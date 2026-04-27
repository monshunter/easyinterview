# CI Pipeline Baseline History

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-04-27

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-27 | 1.0 | 初始创建：锁定 GitHub Actions 平台、`.github/workflows/{ci,nightly,dependabot}.yml` 入口、必跑 job 矩阵（lint-go / lint-ts / lint-config / lint-error-codes / unit-test-go / unit-test-ts / build-api / build-worker / build-frontend / codegen-drift-check / docs-check）、缓存策略、artifact 输出、secret 红线（业务 secret 不进 CI）、branch protection 同步策略；引用 [B1 D-1 idempotent generator](../shared-conventions-codified/spec.md#41-真理源约束) 与 [AGENTS.md / CLAUDE.md §7 git 分支策略](../../../CLAUDE.md#7-git-分支策略)。 | engineering-roadmap/001 Phase 3 |
