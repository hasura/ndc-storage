package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/minio/minio-go/v7/pkg/s3utils"
)

// ListObjects lists objects in a bucket.
func (m *Manager) ListObjects(ctx context.Context, args *common.ListStorageObjectsOptions) ([]common.StorageObject, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	args.Bucket = bucketName

	return client.ListObjects(ctx, args)
}

// ListIncompleteUploads list partially uploaded objects in a bucket.
func (m *Manager) ListIncompleteUploads(ctx context.Context, args *common.ListIncompleteUploadsArguments) ([]common.StorageObjectMultipartInfo, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	args.Bucket = bucketName
	args.Prefix = normalizeObjectName(args.Prefix)

	return client.ListIncompleteUploads(ctx, args)
}

// GetObject returns a stream of the object data. Most of the common errors occur when reading the stream.
func (m *Manager) GetObject(ctx context.Context, args *common.GetStorageObjectOptions) (io.ReadCloser, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	return client.GetObject(ctx, args)
}

// PutObject uploads objects that are less than 128MiB in a single PUT operation. For objects that are greater than 128MiB in size,
// PutObject seamlessly uploads the object as parts of 128MiB or more depending on the actual file size. The max upload size for an object is 5TB.
func (m *Manager) PutObject(ctx context.Context, args *common.PutStorageObjectArguments, data []byte) (*common.StorageUploadInfo, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	return client.PutObject(ctx, args, bytes.NewReader(data), int64(len(data)))
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
	args.Dest.Object = normalizeObjectName(args.Dest.Object)

	if args.Source.Bucket == "" {
		args.Source.Bucket = client.defaultBucket
		args.Source.Object = normalizeObjectName(args.Source.Object)
	}

	return client.CopyObject(ctx, args.Dest, args.Source)
}

// ComposeObject creates an object by concatenating a list of source objects using server-side copying.
func (m *Manager) ComposeObject(ctx context.Context, args *common.ComposeStorageObjectArguments) (*common.StorageUploadInfo, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Dest.Bucket)
	if err != nil {
		return nil, err
	}

	args.Dest.Bucket = bucketName
	args.Dest.Object = normalizeObjectName(args.Dest.Object)
	srcs := make([]common.StorageCopySrcOptions, len(args.Sources))

	for i, src := range args.Sources {
		if src.Bucket == "" {
			src.Bucket = client.defaultBucket
		}

		src.Object = normalizeObjectName(src.Object)

		srcs[i] = src
	}

	return client.ComposeObject(ctx, args.Dest, srcs)
}

// StatObject fetches metadata of an object.
func (m *Manager) StatObject(ctx context.Context, args *common.GetStorageObjectOptions) (*common.StorageObject, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	return client.StatObject(ctx, args)
}

// RemoveObject removes an object with some specified options.
func (m *Manager) RemoveObject(ctx context.Context, args *common.RemoveStorageObjectOptions) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	return client.RemoveObject(ctx, args)
}

// PutObjectRetention applies object retention lock onto an object.
func (m *Manager) PutObjectRetention(ctx context.Context, args *common.PutStorageObjectRetentionOptions) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	return client.PutObjectRetention(ctx, args)
}

// RemoveObjects remove a list of objects obtained from an input channel. The call sends a delete request to the server up to 1000 objects at a time.
// The errors observed are sent over the error channel.
func (m *Manager) RemoveObjects(ctx context.Context, args *common.RemoveStorageObjectsOptions) ([]common.RemoveStorageObjectError, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	args.Bucket = bucketName
	args.Prefix = normalizeObjectName(args.Prefix)

	return client.RemoveObjects(ctx, args), nil
}

// PutObjectLegalHold applies legal-hold onto an object.
func (m *Manager) PutObjectLegalHold(ctx context.Context, args *common.PutStorageObjectLegalHoldOptions) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	return client.PutObjectLegalHold(ctx, args)
}

// GetObjectLegalHold returns legal-hold status on a given object.
func (m *Manager) GetObjectLegalHold(ctx context.Context, args *common.GetStorageObjectLegalHoldOptions) (common.StorageLegalHoldStatus, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return "", err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	return client.GetObjectLegalHold(ctx, args)
}

// PutObjectTagging sets new object Tags to the given object, replaces/overwrites any existing tags.
func (m *Manager) PutObjectTagging(ctx context.Context, args *common.PutStorageObjectTaggingOptions) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	return client.PutObjectTagging(ctx, args)
}

// GetObjectTagging fetches Object Tags from the given object.
func (m *Manager) GetObjectTagging(ctx context.Context, args *common.StorageObjectTaggingOptions) (map[string]string, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	return client.GetObjectTagging(ctx, args)
}

// RemoveObjectTagging removes Object Tags from the given object.
func (m *Manager) RemoveObjectTagging(ctx context.Context, args *common.StorageObjectTaggingOptions) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	return client.RemoveObjectTagging(ctx, args)
}

