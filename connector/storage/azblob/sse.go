package azblob

import (
	"context"

	"github.com/hasura/ndc-storage/connector/storage/common"
)

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
