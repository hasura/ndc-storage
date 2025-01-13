package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/invopop/jsonschema"
)

// BaseClientConfig holds common configurations of a storage client
type BaseClientConfig struct {
	// The unique identity of a client. Use this setting if there are many configured clients.
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	// Cloud provider type of the storage client.
	Type StorageProviderType `json:"type" yaml:"type"`
	// Default bucket name to be set if the user doesn't specify any bucket.
	DefaultBucket utils.EnvString `json:"defaultBucket" yaml:"defaultBucket"`
	// Endpoint of the storage server. Required for other S3 compatible services such as MinIO, Cloudflare R2, DigitalOcean Spaces, etc...
	Endpoint *utils.EnvString `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	// The public host to be used for presigned URL generation.
	PublicHost *utils.EnvString `json:"publicHost,omitempty" yaml:"publicHost,omitempty"`
	// Maximum number of retry times.
	MaxRetries *int `json:"maxRetries,omitempty" jsonschema:"min=1,default=10" yaml:"maxRetries,omitempty"`
	// The default expiry for presigned URL generation. The maximum expiry is 604800 seconds (i.e. 7 days) and minimum is 1 second.
	DefaultPresignedExpiry *string `json:"defaultPresignedExpiry,omitempty" jsonschema:"pattern=[0-9]+(s|m|h),default=24h" yaml:"defaultPresignedExpiry,omitempty"`
	// Allowed buckets. This setting prevents users to get buckets and objects outside the list.
	// However, it's recommended to restrict the permissions for the IAM credentials.
	// This setting is useful to let the connector know which buckets belong to this client.
	// The empty value means all buckets are allowed. The storage server will handle the validation.
	AllowedBuckets []string `json:"allowedBuckets,omitempty" yaml:"allowedBuckets,omitempty"`
}

// Validate checks if the configration is valid.
func (bcc BaseClientConfig) Validate() error {
	if err := bcc.Type.Validate(); err != nil {
		return fmt.Errorf("type: %w", err)
	}

	if _, _, _, err := bcc.ValidateEndpoint(); err != nil {
		return err
	}

	if _, err := bcc.ValidatePublicHost(); err != nil {
		return err
	}

	return nil
}

// ValidatePublicHost validates the public host setting.
func (bcc BaseClientConfig) ValidatePublicHost() (*url.URL, error) {
	if bcc.PublicHost == nil {
		return nil, nil
	}

	publicHost, err := bcc.PublicHost.GetOrDefault("")
	if err != nil {
		return nil, fmt.Errorf("publicHost: %w", err)
	}

	if strings.HasPrefix(publicHost, "http") {
		result, err := url.Parse(publicHost)
		if err != nil {
			return nil, fmt.Errorf("publicHost: %w", err)
		}

		return result, nil
	}

	return &url.URL{
		Host: publicHost,
	}, nil
}

// ValidateEndpoint gets and validates endpoint settings
func (bcc BaseClientConfig) ValidateEndpoint() (string, int, bool, error) {
	port := 80
	if bcc.Endpoint == nil {
		return "", port, false, nil
	}

	var endpoint string
	var useSSL bool

	rawEndpoint, err := bcc.Endpoint.GetOrDefault("")
	if err != nil {
		return "", port, false, fmt.Errorf("endpoint: %w", err)
	}

	if rawEndpoint == "" {
		return endpoint, port, useSSL, nil
	}

	endpointURL, err := url.Parse(rawEndpoint)
	if err != nil {
		return "", port, false, fmt.Errorf("invalid endpoint url: %w", err)
	}

	if !strings.HasPrefix(endpointURL.Scheme, "http") {
		return "", port, false, errors.New("invalid endpoint url http scheme: " + endpointURL.Scheme)
	}

	endpoint = endpointURL.Host

	if endpointURL.Scheme == "https" {
		useSSL = true
		port = 443
	}

	rawPort := endpointURL.Port()
	if rawPort != "" {
		p, err := strconv.Atoi(rawPort)
		if err != nil {
			return "", 0, false, fmt.Errorf("invalid endpoint port: %s", rawPort)
		}

		port = p
	}

	return endpoint, port, useSSL, nil
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
	properties.Set("publicHost", &jsonschema.Schema{
		Description: "The public host to be used for presigned URL generation",
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
		Pattern:     `[0-9]+(s|m|h)`,
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
