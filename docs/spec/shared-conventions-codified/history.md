# Shared Conventions Codified History

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-04-29

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-29 | 1.5 | 按 ADR-Q6 authoritative 边界补齐 AI shared vocabulary：B1 只拥有 `AI_*` 错误码与 Model Profile / AI meta 字段名常量或生成类型；A3 继续拥有 Model Profile schema、`AIClient` runtime、`AICallMeta` runtime 与 provider adapter，A4/E4 负责连接参数与 endpoint。 | plan-review remediation |
| 2026-04-29 | 1.4 | 授权并落地 A3 AI gateway baseline 错误码：`AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_FALLBACK_EXHAUSTED`，作为 `shared/conventions.yaml` 与 Go / TS / OpenAPI codegen 共同消费的唯一真理源；`AICallMeta` 运行时结构仍由 A3 拥有，不进入 B1 共享 DTO。 | ai-gateway-and-model-routing spec remediation |
| 2026-04-28 | 1.3 | 明确 `ApiError` 为错误响应 envelope 内部对象，Go canonical 类型继续归属 `backend/internal/shared/errors.APIError`，B2 OpenAPI 负责外层 `ApiErrorResponse` envelope。 | openapi-v1-contract/001-bootstrap assessment remediation |
| 2026-04-27 | 1.2 | 对齐 A5 单人开发阶段决策：B1 只要求本地 lint/codegen 质量门禁，远端 CI / PR required check / CI drift detection 不作为当前 P0 前置，未来触发条件成立后再由 A5 重新评估。 | 001-bootstrap |
| 2026-04-27 | 1.1 | 回写 001-bootstrap 复盘确认的文档漂移：明确 13 个上游枚举小节 / 14 个生成枚举类型、Go `APIError` 手写归属、TS toolchain 与 Go/TS idempotency 双端验收落点。 | 001-bootstrap |
| 2026-04-26 | 1.0 | 初始创建：锁定跨语言真理源 `shared/conventions.yaml`、Go module 名称（`github.com/monshunter/easyinterview/backend`）、pnpm workspace、UUIDv7 / tmp_ id 规则、错误码 `UPPER_SNAKE_CASE` lint、枚举 `lower_snake_case` 双向生成；引用 [00-shared-conventions.md](../../../easyinterview-tech-docs/00-shared-conventions.md) 13 个上游枚举小节与 6 个已记录错误码示例 | 001-bootstrap |
