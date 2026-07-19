# 001 Home + JD Import + Parse Checklist

> **版本**: 2.35
> **状态**: completed
> **更新日期**: 2026-07-19

**关联计划**: [plan](./plan.md)

## Phase 1: Home 当前入口

- [x] 1.1 Historical Phase 1 delivered the original Home shell, JD input card, ready 简历下拉框、创建简历入口、提交区、最近 3 张模拟面试卡片和 More handoff；Phase 18 owns the current paste-only intake surface.
- [x] 1.2 Historical Phase 1 delivered the original generated-client import baseline, idempotency、错误态和 pending continuation；Phase 18 supersedes the current request and continuation contract.
- [x] 1.3 Historical Phase 1 route carried `resumeId` after explicit ready-resume selection；Phase 20 supersedes the route to `targetJobId` only while the POST body still includes the selected `resumeId`.

## Phase 2: Historical pre-readonly Parse confirmation and handoff

- [x] 2.1 Historical pre-Phase 6 Parse parity covered loading, preview, failed state, editable basics, requirements, hidden signals, round assumptions, resume binding and footer actions; Phase 6 now supersedes success preview with a readonly receipt.
- [x] 2.2 Historical generated-client gates covered Parse `getTargetJob`, `listResumes` and `updateTargetJob` behavior；Phase 20 supersedes ready rendering to Workspace detail, which uses one `getTargetJob`, no `listResumes` and no `updateTargetJob`.
- [x] 2.3 Historical readonly detail ignored route-only `resumeId`；Phase 20 removes that route param entirely, reads the saved TargetJob binding in Workspace detail and disables Start when missing.
- [x] 2.4 Historical Save plan / workspace auto-start handoff was covered before readonly simplification; current success path has no Save plan action and Start enters practice directly.

## Phase 3: 收口验证

- [x] 3.1 `validate_context.py frontend-home-job-picks-and-parse/001 frontend` 通过。
- [x] 3.2 Focused Home/Parse Vitest、frontend typecheck 与 `make validate-fixtures` 通过。
- [x] 3.4 `sync-doc-index --check`、`make docs-check`、`git diff --check` 和 `make lint-core-loop-pruning-surface` 通过。

## Phase 4: Import resume binding remediation

- [x] 4.1 Historical import variants included the selected `resumeId` in generated `importTargetJob` request bodies；Phase 18 preserves the binding in the flattened paste-only body（历史验证：`HomeImport.test.tsx`, `HomeResumeSelection.test.tsx`, `HomeAuthGate.test.tsx` PASS）
- [x] 4.2 Historical Parse route handoff carried `resumeId`；Phase 20 supersedes it with targetJobId-only routing and authoritative `TargetJob.resumeId` recovery。历史 focused tests 仅作开发反馈，当前阶段单测完成由根 `make test` 承接。

## Phase 5: Unified plan detail remediation

- [x] 5.1 Historical UI work renamed the Parse-derived ready visual to `面试规划详情 / 面试上下文确认` while preserving first-import loading；Phase 20 keeps the visual only under Workspace detail（历史验证：`frontend/src`, docs/locales/parity PASS）
- [x] 5.2 Historical parse/workspace routes rendered the same detail DOM；Phase 20 supersedes this so Parse ready replace-navigates and only `route=workspace` with `targetJobId` renders readonly resume binding and Start, while query-free workspace renders `WorkspacePlanList`。历史 focused tests 仅作开发反馈，当前阶段单测完成由根 `make test` 承接。
- [x] 5.3 Shared detail navigation uses declared `TargetJob.currentPracticePlanId` / `TargetJob.resumeId` without fabricating `plan-${targetJobId}` or `resume-unbound`, and out-of-scope independent workspace detail anchors are covered by negative tests（验证：`frontend/src/app/navigation/interviewContext.ts`, `interviewContext.test.ts`, `frontend/src/app/screens/workspace/WorkspaceEmptyState.test.tsx`, `formal frontend component tests` PASS）

