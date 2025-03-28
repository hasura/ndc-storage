package fs

import (
	"context"
	"os"
	"slices"

	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/spf13/afero"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = connector.NewTracer("connector/storage/fs")

// Client represents a file system client wrapper.
type Client struct {
	client             afero.Fs
	clientType         string
	allowedDirectories []string
	permissions        FilePermissionConfig
}

var _ common.StorageClient = &Client{}

// New creates a new generic filesystem client.
func New(client afero.Fs, config *ClientConfig) (*Client, error) {
	defaultDirectory, err := config.DefaultDirectory.GetOrDefault("")
	if err != nil {
		return nil, err
	}

	mc := &Client{
		clientType:         string(config.Type),
		client:             client,
		allowedDirectories: config.AllowedDirectories,
		permissions:        defaultFilePermissions,
	}

	if defaultDirectory != "" && !slices.Contains(mc.allowedDirectories, defaultDirectory) {
		mc.allowedDirectories = append(mc.allowedDirectories, defaultDirectory)
	}

	slices.Sort(mc.allowedDirectories)

	if config.Permissions != nil {
		mc.permissions = *config.Permissions
	}

	return mc, nil
}

// NewOSFileSystem creates a new OS file system client.
func NewOSFileSystem(config *ClientConfig) (*Client, error) {
	return New(afero.NewOsFs(), config)
}

func (c *Client) startOtelSpan(ctx context.Context, name string, bucketName string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		common.NewDBSystemAttribute(),
		attribute.String("rpc.system", c.clientType),
	)

	if bucketName != "" {
		span.SetAttributes(attribute.String("storage.bucket", bucketName))
	}

	return ctx, span
}

func (c *Client) lstatIfPossible(name string) (os.FileInfo, error) {
	return lstatIfPossible(c.client, name)
}

func lstatIfPossible(client afero.Fs, name string) (os.FileInfo, error) {
	if lstater, ok := client.(afero.Lstater); ok {
		result, _, err := lstater.LstatIfPossible(name)

		return result, err
	}

	return client.Stat(name)
}
