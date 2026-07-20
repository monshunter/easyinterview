# Honest Grounded Report Screen Test Checklist

> **版本**: 3.19
> **状态**: completed
> **更新日期**: 2026-07-20

**关联 Test Plan**: [test-plan](./test-plan.md)

## Generating and report behavior

- [x] Polling schedule、pause/resume fences、single-run cap and truthful typed state tests pass.
- [x] Direct report contract、empty/non-empty focus、CTA request and mixed UI/report language tests pass.
- [x] Invalid/over-limit payloads fail closed without raw label, truncation or rewrite.

## Layout and privacy

- [x] Exact 24/64 deterministic fixtures wrap completely at desktop/mobile; 25/65 fail closed.
- [x] Historical formal frontend DOM/style/bbox/viewport tests through Phase 11 pass as code-level visual regression.
- [x] Report/session UUID sentinel DOM/a11y negatives and target/round/resume/interview-record positives pass；frozen-resume URL and privacy-preserving conversation action remain usable.

## Phase 12 report summary hierarchy

- [x] RED proves the current third top readiness+summary metric violates the revised design before implementation.
- [x] DOM tests prove the previously delivered ready group order/count `3/2/2/2/1` and that Overall Summary follows Next Actions；the current four-item Context Strip revision is owned by the unchecked Phase 12 gates below.
- [x] Semantic/i18n tests prove Summary Metrics contain only dimension/evidence counts；Overall Summary contains localized title/readiness plus the unchanged server `summary` exactly once.
- [x] Formal frontend 1440/390 style/bbox/viewport/a11y tests prove desktop full-width bottom summary、mobile same-order single column、complete wrapping and zero horizontal overflow.
- [x] Source negatives reject top readiness/summary、duplicate summary、parallel prototype ownership and any backend/API/persistence/prompt change.

## ReportsScreen and routing

- [x] Current-target isolation、canonical join、current/latest-only、loading/empty/error and stale-response fence tests pass.
- [x] Trusted/untrusted Back matrix、reportId-only route and direct Workspace detail without Parse detour tests pass.
- [x] ReportsScreen-only list consumer and no TopBar/global/history/compatibility entry negatives pass.

## Historical E2E separation and full regression through Phase 11

- [x] P0.099 independently binds current real report/generating API/DB/screenshot evidence; mock/component outputs cannot satisfy it.
- [x] Provider/eval and deterministic parity remain independent code gates, not E2E steps.
- [x] Root `make test` runs the complete frontend/backend unit regression for phase completion.

## Phase 12 current gates

- [x] Root `make test`、typecheck/build/lint and document gates pass on the revised implementation.
- [x] `ConversationReport` and responsive contract tests pass for the four peer context children、canonical resume href、privacy-preserving conversation action、SPA clicks、desktop equal-height detail pairs and mobile single-column reset.
  <!-- verified: 2026-07-15 frontend full suite 126 files / 1003 tests PASS; Chrome desktop/mobile geometry and navigation PASS. -->
- [x] P0.099 README/manual-audit/capture-verification assertions are aligned with four context items、responsive detail-pair alignment、actions and the following bottom interview summary；the current ready behavior is separately accepted through real-backend Chrome desktop/mobile evidence, without reusing historical exact-six images or duplicating unchanged generating/language resources.
  <!-- verified: 2026-07-15 P0.099 evidence tests 8 PASS; current focused Chrome screenshots saved under .test-output/acceptance/report-context-grid/. -->

## Phase 13 report conversation integration

- [x] Focused ReportConversation/ReportsScreen/Report route and safe-Markdown tests pass.
- [x] Focused backend report conversation service/store/handler tests pass with malformed-ID no-read and hidden-404 boundaries.
- [x] OpenAPI fixture/codegen/inventory/diff gates pass and `listPracticeSessions` has no active fixture/generated/runtime consumer.
- [x] Root `make test` passes while deleted Demo/prototype-sync assets remain absent.
  <!-- verified: 2026-07-15 method=focused+root-regression evidence="frontend 989; Python 551+4493 subtests; Go all packages; build/docs/codegen/pruning PASS" -->

## Phase 14 failed report recovery

- [x] Generated-client bodyless POST/IK/typed response contract test passes.
- [x] ReportsScreen failed state, locator separation, oversize, pending/double-click/key/stale/malformed matrices pass.
- [x] Accessibility and raw-error/internal-ID negative tests pass.
- [x] Failed conversation hides Back while its report owner is resolving, then routes a trusted target to Reports or a resolved untrusted result to Workspace.
  <!-- verified: 2026-07-16 method=focused-vitest evidence="ReportConversationScreen 19/19 PASS; the RED first exposed an immediate workspace Back while getFeedbackReport was pending" -->
- [x] Focused frontend, typecheck/build and root `make test` pass after GREEN.
  <!-- verified: 2026-07-16 evidence="focused report recovery/conversation/client suites, typecheck and build PASS; root frontend 126/1026 PASS" -->

## Phase 15 completed-session conversation availability

- [x] RED captures missing latest-attempt conversation action beside queued/generating progress.
- [x] queued/generating/latest-ready/current/same-ID/empty action matrix passes with localized distinct a11y names.
- [x] Focused frontend, typecheck/build and Chrome real-environment acceptance pass after GREEN.
- [x] Root `make test` passes after GREEN.
  <!-- verified: 2026-07-16 evidence="Python 584/4583 subtests, Go all packages, frontend 126/1026 PASS" -->

## Phase 16 reference-aligned report dashboard

- [x] Source/component/responsive tests pass for the 1336px shared grid, Header CTA hierarchy, semantic icons, rounded surfaces, `4/2/2/2/1` and 390px containment.
- [x] Focused owner tests, root regression and formal-frontend fixture Chrome 1916×821 / 390×844 visual checks pass without claiming a real ready-report or full `E2E.P0.099`.<!-- verified: 2026-07-19 evidence="report 30 focused tests; root frontend 1054 tests; typecheck/build; documentOverflow=0 at both viewports" -->

