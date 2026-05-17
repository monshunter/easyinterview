# E2E.P0.081 Resume Create Flow Upload / Paste / Guided Happy Path

> **场景 ID**: E2E.P0.081
> **执行方式**: automated (vitest jsdom)
> **隔离级别**: in-process (vitest worker)
> **状态**: Ready

## 1 Given

- Fixture-backed mock-first client：`Uploads/createUploadPresign.json default`、`Resumes/registerResume.json default / paste-text / guided-answers`、`Resumes/getResume.json default`、`Auth/getRuntimeConfig.json`、`Auth/getMe.json authenticated`
- Mock harness 在 fixture 未覆盖 `parseStatus` 多态时使用 deterministic attempt-aware stepping（[plan.md §6 R2](../../../docs/spec/frontend-resume-workshop/plans/002-create-flow-and-onboarding/plan.md#6-风险与应对) 显式声明）
- 用户：未登录 → 登录态切换 + lang 切换

## 2 When

- 选择 Upload tab → 通过 file input 选择 1KB `.pdf` → 触发 `createUploadPresign` + 浏览器 PUT + `registerResume`
- 选择 Paste tab → 填写 raw text → submit `registerResume` (sourceType=paste)
- 选择 Guided tab → 5 step 各填非空回答 → 最后一步 submit `registerResume` (sourceType=guided)
- 三条路径均进入 ParseFlow → polling `getResume` → transition

## 3 Then

- ResumeCreateFlow 渲染：`resume-create-flow` testid 命中
- 三 tab DOM anchors 覆盖：`resume-create-tab-upload/-paste/-guided` + `data-active=true` 当前 tab
- Upload：`Idempotency-Key` header on presign + register，`fetch(uploadUrl, { method: 'PUT', body: file })` 调用形态
- Paste：textarea / submit disabled-when-empty / IK on register
- Guided：5 step nav `resume-create-guided-step-{1..5}` + payload `{ recentRole, direction, proofProject, metrics, target }`
- ParseFlow：`resume-parse-flow` testid + 7-step ticker DOM
- 隐私：rawText / guidedAnswers / parsedTextSnapshot / parsedSummary / file binary 不出现在 console / URL / pendingAction / localStorage / mock transport log
- 旧入口 grep：`welcome|mistake|growth|drill|followup|STAR|experiences|voice|OnboardingScreen|onboarding=true` 0 命中
- prototype import grep：`ui-design/src/(data|screen-resume-workshop)` 0 命中
- mock harness 切换显式标注 `method=mock-fixture-client`

## 4 Verification Entry

`scripts/trigger.sh` 调用以下 Vitest test files 集中验证：

- `src/app/screens/resume-workshop/create/ResumeCreateFlow.test.tsx`
- `src/app/screens/resume-workshop/create/UploadTab.test.tsx`
- `src/app/screens/resume-workshop/create/PasteGuidedTab.test.tsx`
- `src/app/screens/resume-workshop/create/hooks/useResumePresignUpload.test.tsx`
- `src/app/screens/resume-workshop/create/hooks/useResumeRegistration.test.tsx`
- `src/app/screens/resume-workshop/create/ParsingStage.test.tsx`
- `src/app/screens/resume-workshop/create/CreateFlowLegacyNegative.test.ts`

`scripts/verify.sh` 在 `.test-output/e2e/p0-081-resume-create-flow-upload-paste-guided-happy/trigger.log` 内执行：

- `Test Files +[0-9]+ passed` 匹配
- 关联 test file 名称命中
- privacy / legacy grep 命中 0

## 5 fixture / mock baseline

- 当前 fixture：`Uploads/createUploadPresign.json` 只有 `default` scenario；validation / IK replay 行为通过 Vitest mock client error path 覆盖（见 plan §6 R3）
- `Resumes/getResume.json` 缺少 `queued / generating / failed` parseStatus；本场景使用 `processing` mock 模拟，retrospective 中提议补 fixture

## 6 baseline

- 复刻 [UI 真理源](../../../ui-design/src/screen-resume-workshop.jsx) `ResumeCreateFlow` / `ResumeParseFlow` 顶层 export
- Vitest 测试断言 ≥ 30 testid + ≥ 5 IK / Accept-Language header 断言

## 7 离线限制 / mock-first 标注

- 本场景**不**真实启动 A2 dev stack；所有 fetch 通过 `createFixtureBackedFetch` 拦截到 fixture
- 不真实 PUT 到对象存储；浏览器 PUT 通过全局 fetch spy 验证调用形态
- `method=mock-fixture-client` 已在 plan §6 R3 / R2 与本 README §3 显式标注
