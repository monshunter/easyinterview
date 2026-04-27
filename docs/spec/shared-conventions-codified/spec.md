# Shared Conventions Codified Spec

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-04-27

## 1 背景与目标

[engineering-roadmap spec §5.2](../engineering-roadmap/spec.md#52-layer-b--contract4-份全部-p0) 把 B1 `shared-conventions-codified` 列为 Layer B · Contract 的入口 child（依赖 [A1 `repo-scaffold`](../repo-scaffold/spec.md)）。它是 Wave 0 必须落地的两份 spec 之一，决定了：

- [00-shared-conventions.md](../../../easyinterview-tech-docs/00-shared-conventions.md) 的命名 / ID / 时间 / 枚举 / 错误码 / 异步 Job 约定如何在代码层成为强约束；
- 后端 Go 与前端 TypeScript 在没有 OpenAPI codegen（B2）之前已经能共享的最小类型集合；
- 后续 child（B2 `openapi-v1-contract` / C 全域 / D 全域）在自己的 plan 中只能引用本 spec 已锁定的 enum / error code / id 工具，不允许私造同义字符串。

目标是：

1. **真理源即代码**：把 00-shared-conventions.md 中的 13 个 §5 小节 / 14 个生成枚举类型、6 个已记录错误码示例、ID 规则、时间规则、金额规则同时落到 Go（`backend/internal/shared/types/`）与 TypeScript（`frontend/src/lib/conventions/`）。
2. **跨语言对齐**：Go 与 TS 类型必须共用同一份枚举 / 错误码源（YAML 或 JSON），由本 spec 唯一的 generator 在两侧吐出代码。
3. **lint 强约束**：`UPPER_SNAKE_CASE` 错误码、`lower_snake_case` 枚举值、`camelCase` JSON tag 通过 lint 规则在 PR 阶段拦截，而不是依赖代码 review。
4. **monorepo 名称锁定**：在落地任何业务代码前，先把 `go.mod` 名称、`package.json` 名称、pnpm workspace（如启用）拓扑、共享 lib 目录定下来，避免 W2 多个 child 各自重命名雪球。

本 spec 不实现 OpenAPI 契约（归 B2）、不写业务 handler、不接入数据库（归 B4 与各 C 域）。

## 2 范围

### 2.1 In Scope

- 真理源文件 `shared/conventions.yaml`（或等价 JSON）：包含全部枚举、错误码、ID 前缀、时间格式常量、API 包装结构、异步 Job 状态。
- 跨语言 generator：从 `shared/conventions.yaml` 生成 `backend/internal/shared/types/*.go` 与 `frontend/src/lib/conventions/*.ts`。
- Go 共享 module：`backend/internal/shared/types/`、`backend/internal/shared/idx/`（UUIDv7 + tmp_ id 工具）、`backend/internal/shared/errors/`（错误码常量与 `APIError` 类型）。
- TS 共享 lib：`frontend/src/lib/conventions/`（`PageInfo` / `ApiError` / 枚举字面量类型）、`frontend/src/lib/ids/`（UUID 字符串工具与 tmp_ 前缀校验）。
- monorepo 名称：`go.mod` module name（拟 `github.com/monshunter/easyinterview/backend`）、`frontend/package.json` name、可选 `pnpm-workspace.yaml`。
- Lint 规则：`UPPER_SNAKE_CASE` 错误码常量名、`lower_snake_case` 枚举字面量、`camelCase` JSON tag；B1 提供本地可执行的最小校验，A5 负责把这些校验接入 CI。
- Idempotency-Key 工具：Go 与 TS 双端的 24h TTL 校验 / 生成工具骨架。

### 2.2 Out of Scope

- OpenAPI 契约本身：归 [B2 `openapi-v1-contract`](../engineering-roadmap/spec.md#52-layer-b--contract4-份全部-p0)。
- 事件 envelope / outbox schema：归 B3 `event-and-outbox-contract`。
- DB 表与 migration：归 B4 `db-migrations-baseline`。
- CI 把上述 lint / generator 接到 PR 阶段：归 A5 `ci-pipeline-baseline`。
- prompt / rubric / model 版本表与 LLM Judge：归 F3 `prompt-rubric-registry`。
- 业务域 handler / store / worker（auth / upload / practice / review …）：归 C1–C8。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 跨语言真理源 | `shared/conventions.yaml`（YAML），由 generator 同时输出 Go / TS | 任何枚举或错误码新增必须改一处源；不允许只改 Go 或只改 TS |
| D-2 | Go module 名称 | `github.com/monshunter/easyinterview/backend`（落点 `backend/go.mod`） | 后续所有 Go 包必须以此为根；不允许另起 module |
| D-3 | TS 包管理 | pnpm workspace（启用 `pnpm-workspace.yaml`），前端 package 名 `@easyinterview/frontend` | A2 `local-dev-stack` 与 B2 `openapi-v1-contract` 默认沿用 |
| D-4 | UUID 算法 | UUIDv7（含时序），与 [00-shared-conventions.md §2.2](../../../easyinterview-tech-docs/00-shared-conventions.md#22-id-规则) 对齐；前端临时 id 使用 `tmp_<uuidv4>` | 所有业务主键由 idx 工具生成；不允许 NewV4 直接用作 DB id |
| D-5 | 错误码命名 | `UPPER_SNAKE_CASE`，前缀按 domain：`AUTH_*` / `TARGET_*` / `PRACTICE_*` / `REPORT_*` / `RESUME_*` / `PRIVACY_*` / `RATE_LIMITED` / `VALIDATION_FAILED` | 任何非前缀错误码必须由本 spec 修订决定；business code 直接 import 常量 |
| D-6 | 枚举值书写 | `lower_snake_case`；TS 用 union string literal，Go 用 named string + 常量集 | 严格覆盖 00-shared-conventions §5 的 13 个小节；§5.13 同时包含隐私请求 type/status 两个并行字段，因此生成 14 个枚举类型 |

### 3.2 待确认事项

- generator 工具选型：默认手写 Go template + 简单 YAML loader；如执行阶段发现 schema 复杂度提升，可改用 `cuelang` 或 `quicktype`，由 001-bootstrap plan 在执行时升级并回填 D-1。
- `frontend/src/lib/conventions/` 是否进一步拆为 `enums.ts` / `errors.ts` / `pagination.ts`：默认拆，具体粒度由 002-codegen-pipeline plan 在 W1 末决定。

## 4 设计约束

### 4.1 真理源约束

- `shared/conventions.yaml` 是 [00-shared-conventions.md](../../../easyinterview-tech-docs/00-shared-conventions.md) 在代码侧的唯一镜像；任何 enum / error code / job status 新增必须先改 00 文档（或在 spec history 中说明授权扩展），再由本 spec 同步到 YAML。
- generator 必须保持 idempotent：同一份 YAML 多次生成产出完全一致的 Go / TS 文件；CI 中通过 `git diff --exit-code` 校验未漂移。

### 4.2 命名约束

- 错误码常量在 Go / TS 两侧都必须 `UPPER_SNAKE_CASE`，并以包级常量暴露；TS 侧使用 `as const` 字面量映射，避免 string union 散落。
- 枚举值在 JSON / API / 日志中统一 `lower_snake_case`；Go 类型名 `PascalCase`，常量名 `<TypeName><Value>`（例：`PracticeModeCoreInterview`）。
- `tmp_` 前缀只用于前端浏览器内临时 id；Go 端不得接受任何带 `tmp_` 前缀的字段写入正式业务表，必须在 idx 工具的 `RequireServerID(...)` 校验中拒绝。

### 4.3 边界约束

- 本 spec 输出的 Go module 路径 `backend/internal/shared/...` 不得被任何 child 重命名；后续 child 只能在 `internal/<domain>/` 中 import 这些 shared 类型。
- TS 共享 lib 路径 `frontend/src/lib/conventions/` 与 `frontend/src/lib/ids/` 不得被前端 child（D1–D7）重命名；可由 D1 `frontend-shell` 在自己的 plan 中扩展 path alias。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| 跨语言真理源（YAML） | B1 | 单一源 + generator |
| Go 共享类型 | B1 | `backend/internal/shared/{types,errors,idx}/` |
| TS 共享类型 | B1 | `frontend/src/lib/{conventions,ids}/` |
| Go module 拓扑 | B1 + A1 | A1 创建 `backend/` 根，B1 落地 `go.mod` 名称 |
| pnpm workspace | B1 + A2 | B1 锁名称 + workspace.yaml；A2 在 dev stack 中保证可装 |
| OpenAPI / fixtures | B2 | 引用 B1 的枚举与错误码常量 |
| 事件 envelope | B3 | 引用 B1 的 `eventName` 命名约束、`eventVersion` 字段 |
| Lint 接入 CI | A5 | 把 B1 提供的本地 lint/config 接到 PR pipeline |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 真理源生成 Go 类型 | `shared/conventions.yaml` 已落地 | 执行 `make codegen-conventions`（B1 持有） | `backend/internal/shared/types/*.go` 中 14 个枚举类型常量、`PageInfo` 结构与共享常量按 D-5 / D-6 命名生成；Go `APIError` 结构在 `backend/internal/shared/errors/` 手写，generator 仅补齐错误码常量；`go vet ./backend/...` 通过 | 001-bootstrap |
| C-2 | 真理源生成 TS 类型 | 同 C-1 | 同 C-1 | `frontend/src/lib/conventions/*.ts` 中 14 个 union string literal 类型、`ApiError` / `PageInfo` interface 按 D-6 生成；`pnpm tsc --noEmit` 通过 | 001-bootstrap |
| C-3 | UUIDv7 工具可用 | A1 已落地仓库根 | 在 Go test 与 TS test 中调用 idx 工具 | Go `idx.NewID()` / TS `newId()` 返回 UUIDv7 字符串；输入 `tmp_xxx` 时 `idx.RequireServerID()` / `requireServerId()` 抛错 | 001-bootstrap |
| C-4 | Idempotency-Key 工具 | A1 已落地仓库根 | 生成 + 校验 idempotency key（24h TTL） | Go 与 TS 双端工具产出格式一致的 key；TTL 过期后校验返回 false | 001-bootstrap |
| C-5 | Lint 拦截违规命名 | 任意 PR 中提交一个 `auth_unauthorized`（小写）错误码常量 | 跑 `make lint` | B1 本地 lint/config 能报错：错误码必须 `UPPER_SNAKE_CASE`；A5 只负责 CI 接入，不改变规则语义 | 001-bootstrap |
| C-6 | OpenAPI codegen 复用 B1 | B2 在自己 plan 里生成 OpenAPI types | B2 codegen 完成 | 任何枚举字段直接 import B1 的常量；不出现重复定义 enum 字面量 | engineering-roadmap/001 Phase 3 + B2 自身 plan |

## 7 关联计划

- [001-bootstrap](./plans/001-bootstrap/plan.md)：W0 落地真理源 YAML、generator 框架、Go / TS 共享 lib 骨架、UUID / idempotency 工具、本地 lint gate、monorepo 名称（go.mod / pnpm workspace）。

后续可在本 spec 修订递增版本后追加 `002-codegen-pipeline` 等 plan（覆盖 generator 在 CI 中的 drift detection、prompt / rubric registry 接入、跨语言测试），由 W1 阶段决定是否升格；本 spec 不预先 spawn 第二份 plan。

## 8 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-27 | 1.1 | 回写 `001-bootstrap` 交付复盘确认的 spec-plan 漂移：明确 13 个上游枚举小节对应 14 个生成类型、Go `APIError` 为手写 errors 包类型、TS `ApiError` / `PageInfo` 由 generator 生成，并保持 C-4 Go/TS idempotency 双端验收语义。 | 001-bootstrap |
