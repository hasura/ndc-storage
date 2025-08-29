package connector

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/hasura/ndc-sdk-go/v2/connector"
	"github.com/hasura/ndc-sdk-go/v2/ndctest"
)

func TestConnector(t *testing.T) {
	setConnectorTestEnv(t)

	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		configDir = "../tests/configuration"
	}

	for i, dir := range []string{"01-setup", "02-get", "03-cleanup"} {
		var serverOptions []connector.ServeOption

		if i == 0 {
			logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
				// Level: slog.LevelDebug,
			}))
			serverOptions = append(serverOptions, connector.WithLogger(logger))
		}

		ndctest.TestConnector(t, &Connector{}, ndctest.TestConnectorOptions{
			Configuration: "../tests/configuration",
			TestDataDir:   filepath.Join("testdata", "static", dir),
			ServerOptions: serverOptions,
		})
	}
}

func TestConnectorDynamicCredentials(t *testing.T) {
	configDir := os.Getenv("CONFIG_DIR")
	if configDir != "../tests/configuration" {
		return
	}

	setConnectorTestEnv(t)

	for i, dir := range []string{"01-setup", "02-get", "03-cleanup"} {
		var serverOptions []connector.ServeOption

		if i == 0 {
			logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
				// Level: slog.LevelDebug,
			}))
			serverOptions = append(serverOptions, connector.WithLogger(logger))
		}

		ndctest.TestConnector(t, &Connector{}, ndctest.TestConnectorOptions{
			Configuration: "../tests/configuration",
			TestDataDir:   filepath.Join("testdata", "dynamic", dir),
			ServerOptions: serverOptions,
		})
	}
}

func setConnectorTestEnv(t *testing.T) {
	azureBlobEndpoint := "https://local.hasura.dev:10000"
	azureAccountName := "local"
	azureAccountKey := "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="

	t.Setenv("STORAGE_ENDPOINT", "http://localhost:9000")
	t.Setenv("DEFAULT_BUCKET", "default")
	t.Setenv("ACCESS_KEY_ID", "test-key")
	t.Setenv("SECRET_ACCESS_KEY", "randomsecret")
	t.Setenv("S3_STORAGE_ENDPOINT", "http://localhost:9010")
	t.Setenv("S3_DEFAULT_BUCKET", "bucket1")
	t.Setenv("S3_ACCESS_KEY_ID", "test-key")
	t.Setenv("S3_SECRET_ACCESS_KEY", "randomsecret")
	t.Setenv("AZURE_STORAGE_ENDPOINT", azureBlobEndpoint)
	t.Setenv("AZURE_STORAGE_DEFAULT_BUCKET", "azure-test")
	t.Setenv("AZURE_STORAGE_ACCOUNT_NAME", azureAccountName)
	t.Setenv("AZURE_STORAGE_ACCOUNT_KEY", azureAccountKey)
	t.Setenv("AZURE_STORAGE_CONNECTION_STRING", fmt.Sprintf("DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;BlobEndpoint=%s", azureAccountName, azureAccountKey, azureBlobEndpoint))
	t.Setenv("GOOGLE_STORAGE_DEFAULT_BUCKET", "gcp-bucket")
	t.Setenv("GOOGLE_PROJECT_ID", "test-local-project")
	t.Setenv("GOOGLE_STORAGE_ENDPOINT", "https://local.hasura.dev:4443/storage/v1/")
	t.Setenv("GOOGLE_STORAGE_CREDENTIALS_FILE", "../tests/certs/service_account.json")
	t.Setenv("STORAGE_FS_DEFAULT_DIRECTORY", "../tmp/data")
	t.Setenv("TLS_CERT_FILE", "../tests/certs/tls/client.crt")
	t.Setenv("TLS_KEY_FILE", "../tests/certs/tls/client.key")
	t.Setenv("TLS_CA_FILE", "../tests/certs/tls/ca.crt")
}
