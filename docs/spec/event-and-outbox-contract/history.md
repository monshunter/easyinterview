# Event and Outbox Contract History

> **版本**: 2.3
> **状态**: active
> **更新日期**: 2026-05-11

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-11 | 2.3 | D-14 `ResumeTailorMode` 漂移修复声明阶段：`eventLocalEnums.ResumeTailorMode` 当前 `[inline, rewrite, mirror]`，与 B2 OpenAPI `RequestResumeTailorRequest.mode`（`gap_review / bullet_suggestions`）+ B4 `resume_tailor_runs.mode` 不同步；本次声明对齐为 `[gap_review, bullet_suggestions]`；具体 yaml 修订与 baseline manifest 同步由 002 plan 落地。 | event-and-outbox-contract/002-resume-tailor-mode-drift-fix（声明阶段，docs-only） |
| 2026-05-09 | 2.2 | 对齐 backend-practice Phase 0：B3 generated event refs 中引用 B1 `PracticeMode` 的 surface 跟随二值化，`practice.session.started.mode` 继续只引用 B1 `PracticeMode`，不保留来源语义别名。 | backend-practice/001 Phase 0 |
| 2026-05-08 | 2.1 | 明确 `target.import.requested.sourceType` 是异步导入粗粒度来源：`manual_text` 映射为 `text`，`manual_form` 同步 ready 路径不发该事件；exact API source variant 不进入当前 v1 payload。 | backend-targetjob/001 Phase 0 |
| 2026-05-08 | 2.0 | 对齐 A3 003 Phase 6：删除 C11 当前内部任务占位，canonical job_type 从 10 项收敛为 9 项。 | ai-provider-and-model-routing/003 Phase 6 |
| 2026-05-06 | 1.9 | 对齐 backend-runtime-topology：event producer enum 将独立进程语义 `worker` 替换为 `backend_async`，job/outbox 契约保留并由 backend internal runner 消费。 | backend-runtime-topology/001-worker-consolidation |
| 2026-05-05 | 1.8 | B3 event / job / outbox 契约改为只由本 spec、`shared/events.yaml`、`shared/jobs.yaml`、B4 migrations 与 generated artifacts 承接；移除旧技术草稿名称和旧 shorthand 依赖。 | engineering-roadmap/001-decompose-subspecs |
| 2026-05-03 | 1.6 | 明确当前 event / job / outbox 契约由本 spec、`shared/events.yaml`、`shared/jobs.yaml` 与 B4 migrations 决定，旧 `mistake.*`、18 event inventory 与旧 consumer 不再作为实现依据。 | docs-only |
| 2026-05-03 | 1.5 | 对齐 product-scope v1.2：删除独立 `mistake` 事件 domain 与 `mistake.created` / `mistake.status.changed`，将 `report.generated.mistakeCount` 改为 `questionIssueCount`，将 `debrief.completed.generatedMistakeCount` 改为 `practiceFocusCount`，事件全集 18→16、domain 8→7。 | 001-bootstrap Phase 8 remediation |
| 2026-04-29 | 1.4 | 物化 B3 `001-bootstrap` active plan：新增 committed JSON Schema / job manifest baseline，明确 `email_dispatch` internal-only 口径由 ADR-Q1 同步，JSON Schema refs 由 B3 自有桥接或 inline 值承接，不依赖 B1 必须产出 JSON Schema fragment。 | [001-bootstrap](./plans/001-bootstrap/plan.md) |
| 2026-04-29 | 1.3 | 收口 A/B spec 全面审查 remediation：新增 internal-only `email_dispatch` canonical jobType 与 `email.dispatch` dotted task name，锁定 magic link payload 红线；把 `target.analysis.failed` / `report.generation.failed` / `mistake.status.changed` 改为真正 dot.case，避免 lint 规则拒绝 seed 事件；同步 C9 真实面试复现 P0 范围。 | plan-review remediation |
| 2026-04-29 | 1.2 | 根据 L1 review findings 修订 B3 契约：补齐 18 个事件 v1 payload schema inventory 与 PII 边界；拆分 DB/C8 canonical `job_type` 与 B2 API-facing `JobType` subset；明确 outbox retry operational columns、`traceId` soft-required 语义、B3-owned `codegen-events` 归属与 Go/TS 输出路径。 | plan-review remediation |
| 2026-04-27 | 1.1 | 对齐 A5 单人开发阶段决策：B3 当前只要求本地 `make codegen-events` / `make lint-events` drift 与 breaking-change gate，远端 CI 不作为 P0 前置。 | ci-pipeline-baseline spec-contract remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定 18 个内部事件 v1 全集（target / practice / report / mistake / resume / debrief / source / privacy 8 个 domain）、envelope 字段集、`outbox_events` 字段引用、dispatcher 协议（at-least-once + SKIP LOCKED）、9 项 public `jobType` ↔ Asynq dotted task name 映射；引用 `B3 event-and-outbox-contract`、`B4 db-migrations-baseline §5.9`、[ADR-Q2](../engineering-roadmap/decisions/ADR-Q2-async-orchestration.md) 与 [engineering-roadmap §3.1 D-2 jobType 命名约束](../engineering-roadmap/spec.md#32-adr-q1q6-当前约束)。 | engineering-roadmap/001 Phase 3 |
