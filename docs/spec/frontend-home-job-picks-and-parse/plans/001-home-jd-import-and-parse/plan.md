# 001 Home + JD Import + Parse

> **版本**: 2.32
> **状态**: completed
> **更新日期**: 2026-07-15

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本计划交付当前 Home + Parse 新建模拟面试入口，并维护统一 TargetJob structured round mapper。Phase 20 最终锁定 route 分工与请求基线：POST import 后只进入 `/parse?targetJobId`；Parse 只承载 queued/processing；ready 使用 replace 进入 `/workspace?targetJobId`；Home ready 卡片直接进入 workspace detail；Home/list safe-read 依赖 shell single-flight，StrictMode 下同 key 初载底层请求为 1。

交付后的当前链路：

```text
Home pasted JD + ready Resume
  -> importTargetJob({ rawText, targetLanguage, resumeId })
  -> /parse?targetJobId queued/processing
  -> replace /workspace?targetJobId
  -> Unified Plan Detail / Context Confirm
  -> Start interview
  -> practice
```

## 2 背景

`frontend-shell` 提供 App 壳、route normalization、auth pending action、runtime config、generated client bootstrap 与 fixture-backed transport。本 owner 负责 `home`、`parse` loading 和统一详情母版。Workspace 无上下文列表仍归 `frontend-workspace-and-practice`；详情页 Start action 直接调用 practice handoff helper，不再通过 workspace auto-start route 制造副作用。

UI 必须源级追溯到 `frontend/src` 与 `frontend/src`。正式前端只允许为真实数据、generated client、鉴权接续和可访问性做工程适配。

当前 API 合同来自 `openapi/openapi.yaml` 与 fixtures：

| operationId | Fixture | Frontend consumer | Backend owner | Persistence | AI dependency | Scenario |
|-------------|---------|-------------------|---------------|-------------|---------------|----------|
| `listTargetJobs` | current list fixture | Home recent cards | backend-targetjob | TargetJob read | none | `E2E.P0.098` 仅 Home progress refresh |
| `listResumes` | current list fixture | Home resume select | backend-resume | Resume read | none | 当前无真实 E2E owner；root `make test` |
| `importTargetJob` | paste success/failure fixtures | Home submit | backend-targetjob | TargetJob + parse job | backend parse | 当前无真实 E2E owner；root `make test` |
| `getTargetJob` | current detail/progress fixture | Parse polling + Workspace detail | backend-targetjob | TargetJob summary/progress | backend parse output | `E2E.P0.098` 仅 progress/detail read；import/parse 无 owner |
| `createPracticePlan` / `getPracticePlan` / `startPracticeSession` | current practice fixtures | Workspace/Home start | backend-practice | plan/session | none at frontend | 当前无真实 E2E owner；root `make test` |

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `contract`
- **TDD 策略**: focused UI/contract/backend tests provide development feedback；阶段完成由根 `make test` 承接。Parse command progress 与 Workspace detail 仍通过独立 UI/parity gates 验证。
- **BDD 策略**: Home JD import/parse 保留 Given/When/Then 合同但当前无真实 E2E owner；`E2E.P0.098` 只覆盖 completion 后 Home/Workspace/TargetJob progress refresh/detail read。
- **替代验证 gate**: OpenAPI/codegen/fixture drift、UI parity、typecheck/build、privacy and stale-contract negative searches。

## 4 当前实现合同

### 4.1 Home

- 渲染 Hero label/title、只含 `home-jd-textarea` 的 `home-jd-input-card`、`home-resume-row`、`home-resume-select`、`home-resume-create`、`home-submit-row` 与 `home-jd-submit`；旧 source control / trigger / modal DOM 必须不存在。
- `listResumes` 只把 ready 且可用的简历作为下拉选项；JD 为空或用户未显式选择简历时不得提交。
- Home 只通过 generated client 提交 `importTargetJob({ rawText, targetLanguage, resumeId })`，并携带 side-effect idempotency key。
- 成功 import 后导航到 `/parse?targetJobId=...`；不把 `resumeId`、raw JD 或 intake 类型写入 route。
- `listTargetJobs` 请求必须带 `analysisStatus=ready`；Home 卡片主体直接进入 `/workspace?targetJobId=...`，不经过 Parse/动画；quick-start 与「更多」保持现有语义。`listTargetJobs` / `listResumes` 依赖 shell safe-read single-flight 与稳定 effect dependencies，同 key 初载底层 request count 各为 1。

