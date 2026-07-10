# Repo Scaffold History

> **版本**: 1.8
> **状态**: active
> **更新日期**: 2026-07-10

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-10 | 1.8 | 收敛 git hook 当前事实：`pre-commit` 已委托 A4 secret scan，`commit-msg` 已执行 ASCII-only message gate，不再描述为占位骨架。 | repo-scaffold/001-bootstrap |
| 2026-07-10 | 1.7 | 将根 `Makefile` D-3 / C-2 从 TODO exit-zero 占位口径收敛为当前真实委托：dev-stack、codegen 链与 backend migrate CLI 均由对应 owner 承接。 | repo-scaffold/001-bootstrap |
| 2026-07-10 | 1.6 | 根 `make fmt` 从 child Makefile TODO fallback 改为真实 `gofmt` 入口；删除 `recurse_target` fallback。 | repo-scaffold/001-bootstrap |
| 2026-05-22 | 1.4 | 对齐方案 A：`deploy/` 根容器当前只承载 A2 Docker Compose 外部依赖栈；Helm / Kind / K8s / staging / production 部署资产不再作为当前默认前提，后续由 E4 按需设计。 | local-dev-stack/001 post-pass revision |
| 2026-05-05 | 1.3 | 配置根容器边界改为引用 A3 单文件 provider registry / model profile catalog，移除旧 `config/ai-profiles/` directory 口径。 | ai-provider-and-model-routing/003 L2 remediation |
| 2026-05-05 | 1.2 | A1 根目录契约改为只引用当前 roadmap、B1/B2/B4、A3/A4/F3 owner 和顶层目录；移除旧技术草稿名称与旧目录依赖。 | engineering-roadmap/001-decompose-subspecs |
| 2026-04-29 | 1.1 | 原地修订 A1 根目录契约：把 `shared/` 与 `config/` 纳入 A1 锁定的顶层根容器，明确 B1/B3 在 `shared/` 下维护共享真理源，A3/A4/F3 在 `config/` 下维护配置、AI profile 与 prompt/rubric 相关路径，消除 A1 与 B1/A3/A4 的落点冲突。 | plan-review remediation |
| 2026-04-26 | 1.0 | 初始创建：锁定 7 个根容器目录（backend/ frontend/ openapi/ migrations/ scripts/ test/ deploy/）、顶层 Makefile target 名称、`.editorconfig` / `.tool-versions` / git hooks 落点；引用 `engineering-roadmap decisions §5.2 / §6.1` 对齐后端 / 前端目录约束 | 001-bootstrap |
