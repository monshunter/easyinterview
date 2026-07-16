# OpenAPI v1 Contract Fixtures & Mock Source

> **版本**: 1.23
> **状态**: completed
> **更新日期**: 2026-07-16

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

维护 `openapi/fixtures/` 作为当前 HTTP mock 数据的唯一真理源：当前 10 个 tag / 38 个 operationId 必须各有一份 fixture，`default` scenario 覆盖规范响应，其它 named scenario 由对应 consumer owner 在同一 fixture 中维护，fixtures 再投影为 Prism / 文档站消费的 OpenAPI named examples。

本 plan 只拥有 fixture 数据、fixture validator、prototype sync、fixture example render、Prism byte-equal smoke 和对应文档。正式 mock server 运行壳、前端 MSW runtime、后端 handler、OpenAPI schema 变更与 breaking-change policy 分别归对应 owner；它们只能消费这里的 fixture truth source，不在这里重建第二份 example。

当前 Phase 14 additive 增加 failed report regeneration fixture，inventory 变为 38/38；既有 Phase 12/13 gates 保持有效，不因本次追加而跳过。

## 2 当前合同

- `openapi/fixtures/<tag>/<operationId>.json` 当前必须覆盖 `openapi/openapi.yaml` 的 38 个 operationId，目录 tag 顺序跟随 OpenAPI spec。
- 每个 fixture 必须包含 `scenarios.default`，并且该 key 是 `scenarios` 的第一项。声明 requestBody 的 operation 必须给出 `request.body`；header-only idempotent operation 可只给 `request.headers`。
- `response.status` 必须是 operation 声明的状态码，或被 `default` error response 覆盖。`requestPrivacyExport` 固定返回 `501 + PRIVACY_EXPORT_NOT_AVAILABLE`；`exportResume` 固定返回 `501 + RESUME_EXPORT_NOT_AVAILABLE`。
- 所有 scenario 的 request/response body 必须按 `openapi/openapi.yaml` schema 校验通过。AI 生成相关 schema 必须带非空 `provenance`；隐私字段只能使用保留域名、保留电话号码和通用公司名；所有 UUID 字段使用 UUIDv7 字面量；`tmp_` id 直接失败。
- named scenarios 由 fixture consumer owner 直接维护；不从正式前端源码或已删除的 UI Demo 反向生成 mock 数据。
- `make render-openapi-fixture-examples` 从 fixtures 生成 `openapi/.generated/openapi-with-fixtures.yaml`。OpenAPI 主文件不得手写 response examples；Prism smoke 只使用生成物。

## 3 质量门禁分类

- **Plan 类型**: `contract + tooling + mock-source`
- **TDD 策略**: fixture coverage、schema validation、provenance、privacy allowlist、UUIDv7 / `tmp_` id scan、prototype sync idempotency、example projection 和 Prism byte-equal smoke 是可执行断言。重进本 plan 时必须先运行对应 gate 暴露 drift，再最小修复 fixture 或工具。
- **BDD 策略**: 不适用。本 plan 只交付内部 fixture/mock data truth source，不创建 BDD 文件或引用场景编号。用户行为由 domain owner 独立验收；fixture markers 只用于 contract/consumer tests。
- **替代验证 gate**: `make validate-fixtures`、`make render-openapi-fixture-examples`、`python3 scripts/codegen/prism_fixture_smoke.py`、fixture render/unit tests、`make lint-openapi`、`make codegen-check`、`sync-doc-index --check`。

## 4 交付范围

### 4.1 Fixture inventory and validation

`openapi/fixtures/` 在 Phase 14 前保有 37 个 JSON fixture；Phase 14 增加 `regenerateFeedbackReport` 后当前目标为 38 个，并始终与 OpenAPI operationId 一一对应。`scripts/lint/validate_fixtures.py` 负责以下检查：

