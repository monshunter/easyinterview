# Secrets and Config History

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-04-28

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-28 | 1.5 | 根据 L1 plan-review 修订 A4 spec：移除 JWT/access-token 口径并对齐 ADR-Q1 session cookie + magic link secret；统一 PostHog env key 到 ADR-Q3；补齐 runtime-config public allowlist / analytics opt-out 边界；新增 canonical config schema 分类与对应验收项。 | plan-review remediation |
| 2026-04-27 | 1.4 | 清理剩余 CI gitleaks 表述：当前 secret 防护由 pre-commit 与本地 gitleaks 收口，远端 CI secret scan 仅在 A5 触发条件成立后再接入。 | ci-pipeline-baseline spec-contract remediation |
| 2026-04-27 | 1.3 | 对齐 A5 单人开发阶段决策：A4 当前只要求本地 lint / pre-commit / gitleaks 质量门禁，不把 PR 阶段或 CI secret scan 写成 P0 前置。 | ci-pipeline-baseline spec-contract remediation |
| 2026-04-27 | 1.2 | 对齐 A3 v1.3 与 ADR-Q6 v1.1：`AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY` 对 local deploy（docker compose / Kind）、staging、prod 必填，unit test 可空并走 stub；非 test 环境缺失真实 AI provider / gateway 配置时必须 fail-fast。 | local-dev-stack/001-bootstrap review remediation |
| 2026-04-27 | 1.1 | 对齐 A2 local-dev-stack v1.2：`OTEL_EXPORTER_OTLP_ENDPOINT` 改为可选观测 / 生产条件字段，普通本地 dev 默认不上报；PostHog 部署归 F2/E4，A2 默认本地栈只要求 no-op / file-backed mode 不阻塞启动。 | local-dev-stack/001-bootstrap review remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定三层 config 优先级（runtime secret > env > config/{env}.yaml > config/config.yaml）、`SecretSource` / `FeatureFlagClient` 接口、22 项 P0 必备 env key 字典、`*.secret.yaml` `.gitignore` 红线与 pre-commit / CI gitleaks 拦截策略、`/api/v1/runtime-config` 端点契约；引用 [ADR-Q3](../engineering-roadmap/decisions/ADR-Q3-analytics-platform.md) 自托管 PostHog、[ADR-Q6](../engineering-roadmap/decisions/ADR-Q6-ai-gateway-and-model-routing.md) AI Gateway env、[01-technical-architecture.md §15](../../../easyinterview-tech-docs/01-technical-architecture.md#15-发布与灰度) feature flag baseline。 | engineering-roadmap/001 Phase 3 |
