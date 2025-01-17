package storage

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/minio/minio-go/v7/pkg/s3utils"
)

// ListObjects lists objects in a bucket.
func (m *Manager) ListObjects(ctx context.Context, bucketInfo common.StorageBucketArguments, opts *common.ListStorageObjectsOptions, predicate func(string) bool) (*common.StorageObjectListResults, error) {
	client, bucketName, err := m.GetClientAndBucket(bucketInfo.ClientID, bucketInfo.Bucket)
	if err != nil {
		return nil, err
	}

	results, err := client.ListObjects(ctx, bucketName, opts, predicate)
	if err != nil {
		return nil, err
	}

	for i := range results.Objects {
		results.Objects[i].ClientID = string(client.id)
		results.Objects[i].Bucket = bucketName
	}

	return results, nil
}

// ListIncompleteUploads list partially uploaded objects in a bucket.
func (m *Manager) ListIncompleteUploads(ctx context.Context, bucketInfo common.StorageBucketArguments, opts common.ListIncompleteUploadsOptions) ([]common.StorageObjectMultipartInfo, error) {
	client, bucketName, err := m.GetClientAndBucket(bucketInfo.ClientID, bucketInfo.Bucket)
	if err != nil {
		return nil, err
	}

	return client.ListIncompleteUploads(ctx, bucketName, opts)
}

// GetObject returns a stream of the object data. Most of the common errors occur when reading the stream.
func (m *Manager) GetObject(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, opts common.GetStorageObjectOptions) (io.ReadCloser, error) {
	client, bucketName, err := m.GetClientAndBucket(bucketInfo.ClientID, bucketInfo.Bucket)
	if err != nil {
		return nil, err
	}

	return client.GetObject(ctx, bucketName, objectName, opts)
}

// PutObject uploads objects that are less than 128MiB in a single PUT operation. For objects that are greater than 128MiB in size,
// PutObject seamlessly uploads the object as parts of 128MiB or more depending on the actual file size. The max upload size for an object is 5TB.
func (m *Manager) PutObject(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, opts *common.PutStorageObjectOptions, data []byte) (*common.StorageUploadInfo, error) {
	client, bucketName, err := m.GetClientAndBucket(bucketInfo.ClientID, bucketInfo.Bucket)
	if err != nil {
		return nil, err
	}

	result, err := client.PutObject(ctx, bucketName, objectName, opts, bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	result.Bucket = bucketName
	result.ClientID = string(client.id)

	return result, nil
}

// CopyObject creates or replaces an object through server-side copying of an existing object.
// It supports conditional copying, copying a part of an object and server-side encryption of destination and decryption of source.
// To copy multiple source objects into a single destination object see the ComposeObject API.
func (m *Manager) CopyObject(ctx context.Context, args *common.CopyStorageObjectArguments) (*common.StorageUploadInfo, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Dest.Bucket)
	if err != nil {
		return nil, err
	}

	args.Dest.Bucket = bucketName

	if args.Source.Bucket == "" {
		args.Source.Bucket = client.defaultBucket
	}

	result, err := client.CopyObject(ctx, args.Dest, args.Source)
	if err != nil {
		return nil, err
	}

	result.ClientID = string(client.id)

	return result, nil
}

// ComposeObject creates an object by concatenating a list of source objects using server-side copying.
func (m *Manager) ComposeObject(ctx context.Context, args *common.ComposeStorageObjectArguments) (*common.StorageUploadInfo, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Dest.Bucket)
	if err != nil {
		return nil, err
	}

	args.Dest.Bucket = bucketName
	srcs := make([]common.StorageCopySrcOptions, len(args.Sources))

	for i, src := range args.Sources {
		if src.Bucket == "" {
			src.Bucket = client.defaultBucket
		}

		srcs[i] = src
	}

	result, err := client.ComposeObject(ctx, args.Dest, srcs)
	if err != nil {
		return nil, err
	}

	result.ClientID = string(client.id)

	return result, nil
}

// StatObject fetches metadata of an object.
func (m *Manager) StatObject(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, opts common.GetStorageObjectOptions) (*common.StorageObject, error) {
	client, bucketName, err := m.GetClientAndBucket(bucketInfo.ClientID, bucketInfo.Bucket)
	if err != nil {
		return nil, err
	}

	result, err := client.StatObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return nil, err
	}

	result.ClientID = string(client.id)
	result.Bucket = bucketName

	return result, nil
}

