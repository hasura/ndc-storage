package gcs

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"cloud.google.com/go/storage"
	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = connector.NewTracer("connector/storage/gcs")

// Client represents a Minio client wrapper.
type Client struct {
	publicHost *url.URL
	client     *storage.Client
	projectID  string
	isDebug    bool
}

var _ common.StorageClient = &Client{}

// New creates a new Minio client.
func New(ctx context.Context, config *ClientConfig, logger *slog.Logger, version string) (*Client, error) {
	publicHost, err := config.ValidatePublicHost()
	if err != nil {
		return nil, err
	}

	projectID, err := config.ProjectID.GetOrDefault("")
	if err != nil {
		return nil, fmt.Errorf("projectId: %w", err)
	}

	if projectID == "" && config.Authentication.Type == AuthTypeCredentials {
		return nil, errRequireProjectID
	}

	opts, err := config.toClientOptions(ctx, logger, version)
	if err != nil {
		return nil, err
	}

	mc := &Client{
		publicHost: publicHost,
		projectID:  projectID,
		isDebug:    utils.IsDebug(logger),
	}

	if config.UseGRPC {
		mc.client, err = storage.NewGRPCClient(ctx, opts...)
	} else {
		mc.client, err = storage.NewClient(ctx, opts...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize the Google Cloud Storage client: %w", err)
	}

	return mc, nil
}

func (c *Client) startOtelSpan(ctx context.Context, name string, bucketName string) (context.Context, trace.Span) {
	spanKind := trace.SpanKindClient
	if c.isDebug {
		spanKind = trace.SpanKindInternal
	}

	ctx, span := tracer.Start(ctx, name, trace.WithSpanKind(spanKind))
	span.SetAttributes(
		common.NewDBSystemAttribute(),
		attribute.String("rpc.system", "gcs"),
	)

	if bucketName != "" {
		span.SetAttributes(attribute.String("storage.bucket", bucketName))
	}

	return ctx, span
}
