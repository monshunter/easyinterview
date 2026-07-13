# 001 - OpenAPI v1 Contract Bootstrap

> **版本**: 1.22
> **状态**: active
> **更新日期**: 2026-07-13

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

This owner plan remains the executable contract/codegen evidence index and is reopened for the accepted OPENAPI-002 paste-only correction.

## 2 Current Contract

| Surface | Current contract | Gate |
|---------|------------------|------|
| OpenAPI inventory | 10 tags, 37 operations, `/api/v1` prefix, session-cookie auth, public/protected operation security | `make lint-openapi`, inventory tests |
| Error envelope | B1 `ApiError` inner object + B2 `ApiErrorResponse` wire envelope | generator tests, codegen-check |
| Shared types | B1 enum/page/error conventions are reused; OpenAPI does not duplicate shared enum ownership | conventions drift and generated tests |
| Codegen | Go server/types and TS client/types are reproducible from `openapi/openapi.yaml` | `make codegen-openapi`, `make codegen-check` |
| Local docs | Redocly CLI renders `openapi/dist/index.html` without committing generated docs | `make docs-openapi` |
| Downstream handoff | 002 owns fixtures/mock source; 003 owns breaking-change baseline/gate; 004 owns resume additive coverage | plans INDEX and context validation |

## 3 Current Operation Inventory

| Tag | Operations |
|-----|------------|
| Auth | `getMe`, `completeMyProfile`, `deleteMe`, `startAuthEmailChallenge`, `verifyAuthEmailChallenge`, `logout`, `getRuntimeConfig` |
| Uploads | `createUploadPresign` |
| Resumes | `listResumes`, `registerResume`, `getResume`, `getResumeSource`, `updateResume`, `duplicateResume`, `archiveResume`, `exportResume` |
| TargetJobs | `importTargetJob`, `listTargetJobs`, `getTargetJob`, `updateTargetJob`, `archiveTargetJob` |
| PracticePlans | `createPracticePlan`, `getPracticePlan` |
| PracticeSessions | `listPracticeSessions`, `startPracticeSession`, `getPracticeSession`, `sendPracticeMessage`, `completePracticeSession`, `createPracticeVoiceTurn` |
| Reports | `getFeedbackReport`, `listTargetJobReports` |
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

本 plan 不新建本地 BDD 文件，因为它只拥有 API schema、codegen 与 contract gate；但 Practice recovery 是用户可见行为，不能只用内部 contract test 收口。Phase 14 必须以 frontend-workspace-and-practice/002 的 BDD 与 P0.046 failure/recovery scenario 作为 mandatory `BDD-Gate`，证明 optimistic user message、pending lock、reload projection、typed failure branching 与 same-ID retry。未取得该 handoff evidence 时，本 plan 不得恢复 completed。

## 7 Revision Log

| Date | Version | Change |
|------|---------|--------|
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
| `createPracticePlan` | `PracticePlans/createPracticePlan.json` round variants | shared parse/workspace/report start helper | backend-practice generated adapter/service/store | insert `practice_plans.round_id/round_sequence` + idempotency/audit | none | P0.022/P0.070/P0.072/P0.098 |
| `getPracticePlan` | `PracticePlans/getPracticePlan.json` current + legacy-null | shared start exact-pair reuse; Practice budget | backend-practice generated adapter/store | read nullable legacy/current plan identity | none | P0.022/P0.070/P0.098 |
| `listTargetJobs` | `TargetJobs/listTargetJobs.json` zero/partial/final progress | Home/Workspace cards and quick-start | backend-targetjob list handler/store/service projection | TargetJob summary + plans/sessions/completion events; no mutable progress column | none | P0.018/P0.098 |
| `getTargetJob` | `TargetJobs/getTargetJob.json` zero/partial/final progress | Parse/Report/current-round handoff | backend-targetjob get handler/store/service projection | same ledger projection as list | none after persisted JD parse | P0.057/P0.098 |

## 12 OPENAPI-001 grounded direct report contract

### 12.1 Closed wire shape

After OPENAPI-001 is accepted and B1 emits `REPORT_CONTEXT_TOO_LARGE_CONVENTIONS_PASS`, update the proposed `FeedbackReport` while the old baseline remains unchanged. Require nullable-until-state-resolved `summary` / `preparednessLevel` / `provenance`, non-null minimal `ReportContextSnapshot` with `hasNextRound`, code+label dimensions, dimensionCode evidence and `retryFocusDimensionCodes`. Close the report state machine: ready requires non-null summary/preparedness/provenance plus non-empty dimensions/actions; failed alone requires non-null errorCode and all other states require null errorCode. Keep CreatePracticePlanRequest as `type: object` with typed superset properties and conditional `oneOf`: baseline requires its existing non-focus fields and forbids sourceReportId; retry/next each require a non-null UUID sourceReportId and allow only goal+sourceReportId. Existing generators must produce Go/TS object types rather than `any`; schema/runtime fixtures reject invalid report states, baseline source, derived missing/null/blank source and every derived extra, while valid ready/failed and minimal retry/next requests pass. Delete `DimensionResult` and old dimension/focus names, close objects and apply ADR bounds. Synchronize `REPORT_CONTEXT_TOO_LARGE` into `ApiErrorCode` only through the B1 source. Run 003 base-ref exact audit across breaking and additive findings before codegen/fixtures.

