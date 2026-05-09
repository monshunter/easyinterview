# 001 — Plan and Session Orchestration BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-09

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.022 createPracticePlan baseline + getPracticePlan + cross-user 隔离

- [ ] 创建场景目录 `test/scenarios/e2e/p0-022-practice-plan-baseline-create-and-read/`
- [ ] 编写 `README.md`：Given / When / Then、关联需求、依赖组件
- [ ] 准备 `data/seed-input.md`：用户 A / B fixtures，target_job_id, resume_asset_id, idempotency_key
- [ ] 准备 `data/expected-outcome.md`：201 PracticePlan 响应字段、DB 行、audit row、cross-user 404 envelope
- [ ] 实现 `scripts/setup.sh`：登录用户 A / B + 准备 target_job + resume_asset + 清理同名 idempotency_key
- [ ] 实现 `scripts/trigger.sh`：用户 A POST /practice/plans → GET /practice/plans/{id}；用户 B GET /practice/plans/{id}
- [ ] 实现 `scripts/verify.sh`：断言 201 + DB 写入 + audit 摘要无 question/answer 文本 + cross-user 404 + grep 隐私红线（PracticeMode 上下文 `legacy debrief replay value` 零出现）
- [ ] 实现 `scripts/cleanup.sh`：删除场景自身 plan / idempotency / audit / users（按 §5 清理顺序）
- [ ] 在 `test/scenarios/e2e/INDEX.md` 追加 P0.022 行
- [ ] 执行 `bash scripts/{setup,trigger,verify,cleanup}.sh` 通过
- [ ] 记录验证证据（HTTP response + DB snapshot + audit redaction）到 `.test-output/`

## E2E.P0.023 startPracticeSession 同步首题 + getPracticeSession + outbox started

- [ ] 创建场景目录 `test/scenarios/e2e/p0-023-practice-session-start-and-first-question/`
- [ ] 编写 `README.md`
- [ ] 准备 `data/seed-input.md`：用户 + ready plan（baseline）+ F3 / A3 配置
- [ ] 准备 `data/expected-outcome.md`：201 PracticeSession 含 currentTurn(turnIndex=1, status='asked')、DB sessions/turns/events 行、outbox `practice.session.started` 行 payload
- [ ] 实现 `scripts/setup.sh`：登录用户 + 创建 baseline plan（若 P0.022 fixture 可复用则引用，但保持隔离）
- [ ] 实现 `scripts/trigger.sh`：POST /practice/sessions → GET /practice/sessions/{id}
- [ ] 实现 `scripts/verify.sh`：断言 201 + currentTurn 同步返回 + DB 状态 + outbox row 存在且 payload 与 B3 schema 一致 + AI 调用不在 DB tx 内（lock 检测 / 时序断言）+ grep 隐私红线
- [ ] 实现 `scripts/cleanup.sh`：删除 sessions / turns / events / outbox / plan / idempotency / users
- [ ] 执行通过
- [ ] 记录验证证据

## E2E.P0.024 AI 失败 → reservation failed_retryable → 同 key 重试成功

- [ ] 创建场景目录 `test/scenarios/e2e/p0-024-practice-session-ai-failure-retry/`
- [ ] 编写 `README.md`
- [ ] 准备 `data/seed-input.md`：用户 + ready plan + fake AIClient 配置（首次注入 timeout，二次成功）
- [ ] 准备 `data/expected-outcome.md`：首次 502 envelope（不含 prompt/response 明文）+ DB failed reservation；二次 201 + currentTurn + DB succeeded reservation + outbox 行出现一次
- [ ] 实现 `scripts/setup.sh`：登录 + 创建 baseline plan + 注入 fake AIClient 失败模式
- [ ] 实现 `scripts/trigger.sh`：第一次 POST → 第二次同 key POST
- [ ] 实现 `scripts/verify.sh`：断言两次响应、idempotency_records 状态机、practice_sessions 状态、outbox 行计数 = 1、envelope 无明文
- [ ] 实现 `scripts/cleanup.sh`：完整清理 + 重置 fake AIClient
- [ ] 执行通过
- [ ] 记录验证证据

## E2E.P0.025 副作用 endpoint Idempotency-Key 行为矩阵

- [ ] 创建场景目录 `test/scenarios/e2e/p0-025-practice-idempotency-and-isolation-matrix/`
- [ ] 编写 `README.md`：5 个子场景矩阵
- [ ] 准备 `data/seed-input.md`：用户 A / B + planId + planId2 + 多组 idempotency key + fingerprint
- [ ] 准备 `data/expected-outcome.md`：5 个子场景的预期响应、DB 状态、conflict envelope
- [ ] 实现 `scripts/setup.sh`：登录 A / B + 创建 plans + 准备并发 fixture
- [ ] 实现 `scripts/trigger.sh`：① replay；② mismatch；③ 跨用户；④ 同 user + 不同 key + 同 plan 并发（goroutine 并发或 parallel curl）；⑤ 跨用户 GET
- [ ] 实现 `scripts/verify.sh`：每子场景断言；对并发场景断言只产生一份业务副作用；对 mismatch 断言 envelope 不泄露首次资源；对跨用户 GET 断言 404 + 错误码
- [ ] 实现 `scripts/cleanup.sh`：完整清理两个用户的所有 plan / session / idempotency / audit / users
- [ ] 执行通过
- [ ] 记录验证证据

## E2E.P0.026 隐私红线 + AI metric 完整 + legacy-negative grep

- [ ] 创建场景目录 `test/scenarios/e2e/p0-026-practice-observability-and-privacy-redlines/`
- [ ] 编写 `README.md`
- [ ] 准备 `data/seed-input.md`：依赖 P0.022 / P0.023 / P0.024 已运行的 fixture（或本场景 setup 内重放）
- [ ] 准备 `data/expected-outcome.md`：log/metric/audit/outbox 中 zero-trace、ai_task_runs typed columns 完整、legacy-negative grep zero-hit
- [ ] 实现 `scripts/setup.sh`：触发若干 createPlan + startSession + 失败重试 → 收集 log / metric / audit / outbox 快照
- [ ] 实现 `scripts/trigger.sh`：触发 grep / metric scrape / DB query
- [ ] 实现 `scripts/verify.sh`：① 全集合 grep 红线（question_text / answer_text / hint_text / prompt body / response body / provider secret） 必须 0 命中；② `ai_task_runs` 行包含全部 D-15 typed columns；③ A3 `ai_task_*` metric label 命中 F1 allowlist，且不包含 `feature_key` / prompt / rubric version；④ repo-wide grep retired 术语（warmup / single_drill / drill_builder / mistake_queue / growth_center / 独立 voice route / practiceModeCard）必须 0 命中
- [ ] 实现 `scripts/cleanup.sh`：清理本场景与依赖场景的 fixture
- [ ] 执行通过
- [ ] 记录验证证据
