package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/storage/minio"
	"github.com/invopop/jsonschema"
)

var (
	errRequireAccessKeyID     = errors.New("accessKeyId is required")
	errRequireSecretAccessKey = errors.New("secretAccessKey is required")
	errRequireStorageEndpoint = errors.New("endpoint is required")
)

// Client wraps the storage client with additional information.
type Client struct {
	id                     common.StorageClientID
	defaultBucket          string
	defaultPresignedExpiry *time.Duration
	allowedBuckets         []string

	common.StorageClient
}

// ValidateBucket checks if the bucket name is valid, or returns the default bucket if empty.
func (c *Client) ValidateBucket(key string) (string, error) {
	if key != "" {
		if key == c.defaultBucket || len(c.allowedBuckets) == 0 || slices.Contains(c.allowedBuckets, key) {
			return key, nil
		}

		return "", schema.UnprocessableContentError(fmt.Sprintf("you are not allowed to access `%s` bucket, client id `%s`", key, c.id), nil)
	}

	if c.defaultBucket == "" {
		return "", schema.UnprocessableContentError("bucket name is required", nil)
	}

	return c.defaultBucket, nil
}

// EnvStorageProviderType represents the env configuration for the storage provider type.
type EnvStorageProviderType struct {
	utils.EnvString `yaml:",inline"`
}

// Validate checks if the configration is valid.
func (espt EnvStorageProviderType) Validate() (common.StorageProviderType, error) {
	rawProviderType, err := espt.EnvString.GetOrDefault("")
	if err != nil {
		return "", err
	}

	providerType, err := common.ParseStorageProviderType(rawProviderType)
	if err != nil {
		return "", err
	}

	return providerType, nil
}

// JSONSchema is used to generate a custom jsonschema.
func (espt EnvStorageProviderType) JSONSchema() *jsonschema.Schema {
	result := &jsonschema.Schema{
		Type:       "object",
		Properties: jsonschema.NewProperties(),
		AnyOf: []*jsonschema.Schema{
			{
				Required: []string{"value"},
			},
			{
				Required: []string{"env"},
			},
		},
	}

	result.Properties.Set("env", &jsonschema.Schema{
		Type: "string",
	})

	result.Properties.Set("value", common.StorageProviderType("").JSONSchema())

	return result
}

