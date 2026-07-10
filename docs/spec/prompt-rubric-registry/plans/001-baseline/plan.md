# F3 Baseline Registry, Resolve and Lint Gates

> **版本**: 1.10
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本 owner 固化 F3 baseline registry 的当前可执行合同：9 个 baseline `feature_key` 使用 canonical `multi` prompt / rubric / output schema 真理源，`backend/internal/ai/registry` 提供加载、hash 校验、`ResolveActive`、cache、Judge interface 与 registry client，TargetJob 解析链路通过 adapter 消费 registry resolution，AI task run 与 OpenAPI provenance 保持字段闭环。

本 owner 不实现 prompt 编辑 UI、不直接调用 AI provider、不持有 provider/model 字符串、不改变 A3 model profile 状态。真实 provider routing 归 A3，output schema 深化归 F3 `002`，language coordinate 收敛归 F3 `003`，real judge / eval归 F3 `004`。

## 2 当前真理源

| 范围 | 当前来源 | 可执行落点 |
|------|----------|------------|
| Baseline feature keys | `docs/spec/prompt-rubric-registry/spec.md` §3.1.1 | 9 个 feature_key 目录 |
| Prompt meta/body | `config/prompts/README.md`、`config/prompts/<feature_key>/v0.1.0.{yaml,md}` | `scripts/lint/prompt_lint.py`、Go loader |
| Output schema | `config/prompts/<feature_key>/v0.1.0.schema.json` | registry `OutputSchema`、prompt/schema/struct lint |
| Rubric schema | `config/rubrics/README.md`、`config/rubrics/<feature_key>/v0.1.0.yaml` | `scripts/lint/rubric_lint.py`、Go loader |
| Registry runtime | `backend/internal/ai/registry` | `ResolveActive`、`GetPrompt`、`GetRubric`、cache、Judge interface |
| TargetJob bridge | `backend/internal/targetjob/prompt_registry.go` | `RegistryAdapter` |
| Provenance persistence | `shared/conventions.yaml`、B4 migrations、A3 aiclient metadata | `ai_task_runs` typed columns、OpenAPI `GenerationProvenance` |

## 3 质量门禁

- **Plan 类型**: `code-internal` + `contract` + `truth-source` + `migration` + `tooling`
- **TDD 策略**: 通过 `/implement prompt-rubric-registry/001-baseline backend` 进入 `/tdd`。实现项必须由 lint negative fixtures、Go unit tests、migration static gates、fixture validation 或 cross-layer tests 先行约束。
- **BDD 策略**: 不适用。该 owner 只提供内部 Go 包、config truth source、lint scripts、migration seed 与 TargetJob adapter，不新增用户可见 UI、HTTP API 行为或端到端业务流程。
- **替代验证 gate**: `lint-prompts`、`lint-rubrics`、`lint-prompts-hardcode`、`lint-ai-profile-coverage`、registry / targetjob / aiclient Go tests、migration static lint、`make validate-fixtures`、context validator、sync-doc-index 和 product-scope pruning surface lint。

## 4 当前合同

### 4.1 Config truth source

`config/prompts/` 和 `config/rubrics/` 只维护当前 9 个 baseline feature_key。每个 prompt feature_key 有 `v0.1.0.yaml`、`v0.1.0.md`、`v0.1.0.schema.json`；每个 rubric feature_key 有 `v0.1.0.yaml`。Baseline 坐标为 `language: multi`；语言 override 只在 spec/plan 记录真实语义差异时允许。

`template_hash` 算法由 prompt README、Python lint 和 Go loader 共享：`sha256(template_body || canonical_json(meta_without_template_hash))`。hash drift、字段顺序漂移、stub marker、provider/model 字符串和 schema/prompt/struct 字段漂移都必须让 lint 失败。

### 4.2 Registry runtime

`backend/internal/ai/registry.Client` 启动时加载 prompt/rubric/schema truth source，校验 prompt/rubric coverage 和 hash，使用 atomic snapshot + TTL cache。`ResolveActive(featureKey, language)` 精确匹配语言后 fallback 到 `multi`，返回 `PromptResolution`：`FeatureKey`、`PromptVersion`、`RubricVersion`、`ModelProfileName`、`DataSourceVersion`、`FeatureFlag`、system/user template 与可选 `OutputSchema`。

`Judge` interface 作为 registry 包契约导出；real judge implementation 由 F3 `004` 承接，001 只保证 interface、错误类型和默认边界。

### 4.3 TargetJob adapter

`backend/internal/targetjob.RegistryAdapter` 将 F3 registry resolution 映射到 TargetJob parse pipeline 所需的 7 字段，同时保持 `importTargetJob` / `getTargetJob` fixture 与 OpenAPI `GenerationProvenance` 字段一致。TargetJob runtime 不再维护独立 static prompt registry。

### 4.4 Provenance and migrations

