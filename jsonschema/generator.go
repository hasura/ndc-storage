package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hasura/ndc-storage/connector/types"
	"github.com/invopop/jsonschema"
)

func main() {
	err := jsonSchemaConfiguration()
	if err != nil {
		panic(fmt.Errorf("failed to write jsonschema for Configuration: %w", err))
	}
}

func jsonSchemaConfiguration() error {
	r := new(jsonschema.Reflector)
	if err := r.AddGoComments("github.com/hasura/ndc-storage/connector/types", "../connector/types", jsonschema.WithFullComment()); err != nil {
		return err
	}

	if err := r.AddGoComments("github.com/hasura/ndc-storage/connector/storage", "../connector/storage", jsonschema.WithFullComment()); err != nil {
		return err
	}

	if err := r.AddGoComments("github.com/hasura/ndc-storage/connector/storage/common", "../connector/storage/common", jsonschema.WithFullComment()); err != nil {
		return err
	}

	reflectSchema := r.Reflect(types.Configuration{})
	envString := &jsonschema.Schema{
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
	envString.Properties.Set("env", &jsonschema.Schema{
		Type: "string",
	})
	envString.Properties.Set("value", &jsonschema.Schema{
		Type: "string",
	})

	reflectSchema.Definitions["EnvString"] = envString

	schemaBytes, err := json.MarshalIndent(reflectSchema, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("configuration.schema.json", schemaBytes, 0o644)
}
