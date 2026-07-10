# Frontend Resume Workshop Create Flow Checklist

> **版本**: 1.11
> **状态**: completed
> **更新日期**: 2026-07-10

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

## Phase 8: Home existing resume selection regression

- [x] 8.1 BDD-Gate: E2E.P0.084 用 focused Home/Parse regression tests 覆盖 `listResumes` 返回非归档且已有可读正文的简历时，首页下拉不得为空或禁用。<!-- verified: 2026-07-08 method=vitest tests=HomeResumeSelection.test.tsx,ParseResumeBinding.test.tsx -->
- [x] 8.2 Home 和 Parse 复用同一可选简历判断：`ready` 或已有 `parsedTextSnapshot` / `originalText` / structured profile 的非归档简历可选；无可读证据的 queued/processing 简历仍不可选。<!-- verified: 2026-07-08 method=vitest tests=selectableResume.test.ts -->
- [x] 8.3 截图闭环：浏览器打开 Home，展开或聚焦已有简历选择控件，截图证明选项可见且不再显示 `还没有可用简历`。<!-- verified: 2026-07-08 method=playwright-screenshot artifact=.test-output/screenshots/home-resume-picker-fixed-2026-07-08.png -->

## Phase 9: Zero-reference stage type removal

- [x] 9.1 RED/GREEN: create-flow source gate detects and then rejects the unconsumed `CreateStage` declaration while retaining `data-stage="input"`.<!-- verified: 2026-07-10 method=vitest-red-green evidence="RED failed only on ResumeCreateFlow.tsx; GREEN passed 3/3 after deleting the declaration." -->
- [x] 9.2 Focused create-flow tests and frontend typecheck pass without a replacement stage abstraction.<!-- verified: 2026-07-10 method=vitest+typecheck evidence="Create-flow passed 6 files/32 tests; frontend typecheck passed; production source inventory returned zero CreateStage matches." -->
