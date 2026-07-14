# Honest Grounded Report Screen BDD Plan

> **版本**: 3.6
> **状态**: active
> **更新日期**: 2026-07-15

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.REPORT.UI.001` | report 处于 generating、ready、failed 或 identity/context 不合法 | 页面轮询、恢复、展示报告或触发 replay/next/back | 状态与 CTA 只来自 API truth；ready 内容完整可读，typed failure/route recovery fail closed 且不泄露 private IDs | `frontend/src/app/screens/generating/__tests__/GeneratingScreen.test.tsx` + `frontend/src/app/screens/report/__tests__/ConversationReport.test.tsx`，由根 `make test` 承接 |
| `BDD.REPORT.CONVERSATION.001` | owned report 已创建且具有 ordered terminal messages，状态可为 queued/generating/ready/failed | 用户从 Report 或 ReportsScreen current report 行打开会话记录并返回 | 以 reportId-only 显示安全 Markdown 只读 transcript；无 live controls/内部 IDs/session list，Back 返回同一父报告；missing/cross-user/invalid payload fail closed | `frontend/src/app/screens/report/__tests__/ReportConversationScreen.test.tsx` + generated client/route/source parity tests，由根 `make test` 承接 |

## Real E2E handoff

| ID | Type | Phase | Given | When | Then |
|----|------|-------|-------|------|------|
| E2E.P0.099 | real full-stack report/generating/conversation UI | 8/12 | shared host-run frontend/backend/provider with current-run en/zh ready reports and one honest generating resource | browser keeps the exact six report/generating images, then performs Report -> Conversation -> Back while runner binds authenticated report-conversation API plus read-only DB evidence | exact-six visual contract remains unchanged；the real transcript belongs to the same report and returns to the same parent state without exposing session/message IDs |

## Evidence boundary

- Polling、typed failure、CTA、server-owned focus、route recovery、ReportsScreen isolation and deterministic parity are frontend/backend code tests, not E2E scenarios.
- Exact 24/64 boundary uses deterministic code-level fixtures. P0.099 only proves current legal real report content is fully visible at desktop/mobile.
- Fixture transport、dev mock、jsdom、route interception、component test server or Playwright against mock data cannot satisfy P0.099.
- Root `make test` runs the complete frontend/backend unit regression independently and is never an E2E step or PASS marker.