### 4.2 Parse

- Parse 只在 queued/processing 展示 4 步 loading gate；首读 ready 或轮询转 ready 时立即 replace 到 `/workspace?targetJobId=...`，不在 Parse 渲染 ready detail，Back 不返回动画。
- Workspace ready detail 对用户命名为“面试规划详情 / 面试上下文确认”，只读渲染 API response 中的 title、companyName、locationText、requirements、summary、fitSummary、round assumptions 与已绑定 resume 摘要；详情初载只执行同 key `getTargetJob`，不调用 `listResumes`。
- Round assumptions 的数组长度必须为 2~5，R 序号、标题、轮次类型、时长和 focus 均来自 `TargetJob.summary.interviewRounds[]`。该数组由后端 LLM 根据 JD、岗位级别、公司/行业性质、团队/业务上下文、职责范围、招聘流程线索和同类岗位常见面试实践推断。前端只负责展示 API 保存的 round 数组，不得用 locale 静态文案或本地常量补齐固定 4 轮、固定 HR/技术/经理面类型或固定分钟数。
- Basic fields、requirements evidence、hidden signals、round assumptions 和 resume binding 均不可在详情页修改；详情页不提供 notes 编辑、requirements hit toggle、hidden signal 移除、resume picker、创建简历兜底、重新解析、取消或仅保存规划入口。
- 若 TargetJob 缺少有效 `resumeId`，Workspace detail 的 Start 保持 disabled 并展示缺失上下文状态；不得调用 `listResumes`、不得从 route-only `resumeId` 补绑简历，也不提供 picker/rebind。
- Workspace detail 的 Start interview 不调用 `updateTargetJob`，直接使用已保存 `targetJobId/resumeId/roundId/currentPracticePlanId` 调 `createPracticePlan` / `getPracticePlan` / `startPracticeSession` 并进入 practice。
- `/workspace?targetJobId=...` 普通回访直接拉取一次同 key `getTargetJob` 并渲染详情 ready state；不得 import、poll、播放 Parse loading，也不读取 `autoStartPractice` 或在 route side 启动 session。
- TargetJob ready 后在 Workspace detail 标题旁渲染“绑定简历”查看链接，并在标题下首行动作行从左依次渲染“立即面试 + 面试报告”；报告点击精确导航 `{ name: "reports", params: { targetJobId } }`，绑定链接精确导航 `{ name: "resume_versions", params: { resumeId } }`。不在 TopBar/页尾增加入口，也不把 report/status/round/resume authority 写入 route。
- Parse 与 Workspace detail 都不消费 `listTargetJobReports`、不渲染 per-round reports section，也不保留报告列表 loading/error/empty state。列表数据、current/latest 状态和 report/generating 链接由 `frontend-report-dashboard` 的独立 ReportsScreen 负责。
- Parse/Workspace route 不接受 `section=reports` 或其他报告 query；旧 section 锚点、滚动/聚焦 effect、兼容解析与测试 helper 必须删除。未知 section 由 shared route filter 丢弃，不能影响 TargetJob identity 或业务状态。

### 4.3 Privacy / Auth

- JD 原文不进入 URL、localStorage、console 或 telemetry。
- 未登录 Home import 的 `pendingAction` 只携带 `opaquePendingImportId`；exact `{ rawText, targetLanguage, resumeId }` intent、同一次 import 的 idempotency key 与 expiry 只存在于当前进程的一次性内存 vault，登录成功后原子 consume 一次。
- refresh / 进程重启导致 vault 丢失、entry 过期或 duplicate consume 时必须 fail closed：不调用 `importTargetJob`，清除无效 pending action，返回 Home 并显示本地化重新粘贴 JD / 选择简历提示；不得把 raw JD 或 vault entry 写入 URL、localStorage、sessionStorage、IndexedDB、console 或 telemetry。
- Workspace detail Start 只有在真实 `resumeId` 已绑定时才触发 auth continuation。
- 前端只允许调用当前 generated TargetJobs / Resumes client；Resume 自己的 upload consumer 继续留在 Resume owner。Home 不得直接调用 upload、AI provider、prompt registry、provider-specific endpoint 或 ad hoc parse fetch。

