package functions

import (
	"context"
	"io"

	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/collection"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/types"
)

// FunctionStorageIncompleteUploads list partially uploaded objects in a bucket.
func FunctionStorageIncompleteUploads(ctx context.Context, state *types.State, args *common.ListIncompleteUploadsArguments) ([]common.StorageObjectMultipartInfo, error) {
	return state.Storage.ListIncompleteUploads(ctx, args)
}

// FunctionDownloadStorageObject returns a stream of the object data. Most of the common errors occur when reading the stream.
func FunctionDownloadStorageObject(ctx context.Context, state *types.State, args *common.GetStorageObjectArguments) (*scalar.Bytes, error) {
	reader, err := downloadStorageObject(ctx, state, args)
	if err != nil {
		return nil, err
	}

	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, schema.InternalServerError(err.Error(), nil)
	}

	return scalar.NewBytes(data), nil
}

// FunctionDownloadStorageObjectText returns the object content in plain text. Use this function only if you know exactly the file as an text file.
func FunctionDownloadStorageObjectText(ctx context.Context, state *types.State, args *common.GetStorageObjectArguments) (*string, error) {
	reader, err := downloadStorageObject(ctx, state, args)
	if err != nil {
		return nil, err
	}

	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, schema.InternalServerError(err.Error(), nil)
	}

	dataStr := string(data)

	return &dataStr, nil
}

func downloadStorageObject(ctx context.Context, state *types.State, args *common.GetStorageObjectArguments) (io.ReadCloser, error) {
	request, err := collection.EvalCollectionObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return nil, err
	}

	if !request.IsValid {
		return nil, nil
	}

	return state.Storage.GetObject(ctx, request.StorageBucketArguments, request.Options.Prefix, args.GetStorageObjectOptions)
}

// FunctionStorageObject fetches metadata of an object.
func FunctionStorageObject(ctx context.Context, state *types.State, args *common.GetStorageObjectArguments) (*common.StorageObject, error) {
	request, err := collection.EvalCollectionObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return nil, err
	}

	if !request.IsValid {
		return nil, nil
	}

	return state.Storage.StatObject(ctx, request.StorageBucketArguments, request.Options.Prefix, args.GetStorageObjectOptions)
}

// ProcedureRemoveStorageObject removes an object with some specified options.
func ProcedureRemoveStorageObject(ctx context.Context, state *types.State, args *common.RemoveStorageObjectArguments) (bool, error) {
	request, err := collection.EvalCollectionObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return false, err
	}

	if !request.IsValid {
		return false, nil
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
	request, err := collection.EvalCollectionObjectPredicate(args.StorageBucketArguments, "", args.Where, types.QueryVariablesFromContext(ctx))
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

// ProcedurePutStorageObjectTags sets new object Tags to the given object, replaces/overwrites any existing tags.
func ProcedurePutStorageObjectTags(ctx context.Context, state *types.State, args *common.PutStorageObjectTaggingOptions) (bool, error) {
	if err := state.Storage.PutObjectTagging(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}

// FunctionStorageObjectTags fetches Object Tags from the given object.
func FunctionStorageObjectTags(ctx context.Context, state *types.State, args *common.StorageObjectTaggingOptions) (map[string]string, error) {
	return state.Storage.GetObjectTagging(ctx, args)
}

// ProcedureRemoveStorageObjectTags removes Object Tags from the given object.
func ProcedureRemoveStorageObjectTags(ctx context.Context, state *types.State, args *common.StorageObjectTaggingOptions) (bool, error) {
	if err := state.Storage.RemoveObjectTagging(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}

// FunctionStorageObjectAttributes returns a stream of the object data. Most of the common errors occur when reading the stream.
func FunctionStorageObjectAttributes(ctx context.Context, state *types.State, args *common.StorageObjectAttributesOptions) (*common.StorageObjectAttributes, error) {
	return state.Storage.GetObjectAttributes(ctx, args)
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
	request, err := collection.EvalCollectionObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
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
	request, err := collection.EvalCollectionObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return nil, err
	}

	if !request.IsValid {
		return nil, nil
	}

	return state.Storage.PresignedPutObject(ctx, request.StorageBucketArguments, request.Options.Prefix, args.Expiry)
}
