# Product Scope History

> **版本**: 2.2
> **状态**: active
> **更新日期**: 2026-06-29

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-06-29 | 2.2 | 用户确认方案 B：删除真实面试复盘和用户画像，将 P0 收敛为 JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮；新增 D-22、C-19 和 `001-core-loop-module-pruning` owner plan。 | product-scope/001-core-loop-module-pruning |
| 2026-05-05 | 1.8 | owner matrix 的 AI provider / model profile 可执行真理源改为 `config/ai-providers.yaml` 与 `config/ai-profiles.yaml`，移除旧 profile directory 口径。 | ai-provider-and-model-routing/003 L2 remediation |
| 2026-05-05 | 1.7 | 明确技术契约 owner matrix 是当前唯一分层入口：旧技术草稿实体与名称均不再保留，API / DB / event / metrics / logging / config / AI 等责任由当前 owner spec 与编码 truth source 独立承接。 | engineering-roadmap/001-decompose-subspecs |
| 2026-05-05 | 1.6 | 增加当前技术契约统一 owner matrix，把 API / DB / event / metrics / logging 等职责映射到当前 A/B/F owner spec 与编码 truth source。 | docs-only |
| 2026-05-03 | 1.5 | 同步 engineering-roadmap v3.0：删除 `docs/spec/INDEX.md` pending child 占位模型，后续 child 仅按当前产品 / UI 已保留能力 on-demand 创建，旧 route / 旧 pending 名称不得恢复已丢弃模块。 | engineering-roadmap/001-decompose-subspecs |
| 2026-05-03 | 1.4 | 明确当前 API / DB / event / metrics 以 Layer B/F active spec 与已编码 truth source 为准，避免历史背景误导实现。 | docs-only |
| 2026-05-03 | 1.3 | 删除旧 `voice` route alias 语义，明确语音面试只通过 `practice` 显式携带 `mode=voice` / `modality=voice` 进入；同步 UI 静态原型删除历史独立语音页组件。 | docs-only |
| 2026-05-03 | 1.2 | 记录 engineering-roadmap v2.2 已完成产品范围对齐：P0 前端 child 改为当前 UI 五入口相关拆分，旧 growth / mistakes / drill / plan / voice page 不再作为后续 pending child 恢复依据。 | engineering-roadmap/001-decompose-subspecs |
| 2026-05-03 | 1.1 | 修正范围语义：把热身、反问专练、单题深钻、独立错题本、独立成长中心等从“P0 不做”改为“已决策丢弃 / 长期不恢复”，明确当前 UI / UI 文档未保留且本 spec 未列为规划例外的旧能力默认丢弃；仅保留 `全球多平台搜岗` 作为规划例外。 | docs-only |
| 2026-05-03 | 1.0 | 初始创建：将产品范围真理源从根目录 `easyinterview-spec-v1-0.md` 迁入 spec-centric 文档体系，吸收当前 `ui-design/` 与 `docs/ui-design/` 已确认的导航、模块边界、报告、简历、复盘和删除范围，并保留旧 spec 的目标用户、阶段路线、隐私伦理与质量评估水准。 | docs-only |
