# Engineering Roadmap Plan Code Review Alignment 交付复盘报告

> **日期**: 2026-05-04
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付覆盖 `engineering-roadmap/001-decompose-subspecs` 的 L2 plan-code-review remediation：对齐 engineering roadmap 关联的 active specs、plans、ADRs、OpenAPI / DB / events / config / AI gateway 契约、UI 设计文档与静态原型事实，移除旧 wave / route / Mistakes / Growth / Drill / onboarding 执行口径残留，并修复 codegen conventions 测试仍期望旧 `PracticeMode=core_interview` 的漂移。

二次检查继续覆盖仓库入口 README 与本地 dev/test/deploy 说明，修复根 README、backend/frontend/shared/migrations/test/deploy/config/openapi fixtures 等文档中仍残留的旧 wave / pending child / 独立错题本 / 不存在 `test/scenarios/` 路径说明，使入口文档与 roadmap v3 的 active spec + on-demand workstream 口径一致。

已通过的证据：

- `make docs-check`
- `make codegen-check`
- `make lint-openapi`
- `make validate-fixtures`
- `make lint-events`
- `make lint-config`（gitleaks 本地未安装，脚本按现有策略跳过第二层扫描）
- `python3 scripts/lint/migrations_lint.py --repo-root .`
- `cd backend && go test ./internal/migrations ./cmd/migrate ./cmd/codegen/conventions -count=1`
- `node --test ui-design/ui-design-contract.test.mjs`
- `python3 -m pytest scripts/lint/migrations_lint_test.py scripts/lint/events_inventory_test.py scripts/lint/lint_events_test.py -q`
- `make test`
- `make build`
- `git diff --check`
- `python3 scripts/lint/check_md_links.py .`
- 额外 `docs/spec` 内部锚点审计 `TOTAL 0`

## 2 会话中的主要阻点/痛点

- 普通 `docs-check` 能发现文件链接与 Header / INDEX drift，但不覆盖 Markdown heading anchor 漂移。
  - **证据**：额外锚点审计最初发现 6 个旧锚点，修复后为 `TOTAL 0`。
  - **影响**：旧 roadmap 章节重命名后，部分跨 spec / plan 深链会静默失效。

- 可执行测试期望没有随 `shared/conventions.yaml` 的当前枚举同步。
  - **证据**：`cd backend && go test ./cmd/codegen/conventions -count=1` 曾失败于仍期望 `PracticeMode = "core_interview"`；修复后通过。
  - **影响**：codegen package 的单测与当前 B1 truth source 不一致，可能掩盖后续 enum drift。

- 历史 remediation 证据与当前执行口径混在同一 plan 正文时，容易把负向断言误读为当前要求。
  - **证据**：OpenAPI 001/002 plan 正文中仍出现当前副作用 endpoint 列表包含旧 `/mistakes/{mistakeId}/retest`、fixture provenance 扫描仍包含独立 `MistakeEntry`。
  - **影响**：后续实现者可能按旧 endpoint/schema 恢复已删除模块。

- 仓库入口 README 没有被 `docs-check` 覆盖，仍可能保留旧 roadmap 形态或指向不存在的路径。
  - **证据**：二次检查发现根 `README.md` 仍描述“6 层 38 child / 6 wave”、`migrations/README.md` 仍写“待 W1 spawn”、`test/README.md` 与 `deploy/dev-stack/README.md` 指向当前仓库不存在的 `test/scenarios/`。
  - **影响**：新实现者会从入口文档进入旧执行模型，或按不存在的测试框架路径执行操作。

## 3 根因归类

- 锚点漂移属于 `spec-plan` + tooling gap：现有 docs gate 验证链接存在，但未验证 fragment anchor 是否存在。
- codegen 测试期望漂移属于 `spec-plan` + test gap：shared truth source 更新后，package-level tests 仍保留旧字面量断言。
- plan 正文残留属于 `spec-plan` gap：完成态 plan 的 remediation phase 追加后，正文前半段未完全 rebaseline。
- 入口 README 漂移属于 README + tooling gap：默认 `make docs-check` 只检查 `docs/`，不会覆盖根 README 与仓库子目录 README。

## 4 对流程资产的改进建议

- 将 `docs/spec` fragment anchor 审计沉淀为可复用脚本，并纳入 `make docs-check`。
  - **落点**：README / Makefile / scripts
  - **优先级**：high

- 为 codegen conventions 测试增加“从当前 `shared/conventions.yaml` 动态取第一项 / 当前项”的断言，减少硬编码旧枚举值。
  - **落点**：spec-plan / test
  - **优先级**：medium

- 对完成态 plan 的 remediation phase 增加一条审查 checklist：新增 remediation 后必须重新扫描 plan 正文中的旧 current-scope 句子。
  - **落点**：skill 或 docs/spec README
  - **优先级**：medium

- 将根 README 与关键子目录 README 纳入轻量链接 / 旧 wave 术语审计，至少覆盖 `_pending_`、`待 W* spawn`、旧 `docs/spec/INDEX.md#*-layer-*` 锚点和不存在路径。
  - **落点**：README / Makefile / scripts
  - **优先级**：medium

## 5 建议优先级与后续动作

最高价值的后续动作是把 fragment anchor audit 变成正式 `docs-check` 子步骤；这能直接防止 roadmap/spec 章节重命名后的静默深链失效。

其次是收敛 codegen 测试的硬编码枚举期望，让测试自然跟随当前 truth source，而不是在 product-scope 改动后再次留下旧字面量。

第三个动作是把仓库入口 README 纳入同一套漂移门禁，避免 `docs/` 已对齐但开发者第一眼看到的入口仍停留在旧 roadmap。