## 5 实施步骤

### Phase 1: Home 当前入口

#### 1.1 UI formal implementation contract

Home DOM、布局、控件密度、主题、i18n 与响应式行为对齐 `frontend/src`。

#### 1.2 Generated client contract

Home 使用 `listResumes`、`listTargetJobs` 和 `importTargetJob`。所有 request body、headers、route params 和错误态由 Vitest 覆盖。



### Phase 2: Parse progress 与 Workspace detail handoff

#### 2.1 UI formal implementation contract

Parse loading/failed state 与 Workspace detail 的只读 resume binding/footer actions/响应式行为对齐 `frontend/src` 的共享视觉；ready DOM 只由 Workspace route 渲染。

#### 2.2 Generated client contract

Parse 只使用 `getTargetJob` 分类/轮询；Workspace detail 使用单次同 key `getTargetJob` 和 practice handoff generated client，不调用 `listResumes`。Readonly detail、target switch、failed state、真实 `resumeId` handoff、无 `updateTargetJob` patch 与 auth continuation 均由 Vitest 覆盖。



### Phase 3: 收口验证

#### 3.1 Focused frontend gates

运行 Home/Parse focused Vitest、frontend typecheck、fixture validation 与 real-mode generated-client gate。



#### 3.3 Repo gates

运行 context validation、doc index check、docs-check、diff whitespace check 与 core-loop pruning surface lint。

### Phase 4: Import resume binding remediation

#### 4.1 Generated client request contract

Home must include the selected ready `resumeId` in the single `importTargetJob({ rawText, targetLanguage, resumeId })` request. Missing raw text or resume remains a client-side block before request dispatch.

#### 4.2 Route continuity

Successful import navigates to Parse with only `targetJobId`; route params never carry or restore `resumeId`, and the backend TargetJob response is authoritative after ready replace, reload or list re-entry.



### Phase 5: Unified plan detail remediation

#### 5.1 UI design document and copy

Rename the shared ready-detail visual from "JD parse result" to "Interview Plan Detail / Context Confirm" in `frontend/src`, `docs/ui-design/`, formal locales and component/responsive expectations；render it only under Workspace while keeping the 4-step Parse loading state for first import only.

#### 5.2 Shared route implementation

Refactor the Parse-derived detail so only `route=workspace` with `targetJobId` renders the ready DOM structure, fields, readonly resume binding and Start action；`route=parse` ready immediately replaces to Workspace. Workspace no-context list remains in `WorkspacePlanList`; practice startup is triggered only by the explicit Start action from the readonly detail.

#### 5.3 Route context and out-of-scope negative

Stop fabricating `plan-${targetJobId}` or `resume-unbound` from shared detail navigation; use declared `TargetJob.currentPracticePlanId` / `TargetJob.resumeId` when present, omit absent IDs, and add negative coverage for the out-of-scope independent workspace detail anchors.



### Phase 6: Readonly plan detail simplification

#### 6.1 UI design document and copy

Update `frontend/src`, `docs/ui-design/` and locales so the Workspace success detail is a readonly context receipt: API-derived fields, requirement evidence, hidden signals, round assumptions and bound resume are display-only. The only success footer action is Start interview.

#### 6.2 Generated client contract

Remove ready-detail PATCH behavior. Focused tests must prove Workspace Start does not call `updateTargetJob`, uses the bound resume from TargetJob only, and blocks only when the saved plan is missing a usable bound resume.

#### 6.3 Removed controls negative gate

Vitest component/responsive assertions and any applicable real API/UI scenario gates must assert the absence of editable inputs, requirements toggles, hidden-signal remove controls, resume picker / create-resume fallback, success-state Re-parse, Save plan and Cancel controls.



### Phase 7: LLM-derived round assumptions shared data binding

#### 7.1 UI design document and formal contract

Historical note: this phase first moved Parse/Home/navigation off purely local copy and onto backend-provided round-assumption data. Phase 8 supersedes the string-only shape with structured `TargetJob.summary.interviewRounds[]`; current UI truth no longer uses `TargetJob.summary.interviewHypotheses`, fixed four-card assumptions, or missing-slot static fallback.

#### 7.2 Frontend TDD