- fixture 文件名、`operationId` 字段和 OpenAPI operationId 一致。
- 所有 operationId 都有 fixture，且没有 OpenAPI 不认识的 fixture。
- `default` scenario 必填且排在第一位，额外 scenario 同样校验 schema。
- request/response body 按 operation 的 requestBody 与 response schema 校验。
- AI schema 的 `provenance` 字段非空。
- 隐私 allowlist、黑名单、UUIDv7 和 `tmp_` id rule 通过。

### 4.2 Named scenario ownership

`openapi/fixtures/<tag>/<operationId>.json` 同时承载 `default` 与 consumer-owned named scenarios。所有 scenario 使用同一 validator，不得从前端源码或第二套数据集同步。

### 4.3 Example projection and Prism smoke

`make render-openapi-fixture-examples` 把每个 fixture 的 `scenarios.default.response.body` 投影到 `openapi/.generated/openapi-with-fixtures.yaml`。Phase 14 后投影必须覆盖 38 个 operationId，并保证 named example body 与 fixture body 字节级一致。

Prism smoke 使用生成物启动本地 mock，并用固定 operation matrix 校验响应 body 与 fixture body 字节级一致。当前固定 matrix 包括 `getMe`、`listTargetJobs`、`getPracticeSession`、`getFeedbackReport`、`requestPrivacyExport`。

### 4.4 Consumer contract

Mock consumer 的 scenario 选择规则固定为：

1. 请求显式指定 scenario 时，命中则使用该 scenario。
2. 未指定 scenario 时使用 `default`。
3. 指定了 fixture 未声明的 scenario 时失败，不静默回退。

前端 MSW、后端 mock server、Prism 和文档站都必须消费 `openapi/fixtures/` 或由它生成的 OpenAPI examples。需要新增 mock variant 时，应在这里新增 scenario，并通过 validator 与 consumer gate。

### 4.5 Script inventory

当前 fixture 工具面只保留可执行、可重复验证的 validator、prototype sync、example renderer 与 Prism smoke。没有当前入口或 owner 引用的一次性 bootstrap 记录不属于 fixture truth source，也不作为历史说明文件保留。

## 5 验收标准

- `openapi/fixtures/` 覆盖当前 38 个 operationId，没有多余 operation fixture。
- `make validate-fixtures` 通过，并能拒绝缺 fixture、schema drift、缺 provenance、非保留隐私值、非 UUIDv7 id 和 `tmp_` id。
- 所有 consumer-owned named scenarios 非空并 schema-valid，且不存在平行 mock 数据真理源。
- `make render-openapi-fixture-examples` 通过，生成 examples 与 fixture body 字节级一致。
- Prism smoke 固定 matrix 通过，其中 `requestPrivacyExport` 返回 `501 + PRIVACY_EXPORT_NOT_AVAILABLE`。
- `openapi/fixtures/README.md`、`openapi/README.md` 和本 owner docs 均只描述当前 fixture truth source 与 consumer contract。
- `scripts/` 下的生产工具必须具有当前 entry point、caller 或 owner 引用；一次性 fixture bootstrap 记录不得留在生产脚本目录。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| Consumer scenario 与 OpenAPI schema 字段不同步 | validator fail-fast，并要求在当前 fixture owner 中修正；不在 consumer 中静默改名 |
| fixture 手写 response 漂出 schema | `make validate-fixtures` 校验所有 scenario，fixture edit 必须伴随 validator 通过 |
| privacy export 被误写成成功响应 | validator 对 `requestPrivacyExport` 固定检查 `501 + PRIVACY_EXPORT_NOT_AVAILABLE` |
| AI provenance 被写成空值 | validator 强制 provenance 字段存在且非空 |
| consumer 私自复制 mock body | consumer owner 必须引用本目录或生成 examples；新增 variant 只能通过 fixture scenario 增加 |

## 7 Practice round fixture projection

- `createPracticePlan` / `getPracticePlan` fixtures must include paired `roundId + roundSequence` for current records and a legacy-null negative scenario that is never reusable.
- `listTargetJobs` / `getTargetJob` fixtures must include `practiceProgress` for not-started, partially completed, and all-completed rounds; `completedRounds` is ordered/deduplicated and final `currentRound` is null.
- round-related named scenarios must project the same current/completed semantics and must not derive a round from TargetJob lifecycle `status`.

