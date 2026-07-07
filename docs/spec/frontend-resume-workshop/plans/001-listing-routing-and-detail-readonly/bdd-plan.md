# 001 BDD Plan

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 类别 | 关联 Phase | 关联 Spec C-* | 关联 BDD-Gate（主 checklist） |
|---------|------|-----------|--------------|----------------------------|
| E2E.P0.036 | primary + auth-boundary · flat list render, route shell, fixture-derived rows, safe pending action, non-current negative | Phase 1 + 2 + 4 + 5 | C-1, C-2, C-10, C-11 | Phase 5.1 |
| E2E.P0.037 | primary + boundary + failure · detail preview, original modal, export 501, 404 fallback, privacy and a11y | Phase 1 + 3 + 4 + 5 | C-3, C-10, C-11 | Phase 5.2 |

## Phase 1 + 2 + 4 + 5: flat list + auth boundary

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.036 | Resume Workshop flat list + auth boundary | fixture-backed frontend client 可用；`listResumes` default / empty / paginated fixture 可用；用户可处于未登录或已登录态 | 未登录访问 `resume_versions`；已登录访问 list；点击 flat row open action；扫描 DOM / source / scenario evidence | 未登录先到 `auth_login` 且不触发 Resume API；已登录渲染 `ResumeWorkshopScreen` 和 flat table；每个 fixture item 对应 `resume-list-row-{resumeId}`；打开行进入 `resume_versions?resumeId=<id>&tab=preview`；tree/view-switcher/stats/版本参数/非当前 route testid/prototype runtime import 不出现；raw resume content 不进入 route / pending action / localStorage / generic logs | `test/scenarios/e2e/p0-036-resume-flat-list-auth-boundary/` |

## Phase 1 + 3 + 4 + 5: detail preview + original modal

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.037 | Resume detail preview + original modal + export 501 + 404 fallback | fixture-backed frontend client 可用；`getResume` default / not-found 和 `exportResume` unavailable fixture 可用；用户已登录 | 打开 `resume_versions?resumeId=<id>`；切换 `tab=rewrites`；点击 View original；点击 Export PDF；访问不存在的 resumeId | detail 默认 preview 且显式 tab 不被改写；Preview 渲染 structured projection；original modal 有 dialog / focus / ESC 行为；Export PDF 携带 `Idempotency-Key` 并显示 P0 unavailable toast，不写 localStorage；404 使用 generic copy 和返回列表 CTA，不回显 fixture `error.code`；证据不出现非当前 route testid 或 fallback 文本 | `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/` |
