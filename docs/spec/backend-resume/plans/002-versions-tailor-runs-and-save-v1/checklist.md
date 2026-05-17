# Backend Resume Versions, Tailor Runs and Save v1 Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-17

**关联计划**: [plan](./plan.md)

## Phase 0: preflight + spec/plan 锁定

- [x] 0.1 spec.md 1.1 → 1.2：新增 D-10 / D-11 / D-12 锁定决策；新增 C-14 / C-15 / C-16 验收标准；§1 op 总数 13 → 14；§3.2 待确认事项与 §7 plan 002 行更新（验证：`grep -n 'D-10\|D-11\|D-12\|C-14\|C-15\|C-16' docs/spec/backend-resume/spec.md` 命中且 §1 含 "14 个 operationId"）
- [x] 0.2 history.md 追加 1.2 行（验证：`head -15 docs/spec/backend-resume/history.md` 含 2026-05-17 1.2 行）
- [x] 0.3 F3 feature_key preflight：验证 `config/prompts/resume.tailor.gap_review/v*` 与 `config/prompts/resume.tailor.bullet_suggestions/v*` 存在；如缺失通知 F3 owner，不在本 plan 修订 F3（验证：`find config/prompts -maxdepth 2 -type f \( -path '*/resume.tailor.gap_review/*' -o -path '*/resume.tailor.bullet_suggestions/*' \) -print | sort` 命中 8 个 prompt 文件；`find config/rubrics ...` 命中 4 个 rubric 文件；`python3 scripts/lint/prompt_lint.py --prompts-dir config/prompts --migrations-dir migrations` PASS；`python3 scripts/lint/rubric_lint.py --rubrics-dir config/rubrics` PASS）
- [x] 0.4 A3 AIClient preflight：验证 `resume.tailor.*` profile 可注入 stub provider，与 resume.parse 同款 stub harness 可加载（验证：`rg -n "resume\\.tailor|resume\\.parse|profiles:|name: .*resume" config/ai-profiles.yaml config/ai-providers.yaml backend/internal/ai` 命中 `resume.parse.default` / `resume.tailor.default` 与 resolver；`cd backend && go test ./internal/ai/aiclient/... -count=1` PASS；`cd backend && go test ./internal/ai/registry/... -count=1` PASS；`python3 scripts/lint/ai_profile_coverage.py --repo-root .` PASS）
- [x] 0.5 retired module / events drift baseline grep `grep -nE 'inline|rewrite|mirror'` + `grep -nE 'mistakes|growth|drill|inline-debrief-record'` 在 `backend/internal/resume/` 0 命中；当前 plan / BDD gate prose 与历史 out-of-scope 文档不纳入 raw zero-hit 扫描（验证：`git grep -nE 'inline|rewrite|mirror' -- backend/internal/resume/` exit 1 / 0 matches；`git grep -nE 'mistakes|growth|drill|inline-debrief-record' -- backend/internal/resume/` exit 1 / 0 matches；Phase 9.1 再次验证）

## Phase 1: B2 D-18 additive `confirmResumeStructuredMaster` + B1 错误码增补

