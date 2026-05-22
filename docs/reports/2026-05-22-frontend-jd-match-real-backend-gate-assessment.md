# Frontend JD Match Real Backend Gate 交付复盘报告

> **日期**: 2026-05-22
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`frontend-home-job-picks-and-parse/002-jd-match-recommendations` L2 remediation；原地关闭 frontend completed plan 仍停留在 fixture-backed / backend future 口径的问题。
- 关键修复：新增 `frontend/src/api/jdMatch.realApiMode.test.ts`；P0.027-P0.031 trigger/verify 接入 `VITE_EI_API_MODE=real` gate；更新 plan/spec/BDD/OpenAPI/generated spec/scenario docs；记录 [BUG-0085](../bugs/BUG-0085.md)。
- 成功证据：focused real-mode Vitest PASS；P0.027-P0.031 全部 setup→trigger→verify→cleanup PASS；backend P0.094-P0.097 全部 setup→trigger→verify→cleanup PASS；`make lint-openapi` PASS；二次 `make codegen-openapi` 后 generated diff 不再变化；`make codegen-check` PASS；`make docs-check` 与 `git diff --check` PASS。

## 2 会话中的主要阻点/痛点

- Completed plan 的历史文字误导了当前实现状态。
  - **证据**：plan §3.6 仍写 12 个 backend handler `not-yet-implemented`，scenario docs 仍写 no network，但 backend P0.094-P0.097 已有 live API proof。
  - **影响**：如果只看历史 checklist PASS，会把 fixture UI gate 误当作真实 API 闭环。
- Frontend scenario gate 没有表达“fixture UI variants + real API generated-client gate”的双层语义。
  - **证据**：P0.027-P0.031 trigger 原先只跑组件/fixture Vitest；新增 gate 后才显式证明 12 个 JobMatch operation 的 real base URL、credentials、IK 与 provenance。
  - **影响**：backend 完成后缺少跨 owner 反向接线证据。
- P0.030 verify 保留了旧 pixel gate 字符串。
  - **证据**：verify 仍 grep `toHaveScreenshot`，但当前 pixel spec 已切换为 non-empty screenshot buffer 以支持 clean checkout。
  - **影响**：场景验证会因旧断言名失败，掩盖真实当前 gate 已经通过。

## 3 根因归类

- Completed plan 没有 backend owner 完成后的反向 reconciliation gate。
  - **类别**：spec-plan
- Scenario README/trigger/verify 未明确区分 deterministic UI fixture variants 与 production generated-client routing proof。
  - **类别**：spec-plan
- Pixel gate 迁移后，下游 scenario verify 没有同步检查当前源码锚点。
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- 在 frontend-first plan 的 checklist 中固化“backend owner 完成后必须回到原 plan 补 real-mode gate”的复查项。
  - **落点**：spec-plan
  - **优先级**：high
- 对所有 fixture-first frontend scenario 采用固定 wording：UI variants remain fixture-backed；production generated-client gate must be separate and explicit。
  - **落点**：spec-plan
  - **优先级**：medium
- Scenario verify 引用测试实现时检查当前稳定源码锚点，不检查已废弃的断言 API 名称。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：在后续 frontend-first → backend-real handoff 的计划中，先找原 completed frontend owner plan 原地补 real-mode gate，不新建 sibling。
- 次优先级：审查其它仍写 `fixture-backed` / `not-yet-implemented` 的 frontend plan，确认对应 backend owner 是否已经完成，避免同类漂移。
- 可延后：抽象一个共享 scenario helper，统一注入 `VITE_EI_API_MODE=real` gate 与 verify grep，减少五个脚本重复维护。
