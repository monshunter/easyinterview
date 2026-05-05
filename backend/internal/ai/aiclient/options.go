package aiclient

// Option mutates Client construction. The functional-option pattern lets
// tests opt in to stub allowance and lets future plans inject custom
// observability writers without breaking the New(cfg) signature.
type Option func(*clientOptions)

type clientOptions struct {
	stubAllowed      bool
	resolver         ProfileResolver
	providers        map[string]Provider
	providerResolver ProviderResolver
	taskRunWriter    AITaskRunWriter
	auditWriter      AuditEventWriter
}

// WithStubAllowed permits the stub provider to be instantiated even when
// APP_ENV is not "test". Callers must pass true explicitly; spec §4.4 forbids
// silent stub fallback in any deployment.
func WithStubAllowed(allowed bool) Option {
	return func(o *clientOptions) { o.stubAllowed = allowed }
}

// WithProfileResolver injects a custom resolver. Tests use this to bypass
// the YAML loader; production wiring uses the loader's resolver.
func WithProfileResolver(r ProfileResolver) Option {
	return func(o *clientOptions) { o.resolver = r }
}

// WithProvider registers a Provider under its canonical name. Multiple calls
// accumulate.
func WithProvider(p Provider) Option {
	return func(o *clientOptions) {
		if o.providers == nil {
			o.providers = map[string]Provider{}
		}
		o.providers[p.Name()] = p
	}
}

// WithProviderResolver injects registry-backed provider materialization.
// Tests usually use WithProvider; production wiring uses this option.
func WithProviderResolver(r ProviderResolver) Option {
	return func(o *clientOptions) { o.providerResolver = r }
}

// WithAITaskRunWriter wires the persistence path for ai_task_runs rows.
// B4 owns the schema; A3 only fills typed columns through this interface.
func WithAITaskRunWriter(w AITaskRunWriter) Option {
	return func(o *clientOptions) { o.taskRunWriter = w }
}

// WithAuditEventWriter wires the persistence path for audit_events rows.
func WithAuditEventWriter(w AuditEventWriter) Option {
	return func(o *clientOptions) { o.auditWriter = w }
}
