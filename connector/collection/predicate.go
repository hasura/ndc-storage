package collection

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hasura/ndc-sdk-go/v2/schema"
	"github.com/hasura/ndc-sdk-go/v2/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

// PredicateEvaluator the structured predicate result which is evaluated from the raw expression.
type PredicateEvaluator struct {
	common.StorageClientCredentialArguments

	IsValid           bool
	Include           common.StorageObjectIncludeOptions
	IncludeObjectLock bool

	variables           map[string]any
	BucketPredicate     StringFilterPredicate
	ObjectNamePredicate StringFilterPredicate
}

// EvalBucketPredicate evaluates the predicate bucket condition of the query request.
func EvalBucketPredicate(
	bucketArguments common.StorageClientCredentialArguments,
	preOperator *StringComparisonOperator,
	predicate schema.Expression,
	variables map[string]any,
) (*PredicateEvaluator, error) {
	result := &PredicateEvaluator{
		StorageClientCredentialArguments: bucketArguments,
		Include:                          common.StorageObjectIncludeOptions{},
		variables:                        variables,
	}

	if preOperator != nil {
		result.BucketPredicate.Pre = preOperator
	}

	if len(predicate) > 0 {
		ok, err := result.evalQueryPredicate(predicate, true)
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

// EvalObjectPredicate evaluates the predicate object condition of the query request.
func EvalObjectPredicate(
	bucketInfo common.StorageBucketArguments,
	preOperator *StringComparisonOperator,
	predicate schema.Expression,
	variables map[string]any,
) (*PredicateEvaluator, error) {
	result := &PredicateEvaluator{
		StorageClientCredentialArguments: bucketInfo.StorageClientCredentialArguments,
		Include:                          common.StorageObjectIncludeOptions{},
		variables:                        variables,
	}

	if bucketInfo.Bucket != "" {
		result.BucketPredicate.Pre = &StringComparisonOperator{
			Value:    bucketInfo.Bucket,
			Operator: OperatorEqual,
		}
	}

	if preOperator != nil {
		result.ObjectNamePredicate.Pre = &StringComparisonOperator{
			Value:    normalizeObjectName(preOperator.Value),
			Operator: preOperator.Operator,
		}
	}

	if len(predicate) > 0 {
		ok, err := result.evalQueryPredicate(predicate, false)
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

// GetBucketArguments get bucket arguments information.
func (pe PredicateEvaluator) GetBucketArguments() common.StorageBucketArguments {
	result := common.StorageBucketArguments{
		StorageClientCredentialArguments: common.StorageClientCredentialArguments{
			ClientID:        pe.ClientID,
			ClientType:      pe.ClientType,
			Endpoint:        pe.Endpoint,
			AccessKeyID:     pe.AccessKeyID,
			SecretAccessKey: pe.SecretAccessKey,
		},
		Bucket: pe.BucketPredicate.GetPrefix(),
	}

	return result
}

// EvalArguments evaluate other request arguments.
func (pe *PredicateEvaluator) EvalArguments(arguments map[string]any) error {
	if clientType, err := utils.GetNullableString(arguments, ArgumentClientType); err != nil {
		return schema.UnprocessableContentError(err.Error(), nil)
	} else if clientType != nil {
		pe.ClientType = (*common.StorageProviderType)(clientType)
	}

	if endpoint, err := utils.GetNullableString(arguments, ArgumentEndpoint); err != nil {
		return schema.UnprocessableContentError(err.Error(), nil)
	} else if endpoint != nil {
		pe.Endpoint = *endpoint
	}

	if accessKey, err := utils.GetNullableString(arguments, ArgumentAccessKeyID); err != nil {
		return schema.UnprocessableContentError(err.Error(), nil)
	} else if accessKey != nil {
		pe.AccessKeyID = *accessKey
	}

	if secretAccessKey, err := utils.GetNullableString(arguments, ArgumentSecretAccessKey); err != nil {
		return schema.UnprocessableContentError(err.Error(), nil)
	} else if secretAccessKey != nil {
		pe.SecretAccessKey = *secretAccessKey
	}

	return nil
}

func (pe *PredicateEvaluator) EvalSelection(selection schema.NestedField) error {
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

		pe.evalQuerySelectionFields(expr.Fields)
	}

	return nil
}

func (pe *PredicateEvaluator) evalQuerySelectionFields(fields map[string]schema.Field) {
	for _, key := range []string{"metadata", "raw_metadata"} {
		if _, ok := fields[key]; ok {
			pe.Include.Metadata = true

			break
		}
	}

	for _, key := range checksumColumnNames {
		if _, ok := fields[key]; ok {
			pe.Include.Checksum = true

			break
		}
	}

	if _, ok := fields["tags"]; ok {
		pe.Include.Tags = true
	}

	if _, ok := fields["copy"]; ok {
		pe.Include.Copy = true
	}

	for _, key := range []string{"version_id", "versioning"} {
		if _, ok := fields[key]; ok {
			pe.Include.Versions = true

			break
		}
	}

	if _, legalHoldExists := fields["legal_hold"]; legalHoldExists {
		pe.Include.LegalHold = true
	}

	if _, ok := fields["lifecycle"]; ok {
		pe.Include.Lifecycle = true
	}

	if _, ok := fields["encryption"]; ok {
		pe.Include.Encryption = true
	}

	if _, ok := fields["object_lock"]; ok {
		pe.IncludeObjectLock = true
	}
}

func (pe *PredicateEvaluator) evalQueryPredicate(
	expression schema.Expression,
	forBucket bool,
) (bool, error) {
	exprT, err := expression.InterfaceT()
	if err != nil {
		return false, err
	}

	switch expr := exprT.(type) {
	case *schema.ExpressionAnd:
		for _, nestedExpr := range expr.Expressions {
			ok, err := pe.evalQueryPredicate(nestedExpr, forBucket)
			if err != nil {
				return false, err
			}

			if !ok {
				return false, nil
			}
		}

		return true, nil
	case *schema.ExpressionBinaryComparisonOperator:
		return pe.evalExpressionBinaryComparisonOperator(expr, forBucket)
	default:
		return false, fmt.Errorf("unsupported expression: %+v", expression)
	}
}

func (pe *PredicateEvaluator) evalExpressionBinaryComparisonOperator(
	expr *schema.ExpressionBinaryComparisonOperator,
	forBucket bool,
) (bool, error) {
	columnT, err := expr.Column.InterfaceT()
	if err != nil {
		return false, err
	}

	switch column := columnT.(type) {
	case *schema.ComparisonTargetColumn:
		isNull, err := pe.evalIsNullBoolExp(expr)
		if err != nil {
			return false, fmt.Errorf("%s: %w", column.Name, err)
		}

		if isNull != nil && *isNull {
			return false, nil
		}

		switch column.Name {
		case StorageObjectColumnClientID:
			return pe.evalPredicateClientID(expr)
		case StorageObjectColumnBucket:
			ok, err := pe.evalStringFilter(&pe.BucketPredicate, expr)
			if err != nil {
				return false, fmt.Errorf("%s: %w", StorageObjectColumnBucket, err)
			}

			return ok, nil
		case StorageObjectColumnName:
			var ok bool

			var err error

			if forBucket {
				ok, err = pe.evalStringFilter(&pe.BucketPredicate, expr)
			} else {
				ok, err = pe.evalStringFilter(&pe.ObjectNamePredicate, expr)
			}

			if err != nil {
				return false, fmt.Errorf("%s: %w", StorageObjectColumnName, err)
			}

			return ok, nil
		default:
			return false, errors.New("unsupported predicate on column " + column.Name)
		}
	default:
		return false, fmt.Errorf("unsupported comparison target `%v`", columnT)
	}
}

func (pe *PredicateEvaluator) evalIsNullBoolExp(
	expr *schema.ExpressionBinaryComparisonOperator,
) (*bool, error) {
	if expr.Operator != OperatorIsNull {
		return nil, nil
	}

	return getComparisonValueBoolean(expr.Value, pe.variables)
}

func (pe *PredicateEvaluator) evalPredicateClientID(
	expr *schema.ExpressionBinaryComparisonOperator,
) (bool, error) {
	switch expr.Operator {
	case OperatorEqual:
		value, err := getComparisonValueString(expr.Value, pe.variables)
		if err != nil {
			return false, fmt.Errorf("client_id: %w", err)
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
		return false, fmt.Errorf("unsupported operator `%s` for client_id", expr.Operator)
	}
}

func (pe *PredicateEvaluator) evalStringFilter(
	predicate *StringFilterPredicate,
	expr *schema.ExpressionBinaryComparisonOperator,
) (bool, error) {
	value, err := getComparisonValueString(expr.Value, pe.variables)
	if err != nil {
		return false, err
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
			return strings.Contains(
				strings.ToLower(predicate.Pre.Value),
				strings.ToLower(valueStr),
			), nil
		}
	default:
		return false, fmt.Errorf(
			"unsupported operator `%s` for string filter expression",
			expr.Operator,
		)
	}

	return true, nil
}

// StringFilterPredicate the structured predicate result which is evaluated from the raw expression.
type StringFilterPredicate struct {
	Pre  *StringComparisonOperator
	Post []StringComparisonOperator
}

// HasPostPredicate checks if the request has post-predicate expressions.
func (sfp StringFilterPredicate) GetPrefix() string {
	if sfp.Pre != nil {
		return sfp.Pre.Value
	}

	return ""
}

// HasPostPredicate checks if the request has post-predicate expressions.
func (sfp StringFilterPredicate) HasPostPredicate() bool {
	return len(sfp.Post) > 0
}

// CheckPostObjectPredicate the predicate function to filter the object with post conditions.
func (sfp StringFilterPredicate) CheckPostPredicate(input string) bool {
	for _, pred := range sfp.Post {
		if (pred.Operator == OperatorContains && !strings.Contains(input, pred.Value)) ||
			(pred.Operator == OperatorInsensitiveContains && !strings.Contains(strings.ToLower(input), strings.ToLower(pred.Value))) {
			return false
		}
	}

	return true
}
