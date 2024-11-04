package metadata

import (
	"fmt"

	"github.com/hasura/ndc-sdk-go/schema"
)

type connectorSchemaBuilder struct {
	Metadata    *Metadata
	ScalarTypes schema.SchemaResponseScalarTypes
	ObjectTypes schema.SchemaResponseObjectTypes
	Collections map[string]schema.CollectionInfo
	Functions   map[string]schema.FunctionInfo
}

// BuildConnectorSchema builds the schema for the data connector from metadata
func BuildConnectorSchema(metadata *Metadata) (*schema.SchemaResponse, error) {
	builder := &connectorSchemaBuilder{
		Metadata:    metadata,
		ScalarTypes: defaultScalars,
		ObjectTypes: defaultObjectTypes,
		Functions:   map[string]schema.FunctionInfo{},
		Collections: map[string]schema.CollectionInfo{},
	}

	if err := builder.buildModels(); err != nil {
		return nil, err
	}
	if err := builder.buildNativeQueries(); err != nil {
		return nil, err
	}

	return builder.buildSchemaResponse(), nil
}

func (scb *connectorSchemaBuilder) buildSchemaResponse() *schema.SchemaResponse {
	functions := make([]schema.FunctionInfo, 0, len(scb.Functions))
	collections := make([]schema.CollectionInfo, 0, len(scb.Collections))
	for _, fn := range scb.Functions {
		functions = append(functions, fn)
	}

	for _, collection := range scb.Collections {
		collections = append(collections, collection)
	}

	return &schema.SchemaResponse{
		Collections: collections,
		ObjectTypes: scb.ObjectTypes,
		Procedures:  []schema.ProcedureInfo{},
		ScalarTypes: scb.ScalarTypes,
		Functions:   functions,
	}
}

func (scb *connectorSchemaBuilder) checkDuplicatedOperation(name string) error {
	err := fmt.Errorf("duplicated operation name: %s", name)
	if _, ok := scb.Functions[name]; ok {
		return err
	}
	if _, ok := scb.Collections[name]; ok {
		return err
	}

	return nil
}
