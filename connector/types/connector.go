package types

import (
	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-storage/connector/storage"
)

// State is the global state which is shared for every connector request.
type State struct {
	*connector.TelemetryState
	Storage *storage.Manager
}
