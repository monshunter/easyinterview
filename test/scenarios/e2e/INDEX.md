# E2E 场景索引

> 仅登记通过真实 HTTP API 或连接真实 backend 的浏览器 UI 驱动的场景；代码层单测统一由根 `make test` 承接。

## P0 核心闭环

| 场景 ID | 关联需求 | 目录 | 描述 | 执行方式 | 状态 |
|---------|----------|------|------|----------|------|
| E2E.P0.098 | e2e-scenarios-p0 D-1 / §6 | `p0-098-practice-completion-progress-refresh/` | 真实邮箱登录与 completion API 后，Home / Workspace / TargetJob 投影一致刷新到下一轮 | automated | Ready |
| E2E.P0.099 | e2e-scenarios-p0 D-2 / §6; frontend-report-dashboard C-7 | `p0-099-report-generating-live-ui/` | 真实 frontend/backend/provider 下的 report / generating 六图与 live report API 验收 | hybrid | Ready |
| E2E.P0.101 | backend-auth C-9; frontend-shell C-8; local-dev-stack C-10/C-15 | `p0-101-auth-email-code-profile-setup/` | 真实 frontend/backend/Mailpit 的邮箱验证码登录、首次资料补全与再次登录闭环 | automated | Ready |
