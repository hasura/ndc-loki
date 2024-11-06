package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hasura/ndc-loki/connector/metadata"
	"github.com/invopop/jsonschema"
)

func main() {
	if err := jsonSchemaConfiguration(); err != nil {
		panic(fmt.Errorf("failed to write jsonschema for configuration: %w", err))
	}
}

func jsonSchemaConfiguration() error {
	r := new(jsonschema.Reflector)
	if err := r.AddGoComments("github.com/hasura/ndc-loki/connector/client", "../connector/client"); err != nil {
		return err
	}
	if err := r.AddGoComments("github.com/hasura/ndc-loki/connector/metadata", "../connector/metadata"); err != nil {
		return err
	}

	reflectSchema := r.Reflect(&metadata.Configuration{})
	modelFieldSchema := r.Reflect(&metadata.ModelField{})
	for key, def := range modelFieldSchema.Definitions {
		reflectSchema.Definitions[key] = def
	}
	labelFormatRuleSchema := r.Reflect(&metadata.LabelFormatRule{})
	for key, def := range labelFormatRuleSchema.Definitions {
		reflectSchema.Definitions[key] = def
	}

	schemaBytes, err := json.MarshalIndent(reflectSchema, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("configuration.schema.json", schemaBytes, 0644)
}
