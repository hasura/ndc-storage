// Code generated by github.com/hasura/ndc-sdk-go/cmd/hasura-ndc-go, DO NOT EDIT.
package common

import (
	"encoding/json"
	"errors"
	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"slices"
)

// FromValue decodes values from map
func (j *GetStorageObjectArguments) FromValue(input map[string]any) error {
	var err error
	j.GetStorageObjectOptions, err = utils.DecodeObject[GetStorageObjectOptions](input)
	if err != nil {
		return err
	}
	j.StorageBucketArguments, err = utils.DecodeObject[StorageBucketArguments](input)
	if err != nil {
		return err
	}
	j.Object, err = utils.GetString(input, "object")
	if err != nil {
		return err
	}
	j.Where, err = utils.DecodeObjectValueDefault[schema.Expression](input, "where")
	if err != nil {
		return err
	}
	return nil
}

// FromValue decodes values from map
func (j *GetStorageObjectLegalHoldOptions) FromValue(input map[string]any) error {
	var err error
	j.StorageBucketArguments, err = utils.DecodeObject[StorageBucketArguments](input)
	if err != nil {
		return err
	}
	j.Object, err = utils.GetString(input, "object")
	if err != nil {
		return err
	}
	j.VersionID, err = utils.GetStringDefault(input, "versionId")
	if err != nil {
		return err
	}
	return nil
}

// FromValue decodes values from map
func (j *GetStorageObjectOptions) FromValue(input map[string]any) error {
	var err error
	j.Checksum, err = utils.GetNullableBoolean(input, "checksum")
	if err != nil {
		return err
	}
	j.Headers, err = utils.DecodeObjectValueDefault[map[string]string](input, "headers")
	if err != nil {
		return err
	}
	j.PartNumber, err = utils.GetNullableInt[int](input, "partNumber")
	if err != nil {
		return err
	}
	j.RequestParams, err = utils.DecodeObjectValueDefault[map[string][]string](input, "requestParams")
	if err != nil {
		return err
	}
	j.VersionID, err = utils.GetNullableString(input, "versionId")
	if err != nil {
		return err
	}
	return nil
}

// FromValue decodes values from map
func (j *ListIncompleteUploadsArguments) FromValue(input map[string]any) error {
	var err error
	j.StorageBucketArguments, err = utils.DecodeObject[StorageBucketArguments](input)
	if err != nil {
		return err
	}
	j.Prefix, err = utils.GetString(input, "prefix")
	if err != nil {
		return err
	}
	j.Recursive, err = utils.GetBooleanDefault(input, "recursive")
	if err != nil {
		return err
	}
	return nil
}

// FromValue decodes values from map
func (j *ListStorageBucketArguments) FromValue(input map[string]any) error {
	var err error
	j.ClientID, err = utils.DecodeObjectValueDefault[StorageClientID](input, "clientId")
	if err != nil {
		return err
	}
	return nil
}

// FromValue decodes values from map
func (j *PresignedGetStorageObjectArguments) FromValue(input map[string]any) error {
	var err error
	j.PresignedGetStorageObjectOptions, err = utils.DecodeObject[PresignedGetStorageObjectOptions](input)
	if err != nil {
		return err
	}
	j.StorageBucketArguments, err = utils.DecodeObject[StorageBucketArguments](input)
	if err != nil {
		return err
	}
	j.Object, err = utils.GetString(input, "object")
	if err != nil {
		return err
	}
	j.Where, err = utils.DecodeObjectValueDefault[schema.Expression](input, "where")
	if err != nil {
		return err
	}
	return nil
}

// FromValue decodes values from map
func (j *PresignedGetStorageObjectOptions) FromValue(input map[string]any) error {
	var err error
	j.Expiry, err = utils.DecodeNullableObjectValue[scalar.Duration](input, "expiry")
	if err != nil {
		return err
	}
	j.RequestParams, err = utils.DecodeObjectValueDefault[map[string][]string](input, "requestParams")
	if err != nil {
		return err
	}
	return nil
}

// FromValue decodes values from map
func (j *PresignedPutStorageObjectArguments) FromValue(input map[string]any) error {
	var err error
	j.StorageBucketArguments, err = utils.DecodeObject[StorageBucketArguments](input)
	if err != nil {
		return err
	}
	j.Expiry, err = utils.DecodeNullableObjectValue[scalar.Duration](input, "expiry")
	if err != nil {
		return err
	}
	j.Object, err = utils.GetString(input, "object")
	if err != nil {
		return err
	}
	j.Where, err = utils.DecodeObjectValueDefault[schema.Expression](input, "where")
	if err != nil {
		return err
	}
	return nil
}

// FromValue decodes values from map
func (j *StorageBucketArguments) FromValue(input map[string]any) error {
	var err error
	j.Bucket, err = utils.GetStringDefault(input, "bucket")
	if err != nil {
		return err
	}
	j.ClientID, err = utils.DecodeNullableObjectValue[StorageClientID](input, "clientId")
	if err != nil {
		return err
	}
	return nil
}

