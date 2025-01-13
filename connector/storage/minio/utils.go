package minio

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/notification"
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

func serializeObjectInfo(obj minio.ObjectInfo, fromList bool) common.StorageObject { //nolint:funlen,gocognit,gocyclo,cyclop
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
		Metadata:              map[string]string{},
		UserMetadata:          map[string]string{},
		UserTags:              obj.UserTags,
		UserTagCount:          obj.UserTagCount,
		Grant:                 grants,
		IsLatest:              &obj.IsLatest,
		Deleted:               &obj.IsDeleteMarker,
		ReplicationReady:      &obj.ReplicationReady,
		StorageObjectChecksum: checksum,
	}

	if fromList {
		object.Metadata = obj.UserMetadata

		for key, value := range obj.UserMetadata {
			lowerKey := strings.ToLower(key)
			if strings.HasPrefix(lowerKey, userMetadataHeaderPrefix) {
				object.UserMetadata[key[len(userMetadataHeaderPrefix):]] = value

				continue
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
		object.UserMetadata = obj.UserMetadata

		for key, values := range obj.Metadata {
			if len(values) == 0 {
				continue
			}

			value := strings.Join(values, ", ")
			object.Metadata[key] = value

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
	if mc.providerType == common.GoogleStorage && opts.WithVersions {
		// Force versioning off. GCS doesn't support AWS S3 compatible versioning API.
		opts.WithVersions = false
	}

	span.SetAttributes(
		attribute.Bool("storage.options.recursive", opts.Recursive),
		attribute.Bool("storage.options.with_versions", opts.WithVersions),
		attribute.Bool("storage.options.with_metadata", opts.WithMetadata),
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
		WithVersions: opts.WithVersions,
		WithMetadata: opts.WithMetadata,
		Prefix:       opts.Prefix,
		Recursive:    opts.Recursive,
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
		ETag:                  obj.ETag,
		Name:                  obj.Key,
		Size:                  obj.Size,
		StorageObjectChecksum: checksum,
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
	options := minio.GetObjectOptions{}
	if opts.VersionID != nil && !isStringNull(*opts.VersionID) {
		options.VersionID = *opts.VersionID
		span.SetAttributes(attribute.String("storage.request_object_version", options.VersionID))
	}

	if opts.PartNumber != nil {
		options.PartNumber = *opts.PartNumber
		span.SetAttributes(attribute.Int("storage.part_number", options.PartNumber))
	}

	if opts.Checksum != nil {
		options.Checksum = *opts.Checksum
		span.SetAttributes(attribute.Bool("storage.request_object_checksum", options.Checksum))
	}

	for key, value := range opts.Headers {
		span.SetAttributes(attribute.StringSlice("http.request.header."+key, []string{value}))
		options.Set(key, value)
	}

	if len(opts.RequestParams) > 0 {
		q := url.Values{}

		for key, values := range opts.RequestParams {
			for _, value := range values {
				options.AddReqParam(key, value)
				q.Add(key, value)
			}
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

func convertCopyDestOptions(dst common.StorageCopyDestOptions) (*minio.CopyDestOptions, error) {
	destOptions := minio.CopyDestOptions{
		Bucket:          dst.Bucket,
		Object:          dst.Object,
		UserMetadata:    dst.UserMetadata,
		ReplaceMetadata: dst.ReplaceMetadata,
		UserTags:        dst.UserTags,
		ReplaceTags:     dst.ReplaceTags,
		Size:            dst.Size,
	}

	if dst.RetainUntilDate != nil {
		destOptions.RetainUntilDate = *dst.RetainUntilDate
	}

	if dst.Mode != nil {
		mode := minio.RetentionMode(string(*dst.Mode))
		if !mode.IsValid() {
			return nil, fmt.Errorf("invalid RetentionMode: %s", *dst.Mode)
		}

		destOptions.Mode = mode
	}

	if dst.LegalHold != nil {
		legalHold := minio.LegalHoldStatus(*dst.LegalHold)
		if !legalHold.IsValid() {
			return nil, fmt.Errorf("invalid LegalHoldStatus: %s", *dst.LegalHold)
		}

		destOptions.LegalHold = legalHold
	}

	return &destOptions, nil
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

func isStringNull(input string) bool {
	return input == "" || input == "null"
}
