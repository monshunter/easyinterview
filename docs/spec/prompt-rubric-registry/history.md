# Prompt Rubric Registry History

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-04-27

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-27 | 1.1 | 修正 W1 gate 口径：parent Phase 3 只锁定 F3 feature_key / version / language 坐标、文件落点与 Resolve 契约；真实 baseline prompt/rubric 文件、loader 与 lint 由 F3 child `001` plan 验证后再放行依赖 F3 的 W2 implementation | engineering-roadmap/001 Phase 3 remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定 prompt / rubric 三元组 `(feature_key, version, language)`、`config/{prompts,rubrics}/<feature_key>/<version>` 文件落点、`RegistryClient.Resolve` 业务调用契约、template_hash 校验、灰度规则、LLM Judge 接口锁定（实现归 W3）；§3.1.1 13 个 P0 feature_key 字典覆盖 C4-C7 + C9 + C11 全部 AI task 命名空间；引用 [ADR-Q6 §3.6 F3 解耦](../engineering-roadmap/decisions/ADR-Q6-ai-gateway-and-model-routing.md)、[03-db-definition.md §5.8](../../../easyinterview-tech-docs/03-db-definition.md)、[01-technical-architecture.md §10](../../../easyinterview-tech-docs/01-technical-architecture.md#10-ai-编排层设计)、[engineering-roadmap §5.7 W1 baseline prompt spec-contract lock](../engineering-roadmap/spec.md#57-实施-wave-顺序)。 | engineering-roadmap/001 Phase 3 |
