# Prompt Rubric Registry 001-baseline 交付复盘报告

> **日期**: 2026-05-09
> **审查人**: Claude

**关联计划**: [prompt-rubric-registry/001-baseline](../spec/prompt-rubric-registry/plans/001-baseline/plan.md)

## 1 复盘范围与成功证据

本次交付把 [prompt-rubric-registry spec](../spec/prompt-rubric-registry/spec.md) v2.1 的 §3.1.1 10 个 baseline `feature_key`、§3.1 D-1~D-12 决策与 §6 C-1~C-11 验收端到端落到一个独立可部署、可验证的 vertical slice 中：

- **truth source**：`config/prompts/<feature_key>/v0.1.0[.<lang>].{yaml,md}` 与 `config/rubrics/<feature_key>/v0.1.0[.<lang>].yaml` 共 60 个文件，10 feature_keys × 2 language 坐标，`template_hash` canonical 算法在 `config/prompts/README.md` §3 描述并由 Python lint + Go loader 共用
- **Go 包**：`backend/internal/ai/registry/`（doc.go / types.go / loader.go / resolver.go / cache.go / judge.go / registry.go + 6 个 _test.go），`ResolveActive` benchmark 235 ns/op，`TestStartupBudget` < 1s
- **lint gates**：`scripts/lint/{prompt_lint.py,rubric_lint.py,prompt_hardcode_lint.{go,py}}` 三套 + Makefile target 加入聚合 `lint`，14 个 unit tests 全过
- **targetjob retire**：`StaticPromptRegistry` + 4 个 `defaultTargetImport*` 常量删除，`RegistryAdapter` 映射 7 字段，`backend/cmd/api` DI 切到 `registry.NewRegistryClient`
- **B1/B4/A3 provenance remediation**：B1 vocabulary 加 `feature_key`，B4 `ai_task_runs` 三个 typed columns，A3 `CallMetadata.FeatureFlag` + `AICallMeta` / `AITaskRunRow` / decorator 全链路承载
- **DB seed migration**：`migrations/000002_seed_baseline_prompt_rubric_versions.{up,down}.sql` 写 20 prompt + 20 rubric rows，`ON CONFLICT DO NOTHING` 保证 idempotent

成功证据：

| 证据 | 命令 / 输出 |
|------|-------------|
| `make lint-prompts` | `prompt_lint: 20 files clean` |
| `make lint-rubrics` | `rubric_lint: 20 files clean` |
| `make lint-prompts-hardcode` | exit 0（zero hardcoded prompts） |
| `make lint-ai-profile-coverage` | `ai_profile_coverage: OK` |
| `python3 scripts/lint/migrations_lint.py` | `migration lint: ok` |
| `python3 scripts/lint/conventions_drift.py` | `OK: 10 conventions generated outputs match shared/conventions.yaml` |
| `make validate-fixtures` | `OK — 34 fixtures` |
| `go test ./backend/... -race`（registry / targetjob / cmd-api / aiclient / shared-ai / shared-types） | 全过；`BenchmarkResolve` ≈ 235 ns/op；`TestStartupBudget` < 1s |
| `python3 -m pytest scripts/lint/{prompt_lint_test,rubric_lint_test,prompt_hardcode_lint_test,migrations_lint_test}.py` | 27 passed |
| `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` | Zero drift detected |
| `python3 .agent-skills/implement/shared/scripts/validate_context.py` | exit 0 |
| 11 AC self-check（plan §8.4） | C-1~C-9、C-11 PASS；C-5 / C-10 按 spec §2.2 推到 002 / 003 |

6 个 phase commits 已推到 `feat/prompt-rubric-registry-001-baseline-0509`：`d3842e8` Phase 0 → `8c45d4f` Phase 1 → `740fc44` Phase 2 → `3022998` Phase 3 → `77dc848` Phase 4 → `4cbe760` Phase 5。

## 2 会话中的主要阻点/痛点

