# OpenAPI v1 Contract Breaking-Change Gate

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本 plan 承接 [openapi-v1-contract spec](../../spec.md) 的 v1.0.0 freeze gate：

- `openapi/baseline/openapi-v1.0.0.yaml` 是当前 35 operation / 10 tag contract 的 baseline snapshot。
- `make openapi-diff` 比对 baseline 与 `openapi/openapi.yaml`，并按 B2 additive-only rules 拦截 breaking changes。
- wrapper rules reject endpoint deletion, method/path change, schema field deletion, type change, required-field addition and enum value removal.
- additive changes are allowed only when OpenAPI, fixtures, generated artifacts, inventory lint and spec records move together.
- `POST /api/v1/privacy/exports` P0 `501` to P1 `202` is the only response-status whitelist; it still requires a same-change contract record.
- breaking changes outside the whitelist require an accepted ADR before OpenAPI changes land.

本 plan 不实现 API handler、frontend consumer、fixtures or codegen outputs; those remain under their respective owner plans. It owns the local gate that prevents contract drift from passing silently.

## 2 当前合同

### 2.1 Gate Matrix

| Surface | Current contract | Verification |
|---------|------------------|--------------|
| Baseline snapshot | `openapi/baseline/openapi-v1.0.0.yaml` mirrors current freeze shape | `make openapi-diff` confirms baseline/current compatibility |
| Diff wrapper | `scripts/lint/openapi_diff.py` reclassifies raw diff output to B2 rules | wrapper unit tests + `make openapi-diff` |
| Diff config | `openapi/diff-config.yaml` pins 35 operation inventory and privacy export whitelist | `make openapi-diff`, openapi diff unit tests |
| ADR workflow | `docs/spec/openapi-v1-contract/decisions/TEMPLATE.md` defines accepted ADR shape for breaking changes | docs link checks + owner review |
| Baseline guide | `openapi/baseline/README.md` defines baseline selection and versioning rules | docs link checks + owner review |
| Contract handoff | B2 001/002/003 gates must pass before implementation owners consume generated clients | `make lint-openapi`, `make validate-fixtures`, `make codegen-check`, `make openapi-diff` |

### 2.2 Breaking / Additive Rules

Breaking:

- remove endpoint, method or path
- remove schema field
- change field type
- add a required request/response field
- remove enum value
- change response status semantics outside explicit whitelist

Additive:

- add endpoint or tag with updated inventory
- add optional schema field
- add enum value
- add optional query/header parameter
- add fixture scenario or example without changing existing semantics

### 2.3 Privacy Export Whitelist

The only status-transition whitelist is:

```yaml
path: /api/v1/privacy/exports
method: POST
from: "501"
to: "202"
```

The wrapper must verify the exact path, method and statuses. Any other status transition is breaking unless accepted through ADR and spec revision.

## 3 质量门禁

- **Plan 类型**: `contract + tooling + governance`。
- **TDD 策略**: 适用。Wrapper unit tests cover breaking/additive reclassification, composition schema diff, privacy export whitelist and contract-record requirements.
- **BDD 策略**: 不适用。本 plan is internal contract evolution tooling and has no user behavior flow.
- **替代验证 gate**:
  - `make openapi-diff`
  - `python3 -m unittest scripts.lint.openapi_diff_test`
  - `make lint-openapi`
  - `make validate-fixtures`
  - `make codegen-check`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/openapi-v1-contract/plans/003-breaking-change-gate/context.yaml --target contract`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

## 4 实施步骤

### Phase 1: baseline and diff entrypoint

- Keep `openapi/baseline/openapi-v1.0.0.yaml` aligned to the current freeze baseline.
- Keep `make openapi-diff` wired through the repo Makefile.
- Support explicit baseline version selection when future baselines are added.

### Phase 2: ruleset and wrapper

- Enforce B2 breaking/additive rules through `scripts/lint/openapi_diff.py`.
- Detect composition changes under `oneOf` / `allOf` / `anyOf`.
- Keep privacy export whitelist exact and isolated.

### Phase 3: contract governance

- Maintain the ADR template for accepted breaking changes.
- Maintain baseline README guidance for baseline selection and versioning.
- Require contract records for response status transitions and baseline changes.

### Phase 4: closeout

- Run OpenAPI validation, fixture validation, codegen drift check and diff gate.
- Keep 001/002/003 B2 handoff green before downstream implementation plans consume generated clients.
- Keep product-scope pruning evidence tied to current OpenAPI truth source.

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | Current OpenAPI and baseline are compatible | `make openapi-diff` |
| A-2 | Breaking schema/endpoint changes fail | `python3 -m unittest scripts.lint.openapi_diff_test` |
| A-3 | Additive optional changes pass | `python3 -m unittest scripts.lint.openapi_diff_test` |
| A-4 | Privacy export whitelist is exact and isolated | `make openapi-diff`, wrapper unit tests |
| A-5 | Generated artifacts and fixtures remain current | `make lint-openapi`, `make validate-fixtures`, `make codegen-check` |

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-07 | 1.6 | Compress owner docs to the current 35-operation breaking-change gate contract. |
| 2026-05-04 | 1.5 | Add quality-gate classification for the completed breaking-change gate. |
