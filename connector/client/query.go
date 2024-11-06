package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/prometheus/common/model"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// QueryDirection determines the sort order of logs
// @enum forward,backward
type QueryDirection string

// QueryRangeParams request parameters for the query range function
type QueryRangeParams struct {
	// The LogQL query to perform.
	Query string `json:"query"`
	// The max number of entries to return. It defaults to 100. Only applies to query types which produce a stream (log lines) response.
	Limit int `json:"limit,omitempty"`
	// The start time for the query as a nanosecond Unix epoch or RFC3339 format. Defaults to one hour ago.
	// Loki returns results with timestamp greater or equal to this value.
	Start *time.Time `json:"start"`
	// The end time for the query as a nanosecond Unix epoch or RFC3339 format.
	// Defaults to now. Loki returns results with timestamp lower than this value.
	End *time.Time `json:"end"`
	// A duration used to calculate start relative to end. If end is in the future, start is calculated as this duration before now.
	// Any value specified for start supersedes this parameter.
	Since *scalar.Duration `json:"since"`
	// Query resolution step width in duration format or float number of seconds.
	// Only applies to query types which produce a matrix response.
	Step *scalar.Duration `json:"step"`
	// Only return entries at (or greater than) the specified interval, can be a duration format or float number of seconds.
	// Only applies to queries which produce a stream response.
	Interval *scalar.Duration `json:"interval"`
	// Determines the sort order of logs. Supported values are forward or backward. Defaults to backward.
	Direction *QueryDirection `json:"direction"`
}

// QueryRange [queries logs within a range of time]. This type of query is often referred to as a range query.
// Range queries are used for both log and metric type LogQL queries
//
// [queries logs within a range of time]: https://grafana.com/docs/loki/latest/reference/loki-http-api/#query-logs-within-a-range-of-time
func (c *Client) QueryRange(ctx context.Context, params *QueryRangeParams) (*QueryRangeData, error) {
	if params == nil || params.Query == "" {
		return nil, errQueryRequired
	}
	req, span, cancel, err := c.createRequestSpan(ctx, http.MethodGet, "/loki/api/v1/query_range", nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	span.SetAttributes(attribute.String("db.query.text", params.Query))

	q := req.URL.Query()
	if err := applyQueryTimeRange(&q, params.Start, params.End, params.Since, c.maxTimeRange); err != nil {
		span.SetStatus(codes.Error, "validation failure")
		span.RecordError(err)

		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	q.Set("query", params.Query)
	if params.Limit > 0 {
		q.Set("limit", strconv.FormatInt(int64(params.Limit), 32))
	}
	if params.Step != nil && params.Step.Duration > 0 {
		q.Set("step", params.Step.String())
	}
	if params.Interval != nil && params.Interval.Duration > 0 {
		q.Set("interval", params.Interval.String())
	}
	if params.Direction != nil {
		q.Set("direction", string(*params.Direction))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req, span)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result queryRangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.SetStatus(codes.Error, "failed to decode json response")
		span.RecordError(err)

		return nil, schema.InternalServerError(fmt.Sprintf("failed to decode json response: %s", err), nil)
	}

	return &result.Data, nil
}

// QueryParams request parameters for the query function
type QueryParams struct {
	// The LogQL query to perform.
	Query string `json:"query"`
	// The max number of entries to return. It defaults to 100. Only applies to query types which produce a stream (log lines) response.
	Limit int `json:"limit,omitempty"`
	// The evaluation time for the query as a nanosecond Unix epoch or RFC3339 format. Defaults to now.
	Time *time.Time `json:"time"`
	// Determines the sort order of logs. Supported values are forward or backward. Defaults to backward.
	Direction *QueryDirection `json:"direction"`
}

// Query allows for doing [queries against a single point in time]
//
// [queries against a single point in time]: https://grafana.com/docs/loki/latest/reference/loki-http-api/#query-logs-at-a-single-point-in-time
func (c *Client) Query(ctx context.Context, params *QueryParams) (*QueryData, error) {
	if params == nil || params.Query == "" {
		return nil, errQueryRequired
	}
	req, span, cancel, err := c.createRequestSpan(ctx, http.MethodGet, "/loki/api/v1/query", nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	span.SetAttributes(attribute.String("db.query.text", params.Query))

	q := req.URL.Query()
	q.Set("query", params.Query)
	if params.Time != nil {
		q.Set("time", FormatUnixNanoTimestamp(*params.Time))
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.FormatInt(int64(params.Limit), 32))
	}
	if params.Direction != nil {
		q.Set("direction", string(*params.Direction))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req, span)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result queryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.SetStatus(codes.Error, "failed to decode json response")
		span.RecordError(err)

		return nil, schema.InternalServerError(fmt.Sprintf("failed to decode json response: %s", err), nil)
	}

	return &result.Data, nil
}

// StreamValue holds a timestamp and the content of a log line.
type StreamValue struct {
	Time  time.Time `json:"time"`
	Value string    `json:"value"`
}

// StreamValues holds a label key value pairs for the log stream and a list of values
type StreamValues struct {
	Stream map[string]string `json:"stream"`
	Values []StreamValue     `json:"values"`
}

type rawStreamValues struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *StreamValues) UnmarshalJSON(b []byte) error {
	var s rawStreamValues
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	a.Stream = s.Stream
	a.Values = make([]StreamValue, len(s.Values))

	baseTime := time.Unix(0, 0)
	for i, value := range s.Values {
		if len(value) != 2 {
			return fmt.Errorf("unexpected value length %d", len(value))
		}
		intTime, err := strconv.ParseInt(value[0], 10, 64)
		if err != nil {
			return fmt.Errorf("failed to unmarshal stream value at index %d: %w", i, err)
		}

		t := baseTime.Add(time.Duration(intTime))
		a.Values[i] = StreamValue{
			Time:  t,
			Value: value[1],
		}
	}

	return nil
}

