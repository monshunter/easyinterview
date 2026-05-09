# F3 Baseline Registry, Resolve and Lint Gates

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-09

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [prompt-rubric-registry spec](../../spec.md) v2.1 已锁定的 §3.1.1 10 个 baseline `feature_key`、§3.1 D-1~D-12 决策与 §6 C-1~C-11 验收，端到端落到一个独立可部署、可验证的实施切片中。完成后：

- `config/prompts/<feature_key>/v0.1.0*.{yaml,md}` 与 `config/rubrics/<feature_key>/v0.1.0*.yaml` 覆盖 10 个 feature_key 的 baseline truth source 全部就位（schema 完整，每个 feature_key 至少 2 个 language 坐标，含 `multi`，以验证 `multi` fallback；文件命名遵守 spec D-2）。
- `backend/internal/ai/registry/` Go 包提供 `RegistryClient.ResolveActive` / `GetPrompt` / `GetRubric` 与 LLM `Judge` 接口 stub（NotImplementedJudge 默认实现）；`Resolve` 单调用 P95 ≤ 5ms，启动加载 ≤ 1s。
- `scripts/lint/{prompt_lint,rubric_lint,prompt_hardcode_lint}.py` 三套 lint script + Makefile target（`lint-prompts` / `lint-rubrics` / `lint-prompts-hardcode`）+ 已有 `lint-ai-profile-coverage` 联动，关闭 spec C-2 / C-4 / C-11。
- `migrations/NNNNNN_seed_baseline_prompt_rubric_versions.up.sql` 把 10 个 feature_key 的 prompt/rubric baseline language 坐标写入 `prompt_versions` / `rubric_versions`；同时按实事求是口径修订 B1/B4/A3 上游契约与实现，让 `ai_task_runs` typed columns 显式承载 `feature_key` / `prompt_version` / `rubric_version` / `model_profile_name` / `feature_flag` / `data_source_version`，关闭 C-9。
- `backend/internal/targetjob/StaticPromptRegistry` 与 4 个 `defaultTargetImport*` 常量被 retire，`backend/cmd/api/` 注入新 `RegistryClient` + 本地 `RegistryAdapter`。
- 关闭 §2.1 业务调用规约 + 全部 11 个 AC 自检；`docs/spec/backend-practice` D-29 前置依赖解锁，AI 首题 / 追问 / hint handler 实现 plan 可立即启动。

本 plan 不切真实 Model Profile（推到 002-real-model-profile-and-evals）、不实现 LLM Judge 业务逻辑（推到 002）、不实现 PostHog 灰度分桶（推到 003-grayscale-and-quality-feedback）、不修改 A3 `config/ai-profiles.yaml` 中任何 profile 的 status（仅校验 entry 存在 + 状态合法 + `unsupported_reason` 完整）。本 plan 会显式修改上游 B1/B4/A3 的设计文档与实现代码，但仅限 prompt/rubric provenance 闭环所必需的 `feature_key` / `feature_flag` / `data_source_version` 承载，不借机重开模型路由或指标标签设计。

## 2 背景

`prompt-rubric-registry` spec v2.1（2026-05-09）由 `engineering-roadmap/001-decompose-subspecs` 的 contract lock 创建，并在 L1 plan review 中补齐 C-9 provenance 要求，已锁定 10 个 baseline feature_key 的坐标、文件落点、Resolve 契约、template_hash 计算公式、灰度规则、A3 profile coverage gate 与 LLM Judge 接口签名。

当前 `001-baseline` impl plan 已派生为 active plan；`config/prompts/`、`config/rubrics/` 目录不存在；`backend/internal/ai/registry/` Go 包不存在；3 套 lint script 与对应 Makefile target 缺失；`prompt_versions` / `rubric_versions` 表已就位但无 baseline seed。

L1 review 同时确认：B2 `GenerationProvenance` 已要求 `featureFlag` / `dataSourceVersion`，B1 AI vocabulary 已有 `feature_flag` / `data_source_version`，但 B4 `ai_task_runs` schema、A3 `AICallMeta` / `AITaskRunRow` 与 observability writer 尚未完整承载这些字段；且仅靠 `prompt_version = v0.1.0` 无法在多 `feature_key` 下唯一定位 prompt/rubric baseline。因此本 plan 实施时必须先修订上游 B1/B4/A3 设计与实现，再实施 F3 registry 与 targetjob retire，避免把真实 cross-layer 缺口误删成“plan 断言错误”。

`docs/spec/backend-practice/spec.md` v1.3 D-29 明确：「F3 `001-baseline` 必须独立派生并完成（Resolve / prompt / rubric / lint gates），backend-practice 才能进入依赖 AI 输出（首题 / 追问 / hint）的实施阶段」。当前 `backend/internal/targetjob/prompt_registry.go` 的 `StaticPromptRegistry` 仅硬编码 `target.import.parse` 一条 resolution，是 backend-targetjob/001 的过渡 bridge；spec §2.1 + §5 明确 F3 owner 是 `internal/ai/registry/`，必须在本 plan 完成包结构升级与 retire。

本 plan 目的是把 spec C-1~C-11 全部 cover 在一个 vertical behavior slice 内，让 backend-practice 等待时间最短，同时落地 deep reconcile 红线（[CLAUDE.md §2.1.2](../../../../../CLAUDE.md)）：禁止旧 `StaticPromptRegistry` / 4 个 `defaultTargetImport*` 常量残留、禁止业务包出现裸 prompt 字面量、禁止 `mistakes` / `growth` / `drill` / `mistake.extract` 等已退役模块名反向回流到 prompt / rubric 命名空间。

用户 2026-05-09 决策（写入本 plan §3 与各 phase）：

| 决策点 | 选择 | 落入方式 |
|---|---|---|
| A3 profile coverage（C-11 / D-11） | 仅校验 entry 存在 + 状态合法 | Phase 5 跑 `make lint-ai-profile-coverage`；本 plan 不动 `config/ai-profiles.yaml` 的 status；`disabled` / `unsupported` 必须携带 `unsupported_reason` |
| C-9 DB seed | Phase 4 纳入本 plan | 新增 seed migration + dockertest 集成测试 |
| Prompt body depth | schema 完整 + 文本可用文案 | 10 个 feature_key 的 baseline prompt 全部写真实可用文案（system message 骨架 + user template + 输出 schema 提示），002 切真实模型时再迭代 |
| LLM Judge stub | interface + 空实现 + `ErrJudgeNotImplemented` | `backend/internal/ai/registry/judge.go` 导出 `Judge` interface + `NotImplementedJudge`；002 plan 仅替换实现 |
| `ai_task_runs` provenance | 保留并补齐 typed columns，不删除 C-9 断言 | Phase 0 复核 B1/B4/A3 当前缺口；Phase 4 先修订上游 spec / lint / schema / writer，再写 seed 与 cross-layer tests |

## 3 质量门禁分类

