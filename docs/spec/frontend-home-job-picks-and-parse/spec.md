# Frontend Home / Parse Spec

> **版本**: 2.25
> **状态**: completed
> **更新日期**: 2026-07-14

## 1 背景与目标

`frontend-home-job-picks-and-parse` 是当前新建模拟面试入口的前端 owner，负责 `home` 与 command-only `parse` 进度屏，并维护统一面试规划详情母版组件。Canonical ready 详情 route 是 `/workspace?targetJobId=...`；`/parse?targetJobId=...` 只服务刚导入 TargetJob 的 queued/processing 阶段。

当前目标链路：

```text
Home 粘贴 JD
  -> 显式选择 ready Resume
  -> /parse?targetJobId queued/processing
  -> replace /workspace?targetJobId
  -> 面试规划详情 / 面试上下文确认
  -> 立即面试
  -> Practice handoff
```

本 subspec 维护 Home + Parse progress + 统一详情母版组件。Workspace `/workspace` 列表与 `/workspace?targetJobId` ready 详情 route、Practice、Report、Resume 管理、TargetJob 后端、Upload 后端、AI 解析与 persistence 分属各自 owner。

## 2 范围

### 2.1 In Scope

- Home 屏（`route=home`）：
  - 源级复刻 `ui-design/src/screen-home.jsx::HomeScreen` 当前结构。
  - Hero 只保留 label + title。
  - JD 输入卡只承载 textarea；不展示或挂载其他 JD intake 控件、弹窗或隐藏分支。
  - `listResumes` 读取 ready 且未归档的简历；用户必须显式选择一份简历后才能点击「立即面试」。
  - `还没有简历？1 分钟创建 ->` 与下拉框同行，点击进入 `resume_versions?flow=create`。
  - `listTargetJobs` 渲染最近 3 张 ready mock interview card；卡片主体直接进入 `/workspace?targetJobId=...`，不经过 Parse/动画；超过 3 条时展示「更多」并跳转 `/workspace`。
  - Empty state 引导继续创建模拟面试，不展示示例业务数据。
  - 未登录 import 的 `pendingAction` 只携带 `opaquePendingImportId`；`rawText`、`targetLanguage`、`resumeId` 与同一次 import 的 idempotency key 只存在于当前进程的一次性内存 vault，不进入 route 或任何浏览器持久化介质。
  - i18n 支持 zh/en，所有文案通过 typed locale helper。
- Parse command progress / Unified Plan Detail component（`route=parse` 仅首次导入处理中；`route=workspace&targetJobId` 独占 ready 详情）：
  - 源级复刻 `ui-design/src/screens-p0-complete.jsx::ParseScreen` 当前结构。
  - Loading 阶段只渲染 4 步进度与面向用户的等待说明；不得展示 model/provider、rubric/prompt/version/hash、provenance、典型耗时等内部调试或实现元数据。
  - 通过 generated `getTargetJob(targetJobId)` 分类/轮询 `analysisStatus`；仅 queued/processing 留在 Parse。首读 ready 或轮询转 ready 必须 replace 到 `/workspace?targetJobId=...`，不在 Parse 渲染 preview。
  - Workspace detail 用户可见名称为“面试规划详情 / 面试上下文确认”，只读展示 Basic fields、requirements evidence、hidden signals、round assumptions 和 TargetJob 已绑定 ready 简历摘要；详情初载只执行同 key `getTargetJob`，不调用 `listResumes`。
  - Workspace detail 内容区标题行右上角展示“面试报告”页面级入口，点击后仅携带当前可信 `targetJobId` 进入 `/reports?targetJobId=<uuid>`；该入口不加入全局 TopBar。
  - Workspace detail 不嵌入报告列表、不调用 `listTargetJobReports`，也不接受 `section=reports` 或其他报告相关 query；独立报告列表与状态由 `frontend-report-dashboard` owner 承接。Parse 不渲染 ready detail 或报告入口。
  - Workspace detail 不提供字段编辑、requirements toggle、hidden signal 移除、重新解析、保存规划、取消或更换简历入口；解析成功即表示规划已保存，若用户想换 JD/简历，必须回到 Home 创建新规划。
  - Footer actions 只保留「立即面试」，并从受保护 TargetJob 事实读取真实 `targetJobId`、`resumeId`、可选 `currentPracticePlanId` 和 `roundId` 进入 practice handoff；route 仍只携带 `targetJobId`。
  - 未登录启动通过 auth continuation 接续到 practice。
