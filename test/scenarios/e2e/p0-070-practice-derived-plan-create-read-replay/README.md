# E2E.P0.070 Practice derived plan create/read/replay

> **场景 ID**: E2E.P0.070
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

验证 `createPracticePlan` 对 `retry_current_round`、`next_round`、`debrief` 的 source 字段写入、`getPracticePlan` 读取和同 key replay 响应保持一致。
