package functions

import (
	"context"

	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/collection"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/types"
)

// PutStorageObjectArguments represents input arguments of the PutObject method.
type PutStorageObjectArguments struct {
	common.StorageBucketArguments

	Object  string                         `json:"object"`
	Options common.PutStorageObjectOptions `json:"options,omitempty"`
	Where   schema.Expression              `json:"where"             ndc:"predicate=StorageObjectSimple"`
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
	request, err := collection.EvalCollectionObjectPredicate(args.StorageBucketArguments, args.Object, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return common.StorageUploadInfo{}, err
	}

	if !request.IsValid {
		return common.StorageUploadInfo{}, schema.ForbiddenError("permission dennied", nil)
	}

	result, err := state.Storage.PutObject(ctx, request.StorageBucketArguments, request.Options.Prefix, &args.Options, data)
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
