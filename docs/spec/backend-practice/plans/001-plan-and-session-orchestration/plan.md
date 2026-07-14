# 001 — Plan and Session Orchestration

> **版本**: 2.5
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把 Practice plan/session foundation 从 question/turn 模型原地重构为 conversation/message 模型：

- plan 删除 question budget、mode 和 hint configuration。
- baseline migration 删除 `practice_turns/question_assessments`，新增 `practice_messages`。
- OpenAPI 删除 PracticeTurn/AssistantAction/event-answer schema，新增 message schemas 与 `sendPracticeMessage` operation。
- `startPracticeSession` 通过 `practice.session.chat` 生成 opening assistant message。
- `getPracticeSession` 返回 ordered messages。
- 保持 user isolation、idempotency、AI failure recovery、privacy 和 codegen drift gates。

## 2 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticePlan` | current plan fixtures | start helpers | practice plan owner | plan + idempotency/audit | none | 当前无真实 E2E owner；root `make test` |
| `getPracticePlan` | current plan fixtures | start/read helpers | practice plan read owner | plan read | none | 当前无真实 E2E owner；root `make test` |
| `startPracticeSession` | current session fixtures | start helpers | practice session owner | session + opening message | `practice.session.chat` | 当前无真实 E2E owner；root `make test` |
| `getPracticeSession` | current session fixtures | Practice loader | practice read owner | session + messages | none | 当前无真实 E2E owner；root `make test` |

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + contract + migration + backend + codegen。
- **TDD 策略**: 每个 phase 可先运行 focused Red test 获取开发反馈，再修改 OpenAPI/shared/migration/service/store；阶段完成由根 `make test` 承接。
- **BDD 策略**: plan create/read、session start/opening 与 isolation/privacy 保留 Given/When/Then 合同；当前无真实 API/UI E2E owner。
- **替代验证 gate**: OpenAPI codegen/fixture、conventions codegen、migration up-down-up、prompt/schema/eval lint、privacy negative search。

## 4 Coverage Matrix

| Source | Category | Plan phase | Verification | Negative scope |
|--------|----------|------------|--------------|----------------|
| D-24 conversation | cross-layer contract | 1 | codegen/fixture/migration/prompt gates | questionBudget/PracticeTurn/QuestionAssessment |
| session read | boundary | 4 | ordered/empty/cross-user tests | local fixture transcript |

## 5 实施步骤

### Phase 1: Contract and baseline rebase

- Red tests lock the new 37-operation inventory, message schemas, 11 shared enums, 21 app tables and 6 prompt coordinates.
- Replace `appendSessionEvent` HTTP operation with `sendPracticeMessage` while keeping total endpoint count stable.
- Remove `PracticeMode` / `QuestionReviewStatus` and all question/report-review schemas.
- Rewrite baseline SQL, enum sources, seeds, prompt/rubric/profile/eval coordinates for `practice.session.chat`.
- Regenerate Go/TS artifacts and validate fixtures/baseline.

### Phase 2: PracticePlan simplification

- Remove question budget, mode and hints from request/domain/store.
- Preserve goal, interviewer persona, difficulty, language, time budget, resume/source/focus context.
- Cover baseline and report-derived plans, idempotency and cross-user isolation.

### Phase 3: Session start with opening message

- Replace first-question generator with chat opening using empty ordered history.
- Reserve session, call AI outside transaction, persist one assistant message and session_started fact.
- Retry same IK after timeout without duplicate session/opening/outbox.
- Validate language and schema with one repair; never emit canned text.

### Phase 4: Session read model

- Return ordered messages with stable roles/seq/timestamps.
- Cover empty/loading states, pagination decision (P0 full bounded list), cross-user 404 and deleted context.

### Phase 5: Privacy, migration and BDD closeout

- Prove raw message content exists only in `practice_messages` and authorized prompt/read/report input.
- Sync owner docs/index/context; leave plan active until downstream 002/003 and frontend/report owners close.

### Phase 6: Complete resume grounding for session start

- RED store/service tests prove `startPracticeSession` prefers complete `parsed_text_snapshot`, falls back to `original_text` then non-empty `structured_profile`, preserves a long input tail marker in the AI payload, and makes zero AI calls when all three are empty.
- GREEN both the start reservation SQL and shared conversation context use one `ResumeContext` field with no character/token slicing; remove the `resume context unavailable` model fallback and return typed `VALIDATION_FAILED` before prompt resolve/AI.
- Revalidate `practice_plans.resume_id = target_jobs.resume_id` when reserving both session start and follow-up messages so an old plan/session cannot continue with a stale same-user resume after TargetJob rebinding.
- Harden `practice.session.chat` prompt/schema/eval contract: immutable policy is sent as a `system` message while JD, complete resume, persisted round, persona and ordered history are JSON-encoded as untrusted user data. Candidate company/project/product/technology facts require persisted resume or candidate-authored `user` evidence; `assistant` history is continuity-only and cannot turn a prior model hallucination into a candidate fact. Unnamed projects trigger a clarification question rather than an invented project. Persona controls tone/perspective only and cannot create facts or replace round identity.

