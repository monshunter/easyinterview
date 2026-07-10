# Backend Practice Mode Policies and Provenance

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

本 plan 承接 `appendSessionEvent` 的 hint 分支、mode policy、AssistantAction provenance 和 hint AI observability 合同：

- `mode='assisted'` 或 `mode='strict'` 时，`hint_requested` 通过 F3 `practice.turn.lightweight_observe` + A3 observed AIClient 返回 `AssistantAction{type:'show_hint'}`，并写入 `practice_turns.hint_text`。
- goal 仅决定练习来源，不改变 hint 可用性；`baseline` / `retry_current_round` / `next_round` 均遵守同一 optional hint policy。
- hint AI 失败走 graceful degrade：HTTP 200 + `AssistantAction{type:'session_wait'}`，session 保持 running，不写 `failure_code`，失败摘要进入 `ai_task_runs(task_type='hint_generate')`。
- `AssistantAction.provenance` wire JSON 只暴露 B2 `GenerationProvenance` 六字段；runtime 字段只进入 typed task run / service-local evidence。
- hint 路径不递增 turn count、不发 `practice.turn.completed` outbox、不写 domain audit event，且 payload / logs / metrics / ai_task_runs 不包含 question、answer、hint、prompt 或 provider secret 明文。

## 2 当前合同

### 2.1 Operation Matrix

| surface | fixture / scenario | backend behavior | persistence | AI dependency | coverage |
|---------|--------------------|------------------|-------------|---------------|----------|
| `appendSessionEvent` assisted hint | `appendSessionEvent.json` `hint-assisted-show` | `hint_requested` returns `200 + show_hint`; replay returns original hint snapshot | `practice_session_events`, `practice_turns.hint_text`, `ai_task_runs(hint_generate)` | F3 `practice.turn.lightweight_observe`, A3 Chat profile `practice.turn_observe.default` | `E2E.P0.048`, unit/store tests |
| `appendSessionEvent` strict-mode hint | `appendSessionEvent.json` `show-hint` | `hint_requested` returns `200 + show_hint`; replay returns original hint snapshot and leaves no pending event row | `practice_session_events`, `practice_turns.hint_text`, `ai_task_runs(hint_generate)` | F3 `practice.turn.lightweight_observe`, A3 Chat profile `practice.turn_observe.default` | `E2E.P0.049`, unit/store tests |
| AssistantAction provenance | current generated `GenerationProvenance` | response provenance key set is exactly six wire fields for show_hint / ask_question / ask_follow_up / session_wait / session_completed | runtime metadata excluded from wire | only AI-backed actions call A3 | `E2E.P0.050`, provenance tests |
| hint graceful degrade | `appendSessionEvent.json` `hint-assisted-ai-failed-degrade` | F3/A3/parser failures return `200 + session_wait` and keep session running | failed `ai_task_runs(hint_generate)` row where applicable | F3/A3 failure branches | `E2E.P0.051`, service tests |
| privacy / runtime boundary | no public fixture | no hint text, prompt, answer text, provider secret, or raw response in log/metric/audit/event/task-run payloads | sanitized event payloads and typed task columns only | observed AIClient redaction | `E2E.P0.051`, backend-practice out-of-scope lint |

### 2.2 Persistence And Event Boundary

- `practice_session_events` records the user event and replay envelope.
- `practice_turns.hint_text` is updated only for assisted success.
- `ai_task_runs.task_type='hint_generate'` records success/failure observability.
- hint does not create `practice.turn.completed` outbox events.
- hint does not update `practice_sessions.turn_count` and does not consume question budget.

### 2.3 Wire Boundary

- `AssistantAction.provenance` uses only `promptVersion`, `rubricVersion`, `modelId`, `language`, `featureFlag`, `dataSourceVersion`.
- `show_hint` uses `rubricVersion='not_applicable'`.
- `feature_key`, `model_profile_name`, provider, cost and latency do not leave runtime/typed-observability surfaces.

## 3 质量门禁

