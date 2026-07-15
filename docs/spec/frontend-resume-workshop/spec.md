# Frontend Resume Workshop Spec

> **版本**: 2.17
> **状态**: completed
> **更新日期**: 2026-07-15

## 1 背景与目标

`frontend-resume-workshop` 是当前 `resume_versions` 路由的前端 owner。正式前端必须实现 [`docs/ui-design/resume-module.md`](../../ui-design/resume-module.md)、[`resume-onboarding.md`](../../ui-design/resume-onboarding.md) 与 [`ui-architecture.md`](../../ui-design/ui-architecture.md) 中的当前 flat Resume Workshop 设计。

当前目标：

1. **路由接管**：`resume_versions` route 渲染 `ResumeWorkshopScreen`，支持 list / create / detail 三类视图。
2. **Flat Resume UI**：Resume 是平铺资产；列表使用响应式卡片网格，桌面固定最大列宽并左对齐、移动端单列；详情页是只读简历正文，上传 PDF 使用同源 source endpoint 渲染为从上到下平铺的 PDF 页面栈，粘贴、Markdown 文件和 TXT 文件使用 Markdown 渲染引擎；Markdown body 区域只渲染简历正文，不额外注入 `displayName`、header 名称、summary 或来源元数据；PDF 与 Markdown 使用统一阅读背景板和 page surface 节奏；不提供 preview / rewrites / edit tab、导出、复制、原件弹层、结构化草稿确认、PDF viewer 工具栏或二次编辑入口；所有前端数据投影都以 `resumeId` 识别简历。
3. **CreateFlow**：`flow=create` 只提供 upload / paste 输入；upload 仅支持 PDF / Markdown / TXT；注册成功后进入 `resume_versions?resumeId=<id>` 的解析等待态，直到 backend parse 成功后展示来源格式自适应详情，或失败后展示可恢复失败态；CreateFlow 本身不渲染预览确认页或确认保存页。
4. **真实 client 与 fixture fallback**：frontend 使用 generated client；列表只消费 closed `ResumeSummary`，详情才消费完整 `Resume`；real backend mode 与 fixture-backed dev path 都必须有测试护栏。
5. **UI parity 可执行**：用户可见变更必须有 DOM anchor、computed style、bounding box、viewport screenshot smoke 或对应 owner gate。
6. **幂等初始读取**：React StrictMode 下，相同已认证参数的并发 `listResumes` / `getResume` 初始读取只允许一次实际 transport；失败请求必须从 in-flight registry 驱逐，用户重试会发起新的 transport。

本 subject 不实现 backend handler、OpenAPI schema、migration、object storage、AI parsing 或真实 PDF 生成。

## 2 范围

### 2.1 In Scope

- **Route shell**：`ResumeWorkshopScreen` 解析 `flow=create|list`、`resumeId` 和 `createMode=upload|paste`；out-of-scope `tab` / `tailorRunId` 参数被过滤或忽略，并与 app shell route / TopBar 状态一致。
- **List view**：`ResumeListView` 只消费 `ResumeSummary` closed DTO，渲染响应式卡片网格、统计、唯一创建入口、详情入口和删除入口；卡片展示名称、摘要、来源、语言和最近编辑，底部保留“打开”，右上角保留删除；桌面固定最大列宽并左对齐，移动端单列。不得通过列表响应携带或读取详情正文、结构化档案、文件对象或审计时间字段；列表底部不再重复“上传或粘贴另一份简历”CTA。
- **Detail view**：`ResumeDetailView` 在 `queued/processing` 且正文快照为空时渲染解析等待态并轮询；`ready` 后根据来源格式展示 PDF 页面栈或 Markdown 正文；`failed` 且无可读正文时渲染失败态；`parsedTextSnapshot` / `originalText` 是 Markdown 渲染主要正文来源，结构化字段只能作为无原文时的降级兜底。
- **Preview body**：`ResumePreviewTab` 作为只读正文投影，PDF 上传自动使用 source endpoint 的 PDF 页面栈 renderer，所有页面从上到下平铺展示，不使用浏览器内置 PDF viewer toolbar / sidebar / pagination controls；粘贴 / Markdown / TXT 自动使用 Markdown engine，body card 只包含 `parsedTextSnapshot` / `originalText` / fallback body 本身，不额外 prepend `displayName`、详情 header 名称、summary 或来源元数据；PDF 与 Markdown 共用阅读背景板，Markdown 正文也位于背景板内的白色 page surface；不渲染复制、导出、原件弹层、改写建议、结构化草稿确认或编辑控件。
- **Create flow**：`ResumeCreateFlow` upload / paste 两路径；upload 只允许 `.pdf,.md,.markdown,.txt`；`createUploadPresign`、browser PUT、`registerResume` generated-client contract；upload 10MiB 与 paste 384KiB 默认边界从 `AppRuntimeProvider.contentLimits` 读取并按 UTF-8 bytes 本地校验，backend 仍作最终裁决；注册成功后导航到详情等待/终态页，不在创建流内 `getResume` 轮询或 `updateResume` 保存；右侧说明 rail 不再渲染。
- **Home entry handoff**：Home `还没有简历？1 分钟创建` 进入当前 CreateFlow；Home `选择已有简历` 消费 `listResumes`，对非归档且已有可读简历证据的记录保持可选，不因 `parseStatus` 仍为 `queued` / `processing` / `failed` 但已有正文快照而显示空态。
- **i18n / a11y**：中英双语、只读正文语义、错误/空态和 keyboard behavior。
- **Auth boundary**：未登录只能显示登录引导 / pending action；pending action 只保存安全 route params。
- **Privacy boundary**：raw resume text、parsed summary、structured profile、rewrite text 不进入 console、URL、localStorage、telemetry 或 generic transport logs。
- **Parity gates**：每个 user-visible owner plan 保留 DOM / style / layout / screenshot smoke 或 scenario gate。

