package minio

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/hasura/ndc-sdk-go/v2/connector"
	"github.com/hasura/ndc-sdk-go/v2/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/tags"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

// ListObjects list objects in a bucket.
func (mc *Client) ListObjects(
	ctx context.Context,
	bucketName string,
	opts *common.ListStorageObjectsOptions,
	predicate func(string) bool,
) (*common.StorageObjectListResults, error) {
	ctx, span := mc.startOtelSpan(ctx, "ListObjects", bucketName)
	defer span.End()

	logger := connector.GetLogger(ctx)
	maxResults := opts.MaxResults

	objChan := mc.client.ListObjects(ctx, bucketName, mc.validateListObjectsOptions(span, opts))
	minioObjects := []minio.ObjectInfo{}

	for obj := range objChan {
		if obj.Err != nil {
			span.SetStatus(codes.Error, obj.Err.Error())
			span.RecordError(obj.Err)

			return nil, serializeErrorResponse(obj.Err)
		}

		if predicate != nil && !predicate(obj.Key) {
			continue
		}

		minioObjects = append(minioObjects, obj)
	}

	if len(minioObjects) == 0 {
		span.SetAttributes(attribute.Int("storage.object_count", 0))

		return &common.StorageObjectListResults{
			Objects: []common.StorageObject{},
		}, nil
	}

	maxLength := len(minioObjects)
	pageInfo := common.StoragePaginationInfo{}

	if maxResults > 0 && maxResults < maxLength {
		maxLength = maxResults
		pageInfo.HasNextPage = true
	}

	objects := make([]common.StorageObject, maxLength)

	for i := range maxLength {
		object := serializeObjectInfo(&minioObjects[i], true)
		object.Bucket = bucketName

		objects[i] = object
	}

	span.SetAttributes(attribute.Int("storage.object_count", maxLength))

	if opts.Include.IsEmpty() {
		return &common.StorageObjectListResults{
			Objects:  objects,
			PageInfo: pageInfo,
		}, nil
	}

	if opts.NumThreads <= 1 {
		for i, object := range objects {
			err := mc.populateObject(ctx, &object, opts.Include)
			if err == nil {
				objects[i] = object
			}
		}

		return &common.StorageObjectListResults{
			Objects: objects,
		}, nil
	}

	eg := errgroup.Group{}
	eg.SetLimit(opts.NumThreads)

	results := make([]common.StorageObject, len(objects))

	lhFunc := func(obj common.StorageObject, index int) {
		eg.Go(func() error {
			err := mc.populateObject(ctx, &obj, opts.Include)
			if err == nil {
				results[index] = obj
			}

			return nil
		})
	}

	for i, object := range objects {
		lhFunc(object, i)
	}

	err := eg.Wait()
	if err != nil {
		logger.Error("failed to include object data: " + err.Error())
		span.AddEvent(
			"fetch_legal_holds_error",
			trace.WithAttributes(attribute.String("error", err.Error())),
		)
	}

	return &common.StorageObjectListResults{
		Objects:  objects,
		PageInfo: pageInfo,
	}, nil
}

