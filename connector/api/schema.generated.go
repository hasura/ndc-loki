// Code generated by github.com/hasura/ndc-sdk-go/cmd/hasura-ndc-go, DO NOT EDIT.
package api

import (
	"github.com/hasura/ndc-sdk-go/schema"
)

func toPtr[V any](value V) *V {
	return &value
}

// GetConnectorSchema gets the generated connector schema
func GetConnectorSchema() *schema.SchemaResponse {
	return &schema.SchemaResponse{
		Collections: []schema.CollectionInfo{},
		ObjectTypes: schema.SchemaResponseObjectTypes{
			"LabelsParams": schema.ObjectType{
				Description: toPtr("represent parameters of the labels request"),
				Fields: schema.ObjectTypeFields{
					"end": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("TimestampTZ")).Encode(),
					},
					"query": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					"since": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("Duration")).Encode(),
					},
					"start": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("TimestampTZ")).Encode(),
					},
				},
			},
			"LogDeletionRequest": schema.ObjectType{
				Description: toPtr("the log deletion request item"),
				Fields: schema.ObjectTypeFields{
					"end_time": schema.ObjectField{
						Type: schema.NewNamedType("Int64").Encode(),
					},
					"query": schema.ObjectField{
						Type: schema.NewNamedType("String").Encode(),
					},
					"start_time": schema.ObjectField{
						Type: schema.NewNamedType("Int64").Encode(),
					},
					"status": schema.ObjectField{
						Type: schema.NewNamedType("String").Encode(),
					},
				},
			},
			"LogLineInput": schema.ObjectType{
				Description: toPtr("represents a log line item"),
				Fields: schema.ObjectTypeFields{
					"line": schema.ObjectField{
						Type: schema.NewNamedType("String").Encode(),
					},
					"structured_metadata": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("JSON")).Encode(),
					},
					"timestamp": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("TimestampTZ")).Encode(),
					},
				},
			},
			"MatrixValue": schema.ObjectType{
				Description: toPtr("holds a single timestamp and a metric value"),
				Fields: schema.ObjectTypeFields{
					"time": schema.ObjectField{
						Type: schema.NewNamedType("TimestampTZ").Encode(),
					},
					"value": schema.ObjectField{
						Type: schema.NewNamedType("Float64").Encode(),
					},
				},
			},
			"MatrixValues": schema.ObjectType{
				Description: toPtr("holds a label key value pairs for the metric and a list of values"),
				Fields: schema.ObjectTypeFields{
					"metric": schema.ObjectField{
						Type: schema.NewNamedType("JSON").Encode(),
					},
					"values": schema.ObjectField{
						Type: schema.NewArrayType(schema.NewNamedType("MatrixValue")).Encode(),
					},
				},
			},
			"QueryData": schema.ObjectType{
				Description: toPtr("holds the result type and a metric vector value of the instant query response"),
				Fields: schema.ObjectTypeFields{
					"encodingFlags": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewArrayType(schema.NewNamedType("String"))).Encode(),
					},
					"resultType": schema.ObjectField{
						Type: schema.NewNamedType("String").Encode(),
					},
					"vector": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewArrayType(schema.NewNamedType("VectorValue"))).Encode(),
					},
				},
			},
			"QueryRangeData": schema.ObjectType{
				Description: toPtr("holds the result type and a list of stream or metric values of the query range response"),
				Fields: schema.ObjectTypeFields{
					"encodingFlags": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewArrayType(schema.NewNamedType("String"))).Encode(),
					},
					"matrix": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewArrayType(schema.NewNamedType("MatrixValues"))).Encode(),
					},
					"resultType": schema.ObjectField{
						Type: schema.NewNamedType("String").Encode(),
					},
					"stream": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewArrayType(schema.NewNamedType("StreamValues"))).Encode(),
					},
				},
			},
			"StreamInput": schema.ObjectType{
				Description: toPtr("represents a stream input object."),
				Fields: schema.ObjectTypeFields{
					"stream": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("JSON")).Encode(),
					},
					"values": schema.ObjectField{
						Type: schema.NewArrayType(schema.NewNamedType("LogLineInput")).Encode(),
					},
				},
			},
			"StreamValue": schema.ObjectType{
				Description: toPtr("holds a timestamp and the content of a log line."),
				Fields: schema.ObjectTypeFields{
					"time": schema.ObjectField{
						Type: schema.NewNamedType("TimestampTZ").Encode(),
					},
					"value": schema.ObjectField{
						Type: schema.NewNamedType("String").Encode(),
					},
				},
			},
			"StreamValues": schema.ObjectType{
				Description: toPtr("holds a label key value pairs for the log stream and a list of values"),
				Fields: schema.ObjectTypeFields{
					"stream": schema.ObjectField{
						Type: schema.NewNamedType("JSON").Encode(),
					},
					"values": schema.ObjectField{
						Type: schema.NewArrayType(schema.NewNamedType("StreamValue")).Encode(),
					},
				},
			},
			"VectorValue": schema.ObjectType{
				Description: toPtr("holds a label key value pairs for the metric and single timestamp and value"),
				Fields: schema.ObjectTypeFields{
					"metric": schema.ObjectField{
						Type: schema.NewNamedType("JSON").Encode(),
					},
					"time": schema.ObjectField{
						Type: schema.NewNamedType("TimestampTZ").Encode(),
					},
					"value": schema.ObjectField{
						Type: schema.NewNamedType("Float64").Encode(),
					},
				},
			},
		},
		Functions: []schema.FunctionInfo{
			{
				Name:        "loki_label_values",
				Description: toPtr("retrieve the list of known values for a given label within a given time span."),
				ResultType:  schema.NewArrayType(schema.NewNamedType("String")).Encode(),
				Arguments: map[string]schema.ArgumentInfo{
					"end": {
						Type: schema.NewNullableType(schema.NewNamedType("TimestampTZ")).Encode(),
					},
					"name": {
						Type: schema.NewNamedType("String").Encode(),
					},
					"query": {
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					"since": {
						Type: schema.NewNullableType(schema.NewNamedType("Duration")).Encode(),
					},
					"start": {
						Type: schema.NewNullableType(schema.NewNamedType("TimestampTZ")).Encode(),
					},
				},
			},
			{
				Name:        "loki_labels",
				Description: toPtr("return the list of known labels within a given time span."),
				ResultType:  schema.NewArrayType(schema.NewNamedType("String")).Encode(),
				Arguments: map[string]schema.ArgumentInfo{
					"end": {
						Type: schema.NewNullableType(schema.NewNamedType("TimestampTZ")).Encode(),
					},
					"query": {
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					"since": {
						Type: schema.NewNullableType(schema.NewNamedType("Duration")).Encode(),
					},
					"start": {
						Type: schema.NewNullableType(schema.NewNamedType("TimestampTZ")).Encode(),
					},
				},
			},
			{
				Name:        "loki_log_deletion_requests",
				Description: toPtr("list the existing delete requests for the authenticated tenant"),
				ResultType:  schema.NewArrayType(schema.NewNamedType("LogDeletionRequest")).Encode(),
				Arguments:   map[string]schema.ArgumentInfo{},
			},
			{
				Name:        "loki_query",
				Description: toPtr("allows for doing [queries against a single point in time]"),
				ResultType:  schema.NewNamedType("QueryData").Encode(),
				Arguments: map[string]schema.ArgumentInfo{
					"direction": {
						Type: schema.NewNullableType(schema.NewNamedType("QueryDirection")).Encode(),
					},
					"limit": {
						Type: schema.NewNullableType(schema.NewNamedType("Int32")).Encode(),
					},
					"query": {
						Type: schema.NewNamedType("String").Encode(),
					},
					"time": {
						Type: schema.NewNullableType(schema.NewNamedType("TimestampTZ")).Encode(),
					},
				},
			},
			{
				Name:       "loki_query_range",
				ResultType: schema.NewNamedType("QueryRangeData").Encode(),
				Arguments: map[string]schema.ArgumentInfo{
					"direction": {
						Type: schema.NewNullableType(schema.NewNamedType("QueryDirection")).Encode(),
					},
					"end": {
						Type: schema.NewNullableType(schema.NewNamedType("TimestampTZ")).Encode(),
					},
					"interval": {
						Type: schema.NewNullableType(schema.NewNamedType("Duration")).Encode(),
					},
					"limit": {
						Type: schema.NewNullableType(schema.NewNamedType("Int32")).Encode(),
					},
					"query": {
						Type: schema.NewNamedType("String").Encode(),
					},
					"since": {
						Type: schema.NewNullableType(schema.NewNamedType("Duration")).Encode(),
					},
					"start": {
						Type: schema.NewNullableType(schema.NewNamedType("TimestampTZ")).Encode(),
					},
					"step": {
						Type: schema.NewNullableType(schema.NewNamedType("Duration")).Encode(),
					},
				},
			},
		},
		Procedures: []schema.ProcedureInfo{
			{
				Name:        "loki_cancel_log_deletion_request",
				Description: toPtr("cancels a new log deletion request for the authenticated tenant"),
				ResultType:  schema.NewNamedType("Boolean").Encode(),
				Arguments: map[string]schema.ArgumentInfo{
					"force": {
						Type: schema.NewNullableType(schema.NewNamedType("Boolean")).Encode(),
					},
					"request_id": {
						Type: schema.NewNamedType("String").Encode(),
					},
				},
			},
			{
				Name:        "loki_create_log_deletion_request",
				Description: toPtr("creates a new log deletion request for the authenticated tenant"),
				ResultType:  schema.NewNamedType("Boolean").Encode(),
				Arguments: map[string]schema.ArgumentInfo{
					"end": {
						Type: schema.NewNullableType(schema.NewNamedType("TimestampTZ")).Encode(),
					},
					"query": {
						Type: schema.NewNamedType("String").Encode(),
					},
					"start": {
						Type: schema.NewNullableType(schema.NewNamedType("TimestampTZ")).Encode(),
					},
				},
			},
			{
				Name:        "loki_push_log_lines",
				Description: toPtr("pushes log lines to Loki"),
				ResultType:  schema.NewNamedType("Boolean").Encode(),
				Arguments: map[string]schema.ArgumentInfo{
					"streams": {
						Type: schema.NewArrayType(schema.NewNamedType("StreamInput")).Encode(),
					},
				},
			},
		},
		ScalarTypes: schema.SchemaResponseScalarTypes{
			"Boolean": schema.ScalarType{
				AggregateFunctions:  schema.ScalarTypeAggregateFunctions{},
				ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{},
				Representation:      schema.NewTypeRepresentationBoolean().Encode(),
			},
			"Duration": schema.ScalarType{
				AggregateFunctions:  schema.ScalarTypeAggregateFunctions{},
				ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{},
				Representation:      schema.NewTypeRepresentationJSON().Encode(),
			},
			"Float64": schema.ScalarType{
				AggregateFunctions:  schema.ScalarTypeAggregateFunctions{},
				ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{},
				Representation:      schema.NewTypeRepresentationFloat64().Encode(),
			},
			"Int32": schema.ScalarType{
				AggregateFunctions:  schema.ScalarTypeAggregateFunctions{},
				ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{},
				Representation:      schema.NewTypeRepresentationInt32().Encode(),
			},
			"Int64": schema.ScalarType{
				AggregateFunctions:  schema.ScalarTypeAggregateFunctions{},
				ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{},
				Representation:      schema.NewTypeRepresentationInt64().Encode(),
			},
			"JSON": schema.ScalarType{
				AggregateFunctions:  schema.ScalarTypeAggregateFunctions{},
				ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{},
				Representation:      schema.NewTypeRepresentationJSON().Encode(),
			},
			"QueryDirection": schema.ScalarType{
				AggregateFunctions:  schema.ScalarTypeAggregateFunctions{},
				ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{},
				Representation:      schema.NewTypeRepresentationEnum([]string{"forward", "backward"}).Encode(),
			},
			"String": schema.ScalarType{
				AggregateFunctions:  schema.ScalarTypeAggregateFunctions{},
				ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{},
				Representation:      schema.NewTypeRepresentationString().Encode(),
			},
			"TimestampTZ": schema.ScalarType{
				AggregateFunctions:  schema.ScalarTypeAggregateFunctions{},
				ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{},
				Representation:      schema.NewTypeRepresentationTimestampTZ().Encode(),
			},
		},
	}
}
