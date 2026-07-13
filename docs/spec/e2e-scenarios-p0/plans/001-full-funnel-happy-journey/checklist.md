# 001 Current Conversation Funnel Journey Checklist

> **版本**: 3.5
> **状态**: completed
> **更新日期**: 2026-07-13

**关联计划**: [plan](./plan.md)

## Phase 1: Real-path Red/Green

- [x] 1.1 RED-GREEN: Resume parse AI uses shared observability and writes `ai_task_runs`.
- [x] 1.2 RED-GREEN: Empty focus codes persist as non-null PostgreSQL `{}`.
- [x] 1.3 RED-GREEN: Completion uses lifecycle-only session-event columns.
- [x] 1.4 RED-GREEN: Report retry focus uses PostgreSQL `text[]`; generating retries are idempotent.

## Phase 2: Scenario reconciliation

- [x] 2.1 Rebase P0.098 onto current contract composition.
- [x] 2.2 Rebase P0.099 onto shared real-environment hybrid browser evidence.
- [x] 2.3 Delete orphaned dedicated Playwright full-funnel server/config/spec.
- [x] 2.4 Align P0.100 operation/profile terminology with continuous chat.

## Phase 3: Real browser acceptance

- [x] 3.1 Shared environment reset/redeploy and readiness verification pass.
- [x] 3.2 Real Mailpit login, resume/JD import and continuous message exchange pass.
- [x] 3.3 Voice is natively disabled and no structured question UI is visible.
- [x] 3.4 Completion/report generation reaches ready after recovery fixes.
- [x] 3.5 Desktop/mobile practice and report screenshots are recorded.
- [x] 3.6 BDD-Gate: P0.098 and P0.099 four-stage scripts pass with current evidence.

## Phase 4: Closeout

- [x] 4.1 Run focused/full backend/frontend, codegen, migration, prompt/eval and scenario gates.
- [x] 4.2 Run docs/index/diff and active negative-reference gates.
- [x] 4.3 Complete bug record, retrospective and work journal; restore owner documents to completed.

## Phase 5: P0.099 exact six-image report acceptance

- [x] 5.1 RED-GREEN: P0.099 setup/trigger/verify/cleanup require two isolated long-content real reports + generating state and exact six-image manifest; reject four-image/unnamed/cookie-bearing evidence.
  <!-- verified: 2026-07-12 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q -k p0_099_exact_six; python3 -m pytest scripts/lint/scenario_script_contract_test.py -q; setup.sh && trigger.sh && verify.sh && cleanup.sh" evidence="exact-six validator accepted the canonical six-row full-page manifest fixture and rejected five rows; 23/23 scenario-env and 9/9 scenario-script contracts pass; missing current browser manifest returns MANUAL_REQUIRED rather than PASS" -->
- [x] 5.2 Create current-run en/zh ready rows and capture their desktop/390x844 reports plus generating desktop/mobile。For each ready row, bind DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` and report/session/context digest。Desktop and mobile real report images cover action regions and actual `<=24-whitespace-word` / `<=64-Unicode-code-point` labels are fully visible with no clipping/ellipsis/hiding/overflow；P0.100 output digest is not a prerequisite。
  <!-- verified: 2026-07-13 run="e2e-p0-099-20260713T095144Z-12381" evidence="exact six full-page desktop/mobile images for ready-needs-practice, ready-well-prepared and generating; DB/API canonical report, screenshot, report/session and frozen-context digests bind; manual content audit passes; raw-debug absent; trigger+verify PASS" -->
- [x] 5.3 BDD-Gate: exact-six manifest closes every ready row's current-run canonical content/action/content-audit/screenshot/report/session/context digest chain；separate deterministic fixtures prove exact24/64 pixel parity，and200-code-point fuse or18/52 repair margin cannot satisfy UX PASS。
  <!-- verified: 2026-07-13 evidence="ui-design contract 54/54 and focused Playwright 34/34 PASS for exact en24/zh64 desktop/mobile; scenario privacy tests 38/38 and script tests 9/9 PASS" -->
