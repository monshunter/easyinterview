# Mock Contract Suite Spec

> **版本**: 1.21
> **状态**: active
> **更新日期**: 2026-07-15

## 1 背景与目标

`engineering-roadmap` S1 要求先建立 contract-backed mock runway，让当前核心入口和会话级页面能基于 B2 OpenAPI fixtures 跑通 P0 happy path。`mock-contract-suite` 是这条 runway 的工程 owner：它把 B2 fixtures、generated API types、runtime config 和前后端 mock 入口组织成可测试、可复用、可漂移检查的 mock 运行层。

本 subject 的目标是：

1. 让前端开发只消费 B2 fixtures 投影出的 API shape，不再直接读取 `frontend/src` 作为实现数据源。
2. 让后端或本地 dev 环境可以用同一批 fixtures 提供稳定 mock response。
3. 为后续 `frontend-shell`、D2-D6 前端 workstream 和后端切真 API 提供一致的 fixture-backed backend mock runtime。
4. 把 fixture drift、operation coverage 和当前范围负向搜索纳入可执行 gate。
5. 让正式前端的 Vite dev preview 默认消费同一套 fixtures，使已开发页面在没有真实 backend 时也可见；显式切真实 backend 时必须通过配置开关完成。

## 2 范围

### 2.1 In Scope

- 读取 `openapi/fixtures/` 当前 10 tag / 38 operation fixtures；Auth `updateMe`、扁平 `Resumes`、`getResumeSource` PDF 原件预览、TargetJob paste-only import/archive、report-owned conversation、PracticeVoiceTurn、failed report regeneration 与 runtime config 均属于当前 fixture coverage。公共 PracticeSessions listing 已按 OPENAPI-001 v1.7 删除。TargetJob paste-only contract 由 accepted [OPENAPI-002](../openapi-v1-contract/decisions/OPENAPI-002-targetjob-paste-only.md) 与 openapi-v1-contract 001/002/003 承接；其它落地路径由 [openapi-v1-contract/004-resume-additive-coverage](../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md)、[backend-resume](../backend-resume/spec.md)、[backend-auth/001](../backend-auth/plans/001-email-code-session-bootstrap/plan.md) 与 [frontend-shell/001](../frontend-shell/plans/001-app-shell-auth-settings/plan.md) 承接。
- 基于 generated OpenAPI types 为前端提供 fixture-backed API client 或 mock transport。
- 为本地后端或开发服务器提供同源 mock handler / router。
- 校验 fixtures 与 `openapi/openapi.yaml`、generated packages 和 fixture consumer registry 的一致性。
- 统一 mock response 中的 auth/session、target job、practice plan、practice session、report、resume、privacy 和 runtime config 基线。
- 为代码层 fixture consumers 与 dev preview 提供可重置的 seed profile；seed profile 必须表达为 `openapi/fixtures/<tag>/<operationId>.json` 内的 named scenarios，不得引入第二套 seed 数据源。
- 前端 Vite dev preview 的默认 API client wiring：`pnpm --filter @easyinterview/frontend dev` 在未显式选择真实 backend 时必须使用 fixture-backed transport。

### 2.2 Out of Scope

- 不新增或修改 OpenAPI operation；破坏性 API 变更归 B2 `openapi-v1-contract`。
- 不实现真实业务 store、AI 调用、文件上传、邮箱发送或 backend internal runner。
- 不新增 product-scope 当前范围之外的 route、tag、operation、schema key 或 runtime config 口径。
- 不把 `frontend/src` 作为 mock 数据真理源；正式前端只通过 generated client 与 fixture-backed transport 消费合同。
- 不替代后续 `e2e-scenarios-p0` 的真实端到端验证。
- 不在 production build 默认启用 fixture-backed mock；真实部署仍通过 same-origin `/api/v1` 访问 backend。

