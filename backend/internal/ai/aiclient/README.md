# aiclient

Provider-neutral AIClient for every chat/LLM and STT transcription call
inside the easyinterview backend. This package owns the public Go interface
[`AIClient`](./aiclient.go), the runtime [`AICallMeta`](./meta.go), the
Provider Registry loader in [`providerregistry/`](./providerregistry/loader.go),
Model Profile schema in [`profile/`](./profile/loader.go), and the deterministic
stub plus OpenAI-compatible adapter in [`providers/`](./providers/).

Spec authority: [docs/spec/ai-provider-and-model-routing/spec.md](../../../../docs/spec/ai-provider-and-model-routing/spec.md).
ADR authority: [docs/spec/engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md](../../../../docs/spec/engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md).

## Hard rules

- **Adapter-private SDK boundary.** Business code MUST depend only on the
  `aiclient` package and a Model Profile name. The pinned official
  `openai-go/v3` dependency is allowed only inside `providers/openai_compatible`,
  `providers/judge_compatible`, or the exact shared
  `providers/internal/openaisdk` helper. Public A3 types, business packages,
  config/profile contracts, observability code, and speech adapters MUST NOT
  import or expose SDK types. Other vendor SDKs remain forbidden.
- **AICallMeta is owned here.** Callers receive it as the second return
  value alongside the structured response and cannot mutate it. New
  fields require a spec version bump.
- **Privacy red line.** Log fields, metric labels, `ai_task_runs.metadata`,
  and `audit_events.metadata` MUST NOT contain plaintext prompt or response
  content. Only sha256 hashes, character lengths, and the profile name are
  permitted (spec §4.3 / D-7). The
  [`observability/privacy_test.go`](./observability/privacy_test.go) holds
  the line.
- **Provider registry is the connection contract.** Checked-in
  `config/ai-providers.yaml` stores provider refs, protocols, capabilities, and
  secret env names only. Secret values come from A4 `SecretSource` and must
  never be written to YAML, logs, metrics, DB metadata, or audit metadata.
- **Fail closed on unsupported capability.** Profiles with
  `status=disabled|unsupported`, or capabilities whose adapter is not active,
  return an AI error instead of falling back to chat or stub.

## Stub provider activation matrix

The deterministic stub provider is allowed only in the situations below;
every other path returns [`stub.ErrNotAllowed`](./providers/stub/stub.go)
and refuses to construct.

| `cfg.AppEnv` | `aiclient.WithStubAllowed` | `stub.WithAppEnv` | `stub.WithAllowed` | Result |
| --- | --- | --- | --- | --- |
| `test` | `true` | `test` | — | OK (the standard unit-test setup) |
| `test` | unset | `test` | — | `aiclient.New` → `ErrMissingProviderConfig` |
| `dev`/`staging`/`prod`/`docker compose`/`Kind` | any | any | any | `aiclient.New` → `ErrMissingProviderConfig` unless real provider env vars are set |
| any | any | non-`test` | `true` | OK at the stub layer (only used by integration tests that explicitly opt in) |
| any | any | non-`test` | `false` | `stub.New` → `ErrNotAllowed` |

`stub.New` requires the boot `APP_ENV` to be passed via
[`stub.WithAppEnv`](./providers/stub/stub.go) — direct `os.Getenv` reads
are forbidden by the secrets-and-config boundary lint.

## Local deployment / smoke verification

Local docker compose, Kind, staging, and prod must provide the registry path,
profile catalog path, and any provider-specific env refs selected by active
profiles. `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` may be referenced by
the `deepseek` provider ref, but they are no longer the global AI provider
contract. Current development profiles use `deepseek-v4-flash` for low-latency
interactive work and `deepseek-v4-pro` for report / assessment / rewrite work:

```sh
export APP_ENV=dev
export AI_PROVIDER_REGISTRY_PATH=$(pwd)/config/ai-providers.yaml
export AI_MODEL_PROFILE_PATH=$(pwd)/config/ai-profiles.yaml
export AI_PROVIDER_BASE_URL=https://api.deepseek.com
export AI_PROVIDER_API_KEY=sk-...                # NEVER commit
export AI_DEBUG_PRINT_RAW_OUTPUT=true            # local test/integration default; keep staging/prod off
```

