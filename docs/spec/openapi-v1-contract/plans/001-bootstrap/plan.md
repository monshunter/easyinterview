# 001 - OpenAPI v1 Contract Bootstrap

> **版本**: 1.31
> **状态**: completed
> **更新日期**: 2026-07-15

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

落地当前 B2 OpenAPI v1 contract bootstrap：

- `openapi/openapi.yaml` is the single HTTP contract truth source for current 37 operations / 10 tags.
- Go generated server/types live in `backend/internal/api/generated/`.
- TypeScript generated client/types live in `frontend/src/api/generated/`.
- Root Make targets provide `codegen-openapi`, `codegen-check`, `lint-openapi`, and `docs-openapi`.
- B1 shared conventions are referenced through generated/shared types and error envelope rules.
- Fixtures and breaking-change gates consume this bootstrap output through sibling B2 plans.

This owner plan remains the executable contract/codegen evidence index；Phase 18 retains the in-flight report-conversation handoff，and Phase 19 adds the approved OPENAPI-007 minimal `UserContext` correction without replacing or weakening Phase 18 gates。

## 2 Current Contract

| Surface | Current contract | Gate |
|---------|------------------|------|
| OpenAPI inventory | 10 tags, 37 operations, `/api/v1` prefix, session-cookie auth, public/protected operation security | `make lint-openapi`, inventory tests |
| Error envelope | B1 `ApiError` inner object + B2 `ApiErrorResponse` wire envelope | generator tests, codegen-check |
| Shared types | B1 enum/page/error conventions are reused; OpenAPI does not duplicate shared enum ownership | conventions drift and generated tests |
| Codegen | Go server/types and TS client/types are reproducible from `openapi/openapi.yaml` | `make codegen-openapi`, `make codegen-check` |
| Local docs | Redocly CLI renders `openapi/dist/index.html` without committing generated docs | `make docs-openapi` |
| Downstream handoff | 002 owns fixtures/mock source; 003 owns breaking-change baseline/gate; 004 owns resume additive coverage | plans INDEX and context validation |

## 3 质量门禁分类

- **Plan 类型**: `contract + tooling + feature-behavior handoff`
- **TDD 策略**: schema inventory、semantic lint、Go/TS generator structure、codegen idempotency 与 negative surface tests 必须按 Red-Green-Refactor 执行；每个 correction Phase 的 checklist 明确对应断言与命令入口。
- **BDD 策略**: 本 plan 不复制场景资产；用户可见 contract correction 必须引用下游 owner BDD。Phase 18 由 report owners / `E2E.P0.099` 承接；Phase 19 由 frontend-shell settings BDD 与原地扩展的 `E2E.P0.101` 承接。本 plan 只拥有 schema/codegen correction，因此不新建 BDD 文件。
- **替代验证 gate**: `make lint-openapi`、generator tests、`make codegen-check`、`make openapi-diff`、scoped zero-reference 与 downstream compile/consumer gates。

### 3.1 Current Operation Inventory

| Tag | Operations |
|-----|------------|
| Auth | `getMe`, `completeMyProfile`, `deleteMe`, `startAuthEmailChallenge`, `verifyAuthEmailChallenge`, `logout`, `getRuntimeConfig` |
| Uploads | `createUploadPresign` |
| Resumes | `listResumes`, `registerResume`, `getResume`, `getResumeSource`, `updateResume`, `duplicateResume`, `archiveResume`, `exportResume` |
| TargetJobs | `importTargetJob`, `listTargetJobs`, `getTargetJob`, `updateTargetJob`, `archiveTargetJob` |
| PracticePlans | `createPracticePlan`, `getPracticePlan` |
| PracticeSessions | `startPracticeSession`, `getPracticeSession`, `sendPracticeMessage`, `completePracticeSession`, `createPracticeVoiceTurn` |
| Reports | `getFeedbackReport`, `getReportConversation`, `listTargetJobReports` |
| ResumeTailor | `requestResumeTailor`, `getResumeTailorRun` |
| Jobs | `getJob` |
| Privacy | `requestPrivacyExport`, `requestPrivacyDelete`, `getPrivacyRequest` |

## 4 Completed Implementation Scope

