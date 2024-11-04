package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"go.opentelemetry.io/otel/codes"
)

// LabelsParams represent parameters of the labels request
type LabelsParams struct {
	// The start time for the query as a nanosecond Unix epoch. Defaults to 6 hours ago.
	Start *time.Time `json:"start"`
	// The end time for the query as a nanosecond Unix epoch. Defaults to now.
	End *time.Time `json:"end"`
	// A duration used to calculate start relative to end.
	// If end is in the future, start is calculated as this duration before now.
	// Any value specified for start supersedes this parameter.
	Since *scalar.Duration `json:"since"`
	// Log stream selector that selects the streams to match and return label names. Example: {app="myapp", environment="dev"}
	Query string `json:"query,omitempty"`
}

// ApplyQueryParams apply values to query parameters
func (lp LabelsParams) ApplyQueryParams(q url.Values) url.Values {
	if lp.Query != "" {
		q.Set("query", lp.Query)
	}
	if lp.Start != nil {
		q.Set("start", FormatUnixTimestamp(*lp.Start))
	}
	if lp.End != nil {
		q.Set("end", FormatUnixTimestamp(*lp.End))
	}
	if lp.Since != nil && lp.Since.Duration > 0 {
		q.Set("since", lp.Since.String())
	}

	return q
}

type labelsResponse struct {
	Data []string `json:"data"`
}

// Labels [retrieve the list of known labels] within a given time span
//
// [retrieve the list of known labels]: https://grafana.com/docs/loki/latest/reference/loki-http-api/#query-labels
func (c *Client) Labels(ctx context.Context, params *LabelsParams) ([]string, error) {
	req, span, cancel, err := c.createRequestSpan(ctx, http.MethodGet, "/loki/api/v1/labels", nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	q := params.ApplyQueryParams(req.URL.Query())
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req, span)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result labelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.SetStatus(codes.Error, "failed to decode json response")
		span.RecordError(err)

		return nil, schema.InternalServerError(fmt.Sprintf("failed to decode json response: %s", err), nil)
	}

	return result.Data, nil
}

// LabelValuesParams represent parameters of the label values request
type LabelValuesParams struct {
	Name string `json:"name"`

	LabelsParams
}

// LabelValues retrieve the list of known values for a given label within a given time span.
//
// [list of known values for a given label]: https://grafana.com/docs/loki/latest/reference/loki-http-api/#query-label-values
func (c *Client) LabelValues(ctx context.Context, params *LabelValuesParams) ([]string, error) {
	if params.Name == "" {
		return nil, errLabelNameRequired
	}

	req, span, cancel, err := c.createRequestSpan(ctx, http.MethodGet, fmt.Sprintf("/loki/api/v1/label/%s/values", url.PathEscape(params.Name)), nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	q := params.ApplyQueryParams(req.URL.Query())
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req, span)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result labelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.SetStatus(codes.Error, "failed to decode json response")
		span.RecordError(err)

		return nil, schema.InternalServerError(fmt.Sprintf("failed to decode json response: %s", err), nil)
	}

	return result.Data, nil
}
