package common

import (
	"context"
	"io"
	"time"

	"github.com/hasura/ndc-sdk-go/scalar"
)

// StorageClient abstracts required methods of the storage client.
type StorageClient interface { //nolint:interfacebloat
	// MakeBucket creates a new bucket.
	MakeBucket(ctx context.Context, options *MakeStorageBucketOptions) error
	// ListBuckets list all buckets.
	ListBuckets(ctx context.Context, options *ListStorageBucketsOptions, predicate func(string) bool) (*StorageBucketListResults, error)
	// GetBucket gets a bucket by name.
	GetBucket(ctx context.Context, name string, options BucketOptions) (*StorageBucket, error)
	// BucketExists checks if a bucket exists.
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	// RemoveBucket removes a bucket, bucket should be empty to be successfully removed.
	RemoveBucket(ctx context.Context, bucketName string) error
	// UpdateBucket updates configurations for the bucket.
	UpdateBucket(ctx context.Context, bucketName string, opts UpdateStorageBucketOptions) error
	// ListObjects list objects in a bucket.
	ListObjects(ctx context.Context, bucketName string, opts *ListStorageObjectsOptions, predicate func(string) bool) (*StorageObjectListResults, error)
	// ListIncompleteUploads list partially uploaded objects in a bucket.
	ListIncompleteUploads(ctx context.Context, bucketName string, args ListIncompleteUploadsOptions) ([]StorageObjectMultipartInfo, error)
	// ListDeletedObjects list deleted objects in a bucket.
	ListDeletedObjects(ctx context.Context, bucketName string, opts *ListStorageObjectsOptions, predicate func(string) bool) (*StorageObjectListResults, error)
	// GetObject returns a stream of the object data. Most of the common errors occur when reading the stream.
	GetObject(ctx context.Context, bucketName string, objectName string, opts GetStorageObjectOptions) (io.ReadCloser, error)
	// PutObject uploads objects that are less than 128MiB in a single PUT operation. For objects that are greater than 128MiB in size,
	// PutObject seamlessly uploads the object as parts of 128MiB or more depending on the actual file size. The max upload size for an object is 5TB.
	PutObject(ctx context.Context, bucketName string, objectName string, opts *PutStorageObjectOptions, reader io.Reader, objectSize int64) (*StorageUploadInfo, error)
	// CopyObject creates or replaces an object through server-side copying of an existing object.
	// It supports conditional copying, copying a part of an object and server-side encryption of destination and decryption of source.
	// To copy multiple source objects into a single destination object see the ComposeObject API.
	CopyObject(ctx context.Context, dest StorageCopyDestOptions, src StorageCopySrcOptions) (*StorageUploadInfo, error)
	// ComposeObject creates an object by concatenating a list of source objects using server-side copying.
	ComposeObject(ctx context.Context, dest StorageCopyDestOptions, srcs []StorageCopySrcOptions) (*StorageUploadInfo, error)
	// StatObject fetches metadata of an object.
	StatObject(ctx context.Context, bucketName string, objectName string, opts GetStorageObjectOptions) (*StorageObject, error)
	// RemoveObject removes an object with some specified options
	RemoveObject(ctx context.Context, bucketName string, objectName string, opts RemoveStorageObjectOptions) error
	// RemoveObjects remove a list of objects obtained from an input channel. The call sends a delete request to the server up to 1000 objects at a time.
	// The errors observed are sent over the error channel.
	RemoveObjects(ctx context.Context, bucketName string, opts *RemoveStorageObjectsOptions, predicate func(string) bool) []RemoveStorageObjectError
	// UpdateObject updates object configurations.
	UpdateObject(ctx context.Context, bucketName string, objectName string, opts UpdateStorageObjectOptions) error
	// RestoreObject restores a soft-deleted object.
	RestoreObject(ctx context.Context, bucketName string, objectName string) error
	// RemoveIncompleteUpload removes a partially uploaded object.
	RemoveIncompleteUpload(ctx context.Context, bucketName string, objectName string) error
	// PresignedGetObject generates a presigned URL for HTTP GET operations. Browsers/Mobile clients may point to this URL to directly download objects even if the bucket is private.
	// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
	// The maximum expiry is 604800 seconds (i.e. 7 days) and minimum is 1 second.
	PresignedGetObject(ctx context.Context, bucketName string, objectName string, opts PresignedGetStorageObjectOptions) (string, error)
	// PresignedPutObject generates a presigned URL for HTTP PUT operations.
	// Browsers/Mobile clients may point to this URL to upload objects directly to a bucket even if it is private.
	// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
	// The default expiry is set to 7 days.
	PresignedPutObject(ctx context.Context, bucketName string, objectName string, expiry time.Duration) (string, error)
}

