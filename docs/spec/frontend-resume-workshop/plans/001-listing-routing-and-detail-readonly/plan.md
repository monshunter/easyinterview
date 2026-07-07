# Frontend Resume Workshop Listing Routing and Detail Readonly

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本计划承接当前 `frontend-resume-workshop` 的首屏与只读详情边界：

- `resume_versions` route 渲染 `ResumeWorkshopScreen`，TopBar 选中简历入口。
- route params 只使用当前 flat Resume 合同：`flow=create|list`、`resumeId`、`tab=preview|rewrites|edit`、`targetJobId`、`tailorRunId`、`createMode=upload|paste`。
- `ResumeListView` 使用 `listResumes` 渲染单层平铺表格、创建入口、详情入口、loading / empty / retry / pagination 状态。
- `ResumeDetailView` 使用 `getResume(resumeId)` 渲染 preview tab、copy、original modal、export fallback 和 generic 404 fallback。
- 未登录态不触发 Resume API，请求登录时 pending action 只保存安全 route params。
- 可见 UI 继续追溯 `ui-design/src/screen-resume-workshop.jsx`、`ui-design/src/primitives.jsx`、`ui-design/src/app.jsx` 和 `docs/ui-design/`。

本计划不拥有 CreateFlow、Rewrites save、Edit save、tailor polling 或 duplicate/save-as-new 行为；这些由 002 / 003 owner 承接。

## 2 背景

