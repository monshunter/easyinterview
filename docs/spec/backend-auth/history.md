# Backend Auth History

> **版本**: 2.7
> **状态**: active
> **更新日期**: 2026-07-16

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-16 | 2.7 | 修复中文 SMTP 邮件的 RFC 2047 Subject 与 MIME transfer encoding，确保标准 relay/client 可无损解析 UTF-8 text/plain 和 text/html。 | 001-email-code-session-bootstrap Phase 13 remediation |
| 2026-07-16 | 2.6 | 修复生产 SMTP 会话取消/超时与 DATA 已接受后的重复投递风险，并保证 Redis secret 写入失败不留下会污染限流的 challenge。 | 001-email-code-session-bootstrap Phase 12 remediation |
| 2026-07-16 | 2.5 | 复用现有 Redis 保存加密且带 5 分钟 TTL 的 6 位验证码 delivery secret，支持跨 backend 实例消费。 | 001-email-code-session-bootstrap Phase 12 |
| 2026-07-16 | 2.4 | 支持通过环境变量选择 Mailpit 或带 TLS/auth 的标准 SMTP，删除未消费的 provider API key 合同。 | 001-email-code-session-bootstrap Phase 11 |
| 2026-07-10 | 2.1 | 技术债口径清理：将 auth 邮件派发 owner 从旧 goroutine / 后台派发器表述收敛为当前 `EmailDispatchEnqueuer` + `email_dispatch` async job + backend-async-runner `EmailDispatchHandler` 事实。 | 001-email-code-session-bootstrap |
| 2026-07-07 | 2.0 | 当前 backend-auth 统一为 email-code session bootstrap：`EmailCodeService` 承接邮箱验证码、first-party session、资料补全、logout、`DELETE /me` handoff 与 `email_dispatch` code-only redaction contract；P0.003 / P0.101 场景目录和 active owner plan 均使用当前命名。 | 001-email-code-session-bootstrap; product-scope/001-core-loop-module-pruning |