// ListStorageBucketsOptions holds all options of a list bucket request.
type ListStorageBucketsOptions struct {
	// Only list objects with the prefix
	Prefix string
	// The maximum number of objects requested per
	// batch, advanced use-case not useful for most
	// applications
	MaxResults *int
	// StartAfter start listing lexically at this object onwards.
	StartAfter string
	// Options to be included for the object information.
	Include    BucketIncludeOptions
	NumThreads int
}

// BucketIncludeOptions contain include options for getting bucket information.
type BucketIncludeOptions struct {
	Tags       bool
	Versioning bool
	Lifecycle  bool
	Encryption bool
	ObjectLock bool
}

// BucketOptions hold options to get bucket information.
type BucketOptions struct {
	Prefix     string `json:"prefix"`
	NumThreads int
	Include    BucketIncludeOptions
}

// StorageBucketListResults hold the paginated results of the storage bucket list.
type StorageBucketListResults struct {
	Buckets  []StorageBucket       `json:"buckets"`
	PageInfo StoragePaginationInfo `json:"pageInfo"`
}

// StorageBucket the container for bucket metadata.
type StorageBucket struct {
	// Client ID
	ClientID string `json:"client_id"`
	// The name of the bucket.
	Name string `json:"name"`
	// Bucket tags or metadata.
	Tags []StorageKeyValue `json:"tags,omitempty"`
	// The versioning configuration
	Versioning *StorageBucketVersioningConfiguration `json:"versioning"`
	// The versioning configuration
	Lifecycle *ObjectLifecycleConfiguration `json:"lifecycle"`
	// The server-side encryption configuration.
	Encryption *ServerSideEncryptionConfiguration `json:"encryption"`

	// Retention policy enforces a minimum retention time for all objects
	// contained in the bucket. A RetentionPolicy of nil implies the bucket
	// has no minimum data retention.
	ObjectLock *StorageObjectLockConfig `json:"object_lock"`

	// The location of the bucket.
	Region *string `json:"region"`

	// The bucket's custom placement configuration that holds a list of
	// regional locations for custom dual regions.
	CustomPlacementConfig *CustomPlacementConfig `json:"custom_placement_config"`

	// DefaultEventBasedHold is the default value for event-based hold on newly created objects in this bucket. It defaults to false.
	DefaultEventBasedHold *bool `json:"default_event_based_hold"`

	// StorageClass is the default storage class of the bucket. This defines
	// how objects in the bucket are stored and determines the SLA and the cost of storage.
	StorageClass *string `json:"storage_class"`

	// Date time the bucket was created.
	CreationTime *time.Time `json:"creation_time"`
	// Date time the bucket was created.
	LastModified *time.Time `json:"last_modified"`

	// RequesterPays reports whether the bucket is a Requester Pays bucket.
	// Clients performing operations on Requester Pays buckets must provide
	// a user project (see BucketHandle.UserProject), which will be billed
	// for the operations.
	RequesterPays *bool `json:"requester_pays"`

	// The bucket's Cross-Origin Resource Sharing (CORS) configuration.
	CORS []BucketCors `json:"cors,omitempty"`

	// The logging configuration.
	Logging *BucketLogging `json:"logging"`

	// The website configuration.
	Website *BucketWebsite `json:"website,omitempty"`

	// Etag is the HTTP/1.1 Entity tag for the bucket.
	// This field is read-only.
	Etag *string `json:"etag"`

	// LocationType describes how data is stored and replicated.
	// Typical values are "multi-region", "region" and "dual-region".
	LocationType *string `json:"location_type"`

	// RPO configures the Recovery Point Objective (RPO) policy of the bucket.
	// Set to RPOAsyncTurbo to turn on Turbo Replication for a bucket.
	// See https://cloud.google.com/storage/docs/managing-turbo-replication for
	// more information.
	RPO *GoogleStorageRPO `json:"rpo"`

	// Autoclass holds the bucket's autoclass configuration. If enabled,
	// allows for the automatic selection of the best storage class
	// based on object access patterns.
	Autoclass *BucketAutoclass `json:"autoclass"`

	// SoftDeletePolicy contains the bucket's soft delete policy, which defines
	// the period of time that soft-deleted objects will be retained, and cannot
	// be permanently deleted.
	SoftDeletePolicy *StorageObjectSoftDeletePolicy `json:"soft_delete_policy"`

	// HierarchicalNamespace contains the bucket's hierarchical namespace
	// configuration. Hierarchical namespace enabled buckets can contain
	// [cloud.google.com/go/storage/control/apiv2/controlpb.Folder] resources.
	// It cannot be modified after bucket creation time.
	// UniformBucketLevelAccess must also be enabled on the bucket.
	HierarchicalNamespace *BucketHierarchicalNamespace `json:"hierarchical_namespace"`
}

