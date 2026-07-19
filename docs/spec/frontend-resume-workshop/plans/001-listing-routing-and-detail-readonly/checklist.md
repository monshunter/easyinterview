# Frontend Resume Workshop Listing Routing and Detail Readonly Checklist

> **版本**: 4.3
> **状态**: completed
> **更新日期**: 2026-07-19

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

## Phase 20: Responsive resume card list

- [x] 20.1 RED：`ResumeListView.test.tsx` 与 CSS/responsive owner gate 先证明当前 table/header/row DOM 和整行布局不符合卡片合同；同时锁定 closed `ResumeSummary`、loading/empty/error/pagination、打开与删除的既有行为。（验证：focused RED 7 failed；GREEN 后 Resume Workshop 20 files / 118 tests PASS）
- [x] 20.2 GREEN：`ResumeListView` 改为 list/card DOM；卡片展示名称、可选摘要、来源、语言和最近编辑，缺摘要不伪造，底部“打开”、右上角 trash 删除；desktop 固定最大列宽 + `auto-fill` + 左对齐，mobile 同序单列，1/2/3 卡片规格稳定且单卡不拉伸整行。（验证：`ResumeListView.test.tsx`、`ResumeWorkshopCssParity.test.ts`、i18n/a11y 与 StrictMode list boundary PASS；旧 selector 仅剩负向断言）
- [x] 20.3 A11Y/PARITY：打开与删除具有独立可访问名称、键盘焦点和触控区域；超长名称/摘要/来源完整换行；desktop/mobile geometry、no-overflow 与正式 screenshot acceptance 通过，不新增 nested-card-button 冲突。（验证：Chrome 1440 卡宽 360px/左对齐、390 单列 354px、两端无横溢；打开精确进入只读详情）
- [x] 20.4 REGRESSION：focused Vitest 只作开发反馈；执行根 `make test`、frontend typecheck/build、owner context、`sync-doc-index --check`、`make docs-check`、`git diff --check`，同步证据后恢复 completed。（验证：根后端 551 tests/4493 subtests、前端 125 files/993 tests PASS；lint/build/context/docs/index/diff PASS）

## Phase 21: Parse waiting motion stability

- [x] 21.1 RED：`ResumeDetailView.test.tsx` 以首轮 processing + 第二轮 pending Promise 复现“正在解析简历”被通用 loading 替换；`ResumeWorkshopCssParity.test.ts` 同时锁定 keyframes 不包含 `transform: scale(...)` / `translate(...)`。（验证：组件旧实现精确失败于第二轮 pending 时 `resume-detail-parse-waiting` 消失且 DOM 显示 `Loading resume…`；CSS 旧实现失败于缺少 box-shadow 并仍含 scale）
- [x] 21.2 GREEN：`useResume` 区分首次读取与已有 pending data 的后台轮询，后者保留等待态直到终态原子替换；同时移除图标循环缩放，保留 opacity/box-shadow 动画，并补 `prefers-reduced-motion: reduce`。（验证：focused CSS/component 17 tests PASS，frontend typecheck PASS）
- [x] 21.3 REGRESSION：真实 Chrome 连续 40 次/50ms 采样覆盖多个 250ms poll，解析等待态 40/40、通用 loading 闪现 0 次；10 帧 geometry 采样中图标/标题/说明各只有 1 个唯一边界，computed transform 恒为 `none`；恢复合成 Resume 为 ready 后平滑进入详情。根 `make test` 通过 Python 584 tests/4583 subtests、Go 全包、frontend 126 files/1029 tests；owner context、docs/index/diff gates 单独执行。

## Phase 22: Resume list reference alignment

- [x] 22.1 RED：`ResumeListView.test.tsx` / `ResumeWorkshopCssParity.test.ts` 锁定标题区、desktop 每行两张等宽卡、与 Workspace 一致的 22px 圆圈加号、文件 icon、名称/摘要、meta、删除与 footer 层级，并拒绝 14px 裸加号、单列 918px 规则和语言 tag。（验证：用户补充双列后 focused Vitest 2 项 CSS 断言预期失败；图标补充断言以实际 `width=14` 预期失败）
- [x] 22.2 GREEN：重构 `ResumeListView`、`ResumeWorkshopIcon` plus 几何与 `.ei-resume-workshop-*` scoped CSS；closed summary、排序、loading/empty/error/pagination、create/open/archive route/API 保持通过。（验证：create icon focused 8 tests PASS；此前双列修订 focused 16 tests 与 `npm run typecheck` PASS）
- [x] 22.3 BDD-Gate: 完成 `BDD.RESUME.LIST.VISUAL.003` domain behavior evidence，不声明真实 E2E PASS。（验证：owner scope 24 files / 151 tests PASS；Chrome UI evidence 由 22.4 独立承接）
- [x] 22.4 CHROME/REGRESSION：1916×821 下两张简历卡同排，分别 `x=254/972`、`width=690`、间距 28px；Resume/Workspace create icon 均实测为 22×22、`viewBox="0 0 24 24"`、`strokeWidth=1.8`、圆 `r=9` 与同一十字路径。390×844 下按钮和网格均为 358px 单列、icon 仍为 22×22、overflow 0；键盘 Enter、主题切换和 console 0 error 沿用本轮行为验收。截图保存于 `.test-output/list-ui-acceptance/`；focused 24 files / 151 tests、typecheck/build、根 `make test`（615 tests / 4615 subtests）、context/docs/index/diff gates 均通过后恢复 completed。

