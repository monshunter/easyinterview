# Honest Grounded Report Screen Test Plan

> **版本**: 3.17
> **状态**: completed
> **更新日期**: 2026-07-20

## 1 Unit-test ownership

- Frontend code tests live with Report、Generating、ReportsScreen、routing、i18n and generated-client consumers.
- Focused Vitest/Playwright component tests may be used during development; phase completion and CI run root `make test` for the entire backend/frontend unit regression.
- Code-level tests、typecheck、build、lint、source contract and deterministic parity are not wrapped in `test/scenarios/e2e`.

## 2 Generating truthfulness

- Fake-clock tests cover queued/generating polling、49-call cap、1.5×1.5 backoff with 8s cap、hidden/blur pause during wait/in-flight request、resume at n+1、no duplicate/concurrent request and explicit continue-check reset.
- Component tests cover typed failed/not-found/invalid/context-too-large/client-timeout/network states and truthful Back/continue actions without fake progress or internal attempt leakage.

## 3 Report contract and actions

- Generated-client/fixture/component tests cover summary、immutable context、code+label dimensions、evidence、focus and fail-closed unknown/malformed fields.
- Request tests prove replay/next sends no client-derived focus/settings and consumes server-owned identity.
- Language tests prove UI chrome and report prose use their separate language sources without translating model output.

## 4 Layout, privacy and parity

- Deterministic tests cover wire fuse、English 24/25 whitespace words、zh-CN 64/65 code points、U+FEFF/U+0085 delimiter parity and typed invalid/no raw over-limit output.
- Formal frontend component/browser tests assert the ready DOM order and exact group counts `4/2/2/2/1`: Context Strip 4、Summary Metrics 2、two detail rows of 2、one bottom Overall Summary.
- Link/action contract tests prove the frozen resume child exposes canonical `/resume-versions?resumeId=...` href with SPA/copy/new-tab semantics；the interview-record child uses an in-strip button action so reportId remains absent from DOM attributes.
- Desktop geometry tests prove each detail pair has matching top/bottom bounds and the shorter card fills its row with internal whitespace；mobile returns each panel to content height in one column.
- Source/semantic assertions prove readiness and `summary` are absent from the top metrics, the localized Overall Summary contains both, and the server `summary` renders exactly once without client rewrite.
- At 1440 the Overall Summary spans the full content grid after Next Actions；at 390 every group is single-column in the same DOM order. Computed style、bbox、viewport、wrapping、scroll width and accessible names are code-level gates, not E2E.
- Formal frontend deterministic visual regression uses fixed locale/time/DPR/fonts/motion and controlled actual/expected/diff artifacts; no parallel prototype runtime is an acceptance source.
- UUID sentinel tests reject report/session IDs from visible text、tooltip、ARIA and accessible names while keeping target/round/resume/interview record visible, the frozen-resume canonical URL usable, and conversation navigation free of reportId DOM attributes.

## 5 ReportsScreen and route recovery

- Table tests cover current-target canonical join、current/latest-only display、loading/empty/error、cross-target/mismatch fail closed and stale response fences.
- Route tests cover trusted target Back to Reports、resolved untrusted fallback to Workspace、failed owner resolving without an actionable Back、reportId-only report routes and Reports Back directly to targetJobId-only Workspace detail without Parse detour.
- Source negatives prove ReportsScreen is the sole list consumer and no TopBar/global/history report entry or compatibility query is introduced.

## 6 Real E2E handoff

- P0.099 alone captures current real report/generating desktop/mobile UI and binds authenticated API/read-only DB evidence；Phase 12 first aligns its README/manual visual audit and capture/verification contract so ready full-page images explicitly include the action region and following bottom Overall Summary.
- Code-level exact boundary/parity and provider/eval results are independent; neither is copied into E2E PASS markers.
- Historical exact-six evidence cannot satisfy the revised hierarchy；only an explicitly run current environment produces the Phase 12 E2E result.

## 7 Report conversation integration

- `ReportConversationScreen` tests cover loading/empty/ready/unavailable states, reportId switch stale fences, strict closed projection, safe Markdown/GFM, ready/non-ready Back destinations and the failed-owner resolving fence before any Back action is rendered.
- `ReportsScreen` and `ReportDashboard` tests prove both entry points navigate to the same reportId-only route without adding a Header CTA.
- Backend service/store/handler tests cover owner authorization, malformed locator no-read, report-to-session binding, strict ordered projection and `Cache-Control: no-store`.
- OpenAPI/fixture/codegen and deleted-session-list negatives prove `getReportConversation` is the only public read surface; root `make test` is the aggregate code regression.

## 8 Failed report recovery

