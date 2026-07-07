# E2E.P0.081 Resume Create Flow Upload / Paste Happy Path

> **场景 ID**: E2E.P0.081
> **执行方式**: automated (vitest jsdom)
> **隔离级别**: in-process (vitest worker)
> **状态**: Ready

## 1 Given

- Fixture-backed mock-first client：`Uploads/createUploadPresign.json default`、`Resumes/registerResume.json default / paste-text`、`Auth/getRuntimeConfig.json`、`Auth/getMe.json authenticated`
- 用户：未登录 → 登录态切换 + lang 切换

## 2 When

- 选择 Upload tab → 通过 file input 选择 1KB `.pdf` → 触发 `createUploadPresign` + 浏览器 PUT + `registerResume`
- 选择 Paste tab → 填写 raw text → submit `registerResume` (sourceType=paste)
- 两条路径在 register 成功后直接导航到 `resume_versions?resumeId=<id>`

## 3 Then

- ResumeCreateFlow 渲染：`resume-create-flow` testid 命中
- 两个 tab DOM anchors 覆盖：`resume-create-tab-upload/-paste` + `data-active=true` 当前 tab；non-current guided tab/panel 不渲染
- Upload：`Idempotency-Key` header on presign + register，`fetch(uploadUrl, { method: 'PUT', body: file })` 调用形态
- Paste：textarea / submit disabled-when-empty / IK on register / title 使用中性来源标题，不能提交 raw resume 第一行作为可见名称
- Direct detail：register success 后不渲染 `resume-parse-flow` / `resume-preview-confirm`
- 隐私：rawText / parsedTextSnapshot / parsedSummary / file binary 不出现在 console / URL / pendingAction / localStorage / mock transport log
- 非当前入口 grep：`welcome|mistake|growth|drill|followup|STAR|experiences|voice|OnboardingScreen|onboarding=true|ResumeParseFlow|ParsingStage|PreviewStage|ResumePreviewConfirm` 0 命中
- prototype import grep：`ui-design/src/(data|screen-resume-workshop)` 0 命中
- mock harness 切换显式标注 `method=mock-fixture-client`

## 4 Verification Entry

`scripts/trigger.sh` 调用以下 Vitest test files 集中验证：

- `src/app/screens/resume-workshop/create/ResumeCreateFlow.test.tsx`
- `src/app/screens/resume-workshop/create/UploadTab.test.tsx`
- `src/app/screens/resume-workshop/create/hooks/useResumePresignUpload.test.tsx`
- `src/app/screens/resume-workshop/create/hooks/useResumeRegistration.test.tsx`
- `src/app/screens/resume-workshop/create/CreateFlowNonCurrentNegative.test.ts`

`scripts/verify.sh` 在 `.test-output/e2e/p0-081-resume-create-flow-upload-paste-guided-happy/trigger.log` 内执行：

- `Test Files +[0-9]+ passed` 匹配
- 关联 test file 名称命中
- privacy / non-current grep 命中 0

## 5 fixture / mock baseline

- 当前 fixture：`Uploads/createUploadPresign.json` 只有 `default` scenario；validation / IK replay 行为通过 Vitest mock client error path 覆盖（见 plan §6 R3）

## 6 baseline

- 复刻 [UI 真理源](../../../ui-design/src/screen-resume-workshop.jsx) `ResumeCreateFlow` 顶层 export
- Vitest 测试断言 ≥ 30 testid + ≥ 5 IK / Accept-Language header 断言

## 7 离线限制 / mock-first 标注

- 本场景**不**真实启动 A2 dev stack；所有 fetch 通过 `createFixtureBackedFetch` 拦截到 fixture
- 不真实 PUT 到对象存储；浏览器 PUT 通过全局 fetch spy 验证调用形态
- `method=mock-fixture-client` 已在 plan §6 R3 / R2 与本 README §3 显式标注
