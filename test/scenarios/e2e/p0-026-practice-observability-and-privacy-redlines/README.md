# E2E.P0.026 Practice Observability And Privacy Redlines

> **场景 ID**: E2E.P0.026
> **执行方式**: automated
> **隔离级别**: in-process (Go HTTP tests)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

用户 A 拥有一个 ready baseline plan，并通过 observed AI client 启动面试。场景使用包含敏感 marker strings 的 AI 输入与输出作为负向 fixture。

## 2 When

场景触发 `startPracticeSession`，采集 in-process store 中的 audit / outbox 快照、A3 observed AI logs / metrics / audit rows / `ai_task_runs` 行，并运行 backend-practice out-of-scope gate。

## 3 Then

用户获得 running session 与同步首题；observability typed columns 完整；metric label 使用 A3 allowlist；audit / outbox / logs / metrics / AI task run metadata 只保留允许的证据摘要；backend-practice 当前实现面不包含 out-of-scope 模块入口。

## 4 执行

```bash
./test/scenarios/e2e/p0-026-practice-observability-and-privacy-redlines/scripts/setup.sh
./test/scenarios/e2e/p0-026-practice-observability-and-privacy-redlines/scripts/trigger.sh
./test/scenarios/e2e/p0-026-practice-observability-and-privacy-redlines/scripts/verify.sh
./test/scenarios/e2e/p0-026-practice-observability-and-privacy-redlines/scripts/cleanup.sh
```

## 5 污染控制

当前脚本使用 `cmd/api` HTTP 场景测试与 in-process collectors，不写共享数据库。`cleanup.sh` 只清理 setup marker，保留 `.test-output/e2e/p0-026-practice-observability-and-privacy-redlines/trigger.log` 与 `result.json` 作为 BDD evidence。
