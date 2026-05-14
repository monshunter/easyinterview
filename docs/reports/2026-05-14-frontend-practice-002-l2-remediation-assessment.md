# Frontend Practice 002 L2 Remediation 交付复盘报告

> **日期**: 2026-05-14
> **审查人**: Codex

## 1 复盘范围与成功证据

- 修复范围：`frontend-workspace-and-practice/002-practice-text-event-loop` L2 review remediation，覆盖 practice runtime 错误恢复、隐私红线、pixel parity、P0.046 / P0.047 scenario trigger、legacy negative gate 与 plan/checklist/BDD lifecycle 证据同步。
- Bug 资产：[BUG-0057](../bugs/BUG-0057.md) 记录本次 recovery / parity / completed-evidence drift，状态 `resolved`。
- 计划资产：原 plan 002 `plan.md` / `checklist.md` / `test-checklist.md` / `bdd-checklist.md` 已原地 bump 到 v1.3，`plans/INDEX.md` 已由 `/sync-doc-index --fix-index` 同步。
- 验证证据：practice 新增 recovery/privacy/negative tests、practice focused suite 27 files / 120 tests、frontend full Vitest 154 files / 907 tests、typecheck、frontend build、practice Playwright pixel parity 11 passed / 1 skipped、P0.044-P0.047 scenario loop、workspace P0.018-P0.021 regression、backend-practice P0.022-P0.026 regression、backend-practice 002 Go focus P0.038-P0.043、`make build`、`validate_fixtures.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check` 均通过。

## 2 会话中的主要阻点/痛点

- Completed lifecycle 与真实证据不一致。
  - **证据**：plan 002 已是 `completed`，但 checklist / test-checklist 仍有 `partial` 证据，bdd-checklist 所有条目仍未勾选。
  - **影响**：若只信 lifecycle，会把未闭合的错误恢复、隐私和 BDD wrapper 覆盖误判为 PASS。
- Scenario wrapper 没有绑定专用 runtime 断言。
  - **证据**：P0.046 / P0.047 trigger 原先没有执行 `practiceErrors`、`practiceClientEventConflict`、`practiceConflict`、`practicePrivacy`；verify 文案无法替代缺失的测试执行。
  - **影响**：AI timeout retry、409 mismatch refresh、strict conflict warning、raw text 泄漏红线都可能在 scenario 层假绿。
- UI parity gate 没有覆盖 token computed style。
  - **证据**：新增 Playwright practice spec 后发现 practice runtime 仍引用旧 token 变量；仅靠 DOM anchor / full Vitest 无法捕获这类视觉漂移。
  - **影响**：正式 frontend 可能看似结构正确，但与 `ui-design` token 真理源不一致。

## 3 根因归类

- Completed evidence drift。
  - **类别**：skill / spec-plan
  - **根因**：L2 review 规则没有显式把 `partial`、未勾选 BDD、asset-readiness 语言列为阻塞 drift；plan close-out 允许历史资产状态遮蔽真实 gate 状态。
- Scenario trigger coverage gap。
  - **类别**：spec-plan
  - **根因**：BDD checklist 描述了目标行为，但 scenario trigger 没有强制绑定 dedicated tests，导致 README / INDEX / verify 文案和实际执行测试之间脱节。
- UI token drift。
  - **类别**：spec-plan
  - **根因**：pixel parity gate 原先未覆盖 practice route 的 computed style 与 mobile layout，旧 token 命名没有负向搜索或 runtime 断言。

## 4 对流程资产的改进建议

- 已在 `plan-code-review` skill 中固化 completed evidence drift 规则。
  - **落点**：skill
  - **优先级**：high
  - 建议：后续 L2 review 遇到 completed 文档中的 `partial` / `pending` / 未勾选 BDD / asset-readiness 语言时，直接记录阻塞 finding，并在 `--fix` 中先补 executable gate 再恢复 completed。
- 后续 feature plan 的 BDD close-out 应把 `trigger.sh` / `verify.sh` 内容纳入 checklist 证据。
  - **落点**：spec-plan
  - **优先级**：high
  - 建议：scenario 不是只检查目录和 INDEX；必须证明 trigger 执行了对应 runtime tests，verify 覆盖主路径、失败恢复、隐私与 legacy negative 条件。
- UI parity plan 应固定 token computed style 与 responsive bounding box gate。
  - **落点**：spec-plan
  - **优先级**：medium
  - 建议：所有正式 frontend 页面迁移 `ui-design` 时，至少保留一条 Playwright computed style 断言和一条 mobile overflow / layout 断言。

## 5 建议优先级与后续动作

- 最高优先级：进入下一个 frontend practice owner workstream 前，先对其 plan checklist 做一次 L1/L2 gate audit，确认所有 BDD wrapper 都绑定 dedicated trigger tests，且 completed 文档没有 `partial` 或 asset-readiness 证据。
- 次优先级：如果后续 practice route 增加 real backend hint / provenance 能力，优先把 P0.046 / P0.047 的 fixture-backed 断言扩展为 real handler regression，而不是只增加新的 fixture variant。