// HierarchicalNamespace contains the bucket's hierarchical namespace
// configuration. Hierarchical namespace enabled buckets can contain
// [cloud.google.com/go/storage/control/apiv2/controlpb.Folder] resources.
type BucketHierarchicalNamespace struct {
	// Enabled indicates whether hierarchical namespace features are enabled on
	// the bucket. This can only be set at bucket creation time currently.
	Enabled bool `json:"enabled"`
}

// BucketLogging holds the bucket's logging configuration, which defines the
// destination bucket and optional name prefix for the current bucket's logs.
type BucketLogging struct {
	// The destination bucket where the current bucket's logs
	// should be placed.
	LogBucket string `json:"log_bucket"`

	// A prefix for log object names.
	LogObjectPrefix string `json:"log_object_prefix"`
}

// GoogleStorageRPO (Recovery Point Objective) configures the turbo replication feature. See
// https://cloud.google.com/storage/docs/managing-turbo-replication for more information.
// @enum DEFAULT,ASYNC_TURBO
type GoogleStorageRPO string

// BucketCors is the bucket's Cross-Origin Resource Sharing (CORS) configuration.
type BucketCors struct {
	// MaxAge is the value to return in the Access-Control-Max-Age
	// header used in preflight responses.
	MaxAge scalar.DurationString `json:"max_age"`

	// Methods is the list of HTTP methods on which to include CORS response
	// headers, (GET, OPTIONS, POST, etc) Note: "*" is permitted in the list
	// of methods, and means "any method".
	Methods []string `json:"methods"`

	// Origins is the list of Origins eligible to receive CORS response
	// headers. Note: "*" is permitted in the list of origins, and means
	// "any Origin".
	Origins []string `json:"origins"`

	// ResponseHeaders is the list of HTTP headers other than the simple
	// response headers to give permission for the user-agent to share
	// across domains.
	ResponseHeaders []string `json:"response_headers"`
}

// BucketWebsite holds the bucket's website configuration, controlling how the
// service behaves when accessing bucket contents as a web site. See
// https://cloud.google.com/storage/docs/static-website for more information.
type BucketWebsite struct {
	// If the requested object path is missing, the service will ensure the path has
	// a trailing '/', append this suffix, and attempt to retrieve the resulting
	// object. This allows the creation of index.html objects to represent directory
	// pages.
	MainPageSuffix string `json:"main_page_suffix"`

	// If the requested object path is missing, and any mainPageSuffix object is
	// missing, if applicable, the service will return the named object from this
	// bucket as the content for a 404 Not Found result.
	NotFoundPage *string `json:"not_found_page,omitempty"`
}

// CustomPlacementConfig holds the bucket's custom placement
// configuration for Custom Dual Regions. See
// https://cloud.google.com/storage/docs/locations#location-dr for more information.
type CustomPlacementConfig struct {
	// The list of regional locations in which data is placed.
	// Custom Dual Regions require exactly 2 regional locations.
	DataLocations []string `json:"data_locations"`
}

// Autoclass holds the bucket's autoclass configuration. If enabled,
// allows for the automatic selection of the best storage class
// based on object access patterns. See
// https://cloud.google.com/storage/docs/using-autoclass for more information.
type BucketAutoclass struct {
	// Enabled specifies whether the autoclass feature is enabled
	// on the bucket.
	Enabled bool `json:"enabled"`
	// ToggleTime is the time from which Autoclass was last toggled.
	// If Autoclass is enabled when the bucket is created, the ToggleTime
	// is set to the bucket creation time. This field is read-only.
	ToggleTime time.Time `json:"toggle_time"`
	// TerminalStorageClass: The storage class that objects in the bucket
	// eventually transition to if they are not read for a certain length of
	// time. Valid values are NEARLINE and ARCHIVE.
	// To modify TerminalStorageClass, Enabled must be set to true.
	TerminalStorageClass string `json:"terminal_storage_class"`
	// TerminalStorageClassUpdateTime represents the time of the most recent
	// update to "TerminalStorageClass".
	TerminalStorageClassUpdateTime time.Time `json:"terminal_storage_class_update_time"`
}

