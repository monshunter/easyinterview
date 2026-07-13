# OpenAPI v1 Contract Fixtures & Mock Source Checklist

> **版本**: 1.16
> **状态**: active
> **更新日期**: 2026-07-13

**关联计划**: [plan](./plan.md)

## 1 Fixture inventory and validation

- [x] 1.1 `openapi/fixtures/` 覆盖当前 10 tag / 37 operationId，一份 operationId 对应一份 JSON fixture。<!-- verified: 2026-07-10 method=make target=validate-fixtures fixtures=37 -->
- [x] 1.2 每份 fixture 的 `operationId` 与文件名一致，`scenarios.default` 必填且排在第一位；声明 requestBody 的 operation 带 `request.body`。
- [x] 1.3 `scripts/lint/validate_fixtures.py` 校验 operation coverage、request/response schema、response status、AI provenance、privacy allowlist / blacklist、UUIDv7 和 `tmp_` id rule。
- [x] 1.4 P0 export exceptions 固定：`requestPrivacyExport` 返回 `501 + PRIVACY_EXPORT_NOT_AVAILABLE`，`exportResume` 返回 `501 + RESUME_EXPORT_NOT_AVAILABLE`。

## 2 Prototype baseline sync

- [x] 2.1 `openapi/fixtures/PROTOTYPE_MAPPING.md` 声明 `ui-design/src/data.jsx` 到 operationId 的映射。
- [x] 2.2 `make sync-fixtures-from-prototype` 只写入受支持 fixture 的 `prototype-baseline` scenario，并在写入后执行 fixture validation。
- [x] 2.3 同步命令幂等；重复运行不产生新的 `openapi/fixtures` diff。
- [x] 2.4 P0 closed-loop endpoints 的 `prototype-baseline` scenario 非空且 schema-valid。

## 3 Example projection and Prism smoke

- [x] 3.1 `make render-openapi-fixture-examples` 从 fixtures 生成 `openapi/.generated/openapi-with-fixtures.yaml`，覆盖 37 个 operationId。<!-- verified: 2026-07-10 method=make target=render-openapi-fixture-examples -->
- [x] 3.2 生成的 OpenAPI named example body 与 fixture `scenarios.default.response.body` 字节级一致。
- [x] 3.3 Prism smoke 固定 matrix 校验 `getMe`、`listTargetJobs`、`getPracticeSession`、`getFeedbackReport`、`requestPrivacyExport` 的 response body 与 fixture body 字节级一致。
- [x] 3.4 OpenAPI 主文件不手写 response examples；mock / docs consumer 只消费 fixtures 或生成 examples。

## 4 Consumer contract and docs

- [x] 4.1 Mock consumer scenario 选择规则固定：显式 scenario 命中则使用；未指定时使用 `default`；指定不存在 scenario 时失败。
- [x] 4.2 前端 MSW、后端 mock server、Prism 和文档站必须共享 `openapi/fixtures/` 或生成 examples；需要新增 mock variant 时在 fixture scenario 中增加。
- [x] 4.3 `openapi/fixtures/README.md`、`openapi/README.md` 与本 owner docs 只描述当前 fixture truth source、命令和 consumer contract。
- [x] 4.4 本地 BDD 文件不适用；用户可见行为由当前 P0 scenario owner 验证，新增 Practice recovery 必须通过 Phase 9.6 downstream `BDD-Gate`。

## 5 Current owner compression gate

- [x] 5.1 `plan.md`、`checklist.md`、`context.yaml` 与 plans INDEX 对齐当前 37-operation fixture/mock-source contract。<!-- verified: 2026-07-10 method=context-validation+sync-doc-index target=openapi-v1-contract/002 -->
- [x] 5.2 通用 production-script inventory 先红后绿；删除无 entry point、caller 或 owner 引用的一次性 fixture bootstrap 记录，并通过 fixture/codegen/owner gates。
  <!-- verified: 2026-07-10 method=production-script-inventory-and-openapi-gates evidence="Expanded inventory red reported exactly one one-shot fixture bootstrap record. Deleted the unreferenced 27,754-byte production script without a placeholder. Green inventory passes; prototype sync is hash-idempotent; 37 fixtures validate; sync/render tests pass 9 tests and 16 subtests; example rendering, 10-tag/37-operation OpenAPI lint, and full scripts/lint 293 tests plus 4248 subtests pass." -->
