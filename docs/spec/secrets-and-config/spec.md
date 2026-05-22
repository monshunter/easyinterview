# Secrets and Config Spec

> **版本**: 2.5
> **状态**: active
> **更新日期**: 2026-05-12

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将历史 A4 `secrets-and-config` 保留为当前 active Foundation spec（依赖 [A1 `repo-scaffold`](../repo-scaffold/spec.md)）。它承接 [ADR-Q6](../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md) 第 4 段「运维注入」与 [ADR-Q3](../engineering-roadmap/decisions/ADR-Q3-analytics-platform.md) 自托管 PostHog 的接入凭证落点，决定了：

- 后端 API / backend internal runner / 前端 dev / staging / prod（以及未来需要时的 CI）各类环境如何拿到自己需要的连接串、API key、AI provider registry、feature flag 状态；
- secrets / config 在仓库里如何 layered（默认值、env override、运行时 secret），不被 hardcode；
- feature flag 如何接入 PostHog 但不让业务代码直接 import PostHog SDK。

目标是：

1. **三层 config**：`config.yaml`（默认值，仓库版本化）→ `.env` / 环境变量（环境差异）→ runtime secret（敏感凭证）；任何业务模块只通过 `internal/platform/config` 包读取，不直接读 `os.Getenv`。
2. **secrets 抽象**：`SecretSource` 接口在 P0 仅实现 env-based provider；P1 以后可扩展到 K8s Secret / Vault / SOPS（ADR-Q4 已留接口）；业务代码只依赖接口，不依赖具体 provider。
3. **feature flag 抽象**：`FeatureFlagClient` 接口；本 spec 提供 `FileFlagProvider`（YAML 文件，dev / 单测默认）与 `PostHogFlagProvider`（指向自托管 PostHog 的 HTTP API）；ADR-Q3 已锁定 self-host PostHog。
4. **lint 红线**：在本地质量门禁中拒绝 `os.Getenv(...)` 出现在 `internal/<domain>/...` 包；secrets 文件名 (`*.secret.yaml`) 一律加入 `.gitignore`；提交前 hook 拦截已知敏感前缀（`AKIA*` / `sk-*`）。

