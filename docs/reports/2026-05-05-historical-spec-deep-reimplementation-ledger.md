# Historical Spec Deep Reimplementation Ledger

> **日期**: 2026-05-05
> **验证人**: Codex

**关联计划**: [historical-spec-implementation-review](../spec/historical-spec-implementation-review/plans/001-implement-review-runway/plan.md)

## 1 执行原则

本台账用于记录 2026-05-05 重新执行 historical spec/plan 收口的逐项证据。执行时忽略各目标 plan/checklist 既有 Header 状态、完成标记和历史 PASS 记录，只把它们作为待复核条目来源。

每个目标必须完成以下检查后才能判定为本轮 PASS：

1. 读取当前 spec / plan / checklist / context 原文。
2. 将 plan/checklist 条目映射到当前代码、配置、生成物、脚本、README 或报告 artifact。
3. 针对旧产品范围、旧 route、旧 API/DB/event/config/AI 假设执行 grep / 结构化检查。
4. 运行对应 focused tests、lint、drift gate 或 smoke；必要时构造负向验证。
5. 发现漂移则原地修复，再重新运行目标验证。

### 1.1 后续执行规章（2026-05-05 反馈纳入）

用户反馈指出：diff 小、历史 gate 通过、历史 checklist 完成，都不能证明新产品 spec / 交互重构后的实现仍然正确。本轮后续目标从 `event-and-outbox-contract` 开始必须追加以下执行规章：

1. **不得把 diff 大小当证据**：每个 target 必须从当前 `product-scope` / `docs/ui-design` / `ui-design` 不变量反向审计 artifact；diff 只说明改动量，不说明语义对齐。
2. **旧 gate 只算必要条件**：若 gate 只覆盖结构数量或历史断言，必须补做语义审计；发现 gate 未覆盖新版语义时，优先把反馈固化为 lint / unit test / negative fixture，再继续下一个 target。
3. **每项必须有 artifact-level 反查**：除读取 plan/checklist 外，必须直接解析或搜索真理源、生成物、fixtures、baseline、runtime config、DDL、tests、README 与脚本入口；不得只用 checklist 勾选状态推断完成。
4. **每项必须执行旧口径负向搜索**：至少覆盖旧一级入口、旧 route/tag/schema/table/event/job/config flag、旧 AI model/provider 假设、旧 `feature_key` / `featureKey` 口径、旧 mistakes/growth/drill/voice 独立模块口径。
5. **用户反馈必须立即落地为规则**：如果审查中暴露出“工作方式或 gate 不够深”的问题，必须先更新本台账执行规章或对应 target gate，再继续推进后续计划。

## 2 Scope Inventory

| 分类 | 目标 | 本轮状态 | 依据 |
|------|------|----------|------|
| review target | `repo-scaffold/001-bootstrap` | PASS after fix | 重新验证 context；artifact-level L2 发现并修复 README / dry-run 漂移 |
| review target | `shared-conventions-codified/001-bootstrap` | PASS after fix | context discovery、truth source、Go/TS generated output、lint/codegen drift gate 均已重新核对 |
| review target | `shared-conventions-codified/002-codegen-pipeline` | PASS after fix | context version、AI vocabulary bridge、cross-language parity、旧口径搜索均已重新核对 |
| review target | `openapi-v1-contract/001-bootstrap` | PASS after fix | 34 operation / 12 tag contract、generated Go/TS、OpenAPI inventory 与 current UI/product scope 已重审 |
| review target | `openapi-v1-contract/002-fixtures-and-mock-source` | PASS after fix | fixtures / prototype-baseline / generated examples / provenance / privacy gate 已重审并修复模型硬编码 |
| review target | `openapi-v1-contract/003-breaking-change-gate` | PASS | baseline / additive-only wrapper / privacy whitelist / diff gate 已重审 |
| review target | `event-and-outbox-contract/001-bootstrap` | PASS after fix | 16 events、10 jobs、generated artifacts、email/privacy 红线、B2/B4 对齐均已重审 |
| review target | `db-migrations-baseline/001-bootstrap` | PASS after fix; DB-backed migrate-check deferred to A2/global | 重新验证 context；DDL / enum / privacy / lint 已 artifact-level L2 |
| review target | `secrets-and-config/001-bootstrap` | PASS after fix | 重新验证 context；env/config/flag/secret/runtime-config/lint boundary 已 artifact-level L2 |
| review target | `ai-gateway-and-model-routing/001-aiclient-and-profile-bootstrap` | PASS after fix | 重新验证 context；AIClient/profile/stub/openai-compatible/observability/config 已 artifact-level L2 |
| review target | `local-dev-stack/001-bootstrap` | PASS after fix | 重新验证 context；compose/doctor/Make lifecycle/volume/port conflict 已 artifact-level L2 |
| review target | `ci-pipeline-baseline/001-local-quality-gates` | PASS after fix | 重新验证 context；docs/codegen/lint gates、deferred CI guard、B3 drift path 已 artifact-level L2 |
| L1 target | `engineering-roadmap/001-decompose-subspecs` | PASS | 重新验证 context；Phase 3 仅为后续 child 创建规则，INDEX 与真实 spec 目录未创建候选 P0 workstream |
| excluded | `ai-gateway-and-model-routing/002-tools-streaming-and-stt` | draft-gated | Header draft；不进入本轮 implementation |
| excluded | `historical-spec-implementation-review/001-implement-review-runway` | docs-only orchestration | 本轮执行入口，不作为业务/代码 implementation target |

