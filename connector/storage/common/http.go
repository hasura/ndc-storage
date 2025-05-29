package common

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/hasura/ndc-http/exhttp"
	"github.com/hasura/ndc-sdk-go/schema"
)

// @enum GET,POST.
type DownloadHTTPMethod string

// RequestOptions hold HTTP request options.
type HTTPRequestOptions struct {
	URL      string              `json:"url"`
	Method   *DownloadHTTPMethod `json:"method"`
	Headers  []StorageKeyValue   `json:"headers,omitempty"`
	BodyText string              `json:"body_text,omitempty"`
}

// HTTPClient extends the native http.Client with custom configurations and methods.
type HTTPClient struct {
	client *http.Client
}

// NewTransport creates a new http transport from config.
func NewTransport(
	config *exhttp.HTTPTransportTLSConfig,
	telemetry exhttp.TelemetryConfig,
) (http.RoundTripper, error) {
	var httpTransport *http.Transport

	if config != nil {
		var err error

		httpTransport, err = config.ToTransport(telemetry.Logger)
		if err != nil {
			return nil, err
		}
	} else {
		httpTransport = exhttp.HTTPTransportConfig{}.ToTransport()
	}

	httpTransport.DisableCompression = true

	return exhttp.NewTelemetryTransport(httpTransport, telemetry), nil
}

// NewHTTPClient creates an HTTP client from an HTTP transport configuration.
func NewHTTPClient(
	config *exhttp.HTTPTransportTLSConfig,
	logger *slog.Logger,
) (*HTTPClient, error) {
	transport, err := NewTransport(config, exhttp.TelemetryConfig{
		Logger:                     logger,
		DisableHighCardinalityPath: true,
	})
	if err != nil {
		return nil, err
	}

	return &HTTPClient{
		client: &http.Client{
			Transport: transport,
		},
	}, nil
}

// Request sends a HTTP request to the remote endpoint.
func (hc HTTPClient) Request(
	ctx context.Context,
	options *HTTPRequestOptions,
) (*http.Response, error) {
	method := http.MethodGet

	if options.Method != nil && *options.Method != "" {
		method = string(*options.Method)
	}

	_, err := exhttp.ParseHttpURL(options.URL)
	if err != nil {
		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	var body io.Reader

	if method == http.MethodPost && options.BodyText != "" {
		body = strings.NewReader(options.BodyText)
	}

	req, err := http.NewRequestWithContext(ctx, method, options.URL, body)
	if err != nil {
		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	for _, kv := range options.Headers {
		if kv.Key == "" {
			continue
		}

		req.Header.Set(kv.Key, kv.Value)
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	return resp, nil
}
