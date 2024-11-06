package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-sdk-go/schema"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Client is a HTTP client to request Loki API resources
type Client struct {
	httpClient   *http.Client
	baseURL      string
	timeout      uint
	maxTimeRange time.Duration

	// pre-calculated host and port of the Loki server URL
	serverAddress string
	serverPort    int

	tracer *connector.Tracer
}

// New creates a new Loki client
func New(cfg ClientSettings) (*Client, error) {
	baseURL, err := cfg.URL.Get()
	if err != nil {
		return nil, fmt.Errorf("invalid Loki URL: %w", err)
	}
	if baseURL == "" {
		return nil, errEndpointRequired
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Loki URL: %w", err)
	}

	rt := &roundTripper{
		next:       http.DefaultTransport,
		headers:    map[string]string{},
		propagator: otel.GetTextMapPropagator(),
	}

	for key, header := range cfg.Headers {
		value, err := header.Get()
		if err != nil {
			return nil, fmt.Errorf("invalid value in header %s: %w", key, err)
		}
		if value != "" {
			rt.headers[key] = value
		}
	}

	c := &Client{
		httpClient: &http.Client{
			Transport: rt,
		},
		baseURL:       baseURL,
		timeout:       cfg.Timeout,
		maxTimeRange:  time.Duration(cfg.MaxTimeRange),
		serverAddress: u.Hostname(),
		tracer:        connector.NewTracer("LokiClient"),
	}

	rawPort := u.Port()
	switch {
	case rawPort != "":

		serverPort, err := strconv.ParseInt(rawPort, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid port in the connection URL: %w", err)
		}
		c.serverPort = int(serverPort)
	case u.Scheme == "https":
		c.serverPort = 443
	default:
		c.serverPort = 80
	}

	return c, nil
}

func (c *Client) createRequestSpan(ctx context.Context, method string, apiPath string, body io.Reader) (*http.Request, trace.Span, context.CancelFunc, error) {
	ctx, span := c.tracer.Start(ctx, fmt.Sprintf("%s %s", method, apiPath), trace.WithSpanKind(trace.SpanKindClient))

	span.SetAttributes(
		attribute.String("db.system", "loki"),
		attribute.String("server.address", c.serverAddress),
		attribute.String("http.request.method", method),
		attribute.String("server.address", c.serverAddress),
		attribute.Int("server.port", c.serverPort),
	)

	apiEndpoint := c.baseURL + apiPath
	cancel := noopCancel
	if c.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(c.timeout)*time.Second)
	}

	cancelFunc := func() {
		span.End()
		cancel()
	}
	req, err := http.NewRequestWithContext(ctx, method, apiEndpoint, body)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create request")
		span.RecordError(err)
		span.End()
		cancelFunc()

		return nil, nil, nil, schema.InternalServerError("failed to create request: "+err.Error(), nil)
	}

	return req, span, cancelFunc, nil
}

// set semantic attributes of the http request to the span
func (c *Client) setHTTPRequestAttributes(span trace.Span, req *http.Request) {
	span.SetAttributes(attribute.String("url.full", req.URL.String()))
	if req.ContentLength > 0 {
		span.SetAttributes(attribute.Int64("http.request.body.size", req.ContentLength))
	}
	for key, values := range req.Header {
		if key != XScopeOrgID {
			span.SetAttributes(attribute.StringSlice("http.request.header."+key, values))
		}
	}
}

// set semantic attributes of the http response to the span
func (c *Client) setHTTPResponseAttributes(span trace.Span, res *http.Response) {
	span.SetAttributes(
		attribute.Int("http.response.status_code", res.StatusCode),
	)

	if res.ContentLength > 0 {
		span.SetAttributes(attribute.Int64("http.response.body.size", res.ContentLength))
	}
	for key, values := range res.Header {
		span.SetAttributes(attribute.StringSlice("http.response.header."+key, values))
	}
}

func (c *Client) do(req *http.Request, span trace.Span) (*http.Response, error) {
	c.setHTTPRequestAttributes(span, req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		span.SetStatus(codes.Error, "failed to execute request")
		span.RecordError(err)

		return nil, schema.InternalServerError(err.Error(), nil)
	}

	c.setHTTPResponseAttributes(span, res)
	if res.StatusCode >= http.StatusBadRequest {
		defer res.Body.Close()
		buf, err := io.ReadAll(res.Body)
		if err != nil {
			span.SetStatus(codes.Error, "reading request failed with")
			span.RecordError(err)

			return nil, schema.InternalServerError(fmt.Sprintf("reading request failed with status code %v: %s", res.StatusCode, err), nil)
		}
		err = errors.New(string(buf))
		span.SetStatus(codes.Error, "request failed")
		span.RecordError(err)

		statusCode := res.StatusCode
		if res.StatusCode == http.StatusBadRequest {
			statusCode = http.StatusUnprocessableEntity
		}

		return res, schema.NewConnectorError(statusCode, fmt.Sprintf("request failed: %s", err), nil)
	}

	return res, nil
}

func (c *Client) doEmptyResponse(req *http.Request, span trace.Span) error {
	res, err := c.do(req, span)
	if err != nil {
		return err
	}
	if res.Body != nil {
		res.Body.Close()
	}

	return nil
}

type roundTripper struct {
	headers    map[string]string
	propagator propagation.TextMapPropagator
	next       http.RoundTripper
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range r.headers {
		req.Header.Add(key, value)
	}
	r.propagator.Inject(req.Context(), propagation.HeaderCarrier(req.Header))

	return r.next.RoundTrip(req)
}