## 3 Current Product / UI Semantic Invariants

旧版 plan gate 只作为必要条件；本轮所有 artifact 还必须符合以下 current product / UI 不变量。

| ID | 不变量 | 当前真理源 | 审查影响 |
|----|--------|------------|----------|
| P-1 | 一级入口只允许 `首页 / 岗位推荐 / 模拟面试 / 简历 / 复盘` | `product-scope` D-4、`docs/ui-design/INDEX.md` | README、OpenAPI、fixtures、runtime config、UI prototype 不得恢复报告 / 成长 / 当前岗位 / 语音等独立一级入口 |
| P-2 | 面试是一场完整 session；不提供入口前热身、反问专练、单题深钻、独立 Drill | `product-scope` D-5 / §6.8 | shared enum、OpenAPI schema、fixtures、events、DB check、AI profiles 不得恢复旧 practice mode / goal |
| P-3 | 语音只是 `practice` 或 `debrief` 的形式；不得有独立 `voice` route / module | `product-scope` D-6、UI contract | route/API/config/feature flag/artifact 不得暴露 `voice` 独立入口；允许 `practice?mode=voice&modality=voice` |
| P-4 | 报告只有 session-scoped Dashboard；无上下文不得展示假报告 | `product-scope` D-7、`report-dashboard.md` | OpenAPI、fixtures、UI prototype 和 docs 必须携带 `sessionId` / 会话上下文 |
| P-5 | 错题价值只在报告题目回顾与本轮复练中承载；不恢复独立错题队列 | `product-scope` D-8 | OpenAPI 不得有 `Mistakes`；DB 不得有 `mistake_entries`；events 不得有 `mistake.*`；fixtures 不得有 drill retry |
| P-6 | 复盘面向真实面试；文本和语音添加共享同一份记录 | `product-scope` D-9、`review-module.md` | OpenAPI / fixtures / events / DB 不得把模拟报告错题自动写成复盘；UI/contract 要保留 shared debrief record 语义 |
| P-7 | AI 生成物必须可追溯 prompt / rubric / model / feature flag / data source metadata，但不得硬编码厂商模型假设 | `product-scope` D-10、A3/F3/B1 | AI gateway、OpenAPI provenance、events、DB `ai_task_runs` 必须保留 provenance；A3 不得依赖具体 vendor model naming |
| P-8 | 未创建真实 spec 的 P0 workstream 不进入 INDEX，也不在本轮创建 child | `engineering-roadmap` D-5 / §5.2 | `engineering-roadmap/001` Phase 3 只能是 future creation rule，不是 implementation target |

本轮已执行 semantic baseline：

- `node --test ui-design/ui-design-contract.test.mjs`：15 tests PASS。
- `rg` 抽查 `product-scope` / `docs/ui-design` / `ui-design/src` 当前约束：确认新产品不变量以当前文档和静态原型为准。

## 4 Review Ledger

### 4.1 `repo-scaffold/001-bootstrap`

| 检查面 | 本轮证据 | 结论 |
|--------|----------|------|
| plan/checklist 原文 | 读取 `repo-scaffold/spec.md`、`001-bootstrap/plan.md`、`checklist.md`、`context.yaml`、`plans/INDEX.md`；不采信完成标记，只用条目建立 artifact map | PASS |
| 根目录容器与 README | `backend/`、`frontend/`、`openapi/`、`migrations/`、`scripts/`、`test/`、`deploy/`、`shared/`、`config/` 9 个根容器均有 README；`test/README.md` 明确 `test/scenarios/` 尚未落地 | PASS |
| `.editorconfig` / `.gitignore` / `.tool-versions` | `.editorconfig` 锁 UTF-8/LF/末行换行/Go tab；`.gitignore` 覆盖 Go/Node/Python/secret/local config；`.tool-versions` 声明 golang/nodejs/pnpm/python | PASS |
| Makefile target 与 help | `make help` 列出 A1 target 以及后续 owner 扩展 target；`make -n dev-up` 不再实际调用 docker；dry-run safety 由新增 `scripts/lint/makefile_dry_run_test.py` 锁定 | PASS after fix |
| git hooks 与 bootstrap | `make install-hooks` 实际安装 `.git/hooks/pre-commit` / `commit-msg` symlink；hook 文件保持 A1 锁定路径并允许 A4 secret scan 扩展；`bash scripts/bootstrap.sh` 输出 declared/current toolchain | PASS |
| 根 README 与当前仓库事实 | 根 README 64 行，≤80；结构表含 9 个 A1 根容器；修订 P0 主闭环为 `目标面试规划`，`codegen` / `codegen-check` 明确 B3 events/jobs；`frontend/README.md` 不再写旧 `目标岗位工作台` | PASS after fix |
| 修复 | 修复 `README.md` codegen 说明、P0 闭环用词、`frontend/README.md` 用词；修复 `deploy/dev-stack/Makefile` 中 dry-run 会执行 docker 的递归 make 语义；新增 dry-run 回归测试 | FIXED |
| 验证命令 | `python3 -m unittest scripts/lint/makefile_dry_run_test.py` PASS；`validate_context.py --context repo-scaffold/... --target repo` PASS；`make help` PASS；`make install-hooks` PASS；`bash scripts/bootstrap.sh` PASS（本机 locale warning 与工具 patch 版本差异为环境提示） | PASS |