- OpenAPI 3.1 document with fixed server prefix, tags, security schemes, shared components, idempotency headers, request/response schemas, and default error envelope.
- OpenAPI inventory lint for operation/tag count, operation IDs, idempotency rules, privacy export exception, and schema provenance requirements.
- Go and TS codegen pipeline with reproducible generated artifacts.
- Local API docs renderer using the current Redocly CLI target.
- Codegen and inventory validation integrated into root Make targets.
- Handoff to fixture/mock source and breaking-change gate owner plans.

## 2.3 Make 入口

Current root Make targets owned by this plan:

- `make lint-openapi`: validates `openapi/openapi.yaml` and the current 10-tag / 37-operation inventory.
- `make codegen-openapi`: regenerates Go and TypeScript OpenAPI artifacts.
- `make codegen-check`: verifies generated OpenAPI artifacts are reproducible and not drifted.
- `make docs-openapi`: renders the local OpenAPI HTML document.

## 5 Verification Commands

```bash
make lint-openapi
make codegen-openapi
make codegen-check
cd backend && go test ./cmd/codegen/openapi -count=1
make docs-openapi
python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/openapi-v1-contract/plans/001-bootstrap/context.yaml --target contract
python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check
make docs-check
git diff --check
```

## 6 BDD Applicability

BDD 不适用：本 plan 只拥有 API schema、codegen 与 contract gate，不创建 BDD 文件或引用场景编号。用户可见行为由 domain owner 独立验收；本 plan 只等待 fixture、backend/frontend consumer 与 generated contract handoff。阶段收口从仓库根执行 `make test`。

## 7 Revision Log

| Date | Version | Change |
|------|---------|--------|
| 2026-07-15 | 1.31 | Revise Phase 19 to expose the complete authenticated account email as `email`, delete `emailMasked`, and preserve log/evidence privacy boundaries. |
| 2026-07-15 | 1.30 | Add Phase 19 for OPENAPI-007 four-field UserContext, Auth fixture/codegen migration and old-language-field zero-reference gates. |
| 2026-07-15 | 1.29 | Reopen Phase 18 for the approved one-for-one replacement of public listPracticeSessions with report-owned getReportConversation while preserving 37/10. |
| 2026-07-14 | 1.28 | Reopen Phase 17 for closed RuntimeConfig ContentLimits public projection and generated consumer handoff. |
| 2026-07-14 | 1.27 | Reopen Phase 16 for OPENAPI-005 closed ResumeSummary list projection, full getResume detail and all-consumer handoff. |
| 2026-07-14 | 1.23 | Correct OPENAPI-002 to an exact 17-finding boundary including both source-only ApiErrorCode removals; keep the separate D-35 Practice machine oracle non-ADR. |
| 2026-07-13 | 1.22 | Add Practice durable reply-state, same-ID recovery and typed TypeScript ApiClientError phase; tighten OPENAPI-002 rawText/oracle/invariant gates. |
| 2026-07-13 | 1.21 | Reopen Phase 13 for OPENAPI-002 TargetJob paste-only schema, generated artifacts and consumer handoff. |
| 2026-07-13 | 1.20 | Keep max4 generation/judge audit internal-only；no attempt/retry/reason/scope/progress field or retry endpoint，and no new expected OpenAPI finding. |
| 2026-07-13 | 1.19 | Clarify downstream product full-validator repair responsibility without changing expected OpenAPI finding maxLength200；audit/codegen remain pending. |
| 2026-07-13 | 1.18 | Finalize A：wire fuse200 code points；semantic/UX 24/64；targeted action-label repair margin18/52；keep audit/re-freeze/codegen pending. |
| 2026-07-12 | 1.15 | Complete baseline/derived CreatePracticePlanRequest conditional positive/negative matrix and non-null source contract. |
| 2026-07-12 | 1.14 | Require B1 oversized-context enum marker and include its additive-only oracle classification before report codegen. |
| 2026-07-13 | 1.17 | A-200：set ReportNextAction.label fuse200；keep14/40 UX gate and reopen audit/re-freeze/codegen evidence. |
| 2026-07-13 | 1.16 | Separate wire fuse from14/40 UX responsibility；specific bound superseded by A-200. |
| 2026-07-12 | 1.13 | Reopen Phase 12 for OPENAPI-001 closed grounded direct report schema and Go/TS codegen. |
| 2026-07-12 | 1.12 | Reopen Phase 11 for additive practice round identity and TargetJob progress projection contract. |
| 2026-07-10 | 1.11 | Remove the unconsumed frontend raw OpenAPI snapshot output and its dedicated generator surface. |
| 2026-07-10 | 1.10 | Remove the unreferenced provenance ref constant from the inventory linter. |
| 2026-07-10 | 1.9 | Move the test-only snapshot hash calculation out of the production codegen package. |
| 2026-07-10 | 1.8 | Align owner inventory with the current 37-operation contract including `getResumeSource` and `archiveTargetJob`. |
| 2026-07-07 | 1.7 | Compress owner plan to the 2026-07-07 36-operation / 10-tag OpenAPI contract and executable evidence index. |
| 2026-05-04 | 1.6 | Complete OpenAPI v1 bootstrap delivery. |

