# OpenAPI v1 Contract Fixtures & Mock Source Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## 1 Fixture inventory and validation

- [x] 1.1 `openapi/fixtures/` 覆盖当前 10 tag / 35 operationId，一份 operationId 对应一份 JSON fixture。
- [x] 1.2 每份 fixture 的 `operationId` 与文件名一致，`scenarios.default` 必填且排在第一位；声明 requestBody 的 operation 带 `request.body`。
- [x] 1.3 `scripts/lint/validate_fixtures.py` 校验 operation coverage、request/response schema、response status、AI provenance、privacy allowlist / blacklist、UUIDv7 和 `tmp_` id rule。
- [x] 1.4 P0 export exceptions 固定：`requestPrivacyExport` 返回 `501 + PRIVACY_EXPORT_NOT_AVAILABLE`，`exportResume` 返回 `501 + RESUME_EXPORT_NOT_AVAILABLE`。

## 2 Prototype baseline sync

- [x] 2.1 `openapi/fixtures/PROTOTYPE_MAPPING.md` 声明 `ui-design/src/data.jsx` 到 operationId 的映射。
- [x] 2.2 `make sync-fixtures-from-prototype` 只写入受支持 fixture 的 `prototype-baseline` scenario，并在写入后执行 fixture validation。
- [x] 2.3 同步命令幂等；重复运行不产生新的 `openapi/fixtures` diff。
- [x] 2.4 P0 closed-loop endpoints 的 `prototype-baseline` scenario 非空且 schema-valid。

## 3 Example projection and Prism smoke

- [x] 3.1 `make render-openapi-fixture-examples` 从 fixtures 生成 `openapi/.generated/openapi-with-fixtures.yaml`，覆盖 35 个 operationId。
- [x] 3.2 生成的 OpenAPI named example body 与 fixture `scenarios.default.response.body` 字节级一致。
- [x] 3.3 Prism smoke 固定 matrix 校验 `getMe`、`listTargetJobs`、`getPracticeSession`、`getFeedbackReport`、`requestPrivacyExport` 的 response body 与 fixture body 字节级一致。
- [x] 3.4 OpenAPI 主文件不手写 response examples；mock / docs consumer 只消费 fixtures 或生成 examples。

## 4 Consumer contract and docs

- [x] 4.1 Mock consumer scenario 选择规则固定：显式 scenario 命中则使用；未指定时使用 `default`；指定不存在 scenario 时失败。
- [x] 4.2 前端 MSW、后端 mock server、Prism 和文档站必须共享 `openapi/fixtures/` 或生成 examples；需要新增 mock variant 时在 fixture scenario 中增加。
- [x] 4.3 `openapi/fixtures/README.md`、`openapi/README.md` 与本 owner docs 只描述当前 fixture truth source、命令和 consumer contract。
- [x] 4.4 BDD 不适用；本 plan 的用户可见行为由当前 P0 scenario owner 验证。

## 5 Current owner compression gate

- [x] 5.1 `plan.md`、`checklist.md`、`context.yaml` 与 plans INDEX 对齐当前 35-operation fixture/mock-source contract。
  <!-- verified: 2026-07-07 method=current-owner-compression evidence="Updated plan.md to v1.5, checklist.md to v1.4, and context specVersion to v1.35; sync-doc-index --fix-index updated openapi-v1-contract plans INDEX. PASS: targeted stale-wording grep returned no matches; validate_context.py openapi-v1-contract/002 contract PASS; make sync-fixtures-from-prototype PASS (5 fixtures populated); make render-openapi-fixture-examples PASS; make validate-fixtures PASS (35 fixtures); python3 -m unittest scripts.codegen.render_openapi_fixture_examples_test scripts.lint.validate_fixtures_test scripts.codegen.sync_fixtures_from_prototype_test PASS (32 tests); make lint-openapi PASS (10 tags, 35 operations); make codegen-check PASS." -->