| operationId | fixture scenarios | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|-------------------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticePlan` | baseline round, derived round, mismatch request | shared start helper | backend-practice | normalized plan pair + IK/audit | none | fixture + generated consumer tests |
| `getPracticePlan` | current pair, legacy null pair | exact plan reuse / Practice budget | backend-practice | nullable read compatibility | none | fixture + generated consumer tests |
| `listTargetJobs` | not-started, partial, completed | Home/Workspace rail + quick-start | backend-targetjob | completion-ledger projection | none | fixture + generated consumer tests |
| `getTargetJob` | not-started, partial, completed | Parse/Report current-round gate | backend-targetjob | completion-ledger projection | none after JD parse | fixture + generated consumer tests |

## 9 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-15 | 1.22 | Revise Phase 13 Auth fixtures to the complete authenticated `email`, remove `emailMasked`, and keep raw email out of logs/evidence. | OPENAPI-007 |
| 2026-07-15 | 1.21 | Add Phase 13 for OPENAPI-007 four-field UserContext Auth fixtures, dev-mock parity and old-language-field negative gates. | OPENAPI-007 |
| 2026-07-15 | 1.20 | Add Phase 12 for report-owned conversation fixtures and removal of the public PracticeSessions list fixture while keeping 37/37 parity. | OPENAPI-001 v1.7 |
| 2026-07-14 | 1.19 | Add Phase 11 for OPENAPI-005 summary-only list fixture, full detail fixture and Prism/mock consumer handoff. | OPENAPI-005 |
| 2026-07-14 | 1.16 | Reopen for OPENAPI-004 canonical-round report overview fixtures, prototype projection, Prism parity and latest-report-pointer removal. | OPENAPI-004 |
| 2026-07-13 | 1.15 | Add canonical blank-rawText 422 validation fixture and Practice reload/same-ID recovery plus typed failure fixture matrix. | openapi-v1-contract 1.54 |
| 2026-07-13 | 1.14 | Reopen fixture owner for OPENAPI-002 paste-only TargetJob requests/responses, upload purpose cleanup and runtime projection gates. | OPENAPI-002 + 001/003 + mock-contract-suite/001 |
| 2026-07-12 | 1.13 | Add exact baseline/derived CreatePracticePlanRequest positive and negative fixture matrix. | openapi-v1-contract 1.46 |
| 2026-07-12 | 1.12 | Add the canonical REPORT_CONTEXT_TOO_LARGE failed-report scenario sourced from B1. | openapi-v1-contract 1.45 + shared-conventions 1.30 |
| 2026-07-12 | 1.11 | Reopen Phase 7 for OPENAPI-001 closed Reports/PracticePlans fixtures, prototype projection and Prism proof. | openapi-v1-contract 1.44 |
| 2026-07-12 | 1.10 | Reopen fixture owner for practice round identity and TargetJob practice progress scenarios. | openapi-v1-contract 1.43 |
| 2026-07-10 | 1.9 | 删除 fixture example renderer 中未读取的 path/method 遍历绑定。 | tech-debt pruning |
| 2026-07-10 | 1.8 | 删除无当前入口的一次性 fixture bootstrap 记录，并将生产脚本可达性纳入通用 inventory gate。 | product-scope/001-core-loop-module-pruning |
| 2026-07-10 | 1.7 | 对齐当前 37-operation fixture truth source，包含 `archiveTargetJob`。 | tech-debt pruning |
| 2026-07-07 | 1.6 | 新增 `getResumeSource` fixture，fixture truth source 与 example projection 覆盖当时 36-operation contract。 | backend-resume/001 Phase 12 |
| 2026-07-07 | 1.5 | 压缩 owner 文档为当时 fixture truth source、prototype sync、example projection and Prism smoke contract。 | product-scope/001-core-loop-module-pruning |
| 2026-05-04 | 1.4 | 补齐质量门禁分类。 | docs-only L1 review |
| 2026-05-03 | 1.3 | 刷新 fixture / example coverage 与 prototype-baseline endpoint 范围。 | product-scope v1.2 / openapi-v1-contract v1.9 |
| 2026-05-03 | 1.2 | 调整 fixture coverage、报告字段与 prototype mapping。 | openapi-v1-contract v1.9 |
| 2026-04-29 | 1.1 | 补齐 `Auth/deleteMe` fixture 与 operation coverage gate。 | plan review |

## 8 OPENAPI-001 report fixture migration

Replace every `getFeedbackReport` scenario with queued/generating/ready-needs-practice/ready-well-prepared/failed/failed-context-too-large/invalid-contract/long-content current-shape variants. Every body includes frozen minimal context; ready variants include summary, code+label dimensions, dimensionCode evidence, actions and report-local focus. The oversized-context failed variant uses exactly B1 `REPORT_CONTEXT_TOO_LARGE`; no local alias is permitted. Replace `listTargetJobReports` and `createPracticePlan` focus-input scenarios, then validate and render the fixture tree.

Fixture/schema negative tests must prove old `dimension`, `retryFocusCompetencyCodes`, question fields and arbitrary additional properties fail. Render examples and run Prism byte-equal smoke for both Reports operations plus createPracticePlan.

## 10 OPENAPI-002 paste-only fixtures and projection

### 10.1 TargetJob request/response fixtures

- Replace `importTargetJob` `default` and `paste-primary` requests with the flattened exact body `{rawText,targetLanguage,resumeId}` and non-whitespace `rawText`. Add canonical negative scenario `validation-blank-raw-text`: whitespace-only request, `422`, `ApiErrorResponse.error.code=VALIDATION_FAILED`, `retryable=false`, `details.field=rawText`.
- Extend the fixture validator with an exact negative-request assertion for this one scenario: it must fail at `/rawText` because of the non-whitespace rule while the 422 response validates normally. A generic “skip request validation” flag, wildcard pointer or file-level exemption is forbidden.
- Delete URL, file and manual-form positive scenarios; do not rename them into historical compatibility cases.
- Remove `sourceType` / `sourceUrl` from every TargetJob response scenario, prototype projection and generated example. Parsed-ready, cross-user hidden and invalid-transition coverage remains, but source provenance is no longer part of the wire shape.
- Update `createUploadPresign` fixtures to remove `target_job_attachment` while preserving resume/privacy purpose coverage and the operation itself.

### 10.2 TDD negative matrix

Focused validator tests must first fail on current fixtures, then prove old `source` wrapper, URL, `fileObjectId`, manual-form fields, `titleHint`, `companyNameHint`, unknown properties, TargetJob source response fields and `purpose=target_job_attachment` are schema-invalid. Construct negative payloads in test code or dedicated negative fixtures; do not keep retired variants as positive named scenarios.

### 10.3 Prototype, examples and runtime handoff

Update the TargetJob fixture scenarios directly so they can only express paste-only requests/responses. Validate the complete fixture tree, render examples and run Prism byte-equal smoke for `importTargetJob`, `listTargetJobs`, `getTargetJob` and `createUploadPresign`; hand the exact fixture markers to mock-contract-suite/001 and frontend/backend consumers.

### 10.4 BDD and zero-reference gates

Fixtures provide paste success/failure contract states；this plan does not prove user flow. Current positive fixture/prototype/example surfaces must have zero positive/runtime `TargetJobImportSource*`, `target_job_attachment`, TargetJob `sourceType/sourceUrl`, URL/file/manual-form requests or compatibility aliases. Accepted ADR/oracle and exact negative test/fixture declarations may retain rejected tokens；searches must classify positive/runtime reachability and may not exclude a whole file or test directory. Inventory remains exactly 37 fixtures for 37 operations.

## 11 Practice message recovery fixtures

### 11.1 Authoritative get-session projections

`PracticeSessions/getPracticeSession.json` must add schema-valid named scenarios for `reply-pending`、`reply-retryable-failed`、`reply-terminal-failed` 与 `reply-complete`。All user messages carry stable `clientMessageId` + exact `replyStatus`; assistant messages carry neither. The scenario family must share one session/user message identity so tests can prove reload changes only authoritative reply state, not fabricate a new message. Exact invariants:

- pending/retryable/terminal contain zero assistant reply for that user message；
- complete contains exactly one assistant reply after that user message；
- no scenario derives status merely from assistant absence；the explicit persisted field is the assertion source；
- a reload marker identifies the retryable-failed user message that downstream consumer tests retry with the same ID/text.

### 11.2 Planned send failure matrix

`PracticeSessions/sendPracticeMessage.json` keeps `default` as 200 complete success and adds this exact planned matrix:

| scenario | status | error code | retryable | exact marker / persistence expectation |
|----------|--------|------------|-----------|----------------------------------------|
| `validation-empty-text` | 422 | `VALIDATION_FAILED` | false | `details.field=text`; no user reservation |
| `auth-unauthorized` | 401 | `AUTH_UNAUTHORIZED` | false | no user reservation |
| `session-not-found` | 404 | `PRACTICE_SESSION_NOT_FOUND` | false | no user reservation |
| `reply-pending-conflict` | 409 | `PRACTICE_SESSION_CONFLICT` | false | existing pending reservation unchanged |
| `client-message-mismatch` | 409 | `IDEMPOTENCY_KEY_MISMATCH` | false | same ID with different text; existing reservation unchanged |
| `ai-timeout-retryable` | 502 | `AI_PROVIDER_TIMEOUT` | true | user reservation persisted as `retryable_failed`; no assistant |

Every scenario must use `ApiErrorResponse` and lock status/code/retryable/details without parsing `message`. Add a paired retry-success scenario or cross-fixture contract that sends the same `clientMessageId` + same text after `ai-timeout-retryable`, transitions the existing user projection to `complete`, and yields exactly one assistant reply. Unknown scenario selection remains fail-loudly.

### 11.3 Validator, projection and BDD handoff

Focused fixture tests first RED on missing role-specific fields, assistant recovery fields, invalid reply enum, duplicate user/assistant IDs, wrong failure marker and retry success that allocates a second message. GREEN then runs fixture validation, example rendering and Prism byte parity for `getPracticeSession` and `sendPracticeMessage` without adding operations or tags. Hand exact status/body markers to mock-contract-suite/001, frontend-workspace-and-practice/002 and backend-practice/002；local fixture tests do not substitute for user-flow evidence。

## 12 OPENAPI-004 TargetJob report overview fixtures

### 12.1 Canonical scenario matrix

Replace `Reports/listTargetJobReports.json` flat full-report pages with canonical-round overview scenarios: all rounds empty; current ready only; old current ready plus newer queued/generating/failed latest attempt; latest ready shared by current/latest; duplicate ready tie-break fixtures; cross-user/not-found; invalid/missing frozen context fail-closed. Every round uses `PracticeRoundRef`, every nullable field is explicitly null when absent, and no full report/provenance/model/rubric/session/plan/pagination field appears.

### 12.2 TargetJob/prototype sync

Remove `latestReportId` from all TargetJob fixtures and prototype sync logic. Map `frontend/src` to the plan-detail report section without recreating a TargetJob pointer; canonical display names remain sourced from TargetJob summary while the overview supplies only round identity/current/latest state. Run sync twice and require byte-idempotency.

### 12.3 Parity and handoff

Validate fixtures, render examples and run live Prism byte parity for `listTargetJobReports` plus affected list/get TargetJob defaults. Pass exact markers to backend-review/frontend-report/mock owners；Parse retains only the unchanged entry negative contract. Positive/runtime fixtures must have zero cursor/pageInfo/full report/latestReportId compatibility fields.

## 13 Phase 11: OPENAPI-005 Resume list summary fixtures

### 11.1 Summary-only list fixture

Replace every `Resumes/listResumes.json` item with the exact closed nine-field projection `id/title/displayName/language/sourceType/parseStatus/summaryHeadline/hasReadableContent/updatedAt`. Cover upload/paste, queued/processing/ready/failed, nullable headline and both readable-content boolean values without copying raw/source/profile/provenance detail. Cases must make the projection semantics observable: whitespace-only headline/body and empty `structured_profile` map to null/false, while a trim-nonempty snapshot/original or nonempty-object profile maps to true; `fileObjectId`、`sourceType`、`parseStatus` alone never map to true. Focused fixture tests must RED on any omitted required field, extra property or old full `Resume` item.

### 11.2 Full detail fixture and validator split

Keep `Resumes/getResume.json` on the complete `Resume` contract, including source/body/structured profile/provenance cases required by detail consumers. Update validator provenance routing so full `Resume` remains checked while `listResumes.items[*].structuredProfile.provenance` is no longer expected. Fixture/schema tests must prove list and detail cannot be substituted for each other.

### 11.3 Example, Prism, mock and BDD handoff

Render examples and run byte-equal Prism/mock parity for both `listResumes` and `getResume`; inventory remains 37 fixtures / 37 operations. Hand the summary/full markers to backend-resume, all generated frontend consumers and mock-contract-suite in the same batch. No fixture compatibility scenario or frontend detail-fetch fallback is allowed；阶段收口执行根 `make test`。

## 14 Phase 12: Report-owned conversation fixtures

### 12.1 One-for-one fixture replacement

Delete `openapi/fixtures/PracticeSessions/listPracticeSessions.json` and add `openapi/fixtures/Reports/getReportConversation.json`, preserving exactly 37 default fixtures for 37 operations. No archived, disabled or compatibility session-list scenario remains. Focused fixture tests RED on the old path/operation/schema or any missing new fixture.

### 12.2 Closed conversation scenario matrix

Cover ready, queued/generating/failed with an existing owned report row, empty messages, Markdown/GFM content, hidden cross-user/not-found 404, and fail-closed identity/role/sequence/binding cases. Successful bodies contain only `reportId/reportStatus/context/messages`; messages contain only `sequence/role/content/createdAt` and are strictly increasing. No session/message/replay/anchor locator is allowed.

### 12.3 Example, Prism, mock and downstream handoff

Render examples and run live byte-equal Prism/mock parity for `getReportConversation`; prove the deleted list operation cannot be selected by path or scenario. Hand exact markers to backend-review, frontend-report-dashboard, mock-contract-suite and extended `E2E.P0.099`. BDD behavior remains downstream-owned; this fixture phase closes with validation, rendering, Prism parity, zero-reference and root `make test`.

## 15 Phase 13: OPENAPI-007 Settings UserContext fixtures

Update `Auth/getMe.json` authenticated/profileIncomplete responses and `Auth/completeMyProfile.json` success to exact `{id,email,displayName,profileCompletionRequired}`. The authenticated fixture uses a complete reserved-domain synthetic email under the existing fixture privacy allowlist because it is the contract value rendered by Settings；unauthenticated errors and all logs/E2E evidence remain email-free or redacted. Validator mutations must exercise the source `additionalProperties: false` closure and reject either old language field, `emailMasked`, any arbitrary extra field, a missing required field or a non-reserved real email domain.

Regenerate OpenAPI examples and prove Prism/dev-mock byte parity for `getMe` and `completeMyProfile`. Frontend dev mock and typed builders must consume the same fixture projection；no fallback constants or compatibility scenarios. This phase has no independent BDD: it hands real account values to frontend-shell settings BDD and the existing `E2E.P0.101` extension.

## 16 Phase 14: Failed report regeneration fixtures

Add `Reports/regenerateFeedbackReport.json` with `default`, replay, hidden 404, invalid state, active-job not-ready, context-too-large and idempotency-mismatch scenarios. Every success returns the same path `reportId` and a matching queued report-generation job；header-only requests carry `Idempotency-Key` and no body. Validate all 38 fixtures, render examples, extend Prism/dev-mock operation parity and prove raw report/provider content is absent. User behavior remains in backend/frontend report BDD；this phase is fixture/codegen support only.
