# Frontend Home Source Layout 交付复盘报告

> **日期**: 2026-07-06
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复 Home 首页截图反馈：把粘贴 JD 和上传文件拆成两个 source panel；将 `选择已有简历` 下拉框收敛为适度宽度并与 `还没有简历？1 分钟创建 →` 同行；将 `立即面试` 放到简历选择下方，离开 textarea card。
- 同步范围覆盖 `ui-design/` 静态原型、`docs/ui-design/`、owner spec/plan/checklist/BDD、正式 frontend、Home focused tests、pixel parity、P0.014/P0.015 场景文档和 BUG-0133。
- 成功证据：`HomeLayout.test.tsx` red run 在旧实现下 3/3 失败；修复后 Home focused command 通过 7 files / 41 tests，覆盖布局、导入、简历选择、auth gate、recent mocks 和 i18n。
- UI/parity 证据：`node ui-design/ui-design-contract.test.mjs` 29 tests PASS；`playwright test tests/pixel-parity/home.spec.ts` desktop/mobile 10 tests PASS；frontend `typecheck` 与 `build` PASS。
- BDD 成功证据：`test/scenarios/env-setup.sh` / `env-verify.sh` PASS；P0.014 与 P0.015 的 `setup.sh` / `trigger.sh` / `verify.sh` / `cleanup.sh` 均 PASS。
- 收尾证据：`sync-doc-index --check`、`make docs-check`、`git diff --check` PASS；owner spec / plan / checklist / BDD 已恢复 `completed`。

## 2 会话中的主要阻点/痛点

- 首页导入区仍沿用旧“textarea footer 辅助上传 + 右下角提交”结构。
  - **证据**：用户截图标出上传/URL/提交按钮仍和 textarea 卡片绑定；`HomeLayout.test.tsx` 红灯显示缺少 `home-source-layout` 和 `home-jd-input-card`，上传按钮仍未独立成 source panel。
  - **影响**：页面表达仍偏向“解析 JD”，没有形成“先选 source，再选简历，再立即面试”的规划表单顺序。

- 简历选择虽然已改成 dropdown，但缺少布局边界。
  - **证据**：用户指出下拉框过长且创建简历入口不在同一行；新测试和 pixel parity 增加 `home-resume-row`、360px select、desktop 同行几何断言。
  - **影响**：用户难以把“选择已有简历”和“创建简历”视为同一个决策点，主按钮也显得位置尴尬。

- 上一轮 BDD 只覆盖 dropdown 和 recent cap，没有覆盖 source/submit 所属区域。
  - **证据**：本轮需要补 P0.014/P0.015 README、expected outcome、trigger/verify 脚本，并把 `HomeLayout.test.tsx` 纳入场景。
  - **影响**：若只改正式前端，未来场景仍可能接受旧 textarea-footer 布局。

## 3 根因归类

- 首页新建规划的步骤顺序没有在 Phase 8/9 中被完全形式化。
  - **类别**：spec-plan
  - **说明**：业务已要求绑定简历，但布局 gate 没有明确 source panel、resume row、submit row 的层级关系。

- UI feedback 中的“空间关系”没有转成 bounding box / DOM hierarchy gate。
  - **类别**：spec-plan
  - **说明**：本轮已补 `HomeLayout.test.tsx` 和 Home pixel parity，但该规则此前没有被 owner plan 固化。

- 源级复刻需要同时覆盖 `ui-design` 与正式前端，而不是只修正在浏览器里看到的实现。
  - **类别**：spec-plan
  - **说明**：本轮先修订 owner docs 和 `ui-design/src/screen-home.jsx`，再进入正式实现和场景 gate，避免 truth source 继续漂移。

## 4 对流程资产的改进建议

- 后续 Home / Workspace 表单类 UI 调整，应在 owner plan 中写清每个 action 的所属 row/panel、顺序和负向 containment 断言。
  - **落点**：spec-plan
  - **优先级**：high

- 对截图反馈中出现的“太宽”“同一行”“下方”等布局词，应默认补一个 DOM hierarchy test 和一个 Playwright bounding box test。
  - **落点**：spec-plan
  - **优先级**：high

- P0 场景脚本除了检查 test 文件名，还可以在 verify 中增加 source-level grep，确保关键 testid 或 layout markers 出现在正式前端和 ui-design 两边。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：下一轮处理 `workspace` 模拟面试列表页时，把 Home `更多` 的落点、列表页筛选/排序、空态和卡片密度写成明确 owner plan gate。
- 中优先级：若 Home 继续增加入口，先更新 `ui-design` 并补 containment/bounding-box tests，再迁移正式前端。
- 可延后：如果类似截图反馈再次出现，把“布局词自动转 DOM + bounding box gate”的规则沉淀到前端 owner plan 模板或 UI design README。
