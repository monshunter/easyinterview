# 001 Home + JD Import + Parse

> **版本**: 2.27
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本计划交付当前 Home + Parse 新建模拟面试入口，并在 v2.10 原地修订中把轮次假设升级为结构化 LLM/JD parse 合同：Parse 详情、Home 最近模拟面试卡片、Workspace 规划回访 handoff 和共享导航上下文必须使用同一份 TargetJob structured round mapper；卡片视觉仍复刻 UI 真理源，但轮次数量必须为 2~5，轮次类型、标题、时长和 focus 都必须来自后端保存的 `TargetJob.summary.interviewRounds[]`。v2.11-v2.14 继续收敛 recent card、workspace 详情与 quick-start。Phase 17 删除 Parse loading 内部元数据。Phase 18 把 Home JD intake 收敛为唯一粘贴文本框。Phase 19 在 Parse 内容区标题行右上角增加页面级“面试报告”入口，移除嵌入式列表、列表请求和 `section=reports` 兼容逻辑，并把独立列表交给现有 report owner。

交付后的当前链路：

```text
Home pasted JD + ready Resume
  -> importTargetJob({ rawText, targetLanguage, resumeId })
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
| `importTargetJob` | `openapi/fixtures/TargetJobs/importTargetJob.json`（paste success + validation/failure） | Home `{ rawText, targetLanguage, resumeId }` submit | `backend-targetjob` | TargetJob create + saved `resume_id` + parse job；无 source-specific side branch | backend-only | `E2E.P0.015` |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` | Parse polling + readonly preview, including structured rounds from `summary.interviewRounds[]` | `backend-targetjob` | TargetJob read from `target_jobs.summary` | backend-only `target.import.parse` structured round result | `E2E.P0.015` / `E2E.P0.016` |
| `createPracticePlan` / `getPracticePlan` / `startPracticeSession` | `openapi/fixtures/PracticePlans/*`, `openapi/fixtures/PracticeSessions/*` | Parse readonly detail Start action and Home recent quick start | `backend-practice` | PracticePlan / PracticeSession create-read-start | none | `E2E.P0.016` |

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `contract`
- **TDD 策略**: 通过 `/implement` -> `/tdd` 执行。Phase 18 先让 UI contract、Home Vitest、OpenAPI lint/fixture/generated drift、backend TargetJob tests 与 scenario contract 对旧多入口和旧 discriminator 失败，再最小删除 UI/modal/i18n、contract/generated、backend 分支与专属 scenario，最后重构 paste-only 提交 helper 与 opaque-ID one-shot auth vault；auth RED/GREEN 必须覆盖正常 consume、refresh/lost、expired 与 duplicate consume。Phase 19 先让 prototype/source contract、Parse component/effect/route tests 对缺少页面级入口和仍存在的嵌入列表失败，再最小新增入口并删除列表请求、嵌入 DOM 与 `section=reports` 逻辑。
- **BDD 策略**: Feature plan requires BDD；当前 BDD gate 为 `E2E.P0.014`、`E2E.P0.015`、`E2E.P0.016` 与 `E2E.P0.018`。Phase 18 原地修订 P0.014/P0.015 为 paste-only；Phase 19 原地扩展 P0.016 为规划详情右上角入口与 Parse 零嵌入/零列表请求。两者均要求 1440×900 / 390×844 DOM/style/bbox/viewport/screenshot 证据；不创建 sibling 场景或全局报告中心。
- **替代验证 gate**: 不适用；本计划具备 TDD + BDD 双层验证。

## 4 当前实现合同

### 4.1 Home

- 渲染 Hero label/title、只含 `home-jd-textarea` 的 `home-jd-input-card`、`home-resume-row`、`home-resume-select`、`home-resume-create`、`home-submit-row` 与 `home-jd-submit`；旧 source control / trigger / modal DOM 必须不存在。
- `listResumes` 只把 ready 且可用的简历作为下拉选项；JD 为空或用户未显式选择简历时不得提交。
- Home 只通过 generated client 提交 `importTargetJob({ rawText, targetLanguage, resumeId })`，并携带 side-effect idempotency key。
- 成功 import 后导航到 `parse`，params 只包含 `targetJobId` 与真实 `resumeId`；不把 raw JD 或 intake 类型写入 route。
- `listTargetJobs` 请求必须带 `analysisStatus=ready`，UI 层防御性排除 failed / processing / queued / 空标题 TargetJob，只渲染最近 3 张模拟面试卡片，排序按 `updatedAt desc`；卡片 grid 使用固定最大列宽，单卡不得被 `1fr` 拉伸；`MockInterviewCard` 主体也被 workspace 面试列表复用；Home 卡片点击主体进入统一规划详情，`立即面试` 主按钮启动 practice 并携带结构化 `roundId/roundName`，且不展示删除按钮；`更多` 进入 `workspace`。

