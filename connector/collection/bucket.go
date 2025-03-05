package collection

import (
	"context"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

// CollectionBucketExecutor evaluates and executes a storage bucket collection query.
type CollectionBucketExecutor struct {
	Storage     *storage.Manager
	Request     *schema.QueryRequest
	Arguments   map[string]any
	Variables   map[string]any
	Concurrency int
}

// Execute executes the query request to get list of storage buckets.
func (coe *CollectionBucketExecutor) Execute(ctx context.Context) (*schema.RowSet, error) {
	if coe.Request.Query.Offset != nil && *coe.Request.Query.Offset < 0 {
		return nil, schema.UnprocessableContentError("offset must be positive", nil)
	}

	if coe.Request.Query.Limit != nil && *coe.Request.Query.Limit <= 0 {
		return &schema.RowSet{
			Aggregates: schema.RowSetAggregates{},
			Rows:       []map[string]any{},
		}, nil
	}

	request, err := EvalBucketPredicate(common.StorageClientCredentialArguments{}, nil, coe.Request.Query.Predicate, coe.Variables)
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

	if err := request.EvalArguments(coe.Arguments); err != nil {
		return nil, err
	}

	request.evalQuerySelectionFields(coe.Request.Query.Fields)

	predicate := request.BucketPredicate.CheckPostPredicate
	if !request.BucketPredicate.HasPostPredicate() {
		predicate = nil
	}

	options := &common.ListStorageBucketsOptions{
		Prefix: request.BucketPredicate.GetPrefix(),
		Include: common.BucketIncludeOptions{
			Tags:       request.Include.Tags,
			Versioning: request.Include.Versions,
			Lifecycle:  request.Include.Lifecycle,
			Encryption: request.Include.Encryption,
			ObjectLock: request.IncludeObjectLock,
		},
		NumThreads: coe.Concurrency,
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
		maxResults := offset + limit

		options.MaxResults = &maxResults
	}

	response, err := coe.Storage.ListBuckets(ctx, request.GetBucketArguments().StorageClientCredentialArguments, options, predicate)
	if err != nil {
		return nil, err
	}

	buckets := response.Buckets

	if offset > 0 {
		if len(buckets) <= offset {
			return &schema.RowSet{
				Aggregates: schema.RowSetAggregates{},
				Rows:       []map[string]any{},
			}, nil
		}

		buckets = buckets[offset:]
	}

	rawResults := make([]map[string]any, len(buckets))
	for i, object := range buckets {
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
