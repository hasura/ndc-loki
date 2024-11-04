package internal

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/hasura/ndc-loki/connector/client"
	"github.com/hasura/ndc-loki/connector/metadata"
	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
)

var (
	errTimestampUnsupportedMultipleLtExpression    = errors.New("unsupported multiple _lt expressions for the timestamp")
	errTimestampUnsupportedMultipleGtExpression    = errors.New("unsupported multiple _gt expressions for the timestamp")
	errTimestampUnsupportedMultipleSinceExpression = errors.New("unsupported multiple _since expressions for the timestamp")
	errMetricValueUnsupportedMultipleComparisons   = errors.New("unsupported multiple comparisons for the metric value")
	errStreamValueUnsupportedMultipleComparisons   = errors.New("unsupported multiple comparisons for the stream value")
	errUnsupportedMixedLabelAndFieldExpressions    = errors.New("unsupported mixed label and field expressions in _and and _or operators")
	errUnsupportedOrLabels                         = errors.New("unsupported _or operator for labels")
	errQueryAggregationsRequired                   = errors.New("`aggregations` argument is required")
	errQuantileOverflow                            = errors.New("the φ-quantile must be in between 0 and 1")
	errOnlyOneAggregationAllowed                   = errors.New("each aggregation item must have 1 function only")
)

var valueBinaryOperators = map[string]string{
	metadata.Equal:          "=",
	metadata.NotEqual:       "!=",
	metadata.Least:          "<",
	metadata.LeastOrEqual:   "<=",
	metadata.Regex:          "=~",
	metadata.NotRegex:       "!~",
	metadata.In:             "=~",
	metadata.NotIn:          "!~",
	metadata.Greater:        ">",
	metadata.GreaterOrEqual: ">=",
	metadata.ContainIP:      "=",
	metadata.NotContainIP:   "!=",
}

// ColumnOrder the structured sorting columns
type ColumnOrder struct {
	Name       string
	Descending bool
}

// KeyValue represents a key-value pair
type KeyValue struct {
	Key   string
	Value any
}

// BaseQueryRequest represents the common query request structure
type BaseQueryRequest struct {
	Timestamp *time.Time
	Start     *time.Time
	End       *time.Time
	Since     *scalar.Duration
	Interval  *scalar.Duration
	Step      *scalar.Duration
	Flat      *bool
	OrderBy   []ColumnOrder
	Variables map[string]any

	runtime *metadata.RuntimeSettings
}

// GetFlat gets the flat boolean value
func (bqr BaseQueryRequest) GetFlat() bool {
	return bqr.Flat != nil && *bqr.Flat
}

// FindTimestampOrderBy checks and returns the first timestamp order by item if exist
func (bqr BaseQueryRequest) FindTimestampOrderBy() *ColumnOrder {
	if len(bqr.OrderBy) > 0 && bqr.OrderBy[0].Name == metadata.TimestampKey {
		return &bqr.OrderBy[0]
	}

	return nil
}

// GetQueryDirection gets the query direction if exist
func (bqr BaseQueryRequest) GetQueryDirection() *client.QueryDirection {
	tsOrder := bqr.FindTimestampOrderBy()
	if tsOrder == nil {
		return nil
	}
	direction := client.QueryDirectionBackward
	if tsOrder.Descending {
		direction = client.QueryDirectionBackward
	}

	return &direction
}

func (bqr *BaseQueryRequest) evalTimestampFromBinaryComparisonOperator(expr *schema.ExpressionBinaryComparisonOperator, variables map[string]any, unixTimeUnix metadata.UnixTimeUnit) error {
	switch expr.Operator {
	case metadata.Equal:
		if bqr.Timestamp != nil {
			return errTimestampUnsupportedMultipleLtExpression
		}
		ts, err := getComparisonValueTimestamp(expr.Value, variables, unixTimeUnix)
		if err != nil {
			return err
		}
		bqr.Timestamp = ts
	case metadata.Least:
		if bqr.End != nil {
			return errTimestampUnsupportedMultipleLtExpression
		}
		ts, err := getComparisonValueTimestamp(expr.Value, variables, unixTimeUnix)
		if err != nil {
			return err
		}
		bqr.End = ts
	case metadata.Greater:
		if bqr.Start != nil {
			return errTimestampUnsupportedMultipleGtExpression
		}
		ts, err := getComparisonValueTimestamp(expr.Value, variables, unixTimeUnix)
		if err != nil {
			return err
		}
		bqr.Start = ts
	case metadata.Since:
		if bqr.Since != nil {
			return errTimestampUnsupportedMultipleGtExpression
		}
		dur, err := getComparisonValueDuration(expr.Value, variables, unixTimeUnix)
		if err != nil {
			return err
		}
		if dur != nil {
			bqr.Since = &scalar.Duration{Duration: *dur}
		}
	default:
		return fmt.Errorf("unsupported operator `%s` for the timestamp", expr.Operator)
	}

	return nil
}