## 8 Test-only snapshot hash cleanup

删除只被 `run_test.go` 使用的 production `sha256.go`。幂等测试在 snapshot traversal 内直接计算 SHA-256，保持 byte-identical generated artifact 断言不变。

## 9 Inventory linter dead constant cleanup

删除 `scripts/lint/openapi_inventory.py` 中无读取方的 `PROVENANCE_REF`；现有 `GenerationProvenance` schema shape 与可达性检查继续由真实 schema-name traversal 承担。

## 10 Frontend raw-spec snapshot removal

TypeScript codegen 只输出正式消费的 `client.ts` 与 `types.ts`。删除没有 import、未进入 Vite bundle、也不被 docs/mock tooling 读取的 raw OpenAPI 字符串快照，同时删除专用 TS template 与只服务该快照的字符串转义 helper；保留 `openapi/openapi.yaml`、backend generated spec 镜像、Redocly 文档和所有 wire/API contract 不变。

## 11 Practice round identity and progress projection

### 11.1 Additive wire contract

- `CreatePracticePlanRequest` 新增可选 `roundId`，只表达客户端选择的结构化轮次；`roundSequence` 必须由服务端从 TargetJob summary 推导，客户端不得提交。
- `PracticePlan` 新增可选 `roundId` / `roundSequence`。新创建记录必须成对返回；字段可选只用于读取 legacy null identity，不授权新路径省略。
- `TargetJob` 新增可选 `practiceProgress: PracticeProgress`；`PracticeProgress` 包含 `status=not_started|in_progress|completed`、有序去重的 `completedRounds: PracticeRoundRef[]` 与 nullable `currentRound`。

### 11.2 Compatibility and generated artifacts

变更只新增 schema / optional property，不新增 endpoint、不修改现有 required 字段或状态码。同步 `openapi/baseline/openapi-v1.0.0.yaml`、Go/TS generated artifacts，并运行 `make lint-openapi`、`make codegen-openapi`、`make codegen-check`、`make openapi-diff`；diff 必须分类为 additive。

### 11.3 Consumer invariant

TargetJob lifecycle `status` 只表示岗位生命周期，不能解释为面试轮次。backend-targetjob 只从完成 session 事实投影 `practiceProgress`；frontend 只消费 `practiceProgress.currentRound` 选择卡片当前轮和 quick-start，legacy null plan 不得按时长碰撞复用。

### 11.4 Operation matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticePlan` | `PracticePlans/createPracticePlan.json` round variants | shared parse/workspace/report start helper | backend-practice generated adapter/service/store | insert `practice_plans.round_id/round_sequence` + idempotency/audit | none | fixture + generated consumer tests |
| `getPracticePlan` | `PracticePlans/getPracticePlan.json` current + legacy-null | shared start exact-pair reuse; Practice budget | backend-practice generated adapter/store | read nullable legacy/current plan identity | none | fixture + generated consumer tests |
| `listTargetJobs` | `TargetJobs/listTargetJobs.json` zero/partial/final progress | Home/Workspace cards and quick-start | backend-targetjob list handler/store/service projection | TargetJob summary + plans/sessions/completion events; no mutable progress column | none | fixture + generated consumer tests |
| `getTargetJob` | `TargetJobs/getTargetJob.json` zero/partial/final progress | Parse/Report/current-round handoff | backend-targetjob get handler/store/service projection | same ledger projection as list | none after persisted JD parse | fixture + generated consumer tests |

