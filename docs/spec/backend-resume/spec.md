# Backend Resume Spec

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-12

## 1 背景与目标

[engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 标记 Resume Workshop workstream 候选 subject 包含 `backend-resume`（C7 ownerDomain，对齐 [shared/jobs.yaml](../event-and-outbox-contract/spec.md#311-dbbackend-runner-canonical-job_type--asynq-dotted-task-name-映射)）。本 subject 是 Resume 业务域的后端 owner：承载 `resume_assets`（原始简历）/ `resume_versions`（结构化主版本 + 岗位定制版本）/ `resume_tailor_runs`（改写 run）三张表的 store 与 handler，组织 resume.parse / resume.tailor 异步 AI 编排，并把 Resume 业务事件（`resume.parse.completed` / `resume.tailor.completed`）发射到 [B3 events](../event-and-outbox-contract/spec.md)。

目标：

1. **完整业务域**：所有 Resume HTTP endpoint（B2 D-18 13 个 op：register / get / list / listVersions / getVersion / branch / update / accept / reject / requestTailor / getTailorRun / archive / export）都在本 subject 落地 handler；frontend 与 mock-contract-suite 消费同一份字节级响应。
2. **AI 编排封装**：resume.parse 与 resume.tailor 通过 [A3 AIClient + Capability Model Profile](../ai-provider-and-model-routing/spec.md) 调用；prompt / rubric / 模型版本通过 [F3](../prompt-rubric-registry/spec.md) 注册（3 个 baseline feature_key：`resume.parse` / `resume.tailor.gap_review` / `resume.tailor.bullet_suggestions`）；业务代码不 import 厂商 SDK。
3. **版本树语义**：`resume_versions.parent_version_id` 自引用支持版本链；branch 流程支持 3 种 `seed_strategy`（`copy_master` 同步返回 / `blank` 同步 / `ai_select` 入队 tailor job）。
4. **改写建议状态机**：`resume_version_suggestions.status` ∈ `pending → accepted | rejected`；accept 与 reject 都是终态（不可再切；如需重做必须新建 suggestion 或 branch 新 version）。
5. **mock-first 切真**：本 subject 落地的 13 个 endpoint 必须与 [B2 D-18 fixtures](../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) 字节级一致，frontend-resume-workshop 在切真时无需修订；frontend mock-first 路径不阻塞本 subject 进度。

本 subject 不实现 frontend UI（归 [frontend-resume-workshop](../frontend-resume-workshop/spec.md)）；不实现 file 上传（归 [backend-upload](../backend-upload/spec.md)）；不实现 mistakes / growth / drill 等已丢弃模块（[roadmap D-6](../engineering-roadmap/spec.md#31-已锁定决策) 禁止恢复）。

## 2 范围

### 2.1 In Scope

- **HTTP handler**：实现 [B2 §3.1.1](../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) Resumes + ResumeTailor tag 全部 13 个 operationId（含 D-18 9 个新 op）。
- **store layer**：`resume_assets` / `resume_versions` / `resume_version_suggestions` / `resume_tailor_runs` 4 张表 Repository（前 3 张由 [B4 002 resume-versions-additive](../db-migrations-baseline/plans/002-resume-versions-additive/plan.md) 落地 schema）。
- **AI 编排**：
  - `resume.parse` async job（[B3 jobs.yaml](../event-and-outbox-contract/spec.md#311-dbbackend-runner-canonical-job_type--asynq-dotted-task-name-映射) C7 owner）：解析 file_object → 提取 `structuredProfile` → 写 `resume_assets.parsed_summary` → 发射 `resume.parse.completed`。
  - `resume.tailor` async job：基于 targetJobId + 简历 master version 生成 suggestion → 写 `resume_tailor_runs` + `resume_version_suggestions` → 发射 `resume.tailor.completed`。
- **branch 业务逻辑**：`branchResumeVersion` 支持 3 个 seed_strategy；`ai_select` 同步返回 `resume_version_id` + 入队 `resume_tailor` job（202 + Job）。
- **suggestion 状态机**：accept → 写入 `decided_at` + 同步更新 `resume_versions.structured_profile`（可选）；reject → 仅写入 `decided_at`。
- **隐私链路**：privacy_delete 调用 backend-resume 提供的 `DeleteResumeAssetsForUser` API；级联 versions / suggestions / tailor_runs；调 [backend-upload `DeleteFileObjectsForUser`](../backend-upload/spec.md) 删除 file binary。
- **B3 events 发射**：`resume.parse.completed` / `resume.tailor.completed` 在 job 完成时通过 outbox 写入；payload 字段必须与 [B3 §3.1.4](../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory) 一致。
- **mock-first 对齐**：本 subject 实现的 handler 响应字段集 / status code / IK 行为与 B2 fixture 字节比对，[mock-contract-suite C-9](../mock-contract-suite/spec.md#6-验收标准) 强制 enforce。

### 2.2 Out of Scope

- 前端 Resume Workshop UI（[frontend-resume-workshop](../frontend-resume-workshop/spec.md)）。
- file 上传 / 对象存储（[backend-upload](../backend-upload/spec.md)）；本 subject 通过 `backend-upload.RegisterFileObject(fileObjectId, expectedPurpose=resume, ownerUserId)` 校验 file 后，把 `resume_assets.file_object_id` 写入自身业务表。
- 真实 PDF 导出（`exportResumeVersion` P0 行为：`501 + RESUME_EXPORT_NOT_AVAILABLE`，[B2 D-18](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围)）；P1 实现归未来 plan。
- 简历内容的进阶 AI 能力：JD 匹配评分自动化 / 主动改写推送 / 知识检索：P0 不实现。
- `resume_version_edits` 表（手动编辑历史，[B4 D-17](../db-migrations-baseline/spec.md#31-已锁定决策) 标 P1 延后）：本 subject 第一批 plan 不实现。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 术语映射 | 后端真理源：`ResumeAsset`（OpenAPI）/ `ResumeVersion`（新 schema）/ `ResumeTailorRun`；UI 文档使用 `ResumeSource` / `ResumeVersion` 与之对应；前端通过 adapter 层 wrap UI 命名 | 与 [B1 D-10](../shared-conventions-codified/spec.md#31-已锁定决策) + [B2 D-18](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) 对齐；不重命名 OpenAPI 已发布 schema |
| D-2 | seed_strategy AI 触发 | `copy_master`：同步返回 + structured_profile 拷贝；`blank`：同步返回 + structured_profile 空；`ai_select`：同步返回 resume_version_id + 入队 `resume_tailor` job（202 + Job） | 前端在 ai_select 路径必须轮询 `getResumeTailorRun` 或监听对应 event；branch 必带 IK |
| D-3 | suggestion 状态终态 | `pending → accepted | rejected`，accept 与 reject 都是终态；改写历史不维护（如需变更必须 branch 新 version 或新 suggestion） | UI 与 store 都不实现 "撤销 accept" / "撤销 reject"；ConfirmDialog 由前端在用户行为前提示 |
| D-4 | parse 路径区分 | `RegisterResumeRequest.sourceType` ∈ `upload | paste | guided`：upload 必带 `fileObjectId`；paste 必带 `rawText`；guided 必带 `guidedAnswers`；其他组合返回 422。guided 答案原样持久到 `resume_assets.guided_answers` jsonb，不序列化进 `original_text` | [B2 D-18](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) schema 已约束；本 spec 进一步在 handler 层 enforce；字段由 [B4 D-17](../db-migrations-baseline/spec.md#31-已锁定决策) 提供 |
| D-5 | tailor 模式 | `RequestResumeTailorRequest.mode` ∈ `gap_review | bullet_suggestions`（与 [B3 D-14](../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表) 对齐）；不复活旧 `inline | rewrite | mirror` | events / API / DB 三层 mode enum 同源 |
| D-6 | RESUME_EXPORT_NOT_AVAILABLE 行为 | `exportResumeVersion` P0 默认返回 `501` + `error.code = "RESUME_EXPORT_NOT_AVAILABLE"`；P1 切到 `202 + Job(jobType=resume_export)` 属 additive 行为变化 | 类比 [B2 D-12 privacy export 例外](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围)；frontend toast 兜底 + copyText 真实可用 |
| D-7 | listResumes / listResumeVersions pagination | 默认 pageSize=20，cursor 分页；返回 `PaginatedResumeAsset` / `PaginatedResumeVersion`（B2 D-18 schema） | 与 [B2 D-5](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) 分页规则一致 |
| D-8 | branch / update / accept / reject / archive / export 必带 IK | 6 个 side-effect operation 必带 `Idempotency-Key`（与 [B2 D-18](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) 一致） | 防止网络抖动产生重复 version / 重复 accept |
| D-9 | 首次创建保存边界 | `registerResume` 只登记 `ResumeAsset` source 并触发 `resume.parse`；parse job 只产出解析草稿（`parsed_summary` / `parsed_text_snapshot`）和 parse 状态，不在用户 Preview Confirm 前创建正式 `structured_master` `ResumeVersion`。确认保存 v1 由后续 backend-resume/002 + frontend-resume-workshop/002 承接。 | 对齐 `docs/ui-design/resume-onboarding.md` 的 `输入 -> Agent 解析 -> 预览确认 -> 保存 v1`；防止未确认草稿成为正式简历版本 |

### 3.2 待确认事项

- accept suggestion 时是否自动更新 `resume_versions.structured_profile`：默认 P0 不更新 structured_profile（仅写 suggestion.status = accepted）；P1 评估，由 frontend 与 backend 联合决定字段同步语义。
- branchResumeVersion seedStrategy=ai_select 入队的 tailor job 是否必须立即同步返回 suggestion 列表：默认 P0 异步（202 + Job + frontend 轮询）；如 frontend UX 需要同步响应，由本 spec 修订评估。
- `RESUME_*` 错误码扩展（如 `RESUME_VERSION_BRANCH_FROM_INVALID_PARENT`）：默认走 [B1 D-5](../shared-conventions-codified/spec.md#31-已锁定决策) 通用 `VALIDATION_FAILED`；如业务需要更细分类再修订 B1。

## 4 设计约束

### 4.1 契约约束

- 实现 [B2 §3.1.1](../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) 全部 Resume operation 的 generated server interface；不允许私造 handler 签名。
- 响应字段集 / status code / IK 行为与 [B2 fixtures](../mock-contract-suite/spec.md) 字节比对；新增 scenario 必须 B2 plan 修订同步。
- 错误码必须 `$ref` [B1 D-5](../shared-conventions-codified/spec.md#31-已锁定决策) 已锁定的常量集 + [B1 D-10 RESUME_EXPORT_NOT_AVAILABLE](../shared-conventions-codified/spec.md#31-已锁定决策)；不私造未登记错误码。
- 异步 job 必须通过 [B3 jobs.yaml](../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表) 已登记的 `resume_parse` / `resume_tailor` canonical job_type；不私造 dotted task name。

### 4.2 AI 约束

- resume.parse / resume.tailor 必须通过 [A3 AIClient](../ai-provider-and-model-routing/spec.md) 调用；不允许业务代码 import 厂商 SDK / 直接 HTTP 调 model endpoint。
- prompt / rubric / 模型版本必须通过 [F3 registered feature_key](../prompt-rubric-registry/spec.md) 引用：`resume.parse`（model profile `resume.parse.default`）/ `resume.tailor.gap_review` / `resume.tailor.bullet_suggestions`；本 subject 不 hardcode prompt 正文。
- AI 输出必须含 `GenerationProvenance`（[B2 §4.6](../openapi-v1-contract/spec.md#46-ai-生成结果-provenance-约束)）；写入 `ai_task_runs.model_profile_*` typed columns + `resume_versions.{prompt_version, rubric_version, model_id, provider}` columns。
- AI capability 仅消费 `chat`（[B1 D-8](../shared-conventions-codified/spec.md#31-已锁定决策)）；不引入 stt / realtime / judge / 向量检索。

### 4.3 存储约束

- `resume_versions.user_id` / `resume_version_suggestions` 通过 FK 与 `resume_versions` 关联（[B4 002 资源版本 migration](../db-migrations-baseline/plans/002-resume-versions-additive/plan.md)）；不绕过 store 层直接 SQL。
- 跨用户隔离：所有 read endpoint 必须以 `user_id = current_user_id` 过滤；cross-user 访问返回 404（不暴露存在）。
- 隐私删除调用 `DeleteResumeAssetsForUser(userId)`：先 `resume_version_suggestions` → `resume_versions` → `resume_tailor_runs` → `resume_assets`；file binary 与 `file_objects` 删除由 backend-upload `DeleteFileObjectsForUser` 在同一 privacy request 中按 B4 matrix 协调（对象存储删除成功后再 hard delete DB 行）。
- raw resume text 与 guided answers（`resume_assets.original_text` / `resume_assets.guided_answers` / `raw_text` / `parsed_text_snapshot`）不出现在 audit_events / outbox / log 中（[B3 §3.1.4 PII 边界](../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory)）。

### 4.4 BDD / TDD 约束

- 每个 endpoint 必须有 handler unit test（参数校验 + IK + 错误路径）+ store integration test（state transition + cross-user isolation）+ AI 调用 unit test（stub provider，验证 prompt/profile 路由正确）。
- 用户可见行为（register / list / branch / accept / reject / parse 完成）必须有 BDD scenario 覆盖。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| 13 个 Resume HTTP handler | backend-resume | 真实业务逻辑 |
| `resume_assets` / `resume_versions` / `resume_version_suggestions` / `resume_tailor_runs` 表 schema | [B4 db-migrations-baseline](../db-migrations-baseline/spec.md) + [B4 002 plan](../db-migrations-baseline/plans/002-resume-versions-additive/plan.md) | 字段 / 索引 / FK / check constraint |
| file_object 引用 | [backend-upload](../backend-upload/spec.md) `Register` internal API | resume_assets 通过 backend-upload 引用 file_object |
| `resume.parse` / `resume.tailor` async job | backend-resume + backend-runtime-topology | job handler 注册到 backend internal runner |
| AI 调用 | [A3 AIClient](../ai-provider-and-model-routing/spec.md) + [F3 feature_key](../prompt-rubric-registry/spec.md) | backend-resume 只引用 profile，不绑定 provider |
| 隐私删除调用 | backend internal privacy runner（[backend-runtime-topology](../backend-runtime-topology/spec.md)） | 调用 `DeleteResumeAssetsForUser` |
| frontend Resume Workshop UI | [frontend-resume-workshop](../frontend-resume-workshop/spec.md) | 消费 generated TS client |
| mock-first fixtures | [B2 fixtures](../openapi-v1-contract/spec.md) + [openapi-v1-contract/004](../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) | backend-resume handler 响应字节比对 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | registerResume (upload) 主路径 | 已登录 + 有效 file_object (purpose=resume) + IK | 调 `POST /api/v1/resumes` `{sourceType: upload, fileObjectId, title, language}` | 返回 202 + `ResumeAssetWithJob{resumeAssetId, job(jobType=resume_parse, status=queued)}`；DB `resume_assets` 行 `parse_status='queued'`；触发 `resume.parse` async job | 001-asset-register-parse-and-listing |
| C-2 | registerResume (paste / guided) | sourceType=paste with rawText OR guided with guidedAnswers | 调 register | 同 C-1 行为，但 `resume_assets.file_object_id` 为 NULL；paste 写 `source_type='paste'` + `original_text`；guided 写 `source_type='guided'` + `guided_answers` jsonb；`parsed_text_snapshot` 仅由 parse job 后续写入 | 001 |
| C-3 | resume.parse async 完成 | resume.parse job consumer 处理 queued 行 | 通过 [A3 AIClient](../ai-provider-and-model-routing/spec.md) 调 model + parse JSON | DB `resume_assets.parse_status='ready'` + `parsed_summary` / `parsed_text_snapshot` 写入；触发 outbox `resume.parse.completed`（envelope 字段集 [B3 §3.1.4](../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory) 一致）；ai_task_runs 行写入 typed columns；用户 Preview Confirm 前不得创建正式 `structured_master` `ResumeVersion` | 001 |
| C-4 | resume.parse 失败 retryable | AI provider 返回 timeout / output_invalid | resume.parse 失败 | DB `resume_assets.parse_status='failed'` + `error_code='AI_PROVIDER_TIMEOUT'`；retryable 重试上限到达后停止；privacy 红线：error 不含 prompt / response 摘要 | 001 |
| C-5 | listResumes pagination | 用户 A 有 25 个 resume_asset | 调 `GET /api/v1/resumes?pageSize=20` 然后 cursor | 第一页返回 20 行 + `pageInfo.nextCursor`；第二页返回 5 行 + `hasMore=false`；按 `updated_at DESC` 排序；cross-user 不可见 | 001 |
| C-6 | cross-user 隔离 | 用户 A 有 resume；用户 B 调 `getResume(A.resumeAssetId)` | – | 404；不暴露存在；audit_events 不写入敏感字段 | 001 + 后续 plan |
| C-7 | IK replay | register 同 IK 重复调用 | – | 返回首次 `resumeAssetId`；不创建新 DB 行 | 001 |
| C-8 | mock-first 字节比对 | B2 fixture `registerResume.json` `default` scenario | 调真实 handler | 响应字段集 / status / header 字节一致 | 001 + mock-contract-suite |
| C-9 | privacy 删除链路 | 用户 A 有 3 resume_asset + 5 version + 10 suggestion + 2 tailor_run | privacy_delete job 触发 | backend-resume 删除顺序：suggestions → versions → tailor_runs → assets；backend-upload 同一 privacy request 删除 file binary / file_objects（对象存储先删，成功后 DB hard delete）；audit tombstone 仅保留 ID / 删除时间，不含内容 | 后续 plan |
| C-10 | branchResumeVersion seedStrategy 三路 | 用户 A 有 master version | 分别调 `copy_master` / `blank` / `ai_select` branch | `copy_master` 同步返回 + structured_profile 拷贝；`blank` 同步返回 + structured_profile 空；`ai_select` 同步返回 resume_version_id + 入队 resume.tailor job | 后续 plan |
| C-11 | suggestion accept/reject 状态机 | suggestion `pending` | 调 accept；再调 accept | 首次返回 200 + `decided_at` 写入；第二次返回 409 + `error.code = "VALIDATION_FAILED"`（或按 IK 语义幂等返回首次结果）；不私造未登记状态迁移错误码 | 后续 plan |
| C-12 | exportResumeVersion P0 | 调 `POST /api/v1/resume-versions/{id}/exports` | – | 返回 501 + `error.code="RESUME_EXPORT_NOT_AVAILABLE"`；ai_task_runs 不写入；不消耗 model 配额 | 后续 plan |
| C-13 | events 漂移负向 | grep `inline\|rewrite\|mirror` 在 events / job / dispatcher 上下文 | – | 0 命中（与 [B3 D-14](../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表) 同步） | 001 |

## 7 关联计划

- [001-asset-register-parse-and-listing](./plans/001-asset-register-parse-and-listing/plan.md)：第一批 plan，落地 `registerResume` + `getResume` + `listResumes` + `resume.parse` async job + sourceType 三路 + `resume.parse.completed` event；BDD 覆盖 register → parse → list 主路径。
- `002-versions-and-tailor-runs`（未创建，由 001 完成后启动）：落地 Preview Confirm 保存 v1 `structured_master`、`listResumeVersions` / `branchResumeVersion` / `updateResumeVersion` / `requestResumeTailor` / `getResumeTailorRun` / `acceptSuggestion` / `rejectSuggestion` + `resume.tailor.completed` event。
- `003-export-and-archive-and-delete`（P1 延后）：落地 `exportResumeVersion` 真实 PDF 生成 + `archiveResumeAsset` + privacy delete 链路 fully integrate。
