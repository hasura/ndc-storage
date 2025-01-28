package azblob

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/tracing/azotel"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/invopop/jsonschema"
	"go.opentelemetry.io/otel"
)

var (
	errRequireAccountName      = errors.New("accountName is required")
	errRequireAccountKey       = errors.New("accountKey is required")
	errRequireStorageEndpoint  = errors.New("endpoint is required")
	errRequireConnectionString = errors.New("azure connection string is required")
	errRequireAuthTenantID     = errors.New("azure authentication.tenantId is required")
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
	endpointURL, port, useSSL, err := cc.BaseClientConfig.ValidateEndpoint()
	if err != nil {
		return nil, err
	}

	maxRetries := 0
	if cc.MaxRetries != nil && *cc.MaxRetries > 0 {
		maxRetries = *cc.MaxRetries
	}

	isDebug := utils.IsDebug(logger)
	transport := common.NewTransport(logger, port, true)

	opts := &azblob.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Retry: policy.RetryOptions{
				MaxRetries: int32(maxRetries),
			},
			Logging: policy.LogOptions{
				IncludeBody: isDebug,
			},
			InsecureAllowCredentialWithHTTP: !useSSL,
			TracingProvider:                 azotel.NewTracingProvider(otel.GetTracerProvider(), nil),
			Transport: &http.Client{
				Transport: transport,
			},
		},
	}

	var endpoint string
	if endpointURL != nil {
		endpoint = endpointURL.String()
	}

	return cc.Authentication.toAzureBlobClient(endpoint, opts)
}

// OtherConfig holds MinIO-specific configurations
type OtherConfig struct {
	// Authentication credentials.
	Authentication AuthCredentials `json:"authentication" mapstructure:"authentication" yaml:"authentication"`
}

// AuthType represents the authentication type enum.
type AuthType string

const (
	AuthTypeAnonymous        AuthType = "anonymous"
	AuthTypeSharedKey        AuthType = "sharedKey"
	AuthTypeEntra            AuthType = "entra"
	AuthTypeConnectionString AuthType = "connectionString"
)

