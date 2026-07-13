# Event and Outbox Contract Bootstrap

> **版本**: 1.14
> **状态**: active
> **更新日期**: 2026-07-13

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

维护 B3 当前 event/job/outbox contract：`shared/events.yaml` 定义 12 个 internal events，`shared/jobs.yaml` 定义 7 个 canonical job_type 与 6 个 B2 API-facing job_type subset，`backend/cmd/codegen/events` 生成 Go / TS / JSON Schema / baseline artifacts，`make lint-events` 和 focused tests 负责 drift 与 boundary gate。

本 plan 不实现 backend internal runner、dispatcher loop、业务 producer/consumer、DB migration、PostHog analytics 或 OTel SDK。它只拥有 event/job truth source、generated contract、baseline、lint/codegen gate 和下游 handoff contract。

## 2 当前合同

- `shared/events.yaml` 是 12 个 eventName、envelope、payload schema、producer、aggregateType、PII boundary 和 event-local enum 的真理源。
- 当前 event domains 固定为 `target` / `practice` / `report` / `resume` / `privacy`。
- `shared/jobs.yaml` 是 7 个 canonical job_type、Asynq dotted task name、`triggerEventSemantic`、priority、owner domain、API-facing subset 和 `email_dispatch` redaction policy 的真理源。
- API-facing job_type subset 固定为 `target_import` / `resume_parse` / `report_generate` / `resume_tailor` / `privacy_export` / `privacy_delete`。`email_dispatch` 是唯一 internal-only job。
- `target.import.requested` payload 只保留 `targetJobId` / `userId` / `targetLanguage`；`source.refreshed` 与 refresh job/task contract 不再属于 current inventory。独立 `source_records` persistence 保留，不由本 plan 删除。
- `email_dispatch` payload 只允许审计字段；raw auth token、auth URL、邮箱明文和邮件正文不得进入 `async_jobs.payload`、outbox 或 log。
- Generated artifacts include backend event/job constants and payload structs, frontend event/job constants and payload types, JSON Schemas under `shared/events/schemas`, refs under `shared/events/refs`, and baselines under `shared/events/baseline` / `shared/jobs/baseline`.

## 3 质量门禁分类

- **Plan 类型**: `tooling + code-internal + contract`
- **TDD 策略**: event inventory, jobs inventory, generator output, lint-events, Go/TS generated packages, schema baseline, breaking/additive fixtures and email_dispatch redaction all have focused tests. Re-entry starts from the failing gate and applies the smallest truth-source or generator fix.
- **BDD 策略**: BDD 不适用。本 plan 不产生浏览器 UI、外部 API 行为、用户业务流程或场景测试可观察行为。
- **替代验证 gate**: `python3 scripts/lint/events_inventory.py shared/events.yaml shared/jobs.yaml shared/conventions.yaml`、`make codegen-events`、`make lint-events`、`go test ./backend/cmd/codegen/events ./backend/internal/shared/events ./backend/internal/shared/jobs -count=1`、`pnpm --dir frontend test src/lib/events src/lib/jobs`、`make codegen-check`、`validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check`。

## 4 交付范围

### 4.1 Event truth source

`shared/events.yaml` holds the current event inventory and payload fields. Inventory lint verifies event count, domain set, dot.case names, producer values, payload required fields, event-local enums and B1 enum references.

### 4.2 Job truth source

`shared/jobs.yaml` holds canonical job_type values, Asynq dotted task names, API-facing subset and `triggerEventSemantic`. Inventory lint verifies the 7-job set, 6 API-facing subset, dotted mapping and `email_dispatch` redaction schema.

### 4.3 Codegen and baselines

`make codegen-events` renders Go, TS, JSON Schema and baseline artifacts from the two yaml truth sources. Generated output is deterministic; missing generated contract files, direct edits or baseline drift are caught by `make lint-events` and `make codegen-check`.

### 4.4 Contract tests

Focused tests cover envelope round-trip, traceId soft-required behavior, breaking-change fixtures, additive optional fields, email_dispatch redaction, B1 enum ref generation, generated file presence and current inventory.

### 4.5 Downstream handoff

Downstream owners consume this contract:

- B4 owns `outbox_events` / `async_jobs` DDL and indexes.
- `backend-async-runner` owns dispatcher polling, retry, at-least-once publish and handler execution.
- F1 owns outbox metrics and trace visualization.
- F2 owns product analytics namespace and must not reuse B3 internal event names.
- C1 owns email-code dispatch producer behavior and must use generated redaction helpers.

### 4.6 Canonical generator entrypoint

