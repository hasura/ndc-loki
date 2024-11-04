package metadata

import (
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
)

type ScalarName string

const (
	ScalarBoolean            ScalarName = "Boolean"
	ScalarInt64              ScalarName = "Int64"
	ScalarFloat64            ScalarName = "Float64"
	ScalarString             ScalarName = "String"
	ScalarLogLine            ScalarName = "LogLine"
	ScalarDecimal            ScalarName = "Decimal"
	ScalarTimestamp          ScalarName = "Timestamp"
	ScalarLabel              ScalarName = "Label"
	ScalarLabelSet           ScalarName = "LabelSet"
	ScalarDuration           ScalarName = "Duration"
	ScalarJSON               ScalarName = "JSON"
	ScalarConversionFunction ScalarName = "ConversionFunction"
	ScalarOrdering           ScalarName = "Ordering"
)

const (
	Equal          = "_eq"
	NotEqual       = "_neq"
	Like           = "_like"
	ILike          = "_ilike"
	NotLike        = "_nlike"
	NotILike       = "_nilike"
	In             = "_in"
	NotIn          = "_nin"
	Regex          = "_regex"
	NotRegex       = "_nregex"
	Least          = "_lt"
	LeastOrEqual   = "_lte"
	Greater        = "_gt"
	GreaterOrEqual = "_gte"
	ContainIP      = "_ip"
	NotContainIP   = "_nip"
	Since          = "_since"
)

const (
	TimestampKey          = "timestamp"
	ValueKey              = "value"
	LogLineKey            = "log_line"
	LogLinesKey           = "log_lines"
	MetricValueKey        = "metric_value"
	MetricValuesKey       = "metric_values"
	OriginalLabelsKey     = "original_labels"
	KKey                  = "k"
	ByKey                 = "by"
	WithoutKey            = "without"
	RangeKey              = "range"
	UnwrapKey             = "unwrap"
	ConversionFunctionKey = "conversion_function"
	QuantileKey           = "quantile"
)

