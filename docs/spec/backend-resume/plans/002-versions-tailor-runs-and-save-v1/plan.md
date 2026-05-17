# Backend Resume Versions, Tailor Runs and Save v1

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-17

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [backend-resume spec](../../spec.md) §6 C-9（partial）/ C-10 / C-11 / C-13 / C-14 / C-15 / C-16 落到 backend Go handler + store + AI 编排 + outbox：

- 实施 D-10 `confirmResumeStructuredMaster` 新 operationId 的 B2 D-18 additive cross-owner change：在 `openapi/openapi.yaml` 增加 `POST /api/v1/resumes/{resumeAssetId}/structured-master` + `ConfirmResumeStructuredMasterRequest` schema，新增 fixture（`default` / `idempotency-replay` / `already-exists-409` / `validation-422`），同步 OpenAPI inventory lint、Go/TS generated server/client artifacts、B2 spec §3.1.1 行与 B1 错误码 `RESUME_STRUCTURED_MASTER_ALREADY_EXISTS`；
- 携带 B4 cross-owner addendum migration `000007_resume_versions_structured_master_unique`：partial UNIQUE INDEX `resume_versions(resume_asset_id) WHERE version_type='structured_master' AND deleted_at IS NULL`，在 db-migrations-baseline history.md 写入跨 owner 行；
- 实现 `confirmResumeStructuredMaster` handler + store（`resume_versions.CreateStructuredMasterFromAsset`）：handler 内 SELECT FOR UPDATE + partial UNIQUE INDEX 双层兜底，IK 必带，cross-user 404，覆盖 spec C-14 / C-15；
- 实现 `getResumeVersion` / `listResumeVersions` handler + store：cursor pagination（按 `updated_at DESC, id DESC` 唯一稳定序）、cross-user 404、asset ownership check；
- 实现 `updateResumeVersion` handler + store：IK + partial structured_profile merge + 不可编辑字段（`versionType` / `parentVersionId` / `resumeAssetId`）拒绝；
- 实现 `branchResumeVersion` 三路 seed_strategy（D-2）：`copy_master`（同步 201 + structured_profile 拷贝）/ `blank`（同步 201 + structured_profile 空）/ `ai_select`（同步 201 resume_version 行 + 入队 `resume_tailor` job → 202 + `BranchResumeVersionAccepted`）；parent existence + cross-user + IK；
- 实现 `requestResumeTailor` + `getResumeTailorRun` handler + store：202 + Job(queued) + queued → generating → ready/failed 状态机；mode ∈ `gap_review | bullet_suggestions`（D-5）；IK；
- 实现 `resume.tailor` async job handler + AIClient（A3）+ F3 `resume.tailor.gap_review` / `resume.tailor.bullet_suggestions` feature_key：按 backend-resume/001 同款 `cmd/api` in-process drainer 注册（不引入独立 worker），ready 成功写 `resume_version_suggestions`（status='pending'）+ outbox `resume.tailor.completed`（ready-only，envelope `tailorRunId / resumeAssetId / targetJobId / mode / status`）；
- 实现 `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` 终态状态机：pending → accepted | rejected；accept 仅写 `decided_at` + `status='accepted'`，不自动改 `resume_versions.structured_profile`（D-12）；IK 必带；二次 accept 走 idempotency middleware replay；不同 fingerprint 同 key 409；cross-user 404；
- 在 `cmd/api` 挂载 9 个新 route，验证 session middleware、IK middleware、in-process resume_tailor drainer wiring 都走真实 runtime；
- 通过 spec §6 C-9 partial / C-10 / C-11 / C-13 / C-14 / C-15 / C-16 验收 + 新增 7 个 BDD 场景 `E2E.P0.074 – E2E.P0.080`；
- 解除 [frontend-resume-workshop/002](../../../frontend-resume-workshop/) （未来）切真路径阻塞；不实现 `exportResumeVersion` 真实 PDF / `archiveResumeAsset` / 完整 privacy delete 链路（归 plan 003 P1）。

## 2 背景

[backend-resume spec](../../spec.md) 已锁定 14 个 Resume operationId（含 D-10 additive），其中 5 个由 [001-asset-register-parse-and-listing](../001-asset-register-parse-and-listing/plan.md) 落地（`registerResume` / `getResume` / `listResumes` / resume.parse async / resume.parse.completed）。剩余 9 个 op + Preview Confirm 保存 v1 主版本 + resume.tailor async job + `resume.tailor.completed` event 由本 plan 承接，是 Resume Workshop 阶段 2 的核心代码切片：完成后 [frontend-resume-workshop/002](../../../frontend-resume-workshop/) （未来 plan）可以从 mock-first 切到真实 backend，承接 Preview Confirm → save v1 → branch → 改写 run → accept/reject suggestion 的完整用户路径。

每个 phase 是可独立验证的纵向行为切片：Phase 1 起来就有 OpenAPI + fixtures + generated artifacts；Phase 2 起来就有 confirmResumeStructuredMaster handler + DB UNIQUE INDEX；Phase 3 起来就有 getResumeVersion + listResumeVersions；Phase 4 起来就有 updateResumeVersion；Phase 5 起来就有 branchResumeVersion 三路；Phase 6 起来就有 requestResumeTailor + getResumeTailorRun；Phase 7 起来就有 resume.tailor async job + AIClient 集成 + outbox event；Phase 8 起来就有 accept/reject 终态状态机；Phase 9 收口 + BDD + 解除 frontend-resume-workshop/002 阻塞。

