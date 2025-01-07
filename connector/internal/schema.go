package internal

import (
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
)

const (
	CollectionStorageObject         = "storageObjects"
	StorageObjectName               = "StorageObject"
	StorageObjectColumnClientID     = "clientId"
	StorageObjectColumnName         = "name"
	StorageObjectColumnBucket       = "bucket"
	StorageObjectColumnLastModified = "lastModified"
	StorageObjectColumnMetadata     = "metadata"
	StorageObjectColumnUserMetadata = "userMetadata"
	StorageObjectColumnVersionID    = "versionId"
	StorageObjectArgumentRecursive  = "recursive"
)

const (
	OperatorEqual      = "_eq"
	OperatorStartsWith = "_starts_with"
	OperatorGreater    = "_gt"
	OperatorIsNull     = "_is_null"
)

const (
	ScalarStorageClientID = "StorageClientID"
	ScalarBucketName      = "BucketName"
	ScalarObjectPath      = "ObjectPath"
	ScalarFilterTimestamp = "FilterTimestamp"
)

// GetConnectorSchema returns connector schema for object collections.
func GetConnectorSchema(clientIDs []string) *schema.SchemaResponse { //nolint:funlen
	return &schema.SchemaResponse{
		Collections: []schema.CollectionInfo{
			{
				Name:        CollectionStorageObject,
				Description: utils.ToPtr("The information of an storage object"),
				Type:        StorageObjectName,
				Arguments: schema.CollectionInfoArguments{
					StorageObjectArgumentRecursive: schema.ArgumentInfo{
						Type: schema.NewNullableType(schema.NewNamedType("Boolean")).Encode(),
					},
				},
				UniquenessConstraints: schema.CollectionInfoUniquenessConstraints{},
				ForeignKeys:           schema.CollectionInfoForeignKeys{},
			},
		},
		ObjectTypes: schema.SchemaResponseObjectTypes{
			StorageObjectName: schema.ObjectType{
				Fields: schema.ObjectTypeFields{
					StorageObjectColumnClientID: schema.ObjectField{
						Type: schema.NewNamedType(ScalarStorageClientID).Encode(),
					},
					StorageObjectColumnBucket: schema.ObjectField{
						Type: schema.NewNamedType(ScalarBucketName).Encode(),
					},
					StorageObjectColumnName: schema.ObjectField{
						Type: schema.NewNamedType(ScalarObjectPath).Encode(),
					},
					StorageObjectColumnLastModified: schema.ObjectField{
						Type: schema.NewNamedType(ScalarFilterTimestamp).Encode(),
					},
					"checksumCrc32": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					"checksumCrc32C": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					"checksumCrc64Nvme": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					"checksumSha1": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					"checksumSha256": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					"contentType": schema.ObjectField{
						Type: schema.NewNamedType("String").Encode(),
					},
					"etag": schema.ObjectField{
						Type: schema.NewNamedType("String").Encode(),
					},
					"expiration": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("TimestampTZ")).Encode(),
					},
					"expirationRuleId": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					"expires": schema.ObjectField{
						Type: schema.NewNamedType("TimestampTZ").Encode(),
					},
					"grant": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewArrayType(schema.NewNamedType("StorageGrant"))).Encode(),
					},
					"isDeleteMarker": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("Boolean")).Encode(),
					},
					"isLatest": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("Boolean")).Encode(),
					},
					StorageObjectColumnMetadata: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("JSON")).Encode(),
					},
					"owner": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("StorageOwner")).Encode(),
					},
					"replicationReady": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("Boolean")).Encode(),
					},
					"replicationStatus": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					"restore": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("StorageRestoreInfo")).Encode(),
					},
					"size": schema.ObjectField{
						Type: schema.NewNamedType("Int64").Encode(),
					},
					"storageClass": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					StorageObjectColumnUserMetadata: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("JSON")).Encode(),
					},
					"userTagCount": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("Int32")).Encode(),
					},
					"userTags": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("JSON")).Encode(),
					},
					StorageObjectColumnVersionID: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
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
			ScalarObjectPath: schema.ScalarType{
				AggregateFunctions: schema.ScalarTypeAggregateFunctions{},
				ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{
					OperatorStartsWith: schema.NewComparisonOperatorCustom(schema.NewNamedType(ScalarObjectPath)).Encode(),
				},
				Representation: schema.NewTypeRepresentationString().Encode(),
			},
			ScalarFilterTimestamp: schema.ScalarType{
				AggregateFunctions: schema.ScalarTypeAggregateFunctions{},
				ComparisonOperators: map[string]schema.ComparisonOperatorDefinition{
					OperatorGreater: schema.NewComparisonOperatorCustom(schema.NewNamedType("TimestampTZ")).Encode(),
				},
				Representation: schema.NewTypeRepresentationTimestampTZ().Encode(),
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
