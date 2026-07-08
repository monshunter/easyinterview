# Interview Plan List Card UI 交付复盘报告

> **日期**: 2026-07-08
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付范围：`frontend-workspace-and-practice/001-workspace-and-interview-context` Phase 8 / Phase 9，修复一级 `面试` 无上下文规划列表看起来无样式、卡片内容噪声过多、CTA 与主题不一致的问题。
- 成功证据：
  - `pnpm --filter @easyinterview/frontend test src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceEmptyState.test.tsx src/app/screens/workspace/WorkspaceHandoff.test.tsx` 通过，31 tests。
  - `pnpm --filter @easyinterview/frontend typecheck` 通过。
  - `pnpm --filter @easyinterview/frontend build` 通过。
  - `pnpm --filter @easyinterview/frontend test:pixel-parity tests/pixel-parity/workspace.spec.ts` 通过，30 tests，覆盖 desktop/mobile card computed style、footer 简洁度和 accent CTA。
  - `E2E.P0.018` setup/trigger/verify/cleanup 通过，12 test files / 95 tests，并反查 card body/footer、elevation token、无 source/language metadata 回流。
  - `sync-doc-index --check`、`make docs-check`、`git diff --check` 通过。
- 关联 Bug：[BUG-0143](../bugs/BUG-0143.md)。

## 2 会话中的主要阻点/痛点

- Phase 7 只证明了列表存在和卡片可点击，没有证明卡片在浏览器里拥有可见边框、阴影和分区。
  - **证据**：用户截图显示列表区域只有松散文本列；新增 Playwright computed style gate 初次失败，`borderTopWidth` 为 `0px`。
  - **影响**：需要重新打开 completed owner plan，补 spec/plan/checklist/BDD gate 后再修实现。

- 正式前端沿用了 `ui-design` 原型对象字段命名风格，导致 CSS variable 在浏览器里无效。
  - **证据**：`WorkspacePlanList` 使用 `--ei-color-bgCard` / `--ei-color-rule` / `--ei-color-ink*`，而正式 token 是 `--ei-color-bg-card` / `--ei-color-rule-strong` / `--ei-color-fg-*`。
  - **影响**：DOM 和文本都存在，但 computed style 无法生效，产生“页面无样式”的用户感知。

- Phase 8 只定义了“卡片有容器”，没有定义“卡片应该省略什么”。
  - **证据**：用户第二张截图指出卡片中的 `手动输入`、重复来源和语言是无意义字段，`进入规划` 仍是 secondary 样式，卡片和页面背景不易区分。
  - **影响**：即使样式修复，列表仍显得啰嗦且主动作不明确，需要再次重开 owner plan 补 Phase 9。

## 3 根因归类

- 视觉合同不足。
  - **类别**：spec-plan。
  - Phase 7 缺少“卡片 affordance”这类用户可感知视觉合同，只写了 `Plan Cards` 信息结构。

- Parity gate 断言层级不足。
  - **类别**：spec-plan。
  - 既有 gate 更关注 anchor、导航和截图非空，没有针对新增列表卡片断言 resolved token、border、shadow 和分区结构。

- 原型到正式前端 token 映射容易误用。
  - **类别**：README / spec-plan。
  - `ui-design` 使用 `T.bgCard/T.rule/T.ink`，正式前端必须改为 kebab-case CSS token；这条迁移边界需要在视觉实现 checklist 中显式验证。

- 卡片信息架构缺少负向约束。
  - **类别**：spec-plan。
  - Phase 8 只写了可见卡片、body/footer 和按钮，没有写“不得展示 source/language/手动输入”等低价值字段，也没有把 CTA 样式绑定到主题 accent。

## 4 对流程资产的改进建议

- 对所有新 landing / list / card UI，checklist 应包含 computed style gate。
  - **落点**：spec-plan。
  - **优先级**：high。

- 在 frontend 视觉迁移说明中补一句：从 `ui-design` 的 `T.*` 字段迁移到正式前端时必须映射到 `themes.css` 中实际存在的 kebab-case CSS variables。
  - **落点**：README。
  - **优先级**：medium。

- 对新页面的 BDD verify，不只反查 anchor，还应反查至少一个视觉 token 或 computed-style 相关测试文件名。
  - **落点**：spec-plan。
  - **优先级**：medium。

- 对列表卡片，owner spec/checklist 应同时写正向信息层级和负向字段边界。
  - **落点**：spec-plan。
  - **优先级**：high。

## 5 建议优先级与后续动作

- 下一步最值得做的是在后续 UI 实施计划模板或 owner checklist 里固定“新增卡片/列表必须有 computed style parity gate + 低价值字段负向断言”的要求，避免 anchor-only 和 field-dump 假绿。
- 可以延后处理的是对历史 workspace 详情页旧 token 的全面审计；本次按用户截图只修正无上下文规划列表，未扩大到当前规划详情或 modal 的历史样式债。
