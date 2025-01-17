package gcs

import (
	"encoding/base64"
	"errors"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/api/googleapi"
)

const (
	size256K = 256 * 1024
)

var errNotSupported = schema.NotSupportedError("Google Cloud Storage doesn't support this method", nil)

func serializeBucketInfo(bucket *storage.BucketAttrs) common.StorageBucketInfo {
	result := common.StorageBucketInfo{
		Name:         bucket.Name,
		Tags:         bucket.Labels,
		CreationDate: bucket.Created,
		Lifecycle:    serializeLifecycleConfiguration(bucket.Lifecycle),
		Versioning: &common.StorageBucketVersioningConfiguration{
			Enabled: bucket.VersioningEnabled,
		},
	}

	if bucket.Encryption != nil && bucket.Encryption.DefaultKMSKeyName != "" {
		result.Encryption = &common.ServerSideEncryptionConfiguration{
			KmsMasterKeyID: bucket.Encryption.DefaultKMSKeyName,
		}
	}

	return result
}

func serializeObjectInfo(obj *storage.ObjectAttrs) common.StorageObject {
	object := common.StorageObject{
		Bucket:       obj.Bucket,
		Name:         obj.Name,
		CreationTime: &obj.Created,
		LastModified: obj.Updated,
		Size:         &obj.Size,
		Metadata:     obj.Metadata,
		StorageClass: &obj.StorageClass,
		LegalHold:    &obj.TemporaryHold,
	}

	if obj.Etag != "" {
		object.ETag = &obj.Etag
	}

	if obj.ContentType != "" {
		object.ContentType = &obj.ContentType
	}

	if obj.CacheControl != "" {
		object.CacheControl = &obj.CacheControl
	}

	if obj.ContentDisposition != "" {
		object.ContentDisposition = &obj.ContentDisposition
	}

	if obj.ContentEncoding != "" {
		object.ContentEncoding = &obj.ContentEncoding
	}

	if obj.ContentLanguage != "" {
		object.ContentLanguage = &obj.ContentLanguage
	}

	if obj.CustomerKeySHA256 != "" {
		object.CustomerProvidedKeySHA256 = &obj.CustomerKeySHA256
	}

	if obj.KMSKeyName != "" {
		object.KMSKeyName = &obj.KMSKeyName
	}

	if obj.Owner != "" {
		object.Owner = &common.StorageOwner{
			DisplayName: &obj.Owner,
		}
	}

	if obj.Retention != nil {
		if obj.Retention.Mode != "" {
			object.RetentionMode = &obj.Retention.Mode
		}

		if !obj.Retention.RetainUntil.IsZero() {
			object.RetentionUntilDate = &obj.Retention.RetainUntil
		}
	}

	if !obj.RetentionExpirationTime.IsZero() {
		object.Expiration = &obj.RetentionExpirationTime
	}

	if !obj.Deleted.IsZero() {
		deleted := true
		object.Deleted = &deleted
		object.DeletedTime = &obj.Deleted
	}

	if obj.Generation > 0 {
		versionID := strconv.Itoa(int(obj.Generation))
		object.VersionID = &versionID
	}

	if len(obj.MD5) > 0 {
		contentMd5 := base64.StdEncoding.EncodeToString(obj.MD5)
		object.ContentMD5 = &contentMd5
	}

	return object
}

func (c *Client) validateListObjectsOptions(span trace.Span, opts *common.ListStorageObjectsOptions) *storage.Query {
	span.SetAttributes(
		attribute.Bool("storage.options.recursive", opts.Recursive),
		attribute.Bool("storage.options.with_versions", opts.Include.Versions),
		attribute.Bool("storage.options.with_metadata", opts.Include.Metadata),
	)

	if opts.Prefix != "" {
		span.SetAttributes(attribute.String("storage.options.prefix", opts.Prefix))
	}

	if opts.StartAfter != "" {
		span.SetAttributes(attribute.String("storage.options.start_after", opts.StartAfter))
	}

	if opts.MaxResults > 0 {
		span.SetAttributes(attribute.Int("storage.options.max_results", opts.MaxResults))
	}

	return &storage.Query{
		Versions:    opts.Include.Versions,
		Prefix:      opts.Prefix,
		StartOffset: opts.StartAfter,
	}
}

func serializeUploadObjectInfo(obj *storage.Writer) common.StorageUploadInfo {
	object := common.StorageUploadInfo{
		Bucket: obj.Bucket,
		Name:   obj.Name,
	}

	if obj.Etag != "" {
		object.ETag = &obj.Etag
	}

	if obj.Size > 0 {
		object.Size = &obj.Size
	}

	if !obj.Updated.IsZero() {
		object.LastModified = &obj.Updated
	} else if !obj.Created.IsZero() {
		object.LastModified = &obj.Created
	}

	if !obj.RetentionExpirationTime.IsZero() {
		object.Expiration = &obj.RetentionExpirationTime
	}

	versionID := strconv.Itoa(int(obj.Generation))
	object.VersionID = &versionID

	if len(obj.MD5) > 0 {
		contentMd5 := base64.StdEncoding.EncodeToString(obj.MD5)
		object.ContentMD5 = &contentMd5
	}

	return object
}

func evalGoogleErrorResponse(err *googleapi.Error) *schema.ConnectorError {
	details := map[string]any{
		"statusCode": err.Code,
		"details":    err.Details,
	}

	if err.Code >= 500 {
		return schema.NewConnectorError(err.Code, err.Message, details)
	}

	return schema.UnprocessableContentError(err.Message, details)
}

func serializeErrorResponse(err error) *schema.ConnectorError {
	var e *googleapi.Error
	if ok := errors.As(err, &e); ok {
		return evalGoogleErrorResponse(e)
	}

	return schema.UnprocessableContentError(err.Error(), nil)
}
