# TargetJob Profile Schema Drift 交付复盘报告

> **日期**: 2026-07-08
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `/parse?targetJobId=...` 读取 JD 解析状态时 `GET /api/v1/targets/{targetJobId}` 返回 500 的问题。关联 Bug：[BUG-0142](../bugs/BUG-0142.md)，owner plan：[backend-targetjob/001-targetjob-import-and-parse-bootstrap](../spec/backend-targetjob/plans/001-targetjob-import-and-parse-bootstrap/plan.md)。
- 成功证据：
  - Red 阶段：`DATABASE_URL=<local-dev-postgres-dsn> go test -tags=integration ./internal/targetjob -run TestSQLStoreIntegration_GetTargetJobByUser_AllowsFailedJobWithoutRequirements -count=1` 失败，错误为 `pq: column "profile_id" does not exist`。
  - Focused green：`go test ./internal/targetjob -count=1` 通过。
  - Real DB gate：`DATABASE_URL=<local-dev-postgres-dsn> go test -tags=integration ./internal/targetjob -run TestSQLStoreIntegration_GetTargetJobByUser_AllowsFailedJobWithoutRequirements -count=1` 通过。
  - HTTP scenario regression：`go test ./cmd/api -run 'TestBuildTargetJobRuntimeWiresDrainerAndAIClient|TestBuildAPIHandlerMountsTargetJobRoutesBehindSessionMiddleware|TestE2EP0012HTTPParseFailureRetryableAndNonRetryable' -count=1` 通过。
  - Runtime smoke：`test/scenarios/env-verify.sh` 通过；截图同款 `GET /api/v1/targets/{targetJobId}` 返回 200，body 包含 `analysisStatus="failed"` 和空 `requirements`。
  - 文档与负向 gate：`validate_context.py`、`sync-doc-index --check` 通过；`rg 'profile_id|ProfileID|profileID' backend/internal/targetjob backend/cmd/api openapi/fixtures/TargetJobs openapi/openapi.yaml migrations shared` 无 active 命中。

## 2 会话中的主要阻点/痛点

- sqlmock 测试与 production SQL 一起保留了已退役列。
  - **证据**：修复前 `store_test.go` 的 mock rows 仍包含 `profile_id`，包级 mock 测试无法发现真实 Postgres schema 中该列已不存在。
  - **影响**：历史 completed plan 和 focused unit gate 给出假阴性，真实详情接口在本地运行时仍返回 500。
- 失败态 TargetJob 详情读取缺少真实 DB-backed gate。
  - **证据**：新增 integration red test 后，最小失败态 row 即复现 `pq: column "profile_id" does not exist`；该路径不依赖 AI runner 成功，也不需要 requirements。
  - **影响**：解析任务失败本应是可展示业务状态，却被 store SQL error 包装成 `TARGET_IMPORT_FAILED` 500。
- owner plan 没有把退役 profile schema 列作为长期负向 invariant。
  - **证据**：`backend-targetjob` spec / plan 原先覆盖 import / parse / source / BDD，但没有要求 active SQL 与当前 B4 DDL 做 `profile_id` zero-reference 搜索。
  - **影响**：D-20/D-17 退役 profile 后，TargetJob owner 没有独立防止旧列回流的 gate。

## 3 根因归类

- 根因：TargetJob store 的 active SQL 与当前 `target_jobs` DDL 漂移，mock 列集合同步漂移掩盖了真实 schema 错误。
  - **类别**：spec-plan。
- 根因：plan/checklist 缺少“退役列 zero active reference + real DB-backed failed-state detail read”的组合 gate。
  - **类别**：spec-plan。
- 根因：接口错误 envelope 只暴露通用 `TARGET_IMPORT_FAILED`，定位时需要回查 store/integration 错误链才能看到真实 SQL 根因。
  - **类别**：无需仓库改动。本次已通过 Bug 记录和 plan gate 固化诊断入口，暂无必要改 handler envelope。

## 4 对流程资产的改进建议

- 对后续 schema pruning / module pruning owner plan，默认加入“active SQL scoped negative grep + real DB-backed minimal row read/write gate”。
  - **落点**：spec-plan
  - **优先级**：high
- 对 mock-heavy store 包，凡 SQL 列集合跟随 DDL 删除或新增，都要至少有一条 integration test 证明当前 DDL 下 production store 能读写最小有效 row。
  - **落点**：spec-plan
  - **优先级**：high
- 对业务失败态详情页，测试不应只覆盖 happy-path ready；应覆盖 failed / no child rows / source snapshot present 的可读失败态。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高价值后续动作：对其它仍经历 profile/resume schema pruning 的 backend owner 做一次窄扫，优先检查 active SQL 是否还引用退役 owner 字段，并补真实 DB-backed gate。
- 可延后动作：评估 handler 对底层 SQL schema 错误的内部日志是否足够可观测；如果后续同类 500 定位仍慢，再把 store error classification / logging 提升为平台级改进。
