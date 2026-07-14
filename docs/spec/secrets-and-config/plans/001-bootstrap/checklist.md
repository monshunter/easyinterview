# Secrets and Config Bootstrap Checklist

> **版本**: 1.18
> **状态**: active
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

> Phase 1-12 的已勾选项只保留为历史交付证据；Phase 13 是当前内容大小配置与代码缺省合同。旧 Phase 中出现的附件配置或散落硬编码不构成当前实现、验收或兼容要求。

## Phase 1: Three-tier config loader 与 redactor

- [x] 1.1 落地 `backend/internal/platform/config/` 包骨架（`loader.go` / `validator.go` / `redactor.go` / `getters.go` / `doc.go`），module path 沿用 B1 锁定的 `github.com/monshunter/easyinterview/backend`，`doc.go` godoc 概述 D-1 三层优先级与 `Get*` API 命名约定
- [x] 1.2 引入 `github.com/knadh/koanf/v2` + `koanf/parsers/yaml` + `koanf/providers/env` + `koanf/providers/file`；`loader.go` 按 D-1 顺序串行合并 `config/config.yaml` → `config/{APP_ENV}.yaml` → env var → runtime secret，禁止并发 `Load` 引发 merge 顺序歧义
- [x] 1.3 落地 `Get*` API（`GetString` / `GetInt` / `GetBool` / `GetSecret`），点路径键名（如 `app.listenAddr`）与 spec §3.1.2 `Config path` 列对齐；错误处理由 validator 集中，不在 getter 内 panic
- [x] 1.4 实现 `RedactedString` 类型与 redactor：`String()` / `GoString()` / `MarshalJSON()` / `MarshalText()` / `Format(...)` 五方法返回 `***`；底层字段不导出；`Reveal()` 是唯一明文路径；redactor 必须覆盖结构化 JSON 日志、`fmt.Errorf %w` 错误包装链、嵌套 config dump 三种路径
- [x] 1.5 启动期 fail-fast validator：`APP_ENV=test` 允许缺 AI / Email / Session secret；`APP_ENV=staging|prod` 必填 secret 缺失返回 error；错误信息列出缺失 key 名（关闭 spec C-2 报错预期）
- [x] 1.6 Phase 1 自检：`go test ./backend/internal/platform/config/...` 通过 `loader_test.go`（四层合并）/ `redactor_test.go`（三路径 `***` 断言）/ `validator_test.go`（test/staging/prod 三档 fail-fast 路径）

## Phase 2: SecretSource + FeatureFlagClient 抽象与 provider 实现

- [x] 2.1 落地 `backend/internal/platform/secrets/`：`SecretSource` 接口签名锁定 `Get(name string) (string, error)`；`EnvSecretSource` 实现 + 注释说明 `os.Getenv` 边界；接入 Phase 1 loader 作为 D-1 第一层
- [x] 2.2 落地 `backend/internal/platform/featureflag/featureflag.go` 接口：`IsEnabled(key, ctx FlagContext) bool` + `Variant(key, ctx FlagContext) string`；`FlagContext` 仅承载 anonymous distinct id / authenticated user public id / app env
- [x] 2.3 落地 `FileFlagProvider`（`config/feature-flags.yaml`，YAML schema `flags: { <key>: { enabled, variant?, public } }`），通过 mtime + 内容 hash ≤30s 热加载，使用 `sync.RWMutex` 保护内部 map，加载失败保留上一次快照并写 warn 日志（关闭 spec C-3）
- [x] 2.4 落地 `PostHogFlagProvider`：仅用 `net/http` 调用自托管 `POST {POSTHOG_HOST}/decide?v=3`，禁止 import `posthog-go` SDK；缓存 ≤30s 与 D-7 对齐；网络错误 / 5xx 返回 last-known-good 内存快照并 warn，无快照时 degraded/error，不静默回退 FileFlagProvider；`POSTHOG_SELF_HOSTED=false` 在 staging/prod 启动 fail-fast（关闭 spec C-4）
- [x] 2.5 Phase 2 自检：`secrets_test.go` 覆盖 `EnvSecretSource.Get` 缺失 error；`file_provider_test.go` 覆盖热加载与启动 race；`posthog_provider_test.go` 通过 `httptest.NewServer` mock `/decide` 命中、5xx/timeout last-known-good、无缓存 degraded/error；`POSTHOG_SELF_HOSTED=false` 在 staging/prod 启动失败用例必跑

