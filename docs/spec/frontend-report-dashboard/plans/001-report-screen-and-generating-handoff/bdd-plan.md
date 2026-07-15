# Honest Grounded Report Screen BDD Plan

> **版本**: 3.7
> **状态**: active
> **更新日期**: 2026-07-15

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.REPORT.UI.001` | report 处于 generating、ready、failed 或 identity/context 不合法 | 页面轮询、恢复、展示报告或触发 replay/next/back | 状态与 CTA 只来自 API truth；ready desktop 按 `3/2/2/2/1` 展示，mobile 同序单列，readiness 与唯一服务端 summary 位于四个内容区之后的全宽面试总评；typed failure/route recovery fail closed 且不泄露 private IDs | `frontend/src/app/screens/generating/__tests__/GeneratingScreen.test.tsx` + `frontend/src/app/screens/report/__tests__/ConversationReport.test.tsx`，由根 `make test` 承接 |
| `BDD.REPORT.CONVERSATION.001` | owned report 存在，状态为 queued/generating/ready/failed，消息投影可为合法或损坏 | 用户从 Report 或 ReportsScreen 打开记录、切换 reportId 或返回父页 | 只以 reportId 读取并按 sequence 显示安全只读 Markdown；四状态返回正确父页；跨用户/乱序/非法 role/stale response 整体 fail closed，无 session list/live controls/internal IDs | `frontend/src/app/screens/report-conversation/__tests__/ReportConversationScreen.test.tsx` + `frontend/src/app/screens/reports/__tests__/ReportsScreen.test.tsx`，由根 `make test` 承接 |

## Real E2E handoff

| ID | Type | Phase | Given | When | Then |
|----|------|-------|-------|------|------|
| E2E.P0.099 | real full-stack report/generating UI | 12 | shared host-run frontend/backend/provider with current-run en/zh ready reports and one honest generating resource | browser captures exact six full-page desktop/mobile images and runner binds authenticated report API plus read-only DB evidence | current ready/generating state is visible；each row binds current report/session/context/screenshot digests；ready action regions and following bottom interview summary are complete with no clipping/ellipsis/hiding/overflow |

## Evidence boundary

- Polling、typed failure、CTA、server-owned focus、route recovery、ReportsScreen isolation、`3/2/2/2/1` DOM/geometry and deterministic visual regression are frontend/backend code tests, not E2E scenarios.
- Exact 24/64 boundary uses deterministic code-level fixtures. P0.099 only proves current legal real report content is fully visible at desktop/mobile.
- Fixture transport、dev mock、jsdom、route interception、component test server or Playwright against mock data cannot satisfy P0.099.
- Root `make test` runs the complete frontend/backend unit regression independently and is never an E2E step or PASS marker.
