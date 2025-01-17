package common

import (
	"fmt"
	"slices"

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
type StorageProviderType string

const (
	S3             StorageProviderType = "s3"
	GoogleStorage  StorageProviderType = "gs"
	AzureBlobStore StorageProviderType = "azblob"
)

var enumValues_StorageProviderType = []StorageProviderType{
	S3, GoogleStorage, AzureBlobStore,
}

// ParseStorageProviderType parses the StorageProviderType from string.
func ParseStorageProviderType(input string) (StorageProviderType, error) {
	result := StorageProviderType(input)
	if !slices.Contains(enumValues_StorageProviderType, result) {
		return "", fmt.Errorf("invalid StorageProviderType, expected one of %v, got: %s", enumValues_StorageProviderType, input)
	}

	return result, nil
}

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