The current focused regression coverage proves `parse-round-*` cards and `home-recent-mock-rail-*` labels render backend-provided structured rounds when present and do not use static `parse.round*Focus` / `DEFAULT_ROUNDS` strings for those slots.

#### 7.3 Shared implementation

Replace per-surface static round arrays with a shared TargetJob round assumption mapper consumed by Parse detail, Home recent mock cards, and `interviewContextFromTargetJob` route params. Workspace plan cards remain compact, but their open-plan handoff must not fabricate a conflicting static round name.



### Phase 8: Structured LLM-derived interview rounds

#### 8.1 Contract and prompt schema

Upgrade `target.import.parse` output, OpenAPI `TargetJobSummary`, fixtures and generated Go/TS artifacts from string-only `interviewHypotheses` to structured `interviewRounds[]`. The array must contain 2~5 rounds. Each round must carry `sequence`, `type`, `name`, `durationMinutes` and `focus`; the LLM parse result is authoritative for round count, round type/name and duration and must be inferred from JD evidence plus role seniority, company/industry nature, team/business context, hiring-process hints and common interview practices for similar roles.

#### 8.2 Backend parser and persistence

Update backend targetjob parse executor and tests so successful JD parse validates and persists structured rounds into `target_jobs.summary.interviewRounds[]`, preserving provenance and rejecting malformed round entries instead of silently fabricating default rounds.

#### 8.3 Frontend structured round mapper

Update Workspace detail, Home recent card rail and `interviewContextFromTargetJob` to consume `summary.interviewRounds[]` directly. Focused tests must prove variable round counts and variable durations render from fixtures, and hardcoded strings such as `HR 初筛 · 20m` are not used when structured rounds exist.



### Phase 9: Recent card fixed grid and workspace fusion

#### 9.1 UI design document

Update `frontend/src` and `docs/ui-design/` so Home recent mock cards use the same fixed maximum column width as the workspace plan list. A single recent card must not stretch to fill the row.

#### 9.2 Shared implementation

Extend `MockInterviewCard` as the shared card body for Home recent cards and workspace plan-list cards. Home keeps card-click navigation and no footer; workspace passes workspace-owned testids and appends an `Open plan` / `进入规划` footer CTA.

#### 9.3 Regression gates

Focused tests must prove `home-recent-mock-grid` and `workspace-plan-list-grid` reject `1fr` stretching, workspace cards expose `workspace-plan-list-rail-*`, and `MockInterviewCard` supports workspace testids/footer without changing Home recent semantics.

### Phase 10: Home recent shared action card

#### 10.1 UI design document

Update `frontend/src` and `docs/ui-design/` so Home recent cards reuse the Interview list card action model: card body click opens the unified plan detail, footer shows `立即面试 / Start interview now`, and the delete icon is absent on Home.

#### 10.2 Shared implementation

Extend `MockInterviewCard` with reusable action props so Home can pass a quick-start action without a delete action, while Workspace can pass both quick-start and delete actions.

#### 10.3 Regression gates

Focused tests must prove Home recent cards show the quick-start action, do not show delete controls, request/filter ready TargetJob records only, and quick-start uses the generated practice handoff with structured `roundId/roundName` instead of navigating to the planning detail.



### Phase 12: Pending-import test API removal

The in-memory pending import store exposes only the production `storePendingImportSource` and one-shot `consumePendingImportSource` operations. Remove `clearPendingImportSourcesForTests` and its redundant teardown call: the sole test-created entry is consumed by the authenticated continuation path, and later tests cannot address an unknown generated id. A source negative gate prevents test-only reset APIs from returning to the production module.

### Phase 13: Current fixture inventory wording

Align the BDD closeout checklist with the current B2 truth source: `make validate-fixtures` covers 37 operations. This is a documentation-only inventory correction; Home/Parse scenarios, fixtures, generated clients and runtime behavior remain unchanged.

### Phase 14: Home copy-table orphan cleanup

删除 `frontend/src` 中定义但未渲染的 `uploadSourceSub` 双语属性，以及正式 locale catalog / 自证测试中的同名孤儿 key；Home DOM、可见 copy 与交互保持不变。

### Phase 15: MiniRoundRail prototype call-surface pruning

