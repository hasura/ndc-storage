package azblob

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// ListObjects list objects in a bucket.
func (c *Client) ListObjects(ctx context.Context, bucketName string, opts *common.ListStorageObjectsOptions) ([]common.StorageObject, error) {
	ctx, span := c.startOtelSpan(ctx, "ListObjects", bucketName)
	defer span.End()

	options := &container.ListBlobsFlatOptions{
		Include: container.ListBlobsInclude{
			Versions:         opts.WithVersions,
			Metadata:         opts.WithMetadata,
			Tags:             opts.WithTags,
			UncommittedBlobs: false,
		},
	}

	if opts.Prefix != "" {
		options.Prefix = &opts.Prefix
	}

	if opts.MaxResults > 0 {
		maxResults := int32(opts.MaxResults)
		options.MaxResults = &maxResults
	}

	if opts.StartAfter != "" {
		options.Marker = &opts.StartAfter
	}

	objects := make([]common.StorageObject, 0)
	pager := c.client.NewListBlobsFlatPager(bucketName, options)

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, err
		}

		for _, item := range resp.Segment.BlobItems {
			objects = append(objects, serializeObjectInfo(item))
		}
	}

	span.SetAttributes(attribute.Int("storage.object_count", len(objects)))

	return objects, nil
}

// ListIncompleteUploads list partially uploaded objects in a bucket.
func (c *Client) ListIncompleteUploads(ctx context.Context, args *common.ListIncompleteUploadsArguments) ([]common.StorageObjectMultipartInfo, error) {
	return nil, errNotSupported
}

// RemoveIncompleteUpload removes a partially uploaded object.
func (c *Client) RemoveIncompleteUpload(ctx context.Context, bucketName string, objectName string) error {
	return c.removeObject(ctx, "RemoveIncompleteUpload", bucketName, objectName)
}

// GetObject returns a stream of the object data. Most of the common errors occur when reading the stream.
func (c *Client) GetObject(ctx context.Context, bucketName, objectName string, opts common.GetStorageObjectOptions) (io.ReadCloser, error) {
	ctx, span := c.startOtelSpan(ctx, "GetObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	result, err := c.client.DownloadStream(ctx, bucketName, objectName, &blob.DownloadStreamOptions{})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, err
	}

	return result.Body, nil
}

// PutObject uploads objects that are less than 128MiB in a single PUT operation. For objects that are greater than 128MiB in size,
// PutObject seamlessly uploads the object as parts of 128MiB or more depending on the actual file size. The max upload size for an object is 5TB.
func (c *Client) PutObject(ctx context.Context, bucketName string, objectName string, opts *common.PutStorageObjectOptions, reader io.Reader, objectSize int64) (*common.StorageUploadInfo, error) {
	ctx, span := c.startOtelSpan(ctx, "PutObject", bucketName)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", objectName),
		attribute.Int64("http.response.body.size", objectSize),
	)

	uploadOptions := &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{},
		Tags:        opts.UserTags,
		Metadata:    map[string]*string{},
		Concurrency: int(opts.NumThreads),
		BlockSize:   int64(opts.PartSize),
	}

	for key, value := range opts.UserMetadata {
		if value != "" {
			uploadOptions.Metadata[key] = &value
		}
	}

	if opts.CacheControl != "" {
		uploadOptions.HTTPHeaders.BlobCacheControl = &opts.CacheControl
	}

	if opts.ContentDisposition != "" {
		uploadOptions.HTTPHeaders.BlobContentDisposition = &opts.ContentDisposition
	}

	if opts.ContentEncoding != "" {
		uploadOptions.HTTPHeaders.BlobContentEncoding = &opts.ContentEncoding
	}

	if opts.ContentLanguage != "" {
		uploadOptions.HTTPHeaders.BlobContentLanguage = &opts.ContentLanguage
	}

	if opts.SendContentMd5 {
		var hash []byte
		var err error

		reader, hash, err = common.CalculateContentMd5(reader)
		if err != nil {
			span.SetStatus(codes.Error, "failed to calculate content md5")
			span.RecordError(err)

			return nil, fmt.Errorf("failed to calculate content md5: %w", err)
		}

		uploadOptions.HTTPHeaders.BlobContentMD5 = hash
	}

	resp, err := c.client.UploadStream(ctx, bucketName, objectName, reader, uploadOptions)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, err
	}

	result := serializeUploadObjectInfo(resp)
	common.SetUploadInfoAttributes(span, &result)

	return &result, nil
}

// CopyObject creates or replaces an object through server-side copying of an existing object.
// It supports conditional copying, copying a part of an object and server-side encryption of destination and decryption of source.
// To copy multiple source objects into a single destination object see the ComposeObject API.
func (c *Client) CopyObject(ctx context.Context, dest common.StorageCopyDestOptions, src common.StorageCopySrcOptions) (*common.StorageUploadInfo, error) {
	return nil, errNotSupported
}

// ComposeObject creates an object by concatenating a list of source objects using server-side copying.
func (c *Client) ComposeObject(ctx context.Context, dest common.StorageCopyDestOptions, sources []common.StorageCopySrcOptions) (*common.StorageUploadInfo, error) {
	return nil, errNotSupported
}

