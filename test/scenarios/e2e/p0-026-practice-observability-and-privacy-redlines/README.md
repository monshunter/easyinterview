# E2E.P0.026 Practice Observability And Privacy Redlines

> **场景 ID**: E2E.P0.026
> **执行方式**: automated
> **隔离级别**: in-process (focused privacy/observability tests)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

场景使用包含敏感 marker strings 的 conversation/report 输入与输出作为负向 fixture。

## 2 When

场景检查 lifecycle-only outbox、AI observability plaintext redaction、metric label allowlist、会话级报告单次调用边界，并运行 backend-practice out-of-scope gate。

## 3 Then

消息正文不进入 outbox / logs / metrics / task metadata；metric label 使用 allowlist；报告只执行一次会话级 AI 调用；backend-practice 当前实现面不包含旧题目/turn/hint 模块入口。

## 4 执行

```bash
./test/scenarios/e2e/p0-026-practice-observability-and-privacy-redlines/scripts/setup.sh
./test/scenarios/e2e/p0-026-practice-observability-and-privacy-redlines/scripts/trigger.sh
./test/scenarios/e2e/p0-026-practice-observability-and-privacy-redlines/scripts/verify.sh
./test/scenarios/e2e/p0-026-practice-observability-and-privacy-redlines/scripts/cleanup.sh
```

## 5 污染控制

当前脚本使用 focused store/observability/review tests，不写共享数据库，并显式拒绝 `no tests to run` 假阳性。`cleanup.sh` 只清理 setup marker，保留 `.test-output/e2e/p0-026-practice-observability-and-privacy-redlines/trigger.log` 与 `result.json` 作为 BDD evidence。
