# 001 Home + JD Import + Parse Checklist

> **版本**: 2.23
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Phase 1: Home 当前入口

- [x] 1.1 Home 源级复刻当前 `ui-design/src/screen-home.jsx::HomeScreen`：Hero label/title、JD 输入卡、输入卡底部 upload/URL source actions、ready 简历下拉框、创建简历入口、提交区、最近 3 张模拟面试卡片和 More handoff。
- [x] 1.2 Home 使用 generated client 调 `listResumes`、`listTargetJobs`、`createUploadPresign`、`importTargetJob`；paste/file/URL source discriminator、side-effect idempotency key、错误态和 pending import continuation 均有 focused Vitest 覆盖。
- [x] 1.3 Home import 前必须显式选择 ready 简历；成功进入 `parse` 时 params 携带真实 `resumeId`。
- [x] 1.4 BDD-Gate: `E2E.P0.014` 覆盖默认渲染、empty/one/twelve-plus fixtures、ready filter/sort/3-card cap、More/quick-start handoff、英文 i18n 和 source/resume/submit layout。
- [x] 1.5 BDD-Gate: `E2E.P0.015` 覆盖 paste/upload/URL import、4xx/failed path、privacy gate、generated client request contract 和 real-mode generated-client preflight。

## Phase 2: Historical pre-readonly Parse confirmation and handoff

- [x] 2.1 Historical pre-Phase 6 Parse parity covered loading, preview, failed state, editable basics, requirements, hidden signals, round assumptions, resume binding and footer actions; Phase 6 now supersedes success preview with a readonly receipt.
- [x] 2.2 Historical generated-client gates covered `getTargetJob`, `listResumes` and `updateTargetJob` contract behavior; current Parse success detail uses `getTargetJob` / `listResumes` / practice handoff and must not consume `updateTargetJob`.
- [x] 2.3 Current Parse success detail ignores route-only `resumeId` for binding, only displays the saved TargetJob binding, and disables Start when that binding is missing.
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
- [x] 5.3 Shared detail navigation uses declared `TargetJob.currentPracticePlanId` / `TargetJob.resumeId` without fabricating `plan-${targetJobId}` or `resume-unbound`, and out-of-scope independent workspace detail anchors are covered by negative tests（验证：`frontend/src/app/navigation/interviewContext.ts`, `interviewContext.test.ts`, `frontend/src/app/screens/workspace/WorkspaceEmptyState.test.tsx`, `frontend/tests/pixel-parity/workspace.spec.ts` PASS）
- [x] 5.4 BDD-Gate: `E2E.P0.016` and `E2E.P0.018` prove first-import detail and workspace list re-entry land on the same unified detail mother page（验证：scenario trigger/verify PASS）

## Phase 6: Readonly plan detail simplification

- [x] 6.1 UI truth source and formal copy make Parse success detail a readonly context receipt with only Start interview as the success footer action（验证：`node --test ui-design/ui-design-contract.test.mjs` PASS；focused Playwright parse/workspace PASS）
- [x] 6.2 Parse success detail removes field edit state, requirement toggles, hidden-signal remove controls, resume picker / create fallback, success Re-parse, Save plan and Cancel controls（验证：`ParseScreen.test.tsx`, `ParseEdit.test.tsx`, `ParseResumeBinding.test.tsx`, `ParseAuthGate.test.tsx` PASS）
- [x] 6.3 Start interview uses the saved `targetJobId/resumeId/roundId/currentPracticePlanId` snapshot and must not call `updateTargetJob`; missing bound resume blocks Start without offering in-place binding（验证：focused Parse tests + generated client spy PASS）
- [x] 6.4 BDD-Gate: `E2E.P0.016` proves readonly receipt and direct Start handoff; `E2E.P0.018` proves workspace list re-entry lands on the same readonly detail mother page（验证：P0.016 trigger/verify PASS；focused workspace pixel parity PASS）
- [x] 6.5 Repo gates pass after doc/code/test changes（验证：context validation, sync-doc-index, docs-check, diff whitespace check, touched frontend tests/typecheck PASS）

## Phase 7: LLM-derived round assumptions shared data binding

