package minio

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/tags"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// ListObjects list objects in a bucket.
func (mc *Client) ListObjects(ctx context.Context, bucketName string, opts *common.ListStorageObjectsOptions, predicate func(string) bool) (*common.StorageObjectListResults, error) {
	ctx, span := mc.startOtelSpan(ctx, "ListObjects", bucketName)
	defer span.End()

	maxResults := opts.MaxResults
	// Do not set the limit if the post-predicate function exists.
	// Results will be filtered and paginated by the client.
	if predicate != nil {
		opts.MaxResults = 0
	}

	var count int
	objChan := mc.client.ListObjects(ctx, bucketName, mc.validateListObjectsOptions(span, opts))
	objects := make([]common.StorageObject, 0)

	for obj := range objChan {
		if obj.Err != nil {
			span.SetStatus(codes.Error, obj.Err.Error())
			span.RecordError(obj.Err)

			return nil, serializeErrorResponse(obj.Err)
		}

		if predicate != nil && !predicate(obj.Key) {
			continue
		}

		if maxResults > 0 && count < maxResults {
			object := serializeObjectInfo(obj, true)
			object.Bucket = bucketName

			if opts.Include.LegalHold {
				lhStatus, err := mc.GetObjectLegalHold(ctx, bucketName, object.Name, object.VersionID)
				if err == nil {
					object.LegalHold = &lhStatus
				}
			}

			objects = append(objects, object)
			count++
		}
	}

	span.SetAttributes(attribute.Int("storage.object_count", count))

	return &common.StorageObjectListResults{
		Objects: objects,
	}, nil
}

// ListIncompleteUploads list partially uploaded objects in a bucket.
func (mc *Client) ListIncompleteUploads(ctx context.Context, bucketName string, args common.ListIncompleteUploadsOptions) ([]common.StorageObjectMultipartInfo, error) {
	ctx, span := mc.startOtelSpan(ctx, "ListIncompleteUploads", bucketName)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.object_prefix", args.Prefix),
		attribute.Bool("storage.options.recursive", args.Recursive),
	)

	objChan := mc.client.ListIncompleteUploads(ctx, bucketName, args.Prefix, args.Recursive)
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

// RemoveIncompleteUpload removes a partially uploaded object.
func (mc *Client) RemoveIncompleteUpload(ctx context.Context, bucketName string, objectName string) error {
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
func (mc *Client) GetObject(ctx context.Context, bucketName, objectName string, opts common.GetStorageObjectOptions) (io.ReadCloser, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))
	options := serializeGetObjectOptions(span, opts)

	object, err := mc.client.GetObject(ctx, bucketName, objectName, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	return object, nil
}

