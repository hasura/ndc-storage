package storage

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"time"

	"github.com/hasura/ndc-sdk-go/v2/connector"
	"github.com/hasura/ndc-sdk-go/v2/schema"
	"github.com/hasura/ndc-sdk-go/v2/utils"
	"github.com/hasura/ndc-storage/connector/storage/azblob"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/storage/minio"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = connector.NewTracer("connector/storage")

// Manager represents the high-level client that manages internal clients and configurations.
type Manager struct {
	clients    []Client
	httpClient *common.HTTPClient
	runtime    RuntimeSettings
	logger     *slog.Logger
}

// NewManager creates a storage client manager instance.
func NewManager(
	ctx context.Context,
	configs []ClientConfig,
	runtimeSettings RuntimeSettings,
	logger *slog.Logger,
) (*Manager, error) {
	httpClient, err := common.NewHTTPClient(runtimeSettings.HTTP, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the http client: %w", err)
	}

	result := &Manager{
		clients:    make([]Client, len(configs)),
		httpClient: httpClient,
		runtime:    runtimeSettings,
		logger:     logger,
	}

	for i, config := range configs {
		baseConfig, client, err := config.ToStorageClient(ctx, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize storage client %d: %w", i, err)
		}

		configID := baseConfig.ID
		if configID == "" {
			configID = strconv.Itoa(i)
		}

		defaultBucket, err := baseConfig.DefaultBucket.GetOrDefault("")
		if err != nil {
			return nil, fmt.Errorf(
				"failed to initialize storage client %s; defaultBucket: %w",
				configID,
				err,
			)
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
				return nil, fmt.Errorf(
					"failed to parse defaultPresignedExpiry in client %s: %w",
					configID,
					err,
				)
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

func (m *Manager) GetOrCreateClient(
	ctx context.Context,
	arguments common.StorageClientCredentialArguments,
) (*Client, error) {
	if len(m.clients) == 0 || !arguments.IsEmpty() {
		return m.createTemporaryClient(ctx, arguments)
	}

	client, ok := m.GetClient(arguments.ClientID)
	if !ok {
		return nil, nil
	}

	return client, nil
}

// GetClient gets the inner client by key and bucket name.
func (m *Manager) GetClientAndBucket(
	ctx context.Context,
	arguments common.StorageBucketArguments,
) (*Client, string, error) {
	if len(m.clients) == 0 || !arguments.IsEmpty() {
		if arguments.Bucket == "" {
			return nil, "", schema.UnprocessableContentError("bucket is required", nil)
		}

		client, err := m.createTemporaryClient(ctx, arguments.StorageClientCredentialArguments)
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
			return nil, "", schema.InternalServerError(
				"client not found: "+string(*arguments.ClientID),
				nil,
			)
		}

		bucketName, err := client.ValidateBucket(arguments.Bucket)
		if err != nil {
			return nil, "", err
		}

		return client, bucketName, nil
	}

	for _, c := range m.clients {
		if c.defaultBucket == arguments.Bucket ||
			slices.Contains(c.allowedBuckets, arguments.Bucket) {
			return &c, arguments.Bucket, nil
		}
	}

	// return the first client by default
	return &m.clients[0], arguments.Bucket, nil
}

func (m *Manager) createTemporaryClient(
	ctx context.Context,
	arguments common.StorageClientCredentialArguments,
) (*Client, error) {
	_, span := tracer.Start(ctx, "createTemporaryClient")
	defer span.End()

	clientType := common.StorageProviderTypeS3

	if arguments.ClientType != nil && *arguments.ClientType != "" {
		clientType = *arguments.ClientType
	}

	clientId := common.StorageClientID(fmt.Sprintf("%s-temp", clientType))
	defaultPresignedExpiry := 24 * time.Hour

	span.SetAttributes(attribute.String("storage.client.type", string(clientType)))

	switch clientType {
	case common.StorageProviderTypeAzblob:
		return m.createTemporaryAzblobClient(ctx, arguments)
	case common.StorageProviderTypeS3, common.StorageProviderTypeGcs:
		fallthrough
	default:
		if arguments.AccessKeyID == "" && arguments.SecretAccessKey == "" {
			return nil, schema.UnprocessableContentError(
				"accessKeyId and secretAccessKey arguments are required",
				nil,
			)
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
				HTTP: m.runtime.HTTP,
			},
		}

		if arguments.Endpoint != "" {
			clientConfig.Endpoint = &utils.EnvString{
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

func (m *Manager) createTemporaryAzblobClient(
	ctx context.Context,
	arguments common.StorageClientCredentialArguments,
) (*Client, error) {
	if arguments.Endpoint == "" {
		return nil, schema.UnprocessableContentError("endpoint is required for azblob", nil)
	}

	clientId := "azblob-temp"
	defaultPresignedExpiry := 24 * time.Hour
	clientConfig := &azblob.ClientConfig{
		BaseClientConfig: common.BaseClientConfig{
			ID: clientId,
		},
		OtherConfig: azblob.OtherConfig{
			HTTP: m.runtime.HTTP,
		},
	}

	if arguments.AccessKeyID != "" || arguments.SecretAccessKey != "" {
		clientConfig.Endpoint = &utils.EnvString{
			Value: &arguments.Endpoint,
		}

		clientConfig.Authentication = azblob.AuthCredentials{
			Type: azblob.AuthTypeSharedKey,
			AccountName: &utils.EnvString{
				Value: utils.ToPtr(arguments.AccessKeyID),
			},
			AccountKey: &utils.EnvString{
				Value: utils.ToPtr(arguments.SecretAccessKey),
			},
		}
	} else {
		clientConfig.Authentication = azblob.AuthCredentials{
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
		id:                     common.StorageClientID(clientId),
		StorageClient:          client,
		defaultPresignedExpiry: &defaultPresignedExpiry,
	}, nil
}
