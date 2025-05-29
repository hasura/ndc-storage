package gcs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/iterator"
)

// ListObjects list objects in a bucket.
func (c *Client) ListObjects(
	ctx context.Context,
	bucketName string,
	opts *common.ListStorageObjectsOptions,
	predicate func(string) bool,
) (*common.StorageObjectListResults, error) {
	ctx, span := c.startOtelSpan(ctx, "ListObjects", bucketName)
	defer span.End()

	var count int

	maxResults := opts.MaxResults
	objects := make([]common.StorageObject, 0)
	q := c.validateListObjectsOptions(span, opts, false)
	pager := c.client.Bucket(bucketName).Objects(ctx, q)
	pageInfo := common.StoragePaginationInfo{}
	started := opts.StartAfter == ""

	for {
		object, err := pager.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}

			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, serializeErrorResponse(err)
		}

		if !started {
			started = object.Name == opts.StartAfter ||
				strings.TrimRight(object.Prefix, "/") == strings.TrimRight(opts.StartAfter, "/")

			continue
		}

		result := serializeObjectInfo(object)
		if predicate == nil || predicate(result.Name) {
			objects = append(objects, result)
			count++

			if maxResults > 0 && count >= maxResults {
				pageInfo.HasNextPage = pager.PageInfo().Remaining() > 0

				break
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
	return []common.StorageObjectMultipartInfo{}, nil
}

// RemoveIncompleteUpload removes a partially uploaded object.
func (c *Client) RemoveIncompleteUpload(
	ctx context.Context,
	bucketName string,
	objectName string,
) error {
	return nil
}

// ListDeletedObjects list deleted objects in a bucket.
func (c *Client) ListDeletedObjects(
	ctx context.Context,
	bucketName string,
	opts *common.ListStorageObjectsOptions,
	predicate func(string) bool,
) (*common.StorageObjectListResults, error) {
	ctx, span := c.startOtelSpan(ctx, "ListDeletedObjects", bucketName)
	defer span.End()

	var count int

	maxResults := opts.MaxResults
	objects := make([]common.StorageObject, 0)
	q := c.validateListObjectsOptions(span, opts, true)
	pager := c.client.Bucket(bucketName).Objects(ctx, q)
	pageInfo := common.StoragePaginationInfo{}

	for {
		object, err := pager.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}

			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, serializeErrorResponse(err)
		}

		if (object.Deleted.IsZero() && object.SoftDeleteTime.IsZero()) ||
			(predicate != nil && !predicate(object.Name)) {
			continue
		}

		result := serializeObjectInfo(object)
		objects = append(objects, result)
		count++

		if maxResults > 0 && count >= maxResults {
			pageInfo.HasNextPage = pager.PageInfo().Remaining() > 0

			break
		}
	}

	span.SetAttributes(attribute.Int("storage.object_count", count))

	results := &common.StorageObjectListResults{
		Objects:  objects,
		PageInfo: pageInfo,
	}

	return results, nil
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

	object, err := c.client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	return object, nil
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

	// estimate the chunk size. If the object size < 16MiB,
	// the chunk size will be rounded up to the nearest multiple of 256K
	chunkSize := 16 * 1024 * 1024
	if objectSize < int64(chunkSize) {
		chunkSize = (int(objectSize/size256K) + 1) * size256K
	}

	w := c.client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	w.ChunkSize = chunkSize
	w.Metadata = common.KeyValuesToStringMap(opts.Metadata)
	w.CacheControl = opts.CacheControl
	w.ContentDisposition = opts.ContentDisposition
	w.ContentEncoding = opts.ContentEncoding
	w.ContentLanguage = opts.ContentLanguage
	w.ContentType = opts.ContentType
	w.TemporaryHold = opts.LegalHold != nil && *opts.LegalHold
	w.StorageClass = opts.StorageClass

	if opts.Retention != nil {
		retention := &storage.ObjectRetention{
			Mode:        string(opts.Retention.Mode),
			RetainUntil: opts.Retention.RetainUntilDate,
		}

		w.Retention = retention
	}

	_, err := io.Copy(w, reader)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	if err := w.Close(); err != nil {
		span.SetStatus(codes.Error, "failed to close the file stream")
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := serializeUploadObjectInfo(w)
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

	srcHandle := c.client.Bucket(src.Bucket).Object(src.Name)
	copier := c.client.Bucket(dest.Bucket).Object(dest.Name).CopierFrom(srcHandle)

	object, err := copier.Run(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := serializeUploadObjectInfo(&storage.Writer{
		ObjectAttrs: *object,
	})
	common.SetUploadInfoAttributes(span, &result)

	return &result, nil
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

	object, err := c.client.Bucket(bucketName).Object(objectName).Attrs(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, nil
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := serializeObjectInfo(object)

	return &result, nil
}

// RemoveObject removes an object with some specified options.
func (c *Client) RemoveObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts common.RemoveStorageObjectOptions,
) error {
	ctx, span := c.startOtelSpan(ctx, "RemoveObject", bucketName)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", objectName),
		attribute.Bool("storage.options.force_delete", opts.ForceDelete),
		attribute.Bool("storage.options.governance_bypass", opts.GovernanceBypass),
	)

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))
	}

	objectClient := c.client.Bucket(bucketName).Object(objectName)

	if opts.SoftDelete {
		objectClient = objectClient.SoftDeleted()
	}

	err := objectClient.SoftDeleted().Delete(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
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
	q := c.validateListObjectsOptions(span, &listOptions, false)
	pager := c.client.Bucket(bucketName).Objects(ctx, q)
	objects := []*storage.ObjectAttrs{}

	for {
		object, err := pager.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}

			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return []common.RemoveStorageObjectError{
				{
					Error: err.Error(),
				},
			}
		}

		objects = append(objects, object)
	}

	span.SetAttributes(attribute.Int("storage.object_count", len(objects)))

	if len(objects) == 0 {
		return nil
	}

	errs := make([]common.RemoveStorageObjectError, 0)

	if opts.NumThreads <= 1 {
		for _, item := range objects {
			err := c.client.Bucket(bucketName).Object(item.Name).Delete(ctx)
			if err != nil {
				errs = append(errs, common.RemoveStorageObjectError{
					ObjectName: item.Name,
					VersionID:  strconv.Itoa(int(item.Generation)),
					Error:      err.Error(),
				})
			}
		}
	} else {
		eg := errgroup.Group{}
		eg.SetLimit(opts.NumThreads)

		removeFunc := func(name string) {
			eg.Go(func() error {
				err := c.client.Bucket(bucketName).Object(name).Delete(ctx)
				if err != nil {
					errs = append(errs, common.RemoveStorageObjectError{
						ObjectName: name,
						Error:      err.Error(),
					})
				}

				return nil
			})
		}

		for _, item := range objects {
			removeFunc(item.Name)
		}

		if err := eg.Wait(); err != nil {
			return []common.RemoveStorageObjectError{
				{
					Error: err.Error(),
				},
			}
		}
	}

	if len(errs) > 0 {
		bs, err := json.Marshal(errs)
		if err != nil {
			slog.Error(err.Error())
		}

		span.SetAttributes(attribute.String("errors", string(bs)))
		span.SetStatus(codes.Error, "failed to remove objects")
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

	handle := c.client.Bucket(bucketName).Object(objectName)

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))

		gen, err := strconv.ParseInt(opts.VersionID, 10, 64)
		if err != nil {
			return schema.UnprocessableContentError(
				fmt.Sprintf("invalid generation version: %s", err),
				nil,
			)
		}

		handle = handle.Generation(gen)
	}

	updateAttrs := storage.ObjectAttrsToUpdate{}

	if opts.LegalHold != nil {
		updateAttrs.TemporaryHold = *opts.LegalHold
	}

	if opts.Retention != nil && opts.Retention.Mode != nil {
		updateAttrs.Retention = &storage.ObjectRetention{
			Mode: string(*opts.Retention.Mode),
		}

		if opts.Retention.RetainUntilDate != nil {
			updateAttrs.Retention.RetainUntil = *opts.Retention.RetainUntilDate
		}
	}

	if opts.Metadata != nil {
		updateAttrs.Metadata = common.KeyValuesToStringMap(*opts.Metadata)
	}

	_, err := handle.Update(ctx, updateAttrs)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// RestoreObject restores a soft-deleted object.