## Phase 3: 配置文件骨架与 .env.example 字典对齐

- [x] 3.1 落地仓库根 `config/config.yaml`：覆盖 spec §3.1.2 canonical config schema；secret 字段留空字符串，禁止真实凭证；`auth.sessionCookieName` 固定 `ei_session`（与 ADR-Q1 §3 / spec D-8 一致），不允许任何 env key 覆盖；`async.queueWeights` 默认 `critical:6/default:3/low:1`
- [x] 3.2 落地 `config/dev.yaml` / `config/staging.yaml` / `config/prod.yaml`：env override 文件不含任何 secret；`dev.yaml` 设 `log.level: debug` + `featureFlag.source: file`；`staging.yaml` / `prod.yaml` 设 `featureFlag.source: posthog` + `featureFlag.posthogSelfHosted: true`
- [x] 3.3 落地仓库根 `.env.example`：覆盖 spec §3.1.1 全部 24 项 env key，secret 字段只写说明注释；每行注释标注 Owner subspec 与 prod required（yes/no/conditional），与 spec §3.1.1 表格一一对应；`async.queueWeights` 不进入 env 字典
- [x] 3.4 落地 `config/feature-flags.yaml` current baseline：4 项 P0 flag 为 `report_evidence_v2_enabled` / `report_retry_plan_enabled` / `readiness_signals_enabled` / `ai_fallback_model_enabled`；显式标注 `public: true|false`；`ai_fallback_model_enabled` 必须 `public: false`；三项 out-of-scope key 仅作为负向测试输入
- [x] 3.5 落地 `config/README.md`：覆盖三层优先级（D-1）、5 个文件用途、新增 env key 4 步流程、`RedactedString` 使用示范、`runtime-config` allowlist 边界，包含 spec §3.1 / §4 cross-link
- [x] 3.6 在 A1 已有的根 `.gitignore` 中追加独立段（注释 `# A4 secrets-and-config red lines`）：`*.secret.yaml` / `*.secret.json` / `config/local.*.yaml` / `.env` / `.env.local` / `config/feature-flags.local.yaml`

## Phase 4: Lint / pre-commit hook / `make lint-config`

- [x] 4.1 在 B1 已落地的 `backend/.golangci.yml` 中追加本地可执行规则：优先 `revive` 自定义 rule，必要时落地 `scripts/lint/getenv_boundary.go`（Go AST checker）；allowlist 仅放行 `internal/platform/config/` / `internal/platform/secrets/` / `cmd/{api,migrate}/`，其它包出现 `os.Getenv` lint 失败（关闭 spec C-7）
- [x] 4.2 落地 `scripts/lint/env_dict.py`（或 `.sh`）：解析 `.env.example` + 代码侧 `os.Getenv` / `Get*` 调用 + spec §3.1.1 表，三方求差集；`Makefile` 新增 `.PHONY: lint-config` 并入 `make lint`，缺失 key 必须 fail（关闭 spec C-9 / C-11）
- [x] 4.3 落地 `scripts/git-hooks/pre-commit-secrets.sh`：扫描 `git diff --cached` 命中 `AKIA[0-9A-Z]{16}` / `sk-[A-Za-z0-9]{20,}` / `xox[baprs]-[A-Za-z0-9-]+` 即 fail；错误信息列出文件名 + 行号但不输出命中 secret 字面量；通过 A1 已建立的 hook 入口注册（关闭 spec C-8 第一层）
- [x] 4.4 落地 `scripts/lint/gitleaks.sh` 第二层：调用本地 `gitleaks detect --no-git --redact`；未安装时打印安装提示并 exit 0 不阻塞；`make lint` 调用此脚本；远端 CI secret scan 仅在 A5 触发条件成立后再接入
- [x] 4.5 Phase 4 自检：构造越界 `os.Getenv` 改动 + 删除 `.env.example` 中 `AI_PROVIDER_BASE_URL` + 临时生成命中 secret 正则的假数据三类故意失败场景，确认 `make lint` / `make lint-config` / pre-commit hook 全部拦截；自检后立即 revert，不污染主分支；文档与 fixture 不保留真实形态 secret 样本

