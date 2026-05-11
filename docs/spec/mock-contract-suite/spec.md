# Mock Contract Suite Spec

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-05-11

## 1 背景与目标

`engineering-roadmap` S1 要求先建立 contract-backed mock runway，让当前 UI 五入口和会话级页面能基于 B2 OpenAPI fixtures 跑通 P0 happy path。`mock-contract-suite` 是这条 runway 的工程 owner：它把 B2 fixtures、generated API types、runtime config 和前后端 mock 入口组织成可测试、可复用、可漂移检查的 mock 运行层。

本 subject 的目标是：

1. 让前端开发只消费 B2 fixtures 投影出的 API shape，不再直接读取 `ui-design/src/data.jsx` 作为实现数据源。
2. 让后端或本地 dev 环境可以用同一批 fixtures 提供稳定 mock response。
3. 为后续 `frontend-shell`、D2-D6 前端 workstream 和后端切真 API 提供一致的 fake backend。
4. 把 fixture drift、operation coverage 和 retired UI/module terminology 纳入可执行 gate。
5. 让正式前端的 Vite dev preview 默认消费同一套 fixtures，使已开发页面在没有真实 backend 时也可见；显式切真实 backend 时必须通过配置开关完成。

## 2 范围

### 2.1 In Scope

- 读取 `openapi/fixtures/` 当前 13 tag / 46 operation fixtures（B2 spec D-17 additive 升级，JobMatch tag 12 operation 由 `frontend-home-job-picks-and-parse/002-jd-match-recommendations` 落地 fixture）；**B2 spec D-18 Resume Workshop additive 升级声明阶段**（B2 spec 1.16）已声明扩到 13 tag / 55 operation（保留 `Resumes` tag 扩容，新增 9 operationId + 多 variant fixtures），落地路径由 [openapi-v1-contract/004-resume-additive-coverage](../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) 承接；本 spec §2.1 / §6 C-1 / `openapi/fixtures/README.md` 中的 46 数字在 B2 plan 004 落地后由本 spec 同步 1.5 → 1.6 升级到 55。
- 基于 generated OpenAPI types 为前端提供 fixture-backed API client 或 mock transport。
- 为本地后端或开发服务器提供同源 mock handler / router。
- 校验 fixtures 与 `openapi/openapi.yaml`、generated packages 和 `openapi/fixtures/PROTOTYPE_MAPPING.md` 的一致性。
- 统一 mock response 中的 auth/session、target job、practice plan、practice session、report、resume、debrief、privacy 和 runtime config 基线。
- 为后续 scenario / BDD gate 提供可重置的 seed profile；seed profile 必须表达为 `openapi/fixtures/<tag>/<operationId>.json` 内的 named scenarios，不得引入第二套 seed 数据源。
- 前端 Vite dev preview 的默认 API client wiring：`pnpm --filter @easyinterview/frontend dev` 在未显式选择真实 backend 时必须使用 fixture-backed transport。

### 2.2 Out of Scope

- 不新增或修改 OpenAPI operation；破坏性 API 变更归 B2 `openapi-v1-contract`。
- 不实现真实业务 store、AI 调用、文件上传、邮箱发送或 backend internal runner。
- 不直接恢复旧 Growth、Mistakes、Drill、独立 Voice、旧 `resume` route 或旧 `task_type` profile 口径。
- 不把 `ui-design/src/data.jsx` 作为运行时真理源；它只保留为 prototype-baseline 对照输入。
- 不替代后续 `e2e-scenarios-p0` 的真实端到端验证。
- 不在 production build 默认启用 fixture-backed mock；真实部署仍通过 same-origin `/api/v1` 访问 backend。

## 3 用户决策 / 待确认事项

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | Mock 数据真理源 | B2 `openapi/fixtures/` | 前端和后端 mock 必须从 fixtures 投影，不私造业务数据 |
| D-2 | Prototype 数据定位 | `ui-design/src/data.jsx` 只做 baseline 映射参考 | 实现不能直接 import prototype data |
| D-3 | Mock 范围 | P0 happy path + 高风险错误态 | 不扩展 P1/P2 future candidate 空壳 |
| D-4 | Drift gate | mock runtime 必须跑 fixture coverage、OpenAPI diff / validation 和 retired-name negative search | 后续 UI / API 改动要先更新 owner truth source |
| D-5 | Frontend dev preview 默认行为 | Vite dev 默认 fixture-backed，`VITE_EI_API_MODE=real` 才打真实 backend；real 模式默认 base URL 为 `http://localhost:8080/api/v1`，可用 `VITE_EI_API_BASE_URL` 覆盖 | 避免本地开发时大量真实接口报错导致页面不可见，且避免相对 `/api/v1` 隐式打到 frontend 5173 |

