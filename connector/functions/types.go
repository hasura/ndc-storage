package functions

import (
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/functions/internal"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

var GetBaseConnectorSchema = internal.GetConnectorSchema

var errPermissionDenied = schema.ForbiddenError("permission dennied", nil)

// StorageConnectionEdge the connection information of the relay pagination response.
type StorageConnection[T any] struct {
	Edges    []StorageConnectionEdge[T]   `json:"edges"`
	PageInfo common.StoragePaginationInfo `json:"pageInfo"`
}

// StorageConnectionEdge the connection edge of the relay pagination response.
type StorageConnectionEdge[T any] struct {
	Node   T      `json:"node"`
	Cursor string `json:"cursor"`
}
