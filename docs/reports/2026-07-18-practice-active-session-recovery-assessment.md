# Practice Active Session Recovery 交付复盘报告

> **日期**: 2026-07-18
> **审查人**: Codex

**关联计划**: [Backend Practice Plan And Session Orchestration](../spec/backend-practice/plans/001-plan-and-session-orchestration/plan.md)
**关联 Bug**: [BUG-0189](../bugs/BUG-0189.md)

## 1 复盘范围与成功证据

- 本次交付在原 `backend-practice/001-plan-and-session-orchestration` owner 上补齐 Phase 9：同 user/plan 已有 queued/running session 时，fresh-key repeated start 恢复原 session，不创建第二个活动 session，也不重复生成 opening。
- Service focused tests 覆盖 running recovery、queued 等待与 context cancel；真实 PostgreSQL integration 覆盖 different-key 并发、same-key replay/pending/mismatch、cross-user plan 隔离和零副作用，输出 `active-session-start-recovery=PASS`。
- 根 `make test` 通过 Python 584 tests / 4583 subtests、Go 全包及 frontend 127 files / 1035 tests；`make build`、OpenAPI diff、fixture、docs 与索引 gate 通过。
- 后端重部署并通过环境 verify 后，Chrome 从 Workspace 正式入口进入原 running session；PostgreSQL 前后保持 session/message/event/outbox/audit/AI task 数量为 `1/1/1/1/1/0`，仅新增一条指向原 session 的 succeeded idempotency record。

## 2 会话中的主要阻点/痛点

- 浏览器窗口生命周期与服务端 session lifecycle 被错误地直觉关联。
  - **证据**：用户关闭未回答的面试窗口后，三个近期 plan 的 session 仍分别处于 `running`；前端没有发送 cancel/complete 命令。
  - **影响**：最初看起来像账号级全局故障，实际是多个 plan 各自命中了相同的活动会话冲突。
- 既有启动契约只建模 same-key replay 与 fresh create，没有建模 fresh-key recover。
  - **证据**：活动会话 partial unique index 正确阻止第二条 queued/running row，但 Store 将该约束失败统一映射为 `PRACTICE_SESSION_CONFLICT`；原 test matrix 没有 fresh-key repeated-start 分支。
  - **影响**：合法的已有资源被暴露为不可恢复冲突；直接放宽唯一索引又会引入重复会话和 opening 副作用风险。
- 恢复语义同时涉及幂等、业务 identity 与并发边界，不能只在 handler 层改错误映射。
  - **证据**：不同 start keys 可以并发越过普通存在性检查；queued session 可能仍在原请求的 opening 阶段，恢复请求不能提前固化不完整 snapshot。
  - **影响**：修复需要 user/plan-scoped transaction lock、queued 等待和独立 recovery finalization，而不是单点 catch unique violation。
- 运行时核验的第一版临时 SQL 使用了逻辑名称而非真实 schema 名称。
  - **证据**：`practice_outbox` / `operation_id` 查询失败；迁移和 integration helper 的真实名称是 `outbox_events` / `operation`。
  - **影响**：增加了一次只读核验重试，但没有影响实现或数据；改用当前 migration/test helper 后完成前后对比。

## 3 根因归类

- Fresh-key repeated start 缺少 recover 分支属于 **spec-plan**：原 operation matrix 和 concurrency matrix 未定义“已有合法活动资源”是恢复还是冲突。
- 浏览器关闭不改变服务端 session 属于 **spec-plan**：这是业务 lifecycle 事实，应由 active owner 明确，不应依赖前端窗口行为推断。
- 临时验收 SQL 的表/列误写属于 **无需仓库改动**：真实 schema 和 integration helper 已有明确 owner，本次按 owner 修正后即闭环，不足以证明流程资产存在系统性缺口。

## 4 对流程资产的改进建议

- 保留本次已写入 backend-practice spec/plan 的 create/recover/replay 三分支，以及 user/plan-scoped concurrency 和零副作用 gate。
  - **落点**：spec-plan
  - **优先级**：high（已在本次完成）
- 若产品后续需要“关闭窗口即放弃本轮”，先单独定义显式 abandon/cancel command、终态、审计和恢复 UX；不要把浏览器 disconnect 当成隐式业务命令。
  - **落点**：spec-plan
  - **优先级**：medium
- 当场景测试具备可稳定 seed 既有 running session 的真实 API/UI 资产时，可把 repeated start 恢复固化为独立回归场景；在此之前保留真实 PostgreSQL integration 为 owner gate、Chrome 为运行时补充证据。
  - **落点**：spec-plan / scenario owner
  - **优先级**：low

## 5 建议优先级与后续动作

- 当前最高价值项已经完成：原 owner 文档、service/store 实现、真实 PostgreSQL 并发测试和 Chrome 验收对同一恢复合同一致，不需要再修改 AGENTS.md 或 Skill。
- 下一步应对 Phase 9 变更执行一次 plan-grounded code review，重点复查 plan lock、queued wait cancellation 和 idempotency snapshot 原子性；通过后再合并 feature branch。
- 显式 abandon/cancel 属于新的产品 lifecycle 决策，不是 BUG-0189 的必要补丁；只有用户确认需要“关闭后自动结束/可手动放弃”时再进入 `/change-intake` 设计。
