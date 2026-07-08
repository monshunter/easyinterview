# Resume PDF Page Stack 交付复盘报告

> **日期**: 2026-07-08
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖 `frontend-resume-workshop/001-listing-routing-and-detail-readonly` 的 PDF 详情渲染：upload PDF 从浏览器原生 PDF object 改为 PDF.js 纵向页面栈，去掉 native viewer toolbar / sidebar / download / print / pagination controls。
- 成功证据：
  - `pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/adapters/resume.test.ts src/app/screens/resume-workshop/components/ResumePreviewTab.test.tsx src/app/screens/resume-workshop/components/PdfPageStackPreview.test.tsx src/app/screens/resume-workshop/ResumeWorkshopCssParity.test.ts`
  - `pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx src/app/screens/resume-workshop/components/ResumePreviewTab.test.tsx src/app/screens/resume-workshop/components/PdfPageStackPreview.test.tsx`
  - `pnpm --filter @easyinterview/frontend typecheck`
  - `pnpm --filter @easyinterview/frontend build`
  - `pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/resume-workshop.spec.ts`
  - `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/scripts/setup.sh`、`trigger.sh`、`verify.sh`、`cleanup.sh`
  - `make docs-check`、`git diff --check`

## 2 会话中的主要阻点/痛点

- 旧测试仍把 upload PDF ready / failed-with-snapshot 断言为展示解析文本正文。
  - **证据**：全量前端 Vitest 首次失败于 `ResumeDetailView.test.tsx` 和 `p0-037-resume-detail-preview-readonly.test.tsx`，期望 `AI Workflow` / `service-registry-operator`，实际为 PDF page-stack loading / fallback。
  - **影响**：实现已符合新合同，但场景测试和组件测试需要同步改为 source page stack 断言。
- 依赖更新需要匹配现有 pnpm store。
  - **证据**：`corepack pnpm --filter @easyinterview/frontend add pdfjs-dist@5.4.296` 因 store v10/v11 不一致失败，改用当前 node_modules 匹配的 `pnpm 11.7.0` 后成功。
  - **影响**：锁文件出现 pnpm 11 生成的无关 `libc` 元数据，需要机械清理后保留实际依赖变更。
- 全量前端 Vitest 仍有非本次范围失败。
  - **证据**：`pnpm --filter @easyinterview/frontend test` 最终剩余 `src/app/screens/report/__tests__/ReplayCta.test.tsx` 3 个失败，均为 mock client 缺少 `listResumes`。
  - **影响**：PDF 相关 gate 已通过，但仓库全量前端包仍不是 green。

## 3 根因归类

- PDF 详情合同变更后，P0.037 和组件测试没有自动从“解析文本正文”迁移到“source page stack”断言。
  - **类别**：spec-plan
- pnpm 写锁命令与当前安装的 pnpm major 不一致。
  - **类别**：README
- `ReplayCta` 测试 mock client 与当前 workspace/resume picker 依赖面不一致。
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- 在 `frontend-resume-workshop/001` 的 BDD/checklist 中保留 PDF page-stack negative gate：禁止 `object/iframe/embed`，并要求 P0.037 同步检查 source page stack。
  - **落点**：spec-plan
  - **优先级**：high
- 在前端依赖更新说明中记录：若 `corepack pnpm` 与现有 `node_modules` store major 不一致，应先确认当前安装来源，避免无关锁文件 churn。
  - **落点**：README
  - **优先级**：medium
- 对 `frontend-report-dashboard` 的 Replay CTA 测试补一个独立 follow-up：为测试 harness mock client 补齐 `listResumes`，或避免 replay 路径意外打开 ResumePickerModal。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：保持 001 owner 当前 page-stack gate，不再允许 PDF object 断言回流。
- 下一步建议：单独走 `frontend-report-dashboard` owner 的小修复，处理 `ReplayCta.test.tsx` mock client 缺少 `listResumes` 的全量前端 Vitest 残余失败。
