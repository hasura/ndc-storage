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
	StorageClientCredentialArguments

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
	StorageClientCredentialArguments

	Name  string            `json:"name"`
	Where schema.Expression `json:"where" ndc:"predicate=StorageBucketFilter"`
}

// ToStorageBucketArguments convert to the StorageBucketArguments instance.
func (arg GetStorageBucketArguments) ToStorageBucketArguments() *StorageBucketArguments {
	return &StorageBucketArguments{
		Bucket:                           arg.Name,
		StorageClientCredentialArguments: arg.StorageClientCredentialArguments,
	}
}

// StorageBucketArguments represent the common input arguments for bucket-related methods.
type StorageBucketArguments struct {
	// The bucket name.
	Bucket string `json:"bucket,omitempty"`

	StorageClientCredentialArguments
}

// StorageClientCredentials hold common storage client credential arguments.
type StorageClientCredentialArguments struct {
	ClientID        *StorageClientID     `json:"client_id,omitempty"`
	ClientType      *StorageProviderType `json:"client_type,omitempty"`
	Endpoint        string               `json:"endpoint,omitempty"`
	AccessKeyID     string               `json:"access_key_id,omitempty"`
	SecretAccessKey string               `json:"secret_access_key,omitempty"`
}

// IsEmpty checks if all properties are empty.
func (ca StorageClientCredentialArguments) IsEmpty() bool {
	return ca.ClientType == nil || !ca.ClientType.IsValid() || (ca.AccessKeyID == "" && ca.SecretAccessKey == "" && ca.Endpoint == "")
}

// MakeStorageBucketArguments holds all arguments to tweak bucket creation.
type MakeStorageBucketArguments struct {
	StorageClientCredentialArguments
	MakeStorageBucketOptions
}

// CopyStorageObjectArguments represent input arguments of the CopyObject method.
type CopyStorageObjectArguments struct {
	// The storage client ID
	ClientID *StorageClientID       `json:"client_id,omitempty"`
	Dest     StorageCopyDestOptions `json:"dest"`
	Source   StorageCopySrcOptions  `json:"source"`
}

// ComposeStorageObjectArguments represent input arguments of the ComposeObject method.
type ComposeStorageObjectArguments struct {
	// The storage client ID
	ClientID *StorageClientID        `json:"client_id,omitempty"`
	Dest     StorageCopyDestOptions  `json:"dest"`
	Sources  []StorageCopySrcOptions `json:"sources"`
}

// MakeStorageBucketOptions holds all options to tweak bucket creation.
type MakeStorageBucketOptions struct {
	// Bucket name
	Name string `json:"name"`
	// Bucket location
	Region string `json:"region,omitempty"`
	// Enable object locking
	ObjectLock bool `json:"object_lock,omitempty"`
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

	Name string `json:"name"`
}

// PresignedGetStorageObjectArguments represent the input arguments for the PresignedGetObject method.
type PresignedGetStorageObjectArguments struct {
	StorageBucketArguments
	PresignedGetStorageObjectOptions

	Name  string            `json:"name"`
	Where schema.Expression `json:"where" ndc:"predicate=StorageObjectFilter"`
}

// PresignedGetStorageObjectOptions represent the options for the PresignedGetObject method.
type PresignedGetStorageObjectOptions struct {
	Expiry        *scalar.DurationString `json:"expiry"`
	RequestParams []StorageKeyValue      `json:"request_params,omitempty"`
}

// PresignedPutStorageObjectArguments represent the input arguments for the PresignedPutObject method.
type PresignedPutStorageObjectArguments struct {
	StorageBucketArguments

	Name   string                 `json:"name"`
	Expiry *scalar.DurationString `json:"expiry"`
	Where  schema.Expression      `json:"where"  ndc:"predicate=StorageObjectFilter"`
}

// ListStorageObjectsArguments holds all arguments of a list object request.
type ListStorageObjectsArguments struct {
	StorageBucketArguments

	// Returns the list of objects with the prefix.
	Prefix string `json:"prefix,omitempty"`
	// Returns objects in the recursive order.
	Recursive bool `json:"recursive,omitempty"`
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
	// Find objects recursively.
	Recursive bool
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

	Name  string            `json:"name"`
	Where schema.Expression `json:"where" ndc:"predicate=StorageObjectFilter"`
}