- [x] 1.1 在 `openapi/openapi.yaml` 新增 `POST /api/v1/resumes/{resumeAssetId}/structured-master` operation（tag=Resumes, operationId=`confirmResumeStructuredMaster`, IK + XRequestID + Traceparent + AcceptLanguage + XClientVersion）；响应 201 `ResumeVersion` + 409 + default（验证：`make lint-openapi` PASS，`openapi inventory OK: 13 tags, 59 operations`）
- [x] 1.2 新增 `ConfirmResumeStructuredMasterRequest` schema：`required: [structuredProfile, displayName]`；`structuredProfile.provenance` $ref `GenerationProvenance`；optional `language`（验证：`make lint-openapi` PASS；`rg -n "ConfirmResumeStructuredMasterRequest|structured-master|GenerationProvenance" openapi/openapi.yaml` 命中新 operation / schema / provenance ref）
- [x] 1.3 在 `openapi/openapi.yaml` `ApiErrorResponse.error.code` enum 列表新增 `RESUME_STRUCTURED_MASTER_ALREADY_EXISTS`（验证：`rg -n "RESUME_STRUCTURED_MASTER_ALREADY_EXISTS" openapi/openapi.yaml shared/conventions.yaml backend/internal/shared/errors/codes.go frontend/src/lib/conventions/errors.ts` 命中；`make lint-openapi` PASS）
- [x] 1.4 新增 `openapi/fixtures/Resumes/confirmResumeStructuredMaster.json` 四个 scenario：`default`（201 + structured_master ResumeVersion）/ `idempotency-replay`（同 IK 返回首次结果）/ `already-exists-409`（409 `RESUME_STRUCTURED_MASTER_ALREADY_EXISTS`）/ `validation-422`（schema-valid blank displayName → 422 `VALIDATION_FAILED`；fixture request 必须保持 schema-valid）（验证：`make validate-fixtures` PASS，`validate-fixtures: OK — 59 fixtures`）
- [x] 1.5 在 `scripts/lint/openapi_inventory.py` baseline / IK_REQUIRED 与 openapi-v1-contract spec §3.1.1 endpoint 列表追加 `confirmResumeStructuredMaster`；全局 endpoint count 58 → 59，backend-resume owner surface 13 → 14（验证：`python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml` PASS；`docs/spec/openapi-v1-contract/spec.md` §3.1.1 新增 #49 行）
- [x] 1.6 `make codegen-openapi` + `make codegen-check`：Go server `ConfirmResumeStructuredMaster` 方法 + TS client SDK 类型出现在 generated artifacts；提交 generated diff（验证：`make codegen-check` exit 0；`backend/internal/api/generated/server.gen.go` 含 `ConfirmResumeStructuredMaster`，`frontend/src/api/generated/client.ts` 含 `confirmResumeStructuredMaster`）
- [x] 1.7 在 `docs/spec/shared-conventions-codified/spec.md` §3.1 D-5 / D-10 错误码 lookup table 增补 `RESUME_STRUCTURED_MASTER_ALREADY_EXISTS`，同步 history（验证：`rg -n "RESUME_STRUCTURED_MASTER_ALREADY_EXISTS|1\\.19" docs/spec/shared-conventions-codified/spec.md docs/spec/shared-conventions-codified/history.md` 命中；`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS）
- [x] 1.8 修订 `docs/spec/openapi-v1-contract/spec.md` §3.1.1 + history（加入第 14 个 Resume op 行）（验证：`rg -n "confirmResumeStructuredMaster|1\\.24|59 endpoint" docs/spec/openapi-v1-contract/spec.md docs/spec/openapi-v1-contract/history.md` 命中；`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS）
- [x] 1.9 `frontend/src/api/devMockClient.ts` 注册 `confirmResumeStructuredMaster` 与四个 fixture scenario（验证：`pnpm --dir frontend test src/api/devMockClient.test.ts` PASS，7 tests passed）

## Phase 2: B4 cross-owner addendum migration + `confirmResumeStructuredMaster` handler