- [x] 7.1 Historical UI truth source and owner docs first moved TargetJob round assumptions off local-only copy; Phase 8 supersedes the string-only shape with `TargetJob.summary.interviewRounds[]` across Parse, Home recent cards, Workspace handoff and shared navigation context（当前验证见 Phase 8.1-8.5).
- [x] 7.2 Focused Parse tests prove `parse-round-*` cards render backend-provided structured rounds and do not show static locale focus when structured rounds exist（当前验证见 8.3).
- [x] 7.3 Focused Home/navigation tests prove `home-recent-mock-rail-*` and `interviewContextFromTargetJob` consume the same backend-provided structured rounds instead of a local `DEFAULT_ROUNDS` / static round name（当前验证见 8.3).
- [x] 7.4 Frontend implementation uses a shared TargetJob round assumption mapper without changing the Parse layout, Home card layout, Workspace list layout or Start handoff（验证：`frontend/src/app/interview-context/roundAssumptions.ts`, focused frontend tests, `pnpm --dir frontend typecheck` PASS).
- [x] 7.5 BDD-Gate: `E2E.P0.016` / focused equivalent proves readonly detail and related Home recent surface consume saved backend structured rounds for round assumptions and reject fixed positive-path fallback（当前验证见 8.5).
- [x] 7.6 Repo gates pass after doc/code/test changes（验证：context validation PASS; `node --test ui-design/ui-design-contract.test.mjs` PASS; `pnpm --dir frontend typecheck` PASS; `pnpm --dir frontend test` PASS, 140 files / 825 tests; `E2E.P0.016` setup/trigger/verify/cleanup PASS; `sync-doc-index --check`, `make docs-check`, `git diff --check`, `make lint-core-loop-pruning-surface` PASS).

## Phase 8: Structured LLM-derived interview rounds

- [x] 8.1 OpenAPI / prompt / fixture contract defines `TargetJob.summary.interviewRounds[]` with 2~5 LLM-derived rounds, each carrying `sequence`, `type`, `name`, `durationMinutes` and `focus`; prompt explicitly instructs inference from JD, role seniority, company/industry nature, team/business context, hiring-process hints and common interview practices, and generated Go/TS artifacts are refreshed（验证：`make lint-prompts`, `cd backend && go test ./internal/targetjob -run TestTargetImportPromptMatchesParseResponseSchema -count=1`, `make codegen-openapi`, `make lint-openapi`, `make validate-fixtures` PASS).
- [x] 8.2 Backend parser validates and persists 2~5 structured `interviewRounds[]` from `target.import.parse` without fabricating fixed 4-round defaults（验证：`cd backend && go test ./internal/targetjob -run 'TestParseExecutor|TestDecodeParseResponse|TestTargetImportParse' -count=1` focused PASS).
- [x] 8.3 Frontend Parse/Home/navigation consume structured rounds with variable count, type/name and duration from `summary.interviewRounds[]`; fixed strings such as `HR 初筛 · 20m` / `技术一面 · 45m` are not used when structured rounds exist（验证：`cd frontend && pnpm test --run src/app/navigation/interviewContext.test.ts src/app/screens/home/MockInterviewCard.test.tsx src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/parse/ParseScreen.test.tsx src/app/screens/parse/ParseEdit.test.tsx src/api/targetJob.realApiMode.test.ts`; `cd frontend && pnpm typecheck` PASS).
- [x] 8.4 UI truth source and docs define structured LLM rounds across Parse and Home recent rail（验证：`node --test ui-design/ui-design-contract.test.mjs`; `make sync-fixtures-from-prototype`; `make validate-fixtures` PASS).
- [x] 8.5 BDD-Gate: `E2E.P0.016` proves readonly detail and related Home recent surface consume structured backend rounds for 2~5 count/type/duration/focus, with no fixed 4-round template in positive structured data paths; Playwright attaches the readonly-detail screenshot and emits `screenshotBytes=` marker（验证：`test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/trigger.sh && test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/verify.sh` PASS).
- [x] 8.6 Repo gates pass after structured round contract changes（验证：`python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/context.yaml --docs-root docs --target frontend`; `cd frontend && pnpm typecheck`; focused frontend tests; backend targetjob focused tests; `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`; `make docs-check`; `git diff --check`; `make lint-core-loop-pruning-surface` PASS).

