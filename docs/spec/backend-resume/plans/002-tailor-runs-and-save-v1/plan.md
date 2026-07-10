# Backend Resume Tailor Runs and Save v1

> **版本**: 1.6
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
| `getResume` / `listResumes` | `openapi/fixtures/Resumes/getResume.json`, `listResumes.json` | Resume Workshop list/detail, workspace resume picker | `backend/internal/resume/handler/get.go`, `list.go` + `cmd/api` flat resume routes | `resumes` | none | `E2E.P0.074` |
| `updateResume` | `openapi/fixtures/Resumes/updateResume.json` | backend flat save fixture parity / future internal overwrite consumers | `backend/internal/resume/handler/update.go` + `PATCH /api/v1/resumes/{resumeId}` with IK middleware | `resumes.structured_profile`, `resumes.display_name` | none | `E2E.P0.075`, `E2E.P0.079` |
| `duplicateResume` | `openapi/fixtures/Resumes/duplicateResume.json` | backend flat save fixture parity / future internal duplicate consumers | `backend/internal/resume/handler/duplicate.go` + `POST /api/v1/resumes/{resumeId}/duplicate` with IK middleware | new `resumes` row copied from source with editable overlay | none | `E2E.P0.076`, `E2E.P0.079` |
| `requestResumeTailor` | `openapi/fixtures/ResumeTailor/requestResumeTailor.json` | Resume Workshop tailor run request | `backend/internal/resume/handler/request_tailor.go` + `POST /api/v1/resume/tailor` with IK middleware | `async_jobs(job_type='resume_tailor', resource_type='resume_tailor_run', payload.resumeId)` | async job calls F3 `resume.tailor.gap_review` / `resume.tailor.bullet_suggestions` | `E2E.P0.077` |
| `getResumeTailorRun` | `openapi/fixtures/ResumeTailor/getResumeTailorRun.json` | Resume Workshop tailor polling | `backend/internal/resume/handler/get_tailor_run.go` + `GET /api/v1/resume/tailor-runs/{tailorRunId}` | reads `async_jobs` status/result scoped through `resumes` and `payload.resumeId`; suggestions are task output | none at read time | `E2E.P0.077`, `E2E.P0.078` |
| `resume.tailor.completed` event | N/A | downstream event consumers | `backend/internal/resume/jobs/tailor.go`, `backend/internal/resume/store/tailor_runs.go` | `outbox_events` + typed `ai_task_runs` | A3 AIClient via F3 feature_key | `E2E.P0.077`, `E2E.P0.078`, `E2E.P0.080` |
| removed route family boundary | N/A | current frontend uses flat resume routes | `backend/cmd/api/resume_versions_out_of_scope_test.go` asserts route catalog and runtime absence | no current table dependency | none | `E2E.P0.074`, `E2E.P0.079`, `E2E.P0.080` |

### 2.2 Persistence Contract

- `resumes` is the only current Resume business table.
- `requestResumeTailor` writes `async_jobs` and `ai_task_runs`; returned suggestions live in task output.
- `updateResume` mutates the selected resume in place.
- `duplicateResume` creates a separate resume and leaves the source unchanged.
- Privacy delete removes `resumes` through backend-resume and file binaries through backend-upload; logs, audit events and outbox payloads must not serialize resume body or suggestion text.

### 2.3 Runtime Boundary

- All handlers implement generated OpenAPI server interfaces.
- Session and IK middleware match the B2 contract.
- `cmd/api` owns in-process resume.parse and resume.tailor drainer wiring; this plan does not introduce another worker binary.
- Generated route catalog must expose only the current 10 Resume / ResumeTailor operations.

## 3 质量门禁

- **Plan 类型**: `contract + code-internal + feature-behavior`。
- **TDD 策略**: 适用。Focused gates cover generated route catalog, handler unit tests, store tests, `cmd/api` route/drainer tests, OpenAPI fixture parity and privacy assertions.
- **BDD 策略**: 适用。`E2E.P0.074` - `E2E.P0.080` cover flat reads, update, duplicate, tailor happy/failure paths, flat save fixture parity, read-only detail boundary and privacy/boundary negatives.
- **替代验证 gate**:
  - `make lint-openapi`
  - `make validate-fixtures`
  - `make codegen-check`
  - `go test ./backend/internal/resume/... ./backend/cmd/api -count=1`
  - `go test ./backend/cmd/api -run 'TestResumeVersionRoutesRemainUnmountedPerD20|TestGeneratedRouteCatalogHasNoResumeVersionOperations' -count=1`
  - P0.074-P0.080 scenario `setup -> trigger -> verify -> cleanup`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-resume/plans/002-tailor-runs-and-save-v1/context.yaml --target backend`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

## 4 实施步骤

### Phase 0: current contract preflight

- Read `docs/development.md` §2, backend/openapi/scenario READMEs, [backend-resume spec](../../spec.md), B2 OpenAPI inventory, fixtures, generated artifacts and current handler/store/job code.
- Confirm B2 exposes 10 current Resume / ResumeTailor operationIds and 36 total OpenAPI operations.
- Confirm backend-resume context points at current handler/store/job packages, fixtures and scenario directories.

### Phase 1: flat API and removed-route boundary

- Verify `getResume` / `listResumes` fixture parity, cursor stability and cross-user 404 behavior.
- Verify removed route family runtime inputs return 404 and generated route catalog excludes removed operationIds.
- Keep boundary checks in executable tests and BDD scenario scripts rather than prose-only assertions.

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
- Ensure `cmd/api` drainer lifecycle owns `resume_tailor` alongside the existing runtime composition.

### Phase 6: BDD and closeout

- Run P0.074-P0.080 scenario gates.
- Run context validation, OpenAPI/fixture/codegen checks, docs/index checks and diff whitespace checks.
- Keep product-scope owner evidence tied to current route/API/runtime truth sources.

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | Current Resume route catalog has 10 operationIds and no removed route family | `make lint-openapi`; `TestGeneratedRouteCatalogHasNoResumeVersionOperations` |
| A-2 | Flat read/update/duplicate APIs match fixtures and enforce IK/cross-user rules | handler/service/store tests + `cmd/api` scenarios + P0.074-P0.076 |
| A-3 | Tailor request/read/job flow uses `resumeId`, `async_jobs`, typed `ai_task_runs` and task output suggestions | tailor handler/store/job/drainer tests + P0.077-P0.078 |
| A-4 | `resume.tailor.completed` is ready-only and privacy-safe | outbox tests + P0.080 |
| A-5 | Docs, INDEX and plan context resolve without drift | context validation + `sync-doc-index --check` + `make docs-check` |

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-10 | 1.6 | Align tailor privacy negative gate wording to out-of-scope terminology without behavior changes. |
| 2026-07-07 | 1.5 | Compress owner plan to the current flat Resume save and resume.tailor contract; rename package to current owner wording. |
| 2026-06-14 | 1.4 | Complete flat Resume save / tailor acceptance after D-20 simplification. |