`backend/cmd/codegen/events/main.go` 与 generator tests 统一调用显式接收 conventions 输入的 `RunWithConventions` / `RunFromBytesWithConventions`。删除仅由测试调用、隐式推导 conventions 路径的 `Run` / `RunFromBytes` wrappers；测试 helper 只负责组织真实入口参数，不在生产包保留平行入口。

## 5 验收标准

- Current event inventory is exactly 12 events across 5 domains.
- Current job inventory is exactly 7 canonical job_type values with 6 API-facing values.
- `make codegen-events` is deterministic and generated artifacts match yaml truth sources.
- `make lint-events` rejects bare event/job literals, missing generated files, breaking payload changes and email_dispatch redaction violations.
- Go and TS shared event/job tests pass.
- Docs/context/index gates pass and active docs describe the current contract.

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| Business code hardcodes event or job strings | `make lint-events` scans backend/frontend source outside generated packages |
| YAML truth source drifts from generated outputs | `make codegen-events` plus generated diff checks enforce determinism |
| API-facing job subset expands silently | inventory lint verifies subset length and `apiFacing` flags |
| email_dispatch carries sensitive content | generated helpers and lint reject redacted fields |
| event-local enum diverges from B1 enum ownership | inventory lint checks event-local enum names against B1 enum names |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-13 | 1.14 | Reopen Phase 9 for OPENAPI-002 paste-only event/job contraction, regenerated baselines and zero-reference gates. | event-and-outbox-contract 2.15 + backend-async-runner 1.16 |
| 2026-07-12 | 1.13 | 对齐 Practice 连续会话合同，删除 turn-level event 与题目/模式计数字段，事件全集收敛为 13。 | backend-practice/001 Phase 1 |
| 2026-07-10 | 1.12 | 对齐事件 inventory contract tests 与当前 14-event / 8-job 数量口径。 | tech-debt pruning |
| 2026-07-10 | 1.11 | 删除两个测试专用 generator wrapper，测试与 CLI 统一使用显式 conventions 入口。 | tech-debt pruning |
| 2026-07-07 | 1.10 | 压缩 owner 文档为当前 14-event / 8-job / 6 API-facing subset event-job contract。 | product-scope/001-core-loop-module-pruning |
| 2026-07-06 | 1.9 | 对齐当前 14 events / 8 canonical jobs / 6 API-facing JobType subset。 | product-scope/001-core-loop-module-pruning |
| 2026-04-30 | 1.3 | 完成 B3 event/job contract bootstrap implementation。 | implementation close-out |

## Phase 9: OPENAPI-002 paste-only event/job contraction

### 9.1 RED contract inventory

Add focused inventory/generator tests that fail against the pre-contraction truth source. RED must require `target.import.requested.requiredPayload` to omit `sourceType`, reject event-local `TargetImportSourceType` / `SourceFreshnessStatus`, reject `source.refreshed`, and reject the refresh canonical job/dotted task/priority mapping. Expected inventory is exactly 12 events / 5 domains / 7 canonical jobs / 6 API-facing jobs. Tests must also assert `source_records` is not a deletion target of this contract phase.

### 9.2 GREEN truth source, generated artifacts and baselines

Update `shared/events.yaml` and `shared/jobs.yaml`, then regenerate backend/frontend event/job types, JSON Schemas, refs and committed v1 baseline manifests. `target.import.requested` retains required `targetJobId` / `userId` / `targetLanguage` so the runner can read the persisted paste text by ID; raw JD text does not enter outbox payload. Delete generated refresh event/job symbols and schema files without compatibility aliases or deprecated wrappers.

Because the project is pre-launch, this accepted contraction re-freezes the current v1 manifests after preserving a before/after inventory audit. A baseline-only edit without proposed truth-source/generated diffs must fail.

### 9.3 Handoff and zero-reference

Hand regenerated job types to backend-async-runner Phase 7, B4 job constraint owner and TargetJob producer/consumer owners in the same implementation batch. Current truth source/generated/schema/baseline/runtime/producer/consumer surfaces must contain zero positive source discriminator enum, refresh event, refresh job, dotted refresh task or queue assignment; historical docs and explicit negative-test tokens may be excluded. `source_records` table/model/query ownership remains independently intact and must not be caught by this zero-reference gate.

### 9.4 Verification strategy

BDD is not applicable: this plan changes an internal contract and codegen surface, not a new user workflow. Replacement gates are RED/GREEN inventory tests, generator tests, `make codegen-events`, `make lint-events`, Go/TS generated-package tests, baseline drift checks, backend runner focused tests and scoped zero-reference searches.
