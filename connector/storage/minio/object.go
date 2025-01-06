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
func (mc *Client) ListObjects(ctx context.Context, opts *common.ListStorageObjectsOptions) ([]common.StorageObject, error) {
	ctx, span := mc.startOtelSpan(ctx, "ListObjects", opts.Bucket)
	defer span.End()

	objChan := mc.client.ListObjects(ctx, opts.Bucket, serializeListObjectsOptions(span, opts))
	objects := make([]common.StorageObject, 0)

	for obj := range objChan {
		if obj.Err != nil {
			span.SetStatus(codes.Error, obj.Err.Error())
			span.RecordError(obj.Err)

			return nil, serializeErrorResponse(obj.Err)
		}

		object := serializeObjectInfo(obj)
		object.Bucket = opts.Bucket
		objects = append(objects, object)
	}

	span.SetAttributes(attribute.Int("storage.object_count", len(objects)))

	return objects, nil
}

// ListIncompleteUploads list partially uploaded objects in a bucket.
func (mc *Client) ListIncompleteUploads(ctx context.Context, args *common.ListIncompleteUploadsArguments) ([]common.StorageObjectMultipartInfo, error) {
	ctx, span := mc.startOtelSpan(ctx, "ListIncompleteUploads", args.Bucket)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.object_prefix", args.Prefix),
		attribute.Bool("storage.options.recursive", args.Recursive),
	)

	objChan := mc.client.ListIncompleteUploads(ctx, args.Bucket, args.Prefix, args.Recursive)
	objects := make([]common.StorageObjectMultipartInfo, 0)

	for obj := range objChan {
		if obj.Err != nil {
			span.SetStatus(codes.Error, obj.Err.Error())
			span.RecordError(obj.Err)

			return nil, serializeErrorResponse(obj.Err)
		}

		object := common.StorageObjectMultipartInfo{
			Key:          obj.Key,
			Initiated:    &obj.Initiated,
			StorageClass: obj.StorageClass,
			Size:         obj.Size,
			UploadID:     obj.UploadID,
		}

		objects = append(objects, object)
	}

	span.SetAttributes(attribute.Int("storage.object_count", len(objects)))

	return objects, nil
}

// RemoveIncompleteUpload removes a partially uploaded object.
func (mc *Client) RemoveIncompleteUpload(ctx context.Context, args *common.RemoveIncompleteUploadArguments) error {
	ctx, span := mc.startOtelSpan(ctx, "RemoveIncompleteUpload", args.Bucket)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", args.Object))

	err := mc.client.RemoveIncompleteUpload(ctx, args.Bucket, args.Object)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// GetObject returns a stream of the object data. Most of the common errors occur when reading the stream.
func (mc *Client) GetObject(ctx context.Context, opts *common.GetStorageObjectOptions) (io.ReadCloser, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetObject", opts.Bucket)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", opts.Object))
	options := serializeGetObjectOptions(span, opts)

	object, err := mc.client.GetObject(ctx, opts.Bucket, opts.Object, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	return object, nil
}

