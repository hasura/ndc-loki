package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hasura/ndc-sdk-go/schema"
	"go.opentelemetry.io/otel/codes"
)

type seriesResponse struct {
	Data []map[string]string `json:"data"`
}

// Series return a series query result
func (c *Client) Series(ctx context.Context, matcher string) ([]map[string]string, error) {
	req, span, cancel, err := c.createRequestSpan(ctx, http.MethodGet, "/loki/api/v1/series", nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	q := req.URL.Query()
	q.Set("match[]", matcher)
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req, span)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result seriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.SetStatus(codes.Error, "failed to decode json response")
		span.RecordError(err)

		return nil, schema.InternalServerError(fmt.Sprintf("failed to decode json response: %s", err), nil)
	}

	return result.Data, nil
}
