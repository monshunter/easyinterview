# 001 Home + JD Import + Parse Checklist

> **版本**: 2.29
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: Home 当前入口

- [x] 1.1 Historical Phase 1 delivered the original Home shell, JD input card, ready 简历下拉框、创建简历入口、提交区、最近 3 张模拟面试卡片和 More handoff；Phase 18 owns the current paste-only intake surface.
- [x] 1.2 Historical Phase 1 delivered the original generated-client import baseline, idempotency、错误态和 pending continuation；Phase 18 supersedes the current request and continuation contract.
- [x] 1.3 Historical Phase 1 route carried `resumeId` after explicit ready-resume selection；Phase 20 supersedes the route to `targetJobId` only while the POST body still includes the selected `resumeId`.
- [x] 1.4 Historical BDD-Gate: `E2E.P0.014` 覆盖默认渲染、empty/one/twelve-plus fixtures、ready filter/sort/3-card cap、More/quick-start handoff 和英文 i18n；Phase 18 reruns the current layout and screenshot gate.
- [x] 1.5 Historical BDD-Gate: `E2E.P0.015` 覆盖原始 import、4xx/failed path、privacy gate、generated client request contract 和 real-mode generated-client preflight；Phase 18 replaces its current intake assertions.

## Phase 2: Historical pre-readonly Parse confirmation and handoff

- [x] 2.1 Historical pre-Phase 6 Parse parity covered loading, preview, failed state, editable basics, requirements, hidden signals, round assumptions, resume binding and footer actions; Phase 6 now supersedes success preview with a readonly receipt.
- [x] 2.2 Historical generated-client gates covered Parse `getTargetJob`, `listResumes` and `updateTargetJob` behavior；Phase 20 supersedes ready rendering to Workspace detail, which uses one `getTargetJob`, no `listResumes` and no `updateTargetJob`.
- [x] 2.3 Historical readonly detail ignored route-only `resumeId`；Phase 20 removes that route param entirely, reads the saved TargetJob binding in Workspace detail and disables Start when missing.
- [x] 2.4 Historical Save plan / workspace auto-start handoff was covered before readonly simplification; current success path has no Save plan action and Start enters practice directly.
- [x] 2.5 BDD-Gate: `E2E.P0.016` keeps the historical import-to-detail lineage and now covers readonly receipt, direct Start handoff, auth continuation and privacy checks.

## Phase 3: 收口验证

- [x] 3.1 `validate_context.py frontend-home-job-picks-and-parse/001 frontend` 通过。
- [x] 3.2 Focused Home/Parse Vitest、frontend typecheck 与 `make validate-fixtures` 通过。
- [x] 3.3 `E2E.P0.014`、`E2E.P0.015`、`E2E.P0.016` 的 `setup -> trigger -> verify -> cleanup` 通过。
- [x] 3.4 `sync-doc-index --check`、`make docs-check`、`git diff --check` 和 `make lint-core-loop-pruning-surface` 通过。

## Phase 4: Import resume binding remediation

- [x] 4.1 Historical import variants included the selected `resumeId` in generated `importTargetJob` request bodies；Phase 18 preserves the binding in the flattened paste-only body（历史验证：`HomeImport.test.tsx`, `HomeResumeSelection.test.tsx`, `HomeAuthGate.test.tsx` PASS）
- [x] 4.2 Historical Parse route handoff carried `resumeId`；Phase 20 supersedes it with targetJobId-only routing and authoritative `TargetJob.resumeId` recovery（历史验证：Workspace focused tests and `InterviewContext` merge tests PASS）
- [x] 4.3 BDD-Gate: `E2E.P0.015` import request contract remains aligned with allowed `resumeId` and privacy redlines（验证：focused equivalent Home import tests + `make validate-fixtures` PASS）

## Phase 5: Unified plan detail remediation

