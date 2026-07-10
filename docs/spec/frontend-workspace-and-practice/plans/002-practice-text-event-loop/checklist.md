# 002 — Practice Text Event Loop Checklist

> **版本**: 1.16
> **状态**: active
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Current Contract Snapshot

- [x] `PracticeScreen` renders the current real-interview shell with TopBar, SessionMap, center content and global Finish CTA.
- [x] `usePracticeSessionLoader` 只通过 generated `getPracticeSession` 读取 session，覆盖 loading、data、refresh、404 和 error 状态。
- [x] `usePracticeEvents` 只通过 generated `appendSessionEvent` 提交 answer / hint / pause / resume，body 含 `clientEventId`，request 不带 `Idempotency-Key`；当前 UI 不提供 skip 正向路径。
- [x] AssistantAction renderer 覆盖 `ask_question / ask_follow_up / show_hint / session_wait / session_completed`，provenance 只进 AI transparency UI。
- [x] `PracticeScreen` 始终提供会话内可选 hint；`baseline / retry_current_round / next_round` 和 out-of-scope strict input 对显隐无副作用。
- [x] `useCompletePracticeSession` 只通过 generated `completePracticeSession` 完成会话，body 只含 `clientCompletedAt`，side-effect request 带 `Idempotency-Key` 并处理 replay / mismatch / 5xx / StrictMode 防抖。
- [x] `buildPracticeHandoffParams` 使用 `resumeId` 与稳定 owner IDs handoff 到 `generating`，不携带 answer / question / hint / prompt / model provenance。
- [x] voice turn 只允许出现在 `hooks/usePracticeVoiceTurn.ts`；text event loop 不调用 `getFeedbackReport`、不直接轮询 report、不绕过 generated client。
- [x] BDD-Gate: `E2E.P0.044` assisted happy path、`E2E.P0.045` mode policy、`E2E.P0.046` failure recovery、`E2E.P0.047` completion handoff 场景资产和 verify scripts 齐全。

## Verification

- [x] `validate_context.py frontend-workspace-and-practice/002 frontend`
- [x] Focused practice Vitest suite covers screen, hooks, components, handoff utils and privacy gates.
- [x] `pnpm --filter @easyinterview/frontend exec tsc --noEmit`
- [x] `make validate-fixtures`
- [x] P0.044-P0.047 scenario scripts execute `setup -> trigger -> verify -> cleanup`.
- [x] Owner wording grep has no pre-D20 resume binding field, placeholder voice surface token, stepwise build narrative, plan-reservation note, or out-of-owner generated-client positive surface.
- [x] `sync-doc-index --check`
- [x] `make docs-check`
- [x] `git diff --check`

## Phase 6: Real-interview session simplification

- [x] 6.1 UI truth source: 更新 `docs/ui-design/module-practice-review.md` 和 `ui-design/src/screen-practice.jsx`，当前真实面试 UI 不包含独立辅助信息栏、语音分析、语音转文字、跳过、会话内本地 persona switch、用户可见 strict switch，并新增电话模式字幕、切断、重新开始；验证: `node --test ui-design/ui-design-contract.test.mjs` + focused source grep。
  <!-- verified: 2026-07-09 method=tdd-red-green command="node --test ui-design/ui-design-contract.test.mjs" grep="rg -n forbidden terms ui-design/src/screen-practice.jsx ui-design/canvas.html" -->
- [x] 6.2 Frontend runtime: 当前 runtime 不 materialize 独立辅助信息栏、语音分析卡、dictation/skip controls、会话内本地 persona switch、strict switch 和 voice expression metrics；结束 CTA 位于全局 topbar；验证: focused `PracticeScreen` / `InputBar` / `TopBar` / `practiceModeSwitch` / negative Vitest。
  <!-- verified: 2026-07-09 method=tdd-red-green commands="corepack pnpm --filter @easyinterview/frontend test src/app/screens/practice/PracticeScreen.test.tsx ... outOfScopeNegative.test.ts (13 files, 61 tests); corepack pnpm --filter @easyinterview/frontend exec tsc --noEmit" -->
- [x] 6.3 Contract/backend: 正向 OpenAPI fixtures、generated artifacts、frontend hooks、backend service/handler tests 和 scenario scripts 不把 `turn_skipped` 作为用户动作；验证: `make codegen-openapi`、`make lint-openapi`、`make validate-fixtures`、focused backend practice tests。
  <!-- verified: 2026-07-09 method=tdd-red-green commands="make codegen-openapi; make lint-openapi; make validate-fixtures; python3 scripts/lint/migrations_lint.py --repo-root .; go test ./backend/internal/practice ./backend/internal/api/practice ./backend/internal/store/practice ./backend/cmd/api -run 'TestSessionEventService|TestTurnStatus|TestAppendSessionEvent|TestServiceAppliesHint|TestApplyHint|TestE2EP0039|TestE2EP0048|TestE2EP0049|TestE2EP0050|TestSQLRepositoryAppendSessionEvent|TestSQLRepositoryRecordPracticeVoiceTurn|TestMarshalAppendEvent' -count=1" -->
- [x] 6.4 Phone mode: 用户可见 `voice` 文案改为 `电话模式 / Phone`，phone surface 支持显示字幕、切断和重新开始，并隐藏 `开始录音` / `提交本轮` 主流程；验证: focused phone-mode Vitest + pixel parity practice spec。
  <!-- verified: 2026-07-09 commands="node --test ui-design/ui-design-contract.test.mjs; corepack pnpm --filter @easyinterview/frontend test src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx src/app/screens/practice/__tests__/practiceModeSwitch.test.tsx src/app/screens/practice/PracticeScreen.test.tsx; ./test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/setup.sh && ./test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/trigger.sh && ./test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/verify.sh" result="PASS" -->
