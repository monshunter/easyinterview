# 001 Plan and Session Orchestration Test Plan

> **版本**: 2.4
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

## Phase 6: Resume grounding

- SQL precedence tests cover non-empty `parsed_text_snapshot`, `original_text`, `structured_profile`, plus all-empty context.
- Start service tests use a long synthetic resume with a unique tail marker and assert the complete AI payload contains it without slicing.
- Empty context returns `VALIDATION_FAILED` before prompt resolution/AI and persists no opening assistant reply.
- Prompt lint/eval rejects invented resume projects, requires clarification for unnamed projects, and proves an assistant-only “智能客服” claim cannot become candidate evidence in the next message.
- Payload-role tests require immutable policy in `system`; JD/resume/round/persona/history are JSON-encoded as untrusted user data, including quote/newline/closing-tag injection cases. Persona changes style only and cannot replace facts or round identity; candidate facts can come only from persisted resume or candidate-authored `user` messages.
- Runtime tests split the trusted template before rendering, normalize the system-language tag, reject raw untrusted placeholders in registry preflight, and prove `finish_reason=length` repairs once then fails closed without a committed assistant message.

## Phase 7: Round identity and plan selection

- Contract/domain/store tests cover optional request round intent, server-derived sequence, paired persistence/readback and IK replay/mismatch.
- Transactional plan selection first requires request/source/completion resume to equal the TargetJob binding, then covers baseline first-incomplete, exact report retry round, immediate existing successor, duplicate completion facts and all-complete failure.
- Start reservation/template tests require exact round name/type/focus and reject the current persona-as-round substitution plus legacy/mismatched plan identity.
- Real Postgres tests rebind the TargetJob to a different same-user resume after plan/session creation and require both start and send reservations to fail closed.
- Negative matrix requires non-empty summary provenance, lowercase allowlisted type and positive int32 strictly increasing/unique sequences; it accepts `1,2,4` and chooses `4` after `2`, while rejecting overflow, case drift, same-duration ambiguity, requested round/budget mismatch, wrong-resume facts, legacy null source/ready plan and cross-user source.