// PutObject uploads objects that are less than 128MiB in a single PUT operation. For objects that are greater than 128MiB in size,
// PutObject seamlessly uploads the object as parts of 128MiB or more depending on the actual file size. The max upload size for an object is 5TB.
func (mc *Client) PutObject(ctx context.Context, bucketName string, objectName string, opts *common.PutStorageObjectOptions, reader io.Reader, objectSize int64) (*common.StorageUploadInfo, error) {
	ctx, span := mc.startOtelSpan(ctx, "PutObject", bucketName)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", objectName),
		attribute.Int64("http.response.body.size", objectSize),
	)

	options := minio.PutObjectOptions{
		UserMetadata:            opts.UserMetadata,
		UserTags:                opts.UserTags,
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

	if opts.RetainUntilDate != nil {
		options.RetainUntilDate = *opts.RetainUntilDate
		span.SetAttributes(attribute.String("storage.options.retain_util_date", opts.RetainUntilDate.Format(time.RFC3339)))
	}

	if opts.Mode != nil {
		mode := minio.RetentionMode(string(*opts.Mode))
		if !mode.IsValid() {
			errorMsg := fmt.Sprintf("invalid RetentionMode: %s", *opts.Mode)
			span.SetStatus(codes.Error, errorMsg)

			return nil, schema.UnprocessableContentError(errorMsg, nil)
		}

		options.Mode = mode
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
func (mc *Client) CopyObject(ctx context.Context, dest common.StorageCopyDestOptions, src common.StorageCopySrcOptions) (*common.StorageUploadInfo, error) {
	ctx, span := mc.startOtelSpan(ctx, "CopyObject", dest.Bucket)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", dest.Object),
		attribute.String("storage.copy_source", src.Object),
	)

	destOptions, err := convertCopyDestOptions(dest)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

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
func (mc *Client) ComposeObject(ctx context.Context, dest common.StorageCopyDestOptions, sources []common.StorageCopySrcOptions) (*common.StorageUploadInfo, error) {
	ctx, span := mc.startOtelSpan(ctx, "ComposeObject", dest.Bucket)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", dest.Object))

	srcKeys := make([]string, len(sources))
	srcOptions := make([]minio.CopySrcOptions, len(sources))

	for i, src := range sources {
		srcKeys[i] = src.Object
		source := serializeCopySourceOptions(src)
		srcOptions[i] = source
	}

	span.SetAttributes(attribute.StringSlice("storage.copy_sources", srcKeys))

	destOptions, err := convertCopyDestOptions(dest)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

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
func (mc *Client) StatObject(ctx context.Context, bucketName, objectName string, opts common.GetStorageObjectOptions) (*common.StorageObject, error) {
	ctx, span := mc.startOtelSpan(ctx, "StatObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))
	options := serializeGetObjectOptions(span, opts)

	object, err := mc.client.StatObject(ctx, bucketName, objectName, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := serializeObjectInfo(object, false)
	result.Bucket = bucketName

	if opts.Include.Tags {
		userTags, err := mc.GetObjectTagging(ctx, bucketName, objectName, opts.VersionID)
		if err != nil {
			return nil, err
		}

		result.UserTags = userTags
	}

	if opts.Include.LegalHold {
		var versionID *string
		if object.VersionID != "" {
			versionID = &object.VersionID
		}

		lhStatus, err := mc.GetObjectLegalHold(ctx, bucketName, object.Key, versionID)
		if err == nil {
			result.LegalHold = &lhStatus
		}
	}

	common.SetObjectInfoSpanAttributes(span, &result)

	return &result, nil
}

// RemoveObject removes an object with some specified options.
func (mc *Client) RemoveObject(ctx context.Context, bucketName string, objectName string, opts common.RemoveStorageObjectOptions) error {
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
func (mc *Client) RemoveObjects(ctx context.Context, bucketName string, opts *common.RemoveStorageObjectsOptions, predicate func(string) bool) []common.RemoveStorageObjectError {
	ctx, span := mc.startOtelSpan(ctx, "RemoveObjects", bucketName)
	defer span.End()

	listOptions := mc.validateListObjectsOptions(span, &opts.ListStorageObjectsOptions)
	span.SetAttributes(attribute.Bool("storage.options.governance_bypass", opts.GovernanceBypass))

	objectChan := mc.client.ListObjects(ctx, bucketName, listOptions)

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
			Error:      err.Err,
		})
	}

	return errs
}

// PutObjectRetention applies object retention lock onto an object.
func (mc *Client) PutObjectRetention(ctx context.Context, opts *common.PutStorageObjectRetentionOptions) error {
	ctx, span := mc.startOtelSpan(ctx, "PutObjectRetention", opts.Bucket)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", opts.Object),
		attribute.Bool("storage.options.governance_bypass", opts.GovernanceBypass),
	)

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))
	}

	if opts.RetainUntilDate != nil {
		span.SetAttributes(attribute.String("storage.options.retain_util_date", opts.RetainUntilDate.Format(time.RFC3339)))
	}

	options := minio.PutObjectRetentionOptions{
		GovernanceBypass: opts.GovernanceBypass,
		VersionID:        opts.VersionID,
		RetainUntilDate:  opts.RetainUntilDate,
	}

	if opts.Mode != nil {
		mode := minio.RetentionMode(string(*opts.Mode))
		if !mode.IsValid() {
			errorMsg := fmt.Sprintf("invalid RetentionMode: %s", *opts.Mode)
			span.SetStatus(codes.Error, errorMsg)

			return schema.UnprocessableContentError(errorMsg, nil)
		}

		options.Mode = &mode
	}

	err := mc.client.PutObjectRetention(ctx, opts.Bucket, opts.Object, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// SetObjectLegalHold applies legal-hold onto an object.
func (mc *Client) SetObjectLegalHold(ctx context.Context, bucketName string, objectName string, opts common.SetStorageObjectLegalHoldOptions) error {
	ctx, span := mc.startOtelSpan(ctx, "SetObjectLegalHold", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	options := minio.PutObjectLegalHoldOptions{
		VersionID: opts.VersionID,
	}

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))
	}

	if opts.Status != nil {
		span.SetAttributes(attribute.Bool("storage.options.status", *opts.Status))
		legalHold := validateLegalHoldStatus(opts.Status)
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
func (mc *Client) GetObjectLegalHold(ctx context.Context, bucketName string, objectName string, versionID *string) (bool, error) {
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

// PutObjectTagging sets new object Tags to the given object, replaces/overwrites any existing tags.
func (mc *Client) SetObjectTags(ctx context.Context, bucketName string, objectName string, opts common.SetStorageObjectTagsOptions) error {
	if len(opts.Tags) == 0 {
		return mc.RemoveObjectTagging(ctx, bucketName, objectName, opts.VersionID)
	}

	ctx, span := mc.startOtelSpan(ctx, "SetStorageObjectTags", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	options := minio.PutObjectTaggingOptions{}

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))
	}

	inputTags, err := tags.NewTags(opts.Tags, false)
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

// GetObjectTagging fetches Object Tags from the given object.
func (mc *Client) GetObjectTagging(ctx context.Context, bucketName, objectName string, versionID *string) (map[string]string, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetObjectTagging", bucketName)
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

// RemoveObjectTagging removes Object Tags from the given object.
func (mc *Client) RemoveObjectTagging(ctx context.Context, bucketName, objectName, versionID string) error {
	ctx, span := mc.startOtelSpan(ctx, "RemoveObjectTagging", bucketName)
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
func (mc *Client) PresignedGetObject(ctx context.Context, bucketName string, objectName string, opts common.PresignedGetStorageObjectOptions) (string, error) {
	return mc.presignObject(ctx, http.MethodGet, bucketName, objectName, opts)
}

// PresignedHeadObject generates a presigned URL for HTTP HEAD operations.
// Browsers/Mobile clients may point to this URL to directly get metadata from objects even if the bucket is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational. The default expiry is set to 7 days.
func (mc *Client) PresignedHeadObject(ctx context.Context, bucketName string, objectName string, opts common.PresignedGetStorageObjectOptions) (string, error) {
	return mc.presignObject(ctx, http.MethodHead, bucketName, objectName, opts)
}

func (mc *Client) presignObject(ctx context.Context, method string, bucketName string, objectName string, opts common.PresignedGetStorageObjectOptions) (string, error) {
	ctx, span := mc.startOtelSpan(ctx, method+" PresignedObject", bucketName)
	defer span.End()

	reqParams := url.Values{}

	for key, params := range opts.RequestParams {
		for _, param := range params {
			reqParams.Add(key, param)
		}
	}

	span.SetAttributes(
		attribute.String("storage.key", objectName),
		attribute.String("url.query", reqParams.Encode()),
	)

	if opts.Expiry != nil {
		span.SetAttributes(attribute.String("storage.expiry", opts.Expiry.String()))
	}

	fileName := filepath.Base(objectName)
	// Set request Parameters: for content-disposition.
	reqParams.Set("response-content-disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))

	var result *url.URL
	var err error
	header := http.Header{}

	if mc.publicHost != nil {
		header.Set("Host", mc.publicHost.Host)
	}

	result, err = mc.client.PresignHeader(ctx, method, bucketName, objectName, opts.Expiry.Duration, reqParams, header)
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
func (mc *Client) PresignedPutObject(ctx context.Context, bucketName string, objectName string, expiry time.Duration) (string, error) {
	ctx, span := mc.startOtelSpan(ctx, "PresignedPutObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))
	span.SetAttributes(attribute.String("storage.expiry", expiry.String()))

	header := http.Header{}

	if mc.publicHost != nil {
		header.Set("Host", mc.publicHost.Host)
	}

	result, err := mc.client.PresignHeader(ctx, http.MethodPut, bucketName, objectName, expiry, url.Values{}, header)
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
func (mc *Client) SetObjectLockConfig(ctx context.Context, bucketname string, opts common.SetStorageObjectLockConfig) error {
	ctx, span := mc.startOtelSpan(ctx, "SetObjectLockConfig", bucketname)
	defer span.End()

	if opts.Mode != nil {
		span.SetAttributes(attribute.String("storage.lock_mode", string(*opts.Mode)))
	}

	if opts.Unit != nil {
		span.SetAttributes(attribute.String("storage.lock_unit", string(*opts.Unit)))
	}

	if opts.Validity != nil {
		span.SetAttributes(attribute.Int("storage.lock_validity", int(*opts.Validity)))
	}

	err := mc.client.SetObjectLockConfig(ctx, bucketname, (*minio.RetentionMode)(opts.Mode), opts.Validity, (*minio.ValidityUnit)(opts.Unit))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// Get object lock configuration of given bucket.
func (mc *Client) GetObjectLockConfig(ctx context.Context, bucketName string) (*common.StorageObjectLockConfig, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetObjectLockConfig", bucketName)
	defer span.End()

	objectLock, mode, validity, unit, err := mc.client.GetObjectLockConfig(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := &common.StorageObjectLockConfig{
		ObjectLock: objectLock,
		SetStorageObjectLockConfig: common.SetStorageObjectLockConfig{
			Mode:     (*common.StorageRetentionMode)(mode),
			Validity: validity,
			Unit:     (*common.StorageRetentionValidityUnit)(unit),
		},
	}

	return result, nil
}
