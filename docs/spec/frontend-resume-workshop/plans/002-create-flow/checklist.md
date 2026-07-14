# Frontend Resume Workshop Create Flow Checklist

> **版本**: 1.19
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: Create Flow shell

- [x] 1.1 `flow=create` renders `ResumeCreateFlow`.
- [x] 1.2 CreateFlow keeps upload / paste tabs only.
- [x] 1.3 Auth pending action carries route state only and excludes raw resume content.

## Phase 2: Upload / paste registration

- [x] 2.1 Upload path validates file shape, obtains presign data, performs browser PUT, then calls `registerResume`.
- [x] 2.2 Paste path calls `registerResume` with paste payload after non-empty validation and sends a neutral source title; raw resume first line must not be submitted or displayed as the resume name. <!-- verified: 2026-07-07 method=vitest+scenario tests=ResumeCreateFlow.test.tsx,E2E.P0.081 -->
- [x] 2.3 Side-effect requests include `Idempotency-Key`; polling requests do not.

## Phase 3: Direct-to-detail navigation

- [x] 3.1 Upload registration success navigates directly to `resume_versions?resumeId=<id>` and does not render `resume-parse-flow` / `resume-preview-confirm`.
- [x] 3.2 Paste registration success navigates directly to `resume_versions?resumeId=<id>` and does not render parser animation, preview confirm, or call create-flow `updateResume`.
- [x] 3.3 Direct navigation does not persist raw resume content into URL, pending action, storage or logs.

## Phase 4: Parser / preview-confirm absence

- [x] 4.1 `ResumeCreateFlow` no longer imports or renders `ParsingStage`, `ResumeParseFlow`, `PreviewStage`, or `ResumePreviewConfirm`.
- [x] 4.2 Create-flow tests and scenario scripts no longer execute parser/preview-confirm positive tests.
- [x] 4.3 Source negative scan fails on user-visible copy for “正在阅读你的原始内容 / 结构化草稿如下 / Confirm and save resume” inside current create-flow runtime.

## Phase 5: BDD / integration gates

- [x] 5.1 BDD-Gate: E2E.P0.081 create-flow upload/paste direct-to-detail happy path is maintained.
- [x] 5.2 Absence gate: E2E.P0.082 no longer validates parser failure UI; verify script records parser flow as absent from current create flow.
- [x] 5.3 BDD-Gate: E2E.P0.083 Home CTA direct-create handoff is maintained without preview confirm.
- [x] 5.4 `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/create` PASS.
- [x] 5.5 `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/ResumeWorkshopScreen.test.tsx src/app/screens/resume-workshop/fixture-parity.test.ts` PASS.

## Phase 6: Resume module UX optimization

- [x] 6.1 `UploadTab` 将默认文件大小上限改为 2MiB，并在本地校验中过大文件不触发 presign/register；验证: `UploadTab.test.tsx` focused red/green。<!-- verified: 2026-07-07 method=vitest command="corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/create/UploadTab.test.tsx" -->
- [x] 6.2 `ResumeCreateFlow` 删除右侧“会保存什么 / 接下来”sidebar，静态原型与正式前端 DOM 均不再出现对应 testid/copy；验证: `ResumeCreateFlow.test.tsx` + `ui-design` source grep。<!-- verified: 2026-07-07 method=vitest+source-grep tests=ResumeCreateFlow.test.tsx -->
- [x] 6.3 上传/粘贴注册成功后仍只导航到 `resume_versions?resumeId=<id>`，等待/成功/失败显示由 `ResumeDetailView` owner 接管，create-flow 不恢复 preview confirm 或 `updateResume`；验证: create focused tests。<!-- verified: 2026-07-07 method=vitest tests=ResumeCreateFlow.test.tsx,ResumeDetailView.test.tsx -->

## Phase 7: Resume upload source format support

- [x] 7.1 `UploadTab` 仅接受 PDF / Markdown / TXT，`.docx` 在本地校验中被拒绝且不会触发 presign/register；文案、`accept` 和静态原型同步；验证: `UploadTab.test.tsx` + UI source grep。<!-- verified: 2026-07-07 method=vitest+source-grep tests=UploadTab.test.tsx ui=ui-design/src/screen-resume-workshop.jsx -->

