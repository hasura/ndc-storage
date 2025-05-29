package common

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/hasura/ndc-http/exhttp"
	"github.com/hasura/ndc-sdk-go/connector"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var httpTracer = connector.NewTracer("connector/storage/common/http")

// HTTPTransportOptions hold optional options for the http transport.
type HTTPTransportOptions struct {
	Logger                     *slog.Logger
	Port                       int
	DisableCompression         bool
	DisableHighCardinalityPath bool
}

// NewTransport creates a new HTTP transporter for the storage client.
func NewTransport(transport *http.Transport, options HTTPTransportOptions) http.RoundTripper {
	if transport == nil {
		transport = exhttp.HTTPTransportConfig{}.ToTransport()
	}

	transport.DisableCompression = options.DisableCompression
	logger := options.Logger

	if options.Logger == nil {
		logger = slog.Default()
	}

	return roundTripper{
		transport:                  transport,
		propagator:                 otel.GetTextMapPropagator(),
		port:                       options.Port,
		disableHighCardinalityPath: options.DisableHighCardinalityPath,
		logger:                     logger,
	}
}

type roundTripper struct {
	transport                  *http.Transport
	propagator                 propagation.TextMapPropagator
	logger                     *slog.Logger
	port                       int
	disableHighCardinalityPath bool
}

func (mrt roundTripper) getRequestSpanName(req *http.Request) string {
	spanName := req.Method
	if !mrt.disableHighCardinalityPath {
		spanName += " " + req.URL.Path
	}

	return spanName
}

// RoundTrip wraps the base RoundTripper with telemetry.
func (mrt roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, span := httpTracer.Start(
		req.Context(),
		mrt.getRequestSpanName(req),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	port := mrt.port
	if port == 0 {
		port, _ = exhttp.ParsePort(req.URL.Port(), req.URL.Scheme)
	}

	span.SetAttributes(
		attribute.String("http.request.method", req.Method),
		attribute.String("url.full", req.URL.String()),
		attribute.String("server.address", req.URL.Hostname()),
		attribute.Int("server.port", port),
		attribute.String("network.protocol.name", "http"),
	)

	connector.SetSpanHeaderAttributes(span, "http.request.header.", req.Header)

	if req.ContentLength > 0 {
		span.SetAttributes(attribute.Int64("http.request.body.size", req.ContentLength))
	}

	isDebug := mrt.logger.Enabled(ctx, slog.LevelDebug)
	mrt.propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
	requestLogAttrs := map[string]any{
		"url":     req.URL.String(),
		"method":  req.Method,
		"headers": connector.NewTelemetryHeaders(req.Header),
	}

	if isDebug && req.Body != nil && req.ContentLength > 0 && req.ContentLength <= 100*1024 {
		rawBody, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		requestLogAttrs["body"] = string(rawBody)

		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewBuffer(rawBody))
	}

	logAttrs := []any{
		slog.String("type", "http-client"),
		slog.Any("request", requestLogAttrs),
	}

	resp, err := mrt.transport.RoundTrip(req)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		slog.Debug("failed to execute the request: "+err.Error(), logAttrs...)

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

	if isDebug && resp.ContentLength > 0 && resp.ContentLength < 1024*1024 && resp.Body != nil {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			logAttrs = append(logAttrs, slog.Any("response", respLogAttrs))

			slog.Debug("failed to read response body: "+err.Error(), logAttrs...)

			_ = resp.Body.Close()

			return resp, err
		}

		respLogAttrs["body"] = string(respBody)
		logAttrs = append(logAttrs, slog.Any("response", respLogAttrs))

		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

		span.SetAttributes(attribute.Int("http.response.size", len(respBody)))
	}

	slog.Debug(resp.Status, logAttrs...)

	if resp.StatusCode >= http.StatusBadRequest {
		span.SetStatus(codes.Error, resp.Status)
	}

	return resp, err
}
