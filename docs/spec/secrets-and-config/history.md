# Secrets and Config History

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-05-05

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-05 | 2.0 | 将 AI 连接 env/config 真理源收口为 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 与 `ai.provider*`，确认非 test 环境缺失 provider config 时 fail-fast，不保留旧连接参数兼容层。 | ai-provider-and-model-routing/001 remediation |
| 2026-05-03 | 1.9 | 对齐 product-scope v1.2：feature flag baseline 删除独立错题本导出、独立成长中心与旧 dual-track 表述，改为报告复练计划、准备度信号和会话内辅助程度开关。 | 001-bootstrap Phase 8 remediation |
| 2026-04-30 | 1.8 | §4.1 边界约束扩展 allow-list：将 `backend/cmd/migrate/main.go` 与 cmd/api / cmd/worker 同列纳入 `os.Getenv` 允许前缀，同步 `scripts/lint/getenv_boundary.go` `defaultAllowlist`。原因：B4 db-migrations-baseline 引入 `cmd/migrate` CLI 后 A4 spec 未跟进，导致 A5 `make lint-getenv-boundary` 把 cmd/migrate 当作违规。本次修订承认 cmd/migrate 与 cmd/api / cmd/worker 是同类 CLI 入口。 | ci-pipeline-baseline/001-local-quality-gates 验证发现 |
| 2026-04-29 | 1.7 | 收口 A/B plan-review：确认 P0 env key 字典为 24 项；新增 `async.queueWeights` config-only 字段供 C8 Asynq 队列权重消费；PostHog 临时不可用时只允许 last-known-good 缓存降级；移除文档中的真实形态 secret 样本。 | plan-review remediation |
| 2026-04-28 | 1.6 | 对齐 ADR-Q1 v1.2：锁定 session cookie 字面量 `ei_session`，明确 A4 只管理 session secret，不提供 cookie name env/config override。 | openapi-v1-contract/001-bootstrap assessment remediation |
| 2026-04-28 | 1.5 | 根据 L1 plan-review 修订 A4 spec：移除 JWT/access-token 口径并对齐 ADR-Q1 session cookie + magic link secret；统一 PostHog env key 到 ADR-Q3；补齐 runtime-config public allowlist / analytics opt-out 边界；新增 canonical config schema 分类与对应验收项。 | plan-review remediation |
| 2026-04-27 | 1.4 | 清理剩余 CI gitleaks 表述：当前 secret 防护由 pre-commit 与本地 gitleaks 收口，远端 CI secret scan 仅在 A5 触发条件成立后再接入。 | ci-pipeline-baseline spec-contract remediation |
| 2026-04-27 | 1.3 | 对齐 A5 单人开发阶段决策：A4 当前只要求本地 lint / pre-commit / gitleaks 质量门禁，不把 PR 阶段或 CI secret scan 写成 P0 前置。 | ci-pipeline-baseline spec-contract remediation |
| 2026-04-27 | 1.2 | 对齐 A3 v1.3 与 ADR-Q6 v1.1：`AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 对 local deploy（docker compose / Kind）、staging、prod 必填，unit test 可空并走 stub；非 test 环境缺失真实 AI provider 配置时必须 fail-fast。 | local-dev-stack/001-bootstrap review remediation |
| 2026-04-27 | 1.1 | 对齐 A2 local-dev-stack v1.2：`OTEL_EXPORTER_OTLP_ENDPOINT` 改为可选观测 / 生产条件字段，普通本地 dev 默认不上报；PostHog 部署归 F2/E4，A2 默认本地栈只要求 no-op / file-backed mode 不阻塞启动。 | local-dev-stack/001-bootstrap review remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定三层 config 优先级（runtime secret > env > config/{env}.yaml > config/config.yaml）、`SecretSource` / `FeatureFlagClient` 接口、22 项 P0 必备 env key 字典、`*.secret.yaml` `.gitignore` 红线与 pre-commit / CI gitleaks 拦截策略、`/api/v1/runtime-config` 端点契约；引用 [ADR-Q3](../engineering-roadmap/decisions/ADR-Q3-analytics-platform.md) 自托管 PostHog、[ADR-Q6](../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md) AI Gateway env、[01-technical-architecture.md §15](../../../easyinterview-tech-docs/01-technical-architecture.md#15-发布与灰度) feature flag baseline。 | engineering-roadmap/001 Phase 3 |
