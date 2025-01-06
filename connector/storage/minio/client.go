package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var tracer = connector.NewTracer("connector/storage/minio")

// AuthConfig represent the authentication config of the minio client.
type AuthConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	UseIAMAuth      bool
	IAMAuthEndpoint string
}

// ClientConfig represent the configuration of the minio client.
type ClientConfig struct {
	Type       common.StorageProviderType
	Endpoint   string
	Region     string
	PublicHost *url.URL
	Port       int
	MaxRetries int
	Secure     bool
	// TrailingHeaders indicates server support of trailing headers.
	// Only supported for v4 signatures.
	TrailingHeaders bool

	AuthConfig
}

// Client prepresents a Minio client wrapper.
type Client struct {
	publicHost   *url.URL
	providerType common.StorageProviderType
	isDebug      bool
	client       *minio.Client
}

var _ common.StorageClient = &Client{}

// New creates a new Minio client.
func New(ctx context.Context, cfg *ClientConfig, logger *slog.Logger) (*Client, error) {
	mc := &Client{
		publicHost:   cfg.PublicHost,
		providerType: cfg.Type,
		isDebug:      utils.IsDebug(logger),
	}

	transport, err := minio.DefaultTransport(cfg.Secure)
	if err != nil {
		return nil, err
	}

	opts := &minio.Options{
		Secure:          cfg.Secure,
		Transport:       transport,
		Region:          cfg.Region,
		MaxRetries:      cfg.MaxRetries,
		TrailingHeaders: cfg.TrailingHeaders,
	}

	if utils.IsDebug(logger) {
		opts.Transport = debugRoundTripper{
			transport:  transport,
			propagator: otel.GetTextMapPropagator(),
			port:       cfg.Port,
			logger:     logger,
		}
	} else {
		opts.Transport = roundTripper{
			transport:  transport,
			propagator: otel.GetTextMapPropagator(),
		}
	}

	if cfg.UseIAMAuth {
		opts.Creds = credentials.NewIAM(cfg.IAMAuthEndpoint)
	} else {
		opts.Creds = credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken)
	}

	c, err := minio.New(cfg.Endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the minio client: %w", err)
	}

	mc.client = c

	return mc, nil
}

type debugRoundTripper struct {
	transport  *http.Transport
	propagator propagation.TextMapPropagator
	port       int
	logger     *slog.Logger
}

func (mrt debugRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, span := tracer.Start(req.Context(), fmt.Sprintf("%s %s", req.Method, req.URL.Path), trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attribute.String("http.request.method", req.Method),
		attribute.String("url.full", req.URL.String()),
		attribute.String("server.address", req.URL.Hostname()),
		attribute.Int("server.port", mrt.port),
		attribute.String("network.protocol.name", "http"),
	)

	connector.SetSpanHeaderAttributes(span, "http.request.header.", req.Header)

	if req.ContentLength > 0 {
		span.SetAttributes(attribute.Int64("http.request.body.size", req.ContentLength))
	}

	mrt.propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	requestLogAttrs := map[string]any{
		"url":     req.URL.String(),
		"method":  req.Method,
		"headers": connector.NewTelemetryHeaders(req.Header),
	}

	if req.Body != nil && req.ContentLength > 0 && req.ContentLength <= 100*1024 {
		rawBody, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		requestLogAttrs["body"] = string(rawBody)

		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewBuffer(rawBody))
	}

	logAttrs := []any{
		slog.String("type", "storage-client"),
		slog.Any("request", requestLogAttrs),
	}

	resp, err := mrt.transport.RoundTrip(req)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Debug("failed to execute the request: %s"+err.Error(), logAttrs...)

		return resp, err
	}

	span.SetAttributes(attribute.Int("http.response.status_code", resp.StatusCode))
	connector.SetSpanHeaderAttributes(span, "http.response.header.", resp.Header)

	if resp.ContentLength >= 0 {
		span.SetAttributes(attribute.Int64("http.response.size", resp.ContentLength))
	}

	respLogAttrs := map[string]any{
		"http_status": resp.StatusCode,
		"headers":     resp.Header,
	}

	if resp.Body != nil {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			logAttrs = append(logAttrs, slog.Any("response", respLogAttrs))

			slog.Debug("failed to read response body: "+err.Error(), logAttrs...)
			resp.Body.Close()

			return resp, err
		}

		respLogAttrs["body"] = string(respBody)
		logAttrs = append(logAttrs, slog.Any("response", respLogAttrs))
		slog.Debug(resp.Status, logAttrs...)

		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

		span.SetAttributes(attribute.Int("http.response.size", len(respBody)))
	}

	if resp.StatusCode >= http.StatusBadRequest {
		span.SetStatus(codes.Error, resp.Status)
	} else {
		slog.Debug("executed request successfully", logAttrs...)
	}

	return resp, err
}

type roundTripper struct {
	transport  *http.Transport
	propagator propagation.TextMapPropagator
}

func (rt roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.propagator.Inject(req.Context(), propagation.HeaderCarrier(req.Header))

	return rt.transport.RoundTrip(req)
}

// EnableVersioning enables bucket versioning support.
func (mc *Client) EnableVersioning(ctx context.Context, bucketName string) error {
	ctx, span := mc.startOtelSpan(ctx, "EnableVersioning", bucketName)
	defer span.End()

	err := mc.client.EnableVersioning(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

// SuspendVersioning disables bucket versioning support.
func (mc *Client) SuspendVersioning(ctx context.Context, bucketName string) error {
	ctx, span := mc.startOtelSpan(ctx, "SuspendVersioning", bucketName)
	defer span.End()

	err := mc.client.SuspendVersioning(ctx, bucketName)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return serializeErrorResponse(err)
	}

	return nil
}

func (mc *Client) startOtelSpan(ctx context.Context, name string, bucketName string) (context.Context, trace.Span) {
	spanKind := trace.SpanKindClient
	if mc.isDebug {
		spanKind = trace.SpanKindInternal
	}

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
