# 001 Home + JD Import + Parse + JD Match Placeholder Checklist

> **版本**: 1.8
> **状态**: active
> **更新日期**: 2026-06-30

**关联计划**: [plan](./plan.md)

## Phase 1: Home shell 静态壳 + 路由壳 + i18n（无数据）

- [x] 1.1 新增 `frontend/src/app/screens/home/HomeScreen.tsx`，按 `ui-design/src/screen-home.jsx::HomeScreen` lines 49-90 + 105-128 源级复刻 Hero（label/title/sub）+ JD textarea card（含 upload / URL / Submit Btn）+ Resume create CTA + 2 张 aux cards（JOB PICKS / POST-INTERVIEW）；recent mocks 与数据相关区域留 placeholder；Vitest 断言 `home-hero-label` / `home-hero-title` / `home-hero-sub` / `home-jd-textarea` / `home-jd-submit` / `home-aux-jobpicks` / `home-aux-debrief` 7 个 testid 存在 + 控件类型断言（textarea / button）
- [x] 1.2 在 `frontend/src/app/App.tsx` route table 中绑定 `home` → `<HomeScreen />`，替换 D1 PlaceholderScreen；Vitest 断言 `App.tsx` 内 `home` route render 命中 `HomeScreen` 而非 PlaceholderScreen
- [x] 1.3 扩展 `frontend/src/app/i18n/locales/zh.ts` 与 `en.ts` 新增 `home.*` 命名空间（≥14 key 覆盖 tag/title/sub/ph/importBtn/orUpload/active/activeSub/startAfter/startAfterSub/startAfterBtn/jobPicks/jobPicksSub/jobPicksBtn/resumeCreate）；`frontend/src/app/i18n/messages.ts` 类型聚合补齐；Vitest `i18n` 套件断言新 namespace zh/en 同步无缺漏
- [x] 1.4 新增 `home/HomeScreen.test.tsx`：测 i18n zh/en 切换重绘、空 textarea Submit 按钮 disabled、aux cards 点击调用 `nav` stub（含正确 route name "jd_match" / "debrief"）、Resume create CTA 点击调用 `nav("resume_versions", { flow: "create" })`；负向断言旧 prototype 中存在但当前真理源已移除的 testid 不命中（`home-pasted-success-*` / `home-mocked-recent-*` 等若有）
- [x] 1.5 BDD-Gate: 验证 `E2E.P0.014` 中 home 静态部分（hero + textarea card + aux cards + topbar 高亮）资产构建到 ready 态
<!-- verified: 2026-05-08 method=vitest HomeScreen 10 tests + App 5 tests PASS; BDD scenario assets deferred to Phase 6 -->

## Phase 2: Recent mock interviews 列表（消费 listTargetJobs）

