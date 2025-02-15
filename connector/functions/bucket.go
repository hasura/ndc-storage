package functions

import (
	"context"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/collection"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/types"
)

// ProcedureCreateStorageBucket creates a new bucket.
func ProcedureCreateStorageBucket(ctx context.Context, state *types.State, args *common.MakeStorageBucketArguments) (SuccessResponse, error) {
	if err := state.Storage.MakeBucket(ctx, args.ClientID, &args.MakeStorageBucketOptions); err != nil {
		return SuccessResponse{}, err
	}

	return NewSuccessResponse(), nil
}

// FunctionStorageBucketConnections list all buckets using the relay style.
func FunctionStorageBucketConnections(ctx context.Context, state *types.State, args *common.ListStorageBucketArguments) (StorageConnection[common.StorageBucket], error) {
	if args.First != nil && *args.First <= 0 {
		return StorageConnection[common.StorageBucket]{}, schema.UnprocessableContentError("$first argument must be larger than 0", nil)
	}

	request, err := collection.EvalBucketPredicate(args.ClientID, args.Prefix, args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return StorageConnection[common.StorageBucket]{}, err
	}

	if !request.IsValid {
		return StorageConnection[common.StorageBucket]{
			Edges: []StorageConnectionEdge[common.StorageBucket]{},
		}, nil
	}

	if err := request.EvalSelection(utils.CommandSelectionFieldFromContext(ctx)); err != nil {
		return StorageConnection[common.StorageBucket]{}, err
	}

	predicate := request.BucketPredicate.CheckPostPredicate
	if !request.BucketPredicate.HasPostPredicate() {
		predicate = nil
	}

	buckets, err := state.Storage.ListBuckets(ctx, request.ClientID, &common.ListStorageBucketsOptions{
		Prefix:     request.BucketPredicate.GetPrefix(),
		MaxResults: args.First,
		StartAfter: args.After,
		Include: common.BucketIncludeOptions{
			Tags:       request.Include.Tags,
			Versioning: request.Include.Versions,
			Lifecycle:  request.Include.Lifecycle,
			Encryption: request.Include.Encryption,
			ObjectLock: request.IncludeObjectLock,
		},
		NumThreads: state.Concurrency.Query,
	}, predicate)
	if err != nil {
		return StorageConnection[common.StorageBucket]{}, err
	}

	result := StorageConnection[common.StorageBucket]{
		Edges:    make([]StorageConnectionEdge[common.StorageBucket], len(buckets.Buckets)),
		PageInfo: buckets.PageInfo,
	}

	for i, item := range buckets.Buckets {
		result.Edges[i] = StorageConnectionEdge[common.StorageBucket]{
			Node:   item,
			Cursor: item.Name,
		}
	}

	return result, nil
}

// FunctionStorageBucket gets a bucket by name.
func FunctionStorageBucket(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (*common.StorageBucket, error) {
	request := collection.PredicateEvaluator{}

	if err := request.EvalSelection(utils.CommandSelectionFieldFromContext(ctx)); err != nil {
		return nil, err
	}

	return state.Storage.GetBucket(ctx, args, common.BucketOptions{
		Include: common.BucketIncludeOptions{
			Tags:       request.Include.Tags,
			Versioning: request.Include.Versions,
			Lifecycle:  request.Include.Lifecycle,
			Encryption: request.Include.Encryption,
			ObjectLock: request.IncludeObjectLock,
		},
		NumThreads: state.Concurrency.Query,
	})
}

// FunctionStorageBucketExists checks if a bucket exists.
func FunctionStorageBucketExists(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (ExistsResponse, error) {
	exists, err := state.Storage.BucketExists(ctx, args)
	if err != nil {
		return ExistsResponse{}, err
	}

	return ExistsResponse{Exists: exists}, nil
}

// ProcedureRemoveStorageBucket removes a bucket, bucket should be empty to be successfully removed.
func ProcedureRemoveStorageBucket(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (SuccessResponse, error) {
	if err := state.Storage.RemoveBucket(ctx, args); err != nil {
		return SuccessResponse{}, err
	}

	return NewSuccessResponse(), nil
}

// ProcedureUpdateStorageBucket updates the bucket's configuration.
func ProcedureUpdateStorageBucket(ctx context.Context, state *types.State, args *common.UpdateBucketArguments) (SuccessResponse, error) {
	if err := state.Storage.UpdateBucket(ctx, args); err != nil {
		return SuccessResponse{}, err
	}

	return NewSuccessResponse(), nil
}
