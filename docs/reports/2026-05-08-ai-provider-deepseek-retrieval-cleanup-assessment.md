# AI Provider DeepSeek And Retrieval Cleanup 交付复盘报告

> **日期**: 2026-05-08
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：A3 `003-provider-registry-and-capability-profiles` Phase 6，删除当前 active scope 的 embedding / rerank 实现与基础设施，并将 repo-tracked 开发期 chat provider 收敛为 `deepseek` + `deepseek-v4-flash` / `deepseek-v4-pro`。
- 成功证据：focused Go tests 通过；`python3 scripts/lint/ai_profile_coverage.py --repo-root .` 通过；`python3 scripts/lint/migrations_lint.py --repo-root .` 通过；相关 lint pytest 76 passed / 1962 subtests passed；frontend conventions/jobs Vitest 33 passed；`make lint-config`、`make lint-events`、`make docs-check` 通过；active-scope negative search 仅剩 Go/OpenAPI 自身 `go:embed` / `embedded` 语义命中。

## 2 会话中的主要阻点/痛点

- `make codegen-check` 在 dirty worktree 中报告 intended generated diff。
  - **证据**：`make codegen-check` 输出 jobs generated artifact diff；随后 `make codegen-conventions && make codegen-events` 成功，说明生成器本身可运行，失败点是 check target 与未提交变更的交互。
  - **影响**：验证叙事需要区分“生成器不一致”与“当前变更尚未提交”，否则容易误判。
- dev-stack env 解析对空值行有边界风险。
  - **证据**：`ai_profile_coverage` 首次误报 `AI_MODEL_PROFILE_PATH` missing；原因是 `ENV_LINE_RE` 中 `\s*` 可跨行吞掉空 `AI_PROVIDER_API_KEY=` 后的下一行。
  - **影响**：空 secret 占位是项目常态，lint parser 必须逐行解析，不能使用跨行空白匹配。

## 3 根因归类

- codegen check dirty-worktree 误读：
  - **类别**：skill / README
  - **根因**：当前 workflow 没有明确说明“修改 generated truth source 后，先跑 codegen，再用 staged/clean-state 或 diff snapshot 验证 idempotency”。
- env parser 空值风险：
  - **类别**：no repo change needed
  - **根因**：实现缺陷已在本次修复，现有 `ai_profile_coverage_test.py` 覆盖了空 `AI_PROVIDER_BASE_URL=` / `AI_PROVIDER_API_KEY=` 后仍能解析 `AI_MODEL_PROFILE_PATH` 的路径。

## 4 对流程资产的改进建议

- 在 `/tdd` 或 `/implement` 的 codegen gate 说明中补一句：dirty worktree 中 `make codegen-check` 可能展示 intended generated diff；phase 验证应先运行生成命令，再用 staged-state 或 diff snapshot 证明生成器无额外漂移。
  - **落点**：skill
  - **优先级**：low
- 保留 env parser 的空值回归测试作为长期 gate；新增类似 env parser 时复用 `[ \t]*` 而不是 `\s*` 匹配行内空白。
  - **落点**：README / lint implementation convention
  - **优先级**：low

## 5 建议优先级与后续动作

- low：后续整理 `/tdd` codegen gate 文案，避免 dirty-worktree `codegen-check` 输出被当成真正生成漂移。
- low：若后续新增 env parser，把“空 secret 占位后仍能解析下一行 canonical env”的测试模式复制过去。
