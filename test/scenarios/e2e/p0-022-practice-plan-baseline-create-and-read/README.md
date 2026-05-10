# E2E.P0.022 Practice Plan Baseline Create And Read

> **场景 ID**: E2E.P0.022
> **执行方式**: automated
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

已登录用户 A 拥有 `target_job_id`、`resume_asset_id`、`Idempotency-Key` 与 baseline practice plan 输入；用户 B 已登录但与该 plan 无关。`APP_ENV=test` 使用 `cmd/api` HTTP 场景测试和 in-process store snapshot，不依赖真实 provider。

## 2 When

用户 A 执行 `POST /api/v1/practice/plans` 创建 baseline plan，然后用同一 key 重放一次；用户 A `GET /api/v1/practice/plans/{planId}`；用户 B 访问同一个 `planId`。

## 3 Then

用户 A 收到 `201 + PracticePlan{status:'ready', goal:'baseline'}`；store snapshot 只写入 1 条 practice plan 与 1 条 audit metadata；同 key 重放返回首次响应且不重复副作用；用户 A `GET` 返回完整 plan；用户 B `GET` 返回 `404 + PRACTICE_PLAN_NOT_FOUND`，不泄露存在性；audit / evidence 不包含 question / answer / hint / prompt / response 明文或 retired PracticeMode literal。

## 4 执行

```bash
./test/scenarios/e2e/p0-022-practice-plan-baseline-create-and-read/scripts/setup.sh
./test/scenarios/e2e/p0-022-practice-plan-baseline-create-and-read/scripts/trigger.sh
./test/scenarios/e2e/p0-022-practice-plan-baseline-create-and-read/scripts/verify.sh
./test/scenarios/e2e/p0-022-practice-plan-baseline-create-and-read/scripts/cleanup.sh
```

## 5 污染控制

当前脚本使用 `cmd/api` HTTP 场景测试，覆盖 auth middleware、generated route、practice handler/service、shared idempotency middleware 和 in-process persistence snapshot。`cleanup.sh` 只清理 setup marker，保留 `.test-output/e2e/p0-022-practice-plan-baseline-create-and-read/trigger.log` 与 `result.json` 作为 BDD evidence。`result.json.validBddEvidence=true`。
