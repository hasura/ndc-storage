package fs

import (
	"fmt"

	"github.com/hasura/ndc-storage/connector/storage/common"
)

var defaultFilePermissions FilePermissionConfig = FilePermissionConfig{
	Directory: 644,
	File:      644,
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
	DefaultDirectory string `json:"defaultDirectory" mapstructure:"defaultDirectory" yaml:"defaultDirectory"`
	// Allowed directories. This setting prevents users to browse files outside the list.
	AllowedDirectories []string `json:"allowedDirectories,omitempty" mapstructure:"allowedDirectories" yaml:"allowedDirectories,omitempty"`
}

// Validate checks if the configuration is valid.
func (fpc ClientConfig) Validate() error {
	if fpc.Permissions == nil {
		return nil
	}

	return fpc.Permissions.Validate()
}

// FilePermissionConfig the default file permission configuration.
type FilePermissionConfig struct {
	Directory int `json:"directory" jsonschema:"default=644,min=0,max=777" mapstructure:"directory" yaml:"directory"`
	File      int `json:"file"      jsonschema:"default=644,min=0,max=777" mapstructure:"file"      yaml:"file"`
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
