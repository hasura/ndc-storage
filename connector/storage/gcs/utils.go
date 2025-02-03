package gcs

import (
	"encoding/base64"
	"errors"
	"math"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/api/googleapi"
)

const (
	size256K = 256 * 1024
)

var errNotSupported = schema.NotSupportedError("Google Cloud Storage doesn't support this method", nil)

func serializeBucketInfo(bucket *storage.BucketAttrs) common.StorageBucket {
	result := common.StorageBucket{
		Name:                  bucket.Name,
		Tags:                  bucket.Labels,
		CORS:                  make([]common.BucketCors, len(bucket.CORS)),
		CreationTime:          &bucket.Created,
		LastModified:          &bucket.Updated,
		DefaultEventBasedHold: &bucket.DefaultEventBasedHold,
		RequesterPays:         &bucket.RequesterPays,
		StorageClass:          &bucket.StorageClass,
		ObjectLock:            serializeRetentionPolicy(bucket.RetentionPolicy),
		Lifecycle:             serializeLifecycleConfiguration(bucket.Lifecycle),
		Versioning: &common.StorageBucketVersioningConfiguration{
			Enabled: bucket.VersioningEnabled,
		},
	}

	if bucket.Location != "" {
		result.Region = &bucket.Location
	}

	if bucket.LocationType != "" {
		result.LocationType = &bucket.LocationType
	}

	if bucket.Etag != "" {
		result.Etag = &bucket.Etag
	}

	if bucket.Autoclass != nil {
		result.Autoclass = &common.BucketAutoclass{
			Enabled:                        bucket.Autoclass.Enabled,
			ToggleTime:                     bucket.Autoclass.ToggleTime,
			TerminalStorageClass:           bucket.Autoclass.TerminalStorageClass,
			TerminalStorageClassUpdateTime: bucket.Autoclass.TerminalStorageClassUpdateTime,
		}
	}

	for i, cors := range bucket.CORS {
		result.CORS[i] = common.BucketCors{
			MaxAge:          scalar.NewDuration(cors.MaxAge),
			Methods:         cors.Methods,
			Origins:         cors.Origins,
			ResponseHeaders: cors.ResponseHeaders,
		}
	}

	if bucket.CustomPlacementConfig != nil {
		result.CustomPlacementConfig = &common.CustomPlacementConfig{
			DataLocations: bucket.CustomPlacementConfig.DataLocations,
		}
	}

	if bucket.HierarchicalNamespace != nil {
		result.HierarchicalNamespace = &common.BucketHierarchicalNamespace{
			Enabled: bucket.HierarchicalNamespace.Enabled,
		}
	}

	if bucket.Logging != nil {
		result.Logging = &common.BucketLogging{
			LogBucket:       bucket.Logging.LogBucket,
			LogObjectPrefix: bucket.Logging.LogObjectPrefix,
		}
	}

	if bucket.RPO != storage.RPOUnknown {
		rpo := bucket.RPO.String()
		result.RPO = (*common.GoogleStorageRPO)(&rpo)
	}

	if bucket.SoftDeletePolicy != nil {
		result.SoftDeletePolicy = &common.StorageObjectSoftDeletePolicy{
			EffectiveTime:     bucket.SoftDeletePolicy.EffectiveTime,
			RetentionDuration: scalar.NewDuration(bucket.SoftDeletePolicy.RetentionDuration),
		}
	}

	if bucket.Website != nil {
		result.Website = &common.BucketWebsite{
			MainPageSuffix: bucket.Website.MainPageSuffix,
			NotFoundPage:   &bucket.Website.NotFoundPage,
		}

		if bucket.Website.NotFoundPage != "" {
			result.Website.NotFoundPage = &bucket.Website.NotFoundPage
		}
	}

	if bucket.Encryption != nil && bucket.Encryption.DefaultKMSKeyName != "" {
		result.Encryption = &common.ServerSideEncryptionConfiguration{
			KmsMasterKeyID: bucket.Encryption.DefaultKMSKeyName,
		}
	}

	return result
}

