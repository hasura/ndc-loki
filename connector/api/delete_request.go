package api

import (
	"context"

	"github.com/hasura/ndc-loki/connector/client"
	"github.com/hasura/ndc-loki/connector/metadata"
)

// FunctionLokiLogDeletionRequests list the existing delete requests for the authenticated tenant
func FunctionLokiLogDeletionRequests(ctx context.Context, state *metadata.State) ([]client.LogDeletionRequest, error) {
	return state.Client.GetLogDeletionRequests(ctx)
}

// ProcedureLokiCreateLogDeletionRequest creates a new log deletion request for the authenticated tenant
func ProcedureLokiCreateLogDeletionRequest(ctx context.Context, state *metadata.State, arguments *client.CreateLogDeletionRequestParams) (bool, error) {
	if err := state.Client.CreateLogDeletionRequest(ctx, *arguments); err != nil {
		return false, err
	}

	return true, nil
}

// LokiCancelLogDeletionRequestArguments request arguments of the addDeleteRequest mutation
type LokiCancelLogDeletionRequestArguments struct {
	RequestID string `json:"request_id"`
	Force     *bool  `json:"force"`
}

// ProcedureLokiCancelLogDeletionRequest cancels a new log deletion request for the authenticated tenant
func ProcedureLokiCancelLogDeletionRequest(ctx context.Context, state *metadata.State, arguments *LokiCancelLogDeletionRequestArguments) (bool, error) {
	if err := state.Client.CancelLogDeletionRequest(ctx, arguments.RequestID, arguments.Force); err != nil {
		return false, err
	}

	return true, nil
}
