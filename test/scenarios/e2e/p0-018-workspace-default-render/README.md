# E2E.P0.018 Interview Plan List and Workspace Detail Render

> **场景 ID**: E2E.P0.018
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **状态**: Ready

## 1 Given

用户已登录，workspace fixture 数据就绪：`listTargetJobs=default`、`getTargetJob=default`、`getResume=default`、`getPracticePlan=default(ready)`。无上下文一级 `面试` 入口先进入规划列表；带 `targetJobId/jdId/resumeId/roundId` 的 route param 进入当前规划详情。

## 2 When

点击 TopBar `面试` 进入无上下文 workspace；选择规划卡片进入当前规划详情；打开 Plan Switcher Modal；打开 Resume Picker Modal；切换 zh/en、dark+customAccent。

## 3 Then

- TopBar 中文标签为 `面试`，英文标签为 `Interview`
- 无上下文 workspace 渲染 `workspace-plan-list-*`，不渲染旧 `workspace-empty`
- 规划列表项以卡片呈现：具备卡片背景、1px 边框、轻阴影、body/footer 分区和明确进入按钮
- 选择规划卡片写入 `targetJobId/jobId/jdId/resumeId`；存在 `currentPracticePlanId` 时才携带真实 `planId`，不得伪造 `plan-${targetJobId}`、`resume-unbound` 或 `reportId`
- 带上下文 workspace plan eyebrow、header summary、Interview Launcher、Main Left/Right 渲染
- Plan Switcher Modal 通过 `listTargetJobs` 拉取数据；Resume Picker Modal 通过 flat `listResumes` 渲染 active list
- 两 Modal 支持 ESC/遮罩/X 关闭、focus trap
- zh/en 切换重绘、dark/customAccent 可见变化
- 非当前 prototype testid 0 命中

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```
