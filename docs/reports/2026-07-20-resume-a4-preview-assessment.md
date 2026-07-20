# Resume A4 Preview 交付复盘报告

> **日期**: 2026-07-20
> **审查人**: Codex

**关联计划**: [Resume Listing and Readonly Detail](../spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)

## 1 复盘范围与成功证据

- 本次在原 owner 的 Phase 27 收敛简历详情纸张几何：PDF 页面采用 `794px` A4 宽度与 `210:297` 比例；Markdown 只采用相同纸宽，整份正文保持单个连续长页，不设置 A4 高度、固定高度或最小高度。
- `ResumeWorkshopCssParity.test.ts` 先以 RED 捕获 Markdown 残留的 desktop `aspect-ratio/min-height`，GREEN 后 focused preview tests 10/10、Resume owner 20 files/118 tests、typecheck、production build 与根 `make test` 621 tests/4628 subtests 全部通过。
- 真实 frontend/backend 重新部署后，Chrome 测得 Markdown desktop `794 × 2003px`、mobile `358 × 3429px`，均只有一个 page node、`aspect-ratio:auto`、`min-height:0` 且无横向溢出；PDF desktop `794 × 1123px`、mobile `358 × 506px`，保持 A4 比例并由 canvas 填满可用页宽；console warning/error 为 0。

## 2 会话中的主要阻点/痛点

- 首轮把“Markdown 使用 A4 宽度”误扩展为“Markdown 同时使用 A4 高宽比”。
  - **证据**：用户补充说明 Markdown 只采用 A4 宽度、整体是一页；随后 RED 精确命中 `.ei-resume-detail-markdown-page` 的 `aspect-ratio` 与 desktop/mobile `min-height`。
  - **影响**：产生一次样式与文档返工，但没有进入最终交付。
- PDF 外层页框达到 A4 尺寸后，canvas 的 inline presentation width/height 仍留下内侧空白。
  - **证据**：第一轮 Chrome 实测外层宽 `794px`、canvas 仅约 `745px`；移除 inline 展示尺寸后，desktop/mobile canvas 分别达到 `792px/356px`，只保留页框边界。
  - **影响**：仅靠 CSS source contract 会得到假阳性，必须增加 renderer 级断言并进行实际 DOM 测量。

## 3 根因归类

- Markdown 问题是格式特定几何没有在首轮合同中明确拆分：PDF 的“纸张比例”被错误传播到 Markdown。**类别**：spec-plan。
- PDF 问题是外层容器尺寸与 canvas presentation 尺寸属于两个 owner，原测试只约束前者。**类别**：spec-plan / 无需额外治理变更。

## 4 对流程资产的改进建议

- 保持当前 Phase 27 的格式矩阵：共享的只有 `794px` 纸宽；PDF 单独拥有 `210 / 297`；Markdown 明确拒绝 `aspect-ratio`、fixed height 和 `min-height`。该约束已写入 UI design、Spec、Plan、BDD 与 CSS parity test。**落点**：spec-plan；**优先级**：high；**状态**：本次已完成。
- 对 PDF 预览继续同时断言 outer page bbox 与 canvas bbox，避免只看 A4 页框而漏掉内容区内缩。该约束已写入 renderer test 与 Phase 27 checklist。**落点**：spec-plan；**优先级**：high；**状态**：本次已完成。
- 后续纸张几何调整继续保留真实 Chrome 的 desktop/mobile bbox、page-node count、overflow 与 console 证据，不把代码层 gate 包装为 E2E。**落点**：无需仓库改动；**优先级**：medium。

## 5 建议优先级与后续动作

- 下一轮最值得保持的是格式分治：PDF 验证逐页 A4，Markdown 验证单节点连续长页；两者只共享宽度上限。
- 合并前由用户直接查看 desktop/mobile Markdown 截图；若视觉方向不再调整，可按当前 branch diff 进入 review/merge。
