package featureflag

// FlagContext carries the minimal fields a flag provider needs to evaluate
// a flag. The shape is locked by spec D-4: anonymous distinct id,
// authenticated user public id, and the app environment label. Adding new
// fields requires a spec revision so business code cannot smuggle PII or
// session details into flag evaluation.
type FlagContext struct {
	AnonymousDistinctID string
	AuthenticatedUserID string
	AppEnv              string
}

// DistinctID returns the most specific identifier available, preferring
// authenticated user id over anonymous distinct id. Used by providers that
// need a single id field (e.g. PostHog `distinct_id`).
func (c FlagContext) DistinctID() string {
	if c.AuthenticatedUserID != "" {
		return c.AuthenticatedUserID
	}
	return c.AnonymousDistinctID
}

// FlagDecision is the runtime evaluation result for one flag key. Public
// indicates whether runtime-config (spec D-2) may expose this flag to the
// frontend; provider implementations must always populate it.
type FlagDecision struct {
	Enabled bool
	Variant string
	Public  bool
}

// FeatureFlagClient is the runtime contract every provider must implement.
// IsEnabled / Variant follow spec D-4 exactly; business code should depend
// only on this interface.
type FeatureFlagClient interface {
	IsEnabled(key string, ctx FlagContext) bool
	Variant(key string, ctx FlagContext) string
}

// SnapshotProvider is the runtime-config projection contract. It is separate
// from FeatureFlagClient so the business consumption interface stays locked to
// spec D-4 while the public runtime-config builder can evaluate all flags for
// one request context.
type SnapshotProvider interface {
	Snapshot(ctx FlagContext) map[string]FlagDecision
}