## Phase 6: Readonly plan detail simplification

- [x] 6.1 UI design document and formal copy make the shared success detail a readonly context receipt with only Start interview as the footer action；Phase 20 locates it only at Workspace detail。UI contract/parity 是独立 gate，单测完成由根 `make test` 承接。
- [x] 6.2 Workspace success detail removes field edit state, requirement toggles, hidden-signal remove controls, resume picker / create fallback, success Re-parse, Save plan and Cancel controls（历史 Parse-derived component tests PASS）
- [x] 6.3 Start interview uses the saved `targetJobId/resumeId/roundId/currentPracticePlanId` snapshot and must not call `updateTargetJob`; missing bound resume blocks Start without offering in-place binding。Focused Parse/client-spy tests 仅作开发反馈，阶段单测完成由根 `make test` 承接。
- [x] 6.5 Repo gates pass after doc/code/test changes（验证：context validation, sync-doc-index, docs-check, diff whitespace check, touched frontend tests/typecheck PASS）

## Phase 7: LLM-derived round assumptions shared data binding

- [x] 7.1 Historical UI design document and owner docs first moved TargetJob round assumptions off local-only copy; Phase 8/20 use `TargetJob.summary.interviewRounds[]` across Workspace detail, Home recent cards and shared navigation context（当前验证见 Phase 8.1-8.5).
- [x] 7.2 Focused shared-detail tests prove round cards render backend-provided structured rounds and do not show static locale focus when structured rounds exist（当前验证见 8.3).
- [x] 7.3 Focused Home/navigation tests prove `home-recent-mock-rail-*` and `interviewContextFromTargetJob` consume the same backend-provided structured rounds instead of a local `DEFAULT_ROUNDS` / static round name（当前验证见 8.3).
- [x] 7.4 Frontend implementation uses a shared TargetJob round assumption mapper without changing the Parse layout, Home card layout, Workspace list layout or Start handoff。Focused frontend tests 仅作开发反馈，阶段单测完成由根 `make test` 承接；typecheck 为独立 gate。

## Phase 8: Structured LLM-derived interview rounds

- [x] 8.1 OpenAPI / prompt / fixture contract defines `TargetJob.summary.interviewRounds[]` with 2~5 LLM-derived rounds, each carrying `sequence`, `type`, `name`, `durationMinutes` and `focus`; prompt explicitly instructs inference from JD, role seniority, company/industry nature, team/business context, hiring-process hints and common interview practices, and generated Go/TS artifacts are refreshed（验证：`make lint-prompts`, `cd backend && go test ./internal/targetjob -run TestTargetImportPromptMatchesParseResponseSchema -count=1`, `make codegen-openapi`, `make lint-openapi`, `make validate-fixtures` PASS).
- [x] 8.2 Backend parser validates and persists 2~5 structured `interviewRounds[]` from `target.import.parse` without fabricating fixed 4-round defaults。Focused targetjob tests 仅作开发反馈，阶段单测完成由根 `make test` 承接。
- [x] 8.3 Frontend Workspace detail/Home/navigation consume structured rounds with variable count, type/name and duration from `summary.interviewRounds[]`; fixed strings such as `HR 初筛 · 20m` / `技术一面 · 45m` are not used when structured rounds exist（历史 Parse-derived component tests + current Workspace gate）.
- [x] 8.4 UI design document and docs define structured LLM rounds across Workspace detail and Home recent rail（验证：UI contract/fixtures PASS）.
- [x] 8.6 Repo gates pass after structured round contract changes（验证：`python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/context.yaml --docs-root docs --target frontend`; `cd frontend && pnpm typecheck`; focused frontend tests; backend targetjob focused tests; `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`; `make docs-check`; `git diff --check`; `make lint-core-loop-pruning-surface` PASS).

