package urlfetch

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestFetch_RejectsDNSRebindOnDial(t *testing.T) {
	calls := 0
	resolver := func(_ context.Context, _ string) ([]net.IP, error) {
		calls++
		if calls == 1 {
			return []net.IP{net.ParseIP("8.8.8.8")}, nil
		}
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	}

	f := New(FetcherOptions{
		UserAgent: "EasyInterview JD-Crawler/test",
		Timeout:   50 * time.Millisecond,
		Resolver:  resolver,
	})

	_, err := f.Fetch(context.Background(), "https://jobs.example.test/role")
	if !errors.Is(err, ErrInvalidSource) {
		t.Fatalf("expected ErrInvalidSource when dial resolution rebinds to private IP, got %v", err)
	}
	if calls < 2 {
		t.Fatalf("resolver was called %d time(s); dial path must re-resolve and enforce policy", calls)
	}
}

func TestDialContextRejectsPrivateResolvedAddress(t *testing.T) {
	f := New(FetcherOptions{
		UserAgent: "EasyInterview JD-Crawler/test",
		Resolver: func(_ context.Context, _ string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("169.254.169.254")}, nil
		},
	})

	_, err := f.resolveDialAddress(context.Background(), "tcp", "jobs.example.test:443")
	if !errors.Is(err, ErrInvalidSource) {
		t.Fatalf("expected ErrInvalidSource for private dial resolution, got %v", err)
	}
}