### 4.2 Parse

- 进入 Parse 后先展示 4 步 loading gate，再根据 `getTargetJob.analysisStatus` 进入 detail 或 failed state。
- Preview / workspace 回访详情对用户命名为“面试规划详情 / 面试上下文确认”，只读渲染 API response 中的 title、companyName、locationText、requirements、summary、fitSummary、round assumptions、已绑定 resume 与 provenance 信息。
- Round assumptions 的数组长度必须为 2~5，R 序号、标题、轮次类型、时长和 focus 均来自 `TargetJob.summary.interviewRounds[]`。该数组由后端 LLM 根据 JD、岗位级别、公司/行业性质、团队/业务上下文、职责范围、招聘流程线索和同类岗位常见面试实践推断。前端只负责展示 API 保存的 round 数组，不得用 locale 静态文案或本地常量补齐固定 4 轮、固定 HR/技术/经理面类型或固定分钟数。
- Basic fields、requirements evidence、hidden signals、round assumptions 和 resume binding 均不可在详情页修改；详情页不提供 notes 编辑、requirements hit toggle、hidden signal 移除、resume picker、创建简历兜底、重新解析、取消或仅保存规划入口。
- Parse 读取 ready 简历列表仅用于展示已绑定 `resumeId` 摘要；若 TargetJob 缺少有效 `resumeId`，Start 保持 disabled 并展示缺失上下文状态，不从 route-only `resumeId` 补绑简历。
- Start interview 不调用 `updateTargetJob`，直接使用已保存 `targetJobId/resumeId/roundId/currentPracticePlanId` 调 `createPracticePlan` / `getPracticePlan` / `startPracticeSession` 并进入 practice。
- `workspace?targetJobId=...` 普通回访不得强制播放 parse loading；应直接拉取 `getTargetJob` 并渲染同一详情母版 ready state。`workspace` 不读取 `autoStartPractice`，也不作为启动副作用路由。
- TargetJob ready 后在详情标题行右上角渲染“面试报告”页面级入口；点击精确导航 `{ name: "reports", params: { targetJobId } }`，不在 TopBar 增加入口，也不把 report/status/round 写入 handoff。
- Parse 不消费 `listTargetJobReports`，不渲染 per-round reports section，也不保留报告列表的 loading/error/empty state。列表数据、current/latest 状态和 report/generating 链接由 `frontend-report-dashboard` 的独立 ReportsScreen 负责。
- Parse route 不接受 `section=reports` 或其他报告 query；旧 section 锚点、滚动/聚焦 effect、兼容解析与测试 helper 必须删除。未知 section 由 shared route filter 丢弃，不能影响 TargetJob identity 或业务状态。

### 4.3 Privacy / Auth

- JD 原文不进入 URL、localStorage、console 或 telemetry。
- 未登录 Home import 的 `pendingAction` 只携带 `opaquePendingImportId`；exact `{ rawText, targetLanguage, resumeId }` intent、同一次 import 的 idempotency key 与 expiry 只存在于当前进程的一次性内存 vault，登录成功后原子 consume 一次。
- refresh / 进程重启导致 vault 丢失、entry 过期或 duplicate consume 时必须 fail closed：不调用 `importTargetJob`，清除无效 pending action，返回 Home 并显示本地化重新粘贴 JD / 选择简历提示；不得把 raw JD 或 vault entry 写入 URL、localStorage、sessionStorage、IndexedDB、console 或 telemetry。
- Parse Start 只有在真实 `resumeId` 已绑定时才触发 auth continuation。
- 前端只允许调用当前 generated TargetJobs / Resumes client；Resume 自己的 upload consumer 继续留在 Resume owner。Home 不得直接调用 upload、AI provider、prompt registry、provider-specific endpoint 或 ad hoc parse fetch。

## 5 实施步骤

### Phase 1: Home 当前入口

#### 1.1 UI source parity

Home DOM、布局、控件密度、主题、i18n 与响应式行为对齐 `ui-design/src/screen-home.jsx::HomeScreen`。

#### 1.2 Generated client contract

Home 使用 `listResumes`、`listTargetJobs` 和 `importTargetJob`。所有 request body、headers、route params 和错误态由 Vitest 覆盖。

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

Home must include the selected ready `resumeId` in the single `importTargetJob({ rawText, targetLanguage, resumeId })` request. Missing raw text or resume remains a client-side block before request dispatch.

#### 4.2 Route continuity

Successful import navigates to Parse with `targetJobId` and `resumeId`; route params are not a binding fallback, and the backend TargetJob response is authoritative after reload or list re-entry.