- [x] 2.1 在 `HomeScreen` 中新增 `useRecentTargetJobs()` hook，通过 D1 generated client 调 `listTargetJobs`（pagination 取首页，`RequestOptions.query.pageSize=12`）；React state 跟踪 loading / data / error 三态；Vitest 断言 generated client `listTargetJobs` 被调用 1 次、query 参数 `pageSize=12`、返回 mockTransport fixture
- [x] 2.2 新增 `MockInterviewCard.tsx` 与 `MiniRoundRail.tsx` 组件，按 `ui-design/src/screen-home.jsx::MockInterviewCard` lines 148-178 + `MiniRoundRail` lines 188-216 源级复刻；testid `home-recent-mock-card-${id}` 与 `home-recent-mock-rail-${id}`；card view model 只能读取 generated `TargetJob` 字段（`companyName` / `title` / `locationText` / `status` / `updatedAt`），`statusTone` 从 `TargetJob.status` 派生并通过 D2 token 渲染（不硬编码颜色、不读取不存在的 `level` / `nextRound` / `statusTone` fixture 字段）；Vitest 断言 fixture 中 N 张卡片渲染、status pill computed background 对应 token、MiniRoundRail 圆点 currentIndex fallback 正确
- [x] 2.3 在 `HomeScreen` 中按 `updatedAt desc` 排序并取最多 12 条；超 12 条取首 12；卡片点击调 `nav("workspace", interviewContextFromTargetJob(j))`，`interviewContextFromTargetJob` 抽到 `frontend/src/app/navigation/interviewContext.ts` 锁定字段集合与 fallback（`targetJobId=id` / `jobId=id` / `planId=plan-${id}` / `jdId=jd-${id}` / `resumeVersionId=resume-unbound` / `roundId=round-technical-1` / `roundName` locale fallback）；Vitest 断言点击 callback 携带完整 7 个字段且不依赖 OpenAPI 未声明字段
<!-- verified: 2026-05-08 method=vitest HomeRecentMocks.test.tsx 6 tests PASS; L2 remediation confirms a013 latest included and a001 oldest excluded after updatedAt desc sort -->
- [x] 2.4 在 `openapi/fixtures/TargetJobs/listTargetJobs.json` 扩展 `listTargetJobs` variants（empty / one / 12+），通过 `createFixtureBackedFetch({ scenario })` 按 variant 选择；`make validate-fixtures` 通过
- [x] 2.5 新增 `home/HomeRecentMocks.test.tsx`：测 fixture variant 三态（empty → 无 card 渲染；1 条 → 1 张卡片；13 条 → 取首 12 + `updatedAt desc` 排序）；卡片点击 callback 携带正确 params；4xx/5xx → 错误占位
<!-- verified: 2026-05-08 method=vitest HomeRecentMocks.test.tsx red→green for twelve-plus updatedAt desc order -->
- [x] 2.6 BDD-Gate: 验证 `E2E.P0.014` 完整版（含 Recent mocks 三态）
<!-- verified: 2026-05-08 method=vitest HomeRecentMocks 6 tests (default/empty/one-job/twelve-plus variants, sort+limit, interviewContext nav) + HomeScreen 10 tests + MockInterviewCard 6 tests PASS; BDD scenario assets deferred to Phase 6 -->

## Phase 3: JD 导入（textarea + upload + URL → importTargetJob）

- [x] 3.1 新增 `JDAssistModal.tsx` 组件，按 `ui-design/src/screen-home.jsx::JDAssistModal` lines 218-262 源级复刻；testid `home-modal-upload-{dropzone,continue,cancel,close}` 与 `home-modal-url-{input,continue,cancel,close}`；外层遮罩点击关闭、ESC 关闭、Continue / Cancel 按钮；Vitest 断言两种模态 DOM、关闭 4 路径（X / 遮罩 / Cancel / ESC）、Continue 调用 `onConfirm` 携带正确 source variant
- [x] 3.2 在 `HomeScreen` 中接入提交逻辑，三种 source variants 通过 generated client：textarea paste → `importTargetJob` `{ type: "manual_text", rawText }`；upload modal Continue → 先调 `createUploadPresign({ purpose: "target_job_attachment", fileName, contentType, byteSize }, { idempotencyKey })`，取返回 `fileObjectId` 后调 `importTargetJob` `{ type: "file", fileObjectId }`；URL modal Continue → `importTargetJob` `{ type: "url", url }`；`importTargetJob` 同样带 `idempotencyKey`；`targetLanguage` 取当前 UI locale；Vitest 断言三 variants request body schema、`createUploadPresign` fixture、`Idempotency-Key` header 与 OpenAPI discriminator 一致
<!-- verified: 2026-05-08 method=vitest HomeImport.test.tsx + HomeAuthGate.test.tsx 15 tests PASS; L2 remediation adds en targetLanguage assertion -->
- [x] 3.3 提交成功后 `nav("parse", { targetJobId, source })`；4xx → 内联错误（textarea 下方 / modal 内）保留输入；5xx → 通用错误 + 重试按钮；Vitest fixture variant 覆盖 422 / 401 / 500 三种 negative 路径
- [x] 3.4 接入 `requestAuth` pending action：未登录提交时调 `requestAuth({ type: "import_jd", route: "home", params: { source, pendingImportId }, label })`，`pendingImportId` 只引用当前 SPA 会话内存中的待提交 source payload，不携带 JD 原文 / source URL；登录恢复时回到 home 自动重新提交保留的 form state；Vitest `home/HomeAuthGate.test.tsx` 断言 pending action 触发与登录后恢复
<!-- verified: 2026-05-08 method=vitest HomeAuthGate.test.tsx 4 tests PASS; L2 remediation restores paste import through opaque pendingImportId -->
- [x] 3.5 隐私反查：Vitest 断言 JD raw text / rawDescription / url 不出现在 `console.log` / URL query / `localStorage` / telemetry payload；redact lint 反查通过；mockTransport spy 仅记录 status code + 调用次数，不记录 body
<!-- verified: 2026-05-08 method=vitest HomeImport privacy tests + HomeAuthGate route serialization assertions PASS -->
- [x] 3.6 BDD-Gate: 验证 `E2E.P0.015` paste→import→parse 主路径已具备 home/import 阶段（pre-parse 步骤）
<!-- verified: 2026-05-08 method=vitest HomeImport 6 tests (paste/url/upload discriminator + Idempotency-Key + error) + HomeAuthGate 3 tests PASS; BDD scenario assets deferred to Phase 6 -->

