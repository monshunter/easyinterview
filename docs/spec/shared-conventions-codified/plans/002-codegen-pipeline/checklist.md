# Codegen Pipeline Continuation Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-04-29

**关联计划**: [plan](./plan.md)

## Phase 1: AI shared vocabulary 真理源

- [ ] 1.1 在 `shared/conventions.yaml` 增加 `aiVocabulary` 命名空间，至少包含 `model_profile_name` / `model_profile_version` / `model_family` / `model_id` / `fallback_chain` / `route` / `validation_status` / `output_schema_version` / `prompt_version` / `rubric_version` / `language` / `feature_flag` / `data_source_version`
- [ ] 1.2 generator 输出 Go AI vocabulary 到 `backend/internal/shared/ai/`（或等价 B1-owned 包），TS 输出到 `frontend/src/lib/conventions/ai.ts`；不得把 AI meta 字段名放进 `errors/*`
- [ ] 1.3 生成文件注释明确 B1 owns 字段名 / 校验 helper，A3 owns `AIClient` / `AICallMeta` runtime，A4 owns `AI_GATEWAY_*` 连接参数校验，B4/F1 消费字段名

## Phase 2: Cross-language drift 检测增强

- [ ] 2.1 扩展 `backend/cmd/codegen/conventions/` 或新增 `scripts/lint/conventions_drift.py` wrapper：识别 YAML / Go / TS 三方差异，覆盖 enum、错误码、AI vocabulary
- [ ] 2.2 把扩展接入 `make codegen-check`，针对「YAML 改单侧生成」「YAML 未改但代码私自新增」两种场景明确报错；diff 路径包含 `backend/internal/shared/ai` 与 `frontend/src/lib/conventions/ai.ts`
- [ ] 2.3 不回退或替换 001 已落地的 generator 入口；新增逻辑只追加不替换

## Phase 3: AI vocabulary parity tests

- [ ] 3.1 落地 Go / TS parity tests，断言 AI vocabulary 字段集合、wire snake_case name、Go 常量名、TS 常量名一一对应
- [ ] 3.2 parity 覆盖 A3 当前消费字段：`model_profile_name` / `model_profile_version` / `model_family` / `fallback_chain` / `route` / `validation_status` / `output_schema_version`

## Phase 4: Cross-language contract test

- [ ] 4.1 落地 shared parity fixture 或 generator 临时 fixture，断言 14 个枚举类型字面量集合 + 错误码常量集合（含 `AI_*` baseline）+ AI vocabulary 字段集合两侧严格等价
- [ ] 4.2 断言 `PageInfo` / `ApiError` JSON 序列化经 canonical round-trip 等价，避免 `camelCase` JSON tag 漂移

## Phase 5: Future handoff only

- [ ] 5.1 记录 F3 prompt registry bridge 为 future scope：`feature_key + version` SDK 需 F3 spec 先锁定后新增 plan，本 plan 不存储 prompt body、不实现 `RegistryClient.GetPrompt`
- [ ] 5.2 记录 remote CI drift detection 为 A5 future scope：只有 [A5 spec D-5](../../../ci-pipeline-baseline/spec.md#31-已锁定决策) 命中后由 future `002-remote-ci` 接入，本 plan 不创建 workflow

## Phase 6: Verification

- [ ] 6.1 复跑 `make codegen-conventions` / `make codegen-check` / Go shared package tests / TS typecheck 或 conventions tests，证明扩展未回退 001 既有验收
- [ ] 6.2 临时制造 YAML-only / Go-only / TS-only AI vocabulary drift，确认 `make codegen-check` 失败且错误信息指出缺失方向；revert 后恢复 clean
- [ ] 6.3 本 plan checklist 全部勾选后，将 plan / checklist Header 切 completed，运行 sync-doc-index check/fix，更新 work journal；不修改 A5 workflow、不修改 F3 prompt registry