## Phase 5: `runtime-config` endpoint 接入与 generated-client handoff

- [x] 5.1 落地 `backend/internal/platform/config/runtime_config.go`：`BuildRuntimeConfig(ctx, session) RuntimeConfig` 严格 allowlist `appVersion` / `defaultUiLanguage` / `analyticsEnabled` / `featureFlags` / `postHogPublicKey`；过滤 `public: false` flag；session + `analytics_opt_in=false` 时 `analyticsEnabled=false` 且不返回 `postHogPublicKey`；secret 字段绝对不进 response
- [x] 5.2 落地 `backend/internal/platform/config/runtime_config_handler.go` handler：`GET /api/v1/runtime-config` 调用 builder 序列化 JSON，支持 C1 session-aware resolver 注入，resolver 缺省时使用 anonymous opt-out 默认；OpenAPI schema 真理源由 B2 持有，本 plan 不修改 B2 文件
- [x] 5.3 D1 `AppRuntimeProvider` 通过 B2 generated `EasyInterviewClient.getRuntimeConfig()` 与 `RuntimeConfig` 类型读取 endpoint；A4 不维护平行前端 fetch/cache/type 包
- [x] 5.4 在 builder 顶部 cross-link 到 [B2 openapi-v1-contract spec](../../../openapi-v1-contract/spec.md)，明确 A4 持有 builder/allowlist，B2 持有 schema/generated TS，D1 持有正式 React consumer
- [x] 5.5 Phase 5 自检：`runtime_config_test.go` 覆盖 spec C-6 allowlist；D1 `AppRuntimeProvider` focused tests覆盖 generated-client runtime/auth bootstrap、失败态与 refresh；本地 `curl /api/v1/runtime-config` smoke 通过

## Phase 6: Verification + handoff

- [x] 6.1 AC C-1..C-5 复跑：C-1（三层合并 `app.listenAddr=:9090`、`log.level=debug`）/ C-2（prod 缺 `SESSION_COOKIE_SECRET` 退出码非 0 + stderr 含 `missing required secret`）/ C-3（修改 `feature-flags.yaml` 30s 内热加载）/ C-4（mock `/decide` + `POSTHOG_SELF_HOSTED=false` 双场景）/ C-5（`RedactedString` 三路径 `***`）；命令日志贴入工作日志
- [x] 6.2 AC C-7..C-12 复跑：C-7（越界 `os.Getenv`）/ C-8（`AKIA*` / `sk-*` / `xox*` 三正则临时生成）/ C-9（`.env.example` 删 key）/ C-10（缺 `AI_PROVIDER_*` fail-fast；`APP_ENV=test` 仍走 stub）/ C-11（schema 分类错位 lint 失败）/ C-12（`async.queueWeights` 缺失或非正数 fail-fast）；故意失败 case 验证后立即 revert，不污染主分支；命令日志贴入工作日志
- [x] 6.3 AC C-6 验证 + handoff：A4 builder/handler、B2 schema/generated client/types 与 D1 `AppRuntimeProvider` 构成当前完整链路；前端无第二套 runtime-config client
- [x] 6.4 文档与 INDEX 收口：`config/README.md` Header 完整 + 内容覆盖三层优先级 / 5 文件用途 / 新增 key 4 步流程 / RedactedString 示范 / runtime-config allowlist；`docs/spec/secrets-and-config/plans/INDEX.md` 把本 plan 切到 Completed；`docs/spec/INDEX.md` 中 `secrets-and-config` 行 Header 与 spec 一致；`/sync-doc-index --check` 通过
- [x] 6.5 风险扫尾：按 plan §5 风险表逐条复核 redaction 三路径覆盖、koanf 合并顺序锁定、hot reload race、`.env.example` 与代码侧 env key 对齐、prod fail-fast 与 supervisor restart loop 提示；任一项缺证据本 plan 不切 Completed

