// Package urlfetch implements the JD URL fetch boundary for the targetjob
// domain. It is deliberately minimal: scheme+host hygiene, DNS-resolved IP
// rejection for private / loopback / link-local / metadata ranges, body
// cap, request timeout, and a fixed User-Agent. The SSRF / size matrix
// matches docs/spec/backend-targetjob/spec.md §4.3 and plan §3.3.
package urlfetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrInvalidSource is returned for synchronous policy rejections (bad
// scheme, private IP, malformed URL, oversize body, blank body). Callers
// map this to B1 TARGET_IMPORT_SOURCE_INVALID.
var ErrInvalidSource = errors.New("target import source is invalid")

// ErrSourceUnavailable is returned for upstream timeouts / network errors /
// 5xx responses. Callers map this to B1 TARGET_IMPORT_SOURCE_UNAVAILABLE.
var ErrSourceUnavailable = errors.New("target import source is temporarily unavailable")

// FetcherOptions configures a Fetcher. UserAgent is required; Timeout and
// BodyCap default to spec D-7 values when zero.
type FetcherOptions struct {
	UserAgent string
	Timeout   time.Duration
	BodyCap   int64
	// Resolver returns IPs for the given host. Defaults to net.LookupIP.
	// Tests inject a deterministic resolver to drive private-network checks.
	Resolver func(ctx context.Context, host string) ([]net.IP, error)
	// HTTPClient is used for the GET request. When nil, a default client
	// is constructed with the timeout and a redirect policy that re-runs
	// the IP check on every hop.
	HTTPClient *http.Client
	// Now returns the wall clock used to stamp FetchedAt. Defaults to
	// time.Now().UTC().
	Now func() time.Time
}

// Fetcher pulls JD HTML / text bodies under the spec D-7 envelope.
type Fetcher struct {
	ua       string
	timeout  time.Duration
	bodyCap  int64
	resolver func(ctx context.Context, host string) ([]net.IP, error)
	client   *http.Client
	now      func() time.Time
}

// FetchResult is the structured return of a successful fetch.
type FetchResult struct {
	SanitizedURL string
	Body         string
	ContentType  string
	StatusCode   int
	FetchedAt    time.Time
}

// New constructs a Fetcher. UserAgent must be non-empty (use
// targetjob.URLFetchUserAgent for the canonical value).
func New(opts FetcherOptions) *Fetcher {
	if strings.TrimSpace(opts.UserAgent) == "" {
		panic("urlfetch: UserAgent is required")
	}
	if opts.Timeout == 0 {
		opts.Timeout = 10 * time.Second
	}
	if opts.BodyCap == 0 {
		opts.BodyCap = 1 << 20
	}
	if opts.Now == nil {
		opts.Now = func() time.Time { return time.Now().UTC() }
	}
	resolver := opts.Resolver
	if resolver == nil {
		resolver = func(ctx context.Context, host string) ([]net.IP, error) {
			return net.DefaultResolver.LookupIP(ctx, "ip", host)
		}
	}
	f := &Fetcher{
		ua:       opts.UserAgent,
		timeout:  opts.Timeout,
		bodyCap:  opts.BodyCap,
		resolver: resolver,
		now:      opts.Now,
	}
	if opts.HTTPClient != nil {
		f.client = opts.HTTPClient
	} else {
		f.client = &http.Client{
			Timeout: f.timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				if err := f.checkURLPolicy(req.Context(), req.URL); err != nil {
					return err
				}
				return nil
			},
		}
	}
	return f
}