## 12 OPENAPI-001 grounded direct report contract

### 12.1 Closed wire shape

After OPENAPI-001 is accepted and the B1 conventions gate passes, update the proposed `FeedbackReport` while the old baseline remains unchanged. Require nullable-until-state-resolved `summary` / `preparednessLevel` / `provenance`, non-null minimal `ReportContextSnapshot` with `hasNextRound`, code+label dimensions, dimensionCode evidence and `retryFocusDimensionCodes`. Close the report state machine: ready requires non-null summary/preparedness/provenance plus non-empty dimensions/actions; failed alone requires non-null errorCode and all other states require null errorCode. Keep CreatePracticePlanRequest as `type: object` with typed superset properties and conditional `oneOf`: baseline requires its existing non-focus fields and forbids sourceReportId; retry/next each require a non-null UUID sourceReportId and allow only goal+sourceReportId. Existing generators must produce Go/TS object types rather than `any`; schema/runtime fixtures reject invalid report states, baseline source, derived missing/null/blank source and every derived extra, while valid ready/failed and minimal retry/next requests pass. Delete `DimensionResult` and old dimension/focus names, close objects and apply ADR bounds. Synchronize `REPORT_CONTEXT_TOO_LARGE` into `ApiErrorCode` only through the B1 source. Run 003 base-ref exact audit across breaking and additive findings before codegen/fixtures.

Under approved方案 A，`ReportNextAction.label.maxLength=200` code points is only a malformed-output fuse。F3/runtime/frontend enforce English `<=24` whitespace words / zh-CN `<=64` Unicode code points；targeted action-label repair uses an internal 18/52 generation margin and is revalidated against 200+24/64。Downstream UI owner proves desktop/mobile wrapping and typed-invalid/no-raw behavior。Old baseline clean PASS is invalid until new audit/re-freeze；codegen-check remains separately pending。

Generation/judge max4，attempt_count/retry_count/reason/scope and business/outbox backoff are internal runtime/eval contracts。Do not add HTTP fields、progress or retry-generation operation。`FeedbackReport` remains queued/generating/ready/failed；frontend maxAttempts49 exhaustion is a client continue-check state，not an API failed transition。This decision creates no additional expected finding beyond the existing maxLength200 correction。

### 12.2 Generated artifacts and inventory

Regenerate Go/TS artifacts and add exact inventory/negative tests that reject old report fields and unknown object properties. The operation/tag inventory remains 37/10; no endpoint or status changes.

### 12.3 Handoff

002 must replace every Reports/PracticePlans scenario and prototype projection before backend/frontend consumers compile. 003 must preserve the merge-base breaking artifact before current baseline re-freeze.

## 13 OPENAPI-002 TargetJob paste-only contract

### 13.1 RED contract boundary

After [OPENAPI-002](../../decisions/OPENAPI-002-targetjob-paste-only.md) is accepted, keep the merge-base baseline untouched and add focused schema/inventory tests that require `ImportTargetJobRequest` to be a closed object with exactly required `rawText` / `targetLanguage` / `resumeId`. `rawText` must declare `minLength: 1` and `pattern: '\S'`; positive cases include non-whitespace text and negative cases include empty/space/tab/newline-only values. Negative cases must also reject the old `source` wrapper, every URL/file/manual-form discriminator payload, `fileObjectId`, `titleHint`, `companyNameHint` and arbitrary extra properties. Read-side tests must reject `TargetJob.sourceType` / `sourceUrl`; upload tests must reject `purpose=target_job_attachment` while proving `createUploadPresign` still supports resume/privacy consumers.

### 13.2 GREEN source, codegen and freeze invariant

