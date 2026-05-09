# F3 Baseline Registry, Resolve and Lint Gates Checklist

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-09

**关联计划**: [plan](./plan.md)

## Phase 0: 上游契约复核（无代码 / 契约写入；只记录 handoff snapshot）

- [x] 0.1 抽出 `config/ai-profiles.yaml` 中 spec §3.1.1 列出的 10 个默认 `model_profile_name` 当前 `status` / `capability` / `provider_ref` / `unsupported_reason`，记入 plan §8.1.1 Phase 0 handoff snapshot，作为 Phase 5 `make lint-ai-profile-coverage` 预期输入。验证: 笔记小节包含 10 行；`grep -E "^(target|practice|report|resume|debrief)\\." /tmp/phase0-profile-snapshot.md` 命中 10 行
- [x] 0.2 复核 `migrations/000001_create_baseline.up.sql` L341-396 的 `prompt_versions` / `rubric_versions` / `ai_task_runs` schema，确认 Phase 4 seed migration 字段集与 spec §4.1 一致，并客观记录当前 `ai_task_runs` 缺少 `feature_key` / `feature_flag` / `data_source_version` typed columns 的缺口。验证: schema 字段对照表与缺口表写入 plan §8.1.1 Phase 0 handoff snapshot
- [x] 0.3 复核 `backend/internal/targetjob/prompt_registry.go` 与 `backend/internal/targetjob/parse_executor.go` L20-44，确认 `StaticPromptRegistry` retire scope（含 4 个 `defaultTargetImport*` 常量）与 `PromptResolution` 7 字段，且 `featureKey` 由 Resolve 入参和 A3 metadata 承载。验证: retire 标识符清单写入 plan §8.1.1 Phase 0 handoff snapshot
- [x] 0.4 复核 `docs/spec/backend-practice/spec.md` v1.3 D-29 + `openapi/openapi.yaml` `GenerationProvenance` 6 字段（`promptVersion` / `rubricVersion` / `modelId` / `language` / `featureFlag` / `dataSourceVersion`），作为 Phase 3 cross-layer assertion 对齐目标。验证: 字段对照表写入 plan §8.1.1 Phase 0 handoff snapshot
- [x] 0.5 复核 B1/B4/A3 provenance contract：`shared/conventions.yaml`、`backend/internal/shared/ai/vocabulary.go`、`docs/spec/shared-conventions-codified/spec.md`、`docs/spec/db-migrations-baseline/spec.md`、`backend/internal/ai/aiclient/{payload.go,meta.go,writers.go}` 与 observability decorator。验证: B1 缺 `feature_key`、B4 缺 `ai_task_runs` prompt provenance typed columns、A3 writer 缺三字段的事实表写入 plan §8.1.1 Phase 0 handoff snapshot

## Phase 1: Truth source 文件 + 3 套 lint script