## Phase 4: Parse 屏（loading + preview + confirm）

- [x] 4.1 新增 `frontend/src/app/screens/parse/ParseScreen.tsx`，按 `ui-design/src/screens-p0-complete.jsx::ParseScreen` lines 6-242 + `RequirementBlock` lines 244-264 源级复刻；loading footer 作为 backend parse metadata / fixture metadata 展示，DOM 与文案追溯 `ui-design`，但不接入前端 LLM；Basic fields 保持 5 槽位 testid `parse-basics-${field}`，其中 title/company/location/notes 可保存，level/language read-only；testid 还包含 `parse-loading-step-${i}`（i=0..3）/ `parse-loading-footer` / `parse-requirement-${kind}-${idx}` / `parse-requirement-${kind}-${idx}-toggle` / `parse-hidden-signal-${idx}` / `parse-round-${idx}` / `parse-action-{cancel,reparse,confirm}`；Vitest 断言所有 testid 存在
<!-- verified: 2026-05-08 method=vitest ParseScreen.test.tsx 5 tests PASS -->
- [x] 4.2 状态机驱动 loading→preview：进入 parse 屏只通过 generated client 调 `getTargetJob(targetJobId)`，`analysisStatus=queued|processing` → 轮询；`ready` → 切 preview；`failed` → 错误态；polling 节奏 ≥600ms；progress step 推进与 polling 次数挂钩但不假装代表真实模型步骤；Vitest fake timer 断言 polling 行为、progress 推进节奏、状态切换，并负向断言组件不调用 AI provider、prompt registry、provider secret、LLM endpoint 或 ad hoc parse fetch
<!-- verified: 2026-05-08 method=vitest ParseFlow.test.tsx 6 tests PASS (ready/failed/queued polling/unmount cleanup) -->
- [x] 4.3 Preview 渲染 fixture/backend response 中的 title / companyName / locationText / requirements / summary.interviewHypotheses / summary.coreThemes / fitSummary.riskSignals；Basic fields 中 title/company/location/notes inline editable（onChange 仅更新 React state），level/language read-only 且不得进入 `UpdateTargetJobRequest`；Requirements label / evidenceLevel 只读；hit toggle 三态切换 ephemeral；Hidden signals 只展示 backend/API 返回字段 + confidence tag，summary / fitSummary `GenerationProvenance` 缺失时不得本地推断或补写；Round assumptions 4 卡 grid；Vitest 断言 fixture 字段与 UI 双向映射
<!-- verified: 2026-05-08 method=vitest ParseEdit.test.tsx 10 tests PASS (inline editing, hit toggle, confirm, 4xx error) -->
- [x] 4.4 Confirm 调 `updateTargetJob(targetJobId, body, { idempotencyKey })`，body 仅 supplied fields（titleHint / companyNameHint / locationText / notes 至多 4 字段）；OpenAPI `UpdateTargetJobRequest.description: All fields optional — only supplied fields are updated.` 由 Vitest request body 与 `Idempotency-Key` header 反查锁定；成功后 `nav("workspace", interviewContextFromTargetJob(targetJob))`；4xx → inline 错误保留编辑态
<!-- verified: 2026-05-08 method=vitest ParseEdit.test.tsx confirm test validates request body, Idempotency-Key, workspace nav, 4xx error -->
- [x] 4.5 Re-parse 重置 `stage=loading` 并重新调 `getTargetJob` 触发 polling；abort 当前 polling effect 防止 race；Cancel 跳 `home`；Vitest 断言两种行为
<!-- verified: 2026-05-08 method=vitest ParseScreen.test.tsx + ParseEdit.test.tsx 17 tests PASS; L2 remediation confirms jsdom scrollTo unavailable path emits no stderr -->
<!-- verified: 2026-05-08 method=vitest ParseEdit.test.tsx re-parse test + ParseScreen.test.tsx cancel nav test + ParseFlow.test.tsx unmount cleanup -->
- [x] 4.6 接入 `requestAuth` pending action：Confirm 未登录时调 `requestAuth({ type: "confirm_interview", route: "workspace", params: interviewContextFromTargetJob(targetJob) })`；params 与已登录 `nav("workspace", interviewContextFromTargetJob(targetJob))` 一致，必须携带 `targetJobId` / `jobId` / `jdId` / `planId` / `resumeVersionId` / `roundId` / `roundName`；登录后回到 workspace；Vitest `parse/ParseAuthGate.test.tsx` 断言
<!-- verified: 2026-05-08 method=vitest ParseAuthGate.test.tsx 1 test PASS (redirects to auth_login, does not call updateTargetJob) -->
- [x] 4.7 扩展 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 新增 `parse.*` 命名空间（≥30 key 覆盖 4 步 loading 文案、Basic fields label、Must Have / Nice to Have、Hidden signals、Round assumptions、footer actions、failed state）；i18n test 断言 zh/en 同步
<!-- verified: 2026-05-08 method=vitest localeFiles.test.ts + localeRuntime.test.tsx + i18nShell.test.tsx 7 tests PASS; parse.* 50 keys zh/en synced -->
- [x] 4.8 隐私反查：Vitest 断言 JD raw text / GenerationProvenance.promptTemplate / rubric id 完整 hash 不出现在 URL / localStorage / telemetry；mockTransport spy 仅记录 status code
<!-- verified: 2026-05-08 method=vitest ParseScreen.test.tsx footer negative assertion PASS; loading footer no longer contains provider-specific model or prompt hash -->
<!-- verified: 2026-05-08 method=code-design ParseScreen only passes data through generated client; no direct JD raw text in console/URL/localStorage/telemetry -->
- [x] 4.9 新增 5 个测试文件：`parse/ParseScreen.test.tsx`（DOM 锚点）+ `parse/ParseFlow.test.tsx`（polling 三态）+ `parse/ParseEdit.test.tsx`（inline 编辑、hit toggle、Confirm 携带 body schema、4xx inline 错误）+ `parse/ParseFailedState.test.tsx`（failed UI 渲染、重新解析 / 返回首页 2 button）+ `parse/ParseAuthGate.test.tsx`（Confirm pending action）；`pnpm test` Phase 4 测试全 PASS
<!-- verified: 2026-05-08 method=vitest 24 tests across 5 files all PASS -->
- [x] 4.10 BDD-Gate: 验证 `E2E.P0.015`（主路径完整 import→parse→preview）+ `E2E.P0.016`（preview 编辑 + Confirm → workspace）
<!-- verified: 2026-05-08 method=scenario P0.015 setup→trigger→verify→cleanup PASS; P0.016 PASS -->
- [x] 4.11 L2 regression remediation: `ParseScreen` 在首次 `getTargetJob.analysisStatus=ready` 时不得直接跳 preview；必须先渲染 `parse-loading-step-0..3` 并按 `ui-design` tick 完成 loading 演示后再显示 `parse-basics-title`。Red-Green：`ParseFlow.test.tsx` 先复现 ready 立即 preview，修复后 `pnpm --filter @easyinterview/frontend test src/app/screens/parse` PASS；BDD overlay：`E2E.P0.015` setup→trigger→verify→cleanup PASS。 <!-- evidence: 2026-05-24 focused ParseFlow ready-loading regression PASS; parse suite 27 tests PASS; P0.015 scenario trigger real-mode gate 1/1 + home/parse 54 tests PASS; verify PASS -->
- [x] 4.12 Scenario browser gate hardening: `E2E.P0.015` trigger 必须运行 `frontend/tests/pixel-parity/parse.spec.ts` 的 ready-response Playwright gate；fixture-backed ready `getTargetJob` 响应下截图 `route-parse` loading DOM，断言 `parse-basics-title` 在 loading window 内不存在，tick 完成后才出现；verify.sh 必须 grep browser gate marker 与 screenshot bytes。 <!-- evidence: 2026-05-24 P0.015 trigger includes Playwright parse.spec ready-response browser gate + screenshotBytes marker; verify PASS -->
- [x] 4.13 P0.016 route/context remediation: `ParseScreen` Confirm 必须复用 `interviewContextFromTargetJob(targetJob)`；已登录 navigate 与未登录 `requestAuth(pendingAction)` 均携带 `targetJobId` / `jobId` / `jdId` / `planId` / `resumeVersionId` / `roundId` / `roundName`；`E2E.P0.016` trigger 必须运行 Playwright browser gate，点击 Confirm 后验证 `/workspace` query、`workspace-missing-resume` DOM 与 screenshot bytes marker，verify.sh 必须 grep 完整 contextKeys marker。 <!-- evidence: 2026-05-24 Red: ParseEdit/AuthGate failed missing jobId/roundName; Green: focused ParseEdit/AuthGate 13 tests PASS; frontend build PASS; focused Playwright parse confirm gate desktop/mobile PASS screenshotBytes=20243/83490; P0.016 setup→trigger→verify→cleanup PASS -->
- [x] 4.14 Same-route `targetJobId` switch remediation: 同一 mounted `ParseScreen` 从 preview 收到新的 `targetJobId` 时必须清空旧 `targetJob` / editable fields / hit toggles / error / pending ready state，回到 loading gate，并在 tick 完成后 hydrate 新 TargetJob；`ParseFlow.test.tsx` 必须用 `rerender` 覆盖旧 title 消失、新 loading DOM 出现、新 title 最终渲染。 <!-- evidence: 2026-05-24 Red: focused ParseFlow rerender regression stayed on old preview; Green: focused ParseFlow 7 tests PASS; parse suite 28 tests PASS; frontend build PASS; P0.015/P0.016 setup→trigger→verify PASS -->

