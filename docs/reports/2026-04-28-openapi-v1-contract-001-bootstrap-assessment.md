# openapi-v1-contract/001-bootstrap 交付复盘报告

> **日期**: 2026-04-28
> **审查人**: Claude Opus 4.7

## 1 复盘范围与成功证据

本复盘覆盖 [`openapi-v1-contract/001-bootstrap`](../spec/openapi-v1-contract/plans/001-bootstrap/plan.md) plan 的全部 4 个 phase（14/14 checklist items）。delivery 通过的具体证据：

- **C-1 (OpenAPI 文档结构)** — `npx @apidevtools/swagger-cli@4.0.4 swagger-cli validate openapi/openapi.yaml` exit 0；`scripts/lint/openapi_inventory.py` 输出 `14 tags, 36 operations, ApiError/IK/501/Provenance invariants enforced; B1 enums in sync.`。
- **C-2 / C-3 (Go + TS codegen drift)** — 连续两次 `make codegen-openapi` 后 `git status -- openapi/openapi.yaml backend/internal/api/generated frontend/src/api/generated` 为空；删除任一 generated 文件后 `make codegen-openapi` 还原为字节级一致；故意给 `PracticePlan` 加 `x-optional-metadata-test` 字段后 `make codegen-check` 失败并打印 diff，revert 后恢复 exit 0。
- **C-7 partial (privacy export 501)** — `endpoints declaring 501: [('post', '/privacy/exports')]`，example.error.code 为 `PRIVACY_EXPORT_NOT_AVAILABLE`。
- **C-8 (B1 → B2 drift propagation)** — 在 `shared/conventions.yaml` 注入 `MistakeStatus.values: test_drift_value` → `make codegen-conventions && make codegen-openapi` → `git diff --stat` 显示 openapi.yaml + backend/internal/api/generated/openapi.yaml + frontend/src/api/generated/spec.ts 三处漂移；`make codegen-check` exit 1。
- **C-11 partial (provenance 拓扑)** — 7 个 spec §4.6 AI schema (`TargetJobSummary` / `TargetJobFitSummary` / `AssistantAction` / `FeedbackReport` / `MistakeEntry` / `ResumeTailorRun` / `Debrief`) 全部通过 `$ref` 拓扑可达 `GenerationProvenance`。
- **TDD 单元测试** — `go test ./cmd/codegen/openapi/...` PASS（`TestRun_Idempotent` + `TestRun_DriftPropagatesFromConventions`）；`cd frontend && npx tsc --noEmit` PASS；`cd backend && go build ./...` PASS。
- **文档与 INDEX 同步** — plan/checklist Header `active` → `completed`；`/sync-doc-index --check` zero drift；002 / 003 active 状态保持不变；未触动 engineering-roadmap/001 Phase 3.3。
- **Branch 状态** — 4 个 phase commit 全部 FF-merged 回 `dev`：`40bacab feat(openapi): v1.0.0 骨架与共享 components` / `c4b5f1e feat(openapi): 落地 codegen pipeline 与本地 drift gate` / `29f5cd6 feat(openapi): 接入 docs-openapi 本地 HTML 站点` / `6e8a3cb docs(openapi): 收口 001-bootstrap plan 与 plans/INDEX`。

## 2 会话中的主要阻点/痛点

### 2.1 spec §4.6 列出的 `Session` cookie name 在 ADR-Q1 与 spec 中均未给出具体字面量

- **证据**：plan §1.1 写 “name 见 ADR-Q1”，但 ADR-Q1 §3 只描述属性（HttpOnly/SameSite=Lax/Secure/30d）未给 cookie 名；02-api-definition / db-definition / observability 等 grep 也无既有引用。落地时只能在 OpenAPI 中自选 `ei_session` 作为字面量。
- **影响**：单点判断；如果未来 C1 选择不同名字，OpenAPI security scheme 与 fixture 都需要修订。

### 2.2 `ApiError` 在 B1 Go 与 B1 TS 端的产出不对称