- Parity 与验证：
  - Home / Parse progress / Workspace detail 必须通过 Vitest + jsdom DOM 锚点、generated-client request、privacy checks、Playwright desktop/mobile pixel parity 与 BDD `E2E.P0.014` / `E2E.P0.015` / `E2E.P0.016` / `E2E.P0.018`。

### 2.2 Out of Scope

- `workspace` 无上下文规划列表、workspace 列表删除按钮和 workspace 列表启动实践编排：由 `frontend-workspace-and-practice` 承接。
- JD 文件、岗位链接与结构化表单导入：不属于当前产品范围；对应 UI、OpenAPI discriminator、generated artifacts、backend 分支、专属 fixture 与场景必须删除。Resume 模块自己的文件上传 / 粘贴创建继续由 Resume / Upload owner 承接。
- 独立于 Workspace targetJobId 母版之外的第二套 ready 详情页：不属于当前范围；Parse ready detail 必须为零，Workspace route 复用统一详情组件。
- `practice` / `report` / `resume_versions` 业务屏：由各自 subspec 承接。
- TargetJob import / parse / update handler、AI provider、prompt/rubric、DB migration、event/outbox：由 backend / contract owner 承接。
- 前端不直接调用 LLM、provider-specific endpoint、prompt registry 或 ad hoc parse fetch。

## 3 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | Owner 范围 | 本 subspec 接管 Home、Parse command progress 和被 Workspace targetJobId route 复用的统一详情组件 | Workspace 列表、Practice / Report / Resume 管理仍由各自 owner 承接 |
| D-2 | UI 真理源 | `ui-design/src/screen-home.jsx`、`ui-design/src/screens-p0-complete.jsx::ParseScreen`、`ui-design/src/primitives.jsx` | 正式前端必须源级复刻，不做二次设计 |
| D-3 | Home 提交流程 | 用户在唯一 textarea 粘贴 JD，并显式选择 ready 简历后提交 | POST 成功只进入 `/parse?targetJobId=...`；route 不携带 `resumeId` 或原文 |
| D-4 | Workspace detail handoff | Workspace targetJobId detail 是只读上下文收据；解析成功即已保存规划，唯一 footer CTA 是「立即面试」，直接使用已绑定上下文进入 practice handoff，不先 PATCH `updateTargetJob` 或经由 route-side auto start | Practice session 创建使用已保存 TargetJob / Resume / Round 快照 |
| D-5 | Parse 状态机 | `getTargetJob.analysisStatus` 驱动 queued/processing progress 与 failed；ready 必须 replace 到 workspace detail | Parse 是命令进度，不是 ready 详情回访页 |
| D-6 | Recent mocks | Home 最多展示 3 张最近模拟面试卡片，更多列表入口交给 `workspace` | 首页保持新建任务优先 |
| D-7 | i18n | 只维护当前 `home.*` 与 `parse.*` namespace | 与 typed locale helper 一致 |
| D-8 | Privacy / auth continuation | JD 原文不进入 URL/localStorage/sessionStorage/IndexedDB/console/telemetry；`pendingAction` 的唯一字段是 `opaquePendingImportId` | vault entry 仅在当前进程内保存 `{ rawText, targetLanguage, resumeId, idempotencyKey, expiresAt }` 并原子 consume 一次；refresh / 进程重启、过期或重复 consume 均 fail closed，返回 Home 显示本地化重新粘贴/选择提示，不发起 import，也不尝试从 route 或 storage 恢复原文 |
| D-9 | 统一详情母版 | 原 `JD 解析结果` 视觉改名为“面试规划详情 / 面试上下文确认”，只在 `/workspace?targetJobId` ready route 渲染；首次导入 ready 后 replace 到此，回访直接进入此页 | 用户只学习一个详情页面；Parse 不再作为 ready 详情入口 |
| D-10 | 结构化轮次数据源 | 所有 TargetJob 关联的轮次展示与导航上下文使用 `TargetJob.summary.interviewRounds[]`；数组长度必须为 2~5，轮次类型、标题、时长和 focus 均由后端 LLM 结合 JD、岗位级别、公司/行业性质、团队/业务上下文与招聘流程线索推断并持久化 | 避免 Parse、Home 最近卡片、Workspace 回访或共享上下文保留固定 4 轮 / 固定 HR/技术/经理面 / 固定时长模板 |
| D-11 | Recent card fixed grid and shared body | Home 最近模拟面试卡片使用固定最大列宽，并与 workspace 面试列表共用 `MockInterviewCard` 主体；Home 复用 `立即面试` 主按钮但不展示删除按钮 | 保证 Home recent 与 Interview list 不再表现为两套不同卡片规格 |
| D-12 | Recent card planning and start actions | Home ready 卡片主体直接进入 `/workspace?targetJobId=...`；`立即面试 / Start interview now` 仍用 generated practice handoff 启动 PracticeSession | 已解析规划不经过 Parse 动画；删除按钮只属于 workspace 列表 |
| D-13 | Parse loading 信息层级 | loading 只说明当前进度与等待状态，不暴露 model/provider、rubric/prompt/version/hash、provenance 或典型耗时 | 内部诊断信息留在受控日志/观测面，不进入用户界面 |
| D-14 | JD intake 单一合同 | Home 与 `importTargetJob` 只保留 `{ rawText, targetLanguage, resumeId }` | 不保留 source discriminator；删除其他 JD 导入形态但不影响 Resume 上传 |
| D-15 | 报告记录入口 | Workspace detail 内容区右上角提供页面级“面试报告”入口，进入 `/reports?targetJobId=<uuid>`；入口不进入全局 TopBar，Parse 不渲染 ready 入口、不嵌入或请求报告列表 | 用户在独立页面查看且只查看当前规划的轮次报告；Report/Generating 仍由各自 reportId-only 页面承接 |
| D-16 | Initial GET request count | Home `listResumes` / `listTargetJobs` 与 Parse 每个分类/调度 tick 的 `getTargetJob` 依赖 shell safe-read single-flight | StrictMode mount 不产生紧邻重复底层 GET；轮询只由明确 scheduler 在间隔到期后发起 |
| D-17 | Workspace detail round state | 轮次假设卡片使用与 Home/Workspace mini rail 相同的 `practiceProgress` 投影：完成前缀为 `done/已进行`，首个未完成轮为 `current/即将进行`，其余为 `pending/未进行`；三态必须同时有不同背景、边框、文字标签和可测试状态属性 | 用户无需从顺序猜测进度；不从 lifecycle status、URL 或浏览器存储推断轮次状态 |

