package functions

import (
	"context"
	"io"

	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/functions/internal"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/types"
)

// FunctionStorageObjects lists objects in a bucket.
func FunctionStorageObjects(ctx context.Context, state *types.State, args *common.ListStorageObjectsArguments) (common.StorageObjectListResults, error) {
	if args.MaxResults != nil && *args.MaxResults <= 0 {
		return common.StorageObjectListResults{}, schema.UnprocessableContentError("maxResults must be larger than 0", nil)
	}

	request, err := internal.EvalObjectPredicate(common.StorageBucketArguments{}, "", args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return common.StorageObjectListResults{}, err
	}

	if !request.IsValid {
		return common.StorageObjectListResults{
			Objects: []common.StorageObject{},
		}, nil
	}

	if err := request.EvalSelection(utils.CommandSelectionFieldFromContext(ctx)); err != nil {
		return common.StorageObjectListResults{}, err
	}

	options := &common.ListStorageObjectsOptions{
		Prefix:     request.ObjectNamePredicate.GetPrefix(),
		Recursive:  args.Recursive,
		Include:    request.Include,
		NumThreads: state.Concurrency.Query,
	}

	if args.MaxResults != nil {
		options.MaxResults = *args.MaxResults
	}

	if args.StartAfter != nil {
		options.StartAfter = *args.StartAfter
	}

	predicate := request.ObjectNamePredicate.CheckPostPredicate

	if !request.ObjectNamePredicate.HasPostPredicate() {
		predicate = nil
	}

	objects, err := state.Storage.ListObjects(ctx, request.GetBucketArguments(), options, predicate)
	if err != nil {
		return common.StorageObjectListResults{}, err
	}

	return *objects, nil
}

// FunctionStorageObject fetches metadata of an object.
func FunctionStorageObject(ctx context.Context, state *types.State, args *common.GetStorageObjectArguments) (*common.StorageObject, error) {
	request, err := internal.EvalObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
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
	request, err := internal.EvalObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
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
	request, err := internal.EvalObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
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
	request, err := internal.EvalObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
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

// PutStorageObjectArguments represents input arguments of the PutObject method.
type PutStorageObjectArguments struct {
	common.StorageBucketArguments

	Object  string                         `json:"object"`
	Options common.PutStorageObjectOptions `json:"options,omitempty"`
	Where   schema.Expression              `json:"where"             ndc:"predicate=StorageObjectFilter"`
}

// PutStorageObjectArguments represents input arguments of the PutObject method.
type PutStorageObjectBase64Arguments struct {
	PutStorageObjectArguments

	Data scalar.Bytes `json:"data"`
}

// ProcedureUploadStorageObject uploads object that are less than 128MiB in a single PUT operation. For objects that are greater than 128MiB in size,
// PutObject seamlessly uploads the object as parts of 128MiB or more depending on the actual file size. The max upload size for an object is 5TB.
func ProcedureUploadStorageObject(ctx context.Context, state *types.State, args *PutStorageObjectBase64Arguments) (common.StorageUploadInfo, error) {
	return uploadStorageObject(ctx, state, &args.PutStorageObjectArguments, args.Data.Bytes())
}

func uploadStorageObject(ctx context.Context, state *types.State, args *PutStorageObjectArguments, data []byte) (common.StorageUploadInfo, error) {
	request, err := internal.EvalObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
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

// PutStorageObjectTextArguments represents input arguments of the PutStorageObjectText method.
type PutStorageObjectTextArguments struct {
	PutStorageObjectArguments

	Data string `json:"data"`
}

// ProcedureUploadStorageObjectText uploads object in plain text to the storage server. The file content is not encoded to base64 so the input size is smaller than 30%.
func ProcedureUploadStorageObjectText(ctx context.Context, state *types.State, args *PutStorageObjectTextArguments) (common.StorageUploadInfo, error) {
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
func ProcedureUpdateStorageObject(ctx context.Context, state *types.State, args *common.UpdateStorageObjectArguments) (bool, error) {
	request, err := internal.EvalObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return false, err
	}

	if !request.IsValid {
		return false, errPermissionDenied
	}

	if err := state.Storage.UpdateObject(ctx, request.GetBucketArguments(), request.ObjectNamePredicate.GetPrefix(), args.UpdateStorageObjectOptions); err != nil {
		return false, err
	}

	return true, nil
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

	if err := state.Storage.RemoveObject(ctx, request.GetBucketArguments(), request.ObjectNamePredicate.GetPrefix(), args.RemoveStorageObjectOptions); err != nil {
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

	predicate := request.ObjectNamePredicate.CheckPostPredicate
	if !request.ObjectNamePredicate.HasPostPredicate() {
		predicate = nil
	}

	return state.Storage.RemoveObjects(ctx, request.GetBucketArguments(), &common.RemoveStorageObjectsOptions{
		ListStorageObjectsOptions: common.ListStorageObjectsOptions{
			Prefix:     request.ObjectNamePredicate.GetPrefix(),
			Recursive:  args.Recursive,
			StartAfter: args.StartAfter,
		},
		GovernanceBypass: args.GovernanceBypass,
	}, predicate)
}

// ProcedureRemoveIncompleteStorageUpload removes a partially uploaded object.
func ProcedureRemoveIncompleteStorageUpload(ctx context.Context, state *types.State, args *common.RemoveIncompleteUploadArguments) (bool, error) {
	if err := state.Storage.RemoveIncompleteUpload(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}
