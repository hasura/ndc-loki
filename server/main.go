package main

import (
	loki "github.com/hasura/ndc-loki/connector"
	"github.com/hasura/ndc-sdk-go/connector"
)

// Start the connector server at http://localhost:8080
//
//	go run . serve
//
// See [NDC Go SDK] for more information.
//
// [NDC Go SDK]: https://github.com/hasura/ndc-sdk-go
func main() {
	if err := connector.Start(
		loki.NewLokiConnector(),
		connector.WithMetricsPrefix("ndc_loki"),
		connector.WithDefaultServiceName("ndc_loki"),
	); err != nil {
		panic(err)
	}
}