### 2.2 Out of Scope

- Backend Resume / Upload / Tailor handlers and stores.
- OpenAPI operation/schema design.
- Migration / schema changes.
- Object-storage provider implementation.
- Real PDF generation.
- Product areas not included in the active product-scope core loop.

## 3 用户决策

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | UI 设计文档 | `frontend/src` + primitives + app shell + `docs/ui-design/` | 不从外部设计系统或 AI 审美生成正式前端视觉 |
| D-2 | Data adapter | 列表、Home selector 等集合消费者只消费 closed `ResumeSummary`；详情 route 才消费完整 `Resume`；两者都以 `resumeId` 关联，adapter 只做 display projection 和 fallback | 组件不直接拼 API response shape，也不以完整详情对象充当列表项 |
| D-3 | Route params | `flow=create|list`、`resumeId`、`createMode=upload|paste`；out-of-scope `tab` / `tailorRunId` 不属于当前 route state | Route state 只表达当前 list/create/detail 三态 |
| D-4 | Client mode | generated client 是唯一 API client；fixture-backed dev path 与 real backend mode 都保留测试 | 避免 mock-only drift |
| D-5 | UI parity | DOM anchor、computed style、bounding box、viewport screenshot smoke 为 user-visible gate | 不接受“风格接近”作为完成依据 |
| D-6 | Detail read-only | 简历详情页不提供 export / copy / view-original / rewrites / edit / preview-confirm 操作；原始简历预览就是当前只读 Markdown 正文 | 用奥卡姆剃刀收敛详情页，只保留用户要看的简历内容 |
| D-7 | Negative gate | product-scope pruning gate owns out-of-scope route/module/input regression scan | 防止范围外入口回流 |
| D-8 | Create wait handoff | CreateFlow 只提供 upload / paste；`registerResume` 成功后跳转到详情 route 的解析等待态，等待页轮询到 `ready` 后展示 Markdown 详情，`failed` 后展示失败态；preview confirm 不属于当前流程 | 避免用户在解析过程中长期看到“名称生成中”或空正文，提高提交后的可理解性 |
| D-9 | Display name robustness | 创建后完成态 `displayName` 优先由 backend parse 从 LLM 结构化结果中派生；若 LLM 输出失败但 backend 已抽取出可读正文，backend 必须写入非通用、非文件名、非 raw 第一行直出的可读 fallback 名称。frontend 不展示通用“上传/粘贴的简历”，也不得把 raw resume 第一行、上传文件名或与来源 `title` 相同的文件名 `displayName` 当作列表或详情名称；仅在解析尚未产生名称和正文前显示中性“名称生成中”pending label 或来源信息 | 列表和详情使用可识别名称，避免失败态长期停留在“名称生成中”或误用 Markdown 标题、正文首行、PDF 文件名 |
| D-10 | List actions | 列表只有 Header “新建简历”作为创建入口；每张卡片底部提供“打开”，右上角提供删除（调用 `archiveResume` 软删除并从列表隐藏），删除失败给出可恢复错误；数量上限由 backend `resume.maxActive` 强制，前端只展示服务端错误提示 | 避免重复 CTA，保留清晰的查看、清理资产和解除数量上限路径 |
| D-11 | Markdown body | `parsedTextSnapshot` 成功态是 backend LLM 生成的 Markdown 快照，详情页必须按 Markdown 结构渲染标题、段落和列表；body card 不得额外注入 `displayName`、header 名称、summary 或来源元数据；不得把 Markdown 当普通 txt 段落显示 | 统一后续 UI 渲染输入，同时保留简历行文结构，避免详情 header 信息污染简历正文 |
| D-12 | Source-format renderer | 详情正文区域根据来源格式自动选择 renderer：upload PDF 使用 `/api/v1/resumes/{resumeId}/source` 通过 PDF 页面栈从上到下平铺所有页面；paste、Markdown 文件和 TXT 文件使用 Markdown engine；PDF 与 Markdown 使用统一阅读背景板和 page surface；DOCX 不属于当前 Resume 上传支持范围 | 兼顾用户查看原始 PDF 版式与 LLM 后续交互所需的可读文本，不增加新按钮、二级入口或浏览器 PDF viewer 工具栏 |
| D-13 | List/detail responsibility | `ResumeSummary` 字段集固定为 `id,title,displayName,language,sourceType,parseStatus,summaryHeadline,hasReadableContent,updatedAt`；`originalText`、`parsedTextSnapshot`、`structuredProfile`、`fileObjectId`、`parsedSummary` object、`createdAt`、`deletedAt` 只属于详情或服务端内部，不得进入列表 item | 缩小列表 payload 与隐私面，避免一次列表读取传输所有简历正文和结构化详情 |
| D-14 | StrictMode request identity | 相同 method + normalized URL/query + auth scope 且不带 `AbortSignal` 的并发初始 GET 共享一个 in-flight Promise；settled 后立即驱逐，reject 也必须驱逐；带 `AbortSignal` 的 loader/polling 不进入通用共享，业务轮询只在上一次请求 settle 后继续 | 保留 StrictMode 与合法重试/轮询语义，同时消除同一用户动作导致的重复实际 transport；不引入 TTL cache 或跨时间结果缓存 |
| D-15 | CreateFlow content limits | 只消费 RuntimeConfig `resumeUploadBytes` / `resumePasteTextBytes`，缺 endpoint 字段时由 generated/runtime provider 使用 A4 同值 code default 10MiB/384KiB；按 `TextEncoder` UTF-8 bytes 判断；limit 接受、limit+1 不发 presign/register | 删除 2MiB 本地真理源并与 backend typed config 对齐；UI DOM/样式不变，只更新验证数据与错误文案 |
| D-16 | List card layout | `ResumeListView` 使用语义化 list/card DOM，不再渲染 table/header/row；desktop 使用 `auto-fill` + 固定最大卡片列宽并 `justify-content:start`，mobile 使用同序单列；卡片不得因 1 张数据拉伸为整行宽块 | 复用面试规划卡片的响应式原则，在 PC 与移动端保持稳定规格和可扫描层级 |

