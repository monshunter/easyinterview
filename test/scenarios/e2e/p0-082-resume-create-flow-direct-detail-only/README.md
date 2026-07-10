# E2E.P0.082 Resume Create Flow Direct Detail Only

> **场景 ID**: E2E.P0.082
> **执行方式**: automated (vitest jsdom)
> **隔离级别**: in-process (vitest worker)
> **状态**: Ready

## 1 Given

- Fixture-backed mock-first client：`Resumes/registerResume.json default`
- 用户已登录 lang=zh-CN

## 2 When

- Paste tab → submit → `registerResume` 成功
- Create-flow source negative grep 扫描 parser / preview component names
- Locale and runtime source 扫描解析/预览确认文案缺席

## 3 Then

- `ResumeCreateFlow` register 成功后直接打开 `resume_versions?resumeId=<id>`
- `resume-parse-flow` / `resume-preview-confirm` 不渲染
- `ResumeParseFlow`、`ParsingStage`、`PreviewStage`、`ResumePreviewConfirm` 不再出现在 create-flow runtime source
- 解析失败/重试页面相关测试文件不再作为 gate 运行

## 4 Verification Entry

`scripts/trigger.sh` 调用：

- `src/app/screens/resume-workshop/create/ResumeCreateFlow.test.tsx`
- `src/app/screens/resume-workshop/create/CreateFlowScopeNegative.test.ts`

`scripts/verify.sh` 校验 trigger.log 内：

- `Test Files +\d+ passed` 匹配
- 关联 test file 名称命中
- direct navigation / parser absence negative 关键 case 命中

## 5 fixture / mock baseline

- 当前 create flow 不再依赖 `getResume` polling fixture

## 6 baseline

- Direct-open DOM absence and create-flow source negative grep 在 Vitest 测试中断言

## 7 离线限制 / mock-first 标注

- 不真实启动 dev stack；所有 fetch 通过 mock transport
- `method=mock-fixture-client` 明确标注