当前产品已经收敛为 flat Resume Workshop。001 作为首个前端 owner，只保留当前仍被运行时、场景和 UI 真理源共同承接的 list / preview detail 合同。旧树形列表、版本集合、分叉参数、逐版本导出和占位页接管说明不再作为计划语义存在。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior + frontend + contract-consumer`
- **TDD 策略**: 适用。实现项由 `/implement frontend-resume-workshop/001-listing-routing-and-detail-readonly` 进入 `/tdd`；测试断言来源为 `ResumeWorkshopScreen`、`ResumeWorkshopAuthGate`、`ResumeListView`、`ResumeDetailView`、`ResumeDetailFixtureParity`、`ResumeDetailExport`、`OriginalResumePreviewModal`、`ResumePreviewTab`、`ResumeWorkshopI18nA11y`、`ResumeWorkshopPrivacy`、`fixture-parity` 和 P0.036/P0.037 scenario Vitest。
- **BDD 策略**: 适用。主 checklist 保留 E2E.P0.036 / E2E.P0.037 `BDD-Gate:`，场景细节由 [bdd-plan.md](./bdd-plan.md) 与 [bdd-checklist.md](./bdd-checklist.md) 承接。
- **替代验证 gate**: focused frontend Vitest、P0.036/P0.037 scenario scripts、frontend typecheck/build 或 owner parity gate、context validation、`sync-doc-index --check`、`make docs-check`、`git diff --check`、core-loop pruning surface lint。

### 3.1 Frontend / Backend Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` `default` / `empty` / `paginated` | list hook + `ResumeListView` + `mapResumeToUiSource` | backend-resume real handler | `resumes` | none | E2E.P0.036 |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` `default` / `not-found` | detail hook + `ResumeDetailView` + `ResumePreviewTab` + original modal | backend-resume real handler | `resumes` | none | E2E.P0.037 |
| `exportResume` | `openapi/fixtures/Resumes/exportResume.json` `p0-501-not-available` | `ResumeDetailView` header / preview export action; generated client passes `Idempotency-Key` | backend-resume P0 unavailable response | none in P0 | none | E2E.P0.037 |

## 4 实施步骤

### Phase 1: Route Shell / Auth Boundary

#### 1.1 Route dispatch

`ResumeWorkshopScreen` 解析当前 route params，并在 `flow=create`、`resumeId`、list 三种状态间分派到 CreateFlow、Detail、List。

#### 1.2 Auth gate

runtime 未认证时渲染登录入口；Resume API 请求保持 0 次；pending action 只携带 route params，不携带 raw resume content、parsed summary、structured profile 或 rewrite text。

### Phase 2: Flat List View

#### 2.1 Flat table

`ResumeListView` 从 `listResumes` 读取 flat Resume items，按最近更新时间排序，渲染名称、来源、语言、最近编辑和打开按钮。

#### 2.2 List states

覆盖 loading、empty、retryable error 和 `pageInfo.hasMore` 提示；数量和行 ID 从 fixture / API response 派生。

#### 2.3 Create and detail entry

创建入口导航到 `resume_versions?flow=create`；打开行导航到 `resume_versions?resumeId=<id>&tab=preview`。

### Phase 3: Detail Preview / Original / Export

#### 3.1 Preview detail

`ResumeDetailView` 使用 `getResume(resumeId)` 渲染 crumb、header meta、preview / rewrites / edit tablist；默认 tab 为 `preview`，显式 `tab=rewrites|edit` 不被改写。

#### 3.2 Preview actions

Preview tab 支持 copy text、view original modal 和 Export PDF。original modal 展示原始文本或解析文本快照，具备 `role=dialog`、`aria-modal`、focus return 和 ESC 关闭。

#### 3.3 Export fallback and 404

Export PDF 通过 generated client 调用 `exportResume(resumeId, { idempotencyKey })`，501 / unavailable response 映射为本地 toast，不写 blob / localStorage。不存在的 `resumeId` 渲染 generic NotFoundEmptyState，不回显 fixture `error.code`。

### Phase 4: Privacy / I18n / A11y / Parity

#### 4.1 Privacy

raw resume text、parsedTextSnapshot、parsedSummary、structuredProfile 和 rewrite text 不进入 URL、pending action、localStorage、console、telemetry 或 generic mock transport logs。

#### 4.2 I18n and accessibility

中英 key 由 frontend-shell i18n 体系承接；table、tablist、modal、buttons 和 aria labels 具备可测试语义。

#### 4.3 UI parity

DOM anchor、computed style、bounding box、mobile / desktop layout 和 screenshot smoke 追溯 UI 真理源；截图 diff 只在 baseline 稳定时作为补充 gate。

### Phase 5: BDD / Negative Gate / Closeout

#### 5.1 BDD scenarios

E2E.P0.036 验证 flat list + auth boundary；E2E.P0.037 验证 detail preview + original modal + export 501 + 404 fallback。

#### 5.2 Non-current negative gate

Resume Workshop runtime source、scenario evidence 和 rendered DOM 不出现树形列表、版本 route params、版本集合 operation、分叉参数、prototype runtime import 或 non-current route testid。

#### 5.3 Docs and index

计划、checklist、BDD、context、spec history、scenario INDEX 和 docs/spec INDEX 同步到当前 Header。

## 5 验收标准

- 001 owner docs 只描述当前 flat Resume list / preview detail 合同。
- Operation matrix 只列当前 generated-client operations。
- E2E.P0.036 / E2E.P0.037 scenario assets 指向当前 slug、当前 Vitest entry 和当前 expected outcome。
- Focused frontend tests、context validation、docs/index gates 和 pruning surface lint 通过。

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| Flat list 文档再次回流树形语义 | P0.036、fixture parity 和 pruning surface lint 保留负向断言 |
| Fixture-backed UI 被误认为 real backend 闭环 | scenario trigger 保留 real-mode/generated-client gate，operation matrix 标明真实 handler / fixture 边界 |
| Detail tab owner 交叉 | 001 只拥有 preview/detail shell；Rewrites/Edit save 行为由 003 owner docs 和 tests 承接 |

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-07 | 1.4 | 压缩 001 owner 到当前 flat Resume list/detail preview 合同，移除旧树形/版本集合/分叉参数语义，并同步 P0.036 当前场景 slug。 |
