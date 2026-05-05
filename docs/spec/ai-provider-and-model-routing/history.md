# AI Provider and Model Routing History

> **版本**: 2.3
> **状态**: active
> **更新日期**: 2026-05-05

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-05 | 2.3 | 003 原地修订：将 Model Profile active truth source 从 per-profile YAML 目录收敛为单一 `config/ai-profiles.yaml` catalog，`AI_MODEL_PROFILE_PATH` 改为 catalog 文件路径。 | 003-provider-registry-and-capability-profiles catalog consolidation |
| 2026-05-05 | 2.2 | 003 L2 remediation 收口：补生产 registry/profile bootstrap、profile reload warn、active profile anti-stub gate，并将 003 状态投影恢复为 completed。 | 003-provider-registry-and-capability-profiles L2 remediation |
| 2026-05-05 | 2.1 | 003 Phase 5 负向搜索收口：active scope 部署文案改为 registry/profile/provider ref 契约，不再以单一 AI provider endpoint 作为当前目标架构主语。 | 003-provider-registry-and-capability-profiles Phase 5 |
| 2026-05-05 | 2.0 | 003 实施推进到 A4/B1/F3 集成：A3 runtime `Capability` 改为消费 B1 生成 capability，provider/profile routing 错误码补齐为 B1-owned `AI_*`，profile coverage lint 接入顶层 lint gate。 | 003-provider-registry-and-capability-profiles Phase 4 |
| 2026-05-05 | 1.9 | 基于 product-scope 与 UI 交互重估 AI 使用场景，将 A3 目标从单一 provider endpoint 升级为 provider registry + capability-scoped Model Profile；新增 Product/UI AI capability catalog、central fallback 边界、F3 12 profile coverage gate，并创建 003 实施计划。 | 003-provider-registry-and-capability-profiles |
| 2026-05-05 | 1.8 | 全面收口 AI provider 口径：subject 与 ADR 文件更名，运行时连接参数改为 `AI_PROVIDER_*` / `ai.provider*`，Model Profile route schema 改为 `route`，并新增 active surface 负向 terminology gate。 | 001-aiclient-and-profile-bootstrap Phase 6 |
| 2026-04-29 | 1.7 | 物化 A3 plan 设计：`001-aiclient-and-profile-bootstrap` 切为 active，`002-tools-streaming-and-stt` 保持 draft/blocked；明确 001 只 owns `backend/internal/ai/aiclient/` 与 `config/ai-profiles/` fixture，API/worker entrypoint 是 DI handoff，不由 A3 001 创建。 | plan-review remediation |
| 2026-04-29 | 1.6 | 按 ADR-Q6 authoritative 边界收口：`AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 是 AIClient 的 OpenAI-compatible 连接参数，可指真实 LLM provider 或生产 provider endpoint；fallback 只由连接 provider endpoint route 承担，A3 client 不自行切换模型；B1 提供共享字段名 / 错误码，A3 owns Model Profile schema 与 `AICallMeta` runtime。 | plan-review remediation |
| 2026-04-29 | 1.5 | 修复 L1 review 发现的一致性与完备性问题：`AICallMeta` 明确归 A3 runtime 拥有，B1 提供共享错误码；metric 语义区分 per-call 与 event-only counter；STT / Audio Transcription 降为 C14 P2 预留；补齐 `Stream` 事件合同边界。 | plan-review remediation |
| 2026-04-27 | 1.4 | 对齐 A5 单人开发阶段决策：A3 的测试红线当前由本地 lint / test gate 强制，远端 CI 仅在 A5 触发条件成立后再接入。 | ci-pipeline-baseline spec-contract remediation |
| 2026-04-27 | 1.3 | 明确 stub 只用于单元测试 / 离线契约测试 / 显式 mock；docker compose 与 Kind 本地部署、staging、prod 必须通过 `AI_PROVIDER_BASE_URL` + `AI_PROVIDER_API_KEY` 指向真实 AI provider 或生产 provider endpoint，缺失时 fail-fast，不默认降级到 stub。 | local-dev-stack/001-bootstrap review remediation |
| 2026-04-27 | 1.2 | 对齐 A2 local-dev-stack v1.2：本地开发与单元测试默认走应用内 deterministic stub provider；A2 不再预留或启动 `ai-gateway-mock` compose 服务，AI provider 不作为本地开发依赖。 | local-dev-stack/001-bootstrap review remediation |
| 2026-04-27 | 1.1 | 对齐 F1 baseline metric label contract：fallback metric 使用 `from_model_family` / `to_model_family` / `result`，不把原始 model id 写入 Prometheus label | engineering-roadmap/001 Phase 3 remediation |
| 2026-04-27 | 1.0 | 初始创建：把 [ADR-Q6](../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md) 的 9 项硬约束落到 `AIClient` 接口、Model Profile schema、stub / openai_compatible provider、观测埋点、隐私红线、fallback 边界；引用 `engineering-roadmap decisions §10`、`F1 observability-stack §8`、`F1 observability-stack logging §4.4`、`B4 db-migrations-baseline §5.8` 中的 `ai_task_runs` schema。 | engineering-roadmap/001 Phase 3 |