- **Plan 类型**: `code-internal + contract + truth-source + migration + tooling`。本 plan 落地 F3-owned Go 包 + `config/` truth source + `scripts/lint/` 静态校验 + 一份 idempotent seed migration；同时在同一 owner slice 内修订 B1/B4/A3 prompt provenance 契约与实现。不引入用户可感知 UI、HTTP API 行为或端到端业务工作流，但通过 `ai_task_runs` cross-layer assertion 与 `targetjob.PromptResolution` 字段映射跨包对齐 backend-practice GenerationProvenance schema。
- **TDD 策略**: Code plan requires TDD。所有 checklist item 必须先红后绿：① lint script 先写 negative fixtures（hash drift / dimensions 缺 weight / 业务包注入 `prompt :=`）；② Go 包按 `loader → resolver → cache → judge → registry` 顺序补 focused tests，再写实现；③ targetjob retire 走 grep red-line 单测先红、改 wiring 后绿；④ B1/B4/A3 provenance remediation 先写 schema / writer / lint drift tests，再改 shared conventions、migration spec、A3 writer 与 schema migration；⑤ DB seed migration 先写 dockertest assertion，再写 `up.sql` / `down.sql`。每个 phase 的退出 gate 都是 `go test` / `make lint` / `make migrate-check` 可执行命令。
- **BDD 策略**: BDD 不适用。本 plan 是 F3-owned 内部 Go 包（`backend/internal/ai/registry/`）+ `config/` truth source + `scripts/lint/` 静态校验，不新增用户可见 UI、HTTP API 行为或端到端业务工作流；TargetJobs fixture 仅做既有 provenance 示例对齐并由 `make validate-fixtures` 验证。Resolve 契约是 Go interface 与 yaml schema，不可通过浏览器或外部 API 触发。后续 P0 用户行为（first_question / followup / hint / report 等）由 `backend-practice` / `backend-report` / `backend-resume` 各自 plan 维护 BDD/E2E gate。
- **替代验证 gate**: ① `make lint-prompts` / `make lint-rubrics` / `make lint-prompts-hardcode` / `make lint-ai-profile-coverage` 静态门禁；② `go test ./backend/internal/ai/registry/... -race`（含 `BenchmarkResolve` P95 + `TestStartupBudget` 1s budget）；③ `go test ./backend/internal/targetjob/... -race` 跨包契约对齐 + active-scope negative grep；④ `go test ./backend/internal/ai/aiclient/... -race` 覆盖 `CallMetadata` / `AICallMeta` / `AITaskRunRow` provenance 映射；⑤ `go test -tags=integration ./backend/internal/ai/registry/...` dockertest seed migration cross-layer；⑥ `make migrate-check` 验证 B4 schema / lint / down-up 闭环；⑦ `make validate-fixtures` 验证 TargetJobs fixtures 与 B2 schema / provenance shape；⑧ grep red-line：旧 `StaticPromptRegistry` / 4 个 `defaultTargetImport*` 常量 / 业务包 hardcode prompt / 已退役 `mistakes` / `growth` / `drill` / `mistake.extract` 模块名；⑨ `python3 .agent-skills/implement/shared/scripts/validate_context.py` + `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`。

### 3.1 Cross-layer Operation Matrix

| Operation / Contract | `operationId` / 入口 | Fixture / Truth source | Frontend consumer | Backend handler / writer | Persistence | AI dependency | Scenario coverage |
|---|---|---|---|---|---|---|---|
| Target job import parse request | `importTargetJob` | `config/prompts/target.import.parse/v0.1.0*.{yaml,md}` + `config/rubrics/target.import.parse/v0.1.0*.yaml` + `openapi/fixtures/TargetJobs/importTargetJob.json` | `frontend/src/app/screens/home/HomeScreen.tsx` 调用 generated `client.importTargetJob`；本 plan 不改 UI DOM | `backend/internal/targetjob.Handler.ImportTargetJob` / `Service.ImportTargetJob` / async `ParseExecutor` | `target_jobs` / `target_job_sources` / `async_jobs` / `outbox_events` / `ai_task_runs` | F3 `RegistryClient.ResolveActive("target.import.parse", language)` + A3 model profile routing | BDD 不适用；替代 gate: targetjob HTTP tests + fixture validation |
| Target job parse polling detail | `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` + `TargetJobSummary.provenance` / `TargetJobFitSummary.provenance` in `openapi/openapi.yaml` | `frontend/src/app/screens/parse/ParseScreen.tsx` polling generated `client.getTargetJob` until `analysisStatus=ready|failed` | `backend/internal/targetjob.Handler.GetTargetJob` / `Service.GetTargetJob` / `SQLStore.GetTargetJobByUser` | `target_jobs` / `target_job_requirements` / `target_job_sources` | Reads parse result produced by F3-backed `ParseExecutor`; no new model call on GET | BDD 不适用；替代 gate: `make validate-fixtures` + targetjob HTTP scenario tests |
| Prompt/rubric registry internal resolve | Go interface only | `config/prompts/**/v0.1.0*` + `config/rubrics/**/v0.1.0*` | None | `backend/internal/ai/registry.Client.ResolveActive` | `prompt_versions` / `rubric_versions` seed rows | No direct model call; `Judge` stub only | Go unit + integration tests |
| AI task run observability row | Internal writer | B1 AI vocabulary + B4 `ai_task_runs` schema + A3 `CallMetadata` | None | `backend/internal/ai/aiclient/observability.Decorator` / fake writer tests | `ai_task_runs.feature_key` / `prompt_version` / `rubric_version` / `feature_flag` / `data_source_version` typed columns | A3 `AICallMeta` carries registry provenance; F1 metrics must not add these as labels | aiclient unit tests + targetjob fake writer cross-layer tests |

## 4 实施步骤

### Phase 0: 上游契约复核（无代码 / 契约写入；只记录 handoff snapshot）

#### 0.1 A3 profile catalog 当前状态摘要

读取 `config/ai-profiles.yaml`，按 spec §3.1.1 列出的 10 个默认 `model_profile_name` 抽出每条 entry 的 `status` / `capability` / `provider_ref` / `unsupported_reason`，作为 Phase 5 `make lint-ai-profile-coverage` 的预期输入；记入 §8.1.1 Phase 0 handoff snapshot。本 plan **不**修改任何 profile status；状态 active 化推到 `ai-provider-and-model-routing/003` 后续修订或 F3 002。

#### 0.2 DB schema 边界复核

读取 `migrations/000001_create_baseline.up.sql` L341-396 的 `prompt_versions` / `rubric_versions` / `ai_task_runs` schema，确认本 plan Phase 4 seed migration 写入的字段集（`feature_key, version, language, template_hash, template_body, is_active, created_at` 与 `feature_key, version, language, schema_json, is_active, created_at`）与 spec §4.1 schema 一致；客观记录当前 `ai_task_runs` 缺少 `feature_key` / `feature_flag` / `data_source_version` typed columns 的事实，作为 Phase 4 修订 B4 schema 与 A3 writer 的输入，不把缺口降级为删除 C-9 断言。

#### 0.3 targetjob bridge 退役边界复核

读取 `backend/internal/targetjob/prompt_registry.go` 与 `backend/internal/targetjob/parse_executor.go` L20-44，确认：① 4 个 `defaultTargetImport*` 常量与 `StaticPromptRegistry` 的 retire scope；② `targetjob.PromptResolution` 7 字段（`PromptVersion / RubricVersion / ModelProfileName / DataSourceVersion / FeatureFlag / SystemMessage / UserMessageTemplate`）；③ `targetjob.PromptRegistryClient` interface 签名 = `Resolve(ctx, featureKey, language) (PromptResolution, error)`；新 `RegistryAdapter` 必须保持该签名兼容，并由调用参数与 A3 metadata 显式承载 `featureKey`。

#### 0.4 backend-practice 前置契约复核

读取 `docs/spec/backend-practice/spec.md` v1.3 D-29 与 GenerationProvenance schema 字段集；读取 `openapi/openapi.yaml` 中 `GenerationProvenance` 6 字段（`promptVersion / rubricVersion / modelId / language / featureFlag / dataSourceVersion`），作为 Phase 3 cross-layer assertion 的对齐目标。

#### 0.5 B1/B4/A3 provenance contract 复核