## Phase 7: L2 review remediation

- [x] 7.1 修复 runtime entrypoint env / secret binding 对齐：prod + 完整 env 注入时 current backend loader 校验通过；缺失 secret 仍 fail-fast 并列出 env key；backend-runtime-topology v1.0 后不再保留单独 runtime entrypoint
- [x] 7.2 修复 `AI_PROVIDER_BASE_URL` fail-fast：non-test AIClient-enabled 启动路径缺 base URL 或 API key 任一项都失败；`APP_ENV=test` 仍可缺 AI provider 配置
- [x] 7.3 修复 `scripts/lint/env_dict.py` code-side key 发现：`EnvBindings` / `SecretBindings` 字面量纳入三方求差集；binding map 声明但 `.env.example` 缺 key 的 pytest 必须失败
- [x] 7.4 修复 runtime-config cold PostHog flag projection：首次请求按 `FlagContext` evaluation 后返回 public flags、过滤 operator-only flags；PostHog provider 初始化携带 public allowlist；D-4 业务接口不扩大
- [x] 7.5 修复 prod/staging required config 覆盖：`validator.go` 校验 spec required/conditional P0 keys，database/redis/object storage 在 staging/prod 必须有 runtime override，缺 PostHog host / email provider / AI model profile path 等 fail-fast；补 focused tests 覆盖失败与通过路径；验证: 2026-04-30 `go test ./internal/platform/config -run 'TestValidateProd(AllSecretsPasses|RejectsDevDefaultDeploymentDependencies)' -count=1` 与 `go test ./internal/platform/... ./cmd/api -count=1`

## Phase 8: product-scope v1.2 feature flag remediation

- [x] 8.1 Red: 调整 runtime-config / feature flag tests 后，out-of-scope `mistake_book_export_enabled` / `growth_dashboard_v1_enabled` / `mock_session_dual_track_enabled` 进入 public runtime config 时必须失败
  - 2026-05-03: 更新 `TestBuildRuntimeConfigAllowlistAndOptOut` 后，`cd backend && go test ./internal/platform/config -run TestBuildRuntimeConfigAllowlistAndOptOut -count=1` exit 1，失败于三项 product-scope out-of-scope flag 仍透传到 public runtime config。
- [x] 8.2 Green: 更新 `config/feature-flags.yaml` 与 tests，使用当前 report/readiness flag；2026-07-12 删除已经失去消费者的 practice hint/assistance flags
  - 2026-05-03: `config/feature-flags.yaml` baseline 替换为当前六项 P0 flag；`BuildRuntimeConfig` 改为正向 current-flag allowlist，三项 out-of-scope flag 即使由上游 snapshot 返回也不会进入 runtime config。
- [x] 8.3 Verify: `make lint-config`、focused runtime-config tests 通过；repo 搜索确认实现侧无三项 out-of-scope feature flag key
  - 2026-05-03: `make lint-config` pass（gitleaks 本地未安装，按脚本第二层扫描策略提示并 exit 0）；`cd backend && go test ./internal/platform/config -run TestBuildRuntimeConfigAllowlistAndOptOut -count=1` pass；`cd backend && go test ./internal/platform/config ./internal/platform/featureflag -count=1` pass；`rg -n "mistake_book_export_enabled|growth_dashboard_v1_enabled|mock_session_dual_track_enabled" config backend/internal/platform backend/cmd/api frontend/src/lib/runtime-config scripts/lint/env_dict.py -g '!**/*_test.go'` 无实现侧命中。

## Phase 10: Unused duration getter removal

