package common

import (
	"time"

	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
)

// StorageKeyValue represent a key-value string pair
type StorageKeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ListStorageBucketArguments represent the input arguments for the ListBuckets methods.
type ListStorageBucketArguments struct {
	// The storage client ID.
	ClientID *StorageClientID `json:"clientId,omitempty"`
	// Returns list of bucket with the prefix.
	Prefix string `json:"prefix,omitempty"`
	// The maximum number of objects requested per batch.
	First *int `json:"first"`
	// After start listing lexically at this bucket onwards.
	After string            `json:"after,omitempty"`
	Where schema.Expression `json:"where"           ndc:"predicate=StorageBucketFilter"`
}

// StorageBucketArguments represent the common input arguments for bucket-related methods.
type GetStorageBucketArguments struct {
	StorageBucketArguments

	Where schema.Expression `json:"where" ndc:"predicate=StorageBucketFilter"`
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
	ObjectLock bool `json:"objectLock,omitempty"`
	// Optional tags
	Tags []StorageKeyValue `json:"tags,omitempty"`
}

// ListIncompleteUploadsArguments the input arguments of the ListIncompleteUploads method.
type ListIncompleteUploadsArguments struct {
	StorageBucketArguments
	ListIncompleteUploadsOptions
}

// ListIncompleteUploadsOptions the input arguments of the ListIncompleteUploads method.
type ListIncompleteUploadsOptions struct {
	Prefix string `json:"prefix,omitempty"`
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
	Where  schema.Expression `json:"where"  ndc:"predicate=StorageObjectFilter"`
}

// PresignedGetStorageObjectOptions represent the options for the PresignedGetObject method.
type PresignedGetStorageObjectOptions struct {
	Expiry        *scalar.DurationString `json:"expiry"`
	RequestParams []StorageKeyValue      `json:"requestParams,omitempty"`
}

// PresignedPutStorageObjectArguments represent the input arguments for the PresignedPutObject method.
type PresignedPutStorageObjectArguments struct {
	StorageBucketArguments

	Object string                 `json:"object"`
	Expiry *scalar.DurationString `json:"expiry"`
	Where  schema.Expression      `json:"where"  ndc:"predicate=StorageObjectFilter"`
}

// ListStorageObjectsArguments holds all arguments of a list object request.
type ListStorageObjectsArguments struct {
	StorageBucketArguments

	// Returns the list of objects with the prefix.
	Prefix string `json:"prefix,omitempty"`
	// Returns objects in the hierarchical order.
	Hierarchy bool `json:"hierarchy,omitempty"`
	// The maximum number of objects requested per batch.
	First *int `json:"first"`
	// After start listing lexically at this object onwards.
	After *string `json:"after,omitempty"`

	Where schema.Expression `json:"where" ndc:"predicate=StorageObjectFilter"`
}

// StorageObjectIncludeOptions hold options to be included for the object information.
type StorageObjectIncludeOptions struct {
	// Include any checksums, if object was uploaded with checksum.
	// For multipart objects this is a checksum of part checksums.
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide/checking-object-integrity.html
	Checksum bool
	// Include user tags in the listing
	Tags bool
	// Include objects versions in the listing
	Versions bool
	// Include objects metadata in the listing
	Metadata bool

	Copy        bool
	Snapshots   bool
	LegalHold   bool
	Retention   bool
	Permissions bool
	Lifecycle   bool
	Encryption  bool
}

// IsEmpty checks if all include options are empty
func (soi StorageObjectIncludeOptions) IsEmpty() bool {
	return !soi.Checksum && !soi.Tags && !soi.Versions && !soi.Metadata &&
		!soi.Copy && !soi.Snapshots && !soi.LegalHold && !soi.Retention && !soi.Permissions &&
		!soi.Lifecycle && !soi.Encryption
}

// ListStorageObjectsOptions holds all options of a list object request.
type ListStorageObjectsOptions struct {
	// Only list objects with the prefix
	Prefix string
	// Returns objects in the hierarchical order.
	Hierarchy bool
	// The maximum number of objects requested per
	// batch, advanced use-case not useful for most
	// applications
	MaxResults int
	// StartAfter start listing lexically at this object onwards.
	StartAfter string
	// Options to be included for the object information.
	Include    StorageObjectIncludeOptions
	NumThreads int
}

// GetStorageObjectArguments are used to specify additional headers or options during GET requests.
type GetStorageObjectArguments struct {
	StorageBucketArguments
	GetStorageObjectOptions

	Object string            `json:"object"`
	Where  schema.Expression `json:"where"  ndc:"predicate=StorageObjectFilter"`
}

// GetStorageObjectOptions are used to specify additional headers or options during GET requests.
type GetStorageObjectOptions struct {
	Headers       []StorageKeyValue `json:"headers,omitempty"`
	RequestParams []StorageKeyValue `json:"requestParams,omitempty"`
	// ServerSideEncryption *ServerSideEncryptionMethod `json:"serverSideEncryption"`
	VersionID  *string `json:"versionId"`
	PartNumber *int    `json:"partNumber"`
	// Options to be included for the object information.
	Include       StorageObjectIncludeOptions `json:"-"`
	Base64Encoded bool                        `json:"-"`
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
	Metadata []StorageKeyValue `json:"metadata,omitempty"`

	// `tags` is the user defined object tags to be set on destination.
	Tags []StorageKeyValue `json:"tags,omitempty"`

	// Specifies whether you want to apply a Legal Hold to the copied object.
	LegalHold *bool `json:"legalHold"`

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
	Where  schema.Expression `json:"where"  ndc:"predicate=StorageObjectFilter"`
}

