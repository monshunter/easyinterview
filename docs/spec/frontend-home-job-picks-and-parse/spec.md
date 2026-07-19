# Frontend Home / Parse Spec

> **版本**: 2.30
> **状态**: completed
> **更新日期**: 2026-07-19

## 1 背景与目标

`frontend-home-job-picks-and-parse` 是当前新建模拟面试入口的前端 owner，负责 `home` 与 command-only `parse` 进度屏，并维护统一面试规划详情母版组件。Canonical ready 详情 route 是 `/workspace?targetJobId=...`；`/parse?targetJobId=...` 只服务刚导入 TargetJob 的 queued/processing 阶段。

当前目标链路：

```text
Home 粘贴 JD
  -> 显式选择 selectable Resume
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
  - 按设计合同实现 `frontend/src` 当前结构。
  - Hero 使用主标题双层强调、副标题和右侧轻量插画；不展示旧 uppercase eyebrow。
  - 单一白色 intake card 依次承载 JD label + textarea/count、Resume label + select/create link、右侧主 CTA 与隐私提示；`home-jd-input-card` 仍只包裹 textarea/count，不展示或挂载其他 JD intake 控件、弹窗或隐藏分支。
  - `listResumes` 读取未归档且 `parseStatus=ready` 或已有可读正文/结构化证据的 selectable 简历；用户必须显式选择一份后才能点击「立即面试」。
  - `还没有简历？1 分钟创建 ->` 与下拉框同行，点击进入 `resume_versions?flow=create`。
  - `listTargetJobs` 渲染最近 3 条 ready mock interview record；Home 使用全宽横向列表形态，依次展示公司/岗位、动态轮次 rail、最近使用时间与继续练习；主体直接进入 `/workspace?targetJobId=...`，不经过 Parse/动画；有记录时展示「查看全部」并跳转 `/workspace`。
  - Empty state 引导继续创建模拟面试，不展示示例业务数据。
  - 未登录 import 的 `pendingAction` 只携带 `opaquePendingImportId`；`rawText`、`targetLanguage`、`resumeId` 与同一次 import 的 idempotency key 只存在于当前进程的一次性内存 vault，不进入 route 或任何浏览器持久化介质。
  - i18n 支持 zh/en，所有文案通过 typed locale helper。
- Parse command progress / Unified Plan Detail component（`route=parse` 仅首次导入处理中；`route=workspace&targetJobId` 独占 ready 详情）：
  - 按设计合同实现 `frontend/src` 当前结构。
  - Loading 阶段只渲染 4 步进度与面向用户的等待说明；不得展示 model/provider、rubric/prompt/version/hash、provenance、典型耗时等内部调试或实现元数据。
  - 通过 generated `getTargetJob(targetJobId)` 分类/轮询 `analysisStatus`；仅 queued/processing 留在 Parse。首读 ready 或轮询转 ready 必须 replace 到 `/workspace?targetJobId=...`，不在 Parse 渲染 preview。
  - Workspace detail 用户可见名称为“面试规划详情 / 面试上下文确认”，只读展示 Basic fields、requirements evidence、hidden signals 和 round assumptions；标题旁的“绑定简历”只使用 TargetJob 已保存 `resumeId` 跳转对应 `resume_versions` 详情。详情初载只执行同 key `getTargetJob`，不调用 `listResumes` 或 `getResume`。
  - Workspace detail 标题下方首行动作行从左依次展示“立即面试”与“面试报告”；desktop 同排，mobile 同序响应式换行。报告入口只携带当前可信 `targetJobId` 进入 `/reports?targetJobId=<uuid>`，不加入全局 TopBar 或页尾。
  - Workspace detail 不嵌入报告列表、不调用 `listTargetJobReports`，也不接受 `section=reports` 或其他报告相关 query；独立报告列表与状态由 `frontend-report-dashboard` owner 承接。Parse 不渲染 ready detail 或报告入口。
  - Workspace detail 不提供字段编辑、requirements toggle、hidden signal 移除、重新解析、保存规划、取消或更换简历入口；标题旁绑定简历是查看链接而非 rebind。解析成功即表示规划已保存，若用户想换 JD/简历，必须回到 Home 创建新规划。
  - 删除独立 Interview Launch / 绑定简历大卡片和 Footer actions；「立即面试」前移至首行动作行，并从受保护 TargetJob 事实读取真实 `targetJobId`、`resumeId`、可选 `currentPracticePlanId` 和 `roundId` 进入 practice handoff；route 仍只携带 `targetJobId`。
  - 未登录启动通过 auth continuation 接续到 practice。
- Parity 与验证：
  - Home / Parse / Workspace detail 的前后端单测完成由根 `make test` 统一承接；正式前端 component、responsive 与 accessibility assertions 独立执行。JD import/parse 当前没有真实 E2E owner。

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
| D-2 | UI 设计文档 | `frontend/src` | 正式前端必须按设计合同实现，不做二次设计 |
| D-3 | Home 提交流程 | 用户在唯一 textarea 粘贴 JD，并显式选择 selectable 简历后提交 | POST 成功只进入 `/parse?targetJobId=...`；route 不携带 `resumeId` 或原文 |
| D-4 | Workspace detail handoff | Workspace targetJobId detail 是只读上下文收据；解析成功即已保存规划。标题旁“绑定简历”只查看 saved resume；标题下首行动作行的「立即面试」直接使用已绑定上下文进入 practice handoff，不先 PATCH `updateTargetJob` 或经由 route-side auto start | Practice session 创建使用已保存 TargetJob / Resume / Round 快照；不保留独立 launch/footer 区 |
| D-5 | Parse 状态机 | `getTargetJob.analysisStatus` 驱动 queued/processing progress 与 failed；ready 必须 replace 到 workspace detail | Parse 是命令进度，不是 ready 详情回访页 |
| D-6 | Recent mocks | Home 最多展示 3 条最近模拟面试横向记录，有记录时「查看全部」交给 `workspace` | 首页保持新建任务优先 |
| D-7 | i18n | 只维护当前 `home.*` 与 `parse.*` namespace | 与 typed locale helper 一致 |
| D-8 | Privacy / auth continuation | JD 原文不进入 URL/localStorage/sessionStorage/IndexedDB/console/telemetry；`pendingAction` 的唯一字段是 `opaquePendingImportId` | vault entry 仅在当前进程内保存 `{ rawText, targetLanguage, resumeId, idempotencyKey, expiresAt }` 并原子 consume 一次；refresh / 进程重启、过期或重复 consume 均 fail closed，返回 Home 显示本地化重新粘贴/选择提示，不发起 import，也不尝试从 route 或 storage 恢复原文 |
| D-9 | 统一详情母版 | 原 `JD 解析结果` 视觉改名为“面试规划详情 / 面试上下文确认”，只在 `/workspace?targetJobId` ready route 渲染；首次导入 ready 后 replace 到此，回访直接进入此页 | 用户只学习一个详情页面；Parse 不再作为 ready 详情入口 |
| D-10 | 结构化轮次数据源 | 所有 TargetJob 关联的轮次展示与导航上下文使用 `TargetJob.summary.interviewRounds[]`；数组长度必须为 2~5，轮次类型、标题、时长和 focus 均由后端 LLM 结合 JD、岗位级别、公司/行业性质、团队/业务上下文与招聘流程线索推断并持久化 | 避免 Parse、Home 最近卡片、Workspace 回访或共享上下文保留固定 4 轮 / 固定 HR/技术/经理面 / 固定时长模板 |
| D-11 | Recent record and shared business mapper | Home recent 使用全宽横向 record，Workspace 保持固定最大列宽 card；两种 presentation 共用 `MockInterviewCard` 的 TargetJob/round/progress/action mapper | Home 可按参考图表达信息密度，同时不复制业务推导、路由或启动逻辑 |
| D-12 | Recent card planning and start actions | Home ready 卡片主体直接进入 `/workspace?targetJobId=...`；`立即面试 / Start interview now` 仍用 generated practice handoff 启动 PracticeSession | 已解析规划不经过 Parse 动画；删除按钮只属于 workspace 列表 |
| D-13 | Parse loading 信息层级 | loading 只说明当前进度与等待状态，不暴露 model/provider、rubric/prompt/version/hash、provenance 或典型耗时 | 内部诊断信息留在受控日志/观测面，不进入用户界面 |
| D-14 | JD intake 单一合同 | Home 与 `importTargetJob` 只保留 `{ rawText, targetLanguage, resumeId }` | 不保留 source discriminator；删除其他 JD 导入形态但不影响 Resume 上传 |
| D-15 | JD raw text runtime limit | Home 只消费 RuntimeConfig `targetJobRawTextBytes`，默认/code fallback 98,304 bytes；通过 `TextEncoder` 按 UTF-8 bytes 预检，limit+1 不调用 import；backend 仍作最终裁决 | 不改变 textarea DOM/样式，只统一数据源、字节口径与本地化错误；不把 route/vault/browser storage 当限制事实源 |
| D-15 | 报告记录入口 | Workspace detail 标题下方首行动作行在“立即面试”之后提供“面试报告”，两者左对齐同排；进入 `/reports?targetJobId=<uuid>`。入口不进入全局 TopBar/页尾，Parse 不渲染 ready 入口、不嵌入或请求报告列表 | 用户在独立页面查看且只查看当前规划的轮次报告；Report/Generating 仍由各自 reportId-only 页面承接 |
| D-16 | 绑定简历查看入口 | 删除独立绑定 block；“绑定简历”放在“面试规划详情”旁，点击只使用 `TargetJob.resumeId` 进入 `resume_versions?resumeId=...`；缺失绑定显示非链接异常状态，Start、Reports、复练和下一轮全部 fail closed | 不新增 `getResume`/`listResumes` 读取，不把查看入口误作 rebind，也不从 route/list/recent resume 伪造绑定 |
| D-16 | Initial GET request count | Home `listResumes` / `listTargetJobs` 与 Parse 每个分类/调度 tick 的 `getTargetJob` 依赖 shell safe-read single-flight | StrictMode mount 不产生紧邻重复底层 GET；轮询只由明确 scheduler 在间隔到期后发起 |
| D-17 | Workspace detail round state | 轮次假设卡片使用与 Home/Workspace mini rail 相同的 `practiceProgress` 投影：完成前缀为 `done/已进行`，首个未完成轮为 `current/即将进行`，其余为 `pending/未进行`；三态必须同时有不同背景、边框、文字标签和可测试状态属性 | 用户无需从顺序猜测进度；不从 lifecycle status、URL 或浏览器存储推断轮次状态 |
| D-18 | Selectable Resume 永久前置 | Home 只有在用户显式选择未归档且 `parseStatus=ready` 或已有可读正文/结构化证据的 selectable Resume 后才能提交 exact import；TargetJob 必须保存该 `resumeId`，后续 Start、Reports、复练和下一轮都只消费该持久化事实 | 不实现无简历/JD-only 训练或报告降级；无 selectable 简历的用户只进入创建流程。历史缺失或无效绑定是异常数据并 fail closed，不自动选择最近简历，不从 route/browser storage 补齐 |
| D-19 | Home screenshot-aligned visual hierarchy | Desktop Home 使用 1400px 级居中内容列、浅色渐变/斜切背景、标题强调、单一 intake card 与全宽 recent record；mobile 按 DOM 顺序收敛为单列 | 视觉重排不改变 operation matrix、Resume gate、route、privacy、idempotency 或 TargetJob round mapper；计数器必须显示 runtime owner 的真实上限，不硬编码参考图中的业务值 |

## 4 设计约束

- DOM 构图、控件类型、间距、字体层级、状态、响应式行为和交互节奏必须可追溯到 `frontend/` 当前源码。
- Home `home-intake-card` 是单一视觉容器；其中 `home-jd-input-card` 只承载 `home-jd-textarea` 与真实 runtime count，`home-resume-row` / `home-submit-row` / `home-privacy-note` 同属该视觉容器但不是 textarea DOM 的子节点。旧 source controls、trigger 和 modal 锚点必须为零。
- Home resume select 使用紧凑下拉框；不得平铺所有简历。
- 未登录提交时先创建不可逆推原文的 `opaquePendingImportId`，再把 exact import intent 写入一次性内存 vault；认证路由的 `pendingAction` 不得复制 `rawText`、`targetLanguage`、`resumeId`、intake source 或业务 route params。登录成功后必须原子 consume 一次并使用 vault 中原 idempotency key 提交 exact request；成功、失败、过期和重复 consume 后均不得让同一 entry 再次可读。
- refresh / 进程重启导致 vault 丢失、entry 过期或 ID 已消费时，auth continuation 不调用 `importTargetJob`，清除无效 pending action，返回 Home 并以 zh/en 可访问提示要求用户重新粘贴 JD、选择简历；不得用 `localStorage`、`sessionStorage`、IndexedDB、URL、日志或 telemetry 延长 raw JD 生命周期。
- `route=parse` 只在 queued/processing 展示 loading。首个 `getTargetJob` 已 ready 或轮询转 ready 时立即 `replaceRoute({ name: "workspace", params: { targetJobId } })`；不得先播放动画或把 Parse 留在 Back history。Workspace detail 直接渲染同一 ready 母版。
- Home `listResumes` / `listTargetJobs` 与 Parse `getTargetJob` 使用 shell safe-read single-flight + 稳定 loader dependencies；同 key 初载底层 request count 必须为 1，轮询后续请求必须有 scheduler tick 证据。
- Parse loading 的 DOM、截图和文案负向 gate 必须拒绝 `model`、`provider`、`rubric`、`prompt@`、版本/hash、`provenance`、`typical` 等内部实现标记；不能以折叠、弱化颜色或移动到底部代替删除。
- Workspace detail requirements evidence 只读展示 API 返回的 `evidenceLevel`；前端不得在详情页维护临时 hit toggle 或把确认状态写回后端。
- Workspace detail round assumptions 的卡片布局仍追溯 UI 设计文档，但卡片数量必须来自 2~5 条 `TargetJob.summary.interviewRounds[]`；R 序号、标题、轮次类型、时长和 focus 也必须来自该数组。这些轮次由后端 LLM 根据 JD、行业/公司性质、岗位级别、团队/业务上下文和招聘流程线索推断，前端不得用 locale 或本地常量补齐轮数、HR/技术/经理面类型或分钟数。
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
| Home / Parse progress / Unified Workspace Detail UI | `frontend-home-job-picks-and-parse` | React 组件、route 业务内容、i18n、formal component/responsive/accessibility contract；ready 详情只由 Workspace targetJobId route 复用 |
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
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json` | Home recent ready cards；same-key initial underlying request count = 1 | `backend-targetjob` | `target_jobs` / `target_job_requirements` read | none in frontend | `E2E.P0.098` 仅 Home progress refresh |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` | Home selectable resume list（ready 或有可读证据）；same-key initial underlying request count = 1 | `backend-resume` | `resumes` read | none | 当前无真实 E2E owner；root `make test` |
| `importTargetJob` | `openapi/fixtures/TargetJobs/importTargetJob.json`（paste success + current validation/failure variants） | Home submits `{ rawText, targetLanguage, resumeId }` | `backend-targetjob` | `target_jobs` create + saved `resume_id` + parse job；无并行 source-specific persistence | backend-only parse job | 当前无真实 E2E owner；root `make test` |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` | Parse classification/scheduled polling + Workspace ready detail；each same-key tick count = 1 | `backend-targetjob` | `target_jobs.summary` / requirements read | backend-generated `target.import.parse` structured rounds | `E2E.P0.098` 仅 TargetJob progress/detail read；import/parse 无 owner |
| `createPracticePlan` / `getPracticePlan` / `startPracticeSession` | `openapi/fixtures/PracticePlans/*`, `openapi/fixtures/PracticeSessions/*` | Workspace readonly detail Start action and Home recent quick start | `backend-practice` | `practice_plans` / `practice_sessions` create/read | none | 当前无真实 E2E owner；root `make test` |

