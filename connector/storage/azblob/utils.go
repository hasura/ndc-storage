package azblob

import (
	"encoding/base64"
	"errors"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

var tracer = connector.NewTracer("connector/storage/azblob")

var errNotSupported = schema.NotSupportedError("Azure Blob Storage doesn't support this method", nil)

func serializeObjectInfo(item *container.BlobItem) common.StorageObject { //nolint:funlen,cyclop
	object := common.StorageObject{
		Metadata:  make(map[string]string),
		IsLatest:  item.IsCurrentVersion,
		Deleted:   item.Deleted,
		VersionID: item.VersionID,
	}

	if item.Name != nil {
		object.Name = *item.Name
	}

	if item.BlobTags != nil && len(item.BlobTags.BlobTagSet) > 0 {
		object.Tags = make(map[string]string)

		for _, bt := range item.BlobTags.BlobTagSet {
			if bt.Key == nil || bt.Value == nil {
				continue
			}

			object.Tags[*bt.Key] = *bt.Value
		}
	}

	for key, value := range item.Metadata {
		if value != nil {
			object.Metadata[key] = *value
		}
	}

	if item.Properties == nil {
		return object
	}

	if item.Properties.ETag != nil {
		etag := string(*item.Properties.ETag)
		object.ETag = &etag
	}

	if item.Properties.LastModified != nil {
		object.LastModified = *item.Properties.LastModified
	}

	if item.Properties.ContentType != nil {
		object.ContentType = item.Properties.ContentType
	}

	if item.Properties.CacheControl != nil {
		object.CacheControl = item.Properties.CacheControl
	}

	if item.Properties.ContentDisposition != nil {
		object.ContentDisposition = item.Properties.ContentDisposition
	}

	if len(item.Properties.ContentMD5) > 0 {
		contentMD5 := base64.StdEncoding.EncodeToString(item.Properties.ContentMD5)
		object.ContentMD5 = &contentMD5
	}

	if item.Properties.TagCount != nil {
		object.TagCount = int(*item.Properties.TagCount)
	} else {
		object.TagCount = len(object.Tags)
	}

	if item.Properties.ExpiresOn != nil {
		object.Expires = item.Properties.ExpiresOn
	}

	if item.Properties.Owner != nil {
		object.Owner = &common.StorageOwner{
			DisplayName: item.Properties.Owner,
		}
	}

	if item.Properties.ContentLength != nil {
		object.Size = item.Properties.ContentLength
	}

	object.ACL = item.Properties.ACL
	object.StorageClass = (*string)(item.Properties.AccessTier)
	object.AccessTierChangeTime = item.Properties.AccessTierChangeTime
	object.AccessTierInferred = item.Properties.AccessTierInferred
	object.ArchiveStatus = (*string)(item.Properties.ArchiveStatus)
	object.BlobSequenceNumber = item.Properties.BlobSequenceNumber
	object.BlobType = (*string)(item.Properties.BlobType)

	if item.Properties.CopyID != nil {
		object.Copy = &common.StorageObjectCopyInfo{
			CompletionTime:    item.Properties.CopyCompletionTime,
			ID:                *item.Properties.CopyID,
			Progress:          item.Properties.CopyProgress,
			Source:            item.Properties.CopySource,
			Status:            (*string)(item.Properties.CopyStatus),
			StatusDescription: item.Properties.CopyStatusDescription,
		}
	}

	object.CreationTime = item.Properties.CreationTime
	object.DeletedTime = item.Properties.DeletedTime
	object.CustomerProvidedKeySHA256 = item.Properties.CustomerProvidedKeySHA256
	object.DestinationSnapshot = item.Properties.DestinationSnapshot
	object.ServerEncrypted = item.Properties.ServerEncrypted
	object.KMSKeyName = item.Properties.EncryptionScope
	object.Group = item.Properties.Group
	object.RetentionUntilDate = item.Properties.ImmutabilityPolicyExpiresOn
	object.RetentionMode = (*string)(item.Properties.ImmutabilityPolicyMode)
	object.IncrementalCopy = item.Properties.IncrementalCopy
	object.IsSealed = item.Properties.IsSealed
	object.LastAccessTime = item.Properties.LastAccessedOn
	object.LeaseDuration = (*string)(item.Properties.LeaseDuration)
	object.LeaseState = (*string)(item.Properties.LeaseState)
	object.LeaseStatus = (*string)(item.Properties.LeaseStatus)
	object.LegalHold = item.Properties.LegalHold
	object.Permissions = item.Properties.Permissions
	object.RehydratePriority = (*string)(item.Properties.RehydratePriority)
	object.ResourceType = item.Properties.ResourceType
	object.RemainingRetentionDays = item.Properties.RemainingRetentionDays

	return object
}

func serializeUploadObjectInfo(resp azblob.UploadStreamResponse) common.StorageUploadInfo {
	object := common.StorageUploadInfo{
		LastModified: resp.LastModified,
	}

	if len(resp.ContentCRC64) > 0 {
		crc64 := string(resp.ContentCRC64)
		object.ChecksumCRC64NVME = &crc64
	}

	if len(resp.ContentMD5) > 0 {
		contentMD5 := base64.StdEncoding.EncodeToString(resp.ContentMD5)
		object.ContentMD5 = &contentMD5
	}

	if resp.ETag != nil {
		etag, _ := strconv.Unquote(string(*resp.ETag))
		object.ETag = &etag
	}

	return object
}

func serializeAzureErrorResponse(respErr *azcore.ResponseError) *schema.ConnectorError {
	details := map[string]any{
		"statusCode": respErr.StatusCode,
	}

	if respErr.ErrorCode != "" {
		details["code"] = respErr.ErrorCode
	}

	if respErr.StatusCode >= 500 {
		return schema.NewConnectorError(respErr.StatusCode, respErr.Error(), details)
	}

	return schema.UnprocessableContentError(respErr.Error(), details)
}

func serializeErrorResponse(err error) *schema.ConnectorError {
	var respErr *azcore.ResponseError
	if !errors.As(err, &respErr) {
		return schema.UnprocessableContentError(err.Error(), nil)
	}

	return serializeAzureErrorResponse(respErr)
}
