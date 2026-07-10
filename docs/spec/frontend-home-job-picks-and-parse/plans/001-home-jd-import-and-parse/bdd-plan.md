# 001 BDD Plan

> **版本**: 2.13
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 场景 | 关联 Phase | 关联 Spec C-* | 主 checklist gate |
|---------|------|-----------|--------------|-------------------|
| E2E.P0.014 | Home 默认渲染与最近模拟面试 | Phase 1 | C-1, C-2, C-5, C-10 | 1.4 |
| E2E.P0.015 | Home paste/upload/URL import 到 Parse preview | Phase 1 + 2 | C-2, C-3, C-4, C-6, C-7, C-10 | 1.5 |
| E2E.P0.016 | 面试规划详情只读收据、结构化轮次与 Start handoff | Phase 2 + 5 + 6 + 8 | C-6, C-8, C-9, C-10 | 2.5, 5.4, 6.4, 8.5 |
| E2E.P0.018 | Workspace 列表回访统一面试规划详情 | Phase 5 | C-11 | 5.4 |

## 2 场景定义

| 场景 ID | Given | When | Then | 验证入口 |
|---------|-------|------|------|----------|
| E2E.P0.014 | Home focused tests 使用 fixture-backed runtime；`listTargetJobs` 提供 default / empty / one-job / twelve-plus，`listResumes` 提供当前简历变体；real-mode generated-client test 使用 stub fetch | 执行 generated-client routing test 与五个 Home Vitest 文件 | Home shell/control、source/resume/submit layout、英文 i18n、ready filter、empty/one/twelve-plus、sort/3-card cap、More、card-detail 与 quick-start 断言通过；不把 TopBar/theme/mobile/browser/live-backend 写成该 runner 的证据 | `test/scenarios/e2e/p0-014-home-default-render/` |
| E2E.P0.015 | 用户已登录，Home 有 ready 简历；TargetJobs / Uploads / Resumes fixtures 可用 | 用户先选择 ready 简历，再分别 paste JD、upload JD file、URL import | 未选简历时不提交；三种 source 均通过 generated client；side-effect 调用携带 idempotency key；成功进入 Parse 且 params 携带真实 `resumeId`；Parse 先 loading 再 preview；failed/4xx/ privacy path 通过 | `test/scenarios/e2e/p0-015-jd-import-and-parse/` |
| E2E.P0.016 | 用户进入面试规划详情；TargetJob 已保存 Home 选择的真实 `resumeId` 与 backend-generated 2~5 条 `summary.interviewRounds[]`；`listResumes` 与 practice handoff fixtures 可用 | 用户查看 Home 最近卡片、打开只读详情并点击立即面试 | 有效 `resumeId` 被继承并只读展示；Home recent rail、Parse round assumptions 和 shared navigation context 使用 saved `summary.interviewRounds[]`；轮次数量为 2~5，轮次类型/标题、时长和 focus 均来自 fixture/API，不显示固定 4 轮模板；readonly detail 浏览器验收必须附加截图或输出 `screenshotBytes=` marker；缺失 resume 时 Start disabled 且不提供 picker 兜底；字段编辑、requirements toggle、hidden remove、success Re-parse、Save plan、Cancel 和更换简历均不存在；Start 不调用 `updateTargetJob`，直接进入 practice 链路；auth continuation 与 privacy checks 通过 | `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/` |
| E2E.P0.018 | 用户从 `workspace` 无上下文面试规划列表打开已有规划；TargetJob fixture 提供 `resumeId`，可选 `currentPracticePlanId` | 点击规划卡片进入 `parse` 只读详情 | 页面渲染与 Parse ready state 同源的“面试规划详情 / 面试上下文确认”母版；不出现独立 workspace Header/Launcher/JD card 二次确认；workspace 不执行 `autoStartPractice` 或 session 启动副作用 | `test/scenarios/e2e/p0-018-workspace-default-render/` |

## 3 Real-Mode Overlay

P0.014-P0.016 的 trigger 均先运行 `frontend/src/api/targetJob.realApiMode.test.ts`。该 gate 在 `VITE_EI_API_MODE=real` 与 backend base URL 配置下通过 stub fetch 证明 `listTargetJobs`、`listResumes`、`createUploadPresign`、`importTargetJob`、`getTargetJob`、`updateTargetJob` 的 generated-client URL、cookie credentials 和 side-effect idempotency key；它不发起 live backend 请求。P0.016 额外通过 UI / generated-client spy 证明 Parse success detail 不消费 `updateTargetJob`，并通过 fixture-backed TargetJob 证明 Home recent rail、Parse round assumptions 和 shared navigation context 读取 backend-generated 2~5 条 `summary.interviewRounds[]`，不读取静态 locale focus / local default rounds / fixed duration。各场景只声明自身 trigger/verify 实际执行的 UI、privacy、browser 或 screenshot marker 证据。

## 4 Internal cleanup substitute gate

Phase 12 changes no user-visible behavior and adds no BDD scenario. Source negative, focused Home auth tests and frontend typecheck replace a new scenario; E2E.P0.014-P0.018 remain unchanged.
