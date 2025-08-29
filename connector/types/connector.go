package types

import (
	"context"

	"github.com/hasura/ndc-sdk-go/v2/connector"
	"github.com/hasura/ndc-storage/connector/storage"
)

type ContextKey string

const (
	// QueryVariablesContextKey the context key for the current query variables.
	QueryVariablesContextKey ContextKey = "ndc-query-variables"
)

// State is the global state which is shared for every connector request.
type State struct {
	*connector.TelemetryState
	Storage     *storage.Manager
	Concurrency ConcurrencySettings
}

// QueryVariablesFromContext gets the query variables from context.
func QueryVariablesFromContext(ctx context.Context) map[string]any {
	value := ctx.Value(QueryVariablesContextKey)
	if value != nil {
		if selection, ok := value.(map[string]any); ok {
			return selection
		}
	}

	return nil
}
