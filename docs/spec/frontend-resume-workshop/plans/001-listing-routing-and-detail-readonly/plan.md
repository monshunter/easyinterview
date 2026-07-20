# Frontend Resume Workshop Listing Routing and Detail Readonly

> **版本**: 4.7
> **状态**: completed
> **更新日期**: 2026-07-20

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本计划承接当前 `frontend-resume-workshop` 的首屏与只读详情边界：

- `resume_versions` route 渲染 `ResumeWorkshopScreen`，TopBar 选中简历入口。
- route params 只使用当前 flat Resume 合同：`flow=create|list`、`resumeId`、`createMode=upload|paste`；out-of-scope `tab` / `tailorRunId` 被过滤或忽略。
- `ResumeListView` 使用 `listResumes` 的 closed `ResumeSummary` 渲染 desktop 双列等宽卡列表、唯一 Header 创建入口、每卡片打开/删除入口、loading / empty / retry / pagination 状态；Header 创建入口与 Workspace 创建入口共享 22px 圆圈加号视觉语义；mobile 收敛为同序单列占满可用宽度；列表 consumer 不得读取完整 `Resume` 详情字段，且底部“上传或粘贴另一份简历”重复 CTA 必须删除。
- `ResumeDetailView` 使用 `getResume(resumeId)` 渲染解析等待态、来源格式自适应只读详情、解析失败态和 generic 404 fallback。
- `ResumePreviewTab` 必须按来源格式自动适配：upload PDF 使用 `/api/v1/resumes/{resumeId}/source` 作为 PDF 文件源，并在详情正文中从上到下平铺所有 PDF 页面；不得使用 `<object>` / `<iframe>` / `<embed>` 触发浏览器原生 PDF viewer toolbar、sidebar、download、print 或分页导航；paste、Markdown 文件和 TXT 文件使用正式 Markdown 渲染引擎渲染 `parsedTextSnapshot`，支持 heading / paragraph / list / inline strong / link / GFM 基础语法；不得用手写 block parser 把 `**bold**`、`[link](url)` 等 inline Markdown 当普通文本显示。
- Markdown 正文区域只渲染 resume body 本身，不得额外注入 `displayName`、详情 header 名称、summary 或来源元数据；PDF 页面栈与 Markdown page surface 直接位于详情画布并共用 `794px` A4 纸宽；PDF 每页保持 `210:297`，Markdown 整体是一张由正文自然撑高的连续长页面，不分页、不设置 A4 比例或固定/最小纸高；外层 `article` 不得恢复共享背景板样式。
- 列表与详情不展示通用“上传的简历 / 粘贴的简历 / Uploaded resume / Pasted resume”名称；完成态名称优先使用 backend generated `displayName` 或 LLM structured headline；前端不得把 raw resume 第一行、上传文件名或与来源 `title` 相同的文件名 `displayName` 当作名称。
- 未登录态不触发 Resume API，请求登录时 pending action 只保存安全 route params。
- React StrictMode 下，相同 request identity 的并发 `listResumes` / ready `getResume` 初始读取必须共享一次实际 transport；失败后 registry 清理且用户重试会发出新 transport；queued/processing 详情轮询保持 settle 后串行推进。
- 可见 UI 继续追溯 `frontend/src` 和 `docs/ui-design/`。

本计划不拥有 CreateFlow 输入提交链路、tailor polling、duplicate/save-as-new 或 backend parse 业务规则；本计划消费 `archiveResume`、`getResume`、`listResumes` generated-client 合同，并固化详情页不提供 Rewrites/Edit/export/copy/original modal/preview-confirm 等二次操作。

## 2 背景

