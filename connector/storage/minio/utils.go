package minio

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/notification"
	"github.com/minio/minio-go/v7/pkg/sse"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const userMetadataHeaderPrefix = "x-amz-meta-"

func serializeGrant(grant minio.Grant) common.StorageGrant {
	g := common.StorageGrant{}

	if !isStringNull(grant.Permission) {
		g.Permission = &grant.Permission
	}

	if !isStringNull(grant.Grantee.ID) || !isStringNull(grant.Grantee.DisplayName) || !isStringNull(grant.Grantee.URI) {
		g.Grantee = &common.StorageGrantee{}

		if !isStringNull(grant.Grantee.ID) {
			g.Grantee.ID = &grant.Grantee.ID
		}

		if !isStringNull(grant.Grantee.DisplayName) {
			g.Grantee.DisplayName = &grant.Grantee.DisplayName
		}

		if !isStringNull(grant.Grantee.URI) {
			g.Grantee.URI = &grant.Grantee.URI
		}
	}

	return g
}

func serializeObjectInfo(obj *minio.ObjectInfo, fromList bool) common.StorageObject { //nolint:funlen,gocognit,gocyclo,cyclop
	grants := make([]common.StorageGrant, len(obj.Grant))

	for i, grant := range obj.Grant {
		grants[i] = serializeGrant(grant)
	}

	checksum := common.StorageObjectChecksum{}
	if !isStringNull(obj.ChecksumCRC32) {
		checksum.ChecksumCRC32 = &obj.ChecksumCRC32
	}

	if !isStringNull(obj.ChecksumCRC32C) {
		checksum.ChecksumCRC32C = &obj.ChecksumCRC32C
	}

	if !isStringNull(obj.ChecksumSHA1) {
		checksum.ChecksumSHA1 = &obj.ChecksumSHA1
	}

	if !isStringNull(obj.ChecksumSHA256) {
		checksum.ChecksumSHA256 = &obj.ChecksumSHA256
	}

	object := common.StorageObject{
		Name:                  obj.Key,
		LastModified:          obj.LastModified,
		Size:                  &obj.Size,
		Tags:                  common.StringMapToKeyValues(obj.UserTags),
		TagCount:              obj.UserTagCount,
		Grant:                 grants,
		IsLatest:              &obj.IsLatest,
		Deleted:               &obj.IsDeleteMarker,
		ReplicationReady:      &obj.ReplicationReady,
		StorageObjectChecksum: checksum,
		IsDirectory:           strings.HasSuffix(obj.Key, "/"),
	}

	if object.TagCount == 0 {
		object.TagCount = len(object.Tags)
	}

	if fromList {
		object.RawMetadata = common.StringMapToKeyValues(obj.UserMetadata)

		for key, value := range obj.UserMetadata {
			lowerKey := strings.ToLower(key)
			if strings.HasPrefix(lowerKey, userMetadataHeaderPrefix) {
				object.Metadata = append(object.Metadata, common.StorageKeyValue{
					Key:   key[len(userMetadataHeaderPrefix):],
					Value: value,
				})
			}

			switch lowerKey {
			case common.HeaderContentType:
				object.ContentType = &value
			case common.HeaderCacheControl:
				object.CacheControl = &value
			case common.HeaderContentDisposition:
				object.ContentDisposition = &value
			case common.HeaderContentEncoding:
				object.ContentEncoding = &value
			case common.HeaderContentLanguage:
				object.ContentLanguage = &value
			}
		}
	} else {
		object.Metadata = common.StringMapToKeyValues(obj.UserMetadata)
		keys := utils.GetSortedKeys(obj.Metadata)

		for _, key := range keys {
			values := obj.Metadata[key]

			if len(values) == 0 {
				continue
			}

			value := strings.Join(values, ", ")
			object.RawMetadata = append(object.RawMetadata, common.StorageKeyValue{
				Key:   key,
				Value: value,
			})

			switch strings.ToLower(key) {
			case common.HeaderContentType:
				object.ContentType = &value
			case common.HeaderCacheControl:
				object.CacheControl = &value
			case common.HeaderContentDisposition:
				object.ContentDisposition = &value
			case common.HeaderContentEncoding:
				object.ContentEncoding = &value
			case common.HeaderContentLanguage:
				object.ContentLanguage = &value
			}
		}
	}

	if !isStringNull(obj.ETag) {
		object.ETag = &obj.ETag
	}

	if !isStringNull(obj.ContentType) {
		object.ContentType = &obj.ContentType
	}

	if !obj.Expires.IsZero() {
		object.Expires = &obj.Expires
	}

	if !isStringNull(obj.Owner.DisplayName) || !isStringNull(obj.Owner.ID) {
		object.Owner = &common.StorageOwner{}
		if !isStringNull(obj.Owner.DisplayName) {
			object.Owner.DisplayName = &obj.Owner.DisplayName
		}

		if !isStringNull(obj.Owner.ID) {
			object.Owner.ID = &obj.Owner.ID
		}
	}

	if !isStringNull(obj.StorageClass) {
		object.StorageClass = &obj.StorageClass
	}

	if !isStringNull(obj.VersionID) {
		object.VersionID = &obj.VersionID
	}

	if !isStringNull(obj.ExpirationRuleID) {
		object.ExpirationRuleID = &obj.ExpirationRuleID
	}

	if !obj.Expiration.IsZero() {
		object.Expiration = &obj.Expiration
	}

	if !isStringNull(obj.ReplicationStatus) {
		replicationStatus := common.StorageObjectReplicationStatus(obj.ReplicationStatus)
		object.ReplicationStatus = &replicationStatus
	}

	if obj.Restore != nil {
		object.Restore = &common.StorageRestoreInfo{
			OngoingRestore: obj.Restore.OngoingRestore,
		}

		if !obj.Restore.ExpiryTime.IsZero() {
			object.Restore.ExpiryTime = &obj.Restore.ExpiryTime
		}
	}

	return object
}

