# OpenAPI v1 Contract Spec

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-04-27

## 1 背景与目标

[engineering-roadmap spec §5.2](../engineering-roadmap/spec.md#52-layer-b--contract4-份全部-p0) 把 B2 `openapi-v1-contract` 列为 Layer B · Contract 的核心 child（依赖 [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md)；间接依赖 [A1 `repo-scaffold`](../repo-scaffold/spec.md)）。它是整个 DAG 中**最关键瓶颈节点**——P0 全部 backend（C 域）与 frontend（D 域）child 都依赖本契约的 codegen；任何破坏性变更会触发跨 spec 雪球。

本 spec 由 [001-decompose-subspecs Phase 3.3](../engineering-roadmap/plans/001-decompose-subspecs/checklist.md#phase-3-wave-1基础设施--契约骨架) 锁定为 **W1 spec-contract lock**：parent phase 先固定 `openapi/openapi.yaml` v1.0.0 freeze 范围（32+ endpoints / 14 tags / 字段命名 / additive-only 规则）。真实 OpenAPI 文件、codegen、fixtures 与 breaking-change linter 由 B2 child `001` 系列 plan 验证；未通过前不得启动依赖 B2 的 W2 implementation。

目标是：

1. **唯一真理源**：`openapi/openapi.yaml` 是 P0 所有 HTTP 端点的唯一定义；任何手写 handler stub / 手写 fetch 客户端禁止与之偏离。
2. **双端 codegen**：Go DTO + chi handler 接口在 `backend/internal/api/generated/`；TypeScript SDK 在 `frontend/src/api/generated/`；CI 必须 `git diff --exit-code` 校验未漂移（与 [B1 D-1 idempotent generator](../shared-conventions-codified/spec.md#31-已锁定决策) 一致）。
3. **fixtures 同源**：每个端点的 example response 落 `openapi/fixtures/<tag>/<operationId>.json`，由 [E1 `mock-contract-suite`](../engineering-roadmap/spec.md#55-layer-e--integration4-份) 转 Prism / 自建 mock server；前端 msw 与后端 mock-server 共享同一份 fixtures，**禁止前端 hardcode mock**。
4. **breaking change 拦截**：本 spec 自带 breaking change linter（如 `openapi-diff` / Spectral 规则集）；W1 末 v1.0.0 freeze 后任何 PR 修改 `openapi/openapi.yaml` 时，CI 必须验证只引入 additive 变更；破坏性变更必须通过 ADR + 本 spec 修订流程。

本 spec 不实现具体业务 handler（归各 C 域）、不实现前端业务页面（归各 D 域）、不部署 API 进程（归 [E4](../engineering-roadmap/spec.md#55-layer-e--integration4-份)）。

## 2 范围

### 2.1 In Scope

- **OpenAPI 文档**：`openapi/openapi.yaml` 单根文件（OpenAPI 3.1）；splits 由 generator 在构建时合并；所有路径前缀 `/api/v1`。
- **14 个 tag**（与 [02-api-definition.md §19](../../../easyinterview-tech-docs/02-api-definition.md#19-推荐的-openapi-拆分方式) 一致）：
  1. `Auth`、2. `Uploads`、3. `Profile`、4. `Resumes`、5. `TargetJobs`、6. `PracticePlans`、7. `PracticeSessions`、8. `Reports`、9. `Mistakes`、10. `ResumeTailor`、11. `Debriefs`、12. `Growth`、13. `Jobs`、14. `Privacy`。
- **endpoint 集**：32+ 端点，覆盖 [02-api-definition.md §4–§17](../../../easyinterview-tech-docs/02-api-definition.md) 全部端点；本 spec §3.1.1 列出 v1.0.0 freeze 时的 endpoint 列表。
- **schema 定义**：所有公共对象模型（`Job` / `TargetJob` / `PracticePlan` / `PracticeSession` / `AssistantAction` / `FeedbackReport` / `MistakeEntry` / `Debrief` / `ResumeAsset` / `ResumeTailorRun` / `PrivacyRequest` / `Job` / 共享 `ApiError` / `PageInfo` / `PaginatedXxx`）；引用 [B1 D-6 枚举](../shared-conventions-codified/spec.md#31-已锁定决策) 中 14 个枚举类型与 [D-5 错误码](../shared-conventions-codified/spec.md#31-已锁定决策) 常量。
- **header 与状态码契约**：[02 §2](../../../easyinterview-tech-docs/02-api-definition.md#2-通用约定) 中 `Authorization` / `Idempotency-Key` / `traceparent` / `X-Request-ID` 锁定；状态码使用集合（200/201/202/204/400/401/403/404/409/422/429/500）锁定。
- **codegen pipeline**：`make codegen-openapi`（B2 owner）输出 Go + TS；CI drift 校验。
- **fixtures**：每个 operation 对应一份默认 fixture（`scenario: default`）+ `easyinterview-ui/src/data.jsx` 折出来的 `scenario: prototype-baseline`（与 [engineering-roadmap §4.3 mock-first](../engineering-roadmap/spec.md#43-mock-first-集成策略) 一致）。
- **breaking change linter**：CI 引入 `openapi-diff`（或等价工具）；规则集见 §4.2。
- **API 文档站点**：`make docs-openapi` 输出可阅读 HTML（Redoc / Stoplight）；artifact 由 [A5](./../ci-pipeline-baseline/spec.md) 在 CI 输出。

### 2.2 Out of Scope

- 业务 handler 实现：归各 C 域。
- 前端业务页面：归各 D 域。
- mock server 运行壳：归 [E1 `mock-contract-suite`](../engineering-roadmap/spec.md#55-layer-e--integration4-份)；本 spec 只交付 fixtures。
- WebSocket / SSE / GraphQL：当前 P0 不在范围（练习会话 SSE 未来由本 spec 修订接入）。
- gRPC / Thrift：不在范围。
- 鉴权机制本身（passwordless email / token 颁发）：归 [C1 `backend-auth`](../engineering-roadmap/spec.md#53-layer-c--backend14-份p08--p14--p22) 与 [ADR-Q1](../engineering-roadmap/decisions/ADR-Q1-auth.md)；本 spec 仅锁 `Authorization: Bearer <token>` 形式。
- 限流策略具体阈值：归 [F1](./../observability-stack/spec.md) + 各 C 域；本 spec 仅锁 `429 Too Many Requests` 状态码使用。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策（v1.0.0 freeze 范围）

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 路径前缀 | 所有 endpoint 以 `/api/v1` 起始 | 与 [02 §1](../../../easyinterview-tech-docs/02-api-definition.md#1-概览) 一致 |
| D-2 | 字段命名 | JSON 字段 `camelCase`；URL path 参数 `camelCase`（如 `{targetJobId}`）；query 参数 `camelCase` | 与 [00-shared-conventions](../../../easyinterview-tech-docs/00-shared-conventions.md) 一致 |
| D-3 | 时间格式 | `string` + `format: date-time`，RFC3339 UTC（如 `2026-04-23T13:45:12Z`） | – |
| D-4 | 错误响应 schema | 全部 4xx/5xx 复用 `ApiError`（`error.code` / `error.message` / `error.requestId` / `error.retryable` / `error.details`）；`error.code` 必须出现在 [B1 D-5](../shared-conventions-codified/spec.md#31-已锁定决策) 锁定的错误码常量集合 | 切实业务 handler 不能擅自新增错误码 |
| D-5 | 分页 | 所有列表 endpoint 使用 cursor 分页 + 统一 `pageInfo`（`nextCursor` / `pageSize` / `hasMore`）；不混用 offset 分页 | – |
| D-6 | Idempotency | 所有创建类 endpoint（POST + 副作用）支持 `Idempotency-Key` header（24h TTL，由 [B1 工具](../shared-conventions-codified/spec.md#21-in-scope) 实现） | – |
| D-7 | Job 异步 | 长耗时操作返回 `202 Accepted` + `Job` schema；客户端通过 `GET /jobs/{jobId}` 轮询 | – |
| D-8 | content-type | 仅 `application/json` 与 `multipart/form-data`（仅 upload 端点）；不引入 protobuf / msgpack | – |
| D-9 | v1.0.0 freeze 范围 | §3.1.1 列出 36 个 endpoint + 14 tag；W1 parent phase 锁定范围与 additive-only 规则，B2 child `001` 落地 `openapi/openapi.yaml` 后强制执行（新增 endpoint / 新增可选字段 / 新增枚举值） | 任何 break change 必须 ADR + 本 spec 修订 |
| D-10 | breaking change linter | 默认 `openapi-diff`（OpenAPITools）；规则：禁止删字段、禁止改字段类型、禁止改 required、禁止改枚举（仅允许新增）、禁止删 endpoint | CI 直接失败 |
| D-11 | tags 顺序 | §2.1 14 个 tag 顺序固定；新增 tag 必须递增 spec | – |
| D-12 | privacy export 例外 | 按 [ADR-Q5](../engineering-roadmap/decisions/ADR-Q5-privacy-cadence.md)，`POST /api/v1/privacy/exports` 在 v1.0.0 freeze 中保留路径与 schema，但 P0 必须返回 `501 Not Implemented`（`error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`）；P1 切换实现时是 additive 行为变化，不算 break | 防止 P1 复用时改路径 |

#### 3.1.1 v1.0.0 freeze endpoint 列表

| # | Tag | Method | Path | OperationId | 关联 schema |
|---|-----|--------|------|-------------|-------------|
| 1 | Auth | GET | /api/v1/me | getMe | UserContext |
| 2 | Auth | POST | /api/v1/auth/magic-link | requestMagicLink | – |
| 3 | Auth | POST | /api/v1/auth/sessions | createSession | Session |
| 4 | Auth | DELETE | /api/v1/auth/sessions/current | endSession | – |
| 5 | Uploads | POST | /api/v1/uploads/presign | createUploadPresign | UploadPresign |
| 6 | Profile | GET | /api/v1/profiles/me | getMyProfile | CandidateProfile |
| 7 | Profile | PATCH | /api/v1/profiles/me | updateMyProfile | CandidateProfile |
| 8 | Profile | GET | /api/v1/profiles/me/experience-cards | listExperienceCards | PaginatedExperienceCard |
| 9 | Profile | POST | /api/v1/profiles/me/experience-cards | createExperienceCard | ExperienceCard |
| 10 | Profile | PATCH | /api/v1/profiles/me/experience-cards/{cardId} | updateExperienceCard | ExperienceCard |
| 11 | Resumes | POST | /api/v1/resumes | registerResume | ResumeAssetWithJob |
| 12 | Resumes | GET | /api/v1/resumes/{resumeAssetId} | getResume | ResumeAsset |
| 13 | TargetJobs | POST | /api/v1/targets/import | importTargetJob | TargetJobWithJob |
| 14 | TargetJobs | GET | /api/v1/targets | listTargetJobs | PaginatedTargetJob |
| 15 | TargetJobs | GET | /api/v1/targets/{targetJobId} | getTargetJob | TargetJob |
| 16 | TargetJobs | PATCH | /api/v1/targets/{targetJobId} | updateTargetJob | TargetJob |
| 17 | PracticePlans | POST | /api/v1/practice/plans | createPracticePlan | PracticePlan |
| 18 | PracticePlans | GET | /api/v1/practice/plans/{planId} | getPracticePlan | PracticePlan |
| 19 | PracticeSessions | POST | /api/v1/practice/sessions | startPracticeSession | PracticeSession |
| 20 | PracticeSessions | GET | /api/v1/practice/sessions/{sessionId} | getPracticeSession | PracticeSession |
| 21 | PracticeSessions | POST | /api/v1/practice/sessions/{sessionId}/events | appendSessionEvent | SessionEventResult |
| 22 | PracticeSessions | POST | /api/v1/practice/sessions/{sessionId}/complete | completePracticeSession | ReportWithJob |
| 23 | Reports | GET | /api/v1/reports/{reportId} | getFeedbackReport | FeedbackReport |
| 24 | Reports | GET | /api/v1/targets/{targetJobId}/reports | listTargetJobReports | PaginatedFeedbackReport |
| 25 | Mistakes | GET | /api/v1/mistakes | listMistakes | PaginatedMistakeEntry |
| 26 | Mistakes | POST | /api/v1/mistakes/{mistakeId}/retest | retestMistake | PracticePlanContainer |
| 27 | ResumeTailor | POST | /api/v1/resume/tailor | requestResumeTailor | ResumeTailorRunWithJob |
| 28 | ResumeTailor | GET | /api/v1/resume/tailor-runs/{tailorRunId} | getResumeTailorRun | ResumeTailorRun |
| 29 | Debriefs | POST | /api/v1/debriefs | createDebrief | DebriefWithJob |
| 30 | Debriefs | GET | /api/v1/debriefs/{debriefId} | getDebrief | Debrief |
| 31 | Growth | GET | /api/v1/growth/overview | getGrowthOverview | GrowthOverview |
| 32 | Jobs | GET | /api/v1/jobs/{jobId} | getJob | Job |
| 33 | Privacy | POST | /api/v1/privacy/exports | requestPrivacyExport | PrivacyRequestWithJob（P0 返回 501） |
| 34 | Privacy | POST | /api/v1/privacy/deletions | requestPrivacyDelete | PrivacyRequestWithJob |
| 35 | Privacy | GET | /api/v1/privacy/requests/{privacyRequestId} | getPrivacyRequest | PrivacyRequest |
| 36 | Auth | GET | /api/v1/runtime-config | getRuntimeConfig | RuntimeConfig（[A4 D-2](../secrets-and-config/spec.md#31-已锁定决策含-p0-必备-env-key-字典) owner） |

总计 36 个 endpoint（满足「32+」表述），覆盖 14 tag。

### 3.2 待确认事项

- v1.0.1 / v1.1.0 升级阈值：default 使用 SemVer，破坏性变更 → v2.0.0；v1.x 内累积 ≥ 5 个新 endpoint 触发 v1.1.0；具体由本 spec 修订时决策。
- SSE 子协议（练习会话流式 follow-up）：默认 P0 不上；如 W3 业务域提出，由本 spec 修订决策。
- API 文档发布平台：默认 Redoc；如有偏好可在 001-bootstrap plan 落地时回填。
- 公共 `ResourceType` 枚举（Job / outbox 引用）是否独立成 schema：默认独立（避免重复字符串），由 codegen 引导。

## 4 设计约束

### 4.1 schema 设计约束

- 所有 enum 字段必须以 [B1 D-6 枚举](../shared-conventions-codified/spec.md#31-已锁定决策) 中的 14 个类型为基础；本 spec 不重新定义 enum 字面量，必须 `$ref` 到 B1 共享 enum schema。
- `ApiError` schema 必须 `$ref` 到 B1 提供的共享类型；`error.code` 字段定义为枚举（值集等于 [B1 D-5](../shared-conventions-codified/spec.md#31-已锁定决策) 全部错误码常量），由 generator 自动同步。
- 所有 `id` 字段为 `string`，`format: uuid`；服务端写入字段值必须 UUIDv7（由 B1 idx 工具生成）；前端临时 id（`tmp_<uuid>`）只在前端 state 中存在，不进 API 请求体。
- 所有时间字段统一 `string` + `format: date-time`；不允许某些字段使用 unix epoch number。

### 4.2 breaking change linter 规则集（W1 末 freeze 后强制）

- **禁止**：删除已发布 endpoint / 重命名 path / 修改 method / 删除 schema 字段 / 修改字段类型 / 把 optional 字段改为 required / 删除已发布枚举值。
- **允许（additive）**：新增 endpoint / 新增 tag / 新增 optional 字段 / 新增枚举值（且字段为 string-typed enum） / 新增可选 query 参数 / 新增 example。
- **审计要求**：违反规则的 PR 必须 attach ADR 链接并在本 spec history 表加一行「v2.0.0 升级」记录；CI 中通过 label `breaking-change-approved` 并由 B2 owner approve 才能合入。

### 4.3 codegen 与 drift 约束

- generator 输入：`openapi/openapi.yaml` + `openapi/templates/`（Go / TS 模板）；输出 `backend/internal/api/generated/` 与 `frontend/src/api/generated/`。
- generated 文件必须 idempotent；CI `git diff --exit-code` 阻塞漂移。
- 业务 handler 必须 implement generator 产出的 server interface；不允许业务包定义自己的 DTO 类型。

### 4.4 fixtures 与隐私约束

- `openapi/fixtures/<tag>/<operationId>.json` 必须 schema-valid（CI 中由 `make validate-fixtures` 校验）。
- fixtures 中绝不出现真实用户邮箱 / 真实电话 / 真实公司名敏感信息；统一用 `Acme` / `acme.example` / `alice@example.com`。
- `prototype-baseline` scenario 来自 `easyinterview-ui/src/data.jsx`；维护方式：`make sync-fixtures-from-prototype`（B2 owner）。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `openapi/openapi.yaml` 与 fixtures | B2 | 唯一真理源 |
| 14 个 enum 类型 / `ApiError` / `PageInfo` schema | B1 | B2 通过 `$ref` 引用 |
| 错误码常量列表 | B1 | B2 在 `error.code` 枚举中同步 |
| Go 与 TS codegen | B2 + B1（generator base） | 输出落点固定 |
| 业务 handler 实现 | C 域各 owner | 必须 implement 生成的 server interface |
| 前端 fetch 客户端 | D 域各 owner | 必须使用生成的 TS client |
| mock server 运行壳 | E1 | 消费 fixtures |
| breaking change linter 接入 CI | B2 提供 + A5 接入 |  |
| API 文档发布 | B2（Redoc 集成） + A5（CI artifact） |  |
| 鉴权 token 颁发 | C1 + ADR-Q1 | B2 仅锁 `Authorization` header 形式 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | OpenAPI 文档结构 | `openapi/openapi.yaml` 已落地 | `npx @apidevtools/swagger-cli validate openapi/openapi.yaml` | 通过；含 14 tag、36 endpoint、共享 schema 全部 `$ref` 到 B1 | B2 后续 001 |
| C-2 | Go codegen drift | 修改 `openapi.yaml` 但不跑 codegen | CI | `codegen-drift-check` 失败；artifact `openapi-diff.html` 显示新增字段 | B2 后续 001 + A5 接入 |
| C-3 | TS codegen drift | 同 C-2 | CI | `frontend/src/api/generated/` 漂移；CI 失败 | B2 后续 001 |
| C-4 | breaking change 拦截 | 故意删除 `target_jobs.title` 字段 | CI | `openapi-diff` 失败；阻塞合入；除非 PR 含 `breaking-change-approved` label 且 B2 owner approve | B2 后续 001 + A5 |
| C-5 | additive 通过 | 给 `practice_plans` 新增 `optional metadata` 字段 | CI | `openapi-diff` 仅警告 additive；测试通过 | B2 后续 001 |
| C-6 | fixtures 一致 | 任一 endpoint 缺少 fixtures | `make validate-fixtures` | 失败；列出缺失 operationId | B2 后续 001 |
| C-7 | privacy export 501 | P0 调用 `POST /api/v1/privacy/exports` | E1 mock + 后续 C12 实现 | 返回 501 + `error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"` | B2（fixture）+ C12 P1 实现 |
| C-8 | enum 与 B1 同源 | 在 `openapi.yaml` 引用 `practiceMode` enum | codegen | 生成 TS 与 Go 类型，与 [B1 D-6](../shared-conventions-codified/spec.md#31-已锁定决策) 完全一致；改 B1 后 B2 codegen drift | B2 后续 001 + B1 |
| C-9 | mock 同源（前端 + 后端） | E1 拉起 mock server | 前端 msw 与后端 mock-server 都消费 `openapi/fixtures/` | 同一 endpoint 两端响应字节级一致 | B2 + E1 |
| C-10 | B2 executable freeze handoff | 本 spec 的 contract lock 已完成，B2 后续 `001` 完成 | engineering-roadmap §5.7 W1 准入 gate | `openapi/openapi.yaml` v1.0.0、codegen drift、fixtures 与 breaking-change linter 均通过验证；依赖 B2 的 W2 implementation 可启动；parent Phase 3 只记录 spec-contract lock，不单独冒充本项已通过 | B2 后续 `001` |

## 7 关联计划

B2 在本次 W1 spec 阶段不创建 impl plan（参见 [001-decompose-subspecs §3.1](../engineering-roadmap/plans/001-decompose-subspecs/plan.md#3-实施步骤)）。后续由 B2 自身的 plans 承接（`engineering-roadmap §5.2` 估算 3 plan）：

- `001-bootstrap`（W1 末或 W2 初）：落地 `openapi/openapi.yaml` 框架 + 14 tag 占位 + 32+ endpoint stub schema + B1 enum `$ref` + `make codegen-openapi` + drift CI 接入。
- `002-fixtures-and-mock-source`：每个 operationId 一份 fixtures + `prototype-baseline` 同步工具；E1 接入。
- `003-breaking-change-gate`：linter 规则集 + CI label workflow + ADR 模板。

后续如出现 v1.1.0 / v2.0.0 升级：递增 spec 版本 + history；每次升级在 §3.1.1 中保留 endpoint 完整快照。
