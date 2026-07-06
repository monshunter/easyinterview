# Backend Resume Spec

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-07-06

## 1 背景与目标

[engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 标记 Resume Workshop workstream 候选 subject 包含 `backend-resume`（C7 ownerDomain，对齐 [shared/jobs.yaml](../event-and-outbox-contract/spec.md#311-dbbackend-runner-canonical-job_type--asynq-dotted-task-name-映射)）。本 subject 是 Resume 业务域的后端 owner：承载 `resumes`（product-scope v2.1 D-20 简历扁平化后的单一扁平简历资产表，原 `resume_assets` 重命名 + 合并结构化内容）的 store 与 handler，组织 resume.parse / resume.tailor 异步 AI 编排，并把 Resume 业务事件（`resume.parse.completed` / `resume.tailor.completed`）发射到 [B3 events](../event-and-outbox-contract/spec.md)。

> **D-20 简历扁平化（product-scope v2.1）**：简历是平铺列表中的独立扁平资产，不区分原始 / 结构化主版本 / 岗位定制版本，不做版本树 / 分叉 / 继承。原 `resume_versions` / `resume_version_suggestions` / `resume_tailor_runs` 三表已由 [B4 D-22](../db-migrations-baseline/spec.md) 删除，结构化内容（`structured_profile` / `display_name`）合并进 `resumes`；契约 op 由 [B2 D-26](../openapi-v1-contract/spec.md) 坍缩为 9 个，`resumeAssetId` / `resumeVersionId` 统一为 `resumeId`。本 spec §3.1 的版本树相关决策（D-2 / D-3 / D-9 / D-10 / D-11 / D-12）已由 D-13 退役 / 重塑。

目标：

1. **完整业务域**：所有 Resume HTTP endpoint（B2 D-26 后共 9 个 op：`registerResume` / `getResume` / `listResumes` / `updateResume` / `duplicateResume` / `archiveResume` / `exportResume` / `requestResumeTailor` / `getResumeTailorRun`）都在本 subject 落地 handler，并通过 `cmd/api` session middleware / idempotency middleware 挂到真实 `/api/v1/*` route；frontend 与 mock-contract-suite 消费同一份字节级响应。
2. **AI 编排封装**：resume.parse 与 resume.tailor 通过 [A3 AIClient + Capability Model Profile](../ai-provider-and-model-routing/spec.md) 调用；prompt / rubric / 模型版本通过 [F3](../prompt-rubric-registry/spec.md) 注册（3 个 baseline feature_key：`resume.parse` / `resume.tailor.gap_review` / `resume.tailor.bullet_suggestions`）；业务代码不 import 厂商 SDK。
3. **扁平资产语义（D-20）**：每份 resume 是独立扁平资产，承载只读原始来源快照（`original_text` / `parsed_text_snapshot` / `raw_text` / `file_object_id`）+ 可编辑结构化内容（`structured_profile` / `display_name`）；无版本树 / 分叉 / 主版本 / 岗位定制版本 / `parent_version_id` / `seed_strategy`。`registerResume` + `resume.parse` 直接产出该 resume 的 `structured_profile`，无独立 `confirmStructuredMaster` 主版本确认步骤。
4. **改写建议 ephemeral（D-20）**：`resume.tailor` 生成的 bullet 改写建议经 `getResumeTailorRun` 返回（run + suggestions 持久化在 `ai_task_runs` 的 task 输出中，不再有专属 `resume_tailor_runs` 表与逐条 `resume_version_suggestions` 状态）；用户客户端采纳后经 `updateResume`（覆盖原简历）或 `duplicateResume`（保存为新简历）落盘；不持久化逐条 `accepted | rejected` 状态，不提供 `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` op。
5. **mock-first 切真**：本 subject 落地的 9 个 endpoint 必须与 [B2 fixtures](../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) 及本 plan 新增 fixture 字节级一致，frontend-resume-workshop 在切真时无需修订；frontend mock-first 路径不阻塞本 subject 进度。

本 subject 不实现 frontend UI（归 [frontend-resume-workshop](../frontend-resume-workshop/spec.md)）；不实现 file 上传（归 [backend-upload](../backend-upload/spec.md)）；不实现 mistakes / growth / drill 等已丢弃模块（[roadmap D-6](../engineering-roadmap/spec.md#31-已锁定决策) 禁止恢复）。

## 2 范围

### 2.1 In Scope

- **HTTP handler + runtime wiring**：实现 [B2 §3.1.1](../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) Resumes + ResumeTailor tag 全部 9 个 operationId（D-20 扁平化后：`registerResume` / `getResume` / `listResumes` / `updateResume` / `duplicateResume` / `archiveResume` / `exportResume` / `requestResumeTailor` / `getResumeTailorRun`），并在 `cmd/api` 中按当前 session / IK / generated response envelope 口径挂载真实 route。
- **store layer**：`resumes` 单一扁平表 Repository（原 `resume_assets`，由 [B4 002 D-20 flatten phase](../db-migrations-baseline/plans/002-resume-versions-additive/plan.md) 重命名 + 合并 `structured_profile` / `display_name`，drop `resume_versions` / `resume_version_suggestions` / `resume_tailor_runs`）。`resumes` 字段：`id` / `user_id` / `file_object_id` / `title` / `display_name` / `language` / `source_type`∈{`upload`,`paste`} / `parse_status` / `parsed_summary` / `raw_text` / `original_text` / `parsed_text_snapshot` / `structured_profile` / timestamps / `deleted_at`。
- **AI 编排**：
  - `resume.parse` async job（[B3 jobs.yaml](../event-and-outbox-contract/spec.md#311-dbbackend-runner-canonical-job_type--asynq-dotted-task-name-映射) C7 owner）：解析 file_object / paste text → 提取结构化内容 → 写 `resumes.parsed_summary` / `parsed_text_snapshot` / `structured_profile`；最终 `parse_status='ready'` 时发射 `resume.parse.completed`（`resumeId`），失败路径只写 `ai_task_runs` / audit / async retry metadata，不发 completed event。`registerResume` + parse 直接产出 resume 的结构化内容，无独立主版本确认步骤。
  - `resume.tailor` async job：基于 `resumeId`（可选 `targetJobId` JD-aware 上下文）+ `mode`∈{`gap_review`,`bullet_suggestions`} 生成 ephemeral bullet 改写建议 → 写 `ai_task_runs`（task_type=`resume_tailor`，suggestions 落 task 输出）→ 发射 `resume.tailor.completed`（`tailorRunId` = ai_task_run id，`resumeId`）。不再写专属 `resume_tailor_runs` / `resume_version_suggestions` 表。
- **改写采纳落盘（D-20）**：`getResumeTailorRun` 返回 run 状态 + ephemeral suggestions；用户客户端采纳后经 `updateResume`（覆盖原简历 `structured_profile`）或 `duplicateResume`（从现有 resume 复制 + 应用采纳改写，保存为新 resume）落盘。无服务端逐条 accept/reject 状态机。
- **隐私链路**：privacy_delete 调用 backend-resume 提供的 `DeleteResumesForUser` API；调 [backend-upload `DeleteFileObjectsForUser`](../backend-upload/spec.md) 删除 file binary（对象存储先删，成功后 hard delete `resumes` DB 行）。
- **B3 events 发射**：`resume.parse.completed` / `resume.tailor.completed` 在对应业务结果 ready 成功时通过 outbox 写入；失败路径写 `ai_task_runs` / audit / async retry metadata，不发 `*.completed` 事件。payload 字段（`resumeId` 等）必须与 [B3 §3.1.4](../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory) 一致。
- **mock-first 对齐**：本 subject 实现的 handler 响应字段集 / status code / IK 行为与 B2 fixture 字节比对，[mock-contract-suite C-9](../mock-contract-suite/spec.md#6-验收标准) 强制 enforce。

### 2.2 Out of Scope

- 前端 Resume Workshop UI（[frontend-resume-workshop](../frontend-resume-workshop/spec.md)）。
- file 上传 / 对象存储（[backend-upload](../backend-upload/spec.md)）；本 subject 通过 `backend-upload.RegisterFileObject(fileObjectId, expectedPurpose=resume, ownerUserId)` 校验 file 后，把 `resumes.file_object_id` 写入自身业务表。
- 真实 PDF 导出（`exportResume` P0 行为：`501 + RESUME_EXPORT_NOT_AVAILABLE`，[B2 D-26](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围)）；P1 实现归未来 plan。
- 简历内容的进阶 AI 能力：JD 匹配评分自动化 / 主动改写推送 / 知识检索：P0 不实现。
- 简历版本树 / 主版本 / 岗位定制版本 / 手动编辑历史（`resume_version_edits`）：已由 product-scope v2.1 D-20 整体丢弃，不得恢复。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 术语映射（**D-13 重塑（D-20）**：坍缩为单一 `Resume` / `resumeId`，取代 `ResumeAsset` / `ResumeVersion`） | 后端真理源：`ResumeAsset`（OpenAPI）/ `ResumeVersion`（新 schema）/ `ResumeTailorRun`；UI 文档使用 `ResumeSource` / `ResumeVersion` 与之对应；前端通过 adapter 层 wrap UI 命名 | 与 [B1 D-10](../shared-conventions-codified/spec.md#31-已锁定决策) + [B2 D-18](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) 对齐；不重命名 OpenAPI 已发布 schema |
| D-2 | seed_strategy AI 触发（**已由 D-13 退役（D-20）**：无 branch op / 版本树） | `copy_master`：同步返回 + structured_profile 拷贝；`blank`：同步返回 + structured_profile 空；`ai_select`：同步返回 resume_version_id + 入队 `resume_tailor` job（202 + Job） | 前端在 ai_select 路径必须轮询 `getResumeTailorRun` 或监听对应 event；branch 必带 IK |
| D-3 | suggestion 状态终态（**已由 D-13 退役（D-20）**：改写建议 ephemeral，无持久化 accept/reject 状态） | `pending → accepted | rejected`，accept 与 reject 都是终态；改写历史不维护（如需变更必须 branch 新 version 或新 suggestion） | UI 与 store 都不实现 "撤销 accept" / "撤销 reject"；ConfirmDialog 由前端在用户行为前提示 |
| D-4 | parse 路径区分（**D-13 重塑（D-20）**：`sourceType`∈{`upload`,`paste`}，删除 `guided` + `guided_answers`） | `RegisterResumeRequest.sourceType` ∈ `upload | paste | guided`：upload 必带 `fileObjectId`；paste 必带 `rawText`；guided 必带 `guidedAnswers`；其他组合返回 422。guided 答案原样持久到 `resume_assets.guided_answers` jsonb，不序列化进 `original_text` | [B2 D-18](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) schema 已约束；本 spec 进一步在 handler 层 enforce；字段由 [B4 D-17](../db-migrations-baseline/spec.md#31-已锁定决策) 提供 |
| D-5 | tailor 模式 | `RequestResumeTailorRequest.mode` ∈ `gap_review | bullet_suggestions`（与 [B3 D-14](../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表) 对齐）；不复活旧 `inline | rewrite | mirror` | events / API / DB 三层 mode enum 同源 |
| D-6 | RESUME_EXPORT_NOT_AVAILABLE 行为（**D-13 重塑（D-20）**：`exportResumeVersion`→`exportResume`，作用于 `resumeId`） | `exportResumeVersion` P0 默认返回 `501` + `error.code = "RESUME_EXPORT_NOT_AVAILABLE"`；P1 切到 `202 + Job(jobType=resume_export)` 属 additive 行为变化 | 类比 [B2 D-12 privacy export 例外](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围)；frontend toast 兜底 + copyText 真实可用 |
| D-7 | listResumes / listResumeVersions pagination（**D-13 重塑（D-20）**：仅 `listResumes`，无 `listResumeVersions`） | 默认 pageSize=20，cursor 分页；返回 `PaginatedResumeAsset` / `PaginatedResumeVersion`（B2 D-18 schema） | 与 [B2 D-5](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) 分页规则一致 |
| D-8 | Resume side-effect operation 必带 IK（**D-13 重塑（D-20）**：集合改为 `registerResume` / `updateResume` / `duplicateResume` / `archiveResume` / `exportResume` / `requestResumeTailor` 共 6 个） | `registerResume` / `confirmResumeStructuredMaster` / `branchResumeVersion` / `updateResumeVersion` / `requestResumeTailor` / `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` / `archiveResumeAsset` / `exportResumeVersion` 共 9 个 side-effect operation 必带 `Idempotency-Key`（与 [B2 D-18](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) 及本 spec D-10 additive 一致） | 防止网络抖动产生重复 asset / version / tailor run / accept-reject 决策 |
| D-9 | 首次创建保存边界（**已由 D-13 退役（D-20）**：`registerResume` + parse 直接写 `resumes.structured_profile`，无独立 master 确认步骤） | `registerResume` 只登记 `ResumeAsset` source 并触发 `resume.parse`；parse job 只产出解析草稿（`parsed_summary` / `parsed_text_snapshot`）和 parse 状态，不在用户 Preview Confirm 前创建正式 `structured_master` `ResumeVersion`。确认保存 v1 由后续 backend-resume/002 + frontend-resume-workshop/002 承接。 | 对齐 `docs/ui-design/resume-onboarding.md` 的 `输入 -> Agent 解析 -> 预览确认 -> 保存 v1`；防止未确认草稿成为正式简历版本 |
| D-10 | Preview Confirm 保存 v1 op（**已由 D-13 退役（D-20）**：`confirmResumeStructuredMaster` op 删除，无 master 版本概念） | 新增第 14 个 Resume operationId `confirmResumeStructuredMaster`：`POST /api/v1/resumes/{resumeAssetId}/structured-master`，IK 必带（沿用 D-8 side-effect 集合），Request `ConfirmResumeStructuredMasterRequest{ structuredProfile, displayName, language? }`，Response `201 + ResumeVersion`（`version_type='structured_master'`, `parent_version_id=null`, `target_job_id=null`）。同 IK 重复调用走 idempotency middleware 返回首次结果；不同 fingerprint 同 key 走 409 generic IK conflict；同 asset 已存在未删除 structured_master 时返回 `409 + RESUME_STRUCTURED_MASTER_ALREADY_EXISTS`（B1 cross-owner 新增错误码）。本 op 是 B2 D-18 additive 增补，落地由 [backend-resume/002](./plans/002-versions-tailor-runs-and-save-v1/plan.md) Phase 1 携带 openapi-v1-contract spec / fixtures / inventory / generated artifacts 同步修订完成。 | 对齐 `docs/ui-design/resume-onboarding.md` Preview Confirm 行为；与 D-9 边界配套，使 ResumeAsset 解析草稿能在用户确认后落地为正式 `structured_master`，frontend-resume-workshop/002 切真不需要私造协议 |
| D-11 | structured_master 唯一性（**已由 D-13 退役（D-20）**：无 master 概念 / 无 `resume_versions` 表） | `resume_versions` 表新增 partial UNIQUE INDEX：`UNIQUE (resume_asset_id) WHERE version_type = 'structured_master' AND deleted_at IS NULL`；handler 层在事务内 `SELECT ... FOR UPDATE` 检查后插入。重复调用：若同 IK 走 idempotency replay；若不同 IK / fingerprint 命中已存在 structured_master，返回 `409 + RESUME_STRUCTURED_MASTER_ALREADY_EXISTS`；DB UNIQUE 兜底防止并发双客户端各自插入。 | 防止双客户端 / 双 tab 并发 Preview Confirm 创建两条主版本；DB 与 handler 双层保证；与 backend-practice D-22 idempotency 双层兜底同构 |
| D-12 | accept suggestion 不自动改 structured_profile（**已由 D-13 退役（D-20）**：无服务端 accept/reject op，改写采纳在客户端、经 `updateResume`/`duplicateResume` 落盘） | `acceptResumeTailorSuggestion` P0 仅写 `resume_version_suggestions.status='accepted'` + `decided_at`；不自动 patch `resume_versions.structured_profile`。如需将 suggestion 内容应用到版本，用户后续显式调用 `updateResumeVersion` 完成；reject 同样只写 `status='rejected'` + `decided_at`。 | 保留终态语义清晰；防止 accept 引入隐式 jsonb merge 风险与字段冲突；将"应用建议到版本"显式化为用户操作；与 §3.2 既有待确认事项中默认值一致 |
| D-13 | 简历资产扁平化（product-scope v2.1 D-20） | 简历坍缩为单一扁平 `resumes` 表 + 9 个 op（[B2 D-26](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围)）：`registerResume` / `getResume` / `listResumes` / `updateResume` / `duplicateResume` / `archiveResume` / `exportResume` / `requestResumeTailor` / `getResumeTailorRun`；删除版本树 / 主版本 / 岗位定制 / 改写 run / suggestion 持久化。**退役本 spec 旧决策**：D-2（seed_strategy branch — 无 branch op）、D-3（suggestion 终态状态机 — 无持久化 suggestion）、D-9（parse 草稿 vs master 确认边界 — parse 直接写 `resumes.structured_profile`）、D-10（`confirmResumeStructuredMaster` op 删除）、D-11（structured_master 唯一性 — 无 master 概念）、D-12（服务端 accept/reject — 删除）。**重塑**：D-1 术语坍缩为单一 `Resume`（`resumeId`，取代 `ResumeAsset`/`ResumeVersion`/`resumeAssetId`/`resumeVersionId`）；D-4 `sourceType`∈{`upload`,`paste`}（删除 `guided` + `guided_answers`）；D-5 tailor mode 保留 {`gap_review`,`bullet_suggestions`} 但作用于 `resumeId`、产出 ephemeral suggestions；D-6 `exportResumeVersion`→`exportResume`（保留 `501` + `RESUME_EXPORT_NOT_AVAILABLE`）；D-7 仅 `listResumes` 分页；D-8 side-effect IK 集合改为 `registerResume`/`updateResume`/`duplicateResume`/`archiveResume`/`exportResume`/`requestResumeTailor` 6 个。改写流程：`requestResumeTailor`→`getResumeTailorRun`（run + suggestions 落 `ai_task_runs` task 输出，`tailorRunId` = ai_task_run id）→ 客户端采纳 → `updateResume`（覆盖原简历 `structured_profile`）/ `duplicateResume`（保存为新简历）。 | 对齐 [B2 D-26](../openapi-v1-contract/spec.md) / [B4 D-22](../db-migrations-baseline/spec.md) / [B3 D-17](../event-and-outbox-contract/spec.md) / [B1 D-20](../shared-conventions-codified/spec.md)；§6 验收 C-10/C-11/C-14/C-15 退役、其余重塑为 `resumeId`/`resumes` 口径；由 backend-resume/001 + /002 的 D-20 phase 落地 |

### 3.2 待确认事项

- accept suggestion 时是否自动更新 `resume_versions.structured_profile`：已由 D-12 锁定 P0 不更新（仅写 suggestion.status = accepted）；P1 评估，由 frontend 与 backend 联合决定字段同步语义。
- branchResumeVersion seedStrategy=ai_select 入队的 tailor job 是否必须立即同步返回 suggestion 列表：默认 P0 异步（202 + Job + frontend 轮询）；如 frontend UX 需要同步响应，由本 spec 修订评估。
- `RESUME_*` 错误码扩展：D-10 / D-11 已新增 `RESUME_STRUCTURED_MASTER_ALREADY_EXISTS`；其余错误码（如 `RESUME_VERSION_BRANCH_FROM_INVALID_PARENT`）默认仍走 [B1 D-5](../shared-conventions-codified/spec.md#31-已锁定决策) 通用 `VALIDATION_FAILED`；如业务需要更细分类再修订 B1。

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

- 每个 endpoint 必须有 handler unit test（参数校验 + IK + 错误路径）+ `cmd/api` route wiring test（session middleware / idempotency middleware / path params）+ store integration test（state transition + cross-user isolation）+ AI 调用 unit test（stub provider，验证 prompt/profile 路由正确）。
- 用户可见行为（register / list / branch / accept / reject / parse 完成）必须有 BDD scenario 覆盖；涉及 async job 的场景必须通过 `cmd/api` in-process drainer 或等价真实 runtime harness 证明可执行，不得只验证包级 handler。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| 14 个 Resume / ResumeTailor HTTP handler | backend-resume | 真实业务逻辑 |
| `resume_assets` / `resume_versions` / `resume_version_suggestions` / `resume_tailor_runs` 表 schema | [B4 db-migrations-baseline](../db-migrations-baseline/spec.md) + [B4 002 plan](../db-migrations-baseline/plans/002-resume-versions-additive/plan.md) | 字段 / 索引 / FK / check constraint |
| file_object 引用 | [backend-upload](../backend-upload/spec.md) `Register` internal API | resume_assets 通过 backend-upload 引用 file_object |
| `resume.parse` / `resume.tailor` async job | backend-resume + backend-runtime-topology | job handler 注册到 `cmd/api` in-process drainer / runtime composition |
| `cmd/api` runtime wiring | backend-resume + backend-runtime-topology | 挂载 Resume route、idempotency middleware 与 in-process drainer；不得引入独立 worker 进程 |
| AI 调用 | [A3 AIClient](../ai-provider-and-model-routing/spec.md) + [F3 feature_key](../prompt-rubric-registry/spec.md) | backend-resume 只引用 profile，不绑定 provider |
| 隐私删除调用 | backend internal privacy runner（[backend-runtime-topology](../backend-runtime-topology/spec.md)） | 调用 `DeleteResumeAssetsForUser` |
| frontend Resume Workshop UI | [frontend-resume-workshop](../frontend-resume-workshop/spec.md) | 消费 generated TS client |
| mock-first fixtures | [B2 fixtures](../openapi-v1-contract/spec.md) + [openapi-v1-contract/004](../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) | backend-resume handler 响应字节比对 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | registerResume (upload) 主路径 | 已登录 + 有效 file_object (purpose=resume) + IK | 调 `POST /api/v1/resumes` `{sourceType: upload, fileObjectId, title, language}` | 返回 202 + `ResumeWithJob{resumeId, job(jobType=resume_parse, status=queued)}`；DB `resumes` 行 `parse_status='queued'`；触发 `resume.parse` async job | 001-asset-register-parse-and-listing |
| C-2 | registerResume (paste) | sourceType=paste with rawText | 调 register | 同 C-1 行为，但 `resumes.file_object_id` 为 NULL；paste 写 `source_type='paste'` + `original_text`；`parsed_text_snapshot` / `structured_profile` 仅由 parse job 后续写入（D-20 删除 `guided` 路径与 `guided_answers`） | 001 |
| C-3 | resume.parse async 完成 | resume.parse job consumer 处理 queued 行 | 通过 [A3 AIClient](../ai-provider-and-model-routing/spec.md) 调 model + parse JSON | DB `resumes.parse_status='ready'` + `parsed_summary` / `parsed_text_snapshot` / `structured_profile` 写入（D-20：parse 直接产出扁平 resume 结构化内容，无独立 `structured_master` 确认步骤）；触发 outbox `resume.parse.completed`（envelope `resumeId` 等字段集 [B3 §3.1.4](../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory) 一致）；ai_task_runs 行写入 typed columns | 001 |
| C-4 | resume.parse 失败 retryable | AI provider 返回 timeout / output_invalid | resume.parse 失败 | DB `resumes.parse_status='failed'` + 对应 `error_code`；retryable 由 `async_jobs` attempt metadata 表达；失败路径不发 `resume.parse.completed`；privacy 红线：error 不含 prompt / response 摘要 | 001 |
| C-5 | listResumes pagination | 用户 A 有 25 个 resume | 调 `GET /api/v1/resumes?pageSize=20` 然后 cursor | 第一页返回 20 行 + `pageInfo.nextCursor`；第二页返回 5 行 + `hasMore=false`；按 `updated_at DESC, id DESC` 唯一稳定序排序；cross-user 不可见 | 001 |
| C-6 | cross-user 隔离 | 用户 A 有 resume；用户 B 调 `getResume(A.resumeId)` | – | 404；不暴露存在；audit_events 不写入敏感字段 | 001 + 后续 plan |
| C-7 | IK replay | register 同 IK 重复调用 | – | 返回首次 `resumeId`；不创建新 DB 行 | 001 |
| C-8 | mock-first 字节比对 | B2 fixture `registerResume.json` `default` scenario | 通过 `cmd/api` route 调真实 handler | 响应字段集 / status / header 字节一致；session / IK middleware 不改变 generated response envelope | 001 + mock-contract-suite |
| C-9 | privacy 删除链路 | 用户 A 有 3 resume | privacy_delete job 触发 | backend-resume `DeleteResumesForUser` 删除 `resumes` 单表行（D-20 已删 versions / suggestions / tailor_runs 表，无级联子表）；backend-upload 同一 privacy request 删除 file binary / file_objects（对象存储先删，成功后 DB hard delete）；audit tombstone 仅保留 ID / 删除时间，不含内容 | 后续 plan |
| C-10 | ~~branchResumeVersion seedStrategy 三路~~（**已由 D-13 退役（D-20）**：无 branch / 版本树） | 用户 A 有 master version | 分别调 `copy_master` / `blank` / `ai_select` branch | `copy_master` 同步返回 + structured_profile 拷贝；`blank` 同步返回 + structured_profile 空；`ai_select` 同步返回 resume_version_id + 入队 resume.tailor job | 002-versions-tailor-runs-and-save-v1 |
| C-11 | ~~suggestion accept/reject 状态机~~（**已由 D-13 退役（D-20）**：改写 ephemeral，无服务端 accept/reject） | suggestion `pending` | 调 accept；再调 accept | 首次返回 200 + `decided_at` 写入；第二次返回 409 + `error.code = "VALIDATION_FAILED"`（或按 IK 语义幂等返回首次结果）；不私造未登记状态迁移错误码 | 002-versions-tailor-runs-and-save-v1 |
| C-12 | exportResume P0 | 调 `POST /api/v1/resumes/{resumeId}/exports` | – | 返回 501 + `error.code="RESUME_EXPORT_NOT_AVAILABLE"`；ai_task_runs 不写入；不消耗 model 配额 | 后续 plan |
| C-13 | events 漂移负向 | grep `inline\|rewrite\|mirror` 在 events / job / dispatcher 上下文 | – | 0 命中（与 [B3 D-14](../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表) 同步） | 001 + 002-versions-tailor-runs-and-save-v1 |
| C-14 | ~~confirmResumeStructuredMaster 主路径~~（**已由 D-13 退役（D-20）**：op 删除，parse 直接写 `resumes.structured_profile`） | 已登录 + 用户 A 拥有 `resume_assets` 行（`parse_status='ready'`，已含 `parsed_summary` / `parsed_text_snapshot`）+ 暂无 structured_master version + IK | 调 `POST /api/v1/resumes/{resumeAssetId}/structured-master` body `{ structuredProfile, displayName, language? }` | 返回 201 + `ResumeVersion`（`versionType='structured_master'`, `parentVersionId=null`, `targetJobId=null`，`structuredProfile.provenance` 含 `promptVersion='resume_profile.v1'` 或调用方提供值）；DB 新增 `resume_versions` 行；resume_asset `parse_status` 不变（确认动作不修改解析草稿状态）；IK 二次重放返回首次 ResumeVersion，不创建新行 | 002-versions-tailor-runs-and-save-v1 |
| C-15 | ~~structured_master 唯一性~~（**已由 D-13 退役（D-20）**：无 master 概念 / 无 `resume_versions` 表） | 用户 A 已有 1 行 `resume_versions(version_type='structured_master', deleted_at IS NULL)` 关联到 resume_asset X | 用户 A 用新 IK 再次 `POST /api/v1/resumes/{X}/structured-master` | 返回 `409 + error.code='RESUME_STRUCTURED_MASTER_ALREADY_EXISTS'`，不创建第二条 structured_master；DB partial UNIQUE INDEX 在并发场景下兜底（并发 INSERT 之一返回 409，不出现两条 master） | 002-versions-tailor-runs-and-save-v1 |
| C-16 | resume.tailor.completed envelope | resume.tailor async job 处理 queued 改写请求成功结束 | 通过 [A3 AIClient](../ai-provider-and-model-routing/spec.md) 调 F3 `resume.tailor.gap_review` 或 `resume.tailor.bullet_suggestions` feature_key | DB `ai_task_runs`（task_type=`resume_tailor`）写 typed columns + ephemeral suggestions 落 task 输出（D-20 已删 `resume_tailor_runs` / `resume_version_suggestions` 表，不再持久化逐条 suggestion 状态）；outbox `resume.tailor.completed` 唯一新增（envelope `tailorRunId`(=ai_task_run id) / `resumeId` / `targetJobId` / `mode` / `status` 与 [B3 §3.1.4](../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory) 一致；不含 suggested bullet 内容）；`getResumeTailorRun` 读 ai_task_run 返回 run + suggestions；失败路径（AI timeout / output_invalid / retry exhausted）不发 `resume.tailor.completed`，只写 `ai_task_runs` + `async_jobs` retry metadata | 002-versions-tailor-runs-and-save-v1 |
| C-17 | updateResume 覆盖原简历（D-20） | 用户 A 拥有 resume + 采纳若干改写 + IK | 调 `PATCH /api/v1/resumes/{resumeId}` body `{structuredProfile, displayName?}` | 返回 200 + `Resume`（`structured_profile` / `display_name` 被覆盖）；cross-user 404；IK replay 返回首次结果不重复写；不创建新 resume | 002-versions-tailor-runs-and-save-v1 |
| C-18 | duplicateResume 保存为新简历（D-20） | 用户 A 拥有 resume X + 采纳若干改写 + IK | 调 `POST /api/v1/resumes/{X}/duplicate` body `{structuredProfile?, displayName?}` | 返回 201 + 新 `Resume`（从 X 复制只读来源快照 + 应用传入 `structuredProfile`，分配新 `id`）；原 X 不变；cross-user 404；IK replay 返回首次新 resume 不重复创建 | 002-versions-tailor-runs-and-save-v1 |

## 7 关联计划

- [001-asset-register-parse-and-listing](./plans/001-asset-register-parse-and-listing/plan.md)：第一批 plan，落地 `registerResume` + `getResume` + `listResumes` + `resume.parse` async job + `resume.parse.completed` event；BDD 覆盖 register → parse → list 主路径。**D-20 phase（新增）**：`ResumeAsset`→`Resume`、`resumeAssetId`→`resumeId`、`resume_assets`→`resumes` store；`sourceType` 收敛双路 {`upload`,`paste`}（删除 `guided`）；parse job 直接写 `resumes.structured_profile`（无 master 确认步骤）；handler/store/wiring/test 全量迁移 `resumes` 单表口径。
- [002-versions-tailor-runs-and-save-v1](./plans/002-versions-tailor-runs-and-save-v1/plan.md)：历史 baseline 落地 D-10 `confirmResumeStructuredMaster` + 版本树 / branch / accept-reject suggestion / resume.tailor。**D-20 phase（新增，重塑）**：删除 `confirmResumeStructuredMaster` / `listResumeVersions` / `getResumeVersion` / `updateResumeVersion` / `branchResumeVersion` / `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` 7 个 handler + version/suggestion/tailor_run store + B4 addendum migration `000007`（随表 drop）；`requestResumeTailor` / `getResumeTailorRun` 重塑为作用于 `resumeId`、suggestions ephemeral（落 `ai_task_runs` 输出）；新增 `updateResume`（覆盖）/ `duplicateResume`（另存）handler；`resume.tailor.completed` envelope 改 `resumeId`；BDD `E2E.P0.074–080` 七个版本树场景退役，新增 update/duplicate/tailor-ephemeral 扁平场景。
- `003-export-and-archive-and-delete`（P1 延后）：落地 `exportResume` 真实 PDF 生成 + `archiveResume` + privacy delete 链路 fully integrate。
