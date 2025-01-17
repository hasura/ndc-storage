package gcs

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/invopop/jsonschema"
	"google.golang.org/api/option"
)

var (
	errRequireCredentials = errors.New("require either credential JSON or file")
	errRequireProjectID   = errors.New("projectId is required")
)

// ClientConfig represent the raw configuration of a MinIO client.
type ClientConfig struct {
	common.BaseClientConfig `yaml:",inline"`
	OtherConfig             `yaml:",inline"`
}

// JSONSchema is used to generate a custom jsonschema.
func (cc ClientConfig) JSONSchema() *jsonschema.Schema {
	envStringRef := "#/$defs/EnvString"

	result := cc.BaseClientConfig.GetJSONSchema([]any{common.GoogleStorage})
	result.Required = append(result.Required, "authentication", "projectId")
	result.Properties.Set("authentication", cc.Authentication.JSONSchema())

	result.Properties.Set("publicHost", &jsonschema.Schema{
		Description: "The public host to be used for presigned URL generation",
		Ref:         envStringRef,
	})
	result.Properties.Set("projectId", &jsonschema.Schema{
		Description: "Project ID of the Google Cloud account",
		Ref:         envStringRef,
	})

	return result
}

// OtherConfig holds MinIO-specific configurations
type OtherConfig struct {
	// Project ID of the Google Cloud account.
	ProjectID utils.EnvString `json:"projectId" mapstructure:"projectId" yaml:"projectId"`
	// The public host to be used for presigned URL generation.
	PublicHost *utils.EnvString `json:"publicHost,omitempty" mapstructure:"publicHost" yaml:"publicHost,omitempty"`
	// Authentication credentials.
	Authentication AuthCredentials `json:"authentication" mapstructure:"authentication" yaml:"authentication"`
}

func (cc ClientConfig) toClientOptions(version string) ([]option.ClientOption, error) {
	var opts []option.ClientOption

	cred, err := cc.Authentication.toCredentials()
	if err != nil {
		return nil, err
	}

	opts = append(opts, cred)

	endpointURL, _, _, err := cc.BaseClientConfig.ValidateEndpoint()
	if err != nil {
		return nil, err
	}

	if endpointURL != nil {
		opts = append(opts, option.WithEndpoint(endpointURL.String()))
	}

	opts = append(opts, option.WithUserAgent(fmt.Sprintf("hasura/ndc-storage (%s)", version)))

	return opts, nil
}

// ValidatePublicHost validates the public host setting.
func (cc ClientConfig) ValidatePublicHost() (*url.URL, error) {
	if cc.PublicHost == nil {
		return nil, nil
	}

	publicHost, err := cc.PublicHost.GetOrDefault("")
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

// AuthType represents the authentication type enum.
type AuthType string

const (
	AuthTypeCredentials AuthType = "credentials"
	AuthTypeAnonymous   AuthType = "anonymous"
)

var enumValues_AuthType = []AuthType{
	AuthTypeCredentials, AuthTypeAnonymous,
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

// AuthCredentials represent the authentication credentials information.
type AuthCredentials struct {
	// The authentication type
	Type AuthType `json:"type" mapstructure:"type" yaml:"type"`
	// The given service account or refresh token JSON credentials in JSON string format.
	Credentials *utils.EnvString `json:"credentials,omitempty" mapstructure:"credentials" yaml:"credentials,omitempty"`
	// The given service account or refresh token JSON credentials file.
	CredentialsFile *utils.EnvString `json:"credentialsFile,omitempty" mapstructure:"credentialsFile" yaml:"credentialsFile,omitempty"`
}

// JSONSchema is used to generate a custom jsonschema.
func (ac AuthCredentials) JSONSchema() *jsonschema.Schema {
	envStringRefName := "#/$defs/EnvString"

	credProps := jsonschema.NewProperties()
	credProps.Set("type", &jsonschema.Schema{
		Type:        "string",
		Description: "Authorize with a service account or refresh token JSON credentials",
		Enum:        []any{AuthTypeCredentials},
	})
	credProps.Set("credentials", &jsonschema.Schema{
		Description: "The given service account or refresh token JSON credentials in JSON string format",
		Ref:         envStringRefName,
	})
	credProps.Set("credentialsFile", &jsonschema.Schema{
		Description: "The given service account or refresh token JSON credentials file",
		Ref:         envStringRefName,
	})

	anonymousProps := jsonschema.NewProperties()
	anonymousProps.Set("type", &jsonschema.Schema{
		Type: "string",
		Enum: []any{AuthTypeAnonymous},
	})

	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Type:       "object",
				Properties: credProps,
				Required:   []string{"type"},
				OneOf: []*jsonschema.Schema{
					{Required: []string{"credentials"}},
					{Required: []string{"credentialsFile"}},
				},
			},
			{
				Type:       "object",
				Properties: anonymousProps,
				Required:   []string{"type"},
			},
		},
	}
}

func (ac AuthCredentials) toCredentials() (option.ClientOption, error) {
	switch ac.Type {
	case AuthTypeAnonymous:
		return option.WithoutAuthentication(), nil
	case AuthTypeCredentials:
		return ac.parseServiceAccount()
	default:
		return nil, fmt.Errorf("unsupported auth type %s", ac.Type)
	}
}

func (ac AuthCredentials) parseServiceAccount() (option.ClientOption, error) {
	if ac.Credentials == nil && ac.CredentialsFile == nil {
		return nil, errRequireCredentials
	}

	if ac.Credentials != nil {
		strCred, err := ac.Credentials.GetOrDefault("")
		if err != nil {
			return nil, err
		}

		if strCred != "" {
			return option.WithCredentialsJSON([]byte(strCred)), nil
		}
	}

	credPath, err := ac.CredentialsFile.GetOrDefault("")
	if err != nil {
		return nil, err
	}

	if credPath == "" {
		return nil, errRequireCredentials
	}

	return option.WithCredentialsFile(credPath), nil
}