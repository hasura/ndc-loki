package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PushLogLineInput input arguments of the pushLogLines function.
type PushLogLineInput struct {
	Streams []StreamInput `json:"streams"`
}

// StreamInput represents a stream input object.
type StreamInput struct {
	Stream map[string]string `json:"stream,omitempty"`
	Values []LogLineInput    `json:"values"`
}

// LogLineInput represents a log line item
type LogLineInput struct {
	Line               string            `json:"line"`
	Timestamp          *time.Time        `json:"timestamp,omitempty"`
	StructuredMetadata map[string]string `json:"structured_metadata,omitempty"`
}

// MarshalJSON marshals the type into valid JSON.
func (lli LogLineInput) MarshalJSON() ([]byte, error) {
	var result []any
	ts := time.Now()
	if lli.Timestamp != nil {
		ts = *lli.Timestamp
	}
	if len(lli.StructuredMetadata) > 0 {
		result = []any{FormatUnixNanoTimestamp(ts), lli.Line, lli.StructuredMetadata}
	} else {
		result = []any{FormatUnixNanoTimestamp(ts), lli.Line}
	}

	return json.Marshal(result)
}

// PushLogLines [send log entries] to Loki.
//
// [send log entries]: https://grafana.com/docs/loki/latest/reference/loki-http-api/#ingest-logs
func (c *Client) PushLogLines(ctx context.Context, params *PushLogLineInput) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(params); err != nil {
		return fmt.Errorf("failed to marshal the log line input: %w", err)
	}

	req, span, cancel, err := c.createRequestSpan(ctx, http.MethodPost, "/loki/api/v1/push", &buf)
	if err != nil {
		return err
	}
	defer cancel()

	req.Header.Set("Content-Type", "application/json")

	return c.doEmptyResponse(req, span)
}
