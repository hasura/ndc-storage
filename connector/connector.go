package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/functions"
	"github.com/hasura/ndc-storage/connector/storage"
	"github.com/hasura/ndc-storage/connector/types"
	"gopkg.in/yaml.v3"
)

// Connector implements the SDK interface of NDC specification.
type Connector struct {
	capabilities *schema.RawCapabilitiesResponse
	rawSchema    *schema.RawSchemaResponse
	config       *types.Configuration
	apiHandler   functions.DataConnectorHandler
}

// ParseConfiguration validates the configuration files provided by the user, returning a validated 'Configuration',
// or throwing an error to prevents Connector startup.
func (c *Connector) ParseConfiguration(ctx context.Context, configurationDir string) (*types.Configuration, error) {
	configBytes, err := os.ReadFile(filepath.Join(configurationDir, types.ConfigurationFileName))
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration: %w", err)
	}

	var config types.Configuration
	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to decode configuration: %w", err)
	}

	c.config = &config

	connectorCapabilities := schema.CapabilitiesResponse{
		Version: "0.1.6",
		Capabilities: schema.Capabilities{
			Query: schema.QueryCapabilities{
				Variables:    schema.LeafCapability{},
				NestedFields: schema.NestedFieldCapabilities{},
			},
			Mutation: schema.MutationCapabilities{},
		},
	}

	rawCapabilities, err := json.Marshal(connectorCapabilities)
	if err != nil {
		return nil, fmt.Errorf("failed to encode capabilities: %w", err)
	}

	c.capabilities = schema.NewRawCapabilitiesResponseUnsafe(rawCapabilities)
	c.apiHandler = functions.DataConnectorHandler{}

	return &config, nil
}

// TryInitState initializes the connector's in-memory state.
//
// For example, any connection pools, prepared queries,
// or other managed resources would be allocated here.
//
// In addition, this function should register any
// connector-specific metrics with the metrics registry.
func (c *Connector) TryInitState(ctx context.Context, configuration *types.Configuration, metrics *connector.TelemetryState) (*types.State, error) {
	logger := connector.GetLogger(ctx)

	manager, err := storage.NewManager(ctx, configuration.Clients, logger)
	if err != nil {
		return nil, err
	}

	connectorSchema, errs := utils.MergeSchemas(GetConnectorSchema(), functions.GetBaseConnectorSchema(manager.GetClientIDs()))
	for _, err := range errs {
		slog.Debug(err.Error())
	}

	schemaBytes, err := json.Marshal(connectorSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to encode schema: %w", err)
	}

	c.rawSchema = schema.NewRawSchemaResponseUnsafe(schemaBytes)

	return &types.State{
		Storage:        manager,
		TelemetryState: metrics,
	}, nil
}

// HealthCheck checks the health of the connector.
//
// For example, this function should check that the connector
// is able to reach its data source over the network.
//
// Should throw if the check fails, else resolve.
func (c *Connector) HealthCheck(ctx context.Context, configuration *types.Configuration, state *types.State) error {
	return nil
}

// GetCapabilities get the connector's capabilities.
func (c *Connector) GetCapabilities(configuration *types.Configuration) schema.CapabilitiesResponseMarshaler {
	return c.capabilities
}

// GetSchema gets the connector's schema.
func (c *Connector) GetSchema(ctx context.Context, configuration *types.Configuration, _ *types.State) (schema.SchemaResponseMarshaler, error) {
	return c.rawSchema, nil
}
