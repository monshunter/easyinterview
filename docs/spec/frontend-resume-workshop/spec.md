# Frontend Resume Workshop Spec

> **版本**: 1.7
> **状态**: active
> **更新日期**: 2026-07-07

## 1 背景与目标

`frontend-resume-workshop` 是当前 `resume_versions` 路由的前端 owner。正式前端必须源级复刻 [`ui-design/src/screen-resume-workshop.jsx`](../../../ui-design/src/screen-resume-workshop.jsx)、[`ui-design/src/primitives.jsx`](../../../ui-design/src/primitives.jsx)、[`ui-design/src/app.jsx`](../../../ui-design/src/app.jsx) 与 `docs/ui-design/` 中的当前 flat Resume Workshop 设计。

当前目标：

1. **路由接管**：`resume_versions` route 渲染 `ResumeWorkshopScreen`，支持 list / create / detail 三类视图。
2. **Flat Resume UI**：Resume 是平铺资产；详情页只有 preview / rewrites / edit 三 tab；所有前端数据投影都以 `resumeId` 识别简历。
3. **CreateFlow**：`flow=create` 进入 upload / paste 创建路径，解析完成后 preview confirm，保存回当前 flat Resume。
4. **Rewrites / Edit**：Rewrites tab 采纳建议后通过 `RewriteSaveConfirmModal` 选择覆盖或另存；Edit tab 使用 `updateResume` 保存手动编辑。
5. **真实 client 与 fixture fallback**：frontend 使用 generated client；real backend mode 与 fixture-backed dev path 都必须有测试护栏。
6. **UI parity 可执行**：用户可见变更必须有 DOM anchor、computed style、bounding box、viewport screenshot smoke 或对应 owner gate。

本 subject 不实现 backend handler、OpenAPI schema、migration、object storage、AI parsing 或真实 PDF 生成。

## 2 范围

### 2.1 In Scope

- **Route shell**：`ResumeWorkshopScreen` 解析 `flow=create|list`、`resumeId`、`tab=preview|rewrites|edit`，并与 app shell route / TopBar 状态一致。
- **List view**：`ResumeListView` 渲染平铺列表、统计、创建入口和详情入口。
- **Detail view**：`ResumeDetailView` 渲染 preview / rewrites / edit 三 tab。
- **Preview tab**：结构化预览、复制、原件弹层和 export fallback toast。
- **Create flow**：`ResumeCreateFlow`、`ResumeParseFlow`、`ResumePreviewConfirm`；upload / paste 两路径；`createUploadPresign`、browser PUT、`registerResume`、`getResume`、`updateResume` generated-client contract。
- **Rewrites / Edit**：`requestResumeTailor`、`getResumeTailorRun`、`updateResume`、`duplicateResume` generated-client contract。
- **i18n / a11y**：中英双语、tablist、modal focus trap、ESC、aria-live 和 keyboard behavior。
- **Auth boundary**：未登录只能显示登录引导 / pending action；pending action 只保存安全 route params。
- **Privacy boundary**：raw resume text、parsed summary、structured profile、rewrite text 不进入 console、URL、localStorage、telemetry 或 generic transport logs。
- **Parity gates**：每个 user-visible owner plan 保留 DOM / style / layout / screenshot smoke 或 scenario gate。

### 2.2 Out of Scope

- Backend Resume / Upload / Tailor handlers and stores.
- OpenAPI operation/schema design.
- Migration / schema changes.
- Object-storage provider implementation.
- Real PDF generation.
- Product areas not included in the active product-scope core loop.

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | UI 真理源 | `ui-design/src/screen-resume-workshop.jsx` + primitives + app shell + `docs/ui-design/` | 不从外部设计系统或 AI 审美生成正式前端视觉 |
| D-2 | Data adapter | UI 消费单一 `Resume` / `resumeId` view model；adapter 只做 display projection 和 fallback | 组件不直接拼 API response shape |
| D-3 | Route params | `flow=create|list`、`resumeId`、`tab=preview|rewrites|edit` | Route state 与 UI state 一一对应 |
| D-4 | Client mode | generated client 是唯一 API client；fixture-backed dev path 与 real backend mode 都保留测试 | 避免 mock-only drift |
| D-5 | UI parity | DOM anchor、computed style、bounding box、viewport screenshot smoke 为 user-visible gate | 不接受“风格接近”作为完成依据 |
| D-6 | Export fallback | P0 export button maps unavailable backend response to toast；copy text remains local | PDF 真生成不阻塞 P0 UI |
| D-7 | Negative gate | product-scope pruning gate owns non-current route/module/input regression scan | 防止范围外入口回流 |
| D-8 | Flat CreateFlow | CreateFlow 只提供 upload / paste；preview confirm 保存 flat resume | 与 OpenAPI / backend-resume flat contract 对齐 |
| D-9 | Rewrites save | Rewrites suggestions are task output; accepted suggestions save through overwrite or duplicate | 与 current ResumeTailor generated-client contract 对齐 |

