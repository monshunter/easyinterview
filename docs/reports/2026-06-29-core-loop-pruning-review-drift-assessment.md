# Core Loop Pruning Review Drift 交付复盘报告

> **日期**: 2026-06-29
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖 `product-scope/001-core-loop-module-pruning` L2 review 后的修复：清理 stale `ui-design/canvas.html` Profile / Debrief 正向画板、删除 privacy runner 旧 `ProfileData` hook、修正 backend README domain drift，并在原 owner plan/checklist 中固化 v1.1 gate。
- 关联 Bug： [BUG-0128](../bugs/BUG-0128.md)。
- 成功证据：
  - `node --test ui-design/ui-design-contract.test.mjs` PASS (28 tests)。
  - `go test ./backend/cmd/api ./backend/internal/privacy/runner -count=1` PASS。
  - `make docs-check` PASS。
  - `git diff --check` PASS。
  - `ui-design/canvas.html` retired profile/debrief artboard grep 无命中。
  - privacy runner / backend README legacy hook grep 仅命中测试里的 `FieldByName("ProfileData")` 防回归断言。

## 2 会话中的主要阻点/痛点

- 完成态 zero-reference gate 漏掉了 `ui-design/canvas.html`。
  - **证据**：原 `ui-design` contract tests 已 PASS，但 L2 review 仍在 canvas 画板总览中发现 `route="profile"`、`route="debrief"` 和旧 Profile / Debrief section。
  - **影响**：静态原型 truth source 会让已删除模块以设计画板形式回流。

- privacy boundary gate 没有覆盖 handler option surface。
  - **证据**：DB、OpenAPI、backend profile package 已删除，但 `PrivacyDeleteHandlerOptions.ProfileData` 仍可被 runtime caller 重新注入。
  - **影响**：候选人画像 cleanup 逻辑虽然当前没有在 `cmd/api` 传入，却仍保留为运行时代码扩展点。

- active owner README 被旧领域名污染。
  - **证据**：`backend/README.md` 仍列出 `profile` 作为后端领域模块。
  - **影响**：后续后端 owner 可能把已退役模块视为当前 workstream。

## 3 根因归类

- `canvas.html` 漏扫：
  - **类别**：spec-plan
  - **根因**：plan 的 UI source structure parity gate 强调 `ui-design/src` 和正式前端 route，但没有明确可见 canvas overview 也属于 active artifact。

- handler option surface 漏扫：
  - **类别**：spec-plan / skill
  - **根因**：plan-code-review 的 runtime audit 已覆盖 production wiring，但删除型 owner plan 没有把“public option / interface extension point”列成 retired-domain 检查面。

- README drift：
  - **类别**：README / spec-plan
  - **根因**：zero-reference gate 主要按 runtime path 分类，active README 被当成低风险文档而不是 owner contract。

## 4 对流程资产的改进建议

- 删除型 plan 的 Coverage Matrix 增加 `prototype overview / canvas / owner README / public option surface` 四类 artifact。
  - **落点**：spec-plan
  - **优先级**：high

- 在 plan-code-review 的删除型 review checklist 中增加：对保留 package 的 exported options/interfaces 做 retired-domain grep。
  - **落点**：skill
  - **优先级**：medium

- 对高碰撞词如 `profile`，计划内保持 allowed-hit 分类，同时要求 active README 中避免旧业务域名。
  - **落点**：spec-plan / README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高价值：后续 module pruning / hard-delete 计划先把 `ui-design/canvas.html`、owner README、exported options/interfaces 纳入 zero-reference gate。
- 次高价值：在 `/plan-code-review` 删除型 review 输出中单列 `public extension point` 证据，避免只审 production caller。
