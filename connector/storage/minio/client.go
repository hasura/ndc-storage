package minio

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = connector.NewTracer("connector/storage/minio")

// Client represents a Minio client wrapper.
type Client struct {
	publicHost   *url.URL
	providerType common.StorageProviderType
	isDebug      bool
	client       *minio.Client
}

var _ common.StorageClient = &Client{}

// New creates a new Minio client.
func New(
	ctx context.Context,
	providerType common.StorageProviderType,
	cfg *ClientConfig,
	logger *slog.Logger,
) (*Client, error) {
	publicHost, err := cfg.ValidatePublicHost()
	if err != nil {
		return nil, err
	}

	mc := &Client{
		publicHost:   publicHost,
		providerType: providerType,
		isDebug:      utils.IsDebug(logger),
	}

	opts, endpoint, err := cfg.toMinioOptions(providerType, logger)
	if err != nil {
		return nil, err
	}

	c, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the minio client: %w", err)
	}

	mc.client = c

	return mc, nil
}

func (mc *Client) startOtelSpan(
	ctx context.Context,
	name string,
	bucketName string,
) (context.Context, trace.Span) {
	spanKind := trace.SpanKindClient
	if mc.isDebug {
		spanKind = trace.SpanKindInternal
	}

	return mc.startOtelSpanWithKind(ctx, spanKind, name, bucketName)
}

func (mc *Client) startOtelSpanWithKind(
	ctx context.Context,
	spanKind trace.SpanKind,
	name string,
	bucketName string,
) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, name, trace.WithSpanKind(spanKind))
	span.SetAttributes(
		common.NewDBSystemAttribute(),
		attribute.String("rpc.system", string(mc.providerType)),
	)

	if bucketName != "" {
		span.SetAttributes(attribute.String("storage.bucket", bucketName))
	}

	return ctx, span
}
