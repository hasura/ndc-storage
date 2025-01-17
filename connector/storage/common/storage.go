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
	ListBuckets(ctx context.Context, options BucketOptions) ([]StorageBucketInfo, error)
	// GetBucket gets a bucket by name.
	GetBucket(ctx context.Context, name string, options BucketOptions) (*StorageBucketInfo, error)
	// BucketExists checks if a bucket exists.
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	// RemoveBucket removes a bucket, bucket should be empty to be successfully removed.
	RemoveBucket(ctx context.Context, bucketName string) error
	// UpdateBucket updates configurations for the bucket.
	UpdateBucket(ctx context.Context, bucketName string, opts UpdateStorageBucketOptions) error
	// GetBucketPolicy gets access permissions on a bucket or a prefix.
	GetBucketPolicy(ctx context.Context, bucketName string) (string, error)
	// ListObjects lists objects in a bucket.
	ListObjects(ctx context.Context, bucketName string, opts *ListStorageObjectsOptions, predicate func(string) bool) (*StorageObjectListResults, error)
	// ListIncompleteUploads list partially uploaded objects in a bucket.
	ListIncompleteUploads(ctx context.Context, bucketName string, args ListIncompleteUploadsOptions) ([]StorageObjectMultipartInfo, error)
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
	// GetBucketNotification gets notification configuration on a bucket.
	GetBucketNotification(ctx context.Context, bucketName string) (*NotificationConfig, error)
	// Set a new bucket notification on a bucket.
	SetBucketNotification(ctx context.Context, bucketName string, config NotificationConfig) error
	// Remove all configured bucket notifications on a bucket.
	RemoveAllBucketNotification(ctx context.Context, bucketName string) error
	// SetBucketReplication sets replication configuration on a bucket. Role can be obtained by first defining the replication target on MinIO
	// to associate the source and destination buckets for replication with the replication endpoint.
	SetBucketReplication(ctx context.Context, bucketname string, cfg StorageReplicationConfig) error
	// GetBucketReplication gets current replication config on a bucket.
	GetBucketReplication(ctx context.Context, bucketName string) (*StorageReplicationConfig, error)
	// RemoveBucketReplication removes replication configuration on a bucket.
	RemoveBucketReplication(ctx context.Context, bucketName string) error
}

// BucketOptions hold options to get bucket information.
type BucketOptions struct {
	Prefix            string
	IncludeTags       bool
	IncludeVersioning bool
	IncludeLifecycle  bool
	IncludeEncryption bool
	IncludeObjectLock bool
	NumThreads        int
}

// StorageBucketInfo container for bucket metadata.
type StorageBucketInfo struct {
	// The name of the bucket.
	Name string `json:"name"`
	// Date the bucket was created.
	CreationDate time.Time `json:"creationDate"`
	// Bucket tags or metadata.
	Tags map[string]string `json:"tags,omitempty"`
	// The versioning configuration
	Versioning *StorageBucketVersioningConfiguration `json:"versioning"`
	// The versioning configuration
	Lifecycle *BucketLifecycleConfiguration `json:"lifecycle"`
	// The server-side encryption configuration.
	Encryption *ServerSideEncryptionConfiguration `json:"encryption"`
	ObjectLock *StorageObjectLockConfig           `json:"objectLock"`
}

// StorageOwner name.
type StorageOwner struct {
	DisplayName *string `json:"name"`
	ID          *string `json:"id"`
}

// StorageGrantee represents the person being granted permissions.
type StorageGrantee struct {
	ID          *string `json:"id"`
	DisplayName *string `json:"displayName"`
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
	OngoingRestore bool `json:"ongoingRestore"`
	// When the restored copy of the archived object will be removed
	ExpiryTime *time.Time `json:"expiryTime"`
}

