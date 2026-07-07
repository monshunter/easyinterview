# Event and Outbox Contract Bootstrap Checklist

> **版本**: 1.9
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## 1 Event truth source

- [x] 1.1 `shared/events.yaml` defines the current 14 event inventory.
- [x] 1.2 Event domains are limited to `target`, `practice`, `report`, `resume`, `source`, and `privacy`.
- [x] 1.3 Event payload fields use B1 enum refs where available and event-local enums where B1 ownership does not apply.
- [x] 1.4 `scripts/lint/events_inventory.py` validates event count, names, producers, payload inventory, domain set and enum boundaries.

## 2 Job truth source

- [x] 2.1 `shared/jobs.yaml` defines the current 8 canonical job_type values and Asynq dotted task names.
- [x] 2.2 API-facing job_type subset is fixed to 6 values; `source_refresh` and `email_dispatch` are internal-only.
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

- [x] 6.1 `event-and-outbox-contract/spec.md`, `plan.md`, `checklist.md`, `context.yaml` and plans INDEX align to the current 14-event / 8-job / 6 API-facing event-job contract.
  <!-- verified: 2026-07-07 method=current-owner-compression evidence="Updated event-and-outbox-contract spec.md to v2.12, plan.md to v1.10, checklist.md to v1.9, context specVersion to v2.12, and synced docs/spec plus event-and-outbox plans INDEX. PASS: targeted stale-wording grep returned no matches; validate_context.py event-and-outbox-contract/001 backend PASS; python3 scripts/lint/events_inventory.py shared/events.yaml shared/jobs.yaml shared/conventions.yaml PASS; make codegen-events PASS; make lint-events PASS; go test ./backend/cmd/codegen/events ./backend/internal/shared/events ./backend/internal/shared/jobs -count=1 PASS; pnpm --dir frontend test src/lib/events src/lib/jobs PASS (11 tests); make codegen-check PASS." -->