// GetObjectAttributes returns a stream of the object data. Most of the common errors occur when reading the stream.
func (m *Manager) GetObjectAttributes(ctx context.Context, args *common.StorageObjectAttributesOptions) (*common.StorageObjectAttributes, error) {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return nil, err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	return client.GetObjectAttributes(ctx, args)
}

// RemoveIncompleteUpload removes a partially uploaded object.
func (m *Manager) RemoveIncompleteUpload(ctx context.Context, args *common.RemoveIncompleteUploadArguments) error {
	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	return client.RemoveIncompleteUpload(ctx, args)
}

// PresignedGetObject generates a presigned URL for HTTP GET operations. Browsers/Mobile clients may point to this URL to directly download objects even if the bucket is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The maximum expiry is 604800 seconds (i.e. 7 days) and minimum is 1 second.
func (m *Manager) PresignedGetObject(ctx context.Context, args *common.PresignedGetStorageObjectArguments) (common.PresignedURLResponse, error) {
	if err := s3utils.CheckValidObjectName(args.Object); err != nil {
		return common.PresignedURLResponse{}, schema.UnprocessableContentError(err.Error(), nil)
	}

	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return common.PresignedURLResponse{}, err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	if args.Expiry == nil {
		if client.defaultPresignedExpiry != nil {
			expiry := scalar.NewDuration(*client.defaultPresignedExpiry)
			args.Expiry = &expiry
		}

		return common.PresignedURLResponse{}, schema.UnprocessableContentError("expiry is required", nil)
	}

	rawURL, err := client.PresignedGetObject(ctx, args)
	if err != nil {
		return common.PresignedURLResponse{}, err
	}

	return common.PresignedURLResponse{
		URL:       rawURL.String(),
		ExpiredAt: FormatTimestamp(time.Now().Add(args.Expiry.Duration)),
	}, nil
}

// PresignedPutObject generates a presigned URL for HTTP PUT operations.
// Browsers/Mobile clients may point to this URL to upload objects directly to a bucket even if it is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The default expiry is set to 7 days.
func (m *Manager) PresignedPutObject(ctx context.Context, args *common.PresignedPutStorageObjectArguments) (common.PresignedURLResponse, error) {
	if err := s3utils.CheckValidObjectName(args.Object); err != nil {
		return common.PresignedURLResponse{}, schema.UnprocessableContentError(err.Error(), nil)
	}

	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return common.PresignedURLResponse{}, err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	if args.Expiry == nil {
		if client.defaultPresignedExpiry != nil {
			expiry := scalar.NewDuration(*client.defaultPresignedExpiry)
			args.Expiry = &expiry
		}

		return common.PresignedURLResponse{}, schema.UnprocessableContentError("expiry is required", nil)
	}

	rawURL, err := client.PresignedPutObject(ctx, args)
	if err != nil {
		return common.PresignedURLResponse{}, err
	}

	return common.PresignedURLResponse{
		URL:       rawURL.String(),
		ExpiredAt: FormatTimestamp(time.Now().Add(args.Expiry.Duration)),
	}, nil
}

// PresignedHeadObject generates a presigned URL for HTTP HEAD operations.
// Browsers/Mobile clients may point to this URL to directly get metadata from objects even if the bucket is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational. The default expiry is set to 7 days.
func (m *Manager) PresignedHeadObject(ctx context.Context, args *common.PresignedGetStorageObjectArguments) (common.PresignedURLResponse, error) {
	if err := s3utils.CheckValidObjectName(args.Object); err != nil {
		return common.PresignedURLResponse{}, schema.UnprocessableContentError(err.Error(), nil)
	}

	client, bucketName, err := m.GetClientAndBucket(args.ClientID, args.Bucket)
	if err != nil {
		return common.PresignedURLResponse{}, err
	}

	args.Bucket = bucketName
	args.Object = normalizeObjectName(args.Object)

	if args.Expiry == nil {
		if client.defaultPresignedExpiry != nil {
			expiry := scalar.NewDuration(*client.defaultPresignedExpiry)
			args.Expiry = &expiry
		}

		return common.PresignedURLResponse{}, schema.UnprocessableContentError("expiry is required", nil)
	}

	rawURL, err := client.PresignedHeadObject(ctx, args)
	if err != nil {
		return common.PresignedURLResponse{}, err
	}

	return common.PresignedURLResponse{
		URL:       rawURL.String(),
		ExpiredAt: FormatTimestamp(time.Now().Add(args.Expiry.Duration)),
	}, nil
}

func normalizeObjectName(objectName string) string {
	// replace Unix-compatible backslashes in the file path when run on Windows OS
	return strings.ReplaceAll(objectName, "\\", "/")
}