func (mc *Client) validateListObjectsOptions(span trace.Span, opts *common.ListStorageObjectsOptions) minio.ListObjectsOptions {
	if mc.providerType == common.GoogleStorage && opts.Include.Versions {
		// Force versioning off. GCS doesn't support AWS S3 compatible versioning API.
		opts.Include.Versions = false
	}

	span.SetAttributes(
		attribute.Bool("storage.options.recursive", !opts.Hierarchy),
		attribute.Bool("storage.options.with_versions", opts.Include.Versions),
		attribute.Bool("storage.options.with_metadata", opts.Include.Metadata),
	)

	if opts.Prefix != "" {
		span.SetAttributes(attribute.String("storage.options.prefix", opts.Prefix))
	}

	if !isStringNull(opts.StartAfter) {
		span.SetAttributes(attribute.String("storage.options.start_after", opts.StartAfter))
	}

	if opts.MaxResults > 0 {
		span.SetAttributes(attribute.Int("storage.options.max_results", opts.MaxResults))
	}

	return minio.ListObjectsOptions{
		WithVersions: opts.Include.Versions,
		WithMetadata: opts.Include.Metadata,
		Prefix:       opts.Prefix,
		Recursive:    !opts.Hierarchy,
		MaxKeys:      opts.MaxResults,
		StartAfter:   opts.StartAfter,
	}
}

func serializeUploadObjectInfo(obj minio.UploadInfo) common.StorageUploadInfo {
	checksum := common.StorageObjectChecksum{}
	if !isStringNull(obj.ChecksumCRC32) {
		checksum.ChecksumCRC32 = &obj.ChecksumCRC32
	}

	if !isStringNull(obj.ChecksumCRC32C) {
		checksum.ChecksumCRC32C = &obj.ChecksumCRC32C
	}

	if !isStringNull(obj.ChecksumSHA1) {
		checksum.ChecksumSHA1 = &obj.ChecksumSHA1
	}

	if !isStringNull(obj.ChecksumSHA256) {
		checksum.ChecksumSHA256 = &obj.ChecksumSHA256
	}

	object := common.StorageUploadInfo{
		Bucket:                obj.Bucket,
		Name:                  obj.Key,
		StorageObjectChecksum: checksum,
	}

	if !isStringNull(obj.ETag) {
		object.ETag = &obj.ETag
	}

	if obj.Size > 0 {
		object.Size = &obj.Size
	}

	if !obj.LastModified.IsZero() {
		object.LastModified = &obj.LastModified
	}

	if !obj.Expiration.IsZero() {
		object.Expiration = &obj.Expiration
	}

	if !isStringNull(obj.Location) {
		object.Location = &obj.Location
	}

	if !isStringNull(obj.VersionID) {
		object.VersionID = &obj.VersionID
	}

	if !isStringNull(obj.ExpirationRuleID) {
		object.ExpirationRuleID = &obj.ExpirationRuleID
	}

	return object
}

