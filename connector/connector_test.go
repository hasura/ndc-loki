package connector

import (
	"log/slog"
	"testing"

	"github.com/hasura/ndc-sdk-go/ndctest"
)

func TestConnector(t *testing.T) {
	t.Setenv("CONNECTION_URL", "http://localhost:3131")
	slog.SetLogLoggerLevel(slog.LevelError)
	ndctest.TestConnector(t, NewLokiConnector(), ndctest.TestConnectorOptions{
		Configuration: "../tests/configuration",
		TestDataDir:   "testdata/01-setup",
	})

	ndctest.TestConnector(t, NewLokiConnector(), ndctest.TestConnectorOptions{
		Configuration: "../tests/configuration",
		TestDataDir:   "testdata/02-query",
	})
}