### Phase 7: Persisted canonical round identity and plan selection

- RED OpenAPI/domain/store tests require optional request `roundId`, server-derived `roundSequence`, paired persistence/readback, and no client-supplied sequence.
- GREEN create-plan first requires request `resumeId` to equal the current user's persisted `target_jobs.resume_id`, even when a different resume is owned by the same user. It validates canonical TargetJob rounds and the same-bound-resume completion ledger in one transaction: baseline selects the first incomplete round; retry uses a same-bound-resume source report round; next uses the immediate existing canonical successor and requires it to equal the current incomplete round.
- Canonical summary admission requires non-empty `promptVersion/rubricVersion/modelId/language/dataSourceVersion` provenance, positive `int32` sequences that are unique and strictly increasing while allowing gaps such as `1,2,4`, and exact lowercase round types from the OpenAPI allowlist. Overflow, case normalization, missing provenance and wrong-resume source/completion facts fail closed.
- Session start loads exact round name/type/focus by the persisted pair and renders that structure into the JSON round context; `interviewerPersona` remains a separate field and cannot substitute for round context. A `finish_reason=length` response is repairable once and is never committed merely because its truncated content happens to parse as JSON.
- Reject mismatched requested round, mismatched duration, duplicate/invalid canonical IDs, completed ladders, missing source identity and legacy null reuse. Adjacent equal-duration rounds must remain distinct.
- Preserve idempotency replay: the same key returns the original round pair; a changed round intent with the same key conflicts instead of silently returning another round.

## 6 验收标准

- Contract truth sources contain message/conversation shapes and zero current question/hint/mode shapes.
- start returns exactly one opening assistant message; retries do not duplicate it.
- get session returns ordered messages and preserves isolation/privacy.
- All checklist, test checklist and BDD checklist items pass before completed.
- A parse-failed resume with a complete `parsed_text_snapshot` can start normally and sends the full tail marker to AI; a resume with no readable content cannot call AI or persist an opening assistant message.
- Every newly created plan persists and returns the exact canonical round pair; plan creation cannot infer/reuse a round only from duration, lifecycle status, frontend storage or URL state.
- A plan cannot be created from another same-user resume or a wrong-resume source/completion fact; non-contiguous `1,2,4` ladders select `4` after `2`, not a fabricated `3`.
- Prompt policy remains in the system role; JSON-encoded JD/resume/history/persona content cannot promote embedded instructions into policy, and persona cannot supply resume facts or round identity.
- Start/send fail closed after TargetJob resume rebinding, and repeated `finish_reason=length` output persists no assistant reply.

## 7 风险与应对

| 风险 | 应对 |
|------|------|
| baseline edit leaves stale generated artifacts | codegen-check + baseline diff + negative search |
| AI timeout duplicates opening | three-stage reservation + IK replay tests |
| raw transcript leaks through events/logs | allowlist event payload + redaction tests |
| report consumer still expects turns | downstream backend-review/frontend-report plan gates block closeout |
| empty structured profile silently removes the real resume | Phase 6 SQL precedence and long-input tail-marker tests make the source snapshot the canonical grounding input |
| generic prompt turns an ambiguous user reference into a fabricated project | prompt contract/eval require evidence or clarification and prohibit invented resume facts |
| an assistant-invented project is repeated as conversation evidence | permit candidate facts only from persisted resume or candidate-authored user messages; keep assistant history continuity-only and lock with a negative eval |
| adjacent rounds share the same duration | round pair, not time budget, is the identity and reuse key |
| client or legacy data tries to skip ahead | transactional current-round/source validation fails closed |
| a same-user but unbound resume or report is accepted | TargetJob-bound resume equality is checked before source/completion facts can select or derive a plan |
| a stale plan/session continues after TargetJob resume rebinding | start/send reservations re-check `target_jobs.resume_id = practice_plans.resume_id` before loading any prompt context |
| provider returns parseable but length-truncated JSON | treat `finish_reason=length` as `AI_OUTPUT_INVALID`; repair once, then fail without committing an assistant message |
| summary sequence/type/provenance drifts across AI and OpenAPI layers | reject missing provenance, non-int32/non-increasing sequence and non-lowercase/unknown type instead of normalizing silently |
| resume/JD/history contains prompt-like instructions | encode the entire business context as untrusted JSON in the user message and keep immutable policy in the system message |

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 2.5 | Prevent assistant-authored history from becoming candidate evidence and add the invented-project amplification regression gate. |
| 2026-07-12 | 2.4 | Require TargetJob-bound resume provenance across plan/source/completion facts, admit strict non-contiguous int32 round directories, and separate system policy from JSON-encoded untrusted interview context. |
| 2026-07-12 | 2.3 | Reopen Phase 7 for canonical round identity persistence and ledger-validated plan selection. |
| 2026-07-12 | 2.2 | Reopen start orchestration for full source-snapshot grounding, no-context fail-closed behavior, and evidence-only project questions. |
| 2026-07-12 | 2.0 | Reopen for conversation/message model and opening assistant message. |
