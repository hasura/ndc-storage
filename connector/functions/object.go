package functions

import (
	"context"
	"io"

	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/collection"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/types"
)

// FunctionStorageObjectConnections lists objects in a bucket using the relay style.
func FunctionStorageObjectConnections(ctx context.Context, state *types.State, args *common.ListStorageObjectsArguments) (StorageConnection[common.StorageObject], error) {
	request, options, err := evalStorageObjectsArguments(ctx, state, args)
	if err != nil {
		return StorageConnection[common.StorageObject]{}, err
	}

	if !request.IsValid {
		return StorageConnection[common.StorageObject]{
			Edges: []StorageConnectionEdge[common.StorageObject]{},
		}, nil
	}

	predicate := request.ObjectNamePredicate.CheckPostPredicate

	if !request.ObjectNamePredicate.HasPostPredicate() {
		predicate = nil
	}

	objects, err := state.Storage.ListObjects(ctx, request.GetBucketArguments(), options, predicate)
	if err != nil {
		return StorageConnection[common.StorageObject]{}, err
	}

	result := StorageConnection[common.StorageObject]{
		Edges:    make([]StorageConnectionEdge[common.StorageObject], len(objects.Objects)),
		PageInfo: objects.PageInfo,
	}

	for i, item := range objects.Objects {
		result.Edges[i] = StorageConnectionEdge[common.StorageObject]{
			Node:   item,
			Cursor: item.Name,
		}
	}

	return result, nil
}

// FunctionStorageDeletedObjects list deleted objects in a bucket.
func FunctionStorageDeletedObjects(ctx context.Context, state *types.State, args *common.ListStorageObjectsArguments) (common.StorageObjectListResults, error) {
	request, options, err := evalStorageObjectsArguments(ctx, state, args)
	if err != nil {
		return common.StorageObjectListResults{}, err
	}

	if !request.IsValid {
		return common.StorageObjectListResults{
			Objects: []common.StorageObject{},
		}, nil
	}

	predicate := request.ObjectNamePredicate.CheckPostPredicate

	if !request.ObjectNamePredicate.HasPostPredicate() {
		predicate = nil
	}

	objects, err := state.Storage.ListDeletedObjects(ctx, request.GetBucketArguments(), options, predicate)
	if err != nil {
		return common.StorageObjectListResults{}, err
	}

	return *objects, nil
}

// FunctionStorageObject fetches metadata of an object.
func FunctionStorageObject(ctx context.Context, state *types.State, args *common.GetStorageObjectArguments) (*common.StorageObject, error) {
	request, err := collection.EvalObjectPredicate(args.StorageBucketArguments, &collection.StringComparisonOperator{
		Value:    args.Object,
		Operator: collection.OperatorEqual,
	}, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return nil, err
	}

	if !request.IsValid {
		return nil, nil
	}

	if err := request.EvalSelection(utils.CommandSelectionFieldFromContext(ctx)); err != nil {
		return nil, err
	}

	opts := args.GetStorageObjectOptions
	opts.Include = request.Include

	return state.Storage.StatObject(ctx, request.GetBucketArguments(), request.ObjectNamePredicate.GetPrefix(), opts)
}

// FunctionDownloadStorageObjectAsBase64 returns a stream of the object data. Most of the common errors occur when reading the stream.
func FunctionDownloadStorageObjectAsBase64(ctx context.Context, state *types.State, args *common.GetStorageObjectArguments) (DownloadStorageObjectResponse, error) {
	args.Base64Encoded = true

	reader, err := downloadStorageObject(ctx, state, args)
	if err != nil {
		return DownloadStorageObjectResponse{}, err
	}

	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return DownloadStorageObjectResponse{}, schema.InternalServerError(err.Error(), nil)
	}

	dataBytes := scalar.NewBytes(data)

	return DownloadStorageObjectResponse{Data: dataBytes}, nil
}

// FunctionDownloadStorageObjectAsText returns the object content in plain text. Use this function only if you know exactly the file as an text file.
func FunctionDownloadStorageObjectAsText(ctx context.Context, state *types.State, args *common.GetStorageObjectArguments) (DownloadStorageObjectTextResponse, error) {
	reader, err := downloadStorageObject(ctx, state, args)
	if err != nil {
		return DownloadStorageObjectTextResponse{}, err
	}

	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return DownloadStorageObjectTextResponse{}, schema.InternalServerError(err.Error(), nil)
	}

	dataStr := string(data)

	return DownloadStorageObjectTextResponse{Data: &dataStr}, nil
}

func downloadStorageObject(ctx context.Context, state *types.State, args *common.GetStorageObjectArguments) (io.ReadCloser, error) {
	request, err := collection.EvalObjectPredicate(args.StorageBucketArguments, &collection.StringComparisonOperator{
		Value:    args.Object,
		Operator: collection.OperatorEqual,
	}, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return nil, err
	}

	if !request.IsValid {
		return nil, nil
	}

	return state.Storage.GetObject(ctx, request.GetBucketArguments(), request.ObjectNamePredicate.GetPrefix(), args.GetStorageObjectOptions)
}