## Phase 9: Recent card fixed grid and workspace fusion

- [x] 9.1 UI design document defines Home recent cards and workspace plan-list cards as one shared card body with fixed max-width grid（验证：`docs/ui-design/jd-resume-management.md`, `docs/ui-design/module-job-workspace.md`, `frontend/src`, `python3 scripts/lint/ui_demo_pruning.py` PASS）
- [x] 9.2 Formal `MockInterviewCard` supports Home default testids plus workspace-owned card/body/rail/footer testids and optional footer CTA（验证：`pnpm --filter @easyinterview/frontend test src/app/screens/home/MockInterviewCard.test.tsx` PASS）
- [x] 9.3 Home recent and workspace list focused tests reject `1fr` stretching and verify workspace mini round rail + footer CTA（验证：`pnpm --filter @easyinterview/frontend test src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceEmptyState.test.tsx` PASS）

## Phase 10: Home recent shared action card

- [x] 10.1 UI design document defines Home recent cards as the shared Interview list action card with `立即面试` and without delete controls（验证：`docs/ui-design/jd-resume-management.md`, `docs/ui-design/module-job-workspace.md`, `frontend/src`）
- [x] 10.2 Formal `MockInterviewCard` supports quick-start action props and Home passes no delete action（验证：`MockInterviewCard.test.tsx`, `HomeRecentMocks.test.tsx`）
- [x] 10.3 Home recent quick-start calls shared generated practice handoff with structured `roundId/roundName`, and card-body click remains planning-detail navigation（验证：`HomeRecentMocks.test.tsx` PASS）
- [x] 10.4 Browser screenshot acceptance captures Home recent card with `立即面试` and no delete icon（验证：`.test-output/screenshots/home-recent-action-card.png`）
- [x] 10.5 Home recent requests `listTargetJobs(analysisStatus=ready)` and filters failed / processing / queued / blank-title dirty records before rendering cards（验证：`HomeRecentMocks.test.tsx` PASS）



## Phase 12: Pending-import test API removal

- [x] 12.1 RED/GREEN: Home source gate detects and then rejects the production `clearPendingImportSourcesForTests` export.

## Phase 13: Current fixture inventory wording

- [x] 13.1 Keep OpenAPI inventory, fixture validation, owner contexts and docs/diff/pruning as code/document gates；Home import/parse currently has no real E2E owner.

## Phase 14: Home copy-table orphan cleanup

- [x] 14.1 删除 UI prototype、zh/en locale 与 locale self-test 中无渲染 consumer 的 `uploadSourceSub`；验证 Home/UI contract、locale reachability、owner contexts 与 docs/diff/pruning gates。
  <!-- verified: 2026-07-10 method=home-copy-table-orphan-removal evidence="Deleted uploadSourceSub from both prototype language branches, both formal locale catalogs and the locale self-test requirement. Scoped runtime/prototype search is zero; locale reachability, 35 UI contracts, frontend tests/typecheck/build, Home/product contexts and docs/diff/pruning gates pass with no visible Home DOM or behavior change." -->

## Phase 15: MiniRoundRail prototype call-surface pruning

- [x] 15.1 新增 Home rail 参数消费 contract，并先红证明 `MiniRoundRail` 与调用方仍保留未读取的 `lang`。
  <!-- verified: 2026-07-10 method=home-mini-round-rail-red evidence="UI contract ran 42 tests: the new structured-round dependency contract failed on the existing MiniRoundRail.lang parameter while the prior 41 tests passed; retained assertions pin round count, names, durations and current-index highlighting." -->
- [x] 15.2 删除 rail 的零读取 `lang` 形参与调用方传参；验证：AST `MiniRoundRail` 参数消费 inventory 归零，结构化轮次和 current-index 高亮保持原样。
  <!-- verified: 2026-07-10 method=home-mini-round-rail-green evidence="Removed only MiniRoundRail.lang and its single caller argument; MockInterviewCard.lang remains for visible action copy. UI contract passes 42/42 and Babel binding inventory reports railUnread=[] while retaining structured rounds and currentIndex assertions." -->