### 4.2 `shared-conventions-codified/001-bootstrap` + `002-codegen-pipeline`

| 检查面 | 本轮证据 | 结论 |
|--------|----------|------|
| plan/checklist 原文 | 读取 `shared-conventions-codified/spec.md`、`001-bootstrap/plan.md`、`001-bootstrap/checklist.md`、`002-codegen-pipeline/plan.md`、`002-codegen-pipeline/checklist.md`、两个 `context.yaml`；不采信历史完成标记，只按条目重建 artifact map | PASS |
| context discovery | `001-bootstrap/context.yaml` 原先 default target 指向 docs 且未包含运行时代码包；已改为 `backend`，显式纳入 `shared/conventions.yaml`、Go/TS generated output、generator、lint scripts、根 `Makefile`；`002-codegen-pipeline/context.yaml` `specVersion.to` 更新到 1.8 | PASS after fix |
| `shared/conventions.yaml` truth source | 当前 14 enum、9 error codes、6 job statuses；`PracticeMode=assisted/strict/debrief_replay`、`PracticeGoal=baseline/retry_current_round/next_round/debrief`、`QuestionReviewStatus=open/queued_for_retry/resolved`，与 current product scope 保持一致 | PASS |
| Go generated artifacts | 复核 `backend/internal/shared/types`、`errors`、`idx`、`ai`；`make codegen-conventions` 后生成物无 diff；`go test ./cmd/codegen/conventions ./internal/shared/... -count=1` 与 `go vet ./...` 通过 | PASS |
| TS generated artifacts | 复核 `frontend/src/lib/conventions`、`frontend/src/lib/ids`；`pnpm exec tsc --noEmit`、`pnpm test src/lib/conventions src/lib/ids` 通过，5 files / 36 tests | PASS |
| AI vocabulary drift gate | `scripts/lint/conventions_yaml.py` 已要求 `model_profile_name/version`、`model_family/id`、`fallback_chain`、`route`、`validation_status`、`output_schema_version`、`prompt_version`、`rubric_version`、`language`、`feature_flag`、`data_source_version`；`scripts/lint/error_codes.py` 确认 error code owner boundary clean | PASS |
| 旧 enum / feature_key / model 假设搜索 | 在 `shared/conventions.yaml`、`backend/internal/shared`、`frontend/src/lib/conventions`、`frontend/src/lib/ids` 搜索 `MistakeStatus`、旧 practice mode、`feature_key`、具体 vendor/model token，无命中 | PASS |
| 修复 | 修复两个 context；修复 `scripts/lint/conventions_yaml_test.py`，用真实产品枚举替代 `Enum1..Enum13` 假样本，并新增拒绝 `MistakeStatus` 与旧 `PracticeMode` 值的负向测试 | FIXED |
| 验证命令 | `make lint-conventions && python3 -m unittest scripts/lint/conventions_yaml_test.py scripts/lint/error_codes_test.py scripts/lint/conventions_drift_test.py` PASS；`make codegen-conventions && git diff --exit-code -- shared/conventions.yaml backend/internal/shared/types backend/internal/shared/errors backend/internal/shared/idx backend/internal/shared/ai frontend/src/lib/conventions frontend/src/lib/ids` PASS；`validate_context.py` 对 001/002 `--target backend` PASS | PASS |

### 4.3 `openapi-v1-contract/001` + `002` + `003`

