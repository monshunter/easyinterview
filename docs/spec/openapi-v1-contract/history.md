# OpenAPI v1 Contract History

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-04-27

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-27 | 1.1 | 修正 W1 gate 口径：parent Phase 3 只锁定 B2 v1.0.0 freeze 范围与 additive-only 规则；真实 `openapi/openapi.yaml`、codegen、fixtures 与 breaking-change linter 由 B2 child `001` 系列 plan 验证后再放行依赖 B2 的 W2 implementation | engineering-roadmap/001 Phase 3 remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定 `openapi/openapi.yaml` 唯一真理源、`/api/v1` 路径前缀、`camelCase` 字段命名、RFC3339 时间格式、共享 `ApiError` schema、cursor 分页统一、`Idempotency-Key` 与 `Job 202` 异步契约；§3.1.1 列出 v1.0.0 freeze 时的 36 个 endpoint × 14 tag 完整集合；锁定 breaking change linter 规则集（仅允许 additive）；引用 [02-api-definition.md](../../../easyinterview-tech-docs/02-api-definition.md) 全文与 [B1 D-5/D-6 枚举与错误码](../shared-conventions-codified/spec.md#31-已锁定决策)；记录 [ADR-Q5](../engineering-roadmap/decisions/ADR-Q5-privacy-cadence.md) `POST /privacy/exports` P0 返回 501 的例外。 | engineering-roadmap/001 Phase 3 |
