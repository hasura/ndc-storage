package common

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// NewDBSystemAttribute creates the detault db.system attribute.
func NewDBSystemAttribute() attribute.KeyValue {
	return attribute.String("db.system", "storage")
}

// SetObjectChecksumSpanAttributes sets span attributes from the object checksum.
func SetObjectChecksumSpanAttributes(span trace.Span, object *StorageObjectChecksum) {
	if object.ChecksumCRC32 != nil {
		span.SetAttributes(attribute.String("storage.object.checksum_crc32", *object.ChecksumCRC32))
	}

	if object.ChecksumCRC32C != nil {
		span.SetAttributes(attribute.String("storage.object.checksum_crc32c", *object.ChecksumCRC32C))
	}

	if object.ChecksumCRC64NVME != nil {
		span.SetAttributes(attribute.String("storage.object.checksum_crc64nvme", *object.ChecksumCRC64NVME))
	}

	if object.ChecksumSHA1 != nil {
		span.SetAttributes(attribute.String("storage.object.checksum_sha1", *object.ChecksumSHA1))
	}

	if object.ChecksumSHA256 != nil {
		span.SetAttributes(attribute.String("storage.object.checksum_sha256", *object.ChecksumSHA256))
	}
}

// SetObjectInfoSpanAttributes sets span attributes from the object info.
func SetObjectInfoSpanAttributes(span trace.Span, object *StorageObject) {
	SetObjectChecksumSpanAttributes(span, &object.StorageObjectChecksum)

	if object.Size != nil {
		span.SetAttributes(attribute.Int64("storage.object.size", *object.Size))
	}

	if object.ETag != nil {
		span.SetAttributes(attribute.String("storage.object.etag", *object.ETag))
	}

	if object.StorageClass != nil {
		span.SetAttributes(attribute.String("storage.object.storage_class", *object.StorageClass))
	}

	if object.VersionID != nil {
		span.SetAttributes(attribute.String("storage.object.version", *object.VersionID))
	}

	if object.TagCount > 0 {
		span.SetAttributes(attribute.Int("storage.object.tags_count", object.TagCount))
	}

	if len(object.Metadata) > 0 {
		span.SetAttributes(attribute.Int("storage.object.metadata_count", len(object.Metadata)))
	}

	if object.Expires != nil && !object.Expires.IsZero() {
		span.SetAttributes(attribute.String("storage.object.expires", object.Expires.Format(time.RFC3339)))
	}

	if object.Expiration != nil && !object.Expiration.IsZero() {
		span.SetAttributes(attribute.String("storage.object.expiration", object.Expiration.Format(time.RFC3339)))
	}

	if object.ExpirationRuleID != nil {
		span.SetAttributes(attribute.String("storage.object.expiration_rule_id", *object.ExpirationRuleID))
	}
}

// SetUploadInfoAttributes sets span attributes from the upload info.
func SetUploadInfoAttributes(span trace.Span, object *StorageUploadInfo) {
	SetObjectChecksumSpanAttributes(span, &object.StorageObjectChecksum)

	if object.ETag != nil {
		span.SetAttributes(attribute.String("storage.object.etag", *object.ETag))
	}

	if object.Size != nil {
		span.SetAttributes(attribute.Int64("storage.object.size", *object.Size))
	}

	if object.VersionID != nil {
		span.SetAttributes(attribute.String("storage.object.version", *object.VersionID))
	}

	if object.Expiration != nil && !object.Expiration.IsZero() {
		span.SetAttributes(attribute.String("storage.object.expiration", object.Expiration.Format(time.RFC3339)))
	}

	if object.ExpirationRuleID != nil {
		span.SetAttributes(attribute.String("storage.object.expiration_rule_id", *object.ExpirationRuleID))
	}
}
