package minio

import (
	"context"

	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
	"go.opentelemetry.io/otel/codes"
)

// SetBucketLifecycle sets lifecycle on bucket or an object prefix.
func (mc *Client) SetBucketLifecycle(ctx context.Context, bucketName string, config common.ObjectLifecycleConfiguration) error {
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
func (mc *Client) GetBucketLifecycle(ctx context.Context, bucketName string) (*common.ObjectLifecycleConfiguration, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetBucketLifecycle", bucketName)
	defer span.End()

	rawResult, err := mc.client.GetBucketLifecycle(ctx, bucketName)
	if err != nil {
		respError := evalNotFoundError(err, "NoSuchLifecycleConfiguration")
		if respError == nil {
			return nil, nil
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, respError
	}

	result := serializeLifecycleConfiguration(*rawResult)

	return &result, nil
}

func validateLifecycleRule(rule common.ObjectLifecycleRule) lifecycle.Rule {
	r := lifecycle.Rule{
		ID:         rule.ID,
		Status:     "Enabled",
		Expiration: validateLifecycleExpiration(rule.Expiration),
		RuleFilter: validateLifecycleFilters(rule.RuleFilter),
		Transition: validateLifecycleTransition(rule.Transition),
	}

	if !rule.Enabled {
		r.Status = "Disabled"
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

	return r
}

func validateLifecycleExpiration(input *common.ObjectLifecycleExpiration) lifecycle.Expiration {
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

func validateLifecycleTransition(input *common.ObjectLifecycleTransition) lifecycle.Transition {
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

func validateLifecycleConfiguration(input common.ObjectLifecycleConfiguration) lifecycle.Configuration {
	result := lifecycle.Configuration{
		Rules: make([]lifecycle.Rule, len(input.Rules)),
	}

	for i, rule := range input.Rules {
		r := validateLifecycleRule(rule)
		result.Rules[i] = r
	}

	return result
}

func validateLifecycleFilters(input []common.ObjectLifecycleFilter) lifecycle.Filter {
	result := lifecycle.Filter{}

	inputLen := len(input)
	if inputLen == 0 {
		return result
	}

	for key, value := range input[0].Tags {
		if result.Tag.Key == "" {
			result.Tag.Key = key
			result.Tag.Value = value

			continue
		}

		result.And.Tags = append(result.And.Tags, lifecycle.Tag{
			Key:   key,
			Value: value,
		})
	}

	if len(input[0].MatchesPrefix) > 0 {
		result.Prefix = input[0].MatchesPrefix[0]
	}

	if input[0].ObjectSizeGreaterThan != nil {
		result.ObjectSizeGreaterThan = *input[0].ObjectSizeGreaterThan
	}

	if input[0].ObjectSizeLessThan != nil {
		result.ObjectSizeLessThan = *input[0].ObjectSizeLessThan
	}

	if inputLen == 1 {
		return result
	}

	for key, value := range input[1].Tags {
		result.And.Tags = append(result.And.Tags, lifecycle.Tag{
			Key:   key,
			Value: value,
		})
	}

	if len(input[1].MatchesPrefix) > 0 {
		result.And.Prefix = input[1].MatchesPrefix[0]
	}

	if input[1].ObjectSizeGreaterThan != nil {
		result.And.ObjectSizeGreaterThan = *input[1].ObjectSizeGreaterThan
	}

	if input[1].ObjectSizeLessThan != nil {
		result.And.ObjectSizeLessThan = *input[1].ObjectSizeLessThan
	}

	return result
}

func serializeLifecycleRule(rule lifecycle.Rule) common.ObjectLifecycleRule {
	r := common.ObjectLifecycleRule{
		ID:         rule.ID,
		Enabled:    rule.Status == "Enabled",
		RuleFilter: serializeLifecycleFilter(rule.RuleFilter),
		Transition: serializeLifecycleTransition(rule.Transition),
	}

	if !rule.AbortIncompleteMultipartUpload.IsDaysNull() {
		days := int(rule.AbortIncompleteMultipartUpload.DaysAfterInitiation)
		r.AbortIncompleteMultipartUpload = &common.ObjectAbortIncompleteMultipartUpload{
			DaysAfterInitiation: &days,
		}
	}

	if !rule.AllVersionsExpiration.IsNull() {
		deleteMarker := bool(rule.AllVersionsExpiration.DeleteMarker)
		r.AllVersionsExpiration = &common.ObjectLifecycleAllVersionsExpiration{
			Days:         &rule.AllVersionsExpiration.Days,
			DeleteMarker: &deleteMarker,
		}
	}

	if !rule.DelMarkerExpiration.IsNull() {
		r.DelMarkerExpiration.Days = &rule.DelMarkerExpiration.Days
	}

	if !rule.Expiration.IsNull() {
		r.Expiration = &common.ObjectLifecycleExpiration{
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
		r.NoncurrentVersionExpiration = &common.ObjectLifecycleNoncurrentVersionExpiration{}

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

	return r
}

func serializeLifecycleTransition(input lifecycle.Transition) *common.ObjectLifecycleTransition {
	if input.IsNull() {
		return nil
	}

	result := common.ObjectLifecycleTransition{}

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

func serializeLifecycleConfiguration(input lifecycle.Configuration) common.ObjectLifecycleConfiguration {
	result := common.ObjectLifecycleConfiguration{
		Rules: make([]common.ObjectLifecycleRule, len(input.Rules)),
	}

	for i, rule := range input.Rules {
		r := serializeLifecycleRule(rule)
		result.Rules[i] = r
	}

	return result
}

func serializeLifecycleFilter(input lifecycle.Filter) []common.ObjectLifecycleFilter {
	result := []common.ObjectLifecycleFilter{}

	firstItem := common.ObjectLifecycleFilter{}
	if input.Prefix != "" {
		firstItem.MatchesPrefix = []string{input.Prefix}
	}

	if input.ObjectSizeGreaterThan != 0 {
		firstItem.ObjectSizeGreaterThan = &input.ObjectSizeGreaterThan
	}

	if input.ObjectSizeLessThan != 0 {
		firstItem.ObjectSizeLessThan = &input.ObjectSizeLessThan
	}

	if input.Tag.Key != "" || input.Tag.Value != "" {
		firstItem.Tags = map[string]string{
			input.Tag.Key: input.Tag.Value,
		}
	}

	result = append(result, firstItem)

	if input.And.IsEmpty() {
		return result
	}

	sndItem := common.ObjectLifecycleFilter{}
	if input.And.Prefix != "" {
		sndItem.MatchesPrefix = []string{input.And.Prefix}
	}

	if input.And.ObjectSizeGreaterThan != 0 {
		sndItem.ObjectSizeGreaterThan = &input.And.ObjectSizeGreaterThan
	}

	if input.And.ObjectSizeLessThan != 0 {
		sndItem.ObjectSizeLessThan = &input.And.ObjectSizeLessThan
	}

	sndItem.Tags = make(map[string]string)

	for _, t := range input.And.Tags {
		if t.IsEmpty() {
			continue
		}

		sndItem.Tags[t.Key] = t.Value
	}

	result = append(result, sndItem)

	return result
}
