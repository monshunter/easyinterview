# Unified Plan Detail 交付复盘报告

> **日期**: 2026-07-09
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：将 `JD 解析结果` 页升级为统一的“面试规划详情 / 面试上下文确认”母版，并让 `workspace?targetJobId=...` 普通详情回访复用该母版；`workspace` 无上下文仍进入面试规划列表，`autoStartPractice=1` 继续承担启动面试的 create/start session 合同。
- Owner 文档：`frontend-home-job-picks-and-parse/001` 与 `frontend-workspace-and-practice/001` 已原地修订，覆盖 unified detail、resume binding、旧 workspace 详情锚点负向断言、BDD 场景和 UI parity gate。
- 成功证据：
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/context.yaml --target frontend`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/context.yaml --target frontend`
  - `pnpm --dir frontend exec vitest run src/app/App.test.tsx src/app/navigation/interviewContext.test.ts src/app/screens/parse/ParseScreen.test.tsx src/app/screens/parse/ParseEdit.test.tsx src/app/screens/parse/ParseResumeBinding.test.tsx src/app/screens/workspace`
  - `pnpm --dir frontend build`
  - `pnpm --dir frontend exec playwright test tests/pixel-parity/parse.spec.ts tests/pixel-parity/workspace.spec.ts`
  - `test/scenarios/e2e/p0-018-workspace-default-render/scripts/trigger.sh && test/scenarios/e2e/p0-018-workspace-default-render/scripts/verify.sh`
  - `test/scenarios/e2e/p0-020-workspace-start-practice/scripts/trigger.sh && test/scenarios/e2e/p0-020-workspace-start-practice/scripts/verify.sh`
  - `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/trigger.sh && test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/verify.sh`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`, `make docs-check`, `git diff --check`

## 2 会话中的主要阻点/痛点

- 两个已完成 owner plan 同时拥有同一用户面试上下文体验。
  - **证据**：首次导入后的详情入口属于 `frontend-home-job-picks-and-parse/001`，既有规划回访和 start-practice 合同属于 `frontend-workspace-and-practice/001`；本次必须同时修订两个 spec/plan/checklist/BDD/context，否则统一页面只能覆盖单一路径。
  - **影响**：如果只改 Parse 或只改 Workspace，会留下第二套详情页、旧锚点或旧文案继续作为可见路径。

- 普通详情回访与 `autoStartPractice=1` 启动路径需要明确分界。
  - **证据**：普通 `workspace?targetJobId=...` 已委托 `ParseScreen` 渲染 unified detail；`autoStartPractice=1` 仍需等待 `TargetJob.currentPracticePlanId` 加载后再执行 `useStartPractice`，否则会破坏创建 practice plan、start session 和 auth recovery 的既有合同。
  - **影响**：归一化不能简单删除 Workspace 启动逻辑；需要测试同时证明普通详情旧锚点 0 命中，以及启动面试场景仍 PASS。

- 前端 source-level UI parity 与真实路由复用存在工程耦合。
  - **证据**：统一母版同时要求 `ui-design` 静态原型、`docs/ui-design`、正式 `ParseScreen`、`WorkspaceScreen` delegation、Vitest、Playwright pixel parity 和 scenario wrapper 同步更新。
  - **影响**：只改正式前端会违反 UI 真理源；只改 UI 文档会无法证明 runtime route 的一致性。

## 3 根因归类

- Parse 与 Workspace 的页面 ownership 旧边界按来源划分，而不是按用户任务划分。
  - **类别**：spec-plan
  - **说明**：用户任务都是“核对 JD / 简历 / 轮次并开始面试”，来源差异不应产生两张视觉和交互不同的详情页。

- Workspace start-practice 合同与 Workspace detail UI 旧实现缠在同一组件内。
  - **类别**：spec-plan
  - **说明**：本次通过普通详情委托 Parse、auto-start 保留启动合同的方式缩小风险，但后续仍可考虑把 start-practice orchestration 从旧详情布局中进一步拆出。

## 4 对流程资产的改进建议

- 在两个 owner plan 中长期保留“统一详情页母版”交叉 gate。
  - **落点**：spec-plan
  - **优先级**：high

- 后续重构 `WorkspaceScreen` 时，优先拆分 start-practice orchestration 与 legacy visual layout，让 `autoStartPractice=1` 不再依赖旧 workspace 详情锚点所在的渲染分支。
  - **落点**：spec-plan
  - **优先级**：medium

- 对涉及页面归一化的 UI 任务，默认同时更新 `ui-design`、`docs/ui-design`、runtime tests、pixel parity 和 scenario docs，避免单一层面的“相似但不统一”。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：下一轮进入 `frontend-workspace-and-practice/001` 时，把 start-practice orchestration 从旧 workspace visual branch 里继续剥离，目标是普通详情和自动启动都不再携带第二套详情页视觉锚点。
- 次优先级：继续保留 P0.016 / P0.018 / P0.020 三条场景组合验证，分别覆盖首次导入确认、既有规划回访和自动启动面试。
- 可延后：如果后续新增规划编辑能力，再决定 unified detail 是否需要拆出独立 `InterviewPlanDetail` 组件；当前直接复用 `ParseScreen` 更符合奥卡姆剃刀原则。
