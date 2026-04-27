# Event and Outbox Contract Spec

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-04-27

## 1 背景与目标

[engineering-roadmap spec §5.2](../engineering-roadmap/spec.md#52-layer-b--contract4-份全部-p0) 把 B3 `event-and-outbox-contract` 列为 Layer B · Contract 第三份 child（依赖 [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md)）。它把 [06-event-contracts.md](../../../easyinterview-tech-docs/06-event-contracts.md) 中 18 个内部事件落到代码契约层，决定了：

- 业务跨模块通信（API → Worker / Worker → 下游 handler / Worker → analytics）的统一 envelope；
- `outbox_events` 表与 dispatcher 协议（与 [03-db-definition.md §5.9](../../../easyinterview-tech-docs/03-db-definition.md) 一致）；
- 事件版本化（additive only / `eventVersion + 1` for breaking）与幂等规则。

[ADR-Q2](../engineering-roadmap/decisions/ADR-Q2-async-orchestration.md) 已锁定 Asynq + Redis 为唯一异步 runtime；[engineering-roadmap §3.1 D-2 Q-2](../engineering-roadmap/spec.md#32-w0-已锁定决策hard-gate--全部-accepted) 进一步约束「公共 API / DB / event / metrics 中 `jobType` 沿用 snake_case」「内部 Asynq handler 可用 dotted task name，但必须由 C8 / B3 / B4 显式维护映射」。本 spec 是该映射的 owner 之一。

目标是：

1. **18 个事件 envelope freeze**：每个事件有稳定 `eventName`、`eventVersion=1`、`payload` schema；W1 末锁定后只允许 additive。
2. **outbox 协议清晰**：业务事务 + outbox 双写 → dispatcher 轮询 → Asynq 投递 → consumer 幂等；本 spec 把这套流程定义为可被 [C8 `backend-async-runtime`](../engineering-roadmap/spec.md#53-layer-c--backend14-份p08--p14--p22) 实现的接口。
3. **public jobType ↔ Asynq dotted task name 映射**：B3 owns 这张映射表；新增 jobType 必须先改本 spec 再改 C8。
4. **避免事件爆炸**：本 spec 锁 18 个事件命名空间为 P0 全集；新事件必须有 spec 修订流程。

本 spec 不实现 dispatcher 进程（归 C8）、不实现具体 producer / consumer（归各业务域）、不创建 DB 表（归 [B4 `db-migrations-baseline`](./../db-migrations-baseline/spec.md)）。

## 2 范围

### 2.1 In Scope

- **事件 envelope 类型**：Go `pkg/events/envelope.go` + TS `frontend/src/lib/events/envelope.ts`（虽然前端不消费 outbox，但 analytics SDK 需要事件名 + version 字符串字面量）；由 [B1 generator](../shared-conventions-codified/spec.md#31-已锁定决策) 同源输出。
- **18 个事件 schema**：每个事件有 Go 结构体（`<EventName>Payload`）+ JSON Schema + TS 类型；由 [B1 generator](../shared-conventions-codified/spec.md#31-已锁定决策) 同源输出。
- **outbox 表 schema 引用**：`outbox_events` 表由 B4 落地；本 spec 锁定字段 `event_name` / `event_version` / `aggregate_type` / `aggregate_id` / `payload (jsonb)` / `publish_status` 与命名约束。
- **dispatcher 协议**：dispatcher 必须按 `created_at asc` + `publish_status='pending'` 拉取；至少 once 发布；成功后置 `published`，失败置 `failed` + 重试。
- **public jobType 字典**（与 [03 §5.9 async_jobs.job_type](../../../easyinterview-tech-docs/03-db-definition.md) 一致）：`target_import` / `resume_parse` / `report_generate` / `resume_tailor` / `debrief_generate` / `source_refresh` / `embedding_upsert` / `privacy_export` / `privacy_delete` 共 9 项。
- **public jobType ↔ Asynq dotted task name 映射表**：见 §3.1.1。
- **lint 规则**：禁止业务包 hardcode `eventName` / `jobType` 字符串；必须 `import constants from "events"` 包；A5 接入。
- **tooling**：`make codegen-events`（B1/B3 owner）；CI drift 校验。

### 2.2 Out of Scope

- dispatcher 进程实现 / Asynq handler 注册：归 C8 `backend-async-runtime`。
- 业务 producer（API 何时写 outbox）：归各 C 域。
- 业务 consumer（Worker 何时调用 AI）：归各 C 域。
- analytics 双发去重 / 前端事件埋点：归 [F2 `analytics-funnel`](../engineering-roadmap/spec.md#56-layer-f--quality-横切4-份)。
- DB schema 落地：归 B4。
- 事件 Trace 透传：归 [F1 `observability-stack`](./../observability-stack/spec.md)；本 spec 仅约定 `traceId` 字段必填。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策（含 jobType 映射表）

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | envelope 字段 | `eventId`（UUIDv7） / `eventName`（dot.case） / `eventVersion`（int，从 1 起）/ `aggregateType` / `aggregateId` / `occurredAt`（RFC3339）/ `producer`（`api`/`worker`/`dispatcher`）/ `traceId` / `payload` | 与 [06 §3](../../../easyinterview-tech-docs/06-event-contracts.md#3-标准事件-envelope) 一致 |
| D-2 | 命名规则 | `<domain>.<aggregate>.<verb_past_tense>`；动词必须是过去式（已发生事实），如 `report.generated` / `practice.session.completed`；禁止 `something.updated` / `entity.changed` 等模糊命名 | – |
| D-3 | 18 个事件全集（P0） | 见 §3.1.2；任一新增由本 spec 修订 | – |
| D-4 | 版本化 | additive：新增 optional payload 字段 / 新增消费者；breaking：`eventVersion + 1` 且新旧并存一段时间，consumer 显式分支 | – |
| D-5 | 幂等去重 key | 消费方至少基于 `eventId` 或 `aggregateType + aggregateId + eventName + eventVersion` 去重；Asynq job 基于 `job_type + dedupe_key` | 防重复执行 |
| D-6 | outbox 投递语义 | At-least-once；消费方必须幂等；同一事件可能重复投递 | – |
| D-7 | dispatcher 拉取节奏 | 每秒至少一次扫描 `publish_status='pending'`；批量 ≤ 100；失败按指数退避（max 5 次） | – |
| D-8 | 死信策略 | 重试 5 次后置 `publish_status='failed'`；触发 P2 告警；进入人工排查队列 | 与 [04 §11.2](../../../easyinterview-tech-docs/04-metrics-observability.md#112-p2中优先级) 对齐 |
| D-9 | metric 接入 | 必产 `outbox_events_pending` Gauge / `outbox_publish_duration_seconds` Histogram / `outbox_publish_failures_total` Counter；F1 接入 dashboard | – |
| D-10 | trace 字段必填 | `traceId` 必须从 producer 透传；缺失时 dispatcher 写 warn log 并允许 publish | 故障定位 |
| D-11 | jobType ↔ dotted task name 映射 | 见 §3.1.1；新增 jobType 必须先改本 spec 再改 C8 | 防止 C8 私自加 dotted task name |

#### 3.1.1 public jobType ↔ Asynq dotted task name 映射

| public `jobType`（snake_case） | Asynq dotted task name | 触发事件 | Owner C 域 |
|-------------------------------|------------------------|----------|-----------|
| `target_import` | `target.import` | `target.import.requested` | C4 |
| `resume_parse` | `resume.parse` | API: register resume | C7 |
| `report_generate` | `report.generate` | `practice.session.completed` | C6 |
| `resume_tailor` | `resume.tailor` | API: request tailor | C7 |
| `debrief_generate` | `debrief.generate` | `debrief.created` | C9（P1） |
| `source_refresh` | `source.refresh` | scheduled / `target.parsed` | C13（P2） |
| `embedding_upsert` | `embedding.upsert` | `target.parsed` / `resume.parse.completed` | C11（P1） |
| `privacy_export` | `privacy.export` | `privacy.request.created`（P1） | C12（P1） |
| `privacy_delete` | `privacy.delete` | `privacy.request.created` | C8（P0 删除链路） |

> 备注：[engineering-roadmap §4.4](../engineering-roadmap/spec.md#44-layer-f-横切约束) 已说明 P0 删除链路核心实现下沉到 C8 `privacy_delete` public jobType（内部 Asynq handler 可映射为 `privacy.delete`）。

#### 3.1.2 18 个事件全集（v1）

| # | eventName | producer | consumer 默认集 | aggregateType | 关联 jobType |
|---|-----------|----------|----------------|---------------|--------------|
| 1 | `target.import.requested` | api | dispatcher → `target_import` worker | `target_job` | `target_import` |
| 2 | `target.parsed` | worker | retrieval upsert / analytics | `target_job` | `embedding_upsert`（可选） |
| 3 | `target.analysis_failed` | worker | analytics / alerting | `target_job` | – |
| 4 | `practice.session.started` | api | analytics | `practice_session` | – |
| 5 | `practice.turn.completed` | api | analytics / quality sampler | `practice_turn` | – |
| 6 | `practice.session.completed` | api | report job creator / analytics | `practice_session` | `report_generate` |
| 7 | `report.generation.requested` | api / dispatcher | worker | `feedback_report` | `report_generate` |
| 8 | `report.generated` | worker | mistake updater / growth aggregator / analytics | `feedback_report` | – |
| 9 | `report.generation_failed` | worker | analytics / alerting | `feedback_report` | – |
| 10 | `mistake.created` | worker / review | growth / analytics | `mistake_entry` | – |
| 11 | `mistake.status_changed` | review | growth / analytics | `mistake_entry` | – |
| 12 | `resume.parse.completed` | worker | retrieval upsert / analytics | `resume_asset` | `embedding_upsert`（可选） |
| 13 | `resume.tailor.completed` | worker | analytics | `resume_tailor_run` | – |
| 14 | `debrief.created` | api | debrief worker / analytics | `debrief` | `debrief_generate`（P1） |
| 15 | `debrief.completed` | worker | mistake extractor / growth / analytics | `debrief` | – |
| 16 | `source.refreshed` | worker | source cache / analytics | `source_record` | – |
| 17 | `privacy.request.created` | api | privacy worker / audit | `privacy_request` | `privacy_delete` / `privacy_export`（P1） |
| 18 | `privacy.request.completed` | worker | audit / notification（future） | `privacy_request` | – |

### 3.2 待确认事项

- 是否在 P0 实现「事件重放工具」（从 outbox 历史重新投递给 consumer）：默认不实现；W4/W5 时由 E4 决策。
- 是否引入 Schema Registry（独立 schema 服务）：默认不引入；P0 直接用 generator 出 JSON Schema 文件。
- 跨进程 dedupe 是否使用 Redis SETNX（短 TTL）vs PG unique key：默认 PG unique（与 outbox dedupe_key 一致）；高频事件如发现性能瓶颈再决策。

## 4 设计约束

### 4.1 命名约束

- `eventName` 一律 `dot.case` + 过去式动词；前缀 8 个固定 domain：`target / practice / report / mistake / resume / debrief / source / privacy`。
- `aggregateType` 一律 `snake_case`，与 [03 §4 表名](../../../easyinterview-tech-docs/03-db-definition.md#4-表清单) 单数版本一致（`feedback_report` / `practice_session` 等）。
- public `jobType` 一律 `snake_case`；Asynq dotted task name 一律 `snake_case.snake_case`；映射表见 §3.1.1。

### 4.2 schema 约束

- payload 字段类型必须由 [B1 共享类型](../shared-conventions-codified/spec.md) 提供（避免再造 enum）；新字段必须在 B1 添加后再加入 payload。
- 所有 payload 字段在 v1 中均为 optional 或 required 之一；后续 additive 只允许新增 optional 字段。
- 数值字段（`durationMs` / `tokenCount` / `requirementCount`）必须明确单位（ms / count / chars），单位通过字段名后缀或 schema description 表达。

### 4.3 outbox 协议约束

- producer 必须在同一 DB 事务内写业务表 + 写 `outbox_events` 行；禁止「先写库后发消息」非事务版本。
- dispatcher 必须维护一个进程级单例（同一 DB 行不被多副本同时拉取）；通过 `SELECT ... FOR UPDATE SKIP LOCKED` 实现。
- consumer 必须先校验事件 schema 再处理；schema 失败 → log error + 不更新 status（让重试或人工处理）。

### 4.4 lint 与 codegen 约束

- 业务包不允许出现裸字面量 `"target.parsed"` / `"report_generate"`；必须 import `events` / `jobs` 包常量。
- generator 输入：`shared/events.yaml`（envelope schema + 18 事件清单）+ `shared/jobs.yaml`（jobType ↔ dotted name 映射）；与 [B1 D-1](../shared-conventions-codified/spec.md#31-已锁定决策) 同 generator 进程。
- CI `git diff --exit-code` 校验无漂移。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `shared/events.yaml` + `shared/jobs.yaml` 真理源 | B3 + B1 generator | B3 owns 内容；B1 owns generator |
| Go `internal/shared/events/` + `internal/shared/jobs/` 常量 | B3 | 通过 B1 generator 生成 |
| TS `frontend/src/lib/events/` + `frontend/src/lib/jobs/` 常量 | B3 | 通过 B1 generator 生成 |
| `outbox_events` 表 schema | B4 | B3 提供字段名 + 类型 |
| dispatcher 进程实现 | C8 | 按本 spec 协议实现 |
| 业务 producer / consumer | 各 C 域 | 通过常量包引用 |
| Trace 透传 | F1 + 各 C 域 | B3 仅锁字段必填 |
| analytics 双发去重 | F2 | 与本 spec 18 事件命名空间共享 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | envelope schema 生成 | `shared/events.yaml` 落地 | `make codegen-events` | Go + TS envelope 类型 + 18 个事件 payload 类型生成；CI drift 通过 | B3 后续 001 + B1 generator |
| C-2 | jobType 常量生成 | `shared/jobs.yaml` 落地 | `make codegen-events` | Go `jobs.JobTypeTargetImport` 等 9 个常量；TS 同步；映射表中 dotted name 由 C8 引用 | B3 后续 001 |
| C-3 | outbox 双写 | 业务事务写 `target_jobs` + 写 `outbox_events('target.import.requested')` | 事务提交 / 回滚 | 提交后两行并存；回滚后两行均不存在；不可能出现 `target_jobs` 提交但 outbox 缺失 | B3 后续 001 + B4 + C4 |
| C-4 | dispatcher at-least-once | dispatcher 多次拉取同一行 | dispatcher | 同一行只投递一次成功（`SKIP LOCKED`）；网络抖动可能重复投递；consumer 必须幂等 | B3 后续 001 + C8 |
| C-5 | consumer 幂等 | 同一 `eventId` 投递两次 | consumer | 业务表只更新一次；db unique 约束阻止重复 mistake / report 行 | B3 后续 001 + 各 C 域 |
| C-6 | breaking change 拦截 | 故意把 `report.generated` 的 `mistakeCount` 改为 string | CI | `lint-events` 失败；提示需 `eventVersion + 1` | B3 后续 001 + A5 |
| C-7 | dotted name 映射一致 | 业务包 import `jobs.AsynqTaskTargetImport` | 编译 | 等于 `"target.import"`；与 §3.1.1 表一致 | B3 后续 001 |
| C-8 | privacy.delete P0 路径 | 用户调用 `POST /privacy/deletions` | API + dispatcher | 触发 `privacy.request.created` → dispatcher → Asynq dotted `privacy.delete` → C8 worker | B3 后续 001 + C8 + C12 |
| C-9 | metric 接入 | dispatcher 运行 | F1 dashboard | `outbox_events_pending` 可见；积压 > 100 时告警；与 [04 §11.1](../../../easyinterview-tech-docs/04-metrics-observability.md#111-p1高优先级影响核心主链路) 对齐 | B3 后续 001 + F1 |
| C-10 | analytics 命名空间不冲突 | F2 在 PostHog 注册产品分析事件 | grep 全部 eventName | 产品分析事件名（如 `target_import_requested` snake_case）与本 spec 18 个 internal eventName（`target.import.requested` dot.case）属于不同命名空间，不互相影响 | B3 + F2 |

## 7 关联计划

B3 在本次 W1 spec 阶段不创建 impl plan（参见 [001-decompose-subspecs §3.1](../engineering-roadmap/plans/001-decompose-subspecs/plan.md#3-实施步骤)）。后续由 B3 自身的 `001-bootstrap`（W1 末或 W2 初）承接：

- 落地 `shared/events.yaml` + `shared/jobs.yaml` 真理源。
- 接入 [B1 generator](../shared-conventions-codified/spec.md#21-in-scope) 输出 Go / TS 常量与类型。
- 提供 `make lint-events` 检查业务包是否使用裸字面量。
- 落地 `make codegen-events` 与 CI drift 接入。

后续如需新增事件 / 升级 eventVersion / 新增 jobType：递增 spec 版本 + history；映射表 §3.1.1 全文同步更新。