// RemoveStorageObjectOptions represents options specified by user for RemoveObject call.
type RemoveStorageObjectOptions struct {
	SoftDelete       bool   `json:"softDelete,omitempty"`
	ForceDelete      bool   `json:"forceDelete,omitempty"`
	GovernanceBypass bool   `json:"governanceBypass,omitempty"`
	VersionID        string `json:"versionId,omitempty"`
}

// UpdateStorageObjectArguments represents options specified by user for updating object.
type UpdateStorageObjectArguments struct {
	StorageBucketArguments
	UpdateStorageObjectOptions

	Object string            `json:"object"`
	Where  schema.Expression `json:"where"  ndc:"predicate=StorageObjectFilter"`
}

// UpdateStorageObjectOptions represents options specified by user for updating object.
type UpdateStorageObjectOptions struct {
	VersionID string                            `json:"versionId,omitempty"`
	Retention *SetStorageObjectRetentionOptions `json:"retention"`
	LegalHold *bool                             `json:"legalHold"`
	Metadata  *[]StorageKeyValue                `json:"metadata"`
	Tags      *[]StorageKeyValue                `json:"tags"`
}

// IsEmpty checks if all elements in the option object is null.
func (ubo UpdateStorageObjectOptions) IsEmpty() bool {
	return ubo.VersionID == "" &&
		ubo.Tags == nil &&
		ubo.Retention == nil &&
		ubo.LegalHold == nil &&
		ubo.Metadata == nil
}

// SetStorageObjectRetentionOptions represents options specified by user for PutObject call.
type SetStorageObjectRetentionOptions struct {
	Mode             *StorageRetentionMode `json:"mode"`
	GovernanceBypass bool                  `json:"governanceBypass,omitempty"`
	RetainUntilDate  *time.Time            `json:"retainUntilDate,omitempty"`
}

// RemoveStorageObjectsArguments represents arguments specified by user for RemoveObjects call.
type RemoveStorageObjectsArguments struct {
	StorageBucketArguments
	ListStorageObjectsArguments

	GovernanceBypass bool `json:"governanceBypass,omitempty"`
}

// RemoveStorageObjectsOptions represents options specified by user for RemoveObjects call.
type RemoveStorageObjectsOptions struct {
	ListStorageObjectsOptions

	GovernanceBypass bool
}

// PutStorageObjectRetentionOptions represent options of object retention configuration.
type PutStorageObjectRetentionOptions struct {
	Mode             StorageRetentionMode `json:"mode"`
	RetainUntilDate  time.Time            `json:"retainUntilDate"`
	GovernanceBypass bool                 `json:"governanceBypass,omitempty"`
}

// PutStorageObjectOptions represents options specified by user for PutObject call.
type PutStorageObjectOptions struct {
	Metadata           []StorageKeyValue                 `json:"metadata,omitempty"`
	Tags               []StorageKeyValue                 `json:"tags,omitempty"`
	ContentType        string                            `json:"contentType,omitempty"`
	ContentEncoding    string                            `json:"contentEncoding,omitempty"`
	ContentDisposition string                            `json:"contentDisposition,omitempty"`
	ContentLanguage    string                            `json:"contentLanguage,omitempty"`
	CacheControl       string                            `json:"cacheControl,omitempty"`
	Expires            *time.Time                        `json:"expires,omitempty"`
	Retention          *PutStorageObjectRetentionOptions `json:"retention,omitempty"`
	// ServerSideEncryption    *ServerSideEncryptionMethod `json:"serverSideEncryption,omitempty"`
	NumThreads              uint   `json:"numThreads,omitempty"`
	StorageClass            string `json:"storageClass,omitempty"`
	WebsiteRedirectLocation string `json:"websiteRedirectLocation,omitempty"`
	PartSize                uint64 `json:"partSize,omitempty"`
	LegalHold               *bool  `json:"legalHold"`
	SendContentMd5          bool   `json:"sendContentMd5,omitempty"`
	DisableContentSha256    bool   `json:"disableContentSha256,omitempty"`
	DisableMultipart        bool   `json:"disableMultipart,omitempty"`

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

// PresignedURLResponse holds the presigned URL and expiry information.
type PresignedURLResponse struct {
	URL       string    `json:"url"`
	ExpiredAt time.Time `json:"expiredAt"`
}

// UpdateBucketArguments hold update options for the bucket.
type UpdateBucketArguments struct {
	StorageBucketArguments
	UpdateStorageBucketOptions
}

// UpdateStorageBucketOptions hold update options for the bucket.
type UpdateStorageBucketOptions struct {
	Tags              *[]StorageKeyValue                 `json:"tags"`
	VersioningEnabled *bool                              `json:"versioningEnabled"`
	Lifecycle         *ObjectLifecycleConfiguration      `json:"lifecycle"`
	Encryption        *ServerSideEncryptionConfiguration `json:"encryption"`
	ObjectLock        *SetStorageObjectLockConfig        `json:"objectLock"`
}

// IsEmpty checks if all elements in the option object is null.
func (ubo UpdateStorageBucketOptions) IsEmpty() bool {
	return ubo.VersioningEnabled == nil &&
		ubo.Tags == nil &&
		ubo.Lifecycle == nil &&
		ubo.Encryption == nil &&
		ubo.ObjectLock == nil
}

// RestoreStorageObjectArguments represent arguments specified by user for RestoreObject call.
type RestoreStorageObjectArguments struct {
	StorageBucketArguments

	Object string            `json:"object"`
	Where  schema.Expression `json:"where"  ndc:"predicate=StorageObjectFilter"`
}