当前产品已经收敛为 flat Resume Workshop。001 作为首个前端 owner，只保留当前仍被运行时、场景和 UI 设计文档共同承接的 list / preview detail 合同。旧树形列表、版本集合、分叉参数、逐版本导出和 fallback 页面接管说明不再作为计划语义存在。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior + frontend + contract-consumer`
- **TDD 策略**: Phase 28 先扩展 `ResumeWorkshopCssParity.test.ts`，要求白色 Markdown page 明确拥有深色正文墨水，而不是继承 `--ei-color-fg-primary`。RED 在当前夜间模式继承缺口上失败；GREEN 只增加 page-local ink，现有 renderer/component tests 继续证明 Markdown DOM、A4 几何、PDF.js 与只读边界不变。
- **BDD 策略**: `BDD.RESUME.DETAIL.DARK.008` 使用 `ResumeWorkshopCssParity.test.ts` 与 `ResumePreviewTab.test.tsx` 作为 domain behavior evidence；真实 Chrome 验收 light/dark 的 page/body/list/strong 计算色和正文可读性，不创建或声明 E2E ID。

### 3.1 Frontend / Backend Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `listResumes` | current list fixtures | list hook + flat list + Home selector | backend-resume summary handler | summary projection | none | 当前无真实 E2E owner；root `make test` |
| `getResume` | current detail fixtures | detail hook + readonly detail | backend-resume full-detail handler | full resume projection | parse produces snapshot | 当前无真实 E2E owner；root `make test` |
| `getResumeSource` | current source fixtures | PDF page-stack renderer | backend-resume source handler | file object + object storage | none | 当前无真实 E2E owner；root `make test` + PDF page-stack component tests |
| `archiveResume` | current archive fixture | list-card confirmed delete | backend-resume archive handler | `resumes.deleted_at` | none | 当前无真实 E2E owner；root `make test` |

## 4 实施步骤

### Phase 1: Route Shell / Auth Boundary

#### 1.1 Route dispatch

`ResumeWorkshopScreen` 解析当前 route params，并在 `flow=create`、`resumeId`、list 三种状态间分派到 CreateFlow、Detail、List。

#### 1.2 Auth gate

runtime 未认证时渲染登录入口；Resume API 请求保持 0 次；pending action 只携带 route params，不携带 raw resume content、parsed summary、structured profile 或 rewrite text。

### Phase 2: Responsive Card List View

#### 2.1 Card grid

`ResumeListView` 从 `listResumes` 读取 flat Resume items，按最近更新时间排序，以 desktop 固定最大列宽/mobile 单列的卡片网格渲染名称、摘要、来源、语言、最近编辑、打开和删除动作；Phase 20 承接从旧 table DOM 到该布局的 TDD 实施。

#### 2.2 List states

覆盖 loading、empty、retryable error 和 `pageInfo.hasMore` 提示；数量和卡片 ID 从 fixture / API response 派生。

#### 2.3 Create and detail entry

创建入口只保留 Header “新建简历”按钮并导航到 `resume_versions?flow=create`；打开卡片导航到 `resume_versions?resumeId=<id>`。

#### 2.4 Delete row action

每张卡片提供删除按钮；首次点击只打开二次确认，确认后才调用 `archiveResume` 并从列表隐藏该简历。取消路径零请求；删除失败时保留卡片与对话框并显示 retryable 错误。底部“上传或粘贴另一份简历”CTA 不再渲染。

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

DOM anchor、computed style、bounding box、mobile / desktop layout 和 screenshot smoke 追溯 UI 设计文档；截图 diff 只在 baseline 稳定时作为补充 gate。

### Phase 5: BDD / Negative Gate / Closeout



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

（验证：相关 focused tests 仅作开发反馈；阶段单测完成由仓库根 `make test` 承接）

### Phase 9: PDF Page-stack Refinement

#### 9.1 PDF page-stack renderer

`ResumePreviewTab` 使用 PDF.js renderer 从 `/resumes/{resumeId}/source` 读取 PDF，并在详情正文中渲染稳定的纵向页面栈。测试必须断言 `resume-detail-pdf-preview-stack` / page anchors 出现、`<object>` / `<iframe>` / `<embed>` 不出现，并且 Markdown fallback 文本不会渲染到 PDF 详情正文。

#### 9.2 UI truth and screenshot smoke

`frontend/src`、`docs/ui-design/resume-module.md`、`docs/ui-design/resume-onboarding.md` 和正式 CSS 统一描述 PDF 页面栈，不再描述原生 viewer。正式 component/responsive tests 覆盖 desktop/mobile PDF 详情，并断言页面栈可见且没有 native viewer shell。

### Phase 10: Source-format Reading Surface Alignment

#### 10.1 Markdown body purity

`ResumePreviewTab` 的 Markdown renderer 只消费 `buildResumeBodyMarkdown(resume)` 输出，不得在 body card 内额外 prepend `uiResume.name` / `displayName` / summary / source metadata。详情页 header 仍负责展示简历名称和来源信息。

#### 10.2 Unified reading surface

PDF 与 Markdown renderer 共用同一外层阅读背景板；PDF 页面和 Markdown 页面都作为背景板内的白色 page surface 呈现。CSS parity、component tests 和 pixel smoke 必须覆盖共同背景板、Markdown page anchor 和 PDF page-stack anchor。



### Phase 12: PDF.js On-demand Loading

`PdfPageStackPreview` 首次 render 只渲染现有 loading shell；PDF.js module 与 worker URL 在 component effect 内动态导入，再创建同一 `getDocument` task。取消、失败、page-stack 和 credential 行为保持不变，非 PDF 首屏不得同步打包 PDF.js runtime。







### Phase 16: Zero-consumer Detail CSS Pruning

删除 `screens.css` 中没有正式 DOM、场景或 UI 原型消费者的详情 breadcrumb、旧 structured preview section/skills 和 original-content modal 选择器；同步从共享 back-button rule 中移除旧 preview action/modal button 分支。`ResumeWorkshopCssParity` 必须逐项锁定这些选择器保持不存在，当前 detail back、header、Markdown/PDF reading surface、pending/failed state 和响应式行为不变。

### Phase 17: Detail CSS cascade consolidation

将 `ei-resume-detail-back` 的两段同 specificity 规则合并为一个最终计算值等价的规则：保留后段实际生效的 layout/color/border/font-size，并迁入前段唯一未被覆盖的 border-radius/font-family；删除被覆盖的声明块。删除 `display:flex` 的 `ei-resume-detail-preview` 在 mobile media 中无效的 `grid-template-columns`。BDD 不适用，因为 DOM 与 computed style 不变；替代 gate 为 source RED/GREEN、仓库根 `make test`、typecheck/build、owner contexts 与 docs/diff/pruning gates。

### Phase 18: Empty pending-decision section removal

active spec 只保留当前锁定决策。删除没有任何决策内容、仅声明“当前没有待确认项”的 §3.2，并把第 3 节标题收敛为“用户决策”；不保留空状态、历史说明或替代段落。BDD 不适用，因为本批只校对 owner 文档结构；替代 gate 为 scoped RED/GREEN、两个 owner context、docs/index/link/diff/pruning gates。

### Phase 19: Resume summary consumption and idempotent initial reads

> 依赖 B2 OpenAPI owner 与 backend-resume 001 Phase 15 保持 `PaginatedResume` 外层与 `pageInfo` 不变，只把 `items` 收敛为 `ResumeSummary[]`。不得新增平行 pagination wrapper，也不得在 frontend 以手写窄化类型掩盖仍返回完整详情的服务端合同。

#### 19.1 RED: lock list/detail type separation

为 generated client fixture parity、list hook、`ResumeListView`、Home resume selector 和 adapter 增加失败测试：列表 item 必须只含 `id,title,displayName,language,sourceType,parseStatus,summaryHeadline,hasReadableContent,updatedAt`，且不能访问 `originalText`、`parsedTextSnapshot`、`structuredProfile`、`fileObjectId`、`parsedSummary` object、`createdAt`、`deletedAt`。详情现有 full `Resume` 断言保持可用。

#### 19.2 GREEN: consume ResumeSummary on collection surfaces

更新 generated-client consumer、hooks 与 adapters，使列表和 Home selector 仅接收 `ResumeSummary`；用 `summaryHeadline` 和 `hasReadableContent` 取代客户端从正文/structured profile 推断列表显示或可选性。`ResumeDetailView` 继续只通过 `getResume(resumeId)` 获取完整详情。

#### 19.3 RED/GREEN: one actual transport under StrictMode

在 generated-client transport wrapper / runtime client 的最低共享层增加 request identity 回归，明确断言底层 `fetch`/transport 次数，而不是 hook method invocation 次数。相同 auth scope、method、normalized URL/query 且无 `AbortSignal` 的并发 GET 共享一个 in-flight Promise；resolve/reject 后都立即驱逐。测试先复现 StrictMode 双 effect 产生两次 transport，再修复到一次。

#### 19.4 Retry, abort and polling boundaries

新增失败后重试测试：第一次 `listResumes` / `getResume` reject 后，显式 retry 会产生第二个新 transport 并成功。带 `AbortSignal` 的 route loader 或可取消请求不进入通用共享，避免一个 consumer abort 取消其他 consumer；不增加 TTL cache。queued/processing 详情仅在上一次 `getResume` settle 后按既有节奏轮询，ready/failed/已有正文不继续轮询。

### Phase 20: Responsive resume card list

#### 20.1 RED: reject table/list-row layout

先更新 `ResumeListView.test.tsx`、`ResumeWorkshopCssParity.test.ts` 与正式 responsive owner assertions，使当前 `role=table` / table header / row DOM、整行铺满和移动端横向布局先失败；测试必须同时锁定 card list/card anchors、closed `ResumeSummary` 字段、空态/错误/分页及打开/删除原行为不回退。

#### 20.2 GREEN: card grid and actions

将 `ResumeListView` 改为语义化 list + card：每张卡片展示可识别名称、可选 `summaryHeadline`、来源、语言和最近编辑；摘要缺失时直接省略该行，不伪造内容。“打开”保留在底部 action row，trash 删除固定在右上角。desktop 复用面试规划列表的响应式原则，使用固定最大列宽、`auto-fill` 和左对齐，1/2/3 张卡片规格稳定且单卡不得拉伸整行；mobile 使用同一 DOM 与阅读顺序收敛为单列。删除成功隐藏卡片，失败保留卡片并展示可恢复错误；不得新增 API、详情字段或第二个创建入口。

#### 20.3 Accessibility, responsive and closeout gates

卡片列表暴露 list/item 语义，打开与删除具有独立可访问名称、键盘焦点与最小触控区域；不得把嵌套交互包装成一个不可区分的 card button。超长名称/摘要/来源、缺失摘要、1/2/3 卡片与窄屏必须完整换行且无横向溢出。Focused Vitest 只作开发反馈；阶段收口执行根 `make test`、frontend typecheck/build、desktop/mobile geometry/no-overflow gate、context validation、`sync-doc-index --check`、`make docs-check`、`git diff --check`，完成后才恢复 `completed`。

### Phase 21: Parse waiting motion stability

#### 21.1 Stable animation contract

`ResumeDetailView` 的 queued/processing 等待态保持现有 DOM、56px 图标容器和轮询语义。`useResume` 仅在首次读取或资源身份变化时进入通用 loading；已有 pending Resume 的后台 poll 保留上一份 `data`，不得在请求间隙执行 `setData(null)` 导致“正在加载简历…”与“正在解析简历”交替闪现，终态响应到达后再原子替换。动画不得对圆形容器或 SVG 使用循环 `scale` / `translate`；只通过透明度与不参与布局的柔和 `box-shadow` 表达进行中状态，并在 `prefers-reduced-motion: reduce` 下停止动画。组件 RED/GREEN 锁定后台 poll 不闪回 loading，CSS owner test 锁定无几何变换；真实 Chrome 连续覆盖多个轮询周期时不得出现 loading flash 或图标、标题、说明位移。

### Phase 22: Resume list reference alignment

#### 22.1 RED：锁定参考稿列表层级

扩展 `ResumeListView.test.tsx` 与 `ResumeWorkshopCssParity.test.ts`，先要求页面标题区、唯一创建 CTA、与 Workspace 一致的 22px circled-plus、desktop 每行两张等宽卡、64px 文件 icon、名称/摘要、两行 meta、60px 级删除控件、footer 规则线和右对齐“打开”；旧 14px 裸加号、单列 918px 规则、语言 tag 与过小动作必须先失败。

#### 22.2 GREEN：重构正式 Resume list presentation

只调整 `ResumeListView` DOM class、`ResumeWorkshopIcon` 的 plus 几何与 `.ei-resume-workshop-*` 页面作用域 CSS；create icon 使用 Workspace 同款 22px 圆圈加号。保持 closed `ResumeSummary`、排序、loading/empty/error/pagination、`archiveResume`、`resume_versions?resumeId` 和 create route 不变。desktop 使用双列等宽卡，mobile 保持同一阅读顺序并无横向溢出。

#### 22.3 BDD / A11Y / CHROME

`BDD.RESUME.LIST.VISUAL.003` 由 list/CSS domain behavior tests 承接；1916×821 与 390×844 真实 Chrome 验收标题、卡片 bbox、删除/打开 keyboard、theme、console 与 no-overflow，不把 UI 验收冒充真实 API E2E。

#### 22.4 POST-PASS

执行 focused Vitest、frontend typecheck/build、仓库根 `make test`、owner context、`sync-doc-index --check`、`make docs-check` 与 `git diff --check`；证据同步后恢复 completed lifecycle。



### Phase 23: Resume preview reference composition

#### 23.1 RED: lock the complete detail hierarchy

`ResumeDetailView.test.tsx` 与 `ResumeWorkshopCssParity.test.ts` 先拒绝 breadcrumb 拼接、缺失名称 kicker、`1320/860/720px` 窄构图和仅检查 overflow 的表面 gate，同时继续锁定只读 renderer、正文不注入 header metadata 与无 action/tab/native PDF viewer 的既有边界。

#### 23.2 GREEN: rebuild the reading surface

详情 route 使用约 `1512px` desktop 内容面；Back、蓝色 eyebrow、名称 kicker、主标题和来源/日期 meta 共享左边界；正文使用约 `1310px` 浅色阅读背景板与约 `1150px` 居中白色 PDF/Markdown 纸张。Mobile 同序收敛为满宽背景板和可读内边距，不改变 generated client、轮询、来源格式判断或正文事实。

#### 23.3 ACCEPTANCE/CLOSEOUT

完成 focused owner tests、frontend typecheck/build、根 `make test`、owner context、docs/index/diff gates，并以 Chrome skill 在真实 frontend/backend 上记录 desktop/mobile Header、背景板、纸张 bbox、no-overflow 与截图证据；不把 scoped UI evidence 声明为新的 E2E PASS。

### Phase 24: Screenshot-aligned resume parsing transition

先扩展 `ResumeDetailView`、shared scene 与 CSS parity tests，锁定 `resume` illustration、TopBar/Resume active state、稳定 DOM、返回动作、无伪百分比、reduced-motion 和 mobile containment；当前小型 circle waiting block 必须先失败。随后只替换 queued/processing 无可读正文的 presentation，保留 `useResume` sequential polling、pending data、ready/failed atomic replacement、generated client 与 route。`BDD.RESUME.PARSE.VISUAL.005` 由 owner tests 与 current-run Chrome desktop/mobile 验收承接。

### Phase 25: Remove the shared preview backdrop

先以 component/CSS source tests 锁定 `ResumePreviewTab` 的直接 `article` 不携带 presentation class，并要求正式 CSS 不再定义零 consumer 的共享 preview-card selector。随后删除 desktop/mobile shared backdrop rules 与绑定旧 `1310px` 背景板的断言，只保留 PDF page-stack、Markdown page surface、`1512px` detail shell、只读 renderer 和数据合同。该阶段不修改 generated client、source endpoint、polling、Markdown/PDF 选择或路由。

### Phase 26: Resume delete secondary confirmation

先在 `ResumeListView.test.tsx` 与共享 destructive-dialog test 中锁定首次点击只打开 `role=dialog`、取消初始焦点、focus trap、Escape/遮罩关闭与 trigger focus restore，且确认前 `archiveResume` 调用数为 0。随后接入共享危险操作对话框：生成 dialog lifecycle 级 idempotency key，确认后单次归档；pending 禁止关闭和重复提交，失败保留卡片/弹窗并允许同 key 重试，成功关闭并隐藏卡片。中文/英文 copy 明确“从列表移除、当前无法撤销”，不谎称后端物理删除；不改 OpenAPI、fixture、backend、route 或数量上限事实。

### Phase 27: A4 preview geometry

先在 `ResumeWorkshopCssParity.test.ts` 锁定 `.ei-resume-detail-markdown-page` 与 `.ei-resume-detail-pdf-page` 共用 `width: min(100%, 794px)`，并拒绝旧 `1150px` 纸宽；PDF page 必须保留 `aspect-ratio: 210 / 297`，Markdown page 必须没有 `aspect-ratio`、fixed height 或 `min-height`，整份正文只形成一张连续长页面。RED 必须在当前 CSS 上失败。随后只修改正式 Resume detail CSS：PDF page wrapper 用 A4 纸宽与比例承接加载后的 canvas；PDF.js 只写 canvas intrinsic bitmap 尺寸，不写会覆盖 `width: 100%; height: auto` 的 inline presentation 尺寸，使原始页面按自身比例填满 A4 page content width。Markdown page 只锁定 A4 宽度并由内容自然撑高。不得改变 source URL、credential、page order、renderer choice 或正文内容。Mobile 沿用 `min(100%, 794px)` 在可用宽度内收敛，保留当前内边距与无横向溢出合同。完成 focused owner tests、typecheck/build、根 `make test`、owner/docs/index/diff gates，并用 Chrome skill 在真实 frontend/backend 上验收 PDF/Markdown desktop bbox 与 mobile no-overflow；该证据不声明新的 E2E ID。

### Phase 28: Markdown dark-mode paper ink

先在 `ResumeWorkshopCssParity.test.ts` 增加纸张局部颜色回归：`.ei-resume-detail-markdown-page` 必须同时声明白色背景与 WCAG AA 可读的深色正文墨水，并拒绝使用应用主题 foreground token；当前实现应因缺少局部 `color` 而 RED。随后只修改正式 Resume detail CSS，为白色 Markdown page 设置固定深色正文色，使 paragraph、list、list item、strong 与 inline code 通过继承获得同一纸张墨水；既有 heading/link/page geometry 规则、PDF renderer、应用壳 light/dark token 与 API/正文事实不变。Focused test GREEN 后运行 Resume owner tests、typecheck/build、根 `make test`，再用真实 Chrome 对比 dark/light 的 page/body/list/strong 计算色、截图、console 与 no-overflow；该 scoped UI evidence 不声明 E2E PASS。

## 5 验收标准

- 001 owner docs 只描述当前 flat Resume list / original-content read-only detail 合同。
- 列表只保留一个创建入口，以响应式卡片网格展示；每卡片支持打开与删除成功/失败状态，且无 table/header/row DOM。
- 详情 pending/ready/failed 三态可见；upload PDF 使用 PDF source page-stack preview 且无 native viewer toolbar，paste / Markdown / TXT 使用 Markdown DOM 渲染；inline `strong`、link、GFM list/table 语法不得以源码符号形式露出。
- Markdown body 不额外注入详情 header `displayName` / 名称 / summary；PDF 页面栈和 Markdown page surface 直接位于详情画布并共用 `794px` A4 纸宽；PDF 每页保持 `210:297`，Markdown 为不分页的连续内容高度；窄屏在可用宽度内收敛；外层语义 `article` 不形成额外背景板。
- Operation matrix 只列当前详情实际消费的 generated-client operations。
- 列表/Home 集合消费者只接收 closed `ResumeSummary`，完整 `Resume` 只由详情 route 获取；列表 fixture 与响应不含锁定的详情字段。
- StrictMode 下相同并发初始 GET 只有一次实际 transport；settled/rejected 请求会被驱逐，显式重试和后续合法轮询仍会发起新 transport。
- Focused frontend tests、context validation、docs/index gates 和 pruning surface lint 通过。

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| Fixture-backed UI 被误认为 real backend 闭环 | scenario trigger 保留 real-mode/generated-client gate，operation matrix 标明真实 handler / fixture 边界 |
| 仅在 hook 层隐藏重复调用，实际网络仍重复 | transport spy / fetch mock 直接断言底层次数，测试不把 method call count 当成网络证据 |
| in-flight 共享吞掉重试或轮询 | resolve/reject 均驱逐；不做 TTL cache；AbortSignal 请求不共享；queued polling 只在前次 settle 后发起 |
| 前端本地窄化掩盖服务端仍返回完整详情 | Phase 19 以 B2 generated schema、fixture 和 backend summary projection 为前置，禁止手写 DTO 充当合同完成证据 |

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-20 | 4.7 | Reopen Phase 28 so the white Markdown paper owns readable dark ink instead of inheriting the dark application foreground. |
| 2026-07-20 | 4.6 | Reopen Phase 27 so PDF and Markdown preview pages share a `794px` A4 width; PDF keeps `210:297`, Markdown remains one continuous content-height page, and both remain responsive. |
| 2026-07-20 | 4.5 | Reopen Phase 26 to require accessible secondary confirmation before resume archive while preserving the existing soft-delete API and list behavior. |
| 2026-07-20 | 4.4 | Reopen Phase 25 to remove the zero-consumer shared preview backdrop CSS/tests while preserving the PDF/Markdown page surfaces and readonly detail contract. |
| 2026-07-19 | 4.3 | Reopen Phase 24 for the supplied resume-parsing transition composition while retaining the no-flash polling contract. |
| 2026-07-19 | 4.2 | Reopen Phase 23 to implement the supplied resume preview Header, backdrop and paper composition rather than retaining the narrow historical reading frame. |
| 2026-07-19 | 4.1 | Reopen Phase 22 to align the Resume create icon with the Workspace circled-plus action. |
| 2026-07-19 | 4.0 | Reopen Phase 22 to align the Resume list hierarchy with the supplied desktop reference and the confirmed two-cards-per-row rule while preserving collection and detail contracts. |
| 2026-07-18 | 3.9 | Reopen Phase 21 to preserve the parse waiting DOM across background polls, remove loading-state flashes and eliminate geometry-changing icon motion. |
| 2026-07-15 | 3.8 | Reopen Phase 20 to replace the resume table with a fixed-width desktop and single-column mobile card grid while preserving closed-summary, open and archive behavior. |
| 2026-07-14 | 3.7 | Reopen Phase 19 for closed ResumeSummary list consumption and one underlying transport per StrictMode initial read, with reject eviction and retry coverage. |
| 2026-07-10 | 3.6 | Remove the empty pending-decision section from the active Resume Workshop spec. |
| 2026-07-10 | 3.5 | Consolidate the duplicate detail-back cascade and remove an ineffective flex-grid declaration. |
| 2026-07-10 | 3.4 | Delete zero-consumer breadcrumb, structured-preview and original-modal CSS from the read-only detail owner. |
| 2026-07-10 | 3.3 | Synchronize both pending-PDF tests on the ready visible heading instead of request count alone. |
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
