package minio

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/hasura/ndc-sdk-go/v2/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/replication"
	"github.com/minio/minio-go/v7/pkg/tags"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

// MakeBucket creates a new bucket.
func (mc *Client) MakeBucket(ctx context.Context, args *common.MakeStorageBucketOptions) error {
	ctx, span := mc.startOtelSpan(ctx, "MakeBucket", args.Name)
	defer span.End()

	err := mc.client.MakeBucket(ctx, args.Name, minio.MakeBucketOptions{
		Region:        args.Region,
		ObjectLocking: args.ObjectLock,
	})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	if len(args.Tags) > 0 {
		err := mc.SetBucketTagging(ctx, args.Name, common.KeyValuesToStringMap(args.Tags))
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return err
		}
	}

	return nil
}

// ListBuckets lists all buckets.
func (mc *Client) ListBuckets(
	ctx context.Context,
	options *common.ListStorageBucketsOptions,
	predicate func(string) bool,
) (*common.StorageBucketListResults, error) {
	ctx, span := mc.startOtelSpan(ctx, "ListBuckets", "")
	defer span.End()

	bucketInfos, err := mc.client.ListBuckets(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	filteredBuckets, pageInfo := filterBuckets(bucketInfos, options, predicate)
	span.SetAttributes(attribute.Int("storage.bucket_count", len(bucketInfos)))

	if len(bucketInfos) == 0 {
		return &common.StorageBucketListResults{
			Buckets: []common.StorageBucket{},
		}, nil
	}

	results := make([]common.StorageBucket, len(filteredBuckets))

	if options.NumThreads <= 1 {
		for i, item := range filteredBuckets {
			bucket, err := mc.populateBucket(ctx, item, common.BucketOptions{
				NumThreads: options.NumThreads,
				Include:    options.Include,
			})
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)

				return nil, err
			}

			results[i] = bucket
		}

		return &common.StorageBucketListResults{
			Buckets:  results,
			PageInfo: pageInfo,
		}, nil
	}

	eg := errgroup.Group{}
	eg.SetLimit(options.NumThreads)

	populateFunc := func(item minio.BucketInfo, index int) {
		eg.Go(func() error {
			bucket, err := mc.populateBucket(ctx, item, common.BucketOptions{
				NumThreads: options.NumThreads,
				Include:    options.Include,
			})
			if err != nil {
				return err
			}

			results[index] = bucket

			return nil
		})
	}

	for i, item := range filteredBuckets {
		populateFunc(item, i)
	}

	if err := eg.Wait(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, err
	}

	return &common.StorageBucketListResults{
		Buckets:  results,
		PageInfo: pageInfo,
	}, nil
}