- [x] 5.1 Historical UI work renamed the Parse-derived ready visual to `面试规划详情 / 面试上下文确认` while preserving first-import loading；Phase 20 keeps the visual only under Workspace detail（历史验证：`ui-design/src/screens-p0-complete.jsx`, docs/locales/parity PASS）
- [x] 5.2 Historical parse/workspace routes rendered the same detail DOM；Phase 20 supersedes this so Parse ready replace-navigates and only `route=workspace` with `targetJobId` renders readonly resume binding and Start, while query-free workspace renders `WorkspacePlanList`（历史验证：focused Parse/Workspace tests PASS）
- [x] 5.3 Shared detail navigation uses declared `TargetJob.currentPracticePlanId` / `TargetJob.resumeId` without fabricating `plan-${targetJobId}` or `resume-unbound`, and out-of-scope independent workspace detail anchors are covered by negative tests（验证：`frontend/src/app/navigation/interviewContext.ts`, `interviewContext.test.ts`, `frontend/src/app/screens/workspace/WorkspaceEmptyState.test.tsx`, `frontend/tests/pixel-parity/workspace.spec.ts` PASS）
- [x] 5.4 BDD-Gate: `E2E.P0.016` and `E2E.P0.018` prove first-import detail and workspace list re-entry land on the same unified detail mother page（验证：scenario trigger/verify PASS）

## Phase 6: Readonly plan detail simplification

- [x] 6.1 UI truth source and formal copy make the shared success detail a readonly context receipt with only Start interview as the footer action；Phase 20 locates it only at Workspace detail（历史验证：UI contract + focused parity PASS）
- [x] 6.2 Workspace success detail removes field edit state, requirement toggles, hidden-signal remove controls, resume picker / create fallback, success Re-parse, Save plan and Cancel controls（历史 Parse-derived component tests PASS）
- [x] 6.3 Start interview uses the saved `targetJobId/resumeId/roundId/currentPracticePlanId` snapshot and must not call `updateTargetJob`; missing bound resume blocks Start without offering in-place binding（验证：focused Parse tests + generated client spy PASS）
- [x] 6.4 BDD-Gate: `E2E.P0.016` proves readonly receipt and direct Start handoff; `E2E.P0.018` proves workspace list re-entry lands on the same readonly detail mother page（验证：P0.016 trigger/verify PASS；focused workspace pixel parity PASS）
- [x] 6.5 Repo gates pass after doc/code/test changes（验证：context validation, sync-doc-index, docs-check, diff whitespace check, touched frontend tests/typecheck PASS）

## Phase 7: LLM-derived round assumptions shared data binding

- [x] 7.1 Historical UI truth source and owner docs first moved TargetJob round assumptions off local-only copy; Phase 8/20 use `TargetJob.summary.interviewRounds[]` across Workspace detail, Home recent cards and shared navigation context（当前验证见 Phase 8.1-8.5).
- [x] 7.2 Focused shared-detail tests prove round cards render backend-provided structured rounds and do not show static locale focus when structured rounds exist（当前验证见 8.3).
- [x] 7.3 Focused Home/navigation tests prove `home-recent-mock-rail-*` and `interviewContextFromTargetJob` consume the same backend-provided structured rounds instead of a local `DEFAULT_ROUNDS` / static round name（当前验证见 8.3).
- [x] 7.4 Frontend implementation uses a shared TargetJob round assumption mapper without changing the Parse layout, Home card layout, Workspace list layout or Start handoff（验证：`frontend/src/app/interview-context/roundAssumptions.ts`, focused frontend tests, `pnpm --dir frontend typecheck` PASS).
- [x] 7.5 BDD-Gate: `E2E.P0.016` / focused equivalent proves readonly detail and related Home recent surface consume saved backend structured rounds for round assumptions and reject fixed positive-path fallback（当前验证见 8.5).
- [x] 7.6 Repo gates pass after doc/code/test changes（验证：context validation PASS; `node --test ui-design/ui-design-contract.test.mjs` PASS; `pnpm --dir frontend typecheck` PASS; `pnpm --dir frontend test` PASS, 140 files / 825 tests; `E2E.P0.016` setup/trigger/verify/cleanup PASS; `sync-doc-index --check`, `make docs-check`, `git diff --check`, `make lint-core-loop-pruning-surface` PASS).

