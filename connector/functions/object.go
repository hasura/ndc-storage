package functions

import (
	"context"
	"io"

	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/types"
)

// FunctionStorageIncompleteUploads list partially uploaded objects in a bucket.
func FunctionStorageIncompleteUploads(ctx context.Context, state *types.State, args *common.ListIncompleteUploadsArguments) ([]common.StorageObjectMultipartInfo, error) {
	return state.Storage.ListIncompleteUploads(ctx, args)
}

// FunctionDownloadStorageObject returns a stream of the object data. Most of the common errors occur when reading the stream.
func FunctionDownloadStorageObject(ctx context.Context, state *types.State, args *common.GetStorageObjectOptions) (*scalar.Bytes, error) {
	reader, err := state.Storage.GetObject(ctx, args)
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
func FunctionDownloadStorageObjectText(ctx context.Context, state *types.State, args *common.GetStorageObjectOptions) (*string, error) {
	reader, err := state.Storage.GetObject(ctx, args)
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

// PutStorageObjectArguments represents input arguments of the PutObject method.
type PutStorageObjectBase64Arguments struct {
	common.PutStorageObjectArguments

	Data scalar.Bytes `json:"data"`
}

// ProcedureUploadStorageObject uploads object that are less than 128MiB in a single PUT operation. For objects that are greater than 128MiB in size,
// PutObject seamlessly uploads the object as parts of 128MiB or more depending on the actual file size. The max upload size for an object is 5TB.
func ProcedureUploadStorageObject(ctx context.Context, state *types.State, args *PutStorageObjectBase64Arguments) (common.StorageUploadInfo, error) {
	result, err := state.Storage.PutObject(ctx, &args.PutStorageObjectArguments, args.Data.Bytes())
	if err != nil {
		return common.StorageUploadInfo{}, err
	}

	return *result, nil
}

// PutStorageObjectTextArguments represents input arguments of the PutStorageObjectText method.
type PutStorageObjectTextArguments struct {
	common.PutStorageObjectArguments

	Data string `json:"data"`
}

// ProcedureUploadStorageObjectText uploads object in plain text to the storage server. The file content is not encoded to base64 so the input size is smaller than 30%.
func ProcedureUploadStorageObjectText(ctx context.Context, state *types.State, args *PutStorageObjectTextArguments) (common.StorageUploadInfo, error) {
	result, err := state.Storage.PutObject(ctx, &args.PutStorageObjectArguments, []byte(args.Data))
	if err != nil {
		return common.StorageUploadInfo{}, err
	}

	return *result, nil
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

// FunctionStorageObject fetches metadata of an object.
func FunctionStorageObject(ctx context.Context, state *types.State, args *common.GetStorageObjectOptions) (*common.StorageObject, error) {
	return state.Storage.StatObject(ctx, args)
}

// ProcedureRemoveStorageObject removes an object with some specified options.
func ProcedureRemoveStorageObject(ctx context.Context, state *types.State, args *common.RemoveStorageObjectOptions) (bool, error) {
	if err := state.Storage.RemoveObject(ctx, args); err != nil {
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
func ProcedureRemoveStorageObjects(ctx context.Context, state *types.State, args *common.RemoveStorageObjectsOptions) ([]common.RemoveStorageObjectError, error) {
	return state.Storage.RemoveObjects(ctx, args)
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
func FunctionStoragePresignedDownloadUrl(ctx context.Context, state *types.State, args *common.PresignedGetStorageObjectArguments) (common.PresignedURLResponse, error) {
	return state.Storage.PresignedGetObject(ctx, args)
}

// FunctionStoragePresignedUploadUrl generates a presigned URL for HTTP PUT operations.
// Browsers/Mobile clients may point to this URL to upload objects directly to a bucket even if it is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The default expiry is set to 7 days.
func FunctionStoragePresignedUploadUrl(ctx context.Context, state *types.State, args *common.PresignedPutStorageObjectArguments) (common.PresignedURLResponse, error) {
	return state.Storage.PresignedPutObject(ctx, args)
}

// FunctionStoragePresignedHeadUrl generates a presigned URL for HTTP HEAD operations.
// Browsers/Mobile clients may point to this URL to directly get metadata from objects even if the bucket is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational. The default expiry is set to 7 days.
func FunctionStoragePresignedHeadUrl(ctx context.Context, state *types.State, args *common.PresignedGetStorageObjectArguments) (common.PresignedURLResponse, error) {
	return state.Storage.PresignedHeadObject(ctx, args)
}
