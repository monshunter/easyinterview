# Config

应用配置、feature flag、AI provider registry 与 Model Profile 的根容器。

A1 [`repo-scaffold`](../docs/spec/repo-scaffold/spec.md) 锁定目录位置；A4
[`secrets-and-config`](../docs/spec/secrets-and-config/spec.md) 锁定字段语义、
loader、redactor 与 lint 红线；A3
[`ai-provider-and-model-routing`](../docs/spec/ai-provider-and-model-routing/spec.md)
消费 `ai-providers.yaml` 与 `ai-profiles.yaml`；F3
[`prompt-rubric-registry`](../docs/spec/prompt-rubric-registry/spec.md) 后续在
本目录扩展 prompt / rubric registry 路径。

所有运行时配置都通过 `backend/internal/platform/config` 暴露的 `Get*` API
读取，禁止业务包直接调用 `os.Getenv`（Phase 4 lint 收口）。

## 三层优先级（spec D-1）

| 优先级 | 来源 | 用途 |
|--------|------|------|
| 1（最高） | runtime secret（`SecretSource`，env / future secret manager） | 真实凭证 |
| 2 | 进程 env var | 环境差异、运行时调优 |
| 3 | `config/{APP_ENV}.yaml` | 环境 override（不含 secret） |
| 4（最低） | `config/config.yaml` | 仓库版本化默认值 |

后写覆盖前者；loader 串行 merge，禁止并发 `Load`。缺 prod 必填字段时进程
fail-fast 并打印缺失 key 名（spec C-2）。

## 文件用途

| 文件 | 用途 |
|------|------|
| `config.yaml` | canonical 默认值（spec §3.1.2）；secret 字段保持空字符串 |
| `dev.yaml` | dev 环境 override；`log.level=debug`、`featureFlag.source=file`、本地 AI raw output debug 默认开启 |
| `test.yaml` | test 环境 override；本地测试 AI raw output debug 默认开启，secret 字段仍为空 |
| `staging.yaml` | staging 环境 override；`featureFlag.source=posthog` + `posthogSelfHosted=true` |
| `prod.yaml` | prod 环境 override；同 staging，但生产差异在此处沉淀 |
| `feature-flags.yaml` | `FileFlagProvider` 本地真理源；6 项 P0 baseline flag；显式标 `public: true|false` |
| `ai-providers.yaml` | A3 Provider Registry；只保存 provider ref、protocol、capabilities 与 secret env ref，不保存 secret 明文 |
| `ai-profiles.yaml` | A3 Model Profile catalog；顶层 `profiles[]` 使用 `capability` / `provider_ref` / `status`，不可执行能力必须写 `unsupported_reason` |

当前 P0 baseline flag 固定为 `practice_hint_enabled`、
`report_evidence_v2_enabled`、`report_retry_plan_enabled`、
`readiness_signals_enabled`、`ai_fallback_model_enabled`、
`practice_assistance_mode_enabled`。`runtime-config` 仅按此清单投影公开
flag；非当前错题本、成长看板与双轨 mock session flag 已随 product-scope v1.2
移除，不再作为配置能力保留。

`*.secret.yaml` / `*.secret.json` / `local.*.yaml` / `feature-flags.local.yaml`
默认进 `.gitignore`（spec §4.2 / D-6）。

## 新增 env key 流程（4 步）

任意 env key 增删改必须按顺序：

1. 递增 [`secrets-and-config` spec](../docs/spec/secrets-and-config/spec.md)
   §3.1.1 表格与 history 字段，明确 Owner subspec 与 prod required。
2. 同步本目录的 [`.env.example`](../.env.example)，每行注释包含 Owner 与
   prod required；secret 字段只允许写注入说明，不允许真实凭证。
3. 在 `backend/internal/platform/config/validator.go` 中加入 fail-fast 检查
   （或扩展 `Loader.Validate` 调用链），错误信息列出缺失 key 名。
4. 跑 `make lint-config`：三方求差集（`.env.example` / 代码侧 `os.Getenv` /
   spec §3.1.1 表）必须全部对齐，否则 lint 失败。

