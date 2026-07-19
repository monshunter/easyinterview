# Frontend Resume Workshop Create Flow Checklist

> **版本**: 1.21
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: Create Flow shell

- [x] 1.1 `flow=create` renders `ResumeCreateFlow`.
- [x] 1.2 CreateFlow keeps upload / paste tabs only.
- [x] 1.3 Auth pending action carries route state only and excludes raw resume content.

## Phase 2: Upload / paste registration

- [x] 2.1 Upload path validates file shape, obtains presign data, performs browser PUT, then calls `registerResume`.
- [x] 2.3 Side-effect requests include `Idempotency-Key`; polling requests do not.

## Phase 3: Direct-to-detail navigation

- [x] 3.1 Upload registration success navigates directly to `resume_versions?resumeId=<id>` and does not render `resume-parse-flow` / `resume-preview-confirm`.
- [x] 3.2 Paste registration success navigates directly to `resume_versions?resumeId=<id>` and does not render parser animation, preview confirm, or call create-flow `updateResume`.
- [x] 3.3 Direct navigation does not persist raw resume content into URL, pending action, storage or logs.

## Phase 4: Parser / preview-confirm absence

- [x] 4.1 `ResumeCreateFlow` no longer imports or renders `ParsingStage`, `ResumeParseFlow`, `PreviewStage`, or `ResumePreviewConfirm`.
- [x] 4.2 Create-flow code tests no longer assert parser/preview-confirm positive behavior；当前 BDD 仅保留行为合同且无真实 E2E owner。
- [x] 4.3 Source negative scan fails on user-visible copy for “正在阅读你的原始内容 / 结构化草稿如下 / Confirm and save resume” inside current create-flow runtime.

## Phase 5: BDD / integration gates

- [x] 5.4 `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/create` PASS.
- [x] 5.5 `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/ResumeWorkshopScreen.test.tsx src/app/screens/resume-workshop/fixture-parity.test.ts` PASS.

## Phase 6: Resume module UX optimization

- [x] 6.1 HISTORICAL-SUPERSEDED: 删除 `UploadTab` 的本地 2MiB truth；当前 Phase 13 读取 required runtime field，并仅以小型 metadata focused test 验证 overflow 不触发 presign/register。
- [x] 6.2 `ResumeCreateFlow` 删除右侧“会保存什么 / 接下来”sidebar，静态原型与正式前端 DOM 均不再出现对应 testid/copy；验证: `ResumeCreateFlow.test.tsx` + `ui-design` source grep。
- [x] 6.3 上传/粘贴注册成功后仍只导航到 `resume_versions?resumeId=<id>`，等待/成功/失败显示由 `ResumeDetailView` owner 接管，create-flow 不恢复 preview confirm 或 `updateResume`；验证: create focused tests。

## Phase 7: Resume upload source format support

- [x] 7.1 `UploadTab` 仅接受 PDF / Markdown / TXT，`.docx` 在本地校验中被拒绝且不会触发 presign/register；文案、`accept` 和静态原型同步；验证: `UploadTab.test.tsx` + UI source grep。

## Historical Phase 8: Home existing resume selection regression

The checked assertions below record the 2026-07-08 full-Resume list contract. Active 001 Phase 19 supersedes them with `ResumeSummary.parseStatus/hasReadableContent`; Parse/Workspace detail no longer shares this list predicate.

- [x] 8.2 Home 和 Parse 复用同一可选简历判断：`ready` 或已有 `parsedTextSnapshot` / `originalText` / structured profile 的非归档简历可选；无可读证据的 queued/processing 简历仍不可选。
- [x] 8.3 截图闭环：浏览器打开 Home，展开或聚焦已有简历选择控件，截图证明选项可见且不再显示 `还没有可用简历`。<!-- verified: 2026-07-08 method=playwright-screenshot artifact=.test-output/screenshots/home-resume-picker-fixed-2026-07-08.png -->

## Phase 9: Zero-reference stage type removal

- [x] 9.1 RED/GREEN: create-flow source gate detects and then rejects the unconsumed `CreateStage` declaration while retaining `data-stage="input"`.
- [x] 9.2 Focused create-flow tests provide development feedback without a replacement stage abstraction；阶段单测完成由仓库根 `make test` 承接，typecheck 作为独立代码 gate。

## Phase 10: Prototype create-flow call-surface pruning

- [x] 10.1 Add a prototype prop-consumption contract and prove RED while `ResumeCreateFlow` and its caller still carry unread `nav`.
  <!-- verified: 2026-07-10 method=resume-create-call-surface-red evidence="UI contract ran 44 tests: the new create-flow callback contract failed on the existing ResumeCreateFlow.nav parameter while the prior 43 tests passed; retained assertions pin upload/paste modes plus onBack and onCreateResume ownership." -->
