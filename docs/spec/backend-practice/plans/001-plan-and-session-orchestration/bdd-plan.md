# 001 Plan and Session Orchestration BDD Plan

> **版本**: 2.0
> **状态**: completed
> **更新日期**: 2026-07-12

## 1 Scenario Matrix

| 场景 ID | 类型 | Phase | Given | When | Then |
|---------|------|-------|-------|------|------|
| E2E.P0.022 | primary | 2 | valid target/resume | create/read plan | plan has context fields and no question/mode/hint fields |
| E2E.P0.023 | primary | 3/4 | ready plan | start/read session | exactly one opening assistant message and ordered messages |
| E2E.P0.024 | failure/recovery | 3 | opening AI fails | retry same IK | no duplicate session/message/outbox; eventual success |
| E2E.P0.025 | boundary/security | 3/4 | replay/mismatch/cross-user inputs | call start/read | deterministic replay/conflict/404 isolation |
| E2E.P0.026 | privacy/regression | 5 | conversation start completes | inspect evidence | no raw message leakage or stale question contract |

## 2 Scenario Assets

Existing scenario IDs/directories are revised in place to the conversation contract. Trigger logs must contain actual Go/contract runner evidence; verify scripts must assert pass markers and stale-question negative checks.