读取 `shared/conventions.yaml`、`backend/internal/shared/ai/vocabulary.go`、`docs/spec/shared-conventions-codified/spec.md`、`docs/spec/db-migrations-baseline/spec.md`、`backend/internal/ai/aiclient/{payload.go,meta.go,writers.go}` 与 `backend/internal/ai/aiclient/observability/decorator.go`，记录：① B1 已有 `feature_flag` / `data_source_version` 但缺 `feature_key` AI vocabulary；② B4 C-12 只覆盖模型路由 typed columns，未覆盖 prompt/rubric provenance typed columns；③ A3 `CallMetadata` 有 `FeatureKey` / `DataSourceVersion` 但缺 `FeatureFlag`，`AICallMeta` / `AITaskRunRow` / writer 缺三字段。Phase 4 必须先把这些 upstream contract 修成当前事实源，再继续 seed 与 cross-layer 闭环。

### Phase 1: Truth source 文件 + 3 套 lint script

#### 1.1 `config/prompts/` 与 `config/rubrics/` 命名空间脚手架

新增 `config/prompts/README.md` 与 `config/rubrics/README.md`，描述：① 字段顺序固定为 `feature_key / version / language / template_hash / status / created_at`；② canonical hash 算法 = `sha256(template_body + meta_for_hash_canonical_json)`，其中 `meta_for_hash` 是从 prompt YAML meta 删除 `template_hash` 后的 map；lint 与 Go loader 计算前必须移除该字段，不得把空字符串、旧 hash 或新 hash 写入 hash 输入；meta canonical = key 字典序 + UTF-8 + `\n` 行尾，由 `ruamel.yaml` round-trip 与 Go `encoding/json` Marshal 共享；③ rubric `dimensions[]{name, weight, score_levels[{label, threshold, description}]}` 与 spec §4.1 维度命名 allowlist；④ multi-language 命名约定（`<feature_key>/v0.1.0.yaml` 用 `language: multi`，多语言变体如 `v0.1.0.en.yaml` / `v0.1.0.zh.yaml`）。

#### 1.2 10 个 feature_key 的 v0.1.0 baseline prompt 文件

按 spec §3.1.1 顺序生成 10 个目录及其 v0.1.0 yaml + md：`target.import.parse` / `practice.session.first_question` / `practice.session.follow_up` / `practice.turn.lightweight_observe` / `report.generate` / `report.question_assessment` / `resume.parse` / `resume.tailor.gap_review` / `resume.tailor.bullet_suggestions` / `debrief.generate`。每个 feature_key 至少提供 2 个 language 变体（`multi` + `en` 或 `zh`）以验证 D-6 fallback。template body 写真实可用文案（system message 骨架 + user template + 输出 schema 提示），不写「TBD」字样；变量占位用 `{{...}}` 与 spec §2.1 + targetjob 现有 `UserMessageTemplate` 风格一致（参考 `backend/internal/targetjob/prompt_registry.go` 的 `{{jd_text}}`）。

#### 1.3 10 个 feature_key 的 v0.1.0 baseline rubric 文件

按 spec §3.1.1 + §4.1 schema 生成 10 个 feature_key 的 rubric yaml；每个 feature_key 至少提供 2 个 language 坐标（含 `multi`，文件命名遵守 spec D-2），并与 Phase 1.2 的 prompt language 集合一致；维度命名遵守 §4.1 allowlist（追问相关率 / 报告空泛率 / 异常高分率 / 语言混乱率等 F1/F3 推荐质量指标 + 业务域专有维度）；每个 rubric 至少包含 3 个 dimension，weight 总和 = 1.0（容差 ±0.001），每个 dimension 的 score_levels 至少 3 个 threshold 段。

#### 1.4 `scripts/lint/prompt_lint.py` + 单测

落地 lint：① 字段顺序固定 6 字段；② `status ∈ {draft, active, deprecated}`；③ `version` SemVer 正则；④ `template_hash` 与运行时计算结果一致（drift 检测）；⑤ `language` 枚举（`multi` 或 ISO-639 lower-case）；⑥ canonical algorithm 与 Go runtime 共享。新增 negative fixtures 至少 2 个（hash drift 修改 template body 不改 hash / 字段顺序错乱）；测试文件 `scripts/lint/prompt_lint_test.py` 必须先红后绿。

#### 1.5 `scripts/lint/rubric_lint.py` + 单测

落地 lint：① `dimensions` 非空；② `sum(weight) ≈ 1.0`（容差 ±0.001）；③ `score_levels` 至少 1 个 threshold；④ `language` 枚举；⑤ 维度名 allowlist。Negative fixtures 至少 2 个（dimensions 缺 weight / 维度名违反 allowlist）。

#### 1.6 `scripts/lint/prompt_hardcode_lint.py` + 单测

落地 AST-based lint：用 `go/parser`（通过 subprocess 调用 `go run` helper 或 Python 端口的 `tree-sitter-go`，按仓库已有约定选择）扫 `backend/internal/{practice,report,resume,debrief,targetjob}/**/*.go` 中 `*ast.AssignStmt` RHS 为 raw / 多行 string 且变量名匹配 `~prompt$|^Prompt|systemMessage` 的赋值；维护 allowlist（fixture / test / doc 路径）。Negative fixture 至少 1 个：在 `backend/internal/practice/` 下注入临时 `prompt := "..."` → exit 1。

#### 1.7 Makefile lint targets

在 `Makefile` 追加 `lint-prompts` / `lint-rubrics` / `lint-prompts-hardcode` 三个 target，调用对应 lint script；并把它们加入聚合 `lint` target（与已有 `lint-ai-profile-coverage` 并列）。`migrations_lint` 当前由 `make migrate-check` 触发，本 plan 收口必须同时运行 `make lint` 与 `make migrate-check`，不得声称 `make lint` 已覆盖 migration lint。

#### 1.8 Regression-negative scope grep

`grep -rE "mistakes|growth|drill|mistake.extract" config/prompts/ config/rubrics/` 必须返回 0 行（已退役模块名不能借机重入）；同时确认 spec §3.1.1 中删除的 C11 资料检索类占位不出现在新 baseline。

### Phase 2: `backend/internal/ai/registry/` Go 包

#### 2.1 包脚手架与 doc 红线

新增 `backend/internal/ai/registry/doc.go`：包注释 + 红线（不依赖 `backend/internal/targetjob`、不调 `aiclient`、不写业务 metric / log；只写自身加载状态 log）；`grep -rE "AIClient|aiclient\.|metric\.Counter" backend/internal/ai/registry/` 必须仅命中 doc 注释。

#### 2.2 `types.go`：核心类型 + Judge interface

定义：`PromptResolution`（包含 `FeatureKey` + 兼容 `targetjob.PromptResolution` 的 7 字段 + D-12 预留 `Tools []ToolDescriptor` / `OutputSchema *json.RawMessage` / `StreamWire *string`，本 phase 不消费但字段可见）；`PromptMeta` / `RubricSchema` / `RubricDimension` / `ScoreLevel`；`Judge` interface 签名 = `Judge(ctx context.Context, featureKey string, promptVersion string, output []byte, rubricVersion string) (Score, Reasoning, error)`，与 spec D-9 完全一致；错误类型 `ErrPromptUnsupported` / `ErrLanguageUnsupported` / `ErrJudgeNotImplemented`。

#### 2.3 `loader.go`：启动扫与 hash 校验

启动时扫 `config/prompts/` + `config/rubrics/`，逐项解析 yaml + md；逐项校验 `template_hash = sha256(template_body + meta_for_hash_canonical_json)` 与 yaml 内 `template_hash` 字段一致，其中 `meta_for_hash` 必须排除 `template_hash` 字段（drift → log warn + return error，不污染既有快照）；canonical 算法与 `scripts/lint/prompt_lint.py` 共享描述（`config/prompts/README.md`）。Focused tests：`loader_test.go` 覆盖 happy path / hash drift / missing meta / yaml 解析失败。

#### 2.4 `resolver.go`：ResolveActive + GetPrompt + GetRubric

