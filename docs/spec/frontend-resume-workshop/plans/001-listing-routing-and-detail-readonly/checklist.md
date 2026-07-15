# Frontend Resume Workshop Listing Routing and Detail Readonly Checklist

> **版本**: 3.8
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: Route Shell / Auth Boundary

- [x] 1.1 `ResumeWorkshopScreen` 分派当前 `flow=create|list`、`resumeId`、`targetJobId`、`createMode` route params，忽略 out-of-scope `tab` / `tailorRunId`；验证: `ResumeWorkshopScreen.test.tsx`。

## Phase 2: Flat List View

- [x] 2.2 loading / empty / retryable error / pagination 状态从 API response 派生；验证: `ResumeListView.test.tsx` 与 `fixture-parity.test.ts`。

## Phase 3: Read-only Detail

- [x] 3.5 `mapResumeToUiSource` 对通用“粘贴的简历 / 上传的简历 / Pasted resume / Uploaded resume”做负向过滤，并只从 LLM-derived `displayName` / structured headline 派生可识别名称；不得把 raw resume 第一行作为名称；验证: `adapters/resume.test.ts`、`ResumeListView.test.tsx`、`ResumeDetailView.test.tsx`。
- [x] 3.6 `mapResumeToUiSource` 对上传文件名和与来源 `title` 相同的文件名 `displayName` 做负向过滤，文件名只可作为 `sourceName`，不可成为列表/详情可见简历名称；验证: `adapters/resume.test.ts`。
- [x] 3.8 `ResumeDetailView` 对 `failed` 或已有 `parsedTextSnapshot` / `originalText` 的上传详情停止 `getResume` 轮询，避免同一详情 URL 重复请求；验证: `ResumeDetailView.test.tsx` focused regression。

## Phase 4: Privacy / I18n / A11y / Parity

- [x] 4.1 raw resume text、parsed snapshot、structured profile、rewrite text 不进入 URL / pending action / localStorage / console / generic logs；验证: `ResumeWorkshopPrivacy.test.ts`。
- [x] 4.2 中英文案、Accept-Language、只读详情和 aria 语义可测试；验证: `ResumeWorkshopI18nA11y.test.tsx`。
- [x] 4.3 UI layout、computed style、viewport 与 accessibility 由正式前端 owner gates 承接；验证: `ResumeWorkshopCssParity.test.ts` 与 Resume Workshop component tests。



## Phase 6: Resume module UX optimization

- [x] 6.1 `ResumeListView` 删除底部“上传或粘贴另一份简历”CTA，只保留 Header “新建简历”；验证: `ResumeListView.test.tsx` focused negative。
- [x] 6.2 `ResumeListView` 每行支持删除简历，调用 `archiveResume` 成功后隐藏该 row，失败时保留 row 并显示错误；验证: `ResumeListView.test.tsx` focused success/failure。
- [x] 6.3 `ResumeDetailView` 对 `queued/processing` 且无正文的简历渲染等待动画页并轮询，`ready` 后切换 Markdown 详情，`failed` 且无正文时显示失败态；验证: `ResumeDetailView.test.tsx` focused pending/ready/failed。
- [x] 6.4 `ResumePreviewTab` 将 `parsedTextSnapshot` 按 Markdown 标题、段落、列表渲染，兼容 `originalText` 纯文本 fallback；验证: `ResumePreviewTab.test.tsx` focused Markdown。

## Phase 7: Markdown engine remediation

- [x] 7.1 `ResumePreviewTab` 使用 `react-markdown` + `remark-gfm` 渲染 `parsedTextSnapshot`，覆盖 heading/list/paragraph 以及 inline `strong` / link，禁止 `**...**`、`[label](url)` 以源码形式露出；验证: `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/components/ResumePreviewTab.test.tsx` 与 `corepack pnpm --filter @easyinterview/frontend typecheck`。

## Phase 8: Source-format renderer

- [x] 8.1 `ResumePreviewTab` 对 upload-backed PDF 简历使用 source endpoint renderer，source URL 指向 generated client baseUrl 下的 `/resumes/{resumeId}/source`；不得渲染 Markdown fallback、copy/export/original-modal 或新 tab；原 2026-07-07 native object 实现已由 Phase 9 page-stack refinement 取代。
- [x] 8.2 paste、Markdown upload 和 TXT upload 继续使用 Markdown engine，保持 heading/list/paragraph/inline strong/link DOM 断言；验证: focused frontend tests。

## Phase 9: PDF page-stack refinement

- [x] 9.1 `ResumePreviewTab` 对 upload-backed PDF 简历渲染从上到下平铺的 PDF 页面栈，source URL 仍指向 generated client baseUrl 下的 `/resumes/{resumeId}/source`；不得渲染 `<object>` / `<iframe>` / `<embed>`、browser PDF viewer toolbar、Markdown fallback、copy/export/original-modal 或新 tab；验证: `ResumePreviewTab.test.tsx` focused red/green。
- [x] 9.2 PDF 页面栈样式与 UI design document 同步，desktop/mobile pixel smoke 断言 `resume-detail-pdf-preview-stack` 与 page anchors 可见且没有 native viewer shell；验证: `ResumeWorkshopCssParity.test.ts` + `formal frontend component tests` focused smoke。

## Phase 10: Source-format reading surface alignment

- [x] 10.1 `ResumePreviewTab` Markdown renderer 只渲染 `buildResumeBodyMarkdown(resume)` body，不在 body card 内额外注入 `displayName` / `uiResume.name` / summary / source metadata；验证: `ResumePreviewTab.test.tsx` focused red/green。
- [x] 10.2 PDF 与 Markdown renderer 使用统一阅读背景板；Markdown 也渲染白色 page surface，CSS parity 与 pixel smoke 覆盖 Markdown page anchor / PDF page-stack anchor / shared background；验证: `ResumePreviewTab.test.tsx`、`ResumeWorkshopCssParity.test.ts`、`formal frontend component tests` focused smoke。