// GetStorageObjectOptions are used to specify additional headers or options during GET requests.
type GetStorageObjectOptions struct {
	Headers       []StorageKeyValue `json:"headers,omitempty"`
	RequestParams []StorageKeyValue `json:"request_params,omitempty"`
	// ServerSideEncryption *ServerSideEncryptionMethod `json:"serverSideEncryption"`
	VersionID  *string `json:"version_id"`
	PartNumber *int    `json:"part_number"`
	// Options to be included for the object information.
	Include StorageObjectIncludeOptions `json:"-"`
}

// StorageCopyDestOptions represents options specified by user for CopyObject/ComposeObject APIs.
type StorageCopyDestOptions struct {
	// points to destination bucket
	Bucket string `json:"bucket,omitempty"`
	// points to destination object
	Name string `json:"name"`

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
	LegalHold *bool `json:"legal_hold"`

	// Object Retention related fields
	Mode            *StorageRetentionMode `json:"mode"`
	RetainUntilDate *time.Time            `json:"retain_until_date"`

	// Needs to be specified if progress bar is specified.
	Size int64 `json:"size,omitempty"`
}

// StorageCopySrcOptions represents a source object to be copied, using server-side copying APIs.
type StorageCopySrcOptions struct {
	// source bucket
	Bucket string `json:"bucket,omitempty"`
	// source object
	Name string `json:"name"`

	VersionID            string     `json:"version_id,omitempty"`
	MatchETag            string     `json:"match_etag,omitempty"`
	NoMatchETag          string     `json:"no_match_etag,omitempty"`
	MatchModifiedSince   *time.Time `json:"match_modified_since"`
	MatchUnmodifiedSince *time.Time `json:"match_unmodified_since"`
	MatchRange           bool       `json:"match_range,omitempty"`
	Start                int64      `json:"start,omitempty"`
	End                  int64      `json:"end,omitempty"`
	// Encryption           *ServerSideEncryptionMethod `json:"encryption"`
}

// RemoveStorageObjectArguments represent arguments specified by user for RemoveObject call.
type RemoveStorageObjectArguments struct {
	StorageBucketArguments
	RemoveStorageObjectOptions

	Name  string            `json:"name"`
	Where schema.Expression `json:"where" ndc:"predicate=StorageObjectFilter"`
}

// RemoveStorageObjectOptions represents options specified by user for RemoveObject call.
type RemoveStorageObjectOptions struct {
	SoftDelete       bool   `json:"soft_delete,omitempty"`
	ForceDelete      bool   `json:"force_delete,omitempty"`
	GovernanceBypass bool   `json:"governance_bypass,omitempty"`
	VersionID        string `json:"version_id,omitempty"`
}

// UpdateStorageObjectArguments represents options specified by user for updating object.
type UpdateStorageObjectArguments struct {
	StorageBucketArguments
	UpdateStorageObjectOptions

	Name  string            `json:"name"`
	Where schema.Expression `json:"where" ndc:"predicate=StorageObjectFilter"`
}

// UpdateStorageObjectOptions represents options specified by user for updating object.
type UpdateStorageObjectOptions struct {
	VersionID string                            `json:"version_id,omitempty"`
	Retention *SetStorageObjectRetentionOptions `json:"retention"`
	LegalHold *bool                             `json:"legal_hold"`
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
	GovernanceBypass bool                  `json:"governance_bypass,omitempty"`
	RetainUntilDate  *time.Time            `json:"retain_until_date,omitempty"`
}

// RemoveStorageObjectsArguments represents arguments specified by user for RemoveObjects call.
type RemoveStorageObjectsArguments struct {
	StorageBucketArguments
	ListStorageObjectsArguments

	GovernanceBypass bool `json:"governance_bypass,omitempty"`
}

// RemoveStorageObjectsOptions represents options specified by user for RemoveObjects call.
type RemoveStorageObjectsOptions struct {
	ListStorageObjectsOptions

	GovernanceBypass bool
}

