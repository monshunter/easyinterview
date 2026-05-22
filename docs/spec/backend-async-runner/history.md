# Backend Async Runner History

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-05-22

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-22 | 1.4 | L2 review fix：补齐 scheduler 防饥饿、handler 返回后 fresh timestamp finalize、`report_generate` 失败走 kernel shared backoff 的验收口径与测试 gate。 | 001-internal-job-outbox-runner |
| 2026-05-22 | 1.3 | L2 completion audit 修正文档事实口径：将 spec / plan 中的旧 runner、旧 mail dispatcher 与 outbox dispatcher 缺失描述明确为 `001` 实施前基线，并补写当前 kernel + `OutboxDispatcher` 已接入的完成态事实，避免 active spec 继续表达旧实现为当前事实。 | 001-internal-job-outbox-runner |
| 2026-05-22 | 1.2 | L2 code review remediation：固化 `cmd/api` production bootstrap 必须挂接 `OutboxDispatcher` 的验收口径；补齐 owner BDD rerun 清单（含 `E2E.P0.011` / `053` / `093`）与 p0-033 live gate 重复运行证据。 | 001-internal-job-outbox-runner |
| 2026-05-22 | 1.1 | 合并 `main` 后按当前 B3/shared jobs 与 backend-jobs-recommendations 实现修订：将 `jd_match_agent_scan` 纳入当前可执行 runner 接管范围，明确 `jd_match_search` 仅为 future-async reserved，补齐 operation matrix、Phase 2/4 checklist、BDD rerun 与 context discovery。 | 001-internal-job-outbox-runner |
| 2026-05-21 | 1.0 | 初始创建：锁定 backend in-process 单一 runtime kernel、统一 lease/retry/dead-letter/reaper/shutdown 协议、落地 B3 outbox dispatcher 缺口、保留「不建独立 worker 进程」语义；P3 接管 `email_dispatch` 从 C1 进程内 channel 迁到 `async_jobs` 行。 | 001-internal-job-outbox-runner |