- Generated-client tests lock bodyless POST, required IK, credentials and typed `ReportWithJob` response.
- ReportsScreen tests cover failed-only and old-ready/new-failed rows, oversize exclusion, pending accessibility, double-click suppression, stable-key uncertain retry, explicit-error key reset, target-switch stale fencing and strict response identity validation.
- Error tests prove localization uses typed codes only and never renders provider body, request ID, raw message or report UUID.

## 9 Completed-session conversation availability

- Table tests cover queued/generating/latest-ready-different-current with an independent latest conversation action bound to `latestAttempt.id`.
- same-ID current/latest ready renders one conversation action；both-null renders none；existing failed/oversize/current locator tests remain unchanged.
- zh/en a11y assertions and real Chrome prove progress/regenerate actions never replace the visible interview-record action.

## 10 Reference-aligned report dashboard

- `ReportResponsiveContract.test.ts` and ready-report component/source tests first reject the narrow `1120px` inline-styled implementation, then require a class-based approximately 1336px shared grid, rounded Context/Metric/Detail/Overall surfaces, semantic SVG icons and Header CTA hierarchy.
- Existing `4/2/2/2/1`, frozen context, privacy, failure, replay/next/back and 390px single-column assertions remain the behavior fence. Real ready-report Chrome inspection is scoped visual evidence and does not claim a complete `E2E.P0.099` run.

## 11 Complete target-composition rebuild

- `ReportResponsiveContract.test.ts` locks `1432px`, one Context surface with internal dividers, four semantic Detail icon kinds, class-owned ready styling and desktop/mobile breakpoints; it must fail on a width-only shell even when overflow is zero.
- `ConversationReport.test.tsx` locks the icon-led panel structure and proves Highlights/Issues do not repeat confidence while Dimensions continue to expose localized status/confidence.
- Chrome geometry on the current real ready report checks Header, Context, Metrics, both Detail rows and Overall bboxes in the user's current desktop viewport, including full Overall visibility for the current typical payload. The `390×844` mobile boundary remains covered by the deterministic responsive/component contract unless an exact mobile Chrome viewport is actually executed. These are scoped UI acceptance artifacts, not a complete `E2E.P0.099` run.

## 12 Report list and conversation reference composition

- `ReportsScreen.test.tsx` adds structure/source assertions for the approximately `1372px` shell, decorative Header illustration, real-fact target summary card, numbered timeline, independent round cards and primary/secondary actions while preserving the full current/latest/status/regeneration matrix.
- `ReportConversationScreen.test.tsx` adds structure/source assertions for the same Header, three-column Context Strip, approximately `60px` role avatars, and one shared outlined full-width message-card silhouette for assistant/user while preserving role colors, reportId-only loading/failure/Back/safe-Markdown/privacy behavior.
- Responsive CSS/source tests require desktop shared bounds and mobile same-order single-column containment; current-run Chrome records desktop/mobile screenshots and bbox/overflow evidence against the real frontend/backend without promoting the scoped run to `E2E.P0.099`.

## 13 Report generating transition

- `GeneratingScreen` tests覆盖 queued/generating/recoverable/terminal、reportId switch、visibility pause/resume、trusted Back 和无伪百分比/阶段/通知。
- Shared scene/route/CSS tests覆盖 report illustration、中心白卡、TopBar/Interview active、indeterminate 语义、reduced-motion 与 390px containment。
- Current-run Chrome desktop/mobile 只作为本次视觉验收；完整 `E2E.P0.099` 仍需其独立真实六图矩阵。

## 14 Cardless report generating transition

- `GeneratingScreen` 与 shared-scene source tests 先拒绝 `card` prop/class 及白色背景、边框、阴影、毛玻璃，再确认报告等待内容直接复用共享氛围画布。
- 既有 queued/generating/recoverable/terminal、reportId switch、visibility、trusted Back、TopBar、responsive 与 reduced-motion tests 保持通过。
- Chrome desktop/mobile 检查中心内容 computed background transparent、border 0、box-shadow none、统一“返回 / Back”控件、无横向溢出及 browser warning/error 为零。

## 15 Target-aware Generating Back copy

- `reportDashboardI18nCoverage.test.ts` 与 `backNavigationCopy.test.ts` 先要求新增 `generating.backToReports` 的 zh/en exact copy，并把 Generating 明确登记为 shared `common.back` 的唯一受控 owner 例外。
- `GeneratingScreen.test.tsx` / `GeneratingBackNavigation.test.tsx` 覆盖 polling 与 typed error 两类 trusted Reports 目标显示专用文案并保持原 route；无可信 identity 的 Workspace fallback 继续显示通用文案。
- Chrome 在真实 queued/generating 页面验证“返回面试报告”及其 target-scoped Reports destination；code-level tests 与真实浏览器证据分层记录。
