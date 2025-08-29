package minio

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"

	"github.com/hasura/ndc-http/exhttp"
	"github.com/hasura/ndc-sdk-go/v2/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/invopop/jsonschema"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	errRequireAccessKeyID     = errors.New("accessKeyId is required")
	errRequireSecretAccessKey = errors.New("secretAccessKey is required")
	errRequireStorageEndpoint = errors.New("endpoint is required")
)

// ClientConfig represent the raw configuration of a MinIO client.
type ClientConfig struct {
	common.BaseClientConfig `yaml:",inline"`
	OtherConfig             `yaml:",inline"`
}

// JSONSchema is used to generate a custom jsonschema.
func (cc ClientConfig) JSONSchema() *jsonschema.Schema {
	envStringRef := "#/$defs/EnvString"

	result := cc.GetJSONSchema([]any{common.StorageProviderTypeS3})
	result.Required = append(result.Required, "authentication")

	result.Properties.Set("region", &jsonschema.Schema{
		Description: "Optional region",
		OneOf: []*jsonschema.Schema{
			{Type: "null"},
			{Ref: envStringRef},
		},
	})
	result.Properties.Set("publicHost", &jsonschema.Schema{
		Description: "The public host to be used for presigned URL generation",
		Ref:         envStringRef,
	})
	result.Properties.Set("authentication", cc.Authentication.JSONSchema())
	result.Properties.Set("trailingHeaders", &jsonschema.Schema{
		Description: "TrailingHeaders indicates server support of trailing headers. Only supported for v4 signatures",
		Type:        "boolean",
	})
	result.Properties.Set("http", &jsonschema.Schema{
		Ref: "#/$defs/HTTPTransportTLSConfig",
	})

	return result
}

// OtherConfig holds MinIO-specific configurations.
type OtherConfig struct {
	// The public host to be used for presigned URL generation.
	PublicHost *utils.EnvString `json:"publicHost,omitempty"      mapstructure:"publicHost"      yaml:"publicHost,omitempty"`
	// Optional region.
	Region *utils.EnvString `json:"region,omitempty"          mapstructure:"region"          yaml:"region,omitempty"`
	// Authentication credentials.
	Authentication AuthCredentials `json:"authentication"            mapstructure:"authentication"  yaml:"authentication"`
	// TrailingHeaders indicates server support of trailing headers.
	// Only supported for v4 signatures.
	TrailingHeaders bool `json:"trailingHeaders,omitempty" mapstructure:"trailingHeaders" yaml:"trailingHeaders,omitempty"`
	// Configuration for the http client that is used for uploading files from URL.
	HTTP *exhttp.HTTPTransportTLSConfig `json:"http"                      mapstructure:"http"            yaml:"http"`
}

func (cc ClientConfig) toMinioOptions(
	providerType common.StorageProviderType,
	logger *slog.Logger,
) (*minio.Options, string, error) {
	endpointURL, port, useSSL, err := cc.ValidateEndpoint()
	if err != nil {
		return nil, "", err
	}

	var endpoint string

	if endpointURL == nil {
		switch providerType {
		case common.StorageProviderTypeS3:
			endpoint = "s3.amazonaws.com"
			useSSL = true
		case common.StorageProviderTypeGcs:
			endpoint = "storage.googleapis.com"
			useSSL = true
		default:
			return nil, "", errRequireStorageEndpoint
		}
	} else {
		endpoint = endpointURL.Host
	}

	transport, err := common.NewTransport(cc.HTTP, exhttp.TelemetryConfig{
		Logger: logger,
		Port:   port,
	})
	if err != nil {
		return nil, "", err
	}

	opts := &minio.Options{
		Secure:          useSSL,
		Transport:       transport,
		TrailingHeaders: cc.TrailingHeaders,
	}

	opts.Creds, err = cc.Authentication.toCredentials()
	if err != nil {
		return nil, "", err
	}

	if cc.MaxRetries != nil {
		maxRetries := *cc.MaxRetries
		if maxRetries <= -1 {
			maxRetries = 1
		}

		opts.MaxRetries = maxRetries
	}

	if cc.Region != nil {
		opts.Region, err = cc.Region.GetOrDefault("")
		if err != nil {
			return nil, "", fmt.Errorf("region: %w", err)
		}
	}

	return opts, endpoint, nil
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
	AuthTypeStatic AuthType = "static"
	AuthTypeIAM    AuthType = "iam"
)

var enumValues_AuthType = []AuthType{
	AuthTypeStatic, AuthTypeIAM,
}