`ResolveActive(ctx, featureKey, language) → PromptResolution`，language fallback 优先级精确 → `multi`；fallback 命中写 warn 级 log + bump 计数（C-6）；unknown feature_key → `ErrPromptUnsupported`；unknown language 在尝试 fallback 后仍无命中 → `ErrLanguageUnsupported`。`GetPrompt(featureKey, version, language)` 与 `GetRubric(featureKey, version, language)` 用于 backfill / debug；不接受空字符串。Focused tests：`resolver_test.go` 覆盖精确 language / fallback to multi / unknown feature_key / unknown language warn / 空字符串拒绝。

#### 2.5 `cache.go`：atomic snapshot + 30s TTL + Reload 钩子

用 `atomic.Value` 整体替换 snapshot（map[featureKey]map[language]*entry）；reload 期间旧 snapshot 留存到最后一次读取后 GC；提供 `Reload(ctx)` 测试钩子；TTL = 30s（spec §4.3）。Focused tests：`cache_test.go` 覆盖 TTL expiry、Reload idempotent（连续 5 次后 snapshot 大小恒定）、并发安全（100 goroutine 同时 Resolve + Reload 交错，`-race` 通过）。

#### 2.6 `judge.go`：NotImplementedJudge 默认实现

`NotImplementedJudge` struct 实现 `Judge` interface，所有方法调用始终返回 `ErrJudgeNotImplemented`；导出供业务包 wire default。Focused tests：`judge_test.go` 覆盖签名 freeze（用反射断言 method `Judge` 入参顺序与 spec D-9 一致）、`NotImplementedJudge` 默认返回。

#### 2.7 `registry.go`：NewRegistryClient 构造与全量加载

`NewRegistryClient(opts RegistryOptions) (*Client, error)`：opts 含 `PromptsDir` / `RubricsDir` / `Now func() time.Time` / `Logger`；启动时调用 `loader.go` 全量加载 + 计算 hash + 存入 cache。Focused tests：`registry_test.go` 覆盖默认加载全部 baseline（10 个 feature_key × ≥2 language）+ 不允许悬空 feature_key（即 prompt 有 rubric 无或反之 → 启动 fail）。

#### 2.8 `perf_test.go`：P95 + 启动 budget 硬断言

`BenchmarkResolve` 用 `b.ReportMetric` 输出 P95 latency；测试 helper 反读 metric 后断言 P95 ≤ 5ms（spec C-3）；`TestStartupBudget` 直接 `time.Since(start) < 1*time.Second`（spec §4.3，10 feature × ≥2 language）。

#### 2.9 边界 grep red-line

`grep -rE "AIClient|aiclient\.|metric\.Counter|secret\.|Secret\.|os\.Getenv" backend/internal/ai/registry/` 仅命中 doc 注释或 import 占位（registry 包不持 secret、不调 AI、不写 metric、不读 env）；如有违规，立即修订。

### Phase 3: targetjob StaticPromptRegistry retire + cross-layer 对齐

#### 3.1 `RegistryAdapter` 设计

新增 `backend/internal/targetjob/prompt_registry.go`（替换原文件）：删除 `StaticPromptRegistry` struct、`NewStaticPromptRegistry()`、4 个 `defaultTargetImport*` 常量；新增 `RegistryAdapter` struct，构造函数 `NewRegistryAdapter(client *registry.Client) *RegistryAdapter`，实现 `targetjob.PromptRegistryClient.Resolve` 并把 `registry.PromptResolution` 转换为 `targetjob.PromptResolution`（包括 `SystemMessage` / `UserMessageTemplate` 字段映射）。Adapter 必须 unit test 显式列出每个字段映射（`PromptVersion / RubricVersion / ModelProfileName / DataSourceVersion / FeatureFlag / SystemMessage / UserMessageTemplate`），同时断言入参 `featureKey` 与 registry 返回的 `FeatureKey` 一致；新增字段必须同步 adapter 或在 Phase 4 A3 metadata 中有明确承接。

#### 3.2 `backend/cmd/api/` DI wire 注入

把 `NewStaticPromptRegistry()` 替换为 `registry.NewRegistryClient(opts)` + `targetjob.NewRegistryAdapter(...)`；ParseExecutor 与未来 C5/C6/C7 handler 共享同一 `*registry.Client` 实例。

#### 3.3 `active_scope_negative_test.go`：grep 红线 Go 单测

新增 `backend/internal/targetjob/active_scope_negative_test.go`：用 `go/types` 包反射断言当前 package 不再 export `StaticPromptRegistry` / `NewStaticPromptRegistry` / 4 个 `defaultTargetImport*` 标识符；断言 `mistakes` / `growth` / `drill` / `mistake.extract` 等已退役模块名不出现在 targetjob test fixture / golden output。

#### 3.4 `parse_executor_test.go` cross-layer 对齐补强

在已有 fake AI writer 路径上追加 assertion：① ParseExecutor 收到的 `resolution.PromptVersion` / `RubricVersion` / `ModelProfileName` / `DataSourceVersion` / `FeatureFlag` / `SystemMessage` / `UserMessageTemplate` 与 `config/prompts/target.import.parse/v0.1.0*.yaml` meta 一致；② 输出的 `provenance` JSON 含 `promptVersion / rubricVersion / modelId / language / featureFlag / dataSourceVersion` 6 字段（cross-layer 对齐 backend-practice GenerationProvenance schema，`openapi/openapi.yaml` 中 `GenerationProvenance` 节）。`ai_task_runs` typed row 断言在 Phase 4 完成 B1/B4/A3 remediation 后追加，避免在 schema 未修复前写假绿测试。

#### 3.4.1 TargetJobs fixtures provenance 对齐

更新 `openapi/fixtures/TargetJobs/getTargetJob.json` 中 `TargetJobSummary.provenance` / `TargetJobFitSummary.provenance` 的 prompt/rubric/model profile/data source 示例，使其与 F3 `target.import.parse` baseline 坐标、A3 `target.import.default` profile 与 B2 `GenerationProvenance` 6 字段一致；`openapi/fixtures/TargetJobs/importTargetJob.json` 保持 `202 + Job` request/response shape，但 fixture path 与 operation matrix 必须大小写一致。运行 `make validate-fixtures`，并保留 `frontend/src/app/screens/parse/ParseScreen.tsx` polling consumer 不需要 UI 变更的证据。

#### 3.5 Phase 3 退出 grep red-line

`grep -rE "StaticPromptRegistry|defaultTargetImport(Prompt|Rubric|ModelProfile|DataSource)" backend/` 必须返回 0；`go vet ./backend/...` 无未使用 import / 死代码警告；`go test ./backend/internal/targetjob/... -race` 全绿。

### Phase 4: B1/B4/A3 provenance remediation + DB seed migration

#### 4.1 B1 AI vocabulary contract 修订

先修订 `docs/spec/shared-conventions-codified/spec.md` 与 `shared/conventions.yaml` / `backend/internal/shared/ai/vocabulary.go`：在既有 `feature_flag` / `data_source_version` 旁边增加 `feature_key`，并说明三者是 prompt/rubric provenance 字段，不进入 F1 metric label cardinality。新增或更新 B1 generator / lint 测试：缺少 `feature_key` 时失败，三字段命名必须保持 `snake_case` 与 shared conventions 一致。

#### 4.2 B4 `ai_task_runs` schema contract 修订

修订 `docs/spec/db-migrations-baseline/spec.md` C-12 / D-15 / §2.1，明确 `ai_task_runs` typed columns 除模型路由字段外还必须承载 `feature_key text not null`、`feature_flag text not null default 'none'`、`data_source_version text not null default 'not_applicable'`。按 B4 migration workflow 创建 `NNNNNN_add_ai_task_runs_prompt_provenance.{up,down}.sql`：up 只做 additive columns / backfill / not-null 收紧，down 至少恢复表结构骨架；如果当前 dev baseline 允许直接修订 `000001_create_baseline.up.sql`，仍必须在 §8 记录原因与 `make migrate-check` 证据。同步更新 `scripts/lint/migrations_lint.py`，允许并要求 `ai_task_runs.feature_key`，不得继续只允许 `prompt_versions` / `rubric_versions` 持有 `feature_key`。

