# E2E.P0.083 Resume Create Preview Confirm + 409/422 + Home/Workspace CTA Handoff

> **场景 ID**: E2E.P0.083
> **执行方式**: automated (vitest jsdom)
> **隔离级别**: in-process (vitest worker)
> **状态**: Ready

## 1 Given

- Fixture-backed mock-first client：`Uploads/createUploadPresign.json default`、`Resumes/registerResume.json default`、`Resumes/getResume.json default`、`Resumes/confirmResumeStructuredMaster.json default / idempotency-replay / already-exists-409 / validation-422`、`Resumes/listResumeVersions.json`
- 用户：未登录 → 登录态，lang 默认

## 2 When

- Home `还没有简历？1 分钟创建` CTA click → 路由到 `resume_versions?flow=create`
- 走完 Upload tab → register → polling → PreviewConfirm
- 点击 "确认并保存 v1" 触发 `confirmResumeStructuredMaster`
- 同 IK 二次确认 → `idempotency-replay` → 同 ResumeVersion
- 新 IK + 同 asset 模拟 409 → `listResumeVersions` 查找 → nav 到已存在 master
- 模拟 422 → inline error
- Workspace `WorkspaceMissingResumeState` CTA click → 同样路由到 ResumeCreateFlow

## 3 Then

- `resume-preview-confirm` testid 命中；草稿主体渲染 identity / summary / experience / projects / skills / education
- `confirmResumeStructuredMaster` 请求带 `Idempotency-Key` 与 `Accept-Language` header
- 成功路径：toast `已保存 v1 主版本` + nav 回 list（不携带 structuredProfile 字段进 URL 或 localStorage）
- 409 路径：toast `已存在主版本 · 跳转查看` + nav 到 `resume_versions?versionId=...&tab=preview`
- 422 路径：inline error 渲染 + 不 nav
- Home / Workspace CTA：未登录态显示 auth gate；登录恢复后渲染 ResumeCreateFlow
- pendingAction 只携带 `{ flow: 'create', createMode? }`，不携带 rawText / file binary / guidedAnswers

## 4 Verification Entry

`scripts/trigger.sh` 调用：

- `src/app/screens/resume-workshop/create/PreviewStage.test.tsx`
- `src/app/screens/resume-workshop/create/hooks/useResumeStructuredMasterConfirm.test.tsx`
- `src/app/screens/resume-workshop/create/adapters/mapParsedSummaryToStructuredProfileDraft.test.ts`
- `src/app/screens/resume-workshop/create/CreateFlowIntegration.test.tsx`
- `src/app/screens/resume-workshop/ResumeWorkshopAuthGate.test.tsx`

`scripts/verify.sh` 校验 trigger.log 内：

- `Test Files +\d+ passed` 匹配
- 关联 test file 名称命中
- happy / replay / 409 / 422 / Home CTA / Workspace CTA / auth pendingAction 关键 case 名称命中

## 5 fixture / mock baseline

- `confirmResumeStructuredMaster.json` 四 scenario 已落地，本场景消费
- `listResumeVersions` 用 mock 返回单 master 行模拟 409 fallback

## 6 baseline

- PreviewConfirm DOM + sidebar cards + 确认/返回 CTA + 三种 outcome 分支断言
- Home / Workspace CTA → CreateFlow 路由 / auth gate 行为
- pendingAction params 集合断言

## 7 离线限制 / mock-first 标注

- 不真实启动 dev stack；所有 fetch 通过 fixture
- 不模拟真实 backend confirmResumeStructuredMaster；mock spy 返回 / reject 各 scenario
- `method=mock-fixture-client` 明确标注
