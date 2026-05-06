# Backend Auth History

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-05-06

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-06 | 1.2 | 对齐 backend-runtime-topology：C1 backend-internal mail dispatcher 从过渡 workaround 升为 P0 默认实现，不再引用独立 C8 worker 进程作为未来前置。 | backend-runtime-topology/001-worker-consolidation |
| 2026-05-06 | 1.1 | L1 plan-review remediation：补齐 C1 backend-internal mail dispatcher 过渡实现、B3 `email_dispatch` redaction、B2 generated `deleteMe` idempotent auth handoff、session middleware / logout optional-session、ADR-Q1 rate-limit / TTL / renewal 边界、F1 auth metrics registry gate 与 BDD 错误路径。 | 001-passwordless-session-bootstrap |
| 2026-05-05 | 1.0 | 从 engineering-roadmap S2 与 ADR-Q1 派生 backend auth subject，锁定 passwordless challenge、first-party session、/me、logout 和 runtime-config session resolver 边界。 | 001-passwordless-session-bootstrap |