#### 4.3 A3 aiclient metadata / writer 修订

修订 `backend/internal/ai/aiclient`：`CallMetadata` 增加 `FeatureFlag`；`AICallMeta` / `AITaskRunRow` / observability decorator 显式携带 `FeatureKey` / `FeatureFlag` / `DataSourceVersion`；writer tests 断言三字段非空且从 registry/targetjob resolution 传递到 row。新增 negative test：缺少 `FeatureKey`、空 `FeatureFlag` 或空 `DataSourceVersion` 时 writer 拒绝或填入已登记默认值（`none` / `not_applicable`），不得静默落空字符串。

#### 4.4 seed migration `up.sql`

新增 `migrations/NNNNNN_seed_baseline_prompt_rubric_versions.up.sql`：写入 10 个 feature_key × N language 的 baseline 行（`is_active = true`，`template_hash` 与 yaml 一致，`template_body` 来自同名 prompt md，如 `<feature_key>/v0.1.0.md` 或 `<feature_key>/v0.1.0.en.md`，`schema_json` 来自对应 rubric yaml canonical 化结果）；用 `INSERT ... ON CONFLICT (feature_key, version, language) DO NOTHING` 保证 idempotent；`ON CONFLICT` 不 flip 已有行的 `is_active`。

#### 4.5 seed migration `down.sql`

`DELETE FROM prompt_versions WHERE version = 'v0.1.0' AND feature_key IN (...)`；rubric 同；范围限定避免误删后续版本。

#### 4.6 `prompt_lint.py` cross-file 增强

增强 lint：扫 `migrations/*seed_baseline_prompt_rubric_versions*.up.sql` 中每行的 `template_hash`，与对应 `config/prompts/<feature_key>/v0.1.0*.yaml` 的 `template_hash` 比对一致性；不一致 → exit 1。

#### 4.7 `db_integration_test.go` (build tag: integration)

新增 `backend/internal/ai/registry/db_integration_test.go` 用 `//go:build integration`：dockertest / pgtestdb（沿用 `backend/internal/targetjob/store_test.go` 已有 PG fixture 约定）拉本地 PG，跑当前 migration chain + 本 plan seed migration，断言：① prompt_versions / rubric_versions 行数覆盖 10 个 feature_key × 实际 language 坐标（至少 20 行）且 prompt/rubric language 集合一致；② `is_active = true`；③ `template_hash` 与 yaml 一致；④ `ai_task_runs` 存在 `feature_key` / `feature_flag` / `data_source_version` typed columns；⑤ 同 feature_key 第二行 `is_active = true` 被 unique partial index 拒绝（如 schema 已支持），否则在 RegistryClient 层加运行时校验单测覆盖 D-7。

#### 4.8 ai_task_runs cross-layer 验证

在 `backend/internal/targetjob/parse_executor_test.go` 已有 fake writer 路径（不需要 dockertest）追加 assertion：`ai_task_runs.feature_key / prompt_version / rubric_version / model_profile_name / data_source_version / feature_flag` 6 字段非空且与 RegistryAdapter 返回的 PromptResolution 及调用 `featureKey` 一致。F1 observability tests 必须确认这些 provenance 字段不被新增为高基数 metric labels。

### Phase 5: 收口 + A3 coverage gate + sync-doc-index

#### 5.1 `lint-ai-profile-coverage` 联动

确认 `scripts/lint/ai_profile_coverage.py` 已经把 spec §3.1.1 列出的 10 个默认 `model_profile_name` 列入校验范围；本 plan 不修改 `config/ai-profiles.yaml` profile status；只验证：① 10 个 profile name 全部存在于 catalog；② `disabled` / `unsupported` 必须携带合法 capability 或 provider_ref + `unsupported_reason`。Phase 0 摘要的 profile catalog 当前状态写入 §8 handoff，作为 002 / `ai-provider-and-model-routing/004` 后续 catalog 修订的输入信号。

#### 5.2 顶层 `make lint` + 全量 verification 命令串联

确认顶层 `make lint` 包含本 plan 新增的 `lint-prompts` / `lint-rubrics` / `lint-prompts-hardcode` + 已有 `lint-ai-profile-coverage`；另跑 `make migrate-check` 覆盖 `migrations_lint` 与 migration down/up 闭环。§8 handoff 列出完整 verification one-liner（11 个 AC × 命令证据）。

#### 5.3 Plans INDEX 与 history 同步

确认 `docs/spec/prompt-rubric-registry/plans/INDEX.md` 已写入 `001-baseline` active 行（active → completed 由本 plan 收尾时切换）；更新 `docs/spec/prompt-rubric-registry/history.md` v2.1 entry，记录 `001-baseline` plan 派生、用户决策与 B1/B4/A3 provenance remediation 口径。由于本 L1 修复已修改 spec 文本，`spec.md` Header 必须保持 v2.1，并同步 `docs/spec/INDEX.md`。

#### 5.4 sync-doc-index 校验

`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` 必须通过；如失败按提示修订 Header / INDEX。

#### 5.5 retrospective 候选清单

§8 handoff 记录值得复盘的点：① `StaticPromptRegistry` retire 节奏与 backend-targetjob/001 的 phase boundary 对齐；② `template_hash` Go ↔ Python canonical 算法跨工具对齐策略；③ B1/B4/A3 provenance remediation 由本 plan 接管后，对后续 F1 / backend-practice plan 的影响；不直接生成 retrospective 报告，由用户在本 plan 收尾时通过 `/retrospective` 决定。

## 5 验收标准

- spec §6 C-1~C-11 全部通过本 plan §8 handoff 中列出的命令证据验证。
- §3 替代验证 gate 9 类全部通过。
- `docs/spec/prompt-rubric-registry/plans/INDEX.md` 写入 001-baseline 行；`history.md` 维护 v2.1 entry；`docs/spec/INDEX.md` 与 spec Header 一致。
- `backend-practice` D-29 前置依赖解除：grep 验证 `backend/internal/targetjob/StaticPromptRegistry` 与 4 个 `defaultTargetImport*` 常量已 retire；`backend/internal/ai/registry` 包对外提供 `Client.ResolveActive` + `Judge` interface，AI 首题 / 追问 / hint handler plan 可立即启动。
- BDD 不适用声明已固化在本 plan §3 + checklist BDD-Gate 不适用区段，并附 9 类替代 gate 命令。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| `template_hash` 跨 YAML key 顺序 / 空白漂移 | `prompt_lint.py` 用 `ruamel.yaml` round-trip + canonicalize（删除 `template_hash` 后 key 排序、UTF-8、`\n` 统一）→ 拼 template body 取 sha256；Go 端 `loader.go` 用相同算法（共享一份描述在 `config/prompts/README.md`）；Phase 1 加 Go ↔ Python 跨工具对齐单测。 |
| YAML 字段顺序漂移导致 lint 误报 | `prompt_lint.py` 强制字段顺序 = 固定 6 字段；违反 → exit 1 + 打印期望顺序，不依赖 yaml round-trip 自动排序。 |
| `prompt_hardcode_lint.py` 在 doc / fixture 含 prompt 字样误报 | 用 AST 而非纯 grep；仅扫 `*ast.AssignStmt` 中 prompt 后缀、Prompt 前缀或 `systemMessage` 变量，且 RHS 为 raw / 多行 string；维护 allowlist（fixture / test / doc 路径）。 |
| A3 profile catalog 状态漂移 | 本 plan 只校验「entry 存在 + 状态合法 + reason 完整」，不动 status；profile active 化推到 `ai-provider-and-model-routing/004` 或 F3 002；如 Phase 0 发现 profile 状态与 spec 默认预期不一致，只记录事实并交由后续 owner 决策。 |
| B1/B4/A3 upstream remediation 影响面扩大 | Phase 4 只允许改 prompt/rubric provenance 三字段的设计与实现：B1 vocabulary、B4 `ai_task_runs` typed columns / migration lint、A3 metadata / writer；禁止顺手调整真实模型路由、F1 label schema 或 API response shape。 |
| cache 热更新 + 30s TTL 竞态 | `atomic.Value` 整体替换 snapshot；reload 期间旧 snapshot 留存到最后一次读取后 GC；并发测试 100 goroutine + Reload 交错，`-race` 通过。 |
| golden test fixture drift | 单测不硬编码全部字段，用 `filepath.Walk` + 反射断言 keys 存在 + 类型；version 字符串与 fixture 路径动态读取，避免每次新增 baseline 都改测试。 |
| DB seed migration 重跑 idempotent 风险 | `INSERT ... ON CONFLICT (feature_key, version, language) DO NOTHING`；不在 seed 中 flip 已有行的 `is_active`；新 baseline 走新 migration（→ 002）。 |
| 跨包循环依赖（targetjob → registry） | registry 包不 import targetjob；targetjob 保留本地 interface + adapter，构造时把 `*registry.Client` 当 interface 注入；adapter 单独单测。 |
| `PromptResolution` 7 字段 cross-package 漂移 | Phase 3 adapter 集中映射；adapter unit test 显式列出每个字段映射；新增字段必须同步 adapter，违反 → 编译期失败。 |