func serializeRetentionPolicy(retentionPolicy *storage.RetentionPolicy) *common.StorageObjectLockConfig {
	if retentionPolicy == nil {
		return nil
	}

	unit := common.StorageRetentionValidityUnitDays
	validity := uint(math.Ceil(retentionPolicy.RetentionPeriod.Hours()))
	mode := common.StorageRetentionModeUnlocked

	if retentionPolicy.IsLocked {
		mode = common.StorageRetentionModeLocked
	}

	return &common.StorageObjectLockConfig{
		Enabled: true,
		SetStorageObjectLockConfig: common.SetStorageObjectLockConfig{
			Mode:     &mode,
			Validity: &validity,
			Unit:     &unit,
		},
	}
}

func serializeObjectInfo(obj *storage.ObjectAttrs) common.StorageObject { //nolint:cyclop
	object := common.StorageObject{
		Bucket:       obj.Bucket,
		Name:         obj.Name,
		CreationTime: &obj.Created,
		LastModified: obj.Updated,
		Size:         &obj.Size,
		Metadata:     obj.Metadata,
		LegalHold:    &obj.TemporaryHold,
	}

	if obj.Name == "" && obj.Prefix != "" {
		object.Name = obj.Prefix
		object.IsDirectory = true
	}

	if obj.StorageClass != "" {
		object.StorageClass = &obj.StorageClass
	}

	if obj.Etag != "" {
		object.ETag = &obj.Etag
	}

	if obj.ContentType != "" {
		object.ContentType = &obj.ContentType
	}

	if obj.CacheControl != "" {
		object.CacheControl = &obj.CacheControl
	}

	if obj.ContentDisposition != "" {
		object.ContentDisposition = &obj.ContentDisposition
	}

	if obj.ContentEncoding != "" {
		object.ContentEncoding = &obj.ContentEncoding
	}

	if obj.ContentLanguage != "" {
		object.ContentLanguage = &obj.ContentLanguage
	}

	if obj.CustomerKeySHA256 != "" {
		object.CustomerProvidedKeySHA256 = &obj.CustomerKeySHA256
	}

	if obj.KMSKeyName != "" {
		object.KMSKeyName = &obj.KMSKeyName
	}

	if obj.MediaLink != "" {
		object.MediaLink = &obj.MediaLink
	}

	if obj.Owner != "" {
		object.Owner = &common.StorageOwner{
			DisplayName: &obj.Owner,
		}
	}

	if obj.Retention != nil {
		if obj.Retention.Mode != "" {
			object.RetentionMode = &obj.Retention.Mode
		}

		if !obj.Retention.RetainUntil.IsZero() {
			object.RetentionUntilDate = &obj.Retention.RetainUntil
		}
	}

	if !obj.RetentionExpirationTime.IsZero() {
		object.Expiration = &obj.RetentionExpirationTime
	}

	if !obj.Deleted.IsZero() {
		deleted := true
		object.Deleted = &deleted
		object.DeletedTime = &obj.Deleted
	}

	if obj.Generation > 0 {
		versionID := strconv.Itoa(int(obj.Generation))
		object.VersionID = &versionID
	}

	if len(obj.MD5) > 0 {
		contentMd5 := base64.StdEncoding.EncodeToString(obj.MD5)
		object.ContentMD5 = &contentMd5
	}

	aclRules := make([]ACLRule, len(obj.ACL))

	for i, acl := range obj.ACL {
		aclRules[i] = makeACLRule(acl)
	}

	object.ACL = aclRules

	return object
}