## Phase 12: PDF.js on-demand loading

- [x] 12.1 `PdfPageStackPreview` 在现有 loading shell 后动态导入 PDF.js/runtime worker，保持 credential、cancel/error/page-stack 合同（验证：focused red/green、Resume Workshop/full frontend tests、build+sourcemap 主 chunk/PDF chunk byte delta、owner context/docs gates）



## Phase 14: orphan Resume toast bridge removal



- [x] 15.2 Synchronize both tests on the ready `displayName` heading, then assert the exact two-call polling contract and unchanged PDF page-stack/content negatives.

## Phase 16: zero-consumer detail CSS pruning

- [x] 16.1 Add a source-level RED gate that names every zero-consumer breadcrumb, structured-preview and original-modal selector still present in `screens.css`.
- [x] 16.2 Delete those selectors and the dead shared-button/mobile branches without adding aliases, placeholders or removal markers; preserve all current list/detail selectors.
  <!-- verified: 2026-07-10 method=resume-css-source-green evidence="ResumeWorkshopCssParity passes 4/4. Target selectors are absent from formal CSS/DOM/prototype sources and remain only as negative test literals; current back/header/Markdown/PDF/parse-state selectors retain production consumers." -->
- [x] 16.3 仓库根 `make test` 完成前后端全量单测回归；CSS/component focused tests 仅作开发反馈，typecheck/build、owner/product contexts 与 docs/index/diff/pruning 作为独立 gates；随后恢复 `completed`。

## Phase 17: detail CSS cascade consolidation

- [x] 17.1 Add a source RED gate requiring one detail-back rule with the complete effective declarations and no flex-preview grid declaration.
- [x] 17.2 Merge the effective declarations into one rule and delete the ineffective media block without changing selectors, DOM or computed values.
  <!-- verified: 2026-07-10 method=resume-detail-css-cascade-green evidence="CSS parity passes 5/5. Exactly one detail-back rule retains the effective inline-flex/padding/color/background/border/radius/font declarations, and the flex preview no longer carries grid-template-columns; all remaining grid declarations belong to grid layouts." -->
- [x] 17.3 仓库根 `make test` 完成前后端全量单测回归；typecheck/build、owner/product contexts 与 docs/index/diff/pruning 作为独立 gates；随后恢复 `completed`。

## Phase 18: empty pending-decision section removal

- [x] 18.1 Record a scoped RED gate proving the active spec still contains an empty pending-decision section.
  <!-- verified: 2026-07-10 method=frontend-resume-empty-pending-red evidence="The scoped absence gate failed first on the active spec's combined decisions/pending heading; source inventory confirmed the only 3.2 content was a no-pending status sentence and no active anchor referenced it." -->
- [x] 18.2 Delete the empty section, synchronize spec/history/contexts/indexes, and run owner/product context plus docs/index/link/diff/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=frontend-resume-empty-pending-section-removal evidence="Spec v2.14 keeps only the current decisions heading and has no empty pending section or no-pending status sentence. History, both owner contexts and the top-level spec index are synchronized; 001/002/product contexts, links and final docs/index/diff/pruning gates pass. No code, UI, BDD, Bug/retrospective report, environment restart or data cleanup was needed." -->

## Phase 19: Resume summary consumption and idempotent initial reads

- [x] 19.1 RED：generated client / fixture parity / list hook / `ResumeListView` / Home selector tests 证明当前列表仍可见完整详情字段；新增 exact-key 断言，只允许 `id,title,displayName,language,sourceType,parseStatus,summaryHeadline,hasReadableContent,updatedAt`，并对 `originalText|parsedTextSnapshot|structuredProfile|fileObjectId|parsedSummary|createdAt|deletedAt` 建立编译期或运行时负向 gate。
- [x] 19.2 GREEN：在 B2 generated `ResumeSummary` 与 `PaginatedResume.items: ResumeSummary[]` 就位后，列表与 Home selector 只消费 summary；不得新增 pagination wrapper；`hasReadableContent` / `summaryHeadline` 取代从正文详情推断；`ResumeDetailView` 继续仅由 `getResume` 消费 full `Resume`。Focused adapter/hook/component/fixture parity 仅作开发反馈，typecheck 为独立 gate。
- [x] 19.3 RED/GREEN：StrictMode harness 同时覆盖 `listResumes` 与 ready `getResume`，以底层 `fetch`/transport spy 证明相同 no-signal request identity 当前产生重复 transport，再修复为恰好 1 次；不得删除 StrictMode 或只改测试 method mock。
- [x] 19.4 reject/retry/abort/polling：第一次相同请求 reject 后 registry 清空，用户 retry 发起新的 transport 并成功；resolve 后新用户动作也可发起新 transport；带 `AbortSignal` 请求不共享；queued/processing 详情仅在前次 settle 后轮询，ready/failed/已有正文不轮询。Focused hook/client/component tests 仅作开发反馈。
- [x] 19.7 收口：仓库根 `make test` 完成前后端全量单测回归；typecheck/build、owner contexts、`sync-doc-index --check`、`make docs-check`、`git diff --check` 与 pruning 作为独立 gates；完成后同步 checklist 证据并恢复 completed。

## BDD Gate

- [x] BDD-Gate: `BDD.RESUME.READ.001` 由 [BDD checklist](./bdd-checklist.md) 关联 list/readonly-detail owner behavior tests；不创建或声明真实 E2E PASS。
