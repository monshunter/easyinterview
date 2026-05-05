# Codegen Pipeline Continuation

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-04-30

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [shared-conventions-codified spec](../../spec.md) v1.6 升格的 `002-codegen-pipeline` 范围落到可实施设计：在 [001-bootstrap](../001-bootstrap/plan.md) 已完成的 `shared/conventions.yaml`、generator、Go/TS shared lib 与本地 lint gate 之上，补齐 A3 触发的 AI shared vocabulary、跨语言 drift/parity 检测与 `make codegen-check` 本地接入。

本 plan 当前只处理 AI vocabulary 与本地 drift/parity，不实施 F3 prompt registry bridge，不接入远端 CI workflow，不修改 A5 scope。F3 `feature_key + version` 共享 SDK、remote CI drift detection 只作为 future handoff；触发时新增后续 plan 或修订本 spec，不回填到 002。

## 2 背景

A3 `ai-provider-and-model-routing` 需要稳定字段名来对齐 `AICallMeta` runtime、B4 `ai_task_runs` typed columns、F1 metrics/logs 与 B2 `GenerationProvenance`。这些字段名必须由 B1 提供共享 vocabulary，但 B1 不拥有 `AIClient`、Model Profile schema、provider adapter 或连接参数校验。

001-bootstrap 已经完成错误码、枚举、ID、pagination 与 `ApiError`/`PageInfo` 等共享基础。本 plan 只追加 AI vocabulary 与更强 drift/parity 检测，不替换 001 的 generator 入口，不重命名既有 shared lib 路径。

## 3 质量门禁分类

- **Plan 类型**: `tooling + code-internal + contract`。本 plan 修改 `shared/conventions.yaml`、B1 codegen、Go/TS generated shared lib、lint/drift wrapper 与 parity tests，属于内部契约和工具链交付；不引入用户可感知 UI、HTTP API 行为、业务流程或端到端功能。
- **TDD 策略**: 必须通过 `/tdd --file docs/spec/shared-conventions-codified/plans/002-codegen-pipeline/checklist.md --references docs/spec/shared-conventions-codified/plans/002-codegen-pipeline/plan.md,docs/spec/shared-conventions-codified/spec.md --phase-commit shared-conventions-codified/002-codegen-pipeline` 顺序执行。每个 checklist item 以本 checklist 内的 `验证:` 子句作为 Red-Green-Refactor 断言来源；涉及 generated output 的 item 必须先在 generator 或 drift/parity test 中制造失败，再最小实现并复跑 focused command。
- **BDD 策略**: BDD 不适用。本 plan 不产生浏览器 UI、外部 API、业务工作流或场景测试可观察行为，因此不创建 `bdd-plan.md` / `bdd-checklist.md`，主 checklist 也不设置 `BDD-Gate:`。
- **替代验证 gate**: 使用内部契约 gate 代替 BDD：`make codegen-conventions`、`make codegen-check`、Go generator/shared package tests、TS conventions/ids tests、TS typecheck、AI vocabulary negative drift cases、`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`。

## 4 实施步骤

### Phase 1: AI shared vocabulary 真理源

#### 1.1 `aiVocabulary` 命名空间

在 `shared/conventions.yaml` 增加 `aiVocabulary` 命名空间，至少包含字段名常量：`model_profile_name` / `model_profile_version` / `model_family` / `model_id` / `fallback_chain` / `route` / `validation_status` / `output_schema_version` / `prompt_version` / `rubric_version` / `language` / `feature_flag` / `data_source_version`。snake_case 是 wire 真理源，Go / TS generator 分别输出 idiomatic 常量。

#### 1.2 AI vocabulary 输出落点

generator 输出 Go 到 `backend/internal/shared/ai/`（或等价 B1-owned AI vocabulary 包），TS 到 `frontend/src/lib/conventions/ai.ts`。不得把 AI meta 字段名放进 `backend/internal/shared/errors/` 或 `frontend/src/lib/conventions/errors.ts`；错误码仍由 errors 包拥有，字段名由 AI vocabulary 包拥有。

#### 1.3 A3 / A4 / F1 / B4 handoff 注释

生成文件注释必须声明：B1 owns 字段名 / 校验 helper；A3 owns `AIClient`、Model Profile schema、`AICallMeta` runtime 填充与 provider adapter；A4 owns `AI_PROVIDER_*` 连接参数校验；B4 owns DB columns；F1 owns metric/log consumption。

### Phase 2: Cross-language drift 检测增强

#### 2.1 三方 drift wrapper

