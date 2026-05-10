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
| E2E.P0.006 | frontend-shell C-9 | `p0-006-ui-design-pixel-parity-gate/` | D2 follow-up Playwright + chromium pixel parity：desktop+mobile viewport DOM/computed style/bounding box/screenshot regression + dark/customAccent 状态 | automated | Ready |
| E2E.P0.010 | backend-targetjob C-1/C-3/C-6/C-7/C-12 | `p0-010-targetjob-text-import-parse-ready/` | manual_text TargetJob 导入、异步解析、列表、详情、更新与 idempotency | automated | Ready |
| E2E.P0.011 | backend-targetjob C-2/C-3/C-9 | `p0-011-targetjob-url-import-fetch-and-parse/` | URL TargetJob 导入、SSRF 守护、抓取 snapshot、解析与 source_refresh 占位 | automated | Ready |
| E2E.P0.012 | backend-targetjob C-4/C-5/C-10 | `p0-012-targetjob-parse-failure-retryable/` | TargetJob parse 失败 retryable / non-retryable 语义与隐私红线 | automated | Ready |
| E2E.P0.013 | backend-targetjob C-3/C-6/C-9/C-11/C-13 | `p0-013-targetjob-manual-form-ready/` | manual_form TargetJob 同步 ready、terminal job、列表详情与 no-runner 断言 | automated | Ready |
| E2E.P0.014 | frontend-home-job-picks-and-parse C-1, C-4 | `p0-014-home-default-render/` | Home 默认渲染：empty/non-empty/12+ 三态，DOM 锚点，排序，TopBar 高亮 | automated | Ready |
| E2E.P0.015 | frontend-home-job-picks-and-parse C-2, C-3, C-6 | `p0-015-jd-import-and-parse/` | Paste/Upload/URL → import → parse loading → preview 主路径 + failed | automated | Ready |
| E2E.P0.016 | frontend-home-job-picks-and-parse C-5, C-7 | `p0-016-parse-confirm-to-workspace/` | Parse 编辑 + Confirm → workspace + auth pending action | automated | Ready |
| E2E.P0.017 | frontend-home-job-picks-and-parse C-8 | `p0-017-jd-match-placeholder/` | jd_match 三 tab 数据驱动 smoke + 旧 prototype 负向 grep（plan 002 接力升级） | automated | Ready |
| E2E.P0.018 | frontend-workspace-and-practice C-2, C-7, C-8, C-9 | `p0-018-workspace-default-render/` | Workspace 默认渲染：plan eyebrow + header + Interview Launcher + Main Left/Right + Modals | automated | Ready |
| E2E.P0.019 | frontend-workspace-and-practice C-2, C-3, C-8, C-9 | `p0-019-workspace-context-loading/` | Workspace context loading：empty/missing-resume 空态 + getPracticePlan refresh | automated | Ready |
| E2E.P0.020 | frontend-workspace-and-practice C-1, C-3, C-12 | `p0-020-workspace-start-practice/` | 立即面试 双步契约 + Idempotency-Key + pendingAction 未登录恢复 | automated | Ready |
| E2E.P0.021 | frontend-workspace-and-practice C-7, C-9, C-10, C-12 | `p0-021-workspace-handoff/` | Workspace handoff + 隐私红线 + legacy negative grep | automated | Ready |
| E2E.P0.022 | frontend-home-job-picks-and-parse C-12, C-13, C-15 | `p0-022-jd-match-recommended-and-confirm/` | jd_match Recommended tab 主路径 + 4 button 闭环 + auth pending action + 隐私反查 | automated | Ready |
| E2E.P0.023 | frontend-home-job-picks-and-parse C-14, C-15 | `p0-023-jd-match-search-and-saved/` | jd_match Search tab + Saved searches + 4 chip filter + 5 步 AGENT panel + failure + auth gate + privacy | automated | Ready |
| E2E.P0.024 | frontend-home-job-picks-and-parse C-16 | `p0-024-jd-match-watchlist-and-signals/` | jd_match Watchlist tab + Market signals + chevron handoff + boundary + privacy | automated | Ready |
| E2E.P0.025 | frontend-home-job-picks-and-parse C-12, C-15 | `p0-025-jd-match-profile-and-agent-status/` | jd_match Profile chip + AGENT scan status + Auth pending action 跨 tab 综合 | automated | Ready |
| E2E.P0.026 | frontend-home-job-picks-and-parse C-13 | `p0-026-jd-match-confirm-interview-handoff/` | Confirm interview from jd_match → parse 出口 params 完整性 + parse 屏不破坏 | automated | Ready |
