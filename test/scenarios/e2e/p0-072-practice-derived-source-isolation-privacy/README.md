# E2E.P0.072 Practice derived source isolation privacy

> **场景 ID**: E2E.P0.072
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

验证 missing/cross-user/wrong-target/legacy-null source，以及 stale successor、round/budget mismatch、全部轮次完成均 fail closed，不插入 plan、不泄露 source id；真实 PostgreSQL 证据优先于 sqlmock。
