# OpenAPI v1 Contract Fixtures & Mock Source

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

维护 `openapi/fixtures/` 作为当前 HTTP mock 数据的唯一真理源：当前 10 个 tag / 35 个 operationId 必须各有一份 fixture，`default` scenario 覆盖规范响应，`prototype-baseline` scenario 由 `ui-design/src/data.jsx` 同步，fixtures 再投影为 Prism / 文档站消费的 OpenAPI named examples。

本 plan 只拥有 fixture 数据、fixture validator、prototype sync、fixture example render、Prism byte-equal smoke 和对应文档。正式 mock server 运行壳、前端 MSW runtime、后端 handler、OpenAPI schema 变更与 breaking-change policy 分别归对应 owner；它们只能消费这里的 fixture truth source，不在这里重建第二份 example。

## 2 当前合同

- `openapi/fixtures/<tag>/<operationId>.json` 当前必须覆盖 `openapi/openapi.yaml` 的 35 个 operationId，目录 tag 顺序跟随 OpenAPI spec。
- 每个 fixture 必须包含 `scenarios.default`，并且该 key 是 `scenarios` 的第一项。声明 requestBody 的 operation 必须给出 `request.body`；header-only idempotent operation 可只给 `request.headers`。
- `response.status` 必须是 operation 声明的状态码，或被 `default` error response 覆盖。`requestPrivacyExport` 固定返回 `501 + PRIVACY_EXPORT_NOT_AVAILABLE`；`exportResume` 固定返回 `501 + RESUME_EXPORT_NOT_AVAILABLE`。
- 所有 scenario 的 request/response body 必须按 `openapi/openapi.yaml` schema 校验通过。AI 生成相关 schema 必须带非空 `provenance`；隐私字段只能使用保留域名、保留电话号码和通用公司名；所有 UUID 字段使用 UUIDv7 字面量；`tmp_` id 直接失败。
- `prototype-baseline` 只由 `make sync-fixtures-from-prototype` 写入。源数据来自 `ui-design/src/data.jsx`，映射关系写在 `openapi/fixtures/PROTOTYPE_MAPPING.md`；手工改该 scenario 会被下一次同步覆盖。
- `make render-openapi-fixture-examples` 从 fixtures 生成 `openapi/.generated/openapi-with-fixtures.yaml`。OpenAPI 主文件不得手写 response examples；Prism smoke 只使用生成物。

## 3 质量门禁分类

- **Plan 类型**: `contract + tooling + mock-source`
- **TDD 策略**: fixture coverage、schema validation、provenance、privacy allowlist、UUIDv7 / `tmp_` id scan、prototype sync idempotency、example projection 和 Prism byte-equal smoke 是可执行断言。重进本 plan 时必须先运行对应 gate 暴露 drift，再最小修复 fixture 或工具。
- **BDD 策略**: BDD 不适用。本 plan 只交付内部 mock data truth source，不产生用户行为流；用户可见流程由当前 P0 scenario owner 维护。
- **替代验证 gate**: `make validate-fixtures`、`make sync-fixtures-from-prototype`、`make render-openapi-fixture-examples`、`python3 scripts/codegen/prism_fixture_smoke.py`、fixture render/unit tests、`make lint-openapi`、`make codegen-check`、`sync-doc-index --check`。

## 4 交付范围

### 4.1 Fixture inventory and validation

`openapi/fixtures/` 当前保有 35 个 JSON fixture 文件，和 OpenAPI operationId 一一对应。`scripts/lint/validate_fixtures.py` 负责以下检查：

- fixture 文件名、`operationId` 字段和 OpenAPI operationId 一致。
- 所有 operationId 都有 fixture，且没有 OpenAPI 不认识的 fixture。
- `default` scenario 必填且排在第一位，额外 scenario 同样校验 schema。
- request/response body 按 operation 的 requestBody 与 response schema 校验。
- AI schema 的 `provenance` 字段非空。
- 隐私 allowlist、黑名单、UUIDv7 和 `tmp_` id rule 通过。

