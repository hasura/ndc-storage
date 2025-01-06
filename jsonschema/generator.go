package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hasura/ndc-storage/connector/types"
	"github.com/invopop/jsonschema"
)

func main() {
	if err := jsonSchemaConfiguration(); err != nil {
		panic(fmt.Errorf("failed to write jsonschema for Configuration: %w", err))
	}
}

func jsonSchemaConfiguration() error {
	r := new(jsonschema.Reflector)
	if err := r.AddGoComments("github.com/hasura/ndc-storage/connector/types", "../connector/types"); err != nil {
		return err
	}

	if err := r.AddGoComments("github.com/hasura/ndc-storage/connector/storage", "../connector/storage"); err != nil {
		return err
	}

	if err := r.AddGoComments("github.com/hasura/ndc-storage/connector/storage/common", "../connector/storage/common"); err != nil {
		return err
	}

	reflectSchema := r.Reflect(&types.Configuration{})

	schemaBytes, err := json.MarshalIndent(reflectSchema, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("configuration.schema.json", schemaBytes, 0o644)
}
