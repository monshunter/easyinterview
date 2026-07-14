# 001 BDD Plan

> **版本**: 2.17
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 场景 | 关联 Phase | 关联 Spec C-* | 主 checklist gate |
|---------|------|-----------|--------------|-------------------|
| E2E.P0.014 | Home paste-only 默认渲染与最近模拟面试 | Phase 1 + 18 | C-1, C-2, C-5, C-10 | 1.4, 18.7 |
| E2E.P0.015 | Home paste JD 到 Parse preview | Phase 1 + 2 + 17 + 18 | C-2, C-3, C-4, C-6, C-7, C-10 | 1.5, 17.4, 18.7 |
| E2E.P0.016 | 面试规划详情只读收据、结构化轮次、报告入口与 Start handoff | Phase 2 + 5 + 6 + 8 + 19 | C-6, C-8, C-9, C-10, C-12 | 2.5, 5.4, 6.4, 8.5, 19.6 |
| E2E.P0.018 | Workspace 列表回访统一面试规划详情 | Phase 5 | C-11 | 5.4 |

## 2 场景定义

| 场景 ID | Given | When | Then | 验证入口 |
|---------|-------|------|------|----------|
| E2E.P0.014 | Home focused tests 使用 fixture-backed runtime；`listTargetJobs` 提供 default / empty / one-job / twelve-plus，`listResumes` 提供当前简历变体；real-mode generated-client test 使用 stub fetch | 执行 generated-client routing、Home focused tests 与 1440×900 / 390×844 browser gate | Home intake 只渲染 textarea、ready Resume select 和 CTA；旧 source controls/modal/trigger 不存在；英文 i18n、ready filter、empty/one/twelve-plus、sort/3-card cap、More、card-detail、quick-start 及 DOM/computed-style/bbox/viewport parity 通过 | `test/scenarios/e2e/p0-014-home-default-render/` |
| E2E.P0.015 | Home 有 ready 简历；TargetJobs / Resumes fixtures 可用；未登录分支使用一次性进程内 vault | 用户选择 ready 简历、粘贴 JD 并提交；另验证登录成功、refresh/lost vault、expired 与 duplicate consume | 未选简历或 JD 为空时不提交；`importTargetJob` exact body 为 `{ rawText, targetLanguage, resumeId }` 且携带 idempotency key，不含 source discriminator；`pendingAction` 只含 `opaquePendingImportId`，正常登录原子 consume 后只重放一次，refresh/expired/duplicate 均零 import 并返回 Home 提示重新输入；成功路由只携带真实 `targetJobId + resumeId`；Parse loading/preview、4xx/failed/privacy path 通过；1440×900 / 390×844 截图与 active source 均无旧 intake UI，也无 model/provider、rubric/prompt/version/hash、provenance/typical latency | `test/scenarios/e2e/p0-015-jd-import-and-parse/` |
| E2E.P0.016 | 用户进入面试规划详情；TargetJob 已保存真实 `resumeId` 与 backend-generated 2~5 条 `summary.interviewRounds[]` | 用户打开只读详情、点击内容区右上角“面试报告”，并继续验证立即面试 | 既有 readonly/round/Start 断言保持；入口精确进入 `/reports?targetJobId=...`，全局 TopBar 无报告入口；Parse 不渲染报告列表、不调用 `listTargetJobReports`、不接受 `section=reports`；1440×900 / 390×844 入口 parity 通过 | `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/` |
| E2E.P0.018 | 用户从 `workspace` 无上下文面试规划列表打开已有规划；TargetJob fixture 提供 `resumeId`，可选 `currentPracticePlanId` | 点击规划卡片进入 `parse` 只读详情 | 页面渲染与 Parse ready state 同源的“面试规划详情 / 面试上下文确认”母版；不出现独立 workspace Header/Launcher/JD card 二次确认；workspace 不执行 `autoStartPractice` 或 session 启动副作用 | `test/scenarios/e2e/p0-018-workspace-default-render/` |

## 3 Real-Mode Overlay

P0.014-P0.016 的 trigger 均先运行 `frontend/src/api/targetJob.realApiMode.test.ts`。该 gate 在 `VITE_EI_API_MODE=real` 与 backend base URL 配置下通过 stub fetch 证明 `listTargetJobs`、`listResumes`、`importTargetJob`、`getTargetJob`、`updateTargetJob` 的 generated-client URL、cookie credentials 和 side-effect idempotency key；P0.015 额外证明 `importTargetJob` 使用 exact flattened wire `{ rawText, targetLanguage, resumeId }` 且不存在 source discriminator。它不发起 live backend 请求。P0.016 额外通过 UI / generated-client spy 证明 Parse success detail 不消费 `updateTargetJob`，并通过 fixture-backed TargetJob 证明 Home recent rail、Parse round assumptions 和 shared navigation context 读取 backend-generated 2~5 条 `summary.interviewRounds[]`，不读取静态 locale focus / local default rounds / fixed duration。各场景只声明自身 trigger/verify 实际执行的 UI、privacy、browser 或 screenshot marker 证据。

## 4 Internal cleanup substitute gate

Phase 12 changes no user-visible behavior and adds no BDD scenario. Source negative, focused Home auth tests and frontend typecheck replace a new scenario; E2E.P0.014-P0.018 remain unchanged.

## 5 Phase 18 current-scope gate

Phase 18 原地修订 P0.014/P0.015，不创建同主题 sibling 场景。URL 专属 `E2E.P0.011` 实体目录与 active INDEX 行必须删除且编号不复用。active zero-reference gate 覆盖 UI truth、owner docs、OpenAPI/generated、frontend Home、backend TargetJob 与 active scenarios；排除 work-journal/bug/report 等合法历史证据，并将 Resume upload 明确列为允许保留的独立业务能力。

## 6 Phase 19 current-scope gate

Phase 19 原地扩展 P0.016，不新建 reports sibling scenario。场景只验证 Parse 责任：同一可信 TargetJob 的内容区右上入口、精确 reports handoff、TopBar negative、零嵌入列表、零 `listTargetJobReports` 请求、零 `section=reports` 兼容，以及 desktop/mobile parity。独立列表的 canonical join、current/latest、空/加载/失败、隔离和返回行为移交 report owner P0.058/P0.059。
