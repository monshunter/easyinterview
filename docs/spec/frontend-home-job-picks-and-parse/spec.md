# Frontend Home / Parse Spec

> **版本**: 2.19
> **状态**: active
> **更新日期**: 2026-07-10

## 1 背景与目标

`frontend-home-job-picks-and-parse` 是当前新建模拟面试入口的前端 owner，负责 `home` 与 `parse` 两个屏幕，并拥有由原 `JD 解析结果` 页演进而来的统一面试规划详情母版。它承接 `frontend-shell` 的 App 壳、route normalization、auth continuation、runtime config、generated client 与 fixture-backed transport，把用户从“带着 JD 来”推进到“在面试规划详情 / 面试上下文确认页核对 JD、已绑定简历和轮次，并直接开始面试”。

当前目标链路：

```text
Home 输入 / 上传 / URL 导入 JD
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
  - JD 输入卡承载 textarea、上传 JD 文件与 URL 导入 source actions。
  - `listResumes` 读取 ready 且未归档的简历；用户必须显式选择一份简历后才能点击「立即面试」。
  - `还没有简历？1 分钟创建 ->` 与下拉框同行，点击进入 `resume_versions?flow=create`。
  - `listTargetJobs` 渲染最近 3 张 mock interview card；超过 3 条时展示「更多」并跳转 `workspace`；最近卡片使用固定最大列宽并与 workspace 面试列表共享卡片主体、mini round rail 和 `立即面试 / Start interview now` 主按钮，但不展示删除按钮。
  - Empty state 引导继续创建模拟面试，不展示示例业务数据。
  - 未登录 import 通过 opaque pending import id 接续。
  - i18n 支持 zh/en，所有文案通过 typed locale helper。
- Parse / Unified Plan Detail（`route=parse` 首次导入；`route=workspace` 带上下文回访时复用同一母版）：
  - 源级复刻 `ui-design/src/screens-p0-complete.jsx::ParseScreen` 当前结构。
  - Loading 阶段渲染 4 步进度条与 backend parse metadata footer。
  - 通过 generated `getTargetJob(targetJobId)` 轮询 `analysisStatus`，进入 preview 或 failed state。
  - Preview 阶段用户可见名称为“面试规划详情 / 面试上下文确认”，只读展示 Basic fields、requirements evidence、hidden signals、round assumptions 和已绑定 ready 简历。
  - Preview 成功态不提供字段编辑、requirements toggle、hidden signal 移除、重新解析、保存规划、取消或更换简历入口；解析成功即表示规划已保存，若用户想换 JD/简历，必须回到 Home 创建新规划。
  - Footer actions 只保留「立即面试」，并携带真实 `targetJobId`、`resumeId`、可选 `currentPracticePlanId` 和 `roundId` 进入 practice handoff。
  - 未登录启动通过 auth continuation 接续到 practice。
- Parity 与验证：
  - Home / Parse 必须通过 Vitest + jsdom DOM 锚点、generated-client request、privacy checks、Playwright desktop/mobile pixel parity 与 BDD `E2E.P0.014` / `E2E.P0.015` / `E2E.P0.016`。

### 2.2 Out of Scope

- `workspace` 无上下文规划列表、workspace 列表删除按钮和 workspace 列表启动实践编排：由 `frontend-workspace-and-practice` 承接。
- 独立于 Parse 母版之外的第二套 workspace 当前规划详情页：不属于当前范围，必须删除或改为复用统一详情母版。
- `practice` / `report` / `resume_versions` 业务屏：由各自 subspec 承接。
- 真实 URL fetch、文件对象持久化、TargetJob import / parse / update handler、AI provider、prompt/rubric、DB migration、event/outbox：由 backend / contract owner 承接。
- 前端不直接调用 LLM、provider-specific endpoint、prompt registry 或 ad hoc parse fetch。

## 3 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | Owner 范围 | 本 subspec 只接管 `home / parse` 两个 route 的业务内容 | 避免把 Workspace / Practice / Report / Resume 管理混入本 owner |
| D-2 | UI 真理源 | `ui-design/src/screen-home.jsx`、`ui-design/src/screens-p0-complete.jsx::ParseScreen`、`ui-design/src/primitives.jsx` | 正式前端必须源级复刻，不做二次设计 |
| D-3 | Home 提交流程 | 用户先显式选择 ready 简历，再提交 paste / upload / URL import | 成功 route params 必须包含真实 `resumeId` |
| D-4 | Parse handoff | Parse preview 是只读上下文收据；解析成功即已保存规划，唯一成功 CTA 是「立即面试」，直接使用已绑定上下文进入 practice handoff，不再先 PATCH `updateTargetJob` 或经由 `workspace(autoStartPractice=1)` | Practice session 创建使用已保存 TargetJob / Resume / Round 快照 |
| D-5 | Parse 状态机 | `getTargetJob.analysisStatus` 驱动 loading / preview / failed，不由前端推断解析结果 | AI 与 parsing 结果只来自 backend/API |
| D-6 | Recent mocks | Home 最多展示 3 张最近模拟面试卡片，更多列表入口交给 `workspace` | 首页保持新建任务优先 |
| D-7 | i18n | 只维护当前 `home.*` 与 `parse.*` namespace | 与 typed locale helper 一致 |
| D-8 | Privacy | JD 原文、source URL、rawDescription 不进入 URL/localStorage/console/telemetry | 只允许通过 generated request body 与 React state 传递 |
| D-9 | 统一详情母版 | 原 `JD 解析结果` 页改名为“面试规划详情 / 面试上下文确认”，同时服务首次导入和 workspace 回访 | 用户只学习一个确认页面；workspace 不再维护第二套全页确认 |
| D-10 | 结构化轮次数据源 | 所有 TargetJob 关联的轮次展示与导航上下文使用 `TargetJob.summary.interviewRounds[]`；数组长度必须为 2~5，轮次类型、标题、时长和 focus 均由后端 LLM 结合 JD、岗位级别、公司/行业性质、团队/业务上下文与招聘流程线索推断并持久化 | 避免 Parse、Home 最近卡片、Workspace 回访或共享上下文保留固定 4 轮 / 固定 HR/技术/经理面 / 固定时长模板 |
| D-11 | Recent card fixed grid and shared body | Home 最近模拟面试卡片使用固定最大列宽，并与 workspace 面试列表共用 `MockInterviewCard` 主体；Home 复用 `立即面试` 主按钮但不展示删除按钮 | 保证 Home recent 与 Interview list 不再表现为两套不同卡片规格 |
| D-12 | Recent card planning and start actions | Home 最近模拟面试卡片点击主体进入统一规划详情，`立即面试 / Start interview now` 主按钮直接使用 generated practice handoff 启动 PracticeSession；删除按钮只属于 workspace 面试列表 | 保留继续规划和快速启动两个明确动作，避免 Home recent 与 Interview list 行为分叉 |

## 4 设计约束

- DOM 构图、控件类型、间距、字体层级、状态、响应式行为和交互节奏必须可追溯到 `ui-design/` 当前源码。
- Home `home-jd-input-card` 同时承载 textarea 与 `home-jd-source-controls`；主按钮位于简历选择行下方。
- Home resume select 使用紧凑下拉框；不得平铺所有简历。
- `route=parse` loading 即使首个 `getTargetJob` 已 ready，也必须先展示当前 UI 真理源定义的 loading gate，再进入 detail；`route=workspace` 回访已解析规划时不得强制播放 parse loading，应直接渲染同一详情母版的 ready 状态。
- Parse requirements evidence 只读展示 API 返回的 `evidenceLevel`；前端不得在详情页维护临时 hit toggle 或把确认状态写回后端。
- Parse round assumptions 的卡片布局仍追溯 UI 真理源，但卡片数量必须来自 2~5 条 `TargetJob.summary.interviewRounds[]`；R 序号、标题、轮次类型、时长和 focus 也必须来自该数组。这些轮次由后端 LLM 根据 JD、行业/公司性质、岗位级别、团队/业务上下文和招聘流程线索推断，前端不得用 locale 或本地常量补齐轮数、HR/技术/经理面类型或分钟数。
- Home 最近模拟面试卡片与 Workspace plan handoff 不得维护独立 `DEFAULT_ROUNDS`、固定 4 轮、静态 `roundName` 或静态 duration 分支；相关显示或 route params 必须通过同一个 TargetJob structured round mapper 派生。
- `createUploadPresign`、`importTargetJob` 均为 side-effect operation，必须携带 `Idempotency-Key`。
- Parse success detail 不调用 `updateTargetJob`；规划上下文来自已保存的 TargetJob + Resume binding 快照。
- Dark / customAccent 必须在 Home 与 Parse 生效；移动端不得横向溢出。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| Home / Parse / Unified Plan Detail UI | `frontend-home-job-picks-and-parse` | React 组件、route 业务内容、i18n、source parity、pixel parity；workspace 回访复用该详情母版 |
| App shell / auth / runtime | `frontend-shell` | TopBar、route normalization、auth continuation、generated client bootstrap |
| TargetJobs API | `openapi-v1-contract` + `backend-targetjob` | `importTargetJob` / `listTargetJobs` / `getTargetJob` schema、fixtures、handler；`updateTargetJob` 仍属后端 TargetJobs 合同但不是 Parse preview 成功态 consumer |
| Upload presign | `backend-upload` | `createUploadPresign` handler 与 file object persistence |
| Resume list | `backend-resume` | `listResumes` 只读 ready resume selection |
| Practice handoff | `frontend-workspace-and-practice` | PracticePlan / PracticeSession 创建与 practice 跳转；workspace 仅作为列表回访入口，带上下文详情复用统一母版 |
| Mock transport | `mock-contract-suite` | fixture-backed deterministic variants |

## 6 Operation Matrix

| operationId | Fixture | Frontend consumer | Backend handler | Persistence | AI dependency | Scenario |
|-------------|---------|-------------------|-----------------|-------------|---------------|----------|
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json` | Home recent mock interviews | `backend-targetjob` | `target_jobs` / `target_job_requirements` read | none in frontend | `E2E.P0.014` |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` | Home resume select + Parse bound resume display | `backend-resume` | `resumes` read | none | `E2E.P0.015` / `E2E.P0.016` |
| `createUploadPresign` | `openapi/fixtures/Uploads/createUploadPresign.json` | Home upload source action | `backend-upload` | `file_objects` create | none | `E2E.P0.015` |
| `importTargetJob` | `openapi/fixtures/TargetJobs/importTargetJob.json` | Home paste / file / URL import | `backend-targetjob` | `target_jobs` / `target_job_sources` create | backend-only parse job | `E2E.P0.015` |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` | Parse polling + unified detail readonly preview, including structured `summary.interviewRounds[]` | `backend-targetjob` | `target_jobs.summary` / requirements read | backend-generated `target.import.parse` structured rounds | `E2E.P0.015` / `E2E.P0.016` / `E2E.P0.018` |
| `createPracticePlan` / `getPracticePlan` / `startPracticeSession` | `openapi/fixtures/PracticePlans/*`, `openapi/fixtures/PracticeSessions/*` | Parse readonly detail Start action and Home recent quick start | `backend-practice` | `practice_plans` / `practice_sessions` create/read | none | `E2E.P0.016` |