## 4 设计约束

- DOM 构图、控件类型、间距、字体层级、状态、响应式行为和交互节奏必须可追溯到 `ui-design/` 当前源码。
- Home `home-jd-input-card` 只承载 `home-jd-textarea`；`home-resume-row` 与 `home-submit-row` 位于输入卡下方。旧 source controls、trigger 和 modal 锚点必须为零。
- Home resume select 使用紧凑下拉框；不得平铺所有简历。
- 未登录提交时先创建不可逆推原文的 `opaquePendingImportId`，再把 exact import intent 写入一次性内存 vault；认证路由的 `pendingAction` 不得复制 `rawText`、`targetLanguage`、`resumeId`、intake source 或业务 route params。登录成功后必须原子 consume 一次并使用 vault 中原 idempotency key 提交 exact request；成功、失败、过期和重复 consume 后均不得让同一 entry 再次可读。
- refresh / 进程重启导致 vault 丢失、entry 过期或 ID 已消费时，auth continuation 不调用 `importTargetJob`，清除无效 pending action，返回 Home 并以 zh/en 可访问提示要求用户重新粘贴 JD、选择简历；不得用 `localStorage`、`sessionStorage`、IndexedDB、URL、日志或 telemetry 延长 raw JD 生命周期。
- `route=parse` 只在 queued/processing 展示 loading。首个 `getTargetJob` 已 ready 或轮询转 ready 时立即 `replaceRoute({ name: "workspace", params: { targetJobId } })`；不得先播放动画或把 Parse 留在 Back history。Workspace detail 直接渲染同一 ready 母版。
- Home `listResumes` / `listTargetJobs` 与 Parse `getTargetJob` 使用 shell safe-read single-flight + 稳定 loader dependencies；同 key 初载底层 request count 必须为 1，轮询后续请求必须有 scheduler tick 证据。
- Parse loading 的 DOM、截图和文案负向 gate 必须拒绝 `model`、`provider`、`rubric`、`prompt@`、版本/hash、`provenance`、`typical` 等内部实现标记；不能以折叠、弱化颜色或移动到底部代替删除。
- Workspace detail requirements evidence 只读展示 API 返回的 `evidenceLevel`；前端不得在详情页维护临时 hit toggle 或把确认状态写回后端。
- Workspace detail round assumptions 的卡片布局仍追溯 UI 真理源，但卡片数量必须来自 2~5 条 `TargetJob.summary.interviewRounds[]`；R 序号、标题、轮次类型、时长和 focus 也必须来自该数组。这些轮次由后端 LLM 根据 JD、行业/公司性质、岗位级别、团队/业务上下文和招聘流程线索推断，前端不得用 locale 或本地常量补齐轮数、HR/技术/经理面类型或分钟数。
- Workspace detail round assumptions 必须复用 `resolveTargetJobPracticeProgress` 的 completed-prefix/current-first-incomplete 事实，不另建状态机：`done` 使用成功色 soft 背景与边框，`current` 使用主题 accent soft 背景与边框，`pending` 使用中性 soft 背景与规则线；每张卡同时暴露 `data-round-state` 和本地化的“已进行 / 即将进行 / 未进行”文本。全完成时全部为 done；无效投影不得制造 current/done。
- Workspace detail 的“面试报告”入口只在可信 ready TargetJob 上下文存在时可用，导航参数精确为 `{ targetJobId }`；入口不得复制 reportId/status/round 等业务事实，也不得写入全局 TopBar。
- Parse 与 Workspace detail DOM/effect/generated-client spy/route gate 必须证明报告列表、列表 loading/error/empty state、`listTargetJobReports` 请求和 `section=reports` 兼容逻辑全部不存在。未知 `section` 与报告相关 query 由路由层剔除，不能切换 TargetJob、report identity 或业务状态。
- Home 最近模拟面试卡片与 Workspace plan handoff 不得维护独立 `DEFAULT_ROUNDS`、固定 4 轮、静态 `roundName` 或静态 duration 分支；相关显示或 route params 必须通过同一个 TargetJob structured round mapper 派生。
- `importTargetJob` 是 side-effect operation，必须携带 `Idempotency-Key`；其请求体严格等于 `{ rawText, targetLanguage, resumeId }`，不得带 `source` 或其他 intake-only 字段。
- Workspace success detail 不调用 `updateTargetJob`；规划上下文来自已保存的 TargetJob + Resume binding 快照。
- Dark / customAccent 必须在 Home 与 Parse 生效；移动端不得横向溢出。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| Home / Parse progress / Unified Workspace Detail UI | `frontend-home-job-picks-and-parse` | React 组件、route 业务内容、i18n、source parity、pixel parity；ready 详情只由 Workspace targetJobId route 复用 |
| App shell / auth / runtime | `frontend-shell` | TopBar、route normalization、auth continuation、generated client bootstrap |
| TargetJobs API | `openapi-v1-contract` + `backend-targetjob` | `importTargetJob` / `listTargetJobs` / `getTargetJob` schema、fixtures、handler；`updateTargetJob` 仍属后端 TargetJobs 合同但不是 Parse preview 成功态 consumer |
| Resume upload | `backend-upload` + `backend-resume` | Resume 文件上传与 file object persistence 继续保留；Home JD intake 不消费该能力 |
| Resume list | `backend-resume` | `listResumes` 只读 ready resume selection |
| Practice handoff | `frontend-workspace-and-practice` | PracticePlan / PracticeSession 创建与 practice 跳转；workspace 仅作为列表回访入口，带上下文详情复用统一母版 |
| Mock transport | `mock-contract-suite` | fixture-backed deterministic variants |
| Reports / Report / Generating screens | `frontend-report-dashboard` | 独立 `/reports?targetJobId=...` 列表消费最小 canonical-round overview；Report/Generating 保持 reportId-only，并在可信规划上下文存在时返回该列表 |

