# Backend Resume Tailor Runs and Save v1

> **版本**: 1.13
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本 plan 承接 [backend-resume spec](../../spec.md) 中 flat Resume save 与 resume.tailor backend owner 合同：

- `updateResume`：`PATCH /api/v1/resumes/{resumeId}` 覆盖当前 resume 的 `structured_profile` / `display_name`，使用 Idempotency-Key，cross-user 返回 404。
- `duplicateResume`：`POST /api/v1/resumes/{resumeId}/duplicate` 复制只读来源快照并应用可编辑 overlay，保存为新的独立 resume。
- `requestResumeTailor` / `getResumeTailorRun`：以 `resumeId` 为输入创建和读取 resume.tailor async job；改写建议作为 task output 返回，不持久化逐条采纳状态。
- `resume.tailor.completed`：仅 ready 成功路径写 outbox，payload 只包含 ID / mode / status 等 envelope 字段，不包含 prompt、raw resume、match summary 或 suggested bullet 文本。
- 当前 Resume API、fixtures、generated artifacts、runtime route、BDD 场景和 privacy negative gate 保持一致。

不创建第二套简历对象模型，不恢复分叉、继承或逐条建议状态机；用户采纳后的落盘路径只有 `updateResume` 和 `duplicateResume`。

## 2 当前合同

### 2.1 Operation Matrix

| operationId / surface | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-----------------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getResume` / `listResumes` | current Resume fixtures | Resume Workshop list/detail and selectors | backend-resume get/list handlers | `resumes` | none | 当前无真实 E2E owner；root `make test` |
| `updateResume` | `updateResume.json` | flat save consumers | backend-resume update handler | selected `resumes` row | none | 当前无真实 E2E owner；root `make test` |
| `duplicateResume` | `duplicateResume.json` | duplicate consumers | backend-resume duplicate handler | new `resumes` row | none | 当前无真实 E2E owner；root `make test` |
| `requestResumeTailor` | `requestResumeTailor.json` | tailor request | backend-resume tailor handler | `async_jobs` / `ai_task_runs` | resume-tailor profiles | 当前无真实 E2E owner；root `make test` |
| `getResumeTailorRun` | `getResumeTailorRun.json` | tailor polling | backend-resume tailor read handler | async job result | none at read time | 当前无真实 E2E owner；root `make test` |
| `resume.tailor.completed` event | N/A | downstream consumers | tailor job/store | outbox + task runs | AIClient | 当前无真实 E2E owner；root `make test` |
| removed route family boundary | N/A | current flat routes | route-catalog code contract | none | none | code-level negative gate；root `make test` |

### 2.2 Persistence Contract

- `resumes` is the only current Resume business table.
- `requestResumeTailor` writes `async_jobs` and `ai_task_runs`; returned suggestions live in task output.
- `updateResume` mutates the selected resume in place.
- `duplicateResume` creates a separate resume and leaves the source unchanged.
- Privacy delete removes `resumes` through backend-resume and file binaries through backend-upload; logs, audit events and outbox payloads must not serialize resume body or suggestion text.

### 2.3 Runtime Boundary

- All handlers implement generated OpenAPI server interfaces.
- Session and IK middleware match the B2 contract.
- `cmd/api` owns in-process resume.parse and resume.tailor runner kernel wiring; this plan does not introduce another worker binary.
- Generated route catalog must expose only the current 10 Resume / ResumeTailor operations.

## 3 质量门禁

- **Plan 类型**: `contract + code-internal + feature-behavior`。
- **TDD 策略**: 适用。Focused gates cover generated route catalog, handler unit tests, store tests, `cmd/api` route/runner kernel tests, OpenAPI fixture parity and privacy assertions.
- **替代验证 gate**:
  - `make lint-openapi`
  - `make validate-fixtures`
  - `make codegen-check`
  - `go test ./backend/internal/resume/... ./backend/cmd/api -count=1`
  - `go test ./backend/cmd/api -run 'TestResumeVersionRoutesRemainUnmountedPerD20|TestGeneratedRouteCatalogHasNoResumeVersionOperations' -count=1`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-resume/plans/002-tailor-runs-and-save-v1/context.yaml --target backend`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

## 4 实施步骤

### Phase 0: current contract preflight

- Read `docs/development.md` §2, backend/openapi/scenario READMEs, [backend-resume spec](../../spec.md), B2 OpenAPI inventory, fixtures, generated artifacts and current handler/store/job code.
- Confirm B2 exposes 10 current Resume / ResumeTailor operationIds and 37 total OpenAPI operations.
- Confirm backend-resume context points at current handler/store/job packages, fixtures and scenario directories.

### Phase 1: flat API and removed-route boundary

- Verify `getResume` / `listResumes` fixture parity, cursor stability and cross-user 404 behavior.
- Verify removed route family runtime inputs return 404 and generated route catalog excludes removed operationIds.
- Keep boundary checks in executable code tests and the Given/When/Then BDD contract；当前没有真实 API/UI E2E owner，阶段单测完成由仓库根 `make test` 承接。

