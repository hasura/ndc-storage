package common

import (
	"time"

	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
)

// ListStorageBucketArguments represent the input arguments for the ListBuckets methods.
type ListStorageBucketArguments struct {
	// The storage client ID.
	ClientID *StorageClientID `json:"clientId,omitempty"`
}

// StorageBucketArguments represent the common input arguments for bucket-related methods.
type StorageBucketArguments struct {
	// The storage client ID.
	ClientID *StorageClientID `json:"clientId,omitempty"`
	// The bucket name.
	Bucket string `json:"bucket,omitempty"`
}

// CopyStorageObjectArguments represent input arguments of the CopyObject method.
type CopyStorageObjectArguments struct {
	// The storage client ID
	ClientID *StorageClientID       `json:"clientId,omitempty"`
	Dest     StorageCopyDestOptions `json:"dest"`
	Source   StorageCopySrcOptions  `json:"source"`
}

// ComposeStorageObjectArguments represent input arguments of the ComposeObject method.
type ComposeStorageObjectArguments struct {
	// The storage client ID
	ClientID *StorageClientID        `json:"clientId,omitempty"`
	Dest     StorageCopyDestOptions  `json:"dest"`
	Sources  []StorageCopySrcOptions `json:"sources"`
}

// MakeStorageBucketArguments holds all arguments to tweak bucket creation.
type MakeStorageBucketArguments struct {
	ClientID *StorageClientID `json:"clientId,omitempty"`

	MakeStorageBucketOptions
}

// MakeStorageBucketOptions holds all options to tweak bucket creation.
type MakeStorageBucketOptions struct {
	// Bucket name
	Name string `json:"name"`
	// Bucket location
	Region string `json:"region,omitempty"`
	// Enable object locking
	ObjectLocking bool `json:"objectLocking,omitempty"`
	// Optional tags
	Tags map[string]string `json:"tags,omitempty"`
}

// SetBucketTaggingArguments represent the input arguments for the SetBucketTagging method.
type SetStorageBucketTaggingArguments struct {
	StorageBucketArguments

	Tags map[string]string `json:"tags"`
}

// ListIncompleteUploadsArguments the input arguments of the ListIncompleteUploads method.
type ListIncompleteUploadsArguments struct {
	StorageBucketArguments

	Prefix    string `json:"prefix"`
	Recursive bool   `json:"recursive,omitempty"`
}

// RemoveIncompleteUploadArguments represent the input arguments for the RemoveIncompleteUpload method.
type RemoveIncompleteUploadArguments struct {
	StorageBucketArguments

	Object string `json:"object"`
}

// PresignedGetStorageObjectArguments represent the input arguments for the PresignedGetObject method.
type PresignedGetStorageObjectArguments struct {
	StorageBucketArguments
	PresignedGetStorageObjectOptions

	Object string            `json:"object"`
	Where  schema.Expression `json:"where"  ndc:"predicate=StorageObjectSimple"`
}

// PresignedGetStorageObjectOptions represent the options for the PresignedGetObject method.
type PresignedGetStorageObjectOptions struct {
	Expiry        *scalar.Duration    `json:"expiry"`
	RequestParams map[string][]string `json:"requestParams,omitempty"`
}

// PresignedPutStorageObjectArguments represent the input arguments for the PresignedPutObject method.
type PresignedPutStorageObjectArguments struct {
	StorageBucketArguments

	Object string            `json:"object"`
	Expiry *scalar.Duration  `json:"expiry"`
	Where  schema.Expression `json:"where"  ndc:"predicate=StorageObjectSimple"`
}

// ListStorageObjectsArguments holds all arguments of a list object request.
type ListStorageObjectsArguments struct {
	StorageBucketArguments

	// Ignore '/' delimiter
	Recursive bool `json:"recursive,omitempty"`
	// The maximum number of objects requested per
	// batch, advanced use-case not useful for most
	// applications
	MaxResults *int `json:"maxResults"`
	// StartAfter start listing lexically at this object onwards.
	StartAfter *string `json:"startAfter,omitempty"`

	Where schema.Expression `json:"where" ndc:"predicate=StorageObjectSimple"`
}

// ListStorageObjectsOptions holds all options of a list object request.
type ListStorageObjectsOptions struct {
	// Include objects versions in the listing
	WithVersions bool `json:"withVersions"`
	// Include objects metadata in the listing
	WithMetadata bool `json:"withMetadata"`
	// Include user tags in the listing
	WithTags bool `json:"withTags"`
	// Only list objects with the prefix
	Prefix string `json:"prefix"`
	// Ignore '/' delimiter
	Recursive bool `json:"recursive"`
	// The maximum number of objects requested per
	// batch, advanced use-case not useful for most
	// applications
	MaxResults int `json:"maxResults"`
	// StartAfter start listing lexically at this object onwards.
	StartAfter string `json:"startAfter"`
}

// GetStorageObjectArguments are used to specify additional headers or options during GET requests.
type GetStorageObjectArguments struct {
	StorageBucketArguments
	GetStorageObjectOptions

	Object string            `json:"object"`
	Where  schema.Expression `json:"where"  ndc:"predicate=StorageObjectSimple"`
}

