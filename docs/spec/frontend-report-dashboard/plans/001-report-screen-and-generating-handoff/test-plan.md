# Honest Grounded Report Screen Test Plan

> **版本**: 3.4
> **状态**: completed
> **更新日期**: 2026-07-14

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
- Prototype/formal component Playwright compares DOM、computed style、bbox、viewport and screenshot diff at 1440/390 under fixed locale/time/DPR/fonts/motion. This is code-level visual regression, not E2E.
- UUID sentinel tests reject report/session IDs from visible text、tooltip、ARIA and accessible names while keeping target/round/resume visible.

## 5 ReportsScreen and route recovery

- Table tests cover current-target canonical join、current/latest-only display、loading/empty/error、cross-target/mismatch fail closed and stale response fences.
- Route tests cover trusted target Back to Reports、untrusted fallback to Workspace、reportId-only report routes and Reports Back directly to targetJobId-only Workspace detail without Parse detour.
- Source negatives prove ReportsScreen is the sole list consumer and no TopBar/global/history report entry or compatibility query is introduced.

## 6 Real E2E handoff

- P0.099 alone captures current real report/generating desktop/mobile UI and binds authenticated API/read-only DB evidence.
- Code-level exact boundary/parity and provider/eval results are independent; neither is copied into E2E PASS markers.