func (c *Client) validateListObjectsOptions(span trace.Span, opts *common.ListStorageObjectsOptions, includeDeleted bool) *storage.Query {
	span.SetAttributes(
		attribute.Bool("storage.options.hierarchy", opts.Hierarchy),
		attribute.Bool("storage.options.with_deleted", includeDeleted),
		attribute.Bool("storage.options.with_versions", opts.Include.Versions),
	)

	if opts.Prefix != "" {
		span.SetAttributes(attribute.String("storage.options.prefix", opts.Prefix))
	}

	if opts.StartAfter != "" {
		span.SetAttributes(attribute.String("storage.options.start_after", opts.StartAfter))
	}

	if opts.MaxResults > 0 {
		span.SetAttributes(attribute.Int("storage.options.max_results", opts.MaxResults))
	}

	result := &storage.Query{
		Versions:    opts.Include.Versions,
		Prefix:      opts.Prefix,
		StartOffset: opts.StartAfter,
		SoftDeleted: includeDeleted,
	}

	if opts.Hierarchy {
		result.Delimiter = "/"
	}

	return result
}

func serializeUploadObjectInfo(obj *storage.Writer) common.StorageUploadInfo {
	object := common.StorageUploadInfo{
		Bucket: obj.Bucket,
		Name:   obj.Name,
	}

	if obj.Etag != "" {
		object.ETag = &obj.Etag
	}

	if obj.Size > 0 {
		object.Size = &obj.Size
	}

	if !obj.Updated.IsZero() {
		object.LastModified = &obj.Updated
	} else if !obj.Created.IsZero() {
		object.LastModified = &obj.Created
	}

	if !obj.RetentionExpirationTime.IsZero() {
		object.Expiration = &obj.RetentionExpirationTime
	}

	versionID := strconv.Itoa(int(obj.Generation))
	object.VersionID = &versionID

	if len(obj.MD5) > 0 {
		contentMd5 := base64.StdEncoding.EncodeToString(obj.MD5)
		object.ContentMD5 = &contentMd5
	}

	return object
}

func validateLifecycleRule(rule common.ObjectLifecycleRule) storage.LifecycleRule {
	r := storage.LifecycleRule{}

	for _, filter := range rule.RuleFilter {
		r.Condition.MatchesPrefix = append(r.Condition.MatchesPrefix, filter.MatchesPrefix...)
		r.Condition.MatchesSuffix = append(r.Condition.MatchesSuffix, filter.MatchesSuffix...)
		r.Condition.MatchesStorageClasses = append(r.Condition.MatchesStorageClasses, filter.MatchesStorageClasses...)
	}

	if rule.NoncurrentVersionExpiration != nil {
		if rule.NoncurrentVersionExpiration.NewerNoncurrentVersions != nil {
			r.Condition.NumNewerVersions = int64(*rule.NoncurrentVersionExpiration.NewerNoncurrentVersions)
		}

		if rule.NoncurrentVersionExpiration.NoncurrentDays != nil {
			r.Condition.DaysSinceNoncurrentTime = int64(*rule.NoncurrentVersionExpiration.NoncurrentDays)
		}

		r.Action.Type = storage.DeleteAction
	}

	if rule.NoncurrentVersionTransition != nil {
		if rule.NoncurrentVersionTransition.NewerNoncurrentVersions != nil {
			r.Condition.NumNewerVersions = int64(*rule.NoncurrentVersionTransition.NewerNoncurrentVersions)
		}

		if rule.NoncurrentVersionTransition.NoncurrentDays != nil {
			r.Condition.DaysSinceNoncurrentTime = int64(*rule.NoncurrentVersionTransition.NoncurrentDays)
		}

		if rule.NoncurrentVersionTransition.StorageClass != nil {
			r.Action.StorageClass = *rule.NoncurrentVersionTransition.StorageClass
		}

		r.Action.Type = storage.SetStorageClassAction
	}

	if rule.Expiration != nil && rule.Expiration.Days != nil {
		r.Condition.AgeInDays = int64(*rule.Expiration.Days)
		r.Condition.AllObjects = *rule.Expiration.Days == 0

		r.Action.Type = storage.DeleteAction

		return r
	}

	if rule.Transition != nil || rule.Transition.StorageClass != nil {
		if rule.Transition.Days != nil {
			r.Condition.AgeInDays = int64(*rule.Transition.Days)
		} else if rule.Transition.Date != nil {
			r.Condition.AgeInDays = int64(time.Since(rule.Expiration.Date.Time).Hours() / 24)
		}

		r.Condition.AllObjects = r.Condition.AgeInDays == 0
		r.Action.StorageClass = *rule.Transition.StorageClass
		r.Action.Type = storage.SetStorageClassAction

		return r
	}

	if rule.AbortIncompleteMultipartUpload != nil && rule.AbortIncompleteMultipartUpload.DaysAfterInitiation != nil {
		r.Action.Type = storage.AbortIncompleteMPUAction
		r.Condition.AgeInDays = int64(*rule.AbortIncompleteMultipartUpload.DaysAfterInitiation)
		r.Condition.AllObjects = r.Condition.AgeInDays == 0

		return r
	}

	if rule.AllVersionsExpiration != nil && rule.AllVersionsExpiration.Days != nil {
		r.Condition.AgeInDays = int64(*rule.AllVersionsExpiration.Days)
		r.Condition.AllObjects = r.Condition.AgeInDays == 0
		r.Action.Type = storage.DeleteAction

		return r
	}

	return r
}

