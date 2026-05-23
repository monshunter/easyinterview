# Fixture-backed Mock Runtime

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-05-22

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

落地 `mock-contract-suite` 的首个可执行 runway：用 B2 OpenAPI fixtures 建立前端 mock transport、后端 mock handler/test harness 和 fixture drift gate，让 `frontend-shell` 与后续 D2-D6 前端 workstream 能在不私造 product data 的前提下开发。

## 2 背景

`engineering-roadmap` S1 明确要求先创建或修订 `mock-contract-suite`，把 34 operation fixtures 提供给前端和后端 mock。当前仓库已经有 B2 OpenAPI、fixtures、generated frontend/backend types 和 prototype mapping，但还缺少统一 mock runtime owner。

本 plan 只建立 contract-backed mock runtime，不新增用户可见流程，也不修改 API schema。

## 2.1 修订记录

| 日期 | 版本 | 变更 | 原因 |
|------|------|------|------|
| 2026-05-10 | 1.4 | 合并 Phase 4.5 named scenario truth-source remediation 与 Phase 5 frontend dev preview mock wiring。 | 当前分支和 `main` 均已完成 mock-contract-suite 修订，合并后必须保留两边 gate 并消除同版本并行语义。 |
| 2026-05-10 | 1.3 | 重新激活本 plan，新增 Phase 5 frontend dev preview mock wiring。 | 既有交付只覆盖测试/可注入 client，没有覆盖 `pnpm --filter @easyinterview/frontend dev` 的默认预览路径，导致无真实 backend 时大量 `/api/v1` 报错且页面不可见。 |
| 2026-05-22 | 1.5 | 重新激活本 plan，新增 Phase 6 practice voice contract precision gate。 | repo-wide `make lint` 暴露 retired-token gate 误伤 `createPracticeVoiceTurn`、`/voice-turns` 与 `PracticeVoiceTurn*` generated artifacts；必须保留独立 `/voice` route / `Voice` tag 拦截，同时允许 practice-voice-mvp owner contract。 |

## 3 质量门禁分类

- **Plan 类型**: `code-internal` + `tooling` + `contract`。
- **TDD 策略**: 通过 `/implement mock-contract-suite/001-fixture-backed-mock-runtime tooling` -> `/tdd` 执行；每个 checklist item 先补 focused test 或 lint fixture，再实现最小 runtime / gate；测试断言写在 checklist 的 `验证:` 后。Phase 4.5 必须用后端 focused test 证明 named scenario response expectation 来自 fixture truth source；Phase 5 必须先用 frontend focused tests 证明 dev client 默认 fixture-backed、真实 backend opt-out 保留、fixture coverage 覆盖全部 generated operationId，再接入 `main.tsx`。
- **BDD 策略**: BDD 不适用。本 plan 不引入用户可见 UI、API 行为或业务流程，只提供内部 mock runtime 和 contract gate；用户行为验证归 `frontend-shell` 与后续 D2-D6 plan 的 BDD gate。
- **替代验证 gate**: fixture registry unit tests、frontend mock transport tests、frontend dev mock bootstrap tests、backend mock handler tests、`make validate-fixtures`、`make lint-openapi`、`make codegen-check`、prototype mapping drift check、fixture tag directory set gate、scoped retired route / tag / schema / config token negative search、frontend typecheck/build smoke、`make docs-check`。

## 4 实施步骤

### Phase 1: Fixture registry 与 coverage

#### 1.1 建立 operationId -> fixture registry

读取 `openapi/fixtures/` 目录，以 B2 operationId 作为 key 建立 registry。registry 必须区分 tag、operationId、fixture path、HTTP status 和 schema expectation。

#### 1.2 补齐 fixture coverage lint

新增或扩展 lint，校验 `openapi/openapi.yaml` 的 operationId 与 fixture 文件一一对应；优先复用 `make validate-fixtures` 与 `make lint-openapi` 的 B2 owner gate，缺失、多余、旧 tag 或旧 route 直接失败，且不得维护第二份手写 operation inventory。

### Phase 2: Frontend mock transport

#### 2.1 接入 generated client mock transport

在前端 API 层提供 fixture-backed mock transport，使 generated client 在 dev/test 模式下返回 typed fixture response。

#### 2.2 拦截 prototype data 运行时依赖

补测试或 lint，确认 `frontend/src` 业务实现不 import `ui-design/src/data.jsx` 或 prototype-only fields。

### Phase 3: Backend mock harness

#### 3.1 提供后端 mock handler / router

在后端 API 测试或 dev harness 中复用同一 fixture registry，提供 Auth、TargetJobs、PracticePlans、PracticeSessions、Reports、Resumes、Debriefs、Privacy 和 runtime config 等 P0 operation 的 mock response。

#### 3.2 错误态和 seed profile

为未登录、已登录、缺 session、缺简历、报告生成中、隐私删除请求等状态建立 seed profile。seed profile 必须落在 `openapi/fixtures/<tag>/<operationId>.json` 的 named scenarios 中，由 mock runtime 通过 `Prefer: example=<scenario>` 或等价显式 selector 选择；未知 scenario 必须 fail loudly，不能静默回落到 `default`。

