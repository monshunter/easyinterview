# Event and Outbox Contract Bootstrap Checklist

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-04-29

**关联计划**: [plan](./plan.md)

## Phase 1: `shared/events.yaml` 真理源（envelope + 18 事件 payload）

- [ ] 1.1 落地 `shared/events.yaml` envelope schema：8 字段（`eventId` UUIDv7 / `eventName` dot.case / `eventVersion` int 起始 1 / `aggregateType` snake_case / `aggregateId` UUIDv7 / `occurredAt` RFC3339 / `producer` enum（`api` / `worker` / `dispatcher` / `review`）/ `traceId` optional, soft-required / `payload` polymorphic per eventName）；UUIDv7 / `aggregateId` 通过 `$ref`-style alias 引用 [B1 shared/conventions.yaml](../../../shared-conventions-codified/spec.md)，不复制
- [ ] 1.2 写入 `events:` 列表 18 项，每项含 `name` / `version: 1` / `producer` / `aggregateType` / `requiredPayload` / `optionalPayload` (v1 空 slot) / `piiBoundary`，与 [spec §3.1.3](../../spec.md#313-18-个事件全集v1) 与 [§3.1.4](../../spec.md#314-v1-payload-schema-inventory) 严格对齐；复用 B1 enum 字段以 alias 引用
- [ ] 1.3 在 `eventLocalEnums:` 顶层节落地 `TargetImportSourceType` (`url`/`text`/`file`) / `ResumeTailorMode` (`inline`/`rewrite`/`mirror`) / `SourceFreshnessStatus` (`fresh`/`stale`/`failed`) 三组 event-local enum，并在描述中写明「不进入 B1 / B2 公共 enum」
- [ ] 1.4 落地 `scripts/lint/events_inventory.py`（或等价 `make` 内联脚本）：断言事件数 == 18、`name` 唯一且匹配 dot.case 正则、动词来自过去式白名单、domain 来自 8 个固定前缀、`requiredPayload` 与 §3.1.4 一致、`producer` 与 §3.1.3 一致

## Phase 2: `shared/jobs.yaml` 真理源（10 canonical jobType + `email_dispatch` 红线）

- [ ] 2.1 落地 `shared/jobs.yaml.jobs:` 列表 10 项，每项含 `canonical` / `asynqTask` / `apiFacing` / `triggerEvent` / `ownerDomain` / `priority`，与 [spec §3.1.1](../../spec.md#311-dbc8-canonical-job_type--asynq-dotted-task-name-映射) 与 [ADR-Q2 §3.5](../../../engineering-roadmap/decisions/ADR-Q2-async-orchestration.md) 严格对齐
- [ ] 2.2 落地 `shared/jobs.yaml.apiFacingSubset:` 7 项 (`target_import` / `resume_parse` / `report_generate` / `resume_tailor` / `debrief_generate` / `privacy_export` / `privacy_delete`)；`source_refresh` / `embedding_upsert` / `email_dispatch` 标记 `apiFacing: false` 且不进入 subset
- [ ] 2.3 在 `email_dispatch.payloadSchema` 中只允许 `authChallengeId` / `userId` / `templateKey` / `locale` / `deliverySecretRef` / `dedupeKey`；落地 `redactedFields:` 黑名单 (`rawMagicLinkToken` / `magicLinkUrl` / `recipientEmail` / `recipientEmailHash` / `emailBody` / `emailSubject`)；与 [ADR-Q1 §3.4](../../../engineering-roadmap/decisions/ADR-Q1-auth.md) + [spec D-12](../../spec.md#31-已锁定决策含-jobtype-映射表) 一致
- [ ] 2.4 落地 dotted name 自检脚本：断言每条 `(canonical, asynqTask)` pair 与 §3.1.1 严格相等（`target_import ↔ target.import` 等）；任一笔误 fail

## Phase 3: B3 generator 与 Go / TS / JSON Schema 输出

- [ ] 3.1 落地 `backend/cmd/codegen/events/main.go` 作为 B3-owned generator（独立 Go 程序，**不**并入 [B1 backend/cmd/codegen/conventions](../../../shared-conventions-codified/plans/001-bootstrap/plan.md#13-写入-generator)）；通过 import 引用 B1 generated shared types，不修改 B1 源码
- [ ] 3.2 generator 输出 Go 端：`backend/internal/shared/events/envelope.go`（envelope struct + JSON marshal）+ `backend/internal/shared/events/events.go`（18 个 `EventName*` 常量 + 18 个 `<EventName>Payload` struct）+ `backend/internal/shared/jobs/jobs.go`（10 个 `JobType*` + 10 个 `AsynqTask*` + 7 项 `APIFacingJobTypes` slice + `email_dispatch` redaction policy 常量）
- [ ] 3.3 generator 输出 TS 端：`frontend/src/lib/events/envelope.ts`（discriminated union + `EventEnvelope<T>` generic）+ `frontend/src/lib/events/events.ts`（18 个 `EVENT_NAME_*` 字面量 + payload type alias + `EventNameToPayload` map）+ `frontend/src/lib/jobs/jobs.ts`（10 个 `JOB_TYPE_*` + 10 个 `ASYNQ_TASK_*` + `API_FACING_JOB_TYPES` readonly tuple + `email_dispatch` redaction policy export）
- [ ] 3.4 generator 输出 JSON Schema：`shared/events/schemas/<eventName>.v1.json`（Draft 2020-12，每个事件一个文件，envelope + payload；B1 enum 字段由 generator inline 当前值，或引用 B3-owned `shared/events/refs/<EnumName>.json` 桥接文件）；输出 idempotent
- [ ] 3.5 根 `Makefile` 追加 `codegen-events` target（调用 `go run ./backend/cmd/codegen/events`）；把 `make codegen` 升级为 `codegen: codegen-conventions codegen-events codegen-openapi`；`make help` 自动收录
- [ ] 3.6 落地 committed baseline manifests：`shared/events/baseline/events.v1.json` 与 `shared/jobs/baseline/jobs.v1.json` 由 generator 生成并提交；`make lint-events` 比较当前真理源与 baseline，breaking 变更必须 fail

## Phase 4: `make lint-events` 与本地 drift gate

- [ ] 4.1 落地 `scripts/lint/lint_events.py`（或等价 Go 工具）作为 `make lint-events` 实体：扫描 `backend/` 与 `frontend/`（白名单仅 `backend/internal/shared/{events,jobs}` 与 `frontend/src/lib/{events,jobs}`），拒绝 18 事件名 / 10 canonical jobType / 10 dotted task name 的裸字面量
- [ ] 4.2 `make lint-events` 校验 generated 文件中事件名集合长度 == 18 且与 `shared/events.yaml` 一致；任何 generated 文件之外手写 `EventName*` 常量声明 fail
- [ ] 4.3 `make lint-events` 校验 generated `APIFacingJobTypes` 长度 == 7 且与 `shared/jobs.yaml.apiFacingSubset` 一致；任一项 `apiFacing != true` fail（防止 `email_dispatch` 误扩）
- [ ] 4.4 落地本地 drift gate 命令：`make codegen-events && make lint-events && git diff --exit-code -- shared/events.yaml shared/jobs.yaml backend/internal/shared/events backend/internal/shared/jobs frontend/src/lib/events frontend/src/lib/jobs shared/events/schemas`；在 `Makefile` 注释中说明远端 CI 仅在 [A5 ci-pipeline-baseline](../../../ci-pipeline-baseline/spec.md) 触发条件成立后再接入

## Phase 5: 单元测试（envelope / trace 透传 / breaking-change 拦截 / `email_dispatch` 红线）

- [ ] 5.1 envelope round-trip 测试：Go (`backend/internal/shared/events/envelope_test.go`) + TS (`frontend/src/lib/events/envelope.test.ts`) 至少覆盖 3 个事件（`target.import.requested` / `report.generated` / `mistake.status.changed`），共享 fixture 在 `shared/events/__fixtures__/`，断言序列化 + 反序列化字段等值
- [ ] 5.2 `traceId` soft-required 双码路测试：缺失分支断言 producer 仅 warn log 且 publish 允许通过；存在分支断言 envelope 透传字面量；Go / TS 各 ≥ 2 case
- [ ] 5.3 breaking-change 拦截测试：4 类 negative fixture（类型变更 / required 字段删除 / dot.case 改 snake / enum 成员移除）必须让 `make lint-events` fail，错误信息含 `breaking change requires eventVersion + 1`
- [ ] 5.4 additive change 通过测试：positive fixture 给某事件新增 optional 字段，断言 `make codegen-events` + `make lint-events` 通过；生成的 Go struct 该字段为 pointer，TS 该字段为 `?:`
- [ ] 5.5 `email_dispatch` 红线测试：构造含 `rawMagicLinkToken` / `magicLinkUrl` / `recipientEmail` / `emailBody` 的 payload 必须被 `BuildEmailDispatchPayload` helper 拒绝；只含合法字段必须通过；fixture yaml 中偷加 redacted 字段必须被 Phase 4 lint 拒绝

## Phase 6: Verification + handoff

- [ ] 6.1 自检 spec C-1 / C-2 / C-6 / C-7：`make codegen-events` 跑两次后 `git status` clean；删除生成文件可还原；18 事件常量 + payload struct / type 与 §3.1.4 字段清单逐字段一致；`source_refresh` / `embedding_upsert` / `email_dispatch` 标 internal-only 不进 `APIFacingJobTypes`；临时 breaking change 让 `make lint-events` fail，revert 恢复；`jobs.AsynqTaskTargetImport == "target.import"` 等映射成立
- [ ] 6.2 自检 spec C-10：grep `target.import.requested` / `target_import_requested` / `report.generated` / `report_generated` 四种命名空间；dot.case 仅出现在 yaml / generated / 测试 fixture / 文档；snake_case 在本 plan 输出中**完全不出现**（属 [F2 analytics-funnel](../../../engineering-roadmap/spec.md#56-layer-f--quality-横切4-份) 命名空间）
- [ ] 6.3 在 plan §6.3 handoff 章节内列明 [B4 db-migrations-baseline](../../../db-migrations-baseline/spec.md) 的 `outbox_events` operational columns + 复合索引、[C8 backend-async-runtime](../../../engineering-roadmap/spec.md#53-layer-c--backend14-份p08--p14--p22) dispatcher 协议、[F1 observability-stack](../../../observability-stack/spec.md) 必产 metric、[F2 analytics-funnel](../../../engineering-roadmap/spec.md#56-layer-f--quality-横切4-份) 命名空间分离、[C1 backend-auth](../../../engineering-roadmap/spec.md#53-layer-c--backend14-份p08--p14--p22) `email_dispatch` 必须通过 `BuildEmailDispatchPayload` helper；明示本 plan **不写 SQL**；不新建独立 `handoff.md`
- [ ] 6.4 文档与 INDEX 同步：本 plan 自身 checklist 与 Phase 6 完成后将 plan/checklist Header 切到 `completed`，并用 `/sync-doc-index --fix-index` 同步 [event-and-outbox-contract/plans/INDEX.md](../INDEX.md)；`/sync-doc-index --check` 通过；不修改 [engineering-roadmap/001-decompose-subspecs Phase 3.3](../../../engineering-roadmap/plans/001-decompose-subspecs/checklist.md#phase-3-wave-1基础设施--契约骨架) 的 spawn 项；关键命令日志贴入工作日志
