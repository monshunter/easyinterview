# Frontend Shell TopBar UI-Design Parity 交付复盘报告

> **日期**: 2026-05-08
> **审查人**: Codex

## 1 复盘范围与成功证据

- 交付主题：`frontend-shell/002-app-shell-visual-system --fix` 的 TopBar 源级 parity remediation。用户指出正式实现与 `ui-design` 截图差异明显后，本轮修复把正式 TopBar 从 native select / 独立 custom accent popover 改回 `ui-design/src/app.jsx` 的 theme menu、Custom row 内嵌 AccentPicker、icon-only dark toggle 与 language toggle。
- 修复范围：`frontend/src/app/topbar/TopBar.tsx`、`topbar.css`、TopBar/i18n/scenario/pixel-parity 测试、`frontend-shell` spec/history、001/002 plan/checklist/BDD、frontend README 和受影响 E2E scenario 文案。
- 成功证据：
  - Focused TopBar：`pnpm --filter @easyinterview/frontend test src/app/topbar/TopBarVisual.test.tsx src/app/topbar/TopBar.test.tsx` PASS（20 tests）。
  - 全量前端：`pnpm --filter @easyinterview/frontend test` PASS（40 files / 237 tests）。
  - 类型与构建：`pnpm --filter @easyinterview/frontend typecheck` PASS；`pnpm --filter @easyinterview/frontend build` PASS；`make build` PASS。
  - 真实浏览器：`pnpm --filter @easyinterview/frontend test:pixel-parity` PASS（46 tests，desktop + mobile）。
  - BDD 场景：`E2E.P0.004` setup→trigger→verify→cleanup PASS，日志包含 `language toggle`、English copy 与 `Accept-Language: en`。
  - 文档：`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-index` 后 `All documents are in sync`；`make docs-check` zero drift。
  - Bug 记录：新增 [BUG-0021](../bugs/BUG-0021.md)，记录 TopBar 源级结构漂移、gate 缺口与修复验证。

## 2 会话中的主要阻点/痛点

- 旧 gate 给出了“通过”的错误信号。
  - **证据**：修复前原有 TopBar Vitest / Playwright gate 能通过，但用户截图显示 native select / 独立 custom accent button 与 `ui-design` menu/toggle 明显不同；旧 `topbar.spec.ts` 主要断言高度、padding、border、5 个 nav testid。
  - **影响**：需要额外 Red phase 增加 source-level structure assertions，才能把用户指出的视觉差异变成可执行失败。
- 旧文档口径与当前 UI 真理源冲突。
  - **证据**：`frontend-shell` spec、001 plan/checklist/BDD、frontend README、E2E.P0.001/P0.004/P0.005 仍出现 language dropdown/select、`topbar-theme-select`、`topbar-lang-select`、`topbar-custom-accent-button` 等正向契约。
  - **影响**：仅改代码会留下下一轮实现误读风险，因此本轮必须同时修正 spec/plan/README/scenario 并同步 INDEX。
- 002 与 003 的分工边界容易被误读。
  - **证据**：003 pixel gate 强化了真实浏览器 layout / screenshot，但没有覆盖 TopBar 控件源结构；002 checklist 原先把视觉 parity 证据写得过宽，容易把 layout gate 当成源码结构 gate。
  - **影响**：本轮在 002 checklist 追加 3.3 L2 remediation，并在 Playwright topbar spec 中增加 theme menu / custom picker / language toggle 断言。

## 3 根因归类

- 根因 1：视觉 plan 的 parity gate 没有要求“DOM/interaction shape 反查 `ui-design/src/app.jsx`”。
  - **类别**：spec-plan
  - **处理**：已在 002 checklist 3.3 固化 source-level TopBar parity gate。
- 根因 2：D1 i18n 修复形成的 select/dropdown 旧口径没有在 D2 视觉迁移后被清理。
  - **类别**：spec-plan / README
  - **处理**：已原地修订 spec/history、001/002 plan、frontend README 与 E2E scenario 文案。
- 根因 3：Playwright pixel parity 被用于验证布局，但没有同时承担控件类型和交互层级验证。
  - **类别**：spec-plan
  - **处理**：已扩展 `frontend/tests/pixel-parity/topbar.spec.ts` 的 structure assertions。

## 4 对流程资产的改进建议

- 后续 `/plan-code-review --fix` 对 UI parity 类 target，应强制形成两类 finding：一类是视觉几何/layout，另一类是 source DOM/interaction shape。不要把 computed style 或 screenshot gate 当作源码复刻的充分条件。本规则已在 `.agent-skills/plan-code-review/SKILL.md` Step 4 固化为 UI parity source-level reverse-audit 要求。
  - **落点**：`.agent-skills/plan-code-review/SKILL.md`
  - **优先级**：done
- 视觉 plan checklist 的每个用户可见控件应包含“旧正向口径负向搜索”项，搜索范围至少覆盖 spec/plan/README/scenario/test，而不只覆盖 runtime code。
  - **落点**：spec-plan checklist 模板或 `/design` 视觉类 coverage row
  - **优先级**：medium
- `frontend/README.md` 已经明确 `ui-design` 原生迁移规则；下一轮 D2-D6 页面实现应直接以 README §3 + Playwright gate 为入口，避免再从旧 plan 历史证据推导视觉。
  - **落点**：no repo change needed
  - **优先级**：done

## 5 建议优先级与后续动作

- 最高优先级已闭环：`/plan-code-review` 的 UI parity review 指令已把“源 DOM/interaction shape”列为 L2 必查项。这能直接降低下次“测试通过但截图明显不一致”的风险。
- 下一步业务实现建议：在当前 `frontend-shell` parity gate 通过后，继续推进 `frontend-home-job-picks-and-parse` owner，先把 Home / Job Picks / Parse 的静态原型按相同 source-level parity 方式迁入正式前端。
