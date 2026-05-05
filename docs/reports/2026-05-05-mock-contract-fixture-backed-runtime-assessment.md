# Mock Contract Fixture-backed Runtime 交付复盘报告

> **日期**: 2026-05-05
> **审查人**: Codex

## 1 复盘范围与成功证据

- 交付范围：`mock-contract-suite/001-fixture-backed-mock-runtime` 的 `tooling` target，覆盖 fixture registry、frontend mock transport、backend mock handler / seed profiles、local drift gates 与 frontend-shell handoff。
- 成功证据：
  - `docs/spec/mock-contract-suite/plans/001-fixture-backed-mock-runtime/checklist.md` 全部 8 项已完成，plan / checklist 状态均为 `completed`。
  - `make lint-mock-contract` 通过，覆盖 `validate-fixtures`、`lint-openapi` 与 mock runtime boundary lint。
  - `make codegen-check`、`make docs-check`、backend mockruntime focused tests、frontend focused test/typecheck、`make test`、`make build`、`make lint` 均通过。
  - Phase commits 已 ff-only 合回 `dev`：`c91fe5b`、`8fc7f55`、`4f42d1b`、`bcb1d32`。

## 2 会话中的主要阻点/痛点

- 顶层 `make lint` 在最终验证阶段被非本 plan 的 active AI provider docs 旧术语阻塞。
  - **证据**：`make lint` 首次失败于 `lint-ai-provider-terminology`，命中 `docs/spec/ai-provider-and-model-routing/plans/002-tools-streaming-and-stt/{plan.md,checklist.md}` 中的旧 `gateway` 术语。
  - **影响**：mock-contract 已通过自身 gate 后，仍需额外修复跨 plan active docs 漂移才能完成聚合 lint。
- 前端 mock transport 需要 OpenAPI-derived route metadata，但既有 TS generated client 只导出 operationId。
  - **证据**：2.1 Red test 通过 generated client 发请求时，需要 method/path -> operationId 匹配；最终通过扩展 `openapi/templates/ts/client.tmpl` 导出 `ALL_ROUTES` 解决。
  - **影响**：若后续前端 workstream 手写 route table，容易再次产生 route/operation 漂移。

## 3 根因归类

- Active docs 旧术语漂移：
  - **类别**：spec-plan。
  - **根因**：`ai-provider-and-model-routing/002-tools-streaming-and-stt` 是 draft/blocked plan，但仍属于 active-scope lint 扫描面；该 plan 的 negative-search 文案自身包含 retired terminology。
- TS route metadata 缺口：
  - **类别**：README / no repo change needed。
  - **根因**：B2 TS client 初始只服务 fetch 调用，不需要 mock transport 的反向 route lookup；本次已通过 generated `ALL_ROUTES` 与 README handoff 固化消费方式。

## 4 对流程资产的改进建议

- 在后续 `/plan-review --fix ai-provider-and-model-routing/002-tools-streaming-and-stt` 或激活该 plan 前，显式运行 `make lint` 或至少 `make lint-ai-provider-terminology`。
  - **落点**：spec-plan。
  - **优先级**：high。
- 后续 frontend-shell / D2-D6 plan 应直接消费 `frontend/src/api/README.md` 的 mock runtime handoff，不应手写 operation route inventory。
  - **落点**：README。
  - **优先级**：medium。
- 后续 mock runtime 扩展新 scenario 时，先改 `openapi/fixtures`，再跑 `make lint-mock-contract`，避免把 seed state 写进前端或后端 runtime。
  - **落点**：README。
  - **优先级**：medium。

## 5 建议优先级与后续动作

- **最高优先级**：进入 `frontend-shell/001-app-shell-auth-settings frontend` 实施前，按已写入的 mock runtime handoff 先接 `createFixtureBackedFetch`，证明 auth shell 可以完全基于 B2 fixtures 开发。
- **次优先级**：在 AI provider 002 真正激活前做一次 L1/L2 preflight，避免 draft plan 的 active-scope 文案再次阻塞全局 lint。
