package api

import (
	"context"

	"github.com/hasura/ndc-loki/connector/client"
	"github.com/hasura/ndc-loki/connector/metadata"
)

// FunctionLokiQuery allows for doing [queries against a single point in time]
func FunctionLokiQuery(ctx context.Context, state *metadata.State, arguments *client.QueryParams) (client.QueryData, error) {
	resp, err := state.Client.Query(ctx, arguments)
	if err != nil {
		return client.QueryData{}, err
	}

	return *resp, nil
}

// QueryRange queries logs within a range of time. This type of query is often referred to as a range query.
func FunctionLokiQueryRange(ctx context.Context, state *metadata.State, arguments *client.QueryRangeParams) (client.QueryRangeData, error) {
	resp, err := state.Client.QueryRange(ctx, arguments)
	if err != nil {
		return client.QueryRangeData{}, err
	}

	return *resp, nil
}

// FunctionLokiLabels return the list of known labels within a given time span.
func FunctionLokiLabels(ctx context.Context, state *metadata.State, arguments *client.LabelsParams) ([]string, error) {
	return state.Client.Labels(ctx, arguments)
}

// FunctionLokiLabelValues retrieve the list of known values for a given label within a given time span.
func FunctionLokiLabelValues(ctx context.Context, state *metadata.State, arguments *client.LabelValuesParams) ([]string, error) {
	return state.Client.LabelValues(ctx, arguments)
}
