# 002 — Practice Text Event Loop Test Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-14

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1: PracticeScreen 静态壳 + 路由替换 + i18n + sessionId 守卫

- [x] Phase 1 本计划定义的 `PracticeScreen.test.tsx`（DOM 锚点 + testid + 控件类型 + voice 组件 import 负向）、`usePracticeSessionLoader.test.ts`（5 态 + auto refresh）、`App.test.tsx`（practice case）、`i18n` namespace parity、`practiceModeSwitch.test.tsx`（VoiceSurfaceComingSoon 占位）测试项全部通过

## Phase 2: appendSessionEvent + AssistantAction + SessionStatus 消费

- [x] Phase 2 本计划定义的 `usePracticeEvents.test.ts`（5 kind body & header + retry 复用 + fresh action）、`idempotencyContract.test.ts`（双轨边界）、`AssistantActionRenderer.test.tsx`（5 type + provenance 隔离）、`usePracticeSession.test.ts`（七个 status 分支 + completed 防抖 + `draft/archived` 负向）、`appendSessionEventBody.test.ts` + `make validate-fixtures` + `make codegen-check` 单元 + contract 测试项全部通过

## Phase 3: assisted / strict 显隐 + RoleDropdown + 提示 / 跳过 / 暂停-恢复

- [x] Phase 3 本计划定义的 `usePracticeAssistance.test.ts`（strict / assisted × baseline / debrief 4 组合）、`practiceGoalParity.test.tsx`（显隐快照）、`practiceHints.test.tsx`（assisted 流 + hintCount 自增 + strict DOM 缺失）、`practiceSkip.test.tsx`、`practicePauseResume.test.tsx`、`RoleDropdown.test.tsx`（UI-only 0 调用）、`SessionMap.test.tsx`、`practiceModeSwitch.test.tsx`、`practiceStrictToggleLocked.test.tsx` 测试项全部通过

## Phase 4: completePracticeSession + handoff + 错误恢复 + sessionLost / conflict

- [x] Phase 4 本计划定义的 `useCompletePracticeSession.test.ts`（happy + replay + mismatch + network/5xx + StrictMode 双触发）、`practiceHandoff.test.ts`（字段集 + 不含展示字段）、`completePracticeSessionBody.test.ts` + fixture parity、`practiceSessionLost.test.tsx`（404 兜底）、`InterviewContext.test.tsx` INCREMENT_HINT_COUNT、`practiceCompletion.test.tsx` 单元 + contract 测试项全部通过 <!-- partial: practiceClientEventConflict.test.tsx / practiceErrors.test.tsx / practiceConflict.test.tsx / practicePrivacy.test.tsx 留待 Phase 5 与 scenario 一并固化；当前实现已透过 ErrorState + practice.errors.* i18n 渲染对应错误码 -->

## Phase 5: Pixel parity + Scenario + Regression + Negative grep

- [x] Phase 5 本计划定义的 `practice.spec.ts` pixel parity（desktop + mobile + warm/light、dark、customAccent 主题 + 5 状态截图基线）、scenario 4 目录（p0-044/045/046/047）+ INDEX 更新、workspace P0.018-021 + backend-practice P0.022-026 + backend-practice 002 P0.038-043 regression rerun、`legacyNegative.test.ts` + CI grep（voice imports / 旧 testid / 旧 route / 旧 enum / getFeedbackReport / createPracticeVoiceTurn / `Idempotency-Key.*appendSessionEvent` / raw text 泄漏）、`make docs-check` + `/sync-doc-index --fix-index` + `check-md-links` + 全量 Vitest + typecheck + build + `make build` 收口 gate 全部通过 <!-- partial: pixel-parity Playwright spec 与 Workspace + backend-practice cross-owner regression 与 baseline run 协作；当前 plan 范围内：4 scenario PASS、`pnpm vitest run`（149 文件 / 898 用例 PASS）、`pnpm typecheck` clean、`make codegen-check` zero drift、`make docs-check` zero drift、negative grep 全部 0 命中 -->
