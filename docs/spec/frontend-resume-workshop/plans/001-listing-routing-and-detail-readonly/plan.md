# Frontend Resume Workshop Listing Routing and Detail Readonly

> **版本**: 3.6
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本计划承接当前 `frontend-resume-workshop` 的首屏与只读详情边界：

- `resume_versions` route 渲染 `ResumeWorkshopScreen`，TopBar 选中简历入口。
- route params 只使用当前 flat Resume 合同：`flow=create|list`、`resumeId`、`createMode=upload|paste`；out-of-scope `tab` / `tailorRunId` 被过滤或忽略。
- `ResumeListView` 使用 `listResumes` 渲染单层平铺表格、唯一 Header 创建入口、详情入口、删除入口、loading / empty / retry / pagination 状态；底部“上传或粘贴另一份简历”重复 CTA 必须删除。
- `ResumeDetailView` 使用 `getResume(resumeId)` 渲染解析等待态、来源格式自适应只读详情、解析失败态和 generic 404 fallback。
- `ResumePreviewTab` 必须按来源格式自动适配：upload PDF 使用 `/api/v1/resumes/{resumeId}/source` 作为 PDF 文件源，并在详情正文中从上到下平铺所有 PDF 页面；不得使用 `<object>` / `<iframe>` / `<embed>` 触发浏览器原生 PDF viewer toolbar、sidebar、download、print 或分页导航；paste、Markdown 文件和 TXT 文件使用正式 Markdown 渲染引擎渲染 `parsedTextSnapshot`，支持 heading / paragraph / list / inline strong / link / GFM 基础语法；不得用手写 block parser 把 `**bold**`、`[link](url)` 等 inline Markdown 当普通文本显示。
- Markdown 正文区域只渲染 resume body 本身，不得额外注入 `displayName`、详情 header 名称、summary 或来源元数据；PDF 与 Markdown 必须使用同一阅读背景板节奏，避免同一详情页因来源格式产生割裂观感。
- 列表与详情不展示通用“上传的简历 / 粘贴的简历 / Uploaded resume / Pasted resume”名称；完成态名称优先使用 backend generated `displayName` 或 LLM structured headline；前端不得把 raw resume 第一行、上传文件名或与来源 `title` 相同的文件名 `displayName` 当作名称。
- 未登录态不触发 Resume API，请求登录时 pending action 只保存安全 route params。
- 可见 UI 继续追溯 `ui-design/src/screen-resume-workshop.jsx`、`ui-design/src/primitives.jsx`、`ui-design/src/app.jsx` 和 `docs/ui-design/`。

本计划不拥有 CreateFlow 输入提交链路、tailor polling、duplicate/save-as-new 或 backend parse 业务规则；本计划消费 `archiveResume`、`getResume`、`listResumes` generated-client 合同，并固化详情页不提供 Rewrites/Edit/export/copy/original modal/preview-confirm 等二次操作。

## 2 背景

