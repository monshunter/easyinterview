# 001 Home + JD Import + Parse

> **版本**: 2.6
> **状态**: completed
> **更新日期**: 2026-07-09

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本计划交付当前 Home + Parse 新建模拟面试入口。用户从首页输入、上传或 URL 导入 JD，显式选择一份 ready 简历，进入 Parse 核对岗位解析结果，并选择仅保存规划或立即进入面试链路。

交付后的当前链路：

```text
Home JD source + ready Resume
  -> importTargetJob(resumeId)
  -> Parse loading / preview
  -> updateTargetJob
  -> workspace / workspace(autoStartPractice=1)
```

## 2 背景

`frontend-shell` 提供 App 壳、route normalization、auth pending action、runtime config、generated client bootstrap 与 fixture-backed transport。本 owner 只负责 `home` 与 `parse` 两个业务屏。

UI 必须源级追溯到 `ui-design/src/screen-home.jsx::HomeScreen`、`ui-design/src/screens-p0-complete.jsx::ParseScreen` 与 `ui-design/src/primitives.jsx`。正式前端只允许为真实数据、generated client、鉴权接续和可访问性做工程适配。

当前 API 合同来自 `openapi/openapi.yaml` 与 fixtures：

| operationId | Fixture | Frontend consumer | Backend owner | Persistence | AI dependency | Scenario |
|-------------|---------|-------------------|---------------|-------------|---------------|----------|
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json` | Home recent mock interviews | `backend-targetjob` | TargetJob read | none in frontend | `E2E.P0.014` |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` | Home resume select + Parse binding | `backend-resume` | Resume read | none | `E2E.P0.015` / `E2E.P0.016` |
| `createUploadPresign` | `openapi/fixtures/Uploads/createUploadPresign.json` | Home upload source | `backend-upload` | `file_objects` create | none | `E2E.P0.015` |
| `importTargetJob` | `openapi/fixtures/TargetJobs/importTargetJob.json` | Home paste / file / URL import with selected `resumeId` | `backend-targetjob` | TargetJob create + target job-level resume binding + parse job | backend-only | `E2E.P0.015` |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` | Parse polling + preview | `backend-targetjob` | TargetJob read | backend-only parse result | `E2E.P0.015` |
| `updateTargetJob` | `openapi/fixtures/TargetJobs/updateTargetJob.json` | Parse save / start | `backend-targetjob` | TargetJob partial update | none | `E2E.P0.016` |

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `contract`
- **TDD 策略**: 已通过 `/implement` -> `/tdd` 交付。当前 Red-Green-Refactor 证据保留在 focused Vitest：`frontend/src/app/screens/home/*`、`frontend/src/app/screens/parse/*`、`frontend/src/api/targetJob.realApiMode.test.ts`。
- **BDD 策略**: Feature plan requires BDD；当前 BDD gate 为 `E2E.P0.014`、`E2E.P0.015`、`E2E.P0.016`。
- **替代验证 gate**: 不适用；本计划具备 TDD + BDD 双层验证。

## 4 当前实现合同

### 4.1 Home

- 渲染 Hero label/title、`home-jd-input-card`、`home-jd-textarea`、输入卡底部 `home-jd-source-controls`、upload/URL source actions、`home-resume-row`、`home-resume-select`、`home-resume-create`、`home-submit-row` 与 `home-jd-submit`。
- `listResumes` 只把 ready 且可用的简历作为下拉选项；用户未显式选择简历时，paste / upload / URL import 均不得提交。
- paste 提交 `ImportTargetJobRequest.source.type=manual_text` + selected `resumeId`；upload 先 `createUploadPresign(purpose=target_job_attachment)`，再提交 `source.type=file` + selected `resumeId`；URL 提交 `source.type=url` + selected `resumeId`。
- `createUploadPresign`、`importTargetJob` 都必须通过 generated client 发送，并携带 side-effect idempotency key。
- 成功 import 后导航到 `parse`，params 必须包含 `targetJobId`、source 与真实 `resumeId`。
- `listTargetJobs` 只渲染最近 3 张模拟面试卡片，排序按 `updatedAt desc`；`更多` 进入 `workspace`。

### 4.2 Parse

- 进入 Parse 后先展示 4 步 loading gate，再根据 `getTargetJob.analysisStatus` 进入 preview 或 failed state。
- Preview 只渲染 API response 中的 title、companyName、locationText、requirements、summary、fitSummary、round assumptions 与 provenance 信息。
- Basic fields 的 title/company/location/notes 可编辑；requirements hit toggle 是前端临时确认状态，不写回后端。
- Parse 读取 ready 简历列表；若 route `resumeId` 命中 ready 简历则继承绑定，否则 Save/Start 保持 disabled，直到用户显式选择或创建简历。
- Save plan 与 Start interview 均先调用 `updateTargetJob`，request body 只包含 supplied editable fields，并携带 idempotency key。
- Save plan 进入 `workspace`；Start interview 进入 `workspace` 并携带 `autoStartPractice=1`，由 workspace owner 创建 session。

### 4.3 Privacy / Auth

- JD 原文、source URL 与 raw source content 不进入 URL、localStorage、console 或 telemetry。
- 未登录 Home import 使用 opaque pending import id 接续；pending action 不携带 JD 原文或 source URL。
- Parse Start 只有在真实 `resumeId` 已绑定时才触发 auth continuation。
- 前端只允许调用 generated TargetJobs / Uploads / Resumes client；不得直接调用 AI provider、prompt registry、provider-specific endpoint 或 ad hoc parse fetch。

## 5 实施步骤

### Phase 1: Home 当前入口

#### 1.1 UI source parity

Home DOM、布局、控件密度、主题、i18n 与响应式行为对齐 `ui-design/src/screen-home.jsx::HomeScreen`。

#### 1.2 Generated client contract

Home 使用 `listResumes`、`listTargetJobs`、`createUploadPresign` 和 `importTargetJob`。所有 request body、headers、route params 和错误态由 Vitest 覆盖。

#### 1.3 BDD-Gate

验证 `E2E.P0.014` 与 `E2E.P0.015`。

### Phase 2: Parse 当前确认与 handoff

#### 2.1 UI source parity

Parse loading、preview、failed state、resume binding、footer actions 与响应式行为对齐 `ui-design/src/screens-p0-complete.jsx::ParseScreen`。

#### 2.2 Generated client contract

Parse 使用 `getTargetJob`、`listResumes` 和 `updateTargetJob`。同 route target switch、failed state、partial update body、真实 `resumeId` handoff 与 auth continuation 均由 Vitest 覆盖。

#### 2.3 BDD-Gate

验证 `E2E.P0.015` 与 `E2E.P0.016`。

### Phase 3: 收口验证

#### 3.1 Focused frontend gates

运行 Home/Parse focused Vitest、frontend typecheck、fixture validation 与 real-mode generated-client gate。

#### 3.2 Scenario gates

串行运行 `E2E.P0.014`、`E2E.P0.015`、`E2E.P0.016` 的 `setup -> trigger -> verify -> cleanup`。

#### 3.3 Repo gates

运行 context validation、doc index check、docs-check、diff whitespace check 与 core-loop pruning surface lint。

### Phase 4: Import resume binding remediation

#### 4.1 Generated client request contract

Home must include the selected ready `resumeId` in every `importTargetJob` request body for paste, upload and URL sources. Missing resume remains a client-side block before request dispatch.

#### 4.2 Route continuity

Successful import still navigates to Parse with `targetJobId`, source and `resumeId`, but route params are no longer the only persistence layer for the binding; the backend TargetJob response is authoritative after reload or list re-entry.

#### 4.3 BDD-Gate

`E2E.P0.015` must continue to cover Home import request shape and privacy behavior, with `resumeId` treated as an allowed business identifier and JD raw text/source secrets still excluded from URL/pending action storage.

## 6 验收标准

- Home/Parse owner 文档只描述当前 Home + Parse 合同、operation matrix、BDD gate 和验证入口。
- `context.yaml` 只列当前正向 route、operationId、source package 与场景目录。
- `E2E.P0.014` / `E2E.P0.015` / `E2E.P0.016` 场景文档和脚本覆盖当前 Home/Parse 主路径、失败路径、privacy gate 与 real-mode generated-client gate。
- Home import request bodies include the selected `resumeId`, and backend list/detail can recover the binding without depending on transient Parse route params.
- `sync-doc-index --check`、`make docs-check`、`git diff --check` 和 `make lint-core-loop-pruning-surface` 通过。
