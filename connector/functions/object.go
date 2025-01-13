package functions

import (
	"context"

	"github.com/hasura/ndc-storage/connector/functions/internal"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/types"
)

// FunctionStorageIncompleteUploads list partially uploaded objects in a bucket.
func FunctionStorageIncompleteUploads(ctx context.Context, state *types.State, args *common.ListIncompleteUploadsArguments) ([]common.StorageObjectMultipartInfo, error) {
	return state.Storage.ListIncompleteUploads(ctx, args)
}

// ProcedureRemoveStorageObject removes an object with some specified options.
func ProcedureRemoveStorageObject(ctx context.Context, state *types.State, args *common.RemoveStorageObjectArguments) (bool, error) {
	request, err := internal.EvalObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return false, err
	}

	if !request.IsValid {
		return false, errPermissionDenied
	}

	if err := state.Storage.RemoveObject(ctx, request.StorageBucketArguments, request.Options.Prefix, args.RemoveStorageObjectOptions); err != nil {
		return false, err
	}

	return true, nil
}

// ProcedurePutStorageObjectRetention applies object retention lock onto an object.
func ProcedurePutStorageObjectRetention(ctx context.Context, state *types.State, args *common.PutStorageObjectRetentionOptions) (bool, error) {
	if err := state.Storage.PutObjectRetention(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}

// ProcedureRemoveStorageObjects remove a list of objects obtained from an input channel. The call sends a delete request to the server up to 1000 objects at a time.
// The errors observed are sent over the error channel.
func ProcedureRemoveStorageObjects(ctx context.Context, state *types.State, args *common.RemoveStorageObjectsArguments) ([]common.RemoveStorageObjectError, error) {
	request, err := internal.EvalObjectPredicate(args.StorageBucketArguments, "", args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return nil, err
	}

	if !request.IsValid {
		return nil, nil
	}

	predicate := request.CheckPostObjectNamePredicate
	if !request.HasPostPredicate() {
		predicate = nil
	}

	return state.Storage.RemoveObjects(ctx, request.StorageBucketArguments, &common.RemoveStorageObjectsOptions{
		ListStorageObjectsOptions: request.Options,
		GovernanceBypass:          args.GovernanceBypass,
	}, predicate)
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

// FunctionStoragePresignedDownloadUrl generates a presigned URL for HTTP GET operations.
// Browsers/Mobile clients may point to this URL to directly download objects even if the bucket is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The maximum expiry is 604800 seconds (i.e. 7 days) and minimum is 1 second.
func FunctionStoragePresignedDownloadUrl(ctx context.Context, state *types.State, args *common.PresignedGetStorageObjectArguments) (*common.PresignedURLResponse, error) {
	request, err := internal.EvalObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return nil, err
	}

	if !request.IsValid {
		return nil, nil
	}

	return state.Storage.PresignedGetObject(ctx, request.StorageBucketArguments, request.Options.Prefix, args.PresignedGetStorageObjectOptions)
}

// FunctionStoragePresignedUploadUrl generates a presigned URL for HTTP PUT operations.
// Browsers/Mobile clients may point to this URL to upload objects directly to a bucket even if it is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The default expiry is set to 7 days.
func FunctionStoragePresignedUploadUrl(ctx context.Context, state *types.State, args *common.PresignedPutStorageObjectArguments) (*common.PresignedURLResponse, error) {
	request, err := internal.EvalObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return nil, err
	}

	if !request.IsValid {
		return nil, nil
	}

	return state.Storage.PresignedPutObject(ctx, request.StorageBucketArguments, request.Options.Prefix, args.Expiry)
}
