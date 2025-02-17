package collection

import (
	"context"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

// CollectionObjectExecutor evaluates and executes a storage object collection query
type CollectionObjectExecutor struct {
	Storage     *storage.Manager
	Request     *schema.QueryRequest
	Arguments   map[string]any
	Variables   map[string]any
	Concurrency int
}

// Execute executes the query request to get list of storage objects.
func (coe *CollectionObjectExecutor) Execute(ctx context.Context) (*schema.RowSet, error) {
	if coe.Request.Query.Offset != nil && *coe.Request.Query.Offset < 0 {
		return nil, schema.UnprocessableContentError("offset must be positive", nil)
	}

	if coe.Request.Query.Limit != nil && *coe.Request.Query.Limit <= 0 {
		return &schema.RowSet{
			Aggregates: schema.RowSetAggregates{},
			Rows:       []map[string]any{},
		}, nil
	}

	request, err := EvalObjectPredicate(common.StorageBucketArguments{}, nil, coe.Request.Query.Predicate, coe.Variables)
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

	request.evalQuerySelectionFields(coe.Request.Query.Fields)

	options := &common.ListStorageObjectsOptions{
		Prefix:     request.ObjectNamePredicate.GetPrefix(),
		Include:    request.Include,
		NumThreads: coe.Concurrency,
	}

	if hierarchy, err := utils.GetNullableBoolean(coe.Arguments, argumentHierarchy); err != nil {
		return nil, schema.UnprocessableContentError(err.Error(), nil)
	} else if hierarchy != nil {
		options.Hierarchy = *hierarchy
	}

	if after, err := utils.GetNullableString(coe.Arguments, argumentAfter); err != nil {
		return nil, schema.UnprocessableContentError(err.Error(), nil)
	} else if after != nil && *after != "" {
		options.StartAfter = *after
	}

	var offset, limit int
	if coe.Request.Query.Offset != nil {
		offset = *coe.Request.Query.Offset
	}

	if coe.Request.Query.Limit != nil {
		limit = *coe.Request.Query.Limit
		options.MaxResults = offset + limit
	}

	predicate := request.ObjectNamePredicate.CheckPostPredicate

	if !request.ObjectNamePredicate.HasPostPredicate() {
		predicate = nil
	}

	response, err := coe.Storage.ListObjects(ctx, request.GetBucketArguments(), options, predicate)
	if err != nil {
		return nil, err
	}

	objects := response.Objects

	if offset > 0 {
		if len(response.Objects) <= offset {
			return &schema.RowSet{
				Aggregates: schema.RowSetAggregates{},
				Rows:       []map[string]any{},
			}, nil
		}

		objects = response.Objects[offset:]
	}

	rawResults := make([]map[string]any, len(objects))
	for i, object := range objects {
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
