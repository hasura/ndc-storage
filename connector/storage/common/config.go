package common

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hasura/ndc-http/exhttp"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/invopop/jsonschema"
)

// BaseClientConfig holds common configurations of a storage client
type BaseClientConfig struct {
	// The unique identity of a client. Use this setting if there are many configured clients.
	ID string `json:"id,omitempty" mapstructure:"id" yaml:"id,omitempty"`
	// Cloud provider type of the storage client.
	Type StorageProviderType `json:"type" mapstructure:"type" yaml:"type"`
	// Default bucket name to be set if the user doesn't specify any bucket.
	DefaultBucket utils.EnvString `json:"defaultBucket" mapstructure:"defaultBucket" yaml:"defaultBucket"`
	// Endpoint of the storage server. Required for other S3 compatible services such as MinIO, Cloudflare R2, DigitalOcean Spaces, etc...
	Endpoint *utils.EnvString `json:"endpoint,omitempty" mapstructure:"endpoint" yaml:"endpoint,omitempty"`
	// Maximum number of retry times.
	MaxRetries *int `json:"maxRetries,omitempty" mapstructure:"maxRetries" yaml:"maxRetries,omitempty"`
	// The default expiry for presigned URL generation. The maximum expiry is 604800 seconds (i.e. 7 days) and minimum is 1 second.
	DefaultPresignedExpiry *string `json:"defaultPresignedExpiry,omitempty" mapstructure:"defaultPresignedExpiry" yaml:"defaultPresignedExpiry,omitempty"`
	// Allowed buckets. This setting prevents users to get buckets and objects outside the list.
	// However, it's recommended to restrict the permissions for the IAM credentials.
	// This setting is useful to let the connector know which buckets belong to this client.
	// The empty value means all buckets are allowed. The storage server will handle the validation.
	AllowedBuckets []string `json:"allowedBuckets,omitempty" mapstructure:"allowedBuckets" yaml:"allowedBuckets,omitempty"`
}

// Validate checks if the configration is valid.
func (bcc BaseClientConfig) Validate() error {
	if err := bcc.Type.Validate(); err != nil {
		return fmt.Errorf("type: %w", err)
	}

	if _, _, _, err := bcc.ValidateEndpoint(); err != nil {
		return err
	}

	return nil
}

// ValidateEndpoint gets and validates endpoint settings
func (bcc BaseClientConfig) ValidateEndpoint() (*url.URL, int, bool, error) {
	port := 80
	if bcc.Endpoint == nil {
		return nil, port, false, nil
	}

	var useSSL bool

	rawEndpoint, err := bcc.Endpoint.GetOrDefault("")
	if err != nil {
		return nil, port, false, fmt.Errorf("endpoint: %w", err)
	}

	if rawEndpoint == "" {
		return nil, port, useSSL, nil
	}

	endpointURL, err := exhttp.ParseHttpURL(rawEndpoint)
	if err != nil {
		return nil, port, false, fmt.Errorf("invalid endpoint url: %w", err)
	}

	useSSL = endpointURL.Scheme == "https"

	port, err = exhttp.ParsePort(endpointURL.Port(), endpointURL.Scheme)
	if err != nil {
		return nil, 0, false, err
	}

	return endpointURL, port, useSSL, nil
}

// GetJSONSchema is used to generate a custom jsonschema.
func (bcc BaseClientConfig) GetJSONSchema(providerTypes []any) *jsonschema.Schema {
	envStringRef := "#/$defs/EnvString"

	properties := jsonschema.NewProperties()
	properties.Set("type", &jsonschema.Schema{
		Description: "Cloud provider type of the storage client",
		Type:        "string",
		Enum:        providerTypes,
	})

	properties.Set("id", &jsonschema.Schema{
		Description: "The unique identity of a client. Use this setting if there are many configured clients",
		Type:        "string",
	})
	properties.Set("defaultBucket", &jsonschema.Schema{
		Description: "Default bucket name to be set if the user doesn't specify any bucket",
		Ref:         envStringRef,
	})
	properties.Set("endpoint", &jsonschema.Schema{
		Description: "Endpoint of the storage server. Required for other S3 compatible services such as MinIO, Cloudflare R2, DigitalOcean Spaces, etc...",
		Ref:         envStringRef,
	})
	properties.Set("maxRetries", &jsonschema.Schema{
		Description: "Maximum number of retry times",
		Type:        "integer",
		Minimum:     json.Number("1"),
		Default:     10,
	})
	properties.Set("defaultPresignedExpiry", &jsonschema.Schema{
		Description: "Default bucket name to be set if the user doesn't specify any bucket",
		Type:        "string",
		Pattern:     `^((([0-9]+h)?([0-9]+m)?([0-9]+s))|(([0-9]+h)?([0-9]+m))|([0-9]+h))$`,
		Default:     "24h",
	})
	properties.Set("allowedBuckets", &jsonschema.Schema{
		Description: "Allowed buckets. This setting prevents users to get buckets and objects outside the list. However, it's recommended to restrict the permissions for the IAM credentials",
		Type:        "array",
		Items: &jsonschema.Schema{
			Type: "string",
		},
	})

	return &jsonschema.Schema{
		Type:       "object",
		Properties: properties,
		Required:   []string{"type", "defaultBucket"},
	}
}
