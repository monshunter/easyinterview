# Event and Outbox Contract Bootstrap Checklist

> **版本**: 1.13
> **状态**: active
> **更新日期**: 2026-07-13

**关联计划**: [plan](./plan.md)

## 1 Event truth source

- [x] 1.1 HISTORICAL-FOUNDATION: pre-Phase 9 `shared/events.yaml` defined the 13-event inventory; Phase 9 owns the current 12-event contraction.
- [x] 1.2 Event domains are limited to `target`, `practice`, `report`, `resume`, `source`, and `privacy`.
- [x] 1.3 Event payload fields use B1 enum refs where available and event-local enums where B1 ownership does not apply.
- [x] 1.4 `scripts/lint/events_inventory.py` validates event count, names, producers, payload inventory, domain set and enum boundaries.

## 2 Job truth source

- [x] 2.1 HISTORICAL-FOUNDATION: pre-Phase 9 `shared/jobs.yaml` defined 8 canonical job types; Phase 9 owns the current 7-job contraction.
- [x] 2.2 HISTORICAL-FOUNDATION: API-facing subset was fixed to 6 while two jobs were internal-only; Phase 9 leaves only `email_dispatch` internal-only.
- [x] 2.3 `email_dispatch` payload schema only allows audit fields and lists redacted fields for auth token, auth URL, email address and email body boundaries.
- [x] 2.4 Inventory lint validates job count, dotted mapping, API-facing subset and redaction schema.

## 3 Codegen and baselines

- [x] 3.1 `backend/cmd/codegen/events` remains a B3-owned generator independent from B1 conventions codegen.
- [x] 3.2 `make codegen-events` renders backend event/job packages, frontend event/job packages, JSON Schema files and baseline manifests.
- [x] 3.3 Generated artifacts are deterministic and missing generated contract files fail fast.
- [x] 3.4 B1 enum JSON Schema refs are read from `shared/conventions.yaml`, not duplicated in generator code.

## 4 Lint and tests

- [x] 4.1 `make lint-events` rejects bare event/job literals outside generated packages.
- [x] 4.2 `make lint-events` verifies generated event and job sets match `shared/events.yaml` / `shared/jobs.yaml`.
- [x] 4.3 Go tests cover generator behavior, shared event/job packages and email_dispatch redaction.
- [x] 4.4 Frontend tests cover generated event/job types, envelope round-trip and emailDispatch helpers.

## 5 Downstream contract

- [x] 5.1 B4 owns outbox/async_jobs DDL and consumes B3 field names and job_type set.
- [x] 5.2 `backend-async-runner` owns dispatcher polling, retry and handler execution.
- [x] 5.3 F1/F2/C-domain owners consume metrics, analytics namespace and producer/consumer rules from this contract.

## 6 Current owner compression gate

- [x] 6.1 HISTORICAL-EVIDENCE: owner docs/index previously aligned to the then-current 14-event / 8-job / 6 API-facing contract; Phase 9 owns the current 12/7/6 projection.
  <!-- verified: 2026-07-07 method=current-owner-compression evidence="Updated event-and-outbox-contract spec.md to v2.12, plan.md to v1.10, checklist.md to v1.9, context specVersion to v2.12, and synced docs/spec plus event-and-outbox plans INDEX. PASS: targeted stale-wording grep returned no matches; validate_context.py event-and-outbox-contract/001 backend PASS; python3 scripts/lint/events_inventory.py shared/events.yaml shared/jobs.yaml shared/conventions.yaml PASS; make codegen-events PASS; make lint-events PASS; go test ./backend/cmd/codegen/events ./backend/internal/shared/events ./backend/internal/shared/jobs -count=1 PASS; pnpm --dir frontend test src/lib/events src/lib/jobs PASS (11 tests); make codegen-check PASS." -->

## 7 Canonical generator entrypoint

- [x] 7.1 删除 production-dead 的 `Run` / `RunFromBytes` wrappers，generator tests 改用 `RunWithConventions` / `RunFromBytesWithConventions` 真实入口；验证 production deadcode、symbol inventory、generator/shared Go tests、event lint/codegen drift、owner contexts 与 docs/diff/pruning gates。
  <!-- verified: 2026-07-10 method=event-codegen-wrapper-removal evidence="Production deadcode RED listed Run and RunFromBytes. Deleted both implicit-path wrappers; CLI/tests now share explicit-conventions entrypoints. Generator/shared Go tests, frontend event/job 11 tests, staticcheck, inventory/lint/codegen drift, symbol inventory and owner contexts PASS." -->

## 8 Inventory test naming

- [x] 8.1 HISTORICAL-EVIDENCE: inventory tests were renamed from 16/10 to the then-current 14/8 contract；Phase 9 owns current 12/7 names and assertions.
  <!-- verified: 2026-07-10 method=event-inventory-test-name-reconciliation evidence="Renamed only the two stale count-bearing test methods to 14-event and 8-job. Focused collection and 50 tests, event lint/codegen drift, three Go packages, frontend event/job 11 tests, zero generated consumer diff, both owner contexts and docs/diff/pruning gates PASS." -->

## 9 OPENAPI-002 paste-only event/job contraction

- [x] 9.1 RED/GREEN: inventory/generator tests require 12 events / 5 domains / 7 canonical jobs / 6 API-facing jobs; remove target import source discriminator plus refresh event/local enum/job/dotted task/priority mapping while retaining `targetJobId,userId,targetLanguage` and independent `source_records` persistence.
  <!-- verified: 2026-07-13 method=inventory+generator-red-green evidence="RED reported missing source discriminator/refresh expectations and 13-vs-12 schema count; GREEN has exact 12 events across 5 domains, 7 jobs, 6 API-facing jobs and target.import.requested payload targetJobId/userId/targetLanguage" -->
- [x] 9.2 GENERATED/BASELINE-GATE: regenerate Go/TS types, JSON Schemas/refs and both v1 manifests; preserve before/after audit and reject baseline-only re-freeze or compatibility aliases.
  <!-- verified: 2026-07-13 commands="make codegen-events; make lint-events; Python inventory/lint 50 tests; Go generator/shared events/jobs; frontend generated events/jobs 7 tests" result="PASS; removed hand-written mapper plus generated refresh schema/enum refs; before events/jobs baselines sha256=8df8f5f0.../39605ced... counts=13/8/6, after sha256=7371e762.../8fda2801... counts=12/7/6; truth source and generated artifacts changed with baselines" -->
- [ ] 9.3 HANDOFF/ZERO-REFERENCE: backend runner Phase 7, B4 and TargetJob producers/consumers compile/test; current truth/generated/schema/baseline/runtime surfaces contain zero positive removed enum/event/job/task/queue references outside history/negative tests, explicitly excluding `source_records`.
- [ ] 9.4 BDD-N/A: inventory/generator/contract tests, `make codegen-events`, `make lint-events`, Go/TS generated-package tests, baseline drift and runner focused tests replace BDD; validate context/diff before restoring completed.
