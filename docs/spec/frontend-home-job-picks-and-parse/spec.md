# Frontend Home / Parse Spec

> **版本**: 2.21
> **状态**: active
> **更新日期**: 2026-07-13

## 1 背景与目标

`frontend-home-job-picks-and-parse` 是当前新建模拟面试入口的前端 owner，负责 `home` 与 `parse` 两个屏幕，并拥有由原 `JD 解析结果` 页演进而来的统一面试规划详情母版。它承接 `frontend-shell` 的 App 壳、route normalization、auth continuation、runtime config、generated client 与 fixture-backed transport，把用户从“带着 JD 来”推进到“在面试规划详情 / 面试上下文确认页核对 JD、已绑定简历和轮次，并直接开始面试”。

当前目标链路：

```text
Home 粘贴 JD
  -> 显式选择 ready Resume
  -> Parse loading
  -> 面试规划详情 / 面试上下文确认
  -> 立即面试
  -> Practice handoff
```

本 subspec 维护 Home + Parse loading + 统一详情母版。Workspace 列表、workspace route auto-start、Practice、Report、Resume 管理、TargetJob 后端、Upload 后端、AI 解析与 persistence 分属各自 owner；workspace 带上下文回访时复用本母版，不再维护第二套详情视觉。

## 2 范围

### 2.1 In Scope

- Home 屏（`route=home`）：
  - 源级复刻 `ui-design/src/screen-home.jsx::HomeScreen` 当前结构。
  - Hero 只保留 label + title。
  - JD 输入卡只承载 textarea；不展示或挂载其他 JD intake 控件、弹窗或隐藏分支。
  - `listResumes` 读取 ready 且未归档的简历；用户必须显式选择一份简历后才能点击「立即面试」。
  - `还没有简历？1 分钟创建 ->` 与下拉框同行，点击进入 `resume_versions?flow=create`。
  - `listTargetJobs` 渲染最近 3 张 mock interview card；超过 3 条时展示「更多」并跳转 `workspace`；最近卡片使用固定最大列宽并与 workspace 面试列表共享卡片主体、mini round rail 和 `立即面试 / Start interview now` 主按钮，但不展示删除按钮。
  - Empty state 引导继续创建模拟面试，不展示示例业务数据。
  - 未登录 import 的 `pendingAction` 只携带 `opaquePendingImportId`；`rawText`、`targetLanguage`、`resumeId` 与同一次 import 的 idempotency key 只存在于当前进程的一次性内存 vault，不进入 route 或任何浏览器持久化介质。
  - i18n 支持 zh/en，所有文案通过 typed locale helper。
- Parse / Unified Plan Detail（`route=parse` 首次导入；`route=workspace` 带上下文回访时复用同一母版）：
  - 源级复刻 `ui-design/src/screens-p0-complete.jsx::ParseScreen` 当前结构。
  - Loading 阶段只渲染 4 步进度与面向用户的等待说明；不得展示 model/provider、rubric/prompt/version/hash、provenance、典型耗时等内部调试或实现元数据。
  - 通过 generated `getTargetJob(targetJobId)` 轮询 `analysisStatus`，进入 preview 或 failed state。
  - Preview 阶段用户可见名称为“面试规划详情 / 面试上下文确认”，只读展示 Basic fields、requirements evidence、hidden signals、round assumptions 和已绑定 ready 简历。
  - Preview 成功态不提供字段编辑、requirements toggle、hidden signal 移除、重新解析、保存规划、取消或更换简历入口；解析成功即表示规划已保存，若用户想换 JD/简历，必须回到 Home 创建新规划。
  - Footer actions 只保留「立即面试」，并携带真实 `targetJobId`、`resumeId`、可选 `currentPracticePlanId` 和 `roundId` 进入 practice handoff。
  - 未登录启动通过 auth continuation 接续到 practice。
- Parity 与验证：
  - Home / Parse 必须通过 Vitest + jsdom DOM 锚点、generated-client request、privacy checks、Playwright desktop/mobile pixel parity 与 BDD `E2E.P0.014` / `E2E.P0.015` / `E2E.P0.016`。

### 2.2 Out of Scope