// StorageObject container for object metadata.
type StorageObject struct {
	// An ETag is optionally set to md5sum of an object.  In case of multipart objects,
	// ETag is of the form MD5SUM-N where MD5SUM is md5sum of all individual md5sums of
	// each parts concatenated into one string.
	ETag *string `json:"etag"`

	ClientID           string     `json:"clientId"`     // Client ID
	Bucket             string     `json:"bucket"`       // Name of the bucket
	Name               string     `json:"name"`         // Name of the object
	LastModified       time.Time  `json:"lastModified"` // Date and time the object was last modified.
	Size               *int64     `json:"size"`         // Size in bytes of the object.
	ContentType        *string    `json:"contentType"`  // A standard MIME type describing the format of the object data.
	ContentEncoding    *string    `json:"contentEncoding,omitempty"`
	ContentDisposition *string    `json:"contentDisposition,omitempty"`
	ContentLanguage    *string    `json:"contentLanguage,omitempty"`
	CacheControl       *string    `json:"cacheControl,omitempty"`
	Expires            *time.Time `json:"expires"` // The date and time at which the object is no longer able to be cached.

	// Collection of additional metadata on the object.
	// eg: x-amz-meta-*, content-encoding etc.
	Metadata map[string]string `json:"metadata,omitempty"`

	// x-amz-meta-* headers stripped "x-amz-meta-" prefix containing the first value.
	// Only returned by MinIO servers.
	UserMetadata map[string]string `json:"userMetadata,omitempty"`

	// x-amz-tagging values in their k/v values.
	// Only returned by MinIO servers.
	UserTags map[string]string `json:"userTags,omitempty"`

	// x-amz-tagging-count value
	UserTagCount int `json:"userTagCount,omitempty"`

	// Owner name.
	Owner *StorageOwner `json:"owner"`

	// ACL grant.
	Grant []StorageGrant `json:"grant,omitempty"`

	// The class of storage used to store the object or the access tier on Azure blob storage.
	StorageClass *string `json:"storageClass,omitempty"`

	// Versioning related information
	IsLatest  *bool   `json:"isLatest"`
	Deleted   *bool   `json:"deleted"`
	VersionID *string `json:"versionId,omitempty"`

	// x-amz-replication-status value is either in one of the following states
	ReplicationStatus *StorageObjectReplicationStatus `json:"replicationStatus"`
	// set to true if delete marker has backing object version on target, and eligible to replicate
	ReplicationReady *bool `json:"replicationReady"`
	// Lifecycle expiry-date and ruleID associated with the expiry
	// not to be confused with `Expires` HTTP header.
	Expiration       *time.Time `json:"expiration"`
	ExpirationRuleID *string    `json:"expirationRuleId"`

	Restore *StorageRestoreInfo `json:"restore"`

	// Checksum values
	StorageObjectChecksum

	// Azure Blob Store properties
	ACL                       *string    `json:"acl"`
	AccessTierChangeTime      *time.Time `json:"accessTierChangeTime"`
	AccessTierInferred        *bool      `json:"accessTierInferred"`
	ArchiveStatus             *string    `json:"archiveStatus"`
	BlobSequenceNumber        *int64     `json:"blobSequenceNumber"`
	BlobType                  *string    `json:"blobType"`
	ContentMD5                *string    `json:"contentMd5"`
	CopyCompletionTime        *time.Time `json:"copyCompletionTime"`
	CopyID                    *string    `json:"copyId"`
	CopyProgress              *string    `json:"copyProgress"`
	CopySource                *string    `json:"copySource"`
	CopyStatus                *string    `json:"copyStatus"`
	CopyStatusDescription     *string    `json:"copyStatusDescription"`
	CreationTime              *time.Time `json:"creationTime"`
	DeletedTime               *time.Time `json:"deletedTime"`
	CustomerProvidedKeySHA256 *string    `json:"customerProvidedKeySha256"`
	DestinationSnapshot       *string    `json:"destinationSnapshot"`

	// The name of the encryption scope under which the blob is encrypted.
	KMSKeyName         *string    `json:"kmsKeyName"`
	ServerEncrypted    *bool      `json:"serverEncrypted"`
	Group              *string    `json:"group"`
	RetentionUntilDate *time.Time `json:"retentionUntilDate"`
	RetentionMode      *string    `json:"retentionMode"`
	IncrementalCopy    *bool      `json:"incrementalCopy"`
	IsSealed           *bool      `json:"sealed"`
	LastAccessTime     *time.Time `json:"lastAccessTime"`
	LeaseDuration      *string    `json:"leaseDuration"`
	LeaseState         *string    `json:"leaseState"`
	LeaseStatus        *string    `json:"leaseStatus"`
	LegalHold          *bool      `json:"legalHold"`
	Permissions        *string    `json:"permissions"`

	// If an object is in rehydrate pending state then this header is returned with priority of rehydrate. Valid values are High
	// and Standard.
	RehydratePriority      *string `json:"rehydratePriority"`
	RemainingRetentionDays *int32  `json:"remainingRetentionDays"`
	ResourceType           *string `json:"resourceType"`
}

