# openapi-v1-contract v1.8 artifact remediation 交付复盘报告

> **日期**: 2026-04-29
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付覆盖 `openapi-v1-contract` 三个原地 reopen plan 的 v1.8 artifact remediation：

- `001-bootstrap` Phase 7：`DELETE /api/v1/me` / `deleteMe` 契约、37 operation inventory、Go/TS generated API artefacts、P0 Debrief optional/hidden schema 口径。
- `002-fixtures-and-mock-source` Phase 5：`Auth/deleteMe` fixture、37 fixture/example coverage、Debrief default fixture P1 字段收口。
- `003-breaking-change-gate` Phase 5：`openapi-v1.0.0` baseline 原地重冻、diff inventory 37/37、privacy export 白名单边界保持不扩展。

成功证据：

- `make codegen-check`
- `make validate-fixtures`
- `make openapi-diff`（expected / baseline / current operations 均为 37，0 findings）
- `python3 -m unittest scripts.lint.openapi_inventory_test -v`
- `python3 -m unittest scripts.lint.validate_fixtures_test -v`
- `python3 -m unittest scripts.lint.validate_fixtures_cli_test -v`
- `python3 -m unittest scripts.codegen.render_openapi_fixture_examples_test -v`
- `python3 -m unittest scripts.lint.openapi_diff_test -v`
- `sync-doc-index --check` zero drift（仅既有 pending child 占位 warnings）
- `git diff --check`

## 2 会话中的主要阻点/痛点

- Bare `openapi-v1-contract` 入口需要人工消歧。
  - **证据**：同一 subspec 下存在 `001-bootstrap`、`002-fixtures-and-mock-source`、`003-breaking-change-gate` 三个 active remediation plan；本次需用户回复 `all` 后才按 001→002→003 顺序推进。
  - **影响**：入口选择多一次往返，且执行顺序实际来自 plan 依赖语义而非 manifest 机器可读关系。
- `make codegen-check` 与未提交 contract remediation 存在验证时序摩擦。
  - **证据**：`codegen-check` 用 `git diff --exit-code` 校验 committed baseline；在 001 Phase 7 中，contract/generated 的预期未提交变更会使该 gate 只能在 phase commit 后验证。
  - **影响**：需要先用 focused tests、`make lint-openapi`、`make codegen-openapi` 与提交后 `make codegen-check` 组合证明，无单一命令能覆盖 dirty-tree 阶段。
- 工作树出现与当前 OpenAPI delivery 无关的未提交修改。
  - **证据**：002 完成后 `git status` 显示 `config/README.md` 与 `docs/spec/repo-scaffold/plans/001-bootstrap/checklist.md` dirty；为满足 003 branch gate，先用 stash 暂存并在最终恢复。
  - **影响**：增加了分支切换风险和交付后状态说明成本；未影响 OpenAPI 提交内容。

## 3 根因归类

- Bare subspec alias 不足以表达多 plan 顺序执行。
  - **类别**：skill / spec-plan
- `codegen-check` 同时承担生成物漂移与 committed baseline 校验，dirty-tree remediation 阶段天然需要拆成提交前后两段证据。
  - **类别**：README / spec-plan
- 无关 dirty 文件属于共享工作树协作状态，不是本次 OpenAPI plan 的设计或实现缺陷。
  - **类别**：no repo change needed

## 4 对流程资产的改进建议

- 在 `/implement` 或 plan-context 层支持同一 subspec 下的 ordered remediation batch。
  - **落点**：skill / spec-plan
  - **优先级**：medium
- 在 OpenAPI README 或 001 checklist 中显式说明 dirty-tree remediation 的 codegen 证据组合：提交前跑 `make lint-openapi` + `make codegen-openapi` + focused tests，提交后跑 `make codegen-check`。
  - **落点**：README / spec-plan
  - **优先级**：low
- 保持对无关 dirty 文件的当前处理方式：不纳入计划提交，必要时短暂 stash 并在收尾恢复。
  - **落点**：no repo change needed
  - **优先级**：low

## 5 建议优先级与后续动作

- Medium：为同一 subspec 的多 active remediation plan 增加机器可读顺序或 `/implement all` 约定，减少后续多 plan artifact remediation 的入口摩擦。
- Low：补充 OpenAPI dirty-tree codegen-check 说明，避免未来误把预期未提交 generated diff 当成 gate 失败。
- Low：无关 dirty 文件本次已恢复到工作树，不建议为一次性状态修改治理文档。
