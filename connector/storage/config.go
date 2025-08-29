package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/hasura/ndc-http/exhttp"
	"github.com/hasura/ndc-sdk-go/v2/schema"
	"github.com/hasura/ndc-sdk-go/v2/utils"
	"github.com/hasura/ndc-storage/connector/storage/azblob"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/storage/fs"
	"github.com/hasura/ndc-storage/connector/storage/gcs"
	"github.com/hasura/ndc-storage/connector/storage/minio"
	"github.com/invopop/jsonschema"
)

var errConfigEmpty = errors.New("the configuration is empty")

// Client wraps the storage client with additional information.
type Client struct {
	common.StorageClient

	id                     common.StorageClientID
	defaultBucket          string
	defaultPresignedExpiry *time.Duration
	allowedBuckets         []string
}

// ValidateBucket checks if the bucket name is valid, or returns the default bucket if empty.
func (c *Client) ValidateBucket(key string) (string, error) {
	if key != "" {
		if key == c.defaultBucket || len(c.allowedBuckets) == 0 ||
			slices.Contains(c.allowedBuckets, key) {
			return key, nil
		}

		return "", schema.UnprocessableContentError(
			fmt.Sprintf("you are not allowed to access `%s` bucket, client id `%s`", key, c.id),
			nil,
		)
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
	case common.StorageProviderTypeS3:
		var config minio.ClientConfig

		err := json.Unmarshal(rawConfig, &config)
		if err != nil {
			return err
		}

		return config.Validate()
	case common.StorageProviderTypeGcs:
		var config gcs.ClientConfig

		err := json.Unmarshal(rawConfig, &config)
		if err != nil {
			return err
		}

		return config.Validate()
	case common.StorageProviderTypeAzblob:
		var config azblob.ClientConfig

		err := json.Unmarshal(rawConfig, &config)
		if err != nil {
			return err
		}

		return config.Validate()
	case common.StorageProviderTypeFs:
		var config fs.ClientConfig

		err := json.Unmarshal(rawConfig, &config)
		if err != nil {
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

// ToStorageClient initializes a storage client from the current config.
func (cc ClientConfig) ToStorageClient(
	ctx context.Context,
	logger *slog.Logger,
) (*common.BaseClientConfig, common.StorageClient, error) {
	storageType, err := cc.getStorageType()
	if err != nil {
		return nil, nil, err
	}

	rawConfig, err := json.Marshal(cc)
	if err != nil {
		return nil, nil, err
	}

	switch storageType {
	case common.StorageProviderTypeS3:
		var config minio.ClientConfig
		if err := json.Unmarshal(rawConfig, &config); err != nil {
			return nil, nil, err
		}

		client, err := minio.New(ctx, config.Type, &config, logger)

		return &config.BaseClientConfig, client, err
	case common.StorageProviderTypeGcs:
		var config gcs.ClientConfig
		if err := json.Unmarshal(rawConfig, &config); err != nil {
			return nil, nil, err
		}

		client, err := gcs.New(ctx, &config, logger)

		return &config.BaseClientConfig, client, err
	case common.StorageProviderTypeAzblob:
		var azConfig azblob.ClientConfig
		if err := json.Unmarshal(rawConfig, &azConfig); err != nil {
			return nil, nil, err
		}

		client, err := azblob.New(ctx, &azConfig, logger)

		return &azConfig.BaseClientConfig, client, err
	case common.StorageProviderTypeFs:
		var fsConfig fs.ClientConfig
		if err := json.Unmarshal(rawConfig, &fsConfig); err != nil {
			return nil, nil, err
		}

		client, err := fs.NewOSFileSystem(&fsConfig)

		return fsConfig.ToBaseConfig(), client, err
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
			fs.ClientConfig{}.JSONSchema(),
		},
	}
}

// RuntimeSettings hold runtime settings for the connector.
type RuntimeSettings struct {
	// Maximum size in MB of the object is allowed to download the content in the GraphQL response
	// to avoid memory leaks. Pre-signed URLs are recommended for large files.
	MaxDownloadSizeMBs int64 `json:"maxDownloadSizeMBs" jsonschema:"min=1,default=20" yaml:"maxDownloadSizeMBs"`
	// Maximum size in MB of the object is allowed to upload the content from HTTP URL
	// to avoid memory leaks. Pre-signed URLs are recommended for large files.
	MaxUploadSizeMBs int64 `json:"maxUploadSizeMBs"   jsonschema:"min=1,default=20" yaml:"maxUploadSizeMBs"`
	// Configuration for the http client that is used for uploading files from URL.
	HTTP *exhttp.HTTPTransportTLSConfig `json:"http,omitempty"                                   yaml:"http"`
}
