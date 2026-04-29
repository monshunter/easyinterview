# aiclient

Provider-neutral AIClient for every LLM, embedding, and (future) STT call
inside the easyinterview backend. This package owns the public Go interface
[`AIClient`](./aiclient.go), the runtime [`AICallMeta`](./meta.go), the
Model Profile schema in [`profile/`](./profile/loader.go), and the
deterministic stub plus OpenAI-compatible adapter in [`providers/`](./providers/).

Spec authority: [docs/spec/ai-gateway-and-model-routing/spec.md](../../../../docs/spec/ai-gateway-and-model-routing/spec.md).
ADR authority: [docs/spec/engineering-roadmap/decisions/ADR-Q6-ai-gateway-and-model-routing.md](../../../../docs/spec/engineering-roadmap/decisions/ADR-Q6-ai-gateway-and-model-routing.md).

## Hard rules

- **Zero vendor SDK.** Business code MUST depend only on the `aiclient`
  package and a Model Profile name. Importing `openai-go`,
  `anthropic-sdk-go`, `cohere-go`, `generative-ai-go`, or any other vendor
  SDK from anywhere inside `backend/` is a hard violation (spec §6 C-2).
- **AICallMeta is owned here.** Callers receive it as the second return
  value alongside the structured response and cannot mutate it. New
  fields require a spec version bump.
- **Privacy red line.** Log fields, metric labels, `ai_task_runs.metadata`,
  and `audit_events.metadata` MUST NOT contain plaintext prompt or response
  content. Only sha256 hashes, character lengths, and the profile name are
  permitted (spec §4.3 / D-7). The
  [`observability/privacy_test.go`](./observability/privacy_test.go) holds
  the line.
- **Fail-fast on missing gateway.** Outside `APP_ENV=test`, missing
  `AI_GATEWAY_BASE_URL` or `AI_GATEWAY_API_KEY` returns
  [`ErrMissingGatewayConfig`](./config.go) — never silent stub fallback
  (spec §6 C-9).

## Stub provider activation matrix

The deterministic stub provider is allowed only in the situations below;
every other path returns [`stub.ErrNotAllowed`](./providers/stub/stub.go)
and refuses to construct.

| `cfg.AppEnv` | `aiclient.WithStubAllowed` | `stub.WithAppEnv` | `stub.WithAllowed` | Result |
| --- | --- | --- | --- | --- |
| `test` | `true` | `test` | — | OK (the standard unit-test setup) |
| `test` | unset | `test` | — | `aiclient.New` → `ErrMissingGatewayConfig` |
| `dev`/`staging`/`prod`/`docker compose`/`Kind` | any | any | any | `aiclient.New` → `ErrMissingGatewayConfig` unless real gateway env vars are set |
| any | any | non-`test` | `true` | OK at the stub layer (only used by integration tests that explicitly opt in) |
| any | any | non-`test` | `false` | `stub.New` → `ErrNotAllowed` |

`stub.New` requires the boot `APP_ENV` to be passed via
[`stub.WithAppEnv`](./providers/stub/stub.go) — direct `os.Getenv` reads
are forbidden by the secrets-and-config boundary lint.

## Local deployment / smoke verification

Local docker compose, Kind, staging, and prod must point the AIClient at a
real OpenAI-compatible endpoint (a real LLM provider or a production AI
gateway). Set both env vars or the process must fail to boot:

```sh
export APP_ENV=dev
export AI_GATEWAY_BASE_URL=https://provider.example/v1
export AI_GATEWAY_API_KEY=sk-...                # NEVER commit
export AI_MODEL_PROFILE_PATH=$(pwd)/config/ai-profiles
```

Smoke verification (run only when you want to exercise a real endpoint;
`-tags smoke` is reserved so the smoke suite stays out of the default
`go test ./...`):

```sh
go test -tags smoke ./internal/ai/aiclient/...
```

The smoke build tag is intentionally NOT wired into CI in plan 001. Real
API keys MUST NOT live in test code, fixture YAML, or the repo.

## Hot reload

[`profile.Loader`](./profile/loader.go) re-scans `AI_MODEL_PROFILE_PATH`
periodically so a YAML edit takes effect within the spec §6 C-4 SLA
(≤30 s). Plan 001 ships the polling reloader (default 5 s cadence) which
plan §2.3 lists as the supported fsnotify-fallback path; an fsnotify
watcher can be wired into the same `Loader` API in a follow-up plan
without changing public surface.

`Reload(ctx) error` is exposed for tests to bypass polling and observe
convergence deterministically. Concurrent reads + reloads are guarded by
an `atomic.Pointer[snapshot]` plus a single-flight reload mutex, covered
by `loader_concurrency_test.go` running 100 reload rounds × 8 readers
under `go test -race`.

## Wiring into cmd/api / cmd/worker

Plan 001 does NOT create or rewrite `cmd/api` or `cmd/worker`. A4
secrets-and-config and each consuming C domain plan own the runtime
entrypoint wiring; this package exposes:

- [`aiclient.New`](./client.go) — the constructor + Phase 4.1 fail-fast
  validation.
- [`aiclient.WithProfileResolver`](./options.go) — bind a
  [`profile.Loader`](./profile/loader.go).
- [`aiclient.WithProvider`](./options.go) — register `stub.Provider` or
  `openai_compatible.Adapter`.
- [`observability.New`](./observability/decorator.go) — wrap the inner
  client with the seven metric families, four log events, the
  `ai_task_runs` writer, and the `audit_events` writer.
- [`AITaskRunWriter`](./writers.go) /
  [`AuditEventWriter`](./writers.go) — DI seams that A4 / B4 / F1 will
  bind to real persistence in their own plans.

A non-test caller is expected to run roughly:

```go
loader, err := profile.NewLoader(profile.Options{Dir: cfg.AIModelProfilePath})
if err != nil { return err }

provider, err := openai_compatible.New(openai_compatible.Options{
    BaseURL: cfg.AIGatewayBaseURL,
    APIKey:  cfg.AIGatewayAPIKey,
})
if err != nil { return err }

inner, err := aiclient.New(aiclient.Config{
    AppEnv:           cfg.AppEnv,
    GatewayBaseURL:   cfg.AIGatewayBaseURL,
    GatewayAPIKey:    cfg.AIGatewayAPIKey,
    ModelProfilePath: cfg.AIModelProfilePath,
},
    aiclient.WithProfileResolver(loader),
    aiclient.WithProvider(provider),
)
if err != nil { return err } // ErrMissingGatewayConfig → non-zero exit

client, err := observability.New(inner,
    observability.WithRegisterer(prom),
    observability.WithLogger(logger),
    observability.WithAITaskRunWriter(taskRunStore),
    observability.WithAuditEventWriter(auditStore),
    observability.WithProfileResolver(loader),
)
if err != nil { return err }
```

## Testing

```sh
cd backend
go test ./internal/ai/aiclient/... -count=1 -race
```

This covers:

- `aiclient` end-to-end stub routing + fail-fast matrix
  ([config_test.go](./config_test.go))
- `profile` loader + ≤30 s convergence + concurrent read/reload race
- `providers/stub` deterministic output + APP_ENV gate
- `providers/openai_compatible` contract (chat / embeddings / 5xx /
  4xx envelope / timeout / fallback headers / missing choices)
- `observability` decorator metrics / logs / DB / audit
- `observability` privacy white-box (no plaintext leak across all four
  output sinks)