// StorageObjectPaginationInfo holds the pagination information.
type StorageObjectPaginationInfo struct {
	HasNextPage bool    `json:"hasNextPage"`
	Cursor      *string `json:"cursor"`
	NextCursor  *string `json:"nextCursor"`
}

// StorageObjectListResults hold the paginated results of the storage object list.
type StorageObjectListResults struct {
	Objects  []StorageObject             `json:"objects"`
	PageInfo StorageObjectPaginationInfo `json:"pageInfo"`
}

// StorageObjectReplicationStatus represents the x-amz-replication-status value enum.
// @enum COMPLETED,PENDING,FAILED,REPLICA
type StorageObjectReplicationStatus string

// StorageObjectChecksum represents checksum values of the object.
type StorageObjectChecksum struct {
	ChecksumCRC32     *string `json:"checksumCrc32,omitempty"`
	ChecksumCRC32C    *string `json:"checksumCrc32C,omitempty"`
	ChecksumSHA1      *string `json:"checksumSha1,omitempty"`
	ChecksumSHA256    *string `json:"checksumSha256,omitempty"`
	ChecksumCRC64NVME *string `json:"checksumCrc64Nvme,omitempty"`
}

// StorageUploadInfo represents the information of the uploaded object.
type StorageUploadInfo struct {
	// An ETag is optionally set to md5sum of an object.  In case of multipart objects,
	// ETag is of the form MD5SUM-N where MD5SUM is md5sum of all individual md5sums of
	// each parts concatenated into one string.
	ETag *string `json:"etag"`

	ClientID     string     `json:"clientId"`     // Client ID
	Bucket       string     `json:"bucket"`       // Name of the bucket
	Name         string     `json:"name"`         // Name of the object
	LastModified *time.Time `json:"lastModified"` // Date and time the object was last modified.
	Size         *int64     `json:"size"`         // Size in bytes of the object.
	Location     *string    `json:"location"`
	VersionID    *string    `json:"versionId"`
	ContentMD5   *string    `json:"contentMd5"`

	// Lifecycle expiry-date and ruleID associated with the expiry
	// not to be confused with `Expires` HTTP header.
	Expiration       *time.Time `json:"expiration"`
	ExpirationRuleID *string    `json:"expirationRuleId"`

	// Checksum values
	StorageObjectChecksum
}

// StorageObjectMultipartInfo container for multipart object metadata.
type StorageObjectMultipartInfo struct {
	// Date and time at which the multipart upload was initiated.
	Initiated *time.Time `json:"initiated"`

	// The type of storage to use for the object. Defaults to 'STANDARD'.
	StorageClass *string `json:"storageClass"`

	// Key of the object for which the multipart upload was initiated.
	Name *string `json:"name"`

	// Size in bytes of the object.
	Size *int64 `json:"size"`

	// Upload ID that identifies the multipart upload.
	UploadID *string `json:"uploadId"`
}

// EncryptionMethod represents a server-side-encryption method enum.
// @enum SSE_C,KMS,S3
// type ServerSideEncryptionMethod string

// StorageRetentionMode the object retention mode.
// @enum Locked,Unlocked,Mutable
type StorageRetentionMode string

// RemoveStorageObjectError the container of Multi Delete S3 API error.
type RemoveStorageObjectError struct {
	ObjectName string `json:"objectName"`
	VersionID  string `json:"versionId"`
	Error      error  `json:"error"`
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
	Lambda string `json:"cloudFunction"`
}

