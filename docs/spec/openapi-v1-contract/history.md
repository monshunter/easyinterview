# OpenAPI v1 Contract History

> **版本**: 1.7
> **状态**: active
> **更新日期**: 2026-04-29

## 1 修订规则

[plan 003-breaking-change-gate](./plans/003-breaking-change-gate/plan.md) Phase 3.3 锁定的
`history.md` 写作规则。本节是 [openapi-v1-contract spec §4.4 审计要求](./spec.md#44-breaking-change-linter-规则集w1-末-freeze-后强制) 的执行口径，
与 `make openapi-diff` 白名单 gate（[scripts/lint/openapi_diff.py](../../../scripts/lint/openapi_diff.py)） 一一对应。

| 触发情形 | 必须递增 history 行？ | 行内必须显式标记 |
|----------|----------------------|------------------|
| 任何 schema / endpoint / response status / required 字段集合的变更（包括 additive 与 breaking） | 是 | 关联 plan id（如 `openapi-v1-contract/003-breaking-change-gate`） |
| privacy export 白名单切换：`POST /api/v1/privacy/exports` 从 `501` 切到 `202`（spec §3.1 D-12 / §4.4 P0 例外） | 是；缺则 `make openapi-diff` 升级为 breaking 退出码 1 | 标注「白名单 additive」+ 当前 spec / plan 版本号；`error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"` 的 fixture 同 PR 切换到 `PrivacyRequestWithJob` |
| 白名单外的 breaking change（删字段 / 改 type / required 新增 / 删 enum / 删 endpoint / path rename / method change） | 是 | **必须**引用 `状态: accepted` 的 ADR id（`OPENAPI-NNN-<short>`）；新 baseline 文件名 `openapi-v<MAJOR>.<MINOR>.<PATCH>.yaml`；spec 同 PR 修订 |
| fixture / example / 文案修订（v1.0.x patch） | 是 | 注明「fixture-only / docs-only」；不强制 ADR；不强制 baseline 递增 |
| 工具 / tooling 锁版变更（如 swagger-cli / Redocly / wrapper 版本） | 是 | 注明影响的 `make` target 与 [openapi/diff-config.yaml](../../../openapi/diff-config.yaml) `tooling` 段落 |

写作约定：

- 每行格式：`| <YYYY-MM-DD> | <spec 版本> | <一句话变更，含必要的 path / schema / status code 引用> | <关联 plan id 或 ADR id> |`
- 多行可共享同一 spec 版本号，作为该版本下的 sub-revision（参见现有 `1.3` 双行示例）
- 行顺序按时间倒序（最新置顶）
- 不删除历史行；纠正错误用追加新行的方式覆盖

## 2 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-29 | 1.7 | L2 remediation：`make openapi-diff` wrapper 补齐 `oneOf` / `allOf` / `anyOf` composition schema diff 的 breaking 检测，并把 privacy export `501→202` 白名单 history gate 默认基准改为 base branch merge-base；`Makefile` 新增 `HISTORY_REF=` 覆盖入口。 | 003-breaking-change-gate / plan-code-review --fix |
| 2026-04-28 | 1.6 | 落地 [plan 003-breaking-change-gate](./plans/003-breaking-change-gate/plan.md)：新增 `openapi/baseline/openapi-v1.0.0.yaml` v1.0.0 freeze 快照、`make openapi-diff` 本地 gate（[scripts/lint/openapi_diff.py](../../../scripts/lint/openapi_diff.py) wrapper + [openapi/diff-config.yaml](../../../openapi/diff-config.yaml) ruleset）、privacy export `501→202` 白名单 + 同 PR `history.md` 增量校验、ADR 模板 ([decisions/TEMPLATE.md](./decisions/TEMPLATE.md)) 与 SemVer 升级阈值文档 ([openapi/baseline/README.md](../../../openapi/baseline/README.md))，并在本 history.md 上方落地「修订规则」章节。 | 003-breaking-change-gate |
| 2026-04-28 | 1.5 | 将 `make docs-openapi` 的本地 HTML renderer 从 deprecated `redoc-cli@0.13.21` 迁移为官方推荐的 `@redocly/cli@2.30.1 build-docs`，保持 C-1 validator 仍由 `@apidevtools/swagger-cli@4.0.4` + inventory lint 负责。 | 001-bootstrap |
| 2026-04-28 | 1.4 | 根据 001-bootstrap assessment remediation 锁定 `ei_session` cookie 字面量引用、B2 tooling deprecated-but-accepted 边界、`ResourceType` / `JobType` 字面量集合，并明确 B1 `ApiError` inner object 与 B2 `ApiErrorResponse` envelope 的生成口径。 | 001-bootstrap |
| 2026-04-28 | 1.3 | `/design` 物化 §7 列出的 3 个 child plan：[001-bootstrap](./plans/001-bootstrap/plan.md)（openapi.yaml v1.0.0 骨架 + 双端 codegen + ADR-Q1 Auth + privacy export 501 + GenerationProvenance + 本地 drift gate）、[002-fixtures-and-mock-source](./plans/002-fixtures-and-mock-source/plan.md)（36 operationId default fixtures + prototype-baseline 同步工具 + 隐私脱敏校验 + Prism mock parity smoke）、[003-breaking-change-gate](./plans/003-breaking-change-gate/plan.md)（openapi-diff baseline + ruleset + privacy export 白名单 + ADR 模板 + B2 freeze handoff）。spec 文本不变，依旧是 v1.3。 | 001-bootstrap / 002-fixtures-and-mock-source / 003-breaking-change-gate |
| 2026-04-28 | 1.3 | 根据 L1 plan-review 修订 B2 契约：Auth tag 以 ADR-Q1 email magic link + first-party session cookie 为准；将 privacy export 的 P0 `501 Not Implemented` 作为显式状态码例外并标注 P1 切回 `202` 的兼容判定；补齐 schema inventory、header/idempotency 矩阵与 AI 生成结果 `GenerationProvenance` 约束。 | plan-review remediation |
| 2026-04-27 | 1.2 | 对齐 A5 单人开发阶段决策：B2 当前只要求本地 OpenAPI codegen drift / breaking-change gate，不要求 CI artifact、required check 或 label workflow。 | engineering-roadmap/001 Phase 3 remediation |
| 2026-04-27 | 1.1 | 修正 W1 gate 口径：parent Phase 3 只锁定 B2 v1.0.0 freeze 范围与 additive-only 规则；真实 `openapi/openapi.yaml`、codegen、fixtures 与 breaking-change linter 由 B2 child `001` 系列 plan 验证后再放行依赖 B2 的 W2 implementation | engineering-roadmap/001 Phase 3 remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定 `openapi/openapi.yaml` 唯一真理源、`/api/v1` 路径前缀、`camelCase` 字段命名、RFC3339 时间格式、共享 `ApiError` schema、cursor 分页统一、`Idempotency-Key` 与 `Job 202` 异步契约；§3.1.1 列出 v1.0.0 freeze 时的 36 个 endpoint × 14 tag 完整集合；锁定 breaking change linter 规则集（仅允许 additive）；引用 [02-api-definition.md](../../../easyinterview-tech-docs/02-api-definition.md) 全文与 [B1 D-5/D-6 枚举与错误码](../shared-conventions-codified/spec.md#31-已锁定决策)；记录 [ADR-Q5](../engineering-roadmap/decisions/ADR-Q5-privacy-cadence.md) `POST /privacy/exports` P0 返回 501 的例外。 | engineering-roadmap/001 Phase 3 |