#### 4.3 BDD-Gate

`E2E.P0.015` must cover the exact paste-only request shape and privacy behavior, with `resumeId` treated as an allowed business identifier and JD raw text excluded from URL/pending action storage.

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

### Phase 17: Parse loading internal-metadata removal

先更新 `ui-design/src/screens-p0-complete.jsx::ParseScreen` 与对应 UI 文档，删除 loading footer 中的 model/provider、rubric/prompt/version/hash、provenance 与 typical latency；保留当前四步进度、等待说明、布局和响应式节奏。随后用 RED-GREEN 同步正式 `ParseScreen`，删除同类硬编码与可见 DOM，不改 `getTargetJob` 轮询、ready/failed 分支或 API 合同。

门禁：UI source contract 与正式 DOM 测试先对内部元数据失败后转绿；`E2E.P0.015` 在 1440 desktop 与 390 mobile 捕获 loading 截图并做 source/formal DOM、computed style、bbox 与 viewport parity，截图和 active source negative search 均不得出现上述内部标记。

### Phase 18: Paste-only Home JD intake

#### 18.1 UI truth source and documentation

先更新 `ui-design/src/screen-home.jsx` 与 `docs/ui-design/`：Home 输入卡只保留 textarea，ready Resume 下拉框与「立即面试」CTA 保持当前布局；删除平行 intake 控件、弹窗、双语 copy 和空态中的多入口提示。`ui-design/ui-design-contract.test.mjs` 先红后绿，并固定旧 DOM/testid/copy 为负向。

#### 18.2 OpenAPI and persistence contract

`importTargetJob` 请求收敛为 `{ rawText, targetLanguage, resumeId }`，不再使用 source discriminator。OpenAPI schema、fixtures、generated Go/TS、backend handler/service/store/runner、persistence 与事件 payload 同步删除非当前 intake 分支及来源枚举；`target_jobs.raw_jd_text` 是唯一 JD 原文事实源，不保留 `manual_text` 兼容词汇、来源列或来源表。Resume upload operation、purpose、handler、fixture 与场景保持可用。

#### 18.3 Frontend TDD

RED：Home layout/import/auth/i18n/UI contract/pixel tests 对旧 source controls、modal、额外 locale keys、upload-client call、intake route param，以及 raw-text pending action 失败。GREEN：删除 `JDAssistModal` 及其 tests；`pendingAction` 只保存 `opaquePendingImportId`，一次性内存 vault 保存 exact intent + 原 idempotency key 并原子 consume；正常登录只重放一次，refresh/lost、expired、duplicate consume 均不调用 import 而返回 Home 提示重新输入。成功后仅导航 `targetJobId + resumeId`。REFACTOR：保留一个 paste submit path，不新增 mode enum、兼容 adapter、浏览器持久化或不可达 branch。

#### 18.4 Backend and contract TDD

RED：OpenAPI lint/fixture/generated drift、backend request decode/service/store/runner 和 package-level negative tests先证明公共多源 union、URL fetch/source refresh、JD attachment purpose 与 manual-form branch 仍存在。GREEN：删除当前 source-specific schema、handler、persistence/job/event/config 与专属 scenario；文本成功、validation、idempotency、parse failure/retry、privacy 与 resume binding 必须保持。REFACTOR：共用现有 text parse path，不保留兼容路由或 retired enum。

#### 18.5 BDD, screenshots and zero-reference gate

原地修订 `E2E.P0.014` / `E2E.P0.015`，删除 URL 专属 `E2E.P0.011` 实体目录与 active INDEX 行（编号不复用）。P0.015 覆盖 paste success、当前 4xx/failed、idempotency、privacy 与 Parse loading；P0.014/P0.015 在 1440×900 和 390×844 捕获 Home paste-only 截图并验证 DOM、computed style、bbox、viewport。active truth-source zero-reference gate 必须扫描 `ui-design/`、`docs/ui-design/`、owner specs/plans、OpenAPI/generated、frontend Home、backend TargetJob 与 active scenarios，排除 work-journal/bug/report 等合法历史证据，并明确允许 Resume upload 资产。

### Phase 19: Plan-detail report entry and independent-list handoff

#### 19.1 Prototype and UI contract

先在 `ui-design` Parse ready state 内容区标题行右上角加入“面试报告”页面级入口，并明确它不属于全局 TopBar。删除 Parse 中既有 Reports section；desktop/mobile DOM、style、bbox、viewport 先红后绿，且入口必须在 1440×900 / 390×844 下不挤压标题与说明。

#### 19.2 Generated contract and mapper