`MiniRoundRail` 只消费主题 token、结构化 `rounds` 与 `currentIndex`；轮次名称和时长已由 `TargetJob.summary.interviewRounds[]` 提供，不从 `lang` 推导任何内容。删除从未读取的 `lang` 形参与唯一调用方传参，保留轮次数量、名称、时长和当前轮高亮，不增加空转参数或 wrapper。





### Phase 17: Parse loading internal-metadata removal

先更新 `frontend/src` 与对应 UI 文档，删除 loading footer 中的 model/provider、rubric/prompt/version/hash、provenance 与 typical latency；保留当前四步进度、等待说明、布局和响应式节奏。随后用 RED-GREEN 同步正式 `ParseScreen`，删除同类硬编码与可见 DOM，不改 `getTargetJob` 轮询、ready/failed 分支或 API 合同。


### Phase 18: Paste-only Home JD intake

#### 18.1 UI design document and documentation

先更新 `frontend/src` 与 `docs/ui-design/`：Home 输入卡只保留 textarea，ready Resume 下拉框与「立即面试」CTA 保持当前布局；删除平行 intake 控件、弹窗、双语 copy 和空态中的多入口提示。`scripts/lint/ui_demo_pruning.py` 先红后绿，并固定旧 DOM/testid/copy 为负向。

#### 18.2 OpenAPI and persistence contract

`importTargetJob` 请求收敛为 `{ rawText, targetLanguage, resumeId }`，不再使用 source discriminator。OpenAPI schema、fixtures、generated Go/TS、backend handler/service/store/runner、persistence 与事件 payload 同步删除非当前 intake 分支及来源枚举；`target_jobs.raw_jd_text` 是唯一 JD 原文事实源，不保留 `manual_text` 兼容词汇、来源列或来源表。Resume upload operation、purpose、handler、fixture 与场景保持可用。

#### 18.3 Frontend TDD

RED：Home layout/import/auth/i18n/UI contract/pixel tests 对旧 source controls、modal、额外 locale keys、upload-client call、intake route param，以及 raw-text pending action 失败。GREEN：删除 `JDAssistModal` 及其 tests；`pendingAction` 只保存 `opaquePendingImportId`，一次性内存 vault 保存 exact intent + 原 idempotency key 并原子 consume；正常登录只重放一次，refresh/lost、expired、duplicate consume 均不调用 import 而返回 Home 提示重新输入。成功后 route 仅导航 `targetJobId`。REFACTOR：保留一个 paste submit path，不新增 mode enum、兼容 adapter、浏览器持久化或不可达 branch。

#### 18.4 Backend and contract TDD

RED：OpenAPI lint/fixture/generated drift、backend request decode/service/store/runner 和 package-level negative tests先证明公共多源 union、URL fetch/source refresh、JD attachment purpose 与 manual-form branch 仍存在。GREEN：删除当前 source-specific schema、handler、persistence/job/event/config 与专属 scenario；文本成功、validation、idempotency、parse failure/retry、privacy 与 resume binding 必须保持。REFACTOR：共用现有 text parse path，不保留兼容路由或 retired enum。



### Phase 19: Plan-detail report entry and independent-list handoff

#### 19.1 Prototype and UI contract

本历史阶段先建立 Workspace-only“面试报告”页面级入口并删除既有 Reports section；Phase 23 已 supersede 其标题右侧位置，当前入口必须位于标题下方首行动作行并与“立即面试”左对齐同排。desktop/mobile DOM、style、bbox、viewport 继续保证入口不挤压标题与说明。

#### 19.2 Generated contract and mapper

Workspace detail 只从已验证的当前 TargetJob 取得 `targetJobId` 并导航到 `reports`；删除 shared detail 内 `listTargetJobReports` 调用、overview loader/validator/render state 和相关 i18n。仓库负向 gate 证明 list operation 的正式 UI consumer 只位于 report owner，Parse/Workspace detail effect 与测试 spy 的调用数均为零。

#### 19.3 Interaction and route recovery

入口在可信 Workspace ready TargetJob 上下文存在时可用，点击后精确进入 `/reports?targetJobId=<uuid>`；不通过 route-only target 覆盖当前事实。删除 `section=reports` safe param、ready 后滚动/聚焦和兼容分支；Reports Back 返回 Workspace detail，Report/Generating 的返回路径由 report owner 进入独立列表，Parse 不承接任何 ready 返回锚点。



### Phase 20: Command-only Parse, direct ready detail and exact GET counts