执行本 plan 前必须确认：

- [backend-resume/001](../001-asset-register-parse-and-listing/plan.md) completed（registerResume / getResume / listResumes / resume.parse async / `resume.parse.completed` event 已挂在 `cmd/api`；in-process drainer pattern 已成熟）。
- [B2 D-18](../../../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) Phase 1-5 已完成（Resumes + ResumeTailor tag 13 个 op 的 schema、fixtures、inventory lint、generated server/client artifacts 全部就位；本 plan Phase 1 在此基础上 additive 第 14 个 op）。
- [B3 D-14](../../../event-and-outbox-contract/plans/002-resume-tailor-mode-drift-fix/plan.md) Phase 1-4 已完成（`ResumeTailorMode` enum / baseline manifest / generated 类型 / negative grep 与 B3 spec 描述全部对齐）。
- [B4 002 resume-versions-additive](../../../db-migrations-baseline/plans/002-resume-versions-additive/plan.md) 已完成（`resume_versions` / `resume_version_suggestions` / `resume_tailor_runs` schema 已落地；本 plan Phase 2 在此基础上携带 cross-owner addendum migration 增加 partial UNIQUE INDEX）。
- [F3 001 baseline](../../../prompt-rubric-registry/plans/001-baseline/plan.md) 已 ready（`resume.tailor.gap_review` / `resume.tailor.bullet_suggestions` feature_key + prompt / rubric / model profile 就位）；如有缺失，由本 plan Phase 0 preflight 通知 F3 owner 补齐，本 plan 不直接修订 F3。
- [A3 003](../../../ai-provider-and-model-routing/plans/003-provider-registry-and-capability-profiles/plan.md) 已 ready（AIClient + provider registry + Capability Model Profile）。
- [shared-conventions-codified](../../../shared-conventions-codified/spec.md) 错误码 lookup table 接受 `RESUME_STRUCTURED_MASTER_ALREADY_EXISTS` additive；本 plan Phase 1 同步修订 B1 spec §3.1 D-5 / D-10。

## 3 质量门禁分类

- **Plan 类型**: `contract + migration + code-internal + feature-behavior`。本 plan 携带 B2 D-18 additive 契约修订、B4 cross-owner addendum migration、backend handler / store / async job / AI 调用，以及用户可见 HTTP API 行为。
- **TDD 策略**: 适用。Red-Green-Refactor 入口：
  1. OpenAPI lint / inventory / fixture / codegen drift：`make lint-openapi` + `make validate-fixtures` + `make codegen-openapi` + `make codegen-check`；
  2. migration check：`python3 scripts/lint/migrations_lint.py --repo-root .` + `make migrate-check`（partial UNIQUE INDEX + rollback DOWN）；
  3. handler unit test（共 9 op）：参数校验 + IK + 422 / 404 / 409 / cross-user 隔离 + idempotency middleware replay；
  4. store integration test：CRUD + state machine + cross-user + partial UNIQUE INDEX 并发兜底 + cursor pagination 边界；
  5. resume.tailor job unit test（stub AIClient）：成功路径 / 解析 JSON 失败 / AI provider timeout retryable / output_invalid / retry 复用；
  6. outbox event unit test：envelope 字段集 + ready-only + PII 红线（不含 suggested bullet text）；
  7. `cmd/api` route/runtime test：session middleware、IK middleware、9 个 route path params、resume_tailor in-process drainer wiring 与 shutdown；
  8. legacy / privacy / events drift negative：`grep inline|rewrite|mirror` + `grep mistakes|growth|drill` + outbox payload privacy assertion。
  执行入口：`/implement backend-resume/002-versions-tailor-runs-and-save-v1` → `/tdd`。
- **BDD 策略**: 适用（Feature plan requires BDD）。`E2E.P0.074` – `E2E.P0.080` 共 7 个场景，详见 [bdd-plan.md](./bdd-plan.md) / [bdd-checklist.md](./bdd-checklist.md)。
- **替代验证 gate**:
  - `make lint-openapi`
  - `make validate-fixtures`
  - `make codegen-openapi && make codegen-check`
  - `python3 scripts/lint/migrations_lint.py --repo-root .`
  - `make migrate-check`
  - `cd backend && go test ./...`
  - `cd backend && go test ./internal/resume/... -count=1`
  - `cd backend && go test ./internal/resume/store/... -tags=integration -count=1`
  - `cd backend && go test ./cmd/api -run 'TestBuildResumeRuntime|TestResume.*HTTPScenario|TestResumeTailorDrainer.*' -count=1`
  - smoke：通过 `cmd/api` 真实 route 调 9 个新 endpoint 与 fixture 字节比对
  - grep `inline|rewrite|mirror` in `backend/internal/resume/` + resume_tailor drainer/outbox payload tests（C-13 negative）
  - grep `mistakes|growth|drill|inline-debrief-record` in `backend/internal/resume/`（retired module negative；当前 plan / BDD gate prose 与历史 out-of-scope 文档不纳入 raw zero-hit 扫描，避免自匹配）
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

