# Backend Auth History

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-05-26

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-26 | 1.5 | BUG-0106 修订：对齐 ADR-Q5 与 B4 privacy deletion matrix，明确 `DELETE /api/v1/me` 受理期由 C1 同步软删 `users.deleted_at` / `users.status='deleted'` 并撤销该用户所有 session，backend internal runner 完成逐域删除后负责用户行最终 hard delete。 | backend-async-runner/001-internal-job-outbox-runner BUG-0106 remediation |
| 2026-05-26 | 1.4 | 登记 local-dev-stack Mailpit revision：`EMAIL_PROVIDER=mailpit` 时 `cmd/api` 使用 SMTP `DeliveryWriter` 投递 magic-link 到 Mailpit；writer 从 `auth_challenges` 查询收件人、从 transient delivery secret store 取 token，保持 `email_dispatch` payload redaction，不引入独立后台进程或场景专属 `backend/cmd` helper。 | local-dev-stack/001 Mailpit revision |
| 2026-05-21 | 1.3 | 登记 backend-jobs-recommendations/001 cross-owner additive：新增 `GetUserIdentityForUser(ctx, db, userID) (UserIdentity, error)` 内部 API（`backend/internal/auth/identity.go`），返回 `UserIdentity{DisplayName, AvatarURL, EmailMasked}`；read-only `SELECT email, display_name FROM users WHERE id = $1 AND deleted_at IS NULL`；emailMasked 通过既有 `maskEmail` helper (first + *** + last char of local part + domain) 派生，不返回 raw email；不写 audit_events，不改 user/session 状态；missing user 返回 `ErrUserNotFound`，caller 应 fallback 到非 PII anonymous display name `Candidate`。模块边界表追加 cross-owner internal API 行；新增 spec D-X 决策内容由本 history 行替代（保持 §3.1 D-1..D-N 表稳定）。单元测试 `identity_test.go` 覆盖 seeded happy / missing display_name fallback / ErrUserNotFound / does-not-write-audit / nil-db / empty-userId 6 项断言，emailMasked 显式断言含 `***`、不含 raw local-part `alice`、domain 保留。 | backend-jobs-recommendations/001-jd-match-real-backend-baseline Phase 0.19 |
| 2026-05-06 | 1.2 | 对齐 backend-runtime-topology：C1 backend-internal mail dispatcher 从过渡 workaround 升为 P0 默认实现，不再引用独立 C8 worker 进程作为未来前置。 | backend-runtime-topology/001-worker-consolidation |
| 2026-05-06 | 1.1 | L1 plan-review remediation：补齐 C1 backend-internal mail dispatcher 过渡实现、B3 `email_dispatch` redaction、B2 generated `deleteMe` idempotent auth handoff、session middleware / logout optional-session、ADR-Q1 rate-limit / TTL / renewal 边界、F1 auth metrics registry gate 与 BDD 错误路径。 | 001-passwordless-session-bootstrap |
| 2026-05-05 | 1.0 | 从 engineering-roadmap S2 与 ADR-Q1 派生 backend auth subject，锁定 passwordless challenge、first-party session、/me、logout 和 runtime-config session resolver 边界。 | 001-passwordless-session-bootstrap |