- `workspace` 无上下文规划列表、workspace 列表删除按钮和 workspace 列表启动实践编排：由 `frontend-workspace-and-practice` 承接。
- JD 文件、岗位链接与结构化表单导入：不属于当前产品范围；对应 UI、OpenAPI discriminator、generated artifacts、backend 分支、专属 fixture 与场景必须删除。Resume 模块自己的文件上传 / 粘贴创建继续由 Resume / Upload owner 承接。
- 独立于 Parse 母版之外的第二套 workspace 当前规划详情页：不属于当前范围，必须删除或改为复用统一详情母版。
- `practice` / `report` / `resume_versions` 业务屏：由各自 subspec 承接。
- TargetJob import / parse / update handler、AI provider、prompt/rubric、DB migration、event/outbox：由 backend / contract owner 承接。
- 前端不直接调用 LLM、provider-specific endpoint、prompt registry 或 ad hoc parse fetch。

## 3 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | Owner 范围 | 本 subspec 只接管 `home / parse` 两个 route 的业务内容 | 避免把 Workspace / Practice / Report / Resume 管理混入本 owner |
| D-2 | UI 真理源 | `ui-design/src/screen-home.jsx`、`ui-design/src/screens-p0-complete.jsx::ParseScreen`、`ui-design/src/primitives.jsx` | 正式前端必须源级复刻，不做二次设计 |
| D-3 | Home 提交流程 | 用户在唯一 textarea 粘贴 JD，并显式选择 ready 简历后提交 | 成功 route params 只包含 `targetJobId` 与真实 `resumeId` |
| D-4 | Parse handoff | Parse preview 是只读上下文收据；解析成功即已保存规划，唯一成功 CTA 是「立即面试」，直接使用已绑定上下文进入 practice handoff，不再先 PATCH `updateTargetJob` 或经由 `workspace(autoStartPractice=1)` | Practice session 创建使用已保存 TargetJob / Resume / Round 快照 |
| D-5 | Parse 状态机 | `getTargetJob.analysisStatus` 驱动 loading / preview / failed，不由前端推断解析结果 | AI 与 parsing 结果只来自 backend/API |
| D-6 | Recent mocks | Home 最多展示 3 张最近模拟面试卡片，更多列表入口交给 `workspace` | 首页保持新建任务优先 |
| D-7 | i18n | 只维护当前 `home.*` 与 `parse.*` namespace | 与 typed locale helper 一致 |
| D-8 | Privacy / auth continuation | JD 原文不进入 URL/localStorage/sessionStorage/IndexedDB/console/telemetry；`pendingAction` 的唯一字段是 `opaquePendingImportId` | vault entry 仅在当前进程内保存 `{ rawText, targetLanguage, resumeId, idempotencyKey, expiresAt }` 并原子 consume 一次；refresh / 进程重启、过期或重复 consume 均 fail closed，返回 Home 显示本地化重新粘贴/选择提示，不发起 import，也不尝试从 route 或 storage 恢复原文 |
| D-9 | 统一详情母版 | 原 `JD 解析结果` 页改名为“面试规划详情 / 面试上下文确认”，同时服务首次导入和 workspace 回访 | 用户只学习一个确认页面；workspace 不再维护第二套全页确认 |
| D-10 | 结构化轮次数据源 | 所有 TargetJob 关联的轮次展示与导航上下文使用 `TargetJob.summary.interviewRounds[]`；数组长度必须为 2~5，轮次类型、标题、时长和 focus 均由后端 LLM 结合 JD、岗位级别、公司/行业性质、团队/业务上下文与招聘流程线索推断并持久化 | 避免 Parse、Home 最近卡片、Workspace 回访或共享上下文保留固定 4 轮 / 固定 HR/技术/经理面 / 固定时长模板 |
| D-11 | Recent card fixed grid and shared body | Home 最近模拟面试卡片使用固定最大列宽，并与 workspace 面试列表共用 `MockInterviewCard` 主体；Home 复用 `立即面试` 主按钮但不展示删除按钮 | 保证 Home recent 与 Interview list 不再表现为两套不同卡片规格 |
| D-12 | Recent card planning and start actions | Home 最近模拟面试卡片点击主体进入统一规划详情，`立即面试 / Start interview now` 主按钮直接使用 generated practice handoff 启动 PracticeSession；删除按钮只属于 workspace 面试列表 | 保留继续规划和快速启动两个明确动作，避免 Home recent 与 Interview list 行为分叉 |
| D-13 | Parse loading 信息层级 | loading 只说明当前进度与等待状态，不暴露 model/provider、rubric/prompt/version/hash、provenance 或典型耗时 | 内部诊断信息留在受控日志/观测面，不进入用户界面 |
| D-14 | JD intake 单一合同 | Home 与 `importTargetJob` 只保留 `{ rawText, targetLanguage, resumeId }` | 不保留 source discriminator；删除其他 JD 导入形态但不影响 Resume 上传 |

