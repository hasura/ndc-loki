package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/hasura/ndc-sdk-go/schema"
	"go.opentelemetry.io/otel/codes"
)

const apiLogDeletionPath = "/loki/api/v1/delete"

// CreateLogDeletionRequestParams input parameters to create a log deletion request
type CreateLogDeletionRequestParams struct {
	Query string     `json:"query"`
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

// LogDeletionRequest the log deletion request item
type LogDeletionRequest struct {
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	Query     string `json:"query"`
	Status    string `json:"status"`
}

// CreateLogDeletionRequest [creates a new log deletion request] for the authenticated tenant
//
// [create a new log deletion request]: https://grafana.com/docs/loki/latest/reference/loki-http-api/#request-log-deletion
func (c *Client) CreateLogDeletionRequest(ctx context.Context, params CreateLogDeletionRequestParams) error {
	req, span, cancel, err := c.createRequestSpan(ctx, http.MethodPost, apiLogDeletionPath, nil)
	if err != nil {
		return err
	}
	defer cancel()

	q := req.URL.Query()
	q.Add("query", params.Query)
	if params.Start != nil {
		q.Add("start", params.Start.Format(time.RFC3339))
	}
	if params.End != nil {
		q.Add("end", params.End.Format(time.RFC3339))
	}
	req.URL.RawQuery = q.Encode()

	return c.doEmptyResponse(req, span)
}

// GetLogDeletionRequests [list the existing delete requests] for the authenticated tenant
//
// [list the existing delete requests]: https://grafana.com/docs/loki/latest/reference/loki-http-api/#list-log-deletion-requests
func (c *Client) GetLogDeletionRequests(ctx context.Context) ([]LogDeletionRequest, error) {
	req, span, cancel, err := c.createRequestSpan(ctx, http.MethodGet, apiLogDeletionPath, nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	resp, err := c.do(req, span)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var results []LogDeletionRequest
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		span.SetStatus(codes.Error, "parsing json output failed")
		span.RecordError(err)

		return nil, schema.InternalServerError("parsing json output failed: "+err.Error(), nil)
	}

	return results, nil
}

// CancelLogDeletionRequest [cancels a new log deletion request] for the authenticated tenant
//
// [cancels a new log deletion request]: https://grafana.com/docs/loki/latest/reference/loki-http-api/#request-cancellation-of-a-delete-request
func (c *Client) CancelLogDeletionRequest(ctx context.Context, requestID string, force *bool) error {
	req, span, cancel, err := c.createRequestSpan(ctx, http.MethodDelete, apiLogDeletionPath, nil)
	if err != nil {
		return err
	}
	defer cancel()

	q := req.URL.Query()
	q.Add("request_id", requestID)
	if force != nil {
		q.Add("force", strconv.FormatBool(*force))
	}
	req.URL.RawQuery = q.Encode()

	return c.doEmptyResponse(req, span)
}