## Phase 5: jd_match P1 Placeholder Shell

- [x] 5.1 新增 `frontend/src/app/screens/jd_match/JDMatchScreen.tsx`，按 `ui-design/src/screen-jd-match.jsx::JDMatchScreen` lines 244-300 复刻 hero + profile snapshot chip 静态版本（不连接真实 profile）+ 三 tab 标签（recommended / search / watchlist）；tab 内容区固定渲染 P1 placeholder（testid `jdmatch-placeholder`）+ 引用 `frontend-home-job-picks-and-parse/plans/002-jd-match-recommendations`（保留编号）；testid `jdmatch-hero-{label,title,sub}` + `jdmatch-profile-chip-{title,years,location,skills,sources}` + `jdmatch-tab-{recommended,search,watchlist}` + `jdmatch-placeholder` + `jdmatch-placeholder-cta`
<!-- verified: 2026-05-08 method=vitest JDMatchPlaceholder.test.tsx 6 tests PASS -->
- [x] 5.2 在 `frontend/src/app/App.tsx` route table 中绑定 `jd_match` → `<JDMatchScreen />`，替换 D1 PlaceholderScreen；Vitest 断言 TopBar `topbar-nav-jd_match` 高亮、route 渲染命中 `JDMatchScreen` 而非 PlaceholderScreen；`HomeScreen` Job Picks aux card 点击进入 `jd_match` 路由仍可达
<!-- verified: 2026-05-08 method=vitest route binding verified via JDMatchPlaceholder.test.tsx shell data-route-name assertion -->
- [x] 5.3 扩展 i18n `jdMatch.*` 命名空间（≤10 key：hero.label / hero.title / hero.sub、tab.recommended / tab.search / tab.watchlist、placeholder.copy / placeholder.cta zh/en）；i18n test 断言 zh/en 同步
<!-- verified: 2026-05-08 method=vitest localeFiles.test.ts + localeRuntime.test.tsx PASS; 16 jdMatch keys zh/en synced -->
- [x] 5.4 新增 `jd_match/JDMatchPlaceholder.test.tsx`：测 hero / profile chip / 三 tab 标签 DOM；测 placeholder 文案 zh/en；负向断言旧 prototype 业务 testid（`jdmatch-card-*` / `jdmatch-saved-search-*` / `jdmatch-watchlist-*` / `jdmatch-market-signal-*` / `jdmatch-search-bar` / `jdmatch-search-results` / `jdmatch-jd-detail-*` / `jdmatch-agent-status`）grep 0 命中；TopBar `topbar-nav-jd_match` 高亮断言
<!-- verified: 2026-05-08 method=vitest JDMatchPlaceholder.test.tsx 6 tests PASS (hero, profile chip, 3 tabs, placeholder, negative assertions) -->
- [x] 5.5 BDD-Gate: 验证 `E2E.P0.017` jd_match P1 placeholder smoke
<!-- verified: 2026-05-08 method=scenario P0.017 setup→trigger→verify→cleanup PASS -->

