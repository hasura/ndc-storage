package functions

import (
	"context"

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
func FunctionStorageBuckets(ctx context.Context, state *types.State, args *common.ListStorageBucketArguments) ([]common.StorageBucketInfo, error) {
	request := internal.ObjectPredicate{}

	if err := request.EvalSelection(utils.CommandSelectionFieldFromContext(ctx)); err != nil {
		return nil, err
	}

	return state.Storage.ListBuckets(ctx, args.ClientID, common.BucketOptions{
		Prefix:            args.Prefix,
		IncludeTags:       request.Include.Tags,
		IncludeVersioning: request.Include.Versions,
		IncludeLifecycle:  request.Include.Lifecycle,
		IncludeEncryption: request.Include.Encryption,
		IncludeObjectLock: request.Include.ObjectLock,
		NumThreads:        state.Concurrency.Query,
	})
}

// FunctionStorageBucket gets a bucket by name.
func FunctionStorageBucket(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (*common.StorageBucketInfo, error) {
	request := internal.ObjectPredicate{}

	if err := request.EvalSelection(utils.CommandSelectionFieldFromContext(ctx)); err != nil {
		return nil, err
	}

	return state.Storage.GetBucket(ctx, args, common.BucketOptions{
		IncludeTags:       request.Include.Tags,
		IncludeVersioning: request.Include.Versions,
		IncludeLifecycle:  request.Include.Lifecycle,
		IncludeEncryption: request.Include.Encryption,
		IncludeObjectLock: request.Include.ObjectLock,
		NumThreads:        state.Concurrency.Query,
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