// NotificationConfig the struct that represents a notification configration object.
type NotificationConfig struct {
	LambdaConfigs []NotificationLambdaConfig `json:"cloudFunctionConfigurations"`
	TopicConfigs  []NotificationTopicConfig  `json:"topicConfigurations"`
	QueueConfigs  []NotificationQueueConfig  `json:"queueConfigurations"`
}

// NotificationFilter - a tag in the notification xml structure which carries suffix/prefix filters
type NotificationFilter struct {
	S3Key *NotificationS3Key `json:"s3Key,omitempty"`
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
	FilterRules []NotificationFilterRule `json:"filterRule,omitempty"`
}

// BucketLifecycleRule represents a single rule in lifecycle configuration
type BucketLifecycleRule struct {
	AbortIncompleteMultipartUpload *AbortIncompleteMultipartUpload       `json:"abortIncompleteMultipartUpload"`
	Expiration                     *LifecycleExpiration                  `json:"expiration,omitempty"`
	DelMarkerExpiration            *LifecycleDelMarkerExpiration         `json:"delMarkerExpiration,omitempty"`
	AllVersionsExpiration          *LifecycleAllVersionsExpiration       `json:"allVersionsExpiration,omitempty"`
	ID                             string                                `json:"id"`
	RuleFilter                     *LifecycleFilter                      `json:"filter,omitempty"`
	NoncurrentVersionExpiration    *LifecycleNoncurrentVersionExpiration `json:"noncurrentVersionExpiration,omitempty"`
	NoncurrentVersionTransition    *LifecycleNoncurrentVersionTransition `json:"noncurrentVersionTransition,omitempty"`
	Prefix                         *string                               `json:"prefix,omitempty"`
	Status                         *string                               `json:"status"`
	Transition                     *LifecycleTransition                  `json:"transition,omitempty"`
}

// BucketLifecycleConfiguration is a collection of lifecycle Rule objects.
type BucketLifecycleConfiguration struct {
	Rules []BucketLifecycleRule `json:"rules"`
}

// AbortIncompleteMultipartUpload structure, not supported yet on MinIO
type AbortIncompleteMultipartUpload struct {
	DaysAfterInitiation *int `json:"daysAfterInitiation"`
}

// LifecycleExpiration expiration details of lifecycle configuration
type LifecycleExpiration struct {
	Date         *scalar.Date `json:"date,omitempty"`
	Days         *int         `json:"days,omitempty"`
	DeleteMarker *bool        `json:"expiredObjectDeleteMarker,omitempty"`
	DeleteAll    *bool        `json:"expiredObjectAllVersions,omitempty"`
}

// IsEmpty checks if all properties of the object are empty.
func (fe LifecycleExpiration) IsEmpty() bool {
	return fe.DeleteAll == nil && fe.Date == nil && fe.Days == nil && fe.DeleteMarker == nil
}

// LifecycleTransition transition details of lifecycle configuration
type LifecycleTransition struct {
	Date         *scalar.Date `json:"date"`
	StorageClass *string      `json:"storageClass"`
	Days         *int         `json:"days"`
}

// IsEmpty checks if all properties of the object are empty.
func (fe LifecycleTransition) IsEmpty() bool {
	return fe.StorageClass == nil && fe.Date == nil && fe.Days == nil
}

// LifecycleDelMarkerExpiration represents DelMarkerExpiration actions element in an ILM policy
type LifecycleDelMarkerExpiration struct {
	Days *int `json:"days"`
}

// LifecycleAllVersionsExpiration represents AllVersionsExpiration actions element in an ILM policy
type LifecycleAllVersionsExpiration struct {
	Days         *int  `json:"days"`
	DeleteMarker *bool `json:"deleteMarker"`
}

// LifecycleFilter will be used in selecting rule(s) for lifecycle configuration
type LifecycleFilter struct {
	And                   *LifecycleFilterAnd `json:"and,omitempty"`
	Prefix                *string             `json:"prefix,omitempty"`
	Tag                   *StorageTag         `json:"tag,omitempty"`
	ObjectSizeLessThan    *int64              `json:"objectSizeLessThan,omitempty"`
	ObjectSizeGreaterThan *int64              `json:"objectSizeGreaterThan,omitempty"`
}

