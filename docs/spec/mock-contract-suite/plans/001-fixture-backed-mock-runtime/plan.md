# Fixture-backed Mock Runtime

> **版本**: 1.13
> **状态**: active
> **更新日期**: 2026-07-13

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

维护 `mock-contract-suite` 的当前可执行 mock runtime：前端 dev preview、generated-client mock transport、后端 mockruntime 和 lint gate 都从 B2 `openapi/fixtures/` 读取同一批 37-operation fixture，不复制第二套 mock 数据，也不直接使用 `ui-design/src/data.jsx` 作为运行时数据源。

本 plan 不拥有 OpenAPI schema、fixture 内容、业务 handler、真实 backend store、AI 调用或用户可见 BDD 场景；这些由各自 owner 维护。这里负责让 mock runtime 按当前 fixture truth source 工作，并用 gate 阻止范围外 route / tag / schema / config token 回流。当前原地重开以承接 OPENAPI-002 paste-only fixture/generated handoff。

## 2 当前合同

- `scripts/mock_contract/fixture_registry.py` 以 OpenAPI operationId 为 key 读取 `openapi/fixtures/`，当前 registry 覆盖 10 tag / 37 operationId。
- `frontend/src/api/mockTransport.ts` 与 `frontend/src/api/devMockClient.ts` 返回 generated API types。Vite dev 默认使用 fixture-backed client；`VITE_EI_API_MODE=real` 必须显式提供 `VITE_EI_API_BASE_URL` 才访问真实 backend；production 默认 same-origin `/api/v1`。
- `backend/internal/api/mockruntime` 使用同一 fixture registry 响应 HTTP request。named scenario 选择读取 fixture 中对应 scenario 的 status/body；未知 scenario 返回明确错误。
- `make lint-mock-contract` 串联 `make validate-fixtures`、`make lint-openapi`、fixture registry tests 和 `scripts/lint/mock_runtime_boundary.py`。
- boundary lint 禁止前端运行时代码 import `ui-design/src/data.jsx`，禁止 fixture response 泄漏 prototype-only display field，校验 `openapi/fixtures/` tag 目录严格等于当前 OpenAPI tag 集合，并拦截当前范围外的 mock/API token。
- `createPracticeVoiceTurn`、`/practice/sessions/{sessionId}/voice-turns` 与 `PracticeVoiceTurn*` 属于当前 practice-voice owner 合同；独立 `/voice` route 与 `Voice` tag 仍由 boundary lint 拦截。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + tooling + contract`
- **TDD 策略**: fixture registry、frontend mock transport、dev client factory、backend mockruntime、boundary lint 和 Make target 都有 focused tests。重进本 plan 时先运行对应 focused gate 暴露 drift，再最小修复 mock runtime 或 docs。
- **BDD 策略**: 本 plan 不创建本地 BDD 文件；TargetJob 由 P0.015 验证，Practice recovery exact markers 必须交给 frontend-workspace-and-practice/002 与 P0.046。BDD handoff 未通过不得收口。
- **替代验证 gate**: `make lint-mock-contract`、`make validate-fixtures`、`make lint-openapi`、`make codegen-check`、`python3 -m pytest scripts/lint/mock_runtime_boundary_test.py -q`、`go test ./backend/internal/api/mockruntime -count=1`、`pnpm --filter @easyinterview/frontend test src/api/mockTransport.test.ts src/api/devMockClient.test.ts src/api/clientFactory.test.ts`、`sync-doc-index --check`、`make docs-check`。

## 4 交付范围

### 4.1 Fixture registry

The registry reads fixture metadata from `openapi/fixtures/<tag>/<operationId>.json` and exposes operationId lookup for frontend and backend mock users. Coverage must match OpenAPI exactly: missing fixture, extra fixture, unexpected tag directory, mismatched operationId, or out-of-scope API token is a gate failure.

### 4.2 Frontend mock transport and dev client

The generated client receives fixture-backed fetch in mock mode. Tests cover typed responses, named scenario selection, unknown scenario failure, delay/abort handling, export fallback responses, auth session state, generated operation coverage and client factory mode resolution.

Generated operation coverage is asserted against the keys of the real `createDevMockFixtureRegistry()` result. The production module does not expose a second `getDevMockFixtureOperationIds` view over the private fixture array.

Dev preview defaults to fixture-backed mode so pages render without a real backend. Real backend mode is explicit only:

```sh
VITE_EI_API_MODE=real VITE_EI_API_BASE_URL=http://localhost:8080/api/v1 pnpm --filter @easyinterview/frontend dev
```

### 4.3 Backend mockruntime

The backend mockruntime handler maps HTTP method/path to operationId and returns the fixture scenario response. Tests compare the HTTP response status/body directly with fixture scenario content and verify unknown scenario failure.

### 4.4 Boundary lint

`scripts/lint/mock_runtime_boundary.py` owns the mock runtime boundary:

- frontend runtime must not import prototype data;
- fixture response bodies must not include prototype-only display fields;
- fixture tag directories must match OpenAPI tags;
- current mock/API token scan rejects out-of-scope routes, tags, schema keys and config paths while allowing the current practice-voice contract.

### 4.5 Backend mockruntime test helper cleanup

Delete the unreferenced `assertJSONField` / `lookupJSONPath` helper pair from `mockruntime_test.go`. Existing tests compare complete fixture response status/body and retain the active unknown-scenario assertion; no production or test behavior changes. Use backend staticcheck U1000 as the red gate and mockruntime tests plus `make lint-mock-contract` as green gates.

## 5 验收标准

- 当前 37 operation fixtures 均能被 registry 解析，没有多余 fixture 或 tag 目录。
- 前端 mock transport 与 dev client 返回 generated API types，并支持 named scenario / unknown scenario / auth state / export fallback 行为。
- Vite dev 默认 fixture-backed；dev real mode 缺少 `VITE_EI_API_BASE_URL` 时失败；production 默认 `/api/v1`。
- 后端 mockruntime response status/body 直接跟随 fixture scenario。
- `make lint-mock-contract`、focused frontend tests、backend mockruntime tests、codegen/docs gate 均通过。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| Mock runtime 复制第二套数据 | registry 只读取 B2 fixtures；后端与前端测试都对 fixture body 做断言 |
| 前端运行时代码 import prototype data | boundary lint 扫描 `frontend/src` import |
| Fixture tag 或 operation drift | `make validate-fixtures`、`make lint-openapi` 与 fixture registry tests 一起执行 |
| Dev mock 被当作真实集成验证 | dev real mode 需要显式 `VITE_EI_API_MODE=real` 和 `VITE_EI_API_BASE_URL` |
| out-of-scope token gate 误伤当前 voice turn API | boundary lint 只拦截独立 `/voice` route 与 `Voice` tag，允许 practice-voice owner operation |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-13 | 1.13 | Add Phase 9 Practice recovery fixture/runtime parity and P0.046 handoff; narrow TargetJob zero-reference semantics. | openapi-v1-contract 1.54 |
| 2026-07-13 | 1.12 | Reopen Phase 8 for OPENAPI-002 TargetJob paste-only mock parity, P0.015 handoff and scoped zero-reference gates. | OPENAPI-002 + openapi-v1-contract 002 |
| 2026-07-10 | 1.11 | 删除 dev mock fixture operationId 测试观察器，parity 改查真实 registry keys。 | tech-debt pruning |
| 2026-07-10 | 1.10 | 删除 backend mockruntime 测试中无调用的 JSON field/path helper。 | tech-debt pruning |
| 2026-07-10 | 1.9 | 统一 mock runtime 边界 gate 的 out-of-scope 命名并同步文档版本。 | tech-debt pruning |
| 2026-07-10 | 1.8 | 对齐当前 37-operation fixture-backed runtime，包含 `archiveTargetJob`。 | tech-debt pruning |
| 2026-07-07 | 1.7 | 压缩 owner 文档为当时 36-operation fixture-backed runtime、dev client、backend mockruntime and boundary lint contract。 | product-scope/001-core-loop-module-pruning |
| 2026-07-06 | 1.6 | 对齐当时 36 operationId discovery。 | product-scope D-22 |
| 2026-05-22 | 1.5 | 校准 practice voice contract precision gate。 | practice-voice-mvp |
| 2026-05-10 | 1.4 | 合并 named scenario truth source 与 frontend dev preview mock wiring。 | mock-contract-suite |
| 2026-05-10 | 1.3 | 补充 frontend Vite dev preview fixture-backed wiring。 | mock-contract-suite |

## 8 OPENAPI-002 TargetJob paste-only mock handoff

### 8.1 TDD red/green contract

After openapi-v1-contract 002 publishes the migrated fixtures and 001 publishes generated types, add focused frontend mock transport, backend mockruntime, registry and boundary tests. RED must expose the current URL/file/manual_form source union, TargetJob `sourceType/sourceUrl` response fields and `target_job_attachment` purpose. GREEN consumes the exact flattened `importTargetJob` request `{rawText,targetLanguage,resumeId}` without copying a local DTO or fixture body.

Negative tests must reject old `source` wrapper, `TargetJobImportSource*`, URL, `fileObjectId`, manual-form/title/company fields, TargetJob source response fields, `purpose=target_job_attachment` and compatibility aliases. Positive tests must prove `createUploadPresign` remains registered and resume/privacy fixture scenarios still resolve. Inventory remains 37 operations / 10 tags.

### 8.2 Runtime parity and P0.015 handoff

Both `frontend/src/api/mockTransport.ts` / dev client and `backend/internal/api/mockruntime` must return the exact migrated fixture response for `importTargetJob`, `listTargetJobs`, `getTargetJob` and `createUploadPresign`. Named scenario selection remains fail-loudly. Pass the paste accepted/failure fixture names and expected status/body markers to P0.015; the scenario owner proves the user-visible paste flow and must not retain URL/file/manual_form positive steps.

### 8.3 Alternative gate and zero-reference

Because this phase changes an internal mock adapter rather than defining a new user workflow, local BDD assets are not applicable. Replacement gates are focused registry/transport/mockruntime/boundary tests, `make lint-mock-contract`, fixture/OpenAPI/codegen checks and P0.015 handoff evidence.

Current positive fixture/generated/runtime mock/seed surfaces must contain zero positive/runtime `TargetJobImportSource*`, TargetJob URL/file/manual_form import branches, `sourceType`, `sourceUrl`, `target_job_attachment` or compatibility aliases. Accepted ADR/oracle and exact negative declarations may retain rejected tokens；whole-file/test-directory exclusion is forbidden.

## 9 Practice durable recovery mock handoff

Add focused registry/frontend/backend RED/GREEN tests that consume—not copy—the B2 get-session `pending|retryable_failed|terminal_failed|complete` projections and send `validation-empty-text|auth-unauthorized|session-not-found|reply-pending-conflict|client-message-mismatch|ai-timeout-retryable` scenarios. Exact fixture status/body, role-specific fields and unknown-scenario failure must match in both runtimes. A paired retry proof must load the retryable-failed user, submit the same ID/text, and yield one completed user plus one assistant with no duplicates. Hand markers to P0.046；mock parity does not replace browser/user-flow evidence.
