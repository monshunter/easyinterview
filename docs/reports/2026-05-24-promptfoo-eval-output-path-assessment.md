# Promptfoo Eval Output Path 交付复盘报告

> **日期**: 2026-05-24
> **审查人**: Codex

## 1 复盘范围与成功证据

范围：将最新 `main` 合并到 `design/e2e-scenarios-p0`，解决 work-journal 冲突，并修复 [BUG-0102](../bugs/BUG-0102.md)：`make eval-offline` 不再把 Promptfoo 生成 tests、SQLite state、DB 和 logs 写入 `config/evals/.generated/`，统一迁移到 `.test-output/evals/`。

成功证据：

- `make eval-offline`：drift-check `52 cases / single-source clean`，offline grading `52 cases`，Promptfoo runner `52 passed (100%) / 0 failed / 0 errors`。
- 运行产物落点：`.test-output/evals/promptfoo_tests.yaml`、`.test-output/evals/promptfooconfig.yaml` 与 `.test-output/evals/promptfoo/` 存在；`config/evals/.generated` 不存在。
- override 预检：`make -n eval-offline EVAL_OUTPUT_DIR=/private/tmp/easyinterview-evals-review` 显示 generated tests、渲染后 config、Promptfoo state/logs 均指向同一 override 目录。
- 负向搜索：`Makefile`、`config/evals/promptfooconfig.yaml`、`.gitignore` 中不再保留 `config/evals/.generated` / `.generated/promptfoo` 路径。
- 合并收口：`docs/work-journal/2026-05-24.md` 与 `docs/work-journal/INDEX.md` 保留两侧事实记录并恢复时间顺序。
- 质量 gate：`validate_context.py`、focused Go tests、`make lint-ai-profile-coverage`、`make lint-prompts-hardcode`、`make docs-check`、`git diff --check`、`git diff --cached --check` 均已通过。

## 2 会话中的主要阻点/痛点

1. **Runtime artifact 被放进 config truth source**。
   - **证据**：`main` 上 `make eval-offline` 的 Promptfoo config/state/log 目录指向 `config/evals/.generated/promptfoo`，且 `.gitignore` 隐藏该目录。
   - **影响**：运行产物与 `config/evals/` 下的评估配置、cases、provider/assert 脚本混在同一棵配置树内，后续 review 很容易误判为配置资产或漏掉路径漂移。

2. **work-journal 合并冲突需要语义保留两侧事实**。
   - **证据**：冲突集中在 `docs/work-journal/2026-05-24.md` 与 `docs/work-journal/INDEX.md`；两侧分别包含 prompt-rubric 004 与 e2e-scenarios-p0 的真实提交记录。
   - **影响**：不能使用 `ours` / `theirs` 简化处理，必须按时间顺序合并日记，并让 INDEX 与最终日记事实一致。

3. **原 004 gate 没有禁止 config 下 runtime 产物目录**。
   - **证据**：004 plan/checklist 在 v1.2 已验证 Promptfoo runner 成功，但没有明确要求 generated tests、DB、logs 只能写入 `.test-output/evals/`，也没有禁止 `.gitignore` 隐藏 `config/evals/.generated/`。
   - **影响**：前一轮“隔离用户 home 目录”的修复解决了 home 污染，却把产物放进了 repo config 树。

## 3 根因归类

| 痛点 | 根因 | 归属 |
|------|------|------|
| Runtime artifact 写入 config truth source | eval runner 只要求隔离 Promptfoo state，未明确运行产物目录边界 | `spec-plan` |
| work-journal 冲突需语义合并 | merge 高冲突面是日志 / INDEX，必须按事实保留而非文本优先 | `no repo change needed` |
| 004 gate 未禁止旧路径 | checklist 缺少负向 grep 与 `.gitignore` 反向断言 | `spec-plan` |

## 4 对流程资产的改进建议

1. **运行产物目录 gate 已原地固化**。
   - **落点**：`prompt-rubric-registry/004-real-model-profile-and-evals` plan/checklist v1.3
   - **优先级**：high
   - **状态**：已完成，要求 Promptfoo generated tests/state/logs 只能位于 `.test-output/evals/`，并禁止写入或 gitignore `config/evals/.generated/`。

2. **保留 `.gitignore` 负向断言**。
   - **落点**：004 checklist / future eval plans
   - **优先级**：medium
   - **状态**：已完成于本计划；后续若新增 eval runner，应复用同类 `rg` gate，避免用 ignore 规则掩盖错误产物路径。

3. **work-journal 冲突按事实合并**。
   - **落点**：merge close-out 习惯 / 当前 AGENTS 已覆盖
   - **优先级**：low
   - **状态**：无需新增仓库规则；本次按现有约定完成。

## 5 建议优先级与后续动作

| 优先级 | 动作 | 目标资产 |
|--------|------|----------|
| P1 | 保持 004 v1.3 的 `$(EVAL_OUTPUT_DIR)` 产物 gate，后续 eval runner 不得重新引入 config-local runtime 目录，且覆盖 `EVAL_OUTPUT_DIR` 时 tests/config/state/logs 必须同源移动 | `docs/spec/prompt-rubric-registry/plans/004-real-model-profile-and-evals/checklist.md` |
| P2 | 下一次 prompt-rubric L2 review 若涉及 eval runner，优先复跑 `make eval-offline` + 旧路径负向 grep | `/plan-code-review prompt-rubric-registry/004-real-model-profile-and-evals` |
| P3 | e2e-scenarios-p0 的后续实施继续从 `001-full-funnel-happy-journey` owner plan 进入，不复用 Promptfoo runtime 路径作为场景产物目录 | `/implement e2e-scenarios-p0/001-full-funnel-happy-journey` |
