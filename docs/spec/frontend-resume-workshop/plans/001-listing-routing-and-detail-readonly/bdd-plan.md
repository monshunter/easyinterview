# 001 BDD Plan

> **版本**: 2.0
> **状态**: completed
> **更新日期**: 2026-07-08

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 类别 | 关联 Phase | 关联 Spec C-* | 关联 BDD-Gate（主 checklist） |
|---------|------|-----------|--------------|----------------------------|
| E2E.P0.036 | primary + auth-boundary · flat list render, duplicate create CTA absence, delete action, route shell, fixture-derived rows, safe pending action, non-current negative | Phase 1 + 2 + 4 + 5 + 6 | C-1, C-2, C-8, C-10, C-11 | Phase 5.1 / 6.5 |
| E2E.P0.037 | primary + boundary + failure · waiting state, source-format detail body, Markdown body purity, unified reading surface, generic-name negative, legacy tab ignored, removed actions, 404 fallback, privacy and a11y | Phase 1 + 3 + 4 + 5 + 6 + 8 + 10 | C-3, C-10, C-11 | Phase 5.2 / 6.5 / 8.1 / 8.2 / 10.2 |

## Phase 1 + 2 + 4 + 5: flat list + auth boundary

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.036 | Resume Workshop flat list + auth boundary | fixture-backed frontend client 可用；`listResumes` default / empty / paginated fixture 可用；`archiveResume` fixture 可用；用户可处于未登录或已登录态 | 未登录访问 `resume_versions`；已登录访问 list；点击 flat row open / delete action；扫描 DOM / source / scenario evidence | 未登录先到 `auth_login` 且不触发 Resume API；已登录渲染 `ResumeWorkshopScreen` 和 flat table；每个 fixture item 对应 `resume-list-row-{resumeId}`；Header “新建简历”是唯一创建入口，底部 upload/paste CTA 不出现；打开行进入 `resume_versions?resumeId=<id>`；删除成功隐藏 row，失败保留 row 并提示；tree/view-switcher/stats/版本参数/非当前 route testid/prototype runtime import 不出现；raw resume content 不进入 route / pending action / localStorage / generic logs | `test/scenarios/e2e/p0-036-resume-flat-list-auth-boundary/` |

## Phase 1 + 3 + 4 + 5: read-only detail

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.037 | Resume detail waiting + source-format content + removed actions + 404 fallback | fixture-backed frontend client 可用；`getResume` default / pending-upload / failed-empty / failed-with-snapshot / not-found fixture 可用；`getResumeSource` default / not-found fixture 可用；用户已登录 | 打开 `resume_versions?resumeId=<id>`；带旧 `tab=rewrites` / `tailorRunId` 打开；访问不存在的 resumeId | Pending 且无正文时渲染等待动画并轮询；upload PDF ready 后渲染从上到下平铺的 PDF page stack，source 指向 `/api/v1/resumes/{resumeId}/source`，不显示 browser PDF viewer toolbar / native viewer shell / Markdown fallback；paste、Markdown upload 和 TXT upload 以 Markdown 标题 / 列表 / 段落 / inline DOM 渲染，不能只显示 txt 段落或 structured projection，且 Markdown body card 内不得额外注入 displayName / header 名称 / summary / source metadata；PDF 与 Markdown 共用阅读背景板，Markdown 正文也位于背景板内的白色 page surface；failed 且无正文时显示失败态；failed 或已有可读正文时同一详情 URL 只请求一次并展示终态；显式旧 tab 不 materialize；通用上传/粘贴名称和 raw 第一行名称不出现；Export PDF / Copy text / View original / original modal / Rewrites / Edit / PreviewConfirm DOM 均不存在；不调用 `exportResume` / `requestResumeTailor` / detail `updateResume`；404 使用 generic copy 和返回列表 CTA，不回显 fixture `error.code` | `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/` |
