# E2E.P0.083 Resume Create Preview Save + 422 + Home/Workspace CTA Handoff

> **场景 ID**: E2E.P0.083
> **执行方式**: automated (vitest jsdom)
> **隔离级别**: in-process (vitest worker)
> **状态**: Ready

## 1 Given

- Fixture-backed mock-first client：`Uploads/createUploadPresign.json default`、`Resumes/registerResume.json default`、`Resumes/getResume.json default`、`Resumes/updateResume.json default / idempotency-replay / validation-error-422`
- 用户：未登录 → 登录态，lang 默认

## 2 When

- Home `还没有简历？1 分钟创建` CTA click → 路由到 `resume_versions?flow=create`
- 走完 Upload/Paste tab → register → polling → PreviewStage
- 点击保存触发 `updateResume`
- 模拟 422 → inline error
- Workspace `WorkspaceMissingResumeState` CTA click → 同样路由到 ResumeCreateFlow

## 3 Then

- `resume-preview-confirm` testid 命中；草稿主体渲染 identity / summary / experience / projects / skills / education
- `updateResume` 请求带 `Idempotency-Key` 与 `Accept-Language` header
- 成功路径：toast + nav 回 list（不携带 structuredProfile 字段进 URL 或 localStorage）
- 422 路径：inline error 渲染 + 不 nav
- Home / Workspace CTA：未登录态显示 auth gate；登录恢复后渲染 ResumeCreateFlow
- pendingAction 只携带 `{ flow: 'create', createMode? }`，不携带 rawText / file binary

## 4 Verification Entry

`scripts/trigger.sh` 调用：

- `src/app/screens/resume-workshop/create/PreviewStage.test.tsx`
- `src/app/screens/resume-workshop/create/ResumeCreateFlow.test.tsx`
- `src/app/screens/resume-workshop/create/adapters/mapParsedSummaryToStructuredProfileDraft.test.ts`
- `src/app/screens/resume-workshop/create/CreateFlowIntegration.test.tsx`
- `src/app/screens/resume-workshop/ResumeWorkshopAuthGate.test.tsx`

`scripts/verify.sh` 校验 trigger.log 内：

- `Test Files +\d+ passed` 匹配
- 关联 test file 名称命中
- updateResume / 422 / guided-negative / Home CTA / Workspace CTA / auth pendingAction 关键 case 名称命中

## 5 fixture / mock baseline

- `updateResume.json` scenario 已落地，本场景消费

## 6 baseline

- PreviewConfirm DOM + sidebar cards + 确认/返回 CTA + 三种 outcome 分支断言
- Home / Workspace CTA → CreateFlow 路由 / auth gate 行为
- pendingAction params 集合断言

## 7 离线限制 / mock-first 标注

- 不真实启动 dev stack；所有 fetch 通过 fixture
- 不真实启动 dev stack；updateResume 通过 fixture/mock spy 返回或 reject。
- `method=mock-fixture-client` 明确标注
