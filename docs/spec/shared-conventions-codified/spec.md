# Shared Conventions Codified Spec

> **版本**: 1.30
> **状态**: active
> **更新日期**: 2026-07-12

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 B1 `shared-conventions-codified` 定义为当前 active Contract spec（依赖 [A1 `repo-scaffold`](../repo-scaffold/spec.md)）。它是最早落地的基础契约 spec 之一，决定了：

- 当前共享命名 / ID / 时间 / 错误码 / 枚举 / 分页 / error envelope 以 `shared/conventions.yaml` 与本 spec 为准；
- 后端 Go 与前端 TypeScript 在没有 OpenAPI codegen（B2）之前已经能共享的最小类型集合；
- 后续 child（B2 `openapi-v1-contract` / C 全域 / D 全域）在自己的 plan 中只能引用本 spec 已锁定的 enum / error code / id 工具，不允许私造同义字符串。

目标是：

1. **真理源即代码**：把 product-scope / UI scope 确认的 11 个生成枚举类型、当前错误码、AI shared vocabulary、ID 规则、时间规则和金额规则同时落到 Go 与 TypeScript。
2. **跨语言对齐**：Go 与 TS 类型必须共用同一份枚举 / 错误码源（YAML 或 JSON），由本 spec 唯一的 generator 在两侧吐出代码。
3. **lint 强约束**：`UPPER_SNAKE_CASE` 错误码、`lower_snake_case` 枚举值、`camelCase` JSON tag 通过本地 lint 门禁拦截，而不是依赖代码 review。
4. **monorepo 名称锁定**：在落地任何业务代码前，先把 `go.mod` 名称、`package.json` 名称、pnpm workspace（如启用）拓扑、共享 lib 目录定下来，避免后续多个 subject 各自重命名雪球。

本 spec 不实现 OpenAPI 契约（归 B2）、不写业务 handler、不接入数据库（归 B4 与各 C 域）。

## 2 范围

### 2.1 In Scope

- 真理源文件 `shared/conventions.yaml`（或等价 JSON）：包含全部枚举、错误码、ID 前缀、时间格式常量、API 包装结构、异步 Job 状态。
- 跨语言 generator：从 `shared/conventions.yaml` 生成 `backend/internal/shared/types/*.go` 与 `frontend/src/lib/conventions/*.ts`。
- Go 共享 module：`backend/internal/shared/types/`、`backend/internal/shared/idx/`（UUIDv7 + tmp_ id 工具）、`backend/internal/shared/errors/`（错误码常量与 `APIError` 类型）。
- TS 共享 lib：`frontend/src/lib/conventions/`（`PageInfo` / `ApiError` / 枚举字面量类型）、`frontend/src/lib/ids/`（UUID 字符串工具与 tmp_ 前缀校验）。
- monorepo 名称与拓扑：根 `go.work` 只 use `./backend`，`backend/go.mod` module name 为 `github.com/monshunter/easyinterview/backend`，frontend package name 为 `@easyinterview/frontend`，并启用根 `pnpm-workspace.yaml`。
- Lint 规则：`UPPER_SNAKE_CASE` 错误码常量名、`lower_snake_case` 枚举字面量、`camelCase` JSON tag；B1 提供本地可执行的最小校验，A5 只约束本地质量门禁与远端 CI 延后边界。
- Idempotency-Key 工具：Go 与 TS 双端的 24h TTL 校验 / 生成工具骨架。
- AI 共享 vocabulary：AI capability、provider registry 字段名、Model Profile 字段名、`AICallMeta`/GenerationProvenance/`ai_task_runs` 共同消费的 AI meta 字段名常量或生成类型；B1 不实现 `AIClient`、不拥有 `AICallMeta` runtime 结构体，也不定义 `AI_PROVIDER_*` 连接参数语义。
- AI vocabulary 生成落点独立于错误码：Go 侧输出到 `backend/internal/shared/ai/`（或同等 B1-owned AI vocabulary 包），TS 侧输出到 `frontend/src/lib/conventions/ai.ts`（或同等文件）；不得把 model profile / AI meta 字段名塞进 `errors/*`。

### 2.2 Out of Scope

