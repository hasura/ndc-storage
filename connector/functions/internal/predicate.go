package internal

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

// ObjectPredicate the structured predicate result which is evaluated from the raw expression.
type ObjectPredicate struct {
	common.StorageBucketArguments

	IsValid bool
	Include common.StorageObjectIncludeOptions
	Prefix  string

	variables                map[string]any
	objectNamePrePredicate   *StringComparisonOperator
	objectNamePostPredicates []StringComparisonOperator
}

// EvalObjectPredicate evaluates the predicate condition of the query request.
func EvalObjectPredicate(bucketInfo common.StorageBucketArguments, objectName string, predicate schema.Expression, variables map[string]any) (*ObjectPredicate, error) {
	result := &ObjectPredicate{
		StorageBucketArguments: bucketInfo,
		Include:                common.StorageObjectIncludeOptions{},
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
		result.Prefix = result.objectNamePrePredicate.Value
	}

	result.IsValid = true

	return result, nil
}

// HasPostPredicate checks if the request has post-predicate expressions
func (cor ObjectPredicate) HasPostPredicate() bool {
	return len(cor.objectNamePostPredicates) > 0
}

func (cor *ObjectPredicate) EvalSelection(selection schema.NestedField) error {
	if len(selection) == 0 {
		return nil
	}

	exprT, err := selection.InterfaceT()
	if err != nil {
		return schema.UnprocessableContentError("failed to evaluate selection: "+err.Error(), nil)
	}

	switch expr := exprT.(type) {
	case *schema.NestedArray:
		return cor.EvalSelection(expr.Fields)
	case *schema.NestedObject:
		if objectsField, ok := expr.Fields["objects"]; ok {
			objectsColumn, err := objectsField.AsColumn()
			if err != nil {
				return err
			}

			return cor.EvalSelection(objectsColumn.Fields)
		}

		if _, metadataExists := expr.Fields["metadata"]; metadataExists {
			cor.Include.Metadata = true
		} else if _, metadataExists := expr.Fields["userMetadata"]; metadataExists {
			cor.Include.Metadata = true
		}

		for _, key := range checksumColumnNames {
			if _, ok := expr.Fields[key]; ok {
				cor.Include.Checksum = true

				break
			}
		}

		if _, metadataExists := expr.Fields["userTags"]; metadataExists {
			cor.Include.Tags = true
		} else if _, metadataExists := expr.Fields["tags"]; metadataExists {
			cor.Include.Tags = true
		}

		if _, versionExists := expr.Fields["versionId"]; versionExists {
			cor.Include.Versions = true
		}

		if _, legalHoldExists := expr.Fields["legalHold"]; legalHoldExists {
			cor.Include.LegalHold = true
		}
	}

	return nil
}

func (cor *ObjectPredicate) evalQueryPredicate(expression schema.Expression) (bool, error) {
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
		case StorageObjectColumnObject:
			return cor.evalObjectName(expr)
		default:
			return false, errors.New("unsupport predicate on column " + expr.Column.Name)
		}
	default:
		return false, fmt.Errorf("unsupported expression: %+v", expression)
	}
}

func (cor *ObjectPredicate) evalIsNullBoolExp(expr *schema.ExpressionBinaryComparisonOperator) (*bool, error) {
	if expr.Operator != OperatorIsNull {
		return nil, nil
	}

	boolValue, err := getComparisonValueBoolean(expr.Value, cor.variables)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", expr.Column.Name, err)
	}

	return boolValue, nil
}

func (cor *ObjectPredicate) evalPredicateClientID(expr *schema.ExpressionBinaryComparisonOperator) (bool, error) {
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

func (cor *ObjectPredicate) evalPredicateBucket(expr *schema.ExpressionBinaryComparisonOperator) (bool, error) {
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

func (cor *ObjectPredicate) evalObjectName(expr *schema.ExpressionBinaryComparisonOperator) (bool, error) { //nolint:gocognit,cyclop
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

// CheckPostObjectPredicate the predicate function to filter the object with post conditions
func (cor ObjectPredicate) CheckPostObjectPredicate(input common.StorageObject) bool {
	if len(cor.objectNamePostPredicates) == 0 {
		return true
	}

	return cor.CheckPostObjectNamePredicate(input.Name)
}

// CheckPostObjectPredicate the predicate function to filter the object with post conditions
func (cor ObjectPredicate) CheckPostObjectNamePredicate(name string) bool {
	for _, pred := range cor.objectNamePostPredicates {
		if (pred.Operator == OperatorContains && !strings.Contains(name, pred.Value)) ||
			(pred.Operator == OperatorInsensitiveContains && !strings.Contains(strings.ToLower(name), strings.ToLower(pred.Value))) {
			return false
		}
	}

	return true
}
