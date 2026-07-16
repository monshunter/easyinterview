# Backend Auth BDD Checklist

> **版本**: 1.14
> **状态**: completed
> **更新日期**: 2026-07-16

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.AUTH.EMAIL.001` Email-code session 与资料补全

- [x] Owner behavior tests 覆盖 challenge/verify、四字段 `/me`、session、profile completion、logout/relogin、重放与隐私。
- [x] 根 `make test` 执行对应 Go tests；该结果是代码层行为证据，不是 E2E PASS。

## `E2E.P0.101` 静态 handoff

- [x] 静态资产原地扩展真实设置齿轮、姓名/完整邮箱与 logout，且证据不保存完整邮箱；不另建场景。
- [x] 不使用 fixture transport、进程内 handler 或代码层测试冒充 E2E；账号删除不进入该共享场景。

以上 owner 合同由 `e2e-scenarios-p0/001` 维护真实运行状态；当前 run `e2e-p0-101-20260715114513-19516` 已在真实 frontend/backend/Mailpit 环境 PASS，静态 INDEX 状态仍为 `Ready`。

## `BDD.AUTH.EMAIL.002` Mailpit / production SMTP provider

- [x] Owner behavior tests 覆盖 Mailpit plain/no-auth、SMTP STARTTLS/implicit TLS + AUTH、provider fail-fast、transport failure 与隐私红线。
- [x] 根 `make test` 执行对应 Go tests；该结果是代码层行为证据，不是外部邮箱 E2E PASS。
  <!-- verified: 2026-07-16 method=domain-behavior evidence="auth/config/cmd-api behavior tests plus root make test pass; external SMTP receipt remains a separate live gate" -->
- [x] Owner behavior test 使用标准 MIME reader 解码 `zh-CN` Subject、text/plain 与 text/html，并验证中文标题、说明和验证码无损保留。
  <!-- verified: 2026-07-16 method=domain-behavior evidence="TestSMTPDeliveryWriterEncodesLocalizedMessageAsStandardsCompliantMIME parses the captured SMTP message with net/mail and mime/multipart, decodes RFC 2047 plus quoted-printable, and asserts both body alternatives preserve the localized title, instruction and code." -->
- [x] 重新运行 focused auth test 与根 `make test`，记录本次 domain behavior 证据；不把代码 gate 包装成 E2E。
  <!-- verified: 2026-07-16 method=domain-behavior evidence="Focused localized MIME test and go test ./internal/auth PASS; root make test PASS with Python 567 tests/4481 subtests, Go all packages and frontend 126 files/1004 tests. No scenario asset or E2E status changed." -->

## `BDD.AUTH.EMAIL.003` Cross-instance Redis delivery secret

- [x] Owner domain behavior tests 覆盖实例 A Put/实例 B Get/Delete、加密 value、5 分钟 TTL、SMTP success/failure lifecycle 与脱敏错误。
  <!-- verified: 2026-07-16 method=domain-behavior evidence="TestEmailCodeDeliveryWorksAcrossIndependentRedisBackedInstances plus Redis store and delivery lifecycle focused tests pass" -->
- [x] 使用两个独立真实 Redis client 执行 cross-client integration；该结果属于代码 integration evidence，不是 E2E。
  <!-- verified: 2026-07-16 method=integration evidence="REDIS_URL=redis://127.0.0.1:6379/0 go test -tags=integration ./internal/auth -run TestRedisDeliverySecretStoreCrossClientIntegration -count=1 PASS; covers encrypted raw value, near-5m TTL, cross-client Get/Delete and actual expiry" -->
- [x] 根 `make test` 回归对应 domain behavior tests，并记录 full-container Mailpit/SMTP live evidence。
  <!-- verified: 2026-07-16 evidence="root make test includes auth/cmd-api domain behavior and integration contract tests; full-container Mailpit Chrome challenge->receive->verify->profile PASS with consoleIssues=0; Redis namespace empty after success; external SMTP job succeeded attempts=1" -->
- [x] L2 remediation domain behavior tests 覆盖 runner cancel/SMTP deadline、DATA accepted 后 QUIT 失败仅成功一次、Redis Put 失败无 challenge/rate-limit 污染。
  <!-- verified: 2026-07-16 evidence="runner context propagation, stalled SMTP cancellation, accepted DATA/failed QUIT success, Redis Put failure and canceled-request compensation tests PASS" -->
- [x] 重新运行 owner focused tests 与根 `make test`，记录本次行为证据；不把代码 gate 包装成 E2E。
  <!-- verified: 2026-07-16 evidence="internal/auth and real Redis cross-client integration PASS; root make test PASS with Python 566 tests/4481 subtests, Go all packages and frontend 1004 tests; no E2E status changed" -->
