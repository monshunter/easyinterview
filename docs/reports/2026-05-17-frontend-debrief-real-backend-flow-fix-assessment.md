# Frontend Debrief Real Backend Flow Fix 交付复盘报告

> **日期**: 2026-05-17
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复 debrief 真实后端运行路径中的 6 条 review findings：`createDebrief` answer summary、replay session start、`listPracticeSessions` 后端 wiring、JD `analysisStatus` filter、manual fallback、resume asset re-selection。
- 成功证据：
  - `pnpm --dir frontend exec vitest run src/app/screens/debrief/DebriefScreen.test.tsx src/app/screens/debrief/DebriefPickerRegression.test.tsx src/app/i18n/__tests__/debriefI18nCoverage.test.ts src/app/i18n/localeFiles.test.ts`
  - `go test ./backend/internal/api/practice ./backend/internal/practice ./backend/internal/store/practice ./backend/cmd/api`
  - `pnpm --dir frontend typecheck`
  - `bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh; rc=$?; bash scripts/cleanup.sh; exit $rc` in `test/scenarios/e2e/p0-068-debrief-failure-and-handoff`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-debrief/plans/001-debrief-screen-and-handoff/context.yaml --docs-root docs --target frontend`
  - `make docs-check`
  - `git diff --check`
- Bug KB 已记录为 [BUG-0069](../bugs/BUG-0069.md)。

## 2 会话中的主要阻点/痛点

- Generated client and fixtures existed for `listPracticeSessions`, but real backend route/service/store wiring did not.
  - **证据**：OpenAPI/generated artifacts exposed `GET /practice/sessions`; `backend/cmd/api/main.go` originally registered only POST `/practice/sessions` and GET `/practice/sessions/{sessionId}`.
  - **影响**：fixture-backed tests could pass while the real picker returned 404.
- Step 0 UI allowed entries with only question text.
  - **证据**：`GuidedDebriefRecord` created entries without `myAnswerSummary`; `useSubmitDebrief` mapped missing summaries to `""`.
  - **影响**：backend `createDebrief` validation rejected normal submissions.
- Replay CTA treated debrief replay as route navigation rather than a session lifecycle operation.
  - **证据**：the CTA forwarded a completed mock session ID or no `sessionId`; `PracticeScreen` requires a usable started session.
  - **影响**：users reached session-lost or replayed the wrong session context.

## 3 根因归类

- Mock/generator evidence was accepted without proving the real runtime operation chain.
  - **类别**：spec-plan
- UI submit tests asserted success payload shape too shallowly and did not include required backend fields.
  - **类别**：spec-plan
- The plan handoff did not force replay CTA tests to verify direct session creation before navigation.
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- Add a debrief plan-code-review gate requiring every frontend-consumed generated operation to be traced through `openapi`, fixture, generated client/server, backend handler, service/store, route registration, and at least one focused test.
  - **落点**：spec-plan
  - **优先级**：high
- Add a debrief replay checklist assertion that any CTA entering `practice` must create or resolve a fresh usable `sessionId` before navigation.
  - **落点**：spec-plan
  - **优先级**：high
- Keep form submit tests contract-aware by asserting backend-required non-empty fields, not only navigation or mocked success state.
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- Highest value next step: push the repair branch or open PR review after the real-backend-flow commit lands.
- Lower priority: consider extracting a shared frontend helper for direct practice-session start CTAs once report replay and debrief replay stabilize around the same lifecycle contract.
