package azblob

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
	"github.com/hasura/ndc-sdk-go/v2/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// MakeBucket creates a new bucket.
func (c *Client) MakeBucket(ctx context.Context, args *common.MakeStorageBucketOptions) error {
	ctx, span := c.startOtelSpan(ctx, "MakeBucket", args.Name)
	defer span.End()

	options := &azblob.CreateContainerOptions{
		Metadata: map[string]*string{},
	}

	for _, item := range args.Tags {
		options.Metadata[item.Key] = &item.Value
	}

	_, err := c.client.CreateContainer(ctx, args.Name, options)
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

	opts := &azblob.ListContainersOptions{
		Include: azblob.ListContainersInclude{
			Metadata: options.Include.Tags,
			Deleted:  false,
		},
	}

	if options.Prefix != "" {
		opts.Prefix = &options.Prefix
	}

	var maxResults int32
	if options.MaxResults != nil && *options.MaxResults > 0 && predicate == nil {
		maxResults = int32(*options.MaxResults)
		opts.MaxResults = &maxResults

		span.SetAttributes(attribute.Int("storage.options.max_results", int(maxResults)))
	}

	if options.StartAfter != "" {
		opts.Marker = &options.StartAfter
	}

	pager := c.client.NewListContainersPager(opts)

	var count int32

	var results []common.StorageBucket

	pageInfo := common.StoragePaginationInfo{}

L:
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, serializeErrorResponse(err)
		}

		for i, container := range resp.ContainerItems {
			if container.Name == nil || (predicate != nil && !predicate(*container.Name)) {
				continue
			}

			result := common.StorageBucket{}

			if container.Name != nil {
				result.Name = *container.Name
			}

			for key, value := range container.Metadata {
				if value != nil {
					result.Tags = append(result.Tags, common.StorageKeyValue{
						Key:   key,
						Value: *value,
					})
				}
			}

			if container.Properties != nil {
				result.LastModified = container.Properties.LastModified

				if container.Properties.IsImmutableStorageWithVersioningEnabled != nil {
					result.Versioning = &common.StorageBucketVersioningConfiguration{
						Enabled: *container.Properties.IsImmutableStorageWithVersioningEnabled,
					}
				}

				if container.Properties.DefaultEncryptionScope != nil {
					result.Encryption = &common.ServerSideEncryptionConfiguration{
						KmsMasterKeyID: *container.Properties.DefaultEncryptionScope,
					}
				}

				if container.Properties.RemainingRetentionDays != nil && container.Properties.HasImmutabilityPolicy != nil {
					mode := common.StorageRetentionModeLocked
					unit := common.StorageRetentionValidityUnitDays
					days := uint(*container.Properties.RemainingRetentionDays)

					result.ObjectLock = &common.StorageObjectLockConfig{
						Enabled: true,
						SetStorageObjectLockConfig: common.SetStorageObjectLockConfig{
							Mode:     &mode,
							Unit:     &unit,
							Validity: &days,
						},
					}
				}
			}

			results = append(results, result)
			count++

			if maxResults > 0 && count >= maxResults {
				pageInfo.HasNextPage = i < len(resp.ContainerItems)-1 || pager.More()

				break L
			}
		}
	}

	span.SetAttributes(attribute.Int("storage.bucket_count", len(results)))

	return &common.StorageBucketListResults{
		Buckets:  results,
		PageInfo: pageInfo,
	}, nil
}

// BucketExists checks if a bucket exists.
func (c *Client) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	ctx, span := c.startOtelSpan(ctx, "BucketExists", bucketName)
	defer span.End()

	result, err := c.getBucket(ctx, bucketName, common.BucketOptions{})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return false, err
	}

	span.SetAttributes(attribute.Bool("storage.bucket_exist", result != nil))

	return result != nil, nil
}