## Phase 23: Resume preview reference composition

- [x] 23.1 RED：`ResumeDetailView.test.tsx` / `ResumeWorkshopCssParity.test.ts` 拒绝 breadcrumb 拼接、缺失名称 kicker、窄 `1320/860/720px` 构图与只检查 overflow 的视觉 gate；既有 readonly/PDF/Markdown/privacy negatives 保持通过。<!-- verified: 2026-07-19 method=focused-vitest-red evidence="69 tests run; 8 expected structural failures, 61 existing behavior tests PASS; resume failures identify breadcrumb and 1512/1310/1150 composition gaps" -->
- [x] 23.2 GREEN：实现 Back + 蓝色 eyebrow + 名称 kicker + 主标题 + meta 的 Header 层级，以及约 `1512/1310/1150px` desktop 内容面/背景板/纸张；mobile 同序满宽且无横溢。<!-- verified: 2026-07-19 method=focused-vitest+typecheck evidence="Resume detail component/CSS gates PASS within 69-test focused run; tsc --noEmit PASS" -->
- [x] 23.3 BDD-Gate: `BDD.RESUME.DETAIL.VISUAL.004` 由 owner component/CSS tests 与 current-run Chrome UI evidence 验证，不创建 E2E wrapper。<!-- verified: 2026-07-19 method=chrome-real-local evidence="desktop 1916x821 screen=1512 board=1310 paper=1150 boardTop=346; exact mobile 390x844 board=358 paper=332; overflowX=0" -->
- [x] 23.4 REGRESSION：focused owner tests、frontend typecheck/build、根 `make test`、owner context、docs/index/diff 与 Chrome desktop/mobile evidence 通过后恢复 completed。<!-- verified: 2026-07-19 method=focused+root-regression evidence="owner 32 files/242 tests; root Python 615/4615 subtests; Go all packages; frontend 132 files/1057 tests; typecheck/build/redeploy PASS; dependencies 4/4 OK" -->

## BDD Gate

- [x] BDD-Gate: `BDD.RESUME.READ.001` 由 [BDD checklist](./bdd-checklist.md) 关联 list/readonly-detail owner behavior tests；不创建或声明真实 E2E PASS。
- [x] BDD-Gate: `BDD.RESUME.LIST.002` 由 [BDD checklist](./bdd-checklist.md) 关联 desktop/mobile 卡片列表、打开/删除与非表格语义行为；当前无真实 E2E owner，不把代码 gate 声明为 E2E PASS。
- [x] BDD-Gate: `BDD.RESUME.LIST.VISUAL.003` 由 [BDD checklist](./bdd-checklist.md) 关联 desktop 双列等宽卡、打开/删除与 responsive 行为；current-run Chrome 只作 UI evidence。
- [x] BDD-Gate: `BDD.RESUME.DETAIL.VISUAL.004` 由 [BDD checklist](./bdd-checklist.md) 关联详情 Header、背景板、纸张和 responsive 行为；current-run Chrome 只作 UI evidence。

## Phase 24: Screenshot-aligned resume parsing transition

- [x] 24.1 RED: detail/shared-scene/CSS tests 锁定 resume illustration、共享画布、TopBar active、同一 waiting DOM、返回动作、无伪百分比、reduced-motion 与 mobile containment，旧 circle block 先失败。<!-- verified: 2026-07-19 method=focused-vitest-red evidence="Resume pending-PDF test failed on missing resume variant/illustration while existing polling flow still reached ready" -->
- [x] 24.2 GREEN: queued/processing 无正文时复用 shared `resume` variant；保留 sequential poll、pending data、ready/failed atomic replacement、generated client 与 route。<!-- verified: 2026-07-19 method=focused-vitest-green evidence="Resume detail 9 and CSS parity 8 tests PASS including inter-request waiting DOM" -->
- [x] 24.3 BDD-Gate: `BDD.RESUME.PARSE.VISUAL.005` 覆盖中文/英文、连续轮询、ready replace、desktop Chrome 与 mobile responsive contract，不新增 E2E ID。<!-- verified: 2026-07-19 method=chrome-extension-manual evidence="Two real pasted-resume parses rendered one stable shared waiting DOM, full-bleed x=0 width=1920 canvas, shared TopBar and return action, then atomically replaced with ready details. Locale and mobile boundary remain component contracts." -->
- [x] 24.4 REGRESSION: focused、typecheck/build、根 `make test`、context/docs/diff 与 no-flash intermediate-DOM gate 通过后恢复 completed。<!-- verified: 2026-07-19 evidence="Resume focused 9 plus shared final 89 PASS; production build/redeploy and root make test 615 / 4615 PASS; no legacy loading string in ready DOM and browser console clean." -->
