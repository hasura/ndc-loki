package metadata

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/hasura/ndc-loki/connector/client"
	"go.opentelemetry.io/otel/trace"
)

var (
	errModelInfoLabelsRequired = errors.New("model must define at least 1 label")
)

// State the shared state of the connector.
type State struct {
	Client *client.Client
	Tracer trace.Tracer
}

// TimestampFormat the format for timestamp serialization
type TimestampFormat string

const (
	// Represents the timestamp as a Unix timestamp in RFC3339 string.
	TimestampRFC3339 TimestampFormat = "rfc3339"
	// Represents the timestamp as a Unix timestamp in seconds.
	TimestampUnix TimestampFormat = "unix"
	// Represents the timestamp as a Unix timestamp in milliseconds.
	TimestampUnixMilli TimestampFormat = "unix_ms"
	// Represents the timestamp as a Unix timestamp in microseconds.
	TimestampUnixMicro TimestampFormat = "unix_us"
	// Represents the timestamp as a Unix timestamp in nanoseconds.
	TimestampUnixNano TimestampFormat = "unix_ns"
)

// ValueFormat the format for value serialization
type ValueFormat string

const (
	ValueString  ValueFormat = "string"
	ValueFloat64 ValueFormat = "float64"
)

// UnixTimeUnit the unit for unix timestamp
type UnixTimeUnit string

const (
	UnixTimeSecond UnixTimeUnit = "s"
	UnixTimeMilli  UnixTimeUnit = "ms"
	UnixTimeMicro  UnixTimeUnit = "us"
	UnixTimeNano   UnixTimeUnit = "ns"
)

// Duration returns the duration of the unit
func (ut UnixTimeUnit) Duration() time.Duration {
	switch ut {
	case UnixTimeMilli:
		return time.Millisecond
	case UnixTimeMicro:
		return time.Microsecond
	case UnixTimeNano:
		return time.Nanosecond
	default:
		return time.Second
	}
}

// ConversionFunction represents the conversion function enum
type ConversionFunction string

const (
	ConversionDuration        ConversionFunction = "duration"
	ConversionDurationSeconds ConversionFunction = "duration_seconds"
	ConversionBytes           ConversionFunction = "bytes"
)

var enumValues_ConversionFunction = []string{
	string(ConversionDuration),
	string(ConversionDurationSeconds),
	string(ConversionBytes),
}

// ParseConversionFunction parse the conversion function enum from string
func ParseConversionFunction(input string) (ConversionFunction, error) {
	if !slices.Contains(enumValues_ConversionFunction, input) {
		return "", fmt.Errorf("ConversionFunction value must be one of %v, got: %s", enumValues_ConversionFunction, input)
	}

	return ConversionFunction(input), nil
}

// Ordering represents a sort order enum
type Ordering string

const (
	Ascending  Ordering = "asc"
	Descending Ordering = "desc"
)

var enumValues_Ordering = []string{string(Ascending), string(Descending)}

// ParseOrdering parse the conversion function enum from string
func ParseOrdering(input string) (Ordering, error) {
	if !slices.Contains(enumValues_Ordering, input) {
		return "", fmt.Errorf("Ordering value must be one of %v, got: %s", enumValues_Ordering, input)
	}

	return Ordering(input), nil
}