// StorageObjectSoftDeletePolicy contains the bucket's soft delete policy, which defines the
// period of time that soft-deleted objects will be retained, and cannot be
// permanently deleted.
type StorageObjectSoftDeletePolicy struct {
	// EffectiveTime indicates the time from which the policy, or one with a
	// greater retention, was effective. This field is read-only.
	EffectiveTime time.Time `json:"effective_time"`

	// RetentionDuration is the amount of time that soft-deleted objects in the
	// bucket will be retained and cannot be permanently deleted.
	RetentionDuration scalar.DurationString `json:"retention_duration"`
}

// StorageOwner name.
type StorageOwner struct {
	DisplayName *string `json:"name"`
	ID          *string `json:"id"`
}

// StorageGrantee represents the person being granted permissions.
type StorageGrantee struct {
	ID          *string `json:"id"`
	DisplayName *string `json:"display_name"`
	URI         *string `json:"uri"`
}

// StorageGrant holds grant information.
type StorageGrant struct {
	Grantee    *StorageGrantee `json:"grantee"`
	Permission *string         `json:"permission"`
}

// StorageRestoreInfo contains information of the restore operation of an archived object.
type StorageRestoreInfo struct {
	// Is the restoring operation is still ongoing
	OngoingRestore bool `json:"ongoing_restore"`
	// When the restored copy of the archived object will be removed
	ExpiryTime *time.Time `json:"expiry_time"`
}

// StorageObject container for object metadata.
type StorageObject struct {
	// An ETag is optionally set to md5sum of an object.  In case of multipart objects,
	// ETag is of the form MD5SUM-N where MD5SUM is md5sum of all individual md5sums of
	// each parts concatenated into one string.
	ETag *string `json:"etag"`

	ClientID           string     `json:"client_id"`     // Client ID
	Bucket             string     `json:"bucket"`        // Name of the bucket
	Name               string     `json:"name"`          // Name of the object
	LastModified       time.Time  `json:"last_modified"` // Date and time the object was last modified.
	Size               *int64     `json:"size"`          // Size in bytes of the object.
	ContentType        *string    `json:"content_type"`  // A standard MIME type describing the format of the object data.
	ContentEncoding    *string    `json:"content_encoding,omitempty"`
	ContentDisposition *string    `json:"content_disposition,omitempty"`
	ContentLanguage    *string    `json:"content_language,omitempty"`
	CacheControl       *string    `json:"cache_control,omitempty"`
	Expires            *time.Time `json:"expires"` // The date and time at which the object is no longer able to be cached.
	IsDirectory        bool       `json:"is_directory"`

	// Collection of additional metadata on the object.
	// In MinIO and S3, x-amz-meta-* headers stripped "x-amz-meta-" prefix containing the first value.
	Metadata []StorageKeyValue `json:"metadata,omitempty"`

	// Raw metadata headers, eg: x-amz-meta-*, content-encoding etc... Only returned by MinIO servers.
	RawMetadata []StorageKeyValue `json:"raw_metadata,omitempty"`

	// User tags
	Tags []StorageKeyValue `json:"tags,omitempty"`

	// The total count value of tags
	TagCount int `json:"tag_count,omitempty"`

	// Owner name.
	Owner *StorageOwner `json:"owner"`

	// ACL grant.
	Grant []StorageGrant `json:"grant,omitempty"`

	// The class of storage used to store the object or the access tier on Azure blob storage.
	StorageClass *string `json:"storage_class,omitempty"`

	// Versioning related information
	IsLatest  *bool   `json:"is_latest"`
	Deleted   *bool   `json:"deleted"`
	VersionID *string `json:"version_id,omitempty"`

	// x-amz-replication-status value is either in one of the following states
	ReplicationStatus *StorageObjectReplicationStatus `json:"replication_status"`
	// set to true if delete marker has backing object version on target, and eligible to replicate
	ReplicationReady *bool `json:"replication_ready"`
	// Lifecycle expiry-date and ruleID associated with the expiry
	// not to be confused with `Expires` HTTP header.
	Expiration       *time.Time `json:"expiration"`
	ExpirationRuleID *string    `json:"expiration_rule_id"`

	Restore *StorageRestoreInfo `json:"restore"`

	// Checksum values
	StorageObjectChecksum

	// Azure Blob Store properties
	ACL                       any                    `json:"acl,omitempty"`
	AccessTierChangeTime      *time.Time             `json:"access_tier_change_time"`
	AccessTierInferred        *bool                  `json:"access_tier_inferred"`
	ArchiveStatus             *string                `json:"archive_status"`
	BlobSequenceNumber        *int64                 `json:"blob_sequence_number"`
	BlobType                  *string                `json:"blob_type"`
	ContentMD5                *string                `json:"content_md5"`
	Copy                      *StorageObjectCopyInfo `json:"copy"`
	CreationTime              *time.Time             `json:"creation_time"`
	DeletedTime               *time.Time             `json:"deleted_time"`
	CustomerProvidedKeySHA256 *string                `json:"customer_provided_key_sha256"`
	DestinationSnapshot       *string                `json:"destination_snapshot"`
	MediaLink                 *string                `json:"media_link"`
	// The name of the encryption scope under which the blob is encrypted.
	KMSKeyName         *string    `json:"kms_key_name"`
	ServerEncrypted    *bool      `json:"server_encrypted"`
	Group              *string    `json:"group"`
	RetentionUntilDate *time.Time `json:"retention_until_date"`
	RetentionMode      *string    `json:"retention_mode"`
	IncrementalCopy    *bool      `json:"incremental_copy"`
	IsSealed           *bool      `json:"sealed"`
	LastAccessTime     *time.Time `json:"last_access_time"`
	LeaseDuration      *string    `json:"lease_duration"`
	LeaseState         *string    `json:"lease_state"`
	LeaseStatus        *string    `json:"lease_status"`
	LegalHold          *bool      `json:"legal_hold"`
	Permissions        *string    `json:"permissions"`

	// If an object is in rehydrate pending state then this header is returned with priority of rehydrate. Valid values are High
	// and Standard.
	RehydratePriority      *string `json:"rehydrate_priority"`
	RemainingRetentionDays *int32  `json:"remaining_retention_days"`
	ResourceType           *string `json:"resource_type"`
}

