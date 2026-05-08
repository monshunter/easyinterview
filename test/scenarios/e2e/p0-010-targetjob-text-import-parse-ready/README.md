# E2E.P0.010 TargetJob Text Import Parse Ready

> **场景 ID**: E2E.P0.010
> **执行方式**: automated
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

已登录用户准备 `manual_text` JD、`targetLanguage`、cookie/session 上下文与 `Idempotency-Key`。`APP_ENV=test` 使用 deterministic stub / fake AI 与 in-process drainer，不依赖真实 provider。

## 2 When

场景执行 `importTargetJob -> target_import parse -> listTargetJobs -> getTargetJob -> updateTargetJob`，并覆盖同用户 idempotency dedupe 的 store gate。

## 3 Then

`POST /targets/import` 返回 generated `TargetJobWithJob` 与 queued `target_import` job；解析完成后详情返回 `analysisStatus=ready`、requirements、summary/fitSummary provenance；列表包含该 TargetJob；`PATCH /targets/{id}` 可更新合法 status / notes 且不改 `analysisStatus`；outbox 含 `target.import.requested` 与 `target.parsed`，且 payload 不泄露 JD 原文、prompt 或 response。

## 4 执行

```bash
./test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/scripts/setup.sh
./test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/scripts/trigger.sh
./test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/scripts/verify.sh
./test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/scripts/cleanup.sh
```

## 5 污染控制

当前脚本使用 `cmd/api` HTTP 场景测试，覆盖 auth middleware、generated route、TargetJob handler/service 与 in-process drainer runtime；`cleanup.sh` 只清理 setup marker，保留 `.test-output/e2e/p0-010-targetjob-text-import-parse-ready/trigger.log` 与 `result.json` 作为真实 BDD evidence。`result.json.validBddEvidence=true`。
