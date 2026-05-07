# E2E 场景索引

> 场景按阶段分组，记录编号、关联需求、目录路径和状态。

---

## P0 核心闭环

| 场景 ID | 关联需求 | 目录 | 描述 | 执行方式 | 状态 |
|---------|----------|------|------|----------|------|
| E2E.P0.001 | frontend-shell C-1 | `p0-001-default-home-shell/` | 默认进入首页并呈现五入口 TopBar 与用户菜单 | automated | Ready |
| E2E.P0.002 | frontend-shell C-2 | `p0-002-auth-pending-action-resume/` | 登录打断后恢复原业务动作与上下文 | automated | Ready |
| E2E.P0.003 | backend-auth C-1 | `p0-003-passwordless-session-cookie/` | 邮箱挑战验证后签发 first-party session 并支持 /me 与 logout | automated | Ready |
| E2E.P0.004 | frontend-shell C-7 | `p0-004-app-shell-language-switch/` | App Shell 中英语言切换并携带 Accept-Language display hint | automated | Ready |
| E2E.P0.005 | frontend-shell C-8 | `p0-005-app-shell-visual-system-smoke/` | D2 视觉系统 smoke：DOM/className/CSS-variable/customAccent overlay/legacy 负向 + ui-design 源追溯 | automated | Ready |