// StorageObjectCopyInfo holds the copy information if the object was copied from another object.
type StorageObjectCopyInfo struct {
	ID                string     `json:"id"`
	CompletionTime    *time.Time `json:"completion_time"`
	Progress          *string    `json:"progress"`
	Source            *string    `json:"source"`
	Status            *string    `json:"status"`
	StatusDescription *string    `json:"status_description"`
}

// StoragePaginationInfo holds the pagination information.
type StoragePaginationInfo struct {
	HasNextPage bool `json:"hasNextPage"`
}

// StorageObjectListResults hold the paginated results of the storage object list.
type StorageObjectListResults struct {
	Objects  []StorageObject       `json:"objects"`
	PageInfo StoragePaginationInfo `json:"pageInfo"`
}

// StorageObjectReplicationStatus represents the x-amz-replication-status value enum.
// @enum COMPLETED,PENDING,FAILED,REPLICA
type StorageObjectReplicationStatus string

// StorageObjectChecksum represents checksum values of the object.
type StorageObjectChecksum struct {
	ChecksumCRC32     *string `json:"checksum_crc32,omitempty"`
	ChecksumCRC32C    *string `json:"checksum_crc32c,omitempty"`
	ChecksumSHA1      *string `json:"checksum_sha1,omitempty"`
	ChecksumSHA256    *string `json:"checksum_sha256,omitempty"`
	ChecksumCRC64NVME *string `json:"checksum_crc64_nvme,omitempty"`
}

// StorageUploadInfo represents the information of the uploaded object.
type StorageUploadInfo struct {
	// An ETag is optionally set to md5sum of an object.  In case of multipart objects,
	// ETag is of the form MD5SUM-N where MD5SUM is md5sum of all individual md5sums of
	// each parts concatenated into one string.
	ETag *string `json:"etag"`

	ClientID     string     `json:"client_id"`     // Client ID
	Bucket       string     `json:"bucket"`        // Name of the bucket
	Name         string     `json:"name"`          // Name of the object
	LastModified *time.Time `json:"last_modified"` // Date and time the object was last modified.
	Size         *int64     `json:"size"`          // Size in bytes of the object.
	Location     *string    `json:"location"`
	VersionID    *string    `json:"version_id"`
	ContentMD5   *string    `json:"content_md5"`

	// Lifecycle expiry-date and ruleID associated with the expiry
	// not to be confused with `Expires` HTTP header.
	Expiration       *time.Time `json:"expiration"`
	ExpirationRuleID *string    `json:"expiration_rule_id"`

	// Checksum values
	StorageObjectChecksum
}

