# E2E.P0.072 Practice derived source isolation privacy

> **场景 ID**: E2E.P0.072
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

验证真实 PostgreSQL v19 中 missing / cross-user / non-ready / missing-context source，以及 frozen target / resume / round / persona / language / budget mismatch、unsupported / duplicate non-empty focus 全部 fail closed；每个失败都检查零 `practice_plans` 插入。客户端复制 server-owned identity/settings 也在 service 前置 gate 被拒绝，并执行最终 active runtime/generated/OpenAPI/JSON fixture/scenario legacy identifier 零正向命中 gate。
