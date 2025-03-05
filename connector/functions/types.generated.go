// Code generated by github.com/hasura/ndc-sdk-go/cmd/hasura-ndc-go, DO NOT EDIT.
package functions

import (
	"context"
	"encoding/json"
	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/types"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"slices"
)

// ToMap encodes the struct to a value map
func (j DownloadStorageObjectResponse) ToMap() map[string]any {
	r := make(map[string]any)
	r["data"] = j.Data

	return r
}

// ToMap encodes the struct to a value map
func (j DownloadStorageObjectTextResponse) ToMap() map[string]any {
	r := make(map[string]any)
	r["data"] = j.Data

	return r
}

// ToMap encodes the struct to a value map
func (j ExistsResponse) ToMap() map[string]any {
	r := make(map[string]any)
	r["exists"] = j.Exists

	return r
}

// ToMap encodes the struct to a value map
func (j SuccessResponse) ToMap() map[string]any {
	r := make(map[string]any)
	r["success"] = j.Success

	return r
}

// DataConnectorHandler implements the data connector handler
type DataConnectorHandler struct{}

// QueryExists check if the query name exists
func (dch DataConnectorHandler) QueryExists(name string) bool {
	return slices.Contains(enumValues_FunctionName, name)
}
func (dch DataConnectorHandler) Query(ctx context.Context, state *types.State, request *schema.QueryRequest, rawArgs map[string]any) (*schema.RowSet, error) {
	if !dch.QueryExists(request.Collection) {
		return nil, utils.ErrHandlerNotfound
	}
	queryFields, err := utils.EvalFunctionSelectionFieldValue(request)
	if err != nil {
		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	result, err := dch.execQuery(context.WithValue(ctx, utils.CommandSelectionFieldKey, queryFields), state, request, queryFields, rawArgs)
	if err != nil {
		return nil, err
	}

	return &schema.RowSet{
		Aggregates: schema.RowSetAggregates{},
		Rows: []map[string]any{
			{
				"__value": result,
			},
		},
	}, nil
}

func (dch DataConnectorHandler) execQuery(ctx context.Context, state *types.State, request *schema.QueryRequest, queryFields schema.NestedField, rawArgs map[string]any) (any, error) {
	span := trace.SpanFromContext(ctx)
	logger := connector.GetLogger(ctx)
	switch request.Collection {
	case "downloadStorageObjectAsBase64":

		selection, err := queryFields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.GetStorageObjectArguments
		parseErr := args.FromValue(rawArgs)
		if parseErr != nil {
			return nil, schema.UnprocessableContentError("failed to resolve arguments", map[string]any{
				"cause": parseErr.Error(),
			})
		}

		connector_addSpanEvent(span, logger, "execute_function", map[string]any{
			"arguments": args,
		})
		rawResult, err := FunctionDownloadStorageObjectAsBase64(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)
		if err != nil {
			return nil, err
		}
		return result, nil

	case "downloadStorageObjectAsText":

		selection, err := queryFields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.GetStorageObjectArguments
		parseErr := args.FromValue(rawArgs)
		if parseErr != nil {
			return nil, schema.UnprocessableContentError("failed to resolve arguments", map[string]any{
				"cause": parseErr.Error(),
			})
		}

		connector_addSpanEvent(span, logger, "execute_function", map[string]any{
			"arguments": args,
		})
		rawResult, err := FunctionDownloadStorageObjectAsText(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)
		if err != nil {
			return nil, err
		}
		return result, nil

	case "storageBucket":

		selection, err := queryFields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.GetStorageBucketArguments
		parseErr := args.FromValue(rawArgs)
		if parseErr != nil {
			return nil, schema.UnprocessableContentError("failed to resolve arguments", map[string]any{
				"cause": parseErr.Error(),
			})
		}

		connector_addSpanEvent(span, logger, "execute_function", map[string]any{
			"arguments": args,
		})
		rawResult, err := FunctionStorageBucket(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		if rawResult == nil {
			return nil, nil
		}
		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)
		if err != nil {
			return nil, err
		}
		return result, nil

	case "storageBucketConnections":

		selection, err := queryFields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.ListStorageBucketArguments
		parseErr := args.FromValue(rawArgs)
		if parseErr != nil {
			return nil, schema.UnprocessableContentError("failed to resolve arguments", map[string]any{
				"cause": parseErr.Error(),
			})
		}

		connector_addSpanEvent(span, logger, "execute_function", map[string]any{
			"arguments": args,
		})
		rawResult, err := FunctionStorageBucketConnections(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)
		if err != nil {
			return nil, err
		}
		return result, nil

	case "storageBucketExists":

		selection, err := queryFields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.GetStorageBucketArguments
		parseErr := args.FromValue(rawArgs)
		if parseErr != nil {
			return nil, schema.UnprocessableContentError("failed to resolve arguments", map[string]any{
				"cause": parseErr.Error(),
			})
		}

		connector_addSpanEvent(span, logger, "execute_function", map[string]any{
			"arguments": args,
		})
		rawResult, err := FunctionStorageBucketExists(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)
		if err != nil {
			return nil, err
		}
		return result, nil

	case "storageDeletedObjects":

		selection, err := queryFields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.ListStorageObjectsArguments
		parseErr := args.FromValue(rawArgs)
		if parseErr != nil {
			return nil, schema.UnprocessableContentError("failed to resolve arguments", map[string]any{
				"cause": parseErr.Error(),
			})
		}

		connector_addSpanEvent(span, logger, "execute_function", map[string]any{
			"arguments": args,
		})
		rawResult, err := FunctionStorageDeletedObjects(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)
		if err != nil {
			return nil, err
		}
		return result, nil

	case "storageIncompleteUploads":

		selection, err := queryFields.AsArray()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be array", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.ListIncompleteUploadsArguments
		parseErr := args.FromValue(rawArgs)
		if parseErr != nil {
			return nil, schema.UnprocessableContentError("failed to resolve arguments", map[string]any{
				"cause": parseErr.Error(),
			})
		}

		connector_addSpanEvent(span, logger, "execute_function", map[string]any{
			"arguments": args,
		})
		rawResult, err := FunctionStorageIncompleteUploads(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnArrayIntoSlice(selection, rawResult)
		if err != nil {
			return nil, err
		}
		return result, nil

	case "storageObject":

		selection, err := queryFields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.GetStorageObjectArguments
		parseErr := args.FromValue(rawArgs)
		if parseErr != nil {
			return nil, schema.UnprocessableContentError("failed to resolve arguments", map[string]any{
				"cause": parseErr.Error(),
			})
		}

		connector_addSpanEvent(span, logger, "execute_function", map[string]any{
			"arguments": args,
		})
		rawResult, err := FunctionStorageObject(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		if rawResult == nil {
			return nil, nil
		}
		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)
		if err != nil {
			return nil, err
		}
		return result, nil

	case "storageObjectConnections":

		selection, err := queryFields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.ListStorageObjectsArguments
		parseErr := args.FromValue(rawArgs)
		if parseErr != nil {
			return nil, schema.UnprocessableContentError("failed to resolve arguments", map[string]any{
				"cause": parseErr.Error(),
			})
		}

		connector_addSpanEvent(span, logger, "execute_function", map[string]any{
			"arguments": args,
		})
		rawResult, err := FunctionStorageObjectConnections(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)
		if err != nil {
			return nil, err
		}
		return result, nil

	case "storagePresignedDownloadUrl":

		selection, err := queryFields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.PresignedGetStorageObjectArguments
		parseErr := args.FromValue(rawArgs)
		if parseErr != nil {
			return nil, schema.UnprocessableContentError("failed to resolve arguments", map[string]any{
				"cause": parseErr.Error(),
			})
		}

		connector_addSpanEvent(span, logger, "execute_function", map[string]any{
			"arguments": args,
		})
		rawResult, err := FunctionStoragePresignedDownloadUrl(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		if rawResult == nil {
			return nil, nil
		}
		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)
		if err != nil {
			return nil, err
		}
		return result, nil

	case "storagePresignedUploadUrl":

		selection, err := queryFields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.PresignedPutStorageObjectArguments
		parseErr := args.FromValue(rawArgs)
		if parseErr != nil {
			return nil, schema.UnprocessableContentError("failed to resolve arguments", map[string]any{
				"cause": parseErr.Error(),
			})
		}

		connector_addSpanEvent(span, logger, "execute_function", map[string]any{
			"arguments": args,
		})
		rawResult, err := FunctionStoragePresignedUploadUrl(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		if rawResult == nil {
			return nil, nil
		}
		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)
		if err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, utils.ErrHandlerNotfound
	}
}

var enumValues_FunctionName = []string{"downloadStorageObjectAsBase64", "downloadStorageObjectAsText", "storageBucket", "storageBucketConnections", "storageBucketExists", "storageDeletedObjects", "storageIncompleteUploads", "storageObject", "storageObjectConnections", "storagePresignedDownloadUrl", "storagePresignedUploadUrl"}

// MutationExists check if the mutation name exists
func (dch DataConnectorHandler) MutationExists(name string) bool {
	return slices.Contains(enumValues_ProcedureName, name)
}
func (dch DataConnectorHandler) Mutation(ctx context.Context, state *types.State, operation *schema.MutationOperation) (schema.MutationOperationResults, error) {
	span := trace.SpanFromContext(ctx)
	logger := connector.GetLogger(ctx)
	ctx = context.WithValue(ctx, utils.CommandSelectionFieldKey, operation.Fields)
	connector_addSpanEvent(span, logger, "validate_request", map[string]any{
		"operations_name": operation.Name,
	})

	switch operation.Name {
	case "composeStorageObject":

		selection, err := operation.Fields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.ComposeStorageObjectArguments
		if err := json.Unmarshal(operation.Arguments, &args); err != nil {
			return nil, schema.UnprocessableContentError("failed to decode arguments", map[string]any{
				"cause": err.Error(),
			})
		}
		span.AddEvent("execute_procedure")
		rawResult, err := ProcedureComposeStorageObject(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)

		if err != nil {
			return nil, err
		}
		return schema.NewProcedureResult(result).Encode(), nil

	case "copyStorageObject":

		selection, err := operation.Fields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.CopyStorageObjectArguments
		if err := json.Unmarshal(operation.Arguments, &args); err != nil {
			return nil, schema.UnprocessableContentError("failed to decode arguments", map[string]any{
				"cause": err.Error(),
			})
		}
		span.AddEvent("execute_procedure")
		rawResult, err := ProcedureCopyStorageObject(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)

		if err != nil {
			return nil, err
		}
		return schema.NewProcedureResult(result).Encode(), nil

	case "createStorageBucket":

		selection, err := operation.Fields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.MakeStorageBucketArguments
		if err := json.Unmarshal(operation.Arguments, &args); err != nil {
			return nil, schema.UnprocessableContentError("failed to decode arguments", map[string]any{
				"cause": err.Error(),
			})
		}
		span.AddEvent("execute_procedure")
		rawResult, err := ProcedureCreateStorageBucket(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)

		if err != nil {
			return nil, err
		}
		return schema.NewProcedureResult(result).Encode(), nil

	case "removeIncompleteStorageUpload":

		selection, err := operation.Fields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.RemoveIncompleteUploadArguments
		if err := json.Unmarshal(operation.Arguments, &args); err != nil {
			return nil, schema.UnprocessableContentError("failed to decode arguments", map[string]any{
				"cause": err.Error(),
			})
		}
		span.AddEvent("execute_procedure")
		rawResult, err := ProcedureRemoveIncompleteStorageUpload(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)

		if err != nil {
			return nil, err
		}
		return schema.NewProcedureResult(result).Encode(), nil

	case "removeStorageBucket":

		selection, err := operation.Fields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.GetStorageBucketArguments
		if err := json.Unmarshal(operation.Arguments, &args); err != nil {
			return nil, schema.UnprocessableContentError("failed to decode arguments", map[string]any{
				"cause": err.Error(),
			})
		}
		span.AddEvent("execute_procedure")
		rawResult, err := ProcedureRemoveStorageBucket(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)

		if err != nil {
			return nil, err
		}
		return schema.NewProcedureResult(result).Encode(), nil

	case "removeStorageObject":

		selection, err := operation.Fields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.RemoveStorageObjectArguments
		if err := json.Unmarshal(operation.Arguments, &args); err != nil {
			return nil, schema.UnprocessableContentError("failed to decode arguments", map[string]any{
				"cause": err.Error(),
			})
		}
		span.AddEvent("execute_procedure")
		rawResult, err := ProcedureRemoveStorageObject(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)

		if err != nil {
			return nil, err
		}
		return schema.NewProcedureResult(result).Encode(), nil

	case "removeStorageObjects":

		selection, err := operation.Fields.AsArray()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be array", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.RemoveStorageObjectsArguments
		if err := json.Unmarshal(operation.Arguments, &args); err != nil {
			return nil, schema.UnprocessableContentError("failed to decode arguments", map[string]any{
				"cause": err.Error(),
			})
		}
		span.AddEvent("execute_procedure")
		rawResult, err := ProcedureRemoveStorageObjects(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnArrayIntoSlice(selection, rawResult)

		if err != nil {
			return nil, err
		}
		return schema.NewProcedureResult(result).Encode(), nil

	case "restoreStorageObject":

		selection, err := operation.Fields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.RestoreStorageObjectArguments
		if err := json.Unmarshal(operation.Arguments, &args); err != nil {
			return nil, schema.UnprocessableContentError("failed to decode arguments", map[string]any{
				"cause": err.Error(),
			})
		}
		span.AddEvent("execute_procedure")
		rawResult, err := ProcedureRestoreStorageObject(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)

		if err != nil {
			return nil, err
		}
		return schema.NewProcedureResult(result).Encode(), nil

	case "updateStorageBucket":

		selection, err := operation.Fields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.UpdateBucketArguments
		if err := json.Unmarshal(operation.Arguments, &args); err != nil {
			return nil, schema.UnprocessableContentError("failed to decode arguments", map[string]any{
				"cause": err.Error(),
			})
		}
		span.AddEvent("execute_procedure")
		rawResult, err := ProcedureUpdateStorageBucket(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)

		if err != nil {
			return nil, err
		}
		return schema.NewProcedureResult(result).Encode(), nil

	case "updateStorageObject":

		selection, err := operation.Fields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.UpdateStorageObjectArguments
		if err := json.Unmarshal(operation.Arguments, &args); err != nil {
			return nil, schema.UnprocessableContentError("failed to decode arguments", map[string]any{
				"cause": err.Error(),
			})
		}
		span.AddEvent("execute_procedure")
		rawResult, err := ProcedureUpdateStorageObject(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)

		if err != nil {
			return nil, err
		}
		return schema.NewProcedureResult(result).Encode(), nil

	case "uploadStorageObjectAsBase64":

		selection, err := operation.Fields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args PutStorageObjectBase64Arguments
		if err := json.Unmarshal(operation.Arguments, &args); err != nil {
			return nil, schema.UnprocessableContentError("failed to decode arguments", map[string]any{
				"cause": err.Error(),
			})
		}
		span.AddEvent("execute_procedure")
		rawResult, err := ProcedureUploadStorageObjectAsBase64(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)

		if err != nil {
			return nil, err
		}
		return schema.NewProcedureResult(result).Encode(), nil

	case "uploadStorageObjectAsText":

		selection, err := operation.Fields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args PutStorageObjectTextArguments
		if err := json.Unmarshal(operation.Arguments, &args); err != nil {
			return nil, schema.UnprocessableContentError("failed to decode arguments", map[string]any{
				"cause": err.Error(),
			})
		}
		span.AddEvent("execute_procedure")
		rawResult, err := ProcedureUploadStorageObjectAsText(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)

		if err != nil {
			return nil, err
		}
		return schema.NewProcedureResult(result).Encode(), nil

	case "uploadStorageObjectFromUrl":

		selection, err := operation.Fields.AsObject()
		if err != nil {
			return nil, schema.UnprocessableContentError("the selection field type must be object", map[string]any{
				"cause": err.Error(),
			})
		}
		var args common.UploadStorageObjectFromURLArguments
		if err := json.Unmarshal(operation.Arguments, &args); err != nil {
			return nil, schema.UnprocessableContentError("failed to decode arguments", map[string]any{
				"cause": err.Error(),
			})
		}
		span.AddEvent("execute_procedure")
		rawResult, err := ProcedureUploadStorageObjectFromURL(ctx, state, &args)

		if err != nil {
			return nil, err
		}

		connector_addSpanEvent(span, logger, "evaluate_response_selection", map[string]any{
			"raw_result": rawResult,
		})
		result, err := utils.EvalNestedColumnObject(selection, rawResult)

		if err != nil {
			return nil, err
		}
		return schema.NewProcedureResult(result).Encode(), nil

	default:
		return nil, utils.ErrHandlerNotfound
	}
}

var enumValues_ProcedureName = []string{"composeStorageObject", "copyStorageObject", "createStorageBucket", "removeIncompleteStorageUpload", "removeStorageBucket", "removeStorageObject", "removeStorageObjects", "restoreStorageObject", "updateStorageBucket", "updateStorageObject", "uploadStorageObjectAsBase64", "uploadStorageObjectAsText", "uploadStorageObjectFromUrl"}

func connector_addSpanEvent(span trace.Span, logger *slog.Logger, name string, data map[string]any, options ...trace.EventOption) {
	logger.Debug(name, slog.Any("data", data))
	attrs := utils.DebugJSONAttributes(data, utils.IsDebug(logger))
	span.AddEvent(name, append(options, trace.WithAttributes(attrs...))...)
}
