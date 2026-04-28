# OpenAPI v1 Contract Bootstrap Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-04-28

**关联计划**: [plan](./plan.md)

## Phase 1: openapi.yaml v1.0.0 骨架与共享 components

- [x] 1.1 落地 `openapi/openapi.yaml` 文档头（OpenAPI 3.1、`info.version: 1.0.0`、`servers: [{url: /api/v1}]`）+ 14 个 tag（按 spec §2.1 顺序）+ ADR-Q1 `sessionCookie` security scheme + document-level `security: [{sessionCookie: []}]`；不引入 `Authorization: Bearer` 默认 scheme
- [x] 1.2 在 `components` 中通过 `$ref` 引用 B1 `ApiError` inner object / `PageInfo` / 14 enum / 错误码 enum，并声明 B2 `ApiErrorResponse` envelope；声明 `Idempotency-Key` / `X-Request-ID` / `traceparent` / `Accept-Language` / `X-Client-Version` parameters / headers；落地 `Paginated<T>` `allOf` 模式与 `ResourceType` enum；落地 `GenerationProvenance` schema（6 字段，`rubricVersion` 允许 `not_applicable`）
- [x] 1.3 写入 spec §3.1.1 全部 36 个 operation：每个 operation 含 `tags` / `summary` / `operationId` / `security`（按 §4.1 public/protected 矩阵）/ 必要 parameters / request body（如有）/ success 或 P0 例外 response / `default: $ref ApiErrorResponse`；副作用 endpoint 引用 `Idempotency-Key`，但 ADR-Q1 `POST /api/v1/auth/email/start` 例外；`POST /practice/sessions/{sessionId}/events` 引用 `clientEventId` 且不挂 `Idempotency-Key`；长耗时 operation 走 `202 + *WithJob`；`POST /api/v1/privacy/exports` 唯一声明 `501` + `example.error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`；`GET /api/v1/runtime-config` security `[]`；AI 生成 schema 通过 `$ref` 关联 `GenerationProvenance`
- [x] 1.4 落地 `scripts/lint/openapi_inventory.py`（或等价 `make` 内联脚本）：断言 14 tag 顺序、36 operation 完整集合、每条 operation 都有 `default: $ref ApiErrorResponse`、`Idempotency-Key` 与 ADR-Q1 auth start / `clientEventId` 互斥规则、privacy export 501 唯一性

## Phase 2: Codegen pipeline

- [x] 2.1 落地 `backend/cmd/codegen/openapi/` Go generator（基于 `oapi-codegen` 或等价；模板放 `openapi/templates/go/`），输出 `backend/internal/api/generated/{types,server,spec}.gen.go`；B1 共享类型用 type alias 引用，不复制
- [x] 2.2 同一 generator 二进制额外渲染 TS：基于 `openapi-typescript` / `openapi-fetch`（或等价；模板放 `openapi/templates/ts/`），输出 `frontend/src/api/generated/{types,client,spec}.ts`；B1 TS 类型通过 import alias 引用，不复制
- [x] 2.3 根 `Makefile` 替换 `codegen` 占位行为 `codegen: codegen-conventions codegen-openapi`；新增 `make codegen-openapi`（idempotent）+ `make codegen-check`（跑 codegen 后 `git diff --exit-code -- backend/internal/api/generated frontend/src/api/generated openapi/openapi.yaml`）；`make help` 自动包含新 target
- [x] 2.4 在 `openapi/README.md` 与 `make codegen-openapi` 注释中明确：本地 drift gate 是当前唯一强制 gate；远端 CI required check 仅在 A5 触发条件成立后再接入；本 plan 不修改 A5 workflow

## Phase 3: API 文档站点（本地）

