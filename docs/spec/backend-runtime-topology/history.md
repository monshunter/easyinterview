# Backend Runtime Topology History

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-05-26

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-26 | 1.5 | 对齐 local-dev-stack Mailpit revision：默认本地依赖增加 Mailpit 本地邮箱 sink，仍不新增 worker 或观测消费端。 | local-dev-stack/001 Mailpit revision |
| 2026-05-07 | 1.4 | L2 remediation：补强 `make lint-runtime-topology` 对跨行 YAML / JSON producer `worker` 字段和 owner plan/checklist current handoff 旧 worker 口径的 false-negative 覆盖。 | 001-worker-consolidation |
| 2026-05-07 | 1.3 | L2 remediation：补强 `make lint-runtime-topology` 对 `scripts/` 工具面和 raw `producer: worker` / `"producer": "worker"` 合约形态的 false-negative 覆盖。 | 001-worker-consolidation |
| 2026-05-06 | 1.2 | L2 remediation：补强 `make lint-runtime-topology` false-negative 覆盖，拦截 active handoff 中旧 producer、`app/worker listen addr` 与 `backend-async-runtime` shorthand 回流。 | 001-worker-consolidation |
| 2026-05-06 | 1.1 | L2 remediation：新增 `make lint-runtime-topology` 负向 gate，覆盖 active code/doc handoff 中 non-current standalone worker process 口径回流。 | 001-worker-consolidation |
| 2026-05-06 | 1.0 | 初始创建：锁定 P0 取消独立 worker 进程、后台任务收敛到 backend internal runner、开发期观测消费端不作为 gate。 | 001-worker-consolidation |
