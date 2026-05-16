# E2E.P0.062 Worker AI Failure Graceful + Retry + Permanent Fail

> **场景 ID**: E2E.P0.062
> **关联需求**: backend-debrief C-11, C-12
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

## Given

`debrief_generate` worker 遇到 F3/A3/parse failure，并且 async job 可能低于或达到最大 attempts。

## When

drainer 调用 handler 并 finalize job outcome。

## Then

低于最大 attempts 的失败保持 retryable/backoff；达到最大 attempts 的失败转为 permanent failure；debrief 保持 draft，不发 `debrief.completed`。