## 7 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Home 默认渲染 | 用户进入 App | 打开 `home` | Hero、JD 输入卡、source actions、resume select、create resume CTA、recent mocks/empty state 正常渲染；TopBar 高亮首页 | 001 |
| C-2 | Home resume gate | `listResumes` 返回 ready 简历 | 用户尚未选择简历 | 「立即面试」disabled，不调用 import；选择 ready 简历后才允许提交 | 001 |
| C-3 | Paste JD import | 用户选择 ready 简历并粘贴 JD | 点击「立即面试」 | 调用 `importTargetJob` manual_text，成功进入 `parse` 且 route params 含真实 `resumeId` | 001 |
| C-4 | Upload / URL import | 用户使用 source actions | Confirm | Upload 先 `createUploadPresign` 再 `importTargetJob(file)`；URL 调 `importTargetJob(url)`；均带 `Idempotency-Key` | 001 |
| C-5 | Recent mocks | `listTargetJobs` 返回多条记录 | Home 加载完成 | 只展示最近 3 张，排序按 `updatedAt desc`；卡片固定最大列宽并展示 structured mini round rail；卡片主体点击进入统一规划详情，`立即面试` 直接启动 practice；不展示删除按钮；「更多」进入 `workspace`，workspace 列表复用同一卡片主体与动作模型 | 001 |
| C-6 | Parse ready flow | `getTargetJob` 返回 ready 且 `summary.interviewRounds[]` 已由 LLM 根据 JD、行业/公司性质、岗位级别、团队/业务上下文和招聘流程线索生成 2~5 轮 | 用户进入 `parse` | 先展示 loading gate，再渲染“面试规划详情 / 面试上下文确认”；Basic fields / Hidden signals / requirements / round assumptions / bound resume 只读且只来自 API response 与已绑定 resume；round assumptions 的轮数、类型、标题、时长和 focus 均显示 `summary.interviewRounds[]`，不得退回固定 4 轮模板；验收必须包含 Playwright 截图附件或 `screenshotBytes=` marker | 001 |
| C-7 | Parse failed flow | `analysisStatus=failed` 或轮询超时 | Parse polling | 渲染失败态、重新解析和返回首页；不伪造 preview 数据 | 001 |
| C-8 | Readonly plan receipt | Preview 已绑定 ready 简历 | 用户查看详情 | 不出现字段编辑、requirements toggle、hidden signal 移除、重新解析、保存规划、取消或更换简历入口；缺少绑定简历时只阻断开始，不提供 picker 兜底 | 001 |
| C-9 | Start interview | Preview 已绑定 ready 简历 | 点击「立即面试」 | 不调用 `updateTargetJob`，直接使用已保存 `targetJobId/resumeId/roundId/currentPracticePlanId` 创建或读取 PracticePlan 并启动 PracticeSession | 001 |
| C-10 | Privacy | 用户提交 JD 原文或 URL | 检查 URL/localStorage/console/telemetry | 不出现 raw JD、source URL 或 rawDescription；fixture transport 不记录 request body | 001 |
| C-11 | Workspace 回访统一详情 | `listTargetJobs` 返回已保存规划且有 `targetJobId/resumeId` | 用户从 `workspace` 规划列表打开规划 | 页面渲染同一个面试规划详情母版，不出现独立 workspace Header/Launcher/JD card 二次确认；返回动作回到面试规划列表 | 001 / frontend-workspace-and-practice 001 |

