# Shared Conventions Codified History

> **版本**: 1.23
> **状态**: active
> **更新日期**: 2026-07-06

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-06 | 1.23 | docs-only：将 B1 active spec 正文收敛为当前 16 个生成枚举、flat Resume vocabulary 与 `RESUME_EXPORT_NOT_AVAILABLE` 错误码边界。 | product-scope/001-core-loop-module-pruning |
| 2026-06-29 | 1.22 | product-scope D-22 后同步 shared conventions：`PracticeGoal` 当前只保留 baseline / retry_current_round / next_round；`RESOURCE_NOT_FOUND` 保留为 generic 404。 | product-scope/001-core-loop-module-pruning |
| 2026-05-21 | 1.20 | 授权 backend-profile/001 Phase 1 新增 `RESOURCE_NOT_FOUND` 错误码（`httpStatus: 404`，`retryable: false`，`message: "requested resource not found or not accessible"`），作为 cross-resource generic 404 通用码；同步 `shared/conventions.yaml`、Go/TS generated errors 与 B2 OpenAPI `ApiErrorCode` enum。前缀字典追加 `RESOURCE_NOT_FOUND`。 | backend-profile/001-candidate-profile-and-experience-cards Phase 1 |
| 2026-05-17 | 1.19 | 授权 backend-resume/002 Phase 1 新增 `RESUME_STRUCTURED_MASTER_ALREADY_EXISTS` 错误码（`httpStatus: 409`，`retryable: false`，`message: "structured master resume version already exists for this resume asset"`），用于 `confirmResumeStructuredMaster` 重复确认冲突；同步 `shared/conventions.yaml`、Go/TS generated errors 与 B2 OpenAPI `ApiErrorCode` enum。 | backend-resume/002-versions-tailor-runs-and-save-v1 Phase 1 |
| 2026-05-15 | 1.18 | 授权 backend-review/001 Phase 0.1 新增 `REPORT_NOT_FOUND` 错误码（`httpStatus: 404`，`retryable: false`，`message: "feedback report not found or not accessible"`），用于 cross-user 隔离 404 响应；同步 generated Go `ErrReportNotFound` + generated TS 等价常量；与 [B2 openapi-v1-contract](../openapi-v1-contract/history.md) `FeedbackReport.errorCode` schema + `getFeedbackReport` 404 mapping 同 commit。生成错误码 +1，不引入兼容 alias。 | backend-review/001-report-generation-baseline Phase 0.1 |
| 2026-05-12 | 1.17 | D-10 Resume Workshop additive vocabulary 落地阶段：`shared/conventions.yaml` 新增 `ResumeVersionType` / `ResumeSeedStrategy` / `ResumeTailorSuggestionStatus` 与 `RESUME_EXPORT_NOT_AVAILABLE`；Go/TS generated conventions、parity fixture 与 lint gate 已同步，生成枚举类型 14 → 17。 | openapi-v1-contract/004-resume-additive-coverage |
| 2026-05-11 | 1.16 | D-10 Resume Workshop additive vocabulary 声明阶段：新增 3 个生成枚举（`ResumeVersionType` / `ResumeSeedStrategy` / `ResumeTailorSuggestionStatus`）+ 1 个错误码 `RESUME_EXPORT_NOT_AVAILABLE` + 术语映射决策 UI `ResumeSource` ≡ OpenAPI `ResumeAsset`；具体 `shared/conventions.yaml` 字面量 + Go/TS generated 类型由 openapi-v1-contract/004-resume-additive-coverage 落地后再回填 14 → 17 枚举类型。 | openapi-v1-contract/004-resume-additive-coverage（声明阶段，docs-only） |
| 2026-05-09 | 1.15 | 授权 backend-practice Phase 0 共享契约修订：`PracticeMode` 收敛为 `assisted` / `strict`，新增 `PRACTICE_PLAN_NOT_FOUND` / `PRACTICE_SESSION_NOT_FOUND` 错误码，并同步 `shared/conventions.yaml`、Go/TS generated errors 与 B2 OpenAPI error enum。 | backend-practice/001 Phase 0 |
| 2026-05-08 | 1.14 | 授权 C4 TargetJob 场景错误码：`TARGET_JOB_NOT_FOUND`、`TARGET_IMPORT_SOURCE_INVALID`、`TARGET_IMPORT_SOURCE_UNAVAILABLE`、`TARGET_INVALID_STATE_TRANSITION`；后续 Phase 0 必须同步 `shared/conventions.yaml` 与 Go/TS/OpenAPI generated artifacts。 | backend-targetjob/001 Phase 0 |
| 2026-05-08 | 1.13 | 对齐 A3 003 Phase 6：AI capability vocabulary 删除向量化 / 重排当前能力，只保留 `chat/stt/realtime/judge`。 | ai-provider-and-model-routing/003 Phase 6 |
| 2026-05-06 | 1.12 | 对齐 backend-runtime-topology：业务后台执行边界从 worker 泛称改为 backend background runner，不在 B1 引入独立 worker 前置。 | backend-runtime-topology/001-worker-consolidation |
| 2026-05-05 | 1.11 | 扩展 AI shared vocabulary：新增 AI capability、provider registry/profile 字段名、fallback meta 字段，以及 `AI_UNSUPPORTED_CAPABILITY` / provider config / provider secret 三类 routing 错误码，Go/TS/OpenAPI 生成物同步消费。 | ai-provider-and-model-routing/003 Phase 4 |
| 2026-05-05 | 1.10 | B1 共享约定职责改为由本 spec、`shared/conventions.yaml`、generator 与 lint gate 独立承接；移除旧技术草稿名称和旧 shorthand 依赖。 | engineering-roadmap/001-decompose-subspecs |
| 2026-05-05 | 1.9 | 同步 A3/A4 AI provider 命名：B1 shared AI vocabulary 与 generated owner-boundary 注释只引用 `AI_PROVIDER_*` 连接参数，不传播旧连接命名。 | ai-provider-and-model-routing/001 remediation |
| 2026-05-03 | 1.8 | 明确当前共享约定以 `shared/conventions.yaml` 与本 spec 为准；新增枚举 / 错误码必须先修订当前 owner，不依赖外部草稿。 | docs-only |
| 2026-05-03 | 1.7 | 对齐 product-scope v1.2 / UI scope：练习入口枚举从旧模式卡片改为会话内 `assisted` / `strict`，复练目标改为 `retry_current_round` / `next_round`，旧 `MistakeStatus` 收敛为报告内部 `QuestionReviewStatus`。 | 001-bootstrap Phase 5 remediation |
| 2026-04-29 | 1.6 | 物化 `002-codegen-pipeline` 为 active：补齐 A3 触发的 AI shared vocabulary、跨语言 drift/parity 与本地 codegen-check 接入；F3 prompt bridge 与远端 CI drift detection 仅保留 future handoff。 | [002-codegen-pipeline](./plans/002-codegen-pipeline/plan.md) |
| 2026-04-29 | 1.5 | 按 ADR-Q6 authoritative 边界补齐 AI shared vocabulary：B1 只拥有 `AI_*` 错误码与 Model Profile / AI meta 字段名常量或生成类型；A3 继续拥有 Model Profile schema、`AIClient` runtime、`AICallMeta` runtime 与 provider adapter，A4/E4 负责连接参数与 endpoint。 | plan-review remediation |
| 2026-04-29 | 1.4 | 授权并落地 A3 AI provider baseline 错误码：`AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_FALLBACK_EXHAUSTED`，作为 `shared/conventions.yaml` 与 Go / TS / OpenAPI codegen 共同消费的唯一真理源；`AICallMeta` 运行时结构仍由 A3 拥有，不进入 B1 共享 DTO。 | ai-provider-and-model-routing spec remediation |
| 2026-04-28 | 1.3 | 明确 `ApiError` 为错误响应 envelope 内部对象，Go canonical 类型继续归属 `backend/internal/shared/errors.APIError`，B2 OpenAPI 负责外层 `ApiErrorResponse` envelope。 | openapi-v1-contract/001-bootstrap assessment remediation |
| 2026-04-27 | 1.2 | 对齐 A5 单人开发阶段决策：B1 只要求本地 lint/codegen 质量门禁，远端 CI / PR required check / CI drift detection 不作为当前 P0 前置，未来触发条件成立后再由 A5 重新评估。 | 001-bootstrap |
| 2026-04-27 | 1.1 | 回写 001-bootstrap 复盘确认的文档漂移：明确 13 个枚举 source section / 14 个生成枚举类型、Go `APIError` 手写归属、TS toolchain 与 Go/TS idempotency 双端验收落点。 | 001-bootstrap |
| 2026-04-26 | 1.0 | 初始创建：锁定跨语言真理源 `shared/conventions.yaml`、Go module 名称（`github.com/monshunter/easyinterview/backend`）、pnpm workspace、UUIDv7 / tmp_ id 规则、错误码 `UPPER_SNAKE_CASE` lint、枚举 `lower_snake_case` 双向生成；定义 13 个枚举 source section 与 6 个 baseline 错误码示例。 | 001-bootstrap |