## Phase 17 complete target-composition rebuild

- [x] RED/GREEN source and component tests cover the single Context surface/dividers, four Detail icon kinds, 1432px composition and removal of ready inline visual assembly.<!-- verified: 2026-07-19 evidence="RED missing data-icon/1432 composition; GREEN 2 files/23 tests PASS" -->
- [x] Behavior tests prove Dimensions retain localized status/confidence while Highlights/Issues omit duplicate confidence and preserve server evidence verbatim.<!-- verified: 2026-07-19 evidence="ConversationReport icon-led/no-duplicate-confidence regression PASS" -->
- [x] Current real-report Chrome desktop bbox, screenshot and zero-clipping checks pass at the actually observed viewport; typical desktop content fully shows Overall Summary in the first viewport. The `390×844` boundary is separately covered by the deterministic responsive/component contract unless an exact mobile Chrome viewport is run.<!-- verified: 2026-07-19 evidence="real report desktop 1920x964 and exact mobile 390x844 full-page PASS; overflowX=0; viewport override reset" -->
- [x] Focused owner tests, root regression, typecheck/build and document gates pass after GREEN.<!-- verified: 2026-07-19 evidence="23 focused; root Python 615/4615, Go all, frontend 1055; typecheck/build/docs PASS" -->

## Phase 18 report list and conversation reference composition

- [x] RED/GREEN source and component tests cover the 1372px shared shell, illustrated Header, target summary card, numbered timeline, independent round cards and primary/secondary action hierarchy.<!-- verified: 2026-07-19 evidence="expected structural RED followed by final owner 32 files/242 tests PASS" -->
- [x] RED/GREEN ReportConversation tests cover the three-column Context Strip, 60px role avatars, shared assistant/user outlined message cards and matching avatar silhouettes without changing role colors or reportId-only/safe-Markdown/privacy behavior.<!-- verified: 2026-07-19 evidence="ReportConversation behavior 19 tests plus source visual contract 2 tests PASS; desktop cards/badges share geometry and surface" -->
- [x] Desktop/mobile responsive and current-run Chrome evidence prove shared bounds, same-order single-column containment and zero horizontal overflow without claiming `E2E.P0.099`.<!-- verified: 2026-07-19 evidence="desktop 1920x964 and exact mobile 390x844 PASS; conversation badges 60/48px, shared radius; overflowX=0" -->
- [x] Focused owner tests, root regression, typecheck/build and document gates pass after GREEN.<!-- verified: 2026-07-19 evidence="owner 242; root frontend 1057 plus Python/Go; typecheck/build/redeploy PASS" -->

## Phase 19 report generating transition

- [x] RED/GREEN component tests覆盖 shared report scene、TopBar、中心卡、真实状态和返回动作，并拒绝 percent/phase/notification。
- [x] Polling/visibility/reportId/trusted Back/terminal error 回归保持通过。
- [x] Responsive/reduced-motion contract 与 current-run desktop Chrome 视觉验收通过，不声明完整 `E2E.P0.099` PASS。（1090px centered card；两次真实 ready handoff；console clean。）

## Phase 20 cardless report generating transition

- [x] RED/GREEN tests 覆盖 `card` surface 删除和共享画布直出，并证明报告真实状态与恢复行为未回归。<!-- verified: 2026-07-20 evidence="Expected two-test RED followed by AsyncTransitionScene/GeneratingScreen 12/12 GREEN; affected owner regression included in 65 files/495 tests PASS." -->
- [x] Desktop/mobile Chrome 记录中心内容透明背景、零边框/阴影、统一返回控件、无横向溢出和零 browser warning/error。<!-- verified: 2026-07-20 evidence="Real Generating at 1512x777 and 390x844: rgba(0,0,0,0), border 0, shadow none, backdrop-filter none, Back=返回, overflowX=0, console clean." -->
- [x] Focused、typecheck/build 与根 `make test` 通过。<!-- verified: 2026-07-20 evidence="Affected owner regression 65 files/495 tests; typecheck and production build PASS; root make test PASS with Python 615/4615 subtests, all Go packages, and frontend 136 files/1107 tests." -->

## Phase 21 target-aware Generating Back copy

- [x] Locale/source tests 区分 Generating trusted Reports 专用文案与其它页面/Workspace fallback 的 shared Back。<!-- verified: 2026-07-20 method=focused-vitest evidence="reportDashboardI18nCoverage and backNavigationCopy pass 28/28, including exact zh/en values, the two approved owner sources, common.back retention and retired-key negatives." -->
- [x] Generating waiting/error component tests 同时证明 exact zh/en label 与原导航矩阵不变。<!-- verified: 2026-07-20 method=focused-vitest evidence="GeneratingScreen and GeneratingBackNavigation pass 15/15 for trusted reports copy in waiting/error states, Workspace fallback copy and the existing route matrix." -->
- [x] Focused、typecheck/build、根 `make test` 与 real Chrome trusted Back 验收通过。<!-- verified: 2026-07-20 evidence="Generating/i18n focused 6 files / 72 tests PASS; typecheck and production build PASS; final root make test PASS with 615 tests / 4615 subtests; real Chrome captured the exact trusted label on the transient Generating state." -->

## Phase 22 truthful target-summary date label

- [x] ReportsScreen zh/en 正向标签与旧“面试日期 / Interview date”负向断言通过；i18n parity、focused 与根回归通过。<!-- verified: 2026-07-20 evidence="Focused ReportsScreen/locale 38 tests and root frontend 1115 tests PASS." -->
