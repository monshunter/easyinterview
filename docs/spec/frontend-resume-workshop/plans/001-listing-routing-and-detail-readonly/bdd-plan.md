# 001 BDD Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-11

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 类别 | 关联 Phase | 关联 Spec C-* | 关联 BDD-Gate（主 checklist） |
|---------|------|-----------|--------------|----------------------------|
| E2E.P0.036 | primary + alternate · resume list 默认渲染 + tree/flat 切换 + StatsStrip + 旧入口负向 + UI parity | Phase 1 + 2 + 4 + 5 | C-1, C-2, C-3, C-5, C-6, C-7, C-9 | Phase 5.5 |
| E2E.P0.037 | primary + boundary · resume detail Preview Tab + 原件弹层 + 默认 tab + 404 fallback + UI parity | Phase 3 + 4 + 5 | C-4, C-5, C-6, C-7, C-8 | Phase 5.6 |

---

## Phase 1 + 2 + 4 + 5: resume list 默认渲染 + tree/flat 切换

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.036 | resume list 默认渲染 + tree/flat ViewSwitcher + StatsStrip + i18n + UI parity + retired 负向 | A2 dev stack 拉起；fixture-backed mock-first 启用；用户已登录；fixture `listResumes.json` `default` (5 resume_asset) + `listResumeVersions.json` `default` (12 version 含 3 master + 9 targeted)；用户 lang=zh-CN | （A）加载 `/resume_versions` 路由；（B）点击 ViewSwitcher 切换 flat；（C）点击某 ResumeVersionRow 进入详情（导航行为）；（D）lang 切换 zh → en | （A1）路由替换：`ResumeWorkshopScreen` 渲染而非 `PlaceholderScreen`；TopBar `topbar-nav-resume_versions` 高亮；（A2）StatsStrip 4 项 testid 命中：`resume-workshop-stats-originals/versions/top-match/recent`，数字与 fixture 一致；（A3）ViewSwitcher testid `resume-workshop-view-switcher-tree` (active) / `-flat` 渲染；（A4）默认 tree 视图：5 个 `resume-tree-row-{originalId}` + 12 个 `resume-version-row-{versionId}` 按 originalId 分组 + 折叠可展开；（A5）"选为底稿" / "基于这棵树新建版本" 按钮渲染但 click 显示 toast "即将开放"；（B1）点击 `view-switcher-flat`：渲染 `ResumeFlatView`，12 个 `resume-flat-row-{versionId}` 按 `match DESC nullsLast / updated_at DESC` 排序；（C1）点击 version row → `nav("resume_versions", { versionId })`；URL 含 `versionId` 参数；（D1）lang 切换：StatsStrip 文案 zh→en；`buildResumeData(en)` 投影；generated client 请求 `Accept-Language: en` 携带；（E UI parity）desktop 1440px / mobile 390x844 baseline screenshot 容差 ≤ 0.1%；DOM bounding box parity 命中；（F 旧入口）grep `mistake|growth|drill|onboarding-legacy|experiences-legacy` in `frontend/src/app/screens/resume-workshop/` 0 命中；grep `import.*ui-design/src/data` 0 命中（不允许运行时 import data.jsx）；（G 隐私）raw text / parsed_summary 0 出现在 console.log / URL / localStorage / mock transport log；（H mock-first）字节比对 generated client 期望与 fixture `default` scenario shape 一致 | `test/scenarios/e2e/p0-036-resume-list-tree-flat-toggle/` |

## Phase 3 + 4 + 5: resume detail Preview Tab + 原件弹层

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.037 | resume detail Preview Tab 主路径 + 默认 tab + 原件弹层 a11y + 404 fallback + UI parity | A2 dev stack 拉起；fixture `getResumeVersion.json` `master-default` / `targeted-with-suggestions` / `not-found-404`；用户已登录；lang=zh-CN | （A）从 list view 点击某 MASTER version → 详情；（B）从 list view 点击某 TARGETED version → 详情；（C）点击 "查看原件" → 原件弹层；（D）原件弹层中 ESC / 外层遮罩 / X 按钮；（E）keyboard Tab 在 modal 内 focus trap；（F）访问不存在的 versionId | （A1）渲染 `ResumeDetailView`：testid `resume-detail-breadcrumb` / `resume-detail-branch-graph` / `resume-detail-tab-{preview,rewrites,edit}` / `resume-detail-preview-content` 全部命中；默认 tab=preview (MASTER 用 resumeDefaultTab)；rewrites / edit 渲染 `<ComingSoonTab>` 占位（testid `resume-detail-tab-content-coming-soon-{rewrites,edit}`）；（A2）Preview 内容来自 `buildResumePlainText(zh, masterVersion)` 投影；（B1）TARGETED 进入：默认 tab P0 fallback preview（rewrites 未实现）；rewrites / edit 仍渲染 ComingSoonTab；（C1）点击 `resume-detail-view-original` 按钮 → 原件弹层 modal 渲染 `original-modal-content` + 原件 text；DOM `role="dialog"` + `aria-labelledby` 完整；（D1）ESC 关闭 / 外层遮罩点击关闭 / X 按钮关闭 → focus 回到 view-original 触发按钮；（E1）modal 内 Tab key 在按钮 / 关闭按钮间循环，不漏出 modal；（F1）访问 `getResumeVersion(non-existent)` 返回 404 fixture → 渲染 `<NotFoundEmptyState>` + "返回列表" CTA；（G UI parity）detail view + modal desktop / mobile screenshot 容差 ≤ 0.1%；focus ring computed style 与 ui-design 一致；（H 隐私）原件 text 在 DOM 中渲染但不出现在 console / URL / localStorage；modal 关闭后 DOM unmount（不持久）；（I a11y）axe-core check pass：role / aria / contrast / focus visible | `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/` |