// StorageObjectMultipartInfo container for multipart object metadata.
type StorageObjectMultipartInfo struct {
	// Date and time at which the multipart upload was initiated.
	Initiated *time.Time `json:"initiated"`

	// The type of storage to use for the object. Defaults to 'STANDARD'.
	StorageClass *string `json:"storage_class"`

	// Key of the object for which the multipart upload was initiated.
	Name *string `json:"name"`

	// Size in bytes of the object.
	Size *int64 `json:"size"`

	// Upload ID that identifies the multipart upload.
	UploadID *string `json:"upload_id"`
}

// EncryptionMethod represents a server-side-encryption method enum.
// @enum SSE_C,KMS,S3
// type ServerSideEncryptionMethod string

// StorageRetentionMode the object retention mode.
// @enum Locked,Unlocked,Mutable,Delete
type StorageRetentionMode string

// RemoveStorageObjectError the container of Multi Delete S3 API error.
type RemoveStorageObjectError struct {
	ObjectName string `json:"object_name"`
	VersionID  string `json:"version_id"`
	Error      string `json:"error"`
}

// ChecksumType contains information about the checksum type.
// @enum SHA256,SHA1,CRC32,CRC32C,CRC64NVME,FullObjectCRC32,FullObjectCRC32C,None
type ChecksumType string

// NotificationCommonConfig - represents one single notification configuration
// such as topic, queue or lambda configuration.
type NotificationCommonConfig struct {
	ID     *string             `json:"id,omitempty"`
	Arn    *string             `json:"arn"`
	Events []string            `json:"event"`
	Filter *NotificationFilter `json:"filter,omitempty"`
}

// NotificationTopicConfig carries one single topic notification configuration
type NotificationTopicConfig struct {
	NotificationCommonConfig
	Topic string `json:"topic"`
}

// NotificationQueueConfig carries one single queue notification configuration
type NotificationQueueConfig struct {
	NotificationCommonConfig
	Queue string `json:"queue"`
}

// NotificationLambdaConfig carries one single cloudfunction notification configuration
type NotificationLambdaConfig struct {
	NotificationCommonConfig
	Lambda string `json:"cloud_function"`
}

// NotificationConfig the struct that represents a notification configuration object.
type NotificationConfig struct {
	LambdaConfigs []NotificationLambdaConfig `json:"cloud_function_configurations"`
	TopicConfigs  []NotificationTopicConfig  `json:"topic_configurations"`
	QueueConfigs  []NotificationQueueConfig  `json:"queue_configurations"`
}

// NotificationFilter - a tag in the notification xml structure which carries suffix/prefix filters
type NotificationFilter struct {
	S3Key *NotificationS3Key `json:"s3_key,omitempty"`
}

// NotificationFilterRule child of S3Key, a tag in the notification xml which
// carries suffix/prefix filters
type NotificationFilterRule struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// NotificationS3Key child of Filter, a tag in the notification xml which
// carries suffix/prefix filters
type NotificationS3Key struct {
	FilterRules []NotificationFilterRule `json:"filter_rule,omitempty"`
}

// ObjectLifecycleRule represents a single rule in lifecycle configuration
type ObjectLifecycleRule struct {
	ID                             string                                      `json:"id,omitempty"`
	Enabled                        bool                                        `json:"enabled,omitempty"`
	AbortIncompleteMultipartUpload *ObjectAbortIncompleteMultipartUpload       `json:"abort_incomplete_multipart_upload"`
	Expiration                     *ObjectLifecycleExpiration                  `json:"expiration,omitempty"`
	DelMarkerExpiration            *ObjectLifecycleDelMarkerExpiration         `json:"del_marker_expiration,omitempty"`
	AllVersionsExpiration          *ObjectLifecycleAllVersionsExpiration       `json:"all_versions_expiration,omitempty"`
	RuleFilter                     []ObjectLifecycleFilter                     `json:"filter,omitempty"`
	NoncurrentVersionExpiration    *ObjectLifecycleNoncurrentVersionExpiration `json:"noncurrent_version_expiration,omitempty"`
	NoncurrentVersionTransition    *ObjectLifecycleNoncurrentVersionTransition `json:"noncurrent_version_transition,omitempty"`
	Prefix                         *string                                     `json:"prefix,omitempty"`
	Transition                     *ObjectLifecycleTransition                  `json:"transition,omitempty"`
}