// PutObject uploads objects that are less than 128MiB in a single PUT operation. For objects that are greater than 128MiB in size,
// PutObject seamlessly uploads the object as parts of 128MiB or more depending on the actual file size. The max upload size for an object is 5TB.
func (mc *Client) PutObject(ctx context.Context, args *common.PutStorageObjectArguments, reader io.Reader, objectSize int64) (*common.StorageUploadInfo, error) {
	ctx, span := mc.startOtelSpan(ctx, "PutObject", args.Bucket)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", args.Object),
		attribute.Int64("http.response.body.size", objectSize),
	)

	options := minio.PutObjectOptions{
		UserMetadata:            args.Options.UserMetadata,
		UserTags:                args.Options.UserTags,
		ContentType:             args.Options.ContentType,
		ContentEncoding:         args.Options.ContentEncoding,
		ContentDisposition:      args.Options.ContentDisposition,
		ContentLanguage:         args.Options.ContentLanguage,
		CacheControl:            args.Options.CacheControl,
		NumThreads:              args.Options.NumThreads,
		StorageClass:            args.Options.StorageClass,
		PartSize:                args.Options.PartSize,
		SendContentMd5:          args.Options.SendContentMd5,
		DisableContentSha256:    args.Options.DisableContentSha256,
		DisableMultipart:        args.Options.DisableMultipart,
		WebsiteRedirectLocation: args.Options.WebsiteRedirectLocation,
		ConcurrentStreamParts:   args.Options.ConcurrentStreamParts,
	}

	if args.Options.Expires != nil {
		options.Expires = *args.Options.Expires
	}

	if args.Options.RetainUntilDate != nil {
		options.RetainUntilDate = *args.Options.RetainUntilDate
		span.SetAttributes(attribute.String("storage.options.retain_util_date", args.Options.RetainUntilDate.Format(time.RFC3339)))
	}

	if args.Options.Mode != nil {
		mode := minio.RetentionMode(string(*args.Options.Mode))
		if !mode.IsValid() {
			errorMsg := fmt.Sprintf("invalid RetentionMode: %s", *args.Options.Mode)
			span.SetStatus(codes.Error, errorMsg)

			return nil, schema.UnprocessableContentError(errorMsg, nil)
		}

		options.Mode = mode
	}

	if args.Options.LegalHold != nil {
		legalHold := minio.LegalHoldStatus(*args.Options.LegalHold)
		if !legalHold.IsValid() {
			errorMsg := fmt.Sprintf("invalid LegalHoldStatus: %s", *args.Options.LegalHold)
			span.SetStatus(codes.Error, errorMsg)

			return nil, schema.UnprocessableContentError(errorMsg, nil)
		}

		options.LegalHold = legalHold
	}

	if args.Options.Checksum != nil {
		options.Checksum = parseChecksumType(*args.Options.Checksum)
	}

	if args.Options.AutoChecksum != nil {
		options.AutoChecksum = parseChecksumType(*args.Options.AutoChecksum)
	}

	object, err := mc.client.PutObject(ctx, args.Bucket, args.Object, reader, objectSize, options)
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
func (mc *Client) StatObject(ctx context.Context, opts *common.GetStorageObjectOptions) (*common.StorageObject, error) {
	ctx, span := mc.startOtelSpan(ctx, "StatObject", opts.Bucket)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", opts.Object))
	options := serializeGetObjectOptions(span, opts)

	object, err := mc.client.StatObject(ctx, opts.Bucket, opts.Object, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := serializeObjectInfo(object)
	result.Bucket = opts.Bucket
	common.SetObjectInfoSpanAttributes(span, &result)

	return &result, nil
}

// RemoveObject removes an object with some specified options.
func (mc *Client) RemoveObject(ctx context.Context, opts *common.RemoveStorageObjectOptions) error {
	ctx, span := mc.startOtelSpan(ctx, "RemoveObject", opts.Bucket)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", opts.Object),
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

	err := mc.client.RemoveObject(ctx, opts.Bucket, opts.Object, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// RemoveObjects removes a list of objects obtained from an input channel. The call sends a delete request to the server up to 1000 objects at a time.
// The errors observed are sent over the error channel.
func (mc *Client) RemoveObjects(ctx context.Context, opts *common.RemoveStorageObjectsOptions) []common.RemoveStorageObjectError {
	ctx, span := mc.startOtelSpan(ctx, "RemoveObjects", opts.Bucket)
	defer span.End()

	listOptions := serializeListObjectsOptions(span, &opts.ListStorageObjectsOptions)
	span.SetAttributes(attribute.Bool("storage.options.governance_bypass", opts.GovernanceBypass))

	objectChan := mc.client.ListObjects(ctx, opts.Bucket, listOptions)

	options := minio.RemoveObjectsOptions{
		GovernanceBypass: opts.GovernanceBypass,
	}

	errChan := mc.client.RemoveObjects(ctx, opts.Bucket, objectChan, options)
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

// PutObjectLegalHold applies legal-hold onto an object.
func (mc *Client) PutObjectLegalHold(ctx context.Context, opts *common.PutStorageObjectLegalHoldOptions) error {
	ctx, span := mc.startOtelSpan(ctx, "PutObjectLegalHold", opts.Bucket)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", opts.Object))

	options := minio.PutObjectLegalHoldOptions{
		VersionID: opts.VersionID,
	}

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))
	}

	if opts.Status != nil {
		span.SetAttributes(attribute.String("storage.options.status", string(*opts.Status)))

		legalHold := minio.LegalHoldStatus(*opts.Status)
		if !legalHold.IsValid() {
			errorMsg := "invalid LegalHoldStatus: " + string(*opts.Status)
			span.SetStatus(codes.Error, errorMsg)

			return schema.UnprocessableContentError(errorMsg, nil)
		}

		options.Status = &legalHold
	}

	err := mc.client.PutObjectLegalHold(ctx, opts.Bucket, opts.Object, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// GetObjectLegalHold returns legal-hold status on a given object.
func (mc *Client) GetObjectLegalHold(ctx context.Context, opts *common.GetStorageObjectLegalHoldOptions) (common.StorageLegalHoldStatus, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetObjectLegalHold", opts.Bucket)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", opts.Object))

	options := minio.GetObjectLegalHoldOptions{
		VersionID: opts.VersionID,
	}

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))
	}

	status, err := mc.client.GetObjectLegalHold(ctx, opts.Bucket, opts.Object, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", serializeErrorResponse(err)
	}

	if status == nil {
		return "", nil
	}

	result := common.StorageLegalHoldStatus(string(*status))

	return result, nil
}