## Phase 17: Parse loading internal-metadata removal

- [x] 17.1 RED-GREEN: prototype source contract rejects the Parse loading model/provider, rubric/prompt/version/hash, provenance and typical-latency footer while preserving four progress steps and responsive layout.
- [x] 17.2 RED-GREEN: formal Parse loading DOM removes the same internal metadata without changing polling, ready mapping or failed recovery.
- [x] 17.3 REGRESSION-GATE: 仓库根 `make test` 完成前后端全量单测回归；UI source contract、typecheck/build 与 active negative search 作为独立 gates。

## Phase 18: Paste-only Home JD intake

- [x] 18.1 RED-GREEN: prototype-first 更新 `frontend/src` 与 UI source contract，使 Home intake 仅渲染 textarea、ready Resume select 和 CTA；旧 source controls、辅助弹窗、触发 testid 与多入口 copy 必须先红后删，且 Resume 上传入口保持可用。
- [x] 18.2 RED-GREEN: OpenAPI、fixtures 与 generated Go/TS 将 `importTargetJob` public request 收敛为 exact flattened wire `{ rawText, targetLanguage, resumeId }`，拒绝 source discriminator、嵌套 source payload 和非文本 JD intake；Resume upload operation/fixture 不受影响。
  <!-- verified: 2026-07-13 evidence="OpenAPI inventory 24 tests, generator tests, lint-openapi, fixture validator 37 operations and isolated-index codegen-check PASS; generated Go/TS request is flattened; upload fixture retains resume/privacy only." -->
- [x] 18.3 RED-GREEN: formal Home layout/import/i18n tests 证明只存在一个 paste submit path；删除 source controls/modal、source-specific branches、额外 locale keys 和 JD upload-client 调用，不新增 mode enum、compatibility adapter 或不可达 branch。
- [x] 18.4 RED-GREEN: Home auth `pendingAction` 只保存 `opaquePendingImportId`；一次性进程内 vault 保存 `{ rawText, targetLanguage, resumeId, idempotencyKey, expiresAt }` 并在登录成功后原子 consume 一次。正常路径以原 key 提交同一 exact request；refresh/lost、expired、duplicate consume 均不调用 import，清除 action、返回 Home 并显示本地化重新输入提示；route/browser storage/log/telemetry 均不得携带 JD 原文。
- [x] 18.5 RED-GREEN: backend TargetJob decode/service/store/runner 删除 public 多来源 union、URL fetch/refresh、JD attachment purpose、structured manual-form branch 及 source-specific persistence/event/config，同时保留文本 validation、idempotency、parse failure/retry、privacy 与 resume binding。
- [x] 18.6 REGRESSION-GATE: UI contract、Home focused Vitest、OpenAPI lint/fixture/generated drift、backend TargetJob tests、typecheck/build 与 Resume upload owner tests 全部通过。
- [x] 18.9 PHASE-HANDOFF: owner spec/plan/checklist/BDD/context 与 document INDEX 完成 reconcile，`validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check` 和 pruning gate 通过；因 Phase 19 已重开，计划保持 `active`，最终 `completed` 仅由 19.7 承接。
  <!-- verified: 2026-07-14 decision="用户批准方案 A" method=context+docs+diff+pruning evidence="frontend target context validation, sync-doc-index, docs links/fragments, git diff --check and pruning gate all pass; pruning reports real_residuals=0. Phase 18 hands off without claiming whole-plan completion." current-owner="19.7" -->

## Phase 19: Plan-detail report entry and independent-list handoff

