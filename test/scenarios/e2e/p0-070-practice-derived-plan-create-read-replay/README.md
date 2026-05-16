# E2E.P0.070 Practice derived plan create/read/replay

> **场景 ID**: E2E.P0.070
> **自动化入口**: `cd backend && go test ./cmd/api -run TestE2EP0070PracticeDerivedPlanCreateReadReplay -count=1`

验证 `createPracticePlan` 对 `retry_current_round`、`next_round`、`debrief` 的 source 字段写入、`getPracticePlan` 读取和同 key replay 响应保持一致。