When `AI_DEBUG_PRINT_RAW_OUTPUT=true`, backend observability prints raw
`Complete` responses to stderr so schema and formatting failures can be
debugged. Local test and local frontend/backend integration keep this enabled
by default so agents can inspect real provider output shape. The output remains
outside `ai_task_runs`, `audit_events`, metrics, and structured log fields, and
staging/prod defaults must keep it off.

Smoke verification (run only when you want to exercise a real endpoint;
`-tags smoke` is reserved so the smoke suite stays out of the default
`go test ./...`):

```sh
go test -tags smoke ./internal/ai/aiclient/...
```

The smoke build tag is intentionally NOT wired into CI in plan 001. Real
API keys MUST NOT live in test code, fixture YAML, or the repo.

## Hot reload

[`providerregistry.Loader`](./providerregistry/loader.go) and
[`profile.Loader`](./profile/loader.go) re-scan `AI_MODEL_PROFILE_PATH`
periodically so provider registry or profile YAML edits take effect within the
spec §6 C-4 SLA (≤30 s). Reload failure preserves the last good snapshot and
does not affect in-flight calls.

`Reload(ctx) error` is exposed for tests to bypass polling and observe
convergence deterministically. Concurrent reads + reloads are guarded by
an `atomic.Pointer[snapshot]` plus a single-flight reload mutex, covered
by `loader_concurrency_test.go` running 100 reload rounds × 8 readers
under `go test -race`.

## Wiring into backend runtime

Plan 001 does NOT create or rewrite backend runtime entrypoints. A4
secrets-and-config and each consuming C domain plan own the runtime
entrypoint wiring; this package exposes:

- [`bootstrap.NewClient`](./bootstrap/bootstrap.go) — production runtime
  bootstrap for `AI_PROVIDER_REGISTRY_PATH` + `AI_MODEL_PROFILE_PATH`; it
  validates selected active providers through A4 `SecretSource` and wires
  registry-backed provider refs into the inner client.
- [`aiclient.New`](./client.go) — the small inner constructor used by tests and
  bootstrap after the runtime loaders are ready.
- [`aiclient.WithProfileResolver`](./options.go) — bind a
  [`profile.Loader`](./profile/loader.go).
- [`aiclient.WithProviderResolver`](./options.go) — resolve provider refs from
  the hot-reloaded provider registry.
- [`observability.New`](./observability/decorator.go) — wrap the inner
  client with the seven metric families, four log events, the
  `ai_task_runs` writer, and the `audit_events` writer.
- [`AITaskRunWriter`](./writers.go) /
  [`AuditEventWriter`](./writers.go) — DI seams that A4 / B4 / F1 will
  bind to real persistence in their own plans.

A non-test caller is expected to run roughly:

```go
runtime, err := bootstrap.NewClient(bootstrap.Options{
    Config: aiclient.Config{
        AppEnv:               cfg.AppEnv,
        ProviderRegistryPath: cfg.AIProviderRegistryPath,
        ModelProfilePath:     cfg.AIModelProfilePath,
    },
    SecretSource: secretSource,
    HTTPClient:   httpClient,
    OnWarn:       warn,
})
if err != nil { return err }
defer runtime.Close()

client, err := observability.New(runtime.Client,
    observability.WithRegisterer(prom),
    observability.WithLogger(logger),
    observability.WithAITaskRunWriter(taskRunStore),
    observability.WithAuditEventWriter(auditStore),
    observability.WithProfileResolver(runtime.Client.Resolver()),
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
- `providerregistry` schema / secret resolution / hot reload / negative fixtures
- `profile` capability schema + ≤30 s convergence + concurrent read/reload race
- `providers/stub` deterministic output + APP_ENV gate
- `providers/openai_compatible` contract (chat / STT / 5xx /
  4xx envelope / timeout / fallback headers / missing choices)
- `observability` decorator metrics / logs / DB / audit
- `observability` privacy white-box (no plaintext leak across all four
  output sinks)
