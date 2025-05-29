package gcs

import (
	"context"
	"errors"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/api/iterator"
)

// MakeBucket creates a new bucket.
func (c *Client) MakeBucket(ctx context.Context, args *common.MakeStorageBucketOptions) error {
	ctx, span := c.startOtelSpan(ctx, "MakeBucket", args.Name)
	defer span.End()

	attrs := &storage.BucketAttrs{
		Location: args.Region,
		Labels:   common.KeyValuesToStringMap(args.Tags),
		Name:     args.Name,
	}

	handle := c.client.Bucket(args.Name)

	if args.ObjectLock {
		handle = handle.SetObjectRetention(true)
	}

	err := handle.Create(ctx, c.projectID, attrs)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// ListBuckets lists all buckets.
func (c *Client) ListBuckets(
	ctx context.Context,
	options *common.ListStorageBucketsOptions,
	predicate func(string) bool,
) (*common.StorageBucketListResults, error) {
	ctx, span := c.startOtelSpan(ctx, "ListBuckets", "")
	defer span.End()

	pager := c.client.Buckets(ctx, c.projectID)
	pager.Prefix = options.Prefix

	var count, maxResults int

	var results []common.StorageBucket

	pageInfo := common.StoragePaginationInfo{}

	if options.MaxResults != nil {
		maxResults = *options.MaxResults
	}

	started := options.StartAfter == ""

	for {
		bucket, err := pager.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}

			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, serializeErrorResponse(err)
		}

		if !started {
			started = options.StartAfter == bucket.Name

			continue
		}

		if (options.Prefix != "" && !strings.HasPrefix(bucket.Name, options.Prefix)) ||
			(predicate != nil && !predicate(bucket.Name)) {
			continue
		}

		result := serializeBucketInfo(bucket)
		results = append(results, result)
		count++

		if maxResults > 0 && count >= maxResults {
			pageInfo.HasNextPage = pager.PageInfo().Remaining() > 0

			break
		}
	}

	span.SetAttributes(attribute.Int("storage.bucket_count", len(results)))

	return &common.StorageBucketListResults{
		Buckets:  results,
		PageInfo: pageInfo,
	}, nil
}

// GetBucket gets a bucket by name.
func (c *Client) GetBucket(
	ctx context.Context,
	name string,
	options common.BucketOptions,
) (*common.StorageBucket, error) {
	ctx, span := c.startOtelSpan(ctx, "GetBucket", name)
	defer span.End()

	bucketInfo, err := c.client.Bucket(name).Attrs(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	result := serializeBucketInfo(bucketInfo)

	return &result, nil
}

// BucketExists checks if a bucket exists.
func (c *Client) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	ctx, span := c.startOtelSpan(ctx, "BucketExists", bucketName)
	defer span.End()

	result, err := c.client.Bucket(bucketName).Attrs(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return false, serializeErrorResponse(err)
	}

	existed := result != nil
	span.SetAttributes(attribute.Bool("storage.bucket_exist", existed))

	return existed, nil
}

// UpdateBucket updates configurations for the bucket.
func (c *Client) UpdateBucket(
	ctx context.Context,
	bucketName string,
	opts common.UpdateStorageBucketOptions,
) error {
	ctx, span := c.startOtelSpan(ctx, "UpdateBucket", bucketName)
	defer span.End()

	attrs, err := c.client.Bucket(bucketName).Attrs(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	if attrs == nil {
		return schema.UnprocessableContentError("bucket does not exist", nil)
	}

	inputAttrs := storage.BucketAttrsToUpdate{}

	if opts.Tags != nil {
		tagKeys := map[string]bool{}
		for _, item := range *opts.Tags {
			tagKeys[item.Key] = true

			span.SetAttributes(attribute.String("storage.bucket_tag"+item.Key, item.Value))
		}

		for key := range attrs.Labels {
			if _, ok := tagKeys[key]; !ok {
				inputAttrs.DeleteLabel(key)
			}
		}

		for _, item := range *opts.Tags {
			inputAttrs.SetLabel(item.Key, item.Value)
		}
	}

	if opts.VersioningEnabled != nil {
		inputAttrs.VersioningEnabled = *opts.VersioningEnabled
	}

	if opts.Encryption != nil {
		inputAttrs.Encryption = &storage.BucketEncryption{
			DefaultKMSKeyName: opts.Encryption.KmsMasterKeyID,
		}
	}

	if opts.Lifecycle != nil {
		lc := validateLifecycleConfiguration(*opts.Lifecycle)
		inputAttrs.Lifecycle = lc
	}

	_, err = c.client.Bucket(bucketName).Update(ctx, inputAttrs)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// RemoveBucket removes a bucket, bucket should be empty to be successfully removed.
func (c *Client) RemoveBucket(ctx context.Context, bucketName string) error {
	ctx, span := c.startOtelSpan(ctx, "RemoveBucket", bucketName)
	defer span.End()

	err := c.client.Bucket(bucketName).Delete(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}
