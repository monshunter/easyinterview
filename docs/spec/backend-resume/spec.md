# Backend Resume Spec

> **版本**: 1.7
> **状态**: active
> **更新日期**: 2026-07-07

## 1 背景与目标

[engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 标记 Resume Workshop workstream 候选 subject 包含 `backend-resume`（C7 ownerDomain，对齐 [shared/jobs.yaml](../event-and-outbox-contract/spec.md#311-dbbackend-runner-canonical-job_type--asynq-dotted-task-name-映射)）。本 subject 是 Resume 业务域的后端 owner：承载 `resumes` store 与 handler，组织 resume.parse / resume.tailor 异步 AI 编排，并把 Resume 业务事件（`resume.parse.completed` / `resume.tailor.completed`）发射到 [B3 events](../event-and-outbox-contract/spec.md)。

> **D-20 简历扁平化（product-scope v2.1）**：简历是平铺列表中的独立扁平资产，不区分原始 / 结构化主版本 / 岗位定制版本，不做版本树 / 分叉 / 继承。`resumes` 是当前唯一简历业务表，结构化内容（`structured_profile` / `display_name`）直接归属该表；契约 op 由 [B2 D-26](../openapi-v1-contract/spec.md) 锁定为 9 个，外部标识统一使用 `resumeId`。

目标：

1. **完整业务域**：所有 Resume HTTP endpoint（B2 D-26 后共 9 个 op：`registerResume` / `getResume` / `listResumes` / `updateResume` / `duplicateResume` / `archiveResume` / `exportResume` / `requestResumeTailor` / `getResumeTailorRun`）都在本 subject 落地 handler，并通过 `cmd/api` session middleware / idempotency middleware 挂到真实 `/api/v1/*` route；frontend 与 mock-contract-suite 消费同一份字节级响应。
2. **AI 编排封装**：resume.parse 与 resume.tailor 通过 [A3 AIClient + Capability Model Profile](../ai-provider-and-model-routing/spec.md) 调用；prompt / rubric / 模型版本通过 [F3](../prompt-rubric-registry/spec.md) 注册（3 个 baseline feature_key：`resume.parse` / `resume.tailor.gap_review` / `resume.tailor.bullet_suggestions`）；业务代码不 import 厂商 SDK。
3. **扁平资产语义（D-20）**：每份 resume 是独立扁平资产，承载只读原始来源快照（`original_text` / `parsed_text_snapshot` / `raw_text` / `file_object_id`）+ 可编辑结构化内容（`structured_profile` / `display_name`）；不引入版本树、分叉、主版本或岗位定制版本。`registerResume` + `resume.parse` 直接产出该 resume 的 `structured_profile`。
4. **改写建议 ephemeral（D-20）**：`resume.tailor` 生成的 bullet 改写建议经 `getResumeTailorRun` 返回，run + suggestions 持久化在 `ai_task_runs` 的 task 输出中；用户客户端采纳后经 `updateResume`（覆盖原简历）或 `duplicateResume`（保存为新简历）落盘；服务端不持久化逐条 `accepted | rejected` 状态。
5. **mock-first 切真**：本 subject 落地的 9 个 endpoint 必须与 [B2 fixtures](../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) 及本 plan 新增 fixture 字节级一致，frontend-resume-workshop 在切真时无需修订；frontend mock-first 路径不阻塞本 subject 进度。

本 subject 不实现 frontend UI（归 [frontend-resume-workshop](../frontend-resume-workshop/spec.md)）；不实现 file 上传（归 [backend-upload](../backend-upload/spec.md)）；不实现 mistakes / growth / drill 等当前 P0 外模块（[roadmap D-6](../engineering-roadmap/spec.md#31-已锁定决策) 禁止恢复）。

## 2 范围

### 2.1 In Scope

- **HTTP handler + runtime wiring**：实现 [B2 §3.1.1](../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) Resumes + ResumeTailor tag 全部 9 个 operationId（D-20 扁平化后：`registerResume` / `getResume` / `listResumes` / `updateResume` / `duplicateResume` / `archiveResume` / `exportResume` / `requestResumeTailor` / `getResumeTailorRun`），并在 `cmd/api` 中按当前 session / IK / generated response envelope 口径挂载真实 route。
- **store layer**：`resumes` 单一扁平表 Repository。`resumes` 字段：`id` / `user_id` / `file_object_id` / `title` / `display_name` / `language` / `source_type`∈{`upload`,`paste`} / `parse_status` / `parsed_summary` / `raw_text` / `original_text` / `parsed_text_snapshot` / `structured_profile` / timestamps / `deleted_at`。
- **AI 编排**：
  - `resume.parse` async job（[B3 jobs.yaml](../event-and-outbox-contract/spec.md#311-dbbackend-runner-canonical-job_type--asynq-dotted-task-name-映射) C7 owner）：解析 file_object / paste text → 提取结构化内容 → 写 `resumes.parsed_summary` / `parsed_text_snapshot` / `structured_profile`；最终 `parse_status='ready'` 时发射 `resume.parse.completed`（`resumeId`），失败路径只写 `ai_task_runs` / audit / async retry metadata，不发 completed event。`registerResume` + parse 直接产出 resume 的结构化内容，无独立主版本确认步骤。
  - `resume.tailor` async job：基于 `resumeId`（可选 `targetJobId` JD-aware 上下文）+ `mode`∈{`gap_review`,`bullet_suggestions`} 生成 ephemeral bullet 改写建议 → 写 `ai_task_runs`（task_type=`resume_tailor`，suggestions 落 task 输出）→ 发射 `resume.tailor.completed`（`tailorRunId` = ai_task_run id，`resumeId`）。
- **改写采纳落盘（D-20）**：`getResumeTailorRun` 返回 run 状态 + ephemeral suggestions；用户客户端采纳后经 `updateResume`（覆盖原简历 `structured_profile`）或 `duplicateResume`（从现有 resume 复制 + 应用采纳改写，保存为新 resume）落盘。无服务端逐条 accept/reject 状态机。
- **隐私链路**：privacy_delete 调用 backend-resume 提供的 `DeleteResumesForUser` API；调 [backend-upload `DeleteFileObjectsForUser`](../backend-upload/spec.md) 删除 file binary（对象存储先删，成功后 hard delete `resumes` DB 行）。
- **B3 events 发射**：`resume.parse.completed` / `resume.tailor.completed` 在对应业务结果 ready 成功时通过 outbox 写入；失败路径写 `ai_task_runs` / audit / async retry metadata，不发 `*.completed` 事件。payload 字段（`resumeId` 等）必须与 [B3 §3.1.4](../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory) 一致。
- **mock-first 对齐**：本 subject 实现的 handler 响应字段集 / status code / IK 行为与 B2 fixture 字节比对，[mock-contract-suite C-9](../mock-contract-suite/spec.md#6-验收标准) 强制 enforce。

### 2.2 Out of Scope

- 前端 Resume Workshop UI（[frontend-resume-workshop](../frontend-resume-workshop/spec.md)）。
- file 上传 / 对象存储（[backend-upload](../backend-upload/spec.md)）；本 subject 通过 `backend-upload.RegisterFileObject(fileObjectId, expectedPurpose=resume, ownerUserId)` 校验 file 后，把 `resumes.file_object_id` 写入自身业务表。
- 真实 PDF 导出（`exportResume` P0 行为：`501 + RESUME_EXPORT_NOT_AVAILABLE`，[B2 D-26](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围)）；P1 实现归未来 plan。
- 简历内容的进阶 AI 能力：JD 匹配评分自动化 / 主动改写推送 / 知识检索：P0 不实现。
- 简历版本树 / 主版本 / 岗位定制版本 / 手动编辑记录：不属于当前 P0 Resume 合同。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 术语映射 | 后端、OpenAPI 和前端统一使用单一 `Resume` / `resumeId`；持久化表为 `resumes` | 与 [B2 D-26](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围)、[B4 D-22](../db-migrations-baseline/spec.md) 和当前 frontend-resume-workshop 扁平简历 UI 一致 |
| D-4 | parse 路径区分 | `RegisterResumeRequest.sourceType` 只允许 `upload | paste`：upload 必带 `fileObjectId`；paste 必带 `rawText`；其他组合返回 422 | [B2 D-26](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) schema 已约束；handler 层 enforce；字段由 `resumes` 表承接 |
| D-5 | tailor 模式 | `RequestResumeTailorRequest.mode` ∈ `gap_review | bullet_suggestions`（与 [B3 D-14](../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表) 对齐）；不启用非当前 `inline | rewrite | mirror` | events / API / DB 三层 mode enum 同源 |
| D-6 | RESUME_EXPORT_NOT_AVAILABLE 行为 | `exportResume` P0 默认返回 `501` + `error.code = "RESUME_EXPORT_NOT_AVAILABLE"`；P1 切到异步生成属于 additive 行为变化 | 类比 [B2 D-12 privacy export 例外](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围)；frontend toast 兜底 + copyText 真实可用 |
| D-7 | listResumes pagination | 默认 pageSize=20，cursor 分页；返回 `PaginatedResume` | 与 [B2 D-5](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) 分页规则一致 |
| D-8 | Resume side-effect operation 必带 IK | `registerResume` / `updateResume` / `duplicateResume` / `archiveResume` / `exportResume` / `requestResumeTailor` 共 6 个 side-effect operation 必带 `Idempotency-Key` | 防止网络抖动产生重复 resume、重复改写请求或重复归档 / 导出请求 |
| D-13 | 简历资产扁平化 | 当前合同是单一扁平 `resumes` 表 + 9 个 op：`registerResume` / `getResume` / `listResumes` / `updateResume` / `duplicateResume` / `archiveResume` / `exportResume` / `requestResumeTailor` / `getResumeTailorRun`。`registerResume` + parse 直接写当前 resume 的 `structured_profile`；`requestResumeTailor` / `getResumeTailorRun` 的 run + suggestions 落在 `ai_task_runs` task 输出；客户端采纳后通过 `updateResume` 覆盖或 `duplicateResume` 另存 | 对齐 [B2 D-26](../openapi-v1-contract/spec.md)、[B4 D-22](../db-migrations-baseline/spec.md)、[B3 D-17](../event-and-outbox-contract/spec.md) 与 [B1 D-20](../shared-conventions-codified/spec.md) |

### 3.2 待确认事项

暂无。新增 backend-resume 行为前必须先修订 B2 / B4 / B3 / B1 对应 owner spec，并在本 subject 中明确 operation、persistence、event 和 idempotency 边界。

## 4 设计约束

### 4.1 契约约束

- 实现 [B2 §3.1.1](../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) 全部 Resume operation 的 generated server interface；不允许私造 handler 签名。
- 响应字段集 / status code / IK 行为与 [B2 fixtures](../mock-contract-suite/spec.md) 字节比对；新增 scenario 必须 B2 plan 修订同步。
- 错误码必须 `$ref` [B1 D-5](../shared-conventions-codified/spec.md#31-已锁定决策) 已锁定的常量集 + [B1 D-10 RESUME_EXPORT_NOT_AVAILABLE](../shared-conventions-codified/spec.md#31-已锁定决策)；不私造未登记错误码。
- 异步 job 必须通过 [B3 jobs.yaml](../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表) 已登记的 `resume_parse` / `resume_tailor` canonical job_type；不私造 dotted task name。

### 4.2 AI 约束

- resume.parse / resume.tailor 必须通过 [A3 AIClient](../ai-provider-and-model-routing/spec.md) 调用；不允许业务代码 import 厂商 SDK / 直接 HTTP 调 model endpoint。
- prompt / rubric / 模型版本必须通过 [F3 registered feature_key](../prompt-rubric-registry/spec.md) 引用：`resume.parse`（model profile `resume.parse.default`）/ `resume.tailor.gap_review` / `resume.tailor.bullet_suggestions`；本 subject 不 hardcode prompt 正文。
- AI 输出必须含 `GenerationProvenance`（[B2 §4.6](../openapi-v1-contract/spec.md#46-ai-生成结果-provenance-约束)）；运行元数据写入 `ai_task_runs.model_profile_*` typed columns，当前 resume 内容写入 `resumes.structured_profile` / `display_name`。
- AI capability 仅消费 `chat`（[B1 D-8](../shared-conventions-codified/spec.md#31-已锁定决策)）；不引入 stt / realtime / judge / 向量检索。

### 4.3 存储约束

- `resumes.user_id` 是当前 backend-resume 的用户隔离根；不绕过 store 层直接 SQL。
- 跨用户隔离：所有 read endpoint 必须以 `user_id = current_user_id` 过滤；cross-user 访问返回 404（不暴露存在）。
- 隐私删除调用 `DeleteResumesForUser(userId)` 删除 `resumes` 行；file binary 与 `file_objects` 删除由 backend-upload `DeleteFileObjectsForUser` 在同一 privacy request 中按 B4 matrix 协调（对象存储删除成功后再 hard delete DB 行）。
- raw resume text（`resumes.original_text` / `raw_text` / `parsed_text_snapshot`）不出现在 audit_events / outbox / log 中（[B3 §3.1.4 PII 边界](../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory)）。

### 4.4 BDD / TDD 约束

- 每个 endpoint 必须有 handler unit test（参数校验 + IK + 错误路径）+ `cmd/api` route wiring test（session middleware / idempotency middleware / path params）+ store integration test（state transition + cross-user isolation）+ AI 调用 unit test（stub provider，验证 prompt/profile 路由正确）。
- 用户可见行为（register / list / update / duplicate / archive / tailor / parse 完成）必须有 BDD scenario 覆盖；涉及 async job 的场景必须通过 `cmd/api` in-process drainer 或等价真实 runtime harness 证明可执行，不得只验证包级 handler。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| 9 个 Resume / ResumeTailor HTTP handler | backend-resume | 真实业务逻辑 |
| `resumes` 表 schema | [B4 db-migrations-baseline](../db-migrations-baseline/spec.md) + [B4 002 plan](../db-migrations-baseline/plans/002-flat-resume-migration/plan.md) | 字段 / 索引 / FK / check constraint |
| file_object 引用 | [backend-upload](../backend-upload/spec.md) `Register` internal API | `resumes.file_object_id` 通过 backend-upload 引用 file_object |
| `resume.parse` / `resume.tailor` async job | backend-resume + backend-runtime-topology | job handler 注册到 `cmd/api` in-process drainer / runtime composition |
| `cmd/api` runtime wiring | backend-resume + backend-runtime-topology | 挂载 Resume route、idempotency middleware 与 in-process drainer；不得引入独立 worker 进程 |
| AI 调用 | [A3 AIClient](../ai-provider-and-model-routing/spec.md) + [F3 feature_key](../prompt-rubric-registry/spec.md) | backend-resume 只引用 profile，不绑定 provider |
| 隐私删除调用 | backend internal privacy runner（[backend-runtime-topology](../backend-runtime-topology/spec.md)） | 调用 `DeleteResumesForUser` |
| frontend Resume Workshop UI | [frontend-resume-workshop](../frontend-resume-workshop/spec.md) | 消费 generated TS client |
| mock-first fixtures | [B2 fixtures](../openapi-v1-contract/spec.md) + [openapi-v1-contract/004](../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) | backend-resume handler 响应字节比对 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | registerResume (upload) 主路径 | 已登录 + 有效 file_object (purpose=resume) + IK | 调 `POST /api/v1/resumes` `{sourceType: upload, fileObjectId, title, language}` | 返回 202 + `ResumeWithJob{resumeId, job(jobType=resume_parse, status=queued)}`；DB `resumes` 行 `parse_status='queued'`；触发 `resume.parse` async job | 001-asset-register-parse-and-listing |
| C-2 | registerResume (paste) | sourceType=paste with rawText | 调 register | 同 C-1 行为，但 `resumes.file_object_id` 为 NULL；paste 写 `source_type='paste'` + `original_text`；`parsed_text_snapshot` / `structured_profile` 仅由 parse job 后续写入 | 001 |
| C-3 | resume.parse async 完成 | resume.parse job consumer 处理 queued 行 | 通过 [A3 AIClient](../ai-provider-and-model-routing/spec.md) 调 model + parse JSON | DB `resumes.parse_status='ready'` + `parsed_summary` / `parsed_text_snapshot` / `structured_profile` 写入；触发 outbox `resume.parse.completed`（envelope `resumeId` 等字段集 [B3 §3.1.4](../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory) 一致）；ai_task_runs 行写入 typed columns | 001 |
| C-4 | resume.parse 失败 retryable | AI provider 返回 timeout / output_invalid | resume.parse 失败 | DB `resumes.parse_status='failed'` + 对应 `error_code`；retryable 由 `async_jobs` attempt metadata 表达；失败路径不发 `resume.parse.completed`；privacy 红线：error 不含 prompt / response 摘要 | 001 |
| C-5 | listResumes pagination | 用户 A 有 25 个 resume | 调 `GET /api/v1/resumes?pageSize=20` 然后 cursor | 第一页返回 20 行 + `pageInfo.nextCursor`；第二页返回 5 行 + `hasMore=false`；按 `updated_at DESC, id DESC` 唯一稳定序排序；cross-user 不可见 | 001 |
| C-6 | cross-user 隔离 | 用户 A 有 resume；用户 B 调 `getResume(A.resumeId)` | – | 404；不暴露存在；audit_events 不写入敏感字段 | 001 + 后续 plan |
| C-7 | IK replay | register 同 IK 重复调用 | – | 返回首次 `resumeId`；不创建新 DB 行 | 001 |
| C-8 | mock-first 字节比对 | B2 fixture `registerResume.json` `default` scenario | 通过 `cmd/api` route 调真实 handler | 响应字段集 / status / header 字节一致；session / IK middleware 不改变 generated response envelope | 001 + mock-contract-suite |
| C-9 | privacy 删除链路 | 用户 A 有 3 resume | privacy_delete job 触发 | backend-resume `DeleteResumesForUser` 删除 `resumes` 单表行；backend-upload 同一 privacy request 删除 file binary / file_objects（对象存储先删，成功后 DB hard delete）；audit tombstone 仅保留 ID / 删除时间，不含内容 | 后续 plan |
| C-12 | exportResume P0 | 调 `POST /api/v1/resumes/{resumeId}/exports` | – | 返回 501 + `error.code="RESUME_EXPORT_NOT_AVAILABLE"`；ai_task_runs 不写入；不消耗 model 配额 | 后续 plan |
| C-13 | events 漂移负向 | grep `inline\|rewrite\|mirror` 在 events / job / dispatcher 上下文 | – | 0 命中（与 [B3 D-14](../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表) 同步） | 001 + 002-tailor-runs-and-save-v1 |
| C-16 | resume.tailor.completed envelope | resume.tailor async job 处理 queued 改写请求成功结束 | 通过 [A3 AIClient](../ai-provider-and-model-routing/spec.md) 调 F3 `resume.tailor.gap_review` 或 `resume.tailor.bullet_suggestions` feature_key | DB `ai_task_runs`（task_type=`resume_tailor`）写 typed columns + ephemeral suggestions 落 task 输出；outbox `resume.tailor.completed` 唯一新增（envelope `tailorRunId`(=ai_task_run id) / `resumeId` / `targetJobId` / `mode` / `status` 与 [B3 §3.1.4](../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory) 一致；不含 suggested bullet 内容）；`getResumeTailorRun` 读 ai_task_run 返回 run + suggestions；失败路径（AI timeout / output_invalid / retry exhausted）不发 `resume.tailor.completed`，只写 `ai_task_runs` + `async_jobs` retry metadata | 002-tailor-runs-and-save-v1 |
| C-17 | updateResume 覆盖原简历（D-20） | 用户 A 拥有 resume + 采纳若干改写 + IK | 调 `PATCH /api/v1/resumes/{resumeId}` body `{structuredProfile, displayName?}` | 返回 200 + `Resume`（`structured_profile` / `display_name` 被覆盖）；cross-user 404；IK replay 返回首次结果不重复写；不创建新 resume | 002-tailor-runs-and-save-v1 |
| C-18 | duplicateResume 保存为新简历（D-20） | 用户 A 拥有 resume X + 采纳若干改写 + IK | 调 `POST /api/v1/resumes/{X}/duplicate` body `{structuredProfile?, displayName?}` | 返回 201 + 新 `Resume`（从 X 复制只读来源快照 + 应用传入 `structuredProfile`，分配新 `id`）；原 X 不变；cross-user 404；IK replay 返回首次新 resume 不重复创建 | 002-tailor-runs-and-save-v1 |

## 7 关联计划

- [001-asset-register-parse-and-listing](./plans/001-asset-register-parse-and-listing/plan.md)：第一批 plan，落地 `registerResume` + `getResume` + `listResumes` + `resume.parse` async job + `resume.parse.completed` event；BDD 覆盖 register → parse → list 主路径。D-20 phase 已将 handler / store / wiring / test 迁移到 `Resume` / `resumeId` / `resumes` 单表口径；`sourceType` 只保留 {`upload`,`paste`}；parse job 直接写 `resumes.structured_profile`。
- [002-tailor-runs-and-save-v1](./plans/002-tailor-runs-and-save-v1/plan.md)：当前已完成 flat Resume save / tailor 收口：`requestResumeTailor` / `getResumeTailorRun` 作用于 `resumeId`，suggestions 落 `ai_task_runs` 输出；`updateResume` 覆盖原简历，`duplicateResume` 另存为新简历；`resume.tailor.completed` envelope 使用 `resumeId` + `tailorRunId`。
- `003-export-and-archive-and-delete`（P1 延后）：落地 `exportResume` 真实 PDF 生成 + `archiveResume` + privacy delete 链路 fully integrate。
