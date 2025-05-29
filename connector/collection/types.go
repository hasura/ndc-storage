package collection

import (
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
)

const (
	CollectionStorageObjects    = "storage_objects"
	CollectionStorageBuckets    = "storage_buckets"
	StorageObjectName           = "StorageObject"
	StorageBucketName           = "StorageBucket"
	StorageObjectColumnClientID = "client_id"
	StorageObjectColumnBucket   = "bucket"
	StorageObjectColumnName     = "name"
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
	argumentAfter           = "after"
	argumentRecursive       = "recursive"
	ArgumentClientType      = "client_type"
	ArgumentEndpoint        = "endpoint"
	ArgumentAccessKeyID     = "access_key_id"
	ArgumentSecretAccessKey = "secret_access_key"
)

var checksumColumnNames = []string{"checksum_crc32", "checksum_crc32c", "checksum_crc64_nvme", "checksum_sha1", "checksum_sha256"}

// StringComparisonOperator represents the explicit comparison expression for string columns.
type StringComparisonOperator struct {
	Value    string
	Operator string
}

// GetConnectorSchema returns connector schema for object collections.
func GetConnectorSchema(clientIDs []string, dynamicCredentials bool) *schema.SchemaResponse {
	storageObjectArguments := buildDynamicCredentialArguments(schema.CollectionInfoArguments{
		argumentAfter: {
			Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
		},
		argumentRecursive: {
			Type: schema.NewNullableType(schema.NewNamedType("Boolean")).Encode(),
		},
	}, dynamicCredentials)

	if dynamicCredentials {
		storageObjectArguments[StorageObjectColumnBucket] = schema.ArgumentInfo{
			Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
		}
	}

	return &schema.SchemaResponse{
		Collections: []schema.CollectionInfo{
			{
				Name:                  CollectionStorageObjects,
				Description:           utils.ToPtr("List storage objects"),
				Type:                  StorageObjectName,
				Arguments:             storageObjectArguments,
				UniquenessConstraints: schema.CollectionInfoUniquenessConstraints{},
			},
			{
				Name:        CollectionStorageBuckets,
				Description: utils.ToPtr("List storage buckets"),
				Type:        StorageBucketName,
				Arguments: buildDynamicCredentialArguments(schema.CollectionInfoArguments{
					argumentAfter: {
						Type: schema.NewNullableType(schema.NewNamedType("String")).Encode(),
					},
				}, dynamicCredentials),
				UniquenessConstraints: schema.CollectionInfoUniquenessConstraints{},
			},
		},
		ObjectTypes: schema.SchemaResponseObjectTypes{
			"StorageBucketFilter": schema.ObjectType{
				Fields: schema.ObjectTypeFields{
					StorageObjectColumnClientID: schema.ObjectField{
						Type: schema.NewNamedType(ScalarStorageClientID).Encode(),
					},
					StorageObjectColumnName: schema.ObjectField{
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
					StorageObjectColumnName: schema.ObjectField{
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

func buildDynamicCredentialArguments(arguments map[string]schema.ArgumentInfo, dynamicCredentials bool) map[string]schema.ArgumentInfo {
	if !dynamicCredentials {
		return arguments
	}

	results := map[string]schema.ArgumentInfo{
		ArgumentClientType: {
			Description: utils.ToPtr("The cloud storage provider type"),
			Type:        schema.NewNullableType(schema.NewNamedType("StorageProviderType")).Encode(),
		},
		ArgumentEndpoint: {
			Description: utils.ToPtr("Endpoint of the cloud storage service"),
			Type:        schema.NewNullableType(schema.NewNamedType("String")).Encode(),
		},
		ArgumentAccessKeyID: {
			Description: utils.ToPtr("Access key ID or Account name credential"),
			Type:        schema.NewNullableType(schema.NewNamedType("String")).Encode(),
		},
		ArgumentSecretAccessKey: {
			Description: utils.ToPtr("Secret Access key ID or Account key credential"),
			Type:        schema.NewNullableType(schema.NewNamedType("String")).Encode(),
		},
	}

	for key, arg := range arguments {
		results[key] = arg
	}

	return results
}