- [x] 1.1 新增 `config/prompts/README.md` + `config/rubrics/README.md`：固定字段顺序、canonical hash 算法、维度命名 allowlist、multi-language 命名约定，与 `scripts/lint/prompt_lint.py` 共享算法描述。验证: lint 单测 `prompt_lint_test.py::TestCanonicalHashAgainstReadme` 红→绿
- [x] 1.2 落地 10 个 `config/prompts/<feature_key>/v0.1.0*.{yaml,md}` baseline（每个 feature_key ≥2 language，含 `multi`，文件命名遵守 spec D-2），template body 写真实可用文案，无「TBD」字样。验证: `find config/prompts -mindepth 1 -maxdepth 1 -type d | wc -l` = 10；`find config/prompts -mindepth 2 -name 'v0.1.0*.yaml' | wc -l` ≥ 20；`! grep -rE "\\bTBD\\b|placeholder" config/prompts/`
- [x] 1.3 落地 10 个 `config/rubrics/<feature_key>/v0.1.0*.yaml` baseline（每个 feature_key ≥2 language，含 `multi`，且与 prompt language 集合一致），每个 ≥3 dimensions、weight 总和 = 1.0、score_levels ≥3 段，维度名命中 spec §4.1 allowlist。验证: `find config/rubrics -mindepth 1 -maxdepth 1 -type d | wc -l` = 10；`find config/rubrics -mindepth 2 -name 'v0.1.0*.yaml' | wc -l` ≥ 20；`rubric_lint_test.py::TestWeightSumTolerance` + `TestDimensionNameAllowlist` 通过
- [x] 1.4 实现 `scripts/lint/prompt_lint.py` + `prompt_lint_test.py`，覆盖字段顺序、status enum、SemVer、template_hash drift、language 枚举、canonical algorithm。验证: `python3 -m pytest scripts/lint/prompt_lint_test.py` 全绿；2 个 negative fixture（hash drift / 字段顺序错乱）红→绿
- [x] 1.5 实现 `scripts/lint/rubric_lint.py` + `rubric_lint_test.py`，覆盖 dimensions 非空、weight sum 容差、score_levels schema、维度 allowlist。验证: `python3 -m pytest scripts/lint/rubric_lint_test.py` 全绿；2 个 negative fixture（缺 weight / 违反 allowlist）红→绿
- [x] 1.6 实现 `scripts/lint/prompt_hardcode_lint.py` + `prompt_hardcode_lint_test.py`，AST 扫 `backend/internal/{practice,report,resume,debrief,targetjob}/**/*.go` 中 `prompt :=` / `Prompt = "..."` / `systemMessage := "..."` 字面量。验证: 注入 `backend/internal/practice/` 临时 `prompt := "..."` fixture → exit 1；移除后 → exit 0
- [x] 1.7 在 `Makefile` 追加 `lint-prompts` / `lint-rubrics` / `lint-prompts-hardcode` target，并加入聚合 `lint`；migration lint 仍由 `make migrate-check` 覆盖。验证: `make lint-prompts && make lint-rubrics && make lint-prompts-hardcode && make lint && make migrate-check` 全绿
- [x] 1.8 Regression-negative scope grep：`! grep -rE "mistakes|growth|drill|mistake.extract" config/prompts/ config/rubrics/` 通过；spec §3.1.1 删除的 C11 资料检索类占位不出现。验证: grep 命令 0 行命中（写入 plan §8 handoff Phase 1 收口）

## Phase 2: `backend/internal/ai/registry/` Go 包

- [ ] 2.1 新增 `backend/internal/ai/registry/doc.go`：包注释 + 红线（不依赖 targetjob、不调 aiclient、不写业务 metric / log）。验证: `! grep -rE "github.com/.*/targetjob|aiclient\.|metric\.Counter" backend/internal/ai/registry/` 通过；如需在 doc 注释中记录红线，测试需显式过滤注释而不是放宽源码命中
- [ ] 2.2 实现 `types.go`：`PromptResolution`（包含 `FeatureKey` + 兼容 targetjob 7 字段 + D-12 预留 Tools/OutputSchema/StreamWire）、`PromptMeta`、`RubricSchema`、`RubricDimension`、`ScoreLevel`、`Judge` interface（签名 = spec D-9）、错误类型。验证: `go test ./backend/internal/ai/registry -run TestTypeShape -race` 通过；反射断言 `Judge` 入参顺序与 spec D-9 一致
- [ ] 2.3 实现 `loader.go` + `loader_test.go`：扫 `config/prompts/` + `config/rubrics/`，逐项校验 template_hash drift；canonical algorithm 与 `prompt_lint.py` 共享。验证: `go test ./backend/internal/ai/registry -run TestLoad -race` 全绿（happy / hash drift / missing meta / yaml 解析失败）
- [ ] 2.4 实现 `resolver.go` + `resolver_test.go`：`ResolveActive` 精确 language → `multi` fallback；unknown feature_key → `ErrPromptUnsupported`；空字符串拒绝。验证: `go test ./backend/internal/ai/registry -run TestResolve -race` 全绿（精确 language / fallback to multi 含 warn 计数 / unknown feature_key / unknown language warn / 空字符串）
- [ ] 2.5 实现 `cache.go` + `cache_test.go`：atomic.Value snapshot + 30s TTL + Reload 钩子。验证: `go test ./backend/internal/ai/registry -run TestCache -race` 全绿（TTL expiry / Reload idempotent ×5 / 100 goroutine 并发 + Reload 交错无 race）
- [ ] 2.6 实现 `judge.go` + `judge_test.go`：`NotImplementedJudge` 默认实现，始终返回 `ErrJudgeNotImplemented`；签名 freeze 反射断言。验证: `go test ./backend/internal/ai/registry -run TestJudge -race` 全绿
- [ ] 2.7 实现 `registry.go` + `registry_test.go`：`NewRegistryClient(opts)` 全量加载；不允许悬空 feature_key（prompt 有 rubric 无或反之 → 启动 fail）。验证: `go test ./backend/internal/ai/registry -run TestNewRegistryClient -race` 全绿
- [ ] 2.8 实现 `perf_test.go`：`BenchmarkResolve` P95 ≤ 5ms 硬断言；`TestStartupBudget` < 1s。验证: `go test -bench=. -benchtime=200x ./backend/internal/ai/registry/...` 输出 P95 通过；`go test -run TestStartupBudget ./backend/internal/ai/registry/...` 通过
- [ ] 2.9 边界 grep red-line：registry 包不持 secret、不调 AI、不写 metric、不读 env。验证: `! grep -rE "AIClient|aiclient\.|metric\.Counter|secret\.|Secret\.|os\.Getenv" backend/internal/ai/registry/` 通过；如需在 doc 注释中记录红线，测试需显式过滤注释而不是放宽源码命中