## 7 关联文档导航

> - [Prompt Rubric Registry Spec](../../spec.md)
> - [AI Provider and Model Routing Spec](../../../ai-provider-and-model-routing/spec.md)
> - [DB Migrations Baseline Spec](../../../db-migrations-baseline/spec.md)
> - [Secrets and Config Spec](../../../secrets-and-config/spec.md)
> - [Backend Practice Spec](../../../backend-practice/spec.md)
> - [Backend Targetjob Spec](../../../backend-targetjob/spec.md)
> - [Engineering Roadmap Spec](../../../engineering-roadmap/spec.md)
> - [ADR-Q6 AI Provider and Model Routing](../../../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md)
> - [Observability Stack Spec](../../../observability-stack/spec.md)
> - [Shared Conventions Codified Spec](../../../shared-conventions-codified/spec.md)

## 8 Handoff / Evidence Log

### 8.1 Phase evidence log

| Phase | Evidence slot | Required command / artifact |
|---|---|---|
| Phase 0 | Upstream contract snapshot | Profile catalog snapshot, schema gap table, targetjob bridge inventory, OpenAPI / fixture provenance table, B1/B4/A3 provenance gap table |
| Phase 1 | Truth source + lint | `make lint-prompts && make lint-rubrics && make lint-prompts-hardcode` |
| Phase 2 | Registry package | `go test ./backend/internal/ai/registry/... -race` + Resolve benchmark / startup budget |
| Phase 3 | targetjob retire + TargetJobs fixture provenance | `go test ./backend/internal/targetjob/... -race` + `make validate-fixtures` + checklist 3.6 retire grep gate returns 0 matches |
| Phase 4 | Upstream remediation + DB seed | `go test ./backend/internal/ai/aiclient/... -race` + `go test -tags=integration ./backend/internal/ai/registry/...` + `make migrate-check` |
| Phase 5 | Final reconcile | `make lint` + context validator + sync-doc-index check |

#### 8.1.1 Phase 0 handoff snapshot template

| Snapshot | Required content | Source artifact |
|---|---|---|
| Profile catalog | 10 default `model_profile_name` rows with status / capability / provider_ref / unsupported_reason | `config/ai-profiles.yaml` |
| DB schema gap | `prompt_versions` / `rubric_versions` seed field match plus current `ai_task_runs` prompt provenance missing columns | `migrations/000001_create_baseline.up.sql` |
| targetjob bridge | `StaticPromptRegistry` / `NewStaticPromptRegistry` / 4 `defaultTargetImport*` constants retire list and 7-field `PromptResolution` mapping | `backend/internal/targetjob/prompt_registry.go` + `parse_executor.go` |
| OpenAPI / fixture provenance | `GenerationProvenance` 6 fields, `importTargetJob` fixture path, `getTargetJob` fixture provenance slots, and ParseScreen polling consumer | `openapi/openapi.yaml` + `openapi/fixtures/TargetJobs/*.json` + `frontend/src/app/screens/parse/ParseScreen.tsx` |
| B1/B4/A3 provenance | `feature_key` vocabulary gap, `ai_task_runs` prompt provenance typed-column gap, and A3 metadata / writer field gap | `shared/conventions.yaml` + B4/A3 docs and code |

### 8.1.2 Phase 0 recorded snapshots

#### 8.1.2.1 Profile catalog snapshot (item 0.1)

Source: `config/ai-profiles.yaml` (recorded 2026-05-09). All 10 baseline default profiles are currently `status=active` and use `provider_ref=deepseek` (`deepseek-v4-flash` for chat default, `deepseek-v4-pro` for higher-reasoning routes). `resume.tailor.default` is shared between `resume.tailor.gap_review` and `resume.tailor.bullet_suggestions`, so the catalog has 9 unique profile rows for 10 feature_keys.

| feature_key | model_profile_name | status | capability | provider_ref | unsupported_reason |
|---|---|---|---|---|---|
| `target.import.parse` | `target.import.default` | active | chat | deepseek | – |
| `practice.session.first_question` | `practice.first_question.default` | active | chat | deepseek | – |
| `practice.session.follow_up` | `practice.followup.default` | active | chat | deepseek | – |
| `practice.turn.lightweight_observe` | `practice.turn_observe.default` | active | chat | deepseek | – |
| `report.generate` | `report.generate.default` | active | chat | deepseek | – |
| `report.question_assessment` | `report.assessment.default` | active | chat | deepseek | – |
| `resume.parse` | `resume.parse.default` | active | chat | deepseek | – |
| `resume.tailor.gap_review` | `resume.tailor.default` | active | chat | deepseek | – |
| `resume.tailor.bullet_suggestions` | `resume.tailor.default` (shared) | active | chat | deepseek | – |
| `debrief.generate` | `debrief.generate.default` | active | chat | deepseek | – |

Phase 5 `make lint-ai-profile-coverage` therefore only needs to assert (a) every spec §3.1.1 default profile name resolves in `config/ai-profiles.yaml`, and (b) any future flip to `disabled` / `unsupported` ships with a non-empty `unsupported_reason`. This plan does not modify any profile `status` field; profile lifecycle stays with `ai-provider-and-model-routing/004` and F3 002.

#### 8.1.2.2 DB schema gap snapshot (item 0.2)

Source: `migrations/000001_create_baseline.up.sql` L341-396 (recorded 2026-05-09).

`prompt_versions` columns match spec §4.1 verbatim: `id uuid PK, feature_key text not null, version text not null, language text not null default 'multi', template_hash text not null, template_body text not null, is_active boolean not null default false, created_at timestamptz not null default now(), UNIQUE(feature_key, version, language)`. Phase 4 seed migration writes exactly this column set.

`rubric_versions` columns match spec §4.1: `id uuid PK, feature_key, version, language default 'multi', schema_json jsonb not null, is_active boolean default false, created_at, UNIQUE(feature_key, version, language)`. Phase 4 seed writes the canonical rubric schema_json from yaml.

