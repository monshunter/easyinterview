# E2E.P0.018 Interview Plan List and Workspace Readonly Detail

> **场景 ID**: E2E.P0.018
> **执行方式**: automated (Vitest + Playwright)
> **隔离级别**: fixture-backed frontend
> **状态**: Ready

## 1 Given

用户已登录，workspace fixture 数据就绪：`listTargetJobs=default`。无上下文一级 `面试` 入口进入纯规划列表；ready 规划详情由 `/workspace?targetJobId=...` 承接。

## 2 When

点击 TopBar `面试` 进入无上下文 Workspace；选择 ready 规划卡片直接进入 target-scoped Workspace 只读详情；切换 zh/en、dark+customAccent。

## 3 Then

- TopBar 中文标签为 `面试`，英文标签为 `Interview`
- 无上下文 workspace 渲染 `workspace-plan-list-*`，不渲染 out-of-scope `workspace-empty`
- 规划列表项以卡片呈现：具备卡片背景、1px 边框、轻阴影、body/footer 分区、footer `立即面试` 主按钮和右上角删除图标
- 选择 ready 规划卡片只导航到 `/workspace?targetJobId=...`；不携带 `planId/resumeId/jobId/jdId/reportId`，不发起 import/update/poll，不播放 Parse 动画
- 无参 Workspace 只请求一次 `listTargetJobs`；详情只请求一次 `getTargetJob`，不请求 `listTargetJobs` / `listResumes`
- 带 `targetJobId` 的 Workspace 渲染 `unified-plan-detail` / `parse-*` 共享 DOM 锚点及后端绑定的只读简历；`parse-*` 是组件锚点，不代表路由是 Parse
- 轮次假设必须同时呈现 `done/current/pending` 与“已进行/即将进行/未进行”，三态背景和边框必须有区别
- zh/en 切换重绘、dark/customAccent 可见变化
- out-of-scope prototype testid 0 命中

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```
