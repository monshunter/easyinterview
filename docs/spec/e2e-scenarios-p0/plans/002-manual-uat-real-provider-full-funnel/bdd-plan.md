# BDD Plan

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-05-27

## Phase 5: Real Provider Hybrid Full Funnel

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.100 | 真实 provider hybrid 全漏斗 UAT | 本地 dev-stack 依赖 healthy（含 Mailpit）；schema 已 migrate；`deploy/dev-stack/.env` 是唯一真实本地 env 来源，包含 backend `APP_ENV=dev`、真实 auth secrets、`EMAIL_PROVIDER=mailpit`、真实 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 与 frontend `VITE_EI_API_MODE=real` / `VITE_EI_API_BASE_URL`；验收者使用 synthetic JD / resume / answer materials | AI Agent 先运行 `setup.sh` / `trigger.sh` / `verify.sh` / `cleanup.sh`；若 `deploy/dev-stack/.env` 缺真实凭证或缺浏览器证据则输出 `MANUAL_REQUIRED`；人工或浏览器 Agent 在同一输出目录补齐 Mailpit 登录、Home -> JD import -> Parse -> Workspace -> Practice -> Generating -> Report -> next_round、checklist / 截图 / 脱敏日志，且 `evidence.md` 必须包含与 `setup.env` 相同的 `RUN_ID` / `run_id` | (a) `E2E.P0.100` 登记在 e2e INDEX 且具备标准脚本契约；(b) 前端请求真实 backend base URL，无 fixture `Prefer` header；(c) backend 真实 handler / runner / PostgreSQL 产生资源；(d) AI 调用经真实 provider，证据仅记录 provider/profile/model/latency/task-run count，不记录 prompt/response 明文；(e) 账号、材料、环境、清理说明齐备，且账号入口走 Mailpit magic-link，不新增正式 `backend/cmd` 或直接 session bootstrap；(f) URL/storage/console 不泄露 JD/答案/报告明文；(g) 不使用 `APP_ENV=test`、P0.099 test server、deterministic stub AI 或场景专属 `.env` 作为完成证据；(h) `trigger.sh` 与 `verify.sh` 的 `scan_evidence_redline` 均通过，旧 `evidence.md` 或泄露 provider key / auth secret / session cookie / magic-link token / prompt body / response body 的证据只能得到 `MANUAL_REQUIRED`；(i) env consumer gate 覆盖 `env-setup.sh --with-migrations`、`env-redeploy.sh frontend` 与 `E2E.P0.100` trigger 对 `deploy/dev-stack/.env` 的消费 | `test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/` |

## Scenario Matrix

| 场景 ID | 类型 | 覆盖计划阶段 | 主 checklist Gate |
|---------|------|--------------|-------------------|
| E2E.P0.100 | primary + cross-layer + privacy/security + regression/legacy-negative + hybrid/manual-required | Phase 1 / 2 / 3 / 4 / 5 / 6 / 7 | 3.4 / 4.3 / 5.4 / 7.1 / 7.2 |
