# Prompt Rubric Registry History

> **版本**: 2.2
> **状态**: active
> **更新日期**: 2026-05-16

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-16 | 2.2 | backend-debrief 001 Phase 0.5 新增 `debrief.suggest_questions` baseline feature_key、默认 model profile 与 prompt/rubric/seed truth source。 | backend-debrief/001 Phase 0.5 |
| 2026-05-09 | 2.1 | 派生并修订 `001-baseline` impl plan：把 spec C-1~C-11 与 D-1~D-12 端到端落到 5-phase plan（truth source + lint + Go registry + targetjob retire + DB seed + A3 coverage handoff）。用户决策：A3 profile coverage 仅校验 entry 存在 + status 合法（不动 catalog），DB seed 纳入本 plan，prompt body 写真实可用文案，LLM Judge 落 interface + NotImplementedJudge stub；同时确认 `ai_task_runs` 必须 typed 承载 `feature_key` / `feature_flag` / `data_source_version`，由本 plan 先修 B1/B4/A3 契约再实施。spec.md Header 升至 v2.1。 | prompt-rubric-registry/001-baseline |
| 2026-05-08 | 2.0 | 对齐 A3 003 Phase 6：当前 baseline feature_key 从 12 项收敛为 10 项，删除 C11 资料检索类占位；未来如需要再重新设计。 | ai-provider-and-model-routing/003 Phase 6 |
| 2026-05-06 | 1.9 | 对齐 A3 002 Tools / streaming handoff：记录 F3 Resolve 后续可输出 provider-neutral `tools[]`、`output_schema`、`stream_wire` hints，且不得携带 provider/model 字符串。 | ai-provider-and-model-routing/002 Phase 5 |
| 2026-05-05 | 1.8 | 对齐 A3 003 catalog consolidation：F3 默认 `model_profile_name` coverage gate 指向单一 `config/ai-profiles.yaml` catalog。 | ai-provider-and-model-routing/003 catalog consolidation |
| 2026-05-05 | 1.7 | 将 A3 profile coverage gate 固化为 `make lint-ai-profile-coverage` / 顶层 `make lint` 的可执行门禁，覆盖当时 baseline feature_key 默认 profile 与 A3 Product/UI capability catalog。 | ai-provider-and-model-routing/003 Phase 4 |
| 2026-05-05 | 1.6 | F3 feature_key、prompt/rubric 坐标、model profile reference 与 lint gate 改为只由本 spec、A3/B4 协作边界和后续编码 truth source 承接；移除旧技术草稿名称和旧 shorthand 依赖。 | engineering-roadmap/001-decompose-subspecs |
| 2026-05-05 | 1.5 | 对齐 A3 provider registry + capability profile 设计：F3 仍只输出 `model_profile_name`，但当时 baseline 默认 profile 必须被 A3 profile catalog 覆盖并通过 capability/provider_ref lint。 | ai-provider-and-model-routing/003 |
| 2026-05-03 | 1.4 | 当前 feature_key baseline 收敛为 12 项，删除独立 `mistake.extract`，报告内题目回顾 / 本轮复练由 `report.generate` 与 `report.question_assessment` 承载。 | docs-only |
| 2026-04-29 | 1.3 | 对齐 engineering-roadmap v2.0 的 C9 P0 调整：`debrief.generate` feature_key 现在服务 P0 真实面试复现 / 复盘文本流；感谢信草稿与完整跟进建议仍作为 C9 P1 增强，当时的资料检索命名空间仅作 P1 占位。 | plan-review remediation |
| 2026-04-27 | 1.2 | 对齐 A5 单人开发阶段决策：F3 当前只要求本地 prompt/rubric lint 与 template_hash drift gate，远端 CI 不作为 P0 前置。 | ci-pipeline-baseline spec-contract remediation |
| 2026-04-27 | 1.1 | 修正 W1 gate 口径：parent Phase 3 只锁定 F3 feature_key / version / language 坐标、文件落点与 Resolve 契约；真实 baseline prompt/rubric 文件、loader 与 lint 由 F3 child `001` plan 验证后再放行依赖 F3 的 W2 implementation | engineering-roadmap/001 Phase 3 remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定 prompt / rubric 三元组 `(feature_key, version, language)`、`config/{prompts,rubrics}/<feature_key>/<version>` 文件落点、`RegistryClient.Resolve` 业务调用契约、template_hash 校验、灰度规则、LLM Judge 接口锁定（实现归 W3）；§3.1.1 当时的 P0 feature_key 字典覆盖 C4-C7 + C9 + 资料检索命名空间；引用 [ADR-Q6 §3.6 F3 解耦](../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md)、`B4 db-migrations-baseline §5.8`、`engineering-roadmap decisions §10`、[engineering-roadmap §5.7 W1 baseline prompt spec-contract lock](../engineering-roadmap/spec.md#6-实施顺序)。 | engineering-roadmap/001 Phase 3 |
