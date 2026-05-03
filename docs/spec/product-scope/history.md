# Product Scope History

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-05-03

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-03 | 1.4 | 明确 `easyinterview-tech-docs/` 仅为历史技术输入，当前 API / DB / event / metrics 以 Layer B/F active spec 与已编码 truth source 为准，避免旧技术包按旧产品 spec 误导实现。 | docs-only |
| 2026-05-03 | 1.3 | 删除旧 `voice` route alias 语义，明确语音面试只通过 `practice` 显式携带 `mode=voice` / `modality=voice` 进入；同步 UI 静态原型删除历史独立语音页组件。 | docs-only |
| 2026-05-03 | 1.2 | 记录 engineering-roadmap v2.2 已完成产品范围对齐：P0 前端 child 改为当前 UI 五入口相关拆分，旧 growth / mistakes / drill / plan / voice page 不再作为后续 pending child 恢复依据。 | engineering-roadmap/001-decompose-subspecs |
| 2026-05-03 | 1.1 | 修正范围语义：把热身、反问专练、单题深钻、独立错题本、独立成长中心等从“P0 不做”改为“已决策丢弃 / 长期不恢复”，明确当前 UI / UI 文档未保留且本 spec 未列为规划例外的旧能力默认丢弃；仅保留 `全球多平台搜岗` 作为规划例外。 | docs-only |
| 2026-05-03 | 1.0 | 初始创建：将产品范围真理源从根目录 `easyinterview-spec-v1-0.md` 迁入 spec-centric 文档体系，吸收当前 `ui-design/` 与 `docs/ui-design/` 已确认的导航、模块边界、报告、简历、复盘和删除范围，并保留旧 spec 的目标用户、阶段路线、隐私伦理与质量评估水准。 | docs-only |