## 7 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Home 默认渲染 | 用户进入 App | 打开 `home` | screenshot-aligned Hero/subtitle/illustration、单一 intake card、唯一 JD textarea + 真实上限 count、resume select/create CTA、recent records/empty state 正常渲染；旧 source controls / trigger / modal 锚点不存在；TopBar 高亮首页 | 001 |
| C-2 | Home resume gate | `listResumes` 返回 selectable 简历 | 用户尚未选择简历 | 「立即面试」disabled，不调用 import；选择 selectable 简历后才允许提交 | 001 |
| C-3 | Paste JD import | 用户选择 selectable 简历并粘贴 JD | 点击「立即面试」 | 调用 `importTargetJob({ rawText, targetLanguage, resumeId })` 并携带 `Idempotency-Key`；POST 成功只进入 `/parse?targetJobId=...` | 001 |
| C-4 | 非当前 JD intake 零残留 | paste-only 合同已生效 | 扫描 UI 设计文档、formal frontend、OpenAPI/generated、backend、active fixtures/scenarios | 不存在平行 JD intake UI、source discriminator、专属 handler/persistence/job/scenario；Resume 上传路径仍通过原 owner gate | 001 |
| C-5 | Recent mocks | `listTargetJobs` 返回多条 ready 记录 | Home 加载完成 | Home `listTargetJobs`/`listResumes` 同 key 初载底层请求各 1 次；只展示最近 3 条全宽横向 record；主体直达 `/workspace?targetJobId=...`，不经过 Parse/动画；quick-start 与「查看全部」保持 | 001 |
| C-6 | Parse ready replace | `getTargetJob` 首读 ready 或 queued/processing 轮询转 ready | 用户进入 `/parse?targetJobId=...` | 每个分类/调度 tick 同 key底层 GET 恰好 1；ready 立即 replace 到 `/workspace?targetJobId=...`，Back 不返回动画；Parse 不渲染 ready detail | 001 |
| C-7 | Parse failed flow | `analysisStatus=failed` 或轮询超时 | Parse polling | 渲染失败态、重新解析和返回首页；不伪造 preview 数据 | 001 |
| C-8 | Readonly plan receipt | Workspace detail 已绑定 selectable 简历，或历史 TargetJob 缺失/无效绑定 | 用户查看详情 | 初载同 key `getTargetJob` 底层 count=1 且不调用 `listResumes/getResume`；合法绑定精确进入对应 Resume 详情；无独立 binding/launch block、字段编辑、picker/rebind 或页尾动作；缺绑显示异常状态，Start、Reports、复练和下一轮全部 fail closed | 001 |
| C-9 | Start interview | Workspace detail 已绑定 selectable 简历 | 点击首行动作行「立即面试」 | 不调用 `updateTargetJob`，直接使用已保存 `targetJobId/resumeId/roundId/currentPracticePlanId` 创建或读取 PracticePlan 并启动 PracticeSession；同排“面试报告”不受启动错误影响 | 001 |
| C-10 | Privacy / auth continuation | 未登录用户提交 JD，随后正常登录、刷新导致 vault 丢失、entry 过期或重复触发 continuation | 检查 pendingAction、vault consume、URL/storage/log/telemetry 与 import 调用 | `pendingAction` 只含 `opaquePendingImportId`；raw JD 只在一次性内存 vault。正常登录原子 consume 后 exact request 只提交一次；refresh/expired/duplicate consume 均不发起 import，清除无效 action 并返回 Home 显示本地化重新输入提示；fixture transport 不记录 request body | 001 |
| C-11 | Workspace 回访统一详情 | `listTargetJobs` 返回已保存 ready 规划且有 `targetJobId/resumeId` | 用户从 Home 或 workspace 卡片打开 | 直达 `/workspace?targetJobId=...` 并渲染同一详情母版；无 Parse loading/animation；返回 `/workspace` 列表 | 001 / frontend-workspace-and-practice 001 |
| C-12 | 规划详情报告入口 | ready TargetJob 有可信 `targetJobId` | 打开 Workspace 面试规划详情并点击首行动作行“面试报告” | “立即面试 + 面试报告”从左同排且顺序稳定，报告精确进入 `/reports?targetJobId=<uuid>`；入口不在全局 TopBar/页尾；Parse 不渲染 ready detail/报告入口，Workspace detail 不渲染报告列表、不调用 `listTargetJobReports`、不保留 `section=reports` 兼容逻辑 | 001 |
| C-13 | 规划详情轮次三态 | ready TargetJob 有合法 2~5 轮与 `practiceProgress` 完成前缀/current | 打开或刷新 `/workspace?targetJobId=...` | 每张 round assumption 卡显示且仅显示 done/current/pending 之一，对应“已进行 / 即将进行 / 未进行”及三种不同背景/边框；状态与列表 mini rail 一致，全完成与无效投影 fail closed | 001 |
| C-14 | JD size boundary | owner config 提供 UTF-8 JD byte limit | 点击「立即面试」 | 注入小型 boundary 验证 overflow inline validation 且零 import；默认/override/invalid 由 typed config owner 覆盖，不构造默认大小文本或配置 E2E | 001 Phase 22 |
| C-15 | 强制简历前置零残留 | 用户没有 selectable 简历或尚未显式选择 | 输入合法 JD 并尝试提交，随后扫描 active 产品/UI/owner 文档 | CTA disabled 且 `importTargetJob` 调用为零；创建并形成可读证据后仍须回 Home 显式选择；不存在无简历/JD-only 导入、训练、报告降级或历史缺绑 fallback 承诺 | 001 Phase 24 |
| C-16 | 规划详情参考构图 | ready Workspace detail 有真实 TargetJob、简历和动态轮次 | 在 desktop/mobile 查看并操作 Header 与四层信息卡 | 约 1250px 内容列中 Header 左侧为步骤/标题/简历/说明，右侧为 Start/Reports；基本信息、要求、隐性关注点与动态轮次形成四层响应式卡面，无横向溢出且不改变请求、route、progress 或 fail-closed 行为 | 001 Phase 26 |