func serializeGetObjectOptions(span trace.Span, opts common.GetStorageObjectOptions) minio.GetObjectOptions {
	options := minio.GetObjectOptions{
		Checksum: opts.Include.Checksum,
	}

	span.SetAttributes(attribute.Bool("storage.request_object_checksum", options.Checksum))

	if opts.VersionID != nil && !isStringNull(*opts.VersionID) {
		options.VersionID = *opts.VersionID
		span.SetAttributes(attribute.String("storage.request_object_version", options.VersionID))
	}

	if opts.PartNumber != nil {
		options.PartNumber = *opts.PartNumber
		span.SetAttributes(attribute.Int("storage.part_number", options.PartNumber))
	}

	for _, item := range opts.Headers {
		span.SetAttributes(attribute.StringSlice("http.request.header."+item.Key, []string{item.Value}))
		options.Set(item.Key, item.Value)
	}

	if len(opts.RequestParams) > 0 {
		q := url.Values{}

		for _, item := range opts.RequestParams {
			options.AddReqParam(item.Key, item.Value)
			q.Add(item.Key, item.Value)
		}

		span.SetAttributes(attribute.String("url.query", q.Encode()))
	}

	return options
}

func serializeCopySourceOptions(src common.StorageCopySrcOptions) minio.CopySrcOptions {
	srcOptions := minio.CopySrcOptions{
		Bucket:      src.Bucket,
		Object:      src.Object,
		VersionID:   src.VersionID,
		MatchETag:   src.MatchETag,
		NoMatchETag: src.NoMatchETag,
		MatchRange:  src.MatchRange,
		Start:       src.Start,
		End:         src.End,
	}

	if src.MatchModifiedSince != nil {
		srcOptions.MatchModifiedSince = *src.MatchModifiedSince
	}

	if src.MatchUnmodifiedSince != nil {
		srcOptions.MatchUnmodifiedSince = *src.MatchUnmodifiedSince
	}

	return srcOptions
}

func convertCopyDestOptions(dst common.StorageCopyDestOptions) *minio.CopyDestOptions {
	destOptions := minio.CopyDestOptions{
		Bucket:          dst.Bucket,
		Object:          dst.Object,
		UserMetadata:    common.KeyValuesToStringMap(dst.Metadata),
		ReplaceMetadata: dst.Metadata != nil,
		UserTags:        common.KeyValuesToStringMap(dst.Tags),
		ReplaceTags:     dst.Tags != nil,
		Size:            dst.Size,
		LegalHold:       validateLegalHoldStatus(dst.LegalHold),
	}

	if dst.RetainUntilDate != nil {
		destOptions.RetainUntilDate = *dst.RetainUntilDate
	}

	if dst.Mode != nil {
		destOptions.Mode = validateObjectRetentionMode(*dst.Mode)
	}

	return &destOptions
}

func validateLegalHoldStatus(input *bool) minio.LegalHoldStatus {
	if input == nil {
		return ""
	}

	if *input {
		return minio.LegalHoldEnabled
	}

	return minio.LegalHoldDisabled
}

func validateObjectRetentionMode(input common.StorageRetentionMode) minio.RetentionMode {
	if input == common.StorageRetentionModeLocked {
		return minio.Compliance
	}

	return minio.Governance
}

func serializeObjectRetentionMode(input *minio.RetentionMode) *common.StorageRetentionMode {
	if input == nil {
		return nil
	}

	result := common.StorageRetentionModeUnlocked
	if *input == minio.Compliance {
		result = common.StorageRetentionModeLocked
	}

	return &result
}