// ClientConfig represent the raw configuration of a storage provider client.
type ClientConfig struct {
	// The unique identity of a client. Use this setting if there are many configured clients.
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	// Cloud provider type of the storage client.
	Type EnvStorageProviderType `json:"type" yaml:"type"`
	// Default bucket name to be set if the user doesn't specify any bucket.
	DefaultBucket utils.EnvString `json:"defaultBucket" yaml:"defaultBucket"`
	// Endpoint of the storage server. Required for other S3 compatible services such as MinIO, Cloudflare R2, DigitalOcean Spaces, etc...
	Endpoint *utils.EnvString `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	// The public host to be used for presigned URL generation.
	PublicHost *utils.EnvString `json:"publicHost,omitempty" yaml:"publicHost,omitempty"`
	// Optional region.
	Region *utils.EnvString `json:"region,omitempty" jsonschema:"nullable" yaml:"region,omitempty"`
	// Maximum number of retry times.
	MaxRetries *int `json:"maxRetries,omitempty" jsonschema:"min=1,default=10" yaml:"maxRetries,omitempty"`
	// The default expiry for presigned URL generation. The maximum expiry is 604800 seconds (i.e. 7 days) and minimum is 1 second.
	DefaultPresignedExpiry *string `json:"defaultPresignedExpiry,omitempty" jsonschema:"pattern=[0-9]+(s|m|h),default=24h" yaml:"defaultPresignedExpiry,omitempty"`
	// Authentication credetials.
	Authentication AuthCredentials `json:"authentication" yaml:"authentication"`
	// TrailingHeaders indicates server support of trailing headers.
	// Only supported for v4 signatures.
	TrailingHeaders bool `json:"trailingHeaders,omitempty" yaml:"trailingHeaders,omitempty"`
	// Allowed buckets. This setting prevents users to get buckets and objects outside the list.
	// However, it's recommended to restrict the permissions for the IAM credentials.
	// This setting is useful to let the connector know which buckets belong to this client.
	// The empty value means all buckets are allowed. The storage server will handle the validation.
	AllowedBuckets []string `json:"allowedBuckets,omitempty" yaml:"allowedBuckets,omitempty"`
}

// Validate checks if the configration is valid.
func (sc ClientConfig) Validate() error {
	providerType, err := sc.Type.Validate()
	if err != nil {
		return fmt.Errorf("type: %w", err)
	}

	switch providerType {
	case common.S3, common.GoogleStorage:
		_, err := sc.toMinioConfig(providerType)
		if err != nil && !errors.Is(err, errRequireAccessKeyID) && !errors.Is(err, errRequireSecretAccessKey) && !errors.Is(err, errRequireStorageEndpoint) {
			return err
		}
	}

	return nil
}

// ToStorageClient validates and create the storage client from config.
func (sc ClientConfig) ToStorageClient(ctx context.Context, logger *slog.Logger) (common.StorageClient, error) {
	providerType, err := sc.Type.Validate()
	if err != nil {
		return nil, fmt.Errorf("type: %w", err)
	}

	switch providerType {
	case common.S3, common.GoogleStorage:
		return sc.toMinioClient(ctx, logger, providerType)
	}

	return nil, errors.New("unsupported storage client: " + string(providerType))
}

func (sc ClientConfig) toMinioClient(ctx context.Context, logger *slog.Logger, providerType common.StorageProviderType) (common.StorageClient, error) {
	config, err := sc.toMinioConfig(providerType)
	if err != nil {
		return nil, err
	}

	return minio.New(ctx, config, logger)
}

func (sc ClientConfig) toMinioConfig(providerType common.StorageProviderType) (*minio.ClientConfig, error) {
	endpoint, port, useSSL, err := sc.parseEndpoint()
	if err != nil {
		return nil, err
	}

	if endpoint == "" {
		switch providerType {
		case common.S3:
			endpoint = "s3.amazonaws.com"
			useSSL = true
		case common.GoogleStorage:
			endpoint = "storage.googleapis.com"
			useSSL = true
		default:
			return nil, errRequireStorageEndpoint
		}
	}

	result := &minio.ClientConfig{
		Type:            providerType,
		Endpoint:        endpoint,
		Secure:          useSSL,
		Port:            port,
		TrailingHeaders: sc.TrailingHeaders,
	}

	creds, err := sc.Authentication.toMinioAuthConfig()
	if err != nil {
		return nil, err
	}

	result.AuthConfig = *creds

	if sc.MaxRetries != nil {
		maxRetries := *sc.MaxRetries
		if maxRetries <= -1 {
			maxRetries = 1
		}

		result.MaxRetries = maxRetries
	}

	if sc.PublicHost != nil {
		publicHost, err := sc.PublicHost.GetOrDefault("")
		if err != nil {
			return nil, fmt.Errorf("publicHost: %w", err)
		}

		if strings.HasPrefix(publicHost, "http") {
			result.PublicHost, err = url.Parse(publicHost)
			if err != nil {
				return nil, fmt.Errorf("publicHost: %w", err)
			}
		} else {
			result.PublicHost = &url.URL{
				Host: publicHost,
			}
		}
	}

	if sc.Region != nil {
		result.Region, err = sc.Region.GetOrDefault("")
		if err != nil {
			return nil, fmt.Errorf("region: %w", err)
		}
	}

	return result, nil
}

func (sc ClientConfig) parseEndpoint() (string, int, bool, error) {
	port := 80
	if sc.Endpoint == nil {
		return "", port, false, nil
	}

	var endpoint string
	var useSSL bool

	rawEndpoint, err := sc.Endpoint.GetOrDefault("")
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

// AuthType represents the authentication type enum.
type AuthType string

const (
	AuthTypeStatic = "static"
	AuthTypeIAM    = "iam"
)

var enumValues_AuthType = []AuthType{
	AuthTypeStatic, AuthTypeIAM,
}

// ParseStorageProviderType parses the StorageProviderType from string.
func ParseStorageProviderType(input string) (AuthType, error) {
	result := AuthType(input)
	if !slices.Contains(enumValues_AuthType, result) {
		return "", fmt.Errorf("invalid AuthType, expected one of %v, got: %s", enumValues_AuthType, input)
	}

	return result, nil
}

// Validate checks if the provider type is valid.
func (spt AuthType) Validate() error {
	_, err := ParseStorageProviderType(string(spt))

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

func (ac AuthCredentials) toMinioAuthConfig() (*minio.AuthConfig, error) {
	switch ac.Type {
	case AuthTypeIAM:
		return ac.parseIAMAuth()
	case AuthTypeStatic:
		return ac.parseStaticAccessIDSecret()
	default:
		return nil, fmt.Errorf("unsupported auth type %s", ac.Type)
	}
}

func (ac AuthCredentials) parseIAMAuth() (*minio.AuthConfig, error) {
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

	return &minio.AuthConfig{
		UseIAMAuth:      true,
		IAMAuthEndpoint: rawIAMEndpoint,
	}, nil
}

func (ac AuthCredentials) parseStaticAccessIDSecret() (*minio.AuthConfig, error) {
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

	return &minio.AuthConfig{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
	}, nil
}

// FormatTimestamp formats the Time value to string
func FormatTimestamp(value time.Time) string {
	return value.Format(time.RFC3339)
}
