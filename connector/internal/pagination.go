package internal

import (
	"math"
	"slices"
	"strings"

	"github.com/hasura/ndc-loki/connector/client"
	"github.com/hasura/ndc-loki/connector/metadata"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/prometheus/common/model"
)

func sortVector(vector []client.VectorValue, sortElements []ColumnOrder) {
	if len(sortElements) == 0 {
		return
	}

	slices.SortFunc(vector, func(a client.VectorValue, b client.VectorValue) int {
		for _, elem := range sortElements {
			iOrder := 1
			if elem.Descending {
				iOrder = -1
			}
			switch elem.Name {
			case metadata.MetricValueKey:
				delta := compareMetricValue(a.Value, b.Value)
				if delta == 0 {
					continue
				}

				return delta * iOrder
			case metadata.TimestampKey:
				difference := a.Time.Sub(b.Time)
				if difference == 0 {
					continue
				}

				return int(difference) * iOrder
			default:
				if len(a.Metric) == 0 {
					continue
				}
				labelA, okA := a.Metric[elem.Name]
				labelB, okB := b.Metric[elem.Name]
				if !okA && !okB {
					continue
				}
				difference := strings.Compare(string(labelA), string(labelB))
				if difference == 0 {
					continue
				}

				return difference * iOrder
			}
		}

		return 0
	})
}

func compareMetricValue(a model.SampleValue, b model.SampleValue) int {
	if a.Equal(b) {
		return 0
	}
	if math.IsNaN(float64(a)) {
		return 1
	}
	if math.IsNaN(float64(b)) {
		return -1
	}
	if a > b {
		return 1
	}

	return -1
}

func sortMatrix(matrix []client.MatrixValues, sortElements []ColumnOrder) {
	if len(sortElements) == 0 {
		return
	}

	slices.SortFunc(matrix, func(a client.MatrixValues, b client.MatrixValues) int {
		for _, elem := range sortElements {
			iOrder := 1
			if elem.Descending {
				iOrder = -1
			}
			switch elem.Name {
			case metadata.MetricValueKey, metadata.TimestampKey:
				sortSamplePair(a.Values, elem.Name, iOrder)
				sortSamplePair(b.Values, elem.Name, iOrder)
			default:
				if len(a.Metric) == 0 {
					continue
				}
				labelA, okA := a.Metric[elem.Name]
				labelB, okB := b.Metric[elem.Name]
				if !okA && !okB {
					continue
				}
				difference := strings.Compare(string(labelA), string(labelB))
				if difference == 0 {
					continue
				}

				return difference * iOrder
			}
		}

		return 0
	})
}

func sortSamplePair(values []client.MatrixValue, key string, iOrder int) {
	slices.SortFunc(values, func(a client.MatrixValue, b client.MatrixValue) int {
		switch key {
		case metadata.MetricValueKey:
			if a.Value.Equal(b.Value) {
				return 0
			}
			if math.IsNaN(float64(a.Value)) {
				return 1 * iOrder
			}
			if math.IsNaN(float64(b.Value)) {
				return -1 * iOrder
			}
			if a.Value > b.Value {
				return 1 * iOrder
			} else {
				return -1 * iOrder
			}
		case metadata.TimestampKey:
			return int(a.Time.Sub(b.Time)) * iOrder
		default:
			return 0
		}
	})
}

func paginateVector(vector []client.VectorValue, q schema.Query) []client.VectorValue {
	if q.Offset != nil && *q.Offset > 0 {
		if len(vector) <= *q.Offset {
			return []client.VectorValue{}
		}
		vector = vector[*q.Offset:]
	}
	if q.Limit != nil && *q.Limit < len(vector) {
		vector = vector[:*q.Limit]
	}

	return vector
}

func paginateQueryResults(results []map[string]any, q schema.Query) []map[string]any {
	if q.Offset != nil && *q.Offset > 0 {
		if len(results) <= *q.Offset {
			return []map[string]any{}
		}
		results = results[*q.Offset:]
	}

	if q.Limit != nil && *q.Limit < len(results) {
		results = results[:*q.Limit]
	}

	return results
}
