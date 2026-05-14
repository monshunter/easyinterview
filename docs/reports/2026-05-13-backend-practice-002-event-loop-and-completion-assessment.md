# Backend Practice 002 Event Loop And Completion 交付复盘报告

> **日期**: 2026-05-13
> **审查人**: Codex

## 1 复盘范围与成功证据

- 交付范围：`backend-practice/plans/002-event-loop-and-completion` Phase 2-4，包括 `appendSessionEvent` event-loop、`completePracticeSession` queued report/job handoff、practice turn/session outbox、双轨 idempotency、D-32/D-33/D-34/D-35 收口与计划生命周期关闭。
- 代码证据：新增 `append_session_event_service.go` / `complete_session_service.go` / store append+complete repository / practice HTTP handlers / idempotency resource header 支持 / `cmd/api` route wiring。
- BDD 证据：`backend/cmd/api/practice_http_scenario_test.go` 覆盖 `TestE2EP0038...` / `0039...` / `0040...` / `0041...` / `0042...` / `0043...` 六个场景。
- 验证证据：`cd backend && go test ./...`、focused backend practice/API/store/middleware/cmd tests、`make lint-events`、`make codegen-events-check`、`make codegen-check`、`make validate-fixtures`、`python3 scripts/lint/conventions_drift.py --repo-root .`、`python3 scripts/lint/backend_practice_legacy.py --repo-root .`、`make docs-check`、`git diff --check` 均通过。

## 2 会话中的主要阻点/痛点

- BDD asset 形态与实际验证入口不一致。
  - **证据**：原 `bdd-checklist.md` 要求创建 `test/scenarios/e2e/p0-NNN-*` shell 目录，但本次可执行验证实际落在 `cmd/api` Go HTTP scenario tests。
  - **影响**：如果直接勾选原 checklist 会制造虚假完成；本次已把 BDD plan/checklist 修订为 Go HTTP scenario evidence，并保留未来 Kind/live scenario handoff。
- Complete handler 与 idempotency middleware 的 response resource 传递需要额外机制。
  - **证据**：`completePracticeSession` 需要让 middleware 在 `MarkSucceeded` 时记录 `feedback_report/{reportId}`，最终通过 `idempotency.SetResponseResource` 与内部 header 清理完成。
  - **影响**：计划只写了 snapshot 语义，没有把 handler-to-middleware resource 传递机制作为独立 gate，导致实现时需要补齐 middleware API。
- 并发语义不能只写成 seq_no 序列化。
  - **证据**：[BUG-0053](../bugs/BUG-0053.md) 记录了不同 `clientEventId` 指向同一旧 turn 时必须返回 stale-turn conflict 的问题。
  - **影响**：仅验证 `SELECT FOR UPDATE` / seq_no 不足以证明用户视角正确；需要把 current-turn 游标校验纳入 BDD 与 service 单元测试。

## 3 根因归类

- BDD asset 形态不一致。
  - **类别**：spec-plan
  - **根因**：计划模板把 BDD 默认写成 `test/scenarios` shell 资产，但当前 backend package 已有更快、更贴近 `cmd/api` route/middleware 的 Go HTTP scenario 入口。
- Idempotency resource 传递机制遗漏。
  - **类别**：spec-plan / README
  - **根因**：idempotency middleware 的 reserve/replay/snapshot 语义已有，但缺少“handler 产生 resource id 后如何反馈给 middleware”的包级约定。
- stale-turn conflict 漏断言。
  - **类别**：spec-plan
  - **根因**：并发 gate 过度聚焦 DB 序号唯一性，没有把客户端 stale view 作为同等级 failure path。

## 4 对流程资产的改进建议

- 在 backend-practice 后续 plan 中明确 BDD 执行层级。
  - **落点**：spec-plan
  - **优先级**：high
  - 建议：如果采用 `cmd/api` Go HTTP scenario tests 作为 BDD 证据，plan/bdd-plan 应在创建时直接写明测试函数入口，不再要求 `test/scenarios` shell 目录。
- 为 idempotency middleware 增加 resource handoff 小节。
  - **落点**：README
  - **优先级**：medium
  - 本次已在 `backend/internal/api/practice/README.md` 记录 `SetResponseResource` 约定、内部 header 清理要求，以及 wrapped operation 必须写 `resource_type/resource_id`。
- 把 stale cursor / stale turn 纳入并发 gate 模板。
  - **落点**：spec-plan
  - **优先级**：high
  - 建议：任何 turn/event append 类计划都同时验证 row-lock 序列化与请求游标仍是 current view；BDD 场景名称中避免只写 `seq_no`。

## 5 建议优先级与后续动作

- 最高优先级：后续 `003-mode-policies-and-provenance` 派生或实施时，直接采用“cmd/api HTTP scenario tests 或 test/scenarios shell assets 二选一”的 BDD 入口声明，并把 assisted hint 的 stale-turn / current-turn 校验写入 checklist。
- 次优先级：如果后续其它 backend domain 复用 `SetResponseResource`，再把 `backend/internal/api/practice/README.md` 中的包级说明上移为 shared idempotency middleware README。
