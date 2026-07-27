package common

import (
	"net"
	"testing"
)

func TestValidateEgressURL(t *testing.T) {
	testCases := []struct {
		name    string
		url     string
		mode    URLSafetyMode
		wantErr bool
	}{
		// Cloud metadata and loopback/link-local are rejected in every mode.
		{name: "gcp metadata ip dynamic", url: "http://169.254.169.254", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "gcp metadata ip static", url: "http://169.254.169.254", mode: URLSafetyStaticConfig, wantErr: true},
		{name: "gcp metadata ip path dynamic", url: "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "gcp metadata hostname dynamic", url: "http://metadata.google.internal/x", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "gcp metadata hostname static", url: "http://metadata.google.internal", mode: URLSafetyStaticConfig, wantErr: true},
		{name: "metadata short hostname", url: "http://metadata", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "loopback ip dynamic", url: "http://127.0.0.1:9000", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "loopback ip static", url: "http://127.0.0.1:9000", mode: URLSafetyStaticConfig, wantErr: true},
		{name: "localhost dynamic", url: "http://localhost:9000", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "localhost static", url: "http://localhost:9000", mode: URLSafetyStaticConfig, wantErr: true},
		{name: "subdomain localhost", url: "http://foo.localhost", mode: URLSafetyStaticConfig, wantErr: true},
		{name: "case insensitive host", url: "http://LocalHost", mode: URLSafetyStaticConfig, wantErr: true},
		{name: "ipv6 loopback", url: "http://[::1]:9000", mode: URLSafetyStaticConfig, wantErr: true},
		{name: "unspecified ipv4", url: "http://0.0.0.0", mode: URLSafetyStaticConfig, wantErr: true},

		// Private ranges are only rejected in dynamic-credential mode.
		{name: "10.x dynamic rejected", url: "http://10.0.0.1", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "192.168.x dynamic rejected", url: "http://192.168.1.1", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "172.16.x dynamic rejected", url: "http://172.16.0.1", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "cgnat dynamic rejected", url: "http://100.64.0.1", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "private minio static allowed", url: "http://10.0.0.5:9000", mode: URLSafetyStaticConfig, wantErr: false},
		{name: "private 192.168 static allowed", url: "http://192.168.1.1:9000", mode: URLSafetyStaticConfig, wantErr: false},
		{name: "private 172.16 static allowed", url: "http://172.16.0.1:9000", mode: URLSafetyStaticConfig, wantErr: false},

		// Public endpoints are allowed in both modes.
		{name: "aws s3 https", url: "https://s3.amazonaws.com", mode: URLSafetyDynamicCredential, wantErr: false},
		{name: "gcs https", url: "https://storage.googleapis.com", mode: URLSafetyDynamicCredential, wantErr: false},

		// Non-http schemes are always rejected.
		{name: "file scheme", url: "file:///etc/passwd", mode: URLSafetyStaticConfig, wantErr: true},
		{name: "ftp scheme", url: "ftp://example.com", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "gopher scheme", url: "gopher://169.254.169.254", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "empty url", url: "", mode: URLSafetyStaticConfig, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEgressURL(tc.url, tc.mode)
			if tc.wantErr && err == nil {
				t.Fatalf("ValidateEgressURL(%q, %v): expected error, got nil", tc.url, tc.mode)
			}

			if !tc.wantErr && err != nil {
				t.Fatalf("ValidateEgressURL(%q, %v): expected no error, got %v", tc.url, tc.mode, err)
			}
		})
	}
}

func TestValidateEgressURLResolvedHost(t *testing.T) {
	// Stub DNS resolution so the test does not depend on the network and to
	// confirm that a hostname resolving to a blocked address is rejected.
	original := lookupIP
	t.Cleanup(func() { lookupIP = original })

	lookupIP = func(host string) ([]net.IP, error) {
		switch host {
		case "evil.example.com":
			return []net.IP{net.ParseIP("169.254.169.254")}, nil
		case "internal.example.com":
			return []net.IP{net.ParseIP("10.1.2.3")}, nil
		case "public.example.com":
			return []net.IP{net.ParseIP("93.184.216.34")}, nil
		default:
			return nil, &net.DNSError{Err: "no such host", Name: host}
		}
	}

	cases := []struct {
		name    string
		url     string
		mode    URLSafetyMode
		wantErr bool
	}{
		{name: "resolves to metadata", url: "http://evil.example.com", mode: URLSafetyStaticConfig, wantErr: true},
		{name: "resolves to private dynamic", url: "http://internal.example.com", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "resolves to private static allowed", url: "http://internal.example.com", mode: URLSafetyStaticConfig, wantErr: false},
		{name: "resolves to public", url: "http://public.example.com", mode: URLSafetyDynamicCredential, wantErr: false},
		{name: "unresolvable host allowed", url: "http://does-not-resolve.example.com", mode: URLSafetyDynamicCredential, wantErr: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEgressURL(tc.url, tc.mode)
			if tc.wantErr && err == nil {
				t.Fatalf("ValidateEgressURL(%q): expected error, got nil", tc.url)
			}

			if !tc.wantErr && err != nil {
				t.Fatalf("ValidateEgressURL(%q): expected no error, got %v", tc.url, err)
			}
		})
	}
}

func TestValidateAzureConnectionString(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		mode    URLSafetyMode
		wantErr bool
	}{
		{name: "raw metadata ip", raw: "DefaultEndpointsProtocol=https;BlobEndpoint=http://169.254.169.254/", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "blob endpoint loopback", raw: "BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;AccountName=x", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "blob endpoint private dynamic", raw: "BlobEndpoint=http://10.0.0.5:10000/x", mode: URLSafetyDynamicCredential, wantErr: true},
		{name: "blob endpoint public", raw: "BlobEndpoint=https://myacct.blob.core.windows.net/", mode: URLSafetyDynamicCredential, wantErr: false},
		{name: "plain public url", raw: "https://myacct.blob.core.windows.net/", mode: URLSafetyDynamicCredential, wantErr: false},
		{name: "plain metadata url", raw: "http://169.254.169.254", mode: URLSafetyStaticConfig, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAzureConnectionString(tc.raw, tc.mode)
			if tc.wantErr && err == nil {
				t.Fatalf("ValidateAzureConnectionString(%q): expected error, got nil", tc.raw)
			}

			if !tc.wantErr && err != nil {
				t.Fatalf("ValidateAzureConnectionString(%q): expected no error, got %v", tc.raw, err)
			}
		})
	}
}