Under approved方案 A，`ReportNextAction.label.maxLength=200` code points is only a malformed-output fuse。F3/runtime/frontend enforce English `<=24` whitespace words / zh-CN `<=64` Unicode code points；targeted action-label repair uses an internal 18/52 generation margin and is revalidated against 200+24/64。P0.099 proves desktop+390 wrapping and typed-invalid/no-raw behavior。Old baseline clean PASS is invalid until new audit/re-freeze；codegen-check remains separately pending。

Generation/judge max4，attempt_count/retry_count/reason/scope and business/outbox backoff are internal runtime/eval contracts。Do not add HTTP fields、progress or retry-generation operation。`FeedbackReport` remains queued/generating/ready/failed；frontend maxAttempts49 exhaustion is a client continue-check state，not an API failed transition。This decision creates no additional expected finding beyond the existing maxLength200 correction。

### 12.2 Generated artifacts and inventory

Regenerate Go/TS artifacts and add exact inventory/negative tests that reject old report fields and unknown object properties. The operation/tag inventory remains 37/10; no endpoint or status changes.

### 12.3 Handoff

002 must replace every Reports/PracticePlans scenario and prototype projection before backend/frontend consumers compile. 003 must preserve the merge-base breaking artifact before current baseline re-freeze.

## 13 OPENAPI-002 TargetJob paste-only contract

### 13.1 RED contract boundary

After [OPENAPI-002](../../decisions/OPENAPI-002-targetjob-paste-only.md) is accepted, keep the merge-base baseline untouched and add focused schema/inventory tests that require `ImportTargetJobRequest` to be a closed object with exactly required `rawText` / `targetLanguage` / `resumeId`. `rawText` must declare `minLength: 1` and `pattern: '\S'`; positive cases include non-whitespace text and negative cases include empty/space/tab/newline-only values. Negative cases must also reject the old `source` wrapper, every URL/file/manual-form discriminator payload, `fileObjectId`, `titleHint`, `companyNameHint` and arbitrary extra properties. Read-side tests must reject `TargetJob.sourceType` / `sourceUrl`; upload tests must reject `purpose=target_job_attachment` while proving `createUploadPresign` still supports resume/privacy consumers.

### 13.2 GREEN source, codegen and freeze invariant

Delete all five `TargetJobImportSource*` schemas, flatten and constrain `ImportTargetJobRequest`, remove TargetJob source provenance properties and remove only the TargetJob attachment enum value. Regenerate Go/TS artifacts and assert no compatibility union, discriminator, alias or legacy generated type remains. `importTargetJob` remains operationId `importTargetJob`, `POST /api/v1/targets/import`, `202 + TargetJobWithJob`; `createUploadPresign` remains operationId `createUploadPresign`, `POST /api/v1/uploads/presign`, `201 + UploadPresign`; inventory remains 37 operations / 10 tags. The current baseline cannot be edited until 003 Phase 6 preserves the exact old-baseline audit and all downstream gates pass.

### 13.3 Operation matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `importTargetJob` | `TargetJobs/importTargetJob.json` paste-only default/manual-text | Home paste submit and Parse polling | backend-targetjob generated adapter/service/runner | TargetJob + async job; no URL/file source provenance | parse pasted JD text | P0.010/P0.015 paste-only |
| `createUploadPresign` | `Uploads/createUploadPresign.json` resume/privacy only | Resume Workshop/privacy consumers; no Home consumer | backend-upload generated adapter/service | file object for remaining purposes only | none | existing resume/privacy scenarios; no TargetJob attachment |
| `listTargetJobs` / `getTargetJob` | TargetJobs list/get without source fields | Home/Workspace/Parse/Report | backend-targetjob list/get adapters | current TargetJob projection without source provenance | none after persisted parse | P0.010/P0.015/P0.018/P0.057 |

### 13.4 Zero-reference and handoff gates

Run exact current-contract searches over OpenAPI, generated artifacts, positive fixtures, frontend/backend runtime consumers, mock runtime and active positive scenarios. Positive/runtime occurrences of `TargetJobImportSource*`, `target_job_attachment`, TargetJob `sourceType/sourceUrl`, URL/file/manual-form import branches or compatibility aliases must be zero. Accepted ADR、expected-findings oracle 与显式 negative test/fixture declaration 可包含被拒绝 token；gate 必须按语义/可达面区分正负引用，禁止用整文件或整测试目录排除。BDD assets are owned downstream: P0.010/P0.015 must prove the paste-only main/failure paths, while URL/file/manual-form positive scenarios are removed rather than retained as compatibility coverage.

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
| `getPracticeSession` | `PracticeSessions/getPracticeSession.json`: pending/retryable-failed/terminal-failed/complete + reload recovery | Practice transcript hydration, composer lock and retry affordance | backend-practice get adapter/service/store | user `client_message_id` + durable reply status + ordered assistant projection | none on read | P0.044 + P0.046 |
| `sendPracticeMessage` | `PracticeSessions/sendPracticeMessage.json`: default plus validation/auth/not-found/conflict/mismatch/retryable failure | optimistic row, thinking state and typed retry dispatch | backend-practice generated adapter/service/store | reserve once by `(session_id,client_message_id)`; transition reply status; at-most-one assistant | interviewer generation | P0.044 + P0.046 |

### 14.5 Contract audit and downstream handoff

This union/persistence correction changes existing Practice message validation semantics. 003 must audit old-baseline → proposed findings separately from OPENAPI-002（不得并入其 exact 15 allowset），preserve the artifact before baseline mutation, then wait for 002 fixture matrix、mock-contract-suite parity、backend-practice persistence、frontend-workspace-and-practice typed consumer and P0.046 before re-freeze. Operation/tag inventory stays 37/10；no retry endpoint, client-side business-state store or compatibility message schema is allowed.