## 3 用户决策 / 待确认事项

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | Mock 数据真理源 | B2 `openapi/fixtures/` | 前端和后端 mock 必须从 fixtures 投影，不私造业务数据 |
| D-2 | Prototype 数据定位 | `frontend/src` 只做 baseline 映射参考 | 实现不能直接 import prototype data |
| D-3 | Mock 范围 | P0 happy path + 高风险错误态 | 不扩展当前范围外的空壳 |
| D-4 | Drift gate | mock runtime 必须跑 fixture coverage、OpenAPI diff / validation 和 current-scope negative search | 后续 UI / API 改动要先更新 owner truth source |
| D-5 | Frontend dev preview 默认行为 | Vite dev 默认 fixture-backed；`VITE_EI_API_MODE=real` 必须同时提供 `VITE_EI_API_BASE_URL` 才打真实 backend | 避免本地开发时大量真实接口报错导致页面不可见，且避免相对 `/api/v1` 隐式打到当前 frontend origin；具体端口由 local-dev-stack 配置拥有 |
| D-6 | TargetJob mock paste-only | `importTargetJob` mock request 只接受 flattened `{rawText,targetLanguage,resumeId}`；TargetJob fixture/generated mock response 不含 `sourceType` / `sourceUrl`；URL/file/manual_form 与 `target_job_attachment` 不得作为正向 mock 能力 | 保留通用 `createUploadPresign` 及 resume/privacy purpose；由 registry、frontend transport、backend mockruntime 与 boundary tests 证明代码层 parity |
| D-7 | Practice recovery mock parity | mock runtime 原样消费 B2 role-discriminated messages 与 typed failure fixtures：user 有 `clientMessageId/replyStatus`，assistant 无；get-session 覆盖四种 durable status，send 覆盖 validation/auth/not-found/conflict/mismatch/retryable failure 与 same-ID retry success | 不复制本地 recovery DTO/错误表；unknown scenario 继续 fail loudly；由 fixture-backed frontend/backend tests 证明 exact parity 与 replay semantics |
| D-8 | Report conversation mock replacement | registry 删除 `listPracticeSessions` 并原样消费 `getReportConversation` fixture；只返回 reportId/status/frozen context/ordered messages，不暴露内部 locator | frontend/backend mock 选择同一 Reports fixture；old operation/path/scenario unknown 并 fail loudly；总 coverage 仍为 37/37，不复制本地 transcript DTO |

## 4 设计约束

