# repo-scaffold/001-bootstrap 交付复盘报告

> **日期**: 2026-04-26
> **审查人**: Claude

## 1 复盘范围与成功证据

- 范围：A1 `repo-scaffold` 唯一 plan `001-bootstrap` 的 W0 仓库脚手架落地（plan 三件套：[plan](../spec/repo-scaffold/plans/001-bootstrap/plan.md) / [checklist](../spec/repo-scaffold/plans/001-bootstrap/checklist.md) / [context](../spec/repo-scaffold/plans/001-bootstrap/context.yaml)），交付 7 个根容器目录 + 顶层 Makefile + `.editorconfig` + `.gitignore` + `.tool-versions` + `scripts/git-hooks/` + `scripts/bootstrap.sh` 与全部验证证据。
- 成功证据：
  - Checklist 11/11 全部勾选；plan + checklist Header `状态` 已切到 `completed`，`docs/spec/repo-scaffold/plans/INDEX.md` 已迁组。
  - Phase 3.1：在 macOS zsh + `SHELL=/bin/bash` 下 `make help` / `fmt` / `lint` / `test` / `build` / `dev-up` / `dev-down` / `codegen` / `migrate` 9 条命令全部 `exit 0`，缺失子 Makefile 时按 `recurse_target` 占位规则打印 `TODO: <target> implemented by <child> child subspec`。
  - Phase 3.2：`make install-hooks` 三轮（首次安装、已有 symlink rerun、删除后 rerun）后 `.git/hooks/{pre-commit,commit-msg}` 均为指向 `scripts/git-hooks/` 的 absolute symlink，验证幂等。
  - Phase 3.3：`validate_context.py --target repo` 解析 plan / checklist / spec 通过；`sync-doc-index.py --check` 报告 0 violations / 0 drifts / 0 orphans，36 个 `_pending_` warning 与 W0 spawn 后的占位行数一致。
  - Git 流水：feature branch `feat/repo-scaffold-001-bootstrap-0426` 4 个 commit（Phase 1 / Phase 2 / Phase 3 verify / lifecycle close）通过 `git merge --ff-only` 全部 ff-merge 回父分支 `dev`。
  - Spec / engineering-roadmap parent checklist 未被改动，A1 W0 收口边界与 §6 验收 C-1 至 C-4 一致。

## 2 会话中的主要阻点/痛点

- **父分支语义在 dev trunk 场景下含糊**
  - **证据**：`detect_session_branch.py` 报告 `currentBranch=dev` / `matchesSessionBranch=false`；`AGENTS.md §7` 写「默认父分支: 仓库默认分支（优先自动探测；若未配置则使用当前主开发分支）」，但 `git symbolic-ref refs/remotes/origin/HEAD` 返回 `origin/main`，而所有 spec / plan 文档只在 `dev` 上（`git log main..dev` 列出 15 个 spec/governance commit）。
  - **影响**：会话内需要额外一轮 git log 比对 + 文字推理，确认从 `dev` 而非 `main` 拉 feature branch；如果按字面「优先自动探测」走 `main`，feature branch 将看不到本 plan 的 plan/checklist/spec 文档，无法实施。
- **`sync-doc-index --fix-index` 只改字段不迁分组**
  - **证据**：在 plan / checklist Header 由 `active` → `completed` 后跑 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-index`，输出 `1 file(s) modified ... 状态: active → completed`；紧接的 `Post-fix Verification` 仍报告 `2 drifts ... 状态(group) header=completed index=active`，提示 `auto-fixable: 0, needs LLM: 2`。
  - **影响**：必须额外手工编辑 `docs/spec/repo-scaffold/plans/INDEX.md`，把 `001-bootstrap` 行从 `## 1 进行中（Active）` 段迁到 `## 2 已完成（Completed）` 段，才能让 `--check` 重新归零。后续每个 plan 的 lifecycle close 都会重复一次相同的手工动作。

## 3 根因归类

- AGENTS.md §7 父分支判定优先级与「dev 是真实活动分支、`origin/HEAD` 仍然指向 `main`」的常见状态不匹配；句子里「优先自动探测」被解读为「auto-detect 总是赢」，与「使用当前主开发分支」之间缺少明确的 fallback 触发条件。
  - **类别**：AGENTS.md
- `sync-doc-index --fix-index` 当前实现只覆盖单元格级别的 `状态 / 版本 / 更新日期` 字段同步，未覆盖「行所在的 group 段标题」迁移；当 plan 跨越 active → completed 时这是 lifecycle close 必然触发的场景，不是一次性执行错误。
  - **类别**：skill（`.agent-skills/sync-doc-index/`）

## 4 对流程资产的改进建议

- 收紧 AGENTS.md §7 父分支判定语义。给出 3 条明确判定优先级（建议顺序）：1) `context.yaml metadata.baseBranch` 有显式值则用之；2) 若 plan / spec 文档只存在于某条非默认分支上，则那条分支是父分支（避免 feature branch 拉不到 plan 文档）；3) 否则按 `git symbolic-ref refs/remotes/origin/HEAD` / `init.defaultBranch` 自动探测。同时举一个例子说明「`origin/HEAD=main` 但 plan 仅落在 `dev`」时应当走第 2 条。
  - **落点**：AGENTS.md
  - **优先级**：medium
- 扩展 `sync-doc-index.py --fix-index` 覆盖 group-section 迁移：当某 INDEX.md 的行 `状态 == X` 与该行所属 `## N <Group>` 段不匹配时，自动将该行迁到对应 group 段（active → 进行中、completed → 已完成、deprecated → 废弃 等），group 段缺失时按既有 README / 模板格式追加。需要同时考虑 `docs/spec/INDEX.md`、各 subspec 的 `plans/INDEX.md`、reports / bugs INDEX 等多种索引体例。
  - **落点**：skill（`.agent-skills/sync-doc-index/`）
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮最值得实施：**扩展 `sync-doc-index --fix-index` 支持 group-section 迁移**。原因：W0 之后会有 38 份 child subspec 与至少同样数量的 plan，逐个 lifecycle close 都会触发同一手工迁移动作，自动化收益高且边界清晰。
- 次优：**收紧 AGENTS.md §7 父分支判定**。原因：`/implement` 在 dev trunk 场景下的判断逻辑会被未来每个 child subspec 复用，统一文字阐述能避免重复推理。
- 可以延后：本次会话的一次性执行小问题（首次 verify Phase 3.1 时 bash for-loop 解释成单参数），属于一次性 shell 操作偏差，不构成流程缺陷，不需要资产改动。
