# Secrets and Config Bootstrap

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-05-05

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [secrets-and-config spec](../../spec.md) §3.1 已锁定的 D-1..D-9 决策与 §3.1.1 / §3.1.2 锁定的 24 项 P0 必备 env key 字典、canonical config schema、`async.queueWeights` config-only 字段落到代码：建立 `backend/internal/platform/{config,secrets,featureflag}/` 三个 Go 包真理源、`config/*.yaml` + `.env.example` + `config/feature-flags.yaml` 默认值集合、`make lint-config` 与 `scripts/git-hooks/pre-commit-secrets.sh` 本地质量门禁、`frontend/src/lib/runtime-config/` 前端 fetcher，以及最小 `GET /api/v1/runtime-config` handler stub，并通过本 plan 的 verification phase 证明 [secrets-and-config spec §6](../../spec.md#6-验收标准) C-1..C-12 在本仓库可重复跑通（C-6 与 [B2 `openapi-v1-contract`](../../../openapi-v1-contract/spec.md) 共担最终 schema 一致性，本 plan 完成 A4 侧 builder 与 stub）。

本 plan 是 `secrets-and-config` 唯一的 plan；后续若需扩展（K8s Secret / Vault / SOPS provider，自动 secret rotation，分桶 feature flag），按 §7 约束递增 spec 与本 plan 版本，原地修订，不再开 sibling plan。

## 2 背景

[engineering-roadmap §5.1](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 A4 保留为当前 active Foundation spec；后续 [B2 `openapi-v1-contract`](../../../openapi-v1-contract/spec.md)、[backend-auth](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)、[frontend-shell](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 等 workstream 依赖本计划输出的配置 / secret / feature flag 契约。本 plan 通过 §4 的 6 个 phase 验收 [secrets-and-config spec §6](../../spec.md#6-验收标准) C-1..C-11，关闭 [001-decompose-subspecs](../../../engineering-roadmap/plans/001-decompose-subspecs/checklist.md) 保留的 A4 bootstrap 承诺。

执行本 plan 前必须确认：

- [A1 `repo-scaffold/001-bootstrap`](../../../repo-scaffold/plans/001-bootstrap/plan.md) 已创建根 `Makefile`、`backend/`、`frontend/`、`scripts/git-hooks/`、`.gitignore` 等容器目录与基础 hook 入口；本 plan 只在其上扩展。
- [B1 `shared-conventions-codified/001-bootstrap`](../../../shared-conventions-codified/plans/001-bootstrap/plan.md) 已落地 `backend/go.mod` 与 `backend/internal/shared/` 共享包；本 plan 的 Go 代码引用其常量。
- [A2 `local-dev-stack/001-bootstrap`](../../../local-dev-stack/plans/001-bootstrap/plan.md) 的 `deploy/dev-stack/.env.example` 字段名已与 spec §3.1.1 字典对齐；本 plan 的 `.env.example` 是仓库根真理源，A2 dev stack 在本地启动时复用同一字典。

每个 phase 是可独立部署 / 验证的纵向行为切片：Phase 1 起来即可由 Go 代码 `config.Get*` 读取三层合并的配置；Phase 2 起来即可由业务代码通过 `SecretSource` / `FeatureFlagClient` 接口隔离 provider；Phase 3 起来即有完整的 `.env.example` 与 `config/*.yaml` 字典；Phase 4 起来即有 `make lint-config` 与 pre-commit secret 拦截；Phase 5 起来即有最小 `runtime-config` 端到端链路；Phase 6 收口验证 C-1..C-11 并完成 handoff。

本 plan 不部署 PostHog（归 [F2 `analytics-funnel`](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 与 [E4 `release-gate-and-rollout`](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)），不实现 K8s Secret / Vault / SOPS provider（归 P1 / E4），不冻结 `/api/v1/runtime-config` 的 OpenAPI schema（归 [B2 `openapi-v1-contract`](../../../openapi-v1-contract/spec.md)；A4 在本 plan 中只交付 response builder + 最小 stub handler）。

## 3 质量门禁分类

- **Plan 类型**: `platform-config + code-internal + contract + tooling`。本 plan 修改 backend config/secrets/featureflag packages、config truth source、secret lint hooks、runtime-config builder/stub、frontend runtime-config fetcher 和本地 lint gate；不直接交付用户可见 workflow。
- **TDD 策略**: 历史实现以 checklist 每项的 Go tests、TS tests、lint negative cases、pre-commit secret redline、runtime-config allowlist tests 和 config fail-fast smoke 作为 Red-Green-Refactor 断言来源；重进本 plan 时必须通过 `/implement` -> `/tdd` 顺序执行。
- **BDD 策略**: BDD 不适用。本 plan 是内部配置/secret/feature flag contract 与 tooling；后续 D1/B2/C workstream 把 runtime-config 暴露到用户流程时维护自身 BDD gate。
- **替代验证 gate**: `go test ./backend/internal/platform/config/... ./backend/internal/platform/secrets/... ./backend/internal/platform/featureflag/...`、frontend runtime-config tests/typecheck、`make lint-config`、secret hook negative tests、`make lint`、runtime-config allowlist smoke、`sync-doc-index --check`。

## 4 实施步骤

### Phase 1: Three-tier config loader 与 redactor

#### 1.1 落地 `backend/internal/platform/config/` 包骨架

按 [secrets-and-config spec §5](../../spec.md#5-模块边界) 把 `loader.go` / `validator.go` / `redactor.go` / `getters.go` / `doc.go` 落到 `backend/internal/platform/config/`，module path 沿用 [B1 §1.2](../../../shared-conventions-codified/plans/001-bootstrap/plan.md#12-go-module-初始化) 锁定的 `github.com/monshunter/easyinterview/backend`。`doc.go` 用一段 godoc 概述说明三层优先级（与 [secrets-and-config spec §3.1 D-1](../../spec.md#31-已锁定决策含-p0-必备-env-key-字典) 对齐）以及对外可见的 `Get*` API 命名约定。

#### 1.2 接入 `koanf` 作为 loader 实现

按 [secrets-and-config spec §3.2](../../spec.md#32-待确认事项) 待确认项的默认决议引入 `github.com/knadh/koanf/v2` + `koanf/parsers/yaml` + `koanf/providers/env` + `koanf/providers/file`；不引入 `viper`。`loader.go` 中按 D-1 顺序合并：先 `config/config.yaml`（默认值），再 `config/{APP_ENV}.yaml`（环境 override），再 env var，最后 runtime secret（通过 `SecretSource` 注入）。`koanf` 默认使用最后写入覆盖前者，必须在 loader 中显式约定合并顺序，禁止在多 provider 并行调用 `Load`，避免并发 merge 顺序歧义。

#### 1.3 落地 `Get*` API 与类型化访问器

`getters.go` 暴露 `GetString(key string) string` / `GetInt(key string) int` / `GetBool(key string) bool` / `GetDuration(key string) time.Duration` / `GetSecret(key string) RedactedString`；任何业务包通过这些 API 读取配置。错误路径（缺失 required key / 类型不符）由 validator 统一处理后返回，不由 getter 自行 panic。getter 必须以 `app.listenAddr` 形式接受点路径，与 [secrets-and-config spec §3.1.2](../../spec.md#312-canonical-config-schema-分类) 中的 `Config path` 列对齐。

#### 1.4 实现 `RedactedString` 类型与 redactor

按 [secrets-and-config spec §4.2](../../spec.md#42-安全约束) 在 `redactor.go` 中实现 `RedactedString` 类型：底层为 `string`；`String()` / `GoString()` / `MarshalJSON()` / `MarshalText()` / `Format(...)` 均输出 `***`；只有显式 `Reveal() string` 返回明文。redactor 必须覆盖结构化 JSON 日志、`fmt.Errorf("...: %w", err)` 包装链、嵌套 config dump（递归到 struct 字段时仍返回 `***`）三种路径。Phase 6 自检中通过 `errors_test.go` / `redactor_test.go` 强制这三条路径。

#### 1.5 启动期 fail-fast validator

`validator.go` 在进程启动期消费 spec §3.1.2 标记为 `required` 的字段集合，按 `APP_ENV` 维度判定：`APP_ENV=test` 允许缺 AI / Email / Session 相关 secret；`APP_ENV=staging|prod` 必填 secret 缺失时返回 `error`，由 `cmd/{api,worker}/main.go` 接住后非零退出。错误信息必须明确列出缺失 key 名（与 [secrets-and-config spec §6 C-2](../../spec.md#6-验收标准) 一致），避免 prod 退出后 deployer 不知道补哪个 key。

### Phase 2: SecretSource + FeatureFlagClient 抽象与 provider 实现

#### 2.1 落地 `backend/internal/platform/secrets/` 包

按 [secrets-and-config spec D-3](../../spec.md#31-已锁定决策含-p0-必备-env-key-字典) 在 `secrets/secrets.go` 写入接口：

```go
type SecretSource interface {
    Get(name string) (string, error)
}
```

接口签名固定，P1 升级到 K8s Secret / Vault / SOPS 时只能新增 provider，不允许修改签名。`secrets/env_provider.go` 实现 `EnvSecretSource`：从 `os.Getenv` 读取（注意 `os.Getenv` 仅允许出现在 `internal/platform/config/`、`internal/platform/secrets/` 与 `cmd/{api,worker,migrate}/main.go`，由 Phase 4.1 lint 收口）。Phase 1 loader 把 `EnvSecretSource` 作为默认运行时 secret 来源，与 D-1 第一层（`runtime secret > env var ...`）一致。

#### 2.2 落地 `backend/internal/platform/featureflag/` 包接口

按 [secrets-and-config spec D-4](../../spec.md#31-已锁定决策含-p0-必备-env-key-字典) 在 `featureflag/featureflag.go` 写入接口：

```go
type FeatureFlagClient interface {
    IsEnabled(key string, ctx FlagContext) bool
    Variant(key string, ctx FlagContext) string
}
```

`FlagContext` 仅承载 anonymous distinct id / authenticated user public id / app env 三类字段；任何业务包只调用此接口，绝对不允许直接 import `github.com/posthog/posthog-go` 或前端 `posthog-js`，由 Phase 4.1 lint 收口。

#### 2.3 落地 `FileFlagProvider`（YAML，hot reload ≤30s）

`featureflag/file_provider.go` 读取 `config/feature-flags.yaml`，按 [secrets-and-config spec D-7](../../spec.md#31-已锁定决策含-p0-必备-env-key-字典) 实现 ≤ 30s 热加载：使用 `time.Ticker` 定期对比文件 mtime + 解析后的内容 hash，变更则原子替换内部 map。热加载必须避免在进程刚启动且文件锁未稳定时 race（Phase 6 自检覆盖）；加载失败时保留上一次的内存快照并写一条结构化 warn 日志，禁止 panic。文件 schema 以 `flags: { <key>: { enabled: bool, variant?: string, public: bool } }` 为最小集合，与 `engineering-roadmap decisions §15.1` 列出的 6 项 P0 baseline flag 兼容。

#### 2.4 落地 `PostHogFlagProvider`（自托管 PostHog 原生 HTTP）

`featureflag/posthog_provider.go` 仅使用 `net/http` 直接调用自托管 PostHog 的 `POST {POSTHOG_HOST}/decide?v=3` 端点（与 [ADR-Q3 §3](../../../engineering-roadmap/decisions/ADR-Q3-analytics-platform.md#3-决策) 决议一致），request body 仅携带 `api_key` / `distinct_id` / `groups` / `person_properties`，禁止 import PostHog SDK。staging / prod 启动时如果 `POSTHOG_SELF_HOSTED=false`，validator 必须 fail-fast，与 [secrets-and-config spec §6 C-4](../../spec.md#6-验收标准) 一致；本 plan 不实现 SDK fallback。响应中 `featureFlags` map 缓存 ≤ 30s（与 D-7 热加载 SLA 对齐）；PostHog 网络错误或 5xx 时返回 last-known-good 内存快照并写 warn，若没有可用快照则返回 error / degraded，不静默切回 `FileFlagProvider` 读取 dev baseline，避免 prod flag 口径漂移。

#### 2.5 Phase 2 自检

`secrets_test.go` 与 `featureflag_test.go` 覆盖：`EnvSecretSource.Get` 缺失返回 error；`FileFlagProvider` 解析 YAML、修改 mtime 后 ≤30s 热加载；`PostHogFlagProvider` 在 mock HTTP server（`httptest.NewServer`）返回 `decide` 模拟响应时 `IsEnabled("practice_hint_enabled", ctx)` 返回 true；mock 5xx / timeout 时命中 last-known-good 快照；无快照时返回 degraded/error；`POSTHOG_SELF_HOSTED=false` 在 staging/prod 启动时 validator 返回 fail-fast error。Phase 6 自检会再次串行复跑这些用例。

### Phase 3: 配置文件骨架与 .env.example 字典对齐

#### 3.1 落地 `config/config.yaml` 默认值

按 [secrets-and-config spec §3.1.2](../../spec.md#312-canonical-config-schema-分类) 锁定的 canonical config schema 写入仓库根 `config/config.yaml`（D-1 第一层默认值）。所有 secret 字段（`database.url` / `redis.url` / `objectStorage.accessKey` / `objectStorage.secretKey` / `auth.sessionCookieSecret` / `auth.challengeTokenPepper` / `email.providerApiKey` / `ai.providerApiKey` / `featureFlag.posthogProjectApiKey`）必须留空字符串占位，禁止写入真实凭证；明文字段（如 `runtime.appVersion` / `runtime.defaultUiLanguage` / `app.listenAddr` / `featureFlag.source`）写入 spec 默认值。`auth.sessionCookieName` 固定字面量 `ei_session`，与 [ADR-Q1 §3](../../../engineering-roadmap/decisions/ADR-Q1-auth.md#3-决策) 与 spec D-8 一致；`async.queueWeights` 默认 `critical: 6` / `default: 3` / `low: 1`，不提供 env override；本 plan 不允许任何 env key 覆盖 `ei_session`。

#### 3.2 落地 `config/{dev,staging,prod}.yaml` 环境 override

按 [secrets-and-config spec §2.1](../../spec.md#21-in-scope) 写入三份 env override 文件，禁止包含任何 secret 明文。`dev.yaml` 设 `log.level: debug`、`featureFlag.source: file`；`staging.yaml` 与 `prod.yaml` 设 `featureFlag.source: posthog` 与 `featureFlag.posthogSelfHosted: true`，与 [ADR-Q3](../../../engineering-roadmap/decisions/ADR-Q3-analytics-platform.md) 一致；任何运行时 secret 字段在三份文件中都留空，提示运维通过 env / 真正 secret 注入。

#### 3.3 落地 `.env.example` 与 env key 字典

按 [secrets-and-config spec §3.1.1](../../spec.md#311-p0-必备-env-key-字典25-项) 写入仓库根 `.env.example`，包含全部 env key（`APP_ENV` / `APP_LISTEN_ADDR` / `WORKER_LISTEN_ADDR` / `DATABASE_URL` / `REDIS_URL` / `OBJECT_STORAGE_*` / `OTEL_EXPORTER_OTLP_ENDPOINT` / `LOG_LEVEL` / `SESSION_COOKIE_SECRET` / `AUTH_CHALLENGE_TOKEN_PEPPER` / `AI_PROVIDER_*` / `AI_MODEL_PROFILE_PATH` / `FEATURE_FLAG_*` / `POSTHOG_*` / `EMAIL_PROVIDER` / `EMAIL_PROVIDER_API_KEY`）。所有 secret 字段只写占位说明（如 `# secret; populate via runtime secret in prod`），不允许写真实 key 样本。每行注释必须标注「Owner subspec」与「prod required: yes/no/conditional」，与 spec §3.1.1 表格一一对应。`async.queueWeights` 是 config-only 字段，不进入 `.env.example`。

#### 3.4 落地 `config/feature-flags.yaml` baseline

按 product-scope v1.2 / UI scope 写入 6 项 baseline flag（`practice_hint_enabled` / `report_evidence_v2_enabled` / `report_retry_plan_enabled` / `readiness_signals_enabled` / `ai_fallback_model_enabled` / `practice_assistance_mode_enabled`）。每个 flag 必须显式标注 `public: true|false`：除 `ai_fallback_model_enabled` 外均为 `public: true`（前端可见）；`ai_fallback_model_enabled` 设 `public: false`（operator-only），由 Phase 5 runtime-config builder 在 allowlist 中过滤。旧 `mistake_book_export_enabled` / `growth_dashboard_v1_enabled` / `mock_session_dual_track_enabled` 不得恢复。

#### 3.5 落地 `config/README.md`

按 [secrets-and-config spec §4.3](../../spec.md#43-文档约束) 写一份一屏长度 README，覆盖：三层优先级（D-1）、各文件用途（`config.yaml` / `dev.yaml` / `staging.yaml` / `prod.yaml` / `feature-flags.yaml`）、新增 env key 流程（先递增 spec 版本 + history → 同步 `.env.example` → 同步 §3.1.2 schema → 加 validator → 跑 `make lint-config`）、`RedactedString` 使用示范、`runtime-config` allowlist 边界。README 必须有 cross-link 回到 spec §3.1 / §4。

#### 3.6 .gitignore 红线扩展（与 A1 协作）

按 [secrets-and-config spec §4.2](../../spec.md#42-安全约束) 红线，本 plan 在 [A1 `repo-scaffold/001-bootstrap`](../../../repo-scaffold/plans/001-bootstrap/plan.md) 已创建的根 `.gitignore` 中追加：`*.secret.yaml`、`*.secret.json`、`config/local.*.yaml`、`.env`、`.env.local`、`config/feature-flags.local.yaml`。追加位置必须独立成段并打标注 `# A4 secrets-and-config red lines`，便于后续审查。本 plan 不重写 A1 已有的 hook 入口，只追加 secret 红线条目。

### Phase 4: Lint / pre-commit hook / `make lint-config`

#### 4.1 golangci-lint 自定义规则（拒绝 `os.Getenv` 越界）

按 [secrets-and-config spec §4.1](../../spec.md#41-边界约束) 落地：在 [B1](../../../shared-conventions-codified/plans/001-bootstrap/plan.md#31-go-lint-与错误码校验) 已落地的 `backend/.golangci.yml` 中追加一条本地可执行规则。优先选择 `revive` 自定义 rule（如 `disallow-direct-env-access`）；若 `revive` 表达不动，则在 `scripts/lint/` 下落 `getenv_boundary.go`（Go AST checker），由 `make lint-config` / `make lint` 调用 `go run scripts/lint/getenv_boundary.go -root backend` 扫描 `backend/...`。规则 allowlist 仅放行 `backend/internal/platform/config/`、`backend/internal/platform/secrets/`、`backend/cmd/api/`、`backend/cmd/worker/`、`backend/cmd/migrate/`，与当前 spec §4.1 一致；其它包出现 `os.Getenv` 必须 lint 失败（关闭 [secrets-and-config spec §6 C-7](../../spec.md#6-验收标准)）。

#### 4.2 `make lint-config`：env key dictionary drift 检查

按 [secrets-and-config spec §6 C-9](../../spec.md#6-验收标准) 落地：在 `scripts/lint/env_dict.py`（或 `scripts/lint/env_dict.sh`）实现一次性扫描器：

- 解析 `.env.example` 提取所有 env key 名。
- AST / regex 解析 `backend/internal/platform/config/` 与 `cmd/{api,worker}/main.go` 中所有 `os.Getenv("...")` / `Get*("...")` 调用，提取代码侧已声明的 env key。
- 解析 spec `§3.1.1` 表格（通过定位 markdown 表头与列分隔符）并提取「Key」列。
- 三方求差集：任一方缺失 key 必须 fail，错误信息列出缺失项与三方分别声明的差异。

`Makefile` 新增 `.PHONY: lint-config` target 并入 `make lint`：`lint-config` 失败时整个 `make lint` 失败。脚本只读不改文件，不能在缺 key 时静默自动补齐。

#### 4.3 pre-commit hook：拦截敏感前缀

按 [secrets-and-config spec D-6](../../spec.md#31-已锁定决策含-p0-必备-env-key-字典) 与 [§6 C-8](../../spec.md#6-验收标准) 落地 `scripts/git-hooks/pre-commit-secrets.sh`：扫描 `git diff --cached` 命中 `AKIA[0-9A-Z]{16}` / `sk-[A-Za-z0-9]{20,}` / `xox[baprs]-[A-Za-z0-9-]+` 任一正则即 fail（exit 1），错误信息列出文件名与行号，不输出命中的 secret 字面量本身（避免日志再泄漏一次）。脚本由 [A1](../../../repo-scaffold/plans/001-bootstrap/plan.md) 已建立的 `scripts/git-hooks/` 入口注册，本 plan 只追加文件并扩展安装脚本，不重写 A1 hook 框架。

#### 4.4 本地 gitleaks 第二层

按 [secrets-and-config spec D-6](../../spec.md#31-已锁定决策含-p0-必备-env-key-字典) 在 `scripts/lint/gitleaks.sh` 落入第二层扫描入口：调用本地已安装的 `gitleaks detect --no-git --redact`（不要求开发者必须装；脚本检测到未安装时打印安装提示并 exit 0，避免阻塞，但 README 标注推荐安装）。`make lint` 调用此脚本作为第二层。当前阶段不接入远端 CI secret scan，仅在 [A5 `ci-pipeline-baseline`](../../../ci-pipeline-baseline/spec.md) 触发条件成立后再评估。

#### 4.5 Phase 4 自检

构造一份故意越界改动跑 `make lint`：在 `backend/internal/auth/` 添加一行 `os.Getenv("SESSION_COOKIE_SECRET")` → `make lint` 必须失败并报 `os.Getenv outside platform/config`；删除 `.env.example` 中 `AI_PROVIDER_BASE_URL` → `make lint-config` 失败；在临时文件中程序化生成一行形似真实凭证的值并 `git add` → pre-commit hook 拦截。所有自检完成后必须把越界改动 revert，避免污染主分支；文档与 fixture 中不得长期保存命中 secret 正则的样本文本。

### Phase 5: `runtime-config` endpoint 接入与前端 fetcher

#### 5.1 后端 `runtime-config` builder 与最小 stub handler

按 [secrets-and-config spec D-2 / §3.1.2 / §6 C-6](../../spec.md#31-已锁定决策含-p0-必备-env-key-字典) 在 `backend/internal/platform/config/runtime_config.go` 落地 `BuildRuntimeConfig(ctx, session) RuntimeConfig`：

- 字段 allowlist 严格限定 `appVersion` / `defaultUiLanguage` / `analyticsEnabled` / `featureFlags` / `postHogPublicKey`；
- `featureFlags` 仅纳入 `config/feature-flags.yaml` 或 PostHog 中标 `public: true` 的 flag；`ai_fallback_model_enabled` 等 operator-only flag 必须被过滤；
- 若 `ctx` 携带有效 session 且 `user_settings.analytics_opt_in == false`，则 `analyticsEnabled = false` 且不返回 `postHogPublicKey`（与 D-2 一致）；
- 任何 secret 字段绝对不能进入 response。

`backend/internal/platform/config/runtime_config_handler.go` 落地最小 `GET /api/v1/runtime-config` HTTP handler stub：直接调用 `BuildRuntimeConfig`，序列化为 JSON。OpenAPI schema 真理源由 [B2 `openapi-v1-contract`](../../../openapi-v1-contract/spec.md) 持有；本 plan 仅交付 builder + stub handler，schema 与 fixture 一致性由 B2 在引用 A4 时验证。`user_settings` 真实接入由 [C1 `backend-auth`](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 后续完成；本 plan 用 in-memory fake 与 nil session 路径覆盖测试。

#### 5.2 前端 `frontend/src/lib/runtime-config/` fetcher

按 [secrets-and-config spec §2.1 / D-2 / §5](../../spec.md#21-in-scope) 在 `frontend/src/lib/runtime-config/` 落地：

- `index.ts` 暴露 `fetchRuntimeConfig(): Promise<RuntimeConfig>`：调用 `GET /api/v1/runtime-config`，缓存到 module-scoped `let cached` 直到下一个 page load。
- `types.ts` 与后端 `RuntimeConfig` 字段同名（`appVersion` / `defaultUiLanguage` / `analyticsEnabled` / `featureFlags: Record<string, FlagDecision>` / `postHogPublicKey?`）。
- `hooks.placeholder.ts` 留 `useRuntimeConfig()` 占位 React hook（仅类型签名，不导入 React），由 [D1 `frontend-shell`](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 在自身 plan 中实现完整 hook + provider；本 plan 只锁字段与 fetcher 行为。

前端任何代码不得直接读取 `import.meta.env.VITE_*` 之外的 build-time 变量，与 [secrets-and-config spec §4.1](../../spec.md#41-边界约束) 一致；运行时配置统一通过本 fetcher。

#### 5.3 contract handoff 边界与 cross-link

在 `backend/internal/platform/config/runtime_config.go` 与 `frontend/src/lib/runtime-config/index.ts` 顶部 godoc / TSDoc 注释中 cross-link 到 [B2 `openapi-v1-contract` spec](../../../openapi-v1-contract/spec.md)：明确「OpenAPI schema 真理源在 B2；A4 仅持有 builder 与字段 allowlist」。本 plan 不修改 B2 spec / plan / OpenAPI 文件；如果 B2 后续在 schema 中扩展字段，必须通过 spec 修订递增 A4 spec 版本后才允许扩展 builder allowlist，避免「先暗暗开洞、后补文档」的漂移。

#### 5.4 Phase 5 自检

构造下列单测覆盖 [secrets-and-config spec §6 C-6](../../spec.md#6-验收标准)：

- `runtime_config_test.go`：当 `practice_hint_enabled.public=true`、`ai_fallback_model_enabled.public=false`、`analytics_opt_in=false` 时，response 包含 `practice_hint_enabled` 但不含 `ai_fallback_model_enabled`，`analyticsEnabled=false`，无 `postHogPublicKey`，无任何 secret 字段。
- 前端 `runtime-config.test.ts`（vitest）：`fetchRuntimeConfig` mock 后返回字段正确解析为 TS 类型；缓存命中第二次调用不触发 `fetch`。
- 手工 smoke：本地 `go run ./backend/cmd/api` 启动后 `curl http://localhost:8080/api/v1/runtime-config` 返回 allowlist response。

### Phase 6: Verification + handoff

#### 6.1 AC C-1..C-5 验证（Phase 1 / Phase 2 关闭）

依次复跑：

- **C-1（三层合并）**：`config/config.yaml` 默认 `log.level: info`；`config/dev.yaml` 设 `log.level: debug`；env 设 `APP_LISTEN_ADDR=:9090`；`go run ./backend/cmd/api -dump-config` 必须显示 `app.listenAddr=:9090` 与 `log.level=debug`，其它字段保持默认。
- **C-2（fail-fast）**：`APP_ENV=prod ./bin/api` 缺 `SESSION_COOKIE_SECRET` 时退出码非 0 且 stderr 含 `missing required secret: SESSION_COOKIE_SECRET`。
- **C-3（file flag 热加载）**：`FEATURE_FLAG_SOURCE=file` 启动后修改 `config/feature-flags.yaml`，30s 内 `featureflag.IsEnabled("practice_hint_enabled", ctx)` 反映新值。
- **C-4（posthog flag）**：mock `/decide` server 命中后 `IsEnabled` 正确返回；mock 5xx / timeout 命中 last-known-good 缓存；无缓存时返回 degraded/error；`POSTHOG_SELF_HOSTED=false` 在 staging/prod 启动失败。
- **C-5（redact）**：`go test ./backend/internal/platform/config/...` 包含 `redactor_test.go`，断言 `RedactedString` 在 `fmt.Println` / JSON marshal / error wrapping 三种路径输出 `***`。

#### 6.2 AC C-7..C-11 验证（Phase 3 / Phase 4 / Phase 5 关闭）

依次复跑：

- **C-7（os.Getenv 红线）**：手动构造越界改动验证 `make lint` 失败，验证后立即 revert。
- **C-8（secrets 红线）**：通过临时文件程序化生成命中 `AKIA*` / `sk-*` / `xox*` 三类正则的假数据并 `git add`，验证 pre-commit hook 拦截、本地 gitleaks 二次拦截，验证后立即 revert；不在文档中保留真实形态样本。
- **C-9（env 字典覆盖）**：从 `.env.example` 暂时删除 `AI_PROVIDER_BASE_URL`，`make lint-config` 必须失败并指出代码侧已声明但 example 缺失，验证后恢复。
- **C-10（AI provider fail-fast）**：`APP_ENV=dev` 且加载 AIClient-enabled 组件但缺 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 时进程启动失败；`APP_ENV=test` 仍可走 stub。
- **C-11（schema 分类）**：`make lint-config` 校验 `SESSION_COOKIE_SECRET` 必须 secret 标注、必须不出现在 runtime-config schema；`runtime.defaultUiLanguage` 必须 public 标注、必须出现在 runtime-config schema。任意一侧错位 lint 必须失败。

#### 6.3 AC C-6 部分验证 + handoff

按 [secrets-and-config spec §6 C-6](../../spec.md#6-验收标准) 与 §1 边界，C-6 完整验收需 [B2 `openapi-v1-contract`](../../../openapi-v1-contract/spec.md) 提供 OpenAPI schema 与 fixture，[D1 `frontend-shell`](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 提供前端 React provider 实装，本 plan 只覆盖 builder + stub handler + 前端 fetcher。在工作日志与本 plan §4 验收标准中明确：

- A4 已交付：builder allowlist 实装、handler stub、前端 fetcher、单测断言「不返回 secret / 不返回 operator-only flag / 尊重 analytics_opt_in」。
- 待 B2 在 OpenAPI schema 锁定 response shape 后，A4 在工作日志记录一次跨 plan 验证 token；B2 plan 引用本 plan 时不得反向修改 builder。
- 待 D1 在 frontend-shell plan 接入 React provider 后，C-6 完整闭环关闭；A4 不再额外开 sibling plan。

#### 6.4 文档与 INDEX 收口

- `config/README.md` Header 完整、内容覆盖 §3.5 列表项。
- `docs/spec/secrets-and-config/plans/INDEX.md` 把本 plan 从 Active 切到 Completed。
- `docs/spec/INDEX.md` 中 `secrets-and-config` 行 Header 与 spec 一致。
- 调用 `/sync-doc-index --check` 确认 Header / INDEX 无 drift。
- 把 Phase 1..5 所有自检命令日志、Phase 6 AC 验证证据与跨 plan handoff token 贴入工作日志。

#### 6.5 风险扫尾

按 §5 风险表逐条复核：redaction 是否覆盖结构化 JSON / error wrap / nested dump 三路径；koanf 合并顺序是否被 unit test 锁定；hot reload race 是否在进程启动 1s 内不抖动；`.env.example` 与代码侧 env key 是否对齐；prod fail-fast 是否会触发 supervisor 无限重启循环（在 README 与工作日志中明确 deployer 必须先补齐 secret 再重启，避免 crash loop 噪音）。任一项风险落地证据缺失，本 plan 必须保持 active 状态，不得切 completed。

### Phase 7: L2 review remediation

#### 7.1 修复 worker env / secret binding 对齐

针对 L2 review finding R-7.1：`cmd/worker` 必须复用与 `cmd/api` 一致的 env binding / secret binding 规则，确保 prod 环境中已注入的 `SESSION_COOKIE_SECRET` / `AI_PROVIDER_API_KEY` / `POSTHOG_PROJECT_API_KEY` 等 secret 会被 loader 读取，`loader.Validate()` 不再错误报告缺失。新增或调整 focused Go test / smoke 覆盖 worker 在 prod + 完整 env 下启动校验通过。

#### 7.2 修复 AI provider base URL fail-fast

针对 L2 review finding R-7.2：`validator.go` 必须同时校验 `AI_PROVIDER_BASE_URL` 与 `AI_PROVIDER_API_KEY`。非 test 的 AIClient-enabled 启动路径缺任一字段都必须 fail-fast，并在错误信息中明确列出缺失 env key；`APP_ENV=test` 仍允许缺 AI provider 配置。

#### 7.3 修复 env_dict code-side key 发现

针对 L2 review finding R-7.3：`scripts/lint/env_dict.py` 必须把 `config.Load` 的 `EnvBindings` / `SecretBindings` 静态字面量纳入 code-side env key 集合，确保 `.env.example` / spec §3.1.1 / code-side declared env keys 三方求差集真实生效。新增 pytest 覆盖「binding map 声明但 `.env.example` 缺 key」必须失败，并保持脚本只读。

#### 7.4 修复 runtime-config cold PostHog flag projection

针对 L2 review finding R-7.4：runtime-config builder 必须能按请求 `FlagContext` 触发 provider evaluation，而不是仅依赖 cold `Snapshot()`；PostHog provider 初始化时必须携带 public flag allowlist，使 prod 首次 `/api/v1/runtime-config` 请求也能返回 public flags 并过滤 operator-only flags。保持 `FeatureFlagClient` 的 D-4 两方法接口不漂移，如需 public projection，使用包内辅助接口或显式方法，不扩大业务消费面。

#### 7.5 修复 prod/staging required config 覆盖

针对 L2 review finding R-7.5：`validator.go` 必须覆盖 spec §3.1.1 / §3.1.2 标记为 prod/staging required 或 conditional 的 P0 keys，包括 app/worker listen addr、database、redis、object storage、AI model profile path、feature flag source/file/posthog、email provider 与现有 auth/AI secrets。对 database/redis/object storage 这类 `config/config.yaml` 中含 dev 默认值的部署依赖，staging/prod 必须要求 runtime env/secret override，避免生产静默连接本机 dev 服务。新增 focused tests 覆盖缺 storage/cache/database override 失败、缺 PostHog host 失败、缺 email provider 失败与完整 prod runtime bindings 通过。

### Phase 8: product-scope v1.2 feature flag remediation

#### 8.1 Red

先调整 feature flag / runtime-config tests，要求旧 `mistake_book_export_enabled` / `growth_dashboard_v1_enabled` / `mock_session_dual_track_enabled` 不得出现在 public runtime config。当前配置仍包含旧 flag 时测试必须失败。

#### 8.2 Green

修订 `config/feature-flags.yaml`、runtime-config tests 与相关文档：新增 `report_retry_plan_enabled` / `readiness_signals_enabled` / `practice_assistance_mode_enabled`，删除旧独立错题本、成长中心和 dual-track flag。

#### 8.3 Verify

运行 `make lint-config`、focused runtime-config tests；repo 搜索确认实现侧不再出现旧三项 feature flag key。

## 5 验收标准

- [secrets-and-config spec §6 验收标准](../../spec.md#6-验收标准) C-1..C-5、C-7..C-12 全部成立，证据贴入工作日志；C-6 partial 验收（A4 builder + stub + 前端 fetcher + 单测）成立，跨 plan 完整 verification 由 B2 / D1 后续 plan 关闭并 cross-link 回本工作日志。
- 本 plan checklist 全部勾选；Phase 6 的 AC 验证命令日志贴入工作日志。
- engineering-roadmap/001 保留的 A4 bootstrap 承诺由 Phase 6.3 关闭 partial、Phase 6.4 关闭文档侧；不重复修改父 roadmap checklist。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| `RedactedString` 被绕过：业务包通过 `fmt.Stringer` 接口或反射拿到底层 `string` | Phase 1.4 在 `redactor.go` 实现 `String()` / `GoString()` / `MarshalJSON()` / `MarshalText()` / `Format(...)` 五个方法返回 `***`；底层字段不导出（`type RedactedString struct { v string }`）；Phase 6.1 redactor_test.go 断言三路径输出 `***`；`Reveal()` 调用方在 godoc 标注必须立即用作 SDK 入参，不允许再传入业务层 |
| `koanf` 多 provider 合并顺序歧义导致 dev / prod 行为不同 | Phase 1.2 在 `loader.go` 中显式按 D-1 顺序串行调用 `Load`，禁止并发；写一份 `loader_test.go` 覆盖「config.yaml 默认 + dev.yaml override + env override + secret override」四层场景，断言最终值为 secret 注入值；任何 koanf 升级必须复跑该测试 |
| feature flag 热加载在进程启动 1s 内 race（mtime 抖动 / YAML 解析中途读取） | Phase 2.3 用 `sync.RWMutex` 保护内部 map；首次加载在构造函数同步完成，热加载 goroutine 仅在首次成功后启动；Phase 6.5 在测试中模拟「启动后立即修改文件」场景，断言不 panic |
| `.env.example` 与代码侧 env key 漂移：开发者新增 `os.Getenv` 但忘记更新 example | Phase 4.2 `make lint-config` 三方求差集 lint 在本地必跑；任何新增 env key 必须先递增 spec §3.1.1 表 → 同步 `.env.example` → 加 validator → 跑 `make lint-config`，README 写入这 4 步流程 |
| prod fail-fast 触发 supervisor / k8s 无限重启循环：缺 secret → exit non-zero → restart → 再 exit | Phase 1.5 错误信息明确列出缺失 key 名；`config/README.md` 与 [E4 `release-gate-and-rollout`](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 的 runbook handoff 中提示 deployer 必须先补齐 secret 再恢复 supervisor；本 plan 不实现自动重试 / 自动 backoff，避免在缺 secret 时静默运行 |
| 业务代码绕过 `FeatureFlagClient` 直接 import PostHog SDK，事后切换 provider 时大面积返工 | Phase 4.1 lint 红线扩展：扫描 `import "github.com/posthog/posthog-go"` 与 `import 'posthog-js'` 在 `backend/internal/<domain>/` 与 `frontend/src/<feature>/` 出现即 fail；只允许在 `backend/internal/platform/featureflag/` 与（D1 后续接入时）`frontend/src/lib/analytics/` 中 import；本 plan 不预先在前端 lint 中收口 PostHog 前端 SDK，留给 [D1 `frontend-shell`](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 与 [F2 `analytics-funnel`](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)，但在 `config/README.md` 显式写明此红线，避免后续 plan 漏接 |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-05-05 | 1.6 | L2 深审修正文档 allowlist 旧口径：`cmd/migrate` 是 B4 迁移 CLI 的合法 env 读取入口；plan 与当前 spec §4.1 / `getenv_boundary.go` 对齐。 | historical deep reconcile |
| 2026-05-04 | 1.5 | L1 plan-review remediation：补齐当前强制的质量门禁分类，不改变已完成 config/secret/feature flag 范围。 | historical-spec-implementation-review/001 |
| 2026-05-03 | 1.4 | 原地 reopen，新增 Phase 8 remediation：按 product-scope v1.2 替换旧错题本 / 成长中心 / dual-track feature flag baseline。 | secrets-and-config v1.9 |
| 2026-04-30 | 1.3 | L2 code-review remediation：补 prod/staging required config 覆盖与 dev-default runtime override 防线。 | plan-code-review --fix |
| 2026-04-30 | 1.2 | L2 code-review remediation：worker bindings、AI base URL fail-fast、env_dict code-side binding discovery、runtime-config cold PostHog projection。 | plan-code-review --fix |
| 2026-04-29 | 1.1 | 对齐 spec v1.7：24 项 env key、`async.queueWeights` config-only 字段、PostHog last-known-good 缓存降级、secret 样本只允许临时生成不入文档。 | plan-review remediation |