## 4 设计约束

### 4.1 UI 设计文档约束

- 视觉、DOM、spacing、typography、color、shadow、radius、density、state 和 responsive behavior 必须追溯到 `frontend/` 或 `docs/ui-design/`。
- 正式 frontend 不 import `frontend/src` 作为 runtime component/data source。

### 4.2 数据约束

- Runtime data 只来自 generated client、runtime provider、fixture/mock client 或 user action。
- Adapter 位于 `frontend/src/app/screens/resume-workshop/adapters/` 或 create-flow 局部 adapter。
- `ResumeListView`、Home resume selector 和其它集合投影只能接收 `ResumeSummary`；只有 `ResumeDetailView` 及其详情 renderer 可以接收完整 `Resume`。
- `ResumeSummary` 必须保持 closed field set；类型测试、fixture parity 和 source negative gate 禁止列表 consumer 访问 `originalText`、`parsedTextSnapshot`、`structuredProfile`、`fileObjectId`、`parsedSummary` object、`createdAt` 或 `deletedAt`。
- Route and pending action must never carry raw resume content.

### 4.3 Privacy 约束

- Raw resume text、parsed summary、structured profile and rewrite text are user content.
- Errors and toast messages use enum/generic wording and must not echo raw payloads.
- 详情页没有 copy/export 用户动作；passive logs and route state are not allowed to carry resume content.

### 4.4 Verification 约束