- **README ↔ 验证 grep 自咬**
  - **证据**：Phase 1.2 `! grep -rE "\bTBD\b|placeholder" config/prompts/` 与 Phase 1.8 / 3.6 `! grep -rE "mistakes|growth|drill|mistake.extract"` 都因为 README / doc.go 字面提到禁词而失败；Phase 2.1 / 2.9 `! grep -rE "AIClient|aiclient\.|metric\.Counter|secret\.|os\.Getenv"` 也因 doc.go 红线说明命中 `aiclient.AIClient.Complete` / `os.Getenv` 字面量而失败。
  - **影响**：3 处 README + 1 处 doc.go 各被改写 1-2 次，每次都要二次 verify；plan §3.6 验证最终通过 production-only `--include='*.go' --exclude='*_test.go'` 才能与 spirit 对齐，文字 gate 命令仍保留宽口径。

- **dockertest 假设与现实不符**
  - **证据**：plan §4.7 明确写 "沿用 backend/internal/targetjob/store_test.go 已有 PG fixture 约定"，但实际 `targetjob/store_test.go` 用 `go-sqlmock`，仓库无 dockertest / pgtestdb harness；`grep -rn "dockertest\|pgtestdb" backend/` 0 行命中。
  - **影响**：`db_integration_test.go` 降级为 static SQL parse + cross-file `template_hash` 对照（仍覆盖行数 / feature_key 集合 / hash drift 三项断言），`//go:build integration` 标签未启用；真正的 PG up→down→up 验证仍由 `make migrate-check` 在 `DATABASE_URL` 配置时承担。

- **migrations_lint feature_key 边界对 seed migration 不友好**
  - **证据**：Phase 4.4 seed migration 落地后 `make migrate-check` 的 lint 子段直接报 `feature_key is only allowed in ai_task_runs, prompt_versions, rubric_versions`，原因：原始 `validate_feature_key_scope` 把 SQL 全文按 CREATE TABLE 边界切，对 INSERT VALUES 中的 `feature_key` 列名误报；进一步 `[^;]*;` 的 DML span 检测被 prompt body 内的 `;` 字符截断，强行展开 dollar-quoted block 与 SQL `--` 注释才稳定。
  - **影响**：Phase 4.2 + 4.4 之间至少 3 轮迭代调整 lint 解析逻辑；测试 fixture 也要从 ai_task_runs 改成 users 才能继续验证 allowlist 边界。

- **plan 引用的测试名称在仓库不存在**
  - **证据**：plan §4.8 验证命令 `go test ./backend/internal/targetjob -run TestParseExecutorAITaskRuns -race`，但 targetjob 包没有这个名字的测试；最接近的是 `pipeline_test.go` 中 `TestParseExecutor_*` 系列（fakeRegistry 路径），且 `parse_executor_test.go` 文件本身不存在。
  - **影响**：4.8 的 6 字段 `ai_task_runs` 断言最终落到 `aiclient/observability/decorator_test.go::TestDecorator_SuccessIncrementsRunsAndLogsCompleted`（更准确，因为只有 production decorator 才写 ai_task_runs），但 plan 引用的命令字面执行会报 "no tests to run"。

- **conventions parity fixture 与 codegen 不联动**
  - **证据**：B1 vocabulary 加 `feature_key` 后 `make codegen-conventions` 只更新 `frontend/src/lib/conventions/ai.ts`，没有触发 `shared/fixtures/conventions-parity.json` 同步；`backend/internal/shared/types/conventions_parity_test.go::TestConventionsParityFixture_ErrorCodesAndAIVocabulary` 因此报错 "AllFieldNames length = 22, want 21"，需要手工编辑 fixture。
  - **影响**：单文件手编但增加一个不显然的步骤，Phase 4.1 的 "B1 测试缺 `feature_key` 时红，补齐后绿" 节奏被 fixture drift 中途打断 1 次。

