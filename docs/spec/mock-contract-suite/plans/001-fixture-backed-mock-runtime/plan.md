# Fixture-backed Mock Runtime

> **版本**: 1.9
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

维护 `mock-contract-suite` 的当前可执行 mock runtime：前端 dev preview、generated-client mock transport、后端 mockruntime 和 lint gate 都从 B2 `openapi/fixtures/` 读取同一批 37-operation fixture，不复制第二套 mock 数据，也不直接使用 `ui-design/src/data.jsx` 作为运行时数据源。

本 plan 不拥有 OpenAPI schema、fixture 内容、业务 handler、真实 backend store、AI 调用或用户可见 BDD 场景；这些由各自 owner 维护。这里负责让 mock runtime 按当前 fixture truth source 工作，并用 gate 阻止范围外 route / tag / schema / config token 回流。

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
- **BDD 策略**: BDD 不适用。本 plan 只提供内部 mock runtime；用户行为由 `frontend-shell`、当前 feature owner 和 `test/scenarios/e2e` 验证。
- **替代验证 gate**: `make lint-mock-contract`、`make validate-fixtures`、`make lint-openapi`、`make codegen-check`、`python3 -m pytest scripts/lint/mock_runtime_boundary_test.py -q`、`go test ./backend/internal/api/mockruntime -count=1`、`pnpm --filter @easyinterview/frontend test src/api/mockTransport.test.ts src/api/devMockClient.test.ts src/api/clientFactory.test.ts`、`sync-doc-index --check`、`make docs-check`。

## 4 交付范围

### 4.1 Fixture registry

The registry reads fixture metadata from `openapi/fixtures/<tag>/<operationId>.json` and exposes operationId lookup for frontend and backend mock users. Coverage must match OpenAPI exactly: missing fixture, extra fixture, unexpected tag directory, mismatched operationId, or out-of-scope API token is a gate failure.

### 4.2 Frontend mock transport and dev client

The generated client receives fixture-backed fetch in mock mode. Tests cover typed responses, named scenario selection, unknown scenario failure, delay/abort handling, export fallback responses, auth session state, generated operation coverage and client factory mode resolution.

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
| 2026-07-10 | 1.9 | 统一 mock runtime 边界 gate 的 out-of-scope 命名并同步文档版本。 | tech-debt pruning |
| 2026-07-10 | 1.8 | 对齐当前 37-operation fixture-backed runtime，包含 `archiveTargetJob`。 | tech-debt pruning |
| 2026-07-07 | 1.7 | 压缩 owner 文档为当时 36-operation fixture-backed runtime、dev client、backend mockruntime and boundary lint contract。 | product-scope/001-core-loop-module-pruning |
| 2026-07-06 | 1.6 | 对齐当时 36 operationId discovery。 | product-scope D-22 |
| 2026-05-22 | 1.5 | 校准 practice voice contract precision gate。 | practice-voice-mvp |
| 2026-05-10 | 1.4 | 合并 named scenario truth source 与 frontend dev preview mock wiring。 | mock-contract-suite |
| 2026-05-10 | 1.3 | 补充 frontend Vite dev preview fixture-backed wiring。 | mock-contract-suite |
