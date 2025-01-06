package minio

import (
	"context"

	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/minio/minio-go/v7/pkg/sse"
	"go.opentelemetry.io/otel/codes"
)

// SetBucketEncryption sets default encryption configuration on a bucket.
func (mc *Client) SetBucketEncryption(ctx context.Context, bucketName string, input common.ServerSideEncryptionConfiguration) error {
	ctx, span := mc.startOtelSpan(ctx, "SetBucketEncryption", bucketName)
	defer span.End()

	err := mc.client.SetBucketEncryption(ctx, bucketName, validateBucketEncryptionConfiguration(input))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// GetBucketEncryption gets default encryption configuration set on a bucket.
func (mc *Client) GetBucketEncryption(ctx context.Context, bucketName string) (*common.ServerSideEncryptionConfiguration, error) {
	ctx, span := mc.startOtelSpan(ctx, "GetBucketEncryption", bucketName)
	defer span.End()

	rawResult, err := mc.client.GetBucketEncryption(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, serializeErrorResponse(err)
	}

	return serializeBucketEncryptionConfiguration(rawResult), nil
}

// RemoveBucketEncryption remove default encryption configuration set on a bucket.
func (mc *Client) RemoveBucketEncryption(ctx context.Context, bucketName string) error {
	ctx, span := mc.startOtelSpan(ctx, "RemoveBucketEncryption", bucketName)
	defer span.End()

	err := mc.client.RemoveBucketEncryption(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

func validateBucketEncryptionConfiguration(input common.ServerSideEncryptionConfiguration) *sse.Configuration {
	result := &sse.Configuration{
		Rules: make([]sse.Rule, len(input.Rules)),
	}

	for i, rule := range input.Rules {
		r := sse.Rule{
			Apply: sse.ApplySSEByDefault{
				SSEAlgorithm: rule.Apply.SSEAlgorithm,
			},
		}

		if rule.Apply.KmsMasterKeyID != nil {
			r.Apply.KmsMasterKeyID = *rule.Apply.KmsMasterKeyID
		}

		result.Rules[i] = r
	}

	return result
}

func serializeBucketEncryptionConfiguration(input *sse.Configuration) *common.ServerSideEncryptionConfiguration {
	result := &common.ServerSideEncryptionConfiguration{
		Rules: make([]common.ServerSideEncryptionRule, len(input.Rules)),
	}

	for i, rule := range input.Rules {
		r := common.ServerSideEncryptionRule{
			Apply: common.StorageApplySSEByDefault{
				SSEAlgorithm: rule.Apply.SSEAlgorithm,
			},
		}

		if !isStringNull(rule.Apply.KmsMasterKeyID) {
			r.Apply.KmsMasterKeyID = &rule.Apply.KmsMasterKeyID
		}

		result.Rules[i] = r
	}

	return result
}