- **cmd/api test 缺乏 truth-source 路径配置位**
  - **证据**：Phase 3.2 切 `registry.NewRegistryClient` 之后 `TestBuildTargetJobRuntimeWiresDrainerAndAIClient` 立刻报 `lstat config/prompts: no such file or directory`，因为测试在 `cmd/api/` 工作目录下 relative path 无法解析；解决方案是新增 `ai.promptsDir` / `ai.rubricsDir` config keys + `repoConfigPromptsRubrics` 测试 helper 走 `go.mod` upward search。
  - **影响**：Phase 3.2 跨 1 次 build 失败 + 1 次测试失败，最终通过添加配置位收口；config keys 是新接口，未在 plan §3.2 明文要求，但是 retire 后必然需要的。

## 3 根因归类

- **README / doc 注释与负向 grep gate 字面冲突**
  - **类别**：spec-plan
  - **说明**：plan 各 phase 的负向 grep 命令（`! grep -rE ...`）按裸 token 匹配，没有为 documentation 留出 `--include='*.go' --exclude='*_test.go'` 之类的过滤约定；当 README / doc.go 合理需要描述禁词时只能反复改文案。

- **plan 引用了仓库不存在的 dockertest harness 与测试名**
  - **类别**：spec-plan
  - **说明**：plan derivation 阶段直接复制了 "PG fixture 约定" 与 `TestParseExecutorAITaskRuns` 这类预期标识，没有在派生时 grep 仓库验证；落地阶段被迫降级 / 改名。

- **`migrations_lint.py` 的 SQL 解析对 seed migration 过于脆弱**
  - **类别**：skill
  - **说明**：`validate_feature_key_scope` 用裸 regex 切 CREATE TABLE 块，对 INSERT VALUES + dollar-quoted body + 行内注释的真实 seed migration 多次误判；`migrations_lint.py` 缺少 Postgres-aware 解析或 explicit DML allowlist。

- **`conventions parity` fixture 不在 `make codegen-conventions` 范围内**
  - **类别**：skill
  - **说明**：`scripts/lint/conventions_drift.py` 与 `make codegen-conventions` 只对齐 TS generated；`shared/fixtures/conventions-parity.json` 是 Go side parity gate 的 source of truth 但靠手动维护，drift 检测要靠跑测试才能发现。

- **cmd/api 切外部 truth source 路径时缺少 config-key 模板**
  - **类别**：spec-plan（可同时考虑 README）
  - **说明**：plan §3.2 只描述 wire 替换，没有把 "新增 config keys + 测试 helper" 列为子步骤；任何引入 filesystem-backed runtime client 的 plan 都会遇到同样问题。

- **README/grep 一次性误改 / textwrap 一次性 bug / 工作目录瞬态变化**
  - **类别**：no repo change needed
  - **说明**：会话内一次性误差，已在落地中修正，不构成流程缺陷。

## 4 对流程资产的改进建议

- **plan 派生时强制对仓库实际存在的 harness / 测试名做 grep 验证**
  - **落点**：`/design` 或 `/plan-review` skill
  - **优先级**：high
  - **建议**：在 plan derivation / L1 review 阶段加入一个步骤：对每个验证命令中提到的 test 名 / harness 名 / 工具名，跑一遍 `grep -r "<token>" backend/ scripts/` 或 `find` 验证存在；不存在的需要在 plan §6 风险与应对里明确降级路径，而不是把假设当事实。

- **plan 验证 grep 命令引入统一过滤约定**
  - **落点**：`/plan-review` 与 `prompt-rubric-registry` 后续 plan、其他 retire/regression-negative 主题 plan
  - **优先级**：high
  - **建议**：约定 production-scope grep 默认带 `--include='*.go' --exclude='*_test.go'`（或对应 markdown / yaml 等扩展），test/_test.go / docs README 中的字面引用作为 regression guard 而非违规；同时把这个 idiom 写入 `/plan-review` 的 checklist。

- **`scripts/lint/migrations_lint.py` 升级到 Postgres-aware 解析或显式 DML allowlist**
  - **落点**：skill `/plan-code-review` + `scripts/lint/migrations_lint.py`
  - **优先级**：medium
  - **建议**：把 dollar-quoted body 与 single-line `--` 注释 stripping 沉淀为标准前置步骤；DML span 检测从 `[^;]*;` 升级为按 `;` 边界 + 引号感知的语句切分（参考 sqlparse / pglast）；或把 "feature_key column scope" 检查显式分成 DDL-only 与 DML-only 两个独立 pass。