### Phase 2: `updateResume`

- Implement and verify `PATCH /api/v1/resumes/{resumeId}` with IK, server-owned field validation, structured profile overwrite, display name update, cross-user isolation and fixture parity.
- Ensure replay uses the existing idempotency middleware contract.

### Phase 3: `duplicateResume`

- Implement and verify `POST /api/v1/resumes/{resumeId}/duplicate` with IK, source snapshot copy, editable overlay, rollback safety, cross-user isolation and fixture parity.

### Phase 4: `requestResumeTailor` / `getResumeTailorRun`

- Implement and verify tailor request dispatch with `payload.resumeId`, status readback, queued/generating/ready/failed response variants and fixture parity.
- Ensure suggestions are read from task output and are not persisted as per-suggestion decision rows.

### Phase 5: resume.tailor async job and outbox

- Implement and verify A3/F3 routing, timeout/output-invalid failure handling, retry-to-ready, typed `ai_task_runs`, ready-only outbox write and payload privacy allowlist.
- Ensure `cmd/api` runner kernel lifecycle owns `resume_tailor` alongside the existing runtime composition.

### Phase 6: BDD and closeout

- Run context validation, OpenAPI/fixture/codegen checks, docs/index checks and diff whitespace checks.
- Keep product-scope owner evidence tied to current route/API/runtime truth sources.

### Phase 7: tailor provenance conversion simplification

- Keep the private persisted `tailorResultProvenance` wire boundary and the exported `VersionProvenance` domain type field-identical.
- Replace both write and readback field-by-field copies with explicit type conversions.
- Run store/package tests plus scoped and backend-wide `staticcheck` before restoring completed state.

### Phase 8: tailor code-test negative-gate precision

- Tailor negative contracts remain in code-owner tests；不再维护 shell verify wrapper，focused run 只作开发反馈，阶段完成由仓库根 `make test` 承接。

### Phase 9: current OpenAPI inventory wording

- checklist preflight 使用当前 37-operation B2 inventory；Resume / ResumeTailor operation 子集仍为 10 个。
- 运行当前 OpenAPI inventory、fixture、owner context 与 docs gates；不修改 Resume handler、store、runner、fixture 或 schema。

### Phase 10: flat Resume mutation handler pipeline consolidation

- 保留 `UpdateResumeService` / `DuplicateResumeService` capability、各自 request parser 和 200 / 201 success status。
- 将两个 handler 重复的 user resolution、Idempotency-Key、body read、validation/error mapping、resource header 和 JSON response pipeline 收敛为一个私有 generic helper。

### Phase 11: shared Resume mode negative gate

- `scenario_script_contract_test.py` 直接校验共享实现语义、consumer inventory 和调用方 zero-inline-regex，随后串行重跑六场景。

### Phase 12: unified Resume runtime negative gate

- 将同一批 consumer 仍复制的 `mistakes|growth|drill|inline-debrief-record` production search 合并进共享 gate。
- 把 `resume-mode-negative-gate.sh` 原地更名为涵盖两类边界的 `resume-runtime-negative-gate.sh`；不保留旧名称 wrapper，也不新增第二个 helper。

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | Current Resume route catalog has 10 operationIds and no removed route family | `make lint-openapi`; `TestGeneratedRouteCatalogHasNoResumeVersionOperations` |
| A-5 | Docs, INDEX and plan context resolve without drift | context validation + `sync-doc-index --check` + `make docs-check` |
| A-6 | Tailor provenance JSON round-trips all current fields without duplicated mapping code | store tests + scoped `staticcheck` |

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-10 | 1.13 | Unify Resume mode and module vocabulary scans in one shared runtime gate. |
| 2026-07-10 | 1.12 | Centralize the contextual Resume mode negative gate across six scenarios. |
| 2026-07-10 | 1.11 | Consolidate the duplicate flat Resume update/duplicate HTTP mutation pipeline. |
| 2026-07-10 | 1.10 | Align the preflight inventory wording with the current 37-operation OpenAPI contract. |
| 2026-07-10 | 1.9 | Narrow tailor mode scenario negative gates to contextual production matches and exclude legal HTTP inline disposition. |
| 2026-07-10 | 1.8 | Run resume tailor scenarios through runner.Runtime and update canonical handler/runtime ownership wording. |
| 2026-07-10 | 1.7 | Replace duplicated tailor provenance write/readback mappings with equivalent type conversions. |
| 2026-07-10 | 1.6 | Align tailor privacy negative gate wording to out-of-scope terminology without behavior changes. |
| 2026-07-07 | 1.5 | Compress owner plan to the current flat Resume save and resume.tailor contract; rename package to current owner wording. |
| 2026-06-14 | 1.4 | Complete flat Resume save / tailor acceptance after D-20 simplification. |
