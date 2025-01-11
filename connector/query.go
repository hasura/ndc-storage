package connector

import (
	"context"
	"fmt"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/collection"
	"github.com/hasura/ndc-storage/connector/types"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"
)

// QueryExplain explains a query by creating an execution plan.
func (c *Connector) QueryExplain(ctx context.Context, configuration *types.Configuration, state *types.State, request *schema.QueryRequest) (*schema.ExplainResponse, error) {
	return nil, schema.NotSupportedError("query explain has not been supported yet", nil)
}

// Query executes a query.
func (c *Connector) Query(ctx context.Context, configuration *types.Configuration, state *types.State, request *schema.QueryRequest) (schema.QueryResponse, error) {
	requestVars := request.Variables
	if len(requestVars) == 0 {
		requestVars = []schema.QueryRequestVariablesElem{make(schema.QueryRequestVariablesElem)}
	}

	concurrencyLimit := c.config.Concurrency.Query
	if concurrencyLimit <= 1 || len(request.Variables) <= 1 {
		return c.execQuerySync(ctx, state, request, requestVars)
	}

	return c.execQueryAsync(ctx, state, request, requestVars)
}

func (c *Connector) execQuerySync(ctx context.Context, state *types.State, req *schema.QueryRequest, requestVars []schema.QueryRequestVariablesElem) (schema.QueryResponse, error) {
	rowSets := make([]schema.RowSet, len(requestVars))

	for i, requestVar := range requestVars {
		result, err := c.execQuery(ctx, state, req, requestVar, i)
		if err != nil {
			return nil, err
		}

		rowSets[i] = *result
	}

	return rowSets, nil
}

func (c *Connector) execQueryAsync(ctx context.Context, state *types.State, request *schema.QueryRequest, requestVars []schema.QueryRequestVariablesElem) (schema.QueryResponse, error) {
	rowSets := make([]schema.RowSet, len(requestVars))
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(c.config.Concurrency.Query)

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

func (c *Connector) execQuery(ctx context.Context, state *types.State, request *schema.QueryRequest, variables map[string]any, index int) (*schema.RowSet, error) {
	ctx, span := state.Tracer.Start(ctx, fmt.Sprintf("Execute Query %d", index))
	defer span.End()

	rawArgs, err := utils.ResolveArgumentVariables(request.Arguments, variables)
	if err != nil {
		span.SetStatus(codes.Error, "failed to resolve argument variables")
		span.RecordError(err)

		return nil, schema.UnprocessableContentError("failed to resolve argument variables", map[string]any{
			"cause": err.Error(),
		})
	}

	if request.Collection == collection.CollectionStorageObjects {
		executor := collection.CollectionObjectExecutor{
			Storage:   state.Storage,
			Request:   request,
			Arguments: rawArgs,
		}

		return executor.Execute(ctx)
	}

	result, err := c.apiHandler.Query(context.WithValue(ctx, types.QueryVariablesContextKey, variables), state, request, rawArgs)
	if err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("failed to execute function %d", index))
		span.RecordError(err)

		return nil, err
	}

	return result, nil
}