- [x] 6.5 BDD-Gate: 更新并执行 `E2E.P0.044`-`E2E.P0.047` wrapper，覆盖 text/phone、无独立右侧信息栏、无转写、无跳过、无语音分析、轮次面试官和 generating handoff。
  <!-- verified: 2026-07-09 commands="./test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/setup.sh && ./test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/trigger.sh && ./test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/verify.sh; ./test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/setup.sh && ./test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/trigger.sh && ./test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/verify.sh; ./test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/setup.sh && ./test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/trigger.sh && ./test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/verify.sh; ./test/scenarios/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/scripts/setup.sh && ./test/scenarios/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/scripts/trigger.sh && ./test/scenarios/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/scripts/verify.sh" result="PASS" -->
- [x] 6.6 Real environment close-loop: 启动/验证本地真实前后端环境，浏览器进入 practice 文本与电话模式真实页面，完成核心会话闭环并保存截图证据；验证: scenario-env verify + browser screenshot artifacts。
  <!-- verified: 2026-07-09 commands="./test/scenarios/env-redeploy.sh all; ./test/scenarios/env-verify.sh; Playwright real browser closed-loop" result="PASS" artifacts=".test-output/real-practice-session/practice-text-real-closed-loop.png .test-output/real-practice-session/practice-phone-real-default.png .test-output/real-practice-session/practice-phone-real-captions.png .test-output/real-practice-session/practice-generating-real-stable.png .test-output/real-practice-session/real-browser-closed-loop-result.json" -->
- [x] 6.7 Closeout gates: `validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check`、current-boundary zero-reference search 通过。
  <!-- verified: 2026-07-09 commands="python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-workspace-and-practice/plans/002-practice-text-event-loop/context.yaml --target frontend; python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check; make docs-check; git diff --check; python3 scripts/lint/backend_practice_out_of_scope.py --repo-root . --phase all; make validate-fixtures; make lint-openapi; corepack pnpm --filter @easyinterview/frontend exec tsc --noEmit; runtime current-boundary rg" result="PASS" -->

## Phase 7: constant-only assistance hook removal

- [x] 7.1 Add a scoped RED gate proving the unconsumed hook and its P0.045 trigger path still exist.
  <!-- verified: 2026-07-10 method=constant-assistance-hook-source-trigger-red evidence="Focused practice boundary test failed with exactly two offender labels: practice-runtime and p0-045-trigger; the other three negative gates passed." -->
- [x] 7.2 Delete the hook/test pair and reconcile owner/test/context plus P0.045 assets with rendered policy tests.
  <!-- verified: 2026-07-10 method=constant-assistance-hook-removal evidence="Deleted both files and removed the P0.045 trigger/verify/expected markers. Current owner/test/context now point to rendered goal/hint/mode tests; scoped frontend/scenario symbol search is empty and focused policy tests pass 4 files/13 tests." -->
- [x] 7.3 Run focused practice tests, P0.045 setup/trigger/verify/cleanup, full frontend/typecheck, owner/product contexts, docs, diff and pruning gates.
  <!-- verified: 2026-07-10 method=constant-assistance-hook-removal evidence="P0.045 passes real-mode 1/1 plus rendered policy 6 files/18 tests; practice passes 24 files/107 tests; full frontend passes 137 files/836 tests with zero React update warning; typecheck and scenario trigger-path pytest pass. Owner/product contexts and docs/index/link/diff/pruning gates pass with real_residuals=0." -->

## Phase 8: production test-only handoff inspector removal

- [x] 8.1 Add a scoped source RED assertion for the runtime handoff inspector with no production consumer.
  <!-- verified: 2026-07-10 method=test-only-handoff-inspector-source-red evidence="Focused practice boundary test failed only on utils/practiceHandoffParams.ts; the other four boundary gates passed and repository inventory showed only the unit test imported the inspector." -->
- [x] 8.2 Delete `FORBIDDEN_KEYS` / `findForbiddenHandoffKeys` and replace helper-self-tests with direct complete forbidden-key assertions on `buildPracticeHandoffParams` output.
  <!-- verified: 2026-07-10 method=test-only-handoff-inspector-removal evidence="Deleted both production symbols and the injected-helper self-test. The real output test now checks all 11 forbidden keys directly and no longer uses jsdom; boundary/handoff/privacy pass 3 files/9 tests and scoped runtime symbol inventory is empty." -->
- [x] 8.3 Run focused handoff/privacy tests, P0.047, practice/full frontend tests, typecheck, owner/product contexts, docs, diff and pruning gates.
  <!-- verified: 2026-07-10 method=test-only-handoff-inspector-removal evidence="Focused boundary/handoff/privacy passes 3 files/9 tests; P0.047 passes real-mode 1/1 plus 5 files/14 tests; full frontend passes 137 files/836 tests with zero React update warning and typecheck. Owner/product contexts and docs/index/link/diff/pruning gates pass with real_residuals=0." -->

## Phase 9: current TopBar copy parity

- [x] 9.1 Formal Practice TopBar consumes typed `practice.toolbar.questionTag`, `pause` and `resume` messages and renders the same visible copy as `ui-design/src/screen-practice.jsx`; verify focused TopBar/Practice/pause tests, locale reachability, UI parity and owner/global gates.
  <!-- verified: 2026-07-10 method=practice-topbar-copy-parity evidence="Behavior red showed the formal TopBar rendered only 1/5 and glyph-only pause controls. Green renders typed Question/Pause/Resume copy; focused Practice tests, 46-file/239-test owner directories, full frontend 137 files/841 tests, typecheck/build, 35 UI contracts, Practice Playwright 11 pass plus 1 expected desktop skip, P0.045 and both owner/product contexts pass." -->