// GetStorageObjectOptions are used to specify additional headers or options during GET requests.
type GetStorageObjectOptions struct {
	Headers       map[string]string   `json:"headers,omitempty"`
	RequestParams map[string][]string `json:"requestParams,omitempty"`
	// ServerSideEncryption *ServerSideEncryptionMethod `json:"serverSideEncryption"`
	VersionID  *string `json:"versionId"`
	PartNumber *int    `json:"partNumber"`

	// Include any checksums, if object was uploaded with checksum.
	// For multipart objects this is a checksum of part checksums.
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide/checking-object-integrity.html
	Checksum *bool `json:"checksum"`

	// Include user tags in the listing
	WithTags *bool `json:"withTags"`
}

// StorageCopyDestOptions represents options specified by user for CopyObject/ComposeObject APIs.
type StorageCopyDestOptions struct {
	// points to destination bucket
	Bucket string `json:"bucket,omitempty"`
	// points to destination object
	Object string `json:"object"`

	// `Encryption` is the key info for server-side-encryption with customer
	// provided key. If it is nil, no encryption is performed.
	// Encryption *ServerSideEncryptionMethod `json:"encryption"`

	// `userMeta` is the user-metadata key-value pairs to be set on the
	// destination. The keys are automatically prefixed with `x-amz-meta-`
	// if needed. If nil is passed, and if only a single source (of any
	// size) is provided in the ComposeObject call, then metadata from the
	// source is copied to the destination.
	// if no user-metadata is provided, it is copied from source
	// (when there is only once source object in the compose
	// request)
	UserMetadata map[string]string `json:"userMetadata,omitempty"`
	// UserMetadata is only set to destination if ReplaceMetadata is true
	// other value is UserMetadata is ignored and we preserve src.UserMetadata
	// NOTE: if you set this value to true and now metadata is present
	// in UserMetadata your destination object will not have any metadata
	// set.
	ReplaceMetadata bool `json:"replaceMetadata,omitempty"`

	// `userTags` is the user defined object tags to be set on destination.
	// This will be set only if the `replaceTags` field is set to true.
	// Otherwise this field is ignored
	UserTags    map[string]string `json:"userTags,omitempty"`
	ReplaceTags bool              `json:"replaceTags,omitempty"`

	// Specifies whether you want to apply a Legal Hold to the copied object.
	LegalHold *StorageLegalHoldStatus `json:"legalHold"`

	// Object Retention related fields
	Mode            *StorageRetentionMode `json:"mode"`
	RetainUntilDate *time.Time            `json:"retainUntilDate"`

	// Needs to be specified if progress bar is specified.
	Size int64 `json:"size,omitempty"`
}

// StorageCopySrcOptions represents a source object to be copied, using server-side copying APIs.
type StorageCopySrcOptions struct {
	// source bucket
	Bucket string `json:"bucket,omitempty"`
	// source object
	Object string `json:"object"`

	VersionID            string     `json:"versionId,omitempty"`
	MatchETag            string     `json:"matchETag,omitempty"`
	NoMatchETag          string     `json:"noMatchETag,omitempty"`
	MatchModifiedSince   *time.Time `json:"matchModifiedSince"`
	MatchUnmodifiedSince *time.Time `json:"matchUnmodifiedSince"`
	MatchRange           bool       `json:"matchRange,omitempty"`
	Start                int64      `json:"start,omitempty"`
	End                  int64      `json:"end,omitempty"`
	// Encryption           *ServerSideEncryptionMethod `json:"encryption"`
}

// RemoveStorageObjectArguments represent arguments specified by user for RemoveObject call.
type RemoveStorageObjectArguments struct {
	StorageBucketArguments
	RemoveStorageObjectOptions

	Object string            `json:"object"`
	Where  schema.Expression `json:"where"  ndc:"predicate=StorageObjectSimple"`
}

// RemoveStorageObjectOptions represents options specified by user for RemoveObject call.
type RemoveStorageObjectOptions struct {
	ForceDelete      bool   `json:"forceDelete,omitempty"`
	GovernanceBypass bool   `json:"governanceBypass,omitempty"`
	VersionID        string `json:"versionId,omitempty"`
}

// PutStorageObjectRetentionOptions represents options specified by user for PutObject call.
type PutStorageObjectRetentionOptions struct {
	StorageBucketArguments

	Object           string                `json:"object"`
	GovernanceBypass bool                  `json:"governanceBypass,omitempty"`
	Mode             *StorageRetentionMode `json:"mode"`
	RetainUntilDate  *time.Time            `json:"retainUntilDate,omitempty"`
	VersionID        string                `json:"versionId,omitempty"`
}

// RemoveStorageObjectsArguments represents arguments specified by user for RemoveObjects call.
type RemoveStorageObjectsArguments struct {
	StorageBucketArguments
	RemoveStorageObjectsOptions

	Where schema.Expression `json:"where" ndc:"predicate=StorageObjectSimple"`
}