var enumValues_AuthType = []AuthType{
	AuthTypeAnonymous,
	AuthTypeSharedKey,
	AuthTypeEntra,
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

// AuthCredentials represent the authentication credentials information.
type AuthCredentials struct {
	// The authentication type
	Type AuthType `json:"type" mapstructure:"type" yaml:"type"`
	// Access Key ID.
	AccountName *utils.EnvString `json:"accountName,omitempty" mapstructure:"accountName" yaml:"accountName,omitempty"`
	// Secret Access Key.
	AccountKey *utils.EnvString `json:"accountKey,omitempty" mapstructure:"accountKey" yaml:"accountKey,omitempty"`
	// Connection String.
	ConnectionString *utils.EnvString `json:"connectionString,omitempty" mapstructure:"connectionString" yaml:"connectionString,omitempty"`
	// Azure tenant ID.
	TenantID *utils.EnvString `json:"tenantId,omitempty" mapstructure:"tenantId" yaml:"tenantId,omitempty"`
	// The service principal's client ID.
	ClientID *utils.EnvString `json:"clientId,omitempty" mapstructure:"clientId" yaml:"clientId,omitempty"`
	// One of the service principal's client secrets.
	ClientSecret *utils.EnvString `json:"clientSecret,omitempty" mapstructure:"clientSecret" yaml:"clientSecret,omitempty"`
	// The username (usually an email address).
	Username *utils.EnvString `json:"username,omitempty" mapstructure:"username" yaml:"username,omitempty"`
	// The user's password.
	Password *utils.EnvString `json:"password,omitempty" mapstructure:"password" yaml:"password,omitempty"`
	// Inline PEM or PKCS12 certificate of the private key in base64 format.
	ClientCertificate *utils.EnvString `json:"clientCertificate,omitempty" mapstructure:"clientCertificate" yaml:"clientCertificate,omitempty"`
	// Path to a PEM or PKCS12 certificate file including the private key.
	ClientCertificatePath *utils.EnvString `json:"clientCertificatePath,omitempty" mapstructure:"clientCertificatePath" yaml:"clientCertificatePath,omitempty"`
	// Optional password for the certificate
	ClientCertificatePassword *utils.EnvString `json:"clientCertificatePassword,omitempty" mapstructure:"clientCertificatePassword" yaml:"clientCertificatePassword,omitempty"`
	// SendCertificateChain controls whether the credential sends the public certificate chain in the x5c header of each token request's JWT.
	// This is required for Subject Name/Issuer (SNI) authentication. Defaults to False.
	SendCertificateChain bool `json:"sendCertificateChain,omitempty" mapstructure:"sendCertificateChain" yaml:"sendCertificateChain,omitempty"`
	// TokenFilePath is the path of a file containing a Kubernetes service account token.
	TokenFilePath *utils.EnvString `json:"tokenFilePath,omitempty" mapstructure:"tokenFilePath" yaml:"tokenFilePath,omitempty"`
	// DisableInstanceDiscovery should be set true only by applications authenticating in disconnected clouds, or
	// private clouds such as Azure Stack. It determines whether the credential requests Microsoft Entra instance metadata
	// from https://login.microsoft.com before authenticating. Setting this to true will skip this request, making
	// the application responsible for ensuring the configured authority is valid and trustworthy.
	DisableInstanceDiscovery bool `json:"disableInstanceDiscovery,omitempty" mapstructure:"disableInstanceDiscovery" yaml:"disableInstanceDiscovery,omitempty"`
	// Enable multitenant authentication. The credential may request tokens from in addition to the tenant specified by AZURE_TENANT_ID.
	// Set this value to "*" to enable the credential to request a token from any tenant.
	AdditionallyAllowedTenants []string `json:"additionallyAllowedTenants,omitempty" mapstructure:"additionallyAllowedTenants" yaml:"additionallyAllowedTenants,omitempty"`
	// Audience to use when requesting tokens for Azure Active Directory authentication.
	Audience *utils.EnvString `json:"audience,omitempty" mapstructure:"audience" yaml:"audience,omitempty"`
}

// JSONSchema is used to generate a custom jsonschema.
func (ac AuthCredentials) JSONSchema() *jsonschema.Schema { //nolint:funlen
	envStringRefName := "#/$defs/EnvString"
	staticProps := jsonschema.NewProperties()
	staticProps.Set("type", &jsonschema.Schema{
		Type:        "string",
		Description: "Authorize with an immutable SharedKeyCredential containing the storage account's name and either its primary or secondary key",
		Enum:        []any{AuthTypeSharedKey},
	})
	staticProps.Set("accountName", &jsonschema.Schema{
		Description: "Account Name",
		Ref:         envStringRefName,
	})
	staticProps.Set("accountKey", &jsonschema.Schema{
		Description: "Account Key",
		Ref:         envStringRefName,
	})

	anonymousProps := jsonschema.NewProperties()
	anonymousProps.Set("type", &jsonschema.Schema{
		Type: "string",
		Enum: []any{AuthTypeAnonymous},
	})

	connStringProps := jsonschema.NewProperties()
	connStringProps.Set("type", &jsonschema.Schema{
		Type:        "string",
		Description: "Authorize with a connection string for the desired storage account",
		Enum:        []any{AuthTypeConnectionString},
	})
	connStringProps.Set("connectionString", &jsonschema.Schema{
		Description: "The connection string",
		Ref:         envStringRefName,
	})

	entraProps := jsonschema.NewProperties()
	entraProps.Set("type", &jsonschema.Schema{
		Type: "string",
		Enum: []any{AuthTypeEntra},
	})

	entraProps.Set("tenantId", &jsonschema.Schema{
		Description: "ID of the service principal's tenant. Also called its `directory` ID",
		Ref:         envStringRefName,
	})
	entraProps.Set("clientId", &jsonschema.Schema{
		Description: "The service principal's client ID",
		Ref:         envStringRefName,
	})
	entraProps.Set("clientSecret", &jsonschema.Schema{
		Description: "One of the service principal's client secrets",
		Ref:         envStringRefName,
	})
	entraProps.Set("username", &jsonschema.Schema{
		Description: "The username (usually an email address)",
		Ref:         envStringRefName,
	})

	entraProps.Set("password", &jsonschema.Schema{
		Description: "The user's password",
		Ref:         envStringRefName,
	})

	entraProps.Set("clientCertificate", &jsonschema.Schema{
		Description: "Inline PEM or PKCS12 certificate of the private key in base64 format",
		Ref:         envStringRefName,
	})
	entraProps.Set("clientCertificatePath", &jsonschema.Schema{
		Description: "Path to a PEM or PKCS12 certificate file including the private key",
		Ref:         envStringRefName,
	})
	entraProps.Set("clientCertificatePassword", &jsonschema.Schema{
		Description: "Optional password for the certificate",
		Ref:         envStringRefName,
	})
	entraProps.Set("sendCertificateChain", &jsonschema.Schema{
		Description: "Controls whether the credential sends the public certificate chain in the x5c header of each token request's JWT",
		Type:        "boolean",
	})
	entraProps.Set("tokenFilePath", &jsonschema.Schema{
		Description: "the path of a file containing a Kubernetes service account token",
		Ref:         envStringRefName,
	})
	entraProps.Set("audience", &jsonschema.Schema{
		Description: "Audience to use when requesting tokens for Azure Active Directory authentication",
		Type:        envStringRefName,
	})

	entraProps.Set("disableInstanceDiscovery", &jsonschema.Schema{
		Type: "boolean",
	})
	entraProps.Set("additionallyAllowedTenants", &jsonschema.Schema{
		Type: "array",
		Items: &jsonschema.Schema{
			Type: "string",
		},
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
				Properties: connStringProps,
				Required:   []string{"type", "connectionString"},
			},
			{
				Type:       "object",
				Properties: anonymousProps,
				Required:   []string{"type"},
			},
			{
				Type:       "object",
				Properties: entraProps,
				Required:   []string{"type", "tenantId"},
				DependentRequired: map[string][]string{
					"clientSecret": {"clientId"},
					"username":     {"password"},
					"password":     {"username"},
				},
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
	case AuthTypeAnonymous:
		if serviceURL == "" {
			return nil, errRequireStorageEndpoint
		}

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
	case AuthTypeEntra:
		cred, err := ac.toDefaultAzureCredential(options)
		if err != nil {
			return nil, err
		}

		if ac.Audience != nil {
			audience, err := ac.Audience.GetOrDefault("")
			if err != nil {
				return nil, fmt.Errorf("audience: %w", err)
			}

			options.Audience = audience
		}

		return azblob.NewClient(serviceURL, cred, options)
	case AuthTypeConnectionString:
		if ac.ConnectionString == nil {
			return nil, errRequireConnectionString
		}

		connString, err := ac.ConnectionString.Get()
		if err != nil {
			return nil, fmt.Errorf("failed to get azure connection string: %w", err)
		}

		if connString == "" {
			return nil, errRequireConnectionString
		}

		return azblob.NewClientFromConnectionString(connString, options)
	default:
		return nil, fmt.Errorf("unsupported auth type %s", ac.Type)
	}
}

func (ac AuthCredentials) parseAccountNameAndKey() (string, string, error) {
	var accountName, accountKey string
	var err error

	if ac.AccountName != nil {
		accountName, err = ac.AccountName.GetOrDefault("")
		if err != nil {
			return "", "", fmt.Errorf("accountKey: %w", err)
		}
	}

	if ac.AccountKey != nil {
		accountKey, err = ac.AccountKey.GetOrDefault("")
		if err != nil {
			return "", "", fmt.Errorf("accountKey: %w", err)
		}
	}

	return accountName, accountKey, nil
}

// toDefaultAzureCredential creates a DefaultAzureCredential. Pass nil for options to accept defaults.
func (ac AuthCredentials) toDefaultAzureCredential(options *azblob.ClientOptions) (azcore.TokenCredential, error) {
	var creds []azcore.TokenCredential

	if ac.TenantID == nil {
		return nil, errRequireAuthTenantID
	}

	tenantID, err := ac.TenantID.Get()
	if err != nil {
		return nil, err
	}

	if tenantID == "" {
		return nil, errRequireAuthTenantID
	}

	var clientID string
	if ac.ClientID != nil {
		clientID, err = ac.ClientID.Get()
		if err != nil {
			return nil, err
		}
	}

	secretCred, err := ac.toClientSecretCredential(tenantID, clientID, options)
	if err != nil {
		return nil, err
	}

	if secretCred != nil {
		creds = append(creds, secretCred)
	}

	certCred, err := ac.toCertificateCredential(tenantID, clientID, options)
	if err != nil {
		return nil, err
	}

	if certCred != nil {
		creds = append(creds, certCred)
	}

	userPassCred, err := ac.toUsernamePasswordCredential(tenantID, clientID, options)
	if err != nil {
		return nil, err
	}

	if userPassCred != nil {
		creds = append(creds, userPassCred)
	}

	wic, err := ac.toWorkloadIdentityCredential(tenantID, clientID, options)
	if err != nil {
		return nil, err
	}

	if userPassCred != nil {
		creds = append(creds, wic)
	}

	o := &azidentity.ManagedIdentityCredentialOptions{
		ClientOptions: options.ClientOptions,
	}

	if clientID != "" {
		o.ID = azidentity.ClientID(clientID)
	}

	miCred, err := azidentity.NewManagedIdentityCredential(o)
	if err == nil {
		creds = append(creds, miCred)
	} else {
		slog.Warn("failed to initialize managed identity credential: " + err.Error())
	}

	return azidentity.NewChainedTokenCredential(creds, nil)
}

func (ac AuthCredentials) toClientSecretCredential(tenantID, clientID string, options *azblob.ClientOptions) (azcore.TokenCredential, error) {
	if clientID == "" || ac.ClientSecret == nil {
		return nil, nil
	}

	clientSecret, err := ac.ClientSecret.GetOrDefault("")
	if err != nil {
		return nil, err
	}

	if clientSecret == "" {
		return nil, nil
	}

	o := &azidentity.ClientSecretCredentialOptions{
		AdditionallyAllowedTenants: ac.AdditionallyAllowedTenants,
		ClientOptions:              options.ClientOptions,
		DisableInstanceDiscovery:   ac.DisableInstanceDiscovery,
	}

	return azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, o)
}

func (ac AuthCredentials) toWorkloadIdentityCredential(tenantID, clientID string, options *azblob.ClientOptions) (azcore.TokenCredential, error) {
	if clientID == "" || ac.TokenFilePath == nil {
		return nil, nil
	}

	tokenFilePath, err := ac.TokenFilePath.GetOrDefault("")
	if err != nil {
		return nil, err
	}

	if tokenFilePath == "" {
		return nil, nil
	}

	return azidentity.NewWorkloadIdentityCredential(&azidentity.WorkloadIdentityCredentialOptions{
		ClientOptions:              options.ClientOptions,
		AdditionallyAllowedTenants: ac.AdditionallyAllowedTenants,
		DisableInstanceDiscovery:   ac.DisableInstanceDiscovery,
		ClientID:                   clientID,
		TenantID:                   tenantID,
		TokenFilePath:              tokenFilePath,
	})
}

func (ac AuthCredentials) toUsernamePasswordCredential(tenantID, clientID string, options *azblob.ClientOptions) (azcore.TokenCredential, error) {
	if clientID == "" || ac.Username == nil {
		return nil, nil
	}

	username, err := ac.Username.GetOrDefault("")
	if err != nil {
		return nil, err
	}

	if username == "" {
		return nil, nil
	}

	if ac.Password == nil {
		return nil, errors.New("password is required if the username is set")
	}

	password, err := ac.Password.GetOrDefault("")
	if err != nil {
		return nil, err
	}

	o := &azidentity.UsernamePasswordCredentialOptions{
		AdditionallyAllowedTenants: ac.AdditionallyAllowedTenants,
		ClientOptions:              options.ClientOptions,
		DisableInstanceDiscovery:   ac.DisableInstanceDiscovery,
	}

	return azidentity.NewUsernamePasswordCredential(tenantID, clientID, username, password, o)
}

func (ac AuthCredentials) toCertificateCredential(tenantID, clientID string, options *azblob.ClientOptions) (azcore.TokenCredential, error) {
	if clientID == "" || (ac.ClientCertificate == nil || ac.ClientCertificatePath == nil) {
		return nil, nil
	}

	var certData []byte

	if ac.ClientCertificate != nil {
		inlineCert, err := ac.ClientCertificate.GetOrDefault("")
		if err != nil {
			return nil, err
		}

		if inlineCert != "" {
			b64, err := base64.StdEncoding.DecodeString(inlineCert)
			if err != nil {
				return nil, fmt.Errorf("failed to decode client certificate from base64 string: %w", err)
			}

			certData = b64
		}
	}

	if len(certData) == 0 && ac.ClientCertificatePath != nil {
		certPath, err := ac.ClientCertificatePath.GetOrDefault("")
		if err != nil {
			return nil, err
		}

		certData, err = os.ReadFile(certPath)
		if err != nil {
			return nil, fmt.Errorf(`failed to read certificate file "%s": %w`, certPath, err)
		}
	}

	if len(certData) == 0 {
		return nil, nil
	}

	var password []byte

	if ac.ClientCertificatePassword != nil {
		passwordStr, err := ac.ClientCertificatePassword.Get()
		if err != nil {
			return nil, err
		}

		password = []byte(passwordStr)
	}

	certs, key, err := azidentity.ParseCertificates(certData, password)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate due to error %w. This may be due to a limitation of this module's certificate loader. Consider calling NewClientCertificateCredential instead", err)
	}

	o := &azidentity.ClientCertificateCredentialOptions{
		AdditionallyAllowedTenants: ac.AdditionallyAllowedTenants,
		ClientOptions:              options.ClientOptions,
		DisableInstanceDiscovery:   ac.DisableInstanceDiscovery,
		SendCertificateChain:       ac.SendCertificateChain,
	}

	return azidentity.NewClientCertificateCredential(tenantID, clientID, certs, key, o)
}
