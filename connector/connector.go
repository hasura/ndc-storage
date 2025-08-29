package connector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/hasura/ndc-sdk-go/v2/connector"
	"github.com/hasura/ndc-sdk-go/v2/schema"
	"github.com/hasura/ndc-sdk-go/v2/utils"
	"github.com/hasura/ndc-storage/connector/collection"
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
func (c *Connector) ParseConfiguration(
	ctx context.Context,
	configurationDir string,
) (*types.Configuration, error) {
	configBytes, err := os.ReadFile(filepath.Join(configurationDir, types.ConfigurationFileName))
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration: %w", err)
	}

	var config types.Configuration
	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to decode configuration: %w", err)
	}

	if !config.Generator.DynamicCredentials && len(config.Clients) == 0 {
		return nil, errors.New("failed to initialize storage clients: config is empty")
	}

	c.config = &config

	connectorCapabilities := schema.CapabilitiesResponse{
		Version: schema.NDCVersion,
		Capabilities: schema.Capabilities{
			Query: schema.QueryCapabilities{
				Variables: &schema.LeafCapability{},
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
func (c *Connector) TryInitState(
	ctx context.Context,
	configuration *types.Configuration,
	metrics *connector.TelemetryState,
) (*types.State, error) {
	logger := connector.GetLogger(ctx)

	manager, err := storage.NewManager(ctx, configuration.Clients, configuration.Runtime, logger)
	if err != nil {
		return nil, err
	}

	connectorSchema, errs := utils.MergeSchemas(
		GetConnectorSchema(),
		collection.GetConnectorSchema(
			manager.GetClientIDs(),
			c.config.Generator.DynamicCredentials,
		),
	)
	for _, err := range errs {
		slog.Debug(err.Error())
	}

	c.evalSchema(connectorSchema)

	schemaBytes, err := json.Marshal(connectorSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to encode schema: %w", err)
	}

	c.rawSchema = schema.NewRawSchemaResponseUnsafe(schemaBytes)

	return &types.State{
		Storage:        manager,
		TelemetryState: metrics,
		Concurrency:    c.config.Concurrency,
	}, nil
}

// HealthCheck checks the health of the connector.
//
// For example, this function should check that the connector
// is able to reach its data source over the network.
//
// Should throw if the check fails, else resolve.
func (c *Connector) HealthCheck(
	ctx context.Context,
	configuration *types.Configuration,
	state *types.State,
) error {
	return nil
}

// GetCapabilities get the connector's capabilities.
func (c *Connector) GetCapabilities(
	configuration *types.Configuration,
) schema.CapabilitiesResponseMarshaler {
	return c.capabilities
}

// GetSchema gets the connector's schema.
func (c *Connector) GetSchema(
	ctx context.Context,
	configuration *types.Configuration,
	_ *types.State,
) (schema.SchemaResponseMarshaler, error) {
	return c.rawSchema, nil
}

func (c *Connector) evalSchema(connectorSchema *schema.SchemaResponse) {
	// override field types of the StorageObject object
	objectClientID := connectorSchema.ObjectTypes[collection.StorageObjectName].Fields[collection.StorageObjectColumnClientID]
	objectClientID.Type = schema.NewNamedType(collection.ScalarStorageClientID).Encode()
	connectorSchema.ObjectTypes[collection.StorageObjectName].Fields[collection.StorageObjectColumnClientID] = objectClientID

	objectBucketField := connectorSchema.ObjectTypes[collection.StorageObjectName].Fields[collection.StorageObjectColumnBucket]
	objectBucketField.Type = schema.NewNamedType(collection.ScalarBucketName).Encode()
	connectorSchema.ObjectTypes[collection.StorageObjectName].Fields[collection.StorageObjectColumnBucket] = objectBucketField

	objectNameField := connectorSchema.ObjectTypes[collection.StorageObjectName].Fields[collection.StorageObjectColumnName]
	objectNameField.Type = schema.NewNamedType(collection.ScalarStringFilter).Encode()
	connectorSchema.ObjectTypes[collection.StorageObjectName].Fields[collection.StorageObjectColumnName] = objectNameField

	// override field types of the StorageBucket object
	bucketClientID := connectorSchema.ObjectTypes[collection.StorageBucketName].Fields[collection.StorageObjectColumnClientID]
	bucketClientID.Type = schema.NewNamedType(collection.ScalarStorageClientID).Encode()
	connectorSchema.ObjectTypes[collection.StorageBucketName].Fields[collection.StorageObjectColumnClientID] = bucketClientID

	bucketNameField := connectorSchema.ObjectTypes[collection.StorageBucketName].Fields[collection.StorageObjectColumnName]
	bucketNameField.Type = schema.NewNamedType(collection.ScalarStringFilter).Encode()
	connectorSchema.ObjectTypes[collection.StorageBucketName].Fields[collection.StorageObjectColumnName] = bucketNameField

	dynamicCredentialArguments := []string{
		collection.ArgumentClientType,
		collection.ArgumentEndpoint,
		collection.ArgumentAccessKeyID,
		collection.ArgumentSecretAccessKey,
	}

	for i, f := range connectorSchema.Functions {
		if c.config.Generator.PromptQLCompatible &&
			!slices.Contains(
				[]string{"storageBucketConnections", "storageObjectConnections"},
				f.Name,
			) {
			// remove boolean expression arguments in commands
			delete(f.Arguments, "where")
		}

		if !c.config.Generator.DynamicCredentials {
			// delete dynamic credential arguments
			for _, key := range dynamicCredentialArguments {
				delete(f.Arguments, key)
			}
		}

		if len(c.config.Clients) == 0 {
			delete(f.Arguments, collection.StorageObjectColumnClientID)
		}

		connectorSchema.Functions[i] = f
	}

	procedures := []schema.ProcedureInfo{}

	for _, f := range connectorSchema.Procedures {
		if !c.config.Generator.DynamicCredentials {
			// delete dynamic credential arguments
			for _, key := range dynamicCredentialArguments {
				delete(f.Arguments, key)
			}
		} else if len(c.config.Clients) == 0 && slices.Contains([]string{"composeStorageObject", "copyStorageObject"}, f.Name) {
			continue
		}

		if c.config.Generator.PromptQLCompatible {
			// remove boolean expression arguments in commands
			delete(f.Arguments, "where")
		}

		if len(c.config.Clients) == 0 {
			delete(f.Arguments, collection.StorageObjectColumnClientID)
		}

		procedures = append(procedures, f)
	}

	connectorSchema.Procedures = procedures

	if c.config.Generator.PromptQLCompatible {
		// PromptQL doesn't support bytes representation.
		bytesScalar := schema.NewScalarType()
		bytesScalar.Representation = schema.NewTypeRepresentationString().Encode()
		connectorSchema.ScalarTypes["Bytes"] = *bytesScalar
	}

	if !c.config.Generator.DynamicCredentials {
		for name, object := range connectorSchema.ObjectTypes {
			for _, key := range dynamicCredentialArguments {
				delete(object.Fields, key)
			}

			connectorSchema.ObjectTypes[name] = object
		}
	}
}
