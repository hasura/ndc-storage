package minio

import (
	"context"

	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
	"go.opentelemetry.io/otel/codes"
)

// SetBucketLifecycle sets lifecycle on bucket or an object prefix.
func (mc *Client) SetBucketLifecycle(ctx context.Context, bucketName string, config common.BucketLifecycleConfiguration) error {
	ctx, span := mc.startOtelSpan(ctx, "SetBucketLifecycle", bucketName)
	defer span.End()

	input := validateLifecycleConfiguration(config)

	err := mc.client.SetBucketLifecycle(ctx, bucketName, &input)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// GetBucketLifecycle gets lifecycle on a bucket or a prefix.
func (mc *Client) GetBucketLifecycle(ctx context.Context, bucketName string) (*common.BucketLifecycleConfiguration, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetBucketLifecycle", bucketName)
	defer span.End()

	rawResult, err := mc.client.GetBucketLifecycle(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := serializeLifecycleConfiguration(*rawResult)

	return &result, nil
}

func validateLifecycleRule(rule common.BucketLifecycleRule) lifecycle.Rule {
	r := lifecycle.Rule{
		ID:         rule.ID,
		Expiration: validateLifecycleExpiration(rule.Expiration),
		RuleFilter: validateLifecycleFilter(rule.RuleFilter),
		Transition: validateLifecycleTransition(rule.Transition),
	}

	if rule.AbortIncompleteMultipartUpload != nil && rule.AbortIncompleteMultipartUpload.DaysAfterInitiation != nil {
		r.AbortIncompleteMultipartUpload.DaysAfterInitiation = lifecycle.ExpirationDays(*rule.AbortIncompleteMultipartUpload.DaysAfterInitiation)
	}

	if rule.AllVersionsExpiration != nil && (rule.AllVersionsExpiration.Days != nil || rule.AllVersionsExpiration.DeleteMarker != nil) {
		if rule.AllVersionsExpiration.Days != nil {
			r.AllVersionsExpiration.Days = *rule.AllVersionsExpiration.Days
		}

		if rule.DelMarkerExpiration != nil {
			r.AllVersionsExpiration.DeleteMarker = lifecycle.ExpireDeleteMarker(*rule.AllVersionsExpiration.DeleteMarker)
		}
	}

	if rule.DelMarkerExpiration != nil && rule.DelMarkerExpiration.Days != nil {
		r.DelMarkerExpiration.Days = *rule.DelMarkerExpiration.Days
	}

	if rule.NoncurrentVersionExpiration != nil {
		if rule.NoncurrentVersionExpiration.NewerNoncurrentVersions != nil {
			r.NoncurrentVersionExpiration.NewerNoncurrentVersions = *rule.NoncurrentVersionExpiration.NewerNoncurrentVersions
		}

		if rule.NoncurrentVersionExpiration.NoncurrentDays != nil {
			r.NoncurrentVersionExpiration.NoncurrentDays = lifecycle.ExpirationDays(*rule.NoncurrentVersionExpiration.NoncurrentDays)
		}
	}

	if rule.NoncurrentVersionTransition != nil {
		if rule.NoncurrentVersionTransition.NewerNoncurrentVersions != nil {
			r.NoncurrentVersionTransition.NewerNoncurrentVersions = *rule.NoncurrentVersionTransition.NewerNoncurrentVersions
		}

		if rule.NoncurrentVersionTransition.NoncurrentDays != nil {
			r.NoncurrentVersionTransition.NoncurrentDays = lifecycle.ExpirationDays(*rule.NoncurrentVersionTransition.NoncurrentDays)
		}

		if rule.NoncurrentVersionTransition.StorageClass != nil {
			r.NoncurrentVersionTransition.StorageClass = *rule.NoncurrentVersionTransition.StorageClass
		}
	}

	if rule.Prefix != nil {
		r.Prefix = *rule.Prefix
	}

	if rule.Status != nil {
		r.Status = *rule.Status
	}

	return r
}

func validateLifecycleExpiration(input *common.LifecycleExpiration) lifecycle.Expiration {
	result := lifecycle.Expiration{}

	if input == nil || input.IsEmpty() {
		return result
	}

	if input.Days != nil {
		result.Days = lifecycle.ExpirationDays(*input.Days)
	}

	if input.Date != nil {
		result.Date = lifecycle.ExpirationDate(*input.Date)
	}

	if input.DeleteMarker != nil {
		result.DeleteMarker = lifecycle.ExpireDeleteMarker(*input.DeleteMarker)
	}

	if input.DeleteAll != nil {
		result.DeleteAll = lifecycle.ExpirationBoolean(*input.DeleteAll)
	}

	return result
}

func validateLifecycleTransition(input *common.LifecycleTransition) lifecycle.Transition {
	result := lifecycle.Transition{}

	if input == nil {
		return result
	}

	if input.Days != nil {
		result.Days = lifecycle.ExpirationDays(*input.Days)
	}

	if input.Date != nil {
		result.Date = lifecycle.ExpirationDate(*input.Date)
	}

	if input.StorageClass != nil {
		result.StorageClass = *input.StorageClass
	}

	return result
}

func validateLifecycleConfiguration(input common.BucketLifecycleConfiguration) lifecycle.Configuration {
	result := lifecycle.Configuration{
		Rules: make([]lifecycle.Rule, len(input.Rules)),
	}

	for i, rule := range input.Rules {
		r := validateLifecycleRule(rule)
		result.Rules[i] = r
	}

	return result
}

func validateLifecycleFilter(input *common.LifecycleFilter) lifecycle.Filter {
	result := lifecycle.Filter{}

	if input == nil {
		return result
	}

	if input.Prefix != nil {
		result.Prefix = *input.Prefix
	}

	if input.ObjectSizeGreaterThan != nil {
		result.ObjectSizeGreaterThan = *input.ObjectSizeGreaterThan
	}

	if input.ObjectSizeLessThan != nil {
		result.ObjectSizeLessThan = *input.ObjectSizeLessThan
	}

	if input.Tag != nil {
		if input.Tag.Key != nil {
			result.Tag.Key = *input.Tag.Key
		}

		if input.Tag.Value != nil {
			result.Tag.Value = *input.Tag.Value
		}
	}

	if input.And != nil {
		if input.And.Prefix != nil {
			result.And.Prefix = *input.And.Prefix
		}

		if input.And.ObjectSizeGreaterThan != nil {
			result.And.ObjectSizeGreaterThan = *input.And.ObjectSizeGreaterThan
		}

		if input.And.ObjectSizeLessThan != nil {
			result.And.ObjectSizeLessThan = *input.And.ObjectSizeLessThan
		}

		result.And.Tags = make([]lifecycle.Tag, len(input.And.Tags))

		for i, t := range input.And.Tags {
			tag := lifecycle.Tag{}

			if t.Key != nil {
				tag.Key = *t.Key
			}

			if t.Value != nil {
				tag.Value = *t.Value
			}

			result.And.Tags[i] = tag
		}
	}

	return result
}

func serializeLifecycleRule(rule lifecycle.Rule) common.BucketLifecycleRule {
	r := common.BucketLifecycleRule{
		ID:         rule.ID,
		RuleFilter: serializeLifecycleFilter(rule.RuleFilter),
		Transition: serializeLifecycleTransition(rule.Transition),
	}

	if !rule.AbortIncompleteMultipartUpload.IsDaysNull() {
		days := int(rule.AbortIncompleteMultipartUpload.DaysAfterInitiation)
		r.AbortIncompleteMultipartUpload = &common.AbortIncompleteMultipartUpload{
			DaysAfterInitiation: &days,
		}
	}

	if !rule.AllVersionsExpiration.IsNull() {
		deleteMarker := bool(rule.AllVersionsExpiration.DeleteMarker)
		r.AllVersionsExpiration = &common.LifecycleAllVersionsExpiration{
			Days:         &rule.AllVersionsExpiration.Days,
			DeleteMarker: &deleteMarker,
		}
	}

	if !rule.DelMarkerExpiration.IsNull() {
		r.DelMarkerExpiration.Days = &rule.DelMarkerExpiration.Days
	}

	if !rule.Expiration.IsNull() {
		r.Expiration = &common.LifecycleExpiration{
			DeleteMarker: (*bool)(&rule.Expiration.DeleteMarker),
		}

		if rule.Expiration.Days != 0 {
			r.Expiration.Days = (*int)(&rule.Expiration.Days)
		}

		if !rule.Expiration.Date.IsZero() {
			r.Expiration.Date = &scalar.Date{Time: rule.Expiration.Date.Time}
		}
	}

	if !rule.NoncurrentVersionExpiration.IsDaysNull() || rule.NoncurrentVersionExpiration.NewerNoncurrentVersions != 0 {
		r.NoncurrentVersionExpiration = &common.LifecycleNoncurrentVersionExpiration{}

		if rule.NoncurrentVersionExpiration.NewerNoncurrentVersions != 0 {
			r.NoncurrentVersionExpiration.NewerNoncurrentVersions = &rule.NoncurrentVersionExpiration.NewerNoncurrentVersions
		}

		if !rule.NoncurrentVersionExpiration.IsDaysNull() {
			days := int(rule.NoncurrentVersionExpiration.NoncurrentDays)
			r.NoncurrentVersionExpiration.NoncurrentDays = &days
		}
	}

	if !rule.NoncurrentVersionTransition.IsDaysNull() || rule.NoncurrentVersionTransition.NewerNoncurrentVersions != 0 && rule.NoncurrentVersionTransition.StorageClass != "" {
		if rule.NoncurrentVersionTransition.NewerNoncurrentVersions != 0 {
			r.NoncurrentVersionTransition.NewerNoncurrentVersions = &rule.NoncurrentVersionTransition.NewerNoncurrentVersions
		}

		if rule.NoncurrentVersionTransition.NoncurrentDays != 0 {
			days := int(rule.NoncurrentVersionTransition.NoncurrentDays)
			r.NoncurrentVersionTransition.NoncurrentDays = &days
		}

		if rule.NoncurrentVersionTransition.StorageClass != "" {
			r.NoncurrentVersionTransition.StorageClass = &rule.NoncurrentVersionTransition.StorageClass
		}
	}

	if rule.Prefix != "" {
		r.Prefix = &rule.Prefix
	}

	if rule.Status != "" {
		r.Status = &rule.Status
	}

	return r
}

func serializeLifecycleTransition(input lifecycle.Transition) *common.LifecycleTransition {
	if input.IsNull() {
		return nil
	}

	result := common.LifecycleTransition{}

	if input.Days != 0 {
		result.Days = (*int)(&input.Days)
	}

	if !input.Date.IsZero() {
		result.Date = &scalar.Date{Time: input.Date.Time}
	}

	if input.StorageClass != "" {
		result.StorageClass = &input.StorageClass
	}

	return &result
}

func serializeLifecycleConfiguration(input lifecycle.Configuration) common.BucketLifecycleConfiguration {
	result := common.BucketLifecycleConfiguration{
		Rules: make([]common.BucketLifecycleRule, len(input.Rules)),
	}

	for i, rule := range input.Rules {
		r := serializeLifecycleRule(rule)
		result.Rules[i] = r
	}

	return result
}

func serializeLifecycleFilter(input lifecycle.Filter) *common.LifecycleFilter {
	result := common.LifecycleFilter{}

	if input.Prefix != "" {
		result.Prefix = &input.Prefix
	}

	if input.ObjectSizeGreaterThan != 0 {
		result.ObjectSizeGreaterThan = &input.ObjectSizeGreaterThan
	}

	if input.ObjectSizeLessThan != 0 {
		result.ObjectSizeLessThan = &input.ObjectSizeLessThan
	}

	if input.Tag.Key != "" || input.Tag.Value != "" {
		if input.Tag.Key != "" {
			result.Tag.Key = &input.Tag.Key
		}

		if input.Tag.Value != "" {
			result.Tag.Value = &input.Tag.Value
		}
	}

	if !input.And.IsEmpty() {
		result.And = &common.LifecycleFilterAnd{}

		if input.And.Prefix != "" {
			result.And.Prefix = &input.And.Prefix
		}

		if input.And.ObjectSizeGreaterThan != 0 {
			result.And.ObjectSizeGreaterThan = &input.And.ObjectSizeGreaterThan
		}

		if input.And.ObjectSizeLessThan != 0 {
			result.And.ObjectSizeLessThan = &input.And.ObjectSizeLessThan
		}

		result.And.Tags = make([]common.StorageTag, 0, len(input.And.Tags))

		for _, t := range input.And.Tags {
			if t.IsEmpty() {
				continue
			}

			tag := common.StorageTag{}

			if t.Key != "" {
				tag.Key = &t.Key
			}

			if t.Value != "" {
				tag.Value = &t.Value
			}

			result.And.Tags = append(result.And.Tags, tag)
		}
	}

	return &result
}