## Historical Phase 8: Home existing resume selection regression

The checked assertions below record the 2026-07-08 full-Resume list contract. Active 001 Phase 19 supersedes them with `ResumeSummary.parseStatus/hasReadableContent`; Parse/Workspace detail no longer shares this list predicate.

- [x] 8.1 BDD-Gate: E2E.P0.084 用 focused Home/Parse regression tests 覆盖 `listResumes` 返回非归档且已有可读正文的简历时，首页下拉不得为空或禁用。<!-- verified: 2026-07-08 method=vitest tests=HomeResumeSelection.test.tsx,ParseResumeBinding.test.tsx -->
- [x] 8.2 Home 和 Parse 复用同一可选简历判断：`ready` 或已有 `parsedTextSnapshot` / `originalText` / structured profile 的非归档简历可选；无可读证据的 queued/processing 简历仍不可选。<!-- verified: 2026-07-08 method=vitest tests=selectableResume.test.ts -->
- [x] 8.3 截图闭环：浏览器打开 Home，展开或聚焦已有简历选择控件，截图证明选项可见且不再显示 `还没有可用简历`。<!-- verified: 2026-07-08 method=playwright-screenshot artifact=.test-output/screenshots/home-resume-picker-fixed-2026-07-08.png -->

## Phase 9: Zero-reference stage type removal

- [x] 9.1 RED/GREEN: create-flow source gate detects and then rejects the unconsumed `CreateStage` declaration while retaining `data-stage="input"`.<!-- verified: 2026-07-10 method=vitest-red-green evidence="RED failed only on ResumeCreateFlow.tsx; GREEN passed 3/3 after deleting the declaration." -->
- [x] 9.2 Focused create-flow tests and frontend typecheck pass without a replacement stage abstraction.<!-- verified: 2026-07-10 method=vitest+typecheck evidence="Create-flow passed 6 files/32 tests; frontend typecheck passed; production source inventory returned zero CreateStage matches." -->

## Phase 10: Prototype create-flow call-surface pruning

- [x] 10.1 Add a prototype prop-consumption contract and prove RED while `ResumeCreateFlow` and its caller still carry unread `nav`.
  <!-- verified: 2026-07-10 method=resume-create-call-surface-red evidence="UI contract ran 44 tests: the new create-flow callback contract failed on the existing ResumeCreateFlow.nav parameter while the prior 43 tests passed; retained assertions pin upload/paste modes plus onBack and onCreateResume ownership." -->
- [x] 10.2 Delete the unread prop and caller argument; verify Babel inventory reports zero unread `ResumeCreateFlow` props while `onBack` and `onCreateResume` remain intact.
  <!-- verified: 2026-07-10 method=resume-create-call-surface-green evidence="Removed only ResumeCreateFlow.nav and its matching child argument. UI contract passes 44/44; Babel binding inventory scans 11 prototype files and reports unusedProps=[] and unusedState=[] while retaining onBack/onCreateResume assertions." -->
- [x] 10.3 Run UI contract, focused create-flow, P0.081, and static-browser upload/paste/back/create smoke; successful prototype creation must preserve the local asset while routing from `flow=create` to its waiting/ready detail. Then run full frontend, typecheck/build, owner contexts and docs/diff/pruning gates.
  <!-- red: 2026-07-10 method=ui-contract+browser evidence="UI contract ran 44 tests with only the direct-detail contract failing. Browser reproduction showed Save and open returned to the three-item list because the flow-bearing App key remounted ResumeWorkshopScreen and discarded createdResumes." -->
  <!-- red-2: 2026-07-10 method=ui-contract+browser evidence="After preserving the screen instance, browser replay from the workshop-local New resume action showed params.flow stayed empty and the effect could not exit create mode. The strengthened contract requires addCreatedResume to switch to list before resumeId navigation." -->
  <!-- verified: 2026-07-10 method=red-green+scenario+browser+frontend+owner-gates evidence="UI contract passed 44/44 after two focused RED/GREEN loops; Babel inventory across 11 prototype files returned unusedProps=[] and unusedState=[]; create-flow focused tests passed 6 files/32 tests; P0.081 passed real-mode 1/1 and 5 files/28 tests with setup/trigger/verify/cleanup; browser proved hash create, back, local New resume, Paste, 50ms waiting detail and 1.2s ready detail with no page errors and 200/304 assets; full frontend passed 137 files/841 tests, typecheck and build; both owner contexts, git diff check and pruning surface passed. BUG-0154 records the route-key state loss. No scenario environment restart or data cleanup was performed." -->