- [x] 19.1 HISTORICAL RED-GREEN: prototype/formal shared ready-detail state 增加“面试报告”页面级入口；Phase 20 将该 ready state 独占到 Workspace detail，Parse command progress 不渲染入口。
- [x] 19.2 RED-GREEN: 入口只使用当前可信 TargetJob ID，点击精确导航 `{ name: "reports", params: { targetJobId } }`；缺失/非法上下文不得拼接 report/status/round authority。
- [x] 19.3 RED-GREEN: 从 shared detail prototype/formal component/i18n/tests 删除嵌入式报告列表、current/latest row、loading/error/empty state 和相关交互；Workspace 只读详情与 Start 行为不变。
- [x] 19.4 RED-GREEN: 删除 shared detail 的 `listTargetJobReports` effect/loader/validator consumer；component/generated-client spy 证明 Parse 与 Workspace detail 均零列表请求，正式 consumer 只在 report owner。
- [x] 19.5 RED-GREEN: 删除 `section=reports` safe param、锚点、滚动/聚焦与兼容分支；hostile `section`/report/status/round query 被 shared route 层剔除且不能改变 TargetJob。

## Phase 20: command-only Parse and exact GET counts

- [x] 20.1 RED-GREEN: import success navigates only `/parse?targetJobId`; route has no `resumeId`/raw status, queued/processing remains on progress, and ready initial/poll state calls workspace-detail replace.
- [x] 20.2 RED-GREEN: Home recent ready card body directly opens `/workspace?targetJobId`; quick-start and More remain; no card path enters Parse or starts import/poll/animation.
- [x] 20.3 RED-GREEN: under React StrictMode, bottom transport spies prove Home `listTargetJobs` and `listResumes` each issue exactly one same-key initial GET via shell/001 Phase 13.
- [x] 20.4 RED-GREEN: Parse each classification/scheduler tick issues exactly one same-key `getTargetJob`; later calls require timer advancement, while route/auth/locale/read epoch changes remain distinct.

## Phase 21: Workspace detail round-state affordance

- [x] 21.1 RED-GREEN: prototype round-assumption cards derive `done/current/pending` only from `eiResolvePracticeProgress`, expose state labels/attributes and use success-soft/accent-soft/neutral-soft background+border treatments.<!-- verified: 2026-07-14 method=ui-contract red-green result="new contract RED; 65/65 GREEN" -->
- [x] 21.2 RED-GREEN: formal Workspace detail derives the same exact states from `resolveTargetJobPracticeProgress`; focused tests cover in-progress, all-completed and invalid projections without lifecycle/URL/storage fallback.
- [x] 21.3 PARITY-GATE: UI contract plus desktop/mobile DOM/computed-style/bbox/viewport checks prove the formal three-state cards source-match `frontend/` and remain visually distinct in light/dark/custom themes.<!-- verified: 2026-07-14 method=parse-responsive-browser result="desktop+mobile 2/2; 3 backgrounds; 3 borders; no overflow; screenshots attached" -->

## Phase 22: Required runtime JD guard

- [x] 22.1 RED/GREEN: Home 消费 required `targetJobRawTextBytes` 与共享 UTF-8 helper；小型 injected limit 覆盖多字节、零 request/vault，DOM/style 不变。
- [x] 22.2 FALLBACK-GATE: required 子字段无 per-field fallback；仅整体 runtime source 不可用时保留既有 bootstrap fallback。

## Phase 23: Workspace detail leading resume link and action row

- [x] 23.1 RED：shared Workspace detail tests/source/responsive contract 先拒绝标题右侧 Report、`parse-launch`/`parse-resume-binding` 独立 block 与页尾 Start，并保留 exact GET、round state、route 与 fail-closed 回归。
  <!-- verified: 2026-07-15 method=vitest-red expected-failures="missing title-adjacent resume and leading action row" -->
