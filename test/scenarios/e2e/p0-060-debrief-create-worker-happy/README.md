# E2E.P0.060 Debrief Create + Worker Generation Happy

> **场景 ID**: E2E.P0.060
> **关联需求**: backend-debrief C-1, C-2, C-3, C-5
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

## Given

用户已有目标岗位，并提交真实面试复盘问题、答案摘要和面试官反应。

## When

用户调用 createDebrief，backend 创建 draft debrief、queued `debrief_generate` job，drainer worker 执行 F3/A3 生成分析。

## Then

debrief 从 draft 进入 completed，写入 `debrief.created` / `debrief.completed` outbox，`ai_task_runs.task_type='debrief_generate'` 成功，payload 不泄漏原始问题和答案文本。
