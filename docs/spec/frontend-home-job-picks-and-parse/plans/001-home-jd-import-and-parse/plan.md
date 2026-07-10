# 001 Home + JD Import + Parse

> **版本**: 2.23
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本计划交付当前 Home + Parse 新建模拟面试入口，并在 v2.10 原地修订中把轮次假设升级为结构化 LLM/JD parse 合同：Parse 详情、Home 最近模拟面试卡片、Workspace 规划回访 handoff 和共享导航上下文必须使用同一份 TargetJob structured round mapper；卡片视觉仍复刻 UI 真理源，但轮次数量必须为 2~5，轮次类型、标题、时长和 focus 都必须来自后端保存的 `TargetJob.summary.interviewRounds[]`。v2.11 原地修订把 Home 最近模拟面试卡片和 workspace 面试列表卡片收敛到同一个 `MockInterviewCard` 主体：Home recent grid 使用固定最大列宽，workspace 只追加 footer CTA。本次 v2.12 原地修订要求 Home recent 继续复用 workspace 面试列表卡片动作模型：点击卡片主体进入统一规划详情，`立即面试` 主按钮直接启动 practice，但 Home 不展示删除按钮。v2.13 review remediation 要求 Home recent 只准入 ready TargetJob，并且 quick-start 必须把结构化 `roundId/roundName` 带入 practice route。v2.14 将 workspace detail 负向锚点统一为 out-of-scope 口径。用户从首页输入、上传或 URL 导入 JD，显式选择一份 ready 简历，进入统一详情页核对已保存的 JD、简历和由 LLM 推断的面试轮次；既有规划从 `workspace` 列表回访时也使用同一母版，不再出现第二套 workspace 当前规划详情页。

交付后的当前链路：

```text
Home JD source + ready Resume
  -> importTargetJob(resumeId)
  -> Parse loading
  -> Unified Plan Detail / Context Confirm
  -> Start interview
  -> practice
```

## 2 背景

`frontend-shell` 提供 App 壳、route normalization、auth pending action、runtime config、generated client bootstrap 与 fixture-backed transport。本 owner 负责 `home`、`parse` loading 和统一详情母版。Workspace 无上下文列表仍归 `frontend-workspace-and-practice`；详情页 Start action 直接调用 practice handoff helper，不再通过 workspace auto-start route 制造副作用。

UI 必须源级追溯到 `ui-design/src/screen-home.jsx::HomeScreen`、`ui-design/src/screens-p0-complete.jsx::ParseScreen` 与 `ui-design/src/primitives.jsx`。正式前端只允许为真实数据、generated client、鉴权接续和可访问性做工程适配。

当前 API 合同来自 `openapi/openapi.yaml` 与 fixtures：

