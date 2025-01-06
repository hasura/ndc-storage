package internal

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

// CollectionObjectRequest the structured predicate result which is evaluated from the raw expression.
type CollectionObjectRequest struct {
	IsValid bool
	Options common.ListStorageObjectsOptions
	OrderBy []ColumnOrder

	variables           map[string]any
	objectNamePredicate *StringComparisonOperator
}

// EvalCollectionObjectRequest evaluates the requested collection data of the query request.
func EvalCollectionObjectRequest(request *schema.QueryRequest, arguments map[string]any, variables map[string]any) (*CollectionObjectRequest, error) {
	result := &CollectionObjectRequest{
		variables: variables,
	}

	if len(request.Query.Predicate) > 0 {
		ok, err := result.evalQueryPredicate(request.Query.Predicate)
		if err != nil {
			return nil, err
		}

		if !ok {
			return result, nil
		}
	}

	if result.objectNamePredicate != nil {
		result.Options.Prefix = result.objectNamePredicate.Value
	}

	if err := result.evalArguments(arguments); err != nil {
		return nil, err
	}

	result.evalSelection(request.Query.Fields)

	if request.Query.Limit != nil && *request.Query.Limit > 0 {
		result.Options.MaxKeys = *request.Query.Limit
	}

	orderBy, err := result.evalOrderBy(request.Query.OrderBy)
	if err != nil {
		return nil, err
	}

	result.OrderBy = orderBy
	result.IsValid = true

	return result, nil
}

func (cor *CollectionObjectRequest) evalSelection(selection schema.QueryFields) {
	if _, metadataExists := selection[StorageObjectColumnMetadata]; metadataExists {
		cor.Options.WithMetadata = true
	}

	if _, metadataExists := selection[StorageObjectColumnUserMetadata]; metadataExists {
		cor.Options.WithMetadata = true
	}

	if _, versionExists := selection[StorageObjectColumnVersionID]; versionExists {
		cor.Options.WithVersions = true
	}
}

func (cor *CollectionObjectRequest) evalArguments(arguments map[string]any) error {
	if len(arguments) == 0 {
		return nil
	}

	if rawRecursive, ok := arguments[StorageObjectArgumentRecursive]; ok {
		recursive, err := utils.DecodeNullableBoolean(rawRecursive)
		if err != nil {
			return fmt.Errorf("%s: %w", StorageObjectArgumentRecursive, err)
		}

		if recursive != nil {
			cor.Options.Recursive = *recursive
		}
	}

	return nil
}

func (cor *CollectionObjectRequest) evalQueryPredicate(expression schema.Expression) (bool, error) {
	switch expr := expression.Interface().(type) {
	case *schema.ExpressionAnd:
		for _, nestedExpr := range expr.Expressions {
			ok, err := cor.evalQueryPredicate(nestedExpr)
			if err != nil {
				return false, err
			}

			if !ok {
				return false, nil
			}
		}

		return true, nil
	case *schema.ExpressionBinaryComparisonOperator:
		if expr.Column.Type != schema.ComparisonTargetTypeColumn {
			return false, fmt.Errorf("%s: unsupported comparison target `%s`", expr.Column.Name, expr.Column.Type)
		}

		switch expr.Column.Name {
		case StorageObjectColumnClientID:
			return cor.evalPredicateClientID(expr)
		case StorageObjectColumnBucket:
			return cor.evalPredicateBucket(expr)
		case StorageObjectColumnName:
			return cor.evalObjectName(expr)
		case StorageObjectColumnLastModified:
			switch expr.Operator {
			case OperatorGreater:
				value, err := getComparisonValueDateTime(expr.Value, cor.variables)
				if err != nil {
					return false, fmt.Errorf("lastModified: %w", err)
				}

				if value == nil {
					return true, nil
				}

				valueStr := value.Format(time.RFC3339)
				if cor.Options.StartAfter == "" {
					cor.Options.StartAfter = valueStr

					return true, nil
				}

				return cor.Options.StartAfter == valueStr, nil
			default:
				return false, fmt.Errorf("unsupported operator `%s` for object name", expr.Operator)
			}
		default:
			return false, errors.New("unsupport predicate on column " + expr.Column.Name)
		}
	default:
		return false, fmt.Errorf("unsupported expression: %+v", expression)
	}
}

