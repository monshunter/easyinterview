# Backend Async Runner History

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-22

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-22 | 1.1 | 合并 `main` 后按当前 B3/shared jobs 与 backend-jobs-recommendations 实现修订：将 `jd_match_agent_scan` 纳入当前可执行 runner 接管范围，明确 `jd_match_search` 仅为 future-async reserved，补齐 operation matrix、Phase 2/4 checklist、BDD rerun 与 context discovery。 | 001-internal-job-outbox-runner |
| 2026-05-21 | 1.0 | 初始创建：锁定 backend in-process 单一 runtime kernel、统一 lease/retry/dead-letter/reaper/shutdown 协议、落地 B3 outbox dispatcher 缺口、保留「不建独立 worker 进程」语义；P3 接管 `email_dispatch` 从 C1 进程内 channel 迁到 `async_jobs` 行。 | 001-internal-job-outbox-runner |
