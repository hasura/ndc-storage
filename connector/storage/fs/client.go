package fs

import (
	"context"
	"log/slog"
	"slices"

	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = connector.NewTracer("connector/storage/fs")

// Client represents a file system client wrapper.
type Client struct {
	defaultDirectory   string
	allowedDirectories []string
	permissions        FilePermissionConfig
}

var _ common.StorageClient = &Client{}

// New creates a new Minio client.
func New(ctx context.Context, config *ClientConfig, logger *slog.Logger) (*Client, error) {
	mc := &Client{
		defaultDirectory:   config.DefaultDirectory,
		allowedDirectories: config.AllowedDirectories,
		permissions:        defaultFilePermissions,
	}

	if !slices.Contains(mc.allowedDirectories, mc.defaultDirectory) {
		mc.allowedDirectories = append(mc.allowedDirectories, mc.defaultDirectory)
	}

	slices.Sort(mc.allowedDirectories)

	if config.Permissions != nil {
		mc.permissions = *config.Permissions
	}

	return mc, nil
}

func (c *Client) startOtelSpan(ctx context.Context, name string, bucketName string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		common.NewDBSystemAttribute(),
		attribute.String("rpc.system", "fs"),
	)

	if bucketName != "" {
		span.SetAttributes(attribute.String("storage.bucket", bucketName))
	}

	return ctx, span
}