// FromValue decodes values from map
func (j *StorageObjectAttributesOptions) FromValue(input map[string]any) error {
	var err error
	j.StorageBucketArguments, err = utils.DecodeObject[StorageBucketArguments](input)
	if err != nil {
		return err
	}
	j.MaxParts, err = utils.GetIntDefault[int](input, "maxParts")
	if err != nil {
		return err
	}
	j.Object, err = utils.GetString(input, "object")
	if err != nil {
		return err
	}
	j.PartNumberMarker, err = utils.GetIntDefault[int](input, "partNumberMarker")
	if err != nil {
		return err
	}
	j.VersionID, err = utils.GetStringDefault(input, "versionId")
	if err != nil {
		return err
	}
	return nil
}

// FromValue decodes values from map
func (j *StorageObjectTaggingOptions) FromValue(input map[string]any) error {
	var err error
	j.StorageBucketArguments, err = utils.DecodeObject[StorageBucketArguments](input)
	if err != nil {
		return err
	}
	j.Object, err = utils.GetString(input, "object")
	if err != nil {
		return err
	}
	j.VersionID, err = utils.GetStringDefault(input, "versionId")
	if err != nil {
		return err
	}
	return nil
}

// ToMap encodes the struct to a value map
func (j AbortIncompleteMultipartUpload) ToMap() map[string]any {
	r := make(map[string]any)
	r["daysAfterInitiation"] = j.DaysAfterInitiation

	return r
}

// ToMap encodes the struct to a value map
func (j BucketLifecycleConfiguration) ToMap() map[string]any {
	r := make(map[string]any)
	j_Rules := make([]any, len(j.Rules))
	for i, j_Rules_v := range j.Rules {
		j_Rules[i] = j_Rules_v
	}
	r["rules"] = j_Rules

	return r
}

// ToMap encodes the struct to a value map
func (j BucketLifecycleRule) ToMap() map[string]any {
	r := make(map[string]any)
	if j.AbortIncompleteMultipartUpload != nil {
		r["abortIncompleteMultipartUpload"] = (*j.AbortIncompleteMultipartUpload)
	}
	if j.AllVersionsExpiration != nil {
		r["allVersionsExpiration"] = (*j.AllVersionsExpiration)
	}
	if j.DelMarkerExpiration != nil {
		r["delMarkerExpiration"] = (*j.DelMarkerExpiration)
	}
	if j.Expiration != nil {
		r["expiration"] = (*j.Expiration)
	}
	if j.RuleFilter != nil {
		r["filter"] = (*j.RuleFilter)
	}
	r["id"] = j.ID
	if j.NoncurrentVersionExpiration != nil {
		r["noncurrentVersionExpiration"] = (*j.NoncurrentVersionExpiration)
	}
	if j.NoncurrentVersionTransition != nil {
		r["noncurrentVersionTransition"] = (*j.NoncurrentVersionTransition)
	}
	r["prefix"] = j.Prefix
	r["status"] = j.Status
	if j.Transition != nil {
		r["transition"] = (*j.Transition)
	}

	return r
}

// ToMap encodes the struct to a value map
func (j DeleteMarkerReplication) ToMap() map[string]any {
	r := make(map[string]any)
	r["status"] = j.Status

	return r
}

// ToMap encodes the struct to a value map
func (j DeleteReplication) ToMap() map[string]any {
	r := make(map[string]any)
	r["status"] = j.Status

	return r
}

// ToMap encodes the struct to a value map
func (j ExistingObjectReplication) ToMap() map[string]any {
	r := make(map[string]any)
	r["status"] = j.Status

	return r
}

// ToMap encodes the struct to a value map
func (j GetStorageObjectOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["checksum"] = j.Checksum
	r["headers"] = j.Headers
	r["partNumber"] = j.PartNumber
	r["requestParams"] = j.RequestParams
	r["versionId"] = j.VersionID

	return r
}

// ToMap encodes the struct to a value map
func (j LifecycleAllVersionsExpiration) ToMap() map[string]any {
	r := make(map[string]any)
	r["days"] = j.Days
	r["deleteMarker"] = j.DeleteMarker

	return r
}

// ToMap encodes the struct to a value map
func (j LifecycleDelMarkerExpiration) ToMap() map[string]any {
	r := make(map[string]any)
	r["days"] = j.Days

	return r
}

// ToMap encodes the struct to a value map
func (j LifecycleExpiration) ToMap() map[string]any {
	r := make(map[string]any)
	r["date"] = j.Date
	r["days"] = j.Days
	r["expiredObjectAllVersions"] = j.DeleteAll
	r["expiredObjectDeleteMarker"] = j.DeleteMarker

	return r
}

// ToMap encodes the struct to a value map
func (j LifecycleFilter) ToMap() map[string]any {
	r := make(map[string]any)
	if j.And != nil {
		r["and"] = (*j.And)
	}
	r["objectSizeGreaterThan"] = j.ObjectSizeGreaterThan
	r["objectSizeLessThan"] = j.ObjectSizeLessThan
	r["prefix"] = j.Prefix
	if j.Tag != nil {
		r["tag"] = (*j.Tag)
	}

	return r
}