- **conventions parity fixture 纳入 `make codegen-conventions` 维护范围**
  - **落点**：skill `/codegen-conventions` 或 `scripts/lint/conventions_drift.py`
  - **优先级**：medium
  - **建议**：让 `make codegen-conventions` 在更新 `frontend/src/lib/conventions/ai.ts` 之外，也根据 `shared/conventions.yaml` 自动生成 `shared/fixtures/conventions-parity.json`（或至少 `aiVocabularyFields` 节）；如保留手维，在 drift 出现时直接给出 `--fix` 提示，而不是要求开发者跑跨包测试才看到。

- **runtime filesystem-backed client 切换 plan 模板加入 config-key 子步骤**
  - **落点**：spec-plan template / `/design` skill
  - **优先级**：medium
  - **建议**：当某个 plan 引入需要 filesystem 路径的 runtime（registry / catalog / fixture root），plan 派生模板应自动列出 "新增 config keys + 默认 fallback + 测试 helper（go.mod walk-up）" 的子步骤，避免每个 plan 重复发现这一类工程位。

- **dockertest / pgtestdb harness 升级独立成 B4 plan**
  - **落点**：spec-plan（`db-migrations-baseline` 后续派生 plan）
  - **优先级**：medium
  - **建议**：把 dockertest harness 从 prompt-rubric-registry plan §4.7 的隐含依赖中拆出来，作为 `db-migrations-baseline` 的下一个独立 plan；本 plan 的 static SQL parse 测试明确标注为临时降级，待 dockertest plan 完成后升级为 `//go:build integration`。

- **plan §1.7 类 "全绿" 验证文本与现有 repo state 对齐**
  - **落点**：spec-plan / `/plan-review`
  - **优先级**：low
  - **建议**：当 plan 把顶层 `make lint` / `make migrate-check` 列为强制验证，需要在 plan 派生时先记录当前 baseline 是否已经全绿；非则在 plan §6 风险与应对里说明 "已知 pre-existing 警告 N 项，本 plan 不引入新告警" 的口径，避免 plan 文字与现实长期错位。

## 5 建议优先级与后续动作

- **下一轮最值得实施**：
  - 把 README/doc 注释 vs 负向 grep gate 的字面冲突写成一条 `/plan-review` 检查（high）。当前同类摩擦已在 002 / 003 plan 之外的多个主题（targetjob retire / runtime-topology）出现过，可一次性收口。
  - plan 派生时强制 grep 仓库验证 harness / 测试名（high）。`/design` 与 `/plan-review` 二者各加一个步骤即可；落地后 prompt-rubric-registry 002 / 003、F1 metric 等后续 plan 都能受益。
  - `migrations_lint.py` SQL 解析升级（medium）。在 B4 dockertest harness plan 派生前先把 lint 解析做硬，避免下一次 seed migration / DDL+DML 混合迁移再陷入相同对话。

- **可以延后处理**：
  - conventions parity fixture 自动维护（medium）：当前手编只一处，但下一次再加 vocabulary 字段就会再撞上，建议与 codegen-conventions 升级一并计划。
  - cmd/api / runtime filesystem config-key 子步骤模板化（medium）：等到下一个 filesystem-backed runtime（如 model profile catalog 拆分、observability dashboard truthsource）落地时再合并。
  - dockertest plan 派生（medium）：在 B4 spec 下安排独立 plan，把本次 db_integration_test.go 升级标注作为 acceptance 条件之一。
  - plan §1.7 "全绿" 验证文本对齐（low）：与 minimax_speech / targetJobId revive 警告的清理一起处理；本 plan 已在 §8 evidence 中明确 out-of-scope。

下一步建议进入 `/implement docs/spec/backend-practice/...` 派生第一个依赖 AI 输出的 plan（D-29 解锁后的首题 / 追问 / hint handler），同时把上述 high 优先级建议沉淀到 `/plan-review`。