// PutObjectTagging sets new object Tags to the given object, replaces/overwrites any existing tags.
func (mc *Client) PutObjectTagging(ctx context.Context, opts *common.PutStorageObjectTaggingOptions) error {
	ctx, span := mc.startOtelSpan(ctx, "PutObjectTagging", opts.Bucket)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", opts.Object))

	options := minio.PutObjectTaggingOptions{
		VersionID: opts.VersionID,
	}

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))
	}

	inputTags, err := tags.NewTags(opts.Tags, false)
	if err != nil {
		span.SetStatus(codes.Error, "failed to convert minio tags")
		span.RecordError(err)

		return schema.UnprocessableContentError(err.Error(), nil)
	}

	err = mc.client.PutObjectTagging(ctx, opts.Bucket, opts.Object, inputTags, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// GetObjectTagging fetches Object Tags from the given object.
func (mc *Client) GetObjectTagging(ctx context.Context, opts *common.StorageObjectTaggingOptions) (map[string]string, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetObjectTagging", opts.Bucket)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", opts.Object))

	options := minio.GetObjectTaggingOptions{
		VersionID: opts.VersionID,
	}

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))
	}

	results, err := mc.client.GetObjectTagging(ctx, opts.Bucket, opts.Object, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	return results.ToMap(), nil
}

// RemoveObjectTagging removes Object Tags from the given object.
func (mc *Client) RemoveObjectTagging(ctx context.Context, opts *common.StorageObjectTaggingOptions) error {
	ctx, span := mc.startOtelSpan(ctx, "RemoveObjectTagging", opts.Bucket)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", opts.Object))

	options := minio.RemoveObjectTaggingOptions{
		VersionID: opts.VersionID,
	}

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))
	}

	err := mc.client.RemoveObjectTagging(ctx, opts.Bucket, opts.Object, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// GetObjectAttributes returns a stream of the object data. Most of the common errors occur when reading the stream.
func (mc *Client) GetObjectAttributes(ctx context.Context, opts *common.StorageObjectAttributesOptions) (*common.StorageObjectAttributes, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetObjectAttributes", opts.Bucket)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", opts.Object))

	options := minio.ObjectAttributesOptions{
		VersionID:        opts.VersionID,
		MaxParts:         opts.MaxParts,
		PartNumberMarker: opts.PartNumberMarker,
	}

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))
	}

	attrs, err := mc.client.GetObjectAttributes(ctx, opts.Bucket, opts.Object, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	checksum := common.StorageObjectChecksum{}
	if !isStringNull(attrs.Checksum.ChecksumCRC32) {
		checksum.ChecksumCRC32 = &attrs.Checksum.ChecksumCRC32
	}

	if !isStringNull(attrs.Checksum.ChecksumCRC32C) {
		checksum.ChecksumCRC32C = &attrs.Checksum.ChecksumCRC32C
	}

	if !isStringNull(attrs.Checksum.ChecksumSHA1) {
		checksum.ChecksumSHA1 = &attrs.Checksum.ChecksumSHA1
	}

	if !isStringNull(attrs.Checksum.ChecksumSHA256) {
		checksum.ChecksumSHA256 = &attrs.Checksum.ChecksumSHA256
	}

	result := &common.StorageObjectAttributes{
		LastModified: attrs.LastModified,
		StorageObjectAttributesResponse: common.StorageObjectAttributesResponse{
			ETag:         attrs.ETag,
			StorageClass: attrs.StorageClass,
			ObjectSize:   attrs.ObjectSize,
			Checksum:     checksum,
			ObjectParts: common.StorageObjectParts{
				PartsCount:           attrs.ObjectParts.PartsCount,
				PartNumberMarker:     attrs.ObjectParts.PartNumberMarker,
				NextPartNumberMarker: attrs.ObjectParts.NextPartNumberMarker,
				MaxParts:             attrs.ObjectParts.MaxParts,
				IsTruncated:          attrs.ObjectParts.IsTruncated,
				Parts:                make([]*common.StorageObjectAttributePart, len(attrs.ObjectParts.Parts)),
			},
		},
	}

	if !isStringNull(attrs.VersionID) {
		result.VersionID = &attrs.VersionID
	}

	for i, p := range attrs.ObjectParts.Parts {
		partChecksum := common.StorageObjectChecksum{}
		if !isStringNull(p.ChecksumCRC32) {
			partChecksum.ChecksumCRC32 = &p.ChecksumCRC32
		}

		if !isStringNull(p.ChecksumCRC32C) {
			partChecksum.ChecksumCRC32C = &p.ChecksumCRC32C
		}

		if !isStringNull(p.ChecksumSHA1) {
			partChecksum.ChecksumSHA1 = &p.ChecksumSHA1
		}

		if !isStringNull(p.ChecksumSHA256) {
			partChecksum.ChecksumSHA256 = &p.ChecksumSHA256
		}

		result.ObjectParts.Parts[i] = &common.StorageObjectAttributePart{
			StorageObjectChecksum: partChecksum,
			PartNumber:            p.PartNumber,
			Size:                  p.Size,
		}
	}

	return result, nil
}

