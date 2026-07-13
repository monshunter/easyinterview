# 001 Current Conversation Funnel BDD Checklist

> **版本**: 3.5
> **状态**: completed
> **更新日期**: 2026-07-13

## E2E.P0.098

- [x] Scenario assets describe current operations and persistence.
- [x] Trigger runs current focused tests and verify rejects no-test/failure markers.
- [x] Four-stage lifecycle passes.

## E2E.P0.099

- [x] Shared real environment and Mailpit authentication were used.
- [x] Continuous chat and disabled voice were visually verified.
- [x] Ready report and current PostgreSQL persistence were verified.
- [x] Historical four desktop/mobile screenshots were captured; this evidence is superseded by Phase 5.
- [x] Historical current-run redacted evidence passed the prior contract.

## E2E.P0.099 Phase 5

- [x] Create current-run en/zh ready rows and bind each row's DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` and report/session/context digest；P0.100 output digest is not required.
  <!-- verified: 2026-07-13 run="e2e-p0-099-20260713T095144Z-12381" evidence="two current ready rows plus generating bind canonical DB/API, screenshot, report/session and frozen-context digests; manual content audit passes" -->
- [x] Capture exactly six redacted full-page images；desktop+390 real report images fully cover action regions and show actual `<=24-whitespace-word` / `<=64-Unicode-code-point` labels without clipping/ellipsis/hiding/overflow.
  <!-- verified: 2026-07-13 evidence="exact six full-page images cover three states on desktop/mobile; raw-debug absent; trigger+verify PASS" -->
- [x] Run prototype/formal pixel parity on deterministic ui-design/OpenAPI fixtures with exactly24 English whitespace words /64 zh-CN Unicode code points and prove complete wrapping；reject200-code-point-fuse-only or18/52-repair-margin-only evidence.
  <!-- verified: 2026-07-13 evidence="ui-design contract 54/54 and focused Playwright 34/34 PASS for exact en24/zh64 desktop/mobile; scenario privacy tests 38/38 and script tests 9/9 PASS" -->

> P0.100 的 active BDD checklist 只存在于 `002-manual-uat-real-provider-full-funnel` Phase 8。
