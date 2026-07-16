# Backend Auth History

> **版本**: 2.2
> **状态**: completed
> **更新日期**: 2026-07-10

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-16 | 2.4 | 支持通过环境变量选择 Mailpit 或带 TLS/auth 的标准 SMTP，删除未消费的 provider API key 合同。 | 001-email-code-session-bootstrap Phase 11 |
| 2026-07-10 | 2.1 | 技术债口径清理：将 auth 邮件派发 owner 从旧 goroutine / 后台派发器表述收敛为当前 `EmailDispatchEnqueuer` + `email_dispatch` async job + backend-async-runner `EmailDispatchHandler` 事实。 | 001-email-code-session-bootstrap |
| 2026-07-07 | 2.0 | 当前 backend-auth 统一为 email-code session bootstrap：`EmailCodeService` 承接邮箱验证码、first-party session、资料补全、logout、`DELETE /me` handoff 与 `email_dispatch` code-only redaction contract；P0.003 / P0.101 场景目录和 active owner plan 均使用当前命名。 | 001-email-code-session-bootstrap; product-scope/001-core-loop-module-pruning |