// GetBucket gets a bucket by name.
func (c *Client) GetBucket(
	ctx context.Context,
	bucketName string,
	options common.BucketOptions,
) (*common.StorageBucket, error) {
	ctx, span := c.startOtelSpan(ctx, "GetBucket", bucketName)
	defer span.End()

	result, err := c.getBucket(ctx, bucketName, options)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, err
	}

	return result, nil
}

func (c *Client) getBucket(
	ctx context.Context,
	bucketName string,
	options common.BucketOptions,
) (*common.StorageBucket, error) {
	pager := c.client.NewListContainersPager(&service.ListContainersOptions{
		Prefix: &bucketName,
		Include: service.ListContainersInclude{
			Metadata: options.Include.Tags,
			Deleted:  false,
			System:   false,
		},
	})

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, serializeErrorResponse(err)
		}

		for _, container := range resp.ContainerItems {
			if container.Name == nil || *container.Name != bucketName {
				continue
			}

			result := common.StorageBucket{}

			if container.Name != nil {
				result.Name = *container.Name
			}

			for key, value := range container.Metadata {
				if value != nil {
					result.Tags = append(result.Tags, common.StorageKeyValue{
						Key:   key,
						Value: *value,
					})
				}
			}

			if container.Properties != nil && container.Properties.LastModified != nil {
				result.LastModified = container.Properties.LastModified
			}

			return &result, nil
		}
	}

	return nil, nil
}

// UpdateBucket updates configurations for the bucket.
func (c *Client) UpdateBucket(
	ctx context.Context,
	bucketName string,
	opts common.UpdateStorageBucketOptions,
) error {
	ctx, span := c.startOtelSpan(ctx, "UpdateBucket", bucketName)
	defer span.End()

	if opts.Tags != nil {
		if err := c.SetBucketTagging(ctx, bucketName, common.KeyValuesToStringMap(*opts.Tags)); err != nil {
			return err
		}
	}

	return nil
}

// RemoveBucket removes a bucket, bucket should be empty to be successfully removed.
func (c *Client) RemoveBucket(ctx context.Context, bucketName string) error {
	ctx, span := c.startOtelSpan(ctx, "RemoveBucket", bucketName)
	defer span.End()

	_, err := c.client.DeleteContainer(ctx, bucketName, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if respErr.ErrorCode == string(bloberror.ContainerBeingDeleted) ||
				respErr.ErrorCode == string(bloberror.ContainerNotFound) {
				return nil
			}

			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return serializeAzureErrorResponse(respErr)
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return schema.UnprocessableContentError(err.Error(), nil)
	}

	return nil
}

// SetBucketTagging sets tags to a bucket.
func (c *Client) SetBucketTagging(
	ctx context.Context,
	bucketName string,
	bucketTags map[string]string,
) error {
	ctx, span := c.startOtelSpan(ctx, "SetBucketTagging", bucketName)
	defer span.End()

	var inputTags map[string]*string
	if len(bucketTags) > 0 {
		inputTags = map[string]*string{}

		for key, value := range bucketTags {
			span.SetAttributes(attribute.String("storage.bucket_tag"+key, value))
			inputTags[key] = &value
		}
	}

	containerClient := c.client.ServiceClient().NewContainerClient(bucketName)

	_, err := containerClient.SetMetadata(ctx, &container.SetMetadataOptions{
		Metadata: inputTags,
	})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// GetBucketPolicy gets access permissions on a bucket or a prefix.
func (c *Client) GetBucketPolicy(ctx context.Context, bucketName string) (string, error) {
	ctx, span := c.startOtelSpan(ctx, "SetBucketTagging", bucketName)
	defer span.End()

	containerClient := c.client.ServiceClient().NewContainerClient(bucketName)

	resp, err := containerClient.GetAccessPolicy(ctx, &container.GetAccessPolicyOptions{})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", serializeErrorResponse(err)
	}

	result := ""

	if resp.BlobPublicAccess != nil {
		result = string(*resp.BlobPublicAccess)
	}

	return result, nil
}
