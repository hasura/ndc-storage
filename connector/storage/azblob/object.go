package azblob

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/hasura/ndc-sdk-go/v2/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// ListObjects list objects in a bucket.
func (c *Client) ListObjects(
	ctx context.Context,
	bucketName string,
	opts *common.ListStorageObjectsOptions,
	predicate func(string) bool,
) (*common.StorageObjectListResults, error) {
	if !opts.Recursive {
		return c.listHierarchyObjects(ctx, bucketName, opts, predicate)
	}

	return c.listFlatObjects(ctx, bucketName, opts, predicate)
}

func (c *Client) listFlatObjects(
	ctx context.Context,
	bucketName string,
	opts *common.ListStorageObjectsOptions,
	predicate func(string) bool,
) (*common.StorageObjectListResults, error) {
	ctx, span := c.startOtelSpan(ctx, "ListObjects", bucketName)
	defer span.End()

	options := &container.ListBlobsFlatOptions{
		Include: makeListBlobsInclude(opts.Include),
	}

	if opts.Prefix != "" {
		options.Prefix = &opts.Prefix
	}

	maxResults := int32(opts.MaxResults)
	if opts.MaxResults > 0 && predicate == nil {
		options.MaxResults = &maxResults
	}

	if opts.StartAfter != "" {
		options.Marker = &opts.StartAfter
	}

	var count int32

	objects := make([]common.StorageObject, 0)
	pager := c.client.NewListBlobsFlatPager(bucketName, options)
	pageInfo := common.StoragePaginationInfo{}

L:
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, serializeErrorResponse(err)
		}

		for i, item := range resp.Segment.BlobItems {
			if item.Name == nil || (predicate != nil && !predicate(*item.Name)) {
				continue
			}

			object := serializeObjectInfo(item)
			objects = append(objects, object)
			count++

			if maxResults > 0 && count >= maxResults {
				pageInfo.HasNextPage = i < len(resp.Segment.BlobItems)-1 || pager.More()

				break L
			}
		}
	}

	span.SetAttributes(attribute.Int("storage.object_count", int(count)))

	results := &common.StorageObjectListResults{
		Objects:  objects,
		PageInfo: pageInfo,
	}

	return results, nil
}

func (c *Client) listHierarchyObjects(
	ctx context.Context,
	bucketName string,
	opts *common.ListStorageObjectsOptions,
	predicate func(string) bool,
) (*common.StorageObjectListResults, error) {
	ctx, span := c.startOtelSpan(ctx, "ListObjects", bucketName)
	defer span.End()

	options := &container.ListBlobsHierarchyOptions{
		Include: makeListBlobsInclude(opts.Include),
	}

	if opts.Prefix != "" {
		options.Prefix = &opts.Prefix
	}

	maxResults := opts.MaxResults

	if opts.MaxResults > 0 && predicate == nil {
		mr := int32(opts.MaxResults)
		options.MaxResults = &mr

		if opts.StartAfter != "" {
			*options.MaxResults += 1
		}
	}

	if opts.StartAfter != "" {
		options.Marker = &opts.StartAfter
	}

	var count int

	objects := make([]common.StorageObject, 0)
	pager := c.client.ServiceClient().
		NewContainerClient(bucketName).
		NewListBlobsHierarchyPager("/", options)
	pageInfo := common.StoragePaginationInfo{}

L:
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, serializeErrorResponse(err)
		}

		for i, item := range resp.Segment.BlobPrefixes {
			// azure does not returns results after the marker. We should ignore the start result.
			if item.Name == nil || (opts.StartAfter != "" && strings.TrimRight(*item.Name, "/") == strings.TrimRight(opts.StartAfter, "/")) || (predicate != nil && !predicate(*item.Name)) {
				continue
			}

			object := common.StorageObject{
				Name:        *item.Name,
				IsDirectory: true,
			}
			objects = append(objects, object)
			count++

			if maxResults > 0 && count >= maxResults {
				if i < len(resp.Segment.BlobPrefixes)-1 || len(resp.Segment.BlobItems) > 0 || pager.More() {
					pageInfo.HasNextPage = true
				}

				break L
			}
		}

		for i, item := range resp.Segment.BlobItems {
			if item.Name == nil || (predicate != nil && !predicate(*item.Name)) {
				continue
			}

			object := serializeObjectInfo(item)
			objects = append(objects, object)
			count++

			if maxResults > 0 && count >= maxResults {
				if i < len(resp.Segment.BlobItems)-1 || pager.More() {
					pageInfo.HasNextPage = true
				}

				break L
			}
		}
	}

	span.SetAttributes(attribute.Int("storage.object_count", count))

	results := &common.StorageObjectListResults{
		Objects:  objects,
		PageInfo: pageInfo,
	}

	return results, nil
}