本 spec 不实现具体业务模块的配置消费、不部署 PostHog（[F2 `analytics-funnel`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) / [E4](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 承接）、不锁定 secret 后端实现（P1 以后再决策）。

## 2 范围

### 2.1 In Scope

- **配置文件落点**：
  - `config/config.yaml`（默认值，所有环境共享）
  - `config/dev.yaml` / `config/staging.yaml` / `config/prod.yaml`（环境 override，不含 secrets）
  - `.env.example`（仓库版本化的占位模板，列出所有合法 env key）
  - `config/feature-flags.yaml`（FileFlagProvider 的本地源，dev 默认值）
- **Go 包**：`backend/internal/platform/config/`（loader + validator + redactor）；`backend/internal/platform/secrets/`（`SecretSource` 接口与 env provider）；`backend/internal/platform/featureflag/`（`FeatureFlagClient` 接口与 file / posthog provider）。
- **TS 包**：`frontend/src/lib/runtime-config/`（前端只读取 build-time 注入与运行时 `/api/v1/runtime-config` 端点；不直接读浏览器 env）。
- **配置字段表**：本 spec §3.1 锁定 P0 必备 env key、默认值、config path、secret/public 分类与 runtime-config 暴露规则，全部由 `internal/platform/config` validator 消费。
- **lint / hook**：
  - `make lint-config` 检查 `.env.example` 与代码 `Get*` 调用一致。
  - `scripts/git-hooks/pre-commit-secrets.sh` 拦截敏感前缀（与 [A1 `scripts/git-hooks/`](../repo-scaffold/spec.md#21-in-scope) 集成）。
  - golangci-lint 自定义规则：禁止 `os.Getenv` 出现在 `internal/<domain>/` 包。
- **API 端点契约**：`GET /api/v1/runtime-config`（前端获取 feature flag 状态 + 公开配置项；schema 由 [B2 `openapi-v1-contract`](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 收口）。

### 2.2 Out of Scope

- 真正部署 PostHog：归 [F2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) + [E4](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)；A2 默认本地栈只要求 no-op / file-backed dev mode 不阻塞启动。
- K8s Secret / Vault / SOPS 实施：归 P1 / E4；本 spec 仅锁接口。
- Build-time 注入工具链（Vite envsubst / esbuild defines）：归 [D1 `frontend-shell`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) + [A5 `ci-pipeline-baseline`](../engineering-roadmap/spec.md#51-当前已存在的-active-spec)。
- Auth / session 业务语义（magic link challenge 表、server-side sessions 表、cookie 生命周期、风控阈值）：归 [C1 `backend-auth`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 与 [ADR-Q1](../engineering-roadmap/decisions/ADR-Q1-auth.md)；本 spec 只登记 C1 运行所需 secret/env key，并保证它们进入红线与 redaction。
- 数据库 migration / schema：归 [B4](../engineering-roadmap/spec.md#51-当前已存在的-active-spec)。
- 业务模块的 config 消费现场：归各 C / D 域。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策（含 P0 必备 env key 字典）

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 三层 config 优先级 | `runtime secret > env var > config/{env}.yaml > config/config.yaml` | loader 在启动时按此顺序合并；缺失关键字段直接 fail-fast |
| D-2 | 前端运行时配置入口 | `GET /api/v1/runtime-config` 只返回 allowlisted public 字段（`defaultUiLanguage`、`featureFlags` 中 `public=true` 的 flag、`appVersion`、`analyticsEnabled`、可选 `postHogPublicKey`）；schema 由 B2 收口 | 前端 0 后端调用即可初始化；secret / operator-only flag 永远不出现在前端；若请求携带有效 session，必须合并 `user_settings.analytics_opt_in`，opt-out 时 `analyticsEnabled=false` 且不返回 `postHogPublicKey`；用户分桶 flag 以后走 protected `/api/v1/me/runtime-config` |
| D-3 | secrets 抽象 | `SecretSource` 接口 `Get(name) (string, error)`；P0 默认实现 `EnvSecretSource`（env var）；接口在 spec 锁定后不允许在 P1 升级时变更签名 | 后续 K8s Secret / Vault 切换零业务改动 |
| D-4 | feature flag 抽象 | `FeatureFlagClient` 接口 `IsEnabled(key, ctx) bool` + `Variant(key, ctx) string`；P0 默认 `FileFlagProvider`，prod 切 `PostHogFlagProvider` | ADR-Q3 锁定不依赖第三方 cloud；自托管 PostHog 为生产唯一实现 |
| D-5 | P0 必备 env key 与 config schema | 见下方 §3.1.1 / §3.1.2；任一新增必须递增本 spec 版本 + 同步 `.env.example` + lint 校验；AI provider registry 中引用的 env ref 也进入 `make lint-config` 字典校验 | 防止业务模块偷偷加 env；让 validator / runtime-config / redaction 共用同一真理源 |
| D-6 | secret 红线 | `*.secret.yaml` 默认 `.gitignore`；pre-commit hook 拦截 `AKIA[0-9A-Z]{16}` / `sk-[A-Za-z0-9]{20,}` / `xox[baprs]-[A-Za-z0-9-]+`；本地 gitleaks 复扫；远端 CI secret scan 仅在 A5 触发条件成立后再接入 | 阻断仓库内敏感凭证泄漏 |
| D-7 | 配置热加载 | feature flag 支持热加载（≤ 30s）；其它 config 字段在进程启动时读取，运行时不变；如需热加载，必须递增 spec | 避免业务围绕「config 变了吗」写复杂代码 |
| D-8 | Session cookie 字面量 | `ei_session`，由 [ADR-Q1 §3](../engineering-roadmap/decisions/ADR-Q1-auth.md#3-决策) 锁定；P0 不提供 env/config override | A4 只管理 `SESSION_COOKIE_SECRET` 等 secret，不允许环境差异改 cookie name 导致 B2 OpenAPI / C1 middleware / D1 fetch 口径分裂 |
| D-9 | 后台任务队列权重 | `async.queueWeights` 配置路径固定在 `config/config.yaml` / `config/{env}.yaml`，默认 `critical: 6` / `default: 3` / `low: 1`；P0 不额外增加 env key，backend-async-runner kernel 通过 typed config 读取 | ADR-Q2 的 queue priority 可由配置驱动，同时保持 env 字典稳定为 24 项 |
| D-14 | Runner kernel 时序（additive） | additive 新增 config-only 节点 `async.leaseTimeoutSeconds`(300) / `async.shutdownGraceSeconds`(10) / `async.reaperIntervalSeconds`(60) / `async.scanIntervalSeconds`(5)；不新增 env key（env 字典保持 24 项）；缺失或非正数 fail-fast，不得静默回退为代码常量 | 由 active [`backend-async-runner/001`](../backend-async-runner/spec.md) D-5 / D-8 / D-14 消费，作为 kernel lease loop / reaper / graceful shutdown 时序源 |
| D-10 | 上传基础配置 | `objectStorage.provider=minio|filesystem`、`upload.presignTTLSeconds` 默认 600、`upload.maxBytes.resume=10485760`、`upload.maxBytes.targetJobAttachment=10485760`、`upload.maxBytes.privacyExport=5242880` 均为 config-only path；P0 不新增 `UPLOAD_*` / `OBJECT_STORE_*` env key | backend-upload 通过 typed config 注入 provider / TTL / per-purpose size limit，继续复用现有 `OBJECT_STORAGE_*` secret/env 字典 |

#### 3.1.1 P0 必备 env key 字典（24 项）

| Key | 必填 | 默认值 | 用途 | Owner subspec |
|-----|------|--------|------|---------------|
| `APP_ENV` | 是 | `dev` | `dev` / `staging` / `prod`；驱动 `config/{env}.yaml` 加载 | A4 |
| `APP_LISTEN_ADDR` | 是 | `:8080` | API 进程 HTTP 监听 | A4 |
| `DATABASE_URL` | 是 | `postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable` | Postgres 连接串 | A4（A2 锁本地默认） |
| `REDIS_URL` | 是 | `redis://localhost:6379/0` | Redis 连接串 | A4 |
| `OBJECT_STORAGE_ENDPOINT` | 是 | `http://localhost:9000` | MinIO / S3 endpoint | A4 |
| `OBJECT_STORAGE_BUCKET` | 是 | `easyinterview-dev` | 默认 bucket | A4 |
| `OBJECT_STORAGE_ACCESS_KEY` | 是 | `dev-access-key` | secret，prod 必填 | A4 |
| `OBJECT_STORAGE_SECRET_KEY` | 是 | `dev-secret-key` | secret，prod 必填 | A4 |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | 条件 | `(空，默认不上报)` | 配置了可选观测或生产 OTel Collector 时填写 OTLP HTTP endpoint | A4（F1 复用） |
| `LOG_LEVEL` | 是 | `info` | `debug/info/warn/error` | A4 |
| `SESSION_COOKIE_SECRET` | prod 必填 | `(空，dev 由 init 生成)` | secret；用于 `ei_session` first-party session cookie 签名 / 加密，server-side `sessions` 表仍是会话真理源 | A4（C1 owner，ADR-Q1） |
| `AUTH_CHALLENGE_TOKEN_PEPPER` | prod 必填 | `(空，dev 由 init 生成)` | secret；用于 magic link challenge token hash pepper，原始 token 不入库 / 不进日志 | A4（C1 owner，ADR-Q1） |
| `AI_PROVIDER_REGISTRY_PATH` | 是 | `config/ai-providers.yaml` | AI provider registry 文件路径；registry 内声明 provider ref、protocol、capabilities 与 secret env ref | A4（A3 owner） |
| `AI_PROVIDER_BASE_URL` | 条件 | `(空；仅默认 provider ref 引用时需要)` | 默认 OpenAI-compatible provider ref 可引用的 base URL env；不再代表全局唯一 AI provider contract | A4（A3 owner） |
| `AI_PROVIDER_API_KEY` | 条件 | `(空；仅默认 provider ref 引用时需要)` | 默认 OpenAI-compatible provider ref 可引用的 API key env；非 test 环境中被选中 provider 缺 secret 时 fail-fast | A4（A3 owner） |
| `DOUBAO_SPEECH_BASE_URL` | 条件 | `(空；仅 doubao_speech provider 被选中时需要)` | 豆包语音 provider-specific base URL | A4（A3 owner） |
| `DOUBAO_SPEECH_API_KEY` | 条件 | `(空；仅 doubao_speech provider 被选中时需要)` | secret | A4（A3 owner） |
| `MINIMAX_SPEECH_BASE_URL` | 条件 | `(空；仅 minimax_speech provider 被选中时需要)` | MiniMax 语音 provider-specific base URL | A4（A3 owner） |
| `MINIMAX_SPEECH_API_KEY` | 条件 | `(空；仅 minimax_speech provider 被选中时需要)` | secret | A4（A3 owner） |
| `AI_MODEL_PROFILE_PATH` | 是 | `config/ai-profiles.yaml` | Model Profile catalog 文件路径 | A4（A3 owner） |
| `FEATURE_FLAG_SOURCE` | 是 | `file` | `file` 或 `posthog` | A4 |
| `FEATURE_FLAG_FILE_PATH` | 条件 | `config/feature-flags.yaml` | `FEATURE_FLAG_SOURCE=file` 时必填 | A4 |
| `POSTHOG_HOST` | 条件 | `(空)` | `FEATURE_FLAG_SOURCE=posthog` 时必填；指向自托管 PostHog；普通本地 dev 默认不填 | A4（F2 owner） |
| `POSTHOG_SELF_HOSTED` | 条件 | `false` | staging / prod 使用 PostHog 时必须为 `true`；防止误接 PostHog Cloud | A4（F2 owner） |
| `POSTHOG_PROJECT_API_KEY` | 条件 | `(空)` | secret | A4（F2 owner） |
| `POSTHOG_PUBLIC_KEY` | 条件 | `(空，dev 占位)` | 暴露给前端的 public key；仅前端 analytics 初始化需要 | A4（F2 owner） |
| `EMAIL_PROVIDER` | prod 必填 | `(空)` | passwordless magic link 发件方 | A4（C1 owner，ADR-Q1） |
| `EMAIL_PROVIDER_API_KEY` | prod 必填 | `(空)` | secret | A4（C1 owner） |

#### 3.1.2 Canonical config schema 分类

| Config path | Env key(s) | Secret | Required rule | Runtime-config exposure | Owner |
|-------------|------------|--------|---------------|-------------------------|-------|
| `app.env` | `APP_ENV` | 否 | always | 否 | A4 |
| `app.listenAddr` | `APP_LISTEN_ADDR` | 否 | always | 否 | A4 |
| `runtime.appVersion` / `runtime.defaultUiLanguage` | build metadata / `config.yaml` | 否 | always | 是，固定 allowlist | A4 + D1 |
| `database.url` | `DATABASE_URL` | 是 | always | 否 | A4 |
| `redis.url` | `REDIS_URL` | 是 | always | 否 | A4 |
| `objectStorage.endpoint` / `objectStorage.bucket` | `OBJECT_STORAGE_ENDPOINT` / `OBJECT_STORAGE_BUCKET` | 否 | always | 否 | A4 |
| `objectStorage.accessKey` / `objectStorage.secretKey` | `OBJECT_STORAGE_ACCESS_KEY` / `OBJECT_STORAGE_SECRET_KEY` | 是 | always；prod 必须来自 runtime secret / env | 否 | A4 |
| `objectStorage.provider` | `(config.yaml only)` | 否 | always；`minio|filesystem` | 否 | A4 + backend-upload |
| `upload.presignTTLSeconds` | `(config.yaml only)` | 否 | always；正整数，默认 600 | 否 | A4 + backend-upload |
| `upload.maxBytes.resume` / `upload.maxBytes.targetJobAttachment` / `upload.maxBytes.privacyExport` | `(config.yaml only)` | 否 | always；正整数，默认 10MB / 10MB / 5MB | 否 | A4 + backend-upload |
| `observability.otlpEndpoint` | `OTEL_EXPORTER_OTLP_ENDPOINT` | 否 | optional | 否 | A4 + F1 |
| `log.level` | `LOG_LEVEL` | 否 | always | 否 | A4 |
| `auth.sessionCookieSecret` | `SESSION_COOKIE_SECRET` | 是 | prod required；dev init generated | 否 | A4 + C1 |
| `auth.sessionCookieName` | `(无 env key；ADR-Q1 固定)` | 否 | fixed literal `ei_session` | 否 | ADR-Q1 + A4 + C1 |
| `auth.challengeTokenPepper` | `AUTH_CHALLENGE_TOKEN_PEPPER` | 是 | prod required；dev init generated | 否 | A4 + C1 |
| `email.provider` / `email.providerApiKey` | `EMAIL_PROVIDER` / `EMAIL_PROVIDER_API_KEY` | provider 否；apiKey 是 | prod required | 否 | A4 + C1 |
| `ai.providerRegistryPath` | `AI_PROVIDER_REGISTRY_PATH` | 否 | always | 否 | A4 + A3 |
| `ai.defaultProviderBaseURL` / `ai.defaultProviderApiKey` | `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` | baseURL 否；apiKey 是 | required only when provider registry references these env names and the corresponding AIClient-enabled component starts；`APP_ENV=test` may use stub | 否 | A4 + A3 |
| `ai.doubaoSpeechBaseURL` / `ai.doubaoSpeechApiKey` | `DOUBAO_SPEECH_BASE_URL` / `DOUBAO_SPEECH_API_KEY` | baseURL 否；apiKey 是 | required only when doubao_speech provider is selected；`APP_ENV=test` may use stub | 否 | A4 + A3 |
| `ai.minimaxSpeechBaseURL` / `ai.minimaxSpeechApiKey` | `MINIMAX_SPEECH_BASE_URL` / `MINIMAX_SPEECH_API_KEY` | baseURL 否；apiKey 是 | required only when minimax_speech provider is selected；`APP_ENV=test` may use stub | 否 | A4 + A3 |
| `ai.modelProfilePath` | `AI_MODEL_PROFILE_PATH` | 否 | always | 否 | A4 + A3 |
| `featureFlag.source` / `featureFlag.filePath` | `FEATURE_FLAG_SOURCE` / `FEATURE_FLAG_FILE_PATH` | 否 | always；filePath required when source=file | 否 | A4 |
| `featureFlag.posthogHost` / `featureFlag.posthogSelfHosted` | `POSTHOG_HOST` / `POSTHOG_SELF_HOSTED` | 否 | required when source=posthog; staging/prod must self-host | 否 | A4 + F2 |
| `featureFlag.posthogProjectApiKey` | `POSTHOG_PROJECT_API_KEY` | 是 | required when source=posthog | 否 | A4 + F2 |
| `featureFlag.posthogPublicKey` | `POSTHOG_PUBLIC_KEY` | 否 | optional | 是，仅当 `analyticsEnabled=true` 且已配置 | A4 + F2 + D1 |
| `async.queueWeights` | `(config.yaml only)` | 否 | always；默认 `critical:6/default:3/low:1` | 否 | A4 + backend-async-runner + ADR-Q2 |
| `async.leaseTimeoutSeconds` / `async.shutdownGraceSeconds` / `async.reaperIntervalSeconds` / `async.scanIntervalSeconds` | `(config.yaml only)` | 否 | always；默认 `300/10/60/5`；缺失或非正数 fail-fast | 否 | A4 + active backend-async-runner (D-14) |

### 3.2 待确认事项

- 是否在 P0 引入 `viper` / `koanf` 等成熟 config 库 vs 手写最小 loader：默认 `koanf`（轻量、无 magic），由 001-bootstrap plan 落地时回填。
- `runtime-config` 端点是否需要鉴权：默认开放（仅返回 §3.1.2 allowlist 中标记为可暴露的字段）；如未来加入按用户分桶 feature flag，必须新增带 session 的 `/api/v1/me/runtime-config`，不得扩大 public endpoint。
- secret rotation：默认 P0 不实现自动 rotation；ADR / runbook 由后续 release gate / 运维阶段落地。

## 4 设计约束

### 4.1 边界约束

- `os.Getenv` 与 `flag.String` 等系统级读取只允许出现在 `backend/internal/platform/config/` 与 `backend/cmd/{api,migrate}/main.go` 中；其它包必须通过 `config.Get*` 注入；A5 接入 lint 强制。`cmd/migrate` 作为 B4 db-migrations-baseline 引入的 CLI 入口与 cmd/api 同列，bootstrap 期允许直接读取 `DATABASE_URL` / `APP_ENV` / `MIGRATE_*` 等 env key。P0 不保留 `cmd/worker` 入口或 worker listen addr。
- 前端任何代码不得直接读取 `import.meta.env.VITE_*` 之外的 build-time 变量；运行时配置统一通过 `runtime-config` 端点。
- `config/feature-flags.yaml` 是 dev / 单测真理源；prod 走 PostHogFlagProvider；切换由 `FEATURE_FLAG_SOURCE` 决定。
- PostHog provider 启动时必须校验 `FEATURE_FLAG_SOURCE=posthog` 时 `POSTHOG_HOST` / `POSTHOG_PROJECT_API_KEY` 存在，且 staging/prod `POSTHOG_SELF_HOSTED=true`；启动后 PostHog 临时不可用时只允许回退到 last-known-good 内存缓存并输出 warn，不允许静默切回 file provider 造成 prod flag 口径漂移。
- `runtime-config` 只能序列化 §3.1.2 标记为可暴露的字段；`featureFlags` 只包含 `config/feature-flags.yaml` 或 PostHog 中显式标记 `public=true` 的 flag，`ai_fallback_model_enabled` 等 operator-only flag 不得进入 public response。

### 4.2 安全约束

- 任何 secret 字段的字符串值在 log 中必须 redact（`config.RedactedString` 类型在 `String()` 方法返回 `***`）；A4 提供该类型，其它包必须使用。
- redaction 必须覆盖结构化 JSON 日志、error wrapping 与 nested config dump；`RedactedString` 明文只能通过显式 `Reveal()` 交给需要 secret 的底层 client，普通 `%s` / `%v` / JSON marshal 输出均为 `***`。
- `.gitignore` 必须包含：`*.secret.yaml`、`*.secret.json`、`config/local.*.yaml`、`.env`、`.env.local`。
- pre-commit hook 与本地 gitleaks 双重防护；远端 CI secret scan 仅在 A5 触发条件成立后再接入。

### 4.3 文档约束

- `.env.example` 必须列出全部 P0 必备 key（即 §3.1.1 字典），缺失即 lint 失败。
- `config/README.md` 解释三层优先级、各文件用途、新增 key 的流程。
- 本 spec 修订（新增 / 删除 / 重命名 env key）必须递增版本 + history。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `internal/platform/config/` | A4 | loader / validator / redactor / `Get*` API |
| `internal/platform/secrets/` | A4 | `SecretSource` 接口 + env provider；P1 以后扩展 K8s/Vault |
| `internal/platform/featureflag/` | A4 | `FeatureFlagClient` + file / posthog provider |
| `frontend/src/lib/runtime-config/` | A4 + D1 | `runtime-config` fetcher 与本地缓存；A4 锁字段，D1 集成 React hooks |
| `config/*.yaml` 内容 | 各业务 owner 增量 | A4 锁文件位置与 schema，业务字段由各 child 在 spec 修订时新增 |
| `config/feature-flags.yaml` 字段集 | F2 + 各业务 owner | A4 锁文件位置；当前 6 项 baseline flag 为 `practice_hint_enabled` / `report_evidence_v2_enabled` / `report_retry_plan_enabled` / `readiness_signals_enabled` / `ai_fallback_model_enabled` / `practice_assistance_mode_enabled`；旧 `mistake_book_export_enabled` / `growth_dashboard_v1_enabled` / `mock_session_dual_track_enabled` 已按 product-scope v1.2 删除 |
| AI provider registry / env keys 默认值 | A3（决策） + A4（落 env 字典） | A3 决定 provider registry 与 profile schema；A4 写进 env/config 字典并负责被选中 provider secret 缺失 fail-fast |
| Auth / Email env keys | C1 + A4 | C1 决定字段名（ADR-Q1），A4 写进字典 |
| 部署侧 secret 注入 | E4 + 运维 | A4 提供接口，E4 提供 K8s Secret / Vault 路径 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 三层合并 | `config/config.yaml` 设默认值；`config/dev.yaml` 覆盖 `LOG_LEVEL`；env 设 `APP_LISTEN_ADDR=:9090` | 启动 API 进程 | `config.Get("app.listenAddr") == ":9090"`；`config.Get("log.level") == "debug"`（dev override）；其它字段保持默认 | A4 后续 001 |
| C-2 | 缺失关键字段 fail-fast | `prod` 模式启动但 `SESSION_COOKIE_SECRET` 未设置 | `make build && APP_ENV=prod ./bin/api` | 启动进程退出码非 0，stderr 输出 `missing required secret: SESSION_COOKIE_SECRET`；不得回退到 dev init secret | A4 后续 001 |
| C-3 | feature flag file 模式 | `FEATURE_FLAG_SOURCE=file`；`config/feature-flags.yaml` 设 `practice_hint_enabled: true` | 业务调用 `featureflag.IsEnabled("practice_hint_enabled", ctx)` | 返回 `true`；修改 YAML 后 ≤ 30s 内自动热加载 | A4 后续 001 |
| C-4 | feature flag posthog 模式 | `FEATURE_FLAG_SOURCE=posthog`，`POSTHOG_HOST` 指向 mock，`POSTHOG_SELF_HOSTED=true` | 调用 `IsEnabled` | client 出站 HTTP 命中 PostHog `/decide` 端点；client 不直接 import PostHog SDK；staging/prod 若 `POSTHOG_SELF_HOSTED=false` 则启动失败；PostHog 临时不可用时返回 last-known-good 缓存并写 warn，不静默切 file provider | A4 后续 001 |
| C-5 | secret redact | log / error wrapping / JSON dump 中输出 `config.Get("objectStorage.secretKey")` | 进程产生日志 | 日志中显示 `***`；不出现明文 secret | A4 后续 001 |
| C-6 | runtime-config 端点 | 前端首屏加载，`practice_hint_enabled.public=true`、`ai_fallback_model_enabled.public=false`，且当前用户 `analytics_opt_in=false` | `GET /api/v1/runtime-config` | 返回 `{appVersion, defaultUiLanguage, analyticsEnabled:false, featureFlags{practice_hint_enabled: ...}}`；不返回任何 secret，不返回 operator-only flag，不返回 `postHogPublicKey` | A4 + B2 + D1 |
| C-7 | lint 红线 | 本地改动在 `internal/auth/` 下出现 `os.Getenv("SESSION_COOKIE_SECRET")` | `make lint` | 报错并阻止本地质量门禁通过 | A4 后续 001 |
| C-8 | secrets 红线 | 本地改动包含一行形似真实凭证的测试样本（例如 `OPENAI_API_KEY=<redacted-test-token>`；测试文件通过临时生成内容触发正则，不在文档中写真实形态） | pre-commit / 本地 gitleaks | hook 拦截，gitleaks 拦截；远端 CI secret scan 仅在 A5 触发条件成立后再接入 | A4 后续 001 |
| C-12 | 后台队列权重配置 | `config/config.yaml` 声明 `async.queueWeights`，dev/staging/prod override 可调整权重 | backend internal runner 初始化读取 typed config | 读取到 `critical/default/low` 三档权重，缺失或非正数 fail-fast；不需要新增 env key | A4 后续 001 + backend-runtime-topology |
| C-13 | 上传基础 config-only path | `config/config.yaml` 声明 `objectStorage.provider`、`upload.presignTTLSeconds`、`upload.maxBytes.*` | backend-upload handler / objectstore 初始化读取 typed config | 默认值可读取；非法 provider、非正数 TTL / maxBytes fail-fast；`.env.example` 不新增 `UPLOAD_*` / `OBJECT_STORE_*` | backend-upload/001 |
| C-9 | env 字典覆盖 | `.env.example` 中缺 `AI_PROVIDER_REGISTRY_PATH` 或 registry 引用的 provider secret env | `make lint-config` | 报错：env key 在代码或 registry truth source 出现但 `.env.example` 缺失 | A4 后续 001 + A3 003 |
| C-10 | AI provider 本地部署校验 | `APP_ENV=dev` 且启用了需要 AIClient 的 backend 运行路径，但 provider registry 缺失或选中 provider secret 缺失 | docker compose / Kind 启动进程 | 进程启动失败并报告缺失 provider registry / secret；`APP_ENV=test` 的单元测试仍可走 stub | A4 后续 001 + A3 003 + A2 |
| C-11 | config schema 分类 | `SESSION_COOKIE_SECRET` 标记为 secret，`runtime.defaultUiLanguage` 标记为 public | `make lint-config` / runtime-config schema check | secret 字段缺 redaction 或出现在 runtime-config schema 时失败；public 字段缺 runtime-config schema 时失败 | A4 后续 001 |

## 7 关联计划

A4 当前暂无 active impl plan；后续由 A4 自身的 `001-bootstrap` 承接：

- 落地 `internal/platform/{config,secrets,featureflag}/` Go 包与默认 provider。
- 落地 `config/*.yaml`、`.env.example`、`config/feature-flags.yaml`。
- 落地 `async.queueWeights` typed config，并由 backend internal runner 在后续 plan 中消费。
- 落地 lint 规则与 pre-commit hook（接入 A1 `scripts/git-hooks/`）。
- 提供 `frontend/src/lib/runtime-config/` 与最小 fetcher。

后续 P1 升级（K8s Secret / Vault / SOPS provider）由本 spec 修订递增版本后追加 plan，不创建 sibling spec。
