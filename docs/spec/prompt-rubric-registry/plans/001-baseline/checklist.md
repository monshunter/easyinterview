# F3 Baseline Registry, Resolve and Lint Gates Checklist

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Phase 1: Config truth source

- [x] 1.1 `config/prompts/README.md` 与 `config/rubrics/README.md` 固化文件布局、字段、hash、language coordinate、schema 和 dimension rules；验证: prompt/rubric lint tests 覆盖 README 约束。
- [x] 1.2 9 个 baseline feature_key 均有 `v0.1.0` prompt YAML/MD、语言无关 schema 与 rubric YAML；验证: `make lint-prompts`、`make lint-rubrics` 和 directory-count checks 通过。
- [x] 1.3 Prompt/rubric seed migration 与 config truth source 保持一致；验证: `prompt_lint.py` seed hash gate 与 `migrations_lint.py` static SQL gate 通过。

## Phase 2: Registry runtime

- [x] 2.1 `backend/internal/ai/registry` 提供 types、loader、resolver、cache、registry client、Judge interface 和 errors；验证: registry Go tests 覆盖 load/hash drift/resolve/cache/judge/startup budget。
- [x] 2.2 `ResolveActive` 返回 feature_key、prompt/rubric version、model profile、feature flag、data source version、templates 和 output schema；验证: registry resolver tests 与 schema lint 通过。
- [x] 2.3 Registry 包边界保持 provider-neutral；验证: boundary grep / tests 断言 registry 不直接调用 AIClient、不读 secret、不写业务 metric。

## Phase 3: TargetJob / provenance bridge

- [x] 3.1 `targetjob.RegistryAdapter` 将 registry resolution 映射到 parse pipeline；验证: adapter tests 显式断言 7 字段映射与 feature_key 一致。
- [x] 3.2 `backend/cmd/api` 注入 shared registry client + TargetJob adapter；验证: backend cmd/api 和 targetjob tests 通过。
- [x] 3.3 TargetJobs fixtures 与 OpenAPI `GenerationProvenance` 使用 A3 resolved model id + F3 prompt/rubric provenance；验证: `make validate-fixtures` 通过。

## Phase 4: B1/B4/A3 provenance and migrations

- [x] 4.1 B1 shared vocabulary、B4 migrations、A3 aiclient metadata/writer 都承载 `feature_key`、`feature_flag`、`data_source_version`；验证: shared/aiclient tests 与 migration lint 通过。
- [x] 4.2 `ai_task_runs` typed row 与 TargetJob fake writer tests 对齐；验证: targetjob cross-layer tests 断言 row fields 与 registry/A3 meta 一致。
- [x] 4.3 Baseline prompt/rubric seed migration idempotent 且 down SQL 范围限定；验证: static SQL parse gate 覆盖 9-key `multi` rows。

## Phase 5: Lint / handoff

- [x] 5.1 `lint-prompts`、`lint-rubrics`、`lint-prompts-hardcode`、`lint-ai-profile-coverage` 接入本地 Make targets；验证: 四个 Make targets 通过。
  <!-- verified: 2026-07-07 method=focused-registry-gates evidence="validate_context.py prompt-rubric-registry/001 backend PASS; targeted owner wording grep returned no matches; make lint-prompts PASS (9 files); make lint-rubrics PASS (9 files); make lint-prompts-hardcode PASS; make lint-ai-profile-coverage PASS; python3 scripts/lint/migrations_lint.py --repo-root . PASS; cd backend && go test ./internal/ai/registry/... ./internal/targetjob/... ./internal/ai/aiclient/... ./internal/shared/ai/... ./internal/shared/types/... -count=1 PASS; make validate-fixtures PASS (35 fixtures)" -->
- [x] 5.2 BDD-Gate: 不适用；验证: plan 明确该 owner 只提供内部 registry/config/lint/migration/adapter contract，用户可见 AI 行为由业务 owner 各自维护 BDD。
- [x] 5.3 Owner handoff 保留当前验证入口和与 F3 002/003/004 的分工；验证: context validator 和 sync-doc-index 通过。

## Phase 6: Rubric README stable spec reference

- [x] 6.1 Structural red: `config/rubrics/README.md` 是 active config/lint 文档中唯一仍固定引用 F3 `spec.md v2.9` 的位置
- [x] 6.2 删除固定版本号，保留可点击 spec 路径，不新增专用脚本
- [x] 6.3 运行 rubric lint、F3 001 context、index/docs/diff/pruning 与固定版本负向搜索，并确认状态为 `completed`
  <!-- verified: 2026-07-10 commands="make lint-rubrics; pytest rubric_lint_test.py; validate F3 001 context; index/docs/diff/pruning gates" result="pass; 9 rubric files clean; 6 tests pass; fixed-version search zero; real_residuals=0" -->

## Phase 7: Registry score-level conversion simplification

- [x] 7.1 将同构 `scoreLevelYAML` 显式转换为 `ScoreLevel`，删除逐字段复制并保持 rubric loader 合同（验证：registry package tests、`make lint-rubrics`、scoped `staticcheck`、owner context/docs gates）
  <!-- verified: 2026-07-10 method=registry-score-level-conversion-simplification evidence="S1016 red identified the repeated field copy. All TestLoad cases, rubric lint and scoped registry staticcheck PASS while active; owner/product contexts, sync-doc-index, docs-check, diff-check and pruning surface PASS real_residuals=0. Full registry package reruns after the owner completed header restores its preflight contract." -->
