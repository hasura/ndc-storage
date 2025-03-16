package common

import (
	"github.com/invopop/jsonschema"
)

const (
	HeaderContentType        string = "content-type"
	HeaderCacheControl       string = "cache-control"
	HeaderContentDisposition string = "content-disposition"
	HeaderContentEncoding    string = "content-encoding"
	HeaderContentLanguage    string = "content-language"
)

// StorageClientID the storage client ID enum.
// @scalar StorageClientID string
type StorageClientID string

// StorageProviderType represents a storage provider type enum.
// @enum s3,gcs,azblob,fs
type StorageProviderType string

// Validate checks if the provider type is valid.
func (spt StorageProviderType) Validate() error {
	_, err := ParseStorageProviderType(string(spt))

	return err
}

// JSONSchema is used to generate a custom jsonschema.
func (spt StorageProviderType) JSONSchema() *jsonschema.Schema {
	enumValues := make([]any, len(enumValues_StorageProviderType))
	for i, item := range enumValues_StorageProviderType {
		enumValues[i] = string(item)
	}

	return &jsonschema.Schema{
		Type: "string",
		Enum: enumValues,
	}
}
