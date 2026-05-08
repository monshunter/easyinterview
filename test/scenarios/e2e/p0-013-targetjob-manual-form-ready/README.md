# E2E.P0.013 TargetJob Manual Form Ready

> **场景 ID**: E2E.P0.013
> **执行方式**: automated
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

已登录用户填写 manual form：title、companyName、rawDescription、`targetLanguage` 与 `Idempotency-Key`。A3 / F3 fake 可不可用，本路径不依赖异步 AI parse。

## 2 When

用户调用 `importTargetJob` 的 `manual_form` source，随后读取详情与列表，并执行 repeated idempotency store gate。

## 3 Then

响应保持 B2 `TargetJobWithJob` wire shape，`job.jobType=target_import` 且 `job.status=succeeded`；TargetJob 立即 `analysisStatus=ready` 并拥有至少 1 条 `must_have` draft requirement；不创建待 drainer 消费的 `target_import` async job，不发 `target.import.requested` / `target.parsed`；证据不含 raw JD text、prompt 或 response。

## 4 执行

```bash
./test/scenarios/e2e/p0-013-targetjob-manual-form-ready/scripts/setup.sh
./test/scenarios/e2e/p0-013-targetjob-manual-form-ready/scripts/trigger.sh
./test/scenarios/e2e/p0-013-targetjob-manual-form-ready/scripts/verify.sh
./test/scenarios/e2e/p0-013-targetjob-manual-form-ready/scripts/cleanup.sh
```

## 5 污染控制

场景使用 Go focused tests、fake store 与 sqlmock，不写共享数据库，不启动 Kind cluster；`cleanup.sh` 只清理 setup marker，保留 trigger log 与 result evidence。