- Mock runtime 必须以 OpenAPI operationId 为检索 key，避免 route string 与 fixture 目录漂移。
- 前端 mock adapter 必须返回 generated API types，不能把 `any` 或 prototype-only fields 泄漏到业务组件。
- 后端 mock handler 必须复用同一 fixture registry，不能复制第二套 fixture JSON。
- seed profile 必须覆盖未登录、已登录、缺 session、缺简历、报告生成中、隐私删除请求等 P0 状态；消费者按 `openapi/fixtures/README.md` 的 scenario selection contract 读取，未知 scenario 必须 fail loudly，不能静默回落到 `default`。
- 后端 mock runtime 的 named scenario 回归测试必须以 `openapi/fixtures/<tag>/<operationId>.json` 中的 scenario response 为断言真理源，不得复制一套 hard-coded status / error code / response field 期望；否则 fixture 更新后会出现测试消费者漂移。
- 前端 dev preview mock client 必须从当前 generated operation inventory 反查 fixture 覆盖；新增 operation 后，如果 fixture 没接入 dev mock，应由测试失败暴露，而不是在浏览器里变成真实接口错误。
- 前端 dev preview 必须保留显式真实 backend 逃生口：`VITE_EI_API_MODE=real VITE_EI_API_BASE_URL=<url> pnpm --filter @easyinterview/frontend dev` 使用默认 generated client + real `fetch`；dev real 模式不得隐式使用相对 `/api/v1` 打到当前 frontend origin，也不得在缺少 `VITE_EI_API_BASE_URL` 时猜测 backend 地址或复制 local-dev-stack 端口默认值。
- Mock response 必须只覆盖当前 `openapi/openapi.yaml` inventory、当前 fixtures 与 product-scope current contract；禁止新增当前范围之外的 route、tag、operationId、schema key 或 runtime config path。
- tag / fixture 目录拦截必须覆盖目录名本身，包括空目录和 Git 不跟踪的目录；`openapi/fixtures/` 下只允许当前 10 个 tag 目录。
- TargetJob mock 必须原样消费 OPENAPI-002 迁移后的 fixture/generated types：`importTargetJob` request 精确为 `rawText` / `targetLanguage` / `resumeId`，read response 不含 `sourceType` / `sourceUrl`。旧能力在 positive/runtime mock surface 中必须 zero-reference；accepted ADR/oracle 与 exact negative declarations 可保留 rejected token，禁止 whole-file/directory exclusion。
- `createUploadPresign` operation 必须继续存在，mock coverage 只保留 resume/privacy purpose。TargetJob 收敛不得误删简历原件或隐私导出仍依赖的通用上传合同。
- Practice mock 只投影 B2 fixtures/generated types：四种 get-session reply status 与 send exact failure matrix 的 status/body 必须字节一致；retryable failure → reload → same-ID success 不得新增第二条 user/assistant。Mock adapter 不解析 `Error.message`，也不自行推断 retryability。
- Report conversation mock 只投影 B2 `Reports/getReportConversation.json` 与 generated types；registry 不得保留 `listPracticeSessions` key、path matcher、scenario alias 或 fallback。Success body 与 message row 必须 closed 且 sequence 严格递增；unknown/hidden/fail-closed scenario 原样透传 fixture status/body。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| fixtures | B2 `openapi-v1-contract` | fixture 内容、schema、operation coverage 和 examples provenance |
| frontend mock | `mock-contract-suite` + `frontend-shell` | generated client 的 mock transport 和 dev runtime wiring |
| backend mock | `mock-contract-suite` | 本地 handler/router 或 test harness，供 package-level contract tests 使用 |
| verification | code owners + root `Makefile` | focused tests 用于开发反馈；阶段完成由根 `make test` 承接前后端全量代码回归，mock runtime 不创建 BDD/E2E |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Fixture coverage | B2 当前 38 operation fixtures 已落地（含 Auth `updateMe`、扁平 `Resumes`、`getResumeSource`、`archiveTargetJob`、report-owned conversation、PracticeVoiceTurn、failed report regeneration 与 runtime config） | 运行 mock coverage gate | 每个当前 operationId 都能被 registry 解析且 schema 校验通过；旧 `completeMyProfile` / `listPracticeSessions` 不得解析；范围外 operation 不得作为正向 fixture coverage 目标 | 001-fixture-backed-mock-runtime |
| C-2 | 前端 mock 同源 | 前端请求 generated client | 切到 mock transport | response shape 来自 B2 fixtures，组件不 import prototype data | 001-fixture-backed-mock-runtime |
| C-3 | 后端 mock 同源 | 本地 API smoke 请求 mock handler | 命中任一 P0 operation | handler 返回同一 fixture registry 的 typed response | 001-fixture-backed-mock-runtime |
| C-4 | 当前范围负向搜索 | mock runtime / fixtures / generated artifacts 已生成 | 运行 scoped negative search | 不含当前 product-scope 范围之外的 route / tag / operationId / schema key / config path；不误杀普通业务文案 | 001-fixture-backed-mock-runtime |
| C-5 | 前端 dev client 选择 | Vite dev 未显式选择真实 backend | 创建 app client | 默认选择 fixture-backed client；只有显式 `VITE_EI_API_MODE=real` 且提供 `VITE_EI_API_BASE_URL` 才创建 real client | 001-fixture-backed-mock-runtime |
| C-6 | TargetJob paste-only mock parity | OPENAPI-002 与 openapi-v1-contract 002 已迁移 fixtures/generated artifacts | 运行 mock registry、frontend transport、backend mockruntime 与 boundary focused gates | `importTargetJob` 只接受 `{rawText,targetLanguage,resumeId}`；TargetJob response 无 `sourceType/sourceUrl`；URL/file/manual_form/`TargetJobImportSource*`/`target_job_attachment` 正向 surface 为零；`createUploadPresign` resume/privacy 仍可解析 | 001-fixture-backed-mock-runtime Phase 8 |
| C-7 | Practice recovery mock parity | B2 001/002 发布 role-discriminated generated types 与 planned fixtures | frontend/backend mock 选择 get/send recovery scenarios | exact status/body parity；user/assistant recovery fields 合法；validation/auth/not-found/conflict/mismatch/retryable markers 可选；unknown scenario fail loudly；same-ID retry 无重复消息 | 001-fixture-backed-mock-runtime Phase 9 |
| C-8 | Report conversation mock parity | B2 001/002 一对一替换 operation 与 fixture | frontend/backend mock 选择 ready/non-ready/empty/hidden/fail-closed conversation scenarios | `getReportConversation` exact status/body parity；closed message fields与顺序保持；deleted list operation/path/scenario fail loudly；registry 始终 37 operations | 001-fixture-backed-mock-runtime Phase 10 |

## 7 关联计划

- [001-fixture-backed-mock-runtime](./plans/001-fixture-backed-mock-runtime/plan.md)
