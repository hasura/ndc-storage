package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/azblob"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/storage/minio"
	"github.com/invopop/jsonschema"
)

var errConfigEmpty = errors.New("the configuration is empty")

// Client wraps the storage client with additional information.
type Client struct {
	id                     common.StorageClientID
	defaultBucket          string
	defaultPresignedExpiry *time.Duration
	allowedBuckets         []string

	common.StorageClient
}

// ValidateBucket checks if the bucket name is valid, or returns the default bucket if empty.
func (c *Client) ValidateBucket(key string) (string, error) {
	if key != "" {
		if key == c.defaultBucket || len(c.allowedBuckets) == 0 || slices.Contains(c.allowedBuckets, key) {
			return key, nil
		}

		return "", schema.UnprocessableContentError(fmt.Sprintf("you are not allowed to access `%s` bucket, client id `%s`", key, c.id), nil)
	}

	if c.defaultBucket == "" {
		return "", schema.UnprocessableContentError("bucket name is required", nil)
	}

	return c.defaultBucket, nil
}

// FormatTimestamp formats the Time value to string
func FormatTimestamp(value time.Time) string {
	return value.Format(time.RFC3339)
}

// ClientConfig abstracts a storage client configuration.
type ClientConfig map[string]any

// Validate validates the configuration.
func (cc ClientConfig) Validate() error {
	if len(cc) == 0 {
		return errConfigEmpty
	}

	var baseConfig common.BaseClientConfig
	if err := mapstructure.Decode(cc, &baseConfig); err != nil {
		return err
	}

	if err := baseConfig.Validate(); err != nil {
		return err
	}

	err := baseConfig.Type.Validate()
	if err != nil {
		return err
	}

	switch baseConfig.Type {
	case common.S3, common.GoogleStorage:
		minioConfig := minio.ClientConfig{}

		if err := mapstructure.Decode(cc, &minioConfig.OtherConfig); err != nil {
			return err
		}

		return nil
	case common.AzureBlobStore:
		azConfig := azblob.ClientConfig{}

		if err := mapstructure.Decode(cc, &azConfig.OtherConfig); err != nil {
			return err
		}

		return nil
	}

	return errors.New("unsupported storage client: " + string(baseConfig.Type))
}

func (cc ClientConfig) toStorageClient(ctx context.Context, logger *slog.Logger) (*common.BaseClientConfig, common.StorageClient, error) {
	if len(cc) == 0 {
		return nil, nil, errConfigEmpty
	}

	var baseConfig common.BaseClientConfig
	if err := mapstructure.Decode(cc, &baseConfig); err != nil {
		return nil, nil, err
	}

	err := baseConfig.Type.Validate()
	if err != nil {
		return nil, nil, err
	}

	switch baseConfig.Type {
	case common.S3, common.GoogleStorage:
		minioConfig := minio.ClientConfig{
			BaseClientConfig: baseConfig,
		}

		if err := mapstructure.Decode(cc, &minioConfig.OtherConfig); err != nil {
			return nil, nil, err
		}

		client, err := minio.New(ctx, baseConfig.Type, &minioConfig, logger)

		return &baseConfig, client, err
	case common.AzureBlobStore:
		azConfig := azblob.ClientConfig{
			BaseClientConfig: baseConfig,
		}
		if err := mapstructure.Decode(cc, &azConfig.OtherConfig); err != nil {
			return nil, nil, err
		}

		client, err := azblob.New(ctx, &azConfig, logger)

		return &baseConfig, client, err
	}

	return nil, nil, errors.New("unsupported storage client: " + string(baseConfig.Type))
}

// JSONSchema is used to generate a custom jsonschema.
func (cc ClientConfig) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			minio.ClientConfig{}.JSONSchema(),
			azblob.ClientConfig{}.JSONSchema(),
		},
	}
}