| 检查面 | 本轮证据 | 结论 |
|--------|----------|------|
| plan/checklist 原文 | 读取 `openapi-v1-contract/spec.md`、`001-bootstrap` / `002-fixtures-and-mock-source` / `003-breaking-change-gate` 的 `plan.md`、`checklist.md`、`context.yaml`；不采信 completed/checkmark，只按当前 spec 重建 contract / fixtures / baseline artifact map | PASS |
| 12 tag / 34 operation contract | 结构化解析 `openapi/openapi.yaml`：12 tag 顺序为 `Auth`、`Uploads`、`Profile`、`Resumes`、`TargetJobs`、`PracticePlans`、`PracticeSessions`、`Reports`、`ResumeTailor`、`Debriefs`、`Jobs`、`Privacy`；34 operation 与 spec §3.1.1 一致；`sessionCookie` 为唯一 P0 security scheme，`Authorization: Bearer` 仅作为“不是 P0 默认”的否定说明出现；本轮新增 `openapi_inventory.py` product-scope 语义断言，避免只靠 diff 大小 / 结构数量判断 | PASS after fix |
| current product schema 对齐 | `FeedbackReport` 含 `sessionId` / `targetJobId` / `questionAssessments` / `retryFocusTurnIds`，无独立 mistakes queue；`QuestionAssessment` 使用 `reviewStatus` + `includedInRetryPlan`；`TargetJob` 使用 `openQuestionIssueCount`；`PracticeMode=assisted/strict/debrief_replay`、`PracticeGoal=baseline/retry_current_round/next_round/debrief`、`QuestionReviewStatus=open/queued_for_retry/resolved`；`ReportNextAction.type=retry_current_round/next_round/review_evidence`；`JobType` 仅 7 个 B2 API-facing subset；`Debrief` 保持真实面试复盘且 P1 感谢信 / follow-up 字段 optional/hidden | PASS |
| Go/TS generated API | 复核 `backend/internal/api/generated/{types,server,spec}.gen.go` 与 `frontend/src/api/generated/{types,client,spec}.ts`；`make codegen-check` 通过；`go test ./cmd/codegen/openapi ./internal/api/generated -count=1` 通过；`pnpm exec tsc --noEmit` 通过 | PASS |
| fixtures + prototype mapping + example projection | `openapi/fixtures/` 34 个 operation 全覆盖；`requestPrivacyExport` 为 `501 + PRIVACY_EXPORT_NOT_AVAILABLE`；`deleteMe` / `requestPrivacyDelete` 均为 `privacy_delete` job；`make sync-fixtures-from-prototype` 与 `make render-openapi-fixture-examples` 通过；`openapi/.generated/openapi-with-fixtures.yaml` 已由 fixtures 重渲染 | PASS after fix |
| provenance / privacy | `validate-fixtures` 重新校验 schema、provenance、privacy allowlist、UUIDv7；本轮发现 fixtures 把 `modelId` 写死为 `openrouter:anthropic/claude-sonnet-4.6` / `primary-llm:m4.7-sonnet-2026q1`，旧 gate 未拦截；已改为 provider-neutral `model-profile:contract.default` / `model-profile:prototype-baseline.default`，并新增 validator 负向规则与测试 | PASS after fix |
| baseline + breaking-change gate | `openapi/baseline/openapi-v1.0.0.yaml` 与当前 12 tag / 34 operation freeze 对齐；`make openapi-diff` 返回 breaking/additive/informational 均为 0；unit tests 覆盖 inventory、composition diff、privacy export whitelist history gate | PASS |
| 旧 UI/product scope 搜索 | 在 `openapi/openapi.yaml`、baseline、fixtures、generated examples、Go/TS generated API 中搜索 `Mistakes` / `Growth` / `listMistakes` / `retestMistake` / `getGrowthOverview` / `MistakeEntry` / `MistakeStatus` / 旧 practice mode / 旧 drill 字段 / vendor model token，无命中（否定说明中的 `Authorization: Bearer` 不作为旧 scheme 生效） | PASS |
| 修复 | 修复 `scripts/lint/validate_fixtures.py`、`validate_fixtures_test.py`、`scripts/codegen/_generate_default_fixtures.py`、`scripts/codegen/sync_fixtures_from_prototype.py`、`openapi/fixtures/README.md`、相关 fixtures 与 `openapi/.generated/openapi-with-fixtures.yaml`，锁住 provider-neutral fixture provenance；增强 `scripts/lint/openapi_inventory.py` / `openapi_inventory_test.py`，把当前 product-scope 语义固化为 lint 断言（禁旧 Mistakes/Growth/Voice/Drill、旧 practice values、旧 mistake 字段、非 session-scoped report、JobType subset 漂移） | FIXED |
| 验证命令 | `validate_context.py` 对 001/002/003 `--target contract` PASS；`python3 -m unittest scripts/lint/openapi_inventory_test.py && make lint-openapi` PASS；`make lint-openapi && make validate-fixtures && make openapi-diff` PASS；`python3 -m unittest scripts/lint/openapi_inventory_test.py scripts/lint/openapi_diff_test.py scripts/lint/validate_fixtures_test.py scripts/lint/validate_fixtures_cli_test.py scripts/codegen/render_openapi_fixture_examples_test.py scripts/codegen/sync_fixtures_from_prototype_test.py` 63 tests PASS；`make codegen-check` PASS；`make docs-openapi` PASS | PASS |

### 4.4 `event-and-outbox-contract/001-bootstrap`

| 检查面 | 本轮证据 | 结论 |
|--------|----------|------|
| plan/checklist 原文 | 读取 `event-and-outbox-contract/spec.md`、`001-bootstrap/plan.md`、`checklist.md`、`context.yaml`；不采信 completed/checkmark，只按当前 16-event / 10-job 语义重建 artifact map | PASS |
| context validation | `context.yaml` 原先 `specVersion.to=1.5`，而当前 spec/plan/checklist 均为 1.6；已修为 1.6 并重新 `validate_context.py --target backend` 通过 | PASS after fix |
| 16 events / envelope | 结构化解析 `shared/events.yaml`：16 events、7 domains（target/practice/report/resume/debrief/source/privacy），无 `mistake.*` domain；envelope 8 字段含 optional soft-required `traceId`；`report.generated.questionIssueCount` 与 `debrief.completed.practiceFocusCount` 为当前字段 | PASS |
| 10 jobs / API-facing subset | `shared/jobs.yaml` 10 canonical job types；B3 `apiFacingSubset` 7 项与 B2 OpenAPI `JobType` enum 完全一致；`source_refresh` / `embedding_upsert` / `email_dispatch` internal-only，不进入 API-facing subset | PASS |
| Go/TS/schema/baseline generated artifacts | 复核 `backend/internal/shared/events`、`backend/internal/shared/jobs`、`frontend/src/lib/events`、`frontend/src/lib/jobs`、`shared/events/schemas`、`shared/events/refs`、baseline manifests；`make codegen-events` 后 generated/baseline/schema 路径无 diff | PASS |
| email dispatch privacy redline | Go / TS 生成 `BuildEmailDispatchPayload` / `buildEmailDispatchPayload`，只允许 `authChallengeId` / `userId` / `templateKey` / `locale` / `deliverySecretRef` / `dedupeKey`，拒绝 `rawMagicLinkToken` / `magicLinkUrl` / `recipientEmail` / `emailBody` 等红线字段 | PASS |
| B2/B4 对齐 | B2 `JobType` enum 等于 B3 `apiFacingSubset`；B4 migration `async_jobs.job_type` check constraint 含全部 10 canonical values；`outbox_events` 含 `publish_attempts` / `next_attempt_at` / `locked_at` / `last_error_code` / `last_error_message` 与 pending due 复合索引 | PASS |
| 旧口径 / 模型假设搜索 | 在 `shared/events`、`shared/jobs`、Go/TS generated events/jobs 中搜索旧 `mistake.*`、`MistakeStatus`、`mistakeCount`、旧 practice mode、`feature_key`、vendor/model tokens；仅 email 红线字段以 forbidden-list 形式存在；无旧实现口径 | PASS after fix |
| 修复 | 修复 `context.yaml` specVersion；修复 `frontend/src/lib/events/events.test.ts` 与 `shared/events/__fixtures__/envelopes.json` 中旧模型测试值；增强 `scripts/lint/lint_events.py` / `lint_events_test.py`，事件 contract fixtures/tests 出现 vendor/model token 会失败 | FIXED |
| 验证命令 | `make codegen-events && python3 scripts/lint/events_inventory.py shared/events.yaml shared/jobs.yaml shared/conventions.yaml && make lint-events` PASS；`python3 -m pytest scripts/lint/events_inventory_test.py scripts/lint/lint_events_test.py -q` 41 passed；`cd backend && go test ./cmd/codegen/events ./internal/shared/events ./internal/shared/jobs -count=1` PASS；`pnpm test src/lib/events src/lib/jobs` 9 tests PASS | PASS |

