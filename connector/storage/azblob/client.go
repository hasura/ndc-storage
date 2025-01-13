package azblob

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/storage/azblob/examples_test.go

type ClientConfig struct {
	AccountName string
}

// Client prepresents a Minio client wrapper.
type Client struct {
	client *azblob.Client
}

var _ common.StorageClient = &Client{}

// New creates a new Minio client.
func New(ctx context.Context, cfg *ClientConfig, logger *slog.Logger) (*Client, error) {
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", cfg.AccountName)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	client, err := azblob.NewClient(serviceURL, cred, &azblob.ClientOptions{})
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

func (c *Client) startOtelSpan(ctx context.Context, name string, bucketName string) (context.Context, trace.Span) {
	spanKind := trace.SpanKindClient

	ctx, span := tracer.Start(ctx, name, trace.WithSpanKind(spanKind))
	span.SetAttributes(
		common.NewDBSystemAttribute(),
		attribute.String("rpc.system", "azblob"),
	)

	if bucketName != "" {
		span.SetAttributes(attribute.String("storage.bucket", bucketName))
	}

	return ctx, span
}