- [x] 23.2 GREEN：标题旁“绑定简历”只用 saved `TargetJob.resumeId` 导航 `resume_versions?resumeId=...`；缺失绑定为非链接并禁用 Start，零 `getResume`/`listResumes`/route/list/recent fallback，零 picker/rebind；删除旧 block-only `parse.launch*`、`parse.resumeBound*`、`parse.footerHint` locale key 与断言。
  <!-- verified: 2026-07-15 method=vitest evidence="parse/workspace 11 files 54 tests PASS" -->
- [x] 23.3 GREEN：标题下首行动作行依次为“立即面试” primary 和“面试报告” secondary；desktop 同排、mobile 同序换行，Report 只带可信 target，Start 保持 saved resume/current round，启动错误不阻断 Report。
  <!-- verified: 2026-07-15 method=vitest evidence="parse/workspace 11 files 54 tests PASS" -->
- [x] 23.4 PARITY/A11Y：desktop/mobile DOM、可访问名称、键盘/触控、computed style、bbox/no-overflow 通过；独立 launch/binding、标题右侧 Report、页尾 Start 与 orphan locale key 零残留。（验证：Chrome 1440/390 actions 44px、左对齐同序、标题链接同簇、无横溢；旧 DOM 0）
- [x] 23.5 POST-PASS：根 `make test`、frontend typecheck/build、owner contexts、`sync-doc-index --check`、`make docs-check`、`git diff --check` 通过，与 Workspace owner 同步恢复 completed。（验证：根后端 551 tests/4493 subtests、前端 125 files/993 tests PASS；lint/build/context/docs/index/diff PASS）

## Phase 24: Required Resume product-contract reconciliation

- [x] 24.1 Product/UI truth sources 明确 selectable Resume 是 JD import、Practice、Reports、复练和下一轮的永久强制前置，并删除无简历训练与降级报告承诺。<!-- verified: 2026-07-15 method=in-place-design-reconcile evidence="product-scope D-30/C-28 and four UI owner documents now reject resume-less import, training and report downgrade while preserving ready-or-readable-evidence selection; stale commitment search is zero" -->
- [x] 24.2 Owner spec/plan/BDD/context 将无选择时零 import、形成可读证据后返回 Home 显式选择、历史缺失/无效绑定全链路 fail closed 固化为当前合同。<!-- verified: 2026-07-15 method=focused-contract-evidence evidence="Home/Parse 5 files 38 tests PASS; Home selection predicate, canSubmit/import guard and OpenAPI required resumeId match the revised owner contract" -->
- [x] 24.3 Existing Home/Workspace focused tests、active-doc negative search、context、Header/INDEX、docs、diff 与 pruning gates 全部通过；本阶段无代码变更且不声明新 E2E。<!-- verified: 2026-07-15 method=post-pass-gates evidence="5 focused files 38 tests PASS; stale commitment rg returned zero matches; validate_context, sync-doc-index, docs-check, git diff --check and lint-core-loop-pruning-surface PASS with real_residuals=0" -->

## Phase 25: Screenshot-aligned Home visual system