`ai_task_runs` typed columns currently provide model routing (`provider, model_family, model_id, model_profile_name, model_profile_version, route, language, prompt_version, rubric_version`) but **omit** the prompt/rubric provenance triple required by spec §6 C-9 + plan §3.1: `feature_key`, `feature_flag`, `data_source_version`. Today these only live in untyped `metadata jsonb`, which violates B4 typed-column policy. Phase 4 must add them as typed columns via additive migration `NNNNNN_add_ai_task_runs_prompt_provenance.{up,down}.sql` (or directly amend `000001` with explicit justification logged here, since dev baseline allows pre-launch revision).

D-7 active uniqueness is **not** enforced by DB today: there is no `UNIQUE INDEX ... (feature_key, language) WHERE is_active = true` partial index on either table. Phase 4.7 may add a partial unique index; otherwise registry runtime check covers D-7. Phase 4.7 checklist accepts either option.

Spec §4.1 requires `status ∈ {draft, active, deprecated}` for prompt yaml meta — this is yaml-level only and does not reflect into DB column; DB uses `is_active boolean` for staging/prod gating. The two should not be conflated.

#### 8.1.2.3 Targetjob bridge retire snapshot (item 0.3)

Source: `backend/internal/targetjob/prompt_registry.go` + `parse_executor.go` L17-44 (recorded 2026-05-09).

**Retire list** (Phase 3.1 must delete and Phase 3.3 negative-test must assert absent):

| Identifier | Kind | Location |
|---|---|---|
| `defaultTargetImportPromptVersion` | const | `prompt_registry.go:9` |
| `defaultTargetImportRubricVersion` | const | `prompt_registry.go:10` |
| `defaultTargetImportModelProfileName` | const | `prompt_registry.go:11` |
| `defaultTargetImportDataSourceVersion` | const | `prompt_registry.go:12` |
| `StaticPromptRegistry` | struct | `prompt_registry.go:18` |
| `NewStaticPromptRegistry` | func | `prompt_registry.go:24` |
| `(*StaticPromptRegistry).Resolve` | method | `prompt_registry.go:38` |

**Preserve** (Phase 3.1 RegistryAdapter must satisfy these unchanged):

| Identifier | Kind | Contract |
|---|---|---|
| `FeatureKeyTargetImportParse` | const = `"target.import.parse"` | parse pipeline still resolves through this constant |
| `PromptResolution` | struct (7 fields) | `PromptVersion / RubricVersion / ModelProfileName / DataSourceVersion / FeatureFlag / SystemMessage / UserMessageTemplate` |
| `PromptRegistryClient` | interface | `Resolve(ctx, featureKey, language) (PromptResolution, error)` |
| `ErrPromptUnsupported` | sentinel error | returned when (featureKey, language) is not enabled |

**Adapter contract** (Phase 3.1):

- `NewRegistryAdapter(client *registry.Client) *RegistryAdapter`
- `(*RegistryAdapter).Resolve(ctx, featureKey, language) (PromptResolution, error)` — must map all 7 fields and assert `registry.PromptResolution.FeatureKey == featureKey`; mismatch returns `ErrPromptUnsupported` to keep the targetjob fail-closed contract.

The bridge currently only resolves `target.import.parse`; the adapter must continue to satisfy that single feature_key for the parse pipeline while transparently exposing all 10 baseline feature_keys to other future C-domain consumers.

#### 8.1.2.4 OpenAPI / fixture provenance snapshot (item 0.4)

Source: `openapi/openapi.yaml` L1377-1411 (`GenerationProvenance` schema) + `docs/spec/backend-practice/spec.md` v1.3 L93 D-29 (recorded 2026-05-09).

**`GenerationProvenance` 6 required fields** (cross-layer assertion target for Phase 3.4 + 4.8):

