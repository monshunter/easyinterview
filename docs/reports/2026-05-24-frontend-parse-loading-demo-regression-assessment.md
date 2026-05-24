# Frontend Parse Loading Demo Regression 交付复盘报告

> **日期**: 2026-05-24
> **审查人**: Codex

## 1 复盘范围与成功证据

本次复盘范围是 `frontend-home-job-picks-and-parse/001-home-jd-import-and-parse` 的 parse loading 回归修复：上传、粘贴或 URL 导入 JD 后，即使首次 `getTargetJob` 已返回 `analysisStatus=ready`，正式前端也必须先展示 `ui-design` 4 步 loading 演示，再进入 parsed preview。

已通过的成功证据：

- `pnpm --filter @easyinterview/frontend test src/app/screens/parse/ParseFlow.test.tsx`：6 tests PASS。
- `pnpm --filter @easyinterview/frontend test src/app/screens/parse`：5 files / 27 tests PASS。
- `pnpm --filter @easyinterview/frontend test`：213 files / 1297 tests PASS。
- `pnpm --filter @easyinterview/frontend build`：PASS。
- `E2E.P0.015` setup → trigger → verify → cleanup：PASS。
- `validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check`：PASS。
- [BUG-0099](../bugs/BUG-0099.md) 已建档；owner spec v1.10、plan/checklist v1.3、BDD checklist 与 INDEX 已同步 ready-response loading regression gate。

## 2 会话中的主要阻点/痛点

### 2.1 旧测试把错误行为写成了期望

- **证据**：`ParseFlow.test.tsx` 原 ready fixture 测试断言 ready response 后立即出现 preview；这与 `ui-design/src/screens-p0-complete.jsx::ParseScreen` 的 4 步 loading 初始流程冲突。
- **影响**：先要把测试翻成 regression red case，再修实现；否则单看测试绿灯会误判为当前行为正确。

### 2.2 owner spec 对 ready response 的过渡边界不够硬

- **证据**：D-2 已规定前端只轮询 `analysisStatus` 且不直连 LLM，但未明确首次 ready 也不能跳过 loading demo。
- **影响**：实现容易把 `ready` 理解为“立即 preview”，导致真实 backend 解析快时用户看不到设计里的解析演示流程。

### 2.3 preview 相关测试隐式依赖立即进入 preview

- **证据**：`ParseEdit`、`ParseAuthGate`、`ParseFailedState` 中多个 preview 测试需要新增 fake-timer helper，先推进 3200ms loading gate 后再断言 preview 交互。
- **影响**：修复核心 flow 后，邻近测试需要同步表达新状态机，避免未来测试继续绕过 loading stage。

## 3 根因归类

- **spec-plan**：ready response loading gate 没有在 D-2、BDD checklist 和 plan checklist 中明确固化。
- **skill / review habit**：L2 review 需要把用户可见中间态也作为 UI parity gate 的一部分，而不是只比对最终页面。
- **no repo change needed**：全量前端测试输出仍有既有 `act(...)` warning，但本次 focused parse suite 已消除新增 warning；现有 warning 不影响本次修复结果。

## 4 对流程资产的改进建议

- 在后续 frontend UI parity plan 中，把“过渡态 / loading stage / skeleton / empty-to-ready transition”列为必须反查的 DOM 与 timer gate。
  - **落点**：spec-plan
  - **优先级**：high

- 在 `/plan-code-review` 的 frontend parity 检查口径中补一句：不得只审最终 state，必须覆盖用户可见状态机的进入顺序和最短停留时间。
  - **落点**：skill
  - **优先级**：medium

- 对 E2E.P0.015 增加浏览器级截图或 trace 证据，专门证明 ready response 不会跳过 loading demo。
  - **落点**：scenario / README
  - **优先级**：medium

## 5 建议优先级与后续动作

最高优先级是把“ready response 不能跳过 loading demo”继续下沉到场景级可视证据：当前 unit/Vitest gate 已覆盖状态机，但 E2E.P0.015 更适合长期捕捉真实 import→parse 路径里的过渡态回归。

下一步建议由 `/scenario-create` 或 `/scenario-run E2E.P0.015` owner 承接，给 P0.015 增加 ready-response loading DOM/screenshot 断言；备选路径是先用 `/plan-code-review frontend-home-job-picks-and-parse/001-home-jd-import-and-parse --fix` 做一次更完整的 browser parity 复核。