## Phase 8: Structured LLM-derived interview rounds

- [x] 8.1 OpenAPI / prompt / fixture contract defines `TargetJob.summary.interviewRounds[]` with 2~5 LLM-derived rounds, each carrying `sequence`, `type`, `name`, `durationMinutes` and `focus`; prompt explicitly instructs inference from JD, role seniority, company/industry nature, team/business context, hiring-process hints and common interview practices, and generated Go/TS artifacts are refreshed（验证：`make lint-prompts`, `cd backend && go test ./internal/targetjob -run TestTargetImportPromptMatchesParseResponseSchema -count=1`, `make codegen-openapi`, `make lint-openapi`, `make validate-fixtures` PASS).
- [x] 8.2 Backend parser validates and persists 2~5 structured `interviewRounds[]` from `target.import.parse` without fabricating fixed 4-round defaults（验证：`cd backend && go test ./internal/targetjob -run 'TestParseExecutor|TestDecodeParseResponse|TestTargetImportParse' -count=1` focused PASS).
- [x] 8.3 Frontend Workspace detail/Home/navigation consume structured rounds with variable count, type/name and duration from `summary.interviewRounds[]`; fixed strings such as `HR 初筛 · 20m` / `技术一面 · 45m` are not used when structured rounds exist（历史 Parse-derived component tests + current Workspace gate）.
- [x] 8.4 UI truth source and docs define structured LLM rounds across Workspace detail and Home recent rail（验证：UI contract/fixtures PASS）.
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

## Phase 17: Parse loading internal-metadata removal

- [x] 17.1 RED-GREEN: prototype source contract rejects the Parse loading model/provider, rubric/prompt/version/hash, provenance and typical-latency footer while preserving four progress steps and responsive layout.
- [x] 17.2 RED-GREEN: formal Parse loading DOM removes the same internal metadata without changing polling, ready mapping or failed recovery.
- [x] 17.3 REGRESSION-GATE: focused Parse tests, UI source contract, typecheck/build and active negative search pass.
  <!-- verified: 2026-07-13 evidence="Parse/Report focused Vitest and desktop/mobile parity PASS; UI contract 58/58; full frontend 112 files/786 tests, typecheck and production build PASS; internal metadata and source locator negative assertions PASS." -->
- [x] 17.4 BDD-Gate: `E2E.P0.015` passes with clean 1440/390 loading screenshots plus DOM/style/bbox/viewport parity evidence.
  <!-- verified: 2026-07-13 method=scenario-run+pixel-parity evidence="E2E.P0.015 PASS; ready-response loading is metadata-free at 1440x900 and 390x844; formal/prototype Parse parity changedRatio desktop=0.000556 mobile=0.001499 after explicit scroll normalization." -->

## Phase 18: Paste-only Home JD intake

- [x] 18.1 RED-GREEN: prototype-first 更新 `ui-design/src/screen-home.jsx` 与 UI source contract，使 Home intake 仅渲染 textarea、ready Resume select 和 CTA；旧 source controls、辅助弹窗、触发 testid 与多入口 copy 必须先红后删，且 Resume 上传入口保持可用。
- [x] 18.2 RED-GREEN: OpenAPI、fixtures 与 generated Go/TS 将 `importTargetJob` public request 收敛为 exact flattened wire `{ rawText, targetLanguage, resumeId }`，拒绝 source discriminator、嵌套 source payload 和非文本 JD intake；Resume upload operation/fixture 不受影响。
  <!-- verified: 2026-07-13 evidence="OpenAPI inventory 24 tests, generator tests, lint-openapi, fixture validator 37 operations and isolated-index codegen-check PASS; generated Go/TS request is flattened; upload fixture retains resume/privacy only." -->