// PutStorageObjectRetentionOptions represent options of object retention configuration.
type PutStorageObjectRetentionOptions struct {
	Mode             StorageRetentionMode `json:"mode"`
	RetainUntilDate  time.Time            `json:"retain_until_date"`
	GovernanceBypass bool                 `json:"governance_bypass,omitempty"`
}

// PutStorageObjectOptions represents options specified by user for PutObject call.
type PutStorageObjectOptions struct {
	Metadata           []StorageKeyValue                 `json:"metadata,omitempty"`
	Tags               []StorageKeyValue                 `json:"tags,omitempty"`
	ContentType        string                            `json:"content_type,omitempty"`
	ContentEncoding    string                            `json:"content_encoding,omitempty"`
	ContentDisposition string                            `json:"content_disposition,omitempty"`
	ContentLanguage    string                            `json:"content_language,omitempty"`
	CacheControl       string                            `json:"cache_control,omitempty"`
	Expires            *time.Time                        `json:"expires,omitempty"`
	Retention          *PutStorageObjectRetentionOptions `json:"retention,omitempty"`
	// ServerSideEncryption    *ServerSideEncryptionMethod `json:"serverSideEncryption,omitempty"`
	NumThreads              uint   `json:"num_threads,omitempty"`
	StorageClass            string `json:"storage_class,omitempty"`
	WebsiteRedirectLocation string `json:"website_redirect_location,omitempty"`
	PartSize                uint64 `json:"part_size,omitempty"`
	LegalHold               *bool  `json:"legal_hold"`
	SendContentMd5          bool   `json:"send_content_md5,omitempty"`
	DisableContentSha256    bool   `json:"disable_content_sha256,omitempty"`
	DisableMultipart        bool   `json:"disable_multipart,omitempty"`

	// AutoChecksum is the type of checksum that will be added if no other checksum is added,
	// like MD5 or SHA256 streaming checksum, and it is feasible for the upload type.
	// If none is specified CRC32C is used, since it is generally the fastest.
	AutoChecksum *ChecksumType `json:"auto_checksum"`

	// Checksum will force a checksum of the specific type.
	// This requires that the client was created with "TrailingHeaders:true" option,
	// and that the destination server supports it.
	// Unavailable with V2 signatures & Google endpoints.
	// This will disable content MD5 checksums if set.
	Checksum *ChecksumType `json:"checksum"`

	// ConcurrentStreamParts will create NumThreads buffers of PartSize bytes,
	// fill them serially and upload them in parallel.
	// This can be used for faster uploads on non-seekable or slow-to-seek input.
	ConcurrentStreamParts bool `json:"concurrent_stream_parts,omitempty"`
}

// PresignedURLResponse holds the presigned URL and expiry information.
type PresignedURLResponse struct {
	URL       string    `json:"url"`
	ExpiredAt time.Time `json:"expired_at"`
}

// UpdateBucketArguments hold update options for the bucket.
type UpdateBucketArguments struct {
	GetStorageBucketArguments
	UpdateStorageBucketOptions
}

// UpdateStorageBucketOptions hold update options for the bucket.
type UpdateStorageBucketOptions struct {
	Tags              *[]StorageKeyValue                 `json:"tags"`
	VersioningEnabled *bool                              `json:"versioning_enabled"`
	Lifecycle         *ObjectLifecycleConfiguration      `json:"lifecycle"`
	Encryption        *ServerSideEncryptionConfiguration `json:"encryption"`
	ObjectLock        *SetStorageObjectLockConfig        `json:"object_lock"`
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

	Name  string            `json:"name"`
	Where schema.Expression `json:"where" ndc:"predicate=StorageObjectFilter"`
}

// PutStorageObjectArguments represents input arguments of the PutObject method.
type PutStorageObjectArguments struct {
	StorageBucketArguments

	Name    string                  `json:"name"`
	Options PutStorageObjectOptions `json:"options,omitempty"`
	Where   schema.Expression       `json:"where"             ndc:"predicate=StorageObjectFilter"`
}

// UploadStorageObjectFromURLArguments represents input arguments of the UploadStorageObjectFromURL method.
type UploadStorageObjectFromURLArguments struct {
	PutStorageObjectArguments
	HTTPRequestOptions
}
