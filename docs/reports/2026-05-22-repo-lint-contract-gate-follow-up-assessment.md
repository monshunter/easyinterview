# Repo Lint Contract Gate Follow-up 交付复盘报告

> **日期**: 2026-05-22
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付范围：在 `/plan-code-review local-dev-stack/001-bootstrap repo --fix` 修复 Postgres 18 dev stack 后，继续处理 repo-wide lint 暴露的 practice voice、JD Match rubric、mock runtime boundary、runtime topology wording 和 Go revive 命名阻断，并按 owner spec / plan 原地同步。
- Lint gate 修复：`backend_practice_legacy` 和 `mock_runtime_boundary` 不再把当前 practice voice MVP 的 `createPracticeVoiceTurn`、`/voice-turns`、`PracticeVoiceTurn*` 误判为退役独立 Voice 入口；`rubric_lint` 接受 JD Match F3 当前维度；runtime topology active docs 去除旧 shorthand；Go 内部命名满足 revive。
- 本地栈收口：经用户确认后仅删除并重建 `easyinterview-pg-data`，Redis / MinIO 卷保留；`make dev-up` 后 Postgres / Redis / MinIO 均 healthy，`make dev-doctor` 通过，Postgres `select 1` 通过。
- 数据库验证：使用 `DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable make migrate-up` 完成迁移；`migrate-status` 显示 `version=9 dirty=false`；`schema_migrations` 当前最高版本为 9。
- 静态与测试证据：`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`、`make docs-check`、`git diff --check`、`make lint`、`make test` 均通过；`make test` 覆盖 backend Go tests 与 frontend Vitest 212 files / 1296 tests。

## 2 会话中的主要阻点/痛点

- `make lint` 的 fail-fast 行为使多个历史 gate drift 串行暴露。
  - **证据**：practice voice retired-route false positive 修复后，才继续暴露 JD Match rubric allowlist、mock runtime boundary、runtime topology wording 和 Go revive 命名问题。
  - **影响**：单个 completed plan 的 L2 remediation 被迫跨多个 owner 补齐，验证时间和上下文切换成本明显增加。
- 退役口径 lint 对短 token 采用 broad matching。
  - **证据**：`practice.voice.*`、`/voice-turns`、`voice-turn://...`、`PracticeVoiceTurn*` 和合法 Voice turn 文本被当作退役独立 Voice route/tag。
  - **影响**：当前产品能力会被历史清理 gate 阻断，执行者容易误判为应删除 practice voice MVP，而不是修正 lint 精度。
- 新增 rubric owner plan 没有把共享 allowlist 同步列为硬 gate。
  - **证据**：JD Match F3 rubric 文件已经存在，`rubric_lint.py` 与 `config/rubrics/README.md` 的维度列表却没有同步。
  - **影响**：局部 plan 可以完成，但 repo-wide lint 在后续无关任务中才失败。
- Completed plan 原地修订需要同步多层生命周期字段。
  - **证据**：practice voice MVP、backend jobs recommendations、mock contract suite、backend async runner 均需要 spec / plan / checklist / history / INDEX 对齐。
  - **影响**：如果只改脚本不改 owner docs，会制造下一轮 review 的 false-green。

## 3 根因归类

- Retired-token lint 缺少“当前 owner 复用旧词根”的正向 fixture。
  - **类别**：spec-plan / test
- Rubric 维度新增缺少跨 owner allowlist 同步 gate。
  - **类别**：spec-plan
- Repo lint 审查没有一次性收集所有 blocker 的辅助入口。
  - **类别**：skill
- React `act(...)` warnings 仍在 frontend test 输出中出现，但 exit code 和 assertions 全部通过，本次不归入功能缺陷。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 在涉及退役模块清理的 lint owner checklist 中加入正向 fixture 要求：凡当前 owner contract 合法包含旧词根，必须新增 allow-case test，不能只写 negative search。
  - **落点**：spec-plan / test
  - **优先级**：high
- 在新增或修改 rubric 的 plan gate 中固化三件套：rubric file、`config/rubrics/README.md` allowlist、`rubric_lint` regression test 同步完成。
  - **落点**：spec-plan
  - **优先级**：high
- 为 L2 remediation 增加一个可选 repo lint audit runner，允许继续执行并汇总所有 lint target 失败，再进入 owner-by-owner 修复。
  - **落点**：skill / tooling
  - **优先级**：medium
- 保留现有 owner plan 原地修订策略：命中 completed plan 后先回到当前 owner，同步 spec / plan / checklist / history / INDEX，再恢复 completed。
  - **落点**：AGENTS.md / skill
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮最值得实施：给 `plan-code-review` 或 repo lint 工具补一个“continue-on-error audit mode”，在 L2 remediation 前一次性列出全部 lint blocker，减少串行发现成本。
- 同步可做：在 mock-contract-suite、practice-voice-mvp、backend-jobs-recommendations 的后续计划模板中保留 lint precision / allowlist sync / regression fixture 的显式 gate。
- 可延后：React `act(...)` warnings 可在前端专项清理中处理；当前 `make test` 已通过，不阻塞本次交付。