### 3.2 待确认事项

- 当前没有阻塞本 subject 的产品或架构待确认项。

## 4 设计约束

### 4.1 UI 真理源约束

- 视觉、DOM、spacing、typography、color、shadow、radius、density、state 和 responsive behavior 必须追溯到 `ui-design/` 或 `docs/ui-design/`。
- 正式 frontend 不 import `ui-design/src/*` 作为 runtime component/data source。

### 4.2 数据约束

- Runtime data 只来自 generated client、runtime provider、fixture/mock client 或 user action。
- Adapter 位于 `frontend/src/app/screens/resume-workshop/adapters/` 或 create-flow 局部 adapter。
- Route and pending action must never carry raw resume content.

### 4.3 Privacy 约束

- Raw resume text、parsed summary、structured profile and rewrite text are user content.
- Errors and toast messages use enum/generic wording and must not echo raw payloads.
- Copy/export are explicit user actions; passive logs and route state are not allowed to carry resume content.

### 4.4 Verification 约束

- Component and hook behavior use Vitest.
- Route, auth, privacy and integration flows use focused scenario tests.
- Visual parity follows frontend-shell pixel parity owner patterns.
- Header / INDEX drift uses `/sync-doc-index`.

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `resume_versions` route | frontend-resume-workshop | Resume Workshop shell, list, create, detail |
| UI truth source | `ui-design/` + `docs/ui-design/` | Visual and interaction source |
| Generated client | openapi-v1-contract + frontend adapters | API surface and TS types |
| Upload backend | backend-upload | Presign and object file lifecycle |
| Resume backend | backend-resume | Register, parse, update, duplicate, tailor |
| App shell / auth pending action | frontend-shell | Route normalization and auth continuation |
| Workspace Resume Picker | frontend-workspace-and-practice | Workspace-level resume selection |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Route shell | Authenticated user opens `resume_versions` | Route renders | Resume Workshop shell appears and TopBar highlights resume nav | [001](./plans/001-listing-routing-and-detail-readonly/plan.md) |
| C-2 | List view | `listResumes` returns items | List loads | Flat table, create entrypoint and detail entrypoints render | [001](./plans/001-listing-routing-and-detail-readonly/plan.md) |
| C-3 | Detail preview | User opens a resume | Preview tab renders | Structured preview, copy, export fallback and original modal behave correctly | [001](./plans/001-listing-routing-and-detail-readonly/plan.md) |
| C-4 | Create upload | User selects valid file | Submit | Presign, PUT, register, parse, preview and update save path complete | [002](./plans/002-create-flow/plan.md) |
| C-5 | Create paste | User enters text | Submit | Register, parse, preview and update save path complete | [002](./plans/002-create-flow/plan.md) |
| C-6 | Create recovery | Parse fails, times out, or user cancels | User retries or returns | Input is preserved locally and no raw content leaks | [002](./plans/002-create-flow/plan.md) |
| C-7 | CTA handoff | Home or Workspace create CTA | Click | Route lands on CreateFlow and auth pending action is safe | [002](./plans/002-create-flow/plan.md) |
| C-8 | Rewrites | User opens rewrites tab | Accept and save | Accepted suggestions save through overwrite or duplicate path | [003](./plans/003-branch-rewrites-and-edit/plan.md) |
| C-9 | Edit | User opens edit tab | Save | Manual edit uses `updateResume` and refreshes detail state | [003](./plans/003-branch-rewrites-and-edit/plan.md) |
| C-10 | Privacy | User browses or edits resumes | App logs/routes/stores update | Raw resume content stays out of passive channels | 001 / 002 / 003 |
| C-11 | UI parity | Desktop and mobile viewports | Run owner gates | DOM/style/layout/screenshot smoke remain aligned with UI truth source | 001 / 002 / 003 |

## 7 关联计划

- [001-listing-routing-and-detail-readonly](./plans/001-listing-routing-and-detail-readonly/plan.md)：route shell、list view、detail preview、copy/export fallback and original modal owner.
- [002-create-flow](./plans/002-create-flow/plan.md)：current upload/paste CreateFlow、parse polling、preview update save, CTA handoff, privacy and focused frontend tests.
- [003-branch-rewrites-and-edit](./plans/003-branch-rewrites-and-edit/plan.md)：rewrites, edit, tailor polling, overwrite/duplicate save and detail consistency owner.
