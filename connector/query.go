package connector

import (
	"context"
	"fmt"

	"github.com/hasura/ndc-loki/connector/internal"
	"github.com/hasura/ndc-loki/connector/metadata"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"
)

// QueryExplain explains a query by creating an execution plan.
func (c *LokiConnector) QueryExplain(ctx context.Context, conf *metadata.Configuration, state *metadata.State, request *schema.QueryRequest) (*schema.ExplainResponse, error) {
	if c.apiHandler.QueryExists(request.Collection) {
		return &schema.ExplainResponse{
			Details: schema.ExplainResponseDetails{},
		}, nil
	}

	if c.apiHandler.QueryExists(request.Collection) {
		return &schema.ExplainResponse{
			Details: schema.ExplainResponseDetails{},
		}, nil
	}

	requestVars := request.Variables
	if len(requestVars) == 0 {
		requestVars = []schema.QueryRequestVariablesElem{make(schema.QueryRequestVariablesElem)}
	}

	arguments, err := utils.ResolveArgumentVariables(request.Arguments, requestVars[0])
	if err != nil {
		return nil, err
	}

	if nativeQuery, ok := c.metadata.NativeOperations.Queries[request.Collection]; ok {
		executor := &internal.NativeQueryExecutor{
			Tracer:      state.Tracer,
			Client:      state.Client,
			Request:     request,
			NativeQuery: &nativeQuery,
			Arguments:   arguments,
			Runtime:     c.runtime,
		}
		_, queryString, err := executor.Explain(ctx)
		if err != nil {
			return nil, err
		}

		return &schema.ExplainResponse{
			Details: schema.ExplainResponseDetails{
				"query": queryString,
			},
		}, nil
	}

	collection, queryType := c.metadata.GetModel(request.Collection)
	if collection != nil {
		executor := &internal.QueryCollectionExecutor{
			Tracer:    state.Tracer,
			Client:    state.Client,
			Request:   request,
			Model:     *collection,
			Variables: requestVars[0],
			Arguments: arguments,
			Runtime:   c.runtime,
			QueryType: queryType,
		}

		_, queryString, _, err := executor.Explain(ctx)
		if err != nil {
			return nil, err
		}

		return &schema.ExplainResponse{
			Details: schema.ExplainResponseDetails{
				"query": queryString,
			},
		}, nil
	}

	return nil, fmt.Errorf("%s: unsupported query to explain", request.Collection)
}

// Query executes a query.
func (c *LokiConnector) Query(ctx context.Context, configuration *metadata.Configuration, state *metadata.State, request *schema.QueryRequest) (schema.QueryResponse, error) {
	requestVars := request.Variables
	if len(requestVars) == 0 {
		requestVars = []schema.QueryRequestVariablesElem{make(schema.QueryRequestVariablesElem)}
	}

	if len(requestVars) == 1 || c.runtime.QueryConcurrencyLimit <= 1 {
		return c.execQuerySync(ctx, state, request, requestVars)
	}

	return c.execQueryAsync(ctx, state, request, requestVars)
}

func (c *LokiConnector) execQuerySync(ctx context.Context, state *metadata.State, request *schema.QueryRequest, requestVars []schema.QueryRequestVariablesElem) ([]schema.RowSet, error) {
	rowSets := make([]schema.RowSet, len(requestVars))

	for i, requestVar := range requestVars {
		result, err := c.execQuery(ctx, state, request, requestVar, i)
		if err != nil {
			return nil, err
		}
		rowSets[i] = *result
	}

	return rowSets, nil
}

func (c *LokiConnector) execQueryAsync(ctx context.Context, state *metadata.State, request *schema.QueryRequest, requestVars []schema.QueryRequestVariablesElem) ([]schema.RowSet, error) {
	rowSets := make([]schema.RowSet, len(requestVars))

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(c.runtime.QueryConcurrencyLimit)

	for i, requestVar := range requestVars {
		func(index int, vars schema.QueryRequestVariablesElem) {
			eg.Go(func() error {
				result, err := c.execQuery(ctx, state, request, vars, index)
				if err != nil {
					return err
				}
				rowSets[index] = *result

				return nil
			})
		}(i, requestVar)
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return rowSets, nil
}

func (c *LokiConnector) execQuery(ctx context.Context, state *metadata.State, request *schema.QueryRequest, variables map[string]any, index int) (*schema.RowSet, error) {
	ctx, span := state.Tracer.Start(ctx, fmt.Sprintf("Execute Query %d", index))
	defer span.End()

	arguments, err := utils.ResolveArgumentVariables(request.Arguments, variables)
	if err != nil {
		errorMsg := "failed to resolve argument variables"
		span.SetStatus(codes.Error, errorMsg)
		span.RecordError(err)

		return nil, schema.UnprocessableContentError(errorMsg, map[string]any{
			"cause": err.Error(),
		})
	}
	span.SetAttributes(utils.JSONAttribute("arguments", arguments))

	if c.apiHandler.QueryExists(request.Collection) {
		result, err := c.apiHandler.Query(ctx, state, request, arguments)
		if err != nil {
			span.SetStatus(codes.Error, "failed to execute query")
			span.RecordError(err)

			return nil, err
		}

		return result, nil
	}

	// evaluate native query
	if nativeQuery, ok := c.metadata.NativeOperations.Queries[request.Collection]; ok {
		executor := &internal.NativeQueryExecutor{
			Tracer:      state.Tracer,
			Client:      state.Client,
			Runtime:     c.runtime,
			Request:     request,
			NativeQuery: &nativeQuery,
			Arguments:   arguments,
			Variables:   variables,
		}
		result, err := executor.Execute(ctx)
		if err != nil {
			span.SetStatus(codes.Error, "failed to execute the native query")
			span.RecordError(err)

			return nil, err
		}

		return result, nil
	}

	collection, queryType := c.metadata.GetModel(request.Collection)
	if collection != nil {
		if request.Query.Limit != nil && *request.Query.Limit <= 0 {
			return &schema.RowSet{
				Aggregates: schema.RowSetAggregates{},
				Rows:       []map[string]any{},
			}, nil
		}
		executor := &internal.QueryCollectionExecutor{
			Tracer:    state.Tracer,
			Client:    state.Client,
			Runtime:   c.runtime,
			Request:   request,
			Model:     *collection,
			Arguments: arguments,
			Variables: variables,
			QueryType: queryType,
		}

		result, err := executor.Execute(ctx)
		if err != nil {
			span.SetStatus(codes.Error, "failed to execute the collection query")
			span.RecordError(err)

			return nil, err
		}

		return result, nil
	}

	return nil, schema.UnprocessableContentError(fmt.Sprintf("invalid query `%s`", request.Collection), nil)
}
