package minio

import (
	"context"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/replication"
	"github.com/minio/minio-go/v7/pkg/tags"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// MakeBucket creates a new bucket.
func (mc *Client) MakeBucket(ctx context.Context, args *common.MakeStorageBucketOptions) error {
	ctx, span := mc.startOtelSpan(ctx, "MakeBucket", args.Name)
	defer span.End()

	err := mc.client.MakeBucket(ctx, args.Name, minio.MakeBucketOptions{
		Region:        args.Region,
		ObjectLocking: args.ObjectLocking,
	})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// ListBuckets lists all buckets.
func (mc *Client) ListBuckets(ctx context.Context) ([]common.StorageBucketInfo, error) {
	ctx, span := mc.startOtelSpan(ctx, "ListBuckets", "")
	defer span.End()

	bucketInfos, err := mc.client.ListBuckets(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	results := make([]common.StorageBucketInfo, len(bucketInfos))
	for i, item := range bucketInfos {
		results[i] = common.StorageBucketInfo{
			Name:         item.Name,
			CreationDate: item.CreationDate,
		}
	}

	span.SetAttributes(attribute.Int("storage.bucket_count", len(results)))

	return results, nil
}

// BucketExists checks if a bucket exists.
func (mc *Client) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	ctx, span := mc.startOtelSpan(ctx, "BucketExists", bucketName)
	defer span.End()

	existed, err := mc.client.BucketExists(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return false, serializeErrorResponse(err)
	}

	span.SetAttributes(attribute.Bool("storage.bucket_exist", existed))

	return existed, nil
}

// RemoveBucket removes a bucket, bucket should be empty to be successfully removed.
func (mc *Client) RemoveBucket(ctx context.Context, bucketName string) error {
	ctx, span := mc.startOtelSpan(ctx, "RemoveBucket", bucketName)
	defer span.End()

	err := mc.client.RemoveBucket(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// GetBucketTagging gets tags of a bucket.
func (mc *Client) GetBucketTagging(ctx context.Context, bucketName string) (map[string]string, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetBucketTagging", bucketName)
	defer span.End()

	bucketTags, err := mc.client.GetBucketTagging(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := bucketTags.ToMap()
	for key, value := range result {
		span.SetAttributes(attribute.String("storage.bucket_tag"+key, value))
	}

	return result, nil
}

// Removes all tags on a bucket.
func (mc *Client) RemoveBucketTagging(ctx context.Context, bucketName string) error {
	ctx, span := mc.startOtelSpan(ctx, "RemoveBucketTagging", bucketName)
	defer span.End()

	err := mc.client.RemoveBucketTagging(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// SetBucketTagging sets tags to a bucket.
func (mc *Client) SetBucketTagging(ctx context.Context, args *common.SetStorageBucketTaggingArguments) error {
	ctx, span := mc.startOtelSpan(ctx, "SetBucketTagging", args.Bucket)
	defer span.End()

	for key, value := range args.Tags {
		span.SetAttributes(attribute.String("storage.bucket_tag"+key, value))
	}

	inputTags, err := tags.NewTags(args.Tags, false)
	if err != nil {
		span.SetStatus(codes.Error, "failed to convert minio tags")
		span.RecordError(err)

		return schema.UnprocessableContentError(err.Error(), nil)
	}

	err = mc.client.SetBucketTagging(ctx, args.Bucket, inputTags)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// GetBucketPolicy gets access permissions on a bucket or a prefix.
func (mc *Client) GetBucketPolicy(ctx context.Context, bucketName string) (string, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetBucketPolicy", bucketName)
	defer span.End()

	result, err := mc.client.GetBucketPolicy(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", serializeErrorResponse(err)
	}

	return result, nil
}

// GetBucketNotification gets notification configuration on a bucket.
func (mc *Client) GetBucketNotification(ctx context.Context, bucketName string) (*common.NotificationConfig, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetBucketNotification", bucketName)
	defer span.End()

	result, err := mc.client.GetBucketNotification(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	return serializeBucketNotificationConfig(result), nil
}

// SetBucketNotification sets a new bucket notification on a bucket.
func (mc *Client) SetBucketNotification(ctx context.Context, bucketName string, config common.NotificationConfig) error {
	ctx, span := mc.startOtelSpan(ctx, "SetBucketNotification", bucketName)
	defer span.End()

	input, err := validateBucketNotificationConfig(config)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return schema.UnprocessableContentError(err.Error(), nil)
	}

	if err := mc.client.SetBucketNotification(ctx, bucketName, *input); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// RemoveAllBucketNotification removes all configured bucket notifications on a bucket.
func (mc *Client) RemoveAllBucketNotification(ctx context.Context, bucketName string) error {
	ctx, span := mc.startOtelSpan(ctx, "RemoveAllBucketNotification", bucketName)
	defer span.End()

	err := mc.client.RemoveAllBucketNotification(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// GetBucketVersioning gets the versioning configuration set on a bucket.
func (mc *Client) GetBucketVersioning(ctx context.Context, bucketName string) (*common.StorageBucketVersioningConfiguration, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetBucketVersioning", bucketName)
	defer span.End()

	rawResult, err := mc.client.GetBucketVersioning(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := &common.StorageBucketVersioningConfiguration{
		ExcludedPrefixes: make([]string, len(rawResult.ExcludedPrefixes)),
		ExcludeFolders:   &rawResult.ExcludeFolders,
	}

	if rawResult.Status != "" {
		result.Status = &rawResult.Status
	}

	if rawResult.MFADelete != "" {
		result.MFADelete = &rawResult.MFADelete
	}

	for i, prefix := range rawResult.ExcludedPrefixes {
		result.ExcludedPrefixes[i] = prefix.Prefix
	}

	return result, nil
}

// SetBucketReplication sets replication configuration on a bucket. Role can be obtained by first defining the replication target
// to associate the source and destination buckets for replication with the replication endpoint.
func (mc *Client) SetBucketReplication(ctx context.Context, bucketName string, cfg common.StorageReplicationConfig) error {
	ctx, span := mc.startOtelSpan(ctx, "SetBucketReplication", bucketName)
	defer span.End()

	input := validateBucketReplicationConfig(cfg)

	err := mc.client.SetBucketReplication(ctx, bucketName, input)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// Get current replication config on a bucket.
func (mc *Client) GetBucketReplication(ctx context.Context, bucketName string) (*common.StorageReplicationConfig, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetBucketReplication", bucketName)
	defer span.End()

	result, err := mc.client.GetBucketReplication(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	return serializeBucketReplicationConfig(result), nil
}

// RemoveBucketReplication removes replication configuration on a bucket.
func (mc *Client) RemoveBucketReplication(ctx context.Context, bucketName string) error {
	ctx, span := mc.startOtelSpan(ctx, "RemoveBucketReplication", bucketName)
	defer span.End()

	err := mc.client.RemoveBucketReplication(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

func validateBucketReplicationConfig(input common.StorageReplicationConfig) replication.Config {
	result := replication.Config{
		Rules: make([]replication.Rule, len(input.Rules)),
	}

	if input.Role != nil {
		result.Role = *input.Role
	}

	for i, item := range input.Rules {
		result.Rules[i] = validateBucketReplicationRule(item)
	}

	return result
}

func validateBucketReplicationRule(item common.StorageReplicationRule) replication.Rule {
	rule := replication.Rule{
		Status:   replication.Status(item.Status),
		Priority: item.Priority,
		Filter:   validateBucketReplicationFilter(item.Filter),
	}

	if item.ID != nil {
		rule.ID = *item.ID
	}

	if item.DeleteMarkerReplication != nil && item.DeleteMarkerReplication.Status != "" {
		rule.DeleteMarkerReplication.Status = replication.Status(item.DeleteMarkerReplication.Status)
	}

	if item.DeleteReplication != nil && item.DeleteReplication.Status != "" {
		rule.DeleteReplication.Status = replication.Status(item.DeleteReplication.Status)
	}

	if item.ExistingObjectReplication != nil && item.ExistingObjectReplication.Status != "" {
		rule.ExistingObjectReplication.Status = replication.Status(item.ExistingObjectReplication.Status)
	}

	if item.SourceSelectionCriteria != nil && item.SourceSelectionCriteria.ReplicaModifications != nil && item.SourceSelectionCriteria.ReplicaModifications.Status != "" {
		rule.SourceSelectionCriteria.ReplicaModifications.Status = replication.Status(item.SourceSelectionCriteria.ReplicaModifications.Status)
	}

	if item.Destination != nil {
		rule.Destination = replication.Destination{
			Bucket: item.Destination.Bucket,
		}

		if item.Destination.StorageClass != nil {
			rule.Destination.StorageClass = *item.Destination.StorageClass
		}
	}

	return rule
}

func validateBucketReplicationFilter(input common.StorageReplicationFilter) replication.Filter {
	result := replication.Filter{}

	if input.Prefix != nil {
		result.Prefix = *input.Prefix
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
			result.And.Prefix = *input.Prefix
		}

		result.And.Tags = make([]replication.Tag, len(input.And.Tags))

		for i, tag := range input.And.Tags {
			t := replication.Tag{}
			if tag.Key != nil {
				t.Key = *tag.Key
			}

			if tag.Value != nil {
				t.Value = *tag.Value
			}

			result.And.Tags[i] = t
		}
	}

	return result
}

func serializeBucketReplicationConfig(input replication.Config) *common.StorageReplicationConfig {
	result := common.StorageReplicationConfig{
		Rules: make([]common.StorageReplicationRule, len(input.Rules)),
	}

	if input.Role != "" {
		result.Role = &input.Role
	}

	for i, item := range input.Rules {
		result.Rules[i] = serializeBucketReplicationRule(item)
	}

	return &result
}

func serializeBucketReplicationRule(item replication.Rule) common.StorageReplicationRule {
	rule := common.StorageReplicationRule{
		Status:   common.StorageReplicationRuleStatus(item.Status),
		Priority: item.Priority,
	}

	if item.ID != "" {
		rule.ID = &item.ID
	}

	if item.DeleteMarkerReplication.Status != "" {
		rule.DeleteMarkerReplication.Status = common.StorageReplicationRuleStatus(item.DeleteMarkerReplication.Status)
	}

	if item.DeleteReplication.Status != "" {
		rule.DeleteReplication.Status = common.StorageReplicationRuleStatus(item.DeleteReplication.Status)
	}

	if item.ExistingObjectReplication.Status != "" {
		rule.ExistingObjectReplication.Status = common.StorageReplicationRuleStatus(item.ExistingObjectReplication.Status)
	}

	if item.SourceSelectionCriteria.ReplicaModifications.Status != "" {
		rule.SourceSelectionCriteria = &common.SourceSelectionCriteria{
			ReplicaModifications: &common.ReplicaModifications{
				Status: common.StorageReplicationRuleStatus(item.SourceSelectionCriteria.ReplicaModifications.Status),
			},
		}
	}

	rule.Destination = &common.StorageReplicationDestination{
		Bucket: item.Destination.Bucket,
	}

	if item.Destination.StorageClass != "" {
		rule.Destination.StorageClass = &item.Destination.StorageClass
	}

	if item.Filter.Prefix != "" {
		rule.Filter.Prefix = &item.Filter.Prefix
	}

	if item.Filter.Tag.Key != "" || item.Filter.Tag.Value != "" {
		rule.Filter.Tag = &common.StorageTag{}

		if item.Filter.Tag.Key != "" {
			rule.Filter.Tag.Key = &item.Filter.Tag.Key
		}

		if item.Filter.Tag.Value != "" {
			rule.Filter.Tag.Value = &item.Filter.Tag.Value
		}
	}

	if item.Filter.And.Prefix != "" || len(item.Filter.And.Tags) > 0 {
		rule.Filter.And = &common.StorageReplicationFilterAnd{}
		if item.Filter.And.Prefix != "" {
			rule.Filter.And.Prefix = &item.Filter.Prefix
		}

		rule.Filter.And.Tags = make([]common.StorageTag, len(item.Filter.And.Tags))

		for i, tag := range item.Filter.And.Tags {
			t := common.StorageTag{}
			if tag.Key != "" {
				t.Key = &tag.Key
			}

			if tag.Value != "" {
				t.Value = &tag.Value
			}

			rule.Filter.And.Tags[i] = t
		}
	}

	return rule
}