var defaultScalars = map[string]schema.ScalarType{
	string(ScalarLabel): {
		AggregateFunctions: schema.ScalarTypeAggregateFunctions{},
		ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{
			Equal:    schema.NewComparisonOperatorEqual().Encode(),
			In:       schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarJSON))).Encode(),
			NotEqual: schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLabel))).Encode(),
			Regex:    schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLabel))).Encode(),
			NotRegex: schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLabel))).Encode(),
			NotIn:    schema.NewComparisonOperatorCustom(schema.NewArrayType(schema.NewNamedType(string(ScalarLabel)))).Encode(),
		},
		Representation: schema.NewTypeRepresentationString().Encode(),
	},
	string(ScalarString): {
		AggregateFunctions: schema.ScalarTypeAggregateFunctions{},
		ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{
			Equal:        schema.NewComparisonOperatorEqual().Encode(),
			In:           schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarJSON))).Encode(),
			NotEqual:     schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLabel))).Encode(),
			Regex:        schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLabel))).Encode(),
			NotRegex:     schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLabel))).Encode(),
			NotIn:        schema.NewComparisonOperatorCustom(schema.NewArrayType(schema.NewNamedType(string(ScalarLabel)))).Encode(),
			ContainIP:    schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLabel))).Encode(),
			NotContainIP: schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLabel))).Encode(),
		},
		Representation: schema.NewTypeRepresentationString().Encode(),
	},
	string(ScalarLogLine): {
		AggregateFunctions: schema.ScalarTypeAggregateFunctions{},
		ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{
			ILike:        schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLogLine))).Encode(),
			Like:         schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLogLine))).Encode(),
			NotLike:      schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLogLine))).Encode(),
			NotILike:     schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLogLine))).Encode(),
			Regex:        schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLogLine))).Encode(),
			NotRegex:     schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLogLine))).Encode(),
			ContainIP:    schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLogLine))).Encode(),
			NotContainIP: schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarLogLine))).Encode(),
		},
		Representation: schema.NewTypeRepresentationString().Encode(),
	},
	string(ScalarDecimal): {
		AggregateFunctions: schema.ScalarTypeAggregateFunctions{},
		ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{
			Equal:          schema.NewComparisonOperatorEqual().Encode(),
			NotEqual:       schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarDecimal))).Encode(),
			Least:          schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarDecimal))).Encode(),
			LeastOrEqual:   schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarDecimal))).Encode(),
			Greater:        schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarDecimal))).Encode(),
			GreaterOrEqual: schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarDecimal))).Encode(),
		},
		Representation: schema.NewTypeRepresentationBigDecimal().Encode(),
	},
	string(ScalarDuration): {
		AggregateFunctions:  schema.ScalarTypeAggregateFunctions{},
		ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{},
		Representation:      schema.NewTypeRepresentationJSON().Encode(),
	},
	string(ScalarTimestamp): {
		AggregateFunctions: schema.ScalarTypeAggregateFunctions{},
		ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{
			Equal:   schema.NewComparisonOperatorEqual().Encode(),
			Least:   schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarTimestamp))).Encode(),
			Greater: schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarTimestamp))).Encode(),
			Since:   schema.NewComparisonOperatorCustom(schema.NewNamedType(string(ScalarDuration))).Encode(),
		},
		Representation: schema.NewTypeRepresentationTimestamp().Encode(),
	},
	string(ScalarLabelSet): {
		AggregateFunctions:  schema.ScalarTypeAggregateFunctions{},
		ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{},
		Representation:      schema.NewTypeRepresentationJSON().Encode(),
	},
	string(ScalarConversionFunction): {
		AggregateFunctions:  schema.ScalarTypeAggregateFunctions{},
		ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{},
		Representation:      schema.NewTypeRepresentationEnum(enumValues_ConversionFunction).Encode(),
	},
	string(ScalarOrdering): {
		AggregateFunctions:  schema.ScalarTypeAggregateFunctions{},
		ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{},
		Representation:      schema.NewTypeRepresentationEnum(enumValues_Ordering).Encode(),
	},
}

const (
	objectName_StreamItem  = "StreamItem"
	objectName_MetricItem  = "MetricItem"
	objectName_StreamValue = "StreamValue"
	objectName_MetricValue = "MetricValue"
)

var defaultObjectTypes = map[string]schema.ObjectType{
	objectName_StreamItem: {
		Description: utils.ToPtr("A log stream item of the query result"),
		Fields:      createStreamObjectFields(),
	},
	objectName_MetricItem: {
		Description: utils.ToPtr("A metric item of the query result"),
		Fields:      createMetricObjectFields(),
	},
	objectName_StreamValue: {
		Description: utils.ToPtr("A log stream value of the query result"),
		Fields:      createStreamValueObjectFields(),
	},
	objectName_MetricValue: {
		Description: utils.ToPtr("A metric value of the query result"),
		Fields:      createMetricValueObjectFields(),
	},
}

const (
	ArgumentKeyFlat         = "flat"
	ArgumentKeyTimeout      = "timeout"
	ArgumentKeyStep         = "step"
	ArgumentKeyInterval     = "interval"
	ArgumentKeyOffset       = "offset"
	ArgumentKeyAggregations = "aggregations"
)

var defaultArgumentInfos = map[string]schema.ArgumentInfo{
	ArgumentKeyTimeout: {
		Description: utils.ToPtr("Evaluation timeout"),
		Type:        schema.NewNullableNamedType(string(ScalarDuration)).Encode(),
	},
	ArgumentKeyStep: {
		Description: utils.ToPtr("Query resolution step width in duration format or float number of seconds"),
		Type:        schema.NewNullableNamedType(string(ScalarDuration)).Encode(),
	},
	ArgumentKeyInterval: {
		Description: utils.ToPtr("Only return entries at (or greater than) the specified interval, can be a duration format or float number of seconds"),
		Type:        schema.NewNullableNamedType(string(ScalarDuration)).Encode(),
	},
	ArgumentKeyOffset: {
		Description: utils.ToPtr("The offset modifier allows changing the time offset for individual instant and range vectors in a query"),
		Type:        schema.NewNullableNamedType(string(ScalarDuration)).Encode(),
	},
	ArgumentKeyFlat: {
		Description: utils.ToPtr("Flatten grouped values out the root array"),
		Type:        schema.NewNullableNamedType(string(ScalarBoolean)).Encode(),
	},
}

