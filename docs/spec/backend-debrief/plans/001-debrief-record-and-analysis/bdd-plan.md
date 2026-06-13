# 001 Debrief Record and Analysis BDD Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-06-13

**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 0 目标

为 backend-debrief/001-debrief-record-and-analysis 定义端到端 BDD 场景集。每个场景含 setup / trigger / verify / cleanup 四段，覆盖用户行为流而非内部 unit；前端 frontend-debrief/001 独立分配 scenarios E2E.P0.065-069。

执行入口：每个场景目录必须提供并按顺序运行 `scripts/setup.sh` → `scripts/trigger.sh` → `scripts/verify.sh` → `scripts/cleanup.sh`；`trigger.sh` 必须保留真实 backend runner exit code，Go 测试型 trigger 的 verify.sh 必须断言 `--- PASS` + `ok`，并包含旧口径 grep 反查。

## 1 Scenario Matrix

| 场景 ID | category | 关联 spec AC | 关联 plan phase | 关联 checklist BDD-Gate |
|---------|----------|--------------|-----------------|------------------------|
| E2E.P0.060 | Primary | C-1, C-2, C-3, C-5 | Phase 1-4 | 6.6 |
| E2E.P0.061 | Primary + Privacy | C-6, C-7, C-8 | Phase 5 | 6.7 |
| E2E.P0.062 | Failure/recovery | C-11, C-12 | Phase 4 | 6.8 |
| E2E.P0.063 | Primary + Failure/recovery | C-9, C-10 | Phase 3 | 6.9 |
| E2E.P0.064 | Privacy + Regression/Legacy-negative | C-14, C-15 | Phase 6 | 6.10 |

## 2 场景详情

### E2E.P0.060 — Debrief Create + Worker Generation Happy

| 字段 | 内容 |
|------|------|
| 目录 | `test/scenarios/e2e/p0-060-debrief-create-and-generate/` |
| Phase | Phase 1-4 |
| 关联 spec AC | C-1, C-2, C-3, C-5 |
| 执行入口 | `bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh; bash scripts/cleanup.sh`（在该场景目录内执行） |
| Given | 用户已认证（passwordless session cookie）；`target_jobs(user_id=A, id=T)` 已存在 ready；F3 `debrief.generate` baseline active；A3 mock 返回有效 JSON |
| When | (1) `POST /debriefs` with `Idempotency-Key=IK1` + body `{targetJobId:T, roundType:'behavioral', interviewerRole:'hiring_manager', language:'zh', questions:[{questionText,...}, ...3 items], notes:'...'}` (2) backend drainer lease async_jobs(debrief_generate) (3) debrief.GenerateHandler 处理 |
| Then | (a) HTTP 202 + DebriefWithJob{debriefId=D, job:{jobType:'debrief_generate', status:'queued'}}；(b) DB debriefs(id=D, user_id=A, status='draft', raw_questions=[3 items])；(c) async_jobs(debrief_generate, status='queued', dedupe_key=D)；(d) outbox debrief.created{debriefId:D, targetJobId:T, roundType:'behavioral', questionCount:3}；(e) drainer 处理后 debriefs.status='completed', raw_questions[*].aiAnalysis 注入, risk_items 非空, prompt_version/rubric_version/model_id/provider 4 列填充；(f) async_jobs.status='succeeded'；(g) outbox debrief.completed{debriefId:D, targetJobId:T, riskItemCount:N, practiceFocusCount:N}；(h) ai_task_runs 写一行 task_type='debrief_generate', status='success' |
| Cleanup | 删除 user A + cascade（debriefs / async_jobs / outbox_events / ai_task_runs / audit_events 全部 cascade）；scenario 结束后数据库归零 |
| Privacy 反查 | verify.sh 含 `! grep -E "questionText\|myAnswerSummary\|notes" outbox_events.json metric.log audit_events.json`（不在 evidence file 中出现 raw text） |
| Legacy 反查 | verify.sh 含 `! grep -E "mistakes_count\|generatedMistakeCount" outbox_events.json shared/events.yaml backend/internal/debrief/` |

### E2E.P0.061 — Debrief Get Draft/Completed + Cross-User Isolation

| 字段 | 内容 |
|------|------|
| 目录 | `test/scenarios/e2e/p0-061-debrief-get-and-cross-user/` |
| Phase | Phase 5 |
| 关联 spec AC | C-6, C-7, C-8 |
| 执行入口 | `bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh; bash scripts/cleanup.sh`（在该场景目录内执行） |
| Given | 用户 A 已认证；debriefs(id=D, user_id=A, status='draft') 已存在（未触发 worker）；debriefs(id=E, user_id=A, status='completed', raw_questions[*].aiAnalysis 非空, risk_items 非空) 已存在；用户 B 已认证（独立 session） |
| When | (1) 用户 A `GET /debriefs/D`；(2) 用户 A `GET /debriefs/E`；(3) 用户 B `GET /debriefs/D`；(4) 用户 A `GET /debriefs/NOT_EXIST` |
| Then | (1) 返回 200 + Debrief{status:'draft', questions:[3 items with aiAnalysis:null], riskItems:[], nextRoundChecklist:[], thankYouDraft:null, provenance:null}；(2) 返回 200 + Debrief 全字段 with questions[*].aiAnalysis + riskItems + provenance{6 fields}；(3) 返回 404 + B1 `DEBRIEF_NOT_FOUND`（不泄露存在性）；(4) 返回 404 + `DEBRIEF_NOT_FOUND` |
| Cleanup | 删除用户 A + B + cascade |
| Privacy 反查 | verify.sh assert response body 不含 prompt body / response body / provider secret；provenance 仅 6 wire 字段 |

### E2E.P0.062 — Worker AI Failure Graceful + Retry + Permanent Fail

