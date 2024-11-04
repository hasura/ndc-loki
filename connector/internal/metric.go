package internal

import (
	"fmt"

	"github.com/hasura/ndc-loki/connector/metadata"
	"github.com/hasura/ndc-sdk-go/utils"
)

func (qce *QueryCollectionExecutor) buildMetricQuery(query string, predicate *CollectionRequest) (string, error) {
	valueCondition, err := qce.evalValueComparisonCondition(predicate.MetricValue)
	if err != nil {
		return "", err
	}

	rawOffset, ok := qce.Arguments[metadata.ArgumentKeyOffset]
	if ok {
		offset, err := utils.DecodeDuration(rawOffset, utils.WithBaseUnix(qce.Runtime.UnixTimeUnit.Duration()))
		if err != nil {
			return "", fmt.Errorf("invalid offset argument `%v`", rawOffset)
		}
		if offset > 0 {
			query = fmt.Sprintf("%s offset %s", query, offset.String())
		}
	}

	for _, fn := range predicate.Aggregations {
		query, err = qce.evalAggregation(query, fn)
		if err != nil {
			return "", err
		}
	}

	query += valueCondition

	return query, nil
}

func (qce *QueryCollectionExecutor) evalAggregation(query string, fn KeyValue) (string, error) {
	switch metadata.AggregationOperator(fn.Key) {
	case metadata.RateCounter, metadata.AvgOverTime, metadata.MinOverTime, metadata.MaxOverTime, metadata.SumOverTime, metadata.StddevOverTime, metadata.StdvarOverTime, metadata.FirstOverTime, metadata.LastOverTime:
		input, ok := fn.Value.(map[string]any)
		if !ok {
			return "", fmt.Errorf("%s: invalid value, expected object, got %v", fn.Key, fn.Value)
		}
		rangeUnwrap := &AggregateRangeUnwrapInput{}
		if err := rangeUnwrap.FromValue(input, qce.Runtime.UnixTimeUnit); err != nil {
			return "", fmt.Errorf("%s: %w", fn.Key, err)
		}

		return fmt.Sprintf("%s(%s %s)", fn.Key, query, rangeUnwrap.String()), nil
	case metadata.Sum, metadata.Avg, metadata.Min, metadata.Max, metadata.Count, metadata.Stddev, metadata.Stdvar:
		input, ok := fn.Value.(map[string]any)
		if !ok {
			return "", fmt.Errorf("%s: invalid value, expected object, got %v", fn.Key, fn.Value)
		}
		groupBy := &AggregateGroupBy{}
		if err := groupBy.FromValue(input); err != nil {
			return "", fmt.Errorf("%s: %w", fn.Key, err)
		}

		return fmt.Sprintf("%s%s (%s)", fn.Key, groupBy.String(), query), nil
	case metadata.BottomK, metadata.TopK:
		input, ok := fn.Value.(map[string]any)
		if !ok {
			return "", fmt.Errorf("%s: invalid value, expected object, got %v", fn.Key, fn.Value)
		}
		kArgument := &AggregateK{}
		if err := kArgument.FromValue(input); err != nil {
			return "", fmt.Errorf("%s: %w", fn.Key, err)
		}

		return fmt.Sprintf("%s%s (%d, %s)", fn.Key, kArgument.AggregateGroupBy.String(), kArgument.K, query), nil
	case metadata.QuantileOverTime:
		input, ok := fn.Value.(map[string]any)
		if !ok {
			return "", fmt.Errorf("%s: invalid value, expected object, got %v", fn.Key, fn.Value)
		}
		qot := &QuantileOverTimeInput{}
		if err := qot.FromValue(input, qce.Runtime.UnixTimeUnit); err != nil {
			return "", fmt.Errorf("%s: %w", fn.Key, err)
		}

		return fmt.Sprintf("%s(%f, %s %s)", fn.Key, qot.Quantile, query, qot.AggregateRangeUnwrapInput.String()), nil
	case metadata.AbsentOverTime, metadata.Rate, metadata.BytesRate, metadata.CountOverTime, metadata.BytesOverTime:
		rng, err := utils.DecodeDuration(fn.Value, utils.WithBaseUnix(qce.Runtime.UnixTimeUnit.Duration()))
		if err != nil {
			return "", fmt.Errorf("%s: %s", fn.Key, err)
		}

		query = fmt.Sprintf(`%s(%s [%s])`, fn.Key, query, rng.String())

		return query, nil
	case metadata.Sort:
		rawOrdering, err := utils.DecodeString(fn.Value)
		if err != nil {
			return "", fmt.Errorf("%s: %s", fn.Key, err)
		}
		ordering, err := metadata.ParseOrdering(rawOrdering)
		if err != nil {
			return "", fmt.Errorf("%s: %s", fn.Key, err)
		}
		orderFunc := metadata.Sort
		if ordering == metadata.Descending {
			orderFunc = metadata.SortDesc
		}

		return fmt.Sprintf("%s(%s)", orderFunc, query), nil
	default:
		return "", fmt.Errorf("unsupported aggregate function name `%s`", fn.Key)
	}
}
