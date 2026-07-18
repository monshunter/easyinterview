# Active Session Recovery Hardening 交付复盘报告

> **日期**: 2026-07-19
> **审查人**: Codex

**关联计划**: [Backend Practice 001 Phase 9](../spec/backend-practice/plans/001-plan-and-session-orchestration/plan.md)
**关联 Bug**: [BUG-0190](../bugs/BUG-0190.md)

## 1 复盘范围与成功证据

- 本次交付修复 active-session recovery 代码审查中的两个 P1：恢复最终化与 completion 的竞态，以及 orphaned queued start 的无限等待/轮询；同时补上 timeout 与迟到原 worker 之间的事务 fencing。
- Service tests 证明 queued recovery 使用 100ms 起步、1 秒封顶的有界退避，在 35 秒业务边界后写 retryable `AI_PROVIDER_TIMEOUT`；caller cancellation 继续不修改 session/key，timeout 输给 running transition 时改走成功恢复。
- SQL mock 与真实 PostgreSQL integration 证明 recovery 先锁 session row；completion 先提交时 recovery 返回 conflict 且 key 不伪成功；timeout 先提交时迟到原 worker 的 message/event/outbox/audit 全部回滚。
- practice API/domain/store 单测与全部 store integration 通过；根 `make test` 通过 Python 584 tests / 4583 subtests、Go 全包与 frontend 127 files / 1035 tests。
- build、lint、codegen、OpenAPI diff、fixtures、docs/context/index 与直接 `git diff --check` 均通过；OpenAPI diff 为 0 breaking / 0 additive / 0 informational。

## 2 会话中的主要阻点/痛点

### 2.1 原 Phase 9 只覆盖 recovery happy path，没有覆盖失联与相反提交顺序

- **证据**：既有测试覆盖 running recovery、queued 最终进入 running、context cancellation 和 different-key 并发，但原 starter 永不推进时 service 只按 100ms 无限轮询；`CommitSessionStartRecovery` 也未在读取 snapshot 前锁 session row。
- **影响**：历史 Phase 9 与 BUG-0189 已有全绿证据和真实 UI 验收，仍留下可挂死 HTTP 请求、持续数据库读和 stale succeeded response 的 P1 风险。

### 2.2 单独增加 timeout 会产生迟到 worker 复活问题

- **证据**：`CommitSessionStart` 与 `FailSessionStart` 的 session update 原先都未要求 `status='queued'`。若 recovery timeout 先写 failed，原 worker 仍可随后写 running 与 opening facts。
- **影响**：直觉上的局部修复会引入更隐蔽的双终态/重复事实竞态，需要同时修改 service policy、commit/fail predicate 与 integration 顺序测试。

### 2.3 Closeout 证据引用了不存在的 Make target

- **证据**：执行历史 checklist 中的 `make ... git-diff-check` 时，当前 Makefile 返回 `No rule to make target 'git-diff-check'`；直接 `git diff --check` 通过。
- **影响**：组合门禁被最后一个入口漂移打断，容易把“前序 gate 已通过”和“整条命令失败”混为一谈，也增加不必要的排查成本。

## 3 根因归类

- Recovery owner 合同缺少 producer 失联、terminal transition 与 late producer 三种失败顺序。
  - **类别**：spec/plan
  - **处理状态**：本轮已在 backend-practice spec、Phase 9 plan/test/BDD 中原地补齐。
- 原实现把“事务包住读取与 key update”误当成线性化，但没有锁定可被 completion 修改的 owner row；timeout 也未与原 commit 建立 compare-and-set winner。
  - **类别**：spec/plan
  - **处理状态**：本轮已用 session row lock、queued fence 和真实 PostgreSQL 相反顺序测试固化。
- Diff gate 没有一个当前可执行、稳定命名的 Make owner，历史文档使用了已经不存在的 target 名。
  - **类别**：README / spec-plan
  - **处理状态**：本轮证据改用直接命令并明确备注；通用入口尚可进一步收敛。
- GREEN 初版即使有 35 秒上限仍保留固定 100ms 轮询，单请求最多约 350 次读取。
  - **类别**：无需仓库改动
  - **处理状态**：同轮自审后改为有上限指数退避，并以 deterministic waiter 测试固定负载形态。

## 4 对流程资产的改进建议

- 在 `/plan-code-review` 的并发恢复审查清单中增加“三顺序”反查：producer 永不完成、terminal owner 先提交、recovery timeout 先提交后 producer 迟到；每个顺序同时检查 idempotency state 与全部副作用回滚。
  - **落点**：skill
  - **优先级**：high
- 对所有“轮询等待另一请求推进持久化状态”的计划，要求明确 durable deadline、poll/backoff budget、timeout convergence 与 late-worker fence；caller context 只能作为提前取消，不得成为唯一恢复生命周期。
  - **落点**：spec/plan
  - **优先级**：high
- 统一 diff closeout 入口：要么在 Makefile 恢复一个明确 target，要么把当前 docs/checklist 模板统一写成 `git diff --check`，避免继续复制失效的 `git-diff-check` 名称。
  - **落点**：README / spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

1. 下一轮最高价值动作是更新 `/plan-code-review` 的并发/恢复审查 gate，把本次三种失败顺序推广到所有 reservation → async work → commit 流程；这能在实现完成前发现同类遗漏。
2. 同步审计当前 active plans 中的 polling/recovery 语义，优先检查只依赖 caller context、没有持久化 deadline 或没有 late-worker generation/status fence 的路径。
3. 将 `git-diff-check` 入口漂移作为独立低风险治理修订处理；在统一 owner 前，closeout 证据继续明确记录实际执行的 `git diff --check`。
