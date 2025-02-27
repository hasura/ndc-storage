package storage

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"time"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/azblob"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/storage/minio"
)

// RuntimeSettings hold runtime settings for the connector.
type RuntimeSettings struct {
	// Maximum size in MB of the object is allowed to download content in the GraphQL response
	// to avoid memory leaks. Pre-signed URLs are recommended for large files.
	MaxDownloadSizeMBs int64 `json:"maxDownloadSizeMBs" jsonschema:"min=1,default=20" yaml:"maxDownloadSizeMBs"`
}

// Manager represents the high-level client that manages internal clients and configurations.
type Manager struct {
	clients []Client
	runtime RuntimeSettings
	logger  *slog.Logger
}

// NewManager creates a storage client manager instance.
func NewManager(ctx context.Context, configs []ClientConfig, runtimeSettings RuntimeSettings, logger *slog.Logger) (*Manager, error) {
	result := &Manager{
		clients: make([]Client, len(configs)),
		runtime: runtimeSettings,
		logger:  logger,
	}

	for i, config := range configs {
		baseConfig, client, err := config.toStorageClient(ctx, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize storage client %d: %w", i, err)
		}

		configID := baseConfig.ID
		if configID == "" {
			configID = strconv.Itoa(i)
		}

		defaultBucket, err := baseConfig.DefaultBucket.GetOrDefault("")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize storage client %s; defaultBucket: %w", configID, err)
		}

		c := Client{
			id:             common.StorageClientID(baseConfig.ID),
			defaultBucket:  defaultBucket,
			allowedBuckets: baseConfig.AllowedBuckets,
			StorageClient:  client,
		}

		if baseConfig.DefaultPresignedExpiry != nil {
			presignedExpiry, err := time.ParseDuration(*baseConfig.DefaultPresignedExpiry)
			if err != nil {
				return nil, fmt.Errorf("failied to parse defaultPresignedExpiry in client %s: %w", configID, err)
			}

			c.defaultPresignedExpiry = &presignedExpiry
		}

		if c.id == "" {
			c.id = common.StorageClientID(strconv.Itoa(i))
		}

		result.clients[i] = c
	}

	return result, nil
}

// GetClient gets the inner client by key.
func (m *Manager) GetClient(clientID *common.StorageClientID) (*Client, bool) {
	if len(m.clients) == 0 {
		return nil, false
	}

	if clientID == nil || *clientID == "" {
		return &m.clients[0], true
	}

	for _, c := range m.clients {
		if c.id == *clientID {
			return &c, true
		}
	}

	return nil, false
}

// GetClientIDs gets all client IDs.
func (m *Manager) GetClientIDs() []string {
	results := make([]string, len(m.clients))

	for i, client := range m.clients {
		results[i] = string(client.id)
	}

	return results
}

// GetClient gets the inner client by key and bucket name.
func (m *Manager) GetClientAndBucket(ctx context.Context, arguments common.StorageBucketArguments) (*Client, string, error) {
	if len(m.clients) == 0 {
		if arguments.Bucket == "" {
			return nil, "", schema.UnprocessableContentError("bucket is required", nil)
		}

		client, err := m.createTemporaryClient(ctx, arguments)
		if err != nil {
			return nil, "", err
		}

		return client, arguments.Bucket, nil
	}

	hasClientID := arguments.ClientID != nil && *arguments.ClientID != ""
	if !hasClientID && arguments.Bucket == "" {
		client, _ := m.GetClient(nil)

		return client, client.defaultBucket, nil
	}

	if hasClientID {
		client, ok := m.GetClient(arguments.ClientID)
		if !ok {
			return nil, "", schema.InternalServerError("client not found: "+string(*arguments.ClientID), nil)
		}

		bucketName, err := client.ValidateBucket(arguments.Bucket)
		if err != nil {
			return nil, "", err
		}

		return client, bucketName, nil
	}

	for _, c := range m.clients {
		if c.defaultBucket == arguments.Bucket || slices.Contains(c.allowedBuckets, arguments.Bucket) {
			return &c, arguments.Bucket, nil
		}
	}

	// return the first client by default
	return &m.clients[0], arguments.Bucket, nil
}

func (m *Manager) createTemporaryClient(ctx context.Context, arguments common.StorageBucketArguments) (*Client, error) {
	clientType := common.S3

	if arguments.ClientType != nil && *arguments.ClientType != "" {
		if err := arguments.ClientType.Validate(); err != nil {
			return nil, err
		}

		clientType = *arguments.ClientType
	}

	clientId := common.StorageClientID(fmt.Sprintf("%s-temp", clientType))
	defaultPresignedExpiry := 24 * time.Hour

	switch clientType {
	case common.AzureBlobStore:
		if arguments.Endpoint == "" {
			return nil, schema.UnprocessableContentError("endpoint is required for azblob", nil)
		}

		clientConfig := &azblob.ClientConfig{
			BaseClientConfig: common.BaseClientConfig{
				ID: string(clientId),
			},
		}

		if arguments.AccessKeyID != "" || arguments.SecretAccessKey != "" {
			clientConfig.Endpoint = &utils.EnvString{
				Value: &arguments.Endpoint,
			}

			clientConfig.OtherConfig.Authentication = azblob.AuthCredentials{
				Type: azblob.AuthTypeSharedKey,
				AccountName: &utils.EnvString{
					Value: utils.ToPtr(arguments.AccessKeyID),
				},
				AccountKey: &utils.EnvString{
					Value: utils.ToPtr(arguments.SecretAccessKey),
				},
			}
		} else {
			clientConfig.OtherConfig.Authentication = azblob.AuthCredentials{
				Type: azblob.AuthTypeConnectionString,
				ConnectionString: &utils.EnvString{
					Value: utils.ToPtr(arguments.Endpoint),
				},
			}
		}

		client, err := azblob.New(ctx, clientConfig, m.logger)
		if err != nil {
			return nil, err
		}

		return &Client{
			id:                     clientId,
			StorageClient:          client,
			defaultPresignedExpiry: &defaultPresignedExpiry,
		}, nil
	default:
		if arguments.AccessKeyID == "" && arguments.SecretAccessKey == "" {
			return nil, schema.UnprocessableContentError("accessKeyId and secretAccessKey arguments are required", nil)
		}

		clientConfig := &minio.ClientConfig{
			BaseClientConfig: common.BaseClientConfig{
				ID: string(clientId),
			},
			OtherConfig: minio.OtherConfig{
				Authentication: minio.AuthCredentials{
					Type: minio.AuthTypeStatic,
					AccessKeyID: &utils.EnvString{
						Value: utils.ToPtr(arguments.AccessKeyID),
					},
					SecretAccessKey: &utils.EnvString{
						Value: utils.ToPtr(arguments.SecretAccessKey),
					},
				},
			},
		}

		if arguments.Endpoint != "" {
			clientConfig.BaseClientConfig.Endpoint = &utils.EnvString{
				Value: &arguments.Endpoint,
			}
		}

		client, err := minio.New(ctx, clientType, clientConfig, m.logger)
		if err != nil {
			return nil, err
		}

		return &Client{
			id:                     clientId,
			StorageClient:          client,
			defaultPresignedExpiry: &defaultPresignedExpiry,
		}, nil
	}
}
