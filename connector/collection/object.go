package collection

import (
	"context"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

// CollectionObjectExecutor executes the query to get the list of collection objects.
type CollectionObjectExecutor struct {
	Storage   *storage.Manager
	Request   *schema.QueryRequest
	Arguments map[string]any
	Variables map[string]any
}

// GetMany executes the query request to get list of storage objects.
func (coe *CollectionObjectExecutor) Execute(ctx context.Context) (*schema.RowSet, error) {
	request, err := EvalCollectionObjectsRequest(coe.Request, coe.Arguments, coe.Variables)
	if err != nil {
		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	if !request.IsValid {
		// early returns zero rows
		// the evaluated query always returns empty values
		return &schema.RowSet{
			Aggregates: schema.RowSetAggregates{},
			Rows:       []map[string]any{},
		}, nil
	}

	objects, err := coe.Storage.ListObjects(ctx, request.StorageBucketArguments, &request.Options)
	if err != nil {
		return nil, err
	}

	var filtered []common.StorageObject

	if request.HasPostPredicate() {
		for _, item := range objects {
			if request.CheckPostObjectPredicate(item) {
				filtered = append(filtered, item)
			}
		}
	} else {
		filtered = objects
	}

	rawResults := make([]map[string]any, len(filtered))
	for i, object := range filtered {
		rawResults[i] = object.ToMap()
	}

	result, err := utils.EvalObjectsWithColumnSelection(coe.Request.Query.Fields, rawResults)
	if err != nil {
		return nil, err
	}

	return &schema.RowSet{
		Aggregates: schema.RowSetAggregates{},
		Rows:       result,
	}, nil
}
