# E2E.P0.084 Flat Resume UI Regression

> **场景 ID**: E2E.P0.084
> **执行方式**: automated (vitest jsdom)
> **隔离级别**: in-process (vitest worker)
> **状态**: Ready

## 1 Given

- Fixture-backed mock-first client：`Resumes/listResumes.json default` +
  `Resumes/getResume.json default` + current flat save/tailor fixtures used by
  `ResumeWorkshopScreen` / `ResumeDetailView` / `ResumeRewritesTab` /
  `PreviewStage`。
- 用户：未登录 → 登录态，lang 默认。
- Non-current form and operation tokens remain absent from runtime source.

## 2 When

- 未登录访问 Resume Workshop/detail/create 路由 → 显示 auth gate。
- 登录态渲染 flat `ResumeWorkshopScreen`、`ResumeDetailView`、current
  `ResumeRewritesTab` 与 `PreviewStage`。
- Rewrites accept-only save modal覆盖 overwrite / save-as-new 分支。
- Source grep 检查 non-current operation token 不回流。

## 3 Then

- pendingAction 不携带 non-current form draft 或 wire 字段。
- `ResumeBranchFlow`、`branchResumeVersion`、`seedStrategy`、
  `acceptResumeTailorSuggestion`、`rejectResumeTailorSuggestion`、
  `updateResumeVersion` 在 runtime source 中 0 命中。
- Flat list/detail/create/rewrites surfaces stay functional under Vitest.
- Non-current tailor mode `(inline|rewrite|mirror)` 0 命中；prototype import
  `ui-design/src/(data|screen-resume-workshop)` 0 命中。

## 4 Verification Entry

`scripts/trigger.sh` 通过 Vitest 调用：

- `src/app/screens/resume-workshop/ResumeWorkshopScreen.test.tsx`
- `src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx`
- `src/app/screens/resume-workshop/tabs/ResumeRewritesTab.test.tsx`
- `src/app/screens/resume-workshop/create/PreviewStage.test.tsx`
- `src/app/screens/resume-workshop/ResumeWorkshopAuthGate.test.tsx`

## 5 Output

- `.test-output/e2e/p0-084-resume-flat-ui-regression/trigger.log` Vitest pass output。
- verify.sh 断言 trigger.log 含 vitest RUN 标记 + `Test Files .* passed` + `Tests .* passed`，并显式 grep 每个 spec 文件被执行。

## 6 Baseline

- `make codegen-check` 已通过的 generated client, with non-current operations
  absent.
- Current flat resume fixtures: `listResumes` / `getResume` /
  `updateResume` / `duplicateResume` / `requestResumeTailor`.

## 7 离线限制

本场景纯 fixture-backed Vitest 路径，无需 Docker Compose / Kind 或外网；离线运行 PASS。

## 8 方法标注

`method=fixture-backed-frontend`。Backend non-current-route evidence is covered by
P0.074/P0.079.
