package minio

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"

	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/invopop/jsonschema"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.opentelemetry.io/otel"
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

	result := cc.BaseClientConfig.GetJSONSchema([]any{common.S3, common.GoogleStorage})
	result.Required = append(result.Required, "authentication")

	result.Properties.Set("region", &jsonschema.Schema{
		Description: "Optional region",
		Ref:         envStringRef,
	})
	result.Properties.Set("authentication", cc.Authentication.JSONSchema())
	result.Properties.Set("trailingHeaders", &jsonschema.Schema{
		Description: "TrailingHeaders indicates server support of trailing headers. Only supported for v4 signatures",
		Type:        "boolean",
	})

	return result
}

// OtherConfig holds MinIO-specific configurations
type OtherConfig struct {
	// Optional region.
	Region *utils.EnvString `json:"region,omitempty" jsonschema:"nullable" yaml:"region,omitempty"`
	// Authentication credentials.
	Authentication AuthCredentials `json:"authentication" yaml:"authentication"`
	// TrailingHeaders indicates server support of trailing headers.
	// Only supported for v4 signatures.
	TrailingHeaders bool `json:"trailingHeaders,omitempty" yaml:"trailingHeaders,omitempty"`
}

func (cc ClientConfig) toMinioOptions(providerType common.StorageProviderType, logger *slog.Logger) (*minio.Options, string, error) {
	endpoint, port, useSSL, err := cc.BaseClientConfig.ValidateEndpoint()
	if err != nil {
		return nil, "", err
	}

	if endpoint == "" {
		switch providerType {
		case common.S3:
			endpoint = "s3.amazonaws.com"
			useSSL = true
		case common.GoogleStorage:
			endpoint = "storage.googleapis.com"
			useSSL = true
		case common.AzureBlobStore:
			return nil, "", errRequireStorageEndpoint
		default:
			return nil, "", errRequireStorageEndpoint
		}
	}

	transport, err := minio.DefaultTransport(useSSL)
	if err != nil {
		return nil, "", err
	}

	opts := &minio.Options{
		Secure:          useSSL,
		Transport:       transport,
		TrailingHeaders: cc.TrailingHeaders,
	}

	if utils.IsDebug(logger) {
		opts.Transport = debugRoundTripper{
			transport:  transport,
			propagator: otel.GetTextMapPropagator(),
			port:       port,
			logger:     logger,
		}
	} else {
		opts.Transport = roundTripper{
			transport:  transport,
			propagator: otel.GetTextMapPropagator(),
		}
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
	AccessKeyID *utils.EnvString `json:"accessKeyId,omitempty" yaml:"accessKeyId,omitempty"`
	// Secret Access Key.
	SecretAccessKey *utils.EnvString `json:"secretAccessKey,omitempty" yaml:"secretAccessKey,omitempty"`
	// Optional temporary session token credentials. Used for testing only.
	// See https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_use-resources.html
	SessionToken *utils.EnvString `json:"sessionToken,omitempty" yaml:"sessionToken,omitempty"`
	// Custom endpoint to fetch IAM role credentials.
	IAMAuthEndpoint *utils.EnvString `json:"iamAuthEndpoint,omitempty" yaml:"iamAuthEndpoint,omitempty"`
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
