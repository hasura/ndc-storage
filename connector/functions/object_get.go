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
func FunctionStorageObjects(ctx context.Context, state *types.State, args *common.ListStorageObjectsArguments) ([]common.StorageObject, error) {
	if args.MaxResults != nil && *args.MaxResults <= 0 {
		return nil, schema.UnprocessableContentError("maxResults must be larger than 0", nil)
	}

	request, err := internal.EvalObjectPredicate(args.StorageBucketArguments, "", args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return nil, err
	}

	if !request.IsValid {
		return []common.StorageObject{}, nil
	}

	if err := request.EvalSelection(utils.CommandSelectionFieldFromContext(ctx)); err != nil {
		return nil, err
	}

	if !request.HasPostPredicate() && args.MaxResults != nil {
		request.Options.MaxResults = *args.MaxResults
	}

	if args.StartAfter != nil {
		request.Options.StartAfter = *args.StartAfter
	}

	request.Options.Recursive = args.Recursive

	objects, err := state.Storage.ListObjects(ctx, request.StorageBucketArguments, &request.Options)
	if err != nil {
		return nil, err
	}

	var filtered []common.StorageObject

	if request.HasPostPredicate() {
		for _, item := range objects {
			if request.CheckPostObjectPredicate(item) {
				filtered = append(filtered, item)
			}
		}
	} else {
		filtered = objects
	}

	if args.MaxResults != nil {
		limit := len(filtered)
		if *args.MaxResults < limit {
			limit = *args.MaxResults
		}

		filtered = filtered[:limit]
	}

	return filtered, nil
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

	return state.Storage.StatObject(ctx, request.StorageBucketArguments, request.Options.Prefix, args.GetStorageObjectOptions)
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

	return state.Storage.GetObject(ctx, request.StorageBucketArguments, request.Options.Prefix, args.GetStorageObjectOptions)
}