// GetBucket gets a bucket by name.
func (mc *Client) GetBucket(
	ctx context.Context,
	name string,
	options common.BucketOptions,
) (*common.StorageBucket, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetBucket", "")
	defer span.End()

	bucketInfos, err := mc.client.ListBuckets(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	for _, item := range bucketInfos {
		if item.Name != name {
			continue
		}

		bucket, err := mc.populateBucket(ctx, item, options)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, err
		}

		return &bucket, nil
	}

	return nil, nil
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

// UpdateBucket updates configurations for the bucket.
func (mc *Client) UpdateBucket(
	ctx context.Context,
	bucketName string,
	opts common.UpdateStorageBucketOptions,
) error {
	ctx, span := mc.startOtelSpanWithKind(ctx, trace.SpanKindInternal, "UpdateBucket", bucketName)
	defer span.End()

	if opts.Tags != nil {
		err := mc.SetBucketTagging(ctx, bucketName, common.KeyValuesToStringMap(*opts.Tags))
		if err != nil {
			return err
		}
	}

	if opts.VersioningEnabled != nil {
		if *opts.VersioningEnabled {
			err := mc.EnableVersioning(ctx, bucketName)
			if err != nil {
				return err
			}
		} else {
			err := mc.SuspendVersioning(ctx, bucketName)
			if err != nil {
				return err
			}
		}
	}

	if opts.Lifecycle != nil {
		err := mc.SetBucketLifecycle(ctx, bucketName, *opts.Lifecycle)
		if err != nil {
			return err
		}
	}

	if opts.ObjectLock != nil {
		err := mc.SetObjectLockConfig(ctx, bucketName, *opts.ObjectLock)
		if err != nil {
			return err
		}
	}

	if opts.Encryption == nil {
		return nil
	}

	if opts.Encryption.IsEmpty() {
		return mc.RemoveBucketEncryption(ctx, bucketName)
	}

	return mc.SetBucketEncryption(ctx, bucketName, *opts.Encryption)
}

// GetBucketTagging gets tags of a bucket.
func (mc *Client) GetBucketTagging(
	ctx context.Context,
	bucketName string,
) (map[string]string, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetBucketTagging", bucketName)
	defer span.End()

	bucketTags, err := mc.client.GetBucketTagging(ctx, bucketName)
	if err != nil {
		var errResponse minio.ErrorResponse
		if errors.As(err, &errResponse) {
			if errResponse.StatusCode == http.StatusNotFound {
				return nil, nil
			}

			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, evalMinioErrorResponse(errResponse)
		}

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

// RemoveBucketTagging removes all tags on a bucket.
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
func (mc *Client) SetBucketTagging(
	ctx context.Context,
	bucketName string,
	bucketTags map[string]string,
) error {
	if len(bucketTags) == 0 {
		return mc.RemoveBucketTagging(ctx, bucketName)
	}

	ctx, span := mc.startOtelSpan(ctx, "SetBucketTagging", bucketName)
	defer span.End()

	for key, value := range bucketTags {
		span.SetAttributes(attribute.String("storage.bucket_tag"+key, value))
	}

	inputTags, err := tags.NewTags(bucketTags, false)
	if err != nil {
		span.SetStatus(codes.Error, "failed to convert minio tags")
		span.RecordError(err)

		return schema.UnprocessableContentError(err.Error(), nil)
	}

	err = mc.client.SetBucketTagging(ctx, bucketName, inputTags)
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
func (mc *Client) GetBucketNotification(
	ctx context.Context,
	bucketName string,
) (*common.NotificationConfig, error) {
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
func (mc *Client) SetBucketNotification(
	ctx context.Context,
	bucketName string,
	config common.NotificationConfig,
) error {
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
func (mc *Client) GetBucketVersioning(
	ctx context.Context,
	bucketName string,
) (*common.StorageBucketVersioningConfiguration, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetBucketVersioning", bucketName)
	defer span.End()

	rawResult, err := mc.client.GetBucketVersioning(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := &common.StorageBucketVersioningConfiguration{
		Enabled:          rawResult.Enabled(),
		ExcludedPrefixes: make([]string, len(rawResult.ExcludedPrefixes)),
		ExcludeFolders:   &rawResult.ExcludeFolders,
	}

	if rawResult.MFADelete != "" {
		result.MFADelete = &rawResult.MFADelete
	}

	for i, prefix := range rawResult.ExcludedPrefixes {
		result.ExcludedPrefixes[i] = prefix.Prefix
	}

	return result, nil
}

// EnableVersioning enables bucket versioning support.
func (mc *Client) EnableVersioning(ctx context.Context, bucketName string) error {
	ctx, span := mc.startOtelSpan(ctx, "EnableVersioning", bucketName)
	defer span.End()

	err := mc.client.EnableVersioning(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// SuspendVersioning disables bucket versioning support.
func (mc *Client) SuspendVersioning(ctx context.Context, bucketName string) error {
	ctx, span := mc.startOtelSpan(ctx, "SuspendVersioning", bucketName)
	defer span.End()

	err := mc.client.SuspendVersioning(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// SetBucketReplication sets replication configuration on a bucket. Role can be obtained by first defining the replication target
// to associate the source and destination buckets for replication with the replication endpoint.
func (mc *Client) SetBucketReplication(
	ctx context.Context,
	bucketName string,
	cfg common.StorageReplicationConfig,
) error {
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

// GetBucketReplication gets current replication config on a bucket.
func (mc *Client) GetBucketReplication(
	ctx context.Context,
	bucketName string,
) (*common.StorageReplicationConfig, error) {
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

// SetBucketEncryption sets default encryption configuration on a bucket.
func (mc *Client) SetBucketEncryption(
	ctx context.Context,
	bucketName string,
	input common.ServerSideEncryptionConfiguration,
) error {
	ctx, span := mc.startOtelSpan(ctx, "SetBucketEncryption", bucketName)
	defer span.End()

	err := mc.client.SetBucketEncryption(
		ctx,
		bucketName,
		validateBucketEncryptionConfiguration(input),
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// GetBucketEncryption gets default encryption configuration set on a bucket.
func (mc *Client) GetBucketEncryption(
	ctx context.Context,
	bucketName string,
) (*common.ServerSideEncryptionConfiguration, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetBucketEncryption", bucketName)
	defer span.End()

	rawResult, err := mc.client.GetBucketEncryption(ctx, bucketName)
	if err != nil {
		respError := evalNotFoundError(err, "ServerSideEncryptionConfigurationNotFoundError")
		if respError == nil {
			return nil, nil
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, respError
	}

	return serializeBucketEncryptionConfiguration(rawResult), nil
}

// RemoveBucketEncryption remove default encryption configuration set on a bucket.
func (mc *Client) RemoveBucketEncryption(ctx context.Context, bucketName string) error {
	ctx, span := mc.startOtelSpan(ctx, "RemoveBucketEncryption", bucketName)
	defer span.End()

	err := mc.client.RemoveBucketEncryption(ctx, bucketName)
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
		rule.DeleteMarkerReplication.Status = replication.Status(
			item.DeleteMarkerReplication.Status,
		)
	}

	if item.DeleteReplication != nil && item.DeleteReplication.Status != "" {
		rule.DeleteReplication.Status = replication.Status(item.DeleteReplication.Status)
	}

	if item.ExistingObjectReplication != nil && item.ExistingObjectReplication.Status != "" {
		rule.ExistingObjectReplication.Status = replication.Status(
			item.ExistingObjectReplication.Status,
		)
	}

	if item.SourceSelectionCriteria != nil &&
		item.SourceSelectionCriteria.ReplicaModifications != nil &&
		item.SourceSelectionCriteria.ReplicaModifications.Status != "" {
		rule.SourceSelectionCriteria.ReplicaModifications.Status = replication.Status(
			item.SourceSelectionCriteria.ReplicaModifications.Status,
		)
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

	if len(input.Tag) > 0 {
		for _, item := range input.Tag {
			result.Tag.Key = item.Key
			result.Tag.Value = item.Value

			break
		}
	}

	if input.And != nil {
		if input.And.Prefix != nil {
			result.And.Prefix = *input.Prefix
		}

		for _, item := range input.And.Tags {
			t := replication.Tag{
				Key:   item.Key,
				Value: item.Value,
			}

			result.And.Tags = append(result.And.Tags, t)
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
		rule.DeleteMarkerReplication.Status = common.StorageReplicationRuleStatus(
			item.DeleteMarkerReplication.Status,
		)
	}

	if item.DeleteReplication.Status != "" {
		rule.DeleteReplication.Status = common.StorageReplicationRuleStatus(
			item.DeleteReplication.Status,
		)
	}

	if item.ExistingObjectReplication.Status != "" {
		rule.ExistingObjectReplication.Status = common.StorageReplicationRuleStatus(
			item.ExistingObjectReplication.Status,
		)
	}

	if item.SourceSelectionCriteria.ReplicaModifications.Status != "" {
		rule.SourceSelectionCriteria = &common.SourceSelectionCriteria{
			ReplicaModifications: &common.ReplicaModifications{
				Status: common.StorageReplicationRuleStatus(
					item.SourceSelectionCriteria.ReplicaModifications.Status,
				),
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
		rule.Filter.Tag = []common.StorageKeyValue{
			{
				Key:   item.Filter.Tag.Key,
				Value: item.Filter.Tag.Value,
			},
		}
	}

	if item.Filter.And.Prefix != "" || len(item.Filter.And.Tags) > 0 {
		rule.Filter.And = &common.StorageReplicationFilterAnd{}
		if item.Filter.And.Prefix != "" {
			rule.Filter.And.Prefix = &item.Filter.Prefix
		}

		for _, tag := range item.Filter.And.Tags {
			rule.Filter.And.Tags = append(rule.Filter.And.Tags, common.StorageKeyValue{
				Key:   tag.Key,
				Value: tag.Value,
			})
		}
	}

	return rule
}

func (mc *Client) populateBucket(
	ctx context.Context,
	item minio.BucketInfo,
	options common.BucketOptions,
) (common.StorageBucket, error) {
	bucket := common.StorageBucket{
		Name:         item.Name,
		CreationTime: &item.CreationDate,
	}

	if options.Include.Tags {
		tags, err := mc.GetBucketTagging(ctx, item.Name)
		if err != nil {
			return bucket, err
		}

		bucket.Tags = common.StringMapToKeyValues(tags)
	}

	if options.Include.Versioning {
		versioning, err := mc.GetBucketVersioning(ctx, bucket.Name)
		if err != nil {
			return bucket, err
		}

		bucket.Versioning = versioning
	}

	if options.Include.Lifecycle {
		lc, err := mc.GetBucketLifecycle(ctx, bucket.Name)
		if err != nil {
			return bucket, err
		}

		bucket.Lifecycle = lc
	}

	if options.Include.Encryption {
		encryption, err := mc.GetBucketEncryption(ctx, bucket.Name)
		if err != nil {
			return bucket, err
		}

		bucket.Encryption = encryption
	}

	if options.Include.ObjectLock {
		lock, err := mc.GetObjectLockConfig(ctx, bucket.Name)
		if err != nil {
			return bucket, err
		}

		bucket.ObjectLock = lock
	}

	return bucket, nil
}

func filterBuckets(
	bucketInfos []minio.BucketInfo,
	options *common.ListStorageBucketsOptions,
	predicate func(string) bool,
) ([]minio.BucketInfo, common.StoragePaginationInfo) {
	pageInfo := common.StoragePaginationInfo{}

	if len(bucketInfos) == 0 ||
		(options.Prefix == "" && predicate == nil && options.MaxResults == nil && options.StartAfter == "") {
		return bucketInfos, pageInfo
	}

	var count int

	filteredBuckets := make([]minio.BucketInfo, 0)
	started := options.StartAfter == ""
	bucketLength := len(bucketInfos)

	for i, info := range bucketInfos {
		if !started {
			started = options.StartAfter == info.Name

			continue
		}

		if (options.Prefix != "" && !strings.HasPrefix(info.Name, options.Prefix)) ||
			(predicate != nil && !predicate(info.Name)) {
			continue
		}

		filteredBuckets = append(filteredBuckets, info)
		count++

		if options.MaxResults != nil && count >= *options.MaxResults {
			pageInfo.HasNextPage = i < bucketLength-1

			break
		}
	}

	return filteredBuckets, pageInfo
}
