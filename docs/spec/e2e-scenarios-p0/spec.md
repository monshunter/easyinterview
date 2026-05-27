# E2E Scenarios P0 Spec

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-05-27

## 1 背景与目标

[engineering-roadmap §6.4 S3](../engineering-roadmap/spec.md#64-s3--true-integration-and-release-gate) 把 `e2e-scenarios-p0` 列为「覆盖导入 -> 规划 -> 练习 -> 报告 -> 复练当前轮 / 下一轮 -> 真实复盘」完整 P0 漏斗的 owner subject。

当前 `test/scenarios/e2e` 已登记 87 条切片场景（最高编号 `E2E.P0.097`，编号有空档），但全部是**单一可独立收口的行为切片**（README §3）：每个场景只验证某个 owner spec 的局部 C-* 条件（如 targetjob 导入、workspace 渲染、practice loop、report 渲染、debrief 分析），彼此独立。没有任何一条场景把这些切片**串成跨模块的完整用户旅程**，因此 P0 闭环的「真实 handoff 在真后端下端到端贯通」一直没有被直接验证。

本 subject 创建时的实施前基线是：

- 各 P0 业务域（targetjob / practice / review / resume / debrief / jobs-recommendations / profile / auth / upload）的 plan checklist 已基本收口；近两周大量提交在修复「real backend gate drift」，说明各 owner 正在把 mock 切真后端。
- `backend/cmd/api` 已有成熟的真后端场景范式：`*_http_scenario_test.go` / `jdmatch_live_scenario_test.go` 用 `httptest.NewServer` 组装真实 router/handler/store/internal runner/events，连 `DATABASE_URL`（默认 dev-stack postgres `localhost:5432`）；`config.LoadCanonical(AppEnv:"test")` 负责加载 canonical config 与 AI profile/registry，场景 AI 由 harness 注入确定性 stub / fixture client。
- `make dev-up` 提供 Docker Compose 外部依赖（postgres/redis/minio/mailpit），backend/frontend 默认 host-run 进程（见 [development.md §2](../../development.md#2-frontend--backend-contract-workflow)）。

目标：

1. 验证 P0 完整漏斗跨模块真实贯通：handoff 链 `targetJobId → planId → sessionId → reportId → 派生 planId` 在真实 handler/store/runner/event/DB + scenario stub / fixture AI 下端到端可用。
2. 建立可独立收口的「happy 主干 journey」gate，为后续分支（复练 / 下一轮）、真实复盘回流、失败 / 恢复 journey 与 release gate 奠基。
3. 用两种 driver 同时证明：API-level（后端域间贯通）与 Playwright 全栈（前端在真后端下走完漏斗）。
4. 在自动化 journey 之后提供统一纳入 `e2e` 套件的真实 provider hybrid UAT 场景：AI Agent 先执行环境 preflight、材料/配置/隐私检查和统一 result artifact；人工或浏览器 Agent 再补齐真实前端、真实后端、真实 AI LLM、Mailpit 本地 magic-link 登录、双语输入材料、checklist、证据归档和清理边界。

## 2 范围

### 2.1 In Scope

- 完整漏斗 **happy 主干单条 journey**：`JD 导入 → 解析 ready → 面试规划 → 完整 session（首题 + 事件循环 + 完成）→ 报告 ready → 进入下一轮（next_round）派生`。
- 两种 driver：
  - **API-level**（`E2E.P0.098`）：`backend/cmd/api` 内新增 `httptest` server + 真实 stack 覆盖 9 行 operation matrix（8 个主链必经 operation + `getJob` 备选轮询 / handler gate；`createPracticePlan` 复用 baseline + next_round 两次调用）的 Go scenario test。
  - **Playwright 全栈**（`E2E.P0.099`）：真后端进程 + 前端 build + 真 postgres，Playwright 驱动真实 UI 走完漏斗。
- 真后端全栈 + scenario stub / fixture AI（`APP_ENV=test`），postgres 不可达时 `t.Skip`（沿用现有范式）。
- handoff 链字段真实传递、异步 job（`target_import` / `report_generate`）经真实 internal runner 完成、DB 真实落库、关键写操作幂等、隐私红线、legacy-negative。
- 真实 provider hybrid UAT（`E2E.P0.100`）：标准 `test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/` 场景目录 + `APP_ENV=dev` 后端 + `VITE_EI_API_MODE=real` 前端 + 真实 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` + Mailpit 本地 magic-link 登录 + synthetic JD / resume / answer materials。

### 2.2 Out of Scope

- 复练（`retry_current_round`）分支与真实复盘（`debrief`）回流 journey —— 归后续 plan（`002+`）。
- 失败 / 恢复 journey（报告生成失败、AI timeout、跨用户隔离贯通）—— 归后续 plan。
- 新增或修改任何 OpenAPI operation、DB schema、event payload、prompt / rubric —— 本 subject 只**消费**现有契约。
- 真实 LLM 质量 / 评估 —— 归 [prompt-rubric-registry §7](../prompt-rubric-registry/spec.md) 的 `004-real-model-profile-and-evals` eval workstream owner；本 subject 只用 stub。
- 旧模块 / 旧 route（Welcome / Growth / Plan / Mistakes / Drill / Followup / Experiences / STAR / 独立 Voice），见 [engineering-roadmap §4.1](../engineering-roadmap/spec.md#41-产品与-ui-约束)。
- Kind / K8s / Helm 部署级场景，见 [test/scenarios/e2e/README.md §2](../../../test/scenarios/e2e/README.md)。
- 真实 provider hybrid UAT 不替代 F3 eval 的模型质量基线，也不把具体 AI 输出文案固化为自动化断言；它只要求真实 provider 调用、用户可观察质量记录和脱敏证据。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 后端形态 | 真后端全栈：复用 `backend/cmd/api` httptest + `DATABASE_URL`（dev-stack postgres）+ 真实 router/handler/store/internal runner/events + harness 注入的确定性 AI client；postgres 不可达 `t.Skip` | journey 证明的是真实集成贯通，不是 mock 流转 |
| D-2 | AI provider | stub / recorded scenario client；`config.LoadCanonical(AppEnv:"test")` 与 AI profile/registry 必须在未配置 `AI_PROVIDER_*` 时可加载/解析 | journey 可重复、不依赖外网；AI 质量归 F3 eval workstream |
| D-3 | 覆盖广度 | 首个 plan = 单条 happy 主干 journey（含到 `next_round` 派生一跳）；复练 / 真实复盘回流 / 失败 journey 归后续 plan | 首个 plan 可独立收口部署，规模可控 |
| D-4 | owner 归属 | 新建本 subject；journey 场景在 `test/scenarios/e2e/` 新建目录，与现有 slice 场景区分 | 跨模块 journey 有明确 owner，不挂靠单一域 owner |
| D-5 | journey driver | 两种都要：API-level（`E2E.P0.098`）+ Playwright 全栈（`E2E.P0.099`） | 后端域间贯通与前端全栈体验都被覆盖 |
| D-6 | 场景编号 | 接续 `E2E.P0.098` / `E2E.P0.099`，目录 slug 标注 `full-funnel` journey 以区分 slice | 编号与现有框架一致，语义上标识 journey |
| D-7 | journey 前置 | seed 已认证 user + resume asset（`createPracticePlan` 必需 `resumeAssetId`）；001 默认经 `registerResume` 创建，若环境 bootstrap 使用受控 DB seed，仍必须保留 `registerResume` matrix / fixture / handler gate | happy 链的前置准备，不属于核心 handoff 验证点，但不得让 `registerResume` 退化为未覆盖契约 |
| D-8 | hybrid UAT owner | 真实 provider 人工验收作为 `002-manual-uat-real-provider-full-funnel` 原计划的修订内容，但执行资产必须归入标准 `e2e` 场景目录 `p0-100-real-provider-full-funnel-hybrid/`，不再维护独立 `manual-uat` 套件 | 防止把人工验收材料作为场景框架外文件或把 stub AI 自动化证明冒充真实 LLM 验收 |
| D-9 | hybrid UAT 账号入口 | 002 使用 `local-dev-stack` 提供的 Mailpit 本地 mailbox：synthetic `.example.test` 邮箱触发真实 `startAuthEmailChallenge`，Mailpit 收取 magic-link，`verifyAuthEmailChallenge` 签发 `ei_session`；不得直接写 `sessions` 表或新增正式 `backend/cmd` / Go helper | 支持用户验收时材料齐备，包括账号；本地测试不依赖真实外部邮箱服务或真实邮箱账号，同时避免场景依赖越界进入正式后端进程树 |
| D-10 | hybrid UAT AI provider | 后端必须以 `APP_ENV=dev` 连接真实 provider，默认 DeepSeek OpenAI-compatible endpoint；`APP_ENV=test`、deterministic stub、P0.099 test server 与 frontend fixture mock 都不能作为真实 UAT 完成证据 | 明确真实联调边界 |
| D-11 | scenario helper 语言边界 | `test/scenarios/` 新增场景工具只允许 shell / Python；可编排既有产品 runner，但不得新增 `backend/cmd` / Go helper 作为场景依赖 | 防止人工 UAT / BDD 依赖越界进入正式后端进程树 |
| D-12 | 执行者顺序 | 场景执行者首先是 AI Agent：先运行环境 preflight、四段脚本、材料/配置/隐私检查和 result artifact；人或浏览器 Agent 只在同一场景输出目录补齐真实凭证/浏览器观察证据 | 避免真实前后端联调用例脱离场景框架，保证自动执行与人工执行共享同一入口和证据路径 |

### 3.2 待确认事项

| ID | 待确认事项 | 默认处理 |
|----|------------|----------|
| Q-1 | 后续分支 / 回流 / 失败 journey 的编号段 | 默认接续 `E2E.P0.100+`，由 `002+` plan 在派生时锁定 |
| Q-2 | Playwright 全栈是否需要独立 compose app service | 默认沿用 host-run 进程（[development.md §2 step 7](../../development.md#2-frontend--backend-contract-workflow)），不预设 compose app service |
| Q-3 | 真实 provider hybrid UAT 是否保留现场 | 默认保留 `.test-output/e2e/p0-100-real-provider-full-funnel-hybrid/` 脱敏证据，DB cleanup 仅在用户确认后执行 |

## 4 设计约束

- **真后端范式**：API-level journey 必须复用 `backend/cmd/api` 现有 scenario harness（`httptest.NewServer` + 真实 router/handler/store/internal runner/events + `DATABASE_URL`），不得为 journey 另起一套 mock stack；postgres 不可达时 `t.Skip` 并输出 skip 原因。
- **Playwright 全栈约束**：起真后端进程（连 dev-stack postgres，`APP_ENV=test`，场景 AI 使用确定性 stub / fixture client）+ 前端 build/preview 以 `VITE_EI_API_MODE=real` / `VITE_EI_API_BASE_URL=http://127.0.0.1:<backend-port>/api/v1` 指向真后端 base URL，Playwright 驱动真实 UI；不得用 fixture-backed mock transport 冒充真后端。
- **AI 边界**：journey 全程使用 harness 注入的 stub / fixture AI client，遵循 [development.md §2.4](../../development.md#24-ai-provider-boundary) 与 [A5 §5 secret red line](../ci-pipeline-baseline/spec.md)；canonical test config 与 AI profile/registry 预检必须在未配置 `AI_PROVIDER_*` 等业务 secret 时通过。
- **真实 provider UAT 边界**：hybrid UAT 必须使用 `APP_ENV=dev` 后端真实进程、真实 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY`、真实 PostgreSQL、真实 frontend real mode；不得用 `APP_ENV=test`、deterministic stub、fixture-backed frontend mock transport、`Prefer: example=<scenario>` 或 `EI_E2E_P0_099_SERVER=1` 冒充。
- **账号材料边界**：hybrid UAT 必须提供可重复的 synthetic account sign-in 材料；tracked 文件只能保存邮箱、Mailpit URL、验证步骤和 cleanup 说明，不得保存真实 session cookie、auth secret、AI key 或 magic-link token。`test/scenarios` 只允许 shell/Python 辅助；不得新增 `backend/cmd` / Go helper，也不得直接写 `sessions` 表绕过 auth flow。
- **场景统一管理边界**：真实 provider / 人工观察类用例不得作为 `test/scenarios` 下的独立 companion 套件存在；必须登记在活跃 suite `INDEX.md`，具备标准四段脚本、`data/seed-input.md`、`data/expected-outcome.md` 与统一 result artifact。缺人工证据时使用 `MANUAL_REQUIRED`，不得伪装为 PASS 或退化为 ERROR。
- **契约消费约束**：journey 只消费 [openapi-v1-contract](../openapi-v1-contract/spec.md) 已存在的 operation，不新增 / 修改 operation、schema、event；handoff 字段以 `openapi/openapi.yaml` 真实 schema 为准。
- **框架契约**：遵循 [test/scenarios/README.md](../../../test/scenarios/README.md) 与 [e2e/README.md](../../../test/scenarios/e2e/README.md)：每个场景目录含 `README.md` + `data/` + `scripts/{setup,trigger,verify,cleanup}.sh`；`verify.sh` 必须检查 runner 日志中的真实执行证据（命令 / runner marker + 目标 test 路径 + pass marker），拒绝 no-op；`BDD-Gate` 只引用场景编号。
- **隐私红线**：journey 全程响应 / event / audit / log / metric 只暴露 ID / 状态 / 计数 / 错误码摘要，不泄露 JD 原文、答案文本、报告 prose（[product-scope](../product-scope/spec.md) 隐私红线）。
- **legacy-negative**：journey 树与被消费的 active runtime 不得出现旧 route / 旧模块 / 旧 `mode=debrief` / 旧 feature_key 等被当前设计丢弃的口径；旧 route 反查必须覆盖 `welcome` / `growth` / `plan` / `mistakes` / `drill` / `followup` / `experiences` / `star` / `onboarding` / 独立 `voice`，并用 route-aware pattern 避免误伤合法的 `createPracticePlan`、`practice_plans`、`resumeAssetId`、`resume_assets`。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| scenario | `test/scenarios/e2e/` | journey 场景目录（`p0-098-*` / `p0-099-*`）+ `INDEX.md` 登记；4 段脚本契约 |
| backend journey | `backend/cmd/api/` | API-level journey test（`TestE2EP0098...FullFunnel...`），复用现有 scenario harness |
| frontend journey | `frontend/` Playwright | 全栈 UI journey（`E2E.P0.099`），真后端进程 + 前端 build |
| real provider hybrid scenario | `test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/` | AI Agent first-run preflight、真实 provider runbook、材料、checklist、Mailpit 登录说明、四段脚本与统一 result artifact |
| dev UAT account sign-in | `local-dev-stack` Mailpit + `backend-auth` | 002 使用 synthetic email + Mailpit magic link；不得新增正式 `backend/cmd` / Go helper，不得直接写 session 表 |
| 消费契约（只读） | 各 owner spec / coded truth source | `openapi/openapi.yaml` operations、各域 handler/store/internal runner、`backend-async-runner`、dev-stack；本 subject 不修改这些 owner 资产 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | happy journey 全链贯通 | 真后端 + dev-stack postgres + scenario stub / fixture AI + seed user/resume | 顺序调用 import → poll → createPlan → startSession → appendEvent → complete → poll report → createPlan(next_round) | 每步真实响应 envelope，handoff ID（targetJobId/planId/sessionId/reportId/派生 planId）真实传递并被下一步消费 | 001 Phase 1 / 2 |
| C-2 | 异步 job 真实完成 | `target_import` / `report_generate` 入队 | 真实 internal runner 处理 | resource status 由 queued/processing → ready，DB 真实落库；journey 轮询到 ready | 001 Phase 1 / 2 |
| C-3 | 关键写操作幂等 | start / complete / createPlan 携带 Idempotency-Key | 同 key replay | 无重复副作用（无第二个 session / report / plan，无重复 outbox） | 001 Phase 1 |
| C-4 | 隐私红线 | journey 全程 | 检查响应 / event / log / audit / DB 持久化的可观测面 | 不出现 JD 原文 / 答案文本 / 报告 prose；只 ID / 状态 / 计数 / 错误码 | 001 Phase 1 / 2 |
| C-5 | legacy-negative | journey 树 + 被消费 active runtime | route-aware 负向 grep + frontend scope gate | 旧 route / 旧模块 / 旧 `mode=debrief` / 旧 feature_key 0 命中；`plan` / `resume` 仅作为独立 route key 被拒绝，不误伤合法 `createPracticePlan` / `resumeAssetId` | 001 Phase 1 / 2 |
| C-6 | scenario gate 真实执行 | `E2E.P0.098` / `E2E.P0.099` 场景就绪 | `setup → trigger → verify → cleanup` | verify 校验 runner 日志真实执行证据（命令 marker + 目标 test 路径 + pass marker），拒绝 no-op / skip-as-pass | 001 Phase 3 |
| C-7 | operation matrix 完整 | 跨层 journey plan | 查看 plan | 9 行 operation matrix × fixture / frontend consumer / backend handler / persistence / AI dependency / scenario coverage 全部标明真实状态（8 个主链必经 operation + `getJob` 备选轮询 / handler gate；`createPracticePlan` 复用 baseline + next_round） | 001 §3.1 operation matrix |
| C-8 | 文档一致性 | 文档集创建 / 修订完成 | 运行校验 | `validate_context` / `sync-doc-index --check` / `make docs-check` / `git diff --check` 通过 | 001 Phase 3 |
| C-9 | hybrid UAT 材料完整 | AI Agent / 人工准备真实 provider 验收 | 查看 `test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/` | Mailpit magic-link 登录说明、JD、简历、作答样例、checklist、环境变量模板、证据归档与 cleanup 说明齐备；tracked 文件无真实 secret / PII，且不存在 `backend/cmd/devsession` / `backend/internal/devsession` | 002 Phase 1-3 |
| C-10 | 真实联调启动路径 | 本地机器具备 Docker / Go / Node / pnpm 与真实 AI key | 按 runbook 启动 dev-stack、migrate、backend、frontend，并通过 Mailpit 登录 | backend 为 `APP_ENV=dev` 真进程，frontend 为 `VITE_EI_API_MODE=real`，请求命中真实 backend，AI provider env 指向真实 LLM endpoint，auth flow 通过 Mailpit 本地邮箱完成 | 002 Phase 2 |
| C-11 | 无 mock/stub 冒充 | hybrid UAT runbook / checklist | 审查完成证据 | 不接受 `APP_ENV=test`、P0.099 test server、fixture mock transport、deterministic stub AI 或 `Prefer: example=` 作为真实 UAT 完成证据 | 002 Phase 2-4 |
| C-12 | dev-only 账号安全边界 | hybrid UAT 账号入口存在 | 运行 no-backend-cmd / no-Go scenario negative gate 与材料 secret scan | 只允许 synthetic `.example.test` 邮箱和本地 Mailpit；不输出 auth secret、AI key、session secret、magic-link token 或 cookie value；场景 helper 不进入正式 backend cmd / internal 包 | 002 Phase 1 |
| C-13 | 真实 AI 调用脱敏证据 | 验收者走完整漏斗 | 检查 backend log / DB task-run / manual checklist | 记录 provider/profile/model/latency/task-run count 等摘要，不记录 prompt、response、JD 原文、答案或报告 prose | 002 Phase 3-4 |
| C-14 | 统一场景框架执行 | `E2E.P0.100` 已登记为 `hybrid` Ready 场景 | AI Agent 运行标准四段脚本 | 共享环境先通过顶层 env preflight；脚本写出 `result.json`，缺真实证据时结果为 `MANUAL_REQUIRED`；`test/scenarios/manual-uat` 不再作为独立入口存在 | 002 Phase 5 |

## 7 关联计划

- [001-full-funnel-happy-journey](./plans/001-full-funnel-happy-journey/plan.md)：API-level（`E2E.P0.098`）+ Playwright 全栈（`E2E.P0.099`）两条 happy 主干 journey；后续分支 / 回流 / 失败 journey 由 `002+` 在本 subject 内原地派生。
- [002-manual-uat-real-provider-full-funnel](./plans/002-manual-uat-real-provider-full-funnel/plan.md)：真实 provider hybrid UAT 场景（`E2E.P0.100`），覆盖 AI Agent first-run preflight、Mailpit magic-link 登录、真实前后端、真实 AI LLM、无 mock/stub 边界和脱敏证据。

## 8 关联文档

- [engineering-roadmap](../engineering-roadmap/spec.md) §6.4 S3 —— 本 subject 的上游规划入口
- [product-scope](../product-scope/spec.md) —— P0 漏斗与隐私红线真理源
- [openapi-v1-contract](../openapi-v1-contract/spec.md) —— 被消费 operation 契约
- [backend-async-runner](../backend-async-runner/spec.md) —— 异步 job / outbox runner 承接者
- [backend-targetjob](../backend-targetjob/spec.md) / [backend-practice](../backend-practice/spec.md) / [backend-review](../backend-review/spec.md) —— 漏斗各步后端 owner
- [development.md](../../development.md) §2 —— 前后端契约工作流与真后端运行边界
- [test/scenarios/README.md](../../../test/scenarios/README.md) / [e2e/README.md](../../../test/scenarios/e2e/README.md) —— 场景框架契约