| 字段 | 内容 |
|------|------|
| 目录 | `test/scenarios/e2e/p0-062-debrief-worker-failure-and-retry/` |
| Phase | Phase 4 |
| 关联 spec AC | C-11, C-12 |
| 执行入口 | `bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh; bash scripts/cleanup.sh`（在该场景目录内执行） |
| Given | 用户 A 已认证；A3 mock 配置为：前 4 次调用 returnTimeout，第 5 次也 timeout；F3 active；用户 A 已通过 createDebrief 创建 debriefs(id=D, status='draft') |
| When | drainer 反复 lease + handle 共 5 次 |
| Then | (a) 前 4 次：async_jobs.attempts++, available_at=now()+backoff, status='queued', locked_at=null；debriefs.status 保持 'draft'；ai_task_runs 写 failed row × 4 with status='timeout' + B1 `AI_PROVIDER_TIMEOUT`；outbox debrief.completed 不发出；(b) 第 5 次：async_jobs.status='failed'（permanent）+ locked_at=null；debriefs.status='draft'；ai_task_runs 写 failed row 第 5 行；outbox 仍未发出；(c) 前端通过 getJob 感知 status='failed' + errorCode |
| Cleanup | 删除用户 A + cascade |
| Privacy 反查 | verify.sh assert ai_task_runs error_code 字段是 B1 enum literal（不是 raw provider message） |

### E2E.P0.063 — suggestDebriefQuestions Sync + AI Failure

| 字段 | 内容 |
|------|------|
| 目录 | `test/scenarios/e2e/p0-063-suggest-debrief-questions/` |
| Phase | Phase 3 |
| 关联 spec AC | C-9, C-10 |
| 执行入口 | `bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh; bash scripts/cleanup.sh`（在该场景目录内执行） |
| Given | 用户 A 已认证；target_jobs(user_id=A, id=T) ready；可选 practice_sessions(user_id=A, id=S, target_job_id=T) completed；可选 resume_versions(user_id=A, id=R) ready；F3 `debrief.suggest_questions` baseline active；A3 mock 配置：第一次返回有效 JSON {suggestions:[6 items]}，第二次返回 timeout，第三次返回非 JSON |
| When | (1) `POST /debriefs/question-suggestions` with `{targetJobId:T, sessionId:S, resumeVersionId:R, language:'zh', count:6}`；(2) 再次 `POST` 同 body（A3 第二次 timeout）；(3) 再次 `POST` 同 body（A3 第三次 invalid JSON） |
| Then | (1) HTTP 200 + SuggestDebriefQuestionsResponse{suggestions:[6 items each {questionText, whyLikelyAsked, source: enum value}]}；ai_task_runs 写 success row；audit 一行；(2) HTTP 502 + B1 `AI_PROVIDER_TIMEOUT`；ai_task_runs 写 timeout row；audit 一行 with error_code；(3) HTTP 502 + B1 `AI_OUTPUT_INVALID`；ai_task_runs 写 invalid row |
| Cleanup | 删除用户 A + cascade；不应有 debriefs 行（suggestDebriefQuestions 不写 debriefs） |
| Privacy 反查 | verify.sh assert response suggestions 不含 raw target_job description（只允许 AI 派生 questions）；ai_task_runs 不在 metric label 泄漏 |

### E2E.P0.064 — Debrief Privacy + Legacy Negative

| 字段 | 内容 |
|------|------|
| 目录 | `test/scenarios/e2e/p0-064-debrief-privacy-and-legacy-negative/` |
| Phase | Phase 6 |
| 关联 spec AC | C-14, C-15 |
| 执行入口 | `bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh; bash scripts/cleanup.sh`（在该场景目录内执行） |
| Given | 用户 A 完成 E2E.P0.060 happy flow 所产生的所有 DB 行（debriefs / async_jobs / ai_task_runs / audit_events / outbox_events）；scenario 注入特殊 marker string `__SECRET_RAW_TEXT__` 进入 createDebrief request 的 notes / questionText 字段 |
| When | (1) verify.sh 扫描 outbox_events.payload jsonb 列、audit_events.metadata、metric.log（如可观测）；(2) verify.sh 扫描 backend/internal/debrief 源码、shared/events.yaml、shared/jobs.yaml、openapi/fixtures/Debriefs/ 不出现 retired 标识符 |
| Then | (a) marker string `__SECRET_RAW_TEXT__` 在 outbox / audit / metric 全部 0 命中；(b) retired 标识符 `mistakes_count` / `generatedMistakeCount` / `experience_library` / `drill_builder` / `growth_center` / `star_editor` / `debrief_voice` 在源码 / 契约 / scenario runtime 全部 0 命中；(c) ai_task_runs 行字段完整（feature_key, model_profile_name, status, input_tokens, output_tokens, latency_ms, validation_status, error_code） |
| Cleanup | 删除用户 A + cascade |
| Privacy 反查 | 整个 scenario 就是 privacy + legacy 反查；verify.sh 失败时记录命中位置 + 字符串 |

## 3 编号占用

本 plan 占用 E2E.P0.060 ~ E2E.P0.064（5 个）。下一可用编号 E2E.P0.065 由 frontend-debrief/001-debrief-screen-and-handoff 使用（5 个 P0.065-069）。

## 4 编号策略与与 backend-review/frontend-report-dashboard 的对齐

- backend-review/001 占用 P0.052-055
- frontend-report-dashboard/001 占用 P0.056-059
- backend-debrief/001（本 plan）占用 P0.060-064
- frontend-debrief/001 占用 P0.065-069

完整闭环 P0.001-069（含 backend / frontend 全部 P0 域）持续作为 P0 主漏斗的整体覆盖。
