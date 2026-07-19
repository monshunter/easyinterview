# Honest Grounded Report Screen BDD Plan

> **版本**: 4.5
> **状态**: completed
> **更新日期**: 2026-07-19

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.REPORT.UI.001` | report 处于 generating、ready、failed 或 identity/context 不合法 | 页面轮询、恢复、展示报告、打开冻结简历副本/面试记录或触发 replay/next/back | 状态与 CTA 只来自 API truth；ready desktop 以约 1432px 目标网格按 `4/2/2/2/1` 展示，Context 是一张四列内部竖线分隔整卡，四个 Detail card 带语义 icon 且 evidence 不重复 confidence；典型合法内容的 Overall 在 2048×917 首屏完整出现，长内容和 mobile 同序完整换行；typed failure/route recovery fail closed 且不泄露 private IDs | `frontend/src/app/screens/generating/__tests__/GeneratingScreen.test.tsx` + `frontend/src/app/screens/report/__tests__/ConversationReport.test.tsx` + `ReportResponsiveContract.test.ts`，由根 `make test` 承接 |
| `BDD.REPORT.CONVERSATION.001` | owned report 存在，状态为 queued/generating/ready/failed，消息投影可为合法或损坏 | 用户从 Report 或 ReportsScreen 打开记录、切换 reportId 或返回父页 | ReportsScreen 对每个不同 current/latest locator 独立显示记录入口，progress/regenerate 只作为并列动作；只以 reportId 读取并按 sequence 显示安全只读 Markdown；ready 返回 Report、queued/generating 返回 Generating、failed 以可信 target 直接返回 ReportsScreen；跨用户/乱序/非法 role/stale response 整体 fail closed，无 session list/live controls/internal IDs | `frontend/src/app/screens/report-conversation/__tests__/ReportConversationScreen.test.tsx` + `frontend/src/app/screens/reports/__tests__/ReportsScreen.test.tsx`，由根 `make test` 承接 |
| `BDD.REPORT.REGENERATE.UI.001` | latest attempt 为普通 failed、超限 failed，或旧 ready 与更新 failed 并存 | 用户查看记录、点击重新生成、重试未知网络结果、另一 tab 已改变状态或切换 target | 普通 failed 使用同 reportId/稳定 IK 进入 matching Generating；双击单请求；超限仅记录；旧/新动作 locator 与 accessible name 可区分；typed state conflict 重读 current target + overview，stale/malformed/raw/unknown error fail closed | `frontend/src/app/screens/reports/__tests__/ReportsScreen.test.tsx` + generated-client contract tests，由根 `make test` 承接 |
| `BDD.REPORT.RECORDS.VISUAL.002` | 当前 TargetJob 有 canonical rounds/current/latest，且 report-owned transcript 有合法 assistant/user 消息 | 用户浏览报告列表并打开面试记录 | ReportsScreen 以 1372px Header illustration、真实事实摘要卡和编号时间线展示独立轮次卡；ReportConversation 使用同宽 Header、三列 Context Strip，assistant/user 共用浅色整行卡片、描边、圆角、内边距和同宽方形头像轮廓，只以蓝色 AI / 灰色“我”区分身份；所有 locator/status/regenerate/Back/Markdown/privacy 行为不变，mobile 同序无横溢 | `ReportsScreen.test.tsx` + `ReportConversationScreen.test.tsx` domain behavior tests；current-run Chrome 仅作 UI 证据 |

## Real E2E handoff

| ID | Type | Phase | Given | When | Then |
|----|------|-------|-------|------|------|
| E2E.P0.099 | real full-stack report/generating UI | 12 | shared host-run frontend/backend/provider with current-run en/zh ready reports and one honest generating resource | Chrome captures exact six full-page desktop/mobile images and runner binds authenticated report API plus read-only DB evidence | current ready/generating state is visible；each row binds current report/session/context/screenshot digests；ready four-item context、equal-height detail pairs、action regions and following bottom interview summary are complete with no clipping/ellipsis/hiding/overflow |

## Evidence boundary

- Polling、typed failure、CTA、server-owned focus、route recovery、ReportsScreen isolation、`4/2/2/2/1` DOM/geometry and deterministic visual regression are frontend/backend code tests, not E2E scenarios.
- Exact 24/64 boundary uses deterministic code-level fixtures. P0.099 only proves current legal real report content is fully visible at desktop/mobile.
- Fixture transport、dev mock、jsdom、route interception、component test server or Playwright against mock data cannot satisfy P0.099.
- Root `make test` runs the complete frontend/backend unit regression independently and is never an E2E step or PASS marker.
- Phase 14/15 current-run Chrome extension automation skill must exercise the real local failed-row regenerate/conversation recovery path and queued/generating progress-plus-record coexistence；code-owner behavior additionally proves failed conversation cannot expose Back before its trusted owner resolves. Chrome evidence is required delivery evidence but does not by itself satisfy or mark `E2E.P0.099` PASS.