// ToMap encodes the struct to a value map
func (j LifecycleFilterAnd) ToMap() map[string]any {
	r := make(map[string]any)
	r["objectSizeGreaterThan"] = j.ObjectSizeGreaterThan
	r["objectSizeLessThan"] = j.ObjectSizeLessThan
	r["prefix"] = j.Prefix
	j_Tags := make([]any, len(j.Tags))
	for i, j_Tags_v := range j.Tags {
		j_Tags[i] = j_Tags_v
	}
	r["tags"] = j_Tags

	return r
}

// ToMap encodes the struct to a value map
func (j LifecycleNoncurrentVersionExpiration) ToMap() map[string]any {
	r := make(map[string]any)
	r["newerNoncurrentVersions"] = j.NewerNoncurrentVersions
	r["noncurrentDays"] = j.NoncurrentDays

	return r
}

// ToMap encodes the struct to a value map
func (j LifecycleNoncurrentVersionTransition) ToMap() map[string]any {
	r := make(map[string]any)
	r["newerNoncurrentVersions"] = j.NewerNoncurrentVersions
	r["noncurrentDays"] = j.NoncurrentDays
	r["storageClass"] = j.StorageClass

	return r
}

// ToMap encodes the struct to a value map
func (j LifecycleTransition) ToMap() map[string]any {
	r := make(map[string]any)
	r["date"] = j.Date
	r["days"] = j.Days
	r["storageClass"] = j.StorageClass

	return r
}

// ToMap encodes the struct to a value map
func (j ListStorageObjectsOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["maxKeys"] = j.MaxKeys
	r["prefix"] = j.Prefix
	r["recursive"] = j.Recursive
	r["startAfter"] = j.StartAfter
	r["withMetadata"] = j.WithMetadata
	r["withVersions"] = j.WithVersions

	return r
}

// ToMap encodes the struct to a value map
func (j NotificationCommonConfig) ToMap() map[string]any {
	r := make(map[string]any)
	r["arn"] = j.Arn
	r["event"] = j.Events
	if j.Filter != nil {
		r["filter"] = (*j.Filter)
	}
	r["id"] = j.ID

	return r
}

// ToMap encodes the struct to a value map
func (j NotificationConfig) ToMap() map[string]any {
	r := make(map[string]any)
	j_LambdaConfigs := make([]any, len(j.LambdaConfigs))
	for i, j_LambdaConfigs_v := range j.LambdaConfigs {
		j_LambdaConfigs[i] = j_LambdaConfigs_v
	}
	r["cloudFunctionConfigurations"] = j_LambdaConfigs
	j_QueueConfigs := make([]any, len(j.QueueConfigs))
	for i, j_QueueConfigs_v := range j.QueueConfigs {
		j_QueueConfigs[i] = j_QueueConfigs_v
	}
	r["queueConfigurations"] = j_QueueConfigs
	j_TopicConfigs := make([]any, len(j.TopicConfigs))
	for i, j_TopicConfigs_v := range j.TopicConfigs {
		j_TopicConfigs[i] = j_TopicConfigs_v
	}
	r["topicConfigurations"] = j_TopicConfigs

	return r
}

// ToMap encodes the struct to a value map
func (j NotificationFilter) ToMap() map[string]any {
	r := make(map[string]any)
	if j.S3Key != nil {
		r["s3Key"] = (*j.S3Key)
	}

	return r
}

// ToMap encodes the struct to a value map
func (j NotificationFilterRule) ToMap() map[string]any {
	r := make(map[string]any)
	r["name"] = j.Name
	r["value"] = j.Value

	return r
}

// ToMap encodes the struct to a value map
func (j NotificationLambdaConfig) ToMap() map[string]any {
	r := make(map[string]any)
	r = utils.MergeMap(r, j.NotificationCommonConfig.ToMap())
	r["cloudFunction"] = j.Lambda

	return r
}

// ToMap encodes the struct to a value map
func (j NotificationQueueConfig) ToMap() map[string]any {
	r := make(map[string]any)
	r = utils.MergeMap(r, j.NotificationCommonConfig.ToMap())
	r["queue"] = j.Queue

	return r
}

// ToMap encodes the struct to a value map
func (j NotificationS3Key) ToMap() map[string]any {
	r := make(map[string]any)
	j_FilterRules := make([]any, len(j.FilterRules))
	for i, j_FilterRules_v := range j.FilterRules {
		j_FilterRules[i] = j_FilterRules_v
	}
	r["filterRule"] = j_FilterRules

	return r
}

// ToMap encodes the struct to a value map
func (j NotificationTopicConfig) ToMap() map[string]any {
	r := make(map[string]any)
	r = utils.MergeMap(r, j.NotificationCommonConfig.ToMap())
	r["topic"] = j.Topic

	return r
}

// ToMap encodes the struct to a value map
func (j PresignedGetStorageObjectOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["expiry"] = j.Expiry
	r["requestParams"] = j.RequestParams

	return r
}

// ToMap encodes the struct to a value map
func (j PresignedURLResponse) ToMap() map[string]any {
	r := make(map[string]any)
	r["expiredAt"] = j.ExpiredAt
	r["url"] = j.URL

	return r
}