### 3.1 Frontend / Backend Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `confirmResumeStructuredMaster` (NEW D-10) | `openapi/fixtures/Resumes/confirmResumeStructuredMaster.json` `default` / `idempotency-replay` / `already-exists-409` / `validation-422`（本 plan Phase 1 新增） | `frontend-resume-workshop/002`（未来）Preview Confirm 保存 v1 入口 | `backend/internal/resume/handler/confirm_structured_master.go` real handler + `cmd/api` `POST /api/v1/resumes/{resumeAssetId}/structured-master` route with session + IK middleware | `resume_versions(version_type='structured_master', parent_version_id=null)` 单事务 INSERT；resume_asset 不变；partial UNIQUE INDEX 兜底 | none (用户已确认 structured_profile；本 op 不再调 AI) | `E2E.P0.074` |
| `getResumeVersion` | `openapi/fixtures/Resumes/getResumeVersion.json` `default` / `not-found-404` | `frontend-resume-workshop/002` 详情视图 | `backend/internal/resume/handler/get_version.go` + `cmd/api` `GET /api/v1/resume-versions/{resumeVersionId}` | `resume_versions` read | none | `E2E.P0.074` |
| `listResumeVersions` | `openapi/fixtures/Resumes/listResumeVersions.json` `default` / `empty` / `paginated` | `frontend-resume-workshop/002` 版本列表 / `frontend-workspace-and-practice` ResumePicker 已 ready | `backend/internal/resume/handler/list_versions.go` + `cmd/api` `GET /api/v1/resumes/{resumeAssetId}/versions` | `resume_versions` cursor pagination + asset ownership join | none | `E2E.P0.074` |
| `updateResumeVersion` | `openapi/fixtures/Resumes/updateResumeVersion.json` `default` / `validation-error-422`（本 plan 补齐 `idempotency-replay`，并保留 current fixture scenario name） | `frontend-resume-workshop/002` 编辑器 | `backend/internal/resume/handler/update_version.go` + `cmd/api` `PATCH /api/v1/resume-versions/{resumeVersionId}` with IK middleware | `resume_versions` UPDATE structured_profile / displayName / focusAngle / matchScore | none | `E2E.P0.075` |
| `branchResumeVersion` | `openapi/fixtures/Resumes/branchResumeVersion.json` `default` / `copy-master-sync` / `blank-sync` / `ai-select-202-with-job` / `idempotent-replay` / `validation-error-422` | `frontend-resume-workshop/002` "为岗位定制" 入口 | `backend/internal/resume/handler/branch_version.go` + `cmd/api` `POST /api/v1/resume-versions` with IK middleware | `resume_versions` INSERT + 可选 `resume_tailor_runs` + `async_jobs(resume_tailor)` | none in branch handler；ai_select 路径异步触发 resume.tailor job | `E2E.P0.076` (sync) + `E2E.P0.077` (async dispatch) |
| `requestResumeTailor` | `openapi/fixtures/ResumeTailor/requestResumeTailor.json` `default`（本 plan 修复 default request 缺 `Idempotency-Key`，并补齐 `idempotency-replay`） | `frontend-resume-workshop/002` "运行改写 run" 入口 | `backend/internal/resume/handler/request_tailor.go` + `cmd/api` `POST /api/v1/resume/tailor` with IK middleware | `resume_tailor_runs(status='queued')` + `async_jobs(job_type='resume_tailor')` 单事务 | F3 `resume.tailor.gap_review` 或 `resume.tailor.bullet_suggestions` (实际调用归 Phase 7 async job) | `E2E.P0.077` |
| `getResumeTailorRun` | `openapi/fixtures/ResumeTailor/getResumeTailorRun.json` `default`（本 plan 补齐 `queued` / `generating` / `failed` 三个 status variant） | `frontend-resume-workshop/002` 改写 run 轮询 | `backend/internal/resume/handler/get_tailor_run.go` + `cmd/api` `GET /api/v1/resume/tailor-runs/{tailorRunId}` | `resume_tailor_runs` read + cross-user 404 | none | `E2E.P0.077` + `E2E.P0.078` |
| `acceptResumeTailorSuggestion` | `openapi/fixtures/Resumes/acceptResumeTailorSuggestion.json` `default`（本 plan 补齐 `idempotency-replay` / `already-decided-409`，并将 current `conflict-409` + `TARGET_INVALID_STATE_TRANSITION` 漂移收敛为 `VALIDATION_FAILED` + `detail.reason='SUGGESTION_ALREADY_DECIDED'`） | `frontend-resume-workshop/002` accept CTA | `backend/internal/resume/handler/accept_suggestion.go` + `cmd/api` `POST /api/v1/resume-versions/{resumeVersionId}/suggestions/{suggestionId}/accept` with IK middleware | `resume_version_suggestions.status='accepted' + decided_at`（不自动改 structured_profile，D-12） | none | `E2E.P0.079` |
| `rejectResumeTailorSuggestion` | `openapi/fixtures/Resumes/rejectResumeTailorSuggestion.json` `default`（本 plan 补齐 `idempotency-replay` / `already-decided-409`，并将 current `conflict-409` + `TARGET_INVALID_STATE_TRANSITION` 漂移收敛为 `VALIDATION_FAILED` + `detail.reason='SUGGESTION_ALREADY_DECIDED'`） | `frontend-resume-workshop/002` reject CTA | `backend/internal/resume/handler/reject_suggestion.go` + `cmd/api` `POST /api/v1/resume-versions/{resumeVersionId}/suggestions/{suggestionId}/reject` with IK middleware | `resume_version_suggestions.status='rejected' + decided_at` | none | `E2E.P0.079` |