// FunctionStoragePresignedDownloadUrl generates a presigned URL for HTTP GET operations.
// Browsers/Mobile clients may point to this URL to directly download objects even if the bucket is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The maximum expiry is 604800 seconds (i.e. 7 days) and minimum is 1 second.
func FunctionStoragePresignedDownloadUrl(ctx context.Context, state *types.State, args *common.PresignedGetStorageObjectArguments) (*common.PresignedURLResponse, error) {
	request, err := collection.EvalObjectPredicate(args.StorageBucketArguments, &collection.StringComparisonOperator{
		Value:    args.Object,
		Operator: collection.OperatorEqual,
	}, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return nil, err
	}

	if !request.IsValid {
		return nil, nil
	}

	return state.Storage.PresignedGetObject(ctx, request.GetBucketArguments(), request.ObjectNamePredicate.GetPrefix(), args.PresignedGetStorageObjectOptions)
}

// FunctionStoragePresignedUploadUrl generates a presigned URL for HTTP PUT operations.
// Browsers/Mobile clients may point to this URL to upload objects directly to a bucket even if it is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The default expiry is set to 7 days.
func FunctionStoragePresignedUploadUrl(ctx context.Context, state *types.State, args *common.PresignedPutStorageObjectArguments) (*common.PresignedURLResponse, error) {
	request, err := collection.EvalObjectPredicate(args.StorageBucketArguments, &collection.StringComparisonOperator{
		Value:    args.Object,
		Operator: collection.OperatorEqual,
	}, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return nil, err
	}

	if !request.IsValid {
		return nil, nil
	}

	return state.Storage.PresignedPutObject(ctx, request.GetBucketArguments(), request.ObjectNamePredicate.GetPrefix(), args.Expiry)
}

// FunctionStorageIncompleteUploads list partially uploaded objects in a bucket.
func FunctionStorageIncompleteUploads(ctx context.Context, state *types.State, args *common.ListIncompleteUploadsArguments) ([]common.StorageObjectMultipartInfo, error) {
	return state.Storage.ListIncompleteUploads(ctx, args.StorageBucketArguments, args.ListIncompleteUploadsOptions)
}

// ProcedureUploadStorageObjectAsBase64 uploads object that are less than 128MiB in a single PUT operation. For objects that are greater than 128MiB in size,
// PutObject seamlessly uploads the object as parts of 128MiB or more depending on the actual file size. The max upload size for an object is 5TB.
func ProcedureUploadStorageObjectAsBase64(ctx context.Context, state *types.State, args *PutStorageObjectBase64Arguments) (common.StorageUploadInfo, error) {
	return uploadStorageObject(ctx, state, &args.PutStorageObjectArguments, args.Data.Bytes())
}

func uploadStorageObject(ctx context.Context, state *types.State, args *PutStorageObjectArguments, data []byte) (common.StorageUploadInfo, error) {
	request, err := collection.EvalObjectPredicate(args.StorageBucketArguments, &collection.StringComparisonOperator{
		Value:    args.Object,
		Operator: collection.OperatorEqual,
	}, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return common.StorageUploadInfo{}, err
	}

	if !request.IsValid {
		return common.StorageUploadInfo{}, schema.ForbiddenError("permission denied", nil)
	}

	result, err := state.Storage.PutObject(ctx, request.GetBucketArguments(), request.ObjectNamePredicate.GetPrefix(), &args.Options, data)
	if err != nil {
		return common.StorageUploadInfo{}, err
	}

	return *result, nil
}

// ProcedureUploadStorageObjectAsText uploads object in plain text to the storage server. The file content is not encoded to base64 so the input size is smaller than 30%.
func ProcedureUploadStorageObjectAsText(ctx context.Context, state *types.State, args *PutStorageObjectTextArguments) (common.StorageUploadInfo, error) {
	return uploadStorageObject(ctx, state, &args.PutStorageObjectArguments, []byte(args.Data))
}

// ProcedureCopyStorageObject creates or replaces an object through server-side copying of an existing object.
// It supports conditional copying, copying a part of an object and server-side encryption of destination and decryption of source.
// To copy multiple source objects into a single destination object see the ComposeObject API.
func ProcedureCopyStorageObject(ctx context.Context, state *types.State, args *common.CopyStorageObjectArguments) (common.StorageUploadInfo, error) {
	result, err := state.Storage.CopyObject(ctx, args)
	if err != nil {
		return common.StorageUploadInfo{}, err
	}

	return *result, nil
}

// ProcedureComposeStorageObject creates an object by concatenating a list of source objects using server-side copying.
func ProcedureComposeStorageObject(ctx context.Context, state *types.State, args *common.ComposeStorageObjectArguments) (common.StorageUploadInfo, error) {
	result, err := state.Storage.ComposeObject(ctx, args)
	if err != nil {
		return common.StorageUploadInfo{}, err
	}

	return *result, nil
}