// ObjectLifecycleConfiguration is a collection of lifecycle Rule objects.
type ObjectLifecycleConfiguration struct {
	Rules []ObjectLifecycleRule `json:"rules"`
}

// AbortIncompleteMultipartUpload structure, not supported yet on MinIO
type ObjectAbortIncompleteMultipartUpload struct {
	DaysAfterInitiation *int `json:"days_after_initiation"`
}

// ObjectLifecycleExpiration expiration details of lifecycle configuration
type ObjectLifecycleExpiration struct {
	Date         *scalar.Date `json:"date,omitempty"`
	Days         *int         `json:"days,omitempty"`
	DeleteMarker *bool        `json:"expired_object_delete_marker,omitempty"`
	DeleteAll    *bool        `json:"expired_object_all_versions,omitempty"`
}

// IsEmpty checks if all properties of the object are empty.
func (fe ObjectLifecycleExpiration) IsEmpty() bool {
	return fe.DeleteAll == nil && fe.Date == nil && fe.Days == nil && fe.DeleteMarker == nil
}

// ObjectLifecycleTransition transition details of lifecycle configuration
type ObjectLifecycleTransition struct {
	Date         *scalar.Date `json:"date"`
	StorageClass *string      `json:"storage_class"`
	Days         *int         `json:"days"`
}

// LifecycleDelMarkerExpiration represents DelMarkerExpiration actions element in an ILM policy
type ObjectLifecycleDelMarkerExpiration struct {
	Days *int `json:"days"`
}

// ObjectLifecycleAllVersionsExpiration represents AllVersionsExpiration actions element in an ILM policy
type ObjectLifecycleAllVersionsExpiration struct {
	Days         *int  `json:"days"`
	DeleteMarker *bool `json:"delete_marker"`
}

// ObjectLifecycleFilter will be used in selecting rule(s) for lifecycle configuration
type ObjectLifecycleFilter struct {
	// MatchesPrefix is the condition matching an object if any of the
	// matches_prefix strings are an exact prefix of the object's name.
	MatchesPrefix []string `json:"matches_prefix,omitempty"`

	// MatchesStorageClasses is the condition matching the object's storage
	// class.
	//
	// Values include "STANDARD", "NEARLINE", "COLDLINE" and "ARCHIVE".
	MatchesStorageClasses []string `json:"matches_storage_classes,omitempty"`

	// MatchesSuffix is the condition matching an object if any of the
	// matches_suffix strings are an exact suffix of the object's name.
	MatchesSuffix []string `json:"matches_suffix,omitempty"`

	// Tags structure key/value pair representing an object tag to apply configuration
	Tags                  []StorageKeyValue `json:"tags,omitempty"`
	ObjectSizeLessThan    *int64            `json:"object_size_less_than,omitempty"`
	ObjectSizeGreaterThan *int64            `json:"object_size_greater_than,omitempty"`
}

// ObjectLifecycleNoncurrentVersionExpiration - Specifies when noncurrent object versions expire.
// Upon expiration, server permanently deletes the noncurrent object versions.
// Set this lifecycle configuration action on a bucket that has versioning enabled
// (or suspended) to request server delete noncurrent object versions at a
// specific period in the object's lifetime.
type ObjectLifecycleNoncurrentVersionExpiration struct {
	NoncurrentDays          *int `json:"noncurrent_days,omitempty"`
	NewerNoncurrentVersions *int `json:"newer_noncurrent_versions,omitempty"`
}

// ObjectLifecycleNoncurrentVersionTransition sets this action to request server to
// transition noncurrent object versions to different set storage classes
// at a specific period in the object's lifetime.
type ObjectLifecycleNoncurrentVersionTransition struct {
	StorageClass            *string `json:"storage_class,omitempty"`
	NoncurrentDays          *int    `json:"noncurrent_days"`
	NewerNoncurrentVersions *int    `json:"newer_noncurrent_versions,omitempty"`
}

// ServerSideEncryptionConfiguration is the default encryption configuration structure.
type ServerSideEncryptionConfiguration struct {
	KmsMasterKeyID string `json:"kms_master_key_id,omitempty"`
	SSEAlgorithm   string `json:"sse_algorithm,omitempty"`
}