## Phase 9: Recent card fixed grid and workspace fusion

- [x] 9.1 UI truth source defines Home recent cards and workspace plan-list cards as one shared card body with fixed max-width grid（验证：`docs/ui-design/jd-resume-management.md`, `docs/ui-design/module-job-workspace.md`, `ui-design/src/screen-home.jsx`, `ui-design/src/screen-workspace.jsx`, `node --test ui-design/ui-design-contract.test.mjs` PASS）
- [x] 9.2 Formal `MockInterviewCard` supports Home default testids plus workspace-owned card/body/rail/footer testids and optional footer CTA（验证：`pnpm --filter @easyinterview/frontend test src/app/screens/home/MockInterviewCard.test.tsx` PASS）
- [x] 9.3 Home recent and workspace list focused tests reject `1fr` stretching and verify workspace mini round rail + footer CTA（验证：`pnpm --filter @easyinterview/frontend test src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceEmptyState.test.tsx` PASS）

## Phase 10: Home recent shared action card

- [x] 10.1 UI truth source defines Home recent cards as the shared Interview list action card with `立即面试` and without delete controls（验证：`docs/ui-design/jd-resume-management.md`, `docs/ui-design/module-job-workspace.md`, `ui-design/src/screen-home.jsx`）
- [x] 10.2 Formal `MockInterviewCard` supports quick-start action props and Home passes no delete action（验证：`MockInterviewCard.test.tsx`, `HomeRecentMocks.test.tsx`）
- [x] 10.3 Home recent quick-start calls shared generated practice handoff with structured `roundId/roundName`, and card-body click remains planning-detail navigation（验证：`HomeRecentMocks.test.tsx` PASS）
- [x] 10.4 Browser screenshot acceptance captures Home recent card with `立即面试` and no delete icon（验证：`.test-output/screenshots/home-recent-action-card.png`）
- [x] 10.5 Home recent requests `listTargetJobs(analysisStatus=ready)` and filters failed / processing / queued / blank-title dirty records before rendering cards（验证：`HomeRecentMocks.test.tsx` PASS）

## Phase 11: P0.014 executable-evidence reconciliation

- [x] 11.1 BDD-Gate: P0.014 README/seed/expected 与 trigger 保持同一 Vitest-only 证据面，不宣称 TopBar、theme、mobile、build、Playwright 或 live-backend 覆盖；场景资产合同先红后绿并执行四段脚本。
  <!-- verified: 2026-07-10 method=p0014-executable-evidence-reconciliation evidence="Red asset contract exposed Playwright/chromium/dist/TopBar/theme/mobile/real-backend claims absent from the runner. Scenario assets now describe the stub-fetch generated-client test and five Home Vitest files only. HomeScreen passes 9 tests; P0.014 setup/trigger/verify/cleanup passes with 1 generated-client test and 34 Home tests." -->

## Phase 12: Pending-import test API removal

- [x] 12.1 RED/GREEN: Home source gate detects and then rejects the production `clearPendingImportSourcesForTests` export.<!-- verified: 2026-07-10 method=vitest-red-green evidence="RED failed only on the reset export after path validation; GREEN passed all 6 HomeAuthGate tests after deletion." -->
- [x] 12.2 Delete the redundant Home auth teardown and pass focused Home auth tests plus frontend typecheck without a replacement reset API.<!-- verified: 2026-07-10 method=vitest+scenario+typecheck evidence="Home passed 8 files/64 tests; P0.015 passed generated-client 1/1, Home/Parse 9 files/56 tests, build and Playwright 2/2; typecheck and source inventory passed." -->

## Phase 13: Current fixture inventory wording

- [x] 13.1 Keep the BDD closeout fixture gate on the current 37 operations; verify OpenAPI inventory, fixture validation, owner contexts and docs/diff/pruning gates without changing runtime or scenario assets.
  <!-- verified: 2026-07-10 method=current-openapi-inventory-wording evidence="OpenAPI inventory and fixture validation report 10 tags, 37 operations and 37 fixtures. Home BDD current-contract search has no 35/36 count, owner context passes, and runtime/scenario assets are unchanged." -->

