package common

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var transportTracer = connector.NewTracer("connector/storage/common/transport")

// NewTransport creates a new HTTP transporter for the storage client.
func NewTransport(logger *slog.Logger, port int, secure bool) http.RoundTripper {
	transport, _ := minio.DefaultTransport(secure)

	if utils.IsDebug(logger) {
		return debugRoundTripper{
			transport:  transport,
			propagator: otel.GetTextMapPropagator(),
			port:       port,
			logger:     logger,
		}
	}

	return roundTripper{
		transport:  transport,
		propagator: otel.GetTextMapPropagator(),
	}
}

type debugRoundTripper struct {
	transport  *http.Transport
	propagator propagation.TextMapPropagator
	port       int
	logger     *slog.Logger
}

func (mrt debugRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, span := transportTracer.Start(req.Context(), fmt.Sprintf("%s %s", req.Method, req.URL.Path), trace.WithSpanKind(trace.SpanKindClient))
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

		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

		span.SetAttributes(attribute.Int("http.response.size", len(respBody)))
	}

	slog.Debug(resp.Status, logAttrs...)

	if resp.StatusCode >= http.StatusBadRequest {
		span.SetStatus(codes.Error, resp.Status)
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
