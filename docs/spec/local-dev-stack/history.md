# Local Dev Stack History

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-04-29

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-29 | 1.5 | 文档侧 reconcile：把已落地 compose 的 MinIO / mc 不可变 tag 写回 D-2；A2 executable gate 明确同时覆盖 AI provider fail-fast C-9；§7 从未来计划改为 `001-bootstrap` 已完成事实，不新开 plan。 | plan-review remediation |
| 2026-04-27 | 1.4 | 对齐 A5 单人开发阶段决策：`make dev-doctor` JSON 仍保持可被未来 CI 消费，但当前不把 A5 CI 作为本地开发栈前置。 | [001-bootstrap](./plans/001-bootstrap/plan.md) |
| 2026-04-27 | 1.3 | 对齐 A3 / A4 AI provider 规则：docker compose 本地部署不启动 AI gateway 容器，也不使用单元测试 stub；A2 只传递 `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY` 占位，启用 AIClient 的组件缺真实 provider 配置时 fail-fast。 | [001-bootstrap](./plans/001-bootstrap/plan.md) |
| 2026-04-27 | 1.2 | 按 L1 plan-review 与用户确认修订本地开发栈边界：默认依赖收敛为 Postgres+pgvector / Redis / MinIO；`make dev-up` 改为启动最小依赖 + 当前项目可运行组件；本地观测改为应用 `/metrics` + 容器日志；默认排除 OTel Collector / Grafana / Loki / Prometheus 与 AI gateway。 | [001-bootstrap](./plans/001-bootstrap/plan.md) |
| 2026-04-27 | 1.1 | spawn `001-bootstrap` impl plan：把 spec §3.1 D-1..D-7 与 §6 C-1..C-8 落到 4 个 phase（compose+init / make 生命周期 / dev-doctor / OTel 通路+收口），关闭 [engineering-roadmap/001 Phase 3.6](../engineering-roadmap/plans/001-decompose-subspecs/checklist.md#phase-3-wave-1基础设施--契约骨架) 的「executable gate by A2 child」承诺。spec 内容未变，版本不动 | [001-bootstrap](./plans/001-bootstrap/plan.md) |
| 2026-04-27 | 1.1 | 修正 W1 gate 口径：parent Phase 3 只锁定 A2 spec-contract；真实 `make dev-up` / `make dev-doctor` 可执行 gate 由 A2 child `001` plan 验证后再放行依赖本地栈的 W2 implementation | engineering-roadmap/001 Phase 3 remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定 `deploy/dev-stack/docker-compose.yaml` 落点、7 个本地依赖服务清单（Postgres+pgvector / Redis / MinIO / OTel / Grafana / Loki / Prometheus）、`make dev-up` / `dev-down` / `dev-doctor` / `dev-reset` 契约、健康检查口径与命名卷策略；承接 [engineering-roadmap §5.7](../engineering-roadmap/spec.md#57-实施-wave-顺序) 的 W1 dev-up spec-contract lock；引用 [01-technical-architecture.md §2.3](../../../easyinterview-tech-docs/01-technical-architecture.md#23-数据与基础设施) 的本地依赖列表。 | engineering-roadmap/001 Phase 3 |