// RemoveStorageObjectsOptions represents options specified by user for RemoveObjects call.
type RemoveStorageObjectsOptions struct {
	ListStorageObjectsOptions

	GovernanceBypass bool `json:"governanceBypass,omitempty"`
}

// PutStorageObjectLegalHoldOptions represents options specified by user for PutObjectLegalHold call.
type PutStorageObjectLegalHoldOptions struct {
	StorageBucketArguments

	Object    string                  `json:"object"`
	VersionID string                  `json:"versionId,omitempty"`
	Status    *StorageLegalHoldStatus `json:"status"`
}

// GetStorageObjectLegalHoldOptions represents options specified by user for GetObjectLegalHold call.
type GetStorageObjectLegalHoldOptions struct {
	StorageBucketArguments

	Object    string `json:"object"`
	VersionID string `json:"versionId,omitempty"`
}

// PutStorageObjectOptions represents options specified by user for PutObject call.
type PutStorageObjectOptions struct {
	UserMetadata       map[string]string     `json:"userMetadata,omitempty"`
	UserTags           map[string]string     `json:"userTags,omitempty"`
	ContentType        string                `json:"contentType,omitempty"`
	ContentEncoding    string                `json:"contentEncoding,omitempty"`
	ContentDisposition string                `json:"contentDisposition,omitempty"`
	ContentLanguage    string                `json:"contentLanguage,omitempty"`
	CacheControl       string                `json:"cacheControl,omitempty"`
	Expires            *time.Time            `json:"expires,omitempty"`
	Mode               *StorageRetentionMode `json:"mode,omitempty"`
	RetainUntilDate    *time.Time            `json:"retainUntilDate,omitempty"`
	// ServerSideEncryption    *ServerSideEncryptionMethod `json:"serverSideEncryption,omitempty"`
	NumThreads              uint                    `json:"numThreads,omitempty"`
	StorageClass            string                  `json:"storageClass,omitempty"`
	WebsiteRedirectLocation string                  `json:"websiteRedirectLocation,omitempty"`
	PartSize                uint64                  `json:"partSize,omitempty"`
	LegalHold               *StorageLegalHoldStatus `json:"legalHold"`
	SendContentMd5          bool                    `json:"sendContentMd5,omitempty"`
	DisableContentSha256    bool                    `json:"disableContentSha256,omitempty"`
	DisableMultipart        bool                    `json:"disableMultipart,omitempty"`

	// AutoChecksum is the type of checksum that will be added if no other checksum is added,
	// like MD5 or SHA256 streaming checksum, and it is feasible for the upload type.
	// If none is specified CRC32C is used, since it is generally the fastest.
	AutoChecksum *ChecksumType `json:"autoChecksum"`

	// Checksum will force a checksum of the specific type.
	// This requires that the client was created with "TrailingHeaders:true" option,
	// and that the destination server supports it.
	// Unavailable with V2 signatures & Google endpoints.
	// This will disable content MD5 checksums if set.
	Checksum *ChecksumType `json:"checksum"`

	// ConcurrentStreamParts will create NumThreads buffers of PartSize bytes,
	// fill them serially and upload them in parallel.
	// This can be used for faster uploads on non-seekable or slow-to-seek input.
	ConcurrentStreamParts bool `json:"concurrentStreamParts,omitempty"`
}

// SetStorageObjectTagsArguments holds an object version id to update tag(s) of a specific object version.
type SetStorageObjectTagsArguments struct {
	StorageBucketArguments
	SetStorageObjectTagsOptions

	Object string            `json:"object"`
	Where  schema.Expression `json:"where"  ndc:"predicate=StorageObjectSimple"`
}

// SetStorageObjectTagsOptions holds an object version id to update tag(s) of a specific object version.
type SetStorageObjectTagsOptions struct {
	Tags      map[string]string `json:"tags"`
	VersionID string            `json:"versionId,omitempty"`
}

// SetBucketNotificationArguments represents input arguments for the SetBucketNotification method.
type SetBucketNotificationArguments struct {
	StorageBucketArguments
	NotificationConfig
}

// SetStorageBucketLifecycleArguments represents input arguments for the SetBucketLifecycle method.
type SetStorageBucketLifecycleArguments struct {
	StorageBucketArguments
	BucketLifecycleConfiguration
}

// SetStorageBucketEncryptionArguments represents input arguments for the SetBucketEncryption method.
type SetStorageBucketEncryptionArguments struct {
	StorageBucketArguments
	ServerSideEncryptionConfiguration
}

// SetStorageObjectLockArguments represents input arguments for the SetStorageObjectLock method.
type SetStorageObjectLockArguments struct {
	StorageBucketArguments
	SetStorageObjectLockConfig
}

// SetStorageBucketReplicationArguments storage bucket replication arguments.
type SetStorageBucketReplicationArguments struct {
	StorageBucketArguments
	StorageReplicationConfig
}

// PresignedURLResponse holds the presigned URL and expiry information.
type PresignedURLResponse struct {
	URL       string `json:"url"`
	ExpiredAt string `json:"expiredAt"`
}