// ListIncompleteUploads list partially uploaded objects in a bucket.
func (mc *Client) ListIncompleteUploads(
	ctx context.Context,
	bucketName string,
	args common.ListIncompleteUploadsOptions,
) ([]common.StorageObjectMultipartInfo, error) {
	ctx, span := mc.startOtelSpan(ctx, "ListIncompleteUploads", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.object_prefix", args.Prefix))

	objChan := mc.client.ListIncompleteUploads(ctx, bucketName, args.Prefix, true)
	objects := make([]common.StorageObjectMultipartInfo, 0)

	for obj := range objChan {
		if obj.Err != nil {
			span.SetStatus(codes.Error, obj.Err.Error())
			span.RecordError(obj.Err)

			return nil, serializeErrorResponse(obj.Err)
		}

		object := common.StorageObjectMultipartInfo{
			Initiated: &obj.Initiated,
		}

		if !isStringNull(obj.Key) {
			object.Name = &obj.Key
		}

		if !isStringNull(obj.StorageClass) {
			object.StorageClass = &obj.StorageClass
		}

		if !isStringNull(obj.UploadID) {
			object.UploadID = &obj.UploadID
		}

		if obj.Size > 0 {
			object.Size = &obj.Size
		}

		objects = append(objects, object)
	}

	span.SetAttributes(attribute.Int("storage.object_count", len(objects)))

	return objects, nil
}

// ListDeletedObjects list soft-deleted objects in a bucket.
func (mc *Client) ListDeletedObjects(
	ctx context.Context,
	bucketName string,
	opts *common.ListStorageObjectsOptions,
	predicate func(string) bool,
) (*common.StorageObjectListResults, error) {
	return &common.StorageObjectListResults{
		Objects: []common.StorageObject{},
	}, nil
}

// RemoveIncompleteUpload removes a partially uploaded object.
func (mc *Client) RemoveIncompleteUpload(
	ctx context.Context,
	bucketName string,
	objectName string,
) error {
	ctx, span := mc.startOtelSpan(ctx, "RemoveIncompleteUpload", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	err := mc.client.RemoveIncompleteUpload(ctx, bucketName, objectName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// GetObject returns a stream of the object data. Most of the common errors occur when reading the stream.
func (mc *Client) GetObject(
	ctx context.Context,
	bucketName, objectName string,
	opts common.GetStorageObjectOptions,
) (io.ReadCloser, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))
	options := serializeGetObjectOptions(span, opts)

	object, err := mc.client.GetObject(ctx, bucketName, objectName, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, evalNotFoundError(err, objectNotFoundErrorCode)
	}

	return object, nil
}

// PutObject uploads objects that are less than 128MiB in a single PUT operation. For objects that are greater than 128MiB in size,
// PutObject seamlessly uploads the object as parts of 128MiB or more depending on the actual file size. The max upload size for an object is 5TB.
func (mc *Client) PutObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts *common.PutStorageObjectOptions,
	reader io.Reader,
	objectSize int64,
) (*common.StorageUploadInfo, error) {
	ctx, span := mc.startOtelSpan(ctx, "PutObject", bucketName)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", objectName),
		attribute.Int64("http.response.body.size", objectSize),
	)

	options := minio.PutObjectOptions{
		UserMetadata:            common.KeyValuesToStringMap(opts.Metadata),
		UserTags:                common.KeyValuesToStringMap(opts.Tags),
		ContentType:             opts.ContentType,
		ContentEncoding:         opts.ContentEncoding,
		ContentDisposition:      opts.ContentDisposition,
		ContentLanguage:         opts.ContentLanguage,
		CacheControl:            opts.CacheControl,
		NumThreads:              opts.NumThreads,
		StorageClass:            opts.StorageClass,
		PartSize:                opts.PartSize,
		SendContentMd5:          opts.SendContentMd5,
		DisableContentSha256:    opts.DisableContentSha256,
		DisableMultipart:        opts.DisableMultipart,
		WebsiteRedirectLocation: opts.WebsiteRedirectLocation,
		ConcurrentStreamParts:   opts.ConcurrentStreamParts,
		LegalHold:               validateLegalHoldStatus(opts.LegalHold),
	}

	if opts.Expires != nil {
		options.Expires = *opts.Expires
	}

	if opts.Retention != nil {
		options.Mode = validateObjectRetentionMode(opts.Retention.Mode)
		options.RetainUntilDate = opts.Retention.RetainUntilDate

		span.SetAttributes(
			attribute.String("storage.options.retention_mode", string(opts.Retention.Mode)),
			attribute.String(
				"storage.options.retain_util_date",
				opts.Retention.RetainUntilDate.Format(time.RFC3339),
			),
		)
	}

	if opts.Checksum != nil {
		options.Checksum = parseChecksumType(*opts.Checksum)
	}

	if opts.AutoChecksum != nil {
		options.AutoChecksum = parseChecksumType(*opts.AutoChecksum)
	}

	object, err := mc.client.PutObject(ctx, bucketName, objectName, reader, objectSize, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := serializeUploadObjectInfo(object)
	common.SetUploadInfoAttributes(span, &result)

	return &result, nil
}

// CopyObject creates or replaces an object through server-side copying of an existing object.
// It supports conditional copying, copying a part of an object and server-side encryption of destination and decryption of source.
// To copy multiple source objects into a single destination object see the ComposeObject API.
func (mc *Client) CopyObject(
	ctx context.Context,
	dest common.StorageCopyDestOptions,
	src common.StorageCopySrcOptions,
) (*common.StorageUploadInfo, error) {
	ctx, span := mc.startOtelSpan(ctx, "CopyObject", dest.Bucket)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", dest.Name),
		attribute.String("storage.copy_source", src.Name),
	)

	destOptions := convertCopyDestOptions(dest)
	srcOptions := serializeCopySourceOptions(src)

	object, err := mc.client.CopyObject(ctx, *destOptions, srcOptions)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := serializeUploadObjectInfo(object)
	common.SetUploadInfoAttributes(span, &result)

	return &result, nil
}