- **证据**：B1 generator (`backend/cmd/codegen/conventions/render.go`) 只在 TS 端生成 `export interface ApiError`，Go 端未输出 `sharedtypes.ApiError`。第一次 codegen-openapi 运行后 `cd backend && go build ./...` 报 `internal/api/generated/types.gen.go:34:29: undefined: sharedtypes.ApiError`，需要回退把 ApiError 从 B1 alias 列表里剔除并改为本地 struct（在 hand-authored openapi.yaml 的 components.schemas.ApiError 上 driven）。
- **影响**：第一次 build 失败；返回 render_go.go 修订 `b1OwnedSchemas` 列表。

### 2.3 OpenAPI 3.1 工具链选型存在分裂

- **证据**：plan §3.1 / §4.1 绑定 `npx @apidevtools/swagger-cli validate`，但该包在 npm registry 上已显示 `deprecated`（输出 `npm warn deprecated @apidevtools/swagger-cli@4.0.4`），上游推荐 `@redocly/cli`。试装 `@redocly/cli@1.25.6` 时遇到 `npm error network`（npm install 失败）。最终保留 `swagger-cli@4.0.4`（实测 OpenAPI 3.1 文档 `/tmp/test31.yaml is valid`）作为 spec C-1 验收命令；`make docs-openapi` 选 `redoc-cli@0.13.21`（同一族但不同包名）。
- **影响**：长期看每个外部依赖都有 deprecation / 替代，需要在 spec 修订时显式登记；当前会话只记录在 `openapi/README.md` 的 tooling 节。

### 2.4 spec §3.2 待确认事项 (`ResourceType` enum 字面量 / SSE 协议 / API 文档发布平台) 在 plan 落地时仍是开口

- **证据**：spec §3.2 标注 “默认独立成 schema，由 codegen 引导”，但 ResourceType 字面量集合从未被锁定；只能从 `02-api-definition.md` Job 示例反向枚举出 `target_job / feedback_report / resume_asset / resume_tailor_run / debrief / privacy_request` 6 个值。同样 JobType 只能从 jobType 字符串枚举推断 7 个值。
- **影响**：本 plan 锁定了 6/7 字面量集合，但任何后续 endpoint（如 mistake_entry 直接成为 Job 资源）仍需 spec 修订追加；建议把这种 “由 02-api-definition 反推” 的 enum 列表，作为下一次 spec 修订的明确锁定项。

### 2.5 Phase 2 Go generator 第一次运行碰到 `text/template` 不能在 template arg 处调用方法

- **证据**：第一次 `go run` 报 `Prefix has arguments but cannot be invoked as function`（`{{$.Prefix .ParentName .Value}}` 在 template 不可用）。回退把 const name 预算到 `goEnumValueView.ConstName`，并把 `Prefix` 闭包从 `goTypesData` 删除。
- **影响**：一次返工；如果 skill `tdd` 或 `implement` 给出更具体的 “generator 模板内不可调用 struct 方法 / 必须 precompute” 提示，未来类似 generator 项目可少踩。

### 2.6 OpenAPI 输出文件超出单次写入 token 上限两次

- **证据**：会话出现两次 “Output token limit hit. Resume directly” 系统提示；第一次发生在写 `openapi/openapi.yaml`（最终 ~2,800 行）时，第二次发生在写 codegen render 文件时。最终通过分批 `Edit`（domain schema / paths 分块追加）而非一次 `Write` 落地；正确恢复后未影响交付质量。
- **影响**：执行时间小幅膨胀；流程上没有损失；属于一次性的 token-budget 现象。

## 3 根因归类

- **2.1 cookie name 缺失** — 类别 `spec-plan`（ADR-Q1 + openapi-v1-contract spec 双向缺锁定字面量）。
- **2.2 ApiError Go/TS 不对称** — 类别 `spec-plan`（B1 [shared-conventions-codified spec](../spec/shared-conventions-codified/spec.md) §3.1 D-7 仅锁了 PageInfo Go 输出 / ApiError TS 输出，未对称要求 Go 同时落 ApiError struct）+ `no repo change needed`（本 plan 已通过本地 hand-authored ApiError struct 解决）。
- **2.3 OpenAPI 工具链 deprecation** — 类别 `spec-plan`（B2 spec / plan 直接绑定 `@apidevtools/swagger-cli`，未登记 “deprecated 但仍合规” 的事实，也未登记 redoc-cli 选型）。
- **2.4 ResourceType / JobType 字面量缺锁** — 类别 `spec-plan`（B2 spec §3.2 待确认事项需要在下次 spec 修订时锁定 enum 字面量并显式追加 `ai_task_runs.resource_type` 兼容范围）。
- **2.5 text/template Prefix 错误** — 类别 `no repo change needed`（一次性执行错误，不构成流程缺陷）。
- **2.6 token limit hit** — 类别 `no repo change needed`（环境 / runtime 现象，不属于 governance 资产问题）。

