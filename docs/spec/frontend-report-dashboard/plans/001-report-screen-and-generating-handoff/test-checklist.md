# Honest Grounded Report Screen Test Checklist

> **版本**: 3.2
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1-5: Historical baseline

- [x] Historical conversation report regression tests remain present.

## Phase 6: UI truth and generating

- [x] Phase 6 ownership/source-contract and state-action truthful generating tests pass.
  <!-- verified: 2026-07-12 evidence="ui-design 50/50; full frontend 111 files/762 tests; generating desktop/mobile parity and action matrix PASS" -->
- [x] Phase 6 maxAttempts49/1.5×1.5/cap8 exact timing tests pass；queued/generating during action-local10s/20s/40s waits stays honest，attempt/progress is absent，async job attempts remain independent，and poll exhaustion does not fabricate terminal failed.
  <!-- verified: 2026-07-13 evidence="seven focused frontend files/51 tests PASS; P0.058 v3 timing/reset/separation composition PASS" -->
- [x] Phase 6 in-flight and timer pause/resume tests preserve attempt/delay，resume n+1，reject reset1/repeat-n/concurrency，and keep one run<=49. <!-- verified: 2026-07-13 method=fake-clock-vitest evidence="focused hook17/17; generating23/23; full789 PASS" -->

## Phase 7: Direct dashboard and handoff

- [x] Phase 7 generated direct-contract/frozen-context/empty-generic-focus/non-empty-issue-backed-focus/route-tamper/language-split/CTA/responsive and server-owned handoff tests pass.
  <!-- verified: 2026-07-13 evidence="P0.057 six frontend files/49 tests PASS; P0.070/P0.072 PostgreSQL focus/isolation markers PASS; P0.059 12 desktop/mobile parity tests PASS" -->

## Phase 8: Parity and real UAT

- [x] Phase8 exact-six/current-run audit plus1440x1200+390x844 parity prove legal24/64 labels fully visible；over-limit fixture is typed invalid/no raw。P0.100 output、200-code-point fuse and18/52 repair margin are excluded from UX PASS.
  <!-- verified: 2026-07-13 evidence="P0.099 e2e-p0-099-20260713T095144Z-12381 trigger+verify PASS; exact six/three states/current DB-API digests/manual audit/raw-debug absent" -->

## Phase 9: Internal locator cleanup

- [x] Distinct UUID sentinel DOM/a11y negatives, internal contract/CTA positives, locale zero-reference and refreshed P0.059 parity pass；separate `/agent-browser` acceptance contains only the same ready report's exact formal 1440x1200/390x844 full-page PNGs plus manifest, with recomputed path/hash/state/viewport/fullPage and report/session sentinel-absence audit all passing.
  <!-- verified: 2026-07-14 evidence="Context Strip acceptance 20260714-final contains exactly the two required PNGs and manifest; hashes 36b859b2.../1b0c3ed3..., state/viewport/fullPage and UUID DOM/a11y negatives validate." -->

## Phase 10: Independent current-plan reports list and Back recovery

- [x] ReportsScreen current-target isolation、canonical join、current/latest-only、different-ready status、loading/empty/error、stale-response fence、Back-to-Parse and 1440/390 parity pass；Report/Generating trusted Back-to-Reports、workspace fallback、reportId-only route、single list-consumer and no-TopBar/no-section negatives remain green in P0.058/P0.059.
  <!-- verified: 2026-07-14 evidence="Validator 40/40, ReportsScreen 15/15, final P0.058 setup/trigger/verify/cleanup PASS and P0.059 16/16 Playwright PASS cover the complete matrix." -->
