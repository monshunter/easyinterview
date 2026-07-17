# Harness Simplification 变更历史

> **版本**: 1.8
> **状态**: active
> **更新日期**: 2026-07-18

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-18 | 1.8 | 确立 `environment-build`/`environment-operate` 分工，明确 `local-dev-stack` 非 golden fixture，冻结 20→14 语义化 Skill 映射及四个用户指定删除项，并限定当前会话只对齐 Spec | 001-full-migration |
| 2026-07-18 | 1.7 | 明确 Harness 以“内置抽象/模板参数 + 项目 Spec + 自动实施验证”自举构建项目血肉；冻结 context 退出、Skill 准入收敛与 Arch-first 迁移顺序 | 001-full-migration |
| 2026-07-17 | 1.6 | 纠正“完全外部注入”偏移，确立 `Harness = Skill + Docs Arch + Env`、内置 Blueprint 的 `init-arch` 与 Arch-aware 虚实结合 Skill | 001-full-migration |
| 2026-07-17 | 1.5 | 结束 spec-only 阶段，授权按方案 A 完成新 Harness；旧 gate 在迁移期间降为基线，最终由新 gate 闭环 | 001-full-migration |
| 2026-07-17 | 1.4 | 否决一次性方案 B，确立保留式渐进收敛、独立 Skill、唯一编排 owner、风险证据经济与能力寻源 | 001-full-migration |
| 2026-07-17 | 1.0 | 初始轻量 Harness 草案与一次性迁移尝试；后续已被 1.4 取代 | 001-full-migration |
