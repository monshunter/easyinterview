# Event and Outbox Contract Spec

> **版本**: 2.9
> **状态**: active
> **更新日期**: 2026-07-06

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 B3 `event-and-outbox-contract` 定义为当前 active Contract spec（依赖 [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md)）。它把当前产品范围内的 14 个内部事件落到代码契约层，决定了：

- 业务跨模块通信（API → backend internal runner / runner → 下游 handler / runner → analytics）的统一 envelope；
- `outbox_events` 表与 dispatcher 协议（当前字段与索引以本 spec + B4 为准）；
- 事件版本化（additive only / `eventVersion + 1` for breaking）与幂等规则。

当前 internal event / job / outbox 可执行契约由本 spec、`shared/events.yaml`、`shared/jobs.yaml`、B4 migrations 与后续 generated artifacts 决定。B3 独立承接 `eventName`、`eventVersion`、payload schema、canonical `job_type`、Asynq task mapping、outbox retry/dead-letter 与 breaking baseline。

[ADR-Q2](../engineering-roadmap/decisions/ADR-Q2-async-orchestration.md) 已锁定 B3 job/outbox contract + backend internal runner 为 P0 异步运行形态；[engineering-roadmap §3.2 Q-2](../engineering-roadmap/spec.md#32-adr-q1q6-当前约束) 进一步约束「公共 API / DB / event / metrics 中 `jobType` 沿用 snake_case」「内部 handler 可用 dotted task name，但必须由 backend async runner / B3 / B4 显式维护映射」。本 spec 是该映射的 owner 之一。

目标是：

1. **14 个事件 envelope freeze**：每个事件有稳定 `eventName`、`eventVersion=1`、`payload` schema；baseline 锁定后只允许 additive。
2. **outbox 协议清晰**：业务事务 + outbox 双写 → backend internal runner 轮询 → handler 投递 → consumer 幂等；本 spec 把这套流程定义为可被 backend async runner 实现的接口。
3. **DB canonical jobType ↔ dotted task name 映射**：B3 owns 这张映射表；新增 DB / backend runner jobType 必须先改本 spec 再改 B4 / runner；能暴露到 B2 OpenAPI 的 API-facing subset 另行锁定；`email_dispatch` 是 C1 magic link / 通知派发的 internal-only canonical jobType，不进入 B2 `JobType`。
4. **避免事件爆炸**：本 spec 锁 14 个事件名为当前产品范围全集；新事件必须有 spec 修订流程。

本 spec 不实现 backend internal runner / dispatcher loop（由 active [backend-async-runner](../backend-async-runner/spec.md) subject 承接）、不实现具体 producer / consumer（归各业务域）、不创建 DB 表（归 [B4 `db-migrations-baseline`](./../db-migrations-baseline/spec.md)）。

## 2 范围

### 2.1 In Scope

- **事件 envelope 类型**：Go `backend/internal/shared/events/` + TS `frontend/src/lib/events/`（虽然前端不消费 outbox，但 analytics SDK 需要事件名 + version 字符串字面量）；由 B3 `codegen-events` 生成，复用 [B1](../shared-conventions-codified/spec.md) 已生成的 shared enum / ID / error helpers。
- **14 个事件 schema**：每个事件有 Go 结构体（`<EventName>Payload`）+ JSON Schema + TS 类型；字段清单见 §3.1.4，由 B3 `shared/events.yaml` 作为真理源生成。
- **outbox 表 schema 引用**：`outbox_events` 表由 B4 落地；本 spec 锁定字段 `event_name` / `event_version` / `aggregate_type` / `aggregate_id` / `payload (jsonb)` / `publish_status`，并追加 dispatcher 必需的 `publish_attempts` / `next_attempt_at` / `locked_at` / `last_error_code` / `last_error_message` operational columns。
- **dispatcher 协议**：dispatcher 必须按 `next_attempt_at asc, created_at asc` + `publish_status='pending'` 拉取；至少 once 发布；成功后置 `published`，临时失败保留 `pending` 并后移 `next_attempt_at`，达到上限后置 `failed`。
- **DB/backend runner canonical job_type 字典**：`target_import` / `resume_parse` / `report_generate` / `resume_tailor` / `source_refresh` / `privacy_export` / `privacy_delete` / `email_dispatch` 共 8 项。
- **DB/backend runner canonical job_type ↔ Asynq dotted task name 映射表**：见 §3.1.1；B2 API-facing subset 见 §3.1.2；`triggerEventSemantic` 表达触发事件是否由 dispatcher 创建 job，避免 source fact 被二次消费。
- **lint 规则**：禁止业务包 hardcode `eventName` / `jobType` 字符串；必须 `import constants from "events"` 包；当前由本地 lint gate 接入，远端 CI 仅在 A5 触发条件成立后再接入。
- **tooling**：`make codegen-events`（B3 owner）；本地 drift 校验。
- **breaking baseline**：提交 `shared/events/baseline/events.v1.json` 与 `shared/jobs/baseline/jobs.v1.json` 作为 v1 freeze manifest；`make lint-events` 比较当前 `shared/events.yaml` / `shared/jobs.yaml` 与 baseline，字段删除、类型变化、requiredness 变化、eventName/jobType 删除必须按 breaking 处理。

### 2.2 Out of Scope

- backend internal runner 实现 / handler 注册：归后续 backend async runner subject。
- 业务 producer（API 何时写 outbox）：归各 C 域。
- 业务 consumer（backend internal runner 何时调用 AI）：归各 C 域。
- analytics 双发去重 / 前端事件埋点：归 [F2 `analytics-funnel`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)。
- DB schema 落地：归 B4。
- 事件 Trace SDK / OTel 初始化：归 [F1 `observability-stack`](./../observability-stack/spec.md)；本 spec 仅约定 envelope 中 `traceId` 的 soft-required 透传语义。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策（含 jobType 映射表）

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | envelope 字段 | `eventId`（UUIDv7） / `eventName`（dot.case） / `eventVersion`（int，从 1 起）/ `aggregateType` / `aggregateId` / `occurredAt`（RFC3339）/ `producer`（`api`/`backend_async`/`dispatcher`/`review`）/ `traceId`（optional field, soft-required）/ `payload` | 当前 schema 以本 spec 与 `shared/events.yaml` 为准，允许无 `traceId`，producer 必须尽力从 F1 trace context 透传；`backend_async` 表示 backend 内部后台执行方，不是独立进程 |
| D-2 | 命名规则 | `<domain>[.<aggregate>].<verb_past_tense>`；允许 2 段或 3 段 dot.case；最后一段必须是过去式（已发生事实），如 `target.parsed` / `report.generated` / `practice.session.completed`；禁止 `something.updated` / `entity.changed` 等模糊命名 | – |
| D-3 | 14 个事件全集（当前产品范围） | 见 §3.1.3；任一新增由本 spec 修订；旧 `mistake.*`、`debrief.*`、`jd_match.*` events 均已随 product-scope D-17 / D-22 删除 | – |
| D-4 | 版本化 | additive：新增 optional payload 字段 / 新增消费者；breaking：`eventVersion + 1` 且新旧并存一段时间，consumer 显式分支 | – |
| D-5 | 幂等去重 key | 消费方至少基于 `eventId` 或 `aggregateType + aggregateId + eventName + eventVersion` 去重；Asynq job 基于 `job_type + dedupe_key` | 防重复执行 |
| D-6 | outbox 投递语义 | At-least-once；消费方必须幂等；同一事件可能重复投递 | – |
| D-7 | dispatcher 拉取节奏 | 至少每 5 秒扫描一次 `publish_status='pending' and next_attempt_at <= now()`；批量 ≤ 100；失败按 ADR-Q2 指数退避（30s/2m/10m/1h/6h，max 5 attempts） | 与 [ADR-Q2](../engineering-roadmap/decisions/ADR-Q2-async-orchestration.md) 对齐 |
| D-8 | 死信策略 | `publish_attempts >= 5` 后置 `publish_status='failed'`，保留 `last_error_code` / redacted `last_error_message`；触发 P2 告警；进入人工排查队列 | 告警阈值由 F1 active spec / alert rules 决定 |
| D-9 | metric 接入 | 必产 `outbox_events_pending` Gauge / `outbox_publish_duration_seconds` Histogram / `outbox_publish_failures_total` Counter；F1 接入 dashboard | – |
| D-10 | trace 字段 soft-required | `traceId` schema 上可选；producer 必须尽力从 W3C `traceparent` / active span 派生并写入；缺失时 dispatcher 写 warn log 并允许 publish；F1 backend runner span 只在存在 `traceId` 时重建父链路 | 对齐 F1 trace propagation |
| D-11 | canonical job_type ↔ dotted task name 映射 | 见 §3.1.1；新增 canonical `job_type` 必须先改本 spec，再同步 B4 check constraint、`triggerEventSemantic` 与 backend runner registry | 防止 runner 私自加 dotted task name 或错误消费 source event |
| D-12 | `email_dispatch` payload 红线 | `email_dispatch` 为 internal-only low-priority job；payload 只允许 `authChallengeId` / `userId` / `templateKey` / `locale` / `deliverySecretRef` / `dedupeKey` 等可审计字段，不得把 raw magic-link token、完整 magic-link URL、邮箱明文或邮件正文写入 `async_jobs.payload` / outbox / log；C1 owns `deliverySecretRef` 的一次性解析语义 | 支撑 ADR-Q1 magic link，同时避免 token / 邮件内容落库 |
| D-13 | `target.import.requested.sourceType` 语义 | `sourceType` 是异步导入请求的粗粒度输入来源，固定为 `url` / `text` / `file`；B2 `manual_text` 在事件中映射为 `text`；B2 `manual_form` 是同步 ready 兜底路径，不发 `target.import.requested`，不创建 runner 待处理事件 | 避免把 API source variant 与 async runner payload enum 混为一谈；如未来 analytics 需要 exact variant，只能新增 optional 字段或新事件版本，不能复用当前字段塞 `manual_form` |
| D-14 | `ResumeTailorMode` 漂移修复（已落地） | `shared/events.yaml` 中 `eventLocalEnums.ResumeTailorMode` 字面量已对齐为 `[gap_review, bullet_suggestions]`，与 [B2 `RequestResumeTailorRequest.mode`](../openapi-v1-contract/spec.md#42-schema-inventory-约束) 和 [B4 `resume_tailor_runs.mode`](../db-migrations-baseline/spec.md) check constraint 同步；本次修订属于已有契约漂移修复，baseline 期 `resume.tailor.completed` 无真实 producer/consumer，不按 breaking 处理；[event-and-outbox-contract/002-resume-tailor-mode-drift-fix](./plans/002-resume-tailor-mode-drift-fix/plan.md) 已同步 yaml、baseline manifest、Go/TS 生成类型、JSON Schema refs 与旧字面量精准负向搜索 | `shared/events.yaml` `eventLocalEnums.ResumeTailorMode` 字面量集；`shared/events/baseline/events.v1.json` baseline manifest；`backend/internal/shared/events/` Go 生成类型；`frontend/src/lib/events/` TS 生成类型；§3.1.4 `resume.tailor.completed.mode` 列 enum 值描述同步；与 B2 D-18 / B4 002 / B1 D-10 一并审查 |
| D-15 | `triggerEventSemantic` job ownership 语义 | `shared/jobs.yaml.jobs[*].triggerEventSemantic` 允许 `trigger_creates_job` / `source_event_only`；缺省为 `trigger_creates_job`。`report_generate` 必须显式为 `source_event_only`，表示 `practice.session.completed` 是 source event / analytics fact，不是 dispatcher 二次创建 report job 的指令。B3 codegen 必须生成 `JobTriggerEventSemantic*` 常量与 `IsSourceEventOnly(jobType)` 谓词；未来 backend async runner / dispatcher 必须读取该谓词并跳过 `source_event_only` 任务 | 支撑 backend-practice/002 D-32：`completePracticeSession` 同事务创建 `feedback_reports` placeholder 与 `async_jobs(report_generate)`，outbox 重放不得创建第二个 report/job |
| D-16 | JD Match events + jobs removed | product-scope D-17 已删除岗位推荐模块；`shared/events.yaml` / `shared/jobs.yaml` 当前不再包含 JD Match event 或 canonical job_type。历史 backend-jobs-recommendations additive 只保留在 deprecated subject 和 history 中，不是当前 active contract。 | product-scope/001-core-loop-module-pruning |
| D-17 | D-20 简历扁平化 resume event 重命名 | product-scope D-20 把简历绑定对象从 ResumeVersion 改为扁平 `resumes`（`resumeId`）。`resume.parse.completed` / `resume.tailor.completed` 的 `resumeAssetId` 字段重命名为 `resumeId`（= `resumes.id`），§3.1.3 event #10/#11 resource label `resume_asset` / `resume_tailor_run` 收敛为 `resume`。`resume_tailor` job_type 与 `resume.tailor.completed` 事件**保留**（AI 改写能力延续）：专属 `resume_tailor_runs` 持久化表已由 [B4 D-22](../db-migrations-baseline/spec.md) 删除，改写建议改为 ephemeral 即时返回，`tailorRunId` 改指 AI task run id；`ResumeTailorMode` event-local enum 与 `targetJobId`（可选 JD-aware 上下文）保留，最终随 [backend-resume](../backend-resume/spec.md) C7 D-20 reshape 确认。baseline 期 resume events 无真实 producer/consumer，按 fixture/docs-only 路径处理，不触发 breaking。由 D-20 contract impl phase 同步 `shared/events.yaml`、baseline manifest、Go/TS 生成类型、JSON Schema refs（随 contract regen 落地，不另起 B3 plan phase，参考 [B1 D-20](../shared-conventions-codified/spec.md) 同款无独立 plan phase 模式）。 | `shared/events.yaml` resume event payload；`shared/events/baseline/events.v1.json`；`shared/events/schemas/resume.parse.completed.v1.json` / `resume.tailor.completed.v1.json`；`backend/internal/shared/events/`；`frontend/src/lib/events/`；§3.1.3 / §3.1.4 已同步；与 B4 D-22 / B2 D-20 / B1 D-20 一并审查 |
| D-18 | Debrief events + jobs removed | product-scope D-22 已删除真实面试复盘模块；`debrief.created`、`debrief.completed` 与 `debrief_generate` 不再属于 current event/job truth source。复练能力只通过 `retry_current_round` / `next_round` 与 report-derived practice plan 表达。 | product-scope/001-core-loop-module-pruning |

#### 3.1.1 DB/backend runner canonical job_type ↔ Asynq dotted task name 映射

| canonical `job_type`（snake_case） | API-facing B2 `JobType`? | Asynq dotted task name | 触发事件 / 来源 | `triggerEventSemantic` | Owner C 域 |
|------------------------------------|--------------------------|------------------------|----------------|------------------------|-----------|
| `target_import` | yes | `target.import` | `target.import.requested` | `trigger_creates_job` | C4 |
| `resume_parse` | yes | `resume.parse` | API: register resume | `trigger_creates_job` | C7 |
| `report_generate` | yes | `report.generate` | `practice.session.completed` | `source_event_only` | C6 |
| `resume_tailor` | yes | `resume.tailor` | API: request tailor | `trigger_creates_job` | C7 |
| `source_refresh` | no（internal only） | `source.refresh` | scheduled / `target.parsed` | `trigger_creates_job` | C13（P2） |
| `privacy_export` | yes（P0 endpoint 501; P1 implemented） | `privacy.export` | `privacy.request.created`（P1） | `trigger_creates_job` | C12（P1） |
| `privacy_delete` | yes | `privacy.delete` | `privacy.request.created` | `trigger_creates_job` | C8（P0 删除链路） |
| `email_dispatch` | no（internal only） | `email.dispatch` | API: auth email start / notification producer | `trigger_creates_job` | C1 + C8 |

> 备注：[engineering-roadmap §4.4](../engineering-roadmap/spec.md#63-s2--backend-domain-implementation) 已说明 P0 删除链路核心实现下沉到 C8 `privacy_delete` canonical job_type（同时属于 B2 API-facing subset，内部 Asynq handler 可映射为 `privacy.delete`）。
> `triggerEventSemantic=source_event_only` 的 job 不由 dispatcher 根据 outbox source event 再创建 job；dispatcher / async runner 只能读取 generated `IsSourceEventOnly(jobType)` 谓词跳过该 binding，并由业务 handler 自己创建或查找既有 `async_jobs` row。

#### 3.1.2 B2 API-facing JobType subset

B2 OpenAPI v1.0.0 的 `JobType` enum 只允许以下 6 项：`target_import` / `resume_parse` / `report_generate` / `resume_tailor` / `privacy_export` / `privacy_delete`。`source_refresh` / `email_dispatch` 只能存在于 DB/backend runner 内部，不得出现在 `GET /api/v1/jobs/{jobId}` response、OpenAPI fixture 或前端 SDK 类型中；若未来需要对外暴露，必须先 additive 修订 B2 spec，再同步本 spec 与 B4 check constraint。

#### 3.1.3 14 个事件全集（v1）

| # | eventName | producer | consumer 默认集 | aggregateType | 关联 jobType |
|---|-----------|----------|----------------|---------------|--------------|
| 1 | `target.import.requested` | api | dispatcher → backend internal runner `target_import` | `target_job` | `target_import` |
| 2 | `target.parsed` | backend_async | analytics | `target_job` | – |
| 3 | `target.analysis.failed` | backend_async | analytics / alerting | `target_job` | – |
| 4 | `practice.session.started` | api | analytics | `practice_session` | – |
| 5 | `practice.turn.completed` | api | analytics / quality sampler | `practice_turn` | – |
| 6 | `practice.session.completed` | api | source fact / analytics（report job row 由 `completePracticeSession` 同事务创建） | `practice_session` | `report_generate` |
| 7 | `report.generation.requested` | api / dispatcher | backend internal runner | `feedback_report` | `report_generate` |
| 8 | `report.generated` | backend_async | report question-review / analytics | `feedback_report` | – |
| 9 | `report.generation.failed` | backend_async | analytics / alerting | `feedback_report` | – |
| 10 | `resume.parse.completed` | backend_async | analytics | `resume` | – |
| 11 | `resume.tailor.completed` | backend_async | analytics | `resume` | – |
| 12 | `source.refreshed` | backend_async | source cache / analytics | `source_record` | – |
| 13 | `privacy.request.created` | api | privacy runner / audit | `privacy_request` | `privacy_delete` / `privacy_export`（P1） |
| 14 | `privacy.request.completed` | backend_async | audit / notification（future） | `privacy_request` | – |

#### 3.1.4 v1 payload schema inventory

本表是 `shared/events.yaml` 的语义真理源。`required` 列中的字段在 v1 全部必填；后续只允许新增 optional 字段。`uuidv7` 使用 B1 `idx` 工具；B1 已有 enum 必须直接引用，B1 未覆盖但需要有界值的字段必须在 `shared/events.yaml` 中声明 event-local enum，不能散落为裸字符串。

| eventName | required payload fields | optional payload fields | enum / source | PII / logging boundary |
|-----------|-------------------------|-------------------------|---------------|------------------------|
| `target.import.requested` | `targetJobId:uuidv7`, `userId:uuidv7`, `sourceType:string`, `targetLanguage:string` | – | `sourceType` event-local `TargetImportSourceType` (`url`/`text`/`file`); B2 `manual_text` maps to `text`; B2 `manual_form` does not emit this event; `targetLanguage` BCP-47 | IDs only; no raw JD text / URL body |
| `target.parsed` | `targetJobId:uuidv7`, `userId:uuidv7`, `analysisStatus:TargetJobParseStatus`, `requirementCount:int`, `coreThemes:string[]` | – | B1 `TargetJobParseStatus`; `coreThemes` are controlled slugs | No parsed JD summary or requirement text |
| `target.analysis.failed` | `targetJobId:uuidv7`, `errorCode:string`, `retryable:bool` | – | `errorCode` UPPER_SNAKE_CASE producer-owned code | No raw provider response / prompt / JD text |
| `practice.session.started` | `sessionId:uuidv7`, `planId:uuidv7`, `targetJobId:uuidv7`, `goal:PracticeGoal`, `mode:PracticeMode`, `language:string` | – | B1 `PracticeGoal` / `PracticeMode`; `language` BCP-47 | IDs only; no question or answer text |
| `practice.turn.completed` | `sessionId:uuidv7`, `turnId:uuidv7`, `turnIndex:int`, `questionIntent:string`, `followUpCount:int`, `answerCharLength:int` | – | `questionIntent` controlled slug | Length/count only; no question / answer text |
| `practice.session.completed` | `sessionId:uuidv7`, `planId:uuidv7`, `targetJobId:uuidv7`, `turnCount:int`, `language:string` | – | `language` BCP-47 | IDs and counts only |
| `report.generation.requested` | `reportId:uuidv7`, `sessionId:uuidv7`, `targetJobId:uuidv7` | – | – | IDs only |
| `report.generated` | `reportId:uuidv7`, `sessionId:uuidv7`, `targetJobId:uuidv7`, `preparednessLevel:ReadinessTier`, `questionIssueCount:int`, `promptVersion:string`, `rubricVersion:string`, `modelId:string` | – | B1 `ReadinessTier`; F3 prompt/rubric version ids; A3 model profile id | No report body, answer snippets, raw model response, or prompt body |
| `report.generation.failed` | `reportId:uuidv7`, `sessionId:uuidv7`, `errorCode:string`, `retryable:bool` | – | `errorCode` UPPER_SNAKE_CASE producer-owned code | No raw provider response / prompt / answer text |
| `resume.parse.completed` | `resumeId:uuidv7`, `userId:uuidv7`, `parseStatus:TargetJobParseStatus` | – | B1 `TargetJobParseStatus` reused for queued/processing/ready/failed parse lifecycle；`resumeId` = 扁平 `resumes.id`（D-20，原 `resumeAssetId`） | No resume raw text or parsed summary |
| `resume.tailor.completed` | `tailorRunId:uuidv7`, `resumeId:uuidv7`, `targetJobId:uuidv7`, `mode:string`, `status:ReportStatus` | – | `mode` event-local `ResumeTailorMode`（`[gap_review, bullet_suggestions]`，与 B2 OpenAPI `RequestResumeTailorRequest.mode` 同步）; B1 `ReportStatus` (`ready`/`failed` subset when emitted)；D-20：`resumeId` = 扁平 `resumes.id`（原 `resumeAssetId`），`tailorRunId` = AI task run id（专属 `resume_tailor_runs` 表已由 D-22 删除，改写建议 ephemeral），`targetJobId` 仍可选携带 JD-aware 改写上下文 | No tailored bullet text |
| `source.refreshed` | `sourceRecordId:uuidv7`, `ownerType:string`, `ownerId:uuidv7`, `freshnessStatus:string` | – | `ownerType` event-local resource enum compatible with B2 `ResourceType` where API-facing; `freshnessStatus` event-local `SourceFreshnessStatus` | No source snapshot content or URL secret |
| `privacy.request.created` | `privacyRequestId:uuidv7`, `userId:uuidv7`, `requestType:PrivacyRequestType` | – | B1 `PrivacyRequestType` | Sensitive lifecycle event; no email / file URL / exported data |
| `privacy.request.completed` | `privacyRequestId:uuidv7`, `userId:uuidv7`, `requestType:PrivacyRequestType`, `status:PrivacyRequestStatus` | – | B1 `PrivacyRequestType` / `PrivacyRequestStatus` | Sensitive lifecycle event; no deleted/exported data |

### 3.2 待确认事项

- 是否在 P0 实现「事件重放工具」（从 outbox 历史重新投递给 consumer）：默认不实现；由后续 release workstream 决策。
- 是否引入 Schema Registry（独立 schema 服务）：默认不引入；P0 直接用 generator 出 JSON Schema 文件。
- 跨进程 dedupe 是否使用 Redis SETNX（短 TTL）vs PG unique key：默认 PG unique（与 outbox dedupe_key 一致）；高频事件如发现性能瓶颈再决策。

## 4 设计约束

### 4.1 命名约束

- `eventName` 一律 `dot.case`，允许 2 段或 3 段；最后一段必须是过去式动词；前缀 6 个固定 domain：`target / practice / report / resume / source / privacy`。旧 `mistake`、`debrief`、`jd_match` domain 已随模块裁剪删除；多词状态必须拆为独立 dot segment，例如 `report.generation.failed`，禁止 `generation_failed` 这类 snake segment。
- `aggregateType` 一律 `snake_case`，与 B4 当前表名 / B2 API-facing resource name 的单数语义一致（`feedback_report` / `practice_session` 等）。
- DB/backend runner canonical `job_type` 一律 `snake_case`；B2 API-facing `JobType` 只能使用 §3.1.2 subset；Asynq dotted task name 一律 `snake_case.snake_case`；映射表见 §3.1.1。

### 4.2 schema 约束

- payload 字段类型优先复用 [B1 共享类型](../shared-conventions-codified/spec.md)；B1 未覆盖但需要有界值的事件私有字段，必须在 `shared/events.yaml` 中声明 event-local enum，并在字段描述中写明不进入 B2 / B1 公共 enum。
- 所有 payload 字段在 v1 中均为 required（见 §3.1.4）；后续 additive 只允许新增 optional 字段。字段删除、重命名、类型变化、requiredness 变化、enum 移除/改名均视为 breaking，必须 `eventVersion + 1`。
- 数值字段（`durationMs` / `tokenCount` / `requirementCount`）必须明确单位（ms / count / chars），单位通过字段名后缀或 schema description 表达。
- payload 禁止携带 raw JD / raw resume / answerText / question text / prompt body / model raw response / file URL / email 等敏感明文；只能携带 ID、计数、状态、controlled slug 和版本号。

### 4.3 outbox 协议约束

- producer 必须在同一 DB 事务内写业务表 + 写 `outbox_events` 行；禁止「先写库后发消息」非事务版本。
- B4 必须在 `outbox_events` 中落地：`publish_attempts integer not null default 0`、`next_attempt_at timestamptz not null default now()`、`locked_at timestamptz`、`last_error_code text`、`last_error_message text`；`last_error_message` 必须是 redacted summary，不得写 raw provider response / prompt / answer。
- dispatcher 必须维护一个进程级单例（同一 DB 行不被多副本同时拉取）；通过 `SELECT ... FOR UPDATE SKIP LOCKED` 实现，查询条件必须包含 `publish_status='pending' and next_attempt_at <= now()`，排序为 `next_attempt_at asc, created_at asc`，批量 ≤ 100。
- publish 成功：`publish_status='published'` + `published_at=now()`；可重试失败：`publish_attempts += 1`，`next_attempt_at` 按 D-7 后移并保持 `pending`；达到 5 次：`publish_status='failed'`，保留 redacted last error 并触发告警。
- consumer 必须先校验事件 schema 再处理；schema 失败 → log error + 不更新业务状态，并按 dispatcher retry/dead-letter 语义处理。

### 4.4 lint 与 codegen 约束

- 业务包不允许出现裸字面量 `"target.parsed"` / `"report_generate"`；必须 import `events` / `jobs` 包常量。
- generator 输入：`shared/events.yaml`（envelope schema + 14 事件清单 + §3.1.4 payload schema）+ `shared/jobs.yaml`（8 个 canonical job_type ↔ dotted name 映射 + API-facing subset 标记 + `email_dispatch` payload redaction policy）；B3 owns `backend/cmd/codegen/events`，只 import B1 已生成类型，不复用 B1 generator 进程。
- 本地 drift gate：`make codegen-events && make lint-events && git diff --exit-code -- shared/events.yaml shared/jobs.yaml backend/internal/shared/events/{envelope.go,events.go} backend/internal/shared/jobs/jobs.go frontend/src/lib/events/{envelope.ts,events.ts} frontend/src/lib/jobs/jobs.ts shared/events/{schemas,refs,baseline} shared/jobs/baseline`；手写 `*_test.*` 与 fixtures 由 `make lint-events` / Go / TS 单测覆盖，不作为 generated drift 路径；远端 CI 仅在 A5 触发条件成立后再接入。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `shared/events.yaml` + `shared/jobs.yaml` 真理源 | B3 | B3 owns 内容；引用 B1 generated shared types |
| `backend/cmd/codegen/events` | B3 | B3-owned generator；模式参考 B1/B2 codegen，但不并入 B1 generator |
| Go `backend/internal/shared/events/` + `backend/internal/shared/jobs/` 常量 | B3 | 通过 B3 `codegen-events` 生成 |
| TS `frontend/src/lib/events/` + `frontend/src/lib/jobs/` 常量 | B3 | 通过 B3 `codegen-events` 生成 |
| `outbox_events` 表 schema | B4 | B3 提供字段名 + operational retry columns；B4 落 migration 与索引 |
| backend internal runner 实现 | [backend-async-runner](../backend-async-runner/spec.md) active subject | 按本 spec 协议实现 |
| 业务 producer / consumer | 各 C 域 | 通过常量包引用 |
| Trace 透传 | F1 + 各 C 域 | B3 仅锁 `traceId` optional field + soft-required producer 语义 |
| analytics 双发去重 | F2 | 与本 spec 14 个 internal eventName 命名空间共享 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | envelope schema 生成 | `shared/events.yaml` 落地 | `make codegen-events` + 本地 drift check | Go + TS envelope 类型 + 14 个事件 payload 类型 + JSON Schema 生成；本地 drift 通过；生成类型逐字段覆盖 §3.1.4 | B3 001 |
| C-2 | jobType 常量生成 | `shared/jobs.yaml` 落地 | `make codegen-events` | Go `jobs.JobTypeTargetImport` 等 8 个 canonical 常量 + dotted task name 常量生成；TS 同步；`source_refresh` / `email_dispatch` 标记为 internal-only，不进入 B2 API-facing `JobType` | B3 001 |
| C-3 | outbox 双写 | 业务事务写 `target_jobs` + 写 `outbox_events('target.import.requested')` | 事务提交 / 回滚 | 提交后两行并存；回滚后两行均不存在；不可能出现 `target_jobs` 提交但 outbox 缺失 | B3 后续 001 + B4 + C4 |
| C-4 | dispatcher at-least-once | dispatcher 多次拉取同一行 | dispatcher | 查询使用 `publish_status='pending' and next_attempt_at <= now()` + `FOR UPDATE SKIP LOCKED`；同一行只被一个 dispatcher 实例处理；网络抖动可能重复投递；consumer 必须幂等 | B3 后续 001 + C8 |
| C-5 | consumer 幂等 | 同一 `eventId` 投递两次 | consumer | 业务表只更新一次；db unique 约束阻止重复 report / privacy 行 | B3 后续 001 + 各 C 域 |
| C-6 | breaking change 拦截 | 故意把 `report.generated` 的 `questionIssueCount` 改为 string 或删除 required 字段，或把事件名写成 `report.generation_failed` | 本地 `make lint-events` | `lint-events` 失败；提示 schema breaking 需 `eventVersion + 1`，事件名必须为真正 dot.case；新增 optional 字段通过 | B3 后续 001 |
| C-7 | dotted name 映射一致 | 业务包 import `jobs.AsynqTaskTargetImport` | 编译 | 等于 `"target.import"`；与 §3.1.1 表一致 | B3 后续 001 |
| C-8 | privacy.delete P0 路径 | 用户调用 `POST /privacy/deletions` | API + backend internal runner | 触发 `privacy.request.created` → runner → dotted handler `privacy.delete` | B3 后续 001 + backend async runner |
| C-9 | metric 接入 | dispatcher 运行 | F1 dashboard | `outbox_events_pending` 可见；积压告警阈值以 F1 active spec / alert rules 为准 | B3 后续 001 + F1 |
| C-10 | analytics 命名空间不冲突 | F2 在 PostHog 注册产品分析事件 | grep 全部 eventName | 产品分析事件名（如 `target_import_requested` snake_case）与本 spec 14 个 internal eventName（如 `target.import.requested` dot.case）属于不同命名空间，不互相影响 | B3 + F2 |
| C-11 | outbox retry 字段承载 | B4 baseline migration 已完成 | `select column_name from information_schema.columns where table_name='outbox_events'` + retry 查询 explain | `publish_attempts` / `next_attempt_at` / `locked_at` / `last_error_code` / `last_error_message` 存在；pending + due 查询走索引；失败 5 次后可观察地进入 `failed` | B3 后续 001 + B4 + C8 |
| C-12 | `ResumeTailorMode` 对齐 | B3 002 drift-fix 已修订 `shared/events.yaml` 与 baseline manifest | `make codegen-events && make codegen-check && make lint-events` + 精准 grep + `python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml` | Go/TS generated events 只导出 `gap_review` / `bullet_suggestions`，旧 `inline` / `rewrite` / `mirror` 在 executable/generated/source truth 中 0 命中；B2 `RequestResumeTailorRequest.mode` 与 B4 `resume_tailor_runs.mode` 保持同一枚举集合 | event-and-outbox-contract/002 |

## 7 关联计划

B3 由以下 plan 承接：

- [001-bootstrap](./plans/001-bootstrap/plan.md)（已完成）：
  - 落地 `shared/events.yaml` + `shared/jobs.yaml` 真理源。
  - 落地 B3-owned `backend/cmd/codegen/events`，输出 Go / TS 常量、envelope、payload 类型与 JSON Schema，并复用 B1 generated shared types；`shared/jobs.yaml` 必须包含 internal-only `email_dispatch` 与 payload 红线。
  - 落地 `shared/events/baseline/events.v1.json` 与 `shared/jobs/baseline/jobs.v1.json` committed baseline manifests，供 `make lint-events` 执行 breaking-change 检测。
  - 提供 `make lint-events` 检查业务包是否使用裸字面量。
  - 落地 `make codegen-events` 与本地 drift check。

- [002-resume-tailor-mode-drift-fix](./plans/002-resume-tailor-mode-drift-fix/plan.md)：D-14 `ResumeTailorMode` 漂移修复落地。`shared/events.yaml` `eventLocalEnums.ResumeTailorMode` 已对齐为 `[gap_review, bullet_suggestions]`，与 B2 OpenAPI / B4 DB 同步；同步 `shared/events/baseline/events.v1.json` baseline manifest；codegen drift + grep negative search 验证。

本 spec 确认当前 event contract 不包含独立 `mistake` domain 事件，原错题本价值只保留在报告题目回顾 / 本轮复练字段中。后续如需新增事件 / 升级 eventVersion / 新增 canonical job_type / 调整 B2 API-facing subset：递增 spec 版本 + history；映射表 §3.1.1 / §3.1.2、事件全集 §3.1.3 与 payload schema §3.1.4 全文同步更新。