- [x] 18.3 RED-GREEN: formal Home layout/import/i18n tests 证明只存在一个 paste submit path；删除 source controls/modal、source-specific branches、额外 locale keys 和 JD upload-client 调用，不新增 mode enum、compatibility adapter 或不可达 branch。
  <!-- verified: 2026-07-13 evidence="Home source-control/modal tests were RED before deletion; focused frontend 7 files/34 tests and full 112 files/786 tests PASS; typecheck/build PASS; JDAssistModal and unused locale keys are deleted." -->
- [x] 18.4 RED-GREEN: Home auth `pendingAction` 只保存 `opaquePendingImportId`；一次性进程内 vault 保存 `{ rawText, targetLanguage, resumeId, idempotencyKey, expiresAt }` 并在登录成功后原子 consume 一次。正常路径以原 key 提交同一 exact request；refresh/lost、expired、duplicate consume 均不调用 import，清除 action、返回 Home 并显示本地化重新输入提示；route/browser storage/log/telemetry 均不得携带 JD 原文。
  <!-- verified: 2026-07-13 evidence="One-shot pending import vault tests cover normal consume, lost/expired/duplicate consume, same idempotency key, no import on invalid recovery and no raw JD in route or browser storage; focused/full frontend gates PASS." -->
- [x] 18.5 RED-GREEN: backend TargetJob decode/service/store/runner 删除 public 多来源 union、URL fetch/refresh、JD attachment purpose、structured manual-form branch 及 source-specific persistence/event/config，同时保留文本 validation、idempotency、parse failure/retry、privacy 与 resume binding。
  <!-- verified: 2026-07-13 method=focused-backend+real-postgres+scenario evidence="TargetJob package and HTTP scenarios prove exact rawText import, user-scoped idempotency, resume binding, parse failure/recovery and source-free persistence/event/job paths; obsolete urlfetch package and source refresh registration are absent." -->
- [x] 18.6 REGRESSION-GATE: UI contract、Home focused Vitest、OpenAPI lint/fixture/generated drift、backend TargetJob tests、typecheck/build 与 Resume upload owner tests 全部通过。
  <!-- verified: 2026-07-14 method=focused-full+isolated-index-codegen evidence="UI source contract 59/59; Home/Parse focused Vitest 75/75; TargetJob and Upload Go packages, frontend typecheck/build, lint-openapi and 37-operation fixture validation pass. make codegen is byte-idempotent across 77 outputs, and make codegen-check passes against an isolated index containing the current intended contract without mutating the real index." -->
- [x] 18.7 BDD-Gate: 原地修订 `E2E.P0.014` / `E2E.P0.015`，P0.015 覆盖 paste success、exact request、auth pending replay、4xx/failed、idempotency、privacy 与 Parse loading；在 1440×900 和 390×844 捕获 Home/Parse 截图并验证 DOM、computed style、bbox 与 viewport parity。
  <!-- verified: 2026-07-13 method=scenario-run evidence="E2E.P0.014 and E2E.P0.015 setup/trigger/verify/cleanup PASS serially; Home formal/prototype changedRatio desktop=0.000323 mobile=0.000481; Parse changedRatio desktop=0.000556 mobile=0.001499 with exact viewport, DOM/style/bbox gates." -->
- [x] 18.8 ZERO-REF-GATE: 删除 URL 专属 `E2E.P0.011` 实体目录与 active INDEX 行且不复用编号；扫描 active UI truth、owner docs、OpenAPI/generated、frontend Home、backend TargetJob 与 active scenarios，确认旧 JD source controls/modal/route source/schema/fixture/branch/scenario 为零，排除 work-journal/bug/report 等合法历史证据并明确允许 Resume upload 资产。
  <!-- verified: 2026-07-14 method=exact-negative-scan evidence="P0.011 directory and active INDEX row, JDAssistModal files, urlfetch package, source.refreshed schema and source enum refs are absent. Current UI/OpenAPI/generated/TargetJob/shared/migration/config/prompt production scans have zero retired control/schema/event/job/handler/purpose/token hits; remaining owner-doc and active-scenario hits are normative negative assertions, while Resume upload and independent source_records stay explicitly allowed." -->