// Fetch resolves the supplied URL and returns its body if the entire path
// stays within policy. Errors are wrapped around ErrInvalidSource (caller
// policy violations) or ErrSourceUnavailable (upstream / transport
// problems) so the parse pipeline can map to B1 codes without inspecting
// strings.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) (FetchResult, error) {
	u, err := f.parseAndValidate(rawURL)
	if err != nil {
		return FetchResult{}, err
	}
	if err := f.checkURLPolicy(ctx, u); err != nil {
		return FetchResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return FetchResult{}, fmt.Errorf("%w: build request: %v", ErrInvalidSource, err)
	}
	req.Header.Set("User-Agent", f.ua)
	req.Header.Set("Accept", "text/html, text/plain;q=0.9, application/xhtml+xml;q=0.8")

	resp, err := f.client.Do(req)
	if err != nil {
		// Distinguish timeout / DNS / private-network rejection.
		if isPolicyError(err) {
			return FetchResult{}, fmt.Errorf("%w: %v", ErrInvalidSource, err)
		}
		return FetchResult{}, fmt.Errorf("%w: %v", ErrSourceUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return FetchResult{}, fmt.Errorf("%w: upstream status %d", ErrSourceUnavailable, resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		return FetchResult{}, fmt.Errorf("%w: upstream status %d", ErrInvalidSource, resp.StatusCode)
	}

	// Read at most BodyCap+1 bytes; reject if oversized.
	limited := io.LimitReader(resp.Body, f.bodyCap+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return FetchResult{}, fmt.Errorf("%w: read body: %v", ErrSourceUnavailable, err)
	}
	if int64(len(raw)) > f.bodyCap {
		return FetchResult{}, fmt.Errorf("%w: response body exceeded %d bytes", ErrInvalidSource, f.bodyCap)
	}
	body := strings.TrimSpace(string(raw))
	if body == "" {
		return FetchResult{}, fmt.Errorf("%w: response body is empty", ErrInvalidSource)
	}

	return FetchResult{
		SanitizedURL: u.String(),
		Body:         body,
		ContentType:  resp.Header.Get("Content-Type"),
		StatusCode:   resp.StatusCode,
		FetchedAt:    f.now(),
	}, nil
}

func (f *Fetcher) parseAndValidate(rawURL string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, fmt.Errorf("%w: malformed url: %v", ErrInvalidSource, err)
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return nil, fmt.Errorf("%w: scheme must be https", ErrInvalidSource)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("%w: host is required", ErrInvalidSource)
	}
	// Reject userinfo / fragment so the runner does not accidentally pass
	// credentials, query secrets, or anchor strings.
	u.User = nil
	u.RawQuery = ""
	u.ForceQuery = false
	u.Fragment = ""
	return u, nil
}

func (f *Fetcher) checkURLPolicy(ctx context.Context, u *url.URL) error {
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("%w: host is required", ErrInvalidSource)
	}
	// If the host is already an IP literal, skip resolution and check it
	// directly. Otherwise resolve and check every returned IP.
	if ip := net.ParseIP(host); ip != nil {
		if !isPublicIP(ip) {
			return fmt.Errorf("%w: target host resolves to non-public address", ErrInvalidSource)
		}
		return nil
	}
	ips, err := f.resolver(ctx, host)
	if err != nil {
		return fmt.Errorf("%w: dns lookup: %v", ErrSourceUnavailable, err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("%w: dns returned no addresses", ErrSourceUnavailable)
	}
	for _, ip := range ips {
		if !isPublicIP(ip) {
			return fmt.Errorf("%w: target host resolves to non-public address", ErrInvalidSource)
		}
	}
	return nil
}

// isPublicIP returns true when the IP is allowed for outbound JD fetch:
// the address is not loopback, link-local, multicast, broadcast, private,
// or part of any reserved metadata range.
func isPublicIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsMulticast() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() ||
		ip.IsPrivate() || ip.IsUnspecified() {
		return false
	}
	// Cloud metadata endpoints (AWS / GCP / Azure / Alibaba). 169.254.169.254
	// is already covered by IsLinkLocalUnicast, but assert it explicitly so
	// the test matrix can document it.
	if ip.Equal(net.IPv4(169, 254, 169, 254)) {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		// Reject 100.64/10 carrier-grade NAT and 198.18/15 benchmarking.
		if ip4[0] == 100 && ip4[1]&0xc0 == 64 {
			return false
		}
		if ip4[0] == 198 && (ip4[1] == 18 || ip4[1] == 19) {
			return false
		}
	}
	return true
}

// isPolicyError returns true when the error chain contains a policy
// rejection produced by checkURLPolicy or the redirect handler. This lets
// transport errors with policy roots map to ErrInvalidSource rather than
// ErrSourceUnavailable.
func isPolicyError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrInvalidSource) {
		return true
	}
	// url.Error wraps the redirect error message; substring fallback.
	return strings.Contains(err.Error(), "non-public address") ||
		strings.Contains(err.Error(), "scheme must be https") ||
		strings.Contains(err.Error(), "too many redirects")
}
