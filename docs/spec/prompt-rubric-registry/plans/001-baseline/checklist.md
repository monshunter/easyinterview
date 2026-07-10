# F3 Baseline Registry, Resolve and Lint Gates Checklist

> **版本**: 1.10
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

## Phase 8: Feature key helper removal

- [x] 8.1 删除零消费者 `featurekeys.All` / `FeatureKey.String` 与不存在的 future-codegen 说明，只保留 typed constants；验证：`deadcode -test`、exact symbol inventory、registry/shared package tests、events/profile/prompt lints、scoped `staticcheck` 与 owner docs gates
  <!-- verified: 2026-07-10 method=feature-key-helper-removal evidence="deadcode -test RED reported featurekeys.All; exact inventory found no explicit String consumer and whylive only reached it through generic fmt.Stringer dispatch. Removed both helpers and the unsupported future-codegen claim while retaining all nine typed constants. Events/prompt/profile lints, staticcheck, reachability and symbol inventory PASS; owner-completed focused tests verify runtime/preflight behavior." -->

## Phase 9: Pytest alias cleanup

- [x] 9.1 Rename the three `Test*` lint-test bodies to their pytest-collected `test_*` names and delete forwarding aliases; verify collection remains 22 tests, prompt/rubric suites and Make lints pass, alias inventory is zero, and owner contexts plus docs/diff/pruning gates pass.
  <!-- verified: 2026-07-10 method=prompt-rubric-pytest-alias-cleanup evidence="Collect-only RED showed only the three lowercase wrappers were collected while each uppercase body had one caller. After merging, collection remains 22 and all 22 pass; four prompt/rubric/profile lints, py_compile, alias inventory and owner contexts PASS." -->

## Phase 10: Prompt linter import cleanup

- [x] 10.1 删除 `scripts/lint/prompt_lint.py` 中零读取的 `Iterable` import；验证 AST import inventory、prompt tests/lints、registry consumers、owner contexts 与 docs/diff/pruning gates。
  <!-- verified: 2026-07-10 method=prompt-linter-unused-import-removal evidence="AST RED identified Iterable as the sole unused production Python import. Deleted it without replacement. Prompt/hardcode 22 tests, four F3 lints, migration/fixture gates and all registry/TargetJob/A3/shared Go consumers PASS. Two owner-status preflights failed while the plan was intentionally active and passed after completed restoration; both owner contexts and docs/diff/pruning gates PASS." -->

## Phase 11: Shared config-root test support

- [x] 11.1 RED: registry, benchmark, TargetJob and cmd/api contain four local walkers with the same `backend/go.mod` search and config path projection.
  <!-- red: 2026-07-10 method=config-root-walker-dupl evidence="dupl -t 100 reports cmd/api, registry benchmark and TargetJob walkers as one clone group; registry loader carries the same implementation and all four serve identical config roots." -->
- [x] 11.2 Add a tested `internal/testsupport.ConfigRoots(testing.TB)` and migrate all consumers; delete `repoConfigRoots`, `benchConfigRoots` and `repoConfigPromptsRubrics` definitions with exact-name zero-reference.
  <!-- verified: 2026-07-10 method=shared-config-roots-green evidence="ConfigRoots resolves both current directories in its unit test and serves 26 test/benchmark calls. Four local walkers and exact symbols are gone; integration-tag cmd/api compiles, focused registry/TargetJob/cmd-api tests pass and full backend dupl drops from nine to eight groups." -->
- [x] 11.3 BDD-Gate: 不适用；运行 testsupport/registry/TargetJob/cmd-api focused tests、full backend/vet/staticcheck、F3 lints、contexts 和 docs/pruning gates。
  <!-- verified: 2026-07-10 method=shared-config-roots-closeout evidence="Completed-state testsupport/registry/TargetJob/cmd-api and full backend suites pass; integration-tag cmd/api compiles, vet/staticcheck and four F3 lints pass. Owner/product contexts, index/docs/diff and pruning gates pass with real_residuals=0." -->
