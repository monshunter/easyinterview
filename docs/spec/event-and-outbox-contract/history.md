# Event and Outbox Contract History

> **版本**: 1.6
> **状态**: active
> **更新日期**: 2026-05-03

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-03 | 1.6 | 将 `easyinterview-tech-docs/06` / `03` / `04` 降级为历史事件、DB 与告警输入；当前 event / job / outbox 契约改由本 spec、`shared/events.yaml`、`shared/jobs.yaml` 与 B4 migrations 决定，旧 `mistake.*`、18 event inventory 与旧 consumer 不再作为实现依据。 | docs-only |
| 2026-05-03 | 1.5 | 对齐 product-scope v1.2：删除独立 `mistake` 事件 domain 与 `mistake.created` / `mistake.status.changed`，将 `report.generated.mistakeCount` 改为 `questionIssueCount`，将 `debrief.completed.generatedMistakeCount` 改为 `practiceFocusCount`，事件全集 18→16、domain 8→7。 | 001-bootstrap Phase 8 remediation |
| 2026-04-29 | 1.4 | 物化 B3 `001-bootstrap` active plan：新增 committed JSON Schema / job manifest baseline，明确 `email_dispatch` internal-only 口径由 ADR-Q1 同步，JSON Schema refs 由 B3 自有桥接或 inline 值承接，不依赖 B1 必须产出 JSON Schema fragment。 | [001-bootstrap](./plans/001-bootstrap/plan.md) |
| 2026-04-29 | 1.3 | 收口 A/B spec 全面审查 remediation：新增 internal-only `email_dispatch` canonical jobType 与 `email.dispatch` dotted task name，锁定 magic link payload 红线；把 `target.analysis.failed` / `report.generation.failed` / `mistake.status.changed` 改为真正 dot.case，避免 lint 规则拒绝 seed 事件；同步 C9 真实面试复现 P0 范围。 | plan-review remediation |
| 2026-04-29 | 1.2 | 根据 L1 review findings 修订 B3 契约：补齐 18 个事件 v1 payload schema inventory 与 PII 边界；拆分 DB/C8 canonical `job_type` 与 B2 API-facing `JobType` subset；明确 outbox retry operational columns、`traceId` soft-required 语义、B3-owned `codegen-events` 归属与 Go/TS 输出路径。 | plan-review remediation |
| 2026-04-27 | 1.1 | 对齐 A5 单人开发阶段决策：B3 当前只要求本地 `make codegen-events` / `make lint-events` drift 与 breaking-change gate，远端 CI 不作为 P0 前置。 | ci-pipeline-baseline spec-contract remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定 18 个内部事件 v1 全集（target / practice / report / mistake / resume / debrief / source / privacy 8 个 domain）、envelope 字段集、`outbox_events` 字段引用、dispatcher 协议（at-least-once + SKIP LOCKED）、9 项 public `jobType` ↔ Asynq dotted task name 映射；引用 [06-event-contracts.md](../../../easyinterview-tech-docs/06-event-contracts.md)、[03-db-definition.md §5.9](../../../easyinterview-tech-docs/03-db-definition.md)、[ADR-Q2](../engineering-roadmap/decisions/ADR-Q2-async-orchestration.md) 与 [engineering-roadmap §3.1 D-2 jobType 命名约束](../engineering-roadmap/spec.md#32-w0-已锁定决策hard-gate--全部-accepted)。 | engineering-roadmap/001 Phase 3 |
