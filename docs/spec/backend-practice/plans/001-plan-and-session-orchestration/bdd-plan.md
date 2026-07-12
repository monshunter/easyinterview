# 001 Plan and Session Orchestration BDD Plan

> **版本**: 2.3
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
| E2E.P0.023 | grounding primary | 6 | parse failed/empty structured profile but complete source snapshot exists | start session | opening AI payload contains the complete tail marker and uses only resume-backed facts |
| E2E.P0.024 | grounding failure | 6 | all resume content fields are empty | start session | typed validation, zero AI call, and no opening assistant message |
| E2E.P0.022 | round identity primary | 7 | Real PostgreSQL TargetJob has a bound resume and canonical structured rounds with complete provenance, including equal durations and `1,2,4` sequences | create/read baseline plan | response and DB persist the first incomplete exact `roundId/roundSequence`; no duration-only reuse and successor follows the next existing canonical row |
| E2E.P0.070 | derived plan primary/replay | 7 | same-bound-resume report source round is complete and has an immediate existing successor | create retry/next and replay IK | retry preserves source pair; next chooses current canonical successor; replay returns the same pair |
| E2E.P0.072 | round identity failure/isolation | 7 | request/source/completion resume mismatch, missing provenance, overflow/case-drift round, mismatched budget, all complete, legacy null source or cross-user report | create plan | typed failure, no plan insert, no cross-user or wrong-resume disclosure |

## 2 Scenario Assets

Existing scenario IDs/directories are revised in place to the conversation contract. Trigger logs must contain actual Go/contract runner evidence; verify scripts must assert pass markers and stale-question negative checks.
