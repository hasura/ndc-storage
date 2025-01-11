package collection

import (
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
)

const (
	CollectionStorageObjects             = "storageObjects"
	StorageObjectName                    = "StorageObject"
	StorageObjectColumnClientID          = "clientId"
	StorageObjectColumnObject            = "object"
	StorageObjectColumnName              = "name"
	StorageObjectColumnBucket            = "bucket"
	StorageObjectColumnLastModified      = "lastModified"
	StorageObjectColumnMetadata          = "metadata"
	StorageObjectColumnUserMetadata      = "userMetadata"
	StorageObjectColumnVersionID         = "versionId"
	StorageObjectColumnChecksumCrc32     = "checksumCrc32"
	StorageObjectColumnChecksumCrc32C    = "checksumCrc32C"
	StorageObjectColumnChecksumCrc64Nvme = "checksumCrc64Nvme"
	StorageObjectColumnChecksumSha1      = "checksumSha1"
	StorageObjectColumnChecksumSha256    = "checksumSha256"
	StorageObjectColumnContentType       = "contentType"
	StorageObjectColumnETag              = "etag"
	StorageObjectColumnExpiration        = "expiration"
	StorageObjectColumnExpirationRuleID  = "expirationRuleId"
	StorageObjectColumnExpires           = "expires"
	StorageObjectColumnSize              = "size"
	StorageObjectColumnStorageClass      = "storageClass"
	StorageObjectColumnUserTagCount      = "userTagCount"
	StorageObjectColumnIsDeleteMarker    = "isDeleteMarker"
	StorageObjectColumnIsLatest          = "isLatest"
	StorageObjectColumnReplicationReady  = "replicationReady"
	StorageObjectColumnReplicationStatus = "replicationStatus"

	StorageObjectArgumentWhere     = "where"
	StorageObjectArgumentRecursive = "recursive"
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
	ScalarBucketName      = "BucketName"
	ScalarObjectPath      = "ObjectPath"
	ScalarFilterTimestamp = "FilterTimestamp"
)

// GetConnectorSchema returns connector schema for object collections.
func GetConnectorSchema(clientIDs []string) *schema.SchemaResponse { //nolint:funlen
	return &schema.SchemaResponse{
		Collections: []schema.CollectionInfo{
			{
				Name:        CollectionStorageObjects,
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
			"StorageObjectSimple": schema.ObjectType{
				Fields: schema.ObjectTypeFields{
					StorageObjectColumnClientID: schema.ObjectField{
						Type: schema.NewNamedType(ScalarStorageClientID).Encode(),
					},
					StorageObjectColumnBucket: schema.ObjectField{
						Type: schema.NewNamedType(ScalarBucketName).Encode(),
					},
					StorageObjectColumnObject: schema.ObjectField{
						Type: schema.NewNamedType(ScalarObjectPath).Encode(),
					},
				},
			},
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
					StorageObjectColumnChecksumCrc32: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					StorageObjectColumnChecksumCrc32C: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					StorageObjectColumnChecksumCrc64Nvme: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					StorageObjectColumnChecksumSha1: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					StorageObjectColumnChecksumSha256: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					StorageObjectColumnContentType: schema.ObjectField{
						Type: schema.NewNamedType("String").Encode(),
					},
					StorageObjectColumnETag: schema.ObjectField{
						Type: schema.NewNamedType("String").Encode(),
					},
					StorageObjectColumnExpiration: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("TimestampTZ")).Encode(),
					},
					StorageObjectColumnExpirationRuleID: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					StorageObjectColumnExpires: schema.ObjectField{
						Type: schema.NewNamedType("TimestampTZ").Encode(),
					},
					"grant": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewArrayType(schema.NewNamedType("StorageGrant"))).Encode(),
					},
					StorageObjectColumnIsDeleteMarker: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("Boolean")).Encode(),
					},
					StorageObjectColumnIsLatest: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("Boolean")).Encode(),
					},
					StorageObjectColumnMetadata: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("JSON")).Encode(),
					},
					"owner": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("StorageOwner")).Encode(),
					},
					StorageObjectColumnReplicationReady: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("Boolean")).Encode(),
					},
					StorageObjectColumnReplicationStatus: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					"restore": schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("StorageRestoreInfo")).Encode(),
					},
					StorageObjectColumnSize: schema.ObjectField{
						Type: schema.NewNamedType("Int64").Encode(),
					},
					StorageObjectColumnStorageClass: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
					StorageObjectColumnUserMetadata: schema.ObjectField{
						Type: schema.NewNullableType(schema.NewNamedType("JSON")).Encode(),
					},
					StorageObjectColumnUserTagCount: schema.ObjectField{
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
					OperatorEqual:               schema.NewComparisonOperatorEqual().Encode(),
					OperatorStartsWith:          schema.NewComparisonOperatorCustom(schema.NewNamedType(ScalarObjectPath)).Encode(),
					OperatorContains:            schema.NewComparisonOperatorCustom(schema.NewNamedType(ScalarObjectPath)).Encode(),
					OperatorInsensitiveContains: schema.NewComparisonOperatorCustom(schema.NewNamedType(ScalarObjectPath)).Encode(),
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