## 4 设计约束

- DOM 构图、控件类型、间距、字体层级、状态、响应式行为和交互节奏必须可追溯到 `ui-design/` 当前源码。
- Home `home-jd-input-card` 只承载 `home-jd-textarea`；`home-resume-row` 与 `home-submit-row` 位于输入卡下方。旧 source controls、trigger 和 modal 锚点必须为零。
- Home resume select 使用紧凑下拉框；不得平铺所有简历。
- 未登录提交时先创建不可逆推原文的 `opaquePendingImportId`，再把 exact import intent 写入一次性内存 vault；认证路由的 `pendingAction` 不得复制 `rawText`、`targetLanguage`、`resumeId`、intake source 或业务 route params。登录成功后必须原子 consume 一次并使用 vault 中原 idempotency key 提交 exact request；成功、失败、过期和重复 consume 后均不得让同一 entry 再次可读。
- refresh / 进程重启导致 vault 丢失、entry 过期或 ID 已消费时，auth continuation 不调用 `importTargetJob`，清除无效 pending action，返回 Home 并以 zh/en 可访问提示要求用户重新粘贴 JD、选择简历；不得用 `localStorage`、`sessionStorage`、IndexedDB、URL、日志或 telemetry 延长 raw JD 生命周期。
- `route=parse` loading 即使首个 `getTargetJob` 已 ready，也必须先展示当前 UI 真理源定义的 loading gate，再进入 detail；`route=workspace` 回访已解析规划时不得强制播放 parse loading，应直接渲染同一详情母版的 ready 状态。
- Parse loading 的 DOM、截图和文案负向 gate 必须拒绝 `model`、`provider`、`rubric`、`prompt@`、版本/hash、`provenance`、`typical` 等内部实现标记；不能以折叠、弱化颜色或移动到底部代替删除。
- Parse requirements evidence 只读展示 API 返回的 `evidenceLevel`；前端不得在详情页维护临时 hit toggle 或把确认状态写回后端。
- Parse round assumptions 的卡片布局仍追溯 UI 真理源，但卡片数量必须来自 2~5 条 `TargetJob.summary.interviewRounds[]`；R 序号、标题、轮次类型、时长和 focus 也必须来自该数组。这些轮次由后端 LLM 根据 JD、行业/公司性质、岗位级别、团队/业务上下文和招聘流程线索推断，前端不得用 locale 或本地常量补齐轮数、HR/技术/经理面类型或分钟数。
- Home 最近模拟面试卡片与 Workspace plan handoff 不得维护独立 `DEFAULT_ROUNDS`、固定 4 轮、静态 `roundName` 或静态 duration 分支；相关显示或 route params 必须通过同一个 TargetJob structured round mapper 派生。
- `importTargetJob` 是 side-effect operation，必须携带 `Idempotency-Key`；其请求体严格等于 `{ rawText, targetLanguage, resumeId }`，不得带 `source` 或其他 intake-only 字段。
- Parse success detail 不调用 `updateTargetJob`；规划上下文来自已保存的 TargetJob + Resume binding 快照。
- Dark / customAccent 必须在 Home 与 Parse 生效；移动端不得横向溢出。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| Home / Parse / Unified Plan Detail UI | `frontend-home-job-picks-and-parse` | React 组件、route 业务内容、i18n、source parity、pixel parity；workspace 回访复用该详情母版 |
| App shell / auth / runtime | `frontend-shell` | TopBar、route normalization、auth continuation、generated client bootstrap |
| TargetJobs API | `openapi-v1-contract` + `backend-targetjob` | `importTargetJob` / `listTargetJobs` / `getTargetJob` schema、fixtures、handler；`updateTargetJob` 仍属后端 TargetJobs 合同但不是 Parse preview 成功态 consumer |
| Resume upload | `backend-upload` + `backend-resume` | Resume 文件上传与 file object persistence 继续保留；Home JD intake 不消费该能力 |
| Resume list | `backend-resume` | `listResumes` 只读 ready resume selection |
| Practice handoff | `frontend-workspace-and-practice` | PracticePlan / PracticeSession 创建与 practice 跳转；workspace 仅作为列表回访入口，带上下文详情复用统一母版 |
| Mock transport | `mock-contract-suite` | fixture-backed deterministic variants |