// AggregateK represents the object argument of topK and bottomK object
type AggregateK struct {
	AggregateGroupBy

	K int64
}

// FromValue decodes property values from an input object
func (agb *AggregateK) FromValue(input map[string]any) error {
	k, err := utils.GetInt[int64](input, metadata.KKey)
	if err != nil {
		return err
	}

	agb.K = k

	return agb.AggregateGroupBy.FromValue(input)
}

// AggregateGroupBy represents the aggregate group by object
type AggregateGroupBy struct {
	By      []string
	Without []string
}

// FromValue decodes property values from an input object
func (agb *AggregateGroupBy) FromValue(input map[string]any) error {
	groupBy, err := utils.GetNullableStringSlice(input, metadata.ByKey)
	if err != nil {
		return err
	}
	if groupBy != nil {
		agb.By = *groupBy
	}

	without, err := utils.GetNullableStringSlice(input, metadata.WithoutKey)
	if err != nil {
		return err
	}
	if without != nil {
		agb.Without = *without
	}

	return nil
}

// String implement the fmt.Stringer interface
func (agb AggregateGroupBy) String() string {
	byLength := len(agb.By)
	withoutLength := len(agb.Without)
	if byLength == 0 && withoutLength == 0 {
		return ""
	}

	if withoutLength == 0 {
		return fmt.Sprintf(" by (%s)", strings.Join(agb.By, ", "))
	}

	sameValues := intersection(agb.By, agb.Without)
	sameValuesLength := len(sameValues)
	if byLength == 0 || sameValuesLength == 0 {
		return fmt.Sprintf(" without (%s) ", strings.Join(agb.Without, ", "))
	}

	without := slices.DeleteFunc(agb.Without, func(item string) bool {
		return slices.Contains(sameValues, item)
	})

	if len(without) == 0 {
		return ""
	}

	return fmt.Sprintf(" without (%s) ", strings.Join(without, ", "))
}

// AggregateRangeUnwrapInput represents an aggregate range with unwrap input argument
type AggregateRangeUnwrapInput struct {
	Range              time.Duration
	Unwrap             string
	ConversionFunction *metadata.ConversionFunction
}

// FromValue decodes property values from an input object
func (aru *AggregateRangeUnwrapInput) FromValue(input map[string]any, unitTimeUnit metadata.UnixTimeUnit) error {
	rng, err := utils.GetDuration(input, metadata.RangeKey, utils.WithBaseUnix(unitTimeUnit.Duration()))
	if err != nil {
		return fmt.Errorf("%s: %w", metadata.RangeKey, err)
	}
	aru.Range = rng

	aru.Unwrap, err = utils.GetString(input, metadata.UnwrapKey)
	if err != nil {
		return fmt.Errorf("%s: %w", metadata.UnwrapKey, err)
	}

	if aru.Unwrap == "" {
		return fmt.Errorf(metadata.UnwrapKey + ": label name required")
	}

	rawCf, err := utils.GetNullableString(input, metadata.ConversionFunctionKey)
	if err != nil {
		return fmt.Errorf("%s: %w", metadata.ConversionFunctionKey, err)
	}

	if rawCf != nil {
		cf, err := metadata.ParseConversionFunction(*rawCf)
		if err != nil {
			return fmt.Errorf("%s: %w", metadata.ConversionFunctionKey, err)
		}
		aru.ConversionFunction = &cf
	}

	return nil
}

// String implement the fmt.Stringer interface
func (aru AggregateRangeUnwrapInput) String() string {
	var sb strings.Builder
	sb.WriteString("| unwrap ")
	if aru.ConversionFunction != nil {
		sb.WriteString(string(*aru.ConversionFunction))
		sb.WriteRune('(')
		sb.WriteString(aru.Unwrap)
		sb.WriteRune(')')
	} else {
		sb.WriteString(aru.Unwrap)
	}

	sb.WriteString(" [")
	sb.WriteString(aru.Range.String())
	sb.WriteRune(']')

	return sb.String()
}

// QuantileOverTimeInput represents an input argument for the QuantileOverTime aggregation
type QuantileOverTimeInput struct {
	AggregateRangeUnwrapInput

	Quantile float64
}

// FromValue decodes property values from an input object
func (qot *QuantileOverTimeInput) FromValue(input map[string]any, unitTimeUnit metadata.UnixTimeUnit) error {
	quantile, err := utils.GetFloat[float64](input, metadata.QuantileKey)
	if err != nil {
		return fmt.Errorf("%s: %w", metadata.QuantileKey, err)
	}
	if quantile < 0 || quantile > 1 {
		return errQuantileOverflow
	}
	qot.Quantile = quantile

	if err := qot.AggregateRangeUnwrapInput.FromValue(input, unitTimeUnit); err != nil {
		return err
	}

	return nil
}
