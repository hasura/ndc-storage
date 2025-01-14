package functions

import (
	"context"

	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/types"
)

// FunctionStorageIncompleteUploads list partially uploaded objects in a bucket.
func FunctionStorageIncompleteUploads(ctx context.Context, state *types.State, args *common.ListIncompleteUploadsArguments) ([]common.StorageObjectMultipartInfo, error) {
	return state.Storage.ListIncompleteUploads(ctx, args)
}

// ProcedurePutStorageObjectRetention applies object retention lock onto an object.
func ProcedurePutStorageObjectRetention(ctx context.Context, state *types.State, args *common.PutStorageObjectRetentionOptions) (bool, error) {
	if err := state.Storage.PutObjectRetention(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}

// ProcedurePutStorageObjectLegalHold applies legal-hold onto an object.
func ProcedurePutStorageObjectLegalHold(ctx context.Context, state *types.State, args *common.PutStorageObjectLegalHoldOptions) (bool, error) {
	if err := state.Storage.PutObjectLegalHold(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}

// FunctionStorageObjectLegalHold returns legal-hold status on a given object.
func FunctionStorageObjectLegalHold(ctx context.Context, state *types.State, args *common.GetStorageObjectLegalHoldOptions) (common.StorageLegalHoldStatus, error) {
	return state.Storage.GetObjectLegalHold(ctx, args)
}

// ProcedureRemoveIncompleteStorageUpload removes a partially uploaded object.
func ProcedureRemoveIncompleteStorageUpload(ctx context.Context, state *types.State, args *common.RemoveIncompleteUploadArguments) (bool, error) {
	if err := state.Storage.RemoveIncompleteUpload(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}