## 8 关联计划

- [001-home-jd-import-and-parse](./plans/001-home-jd-import-and-parse/plan.md) — Home + Parse + unified plan detail 当前 owner 计划，覆盖 source parity、generated-client request、real-mode gate、resume selection、recent mocks、parse/workspace unified detail readonly handoff 和 P0.014-P0.016/P0.018 BDD。

## 9 关联文档

- 上游 spec：[`product-scope`](../product-scope/spec.md)、[`engineering-roadmap`](../engineering-roadmap/spec.md)、[`frontend-shell`](../frontend-shell/spec.md)、[`frontend-workspace-and-practice`](../frontend-workspace-and-practice/spec.md)、[`openapi-v1-contract`](../openapi-v1-contract/spec.md)、[`mock-contract-suite`](../mock-contract-suite/spec.md)
- UI 真理源：`ui-design/src/screen-home.jsx`、`ui-design/src/screens-p0-complete.jsx::ParseScreen`、`ui-design/src/primitives.jsx`、[`docs/ui-design/jd-resume-management.md`](../../ui-design/jd-resume-management.md)、[`docs/ui-design/ui-architecture.md`](../../ui-design/ui-architecture.md)、[`docs/ui-design/module-job-workspace.md`](../../ui-design/module-job-workspace.md)、[`docs/ui-design/module-map.md`](../../ui-design/module-map.md)
- 当前正式前端入口：`frontend/src/app/screens/home/`、`frontend/src/app/screens/parse/`、`frontend/src/app/navigation/interviewContext.ts`、`frontend/src/api/generated/`
- 场景：`test/scenarios/e2e/p0-014-home-default-render/`、`test/scenarios/e2e/p0-015-jd-import-and-parse/`、`test/scenarios/e2e/p0-016-parse-confirm-to-workspace/`
- 变更记录：[history.md](./history.md)
