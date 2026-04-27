# AI Gateway and Model Routing History

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-04-27

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-27 | 1.0 | 初始创建：把 [ADR-Q6](../engineering-roadmap/decisions/ADR-Q6-ai-gateway-and-model-routing.md) 的 9 项硬约束落到 `AIClient` 接口、Model Profile schema、stub / openai_compatible provider、观测埋点、隐私红线、fallback 边界；引用 [01-technical-architecture.md §10](../../../easyinterview-tech-docs/01-technical-architecture.md#10-ai-编排层设计)、[04-metrics-observability.md §8](../../../easyinterview-tech-docs/04-metrics-observability.md#8-ai-调用指标)、[05-logging-standard.md §4.4](../../../easyinterview-tech-docs/05-logging-standard.md#44-ai-log-额外字段)、[03-db-definition.md §5.8](../../../easyinterview-tech-docs/03-db-definition.md) 中的 `ai_task_runs` schema。 | engineering-roadmap/001 Phase 3 |
