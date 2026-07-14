# OpenAPI v1 Contract Spec

> **版本**: 1.60
> **状态**: active
> **更新日期**: 2026-07-14

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将原始 B2 `openapi-v1-contract` 保留为当前 active Contract spec（依赖 [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md)；间接依赖 [A1 `repo-scaffold`](../repo-scaffold/spec.md)）。它是当前 P0 backend / frontend workstream 的 HTTP 契约瓶颈节点：后续实现必须复用本契约的 codegen、fixtures 与 breaking-change gate；任何破坏性变更会触发跨 spec 雪球。

本 spec 由 `engineering-roadmap/001-decompose-subspecs` 的 contract lock 创建；当前执行口径以 roadmap active spec 的保留规则为准：`openapi/openapi.yaml` v1.0.0 freeze 范围为当前 37 endpoints / 10 tags、字段命名和 additive-only 规则。真实 OpenAPI 文件、codegen、fixtures 与 breaking-change linter 由 B2 `001-bootstrap` / `002-fixtures-and-mock-source` / `003-breaking-change-gate` / `004-resume-additive-coverage` 分别验证；未通过前不得启动依赖 B2 的 implementation。

当前 HTTP 可执行契约由本 spec、`openapi/openapi.yaml`、OpenAPI fixtures / baseline 与 B1 shared-conventions-codified 决定。B2 独立承接 endpoint inventory、tag、auth 形态、header、status code、schema、fixture provenance 与 breaking-change gate；任何实现或 codegen 都不得绕过这些当前 owner truth source。

目标是：

