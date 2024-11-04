package metadata

import (
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
)

func createQueryResultValuesObjectFields(queryType QueryType) schema.ObjectTypeFields {
	if queryType == QueryTypeMetric {
		return createMetricObjectFields()
	}

	return createStreamObjectFields()
}

func createStreamValueObjectFields() schema.ObjectTypeFields {
	return schema.ObjectTypeFields{
		TimestampKey: schema.ObjectField{
			Description: utils.ToPtr("Timestamp of the log line"),
			Type:        schema.NewNamedType(string(ScalarTimestamp)).Encode(),
		},
		ValueKey: schema.ObjectField{
			Description: utils.ToPtr("The log line content"),
			Type:        schema.NewNamedType(string(ScalarLogLine)).Encode(),
		},
	}
}

func createStreamObjectFields() schema.ObjectTypeFields {
	return schema.ObjectTypeFields{
		OriginalLabelsKey: schema.ObjectField{
			Description: utils.ToPtr("Labels of the metric"),
			Type:        schema.NewNamedType(string(ScalarLabelSet)).Encode(),
		},
		TimestampKey: schema.ObjectField{
			Description: utils.ToPtr("The timestamp of the current log line or the last timestamp of a range query result"),
			Type:        schema.NewNamedType(string(ScalarTimestamp)).Encode(),
		},
		LogLineKey: schema.ObjectField{
			Description: utils.ToPtr("A log line if the flat values setting is enabled"),
			Type:        schema.NewNamedType(string(ScalarLogLine)).Encode(),
		},
		LogLinesKey: schema.ObjectField{
			Description: utils.ToPtr("List of log lines grouped by unique labels"),
			Type:        schema.NewArrayType(schema.NewNamedType(objectName_StreamValue)).Encode(),
		},
	}
}

func createMetricValueObjectFields() schema.ObjectTypeFields {
	return schema.ObjectTypeFields{
		TimestampKey: schema.ObjectField{
			Description: utils.ToPtr("The timestamp when the value is calculated"),
			Type:        schema.NewNamedType(string(ScalarTimestamp)).Encode(),
		},
		ValueKey: schema.ObjectField{
			Description: utils.ToPtr("The metric value"),
			Type:        schema.NewNamedType(string(ScalarDecimal)).Encode(),
		},
	}
}

func createMetricObjectFields() schema.ObjectTypeFields {
	return schema.ObjectTypeFields{
		OriginalLabelsKey: schema.ObjectField{
			Description: utils.ToPtr("Labels of the metric"),
			Type:        schema.NewNamedType(string(ScalarLabelSet)).Encode(),
		},
		TimestampKey: schema.ObjectField{
			Description: utils.ToPtr("An instant timestamp or the last timestamp of a range query result"),
			Type:        schema.NewNamedType(string(ScalarTimestamp)).Encode(),
		},
		LogLineKey: schema.ObjectField{
			Description: utils.ToPtr("A log line if the flat values setting is enabled"),
			Type:        schema.NewNamedType(string(ScalarLogLine)).Encode(),
		},
		MetricValueKey: schema.ObjectField{
			Description: utils.ToPtr("Value of the instant query or the last value of a range query"),
			Type:        schema.NewNamedType(string(ScalarDecimal)).Encode(),
		},
		MetricValuesKey: schema.ObjectField{
			Description: utils.ToPtr("List of metric values grouped by unique labels"),
			Type:        schema.NewArrayType(schema.NewNamedType(objectName_MetricValue)).Encode(),
		},
	}
}

func createCollectionArguments(queryType QueryType) schema.CollectionInfoArguments {
	if queryType == QueryTypeMetric {
		return createMetricArguments()
	}

	return createStreamArguments()
}

func createStreamArguments() schema.CollectionInfoArguments {
	arguments := schema.CollectionInfoArguments{}
	for _, key := range []string{ArgumentKeyInterval, ArgumentKeyFlat} {
		arguments[key] = defaultArgumentInfos[key]
	}

	return arguments
}

func createMetricArguments() schema.CollectionInfoArguments {
	arguments := schema.CollectionInfoArguments{}
	for _, key := range []string{ArgumentKeyStep, ArgumentKeyOffset, ArgumentKeyFlat} {
		arguments[key] = defaultArgumentInfos[key]
	}

	return arguments
}
