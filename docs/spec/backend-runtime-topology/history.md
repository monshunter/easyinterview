# Backend Runtime Topology History

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-06

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-06 | 1.1 | L2 remediation：新增 `make lint-runtime-topology` 负向 gate，覆盖 active code/doc handoff 中 retired standalone worker process 口径回流。 | 001-worker-consolidation |
| 2026-05-06 | 1.0 | 初始创建：锁定 P0 取消独立 worker 进程、后台任务收敛到 backend internal runner、开发期观测消费端不作为 gate。 | 001-worker-consolidation |