- Component and hook behavior use Vitest.
- 请求去重测试必须区分 hook/client method 调用次数与底层实际 transport 次数；StrictMode 双 effect 允许共享同一 in-flight Promise，但同一 request identity 的底层 transport 必须为 1。
- reject / abort / settle 后 registry 必须清理；失败后的显式重试必须产生新的 transport，且 queued/processing 详情轮询不得被永久缓存或吞掉。
- Route, auth, privacy and integration flows use focused scenario tests.
- Visual quality is verified by formal component, responsive and accessibility assertions; real-browser acceptance must use the running frontend/backend when applicable.
- Formal Resume Workshop CSS must not retain breadcrumb, structured-preview, modal or action selectors without a current DOM or prototype consumer.
- Header / INDEX drift uses `/sync-doc-index`.

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `resume_versions` route | frontend-resume-workshop | Resume Workshop shell, list, create, detail |
| UI design document | `frontend/` + `docs/ui-design/` | Visual and interaction source |
| Generated client | openapi-v1-contract + frontend adapters | API surface and TS types |
| Upload backend | backend-upload | Presign and object file lifecycle |
| Resume backend | backend-resume | Register, parse, update, duplicate, tailor |
| App shell / auth pending action | frontend-shell | Route normalization and auth continuation |
| Workspace Resume Picker | frontend-workspace-and-practice | Workspace-level resume selection |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Route shell | Authenticated user opens `resume_versions` | Route renders | Resume Workshop shell appears and TopBar highlights resume nav | [001](./plans/001-listing-routing-and-detail-readonly/plan.md) |
| C-2 | List view | `listResumes` returns `ResumeSummary[]` | List loads | Responsive card grid, header create entrypoint, per-card open and top-right delete actions render; desktop card width stays stable and left-aligned, mobile is single-column, table/header/row semantics are absent; each card exposes only the locked summary fields, forbidden detail fields are absent, and duplicate bottom upload/paste CTA is absent | [001](./plans/001-listing-routing-and-detail-readonly/plan.md) |
| C-3 | Detail read-only | User opens a resume | Detail renders | Full `Resume` is fetched only through `getResume`; pending parse with no readable body shows a waiting state and polls sequentially; upload PDF renders the source endpoint as a top-to-bottom page stack without native PDF viewer toolbar; paste / Markdown / TXT renders Markdown headings / lists / paragraphs without injected displayName/header metadata; PDF and Markdown share the same reading backdrop and page-surface rhythm; failed with no readable body shows a failure state; export / copy / original modal / rewrite / edit surfaces are absent; out-of-scope tab params are ignored | [001](./plans/001-listing-routing-and-detail-readonly/plan.md) |
| C-4 | Create upload/paste | User selects valid file or enters text；owner config provides byte limits | Submit | 注入小型 boundary 验证 overflow inline rejection with zero presign/register；valid input navigates to waiting/detail；默认/override/invalid 由 typed config owner 覆盖，不构造默认大小文件或配置 E2E | [002 Phase 13](./plans/002-create-flow/plan.md) |
| C-5 | Create paste | User enters text | Submit | Register completes and app navigates to the waiting/detail route; request title remains a neutral source title, and visible list/detail name comes from backend generated `displayName` after parse or extracted-text fallback, never from the raw first line or source filename/title fallback | [002](./plans/002-create-flow/plan.md) |
| C-6 | Create recovery | Register or upload fails | User retries from input | Input is preserved locally and no raw content leaks | [002](./plans/002-create-flow/plan.md) |
| C-7 | Home handoff | Home create CTA or Home existing-resume selector | Click / select | Create CTA lands on CreateFlow and auth pending action is safe; Home existing-resume selector shows non-archived readable `listResumes` records and carries the selected `resumeId` into JD import / parse handoff | [002](./plans/002-create-flow/plan.md) |
| C-8 | Delete resume | User deletes a row from Resume list | Archive succeeds or fails | Success hides the row and can free backend count limit; failure shows retryable feedback without removing data | [001](./plans/001-listing-routing-and-detail-readonly/plan.md) |
| C-10 | Privacy | User browses or creates resumes | App logs/routes/stores update | Raw resume content stays out of passive channels | 001 / 002 |
| C-11 | UI parity | Desktop and mobile viewports | Run owner gates | DOM/style/layout/screenshot smoke remain aligned with UI design document | 001 / 002 |
| C-12 | StrictMode list read | Authenticated list mounts under React StrictMode | `listResumes` effects overlap | Identical concurrent reads produce exactly one underlying transport; a rejected read is evicted, and retry produces a new transport and can succeed | [001](./plans/001-listing-routing-and-detail-readonly/plan.md) |
| C-13 | StrictMode detail read | Authenticated ready detail mounts under React StrictMode | `getResume(resumeId)` effects overlap | Initial identical read produces exactly one underlying transport; rejected reads remain retryable; queued/processing polling may issue a later request only after the previous request settles | [001](./plans/001-listing-routing-and-detail-readonly/plan.md) |

## 7 关联计划

- [001-listing-routing-and-detail-readonly](./plans/001-listing-routing-and-detail-readonly/plan.md)：route shell、list view、delete action、waiting/detail Markdown read-only body, display-name fallback and out-of-scope detail-action negative owner.
- [002-create-flow](./plans/002-create-flow/plan.md)：current upload/paste CreateFlow input, RuntimeConfig 10MiB/384KiB validation, waiting/detail handoff, out-of-scope preview-confirm negative owner, CTA handoff, privacy and focused frontend tests.