- [x] 18.9 PHASE-HANDOFF: owner spec/plan/checklist/BDD/context 与 document INDEX 完成 reconcile，`validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check` 和 pruning gate 通过；因 Phase 19 已重开，计划保持 `active`，最终 `completed` 仅由 19.7 承接。
  <!-- verified: 2026-07-14 decision="用户批准方案 A" method=context+docs+diff+pruning evidence="frontend target context validation, sync-doc-index, docs links/fragments, git diff --check and pruning gate all pass; pruning reports real_residuals=0. Phase 18 hands off without claiming whole-plan completion." current-owner="19.7" -->

## Phase 19: Plan-detail report entry and independent-list handoff

- [x] 19.1 HISTORICAL RED-GREEN: prototype/formal shared ready-detail state 增加“面试报告”页面级入口；Phase 20 将该 ready state 独占到 Workspace detail，Parse command progress 不渲染入口。
  <!-- verified: 2026-07-14 method=ui-contract+P0.016+acceptance-screenshots evidence="Prototype/formal source contract and P0.016 pass; exact desktop/mobile acceptance images show the content-header entry and three-item TopBar with no global report entry." -->
- [x] 19.2 RED-GREEN: 入口只使用当前可信 TargetJob ID，点击精确导航 `{ name: "reports", params: { targetJobId } }`；缺失/非法上下文不得拼接 report/status/round authority。
  <!-- verified: 2026-07-14 method=prototype+formal-vitest evidence="UI contract 62/62 and Parse report handoff focused tests prove the page-level entry uses only the loaded matching TargetJob ID and navigates to reports with targetJobId only." -->
- [x] 19.3 RED-GREEN: 从 shared detail prototype/formal component/i18n/tests 删除嵌入式报告列表、current/latest row、loading/error/empty state 和相关交互；Workspace 只读详情与 Start 行为不变。
  <!-- verified: 2026-07-14 method=source-contract+focused-vitest evidence="Prototype contract 62/62 and Parse focused tests pass after embedded report DOM/state/actions and nine unreachable list locale keys were removed; only parse.reports.label remains for the page-level entry." -->
- [x] 19.4 RED-GREEN: 删除 shared detail 的 `listTargetJobReports` effect/loader/validator consumer；component/generated-client spy 证明 Parse 与 Workspace detail 均零列表请求，正式 consumer 只在 report owner。
  <!-- verified: 2026-07-14 method=focused-vitest evidence="ParseReports.test.tsx 5/5 proves ready, loading, failed, hostile section and TargetJob switch make zero listTargetJobReports calls and render no list." -->
- [x] 19.5 RED-GREEN: 删除 `section=reports` safe param、锚点、滚动/聚焦与兼容分支；hostile `section`/report/status/round query 被 shared route 层剔除且不能改变 TargetJob。
  <!-- verified: 2026-07-14 method=route-tdd evidence="Routing/App/auth/privacy focused suite 152/152 and legacy P0.088/P0.089/P0.090 unit regressions 23/23 pass with Parse section/report/status/round inputs stripped and no focus/scroll branch." -->
- [x] 19.6 BDD-Gate: 原地扩展 `E2E.P0.016` 覆盖 Workspace detail 右上入口、精确 target handoff、TopBar/Parse-entry negative、两处无列表/无请求/无 section 兼容、只读详情与 Start 回归及 desktop/mobile parity。
  <!-- verified: 2026-07-14 scenario=E2E.P0.016 evidence="Source 3/3, real API 1/1, Vitest 105/105, production build and Playwright 6/6 pass; formal/prototype changedRatio is 0 at both viewports." -->
- [x] 19.7 POST-PASS: focused Parse/route/i18n tests、UI source contract、typecheck/build、P0.016、owner contexts、docs/diff 与 `listTargetJobReports` consumer negative gate 全部通过后再完成本 Phase。
  <!-- verified: 2026-07-14 method=current-aggregate evidence="Parse/route/i18n, UI 62/62, frontend 121 files/977 tests, typecheck/build, P0.016, context and single-consumer negatives pass; docs/diff/pruning closeout is current." -->