// RemoveObject removes an object with some specified options.
func (m *Manager) RemoveObject(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, opts common.RemoveStorageObjectOptions) error {
	client, bucketName, err := m.GetClientAndBucket(bucketInfo.ClientID, bucketInfo.Bucket)
	if err != nil {
		return err
	}

	return client.RemoveObject(ctx, bucketName, objectName, opts)
}

// SetObjectRetention applies object retention lock onto an object.
func (m *Manager) SetObjectRetention(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, opts common.SetStorageObjectRetentionOptions) error {
	client, bucketName, err := m.GetClientAndBucket(bucketInfo.ClientID, bucketInfo.Bucket)
	if err != nil {
		return err
	}

	return client.SetObjectRetention(ctx, bucketName, objectName, opts)
}

// RemoveObjects remove a list of objects obtained from an input channel. The call sends a delete request to the server up to 1000 objects at a time.
// The errors observed are sent over the error channel.
func (m *Manager) RemoveObjects(ctx context.Context, bucketInfo common.StorageBucketArguments, opts *common.RemoveStorageObjectsOptions, predicate func(string) bool) ([]common.RemoveStorageObjectError, error) {
	client, bucketName, err := m.GetClientAndBucket(bucketInfo.ClientID, bucketInfo.Bucket)
	if err != nil {
		return nil, err
	}

	return client.RemoveObjects(ctx, bucketName, opts, predicate), nil
}

// SetObjectLegalHold applies legal-hold onto an object.
func (m *Manager) SetObjectLegalHold(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, opts common.SetStorageObjectLegalHoldOptions) error {
	client, bucketName, err := m.GetClientAndBucket(bucketInfo.ClientID, bucketInfo.Bucket)
	if err != nil {
		return err
	}

	return client.SetObjectLegalHold(ctx, bucketName, objectName, opts)
}

// PutObjectTagging sets new object Tags to the given object, replaces/overwrites any existing tags.
func (m *Manager) SetObjectTags(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, opts common.SetStorageObjectTagsOptions) error {
	client, bucketName, err := m.GetClientAndBucket(bucketInfo.ClientID, bucketInfo.Bucket)
	if err != nil {
		return err
	}

	return client.SetObjectTags(ctx, bucketName, objectName, opts)
}

// RemoveIncompleteUpload removes a partially uploaded object.
func (m *Manager) RemoveIncompleteUpload(ctx context.Context, args *common.RemoveIncompleteUploadArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	return client.RemoveIncompleteUpload(ctx, bucketName, args.Object)
}

// PresignedGetObject generates a presigned URL for HTTP GET operations. Browsers/Mobile clients may point to this URL to directly download objects even if the bucket is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The maximum expiry is 604800 seconds (i.e. 7 days) and minimum is 1 second.
func (m *Manager) PresignedGetObject(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, opts common.PresignedGetStorageObjectOptions) (*common.PresignedURLResponse, error) {
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	client, bucketName, err := m.GetClientAndBucket(bucketInfo.ClientID, bucketInfo.Bucket)
	if err != nil {
		return nil, err
	}

	var exp time.Duration

	if opts.Expiry != nil {
		exp = opts.Expiry.Duration
	} else if client.defaultPresignedExpiry != nil {
		exp = *client.defaultPresignedExpiry
	}

	if exp == 0 {
		return nil, schema.UnprocessableContentError("expiry is required and must be larger than 0", nil)
	}

	rawURL, err := client.PresignedGetObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return nil, err
	}

	return &common.PresignedURLResponse{
		URL:       rawURL,
		ExpiredAt: FormatTimestamp(time.Now().Add(opts.Expiry.Duration)),
	}, nil
}

// PresignedPutObject generates a presigned URL for HTTP PUT operations.
// Browsers/Mobile clients may point to this URL to upload objects directly to a bucket even if it is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The default expiry is set to 7 days.
func (m *Manager) PresignedPutObject(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, expiry *scalar.Duration) (*common.PresignedURLResponse, error) {
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	client, bucketName, err := m.GetClientAndBucket(bucketInfo.ClientID, bucketInfo.Bucket)
	if err != nil {
		return nil, err
	}

	var exp time.Duration

	if expiry != nil {
		exp = expiry.Duration
	} else if client.defaultPresignedExpiry != nil {
		exp = *client.defaultPresignedExpiry
	}

	if exp == 0 {
		return nil, schema.UnprocessableContentError("expiry is required and must be larger than 0", nil)
	}

	rawURL, err := client.PresignedPutObject(ctx, bucketName, objectName, exp)
	if err != nil {
		return nil, err
	}

	return &common.PresignedURLResponse{
		URL:       rawURL,
		ExpiredAt: FormatTimestamp(time.Now().Add(exp)),
	}, nil
}
