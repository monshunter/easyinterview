# Repo Scaffold History

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-04-29

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-29 | 1.1 | 原地修订 A1 根目录契约：把 `shared/` 与 `config/` 纳入 A1 锁定的顶层根容器，明确 B1/B3 在 `shared/` 下维护共享真理源，A3/A4/F3 在 `config/` 下维护配置、AI profile 与 prompt/rubric 相关路径，消除 A1 与 B1/A3/A4 的落点冲突。 | plan-review remediation |
| 2026-04-26 | 1.0 | 初始创建：锁定 7 个根容器目录（backend/ frontend/ openapi/ migrations/ scripts/ test/ deploy/）、顶层 Makefile target 名称、`.editorconfig` / `.tool-versions` / git hooks 落点；引用 [01-technical-architecture.md §5.2 / §6.1](../../../easyinterview-tech-docs/01-technical-architecture.md) 对齐后端 / 前端目录约束 | 001-bootstrap |
