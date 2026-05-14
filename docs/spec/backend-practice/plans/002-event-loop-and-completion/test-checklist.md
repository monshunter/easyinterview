# 002 — Event Loop and Completion Test Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-14

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 0: 跨 spec 前置修订 + Preflight

- [x] Phase 0 本计划定义的 `triggerEventSemantic` enum + `make lint-events` / `make codegen-events-check` / generated `JobTriggerEventSemantic*` 常量 + `IsSourceEventOnly` 谓词单元测试 / OpenAPI codegen / fixtures validator / sync-doc-index 与 F3 baseline preflight 测试项全部通过（注：runtime outbox→asynq dispatcher 集成测试归 future `backend-async-runner` plan，002 阶段不在范围内）

## Phase 1: AppendSessionEvent state machine 与 turn-status 域

- [x] Phase 1 本计划定义的 `SessionEventService` 状态机、handleAnswerSubmitted / handleHintRequested / handleTurnSkipped / handleSessionPaused / handleSessionResumed、unknown kind、AssistantAction provenance 与 `turn_status` mapping 单元测试项全部通过

## Phase 2: AppendSessionEvent vertical slice

- [x] Phase 2 本计划定义的 repository（主流程 / stale-turn conflict / replay / mismatch / cross-user）、outbox_emitter practice.turn.completed、service（含 F3 follow_up、AI 失败退化、`answer_submitted` 缺失 `payload.answerText` 校验、server-owned `follow_up_count` 决策）、handler（拒 header / required `occurredAt` / 200 wire shape）、error mapping、router 单元 + 集成 + contract 测试项全部通过

## Phase 3: CompletePracticeSession vertical slice

- [x] Phase 3 本计划定义的 repository（主流程 / D-35 replay + `async_jobs.dedupe_key=sessionId` lookup / status guard / cross-user / async_jobs dedupe）、outbox_emitter practice.session.completed、service replay、handler（idempotency middleware + required `clientCompletedAt` + 双 key + cross-user + illegal completion status conflict）、idempotency middleware 复用、error mapping 单元 + 集成 + contract 测试项全部通过

## Phase 4: 隐私 / 观测 / Legacy-Negative

- [x] Phase 4 本计划定义的 redaction、metric label allowlist、out-of-scope boundary、legacy-negative grep、BDD 编号碰撞反查与 `make codegen-check` / `make lint-events` / `make codegen-events-check` / `cd backend && go test ./...` / `python3 scripts/lint/conventions_drift.py --repo-root .` 收口 gate 全部通过
