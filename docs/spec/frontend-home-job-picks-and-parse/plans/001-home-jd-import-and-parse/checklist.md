# 001 Home + JD Import + Parse Checklist

> **版本**: 2.5
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 1: Home 当前入口

- [x] 1.1 Home 源级复刻当前 `ui-design/src/screen-home.jsx::HomeScreen`：Hero label/title、JD 输入卡、输入卡底部 upload/URL source actions、ready 简历下拉框、创建简历入口、提交区、最近 3 张模拟面试卡片和 More handoff。
- [x] 1.2 Home 使用 generated client 调 `listResumes`、`listTargetJobs`、`createUploadPresign`、`importTargetJob`；paste/file/URL source discriminator、side-effect idempotency key、错误态和 pending import continuation 均有 focused Vitest 覆盖。
- [x] 1.3 Home import 前必须显式选择 ready 简历；成功进入 `parse` 时 params 携带真实 `resumeId`。
- [x] 1.4 BDD-Gate: `E2E.P0.014` 覆盖默认渲染、empty/one/twelve-plus fixtures、3-card cap、More handoff、theme/i18n 和 source/resume/submit layout。
- [x] 1.5 BDD-Gate: `E2E.P0.015` 覆盖 paste/upload/URL import、4xx/failed path、privacy gate、generated client request contract 和 real-mode generated-client preflight。

## Phase 2: Parse 当前确认与 handoff

- [x] 2.1 Parse 源级复刻当前 `ui-design/src/screens-p0-complete.jsx::ParseScreen`：4-step loading、preview、failed state、editable basics、requirements、hidden signals、round assumptions、resume binding 和 footer actions。
- [x] 2.2 Parse 使用 generated client 调 `getTargetJob`、`listResumes`、`updateTargetJob`；polling、same-route target switch、partial update body、idempotency key、failed state 与 privacy gate 均有 focused Vitest 覆盖。
- [x] 2.3 Parse 继承有效 route `resumeId`；缺失或无效时 Save/Start disabled，直到用户选择 ready 简历或进入创建流程。
- [x] 2.4 Save plan 进入 `workspace`；Start interview 进入 `workspace(autoStartPractice=1)`；两条路径都必须携带真实 `resumeId`。
- [x] 2.5 BDD-Gate: `E2E.P0.016` 覆盖 route resume inheritance、explicit resume selection、Save/Start browser gates、request body schema、auth continuation 和 privacy checks。

## Phase 3: 收口验证

- [x] 3.1 `validate_context.py frontend-home-job-picks-and-parse/001 frontend` 通过。
- [x] 3.2 Focused Home/Parse Vitest、frontend typecheck 与 `make validate-fixtures` 通过。
- [x] 3.3 `E2E.P0.014`、`E2E.P0.015`、`E2E.P0.016` 的 `setup -> trigger -> verify -> cleanup` 通过。
- [x] 3.4 `sync-doc-index --check`、`make docs-check`、`git diff --check` 和 `make lint-core-loop-pruning-surface` 通过。
