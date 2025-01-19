package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"time"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

// Manager represents the high-level client that manages internal clients and configurations.
type Manager struct {
	clients []Client
}

// NewManager creates a storage client manager instance.
func NewManager(ctx context.Context, configs []ClientConfig, logger *slog.Logger, version string) (*Manager, error) {
	if len(configs) == 0 {
		return nil, errors.New("failed to initialize storage clients: config is empty")
	}

	result := &Manager{
		clients: make([]Client, len(configs)),
	}

	for i, config := range configs {
		baseConfig, client, err := config.toStorageClient(ctx, logger, version)
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
func (m *Manager) GetClientAndBucket(clientID *common.StorageClientID, bucketName string) (*Client, string, error) {
	hasClientID := clientID != nil && *clientID != ""
	if !hasClientID && bucketName == "" {
		client, _ := m.GetClient(nil)

		return client, client.defaultBucket, nil
	}

	if hasClientID {
		client, ok := m.GetClient(clientID)
		if !ok {
			return nil, "", schema.InternalServerError("client not found: "+string(*clientID), nil)
		}

		bucketName, err := client.ValidateBucket(bucketName)
		if err != nil {
			return nil, "", err
		}

		return client, bucketName, nil
	}

	for _, c := range m.clients {
		if c.defaultBucket == bucketName || slices.Contains(c.allowedBuckets, bucketName) {
			return &c, bucketName, nil
		}
	}

	// return the first client by default
	return &m.clients[0], bucketName, nil
}