// LifecycleFilterAnd the And Rule for LifecycleTag, to be used in LifecycleRuleFilter
type LifecycleFilterAnd struct {
	Prefix                *string      `json:"prefix,omitempty"`
	Tags                  []StorageTag `json:"tags,omitempty"`
	ObjectSizeLessThan    *int64       `json:"objectSizeLessThan,omitempty"`
	ObjectSizeGreaterThan *int64       `json:"objectSizeGreaterThan,omitempty"`
}

// StorageTag structure key/value pair representing an object tag to apply configuration
type StorageTag struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`
}

// LifecycleNoncurrentVersionExpiration - Specifies when noncurrent object versions expire.
// Upon expiration, server permanently deletes the noncurrent object versions.
// Set this lifecycle configuration action on a bucket that has versioning enabled
// (or suspended) to request server delete noncurrent object versions at a
// specific period in the object's lifetime.
type LifecycleNoncurrentVersionExpiration struct {
	NoncurrentDays          *int `json:"noncurrentDays,omitempty"`
	NewerNoncurrentVersions *int `json:"newerNoncurrentVersions,omitempty"`
}

// LifecycleNoncurrentVersionTransition sets this action to request server to
// transition noncurrent object versions to different set storage classes
// at a specific period in the object's lifetime.
type LifecycleNoncurrentVersionTransition struct {
	StorageClass            *string `json:"storageClass,omitempty"`
	NoncurrentDays          *int    `json:"noncurrentDays"`
	NewerNoncurrentVersions *int    `json:"newerNoncurrentVersions,omitempty"`
}

// ServerSideEncryptionConfiguration is the default encryption configuration structure.
type ServerSideEncryptionConfiguration struct {
	KmsMasterKeyID string `json:"kmsMasterKeyId,omitempty"`
	SSEAlgorithm   string `json:"sseAlgorithm,omitempty"`
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

	ObjectLock string `json:"objectLock"`
}

// StorageRetentionValidityUnit retention validity unit.
// @enum DAYS,YEARS
type StorageRetentionValidityUnit string

// StorageBucketVersioningConfiguration is the versioning configuration structure
type StorageBucketVersioningConfiguration struct {
	Enabled   bool    `json:"status"`
	MFADelete *string `json:"mfaDelete"`
	// MinIO extension - allows selective, prefix-level versioning exclusion.
	// Requires versioning to be enabled
	ExcludedPrefixes []string `json:"excludedPrefixes,omitempty"`
	ExcludeFolders   *bool    `json:"excludeFolders"`
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
	DeleteMarkerReplication   *DeleteMarkerReplication       `json:"deleteMarkerReplication"`
	DeleteReplication         *DeleteReplication             `json:"deleteReplication"`
	Destination               *StorageReplicationDestination `json:"destination"`
	Filter                    StorageReplicationFilter       `json:"filter"`
	SourceSelectionCriteria   *SourceSelectionCriteria       `json:"sourceSelectionCriteria"`
	ExistingObjectReplication *ExistingObjectReplication     `json:"existingObjectReplication,omitempty"`
}

// Destination the destination in ReplicationConfiguration.
type StorageReplicationDestination struct {
	Bucket       string  `json:"bucket"`
	StorageClass *string `json:"storageClass,omitempty"`
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
	ReplicaModifications *ReplicaModifications `json:"replicaModifications"`
}

// StorageReplicationFilter a filter for a replication configuration Rule.
type StorageReplicationFilter struct {
	Prefix *string                      `json:"rrefix,omitempty"`
	And    *StorageReplicationFilterAnd `json:"and,omitempty"`
	Tag    *StorageTag                  `json:"tag,omitempty"`
}

// StorageReplicationFilterAnd - a tag to combine a prefix and multiple tags for replication configuration rule.
type StorageReplicationFilterAnd struct {
	Prefix *string      `json:"rrefix,omitempty"`
	Tags   []StorageTag `json:"tag,omitempty"`
}
