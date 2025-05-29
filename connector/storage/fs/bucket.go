package fs

import (
	"context"
	"errors"
	"os"
	"slices"
	"strings"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/spf13/afero"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// MakeBucket creates a new bucket.
func (c *Client) MakeBucket(ctx context.Context, args *common.MakeStorageBucketOptions) error {
	_, span := c.startOtelSpan(ctx, "MakeBucket", args.Name)
	defer span.End()

	err := c.client.MkdirAll(args.Name, os.FileMode(c.permissions.Directory))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return schema.UnauthorizeError(err.Error(), nil)
	}

	return nil
}

// ListBuckets lists all buckets.
func (c *Client) ListBuckets(
	ctx context.Context,
	options *common.ListStorageBucketsOptions,
	predicate func(string) bool,
) (*common.StorageBucketListResults, error) {
	_, span := c.startOtelSpan(ctx, "ListBuckets", "")
	defer span.End()

	result := &common.StorageBucketListResults{
		Buckets: make([]common.StorageBucket, 0),
	}

	count, index := 0, 0
	started := options.StartAfter == ""
	total := len(c.allowedDirectories)

	for ; index < total; index++ {
		dir := c.allowedDirectories[index]

		if !started {
			if options.StartAfter == dir {
				started = true
			}

			continue
		}

		if (options.Prefix != "" && !strings.HasPrefix(dir, options.Prefix)) ||
			(predicate != nil && !predicate(dir)) {
			continue
		}

		item := common.StorageBucket{
			Name: dir,
		}

		result.Buckets = append(result.Buckets, item)
		count++

		if options.MaxResults != nil && count >= *options.MaxResults {
			break
		}
	}

	result.PageInfo.HasNextPage = index < total

	span.SetAttributes(attribute.Int("storage.bucket_count", len(result.Buckets)))

	return result, nil
}

// GetBucket gets a bucket by name.
func (c *Client) GetBucket(
	ctx context.Context,
	name string,
	options common.BucketOptions,
) (*common.StorageBucket, error) {
	_, span := c.startOtelSpan(ctx, "GetBucket", name)
	defer span.End()

	if !slices.Contains(c.allowedDirectories, name) {
		return nil, nil
	}

	return &common.StorageBucket{
		Name: name,
	}, nil
}

// BucketExists checks if a bucket exists.
func (c *Client) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	_, span := c.startOtelSpan(ctx, "BucketExists", bucketName)
	defer span.End()

	existed := slices.Contains(c.allowedDirectories, bucketName)
	span.SetAttributes(attribute.Bool("storage.bucket_exist", existed))

	return existed, nil
}

// UpdateBucket updates configurations for the bucket.
func (c *Client) UpdateBucket(
	ctx context.Context,
	bucketName string,
	opts common.UpdateStorageBucketOptions,
) error {
	_, span := c.startOtelSpan(ctx, "UpdateBucket", bucketName)
	defer span.End()

	return nil
}

// RemoveBucket removes a bucket, bucket should be empty to be successfully removed.
func (c *Client) RemoveBucket(ctx context.Context, bucketName string) error {
	_, span := c.startOtelSpan(ctx, "RemoveBucket", bucketName)
	defer span.End()

	dirInfo, err := c.lstatIfPossible(bucketName)
	if err != nil {
		if errors.Is(err, afero.ErrFileNotFound) {
			return nil
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return schema.UnprocessableContentError(err.Error(), nil)
	}

	if !dirInfo.IsDir() {
		err := errors.New("the bucket path must be a directory")
		span.SetStatus(codes.Error, err.Error())

		return schema.UnprocessableContentError(err.Error(), nil)
	}

	err = c.client.RemoveAll(bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return schema.UnprocessableContentError(err.Error(), nil)
	}

	return nil
}
