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
func (j *GetStorageObjectOptions) FromValue(input map[string]any) error {
	var err error
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
	j.ListIncompleteUploadsOptions, err = utils.DecodeObject[ListIncompleteUploadsOptions](input)
	if err != nil {
		return err
	}
	j.StorageBucketArguments, err = utils.DecodeObject[StorageBucketArguments](input)
	if err != nil {
		return err
	}
	return nil
}

// FromValue decodes values from map
func (j *ListIncompleteUploadsOptions) FromValue(input map[string]any) error {
	var err error
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
	j.MaxResults, err = utils.GetIntDefault[int](input, "maxResults")
	if err != nil {
		return err
	}
	j.StartAfter, err = utils.GetStringDefault(input, "startAfter")
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
func (j *ListStorageObjectsArguments) FromValue(input map[string]any) error {
	var err error
	j.MaxResults, err = utils.GetNullableInt[int](input, "maxResults")
	if err != nil {
		return err
	}
	j.Recursive, err = utils.GetBooleanDefault(input, "recursive")
	if err != nil {
		return err
	}
	j.StartAfter, err = utils.GetNullableString(input, "startAfter")
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

// ToMap encodes the struct to a value map
func (j BucketAutoclass) ToMap() map[string]any {
	r := make(map[string]any)
	r["enabled"] = j.Enabled
	r["terminalStorageClass"] = j.TerminalStorageClass
	r["terminalStorageClassUpdateTime"] = j.TerminalStorageClassUpdateTime
	r["toggleTime"] = j.ToggleTime

	return r
}

// ToMap encodes the struct to a value map
func (j BucketCors) ToMap() map[string]any {
	r := make(map[string]any)
	r["maxAge"] = j.MaxAge
	r["methods"] = j.Methods
	r["origins"] = j.Origins
	r["responseHeaders"] = j.ResponseHeaders

	return r
}

// ToMap encodes the struct to a value map
func (j BucketHierarchicalNamespace) ToMap() map[string]any {
	r := make(map[string]any)
	r["enabled"] = j.Enabled

	return r
}

// ToMap encodes the struct to a value map
func (j BucketLogging) ToMap() map[string]any {
	r := make(map[string]any)
	r["logBucket"] = j.LogBucket
	r["logObjectPrefix"] = j.LogObjectPrefix

	return r
}

// ToMap encodes the struct to a value map
func (j BucketWebsite) ToMap() map[string]any {
	r := make(map[string]any)
	r["mainPageSuffix"] = j.MainPageSuffix
	r["notFoundPage"] = j.NotFoundPage

	return r
}

// ToMap encodes the struct to a value map
func (j CustomPlacementConfig) ToMap() map[string]any {
	r := make(map[string]any)
	r["DataLocations"] = j.DataLocations

	return r
}

// ToMap encodes the struct to a value map
func (j GetStorageObjectOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["headers"] = j.Headers
	r["partNumber"] = j.PartNumber
	r["requestParams"] = j.RequestParams
	r["versionId"] = j.VersionID

	return r
}

// ToMap encodes the struct to a value map
func (j ListIncompleteUploadsOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["prefix"] = j.Prefix
	r["recursive"] = j.Recursive

	return r
}

// ToMap encodes the struct to a value map
func (j ListStorageObjectsOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["maxResults"] = j.MaxResults
	r["prefix"] = j.Prefix
	r["recursive"] = j.Recursive
	r["startAfter"] = j.StartAfter

	return r
}

// ToMap encodes the struct to a value map
func (j MakeStorageBucketOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["name"] = j.Name
	r["objectLock"] = j.ObjectLock
	r["region"] = j.Region
	r["tags"] = j.Tags

	return r
}

// ToMap encodes the struct to a value map
func (j ObjectAbortIncompleteMultipartUpload) ToMap() map[string]any {
	r := make(map[string]any)
	r["daysAfterInitiation"] = j.DaysAfterInitiation

	return r
}

// ToMap encodes the struct to a value map
func (j ObjectLifecycleAllVersionsExpiration) ToMap() map[string]any {
	r := make(map[string]any)
	r["days"] = j.Days
	r["deleteMarker"] = j.DeleteMarker

	return r
}

// ToMap encodes the struct to a value map
func (j ObjectLifecycleConfiguration) ToMap() map[string]any {
	r := make(map[string]any)
	j_Rules := make([]any, len(j.Rules))
	for i, j_Rules_v := range j.Rules {
		j_Rules[i] = j_Rules_v
	}
	r["rules"] = j_Rules

	return r
}

// ToMap encodes the struct to a value map
func (j ObjectLifecycleDelMarkerExpiration) ToMap() map[string]any {
	r := make(map[string]any)
	r["days"] = j.Days

	return r
}

// ToMap encodes the struct to a value map
func (j ObjectLifecycleExpiration) ToMap() map[string]any {
	r := make(map[string]any)
	r["date"] = j.Date
	r["days"] = j.Days
	r["expiredObjectAllVersions"] = j.DeleteAll
	r["expiredObjectDeleteMarker"] = j.DeleteMarker

	return r
}

// ToMap encodes the struct to a value map
func (j ObjectLifecycleFilter) ToMap() map[string]any {
	r := make(map[string]any)
	r["matchesPrefix"] = j.MatchesPrefix
	r["matchesStorageClasses"] = j.MatchesStorageClasses
	r["matchesSuffix"] = j.MatchesSuffix
	r["objectSizeGreaterThan"] = j.ObjectSizeGreaterThan
	r["objectSizeLessThan"] = j.ObjectSizeLessThan
	r["tags"] = j.Tags

	return r
}

// ToMap encodes the struct to a value map
func (j ObjectLifecycleNoncurrentVersionExpiration) ToMap() map[string]any {
	r := make(map[string]any)
	r["newerNoncurrentVersions"] = j.NewerNoncurrentVersions
	r["noncurrentDays"] = j.NoncurrentDays

	return r
}

// ToMap encodes the struct to a value map
func (j ObjectLifecycleNoncurrentVersionTransition) ToMap() map[string]any {
	r := make(map[string]any)
	r["newerNoncurrentVersions"] = j.NewerNoncurrentVersions
	r["noncurrentDays"] = j.NoncurrentDays
	r["storageClass"] = j.StorageClass

	return r
}

// ToMap encodes the struct to a value map
func (j ObjectLifecycleRule) ToMap() map[string]any {
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
	r["enabled"] = j.Enabled
	if j.Expiration != nil {
		r["expiration"] = (*j.Expiration)
	}
	j_RuleFilter := make([]any, len(j.RuleFilter))
	for i, j_RuleFilter_v := range j.RuleFilter {
		j_RuleFilter[i] = j_RuleFilter_v
	}
	r["filter"] = j_RuleFilter
	r["id"] = j.ID
	if j.NoncurrentVersionExpiration != nil {
		r["noncurrentVersionExpiration"] = (*j.NoncurrentVersionExpiration)
	}
	if j.NoncurrentVersionTransition != nil {
		r["noncurrentVersionTransition"] = (*j.NoncurrentVersionTransition)
	}
	r["prefix"] = j.Prefix
	if j.Transition != nil {
		r["transition"] = (*j.Transition)
	}

	return r
}

// ToMap encodes the struct to a value map
func (j ObjectLifecycleTransition) ToMap() map[string]any {
	r := make(map[string]any)
	r["date"] = j.Date
	r["days"] = j.Days
	r["storageClass"] = j.StorageClass

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
	r["metadata"] = j.Metadata
	r["numThreads"] = j.NumThreads
	r["partSize"] = j.PartSize
	if j.Retention != nil {
		r["retention"] = (*j.Retention)
	}
	r["sendContentMd5"] = j.SendContentMd5
	r["storageClass"] = j.StorageClass
	r["tags"] = j.Tags
	r["websiteRedirectLocation"] = j.WebsiteRedirectLocation

	return r
}

// ToMap encodes the struct to a value map
func (j PutStorageObjectRetentionOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["governanceBypass"] = j.GovernanceBypass
	r["mode"] = j.Mode
	r["retainUntilDate"] = j.RetainUntilDate

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
	r["softDelete"] = j.SoftDelete
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
func (j ServerSideEncryptionConfiguration) ToMap() map[string]any {
	r := make(map[string]any)
	r["kmsMasterKeyId"] = j.KmsMasterKeyID
	r["sseAlgorithm"] = j.SSEAlgorithm

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
func (j SetStorageObjectRetentionOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["governanceBypass"] = j.GovernanceBypass
	r["mode"] = j.Mode
	r["retainUntilDate"] = j.RetainUntilDate

	return r
}

// ToMap encodes the struct to a value map
func (j StorageBucket) ToMap() map[string]any {
	r := make(map[string]any)
	if j.Autoclass != nil {
		r["autoclass"] = (*j.Autoclass)
	}
	j_CORS := make([]any, len(j.CORS))
	for i, j_CORS_v := range j.CORS {
		j_CORS[i] = j_CORS_v
	}
	r["cors"] = j_CORS
	r["creationTime"] = j.CreationTime
	if j.CustomPlacementConfig != nil {
		r["customPlacementConfig"] = (*j.CustomPlacementConfig)
	}
	r["defaultEventBasedHold"] = j.DefaultEventBasedHold
	if j.Encryption != nil {
		r["encryption"] = (*j.Encryption)
	}
	r["etag"] = j.Etag
	if j.HierarchicalNamespace != nil {
		r["hierarchicalNamespace"] = (*j.HierarchicalNamespace)
	}
	r["lastModified"] = j.LastModified
	if j.Lifecycle != nil {
		r["lifecycle"] = (*j.Lifecycle)
	}
	r["locationType"] = j.LocationType
	if j.Logging != nil {
		r["logging"] = (*j.Logging)
	}
	r["name"] = j.Name
	if j.ObjectLock != nil {
		r["objectLock"] = (*j.ObjectLock)
	}
	r["region"] = j.Region
	r["requesterPays"] = j.RequesterPays
	r["rpo"] = j.RPO
	if j.SoftDeletePolicy != nil {
		r["softDeletePolicy"] = (*j.SoftDeletePolicy)
	}
	r["storageClass"] = j.StorageClass
	r["tags"] = j.Tags
	if j.Versioning != nil {
		r["versioning"] = (*j.Versioning)
	}
	if j.Website != nil {
		r["website"] = (*j.Website)
	}

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
func (j StorageBucketListResults) ToMap() map[string]any {
	r := make(map[string]any)
	j_Buckets := make([]any, len(j.Buckets))
	for i, j_Buckets_v := range j.Buckets {
		j_Buckets[i] = j_Buckets_v
	}
	r["buckets"] = j_Buckets
	r["pageInfo"] = j.PageInfo

	return r
}

// ToMap encodes the struct to a value map
func (j StorageBucketVersioningConfiguration) ToMap() map[string]any {
	r := make(map[string]any)
	r["enabled"] = j.Enabled
	r["excludeFolders"] = j.ExcludeFolders
	r["excludedPrefixes"] = j.ExcludedPrefixes
	r["mfaDelete"] = j.MFADelete

	return r
}

// ToMap encodes the struct to a value map
func (j StorageCopyDestOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["bucket"] = j.Bucket
	r["legalHold"] = j.LegalHold
	r["metadata"] = j.Metadata
	r["mode"] = j.Mode
	r["object"] = j.Object
	r["retainUntilDate"] = j.RetainUntilDate
	r["size"] = j.Size
	r["tags"] = j.Tags

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
	r["accessTierChangeTime"] = j.AccessTierChangeTime
	r["accessTierInferred"] = j.AccessTierInferred
	r["acl"] = j.ACL
	r["archiveStatus"] = j.ArchiveStatus
	r["blobSequenceNumber"] = j.BlobSequenceNumber
	r["blobType"] = j.BlobType
	r["bucket"] = j.Bucket
	r["cacheControl"] = j.CacheControl
	r["clientId"] = j.ClientID
	r["contentDisposition"] = j.ContentDisposition
	r["contentEncoding"] = j.ContentEncoding
	r["contentLanguage"] = j.ContentLanguage
	r["contentMd5"] = j.ContentMD5
	r["contentType"] = j.ContentType
	if j.Copy != nil {
		r["copy"] = (*j.Copy)
	}
	r["creationTime"] = j.CreationTime
	r["customerProvidedKeySha256"] = j.CustomerProvidedKeySHA256
	r["deleted"] = j.Deleted
	r["deletedTime"] = j.DeletedTime
	r["destinationSnapshot"] = j.DestinationSnapshot
	r["etag"] = j.ETag
	r["expiration"] = j.Expiration
	r["expirationRuleId"] = j.ExpirationRuleID
	r["expires"] = j.Expires
	j_Grant := make([]any, len(j.Grant))
	for i, j_Grant_v := range j.Grant {
		j_Grant[i] = j_Grant_v
	}
	r["grant"] = j_Grant
	r["group"] = j.Group
	r["incrementalCopy"] = j.IncrementalCopy
	r["isLatest"] = j.IsLatest
	r["kmsKeyName"] = j.KMSKeyName
	r["lastAccessTime"] = j.LastAccessTime
	r["lastModified"] = j.LastModified
	r["leaseDuration"] = j.LeaseDuration
	r["leaseState"] = j.LeaseState
	r["leaseStatus"] = j.LeaseStatus
	r["legalHold"] = j.LegalHold
	r["mediaLink"] = j.MediaLink
	r["metadata"] = j.Metadata
	r["name"] = j.Name
	if j.Owner != nil {
		r["owner"] = (*j.Owner)
	}
	r["permissions"] = j.Permissions
	r["rawMetadata"] = j.RawMetadata
	r["rehydratePriority"] = j.RehydratePriority
	r["remainingRetentionDays"] = j.RemainingRetentionDays
	r["replicationReady"] = j.ReplicationReady
	r["replicationStatus"] = j.ReplicationStatus
	r["resourceType"] = j.ResourceType
	if j.Restore != nil {
		r["restore"] = (*j.Restore)
	}
	r["retentionMode"] = j.RetentionMode
	r["retentionUntilDate"] = j.RetentionUntilDate
	r["sealed"] = j.IsSealed
	r["serverEncrypted"] = j.ServerEncrypted
	r["size"] = j.Size
	r["storageClass"] = j.StorageClass
	r["tagCount"] = j.TagCount
	r["tags"] = j.Tags
	r["versionId"] = j.VersionID

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
func (j StorageObjectCopyInfo) ToMap() map[string]any {
	r := make(map[string]any)
	r["completionTime"] = j.CompletionTime
	r["id"] = j.ID
	r["progress"] = j.Progress
	r["source"] = j.Source
	r["status"] = j.Status
	r["statusDescription"] = j.StatusDescription

	return r
}

// ToMap encodes the struct to a value map
func (j StorageObjectListResults) ToMap() map[string]any {
	r := make(map[string]any)
	j_Objects := make([]any, len(j.Objects))
	for i, j_Objects_v := range j.Objects {
		j_Objects[i] = j_Objects_v
	}
	r["objects"] = j_Objects
	r["pageInfo"] = j.PageInfo

	return r
}

// ToMap encodes the struct to a value map
func (j StorageObjectLockConfig) ToMap() map[string]any {
	r := make(map[string]any)
	r = utils.MergeMap(r, j.SetStorageObjectLockConfig.ToMap())
	r["enabled"] = j.Enabled

	return r
}

// ToMap encodes the struct to a value map
func (j StorageObjectMultipartInfo) ToMap() map[string]any {
	r := make(map[string]any)
	r["initiated"] = j.Initiated
	r["name"] = j.Name
	r["size"] = j.Size
	r["storageClass"] = j.StorageClass
	r["uploadId"] = j.UploadID

	return r
}

// ToMap encodes the struct to a value map
func (j StorageObjectSoftDeletePolicy) ToMap() map[string]any {
	r := make(map[string]any)
	r["effectiveTime"] = j.EffectiveTime
	r["retentionDuration"] = j.RetentionDuration

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
func (j StoragePaginationInfo) ToMap() map[string]any {
	r := make(map[string]any)
	r["cursor"] = j.Cursor
	r["hasNextPage"] = j.HasNextPage

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
func (j StorageUploadInfo) ToMap() map[string]any {
	r := make(map[string]any)
	r = utils.MergeMap(r, j.StorageObjectChecksum.ToMap())
	r["bucket"] = j.Bucket
	r["clientId"] = j.ClientID
	r["contentMd5"] = j.ContentMD5
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

// ToMap encodes the struct to a value map
func (j UpdateStorageBucketOptions) ToMap() map[string]any {
	r := make(map[string]any)
	if j.Encryption != nil {
		r["encryption"] = (*j.Encryption)
	}
	if j.Lifecycle != nil {
		r["lifecycle"] = (*j.Lifecycle)
	}
	if j.ObjectLock != nil {
		r["objectLock"] = (*j.ObjectLock)
	}
	r["tags"] = j.Tags
	r["versioningEnabled"] = j.VersioningEnabled

	return r
}

// ToMap encodes the struct to a value map
func (j UpdateStorageObjectOptions) ToMap() map[string]any {
	r := make(map[string]any)
	r["legalHold"] = j.LegalHold
	r["metadata"] = j.Metadata
	if j.Retention != nil {
		r["retention"] = (*j.Retention)
	}
	r["tags"] = j.Tags
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
func (j GoogleStorageRPO) ScalarName() string {
	return "GoogleStorageRPO"
}

const (
	GoogleStorageRpoDefault    GoogleStorageRPO = "DEFAULT"
	GoogleStorageRpoAsyncTurbo GoogleStorageRPO = "ASYNC_TURBO"
)

var enumValues_GoogleStorageRpo = []GoogleStorageRPO{GoogleStorageRpoDefault, GoogleStorageRpoAsyncTurbo}

// ParseGoogleStorageRpo parses a GoogleStorageRPO enum from string
func ParseGoogleStorageRpo(input string) (GoogleStorageRPO, error) {
	result := GoogleStorageRPO(input)
	if !slices.Contains(enumValues_GoogleStorageRpo, result) {
		return GoogleStorageRPO(""), errors.New("failed to parse GoogleStorageRPO, expect one of [DEFAULT, ASYNC_TURBO]")
	}

	return result, nil
}

// IsValid checks if the value is invalid
func (j GoogleStorageRPO) IsValid() bool {
	return slices.Contains(enumValues_GoogleStorageRpo, j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *GoogleStorageRPO) UnmarshalJSON(b []byte) error {
	var rawValue string
	if err := json.Unmarshal(b, &rawValue); err != nil {
		return err
	}

	value, err := ParseGoogleStorageRpo(rawValue)
	if err != nil {
		return err
	}

	*j = value
	return nil
}

// FromValue decodes the scalar from an unknown value
func (s *GoogleStorageRPO) FromValue(value any) error {
	valueStr, err := utils.DecodeNullableString(value)
	if err != nil {
		return err
	}
	if valueStr == nil {
		return nil
	}
	result, err := ParseGoogleStorageRpo(*valueStr)
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
func (j StorageRetentionMode) ScalarName() string {
	return "StorageRetentionMode"
}

const (
	StorageRetentionModeLocked   StorageRetentionMode = "Locked"
	StorageRetentionModeUnlocked StorageRetentionMode = "Unlocked"
	StorageRetentionModeMutable  StorageRetentionMode = "Mutable"
	StorageRetentionModeDelete   StorageRetentionMode = "Delete"
)

var enumValues_StorageRetentionMode = []StorageRetentionMode{StorageRetentionModeLocked, StorageRetentionModeUnlocked, StorageRetentionModeMutable, StorageRetentionModeDelete}

// ParseStorageRetentionMode parses a StorageRetentionMode enum from string
func ParseStorageRetentionMode(input string) (StorageRetentionMode, error) {
	result := StorageRetentionMode(input)
	if !slices.Contains(enumValues_StorageRetentionMode, result) {
		return StorageRetentionMode(""), errors.New("failed to parse StorageRetentionMode, expect one of [Locked, Unlocked, Mutable, Delete]")
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
