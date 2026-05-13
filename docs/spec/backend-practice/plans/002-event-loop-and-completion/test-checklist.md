# 002 — Event Loop and Completion Test Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-13

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 0: 跨 spec 前置修订 + Preflight

- [x] Phase 0 本计划定义的 `triggerEventSemantic` enum + `make lint-events` / `make codegen-events-check` / generated `JobTriggerEventSemantic*` 常量 + `IsSourceEventOnly` 谓词单元测试 / OpenAPI codegen / fixtures validator / sync-doc-index 与 F3 baseline preflight 测试项全部通过（注：runtime outbox→asynq dispatcher 集成测试归 future `backend-async-runner` plan，002 阶段不在范围内）

## Phase 1: AppendSessionEvent state machine 与 turn-status 域

- [ ] Phase 1 本计划定义的 `SessionEventService` 状态机、handleAnswerSubmitted / handleHintRequested / handleTurnSkipped / handleSessionPaused / handleSessionResumed、unknown kind、AssistantAction provenance 与 `turn_status` mapping 单元测试项全部通过

## Phase 2: AppendSessionEvent vertical slice

- [ ] Phase 2 本计划定义的 repository（主流程 / concurrent seq_no / replay / mismatch / cross-user）、outbox_emitter practice.turn.completed、service（含 F3 follow_up 与 AI 失败退化）、handler（拒 header / 200 wire shape）、error mapping、router 单元 + 集成 + contract 测试项全部通过

## Phase 3: CompletePracticeSession vertical slice

- [ ] Phase 3 本计划定义的 repository（主流程 / D-35 replay / concurrent / cross-user / async_jobs dedupe）、outbox_emitter practice.session.completed、service replay、handler（idempotency middleware + 双 key + cross-user）、idempotency middleware 复用、error mapping 单元 + 集成 + contract 测试项全部通过

## Phase 4: 隐私 / 观测 / Legacy-Negative

- [ ] Phase 4 本计划定义的 redaction、metric label allowlist、out-of-scope boundary、legacy-negative grep 与 `make codegen-check` / `make lint-events` / `make codegen-events-check` / `cd backend && go test ./...` / `python3 scripts/lint/conventions_drift.py --repo-root .` 收口 gate 全部通过