// ParseAuthType parses the AuthType from string.
func ParseAuthType(input string) (AuthType, error) {
	result := AuthType(input)
	if !slices.Contains(enumValues_AuthType, result) {
		return "", fmt.Errorf(
			"invalid AuthType, expected one of %v, got: %s",
			enumValues_AuthType,
			input,
		)
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
	Type AuthType `json:"type"                      mapstructure:"type"            yaml:"type"`
	// Access Key ID.
	AccessKeyID *utils.EnvString `json:"accessKeyId,omitempty"     mapstructure:"accessKeyId"     yaml:"accessKeyId,omitempty"`
	// Secret Access Key.
	SecretAccessKey *utils.EnvString `json:"secretAccessKey,omitempty" mapstructure:"secretAccessKey" yaml:"secretAccessKey,omitempty"`
	// Optional temporary session token credentials. Used for testing only.
	// See https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_use-resources.html
	SessionToken *utils.EnvString `json:"sessionToken,omitempty"    mapstructure:"sessionToken"    yaml:"sessionToken,omitempty"`
	// Custom endpoint to fetch IAM role credentials.
	IAMAuthEndpoint *utils.EnvString `json:"iamAuthEndpoint,omitempty" mapstructure:"iamAuthEndpoint" yaml:"iamAuthEndpoint,omitempty"`
}

// JSONSchema is used to generate a custom jsonschema.
func (ac AuthCredentials) JSONSchema() *jsonschema.Schema {
	envStringRef := &jsonschema.Schema{
		Ref: "#/$defs/EnvString",
	}

	staticProps := jsonschema.NewProperties()
	staticProps.Set("type", &jsonschema.Schema{
		Type: "string",
		Enum: []any{AuthTypeStatic},
	})
	staticProps.Set("accessKeyId", envStringRef)
	staticProps.Set("secretAccessKey", envStringRef)
	staticProps.Set("sessionToken", envStringRef)

	iamProps := jsonschema.NewProperties()
	iamProps.Set("type", &jsonschema.Schema{
		Type: "string",
		Enum: []any{AuthTypeIAM},
	})
	iamProps.Set("iamAuthEndpoint", envStringRef)

	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Type:       "object",
				Properties: staticProps,
				Required:   []string{"type", "accessKeyId", "secretAccessKey"},
			},
			{
				Type:       "object",
				Properties: iamProps,
				Required:   []string{"type"},
			},
		},
	}
}

func (ac AuthCredentials) toCredentials() (*credentials.Credentials, error) {
	switch ac.Type {
	case AuthTypeIAM:
		return ac.parseIAMAuth()
	case AuthTypeStatic:
		return ac.parseStaticAccessIDSecret()
	default:
		return nil, fmt.Errorf("unsupported auth type %s", ac.Type)
	}
}

func (ac AuthCredentials) parseIAMAuth() (*credentials.Credentials, error) {
	rawIAMEndpoint, err := ac.IAMAuthEndpoint.GetOrDefault("")
	if err != nil {
		return nil, fmt.Errorf("iamAuthEndpoint: %w", err)
	}

	if rawIAMEndpoint != "" {
		iamEndpoint, err := url.Parse(rawIAMEndpoint)
		if err != nil {
			return nil, fmt.Errorf("iamAuthEndpoint: %w", err)
		}

		if !strings.HasPrefix(iamEndpoint.Scheme, "http") {
			return nil, errors.New("iamAuthEndpoint: invalid http scheme " + iamEndpoint.Scheme)
		}
	}

	return credentials.NewIAM(rawIAMEndpoint), nil
}

func (ac AuthCredentials) parseStaticAccessIDSecret() (*credentials.Credentials, error) {
	if ac.AccessKeyID == nil {
		return nil, errRequireAccessKeyID
	}

	if ac.SecretAccessKey == nil {
		return nil, errRequireSecretAccessKey
	}

	accessKeyID, err := ac.AccessKeyID.GetOrDefault("")
	if err != nil {
		return nil, fmt.Errorf("accessKeyID: %w", err)
	}

	if accessKeyID == "" {
		return nil, errRequireAccessKeyID
	}

	secretAccessKey, err := ac.SecretAccessKey.GetOrDefault("")
	if err != nil {
		return nil, fmt.Errorf("secretAccessKey: %w", err)
	}

	if secretAccessKey == "" {
		return nil, errRequireSecretAccessKey
	}

	var sessionToken string
	if ac.SessionToken != nil {
		sessionToken, err = ac.SessionToken.GetOrDefault("")
		if err != nil {
			return nil, fmt.Errorf("sessionToken: %w", err)
		}
	}

	return credentials.NewStaticV4(accessKeyID, secretAccessKey, sessionToken), nil
}