## 4 实施步骤

### Phase 0: preflight + spec/plan 锁定

#### 0.1 spec L1 修订（本 plan 同步落地）
- spec.md 1.1 → 1.2：D-10 / D-11 / D-12 新增；C-14 / C-15 / C-16 新增；§1 op 总数 13 → 14；§3.2 待确认事项与 §7 plan 002 行更新。
- history.md 追加 1.2 行（已完成）。

#### 0.2 F3 feature_key 与 AIClient profile preflight
- 验证 `config/prompts/resume.tailor.gap_review/v*` 与 `config/prompts/resume.tailor.bullet_suggestions/v*` 在 [F3 baseline](../../../prompt-rubric-registry/plans/001-baseline/plan.md) 已 ready；如缺失，通知 F3 owner，由 F3 plan 修订补齐；本 plan 不直接修订 F3。
- 验证 [A3 AIClient](../../../ai-provider-and-model-routing/spec.md) Capability Model Profile `resume.tailor.*` 路由可注入 stub provider（与 resume.parse / 001 同款 stub harness）。

#### 0.3 retired module / events drift baseline
- 在 `backend/internal/resume/` 执行 `grep -nE 'inline|rewrite|mirror'`（C-13）+ `grep -nE 'mistakes|growth|drill|inline-debrief-record'` baseline：0 命中前置，Phase 9 收口时再次验证；当前 plan / BDD 文档中的 gate literal 和历史 out-of-scope 说明不作为 raw zero-hit 扫描对象。

### Phase 1: B2 D-18 additive `confirmResumeStructuredMaster` + B1 错误码增补

#### 1.1 OpenAPI schema + operation
- 在 `openapi/openapi.yaml` 新增 `POST /api/v1/resumes/{resumeAssetId}/structured-master` operation：tag `Resumes`，IK + XRequestID + Traceparent + AcceptLanguage + XClientVersion 参数集合与 D-8 一致；RequestBody `$ref: ConfirmResumeStructuredMasterRequest`；响应 201 `ResumeVersion` + 409 `ApiErrorResponse` + default `ApiErrorResponse`。
- 新增 `ConfirmResumeStructuredMasterRequest` schema：`required: [structuredProfile, displayName]`；`structuredProfile.provenance` `$ref: GenerationProvenance`；可选 `language`。

#### 1.2 fixtures
- 新增 `openapi/fixtures/Resumes/confirmResumeStructuredMaster.json` 四个 scenario：`default`（201 + ResumeVersion structured_master）/ `idempotency-replay`（同 IK 返回首次结果）/ `already-exists-409`（已存在 master → 409 `RESUME_STRUCTURED_MASTER_ALREADY_EXISTS`）/ `validation-422`（schema-valid blank displayName → 422 `VALIDATION_FAILED`；fixture request 必须保持 schema-valid）。

#### 1.3 inventory / lint / codegen
- 在 `scripts/lint/openapi_inventory.py` baseline 注册新 operation，并把 IK_REQUIRED 增补 `POST /resumes/{resumeAssetId}/structured-master`。
- 在 `openapi-v1-contract/spec.md` §3.1.1 endpoint 列表追加 `confirmResumeStructuredMaster`（全局 endpoint count 58 → 59，backend-resume owner surface 13 → 14）；history 追加版本行。
- 跑 `make lint-openapi` + `make validate-fixtures` + `make codegen-openapi` + `make codegen-check`：generated server/client artifacts 含 `ConfirmResumeStructuredMaster` 方法与 client SDK 类型。

#### 1.4 B1 错误码增补
- 在 `docs/spec/shared-conventions-codified/spec.md` §3.1 D-5 / D-10 错误码 lookup table 增补 `RESUME_STRUCTURED_MASTER_ALREADY_EXISTS`（与 `RESUME_EXPORT_NOT_AVAILABLE` 同列）；同步 history。
- 修订 `openapi/openapi.yaml` ApiErrorResponse `error.code` enum 列表加入新错误码。

#### 1.5 dev mock client + frontend mock harness
- `frontend/src/api/devMockClient.ts` 注册新 operation 与 fixture scenarios，避免 dev-mock fixture coverage gate 退化（与 backend-practice/listPracticeSessions 同款补齐）。

### Phase 2: B4 cross-owner addendum migration + `confirmResumeStructuredMaster` handler

#### 2.1 migration 000007
- 新增 `migrations/000007_resume_versions_structured_master_unique.up.sql`：
  ```sql
  CREATE UNIQUE INDEX IF NOT EXISTS uq_resume_versions_structured_master_per_asset
    ON resume_versions (resume_asset_id)
    WHERE version_type = 'structured_master' AND deleted_at IS NULL;
  ```
