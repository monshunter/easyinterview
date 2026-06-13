# 002 — Practice Text Event Loop Test Checklist

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-06-13

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1: PracticeScreen 静态壳 + 路由替换 + i18n + sessionId 守卫

- [x] Phase 1 本计划定义的 `PracticeScreen.test.tsx`（DOM 锚点 + testid + 控件类型 + voice owner co-location 边界）、`usePracticeSessionLoader.test.ts`（5 态 + auto refresh）、`App.test.tsx`（practice case）、`i18n` namespace parity、`practiceModeSwitch.test.tsx` 测试项全部通过

## Phase 2: appendSessionEvent + AssistantAction + SessionStatus 消费

- [x] Phase 2 本计划定义的 `usePracticeEvents.test.ts`（5 kind body & header + retry 复用 + fresh action）、`idempotencyContract.test.ts`（双轨边界）、`AssistantActionRenderer.test.tsx`（5 type + provenance 隔离）、`usePracticeSession.test.ts`（七个 status 分支 + completed 防抖 + `draft/archived` 负向）、`appendSessionEventBody.test.ts` + `make validate-fixtures` + `make codegen-check` 单元 + contract 测试项全部通过

## Phase 3: assisted / strict 显隐 + RoleDropdown + 提示 / 跳过 / 暂停-恢复

- [x] Phase 3 本计划定义的 `usePracticeAssistance.test.ts`（strict / assisted × baseline / debrief 4 组合）、`practiceGoalParity.test.tsx`（显隐快照）、`practiceHints.test.tsx`（assisted 流 + hintCount 自增 + strict DOM 缺失）、`practiceSkip.test.tsx`、`practicePauseResume.test.tsx`、`RoleDropdown.test.tsx`（UI-only 0 调用）、`SessionMap.test.tsx`、`practiceModeSwitch.test.tsx`、`practiceStrictToggleLocked.test.tsx` 测试项全部通过

## Phase 4: completePracticeSession + handoff + 错误恢复 + sessionLost / conflict

- [x] Phase 4 本计划定义的 `useCompletePracticeSession.test.ts`（happy + replay + mismatch + network/5xx + StrictMode 双触发）、`practiceHandoffParams.test.ts`（字段集 + 不含展示字段）、`completePracticeSessionBody.test.ts` + fixture parity、`practiceSessionLost.test.tsx`（404 兜底）、`InterviewContext.test.tsx` INCREMENT_HINT_COUNT、`practiceCompletion.test.tsx`、`practiceClientEventConflict.test.tsx`、`practiceErrors.test.tsx`、`practiceConflict.test.tsx`、`practicePrivacy.test.tsx` 单元 + contract 测试项全部通过 <!-- verified: 2026-05-14 evidence=practice focused suite 27 files / 120 tests PASS; p0-046 and p0-047 scenario PASS -->

## Phase 5: Pixel parity + Scenario + Regression + Negative grep

- [x] Phase 5 本计划定义的 `practice.spec.ts` pixel parity（desktop + mobile + warm/light、dark、customAccent 主题 + 5 状态截图基线）、scenario 4 目录（p0-044/045/046/047）+ INDEX 更新、workspace P0.018-021 + backend-practice P0.022-026 + backend-practice 002 P0.038-043 regression rerun、`legacyNegative.test.ts` + CI grep（旧 testid / 旧 route / 旧 enum / getFeedbackReport / `createPracticeVoiceTurn` 非 voice-owner-hook 调用 / `Idempotency-Key.*appendSessionEvent` / raw text 泄漏）、`make docs-check` + `/sync-doc-index --fix-index` + `check-md-links` + 全量 Vitest + typecheck + build + `make build` 收口 gate 全部通过 <!-- verified: 2026-05-14 evidence=practice Playwright 11 passed / 1 skipped; p0-044..047 scenario PASS; workspace p0-018..021 PASS; backend-practice p0-022..026 and Go p0-038..043 PASS; full frontend Vitest 154 files / 907 tests PASS; typecheck/build/make build PASS; sync-doc-index --fix-index PASS; make docs-check PASS; revised 2026-05-23 for practice-voice-mvp owner boundary -->