#### 20.1 Route and transition RED-GREEN

POST `importTargetJob` 成功后只导航 `/parse?targetJobId=...`，不得复制 `resumeId` 或 ready detail 状态。Parse 首读 queued/processing 时展示进度并按现有 scheduler 轮询；首读 ready 或任一 tick 转 ready 时调用 `replaceRoute({ name: "workspace", params: { targetJobId } })`。failed/timeout 恢复保持现有合同。

#### 20.2 Direct ready-card detail

Home recent ready card body 直接进入 `/workspace?targetJobId=...`；不得经过 Parse、播放解析动画或创建新的 import/poll。Quick-start 主按钮仍直接走 practice handoff，「更多」仍进入 query-free `/workspace`。

#### 20.3 Request-count and dependency gate

Home `listTargetJobs` / `listResumes` 与 Parse 每个 `getTargetJob` 分类/调度 tick 通过 frontend-shell/001 Phase 13 safe-read single-flight。focused RED/GREEN 必须读取底层 transport spy，而不是 hook invocation count：StrictMode 同 key 初载恰好 1 个底层 GET；后续 polling 只能在 scheduler interval 到期后出现；route/auth/locale/read epoch 变化按 shell 合同产生独立 GET。



### Phase 21: Workspace detail round-state affordance

#### 21.1 Prototype-first state contract

在 `frontend/src` 的 Workspace ready-detail 母版中，复用既有 `eiResolvePracticeProgress` 结果，为每张 round assumption 卡派生 `done/current/pending`。三态分别使用现有 success-soft、accent-soft、neutral-soft token，并显示本地化“已进行 / 即将进行 / 未进行”；不得新增生命周期状态推断或独立 round cursor。

#### 21.2 Formal formal implementation contract

正式 `ParseScreen` 继续只读取 `resolveTargetJobPracticeProgress(targetJob)`：index 小于 `completedCount` 为 done，等于合法 `currentIndex` 为 current，其余为 pending。每张卡必须提供 `data-round-state`、状态文案、不同 background/border；全完成全部 done，无效投影不显示伪造 done/current。DOM、样式与 prototype 一一可追溯。

#### 21.3 Focused and parity gates

先扩展 `ParseEdit.test.tsx` / UI source contract 形成 RED，覆盖进行中、全完成、无效投影和三态 computed style；GREEN 后运行 round mapper、Workspace detail、UI contract、typecheck/build 与 desktop/mobile parity。负向 gate 拒绝从 `TargetJob.status`、URL、localStorage/sessionStorage 推导状态。



### Phase 22: Required runtime JD guard

#### 22.1 Focused Home validation

Home consumes the required `AppRuntimeProvider.contentLimits.targetJobRawTextBytes` field and a shared UTF-8 byte helper. A small injected limit covers ASCII/multibyte acceptance and local rejection with zero import/pending-vault side effects while preserving the textarea DOM/styles。Required 子字段不得单独 fallback；只有整体 runtime source 不可用时才允许沿用既有 bootstrap fallback。

### Phase 23: Workspace detail leading resume link and action row

#### 23.1 RED: current hierarchy and stale controls

扩展 `ParseScreen.test.tsx`、`ParseResumeBinding.test.tsx`、App route tests 与 formal responsive/source contract，使当前标题右侧 Report、独立 `parse-launch` / `parse-resume-binding` block、页尾 Start 先失败；同时保留 targetJobId-only Workspace detail、exact `getTargetJob` transport、轮次三态、Start/Reports route 与缺失绑定 fail-closed 回归。

#### 23.2 GREEN: saved resume viewer beside title

标题 cluster 在“面试规划详情”旁渲染“绑定简历”链接；点击只使用受保护 `TargetJob.resumeId` 导航 `resume_versions?resumeId=...`，不调用 `getResume`/`listResumes`，不从 route/list item/最近简历推断，不提供 picker 或 in-place rebind。缺失/空绑定渲染非链接状态并禁用 Start。删除仅供旧 block 使用的 `parse.launch*`、`parse.resumeBound*`、`parse.footerHint` locale key/test 断言，保留新链接与缺失态所需文案。

#### 23.3 GREEN: leading Start and Reports actions

