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
- `cmd/api` owns in-process resume.parse and resume.tailor runner kernel wiring; this plan does not introduce another worker binary.
- Generated route catalog must expose only the current 10 Resume / ResumeTailor operations.

## 3 质量门禁

- **Plan 类型**: `contract + code-internal + feature-behavior`。
- **TDD 策略**: 适用。Focused gates cover generated route catalog, handler unit tests, store tests, `cmd/api` route/runner kernel tests, OpenAPI fixture parity and privacy assertions.
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
- Confirm B2 exposes 10 current Resume / ResumeTailor operationIds and 37 total OpenAPI operations.
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
- Ensure `cmd/api` runner kernel lifecycle owns `resume_tailor` alongside the existing runtime composition.

### Phase 6: BDD and closeout

- Run P0.074-P0.080 scenario gates.
- Run context validation, OpenAPI/fixture/codegen checks, docs/index checks and diff whitespace checks.
- Keep product-scope owner evidence tied to current route/API/runtime truth sources.

### Phase 7: tailor provenance conversion simplification

- Keep the private persisted `tailorResultProvenance` wire boundary and the exported `VersionProvenance` domain type field-identical.
- Replace both write and readback field-by-field copies with explicit type conversions.
- Run store/package tests plus scoped and backend-wide `staticcheck` before restoring completed state.

### Phase 8: tailor scenario negative-gate precision

- P0.075-P0.080 的 Resume mode 负向搜索必须要求 `tailor|mode` 与 `inline|rewrite|mirror` 同行出现，并排除 `*_test.go`；合法 HTTP `Content-Disposition: inline` 不得触发。
- `scenario_script_contract_test.py` 固化六份 verify script 的同构正则和 test-file exclusion；行为相关场景按变更批次串行重跑。

### Phase 9: current OpenAPI inventory wording

- checklist preflight 使用当前 37-operation B2 inventory；Resume / ResumeTailor operation 子集仍为 10 个。
- 运行当前 OpenAPI inventory、fixture、owner context 与 docs gates；不修改 Resume handler、store、runner、fixture 或 schema。

### Phase 10: flat Resume mutation handler pipeline consolidation

- 保留 `UpdateResumeService` / `DuplicateResumeService` capability、各自 request parser 和 200 / 201 success status。
- 将两个 handler 重复的 user resolution、Idempotency-Key、body read、validation/error mapping、resource header 和 JSON response pipeline 收敛为一个私有 generic helper。
- 串行重跑 P0.075 / P0.076，证明 update / duplicate 的 IK、validation、cross-user、fixture 和 persistence 行为不变。

### Phase 11: shared Resume mode negative gate

- 将 P0.075-P0.080 verify 与 P0.080 trigger 重复的 contextual Resume mode 搜索收敛为一个 `_shared` executable gate。
- 六份 verify 与 P0.080 trigger 只调用共享 gate，不再复制正则、glob exclusion 或错误消息。
- `scenario_script_contract_test.py` 直接校验共享实现语义、consumer inventory 和调用方 zero-inline-regex，随后串行重跑六场景。

### Phase 12: unified Resume runtime negative gate

- 将同一批 consumer 仍复制的 `mistakes|growth|drill|inline-debrief-record` production search 合并进共享 gate。
- 把 `resume-mode-negative-gate.sh` 原地更名为涵盖两类边界的 `resume-runtime-negative-gate.sh`；不保留旧名称 wrapper，也不新增第二个 helper。
- 共享 helper 使用一个 scan function 统一 match、clean 和 `rg` error 语义；P0.075/P0.076 同步排除 `*_test.go`。

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | Current Resume route catalog has 10 operationIds and no removed route family | `make lint-openapi`; `TestGeneratedRouteCatalogHasNoResumeVersionOperations` |
| A-2 | Flat read/update/duplicate APIs match fixtures and enforce IK/cross-user rules | handler/service/store tests + `cmd/api` scenarios + P0.074-P0.076 |
| A-3 | Tailor request/read/job flow uses `resumeId`, `async_jobs`, typed `ai_task_runs` and task output suggestions | tailor handler/store/job/runner kernel tests + P0.077-P0.078 |
| A-4 | `resume.tailor.completed` is ready-only and privacy-safe | outbox tests + P0.080 |
| A-5 | Docs, INDEX and plan context resolve without drift | context validation + `sync-doc-index --check` + `make docs-check` |
| A-6 | Tailor provenance JSON round-trips all current fields without duplicated mapping code | store tests + scoped `staticcheck` |
| A-7 | Resume mode negative gates reject only contextual mode vocabulary and ignore legal HTTP inline disposition | scenario script contract + P0.075-P0.080 verify |
| A-8 | Flat update / duplicate handlers share one mutation pipeline while preserving operation-specific validation, service capability and success status | scoped `dupl` + handler tests + P0.075-P0.076 |
| A-9 | P0.075-P0.080 use one shared contextual Resume mode negative gate with no caller-owned regex copies | scenario script contract + `bash -n` + P0.075-P0.080 |
| A-10 | One shared Resume runtime gate owns both mode and module boundary scans; callers and the old helper name contain zero copies | scenario script contract + source negative search + P0.075-P0.080 |

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
