# 001 BDD Plan

> **版本**: 2.6
> **状态**: completed
> **更新日期**: 2026-07-09

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 场景 | 关联 Phase | 关联 Spec C-* | 主 checklist gate |
|---------|------|-----------|--------------|-------------------|
| E2E.P0.014 | Home 默认渲染与最近模拟面试 | Phase 1 | C-1, C-2, C-5, C-10 | 1.4 |
| E2E.P0.015 | Home paste/upload/URL import 到 Parse preview | Phase 1 + 2 | C-2, C-3, C-4, C-6, C-7, C-10 | 1.5 |
| E2E.P0.016 | 面试规划详情编辑、简历绑定、Save/Start handoff | Phase 2 + 5 | C-6, C-8, C-9, C-10 | 2.5, 5.4 |
| E2E.P0.018 | Workspace 列表回访统一面试规划详情 | Phase 5 | C-11 | 5.4 |

## 2 场景定义

| 场景 ID | Given | When | Then | 验证入口 |
|---------|-------|------|------|----------|
| E2E.P0.014 | 用户打开 Home；`listTargetJobs` fixture 提供 empty / one-job / twelve-plus 变体；`listResumes` fixture 提供 ready 简历 | 切换 fixture 变体并加载 `home` | Home 渲染 JD 输入卡、输入卡底部 source controls、ready 简历下拉框、创建简历入口、提交区、empty state、最近 3 张卡片和 More handoff；TopBar 高亮 Home；zh/en、dark/customAccent 和 responsive layout 通过 | `test/scenarios/e2e/p0-014-home-default-render/` |
| E2E.P0.015 | 用户已登录，Home 有 ready 简历；TargetJobs / Uploads / Resumes fixtures 可用 | 用户先选择 ready 简历，再分别 paste JD、upload JD file、URL import | 未选简历时不提交；三种 source 均通过 generated client；side-effect 调用携带 idempotency key；成功进入 Parse 且 params 携带真实 `resumeId`；Parse 先 loading 再 preview；failed/4xx/ privacy path 通过 | `test/scenarios/e2e/p0-015-jd-import-and-parse/` |
| E2E.P0.016 | 用户进入面试规划详情；route 可携带 Home 选择的真实 `resumeId`；`listResumes` 与 `updateTargetJob` fixtures 可用 | 用户编辑字段，点击仅保存规划或立即面试 | 有效 route `resumeId` 被继承；缺失时必须显式选择 ready 简历；`updateTargetJob` body 只含 supplied fields；Save 进入 `workspace` 同一详情上下文；Start 使用 `autoStartPractice=1` 进入 practice 链路；auth continuation 与 privacy checks 通过 | `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/` |
| E2E.P0.018 | 用户从 `workspace` 无上下文面试规划列表打开已有规划；TargetJob fixture 提供 `resumeId`，可选 `currentPracticePlanId` | 点击规划卡片进入带上下文 `workspace` | 页面渲染与 Parse ready state 同源的“面试规划详情 / 面试上下文确认”母版；不出现独立 workspace Header/Launcher/JD card 二次确认；`autoStartPractice=1` 仍由 workspace owner 创建 session | `test/scenarios/e2e/p0-018-workspace-default-render/` |

## 3 Real-Mode Overlay

P0.014-P0.016 的 trigger 均先运行 `frontend/src/api/targetJob.realApiMode.test.ts`。该 gate 在 `VITE_EI_API_MODE=real` 与真实 backend base URL 下证明 `listTargetJobs`、`listResumes`、`createUploadPresign`、`importTargetJob`、`getTargetJob`、`updateTargetJob` 通过 generated client 使用 cookie credentials、正确 base URL 和 side-effect idempotency key。UI 子用例继续使用 fixture-backed transport 以稳定覆盖 DOM、source variants、错误态、privacy 和 responsive parity。