## Phase 14: Home copy-table orphan cleanup

- [x] 14.1 删除 UI prototype、zh/en locale 与 locale self-test 中无渲染 consumer 的 `uploadSourceSub`；验证 Home/UI contract、locale reachability、owner contexts 与 docs/diff/pruning gates。
  <!-- verified: 2026-07-10 method=home-copy-table-orphan-removal evidence="Deleted uploadSourceSub from both prototype language branches, both formal locale catalogs and the locale self-test requirement. Scoped runtime/prototype search is zero; locale reachability, 35 UI contracts, frontend tests/typecheck/build, Home/product contexts and docs/diff/pruning gates pass with no visible Home DOM or behavior change." -->

## Phase 15: MiniRoundRail prototype call-surface pruning

- [x] 15.1 新增 Home rail 参数消费 contract，并先红证明 `MiniRoundRail` 与调用方仍保留未读取的 `lang`。
  <!-- verified: 2026-07-10 method=home-mini-round-rail-red evidence="UI contract ran 42 tests: the new structured-round dependency contract failed on the existing MiniRoundRail.lang parameter while the prior 41 tests passed; retained assertions pin round count, names, durations and current-index highlighting." -->
- [x] 15.2 删除 rail 的零读取 `lang` 形参与调用方传参；验证：AST `MiniRoundRail` 参数消费 inventory 归零，结构化轮次和 current-index 高亮保持原样。
  <!-- verified: 2026-07-10 method=home-mini-round-rail-green evidence="Removed only MiniRoundRail.lang and its single caller argument; MockInterviewCard.lang remains for visible action copy. UI contract passes 42/42 and Babel binding inventory reports railUnread=[] while retaining structured rounds and currentIndex assertions." -->
- [x] 15.3 运行 UI contract、focused Home、P0.014/P0.016、静态浏览器 Home rail、full frontend、typecheck/build、owner contexts 与 docs/diff/pruning gates。
  <!-- verified: 2026-07-10 method=home-mini-round-rail-regression-closeout evidence="UI contract passes 42/42 and focused Home passes 3 files/27 tests. P0.014 setup/trigger/verify/cleanup passes generated-client 1 plus Home 34; P0.016 passes generated-client 1, focused 37, build and desktop/mobile Playwright 4. Static browser renders signed-in Home with three recent cards and four structured round names/durations, no errors and only 200/304 requests. Full frontend passes 137 files/841 tests and typecheck passes. Both owner contexts and diff/pruning gates pass with real_residuals=0. No scenario environment restart or data cleanup occurred." -->

## Phase 16: Home/Parse real-backend verifier convergence

- [x] 16.1 共享 helper 支持可配置 owner test 文件，P0.014/P0.015/P0.016 删除内联 real-mode/Vitest 通用解析与冗余 PASS grep；验证 helper/caller RED/GREEN、三个 wrapper 生命周期、owner/product contexts 与 docs/diff/pruning gates。
  <!-- red: 2026-07-10 method=scenario-env-contract evidence="The focused contract suite failed only the two new checks while the prior 16 tests passed: the helper ignored the requested targetJob owner test marker, and all three Home/Parse callers still contained inline real-mode and generic Vitest summary parsing." -->
  <!-- verified: 2026-07-10 method=home-parse-real-backend-verifier-convergence evidence="The helper now accepts an optional owner test marker while preserving frontendOwners as its default. P0.014/P0.015/P0.016 pass targetJob.realApiMode.test.ts explicitly and retain their scenario-specific test, privacy, browser and out-of-scope assertions with no duplicate generic parsing. Contract tests pass 18/18 and all touched shell scripts pass bash syntax. Complete wrapper lifecycles pass: P0.014 real mode 1 plus Home 34; P0.015 real mode 1 plus Home/Parse 56, build and Playwright 2; P0.016 real mode 1 plus focused 37, build and Playwright 4; default-argument P0.018 real mode 1 plus focused 57. Both owner contexts, git diff check and pruning surface pass with real_residuals=0. No Bug or retrospective report was needed because scenario behavior and coverage did not change. No environment restart or data cleanup occurred." -->
