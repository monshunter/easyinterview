# E2E.P0.011 TargetJob URL Import Fetch And Parse

> **场景 ID**: E2E.P0.011
> **执行方式**: automated
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

已登录用户准备合规 HTTPS JD URL、非法 URL 集合和 deterministic fake AI。URL fetcher 通过 repo-tracked focused tests 覆盖 public IP、私网、metadata、超长、空白和 4xx/5xx。

## 2 When

场景执行 `url import -> drainer fetch -> parse -> source snapshot persist`，并对非法 URL 执行拒绝路径。

## 3 Then

合法 URL 被规范化并剥离 query / fragment / userinfo；抓取正文写入 `target_job_sources.snapshot_text` 并驱动 parse；`target.parsed` 与 internal-only `source_refresh` 占位 job 写入；非法 URL 映射 `TARGET_IMPORT_SOURCE_INVALID` 或 `TARGET_IMPORT_SOURCE_UNAVAILABLE`；事件 / 日志 / metric 证据不包含完整 URL query、内网响应或 prompt。

## 4 执行

```bash
./test/scenarios/e2e/p0-011-targetjob-url-import-fetch-and-parse/scripts/setup.sh
./test/scenarios/e2e/p0-011-targetjob-url-import-fetch-and-parse/scripts/trigger.sh
./test/scenarios/e2e/p0-011-targetjob-url-import-fetch-and-parse/scripts/verify.sh
./test/scenarios/e2e/p0-011-targetjob-url-import-fetch-and-parse/scripts/cleanup.sh
```

## 5 污染控制

当前脚本使用 `cmd/api` HTTP 场景测试，覆盖 auth middleware、generated route、TargetJob handler/service、in-process drainer runtime 与 URL fetch boundary；`cleanup.sh` 只清理 setup marker，保留 trigger log 与 result evidence。`result.json.validBddEvidence=true`。