删除独立 Interview Launch/绑定简历大卡片与页尾 action 区；标题下方首行动作行从左依次渲染“立即面试” primary 和“面试报告” secondary。desktop 同排，mobile 保持 DOM/阅读顺序并在空间不足时换行；Report 只携带可信 `targetJobId`，Start 继续读取 saved resume/current round，启动错误紧邻 action row 且不阻断 Report。

#### 23.4 Gates and owner handoff

Focused tests 只作开发反馈；执行根 `make test`、frontend typecheck/build、desktop/mobile DOM/style/bbox/no-overflow、owner contexts、`sync-doc-index --check`、`make docs-check`、`git diff --check` 与旧标题右侧/独立 block/footer action/orphan locale key 零残留。完成后与 `frontend-workspace-and-practice/001` 同步恢复 completed。

### Phase 24: Required Resume product-contract reconciliation

#### 24.1 Product and design contracts

原地修订 product-scope 与 Home/Resume/Workspace 用户流程，把 selectable Resume 锁定为当前及未来 `importTargetJob`、Practice、Reports、复练和下一轮的强制前置。selectable 延用正式代码合同：未归档且 `parseStatus=ready` 或已有可读正文/结构化证据。无该类简历的用户只进入 `resume_versions(flow=create)`；形成可读证据后返回 Home 显式选择，不自动绑定最近简历。

#### 24.2 Owner and legacy failure contract

Home exact request 保持 `{ rawText, targetLanguage, resumeId }` 且无选择时零 request。历史 TargetJob 缺失或无效 `resumeId` 属于异常数据：Workspace 显示非链接缺失态，Start、Reports、复练和下一轮全部 fail closed；不得提供 picker/rebind、route/storage fallback、JD-only 训练或报告降级。

#### 24.3 Documentation and behavior gates

本阶段只纠正设计合同，不修改 frontend/OpenAPI/backend 实现，也不新建 BDD 文件或伪 E2E。复用 `BDD.HOME.JD.001/002` 和现有 Home/Workspace focused tests证明当前实现已要求简历；以 active docs 负向搜索拒绝“跳过简历训练”“无简历报告”“只阻断 Start”等旧承诺，并执行 context、Header/INDEX、docs links、diff 与 pruning gates。



## 6 验收标准