## Phase 6: 验证收口（pixel parity + scenario + regression rerun）

- [x] 6.1 新增 `frontend/tests/pixel-parity/home.spec.ts` 覆盖 desktop (1440×900) + mobile (390×844) 两 chromium project：DOM 锚点 + bounding box stays in viewport, no overlap + warm/light → dark → customAccent 三态切换 computed 颜色变化 + toHaveScreenshot baseline；mobile 断言 textarea card 不溢出、Recent mocks 网格自然成单列、aux cards 折叠
<!-- verified: 2026-05-08 method=playwright --list discovers 4 home tests × 2 viewports; execution requires pnpm build + pixel parity server -->
- [x] 6.2 新增 `frontend/tests/pixel-parity/parse.spec.ts` 覆盖 desktop + mobile：loading 4 步进度条与 footer DOM；preview Basic fields / Requirements 双列 / Hidden signals / Round assumptions / footer actions 锚点；mobile 断言 Requirements 折单列、Round assumptions grid 折单/双列；warm/light → dark → customAccent 三态可见变化
<!-- verified: 2026-05-08 method=playwright --list discovers 3 parse tests × 2 viewports; full fixture-backed parse flow deferred to E2E.P0.015/016 scenarios -->
- [x] 6.3 新增 `frontend/tests/pixel-parity/jd_match.spec.ts` 覆盖 desktop + mobile：hero + profile chip + 三 tab 标签 + placeholder DOM；负向断言旧业务 testid 0 命中；mobile 不溢出
<!-- verified: 2026-05-08 method=playwright --list discovers 3 jd_match tests × 2 viewports; negative testid assertions cover jdmatch-card-*, jdmatch-saved-search-*, jdmatch-watchlist-*, jdmatch-market-signal-*, jdmatch-search-bar -->
- [x] 6.4 `pnpm --filter @easyinterview/frontend test:pixel-parity` 在 D2/D3 现有 21 spec × 2 viewport = 42 项基础上累加 home/parse/jd_match 新增 spec；总数全 PASS
<!-- verified: 2026-05-08 method=playwright 68/68 PASS (34 specs × 2 viewports); baseline updated for home screen changes -->
- [x] 6.5 派生 4 个 scenario 目录 `test/scenarios/e2e/p0-014-home-default-render/` `p0-015-jd-import-and-parse/` `p0-016-parse-confirm-to-workspace/` `p0-017-jd-match-placeholder/`，每个含 `README.md`（§6 baseline + §7 离线限制）+ `scripts/{setup,trigger,verify,cleanup}.sh`，按 `test/scenarios/README.md` + `test/scenarios/e2e/README.md` 规范实现；verify 脚本断言对应 testid 命中、retired-entry grep 0 命中、新增 spec 全 PASS marker
<!-- verified: 2026-05-08 method=scenario P0.014/P0.015/P0.016/P0.017 setup→trigger→verify→cleanup PASS; P0.006 verify repaired to current 68-test parity suite -->
- [x] 6.6 `test/scenarios/e2e/INDEX.md` P0 表追加 4 行（007 home 默认渲染 / 008 JD 导入与解析 / 009 Parse 确认进 workspace / 010 jd_match P1 placeholder smoke），关联需求列指向 `frontend-home-job-picks-and-parse C-1～C-10`，状态 Ready，执行方式 automated
<!-- verified: 2026-05-08 method=fs INDEX.md updated with P0.014-P0.017 rows -->
- [x] 6.7 Regression 重跑：`E2E.P0.001 / 002 / 004 / 005 / 006` 全部 setup→trigger→verify→cleanup PASS
<!-- verified: 2026-05-08 method=scenario P0.001/P0.002/P0.004/P0.005/P0.006 setup→trigger→verify→cleanup PASS; P0.006 Playwright 68/68 PASS -->
- [x] 6.8 全量验证：`pnpm --filter @easyinterview/frontend test`、`pnpm --filter @easyinterview/frontend typecheck`、`pnpm --filter @easyinterview/frontend build` 全 PASS；`make build` 全 PASS
<!-- verified: 2026-05-08 method=vitest frontend 52 files / 324 tests PASS; typecheck PASS; frontend build PASS; make build PASS; make validate-fixtures PASS -->
- [x] 6.9 文档与索引同步：`/sync-doc-index --fix-index` 把 `docs/spec/INDEX.md` 与 `docs/spec/frontend-home-job-picks-and-parse/plans/INDEX.md` 同步到 Header 当前；`check_md_links` 双 OK
<!-- verified: 2026-05-08 method=make docs-check zero drift; check_md_links double OK -->
- [x] 6.10 负向搜索：`frontend/src/` 内不 import `ui-design/src/data.jsx` / `window.EI_DATA` 0 命中；旧 prototype jd_match 业务 testid 与旧 route alias 0 命中（除 negative 断言文件与 `normalizeRoute` alias map）；JD raw text 不在 console.log/URL/localStorage/telemetry 0 命中；AI provider key / provider registry / prompt registry / AIClient / LLM endpoint / bypass generated client 的 parse fetch 0 命中（除测试负向断言与纯 UI 文案 fixture）
<!-- verified: 2026-05-08 method=rg runtime source negative searches passed; excluded README/test negative assertions and normalizeRoute alias map where applicable -->
- [x] 6.11 BDD-Gate: 验证 `E2E.P0.014` / `E2E.P0.015` / `E2E.P0.016` / `E2E.P0.017` 全部 setup→trigger→verify→cleanup PASS + D1+D2+D3 P0.001/002/004/005/006 regression PASS
<!-- verified: 2026-05-08 method=scenario P0.014/P0.015/P0.016/P0.017 plus P0.001/P0.002/P0.004/P0.005/P0.006 all setup→trigger→verify→cleanup PASS -->
- [x] 6.12 L2 remediation：真实 backend 联调闭环。新增 `frontend/src/api/targetJob.realApiMode.test.ts` 覆盖 `VITE_EI_API_MODE=real` 下 `listTargetJobs` / `createUploadPresign` / `importTargetJob` / `getTargetJob` / `updateTargetJob` 的真实 backend base URL、`credentials: "include"`、默认无 fixture `Prefer` header、3 个 side-effect `Idempotency-Key` 与 `GenerationProvenance` roundtrip；P0.014-P0.016 trigger/verify 必须先跑该 real-mode gate 再跑 fixture-backed UI variants；原地更新 plan/spec/BDD/scenario docs，删除 TargetJobs/import/parse 仍为 `not-yet-implemented` 的 stale 口径；重跑 P0.014-P0.016 + backend P0.010-P0.013 + upload focused route/handler tests + docs drift gates。 <!-- evidence: 2026-05-22 focused real-mode vitest PASS (1 file / 1 test); P0.014 PASS (real gate 1/1 + Home 3 files / 22 tests); P0.015 PASS (real gate 1/1 + Home/Parse import flow 7 files / 54 tests; existing React act warnings only); P0.016 PASS (real gate 1/1 + Parse confirm 2 files / 13 tests); backend P0.010/P0.011/P0.012/P0.013 all setup→trigger→verify→cleanup PASS; backend upload focused tests PASS (`go test ./cmd/api -run TestBuildUploadRoutesAlignsIdempotencyTTLWithPresignTTL -count=1`; `go test ./internal/upload/handler -run 'TestCreateUploadPresignReturnsCreatedResponse|TestCreateUploadPresignIdempotencyReplayAndTTL' -count=1`) -->

