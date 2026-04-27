# Event and Outbox Contract History

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-04-27

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-27 | 1.0 | 初始创建：锁定 18 个内部事件 v1 全集（target / practice / report / mistake / resume / debrief / source / privacy 8 个 domain）、envelope 字段集、`outbox_events` 字段引用、dispatcher 协议（at-least-once + SKIP LOCKED）、9 项 public `jobType` ↔ Asynq dotted task name 映射；引用 [06-event-contracts.md](../../../easyinterview-tech-docs/06-event-contracts.md)、[03-db-definition.md §5.9](../../../easyinterview-tech-docs/03-db-definition.md)、[ADR-Q2](../engineering-roadmap/decisions/ADR-Q2-async-orchestration.md) 与 [engineering-roadmap §3.1 D-2 jobType 命名约束](../engineering-roadmap/spec.md#32-w0-已锁定决策hard-gate--全部-accepted)。 | engineering-roadmap/001 Phase 3 |
