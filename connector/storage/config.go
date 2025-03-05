package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/azblob"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/storage/gcs"
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

// ClientConfig abstracts a storage client configuration.
type ClientConfig map[string]any

// Validate validates the configuration.
func (cc ClientConfig) Validate() error {
	storageType, err := cc.getStorageType()
	if err != nil {
		return err
	}

	rawConfig, err := json.Marshal(cc)
	if err != nil {
		return err
	}

	switch storageType {
	case common.S3:
		var config minio.ClientConfig
		if err := json.Unmarshal(rawConfig, &config); err != nil {
			return err
		}

		return config.Validate()
	case common.GoogleStorage:
		var config gcs.ClientConfig
		if err := json.Unmarshal(rawConfig, &config); err != nil {
			return err
		}

		return config.Validate()
	case common.AzureBlobStore:
		var config azblob.ClientConfig
		if err := json.Unmarshal(rawConfig, &config); err != nil {
			return err
		}

		return config.Validate()
	}

	return errors.New("unsupported storage client: " + string(storageType))
}

func (cc ClientConfig) getStorageType() (common.StorageProviderType, error) {
	if len(cc) == 0 {
		return "", errConfigEmpty
	}

	rawStorageType, err := utils.GetString(cc, "type")
	if err != nil {
		return "", err
	}

	storageType, err := common.ParseStorageProviderType(rawStorageType)
	if err != nil {
		return "", err
	}

	return storageType, nil
}

func (cc ClientConfig) toStorageClient(ctx context.Context, logger *slog.Logger) (*common.BaseClientConfig, common.StorageClient, error) {
	storageType, err := cc.getStorageType()
	if err != nil {
		return nil, nil, err
	}

	rawConfig, err := json.Marshal(cc)
	if err != nil {
		return nil, nil, err
	}

	switch storageType {
	case common.S3:
		var config minio.ClientConfig
		if err := json.Unmarshal(rawConfig, &config); err != nil {
			return nil, nil, err
		}

		client, err := minio.New(ctx, config.Type, &config, logger)

		return &config.BaseClientConfig, client, err
	case common.GoogleStorage:
		var config gcs.ClientConfig
		if err := json.Unmarshal(rawConfig, &config); err != nil {
			return nil, nil, err
		}

		client, err := gcs.New(ctx, &config, logger)

		return &config.BaseClientConfig, client, err
	case common.AzureBlobStore:
		var azConfig azblob.ClientConfig
		if err := json.Unmarshal(rawConfig, &azConfig); err != nil {
			return nil, nil, err
		}

		client, err := azblob.New(ctx, &azConfig, logger)

		return &azConfig.BaseClientConfig, client, err
	}

	return nil, nil, errors.New("unsupported storage client: " + string(storageType))
}

// JSONSchema is used to generate a custom jsonschema.
func (cc ClientConfig) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			minio.ClientConfig{}.JSONSchema(),
			azblob.ClientConfig{}.JSONSchema(),
			gcs.ClientConfig{}.JSONSchema(),
		},
	}
}