Delete all five `TargetJobImportSource*` schemas, flatten and constrain `ImportTargetJobRequest`, remove TargetJob source provenance properties and remove only the TargetJob attachment enum value. The same correction removes source-only `TARGET_IMPORT_SOURCE_INVALID` / `TARGET_IMPORT_SOURCE_UNAVAILABLE` from `ApiErrorCode` while retaining `VALIDATION_FAILED` / `TARGET_IMPORT_FAILED`; both removals are independent entries in the exact 17-finding oracle. Regenerate Go/TS artifacts and assert no compatibility union, discriminator, alias or legacy generated type remains. `importTargetJob` remains operationId `importTargetJob`, `POST /api/v1/targets/import`, `202 + TargetJobWithJob`; `createUploadPresign` remains operationId `createUploadPresign`, `POST /api/v1/uploads/presign`, `201 + UploadPresign`; inventory remains 37 operations / 10 tags. The current baseline cannot be edited until 003 Phase 6 preserves the exact old-baseline audit and all downstream gates pass.

### 13.3 Operation matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `importTargetJob` | `TargetJobs/importTargetJob.json` paste-only `default` / `paste-primary` | Home paste submit and Parse polling | backend-targetjob generated adapter/service/runner | TargetJob + async job; no URL/file source provenance | parse pasted JD text | fixture + backend/frontend consumer tests |
| `createUploadPresign` | `Uploads/createUploadPresign.json` resume/privacy only | Resume Workshop/privacy consumers; no Home consumer | backend-upload generated adapter/service | file object for remaining purposes only | none | existing resume/privacy scenarios; no TargetJob attachment |
| `listTargetJobs` / `getTargetJob` | TargetJobs list/get without source fields | Home/Workspace/Parse/Report | backend-targetjob list/get adapters | current TargetJob projection without source provenance | none after persisted parse | fixture + backend/frontend consumer tests |

### 13.4 Zero-reference and handoff gates

Run exact current-contract searches over OpenAPI, generated artifacts, positive fixtures, frontend/backend runtime consumers and mock runtime. Positive/runtime occurrences of `TargetJobImportSource*`, `target_job_attachment`, TargetJob `sourceType/sourceUrl`, URL/file/manual-form import branches or compatibility aliases must be zero. Accepted ADR、expected-findings oracle 与显式 negative test/fixture declaration 可包含被拒绝 token；gate 必须按语义/可达面区分正负引用，禁止用整文件或整测试目录排除。用户流程由 downstream owner 验收；URL/file/manual-form 正向资产直接删除，不保留兼容覆盖。

## 14 Practice durable message recovery（方案 A）

### 14.1 RED: role-discriminated message contract

先新增 schema/generator RED，要求 `PracticeMessage` 是 closed role-discriminated union，而不是所有字段都 optional 的宽对象：

- user projection required `clientMessageId` 与 `replyStatus`；`replyStatus` exact enum 为 `pending|retryable_failed|terminal_failed|complete`；
- assistant projection 禁止 `clientMessageId` / `replyStatus`；
- `pending` 表示 user reservation 已持久化但 reply 尚未有权威终态；`retryable_failed` 表示 reservation 后发生 typed retryable failure；`terminal_failed` 表示 reservation 后发生 non-retryable reply failure；`complete` 必须关联恰好一条 assistant reply；
- validation/auth/not-found/conflict/mismatch 若在 reservation 前失败，不得创建新的 user message；transport outcome 不确定时 consumer 先通过 `getPracticeSession` 对账，不能从本地 error 文案猜测服务端是否已持久化。

Go/TS generated type tests must reject assistant recovery fields, user missing fields, unknown status and any fallback to `any`.

### 14.2 GREEN: persistence/read-side and replay semantics

`replyStatus` 与原始 `clientMessageId` 由 backend-practice persistence/read model 持久化，前端 memory、URL、localStorage/sessionStorage/IndexedDB 或“是否已有 assistant”均不能充当事实源。`getPracticeSession` reload 必须返回 ordered authoritative messages：

- `pending`：composer remains locked；client may reconcile/poll but cannot submit another text；
- `retryable_failed`：只允许用相同 `clientMessageId` 与完全相同 text 执行 retry；
- `terminal_failed`：不提供 retry，consumer 使用 reload/auth/session-lost 等 typed recovery；
- `complete`：same-ID replay 返回既有完成结果，不新增 user 或 assistant；
- same ID + different text/session 返回 `409 + IDEMPOTENCY_KEY_MISMATCH`；session 已有 unresolved reply conflict 返回 `409 + PRACTICE_SESSION_CONFLICT`。

