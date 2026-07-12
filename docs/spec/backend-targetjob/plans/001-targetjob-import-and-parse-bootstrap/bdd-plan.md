# Backend TargetJob BDD Plan

> **版本**: 1.11
> **状态**: completed
> **更新日期**: 2026-07-12

## 0 场景矩阵

| 场景 ID | 类别 | 关联 Plan Phase | 关联主 checklist BDD-Gate | 关联 spec C-* |
|---------|------|----------------|---------------------------|----------------|
| E2E.P0.010 | primary | Phase 2 + Phase 3.1 + Phase 4 + Phase 10 | 6.1 / 10.4 | C-1 / C-3 / C-6 / C-7 / C-9 / C-12 / C-16 |
| E2E.P0.011 | alternate (URL source) | Phase 3.3 + Phase 4 + Phase 5 | 6.2 | C-2 / C-3 / C-9 |
| E2E.P0.012 | failure / recovery | Phase 4.4 + Phase 5 + Phase 11 | 6.3 / 11.4 | C-4 / C-5 / C-9 / C-10 |
| E2E.P0.013 | primary / manual fallback | Phase 2.1 + Phase 3.1 | 6.4 | C-3 / C-6 / C-9 / C-11 / C-13 |
| E2E.P0.098 | backend progress / persistence / recovery | Phase 17 | 17.5 | C-17 |
| E2E.P0.018 | primary / workspace delete archive | Phase 12 | 12.4 | C-7a / C-8 |

> 备注：编号承接 `practice-voice-mvp/spec.md §4.3` 已预留的 `E2E.P0.007` / `E2E.P0.008` / `E2E.P0.009`；本计划接续使用 `E2E.P0.010` / `E2E.P0.011` / `E2E.P0.012` / `E2E.P0.013`。
>
> L2 remediation 备注（2026-05-08）：`test/scenarios/e2e/p0-010..013-*` 使用 `cmd/api` HTTP 场景 harness，覆盖 auth middleware / HTTP API / TargetJob handler-service / cmd/api in-process runner kernel / ParseExecutor。`verify.sh` 输出 `method=cmd-api-http` 与 `validBddEvidence=true`；focused tests 作为 TDD 辅助证据。
>
> BUG-0146 备注（2026-07-09）：`E2E.P0.010` 的 primary path 追加 C-16 回归补证。除 deterministic HTTP harness 外，本次 closeout 需要一条真实 provider + host-run frontend/browser smoke，证明有效 JD 缺少公司名时仍进入 `analysisStatus='ready'`，Company 展示语言相关兜底值，且 `ai_task_runs` 存在 `jd_parse` 证据。
>
> Phase 17 证据边界（2026-07-12）：`E2E.P0.098` 在本 owner 只认真实 PostgreSQL plan/session/event 与 TargetJob Get/List read model 证据，并可组合 frontend unit/static negative gate；它不自动等价于真实浏览器刷新证明。live frontend/backend browser refresh + quick-start 由 `frontend-workspace-and-practice/001` 保持未完成，直到存在实际执行记录。