- OpenAPI 契约本身：归 [B2 `openapi-v1-contract`](../engineering-roadmap/spec.md#51-当前已存在的-active-spec)。
- 事件 envelope / outbox schema：归 B3 `event-and-outbox-contract`。
- DB 表与 migration：归 B4 `db-migrations-baseline`。
- 远端 CI 把上述 lint / generator 接到 PR 阶段：当前单人阶段不做；触发多人协作 / 公开 release / 自动发版等条件后再由 A5 `ci-pipeline-baseline` 重新评估。
- prompt / rubric / model 版本表与 LLM Judge：归 F3 `prompt-rubric-registry`。
- `AIClient` runtime、Model Profile schema / loader、provider adapter、fallback 消费与 `AI_PROVIDER_*` 连接参数校验：归 A3 / A4 / E4，B1 只提供字段名和错误码真理源。
- 业务域 handler / store / backend background runner（auth / upload / practice / review …）：归 C1–C8 或对应 runtime owner。

## 3 用户决策

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 跨语言真理源 | `shared/conventions.yaml`（YAML），由 generator 同时输出 Go / TS | 任何枚举或错误码新增必须改一处源；不允许只改 Go 或只改 TS |
| D-2 | Go workspace / module | 根 `go.work` 只 use `./backend`；module 名称为 `github.com/monshunter/easyinterview/backend`（落点 `backend/go.mod`）；两者 `go` directive 与 `.tool-versions` 同为 `1.24.5` | 后续所有 Go 包必须以此为根；不允许另起 module 或产生版本漂移 |
| D-3 | TS 包管理 | pnpm workspace（启用 `pnpm-workspace.yaml`），前端 package 名 `@easyinterview/frontend` | A2 `local-dev-stack` 与 B2 `openapi-v1-contract` 默认沿用 |
| D-4 | UUID 算法 | UUIDv7（含时序）；前端临时 id 使用 `tmp_<uuidv4>` | 所有业务主键由 idx 工具生成；不允许 NewV4 直接用作 DB id |
| D-5 | 错误码命名 | `UPPER_SNAKE_CASE`，前缀按 domain：`AUTH_*` / `TARGET_*` / `PRACTICE_*` / `REPORT_*` / `RESUME_*` / `PRIVACY_*` / `AI_*` / `RATE_LIMITED` / `VALIDATION_FAILED` / `RESOURCE_NOT_FOUND` | 任何非前缀错误码必须由本 spec 修订决定；business code 直接 import 常量；A3 已授权 `AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_FALLBACK_EXHAUSTED` / `AI_UNSUPPORTED_CAPABILITY` / `AI_PROVIDER_CONFIG_INVALID` / `AI_PROVIDER_SECRET_MISSING`；C4 已授权 `TARGET_JOB_NOT_FOUND` / `TARGET_IMPORT_SOURCE_INVALID` / `TARGET_IMPORT_SOURCE_UNAVAILABLE` / `TARGET_INVALID_STATE_TRANSITION`；backend-practice/001 Phase 0 已授权 `PRACTICE_PLAN_NOT_FOUND` / `PRACTICE_SESSION_NOT_FOUND`；backend-review/001 已授权 non-retryable `REPORT_CONTEXT_TOO_LARGE`；`RESOURCE_NOT_FOUND` 是当前 cross-resource generic 404 |
| D-6 | 枚举值书写 | `lower_snake_case`；TS 用 union string literal，Go 用 named string + 常量集 | 覆盖 `shared/conventions.yaml` 当前 11 个生成枚举类型；增删枚举必须先修订 owner spec |
| D-7 | `ApiError` inner object 归属 | `shared/conventions.yaml#structures.ApiError` 表示错误响应 envelope 内部的 `error` 对象（`code` / `message` / `requestId` / `retryable` / `details`），不表示外层 `{error: ...}` envelope；Go 侧 canonical 类型是手写 `backend/internal/shared/errors.APIError` + generated `errors.AllCodes`，TS 侧 canonical 类型是 generated `frontend/src/lib/conventions.ApiError` | B2 OpenAPI 必须把 wire response body 建模为 `ApiErrorResponse` envelope，并在 envelope 内 `$ref` B1 `ApiError` inner object；不得把 Go 侧误写为 `sharedtypes.ApiError` |
| D-8 | AI shared vocabulary 归属 | B1 提供 `AI_*` 错误码、AI capability、Provider Registry 字段名、Model Profile 字段名、AI meta 字段名常量或生成类型；A3 提供 Model Profile schema、`AIClient` runtime、`AICallMeta` runtime 填充与 OpenAI-compatible provider adapter；A4 校验 `AI_PROVIDER_*` 连接参数 | 避免 B1/A3/B4/F1 对同一 AI 字段私造名称；同时避免把运行时或连接配置误下沉到 shared conventions |
| D-9 | 当前 Practice / Report enum | 只保留 `PracticeGoal = baseline / retry_current_round / next_round`；删除 `PracticeMode` 与 `QuestionReviewStatus` | 连续 conversation 不存在 strict/assisted 或 question review 生命周期 |
| D-10 | Resume contract shared vocabulary | 当前 B1 仅保留 Resume shared error code `RESUME_EXPORT_NOT_AVAILABLE`，用于 B2 `exportResume` P0 `501` 响应；其它 Resume request/response schema、provenance 与 fixture contract 由 B2 OpenAPI 承接 | 业务代码 (`backend-resume` / `frontend-resume-workshop`) 不得绕过 B1 私造错误码；B2 schema 必须通过 `$ref` 引用本 spec 锁定的错误码字面量 |
| D-11 | Generator 实现 | `backend/cmd/codegen/conventions` 使用 `yaml.v3` 将 `shared/conventions.yaml` 解码为 typed `Spec`，再由手写 Go renderer 与 `go/format` 生成当前 Go/TS 资产 | generator 保持单入口；输出完整性和幂等性由 focused tests 与 `make codegen-check` 固化 |
| D-12 | TypeScript 生成边界 | conventions 输出固定为 `enums.ts`、`errors.ts`、`ai.ts`、`pagination.ts`，ID 输出固定为 `frontend/src/lib/ids/generated.ts` | 消费方从 conventions barrel 或 IDs 模块引用；不得把已拆分资产重新合并为单文件或复制出第二套生成入口 |
| D-13 | Report context size error | `REPORT_CONTEXT_TOO_LARGE` 是 B1 canonical report error code，message 为 `report context exceeds supported generation size`，`retryable: false`。它由 `shared/conventions.yaml` 单源生成 Go/TS constants，并以 string enum additive 同步到 B2 `ApiErrorCode`。 | backend-review 在最终 UTF-8 prompt 超过 48,000 bytes 时 terminal fail 且不调用 provider/repair；B1 不拥有该业务阈值，只拥有跨语言字面量与 retryability |
| D-20 | 扁平 Resume vocabulary boundary | Resume 是单一实体，API path / request / response 使用 `resumeId` 与 `Resume`；UI resume ≡ OpenAPI `Resume` | 由 [openapi-v1-contract/004](../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) 与 [backend-resume](../backend-resume/spec.md) 同步 `shared/conventions.yaml`、Go/TS generated errors、B2 `ApiErrorCode` 与 parity fixtures；新增 Resume shared vocabulary 必须先修订 owner spec |

## 4 设计约束

### 4.1 真理源约束

- `shared/conventions.yaml` 是当前共享枚举、错误码、ID、分页、错误 envelope 和 AI shared vocabulary 的可执行真理源。任何 enum / error code / job status 新增必须先修订本 spec（或对应 owner spec），再同步到 YAML 和生成代码；不得绕过 B1 直接修改生成物。
- generator 必须保持 idempotent：同一份 YAML 多次生成产出完全一致的 Go / TS 文件；当前通过本地 `make codegen-check` 或 `git diff --exit-code` 校验未漂移。

### 4.2 命名约束

- 错误码常量在 Go / TS 两侧都必须 `UPPER_SNAKE_CASE`，并以包级常量暴露；TS 侧使用 `as const` 字面量映射，避免 string union 散落。
- 枚举值在 JSON / API / 日志中统一 `lower_snake_case`；Go 类型名 `PascalCase`，常量名 `<TypeName><Value>`（例：`PracticeGoalBaseline`）。
- `tmp_` 前缀只用于前端浏览器内临时 id；Go 端不得接受任何带 `tmp_` 前缀的字段写入正式业务表，必须在 idx 工具的 `RequireServerID(...)` 校验中拒绝。

### 4.3 边界约束

- 根 `go.work` 只 use `./backend`，并与 `backend/go.mod`、`.tool-versions` 使用同一 Go 版本；本 spec 输出的 Go module 路径 `backend/internal/shared/...` 不得被任何 child 重命名，后续 child 只能在 `internal/<domain>/` 中 import 这些 shared 类型。
- TS 共享 lib 路径 `frontend/src/lib/conventions/` 与 `frontend/src/lib/ids/` 不得被后续 frontend workstream 重命名；可由 `frontend-shell` 在自己的 plan 中扩展 path alias。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| 跨语言真理源（YAML） | B1 | 单一源 + generator |
| Go 共享类型 | B1 | `backend/internal/shared/{types,errors,idx}/` |
| TS 共享类型 | B1 | `frontend/src/lib/{conventions,ids}/` |
| Go module 拓扑 | B1 + A1 | A1 锁工具版本与 `backend/` 根；B1 锁根 `go.work`、`backend/go.mod` 名称和单 module 拓扑 |
| pnpm workspace | B1 + A2 | B1 锁名称 + workspace.yaml；A2 在 dev stack 中保证可装 |
| OpenAPI / fixtures | B2 | 引用 B1 的枚举与错误码常量 |
| 事件 envelope | B3 | 引用 B1 的 `eventName` 命名约束、`eventVersion` 字段 |
| AI shared vocabulary | B1 + A3 | B1 输出字段名 / 错误码常量；A3 owns Model Profile schema、`AIClient` runtime、`AICallMeta` runtime 与 `AI_PROVIDER_*` 连接参数消费 |
| 本地质量门禁 / 未来远端 CI | B1 + A5 | B1 提供本地 lint/config；A5 只记录本地质量门禁与远端 CI 延后条件 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 真理源生成 Go 类型 | YAML 已更新 | 执行 codegen | Go 中 11 个枚举、PageInfo 和错误码生成；PracticeMode/QuestionReviewStatus 零残留 | 001-bootstrap |
| C-2 | 真理源生成 TS 类型 | 同 C-1 | 同 C-1 | TS 中 11 个 union、ApiError/PageInfo/error constants 生成；typecheck 通过 | 001-bootstrap |
| C-3 | UUIDv7 工具可用 | A1 已落地仓库根 | 在 Go test 与 TS test 中调用 idx 工具 | Go `idx.NewID()` / TS `newId()` 返回 UUIDv7 字符串；输入 `tmp_xxx` 时 `idx.RequireServerID()` / `requireServerId()` 抛错 | 001-bootstrap |
| C-4 | Idempotency-Key 工具 | A1 已落地仓库根 | 生成 + 校验 idempotency key（24h TTL） | Go 与 TS 双端工具产出格式一致的 key；TTL 过期后校验返回 false | 001-bootstrap |
| C-5 | Lint 拦截违规命名 | 本地提交前引入一个 `auth_unauthorized`（小写）错误码常量 | 跑 `make lint` | B1 本地 lint/config 能报错：错误码必须 `UPPER_SNAKE_CASE`；A5 只约束本地质量门禁与远端 CI 延后边界，不改变规则语义 | 001-bootstrap |
| C-6 | OpenAPI codegen 复用 B1 | B2 在自己 plan 里生成 OpenAPI types | B2 codegen 完成 | 任何枚举字段直接 import B1 的常量；不出现重复定义 enum 字面量 | B2 自身 plan |
| C-7 | OpenAPI 错误响应 envelope 复用 B1 inner error | B2 渲染 `components.schemas.ApiError` 与 `components.schemas.ApiErrorResponse` | `make codegen-openapi && make codegen-check` | `ApiError` 只包含 inner error 字段；`ApiErrorResponse.error` `$ref` 到 `ApiError`；Go generated 复用 `sharederrors.APIError`，TS generated 复用 `conventions.ApiError` | openapi-v1-contract/001-bootstrap |
| C-8 | AI vocabulary 共享 | A3/B4/F1/B2/TS client 同时消费 AI capability、provider/profile 字段、AI meta 字段与 `AI_*` 错误码 | `make codegen-conventions && make codegen-openapi`，再跑 parity tests / drift gate | `chat/stt/realtime/judge`、provider registry 字段、Model Profile 字段、`model_profile_name` / `capability` / fallback label 等字段名由 B1 生成或校验；F3 prompt/rubric provenance 字段（`feature_key` / `feature_flag` / `data_source_version`）作为 AI vocabulary 的一部分由 [F3 `prompt-rubric-registry/001-baseline`](../prompt-rubric-registry/plans/001-baseline/plan.md) 阶段 4.1 登记，仅服务于 prompt/rubric 来源追溯，不进入 F1 metric label 集合；A3 `AICallMeta` runtime 与 B4 `ai_task_runs` typed columns 使用同一来源；B1 不生成 `AICallMeta` DTO | ai-provider-and-model-routing/003 Phase 6 + db-migrations-baseline remediation + F3 `prompt-rubric-registry/001-baseline` 阶段 4.1 |
| C-9 | TargetJob 场景错误码共享 | C4 `backend-targetjob` 需要区分不存在/越权、非法导入源、暂时不可用导入源、非法状态迁移 | `make codegen-conventions && make codegen-openapi`，再跑 B1/B2 parity tests / drift gate | `TARGET_JOB_NOT_FOUND` / `TARGET_IMPORT_SOURCE_INVALID` / `TARGET_IMPORT_SOURCE_UNAVAILABLE` / `TARGET_INVALID_STATE_TRANSITION` 出现在 `shared/conventions.yaml`、Go/TS generated 错误码和 OpenAPI `ApiErrorCode` enum；handler 只使用上述 canonical codes，不私造 bare aliases | backend-targetjob/001 Phase 0 |
| C-10 | Go workspace/module metadata | 根 workspace 与 backend module 已落地 | 执行 `make lint-go-mod-tidy` | `.tool-versions`、`go.work`、`backend/go.mod` 的 Go 版本均为 `1.24.5`；workspace 只 use backend；tidy 无 diff | 001-bootstrap |
| C-11 | Report oversized-context error parity | backend-review 需要 deterministic terminal code | 更新 B1 truth source 并执行 conventions/OpenAPI parity | YAML、Go、TS 与 B2 `ApiErrorCode` 精确包含 `REPORT_CONTEXT_TOO_LARGE`；retryable 为 false；OpenAPI merge-base audit 只把该 enum widening 记为 additive，不进入 OPENAPI-001 breaking allowset | 001-bootstrap Phase 9 + openapi-v1-contract/003 |

## 7 关联计划

- [001-bootstrap](./plans/001-bootstrap/plan.md)：落地真理源 YAML、generator 框架、Go / TS 共享 lib 骨架、UUID / idempotency 工具、本地 lint gate、monorepo 名称（go.mod / pnpm workspace）。
- [002-codegen-pipeline](./plans/002-codegen-pipeline/plan.md)：当前 active；补齐 A3 触发的 AI shared vocabulary、跨语言 drift/parity gate 与本地 `make codegen-check` 接入。F3 prompt/rubric registry bridge 与远端 CI drift detection 只作为 handoff / future scope，不在 002 当前验收中实施。

后续如果 F3 需要共享 `feature_key + version` SDK，或 A5 D-5 触发远端 CI，再递增本 spec 版本并追加后续 plan；不把 F3 bridge / remote CI scope 塞回 002。

## 8 修订记录

完整修订记录见 [history.md](./history.md)。本 spec 只保留当前修订摘要。

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-12 | 1.30 | 授权 `REPORT_CONTEXT_TOO_LARGE`，锁定 non-retryable YAML/Go/TS/OpenAPI enum parity 与 additive-diff 口径。 | 001-bootstrap Phase 9 + backend-review/001 |
| 2026-07-10 | 1.27 | 将已经实现的 generator 与 TypeScript 输出边界固化为当前决策，删除陈旧待确认段落。 | tech-debt pruning |
| 2026-07-10 | 1.26 | 将当前 Go topology 固化为根 `go.work` 单 use backend，并统一三处 Go 版本与 tidy drift gate。 | tech-debt pruning |
| 2026-07-10 | 1.25 | 将 TargetJob 错误码验收条件收敛为 canonical codes 正向合同，并同步 history 与 001 context。 | tech-debt pruning |
| 2026-07-07 | 1.24 | docs-only：将修订说明统一为记录表述，不改变 B1 shared conventions truth source。 | product-scope/001-core-loop-module-pruning |
| 2026-07-06 | 1.23 | docs-only：将 B1 active spec 收敛为当前 16 个生成枚举、当前 flat Resume vocabulary 与 `RESUME_EXPORT_NOT_AVAILABLE` 错误码边界；详细修订明细只保留在独立 history。 | product-scope/001-core-loop-module-pruning |
