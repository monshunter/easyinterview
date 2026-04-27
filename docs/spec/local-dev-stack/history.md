# Local Dev Stack History

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-04-27

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-27 | 1.1 | 修正 W1 gate 口径：parent Phase 3 只锁定 A2 spec-contract；真实 `make dev-up` / `make dev-doctor` 可执行 gate 由 A2 child `001` plan 验证后再放行依赖本地栈的 W2 implementation | engineering-roadmap/001 Phase 3 remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定 `deploy/dev-stack/docker-compose.yaml` 落点、7 个本地依赖服务清单（Postgres+pgvector / Redis / MinIO / OTel / Grafana / Loki / Prometheus）、`make dev-up` / `dev-down` / `dev-doctor` / `dev-reset` 契约、健康检查口径与命名卷策略；承接 [engineering-roadmap §5.7](../engineering-roadmap/spec.md#57-实施-wave-顺序) 的 W1 dev-up spec-contract lock；引用 [01-technical-architecture.md §2.3](../../../easyinterview-tech-docs/01-technical-architecture.md#23-数据与基础设施) 的本地依赖列表。 | engineering-roadmap/001 Phase 3 |
