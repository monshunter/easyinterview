# OpenAPI v1 Contract Spec

> **版本**: 1.37
> **状态**: active
> **更新日期**: 2026-07-09

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将原始 B2 `openapi-v1-contract` 保留为当前 active Contract spec（依赖 [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md)；间接依赖 [A1 `repo-scaffold`](../repo-scaffold/spec.md)）。它是当前 P0 backend / frontend workstream 的 HTTP 契约瓶颈节点：后续实现必须复用本契约的 codegen、fixtures 与 breaking-change gate；任何破坏性变更会触发跨 spec 雪球。

本 spec 由 `engineering-roadmap/001-decompose-subspecs` 的 contract lock 创建；当前执行口径以 roadmap active spec 的保留规则为准：`openapi/openapi.yaml` v1.0.0 freeze 范围为当前 37 endpoints / 10 tags、字段命名和 additive-only 规则。真实 OpenAPI 文件、codegen、fixtures 与 breaking-change linter 由 B2 `001-bootstrap` / `002-fixtures-and-mock-source` / `003-breaking-change-gate` / `004-resume-additive-coverage` 分别验证；未通过前不得启动依赖 B2 的 implementation。

当前 HTTP 可执行契约由本 spec、`openapi/openapi.yaml`、OpenAPI fixtures / baseline 与 B1 shared-conventions-codified 决定。B2 独立承接 endpoint inventory、tag、auth 形态、header、status code、schema、fixture provenance 与 breaking-change gate；任何实现或 codegen 都不得绕过这些当前 owner truth source。

目标是：

