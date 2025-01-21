package storage

import (
	"context"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

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
func (m *Manager) ListBuckets(ctx context.Context, clientID *common.StorageClientID, options *common.ListStorageBucketsOptions, predicate func(string) bool) (*common.StorageBucketListResults, error) {
	client, ok := m.GetClient(clientID)
	if !ok {
		return nil, schema.InternalServerError("client not found", nil)
	}

	return client.ListBuckets(ctx, options, predicate)
}

// GetBucket gets bucket by name.
func (m *Manager) GetBucket(ctx context.Context, bucketInfo *common.StorageBucketArguments, options common.BucketOptions) (*common.StorageBucket, error) {
	client, bucketName, err := m.GetClientAndBucket(bucketInfo.ClientID, bucketInfo.Bucket)
	if err != nil {
		return nil, err
	}

	return client.GetBucket(ctx, bucketName, options)
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
