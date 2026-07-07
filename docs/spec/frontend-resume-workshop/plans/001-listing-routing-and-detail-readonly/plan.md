# Frontend Resume Workshop Listing Routing and Detail Readonly

> **版本**: 1.9
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本计划承接当前 `frontend-resume-workshop` 的首屏与只读详情边界：

- `resume_versions` route 渲染 `ResumeWorkshopScreen`，TopBar 选中简历入口。
- route params 只使用当前 flat Resume 合同：`flow=create|list`、`resumeId`、`createMode=upload|paste`；旧 `tab` / `tailorRunId` 被过滤或忽略。
- `ResumeListView` 使用 `listResumes` 渲染单层平铺表格、创建入口、详情入口、loading / empty / retry / pagination 状态。
- `ResumeDetailView` 使用 `getResume(resumeId)` 只渲染原始简历内容本身和 generic 404 fallback。
- 列表与详情不展示通用“上传的简历 / 粘贴的简历 / Uploaded resume / Pasted resume”名称；完成态名称优先使用 backend generated `displayName` 或 LLM structured headline；前端不得把 raw resume 第一行、上传文件名或与来源 `title` 相同的文件名 `displayName` 当作名称。
- 未登录态不触发 Resume API，请求登录时 pending action 只保存安全 route params。
- 可见 UI 继续追溯 `ui-design/src/screen-resume-workshop.jsx`、`ui-design/src/primitives.jsx`、`ui-design/src/app.jsx` 和 `docs/ui-design/`。

本计划不拥有 CreateFlow 注册链路、tailor polling、duplicate/save-as-new 或 backend handler 行为；本计划同时固化详情页不提供 Rewrites/Edit/export/copy/original modal/preview-confirm 等二次操作。

## 2 背景