### 4.5 `db-migrations-baseline/001-bootstrap`

| 检查面 | 本轮证据 | 结论 |
|--------|----------|------|
| plan/checklist 原文 | 重新读取 `spec.md` v1.8、`001-bootstrap/plan.md`、`checklist.md`、`context.yaml`；忽略 completed/checkmark，把 Phase 1-5 重新映射到 `migrations/`、`backend/cmd/migrate`、`backend/internal/migrations`、`scripts/lint` 与 Make targets。发现 context/plan 仍写旧 spec v1.6 口径，已修到 v1.8 | PASS after fix |
| baseline DDL / table inventory | 直接审 `migrations/000001_create_baseline.up.sql` / `.down.sql`：当前创建 26 应用表 + 3 auth 支撑表 + `schema_backfills`；`schema_migrations` 由 `golang-migrate` 管理；不存在 `mistake_entries`；`target_jobs.open_question_issue_count`、`feedback_reports.session_id`、`question_assessments.review_status` / `included_in_retry_plan`、`practice_plans.goal/mode` 均对齐当前 product-scope / B1 | PASS |
| enum/check constraints | `migrations/enum-sources.yaml` 与 SQL check 双向比对；B1 shared enum、B2 API-facing async enum、B3 10 job canonical 值、ADR-Q1 auth 状态均有来源登记；本轮新增 product-scope lint，防止旧 `MistakeStatus`、旧 practice mode、非 session-scoped report、错误 `feature_key` 落点和 vendor/model token 回流 | PASS after fix |
| outbox / async job schema | `async_jobs.job_type` check 含 B3 10 项，B2 API-facing subset 仍由 OpenAPI 7 项控制；`outbox_events` 含 retry/dead-letter 字段与 `(publish_status,next_attempt_at,created_at)` pending due 索引 | PASS |
| privacy deletion matrix | `backend/internal/migrations/privacy.go` 是可执行 dry-run truth source；本轮新增 `TestPrivacyMatrixCoversEveryBaselineTableExactly`，断言 31 张 public baseline 表逐一覆盖、无 duplicate、无 `mistake_entries`；`make privacy-delete-dry-run` 输出覆盖 hard delete / cascade / retain / tombstone | PASS after fix |
| migration lint / migrate-check | `python3 scripts/lint/migrations_lint.py --repo-root .` PASS；`APP_ENV=prod make migrate-down` 在 `DATABASE_URL` 校验前 fail-fast 并提示 `MIGRATE_DOWN_FORCE=1`。当前 shell `DATABASE_URL=unset`，DB-backed `make migrate-check` 不在 B4 本地步骤伪造通过，延后到 A2/local-dev-stack 与 global close 使用 dev DB 执行 | PASS offline; DB-backed deferred |
| 修复 | 修复 B4 context/plan specVersion v1.6→v1.8；新增 `migrations_lint.py` product-scope semantic guard 与 5 个 pytest；新增 Go privacy matrix 精确覆盖测试 | FIXED |
| 验证命令 | `validate_context.py --context docs/spec/db-migrations-baseline/plans/001-bootstrap/context.yaml --target backend` PASS；`python3 -m pytest scripts/lint/migrations_lint_test.py -q` 13 passed；`python3 scripts/lint/migrations_lint.py --repo-root .` PASS；`cd backend && go test ./internal/migrations ./cmd/migrate -count=1` PASS；`make privacy-delete-dry-run` PASS；`APP_ENV=prod make migrate-down` expected fail-fast | PASS |

### 4.6 `secrets-and-config/001-bootstrap`

