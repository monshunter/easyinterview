# Backend Practice 001 Orchestration 交付复盘报告

> **日期**: 2026-05-09
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-practice/001-plan-and-session-orchestration` 全计划交付，覆盖 Phase 0 contract preflight、Phase 1 plan/session success path、Phase 2 idempotency recovery、Phase 3 observability/privacy/lifecycle closure。
- 关键代码与场景交付：shared idempotency middleware、practice handler/service/store、session reservation/commit/failure 状态机、A3 observed AI task-run context、session start audit event、Phase 3 legacy lint、`E2E.P0.022` 至 `E2E.P0.026`。
- 成功证据：
  - `cd backend && go test ./cmd/api ./internal/api/practice ./internal/practice ./internal/store/practice ./internal/middleware/idempotency/... -count=1`
  - `python3 -m pytest scripts/lint/backend_practice_legacy_test.py`
  - `make lint-backend-practice-legacy`
  - `make validate-fixtures`
  - `make lint-events`
  - `make docs-check`
  - `python3 scripts/lint/migrations_lint.py --repo-root .`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-practice/plans/001-plan-and-session-orchestration/context.yaml`
  - `test/scenarios/e2e/p0-026-practice-observability-and-privacy-redlines/scripts/{setup,trigger,verify,cleanup}.sh`

## 2 会话中的主要阻点/痛点

- Phase 3 legacy-negative gate 一开始会自扫到旧场景脚本里的负向搜索字面量。
  - **证据**：`backend_practice_legacy.py --phase all` 首次命中 `p0-022 ... verify.sh` 与 `expected-outcome.md` 中的 retired enum literal。
  - **影响**：需要额外修正 P0.022 脚本写法，并在 lint 脚本中明确 Phase 3 扫描范围和 `verify.sh` 自引用排除。
- `3.4` checklist 同时要求 redaction unit test 和 scenario `verify.sh` grep，和 `3.6` 的 P0.026 场景创建/执行存在交叉。
  - **证据**：3.4 的单元测试已通过后，仍必须等 P0.026 `verify.sh` 落地和执行才可勾选。
  - **影响**：执行顺序需要人工判断，增加了 checklist 状态同步的歧义。
- Observed AI path 对 `AITaskRunContext` 的 UUID 校验与场景 harness 的可读字符串用户 ID 不一致。
  - **证据**：Phase 3 接入 observed decorator 后，task-run context 必须满足 A3 `Validate()`；最终将 practice HTTP 场景用户 ID 改为 UUID 格式，并同步 seed 文档。
  - **影响**：若未来场景继续使用非 UUID fixture，observability 场景会在业务断言前因 A3 类型约束失败。
- `3.8` 同时包含 commit、work-journal、retrospective，严格来说 retrospective 必须发生在 phase commit 之后。
  - **证据**：Phase 3 commit 已完成后，才具备 retrospective 的完整提交和验证证据，因此需要一个 follow-up close-out commit 勾选 3.8 并提交复盘报告。
  - **影响**：单个 checklist item 混合了 phase commit 与 post-commit 工作，天然不能在同一 phase commit 内完整收口。

## 3 根因归类

- Phase 3 grep 自引用属于 `spec-plan` + `scenario` 资产边界问题：计划要求 retired terms zero-hit，但没有说明 grep 脚本自身如何表达禁用词。
- 3.4 / 3.6 交叉属于 `spec-plan` gate 粒度问题：一个 checklist item 同时绑定 unit gate 与 BDD script gate。
- UUID fixture 问题属于 `README` / scenario convention 缺口：场景框架说明了目录与隔离，但没有说明命中 typed persistence / A3 task-run validation 时 ID 形态应贴近真实 UUID。
- 3.8 post-commit item 属于 `spec-plan` checklist 编排问题：post-pass retrospective 与 phase commit 的依赖方向没有拆开。

## 4 对流程资产的改进建议

- 建议在未来含 negative grep 的 plan/test-plan 中明确：禁止词不得以连续字面量写入被扫描范围；若脚本必须持有禁用词，应使用变量拼接或把脚本自身列入显式排除。
  - **落点**：spec-plan / scenario-create skill
  - **优先级**：high
- 建议把 checklist 中“unit redaction gate”和“scenario verify gate”拆成两个相邻条目，避免一个条目跨越不同执行时点。
  - **落点**：spec-plan
  - **优先级**：medium
- 建议在 `test/scenarios/e2e/README.md` 或 scenario-create 规则中补一句：当场景经过 typed DB/A3 task-run validation 时，用户、资源、计划、session fixture ID 应使用 UUID 格式。
  - **落点**：README / scenario-create skill
  - **优先级**：medium
- 建议后续 plan 不再把“phase commit + work-journal + retrospective”合并为一个同一提交内完成的 checklist 项；可拆为“phase commit + work-journal”和“post-pass retrospective close-out”。
  - **落点**：spec-plan / design skill
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：更新 scenario-create 或 plan 模板，要求 negative grep gate 避免自引用禁用词。这类问题会直接让 CI gate 误报失败。
- 次优先级：在 e2e README 中补 UUID fixture 约束，避免 observed/persistence 场景继续用可读字符串 ID 触发类型校验失败。
- 可延后：调整 design/plan checklist 模板，把 post-pass retrospective 拆成独立 close-out 条目；这主要减少收尾提交摩擦，不影响功能正确性。
