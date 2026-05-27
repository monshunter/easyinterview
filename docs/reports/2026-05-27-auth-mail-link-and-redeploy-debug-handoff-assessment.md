# Auth Mail-Link And Redeploy Debug Handoff 交付复盘报告

> **日期**: 2026-05-27
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`BUG-0112` 登录/注册邮件链路修复，以及 `BUG-0113` 本地 redeploy 从 build-only 修订为 build + restart + debug handoff。
- 登录链路证据：generated client 已处理 `202` 空响应；frontend `/auth/verify?token=...` 自动兑换 session 并清理 token；Mailpit 最新邮件指向 frontend `/auth/verify` callback，不再指向 backend verify API。
- 环境证据：`test/scenarios/env-redeploy.sh all` 完成 dev-stack verify、backend build、frontend build，并重启 8080 backend 与 5173 frontend；输出 frontend/backend/Mailpit/MinIO 地址、日志路径和 PID 文件。
- 验证证据：focused frontend tests、backend auth/cmd/codegen tests、OpenAPI/codegen/lint/docs gates、scenario env contract pytest、`env-redeploy.sh all` live restart、Mailpit 最新邮件 callback 检查。

## 2 会话中的主要阻点/痛点

- `startAuthEmailChallenge` 合法返回 `202 Accepted` 且无 body，但 generated client 仍尝试 `JSON.parse("")`，导致页面无反馈。
- 邮件链接最初进入 backend verify API，用户无法回到前端登录态恢复路径；frontend token fallback 入口也不明显。
- `env-redeploy.sh all` 只刷新 build artifacts，没有重启宿主机 backend/frontend；浏览器仍连旧进程，Mailpit 新旧邮件混在一起，误导调试。
- 环境脚本没有稳定输出服务地址、日志路径、PID 文件和容器日志命令，开发者无法在 Agent 启动后直接接管。

## 3 根因归类

- `spec/plan`：auth flow 没有把 `202 no body`、frontend callback owner、token scrub 和 Mailpit callback 串成同一个回归 gate。
- `test`：OpenAPI fixture 曾用 `{}` 掩盖真实 `202` 空响应，generated-client 缺少空响应回归。
- `README/skill`：local-dev-stack 的 redeploy 语义停留在 artifact 刷新，没有明确 host-run 服务必须重启。
- `no repo change needed`：`202` 响应码本身正确；它表示邮件 challenge 已被接受并异步投递，最终 session 兑换才是 verify API 的 `200`。

## 4 对流程资产的改进建议

- `spec-plan` 高优先级：auth/email magic-link 相关计划必须同时列出 `startAuthEmailChallenge` 202 空 body、frontend `/auth/verify` callback、URL token scrub、CORS origin 来自 callback URL。
- `test` 高优先级：generated client 保留 202/204/empty body regression；fixture generator 不得用 `{}` 代表 no-body endpoint。
- `README/skill` 高优先级：scenario env/redeploy skill 和 dev-stack README 必须把 host-run redeploy 定义为 build + restart + endpoint/log/PID handoff。
- `test` 中优先级：local-dev-stack contract test 应继续反查 env scripts、skills 和 README，避免 redeploy 语义退回 build-only。

## 5 建议优先级与后续动作

- 最高优先级：用当前已重启的 `http://127.0.0.1:5173/` 再走一次浏览器登录，确认点击 Mailpit 最新邮件后 TopBar 进入已登录态。
- 其次：运行 `E2E.P0.100` hybrid trigger/verify，确保真实 provider 全漏斗场景也消费新的 auth callback 与 redeploy handoff。
- 提交前：用 `/work-journal` 记录本次 `BUG-0112` + `BUG-0113` 修复，并用 commit message 关联两个 Bug。
