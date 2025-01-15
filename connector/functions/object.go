package functions

import (
	"context"

	"github.com/hasura/ndc-storage/connector/functions/internal"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/types"
)

// FunctionStorageIncompleteUploads list partially uploaded objects in a bucket.
func FunctionStorageIncompleteUploads(ctx context.Context, state *types.State, args *common.ListIncompleteUploadsArguments) ([]common.StorageObjectMultipartInfo, error) {
	return state.Storage.ListIncompleteUploads(ctx, args.StorageBucketArguments, args.ListIncompleteUploadsOptions)
}

// ProcedurePutStorageObjectRetention applies object retention lock onto an object.
func ProcedurePutStorageObjectRetention(ctx context.Context, state *types.State, args *common.PutStorageObjectRetentionOptions) (bool, error) {
	if err := state.Storage.PutObjectRetention(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}

// ProcedureSetStorageObjectLegalHold applies legal-hold onto an object.
func ProcedureSetStorageObjectLegalHold(ctx context.Context, state *types.State, args *common.SetStorageObjectLegalHoldArguments) (bool, error) {
	request, err := internal.EvalObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return false, err
	}

	if !request.IsValid {
		return false, errPermissionDenied
	}

	if err := state.Storage.SetObjectLegalHold(ctx, request.StorageBucketArguments, request.Prefix, args.SetStorageObjectLegalHoldOptions); err != nil {
		return false, err
	}

	return true, nil
}

// ProcedureRemoveIncompleteStorageUpload removes a partially uploaded object.
func ProcedureRemoveIncompleteStorageUpload(ctx context.Context, state *types.State, args *common.RemoveIncompleteUploadArguments) (bool, error) {
	if err := state.Storage.RemoveIncompleteUpload(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}
