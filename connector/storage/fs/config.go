package fs

import (
	"encoding/json"
	"fmt"

	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/invopop/jsonschema"
)

var defaultFilePermissions FilePermissionConfig = FilePermissionConfig{
	Directory: 0o755,
	File:      0o644,
}

// ClientConfig represent the raw configuration of a MinIO client.
type ClientConfig struct {
	// The unique identity of a client. Use this setting if there are many configured clients.
	ID string `json:"id,omitempty" mapstructure:"id" yaml:"id,omitempty"`
	// Cloud provider type of the storage client.
	Type common.StorageProviderType `json:"type" mapstructure:"type" yaml:"type"`
	// Default directory for storage files.
	Permissions *FilePermissionConfig `json:"permissions,omitempty" mapstructure:"permissions" yaml:"permissions"`
	// Default directory for storage files.
	DefaultDirectory utils.EnvString `json:"defaultDirectory" mapstructure:"defaultDirectory" yaml:"defaultDirectory"`
	// Allowed directories. This setting prevents users to browse files outside the list.
	AllowedDirectories []string `json:"allowedDirectories,omitempty" mapstructure:"allowedDirectories" yaml:"allowedDirectories,omitempty"`
}

// Validate checks if the configuration is valid.
func (cc ClientConfig) Validate() error {
	if cc.Permissions == nil {
		return nil
	}

	return cc.Permissions.Validate()
}

// ToBaseConfig creates a BaseClientConfig instance.
func (cc ClientConfig) ToBaseConfig() *common.BaseClientConfig {
	return &common.BaseClientConfig{
		ID:             cc.ID,
		Type:           cc.Type,
		DefaultBucket:  cc.DefaultDirectory,
		AllowedBuckets: cc.AllowedDirectories,
	}
}

// JSONSchema is used to generate a custom jsonschema.
func (cc ClientConfig) JSONSchema() *jsonschema.Schema {
	envStringRefName := "#/$defs/EnvString"
	properties := jsonschema.NewProperties()

	properties.Set("type", &jsonschema.Schema{
		Description: "Cloud provider type of the storage client",
		Type:        "string",
		Enum:        []any{common.StorageProviderTypeFs},
	})

	properties.Set("id", &jsonschema.Schema{
		Description: "The unique identity of a client. Use this setting if there are many configured clients",
		Type:        "string",
	})

	properties.Set("defaultDirectory", &jsonschema.Schema{
		Description: "Default directory location to be set if the user doesn't specify any bucket",
		Ref:         envStringRefName,
	})

	properties.Set("allowedDirectories", &jsonschema.Schema{
		Description: "Allowed directories. This setting prevents users to browse files outside the list",
		Type:        "array",
		Items: &jsonschema.Schema{
			Type: "string",
		},
	})

	properties.Set("permissions", FilePermissionConfig{}.JSONSchema())

	return &jsonschema.Schema{
		Type:       "object",
		Properties: properties,
		Required:   []string{"type", "defaultDirectory"},
	}
}

// FilePermissionConfig the default file permission configuration.
type FilePermissionConfig struct {
	// Default directory permission.
	Directory int `json:"directory" mapstructure:"directory" yaml:"directory"`
	// Default file permission.
	File int `json:"file" mapstructure:"file" yaml:"file"`
}

// Validate checks if the configuration is valid.
func (fpc FilePermissionConfig) Validate() error {
	if err := fpc.validatePermission("directory", fpc.Directory); err != nil {
		return err
	}

	if err := fpc.validatePermission("file", fpc.File); err != nil {
		return err
	}

	return nil
}

func (fpc FilePermissionConfig) validatePermission(key string, perm int) error {
	if perm < 0 || perm > 777 {
		return fmt.Errorf("%s permission must be in between 000 and 777, got %d", key, perm)
	}

	return nil
}

// JSONSchema is used to generate a custom jsonschema.
func (fpc FilePermissionConfig) JSONSchema() *jsonschema.Schema {
	properties := jsonschema.NewProperties()

	properties.Set("directory", &jsonschema.Schema{
		Description: "Default directory permission",
		Type:        "integer",
		Default:     644,
		Minimum:     json.Number("0"),
		Maximum:     json.Number("777"),
	})

	properties.Set("file", &jsonschema.Schema{
		Description: "Default file permission",
		Type:        "integer",
		Default:     644,
		Minimum:     json.Number("0"),
		Maximum:     json.Number("777"),
	})

	return &jsonschema.Schema{
		Type:       "object",
		Properties: properties,
		Required:   []string{"directory", "file"},
	}
}
