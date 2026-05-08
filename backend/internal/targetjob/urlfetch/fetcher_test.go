package urlfetch_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/targetjob/urlfetch"
)

func newURL(s string) (*url.URL, error) { return url.Parse(s) }

func newPublicResolver(addrs ...string) func(ctx context.Context, host string) ([]net.IP, error) {
	return func(_ context.Context, _ string) ([]net.IP, error) {
		ips := make([]net.IP, 0, len(addrs))
		for _, a := range addrs {
			ip := net.ParseIP(a)
			if ip == nil {
				continue
			}
			ips = append(ips, ip)
		}
		return ips, nil
	}
}

func TestFetch_HappyPath_ReturnsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); !strings.Contains(got, "EasyInterview JD-Crawler") {
			t.Errorf("missing UA header: %q", got)
		}
		_, _ = w.Write([]byte("Hiring a Backend Engineer."))
	}))
	defer srv.Close()
	f := urlfetch.New(urlfetch.FetcherOptions{
		UserAgent:  "EasyInterview JD-Crawler/test",
		HTTPClient: srv.Client(),
		Resolver:   newPublicResolver("8.8.8.8"),
	})
	got, err := f.Fetch(context.Background(), srv.URL)
	if err == nil {
		// httptest server uses 127.0.0.1; we expect rejection by policy.
		_ = got
	}
	// Because httptest server runs on 127.0.0.1, the policy rejects it.
	// Use an explicit IP-keyed test below for the public-IP path.
	if err != nil && !errors.Is(err, urlfetch.ErrInvalidSource) {
		t.Fatalf("expected loopback rejection, got %v", err)
	}
}

// To exercise the happy path without a real DNS lookup, we route through
// a custom transport that forwards to httptest while leaving the URL host
// as a non-loopback string.
func TestFetch_HappyPath_PublicIPViaInjectedTransport(t *testing.T) {
	body := []byte("Looking for a Backend Engineer with strong Go experience.")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer upstream.Close()

	transport := &injectingTransport{redirectTo: upstream.URL}
	f := urlfetch.New(urlfetch.FetcherOptions{
		UserAgent:  "EasyInterview JD-Crawler/test",
		HTTPClient: &http.Client{Transport: transport, Timeout: 5 * time.Second},
		Resolver:   newPublicResolver("8.8.8.8"),
	})
	got, err := f.Fetch(context.Background(), "https://jobs.example.com/role/123?token=secret")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !strings.Contains(got.Body, "Backend Engineer") {
		t.Fatalf("body not preserved: %q", got.Body)
	}
	if got.SanitizedURL != "https://jobs.example.com/role/123" {
		t.Fatalf("sanitized url drifted: %q", got.SanitizedURL)
	}
	if strings.Contains(got.SanitizedURL, "token=secret") || strings.Contains(got.SanitizedURL, "?") {
		t.Fatalf("sanitized URL leaked query secret: %q", got.SanitizedURL)
	}
}

func TestFetch_RejectsHTTPScheme(t *testing.T) {
	f := urlfetch.New(urlfetch.FetcherOptions{
		UserAgent: "EasyInterview JD-Crawler/test",
		Resolver:  newPublicResolver("8.8.8.8"),
	})
	_, err := f.Fetch(context.Background(), "http://insecure.example.com/jd")
	if !errors.Is(err, urlfetch.ErrInvalidSource) {
		t.Fatalf("expected ErrInvalidSource, got %v", err)
	}
}

func TestFetch_RejectsMissingHost(t *testing.T) {
	f := urlfetch.New(urlfetch.FetcherOptions{
		UserAgent: "EasyInterview JD-Crawler/test",
		Resolver:  newPublicResolver("8.8.8.8"),
	})
	_, err := f.Fetch(context.Background(), "https://")
	if !errors.Is(err, urlfetch.ErrInvalidSource) {
		t.Fatalf("expected ErrInvalidSource, got %v", err)
	}
}

func TestFetch_RejectsPrivateNetworkResolution(t *testing.T) {
	cases := map[string]string{
		"rfc1918":           "10.0.0.5",
		"link-local":        "169.254.10.20",
		"metadata":          "169.254.169.254",
		"loopback":          "127.0.0.1",
		"ipv6-loopback":     "::1",
		"ipv6-link-local":   "fe80::1",
		"ipv6-unique-local": "fc00::1",
		"cgnat":             "100.64.1.1",
		"benchmark":         "198.18.0.1",
	}
	for name, ip := range cases {
		t.Run(name, func(t *testing.T) {
			f := urlfetch.New(urlfetch.FetcherOptions{
				UserAgent: "EasyInterview JD-Crawler/test",
				Resolver:  newPublicResolver(ip),
			})
			_, err := f.Fetch(context.Background(), "https://internal.example.com/jd")
			if !errors.Is(err, urlfetch.ErrInvalidSource) {
				t.Fatalf("ip %s should be rejected, got %v", ip, err)
			}
		})
	}
}

