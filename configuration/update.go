package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/hasura/ndc-storage/connector/storage/minio"
	"github.com/hasura/ndc-storage/connector/types"
	"gopkg.in/yaml.v3"
)

// UpdateArguments represent input arguments of the `update` command.
type UpdateArguments struct {
	Dir string `default:"." env:"HASURA_PLUGIN_CONNECTOR_CONTEXT_PATH" help:"The directory where the configuration.yaml file is present" short:"d"`
}

// UpdateConfig validate and update the configuration.
func UpdateConfig(dir string) error {
	configPath := filepath.Join(dir, types.ConfigurationFileName)

	rawBytes, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return writeConfig(configPath, &defaultConfiguration)
		}

		return err
	}

	var config types.Configuration
	if err := yaml.Unmarshal(rawBytes, &config); err != nil {
		return err
	}

	return config.Validate()
}

func writeConfig(filePath string, config *types.Configuration) error {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	_, _ = writer.WriteString("# yaml-language-server: $schema=https://raw.githubusercontent.com/hasura/ndc-storage/main/jsonschema/configuration.json\n")
	encoder := yaml.NewEncoder(writer)
	encoder.SetIndent(2)

	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode the configuration file: %w", err)
	}

	writer.Flush()

	return os.WriteFile(filePath, buf.Bytes(), 0o644)
}

var defaultConfiguration = types.Configuration{
	Concurrency: types.ConcurrencySettings{
		Query:    5,
		Mutation: 1,
	},
	Runtime: storage.RuntimeSettings{
		MaxDownloadSizeMBs: 20,
		MaxUploadSizeMBs:   20,
	},
	Clients: []storage.ClientConfig{
		{
			"type":          common.S3,
			"endpoint":      utils.ToPtr(utils.NewEnvStringVariable("STORAGE_ENDPOINT")),
			"defaultBucket": utils.NewEnvStringVariable("DEFAULT_BUCKET"),
			"authentication": minio.AuthCredentials{
				Type:            minio.AuthTypeStatic,
				AccessKeyID:     utils.ToPtr(utils.NewEnvStringVariable("ACCESS_KEY_ID")),
				SecretAccessKey: utils.ToPtr(utils.NewEnvStringVariable("SECRET_ACCESS_KEY")),
			},
		},
	},
}
