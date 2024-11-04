package metadata

import (
	"os"
	"strings"

	"github.com/hasura/ndc-loki/connector/client"
	"gopkg.in/yaml.v3"
)

// Configuration the configuration of the Loki connector
type Configuration struct {
	// Connection settings to connect the Loki server
	ConnectionSettings client.ClientSettings `json:"connection_settings" yaml:"connection_settings"`
	// The metadata of log stream and metric models
	Metadata Metadata `json:"metadata" yaml:"metadata"`
	// Runtime settings
	Runtime RuntimeSettings `json:"runtime" yaml:"runtime"`
}

// Metadata the metadata configuration
type Metadata struct {
	// Definition of stream models
	Models map[string]ModelInfo `json:"models" yaml:"models"`
	// The collection of native operations
	NativeOperations NativeOperations `json:"native_operations" yaml:"native_operations"`
}

// GetModel gets a model by name
func (m *Metadata) GetModel(name string) (*ModelInfo, QueryType) {
	collection, ok := m.Models[name]
	if ok {
		return &collection, QueryTypeStream
	}

	if strings.HasSuffix(name, "_aggregate") {
		modelName := strings.TrimSuffix(name, "_aggregate")
		collection, ok = m.Models[modelName]
		if ok {
			return &collection, QueryTypeMetric
		}
	}

	return nil, ""
}

// RuntimeFormatSettings format settings for timestamps and values in runtime
type RuntimeFormatSettings struct {
	// The serialization format for timestamp
	Timestamp TimestampFormat `json:"timestamp" yaml:"timestamp" jsonschema:"enum=rfc3339,enum=unix,default=unix"`
	// The serialization format for value
	Value ValueFormat `json:"value" yaml:"value" jsonschema:"enum=string,enum=float64,default=string"`
	// The serialization format for not-a-number values
	NaN any `json:"nan" yaml:"nan" jsonschema:"oneof_type=string;number;null"`
	// The serialization format for infinite values
	Inf any `json:"inf" yaml:"inf" jsonschema:"oneof_type=string;number;null"`
	// The serialization format for negative infinite values
	NegativeInf any `json:"negative_inf" yaml:"negative_inf" jsonschema:"oneof_type=string;number;null"`
}

// RuntimeSettings contain settings for the runtime engine
type RuntimeSettings struct {
	// Flatten value points to the root array
	Flat bool `json:"flat" yaml:"flat"`
	// The default unit for unix timestamp
	UnixTimeUnit UnixTimeUnit `json:"unix_time_unit" yaml:"unix_time_unit" jsonschema:"enum=s,enum=ms,enum=us,enum=ns,default=s"`
	// The serialization format for response fields
	Format RuntimeFormatSettings `json:"format" yaml:"format"`
	// The concurrency limit of queries if there are many variables in a single query
	QueryConcurrencyLimit int `json:"query_concurrency_limit,omitempty" yaml:"query_concurrency_limit,omitempty"`
	// The concurrency limit of operations if there are many operations in a single mutation
	MutationConcurrencyLimit int `json:"mutation_concurrency_limit,omitempty" yaml:"mutation_concurrency_limit,omitempty"`
}

// ReadConfiguration reads the configuration from file
func ReadConfiguration(configurationDir string) (*Configuration, error) {
	var config Configuration
	configFilePath := configurationDir + "/configuration.yaml"
	yamlBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(yamlBytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