## Phase 20: command-only Parse and exact GET counts

- [x] 20.1 RED-GREEN: import success navigates only `/parse?targetJobId`; route has no `resumeId`/raw status, queued/processing remains on progress, and ready initial/poll state calls workspace-detail replace.<!-- verified: 2026-07-14 method=vitest-red-green evidence="HomeImport and routeUrl RED rejected resumeId/jdId/importId; ParseFlow RED rejected preview delay and duplicate cache-driven GETs. GREEN passed 51 route/import tests and 7 poll/replace tests." -->
- [x] 20.2 RED-GREEN: Home recent ready card body directly opens `/workspace?targetJobId`; quick-start and More remain; no card path enters Parse or starts import/poll/animation.
- [x] 20.3 RED-GREEN: under React StrictMode, bottom transport spies prove Home `listTargetJobs` and `listResumes` each issue exactly one same-key initial GET via shell/001 Phase 13.
- [x] 20.4 RED-GREEN: Parse each classification/scheduler tick issues exactly one same-key `getTargetJob`; later calls require timer advancement, while route/auth/locale/read epoch changes remain distinct.
- [x] 20.5 BDD-Gate: update `E2E.P0.014` for Home exact counts/direct card; `E2E.P0.015` for POST-command/poll/ready replace/Back; `E2E.P0.016` and `E2E.P0.018` for Workspace ready detail with exactly one same-key `getTargetJob`, zero `listResumes`/import/poll/auto-start and no Parse animation.
- [x] 20.6 POST-PASS: focused Home/Parse/transport/route tests, P0.014/P0.015/P0.016/P0.018, frontend typecheck/build, owner contexts, docs/diff and duplicate-request negative gates pass before restoring `completed`.
  <!-- verified: 2026-07-14 evidence="Fresh P0.014/P0.015/P0.016/P0.018 wrappers PASS; P0.015 includes 57 tests, build, desktop/mobile and StrictMode transport initial=1/tick1=2/tick2=3. Browser acceptance proves ready card read-only Workspace handoff and import-only Parse progress." -->

## Phase 21: Workspace detail round-state affordance

- [x] 21.1 RED-GREEN: prototype round-assumption cards derive `done/current/pending` only from `eiResolvePracticeProgress`, expose state labels/attributes and use success-soft/accent-soft/neutral-soft background+border treatments.<!-- verified: 2026-07-14 method=ui-contract red-green result="new contract RED; 65/65 GREEN" -->
- [x] 21.2 RED-GREEN: formal Workspace detail derives the same exact states from `resolveTargetJobPracticeProgress`; focused tests cover in-progress, all-completed and invalid projections without lifecycle/URL/storage fallback.<!-- verified: 2026-07-14 method=ParseRoundStates+roundAssumptions result="29/29 PASS; lifecycle independence included" -->
- [x] 21.3 PARITY-GATE: UI contract plus desktop/mobile DOM/computed-style/bbox/viewport checks prove the formal three-state cards source-match `ui-design/` and remain visually distinct in light/dark/custom themes.<!-- verified: 2026-07-14 method=parse-pixel-parity result="desktop+mobile 2/2; 3 backgrounds; 3 borders; no overflow; screenshots attached" -->
- [x] 21.4 BDD-Gate: existing `E2E.P0.016` browser path captures the three labels/states/backgrounds and confirms consistency with the list-card mini rail; no sibling scenario.
- [x] 21.5 POST-PASS: focused round/detail tests, UI contract, P0.016, typecheck/build, owner context/docs/diff and negative state-source searches pass before restoring `completed`.
  <!-- verified: 2026-07-14 evidence="P0.016 desktop/mobile PASS; UI contract 65/65 and acceptance screenshot 03 prove persisted done/current/pending labels and distinct background/border treatments without URL/storage fallback." -->
