package functions

import (
	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

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

// SuccessResponse represents a common successful response structure.
type SuccessResponse struct {
	Success bool `json:"success"`
}

// NewSuccessResponse creates a SuccessResponse instance with default success=true.
func NewSuccessResponse() SuccessResponse {
	return SuccessResponse{Success: true}
}

// ExistsResponse represents a common existing response structure.
type ExistsResponse struct {
	Exists bool `json:"exists"`
}

// DownloadStorageObjectResponse represents the object data response in base64-encode string format.
type DownloadStorageObjectResponse struct {
	Data *scalar.Bytes `json:"data"`
}

// DownloadStorageObjectTextResponse represents the object data response in string format.
type DownloadStorageObjectTextResponse struct {
	Data *string `json:"data"`
}

// PutStorageObjectArguments represents input arguments of the PutObject method.
type PutStorageObjectBase64Arguments struct {
	common.PutStorageObjectArguments

	Data scalar.Bytes `json:"data"`
}

// PutStorageObjectTextArguments represents input arguments of the PutStorageObjectText method.
type PutStorageObjectTextArguments struct {
	common.PutStorageObjectArguments

	Data string `json:"data"`
}
