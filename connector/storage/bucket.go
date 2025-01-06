package storage

import (
	"context"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

// GetBucketPolicy gets access permissions on a bucket or a prefix.
func (m *Manager) GetBucketPolicy(ctx context.Context, args *common.StorageBucketArguments) (string, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return "", err
	}

	return client.GetBucketPolicy(ctx, bucketName)
}

// GetBucketNotification gets notification configuration on a bucket.
func (m *Manager) GetBucketNotification(ctx context.Context, args *common.StorageBucketArguments) (*common.NotificationConfig, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	return client.GetBucketNotification(ctx, bucketName)
}

// SetBucketNotification sets a new bucket notification on a bucket.
func (m *Manager) SetBucketNotification(ctx context.Context, args *common.SetBucketNotificationArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	return client.SetBucketNotification(ctx, bucketName, args.NotificationConfig)
}

// Remove all configured bucket notifications on a bucket.
func (m *Manager) RemoveAllBucketNotification(ctx context.Context, args *common.StorageBucketArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	return client.RemoveAllBucketNotification(ctx, bucketName)
}

// MakeBucket creates a new bucket.
func (m *Manager) MakeBucket(ctx context.Context, args *common.MakeStorageBucketOptions) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Name)
	if err != nil {
		return err
	}

	args.Name = bucketName

	return client.MakeBucket(ctx, args)
}

// ListBuckets list all buckets.
func (m *Manager) ListBuckets(ctx context.Context, args *common.ListStorageBucketArguments) ([]common.StorageBucketInfo, error) {
	client, ok := m.GetClient(&args.ClientID)
	if !ok {
		return nil, schema.InternalServerError("client not found", nil)
	}

	return client.ListBuckets(ctx)
}

// BucketExists checks if a bucket exists.
func (m *Manager) BucketExists(ctx context.Context, args *common.StorageBucketArguments) (bool, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return false, err
	}

	return client.BucketExists(ctx, bucketName)
}

// RemoveBucket removes a bucket, bucket should be empty to be successfully removed.
func (m *Manager) RemoveBucket(ctx context.Context, args *common.StorageBucketArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	return client.RemoveBucket(ctx, bucketName)
}

// SetBucketTagging sets tags to a bucket.
func (m *Manager) SetBucketTagging(ctx context.Context, args *common.SetStorageBucketTaggingArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	args.Bucket = bucketName

	return client.SetBucketTagging(ctx, args)
}

// GetBucketTagging gets tags of a bucket.
func (m *Manager) GetBucketTagging(ctx context.Context, args *common.StorageBucketArguments) (map[string]string, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	return client.GetBucketTagging(ctx, bucketName)
}

// RemoveBucketTagging removes all tags on a bucket.
func (m *Manager) RemoveBucketTagging(ctx context.Context, args *common.StorageBucketArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	return client.RemoveBucketTagging(ctx, bucketName)
}

// SetBucketLifecycle sets lifecycle on bucket or an object prefix.
func (m *Manager) SetBucketLifecycle(ctx context.Context, args *common.SetStorageBucketLifecycleArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	return client.SetBucketLifecycle(ctx, bucketName, args.BucketLifecycleConfiguration)
}

// GetBucketLifecycle gets lifecycle on a bucket or a prefix.
func (m *Manager) GetBucketLifecycle(ctx context.Context, args *common.StorageBucketArguments) (*common.BucketLifecycleConfiguration, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	return client.GetBucketLifecycle(ctx, bucketName)
}

// SetBucketEncryption sets default encryption configuration on a bucket.
func (m *Manager) SetBucketEncryption(ctx context.Context, args *common.SetStorageBucketEncryptionArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	return client.SetBucketEncryption(ctx, bucketName, args.ServerSideEncryptionConfiguration)
}

// GetBucketEncryption gets default encryption configuration set on a bucket.
func (m *Manager) GetBucketEncryption(ctx context.Context, args *common.StorageBucketArguments) (*common.ServerSideEncryptionConfiguration, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	return client.GetBucketEncryption(ctx, bucketName)
}

// RemoveBucketEncryption removes default encryption configuration set on a bucket.
func (m *Manager) RemoveBucketEncryption(ctx context.Context, args *common.StorageBucketArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	return client.RemoveBucketEncryption(ctx, bucketName)
}

// SetObjectLockConfig sets object lock configuration in given bucket. mode, validity and unit are either all set or all nil.
func (m *Manager) SetObjectLockConfig(ctx context.Context, args *common.SetStorageObjectLockArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	return client.SetObjectLockConfig(ctx, bucketName, args.SetStorageObjectLockConfig)
}

// GetObjectLockConfig gets object lock configuration of given bucket.
func (m *Manager) GetObjectLockConfig(ctx context.Context, args *common.StorageBucketArguments) (*common.StorageObjectLockConfig, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	return client.GetObjectLockConfig(ctx, bucketName)
}

// EnableVersioning enables bucket versioning support.
func (m *Manager) EnableVersioning(ctx context.Context, args *common.StorageBucketArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	return client.EnableVersioning(ctx, bucketName)
}

// SuspendVersioning disables bucket versioning support.
func (m *Manager) SuspendVersioning(ctx context.Context, args *common.StorageBucketArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	return client.SuspendVersioning(ctx, bucketName)
}

// GetBucketVersioning gets versioning configuration set on a bucket.
func (m *Manager) GetBucketVersioning(ctx context.Context, args *common.StorageBucketArguments) (*common.StorageBucketVersioningConfiguration, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	return client.GetBucketVersioning(ctx, bucketName)
}

// SetBucketReplication sets replication configuration on a bucket. Role can be obtained by first defining the replication target on MinIO
// to associate the source and destination buckets for replication with the replication endpoint.
func (m *Manager) SetBucketReplication(ctx context.Context, args *common.SetStorageBucketReplicationArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	return client.SetBucketReplication(ctx, bucketName, args.StorageReplicationConfig)
}

// GetBucketReplication gets current replication config on a bucket.
func (m *Manager) GetBucketReplication(ctx context.Context, args *common.StorageBucketArguments) (*common.StorageReplicationConfig, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	return client.GetBucketReplication(ctx, bucketName)
}

// RemoveBucketReplication removes replication configuration on a bucket.
func (m *Manager) RemoveBucketReplication(ctx context.Context, args *common.StorageBucketArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	return client.RemoveBucketReplication(ctx, bucketName)
}
