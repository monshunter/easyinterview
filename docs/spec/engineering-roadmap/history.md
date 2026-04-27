# Engineering Roadmap History

> **版本**: 1.8
> **状态**: active
> **更新日期**: 2026-04-27

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-27 | 1.8 | 对齐个人单人开发阶段决策：A5 不再作为 P0 远端 CI pipeline，当前只保留本地质量门禁；GitHub Actions、branch protection、artifact、nightly 与 CI secret 延后到多人协作、公开 release 或自动发版触发条件出现后再建。 | 001-decompose-subspecs |
| 2026-04-27 | 1.7 | 同步修订 ADR-Q4/Q6 与 A2/A3/A4 AI provider 边界：unit test / 离线契约测试才走 stub；docker compose 与 Kind 本地部署必须注入真实 AI provider endpoint / key，staging/prod 可指 cluster-internal gateway。 | 001-decompose-subspecs |
| 2026-04-27 | 1.6 | 同步修订 ADR-Q3/Q4 与 A2 本地开发栈边界：PostHog 部署验证改归 F2/E4，普通本地 dev 默认 no-op / file-backed；Kind 仅用于场景集成测试，A2 docker-compose 不再与 Kind manifest 同源。 | 001-decompose-subspecs |
| 2026-04-27 | 1.5 | 对齐 A2 local-dev-stack v1.2：parent roadmap 中 A2 口径改为最小依赖（Postgres+pgvector / Redis / MinIO）+ 项目组件一键启动；F1 口径改为消费应用 `/metrics` / 日志与生产观测配置，不再要求 A2 默认提供 OTel/Grafana/Loki/Prometheus。 | 001-decompose-subspecs |
| 2026-04-27 | 1.4 | 修正 W1 Phase 3 gate 口径：parent plan 只完成 9 份 child spec 的 cross-spec review 与 spec-contract lock；A2/B2/F1/F3 的可执行 gate 交由各 child `001` plan 逐一验证，未通过前不得启动依赖它的 W2 implementation | 001-decompose-subspecs |
| 2026-04-26 | 1.3 | L2 code review remediation：ADR-Q3 从 PostHog Cloud 切换为自托管 PostHog 优先；补齐 async public `jobType` 与内部 Asynq handler 的命名边界；明确 Q-5 P0 导出延后是产品验收项的 W0 例外 | 001-decompose-subspecs |
| 2026-04-26 | 1.2 | W0 hard gate 6 项 ADR 全部 accepted（Q-1 自建 passwordless / Q-2 Asynq+Redis / Q-3 PostHog Cloud EU / Q-4 Kubernetes / Q-5 P0 仅删除 / Q-6 AIClient+Model Profile+外部 AI Gateway）；§3.2 表替换为锁定结论 + ADR 链接；§4.4 / §5.5 E4 / §6 C-1 / C-6 同步引用 ADR | 001-decompose-subspecs |
| 2026-04-26 | 1.1 | 补充 Q-6 AI 网关与模型路由 W0 决策项；A3 改为 provider-neutral `ai-gateway-and-model-routing`；明确 Higress 等 AI Gateway 作为独立部署组件而非业务 SDK | 001-decompose-subspecs |
| 2026-04-26 | 1.0 | 初始创建：定义 6 层 38 child subspec、6 wave 实施顺序、mock-first 集成策略、5 项 W0 hard gate 决策项 | 001-decompose-subspecs |