// ListIncompleteUploads list partially uploaded objects in a bucket.
func (c *Client) ListIncompleteUploads(
	ctx context.Context,
	bucketName string,
	args common.ListIncompleteUploadsOptions,
) ([]common.StorageObjectMultipartInfo, error) {
	ctx, span := c.startOtelSpan(ctx, "ListIncompleteUploads", bucketName)
	defer span.End()

	options := &container.ListBlobsFlatOptions{
		Include: container.ListBlobsInclude{
			UncommittedBlobs: true,
		},
	}

	objects := make([]common.StorageObjectMultipartInfo, 0)
	pager := c.client.NewListBlobsFlatPager(bucketName, options)

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, serializeErrorResponse(err)
		}

		for _, item := range resp.Segment.BlobItems {
			if item.Properties == nil || item.Properties.ETag == nil ||
				*item.Properties.ETag == "" {
				continue
			}

			obj := common.StorageObjectMultipartInfo{
				Name: item.Name,
			}

			if item.Properties != nil {
				if item.Properties.CreationTime != nil {
					obj.Initiated = item.Properties.CreationTime
				} else if item.Properties.LastModified != nil {
					obj.Initiated = item.Properties.CreationTime
				}

				if item.Properties.ContentLength != nil && *item.Properties.ContentLength > 0 {
					obj.Size = item.Properties.ContentLength
				}
			}

			objects = append(objects, obj)
		}
	}

	span.SetAttributes(attribute.Int("storage.object_count", len(objects)))

	return objects, nil
}

// ListDeletedObjects list soft-deleted objects in a bucket.
func (c *Client) ListDeletedObjects(
	ctx context.Context,
	bucketName string,
	opts *common.ListStorageObjectsOptions,
	predicate func(string) bool,
) (*common.StorageObjectListResults, error) {
	ctx, span := c.startOtelSpan(ctx, "ListDeletedObjects", bucketName)
	defer span.End()

	options := &container.ListBlobsFlatOptions{
		Include: container.ListBlobsInclude{
			Versions:            opts.Include.Versions,
			Metadata:            opts.Include.Metadata,
			Tags:                opts.Include.Tags,
			Copy:                opts.Include.Copy,
			Snapshots:           opts.Include.Snapshots,
			LegalHold:           opts.Include.LegalHold,
			ImmutabilityPolicy:  opts.Include.Retention,
			Permissions:         opts.Include.Permissions,
			Deleted:             true,
			DeletedWithVersions: true,
			UncommittedBlobs:    false,
		},
	}

	if opts.Prefix != "" {
		options.Prefix = &opts.Prefix
	}

	maxResults := int32(opts.MaxResults)

	if opts.StartAfter != "" {
		options.Marker = &opts.StartAfter
	}

	var count int32

	objects := make([]common.StorageObject, 0)
	pager := c.client.NewListBlobsFlatPager(bucketName, options)
	pageInfo := common.StoragePaginationInfo{}

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, serializeErrorResponse(err)
		}

		for _, item := range resp.Segment.BlobItems {
			if item.Name == nil || item.Deleted == nil || !*item.Deleted ||
				(predicate != nil && !predicate(*item.Name)) {
				continue
			}

			objects = append(objects, serializeObjectInfo(item))
			count++
		}

		if maxResults > 0 && count >= maxResults {
			if pager.More() {
				pageInfo.HasNextPage = true
			}

			break
		}
	}

	span.SetAttributes(attribute.Int("storage.object_count", int(count)))

	results := &common.StorageObjectListResults{
		Objects:  objects,
		PageInfo: pageInfo,
	}

	return results, nil
}

// RemoveIncompleteUpload removes a partially uploaded object.
func (c *Client) RemoveIncompleteUpload(
	ctx context.Context,
	bucketName string,
	objectName string,
) error {
	return c.removeObject(
		ctx,
		"RemoveIncompleteUpload",
		bucketName,
		objectName,
		common.RemoveStorageObjectOptions{},
	)
}

