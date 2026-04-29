# CI Pipeline Baseline History

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-04-29

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-29 | 1.3 | 物化 `001-local-quality-gates` 为本地门禁聚合 plan；`make docs-check` 改为可执行 sync-doc-index 脚本 + 链接检查命令；远端 CI 明确后续走 `002-remote-ci`，不得塞回 001。 | plan-review remediation |
| 2026-04-27 | 1.2 | 按用户决策修订为个人单人开发阶段不构建远端 CI pipeline：A5 当前只约束本地手动质量门禁，GitHub Actions / branch protection / artifact / nightly / CI secret 全部延后到多人协作、公开 release 或自动发版触发条件出现后再建。 | engineering-roadmap/001-decompose-subspecs remediation |
| 2026-04-27 | 1.1 | 对齐 A3/A4 AI provider 规则：PR CI 不注入 AI provider secrets，单元测试继续走 stub / fixtures；本地部署真实 provider 校验由 A2/A3/A4 本地栈与 Kind 路径承接，不在 CI runner 中执行。 | local-dev-stack/001-bootstrap review remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定 GitHub Actions 平台、`.github/workflows/{ci,nightly,dependabot}.yml` 入口、必跑 job 矩阵（lint-go / lint-ts / lint-config / lint-error-codes / unit-test-go / unit-test-ts / build-api / build-worker / build-frontend / codegen-drift-check / docs-check）、缓存策略、artifact 输出、secret 红线（业务 secret 不进 CI）、branch protection 同步策略；引用 [B1 D-1 idempotent generator](../shared-conventions-codified/spec.md#41-真理源约束) 与 [AGENTS.md / CLAUDE.md §7 git 分支策略](../../../CLAUDE.md#7-git-分支策略)。 | engineering-roadmap/001 Phase 3 |
