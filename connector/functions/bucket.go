package functions

import (
	"context"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/functions/internal"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/types"
)

// ProcedureCreateStorageBucket creates a new bucket.
func ProcedureCreateStorageBucket(ctx context.Context, state *types.State, args *common.MakeStorageBucketArguments) (bool, error) {
	if err := state.Storage.MakeBucket(ctx, args.ClientID, &args.MakeStorageBucketOptions); err != nil {
		return false, err
	}

	return true, nil
}

// FunctionStorageBuckets list all buckets.
func FunctionStorageBuckets(ctx context.Context, state *types.State, args *common.ListStorageBucketArguments) (common.StorageBucketListResults, error) {
	if args.MaxResults <= 0 {
		return common.StorageBucketListResults{}, schema.UnprocessableContentError("maxResults must be larger than 0", nil)
	}

	request, err := internal.EvalObjectPredicate(common.StorageBucketArguments{}, "", args.Where, types.QueryVariablesFromContext(ctx))
	if err != nil {
		return common.StorageBucketListResults{}, err
	}

	if !request.IsValid {
		return common.StorageBucketListResults{
			Buckets: []common.StorageBucket{},
		}, nil
	}

	if err := request.EvalSelection(utils.CommandSelectionFieldFromContext(ctx)); err != nil {
		return common.StorageBucketListResults{}, err
	}

	predicate := request.BucketPredicate.CheckPostPredicate
	if !request.BucketPredicate.HasPostPredicate() {
		predicate = nil
	}

	result, err := state.Storage.ListBuckets(ctx, request.ClientID, &common.ListStorageBucketsOptions{
		Prefix:     request.BucketPredicate.GetPrefix(),
		MaxResults: args.MaxResults,
		StartAfter: args.StartAfter,
		Include: common.BucketIncludeOptions{
			Tags:       request.Include.Tags,
			Versioning: request.Include.Versions,
			Lifecycle:  request.Include.Lifecycle,
			Encryption: request.Include.Encryption,
			ObjectLock: request.Include.ObjectLock,
		},
		NumThreads: state.Concurrency.Query,
	}, predicate)
	if err != nil {
		return common.StorageBucketListResults{}, err
	}

	return *result, nil
}

// FunctionStorageBucket gets a bucket by name.
func FunctionStorageBucket(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (*common.StorageBucket, error) {
	request := internal.PredicateEvaluator{}

	if err := request.EvalSelection(utils.CommandSelectionFieldFromContext(ctx)); err != nil {
		return nil, err
	}

	return state.Storage.GetBucket(ctx, args, common.BucketOptions{
		Include: common.BucketIncludeOptions{
			Tags:       request.Include.Tags,
			Versioning: request.Include.Versions,
			Lifecycle:  request.Include.Lifecycle,
			Encryption: request.Include.Encryption,
			ObjectLock: request.Include.ObjectLock,
		},
		NumThreads: state.Concurrency.Query,
	})
}

// FunctionStorageBucketExists checks if a bucket exists.
func FunctionStorageBucketExists(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (bool, error) {
	return state.Storage.BucketExists(ctx, args)
}

// ProcedureRemoveStorageBucket removes a bucket, bucket should be empty to be successfully removed.
func ProcedureRemoveStorageBucket(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (bool, error) {
	if err := state.Storage.RemoveBucket(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}

// ProcedureUpdateStorageBucket updates the bucket's configuration.
func ProcedureUpdateStorageBucket(ctx context.Context, state *types.State, args *common.UpdateBucketArguments) (bool, error) {
	if err := state.Storage.UpdateBucket(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}
