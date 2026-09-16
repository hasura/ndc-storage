package common

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"
)

// URLSafetyMode controls how strict ValidateEgressURL is about blocking
// destinations that could be abused for server-side request forgery (SSRF).
type URLSafetyMode int

const (
	// URLSafetyStaticConfig is used for endpoints that come from the static
	// connector configuration. It blocks cloud metadata, loopback and
	// link-local destinations but still allows private ranges so that
	// self-hosted services (e.g. a MinIO server on a private network) keep
	// working.
	URLSafetyStaticConfig URLSafetyMode = iota
	// URLSafetyDynamicCredential is used for endpoints supplied at request time
	// via GraphQL dynamic credentials. In addition to the static blocks it also
	// rejects private/internal ranges, since callers must never be able to
	// point the connector at VPC-internal services.
	URLSafetyDynamicCredential
)

// blockedHostnames is the set of hostnames rejected in every mode.
var blockedHostnames = map[string]struct{}{
	"metadata.google.internal": {},
	"metadata":                 {},
	"localhost":                {},
	"localhost.localdomain":    {},
}

// alwaysBlockedRanges are rejected in every mode: cloud metadata, loopback,
// link-local and the unspecified network.
var alwaysBlockedRanges = mustParsePrefixes(
	"169.254.0.0/16", // link-local, includes cloud metadata 169.254.169.254
	"fe80::/10",      // IPv6 link-local
	"127.0.0.0/8",    // IPv4 loopback
	"::1/128",        // IPv6 loopback
	"0.0.0.0/8",      // "this" network / unspecified
)

// privateBlockedRanges are additionally rejected in dynamic-credential mode.
var privateBlockedRanges = mustParsePrefixes(
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"100.64.0.0/10", // carrier-grade NAT / shared address space
	"fc00::/7",      // IPv6 unique local addresses
)

// errBlockedEgressURL is the sentinel error wrapped by egress validation
// failures.
var errBlockedEgressURL = errors.New("the destination host is not allowed for security reasons")

// lookupIP resolves a hostname to its IP addresses. It is a package variable so
// tests can stub DNS resolution.
var lookupIP = net.LookupIP

func mustParsePrefixes(cidrs ...string) []netip.Prefix {
	prefixes := make([]netip.Prefix, len(cidrs))

	for i, c := range cidrs {
		p, err := netip.ParsePrefix(c)
		if err != nil {
			panic("invalid CIDR in SSRF block list: " + c)
		}

		prefixes[i] = p
	}

	return prefixes
}

// ValidateEgressURL validates that raw is a safe URL for the connector to send
// outbound requests to, according to the given safety mode. It returns an error
// if the URL targets a cloud metadata endpoint, a loopback/link-local address
// or (in dynamic-credential mode) a private/internal range.
func ValidateEgressURL(raw string, mode URLSafetyMode) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("%w: empty url", errBlockedEgressURL)
	}

	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("%w: unsupported url scheme %q", errBlockedEgressURL, u.Scheme)
	}

	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("%w: empty host", errBlockedEgressURL)
	}

	return validateEgressHost(host, mode)
}

// ValidateAzureConnectionString validates an Azure blob connection string or
// endpoint for SSRF safety. It best-effort parses the BlobEndpoint= field and
// also rejects any string that references the cloud metadata IP directly.
func ValidateAzureConnectionString(raw string, mode URLSafetyMode) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("%w: empty endpoint", errBlockedEgressURL)
	}

	if strings.Contains(raw, "169.254.169.254") {
		return fmt.Errorf("%w: cloud metadata endpoint", errBlockedEgressURL)
	}

	// A plain URL can be validated directly.
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return ValidateEgressURL(trimmed, mode)
	}

	// Otherwise best-effort parse connection-string segments and validate the
	// BlobEndpoint value if present.
	for part := range strings.SplitSeq(raw, ";") {
		key, value, found := strings.Cut(part, "=")
		if !found {
			continue
		}

		if strings.EqualFold(strings.TrimSpace(key), "BlobEndpoint") {
			if err := ValidateEgressURL(strings.TrimSpace(value), mode); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateEgressHost(host string, mode URLSafetyMode) error {
	host = strings.ToLower(strings.TrimSuffix(host, "."))

	if isBlockedHostname(host) {
		return fmt.Errorf("%w: %s", errBlockedEgressURL, host)
	}

	// If the host is an IP literal, check it directly.
	if addr, err := netip.ParseAddr(host); err == nil {
		if isBlockedIP(addr, mode) {
			return fmt.Errorf("%w: %s", errBlockedEgressURL, host)
		}

		return nil
	}

	// Otherwise resolve the hostname and reject if any resolved address falls
	// into a blocked range. An unresolvable host is not treated as an SSRF hit;
	// the outbound request would simply fail to connect.
	ips, err := lookupIP(host)
	if err != nil {
		return nil //nolint:nilerr // DNS failures are left to the outbound client's error handling.
	}

	for _, ip := range ips {
		addr, ok := netip.AddrFromSlice(ip)
		if !ok {
			continue
		}

		addr = addr.Unmap()
		if isBlockedIP(addr, mode) {
			return fmt.Errorf(
				"%w: %s resolves to blocked address %s",
				errBlockedEgressURL,
				host,
				addr,
			)
		}
	}

	return nil
}

func isBlockedHostname(host string) bool {
	if _, ok := blockedHostnames[host]; ok {
		return true
	}

	// *.localhost is reserved as loopback per RFC 6761.
	return strings.HasSuffix(host, ".localhost")
}

func isBlockedIP(addr netip.Addr, mode URLSafetyMode) bool {
	addr = addr.Unmap()

	for _, p := range alwaysBlockedRanges {
		if p.Contains(addr) {
			return true
		}
	}

	if mode == URLSafetyDynamicCredential {
		for _, p := range privateBlockedRanges {
			if p.Contains(addr) {
				return true
			}
		}
	}

	return false
}