// ToMap encodes the struct to a value map
func (j PutStorageObjectOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["autoChecksum"] = j.AutoChecksum
	r["cacheControl"] = j.CacheControl
	r["checksum"] = j.Checksum
	r["concurrentStreamParts"] = j.ConcurrentStreamParts
	r["contentDisposition"] = j.ContentDisposition
	r["contentEncoding"] = j.ContentEncoding
	r["contentLanguage"] = j.ContentLanguage
	r["contentType"] = j.ContentType
	r["disableContentSha256"] = j.DisableContentSha256
	r["disableMultipart"] = j.DisableMultipart
	r["expires"] = j.Expires
	r["legalHold"] = j.LegalHold
	r["mode"] = j.Mode
	r["numThreads"] = j.NumThreads
	r["partSize"] = j.PartSize
	r["retainUntilDate"] = j.RetainUntilDate
	r["sendContentMd5"] = j.SendContentMd5
	r["storageClass"] = j.StorageClass
	r["userMetadata"] = j.UserMetadata
	r["userTags"] = j.UserTags
	r["websiteRedirectLocation"] = j.WebsiteRedirectLocation

	return r
}

// ToMap encodes the struct to a value map
func (j RemoveStorageObjectError) ToMap() map[string]any {
	r := make(map[string]any)
	r["error"] = j.Error
	r["objectName"] = j.ObjectName
	r["versionId"] = j.VersionID

	return r
}

// ToMap encodes the struct to a value map
func (j RemoveStorageObjectOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["forceDelete"] = j.ForceDelete
	r["governanceBypass"] = j.GovernanceBypass
	r["versionId"] = j.VersionID

	return r
}

// ToMap encodes the struct to a value map
func (j RemoveStorageObjectsOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r = utils.MergeMap(r, j.ListStorageObjectsOptions.ToMap())
	r["governanceBypass"] = j.GovernanceBypass

	return r
}

// ToMap encodes the struct to a value map
func (j ReplicaModifications) ToMap() map[string]any {
	r := make(map[string]any)
	r["status"] = j.Status

	return r
}

// ToMap encodes the struct to a value map
func (j ServerSideEncryptionConfiguration) ToMap() map[string]any {
	r := make(map[string]any)
	j_Rules := make([]any, len(j.Rules))
	for i, j_Rules_v := range j.Rules {
		j_Rules[i] = j_Rules_v
	}
	r["rules"] = j_Rules

	return r
}

// ToMap encodes the struct to a value map
func (j ServerSideEncryptionRule) ToMap() map[string]any {
	r := make(map[string]any)
	r["apply"] = j.Apply

	return r
}

// ToMap encodes the struct to a value map
func (j SetStorageObjectLockConfig) ToMap() map[string]any {
	r := make(map[string]any)
	r["mode"] = j.Mode
	r["unit"] = j.Unit
	r["validity"] = j.Validity

	return r
}

// ToMap encodes the struct to a value map
func (j SourceSelectionCriteria) ToMap() map[string]any {
	r := make(map[string]any)
	if j.ReplicaModifications != nil {
		r["replicaModifications"] = (*j.ReplicaModifications)
	}

	return r
}

// ToMap encodes the struct to a value map
func (j StorageApplySSEByDefault) ToMap() map[string]any {
	r := make(map[string]any)
	r["kmsMasterKeyId"] = j.KmsMasterKeyID
	r["sseAlgorithm"] = j.SSEAlgorithm

	return r
}

// ToMap encodes the struct to a value map
func (j StorageBucketArguments) ToMap() map[string]any {
	r := make(map[string]any)
	r["bucket"] = j.Bucket
	r["clientId"] = j.ClientID

	return r
}

// ToMap encodes the struct to a value map
func (j StorageBucketInfo) ToMap() map[string]any {
	r := make(map[string]any)
	r["creationDate"] = j.CreationDate
	r["name"] = j.Name

	return r
}

// ToMap encodes the struct to a value map
func (j StorageBucketVersioningConfiguration) ToMap() map[string]any {
	r := make(map[string]any)
	r["excludeFolders"] = j.ExcludeFolders
	r["excludedPrefixes"] = j.ExcludedPrefixes
	r["mfaDelete"] = j.MFADelete
	r["status"] = j.Status

	return r
}

// ToMap encodes the struct to a value map
func (j StorageCopyDestOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["bucket"] = j.Bucket
	r["legalHold"] = j.LegalHold
	r["mode"] = j.Mode
	r["object"] = j.Object
	r["replaceMetadata"] = j.ReplaceMetadata
	r["replaceTags"] = j.ReplaceTags
	r["retainUntilDate"] = j.RetainUntilDate
	r["size"] = j.Size
	r["userMetadata"] = j.UserMetadata
	r["userTags"] = j.UserTags

	return r
}

// ToMap encodes the struct to a value map
func (j StorageCopySrcOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["bucket"] = j.Bucket
	r["end"] = j.End
	r["matchETag"] = j.MatchETag
	r["matchModifiedSince"] = j.MatchModifiedSince
	r["matchRange"] = j.MatchRange
	r["matchUnmodifiedSince"] = j.MatchUnmodifiedSince
	r["noMatchETag"] = j.NoMatchETag
	r["object"] = j.Object
	r["start"] = j.Start
	r["versionId"] = j.VersionID

	return r
}

