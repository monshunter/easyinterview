# E2E.P0.018 Interview Plan List and Parse Detail Render

> **场景 ID**: E2E.P0.018
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **状态**: Ready

## 1 Given

用户已登录，workspace fixture 数据就绪：`listTargetJobs=default`。无上下文一级 `面试` 入口进入纯规划列表；规划详情通过 `parse?targetJobId=...` 的统一面试规划详情承接。

## 2 When

点击 TopBar `面试` 进入无上下文 workspace；选择规划卡片进入 parse 统一面试规划详情；切换 zh/en、dark+customAccent。

## 3 Then

- TopBar 中文标签为 `面试`，英文标签为 `Interview`
- 无上下文 workspace 渲染 `workspace-plan-list-*`，不渲染 out-of-scope `workspace-empty`
- 规划列表项以卡片呈现：具备卡片背景、1px 边框、轻阴影、body/footer 分区、footer `立即面试` 主按钮和右上角删除图标
- 选择规划卡片导航到 `parse`，写入 `targetJobId`，并在存在真实绑定时携带 `planId/resumeId`；不得伪造 `jobId`、`jdId`、`plan-${targetJobId}`、`resume-unbound` 或 `reportId`
- 带上下文 workspace 仍渲染列表，不读取 `targetJobId` / `planId`，不渲染 `parse-error` / “缺少目标岗位 ID”
- parse 详情负责 `unified-plan-detail` / `parse-*` 统一详情锚点和简历选择器
- zh/en 切换重绘、dark/customAccent 可见变化
- out-of-scope prototype testid 0 命中

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```
