# 001 Home + JD Import + Parse Checklist

> **版本**: 2.8
> **状态**: completed
> **更新日期**: 2026-07-09

**关联计划**: [plan](./plan.md)

## Phase 1: Home 当前入口

- [x] 1.1 Home 源级复刻当前 `ui-design/src/screen-home.jsx::HomeScreen`：Hero label/title、JD 输入卡、输入卡底部 upload/URL source actions、ready 简历下拉框、创建简历入口、提交区、最近 3 张模拟面试卡片和 More handoff。
- [x] 1.2 Home 使用 generated client 调 `listResumes`、`listTargetJobs`、`createUploadPresign`、`importTargetJob`；paste/file/URL source discriminator、side-effect idempotency key、错误态和 pending import continuation 均有 focused Vitest 覆盖。
- [x] 1.3 Home import 前必须显式选择 ready 简历；成功进入 `parse` 时 params 携带真实 `resumeId`。
- [x] 1.4 BDD-Gate: `E2E.P0.014` 覆盖默认渲染、empty/one/twelve-plus fixtures、3-card cap、More handoff、theme/i18n 和 source/resume/submit layout。
- [x] 1.5 BDD-Gate: `E2E.P0.015` 覆盖 paste/upload/URL import、4xx/failed path、privacy gate、generated client request contract 和 real-mode generated-client preflight。

## Phase 2: Historical pre-readonly Parse confirmation and handoff

- [x] 2.1 Historical pre-Phase 6 Parse parity covered loading, preview, failed state, editable basics, requirements, hidden signals, round assumptions, resume binding and footer actions; Phase 6 now supersedes success preview with a readonly receipt.
- [x] 2.2 Historical generated-client gates covered `getTargetJob`, `listResumes` and `updateTargetJob` contract behavior; current Parse success detail uses `getTargetJob` / `listResumes` / practice handoff and must not consume `updateTargetJob`.
- [x] 2.3 Historical route `resumeId` inheritance and picker fallback were covered; current Parse success detail only displays the saved binding and disables Start when that binding is missing.
- [x] 2.4 Historical Save plan / workspace auto-start handoff was covered before readonly simplification; current success path has no Save plan action and Start enters practice directly.
- [x] 2.5 BDD-Gate: `E2E.P0.016` keeps the historical import-to-detail lineage and now covers readonly receipt, direct Start handoff, auth continuation and privacy checks.

## Phase 3: 收口验证

- [x] 3.1 `validate_context.py frontend-home-job-picks-and-parse/001 frontend` 通过。
- [x] 3.2 Focused Home/Parse Vitest、frontend typecheck 与 `make validate-fixtures` 通过。
- [x] 3.3 `E2E.P0.014`、`E2E.P0.015`、`E2E.P0.016` 的 `setup -> trigger -> verify -> cleanup` 通过。
- [x] 3.4 `sync-doc-index --check`、`make docs-check`、`git diff --check` 和 `make lint-core-loop-pruning-surface` 通过。

## Phase 4: Import resume binding remediation

- [x] 4.1 Home paste/upload/URL imports include the selected `resumeId` in generated `importTargetJob` request bodies（验证：`HomeImport.test.tsx`, `HomeResumeSelection.test.tsx`, `HomeAuthGate.test.tsx` PASS）
- [x] 4.2 Parse route handoff still carries `resumeId`, but reload/list re-entry can recover binding from `TargetJob.resumeId` instead of transient route-only state（验证：Workspace focused tests and `InterviewContext` merge tests PASS）
- [x] 4.3 BDD-Gate: `E2E.P0.015` import request contract remains aligned with allowed `resumeId` and privacy redlines（验证：focused equivalent Home import tests + `make validate-fixtures` PASS）

## Phase 5: Unified plan detail remediation

- [x] 5.1 UI truth source and formal copy rename the Parse preview to `面试规划详情 / 面试上下文确认` while preserving first-import loading（验证：`ui-design/src/screens-p0-complete.jsx`, `docs/ui-design/module-job-workspace.md`, `frontend/src/app/i18n/locales/{zh,en}.ts`, `frontend/tests/pixel-parity/parse.spec.ts` PASS）
- [x] 5.2 `route=parse` ready state and `route=workspace` with `targetJobId` render the same Parse-derived detail DOM, readonly resume binding and Start action; workspace no-context still renders `WorkspacePlanList`（验证：`ParseScreen.test.tsx`, `ParseEdit.test.tsx`, `ParseResumeBinding.test.tsx`, `WorkspaceScreen.test.tsx`, `WorkspaceEmptyState.test.tsx` PASS）
- [x] 5.3 Shared detail navigation uses declared `TargetJob.currentPracticePlanId` / `TargetJob.resumeId` without fabricating `plan-${targetJobId}` or `resume-unbound`, and retired independent workspace detail anchors are covered by negative tests（验证：`frontend/src/app/navigation/interviewContext.ts`, `interviewContext.test.ts`, `WorkspaceHandoff.test.tsx`, `frontend/tests/pixel-parity/workspace.spec.ts` PASS）
- [x] 5.4 BDD-Gate: `E2E.P0.016` and `E2E.P0.018` prove first-import detail and workspace list re-entry land on the same unified detail mother page（验证：scenario trigger/verify PASS）

## Phase 6: Readonly plan detail simplification

- [x] 6.1 UI truth source and formal copy make Parse success detail a readonly context receipt with only Start interview as the success footer action（验证：`node --test ui-design/ui-design-contract.test.mjs` PASS；focused Playwright parse/workspace PASS）
- [x] 6.2 Parse success detail removes field edit state, requirement toggles, hidden-signal remove controls, resume picker / create fallback, success Re-parse, Save plan and Cancel controls（验证：`ParseScreen.test.tsx`, `ParseEdit.test.tsx`, `ParseResumeBinding.test.tsx`, `ParseAuthGate.test.tsx` PASS）
- [x] 6.3 Start interview uses the saved `targetJobId/resumeId/roundId/currentPracticePlanId` snapshot and must not call `updateTargetJob`; missing bound resume blocks Start without offering in-place binding（验证：focused Parse tests + generated client spy PASS）
- [x] 6.4 BDD-Gate: `E2E.P0.016` proves readonly receipt and direct Start handoff; `E2E.P0.018` proves workspace list re-entry lands on the same readonly detail mother page（验证：P0.016 trigger/verify PASS；focused workspace pixel parity PASS）
- [x] 6.5 Repo gates pass after doc/code/test changes（验证：context validation, sync-doc-index, docs-check, diff whitespace check, touched frontend tests/typecheck PASS）