// ToMap encodes the struct to a value map
func (j StorageGrant) ToMap() map[string]any {
	r := make(map[string]any)
	if j.Grantee != nil {
		r["grantee"] = (*j.Grantee)
	}
	r["permission"] = j.Permission

	return r
}

// ToMap encodes the struct to a value map
func (j StorageGrantee) ToMap() map[string]any {
	r := make(map[string]any)
	r["displayName"] = j.DisplayName
	r["id"] = j.ID
	r["uri"] = j.URI

	return r
}

// ToMap encodes the struct to a value map
func (j StorageObject) ToMap() map[string]any {
	r := make(map[string]any)
	r = utils.MergeMap(r, j.StorageObjectChecksum.ToMap())
	r["bucket"] = j.Bucket
	r["clientId"] = j.ClientID
	r["contentType"] = j.ContentType
	r["etag"] = j.ETag
	r["expiration"] = j.Expiration
	r["expirationRuleId"] = j.ExpirationRuleID
	r["expires"] = j.Expires
	j_Grant := make([]any, len(j.Grant))
	for i, j_Grant_v := range j.Grant {
		j_Grant[i] = j_Grant_v
	}
	r["grant"] = j_Grant
	r["isDeleteMarker"] = j.IsDeleteMarker
	r["isLatest"] = j.IsLatest
	r["lastModified"] = j.LastModified
	r["metadata"] = j.Metadata
	r["name"] = j.Name
	if j.Owner != nil {
		r["owner"] = (*j.Owner)
	}
	r["replicationReady"] = j.ReplicationReady
	r["replicationStatus"] = j.ReplicationStatus
	if j.Restore != nil {
		r["restore"] = (*j.Restore)
	}
	r["size"] = j.Size
	r["storageClass"] = j.StorageClass
	r["userMetadata"] = j.UserMetadata
	r["userTagCount"] = j.UserTagCount
	r["userTags"] = j.UserTags
	r["versionId"] = j.VersionID

	return r
}

// ToMap encodes the struct to a value map
func (j StorageObjectAttributePart) ToMap() map[string]any {
	r := make(map[string]any)
	r = utils.MergeMap(r, j.StorageObjectChecksum.ToMap())
	r["partNumber"] = j.PartNumber
	r["size"] = j.Size

	return r
}

// ToMap encodes the struct to a value map
func (j StorageObjectAttributes) ToMap() map[string]any {
	r := make(map[string]any)
	r = utils.MergeMap(r, j.StorageObjectAttributesResponse.ToMap())
	r["lastModified"] = j.LastModified
	r["versionId"] = j.VersionID

	return r
}

// ToMap encodes the struct to a value map
func (j StorageObjectAttributesResponse) ToMap() map[string]any {
	r := make(map[string]any)
	r["checksum"] = j.Checksum
	r["etag"] = j.ETag
	r["objectParts"] = j.ObjectParts
	r["objectSize"] = j.ObjectSize
	r["storageClass"] = j.StorageClass

	return r
}

// ToMap encodes the struct to a value map
func (j StorageObjectChecksum) ToMap() map[string]any {
	r := make(map[string]any)
	r["checksumCrc32"] = j.ChecksumCRC32
	r["checksumCrc32C"] = j.ChecksumCRC32C
	r["checksumCrc64Nvme"] = j.ChecksumCRC64NVME
	r["checksumSha1"] = j.ChecksumSHA1
	r["checksumSha256"] = j.ChecksumSHA256

	return r
}

// ToMap encodes the struct to a value map
func (j StorageObjectLockConfig) ToMap() map[string]any {
	r := make(map[string]any)
	r = utils.MergeMap(r, j.SetStorageObjectLockConfig.ToMap())
	r["objectLock"] = j.ObjectLock

	return r
}

// ToMap encodes the struct to a value map
func (j StorageObjectMultipartInfo) ToMap() map[string]any {
	r := make(map[string]any)
	r["initiated"] = j.Initiated
	r["key"] = j.Key
	r["size"] = j.Size
	r["storageClass"] = j.StorageClass
	r["uploadId"] = j.UploadID

	return r
}

// ToMap encodes the struct to a value map
func (j StorageObjectParts) ToMap() map[string]any {
	r := make(map[string]any)
	r["isTruncated"] = j.IsTruncated
	r["maxParts"] = j.MaxParts
	r["nextPartNumberMarker"] = j.NextPartNumberMarker
	r["partNumberMarker"] = j.PartNumberMarker
	j_Parts := make([]any, len(j.Parts))
	for i, j_Parts_v := range j.Parts {
		if j_Parts_v != nil {
			j_Parts[i] = (*j_Parts_v)
		}
	}
	r["parts"] = j_Parts
	r["partsCount"] = j.PartsCount

	return r
}

// ToMap encodes the struct to a value map
func (j StorageOwner) ToMap() map[string]any {
	r := make(map[string]any)
	r["id"] = j.ID
	r["name"] = j.DisplayName

	return r
}

