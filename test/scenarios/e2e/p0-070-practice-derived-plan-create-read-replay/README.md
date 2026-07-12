# E2E.P0.070 Practice derived plan create/read/replay

> **场景 ID**: E2E.P0.070
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

验证真实 PostgreSQL 中 `retry_current_round` 保留 source report 的 exact round pair，`next_round` 选择 source 后的紧邻 canonical round 且必须等于当前第一个未完成轮；覆盖等时长与 sequence 跳号，不用 duration 或 `sourceSequence+1` 猜测。
