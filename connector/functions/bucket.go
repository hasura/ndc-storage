package functions

import (
	"context"

	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/types"
)

// ProcedureCreateStorageBucket creates a new bucket.
func ProcedureCreateStorageBucket(ctx context.Context, state *types.State, options *common.MakeStorageBucketOptions) (bool, error) {
	if err := state.Storage.MakeBucket(ctx, options); err != nil {
		return false, err
	}

	return true, nil
}

// FunctionStorageBuckets list all buckets.
func FunctionStorageBuckets(ctx context.Context, state *types.State, args *common.ListStorageBucketArguments) ([]common.StorageBucketInfo, error) {
	return state.Storage.ListBuckets(ctx, args)
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

// ProcedureSetStorageBucketTags sets tags to a bucket.
func ProcedureSetStorageBucketTags(ctx context.Context, state *types.State, args *common.SetStorageBucketTaggingArguments) (bool, error) {
	if err := state.Storage.SetBucketTagging(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}

// FunctionStorageBucketTags gets tags of a bucket.
func FunctionStorageBucketTags(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (map[string]string, error) {
	return state.Storage.GetBucketTagging(ctx, args)
}

// ProcedureRemoveStorageBucketTags removes all tags on a bucket.
func ProcedureRemoveStorageBucketTags(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (bool, error) {
	if err := state.Storage.RemoveBucketTagging(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}

// FunctionStorageBucketPolicy gets access permissions on a bucket or a prefix.
func FunctionStorageBucketPolicy(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (string, error) {
	return state.Storage.GetBucketPolicy(ctx, args)
}

// FunctionStorageBucketNotification gets notification configuration on a bucket.
func FunctionStorageBucketNotification(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (*common.NotificationConfig, error) {
	return state.Storage.GetBucketNotification(ctx, args)
}

// ProcedureSetStorageBucketNotification sets a new notification configuration on a bucket.
func ProcedureSetStorageBucketNotification(ctx context.Context, state *types.State, args *common.SetBucketNotificationArguments) (bool, error) {
	if err := state.Storage.SetBucketNotification(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}

// ProcedureSetStorageBucketLifecycle sets lifecycle on bucket or an object prefix.
func ProcedureSetStorageBucketLifecycle(ctx context.Context, state *types.State, args *common.SetStorageBucketLifecycleArguments) (bool, error) {
	err := state.Storage.SetBucketLifecycle(ctx, args)
	if err != nil {
		return false, err
	}

	return true, nil
}

// FunctionStorageBucketLifecycle gets lifecycle on a bucket or a prefix.
func FunctionStorageBucketLifecycle(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (*common.BucketLifecycleConfiguration, error) {
	return state.Storage.GetBucketLifecycle(ctx, args)
}

// ProcedureSetStorageBucketEncryption sets default encryption configuration on a bucket.
func ProcedureSetStorageBucketEncryption(ctx context.Context, state *types.State, args *common.SetStorageBucketEncryptionArguments) (bool, error) {
	err := state.Storage.SetBucketEncryption(ctx, args)
	if err != nil {
		return false, err
	}

	return true, nil
}

// FunctionStorageBucketEncryption gets default encryption configuration set on a bucket.
func FunctionStorageBucketEncryption(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (*common.ServerSideEncryptionConfiguration, error) {
	return state.Storage.GetBucketEncryption(ctx, args)
}

// ProcedureSetObjectLockConfig sets object lock configuration in given bucket. mode, validity and unit are either all set or all nil.
func ProcedureSetStorageObjectLockConfig(ctx context.Context, state *types.State, args *common.SetStorageObjectLockArguments) (bool, error) {
	err := state.Storage.SetObjectLockConfig(ctx, args)
	if err != nil {
		return false, err
	}

	return true, nil
}

// FunctionStorageObjectLockConfig gets object lock configuration of given bucket.
func FunctionStorageObjectLockConfig(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (*common.StorageObjectLockConfig, error) {
	return state.Storage.GetObjectLockConfig(ctx, args)
}

// ProcedureEnableStorageBucketVersioning enables bucket versioning support.
func ProcedureEnableStorageBucketVersioning(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (bool, error) {
	if err := state.Storage.EnableVersioning(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}

// ProcedureSuspendStorageBucketVersioning disables bucket versioning support.
func ProcedureSuspendStorageBucketVersioning(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (bool, error) {
	if err := state.Storage.SuspendVersioning(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}

// FunctionStorageBucketVersioning gets versioning configuration set on a bucket.
func FunctionStorageBucketVersioning(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (*common.StorageBucketVersioningConfiguration, error) {
	return state.Storage.GetBucketVersioning(ctx, args)
}

// ProcedureSetStorageBucketReplication sets replication configuration on a bucket. Role can be obtained by first defining the replication target on MinIO
// to associate the source and destination buckets for replication with the replication endpoint.
func ProcedureSetStorageBucketReplication(ctx context.Context, state *types.State, args *common.SetStorageBucketReplicationArguments) (bool, error) {
	if err := state.Storage.SetBucketReplication(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}

// FunctionGetBucketReplication gets current replication config on a bucket.
func FunctionStorageBucketReplication(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (*common.StorageReplicationConfig, error) {
	return state.Storage.GetBucketReplication(ctx, args)
}

// RemoveBucketReplication removes replication configuration on a bucket.
func ProcedureRemoveStorageBucketReplication(ctx context.Context, state *types.State, args *common.StorageBucketArguments) (bool, error) {
	if err := state.Storage.RemoveBucketReplication(ctx, args); err != nil {
		return false, err
	}

	return true, nil
}