- [x] 5.3 删除 fixture example renderer 中四个未读取的 path/method loop bindings，改为 value-only traversal；验证 AST RED/GREEN、renderer tests/idempotency、fixture/OpenAPI/codegen gates、owner contexts 与 docs/diff/pruning gates。
  <!-- verified: 2026-07-10 method=fixture-renderer-value-only-traversal evidence="Replaced two nested items traversals with values traversals and removed four unused key bindings. Renderer passes 5 tests plus 6 subtests; generated OpenAPI SHA is unchanged; 37 fixtures validate; prototype sync preserves the full fixture-tree hash; OpenAPI/codegen, both owner contexts and docs/diff/pruning gates PASS. Prism CLI was unavailable and is not claimed." -->

## 6 Practice round identity and progress fixtures

- [x] 6.1 RED: fixture validation/consumer tests reject missing paired round identity on current practice plans and missing progress projection on structured TargetJobs.<!-- verified: 2026-07-12 method=validator-red -->
- [x] 6.2 Add plan fixtures for baseline/current round and legacy null identity, plus TargetJob fixtures for zero/partial/all completed round states.<!-- verified: 2026-07-12 method=fixture-validation count=37 variants="not-started,partial,completed,legacy-null,mismatch" -->
- [x] 6.3 Update prototype mapping/data and prove `make sync-fixtures-from-prototype` idempotency without lifecycle-status round inference.<!-- verified: 2026-07-12 method=prototype-sync-twice+4-unit-tests+37-fixture-validation evidence="ui-design/src/data.jsx is the practiceProgress source; sync does not read TargetJob lifecycle status" -->
- [x] 6.4 Run `make validate-fixtures`, example rendering, mock consumer tests, and scenario-owner handoff gates.<!-- verified: 2026-07-12 method=57-python-tests+validate-fixtures+render-examples+generated-consumers -->

## 7 OPENAPI-001 report fixtures

- [x] 7.1 RED-GREEN: validator negatives inject old dimension/focus/question fields and unknown properties into each closed report object and must fail before fixture migration. Create-plan matrix also rejects baseline+sourceReportId; retry/next missing/null/blank/malformed sourceReportId; and every derived extra.
  <!-- verified: 2026-07-12 method=tdd-red-green evidence="Initial fixture validation failed 60 current-schema errors. Focused RED also proved bare const predicates incorrectly selected the ready branch and old prototype projection restored dimension/question fields. GREEN passes the 12-case request matrix, status-conditional branch test, closed-object/bounds negatives and canonical oversize alias rejection." -->
- [x] 7.2 Replace get/list report and create-plan scenarios with current direct shape, frozen context and no client focus input; include queued/generating/two ready/failed/failed-context-too-large/invalid/long-content variants plus valid baseline and minimal retry/next `{goal,sourceReportId}` requests. The oversized variant uses canonical B1 `REPORT_CONTEXT_TOO_LARGE` and fixture validation rejects aliases.
  <!-- verified: 2026-07-12 method=fixture-status-and-derived-matrix evidence="getFeedbackReport includes queued, generating, ready-needs-practice, ready-well-prepared, ready-empty-focus, failed, failed-context-too-large, invalid-contract and long-content direct bodies; list reports is direct; retry-derived/next-derived requests contain only goal+sourceReportId; all 37 fixtures validate." -->
