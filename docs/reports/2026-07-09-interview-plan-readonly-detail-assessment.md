# Interview Plan Readonly Detail 交付复盘报告

> **日期**: 2026-07-09
> **审查人**: Codex

## 1 复盘范围与成功证据

- 交付范围：将 `frontend-home-job-picks-and-parse/001-home-jd-import-and-parse` 的统一面试规划详情从可编辑确认页收敛为只读上下文收据；移除成功态 Re-parse、Save plan、Cancel、字段编辑、requirements toggle、hidden signal remove、resume picker / create fallback；Start 直接使用已保存 TargetJob / Resume / PracticePlan 快照进入 practice，不再调用 `updateTargetJob`。
- 已同步资产：owner spec / plan / checklist / BDD、`docs/ui-design/`、`ui-design/src/screens-p0-complete.jsx`、formal frontend、pixel parity、P0.016 场景资产。
- 通过证据：
  - `pnpm --filter @easyinterview/frontend exec vitest run src/app/screens/parse/ParseScreen.test.tsx src/app/screens/parse/ParseEdit.test.tsx src/app/screens/parse/ParseResumeBinding.test.tsx src/app/screens/parse/ParseAuthGate.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceEmptyState.test.tsx src/app/navigation/interviewContext.test.ts`
  - `node --test ui-design/ui-design-contract.test.mjs`
  - `pnpm --filter @easyinterview/frontend typecheck`
  - `pnpm --filter @easyinterview/frontend build`
  - `pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/parse.spec.ts tests/pixel-parity/workspace.spec.ts --grep "ready target job response|readonly plan detail exposes|start interview hands off directly|parse detail renders|parse detail keeps resume binding"`
  - `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/trigger.sh && test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/verify.sh`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/context.yaml --docs-root docs --target frontend`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`
  - `git diff --check`
  - `make lint-core-loop-pruning-surface`

## 2 会话中的主要阻点/痛点

- 旧 completed 计划和 BDD 仍把 P0.016 定义为编辑、Save、`updateTargetJob` 和 workspace auto-start。
  - **证据**：本次必须原地重开 plan 001、更新 Phase 6，并重写 P0.016 README / seed / expected / trigger / verify。
  - **影响**：如果只改 UI，实现会被旧测试和 scenario wrapper 拉回可编辑语义。
- UI 真理源、正式前端、pixel parity 三处同时保护同一行为，变更面较分散。
  - **证据**：同一旧口径同时存在于 `ui-design/src/screens-p0-complete.jsx`、`ui-design/ui-design-contract.test.mjs`、`ParseScreen.tsx`、parse focused tests 和 pixel tests。
  - **影响**：需要先改 owner 合同，再改静态原型，最后改正式前端和 browser gates；顺序错误会造成通过局部测试但违背 UI truth source。

## 3 根因归类

- 旧行为作为“已完成”场景被测试固化，但产品语义已变化。
  - **类别**：spec-plan
- P0.016 场景名仍叫 `parse-confirm-to-workspace`，但当前行为已经是 parse readonly receipt -> practice。
  - **类别**：spec-plan
- 本次没有发现需要修改 `AGENTS.md` 或 skill 的流程缺陷；现有原地修订、TDD、BDD、retrospective 规则足以覆盖。
  - **类别**：no repo change needed

## 4 对流程资产的改进建议

- 在 `frontend-home-job-picks-and-parse` 后续变更中，继续把 P0.016 作为“readonly receipt + direct Start”主 gate，禁止恢复 Save / PATCH / picker 语义。
  - **落点**：spec-plan
  - **优先级**：high
- 如后续清理场景目录命名，可考虑把 `p0-016-parse-confirm-to-workspace` 改名为更贴近当前语义的 `p0-016-parse-readonly-start-handoff`，但应单独做 rename 以避免和本次行为变更混在一起。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- high：下一轮若继续做面试规划详情，应优先补一个 source negative gate，扫描正式前端正向 runtime 中不得出现 `parse-action-save-plan`、`parse-resume-picker-toggle`、`updateTargetJob` parse success consumer 等旧锚点。
- medium：在没有紧急需求时，再做 P0.016 目录 rename；当前目录名虽然历史化，但 README / scripts / expected outcome 已经表达当前合同。
