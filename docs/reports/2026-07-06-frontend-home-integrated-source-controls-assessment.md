# Frontend Home Integrated Source Controls 交付复盘报告

> **日期**: 2026-07-06
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复 Home 首页最新截图反馈：将上传 JD 文件与 URL 导入从独立右侧 source panel 回收到 `home-jd-input-card` 底部 `home-jd-source-controls`，恢复“粘贴 JD / 上传 JD 文件 / URL 导入在输入框中整合”的简约入口。
- 同步范围覆盖 `ui-design/` 静态原型、正式 frontend、`docs/ui-design/`、owner spec / plan / checklist / BDD、Home focused tests、UI contract、pixel parity、P0.014 / P0.015 场景文档和 BUG-0134。
- 成功证据：`HomeLayout.test.tsx` red run 在旧实现下失败；修复后 Home focused command 通过 7 files / 41 tests，覆盖布局、导入、简历选择、auth gate、recent mocks 和 i18n。
- UI/parity 证据：`node ui-design/ui-design-contract.test.mjs` 29 tests PASS；`playwright test tests/pixel-parity/home.spec.ts` desktop/mobile 10 tests PASS；frontend `typecheck` 与 `build` PASS。
- BDD 成功证据：`test/scenarios/env-setup.sh` / `env-verify.sh` PASS；P0.014 与 P0.015 的 `setup.sh` / `trigger.sh` / `verify.sh` / `cleanup.sh` 均 PASS。
- 收尾证据：`sync-doc-index --check`、`make docs-check`、`git diff --check` PASS；owner spec / plan / checklist / BDD 已恢复 `completed`。

## 2 会话中的主要阻点/痛点

- 上一轮 Phase 10 把“source 和 submit 不混在一起”过度实现为左右 source panel。
  - **证据**：用户最新截图明确指出粘贴 JD、上传 JD 文件和填写 URL 仍按旧 UI 在输入框中整合更简约；旧实现存在 `home-source-layout` 与 `home-upload-source-panel`。
  - **影响**：页面视觉重量增加，Home 作为“模拟面试新建规划快捷方式”的快速入口被拆得过复杂。

- 错误布局已经进入 owner plan 和自动化 gate。
  - **证据**：Phase 10、BDD 文档、`ui-design` contract 和 pixel parity 都接受独立 upload source panel；本轮必须新增 Phase 11 并反向禁止旧结构。
  - **影响**：如果只改代码，后续 contract 或 pixel gate 仍会把错误设计当成期望状态。

- change-intake 初筛会被 `home` / `jd_match` 共享关键词干扰。
  - **证据**：候选包含 `002-jd-match-recommendations`，但实际 owner 是 `001-home-jd-import-and-parse`。
  - **影响**：需要坚持 owner context validation，避免把 Home 新建规划入口的视觉反馈落到已移除或旁支主题。

## 3 根因归类

- 布局要求没有精确到容器归属。
  - **类别**：spec-plan
  - **说明**：Phase 10 写清了 source 分区和 submit row，却没有写清 upload / URL 应保留在 JD input card 内，导致实现合理但不符合用户偏好的简约形态。

- Gate 缺少旧结构负向搜索。
  - **类别**：spec-plan
  - **说明**：本轮已补 `home-source-layout` / `home-upload-source-panel` 负向断言，并把 `home-jd-source-controls` 作为输入卡片内的正向锚点。

- UI 真理源与正式前端需要同改。
  - **类别**：spec-plan
  - **说明**：如果只修正式前端，`ui-design` 仍会推动下一次迁移回错误布局；本轮先改 truth source，再迁移正式实现。

## 4 对流程资产的改进建议

- Home / Workspace 表单类反馈出现“整合”“简约”“旧 UI”时，owner plan 应写明保留的父容器和删除的旧容器名。
  - **落点**：spec-plan
  - **优先级**：high

- 对同一页面连续多轮截图反馈，应默认补旧结构负向断言，而不只新增当前目标锚点。
  - **落点**：spec-plan
  - **优先级**：high

- change-intake 对 `home` 与 `jd_match` 共享关键词的候选排序可以在未来增加“route / visible page / active module”权重，但当前 owner context validation 已足够拦截。
  - **落点**：change-intake
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：下一轮若继续调整 Home，继续在 `frontend-home-job-picks-and-parse/001-home-jd-import-and-parse` 原计划追加 Phase 12，先更新 `ui-design` 和 DOM/bounding-box gate，再动正式前端。
- 中优先级：处理 “更多” 跳转后的模拟面试列表页时，另起对应列表页 owner plan，明确 Home recent card 到列表页的排序、筛选和空态契约。
- 可延后：若类似 Home 连续截图反馈再次出现，把“正向容器归属 + 旧容器负向断言”沉淀到 UI design README 或前端 owner plan 模板。