Storage/handler tests must prove unique user reservation, at-most-one assistant reply, retryable failure → reload → same-ID success and concurrent replay determinism.

### 14.3 Typed TypeScript failure contract

Update the TS generator template and golden tests so generated client exports `ApiClientError` with public `status: number | null` and `apiError: ApiErrorResponse | null`（可另带 stable `kind=http|abort|transport` / `cause`；不得公开 fetch `Response`）。Exact tests cover：

1. non-2xx JSON `ApiErrorResponse` → HTTP status + parsed `apiError`；
2. non-JSON body → HTTP status + `apiError=null`，且不泄漏 raw body；
3. empty body → HTTP status + `apiError=null`；
4. abort → null status/apiError + abort discriminant；
5. transport rejection → null status/apiError + transport discriminant。

Consumer contract tests must fail on `error.message` regex/string parsing；retryability 只读取 `apiError.error.retryable`，status/code/details 只读取 typed fields。

### 14.4 Operation matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getPracticeSession` | `PracticeSessions/getPracticeSession.json`: pending/retryable-failed/terminal-failed/complete + reload recovery | Practice transcript hydration, composer lock and retry affordance | backend-practice get adapter/service/store | user `client_message_id` + durable reply status + ordered assistant projection | none on read | fixture + backend/frontend consumer tests |
| `sendPracticeMessage` | `PracticeSessions/sendPracticeMessage.json`: default plus validation/auth/not-found/conflict/mismatch/retryable failure | optimistic row, thinking state and typed retry dispatch | backend-practice generated adapter/service/store | reserve once by `(session_id,client_message_id)`; transition reply status; at-most-one assistant | interviewer generation | fixture + backend/frontend consumer tests |

### 14.5 Contract audit and downstream handoff

This union/persistence correction changes existing Practice message validation semantics. D-35 + history 1.54 + the product-approved 方案 A remain the sole governance authority. 003 must use a separate Practice machine oracle as D-35's executable five-key finding projection—not as a third `OPENAPI-NNN` ADR—and must never merge those findings into OPENAPI-002's exact 17 allowset. Preserve both owner-specific artifacts before baseline mutation, then wait for 002 fixture matrix、mock-contract-suite parity、backend-practice persistence and frontend-workspace-and-practice typed consumer tests before re-freeze. Operation/tag inventory stays 37/10；no retry endpoint, client-side business-state store or compatibility message schema is allowed.

## 15 OPENAPI-004 TargetJob canonical-round report overview

### 15.1 RED contract boundary

Consume accepted OPENAPI-004 before schema mutation. Focused schema/generator tests must fail while `listTargetJobReports` still accepts cursor/pageSize or returns `PaginatedFeedbackReport`, while `TargetJob.latestReportId` remains, or while the new overview permits unknown/missing fields. Lock the unchanged endpoint method/path/operationId/200 and 37/10 inventory.

### 15.2 GREEN minimal wire

Replace the response with closed `TargetJobReportsOverview{targetJobId,rounds}`. Each round item requires `round: PracticeRoundRef`, nullable `currentReport={id,generatedAt}` and nullable `latestAttempt={id,status,errorCode,createdAt}`; all properties are required, nullable values use explicit null, and failed-only errorCode is enforced. Remove cursor/pageSize, `PaginatedFeedbackReport` and `TargetJob.latestReportId`; do not expose full report, summary, provenance, model/rubric, session/plan locator, compatibility alias or replacement pointer. Regenerate typed Go/TS artifacts.

### 15.3 Operation matrix and handoff

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `listTargetJobReports` | `Reports/listTargetJobReports.json` canonical rounds/current/latest/empty/fail-closed cases | target-scoped ReportsScreen via generated client | backend-review list overview service/store | owned TargetJob canonical summary + `feedback_reports.generation_context/status/generated_at/created_at`；no TargetJob pointer | none | fixture + backend/frontend consumer tests |

003 must exact-audit OPENAPI-004 from old baseline before re-freeze. 002 fixtures/Prism, db/targetjob cleanup, backend-review selection and frontend consumers must all pass before baseline mutation.

## 16 OPENAPI-005 Resume list summary projection

### 16.1 RED contract boundary