## Phase 6: TargetJob backend behavior

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.010 | Text JD import 走完异步解析并可列表 / 详情 / 更新 | 已登录用户、A3 / F3 active、stub `target.import.default` 在 `APP_ENV=test` 可用、cookie jar / `Idempotency-Key` 准备完毕；BUG-0146 回归补证使用真实 provider 与一个未披露公司名的有效 JD | 用户使用 `manual_text` 调 `POST /targets/import`，runner kernel 完成后依次调 `GET /targets`、`GET /targets/{id}`、`PATCH /targets/{id}`；BUG-0146 回归补证打开 authenticated `/parse?...targetJobId=...` | 202 响应携带 generated `TargetJobWithJob` + `Job(type=target_import,status=queued)`；runner kernel 处理后列表可见该 job；详情返回 `analysis_status='ready'` + 至少 1 条 `must_have` requirement + `summary.provenance` 完整；`PATCH` 可更新合法 status / notes 且不修改 `analysis_status`；outbox 含 `target.import.requested` + `target.parsed` 事件；同 key 重复 import 返回相同 `targetJobId`，DB / outbox 不出现重复 row；BUG-0146 回归补证中 `companyName` 用语言相关兜底值，`/parse` 不显示 `JD 解析失败`，`ai_task_runs` 有 `jd_parse` provider/model/status/validation 证据 | `test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/` + BUG-0146 local browser smoke |
| E2E.P0.011 | URL JD import 守护与抓取 | 已登录用户、本地 fixture HTTPS server 暴露合规 JD HTML、A3 / F3 active；同时准备非法目标（私网 IP、metadata 服务、超长 body、HTTP scheme） | 用户用合法 URL `POST /targets/import` 后等待 runner kernel；再依次提交非法目标 | 合法 URL：runner kernel 抓取后写 `target_job_sources.url` 为规范化 URL、`snapshot_text` 为去密正文、`fetched_at` / `freshness_status='fresh'`，`target.parsed` 发出，internal-only `source_refresh` follow-up job 写入且 payload 不含完整 URL。非法目标全部返回 B1 `TARGET_IMPORT_SOURCE_INVALID` 或 `TARGET_IMPORT_SOURCE_UNAVAILABLE`，事件 / 日志 / metric label / audit 不含完整 URL、query 串或 prompt 明文 | `test/scenarios/e2e/p0-011-targetjob-url-import-fetch-and-parse/` |
| E2E.P0.012 | Parse 失败 retryable / non-retryable | 已登录用户、F3 / A3 可被场景注入返回特定错误（test stub 切到 `AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_CONFIG_INVALID`）；准备已 import 的 manual_text TargetJob | runner kernel 处理 job；分别触发 retryable 与 non-retryable 失败 | retryable 失败：`target.analysis.failed.retryable=true` 且失败 TargetJob 资产被删除；non-retryable 失败：事件 `retryable=false`，`GET /targets/{id}` 返回 404 + `TARGET_JOB_NOT_FOUND`，`GET /targets` 不含该 job；error envelope / log / metric 不含 prompt / response 明文；F3 unsupported / disabled profile 与 A3 缺 secret 也走同一失败语义；用户可重新 import 创建新 `targetJobId` 而不被旧失败阻塞 | `test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/` |
| E2E.P0.013 | Manual form import ready 直达 | 已登录用户、用户在表单中直接填写 title / company / rawDescription、准备 `Idempotency-Key`，A3 / F3 可不可用均不影响该路径 | 用户使用 `manual_form` 调 `POST /targets/import`，随后调 `GET /targets/{id}` 与 `GET /targets`，并用同一 key 重复 import | 202 响应仍为 `TargetJobWithJob`，其中 `job.jobType='target_import'` 且 `job.status='succeeded'`；TargetJob 立即 `analysis_status='ready'`，至少 1 条 `must_have` draft requirement；不创建待 runner kernel 处理的 `target_import` async job，不发 `target.import.requested` / `target.parsed`；同 key 重复 import 返回同一 `targetJobId`；日志 / audit / metrics 不含原文 JD | `test/scenarios/e2e/p0-013-targetjob-manual-form-ready/` |
| E2E.P0.018 | Workspace 删除图标持久归档 | 已登录用户已有 ready TargetJob，workspace 使用 real backend / generated client，准备 `Idempotency-Key` | 用户点击 workspace 卡片删除图标；随后刷新 workspace 并直接调用 `GET /targets` / `GET /targets/{id}` | `archiveTargetJob` 返回 archived `TargetJob`，DB 写 `status='archived'` 与 `deleted_at`；workspace 成功后移除卡片且刷新后不回灌；`GET /targets` 不含该 job，`GET /targets/{id}` 返回 404；重复归档返回 `TARGET_INVALID_STATE_TRANSITION` conflict；越权归档返回 `TARGET_JOB_NOT_FOUND` | `test/scenarios/e2e/p0-018-workspace-default-render/` + local real-backend browser smoke |