// GetObject returns a stream of the object data. Most of the common errors occur when reading the stream.
func (c *Client) GetObject(
	ctx context.Context,
	bucketName, objectName string,
	opts common.GetStorageObjectOptions,
) (io.ReadCloser, error) {
	ctx, span := c.startOtelSpan(ctx, "GetObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	result, err := c.client.DownloadStream(
		ctx,
		bucketName,
		objectName,
		&blob.DownloadStreamOptions{},
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	return result.Body, nil
}

// PutObject uploads objects that are less than 128MiB in a single PUT operation. For objects that are greater than 128MiB in size,
// PutObject seamlessly uploads the object as parts of 128MiB or more depending on the actual file size. The max upload size for an object is 5TB.
func (c *Client) PutObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts *common.PutStorageObjectOptions,
	reader io.Reader,
	objectSize int64,
) (*common.StorageUploadInfo, error) {
	ctx, span := c.startOtelSpan(ctx, "PutObject", bucketName)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", objectName),
		attribute.Int64("http.response.body.size", objectSize),
	)

	uploadOptions := &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{},
		Tags:        common.KeyValuesToStringMap(opts.Tags),
		Metadata:    map[string]*string{},
		Concurrency: int(opts.NumThreads),
		BlockSize:   int64(opts.PartSize),
	}

	if opts.StorageClass != "" {
		accessTier := blob.AccessTier(opts.StorageClass)
		if !slices.Contains(blob.PossibleAccessTierValues(), accessTier) {
			return nil, schema.UnprocessableContentError(
				"invalid Azure Blob access tier: "+opts.StorageClass,
				nil,
			)
		}

		uploadOptions.AccessTier = &accessTier
	}

	for _, item := range opts.Metadata {
		uploadOptions.Metadata[item.Key] = &item.Value
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

		return nil, serializeErrorResponse(err)
	}

	if opts.Retention != nil && opts.Retention.Mode == common.StorageRetentionModeLocked {
		err := c.SetObjectRetention(
			ctx,
			bucketName,
			objectName,
			"",
			common.SetStorageObjectRetentionOptions{
				Mode:            &opts.Retention.Mode,
				RetainUntilDate: &opts.Retention.RetainUntilDate,
			},
		)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, err
		}
	}

	result := serializeUploadObjectInfo(resp)
	result.Name = objectName

	common.SetUploadInfoAttributes(span, &result)

	return &result, nil
}

// CopyObject creates or replaces an object through server-side copying of an existing object.
// It supports conditional copying, copying a part of an object and server-side encryption of destination and decryption of source.
// To copy multiple source objects into a single destination object see the ComposeObject API.
func (c *Client) CopyObject(
	ctx context.Context,
	dest common.StorageCopyDestOptions,
	src common.StorageCopySrcOptions,
) (*common.StorageUploadInfo, error) {
	ctx, span := c.startOtelSpan(ctx, "CopyObject", dest.Bucket)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", dest.Name),
		attribute.String("storage.copy_source", src.Name),
	)

	srcURL := c.client.ServiceClient().NewContainerClient(src.Bucket).NewBlobClient(src.Name).URL()
	blobClient := c.client.ServiceClient().NewContainerClient(dest.Bucket).NewBlobClient(dest.Name)

	options := &blob.CopyFromURLOptions{
		BlobTags:  common.KeyValuesToStringMap(dest.Tags),
		Metadata:  make(map[string]*string),
		LegalHold: dest.LegalHold,
	}

	for _, item := range dest.Metadata {
		options.Metadata[item.Key] = &item.Value
	}

	resp, err := blobClient.CopyFromURL(ctx, srcURL, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := &common.StorageUploadInfo{
		Bucket: dest.Bucket,
		Name:   dest.Name,
	}

	if resp.ETag != nil && *resp.ETag != "" {
		etag, _ := strconv.Unquote(string(*resp.ETag))
		result.ETag = &etag
	}

	if len(resp.ContentMD5) > 0 {
		contentMD5 := base64.StdEncoding.EncodeToString(resp.ContentMD5)
		result.ContentMD5 = &contentMD5
	}

	if len(resp.ContentCRC64) > 0 {
		crc64 := base64.StdEncoding.EncodeToString(resp.ContentCRC64)
		result.ChecksumCRC64NVME = &crc64
	}

	if resp.LastModified != nil {
		result.LastModified = resp.LastModified
	}

	if resp.VersionID != nil {
		result.VersionID = resp.VersionID
	}

	common.SetUploadInfoAttributes(span, result)

	return result, nil
}