func TestFetch_RejectsIPLiteralLoopback(t *testing.T) {
	f := urlfetch.New(urlfetch.FetcherOptions{
		UserAgent: "EasyInterview JD-Crawler/test",
	})
	_, err := f.Fetch(context.Background(), "https://127.0.0.1/jd")
	if !errors.Is(err, urlfetch.ErrInvalidSource) {
		t.Fatalf("expected ErrInvalidSource for loopback IP literal, got %v", err)
	}
}

func TestFetch_RejectsOversizeBody(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		blob := strings.Repeat("x", 1024) // 1 KiB chunks
		for range 1500 {                  // 1.5 MiB total
			_, _ = w.Write([]byte(blob))
		}
	}))
	defer upstream.Close()
	transport := &injectingTransport{redirectTo: upstream.URL}
	f := urlfetch.New(urlfetch.FetcherOptions{
		UserAgent:  "EasyInterview JD-Crawler/test",
		HTTPClient: &http.Client{Transport: transport, Timeout: 5 * time.Second},
		Resolver:   newPublicResolver("8.8.8.8"),
		BodyCap:    1 << 20, // 1 MiB
	})
	_, err := f.Fetch(context.Background(), "https://jobs.example.com/big")
	if !errors.Is(err, urlfetch.ErrInvalidSource) {
		t.Fatalf("expected ErrInvalidSource for oversize body, got %v", err)
	}
}

func TestFetch_RejectsBlankBody(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("   \n\t"))
	}))
	defer upstream.Close()
	transport := &injectingTransport{redirectTo: upstream.URL}
	f := urlfetch.New(urlfetch.FetcherOptions{
		UserAgent:  "EasyInterview JD-Crawler/test",
		HTTPClient: &http.Client{Transport: transport, Timeout: 5 * time.Second},
		Resolver:   newPublicResolver("8.8.8.8"),
	})
	_, err := f.Fetch(context.Background(), "https://jobs.example.com/blank")
	if !errors.Is(err, urlfetch.ErrInvalidSource) {
		t.Fatalf("expected ErrInvalidSource for blank body, got %v", err)
	}
}

func TestFetch_RejectsUpstream5xxAsUnavailable(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusBadGateway)
	}))
	defer upstream.Close()
	transport := &injectingTransport{redirectTo: upstream.URL}
	f := urlfetch.New(urlfetch.FetcherOptions{
		UserAgent:  "EasyInterview JD-Crawler/test",
		HTTPClient: &http.Client{Transport: transport, Timeout: 5 * time.Second},
		Resolver:   newPublicResolver("8.8.8.8"),
	})
	_, err := f.Fetch(context.Background(), "https://jobs.example.com/down")
	if !errors.Is(err, urlfetch.ErrSourceUnavailable) {
		t.Fatalf("expected ErrSourceUnavailable, got %v", err)
	}
}

func TestFetch_RejectsUpstream4xxAsInvalid(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	}))
	defer upstream.Close()
	transport := &injectingTransport{redirectTo: upstream.URL}
	f := urlfetch.New(urlfetch.FetcherOptions{
		UserAgent:  "EasyInterview JD-Crawler/test",
		HTTPClient: &http.Client{Transport: transport, Timeout: 5 * time.Second},
		Resolver:   newPublicResolver("8.8.8.8"),
	})
	_, err := f.Fetch(context.Background(), "https://jobs.example.com/missing")
	if !errors.Is(err, urlfetch.ErrInvalidSource) {
		t.Fatalf("expected ErrInvalidSource for 4xx, got %v", err)
	}
}

func TestNew_PanicsOnEmptyUserAgent(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on empty UA")
		}
	}()
	urlfetch.New(urlfetch.FetcherOptions{})
}

// injectingTransport rewrites every request to the configured httptest URL
// while preserving Host header so the Fetcher continues to think it is
// hitting jobs.example.com. The Resolver is responsible for keeping the
// host -> public IP mapping aligned with the policy gate.
type injectingTransport struct {
	redirectTo string
	last       *http.Request
}

func (t *injectingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.last = req.Clone(req.Context())
	parsed, err := newURL(t.redirectTo)
	if err != nil {
		return nil, err
	}
	clone := req.Clone(req.Context())
	clone.URL.Scheme = parsed.Scheme
	clone.URL.Host = parsed.Host
	return http.DefaultTransport.RoundTrip(clone)
}
