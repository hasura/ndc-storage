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
func (m *Manager) MakeBucket(ctx context.Context, clientID *common.StorageClientID, args *common.MakeStorageBucketOptions) error {
	client, bucketName, err := m.GetClientAndBucket(clientID, args.Name)
	if err != nil {
		return err
	}

	args.Name = bucketName

	return client.MakeBucket(ctx, args)
}

// UpdateBucket updates configurations for the bucket.
func (m *Manager) UpdateBucket(ctx context.Context, args *common.UpdateBucketArguments) error {
	if args.UpdateStorageBucketOptions.IsEmpty() {
		return nil
	}

	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	return client.UpdateBucket(ctx, bucketName, args.UpdateStorageBucketOptions)
}

// ListBuckets list all buckets.
func (m *Manager) ListBuckets(ctx context.Context, clientID *common.StorageClientID, options common.BucketOptions) ([]common.StorageBucketInfo, error) {
	client, ok := m.GetClient(clientID)
	if !ok {
		return nil, schema.InternalServerError("client not found", nil)
	}

	return client.ListBuckets(ctx, options)
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