`async.queueWeights` 是 config-only 字段（spec D-9），不进 env 字典；新增
config-only 字段同样需要先递增 spec 版本再同步 `config.yaml`。

## AI provider registry / profile 规则

- `ai-providers.yaml` 的 `providers[]` 只允许写 `name`、`protocol`、
  `base_url_env`、`api_key_env`、`capabilities[]` 和 `version`；`stub` 可不写
  secret env ref，网络出站 protocol 必须写 env ref。
- `ai-profiles.yaml` 顶层只允许 `profiles[]`，每个 profile 必须使用 `capability` 与
  `default.provider_ref`，不得使用 out-of-scope task-category key 或 provider alias key。
- `status=disabled|unsupported` 的 profile 必须写 `unsupported_reason`，运行时
  必须 fail-closed，不得静默降级到 chat 或 stub。
- 当前开发主力 provider ref 是 `deepseek`，profile 只使用
  `deepseek-v4-flash` / `deepseek-v4-pro` 两个模型 ID。
- `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 只是 `deepseek` provider ref
  引用的 env 名，不是全局唯一 provider contract。
- `AI_DEBUG_PRINT_RAW_OUTPUT` 在 local dev/test 和 `deploy/dev-stack/.env.example`
  中默认开启，用于本地调试真实 provider 输出格式；raw output 只允许进入本机
  backend stderr / `.test-output/` 调试日志，不得进入持久化审计、runtime-config、
  staging 或 prod 默认配置。
- 新增 Product/UI AI 场景或 F3 feature_key 时，必须同步
  `docs/spec/ai-provider-and-model-routing/spec.md` §4.5、
  `docs/spec/prompt-rubric-registry/spec.md` §3.1.1 和
  `config/ai-profiles.yaml`，并运行 `python3 scripts/lint/ai_profile_coverage.py
  --repo-root .`。

## RedactedString 使用示范

```go
import "github.com/monshunter/easyinterview/backend/internal/platform/config"

cookieSecret := loader.GetSecret("auth.sessionCookieSecret")
// 直接打印 / JSON marshal / fmt.Errorf("...: %w") 都输出 ***
log.Info("session cookie secret loaded", "secret", cookieSecret)

// 仅在交给最终 SDK 时一次性 Reveal，不要再传给业务层。
cookie.SignWith(cookieSecret.Reveal())
```

`RedactedString` 在 `String()` / `GoString()` / `MarshalJSON()` /
`MarshalText()` / `Format(...)` 五个路径返回 `***`；底层字段 `v` 不导出，
反射也无法直接拿到明文。

## runtime-config allowlist 边界

`GET /api/v1/runtime-config`（spec D-2）只返回 §3.1.2 标记为可暴露的字段：

- `appVersion`、`defaultUiLanguage`、`analyticsEnabled`、`featureFlags`、可选
  `postHogPublicKey`。
- `featureFlags` 仅纳入 `feature-flags.yaml` / PostHog 中 `public: true` 的
  flag；`ai_fallback_model_enabled` 等 operator-only flag 永远不进 response。
- 任何 secret 字段绝对不能进 response；session 携带 `analytics_opt_in=false`
  时 `analyticsEnabled=false` 且不返回 `postHogPublicKey`。

OpenAPI schema 真理源在 [B2
`openapi-v1-contract`](../docs/spec/openapi-v1-contract/spec.md)；A4 在
`backend/internal/platform/config/runtime_config.go` 持有 builder 与字段
allowlist，B2 引用 A4 时不得反向修改 builder。

## 红线

- `config/feature-flags.local.yaml` 等本地覆盖文件默认 ignore，禁止 commit。
- 业务代码不得 `import "github.com/posthog/posthog-go"` 或前端 `posthog-js`；
  统一通过 `internal/platform/featureflag` 抽象。
- 新增 secret 字段必须默认 `RedactedString`，并通过 SecretSource 注入。
- prod fail-fast 缺失 secret 时退出码非 0；deployer 必须先补齐 secret 再恢复
  supervisor，避免 crash loop（后续 release workstream 按
  [engineering-roadmap S3](../docs/spec/engineering-roadmap/spec.md#64-s3--true-integration-and-release-gate)
  承接 runbook handoff）。