// ComposeObject creates an object by concatenating a list of source objects using server-side copying.
func (c *Client) ComposeObject(
	ctx context.Context,
	dest common.StorageCopyDestOptions,
	sources []common.StorageCopySrcOptions,
) (*common.StorageUploadInfo, error) {
	return nil, errNotSupported
}

// StatObject fetches metadata of an object.
func (c *Client) StatObject(
	ctx context.Context,
	bucketName, objectName string,
	opts common.GetStorageObjectOptions,
) (*common.StorageObject, error) {
	ctx, span := c.startOtelSpan(ctx, "StatObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	results, err := c.ListObjects(ctx, bucketName, &common.ListStorageObjectsOptions{
		Prefix:     objectName,
		MaxResults: 1,
		Include:    opts.Include,
	}, nil)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	for _, obj := range results.Objects {
		if obj.Name == objectName {
			return &obj, nil
		}
	}

	return nil, nil
}

// RemoveObject removes an object with some specified options.
func (c *Client) RemoveObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts common.RemoveStorageObjectOptions,
) error {
	return c.removeObject(ctx, "RemoveObject", bucketName, objectName, opts)
}

// RemoveObject removes an object with some specified options.
func (c *Client) removeObject(
	ctx context.Context,
	spanName string,
	bucketName string,
	objectName string,
	opts common.RemoveStorageObjectOptions,
) error {
	ctx, span := c.startOtelSpan(ctx, spanName, bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	options := &azblob.DeleteBlobOptions{}

	if opts.ForceDelete {
		deleteType := blob.DeleteTypePermanent
		deleteSnapshots := blob.DeleteSnapshotsOptionTypeInclude
		options.DeleteSnapshots = &deleteSnapshots
		options.BlobDeleteType = &deleteType
	}

	_, err := c.client.DeleteBlob(ctx, bucketName, objectName, options)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if respErr.ErrorCode == string(bloberror.BlobNotFound) {
				return nil
			}

			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return serializeAzureErrorResponse(respErr)
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return schema.UnprocessableContentError(err.Error(), nil)
	}

	return nil
}

// RemoveObjects removes a list of objects obtained from an input channel. The call sends a delete request to the server up to 1000 objects at a time.
// The errors observed are sent over the error channel.
func (c *Client) RemoveObjects(
	ctx context.Context,
	bucketName string,
	opts *common.RemoveStorageObjectsOptions,
	predicate func(string) bool,
) []common.RemoveStorageObjectError {
	ctx, span := c.startOtelSpan(ctx, "RemoveObjects", bucketName)
	defer span.End()

	listOptions := opts.ListStorageObjectsOptions
	listOptions.Recursive = true

	results, err := c.ListObjects(ctx, bucketName, &listOptions, predicate)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return []common.RemoveStorageObjectError{
			{
				Error: err.Error(),
			},
		}
	}

	errs := make([]common.RemoveStorageObjectError, 0)
	containerClient := c.client.ServiceClient().NewContainerClient(bucketName)

	for chunk := range slices.Chunk(results.Objects, 256) {
		batchBuilder, err := containerClient.NewBatchBuilder()
		if err != nil {
			return []common.RemoveStorageObjectError{
				{
					Error: err.Error(),
				},
			}
		}

		for _, obj := range chunk {
			err := batchBuilder.Delete(obj.Name, &container.BatchDeleteOptions{})
			if err != nil {
				return []common.RemoveStorageObjectError{
					{
						Error: err.Error(),
					},
				}
			}
		}

		if _, err := containerClient.SubmitBatch(ctx, batchBuilder, nil); err != nil {
			errs = append(errs, common.RemoveStorageObjectError{
				Error: err.Error(),
			})
		}
	}

	return errs
}

// UpdateObject updates object configurations.
func (c *Client) UpdateObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts common.UpdateStorageObjectOptions,
) error {
	ctx, span := c.startOtelSpan(ctx, "UpdateObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))
	}

	if opts.LegalHold != nil {
		err := c.SetObjectLegalHold(ctx, bucketName, objectName, opts.VersionID, *opts.LegalHold)
		if err != nil {
			return err
		}
	}

	if opts.Tags != nil {
		err := c.SetObjectTags(
			ctx,
			bucketName,
			objectName,
			opts.VersionID,
			common.KeyValuesToStringMap(*opts.Tags),
		)
		if err != nil {
			return err
		}
	}

	if opts.Retention != nil {
		err := c.SetObjectRetention(ctx, bucketName, objectName, opts.VersionID, *opts.Retention)
		if err != nil {
			return err
		}
	}

	return nil
}

