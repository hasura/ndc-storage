package collection

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
	common.StorageBucketArguments

	IsValid bool
	Options common.ListStorageObjectsOptions
	OrderBy []ColumnOrder

	variables                map[string]any
	objectNamePrePredicate   *StringComparisonOperator
	objectNamePostPredicates []StringComparisonOperator
}

// EvalCollectionObjectsRequest evaluates the requested collection data of the query request.
func EvalCollectionObjectsRequest(request *schema.QueryRequest, arguments map[string]any, variables map[string]any) (*CollectionObjectRequest, error) {
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

	if result.objectNamePrePredicate != nil {
		result.Options.Prefix = result.objectNamePrePredicate.Value
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

// EvalCollectionObjectPredicate evaluates the predicate condition of the query request.
func EvalCollectionObjectPredicate(bucketInfo common.StorageBucketArguments, objectName string, predicate schema.Expression, variables map[string]any) (*CollectionObjectRequest, error) {
	result := &CollectionObjectRequest{
		StorageBucketArguments: bucketInfo,
		Options:                common.ListStorageObjectsOptions{},
		variables:              variables,
	}

	if objectName != "" {
		result.objectNamePrePredicate = &StringComparisonOperator{
			Value:    normalizeObjectName(objectName),
			Operator: OperatorEqual,
		}
	}

	if len(predicate) > 0 {
		ok, err := result.evalQueryPredicate(predicate)
		if err != nil {
			return nil, err
		}

		if !ok {
			return result, nil
		}
	}

	if result.objectNamePrePredicate != nil {
		result.Options.Prefix = result.objectNamePrePredicate.Value
	}

	result.IsValid = true

	return result, nil
}

// HasPostPredicate checks if the request has post-predicate expressions
func (cor CollectionObjectRequest) HasPostPredicate() bool {
	return len(cor.objectNamePostPredicates) > 0
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
	exprT, err := expression.InterfaceT()
	if err != nil {
		return false, err
	}

	switch expr := exprT.(type) {
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

		isNull, err := cor.evalIsNullBoolExp(expr)
		if err != nil {
			return false, err
		}

		if isNull != nil && *isNull {
			return false, nil
		}

		switch expr.Column.Name {
		case StorageObjectColumnClientID:
			return cor.evalPredicateClientID(expr)
		case StorageObjectColumnBucket:
			return cor.evalPredicateBucket(expr)
		case StorageObjectColumnName:
			return cor.evalObjectName(expr)
		case StorageObjectColumnLastModified:
			return cor.evalPredicateLastModified(expr)
		default:
			return false, errors.New("unsupport predicate on column " + expr.Column.Name)
		}
	default:
		return false, fmt.Errorf("unsupported expression: %+v", expression)
	}
}

func (cor *CollectionObjectRequest) evalPredicateLastModified(expr *schema.ExpressionBinaryComparisonOperator) (bool, error) {
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
}

func (cor *CollectionObjectRequest) evalIsNullBoolExp(expr *schema.ExpressionBinaryComparisonOperator) (*bool, error) {
	if expr.Operator != OperatorIsNull {
		return nil, nil
	}

	boolValue, err := getComparisonValueBoolean(expr.Value, cor.variables)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", expr.Column.Name, err)
	}

	return boolValue, nil
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

		if cor.ClientID == nil || *cor.ClientID == "" {
			clientID := common.StorageClientID(*value)
			cor.ClientID = &clientID

			return true, nil
		}

		return string(*cor.ClientID) == *value, nil
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

		if cor.Bucket == "" {
			cor.Bucket = *value

			return true, nil
		}

		return cor.Bucket == *value, nil
	default:
		return false, fmt.Errorf("unsupported operator `%s` for bucket", expr.Operator)
	}
}

func (cor *CollectionObjectRequest) evalObjectName(expr *schema.ExpressionBinaryComparisonOperator) (bool, error) { //nolint:gocognit,cyclop
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

	valueStr := normalizeObjectName(*value)

	if cor.objectNamePrePredicate == nil {
		if expr.Operator == OperatorStartsWith || expr.Operator == OperatorEqual {
			cor.objectNamePrePredicate = &StringComparisonOperator{
				Value:    valueStr,
				Operator: expr.Operator,
			}
		} else {
			cor.objectNamePostPredicates = append(cor.objectNamePostPredicates, StringComparisonOperator{
				Value:    valueStr,
				Operator: expr.Operator,
			})
		}

		return true, nil
	}

	switch expr.Operator {
	case OperatorStartsWith:
		switch cor.objectNamePrePredicate.Operator {
		case OperatorStartsWith:
			if len(cor.objectNamePrePredicate.Value) >= len(valueStr) {
				return strings.HasPrefix(cor.objectNamePrePredicate.Value, valueStr), nil
			}

			if !strings.HasPrefix(valueStr, cor.objectNamePrePredicate.Value) {
				return false, nil
			}

			cor.objectNamePrePredicate.Value = valueStr
		case OperatorEqual:
			return strings.HasPrefix(cor.objectNamePrePredicate.Value, valueStr), nil
		}
	case OperatorEqual:
		switch cor.objectNamePrePredicate.Operator {
		case OperatorStartsWith:
			if !strings.HasPrefix(cor.objectNamePrePredicate.Value, valueStr) {
				return false, nil
			}

			cor.objectNamePrePredicate = &StringComparisonOperator{
				Value:    valueStr,
				Operator: OperatorEqual,
			}
		case OperatorEqual:
			return cor.objectNamePrePredicate.Value == valueStr, nil
		}
	case OperatorContains:
		switch cor.objectNamePrePredicate.Operator {
		case OperatorStartsWith:
			if strings.Contains(cor.objectNamePrePredicate.Value, valueStr) {
				return true, nil
			}

			cor.objectNamePostPredicates = append(cor.objectNamePostPredicates, StringComparisonOperator{
				Value:    valueStr,
				Operator: expr.Operator,
			})
		case OperatorEqual:
			return strings.Contains(cor.objectNamePrePredicate.Value, valueStr), nil
		}
	case OperatorInsensitiveContains:
		switch cor.objectNamePrePredicate.Operator {
		case OperatorStartsWith:
			if strings.Contains(strings.ToLower(cor.objectNamePrePredicate.Value), strings.ToLower(valueStr)) {
				return true, nil
			}

			cor.objectNamePostPredicates = append(cor.objectNamePostPredicates, StringComparisonOperator{
				Value:    valueStr,
				Operator: expr.Operator,
			})
		case OperatorEqual:
			return strings.Contains(strings.ToLower(cor.objectNamePrePredicate.Value), strings.ToLower(valueStr)), nil
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

// CheckPostObjectPredicate the predicate function to filter the object with post conditions
func (cor CollectionObjectRequest) CheckPostObjectPredicate(input common.StorageObject) bool {
	if len(cor.objectNamePostPredicates) == 0 {
		return true
	}

	return cor.CheckPostObjectNamePredicate(input.Name)
}

// CheckPostObjectPredicate the predicate function to filter the object with post conditions
func (cor CollectionObjectRequest) CheckPostObjectNamePredicate(name string) bool {
	for _, pred := range cor.objectNamePostPredicates {
		if (pred.Operator == OperatorContains && !strings.Contains(name, pred.Value)) ||
			(pred.Operator == OperatorInsensitiveContains && !strings.Contains(strings.ToLower(name), strings.ToLower(pred.Value))) {
			return false
		}
	}

	return true
}