- [x] 10.1 删除零消费者 `Loader.GetDuration`、自测分支与 package doc 引用；验证：production `deadcode` RED/GREEN、config tests/staticcheck、config lints、owner docs gates。
  <!-- verified: 2026-07-10 method=unused-config-duration-getter-removal evidence="Production deadcode RED identified Loader.GetDuration as test-only. Removed the method, getter fixture/assertion and package/current-contract references. Config tests, staticcheck, code inventory and make lint-config PASS." -->

## Phase 11: Parallel frontend runtime-config client removal

- [x] 11.1 删除 main-entry 不可达且仅自身测试消费的平行 frontend runtime-config client 包，收敛到 generated client + `AppRuntimeProvider`；验证 source contract RED/GREEN、focused/full frontend、config/codegen、owner contexts 与 docs/diff/pruning gates。
  <!-- red: 2026-07-10 method=main-entry-reachability+pruning-contract evidence="The TypeScript main-entry graph reported six unreachable non-test files, of which the two hand-written runtime-config modules were consumed only by their own five-test file. The focused pruning suite failed only the new package-absence contract while the prior nine tests passed." -->
  <!-- verified: 2026-07-10 method=parallel-frontend-runtime-config-client-removal evidence="Deleted the fetch/cache implementation, duplicate hand-written types and five self-only tests without a wrapper or replacement. AppRuntimeProvider continues to use generated EasyInterviewClient/RuntimeConfig and getRuntimeConfig. Focused frontend passes 14 tests; full frontend passes 136 files/836 tests plus typecheck/build; backend config/secrets/featureflag packages, make lint-config and make codegen-check pass. The post-delete graph reports 264 source files, 121 runtime-reachable files and only four unreachable non-test generated B2/B3 contract assets. A4, D1 and product contexts, git diff check and pruning surface pass with real_residuals=0. No Bug/retrospective report, environment restart or data cleanup was needed." -->

## Phase 12: TargetJob attachment maxBytes config removal

- [x] 12.1 RED: config schema、validator 与 backend API composition tests 要求 maxBytes 当前集合仅含 `resume` / `privacyExport`，并固定默认值 10MB / 5MB；旧 TargetJob attachment key 仍存在时测试失败。
  <!-- verified: 2026-07-13 method=default-config-red evidence="focused test failed because upload.maxBytes.targetJobAttachment still resolved to 10485760 instead of being absent" -->
- [x] 12.2 GREEN: 删除 `config/config.yaml`、platform config validator/fixtures/tests 与 backend API composition 中的旧 TargetJob attachment maxBytes binding，不增加 alias、fallback 或兼容读取。
  <!-- verified: 2026-07-13 method=config+composition-green evidence="canonical config, validator required paths, cmd/api purpose-limit composition and all config fixtures now expose only resume=10MB/privacyExport=5MB; no alias or fallback" -->
- [x] 12.3 REGRESSION-GATE: `make lint-config`、platform config focused tests 与 backend API composition tests 通过；Resume/Privacy 非正数仍 fail-fast，presign TTL 和 10MB/5MB 默认值不变。
  <!-- verified: 2026-07-14 method=lint+focused-regression evidence="make lint-config passed after shortening a pre-existing SHA-256 evidence string that gitleaks misclassified; go test ./internal/platform/config ./cmd/api -count=1 passed and retained TTL plus 10MiB/5MiB validation." -->
- [x] 12.4 ZERO-REF/BDD-GATE: active `config/`、platform config、backend API composition 与 A4 current docs 对旧 key 零命中；合法历史/显式负测排除。BDD 不适用，以 Red/Green focused tests、config lint、typed composition 与 zero-reference 作为替代 gate。
  <!-- verified: 2026-07-14 method=zero-reference+substitute-gate evidence="Production config, platform config, and cmd/api scopes have zero targetJobAttachment key matches; remaining matches are the explicit validator/integration negative tests and historical RED evidence. Phase 12 lint and focused typed composition tests pass." -->

## Phase 13: Runtime content size defaults and boundary alignment