// ProcedureUpdateStorageObject updates the object's configuration.
func ProcedureUpdateStorageObject(ctx context.Context, state *types.State, args *common.UpdateStorageObjectArguments) (SuccessResponse, error) {
	request, err := collection.EvalObjectPredicate(args.StorageBucketArguments, &collection.StringComparisonOperator{
		Value:    args.Object,
		Operator: collection.OperatorEqual,
	}, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return SuccessResponse{}, err
	}

	if !request.IsValid {
		return SuccessResponse{}, errPermissionDenied
	}

	if err := state.Storage.UpdateObject(ctx, request.GetBucketArguments(), request.ObjectNamePredicate.GetPrefix(), args.UpdateStorageObjectOptions); err != nil {
		return SuccessResponse{}, err
	}

	return NewSuccessResponse(), nil
}

// ProcedureRemoveStorageObject removes an object with some specified options.
func ProcedureRemoveStorageObject(ctx context.Context, state *types.State, args *common.RemoveStorageObjectArguments) (SuccessResponse, error) {
	request, err := collection.EvalObjectPredicate(args.StorageBucketArguments, &collection.StringComparisonOperator{
		Value:    args.Object,
		Operator: collection.OperatorEqual,
	}, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return SuccessResponse{}, err
	}

	if !request.IsValid {
		return SuccessResponse{}, errPermissionDenied
	}

	if err := state.Storage.RemoveObject(ctx, request.GetBucketArguments(), request.ObjectNamePredicate.GetPrefix(), args.RemoveStorageObjectOptions); err != nil {
		return SuccessResponse{}, err
	}

	return NewSuccessResponse(), nil
}

// ProcedureRemoveStorageObjects remove a list of objects obtained from an input channel. The call sends a delete request to the server up to 1000 objects at a time.
// The errors observed are sent over the error channel.
func ProcedureRemoveStorageObjects(ctx context.Context, state *types.State, args *common.RemoveStorageObjectsArguments) ([]common.RemoveStorageObjectError, error) {
	request, options, err := evalStorageObjectsArguments(ctx, state, &args.ListStorageObjectsArguments)
	if err != nil {
		return []common.RemoveStorageObjectError{}, err
	}

	if !request.IsValid {
		return []common.RemoveStorageObjectError{}, nil
	}

	predicate := request.ObjectNamePredicate.CheckPostPredicate

	if !request.ObjectNamePredicate.HasPostPredicate() {
		predicate = nil
	}

	options.Hierarchy = false

	return state.Storage.RemoveObjects(ctx, request.GetBucketArguments(), &common.RemoveStorageObjectsOptions{
		ListStorageObjectsOptions: *options,
		GovernanceBypass:          args.GovernanceBypass,
	}, predicate)
}

// ProcedureRemoveIncompleteStorageUpload removes a partially uploaded object.
func ProcedureRemoveIncompleteStorageUpload(ctx context.Context, state *types.State, args *common.RemoveIncompleteUploadArguments) (SuccessResponse, error) {
	if err := state.Storage.RemoveIncompleteUpload(ctx, args); err != nil {
		return SuccessResponse{}, err
	}

	return NewSuccessResponse(), nil
}

// ProcedureRestoreStorageObject restore a soft-deleted object.
func ProcedureRestoreStorageObject(ctx context.Context, state *types.State, args *common.RestoreStorageObjectArguments) (SuccessResponse, error) {
	request, err := collection.EvalObjectPredicate(args.StorageBucketArguments, &collection.StringComparisonOperator{
		Value:    args.Object,
		Operator: collection.OperatorEqual,
	}, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return SuccessResponse{}, err
	}

	if !request.IsValid {
		return SuccessResponse{}, errPermissionDenied
	}

	if err := state.Storage.RestoreObject(ctx, request.GetBucketArguments(), request.ObjectNamePredicate.GetPrefix()); err != nil {
		return SuccessResponse{}, err
	}

	return NewSuccessResponse(), nil
}

func evalStorageObjectsArguments(ctx context.Context, state *types.State, args *common.ListStorageObjectsArguments) (*collection.PredicateEvaluator, *common.ListStorageObjectsOptions, error) {
	if args.First != nil && *args.First <= 0 {
		return nil, nil, schema.UnprocessableContentError("$first argument must be larger than 0", nil)
	}

	request, err := collection.EvalObjectPredicate(args.StorageBucketArguments, &collection.StringComparisonOperator{
		Value:    args.Prefix,
		Operator: collection.OperatorStartsWith,
	}, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return nil, nil, err
	}

	if !request.IsValid {
		return request, nil, nil
	}

	if err := request.EvalSelection(utils.CommandSelectionFieldFromContext(ctx)); err != nil {
		return nil, nil, err
	}

	options := &common.ListStorageObjectsOptions{
		Prefix:     request.ObjectNamePredicate.GetPrefix(),
		Hierarchy:  args.Hierarchy,
		Include:    request.Include,
		NumThreads: state.Concurrency.Query,
	}

	if args.First != nil {
		options.MaxResults = *args.First
	}

	if args.After != nil {
		options.StartAfter = *args.After
	}

	return request, options, nil
}
