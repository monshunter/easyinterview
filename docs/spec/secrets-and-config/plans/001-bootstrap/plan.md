# Secrets and Config Bootstrap

> **版本**: 1.22
> **状态**: active
> **更新日期**: 2026-07-16

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [secrets-and-config spec](../../spec.md) §3.1 已锁定的 D-1..D-9 决策与 §3.1.1 / §3.1.2 锁定的 P0 必备 env key 字典、canonical config schema、`async.queueWeights` config-only 字段落到代码：建立 `backend/internal/platform/{config,secrets,featureflag}/` 三个 Go 包真理源、`config/*.yaml` + `.env.example` + `config/feature-flags.yaml` 默认值集合、`make lint-config` 与 `scripts/git-hooks/pre-commit-secrets.sh` 本地质量门禁，以及 `GET /api/v1/runtime-config` builder/handler；前端由 B2 generated client/types 与 D1 `AppRuntimeProvider` 构成唯一消费链。通过 verification phase 证明 [secrets-and-config spec §6](../../spec.md#6-验收标准) C-1..C-12 在本仓库可重复跑通。

本 plan 是 `secrets-and-config` 唯一的 plan；后续若需扩展（Vault / SOPS / platform secret / K8s Secret provider，自动 secret rotation，分桶 feature flag），按 §7 约束递增 spec 与本 plan 版本，原地修订，不再开 sibling plan。

本次 v1.15 技术债清理删除只被自身单测消费的平行 runtime-config fetch/cache 包；A4 只拥有 backend allowlist 与 endpoint，正式前端由 D1 `AppRuntimeProvider` 通过 B2 generated client/types 读取配置。

本次 v1.16 随连续对话简化删除 `practice_hint_enabled` / `practice_assistance_mode_enabled` 及 public runtime allowlist，baseline 收敛为 4 项仍有当前消费者的 report/readiness/operator flags。

本次 v1.17 随 Home JD intake 收敛为 raw text，删除 A4 中已无当前消费者的 TargetJob attachment maxBytes config、validator 与 typed composition binding；`upload.maxBytes.resume` 和 `upload.maxBytes.privacyExport` 的默认值、验证与业务边界保持不变。

本次 v1.18 按用户确认的方案 A 统一运行时内容大小配置：以真实失败样本、当前 1,000,000-token report profile 与各业务输入形态重新校准默认值；所有可配置 size 参数必须同时拥有 typed code default，缺 key 使用缺省值，显式非法值启动失败；公共 `runtime-config` 仅投影前端提交前需要的五项限制。

本次 v1.19 按奥卡姆剃刀收敛配置测试：默认值、override 与非法组合由 platform config 的单一 typed contract suite 持有；消费者仅保留无法由类型/构造保证的业务分支，使用小型注入值或 metadata。配置数值传播不再扩展或重复运行 domain BDD / scenario，也不构造默认大小文件或字符串。

本次 v1.10 技术债清理同步当前实现事实：`runtime_config_handler.go` 支持由 C1 backend-auth 注入 session-aware resolver；resolver 缺省时才使用 anonymous opt-out 默认，不再将 handler 描述为 stub。

本次 v1.11 技术债清理将 `config/config.yaml` 与 `.env.example` 的 secret 默认值描述收敛为空字符串 / 说明注释，不改变配置文件合同或 lint 行为。

本次 v1.12 技术债清理统一 feature flag 范围术语为 `out-of-scope`，并将 plan context 对齐当前 spec 2.13；六项 current baseline 与负向回归输入保持不变。

本次 v1.13 技术债清理删除仅供单测清理 module cache 的 `_resetRuntimeConfigCache` export。各用例改用生产 `forceRefresh` 选项建立独立缓存边界，缓存、失败恢复和刷新行为不变。

Phase 1-14 及其已完成条目只保留为历史交付证据；Phase 15 是当前本地 AI raw I/O 配置与隐私边界合同。旧 Phase 中被 Phase 13 清退的附件配置、散落硬编码和配置专用场景口径，以及被 Phase 15 删除的 stderr raw-output 开关，均不得作为当前实现、验收或兼容要求。

## 2 背景

[engineering-roadmap §5.1](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 A4 保留为当前 active Foundation spec；后续 [B2 `openapi-v1-contract`](../../../openapi-v1-contract/spec.md)、[backend-auth](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)、[frontend-shell](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 等 workstream 依赖本计划输出的配置 / secret / feature flag 契约。本 plan 通过 §4 的 6 个 phase 验收 [secrets-and-config spec §6](../../spec.md#6-验收标准) C-1..C-11，关闭 [001-decompose-subspecs](../../../engineering-roadmap/plans/001-decompose-subspecs/checklist.md) 保留的 A4 bootstrap 承诺。

执行本 plan 前必须确认：

- [A1 `repo-scaffold/001-bootstrap`](../../../repo-scaffold/plans/001-bootstrap/plan.md) 已创建根 `Makefile`、`backend/`、`frontend/`、`scripts/git-hooks/`、`.gitignore` 等容器目录与基础 hook 入口；本 plan 只在其上扩展。
- [B1 `shared-conventions-codified/001-bootstrap`](../../../shared-conventions-codified/plans/001-bootstrap/plan.md) 已落地 `backend/go.mod` 与 `backend/internal/shared/` 共享包；本 plan 的 Go 代码引用其常量。
- [A2 `local-dev-stack/001-bootstrap`](../../../local-dev-stack/plans/001-bootstrap/plan.md) 的 `deploy/dev-stack/.env.example` 字段名已与 spec §3.1.1 字典对齐；本 plan 的 `.env.example` 是仓库根真理源，A2 dev stack 在本地启动时复用同一字典。

每个 phase 是可独立部署 / 验证的纵向行为切片：Phase 1 起来即可由 Go 代码 `config.Get*` 读取三层合并的配置；Phase 2 起来即可由业务代码通过 `SecretSource` / `FeatureFlagClient` 接口隔离 provider；Phase 3 起来即有完整的 `.env.example` 与 `config/*.yaml` 字典；Phase 4 起来即有 `make lint-config` 与 pre-commit secret 拦截；Phase 5 起来即有 `runtime-config` 端到端链路；Phase 6 收口验证 C-1..C-11 并完成 handoff。

本 plan 不部署 PostHog（归 [F2 `analytics-funnel`](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 与 [E4 `release-gate-and-rollout`](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)），不实现 Vault / SOPS / platform secret / K8s Secret provider（归 P1 / E4），不冻结 `/api/v1/runtime-config` 的 OpenAPI schema（归 [B2 `openapi-v1-contract`](../../../openapi-v1-contract/spec.md)；A4 在本 plan 中只交付 response builder + handler）。

## 3 质量门禁分类

- **Plan 类型**: `platform-config + code-internal + contract + tooling`。Phase 14 的 SMTP provider 配置不新增独立用户流程；用户可感知的邮件投递行为由 backend-auth 持有，A4 只持有配置与 secret 合同。
- **TDD 策略**: 必须通过 `/implement secrets-and-config/001-bootstrap platform-config` → `/tdd`。platform config 保留一组表驱动契约测试覆盖缺 key、合法 override、显式非法值与跨字段约束；消费者只为错误映射、持久化原子性、provider call/no-call、协议读取上限等非平凡分支保留 focused test，并注入小型边界值或 metadata。
- **BDD 策略**: 不适用。默认数值、配置注入和 public projection 不是独立用户流程；不得为其新增、扩展或重复运行任何 E2E。真实用户流程由 domain owner 独立验证，但不充当 A4 配置 gate。
- **替代验证 gate**: platform config typed contract、OpenAPI schema/codegen drift、AI profile catalog lint、必要的 provider/domain/frontend focused tests、旧硬编码 negative search、`sync-doc-index --check`、`make docs-check`、`git diff --check`。

## 4 实施步骤

### Phase 1: Three-tier config loader 与 redactor

#### 1.1 落地 `backend/internal/platform/config/` 包骨架

按 [secrets-and-config spec §5](../../spec.md#5-模块边界) 把 `loader.go` / `validator.go` / `redactor.go` / `getters.go` / `doc.go` 落到 `backend/internal/platform/config/`，module path 沿用 [B1 shared helper contract](../../../shared-conventions-codified/plans/001-bootstrap/plan.md#phase-2-go--ts-shared-helpers) 锁定的 `github.com/monshunter/easyinterview/backend`。`doc.go` 用一段 godoc 概述说明三层优先级（与 [secrets-and-config spec §3.1 D-1](../../spec.md#31-已锁定决策含-p0-必备-env-key-字典) 对齐）以及对外可见的 `Get*` API 命名约定。

#### 1.2 接入 `koanf` 作为 loader 实现

按 [secrets-and-config spec §3.2](../../spec.md#32-待确认事项) 待确认项的默认决议引入 `github.com/knadh/koanf/v2` + `koanf/parsers/yaml` + `koanf/providers/env` + `koanf/providers/file`；不引入 `viper`。`loader.go` 中按 D-1 顺序合并：先 `config/config.yaml`（默认值），再 `config/{APP_ENV}.yaml`（环境 override），再 env var，最后 runtime secret（通过 `SecretSource` 注入）。`koanf` 默认使用最后写入覆盖前者，必须在 loader 中显式约定合并顺序，禁止在多 provider 并行调用 `Load`，避免并发 merge 顺序歧义。

#### 1.3 落地 `Get*` API 与类型化访问器

`getters.go` 暴露 `GetString(key string) string` / `GetInt(key string) int` / `GetBool(key string) bool` / `GetSecret(key string) RedactedString`；任何业务包通过这些 API 读取配置。错误路径（缺失 required key / 类型不符）由 validator 统一处理后返回，不由 getter 自行 panic。getter 必须以 `app.listenAddr` 形式接受点路径，与 [secrets-and-config spec §3.1.2](../../spec.md#312-canonical-config-schema-分类) 中的 `Config path` 列对齐。

#### 1.4 实现 `RedactedString` 类型与 redactor

按 [secrets-and-config spec §4.2](../../spec.md#42-安全约束) 在 `redactor.go` 中实现 `RedactedString` 类型：底层为 `string`；`String()` / `GoString()` / `MarshalJSON()` / `MarshalText()` / `Format(...)` 均输出 `***`；只有显式 `Reveal() string` 返回明文。redactor 必须覆盖结构化 JSON 日志、`fmt.Errorf("...: %w", err)` 包装链、嵌套 config dump（递归到 struct 字段时仍返回 `***`）三种路径。Phase 6 自检中通过 `errors_test.go` / `redactor_test.go` 强制这三条路径。

#### 1.5 启动期 fail-fast validator

`validator.go` 在进程启动期消费 spec §3.1.2 标记为 `required` 的字段集合，按 `APP_ENV` 维度判定：`APP_ENV=test` 允许缺 AI / Email / Session 相关 secret；`APP_ENV=staging|prod` 必填 secret 缺失时返回 `error`，由当前 backend runtime entrypoint 接住后非零退出。错误信息必须明确列出缺失 key 名（与 [secrets-and-config spec §6 C-2](../../spec.md#6-验收标准) 一致），避免 prod 退出后 deployer 不知道补哪个 key。

### Phase 2: SecretSource + FeatureFlagClient 抽象与 provider 实现

#### 2.1 落地 `backend/internal/platform/secrets/` 包

按 [secrets-and-config spec D-3](../../spec.md#31-已锁定决策含-p0-必备-env-key-字典) 在 `secrets/secrets.go` 写入接口：

```go
type SecretSource interface {
    Get(name string) (string, error)
}
```

接口签名固定，P1 升级到 Vault / SOPS / platform secret / K8s Secret 时只能新增 provider，不允许修改签名。`secrets/env_provider.go` 实现 `EnvSecretSource`：从 `os.Getenv` 读取（注意 `os.Getenv` 仅允许出现在 `internal/platform/config/`、`internal/platform/secrets/` 与 `cmd/{api,migrate}/main.go`，由 Phase 4.1 lint 收口）。Phase 1 loader 把 `EnvSecretSource` 作为默认运行时 secret 来源，与 D-1 第一层（`runtime secret > env var ...`）一致。

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

`secrets_test.go` 与 `featureflag_test.go` 覆盖：`EnvSecretSource.Get` 缺失返回 error；`FileFlagProvider` 解析 YAML、修改 mtime 后 ≤30s 热加载；`PostHogFlagProvider` 在 mock HTTP server（`httptest.NewServer`）返回 `decide` 模拟响应时 `IsEnabled("sample_public_flag", ctx)` 返回 true；mock 5xx / timeout 时命中 last-known-good 快照；无快照时返回 degraded/error；`POSTHOG_SELF_HOSTED=false` 在 staging/prod 启动时 validator 返回 fail-fast error。Phase 6 自检会再次串行复跑这些用例。

### Phase 3: 配置文件骨架与 .env.example 字典对齐

#### 3.1 落地 `config/config.yaml` 默认值

按 [secrets-and-config spec §3.1.2](../../spec.md#312-canonical-config-schema-分类) 锁定的 canonical config schema 写入仓库根 `config/config.yaml`（D-1 第一层默认值）。所有 secret 字段（`database.url` / `redis.url` / `objectStorage.accessKey` / `objectStorage.secretKey` / `auth.sessionCookieSecret` / `auth.challengeTokenPepper` / `email.providerApiKey` / `ai.providerApiKey` / `featureFlag.posthogProjectApiKey`）必须留空字符串，禁止写入真实凭证；明文字段（如 `runtime.appVersion` / `runtime.defaultUiLanguage` / `app.listenAddr` / `featureFlag.source`）写入 spec 默认值。`auth.sessionCookieName` 固定字面量 `ei_session`，与 [ADR-Q1 §3](../../../engineering-roadmap/decisions/ADR-Q1-auth.md#3-决策) 与 spec D-8 一致；`async.queueWeights` 默认 `critical: 6` / `default: 3` / `low: 1`，不提供 env override；本 plan 不允许任何 env key 覆盖 `ei_session`。

#### 3.2 落地 `config/{dev,staging,prod}.yaml` 环境 override

按 [secrets-and-config spec §2.1](../../spec.md#21-in-scope) 写入三份 env override 文件，禁止包含任何 secret 明文。`dev.yaml` 设 `log.level: debug`、`featureFlag.source: file`；`staging.yaml` 与 `prod.yaml` 设 `featureFlag.source: posthog` 与 `featureFlag.posthogSelfHosted: true`，与 [ADR-Q3](../../../engineering-roadmap/decisions/ADR-Q3-analytics-platform.md) 一致；任何运行时 secret 字段在三份文件中都留空，提示运维通过 env / 真正 secret 注入。

#### 3.3 落地 `.env.example` 与 env key 字典

按 [secrets-and-config spec §3.1.1](../../spec.md#311-p0-必备-env-key-字典) 写入仓库根 `.env.example`，包含全部 env key（`APP_ENV` / `APP_LISTEN_ADDR` / `DATABASE_URL` / `REDIS_URL` / `OBJECT_STORAGE_*` / `OTEL_EXPORTER_OTLP_ENDPOINT` / `LOG_LEVEL` / `SESSION_COOKIE_SECRET` / `AUTH_CHALLENGE_TOKEN_PEPPER` / `AI_PROVIDER_*` / `AI_MODEL_PROFILE_PATH` / `FEATURE_FLAG_*` / `POSTHOG_*` / `EMAIL_PROVIDER` / `EMAIL_SMTP_*` / `EMAIL_FROM_ADDRESS` / `EMAIL_VERIFY_BASE_URL`）。所有 secret 字段只写说明注释（如 `# secret; populate via runtime secret in prod`），不允许写真实 key 样本。每行注释必须标注「Owner subspec」与「prod required: yes/no/conditional」，与 spec §3.1.1 表格一一对应。`async.queueWeights` 是 config-only 字段，不进入 `.env.example`。

#### 3.4 落地 `config/feature-flags.yaml` baseline

按当前 product/UI scope 写入 4 项 baseline flag（`report_evidence_v2_enabled` / `report_retry_plan_enabled` / `readiness_signals_enabled` / `ai_fallback_model_enabled`）。每个 flag 必须显式标注 `public: true|false`：前三项为 `public: true`；`ai_fallback_model_enabled` 设 `public: false`（operator-only），由 Phase 5 runtime-config builder 在 allowlist 中过滤。已删除的 practice hint/assistance flags 不得保留兼容投影。

#### 3.5 落地 `config/README.md`

按 [secrets-and-config spec §4.3](../../spec.md#43-文档约束) 写一份一屏长度 README，覆盖：三层优先级（D-1）、各文件用途（`config.yaml` / `dev.yaml` / `staging.yaml` / `prod.yaml` / `feature-flags.yaml`）、新增 env key 流程（先递增 spec 版本 + history → 同步 `.env.example` → 同步 §3.1.2 schema → 加 validator → 跑 `make lint-config`）、`RedactedString` 使用示范、`runtime-config` allowlist 边界。README 必须有 cross-link 回到 spec §3.1 / §4。

#### 3.6 .gitignore 红线扩展（与 A1 协作）

按 [secrets-and-config spec §4.2](../../spec.md#42-安全约束) 红线，本 plan 在 [A1 `repo-scaffold/001-bootstrap`](../../../repo-scaffold/plans/001-bootstrap/plan.md) 已创建的根 `.gitignore` 中追加：`*.secret.yaml`、`*.secret.json`、`config/local.*.yaml`、`.env`、`.env.local`、`config/feature-flags.local.yaml`。追加位置必须独立成段并打标注 `# A4 secrets-and-config red lines`，便于后续审查。本 plan 不重写 A1 已有的 hook 入口，只追加 secret 红线条目。

### Phase 4: Lint / pre-commit hook / `make lint-config`

#### 4.1 golangci-lint 自定义规则（拒绝 `os.Getenv` 越界）

按 [secrets-and-config spec §4.1](../../spec.md#41-边界约束) 落地：在 [B1 lint gate](../../../shared-conventions-codified/plans/001-bootstrap/plan.md#phase-3-lint-and-naming-gates) 已落地的 `backend/.golangci.yml` 中追加一条本地可执行规则。优先选择 `revive` 自定义 rule（如 `disallow-direct-env-access`）；若 `revive` 表达不动，则在 `scripts/lint/` 下落 `getenv_boundary.go`（Go AST checker），由 `make lint-config` / `make lint` 调用 `go run scripts/lint/getenv_boundary.go -root backend` 扫描 `backend/...`。规则 allowlist 仅放行 `backend/internal/platform/config/`、`backend/internal/platform/secrets/`、`backend/cmd/api/`、`backend/cmd/migrate/`，与当前 spec §4.1 一致；其它包出现 `os.Getenv` 必须 lint 失败（关闭 [secrets-and-config spec §6 C-7](../../spec.md#6-验收标准)）。

#### 4.2 `make lint-config`：env key dictionary drift 检查

按 [secrets-and-config spec §6 C-9](../../spec.md#6-验收标准) 落地：在 `scripts/lint/env_dict.py`（或 `scripts/lint/env_dict.sh`）实现一次性扫描器：

- 解析 `.env.example` 提取所有 env key 名。
- AST / regex 解析 `backend/internal/platform/config/` 与 `backend/cmd/{api,migrate}/main.go` 中所有 `os.Getenv("...")` / `Get*("...")` 调用，提取代码侧已声明的 env key。
- 解析 spec `§3.1.1` 表格（通过定位 markdown 表头与列分隔符）并提取「Key」列。
- 三方求差集：任一方缺失 key 必须 fail，错误信息列出缺失项与三方分别声明的差异。

`Makefile` 新增 `.PHONY: lint-config` target 并入 `make lint`：`lint-config` 失败时整个 `make lint` 失败。脚本只读不改文件，不能在缺 key 时静默自动补齐。

#### 4.3 pre-commit hook：拦截敏感前缀

按 [secrets-and-config spec D-6](../../spec.md#31-已锁定决策含-p0-必备-env-key-字典) 与 [§6 C-8](../../spec.md#6-验收标准) 落地 `scripts/git-hooks/pre-commit-secrets.sh`：扫描 `git diff --cached` 命中 `AKIA[0-9A-Z]{16}` / `sk-[A-Za-z0-9]{20,}` / `xox[baprs]-[A-Za-z0-9-]+` 任一正则即 fail（exit 1），错误信息列出文件名与行号，不输出命中的 secret 字面量本身（避免日志再泄漏一次）。脚本由 [A1](../../../repo-scaffold/plans/001-bootstrap/plan.md) 已建立的 `scripts/git-hooks/` 入口注册，本 plan 只追加文件并扩展安装脚本，不重写 A1 hook 框架。

#### 4.4 本地 gitleaks 第二层

按 [secrets-and-config spec D-6](../../spec.md#31-已锁定决策含-p0-必备-env-key-字典) 在 `scripts/lint/gitleaks.sh` 落入第二层扫描入口：调用本地已安装的 `gitleaks detect --no-git --redact`（不要求开发者必须装；脚本检测到未安装时打印安装提示并 exit 0，避免阻塞，但 README 标注推荐安装）。`make lint` 调用此脚本作为第二层。当前阶段不接入远端 CI secret scan，仅在 [A5 `ci-pipeline-baseline`](../../../ci-pipeline-baseline/spec.md) 触发条件成立后再评估。

#### 4.5 Phase 4 自检

构造一份故意越界改动跑 `make lint`：在 `backend/internal/auth/` 添加一行 `os.Getenv("SESSION_COOKIE_SECRET")` → `make lint` 必须失败并报 `os.Getenv outside platform/config`；删除 `.env.example` 中 `AI_PROVIDER_BASE_URL` → `make lint-config` 失败；在临时文件中程序化生成一行形似真实凭证的值并 `git add` → pre-commit hook 拦截。所有自检完成后必须把越界改动 revert，避免污染主分支；文档与 fixture 中不得长期保存命中 secret 正则的样本文本。

### Phase 5: `runtime-config` endpoint 接入与 generated-client handoff

#### 5.1 后端 `runtime-config` builder 与 handler

按 [secrets-and-config spec D-2 / §3.1.2 / §6 C-6](../../spec.md#31-已锁定决策含-p0-必备-env-key-字典) 在 `backend/internal/platform/config/runtime_config.go` 落地 `BuildRuntimeConfig(ctx, session) RuntimeConfig`：

- 字段 allowlist 严格限定 `appVersion` / `defaultUiLanguage` / `analyticsEnabled` / `featureFlags` / `postHogPublicKey`；
- `featureFlags` 仅纳入 `config/feature-flags.yaml` 或 PostHog 中标 `public: true` 的 flag；`ai_fallback_model_enabled` 等 operator-only flag 必须被过滤；
- 若 `ctx` 携带有效 session 且 `user_settings.analytics_opt_in == false`，则 `analyticsEnabled = false` 且不返回 `postHogPublicKey`（与 D-2 一致）；
- 任何 secret 字段绝对不能进入 response。

`backend/internal/platform/config/runtime_config_handler.go` 落地 `GET /api/v1/runtime-config` HTTP handler：直接调用 `BuildRuntimeConfig`，序列化为 JSON，并接受 C1 backend-auth 注入的 session-aware resolver；resolver 缺省时使用 anonymous opt-out 默认。OpenAPI schema 真理源由 [B2 `openapi-v1-contract`](../../../openapi-v1-contract/spec.md) 持有；本 plan 仅交付 builder + handler，schema 与 fixture 一致性由 B2 在引用 A4 时验证。

#### 5.2 前端 generated-client handoff

[D1 `frontend-shell`](../../../frontend-shell/spec.md) 的 `AppRuntimeProvider` 通过 B2 `EasyInterviewClient.getRuntimeConfig()` 与 generated `RuntimeConfig` 类型读取当前 endpoint，并与 auth bootstrap 共用同一个 client。A4 不维护第二套 fetch/cache/type 边界；前端任何代码不得直接读取 `import.meta.env.VITE_*` 之外的 build-time 变量。

#### 5.3 contract handoff 边界与 cross-link

在 `backend/internal/platform/config/runtime_config.go` cross-link 到 [B2 `openapi-v1-contract` spec](../../../openapi-v1-contract/spec.md)，明确「OpenAPI schema 与 generated TS 类型真理源在 B2；A4 仅持有 builder 与字段 allowlist」。如果 B2 后续扩展 schema，必须先递增 A4 spec 版本并扩展 builder allowlist。

#### 5.4 Phase 5 自检

构造下列单测覆盖 [secrets-and-config spec §6 C-6](../../spec.md#6-验收标准)：

- `runtime_config_test.go`：当 `report_evidence_v2_enabled.public=true`、`ai_fallback_model_enabled.public=false`、`analytics_opt_in=false` 时，response 包含公开 report flag 但不含 `ai_fallback_model_enabled`，`analyticsEnabled=false`，无 `postHogPublicKey`，无任何 secret 字段。
- D1 `AppRuntimeProvider` focused tests：generated client 的 `getRuntimeConfig` 与 `getMe` bootstrap、失败态及 refresh 行为保持通过。
- 手工 smoke：本地 `go run ./backend/cmd/api` 启动后 `curl http://127.0.0.1:10901/api/v1/runtime-config` 返回 allowlist response。

### Phase 6: Verification + handoff

#### 6.1 AC C-1..C-5 验证（Phase 1 / Phase 2 关闭）

依次复跑：

- **C-1（三层合并）**：`config/config.yaml` 默认 `log.level: info`；`config/dev.yaml` 设 `log.level: debug`；env 设 `APP_LISTEN_ADDR=:9090`；`go run ./backend/cmd/api -dump-config` 必须显示 `app.listenAddr=:9090` 与 `log.level=debug`，其它字段保持默认。
- **C-2（fail-fast）**：`APP_ENV=prod ./bin/api` 缺 `SESSION_COOKIE_SECRET` 时退出码非 0 且 stderr 含 `missing required secret: SESSION_COOKIE_SECRET`。
- **C-3（file flag 热加载）**：`FEATURE_FLAG_SOURCE=file` 启动后修改 `config/feature-flags.yaml`，30s 内 `featureflag.IsEnabled("sample_public_flag", ctx)` 反映新值。
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

按 [secrets-and-config spec §6 C-6](../../spec.md#6-验收标准) 与 §1 边界，C-6 完整验收需 [B2 `openapi-v1-contract`](../../../openapi-v1-contract/spec.md) 提供 OpenAPI schema 与 fixture，[D1 `frontend-shell`](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 提供前端 React provider 实装，本 plan 只覆盖 builder + handler + 前端 fetcher。在工作日志与本 plan §4 验收标准中明确：

- A4 已交付：builder allowlist 实装、handler、前端 fetcher、单测断言「不返回 secret / 不返回 operator-only flag / 尊重 analytics_opt_in」。
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

#### 7.1 修复 runtime entrypoint env / secret binding 对齐

针对 L2 review finding R-7.1：runtime entrypoint 必须复用 `cmd/api` 的 canonical env binding / secret binding 规则，确保 prod 环境中已注入的 `SESSION_COOKIE_SECRET` / `AI_PROVIDER_API_KEY` / `POSTHOG_PROJECT_API_KEY` 等 secret 会被 loader 读取，`loader.Validate()` 不再错误报告缺失。backend-runtime-topology v1.0 后不再保留单独 runtime entrypoint；验证以 current backend entrypoint 与 config package focused tests 为准。

#### 7.2 修复 AI provider base URL fail-fast

针对 L2 review finding R-7.2：`validator.go` 必须同时校验 `AI_PROVIDER_BASE_URL` 与 `AI_PROVIDER_API_KEY`。非 test 的 AIClient-enabled 启动路径缺任一字段都必须 fail-fast，并在错误信息中明确列出缺失 env key；`APP_ENV=test` 仍允许缺 AI provider 配置。

#### 7.3 修复 env_dict code-side key 发现

针对 L2 review finding R-7.3：`scripts/lint/env_dict.py` 必须把 `config.Load` 的 `EnvBindings` / `SecretBindings` 静态字面量纳入 code-side env key 集合，确保 `.env.example` / spec §3.1.1 / code-side declared env keys 三方求差集真实生效。新增 pytest 覆盖「binding map 声明但 `.env.example` 缺 key」必须失败，并保持脚本只读。

#### 7.4 修复 runtime-config cold PostHog flag projection

针对 L2 review finding R-7.4：runtime-config builder 必须能按请求 `FlagContext` 触发 provider evaluation，而不是仅依赖 cold `Snapshot()`；PostHog provider 初始化时必须携带 public flag allowlist，使 prod 首次 `/api/v1/runtime-config` 请求也能返回 public flags 并过滤 operator-only flags。保持 `FeatureFlagClient` 的 D-4 两方法接口不漂移，如需 public projection，使用包内辅助接口或显式方法，不扩大业务消费面。

#### 7.5 修复 prod/staging required config 覆盖

针对 L2 review finding R-7.5：`validator.go` 必须覆盖 spec §3.1.1 / §3.1.2 标记为 prod/staging required 或 conditional 的 P0 keys，包括 app listen addr、database、redis、object storage、AI model profile path、feature flag source/file/posthog、email provider 与现有 auth/AI secrets。对 database/redis/object storage 这类 `config/config.yaml` 中含 dev 默认值的部署依赖，staging/prod 必须要求 runtime env/secret override，避免生产静默连接本机 dev 服务。新增 focused tests 覆盖缺 storage/cache/database override 失败、缺 PostHog host 失败、缺 email provider 失败与完整 prod runtime bindings 通过。

### Phase 8: product-scope v1.2 feature flag remediation

#### 8.1 Red

先调整 feature flag / runtime-config tests，要求 out-of-scope `mistake_book_export_enabled` / `growth_dashboard_v1_enabled` / `mock_session_dual_track_enabled` 不得出现在 public runtime config。当前配置包含 out-of-scope flag 时测试必须失败。

#### 8.2 Green

修订 `config/feature-flags.yaml`、runtime-config tests 与相关文档：current baseline 仅使用仍有消费者的 report/readiness flags；practice hint/assistance flags 与范围外独立错题本、成长中心和 dual-track flag 不进入 public runtime config。

#### 8.3 Verify

运行 `make lint-config`、focused runtime-config tests；repo 搜索确认实现侧不出现三项 out-of-scope feature flag key。

### Phase 10: Unused duration getter removal

删除零生产消费者的 `Loader.GetDuration`、getter 自测分支与 package doc 引用。当前 runtime 时长继续通过 typed config 或带明确单位的 `GetInt` 转换读取，不保留未使用 accessor。

### Phase 11: Parallel frontend runtime-config client removal

删除仅由自身单测消费、正式 `src/main.tsx` 依赖图不可达的平行 frontend runtime-config fetch/cache/type/test 包。D1 `AppRuntimeProvider` 继续通过 B2 generated client/types 获取同一 endpoint；同步 frontend README、A4 spec/plan/checklist/context 与 D1 discovery context，不保留 wrapper、兼容 export 或退役标记。

### Phase 12: TargetJob attachment maxBytes config removal

本批次依赖顺序固定为：统一 RED → B1/B3/OpenAPI 真理源与生成物 → A4/B4/F3/backend-upload/backend-async-runner 各自 owner surface → backend-targetjob Phase 18 集成 → BDD/全局 zero-reference。A4 只在 B1/B3/OpenAPI 当前合同可消费后修改自己拥有的 config、validator 与 backend API composition surface；purpose/DB constraint 删除分别归 backend-upload/B4，任一上游 handoff 未完成时不得宣称本 Phase 完成。

#### 12.1 Red: pin the current two-purpose maxBytes contract

先更新 config schema、validator 与 backend API composition tests，使它们要求 maxBytes 当前集合只包含 `resume` 和 `privacyExport`；旧 TargetJob attachment key 仍存在时，focused tests 与 active-key inventory 必须失败。负向断言同时固定 Resume 10MB、Privacy Export 5MB 与 presign TTL 不变。

#### 12.2 Green: remove the orphaned config and validator surface

从 `config/config.yaml`、`backend/internal/platform/config/validator.go`、validator fixtures/tests 与 `backend/cmd/api` typed upload-limit composition 删除旧 TargetJob attachment maxBytes binding，不增加 alias、默认回退或兼容读取。文件 purpose/schema 的跨 owner 删除由 backend-upload / TargetJob owner 承接；A4 只删除自己拥有的 config、validator 与 composition surface。

#### 12.3 Verify: lint, focused tests and zero-reference

运行 `make lint-config`、platform config focused tests 与 backend API composition tests，证明 Resume/Privacy 非正数继续 fail-fast、默认值保持 10MB/5MB。active zero-reference 扫描覆盖 `config/`、platform config、backend API composition 与 A4 current docs；work-journal/history/bug/report 和显式 negative tests 可作为合法历史证据保留。

#### 12.4 BDD substitute gate

BDD 不适用：本 phase 删除内部 config/validator orphan，不新增 UI、HTTP wire 或用户业务流程。替代 gate 为 Red/Green focused tests、`make lint-config`、typed composition test、current-key inventory 与 active zero-reference。

### Phase 13: Runtime content size defaults and boundary alignment

#### 13.1 锁定统一 typed defaults

在 platform config 建立单一 typed `ContentLimits` 缺省真理源，并让 repo-tracked YAML 精确镜像：HTTP request 10MiB、Resume upload 10MiB、Privacy Export upload 5MiB、Resume active 10、Resume extracted/paste 各 384KiB、TargetJob raw text 96KiB、Practice message 32KiB、Practice session text 256KiB、Report framed input 896KiB、AI provider response body 4MiB。缺 key 使用代码缺省；显式 `0`、负数或跨字段非法组合必须 fail-fast。

#### 13.2 统一 backend 注入与 byte 边界

删除 report 48,000 bytes、Practice 8,000 runes、Resume parse 8MiB、idempotency-only 10MiB 与各 domain/provider 重复常量。所有用户文本按 UTF-8 bytes 判断，limit 接受、limit+1 返回可识别错误；不得静默截断，也不得在越界后调用 AI provider。全局 HTTP body cap 保护 API JSON body，同时 domain limit 保留业务语义错误。

#### 13.3 对齐 AI profile、provider response 与容量 gate

canonical AI profile 的 `max_tokens` / `context_window_tokens` 继续由 A3 catalog 持有；六个 active profile 的 `max_tokens` 不低于 16384，缺失字段使用 typed code defaults。四个 provider adapter 统一使用注入的 `ai.maxResponseBodyBytes`。配置层不把 bytes 与 tokens 直接相加，不维护 report budget test、exact-profile lint 或真实 provider smoke；profile 合法性由 A3 loader owner 契约与 active-budget floor lint 承接。

#### 13.4 投影 public runtime-config 并接入前端

`RuntimeConfig.contentLimits` 精确投影 `resumeUploadBytes`、`resumePasteTextBytes`、`targetJobRawTextBytes`、`practiceMessageBytes`、`practiceSessionTextBytes`。正式前端通过 generated client + `AppRuntimeProvider` 消费，不保留 2MiB 或 rune-count 本地真理源；report/HTTP/provider/profile 限制不得公开。

#### 13.5 跨 owner 最小验证矩阵

| 路径 | Config owner | Backend consumer | Frontend consumer | 最小验证 owner |
|------|--------------|------------------|-------------------|------------------|
| Resume upload/paste | A4 + backend-upload/resume | upload handler + resume parse | frontend-resume-workshop | A4 typed contract + 一个小型 consumer boundary test；不构造 10MiB/384KiB 材料 |
| TargetJob raw JD | A4 + backend-targetjob | target-job import | frontend-home | A4 typed contract + 既有 domain UTF-8 boundary test；不扩展场景 |
| Practice message/session | A4 + backend-practice | chat handler + message store aggregate | frontend-practice | A4 typed contract + 既有持久化/错误语义 focused test；不扩展场景 |
| Report framed input | A4 + A3 + backend-review | report context builder + provider adapter | internal-only error receipt | 小型 framed-input business test + provider call/no-call；不以真实大文件证明配置 |

#### 13.6 最小回归与 post-pass

只运行 13.5 中的 owner contract 与必要 focused gates，不为配置注入启动场景环境，也不在多个 consumer 重复默认值或 `limit / limit+1`。负向搜索必须证明旧 48,000、2MiB、8,000-rune、8MiB 与 adapter-local 4MiB 不再是生产真理源；合法历史与针对独立业务缺陷的 boundary test 可保留。成功后执行 doc reconcile、Bug 记录评估与 retrospective。

### Phase 14: Standard SMTP provider config

#### 14.1 Typed provider contract

在 `backend/internal/platform/config` 先补表驱动 RED：`mailpit` 要求 host/port/from、`TLS_MODE=none` 且无认证；`smtp` 要求 host/port/from/username/password、`TLS_MODE=starttls|tls`；未知 provider/mode、非法 port、staging/prod Mailpit/none 全部 fail-fast。然后最小修改 bindings / secret bindings / validator。

#### 14.2 Dictionary and drift cleanup

新增 `EMAIL_SMTP_USERNAME` / `EMAIL_SMTP_PASSWORD` / `EMAIL_SMTP_TLS_MODE`，删除未被 runtime 消费的 `EMAIL_PROVIDER_API_KEY`；同步根与 dev-stack `.env.example`、Compose、当前 owner 文档和 config lint 字典。历史修订记录可保留，current contract zero-reference 必须通过。

BDD-N/A：本阶段只定义配置合法性与 secret 边界；用户可感知的邮件投递行为由 backend-auth `BDD.AUTH.EMAIL.002` 验证。替代 gate 为 owner table tests、`make lint-config`、secret redaction 与 current-scope zero-reference。

### Phase 15: Local AI raw I/O capture config

#### 15.1 Typed config RED/GREEN

在 `backend/internal/platform/config` owner 的表驱动 suite 先写 RED，覆盖 dev/test 缺省开启、staging/prod 缺省关闭、staging/prod 显式开启拒绝、capture 开启但 path 为空拒绝、合法 override，以及相对路径稳定按 resolved `ConfigDir` parent 解析；随后落地 `ai.debugCaptureRawIO` / `ai.debugRawIOPath` typed bindings、env bindings、YAML 默认、effective absolute path 和 validator。

#### 15.2 Dictionary, privacy and removal gates

根与 dev-stack `.env.example` 只声明 `AI_DEBUG_CAPTURE_RAW_IO` / `AI_DEBUG_RAW_IO_PATH`；runtime-config/config dump 不暴露开关、路径或内容；当前 config、env 字典、生产代码删除 `AI_DEBUG_PRINT_RAW_OUTPUT` / `ai.debugPrintRawOutput`，不保留 alias。A3 Phase 16 消费 resolved path 建立 symlink-safe mode-0600 NDJSON recorder。

BDD-N/A：本阶段只定义启动期配置、环境隔离与隐私边界，不产生独立用户行为。替代 gate 为 owner table tests、`make lint-config`、runtime-config negative contract 与旧 key zero-reference。

## 5 验收标准

- [secrets-and-config spec §6 验收标准](../../spec.md#6-验收标准) C-1..C-15 全部成立；C-6 由 A4 builder/handler、B2 generated contract 与 D1 `AppRuntimeProvider` 共同闭环，无平行前端 fetcher；C-14 证明 typed defaults、override、非法值、public allowlist 与旧硬编码清退；C-15 锁定本地 raw capture 的环境默认、生产禁用与非公开边界。
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
| 2026-07-16 | 1.22 | local dev backend/frontend 默认入口随 A2 统一为 10901/10900；A4 同步 APP_LISTEN_ADDR 与 CORS origin 默认。 | local port-conflict feedback |
| 2026-07-14 | 1.18 | 方案 A：统一内容大小 typed defaults、代码缺省、AI 容量 gate、runtime-config 投影与跨 owner BDD。 | runtime size limits recalibration |
| 2026-07-14 | 1.19 | 按奥卡姆剃刀删除跨层重复配置测试与配置专用场景 gate，保留单一 typed contract 和必要业务缺陷回归。 | config test proportionality |
| 2026-07-16 | 1.20 | 新增标准 SMTP username/password/TLS mode 条件合同，删除未消费的 provider API key。 | backend-auth production SMTP |
| 2026-07-16 | 1.21 | 原地 reopen，新增本地 AI raw I/O typed config、生产禁用、独立路径与旧 stderr 开关清退。 | report generation failure investigation |
| 2026-07-13 | 1.17 | 删除无当前消费者的 TargetJob attachment maxBytes config/validator/composition binding，保留 Resume 与 Privacy Export 配额。 | Home paste-only JD intake |
| 2026-07-10 | 1.15 | 删除无正式入口消费者的平行 frontend runtime-config fetch/cache/type/test 包。 | tech-debt pruning |
| 2026-07-10 | 1.14 | 删除零消费者 `Loader.GetDuration` 与旧 bootstrap 合同引用。 | tech-debt pruning |
| 2026-07-10 | 1.13 | 删除 runtime-config 测试 reset export，单测改用生产 forceRefresh 边界。 | tech-debt pruning |
| 2026-07-10 | 1.12 | 统一 feature flag 范围术语并将 context 对齐 spec 2.13；不改变 current baseline 或负向回归输入。 | tech-debt pruning |
| 2026-07-10 | 1.11 | 将 `config/config.yaml` 与 `.env.example` 的 secret 默认值描述收敛为空字符串 / 说明注释。 | tech-debt pruning |
| 2026-07-10 | 1.10 | 将 runtime-config handler 口径从 stub 收敛为当前 session-aware handler + anonymous opt-out default。 | tech-debt pruning |
| 2026-07-07 | 1.8 | Wording cleanup：收敛 feature flag 与 TDD gate 说明为 out-of-scope flag / 既有实现口径，不改变可执行契约。 | product-scope/001 Phase 6.89 |
| 2026-05-05 | 1.6 | L2 深审修正文档 allowlist 范围外口径：`cmd/migrate` 是 B4 迁移 CLI 的合法 env 读取入口；plan 与当前 spec §4.1 / `getenv_boundary.go` 对齐。 | deep reconcile |
| 2026-05-04 | 1.5 | L1 plan-review remediation：补齐当前强制的质量门禁分类，不改变已完成 config/secret/feature flag 范围。 | docs-only L1 remediation |
| 2026-05-03 | 1.4 | 原地 reopen，新增 Phase 8 remediation：按 product-scope v1.2 替换范围外错题本 / 成长中心 / dual-track feature flag baseline。 | secrets-and-config v1.9 |
| 2026-04-30 | 1.3 | L2 code-review remediation：补 prod/staging required config 覆盖与 dev-default runtime override 防线。 | plan-code-review --fix |
| 2026-04-30 | 1.2 | L2 code-review remediation：out-of-scope worker config bindings、AI base URL fail-fast、env_dict code-side binding discovery、runtime-config cold PostHog projection。 | plan-code-review --fix |
| 2026-04-29 | 1.1 | 对齐 spec v1.7：24 项 env key、`async.queueWeights` config-only 字段、PostHog last-known-good 缓存降级、secret 样本只允许临时生成不入文档。 | plan-review remediation |