## Phase 3: targetjob StaticPromptRegistry retire + cross-layer 对齐

- [ ] 3.1 替换 `backend/internal/targetjob/prompt_registry.go`：删除 `StaticPromptRegistry` + 4 个 `defaultTargetImport*` 常量；新增 `RegistryAdapter`，构造 `NewRegistryAdapter(client *registry.Client)`，实现 `targetjob.PromptRegistryClient.Resolve` 并映射 7 字段，同时断言 Resolve 入参 `featureKey` 与 registry `FeatureKey` 一致。验证: `go test ./backend/internal/targetjob -run TestRegistryAdapter -race` 全绿，单测显式列出 7 字段映射
- [ ] 3.2 更新 `backend/cmd/api/` DI wire：`NewStaticPromptRegistry()` → `registry.NewRegistryClient(opts)` + `targetjob.NewRegistryAdapter(...)`；ParseExecutor 与未来 C5/C6/C7 handler 共享同一 `*registry.Client`。验证: `go build ./backend/cmd/api/...` 通过；`go vet ./backend/...` 无未使用 import
- [ ] 3.3 新增 `backend/internal/targetjob/active_scope_negative_test.go`：用 `go/types` 断言 package 不再 export `StaticPromptRegistry` / `NewStaticPromptRegistry` / 4 个 `defaultTargetImport*`；断言 `mistakes` / `growth` / `drill` / `mistake.extract` 不出现在 targetjob test fixture / golden output。验证: `go test ./backend/internal/targetjob -run TestActiveScopeNegative -race` 全绿
- [ ] 3.4 在 `parse_executor_test.go` 已有 fake AI writer 路径追加 cross-layer assertion：① ParseExecutor 收到的 7 字段 PromptResolution 与 `config/prompts/target.import.parse/v0.1.0*.yaml` meta 一致；② `provenance` JSON 含 `promptVersion / rubricVersion / modelId / language / featureFlag / dataSourceVersion` 6 字段。验证: `go test ./backend/internal/targetjob -run TestParseExecutor -race` 全绿
- [ ] 3.5 更新 `openapi/fixtures/TargetJobs/getTargetJob.json` 中 `summary.provenance` / `fitSummary.provenance` 示例，与 F3 `target.import.parse` baseline 坐标和 B2 `GenerationProvenance` 6 字段一致；确认 `openapi/fixtures/TargetJobs/importTargetJob.json` 路径大小写与 operation matrix 一致且 request/response shape 不变。验证: `make validate-fixtures` 全绿；`frontend/src/app/screens/parse/ParseScreen.tsx` 仍通过 generated `client.getTargetJob` polling
- [ ] 3.6 Phase 3 退出 grep red-line：`! grep -rE "StaticPromptRegistry|defaultTargetImport(Prompt|Rubric|ModelProfile|DataSource)" backend/` 通过；`go vet ./backend/...` 无警告。验证: grep 命令 0 行命中（写入 plan §8 handoff Phase 3 收口）

