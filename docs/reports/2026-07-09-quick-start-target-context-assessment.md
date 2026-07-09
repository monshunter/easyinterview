# Quick Start Target Context 交付复盘报告

> **日期**: 2026-07-09
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 Home recent 与 Workspace plan-list quick-start 丢失结构化 `roundId/roundName`，并修复 Home recent 未限制 ready TargetJob 的准入问题。关联 Bug: [BUG-0152](../bugs/BUG-0152.md)。
- 成功证据：
  - `pnpm --filter @easyinterview/frontend exec vitest run src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/workspace/WorkspaceEmptyState.test.tsx` PASS
  - `pnpm --filter @easyinterview/frontend exec vitest run src/app/screens/parse/ParseFlow.test.tsx src/app/screens/parse/ParseResumeBinding.test.tsx src/app/screens/parse/ParseAuthGate.test.tsx` PASS
  - `pnpm --filter @easyinterview/frontend typecheck` PASS
  - `validate_context.py` for `frontend-home-job-picks-and-parse/001` and `frontend-workspace-and-practice/001` PASS
  - `sync-doc-index --check`, `make docs-check`, `git diff --check` PASS

## 2 会话中的主要阻点/痛点

- Home / Workspace / Parse had separate practice route-param construction.
  - **证据**：review red tests failed because Home and Workspace quick-start produced practice params without `roundId/roundName`, while Parse already used structured TargetJob context.
  - **影响**：same user action family had different context fidelity depending on entry point.

- Home recent list admission lagged behind Workspace list admission.
  - **证据**：red test showed Home called `listTargetJobs({ pageSize: "12" })` and rendered a processing TargetJob fixture.
  - **影响**：non-ready TargetJobs could appear in a startable shortcut surface.

## 3 根因归类

- Duplicated shortcut handoff logic.
  - **类别**：spec-plan
  - **说明**：owner plan said quick-start should use saved round context, but focused tests did not require the same route params as Parse start.

- List admission contract was not shared across same-card surfaces.
  - **类别**：spec-plan
  - **说明**：Workspace had ready-only query/filter assertions; Home recent reused the visual card but not the admission gate.

## 4 对流程资产的改进建议

- Keep shortcut start gates explicit in owner checklists.
  - **落点**：spec-plan
  - **优先级**：high
  - **建议**：when a list/card surface can start practice, checklist items should assert both route context fields (`targetJobId`, `resumeId`, `planId`, `roundId`, `roundName`) and the source list admission query/filter.

- Prefer shared route-param helpers for entry points that launch the same downstream flow.
  - **落点**：spec-plan
  - **优先级**：medium
  - **建议**：future frontend plans should call out shared helpers as an implementation invariant when Home, Workspace, Parse, or Report start the same practice handoff.

## 5 建议优先级与后续动作

- Highest value next action: include `BUG-0152` in the work-journal / commit closeout so the review remediation, Bug record, and owner plan updates stay linked.
- Follow-up to defer: consider adding a broader frontend handoff pattern only if another shortcut path repeats the same context-field drift.
