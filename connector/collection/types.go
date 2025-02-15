package collection

import (
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
)

const (
	CollectionStorageObjects    = "storageObjects"
	CollectionStorageBuckets    = "storageBuckets"
	StorageObjectName           = "StorageObject"
	StorageBucketName           = "StorageBucket"
	StorageObjectColumnClientID = "clientId"
	StorageObjectColumnObject   = "object"
	StorageObjectColumnBucket   = "bucket"
)

const (
	OperatorEqual               = "_eq"
	OperatorStartsWith          = "_starts_with"
	OperatorContains            = "_contains"
	OperatorInsensitiveContains = "_icontains"
	OperatorGreater             = "_gt"
	OperatorIsNull              = "_is_null"
)

const (
	ScalarStorageClientID = "StorageClientID"
	ScalarBucketName      = "StorageBucketName"
	ScalarStringFilter    = "StorageStringFilter"
)

const (
	argumentAfter     = "after"
	argumentHierarchy = "hierarchy"
)

var checksumColumnNames = []string{"checksumCrc32", "checksumCrc32C", "checksumCrc64Nvme", "checksumSha1", "checksumSha256"}

// StringComparisonOperator represents the explicit comparison expression for string columns.
type StringComparisonOperator struct {
	Value    string
	Operator string
}

// GetConnectorSchema returns connector schema for object collections.
func GetConnectorSchema(clientIDs []string) *schema.SchemaResponse {
	return &schema.SchemaResponse{
		Collections: []schema.CollectionInfo{
			{
				Name:        CollectionStorageObjects,
				Description: utils.ToPtr("List storage objects"),
				Type:        StorageObjectName,
				Arguments: schema.CollectionInfoArguments{
					argumentAfter: {
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					argumentHierarchy: {
						Type: schema.NewNullableType(schema.NewNamedType("Boolean")).Encode(),
					},
				},
				UniquenessConstraints: schema.CollectionInfoUniquenessConstraints{},
				ForeignKeys:           schema.CollectionInfoForeignKeys{},
			},
			{
				Name:        CollectionStorageBuckets,
				Description: utils.ToPtr("List storage buckets"),
				Type:        StorageBucketName,
				Arguments: schema.CollectionInfoArguments{
					argumentAfter: {
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
				},
				UniquenessConstraints: schema.CollectionInfoUniquenessConstraints{},
				ForeignKeys:           schema.CollectionInfoForeignKeys{},
			},
		},
		ObjectTypes: schema.SchemaResponseObjectTypes{
			"StorageBucketFilter": schema.ObjectType{
				Fields: schema.ObjectTypeFields{
					StorageObjectColumnClientID: schema.ObjectField{
						Type: schema.NewNamedType(ScalarStorageClientID).Encode(),
					},
					StorageObjectColumnBucket: schema.ObjectField{
						Type: schema.NewNamedType(ScalarStringFilter).Encode(),
					},
				},
			},
			"StorageObjectFilter": schema.ObjectType{
				Fields: schema.ObjectTypeFields{
					StorageObjectColumnClientID: schema.ObjectField{
						Type: schema.NewNamedType(ScalarStorageClientID).Encode(),
					},
					StorageObjectColumnBucket: schema.ObjectField{
						Type: schema.NewNamedType(ScalarBucketName).Encode(),
					},
					StorageObjectColumnObject: schema.ObjectField{
						Type: schema.NewNamedType(ScalarStringFilter).Encode(),
					},
				},
			},
		},
		ScalarTypes: schema.SchemaResponseScalarTypes{
			ScalarBucketName: schema.ScalarType{
				AggregateFunctions: schema.ScalarTypeAggregateFunctions{},
				ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{
					OperatorEqual: schema.NewComparisonOperatorEqual().Encode(),
				},
				Representation: schema.NewTypeRepresentationString().Encode(),
			},
			ScalarStringFilter: schema.ScalarType{
				AggregateFunctions: schema.ScalarTypeAggregateFunctions{},
				ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{
					OperatorEqual:               schema.NewComparisonOperatorEqual().Encode(),
					OperatorStartsWith:          schema.NewComparisonOperatorCustom(schema.NewNamedType(ScalarStringFilter)).Encode(),
					OperatorContains:            schema.NewComparisonOperatorCustom(schema.NewNamedType(ScalarStringFilter)).Encode(),
					OperatorInsensitiveContains: schema.NewComparisonOperatorCustom(schema.NewNamedType(ScalarStringFilter)).Encode(),
				},
				Representation: schema.NewTypeRepresentationString().Encode(),
			},
			ScalarStorageClientID: schema.ScalarType{
				AggregateFunctions: schema.ScalarTypeAggregateFunctions{},
				ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{
					OperatorEqual: schema.NewComparisonOperatorEqual().Encode(),
				},
				Representation: schema.NewTypeRepresentationEnum(clientIDs).Encode(),
			},
		},
	}
}