// AggregationOperator represents an aggregate operator enum
type AggregationOperator string

const (
	// Calculate sum over labels
	Sum AggregationOperator = "sum"
	// Calculate the average over labels
	Avg AggregationOperator = "avg"
	// Select minimum over labels
	Min AggregationOperator = "min"
	// Select maximum over labels
	Max AggregationOperator = "max"
	// Calculate the population standard deviation over labels
	Stddev AggregationOperator = "stddev"
	// Calculate the population standard variance over labels
	Stdvar AggregationOperator = "stdvar"
	// Count number of elements in the vector
	Count AggregationOperator = "count"
	// Select largest k elements by sample value
	TopK AggregationOperator = "topk"
	// Select smallest k elements by sample value
	BottomK AggregationOperator = "bottomk"
	// returns vector elements sorted by their sample values, in ascending order.
	Sort AggregationOperator = "sort"
	// Same as sort, but sorts in descending order.
	SortDesc AggregationOperator = "sort_desc"
	// calculates the number of entries per second
	Rate AggregationOperator = "rate"
	// calculates per second rate of the values in the specified interval and treating them as “counter metric”.
	RateCounter AggregationOperator = "rate_counter"
	// counts the entries for each log stream within the given range.
	CountOverTime AggregationOperator = "count_over_time"
	// calculates the number of bytes per second for each stream.
	BytesRate AggregationOperator = "bytes_rate"
	// counts the amount of bytes used by each log stream for a given range.
	BytesOverTime AggregationOperator = "bytes_over_time"
	// returns an empty vector if the range vector passed to it has any elements and a 1-element vector with the value 1 if the range vector passed to it has no elements.
	// (absent_over_time is useful for alerting on when no time series and logs stream exist for label combination for a certain amount of time.)
	AbsentOverTime AggregationOperator = "absent_over_time"
	// the average value of all points in the specified interval.
	AvgOverTime AggregationOperator = "avg_over_time"
	// the minimum value of all points in the specified interval
	MinOverTime AggregationOperator = "min_over_time"
	// the maximum value of all points in the specified interval.
	MaxOverTime AggregationOperator = "max_over_time"
	// the sum of all values in the specified interval.
	SumOverTime AggregationOperator = "sum_over_time"
	// the φ-quantile (0 ≤ φ ≤ 1) of the values in the specified interval.
	QuantileOverTime AggregationOperator = "quantile_over_time"
	// the population standard deviation of the values in the specified interval.
	StddevOverTime AggregationOperator = "stddev_over_time"
	// the population standard variance of the values in the specified interval.
	StdvarOverTime AggregationOperator = "stdvar_over_time"
	// the first value of all points in the specified interval
	FirstOverTime AggregationOperator = "first_over_time"
	// the last value of all points in the specified interval
	LastOverTime AggregationOperator = "last_over_time"
)

var lineFilterOperators = map[string]string{
	Like:         "|=",
	NotLike:      "!=",
	ILike:        "|~",
	NotILike:     "!~",
	Regex:        "|~",
	NotRegex:     "!~",
	ContainIP:    "|=",
	NotContainIP: "!=",
}

var labelFilterOperators = map[string]string{
	Equal:          "=",
	NotEqual:       "!=",
	Least:          "<",
	LeastOrEqual:   "<=",
	Regex:          "=~",
	NotRegex:       "!~",
	In:             "=~",
	NotIn:          "!~",
	Greater:        ">",
	GreaterOrEqual: ">=",
	ContainIP:      "=",
	NotContainIP:   "!=",
}