- 新增 `migrations/000007_resume_versions_structured_master_unique.down.sql`：`DROP INDEX IF EXISTS uq_resume_versions_structured_master_per_asset;`。
- 在 `docs/spec/db-migrations-baseline/history.md` 追加 cross-owner addendum 行（由 backend-resume/002 owner 写入）。
- `python3 scripts/lint/migrations_lint.py --repo-root .` + `make migrate-check` PASS。

#### 2.2 handler `confirm_structured_master.go`
- 实现 generated server interface `ConfirmResumeStructuredMaster`。
- IK 必带（缺失 422，复用 idempotency middleware）。
- cross-user：resume_asset 不属本用户 → 404 不暴露存在。
- 调用 service `ConfirmStructuredMaster(ctx, userID, resumeAssetID, req)`，handler 仅做参数解析 + 错误码映射。
- 返回 201 + `ResumeVersion`（含 `structuredProfile.provenance`）。

#### 2.3 service / store `resume_versions.CreateStructuredMasterFromAsset`
- 单事务：`SELECT ... FOR UPDATE` resume_assets 行 → 检查 ownership 与 `parse_status='ready'`（如非 ready 返回 422 `VALIDATION_FAILED` + detail.reason=`PARSE_NOT_READY`）→ INSERT `resume_versions` 行（`version_type='structured_master'`, `parent_version_id=null`, `target_job_id=null`, `seed_strategy=null`）。
- 命中 partial UNIQUE INDEX 时映射为 `409 + RESUME_STRUCTURED_MASTER_ALREADY_EXISTS`。
- handler 单元测试覆盖 happy / IK replay / 409 / 422 / cross-user / parse_not_ready 各分支。
- store integration test 覆盖：（a）并发 INSERT 仅一行成功；（b）IK replay 不创建新行；（c）cross-user filter。

#### 2.4 `cmd/api` route wiring
- 注册新 route 进 `buildResumeRuntime` 或等价 helper；session + IK middleware 链与 backend-resume/001 一致。
- HTTP scenario test 覆盖：route 存在、缺 session 401、缺 IK 422、IK replay 返回首次结果、并发 409。

BDD-Gate: 验证 `E2E.P0.074` 通过（覆盖 confirmStructuredMaster + getResumeVersion + listResumeVersions 主路径，详见 Phase 3 后联动收口）

### Phase 3: `getResumeVersion` + `listResumeVersions` handler

#### 3.1 handler `get_version.go`
- 实现 generated server interface `GetResumeVersion`。
- cross-user 返回 404。
- 返回 `ResumeVersion` 字段（按 B2 schema）。

#### 3.2 handler `list_versions.go`
- 实现 generated server interface `ListResumeVersions`。
- asset ownership：先校验 resume_asset 属本用户（404 不暴露存在），再查询。
- cursor pagination：`updated_at DESC, id DESC` 唯一稳定序。
- 返回 `PaginatedResumeVersion{items, pageInfo}`。

#### 3.3 store `resume_versions.GetByID` / `ListByAsset(cursor, pageSize)`
- 索引：复用 B4 002 已建立的 `idx_resume_versions_asset_type` + `idx_resume_versions_user_updated`。
- 严格 user-scoped 过滤；cursor invalid → 422。

#### 3.4 unit / integration test
- handler unit：200 / 404 cross-user / 404 not exist / 422 invalid cursor。
- store integration：25 行版本 + 第二页 + `hasMore=false`；cross-user 不可见；按 `updated_at DESC, id DESC` 排序。

#### 3.5 `cmd/api` route wiring
- 挂载 `GET /api/v1/resume-versions/{resumeVersionId}` + `GET /api/v1/resumes/{resumeAssetId}/versions`。
- HTTP scenario test 覆盖 4 个 case：A 拥有 → 200；B 访问 A 的 → 404；asset 不存在 → 404；invalid cursor → 422。

BDD-Gate: 验证 `E2E.P0.074` 通过

### Phase 4: `updateResumeVersion` handler

#### 4.1 handler `update_version.go`
- 实现 generated server interface `UpdateResumeVersion`。
- PATCH semantic：仅修改 `displayName` / `focusAngle` / `structuredProfile` / `matchScore` 字段；不可改 `versionType` / `resumeAssetId` / `parentVersionId` / `targetJobId` / `seedStrategy`（请求体中包含这些字段直接 422）。
- structured_profile partial merge：客户端发送的 partial jsonb 与现有值合并；`provenance` 字段强制由 server 重置（不可由 client 任意写入）。
- IK 必带（D-8）；replay 返回首次结果；mismatch 409。

#### 4.2 store `resume_versions.UpdatePatch`
- 单事务 UPDATE；返回 `ResumeVersion`。
- cross-user 404。
- 已 `deleted_at` 行：404。

#### 4.3 unit / integration test
- handler unit：happy / IK replay / IK mismatch / 422 不可编辑字段 / cross-user 404。
- store integration：merge semantic + concurrent update + ETag-like updated_at 推进。