// ComposeObject creates an object by concatenating a list of source objects using server-side copying.
func (mc *Client) ComposeObject(
	ctx context.Context,
	dest common.StorageCopyDestOptions,
	sources []common.StorageCopySrcOptions,
) (*common.StorageUploadInfo, error) {
	ctx, span := mc.startOtelSpan(ctx, "ComposeObject", dest.Bucket)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", dest.Name))

	srcKeys := make([]string, len(sources))
	srcOptions := make([]minio.CopySrcOptions, len(sources))

	for i, src := range sources {
		srcKeys[i] = src.Name
		source := serializeCopySourceOptions(src)
		srcOptions[i] = source
	}

	span.SetAttributes(attribute.StringSlice("storage.copy_sources", srcKeys))

	destOptions := convertCopyDestOptions(dest)

	object, err := mc.client.ComposeObject(ctx, *destOptions, srcOptions...)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := serializeUploadObjectInfo(object)
	common.SetUploadInfoAttributes(span, &result)

	return &result, nil
}

// StatObject fetches metadata of an object.
func (mc *Client) StatObject(
	ctx context.Context,
	bucketName, objectName string,
	opts common.GetStorageObjectOptions,
) (*common.StorageObject, error) {
	ctx, span := mc.startOtelSpan(ctx, "StatObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))
	options := serializeGetObjectOptions(span, opts)

	object, err := mc.client.StatObject(ctx, bucketName, objectName, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, evalNotFoundError(err, objectNotFoundErrorCode)
	}

	result := serializeObjectInfo(&object, false)
	result.Bucket = bucketName

	err = mc.populateObject(ctx, &result, opts.Include)
	if err != nil {
		return nil, err
	}

	common.SetObjectInfoSpanAttributes(span, &result)

	return &result, nil
}

// RemoveObject removes an object with some specified options.
func (mc *Client) RemoveObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts common.RemoveStorageObjectOptions,
) error {
	ctx, span := mc.startOtelSpan(ctx, "RemoveObject", bucketName)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", objectName),
		attribute.Bool("storage.options.force_delete", opts.ForceDelete),
		attribute.Bool("storage.options.governance_bypass", opts.GovernanceBypass),
	)

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))
	}

	options := minio.RemoveObjectOptions{
		ForceDelete:      opts.ForceDelete,
		GovernanceBypass: opts.GovernanceBypass,
		VersionID:        opts.VersionID,
	}

	err := mc.client.RemoveObject(ctx, bucketName, objectName, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// RemoveObjects removes a list of objects obtained from an input channel. The call sends a delete request to the server up to 1000 objects at a time.
// The errors observed are sent over the error channel.
func (mc *Client) RemoveObjects(
	ctx context.Context,
	bucketName string,
	opts *common.RemoveStorageObjectsOptions,
	predicate func(string) bool,
) []common.RemoveStorageObjectError {
	ctx, span := mc.startOtelSpan(ctx, "RemoveObjects", bucketName)
	defer span.End()

	listOptions := mc.validateListObjectsOptions(span, &opts.ListStorageObjectsOptions)
	listOptions.Recursive = true
	objectChan := mc.client.ListObjects(ctx, bucketName, listOptions)

	span.SetAttributes(attribute.Bool("storage.options.governance_bypass", opts.GovernanceBypass))

	options := minio.RemoveObjectsOptions{
		GovernanceBypass: opts.GovernanceBypass,
	}

	removeObjectChan := objectChan
	if predicate != nil {
		removeObjectChan := make(chan minio.ObjectInfo, 1)
		defer close(removeObjectChan)

		go func() {
			for ch := range objectChan {
				if !predicate(ch.Key) {
					continue
				}

				removeObjectChan <- ch
			}

			close(removeObjectChan)
		}()
	}

	errChan := mc.client.RemoveObjects(ctx, bucketName, removeObjectChan, options)
	errs := make([]common.RemoveStorageObjectError, 0)

	for err := range errChan {
		errs = append(errs, common.RemoveStorageObjectError{
			ObjectName: err.ObjectName,
			VersionID:  err.VersionID,
			Error:      err.Err.Error(),
		})
	}

	return errs
}

// UpdateObject updates object configurations.
func (mc *Client) UpdateObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts common.UpdateStorageObjectOptions,
) error {
	ctx, span := mc.startOtelSpanWithKind(ctx, trace.SpanKindInternal, "UpdateObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))
	}

	if opts.LegalHold != nil {
		err := mc.SetObjectLegalHold(ctx, bucketName, objectName, opts.VersionID, opts.LegalHold)
		if err != nil {
			return err
		}
	}

	if opts.Tags != nil {
		err := mc.SetObjectTags(
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
		err := mc.SetObjectRetention(ctx, bucketName, objectName, opts.VersionID, *opts.Retention)
		if err != nil {
			return err
		}
	}

	return nil
}

// RestoreObject restores a soft-deleted object.
func (mc *Client) RestoreObject(ctx context.Context, bucketName string, objectName string) error {
	return schema.NotSupportedError("MinIO does not support this function", nil)
}

// SetObjectRetention applies object retention lock onto an object.
func (mc *Client) SetObjectRetention(
	ctx context.Context,
	bucketName string,
	objectName, versionID string,
	opts common.SetStorageObjectRetentionOptions,
) error {
	ctx, span := mc.startOtelSpan(ctx, "SetObjectRetention", bucketName)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", objectName),
		attribute.Bool("storage.options.governance_bypass", opts.GovernanceBypass),
	)

	if versionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", versionID))
	}

	if opts.RetainUntilDate != nil {
		span.SetAttributes(
			attribute.String(
				"storage.options.retain_util_date",
				opts.RetainUntilDate.Format(time.RFC3339),
			),
		)
	}

	options := minio.PutObjectRetentionOptions{
		GovernanceBypass: opts.GovernanceBypass,
		VersionID:        versionID,
		RetainUntilDate:  opts.RetainUntilDate,
	}

	if opts.Mode != nil {
		mode := validateObjectRetentionMode(*opts.Mode)

		options.Mode = &mode
	}

	err := mc.client.PutObjectRetention(ctx, bucketName, objectName, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// SetObjectLegalHold applies legal-hold onto an object.
func (mc *Client) SetObjectLegalHold(
	ctx context.Context,
	bucketName, objectName, versionID string,
	status *bool,
) error {
	ctx, span := mc.startOtelSpan(ctx, "SetObjectLegalHold", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	options := minio.PutObjectLegalHoldOptions{
		VersionID: versionID,
	}

	if versionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", versionID))
	}

	if status != nil {
		span.SetAttributes(attribute.Bool("storage.options.status", *status))
		legalHold := validateLegalHoldStatus(status)
		options.Status = &legalHold
	}

	err := mc.client.PutObjectLegalHold(ctx, bucketName, objectName, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// GetObjectLegalHold returns legal-hold status on a given object.
func (mc *Client) GetObjectLegalHold(
	ctx context.Context,
	bucketName string,
	objectName string,
	versionID *string,
) (bool, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetObjectLegalHold", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	options := minio.GetObjectLegalHoldOptions{}

	if versionID != nil && *versionID != "" {
		options.VersionID = *versionID
		span.SetAttributes(attribute.String("storage.options.version", *versionID))
	}

	status, err := mc.client.GetObjectLegalHold(ctx, bucketName, objectName, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return false, serializeErrorResponse(err)
	}

	return status != nil && *status == minio.LegalHoldEnabled, nil
}

// SetObjectTags sets new object Tags to the given object, replaces/overwrites any existing tags.
func (mc *Client) SetObjectTags(
	ctx context.Context,
	bucketName string,
	objectName, versionID string,
	objectTags map[string]string,
) error {
	if len(objectTags) == 0 {
		return mc.RemoveObjectTags(ctx, bucketName, objectName, versionID)
	}

	ctx, span := mc.startOtelSpan(ctx, "SetObjectTags", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	options := minio.PutObjectTaggingOptions{}

	if versionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", versionID))
	}

	inputTags, err := tags.NewTags(objectTags, false)
	if err != nil {
		span.SetStatus(codes.Error, "failed to convert minio tags")
		span.RecordError(err)

		return schema.UnprocessableContentError(err.Error(), nil)
	}

	err = mc.client.PutObjectTagging(ctx, bucketName, objectName, inputTags, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// GetObjectTags fetches Object Tags from the given object.
func (mc *Client) GetObjectTags(
	ctx context.Context,
	bucketName, objectName string,
	versionID *string,
) (map[string]string, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetObjectTags", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	options := minio.GetObjectTaggingOptions{}

	if versionID != nil {
		options.VersionID = *versionID
		span.SetAttributes(attribute.String("storage.options.version", *versionID))
	}

	results, err := mc.client.GetObjectTagging(ctx, bucketName, objectName, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	return results.ToMap(), nil
}

// RemoveObjectTags removes Object Tags from the given object.
func (mc *Client) RemoveObjectTags(
	ctx context.Context,
	bucketName, objectName, versionID string,
) error {
	ctx, span := mc.startOtelSpan(ctx, "RemoveObjectTags", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	options := minio.RemoveObjectTaggingOptions{
		VersionID: versionID,
	}

	if versionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", versionID))
	}

	err := mc.client.RemoveObjectTagging(ctx, bucketName, objectName, options)
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
func (mc *Client) PresignedGetObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts common.PresignedGetStorageObjectOptions,
) (string, error) {
	return mc.presignObject(ctx, http.MethodGet, bucketName, objectName, opts)
}

// PresignedHeadObject generates a presigned URL for HTTP HEAD operations.
// Browsers/Mobile clients may point to this URL to directly get metadata from objects even if the bucket is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational. The default expiry is set to 7 days.
func (mc *Client) PresignedHeadObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts common.PresignedGetStorageObjectOptions,
) (string, error) {
	return mc.presignObject(ctx, http.MethodHead, bucketName, objectName, opts)
}

func (mc *Client) presignObject(
	ctx context.Context,
	method string,
	bucketName string,
	objectName string,
	opts common.PresignedGetStorageObjectOptions,
) (string, error) {
	ctx, span := mc.startOtelSpan(ctx, method+" PresignedObject", bucketName)
	defer span.End()

	reqParams := url.Values(common.KeyValuesToHeaders(opts.RequestParams))

	span.SetAttributes(
		attribute.String("storage.key", objectName),
		attribute.String("url.query", reqParams.Encode()),
	)

	expiry := time.Hour
	if opts.Expiry != nil && opts.Expiry.Duration > 0 {
		expiry = opts.Expiry.Duration
		span.SetAttributes(attribute.String("storage.expiry", opts.Expiry.String()))
	}

	fileName := filepath.Base(objectName)
	// Set request Parameters: for content-disposition.
	reqParams.Set(
		"response-content-disposition",
		fmt.Sprintf(`attachment; filename="%s"`, fileName),
	)

	var result *url.URL

	var err error

	header := http.Header{}

	if mc.publicHost != nil {
		header.Set("Host", mc.publicHost.Host)
	}

	result, err = mc.client.PresignHeader(
		ctx,
		method,
		bucketName,
		objectName,
		expiry,
		reqParams,
		header,
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", serializeErrorResponse(err)
	}

	if mc.publicHost != nil {
		result.Host = mc.publicHost.Host
		if mc.publicHost.Scheme != "" {
			result.Scheme = mc.publicHost.Scheme
		}
	}

	return result.String(), nil
}

// PresignedPutObject generates a presigned URL for HTTP PUT operations. Browsers/Mobile clients may point to this URL to upload objects directly to a bucket even if it is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational. The default expiry is set to 7 days.
func (mc *Client) PresignedPutObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	expiry time.Duration,
) (string, error) {
	ctx, span := mc.startOtelSpan(ctx, "PresignedPutObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))
	span.SetAttributes(attribute.String("storage.expiry", expiry.String()))

	header := http.Header{}

	if mc.publicHost != nil {
		header.Set("Host", mc.publicHost.Host)
	}

	result, err := mc.client.PresignHeader(
		ctx,
		http.MethodPut,
		bucketName,
		objectName,
		expiry,
		url.Values{},
		header,
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", serializeErrorResponse(err)
	}

	if mc.publicHost != nil {
		result.Host = mc.publicHost.Host
		if mc.publicHost.Scheme != "" {
			result.Scheme = mc.publicHost.Scheme
		}
	}

	return result.String(), nil
}

// Set object lock configuration in given bucket. mode, validity and unit are either all set or all nil.
func (mc *Client) SetObjectLockConfig(
	ctx context.Context,
	bucketname string,
	opts common.SetStorageObjectLockConfig,
) error {
	ctx, span := mc.startOtelSpan(ctx, "SetObjectLockConfig", bucketname)
	defer span.End()

	var retentionMode *minio.RetentionMode

	if opts.Mode != nil {
		span.SetAttributes(attribute.String("storage.lock_mode", string(*opts.Mode)))
		mode := validateObjectRetentionMode(*opts.Mode)
		retentionMode = &mode
	}

	if opts.Unit != nil {
		span.SetAttributes(attribute.String("storage.lock_unit", string(*opts.Unit)))
	}

	if opts.Validity != nil {
		span.SetAttributes(attribute.Int("storage.lock_validity", int(*opts.Validity)))
	}

	err := mc.client.SetObjectLockConfig(
		ctx,
		bucketname,
		retentionMode,
		opts.Validity,
		(*minio.ValidityUnit)(opts.Unit),
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// Get object lock configuration of given bucket.
func (mc *Client) GetObjectLockConfig(
	ctx context.Context,
	bucketName string,
) (*common.StorageObjectLockConfig, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetObjectLockConfig", bucketName)
	defer span.End()

	// ObjectLockConfigurationNotFoundError
	objectLock, mode, validity, unit, err := mc.client.GetObjectLockConfig(ctx, bucketName)
	if err != nil {
		respError := evalNotFoundError(err, "ObjectLockConfigurationNotFoundError")
		if respError == nil {
			return nil, nil
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, respError
	}

	result := &common.StorageObjectLockConfig{
		Enabled: objectLock == "Enabled",
		SetStorageObjectLockConfig: common.SetStorageObjectLockConfig{
			Mode:     serializeObjectRetentionMode(mode),
			Validity: validity,
			Unit:     (*common.StorageRetentionValidityUnit)(unit),
		},
	}

	return result, nil
}

func (mc *Client) populateObject(
	ctx context.Context,
	result *common.StorageObject,
	include common.StorageObjectIncludeOptions,
) error {
	if include.Tags {
		userTags, err := mc.GetObjectTags(ctx, result.Bucket, result.Name, result.VersionID)
		if err != nil {
			return err
		}

		result.Tags = common.StringMapToKeyValues(userTags)
	}

	if include.LegalHold {
		lhStatus, err := mc.GetObjectLegalHold(ctx, result.Bucket, result.Name, result.VersionID)
		if err == nil {
			result.LegalHold = &lhStatus
		}
	}

	return nil
}
