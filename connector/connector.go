package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hasura/ndc-loki/connector/api"
	"github.com/hasura/ndc-loki/connector/client"
	"github.com/hasura/ndc-loki/connector/metadata"
	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
)

// LokiConnector implements a data connector for Loki API
type LokiConnector struct {
	capabilities *schema.RawCapabilitiesResponse
	rawSchema    *schema.RawSchemaResponse
	metadata     *metadata.Metadata
	runtime      *metadata.RuntimeSettings
	apiHandler   api.DataConnectorHandler
}

// NewLokiConnector creates a Loki connector instance
func NewLokiConnector() *LokiConnector {
	return &LokiConnector{
		apiHandler: api.DataConnectorHandler{},
	}
}

// ParseConfiguration validates the configuration files provided by the user, returning a validated 'Configuration',
// or throwing an error to prevents Connector startup.
func (c *LokiConnector) ParseConfiguration(ctx context.Context, configurationDir string) (*metadata.Configuration, error) {
	restCapabilities := schema.CapabilitiesResponse{
		Version: "0.1.6",
		Capabilities: schema.Capabilities{
			Query: schema.QueryCapabilities{
				Variables:    schema.LeafCapability{},
				NestedFields: schema.NestedFieldCapabilities{},
				Explain:      map[string]any{},
			},
			Mutation: schema.MutationCapabilities{},
		},
	}
	rawCapabilities, err := json.Marshal(restCapabilities)
	if err != nil {
		return nil, fmt.Errorf("failed to encode capabilities: %w", err)
	}
	c.capabilities = schema.NewRawCapabilitiesResponseUnsafe(rawCapabilities)
	config, err := metadata.ReadConfiguration(configurationDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read the configuration file at %s: %w", configurationDir, err)
	}

	c.metadata = &config.Metadata
	c.runtime = &config.Runtime

	return config, nil
}

// TryInitState initializes the connector's in-memory state.
//
// For example, any connection pools, prepared queries,
// or other managed resources would be allocated here.
//
// In addition, this function should register any
// connector-specific metrics with the metrics registry.
func (c *LokiConnector) TryInitState(ctx context.Context, conf *metadata.Configuration, metrics *connector.TelemetryState) (*metadata.State, error) {
	_, span := metrics.Tracer.StartInternal(ctx, "Initialize")
	defer span.End()

	querySchema, err := metadata.BuildConnectorSchema(&conf.Metadata)
	if err != nil {
		return nil, err
	}
	ndcSchema, errs := utils.MergeSchemas(api.GetConnectorSchema(), querySchema)
	for _, e := range errs {
		slog.Debug(e.Error())
	}

	rawSchema, err := json.Marshal(ndcSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to encode schema to json: %w", err)
	}
	c.rawSchema = schema.NewRawSchemaResponseUnsafe(rawSchema)

	client, err := client.New(conf.ConnectionSettings)
	if err != nil {
		return nil, err
	}

	return &metadata.State{
		Client: client,
		Tracer: metrics.Tracer,
	}, nil
}

// GetSchema gets the connector's schema.
func (c *LokiConnector) GetSchema(ctx context.Context, configuration *metadata.Configuration, _ *metadata.State) (schema.SchemaResponseMarshaler, error) {
	return c.rawSchema, nil
}

// HealthCheck checks the health of the connector.
//
// For example, this function should check that the connector
// is able to reach its data source over the network.
//
// Should throw if the check fails, else resolve.
func (c *LokiConnector) HealthCheck(ctx context.Context, conf *metadata.Configuration, state *metadata.State) error {
	// return state.Client.Healthy(ctx)
	return nil
}

// GetCapabilities get the connector's capabilities.
func (c *LokiConnector) GetCapabilities(conf *metadata.Configuration) schema.CapabilitiesResponseMarshaler {
	return c.capabilities
}