#### 4.4 `cmd/api` route wiring
- 挂载 `PATCH /api/v1/resume-versions/{resumeVersionId}` 含 IK middleware。
- HTTP scenario test 覆盖 5 个 case：happy / IK replay / IK mismatch / 不可编辑字段 / cross-user。

BDD-Gate: 验证 `E2E.P0.075` 通过

### Phase 5: `branchResumeVersion` 三路 seed_strategy

#### 5.1 handler `branch_version.go`
- 实现 generated server interface `BranchResumeVersion`。
- IK 必带；parent existence + cross-user（parent 不属本用户 → 404）。
- seed_strategy 路由：
  - `copy_master` → 同步 INSERT `resume_versions(version_type='targeted', parent_version_id=parent.id, target_job_id=req.targetJobId, structured_profile=parent.structured_profile copy)`；返回 201 + `ResumeVersion`。
  - `blank` → 同步 INSERT `resume_versions(version_type='targeted', parent_version_id=parent.id, target_job_id=req.targetJobId, structured_profile=empty + provenance.dataSourceVersion='resume_version.v1')`；返回 201 + `ResumeVersion`。
  - `ai_select` → 同事务 INSERT `resume_versions(version_type='targeted', ...)` + `resume_tailor_runs(status='queued', mode='gap_review' 默认)` + `async_jobs(job_type='resume_tailor', resource_type='resume_tailor_run')`；返回 202 + `BranchResumeVersionAccepted{resumeVersionId, version, job}`。
- 不可编辑字段同 D-2；`focusAngle` 可选。

#### 5.2 store `resume_versions.BranchFromParent`
- 单事务三种分支；ai_select 路径必须保证 resume_version + tailor_run + async_jobs 三行原子提交，否则 outbox / drainer 会捕获 orphan。
- target_job_id 存在性校验：targetJobId 不属本用户 → 404。

#### 5.3 unit / integration test
- handler unit：copy_master / blank / ai_select 三路 + IK replay + parent 不存在 404 + targetJob 不存在 404 + 422 invalid seed_strategy。
- store integration：三路 INSERT + ai_select rollback case + cross-user。

#### 5.4 `cmd/api` route wiring
- 挂载 `POST /api/v1/resume-versions` 含 IK middleware。
- HTTP scenario test 覆盖：copy_master 同步 201 + 字节比对 fixture / blank 同步 201 / ai_select 异步 202 + Job(queued) + DB 三行原子。

BDD-Gate: 验证 `E2E.P0.076` 通过（sync）
BDD-Gate: 验证 `E2E.P0.077` 通过（async dispatch handoff Phase 7）

### Phase 6: `requestResumeTailor` + `getResumeTailorRun` handler

#### 6.1 handler `request_tailor.go`
- 实现 generated server interface `RequestResumeTailor`。
- IK 必带；mode ∈ `gap_review | bullet_suggestions`（B3 D-14 / 本 spec D-5）；其他值 422 `VALIDATION_FAILED`。
- 同事务创建 `resume_tailor_runs(status='queued', mode)` + `async_jobs(job_type='resume_tailor')`；返回 202 + `{tailorRunId, job(jobType=resume_tailor, status=queued)}`。
- targetJobId / resumeAssetId 都必须属本用户（404 不暴露存在）。

#### 6.2 handler `get_tailor_run.go`
- 实现 generated server interface `GetResumeTailorRun`。
- cross-user 404；返回 `ResumeTailorRun` 字段（status / matchSummary / suggestions / provenance / timestamps）。
- 处理 status ∈ `queued | generating | ready | failed`（与 `openapi/openapi.yaml` / `migrations/000001_create_baseline.up.sql` / `migrations/enum-sources.yaml` 当前 truth source 一致）。

#### 6.3 store `resume_tailor_runs` Repository
- `Create / Get / MarkGenerating / MarkReady(matchSummary, suggestions) / MarkFailed(errorCode)`。
- state machine：`queued → generating → ready | failed`。

#### 6.4 unit / integration test
- handler unit：request happy / IK replay / mode 422 / cross-user 404 + get happy / cross-user 404 / status 四态。
- store integration：state transition + cross-user + concurrent claim from queued。

#### 6.5 `cmd/api` route wiring
- 挂载 `POST /api/v1/resume/tailor`（含 IK middleware）+ `GET /api/v1/resume/tailor-runs/{tailorRunId}`。
- HTTP scenario test：handler 与真实 route 字节比对 fixture。

BDD-Gate: 验证 `E2E.P0.077` 通过

### Phase 7: resume.tailor async job + AIClient + outbox `resume.tailor.completed`

