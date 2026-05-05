# Observability Stack History

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-05-05

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-05 | 1.5 | F1 metric / logging / trace / dashboard / alerting 契约改为只由本 spec 与后续编码 truth source 承接；移除旧技术草稿名称和旧 shorthand 依赖。 | engineering-roadmap/001-decompose-subspecs |
| 2026-05-03 | 1.4 | 明确当前观测契约由本 spec、F1 后续编码 truth source 与 product-scope 当前范围决定，并从业务域 metric 前缀中移除独立 `mistake_` / `growth_`。 | docs-only |
| 2026-05-03 | 1.3 | 对齐 product-scope v1.2：metric 命名示例移除旧 `mistake_entries` 表语义，改为报告生成相关示例；不恢复独立错题本 / 成长中心指标口径。 | docs consistency remediation |
| 2026-04-27 | 1.2 | 对齐 A2 local-dev-stack v1.2：F1 消费默认本地环境提供的应用 `/metrics` 与容器日志；OTel Collector / Grafana / Loki / Prometheus 仅作为 F1/E4 可选观测或生产部署路径，不再要求 A2 默认 `make dev-up` 提供。 | local-dev-stack/001-bootstrap review remediation |
| 2026-04-27 | 1.1 | 修正 baseline metric label contract：将 `operation`、`from_model_family`、`to_model_family` 纳入有界 allowed labels；禁止原始 model id / `from_model` / `to_model` 进入 metric；`ai_fallback_total` 改用 model family label，避免 baseline metric 与 lint 规则互相冲突 | engineering-roadmap/001 Phase 3 remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定 Prometheus metric 命名 / allowed labels / forbidden labels / 明文红线 / OTel pipeline 框架 / 5 dashboard 名称 / Sentry 接线 / 健康检查端点；§3.1.1 W1 freeze 24 个 baseline metric（HTTP / DB / Async / Outbox / AI），承接 [engineering-roadmap §5.7 W1 baseline 指标命名约定锁定](../engineering-roadmap/spec.md#6-实施顺序) spec-contract lock；引用 `F1 observability-stack`、`F1 observability-stack logging`、[ADR-Q5](../engineering-roadmap/decisions/ADR-Q5-privacy-cadence.md) 与 [A4 env 字典](../secrets-and-config/spec.md#311-p0-必备-env-key-字典25-项)。 | engineering-roadmap/001 Phase 3 |