## Phase 7: Parse 简历绑定强制门禁（2026-06-30 修订）

- [x] 7.1 新增 `frontend/src/app/screens/parse/ParseResumeBinding.test.tsx` 红灯：ready Parse preview 必须调用 `listResumes`，渲染 `parse-launch`、`parse-resume-binding`、`parse-action-save-plan`、`parse-action-start-interview`；有 ready 简历时不得默认选中，Save/Start 在用户显式选择前必须 disabled。
  - Evidence 2026-06-30: Red `CI=true COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend test src/app/screens/parse/ParseResumeBinding.test.tsx` failed with `expected "listResumes" to be called 1 times, but got 0 times`; Green same command passed after `ParseScreen` added resume binding.
- [x] 7.2 实现 Parse resume binding：`ParseScreen` 读取 `listResumes`，过滤 `parseStatus=ready` 且未 archived 的简历，不默认选中；渲染简历绑定卡、显式选择列表 / 弹窗锚点和创建简历入口；无 ready 简历或读取失败时禁用 `立即面试` 与 `仅保存规划`，点击 `parse-resume-create` 导航 `resume_versions` `{ flow: "create" }`。
  - Evidence 2026-06-30: `CI=true COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend test src/app/screens/parse/ParseResumeBinding.test.tsx` passed after explicit-selection remediation, covering ready list no-default disabled state, user click enablement, empty-state disabled actions, and Start handoff with the clicked resume id.
