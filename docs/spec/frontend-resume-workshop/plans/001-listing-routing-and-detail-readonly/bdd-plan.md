# 001 BDD Plan

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 类别 | 关联 Phase | 关联 Spec C-* | 关联 BDD-Gate（主 checklist） |
|---------|------|-----------|--------------|----------------------------|
| E2E.P0.036 | primary + auth-boundary · flat list render, route shell, fixture-derived rows, safe pending action, non-current negative | Phase 1 + 2 + 4 + 5 | C-1, C-2, C-10, C-11 | Phase 5.1 |
| E2E.P0.037 | primary + boundary + failure · read-only original-content detail, generic-name negative, legacy tab ignored, removed actions, 404 fallback, privacy and a11y | Phase 1 + 3 + 4 + 5 | C-3, C-10, C-11 | Phase 5.2 |

## Phase 1 + 2 + 4 + 5: flat list + auth boundary

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.036 | Resume Workshop flat list + auth boundary | fixture-backed frontend client 可用；`listResumes` default / empty / paginated fixture 可用；用户可处于未登录或已登录态 | 未登录访问 `resume_versions`；已登录访问 list；点击 flat row open action；扫描 DOM / source / scenario evidence | 未登录先到 `auth_login` 且不触发 Resume API；已登录渲染 `ResumeWorkshopScreen` 和 flat table；每个 fixture item 对应 `resume-list-row-{resumeId}`；打开行进入 `resume_versions?resumeId=<id>`；tree/view-switcher/stats/版本参数/非当前 route testid/prototype runtime import 不出现；raw resume content 不进入 route / pending action / localStorage / generic logs | `test/scenarios/e2e/p0-036-resume-flat-list-auth-boundary/` |

## Phase 1 + 3 + 4 + 5: read-only detail

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.037 | Resume detail read-only original content + removed actions + 404 fallback | fixture-backed frontend client 可用；`getResume` default / pending-upload / not-found fixture 可用；用户已登录 | 打开 `resume_versions?resumeId=<id>`；带旧 `tab=rewrites` / `tailorRunId` 打开；访问不存在的 resumeId | Detail 优先渲染 `parsedTextSnapshot` / `originalText` 原文，不能只显示 structured projection；pending upload 原文快照为空时在详情页内轮询到快照可见；显式旧 tab 不 materialize；通用上传/粘贴名称和 raw 第一行名称不出现；Export PDF / Copy text / View original / original modal / Rewrites / Edit / PreviewConfirm DOM 均不存在；不调用 `exportResume` / `requestResumeTailor` / detail `updateResume`；404 使用 generic copy 和返回列表 CTA，不回显 fixture `error.code` | `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/` |
