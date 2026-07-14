# 001 Real API/UI Journeys Checklist

> **版本**: 3.8
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
- [ ] 3.4 CONVERSATION-GATE: extend the real flow with Report → Conversation → Back；route uses only reportId, API/DB binding and strict sequence digests agree, transcript prose is not stored, and no public session-list request occurs.
- [ ] 3.5 EVIDENCE-GATE: screenshot directory/manifest/manual audit remain exactly six images；conversation adds only bounded non-image evidence and cannot introduce a seventh screenshot.

## Phase 4: Closeout

- [x] 4.1 Run root `make test` as the independent whole-repository backend/frontend unit regression gate.
- [x] 4.2 Keep owner-specific codegen/migration/lint/build/prompt/eval gates independent; none is an E2E step or marker.
- [x] 4.3 Run static scenario structure/syntax/interception, docs/index/diff and deleted-ID negative checks. No real environment run is claimed; scenarios remain `Ready` until an explicit `/scenario-run`.
- [ ] 4.4 BDD-Gate: run `E2E.P0.098` against the current real environment and record current-run PASS.
- [ ] 4.5 BDD-Gate: run `E2E.P0.099` against the current real environment, complete exact-six no-OCR audit plus bounded conversation navigation/API/DB evidence, and record current-run PASS.
- [ ] 4.6 BDD-Gate: run `E2E.P0.101` against the current real environment and record current-run PASS.
