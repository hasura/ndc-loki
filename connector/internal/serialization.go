package internal

import (
	"github.com/hasura/ndc-loki/connector/client"
	"github.com/hasura/ndc-loki/connector/metadata"
)

// serializeStreamValues serialize log stream values to maps
func serializeStreamValues(data []client.StreamValues, labels map[string]metadata.LabelInfo, fields map[string]metadata.ModelFieldValue, runtimeSettings *metadata.RuntimeSettings, flat bool) []map[string]any {
	if flat {
		return serializeFlatStreamValues(data, labels, fields, runtimeSettings)
	}

	return serializeGroupedStreamValues(data, labels, fields, runtimeSettings)
}

func serializeGroupedStreamValues(data []client.StreamValues, labels map[string]metadata.LabelInfo, fields map[string]metadata.ModelFieldValue, runtimeSettings *metadata.RuntimeSettings) []map[string]any {
	results := make([]map[string]any, len(data))
	for i, item := range data {
		valueLen := len(item.Values)
		if valueLen == 0 {
			continue
		}
		lastValue := item.Values[valueLen-1]
		result := extractLabelAndFieldValues(item.Stream, labels, fields)
		result[metadata.TimestampKey] = formatTimestamp(lastValue.Time, runtimeSettings.Format.Timestamp)
		result[metadata.LogLineKey] = lastValue.Value

		values := make([]map[string]any, len(item.Values))
		for j, rawValue := range item.Values {
			values[j] = map[string]any{
				metadata.TimestampKey: formatTimestamp(lastValue.Time, runtimeSettings.Format.Timestamp),
				metadata.ValueKey:     rawValue.Value,
			}
		}
		result[metadata.LogLinesKey] = values
		results[i] = result
	}

	return results
}

func serializeFlatStreamValues(data []client.StreamValues, labels map[string]metadata.LabelInfo, fields map[string]metadata.ModelFieldValue, runtimeSettings *metadata.RuntimeSettings) []map[string]any {
	var results []map[string]any
	for _, item := range data {
		valueLen := len(item.Values)
		if valueLen == 0 {
			continue
		}

		for _, rawValue := range item.Values {
			result := extractLabelAndFieldValues(item.Stream, labels, fields)
			result[metadata.TimestampKey] = formatTimestamp(rawValue.Time, runtimeSettings.Format.Timestamp)
			result[metadata.LogLineKey] = rawValue.Value
			results = append(results, result)
		}
	}

	return results
}

// serialize log stream values to maps
func serializeMetricMatrix(data []client.MatrixValues, labels map[string]metadata.LabelInfo, fields map[string]metadata.ModelFieldValue, runtimeSettings *metadata.RuntimeSettings, flat bool) []map[string]any {
	if flat {
		return serializeFlatMetricMatrix(data, labels, fields, runtimeSettings)
	}

	return serializeGroupedMetricMatrix(data, labels, fields, runtimeSettings)
}

func serializeGroupedMetricMatrix(data []client.MatrixValues, labels map[string]metadata.LabelInfo, fields map[string]metadata.ModelFieldValue, runtimeSettings *metadata.RuntimeSettings) []map[string]any {
	results := make([]map[string]any, len(data))
	for i, item := range data {
		valueLen := len(item.Values)
		if valueLen == 0 {
			continue
		}
		lastValue := item.Values[valueLen-1]
		result := extractLabelAndFieldValues(item.Metric, labels, fields)
		result[metadata.TimestampKey] = formatTimestamp(lastValue.Time, runtimeSettings.Format.Timestamp)
		result[metadata.MetricValueKey] = formatValue(lastValue.Value, runtimeSettings.Format)

		for label := range labels {
			result[label] = string(item.Metric[label])
		}
		for key := range fields {
			result[key] = string(item.Metric[key])
		}
		values := make([]map[string]any, len(item.Values))
		for j, rawValue := range item.Values {
			ts := formatTimestamp(rawValue.Time, runtimeSettings.Format.Timestamp)
			value := formatValue(rawValue.Value, runtimeSettings.Format)
			values[j] = map[string]any{
				metadata.TimestampKey: ts,
				metadata.ValueKey:     value,
			}
		}
		result[metadata.MetricValuesKey] = values
		results[i] = result
	}

	return results
}

func serializeFlatMetricMatrix(data []client.MatrixValues, labels map[string]metadata.LabelInfo, fields map[string]metadata.ModelFieldValue, runtimeSettings *metadata.RuntimeSettings) []map[string]any {
	var results []map[string]any
	for _, item := range data {
		valueLen := len(item.Values)
		if valueLen == 0 {
			continue
		}

		for _, rawValue := range item.Values {
			ts := formatTimestamp(rawValue.Time, runtimeSettings.Format.Timestamp)
			value := formatValue(rawValue.Value, runtimeSettings.Format)
			result := extractLabelAndFieldValues(item.Metric, labels, fields)
			result[metadata.TimestampKey] = ts
			result[metadata.MetricValueKey] = value
			results = append(results, result)
		}
	}

	return results
}

func serializeMetricVector(data []client.VectorValue, labels map[string]metadata.LabelInfo, fields map[string]metadata.ModelFieldValue, runtimeSettings *metadata.RuntimeSettings, flat bool) []map[string]any {
	results := make([]map[string]any, 0)
	for _, item := range data {
		ts := formatTimestamp(item.Time, runtimeSettings.Format.Timestamp)
		value := formatValue(item.Value, runtimeSettings.Format)

		result := extractLabelAndFieldValues(item.Metric, labels, fields)
		result[metadata.TimestampKey] = ts
		result[metadata.MetricValueKey] = value
		if !flat {
			result[metadata.MetricValuesKey] = []map[string]any{
				{
					metadata.TimestampKey: ts,
					metadata.ValueKey:     value,
				},
			}
		}
		results = append(results, result)
	}

	return results
}
