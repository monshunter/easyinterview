package featureflag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	// posthogDecidePath is the self-hosted PostHog decision endpoint
	// (ADR-Q3 §3 / spec D-4). Path + ?v=3 are pinned together; bumping the
	// version requires a spec revision because the response shape changes.
	posthogDecidePath       = "/decide"
	posthogDecideAPIVersion = "3"
	defaultPostHogTimeout   = 2 * time.Second
	defaultPostHogCacheTTL  = 30 * time.Second
)

// PostHogProviderOptions configures PostHogFlagProvider.
type PostHogProviderOptions struct {
	// Host is the self-hosted PostHog base URL (no trailing slash). The
	// provider rejects empty or PostHog Cloud values when SelfHosted=false
	// in staging/prod.
	Host string
	// APIKey is the PostHog project API key (POSTHOG_PROJECT_API_KEY).
	APIKey string
	// SelfHosted must be true in staging/prod. The constructor fails-fast
	// otherwise (spec C-4).
	SelfHosted bool
	// AppEnv is the runtime environment label, used by the self-hosted
	// guard above.
	AppEnv string
	// Public maps flag key -> public visibility. Mirrors the
	// FileFlagProvider's `public` field; used by the runtime-config
	// builder allowlist.
	Public map[string]bool
	// CacheTTL caches /decide responses per distinct id. Defaults to 30s.
	CacheTTL time.Duration
	// HTTPClient is injected for tests. Defaults to http.DefaultClient.
	HTTPClient *http.Client
	// EvalTimeout bounds each /decide call so flag evaluation cannot stall
	// request handling.
	EvalTimeout time.Duration
	// Logger receives degraded-mode warnings.
	Logger *slog.Logger
}

// PostHogFlagProvider evaluates flags against a self-hosted PostHog
// instance via a thin net/http client. The provider intentionally does
// not import github.com/posthog/posthog-go (spec §4.1, lint enforced by
// Phase 4) so business code that ends up vendoring this package never
// gains transitive access to the SDK either.
type PostHogFlagProvider struct {
	host        string
	apiKey      string
	public      map[string]bool
	httpClient  *http.Client
	cacheTTL    time.Duration
	evalTimeout time.Duration
	appEnv      string
	logger      *slog.Logger

	mu             sync.RWMutex
	lastSnapshot   map[string]FlagDecision
	lastFetched    time.Time
	lastDistinctID string
}

// NewPostHogProvider validates options and returns a ready provider. It
// does not call /decide synchronously; the first IsEnabled / Variant call
// triggers the first network request.
func NewPostHogProvider(opts PostHogProviderOptions) (*PostHogFlagProvider, error) {
	env := strings.ToLower(strings.TrimSpace(opts.AppEnv))
	if (env == "staging" || env == "prod") && !opts.SelfHosted {
		return nil, fmt.Errorf("featureflag: POSTHOG_SELF_HOSTED must be true in %s (spec §4.1)", env)
	}
	if opts.Host == "" {
		return nil, fmt.Errorf("featureflag: PostHog host is required")
	}
	if _, err := url.Parse(opts.Host); err != nil {
		return nil, fmt.Errorf("featureflag: invalid PostHog host %q: %w", opts.Host, err)
	}
	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: defaultPostHogTimeout}
	}
	cacheTTL := opts.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = defaultPostHogCacheTTL
	}
	evalTimeout := opts.EvalTimeout
	if evalTimeout <= 0 {
		evalTimeout = defaultPostHogTimeout
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	publicCopy := make(map[string]bool, len(opts.Public))
	for k, v := range opts.Public {
		publicCopy[k] = v
	}
	return &PostHogFlagProvider{
		host:         strings.TrimRight(opts.Host, "/"),
		apiKey:       opts.APIKey,
		public:       publicCopy,
		httpClient:   client,
		cacheTTL:     cacheTTL,
		evalTimeout:  evalTimeout,
		appEnv:       env,
		logger:       logger,
		lastSnapshot: map[string]FlagDecision{},
	}, nil
}

// IsEnabled implements FeatureFlagClient.
func (p *PostHogFlagProvider) IsEnabled(key string, ctx FlagContext) bool {
	return p.evaluate(ctx)[key].Enabled
}

// Variant implements FeatureFlagClient.
func (p *PostHogFlagProvider) Variant(key string, ctx FlagContext) string {
	return p.evaluate(ctx)[key].Variant
}

// Snapshot returns the evaluated map for this request context, using the
// provider cache when fresh. A cold call triggers /decide so public
// runtime-config does not omit flags on first request.
func (p *PostHogFlagProvider) Snapshot(ctx FlagContext) map[string]FlagDecision {
	return p.evaluate(ctx)
}

func (p *PostHogFlagProvider) evaluate(ctx FlagContext) map[string]FlagDecision {
	if cached := p.cachedFor(ctx); cached != nil {
		return cached
	}
	c, cancel := context.WithTimeout(context.Background(), p.evalTimeout)
	defer cancel()
	decisions, err := p.fetchDecisions(c, ctx)
	if err != nil {
		p.logger.Warn("featureflag: PostHog /decide failed; using last-known-good", "error", err.Error())
		p.mu.RLock()
		defer p.mu.RUnlock()
		out := make(map[string]FlagDecision, len(p.lastSnapshot))
		for k, v := range p.lastSnapshot {
			out[k] = v
		}
		return out
	}
	p.mu.Lock()
	p.lastSnapshot = decisions
	p.lastFetched = time.Now()
	p.lastDistinctID = ctx.DistinctID()
	p.mu.Unlock()
	return decisions
}

func (p *PostHogFlagProvider) cachedFor(ctx FlagContext) map[string]FlagDecision {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.lastFetched.IsZero() {
		return nil
	}
	if time.Since(p.lastFetched) >= p.cacheTTL {
		return nil
	}
	if p.lastDistinctID != ctx.DistinctID() {
		return nil
	}
	out := make(map[string]FlagDecision, len(p.lastSnapshot))
	for k, v := range p.lastSnapshot {
		out[k] = v
	}
	return out
}

type decideResponse struct {
	FeatureFlags map[string]any `json:"featureFlags"`
}

func (p *PostHogFlagProvider) fetchDecisions(ctx context.Context, fc FlagContext) (map[string]FlagDecision, error) {
	body, err := json.Marshal(map[string]any{
		"api_key":     p.apiKey,
		"distinct_id": fc.DistinctID(),
	})
	if err != nil {
		return nil, err
	}
	endpoint := p.host + posthogDecidePath + "?v=" + posthogDecideAPIVersion
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("posthog status %d", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("posthog status %d", resp.StatusCode)
	}
	var parsed decideResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode /decide: %w", err)
	}
	out := make(map[string]FlagDecision, len(parsed.FeatureFlags))
	for key, raw := range parsed.FeatureFlags {
		decision := FlagDecision{Public: p.public[key]}
		switch v := raw.(type) {
		case bool:
			decision.Enabled = v
		case string:
			decision.Enabled = true
			decision.Variant = v
		default:
			// Unknown shape: treat as disabled but keep public flag.
		}
		out[key] = decision
	}
	return out, nil
}
