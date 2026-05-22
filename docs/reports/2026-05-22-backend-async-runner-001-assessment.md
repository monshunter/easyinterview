# Backend Async Runner 001 交付复盘

> **日期**: 2026-05-22
> **审查人**: Claude

## 1 复盘范围与成功证据

范围：`backend-async-runner/001-internal-job-outbox-runner` 全量实施（Phase 1-4），把散落的 `targetjob.Drainer` / `review.Runner+Reaper` / `auth.BackgroundMailDispatcher` / `jdmatchRuntime.Drainer` 收敛为单一 `runner.Runtime` kernel，并落地 outbox dispatcher 与 `email_dispatch` 收口。

成功证据：

- 四个 phase 提交：`fd17247` kernel 基础设施、`de1459f` handler 迁移、`bea7eb0` outbox+email、`24e9cd0` 收口。
- checklist 1.1-4.16 与 test-checklist Phase 1-4 全部勾选；plan / checklist / test-plan / test-checklist 状态 `completed`，`sync-doc-index --check` 零漂移。
- 单测全绿：`go test ./...` 无失败；真 PG 集成（lease/finalize/reclaim、outbox claim/batch/dead-letter/redaction/idempotency、auth email end-to-end magic link ≤1 scan 周期）通过。
- BDD-Gate：p0-003 / 010 / 012 / 035 / 062 / 077 / 078 / 093 / 094 / 097 共 10 个 owner 场景 `trigger.sh`+`verify.sh` 全 PASS；report_generate 经 `reports_http_scenario_test` 验证 kernel `GenerateHandler` 路径。
- Drift gate：`go build` / `go vet` / `validate_context.py` / `make lint-runner-legacy` / `git diff --check` 全部 PASS；10 个 owner spec 负向 grep 0 命中。

## 2 会话中的主要阻点/痛点

1. **retryable finalize SQL 的 PG 类型推断 bug**：`completed_at = case when ... then $1 else null end` 因参数 $1 在多处使用导致 PG 推断为 text，集成测试报 `column "completed_at" is of type timestamptz but expression is of type text`。由真 PG 集成测试在 Phase 1 当场捕获，加 `$1::timestamptz` 显式 cast 解决。
2. **结构负向测试误伤文档注释**：`TestNoBackgroundDispatcher` 扫描整文件文本，flag 了 `mail.go` / `email_dispatch_handler.go` / `targetjob/drainer.go` 中提及旧类名的注释，需要改写注释（而非代码）才通过。
3. **email_dispatch handler 测试路径在 plan 内不一致**：test-plan §2.6 / O-4 把 handler 集成测试落点写为 `backend/internal/runner/email_dispatch_integration_test.go`，但 checklist 3.8 把 handler 落点定为 `backend/internal/auth/email_dispatch_handler.go`。实施按 checklist（auth 包），测试随之放在 auth，与 test-plan 文字路径不符。
4. **集成测试固定 UUID 复跑冲突**：`TestAuthEmailEndToEnd` 用固定 challengeID，仅在 `t.Cleanup` 删除；连续复跑时上一轮残留触发 `auth_challenges_pkey` 重复键失败，补充测试开头 pre-delete 解决。
5. **upload_roundtrip 集成 smoke 受 minio 本地 env 限制**：privacy_delete 的 cmd/api 全链路 smoke（DELETE /me）因本地 minio test harness env（endpoint 读取）失败于 minio bucket 初始化（line 48，早于本次迁移改动），由 privacy/runner handler 单测 + kernel PG 集成 + BDD p0-093/097 复核覆盖。

## 3 根因归类

- 痛点 1（SQL 类型 cast）：症状层的一次性实现问题，由集成测试正确拦截，**no repo change needed**；反而印证 spec「真 PG integration 覆盖 lease/finalize」gate 的价值。
- 痛点 2（结构测试误伤注释）：根因属 **skill / 实施规约**——zero-reference 结构测试与 lint 应明确「标识符引用」与「文档注释」的边界，或production 注释禁止保留 retired 类名。本次以「production 注释不留旧名」收敛。
- 痛点 3（plan 内测试路径不一致）：根因属 **spec/plan**——test-plan 与 checklist 对同一 artifact 落点表述分叉，属 L1 文档一致性缺口。
- 痛点 4（固定 ID 复跑冲突）：根因属 **skill / 测试规约**——DB 集成测试用固定主键时应「先删后插」，否则复跑不幂等。
- 痛点 5（minio env smoke 受限）：根因属 **环境/README**——pre-existing 环境约束，非本次迁移引入；属已知 harness 局限。

## 4 对流程资产的改进建议

| 建议 | 目标资产 | 说明 |
|------|----------|------|
| zero-reference 结构测试 / lint 应区分标识符与注释，或在规约中明确「production 注释不得保留 retired 类名」 | skill（`/plan-code-review` 或 lint 约定）/ AGENTS.md | 避免负向 gate 把历史注释误判为回流，减少反复改写注释 |
| DB 集成测试使用固定主键时必须「先删后插」保证复跑幂等 | README（`test/scenarios/` 或 backend 测试约定） | 把固定-ID pre-delete 写入测试编写约定，避免连续复跑残留冲突 |
| plan 内 test-plan 与 checklist 对同一 artifact 的落点路径必须一致 | spec/plan（`/plan-review` gate） | L1 审查应交叉校验 test-plan 测试入口路径与 checklist 实现/测试落点一致 |
| privacy_delete cmd/api 全链路 smoke 的 minio env 入口需在 scenario/README 固化或提供 db-only 替代 | README（`deploy/dev-stack` / `test/scenarios`） | 让 env 受限场景有明确的替代验证路径，避免误判为回归 |

## 5 建议优先级与后续动作

1. **高**：在 `/plan-review` 增加 test-plan↔checklist artifact 路径交叉一致性检查（痛点 3 根因），成本低、可复用到所有 plan。
2. **中**：把「DB 集成测试固定主键先删后插」「production 注释不留 retired 类名」沉淀到 backend 测试约定 / lint 规约（痛点 2、4）。
3. **低**：为 minio-依赖 smoke 提供 db-only 替代或在 README 固化 env 入口（痛点 5），属环境工程优化。

无需创建 BUG 记录：本次无运行期缺陷修复，痛点 1 为实施期由集成测试拦截的一次性问题。