func serializeBucketNotificationCommonConfig(item notification.Config) common.NotificationCommonConfig {
	cfg := common.NotificationCommonConfig{
		Events: make([]string, len(item.Events)),
	}

	if !isStringNull(item.ID) {
		cfg.ID = &item.ID
	}

	if item.Filter != nil {
		cfg.Filter.S3Key.FilterRules = make([]common.NotificationFilterRule, len(item.Filter.S3Key.FilterRules))
		for i, rule := range item.Filter.S3Key.FilterRules {
			cfg.Filter.S3Key.FilterRules[i] = common.NotificationFilterRule(rule)
		}
	}

	if item.Arn.AccountID != "" || item.Arn.Partition != "" || item.Arn.Resource != "" || item.Arn.Service != "" {
		arn := item.Arn.String()
		cfg.Arn = &arn
	}

	for i, eventType := range item.Events {
		cfg.Events[i] = string(eventType)
	}

	return cfg
}

func validateBucketNotificationCommonConfig(item common.NotificationCommonConfig) (*notification.Config, error) {
	cfg := notification.Config{
		Events: make([]notification.EventType, len(item.Events)),
	}

	if item.ID != nil {
		cfg.ID = *item.ID
	}

	if item.Filter != nil && item.Filter.S3Key != nil {
		cfg.Filter = &notification.Filter{
			S3Key: notification.S3Key{
				FilterRules: make([]notification.FilterRule, len(item.Filter.S3Key.FilterRules)),
			},
		}

		for i, rule := range item.Filter.S3Key.FilterRules {
			cfg.Filter.S3Key.FilterRules[i] = notification.FilterRule(rule)
		}
	}

	if item.Arn != nil {
		arn, err := notification.NewArnFromString(*item.Arn)
		if err != nil {
			return nil, err
		}

		cfg.Arn = arn
	}

	for i, eventType := range item.Events {
		cfg.Events[i] = notification.EventType(eventType)
	}

	return &cfg, nil
}

func validateBucketNotificationConfig(input common.NotificationConfig) (*notification.Configuration, error) {
	result := notification.Configuration{
		LambdaConfigs: make([]notification.LambdaConfig, len(input.LambdaConfigs)),
		TopicConfigs:  make([]notification.TopicConfig, len(input.TopicConfigs)),
		QueueConfigs:  make([]notification.QueueConfig, len(input.QueueConfigs)),
	}

	for i, item := range input.LambdaConfigs {
		commonCfg, err := validateBucketNotificationCommonConfig(item.NotificationCommonConfig)
		if err != nil {
			return nil, fmt.Errorf("cloudFunctionConfigurations[%d]: %w", i, err)
		}

		cfg := notification.LambdaConfig{
			Lambda: item.Lambda,
			Config: *commonCfg,
		}

		result.LambdaConfigs[i] = cfg
	}

	for i, item := range input.QueueConfigs {
		commonCfg, err := validateBucketNotificationCommonConfig(item.NotificationCommonConfig)
		if err != nil {
			return nil, fmt.Errorf("queueConfigurations[%d]: %w", i, err)
		}

		cfg := notification.QueueConfig{
			Queue:  item.Queue,
			Config: *commonCfg,
		}

		result.QueueConfigs[i] = cfg
	}

	for i, item := range input.TopicConfigs {
		commonCfg, err := validateBucketNotificationCommonConfig(item.NotificationCommonConfig)
		if err != nil {
			return nil, fmt.Errorf("topicConfigurations[%d]: %w", i, err)
		}

		cfg := notification.TopicConfig{
			Topic:  item.Topic,
			Config: *commonCfg,
		}

		result.TopicConfigs[i] = cfg
	}

	return &result, nil
}