- [x] 13.1 RED-CONFIG: 为全部 size key 增加缺 key、合法 override、显式 `0`/负数、`paste > extracted`、`message > session` 用例；在 typed defaults 与 validator 未落地前测试按预期失败。
  <!-- verified: 2026-07-14 method=focused-red evidence="go test ./internal/platform/config -run TestContentLimits -count=1 failed to compile because Loader.ContentLimits, config.ContentLimits, and DefaultContentLimits do not yet exist; the new tests cover every key with missing/default, override, zero, negative, and both cross-field invalid combinations." -->
- [x] 13.2 GREEN-CONFIG: 落地统一 typed code defaults 与 YAML 镜像：HTTP 10MiB、Resume upload 10MiB、Privacy Export 5MiB、Resume active 10、Resume extracted/paste 384KiB、TargetJob raw 96KiB、Practice message 32KiB/session 256KiB、Report framed 896KiB、AI response 4MiB。
  <!-- verified: 2026-07-14 method=typed-defaults-green evidence="ContentLimits and DefaultContentLimits now resolve missing keys, honor YAML overrides, reject all explicit non-positive values plus invalid paste/extracted and message/session combinations; focused and full platform config tests, make lint-config, and git diff --check pass." -->
- [x] 13.3 BACKEND-GREEN: 将全局 HTTP body、upload/resume/target-job/practice/report consumers 与四个 AI provider adapters 改为注入配置；UTF-8 bytes 的 limit 接受、limit+1 拒绝且不调用 provider；删除重复生产常量。
  <!-- verified: 2026-07-14 method=focused+full+race evidence="All injected consumers pass exact/+1 tests; backend go test ./... and selected config/practice/review/provider race packages pass." -->
- [x] 13.4 REPORT-CAPACITY: A3 profile code fallback 与 canonical catalog 一致；测试锁定 `917504 + 2048 + 6144 = 925696 < 1000000`，真实 62,397-byte 失败样本可进入 provider 路径，TPM 不再作为单请求 hard cap。
  <!-- verified: 2026-07-14 method=in-memory-capacity+P0.056 evidence="62,397 and 917,504 bytes each reach provider once; 917,505 is terminal before provider; formula/profile fallback tests pass." -->
- [x] 13.5 CONTRACT: OpenAPI `RuntimeConfig.contentLimits` 只含五项 public limit；backend builder、fixture、generated Go/TS 同步；内部 report/HTTP/provider/profile 上限不泄漏。
  <!-- verified: 2026-07-14 method=OPENAPI-006-exact-audit evidence="Exact 1 breaking + 8 additive audit preserved; 37/10 lint, fixtures, generated artifacts and 52 wrapper tests pass. Final post-commit codegen-check remains owned by 13.8." -->
- [x] 13.6 FRONTEND: Resume upload/paste、Home JD raw text、Practice message/session 校验统一消费 runtime config 并按 UTF-8 bytes 判断；删除 2MiB 与 rune/character 本地真理源，limit/limit+1 focused tests 通过。
  <!-- verified: 2026-07-14 method=vitest+build+BDD evidence="Frontend full 126 files/1018 tests and production build pass; P0.015/P0.046/P0.081 exact/+1 assertions pass." -->
- [x] 13.7 BDD-GATE: [`bdd-checklist.md`](./bdd-checklist.md) 中 `E2E.P0.010`、`E2E.P0.046`、`E2E.P0.081`、`E2E.P0.056` 全部通过并记录当前证据。
  <!-- verified: 2026-07-14 method=serial-scenario-run evidence="Fresh P0.010/P0.015/P0.034/P0.035/P0.046/P0.056/P0.058/P0.081 pass; P0.046 isolated PostgreSQL residual=0." -->
- [ ] 13.8 REGRESSION/POST-PASS: config/profile/provider/domain/OpenAPI/frontend focused/full gates、`make lint-config`、`make codegen-check`、旧硬编码 negative search、`sync-doc-index --check`、`make docs-check`、`git diff --check` 全部通过；完成 Bug 记录评估与 retrospective。
