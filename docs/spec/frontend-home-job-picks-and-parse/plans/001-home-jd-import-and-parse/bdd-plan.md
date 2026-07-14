# 001 BDD Plan

> **版本**: 2.20
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 场景 | 关联 Phase | 关联 Spec C-* | 主 checklist gate |
|---------|------|-----------|--------------|-------------------|
| E2E.P0.014 | Home paste-only 默认渲染、exact GET 与最近模拟面试 | Phase 1 + 18 + 20 | C-1, C-2, C-5, C-10 | 1.4, 18.7, 20.5 |
| E2E.P0.015 | Home paste JD 到 command-only Parse + 96KiB boundary | Phase 1 + 2 + 17 + 18 + 20 + 22 | C-2, C-3, C-4, C-6, C-7, C-10, C-14 | 1.5, 17.4, 18.7, 20.5, 22.3 |
| E2E.P0.016 | Workspace 面试规划详情只读收据、结构化轮次三态、报告入口与 Start handoff | Phase 2 + 5 + 6 + 8 + 19 + 20 + 21 | C-6, C-8, C-9, C-10, C-12, C-13 | 2.5, 5.4, 6.4, 8.5, 19.6, 20.5, 21.4 |
| E2E.P0.018 | Workspace 列表直达统一面试规划详情 | Phase 5 + 20 | C-11 | 5.4, 20.5 |

## 2 场景定义

| 场景 ID | Given | When | Then | 验证入口 |
|---------|-------|------|------|----------|
| E2E.P0.014 | Home focused tests 使用 fixture-backed runtime；底层 transport 可计数；`listTargetJobs` / `listResumes` 提供当前变体 | 在 React StrictMode 打开 Home 并点击 ready recent card | 两项同 key 初载底层 GET 各恰好 1；无紧邻重复 pair；ready card 直达 `/workspace?targetJobId`，不进入 Parse；原 paste-only/quick-start/More/parity 合同保持 | `test/scenarios/e2e/p0-014-home-default-render/` |
| E2E.P0.015 | Home 有 ready 简历；RuntimeConfig/default 98,304 bytes；import 可经历 queued/processing/ready | 提交 UTF-8 limit 与 limit+1，推进合法 import polling 到 ready，再执行 Back/auth continuation | limit POST 成功并进入 command route；+1 inline reject 且零 POST/vault；每 tick GET=1；ready replace；auth/privacy/idempotency 保持 | `test/scenarios/e2e/p0-015-jd-import-and-parse/` |
| E2E.P0.016 | 用户进入 Workspace 面试规划详情；TargetJob 已保存真实 `resumeId`、backend-generated 2~5 条 rounds 与合法 completed/current progress | 用户打开只读详情、观察轮次三态、点击内容区右上角“面试报告”，并继续验证立即面试 | 轮次卡显示 done/current/pending 对应“已进行/即将进行/未进行”，背景和边框三态不同且与列表 rail 一致；既有 readonly/report/Start/zero-list 断言保持；1440×900 / 390×844 parity 通过 | `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/` |
| E2E.P0.018 | 用户从 query-free workspace 列表打开已有 ready 规划；TargetJob fixture 提供绑定上下文 | 点击规划卡片 | 直接进入 `/workspace?targetJobId`；详情只读复用统一母版，不播放 Parse animation，不 import/poll/auto-start；详情 `getTargetJob` 同 key 初载底层请求恰好 1 | `test/scenarios/e2e/p0-018-workspace-default-render/` |

## 3 Real-Mode Overlay

P0.014-P0.016 的 trigger 均先运行 `frontend/src/api/targetJob.realApiMode.test.ts`。该 gate 在 `VITE_EI_API_MODE=real` 与 backend base URL 配置下通过 stub fetch 证明 `listTargetJobs`、`listResumes`、`importTargetJob`、`getTargetJob`、`updateTargetJob` 的 generated-client URL、cookie credentials 和 side-effect idempotency key；P0.015 额外证明 `importTargetJob` 使用 exact flattened wire `{ rawText, targetLanguage, resumeId }` 且不存在 source discriminator。它不发起 live backend 请求。P0.016 额外通过 UI / generated-client spy 证明 Workspace success detail 不消费 `updateTargetJob`/`listResumes`，并通过 fixture-backed TargetJob 证明 Home recent rail、Workspace round assumptions 和 shared navigation context 读取 backend-generated 2~5 条 `summary.interviewRounds[]`，不读取静态 locale focus / local default rounds / fixed duration。各场景只声明自身 trigger/verify 实际执行的 UI、privacy、browser 或 screenshot marker 证据。

## 4 Internal cleanup substitute gate

Phase 12 changes no user-visible behavior and adds no BDD scenario. Source negative, focused Home auth tests and frontend typecheck replace a new scenario; E2E.P0.014-P0.018 remain unchanged.

## 5 Phase 18 current-scope gate

Phase 18 原地修订 P0.014/P0.015，不创建同主题 sibling 场景。URL 专属 `E2E.P0.011` 实体目录与 active INDEX 行必须删除且编号不复用。active zero-reference gate 覆盖 UI truth、owner docs、OpenAPI/generated、frontend Home、backend TargetJob 与 active scenarios；排除 work-journal/bug/report 等合法历史证据，并将 Resume upload 明确列为允许保留的独立业务能力。

## 6 Phase 19 current-scope gate

Phase 19 原地扩展 P0.016，不新建 reports sibling scenario；Phase 20 将其 ready-detail route 归属 Workspace。当前场景验证同一可信 TargetJob 的 Workspace detail 右上入口、精确 reports handoff、TopBar/Parse-entry negative、零嵌入列表、零 `listTargetJobReports` 请求、零 `section=reports` 兼容，以及 desktop/mobile parity。独立列表的 canonical join、current/latest、空/加载/失败、隔离和返回行为移交 report owner P0.058/P0.059。

## 7 Phase 20 current-scope gate

Phase 20 只原地扩展 P0.014/P0.015/P0.016/P0.018。请求数量以底层 transport request 为准，不以 hook/effect invocation 为准；不得通过关闭 StrictMode 获得 exact-one。Parse polling 的第二次及后续 GET 必须由可验证 timer interval 驱动，ready handoff 必须是 replace，且既有场景 ID 不复用、不新建 sibling。

## 8 Phase 21 current-scope gate

Phase 21 只原地扩展 P0.016。三态只允许由 persisted `practiceProgress` 的 completed prefix / exact current 派生；浏览器 gate 同时读取可见标签、`data-round-state` 与 computed background/border，并对全完成、无效投影和列表 rail 一致性做回归；不得新增 sibling scenario。