func validateLifecycleConfiguration(input common.ObjectLifecycleConfiguration) *storage.Lifecycle {
	result := &storage.Lifecycle{
		Rules: make([]storage.LifecycleRule, len(input.Rules)),
	}

	for i, rule := range input.Rules {
		if !rule.Enabled {
			continue
		}

		r := validateLifecycleRule(rule)
		result.Rules[i] = r
	}

	return result
}

func serializeLifecycleConfiguration(input storage.Lifecycle) *common.ObjectLifecycleConfiguration {
	result := &common.ObjectLifecycleConfiguration{
		Rules: make([]common.ObjectLifecycleRule, len(input.Rules)),
	}

	for i, rule := range input.Rules {
		r := serializeLifecycleRule(rule)
		result.Rules[i] = r
	}

	return result
}

func serializeLifecycleRule(rule storage.LifecycleRule) common.ObjectLifecycleRule {
	r := common.ObjectLifecycleRule{}

	if len(rule.Condition.MatchesPrefix) > 0 || len(rule.Condition.MatchesSuffix) > 0 || len(rule.Condition.MatchesStorageClasses) > 0 {
		r.RuleFilter = []common.ObjectLifecycleFilter{
			{
				MatchesPrefix:         rule.Condition.MatchesPrefix,
				MatchesSuffix:         rule.Condition.MatchesSuffix,
				MatchesStorageClasses: rule.Condition.MatchesStorageClasses,
			},
		}
	}

	ageInDays := int(rule.Condition.AgeInDays)

	switch rule.Action.Type {
	case storage.SetStorageClassAction:
		r.Transition = &common.ObjectLifecycleTransition{
			Days:         &ageInDays,
			StorageClass: &rule.Action.StorageClass,
		}
	case storage.DeleteAction:
		r.Expiration = &common.ObjectLifecycleExpiration{
			Days: &ageInDays,
		}
	case storage.AbortIncompleteMPUAction:
		r.AbortIncompleteMultipartUpload = &common.ObjectAbortIncompleteMultipartUpload{
			DaysAfterInitiation: &ageInDays,
		}
	}

	return r
}

func evalGoogleErrorResponse(err *googleapi.Error) *schema.ConnectorError {
	details := map[string]any{
		"statusCode": err.Code,
		"details":    err.Details,
	}

	if err.Code >= 500 {
		return schema.NewConnectorError(err.Code, err.Message, details)
	}

	return schema.UnprocessableContentError(err.Message, details)
}

func serializeErrorResponse(err error) *schema.ConnectorError {
	var e *googleapi.Error
	if ok := errors.As(err, &e); ok {
		return evalGoogleErrorResponse(e)
	}

	return schema.UnprocessableContentError(err.Error(), nil)
}