- [x] 10.2 Delete the unread prop and caller argument; verify Babel inventory reports zero unread `ResumeCreateFlow` props while `onBack` and `onCreateResume` remain intact.
  <!-- verified: 2026-07-10 method=resume-create-call-surface-green evidence="Removed only ResumeCreateFlow.nav and its matching child argument. UI contract passes 44/44; Babel binding inventory scans 11 prototype files and reports unusedProps=[] and unusedState=[] while retaining onBack/onCreateResume assertions." -->
  <!-- red: 2026-07-10 method=ui-contract+browser evidence="UI contract ran 44 tests with only the direct-detail contract failing. Browser reproduction showed Save and open returned to the three-item list because the flow-bearing App key remounted ResumeWorkshopScreen and discarded createdResumes." -->
  <!-- red-2: 2026-07-10 method=ui-contract+browser evidence="After preserving the screen instance, browser replay from the workshop-local New resume action showed params.flow stayed empty and the effect could not exit create mode. The strengthened contract requires addCreatedResume to switch to list before resumeId navigation." -->

## Phase 11: zero-consumer ghost CTA CSS pruning

- [x] 11.1 Add a create-flow-owned source RED gate for the ghost CTA selector with no formal DOM or prototype consumer.
- [x] 11.2 Delete its base/variant/disabled CSS branches without aliases, placeholders or removal markers; retain the current accent CTA and disabled state.
  <!-- verified: 2026-07-10 method=create-flow-css-source-green evidence="CreateFlowScopeNegative passes 4/4; ghost is absent outside its negative literal, accent base/variant/disabled CSS remains, and UploadTab/PasteTab retain the two production consumers. CSS class reachability now reports only prototype-backed ei-scroll." -->

## Phase 12: accent CTA rule consolidation

- [x] 12.1 Add a source RED gate requiring one accent CTA rule with the complete layout, typography, interaction, color and border declarations.
- [x] 12.2 Merge both declaration blocks into one rule without changing the disabled rule or component DOM.
  <!-- verified: 2026-07-10 method=create-accent-css-cascade-green evidence="CreateFlowScopeNegative passes 5/5; exactly one accent CTA rule contains all effective layout, typography, interaction, background, border and color declarations, while the disabled rule and two component consumers remain unchanged." -->
- [x] 12.3 Run focused CreateFlow, full frontend, typecheck/build, owner/product contexts and docs/index/diff/pruning gates; then restore the owner to `completed`.

## Phase 13: Required runtime upload and paste guards

- [x] 13.1 RED/GREEN: UploadTab/PasteTab 消费 required RuntimeConfig fields 与共享 UTF-8 helper；小型 metadata/text limit 覆盖 zero request，UI DOM/style 不变。
- [x] 13.2 FALLBACK-GATE: required 子字段无 per-field fallback；仅整体 runtime source 不可用时保留既有 bootstrap fallback。

## Phase 14: Reference-aligned create canvas

- [x] 14.1 RED: component/source tests 固化 full-viewport create canvas、约 1470px desktop 内容面、Header/tab/dropzone/CTA/capability-chip DOM 与 390px containment。<!-- verified: 2026-07-19 method=vitest-red evidence="ResumeCreateVisual 2 expected failures on old narrow canvas and missing art/capabilities" -->
- [x] 14.2 GREEN: 重构 CreateFlow / UploadTab / PasteTab 与 owner CSS，保留 upload/paste 注册、runtime limit、错误恢复、隐私和 direct-detail 行为。<!-- verified: 2026-07-19 method=focused-vitest evidence="4 files / 21 tests PASS" -->
- [x] 14.3 BDD-Gate: `BDD.RESUME.CREATE.001` 在新视觉层级下继续覆盖 upload/paste、错误恢复与隐私；Chrome 1916×821 / 390×844 视图验收无横向溢出。<!-- verified: 2026-07-19 method=chrome-formal-local evidence="10900 real-mode frontend create flow: desktop shell/card width 1470px; mobile x=16 width=358px; header art hides at mobile; both viewports documentOverflow=0" -->
- [x] 14.4 REGRESSION-GATE: focused frontend、根 `make test`、typecheck/build、context/docs/index/diff 通过后恢复 completed。<!-- verified: 2026-07-19 method=focused+root-regression evidence="create focused 4 files/21 tests PASS; root Python 615/4615 subtests, Go all packages, frontend 131 files/1054 tests PASS; typecheck/build PASS" -->

## BDD Gate

- [x] BDD-Gate: `BDD.RESUME.CREATE.001` 由 [BDD checklist](./bdd-checklist.md) 关联 upload/paste create-flow owner behavior tests；不创建或声明真实 E2E PASS。