### Phase 4: Drift gates and handoff

#### 4.1 接入本地质量门禁

把 fixture coverage、operation registry metadata test、prototype import boundary 和旧口径 negative search 接入本地 lint 或 codegen gate。旧口径 gate 必须使用精确 retired token（如 `/mistakes`、`/growth`、`/drill`、`/voice`、`Mistakes` / `Growth` tag、`single_drill`、`gateway_route`、`ai.gateway*`、`default.provider`、`task_type`），避免误杀 `growth-stage` 等普通业务文案。

#### 4.2 Handoff 给 frontend-shell

记录 `frontend-shell` plan 可消费的 mock runtime 入口、seed profile 和阻塞条件，确保后续 `/implement frontend-shell/001-app-shell-auth-settings frontend` 不需要重新设计 mock 数据源。

#### 4.3 L2 runtime drift remediation

补强 `lint-mock-contract` 的 operation registry metadata、Go route table operation count 和前端 named scenario 回归覆盖，确保历史 36-row 与 scenario 静默回退不会重新进入 runtime。

#### 4.4 Fixture tag directory gate

扩展 mock runtime boundary lint，校验 `openapi/fixtures/` 的 tag 目录集合严格等于当前 OpenAPI 12 tag；即使旧 `Growth` / `Mistakes` 为空目录或 Git 不跟踪，也必须被 gate 捕获并清理。

#### 4.5 Remediation: named scenario expectations follow fixture truth source

修复后端 mock runtime named scenario 回归测试中的 hard-coded response expectation。`TestHandlerSelectsNamedSeedScenariosAndFailsUnknown` 必须读取对应 fixture scenario 的 `response.status` 和 `response.body` 作为断言真理源，覆盖 `getPracticeSession` 的 `missing-session` scenario 当前为 `404 PRACTICE_SESSION_NOT_FOUND`，避免测试继续期待旧的 `401 AUTH_UNAUTHORIZED`。

### Phase 5: Frontend dev preview mock wiring

#### 5.1 建立 dev preview mock client factory

在 `frontend/src/api` 提供 dev-only fixture-backed client 工厂，默认读取全部 `openapi/fixtures/<tag>/<operationId>.json` 并校验 generated operationId 覆盖；测试必须证明 `DEV` 环境默认返回 fixture 数据，且 `VITE_EI_API_MODE=real` 会回到真实 generated client。

#### 5.2 接入 Vite frontend bootstrap

把 `frontend/src/main.tsx` 的默认 client 构造改为通过 factory 解析：Vite dev 默认 fixture-backed，production build 默认真实 backend，显式 `VITE_EI_API_MODE=real` 可在 dev 下打真实 backend。dev real 模式必须使用 `VITE_EI_API_BASE_URL` 或 backend 默认 `http://localhost:8080/api/v1`，不得继续用相对 `/api/v1` 打到 frontend 5173。

#### 5.3 文档与本地 smoke 收口

更新 frontend API handoff 文档，说明 dev mock 默认行为、真实 backend opt-out、`VITE_EI_API_BASE_URL` 与新增 fixture scenario 的维护入口；运行 focused frontend tests、typecheck/build，并用本地 dev server smoke 证明页面不再因真实接口错误不可见。

### Phase 6: Practice voice contract precision

#### 6.1 Retired token gate precision

收窄 `lint-mock-contract` 的 retired token 匹配：继续拦截独立 `/voice` route、`Voice` tag 和旧模块 token，但允许 `practice-voice-mvp` 拥有的 `createPracticeVoiceTurn`、`/practice/sessions/{sessionId}/voice-turns`、`PracticeVoiceTurn*` generated artifacts 与 fixture 字段。验证必须覆盖 `python3 -m pytest scripts/lint/mock_runtime_boundary_test.py -q`、`make lint-mock-contract` 与 `make lint`。

## 5 验收标准

- 34 operation fixtures 均能被 operationId registry 解析。
- 前端 mock transport 返回 generated API types，不依赖 prototype data。
- 后端 mock harness 与前端 mock 使用同一 fixture registry。
- Vite dev frontend 在没有真实 backend 时默认使用 fixture-backed client，已开发页面可渲染；真实 backend opt-out 只能通过显式 `VITE_EI_API_MODE=real`。
- scoped retired route / tag / schema key / config path negative search 通过，且不误杀普通业务文案。
- 所有 checklist item 的 focused tests / lint / drift gates 实际运行通过。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| Mock runtime 复制第二套数据 | Phase 1 要求 registry 只读取 B2 fixtures |
| 前端为了速度继续 import prototype data | Phase 2.2 增加 import boundary test / lint |
| Mock handler 与真实 OpenAPI 漂移 | Phase 1.2 与 Phase 4.1 接入 fixture coverage 和 OpenAPI validation |
| 场景测试误把 mock smoke 当真实 E2E | 本 plan 不创建 BDD-Gate；真实用户行为由后续 feature plan 和 `test/scenarios/e2e` 承接 |
| Dev preview mock 被误用于真实集成验证 | README 与 factory opt-out 明确 `VITE_EI_API_MODE=real`，operation matrix 仍区分 fixture-backed 与真实 backend |