## Phase 4: B1/B4/A3 provenance remediation + DB seed migration

- [ ] 4.1 修订 B1 AI vocabulary：`docs/spec/shared-conventions-codified/spec.md`、`shared/conventions.yaml`、`backend/internal/shared/ai/vocabulary.go` 在既有 `feature_flag` / `data_source_version` 旁增加 `feature_key`，并说明三者不进入 F1 metric labels。验证: B1 generator/lint 测试缺 `feature_key` 时红，补齐后绿
- [ ] 4.2 修订 B4 `ai_task_runs` schema contract：`docs/spec/db-migrations-baseline/spec.md` C-12 / D-15 / §2.1 增加 `feature_key` / `feature_flag` / `data_source_version` typed columns；按 migration workflow 新增 `NNNNNN_add_ai_task_runs_prompt_provenance.{up,down}.sql` 或在 dev baseline 允许时修订 `000001` 并记录原因；更新 `scripts/lint/migrations_lint.py` 允许且要求 `ai_task_runs.feature_key`。验证: `make migrate-check` 全绿；migration lint 对缺列 fixture 失败
- [ ] 4.3 修订 A3 aiclient metadata / writer：`CallMetadata` 增加 `FeatureFlag`；`AICallMeta` / `AITaskRunRow` / observability decorator 显式携带 `FeatureKey` / `FeatureFlag` / `DataSourceVersion`。验证: `go test ./backend/internal/ai/aiclient/... -race` 全绿；缺三字段 negative test 红→绿
- [ ] 4.4 新增 `migrations/NNNNNN_seed_baseline_prompt_rubric_versions.up.sql`：写入 10 个 feature_key × N language baseline 行（`version='v0.1.0'`、`is_active=true`、`template_hash` 与 yaml 一致、`schema_json` 来自 rubric yaml canonical 化）；`INSERT ... ON CONFLICT DO NOTHING`。验证: `make migrate-up` 在 dockertest 环境绿；二次执行无 row 增量（idempotent）
- [ ] 4.5 新增 `migrations/NNNNNN_seed_baseline_prompt_rubric_versions.down.sql`：`DELETE FROM prompt_versions WHERE version='v0.1.0' AND feature_key IN (...)`，rubric 同；范围限定避免误删后续版本。验证: `make migrate-down && make migrate-up` 后 row 数与首次 up 一致
- [ ] 4.6 增强 `scripts/lint/prompt_lint.py`：扫 `migrations/*seed_baseline*.up.sql` 中 `template_hash` 与对应 `config/prompts/<feature_key>/v0.1.0*.yaml` 的 `template_hash` 一致性。验证: 注入 hash 不一致 fixture → `make lint-prompts` exit 1；恢复后 → exit 0
- [ ] 4.7 新增 `backend/internal/ai/registry/db_integration_test.go`（`//go:build integration`）：dockertest 拉 PG，跑当前 migration chain + 本 plan seed migration，断言 prompt_versions / rubric_versions 行覆盖 10 个 feature_key × 实际 language 坐标（至少 20 行）且 prompt/rubric language 集合一致 + `is_active=true` + `template_hash` 与 yaml 一致 + `ai_task_runs` 三个 provenance typed columns 存在 + 同 feature_key 第二行 active 被拒（unique partial index 或 RegistryClient 运行时校验）。验证: `go test -tags=integration ./backend/internal/ai/registry/...` 全绿
- [ ] 4.8 在 `parse_executor_test.go` 已有 fake writer 路径追加 `ai_task_runs.feature_key / prompt_version / rubric_version / model_profile_name / data_source_version / feature_flag` 6 字段 cross-layer assertion（无需 dockertest），并确认 F1 不新增高基数 metric labels。验证: `go test ./backend/internal/targetjob -run TestParseExecutorAITaskRuns -race` 全绿

