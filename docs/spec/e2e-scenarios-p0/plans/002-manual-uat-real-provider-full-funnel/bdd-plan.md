# BDD Plan

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-05-26

## Phase 3: Manual UAT Real Provider Full Funnel

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.100 | 真实 provider 人工全漏斗 UAT | 本地 dev-stack 依赖 healthy（含 Mailpit）；schema 已 migrate；backend 以 `APP_ENV=dev`、真实 auth secrets、`EMAIL_PROVIDER=mailpit`、真实 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 启动；frontend 以 `VITE_EI_API_MODE=real` 指向 backend；验收者使用 synthetic JD / resume / answer materials | 验收者在浏览器中输入 `manual-uat-full-funnel@example.test`，从 Mailpit 打开 magic link 完成登录，走 Home -> JD import -> Parse -> Workspace -> Practice -> Generating -> Report -> next_round，并记录 checklist / 截图 / 脱敏日志 | (a) 前端请求真实 backend base URL，无 fixture `Prefer` header；(b) backend 真实 handler / runner / PostgreSQL 产生资源；(c) AI 调用经真实 provider，证据仅记录 provider/profile/model/latency/task-run count，不记录 prompt/response 明文；(d) 账号、材料、环境、清理说明齐备，且账号入口走 Mailpit magic-link，不新增正式 `backend/cmd` 或直接 session bootstrap；(e) URL/storage/console 不泄露 JD/答案/报告明文；(f) 不使用 `APP_ENV=test`、P0.099 test server 或 deterministic stub AI 作为完成证据 | `test/scenarios/manual-uat/full-funnel/` |

## Scenario Matrix

| 场景 ID | 类型 | 覆盖计划阶段 | 主 checklist Gate |
|---------|------|--------------|-------------------|
| E2E.P0.100 | primary + cross-layer + privacy/security + regression/legacy-negative | Phase 1 / 2 / 3 / 4 | 3.4 / 4.3 |