- Home/Parse owner 文档只描述当前 Home + Parse 合同、operation matrix、BDD gate 和验证入口。
- `context.yaml` 只列当前正向 route、operationId、source package 与场景目录。
- Home import request bodies include the selected `resumeId`, and backend list/detail must recover the binding without depending on transient Parse route params.
- Only `/workspace?targetJobId` renders the "面试规划详情 / 面试上下文确认" page structure, title-adjacent bound-resume viewer and leading Start/Reports action semantics；Parse ready immediately replaces and never renders a second detail.
- Workspace detail round assumptions, Home recent mock rails and shared TargetJob navigation context display/use backend/LLM `TargetJob.summary.interviewRounds[]`; round count is 2~5, and type/name, duration and focus are not front-end fixed values.
- Workspace detail round assumption cards derive done/current/pending from the same persisted `practiceProgress` mapper as list rails and render distinguishable success/accent/neutral backgrounds, borders and localized state labels.
- Home recent mock cards and workspace plan-list cards share the same `MockInterviewCard` body, mini round rail, fixed max-width grid and quick-start action; quick-start preserves structured `roundId/roundName`; Home omits delete controls while workspace includes them.
- The pending import module exposes no test-only reset API；`pendingAction` only carries `opaquePendingImportId`，while raw JD remains in a process-memory one-shot vault. Home auth continuation tests cover normal atomic consume, refresh/lost vault, expiry and duplicate consume；only the normal path dispatches one exact request with the original idempotency key.
- Parse loading 只展示用户可理解的进度/等待状态；prototype、formal、desktop/mobile 截图和 active source 均不含 model/provider、rubric/prompt/version/hash、provenance 或 typical latency。
- Home 只展示 JD textarea、ready Resume 下拉框和主 CTA；`importTargetJob` 只接受 `{ rawText, targetLanguage, resumeId }`，route 只携带 `targetJobId`。
- selectable Resume 是 import、Practice、Reports、复练和下一轮的永久前置；无 selectable 简历只进入创建流程，历史缺失/无效绑定全链路 fail closed，active docs 不得保留 JD-only 降级承诺。
- 非当前 JD intake 的 UI、public schema、generated type、backend branch、专属 fixture/scenario 和 active docs 为零；Resume 上传路径继续通过原 owner gates。
- Workspace detail 标题旁“绑定简历”精确进入 saved Resume 详情，标题下首行动作行从左展示“立即面试 + 面试报告”，后者精确进入 `reports?targetJobId=...`；全局 TopBar、页尾与 Parse command progress 无该入口，独立 launch/binding block 为零。
- Parse 与 Workspace detail 中 `listTargetJobReports` 调用、嵌入式报告 DOM、列表状态、`section=reports` safe param 与滚动/聚焦兼容逻辑为零；独立 ReportsScreen 与返回路径由 report/shell owner 验证。
- POST import 只进入 `/parse?targetJobId`；Home ready card 直达 `/workspace?targetJobId`；Parse ready 使用 replace，Back 不重播动画。
- Home `listTargetJobs` / `listResumes` 与 Parse 每个 `getTargetJob` tick 的同 key 底层 request count 为 1；polling 必须有 scheduler 时间证据。
- `sync-doc-index --check`、`make docs-check`、`git diff --check` 和 `make lint-core-loop-pruning-surface` 通过。

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-15 | 2.32 | Add Phase 24 to reconcile the permanent selectable-Resume prerequisite across product, UI and owner documents without adding a resume-less compatibility path. |
| 2026-07-15 | 2.31 | Reopen Phase 23 to replace the standalone launch/binding block with a title-adjacent saved-resume link and a leading Start/Reports row. |
| 2026-07-14 | 2.28 | Add Phase 20 command-only Parse, ready-card direct Workspace detail, ready replace, Workspace-owned report/start detail language, targetJobId-only routes and exact safe-read GET count gates. |
| 2026-07-14 | 2.27 | Revise Phase 19 in place: move reports to an independent target-scoped page, keep only the plan-detail header entry, and delete Parse list requests, embedded UI, and section compatibility. |
| 2026-07-13 | 2.25 | Reopen in place to make Home JD intake paste-only across UI, contract, backend and scenarios, with exact request shape, zero-reference gates and desktop/mobile screenshots. |
| 2026-07-13 | 2.24 | Reopen in place to remove internal parse model/rubric/provenance/latency metadata and require clean desktop/mobile loading screenshots. |
| 2026-07-10 | 2.23 | Reuse the shared real-backend verifier across the three Home/Parse scenarios. |
| 2026-07-10 | 2.22 | Remove the unread MiniRoundRail language prop and caller argument. |
| 2026-07-10 | 2.21 | Remove the unrendered Home upload-source subtitle from prototype and locale assets. |
| 2026-07-10 | 2.20 | Align the BDD fixture gate wording with the current 37-operation OpenAPI contract. |
| 2026-07-10 | 2.19 | Remove the redundant pending-import test reset API and teardown. |
| 2026-07-10 | 2.17 | Normalize workspace detail out-of-scope and hardcoded-round negative wording without behavior changes. |
| 2026-07-10 | 2.16 | Align workspace detail and fixed-round negative wording without behavior changes. |
| 2026-07-10 | 2.15 | Parse success detail ignores route-only `resumeId` for binding; Start requires saved TargetJob binding. |
| 2026-07-09 | 2.13 | Review remediation: constrain Home recent to ready TargetJobs and preserve structured round params in quick-start practice handoff. |
| 2026-07-09 | 2.12 | Reopen owner plan so Home recent cards reuse the Interview list action card, show quick start, and omit delete. |
| 2026-07-09 | 2.11 | Reopen owner plan to give Home recent cards a fixed max-width grid and share `MockInterviewCard` with the workspace plan list. |
| 2026-07-09 | 2.10 | Reopen owner plan to upgrade round assumptions from string-only `interviewHypotheses` focus text to structured LLM-derived `interviewRounds[]` covering round count, type/name, duration and focus. |
| 2026-07-09 | 2.9 | Reopen owner plan to bind Parse, Home recent card rails and shared TargetJob navigation context to backend-generated `summary.interviewHypotheses` instead of static round focus text. |
| 2026-07-09 | 2.7 | Reopen owner plan to unify Parse preview and workspace current-plan detail into one Interview Plan Detail / Context Confirm mother page. |
