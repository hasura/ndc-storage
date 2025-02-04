package types

import (
	"errors"
	"fmt"

	"github.com/hasura/ndc-storage/connector/storage"
)

const (
	ConfigurationFileName = "configuration.yaml"
)

// Configuration contains required settings for the connector.
type Configuration struct {
	// List of storage client configurations and credentials.
	Clients []storage.ClientConfig `json:"clients" yaml:"clients"`
	// Settings for concurrent webhook executions to remote servers.
	Concurrency ConcurrencySettings `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
	// Common runtime settings for all clients.
	Runtime storage.RuntimeSettings `json:"runtime" yaml:"runtime"`
}

// Validate checks if the configuration is valid.
func (c Configuration) Validate() error {
	if len(c.Clients) == 0 {
		return errors.New("require at least 1 element in the clients array")
	}

	for i, c := range c.Clients {
		if err := c.Validate(); err != nil {
			return fmt.Errorf("invalid client configuration at %d: %w", i, err)
		}
	}

	if c.Runtime.MaxDownloadSizeMBs <= 0 {
		return errors.New("maxDownloadSizeMBs must be larger than 0")
	}

	return nil
}

// ConcurrencySettings represent settings for concurrent webhook executions to remote servers.
type ConcurrencySettings struct {
	// Maximum number of concurrent executions if there are many query variables.
	Query int `json:"query" jsonschema:"min=1,default=5" yaml:"query"`
	// Maximum number of concurrent executions if there are many mutation operations.
	Mutation int `json:"mutation" jsonschema:"min=1,default=1" yaml:"mutation"`
}
