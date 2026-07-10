# Backend TargetJob Spec

> **版本**: 2.7
> **状态**: active
> **更新日期**: 2026-07-10

## 1 背景与目标

`backend-targetjob` 承接 [engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 中 JD import / parse 后端域，落地 P0 用户路径：用户在首页粘贴 / 上传 / URL 导入 / 手工录入 JD 后，后端创建 `TargetJob`，异步完成 JD 解析，并把岗位需求 / 摘要 / fit 信号写入 [B4 `target_jobs` / `target_job_requirements` / `target_job_sources`](../db-migrations-baseline/spec.md) 三张 baseline 表，让前端 `parse` 屏与后续 `Mock Interview Plan` 能从 mock fixture 切到真实数据。

[B2 OpenAPI v1](../openapi-v1-contract/spec.md) 已定义本域承接的 5 个 operation：`importTargetJob` / `listTargetJobs` / `getTargetJob` / `updateTargetJob` / `archiveTargetJob`，[B3 `event-and-outbox-contract`](../event-and-outbox-contract/spec.md) 已冻结 `target.import.requested` / `target.parsed` / `target.analysis.failed` 三个事件与 `target_import` 异步 job（`asynqTask: target.import`，apiFacing），[F3 `prompt-rubric-registry`](../prompt-rubric-registry/spec.md) 已锁定 feature_key `target.import.parse` 与默认 model profile `target.import.default`。本 subject 把这些已就位的契约、表结构、事件和 feature_key 缝合成一个 P0 后端域，明确 owner 边界、隐私红线、解析失败语义、idempotency、TargetJob 归档与 source-fetch 安全约束。

本 subject 不重新设计 OpenAPI / DB / event / feature_key。任何契约变更必须先回到 owner spec 修订，再落到本 subject 的 plan。

## 2 范围

### 2.1 In Scope

- 5 个 TargetJob operation 的 backend handler + service + store：
  - `POST /targets/import` `importTargetJob`：202 + `TargetJobWithJob`，要求 `Idempotency-Key`；`url` / `manual_text` / `file` 派发 `target_import` 异步 job，`manual_form` 同步落库 `analysisStatus=ready` 并返回 terminal `Job(jobType=target_import,status=succeeded)`。
  - `GET /targets` `listTargetJobs`：按 `status` / `analysisStatus` / `q` / `cursor` / `pageSize` 列表，cursor-paginated。
  - `GET /targets/{targetJobId}` `getTargetJob`：返回完整 `TargetJob` 工作台对象。
  - `PATCH /targets/{targetJobId}` `updateTargetJob`：更新 lifecycle status / location / notes / hint，要求 `Idempotency-Key`。
  - `POST /targets/{targetJobId}/archive` `archiveTargetJob`：持久软归档当前用户的 TargetJob，要求 `Idempotency-Key`；写入 `status='archived'` 与 `deleted_at`，随后列表 / 详情不可见。
- 4 类导入源（`url` / `manual_text` / `file` / `manual_form`）的写入路径与归一化：source snapshot 写入 `target_job_sources`，`raw_jd_text` 只在 `target_jobs` 表内部保留，事件 / 日志 / metric / audit 不携带原文。
- URL 导入的 fetch 边界：scheme / DNS-IP SSRF 防护、length cap、timeout、`fetched_at` 与 `freshness_status` 写入；不抓登录后内容、不绕过 robots / ToS、不抓非 HTML 资源。
- File 导入的 `file_objects` 引用：仅接受 `purpose='target_job_attachment'` 的 file object；缺失或越权返回 B1 error envelope，不泄露文件内容。
- 异步 JD 解析管线：消费 `target_import` job → 使用 [F3 `RegistryClient.Resolve("target.import.parse", language)`](../prompt-rubric-registry/spec.md) 解析三元组 → 调用 [A3 `AIClient`](../ai-provider-and-model-routing/spec.md) → 写入 `target_job_requirements` + `target_jobs.summary` + `target_jobs.fit_summary` + `provenance`，事务内更新 `analysis_status`，并与 `target.parsed` / `target.analysis.failed` outbox 事件及 internal-only `source_refresh` follow-up job 保持原子提交。
- 异步执行边界：`target_import` / `source_refresh` 已由 [backend-async-runner](../backend-async-runner/spec.md) 的单一 backend-internal kernel 接管；本 subject 保留 TargetJob handler / service / store / executor 业务实现，不引入独立 worker 进程，也不自建 Asynq 集群。
- Cross-user / cross-tenant 隔离：所有 read / write 必须按 `user_id` 过滤；越权访问返回 404，不泄露目标存在性。
- Idempotency：`importTargetJob` / `updateTargetJob` / `archiveTargetJob` 必须按 `(user_id, idempotency_key)` 去重；同 key 重复请求返回同一 `targetJobId` / 同一 `target_import` job 或同一 archived TargetJob，不创建多余记录或多余 job。
- 隐私 / 观测红线：`raw_jd_text`、`source_url` 完整 query 串、文件对象 URL、AI prompt body / response body、provider secret 不进入 log / metric label / audit / 事件 payload；只允许 hash、长度、status、profile、provider、cost micros、error code 摘要。
- F1 metric 注册边界：所有新增的 target-job metric / audit 类型必须先在 [F1 `observability-stack`](../observability-stack/spec.md) baseline 字典登记或由 F1 owner 承接，不得在本域私造 metric / label。
- 解析失败准入语义：AI 调用失败、超时、provider 缺 secret、URL fetch 失败、超长输入、空白文本都必须映射到 B1 共享错误码 `AI_*` / `TARGET_*`，并通过 `target.analysis.failed` 事件保留失败诊断；失败解析不得留下可见 TargetJob / JD 资产，`target_jobs` 与级联 source / requirement 行必须在失败提交事务中删除，`GET /targets` 与 `GET /targets/{id}` 不得返回失败解析产物。retryable 标志必须与 A3 / B1 错误码语义一致。实施前 Phase 0 必须先把本 spec 需要的 TargetJob 错误码同步到 B1/B2 可执行契约。

### 2.2 Out of Scope

- 不实现岗位推荐（Job Picks / JD Match）：该模块按 product-scope D-17 不在当前范围；本 subject 不新增 recommendation endpoint、JD search endpoint 或 data-source plan。
- 不实现 `MockInterviewPlan` / `practice_plans` 创建、修改、列表；归 `backend-practice` owner。
- 不实现简历或 `resume_versions` 的解析、改写、绑定到 plan；归 `backend-resume` owner。
- 不实现独立 `company_intel` 抓取、聚合、刷新或详情页数据源；本 spec 只允许 `target.parsed` 事件触发 internal-only `source_refresh` job，由 handler 将 source 标记为 `stale`，供 workspace 内嵌公司轻情报摘要消费当前 TargetJob 公开摘要；真实外部抓取刷新不在本 subject 内，需先回 owner spec 重新设计。
- 不实现独立 worker / Asynq dispatcher / 生产级 outbox consumer；P0 用 backend-internal drainer 完成本地与 BDD 验证。
- 不修改 B2 OpenAPI、B3 events.yaml / jobs.yaml、B4 baseline 表结构、A3 provider 协议、F3 `target.import.parse` baseline prompt / rubric 文本；任何修改先回到 owner spec。
- 不实现报告生成、证据回收或错题回顾；这些归 `backend-review` 等 owner。真实面试复盘按 product-scope D-22 不在当前范围，不再作为 downstream owner。
- 不实现完整 privacy export；`target_jobs` 软删 + 删除矩阵 dry-run schema 由 [B4](../db-migrations-baseline/spec.md) 承接，`privacy_delete` 运行链路已由 [`backend-async-runner/001`](../backend-async-runner/spec.md) kernel 接管（`privacy_export` 仍为 reserved，不由本 plan 注册）；本 spec 只保证 `deleted_at` 软删字段与 cascade 关系不被违反。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | API 契约来源 | 本域只消费 [B2 OpenAPI](../openapi-v1-contract/spec.md) 已定义的 `importTargetJob` / `listTargetJobs` / `getTargetJob` / `updateTargetJob` / `archiveTargetJob`；不私造 endpoint、不重写 schema | 任何字段 / 新 operation 先在 B2 spec / `openapi.yaml` 修订 |
| D-2 | DB 真理源 | 复用 [B4 baseline](../db-migrations-baseline/spec.md) 的 `target_jobs` / `target_job_requirements` / `target_job_sources` 与现有索引、CHECK 约束、软删字段 | 不在本 plan 添加 migration；如需新列必须先修订 B4 |
| D-3 | 事件契约 | 复用 [B3](../event-and-outbox-contract/spec.md) 已冻结的 `target.import.requested` / `target.parsed` / `target.analysis.failed` 与 `target_import` job | 事件 payload 与 PII 边界不得扩张；新增字段先回到 B3 spec |
| D-4 | AI 调用形态 | 业务侧调用 [F3 `RegistryClient.Resolve("target.import.parse", language)`](../prompt-rubric-registry/spec.md) → 拿三元组 → 调用 [A3 `AIClient`](../ai-provider-and-model-routing/spec.md) `Complete`；payload 必须携带 `feature_key + prompt_version + rubric_version + model_profile_name + language + data_source_version` | 业务包不得 hardcode prompt 文本，不得直接持有 provider / model 字符串 |
| D-5 | Async runner 边界（已收干） | `target_import` 与下游 `source_refresh` 的运行已由 [`backend-async-runner/001`](../backend-async-runner/spec.md) kernel 接管：`targetjob.ParseExecutor` / `SourceRefreshHandler` 经 `runner.FromTargetjobHandler` adapter 注册到单一 `runner.Runtime`；不启动独立 worker 进程，也不引入 Asynq | kernel 复用同一 B3 payload red-line；本 spec 保留 handler 业务实现，`targetjob.Drainer` 抽象仅保留给 focused 测试 |
| D-6 | Idempotency | `importTargetJob` / `updateTargetJob` / `archiveTargetJob` 按 `(user_id, idempotency_key)` 去重；重复请求返回同一 `targetJobId`、同一 active `target_import` job 或同一 archived TargetJob；解析失败重试由用户显式 `PATCH` 或后续 retry plan 决策 | 防止重复创建 / 重复派发 / 重复写入事件 / 重复归档 |
| D-14 | TargetJob archive semantics | `archiveTargetJob` 复用简历 archive 先例：软归档而非隐私删除；成功写 `status='archived'` 与 `deleted_at`，read-side 继续通过 `deleted_at is null` 过滤；同用户重复归档返回 `TARGET_INVALID_STATE_TRANSITION` conflict，越权/不存在返回 `TARGET_JOB_NOT_FOUND` | workspace 删除图标刷新后不再回灌已归档卡片，同时保留 privacy delete 独立 owner |
| D-7 | URL fetch 安全 | 仅允许 `https` scheme；阻止私网 / 链路本地 / 元数据服务 IP；总 body cap 1 MiB；fetch 超时 10s；不跟随 cross-origin redirect 进入私网；user agent 显式标注 EasyInterview crawler 版本；保存 snapshot 时去除 query secret | 防止 SSRF、爬虫滥用与日志泄露 |
| D-8 | 隐私红线 | 事件 / metric label / log / audit / async payload 不得包含 `raw_jd_text`、`source_url` 完整路径、文件 object URL、AI prompt / response body、provider secret；只允许 hash、长度、language、status、profile、provider、model_id、cost micros、error code | 与 product-scope §9.3 / F1 一致 |
| D-9 | Cross-user 隔离 | 所有 read / write SQL 必须按 `user_id` 过滤；越权访问 `getTargetJob` / `updateTargetJob` 返回 HTTP 404 + B1 `TARGET_JOB_NOT_FOUND` 而不是 `FORBIDDEN`，避免泄露存在性 | 与 [backend-auth `DELETE /me` 同 key 用户隔离](../backend-auth/spec.md) 一致 |
| D-10 | 解析失败可重试 | A3 返回 `AI_PROVIDER_TIMEOUT` / `AI_FALLBACK_EXHAUSTED` 等 retryable 错误时 `target.analysis.failed.retryable=true`；`AI_OUTPUT_INVALID` / `AI_UNSUPPORTED_CAPABILITY` / `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_CONFIG_INVALID` / `TARGET_IMPORT_SOURCE_INVALID` 为 `retryable=false`；URL 上游暂时不可达映射 `TARGET_IMPORT_SOURCE_UNAVAILABLE` 且 `retryable=true` | 让前端、运维与未来 retry plan 有一致重试语义；禁止使用未登记的 out-of-scope alias |
| D-11 | manual_form 同步路径 | 用户手工录入的 `manual_form` 直接写入 requirements 草稿且 `analysis_status='ready'`，跳过异步解析；HTTP 仍返回 `202 + TargetJobWithJob`，其中 `job.status='succeeded'` 表示 terminal compatibility job，不派发 `target.import.requested` / `target.parsed` | 让“手工兜底”不被 AI 故障阻塞，同时保持 B2 wire shape 与 generated client 不分叉 |
| D-12 | 重新解析 owner 边界 | `ui-design/src/screens-p0-complete.jsx::ParseScreen` 已提供前端 `Re-parse / 重新解析` 体验；后端 P0 不新增 rerun endpoint 或 rerun job，前端重新解析需要落地真实数据时通过现有 `importTargetJob` 创建新的 import / TargetJob，或在后续 frontend plan 内消费既有 generated client | 避免后端 plan 把前端交互按钮误列为后端待确认事项 |
| D-13 | import source event 映射 | `target.import.requested.sourceType` 复用 B3 event-local `url/text/file`：B2 `manual_text` 映射为 `text`，B2 `manual_form` 不发该事件；exact API source variant 只存在于 B2 request / DB `source_type` | 保持事件 payload 表达“需要异步导入/解析的请求”，避免同步兜底场景污染 runner 语义 |

### 3.2 非后端 owner 决策

| ID | 事项 | Owner | 本域处理 |
|----|------------|------|----------|
| Q-1 | TargetJob 重新解析（rerun parse） | frontend / ui-design | 已由 `ui-design` 提供 `Re-parse / 重新解析` 前端体验；后端只保留既有 `importTargetJob` 契约，不新增 rerun operation |
| Q-2 | URL fetch 统一出网代理 | nginx / 接入层 / 部署平台 | 不进入 easyinterview 后端项目配置，不要求 A4 提供 app-level proxy key |

### 3.3 待确认事项

| ID | 待确认事项 | 影响 | 默认处理 |
|----|------------|------|----------|
| Q-3 | 解析失败的用户级 backoff | 同一 `targetJobId` 短时间内多次失败是否限制后续 import | 默认不在 P0 实施，由观测数据推动后续 plan |

## 4 设计约束

### 4.1 API 契约约束

- 必须使用 [B2 generated `ServerInterface`](../openapi-v1-contract/spec.md) 注册 handler，不得绕过 generated types 自造 router。
- 入参反序列化必须使用 generated request types；响应必须使用 generated response types；fixture 与真实 handler 共用同一 schema，不得 cherry-pick 字段。
- `importTargetJob` 必须在事务内创建 `target_jobs` 行 + `target_job_sources` 行（除 `manual_form` 外）+ 派发 `target_import` job，并在响应里返回 generated `Job` 对象；`manual_form` 返回 terminal `target_import/succeeded` compatibility job。
- `getTargetJob` / `listTargetJobs` 必须返回 `requirements` 完整数组与 `summary` / `fitSummary` 的 `provenance` 字段；`provenance.featureFlag` 缺省 `none`，`rubricVersion` 非评分场景写 `not_applicable`，与 [GenerationProvenance](../openapi-v1-contract/spec.md) 描述一致。
- `updateTargetJob` 必须验证 `status` 状态机合法迁移（如 `draft → preparing → applied → interviewing → offer / rejected → archived`），非法迁移返回 B1 错误。
- `archiveTargetJob` 必须使用 generated `ServerInterface`、generated `TargetJob` response 与 B1 error envelope；缺 `Idempotency-Key` 返回 validation error，已归档返回 `TARGET_INVALID_STATE_TRANSITION` conflict。

### 4.2 数据约束

- 所有写入必须在事务内完成；`target.import.requested` / `target.parsed` / `target.analysis.failed` 必须通过 outbox（B3 工作集）发出，与业务写入同事务，避免双写。
- `target_job_requirements.kind` 仅允许 B4 已有 CHECK 列表（`must_have` / `nice_to_have` / `hidden_signal` / `interview_focus`）；B2 OpenAPI / fixtures 必须在 Phase 0 additive 扩展到同一列表后，backend 才能在 API response 中返回 `hidden_signal` / `interview_focus`。
- `target_job_sources.freshness_status` 默认 `fresh`，URL 抓取后写 `fetched_at`、sanitized `url` 与 `snapshot_text`；后续 internal-only `source_refresh` job 由 `target.parsed` 触发，handler 将 source 标记为 `stale`，不执行真实外部抓取刷新。
- TargetJob store / handler / runner SQL 只能引用当前 B4 `target_jobs` 表真实列；已移除 profile 模块的 `profile_id` 不得出现在 active TargetJob 读写路径、sqlmock 列集合或 integration gate 中。
- 软删：`archiveTargetJob` 必须原子写入 `target_jobs.status='archived'` 与 `deleted_at`；`deleted_at IS NOT NULL` 的 `target_jobs` 不参与列表 / 详情 / 解析；查询必须显式过滤。
- `target.import.parse` 输出的 `title` 是必需岗位身份；`companyName` 是可选展示字段。JD 未披露公司名时，parse success path 必须写入语言相关兜底展示值（`zh-*` 为 `未提供`，其他语言为 `Unknown company`），不得把有效 JD 标记为 `AI_OUTPUT_INVALID`。

### 4.3 安全 / 隐私约束

- URL 抓取的 user agent 必须包含 `EasyInterview JD-Crawler/<version>` 与联系页链接；不得伪装其它来源。
- DNS 解析必须解析后再 IP 校验，阻止解析到 RFC1918 / 169.254 / `::1` 等私网范围；不得跟随 redirect 至私网。
- 抓取 body cap 1 MiB，超长直接拒绝并标 `TARGET_IMPORT_SOURCE_INVALID`；不得保留超长片段。
- AI 调用必须 fail-closed：F3 `Resolve` 返回 unsupported / disabled profile 或 A3 缺 provider secret 时，整个 import 必须返回 B1 错误并写 `target.analysis.failed`；不得静默回退到 stub provider（除非 `APP_ENV=test`）。
- log / metric label / audit / 事件 payload 不得出现 `raw_jd_text`、`source_url.path`、`source_url.query`、文件 object URL、AI prompt / response body、provider secret；仅允许 hash / 长度 / status / profile / provider / model_id / cost micros / error code 摘要。

### 4.4 异步 / 可观测约束

- 必须复用 [backend-auth](../backend-auth/spec.md) 同款 backend-internal goroutine drainer 模式，drainer 必须有 graceful shutdown / drain timeout 测试。
- AI 调用与 outbox 写入必须可观测：`target_job_imports_total` / `target_job_parse_duration_seconds` / `target_job_parse_failures_total` 等 metric 名实施前先在 [F1 baseline](../observability-stack/spec.md) 字典登记；label 仅使用 F1 allowed labels，`error_code` 来自 B1，`source_type` 来自 B2/B3 有界枚举。
- `cmd/api` TargetJob parse AI runtime 必须复用 A3 observability decorator 并写入 `ai_task_runs`，记录 task type、resource id、provider/model、status、validation status 与错误摘要；真实 provider 失败不得缺少 task-run 证据。
- 解析延迟 P95 要求作为观测目标登记，但不作为本 plan 验收 gate；评估在后续质量 / SLO plan 决策。

### 4.5 文档治理约束

- 本 spec 后续修订必须原地更新；不允许创建同主题 sibling spec。
- 涉及 OpenAPI / events / migrations / shared enums / runtime config 的修改必须先回 owner spec / `*.yaml` truth source。
- 涉及用户行为流的 plan 必须维护 BDD gate；本域所有 backend operation 都属于用户可见 API 行为，必须有 BDD 场景。
- 命中 `completed` plan 时不创建同主题 sibling follow-up plan，原地修订即可。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| API contract | [B2 `openapi-v1-contract`](../openapi-v1-contract/spec.md) | 5 个 TargetJob operation 的 schema、fixtures、generated client / server |
| Backend domain | `backend-targetjob` | handler / service / store / drainer / parse executor / outbox emit |
| DB schema | [B4 `db-migrations-baseline`](../db-migrations-baseline/spec.md) | `target_jobs` / `target_job_requirements` / `target_job_sources` 列与索引；删除矩阵 dry-run |
| Event / job contract | [B3 `event-and-outbox-contract`](../event-and-outbox-contract/spec.md) | `target.import.requested` / `target.parsed` / `target.analysis.failed` 与 `target_import` job |
| AI provider | [A3 `ai-provider-and-model-routing`](../ai-provider-and-model-routing/spec.md) | `AIClient.Complete`、provider registry、model profile、observability decorator |
| Prompt / rubric | [F3 `prompt-rubric-registry`](../prompt-rubric-registry/spec.md) | `target.import.parse` feature_key、Resolve 实现、baseline prompt / rubric 文件 |
| Config / secret | [A4 `secrets-and-config`](../secrets-and-config/spec.md) | provider secret、feature flag；URL fetch proxy 不属于 app-level 配置 |
| Observability | [F1 `observability-stack`](../observability-stack/spec.md) | metric / audit 类型登记、label allowlist、隐私红线 |
| Frontend consumer | `frontend-home-job-picks-and-parse` | Home / Parse 通过 generated client 消费 TargetJob 导入与读取 |
| Scenario coverage | scenarios owner + 本 subject | `E2E.P0.010` / `E2E.P0.011` / `E2E.P0.012` / `E2E.P0.013` setup / trigger / verify / cleanup |
| Backend internal runner | [backend-async-runner](../backend-async-runner/spec.md) | `target_import` / `source_refresh` handler 注册到单一 runner kernel，并沿用 B3 payload red-line |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | TargetJob 文本导入 | 已登录用户提交 `manual_text` JD 与 `targetLanguage` | `POST /targets/import` | 返回 202 + `TargetJobWithJob`，DB 写入 `target_jobs` + `target_job_sources`，派发 `target_import` job，发出 `target.import.requested` | 001 |
| C-2 | URL 导入与抓取 | 用户提交合规 `https` JD URL | `POST /targets/import` 后 drainer 处理 | URL 抓取受 SSRF / 长度 / timeout 守护，snapshot 写入 `target_job_sources.fetched_at` / `freshness_status`，敏感 query 串不入库 / 不入日志 | 001 |
| C-3 | 异步解析成功 | `target_import` job 入队，F3 / A3 active | drainer 处理 job | 业务侧 Resolve `target.import.parse` → A3 `Complete` → 写入 requirements + summary + fit + provenance，事务内 `analysis_status='ready'`，发出 `target.parsed` | 001 |
| C-4 | 异步解析失败 retryable | 解析过程 A3 返回 `AI_PROVIDER_TIMEOUT` | drainer 处理 job | 失败事务写入 `target.analysis.failed.retryable=true` 并删除失败 TargetJob 资产；`GET /targets` 不返回该 job，`GET /targets/{id}` 返回 404 + `TARGET_JOB_NOT_FOUND`；error envelope / log 不含 prompt / response 明文 | 001 |
| C-5 | 解析失败 non-retryable | 解析返回 `AI_OUTPUT_INVALID` 或 `TARGET_IMPORT_SOURCE_INVALID` | drainer 处理 job | 失败事务写入 `target.analysis.failed.retryable=false` 并删除失败 TargetJob 资产；失败 JD 原文、空标题、source snapshot 不作为可继续规划持久化，用户重试必须重新 import 创建新 `targetJobId` | 001 |
| C-6 | 列表与游标 | 用户已有多个 TargetJob，含不同 `status` / `analysisStatus` | `GET /targets` 携带过滤 + cursor | 仅返回当前用户的活跃记录，按 `updated_at DESC` 排序，分页字段符合 `PaginatedTargetJob` | 001 |
| C-7 | 详情与状态更新 | 用户拥有某 `targetJobId` | `GET /targets/{id}` + `PATCH /targets/{id}` | 详情包含 requirements / summary / fitSummary / provenance；patch 验证状态机迁移合法并要求 `Idempotency-Key` | 001 |
| C-7a | TargetJob 持久归档 | 用户在 workspace 列表点击删除图标 | `POST /targets/{targetJobId}/archive` | 返回 archived `TargetJob`，DB 写 `status='archived'` 与 `deleted_at`；随后 `GET /targets` 不再返回该 job，`GET /targets/{id}` 返回 404；重复归档返回 `TARGET_INVALID_STATE_TRANSITION` conflict；若已有 `target_import` job 后续读取到该 TargetJob 已归档 / 不可见，worker 必须以 non-retryable `TARGET_JOB_NOT_FOUND` 终结，不写二次 `target.analysis.failed` 清理事件 | 001 |
| C-8 | Cross-user 隔离 | 用户 A 持有 jobX，用户 B 携带相同 `Idempotency-Key` 与不同 source 调 import / get / patch / archive | 用户 B 调用 5 个 operation | 用户 B 不能读到 / 改到 / 归档 jobX；Idempotency dedupe 仅在同 user 范围生效；越权返回 HTTP 404 + B1 `TARGET_JOB_NOT_FOUND` | 001 |
| C-9 | 隐私红线 | 任意 import / parse / failure 完成 | 检查 log / metric label / audit / 事件 payload | 不含 `raw_jd_text`、`source_url` 完整路径、文件 object URL、AI prompt / response body、provider secret；只含 hash / 长度 / status / profile / provider / model_id / cost / error code | 001 |
| C-10 | F3 / A3 fail-closed | 选中的 `target.import.default` profile 不可解析或 provider 缺 secret | drainer 处理 job 或 import 启动 | 整个 import 返回 B1 错误并写 `target.analysis.failed`；缺 secret 映射 `AI_PROVIDER_SECRET_MISSING`，配置无效映射 `AI_PROVIDER_CONFIG_INVALID`，不静默回退 stub（除 `APP_ENV=test`） | 001 |
| C-11 | manual_form 同步路径 | 用户提交 `manual_form` 含 `title` / `companyName` / `rawDescription` | `POST /targets/import` | 同步写入 `target_jobs.analysis_status='ready'` + 草稿 requirements；返回 `202 + TargetJobWithJob` 且 `job.status='succeeded'`；不派发 runner job、不发出 `target.import.requested` / `target.parsed` | 001 |
| C-12 | Idempotency dedupe | 用户用同一 `Idempotency-Key` 重复 `importTargetJob` | 同一秒内重复发起 | 返回同一 `targetJobId` 与同一 active `target_import` job，DB 不出现重复 row，outbox 不重复发事件 | 001 |
| C-13 | 契约前置修订 | B1/B2/B3/F1 当前 contract 缺少本域所需错误码、fixture scenario、sourceType 映射说明或 metric 名 | 执行 001 Phase 0 | `shared/conventions.yaml` / OpenAPI fixtures / B3 sourceType 说明 / F1 metric dictionary 与本 spec 场景一致，`make codegen-check` / `make validate-fixtures` / `make lint-events` / metric registry tests 可执行 | 001 |
| C-14 | 文档与修订记录治理 | 本 spec 状态变更或字段调整 | 更新 spec / history / `plans/INDEX.md` / `docs/spec/INDEX.md` | 文档保持单一 owner，无 sibling spec；out-of-scope `feature_key` / `voice` route / `mistake.*` 等口径不出现在 active 文档 | docs-only |
| C-15 | 当前 B4 schema 对齐 | D-20/D-17 后 profile 模块已移除，`target_jobs.profile_id` 不存在 | `GET /targets/{id}` / `GET /targets` / `PATCH /targets/{id}` / `target_import` drainer 读取 TargetJob | active SQL 不引用 `profile_id`；解析成功/ready TargetJob 详情不因已移除列引用漂移返回 500；解析失败 TargetJob 已被删除，详情返回 404 而非脏失败态资产 | 001 |
| C-16 | 有效 JD 未披露公司名 | 用户提交有效 JD，AI 输出包含 title / requirements 但 `companyName` 为空 | `target_import` drainer 调用真实 provider 并完成 parse | TargetJob 进入 `analysisStatus='ready'`，`companyName` 写入语言相关兜底值，requirements 可见，`ai_task_runs` 记录 jd_parse provider/model/status/validation 摘要；markdown fenced JSON 可解析，带 prose 的输出仍失败 | 001 |

## 7 关联计划

- [001-targetjob-import-and-parse-bootstrap](./plans/001-targetjob-import-and-parse-bootstrap/plan.md)（active）：先修 B1/B2/B3/F1 owner 契约，再落地 5 个 TargetJob operation、4 类导入源处理、`target_import` 异步解析管线、持久归档、隐私 / 观测红线、`E2E.P0.010` / `E2E.P0.011` / `E2E.P0.012` / `E2E.P0.013` / `E2E.P0.018` BDD 场景；2026-07-08 已追加 `profile_id` 已移除列漂移修复 gate；2026-07-09 已追加有效 JD 缺公司名兜底、parse AI task-run observability gate、解析失败不得持久化 TargetJob 资产的准入修复 gate，以及 workspace 删除图标调用 `archiveTargetJob` 的持久归档 gate。

## 8 相关文档

- [Product Scope §6.7 M2 目标岗位工作台](../product-scope/spec.md#66-m2目标岗位工作台)
- [docs/ui-design/jd-resume-management.md](../../ui-design/jd-resume-management.md)
- [docs/ui-design/auth-and-entry.md](../../ui-design/auth-and-entry.md)
- [openapi-v1-contract](../openapi-v1-contract/spec.md)
- [event-and-outbox-contract](../event-and-outbox-contract/spec.md)
- [db-migrations-baseline](../db-migrations-baseline/spec.md)
- [ai-provider-and-model-routing](../ai-provider-and-model-routing/spec.md)
- [prompt-rubric-registry](../prompt-rubric-registry/spec.md)
- [secrets-and-config](../secrets-and-config/spec.md)
- [observability-stack](../observability-stack/spec.md)
- [backend-auth](../backend-auth/spec.md)（同款 in-process drainer 与隐私红线先例）
- [docs/development.md §2 Frontend / Backend Contract Workflow](../../development.md)
