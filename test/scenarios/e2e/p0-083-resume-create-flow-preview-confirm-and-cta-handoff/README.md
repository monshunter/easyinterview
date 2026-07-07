# E2E.P0.083 Resume Create Direct Handoff + CTA

> **场景 ID**: E2E.P0.083
> **执行方式**: automated (vitest jsdom)
> **隔离级别**: in-process (vitest worker)
> **状态**: Ready

## 1 Given

- Fixture-backed mock-first client：`Uploads/createUploadPresign.json default`、`Resumes/registerResume.json default`
- 用户：未登录 → 登录态，lang 默认

## 2 When

- Home `还没有简历？1 分钟创建` CTA click → 路由到 `resume_versions?flow=create`
- 走完 Upload/Paste tab → register success → 直接打开 `resume_versions?resumeId=<id>`
- Workspace `WorkspaceMissingResumeState` CTA click → 同样路由到 ResumeCreateFlow

## 3 Then

- `resume-preview-confirm` 不渲染；create flow 不调用 `updateResume`
- 成功路径：nav 到 detail（不携带 rawText / structuredProfile 字段进 URL 或 localStorage）
- Home / Workspace CTA：未登录态显示 auth gate；登录接续后渲染 ResumeCreateFlow
- pendingAction 只携带 `{ flow: 'create', createMode? }`，不携带 rawText / file binary

## 4 Verification Entry

`scripts/trigger.sh` 调用：

- `src/app/screens/resume-workshop/create/ResumeCreateFlow.test.tsx`
- `src/app/screens/resume-workshop/create/UploadTab.test.tsx`
- `src/app/screens/resume-workshop/create/CreateFlowIntegration.test.tsx`
- `src/app/screens/resume-workshop/ResumeWorkshopAuthGate.test.tsx`
- `src/app/screens/workspace/WorkspaceHandoff.test.tsx`

`scripts/verify.sh` 校验 trigger.log 内：

- `Test Files +\d+ passed` 匹配
- 关联 test file 名称命中
- Home CTA / Workspace CTA / auth pendingAction / direct detail 关键 case 名称命中

## 5 fixture / mock baseline

- 本场景不再消费 `updateResume.json`

## 6 baseline

- Home / Workspace CTA → CreateFlow 路由 / auth gate 行为
- CreateFlow register success → detail route
- pendingAction params 集合断言

## 7 离线限制 / mock-first 标注

- 不真实启动 dev stack；所有 fetch 通过 fixture
- `method=mock-fixture-client` 明确标注