| 检查面 | 本轮证据 | 结论 |
|--------|----------|------|
| plan/checklist 原文 | 重新读取 spec v1.9、plan/checklist/context；真实 target 为 `platform-config`，`backend` target 不存在；按 context discovery 直接审 `backend/internal/platform/{config,secrets,featureflag}`、`frontend/src/lib/runtime-config`、`config/`、hooks、Makefile | PASS |
| config loader / validator | `LoadCanonical` 绑定 24 项 env key 与 secret bindings；validator 覆盖 staging/prod runtime override、防 dev default DB/Redis/object storage、AI gateway、PostHog、email、auth secret 与 `async.queueWeights` 正数；cmd/api 与 cmd/worker 复用 canonical loader | PASS |
| env dictionary / `.env.example` | `.env.example` 24 key、spec §3.1.1 24 key、code-side bindings 24 key 三方一致；`auth.sessionCookieName=ei_session` 固定，无 env override；`async.queueWeights` 保持 config-only | PASS |
| feature flags / runtime config | `config/feature-flags.yaml` 当前 6 项 P0 flag；旧 `mistake_book_export_enabled` / `growth_dashboard_v1_enabled` / `mock_session_dual_track_enabled` 不在实现侧；本轮修复 runtime-config builder 的防御 allowlist，`ai_fallback_model_enabled` 即使上游误标 public 也不会出现在前端响应；PostHog provider 缺 `POSTHOG_PROJECT_API_KEY` fail-fast | PASS after fix |
| secret boundary / hook / gitleaks fallback | `getenv_boundary.go` allowlist 与当前 spec §4.1 对齐，包含 `cmd/{api,worker,migrate}`；pre-commit umbrella 调用 `pre-commit-secrets.sh` 且不打印 secret 字面量；`make lint-config` 本地 gitleaks 未安装时按 A4 策略提示并 exit 0 | PASS |
| current product + runtime truth source | `config/README.md`、runtime-config Go/TS types、cmd/api stub 与 OpenAPI B2 边界一致；前端 runtime-config fetcher 不读 `import.meta.env` / `VITE_*`；无 PostHog SDK import；旧 feature flag 只作为文档说明或负向测试存在 | PASS |
| 修复 | 修复 `runtime_config.go` public allowlist；新增/调整 runtime-config 与 PostHog provider tests；修复 plan/checklist/INDEX 中 `cmd/migrate` allowlist 旧口径并同步文档索引 | FIXED |
| 验证命令 | `validate_context.py --target platform-config` PASS；`make lint-config` PASS（gitleaks absent skip as designed）；`python3 -m pytest scripts/lint/env_dict_test.py -q` 5 passed；`cd backend && go test ./internal/platform/config ./internal/platform/secrets ./internal/platform/featureflag ./cmd/api ./cmd/worker -count=1` PASS；`cd frontend && pnpm test src/lib/runtime-config` PASS；`sync-doc-index.py --check` zero drift | PASS |

### 4.7 `ai-gateway-and-model-routing/001-aiclient-and-profile-bootstrap`

| 检查面 | 本轮证据 | 结论 |
|--------|----------|------|
| plan/checklist 原文 | 重新读取 spec v1.7、plan v1.4、checklist v1.3、context；忽略 completed/checkmark，把 Phase 1-5 逐项映射到 `backend/internal/ai/aiclient`、`providers/{stub,openai_compatible}`、`profile/loader.go`、`observability/`、`config/ai-profiles` 与 A4 config handoff | PASS |
| AIClient interface / meta | `AIClient` 只暴露 `Complete` / `Embed` / `Stream`；`AICallMeta` 字段顺序与 spec §4.1 一致；`CallMetadata.FeatureKey` 仅作为 per-call metadata 字段存在，未参与 `dispatch`、provider selection、model selection 或 feature flag routing；`metaBuilder` 只合流 profile/provider/call metadata | PASS |
| profile loader / config integration | `profile.Loader` 结构化读取 `name` / `task_type` / `default` / `fallback` / `timeout_ms` / `max_tokens` / `rate_limit` / `gateway_route` / `version`，错误含 file path + line；新增 focused test 锁住 `task_type=stt` 只作为 loader 预留值，client 调用返回 `ErrTaskTypeNotImplemented`；`New(cfg)` 在 non-test 缺 `AI_GATEWAY_*` 时 fail-fast | PASS after fix |
| stub + OpenAI-compatible + mockserver | stub factory 只允许 `APP_ENV=test` 或显式 override，deterministic output 无时间/随机数；`openai_compatible` 仅用标准库 `net/http` + `encoding/json`，只调用 `/v1/chat/completions` / `/v1/embeddings`，无 `/v1/audio/transcriptions` / `Transcribe`；fallback 只消费 endpoint/gateway header，不遍历 `profile.Fallback` 自行切模型；mockserver 覆盖 chat/embeddings/timeout/5xx/4xx/fallback header/missing choices | PASS |
| observability wrapper | 7 个 metric family、4 类 log、`ai_task_runs`、`audit_events` 均由 decorator 覆盖；privacy test 验证 prompt/response 明文不进入 metric label/log/DB/audit metadata；本轮发现 fallback metric helper 仍按最后一个 `-` 盲目推断 model family，已加红测并修为仅剥离明确 `YYYY-MM-DD` 日期后缀 | PASS after fix |
| 旧 feature_key / model naming 假设 | vendor SDK import grep 无命中；`claude` / `sonnet` / `gpt-` / `anthropic` 在 A3 runtime/config 无实现命中，仅 README/doc.go 禁令说明；`featureKey` 只在 `CallMetadata` 与 tests 中出现；STT 只作为 `TaskTypeSTT` 预留；旧模型命名假设由 `TestDecorator_FallbackCounterDerivesModelFamilyOnlyFromDateSuffix` 锁住 | PASS after fix |
| 修复 | 修复 `observability/decorator.go` fallback model-family 归一化；新增 `TestDecorator_FallbackCounterDerivesModelFamilyOnlyFromDateSuffix`、`TestSTTProfileReturnsTaskTypeNotImplemented`、`TestLoaderAcceptsSTTAsReservedTaskType`，防止旧模型命名与 P0 STT 假实现回流 | FIXED |
| 验证命令 | `validate_context.py --target backend` PASS；`go test ./internal/ai/aiclient/... -count=1` PASS；`go test -race ./internal/ai/aiclient/profile -run TestLoaderConcurrentReadAndReload -count=1` PASS；vendor SDK import grep 无命中；旧模型/provider token grep 仅命中文档禁令与 `FeatureKey` metadata 字段 | PASS |