- [x] 7.3 修复 Parse handoff：用户显式选择简历后，`仅保存规划` 保存 `updateTargetJob` 后进入 `workspace` 并携带真实 `resumeId`；`立即面试` 保存同一编辑字段后进入 `workspace` 并携带 `autoStartPractice=1`，由现有 workspace `useStartPractice` 链路创建 session 后进入 `practice`；focused tests 反向断言 `resume-unbound` 不在成功 params 中，未登录且无 verified ready 简历时不得产生成功 pendingAction。
  - Evidence 2026-06-30: `CI=true COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend test src/app/screens/parse` passed 6 files / 32 tests, including Save plan real `resumeId`, Start interview `autoStartPractice=1`, and unauthenticated disabled no-handoff negative coverage.
- [x] 7.4 BDD-Gate: 修订并验证 `E2E.P0.016`：trigger/verify/README/expected outcome 证明 Parse 成功出口不再渲染 `workspace-missing-resume`，并拒绝 `resume-unbound` 成功 marker。
  - Evidence 2026-06-30: `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/setup.sh`, `trigger.sh`, `verify.sh`, `cleanup.sh` all PASS. Trigger includes real API gate, focused Parse Vitest, frontend build, and Playwright desktop/mobile Save/Start browser gates with real ready `resumeId`.