// ToMap encodes the struct to a value map
func (j StorageReplicationConfig) ToMap() map[string]any {
	r := make(map[string]any)
	r["role"] = j.Role
	j_Rules := make([]any, len(j.Rules))
	for i, j_Rules_v := range j.Rules {
		j_Rules[i] = j_Rules_v
	}
	r["rules"] = j_Rules

	return r
}

// ToMap encodes the struct to a value map
func (j StorageReplicationDestination) ToMap() map[string]any {
	r := make(map[string]any)
	r["bucket"] = j.Bucket
	r["storageClass"] = j.StorageClass

	return r
}

// ToMap encodes the struct to a value map
func (j StorageReplicationFilter) ToMap() map[string]any {
	r := make(map[string]any)
	if j.And != nil {
		r["and"] = (*j.And)
	}
	r["rrefix"] = j.Prefix
	if j.Tag != nil {
		r["tag"] = (*j.Tag)
	}

	return r
}

// ToMap encodes the struct to a value map
func (j StorageReplicationFilterAnd) ToMap() map[string]any {
	r := make(map[string]any)
	r["rrefix"] = j.Prefix
	j_Tags := make([]any, len(j.Tags))
	for i, j_Tags_v := range j.Tags {
		j_Tags[i] = j_Tags_v
	}
	r["tag"] = j_Tags

	return r
}

// ToMap encodes the struct to a value map
func (j StorageReplicationRule) ToMap() map[string]any {
	r := make(map[string]any)
	if j.DeleteMarkerReplication != nil {
		r["deleteMarkerReplication"] = (*j.DeleteMarkerReplication)
	}
	if j.DeleteReplication != nil {
		r["deleteReplication"] = (*j.DeleteReplication)
	}
	if j.Destination != nil {
		r["destination"] = (*j.Destination)
	}
	if j.ExistingObjectReplication != nil {
		r["existingObjectReplication"] = (*j.ExistingObjectReplication)
	}
	r["filter"] = j.Filter
	r["id"] = j.ID
	r["priority"] = j.Priority
	if j.SourceSelectionCriteria != nil {
		r["sourceSelectionCriteria"] = (*j.SourceSelectionCriteria)
	}
	r["status"] = j.Status

	return r
}

// ToMap encodes the struct to a value map
func (j StorageRestoreInfo) ToMap() map[string]any {
	r := make(map[string]any)
	r["expiryTime"] = j.ExpiryTime
	r["ongoingRestore"] = j.OngoingRestore

	return r
}

// ToMap encodes the struct to a value map
func (j StorageTag) ToMap() map[string]any {
	r := make(map[string]any)
	r["key"] = j.Key
	r["value"] = j.Value

	return r
}

// ToMap encodes the struct to a value map
func (j StorageUploadInfo) ToMap() map[string]any {
	r := make(map[string]any)
	r = utils.MergeMap(r, j.StorageObjectChecksum.ToMap())
	r["bucket"] = j.Bucket
	r["clientId"] = j.ClientID
	r["etag"] = j.ETag
	r["expiration"] = j.Expiration
	r["expirationRuleId"] = j.ExpirationRuleID
	r["lastModified"] = j.LastModified
	r["location"] = j.Location
	r["name"] = j.Name
	r["size"] = j.Size
	r["versionId"] = j.VersionID

	return r
}

// ScalarName get the schema name of the scalar
func (j ChecksumType) ScalarName() string {
	return "ChecksumType"
}

const (
	ChecksumTypeSha256           ChecksumType = "SHA256"
	ChecksumTypeSha1             ChecksumType = "SHA1"
	ChecksumTypeCrc32            ChecksumType = "CRC32"
	ChecksumTypeCrc32C           ChecksumType = "CRC32C"
	ChecksumTypeCrc64Nvme        ChecksumType = "CRC64NVME"
	ChecksumTypeFullObjectCrc32  ChecksumType = "FullObjectCRC32"
	ChecksumTypeFullObjectCrc32C ChecksumType = "FullObjectCRC32C"
	ChecksumTypeNone             ChecksumType = "None"
)

var enumValues_ChecksumType = []ChecksumType{ChecksumTypeSha256, ChecksumTypeSha1, ChecksumTypeCrc32, ChecksumTypeCrc32C, ChecksumTypeCrc64Nvme, ChecksumTypeFullObjectCrc32, ChecksumTypeFullObjectCrc32C, ChecksumTypeNone}

// ParseChecksumType parses a ChecksumType enum from string
func ParseChecksumType(input string) (ChecksumType, error) {
	result := ChecksumType(input)
	if !slices.Contains(enumValues_ChecksumType, result) {
		return ChecksumType(""), errors.New("failed to parse ChecksumType, expect one of [SHA256, SHA1, CRC32, CRC32C, CRC64NVME, FullObjectCRC32, FullObjectCRC32C, None]")
	}

	return result, nil
}