## 4 对流程资产的改进建议

| # | 建议 | 落点 | 优先级 |
|---|------|------|--------|
| R1 | 在 ADR-Q1 §3 第 2 条直接锁定 cookie 字面量名（候选 `ei_session`），并把它作为 `secrets-and-config/spec.md` §3.1.1 的关联说明，把双方 spec 对齐到同一字面量 | spec-plan（ADR-Q1 + secrets-and-config + openapi-v1-contract spec） | high |
| R2 | 在 [shared-conventions-codified/spec.md](../spec/shared-conventions-codified/spec.md) D-7（已有 `structures.ApiError` 锁定）增加对称要求：B1 Go generator 必须输出 `sharedtypes.ApiError` 与 `sharedtypes.AllErrorCodes` Go 类型，避免 Go/TS 端不对称 | spec-plan（shared-conventions-codified） | high |
| R3 | 在 [openapi-v1-contract/spec.md](../spec/openapi-v1-contract/spec.md) §2.1 / §3.1 增加 “tooling 锁定” 小节：明确 swagger-cli 已 deprecated 但 still-valid for OpenAPI 3.1，禁止换用 redocly/cli 直至本 spec 修订；docs-openapi 锁 `redoc-cli@0.13.21` | spec-plan（openapi-v1-contract） | medium |
| R4 | 在 [openapi-v1-contract/spec.md](../spec/openapi-v1-contract/spec.md) §3.2 把 `ResourceType` (6 值) 与 `JobType` (7 值) 字面量从待确认事项升级为已锁定决策；同步 `02-api-definition.md` 与 `03-db-definition.md#ai_task_runs.resource_type` 引用 | spec-plan（openapi-v1-contract） | medium |
| R5 | 在 [.agent-skills/tdd/SKILL.md](../../.agent-skills/tdd/SKILL.md) `## Test Completeness Requirements` 或 generator-related 章节增加 “大文件交付时优先 Edit 分块而非一次 Write”、“text/template 不可调用 struct 方法，需在 builder 阶段预算字段” 的简短提示，避免 generator 类任务首次运行返工 | skill（`/tdd`） | low |
| R6 | 在 `openapi/README.md` Tooling 节扩展 “每个外部 npm 依赖须显式锁版 + 加 ‘deprecated 但仍合规’ 注释” 的写作约定，作为后续 002 / 003 plan 在引入工具时的引用范本 | README（openapi/） | low |

## 5 建议优先级与后续动作

**最值得立即做**（阻断 / 阻碍下一轮）：

1. R1 — Cookie 名锁定。002 fixtures 与 C1 backend-auth 实现都会引用，越早锁越省回归。
2. R2 — B1 Go ApiError 对称落地。这是当前 002 fixtures 落 Go 端验证、以及未来 C-domain handler 实现的常踩坑；当前用了 hand-authored ApiError struct 绕过，但长期不优雅。

**可以延后**：

3. R3 / R6 — 工具锁定与文档约定，跟随 002 / 003 plan 修订时一起处理即可。
4. R4 — ResourceType / JobType 字面量锁定，下一次 02-api-definition 触动或新增 endpoint 时 batch 同步。
5. R5 — `/tdd` 提示，性价比相对低；可以在下次 generator 类 skill 使用频次累计后再考虑。

**无需后续动作**：

- 2.5 (text/template) — 一次性执行错误。
- 2.6 (token limit hit) — 环境现象，不进 governance 资产。

后续主路径仍是 `/implement openapi-v1-contract/002-fixtures-and-mock-source` → `/implement openapi-v1-contract/003-breaking-change-gate`，二者在 spec §7 已显式 spawn 为本 plan 的下游。
