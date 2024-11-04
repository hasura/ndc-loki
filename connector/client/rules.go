package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hasura/ndc-sdk-go/schema"
	"go.opentelemetry.io/otel/codes"
)

// RulesResponse the response data of Loki rules
type RulesResponse struct {
	Status string    `json:"status"`
	Data   RulesData `json:"data"`
}

// RulesData the data of rules
type RulesData struct {
	Groups []Rules `json:"groups"`
}

// Rules the rule item
type Rules struct {
	Name  string        `json:"name"`
	File  string        `json:"file"`
	Rules []interface{} `json:"rules"`
}

// GetRules returns the loki ruler rules
func (c *Client) GetRules(ctx context.Context) (*RulesResponse, error) {
	req, span, cancel, err := c.createRequestSpan(ctx, http.MethodGet, "/prometheus/api/v1/rules", nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	resp, err := c.do(req, span)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result RulesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.SetStatus(codes.Error, "failed to decode json response")
		span.RecordError(err)

		return nil, schema.InternalServerError(fmt.Sprintf("failed to decode json response: %s", err), nil)
	}

	return &result, nil
}
