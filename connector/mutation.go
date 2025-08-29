package connector

import (
	"context"
	"fmt"

	"github.com/hasura/ndc-sdk-go/v2/schema"
	"github.com/hasura/ndc-storage/connector/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"
)

// MutationExplain explains a mutation by creating an execution plan.
func (c *Connector) MutationExplain(
	ctx context.Context,
	configuration *types.Configuration,
	state *types.State,
	request *schema.MutationRequest,
) (*schema.ExplainResponse, error) {
	return nil, schema.NotSupportedError("mutation explain has not been supported yet", nil)
}

// Mutation executes a mutation.
func (c *Connector) Mutation(
	ctx context.Context,
	configuration *types.Configuration,
	state *types.State,
	request *schema.MutationRequest,
) (*schema.MutationResponse, error) {
	concurrencyLimit := c.config.Concurrency.Mutation
	if len(request.Operations) <= 1 || concurrencyLimit <= 1 {
		return c.execMutationSync(ctx, state, request)
	}

	return c.execMutationAsync(ctx, state, request)
}

func (c *Connector) execMutationSync(
	ctx context.Context,
	state *types.State,
	request *schema.MutationRequest,
) (*schema.MutationResponse, error) {
	operationResults := make([]schema.MutationOperationResults, len(request.Operations))

	for i, operation := range request.Operations {
		result, err := c.execMutation(ctx, state, operation, i)
		if err != nil {
			return nil, err
		}

		operationResults[i] = result
	}

	return &schema.MutationResponse{
		OperationResults: operationResults,
	}, nil
}

func (c *Connector) execMutationAsync(
	ctx context.Context,
	state *types.State,
	request *schema.MutationRequest,
) (*schema.MutationResponse, error) {
	operationResults := make([]schema.MutationOperationResults, len(request.Operations))
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(c.config.Concurrency.Mutation)

	for i, operation := range request.Operations {
		func(index int, op schema.MutationOperation) {
			eg.Go(func() error {
				result, err := c.execMutation(ctx, state, op, index)
				if err != nil {
					return err
				}

				operationResults[index] = result

				return nil
			})
		}(i, operation)
	}

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	return &schema.MutationResponse{
		OperationResults: operationResults,
	}, nil
}

func (c *Connector) execMutation(
	ctx context.Context,
	state *types.State,
	operation schema.MutationOperation,
	index int,
) (schema.MutationOperationResults, error) {
	ctx, span := state.Tracer.Start(ctx, fmt.Sprintf("Execute Procedure %d", index))
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", string(operation.Type)),
		attribute.String("operation.name", operation.Name),
	)

	switch operation.Type {
	case schema.MutationOperationProcedure:
		result, err := c.execProcedure(ctx, state, &operation)
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("failed to execute procedure %d", index))
			span.RecordError(err)

			return nil, err
		}

		return result, nil
	default:
		errorMsg := fmt.Sprintf("invalid operation type: %s", operation.Type)
		span.SetStatus(codes.Error, errorMsg)

		return nil, schema.UnprocessableContentError(errorMsg, nil)
	}
}

func (c *Connector) execProcedure(
	ctx context.Context,
	state *types.State,
	operation *schema.MutationOperation,
) (schema.MutationOperationResults, error) {
	return c.apiHandler.Mutation(ctx, state, operation)
}
