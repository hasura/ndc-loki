package internal

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"time"

	"github.com/hasura/ndc-loki/connector/metadata"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/prometheus/common/model"
)

func formatTimestamp(ts time.Time, format metadata.TimestampFormat) any {
	switch format {
	case metadata.TimestampUnix:
		return ts.Unix()
	case metadata.TimestampUnixMilli:
		return ts.UnixMilli()
	case metadata.TimestampUnixMicro:
		return ts.UnixMicro()
	case metadata.TimestampUnixNano:
		return strconv.FormatInt(ts.UnixNano(), 10)
	default:
		return ts.Format(time.RFC3339)
	}
}

func formatValue(value model.SampleValue, format metadata.RuntimeFormatSettings) any {
	switch format.Value {
	case metadata.ValueFloat64:
		if math.IsNaN(float64(value)) {
			return format.NaN
		}
		if value > 0 && math.IsInf(float64(value), 1) {
			return format.Inf
		}
		if value < 0 && math.IsInf(float64(value), -1) {
			return format.NegativeInf
		}

		return float64(value)
	default:
		return value.String()
	}
}

func intersection[T comparable](sliceA []T, sliceB []T) []T {
	var result []T
	if len(sliceA) == 0 || len(sliceB) == 0 {
		return result
	}

	for _, a := range sliceA {
		if slices.Contains(sliceB, a) {
			result = append(result, a)
		}
	}

	return result
}

func getComparisonValue(input schema.ComparisonValue, variables map[string]any) (any, error) {
	if len(input) == 0 {
		return nil, nil
	}

	switch v := input.Interface().(type) {
	case *schema.ComparisonValueScalar:
		return v.Value, nil
	case *schema.ComparisonValueVariable:
		if value, ok := variables[v.Name]; ok {
			return value, nil
		}

		return nil, fmt.Errorf("variable %s does not exist", v.Name)
	default:
		return nil, fmt.Errorf("invalid comparison value: %v", input)
	}
}

func getComparisonValueFloat64(input schema.ComparisonValue, variables map[string]any) (*float64, error) {
	rawValue, err := getComparisonValue(input, variables)
	if err != nil {
		return nil, err
	}

	return utils.DecodeNullableFloat[float64](rawValue)
}

func getComparisonValueString(input schema.ComparisonValue, variables map[string]any) (*string, error) {
	rawValue, err := getComparisonValue(input, variables)
	if err != nil {
		return nil, err
	}

	return utils.DecodeNullableString(rawValue)
}

func getComparisonValueTimestamp(input schema.ComparisonValue, variables map[string]any, unixTimeUnit metadata.UnixTimeUnit) (*time.Time, error) {
	rawValue, err := getComparisonValue(input, variables)
	if err != nil {
		return nil, err
	}

	return utils.DecodeNullableDateTime(rawValue, utils.WithBaseUnix(unixTimeUnit.Duration()))
}

func getComparisonValueDuration(input schema.ComparisonValue, variables map[string]any, unixTimeUnit metadata.UnixTimeUnit) (*time.Duration, error) {
	rawValue, err := getComparisonValue(input, variables)
	if err != nil {
		return nil, err
	}

	return utils.DecodeNullableDuration(rawValue, utils.WithBaseUnix(unixTimeUnit.Duration()))
}

func getComparisonValueStringSlice(input schema.ComparisonValue, variables map[string]any) ([]string, error) {
	rawValue, err := getComparisonValue(input, variables)
	if err != nil {
		return nil, err
	}

	return metadata.DecodeStringSlice(rawValue)
}

func evalCollectionOrderBy(orderBy *schema.OrderBy) ([]ColumnOrder, error) {
	var results []ColumnOrder
	if orderBy == nil {
		return results, nil
	}
	for _, elem := range orderBy.Elements {
		switch target := elem.Target.Interface().(type) {
		case *schema.OrderByColumn:
			if slices.Contains([]string{metadata.OriginalLabelsKey, metadata.LogLineKey, metadata.MetricValuesKey}, target.Name) {
				return nil, fmt.Errorf("ordering by `%s` is unsupported", target.Name)
			}

			orderBy := ColumnOrder{
				Name:       target.Name,
				Descending: elem.OrderDirection == schema.OrderDirectionDesc,
			}
			results = append(results, orderBy)
		default:
			return nil, fmt.Errorf("support ordering by column only, got: %v", elem.Target)
		}
	}

	return results, nil
}

func getLabelInfosFromModelLabels(input map[string]metadata.ModelLabelInfo) map[string]metadata.LabelInfo {
	results := make(map[string]metadata.LabelInfo)
	for key, info := range input {
		results[key] = info.LabelInfo
	}

	return results
}

// merge label expression maps
func mergeLabelExpressions(dest map[string]LabelExpression, others ...map[string]LabelExpression) map[string]LabelExpression {
	if dest == nil {
		dest = make(map[string]LabelExpression)
	}
	for _, item := range others {
		for key, value := range item {
			existedItem, ok := dest[key]
			if ok {
				existedItem.Expressions = append(existedItem.Expressions, value.Expressions...)
				dest[key] = existedItem
			} else {
				dest[key] = value
			}
		}
	}

	return dest
}

func extractLabelAndFieldValues(data map[string]string, labels map[string]metadata.LabelInfo, fields map[string]metadata.ModelFieldValue) map[string]any {
	result := map[string]any{
		metadata.OriginalLabelsKey: data,
	}
	for label, info := range labels {
		key := label
		if info.Source != "" {
			key = info.Source
		}
		value := data[key]
		result[label] = value
	}

	for key := range fields {
		result[key] = string(data[key])
	}

	return result
}
