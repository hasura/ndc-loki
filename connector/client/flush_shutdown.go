package client

import (
	"context"
	"net/http"
)

// Flush all in-memory chunks held by the ingesters to the backing store
//
// [all in-memory chunks]: https://grafana.com/docs/loki/latest/reference/loki-http-api/#flush-in-memory-chunks-to-backing-store
func (c *Client) Flush(ctx context.Context) error {
	req, span, cancel, err := c.createRequestSpan(ctx, http.MethodPost, "/flush", nil)
	if err != nil {
		return err
	}
	defer cancel()

	req.Header.Set("Content-Type", "application/json")

	return c.doEmptyResponse(req, span)
}