### 4.8 `local-dev-stack/001-bootstrap`

| 检查面 | 本轮证据 | 结论 |
|--------|----------|------|
| plan/checklist 原文 | 重新读取 spec v1.5、plan v1.4、checklist v1.3、context；忽略 completed/checkmark，把 Phase 1-4 映射到 `deploy/dev-stack/docker-compose.yaml`、init scripts、`deploy/dev-stack/Makefile`、root `Makefile`、`dev-doctor.sh`、`.env.example`、README | PASS after fix |
| docker compose services / volumes / init | `docker compose config --quiet` PASS；默认服务只有 Postgres+pgvector / Redis / MinIO / minio-init，无 OTel/Grafana/Loki/Prometheus/AI gateway/PostHog 默认服务；3 个命名卷固定；`minio-init` 幂等创建 bucket；真实 `make dev-up` 后 doctor 3/3 OK | PASS |
| dev-doctor contract | `dev-doctor.sh` 177 行，POSIX sh + jq；JSON 输出 `services[]` + `summary{ok,degraded,down,total}`；PG probe 覆盖 `pg_isready` / `select 1` / `pgvector`，Redis 覆盖 set/get/del，MinIO 覆盖 `mc ls` bucket；停止 `redis-dev` 后 `make dev-doctor` 返回 `summary.down=1` 并非 0 退出 | PASS |
| Make targets / idempotency / reset boundary | root `dev-*` target 递归到 `deploy/dev-stack`; `make -n dev-up` 不执行容器操作；真实重复 `make dev-up` 输出 `already healthy`；写入 probe row 后 `dev-down -> dev-up` 数据仍在；`printf no | make dev-reset` abort 且卷保留；`DEV_RESET_FORCE=1 make dev-reset` 删除 3 个命名卷；双栈占用 5432 时 `make dev-up` 非 0 且 stderr 指出 Python 端口占用 | PASS |
| README 与当前本地验证入口 | `deploy/dev-stack/README.md` 记录 Docker/Compose/JQ/lsof/curl 前置、服务表、项目组件 label 合同、`make dev-*` 命令、AI provider 真实配置、不走 stub、与 Kind/scenario 双轨；`.env.example` 中 `AI_GATEWAY_API_KEY` 为空占位且 `deploy/dev-stack/.env` 被 `.gitignore` 命中 | PASS |
| 修复 | 修复 `docs/spec/local-dev-stack/plans/001-bootstrap/context.yaml` `specVersion.to` v1.4→v1.5，避免 context PASS 掩盖当前 spec Header 版本漂移 | FIXED |
| 验证命令 | `validate_context.py --target repo` PASS；`bash -n dev-doctor.sh` PASS；`docker compose config --quiet` PASS；`make dev-up` PASS；`make dev-doctor` PASS；`psql ... extname='vector'` 返回 `vector`；data persist / reset / port conflict / redis-down failure paths 均已实际执行；当前 dev stack 已重新恢复为 3/3 OK | PASS |

### 4.9 `ci-pipeline-baseline/001-local-quality-gates`