- [x] 2.1 新增 `migrations/000007_resume_versions_structured_master_unique.up.sql` + `.down.sql`：partial UNIQUE INDEX `uq_resume_versions_structured_master_per_asset ON resume_versions (resume_asset_id) WHERE version_type='structured_master' AND deleted_at IS NULL`（验证：`python3 scripts/lint/migrations_lint.py --repo-root .` PASS）
- [x] 2.2 `make migrate-check` 静态 migration lint PASS；如本地具备 DATABASE_URL，跑 up/down roundtrip 并验证 INDEX 创建/删除（验证：`make dev-up` 已健康；`DATABASE_URL='postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' make migrate-check` PASS；`go run ./cmd/migrate ... status` 输出 `version=7 dirty=false`）
- [x] 2.3 在 `docs/spec/db-migrations-baseline/history.md` 追加 cross-owner addendum 行（owner: backend-resume/002；说明 partial UNIQUE INDEX 用途）（验证：`make docs-check` PASS；`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS）
- [x] 2.4 实现 `backend/internal/resume/handler/confirm_structured_master.go`：generated server interface `ConfirmResumeStructuredMaster`；IK 必带；cross-user 404；调 service（验证：`cd backend && go vet ./internal/resume/... ./cmd/api` PASS）
- [x] 2.5 实现 service `ConfirmStructuredMaster(ctx, userID, resumeAssetID, req)`：事务内 SELECT FOR UPDATE resume_assets → ownership + `parse_status='ready'` 校验（非 ready → 422 `VALIDATION_FAILED` + detail.reason=`PARSE_NOT_READY`）→ INSERT resume_versions structured_master 行；命中 partial UNIQUE INDEX 映射为 `409 + RESUME_STRUCTURED_MASTER_ALREADY_EXISTS`（验证：`cd backend && go test ./internal/resume/... -run TestConfirmStructuredMaster -count=1` PASS）
- [x] 2.6 实现 store `resume_versions.CreateStructuredMasterFromAsset`：单事务 INSERT + provenance 构造（验证：`cd backend && go test ./internal/resume/store/... -run 'TestCreateStructuredMaster|TestRepositoryExposesResumeAssetMethods' -count=1` PASS；`DATABASE_URL='postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' go test ./internal/resume/store/... -tags=integration -run TestStructuredMasterUnique -count=1 -v` PASS）
- [x] 2.7 handler unit test：happy / IK replay / 409 already-exists / 422 缺 structuredProfile / 422 parse_not_ready / 404 cross-user / 404 asset not exist 全分支（验证：`cd backend && go test ./internal/resume/handler/... -run TestConfirmStructuredMaster -count=1` PASS；`cd backend && go test ./internal/resume/handler/... -count=1` PASS）
- [x] 2.8 store integration test：（a）partial UNIQUE INDEX 在并发 INSERT 仅一行成功；（b）IK replay 不创建新行；（c）cross-user filter；（d）deleted_at IS NOT NULL 后允许新建 structured_master（验证：`cd backend && DATABASE_URL='postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' go test ./internal/resume/store/... -tags=integration -run TestStructuredMasterUnique -count=1 -v` PASS；IK replay 由 `TestResumeConfirmStructuredMasterHTTPScenario/idempotency_replay_bypasses_service` 覆盖）
- [x] 2.9 在 `cmd/api` `buildResumeRuntime` 注册新 route `POST /api/v1/resumes/{resumeAssetId}/structured-master`，session + IK middleware 链生效（验证：`cd backend && go test ./cmd/api -run 'TestBuildAPIHandlerMountsResumeRoutesBehindSessionMiddleware|TestBuildResumeRuntime|TestResumeConfirmStructuredMasterHTTPScenario' -count=1` PASS）
- [x] 2.10 `cmd/api` HTTP scenario test：route 存在、缺 session 401、缺 IK 422、IK replay 返回首次结果、并发 409（验证：`cd backend && go test ./cmd/api -run TestResumeConfirmStructuredMasterHTTPScenario -count=1` PASS）

## Phase 3: `getResumeVersion` + `listResumeVersions` handler

- [ ] 3.1 实现 `backend/internal/resume/handler/get_version.go`：generated server interface `GetResumeVersion`；cross-user 返回 404 不暴露存在（验证：编译 PASS）
- [ ] 3.2 实现 `backend/internal/resume/handler/list_versions.go`：generated server interface `ListResumeVersions`；asset ownership 404；cursor pagination 按 `updated_at DESC, id DESC` 唯一稳定序；invalid cursor → 422 `VALIDATION_FAILED`（验证：编译 PASS）
- [ ] 3.3 实现 store `resume_versions.GetByID` + `ListByAsset(userID, resumeAssetID, cursor, pageSize)`，复用 B4 002 索引（验证：store unit + integration test PASS）
- [ ] 3.4 handler unit test：getResumeVersion 200 / 404 cross-user / 404 not exist；listResumeVersions empty / 25 行 + 第二页 + `hasMore=false` / asset 不属本人 404 / invalid cursor 422（验证：`cd backend && go test ./internal/resume/handler/... -run 'TestGetResumeVersion|TestListResumeVersions' -count=1` PASS）
- [ ] 3.5 store integration test：cursor 边界 + cross-user + 排序稳定性（验证：`cd backend && go test ./internal/resume/store/... -tags=integration -run TestResumeVersionListPagination -count=1` PASS）
- [ ] 3.6 在 `cmd/api` 挂载 `GET /api/v1/resume-versions/{resumeVersionId}` + `GET /api/v1/resumes/{resumeAssetId}/versions`（验证：`cd backend && go test ./cmd/api -run TestResumeVersionReadHTTPScenario -count=1` PASS）
- [ ] 3.7 字节比对 [B2 fixture](../../../mock-contract-suite/spec.md) `getResumeVersion.json` `default` / `not-found-404` + `listResumeVersions.json` `default` / `empty` / `paginated`（验证：fixture parity test PASS）
- [ ] 3.8 BDD-Gate: 验证 `E2E.P0.074` 通过（confirmStructuredMaster + getResumeVersion + listResumeVersions 综合主路径，详见 [bdd-checklist.md](./bdd-checklist.md)）

## Phase 4: `updateResumeVersion` handler

- [ ] 4.1 实现 `backend/internal/resume/handler/update_version.go`：generated server interface `UpdateResumeVersion`；PATCH 语义；不可编辑字段（`versionType` / `resumeAssetId` / `parentVersionId` / `targetJobId` / `seedStrategy` / `provenance.promptVersion` 等 server-owned 字段）出现在请求体即 422（验证：编译 PASS）
- [ ] 4.2 structured_profile partial merge：客户端 partial jsonb 与现有值合并；`provenance` 字段强制由 server 重置；`displayName` / `focusAngle` / `matchScore` 直接覆盖（验证：service unit test 覆盖 merge case + server 重置 provenance case）
- [ ] 4.3 IK 必带（D-8）；同 IK + 同 fingerprint 走 idempotency middleware replay；同 IK + 不同 fingerprint 409（验证：handler unit + middleware test）
- [ ] 4.4 实现 store `resume_versions.UpdatePatch`：单事务 UPDATE + 返回更新行；cross-user 404；deleted_at 行 404（验证：store integration test PASS）
- [ ] 4.5 handler unit test：happy（包括 structured_profile partial merge）/ IK replay / IK mismatch 409 / 422 不可编辑字段 / cross-user 404 / 404 deleted_at（验证：`cd backend && go test ./internal/resume/handler/... -run TestUpdateResumeVersion -count=1` PASS）
- [ ] 4.6 在 `cmd/api` 挂载 `PATCH /api/v1/resume-versions/{resumeVersionId}` 含 IK middleware（验证：`cd backend && go test ./cmd/api -run TestResumeUpdateVersionHTTPScenario -count=1` PASS）
- [ ] 4.7 字节比对 [B2 fixture `updateResumeVersion.json`](../../../mock-contract-suite/spec.md) `default` / `validation-error-422`，并补齐 `idempotency-replay` variant（验证：fixture parity test PASS）
- [ ] 4.8 BDD-Gate: 验证 `E2E.P0.075` 通过

## Phase 5: `branchResumeVersion` 三路 seed_strategy

- [ ] 5.1 实现 `backend/internal/resume/handler/branch_version.go`：generated server interface `BranchResumeVersion`；IK 必带；parent existence + cross-user 404；targetJobId 不属本用户 404（验证：编译 PASS）
- [ ] 5.2 seed_strategy 路由实现：
  - `copy_master` → 同步 INSERT `resume_versions(version_type='targeted', parent_version_id=parent.id, target_job_id=req.targetJobId, structured_profile=parent.structured_profile copy)`；返回 201 + `ResumeVersion`
  - `blank` → 同步 INSERT empty structured_profile；返回 201 + `ResumeVersion`
  - `ai_select` → 同事务 INSERT `resume_versions` + `resume_tailor_runs(status='queued', mode='gap_review')` + `async_jobs(job_type='resume_tailor', resource_type='resume_tailor_run')`；返回 202 + `BranchResumeVersionAccepted{resumeVersionId, version, job}`
  （验证：service unit test 三路 + ai_select rollback case）
- [ ] 5.3 实现 store `resume_versions.BranchFromParent`：三路单事务 INSERT；ai_select 路径三行原子；rollback 不留 orphan（验证：store integration test PASS）
- [ ] 5.4 handler unit test：copy_master 同步 201 + 字段拷贝；blank 同步 201 + 空 structured_profile；ai_select 异步 202 + Job(queued)；IK replay 三路；parent 不存在 404；targetJob 不存在 404；invalid seed_strategy 422（验证：`cd backend && go test ./internal/resume/handler/... -run TestBranchResumeVersion -count=1` PASS）
- [ ] 5.5 store integration test：三路 INSERT 完整 + ai_select rollback（模拟 async_jobs 写失败时 resume_versions / resume_tailor_runs 不留 orphan）+ cross-user filter（验证：`cd backend && go test ./internal/resume/store/... -tags=integration -run TestBranchVersion -count=1` PASS）
- [ ] 5.6 在 `cmd/api` 挂载 `POST /api/v1/resume-versions` 含 IK middleware（验证：`cd backend && go test ./cmd/api -run TestResumeBranchVersionHTTPScenario -count=1` PASS）
- [ ] 5.7 字节比对 [B2 fixture `branchResumeVersion.json`](../../../mock-contract-suite/spec.md) `default` / `copy-master-sync` / `blank-sync` / `ai-select-202-with-job` / `idempotent-replay` / `validation-error-422`（验证：fixture parity test PASS）
- [ ] 5.8 BDD-Gate: 验证 `E2E.P0.076` 通过（copy_master / blank 同步）
- [ ] 5.9 BDD-Gate: 验证 `E2E.P0.077` 通过的 ai_select dispatch 部分（async job 后续逻辑由 Phase 7 落地）

## Phase 6: `requestResumeTailor` + `getResumeTailorRun` handler

- [ ] 6.1 实现 `backend/internal/resume/handler/request_tailor.go`：generated server interface `RequestResumeTailor`；IK 必带；mode ∈ `gap_review | bullet_suggestions`（B3 D-14 / 本 spec D-5）；其他值 422；targetJobId / resumeAssetId 不属本用户 404（验证：编译 PASS）
- [ ] 6.2 同事务创建 `resume_tailor_runs(status='queued', mode)` + `async_jobs(job_type='resume_tailor')`；返回 202 + `{tailorRunId, job(jobType=resume_tailor, status=queued)}`（验证：service unit + store integration test PASS）
- [ ] 6.3 实现 `backend/internal/resume/handler/get_tailor_run.go`：generated server interface `GetResumeTailorRun`；cross-user 404；status 四态读（queued / generating / ready / failed）（验证：编译 PASS）
- [ ] 6.4 实现 store `resume_tailor_runs` Repository：`Create / Get / MarkGenerating / MarkReady(matchSummary, suggestions) / MarkFailed(errorCode)`；state machine `queued → generating → ready | failed`（验证：store unit + integration test PASS）
- [ ] 6.5 handler unit test：requestTailor happy / IK replay / mode 422 / targetJob 404 / resumeAsset 404；getTailorRun happy / cross-user 404 / status 四态（验证：`cd backend && go test ./internal/resume/handler/... -run 'TestRequestResumeTailor|TestGetResumeTailorRun' -count=1` PASS）
- [ ] 6.6 store integration test：state transition + cross-user + concurrent claim from queued（验证：`cd backend && go test ./internal/resume/store/... -tags=integration -run TestResumeTailorRunStore -count=1` PASS）
- [ ] 6.7 在 `cmd/api` 挂载 `POST /api/v1/resume/tailor` + `GET /api/v1/resume/tailor-runs/{tailorRunId}` 含 IK middleware（验证：`cd backend && go test ./cmd/api -run TestResumeTailorEndpointsHTTPScenario -count=1` PASS）
- [ ] 6.8 字节比对 [B2 fixture `requestResumeTailor.json`](../../../mock-contract-suite/spec.md) `default`（修复 request header 缺 `Idempotency-Key`）+ 本 plan 补齐 `idempotency-replay`；`getResumeTailorRun.json` `default` + 本 plan 补齐 `queued` / `generating` / `failed` status variant（验证：fixture parity test PASS）
- [ ] 6.9 BDD-Gate: 验证 `E2E.P0.077` 通过的 requestTailor + getTailorRun 部分

## Phase 7: resume.tailor async job + AIClient + outbox `resume.tailor.completed`

- [ ] 7.1 实现 `backend/internal/resume/jobs/tailor.go` 与 resume_tailor in-process drainer 注册（job_type=resume_tailor / dotted=resume.tailor）；不引入独立 worker binary / `WORKER_*` config / `backend-async-runtime` 旧 shorthand（验证：runner registry / topology negative 测试 PASS）
- [ ] 7.2 从 `resume_tailor_runs` 读 `resume_asset_id / target_job_id / mode`，构造 prompt input；按 mode 路由 [F3 feature_key](../../../prompt-rubric-registry/spec.md) `resume.tailor.gap_review` / `resume.tailor.bullet_suggestions`；不 hardcode prompt 正文（验证：unit test stub AIClient verify profile / feature_key 路由）
- [ ] 7.3 通过 [A3 AIClient](../../../ai-provider-and-model-routing/spec.md) 调用 → 解析 LLM JSON 输出 → 写 `resume_tailor_runs.status='ready'` + `match_summary` + `provenance`；写 N 行 `resume_version_suggestions(status='pending')`（验证：jobs unit test happy path PASS）
- [ ] 7.4 失败路径：AI timeout / output_invalid → `resume_tailor_runs.status='failed'` + `error_code`；retryable 信息落在 `async_jobs` retry metadata，不向 `resume_tailor_runs.status` 私加 `failed_retryable`；retry 时允许从 `failed` 回到 `generating`（验证：unit test `TestResumeTailorFailureRetryable` PASS）
- [ ] 7.5 写入 `ai_task_runs` typed columns：model_profile_name / model_profile_version / prompt_version / rubric_version / route / validation_status / feature_key（验证：integration test verify `ai_task_runs` 行）
- [ ] 7.6 outbox `resume.tailor.completed`：仅最终 `status='ready'` 写入；envelope 字段 `tailorRunId / resumeAssetId / targetJobId / mode / status`（按 [B3 §3.1.4](../../../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory)）；AI output invalid / provider timeout / retry exhausted 等失败路径不发 completed event；PII 边界：不含 `original_bullet` / `suggested_bullet` / `match_summary` 文本 / prompt input / model raw response（验证：outbox unit test 字段集 + 失败 negative + PII 字符串 grep）
- [ ] 7.7 resume_tailor in-process drainer wiring：`cmd/api` 进程内 claim `async_jobs(job_type=resume_tailor)`；deterministic `RunOnce(resume_tailor)` 入口；`Start(ctx)` / `Shutdown(ctx)` 随 `cmd/api` lifecycle；与 resume_parse drainer 并发不互相 claim（验证：`cd backend && go test ./internal/resume/jobs/... -run TestResumeTailorDrainer -count=1` PASS）
- [ ] 7.8 PII leak negative：log / audit / outbox payload 路径不序列化 prompt body / model raw response / suggested bullet text / match_summary 内容；允许 SQL/store 层出现必要列名，禁止把列值写入日志或事件（验证：专用 lint / unit test 覆盖 log sink 与 outbox payload）
- [ ] 7.9 `cmd/api` 把 resume_tailor drainer 纳入 `Start(ctx)` / `Shutdown(ctx)` lifecycle（验证：`cd backend && go test ./cmd/api -run 'TestBuildResumeRuntime|TestResumeTailorDrainerHTTPScenario' -count=1` PASS）
- [ ] 7.10 BDD-Gate: 验证 `E2E.P0.077` 通过（resume.tailor async happy + outbox event）
- [ ] 7.11 BDD-Gate: 验证 `E2E.P0.078` 通过（resume.tailor failure retryable / non-retryable + ready-only outbox）

## Phase 8: accept / reject suggestion 终态状态机

- [ ] 8.1 实现 `backend/internal/resume/handler/accept_suggestion.go` 与 `reject_suggestion.go`：generated server interface `AcceptResumeTailorSuggestion` / `RejectResumeTailorSuggestion`；IK 必带（D-8）；handler 仅做参数解析与错误码映射（验证：编译 PASS）
- [ ] 8.2 终态状态机：suggestion `pending` → accept 写 `status='accepted' + decided_at`；reject 写 `status='rejected' + decided_at`；已终态再调走 IK replay 或 409 `error.code='VALIDATION_FAILED' + detail.reason='SUGGESTION_ALREADY_DECIDED'`；accept 不自动更新 `resume_versions.structured_profile`（D-12）（验证：service unit test 状态机分支 + accept-no-structured-profile-mutation 断言）
- [ ] 8.3 实现 store `resume_version_suggestions.Decide(suggestionID, decision)`：CAS pending → accepted/rejected；其他状态返回 already-decided error；返回更新后的 `ResumeVersion`（验证：store integration test CAS 并发只有一方成功）
- [ ] 8.4 cross-user：suggestion 关联 version 不属本用户 → 404（验证：handler unit test + store integration test）
- [ ] 8.5 handler unit test：accept happy / reject happy / IK replay / IK mismatch 409 / already-decided 409 / cross-user 404（验证：`cd backend && go test ./internal/resume/handler/... -run 'TestAcceptSuggestion|TestRejectSuggestion' -count=1` PASS）
- [ ] 8.6 在 `cmd/api` 挂载 `POST /api/v1/resume-versions/{resumeVersionId}/suggestions/{suggestionId}/accept` + `.../reject` 含 IK middleware（验证：`cd backend && go test ./cmd/api -run TestResumeSuggestionAcceptRejectHTTPScenario -count=1` PASS）
- [ ] 8.7 字节比对 [B2 fixture `acceptResumeTailorSuggestion.json` / `rejectResumeTailorSuggestion.json`](../../../mock-contract-suite/spec.md) `default` + 本 plan 补齐 `idempotency-replay` / `already-decided-409`，并将 current `conflict-409` + `TARGET_INVALID_STATE_TRANSITION` 漂移收敛为 `VALIDATION_FAILED` + `detail.reason='SUGGESTION_ALREADY_DECIDED'`（验证：fixture parity test PASS + grep `TARGET_INVALID_STATE_TRANSITION` 在两个 suggestion fixture 0 命中）
- [ ] 8.8 BDD-Gate: 验证 `E2E.P0.079` 通过

## Phase 9: 收口 + BDD + 解锁前端 002

- [ ] 9.1 跑 `cd backend && go test ./...` + `cd backend && go test ./internal/resume/...` + `cd backend && go test ./cmd/api -run 'TestBuildResumeRuntime|TestResume.*HTTPScenario|TestResumeTailorDrainer.*' -count=1` 全 PASS（验证：exit 0）
- [ ] 9.2 mock-first 对齐：9 个新 op 真实响应与 fixture 字节比对 PASS（含 `confirmResumeStructuredMaster.json` 全 4 scenario / `branchResumeVersion.json` 4 scenario / `requestResumeTailor.json` + `getResumeTailorRun.json` status 四态 / `acceptResumeTailorSuggestion.json` + `rejectResumeTailorSuggestion.json` 3 scenario）
- [ ] 9.3 grep `inline|rewrite|mirror` in `backend/internal/resume/` + resume_tailor drainer/outbox payload：0 命中（C-13 negative）（验证：`git grep -nE 'inline|rewrite|mirror' backend/internal/resume/` 0 命中 + payload assertion）
- [ ] 9.4 grep `mistakes|growth|drill|inline-debrief-record` in `backend/internal/resume/`：0 命中（验证：`git grep -nE 'mistakes|growth|drill|inline-debrief-record' backend/internal/resume/` 0 命中；当前 plan / BDD gate prose 与历史 out-of-scope 文档不纳入 raw zero-hit 扫描）
- [ ] 9.5 BDD-Gate: E2E.P0.074 resume-confirm-master-and-version-reads PASS（详见 [bdd-checklist.md](./bdd-checklist.md)）
- [ ] 9.6 BDD-Gate: E2E.P0.075 resume-update-version-merge-and-ik PASS
- [ ] 9.7 BDD-Gate: E2E.P0.076 resume-branch-version-sync-paths PASS
- [ ] 9.8 BDD-Gate: E2E.P0.077 resume-tailor-async-dispatch-and-ready PASS
- [ ] 9.9 BDD-Gate: E2E.P0.078 resume-tailor-failure-and-retry PASS
- [ ] 9.10 BDD-Gate: E2E.P0.079 resume-suggestion-accept-reject-terminal PASS
- [ ] 9.11 BDD-Gate: E2E.P0.080 resume-versions-privacy-and-legacy-negative PASS
- [ ] 9.12 在 `test/scenarios/e2e/INDEX.md` 追加 P0.074 – P0.080 七行（关联需求 `backend-resume C-9, C-10, C-11, C-13, C-14, C-15, C-16`）
- [ ] 9.13 通知 [frontend-resume-workshop](../../../frontend-resume-workshop/) owner：9 个 op 已就位，可启动 mock-first → real backend 切真原地修订（验证：cross-plan 信号 commit + frontend-resume-workshop spec / plan 修订入口准备）
- [ ] 9.14 同步 `docs/spec/INDEX.md` backend-resume 行（1.1 → 1.2）；同步 `docs/spec/backend-resume/plans/INDEX.md`，将 002 行从 Active → Completed（验证：`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS）
- [ ] 9.15 spec / history 收口：再次 `make docs-check` PASS + `git diff --check` 0 残留 trailing whitespace
