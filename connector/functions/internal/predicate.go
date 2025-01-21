package internal

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

// PredicateEvaluator the structured predicate result which is evaluated from the raw expression.
type PredicateEvaluator struct {
	ClientID *common.StorageClientID
	IsValid  bool
	Include  common.StorageObjectIncludeOptions

	variables           map[string]any
	BucketPredicate     StringFilterPredicate
	ObjectNamePredicate StringFilterPredicate
}

// EvalObjectPredicate evaluates the predicate condition of the query request.
func EvalObjectPredicate(bucketInfo common.StorageBucketArguments, objectName string, predicate schema.Expression, variables map[string]any) (*PredicateEvaluator, error) {
	result := &PredicateEvaluator{
		ClientID:  bucketInfo.ClientID,
		Include:   common.StorageObjectIncludeOptions{},
		variables: variables,
	}

	if bucketInfo.Bucket != "" {
		result.BucketPredicate.Pre = &StringComparisonOperator{
			Value:    bucketInfo.Bucket,
			Operator: OperatorEqual,
		}
	}

	if objectName != "" {
		result.ObjectNamePredicate.Pre = &StringComparisonOperator{
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

	result.IsValid = true

	return result, nil
}

// GetBucketArguments get bucket arguments information
func (pe PredicateEvaluator) GetBucketArguments() common.StorageBucketArguments {
	result := common.StorageBucketArguments{
		ClientID: pe.ClientID,
		Bucket:   pe.BucketPredicate.GetPrefix(),
	}

	return result
}

func (pe *PredicateEvaluator) EvalSelection(selection schema.NestedField) error { //nolint:gocognit
	if len(selection) == 0 {
		return nil
	}

	exprT, err := selection.InterfaceT()
	if err != nil {
		return schema.UnprocessableContentError("failed to evaluate selection: "+err.Error(), nil)
	}

	switch expr := exprT.(type) {
	case *schema.NestedArray:
		return pe.EvalSelection(expr.Fields)
	case *schema.NestedObject:
		if objectsField, ok := expr.Fields["objects"]; ok {
			objectsColumn, err := objectsField.AsColumn()
			if err != nil {
				return err
			}

			return pe.EvalSelection(objectsColumn.Fields)
		}

		for _, key := range []string{"metadata", "rawMetadata"} {
			if _, ok := expr.Fields[key]; ok {
				pe.Include.Metadata = true

				break
			}
		}

		for _, key := range checksumColumnNames {
			if _, ok := expr.Fields[key]; ok {
				pe.Include.Checksum = true

				break
			}
		}

		if _, metadataExists := expr.Fields["tags"]; metadataExists {
			pe.Include.Tags = true
		}

		for _, key := range []string{"versionId", "versioning"} {
			if _, ok := expr.Fields[key]; ok {
				pe.Include.Versions = true

				break
			}
		}

		if _, legalHoldExists := expr.Fields["legalHold"]; legalHoldExists {
			pe.Include.LegalHold = true
		}

		if _, ok := expr.Fields["lifecycle"]; ok {
			pe.Include.Lifecycle = true
		}

		if _, ok := expr.Fields["encryption"]; ok {
			pe.Include.Encryption = true
		}

		if _, ok := expr.Fields["objectLock"]; ok {
			pe.Include.ObjectLock = true
		}
	}

	return nil
}

func (pe *PredicateEvaluator) evalQueryPredicate(expression schema.Expression) (bool, error) {
	exprT, err := expression.InterfaceT()
	if err != nil {
		return false, err
	}

	switch expr := exprT.(type) {
	case *schema.ExpressionAnd:
		for _, nestedExpr := range expr.Expressions {
			ok, err := pe.evalQueryPredicate(nestedExpr)
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

		isNull, err := pe.evalIsNullBoolExp(expr)
		if err != nil {
			return false, err
		}

		if isNull != nil && *isNull {
			return false, nil
		}

		switch expr.Column.Name {
		case StorageObjectColumnClientID:
			return pe.evalPredicateClientID(expr)
		case StorageObjectColumnBucket:
			return pe.evalStringFilter(&pe.BucketPredicate, expr)
		case StorageObjectColumnObject:
			return pe.evalStringFilter(&pe.ObjectNamePredicate, expr)
		default:
			return false, errors.New("unsupported predicate on column " + expr.Column.Name)
		}
	default:
		return false, fmt.Errorf("unsupported expression: %+v", expression)
	}
}

func (pe *PredicateEvaluator) evalIsNullBoolExp(expr *schema.ExpressionBinaryComparisonOperator) (*bool, error) {
	if expr.Operator != OperatorIsNull {
		return nil, nil
	}

	boolValue, err := getComparisonValueBoolean(expr.Value, pe.variables)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", expr.Column.Name, err)
	}

	return boolValue, nil
}

func (pe *PredicateEvaluator) evalPredicateClientID(expr *schema.ExpressionBinaryComparisonOperator) (bool, error) {
	switch expr.Operator {
	case OperatorEqual:
		value, err := getComparisonValueString(expr.Value, pe.variables)
		if err != nil {
			return false, fmt.Errorf("clientId: %w", err)
		}

		if value == nil {
			return true, nil
		}

		if pe.ClientID == nil || *pe.ClientID == "" {
			clientID := common.StorageClientID(*value)
			pe.ClientID = &clientID

			return true, nil
		}

		return string(*pe.ClientID) == *value, nil
	default:
		return false, fmt.Errorf("unsupported operator `%s` for clientId", expr.Operator)
	}
}

func (pe *PredicateEvaluator) evalStringFilter(predicate *StringFilterPredicate, expr *schema.ExpressionBinaryComparisonOperator) (bool, error) { //nolint:gocognit,cyclop
	if !slices.Contains([]string{OperatorStartsWith, OperatorEqual}, expr.Operator) {
		return false, fmt.Errorf("unsupported operator `%s` for string filter expression", expr.Operator)
	}

	value, err := getComparisonValueString(expr.Value, pe.variables)
	if err != nil {
		return false, fmt.Errorf("bucket: %w", err)
	}

	if value == nil {
		return true, nil
	}

	valueStr := normalizeObjectName(*value)

	if predicate.Pre == nil {
		if expr.Operator == OperatorStartsWith || expr.Operator == OperatorEqual {
			predicate.Pre = &StringComparisonOperator{
				Value:    valueStr,
				Operator: expr.Operator,
			}
		} else {
			predicate.Post = append(predicate.Post, StringComparisonOperator{
				Value:    valueStr,
				Operator: expr.Operator,
			})
		}

		return true, nil
	}

	switch expr.Operator {
	case OperatorStartsWith:
		switch predicate.Pre.Operator {
		case OperatorStartsWith:
			if len(predicate.Pre.Value) >= len(valueStr) {
				return strings.HasPrefix(predicate.Pre.Value, valueStr), nil
			}

			if !strings.HasPrefix(valueStr, predicate.Pre.Value) {
				return false, nil
			}

			predicate.Pre.Value = valueStr
		case OperatorEqual:
			return strings.HasPrefix(predicate.Pre.Value, valueStr), nil
		}
	case OperatorEqual:
		switch predicate.Pre.Operator {
		case OperatorStartsWith:
			if !strings.HasPrefix(predicate.Pre.Value, valueStr) {
				return false, nil
			}

			predicate.Pre = &StringComparisonOperator{
				Value:    valueStr,
				Operator: OperatorEqual,
			}
		case OperatorEqual:
			return predicate.Pre.Value == valueStr, nil
		}
	case OperatorContains:
		switch predicate.Pre.Operator {
		case OperatorStartsWith:
			if strings.Contains(predicate.Pre.Value, valueStr) {
				return true, nil
			}

			predicate.Post = append(predicate.Post, StringComparisonOperator{
				Value:    valueStr,
				Operator: expr.Operator,
			})
		case OperatorEqual:
			return strings.Contains(predicate.Pre.Value, valueStr), nil
		}
	case OperatorInsensitiveContains:
		switch predicate.Pre.Operator {
		case OperatorStartsWith:
			if strings.Contains(strings.ToLower(predicate.Pre.Value), strings.ToLower(valueStr)) {
				return true, nil
			}

			predicate.Post = append(predicate.Post, StringComparisonOperator{
				Value:    valueStr,
				Operator: expr.Operator,
			})
		case OperatorEqual:
			return strings.Contains(strings.ToLower(predicate.Pre.Value), strings.ToLower(valueStr)), nil
		}
	}

	return true, nil
}

// StringFilterPredicate the structured predicate result which is evaluated from the raw expression.
type StringFilterPredicate struct {
	Pre  *StringComparisonOperator
	Post []StringComparisonOperator
}

// HasPostPredicate checks if the request has post-predicate expressions
func (sfp StringFilterPredicate) GetPrefix() string {
	if sfp.Pre != nil {
		return sfp.Pre.Value
	}

	return ""
}

// HasPostPredicate checks if the request has post-predicate expressions
func (sfp StringFilterPredicate) HasPostPredicate() bool {
	return len(sfp.Post) > 0
}

// CheckPostObjectPredicate the predicate function to filter the object with post conditions
func (sfp StringFilterPredicate) CheckPostPredicate(input string) bool {
	for _, pred := range sfp.Post {
		if (pred.Operator == OperatorContains && !strings.Contains(input, pred.Value)) ||
			(pred.Operator == OperatorInsensitiveContains && !strings.Contains(strings.ToLower(input), strings.ToLower(pred.Value))) {
			return false
		}
	}

	return true
}