## 6 Operation Matrix

| operationId | Fixture | Frontend consumer | Backend handler | Persistence | AI dependency | Scenario |
|-------------|---------|-------------------|-----------------|-------------|---------------|----------|
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json` | Home recent ready cards；same-key initial underlying request count = 1 | `backend-targetjob` | `target_jobs` / `target_job_requirements` read | none in frontend | `E2E.P0.014` |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` | Home ready resume select；same-key initial underlying request count = 1 | `backend-resume` | `resumes` read | none | `E2E.P0.014` / `E2E.P0.015` |
| `importTargetJob` | `openapi/fixtures/TargetJobs/importTargetJob.json`（paste success + current validation/failure variants） | Home submits `{ rawText, targetLanguage, resumeId }` | `backend-targetjob` | `target_jobs` create + saved `resume_id` + parse job；无并行 source-specific persistence | backend-only parse job | `E2E.P0.015` |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` | Parse classification/scheduled polling + Workspace ready detail；each same-key tick count = 1 | `backend-targetjob` | `target_jobs.summary` / requirements read | backend-generated `target.import.parse` structured rounds | `E2E.P0.015` / `E2E.P0.016` / `E2E.P0.018` |
| `createPracticePlan` / `getPracticePlan` / `startPracticeSession` | `openapi/fixtures/PracticePlans/*`, `openapi/fixtures/PracticeSessions/*` | Workspace readonly detail Start action and Home recent quick start | `backend-practice` | `practice_plans` / `practice_sessions` create/read | none | `E2E.P0.016`, `E2E.P0.018` |

## 7 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Home 默认渲染 | 用户进入 App | 打开 `home` | Hero、唯一 JD textarea、resume select、create resume CTA、recent mocks/empty state 正常渲染；旧 source controls / trigger / modal 锚点不存在；TopBar 高亮首页 | 001 |
| C-2 | Home resume gate | `listResumes` 返回 ready 简历 | 用户尚未选择简历 | 「立即面试」disabled，不调用 import；选择 ready 简历后才允许提交 | 001 |
| C-3 | Paste JD import | 用户选择 ready 简历并粘贴 JD | 点击「立即面试」 | 调用 `importTargetJob({ rawText, targetLanguage, resumeId })` 并携带 `Idempotency-Key`；POST 成功只进入 `/parse?targetJobId=...` | 001 |
| C-4 | 非当前 JD intake 零残留 | paste-only 合同已生效 | 扫描 UI 真理源、formal frontend、OpenAPI/generated、backend、active fixtures/scenarios | 不存在平行 JD intake UI、source discriminator、专属 handler/persistence/job/scenario；Resume 上传路径仍通过原 owner gate | 001 |
| C-5 | Recent mocks | `listTargetJobs` 返回多条 ready 记录 | Home 加载完成 | Home `listTargetJobs`/`listResumes` 同 key 初载底层请求各 1 次；只展示最近 3 张；卡片主体直达 `/workspace?targetJobId=...`，不经过 Parse/动画；quick-start 与 More 保持 | 001 |
| C-6 | Parse ready replace | `getTargetJob` 首读 ready 或 queued/processing 轮询转 ready | 用户进入 `/parse?targetJobId=...` | 每个分类/调度 tick 同 key底层 GET 恰好 1；ready 立即 replace 到 `/workspace?targetJobId=...`，Back 不返回动画；Parse 不渲染 ready detail | 001 |
| C-7 | Parse failed flow | `analysisStatus=failed` 或轮询超时 | Parse polling | 渲染失败态、重新解析和返回首页；不伪造 preview 数据 | 001 |
| C-8 | Readonly plan receipt | Workspace detail 已绑定 ready 简历 | 用户查看详情 | 初载同 key `getTargetJob` 底层 count=1 且不调用 `listResumes`；不出现字段编辑、requirements toggle、hidden signal 移除、重新解析、保存规划、取消或更换简历入口；缺少绑定简历时只阻断开始，不提供 picker 兜底 | 001 |
| C-9 | Start interview | Workspace detail 已绑定 ready 简历 | 点击「立即面试」 | 不调用 `updateTargetJob`，直接使用已保存 `targetJobId/resumeId/roundId/currentPracticePlanId` 创建或读取 PracticePlan 并启动 PracticeSession | 001 |
| C-10 | Privacy / auth continuation | 未登录用户提交 JD，随后正常登录、刷新导致 vault 丢失、entry 过期或重复触发 continuation | 检查 pendingAction、vault consume、URL/storage/log/telemetry 与 import 调用 | `pendingAction` 只含 `opaquePendingImportId`；raw JD 只在一次性内存 vault。正常登录原子 consume 后 exact request 只提交一次；refresh/expired/duplicate consume 均不发起 import，清除无效 action 并返回 Home 显示本地化重新输入提示；fixture transport 不记录 request body | 001 |
| C-11 | Workspace 回访统一详情 | `listTargetJobs` 返回已保存 ready 规划且有 `targetJobId/resumeId` | 用户从 Home 或 workspace 卡片打开 | 直达 `/workspace?targetJobId=...` 并渲染同一详情母版；无 Parse loading/animation；返回 `/workspace` 列表 | 001 / frontend-workspace-and-practice 001 |
| C-12 | 规划详情报告入口 | ready TargetJob 有可信 `targetJobId` | 打开 Workspace 面试规划详情并点击标题行右上角“面试报告” | 精确进入 `/reports?targetJobId=<uuid>`；入口不在全局 TopBar；Parse 不渲染 ready detail/报告入口，Workspace detail 不渲染报告列表、不调用 `listTargetJobReports`、不保留 `section=reports` 兼容逻辑，Start 不受影响 | 001 |
| C-13 | 规划详情轮次三态 | ready TargetJob 有合法 2~5 轮与 `practiceProgress` 完成前缀/current | 打开或刷新 `/workspace?targetJobId=...` | 每张 round assumption 卡显示且仅显示 done/current/pending 之一，对应“已进行 / 即将进行 / 未进行”及三种不同背景/边框；状态与列表 mini rail 一致，全完成与无效投影 fail closed | 001 |

## 8 关联计划

- [001-home-jd-import-and-parse](./plans/001-home-jd-import-and-parse/plan.md) — Home + Parse command progress + Workspace unified plan detail 当前 owner 计划，覆盖 paste-only UI/API contract、generated-client request、real-mode gate、resume selection、recent mocks、规划详情报告入口、ready replace/direct detail handoff 和 P0.014-P0.016/P0.018 BDD。

## 9 关联文档

- 上游 spec：[`product-scope`](../product-scope/spec.md)、[`engineering-roadmap`](../engineering-roadmap/spec.md)、[`frontend-shell`](../frontend-shell/spec.md)、[`frontend-workspace-and-practice`](../frontend-workspace-and-practice/spec.md)、[`openapi-v1-contract`](../openapi-v1-contract/spec.md)、[`mock-contract-suite`](../mock-contract-suite/spec.md)
- UI 真理源：`ui-design/src/screen-home.jsx`、`ui-design/src/screens-p0-complete.jsx::ParseScreen`、`ui-design/src/primitives.jsx`、[`docs/ui-design/jd-resume-management.md`](../../ui-design/jd-resume-management.md)、[`docs/ui-design/ui-architecture.md`](../../ui-design/ui-architecture.md)、[`docs/ui-design/module-job-workspace.md`](../../ui-design/module-job-workspace.md)、[`docs/ui-design/module-map.md`](../../ui-design/module-map.md)
- 当前正式前端入口：`frontend/src/app/screens/home/`、`frontend/src/app/screens/parse/`、`frontend/src/app/navigation/interviewContext.ts`、`frontend/src/api/generated/`
- 场景：`test/scenarios/e2e/p0-014-home-default-render/`、`test/scenarios/e2e/p0-015-jd-import-and-parse/`、`test/scenarios/e2e/p0-016-parse-confirm-to-workspace/`
- 变更记录：[history.md](./history.md)

## 10 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 2.25 | 2026-07-14 | Add Workspace detail round-assumption done/current/pending visual states derived from the same backend practice-progress projection as list-card rails. |
| 2.24 | 2026-07-14 | Make Parse command-progress-only, replace ready targets into Workspace detail, route ready cards directly, move ready detail/report/start language to Workspace, and require exact initial GET counts. |
