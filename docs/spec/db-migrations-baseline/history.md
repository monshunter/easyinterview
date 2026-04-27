# DB Migrations Baseline History

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-04-27

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-27 | 1.0 | 初始创建：锁定 `golang-migrate` 工具、`migrations/` 目录与 `NNNNNN_<verb>_<noun>.{up,down}.sql` 文件命名、27 张 P0 应用表（与 [03 §4](../../../easyinterview-tech-docs/03-db-definition.md#4-表清单) 一致）+ `schema_migrations` 元数据 + pgvector 扩展、03 §7 全部索引、`make migrate-{up,down,status,create}` target、可逆 + 数据回填策略、prod 防呆与 enum 与 B1 同源约束；引用 [03 §9 迁移策略](../../../easyinterview-tech-docs/03-db-definition.md#9-迁移策略) 与 [B1 D-6 枚举](../shared-conventions-codified/spec.md#31-已锁定决策)。 | engineering-roadmap/001 Phase 3 |