### 4.2 Prototype sync

`openapi/fixtures/PROTOTYPE_MAPPING.md` 声明 `ui-design/src/data.jsx` 到 operationId 的映射。`make sync-fixtures-from-prototype` 只更新受支持 fixture 的 `prototype-baseline` scenario，并在写入后执行 fixture validation。该命令必须幂等：重复运行不会制造新的 fixture diff。

### 4.3 Example projection and Prism smoke

`make render-openapi-fixture-examples` 把每个 fixture 的 `scenarios.default.response.body` 投影到 `openapi/.generated/openapi-with-fixtures.yaml`。投影必须覆盖 35 个 operationId，并保证 named example body 与 fixture body 字节级一致。

Prism smoke 使用生成物启动本地 mock，并用固定 operation matrix 校验响应 body 与 fixture body 字节级一致。当前固定 matrix 包括 `getMe`、`listTargetJobs`、`getPracticeSession`、`getFeedbackReport`、`requestPrivacyExport`。

### 4.4 Consumer contract

Mock consumer 的 scenario 选择规则固定为：

1. 请求显式指定 scenario 时，命中则使用该 scenario。
2. 未指定 scenario 时使用 `default`。
3. 指定了 fixture 未声明的 scenario 时失败，不静默回退。

前端 MSW、后端 mock server、Prism 和文档站都必须消费 `openapi/fixtures/` 或由它生成的 OpenAPI examples。需要新增 mock variant 时，应在这里新增 scenario，并通过 validator 与 consumer gate。

## 5 验收标准

- `openapi/fixtures/` 覆盖当前 35 个 operationId，没有多余 operation fixture。
- `make validate-fixtures` 通过，并能拒绝缺 fixture、schema drift、缺 provenance、非保留隐私值、非 UUIDv7 id 和 `tmp_` id。
- `make sync-fixtures-from-prototype` 幂等，且 P0 closed-loop endpoints 的 `prototype-baseline` scenario 非空并 schema-valid。
- `make render-openapi-fixture-examples` 通过，生成 examples 与 fixture body 字节级一致。
- Prism smoke 固定 matrix 通过，其中 `requestPrivacyExport` 返回 `501 + PRIVACY_EXPORT_NOT_AVAILABLE`。
- `openapi/fixtures/README.md`、`openapi/README.md` 和本 owner docs 均只描述当前 fixture truth source 与 consumer contract。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| Prototype data 与 OpenAPI schema 字段不同步 | sync 工具 fail-fast，并要求先修正 `PROTOTYPE_MAPPING.md` 或 `ui-design/src/data.jsx`；不在脚本里静默改名 |
| fixture 手写 response 漂出 schema | `make validate-fixtures` 校验所有 scenario，fixture edit 必须伴随 validator 通过 |
| privacy export 被误写成成功响应 | validator 对 `requestPrivacyExport` 固定检查 `501 + PRIVACY_EXPORT_NOT_AVAILABLE` |
| AI provenance 被写成空值 | validator 强制 provenance 字段存在且非空 |
| consumer 私自复制 mock body | consumer owner 必须引用本目录或生成 examples；新增 variant 只能通过 fixture scenario 增加 |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-07 | 1.5 | 压缩 owner 文档为当前 35-operation fixture truth source、prototype sync、example projection and Prism smoke contract。 | product-scope/001-core-loop-module-pruning |
| 2026-05-04 | 1.4 | 补齐质量门禁分类。 | docs-only L1 review |
| 2026-05-03 | 1.3 | 刷新 fixture / example coverage 与 prototype-baseline endpoint 范围。 | product-scope v1.2 / openapi-v1-contract v1.9 |
| 2026-05-03 | 1.2 | 调整 fixture coverage、报告字段与 prototype mapping。 | openapi-v1-contract v1.9 |
| 2026-04-29 | 1.1 | 补齐 `Auth/deleteMe` fixture 与 operation coverage gate。 | plan review |