## 4 设计约束

- Mock runtime 必须以 OpenAPI operationId 为检索 key，避免 route string 与 fixture 目录漂移。
- 前端 mock adapter 必须返回 generated API types，不能把 `any` 或 prototype-only fields 泄漏到业务组件。
- 后端 mock handler 必须复用同一 fixture registry，不能复制第二套 fixture JSON。
- seed profile 必须覆盖未登录、已登录、缺 session、缺简历、报告生成中、隐私删除请求等 P0 状态；消费者按 `openapi/fixtures/README.md` 的 scenario selection contract 读取，未知 scenario 必须 fail loudly，不能静默回落到 `default`。
- 后端 mock runtime 的 named scenario 回归测试必须以 `openapi/fixtures/<tag>/<operationId>.json` 中的 scenario response 为断言真理源，不得复制一套 hard-coded status / error code / response field 期望；否则 fixture 更新后会出现测试消费者漂移。
- 前端 dev preview mock client 必须从当前 generated operation inventory 反查 fixture 覆盖；新增 operation 后，如果 fixture 没接入 dev mock，应由测试失败暴露，而不是在浏览器里变成真实接口错误。
- 前端 dev preview 必须保留显式真实 backend 逃生口：`VITE_EI_API_MODE=real pnpm --filter @easyinterview/frontend dev` 使用默认 generated client + real `fetch`，且 dev real 模式不得隐式使用相对 `/api/v1` 打到 frontend 5173；未设置 `VITE_EI_API_BASE_URL` 时使用 backend 默认 `http://localhost:8080/api/v1`。
- Mock response 中不得出现旧模块口径：独立 `/mistakes`、`/growth`、`/drill`、`/voice` route，`Mistakes` / `Growth` / `Drill` / `Voice` tag，`listMistakes` / `getGrowthOverview` operationId，旧 `single_drill` practice mode，旧 `gateway_route` / `ai.gateway*` / AI gateway 运行时配置，旧 `default.provider` 或 `task_type` schema key。普通业务文案中的 `growth-stage` 等非模块含义词不属于禁止项。
- 旧 tag 拦截必须覆盖 fixture 目录名本身，包括空目录和 Git 不跟踪的目录；`openapi/fixtures/` 下不得残留 `Growth`、`Mistakes` 等 retired tag 目录。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| fixtures | B2 `openapi-v1-contract` | fixture 内容、schema、operation coverage 和 examples provenance |
| frontend mock | `mock-contract-suite` + `frontend-shell` | generated client 的 mock transport 和 dev runtime wiring |
| backend mock | `mock-contract-suite` | 本地 handler/router 或 test harness，供 API smoke 使用 |
| scenarios | `test/scenarios/e2e` | 用户行为场景资产，不由 mock suite 直接标记 ready |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Fixture coverage | B2 已有 46 operation fixtures（含 JobMatch tag 12 operation；B2 D-18 声明阶段扩到 55 operation，由 openapi-v1-contract/004-resume-additive-coverage 落地） | 运行 mock coverage gate | 每个 operationId 都能被 registry 解析且 schema 校验通过；B2 D-18 落地后本 C-1 数字由 46 同步升 55 | 001-fixture-backed-mock-runtime（C-1 数字升级跟随 openapi-v1-contract/004） |
| C-2 | 前端 mock 同源 | 前端请求 generated client | 切到 mock transport | response shape 来自 B2 fixtures，组件不 import prototype data | 001-fixture-backed-mock-runtime |
| C-3 | 后端 mock 同源 | 本地 API smoke 请求 mock handler | 命中任一 P0 operation | handler 返回同一 fixture registry 的 typed response | 001-fixture-backed-mock-runtime |
| C-4 | 旧口径拦截 | mock runtime / fixtures / generated artifacts 已生成 | 运行 scoped negative search | 不含旧 route / tag / operationId / schema key / config path 等 retired token；不误杀普通业务文案 | 001-fixture-backed-mock-runtime |
| C-5 | 前端 dev preview 可见 | 没有启动真实 backend | 运行 Vite dev frontend 并打开已开发页面 | 默认 fixture-backed client 返回 runtime/auth/业务 fixtures，页面可渲染；只有显式 `VITE_EI_API_MODE=real` 才访问真实 `/api/v1` | 001-fixture-backed-mock-runtime |

## 7 关联计划

- [001-fixture-backed-mock-runtime](./plans/001-fixture-backed-mock-runtime/plan.md)
