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

	ContentTypeTextPlain       string = "text/plain"
	ContentTypeTextCSV         string = "text/csv"
	ContentTypeApplicationJSON string = "application/json"
)

// StorageClientID the storage client ID enum.
// @scalar StorageClientID string
type StorageClientID string

// StorageProviderType represents a storage provider type enum.
// @enum s3,gcs,azblob
type StorageProviderType string

const (
	S3             StorageProviderType = "s3"
	GoogleStorage  StorageProviderType = "gcs"
	AzureBlobStore StorageProviderType = "azblob"
)

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