当前产品已经收敛为 flat Resume Workshop。001 作为首个前端 owner，只保留当前仍被运行时、场景和 UI 真理源共同承接的 list / preview detail 合同。旧树形列表、版本集合、分叉参数、逐版本导出和 fallback 页面接管说明不再作为计划语义存在。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior + frontend + contract-consumer`
- **TDD 策略**: 适用。实现项由 `/implement frontend-resume-workshop/001-listing-routing-and-detail-readonly` 进入 `/tdd`；测试断言来源为 `ResumeWorkshopScreen`、`ResumeWorkshopAuthGate`、`ResumeListView`、`ResumeDetailView`、`ResumeDetailFixtureParity`、`ResumeDetailExport`、`ResumePreviewTab`、`ResumeWorkshopI18nA11y`、`ResumeWorkshopPrivacy`、`fixture-parity` 和 P0.036/P0.037 scenario Vitest。
- **BDD 策略**: 适用。主 checklist 保留 E2E.P0.036 / E2E.P0.037 `BDD-Gate:`，场景细节由 [bdd-plan.md](./bdd-plan.md) 与 [bdd-checklist.md](./bdd-checklist.md) 承接。
- **替代验证 gate**: focused frontend Vitest、P0.036/P0.037 scenario scripts、frontend typecheck/build 或 owner parity gate、context validation、`sync-doc-index --check`、`make docs-check`、`git diff --check`、core-loop pruning surface lint。

### 3.1 Frontend / Backend Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` `default` / `empty` / `paginated` | list hook + `ResumeListView` + `mapResumeToUiSource` display-name fallback | backend-resume real handler | `resumes` | none | E2E.P0.036 |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` `default` / `not-found` | detail hook + `ResumeDetailView` + `ResumePreviewTab` waiting / Markdown / failed states | backend-resume real handler | `resumes` | `resume.parse` produces Markdown snapshot | E2E.P0.037 |
| `getResumeSource` | `openapi/fixtures/Resumes/getResumeSource.json` `default` / `not-found` | `ResumePreviewTab` PDF page-stack renderer source URL | backend-resume real handler | `file_objects` + object storage | none | E2E.P0.037 + focused component + pixel smoke |
| `archiveResume` | `openapi/fixtures/Resumes/archiveResume.json` `default` | list row delete action | backend-resume real handler | `resumes.deleted_at` soft-hide | none | E2E.P0.036 focused regression |

## 4 实施步骤

### Phase 1: Route Shell / Auth Boundary

#### 1.1 Route dispatch

`ResumeWorkshopScreen` 解析当前 route params，并在 `flow=create`、`resumeId`、list 三种状态间分派到 CreateFlow、Detail、List。

#### 1.2 Auth gate

runtime 未认证时渲染登录入口；Resume API 请求保持 0 次；pending action 只携带 route params，不携带 raw resume content、parsed summary、structured profile 或 rewrite text。

### Phase 2: Flat List View

#### 2.1 Flat table

`ResumeListView` 从 `listResumes` 读取 flat Resume items，按最近更新时间排序，渲染名称、来源、语言、最近编辑和打开按钮。

#### 2.2 List states

覆盖 loading、empty、retryable error 和 `pageInfo.hasMore` 提示；数量和行 ID 从 fixture / API response 派生。

#### 2.3 Create and detail entry

创建入口只保留 Header “新建简历”按钮并导航到 `resume_versions?flow=create`；打开行导航到 `resume_versions?resumeId=<id>`。

#### 2.4 Delete row action

每行提供删除按钮，调用 `archiveResume` 后从列表隐藏该简历；删除失败时保留原行并显示 retryable 错误。底部“上传或粘贴另一份简历”CTA 不再渲染。

### Phase 3: Read-only Detail

#### 3.1 Waiting and read-only detail

`ResumeDetailView` 使用 `getResume(resumeId)`；当 `parseStatus in queued|processing` 且无可读正文时渲染等待动画页并轮询；当 `ready` 或已有可读正文时渲染 crumb、header meta 和来源格式自适应只读正文：PDF upload 渲染 page stack，paste / Markdown / TXT 渲染 Markdown 正文；当 `failed` 且无可读正文时渲染解析失败页。显式 `tab=preview|rewrites|edit` 不 materialize 任何 tab 或二次编辑 surface。

#### 3.2 Removed actions

详情页不渲染 Export PDF、Copy text、View original/original modal、Rewrites、Edit 或 preview-confirm；原始简历预览就是当前来源格式自适应正文区域。

#### 3.4 Original-content projection and meaningful names

`ResumePreviewTab` 先根据来源格式选择 renderer：PDF upload 使用 source endpoint 的 PDF page-stack preview，从上到下平铺所有 PDF 页面；paste、Markdown upload 和 TXT upload 优先将 `parsedTextSnapshot` 作为 Markdown 渲染，其次将 `originalText` 作为纯文本兼容输入，最后才降级到结构化字段的只读摘要。PDF renderer 不得使用 `<object>` / `<iframe>` / `<embed>` 或暴露浏览器 PDF viewer toolbar；Markdown 渲染必须由 `react-markdown` + `remark-gfm` 等正式开源引擎承接，支持 inline strong / link 等真实 DOM 输出；上传文件刚注册后若 `parseStatus` 仍为 `queued/processing` 且正文快照为空，详情页显示等待动画并轻量轮询 `getResume`；若 `parseStatus='failed'` 或任一可读正文已到达，详情必须停止轮询并展示终态。列表和详情 header 对通用 `displayName` 做负向过滤，使用 backend generated 名称或 structured headline；raw resume 第一行、上传文件名或与来源 `title` 相同的文件名 `displayName` 不得作为名称兜底。

#### 3.3 404

不存在的 `resumeId` 渲染 generic NotFoundEmptyState，不回显 fixture `error.code`。

### Phase 4: Privacy / I18n / A11y / Parity

#### 4.1 Privacy

raw resume text、parsedTextSnapshot、parsedSummary、structuredProfile 和 rewrite text 不进入 URL、pending action、localStorage、console、telemetry 或 generic mock transport logs。

#### 4.2 I18n and accessibility

中英 key 由 frontend-shell i18n 体系承接；table、只读正文、buttons 和 aria labels 具备可测试语义。

#### 4.3 UI parity

DOM anchor、computed style、bounding box、mobile / desktop layout 和 screenshot smoke 追溯 UI 真理源；截图 diff 只在 baseline 稳定时作为补充 gate。

### Phase 5: BDD / Negative Gate / Closeout

#### 5.1 BDD scenarios

E2E.P0.036 验证 flat list + auth boundary；E2E.P0.037 验证 read-only detail + out-of-scope tab negative + removed actions + 404 fallback。

#### 5.2 Out-of-scope negative gate

Resume Workshop runtime source、scenario evidence 和 rendered DOM 不出现树形列表、版本 route params、版本集合 operation、分叉参数、prototype runtime import 或 out-of-scope route testid。

#### 5.3 Docs and index

计划、checklist、BDD、context、spec history、scenario INDEX 和 docs/spec INDEX 同步到当前 Header。

### Phase 8: Source-format Renderer

#### 8.1 PDF source renderer

`ResumePreviewTab` 对 upload-backed `.pdf` resume 渲染同源 PDF page-stack preview，URL 为 generated client baseUrl + `/resumes/{resumeId}/source`；该 renderer 不展示 `parsedTextSnapshot` / `originalText` fallback 文本，也不新增按钮、tab、原件弹层、浏览器 viewer toolbar、download/print 控件或分页导航。

（验证：`corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/adapters/resume.test.ts src/app/screens/resume-workshop/components/ResumePreviewTab.test.tsx` PASS）

#### 8.2 Markdown source renderer

paste、Markdown upload 和 TXT upload 继续使用 Markdown engine，并保留 inline strong / link / list / heading DOM 断言。

（验证：`ResumePreviewTab.test.tsx` / `adapters/resume.test.ts` focused tests PASS）

### Phase 9: PDF Page-stack Refinement

#### 9.1 PDF page-stack renderer

`ResumePreviewTab` 使用 PDF.js renderer 从 `/resumes/{resumeId}/source` 读取 PDF，并在详情正文中渲染稳定的纵向页面栈。测试必须断言 `resume-detail-pdf-preview-stack` / page anchors 出现、`<object>` / `<iframe>` / `<embed>` 不出现，并且 Markdown fallback 文本不会渲染到 PDF 详情正文。

#### 9.2 UI truth and screenshot smoke

`ui-design/src/screen-resume-workshop.jsx`、`docs/ui-design/resume-module.md`、`docs/ui-design/resume-onboarding.md` 和正式 CSS 统一描述 PDF 页面栈，不再描述原生 viewer。Pixel parity smoke 覆盖 desktop/mobile PDF 详情，并断言页面栈可见且没有 native viewer shell。

### Phase 10: Source-format Reading Surface Alignment

#### 10.1 Markdown body purity

`ResumePreviewTab` 的 Markdown renderer 只消费 `buildResumeBodyMarkdown(resume)` 输出，不得在 body card 内额外 prepend `uiResume.name` / `displayName` / summary / source metadata。详情页 header 仍负责展示简历名称和来源信息。

#### 10.2 Unified reading surface

PDF 与 Markdown renderer 共用同一外层阅读背景板；PDF 页面和 Markdown 页面都作为背景板内的白色 page surface 呈现。CSS parity、component tests 和 pixel smoke 必须覆盖共同背景板、Markdown page anchor 和 PDF page-stack anchor。

### Phase 11: P0.036 Test Lifecycle Isolation

P0.036 的同步 out-of-scope negative test 在断言后显式 unmount，避免 fixture-backed runtime 与 InterviewContext effects 在用例结束后回写；业务断言和生产行为不变。

### Phase 12: PDF.js On-demand Loading

`PdfPageStackPreview` 首次 render 只渲染现有 loading shell；PDF.js module 与 worker URL 在 component effect 内动态导入，再创建同一 `getDocument` task。取消、失败、page-stack 和 credential 行为保持不变，非 PDF 首屏不得同步打包 PDF.js runtime。

### Phase 13: P0.037 Async Test Lifecycle

P0.037 trigger 同时记录 stdout/stderr，verify 将未被 `act(...)` 接管的 React update warning 视为失败；场景测试及其 `ResumeDetailView` owner mirror 都通过 Testing Library `act` 等待 failed-with-snapshot PDF 单次请求观察窗口，保留 350ms 轮询负向断言和全部业务行为，不修改生产 PDF renderer。

### Phase 14: Orphan Resume Toast Bridge Removal

删除正式 Resume Workshop 中无消费者的 `components/toast.ts`，并删除 P0.036 仅用于证明旧占位 toast 不出现的 `window.eiToast` capture；保留一个 scoped source gate，要求正式 Resume Workshop 与 P0.036 不再出现该 prototype bridge。`ui-design/` 原型 toast 实现不属于本批修改范围。

### Phase 15: P0.037 Ready DOM Synchronization

P0.037 pending-PDF 场景及其 `ResumeDetailView` owner mirror 必须等待 ready `displayName` 标题实际提交到 DOM，再断言 page stack 和后续只读内容。第二次 `getResume` 调用次数只作为轮询证据，不能充当 React 可见状态已经提交的同步屏障；测试修复不修改生产轮询、PDF renderer 或业务断言。诊断记录见 [BUG-0153](../../../../bugs/BUG-0153.md)。

### Phase 16: Zero-consumer Detail CSS Pruning

删除 `screens.css` 中没有正式 DOM、场景或 UI 原型消费者的详情 breadcrumb、旧 structured preview section/skills 和 original-content modal 选择器；同步从共享 back-button rule 中移除旧 preview action/modal button 分支。`ResumeWorkshopCssParity` 必须逐项锁定这些选择器保持不存在，当前 detail back、header、Markdown/PDF reading surface、pending/failed state 和响应式行为不变。

### Phase 17: Detail CSS cascade consolidation

将 `ei-resume-detail-back` 的两段同 specificity 规则合并为一个最终计算值等价的规则：保留后段实际生效的 layout/color/border/font-size，并迁入前段唯一未被覆盖的 border-radius/font-family；删除被覆盖的声明块。删除 `display:flex` 的 `ei-resume-detail-preview` 在 mobile media 中无效的 `grid-template-columns`。BDD 不适用，因为 DOM 与 computed style 不变；替代 gate 为 source RED/GREEN、focused/full Resume Workshop、typecheck/build、owner contexts 与 docs/diff/pruning gates。

### Phase 18: Empty pending-decision section removal

active spec 只保留当前锁定决策。删除没有任何决策内容、仅声明“当前没有待确认项”的 §3.2，并把第 3 节标题收敛为“用户决策”；不保留空状态、历史说明或替代段落。BDD 不适用，因为本批只校对 owner 文档结构；替代 gate 为 scoped RED/GREEN、两个 owner context、docs/index/link/diff/pruning gates。

## 5 验收标准

- 001 owner docs 只描述当前 flat Resume list / original-content read-only detail 合同。
- 列表只保留一个创建入口，且支持删除 row 的成功/失败状态。
- 详情 pending/ready/failed 三态可见；upload PDF 使用 PDF source page-stack preview 且无 native viewer toolbar，paste / Markdown / TXT 使用 Markdown DOM 渲染；inline `strong`、link、GFM list/table 语法不得以源码符号形式露出。
- Markdown body card 不额外注入详情 header `displayName` / 名称 / summary；PDF 与 Markdown 详情正文区域使用统一阅读背景板和 page surface 视觉节奏。
- Operation matrix 只列当前详情实际消费的 generated-client operations。
- E2E.P0.036 / E2E.P0.037 scenario assets 指向当前 slug、当前 Vitest entry 和当前 expected outcome。
- Focused frontend tests、context validation、docs/index gates 和 pruning surface lint 通过。

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| Flat list 文档再次回流树形语义 | P0.036、fixture parity 和 pruning surface lint 保留负向断言 |
| Fixture-backed UI 被误认为 real backend 闭环 | scenario trigger 保留 real-mode/generated-client gate，operation matrix 标明真实 handler / fixture 边界 |
| 旧详情动作回流 | P0.037、pixel parity 和 negative grep 固化 Export/Copy/Original/Rewrites/Edit absence |

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-10 | 3.6 | Remove the empty pending-decision section from the active Resume Workshop spec. |
| 2026-07-10 | 3.5 | Consolidate the duplicate detail-back cascade and remove an ineffective flex-grid declaration. |
| 2026-07-10 | 3.4 | Delete zero-consumer breadcrumb, structured-preview and original-modal CSS from the read-only detail owner. |
| 2026-07-10 | 3.3 | Synchronize both pending-PDF tests on the ready visible heading instead of request count alone. |
| 2026-07-10 | 3.2 | Delete the orphan Resume Workshop toast bridge and its P0.036 self-only capture. |
| 2026-07-10 | 3.1 | Make P0.037 fail on unwrapped React updates and isolate its failed-PDF observation wait. |
| 2026-07-10 | 3.0 | Load PDF.js on demand behind the existing page-stack loading shell and document the completed P0.036 test lifecycle isolation. |
| 2026-07-10 | 2.9 | Isolate the synchronous P0.036 out-of-scope test lifecycle with explicit cleanup; keep Resume Workshop behavior unchanged. |
| 2026-07-10 | 2.8 | 将 detail route、fallback 和场景负向 gate 表述统一为 out-of-scope 口径；行为不变。 |
| 2026-07-10 | 2.7 | 将 fallback page 负向历史表述统一为 out-of-scope wording；行为不变。 |
| 2026-07-10 | 2.6 | 将 detail route 负向输入统一为 out-of-scope tab/query 口径；行为不变。 |
| 2026-07-10 | 2.5 | 将 out-of-scope fallback 页面和通用 displayName 文档口径收敛为当前 list/detail 合同用语。 |
| 2026-07-08 | 2.4 | 修订来源格式阅读面：Markdown body 禁止注入 displayName/header 元数据，PDF 与 Markdown 共用阅读背景板和 page surface。 |
| 2026-07-08 | 2.3 | 将 upload PDF 详情从浏览器原生 PDF object 改为无工具栏的纵向 page-stack renderer。 |
| 2026-07-07 | 2.2 | 新增来源格式渲染：upload PDF 使用同源 source endpoint PDF preview，paste / Markdown / TXT 使用 Markdown engine。 |
| 2026-07-07 | 2.1 | 修复 Markdown 渲染引擎缺口：`ResumePreviewTab` 改用 `react-markdown` + `remark-gfm`，覆盖 inline strong / link 等真实 DOM 渲染。 |
| 2026-07-07 | 2.0 | 本轮优化：列表删除重复上传 CTA，新增删除 row；详情新增解析等待 / 失败态并用 Markdown 渲染正文。 |
| 2026-07-07 | 1.9 | 修订上传详情性能回归：`failed` 或已有可读正文时停止 `getResume` 轮询；名称消费改为 backend generated displayName 优先。 |
| 2026-07-07 | 1.8 | 修订未闭环回归：禁止上传文件名 / 与来源 title 相同的文件名 displayName 作为可见名称；failed resume 只要已有 parsedTextSnapshot 仍展示原文。 |
| 2026-07-07 | 1.6 | 修订未闭环回归：详情正文改为优先展示原始内容快照，列表/详情过滤通用上传/粘贴名称并增加内容派生兜底。 |
| 2026-07-07 | 1.7 | 修订未闭环回归：禁止 raw 第一行作为可见名称；上传详情在原文快照到达前轻量轮询，避免 PDF 详情空白。 |
| 2026-07-07 | 1.5 | 将详情页收敛为只读简历正文，移除 export/copy/original modal/Rewrites/Edit 正向 gate，并过滤 out-of-scope `tab` / `tailorRunId` route 口径。 |
| 2026-07-07 | 1.4 | 压缩 001 owner 到当前 flat Resume list/detail preview 合同，移除旧树形/版本集合/分叉参数语义，并同步 P0.036 当前场景 slug。 |
