# Frontend Resume Workshop 002 L2 Remediation 交付复盘报告

> **日期**: 2026-05-17
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`plan-code-review frontend-resume-workshop/002-create-flow-and-onboarding --fix`，覆盖 create flow parsing retry、BDD wrapper 证据、legacy negative grep 和 Playwright parity matrix。
- 关键修复：`useResumeParsingPolling.retry()` failed terminal state 后重新启动 polling；P0.081/P0.082/P0.083 verifier 拒绝 no-test / failed summary；legacy negative grep raw gate 真 0；`resume-workshop-create.spec.ts` 补齐 ParseFlow / PreviewConfirm。
- 成功证据：
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-resume-workshop/plans/002-create-flow-and-onboarding/context.yaml --docs-root docs --target frontend`
  - `pnpm --filter @easyinterview/frontend exec vitest run --reporter=verbose src/app/screens/resume-workshop/create/hooks/useResumeParsingPolling.test.tsx`：5 passed。
  - P0.081 `trigger.sh` + `verify.sh`：7 files / 35 tests PASS。
  - P0.082 `setup.sh` + `trigger.sh` + `verify.sh` + `cleanup.sh`：3 files / 18 tests PASS。
  - P0.083 `setup.sh` + `trigger.sh` + `verify.sh` + `cleanup.sh`：5 files / 31 tests PASS。
  - `pnpm --filter @easyinterview/frontend test:pixel-parity tests/pixel-parity/resume-workshop-create.spec.ts`：10 passed。
  - `pnpm --filter @easyinterview/frontend typecheck`、`pnpm --filter @easyinterview/frontend build`、`make docs-check`、`git diff --check` PASS。

## 2 会话中的主要阻点/痛点

- Retry behavior was documented and checked off, but no test exercised failed → retry → ready.
  - **证据**：新增回归测试 Red 阶段在旧实现下停留 `polling`。
  - **影响**：P0.082 recovery path would not recover for users after parse failure.
- Scenario verification and pixel parity had false-green risk.
  - **证据**：three `verify.sh` scripts lacked no-test / failed summary rejection; Playwright spec declared 5 screens but ran only 3 logical screens.
  - **影响**：completed BDD-Gate could pass without proving all declared UI states.
- The negative grep test polluted the raw grep it was supposed to enforce.
  - **证据**：raw retired-term and prototype import grep matched `CreateFlowLegacyNegative.test.ts` before the test strings were split.
  - **影响**：checklist 6.11 / 6.12 evidence was not reproducible from the listed commands.

## 3 根因归类

- Missing retry test coverage.
  - **类别**: spec-plan
  - The plan named retry behavior, but the executable test matrix did not include the terminal failed retry path.
- Wrapper and parity evidence were trusted through summary signals.
  - **类别**: skill
  - `plan-code-review` now caught this by reading wrappers and parity specs directly; no immediate skill change is required, but this remains the highest-value review habit.
- Negative grep tests used forbidden literals directly.
  - **类别**: no repo change needed
  - The local test was fixed by constructing terms from fragments; no general framework change is needed.

## 4 对流程资产的改进建议

- For future UI parity plans, explicitly compare the declared screen matrix with the Playwright logical tests before accepting the parity gate.
  - **落点**: plan-code-review practice / reviewer checklist
  - **优先级**: high
- For negative grep gates, prefer either raw command execution in review or helper tests that build forbidden terms from fragments.
  - **落点**: spec-plan checklist authoring guidance
  - **优先级**: medium
- For recovery paths, require one executable test that proves the user-visible recovery action resumes the underlying async operation, not only that the button renders.
  - **落点**: plan checklist / BDD checklist
  - **优先级**: medium

## 5 建议优先级与后续动作

- Highest value next action: apply the same L2 screen-matrix audit to `frontend-resume-workshop/003-branch-rewrites-and-edit` before implementation starts, because it will add multiple visible states and cross-operation gates.
- Keep Pattern 4 as the active reusable bug pattern; this session produced another example but did not require a new pattern entry.