// MatrixValues holds a label key value pairs for the metric and a list of values
type MatrixValues struct {
	Metric map[string]string `json:"metric"`
	Values []MatrixValue     `json:"values"`
}

// MatrixValue holds a single timestamp and a metric value
type MatrixValue struct {
	Time  time.Time         `json:"time"`
	Value model.SampleValue `json:"value"`
}

type rawMatrixValues struct {
	Metric map[string]string `json:"metric"`
	Values [][]any           `json:"values"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *MatrixValues) UnmarshalJSON(b []byte) error {
	var s rawMatrixValues
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	a.Metric = s.Metric
	a.Values = make([]MatrixValue, len(s.Values))
	for i, value := range s.Values {
		if len(value) != 2 {
			return fmt.Errorf("unexpected value length %d", len(value))
		}
		item := MatrixValue{}
		if ts, ok := value[0].(int64); ok {
			item.Time = time.Unix(ts, 0)
		} else if ts, ok := value[0].(float64); ok {
			item.Time = time.Unix(int64(ts), 0)
		}

		if val, ok := value[1].(string); ok {
			floatValue, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("failed to decode metric value: %w", err)
			}
			item.Value = model.SampleValue(floatValue)
		}
		a.Values[i] = item
	}

	return nil
}

// VectorValue holds a label key value pairs for the metric and single timestamp and value
type VectorValue struct {
	Metric map[string]string `json:"metric"`
	Time   time.Time         `json:"time"`
	Value  model.SampleValue `json:"value"`
}

type rawVectorValue struct {
	Metric map[string]string `json:"metric"`
	Value  []any             `json:"value"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *VectorValue) UnmarshalJSON(b []byte) error {
	var s rawVectorValue
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	a.Metric = s.Metric
	if len(s.Value) != 2 {
		return fmt.Errorf("unexpected value length %d", len(s.Value))
	}
	if ts, ok := s.Value[0].(int64); ok {
		a.Time = time.Unix(ts, 0)
	} else if ts, ok := s.Value[0].(float64); ok {
		a.Time = time.Unix(int64(ts), 0)
	}
	if val, ok := s.Value[1].(string); ok {
		floatValue, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("failed to decode metric value: %w", err)
		}
		a.Value = model.SampleValue(floatValue)
	}

	return nil
}

// QueryResultType represents the enum of query result types
type QueryResultType string

const (
	ResultTypeVector  QueryResultType = "vector"
	ResultTypeStreams QueryResultType = "streams"
	ResultTypeMatrix  QueryResultType = "matrix"
)

// QueryData holds the result type and a metric vector value of the instant query response
type QueryData struct {
	ResultType    QueryResultType `json:"resultType"`
	Vector        []VectorValue   `json:"vector,omitempty"`
	EncodingFlags []string        `json:"encodingFlags,omitempty"`
}

type rawQueryData struct {
	ResultType    QueryResultType `json:"resultType"`
	EncodingFlags []string        `json:"encodingFlags"`
	Result        json.RawMessage `json:"result"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *QueryData) UnmarshalJSON(b []byte) error {
	var s rawQueryData
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s.ResultType {
	case ResultTypeVector:
		if err := json.Unmarshal(s.Result, &a.Vector); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown query result type %s", s.ResultType)
	}
	a.ResultType = s.ResultType
	a.EncodingFlags = s.EncodingFlags

	return nil
}

// QueryRangeData holds the result type and a list of stream or metric values of the query range response
type QueryRangeData struct {
	ResultType    QueryResultType `json:"resultType"`
	Stream        []StreamValues  `json:"stream,omitempty"`
	Matrix        []MatrixValues  `json:"matrix,omitempty"`
	EncodingFlags []string        `json:"encodingFlags,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *QueryRangeData) UnmarshalJSON(b []byte) error {
	var s rawQueryData
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s.ResultType {
	case ResultTypeStreams:
		if err := json.Unmarshal(s.Result, &a.Stream); err != nil {
			return err
		}
	case ResultTypeMatrix:
		if err := json.Unmarshal(s.Result, &a.Matrix); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown query range result type %s", s.ResultType)
	}
	a.ResultType = s.ResultType
	a.EncodingFlags = s.EncodingFlags

	return nil
}

// queryResponse holds the status and data of the instant query request
type queryResponse struct {
	Status string    `json:"status"`
	Data   QueryData `json:"data"`
}

// queryRangeResponse holds the status and data of the range query request
type queryRangeResponse struct {
	Status string         `json:"status"`
	Data   QueryRangeData `json:"data"`
}

type formatQueryResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

// FormatQuery [formats a LogQL query].
//
// [formats a LogQL query]: https://grafana.com/docs/loki/latest/reference/loki-http-api/#format-a-logql-query
func (c *Client) FormatQuery(ctx context.Context, query string) (string, error) {
	if query == "" {
		return "", errQueryRequired
	}
	req, span, cancel, err := c.createRequestSpan(ctx, http.MethodGet, "/loki/api/v1/format_query", nil)
	if err != nil {
		return "", err
	}
	defer cancel()

	span.SetAttributes(attribute.String("db.query.text", query))

	q := req.URL.Query()
	q.Set("query", query)
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req, span)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result formatQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.SetStatus(codes.Error, "failed to decode json response")
		span.RecordError(err)

		return "", schema.InternalServerError(fmt.Sprintf("failed to decode json response: %s", err), nil)
	}

	return result.Data, nil
}
