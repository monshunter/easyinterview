# 001 — Plan and Session Orchestration BDD Plan

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-06-13

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 0 BDD 框架与编号

本 plan 全部 BDD 场景按 [`test/scenarios/README.md`](../../../../../test/scenarios/README.md) 与 [`test/scenarios/e2e/README.md`](../../../../../test/scenarios/e2e/README.md) 约定执行：

- 套件: `e2e`
- 阶段: `P0`
- 编号起点: `E2E.P0.022`（[`test/scenarios/e2e/INDEX.md`](../../../../../test/scenarios/e2e/INDEX.md) 当前最高 `E2E.P0.021`）
- 编号范围: `E2E.P0.022 ~ E2E.P0.026`（5 个场景）
- 目录格式: `test/scenarios/e2e/p0-NNN-<slug>/`
- 场景资产: `README.md` + `scripts/{setup,trigger,verify,cleanup}.sh` + `data/{seed-input.md,expected-outcome.md}`

每个场景的 setup / trigger / verify / cleanup / 数据 / 证据捕获在 [bdd-checklist](./bdd-checklist.md) 跟踪；本文件只记录场景的 Given / When / Then 与覆盖范围，不出现执行 checkbox。

## 1 场景矩阵

| 场景 ID | 名称 | 类别 | 关联 Plan Phase | 关联 spec AC |
|---------|------|------|----------------|---------------|
| `E2E.P0.022` | createPracticePlan baseline + getPracticePlan + cross-user 隔离 | primary | Phase 1 | C-1, C-13（plan 维度子集） |
| `E2E.P0.023` | startPracticeSession 同步首题 + getPracticeSession + outbox started | primary | Phase 1 | C-4 |
| `E2E.P0.024` | AI 失败 → reservation failed_retryable → 同 key 重试成功 | failure / recovery | Phase 2 | C-5, C-21（startSession 子集）, C-23 |
| `E2E.P0.025` | 副作用 endpoint Idempotency-Key 行为矩阵（replay / mismatch / 跨用户 / 并发 / 同 plan 多 key 并发 / 越权 404） | boundary + alternate + regression | Phase 2 | C-10, C-13, C-22, C-23, C-24, C-25 |
| `E2E.P0.026` | 隐私红线 + AI metric 完整 + legacy-negative grep | privacy + observability + regression | Phase 3 | C-16（001 子集）, D-11 |