#### 7.1 实现 `internal/resume/jobs/tailor.go`
- 注册到 `cmd/api` in-process drainer（job_type=resume_tailor, dotted=resume.tailor），复用 backend-resume/001 已建立的 drainer pattern；不引入独立 worker binary、`WORKER_*` config 或 `backend-async-runtime` 旧 shorthand。
- 从 `resume_tailor_runs` 读 `resume_asset_id` / `target_job_id` / `mode` → 通过 [A3 AIClient](../../../ai-provider-and-model-routing/spec.md) 调 [F3 `resume.tailor.gap_review` / `resume.tailor.bullet_suggestions` feature_key](../../../prompt-rubric-registry/spec.md) 之一（按 mode 路由）。
- 解析 LLM JSON 输出 → 写 `resume_tailor_runs.status='ready'` + `match_summary` + `provenance`，并在 `resume_version_suggestions` 写 N 行 `status='pending'`。
- 用户后续 accept/reject 才改 suggestion 状态；本 job 不预设 accepted/rejected。
- 失败路径：`status='failed'` + `error_code`；retryable 信息由 `async_jobs` retry metadata 表达，不向 `resume_tailor_runs.status` 私加 `failed_retryable`。
- 写入 `ai_task_runs` typed columns（model_profile_name / version / prompt_version / rubric_version / route / validation_status / feature_key）。

#### 7.2 outbox event `resume.tailor.completed`
- envelope 字段集（[B3 §3.1.4](../../../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory)）：`tailorRunId / resumeAssetId / targetJobId / mode / status`。
- 仅在最终 `status='ready'` 时写入；AI output invalid / provider timeout / retryable exhausted 等失败路径不发 `resume.tailor.completed`。
- PII 边界：不含 `original_bullet` / `suggested_bullet` / `match_summary.strengths` / `match_summary.gaps` 文本；不含 prompt input 与 model raw response。

#### 7.3 resume_tailor in-process drainer wiring
- 沿用 backend-resume/001 的 in-process drainer 口径：`cmd/api` 进程内 claim `async_jobs(job_type=resume_tailor)` 并调用 `backend/internal/resume/jobs/tailor.go`。
- 提供 `RunOnce` 或等价 deterministic stepping；与 resume_parse drainer 并发协作（不互相 claim 跨类型 job）。
- `Start(ctx)` / `Shutdown(ctx)` 必须随 `cmd/api` lifecycle 管理。

#### 7.4 unit test
- `tailor_test.go`（stub AIClient）：成功 / parse JSON 失败 / AI timeout retryable / output_invalid。
- `drainer_test.go`：`Handles(resume_tailor)`、`RunOnce` 成功处理、shutdown 不泄漏 goroutine、未知 job type 不被本 drainer claim。
- outbox unit test：envelope 字段集 + ready-only + PII 红线。

#### 7.5 Remediation: persisted ready provenance completeness
- `resume_tailor_runs` 必须持久化并可重新读取完整 `GenerationProvenance`：`promptVersion / rubricVersion / modelId / language / featureFlag / dataSourceVersion`。
- `CompleteTailorRunSuccess` 后通过 `GetTailorRun` 读取 ready run 时，`getResumeTailorRun` 返回的 `provenance` 不得只保留 `prompt_version / rubric_version / model_id / provider` typed columns；`language / feature_flag / data_source_version` 也必须来自同一 run 持久化状态。
- store integration 必须覆盖 write-after-read 断言，避免 fixture-only 或 fake store 测试漏掉 DB roundtrip。

BDD-Gate: 验证 `E2E.P0.077` 通过（happy path）
BDD-Gate: 验证 `E2E.P0.078` 通过（failure retryable / non-retryable）

### Phase 8: accept / reject suggestion 终态状态机

#### 8.1 handler `accept_suggestion.go` / `reject_suggestion.go`
- 实现 generated server interface `AcceptResumeTailorSuggestion` / `RejectResumeTailorSuggestion`。
- IK 必带（D-8）；二次 accept 同 IK 走 idempotency middleware replay；不同 fingerprint 同 key 409 generic IK conflict。
- 终态状态机：suggestion `pending` → accept 写 `status='accepted' + decided_at=now()`；reject 写 `status='rejected' + decided_at=now()`；已终态再调走 IK replay 或返回 409 `error.code='VALIDATION_FAILED' + detail.reason='SUGGESTION_ALREADY_DECIDED'`。
- accept 不自动更新 `resume_versions.structured_profile`（D-12）；客户端如需应用建议，后续调用 `updateResumeVersion` 显式 patch。
- cross-user：suggestion 关联 version 不属本用户 → 404；reject 同样。

#### 8.2 store `resume_version_suggestions.Decide`
- 单事务 UPDATE：CAS pending → accepted/rejected；其他状态返回 already-decided error。
- 返回更新后的 `ResumeVersion`（按 B2 fixture 响应 schema）。

#### 8.3 unit / integration test
- handler unit：accept happy / reject happy / IK replay / IK mismatch 409 / already-decided 409 / cross-user 404。
- store integration：CAS 并发只有一方成功；其他方收到 already-decided。

#### 8.4 `cmd/api` route wiring
- 挂载 `POST /api/v1/resume-versions/{resumeVersionId}/suggestions/{suggestionId}/accept` + `.../reject`，含 IK middleware。
- HTTP scenario test 覆盖 accept happy / replay / 409 / reject happy / cross-user。

BDD-Gate: 验证 `E2E.P0.079` 通过

### Phase 9: 收口 + BDD + 解锁前端 002

