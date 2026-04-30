# db-migrations-baseline/001-bootstrap 交付复盘报告

> **日期**: 2026-04-30
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖 `db-migrations-baseline/001-bootstrap`：baseline SQL migration、`backend/cmd/migrate` wrapper、`golang-migrate` runner、backfill ledger / manifest、migration lint、enum-source drift gate、privacy delete dry-run 与 plan lifecycle close-out。
- 关联计划已收口：[plan](../spec/db-migrations-baseline/plans/001-bootstrap/plan.md) / [checklist](../spec/db-migrations-baseline/plans/001-bootstrap/checklist.md) Header 均为 `completed`，21/21 checklist items 完成，`plans/INDEX.md` 已移入 Completed。
- 通过验证：
  - `go test ./backend/cmd/migrate ./backend/internal/migrations/... -count=1`
  - `python3 -m pytest scripts/lint/migrations_lint_test.py -q`
  - `python3 scripts/lint/migrations_lint.py --repo-root .`
  - clean pgvector Postgres 上 `DATABASE_URL=... APP_ENV=dev make migrate-check`
  - `APP_ENV=prod make migrate-down` fail-fast，并提示 `MIGRATE_DOWN_FORCE=1`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `git diff --check`
- SQL probes 验证 table count 32、outbox retry columns、AI typed columns、pending due / dashboard / B-Tree / ivfflat / GIN 索引存在且 explain 命中；privacy dry-run 覆盖用户、auth/session、AI、async/outbox、prompt/rubric 与 migration metadata 表组。

## 2 会话中的主要阻点/痛点

- 旧 active plan 缺少新质量门禁分类。
  - **证据**：实施入口发现 `plan.md` 缺 `## 3 质量门禁分类`，checklist 缺逐项 `验证:` 断言；必须先经用户确认补齐文档，再进入 TDD 实施。
  - **影响**：实现开始前增加一次文档修订和确认往返；修订本身不改变 scope，但暴露了旧计划与当前治理规则的漂移。
- 同仓库并行推进导致分支与 `dev` 位置变化。
  - **证据**：执行过程中 `dev` 先后包含 event-outbox commits，当前分支一度切到 `feat/event-and-outbox-contract-001-bootstrap-0430`；最终需要把 migration branch 重新放到当前 `dev` 顶端，并显式只 staging B4 close-out 文件。
  - **影响**：增加了提交风险和收尾成本；若直接 `git add .` 或在错误分支提交，会混入 CI / event-outbox / link-check 相关未完成改动。
- `make migrate-check` 的成功输出偏少。
  - **证据**：clean Postgres 验证时 make target 通过，但主要可见输出来自 lint；apply/down/up 与 backfill ledger 去重需要额外 SQL probes 与 work journal 记录补证据。
  - **影响**：验证是可靠的，但审计证据分散在命令、SQL probe 和日志中，后续执行者需要重复整理。
- Negative drift 仍依赖临时改写真实源文件。
  - **证据**：enum-source checksum drift 通过临时修改 `migrations/enum-sources.yaml` 验证失败路径，再恢复后重跑 lint。
  - **影响**：失败路径有效，但手工恢复要求高；未来 migration drift gate 增多时容易误留临时改动。

## 3 根因归类

- 质量门禁治理升级早于本 plan 修订。
  - **类别**：spec-plan
  - **说明**：`001-bootstrap` 创建时没有当前 TDD/BDD classification 规则；实施前补齐是正确路径，但说明 active plans 仍可能存在同类漂移。
- `/implement` / `/work-journal` 对并行 plan 推进的分支状态缺少明确收敛提示。
  - **类别**：skill
  - **说明**：当前 workflow 能通过人工谨慎 staging 避免混入别的 owner 文件，但没有一等提示来处理“dev 已被 sibling plan 快进、当前 branch 不是本 plan branch”的收尾状态。
- Migration smoke gate 缺少 machine-readable summary。
  - **类别**：README / tooling
  - **说明**：`make migrate-check` 已执行核心动作，但不输出迁移版本、table count、backfill dry-run/apply ledger count 等摘要，导致证据需要额外 probe 拼接。
- 临时漂移验证是当前 repo 常见模式。
  - **类别**：no repo change needed / tooling
  - **说明**：本轮已恢复现场并通过 `git diff --check`，不构成缺陷；但若同类 gate 继续扩张，可以提升为 repo-tracked helper。

## 4 对流程资产的改进建议

- 对 active plans 做一次质量门禁 sweep。
  - **落点**：spec-plan / `/plan-review`
  - **优先级**：medium
  - **建议**：检查 active plan 是否缺 `## 3 质量门禁分类`、BDD 不适用说明、checklist `验证:` 子句，避免 `/implement` 执行时才发现。
- 在 `/implement` 或 `/tdd` phase-commit 文档中补“并行 plan 快进 dev 后的分支收敛”说明。
  - **落点**：skill
  - **优先级**：medium
  - **建议**：当 `dev` 已包含 sibling plan commits，且当前分支与本 plan branch 不一致时，要求先确认 HEAD 基线、显式 path staging，并在 journal 中记录未纳入的 unrelated dirty files。
- 为 `make migrate-check` 增加可审计 summary 输出。
  - **落点**：README / tooling
  - **优先级**：medium
  - **建议**：在 target 成功后输出 current migration version、table count、backfill ledger dry-run/apply success counts、privacy dry-run retain 表组摘要，减少手工 SQL probe 整理。
- 为 migration drift negative checks 增加临时目录 helper。
  - **落点**：tooling
  - **优先级**：low
  - **建议**：在 `scripts/lint/migrations_lint_test.py` 或新 helper 中模拟 checksum drift / unregistered check / forbidden secret 字段，不直接修改 repo 工作树。

## 5 建议优先级与后续动作

- **Medium**：优先做 active plan quality-gate sweep，能直接减少下一轮 `/implement` 的文档阻塞。
- **Medium**：给 phase-commit / branch 收敛补规则，降低多 plan 并行时误提交别的 owner 文件的风险。
- **Medium**：增强 `make migrate-check` summary，让 B4 migration gate 以后自带审计证据。
- **Low**：migration negative drift helper 可延后；当前 pytest 已覆盖大部分静态失败路径，手工 e2e drift 只是收尾成本偏高。