- [x] 25.1 RED: `HomeLayout.test.tsx` / `HomeRecentMocks.test.tsx` 固化 screenshot-aligned hierarchy、single intake card、真实 runtime count 与横向 recent record；旧 Home fixed-card grid/eyebrow/分散表单结构先失败。TopBar 由 `frontend-shell/002` Phase 22 独立承接。<!-- verified: focused Home RED 32 tests, 10 failed / 22 passed; failures covered the complete retired visual structure -->
- [x] 25.2 GREEN: `.ei-home-screen` page-scoped CSS 与正式 Home/MockInterviewCard 实现 1:1 还原参考 viewport；保留 generated client、Resume gate、privacy、route、idempotency、structured rounds 和 practice handoff。<!-- verified: Home focused 32/32 PASS; frontend typecheck PASS; production source keeps client/auth/import/practice chain -->
- [x] 25.3 BDD-Gate: 执行 `BDD.HOME.JD.003` owner behavior tests，完成 zh/en、loading/empty/error、disabled/enabled、1~3 recent records、theme/dark 与 accessibility 断言。<!-- verified: Home/TopBar/i18n focused 102/102 PASS plus MockInterviewCard keyboard regression 14/14 PASS; root frontend 1044/1044 PASS -->
- [x] 25.4 CHROME-GATE: 用户现有 Chrome 在 `1916x821` 对照参考图检查 bbox/spacing/type/color/no-overflow，并在 `390x844` 检查单列、触控与 document containment；记录截图路径和 console 结果，不声明真实 E2E。<!-- verified: Chrome desktop card y=242 h=325, recent record y=679 h=130, scrollWidth=1916; mobile scrollWidth=390, light/dark and disabled/enabled transitions PASS; console warnings/errors=0; evidence=.test-output/home-ui-acceptance -->
- [x] 25.5 REGRESSION-GATE: focused tests、根 `make test`、frontend typecheck/build、owner contexts、Header/INDEX/docs/diff/pruning gates 全部通过后恢复 completed。<!-- verified: focused Home/TopBar/i18n 102/102, keyboard 14/14, root 615 tooling plus backend plus frontend 1044/1044, build/context/docs/diff/pruning PASS -->

## BDD Gate

- [x] BDD-Gate: `BDD.HOME.JD.001` 由 [BDD checklist](./bdd-checklist.md) 关联 import/parse/handoff owner behavior tests；P0.098 不承接该流程。
- [x] BDD-Gate: `BDD.HOME.JD.002` 由 [BDD checklist](./bdd-checklist.md) 关联共享 Workspace ready-detail 的绑定简历查看与首行动作行行为；当前无该 UI 的真实 E2E owner。

## Phase 26: Screenshot-aligned Workspace plan detail

- [x] 26.1 RED: 新增 `ParsePlanVisual.test.ts` 并扩展 detail tests，锁定 `1250px` Header 右侧动作与基本信息、要求、隐性关注点、动态轮次四层卡面；旧 inline/左对齐动作构图先失败。<!-- verified: 2026-07-19 method=focused-vitest-red evidence="ParsePlanVisual failed 3/3 on the old inline 1200px detail shell, absent header grid, absent four-layer class anchors and absent mobile CSS while the shared focused run kept 24 prior assertions green." -->
- [x] 26.2 GREEN: 以 `.ei-plan-detail-*` page-scoped CSS 与仓库内 SVG/CSS icon 实施；保留 saved resume、Start/Reports、dynamic rounds、progress、route、request count 与 fail-closed 行为。<!-- verified: 2026-07-19 method=focused-vitest-green evidence="Parse visual, detail and saved-resume suites pass 18/18 in the shared 72-test run; frontend typecheck passes. Start/Reports remain ordered in the header and all existing route/client/fail-closed tests stay green." -->
- [x] 26.3 BDD-Gate: `BDD.HOME.PLAN.VISUAL.004` 由合法/缺绑/无效 progress、desktop/mobile responsive/a11y tests 和 current-run Chrome manual acceptance 承接，不新增 E2E ID。<!-- verified: 2026-07-19 method=chrome-extension-manual evidence="Chrome at 1916x821 verified the 1250px centered detail column, header actions and four card layers; 390x844 verified stacked actions, single-column cards and scrollWidth=viewportWidth with no horizontal overflow." -->
- [x] 26.4 REGRESSION: Parse/Workspace focused、frontend typecheck/build、根 `make test`、owner context、Header/INDEX/docs/diff/pruning gates 全部通过后恢复 completed。<!-- verified: 2026-07-19 method=full-regression evidence="Parse visual/detail/binding suites pass 18/18; shared visual/detail suites pass 95/95; frontend typecheck/build pass; root make test passes Python 615 with 4615 subtests, Go all packages, and frontend 133 files / 1066 tests. Owner context, doc/index, pruning and diff gates pass." -->
