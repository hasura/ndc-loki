package connector

import (
	"context"
	"fmt"

	"github.com/hasura/ndc-loki/connector/metadata"
	"github.com/hasura/ndc-sdk-go/schema"
	"golang.org/x/sync/errgroup"
)

// MutationExplain explains a mutation by creating an execution plan.
func (c *LokiConnector) MutationExplain(ctx context.Context, conf *metadata.Configuration, state *metadata.State, request *schema.MutationRequest) (*schema.ExplainResponse, error) {
	return nil, schema.NotSupportedError("mutation explain has not been supported yet", nil)
}

// Mutation executes a mutation.
func (c *LokiConnector) Mutation(ctx context.Context, configuration *metadata.Configuration, state *metadata.State, request *schema.MutationRequest) (*schema.MutationResponse, error) {
	if len(request.Operations) == 1 || c.runtime.MutationConcurrencyLimit <= 1 {
		return c.execMutationSync(ctx, state, request)
	}

	return c.execMutationAsync(ctx, state, request)
}

func (c *LokiConnector) execMutationSync(ctx context.Context, state *metadata.State, request *schema.MutationRequest) (*schema.MutationResponse, error) {
	operationResults := make([]schema.MutationOperationResults, len(request.Operations))
	for i, operation := range request.Operations {
		result, err := c.execMutationOperation(ctx, state, operation, i)
		if err != nil {
			return nil, err
		}
		operationResults[i] = result
	}

	return &schema.MutationResponse{
		OperationResults: operationResults,
	}, nil
}

func (c *LokiConnector) execMutationAsync(ctx context.Context, state *metadata.State, request *schema.MutationRequest) (*schema.MutationResponse, error) {
	operationResults := make([]schema.MutationOperationResults, len(request.Operations))

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(c.runtime.QueryConcurrencyLimit)

	for i, operation := range request.Operations {
		func(index int, op schema.MutationOperation) {
			eg.Go(func() error {
				result, err := c.execMutationOperation(ctx, state, op, index)
				if err != nil {
					return err
				}
				operationResults[index] = result

				return nil
			})
		}(i, operation)
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return &schema.MutationResponse{
		OperationResults: operationResults,
	}, nil
}

func (c *LokiConnector) execMutationOperation(parentCtx context.Context, state *metadata.State, operation schema.MutationOperation, index int) (schema.MutationOperationResults, error) {
	ctx, span := state.Tracer.Start(parentCtx, fmt.Sprintf("Execute Operation %d", index))
	defer span.End()

	switch operation.Type {
	case schema.MutationOperationProcedure:
		return c.apiHandler.Mutation(ctx, state, &operation)
	default:
		return nil, schema.BadRequestError("invalid operation type: "+string(operation.Type), nil)
	}
}
