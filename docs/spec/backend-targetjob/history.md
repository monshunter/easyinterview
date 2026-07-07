# Backend TargetJob History

> **版本**: 1.8
> **状态**: active
> **更新日期**: 2026-07-07

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-07 | 1.8 | 同步当前 backend internal runner 事实：`target_import` / `source_refresh` 已由 `backend-async-runner` kernel 接管，backend-targetjob 仅保留 handler / service / store / executor 业务实现与 B3 payload red-line。 | product-scope/001-core-loop-module-pruning Phase 6 |
| 2026-07-06 | 1.7 | 对齐 product-scope D-17/D-18 后的 TargetJob active owner 边界：背景收敛为 JD import / parse；不新增 recommendation/search/data-source plan；`CountTargetJobsForUser` 不作为当前 cross-owner 正向边界。 | product-scope/001-core-loop-module-pruning Phase 6 |
| 2026-06-29 | 1.6 | product-scope D-22 后同步 downstream 边界：backend-targetjob 只声明当前 TargetJob / practice / report downstream owner。 | product-scope/001-core-loop-module-pruning |
| 2026-05-21 | 1.5 | 登记 backend-jobs-recommendations/001 cross-owner additive：新增 `CountTargetJobsForUser(ctx, db, userID) (int, error)` 内部 API（`backend/internal/targetjob/count.go`），read-only `SELECT COUNT(*) FROM target_jobs WHERE user_id = $1 AND deleted_at IS NULL`；cross-user 隔离由 caller userId 保证；不写 audit_events。单元测试 `count_test.go` 覆盖 happy / cross-user / nil-db / empty-userId。 | backend-jobs-recommendations/001-jd-match-real-backend-baseline Phase 0.13 |
| 2026-05-08 | 1.4 | 完成 001 plan 真实 HTTP BDD gate：p0-010..013 场景脚本迁移为 `cmd/api` HTTP harness，覆盖 auth middleware、generated route、TargetJob handler/service、in-process drainer、F3 contract bridge、A3 test fixture 与 URL fetch，verify 输出 `method=cmd-api-http` / `validBddEvidence=true`。 | 001-targetjob-import-and-parse-bootstrap |
| 2026-05-08 | 1.3 | L2 plan-code-review remediation：重新打开 001 plan 与 BDD gate，记录 `cmd/api` 缺真实 TargetJob drainer / F3 runtime wiring 的 blocker，并补充 URL fetch DNS rebinding 与 update 状态机事务内校验修复项。 | 001-targetjob-import-and-parse-bootstrap |
| 2026-05-08 | 1.2 | 完成 001 plan 交付：新增 E2E.P0.010 / 011 / 012 / 013 场景资产与脚本证据，补齐 generated TargetJob summary / fitSummary provenance 映射，并修复 `AI_PROVIDER_SECRET_MISSING` 合法 B1 错误码被 payload redline 误杀的问题。 | 001-targetjob-import-and-parse-bootstrap |
| 2026-05-08 | 1.1 | L1 plan-review remediation：按具体场景补齐 B1/B2/B3/F1 owner 契约前置、manual_form terminal job 语义、B3 sourceType 映射、TargetJob 错误码、F1 指标名与 BDD.P0.013 manual_form 场景。 | 001-targetjob-import-and-parse-bootstrap |
| 2026-05-08 | 1.0 | 初始创建：固定 4 个 TargetJob operation 的 backend owner 边界、4 类导入源处理、target_import 异步解析管线、隐私/观测红线、cross-user 隔离、URL fetch SSRF 守护、F3/A3 fail-closed、manual_form 同步路径与 idempotency dedupe；派生 001-targetjob-import-and-parse-bootstrap plan，BDD 占用 E2E.P0.010 / E2E.P0.011 / E2E.P0.012 三个场景 ID（接续 practice-voice-mvp 已预留的 P0.007-P0.009） | 001-targetjob-import-and-parse-bootstrap |
