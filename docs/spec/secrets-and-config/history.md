# Secrets and Config History

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-04-27

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-27 | 1.0 | 初始创建：锁定三层 config 优先级（runtime secret > env > config/{env}.yaml > config/config.yaml）、`SecretSource` / `FeatureFlagClient` 接口、22 项 P0 必备 env key 字典、`*.secret.yaml` `.gitignore` 红线与 pre-commit / CI gitleaks 拦截策略、`/api/v1/runtime-config` 端点契约；引用 [ADR-Q3](../engineering-roadmap/decisions/ADR-Q3-analytics-platform.md) 自托管 PostHog、[ADR-Q6](../engineering-roadmap/decisions/ADR-Q6-ai-gateway-and-model-routing.md) AI Gateway env、[01-technical-architecture.md §15](../../../easyinterview-tech-docs/01-technical-architecture.md#15-发布与灰度) feature flag baseline。 | engineering-roadmap/001 Phase 3 |
