# Backend Async Runner History

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-21

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-21 | 1.0 | 初始创建：锁定 backend in-process 单一 runtime kernel、统一 lease/retry/dead-letter/reaper/shutdown 协议、落地 B3 outbox dispatcher 缺口、保留「不建独立 worker 进程」语义；P3 接管 `email_dispatch` 从 C1 进程内 channel 迁到 `async_jobs` 行。 | 001-internal-job-outbox-runner |
