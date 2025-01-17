package azblob

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
	"github.com/hasura/ndc-sdk-go/schema"
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

	for key, value := range args.Tags {
		options.Metadata[key] = &value
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
func (c *Client) ListBuckets(ctx context.Context, options common.BucketOptions) ([]common.StorageBucketInfo, error) {
	ctx, span := c.startOtelSpan(ctx, "ListBuckets", "")
	defer span.End()

	pager := c.client.NewListContainersPager(&azblob.ListContainersOptions{
		Include: azblob.ListContainersInclude{
			Metadata: options.IncludeTags,
			Deleted:  false,
		},
	})

	var results []common.StorageBucketInfo

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, serializeErrorResponse(err)
		}

		for _, container := range resp.ContainerItems {
			result := common.StorageBucketInfo{}

			if container.Name != nil {
				result.Name = *container.Name
			}

			if container.Metadata != nil {
				result.Tags = map[string]string{}

				for key, value := range container.Metadata {
					if value != nil {
						result.Tags[key] = *value
					}
				}
			}

			if container.Properties != nil && container.Properties.LastModified != nil {
				result.CreationDate = *container.Properties.LastModified
			}

			results = append(results, result)
		}
	}

	span.SetAttributes(attribute.Int("storage.bucket_count", len(results)))

	return results, nil
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
func (c *Client) GetBucket(ctx context.Context, bucketName string, options common.BucketOptions) (*common.StorageBucketInfo, error) {
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

func (c *Client) getBucket(ctx context.Context, bucketName string, options common.BucketOptions) (*common.StorageBucketInfo, error) {
	pager := c.client.NewListContainersPager(&service.ListContainersOptions{
		Prefix: &bucketName,
		Include: service.ListContainersInclude{
			Metadata: options.IncludeTags,
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

			result := common.StorageBucketInfo{}

			if container.Name != nil {
				result.Name = *container.Name
			}

			if container.Metadata != nil {
				result.Tags = map[string]string{}

				for key, value := range container.Metadata {
					if value != nil {
						result.Tags[key] = *value
					}
				}
			}

			if container.Properties != nil && container.Properties.LastModified != nil {
				result.CreationDate = *container.Properties.LastModified
			}

			return &result, nil
		}
	}

	return nil, nil
}

// RemoveBucket removes a bucket, bucket should be empty to be successfully removed.
func (c *Client) RemoveBucket(ctx context.Context, bucketName string) error {
	ctx, span := c.startOtelSpan(ctx, "RemoveBucket", bucketName)
	defer span.End()

	_, err := c.client.DeleteContainer(ctx, bucketName, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if respErr.ErrorCode == string(bloberror.ContainerBeingDeleted) || respErr.ErrorCode == string(bloberror.ContainerNotFound) {
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
func (c *Client) SetBucketTagging(ctx context.Context, bucketName string, bucketTags map[string]string) error {
	ctx, span := c.startOtelSpan(ctx, "SetBucketTagging", bucketName)
	defer span.End()

	var inputTags map[string]*string
	if len(bucketTags) > 0 {
		inputTags = map[string]*string{}

		for key, value := range bucketTags {
			if value == "" {
				continue
			}

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

// GetBucketNotification gets notification configuration on a bucket.
func (c *Client) GetBucketNotification(ctx context.Context, bucketName string) (*common.NotificationConfig, error) {
	return nil, errNotSupported
}

// SetBucketNotification sets a new bucket notification on a bucket.
func (c *Client) SetBucketNotification(ctx context.Context, bucketName string, config common.NotificationConfig) error {
	return errNotSupported
}

// RemoveAllBucketNotification removes all configured bucket notifications on a bucket.
func (c *Client) RemoveAllBucketNotification(ctx context.Context, bucketName string) error {
	return errNotSupported
}

// GetBucketVersioning gets the versioning configuration set on a bucket.
func (c *Client) GetBucketVersioning(ctx context.Context, bucketName string) (*common.StorageBucketVersioningConfiguration, error) {
	return nil, errNotSupported
}

// SetBucketReplication sets replication configuration on a bucket. Role can be obtained by first defining the replication target
// to associate the source and destination buckets for replication with the replication endpoint.
func (c *Client) SetBucketReplication(ctx context.Context, bucketName string, cfg common.StorageReplicationConfig) error {
	return errNotSupported
}

// Get current replication config on a bucket.
func (c *Client) GetBucketReplication(ctx context.Context, bucketName string) (*common.StorageReplicationConfig, error) {
	return nil, errNotSupported
}

// RemoveBucketReplication removes replication configuration on a bucket.
func (c *Client) RemoveBucketReplication(ctx context.Context, bucketName string) error {
	return errNotSupported
}

// EnableVersioning enables bucket versioning support.
func (c *Client) EnableVersioning(ctx context.Context, bucketName string) error {
	return errNotSupported
}

// SuspendVersioning disables bucket versioning support.
func (c *Client) SuspendVersioning(ctx context.Context, bucketName string) error {
	return errNotSupported
}

// SetBucketLifecycle sets lifecycle on bucket or an object prefix.
func (c *Client) SetBucketLifecycle(ctx context.Context, bucketName string, config common.BucketLifecycleConfiguration) error {
	return errNotSupported
}

// GetBucketLifecycle gets lifecycle on a bucket or a prefix.
func (c *Client) GetBucketLifecycle(ctx context.Context, bucketName string) (*common.BucketLifecycleConfiguration, error) {
	return nil, errNotSupported
}

// SetBucketEncryption sets default encryption configuration on a bucket.
func (c *Client) SetBucketEncryption(ctx context.Context, bucketName string, input common.ServerSideEncryptionConfiguration) error {
	return errNotSupported
}

// GetBucketEncryption gets default encryption configuration set on a bucket.
func (c *Client) GetBucketEncryption(ctx context.Context, bucketName string) (*common.ServerSideEncryptionConfiguration, error) {
	return nil, errNotSupported
}

// RemoveBucketEncryption remove default encryption configuration set on a bucket.
func (c *Client) RemoveBucketEncryption(ctx context.Context, bucketName string) error {
	return errNotSupported
}