func (cor *CollectionObjectRequest) evalPredicateClientID(expr *schema.ExpressionBinaryComparisonOperator) (bool, error) {
	switch expr.Operator {
	case OperatorEqual:
		value, err := getComparisonValueString(expr.Value, cor.variables)
		if err != nil {
			return false, fmt.Errorf("clientId: %w", err)
		}

		if value == nil {
			return true, nil
		}

		if cor.Options.ClientID == nil || *cor.Options.ClientID == "" {
			clientID := common.StorageClientID(*value)
			cor.Options.ClientID = &clientID

			return true, nil
		}

		return string(*cor.Options.ClientID) == *value, nil
	default:
		return false, fmt.Errorf("unsupported operator `%s` for clientId", expr.Operator)
	}
}

func (cor *CollectionObjectRequest) evalPredicateBucket(expr *schema.ExpressionBinaryComparisonOperator) (bool, error) {
	switch expr.Operator {
	case OperatorEqual:
		value, err := getComparisonValueString(expr.Value, cor.variables)
		if err != nil {
			return false, fmt.Errorf("bucket: %w", err)
		}

		if value == nil {
			return true, nil
		}

		if cor.Options.Bucket == "" {
			cor.Options.Bucket = *value

			return true, nil
		}

		return cor.Options.Bucket == *value, nil
	default:
		return false, fmt.Errorf("unsupported operator `%s` for bucket", expr.Operator)
	}
}

func (cor *CollectionObjectRequest) evalObjectName(expr *schema.ExpressionBinaryComparisonOperator) (bool, error) {
	if !slices.Contains([]string{OperatorStartsWith, OperatorEqual}, expr.Operator) {
		return false, fmt.Errorf("unsupported operator `%s` for object name", expr.Operator)
	}

	value, err := getComparisonValueString(expr.Value, cor.variables)
	if err != nil {
		return false, fmt.Errorf("bucket: %w", err)
	}

	if value == nil {
		return true, nil
	}

	if cor.objectNamePredicate == nil {
		cor.objectNamePredicate = &StringComparisonOperator{
			Value:    *value,
			Operator: expr.Operator,
		}

		return true, nil
	}

	switch expr.Operator {
	case OperatorStartsWith:
		switch cor.objectNamePredicate.Operator {
		case OperatorStartsWith:
			if len(cor.objectNamePredicate.Value) >= len(*value) {
				return strings.HasPrefix(cor.objectNamePredicate.Value, *value), nil
			}

			if !strings.HasPrefix(*value, cor.objectNamePredicate.Value) {
				return false, nil
			}

			cor.objectNamePredicate.Value = *value
		case OperatorEqual:
			return strings.HasPrefix(cor.objectNamePredicate.Value, *value), nil
		}
	case OperatorEqual:
		switch cor.objectNamePredicate.Operator {
		case OperatorStartsWith:
			if !strings.HasPrefix(cor.objectNamePredicate.Value, *value) {
				return false, nil
			}

			cor.objectNamePredicate = &StringComparisonOperator{
				Value:    *value,
				Operator: OperatorEqual,
			}
		case OperatorEqual:
			return cor.objectNamePredicate.Value == *value, nil
		}
	}

	return true, nil
}

func (cor *CollectionObjectRequest) evalOrderBy(orderBy *schema.OrderBy) ([]ColumnOrder, error) {
	var results []ColumnOrder
	if orderBy == nil {
		return results, nil
	}

	for _, elem := range orderBy.Elements {
		switch target := elem.Target.Interface().(type) {
		case *schema.OrderByColumn:
			// if slices.Contains([]string{metadata.LabelsKey, metadata.ValuesKey}, target.Name) {
			// 	return nil, fmt.Errorf("ordering by `%s` is unsupported", target.Name)
			// }
			orderBy := ColumnOrder{
				Name:       target.Name,
				Descending: elem.OrderDirection == schema.OrderDirectionDesc,
			}
			results = append(results, orderBy)
		default:
			return nil, fmt.Errorf("support ordering by column only, got: %v", elem.Target)
		}
	}

	return results, nil
}
