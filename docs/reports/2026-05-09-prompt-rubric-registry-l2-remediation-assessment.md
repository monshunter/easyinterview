# Prompt Rubric Registry L2 Remediation 交付复盘报告

> **日期**: 2026-05-09
> **审查人**: Codex

**关联计划**: [prompt-rubric-registry/001-baseline](../spec/prompt-rubric-registry/plans/001-baseline/plan.md)
**关联 Bug**: [BUG-0030](../bugs/BUG-0030.md)

## 1 复盘范围与成功证据

本次复盘覆盖 `/plan-code-review prompt-rubric-registry --fix` 后的 Phase 6 remediation：修复 `ParseExecutor` provenance model id 来源、A3 `CallMetadata.FeatureFlag` / B4 `TaskRun` 传递、TargetJobs fixture validator 的 provider-neutral model id contract，以及 4.7 / 4.8 / 5.2 gate evidence 误记。

成功证据：

- `cd backend && go test ./internal/targetjob -run 'TestParseExecutor_(HappyPath|MetadataCarriesF3Triple|AITaskRuns)' -count=1 -race`
- `cd backend && go test ./internal/targetjob -count=1 -race`
- `cd backend && go test ./internal/ai/registry/... -count=1 -race`
- `cd backend && go test ./internal/ai/aiclient/... -count=1 -race`
- `cd backend && go test ./cmd/api -count=1 -race`
- `python3 -m pytest scripts/lint/validate_fixtures_test.py -q`（20 passed, 2384 subtests）
- `make validate-fixtures`（34 fixtures OK）
- `python3 scripts/lint/migrations_lint.py --repo-root .`
- context validator 与 `sync-doc-index --check` 均通过；plan/checklist v1.2 已恢复 `completed`
- `make migrate-check` 仍因本地缺 `DATABASE_URL` 停在 live DB 子步骤；已在 plan/checklist 和 BUG-0030 中记录为 blocker，不再记为绿色 DB gate

## 2 会话中的主要阻点/痛点

- **历史 checklist 记录了 no-op test**
  证据：`go test ./backend/internal/targetjob -run TestParseExecutorAITaskRuns -race` 之前返回 `[no tests to run]`，但 checklist 4.8 记录为全绿。影响是 `ai_task_runs` typed row 没有被 targetjob ParseExecutor 真实路径覆盖。

- **`modelId` 与 `model_profile_name` 层次混淆**
  证据：plan §8.1.2.4 已写明 `modelId` 来自 A3 resolved model id，但 `parse_executor.go` 丢弃 `AICallMeta` 并写入 registry `ModelProfileName`；fixture validator 也只允许 `model-profile:*`。影响是 wire provenance 示例和 runtime 行为都偏向旧口径。

- **live DB gate 与 static gate 混写**
  证据：`db_integration_test.go` 当前是 static SQL parse，不是 build-tagged PG integration；`make migrate-check` 本地输出 `DATABASE_URL is required`。影响是 checklist 4.7 / 5.2 容易把未执行的 live migration gate 误读为已通过。

- **相邻 cmd/api 测试 stub 没有完整 A3 meta**
  证据：新增 `AICallMeta.ModelID` fail-closed 后，`cmd/api` invalid-output case 先变成 `AI_PROVIDER_CONFIG_INVALID`。影响是测试原本要覆盖 invalid AI output，却被测试 stub 的 metadata 缺口抢先命中。

## 3 根因归类

- **spec-plan**：历史计划没有把 test-name existence / no-op detection 固化为 gate，导致不存在的 focused test 被写成 pass。
- **spec-plan**：`GenerationProvenance.modelId` 字段语义虽然写在 handoff 里，但 checklist 3.4/3.5 没要求反向断言 actual A3 meta，fixture validator 也没有同步更新。
- **README / plan evidence**：migration gate 没区分 static lint / static SQL parse / live DB migration 三层。
- **no repo change needed**：cmd/api stub 问题是本次新增 fail-closed 检查后的测试假体补齐，已在同一修复中处理。

## 4 对流程资产的改进建议

- **落点：`/plan-review` skill，优先级 high**
  增加 test-name / harness existence 检查：当 checklist 写 `go test -run TestX` 时，review 必须用 grep 或实际 `go test -run` 确认不是 `[no tests to run]`。

- **落点：spec/plan 模板，优先级 high**
  对 cross-layer provenance 字段要求同时写清 `source field`、`writer field`、`wire field` 和至少一个 end-to-end test owner，避免 `modelId` / `model_profile_name` 这类同名近义层次混淆。

- **落点：migration README / B4 后续 plan，优先级 medium**
  把 migration verification 分成三类：static lint、static SQL parse、live DB down/up。没有 `DATABASE_URL` 时只能记录 blocker，不能作为绿色 gate。

- **落点：fixture README，优先级 medium**
  已在本次把 `model-profile:*` / `fixture-model:*` 的含义写入 `openapi/fixtures/README.md`；后续如果引入更多 provider-neutral fixture id 前缀，需要同步 validator test。

## 5 建议优先级与后续动作

最高价值的下一步是修 `/plan-review` 或 `/plan-code-review` 的 checklist 验证规则：任何 `go test -run` gate 必须确认执行到测试，不能接受 `[no tests to run]`。其次，建议为 `db-migrations-baseline` 派生 live PG harness plan，把本次 static SQL parse gate 升级成真正的 migration down/up 验证。
