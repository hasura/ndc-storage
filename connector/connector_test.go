package connector

import (
	"path/filepath"
	"testing"

	"github.com/hasura/ndc-sdk-go/ndctest"
)

func TestConnector(t *testing.T) {
	t.Setenv("STORAGE_ENDPOINT", "http://localhost:9000")
	t.Setenv("DEFAULT_BUCKET", "default")
	t.Setenv("ACCESS_KEY_ID", "test-key")
	t.Setenv("SECRET_ACCESS_KEY", "randomsecret")
	t.Setenv("S3_STORAGE_ENDPOINT", "http://localhost:9010")
	t.Setenv("S3_DEFAULT_BUCKET", "bucket1")
	t.Setenv("S3_ACCESS_KEY_ID", "test-key")
	t.Setenv("S3_SECRET_ACCESS_KEY", "randomsecret")
	t.Setenv("AZURE_STORAGE_ENDPOINT", "http://local.hasura.dev:10000")
	t.Setenv("AZURE_STORAGE_DEFAULT_BUCKET", "azure-test")
	t.Setenv("AZURE_STORAGE_ACCOUNT_NAME", "local")
	t.Setenv("AZURE_STORAGE_ACCOUNT_KEY", "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==")

	for _, dir := range []string{"01-setup", "02-get", "03-cleanup"} {
		ndctest.TestConnector(t, &Connector{}, ndctest.TestConnectorOptions{
			Configuration: "../tests/configuration",
			TestDataDir:   filepath.Join("testdata", dir),
		})
	}
}
