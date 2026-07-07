# Backend Auth History

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-07

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-07 | 2.0 | 当前 backend-auth 统一为 email-code session bootstrap：`EmailCodeService` 承接邮箱验证码、first-party session、资料补全、logout、`DELETE /me` handoff 与 `email_dispatch` code-only redaction contract；P0.003 / P0.101 场景目录和 active owner plan 均使用当前命名。 | 001-email-code-session-bootstrap; product-scope/001-core-loop-module-pruning |
