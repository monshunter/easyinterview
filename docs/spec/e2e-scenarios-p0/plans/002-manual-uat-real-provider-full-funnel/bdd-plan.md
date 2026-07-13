# BDD Plan

> **版本**: 2.10
> **状态**: active
> **更新日期**: 2026-07-13

## Phase 5: Real Provider Hybrid Full Funnel

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.100 | 真实 provider hybrid 全漏斗 UAT | 本地 dev-stack 依赖 healthy（含 Mailpit）；schema 已 migrate；`deploy/dev-stack/.env` 是唯一真实本地 env 来源，包含 backend `APP_ENV=dev`、真实 auth secrets、`EMAIL_PROVIDER=mailpit`、真实 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 与 frontend `VITE_EI_API_MODE=real` / `VITE_EI_API_BASE_URL`；验收者使用 synthetic JD / resume / answer materials | AI Agent 先运行 `setup.sh` / `trigger.sh` / `verify.sh` / `cleanup.sh`；若 `deploy/dev-stack/.env` 缺真实凭证或缺浏览器证据则输出 `MANUAL_REQUIRED`；人工或浏览器 Agent 在同一输出目录补齐 Mailpit email-code 登录、Home -> JD import -> Parse & Confirm -> Practice -> Generating -> Report -> retry/next、checklist / 截图 / 脱敏日志；Workspace 只作回访/返回枢纽，且 `evidence.md` 必须包含与 `setup.env` 相同的 `RUN_ID` / `run_id` | (a) `E2E.P0.100` 登记在 e2e INDEX 且具备标准脚本契约；(b) 前端请求真实 backend base URL，无 fixture `Prefer` header；(c) backend 真实 handler / runner / PostgreSQL 产生资源；(d) AI 调用经真实 provider，证据仅记录 provider/profile/model/latency/task-run count，不记录 prompt/response 明文；(e) 账号、材料、环境、清理说明齐备，且账号入口走 Mailpit 6 位 code，不新增正式 `backend/cmd` 或直接 session bootstrap；(f) URL/storage/console 不泄露 JD/答案/报告明文；(g) 不使用 `APP_ENV=test`、P0.099 test server、deterministic stub AI 或场景专属 `.env` 作为完成证据；(h) `trigger.sh` 与 `verify.sh` 的 `scan_evidence_redline` 均通过，旧 `evidence.md` 或泄露 provider key / auth secret / session cookie / raw email code / prompt body / response body 的证据只能得到 `MANUAL_REQUIRED`；(i) env consumer gate 覆盖 `env-setup.sh --with-migrations`、`env-redeploy.sh frontend` 与 `E2E.P0.100` trigger 对 `deploy/dev-stack/.env` 的消费 | `test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/` |

## Phase 8: Grounded Report Content Reliability

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.100 | 不同上下文的报告真实可靠与bounded retry | corrected F3 markers；final paired-example prompt；product retry evidence current | run deterministic action-local/reset/separation matrix，then run fixed five representative semantic cases and retain the stricter 11-attempt diagnostic separately | product invocation initial+3/10s-20s-40s/destroy-reset；async attempts independent；mechanical final outputs 100%；fixed cases至少4/5；strict P0.100 only PASS on11/11+blind review；run59381 product-accepted but strict-FAIL | P0.100 Phase 8-9 evidence |

### Phase 8-9 retry variants

| Variant | 类型 | Given | When | Then |
|---------|------|-------|------|------|
| generation provider transient | failure/recovery | attempt1 retryable provider/protocol failure | action-local waiter10s | attempt2 succeeds；aggregate audit连续 |
| generation invalid multi-round | failure/recovery | outputs alternate sole-label and mixed violations | attempts1-4 validate | scope follows current violations；success may occur2/3/4 |
| generation exhausted | boundary | attempt4 remains invalid or retryable-failed | current action exits | terminal failed for that action，no action-local call5 |
| generation nonretryable | failure | config/secret/unsupported/context-too-large/cancel | failure classified | immediate terminal，zero retry |
| independent generation reset | lifecycle | first invocation exhausts four calls and returns | second invocation starts | retry state destroyed；second action begins at attempt1 |
| async attempt separation | boundary | async job attempts/max_attempts vary | product invocation starts | product attempt remains initial；runner values do not schedule action waits |
| lease takeover | concurrency | attempt1 reaped then attempt2 claimed；attempt1 returns late | domain persistence/finalize | stale attempt1 writes zero report/outbox/audit/job；attempt2 preserved |
| judge invalid recovery | failure/recovery | provider retryable or protocol/schema invalid | judge retries | may succeed by attempt4；audit aggregated |
| judge valid negative | content rejection | structurally valid unsupported/causal/critical verdict | evaluator returns | typed terminal FAIL，zero retry |
| frontend in-flight pause | UI recovery | hidden/blur during unresolved poll | visible/focus | resume n+1，no reset1/duplicate；single run<=49 |

## Scenario Matrix

| 场景 ID | 类型 | 覆盖计划阶段 | 主 checklist Gate |
|---------|------|--------------|-------------------|
| E2E.P0.100 | primary + cross-layer + content-reliability + privacy/security + regression/out-of-scope-negative + hybrid/manual-required | Phase 1 / 2 / 3 / 4 / 5 / 6 / 7 / 8 | 3.4 / 4.3 / 5.4 / 7.1 / 7.2 / 8.8 |
