# E2E.P0.018 Workspace Default Render

> **场景 ID**: E2E.P0.018
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **状态**: Ready

## 1 Given

用户已登录，workspace fixture 数据就绪：`getTargetJob=with-rounds`、`getResume=default`、`getPracticePlan=default(ready)`。`InterviewContext` 通过 route param 传入了 `targetJobId/jdId/resumeId/roundId`。

## 2 When

进入 workspace route，渲染 WorkspaceScreen；打开 Plan Switcher Modal；打开 Resume Picker Modal；切换 zh/en、dark+customAccent。

## 3 Then

- workspace plan eyebrow、header summary、Interview Launcher、Main Left/Right 渲染
- Plan Switcher Modal 通过 `listTargetJobs` 拉取数据；Resume Picker Modal 通过 flat `listResumes` 渲染 active list
- 两 Modal 支持 ESC/遮罩/X 关闭、focus trap
- zh/en 切换重绘、dark/customAccent 可见变化
- 非当前 prototype testid 0 命中

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```