1. **唯一真理源**：`openapi/openapi.yaml` 是 P0 所有 HTTP 端点的唯一定义；任何脱离 codegen 的 handler surface / 手写 fetch 客户端禁止与之偏离。
2. **双端 codegen**：Go DTO + chi handler 接口在 `backend/internal/api/generated/`；TypeScript SDK 只生成 `frontend/src/api/generated/client.ts` 与 `types.ts`，不复制 raw OpenAPI 文本；本地 `make codegen-openapi` / `make codegen-check` 必须能用 `git diff --exit-code` 校验未漂移（与 [B1 D-1 idempotent generator](../shared-conventions-codified/spec.md#31-已锁定决策) 一致）。
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
| D-6 | Idempotency | 副作用 endpoint 的 `Idempotency-Key` 支持范围由本 spec §4.1 与 B1 幂等工具共同决定；`sendPracticeMessage` 使用 body `clientMessageId` 去重，不混用 `Idempotency-Key` | 同一用户消息可安全重试且不会生成重复 assistant reply |
| D-7 | Job 异步 | 长耗时操作返回 `202 Accepted` + `Job` schema；客户端通过 `GET /jobs/{jobId}` 轮询 | – |
| D-8 | content-type | 仅 `application/json` 与 `multipart/form-data`（仅 upload 端点）；不引入 protobuf / msgpack | – |
| D-9 | v1.0.0 freeze 范围 | §3.1.1 列出当前 37 个 endpoint + 10 tag；本 spec 锁定范围与 additive-only 规则，B2 `001` 落地 `openapi/openapi.yaml` 后强制执行（新增 endpoint / 新增可选字段 / 新增枚举值）；Auth tag 以 ADR-Q1 的 email-code challenge + session cookie 路径为准，`startAuthEmailChallenge` 只接收邮箱和 `returnTo`，不得用 `purpose` 或 `displayName` 区分注册/登录；`verifyAuthEmailChallenge` 消费 query `token` 但其语义为 6 位 code；`DELETE /api/v1/me` 按 ADR-Q5 纳入 P0 删除入口；`listPracticeSessions`、`createPracticeVoiceTurn`、`completeMyProfile`、扁平 Resume operations、`duplicateResume`、`getResumeSource` 和 `archiveTargetJob` 均属于当前 freeze | 任何 break change 必须 ADR + 本 spec 修订；当前 project pre-launch 阶段允许由 product-scope owner 授权 v1.0.0 freeze correction，并必须同步 history、baseline、fixtures、generated artifacts 和 diff-config |
| D-10 | breaking change linter | 默认 `openapi-diff`（OpenAPITools）；规则：禁止删字段、禁止改字段类型、禁止改 required、禁止改枚举（仅允许新增）、禁止删 endpoint | 本地 gate 直接失败；远端 CI 接入由 A5 后续触发条件决定 |
| D-11 | tags 顺序 | §2.1 10 个 tag 顺序固定；新增 tag 必须递增 spec | – |
| D-12 | privacy export 例外 | 按 [ADR-Q5](../engineering-roadmap/decisions/ADR-Q5-privacy-cadence.md)，`POST /api/v1/privacy/exports` 在 v1.0.0 freeze 中保留路径与 schema，但 P0 必须返回 `501 unavailable response`（HTTP status 仍为 501，`error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`）；P1 切换实现时是 additive 行为变化，不算 break | 防止 P1 复用时改路径 |
| D-13 | OpenAPI tooling 锁版 | validation: `@apidevtools/swagger-cli@4.0.4`；docs: `@redocly/cli@2.30.1 build-docs`。禁止未修订 spec 时替换 C-1 validator | 避免 002 / 003 在不同 validation gate 间产生不一致错误面；docs renderer 升级必须记录实测兼容证据 |
| D-14 | B2 专属 async enum 字面量 | `ResourceType` 与 `JobType` 独立成 OpenAPI schema；字面量见 §3.1.2。它们来自当前 B2 API-facing async response set 与 P0 privacy exception，后续新增 endpoint / async job 时必须递增本 spec 并 additive 追加 enum 值 | 不再把 `ResourceType` 留作待确认；fixtures / mock / generated DTO 可直接依赖 |
| D-15 | 错误响应 envelope | B1 `ApiError` 表示 `error` inner object；B2 `ApiErrorResponse` 表示 wire body `{error: ApiError}`。所有 default 4xx/5xx 与 privacy export 501 响应使用 `ApiErrorResponse` envelope | 消除 Go/TS codegen 对 `ApiError` 名称的歧义 |
| D-16 | TargetJob import 场景契约 | accepted [OPENAPI-002](./decisions/OPENAPI-002-targetjob-paste-only.md)：`importTargetJob` 只接受 closed flattened `{rawText,targetLanguage,resumeId}`；`rawText` 同时使用 `minLength: 1` 与 `pattern: '\S'` 拒绝空字符串和纯空白；继续返回 `202 + TargetJobWithJob`。删除 URL/file/manual-form source union、title/company hints、`TargetJob.sourceType/sourceUrl` 与 `target_job_attachment` upload purpose。标题、公司和结构化要求只由服务端从粘贴文本解析。 | 这是未上线 v1.0.0 freeze correction；001/002/003、mock 与 frontend/backend consumer 必须同批迁移，不保留 discriminator、alias 或兼容分支；37 operation / 10 tag inventory 与通用 `createUploadPresign` 保持不变 |
| D-19 | Practice conversation pre-launch rebase | 删除 `PracticeTurn`、`PracticeSession.currentTurn/turnCount`、question budget、event answer/hint action 和 AssistantAction；新增 `PracticeMessage`、`SendPracticeMessageRequest/Result`，并以 `clientMessageId` 作为消息 replay key。 | v1.0.0 尚未发布，baseline / fixture / generated artifacts 原地同步，不保留兼容字段或旧 endpoint |
| D-21 | Practice sessions listing | 当前 37-operation freeze 包含 `GET /api/v1/practice/sessions` operationId `listPracticeSessions`，query 为 `targetJobId?` / `status?` / `cursor?` / `pageSize?`，response schema 为 `PaginatedPracticeSession`。该 endpoint 是 read-only，不挂 `Idempotency-Key`。 | `openapi/openapi.yaml`、`openapi/fixtures/PracticeSessions/listPracticeSessions.json`、`scripts/lint/openapi_inventory.py`、Go/TS generated artifacts 与 codegen test 保持同步 |
| D-22 | Practice voice disabled | 继续保留 `createPracticeVoiceTurn` 路径作为 typed disabled 边界，但当前只返回 `AI_UNSUPPORTED_CAPABILITY`；request 不得进入 provider/persistence happy path，正向 voice event kinds 不属于当前 contract。 | fixture 只保留 disabled negative scenario；重新启用必须修订 Product/UI/B2/privacy/provider owner |
| D-30 | Conversation-level report rebase | `FeedbackReport` 使用 `dimensionAssessments` 与 `retryFocusCompetencyCodes`；删除 `QuestionAssessment`、`questionAssessments`、`retryFocusTurnIds` 和 `QuestionReviewStatus`。 | report fixtures、generated DTO、backend-review、frontend dashboard 与 baseline 同步 |
| D-31 | Practice 轮次身份与进度投影 | `CreatePracticePlanRequest` additive 新增可选 `roundId` 作为客户端轮次意图；`PracticePlan` additive 返回 `roundId` / `roundSequence`，新建记录必须同时有值，legacy 无身份记录可为 null；`roundSequence` 是 `1..2147483647` 的 int32，logical `roundId` 的 sequence 段最多 10 位并须由服务端与 pair 精确校验；`TargetJob` additive 返回 `practiceProgress`，其中 `completedRounds` 与 `currentRound` 由已完成 session 台账投影，禁止把 TargetJob lifecycle `status` 当作面试轮次。 | 不新增进度 endpoint 或第二份可变状态；超 int32/错误 pair fail closed；fixtures、baseline、Go/TS generated artifacts、backend-practice、backend-targetjob 与 frontend workspace 同步 |
| D-32 | Grounded direct report pre-release correction | accepted [OPENAPI-001](./decisions/OPENAPI-001-report-direct-semantics.md)：FeedbackReport 必带 nullable-until-ready `summary` / `preparednessLevel` / `provenance` 与非空 frozen `context`，维度/证据使用 closed `code + label + dimensionCode`，focus 改为 `retryFocusDimensionCodes`；`ready` 必须同时提供 non-null summary/preparedness/provenance 与非空 dimensions/actions，`failed` 必须提供 non-null errorCode，其余状态 errorCode 必须 null。删除旧 dimension 字段、`DimensionResult`、request `focusCompetencyCodes` 与兼容层。CreatePracticePlanRequest 是 typed closed conditional object：baseline 禁止 sourceReportId，derived 两个 goal 均只允许并要求 non-null UUID sourceReportId。B1 同批新增 non-retryable `REPORT_CONTEXT_TOO_LARGE` 并 additive 同步到 `ApiErrorCode`。 | 这是未上线 v1.0.0 freeze correction；001/002/003 同批重开，先用 merge-base 旧 baseline 按 `severity+path+kind+before+after` 审计机器 oracle exact set，再同步 fixtures/codegen/consumers并以 report state + baseline/derived positive/negative matrix 证明 conditional contract，最后原地 re-freeze current baseline。 |
| D-33 | Report action / retry responsibility | `ReportNextAction.label.minLength=1/maxLength=200`只作wire/schema malformed fuse。24/64与18/52由F3/runtime/frontend承接。Generation/judge各自最多4调用、dynamic scope、attempt/retry/reason/scope与业务/infra backoff均是内部合同；HTTP只保留`queued/generating/ready/failed`和既有payload/error/provenance，不新增attempt fields、client retry endpoint或progress。客户端窗口耗尽不能伪造server failed。 |
| D-34 | TargetJob paste-only pre-release correction | OPENAPI-002 的 old-baseline oracle 必须精确记录 17 个 authorized breaking findings：删除五个 source schemas、source/title/company request properties、TargetJob source fields、`target_job_attachment` enum value，以及 source-only `TARGET_IMPORT_SOURCE_INVALID` / `TARGET_IMPORT_SOURCE_UNAVAILABLE` 两个 `ApiErrorCode` enum values；新增 constrained required `rawText`、关闭 request object，并更新两个 required sets。新 property 的初始 `minLength=1,pattern=\S` 归一到同一个 `required_property_added.after`，不另增 finding；wrapper RED 必须拒绝 stale 15-finding oracle，GREEN 锁 exact 17。`VALIDATION_FAILED` 与 `TARGET_IMPORT_FAILED` 保留；`importTargetJob` method/path/operationId/status/response、`createUploadPresign` method/path/operationId/201/response 与 37/10 inventory 是不变量。 | 003 必须在 baseline 编辑前保存 merge-base exact finding artifact；001/002 与所有消费者完成 paste-only positive/negative gate及旧能力 positive/runtime zero-reference 后才允许 re-freeze。ADR、oracle 和显式 negative declarations 可保留旧 token，禁止整目录豁免。 |
| D-35 | Practice message durable recovery（方案 A） | `PracticeMessage` 改为 role-discriminated closed union：user projection 必须携带原始 `clientMessageId` 与 `replyStatus=pending|retryable_failed|terminal_failed|complete`，assistant projection 必须禁止这两个字段。`getPracticeSession` 是 reload 后的权威投影；AI failure 后保留同一 user message，只有 `retryable_failed` 可用相同 `clientMessageId` 与原文重试，且重放不得重复 user/assistant message。 | reply status 是后端持久化/read-side 事实，不得由 localStorage、URL 或“是否存在 assistant”临时推断；`sendPracticeMessage` / fixtures / Go+TS codegen / backend-practice / frontend-workspace-and-practice 同批迁移。Generated TS client 必须抛 typed `ApiClientError`，消费者禁止解析 `message`。D-35 + history 1.54 + 产品批准的方案 A 是唯一治理 authority；独立 Practice machine oracle 只是该决策的可执行 finding 投影，不创建第三个 `OPENAPI-NNN` ADR，也不得并入 OPENAPI-002。 |
| D-36 | TargetJob canonical-round report overview | accepted [OPENAPI-004](./decisions/OPENAPI-004-targetjob-report-overview.md)：原地保留 `GET /targets/{targetJobId}/reports` 与 operationId `listTargetJobReports`，删除 cursor/pageSize 和 flat `PaginatedFeedbackReport` response，改为 closed `TargetJobReportsOverview`。每个 canonical round 返回 required `round: PracticeRoundRef`、nullable `currentReport={id,generatedAt}` 与 nullable `latestAttempt={id,status,errorCode,createdAt}`；`TargetJob.latestReportId` 同批删除。 | rounds 覆盖 owned TargetJob 当前 canonical catalog 全量并按 sequence 排序；current ready 与 latest attempt 独立选择，较新的 queued/generating/failed 不替换旧 ready。所有 report 必须用 frozen context 精确匹配当前 TargetJob/round，缺失或非法时整份 fail closed；不保留 pagination 或 pointer 兼容层，37/10 与 endpoint method/path/operationId/status 不变。 |
| D-37 | Resume list summary projection | accepted [OPENAPI-005](./decisions/OPENAPI-005-resume-list-summary.md)：`listResumes` method/path/operationId/200 与 `PaginatedResume` pagination envelope 不变，`items` 从完整 `Resume` 改为 closed `ResumeSummary`；required 字段精确为 `id/title/displayName/language/sourceType/parseStatus/summaryHeadline/hasReadableContent/updatedAt`。`summaryHeadline` 按 `parsed_summary.headline` → `parsed_summary.basics.headline` → `structured_profile.headline` → `structured_profile.basics.headline` 取首个 trim 后非空 string；`hasReadableContent=true` 当且仅当 trim 后 `parsed_text_snapshot` 或 `original_text` 非空，或 `structured_profile` 是非空 object。`getResume` 继续返回完整 `Resume`。 | list 禁止 `fileObjectId/originalText/parsedTextSnapshot/parsedSummary/structuredProfile/createdAt/deletedAt/status/provenance` 与任意额外字段；不得按 `fileObjectId/sourceType/parseStatus` 猜测可读性，frontend 不通过详情字段或额外 fetch 推断。001/002/003/004、backend/frontend/mock/BDD 全部同批迁移，不保留 alias、兼容字段或第二个列表 endpoint。 |
| D-38 | Runtime content-limit projection | accepted [OPENAPI-006](./decisions/OPENAPI-006-runtime-content-limits.md)：`RuntimeConfig` 新增 required `contentLimits`，引用 closed required `ContentLimits`；字段精确为 `resumeUploadBytes/resumePasteTextBytes/targetJobRawTextBytes/practiceMessageBytes/practiceSessionTextBytes` 五个 positive int64。 | 这是用户于 2026-07-14 批准方案 A 与修订默认值后的未上线 v1.0.0 freeze correction；只公开前端预检所需字段，report/HTTP/provider/profile 内部限制不得泄漏。001/003、fixture、Go/TS generated、backend builder 与 Resume/Home/Practice consumers 同批迁移；先保存 exact finding artifact，再原地 re-freeze。 |
| D-25 | Auth single-entry profile completion | 当前 37-operation freeze 包含 protected `PATCH /api/v1/me` operationId `completeMyProfile`；邮箱是唯一账号标识，用户只从 `startAuthEmailChallenge` 发起同一个邮箱验证码登录入口；`AuthEmailStartRequest` 不使用 `purpose` / `displayName` 区分注册/登录，避免发码前泄露邮箱存在性；`UserContext.profileCompletionRequired` 标识首次登录资料未完成。`displayName` 不唯一，不参与账号去重；资料未完成账号每次登录后必须先进入资料补全。 | `openapi/openapi.yaml`、Auth fixtures、inventory lint、Go/TS generated artifacts、frontend-shell/001 与 backend-auth/001 保持同步；真实登录流程由业务 owner 独立验收 |
| D-26 | 简历资产扁平化 contract collapse（product-scope D-20） | 当前 freeze 中，Resumes tag 使用单一扁平 `Resume` 详情实体、`resumeId` 路径参数、`UpdateResumeRequest` 覆盖保存和 `DuplicateResumeRequest` 另存为新简历；ResumeTailor 生成 ephemeral suggestions，采纳后通过 `updateResume` 或 `duplicateResume` 落盘；`RegisterResumeRequest.sourceType` 为 `upload` / `paste`；`Resume.structuredProfile` 承载结构化内容，`AI_PROVENANCE_SCHEMAS` 覆盖 `Resume`。列表 read model 由 D-37 的 `ResumeSummary` 单独约束，不恢复版本树。 | `openapi/openapi.yaml`、`openapi/fixtures/Resumes/*`、`openapi/fixtures/ResumeTailor/*`、`scripts/lint/openapi_inventory.py`、`scripts/lint/validate_fixtures.py`、`openapi/README.md`、`openapi/fixtures/README.md`、`openapi/baseline/openapi-v1.0.0.yaml`、Go/TS generated client/server/types、`docs/spec/mock-contract-suite` / `docs/spec/engineering-roadmap` 计数同步；由 [openapi-v1-contract/004](./plans/004-resume-additive-coverage/plan.md) 与 product-scope current contract owner 一并审查 |
| D-27 | 核心闭环 contract boundary（product-scope D-22） | 当前 freeze 为 **37 endpoint / 10 tag**。账号资料补全由 Auth `completeMyProfile` 承接；practice 派生计划只接受 report-derived `retry_current_round` / `next_round`；Reports / PracticeSessions / Resumes / ResumeTailor / TargetJobs / Privacy 共同承接当前核心闭环。 | OpenAPI source、fixtures、inventory/validation lint、baseline、Go/TS generated artifacts 与 frontend/backend consumers 同步；真实 API/UI 核心闭环由 product-scope owner 独立验收。 |
| D-29 | TargetJob archive | 当前 freeze 包含 protected `POST /api/v1/targets/{targetJobId}/archive` operationId `archiveTargetJob`，用于 workspace 删除图标持久软归档 TargetJob。该 endpoint 是 side-effect operation，必须声明 `Idempotency-Key`，成功返回 archived `TargetJob`，read-side 继续通过 `deleted_at is null` 隐藏归档记录。 | OpenAPI source、TargetJobs fixture、inventory lint、Go/TS generated artifacts、backend-targetjob/001 与 frontend-workspace-and-practice/001 保持同步 |
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
| 26 | PracticeSessions | POST | /api/v1/practice/sessions/{sessionId}/messages | sendPracticeMessage | SendPracticeMessageRequest / SendPracticeMessageResult |
| 27 | PracticeSessions | POST | /api/v1/practice/sessions/{sessionId}/complete | completePracticeSession | ReportWithJob |
| 28 | PracticeSessions | POST | /api/v1/practice/sessions/{sessionId}/voice-turns | createPracticeVoiceTurn | CreatePracticeVoiceTurnRequest / PracticeVoiceTurnResult（IK 必带） |
| 29 | Reports | GET | /api/v1/reports/{reportId} | getFeedbackReport | FeedbackReport |
| 30 | Reports | GET | /api/v1/targets/{targetJobId}/reports | listTargetJobReports | TargetJobReportsOverview |
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
| `JobType` | `target_import` / `resume_parse` / `report_generate` / `resume_tailor` / `privacy_export` / `privacy_delete` | P0 API async job response set；DB/backend runner 只额外保留 internal-only `email_dispatch`。`source_refresh` 已由 B3 D-20 删除，不得作为当前 job type 回流；`email_dispatch` 不得出现在 v1.0.0 `GET /api/v1/jobs/{jobId}` response 中，除非本 spec 修订 additive 追加 |

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
| P0 显式例外 | 当前已落地 `501 unavailable response` 仅允许 `POST /api/v1/privacy/exports` 与 `POST /api/v1/resumes/{resumeId}/exports` | privacy export 返回 `ApiErrorResponse.error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`；resume export 返回 `ApiErrorResponse.error.code = "RESUME_EXPORT_NOT_AVAILABLE"`，作用于扁平 `resumeId`；P1 将任一 endpoint 切回 `202 + *WithJob` 属于“预留能力变为可用”的兼容行为，不算 breaking change，但必须递增 spec/history、更新 fixture 与 release gate 例外记录 |
| Auth public endpoints | `/api/v1/auth/email/start`、`/api/v1/auth/email/verify`、`/api/v1/runtime-config` 不要求既有 session | auth start/verify 归 ADR-Q1；runtime-config 只能返回非敏感公开配置 |
| Protected endpoints | 除 public endpoints 外，P0 默认要求有效 first-party session cookie | `Authorization: Bearer` 不作为 P0 默认认证形态；如重新启用必须修订 ADR-Q1 与本 spec |
| Account deletion | `DELETE /api/v1/me` 是 protected endpoint，成功返回 `202 + PrivacyRequestWithJob` | 与 `POST /api/v1/privacy/deletions` 同义；必须支持 `Idempotency-Key` 或等价 active-request dedupe，重复删除请求返回同一未完成 `privacy_delete` job；先撤销 session / 软删用户，再由 backend internal runner 按 B4 table matrix 异步硬删 |
| Request headers | `X-Request-ID` / `traceparent` / `Accept-Language` / `X-Client-Version` 按本 spec 与 B1 当前 shared conventions 入 OpenAPI components | `Accept-Language` 只影响展示语言默认值，不覆盖 `targetLanguage` / `language` 等持久业务字段 |
| Idempotency-Key | 仅本 spec 标记的副作用 endpoint 必须声明并校验；B1 提供 key 格式与 TTL 工具语义 | `sendPracticeMessage` 使用 body `clientMessageId`，不挂 header IK；auth email start 使用 ADR-Q1 rate limit / challenge TTL |

### 4.2 schema inventory 约束

| 类别 | 必须覆盖的 schema | 来源 / 约束 |
|------|-------------------|-------------|
| B1 shared | `ApiError` inner object、`PageInfo`、`Paginated<T>`、当前共享枚举、错误码 enum、`IdempotencyKey` 工具语义 | `$ref` / codegen 复用 B1；OpenAPI 不重复维护 B1 enum 字面量；wire error body 另用 B2 `ApiErrorResponse` envelope |
| Auth / runtime | `UserContext`、`AuthEmailStartRequest`、`CompleteProfileRequest`、`AuthEmailVerifyQuery`、`Session`、`RuntimeConfig`、`ContentLimits`、`DeleteMeResponse`（alias `PrivacyRequestWithJob`） | `ContentLimits` 是 closed required object，精确含 `resumeUploadBytes`、`resumePasteTextBytes`、`targetJobRawTextBytes`、`practiceMessageBytes`、`practiceSessionTextBytes` 五个 positive int64；runtime-config 不公开 report/HTTP/provider/profile limits；其余字段以 A4 D-2 为准 |
| Uploads / resumes | `UploadPresignRequest`、`UploadPresign`、`RegisterResumeRequest`、`ResumeSummary`、`Resume`、`ResumeWithJob`、`PaginatedResume`、`UpdateResumeRequest`、`DuplicateResumeRequest` | B2 owns request/response schema and fixture provenance；`PaginatedResume.items` 只引用 D-37 closed `ResumeSummary`，`getResume` 继续返回 full `Resume`；`Resume.structuredProfile` 承载结构化内容；`updateResume` 覆盖保存，`duplicateResume` 另存为新简历 |
| TargetJobs | `ImportTargetJobRequest`、`TargetJobWithJob`、`TargetJob`、`UpdateTargetJobRequest`、`TargetJobRequirement`、`TargetJobSummary`、`TargetJobFitSummary`、`PracticeProgress`、`PaginatedTargetJob` | `ImportTargetJobRequest` 只允许 required `rawText` / `targetLanguage` / `resumeId`；`rawText` 必须 `minLength: 1` 且 `pattern: '\S'`；不得出现 source wrapper、URL、file、manual form 或客户端 title/company hints；`TargetJobRequirement.kind` 覆盖 `must_have` / `nice_to_have` / `hidden_signal` / `interview_focus`；structured round TargetJob runtime 必须返回 backend ledger projection |
| Practice | `CreatePracticePlanRequest`、`PracticePlan`、`PracticeRoundRef`、`StartPracticeSessionRequest`、`PracticeSession`、role-discriminated `PracticeMessage`（user / assistant）、`PracticeReplyStatus`、`SendPracticeMessageRequest`、`SendPracticeMessageResult`、`CompletePracticeSessionRequest`、`ReportWithJob` | plan/session/message schemas 不含 question/mode/hint；user message 必带 `clientMessageId` 和 persisted `replyStatus`，assistant message 禁止二者；session 返回 ordered messages；same-ID 是唯一 replay key；round identity 为 paired logical key + sequence |
| Review | `FeedbackReport`、`ReportContextSnapshot`、`DimensionAssessment`、`ReportHighlight`、`ReportIssue`、`ReportNextAction`、`TargetJobReportsOverview` 及其轻量 round summaries | closed report schemas；详情输出 grounded direct semantics，TargetJob overview 只输出 canonical round identity、current ready locator/timestamp 与 latest attempt status/error/timestamp，不输出 full report、provenance/model/rubric 或内部 locator |
| ResumeTailor | `RequestResumeTailorRequest`、`ResumeTailorRun`、`ResumeTailorRunWithJob` | 简历定制必须携带 provenance；`RequestResumeTailorRequest.resumeId` 指向扁平 resume，`targetJobId` 可选用于 JD-aware 改写上下文；`ResumeTailorRun.suggestions` 为 ephemeral，用户采纳后经 `updateResume`（覆盖）/ `duplicateResume`（另存）落盘；感谢信草稿与完整跟进建议字段在 P1 以前必须 optional / hidden，不得阻塞 P0 |
| Jobs / privacy | `Job`、`ResourceType`、`JobType`、`PrivacyRequest`、`PrivacyRequestWithJob`、`ApiErrorResponse` 501 example | privacy export P0 fixture 必须是 `501 + ApiErrorResponse.error.code = PRIVACY_EXPORT_NOT_AVAILABLE`；resume export P0 fixture 必须是 `501 + ApiErrorResponse.error.code = RESUME_EXPORT_NOT_AVAILABLE`；privacy deletion 保持 `202 + PrivacyRequestWithJob` |

`RuntimeConfig.contentLimits` 是 runtime public projection，不把数值复制到各业务 request schema 的静态 `maxLength`/`maximum` 作为可覆盖配置真理源。业务 request wire 保持不变；backend domain validation 仍是权威边界。

每个 §3.1.1 endpoint 在 `openapi/openapi.yaml` 中必须同时声明 `operationId`、request body（若有）、success / P0 例外 response schema 与 error response `$ref`；缺任一项时 `make codegen-openapi` 或 inventory lint 不得通过。每个 operationId 的 default fixture 由 [002-fixtures-and-mock-source](./plans/002-fixtures-and-mock-source/plan.md) 交付，缺失 fixture 时 `make validate-fixtures` 不得通过；Prism / 文档站所需的 OpenAPI examples 必须由 fixtures 投影生成，并由 002 的 examples 同步门禁校验不漂移。

### 4.3 schema 设计约束

- 所有 enum 字段必须以 [B1 D-6 / D-10 枚举](../shared-conventions-codified/spec.md#31-已锁定决策) 中的 16 个类型为基础；本 spec 不重新定义 enum 字面量，必须 `$ref` 到 B1 共享 enum schema。
- `ApiError` schema 必须表示 B1 提供的 inner error object；`ApiErrorResponse` schema 必须是 `{error: ApiError}` envelope。`error.code` 字段定义为枚举（值集等于 [B1 D-5/D-10](../shared-conventions-codified/spec.md#31-已锁定决策) 全部错误码常量，含 `PRIVACY_EXPORT_NOT_AVAILABLE` 与 `RESUME_EXPORT_NOT_AVAILABLE`），由 generator 自动同步。
- 除 `roundId` 外，所有 `id` 字段为 `string`，`format: uuid`；服务端写入字段值必须 UUIDv7（由 B1 idx 工具生成）；前端临时 id（`tmp_<uuid>`）只在前端 state 中存在，不进 API 请求体。`roundId` 是唯一的 deterministic logical-key 例外，格式为 `round-{positive-sequence}-{type}`，使用 string pattern/description，不得声明 `format: uuid`。
- 所有时间字段统一 `string` + `format: date-time`；不允许某些字段使用 unix epoch number。
- 所有语言字段统一 BCP 47（如 `en` / `zh-CN` / `en-SG`）；OpenAPI schema 使用 `string` + pattern / example，实际允许集由产品 i18n 与质量评估 gate 控制。

### 4.4 breaking change linter 规则集（v1.0.0 freeze 后强制）

- **禁止**：删除已发布 endpoint / 重命名 path / 修改 method / 删除 schema 字段 / 修改字段类型 / 把 optional 字段改为 required / 删除已发布枚举值。
- **允许（additive）**：新增 endpoint / 新增 tag / 新增 optional 字段 / 新增枚举值（且字段为 string-typed enum） / 新增可选 query 参数 / 新增 example。
- **P0 例外**：`POST /api/v1/privacy/exports` 从 P0 `501 ApiErrorResponse` 切到 P1 `202 PrivacyRequestWithJob`、`POST /api/v1/resumes/{resumeId}/exports` 从 P0 `501 ApiErrorResponse` 切到 P1 `202 + Job(jobType=resume_export)`，均是已预留能力变为可用；该行为必须递增 spec/history 和 fixture，但不按 breaking change 处理。
- **未上线 freeze correction**：仅在 product owner 明确授权、accepted ADR、所有 consumer 同批迁移且未发布 baseline 时允许原地修订 v1.0.0。gate 必须从 merge-base 读取旧 baseline 与新 OpenAPI 比对，全部 normalized findings 按 `severity + path + kind + before + after` 与 ADR 机器 oracle exact-match 后才允许 re-freeze；authorized breaking、ordinary additive 和 informational 不得互相替代。同时替换 current/baseline 导致的零 finding 不能作为证据。
- **审计要求**：已发布 baseline 的 breaking change 仍必须 v2.0.0；未上线 correction 必须 attach accepted ADR、base-ref finding artifact 与本 spec/history 增量。远端 CI label workflow 仅在 A5 触发条件成立后再接入，当前以本地 gate + owner review 为准。

### 4.5 codegen 与 drift 约束

- generator 输入：`openapi/openapi.yaml` + `openapi/templates/`（Go / TS 模板）；输出 `backend/internal/api/generated/` 与 frontend `client.ts` / `types.ts`，不生成无消费方的 raw-spec 字符串快照。
- generated 文件必须 idempotent；本地 `make codegen-check` / `git diff --exit-code` 阻塞漂移。远端 CI 接入由 A5 后续触发条件决定。
- 业务 handler 必须 implement generator 产出的 server interface；不允许业务包定义自己的 DTO 类型。
- TypeScript generator 必须导出 typed `ApiClientError`：public fields 精确为 `status: number | null` 与 `apiError: ApiErrorResponse | null`；可另带稳定 `kind` / `cause` 区分 HTTP、abort 与 transport，但不得暴露 fetch `Response` 或把 raw response body、provider 信息、secret 拼入错误文案。
- Generated client 必须把五类 failure 收敛到 `ApiClientError`：non-2xx valid JSON error、non-JSON body、empty body、`AbortError`、fetch/transport rejection。只有第一类携带 parsed `apiError`；后四类 `apiError=null` 且不得伪造 `error.code` / `retryable`。业务消费者只能读取 `status` 与 `apiError.error.code/retryable/details`，禁止解析 `Error.message` 判断 validation/auth/not-found/conflict/mismatch/retryable。

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

至少以下 schema 必须包含或可追溯到 `GenerationProvenance`：`TargetJob.summary` / `fitSummary`、`AssistantAction`、`FeedbackReport`、`ResumeTailorRun`、`Resume`。`Resume` 通过 `structuredProfile.provenance` 覆盖 AI 抽取的结构化简历内容；tailor 改写建议 provenance 由 `ResumeTailorRun` 承载。D-37 `ResumeSummary` 是非生成详情的最小列表投影，不携带 provenance，fixture validator 不得要求 `listResumes.items[*].structuredProfile.provenance`。缺失 full `Resume` provenance 的 fixture 仍不得通过 `make validate-fixtures`。

### 4.7 fixtures 与隐私约束

- `openapi/fixtures/<tag>/<operationId>.json` 必须 schema-valid（本地由 `make validate-fixtures` 校验；远端 CI 接入由 A5 后续触发条件决定）。
- fixtures 中绝不出现真实用户邮箱 / 真实电话 / 真实公司名敏感信息；统一用 `Acme` / `acme.example` / `alice@example.com`。
- `prototype-baseline` scenario 来自 `ui-design/src/data.jsx`；维护方式：`make sync-fixtures-from-prototype`（B2 owner）。
- `listResumes` fixture 的每个 item 必须精确匹配 D-37 九字段 closed `ResumeSummary`，不得包含 full `Resume` detail/provenance；`getResume` fixture 继续验证完整 `Resume`。fixture/example/Prism/mock 与全部消费者必须同批切换，禁止用 frontend N+1 `getResume` 补齐列表。
- TargetJob fixtures 必须只表达 paste intake：`importTargetJob` 的 `default` 与 `paste-primary` request body 精确为 `{rawText,targetLanguage,resumeId}`；validator 必须拒绝 source wrapper、URL、fileObjectId、manual-form、title/company hints 和任意 additional property。Canonical `validation-blank-raw-text` 使用纯空白 `rawText`，返回 `422 + ApiErrorResponse`，marker 精确包含 `error.code=VALIDATION_FAILED`、`error.retryable=false`、`error.details.field=rawText`；negative-case harness 必须断言 request 只因 `/rawText` 的 non-whitespace schema rule 失败并正常校验 error response，禁止跳过全部 request validation。`getTargetJob` / `listTargetJobs` / `updateTargetJob` 仍至少覆盖 parsed ready、cross-user hidden 404、invalid state transition 三类 scenario，但 response 不得含 `sourceType` / `sourceUrl`。新增 scenario 必须通过 `make validate-fixtures`，并被对应 BDD 场景引用。
- Practice fixtures 必须同时覆盖 `getPracticeSession` 的 `pending`、`retryable_failed`、`terminal_failed`、`complete` user projections，以及 `sendPracticeMessage` 的 `default` success 和 planned `validation-empty-text`、`auth-unauthorized`、`session-not-found`、`reply-pending-conflict`、`client-message-mismatch`、`ai-timeout-retryable` failures。每个 error scenario 锁 status、`ApiErrorResponse.error.code`、`retryable` 与必要 details marker；reload/retry pair 必须证明 AI failure 后同一 `clientMessageId` 可见、same-ID retry 成功且没有重复 user/assistant message。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `openapi/openapi.yaml` 与 fixtures | B2 | 唯一真理源 |
| 当前 enum 类型 / `ApiError` inner object / `PageInfo` schema | B1 | B2 通过 `$ref` 引用；B2 自身维护 `ApiErrorResponse` envelope；`PracticeMode` / `QuestionReviewStatus` 不再属于当前 source |
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
| C-8 | enum 与 B1 同源 | 在 `openapi.yaml` 引用共享 enum | codegen | 生成 TS 与 Go 类型，与 [B1 D-6](../shared-conventions-codified/spec.md#31-已锁定决策) 完全一致；不得重新生成已删除的 PracticeMode / QuestionReviewStatus | B2 001 + B1 |
| C-9 | mock 同源（前端 + 后端） | E1 拉起 mock server | 前端 msw 与后端 mock-server 都消费 `openapi/fixtures/` | 同一 endpoint 两端响应字节级一致；B2 002 先证明 fixture → OpenAPI example → Prism response 的 default scenario 字节级一致 | B2 002（partial）+ E1 |
| C-10 | B2 executable freeze handoff | 本 spec 的 contract lock 已完成，B2 001 / 002 / 003 均完成 | 当前 active spec 关系已保留 | `openapi/openapi.yaml` v1.0.0、codegen drift、fixtures 与 breaking-change linter 均通过验证；依赖 B2 的后续 implementation 可启动；roadmap 只保留 active spec 关系，不单独冒充本项已通过 | B2 003（汇总 001 / 002 证据） |
| C-14 | report pre-release correction | OPENAPI-001 accepted 且旧 baseline 尚未发布 | 修改 direct report contract | merge-base findings按 severity/path/kind/before/after 与 JSON oracle exact-match；ready non-null payload / failed-only errorCode state matrix、closed fixtures/codegen/consumers 全切换；新 baseline 与 current 一致且无未记录 finding | B2 001/002/003 |
| C-15 | oversized-context error enum | B1 001 marker 已通过 | 同步 `ApiErrorCode` | enum 精确新增 `REPORT_CONTEXT_TOO_LARGE` 且 oracle severity=additive；该值不进入 breaking allowset；failed report fixture 使用 canonical code | B1 001 + B2 001/002/003 |
| C-16 | report action wire fuse | `ReportNextAction.label` schema | schema validation/codegen | 1..200 code-point 边界 fail closed 且 generated 同步；只表示 malformed 防线，不能替代 24/64 semantic gate、18/52 targeted-repair margin 或 downstream desktop/mobile UX evidence | B2 001 + F3 002/004 |
| C-17 | report retry internal-only | product/evalkit使用generation/judge max4与内部attempt audit | schema/inventory/fixture/generated negative | `FeedbackReport`仍只有queued/generating/ready/failed；无attemptCount/retryCount/reason/scope/progress或retry-generation endpoint；frontend polling timeout不改API terminal state | B2 001 + backend-review/001 + frontend-report-dashboard/001 |
| C-18 | TargetJob paste-only freeze correction | OPENAPI-002 accepted 且旧 baseline 尚未发布 | 修改 TargetJob import/read 与 upload purpose contract | merge-base findings按 severity/path/kind/before/after 精确等于 OPENAPI-002 oracle；`importTargetJob` 只接受 closed `{rawText,targetLanguage,resumeId}`，TargetJob response 不含 source provenance，`target_job_attachment` purpose 不可用；37/10 inventory、import 202 response 和通用 `createUploadPresign` 不变；fixtures/codegen/mock/consumers 同批迁移且 URL/file/manual-form/source schema scoped zero-reference 后才可 re-freeze | B2 001/002/003 + mock-contract-suite/001 + frontend/backend TargetJob owners |
| C-19 | Practice failure reload and same-ID retry | user message 已持久化，AI reply 失败或 request outcome 不确定 | `getPracticeSession` reload 后按 typed error/status 决定 retry，再次 `sendPracticeMessage` | user projection 保留原 `clientMessageId` 与 durable `replyStatus`；assistant 不携带 recovery fields；仅 `retryable_failed` 显示 same-ID retry，terminal/auth/not-found/conflict/mismatch 不重试；success 后状态为 `complete` 且全会话只有一条对应 user 与 assistant；TS consumer 全程不解析 error message | B2 001/002/003 + mock-contract-suite/001 + backend-practice/002 + frontend-workspace-and-practice/002 |
| C-20 | TargetJob canonical-round report overview | owned TargetJob 有 2..5 canonical rounds，可能存在多个 ready 与更新的 queued/generating/failed attempt | `listTargetJobReports` | response 覆盖每个 round；currentReport 按 `generatedAt/createdAt/id` 取最新合法 ready，latestAttempt 按 `createdAt/id` 取最新任意状态；二者可指向不同 report；无 report 为双 null。cursor/pageSize/pageInfo/full report 与 `TargetJob.latestReportId` 不存在；唯一 UI consumer 是 target-scoped ReportsScreen，Parse 只负责入口 | B2 001/002/003 + backend-review/001 + frontend-report/001 |
| C-21 | Resume list summary / detail split | 用户有 upload/paste、不同 parse 状态与有/无可读正文的简历 | 调用 `listResumes` 后再打开一条 `getResume` | list method/path/operationId/200 与 pagination 不变，items 只含九个 required `ResumeSummary` 字段；`summaryHeadline` 和 `hasReadableContent` 严格按 D-37 投影，空白正文、空 object、仅有 file/source/status 都不产生假阳性；不含任何详情/provenance；打开详情时 `getResume` 仍返回完整 `Resume`。Go/TS types、fixture/mock、store projection、service/handler 与全部 frontend consumers 同批通过，无兼容层或列表 N+1 detail fetch | B2 001/002/003/004 + backend-resume + frontend-home/frontend-resume-workshop |
| C-11 | provenance 完整性 | 任一 AI 生成 response fixture 缺少 `GenerationProvenance` 或不可追溯到含 provenance 的 job/resource | `make validate-fixtures` | 失败；列出 operationId 与缺失字段；001 只锁 schema 可追溯关系，fixture 内容由 002 验证 | B2 001（schema）+ B2 002（fixtures）+ F3 |
| C-12 | resume export 501 例外 | D-26 已落地 `exportResume` operation | `make lint-openapi` + `make validate-fixtures` | `exportResume` 允许 P0 `501 + ApiErrorResponse.error.code="RESUME_EXPORT_NOT_AVAILABLE"`；除 `requestPrivacyExport` / `exportResume` 外的 endpoint 返回 501 会被 inventory lint 拒绝；未来切到 `202 + Job(jobType=resume_export)` 必须递增 spec/history 与 fixture | openapi-v1-contract/004 |
| C-13 | Practice 轮次身份与 TargetJob 进度 | TargetJob 有 1..N 个结构化轮次，practice plan/session 可含 legacy 数据、超 int32 sequence 与重复完成请求 | 创建/读取 plan，完成 session，再读取 TargetJob | 新 plan 返回规范化 `roundId + roundSequence`；OpenAPI/baseline/generated/validator 对 int32 maximum 和 logical-key 位数一致；TargetJob 返回去重且有序的 `completedRounds` 与第一个未完成 `currentRound`，全部完成时 `currentRound=null/status=completed`；legacy null/overflow plan 不得被当作当前轮复用；字段变更通过 additive diff、fixture 与 codegen gate | B2 001/002/003 + backend-practice 001/002 + backend-targetjob 001 + frontend-workspace-and-practice 001 |

## 7 关联计划

B2 当前由本 spec 保留 active contract lock；真实 executable contract 由 B2 自身 plans 与当前 product-scope owner 承接（[engineering-roadmap §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec)）：

- `001-bootstrap`：落地 `openapi/openapi.yaml` 框架、ADR-Q1 Auth 路径、`DELETE /api/v1/me` privacy deletion endpoint、privacy export 501 例外、B1 enum `$ref`、`GenerationProvenance`、`make codegen-openapi` 与本地 drift check；当前 freeze 为 10 tag / 37 endpoint。
- `002-fixtures-and-mock-source`：每个 operationId 一份 fixtures + `prototype-baseline` 同步工具；E1 接入。
- `003-breaking-change-gate`：linter 规则集 + ADR 模板；远端 CI label workflow 仅在 A5 触发条件成立后再评估。
- `004-resume-additive-coverage`：承接当前扁平 Resume contract、D-37 list summary/detail full split、`RESUME_EXPORT_NOT_AVAILABLE`、Resumes / ResumeTailor fixtures、codegen drift、inventory lint 与 consumer handoff；BDD 不适用。
- `backend-auth/001-email-code-session-bootstrap` + `frontend-shell/001-app-shell-auth-settings`：D-25 Auth single-entry profile completion pre-launch correction 落地：`AuthEmailStartRequest` 不含 `purpose` / `displayName`，`UserContext` 新增 `profileCompletionRequired`，Auth tag 包含 `PATCH /api/v1/me completeMyProfile` + `CompleteProfileRequest` + fixture + Go/TS generated artifacts；真实 handler 由 backend-auth 承接。
- `product-scope/001-core-loop-module-pruning`：当前核心闭环范围 owner，负责 10 tag / 37 endpoint freeze、baseline、fixtures、generated artifacts、runtime route-negative tests 与 active spec zero-reference gates。
- `practice-voice-mvp/001-cascaded-stt-llm-tts`：D-22 Practice voice turn additive 升级落地：`PracticeSessions` tag 包含 `createPracticeVoiceTurn` operation + voice turn request/response schema + fixture + Go/TS generated artifacts；真实 handler 由 practice voice Phase 5 补齐。

在放行依赖 B2 的后续业务实现前，必须确认 B2 001/002/003/004 与 product-scope owner gate 已补齐 `openapi/openapi.yaml`、fixtures、baseline、generated artifacts 与 diff whitelist，不得只停留在本 spec 文本。

后续如出现 v1.1.0 / v2.0.0 升级：递增 spec 版本 + history；每次升级在 §3.1.1 中保留 endpoint 完整快照。
