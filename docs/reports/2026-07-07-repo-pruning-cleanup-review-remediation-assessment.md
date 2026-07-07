# Repo Pruning Cleanup Review Remediation 交付复盘报告

> **日期**: 2026-07-07
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖 `fix/repo-pruning-cleanup-0706` 相较于 `main` 的 review remediation：repo pruning lint false-negative、workspace resume picker stale id guard、pnpm 11 build approval 配置、gofmt 漂移，以及 stricter pruning gate 曝光的 active 文本残留。
- 成功证据：
  - `python3 -m pytest -q scripts/lint/core_loop_pruning_surface_test.py`
  - `python3 -m pytest -q scripts/lint/core_loop_pruning_surface_test.py scripts/lint/makefile_dry_run_test.py`
  - `make lint-core-loop-pruning-surface`
  - `CI=true pnpm install --config.confirmModulesPurge=false --fetch-timeout=300000`
  - `CI=true pnpm --filter @easyinterview/frontend test src/app/screens/workspace/modals/ResumePickerModal.test.tsx`
  - `CI=true pnpm --filter @easyinterview/frontend typecheck`
  - `gofmt -l $(git diff --name-only -- '*.go')`
  - `make docs-check`
  - `git diff --check`

## 2 会话中的主要阻点/痛点

- Pruning lint 旧 gate 存在 false-negative。
  - **证据**：新增生产路径反例测试后，旧实现把 `backend/internal/api/debriefs/handler.go` 中带 `delete` 的 debrief 残留归入 `negative_tests`。
  - **影响**：non-current runtime 残留可以通过措辞逃过 repo-wide gate，降低 core-loop pruning 的 zero residual 可信度。
- Frontend focused test 首次被 pnpm 配置阻塞。
  - **证据**：前端 Vitest 执行前 pnpm 11 报 `[ERR_PNPM_IGNORED_BUILDS]`，原因是 `pnpm-workspace.yaml` 的 `allowBuilds` 仍为占位值。
  - **影响**：测试代码尚未运行就失败，容易把环境配置问题误判为前端用例问题。
- Resume picker 缺少 stale bound id 回归测试。
  - **证据**：review 指出 `boundResumeId="resume-unbound"` 能启用 confirm；补充 focused Vitest 后覆盖该场景。
  - **影响**：workspace flow 可能向上层提交不存在于 active resume 列表的 id。

## 3 根因归类

- Pruning lint false-negative：
  - **类别**：spec-plan
  - **说明**：core-loop pruning owner gate 已要求 zero residual，但脚本测试缺少 production path negative wording 反例。
- pnpm allowBuilds 占位阻塞：
  - **类别**：README
  - **说明**：当前前端 README / dev workflow 对 pnpm 11 approval 配置没有显式提醒，配置漂移只在执行 test/install 时暴露。
- Resume picker stale id guard：
  - **类别**：no repo change needed
  - **说明**：属于具体组件 guard 缺口，已用 focused regression test 和实现修复闭环。

## 4 对流程资产的改进建议

- 在 `product-scope/001-core-loop-module-pruning` 的后续 gate 说明中补充：repo-wide pruning lint 必须包含 production path + negative wording 反例。
  - **落点**：spec-plan
  - **优先级**：medium
- 在 frontend 或 root dev workflow 中记录 pnpm 11 `allowBuilds` 必须使用布尔值，并提示 `package.json` 中旧 `pnpm.onlyBuiltDependencies` 已不再被 pnpm 读取。
  - **落点**：README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 优先补齐 owner plan / checklist 中的 pruning lint semantic negative fixture 说明，避免未来再次把 negative wording 当成全局 allowlist。
- 其次清理或迁移 `package.json` 中被 pnpm 11 忽略的 `pnpm.onlyBuiltDependencies`，减少前端 gate 的配置噪音。
