# 001 BDD Plan

> **版本**: 2.5
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 类别 | 关联 Phase | 关联 Spec C-* | 关联 BDD-Gate（主 checklist） |
|---------|------|-----------|--------------|----------------------------|
| E2E.P0.036 | primary + auth-boundary · flat summary list, duplicate create CTA absence, delete action, route shell, fixture-derived rows, safe pending action, StrictMode single transport/retry, out-of-scope negative | Phase 1 + 2 + 4 + 5 + 6 + 19 | C-1, C-2, C-8, C-10, C-11, C-12 | Phase 5.1 / 6.5 / 19.5 |
| E2E.P0.037 | primary + boundary + failure · full detail only, waiting state, source-format detail body, Markdown body purity, unified reading surface, StrictMode single initial transport/retry, sequential polling, generic-name negative, out-of-scope tab ignored, removed actions, 404 fallback, privacy and a11y | Phase 1 + 3 + 4 + 5 + 6 + 8 + 10 + 19 | C-3, C-10, C-11, C-13 | Phase 5.2 / 6.5 / 8.1 / 8.2 / 10.2 / 19.6 |

## Phase 1 + 2 + 4 + 5: flat list + auth boundary

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.036 | Resume Workshop flat summary list + auth boundary | fixture-backed frontend client 可用；`listResumes` default / empty / paginated fixture 外层保持 `PaginatedResume` 且 `items` 为 `ResumeSummary[]`；`archiveResume` fixture 可用；用户可处于未登录或已登录态 | 未登录访问 `resume_versions`；在 React StrictMode 下已登录访问 list；令首个 list transport reject 后点击 retry；点击 flat row open / delete action；扫描 response keys、DOM、source 与 scenario evidence | 未登录先到 `auth_login` 且不触发 Resume API；已登录渲染 `ResumeWorkshopScreen` 和 flat table；每个 item exact keys 仅为 `id,title,displayName,language,sourceType,parseStatus,summaryHeadline,hasReadableContent,updatedAt`，详情 forbidden fields absent；相同初始 request identity 的底层 transport 恰好 1 次；reject 后 retry 发起新 transport 并成功；每个 fixture item 对应 `resume-list-row-{resumeId}`；Header “新建简历”是唯一创建入口，底部 upload/paste CTA 不出现；打开行进入详情；删除成功隐藏 row，失败保留 row 并提示；tree/view-switcher/stats/版本参数/范围外 route testid/prototype runtime import 不出现；raw resume content 不进入 route / pending action / localStorage / generic logs | `test/scenarios/e2e/p0-036-resume-flat-list-auth-boundary/` |

## Phase 1 + 3 + 4 + 5: read-only detail

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.037 | Resume full detail waiting + source-format content + removed actions + 404 fallback | fixture-backed frontend client 可用；`getResume` full-detail default / pending-upload / failed-empty / failed-with-snapshot / not-found fixture 可用；`getResumeSource` default / not-found fixture 可用；用户已登录 | 在 React StrictMode 下打开 ready / pending `resume_versions?resumeId=<id>`；令首个 detail transport reject 后点击 retry；带 out-of-scope `tab=rewrites` / `tailorRunId` 打开；访问不存在的 resumeId | 完整正文只来自 `getResume`，不由 `listResumes` item 透传；ready 初始相同 request identity 的底层 transport恰好 1 次；reject 后 retry 发起新 transport并成功；pending 且无正文时渲染等待动画，且只在上一次请求 settle 后发起下一次轮询；upload PDF ready 后渲染从上到下平铺的 PDF page stack，source 指向 `/api/v1/resumes/{resumeId}/source`，不显示 browser PDF viewer toolbar / native viewer shell / Markdown fallback；paste、Markdown upload 和 TXT upload 以 Markdown DOM 渲染，且 body card 不注入 displayName / header / summary / source metadata；PDF 与 Markdown 共用阅读背景板；failed 且无正文时显示失败态；failed 或已有可读正文时不继续轮询；显式 out-of-scope tab 不 materialize；通用名称、removed actions 均不存在；404 使用 generic copy，不回显 fixture `error.code` | `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/` |