## Phase 11: zero-consumer ghost CTA CSS pruning

- [x] 11.1 Add a create-flow-owned source RED gate for the ghost CTA selector with no formal DOM or prototype consumer.
  <!-- verified: 2026-07-10 method=create-flow-css-source-red evidence="Focused CreateFlowScopeNegative ran 4 tests: all 3 existing source contracts passed and only the new ghost CTA ownership gate failed." -->
- [x] 11.2 Delete its base/variant/disabled CSS branches without aliases, placeholders or removal markers; retain the current accent CTA and disabled state.
  <!-- verified: 2026-07-10 method=create-flow-css-source-green evidence="CreateFlowScopeNegative passes 4/4; ghost is absent outside its negative literal, accent base/variant/disabled CSS remains, and UploadTab/PasteTab retain the two production consumers. CSS class reachability now reports only prototype-backed ei-scroll." -->
- [x] 11.3 Run focused CreateFlow/P0.081, full frontend, typecheck/build, owner/product contexts and docs/index/diff/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=resume-create-ghost-cta-css-pruning evidence="Source gate passes 4/4; CreateFlow owner passes 6 files/33 tests; P0.081 setup/trigger/verify/cleanup passes real-mode 1 plus focused 29; full frontend passes 136 files/842 tests; typecheck/build and both contexts pass. Ghost runtime inventory is zero, accent CTA retains two consumers, and CSS reachability reports only prototype-backed ei-scroll. Final docs/index/diff/pruning gates run during closeout. No Bug/retrospective report, environment restart or data cleanup was needed." -->

## Phase 12: accent CTA rule consolidation

- [x] 12.1 Add a source RED gate requiring one accent CTA rule with the complete layout, typography, interaction, color and border declarations.
  <!-- verified: 2026-07-10 method=create-accent-css-cascade-red evidence="Focused CreateFlowScopeNegative ran 5 tests: all 4 existing contracts passed and only the new unique accent CTA rule gate failed because two blocks remained." -->
- [x] 12.2 Merge both declaration blocks into one rule without changing the disabled rule or component DOM.
  <!-- verified: 2026-07-10 method=create-accent-css-cascade-green evidence="CreateFlowScopeNegative passes 5/5; exactly one accent CTA rule contains all effective layout, typography, interaction, background, border and color declarations, while the disabled rule and two component consumers remain unchanged." -->
- [x] 12.3 Run focused CreateFlow, full frontend, typecheck/build, owner/product contexts and docs/index/diff/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=create-accent-css-cascade-consolidation evidence="Source gate passes 5/5; CreateFlow owner passes 6 files/34 tests; full frontend passes 136 files/844 tests; typecheck/build and both contexts pass. Exactly one complete accent CTA rule remains, while the disabled rule and both production consumers are unchanged. Final docs/index/diff/pruning gates pass during closeout. No Bug/retrospective report, environment restart or data cleanup was needed." -->

## Phase 13: Runtime-configured upload and paste boundaries

- [x] 13.1 RED: runtime 10MiB/384KiB、UTF-8 limit/limit+1 与 zero-request tests 在旧 2MiB constant 下失败。
- [x] 13.2 GREEN: UploadTab/PasteTab 消费 AppRuntimeProvider 两字段和共享 byte helper；缺字段使用同值 code default，UI DOM/style 不变。
- [x] 13.3 BDD-Gate: P0.081 upload/paste boundary、DOCX、direct detail、privacy/recovery 当前证据通过。
- [x] 13.4 focused/full frontend、typecheck/build、OpenAPI/generated、parity、contexts/docs/diff 与旧 2MiB production-truth negative search 通过。
  <!-- verified: 2026-07-14 evidence="Frontend full 126 files/1018 tests, build and fresh P0.081 5 files/33 tests pass." -->
