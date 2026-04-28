# openapi-v1-contract/001-bootstrap Remediation 交付复盘报告

> **日期**: 2026-04-28
> **审查人**: Codex

## 1 复盘范围与成功证据

本次复盘覆盖对 [openapi-v1-contract/001-bootstrap assessment](./2026-04-28-openapi-v1-contract-001-bootstrap-assessment.md) 的核验与修复。实际修复范围包括：

- R1：ADR-Q1 锁定 session cookie 字面量 `ei_session`，A4 config spec 明确不允许通过 env/config 改名。
- R2：把 B1 `ApiError` inner object 与 B2 `ApiErrorResponse` wire envelope 分离，并修复 Go / TS generated shape。
- R3 / R6：B2 spec 与 `openapi/README.md` 锁定 `@apidevtools/swagger-cli@4.0.4` deprecated-but-accepted validation 边界；随后根据 `make docs-openapi` 的 deprecated 输出，将本地 docs renderer 从 `redoc-cli@0.13.21` 迁移为 `@redocly/cli@2.30.1 build-docs`。
- R4：B2 spec、02-api-definition、03-db-definition 锁定 `ResourceType` / `JobType` API-facing 字面量。
- R5：归类为低价值 no-op，未修改 `.agent-skills/tdd/SKILL.md`。

通过的验证：

- `go test ./backend/cmd/codegen/openapi/...`
- `make lint-openapi`
- `make docs-openapi`
- `cd backend && go build ./...`
- `cd frontend && npx tsc --noEmit`
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
- `git diff --check`

`make codegen-check` 的 validate / inventory 部分通过；最终 `git diff --exit-code` 因本次存在未提交的预期 generated 变更返回非 0，已记录在 checklist Phase 5.4。

## 2 会话中的主要阻点/痛点

- `make codegen-check` 同时承担“重生成验证”和“工作树 clean gate”两种语义。
  - **证据**：本次在修复 generated artefacts 后运行该 target，OpenAPI validate 与 inventory 均通过，但最终因预期 diff 返回非 0。
  - **影响**：未提交 remediation 时不能把它作为单独的“生成结果正确”通过证据，必须补充 focused idempotency test 与 lint/build/tsc。

## 3 根因归类

- `ApiError` 命名歧义。
  - **类别**：spec-plan + codegen
  - B1 `ApiError` 原本代表 inner object，B2 OpenAPI 曾把同名 schema 用作 wire envelope，导致 TS client 与 Go generated DTO 对 `ApiError` 形状产生不同理解。
- `codegen-check` dirty-tree 限制。
  - **类别**：README / plan usage note
  - 当前目标仍适合作为提交前 clean gate；在未提交变更的 remediation 中，需要额外说明它的最终 diff failure 不等价于生成失败。
- `redoc-cli` 与 `@apidevtools/swagger-cli` 是两个独立弃用面。
  - **类别**：spec-plan + README
  - validator 可以暂时接受 deprecated `@apidevtools/swagger-cli@4.0.4`，但 docs renderer 不应继续使用已提示官方替代命令的 `redoc-cli@0.13.21`。

## 4 对流程资产的改进建议

- 在后续需要更频繁执行 generated remediation 时，可新增一个只做重生成 + lint、不做 `git diff --exit-code` 的 `make codegen-openapi-verify` 或文档化“dirty tree 验证组合”。
  - **落点**：`openapi/README.md` / B2 后续 plan
  - **优先级**：low

## 5 建议优先级与后续动作

本次 remediation 已完成，后续主路径仍是 `openapi-v1-contract/002-fixtures-and-mock-source` 与 `003-breaking-change-gate`。短期不需要再修改 skill；若后续同类 generated remediation 增多，再考虑补充 dirty-tree codegen 验证 target。