// IsValid checks if the value is invalid
func (j ChecksumType) IsValid() bool {
	return slices.Contains(enumValues_ChecksumType, j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ChecksumType) UnmarshalJSON(b []byte) error {
	var rawValue string
	if err := json.Unmarshal(b, &rawValue); err != nil {
		return err
	}

	value, err := ParseChecksumType(rawValue)
	if err != nil {
		return err
	}

	*j = value
	return nil
}

// FromValue decodes the scalar from an unknown value
func (s *ChecksumType) FromValue(value any) error {
	valueStr, err := utils.DecodeNullableString(value)
	if err != nil {
		return err
	}
	if valueStr == nil {
		return nil
	}
	result, err := ParseChecksumType(*valueStr)
	if err != nil {
		return err
	}

	*s = result
	return nil
}

// ScalarName get the schema name of the scalar
func (j StorageClientID) ScalarName() string {
	return "StorageClientID"
}

// ScalarName get the schema name of the scalar
func (j StorageLegalHoldStatus) ScalarName() string {
	return "StorageLegalHoldStatus"
}

const (
	StorageLegalHoldStatusOn  StorageLegalHoldStatus = "ON"
	StorageLegalHoldStatusOff StorageLegalHoldStatus = "OFF"
)

var enumValues_StorageLegalHoldStatus = []StorageLegalHoldStatus{StorageLegalHoldStatusOn, StorageLegalHoldStatusOff}

// ParseStorageLegalHoldStatus parses a StorageLegalHoldStatus enum from string
func ParseStorageLegalHoldStatus(input string) (StorageLegalHoldStatus, error) {
	result := StorageLegalHoldStatus(input)
	if !slices.Contains(enumValues_StorageLegalHoldStatus, result) {
		return StorageLegalHoldStatus(""), errors.New("failed to parse StorageLegalHoldStatus, expect one of [ON, OFF]")
	}

	return result, nil
}

// IsValid checks if the value is invalid
func (j StorageLegalHoldStatus) IsValid() bool {
	return slices.Contains(enumValues_StorageLegalHoldStatus, j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *StorageLegalHoldStatus) UnmarshalJSON(b []byte) error {
	var rawValue string
	if err := json.Unmarshal(b, &rawValue); err != nil {
		return err
	}

	value, err := ParseStorageLegalHoldStatus(rawValue)
	if err != nil {
		return err
	}

	*j = value
	return nil
}

// FromValue decodes the scalar from an unknown value
func (s *StorageLegalHoldStatus) FromValue(value any) error {
	valueStr, err := utils.DecodeNullableString(value)
	if err != nil {
		return err
	}
	if valueStr == nil {
		return nil
	}
	result, err := ParseStorageLegalHoldStatus(*valueStr)
	if err != nil {
		return err
	}

	*s = result
	return nil
}

// ScalarName get the schema name of the scalar
func (j StorageObjectReplicationStatus) ScalarName() string {
	return "StorageObjectReplicationStatus"
}

const (
	StorageObjectReplicationStatusCompleted StorageObjectReplicationStatus = "COMPLETED"
	StorageObjectReplicationStatusPending   StorageObjectReplicationStatus = "PENDING"
	StorageObjectReplicationStatusFailed    StorageObjectReplicationStatus = "FAILED"
	StorageObjectReplicationStatusReplica   StorageObjectReplicationStatus = "REPLICA"
)

var enumValues_StorageObjectReplicationStatus = []StorageObjectReplicationStatus{StorageObjectReplicationStatusCompleted, StorageObjectReplicationStatusPending, StorageObjectReplicationStatusFailed, StorageObjectReplicationStatusReplica}

// ParseStorageObjectReplicationStatus parses a StorageObjectReplicationStatus enum from string
func ParseStorageObjectReplicationStatus(input string) (StorageObjectReplicationStatus, error) {
	result := StorageObjectReplicationStatus(input)
	if !slices.Contains(enumValues_StorageObjectReplicationStatus, result) {
		return StorageObjectReplicationStatus(""), errors.New("failed to parse StorageObjectReplicationStatus, expect one of [COMPLETED, PENDING, FAILED, REPLICA]")
	}

	return result, nil
}

// IsValid checks if the value is invalid
func (j StorageObjectReplicationStatus) IsValid() bool {
	return slices.Contains(enumValues_StorageObjectReplicationStatus, j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *StorageObjectReplicationStatus) UnmarshalJSON(b []byte) error {
	var rawValue string
	if err := json.Unmarshal(b, &rawValue); err != nil {
		return err
	}

	value, err := ParseStorageObjectReplicationStatus(rawValue)
	if err != nil {
		return err
	}

	*j = value
	return nil
}

// FromValue decodes the scalar from an unknown value
func (s *StorageObjectReplicationStatus) FromValue(value any) error {
	valueStr, err := utils.DecodeNullableString(value)
	if err != nil {
		return err
	}
	if valueStr == nil {
		return nil
	}
	result, err := ParseStorageObjectReplicationStatus(*valueStr)
	if err != nil {
		return err
	}

	*s = result
	return nil
}

// ScalarName get the schema name of the scalar
func (j StorageReplicationRuleStatus) ScalarName() string {
	return "StorageReplicationRuleStatus"
}

const (
	StorageReplicationRuleStatusEnabled  StorageReplicationRuleStatus = "Enabled"
	StorageReplicationRuleStatusDisabled StorageReplicationRuleStatus = "Disabled"
)

var enumValues_StorageReplicationRuleStatus = []StorageReplicationRuleStatus{StorageReplicationRuleStatusEnabled, StorageReplicationRuleStatusDisabled}

// ParseStorageReplicationRuleStatus parses a StorageReplicationRuleStatus enum from string
func ParseStorageReplicationRuleStatus(input string) (StorageReplicationRuleStatus, error) {
	result := StorageReplicationRuleStatus(input)
	if !slices.Contains(enumValues_StorageReplicationRuleStatus, result) {
		return StorageReplicationRuleStatus(""), errors.New("failed to parse StorageReplicationRuleStatus, expect one of [Enabled, Disabled]")
	}

	return result, nil
}

// IsValid checks if the value is invalid
func (j StorageReplicationRuleStatus) IsValid() bool {
	return slices.Contains(enumValues_StorageReplicationRuleStatus, j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *StorageReplicationRuleStatus) UnmarshalJSON(b []byte) error {
	var rawValue string
	if err := json.Unmarshal(b, &rawValue); err != nil {
		return err
	}

	value, err := ParseStorageReplicationRuleStatus(rawValue)
	if err != nil {
		return err
	}

	*j = value
	return nil
}

// FromValue decodes the scalar from an unknown value
func (s *StorageReplicationRuleStatus) FromValue(value any) error {
	valueStr, err := utils.DecodeNullableString(value)
	if err != nil {
		return err
	}
	if valueStr == nil {
		return nil
	}
	result, err := ParseStorageReplicationRuleStatus(*valueStr)
	if err != nil {
		return err
	}

	*s = result
	return nil
}

// ScalarName get the schema name of the scalar
func (j StorageRetentionMode) ScalarName() string {
	return "StorageRetentionMode"
}

const (
	StorageRetentionModeGovernance StorageRetentionMode = "GOVERNANCE"
	StorageRetentionModeCompliance StorageRetentionMode = "COMPLIANCE"
)

var enumValues_StorageRetentionMode = []StorageRetentionMode{StorageRetentionModeGovernance, StorageRetentionModeCompliance}

// ParseStorageRetentionMode parses a StorageRetentionMode enum from string
func ParseStorageRetentionMode(input string) (StorageRetentionMode, error) {
	result := StorageRetentionMode(input)
	if !slices.Contains(enumValues_StorageRetentionMode, result) {
		return StorageRetentionMode(""), errors.New("failed to parse StorageRetentionMode, expect one of [GOVERNANCE, COMPLIANCE]")
	}

	return result, nil
}

// IsValid checks if the value is invalid
func (j StorageRetentionMode) IsValid() bool {
	return slices.Contains(enumValues_StorageRetentionMode, j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *StorageRetentionMode) UnmarshalJSON(b []byte) error {
	var rawValue string
	if err := json.Unmarshal(b, &rawValue); err != nil {
		return err
	}

	value, err := ParseStorageRetentionMode(rawValue)
	if err != nil {
		return err
	}

	*j = value
	return nil
}

// FromValue decodes the scalar from an unknown value
func (s *StorageRetentionMode) FromValue(value any) error {
	valueStr, err := utils.DecodeNullableString(value)
	if err != nil {
		return err
	}
	if valueStr == nil {
		return nil
	}
	result, err := ParseStorageRetentionMode(*valueStr)
	if err != nil {
		return err
	}

	*s = result
	return nil
}

// ScalarName get the schema name of the scalar
func (j StorageRetentionValidityUnit) ScalarName() string {
	return "StorageRetentionValidityUnit"
}

const (
	StorageRetentionValidityUnitDays  StorageRetentionValidityUnit = "DAYS"
	StorageRetentionValidityUnitYears StorageRetentionValidityUnit = "YEARS"
)

var enumValues_StorageRetentionValidityUnit = []StorageRetentionValidityUnit{StorageRetentionValidityUnitDays, StorageRetentionValidityUnitYears}

// ParseStorageRetentionValidityUnit parses a StorageRetentionValidityUnit enum from string
func ParseStorageRetentionValidityUnit(input string) (StorageRetentionValidityUnit, error) {
	result := StorageRetentionValidityUnit(input)
	if !slices.Contains(enumValues_StorageRetentionValidityUnit, result) {
		return StorageRetentionValidityUnit(""), errors.New("failed to parse StorageRetentionValidityUnit, expect one of [DAYS, YEARS]")
	}

	return result, nil
}

// IsValid checks if the value is invalid
func (j StorageRetentionValidityUnit) IsValid() bool {
	return slices.Contains(enumValues_StorageRetentionValidityUnit, j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *StorageRetentionValidityUnit) UnmarshalJSON(b []byte) error {
	var rawValue string
	if err := json.Unmarshal(b, &rawValue); err != nil {
		return err
	}

	value, err := ParseStorageRetentionValidityUnit(rawValue)
	if err != nil {
		return err
	}

	*j = value
	return nil
}

// FromValue decodes the scalar from an unknown value
func (s *StorageRetentionValidityUnit) FromValue(value any) error {
	valueStr, err := utils.DecodeNullableString(value)
	if err != nil {
		return err
	}
	if valueStr == nil {
		return nil
	}
	result, err := ParseStorageRetentionValidityUnit(*valueStr)
	if err != nil {
		return err
	}

	*s = result
	return nil
}
