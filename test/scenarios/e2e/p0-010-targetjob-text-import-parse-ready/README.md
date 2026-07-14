# E2E.P0.010 TargetJob Paste Import Parse Ready

> **场景 ID**: E2E.P0.010
> **执行方式**: automated
> **隔离级别**: in-process (Go HTTP tests)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

已登录用户准备合法的 `{rawText,targetLanguage,resumeId}`、cookie/session 上下文与 `Idempotency-Key`。`APP_ENV=test` 使用 deterministic stub / fake AI 与 in-process runner kernel，不依赖真实 provider。

## 2 When

场景执行 paste-only `importTargetJob -> target_import parse -> listTargetJobs -> getTargetJob -> updateTargetJob`，并以同一 idempotency key 重放导入。

## 3 Then

`POST /targets/import` 返回 generated `TargetJobWithJob` 与 queued `target_import` job；幂等重放不新增 TargetJob；`rawText` 只作为 `target_jobs.raw_jd_text` 持久化。解析完成后详情返回 `analysisStatus=ready`、requirements、summary/fitSummary provenance；列表包含该 TargetJob；`PATCH /targets/{id}` 可更新合法 status / notes 且不改 `analysisStatus`。`target.import.requested` / `target.parsed` 与 job payload 不含 JD 原文、`sourceType`、`sourceUrl`、prompt 或 response，且不产生 source refresh job。

## 4 执行

```bash
./test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/scripts/setup.sh
./test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/scripts/trigger.sh
./test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/scripts/verify.sh
./test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/scripts/cleanup.sh
```

## 5 污染控制

当前脚本使用 `cmd/api` HTTP 场景测试，覆盖 auth middleware、generated route、TargetJob handler/service 与 in-process runner kernel runtime；`cleanup.sh` 只清理 setup marker，保留 `.test-output/e2e/p0-010-targetjob-text-import-parse-ready/trigger.log` 与 `result.json` 作为真实 BDD evidence。`result.json.validBddEvidence=true`。
