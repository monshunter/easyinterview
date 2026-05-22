# Frontend Home TargetJob Real Backend Gate 交付复盘报告

> **日期**: 2026-05-22
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`frontend-home-job-picks-and-parse/001-home-jd-import-and-parse` 的 `/plan-code-review --fix`，重点校对 TargetJobs/import/parse fixture-first handoff 是否仍符合当前 backend owner 事实。
- 修复：新增 `frontend/src/api/targetJob.realApiMode.test.ts`；P0.014-P0.016 trigger/verify/README 加 real-mode overlay；plan/spec/BDD/history/index、BUG-0086 与 Bug pattern 同步。
- 证据：P0.014/P0.015/P0.016 setup→trigger→verify→cleanup PASS；backend P0.010/P0.011/P0.012/P0.013 PASS；backend-upload focused route/handler tests PASS；`sync-doc-index --fix-index` zero drift。

## 2 会话中的主要阻点/痛点

- Completed plan 的历史 `not-yet-implemented` wording 已过期。
  - **证据**：plan001 §3.6 仍写 5 个 operation 的 backend handler `not-yet-implemented`；`backend/cmd/api/main.go` 已挂载 `/targets`、`/targets/import`、`/targets/{id}` 与 `/uploads/presign`。
  - **影响**：review 不能只看 checklist completed；必须反查 backend owner 和 runtime route。
- Frontend scenarios 原先只证明 fixture UI variants。
  - **证据**：P0.014-P0.016 trigger 原先只跑 Home/Parse Vitest，没有 real API mode generated-client gate。
  - **影响**：fixture PASS 容易被误读为真实 backend 联调完成。

## 3 根因归类

- 根因：frontend-first plan 完成时 backend owner 尚未完成，后续 backend-targetjob/backend-upload 完成后没有自动回填 frontend owner 的 real-mode gate。
  - **类别**：spec-plan / scenario gate。
- 根因：仓库已有 frontend-first handoff 规则，但缺少“backend owner landed 后 sweep sibling frontend plans”的显式模式。
  - **类别**：Bug pattern。

## 4 对流程资产的改进建议

- 已落地：在 BUG-0086 与 `docs/bugs/PATTERNS.md` 增加 frontend-first handoff 回填模式，要求 completed frontend plan 反查 backend owner、补 real-mode generated-client gate、verify 检查 real-mode marker。
  - **落点**：Bug knowledge base。
  - **优先级**：high。
- 建议后续：在 `/plan-code-review` skill 的 frontend-first review checklist 中增加“same-pattern sibling sweep”提示，尤其是同 subspec 内已有一个 plan 被修复 handoff drift 时。
  - **落点**：`.agent-skills/plan-code-review/SKILL.md` 或共享 review checklist。
  - **优先级**：medium。

## 5 建议优先级与后续动作

- 下一步最值得实施：用 `/plan-code-review frontend-home-job-picks-and-parse/001-home-jd-import-and-parse --fix` 的结果作为模板，抽一个小型 review helper/gate，自动列出 frontend-first plan matrix 中仍为 `not-yet-implemented` 的 backend rows，并要求输入当前 backend owner evidence。
- 可延后：把 P0.014-P0.016 的 fixture UI variants 拆成更细的 named scenario assets；当前 real backend drift 已关闭，拆分只影响可读性，不影响本次联调证据。
