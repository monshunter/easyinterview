# OpenAPI v1 Contract Bootstrap

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-04-29

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [openapi-v1-contract spec](../../spec.md) §3.1 已锁定的 D-1..D-15 与 §3.1.1 列出的 37 endpoint × 14 tag 落到 `openapi/openapi.yaml` v1.0.0；落地双端 codegen pipeline（Go DTO + chi server interface 在 `backend/internal/api/generated/`，TS SDK 在 `frontend/src/api/generated/`）；接入根 `Makefile` 的 `codegen-openapi` / `codegen-check` 入口；锁定 ADR-Q1 Auth 路径（email magic link + first-party session cookie）、`DELETE /api/v1/me` account deletion、`POST /api/v1/privacy/exports` P0 的 `501 PRIVACY_EXPORT_NOT_AVAILABLE` 例外、B1 共享 enum / 错误码 / `ApiError` 的 `$ref` 拓扑、§4.6 `GenerationProvenance` schema；通过本 plan Phase 4 的本地命令证明 spec §6 中 C-1 的 contract/schema 部分、C-2 / C-3 / C-8、C-7 / C-11 的 contract/schema 部分已成立，并为 [002-fixtures-and-mock-source](../002-fixtures-and-mock-source/plan.md) 与 [003-breaking-change-gate](../003-breaking-change-gate/plan.md) 提供可消费的契约源。

