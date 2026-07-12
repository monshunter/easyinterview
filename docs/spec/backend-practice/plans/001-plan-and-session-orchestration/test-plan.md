# 001 Plan and Session Orchestration Test Plan

> **版本**: 2.0
> **状态**: completed
> **更新日期**: 2026-07-12

## Phase 1: Contract tests

- OpenAPI inventory/schema/fixture/codegen tests reject old question/event/report shapes.
- Shared generator tests assert PracticeMode/QuestionReviewStatus absence.
- Migration contract/probe asserts practice_messages and absence of practice_turns/question_assessments.
- Prompt/rubric/seed/eval inventory asserts practice.session.chat and six coordinates.

## Phase 2: Plan tests

- Request validation, store SQL, idempotency and report-derived source isolation without question/mode/hint fields.

## Phase 3: Start tests

- Opening happy path, language repair, invalid output, timeout retry, IK replay/mismatch, outbox/message uniqueness and privacy.

## Phase 4: Read tests

- Ordered messages, empty message list, cross-user 404, list pagination and generated DTO mapping.

## Phase 5: Gate set

- Focused Practice/OpenAPI/migration/prompt tests, full backend, staticcheck, codegen/fixture/migrate/docs/diff gates.