当前产品已经收敛为 flat Resume Workshop。001 作为首个前端 owner，只保留当前仍被运行时、场景和 UI 真理源共同承接的 list / preview detail 合同。旧树形列表、版本集合、分叉参数、逐版本导出和占位页接管说明不再作为计划语义存在。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior + frontend + contract-consumer`
- **TDD 策略**: 适用。实现项由 `/implement frontend-resume-workshop/001-listing-routing-and-detail-readonly` 进入 `/tdd`；测试断言来源为 `ResumeWorkshopScreen`、`ResumeWorkshopAuthGate`、`ResumeListView`、`ResumeDetailView`、`ResumeDetailFixtureParity`、`ResumeDetailExport`、`ResumePreviewTab`、`ResumeWorkshopI18nA11y`、`ResumeWorkshopPrivacy`、`fixture-parity` 和 P0.036/P0.037 scenario Vitest。
- **BDD 策略**: 适用。主 checklist 保留 E2E.P0.036 / E2E.P0.037 `BDD-Gate:`，场景细节由 [bdd-plan.md](./bdd-plan.md) 与 [bdd-checklist.md](./bdd-checklist.md) 承接。
- **替代验证 gate**: focused frontend Vitest、P0.036/P0.037 scenario scripts、frontend typecheck/build 或 owner parity gate、context validation、`sync-doc-index --check`、`make docs-check`、`git diff --check`、core-loop pruning surface lint。

### 3.1 Frontend / Backend Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` `default` / `empty` / `paginated` | list hook + `ResumeListView` + `mapResumeToUiSource` display-name fallback | backend-resume real handler | `resumes` | none | E2E.P0.036 |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` `default` / `not-found` | detail hook + `ResumeDetailView` + `ResumePreviewTab` original-content body | backend-resume real handler | `resumes` | none | E2E.P0.037 |

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

创建入口导航到 `resume_versions?flow=create`；打开行导航到 `resume_versions?resumeId=<id>`。

### Phase 3: Read-only Detail

#### 3.1 Read-only detail

`ResumeDetailView` 使用 `getResume(resumeId)` 渲染 crumb、header meta 和只读原始简历正文；显式 `tab=preview|rewrites|edit` 不 materialize 任何 tab 或二次编辑 surface。

#### 3.2 Removed actions

详情页不渲染 Export PDF、Copy text、View original/original modal、Rewrites、Edit 或 preview-confirm；原始简历预览就是当前原文正文。

#### 3.4 Original-content projection and meaningful names

`ResumePreviewTab` 优先展示 `parsedTextSnapshot`，其次 `originalText`，最后才降级到结构化字段的只读摘要；上传文件刚注册后若 `parseStatus` 仍为 `queued/processing` 且正文快照为空，详情页可以轻量轮询 `getResume`，但不得恢复 parser animation 或 preview-confirm；若 `parseStatus='failed'` 或任一可读正文已到达，详情必须停止轮询并展示该原文快照。列表和详情 header 对通用占位 `displayName` 做负向过滤，使用 backend generated 名称或 structured headline；raw resume 第一行、上传文件名或与来源 `title` 相同的文件名 `displayName` 不得作为名称兜底。

#### 3.3 404

不存在的 `resumeId` 渲染 generic NotFoundEmptyState，不回显 fixture `error.code`。

### Phase 4: Privacy / I18n / A11y / Parity

#### 4.1 Privacy

raw resume text、parsedTextSnapshot、parsedSummary、structuredProfile 和 rewrite text 不进入 URL、pending action、localStorage、console、telemetry 或 generic mock transport logs。

#### 4.2 I18n and accessibility

中英 key 由 frontend-shell i18n 体系承接；table、只读正文、buttons 和 aria labels 具备可测试语义。

#### 4.3 UI parity

DOM anchor、computed style、bounding box、mobile / desktop layout 和 screenshot smoke 追溯 UI 真理源；截图 diff 只在 baseline 稳定时作为补充 gate。

### Phase 5: BDD / Negative Gate / Closeout

#### 5.1 BDD scenarios

E2E.P0.036 验证 flat list + auth boundary；E2E.P0.037 验证 read-only detail + legacy tab negative + removed actions + 404 fallback。

#### 5.2 Non-current negative gate

Resume Workshop runtime source、scenario evidence 和 rendered DOM 不出现树形列表、版本 route params、版本集合 operation、分叉参数、prototype runtime import 或 non-current route testid。

#### 5.3 Docs and index

计划、checklist、BDD、context、spec history、scenario INDEX 和 docs/spec INDEX 同步到当前 Header。

## 5 验收标准

- 001 owner docs 只描述当前 flat Resume list / original-content read-only detail 合同。
- Operation matrix 只列当前详情实际消费的 generated-client operations。
- E2E.P0.036 / E2E.P0.037 scenario assets 指向当前 slug、当前 Vitest entry 和当前 expected outcome。
- Focused frontend tests、context validation、docs/index gates 和 pruning surface lint 通过。

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| Flat list 文档再次回流树形语义 | P0.036、fixture parity 和 pruning surface lint 保留负向断言 |
| Fixture-backed UI 被误认为 real backend 闭环 | scenario trigger 保留 real-mode/generated-client gate，operation matrix 标明真实 handler / fixture 边界 |
| 旧详情动作回流 | P0.037、pixel parity 和 negative grep 固化 Export/Copy/Original/Rewrites/Edit absence |

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-07 | 1.9 | 修订上传详情性能回归：`failed` 或已有可读正文时停止 `getResume` 轮询；名称消费改为 backend generated displayName 优先。 |
| 2026-07-07 | 1.8 | 修订未闭环回归：禁止上传文件名 / 与来源 title 相同的文件名 displayName 作为可见名称；failed resume 只要已有 parsedTextSnapshot 仍展示原文。 |
| 2026-07-07 | 1.6 | 修订未闭环回归：详情正文改为优先展示原始内容快照，列表/详情过滤通用上传/粘贴名称并增加内容派生兜底。 |
| 2026-07-07 | 1.7 | 修订未闭环回归：禁止 raw 第一行作为可见名称；上传详情在原文快照到达前轻量轮询，避免 PDF 详情空白。 |
| 2026-07-07 | 1.5 | 将详情页收敛为只读简历正文，移除 export/copy/original modal/Rewrites/Edit 正向 gate，并过滤旧 `tab` / `tailorRunId` route 口径。 |
| 2026-07-07 | 1.4 | 压缩 001 owner 到当前 flat Resume list/detail preview 合同，移除旧树形/版本集合/分叉参数语义，并同步 P0.036 当前场景 slug。 |
