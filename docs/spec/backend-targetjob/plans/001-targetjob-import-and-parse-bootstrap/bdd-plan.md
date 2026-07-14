# Backend TargetJob BDD Plan

> **版本**: 1.13
> **状态**: completed
> **更新日期**: 2026-07-14

## 0 场景矩阵

| 场景 ID | 类别 | 关联 Plan Phase | 关联主 checklist BDD-Gate | 关联 spec C-* |
|---------|------|----------------|---------------------------|----------------|
| E2E.P0.010 | paste-only primary + size boundary | Phase 18 + 20 | 18.4 / 20.3 | C-1 / C-2 / C-3 / C-6 / C-7 / C-9 / C-12 / C-16 / C-19 |
| E2E.P0.012 | failure / recovery | Phase 18 | 18.4 | C-4 / C-5 / C-9 / C-10 |
| E2E.P0.098 | backend progress / persistence / recovery | Phase 17 | 17.5 | C-17 |
| E2E.P0.018 | primary / workspace delete archive | Phase 12 | 12.4 | C-7a / C-8 |

> Phase 18 当前 TargetJob 导入场景只保留 `E2E.P0.010` 与 `E2E.P0.012`；旧 URL 与同步手工表单场景从当前 suite 删除。
>
> L2 remediation 备注（2026-05-08）：`test/scenarios/e2e/p0-010..013-*` 使用 `cmd/api` HTTP 场景 harness，覆盖 auth middleware / HTTP API / TargetJob handler-service / cmd/api in-process runner kernel / ParseExecutor。`verify.sh` 输出 `method=cmd-api-http` 与 `validBddEvidence=true`；focused tests 作为 TDD 辅助证据。
>
> BUG-0146 备注（2026-07-09）：`E2E.P0.010` 的 primary path 追加 C-16 回归补证。除 deterministic HTTP harness 外，本次 closeout 需要一条真实 provider + host-run frontend/browser smoke，证明有效 JD 缺少公司名时仍进入 `analysisStatus='ready'`，Company 展示语言相关兜底值，且 `ai_task_runs` 存在 `jd_parse` 证据。
>
> Phase 17 证据边界（2026-07-12）：`E2E.P0.098` 在本 owner 只认真实 PostgreSQL plan/session/event 与 TargetJob Get/List read model 证据，并可组合 frontend unit/static negative gate；它不自动等价于真实浏览器刷新证明。live frontend/backend browser refresh + quick-start 由 `frontend-workspace-and-practice/001` 保持未完成，直到存在实际执行记录。

## Phase 6: TargetJob backend behavior

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.010 | 粘贴 JD 走完异步解析并验证 96KiB 边界 | 已登录用户、A3/F3 active、RuntimeConfig/default 98,304 bytes 与合法 `{rawText,targetLanguage,resumeId}` | 提交 UTF-8 limit、limit+1，runner 完成 limit case 后 list/get/patch，并 replay 同 key | limit 返回 queued 并 ready；limit+1 typed validation 且零 TargetJob/job/outbox/provider；replay 不重复；raw_jd_text 是唯一原文事实源 | `test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/` + P0.015 handoff |
| E2E.P0.012 | Parse 失败 retryable / non-retryable | 已登录用户使用同一 paste-only wire，F3/A3 可注入 timeout/output-invalid/secret/config 错误 | runner 分别处理 retryable 与 non-retryable 失败 | `target.analysis.failed.retryable` 与 AI 错误语义一致，失败 TargetJob 不可见；用户可重新粘贴创建新 TargetJob；无 source row/file ref/refresh job，日志与事件不泄漏原文 | `test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/` |
| E2E.P0.018 | Workspace 删除图标持久归档 | 已登录用户已有 ready TargetJob，workspace 使用 real backend / generated client，准备 `Idempotency-Key` | 用户点击 workspace 卡片删除图标；随后刷新 workspace 并直接调用 `GET /targets` / `GET /targets/{id}` | `archiveTargetJob` 返回 archived `TargetJob`，DB 写 `status='archived'` 与 `deleted_at`；workspace 成功后移除卡片且刷新后不回灌；`GET /targets` 不含该 job，`GET /targets/{id}` 返回 404；重复归档返回 `TARGET_INVALID_STATE_TRANSITION` conflict；越权归档返回 `TARGET_JOB_NOT_FOUND` | `test/scenarios/e2e/p0-018-workspace-default-render/` + local real-backend browser smoke |