// StatObject fetches metadata of an object.
func (c *Client) StatObject(ctx context.Context, bucketName, objectName string, opts common.GetStorageObjectOptions) (*common.StorageObject, error) {
	ctx, span := c.startOtelSpan(ctx, "StatObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	objects, err := c.ListObjects(ctx, bucketName, &common.ListStorageObjectsOptions{
		Prefix:       objectName,
		WithVersions: true,
		WithMetadata: true,
		WithTags:     opts.WithTags != nil && *opts.WithTags,
		MaxResults:   1,
	})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, err
	}

	for _, obj := range objects {
		if obj.Name == objectName {
			return &obj, nil
		}
	}

	return nil, nil
}

// RemoveObject removes an object with some specified options.
func (c *Client) RemoveObject(ctx context.Context, bucketName string, objectName string, opts common.RemoveStorageObjectOptions) error {
	return c.removeObject(ctx, "RemoveObject", bucketName, objectName)
}

// RemoveObject removes an object with some specified options.
func (c *Client) removeObject(ctx context.Context, spanName string, bucketName string, objectName string) error {
	ctx, span := c.startOtelSpan(ctx, spanName, bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	deleteType := blob.DeleteTypePermanent
	deleteSnapshots := blob.DeleteSnapshotsOptionTypeInclude
	options := &azblob.DeleteBlobOptions{
		BlobDeleteType:  &deleteType,
		DeleteSnapshots: &deleteSnapshots,
	}

	_, err := c.client.DeleteBlob(ctx, bucketName, objectName, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	}

	return err
}

// RemoveObjects removes a list of objects obtained from an input channel. The call sends a delete request to the server up to 1000 objects at a time.
// The errors observed are sent over the error channel.
func (c *Client) RemoveObjects(ctx context.Context, bucketName string, opts *common.RemoveStorageObjectsOptions, predicate func(string) bool) []common.RemoveStorageObjectError {
	ctx, span := c.startOtelSpan(ctx, "RemoveObjects", bucketName)
	defer span.End()

	objects, err := c.ListObjects(ctx, bucketName, &opts.ListStorageObjectsOptions)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return []common.RemoveStorageObjectError{
			{
				Error: err,
			},
		}
	}

	errs := make([]common.RemoveStorageObjectError, 0)

	for _, obj := range objects {
		if err := c.RemoveObject(ctx, bucketName, obj.Name, common.RemoveStorageObjectOptions{
			ForceDelete: true,
		}); err != nil {
			errs = append(errs, common.RemoveStorageObjectError{
				ObjectName: obj.Name,
				Error:      err,
			})
		}
	}

	return errs
}

// PutObjectRetention applies object retention lock onto an object.
func (c *Client) PutObjectRetention(ctx context.Context, opts *common.PutStorageObjectRetentionOptions) error {
	return errNotSupported
}

// PutObjectLegalHold applies legal-hold onto an object.
func (c *Client) PutObjectLegalHold(ctx context.Context, opts *common.PutStorageObjectLegalHoldOptions) error {
	return errNotSupported
}

// GetObjectLegalHold returns legal-hold status on a given object.
func (c *Client) GetObjectLegalHold(ctx context.Context, opts *common.GetStorageObjectLegalHoldOptions) (common.StorageLegalHoldStatus, error) {
	return "", errNotSupported
}

// PutObjectTagging sets new object Tags to the given object, replaces/overwrites any existing tags.
func (c *Client) SetObjectTags(ctx context.Context, bucketName string, objectName string, options common.SetStorageObjectTagsOptions) error {
	return errNotSupported
}

// PresignedGetObject generates a presigned URL for HTTP GET operations. Browsers/Mobile clients may point to this URL to directly download objects even if the bucket is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The maximum expiry is 604800 seconds (i.e. 7 days) and minimum is 1 second.
func (c *Client) PresignedGetObject(ctx context.Context, bucketName string, objectName string, opts common.PresignedGetStorageObjectOptions) (string, error) {
	expiry := time.Hour

	if opts.Expiry != nil {
		expiry = opts.Expiry.Duration
	}

	return c.presignedObject(ctx, "GET", bucketName, objectName, expiry, sas.AccountPermissions{
		Read: true,
	})
}

// PresignedPutObject generates a presigned URL for HTTP PUT operations. Browsers/Mobile clients may point to this URL to upload objects directly to a bucket even if it is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational. The default expiry is set to 7 days.
func (c *Client) PresignedPutObject(ctx context.Context, bucketName string, objectName string, expiry time.Duration) (string, error) {
	return c.presignedObject(ctx, "PUT", bucketName, objectName, expiry, sas.AccountPermissions{
		Write: true,
	})
}

func (c *Client) presignedObject(ctx context.Context, method, bucketName, objectName string, expiry time.Duration, permissions sas.AccountPermissions) (string, error) {
	_, span := c.startOtelSpan(ctx, method+"PresignedObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))
	span.SetAttributes(attribute.String("storage.expiry", expiry.String()))

	expiredAt := time.Now().Add(expiry)

	result, err := c.client.ServiceClient().GetSASURL(sas.AccountResourceTypes{
		Object: true,
	}, permissions, expiredAt, nil)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", err
	}

	return result, nil
}

// Set object lock configuration in given bucket. mode, validity and unit are either all set or all nil.
func (c *Client) SetObjectLockConfig(ctx context.Context, bucketname string, opts common.SetStorageObjectLockConfig) error {
	return errNotSupported
}

// Get object lock configuration of given bucket.
func (c *Client) GetObjectLockConfig(ctx context.Context, bucketName string) (*common.StorageObjectLockConfig, error) {
	return nil, errNotSupported
}