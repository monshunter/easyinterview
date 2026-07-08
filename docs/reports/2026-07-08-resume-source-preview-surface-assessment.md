# Resume Source Preview Surface 交付复盘报告

> **日期**: 2026-07-08
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `frontend-resume-workshop/001-listing-routing-and-detail-readonly` 中 Markdown body 注入 `displayName` / header 元数据，以及 PDF / Markdown 详情正文区域视觉背景不一致的问题。关联 Bug：[BUG-0140](../bugs/BUG-0140.md)。
- 成功证据：
  - Red 阶段：`corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/components/ResumePreviewTab.test.tsx src/app/screens/resume-workshop/ResumeWorkshopCssParity.test.ts` 失败，确认 `resume-detail-markdown-page` 缺失、Markdown body 被额外插入名称、PDF 专属 modifier 仍存在。
  - Green 阶段：同一 focused Vitest 通过。
  - 回归范围：`corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx src/app/screens/resume-workshop/components/ResumePreviewTab.test.tsx src/app/screens/resume-workshop/ResumeWorkshopCssParity.test.ts` 通过。
  - UI truth / browser gate：`node --test ui-design/ui-design-contract.test.mjs`、`corepack pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/resume-workshop.spec.ts` 通过。
  - 工程 gate：`corepack pnpm --filter @easyinterview/frontend typecheck`、`corepack pnpm --filter @easyinterview/frontend build`、`make docs-check`、`git diff --check` 通过。

## 2 会话中的主要阻点/痛点

- 现有 Markdown renderer 测试只验证了 heading/list/strong/link DOM，没有负向断言 body 区不得包含详情 header metadata。
  - **证据**：新增 Red test 在 `ResumePreviewTab.test.tsx` 中构造唯一 `displayName` 后失败，DOM 显示 `<h3 class="ei-text-title">Injected Display Name Must Stay In Header</h3>`。
  - **影响**：Markdown 正文污染在前一轮 Markdown engine 修复后仍能漏出。
- PDF page-stack refinement 只把背景板合同落在 PDF 分支，没有同步要求 Markdown 分支使用同一阅读面。
  - **证据**：CSS parity Red 阶段显示 `.ei-resume-detail-preview-card--pdf` 仍存在，且 `.ei-resume-detail-markdown-page` 缺失。
  - **影响**：同一个详情页因来源格式产生割裂观感，截图复核阶段才暴露。

## 3 根因归类

- 根因：`frontend-resume-workshop/001` 的 Phase 9 只覆盖 PDF page-stack 正向行为，没有把 Markdown body purity 和 shared reading surface 写成执行项。
  - **类别**：spec-plan。
- 根因：UI truth source 和 CSS parity 缺少 Markdown page-level anchor，只能证明 Markdown DOM 语义存在，不能证明它处在正确的阅读容器中。
  - **类别**：spec-plan / no repo change needed（本轮已补入 owner plan、BDD、CSS parity 和 pixel smoke）。

## 4 对流程资产的改进建议

- 后续任何 source-adaptive renderer 计划都应同时写正向和负向 UI contract：正向断言各 renderer anchor，负向断言非正文 metadata 不进入 body。
  - **落点**：spec-plan
  - **优先级**：medium
- Pixel parity 测试在检查页面截图前，应优先检查共享容器与每个 renderer 的 page-level anchor。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高价值后续动作：在后续 resume detail 相关改动中，把 `ResumePreviewTab.test.tsx` 的 “body purity” 测试当作必跑 focused gate，避免 header/source metadata 再次进入简历正文。
- 可延后动作：若后续新增 DOCX 或其它 source renderer，再抽象成更通用的 source renderer parity helper；当前 PDF/Markdown 两分支不需要额外抽象。
