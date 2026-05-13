# Backend Practice 002 Event Loop Replay Contracts 交付复盘报告

> **日期**: 2026-05-14
> **审查人**: Codex

## 1 复盘范围与成功证据

- 修复范围：`backend-practice/002-event-loop-and-completion` L2 follow-up，覆盖 `answer_submitted` server-owned follow-up state、`occurredAt` / `clientCompletedAt` required validation、D-35 complete replay lookup，以及 P0.038 / P0.039 / P0.040 HTTP scenario 语义加固。
- 代码证据：`SessionEventService` 改为读取 `practice_turns.follow_up_count`，忽略客户端 `payload.followUpCount`；append store 写入 outcome 计算后的 follow-up count；API handler 对缺失 timestamp 返回 422；complete replay SQL 绑定 `job_type` / `resource_type` / `dedupe_key`。
- 计划资产：原 002 plan / checklist / test-plan / test-checklist / bdd-plan / bdd-checklist 已原地 bump 到 v1.2，并补齐本次修复的 operation matrix、BDD 和测试 gate。
- Bug 资产：[BUG-0055](../bugs/BUG-0055.md) 记录 client follow-up state trust、timestamp required drift 与 D-35 replay lookup drift，状态 `resolved`。
- 验证证据：`cd backend && go test ./...`、focused backend practice/API/store/cmd tests、`python3 -m pytest scripts/lint/backend_practice_legacy_test.py -q`、`python3 scripts/lint/backend_practice_legacy.py --repo-root .`、`python3 scripts/lint/conventions_drift.py --repo-root .`、`make lint-events`、`make validate-fixtures`、`make docs-check`、`make codegen-events-check`、`make codegen-check`、`git diff --check` 均通过。

## 2 会话中的主要阻点/痛点

- 历史 P0.038 只验证了 action sequence，没有验证 follow-up count 的真实 owner。
  - **证据**：修复前客户端提交 `payload.followUpCount: 99` 可绕过首轮 follow-up 分支，直接进入 `ask_question`。
  - **影响**：用户第一轮回答会跳过追问，且 turn completed outbox 可能提前发出。
- OpenAPI required 字段与 handler 容错实现存在漂移。
  - **证据**：`occurredAt` / `clientCompletedAt` 在 contract 中 required，但 handler 允许空值并写入 zero time。
  - **影响**：invalid replay snapshot 和 audit 时间会被当成有效业务事件。
- D-35 replay 查找条件不够完整。
  - **证据**：complete replay SQL 只按 `session_id` 和 job type join，未绑定 `async_jobs.resource_type` / `dedupe_key`。
  - **影响**：同一 resource id 或异常 job 行可能污染 idempotency replay 语义。

## 3 根因归类

- Server-owned state 未被显式写进 scenario redline。
  - **类别**：spec-plan
  - **根因**：计划写明了 follow-up budget 和 action 分支，但没有把“客户端 payload 不可信，DB turn state 是唯一 owner”作为独立 negative gate。
- Required timestamp validation 过度依赖 contract 文档。
  - **类别**：spec-plan
  - **根因**：OpenAPI required 字段没有同步落到 handler negative tests 与 HTTP scenario coverage。
- Replay lookup invariant 分散在实现细节中。
  - **类别**：spec-plan
  - **根因**：D-35 只强调 replay 结果一致，没有枚举 lookup 必须同时绑定 `job_type`、`resource_type`、`dedupe_key`。

## 4 对流程资产的改进建议

- 后续 event-loop 类 plan 应固定 server-owned state negative gate。
  - **落点**：spec-plan
  - **优先级**：high
  - 建议：凡状态机字段可由客户端 payload 伪造时，BDD 必须提交恶意 payload 并断言服务端仍以 DB / store state 为准。
- API required 字段应在 handler 层形成红线测试。
  - **落点**：spec-plan
  - **优先级**：high
  - 建议：operation matrix 中的 required request field 不只列 contract，也要列 handler negative test 和 HTTP scenario assertion。
- Complete / replay 类 gate 应写明 lookup invariant。
  - **落点**：spec-plan
  - **优先级**：medium
  - 建议：D-35 类条目直接枚举 replay 查询条件，避免只验证 happy-path replay body。

## 5 建议优先级与后续动作

- 最高优先级：进入 `backend-practice/003-mode-policies-and-provenance` 前，先用 `/plan-review --fix` 检查 mode policy 中是否存在同类 client-owned payload 风险，并把 server-owned state redline 写入 BDD。
- 次优先级：将 required request field negative gate 作为 backend API plan 的默认 checklist 项，尤其是 event append、completion、async handoff 这类会写 audit / outbox 的接口。