func (c *Client) RestoreObject(ctx context.Context, bucketName string, objectName string) error {
	ctx, span := c.startOtelSpan(ctx, "RestoreObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	blobClient := c.client.Bucket(bucketName).Object(objectName)

	_, err := blobClient.Restore(ctx, &storage.RestoreOptions{})
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
	return c.presignObject(ctx, http.MethodGet, bucketName, objectName, opts)
}

func (c *Client) presignObject(
	ctx context.Context,
	method string,
	bucketName string,
	objectName string,
	opts common.PresignedGetStorageObjectOptions,
) (string, error) {
	_, span := c.startOtelSpan(ctx, method+" PresignedObject", bucketName)
	defer span.End()

	reqParams := url.Values(common.KeyValuesToHeaders(opts.RequestParams))
	span.SetAttributes(
		attribute.String("storage.key", objectName),
		attribute.String("url.query", reqParams.Encode()),
	)

	if opts.Expiry != nil {
		span.SetAttributes(attribute.String("storage.expiry", opts.Expiry.String()))
	}

	fileName := filepath.Base(objectName)
	// Set request Parameters: for content-disposition.
	reqParams.Set(
		"response-content-disposition",
		fmt.Sprintf(`attachment; filename="%s"`, fileName),
	)

	options := &storage.SignedURLOptions{
		Method:          method,
		Expires:         time.Now().Add(opts.Expiry.Duration),
		QueryParameters: reqParams,
	}

	if c.publicHost != nil {
		options.Hostname = c.publicHost.Host
	}

	result, err := c.client.Bucket(bucketName).SignedURL(objectName, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", serializeErrorResponse(err)
	}

	return result, nil
}

// PresignedPutObject generates a presigned URL for HTTP PUT operations. Browsers/Mobile clients may point to this URL to upload objects directly to a bucket even if it is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational. The default expiry is set to 7 days.
func (c *Client) PresignedPutObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	expiry time.Duration,
) (string, error) {
	return c.presignObject(
		ctx,
		http.MethodPut,
		bucketName,
		objectName,
		common.PresignedGetStorageObjectOptions{
			Expiry: &scalar.DurationString{Duration: expiry},
		},
	)
}
