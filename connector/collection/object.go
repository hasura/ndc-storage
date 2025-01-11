package collection

import (
	"context"
	"slices"
	"strings"

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
	if coe.Request.Query.Limit != nil && *coe.Request.Query.Limit == 0 {
		return &schema.RowSet{
			Aggregates: schema.RowSetAggregates{},
			Rows:       []map[string]any{},
		}, nil
	}

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

	if len(request.OrderBy) > 0 {
		filtered = coe.sortObjects(filtered, request.OrderBy)
	}

	if coe.Request.Query.Offset != nil && *coe.Request.Query.Offset > 0 {
		if *coe.Request.Query.Offset >= len(filtered) {
			return &schema.RowSet{
				Aggregates: schema.RowSetAggregates{},
				Rows:       []map[string]any{},
			}, nil
		}

		filtered = filtered[*coe.Request.Query.Offset:]
	}

	if coe.Request.Query.Limit != nil {
		limit := len(filtered)
		if *coe.Request.Query.Limit < limit {
			limit = *coe.Request.Query.Limit
		}

		filtered = filtered[:limit]
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

func (coe *CollectionObjectExecutor) sortObjects(objects []common.StorageObject, orderBys []ColumnOrder) []common.StorageObject { //nolint:funlen,gocognit,gocyclo
	slices.SortFunc(objects, func(a, b common.StorageObject) int {
		for _, ob := range orderBys {
			ordering := 1
			if ob.Descending {
				ordering = -1
			}

			switch ob.Name {
			case StorageObjectColumnName:
				if a.Name == b.Name {
					continue
				}

				return strings.Compare(a.Name, b.Name) * ordering
			case StorageObjectColumnLastModified:
				if a.LastModified.Equal(b.LastModified) {
					continue
				}

				return int(a.LastModified.Sub(b.LastModified)) * ordering
			case StorageObjectColumnChecksumCrc32:
				if cmpResult := compareNullableString(a.ChecksumCRC32, b.ChecksumCRC32); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnChecksumCrc32C:
				if cmpResult := compareNullableString(a.ChecksumCRC32C, b.ChecksumCRC32C); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnChecksumCrc64Nvme:
				if cmpResult := compareNullableString(a.ChecksumCRC64NVME, b.ChecksumCRC64NVME); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnChecksumSha1:
				if cmpResult := compareNullableString(a.ChecksumSHA1, b.ChecksumSHA1); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnChecksumSha256:
				if cmpResult := compareNullableString(a.ChecksumSHA256, b.ChecksumSHA256); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnETag:
				if cmpResult := compareNullableString(a.ETag, b.ETag); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnStorageClass:
				if cmpResult := compareNullableString(a.StorageClass, b.StorageClass); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnContentType:
				if cmpResult := compareNullableString(a.ContentType, b.ContentType); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnVersionID:
				if cmpResult := compareNullableString(a.VersionID, b.VersionID); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnExpirationRuleID:
				if cmpResult := compareNullableString(a.ExpirationRuleID, b.ExpirationRuleID); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnReplicationStatus:
				if cmpResult := compareNullableString((*string)(a.ReplicationStatus), (*string)(b.ReplicationStatus)); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnSize:
				if a.Size == b.Size {
					continue
				}

				return int(a.Size-b.Size) * ordering
			case StorageObjectColumnUserTagCount:
				if a.UserTagCount == b.UserTagCount {
					continue
				}

				return (a.UserTagCount - b.UserTagCount) * ordering
			case StorageObjectColumnExpires:
				if cmpResult := compareNullableTime(a.Expires, b.Expires); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnExpiration:
				if cmpResult := compareNullableTime(a.Expiration, b.Expiration); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnIsDeleteMarker:
				if cmpResult := compareNullableBoolean(a.IsDeleteMarker, b.IsDeleteMarker); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnIsLatest:
				if cmpResult := compareNullableBoolean(a.IsLatest, b.IsLatest); cmpResult != 0 {
					return cmpResult * ordering
				}
			case StorageObjectColumnReplicationReady:
				if cmpResult := compareNullableBoolean(a.ReplicationReady, b.ReplicationReady); cmpResult != 0 {
					return cmpResult * ordering
				}
			}
		}

		return 0
	})

	return objects
}