| 检查面 | 本轮证据 | 结论 |
|--------|----------|------|
| plan/checklist 原文 | 重新读取 spec v1.3、plan/checklist v1.5、context；忽略 completed/checkmark，把 Phase 1-5 映射到 root `Makefile`、`scripts/lint/check_md_links.py` / tests、sync-doc-index、B1/B2/B3/A4 owner gates 与 deferred CI 文档 | PASS |
| `docs-check` | `docs-check` 串行执行 sync-doc-index、全 `docs/` 相对链接、`docs/spec` fragment anchor；本轮实际先失败于 `BUG-0006` 指向未创建 work-journal 的链接，已修为 pending plain text 并删除过早 global-pass 表述；重跑 `make docs-check` PASS | PASS after fix |
| `codegen-check` | `codegen-check` 覆盖 B1 conventions drift、B3 events/jobs drift、B2 OpenAPI drift 与 OpenAPI inventory；本轮实际先失败于 B3 `codegen-events-check` 把手写 `frontend/src/lib/events/events.test.ts` 当 generated path，已收窄为 generator 实际输出文件并补 dry-run 回归测试；重跑 PASS | PASS after fix |
| `lint` / `test` / `build` | `make lint` 实际执行 B1 conventions、A4 config、F1 placeholder、backend `golangci-lint`、frontend lint placeholder并 exit 0；`make test` / `make build` 留到 Final Reconcile 全局 gate 执行，A5 本轮已核对 Makefile 聚合入口与当前 owner 边界 | PASS; test/build deferred to global |
| NOT-YET-LANDED / deferred CI 边界 | `find .github/workflows` 无文件；README / `docs/development.md` / A5 spec 将 branch protection / required checks / workflows 全部写为 deferred/out of scope；`lint-observability` 仍是唯一 F1 placeholder 且输出 `not implemented yet: F1 observability-stack` exit 0 | PASS |
| 修复 | 修复 `BUG-0006.md` 虚假 work-journal 链接与过早 Global pass；修复 root `Makefile` B3 generated drift path；新增 `scripts/lint/makefile_dry_run_test.py` 覆盖 `codegen-events-check` 只检查生成物；同步修订 B3 spec/plan/checklist/context/INDEX 至 v1.7 | FIXED |
| 验证命令 | `validate_context.py --target repo` PASS；`python3 -m unittest scripts/lint/check_md_links_test.py scripts/lint/makefile_dry_run_test.py` PASS；`make docs-check` PASS；`make codegen-check` PASS；`make lint` PASS；`sync-doc-index.py --check` zero drift；`.github/workflows` zero files | PASS |

### 4.10 `engineering-roadmap/001-decompose-subspecs`

| 检查面 | 本轮证据 | 结论 |
|--------|----------|------|
| plan/checklist 原文 | 重新读取 `engineering-roadmap/spec.md`、`001-decompose-subspecs/plan.md`、`checklist.md`、`context.yaml`、`plans/INDEX.md`；spec Header 为 v3.0 active，plan/checklist 为 v3.1 completed，context `specVersion.to=3.0` 正确指向 spec Header，plans/INDEX 正确投影 plan v3.1 | PASS |
| Phase 3 future rule 语义 | `plan.md` Phase 3 明确“只记录后续 P0 workstream 的创建规则，不创建 child spec / plan，也不把这些候选项转成当前 implementation target”；checklist 3.1-3.4 的 2026-05-05 verification comments 均标注为 future rule only；spec D-5 / §5.2 / §5.3 / §6.5 也把候选 workstream 和 future candidates 定义为 on-demand 创建 | PASS |
| 未创建 child spec / plan | `find docs/spec -maxdepth 2 -name spec.md` 只列出 14 个真实现存 spec；不存在 `mock-contract-suite`、`frontend-shell`、`backend-auth`、`backend-targetjob`、`backend-practice`、`e2e-scenarios-p0`、`analytics-funnel`、`release-gate-and-rollout` 等候选 child spec 目录 | PASS |
| INDEX / context 候选语义 | `docs/spec/INDEX.md` 只投影真实 spec，且开头明确 future workstream 不进入索引；`rg` 检查 `docs/spec/INDEX.md` 与全部 `plans/INDEX.md` 未命中 `_pending_`、`待 spawn`、候选 P0 child 或 future candidate；`context.yaml` discovery keyword 保留 `on-demand child spec` 作为检索语义，不把 child 写成 target | PASS |
| 修复 | 本轮 L1 未发现需要修改 roadmap 文档的 material finding；用户关于“不得浅审 / 不得用历史 PASS 或 diff 大小作证据”的反馈已写入本台账 §1.1，并作为后续与 final reconcile 的执行规章 | NO-OP |
| 验证命令 | `validate_context.py --context docs/spec/engineering-roadmap/plans/001-decompose-subspecs/context.yaml --target docs` PASS；`sync-doc-index.py --check` zero drift；INDEX 负向 `rg` 无输出；真实 spec 目录清点无候选 child | PASS |

## 5 Final Reconcile

| Gate | 本轮结果 | 备注 |
|------|----------|------|
| `make docs-check` | PASS | Header/INDEX zero drift；全 `docs/` 相对链接 OK；`docs/spec` heading fragments OK |
| `make codegen-check` | PASS | B1 conventions、B3 events/jobs、B2 OpenAPI generated output 与 inventory 均无 drift |
| DB-backed migration gate | PASS | `make dev-doctor` 3/3 OK；`DATABASE_URL='postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' make migrate-check` PASS |
| focused tests | PASS | Python unittest 89；pytest 59；后端 A3/A4/B4 focused Go packages PASS；前端 10 test files / 49 tests PASS |
| `make lint` | PASS | B1/A4/F1/backend/frontend lint 聚合通过；gitleaks 本机未安装，按 A4 second-layer skip 策略提示并 exit 0 |
| `make test` | PASS | 后端 Go packages PASS；前端 Vitest 10 files / 49 tests PASS |
| `make build` | PASS | 后端 cmd build 通过；前端 build 仍为 D1 placeholder：`TODO: build implemented by D1 frontend-shell` |
| retrospective / bug-report 判断 | PASS | BUG-0006 已记录并更新；新增 [Historical Spec Deep Reimplementation 交付复盘](./2026-05-05-historical-spec-deep-reimplementation-assessment.md) |
| work-journal / commit | PASS | 新增 [2026-05-05 工作日志](../work-journal/2026-05-05.md)，关联 commit subject 为 `fix(historical-spec): deep reconcile existing plans` |
