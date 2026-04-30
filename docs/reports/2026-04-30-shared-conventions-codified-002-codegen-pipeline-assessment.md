# shared-conventions-codified/002-codegen-pipeline 交付复盘报告

> **日期**: 2026-04-30
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖 `shared-conventions-codified/002-codegen-pipeline`：AI shared vocabulary 真理源、Go/TS generated outputs、drift wrapper、`make codegen-check` 接入、Go/TS parity tests、shared parity fixture、future handoff guard 与 plan lifecycle sync。
- 计划状态已收口：`plan.md` / `checklist.md` Header 均为 `completed`，`checklist.md` 13/13 items 完成，`plans/INDEX.md` 已将 002 移至 Completed。
- 通过验证：
  - `make codegen-conventions`
  - `make codegen-check`
  - `go test ./cmd/codegen/conventions ./internal/shared/... -count=1`
  - `pnpm --dir frontend test src/lib/conventions src/lib/ids`
  - `pnpm --dir frontend typecheck`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
- 三类 negative drift 已验证：YAML-only drift 同时报 Go/TS AI paths；Go-only drift 报 `backend/internal/shared/ai/vocabulary.go`；TS-only drift 报 `frontend/src/lib/conventions/ai.ts`；恢复后 `git diff --check` 与 `make codegen-check` 通过。

## 2 会话中的主要阻点/痛点

- Active plan 命中新质量门禁后无法直接实施。
  - **证据**：`/implement` Step 4.2 发现 002 plan 缺少 `## 3 质量门禁分类` 和 checklist 可执行验证断言，必须先做 `/plan-review --fix` 风格文档修复。
  - **影响**：实现开始前增加一次文档修订和确认往返；所幸修订范围清晰，没有改变 plan scope。
- 分支规则与当前工作分支不一致。
  - **证据**：session branch detector 判定当前 `feature/spec-init` 不符合 `shared-conventions-codified-002-codegen-pipeline` session branch contract；用户选择方案 A 后才继续。
  - **影响**：需要用户显式确认才能在当前分支执行；后续 phase commit 仍可 ff-only merge 回 `dev`。
- Negative drift 验证需要手动临时改写 generated outputs。
  - **证据**：6.2 通过临时修改 YAML、Go generated、TS generated 文件验证 gate failure，再用 generator 恢复。
  - **影响**：验证有效但操作成本高，且要求执行者严格恢复现场。

## 3 根因归类

- 新质量门禁落地后，旧 active plan 未自动补齐分类。
  - **类别**：spec-plan
  - **说明**：002 plan 创建早于 TDD/BDD quality gate codification，缺口属于计划文档随治理升级产生的漂移。
- `/implement` branch gate 对“用户批准沿当前分支继续”的路径没有一等记录机制。
  - **类别**：skill
  - **说明**：当前流程能正确阻止误切分支，但用户批准 override 后只靠会话上下文表达，缺少结构化记录。
- Negative drift gate 缺少 repo-tracked scenario helper。
  - **类别**：README / tooling
  - **说明**：`conventions_drift.py` 有 unit contract，但 `make codegen-check` 的 end-to-end negative cases 仍依赖手工临时编辑。

## 4 对流程资产的改进建议

- 在 quality-gate governance 变更后，对 active plans 做一次 targeted plan-review sweep。
  - **落点**：spec-plan / `/plan-review`
  - **优先级**：medium
  - **建议**：新增一个轻量 checklist 或命令说明，用于发现 active plan 是否缺 `## 3 质量门禁分类`、BDD 不适用说明或 checklist `验证:` 子句。
- 为 `/implement` 增加“用户批准当前分支继续”的记录建议。
  - **落点**：skill
  - **优先级**：low
  - **建议**：当 branch detector mismatch 但用户明确选择继续时，要求最终 work journal 或 plan notes 记录 branch override reason，避免后续审计误判。
- 为 conventions drift negative checks 增加 repo-tracked helper。
  - **落点**：README / tooling
  - **优先级**：medium
  - **建议**：在 `scripts/lint/conventions_drift.py` 或 Makefile 增加测试模式，在临时目录中模拟 YAML-only / Go-only / TS-only drift，减少人工编辑 generated 文件的恢复风险。

## 5 建议优先级与后续动作

- **Medium**：补一个 active plan quality-gate sweep 入口，优先降低未来 `/implement` 到 Step 4.2 才发现文档缺口的概率。
- **Medium**：给 conventions drift 增加 end-to-end negative helper，后续 B1/A5 扩展 drift gate 时能复用。
- **Low**：完善 `/implement` branch override 记录规则；当前已有用户确认和 phase journal，不阻塞后续开发。