- [x] 7.3 Update prototype data/mapping and run `make sync-fixtures-from-prototype` twice; the second run is byte-idempotent and cannot restore old fields.
  <!-- verified: 2026-07-12 method=prototype-sync-twice+unit-tests evidence="Two full fixture-tree SHA-256 manifests are identical; 6 sync tests pass including direct FeedbackReport projection and negative old-field assertions." -->
- [x] 7.4 Run `make validate-fixtures`, example rendering and Prism byte-equal smoke for `getFeedbackReport`, `listTargetJobReports` and `createPracticePlan`; pass exact response markers to backend/frontend owners.
  <!-- verified: 2026-07-12 method=fixture-example-prism evidence="validate-fixtures passes 37; renderer passes 5 tests and emits e4017fcf5a3a...; live Prism 5.14.2 passes 7/7 byte-equal checks including exact getFeedbackReport=200, listTargetJobReports=200 and createPracticePlan=201 defaults." -->

## Phase 8: OPENAPI-002 paste-only fixtures

- [ ] 8.1 RED: focused fixture/schema tests reject empty/space/tab/newline-only `rawText`, old source wrapper, URL/file/manual-form/title/company/extra request fields, TargetJob `sourceType/sourceUrl` responses and `target_job_attachment`; current positive fixtures fail before migration.
- [ ] 8.2 GREEN: make `importTargetJob` default/manual-text requests exactly `{rawText,targetLanguage,resumeId}` with non-whitespace text; add canonical `validation-blank-raw-text`=`422/VALIDATION_FAILED/retryable=false/details.field=rawText`, whose negative harness asserts the exact `/rawText` schema violation without skipping request validation; remove URL/file/manual-form positive scenarios and source response fields; preserve 37-operation fixture coverage.
- [ ] 8.3 GREEN: remove only TargetJob attachment purpose/scenarios from `createUploadPresign`; keep resume/privacy purpose coverage and the generic upload operation.
- [ ] 8.4 Update prototype data/mapping and run `make sync-fixtures-from-prototype` twice; second run is byte-identical and cannot restore old source fields or import variants.
- [ ] 8.5 Run `make validate-fixtures`, example rendering and Prism byte parity for import/list/get TargetJob plus upload presign; hand exact markers to mock/frontend/backend owners.
- [ ] 8.6 BDD/ZERO-REFERENCE-GATE: P0.010/P0.015 consume paste-only states; current positive fixture/prototype/example surfaces contain zero positive/runtime `TargetJobImportSource*|target_job_attachment|sourceType/sourceUrl|url/file/manual_form` import variants. ADR/oracle and exact negative declarations are allowed; whole-file/directory exclusions are forbidden.

## Phase 9: Practice message recovery fixtures

- [ ] 9.1 RED: fixture-validator tests reject user messages missing `clientMessageId/replyStatus`, assistant messages containing recovery fields, invalid status, duplicate retry messages and non-typed error bodies.
- [ ] 9.2 GREEN: `getPracticeSession` provides pending/retryable-failed/terminal-failed/complete projections with stable same user ID; only complete has exactly one assistant reply.
- [ ] 9.3 FAILURE-MATRIX: `sendPracticeMessage` includes exact validation 422, auth 401, not-found 404, pending-conflict 409, same-ID mismatch 409 and retryable AI-timeout 502 scenarios with locked code/retryable/details markers and reservation expectations.
- [ ] 9.4 REPLAY-GATE: paired retry-success uses the same `clientMessageId` and text after retryable failure, transitions the existing user projection to complete and creates exactly one assistant; mismatch/terminal cases never retry.
- [ ] 9.5 PARITY-GATE: validate fixtures, render examples and run Prism byte parity for `getPracticeSession` / `sendPracticeMessage`; mock runtime unknown-scenario behavior stays fail-loudly.
- [ ] 9.6 BDD-Gate: hand exact markers to mock-contract-suite/001, frontend-workspace-and-practice/002, backend-practice/002 and P0.046; scenario proof covers reload → same-ID retry with no duplicate user/assistant message.