## Phase 5: 收口 + A3 coverage gate + sync-doc-index

- [ ] 5.1 跑 `make lint-ai-profile-coverage`：验证 spec §3.1.1 列出的 10 个默认 `model_profile_name` 全部存在于 `config/ai-profiles.yaml`，`disabled` / `unsupported` 携带合法 capability/provider_ref + `unsupported_reason`；本 plan 不修改任何 profile status。验证: `make lint-ai-profile-coverage` 退出码 0；profile catalog 当前状态写入 plan §8 handoff Phase 5 收口
- [ ] 5.2 跑顶层 `make lint`（含 `lint-prompts` / `lint-rubrics` / `lint-prompts-hardcode` / `lint-ai-profile-coverage`）+ `make migrate-check` + 全量 verification one-liner。验证: 全绿；输出写入 plan §8 handoff
- [ ] 5.3 确认 `docs/spec/prompt-rubric-registry/plans/INDEX.md` 已写入 001-baseline active 行；维护 `docs/spec/prompt-rubric-registry/history.md` v2.1 entry；`spec.md` Header 保持 v2.1 并同步 `docs/spec/INDEX.md`。验证: `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` 通过
- [ ] 5.4 跑 `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/prompt-rubric-registry/plans/001-baseline/context.yaml --docs-root docs --target backend`。验证: 退出码 0
- [ ] 5.5 §8 handoff 记录 retrospective 候选清单：StaticPromptRegistry retire 节奏 / template_hash Go ↔ Python canonical 算法对齐策略 / B1/B4/A3 provenance remediation 对后续 F1 / backend-practice plan 的影响；不直接生成 retrospective 报告。验证: handoff 小节存在；由用户决定是否触发 `/retrospective`
- [ ] 5.6 全体 11 AC 自检：plan §8 handoff 列出 C-1~C-11 × 命令证据 × 通过判据；将 plan/checklist Header 切到 `completed`，同步 INDEX 与工作日志。验证: 11 AC 全过；INDEX 显示 completed；最近 `docs/work-journal/` 条目链接到本 plan

## BDD-Gate

> **BDD 不适用**: 本 plan 是 F3-owned 内部 Go 包（`backend/internal/ai/registry/`）+ `config/` truth source + `scripts/lint/` 静态校验，不新增用户可见 UI、HTTP API 行为或端到端业务工作流；TargetJobs fixture 仅做既有 provenance 示例对齐并由 `make validate-fixtures` 验证。Resolve 契约是 Go interface 与 yaml schema，不可通过浏览器或外部 API 触发。后续 P0 用户行为（first_question / followup / hint / report 等）由 `backend-practice` / `backend-report` / `backend-resume` / `backend-debrief` 各自 plan 维护 BDD/E2E gate。
>
> **替代验证 gate**:
>
> 1. `make lint-prompts && make lint-rubrics && make lint-prompts-hardcode && make lint-ai-profile-coverage` 静态门禁
> 2. `go test ./backend/internal/ai/registry/... -race`（含 `BenchmarkResolve` P95 + `TestStartupBudget` 1s budget）
> 3. `go test ./backend/internal/targetjob/... -race` 跨包契约对齐 + active-scope negative grep
> 4. `go test ./backend/internal/ai/aiclient/... -race` 覆盖 `CallMetadata` / `AICallMeta` / `AITaskRunRow` provenance 映射
> 5. `go test -tags=integration ./backend/internal/ai/registry/...` dockertest seed migration cross-layer
> 6. `make migrate-check` 覆盖 B4 schema / lint / down-up 闭环
> 7. `make validate-fixtures` 验证 TargetJobs fixtures 与 B2 schema / provenance shape
> 8. grep red-line：旧 `StaticPromptRegistry` / 4 个 `defaultTargetImport*` 常量 / 业务包 hardcode prompt / 已退役 `mistakes` / `growth` / `drill` / `mistake.extract` 模块名
> 9. `python3 .agent-skills/implement/shared/scripts/validate_context.py` + `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
