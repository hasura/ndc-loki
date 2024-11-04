package api

import (
	"context"

	"github.com/hasura/ndc-loki/connector/client"
	"github.com/hasura/ndc-loki/connector/metadata"
)

// ProcedureLokiPushLogLines pushes log lines to Loki
func ProcedureLokiPushLogLines(ctx context.Context, state *metadata.State, arguments *client.PushLogLineInput) (bool, error) {
	err := state.Client.PushLogLines(ctx, arguments)
	if err != nil {
		return false, err
	}

	return true, nil
}