- [x] 3.1 落地 `make docs-openapi`：调用 Redoc / Stoplight CLI 渲染 `openapi/openapi.yaml` 为单文件 HTML，输出到 `openapi/dist/index.html`；根 `.gitignore` 忽略 `openapi/dist/`；不要求 A5 CI artifact
- [x] 3.2 更新 `openapi/README.md`：yaml 入口、generator 调用与产物落点、14 tag 链接、`make docs-openapi` 用法、B1 `$ref` 拓扑示意、`Authorization: Bearer` 不作为 P0 默认 auth 形态的声明

## Phase 4: Verification + handoff

- [x] 4.1 自检 spec C-1 / C-2 / C-3：`npx @apidevtools/swagger-cli@4.0.4 swagger-cli validate openapi/openapi.yaml` 通过；连续两次 `make codegen-openapi` 后 `git status` clean；删除任一 generated 文件可由 generator 还原；临时新增 schema 字段后 `make codegen-check` 失败、revert 后恢复
- [x] 4.2 自检 spec C-7 partial / C-8 / C-11 partial：privacy export 501 在 yaml 中唯一存在且 `example.error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`；分支上修改 B1 `shared/conventions.yaml` 任一 enum 值后 `codegen-conventions && codegen-openapi && codegen-check` 拦截漂移；`GenerationProvenance` schema 存在且 spec §4.6 固定 AI schema 名单均通过 `$ref` 可追溯
- [x] 4.3 文档与 INDEX 同步：本 plan 自身 checklist 与 Phase 4 验证完成后将 plan/checklist Header 切到 completed，并用 `/sync-doc-index --fix-index` 同步 plans/INDEX.md；不联动 002 / 003 状态；`/sync-doc-index --check` 通过；不修改 engineering-roadmap/001 Phase 3.3
- [x] 4.4 B2 child 协作 handoff：本 plan 输出的 `openapi.yaml` 与 generated 类型为 002 fixtures 校验与 003 baseline 锁定的真理源；不直接登记 spec C-10（C-10 由 003 Phase 4 关闭）

## Phase 5: Assessment remediation

- [x] 5.1 逐项核验 assessment R1-R6：确认 R1/R3/R4/R6 与 R2 的 `ApiError` inner/envelope drift 在当前仓库存在；R5 归类为低价值 no-op，不修改 `.agent-skills/tdd/SKILL.md`
- [x] 5.2 修订 ADR-Q1、A4/B1/B2 spec 与 `openapi/README.md`：锁定 `ei_session`、tooling deprecated-but-accepted 边界、`ResourceType` / `JobType` 字面量、`ApiError` inner object 与 `ApiErrorResponse` envelope 归属
- [x] 5.3 修订 OpenAPI codegen 并重新生成 artefacts：B1-AUTO `ApiError` 输出 inner object，新增 `ApiErrorResponse` envelope，Go/TS generated types 与 response schema 对齐
- [x] 5.4 运行 focused/regression 验证并收口生命周期：generator tests、`make lint-openapi`、backend build、frontend tsc、`/sync-doc-index --check` 通过；`make codegen-check` 的 validate/inventory 通过，最终 `git diff --exit-code` 因本次未提交的预期 generated 变更返回非 0，作为 dirty-tree 限制记录；随后恢复 plan/checklist `completed`

## Phase 6: docs-openapi renderer deprecation remediation

- [x] 6.1 复现 `make docs-openapi` 旧实现：命令 exit 0 且生成 `openapi/dist/index.html`，但打印 `redoc-cli` deprecated 横幅并提示 `@redocly/cli build-docs`
- [x] 6.2 迁移根 `Makefile` 的 `docs-openapi` target 到 `@redocly/cli@2.30.1 redocly build-docs`，保持输入、输出路径与标题不变；同步 `openapi/README.md` tooling 说明
- [x] 6.3 运行验证并收口生命周期：`make docs-openapi` 无 deprecated 横幅且产物成功生成；`make lint-openapi`、`/sync-doc-index --check`、`git diff --check` 通过；随后恢复 plan/checklist `completed`
