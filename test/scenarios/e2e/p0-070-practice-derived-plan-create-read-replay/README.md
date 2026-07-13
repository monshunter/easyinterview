# E2E.P0.070 Practice derived plan create/read/replay

> **场景 ID**: E2E.P0.070
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

验证真实 PostgreSQL v19 中 report-derived 当前契约：请求只有 `goal + sourceReportId`；empty retry 为通用同轮复练，issue-backed retry 原子投影 dimension codes；start/send 只得到 code + label + issue 的 untrusted semantic focus，并由 F3 已激活的 `practice.session.chat` v0.2 pair 消费；`next_round` 选择 frozen canonical successor、使用 successor duration 且 focus 为空。另运行 create-plan Idempotency-Key 指纹错配 gate，证明相同 key 不会触发第二次插入或泄漏 source。