## 2 Phase 1 — 主流程场景

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.022` | createPracticePlan baseline + getPracticePlan + cross-user 隔离 | 已登录用户 A 拥有 `target_job_id` 与 `resume_asset_id`；用户 B 已登录但无关 | 用户 A `POST /practice/plans` 携带 `Idempotency-Key` + `goal=baseline` → 用户 A `GET /practice/plans/{planId}` → 用户 B `GET /practice/plans/{planId}` | 用户 A 收到 `201 + PracticePlan{status:'ready', source*: null}`；DB `practice_plans` 写入；audit_events 写入元数据摘要；用户 A `GET` 拿到完整 plan；用户 B `GET` 返回 `404 + PRACTICE_PLAN_NOT_FOUND`，不泄露存在性 | `test/scenarios/e2e/p0-022-practice-plan-baseline-create-and-read/` |
| `E2E.P0.023` | startPracticeSession 同步首题 + getPracticeSession + outbox started | 用户 A 拥有 `practice_plans(id=planId, status='ready', goal='baseline')`；F3 baseline active；A3 真实或 fake AIClient 配置可返回 first-question | 用户 A `POST /practice/sessions` 携带 `Idempotency-Key` + `planId` → 用户 A `GET /practice/sessions/{sessionId}` | `201 + PracticeSession{status:'running', currentTurn:{turnIndex:1, status:'asked', questionText, questionIntent, askedAt}}`；DB `practice_sessions(status='running')` + `practice_turns(turn_index=1, status='asked')` + `practice_session_events(seq_no=1, event_type='session_started')` 写入；`outbox_events` 表中 `practice.session.started` 行存在且 payload 与 B3 schema 一致；`GET` 拿到完整 session；外部 AI 调用不在 DB tx 内（通过 lock 检测 / 时序断言） | `test/scenarios/e2e/p0-023-practice-session-start-and-first-question/` |

## 3 Phase 2 — 错误与边界场景

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.024` | AI 失败 → reservation failed_retryable → 同 key 重试成功 | 用户 A 拥有 `practice_plans(status='ready', goal='baseline')`；F3 active；A3 fake AIClient 配置首次调用注入 `AI_PROVIDER_TIMEOUT`、第二次调用返回成功 first-question | 用户 A `POST /practice/sessions` 携带 `Idempotency-Key=K1` → 收到 502 后用同 `K1` 再次 `POST /practice/sessions` | 第一次：`502 + ApiError{code:'AI_PROVIDER_TIMEOUT'}`；envelope 不含 prompt / response 明文；DB `idempotency_records(status='failed_retryable')` + `practice_sessions(status='failed', failure_code='AI_PROVIDER_TIMEOUT')`；不发 `practice.session.started` outbox。第二次：`201 + PracticeSession{status:'running', currentTurn:{turnIndex:1, status:'asked'}}`；DB 升级到 `idempotency_records(status='succeeded')` + `practice_sessions(status='running')`；`practice.session.started` outbox 行出现一次 | `test/scenarios/e2e/p0-024-practice-session-ai-failure-retry/` |
| `E2E.P0.025` | 001 副作用 endpoint Idempotency-Key 行为矩阵 | 用户 A 已成功创建 `planId` 与 `sessionId`；用户 B 独立账号；存在 ready 状态的第二个 plan `planId2` | 多组并发 / 顺序请求：① 同 user + 同 key + 同 fingerprint replay；② 同 user + 同 key + 不同 fingerprint mismatch；③ 用户 B 同 key 不同请求；④ 同 user + 不同 key 同 plan 多个 startPracticeSession 并发；⑤ 跨用户 GET planId / sessionId | ① replay 返回首次 response，DB / outbox / audit 不重复；② 返回 `409 PRACTICE_SESSION_CONFLICT`，envelope 不泄露首次资源；③ 用户 B 走自己的 idempotency record，与用户 A 隔离；④ 仅一个 startPracticeSession 成功，其它返回 `409 PRACTICE_SESSION_CONFLICT`，DB partial UNIQUE INDEX 保证同 user+plan 仅一个 active session；⑤ 用户 B GET 返回 `404 + PRACTICE_PLAN_NOT_FOUND` / `PRACTICE_SESSION_NOT_FOUND` | `test/scenarios/e2e/p0-025-practice-idempotency-and-isolation-matrix/` |

## 4 Phase 3 — 观测 / 隐私 / 回归场景

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.026` | 隐私红线 + AI metric label 边界 + legacy-negative grep | E2E.P0.022 + E2E.P0.023 + E2E.P0.024 已运行；log / metric / audit / outbox 收集器可读 | 触发额外 AI 调用，采集本 phase 的 log / metric label / audit_events / outbox payload；运行 repo-wide grep | 全集合中 `question_text` / `answer_text` / `hint_text` / AI prompt body / response body / provider secret 零出现；`ai_task_runs` 行包含 D-15 全部 typed columns；A3 `ai_task_*` metric label 命中 F1 allowlist，且不含 `feature_key` / prompt-rubric version；repo grep 中 `warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route / `practiceModeCard` 在 backend-practice 输出文档与代码 diff 零出现 | `test/scenarios/e2e/p0-026-practice-observability-and-privacy-redlines/` |

## 5 数据隔离与污染恢复

每个场景按 `test/scenarios/e2e/README.md` §5 / §3 / §6 / §8 约定：

- 数据隔离：每个场景使用独立的 `user_id` / `plan_id` / `idempotency_key` 命名空间；不复用其它场景的 fixture
- 清理顺序：cleanup 先删自身 `practice_session_events` → `practice_turns` → `practice_sessions` → `practice_plans` → `idempotency_records` → audit_events → outbox_events → users
- 污染恢复：场景失败时按 README §8 顺序：① 清理场景自身资源；② 定位并恢复 shared 组件（如 dispatcher 状态、F3 cache）；③ 仅在 ① ② 失败时 `test/scenarios/env-cleanup.sh && env-setup.sh` 全量重建
- 不预设 Helm chart / 外部 Git 平台名称；所有命令以本仓库脚本为真理源

## 6 与单元测试边界

本 BDD plan 验证用户可见行为切片（HTTP API + DB 状态 + outbox emit + log/metric/audit 红线）；不重复内部接口签名、序列化结构、错误映射等单元测试覆盖（见 [test-plan](./test-plan.md)）。