B1 shared vocabulary、B4 migrations、A3 aiclient metadata/writer 和 `ai_task_runs` typed columns共同承载 `feature_key`、`prompt_version`、`rubric_version`、`model_profile_name`、`model_id`、`feature_flag`、`data_source_version`。Prompt/rubric baseline seed migration 写入 9 个 canonical `multi` rows，并保持 down SQL 范围限定到 `version='v0.1.0'` 与 9-key 集合。

### 4.5 Lint targets

Makefile 暴露并聚合：

- `make lint-prompts`
- `make lint-rubrics`
- `make lint-prompts-hardcode`
- `make lint-ai-profile-coverage`

这些 gates 共同阻止 prompt/rubric truth source、hardcoded prompt、profile coverage、seed hash 和 schema contract 漂移。

## 5 验收标准

- 9 个 baseline feature_key 均存在 canonical `multi` prompt / rubric / output schema truth source。
- `RegistryClient.ResolveActive` 对 9 个 feature_key 返回完整 resolution，language fallback 到 `multi`，unknown feature/language 走明确错误。
- TargetJob parse pipeline 通过 `RegistryAdapter` 消费 F3 registry，不保留独立 static prompt registry。
- B1/B4/A3 provenance 字段与 `GenerationProvenance`、`ai_task_runs` 和 TargetJob fake writer tests 对齐。
- Prompt / rubric lint、hardcode lint、AI profile coverage、migration static lint、Go focused tests 和 fixture validation 通过。
- BDD 不适用说明和替代验证 gate 保留在 checklist 中。
- `deadcode -test` 与精确 symbol inventory 不再报告未使用的 feature-key helper API。
- Registry、TargetJob 与 cmd/api tests/benchmarks 通过一个 `internal/testsupport.ConfigRoots` 定位 F3 config truth source，不复制 repo walker。

## 6 验证入口

```bash
python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/prompt-rubric-registry/plans/001-baseline/context.yaml --target backend
make lint-prompts
make lint-rubrics
make lint-prompts-hardcode
make lint-ai-profile-coverage
python3 scripts/lint/migrations_lint.py --repo-root .
cd backend && go test ./internal/ai/registry/... ./internal/targetjob/... ./internal/ai/aiclient/... ./internal/shared/ai/... ./internal/shared/types/... -count=1
make validate-fixtures
```

## 7 修订记录

| 日期 | 版本 | 说明 |
|------|------|------|
| 2026-07-10 | 1.10 | Centralize prompt/rubric config-root discovery for tests and benchmarks. |
| 2026-07-10 | 1.9 | Remove the unused `Iterable` import from the prompt linter. |
| 2026-07-10 | 1.8 | Collapse three pytest alias wrappers into directly collected tests. |
| 2026-07-10 | 1.7 | Remove unused feature-key inventory/string helpers and the unsupported future-codegen claim. |
| 2026-07-10 | 1.6 | Simplify rubric score-level loading through the equivalent domain type conversion. |
| 2026-07-10 | 1.5 | Remove the stale fixed F3 spec version from the rubric README and lock a versionless-reference gate. |
| 2026-07-07 | 1.4 | Compress owner docs to the current 9-key multi-coordinate registry, lint, TargetJob adapter and provenance contract. |

## 8 Rubric README stable spec reference

`config/rubrics/README.md` 引用 active F3 spec 时只使用路径，不固定瞬时版本号。以负向搜索锁定 active config/lint 文档中不存在 ``spec.md` vN.N``，并运行 rubric lint、context/index/docs/diff/pruning gates 后确认 `completed`。

验证结果：`make lint-rubrics` 输出 9 files clean，6 个 rubric lint tests 通过；F3 001 context、index/docs/diff/pruning gates 通过，固定版本搜索为零。

## 9 Registry score-level conversion simplification

`scoreLevelYAML` 与 `ScoreLevel` 保持字段名、顺序和类型同构。loader 直接执行显式类型转换，删除逐字段复制的重复映射，同时用全量 registry loader tests、rubric lint 和 scoped `staticcheck` 保持 9 份 rubric 的加载合同不变。

## 10 Feature key helper removal

`backend/internal/shared/featurekeys` 只拥有 registry 与业务 package 实际消费的 typed constants。删除零消费者 `All` inventory、`String` method 与不存在的 future-codegen 说明；当前 coverage 继续由 F3 spec/config lints 和 resolver tests 承担。

## 11 Pytest alias cleanup

`prompt_lint_test.py` 与 `rubric_lint_test.py` 的三个 lint contract 直接使用 pytest 可发现的 `test_*` 名称。删除未被 pytest 收集的 Go-style `Test*` 主体与小写转发 wrapper 双层结构；测试断言、fixture、收集数量和 lint 行为保持不变。

## 12 Shared config-root test support

`backend/internal/testsupport.ConfigRoots` 使用 `testing.TB` 同时服务 registry tests、registry benchmark、TargetJob cross-layer tests 和 cmd/api composition tests。删除各 package 向上查找 `backend/go.mod` 的本地 walker，不保留转发 alias；共享 helper 仍在无法定位 checkout 时 skip，并返回 repo `config/prompts` / `config/rubrics` 绝对路径。
