# 001 Real API/UI Journeys Checklist

> **版本**: 4.1
> **状态**: active
> **更新日期**: 2026-07-15

**关联计划**: [plan](./plan.md)

## Phase 1: E2E boundary reconciliation

- [x] 1.1 Delete scenario ownership that only wrapped code tests, builds, lint, fixture parity or provider CLI/eval.
- [x] 1.2 Keep P0.098 and P0.099 on the shared host-run real stack; reject fixture transport/dev mock/in-process handler evidence.
- [x] 1.3 Keep focused tests for development feedback and require root `make test` as the independent backend/frontend unit regression gate.
- [x] 1.4 Register P0.101 as the real frontend/backend/Mailpit asset whose business contract remains owned by backend-auth/frontend-shell.

## Phase 2: P0.098 completion/progress

- [x] 2.1 The tracked Playwright flow signs in through real frontend/backend/Mailpit and completes the pre-seeded waiting session through the real completion API.
- [x] 2.2 The tracked flow reloads Home and Workspace, opens TargetJob detail, and asserts `current,pending,pending` becomes `done,current,pending` consistently with the real API.
- [x] 2.3 ASSET-GATE: application requests are not intercepted or fulfilled; the scenario does not create a round-2 plan.

## Phase 3: P0.099 report/generating

- [x] 3.1 The tracked runbook requires isolated current-run en/zh ready report resources plus one honest generating resource in the real stack.
- [x] 3.2 The tracked runbook captures exactly six redacted `fullPage: true` images at 1440x1200 and 390x844 and binds each row to current API/DB status and report/session/context/screenshot digests.
- [x] 3.3 ASSET-GATE: the runbook requires direct no-OCR review of ready/generating state, complete action region, clipping/ellipsis/hidden content/overflow and raw private content.
- [x] 3.4 CONVERSATION-GATE: extend the real flow with Report → Conversation → Back；route uses only reportId, API/DB binding and strict sequence digests agree, transcript prose is not stored, and no public session-list request occurs.
  <!-- verified: 2026-07-15 method=tdd-static evidence="six scenario-local tests cover reportId-only navigation, real authenticated API/read-only PostgreSQL projection, strict sequence digest, zero list requests and redaction; runner does not invoke the code-level test" -->
- [x] 3.5 EVIDENCE-GATE: screenshot directory/manifest/manual audit remain exactly six images；conversation adds only bounded non-image evidence and cannot introduce a seventh screenshot.
  <!-- verified: 2026-07-15 method=static-contract evidence="manifest/manual audit schemas remain exactly six rows; conversation-navigation.json is bounded non-image evidence only" -->
- [x] 3.6 PRIVACY-SCOPE-GATE: TDD updates P0.099 evidence validation so project user data and secrets remain fail-closed, while benign development metadata such as PNG `iCCP` is accepted and PNG integrity/digest checks remain enforced.
  <!-- verified: 2026-07-15 method=tdd command="python3 test/scenarios/e2e/p0-099-report-generating-live-ui/scripts/test_report_conversation_evidence.py" result="7 PASS; iCCP accepted; ei_session metadata rejected" -->

## Phase 4: Closeout

- [x] 4.1 Run root `make test` as the independent whole-repository backend/frontend unit regression gate.
- [x] 4.2 Keep owner-specific codegen/migration/lint/build/prompt/eval gates independent; none is an E2E step or marker.
- [x] 4.3 Run static scenario structure/syntax/interception, docs/index/diff and deleted-ID negative checks. `E2E.P0.099` also completed host-run setup/trigger/verify/cleanup and the exact-six Chrome/manual audit against the current real environment.
  <!-- verified: 2026-07-15 commands="python3 test/scenarios/e2e/p0-099-report-generating-live-ui/scripts/test_report_conversation_evidence.py; py_compile; bash -n; env-verify.sh; setup.sh; trigger.sh; Chrome exact-six capture/no-OCR audit; verify.sh; cleanup.sh" result="7 PASS; P0_099_MANUAL_VISUAL_AUDIT_BOUND_PASS; verify PASS" -->
- [ ] 4.4 BDD-Gate: run `E2E.P0.098` against the current real environment and record current-run PASS.
- [x] 4.5 BDD-Gate: run `E2E.P0.099` against the current real environment, complete exact-six no-OCR audit plus bounded conversation navigation/API/DB evidence, and record current-run PASS.
  <!-- verified: 2026-07-15 run_id="e2e-p0-099-20260715T021319Z-57232" result="PASS" evidence="exact six Chrome full-page screenshots; live API/PostgreSQL/conversation/back binding; manual-visual-audit.json; bounded redaction" -->
- [ ] 4.6 BDD-Gate: run `E2E.P0.101` against the current real environment and record current-run PASS.