// RestoreObject restores an object with some specified options.
func (c *Client) RestoreObject(ctx context.Context, bucketName string, objectName string) error {
	ctx, span := c.startOtelSpan(ctx, "RestoreObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	blobClient := c.client.ServiceClient().NewContainerClient(bucketName).NewBlobClient(objectName)

	_, err := blobClient.Undelete(ctx, nil)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// SetObjectRetention applies object retention lock onto an object.
func (c *Client) SetObjectRetention(
	ctx context.Context,
	bucketName string,
	objectName, versionID string,
	opts common.SetStorageObjectRetentionOptions,
) error {
	ctx, span := c.startOtelSpan(ctx, "SetObjectRetention", bucketName)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", objectName),
	)

	if versionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", versionID))
	}

	client := c.client.ServiceClient().NewContainerClient(bucketName).NewBlobClient(objectName)
	if opts.RetainUntilDate == nil {
		_, err := client.DeleteImmutabilityPolicy(ctx, nil)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return serializeErrorResponse(err)
		}

		return nil
	}

	span.SetAttributes(
		attribute.String(
			"storage.options.retain_util_date",
			opts.RetainUntilDate.Format(time.RFC3339),
		),
	)

	_, err := client.SetImmutabilityPolicy(
		ctx,
		*opts.RetainUntilDate,
		&blob.SetImmutabilityPolicyOptions{
			Mode: (*blob.ImmutabilityPolicySetting)(opts.Mode),
		},
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// SetObjectLegalHold applies legal-hold onto an object.
func (c *Client) SetObjectLegalHold(
	ctx context.Context,
	bucketName string,
	objectName, versionID string,
	status bool,
) error {
	ctx, span := c.startOtelSpan(ctx, "SetObjectTags", bucketName)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", objectName),
		attribute.Bool("storage.options.status", status),
	)

	client := c.client.ServiceClient().NewContainerClient(bucketName).NewBlobClient(objectName)

	_, err := client.SetLegalHold(ctx, status, &blob.SetLegalHoldOptions{})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// SetObjectTags sets new object Tags to the given object, replaces/overwrites any existing tags.
func (c *Client) SetObjectTags(
	ctx context.Context,
	bucketName string,
	objectName, versionID string,
	tags map[string]string,
) error {
	ctx, span := c.startOtelSpan(ctx, "SetObjectTags", bucketName)
	defer span.End()

	opts := &blob.SetTagsOptions{
		VersionID: &versionID,
	}

	if versionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", versionID))
		opts.VersionID = &versionID
	}

	client := c.client.ServiceClient().NewContainerClient(bucketName).NewBlobClient(objectName)

	_, err := client.SetTags(ctx, tags, opts)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// PresignedGetObject generates a presigned URL for HTTP GET operations. Browsers/Mobile clients may point to this URL to directly download objects even if the bucket is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The maximum expiry is 604800 seconds (i.e. 7 days) and minimum is 1 second.
func (c *Client) PresignedGetObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts common.PresignedGetStorageObjectOptions,
) (string, error) {
	expiry := time.Hour

	if opts.Expiry != nil {
		expiry = opts.Expiry.Duration
	}

	return c.presignedObject(ctx, "GET", bucketName, objectName, expiry, sas.BlobPermissions{
		Read: true,
	})
}

// PresignedPutObject generates a presigned URL for HTTP PUT operations. Browsers/Mobile clients may point to this URL to upload objects directly to a bucket even if it is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational. The default expiry is set to 7 days.
func (c *Client) PresignedPutObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	expiry time.Duration,
) (string, error) {
	return c.presignedObject(ctx, "PUT", bucketName, objectName, expiry, sas.BlobPermissions{
		Write:  true,
		Add:    true,
		Create: true,
	})
}

func (c *Client) presignedObject(
	ctx context.Context,
	method, bucketName, objectName string,
	expiry time.Duration,
	permissions sas.BlobPermissions,
) (string, error) {
	_, span := c.startOtelSpan(ctx, method+"PresignedObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))
	span.SetAttributes(attribute.String("storage.expiry", expiry.String()))

	expiredAt := time.Now().Add(expiry)

	result, err := c.client.ServiceClient().
		NewContainerClient(bucketName).
		NewBlobClient(objectName).
		GetSASURL(permissions, expiredAt, nil)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", serializeErrorResponse(err)
	}

	return result, nil
}
