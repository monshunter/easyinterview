# 001 BDD Plan

> **版本**: 1.14
> **状态**: active
> **更新日期**: 2026-07-09

**关联 Plan**: [plan](./plan.md)
**关联 Checklist**: [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 场景矩阵

| 场景 ID | 名称 | 类别 | 验证入口 |
|---------|------|------|----------|
| `E2E.P0.018` | 面试入口规划列表 + parse 统一面试规划详情 handoff | primary + alternate + UX regression | `test/scenarios/e2e/p0-018-workspace-default-render/` |
| `E2E.P0.019` | Workspace context loading + empty/missing-resume + plan refresh | primary + boundary + recovery | `test/scenarios/e2e/p0-019-workspace-context-loading/` |
| `E2E.P0.020` | 立即面试 + idempotency + auth recovery | primary + alternate + failure | `test/scenarios/e2e/p0-020-workspace-start-practice/` |
| `E2E.P0.021` | Embedded insight + records placeholder + privacy/non-current negative | regression + privacy | `test/scenarios/e2e/p0-021-workspace-handoff/` |

## 2 场景明细

### E2E.P0.018 面试入口规划列表 + 统一面试规划详情

| Given | When | Then |
|-------|------|------|
| 用户已登录；`listTargetJobs=default` 且可能混入 failed / blank-title 历史脏数据；点击顶部 `面试`；另有 legacy workspace route params 含 `targetJobId / jdId / resumeId / roundId`；real-backend 场景存在可归档 ready TargetJob | 进入 workspace 默认 landing；点击一个规划卡片；进入 parse 统一详情；从详情页再次点击 TopBar `面试`；点击另一个卡片右上角的删除图标；刷新 workspace；切换 zh/en、dark/customAccent | TopBar 只显示 `首页 / 面试 / 简历`；workspace canonicalize 为 `/workspace`、始终渲染面试规划列表且不显示缺 JD 死胡同；列表请求 `analysisStatus=ready`，failed / blank-title TargetJob 不渲染为卡片；规划卡片以卡片背景、边框、轻阴影、body/footer 分区、footer `立即面试` 和右上角删除图标呈现，卡片内不展示 `手动输入` / 来源类型 / 目标语言等导入元信息，并通过 `listTargetJobs` 返回的 target job-level `resumeId` 导航 `parse`；删除图标调用 generated `archiveTargetJob`，成功后卡片消失且刷新后不回灌，失败时保留卡片；legacy workspace params 不触发 `getTargetJob` / `parse-error`；有 `currentPracticePlanId` 时携带真实 `planId`，无 plan 时不得伪造；详情 DOM 与 Parse ready state 同源，旧独立 workspace detail anchors 不出现；非当前 prototype testid 0 命中 |

### E2E.P0.019 Workspace context loading

| Given | When | Then |
|-------|------|------|
| 用户已登录；覆盖 TargetJob ready/not-found/5xx、Resume ready/not-found、PracticePlan ready/not-found variants | 加载 workspace route | Ready path hydrate InterviewContext；缺 JD 显示 `WorkspaceEmptyState`；缺简历显示 `WorkspaceMissingResumeState`；absent plan clears planId and waits for start action；target 5xx renders retry state without fake data |

### E2E.P0.020 立即面试

| Given | When | Then |
|-------|------|------|
| 用户已登录或未登录；PracticePlan absent/ready/not-found；`createPracticePlan` and `startPracticeSession` fixtures include success and failure variants | 点击 `立即面试`，必要时重试；未登录时完成 auth recovery | Absent plan path calls `createPracticePlan` then `startPracticeSession` with idempotency keys；ready plan path skips create；failure shows inline retry; pendingAction returns to workspace and auto-starts without leaking sensitive params |

### E2E.P0.021 Embedded insight and records placeholder

| Given | When | Then |
|-------|------|------|
| 用户已登录；workspace data ready；records typed consumer is outside this completed plan | 点击 insight card；点击 records placeholder；运行 runtime negative grep | Insight action remains in workspace and carries safe params only；records placeholder does not navigate to report；runtime does not call standalone insight API, report API, untyped fixture extension or prototype helper；privacy and non-current negative grep pass |

## 3 执行入口

```bash
test/scenarios/e2e/p0-018-workspace-default-render/scripts/setup.sh && test/scenarios/e2e/p0-018-workspace-default-render/scripts/trigger.sh && test/scenarios/e2e/p0-018-workspace-default-render/scripts/verify.sh && test/scenarios/e2e/p0-018-workspace-default-render/scripts/cleanup.sh
test/scenarios/e2e/p0-019-workspace-context-loading/scripts/setup.sh && test/scenarios/e2e/p0-019-workspace-context-loading/scripts/trigger.sh && test/scenarios/e2e/p0-019-workspace-context-loading/scripts/verify.sh && test/scenarios/e2e/p0-019-workspace-context-loading/scripts/cleanup.sh
test/scenarios/e2e/p0-020-workspace-start-practice/scripts/setup.sh && test/scenarios/e2e/p0-020-workspace-start-practice/scripts/trigger.sh && test/scenarios/e2e/p0-020-workspace-start-practice/scripts/verify.sh && test/scenarios/e2e/p0-020-workspace-start-practice/scripts/cleanup.sh
test/scenarios/e2e/p0-021-workspace-handoff/scripts/setup.sh && test/scenarios/e2e/p0-021-workspace-handoff/scripts/trigger.sh && test/scenarios/e2e/p0-021-workspace-handoff/scripts/verify.sh && test/scenarios/e2e/p0-021-workspace-handoff/scripts/cleanup.sh
```

## 4 AC 映射

| spec AC / decision | 覆盖场景 |
|--------------------|----------|
| C-1 owner route takeover | `E2E.P0.018` |
| C-2 workspace pure plan-list landing and parse detail handoff | `E2E.P0.018`, `E2E.P0.019` |
| C-2a workspace plan-list card visual affordance + concise metadata boundary | `E2E.P0.018` |
| C-3 workspace interactions and start practice | `E2E.P0.018`, `E2E.P0.020` |
| C-3 persistent workspace archive delete | `E2E.P0.018` |
| C-7 downstream handoff boundary | `E2E.P0.021` |
| C-8 / C-9 UI parity | `E2E.P0.018` |
| C-10 non-current negative search | `E2E.P0.021` |
| C-12 privacy redline | `E2E.P0.020`, `E2E.P0.021` |
| C-13 parse detail regression and workspace legacy-param purity | `E2E.P0.018`, parse/report focused gates |
