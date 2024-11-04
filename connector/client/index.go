package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hasura/ndc-sdk-go/schema"
	"go.opentelemetry.io/otel/codes"
)

type statsResponse struct {
	Data []map[string]int `json:"data"`
}

func (c *Client) Stats(ctx context.Context, query string) ([]map[string]int, error) {
	req, span, cancel, err := c.createRequestSpan(ctx, http.MethodGet, "/loki/api/v1/index/stats", nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	q := req.URL.Query()
	q.Set("query", query)
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req, span)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result statsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.SetStatus(codes.Error, "failed to decode json response")
		span.RecordError(err)

		return nil, schema.InternalServerError(fmt.Sprintf("failed to decode json response: %s", err), nil)
	}

	return result.Data, nil
}