#### 9.1 跨 gate 收口
- `cd backend && go test ./...` PASS
- `cd backend && go test ./internal/resume/...` PASS
- `cd backend && go test ./cmd/api -run 'TestBuildResumeRuntime|TestResume.*HTTPScenario|TestResumeTailorDrainer.*' -count=1` PASS
- mock-first 对齐：9 个 op 通过 `cmd/api` 真实 route 的响应与 fixtures 字节比对 PASS（含 `confirmResumeStructuredMaster.json` 全 4 scenario / `branchResumeVersion.json` copy-master / blank / ai-select / `requestResumeTailor.json` + `getResumeTailorRun.json` status 四态等）
- grep `inline|rewrite|mirror` in `backend/internal/resume/` + resume_tailor drainer/outbox payload：0 命中（C-13 negative）
- grep `mistakes|growth|drill|inline-debrief-record` in `backend/internal/resume/`：0 命中（当前 plan / BDD gate prose 与历史 out-of-scope 文档不纳入 raw zero-hit 扫描，避免自匹配）
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS
- `make docs-check` PASS

#### 9.2 BDD 场景验证
- 创建 `test/scenarios/e2e/p0-074-resume-confirm-master-and-version-reads/` 至 `p0-080-resume-versions-privacy-legacy/` 七个场景目录与四段脚本（setup / trigger / verify / cleanup）。
- 在 `test/scenarios/e2e/INDEX.md` 追加 P0.074 – P0.080 行。
- 执行七个场景全 PASS。

#### 9.3 解锁前端 002
- 通知 [frontend-resume-workshop](../../../frontend-resume-workshop/) owner：14 个 Resume operationId 中本 plan 落地的 9 个真实 backend 已就位，Preview Confirm / 版本读写 / branch / tailor / accept-reject 可从 mock-first 切真。
- 在本 plan history（如未单独 plan history.md）/ retrospective 中记录"已可切真"信号；本 plan 不直接修订 frontend 文件。

#### 9.4 spec / history / INDEX 同步
- spec.md 1.2 / history.md 1.2 已写入（Phase 0.1）。
- `docs/spec/INDEX.md` 已同步 backend-resume 行（Phase 0 已统一升级到 1.2）。
- `plans/INDEX.md` 完成行从 Active → Completed（实施完成时切换）。

## 5 验收标准

- 本计划列出的 §4 所有 Phase task 全部完成
- §3 替代验证 gate 全部通过
- spec §6 C-9 partial / C-10 / C-11 / C-13 / C-14 / C-15 / C-16 全部 PASS
- `cmd/api` route/runtime gate PASS：9 个新 route 在真实 runtime 注册，session / IK middleware 行为与现有 backend-resume/001 一致；resume_tailor in-process drainer Start/Shutdown 与 deterministic `RunOnce` 均有测试证据
- 7 个 BDD 场景（E2E.P0.074 – P0.080）PASS
- [frontend-resume-workshop](../../../frontend-resume-workshop/) owner 已收到 9 个 op 切真信号
- `docs/spec/INDEX.md` 与 `plans/INDEX.md` 同步至 1.2 / completed

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: F3 `resume.tailor.gap_review` / `resume.tailor.bullet_suggestions` feature_key 缺失或 prompt schema 与本 plan 假设不一致 | Phase 0.2 preflight 强制校验；缺失则通知 F3 owner 修订 F3，不绕过；schema 校验通过 stub AIClient 早期 fail-fast |
| R2: partial UNIQUE INDEX 与历史 baseline `resume_versions` 数据冲突 | Phase 2.1 migration 仅 CREATE UNIQUE INDEX；如 baseline 已存在多条 structured_master / asset，必须先 dedupe（实际 baseline 应为空，因 backend-resume/001 不创建任何 resume_versions 行） |
| R3: confirmResumeStructuredMaster 与 branchResumeVersion 命名 / 用户预期混淆 | spec D-10 / D-11 明确语义；fixture 命名 `confirmResumeStructuredMaster.json`；frontend-resume-workshop/002 切真时 owner 协同审查 |
| R4: resume.tailor job 与 resume.parse job 共享 in-process drainer 时并发抢占 | drainer claim 限定 `job_type`；deterministic `RunOnce(jobType)` 入口；并发集成测试覆盖 |
| R5: accept suggestion 不写 structured_profile 引发用户体验混乱 | D-12 明确语义；frontend-resume-workshop/002 在 accept 后引导用户显式 patch；本 plan 不变更默认 |
| R6: 9 个新 route 真实 runtime wiring 通过包测试但实际 cmd/api 未挂 | Phase N.x checklist 强制 `cmd/api` HTTP scenario test 覆盖每个新 route，且 BDD scenario verify 必须输出 `method=cmd-api-http` 或等价 live runtime evidence；no-op / skip 视为 fail |
| R7: B2 D-18 additive 与 B1 错误码 lookup table 漂移 | Phase 1 一次性同步修订 openapi.yaml + fixtures + inventory + shared-conventions-codified + generated artifacts；codegen drift 与 `make lint-openapi` PASS 兜底 |
| R8: resume.tailor.completed envelope 含 suggested bullet 文本（PII 泄露） | Phase 7.2 outbox payload 写入路径只允许 `tailorRunId / resumeAssetId / targetJobId / mode / status`；专项 outbox unit test 字段集断言 + privacy negative grep |