本 plan 不落地 fixtures（归 002）、不落地 breaking-change linter（归 003）、不实现业务 handler（归 C 域）、不实现前端 fetch 客户端（归 D 域）、不部署 mock server（归 [E1](../../../engineering-roadmap/spec.md#55-layer-e--integration4-份)）。后续如需扩展（v1.0.1 patch、新 endpoint、SSE 子协议接入），在本 spec / plan 上递增版本，原地修订。

## 2 背景

[engineering-roadmap §5.2](../../../engineering-roadmap/spec.md#52-layer-b--contract4-份全部-p0) 把 B2 列为 Layer B Contract 的 DAG 瓶颈节点：W2 启动的 C 全域和 D 全域 child 都依赖本 plan 产出的 codegen output 与 Auth security scheme。[001-decompose-subspecs Phase 3.3](../../../engineering-roadmap/plans/001-decompose-subspecs/checklist.md#phase-3-wave-1基础设施--契约骨架) 已将 spec-contract lock 标记完成；本 plan 是 §7 关联计划列出的 3 个 child 中第一个，承担 v1.0.0 freeze 的所有结构性产出。

执行本 plan 前必须确认 [B1 shared-conventions-codified/001-bootstrap](../../../shared-conventions-codified/plans/001-bootstrap/plan.md) 已完成：generator 输出的 `backend/internal/shared/types/enums.go`、`frontend/src/lib/conventions/{enums,errors,pagination}.ts`、`shared/conventions.yaml` 是本 plan 的 `$ref` 真理源；若 B1 未完成，先暂停本 plan。

每个 phase 是可独立验证的纵向切片：Phase 1 起来就能用 `npx @apidevtools/swagger-cli@4.0.4 swagger-cli validate` 校验骨架；Phase 2 起来就能 `make codegen-openapi` 双端生成；Phase 3 起来就能 `make codegen-check` 拦截漂移；Phase 4 收口 5 项 AC + 文档 + handoff。本 plan 不引入 BDD 资产（`test/scenarios/` 由 [E2 e2e-scenarios-p0](../../../engineering-roadmap/spec.md#55-layer-e--integration4-份) 在 W4 spawn），AC 验证完全由 `make` 命令与 `git diff --exit-code` 驱动。

## 3 实施步骤

### Phase 1: openapi.yaml v1.0.0 骨架与共享 components

#### 1.1 文档头与 servers / security schemes

在 `openapi/openapi.yaml` 写入 OpenAPI 3.1 文档头、`info.version: 1.0.0`、`servers: [{url: /api/v1}]`（spec D-1）、按 §2.1 顺序声明 14 个 tag（spec D-11）。security schemes 按 [ADR-Q1](../../../engineering-roadmap/decisions/ADR-Q1-auth.md) 写入 `sessionCookie`（type `apiKey`，in `cookie`，name `ei_session`）。`Authorization: Bearer` 不作为 P0 默认 security scheme；如需保留扩展点须在 ADR-Q1 + 本 spec 修订后再加。document-level `security: [{sessionCookie: []}]`，public endpoints（§4.1）在 operation 级别用 `security: []` 显式覆盖。

#### 1.2 共享 components 与 B1 `$ref`

在 `components.schemas` 中只声明 OpenAPI 范畴专属 schema；所有共享定义通过 `$ref` 指向 B1 generator 的产出文件或 `shared/conventions.yaml`：

- B1 `ApiError` inner object、`PageInfo`、14 个 enum、`error.code` 错误码 enum：`$ref` 到 `shared/conventions.yaml` 已存在的节，由 codegen template 在生成阶段把 yaml 节解释为 OpenAPI schema（B1 D-5 / D-7）。OpenAPI 不重复维护 B1 enum 字面量；wire body 另声明 `ApiErrorResponse` envelope（`{error: ApiError}`）。
- `Idempotency-Key` / `X-Request-ID` / `traceparent` / `Accept-Language` / `X-Client-Version` 在 `components.parameters` 与 `components.headers` 中声明一次，由 endpoint 通过 `$ref` 引用（spec §4.1）。
- `Paginated<T>`：使用 `allOf` + `pageInfo: $ref ../PageInfo` 模式（B1 D-5），不为每个列表 endpoint 单独维护字段顺序。
- `GenerationProvenance`（spec §4.6）：6 字段（`promptVersion` / `rubricVersion` / `modelId` / `language` / `featureFlag` / `dataSourceVersion`），其中 `rubricVersion` 显式允许 `not_applicable` 字面量。

公共 `ResourceType` enum（spec §3.2 待确认事项）独立成 schema，避免与 outbox / Job 引用重复。

#### 1.3 37 endpoint operation 骨架

按 spec §3.1.1 表格逐行写入 37 operation：

- 每个 operation 至少声明 `tags`、`summary`、`operationId`、`security`（覆盖 §4.1 public/protected 矩阵）、必要的 path/query/header parameters、request body（如有）、success 或 P0 例外 response、`default: $ref ApiErrorResponse`。
- `POST /api/v1/uploads/presign`、`POST /api/v1/resumes`、`POST /api/v1/targets/import`、`PATCH /api/v1/targets/{targetJobId}`、`POST /api/v1/practice/plans`、`POST /api/v1/practice/sessions`、`POST /api/v1/practice/sessions/{sessionId}/complete`、`POST /api/v1/mistakes/{mistakeId}/retest`、`POST /api/v1/resume/tailor`、`POST /api/v1/debriefs`、`POST /api/v1/privacy/exports`、`POST /api/v1/privacy/deletions` 等副作用 endpoint 必须声明 `Idempotency-Key` header 引用（spec D-6）；ADR-Q1 auth email start 例外见下一条。
- `POST /api/v1/practice/sessions/{sessionId}/events`：声明 `clientEventId` 字段，**不**挂 `Idempotency-Key` header；与其他幂等机制不混用（spec D-6）。
- `POST /api/v1/auth/email/start` 不挂通用 `Idempotency-Key`；rate limit / challenge TTL 归 ADR-Q1（spec §4.1）。
- 长耗时 operation（resume tailor / debrief / target import / practice complete / privacy delete / resume register 等）success response 走 `202 Accepted` + `*WithJob` schema（spec D-7）；客户端通过 `GET /api/v1/jobs/{jobId}` 轮询。
- `POST /api/v1/privacy/exports` P0 例外响应强制写为 `501` + `application/json: { schema: $ref ApiErrorResponse, example.error.code: "PRIVACY_EXPORT_NOT_AVAILABLE" }`（spec D-12 / §4.1 / C-7 partial）。
- `GET /api/v1/runtime-config` schema 引用 [A4 D-2](../../../secrets-and-config/spec.md#31-已锁定决策含-p0-必备-env-key-字典) 的 `RuntimeConfig`；security 设为空（public）。
- `DELETE /api/v1/me` schema 使用 `PrivacyRequestWithJob`，protected，必须声明 `Idempotency-Key` header 或等价 active-request dedupe 语义；operationId 固定 `deleteMe`。
- AI 生成结果 schema（`TargetJob.summary` / `TargetJob.fitSummary` / `AssistantAction` / `FeedbackReport` / 由 AI 创建的 `MistakeEntry` / `ResumeTailorRun` / `Debrief`）必须包含 `provenance: $ref GenerationProvenance` 字段，或所属 `*WithJob` 包装类型在 `job.provenance` 中可追溯到该对象（spec §4.6）。

#### 1.4 endpoint 自检

- `npx @apidevtools/swagger-cli@4.0.4 swagger-cli validate openapi/openapi.yaml` 通过（spec C-1）。
- 写一个 `scripts/lint/openapi_inventory.py`（或等价 `make` target 内联脚本）扫描 yaml，断言：
  - tag 数 == 14 且顺序与 spec §2.1 一致；
  - operation 数 == 37 且 `(tag, method, path, operationId)` 与 spec §3.1.1 完全一致；
  - 每个 operation 都有 `default: $ref ApiErrorResponse`；
  - 除 ADR-Q1 auth email start 与 session event 例外外，spec D-6 涉及的副作用 endpoint 都引用 `Idempotency-Key` header；`POST /api/v1/auth/email/start` 与 `POST /api/v1/practice/sessions/{sessionId}/events` 不引用；
  - `POST /api/v1/privacy/exports` 唯一声明 `501` 响应，`example.error.code == "PRIVACY_EXPORT_NOT_AVAILABLE"`。

### Phase 2: Codegen pipeline

#### 2.1 Go generator

落地 `backend/cmd/codegen/openapi/`（Go 实现，由 `make codegen-openapi` 调用）：

- 输入：`openapi/openapi.yaml` + `openapi/templates/go/`（可基于 [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) 模板配置；如直接复用上游模板，也必须放仓库内 templates 目录可追溯）。
- 输出：`backend/internal/api/generated/types.gen.go`（请求 / 响应 DTO，复用 B1 类型时通过 type alias）、`backend/internal/api/generated/server.gen.go`（chi server interface）、`backend/internal/api/generated/spec.gen.go`（embed 的 yaml）。
- B1 共享类型（`enums.PracticeMode` 等、`errors.Code` 等、`types.PageInfo`）必须以 type alias 引用，不复制类型。

#### 2.2 TS generator

同一二进制额外渲染 TS：

- 输入：同上 + `openapi/templates/ts/`（可基于 [openapi-typescript](https://github.com/openapi-ts/openapi-typescript) 或 [openapi-fetch](https://github.com/openapi-ts/openapi-fetch) 模板）。
- 输出：`frontend/src/api/generated/types.ts`（DTO 与 enum 联合类型）、`frontend/src/api/generated/client.ts`（fetch client + operationId-keyed methods）、`frontend/src/api/generated/spec.ts`（embed 的 yaml 字符串或 JSON）。
- 引用 B1 TS 类型（`@easyinterview/frontend/src/lib/conventions/{enums,errors,pagination}`）通过 import alias，不复制类型。

#### 2.3 Make 入口

根 `Makefile` 在 [B1 codegen-conventions](../../../shared-conventions-codified/plans/001-bootstrap/plan.md#13-写入-generator) 之后追加：

- `make codegen-openapi`：调用 generator；输出必须 idempotent；执行顺序为 `codegen-conventions` → `codegen-openapi`，确保 B1 类型已就绪。
- `make codegen-check`：运行 `codegen-openapi` 后跑 `git diff --exit-code -- backend/internal/api/generated frontend/src/api/generated openapi/openapi.yaml`；漂移即失败（spec C-2 / C-3）。
- 把 `make codegen` 当前 placeholder 行（`echo "TODO: OpenAPI codegen pending B2 openapi-v1-contract"`）替换为实际链路：`codegen: codegen-conventions codegen-openapi`。
- `make help` 自动包含新 target（沿用 `## ` 注释机制）。

#### 2.4 Drift gate 接入边界

本地 `make codegen-check` 是当前唯一强制 gate（spec §4.5 + §5）；远端 CI required check 仅在 [A5 ci-pipeline-baseline](../../../ci-pipeline-baseline/spec.md) 触发条件成立后再接入，本 plan 不修改 A5 workflow。

### Phase 3: API 文档站点（本地）

#### 3.1 `make docs-openapi`

落地 `make docs-openapi` 调用 Redoc / Stoplight CLI（或等价工具）把 `openapi/openapi.yaml` 渲染为单文件 HTML，输出到 `openapi/dist/index.html`，由根 `.gitignore` 忽略产物目录（spec §2.1 末项）。当前单人开发阶段只保留本地产物，不要求 A5 上传 CI artifact（spec §5）。

#### 3.2 `openapi/README.md`

更新 `openapi/README.md`：

- yaml 入口位置、generator 调用方式、产物落点（Go / TS）。
- 14 tag 列表与 spec §2.1 链接。
- `make docs-openapi` 使用方式与产物路径（不在 git 中）。
- 与 [B1 shared-conventions-codified](../../../shared-conventions-codified/spec.md) 的 `$ref` 拓扑示意。
- `Authorization: Bearer` 不作为 P0 默认 auth 形态的明确声明（与 [ADR-Q1](../../../engineering-roadmap/decisions/ADR-Q1-auth.md) 一致）。

### Phase 4: Verification + handoff

#### 4.1 Spec C-1 / C-2 / C-3 自检

- `npx @apidevtools/swagger-cli@4.0.4 swagger-cli validate openapi/openapi.yaml` exit 0。
- 跑两次 `make codegen-openapi`，第二次 `git status` 必须 clean；删除任意一个生成文件再跑可还原（codegen idempotency）。
- 在 `openapi/openapi.yaml` 给某个已有 schema 临时新增 `optional metadata` 字段（不提交）：`make codegen-check` 失败，diff 显示 generated 文件中新增字段；revert 后 gate 恢复 clean。

#### 4.2 Spec C-7 partial / C-8 / C-11 partial 自检

- 验证 `POST /api/v1/privacy/exports` 在 `openapi.yaml` 中唯一声明 `501` 响应且 `example.error.code == "PRIVACY_EXPORT_NOT_AVAILABLE"`；fixture 资产由 [002-fixtures-and-mock-source](../002-fixtures-and-mock-source/plan.md) 闭合。
- 修改 [B1 shared/conventions.yaml](../../../shared-conventions-codified/spec.md#31-已锁定决策) 中任一枚举值（在分支上验证后 revert）：跑 `make codegen-conventions && make codegen-openapi && make codegen-check`，B2 generated 与 OpenAPI yaml 必须同步漂移并被 gate 拦截（C-8）。
- 验证 `GenerationProvenance` schema 已在 `components.schemas` 中存在；spec §4.6 固定 AI schema 名单（`TargetJob.summary` / `TargetJob.fitSummary` / `AssistantAction` / `FeedbackReport` / AI-created `MistakeEntry` / `ResumeTailorRun` / `Debrief`）的字段图通过 `$ref` 链路可追溯到该 schema；fixture 级 provenance 校验由 002 闭合（C-11 partial）。

#### 4.3 文档与 INDEX 同步

- 本 plan checklist 全部勾选；Phase 4 关键命令日志贴入工作日志。
- 本 plan 自身 checklist 与 Phase 4 验证完成后，把本 plan Header 从 active 切到 completed，并用 `/sync-doc-index --fix-index` 同步 [openapi-v1-contract/plans/INDEX.md](../INDEX.md)；不联动 002 / 003 状态。
- 调用 `/sync-doc-index --check` 确认 `docs/spec/INDEX.md` 与 `plans/INDEX.md` 与 spec / plan Header 无 drift。
- 不修改 [engineering-roadmap/001-decompose-subspecs Phase 3.3](../../../engineering-roadmap/plans/001-decompose-subspecs/checklist.md#phase-3-wave-1基础设施--契约骨架) 已完成的 spawn 项；W2 implementation 准入 gate（spec C-10）由 [003-breaking-change-gate](../003-breaking-change-gate/plan.md) Phase 4 关闭，不重复登记父 roadmap。

#### 4.4 B2 child 协作 handoff

- [002-fixtures-and-mock-source](../002-fixtures-and-mock-source/plan.md) 在本 plan 完结后 spawn 自身实施：本 plan 输出的 `openapi.yaml` 与 generated 类型是 002 fixture schema 校验、operationId 列表与 mock parity 的真理源。
- [003-breaking-change-gate](../003-breaking-change-gate/plan.md) 在本 plan 完结后 spawn：本 plan 末态 `openapi/openapi.yaml` 即 v1.0.0 freeze baseline，由 003 拷贝到 `openapi/baseline/openapi-v1.0.0.yaml` 锁定。

### Phase 5: Assessment remediation

#### 5.1 复核 bootstrap assessment 建议

逐项核验 [2026-04-28 openapi-v1-contract/001-bootstrap assessment](../../../../reports/2026-04-28-openapi-v1-contract-001-bootstrap-assessment.md) 的 R1-R6 建议。只修订已被当前仓库文件证实存在的漂移：ADR-Q1 session cookie name 未锁、B2 tooling 选型未登记 deprecated-but-accepted 边界、`ResourceType` / `JobType` 字面量仍停留在待确认事项、`openapi/README.md` 外部工具写作约定不足，以及 B1 `ApiError` inner object 与 OpenAPI error envelope / 双端 generated type 的口径不一致。

#### 5.2 修订契约与文档真理源

更新 ADR-Q1、A4 `secrets-and-config`、B1 `shared-conventions-codified`、B2 `openapi-v1-contract`、`openapi/README.md` 及必要的上游技术文档引用，确保 cookie 字面量、B2 tooling 锁版、`ResourceType` / `JobType` 字面量和 `ApiError` / `ApiErrorResponse` 形状由真理源显式说明，不再依赖执行时反推。

#### 5.3 修复 codegen 并重新生成 artefacts

修订 `backend/cmd/codegen/openapi` 的 B1-AUTO block 与 Go/TS render 逻辑：`ApiError` 表示 B1 共享 inner error object；OpenAPI response body 使用 `ApiErrorResponse` envelope；Go 端复用 `backend/internal/shared/errors.APIError`；TS 端继续复用 `frontend/src/lib/conventions.ApiError`。重新运行 codegen，保证 `openapi/openapi.yaml`、`backend/internal/api/generated/` 与 `frontend/src/api/generated/` 字节级可再生。

#### 5.4 验证与生命周期收口

运行 focused generator tests、`make codegen-check`、`cd backend && go build ./...`、`cd frontend && npx tsc --noEmit`，随后把本 plan/checklist 恢复为 `completed` 并同步 INDEX。R5 中关于大文件写入和 `text/template` 的 skill 提示属于低价值流程优化，本次仅记录为 assessment 中的 no-op，不修改 `.agent-skills/tdd/SKILL.md`。

### Phase 6: docs-openapi renderer deprecation remediation

#### 6.1 复现 deprecated renderer 现象

确认 `make docs-openapi` 旧实现虽然 exit 0 且能生成 `openapi/dist/index.html`，但会打印 `redoc-cli` deprecated 横幅，并提示使用 `npx @redocly/cli build-docs <api>`。该现象属于 local docs renderer tooling drift，不影响 `make lint-openapi` 的 C-1 validation gate。

#### 6.2 迁移本地 HTML renderer

将根 `Makefile` 的 `docs-openapi` target 从 `redoc-cli@0.13.21 redoc-cli bundle` 迁移为 `@redocly/cli@2.30.1 redocly build-docs`，保持输入 `openapi/openapi.yaml`、输出 `openapi/dist/index.html` 和标题 `easyinterview API` 不变；不修改 `make lint-openapi` validator。

#### 6.3 文档与验证收口

同步 `openapi/README.md`、B2 spec/history、plan/checklist 的 tooling 说明；运行 `make docs-openapi` 确认不再出现 deprecated 横幅且产物生成成功，再运行 `make lint-openapi`、`/sync-doc-index --check` 与 `git diff --check`。

### Phase 7: v1.8 contract remediation

#### 7.1 `deleteMe` operation 与 inventory

将 `openapi/openapi.yaml`、inventory lint、generated Go/TS types 与 server/client interfaces 更新到 spec v1.8 的 37 endpoint 集合，新增 `DELETE /api/v1/me` / `operationId=deleteMe` / `202 PrivacyRequestWithJob`，并确保 Auth tag 下 account deletion 与 `POST /api/v1/privacy/deletions` 语义一致。

#### 7.2 Idempotency-Key 与 deletion dedupe

`DELETE /api/v1/me` 必须声明 `Idempotency-Key` header 或等价 active-request dedupe；重复删除请求返回同一 active `privacy_delete` job 或同义终态，避免 account deletion 重复排队。

#### 7.3 P0 debrief schema 收口

P0 `Debrief` / `DebriefWithJob` 只保留真实面试复现与复盘所需字段；感谢信草稿与完整跟进建议保持 P1 optional / hidden，不作为 P0 required 字段。

## 4 验收标准

- spec [§6 验收标准](../../spec.md#6-验收标准) C-1 的 contract/schema 部分、C-2 / C-3 / C-8 全部成立；C-7 / C-11 中本 plan 对应的契约 / schema 部分（非 fixture / 非 baseline）成立，剩余部分由 002 / 003 闭合。
- 本 plan checklist 全部勾选；Phase 4 关键命令日志贴入工作日志。
- B1 共享类型变更通过 `$ref` 自动同步进 OpenAPI 与 codegen，不形成手写副本（spec §3 D-2 / §4.2 / §4.3）。
- Phase 5 remediation 中确认存在的 assessment 问题已修订；`ApiError` inner object 与 OpenAPI response envelope 的 Go/TS 产物一致；未采纳的 R5 有明确 no-op 说明。
- Phase 6 docs renderer 迁移后，`make docs-openapi` 不再触发 `redoc-cli` deprecated 横幅，仍输出 `openapi/dist/index.html`；C-1 validator 保持 `@apidevtools/swagger-cli@4.0.4` + inventory lint。

## 5 风险与应对

| 风险 | 应对措施 |
|------|----------|
| oapi-codegen / openapi-typescript 模板默认输出与项目命名风格（camelCase / type alias）不一致 | Phase 2.1 / 2.2 把模板拷进 `openapi/templates/{go,ts}/` 仓库内可追溯；首次 codegen 后人工对比 generated 文件命名，必要时定制模板再固化 idempotent baseline |
| B1 generator 输出与本 plan 同步运行时形成竞态（先跑 B2 后跑 B1 会得到不一致结果） | Phase 2.3 在 root `Makefile` 强制 `codegen: codegen-conventions codegen-openapi`；`make codegen-openapi` 内部 first step 先 `$(MAKE) codegen-conventions`，避免直接调用时漏 B1 |
| OpenAPI 3.1 与部分 codegen 工具链兼容性差（部分老版本工具默认 3.0） | Phase 2 选定的 generator 必须明确支持 3.1（在 `openapi/README.md` 标注最低版本）；不接受静默 downgrade 到 3.0；如必须 3.0，递增本 plan 版本并加 ADR |
| `Idempotency-Key` / `clientEventId` 误挂混用导致 handler 语义分裂 | Phase 1.4 inventory 脚本强制：除 ADR-Q1 auth email start 与 session event 例外外，spec D-6 涉及的副作用 endpoint 必须挂 `Idempotency-Key`；`POST /api/v1/auth/email/start` 不挂通用 idempotency；`POST /practice/sessions/{sessionId}/events` 必须挂 `clientEventId` 且不挂 `Idempotency-Key`；任何冲突直接 fail |
| `GenerationProvenance` 在大量 AI schema 上误传播（错把 transactional schema 也挂上去） | Phase 1.2 在 `openapi/README.md` 与 schema 注释中明确「至少」名单（`TargetJob.summary` / `fitSummary` / `AssistantAction` / `FeedbackReport` / AI-created `MistakeEntry` / `ResumeTailorRun` / `Debrief`）；非 AI 生成 schema 不应携带 provenance |
| `make codegen-check` 在 IDE auto-format 后误报漂移 | Phase 2.1 / 2.2 generator 必须固定 import 顺序、行尾、缩进；与 `.editorconfig` 对齐；首次 idempotent baseline 由本 plan 锁定，编辑器若改动须修 generator 模板，不放弃 gate |

## 6 修订记录

| 日期 | 版本 | 变更 | 关联材料 |
|------|------|------|----------|
| 2026-04-28 | 1.2 | 根据 `make docs-openapi` deprecated 输出追加 Phase 6：将本地 HTML renderer 从 `redoc-cli@0.13.21` 迁移到 `@redocly/cli@2.30.1 build-docs`，不改变 C-1 validator。 | user report / local reproduction |
| 2026-04-28 | 1.1 | 根据 bootstrap assessment 追加 Phase 5 remediation：锁定 ADR-Q1 cookie name、B2 tooling 边界、`ResourceType` / `JobType` 字面量与 `ApiError` inner/envelope 生成口径。 | [assessment](../../../../reports/2026-04-28-openapi-v1-contract-001-bootstrap-assessment.md) |
| 2026-04-29 | 1.3 | 原地 reopen，新增 Phase 7 remediation：补齐 v1.8 spec 的 37 endpoint inventory、`DELETE /api/v1/me` account deletion、Idempotency-Key 与 generated codegen artifacts。 | plan-review remediation |