// PresignedGetObject generates a presigned URL for HTTP GET operations. Browsers/Mobile clients may point to this URL to directly download objects even if the bucket is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The maximum expiry is 604800 seconds (i.e. 7 days) and minimum is 1 second.
func (mc *Client) PresignedGetObject(ctx context.Context, args *common.PresignedGetStorageObjectArguments) (*url.URL, error) {
	return mc.presignObject(ctx, http.MethodGet, args)
}

// PresignedHeadObject generates a presigned URL for HTTP HEAD operations.
// Browsers/Mobile clients may point to this URL to directly get metadata from objects even if the bucket is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational. The default expiry is set to 7 days.
func (mc *Client) PresignedHeadObject(ctx context.Context, args *common.PresignedGetStorageObjectArguments) (*url.URL, error) {
	return mc.presignObject(ctx, http.MethodHead, args)
}

func (mc *Client) presignObject(ctx context.Context, method string, args *common.PresignedGetStorageObjectArguments) (*url.URL, error) {
	ctx, span := mc.startOtelSpan(ctx, method+" PresignedObject", args.Bucket)
	defer span.End()

	reqParams := url.Values{}

	for key, params := range args.RequestParams {
		for _, param := range params {
			reqParams.Add(key, param)
		}
	}

	span.SetAttributes(
		attribute.String("storage.key", args.Object),
		attribute.String("url.query", reqParams.Encode()),
	)

	if args.Expiry != nil {
		span.SetAttributes(attribute.String("storage.expiry", args.Expiry.String()))
	}

	fileName := filepath.Base(args.Object)
	// Set request Parameters: for content-disposition.
	reqParams.Set("response-content-disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))

	var result *url.URL
	var err error
	header := http.Header{}

	if mc.publicHost != nil {
		header.Set("Host", mc.publicHost.Host)
	}

	result, err = mc.client.PresignHeader(ctx, method, args.Bucket, args.Object, args.Expiry.Duration, reqParams, header)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	if mc.publicHost != nil {
		result.Host = mc.publicHost.Host
		if mc.publicHost.Scheme != "" {
			result.Scheme = mc.publicHost.Scheme
		}
	}

	return result, nil
}

// PresignedPutObject generates a presigned URL for HTTP PUT operations. Browsers/Mobile clients may point to this URL to upload objects directly to a bucket even if it is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational. The default expiry is set to 7 days.
func (mc *Client) PresignedPutObject(ctx context.Context, args *common.PresignedPutStorageObjectArguments) (*url.URL, error) {
	ctx, span := mc.startOtelSpan(ctx, "PresignedPutObject", args.Bucket)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", args.Object))

	if args.Expiry != nil {
		span.SetAttributes(attribute.String("storage.expiry", args.Expiry.String()))
	}

	header := http.Header{}

	if mc.publicHost != nil {
		header.Set("Host", mc.publicHost.Host)
	}

	result, err := mc.client.PresignHeader(ctx, http.MethodPut, args.Bucket, args.Object, args.Expiry.Duration, url.Values{}, header)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	if mc.publicHost != nil {
		result.Host = mc.publicHost.Host
		if mc.publicHost.Scheme != "" {
			result.Scheme = mc.publicHost.Scheme
		}
	}

	return result, nil
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
