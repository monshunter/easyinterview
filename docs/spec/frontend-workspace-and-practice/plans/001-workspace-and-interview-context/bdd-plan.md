# 001 BDD Plan

> **版本**: 1.20
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 Plan**: [plan](./plan.md)
**关联 Checklist**: [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 场景矩阵

| 场景 ID | 名称 | 类别 | 验证入口 |
|---------|------|------|----------|
| `E2E.P0.018` | 面试入口规划列表 + parse 统一面试规划详情 handoff | primary + alternate + UX regression | `test/scenarios/e2e/p0-018-workspace-default-render/` |
| `E2E.P0.021` | Workspace boundary + privacy/out-of-scope negative | regression + privacy | `test/scenarios/e2e/p0-021-workspace-handoff/` |
| `E2E.P0.045` | Practice structured-round budget display | primary + UX regression | `test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/` |
| `E2E.P0.057` | Report retry / next-round handoff boundaries | primary + boundary + recovery | `test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/` |

## 2 场景明细

### E2E.P0.018 面试入口规划列表 + 统一面试规划详情

| Given | When | Then |
|-------|------|------|
| 用户已登录；`listTargetJobs=default` 且可能混入 failed / blank-title stale data；点击顶部 `面试`；另有 out-of-scope workspace route params 含 `targetJobId / jdId / resumeId / roundId`；real-backend 场景存在可归档 ready TargetJob | 进入 workspace 默认 landing；点击一个规划卡片；进入 parse 统一详情；从详情页再次点击 TopBar `面试`；点击另一个卡片右上角的删除图标；刷新 workspace；切换 zh/en、dark/customAccent | TopBar 只显示 `首页 / 面试 / 简历`；workspace canonicalize 为 `/workspace`、始终渲染面试规划列表且不显示缺 JD 死胡同；列表请求 `analysisStatus=ready`，failed / blank-title TargetJob 不渲染为卡片；规划卡片以卡片背景、边框、轻阴影、body/footer 分区、footer `立即面试` 和右上角删除图标呈现，卡片内不展示 `手动输入` / 来源类型 / 目标语言等导入元信息，并通过 `listTargetJobs` 返回的 target job-level `resumeId` 导航 `parse`；删除图标调用 generated `archiveTargetJob`，成功后卡片消失且刷新后不回灌，失败时保留卡片；out-of-scope workspace params 不触发 `getTargetJob` / `parse-error`；有 `currentPracticePlanId` 时携带真实 `planId`，无 plan 时不得伪造；详情 DOM 与 Parse ready state 同源，out-of-scope independent workspace detail anchors 不出现；out-of-scope prototype testid 0 命中 |

### E2E.P0.021 Workspace boundary and privacy

| Given | When | Then |
|-------|------|------|
| 用户已登录；workspace plan-list 是当前 runtime；records typed consumer is outside this completed plan | 运行 workspace source negative、report replay handoff regression、privacy grep 和 out-of-scope grep | Workspace runtime does not call standalone insight API, report API, untyped fixture extension or prototype helper；report replay handoff stays covered by report owner tests；privacy and out-of-scope negative grep pass |

### E2E.P0.045 Practice structured-round budget display

| Given | When | Then |
|-------|------|------|
| 当前 session 对应 ready PracticePlan，plan 的 `timeBudgetMinutes` 来自当前 TargetJob 结构化轮次；另有 plan read loading/failure 变体 | 进入或刷新 Practice | Top Bar 显示 plan budget（例如 60 分钟显示 `60:00`）且不存在固定 `25:00`；loading/failure 不伪造预算；elapsed 超过预算也不自动完成会话 |

### E2E.P0.057 Report retry / next-round handoff boundaries

| Given | When | Then |
|-------|------|------|
| ready report 与 TargetJob 有按 sequence 排序的 1..N 轮；当前 round 可能是中间、末轮、未知或缺失，round list 可能产生重复 ID，round data 可能 loading/failure | 点击复练当前轮或进入下一轮，并覆盖重复点击 | 复练保持当前轮；下一轮只选择紧邻后一轮并用其时长创建 plan/session；重复 ID、末轮、单轮、空/未知轮次、loading/failure 不触发 next start；in-flight CTA disabled，重复点击最多创建一次；不回退第一轮或固定默认轮次 |

## 3 执行入口

```bash
test/scenarios/e2e/p0-018-workspace-default-render/scripts/setup.sh && test/scenarios/e2e/p0-018-workspace-default-render/scripts/trigger.sh && test/scenarios/e2e/p0-018-workspace-default-render/scripts/verify.sh && test/scenarios/e2e/p0-018-workspace-default-render/scripts/cleanup.sh
test/scenarios/e2e/p0-021-workspace-handoff/scripts/setup.sh && test/scenarios/e2e/p0-021-workspace-handoff/scripts/trigger.sh && test/scenarios/e2e/p0-021-workspace-handoff/scripts/verify.sh && test/scenarios/e2e/p0-021-workspace-handoff/scripts/cleanup.sh
test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/setup.sh && test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/trigger.sh && test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/verify.sh && test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/cleanup.sh
test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/scripts/setup.sh && test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/scripts/trigger.sh && test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/scripts/verify.sh && test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/scripts/cleanup.sh
```

## 4 AC 映射

| spec AC / decision | 覆盖场景 |
|--------------------|----------|
| C-1 owner route takeover | `E2E.P0.018` |
| C-2 workspace pure plan-list landing and parse detail handoff | `E2E.P0.018` |
| C-2a workspace plan-list card visual affordance + concise metadata boundary | `E2E.P0.018` |
| C-3 workspace interactions and start practice | `E2E.P0.018` + parse/report focused start-practice gates |
| C-3 persistent workspace archive delete | `E2E.P0.018` |
| C-7 downstream handoff boundary | `E2E.P0.021` |
| C-8 / C-9 UI parity | `E2E.P0.018` |
| C-10 out-of-scope negative search | `E2E.P0.021` |
| C-12 privacy redline | `E2E.P0.021` + parse/report focused handoff gates |
| C-13 parse detail regression and workspace out-of-scope-param purity | `E2E.P0.018`, parse/report focused gates |
| C-11 structured round time budget and next-round progression | `E2E.P0.021`, `E2E.P0.045`, `E2E.P0.057` |