// IsEmpty checks if the configuration is empty.
func (ssec ServerSideEncryptionConfiguration) IsEmpty() bool {
	return ssec.KmsMasterKeyID == "" && ssec.SSEAlgorithm == ""
}

// SetStorageObjectLockConfig represents the object lock configuration options in given bucket
type SetStorageObjectLockConfig struct {
	Mode     *StorageRetentionMode         `json:"mode"`
	Validity *uint                         `json:"validity"`
	Unit     *StorageRetentionValidityUnit `json:"unit"`
}

// StorageObjectLockConfig represents the object lock configuration in given bucket
type StorageObjectLockConfig struct {
	SetStorageObjectLockConfig

	Enabled bool `json:"enabled"`
}

// StorageRetentionValidityUnit retention validity unit.
// @enum DAYS,YEARS
type StorageRetentionValidityUnit string

// StorageBucketVersioningConfiguration is the versioning configuration structure
type StorageBucketVersioningConfiguration struct {
	Enabled   bool    `json:"enabled"`
	MFADelete *string `json:"mfa_delete"`
	// MinIO extension - allows selective, prefix-level versioning exclusion.
	// Requires versioning to be enabled
	ExcludedPrefixes []string `json:"excluded_prefixes,omitempty"`
	ExcludeFolders   *bool    `json:"exclude_folders"`
}

// StorageReplicationConfig replication configuration specified in
// https://docs.aws.amazon.com/AmazonS3/latest/dev/replication-add-config.html
type StorageReplicationConfig struct {
	Rules []StorageReplicationRule `json:"rules"`
	Role  *string                  `json:"role"`
}

// StorageReplicationRule a rule for replication configuration.
type StorageReplicationRule struct {
	ID                        *string                        `json:"id,omitempty"`
	Status                    StorageReplicationRuleStatus   `json:"status"`
	Priority                  int                            `json:"priority"`
	DeleteMarkerReplication   *DeleteMarkerReplication       `json:"delete_marker_replication"`
	DeleteReplication         *DeleteReplication             `json:"delete_replication"`
	Destination               *StorageReplicationDestination `json:"destination"`
	Filter                    StorageReplicationFilter       `json:"filter"`
	SourceSelectionCriteria   *SourceSelectionCriteria       `json:"source_selection_criteria"`
	ExistingObjectReplication *ExistingObjectReplication     `json:"existing_object_replication,omitempty"`
}

// Destination the destination in ReplicationConfiguration.
type StorageReplicationDestination struct {
	Bucket       string  `json:"bucket"`
	StorageClass *string `json:"storage_class,omitempty"`
}

// ExistingObjectReplication whether existing object replication is enabled
type ExistingObjectReplication struct {
	Status StorageReplicationRuleStatus `json:"status"` // should be set to "Disabled" by default
}

// StorageReplicationRuleStatus represents Enabled/Disabled status
// @enum Enabled,Disabled
type StorageReplicationRuleStatus string

// DeleteMarkerReplication whether delete markers are replicated -
// https://docs.aws.amazon.com/AmazonS3/latest/dev/replication-add-config.html
type DeleteMarkerReplication struct {
	Status StorageReplicationRuleStatus `json:"status"` // should be set to "Disabled" by default
}

// DeleteReplication whether versioned deletes are replicated. This is a MinIO specific extension
type DeleteReplication struct {
	Status StorageReplicationRuleStatus `json:"status"` // should be set to "Disabled" by default
}

// ReplicaModifications specifies if replica modification sync is enabled
type ReplicaModifications struct {
	Status StorageReplicationRuleStatus `json:"status"` // should be set to "Enabled" by default
}

// SourceSelectionCriteria specifies additional source selection criteria in ReplicationConfiguration.
type SourceSelectionCriteria struct {
	ReplicaModifications *ReplicaModifications `json:"replica_modifications"`
}

// StorageReplicationFilter a filter for a replication configuration Rule.
type StorageReplicationFilter struct {
	Prefix *string                      `json:"prefix,omitempty"`
	And    *StorageReplicationFilterAnd `json:"and,omitempty"`
	Tag    []StorageKeyValue            `json:"tag,omitempty"`
}

// StorageReplicationFilterAnd - a tag to combine a prefix and multiple tags for replication configuration rule.
type StorageReplicationFilterAnd struct {
	Prefix *string           `json:"prefix,omitempty"`
	Tags   []StorageKeyValue `json:"tag,omitempty"`
}