## 6 Operation Matrix

| operationId | Fixture | Frontend consumer | Backend handler | Persistence | AI dependency | Scenario |
|-------------|---------|-------------------|-----------------|-------------|---------------|----------|
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json` | Home recent mock interviews | `backend-targetjob` | `target_jobs` / `target_job_requirements` read | none in frontend | `E2E.P0.014` |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` | Home resume select + Parse bound resume display | `backend-resume` | `resumes` read | none | `E2E.P0.015` / `E2E.P0.016` |
| `importTargetJob` | `openapi/fixtures/TargetJobs/importTargetJob.json`（paste success + current validation/failure variants） | Home submits `{ rawText, targetLanguage, resumeId }` | `backend-targetjob` | `target_jobs` create + saved `resume_id` + parse job；无并行 source-specific persistence | backend-only parse job | `E2E.P0.015` |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` | Parse polling + unified detail readonly preview, including structured `summary.interviewRounds[]` | `backend-targetjob` | `target_jobs.summary` / requirements read | backend-generated `target.import.parse` structured rounds | `E2E.P0.015` / `E2E.P0.016` / `E2E.P0.018` |
| `createPracticePlan` / `getPracticePlan` / `startPracticeSession` | `openapi/fixtures/PracticePlans/*`, `openapi/fixtures/PracticeSessions/*` | Parse readonly detail Start action and Home recent quick start | `backend-practice` | `practice_plans` / `practice_sessions` create/read | none | `E2E.P0.016` |

## 7 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Home 默认渲染 | 用户进入 App | 打开 `home` | Hero、唯一 JD textarea、resume select、create resume CTA、recent mocks/empty state 正常渲染；旧 source controls / trigger / modal 锚点不存在；TopBar 高亮首页 | 001 |
| C-2 | Home resume gate | `listResumes` 返回 ready 简历 | 用户尚未选择简历 | 「立即面试」disabled，不调用 import；选择 ready 简历后才允许提交 | 001 |
| C-3 | Paste JD import | 用户选择 ready 简历并粘贴 JD | 点击「立即面试」 | 调用 `importTargetJob({ rawText, targetLanguage, resumeId })` 并携带 `Idempotency-Key`；成功进入 `parse`，route 只携带 `targetJobId` 与真实 `resumeId` | 001 |
| C-4 | 非当前 JD intake 零残留 | paste-only 合同已生效 | 扫描 UI 真理源、formal frontend、OpenAPI/generated、backend、active fixtures/scenarios | 不存在平行 JD intake UI、source discriminator、专属 handler/persistence/job/scenario；Resume 上传路径仍通过原 owner gate | 001 |
| C-5 | Recent mocks | `listTargetJobs` 返回多条记录 | Home 加载完成 | 只展示最近 3 张，排序按 `updatedAt desc`；卡片固定最大列宽并展示 structured mini round rail；卡片主体点击进入统一规划详情，`立即面试` 直接启动 practice；不展示删除按钮；「更多」进入 `workspace`，workspace 列表复用同一卡片主体与动作模型 | 001 |
| C-6 | Parse ready flow | `getTargetJob` 返回 ready 且 `summary.interviewRounds[]` 已由 LLM 根据 JD、行业/公司性质、岗位级别、团队/业务上下文和招聘流程线索生成 2~5 轮 | 用户进入 `parse` | 先展示只含进度与等待说明、无内部 model/rubric/prompt/provenance/typical 元数据的 loading gate，再渲染“面试规划详情 / 面试上下文确认”；Basic fields / Hidden signals / requirements / round assumptions / bound resume 只读且只来自 API response 与已绑定 resume；round assumptions 的轮数、类型、标题、时长和 focus 均显示 `summary.interviewRounds[]`，不得退回固定 4 轮模板；验收必须包含 desktop/mobile Playwright 截图附件或 `screenshotBytes=` marker | 001 |
| C-7 | Parse failed flow | `analysisStatus=failed` 或轮询超时 | Parse polling | 渲染失败态、重新解析和返回首页；不伪造 preview 数据 | 001 |
| C-8 | Readonly plan receipt | Preview 已绑定 ready 简历 | 用户查看详情 | 不出现字段编辑、requirements toggle、hidden signal 移除、重新解析、保存规划、取消或更换简历入口；缺少绑定简历时只阻断开始，不提供 picker 兜底 | 001 |
| C-9 | Start interview | Preview 已绑定 ready 简历 | 点击「立即面试」 | 不调用 `updateTargetJob`，直接使用已保存 `targetJobId/resumeId/roundId/currentPracticePlanId` 创建或读取 PracticePlan 并启动 PracticeSession | 001 |
| C-10 | Privacy / auth continuation | 未登录用户提交 JD，随后正常登录、刷新导致 vault 丢失、entry 过期或重复触发 continuation | 检查 pendingAction、vault consume、URL/storage/log/telemetry 与 import 调用 | `pendingAction` 只含 `opaquePendingImportId`；raw JD 只在一次性内存 vault。正常登录原子 consume 后 exact request 只提交一次；refresh/expired/duplicate consume 均不发起 import，清除无效 action 并返回 Home 显示本地化重新输入提示；fixture transport 不记录 request body | 001 |
| C-11 | Workspace 回访统一详情 | `listTargetJobs` 返回已保存规划且有 `targetJobId/resumeId` | 用户从 `workspace` 规划列表打开规划 | 页面渲染同一个面试规划详情母版，不出现独立 workspace Header/Launcher/JD card 二次确认；返回动作回到面试规划列表 | 001 / frontend-workspace-and-practice 001 |

## 8 关联计划

- [001-home-jd-import-and-parse](./plans/001-home-jd-import-and-parse/plan.md) — Home + Parse + unified plan detail 当前 owner 计划，覆盖 paste-only UI/API contract、generated-client request、real-mode gate、resume selection、recent mocks、parse/workspace unified detail readonly handoff 和 P0.014-P0.016/P0.018 BDD。

## 9 关联文档

- 上游 spec：[`product-scope`](../product-scope/spec.md)、[`engineering-roadmap`](../engineering-roadmap/spec.md)、[`frontend-shell`](../frontend-shell/spec.md)、[`frontend-workspace-and-practice`](../frontend-workspace-and-practice/spec.md)、[`openapi-v1-contract`](../openapi-v1-contract/spec.md)、[`mock-contract-suite`](../mock-contract-suite/spec.md)
- UI 真理源：`ui-design/src/screen-home.jsx`、`ui-design/src/screens-p0-complete.jsx::ParseScreen`、`ui-design/src/primitives.jsx`、[`docs/ui-design/jd-resume-management.md`](../../ui-design/jd-resume-management.md)、[`docs/ui-design/ui-architecture.md`](../../ui-design/ui-architecture.md)、[`docs/ui-design/module-job-workspace.md`](../../ui-design/module-job-workspace.md)、[`docs/ui-design/module-map.md`](../../ui-design/module-map.md)
- 当前正式前端入口：`frontend/src/app/screens/home/`、`frontend/src/app/screens/parse/`、`frontend/src/app/navigation/interviewContext.ts`、`frontend/src/api/generated/`
- 场景：`test/scenarios/e2e/p0-014-home-default-render/`、`test/scenarios/e2e/p0-015-jd-import-and-parse/`、`test/scenarios/e2e/p0-016-parse-confirm-to-workspace/`
- 变更记录：[history.md](./history.md)