Consume accepted OPENAPI-005 before schema mutation. Focused schema/inventory/generator tests must fail until `ResumeSummary` is closed, its required property set is exactly `id/title/displayName/language/sourceType/parseStatus/summaryHeadline/hasReadableContent/updatedAt`, and `PaginatedResume.items` references it. Negative assertions reject full-detail/provenance fields and arbitrary extras while locking unchanged 37/10 inventory, `GET /api/v1/resumes` / `listResumes` / 200 / pagination and `GET /api/v1/resumes/{resumeId}` / `getResume` / 200 + full `Resume`.

### 16.2 GREEN source and generated types

Add the closed source schema, switch only `PaginatedResume.items`, and regenerate typed Go/TS artifacts. `summaryHeadline` is required nullable；`hasReadableContent` is required boolean and backend-owned. The handoff contract pins `summaryHeadline` to the first trim-nonempty string in `parsed_summary.headline`、`parsed_summary.basics.headline`、`structured_profile.headline`、`structured_profile.basics.headline`, and `hasReadableContent=true` exactly to trim-nonempty `parsed_text_snapshot` / `original_text` or a nonempty-object `structured_profile`; `fileObjectId`、`sourceType`、`parseStatus` never imply readability. Generated list results must use `ResumeSummary[]` / `[]ResumeSummary`, while `getResume`, update/duplicate/archive responses keep `Resume`. No `ResumeListItem` alias, union, optional detail compatibility fields, second endpoint or `any` is allowed.

### 16.3 Operation matrix and all-consumer handoff

| operationId | response contract | fixture | backend projection | frontend consumer | BDD |
|-------------|-------------------|---------|--------------------|-------------------|-----|
| `listResumes` | `PaginatedResume.items: ResumeSummary[]` | `Resumes/listResumes.json` summary-only examples | dedicated list columns/record/mapper; no detail payload fetch | Home picker, Resume Workshop list and every generated-client list consumer | fixture + backend/frontend consumer tests |
| `getResume` | full `Resume` unchanged | `Resumes/getResume.json` full-detail examples | existing owned detail lookup/mapper | Resume Workshop read-only detail | fixture + backend/frontend consumer tests |

002 Phase 11 must migrate fixture/example/Prism/mock bytes；003 Phase 9 must generate the declared OPENAPI-005 exact oracle from merge-base old baseline before any re-freeze；004 Phase 7 coordinates backend store/service/handler, generated consumers and frontend list/detail migration. All consumers compile and pass focused tests in the same batch；frontend may not issue N+1 `getResume` requests to restore removed list fields.

## 17 RuntimeConfig content limits projection

### 17.1 RED/GREEN closed schema

Add required `RuntimeConfig.contentLimits` referencing a closed required `ContentLimits` object with exactly five positive int64 fields: `resumeUploadBytes`, `resumePasteTextBytes`, `targetJobRawTextBytes`, `practiceMessageBytes`, `practiceSessionTextBytes`. Generator tests reject missing/extra/internal fields and `any`/optional fallbacks.

### 17.2 Fixture/generated handoff

Update getRuntimeConfig fixture, generated Go/TS types and backend builder mapping. Contract fixtures use small positive representative values rather than复制 A4 默认数值。Explicitly reject report framed input, HTTP body, provider response and profile token values from the public schema. `ContentLimits` 及五个子字段均为 required；consumer 不得为缺失子字段逐项 fallback，只有整体 runtime source 不可用时才可沿用既有 bootstrap fallback。

### 17.3 Substitute contract gates

BDD 不适用：本 Phase 不新增用户流程，也不改变业务错误、恢复或持久化语义。替代 gate 为 closed schema semantic lint、fixture validation、Go/TS codegen drift、backend builder focused test、frontend compile/consumer test 与 internal-field negative search；operation/tag inventory remains 37/10 and no business request wire changes。不得要求任何 E2E 作为配置传播证据；阶段收口执行根 `make test`。

## 18 Report-owned conversation replacement

### 18.1 RED contract boundary

