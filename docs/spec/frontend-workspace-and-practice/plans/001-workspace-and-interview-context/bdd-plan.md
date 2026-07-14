# 001 BDD Plan

> **版本**: 1.24
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)
**关联 Checklist**: [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 场景矩阵

| 场景 ID | 名称 | 类别 | 验证入口 |
|---------|------|------|----------|
| `E2E.P0.018` | Workspace 规划列表 + targetJobId 只读详情 + 轮次三态 | primary + alternate + UX/request-count regression | `test/scenarios/e2e/p0-018-workspace-default-render/` |
| `E2E.P0.021` | Workspace boundary + privacy/out-of-scope negative | regression + privacy | `test/scenarios/e2e/p0-021-workspace-handoff/` |
| `E2E.P0.045` | Practice structured-round budget display | primary + UX regression | `test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/` |
| `E2E.P0.057` | Report retry / next-round handoff boundaries | primary + boundary + recovery | `test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/` |
| `E2E.P0.098` | Persisted multi-round progress and quick-start | primary + persistence + recovery | `test/scenarios/e2e/p0-098-full-funnel-import-to-next-round-journey/` |

## 2 场景明细

### E2E.P0.018 Workspace 列表 + 统一只读详情

| Given | When | Then |
|-------|------|------|
| 用户已登录；ready list 可混入 failed/blank-title stale data；底层 transport 可计数；存在一个可信 ready TargetJob，其 progress 可形成 done/current/pending，与 hostile plan/resume/auto-start query | 在 StrictMode 打开 `/workspace`，观察列表 rail，点击 card body，刷新 `/workspace?targetJobId`，再用 TopBar 返回列表；覆盖 not-found/mismatch、归档、quick-start、三态 computed style 和 parity | Query-free route 展示列表且 `listTargetJobs` same-key initial count=1；card 只携带 targetJobId 直达 workspace detail；detail `getTargetJob` same-key initial count=1，统一只读母版无 Parse animation，零 import/poll/route-side start；详情 done/current/pending 标签、属性、背景/边框与 rail 状态序列一致；hostile params stripped；失败 fail closed；归档、quick-start、zh/en/theme 与 desktop/mobile parity 保持 |

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
| ready report 与 TargetJob 有正 int32、唯一、严格递增但可能非连续的 canonical 轮次（如 `1,2,4`）；当前 round 可能是中间、末轮、未知或缺失，round list 可能产生重复 ID，round data 可能 loading/failure | 点击复练当前轮或进入下一轮，并覆盖重复点击 | 复练保持当前轮；下一轮只选择数组中的下一条现有 canonical round（`2→4`）并用其时长创建 plan/session，不计算 `sequence + 1`；重复 ID、末轮、单轮、空/未知轮次、loading/failure 不触发 next start；in-flight CTA disabled，重复点击最多创建一次；不回退第一轮或固定默认轮次 |

### E2E.P0.098 Persisted multi-round progress and quick-start

| Given | When | Then |
|-------|------|------|
| Real backend TargetJob has canonical `1,2,4` rounds; plans persist exact pairs; first round completes, including wrong-resume/duplicate completion/retry/report-state variants; frontend runs in real API mode | In a live browser reload Home/Workspace/Parse, inspect rail, click quick-start, capture the real create/start exchange, then complete remaining rounds | API-projected completed prefix/current next-existing round survives browser refresh and route changes; quick-start uses only exact current plan or creates a plan with current `roundId`; equal-duration/legacy/wrong-resume plan is not reused; all completed shows all nodes done and disables start; no browser-stored business progress is read |

> 证据边界：真实 PostgreSQL Get/List projection、Vitest consumer tests、scope/static negative 与 pixel parity 都是必要支持证据，但不能单独勾选本场景的 live browser reload/quick-start 项。只有实际 host-run frontend/backend browser execution 及请求/响应证据存在后才可完成。

## 3 执行入口

```bash
test/scenarios/e2e/p0-018-workspace-default-render/scripts/setup.sh && test/scenarios/e2e/p0-018-workspace-default-render/scripts/trigger.sh && test/scenarios/e2e/p0-018-workspace-default-render/scripts/verify.sh && test/scenarios/e2e/p0-018-workspace-default-render/scripts/cleanup.sh
test/scenarios/e2e/p0-021-workspace-handoff/scripts/setup.sh && test/scenarios/e2e/p0-021-workspace-handoff/scripts/trigger.sh && test/scenarios/e2e/p0-021-workspace-handoff/scripts/verify.sh && test/scenarios/e2e/p0-021-workspace-handoff/scripts/cleanup.sh
test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/setup.sh && test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/trigger.sh && test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/verify.sh && test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/cleanup.sh
test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/scripts/setup.sh && test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/scripts/trigger.sh && test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/scripts/verify.sh && test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/scripts/cleanup.sh
test/scenarios/e2e/p0-098-full-funnel-import-to-next-round-journey/scripts/setup.sh && test/scenarios/e2e/p0-098-full-funnel-import-to-next-round-journey/scripts/trigger.sh && test/scenarios/e2e/p0-098-full-funnel-import-to-next-round-journey/scripts/verify.sh && test/scenarios/e2e/p0-098-full-funnel-import-to-next-round-journey/scripts/cleanup.sh
```

## 4 AC 映射

| spec AC / decision | 覆盖场景 |
|--------------------|----------|
| C-1 workspace list/detail route split | `E2E.P0.018` |
| C-2 query-free list and targetJobId-only detail | `E2E.P0.018` |
| C-2a workspace plan-list card visual affordance + concise metadata boundary | `E2E.P0.018` |
| C-3 workspace interactions and start practice | `E2E.P0.018` + parse/report focused start-practice gates |
| C-3 persistent workspace archive delete | `E2E.P0.018` |
| C-7 downstream handoff boundary | `E2E.P0.021` |
| C-8 / C-9 UI parity | `E2E.P0.018` |
| C-10 out-of-scope negative search | `E2E.P0.021` |
| C-12 privacy redline | `E2E.P0.021` + parse/report focused handoff gates |
| C-13 workspace detail exact GET, zero Parse animation/import/poll and safe-param purity | `E2E.P0.018`, shell focused gates |
| C-17 workspace detail persisted round-state labels, visual distinction and rail consistency | `E2E.P0.018`, `E2E.P0.098` |
| C-11 structured round time budget and next-round progression | `E2E.P0.021`, `E2E.P0.045`, `E2E.P0.057` |
| C-12 backend-persisted progress, refreshed rail and exact quick-start | `E2E.P0.018`, `E2E.P0.021`, `E2E.P0.057`, `E2E.P0.098` |

## 5 Phase 26-27 current-scope gate

Phase 26-27 原地修订 `E2E.P0.018`，不新建 sibling。Request count 必须来自底层 transport；不得关闭 StrictMode。P0.018 同时证明 workspace detail 零 import/Parse poll/Parse animation/route-side start，并读取可见三态标签、`data-round-state`、computed background/border 与列表 rail 序列；ready replace 与 Back 不重播动画由 shell/004 Phase 12 提供交叉证据。全完成/刷新持久化由 P0.098 交叉覆盖，无效投影由 focused fail-closed gate 覆盖。
