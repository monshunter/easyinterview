# Frontend Home Compact Shortcut UX 交付复盘报告

> **日期**: 2026-07-06
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复 Home 首页截图反馈：将“选择已有简历”从平铺按钮列表改为 dropdown / combobox；Recent mock interviews 只展示最近 3 张卡片；超过 3 条时显示“更多”并跳转 `workspace` 模拟面试列表页。
- 同步范围覆盖 `ui-design/` 静态原型、`docs/ui-design/`、owner spec/plan/checklist/BDD、正式 frontend、Home focused tests、pixel parity、P0.014/P0.015 场景文档和 BUG-0132。
- 成功证据：`HomeResumeSelection.test.tsx` 3 tests PASS，`HomeRecentMocks.test.tsx` 7 tests PASS，`HomeImport.test.tsx` 11 tests PASS，`HomeAuthGate.test.tsx` 5 tests PASS，`HomeScreen.test.tsx` 8 tests PASS，`localeFiles.test.ts` 4 tests PASS。
- UI/parity 证据：`node ui-design/ui-design-contract.test.mjs` 29 tests PASS；`playwright test tests/pixel-parity/home.spec.ts` desktop/mobile 8 tests PASS；frontend `typecheck` 与 `build` PASS。
- BDD 成功证据：P0.014 `setup.sh` / `trigger.sh` / `verify.sh` / `cleanup.sh` PASS；P0.015 `setup.sh` / `trigger.sh` / `verify.sh` / `cleanup.sh` PASS。
- 收尾证据：`sync-doc-index --check`、`make docs-check`、`git diff --check` PASS；owner plan/checklist/BDD 已恢复 `completed`。

## 2 会话中的主要阻点/痛点

- Phase 8 锁定了“必须选择已有简历”，但没有锁定控件形态。
  - **证据**：用户截图指出当前实现把简历全部平铺展示；新增 Red test 证明旧 `home-resume-select` 是 `DIV` 而不是 `SELECT`。
  - **影响**：实现满足了业务绑定，却破坏首页输入区密度，需要补改 UI 真理源、正式前端和测试。

- Recent mocks 继续沿用历史 12-card 列表口径。
  - **证据**：用户要求最近模拟面试只展示三个并增加“更多”；新增 Red test 证明 twelve-plus variant 旧实现渲染 12 张卡片。
  - **影响**：首页从快捷入口变成列表页替代品，遮挡主任务区域，也缺少通往完整“模拟面试”列表页的显式入口。

- 场景 gate 之前只覆盖数据存在和排序，没有覆盖信息架构边界。
  - **证据**：P0.014 / P0.015 需要同步补充 dropdown、3-card cap、`home-recent-more` workspace handoff 和真实 `resumeId` 传递说明。
  - **影响**：如果不更新 BDD，旧布局仍可能被未来回归测试接受。

## 3 根因归类

- 控件语义没有下沉到 owner plan gate。
  - **类别**：spec-plan
  - **说明**：plan 用“选择已有简历”描述业务动作，但缺少 dropdown / combobox、旧平铺列表负向断言和 testid 控件类型要求。

- 首页 recent 区域的角色没有和完整列表页边界分开。
  - **类别**：spec-plan
  - **说明**：历史 `pageSize=12` 是数据获取和旧 UI 展示混合口径；当前产品需要把 backend page size 与首页 preview cap 拆开。

- 截图反馈暴露的 UI 密度约束尚未形成可执行测试。
  - **类别**：spec-plan
  - **说明**：本次已通过 Home focused tests、UI contract test、pixel parity 和 P0.014/P0.015 补齐。

## 4 对流程资产的改进建议

- 后续 UI 需求若出现“选择”“更多”“最近”等高频词，owner plan 应明确控件类型、显示上限、跳转目标和旧形态负向断言。
  - **落点**：spec-plan
  - **优先级**：high

- `docs/ui-design/` 对首页类页面应区分“快捷入口 preview”和“完整列表页”，避免把 API page size 直接等同于首页展示数量。
  - **落点**：spec-plan
  - **优先级**：medium

- P0 场景 expected outcome 应覆盖控件 role / tagName、列表 cap 和 More navigation，不只覆盖文案与 testid 存在。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：下一轮 Home 或 Workspace UI 入口改动时，先在 owner plan 中写清“控件类型 + 展示数量 + 跳转目标 + 负向旧形态”四件套，再进入 `/implement`。
- 中优先级：把 `workspace` 模拟面试列表页的空态、分页和从 Home `更多` 进入后的高亮/过滤行为作为后续 owner plan 的明确验收点。
- 可延后：若后续同类问题重复出现，再把“preview cap 与 API pageSize 分离”沉淀到 `docs/ui-design/` 的通用页面约束。