- **Plan 类型**: `feature-behavior + contract + code-internal`。
- **TDD 策略**: 适用。Focused tests cover optional hint dispatch across assisted / strict mode, AI success, graceful degrade, provenance JSON shape, store persistence, task-run writer, redaction and runtime boundary lint.
- **BDD 策略**: 适用。`E2E.P0.048` - `E2E.P0.051` cover assisted hint, strict-mode optional hint, provenance/task-run boundary and graceful degrade/privacy.
- **替代验证 gate**:
  - `cd backend && go test ./cmd/api -run 'TestE2EP0048|TestE2EP0049|TestE2EP0050|TestE2EP0051' -count=1`
  - `cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice ./internal/ai/aiclient ./internal/ai/registry ./cmd/api -count=1`
  - `python3 scripts/lint/backend_practice_out_of_scope.py --repo-root . --phase all`
  - `python3 -m pytest scripts/lint/backend_practice_out_of_scope_test.py -q`
  - `make validate-fixtures`
  - `make codegen-check`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-practice/plans/003-mode-policies-and-provenance/context.yaml --target backend`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

## 4 实施步骤

### Phase 0: contract preflight

- Confirm backend-practice spec current decisions for `PracticeMode`, graceful degrade, hint lifecycle and provenance.
- Confirm B4 baseline and A3 writer accept `hint_generate`.
- Confirm F3 `practice.turn.lightweight_observe` and model profile `practice.turn_observe.default` resolve in tests.
- Confirm `appendSessionEvent` fixtures include assisted success, strict-mode success and assisted degrade variants.

### Phase 1: optional hint dispatch and strict mode

- Dispatch `hint_requested` as optional in-session assistance.
- Assisted and strict modes return a pending `show_hint` outcome for service AI application.
- Replay preserves the original hint response and leaves no pending event row.
- Unit tests cover mode × goal matrix and strict-mode behavior.

### Phase 2: assisted hint AI and persistence

- Apply F3/A3 hint generation for assisted and strict-mode hint.
- Persist `practice_turns.hint_text` only on hint success.
- Preserve turn/session lifecycle: no turn count increment, no turn status change, no outbox, no audit event.
- Store replay payload preserves the original hint response snapshot.

### Phase 3: provenance and graceful degrade

- Marshal AssistantAction provenance with only the six OpenAPI wire fields across action types.
- Record F3/A3/parser failures as graceful degrade, with failed `ai_task_runs(hint_generate)` where required.
- Keep degrade reason out of wire response and domain audit events.

### Phase 4: privacy, observability and closeout

- Enforce redaction for logs, metrics, audit, events and typed task-run payloads.
- Run backend-practice out-of-scope lint over runtime/scenario/generated surfaces.
- Run BDD gates and update plan/index evidence.

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | Assisted hint returns show_hint and writes hint_text/task-run evidence | `TestE2EP0048PracticeHintAssistedAcrossGoals`, service/store tests |
| A-2 | Strict-mode hint remains available, replayable and leaves no pending reservation | `TestE2EP0049PracticeHintOptionalAcrossStrictModeGoals`, strict-mode replay tests |
| A-3 | AssistantAction provenance wire JSON has exactly six keys | `TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns`, provenance unit tests |
| A-4 | Hint AI failures degrade without failing the session | `TestE2EP0051PracticeHintDegradeAndPrivacy`, `TestApplyHintAIGracefulDegradeMatrix` |
| A-5 | Privacy/runtime boundary has no real residuals | backend-practice out-of-scope lint, redaction tests, pruning-surface lint |

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-07 | 1.4 | Compress owner docs to current hint mode policy, provenance, task-run and privacy contract. |
| 2026-07-06 | 1.3 | Reconcile current goal matrix and out-of-scope gate wording after product-scope pruning. |
| 2026-07-10 | 1.6 | Rename strict-mode wording and test references to optional hint policy while preserving behavior. |
| 2026-07-09 | 1.5 | Align docs with real-interview simplification: strict mode no longer rejects hints; hint remains optional assistance across current goals. |