func serializeBucketNotificationConfig(input notification.Configuration) *common.NotificationConfig {
	result := common.NotificationConfig{
		LambdaConfigs: make([]common.NotificationLambdaConfig, len(input.LambdaConfigs)),
		TopicConfigs:  make([]common.NotificationTopicConfig, len(input.TopicConfigs)),
		QueueConfigs:  make([]common.NotificationQueueConfig, len(input.QueueConfigs)),
	}

	for i, item := range input.LambdaConfigs {
		cfg := common.NotificationLambdaConfig{
			Lambda:                   item.Lambda,
			NotificationCommonConfig: serializeBucketNotificationCommonConfig(item.Config),
		}

		result.LambdaConfigs[i] = cfg
	}

	for i, item := range input.QueueConfigs {
		cfg := common.NotificationQueueConfig{
			Queue:                    item.Queue,
			NotificationCommonConfig: serializeBucketNotificationCommonConfig(item.Config),
		}

		result.QueueConfigs[i] = cfg
	}

	for i, item := range input.TopicConfigs {
		cfg := common.NotificationTopicConfig{
			Topic:                    item.Topic,
			NotificationCommonConfig: serializeBucketNotificationCommonConfig(item.Config),
		}

		result.TopicConfigs[i] = cfg
	}

	return &result
}

func parseChecksumType(input common.ChecksumType) minio.ChecksumType {
	switch string(input) {
	case "SHA256":
		return minio.ChecksumSHA256
	case "SHA1":
		return minio.ChecksumSHA1
	case "CRC32":
		return minio.ChecksumCRC32
	case "CRC32C":
		return minio.ChecksumCRC32C
	case "CRC64NVME":
		return minio.ChecksumCRC64NVME
	case "FullObjectCRC32":
		return minio.ChecksumFullObjectCRC32
	case "FullObjectCRC32C":
		return minio.ChecksumFullObjectCRC32C
	default:
		return minio.ChecksumNone
	}
}

func evalMinioErrorResponse(err minio.ErrorResponse) *schema.ConnectorError {
	details := map[string]any{
		"statusCode": err.StatusCode,
		"server":     err.Server,
	}

	if err.Code != "" {
		details["code"] = err.Code
	}

	if err.BucketName != "" {
		details["bucketName"] = err.BucketName
	}

	if err.Key != "" {
		details["key"] = err.Key
	}

	if err.HostID != "" {
		details["hostId"] = err.HostID
	}

	if err.RequestID != "" {
		details["requestId"] = err.RequestID
	}

	if err.Resource != "" {
		details["resource"] = err.Resource
	}

	if err.Region != "" {
		details["region"] = err.Region
	}

	if err.StatusCode >= 500 {
		return schema.NewConnectorError(err.StatusCode, err.Message, details)
	}

	return schema.UnprocessableContentError(err.Message, details)
}

func serializeErrorResponse(err error) *schema.ConnectorError {
	var errResponse minio.ErrorResponse
	if errors.As(err, &errResponse) {
		return evalMinioErrorResponse(errResponse)
	}

	errRespPtr := &errResponse
	if errors.As(err, &errRespPtr) {
		return evalMinioErrorResponse(*errRespPtr)
	}

	return schema.UnprocessableContentError(err.Error(), nil)
}

func evalNotFoundError(err error, notFoundCode string) *schema.ConnectorError {
	var errResponse minio.ErrorResponse
	if !errors.As(err, &errResponse) {
		errRespPtr := &errResponse
		if errors.As(err, &errRespPtr) {
			errResponse = *errRespPtr
		}
	}

	if errResponse.Code == notFoundCode {
		return nil
	}

	if errResponse.StatusCode > 0 {
		return evalMinioErrorResponse(errResponse)
	}

	return schema.UnprocessableContentError(err.Error(), nil)
}

func isStringNull(input string) bool {
	return input == "" || input == "null"
}

func validateBucketEncryptionConfiguration(input common.ServerSideEncryptionConfiguration) *sse.Configuration {
	if input.SSEAlgorithm == "AES256" {
		return sse.NewConfigurationSSES3()
	}

	return sse.NewConfigurationSSEKMS(input.KmsMasterKeyID)
}

func serializeBucketEncryptionConfiguration(input *sse.Configuration) *common.ServerSideEncryptionConfiguration {
	if input == nil || len(input.Rules) == 0 {
		return nil
	}

	return &common.ServerSideEncryptionConfiguration{
		KmsMasterKeyID: input.Rules[0].Apply.KmsMasterKeyID,
		SSEAlgorithm:   input.Rules[0].Apply.SSEAlgorithm,
	}
}