1. **唯一真理源**：`openapi/openapi.yaml` 是 P0 所有 HTTP 端点的唯一定义；任何手写 handler stub / 手写 fetch 客户端禁止与之偏离。
2. **双端 codegen**：Go DTO + chi handler 接口在 `backend/internal/api/generated/`；TypeScript SDK 在 `frontend/src/api/generated/`；本地 `make codegen-openapi` / `make codegen-check` 必须能用 `git diff --exit-code` 校验未漂移（与 [B1 D-1 idempotent generator](../shared-conventions-codified/spec.md#31-已锁定决策) 一致）。
3. **fixtures 同源**：每个端点的 example response 落 `openapi/fixtures/<tag>/<operationId>.json`，由 [E1 `mock-contract-suite`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 转 Prism / 自建 mock server；需要给 Prism / 文档站消费的 OpenAPI examples 必须由 fixtures 生成，不手写第二份 example；前端 msw 与后端 mock-server 共享同一份 fixtures，**禁止前端 hardcode mock**。
4. **breaking change 拦截**：本 spec 自带 breaking change linter（如 `openapi-diff` / Spectral 规则集）；v1.0.0 freeze 生效后任何修改 `openapi/openapi.yaml` 时，本地 gate 必须验证只引入 additive 变更；破坏性变更必须通过 ADR + 本 spec 修订流程。

本 spec 不实现具体业务 handler（归各 C 域）、不实现前端业务页面（归各 D 域）、不部署 API 进程（归 [E4](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)）。

## 2 范围

### 2.1 In Scope

- **OpenAPI 文档**：`openapi/openapi.yaml` 单根文件（OpenAPI 3.1）；splits 由 generator 在构建时合并；所有路径前缀 `/api/v1`。
- **10 个 tag**：1. `Auth`、2. `Uploads`、3. `Resumes`、4. `TargetJobs`、5. `PracticePlans`、6. `PracticeSessions`、7. `Reports`、8. `ResumeTailor`、9. `Jobs`、10. `Privacy`。当前 tag inventory 以 `openapi/openapi.yaml` 与 `scripts/lint/openapi_inventory.py` 为准。
- **endpoint 集**：37 端点，覆盖当前 P0 contract；本 spec §3.1.1 列出 v1.0.0 freeze 时的 endpoint 列表。任何新增 endpoint、tag、schema 或 fixture 都必须先修订本 spec、OpenAPI baseline、fixtures、generated artifacts 与 inventory lint。
- **schema 定义**：所有 endpoint request / success 或 P0 例外 response / async wrapper / error response 必须出现在 §4.2 schema inventory，或显式声明无 body / 无响应体；共享 `ApiError` inner object / `PageInfo` / `PaginatedXxx` 与 16 个枚举类型引用 [B1 D-5/D-7/D-10](../shared-conventions-codified/spec.md#31-已锁定决策)，OpenAPI 只负责 `ApiErrorResponse` 外层 envelope 与 B2 专属 enum（`ResourceType` / `JobType`），不得重复维护 B1 enum 字面量。
- **header 与状态码契约**：由本 spec §4.1 与 `openapi/openapi.yaml` 的 components 共同承接；认证形态以 [ADR-Q1](../engineering-roadmap/decisions/ADR-Q1-auth.md) 与本 spec 为准：P0 使用 first-party session cookie；`Authorization: Bearer` 不属于当前 P0 contract。状态码矩阵见 §4.1。
- **codegen pipeline**：`make codegen-openapi`（B2 owner）输出 Go + TS；本地 drift 校验。
- **fixtures**：每个 operation 对应一份默认 fixture（`scenario: default`）+ `ui-design/src/data.jsx` 折出来的 `scenario: prototype-baseline`（与 [engineering-roadmap §4.3 mock-first](../engineering-roadmap/spec.md#43-契约与-mock-first-约束) 一致）。
- **breaking change linter**：本地引入 `openapi-diff`（或等价工具）；规则集见 §4.4。
- **API 文档站点**：`make docs-openapi` 输出可阅读 HTML（当前锁 `@redocly/cli@2.30.1 build-docs`）；当前单人阶段只保留本地产物，不要求 A5 上传 CI artifact。
- **tooling 锁定**：`make lint-openapi` 使用 `npx @apidevtools/swagger-cli@4.0.4` + inventory lint；换用 `@redocly/cli` 或其它 validator 作为 validation gate 前必须修订本 spec / plan 并提供实测证据。`make docs-openapi` 使用 Redocly CLI docs renderer，不参与 C-1 validation gate。

### 2.2 Out of Scope

- 业务 handler 实现：归各 C 域。
- 前端业务页面：归各 D 域。
- mock server 运行壳：归 [E1 `mock-contract-suite`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)；本 spec 只交付 fixtures。
- WebSocket / SSE / GraphQL：当前 P0 不在范围（练习会话 SSE 未来由本 spec 修订接入）。
- gRPC / Thrift：不在范围。
- 鉴权机制本身（email-code challenge、session cookie 颁发 / 撤销、风控阈值）：归 [C1 `backend-auth`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 与 [ADR-Q1](../engineering-roadmap/decisions/ADR-Q1-auth.md)；本 spec 只冻结 HTTP contract、public/protected 边界与 OpenAPI security scheme。
- 限流策略具体阈值：归 [F1](./../observability-stack/spec.md) + 各 C 域；本 spec 仅锁 `429 Too Many Requests` 状态码使用。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策（v1.0.0 freeze 范围）

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 路径前缀 | 所有 endpoint 以 `/api/v1` 起始 | 当前 endpoint inventory 以 §3.1.1 与 `openapi/openapi.yaml` 为准 |
| D-2 | 字段命名 | JSON 字段 `camelCase`；URL path 参数 `camelCase`（如 `{targetJobId}`）；query 参数 `camelCase` | 与 B1 当前 shared conventions 一致 |
| D-3 | 时间格式 | `string` + `format: date-time`，RFC3339 UTC（如 `2026-04-23T13:45:12Z`） | – |
| D-4 | 错误响应 schema | 全部 4xx/5xx wire body 复用 `ApiErrorResponse` envelope（`error.code` / `error.message` / `error.requestId` / `error.retryable` / `error.details`）；inner `error` object 复用 B1 `ApiError`；`error.code` 必须出现在 [B1 D-5/D-10](../shared-conventions-codified/spec.md#31-已锁定决策) 锁定的错误码常量集合（含 `PRIVACY_EXPORT_NOT_AVAILABLE` / `RESUME_EXPORT_NOT_AVAILABLE`） | 具体业务 handler 不能擅自新增错误码 |
| D-5 | 分页 | 所有列表 endpoint 使用 cursor 分页 + 统一 `pageInfo`（`nextCursor` / `pageSize` / `hasMore`）；不混用 offset 分页 | – |
| D-6 | Idempotency | 副作用 endpoint 的 `Idempotency-Key` 支持范围由本 spec §4.1 与 B1 幂等工具共同决定；`POST /practice/sessions/{sessionId}/events` 使用 `clientEventId` 去重，不混用 `Idempotency-Key` | 防止不同去重机制叠加导致 handler 语义分裂 |
| D-7 | Job 异步 | 长耗时操作返回 `202 Accepted` + `Job` schema；客户端通过 `GET /jobs/{jobId}` 轮询 | – |
| D-8 | content-type | 仅 `application/json` 与 `multipart/form-data`（仅 upload 端点）；不引入 protobuf / msgpack | – |
| D-9 | v1.0.0 freeze 范围 | §3.1.1 列出当前 37 个 endpoint + 10 tag；本 spec 锁定范围与 additive-only 规则，B2 `001` 落地 `openapi/openapi.yaml` 后强制执行（新增 endpoint / 新增可选字段 / 新增枚举值）；Auth tag 以 ADR-Q1 的 email-code challenge + session cookie 路径为准，`startAuthEmailChallenge` 只接收邮箱和 `returnTo`，不得用 `purpose` 或 `displayName` 区分注册/登录；`verifyAuthEmailChallenge` 消费 query `token` 但其语义为 6 位 code；`DELETE /api/v1/me` 按 ADR-Q5 纳入 P0 删除入口；`listPracticeSessions`、`createPracticeVoiceTurn`、`completeMyProfile`、扁平 Resume operations、`duplicateResume`、`getResumeSource` 和 `archiveTargetJob` 均属于当前 freeze | 任何 break change 必须 ADR + 本 spec 修订；当前 project pre-launch 阶段允许由 product-scope owner 授权 v1.0.0 freeze correction，并必须同步 history、baseline、fixtures、generated artifacts 和 diff-config |
| D-10 | breaking change linter | 默认 `openapi-diff`（OpenAPITools）；规则：禁止删字段、禁止改字段类型、禁止改 required、禁止改枚举（仅允许新增）、禁止删 endpoint | 本地 gate 直接失败；远端 CI 接入由 A5 后续触发条件决定 |
| D-11 | tags 顺序 | §2.1 10 个 tag 顺序固定；新增 tag 必须递增 spec | – |
| D-12 | privacy export 例外 | 按 [ADR-Q5](../engineering-roadmap/decisions/ADR-Q5-privacy-cadence.md)，`POST /api/v1/privacy/exports` 在 v1.0.0 freeze 中保留路径与 schema，但 P0 必须返回 `501 Not Implemented`（`error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`）；P1 切换实现时是 additive 行为变化，不算 break | 防止 P1 复用时改路径 |
| D-13 | OpenAPI tooling 锁版 | validation: `@apidevtools/swagger-cli@4.0.4`；docs: `@redocly/cli@2.30.1 build-docs`。禁止未修订 spec 时替换 C-1 validator | 避免 002 / 003 在不同 validation gate 间产生不一致错误面；docs renderer 升级必须记录实测兼容证据 |
| D-14 | B2 专属 async enum 字面量 | `ResourceType` 与 `JobType` 独立成 OpenAPI schema；字面量见 §3.1.2。它们来自当前 B2 API-facing async response set 与 P0 privacy exception，后续新增 endpoint / async job 时必须递增本 spec 并 additive 追加 enum 值 | 不再把 `ResourceType` 留作待确认；fixtures / mock / generated DTO 可直接依赖 |
| D-15 | 错误响应 envelope | B1 `ApiError` 表示 `error` inner object；B2 `ApiErrorResponse` 表示 wire body `{error: ApiError}`。所有 default 4xx/5xx 与 privacy export 501 响应使用 `ApiErrorResponse` envelope | 消除 Go/TS codegen 对 `ApiError` 名称的歧义 |
| D-16 | TargetJob import 场景契约 | `importTargetJob` 四类 source variant 均保持 `202 + TargetJobWithJob`；`manual_form` 是同步兜底路径，必须返回 terminal `Job(jobType=target_import,status=succeeded)`，不要求 backend async runner 处理；`TargetJobRequirement.kind` 需 additive 扩展为 B4 已有四类 `must_have` / `nice_to_have` / `hidden_signal` / `interview_focus` | `openapi/openapi.yaml`、fixtures、Go/TS generated artifacts 与 breaking-change baseline 必须由 backend-targetjob/001 Phase 0 同步；这是 additive enum / fixture 场景扩展，不删现有值 |
| D-19 | PracticeTurn.status pre-launch baseline rebase | `PracticeTurn.status` wire enum 原地 rebase 为 5 值：`asked` / `answered` / `follow_up_requested` / `assessed` / `skipped`。这是 backend-practice/002 在 v1.0.0 发布前对事件循环 runtime state 的 contract correction，不再把内部 turn lifecycle 压缩为 3 值。 | `openapi/openapi.yaml`、`openapi/baseline/openapi-v1.0.0.yaml`、`backend/internal/api/generated/`、`frontend/src/api/generated/`、inventory lint 与 generated artifact sync test 必须同时更新；`make codegen-openapi` / `make codegen-check` / 后端 build / 前端 typecheck 均为本次 rebase gate |
| D-21 | Practice sessions listing | 当前 37-operation freeze 包含 `GET /api/v1/practice/sessions` operationId `listPracticeSessions`，query 为 `targetJobId?` / `status?` / `cursor?` / `pageSize?`，response schema 为 `PaginatedPracticeSession`。该 endpoint 是 read-only，不挂 `Idempotency-Key`。 | `openapi/openapi.yaml`、`openapi/fixtures/PracticeSessions/listPracticeSessions.json`、`scripts/lint/openapi_inventory.py`、Go/TS generated artifacts 与 codegen test 保持同步 |
| D-22 | Practice voice turn | 当前 37-operation freeze 包含 `POST /api/v1/practice/sessions/{sessionId}/voice-turns` operationId `createPracticeVoiceTurn`，request/response schema 为 `CreatePracticeVoiceTurnRequest` / `PracticeVoiceTurnResult`，用于 voice mode 的 STT -> chat -> TTS 级联 turn。该 endpoint 是 side-effect operation，必须声明 `Idempotency-Key`；`PracticeSessionEventRequest.kind` 支持 voice playback / barge-in / committed context event。 | `openapi/openapi.yaml`、`openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json`、`scripts/lint/openapi_inventory.py`、Go/TS generated artifacts、fixture validator 与 codegen test 保持同步；真实 handler 由 `practice-voice-mvp/001-cascaded-stt-llm-tts` Phase 5 补齐 |
| D-25 | Auth single-entry profile completion | 当前 37-operation freeze 包含 protected `PATCH /api/v1/me` operationId `completeMyProfile`；邮箱是唯一账号标识，用户只从 `startAuthEmailChallenge` 发起同一个邮箱验证码登录入口；`AuthEmailStartRequest` 不使用 `purpose` / `displayName` 区分注册/登录，避免发码前泄露邮箱存在性；`UserContext.profileCompletionRequired` 标识首次登录资料未完成。`displayName` 不唯一，不参与账号去重；资料未完成账号每次登录后必须先进入资料补全。 | `openapi/openapi.yaml`、`openapi/fixtures/Auth/completeMyProfile.json`、`openapi/fixtures/Auth/getMe.json`、`scripts/lint/openapi_inventory.py`、Go/TS generated artifacts、frontend-shell/001 Phase 9、backend-auth/001 Phase 8 与 E2E.P0.101 保持同步 |
| D-26 | 简历资产扁平化 contract collapse（product-scope D-20） | 当前 freeze 中，Resumes tag 使用单一 `Resume` 实体、`resumeId` 路径参数、`UpdateResumeRequest` 覆盖保存和 `DuplicateResumeRequest` 另存为新简历；ResumeTailor 生成 ephemeral suggestions，采纳后通过 `updateResume` 或 `duplicateResume` 落盘；`RegisterResumeRequest.sourceType` 为 `upload` / `paste`；`Resume.structuredProfile` 承载结构化内容，`AI_PROVENANCE_SCHEMAS` 覆盖 `Resume`。 | `openapi/openapi.yaml`、`openapi/fixtures/Resumes/*`、`openapi/fixtures/ResumeTailor/*`、`scripts/lint/openapi_inventory.py`、`scripts/lint/validate_fixtures.py`、`openapi/README.md`、`openapi/fixtures/README.md`、`openapi/baseline/openapi-v1.0.0.yaml`、Go/TS generated client/server/types、`docs/spec/mock-contract-suite` / `docs/spec/engineering-roadmap` 计数同步；由 [openapi-v1-contract/004](./plans/004-resume-additive-coverage/plan.md) 与 product-scope current contract owner 一并审查 |
| D-27 | 核心闭环 contract boundary（product-scope D-22） | 当前 freeze 为 **37 endpoint / 10 tag**。账号资料补全由 Auth `completeMyProfile` 承接；practice 派生计划只接受 report-derived `retry_current_round` / `next_round`；Reports / PracticeSessions / Resumes / ResumeTailor / TargetJobs / Privacy 共同承接当前核心闭环。 | `openapi/openapi.yaml`、`openapi/fixtures/*`、`scripts/lint/openapi_inventory.py`、`scripts/lint/validate_fixtures.py`、`openapi/README.md`、`openapi/fixtures/README.md`、`openapi/baseline/openapi-v1.0.0.yaml`、Go/TS generated client/server/types、frontend/backend consumers 与 E2E.P0.098/P0.099 gate 同步；由 [product-scope/001-core-loop-module-pruning](../product-scope/plans/001-core-loop-module-pruning/plan.md) 落地。 |
| D-29 | TargetJob archive | 当前 freeze 包含 protected `POST /api/v1/targets/{targetJobId}/archive` operationId `archiveTargetJob`，用于 workspace 删除图标持久软归档 TargetJob。该 endpoint 是 side-effect operation，必须声明 `Idempotency-Key`，成功返回 archived `TargetJob`，read-side 继续通过 `deleted_at is null` 隐藏归档记录。 | `openapi/openapi.yaml`、`openapi/fixtures/TargetJobs/archiveTargetJob.json`、`scripts/lint/openapi_inventory.py`、Go/TS generated artifacts、backend-targetjob/001 Phase 12、frontend-workspace-and-practice/001 Phase 18 与 E2E.P0.018 保持同步 |
| D-28 | Resume PDF source preview | 当前 37-operation freeze 包含 protected `GET /api/v1/resumes/{resumeId}/source` operationId `getResumeSource`；仅用于 upload-backed PDF 原件 inline 预览，返回 `application/pdf` binary；paste、Markdown、TXT、missing、archived 和 cross-user 均返回 404。 | `openapi/openapi.yaml`、`openapi/fixtures/Resumes/getResumeSource.json`、`scripts/lint/openapi_inventory.py`、Go/TS generated artifacts、frontend-resume-workshop/001 Phase 8 与 backend-resume/001 Phase 12 保持同步 |

#### 3.1.1 v1.0.0 freeze endpoint 列表

| # | Tag | Method | Path | OperationId | 关联 schema |
|---|-----|--------|------|-------------|-------------|
| 1 | Auth | GET | /api/v1/me | getMe | UserContext |
| 2 | Auth | PATCH | /api/v1/me | completeMyProfile | CompleteProfileRequest / UserContext |
| 3 | Auth | DELETE | /api/v1/me | deleteMe | PrivacyRequestWithJob |
| 4 | Auth | POST | /api/v1/auth/email/start | startAuthEmailChallenge | AuthEmailStartRequest |
| 5 | Auth | GET | /api/v1/auth/email/verify | verifyAuthEmailChallenge | Session |
| 6 | Auth | POST | /api/v1/auth/logout | logout | – |
| 7 | Uploads | POST | /api/v1/uploads/presign | createUploadPresign | UploadPresign |
| 8 | Resumes | GET | /api/v1/resumes | listResumes | PaginatedResume |
| 9 | Resumes | POST | /api/v1/resumes | registerResume | ResumeWithJob |
| 10 | Resumes | GET | /api/v1/resumes/{resumeId} | getResume | Resume |
| 11 | Resumes | GET | /api/v1/resumes/{resumeId}/source | getResumeSource | application/pdf binary（upload-backed PDF inline source preview） |
| 12 | Resumes | PATCH | /api/v1/resumes/{resumeId} | updateResume | UpdateResumeRequest / Resume（IK 必带；覆盖 `structuredProfile` / `displayName` / 采纳改写后覆盖原简历） |
| 13 | Resumes | POST | /api/v1/resumes/{resumeId}/duplicate | duplicateResume | DuplicateResumeRequest / Resume（IK 必带；采纳改写后「保存为新简历」） |
| 14 | Resumes | POST | /api/v1/resumes/{resumeId}/archive | archiveResume | Resume（IK 必带） |
| 15 | Resumes | POST | /api/v1/resumes/{resumeId}/exports | exportResume | ApiErrorResponse（P0 501 + `RESUME_EXPORT_NOT_AVAILABLE`；IK 必带） |
| 16 | TargetJobs | POST | /api/v1/targets/import | importTargetJob | TargetJobWithJob |
| 17 | TargetJobs | GET | /api/v1/targets | listTargetJobs | PaginatedTargetJob |
| 18 | TargetJobs | GET | /api/v1/targets/{targetJobId} | getTargetJob | TargetJob |
| 19 | TargetJobs | PATCH | /api/v1/targets/{targetJobId} | updateTargetJob | TargetJob |
| 20 | TargetJobs | POST | /api/v1/targets/{targetJobId}/archive | archiveTargetJob | TargetJob（IK 必带；软归档） |
| 21 | PracticePlans | POST | /api/v1/practice/plans | createPracticePlan | PracticePlan |
| 22 | PracticePlans | GET | /api/v1/practice/plans/{planId} | getPracticePlan | PracticePlan |
| 23 | PracticeSessions | GET | /api/v1/practice/sessions | listPracticeSessions | PaginatedPracticeSession |
| 24 | PracticeSessions | POST | /api/v1/practice/sessions | startPracticeSession | PracticeSession |
| 25 | PracticeSessions | GET | /api/v1/practice/sessions/{sessionId} | getPracticeSession | PracticeSession |
| 26 | PracticeSessions | POST | /api/v1/practice/sessions/{sessionId}/events | appendSessionEvent | SessionEventResult |
| 27 | PracticeSessions | POST | /api/v1/practice/sessions/{sessionId}/complete | completePracticeSession | ReportWithJob |
| 28 | PracticeSessions | POST | /api/v1/practice/sessions/{sessionId}/voice-turns | createPracticeVoiceTurn | CreatePracticeVoiceTurnRequest / PracticeVoiceTurnResult（IK 必带） |
| 29 | Reports | GET | /api/v1/reports/{reportId} | getFeedbackReport | FeedbackReport |
| 30 | Reports | GET | /api/v1/targets/{targetJobId}/reports | listTargetJobReports | PaginatedFeedbackReport |
| 31 | ResumeTailor | POST | /api/v1/resume/tailor | requestResumeTailor | ResumeTailorRunWithJob |
| 32 | ResumeTailor | GET | /api/v1/resume/tailor-runs/{tailorRunId} | getResumeTailorRun | ResumeTailorRun |
| 33 | Jobs | GET | /api/v1/jobs/{jobId} | getJob | Job |
| 34 | Privacy | POST | /api/v1/privacy/exports | requestPrivacyExport | PrivacyRequestWithJob（P0 返回 501） |
| 35 | Privacy | POST | /api/v1/privacy/deletions | requestPrivacyDelete | PrivacyRequestWithJob |
| 36 | Privacy | GET | /api/v1/privacy/requests/{privacyRequestId} | getPrivacyRequest | PrivacyRequest |
| 37 | Auth | GET | /api/v1/runtime-config | getRuntimeConfig | RuntimeConfig（[A4 D-2](../secrets-and-config/spec.md#31-已锁定决策含-p0-必备-env-key-字典) owner） |

总计 37 个 endpoint，覆盖 10 tag。当前 freeze 包含 Auth 单入口资料补全、扁平 Resume operations、PDF source preview、TargetJob archive、ResumeTailor ephemeral suggestions、PracticeSessions listing / voice turn、Reports、Jobs 与 Privacy request contract。

> Auth single-entry profile completion (#2) 由 `backend-auth/001-email-code-session-bootstrap` 与 `frontend-shell/001-app-shell-auth-settings` 纳入 v1.0.0 freeze：`PATCH /api/v1/me completeMyProfile`、`UserContext.profileCompletionRequired`、`CompleteProfileRequest`、fixture、generated client/server artifact 与 inventory lint 已回填；真实 backend handler 由 backend-auth 承接。

> Practice sessions listing (#21) 保留为核心 practice recovery / session list contract；D-22 后不再承担 debrief picker 语义。

> Practice voice turn (#26) 由 `practice-voice-mvp/001-cascaded-stt-llm-tts` 纳入 v1.0.0 freeze：`PracticeSessions` tag 新增 side-effect `createPracticeVoiceTurn` operation、voice turn request/response schema、fixture、generated client/server artifact、inventory lint 与 fixture validator 已回填；真实 backend handler 由同计划 Phase 5 承接。

#### 3.1.2 B2 专属 async enum 字面量

`ResourceType` 与 `JobType` 不属于 B1 的 16 个共享业务 enum；它们由 B2 OpenAPI 独立锁定，当前 v1.0.0 字面量如下：

| Schema | 字面量 | 来源 |
|--------|--------|------|
| `ResourceType` | `target_job` / `feedback_report` / `resume_asset` / `resume_tailor_run` / `privacy_request` | 当前 B2 API-facing async resource set；`ai_task_runs.resource_type` / `async_jobs.resource_type` 必须兼容这些 API-facing resource names |
| `JobType` | `target_import` / `resume_parse` / `report_generate` / `resume_tailor` / `privacy_export` / `privacy_delete` | P0 API async job response set；DB 内部可保留 `source_refresh` / `email_dispatch` 等非 API-facing job type，但它们不得出现在 v1.0.0 `GET /api/v1/jobs/{jobId}` response 中，除非本 spec 修订 additive 追加 |

### 3.2 待确认事项

- v1.0.1 / v1.1.0 升级阈值：default 使用 SemVer，破坏性变更 → v2.0.0；v1.x 内累积 ≥ 5 个新 endpoint 触发 v1.1.0；具体由本 spec 修订时决策。
- SSE 子协议（练习会话流式 follow-up）：默认 P0 不上；如后续业务域提出，由本 spec 修订决策。

## 4 设计约束

### 4.1 状态码、Header 与幂等矩阵

| 契约项 | P0 锁定规则 | 例外 / 说明 |
|--------|-------------|-------------|
| 成功状态码 | `200` / `201` / `202` / `204` | 长耗时任务统一 `202 + Job`；删除 / logout 等无响应体成功使用 `204` |
| 客户端错误 | `400` / `401` / `403` / `404` / `409` / `422` / `429` | wire body 全部复用 B2 `ApiErrorResponse` envelope，内部 `error` 对象复用 B1 `ApiError`；`409` 覆盖状态冲突与幂等冲突 |
| 服务端错误 | `500` | 未分类内部错误；不得暴露 provider / prompt / secret 细节 |
| P0 显式例外 | 当前已落地 `501 Not Implemented` 仅允许 `POST /api/v1/privacy/exports` 与 `POST /api/v1/resumes/{resumeId}/exports` | privacy export 返回 `ApiErrorResponse.error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`；resume export 返回 `ApiErrorResponse.error.code = "RESUME_EXPORT_NOT_AVAILABLE"`，作用于扁平 `resumeId`；P1 将任一 endpoint 切回 `202 + *WithJob` 属于“预留能力变为可用”的兼容行为，不算 breaking change，但必须递增 spec/history、更新 fixture 与 release gate 例外记录 |
| Auth public endpoints | `/api/v1/auth/email/start`、`/api/v1/auth/email/verify`、`/api/v1/runtime-config` 不要求既有 session | auth start/verify 归 ADR-Q1；runtime-config 只能返回非敏感公开配置 |
| Protected endpoints | 除 public endpoints 外，P0 默认要求有效 first-party session cookie | `Authorization: Bearer` 不作为 P0 默认认证形态；如重新启用必须修订 ADR-Q1 与本 spec |
| Account deletion | `DELETE /api/v1/me` 是 protected endpoint，成功返回 `202 + PrivacyRequestWithJob` | 与 `POST /api/v1/privacy/deletions` 同义；必须支持 `Idempotency-Key` 或等价 active-request dedupe，重复删除请求返回同一未完成 `privacy_delete` job；先撤销 session / 软删用户，再由 backend internal runner 按 B4 table matrix 异步硬删 |
| Request headers | `X-Request-ID` / `traceparent` / `Accept-Language` / `X-Client-Version` 按本 spec 与 B1 当前 shared conventions 入 OpenAPI components | `Accept-Language` 只影响展示语言默认值，不覆盖 `targetLanguage` / `language` 等持久业务字段 |
| Idempotency-Key | 仅本 spec 标记的副作用 endpoint 必须声明并校验；B1 提供 key 格式与 TTL 工具语义 | `POST /practice/sessions/{sessionId}/events` 必须声明 `clientEventId` 去重；auth email start 使用 ADR-Q1 rate limit / challenge TTL，不挂通用 idempotency |

### 4.2 schema inventory 约束

| 类别 | 必须覆盖的 schema | 来源 / 约束 |
|------|-------------------|-------------|
| B1 shared | `ApiError` inner object、`PageInfo`、`Paginated<T>`、16 个枚举类型、错误码 enum、`IdempotencyKey` 工具语义 | `$ref` / codegen 复用 B1；OpenAPI 不重复维护 B1 enum 字面量；wire error body 另用 B2 `ApiErrorResponse` envelope |
| Auth / runtime | `UserContext`、`AuthEmailStartRequest`、`CompleteProfileRequest`、`AuthEmailVerifyQuery`、`Session`、`RuntimeConfig`、`DeleteMeResponse`（alias `PrivacyRequestWithJob`） | Auth 路径以 ADR-Q1 与 D-25 单入口邮箱验证码登录为准；`AuthEmailStartRequest` 不含 `purpose` / `displayName`，`UserContext.profileCompletionRequired` 是首次资料补全强制跳转信号；runtime-config 字段以 [A4 D-2](../secrets-and-config/spec.md#31-已锁定决策含-p0-必备-env-key-字典) 为准；`DELETE /me` 删除语义以 ADR-Q5 / B4 deletion matrix 为准 |
| Uploads / resumes | `UploadPresignRequest`、`UploadPresign`、`RegisterResumeRequest`、`Resume`、`ResumeWithJob`、`PaginatedResume`、`UpdateResumeRequest`、`DuplicateResumeRequest` | B2 owns request/response schema and fixture provenance；`Resume.structuredProfile` 承载结构化内容；`updateResume` 覆盖保存，`duplicateResume` 另存为新简历 |
| TargetJobs | `ImportTargetJobRequest`、`TargetJobWithJob`、`TargetJob`、`UpdateTargetJobRequest`、`TargetJobRequirement`、`TargetJobSummary`、`TargetJobFitSummary`、`PaginatedTargetJob` | 覆盖 URL / text / file / manual form source variants；`manual_form` 返回 terminal `target_import` job；`TargetJobRequirement.kind` 覆盖 `must_have` / `nice_to_have` / `hidden_signal` / `interview_focus` |
| Practice | `CreatePracticePlanRequest`、`PracticePlan`、`StartPracticeSessionRequest`、`PracticeSession`、`PracticeTurn`、`PracticeSessionEventRequest`、`SessionEventResult`、`AssistantAction`、`CompletePracticeSessionRequest`、`ReportWithJob` | `PracticeSessionEventRequest.clientEventId` 是事件幂等真理源；`PracticeTurn.status` wire enum 锁定为 `asked` / `answered` / `follow_up_requested` / `assessed` / `skipped` |
| Review / question review | `FeedbackReport`、`ReportHighlight`、`ReportIssue`、`ReportNextAction`、`QuestionAssessment`、`PaginatedFeedbackReport` | 报告前台只展示准备度档位、维度状态、题目回顾和本轮复练上下文；不输出精确通过率，不暴露独立错题本 endpoint |
| ResumeTailor | `RequestResumeTailorRequest`、`ResumeTailorRun`、`ResumeTailorRunWithJob` | 简历定制必须携带 provenance；`RequestResumeTailorRequest.resumeId` 指向扁平 resume，`targetJobId` 可选用于 JD-aware 改写上下文；`ResumeTailorRun.suggestions` 为 ephemeral，用户采纳后经 `updateResume`（覆盖）/ `duplicateResume`（另存）落盘；感谢信草稿与完整跟进建议字段在 P1 以前必须 optional / hidden，不得阻塞 P0 |
| Jobs / privacy | `Job`、`ResourceType`、`JobType`、`PrivacyRequest`、`PrivacyRequestWithJob`、`ApiErrorResponse` 501 example | privacy export P0 fixture 必须是 `501 + ApiErrorResponse.error.code = PRIVACY_EXPORT_NOT_AVAILABLE`；resume export P0 fixture 必须是 `501 + ApiErrorResponse.error.code = RESUME_EXPORT_NOT_AVAILABLE`；privacy deletion 保持 `202 + PrivacyRequestWithJob` |

每个 §3.1.1 endpoint 在 `openapi/openapi.yaml` 中必须同时声明 `operationId`、request body（若有）、success / P0 例外 response schema 与 error response `$ref`；缺任一项时 `make codegen-openapi` 或 inventory lint 不得通过。每个 operationId 的 default fixture 由 [002-fixtures-and-mock-source](./plans/002-fixtures-and-mock-source/plan.md) 交付，缺失 fixture 时 `make validate-fixtures` 不得通过；Prism / 文档站所需的 OpenAPI examples 必须由 fixtures 投影生成，并由 002 的 examples 同步门禁校验不漂移。

### 4.3 schema 设计约束

- 所有 enum 字段必须以 [B1 D-6 / D-10 枚举](../shared-conventions-codified/spec.md#31-已锁定决策) 中的 16 个类型为基础；本 spec 不重新定义 enum 字面量，必须 `$ref` 到 B1 共享 enum schema。
- `ApiError` schema 必须表示 B1 提供的 inner error object；`ApiErrorResponse` schema 必须是 `{error: ApiError}` envelope。`error.code` 字段定义为枚举（值集等于 [B1 D-5/D-10](../shared-conventions-codified/spec.md#31-已锁定决策) 全部错误码常量，含 `PRIVACY_EXPORT_NOT_AVAILABLE` 与 `RESUME_EXPORT_NOT_AVAILABLE`），由 generator 自动同步。
- 所有 `id` 字段为 `string`，`format: uuid`；服务端写入字段值必须 UUIDv7（由 B1 idx 工具生成）；前端临时 id（`tmp_<uuid>`）只在前端 state 中存在，不进 API 请求体。
- 所有时间字段统一 `string` + `format: date-time`；不允许某些字段使用 unix epoch number。
- 所有语言字段统一 BCP 47（如 `en` / `zh-CN` / `en-SG`）；OpenAPI schema 使用 `string` + pattern / example，实际允许集由产品 i18n 与质量评估 gate 控制。

### 4.4 breaking change linter 规则集（v1.0.0 freeze 后强制）

- **禁止**：删除已发布 endpoint / 重命名 path / 修改 method / 删除 schema 字段 / 修改字段类型 / 把 optional 字段改为 required / 删除已发布枚举值。
- **允许（additive）**：新增 endpoint / 新增 tag / 新增 optional 字段 / 新增枚举值（且字段为 string-typed enum） / 新增可选 query 参数 / 新增 example。
- **P0 例外**：`POST /api/v1/privacy/exports` 从 P0 `501 ApiErrorResponse` 切到 P1 `202 PrivacyRequestWithJob`、`POST /api/v1/resumes/{resumeId}/exports` 从 P0 `501 ApiErrorResponse` 切到 P1 `202 + Job(jobType=resume_export)`，均是已预留能力变为可用；该行为必须递增 spec/history 和 fixture，但不按 breaking change 处理。
- **审计要求**：违反规则的 PR 必须 attach ADR 链接并在本 spec history 表加一行「v2.0.0 升级」记录；远端 CI label workflow 仅在 A5 触发条件成立后再接入，当前单人阶段以本地 gate + owner review 为准。

### 4.5 codegen 与 drift 约束

- generator 输入：`openapi/openapi.yaml` + `openapi/templates/`（Go / TS 模板）；输出 `backend/internal/api/generated/` 与 `frontend/src/api/generated/`。
- generated 文件必须 idempotent；本地 `make codegen-check` / `git diff --exit-code` 阻塞漂移。远端 CI 接入由 A5 后续触发条件决定。
- 业务 handler 必须 implement generator 产出的 server interface；不允许业务包定义自己的 DTO 类型。

### 4.6 AI 生成结果 provenance 约束

OpenAPI 必须提供共享 `GenerationProvenance` schema，并要求所有 AI 生成结果直接包含该对象，或通过响应中的 `job` / `resource` 可追溯到该对象。字段固定为：

| 字段 | 说明 |
|------|------|
| `promptVersion` | prompt registry key / version |
| `rubricVersion` | rubric registry key / version；非评分生成也必须显式填 `not_applicable` |
| `modelId` | provider profile / model id，不暴露 secret |
| `language` | 本次生成使用的 BCP 47 语言 |
| `featureFlag` | 影响生成路径的 feature flag / variant |
| `dataSourceVersion` | 输入数据来源版本或 snapshot id |

至少以下 schema 必须包含或可追溯到 `GenerationProvenance`：`TargetJob.summary` / `fitSummary`、`AssistantAction`、`FeedbackReport`、`ResumeTailorRun`、`Resume`。`Resume` 通过 `structuredProfile.provenance` 覆盖 AI 抽取的结构化简历内容；tailor 改写建议 provenance 由 `ResumeTailorRun` 承载。缺失 provenance 的 fixture 不得通过 `make validate-fixtures`。

### 4.7 fixtures 与隐私约束

- `openapi/fixtures/<tag>/<operationId>.json` 必须 schema-valid（本地由 `make validate-fixtures` 校验；远端 CI 接入由 A5 后续触发条件决定）。
- fixtures 中绝不出现真实用户邮箱 / 真实电话 / 真实公司名敏感信息；统一用 `Acme` / `acme.example` / `alice@example.com`。
- `prototype-baseline` scenario 来自 `ui-design/src/data.jsx`；维护方式：`make sync-fixtures-from-prototype`（B2 owner）。
- TargetJob fixtures 必须覆盖实际用户场景，而不只保留 URL happy path：`importTargetJob` 至少维护 `default`（URL accepted）、`manual-text-primary`、`manual-form-ready-terminal-job`、`url-invalid-source`、`url-source-unavailable`；`getTargetJob` / `listTargetJobs` / `updateTargetJob` 至少覆盖 parsed ready、cross-user hidden 404、invalid state transition 三类 scenario。新增 scenario 必须通过 `make validate-fixtures`，并被对应 BDD 场景引用。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `openapi/openapi.yaml` 与 fixtures | B2 | 唯一真理源 |
| 16 个 enum 类型 / `ApiError` inner object / `PageInfo` schema | B1 | B2 通过 `$ref` 引用；B2 自身维护 `ApiErrorResponse` envelope；`QuestionReviewStatus` 与 Resume contract vocabulary 由 B1 当前 source 承接 |
| 错误码常量列表 | B1 | B2 在 `error.code` 枚举中同步 |
| Go 与 TS codegen | B2 + B1（generator base） | 输出落点固定 |
| 业务 handler 实现 | C 域各 owner | 必须 implement 生成的 server interface |
| 前端 fetch 客户端 | D 域各 owner | 必须使用生成的 TS client |
| mock server 运行壳 | E1 | 消费 fixtures |
| breaking change linter | B2 | 本地 gate；远端 CI 仅在 A5 触发条件成立后再接入 |
| API 文档生成 | B2（Redoc 集成） | 当前只保留本地产物，不要求 CI artifact |
| 鉴权 session 颁发 / 撤销 | C1 + ADR-Q1 | B2 仅锁 Auth tag HTTP contract 与 session cookie security scheme |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | OpenAPI 文档结构 | `openapi/openapi.yaml` 已落地 | `npx -p @apidevtools/swagger-cli@4.0.4 swagger-cli validate openapi/openapi.yaml` + inventory lint | 通过；含 10 tag、37 endpoint；每个 endpoint 有 request/success 或 P0 例外/error schema；`PATCH /api/v1/me` 返回 `200 + UserContext`，`DELETE /api/v1/me` 返回 `202 + PrivacyRequestWithJob`；`archiveTargetJob` 返回 archived `TargetJob` 并要求 `Idempotency-Key`；`ApiError` inner object / B1 shared schema 拓扑一致；Auth 路径与 ADR-Q1 `ei_session` 一致；fixture 完整性由 C-6 单独验证 | B2 001 / B2 004 / product-scope/001-core-loop-module-pruning / practice-voice-mvp 001 / backend-resume 002 / backend-auth 001 / backend-targetjob 001 Phase 12（contract/schema） |
| C-2 | Go codegen drift | 修改 `openapi.yaml` 但不跑 codegen | 本地 `make codegen-check` 或等价 gate | `codegen-drift-check` 失败；本地 diff 显示新增字段 | B2 001 |
| C-3 | TS codegen drift | 同 C-2 | 本地 `make codegen-check` 或等价 gate | `frontend/src/api/generated/` 漂移；本地 gate 失败 | B2 001 |
| C-4 | breaking change 拦截 | 故意删除 `target_jobs.title` 字段 | 本地 `make openapi-diff` / 等价 gate | `openapi-diff` 失败；除非已有 ADR + 本 spec 修订授权，否则不得继续 | B2 003 |
| C-5 | additive 通过 | 给 `practice_plans` 新增 `optional metadata` 字段 | 本地 `make openapi-diff` / 等价 gate | `openapi-diff` 仅警告 additive；测试通过 | B2 003 |
| C-6 | fixtures 一致 | 任一 endpoint 缺少 fixtures | `make validate-fixtures` | 失败；列出缺失 operationId | B2 002 |
| C-7 | privacy export 501 | P0 调用 `POST /api/v1/privacy/exports` | E1 mock + 后续 C12 实现 | 返回 501 + `error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"` | B2 002（fixture）+ C12 P1 实现 |
| C-7a | account deletion endpoint | P0 登录用户调用 `DELETE /api/v1/me` | E1 mock + 后续 backend internal privacy runner | 返回 `202 + PrivacyRequestWithJob`，`job.jobType="privacy_delete"`；重复请求返回同一 active 删除 job 或同义终态；与 `POST /api/v1/privacy/deletions` fixture 的语义一致 | B2 001 + B2 002 + backend-runtime-topology |
| C-8 | enum 与 B1 同源 | 在 `openapi.yaml` 引用 `practiceMode` enum | codegen | 生成 TS 与 Go 类型，与 [B1 D-6](../shared-conventions-codified/spec.md#31-已锁定决策) 完全一致；改 B1 后 B2 codegen drift | B2 001 + B1 |
| C-9 | mock 同源（前端 + 后端） | E1 拉起 mock server | 前端 msw 与后端 mock-server 都消费 `openapi/fixtures/` | 同一 endpoint 两端响应字节级一致；B2 002 先证明 fixture → OpenAPI example → Prism response 的 default scenario 字节级一致 | B2 002（partial）+ E1 |
| C-10 | B2 executable freeze handoff | 本 spec 的 contract lock 已完成，B2 001 / 002 / 003 均完成 | 当前 active spec 关系已保留 | `openapi/openapi.yaml` v1.0.0、codegen drift、fixtures 与 breaking-change linter 均通过验证；依赖 B2 的后续 implementation 可启动；roadmap 只保留 active spec 关系，不单独冒充本项已通过 | B2 003（汇总 001 / 002 证据） |
| C-11 | provenance 完整性 | 任一 AI 生成 response fixture 缺少 `GenerationProvenance` 或不可追溯到含 provenance 的 job/resource | `make validate-fixtures` | 失败；列出 operationId 与缺失字段；001 只锁 schema 可追溯关系，fixture 内容由 002 验证 | B2 001（schema）+ B2 002（fixtures）+ F3 |
| C-12 | resume export 501 例外 | D-26 已落地 `exportResume` operation | `make lint-openapi` + `make validate-fixtures` | `exportResume` 允许 P0 `501 + ApiErrorResponse.error.code="RESUME_EXPORT_NOT_AVAILABLE"`；除 `requestPrivacyExport` / `exportResume` 外的 endpoint 返回 501 会被 inventory lint 拒绝；未来切到 `202 + Job(jobType=resume_export)` 必须递增 spec/history 与 fixture | openapi-v1-contract/004 |

## 7 关联计划

B2 当前由本 spec 保留 active contract lock；真实 executable contract 由 B2 自身 plans 与当前 product-scope owner 承接（[engineering-roadmap §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec)）：

- `001-bootstrap`：落地 `openapi/openapi.yaml` 框架、ADR-Q1 Auth 路径、`DELETE /api/v1/me` privacy deletion endpoint、privacy export 501 例外、B1 enum `$ref`、`GenerationProvenance`、`make codegen-openapi` 与本地 drift check；当前 freeze 为 10 tag / 37 endpoint。
- `002-fixtures-and-mock-source`：每个 operationId 一份 fixtures + `prototype-baseline` 同步工具；E1 接入。
- `003-breaking-change-gate`：linter 规则集 + ADR 模板；远端 CI label workflow 仅在 A5 触发条件成立后再评估。
- `004-resume-additive-coverage`：承接当前扁平 Resume contract、`RESUME_EXPORT_NOT_AVAILABLE`、Resumes / ResumeTailor fixtures、codegen drift、inventory lint 与 TS client response typing gate。
- `backend-auth/001-email-code-session-bootstrap` + `frontend-shell/001-app-shell-auth-settings`：D-25 Auth single-entry profile completion pre-launch correction 落地：`AuthEmailStartRequest` 不含 `purpose` / `displayName`，`UserContext` 新增 `profileCompletionRequired`，Auth tag 包含 `PATCH /api/v1/me completeMyProfile` + `CompleteProfileRequest` + fixture + Go/TS generated artifacts；真实 handler 由 backend-auth 承接。
- `product-scope/001-core-loop-module-pruning`：当前核心闭环范围 owner，负责 10 tag / 37 endpoint freeze、baseline、fixtures、generated artifacts、runtime route-negative tests 与 active spec zero-reference gates。
- `practice-voice-mvp/001-cascaded-stt-llm-tts`：D-22 Practice voice turn additive 升级落地：`PracticeSessions` tag 包含 `createPracticeVoiceTurn` operation + voice turn request/response schema + fixture + Go/TS generated artifacts；真实 handler 由 practice voice Phase 5 补齐。

在放行依赖 B2 的后续业务实现前，必须确认 B2 001/002/003/004 与 product-scope owner gate 已补齐 `openapi/openapi.yaml`、fixtures、baseline、generated artifacts 与 diff whitelist，不得只停留在本 spec 文本。

后续如出现 v1.1.0 / v2.0.0 升级：递增 spec 版本 + history；每次升级在 §3.1.1 中保留 endpoint 完整快照。