## 8 关联计划

- [001-home-jd-import-and-parse](./plans/001-home-jd-import-and-parse/plan.md) — Home + Parse command progress + Workspace unified plan detail 当前 owner；JD import/parse 的 BDD 只保留 Given/When/Then，当前无真实 E2E owner。


## 9 关联文档

- 上游 spec：[`product-scope`](../product-scope/spec.md)、[`engineering-roadmap`](../engineering-roadmap/spec.md)、[`frontend-shell`](../frontend-shell/spec.md)、[`frontend-workspace-and-practice`](../frontend-workspace-and-practice/spec.md)、[`openapi-v1-contract`](../openapi-v1-contract/spec.md)、[`mock-contract-suite`](../mock-contract-suite/spec.md)
- UI 设计文档：`frontend/src`、[`docs/ui-design/jd-resume-management.md`](../../ui-design/jd-resume-management.md)、[`docs/ui-design/ui-architecture.md`](../../ui-design/ui-architecture.md)、[`docs/ui-design/module-job-workspace.md`](../../ui-design/module-job-workspace.md)、[`docs/ui-design/module-map.md`](../../ui-design/module-map.md)
- 当前正式前端入口：`frontend/src/app/screens/home/`、`frontend/src/app/screens/parse/`、`frontend/src/app/navigation/interviewContext.ts`、`frontend/src/api/generated/`
- 变更记录：[history.md](./history.md)

## 10 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 2.30 | 2026-07-19 | Reopen Phase 26 to align the Workspace plan-detail header and four-layer card composition with the supplied reference while preserving TargetJob behavior. |
| 2.29 | 2026-07-19 | Reopen Phase 25 to align the formal Home hierarchy, intake card and recent records with the supplied desktop reference while preserving runtime limits and all business contracts. |
| 2.28 | 2026-07-15 | Lock a selectable Resume as a permanent prerequisite for import, practice and reports; preserve readable-evidence selection, remove resume-less fallback commitments, and fail closed on invalid historical bindings. |
| 2.27 | 2026-07-15 | Move the bound-resume viewer beside the plan title, remove the standalone launch/binding block, and place Start plus Reports in one leading action row. |
| 2.26 | 2026-07-14 | Add RuntimeConfig-backed 96KiB JD UTF-8 boundary to Home without changing the paste-only UI structure. |
| 2.25 | 2026-07-14 | Add Workspace detail round-assumption done/current/pending visual states derived from the same backend practice-progress projection as list-card rails. |
| 2.24 | 2026-07-14 | Make Parse command-progress-only, replace ready targets into Workspace detail, route ready cards directly, move ready detail/report/start language to Workspace, and require exact initial GET counts. |