扩展 `backend/cmd/codegen/conventions/` 或新增 `scripts/lint/conventions_drift.py` wrapper，识别 YAML、Go 输出、TS 输出三方差异，覆盖 enum、错误码、AI vocabulary 三类资产。

#### 2.2 `make codegen-check` 接入

把 Phase 2.1 接入 `make codegen-check`。针对「YAML 改但只生成一侧」「YAML 未改但代码私自新增」两种场景明确报错；diff 路径必须包含 `backend/internal/shared/ai` 与 `frontend/src/lib/conventions/ai.ts`。

#### 2.3 不回退 001 generator

新增 drift 检测只追加，不替换 001 已交付的 `make codegen-conventions` 入口；任何模板重排必须保持既有 generated 文件 idempotent。

### Phase 3: AI vocabulary parity tests

#### 3.1 Go / TS parity

落地 Go / TS parity tests，断言 AI vocabulary 字段集合、wire snake_case name、Go 常量名、TS 常量名一一对应。

#### 3.2 A3 字段覆盖

parity 必须覆盖 A3 当前消费字段：`model_profile_name`、`model_profile_version`、`model_family`、`fallback_chain`、`route`、`validation_status`、`output_schema_version`，并保留 `prompt_version` / `rubric_version` / `language` / `feature_flag` / `data_source_version` 供 B2/F3/F1 消费。

### Phase 4: Cross-language contract test

#### 4.1 shared fixture

落地 `shared/fixtures/conventions-parity.json` 或 generator 临时 fixture，断言 14 个枚举类型、错误码常量集合（含 A3 `AI_*` baseline）、AI vocabulary 字段集合在 Go / TS 两侧严格等价。

#### 4.2 serialization parity

断言 `PageInfo` / `ApiError` JSON 序列化经 canonical round-trip 等价，避免 `camelCase` JSON tag 漂移。

### Phase 5: Future handoff only

#### 5.1 F3 prompt bridge future scope

记录 F3 prompt registry bridge 为 future scope：`feature_key + version` SDK 需 F3 spec 先锁定后新增 plan；本 plan不存储 prompt body，不实现 `RegistryClient.GetPrompt`。

#### 5.2 Remote CI future scope

记录 remote CI drift detection 为 A5 future scope：只有 [A5 spec D-5](../../../ci-pipeline-baseline/spec.md#31-已锁定决策) 命中后由 future `002-remote-ci` 接入；本 plan 不创建 workflow 文件。

### Phase 6: Verification

#### 6.1 generator / lint / test

复跑 `make codegen-conventions`、`make codegen-check`、Go shared package tests、TS typecheck / conventions tests，证明扩展未回退 001 既有验收。

#### 6.2 negative drift cases

临时制造 YAML-only / Go-only / TS-only AI vocabulary drift，确认 `make codegen-check` 失败且错误信息指出缺失方向；revert 后恢复 clean。

#### 6.3 文档与 INDEX

本 plan checklist 全部勾选后，将 plan / checklist Header 切 completed，运行 sync-doc-index check/fix，更新 work journal；不修改 A5 workflow、不修改 F3 prompt registry。

## 5 验收标准

- AI vocabulary 字段名已生成到独立 Go / TS 文件，A3 可引用字段名常量但仍 owns runtime `AICallMeta`。
- `make codegen-check` 能覆盖 enum、错误码、AI vocabulary 三类 drift。
- Go / TS parity tests 覆盖 A3 当前消费字段与既有 `ApiError` / `PageInfo` serialization。
- F3 bridge / remote CI 未被本 plan 提前实现，仅留下 handoff。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| AI meta 字段名被塞进 errors 包，混淆错误码与字段名职责 | Phase 1.2 固定独立输出落点；code review 与 parity test 检查 `errors/*` 不包含 AI meta field constants |
| A3 在 B1 生成前私造跨语言字段常量 | Phase 1.3 handoff 明确 A3 只 owns runtime；B1 002 完成后 A3 切换引用生成常量 |
| Drift wrapper 误伤 001 已完成的 generator 输出 | Phase 2.3 要求只追加检测，不替换入口；Phase 6 复跑 001 原有验收 |
| F3 / remote CI scope 被提前塞入本 plan | Phase 5 明确为 future handoff；触发时新增后续 plan或 A5 `002-remote-ci` |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-04-30 | 1.1 | 补齐 TDD/BDD 质量门禁分类与 checklist 可执行验证断言；确认 BDD 不适用并以内部契约 gate 替代。 | implement gate remediation |
| 2026-04-29 | 1.0 | 升格为 active，并将 scope 收敛为 A3 AI vocabulary、drift/parity 与本地 codegen-check 接入；F3 bridge / remote CI 仅保留 future handoff。 | plan-review remediation |