Parse 只从已验证的当前 TargetJob 取得 `targetJobId` 并导航到 `reports`；删除 Parse 内 `listTargetJobReports` 调用、overview loader/validator/render state 和相关 i18n。仓库负向 gate 证明 list operation 的正式 UI consumer 只位于 report owner，Parse effect 与测试 spy 的调用数为零。

#### 19.3 Interaction and route recovery

入口在可信 ready TargetJob 上下文存在时可用，点击后精确进入 `/reports?targetJobId=<uuid>`；不通过 route-only target 覆盖当前事实。删除 `section=reports` safe param、ready 后滚动/聚焦和兼容分支；Report/Generating 的返回路径改由 report owner 进入独立列表，Parse 不承接返回锚点。

#### 19.4 BDD and parity

原地扩展 `E2E.P0.016` 覆盖内容区右上入口、精确 target handoff、全局 TopBar 无报告入口、Parse 无嵌入列表/列表请求/section 兼容，以及 Start/只读详情回归；在 1440×900 / 390×844 对 prototype/formal 入口执行 DOM/computed-style/bbox/viewport/screenshot parity。独立 ReportsScreen 的数据状态与隔离由 report owner P0.058/P0.059 承接。

## 6 验收标准

- Home/Parse owner 文档只描述当前 Home + Parse 合同、operation matrix、BDD gate 和验证入口。
- `context.yaml` 只列当前正向 route、operationId、source package 与场景目录。
- `E2E.P0.014` / `E2E.P0.015` / `E2E.P0.016` 场景文档和脚本覆盖当前 Home/Parse 主路径、失败路径、privacy gate 与 real-mode generated-client gate。
- Home import request bodies include the selected `resumeId`, and backend list/detail must recover the binding without depending on transient Parse route params.
- Parse and workspace detail routes share the same "面试规划详情 / 面试上下文确认" page structure, copy, resume binding and action semantics; workspace no longer renders an independent full-page current-plan confirmation.
- Parse round assumptions, Home recent mock rails and shared TargetJob navigation context display/use backend/LLM `TargetJob.summary.interviewRounds[]`; round count is 2~5, and type/name, duration and focus are not front-end fixed values.
- Home recent mock cards and workspace plan-list cards share the same `MockInterviewCard` body, mini round rail, fixed max-width grid and quick-start action; quick-start preserves structured `roundId/roundName`; Home omits delete controls while workspace includes them.
- The pending import module exposes no test-only reset API；`pendingAction` only carries `opaquePendingImportId`，while raw JD remains in a process-memory one-shot vault. Home auth continuation tests cover normal atomic consume, refresh/lost vault, expiry and duplicate consume；only the normal path dispatches one exact request with the original idempotency key.
- P0.014/P0.015/P0.016 reuse the shared real-backend verifier with the TargetJob generated-client owner test while retaining scenario-specific evidence checks.
- Parse loading 只展示用户可理解的进度/等待状态；prototype、formal、desktop/mobile 截图和 active source 均不含 model/provider、rubric/prompt/version/hash、provenance 或 typical latency。
- Home 只展示 JD textarea、ready Resume 下拉框和主 CTA；`importTargetJob` 只接受 `{ rawText, targetLanguage, resumeId }`，route 只携带 `targetJobId + resumeId`。
- 非当前 JD intake 的 UI、public schema、generated type、backend branch、专属 fixture/scenario 和 active docs 为零；Resume 上传路径继续通过原 owner gates。
- Parse 内容区标题行右上角展示“面试报告”，点击精确进入 `reports?targetJobId=...`；全局 TopBar 无该入口，既有只读详情与 Start 保持可用。
- Parse 中 `listTargetJobReports` 调用、嵌入式报告 DOM、列表状态、`section=reports` safe param 与滚动/聚焦兼容逻辑为零；独立 ReportsScreen 与返回路径由 report/shell owner 验证。
- `sync-doc-index --check`、`make docs-check`、`git diff --check` 和 `make lint-core-loop-pruning-surface` 通过。

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 2.27 | Revise Phase 19 in place: move reports to an independent target-scoped page, keep only the plan-detail header entry, and delete Parse list requests, embedded UI, and section compatibility. |
| 2026-07-14 | 2.26 | Add unchecked Phase 19 for the Parse canonical-round reports section, minimal overview mapper, typed state links, non-blocking failure, section anchor and P0.016 parity. |
| 2026-07-13 | 2.25 | Reopen in place to make Home JD intake paste-only across UI, contract, backend and scenarios, with exact request shape, zero-reference gates and desktop/mobile screenshots. |
| 2026-07-13 | 2.24 | Reopen in place to remove internal parse model/rubric/provenance/latency metadata and require clean desktop/mobile loading screenshots. |
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