| Field | OpenAPI type | Source authority |
|---|---|---|
| `promptVersion` | string | F3 RegistryClient.PromptResolution.PromptVersion |
| `rubricVersion` | string | F3 RegistryClient.PromptResolution.RubricVersion (literal `not_applicable` allowed for non-scoring) |
| `modelId` | string | A3 resolved adapter model id at call time (NOT registry's ModelProfileName — these are different layers) |
| `language` | string | BCP 47 tag from request context |
| `featureFlag` | string | A3 CallMetadata.FeatureFlag (literal `none` when no PostHog flag is active) |
| `dataSourceVersion` | string | F3 RegistryClient.PromptResolution.DataSourceVersion |

**D-29 frontload** (backend-practice/spec.md v1.3 L93): `prompt-rubric-registry/001-baseline` must independently derive, complete, and pass its Resolve / prompt / rubric / lint gates before backend-practice can enter implementation that depends on AI output (`practice.session.first_question`, `practice.session.follow_up`, `practice.turn.lightweight_observe`). This plan unblocks D-29 by Phase 5.6 closure.

**Cross-layer mapping nuance**: registry's `ModelProfileName` → `ai_task_runs.model_profile_name` typed column (Phase 4 already typed); A3's resolved `modelId` → `ai_task_runs.model_id` typed column → response `GenerationProvenance.modelId`. The wire layer surfaces `modelId` (not profile name) so Phase 3.4 fake AI writer must echo the resolved adapter model rather than the profile name.

**TargetJobs fixture provenance slots** (Phase 3.5 update target): `openapi/fixtures/TargetJobs/getTargetJob.json` carries `summary.provenance` and `fitSummary.provenance` examples; `openapi/fixtures/TargetJobs/importTargetJob.json` is `202 + Job` with `job.provenance`. Phase 3.5 must update example values to match `target.import.parse v0.1.0` baseline coordinates and the A3 `target.import.default` profile-resolved model. ParseScreen.tsx polling consumer is read-only.

#### 8.1.2.5 B1/B4/A3 provenance contract gap snapshot (item 0.5)

Source: `shared/conventions.yaml` + `backend/internal/shared/ai/vocabulary.go` + `backend/internal/ai/aiclient/{payload.go,meta.go,writers.go}` + `backend/internal/ai/aiclient/observability/decorator.go` + `migrations/000001_create_baseline.up.sql` (recorded 2026-05-09).

| Layer | Surface | feature_key | feature_flag | data_source_version | Phase 4 action |
|---|---|---|---|---|---|
| B1 | `shared/conventions.yaml` `ai_provenance_fields[]` | ✗ missing | ✓ present (L312) | ✓ present (L313) | 4.1 add `feature_key` row |
| B1 | `backend/internal/shared/ai/vocabulary.go` | ✗ no `FieldFeatureKey` | ✓ `FieldFeatureFlag` (L144) | ✓ `FieldDataSourceVersion` (L145) | 4.1 add constant + add to allowlist (L169) |
| B4 | `migrations/000001_create_baseline.up.sql` `ai_task_runs` | ✗ no typed col | ✗ no typed col | ✗ no typed col | 4.2 add 3 typed columns (additive migration); update `migrations_lint.py` to allow + require |
| B4 | `prompt_versions` / `rubric_versions` | ✓ typed | n/a | n/a | – |
| A3 | `aiclient.CallMetadata` | ✓ `FeatureKey` (payload.go:17) | ✗ missing | ✓ `DataSourceVersion` (payload.go:21) | 4.3 add `FeatureFlag` field |
| A3 | `aiclient.AICallMeta` | ✗ missing | ✗ missing | ✗ missing | 4.3 add 3 fields, freeze field order per spec §4.1 / ADR-Q6 §3.1 |
| A3 | `aiclient.AITaskRunRow` | ✗ missing | ✗ missing | ✗ missing | 4.3 add 3 fields; writer must reject empty values or use registered defaults (`none` / `not_applicable`) |
| A3 | `aiclient/observability/decorator.go` | ✗ no propagation | ✗ no propagation | ✗ no propagation | 4.3 propagate from CallMetadata → AICallMeta → AITaskRunRow + privacy red line update if needed |

**Conclusions for Phase 4 sequencing**:

1. B1 vocabulary update must land first because vocabulary.go is the cross-language source of truth — a downstream change to AICallMeta without B1 first would fail conventions drift lint.
2. B4 schema migration must land before A3 writer changes — A3 cannot write a typed column that doesn't exist. Phase 4.2 (B4 migration) must precede Phase 4.3 (A3 metadata/writer).
3. A3 CallMetadata is a public API surface on aiclient.AIClient.Complete callers; adding `FeatureFlag` is additive (new field), but every existing caller must populate it (with literal `none` if no flag is active). Phase 4.3 negative test must cover the empty-string rejection.
4. F1 metric labels must NOT add the 3 new columns — they are append-only provenance, not high-cardinality dashboards. Confirm in Phase 4.3 via existing F1 test or registered F1 spec.

### 8.1.3 Phase 1 evidence summary

- `make lint-prompts`: green (`prompt_lint: 20 files clean`).
- `make lint-rubrics`: green (`rubric_lint: 20 files clean`).
- `make lint-prompts-hardcode`: green (no hardcoded prompt assignments in `backend/internal/{practice,report,resume,debrief,targetjob}`).
- `python3 -m pytest scripts/lint/prompt_lint_test.py`: 4 passed (baseline / canonical hash vs README §3 / hash drift negative / field order negative).
- `python3 -m pytest scripts/lint/rubric_lint_test.py`: 4 passed (baseline / weight tolerance / dimension allowlist / missing weight negative).
- `python3 -m pytest scripts/lint/prompt_hardcode_lint_test.py`: 6 passed (default scan / raw string negative / long quoted negative / PromptVersion short-string passes / `_test.go` allowlisted / systemMessage flagged).
- `make migrate-check` substep `migrations_lint.py`: green (`migration lint: ok`); the `cmd/migrate ... check` substep requires `DATABASE_URL` and exits early in this local environment, which is the expected dev shape until Phase 4 dockertest lands.
- Pre-existing aggregate `make lint` warnings noted: `backend/internal/ai/aiclient/providers/minimax_speech/*.go` and `backend/internal/targetjob/handler.go targetJobId` (revive `var-naming`) are emitted by previously committed code (`f2f5fc9` and earlier). Phase 1 introduces no new lint violations; these pre-existing findings are out of scope for this plan and stay with the originating subspecs.

### 8.1.4 Phase 2 evidence summary

- `go test ./backend/internal/ai/registry/... -race`: green (TestTypeShape, TestJudgeSignature, TestLoadHappyPath, TestLoadHashDriftRejected, TestLoadMissingMarkdownBody, TestLoadInvalidYAML, TestResolveExactLanguage, TestResolveFallbackToMulti, TestResolveUnknownFeatureKey, TestResolveEmptyArgsRejected, TestGetPromptExact, TestGetRubricExact, TestCacheReloadIdempotent, TestCacheTTLDrivesReload, TestCacheConcurrentReadsAndReload, TestNotImplementedJudgeAlwaysFails, TestNewRegistryClientLoadsAllBaselines, TestNewRegistryClientRequiresDirs, TestNewRegistryClientRejectsOrphanFeatureKey, TestStartupBudget, TestResolveP95Budget).
- `go test -bench=. -benchtime=200x -run=^$ ./backend/internal/ai/registry/`: `BenchmarkResolve` ~235ns/op (well under spec C-3 5ms P95 budget).
- `! grep -rE "AIClient|aiclient\.|metric\.Counter|secret\.|Secret\.|os\.Getenv" backend/internal/ai/registry/`: green (zero matches; doc.go was rephrased so red-line documentation does not collide with the lint regex).
- `! grep -rE "github.com/.*/targetjob|aiclient\.|metric\.Counter" backend/internal/ai/registry/`: green.
- `go vet ./backend/internal/ai/registry/`: green.
- F3 RegistryClient publishes ResolveActive / GetPrompt / GetRubric, NotImplementedJudge default, atomic.Value-backed snapshot with 30s TTL + explicit Reload, FallbackCount counter for D-6 test assertions, and SnapshotSize accessor for cache idempotency tests.

### 8.1.5 Phase 3 evidence summary

- `go test ./backend/internal/targetjob -run TestRegistryAdapter -race`: green (TestRegistryAdapterMapsAllSevenFields, TestRegistryAdapterRejectsNilClient, TestRegistryAdapterRejectsUnknownFeatureKey).
- `go test ./backend/internal/targetjob -run TestActiveScopeNegative -race`: green (extended forbidden-token list now covers `StaticPromptRegistry`, `NewStaticPromptRegistry`, and all 4 `defaultTargetImport*` constants).
- `go test ./backend/internal/targetjob -run TestParseExecutor -race`: green; new `TestParseExecutorRegistryAdapterCrossLayer` and `TestParseExecutorMetadataCarriesF3Triple` exercise the F3 RegistryAdapter end-to-end against the on-disk `config/prompts/target.import.parse/v0.1.0*.yaml` baseline and confirm the 6-field provenance shape (`language`, `featureFlag`, `promptVersion`, `rubricVersion`, `modelId`, `dataSourceVersion`).
- `go test ./backend/cmd/api -race`: green; `TestBuildTargetJobRuntimeWiresDrainerAndAIClient` now points the loader at the in-repo `config/prompts` and `config/rubrics` roots through new `ai.promptsDir` / `ai.rubricsDir` config keys, replacing the retired `targetjob.NewStaticPromptRegistry()` wiring with `registry.NewRegistryClient` + `targetjob.NewRegistryAdapter`. cmd/api scenario test gets a local `staticTestPromptRegistry` shim mirroring the old shape so HTTP-level assertions stay stable.
- `go vet ./backend/internal/targetjob ./backend/cmd/api ./backend/internal/ai/registry`: green.
- `make validate-fixtures`: green; `openapi/fixtures/TargetJobs/getTargetJob.json` now uses F3 baseline coordinates (`promptVersion=v0.1.0`, `rubricVersion=v0.1.0|not_applicable`, `modelId=model-profile:target.import.default`, `featureFlag=none`, `dataSourceVersion=registry.v1`). `importTargetJob.json` path casing remains aligned with the operation matrix and the response is still `202 + Job` with no provenance shape change.
- `! grep -rE "StaticPromptRegistry|defaultTargetImport(Prompt|Rubric|ModelProfile|DataSource)" backend/` (production scope, `--include='*.go' --exclude='*_test.go'`): zero matches. The full recursive grep still hits the negative-test file `active_scope_negative_test.go` and the cmd/api scenario shim comment, both of which intentionally name the retired identifiers as forbidden tokens — these are regression guards, not residual production references. The 2.9-style filtering convention applies.

### 8.2 Owner handoff

- **B1 shared-conventions-codified**: `feature_key` joins `feature_flag` / `data_source_version` as AI provenance vocabulary; no F1 metric label expansion.
- **B4 db-migrations-baseline**: `ai_task_runs` must typed-carry prompt/rubric provenance columns in addition to existing model routing typed columns; migration lint must allow and require this exact exception for `ai_task_runs.feature_key`.
- **A3 ai-provider-and-model-routing**: `CallMetadata` / `AICallMeta` / `AITaskRunRow` carry `FeatureKey` / `FeatureFlag` / `DataSourceVersion`; writer rejects empty provenance or applies registered defaults only where the spec names them.
- **C4 backend-targetjob**: `StaticPromptRegistry` is retired; targetjob keeps its local `PromptRegistryClient` interface and receives F3 registry through `RegistryAdapter`.
- **F3 002-real-model-profile-and-evals**: receives Judge implementation, real model profile activation/eval, and prompt/rubric quality iteration after this baseline is verified.

### 8.3 Retrospective candidates

- `StaticPromptRegistry` retire timing versus backend-targetjob/001 phase boundary.
- `template_hash` canonical algorithm shared by Python lint and Go runtime.
- Whether future cross-layer plans should require `ai_task_runs` operation matrix rows before implementation starts.
