# Practice Plan and Session Orchestration BDD Checklist

> **版本**: 2.9
> **状态**: active
> **更新日期**: 2026-07-21

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.PRACTICE.PLAN.001` plan/session 幂等与失败恢复

- [x] Owner behavior tests 覆盖 identity match、幂等、隔离、opening failure 与零重复事实。
- [x] 根 `make test` 已执行对应 Go tests；该结果是代码层行为证据，不是 E2E PASS。
- [x] 当前无 plan 创建/session start 真实 E2E owner；不创建 shell wrapper 或场景编号。

## Phase 8 BDD 适用性

- [x] 确认 `listPracticeSessions` 删除不创建新的 Behavior ID、BDD/E2E 场景或伪 PASS；代码层替代 gate 全部由主 checklist/test checklist 执行。
  <!-- verified: 2026-07-15 method=plan+scenario-negative result=PASS -->
- [x] 回归 `BDD.PRACTICE.PLAN.001`，证明保留的 `startPracticeSession` / `getPracticeSession` 行为未因公共列表删除而退化。
  <!-- verified: 2026-07-15 method=focused-owner-behavior-tests result=PASS -->

## Phase 9 活动会话恢复

- [x] Owner behavior test 证明同 user/plan 重复 start 返回同一 running session，零 prompt/AI/opening commit。
  <!-- verified: 2026-07-18 method=go-test tests="TestStartPracticeSessionRecoversRunningSessionWithoutOpeningSideEffects,TestStartPracticeSessionWaitsForQueuedRecoveryBeforeFinalizing" -->
- [x] Real PostgreSQL integration 证明 queued/running recovery、different-key 并发、精确 replay 与 user/plan isolation。
  <!-- verified: 2026-07-18 method=go-test-tags-integration marker="active-session-start-recovery=PASS" -->
- [x] Chrome skill 从正式入口恢复现有受影响 session；该结果是运行时 UI 补充证据，不声明新的 E2E 场景 ID。
  <!-- verified: 2026-07-18 method=chrome formalEntry=workspace result="same running session opened without PRACTICE_SESSION_CONFLICT" -->
- [x] Owner behavior/store tests 证明 orphaned queued start 在有限时间内 retryable 收敛，下一次 start 可恢复；迟到原 worker 与并发 completion 不产生错误 succeeded 快照或重复 opening facts。
  <!-- verified: 2026-07-19 method=owner-behavior+postgres-integration result=PASS evidence="bounded retryable timeout, session-row lock ordering and late-worker rollback" -->

## `BDD.PRACTICE.PLAN.002` 面试官身份与公司来源隔离

- [x] Owner prompt/registry behavior tests 断言 TargetJob/round 是面试官雇主身份的唯一来源，Resume 公司只属于候选人履历。
  <!-- verified: 2026-07-21 method=prompt-lint+registry-preflight result=PASS -->
- [x] Offline identity cases 覆盖目标公司与 Resume 公司不同、匿名目标公司、弱反例和 assistant-history 纠偏。
  <!-- verified: 2026-07-21 method=make-eval-offline result="32/32 offline and Promptfoo PASS; four identity classes present" -->
- [x] 配置可用时执行真实 provider screenshot-equivalent acceptance；只记录脱敏输入类别与是否出现 Resume 公司错误自称，不保存完整简历/prompt/raw output。
  <!-- verified: 2026-07-21 method=evalkit-live-complete coordinate=v0.3.0/v0.3.0 result="5/5 schema-valid; zero Resume-employer identity claims; named target preserved; raw payloads and audits deleted" limitation="live judge response JSON parsing failed, not claimed PASS" -->
- [x] 根 `make test` 与 F3/Practice 独立 gates 通过；若 provider 不可用，明确记录 unavailable，不能标记本行为完整 PASS。
  <!-- verified: 2026-07-21 method=full-closeout result="make test PASS Python 626/4628, Go all, frontend 137/1126; make eval-offline 32/32; focused F3/Practice gates PASS; completion provider available 5/5; live judge parser unavailable and not counted PASS" -->