| operationId | Fixture | Frontend consumer | Backend owner | Persistence | AI dependency | Scenario |
|-------------|---------|-------------------|---------------|-------------|---------------|----------|
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json` | Home recent mock interviews, ready TargetJob only | `backend-targetjob` | TargetJob read | none in frontend | `E2E.P0.014` |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` | Home resume select + Parse readonly bound resume display | `backend-resume` | Resume read | none | `E2E.P0.015` / `E2E.P0.016` |
| `createUploadPresign` | `openapi/fixtures/Uploads/createUploadPresign.json` | Home upload source | `backend-upload` | `file_objects` create | none | `E2E.P0.015` |
| `importTargetJob` | `openapi/fixtures/TargetJobs/importTargetJob.json` | Home paste / file / URL import with selected `resumeId` | `backend-targetjob` | TargetJob create + target job-level resume binding + parse job | backend-only | `E2E.P0.015` |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` | Parse polling + readonly preview, including structured rounds from `summary.interviewRounds[]` | `backend-targetjob` | TargetJob read from `target_jobs.summary` | backend-only `target.import.parse` structured round result | `E2E.P0.015` / `E2E.P0.016` |
| `createPracticePlan` / `getPracticePlan` / `startPracticeSession` | `openapi/fixtures/PracticePlans/*`, `openapi/fixtures/PracticeSessions/*` | Parse readonly detail Start action and Home recent quick start | `backend-practice` | PracticePlan / PracticeSession create-read-start | none | `E2E.P0.016` |

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `contract`
- **TDD 策略**: v2.11 继续通过 `/implement` -> `/tdd` 执行；Red-Green-Refactor 覆盖 `config/prompts/target.import.parse/*`、`openapi/openapi.yaml`、`backend/internal/targetjob/*`、`frontend/src/app/interview-context/roundAssumptions.ts`、`frontend/src/app/screens/{parse,home,workspace}/*`、`frontend/src/app/navigation/interviewContext.ts`、`ui-design/src/*` 和 P0.016/P0.018 scenario wrappers。
- **BDD 策略**: Feature plan requires BDD；当前 BDD gate 为 `E2E.P0.014`、`E2E.P0.015`、`E2E.P0.016`，v2.7 追加 `E2E.P0.018` 作为 workspace 列表进入统一详情的回访 gate；v2.10 继续使用 `E2E.P0.016` 证明结构化轮次在 Parse、Home recent rail 和 shared navigation context 中同源消费，并要求 Playwright 截图附件或 `screenshotBytes=` marker 作为验收证据。
- **替代验证 gate**: 不适用；本计划具备 TDD + BDD 双层验证。

## 4 当前实现合同

### 4.1 Home

- 渲染 Hero label/title、`home-jd-input-card`、`home-jd-textarea`、输入卡底部 `home-jd-source-controls`、upload/URL source actions、`home-resume-row`、`home-resume-select`、`home-resume-create`、`home-submit-row` 与 `home-jd-submit`。
- `listResumes` 只把 ready 且可用的简历作为下拉选项；用户未显式选择简历时，paste / upload / URL import 均不得提交。
- paste 提交 `ImportTargetJobRequest.source.type=manual_text` + selected `resumeId`；upload 先 `createUploadPresign(purpose=target_job_attachment)`，再提交 `source.type=file` + selected `resumeId`；URL 提交 `source.type=url` + selected `resumeId`。
- `createUploadPresign`、`importTargetJob` 都必须通过 generated client 发送，并携带 side-effect idempotency key。
- 成功 import 后导航到 `parse`，params 必须包含 `targetJobId`、source 与真实 `resumeId`。
- `listTargetJobs` 请求必须带 `analysisStatus=ready`，UI 层防御性排除 failed / processing / queued / 空标题 TargetJob，只渲染最近 3 张模拟面试卡片，排序按 `updatedAt desc`；卡片 grid 使用固定最大列宽，单卡不得被 `1fr` 拉伸；`MockInterviewCard` 主体也被 workspace 面试列表复用；Home 卡片点击主体进入统一规划详情，`立即面试` 主按钮启动 practice 并携带结构化 `roundId/roundName`，且不展示删除按钮；`更多` 进入 `workspace`。

### 4.2 Parse

- 进入 Parse 后先展示 4 步 loading gate，再根据 `getTargetJob.analysisStatus` 进入 detail 或 failed state。
- Preview / workspace 回访详情对用户命名为“面试规划详情 / 面试上下文确认”，只读渲染 API response 中的 title、companyName、locationText、requirements、summary、fitSummary、round assumptions、已绑定 resume 与 provenance 信息。
- Round assumptions 的数组长度必须为 2~5，R 序号、标题、轮次类型、时长和 focus 均来自 `TargetJob.summary.interviewRounds[]`。该数组由后端 LLM 根据 JD、岗位级别、公司/行业性质、团队/业务上下文、职责范围、招聘流程线索和同类岗位常见面试实践推断。前端只负责展示 API 保存的 round 数组，不得用 locale 静态文案或本地常量补齐固定 4 轮、固定 HR/技术/经理面类型或固定分钟数。
- Basic fields、requirements evidence、hidden signals、round assumptions 和 resume binding 均不可在详情页修改；详情页不提供 notes 编辑、requirements hit toggle、hidden signal 移除、resume picker、创建简历兜底、重新解析、取消或仅保存规划入口。
- Parse 读取 ready 简历列表仅用于展示已绑定 `resumeId` 摘要；若 TargetJob 缺少有效 `resumeId`，Start 保持 disabled 并展示缺失上下文状态，不从 route-only `resumeId` 补绑简历。
- Start interview 不调用 `updateTargetJob`，直接使用已保存 `targetJobId/resumeId/roundId/currentPracticePlanId` 调 `createPracticePlan` / `getPracticePlan` / `startPracticeSession` 并进入 practice。
- `workspace?targetJobId=...` 普通回访不得强制播放 parse loading；应直接拉取 `getTargetJob` 并渲染同一详情母版 ready state。`workspace` 不读取 `autoStartPractice`，也不作为启动副作用路由。

### 4.3 Privacy / Auth

- JD 原文、source URL 与 raw source content 不进入 URL、localStorage、console 或 telemetry。
- 未登录 Home import 使用 opaque pending import id 接续；pending action 不携带 JD 原文或 source URL。
- Parse Start 只有在真实 `resumeId` 已绑定时才触发 auth continuation。
- 前端只允许调用 generated TargetJobs / Uploads / Resumes client；不得直接调用 AI provider、prompt registry、provider-specific endpoint 或 ad hoc parse fetch。

## 5 实施步骤

### Phase 1: Home 当前入口

#### 1.1 UI source parity

Home DOM、布局、控件密度、主题、i18n 与响应式行为对齐 `ui-design/src/screen-home.jsx::HomeScreen`。

#### 1.2 Generated client contract

Home 使用 `listResumes`、`listTargetJobs`、`createUploadPresign` 和 `importTargetJob`。所有 request body、headers、route params 和错误态由 Vitest 覆盖。

#### 1.3 BDD-Gate

验证 `E2E.P0.014` 与 `E2E.P0.015`。

### Phase 2: Parse 当前确认与 handoff

#### 2.1 UI source parity

Parse loading、preview、failed state、resume binding、footer actions 与响应式行为对齐 `ui-design/src/screens-p0-complete.jsx::ParseScreen`。

#### 2.2 Generated client contract

Parse 使用 `getTargetJob`、`listResumes` 和 practice handoff generated client。Readonly detail、同 route target switch、failed state、真实 `resumeId` handoff、无 `updateTargetJob` patch 与 auth continuation 均由 Vitest 覆盖。

#### 2.3 BDD-Gate

验证 `E2E.P0.015` 与 `E2E.P0.016`。

### Phase 3: 收口验证

#### 3.1 Focused frontend gates

运行 Home/Parse focused Vitest、frontend typecheck、fixture validation 与 real-mode generated-client gate。

#### 3.2 Scenario gates

串行运行 `E2E.P0.014`、`E2E.P0.015`、`E2E.P0.016` 的 `setup -> trigger -> verify -> cleanup`。

#### 3.3 Repo gates

运行 context validation、doc index check、docs-check、diff whitespace check 与 core-loop pruning surface lint。

### Phase 4: Import resume binding remediation

#### 4.1 Generated client request contract

Home must include the selected ready `resumeId` in every `importTargetJob` request body for paste, upload and URL sources. Missing resume remains a client-side block before request dispatch.

#### 4.2 Route continuity

Successful import still navigates to Parse with `targetJobId`, source and `resumeId`, but route params are not a binding fallback; the backend TargetJob response is authoritative after reload or list re-entry.

#### 4.3 BDD-Gate

`E2E.P0.015` must continue to cover Home import request shape and privacy behavior, with `resumeId` treated as an allowed business identifier and JD raw text/source secrets still excluded from URL/pending action storage.

### Phase 5: Unified plan detail remediation

#### 5.1 UI truth source and copy

Rename the Parse preview user-facing concept from "JD parse result" to "Interview Plan Detail / Context Confirm" in `ui-design/src/screens-p0-complete.jsx`, `docs/ui-design/`, formal locales and pixel parity expectations, while keeping the 4-step parse loading state for first import only.

#### 5.2 Shared route implementation

Refactor the Parse-derived detail so `route=parse` after loading and `route=workspace` with `targetJobId` render the same DOM structure, fields, readonly resume binding and Start action. Workspace no-context list remains in `WorkspacePlanList`; practice startup is triggered only by the explicit Start action from the readonly detail.

#### 5.3 Route context and out-of-scope negative

Stop fabricating `plan-${targetJobId}` or `resume-unbound` from shared detail navigation; use declared `TargetJob.currentPracticePlanId` / `TargetJob.resumeId` when present, omit absent IDs, and add negative coverage for the out-of-scope independent workspace detail anchors.

#### 5.4 BDD-Gate

`E2E.P0.016` must continue to prove first-import detail Start handoff, and `E2E.P0.018` must prove workspace list card re-entry lands on the same unified detail mother page rather than a second workspace detail page.

### Phase 6: Readonly plan detail simplification

#### 6.1 UI truth source and copy

Update `ui-design/src/screens-p0-complete.jsx`, `docs/ui-design/` and locales so the success detail is a readonly context receipt: API-derived fields, requirement evidence, hidden signals, round assumptions and bound resume are display-only. The only success footer action is Start interview.

#### 6.2 Generated client contract

Remove Parse success-detail PATCH behavior. Focused tests must prove Start interview does not call `updateTargetJob`, uses the bound resume from TargetJob / route, and blocks only when the saved plan is missing a usable bound resume.

#### 6.3 Removed controls negative gate

Vitest, pixel parity and scenario gates must assert the absence of editable inputs, requirements toggles, hidden-signal remove controls, resume picker / create-resume fallback, success-state Re-parse, Save plan and Cancel controls.

#### 6.4 BDD-Gate

`E2E.P0.016` must prove the readonly plan receipt and direct Start handoff. `E2E.P0.018` must continue to prove workspace list re-entry lands on the same readonly detail mother page.

### Phase 7: LLM-derived round assumptions shared data binding

#### 7.1 UI truth source and formal contract

Historical note: this phase first moved Parse/Home/navigation off purely local copy and onto backend-provided round-assumption data. Phase 8 supersedes the string-only shape with structured `TargetJob.summary.interviewRounds[]`; current UI truth no longer uses `TargetJob.summary.interviewHypotheses`, fixed four-card assumptions, or missing-slot static fallback.

#### 7.2 Frontend TDD

The current focused regression coverage proves `parse-round-*` cards and `home-recent-mock-rail-*` labels render backend-provided structured rounds when present and do not use static `parse.round*Focus` / `DEFAULT_ROUNDS` strings for those slots.

#### 7.3 Shared implementation

Replace per-surface static round arrays with a shared TargetJob round assumption mapper consumed by Parse detail, Home recent mock cards, and `interviewContextFromTargetJob` route params. Workspace plan cards remain compact, but their open-plan handoff must not fabricate a conflicting static round name.

#### 7.4 BDD-Gate

`E2E.P0.016` / focused equivalent must cover the readonly detail and related Home recent card surface showing saved `TargetJob.summary.interviewRounds[]` count/type/name/duration/focus in round assumptions, with no static four-round fallback in positive structured data paths.

### Phase 8: Structured LLM-derived interview rounds

#### 8.1 Contract and prompt schema

Upgrade `target.import.parse` output, OpenAPI `TargetJobSummary`, fixtures and generated Go/TS artifacts from string-only `interviewHypotheses` to structured `interviewRounds[]`. The array must contain 2~5 rounds. Each round must carry `sequence`, `type`, `name`, `durationMinutes` and `focus`; the LLM parse result is authoritative for round count, round type/name and duration and must be inferred from JD evidence plus role seniority, company/industry nature, team/business context, hiring-process hints and common interview practices for similar roles.

#### 8.2 Backend parser and persistence

Update backend targetjob parse executor and tests so successful JD parse validates and persists structured rounds into `target_jobs.summary.interviewRounds[]`, preserving provenance and rejecting malformed round entries instead of silently fabricating default rounds.

#### 8.3 Frontend structured round mapper

Update Parse detail, Home recent card rail and `interviewContextFromTargetJob` to consume `summary.interviewRounds[]` directly. Focused tests must prove variable round counts and variable durations render from fixtures, and hardcoded strings such as `HR 初筛 · 20m` are not used when structured rounds exist.

#### 8.4 UI truth source and BDD gate

Update `ui-design/`, `docs/ui-design/module-job-workspace.md` and `E2E.P0.016` so the visible contract is structured LLM rounds: 2~5 rounds, inferred type/name, inferred duration and inferred focus across Parse, Home recent cards and shared navigation context. The browser acceptance path must attach a screenshot or emit a positive `screenshotBytes=` marker while asserting the rendered round cards.

### Phase 9: Recent card fixed grid and workspace fusion

#### 9.1 UI truth source

Update `ui-design/src/screen-home.jsx` and `docs/ui-design/` so Home recent mock cards use the same fixed maximum column width as the workspace plan list. A single recent card must not stretch to fill the row.

#### 9.2 Shared implementation

Extend `MockInterviewCard` as the shared card body for Home recent cards and workspace plan-list cards. Home keeps card-click navigation and no footer; workspace passes workspace-owned testids and appends an `Open plan` / `进入规划` footer CTA.

#### 9.3 Regression gates

Focused tests must prove `home-recent-mock-grid` and `workspace-plan-list-grid` reject `1fr` stretching, workspace cards expose `workspace-plan-list-rail-*`, and `MockInterviewCard` supports workspace testids/footer without changing Home recent semantics.

### Phase 10: Home recent shared action card

#### 10.1 UI truth source

Update `ui-design/src/screen-home.jsx` and `docs/ui-design/` so Home recent cards reuse the Interview list card action model: card body click opens the unified plan detail, footer shows `立即面试 / Start interview now`, and the delete icon is absent on Home.

#### 10.2 Shared implementation

Extend `MockInterviewCard` with reusable action props so Home can pass a quick-start action without a delete action, while Workspace can pass both quick-start and delete actions.

#### 10.3 Regression gates

Focused tests must prove Home recent cards show the quick-start action, do not show delete controls, request/filter ready TargetJob records only, and quick-start uses the generated practice handoff with structured `roundId/roundName` instead of navigating to the planning detail.

### Phase 11: P0.014 executable-evidence reconciliation

`E2E.P0.014` 只声明 trigger 实际执行的 real-mode generated-client routing test 与五个 Home Vitest 文件：Home shell/control/i18n、source/resume/submit layout、resume selection、recent fixture variants/filter/sort/cap/More/quick-start 和 shared card。TopBar、theme、mobile layout、frontend build、Playwright 与 live backend 不属于该场景证据；browser-level Home parity 继续由 frontend-shell/003 的当前 browser gate 承接。

### Phase 12: Pending-import test API removal

The in-memory pending import store exposes only the production `storePendingImportSource` and one-shot `consumePendingImportSource` operations. Remove `clearPendingImportSourcesForTests` and its redundant teardown call: the sole test-created entry is consumed by the authenticated continuation path, and later tests cannot address an unknown generated id. A source negative gate prevents test-only reset APIs from returning to the production module.

### Phase 13: Current fixture inventory wording

Align the BDD closeout checklist with the current B2 truth source: `make validate-fixtures` covers 37 operations. This is a documentation-only inventory correction; Home/Parse scenarios, fixtures, generated clients and runtime behavior remain unchanged.

### Phase 14: Home copy-table orphan cleanup

删除 `ui-design/src/screen-home.jsx` 中定义但未渲染的 `uploadSourceSub` 双语属性，以及正式 locale catalog / 自证测试中的同名孤儿 key；Home DOM、可见 copy 与交互保持不变。

### Phase 15: MiniRoundRail prototype call-surface pruning

`MiniRoundRail` 只消费主题 token、结构化 `rounds` 与 `currentIndex`；轮次名称和时长已由 `TargetJob.summary.interviewRounds[]` 提供，不从 `lang` 推导任何内容。删除从未读取的 `lang` 形参与唯一调用方传参，保留轮次数量、名称、时长和当前轮高亮，不增加空转参数或 wrapper。

门禁：UI contract 先对当前冗余 rail 签名和调用方传参失败，删除后以 AST 证明 `MiniRoundRail` 参数全部有读取点；focused Home、P0.014/P0.016、静态浏览器 Home rail、full frontend、typecheck/build、owner contexts 与 docs/diff/pruning gates 通过。BDD 不适用，因为本批不改变 Home recent 的可见内容、结构化轮次或导航行为。

### Phase 16: Home/Parse real-backend verifier convergence

让既有 `frontend-real-backend-verify.sh` 接受可选 owner test 文件参数，默认继续校验 `frontendOwners.realApiMode.test.ts`，并让 P0.014/P0.015/P0.016 显式校验 `targetJob.realApiMode.test.ts`。删除三个 caller 内联的 real-mode、base URL 和通用 Vitest summary 解析，以及 P0.015/P0.016 中被更强 summary 检查完全覆盖的 PASS grep；保留每个场景的固定 spec 文件、业务 marker、隐私与 out-of-scope 断言。

门禁：共享 helper 参数行为与三个 caller source contract 先红后绿；P0.014/P0.015/P0.016 的 setup/trigger/verify/cleanup、owner/product contexts、docs/diff/pruning gates 通过。BDD 不适用，因为 trigger 测试集合、场景业务断言、浏览器覆盖和环境生命周期均不改变。

## 6 验收标准

- Home/Parse owner 文档只描述当前 Home + Parse 合同、operation matrix、BDD gate 和验证入口。
- `context.yaml` 只列当前正向 route、operationId、source package 与场景目录。
- `E2E.P0.014` / `E2E.P0.015` / `E2E.P0.016` 场景文档和脚本覆盖当前 Home/Parse 主路径、失败路径、privacy gate 与 real-mode generated-client gate。
- Home import request bodies include the selected `resumeId`, and backend list/detail must recover the binding without depending on transient Parse route params.
- Parse and workspace detail routes share the same "面试规划详情 / 面试上下文确认" page structure, copy, resume binding and action semantics; workspace no longer renders an independent full-page current-plan confirmation.
- Parse round assumptions, Home recent mock rails and shared TargetJob navigation context display/use backend/LLM `TargetJob.summary.interviewRounds[]`; round count is 2~5, and type/name, duration and focus are not front-end fixed values.
- Home recent mock cards and workspace plan-list cards share the same `MockInterviewCard` body, mini round rail, fixed max-width grid and quick-start action; quick-start preserves structured `roundId/roundName`; Home omits delete controls while workspace includes them.
- The pending import module exposes no test-only reset API; Home auth continuation tests cover one-shot store/consume behavior and privacy unchanged.
- P0.014/P0.015/P0.016 reuse the shared real-backend verifier with the TargetJob generated-client owner test while retaining scenario-specific evidence checks.
- `sync-doc-index --check`、`make docs-check`、`git diff --check` 和 `make lint-core-loop-pruning-surface` 通过。

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-10 | 2.23 | Reuse the shared real-backend verifier across the three Home/Parse scenarios. |
| 2026-07-10 | 2.22 | Remove the unread MiniRoundRail language prop and caller argument. |
| 2026-07-10 | 2.21 | Remove the unrendered Home upload-source subtitle from prototype and locale assets. |
| 2026-07-10 | 2.20 | Align the BDD fixture gate wording with the current 37-operation OpenAPI contract. |
| 2026-07-10 | 2.19 | Remove the redundant pending-import test reset API and teardown. |
| 2026-07-10 | 2.18 | Align P0.014 scenario and BDD claims with its generated-client and Home Vitest runner evidence. |
| 2026-07-10 | 2.17 | Normalize workspace detail out-of-scope and hardcoded-round negative wording without behavior changes. |
| 2026-07-10 | 2.16 | Align workspace detail and fixed-round negative wording without behavior changes. |
| 2026-07-10 | 2.15 | Parse success detail ignores route-only `resumeId` for binding; Start requires saved TargetJob binding. |
| 2026-07-09 | 2.13 | Review remediation: constrain Home recent to ready TargetJobs and preserve structured round params in quick-start practice handoff. |
| 2026-07-09 | 2.12 | Reopen owner plan so Home recent cards reuse the Interview list action card, show quick start, and omit delete. |
| 2026-07-09 | 2.11 | Reopen owner plan to give Home recent cards a fixed max-width grid and share `MockInterviewCard` with the workspace plan list. |
| 2026-07-09 | 2.10 | Reopen owner plan to upgrade round assumptions from string-only `interviewHypotheses` focus text to structured LLM-derived `interviewRounds[]` covering round count, type/name, duration and focus. |
| 2026-07-09 | 2.9 | Reopen owner plan to bind Parse, Home recent card rails and shared TargetJob navigation context to backend-generated `summary.interviewHypotheses` instead of static round focus text. |
| 2026-07-09 | 2.7 | Reopen owner plan to unify Parse preview and workspace current-plan detail into one Interview Plan Detail / Context Confirm mother page. |
