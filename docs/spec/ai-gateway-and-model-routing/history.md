# AI Gateway and Model Routing History

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-04-27

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-27 | 1.4 | 对齐 A5 单人开发阶段决策：A3 的测试红线当前由本地 lint / test gate 强制，远端 CI 仅在 A5 触发条件成立后再接入。 | ci-pipeline-baseline spec-contract remediation |
| 2026-04-27 | 1.3 | 明确 stub 只用于单元测试 / 离线契约测试 / 显式 mock；docker compose 与 Kind 本地部署、staging、prod 必须通过 `AI_GATEWAY_BASE_URL` + `AI_GATEWAY_API_KEY` 指向真实 AI provider 或生产 gateway，缺失时 fail-fast，不默认降级到 stub。 | local-dev-stack/001-bootstrap review remediation |
| 2026-04-27 | 1.2 | 对齐 A2 local-dev-stack v1.2：本地开发与单元测试默认走应用内 deterministic stub provider；A2 不再预留或启动 `ai-gateway-mock` compose 服务，AI gateway 不作为本地开发依赖。 | local-dev-stack/001-bootstrap review remediation |
| 2026-04-27 | 1.1 | 对齐 F1 baseline metric label contract：fallback metric 使用 `from_model_family` / `to_model_family` / `result`，不把原始 model id 写入 Prometheus label | engineering-roadmap/001 Phase 3 remediation |
| 2026-04-27 | 1.0 | 初始创建：把 [ADR-Q6](../engineering-roadmap/decisions/ADR-Q6-ai-gateway-and-model-routing.md) 的 9 项硬约束落到 `AIClient` 接口、Model Profile schema、stub / openai_compatible provider、观测埋点、隐私红线、fallback 边界；引用 [01-technical-architecture.md §10](../../../easyinterview-tech-docs/01-technical-architecture.md#10-ai-编排层设计)、[04-metrics-observability.md §8](../../../easyinterview-tech-docs/04-metrics-observability.md#8-ai-调用指标)、[05-logging-standard.md §4.4](../../../easyinterview-tech-docs/05-logging-standard.md#44-ai-log-额外字段)、[03-db-definition.md §5.8](../../../easyinterview-tech-docs/03-db-definition.md) 中的 `ai_task_runs` schema。 | engineering-roadmap/001 Phase 3 |
