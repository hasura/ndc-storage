package azblob

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/invopop/jsonschema"
)

var (
	errRequireAccountName     = errors.New("accountName is required")
	errRequireAccountKey      = errors.New("accountKey is required")
	errRequireStorageEndpoint = errors.New("endpoint is required")
)

// ClientConfig represent the raw configuration of a MinIO client.
type ClientConfig struct {
	common.BaseClientConfig `yaml:",inline"`
	OtherConfig             `yaml:",inline"`
}

// JSONSchema is used to generate a custom jsonschema.
func (cc ClientConfig) JSONSchema() *jsonschema.Schema {
	result := cc.BaseClientConfig.GetJSONSchema([]any{common.AzureBlobStore})
	result.Required = append(result.Required, "authentication")
	result.Properties.Set("authentication", cc.Authentication.JSONSchema())

	return result
}

func (cc ClientConfig) toAzureBlobClient(logger *slog.Logger) (*azblob.Client, error) {
	endpoint, _, useSSL, err := cc.BaseClientConfig.ValidateEndpoint()
	if err != nil {
		return nil, err
	}

	maxRetries := 0
	if cc.MaxRetries != nil && *cc.MaxRetries > 0 {
		maxRetries = *cc.MaxRetries
	}

	isDebug := utils.IsDebug(logger)

	opts := &azblob.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Retry: policy.RetryOptions{
				MaxRetries: int32(maxRetries),
			},
			Logging: policy.LogOptions{
				IncludeBody: isDebug,
			},
			InsecureAllowCredentialWithHTTP: !useSSL,
		},
	}

	return cc.Authentication.toAzureBlobClient(endpoint, opts)
}

// OtherConfig holds MinIO-specific configurations
type OtherConfig struct {
	// Authentication credentials.
	Authentication AuthCredentials `json:"authentication" yaml:"authentication"`
}

// AuthType represents the authentication type enum.
type AuthType string

const (
	AuthTypeNone             AuthType = "none"
	AuthTypeSharedKey        AuthType = "sharedKey"
	AuthTypeActiveDirectory  AuthType = "azureAD"
	AuthTypeConnectionString AuthType = "connectionString"
)

var enumValues_AuthType = []AuthType{
	AuthTypeNone,
	AuthTypeSharedKey,
	AuthTypeActiveDirectory,
	AuthTypeConnectionString,
}

// ParseAuthType parses the AuthType from string.
func ParseAuthType(input string) (AuthType, error) {
	result := AuthType(input)
	if !slices.Contains(enumValues_AuthType, result) {
		return "", fmt.Errorf("invalid AuthType, expected one of %v, got: %s", enumValues_AuthType, input)
	}

	return result, nil
}

// Validate checks if the provider type is valid.
func (at AuthType) Validate() error {
	_, err := ParseAuthType(string(at))

	return err
}

// AuthCredentials represent the authentication credentials infomartion.
type AuthCredentials struct {
	// The authentication type
	Type AuthType `json:"type" yaml:"type"`
	// Access Key ID.
	AccountName *utils.EnvString `json:"accountName,omitempty" yaml:"accountName,omitempty"`
	// Secret Access Key.
	AccountKey *utils.EnvString `json:"accountKey,omitempty" yaml:"accountKey,omitempty"`
}

// JSONSchema is used to generate a custom jsonschema.
func (ac AuthCredentials) JSONSchema() *jsonschema.Schema {
	envStringRef := &jsonschema.Schema{
		Ref: "#/$defs/EnvString",
	}

	staticProps := jsonschema.NewProperties()
	staticProps.Set("type", &jsonschema.Schema{
		Type: "string",
		Enum: []any{AuthTypeSharedKey},
	})
	staticProps.Set("accountName", envStringRef)
	staticProps.Set("accountKey", envStringRef)

	adProps := jsonschema.NewProperties()
	adProps.Set("type", &jsonschema.Schema{
		Type: "string",
		Enum: []any{AuthTypeActiveDirectory, AuthTypeNone},
	})
	adProps.Set("accountName", envStringRef)

	connStringProps := jsonschema.NewProperties()
	connStringProps.Set("type", &jsonschema.Schema{
		Type: "string",
		Enum: []any{AuthTypeConnectionString},
	})

	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Type:       "object",
				Properties: staticProps,
				Required:   []string{"type", "accountName", "accountKey"},
			},
			{
				Type:       "object",
				Properties: adProps,
				Required:   []string{"type"},
			},
			{
				Type:       "object",
				Properties: connStringProps,
				Required:   []string{"type"},
			},
		},
	}
}

func (ac AuthCredentials) toAzureBlobClient(endpoint string, options *azblob.ClientOptions) (*azblob.Client, error) {
	accountName, accountKey, err := ac.parseAccountNameAndKey()
	if err != nil {
		return nil, err
	}

	serviceURL := endpoint
	if accountName != "" && endpoint == "" {
		serviceURL = fmt.Sprintf("https://%s.blob.core.windows.net/", accountKey)
	}

	switch ac.Type {
	case AuthTypeNone:
		return azblob.NewClientWithNoCredential(serviceURL, options)
	case AuthTypeSharedKey:
		if accountName == "" {
			return nil, errRequireAccountName
		}

		if accountKey == "" {
			return nil, errRequireAccountKey
		}

		cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
		if err != nil {
			return nil, err
		}

		return azblob.NewClientWithSharedKeyCredential(serviceURL, cred, options)
	case AuthTypeActiveDirectory:
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, err
		}

		return azblob.NewClient(serviceURL, cred, options)
	case AuthTypeConnectionString:
		if endpoint == "" {
			return nil, errRequireStorageEndpoint
		}

		return azblob.NewClientFromConnectionString(serviceURL, options)
	default:
		return nil, fmt.Errorf("unsupported auth type %s", ac.Type)
	}
}

func (ac AuthCredentials) parseAccountNameAndKey() (string, string, error) {
	accountName, err := ac.AccountName.GetOrDefault("")
	if err != nil {
		return "", "", fmt.Errorf("accountKey: %w", err)
	}

	accountKey, err := ac.AccountKey.GetOrDefault("")
	if err != nil {
		return "", "", fmt.Errorf("accountKey: %w", err)
	}

	return accountName, accountKey, nil
}
