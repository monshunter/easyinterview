# E2E.P0.012 TargetJob Parse Failure Retryable

> **场景 ID**: E2E.P0.012
> **执行方式**: automated
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

已登录用户已有 `manual_text` TargetJob，F3 / A3 fake 可注入 `AI_PROVIDER_TIMEOUT`、`AI_OUTPUT_INVALID`、`AI_PROVIDER_SECRET_MISSING` 与 F3 disabled / unsupported profile。

## 2 When

场景通过 `target_import` parse executor 分别触发 retryable 与 non-retryable 失败路径。

## 3 Then

retryable 失败写入 `target.analysis.failed.retryable=true`；non-retryable 失败写入 `retryable=false`；`target_jobs.analysis_status=failed`，source row 保留；error envelope、outbox payload、log 与 metric 证据不含 prompt body、response body、provider secret 或 `Authorization:`。

## 4 执行

```bash
./test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/scripts/setup.sh
./test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/scripts/trigger.sh
./test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/scripts/verify.sh
./test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/scripts/cleanup.sh
```

## 5 污染控制

场景使用 Go focused tests 与 fake A3/F3，不写共享数据库，不启动 Kind cluster；`cleanup.sh` 只清理 setup marker，保留 trigger log 与 result evidence。