Consume OPENAPI-001 v1.7 and spec/history 1.61 before source mutation. Focused source/inventory/generator tests must fail until `GET /api/v1/practice/sessions listPracticeSessions` and its query/`PaginatedPracticeSession` surface are absent, while `GET /api/v1/reports/{reportId}/conversation getReportConversation` exists with protected 200/404/500 behavior and a closed response. Lock unchanged 37 operations / 10 tags plus preserved `startPracticeSession` and `getPracticeSession` method/path/operationId/status.

### 18.2 GREEN closed report conversation

Add closed `ReportConversation{reportId,reportStatus,context,messages}` and closed `ReportConversationMessage{sequence,role,content,createdAt}`. Reuse `ReportContextSnapshot` and the existing report status/role vocabulary; generated Go/TS types must not expose session/message/replay/anchor locator fields, `any`, optional aliases or compatibility list methods. Remove the list path/schema and regenerate client/server artifacts in the same batch.

### 18.3 Operation matrix and handoff

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getReportConversation` | `Reports/getReportConversation.json` with ready/non-ready/empty/hidden/fail-closed scenarios | `ReportConversationScreen` via generated client; report detail and safe ReportsScreen shortcut | backend-review owned-report lookup and strict mapper | existing `feedback_reports.session_id` + `practice_messages.seq_no`; no migration | none | `BDD.REPORT.CONVERSATION.001`, `BDD.REPORT.CONVERSATION.API.001`, extended `E2E.P0.099` |

Delete `PracticeSessions/listPracticeSessions.json` and every generated/runtime/consumer positive reference. 002 Phase 12 must prove fixture/example/Prism parity; 003 Phase 11 must generate and exact-match the OPENAPI-001 v1.7 oracle from the merge-base old baseline before re-freeze. Backend-practice owns deletion of list runtime surface; backend-review owns the replacement read model; frontend-report owns the reportId-only UI. BDD is downstream-owned; this contract phase closes with schema/inventory/codegen/zero-reference gates plus root `make test`.

## 19 OPENAPI-007 Settings UserContext pruning

### 19.1 RED and invariants

Require accepted [OPENAPI-007](../../decisions/OPENAPI-007-settings-user-context-pruning.md), spec D-39 and history 1.64 before source mutation. Focused schema/generated tests must fail while `UserContext.required/properties` or generated Go/TS types expose `uiLanguage` / `preferredPracticeLanguage` / `emailMasked`, or omit the required complete account `email`. Lock unchanged 37 operations / 10 tags and exact `getMe` / `completeMyProfile` / `deleteMe` method, path, operationId, status and security semantics.

### 19.2 GREEN minimal projection

Set `UserContext.additionalProperties: false` and make it a required four-field object: `id`, `email`, `displayName`, `profileCompletionRequired`. `email` is the complete authenticated account email shown in Settings. Update `openapi/openapi.yaml`, generated Go/TS, embedded backend schema and all typed builders in one batch. Delete `emailMasked`; do not make old fields optional, inject defaults or add aliases. Raw email must not be emitted to logs/E2E evidence or unauthenticated responses.

### 19.3 Operation matrix and handoff

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getMe` | `Auth/getMe.json` authenticated/profileIncomplete/unauthenticated | AppRuntimeProvider + Settings real account fields | backend-auth current-user handler | users/session；analytics opt-in remains internal runtime-config input | none | `BDD.SHELL.SETTINGS.001/.002` + extended `E2E.P0.101` |
| `completeMyProfile` | `Auth/completeMyProfile.json` | AuthProfileSetup + runtime refresh | backend-auth profile completion handler | users display/profile/terms fields | none | existing auth BDD + `E2E.P0.101` |
| `deleteMe` | unchanged `Auth/deleteMe.json` | Settings destructive action | backend-auth delete handoff | user soft delete/session revoke/privacy job | none | `BDD.SHELL.SETTINGS.DELETE.001` + backend contract |

002 Phase 13 updates Auth fixtures and mock parity；003 Phase 12 owns exact old-baseline findings and guarded re-freeze；backend-auth/001 Phase 10 removes old mapper/store fields；frontend-shell/001 consumes runtime user without a second `getMe`；B4 001 Phase 13 drops the four obsolete `user_settings` columns while retaining `analytics_opt_in`。Production OpenAPI/generated/backend/frontend/mock must have zero positive old-field references；accepted ADR/history/plans and explicit negative tests are allowed。
