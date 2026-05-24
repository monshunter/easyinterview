# 001 Full Funnel Happy Journey

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-24

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

交付 P0 完整漏斗的 **happy 主干单条 journey**，用两种 driver 同时证明跨模块真实贯通：

- **API-level**（`E2E.P0.098`）：在 `backend/cmd/api` 内新增 `httptest` server + 真实 stack 的 Go scenario test，按顺序串起 `registerResume`（前置）→ `importTargetJob` → `getTargetJob`（poll ready）→ `createPracticePlan`（baseline）→ `startPracticeSession` → `appendSessionEvent` → `completePracticeSession` → `getFeedbackReport`（poll ready）→ `createPracticePlan`（`next_round` + `sourceReportId`），断言 handoff 链 `targetJobId → planId → sessionId → reportId → 派生 planId` 真实传递、异步 job 经真实 internal runner 完成、关键写操作幂等、隐私红线与 legacy-negative。
- **Playwright 全栈**（`E2E.P0.099`）：起真后端进程（连 dev-stack postgres，`APP_ENV=test` stub AI）+ 前端 build/preview 指向真后端，Playwright 驱动真实 UI 从首页导入走到报告并点击「进入下一轮」CTA，断言跨屏 nav、真实轮询 UI、CTA handoff 与隐私 / legacy 红线。

交付后，本 plan 成为 P0 闭环「真实 handoff 在真后端下端到端贯通」的首个直接 gate；复练 / 下一轮另一分支、真实复盘回流、失败 / 恢复 journey 由本 subject 后续 `002+` plan 原地派生。

## 2 背景

本 plan 由 `e2e-scenarios-p0` spec v1.0 同时段派生。spec §1 已确认实施前基线：97 个 slice 场景无完整漏斗贯通；`backend/cmd/api` 已有 `*_http_scenario_test.go` / `jdmatch_live_scenario_test.go` 真后端 harness 范式（`httptest.NewServer` + 真实 router/handler/store/internal runner/events + `DATABASE_URL` + `config.LoadCanonical(AppEnv:"test")` stub AI）。

漏斗各步消费的 operation、handoff 字段与异步轮询机制已对 `openapi/openapi.yaml` 核实（见 §3 operation matrix）。复练 / 下一轮经 `createPracticePlan` 的 `goal IN ('retry_current_round','next_round') + sourceReportId` 表达（[backend-practice/004](../../../backend-practice/plans/004-derived-plans-debrief/plan.md)），report 的 `nextActions` 只是建议项、不是派生触发器。

**前置依赖**（Phase 0 验证，未就绪则暂停进入 Phase 1）：

- `make dev-up` 的 postgres 可达，`make migrate-up` 已应用到最新 schema。
- `config.LoadCanonical(AppEnv:"test")` 可加载且 AI 落到 stub provider，不读业务 secret。
- 漏斗 8 个 operationId 在 generated server + 真实 handler 已挂载（各 owner plan 已 completed）。
- `backend/cmd/api` 现有 scenario harness 可复用（同包 helper：`testLoader` / `open*ScenarioDB` 等）。

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + contract + tooling（端到端用户旅程 + 跨层契约消费验证 + 场景脚本工具）
- **TDD 策略**: Code plan requires TDD。
  - API-level journey 先写 `TestE2EP0098FullFunnelImportToNextRound`（断言完整 handoff 链 + 异步 ready + 幂等 + 隐私），初始 Red（journey 未实现 / 真实 stack 未贯通），再让 orchestration 通过转 Green；测试文件 `backend/cmd/api/full_funnel_journey_scenario_test.go`，命令 `cd backend && go test ./cmd/api -run 'TestE2EP0098' -count=1`（postgres 不可达 `t.Skip`）。
  - Playwright journey 先写 `frontend/tests/e2e/full-funnel-journey.spec.ts` 断言 UI 走完漏斗，初始 Red，再绿；命令 `pnpm --filter @easyinterview/frontend exec playwright test tests/e2e/full-funnel-journey.spec.ts`。
  - Phase 1 / 2 每个 checklist item 命名其断言来源（见 checklist 各项尾注）。
- **BDD 策略**: Feature plan requires BDD。本 plan 引入端到端业务流程；BDD scenarios `E2E.P0.098`（API-level）+ `E2E.P0.099`（Playwright 全栈）已在 [bdd-plan.md](./bdd-plan.md) 分配，主 [checklist.md](./checklist.md) Phase 3 含 `BDD-Gate:` 项引用每个 scenario ID；执行使用场景框架 `scripts/setup.sh → trigger.sh → verify.sh → cleanup.sh` 四段入口，cleanup 失败时也必须执行。
- **替代验证 gate**:
  - operation matrix 真实性：`grep -rn "importTargetJob\|getTargetJob\|createPracticePlan\|startPracticeSession\|appendSessionEvent\|completePracticeSession\|getFeedbackReport\|registerResume" backend/internal/api/generated/` 命中真实 server 方法。
  - 隐私红线：journey test / verify.sh 断言响应 / event / audit / log / DB 可观测面不含 JD 原文、答案文本、报告 prose；Playwright 侧扫描 URL / localStorage / sessionStorage / console。
  - legacy-negative：`grep -rn "welcome\|growth\|mistakes\|drill\|followup\|mode=debrief\|experiences" test/scenarios/e2e/p0-098-* test/scenarios/e2e/p0-099-*` 0 命中。
  - 文档一致性：`validate_context.py` / `sync-doc-index --check` / `make docs-check` / `git diff --check`。

### 3.1 Operation Matrix

漏斗按真实 `openapi/openapi.yaml` 消费以下 operation（不新增 / 修改任何 operation）：

| # | operationId | method + path | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|---|-------------|---------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| 0 | `registerResume` | POST `/resumes` | `Resumes/registerResume.json` | ResumeCreateFlow（P0.099 前置 / P0.098 seed） | real（resume handler/store） | `resume_assets` | none | P0.098 / P0.099 前置 |
| 1 | `importTargetJob` | POST `/targets/import` | `TargetJobs/importTargetJob.json` | HomeScreen 导入 | real（targetjob handler + drainer） | `target_jobs` + `jobs` | `target_import` via stub | P0.098 / P0.099 |
| 2 | `getTargetJob` | GET `/targets/{targetJobId}` | `TargetJobs/getTargetJob.json` | ParseScreen 轮询 | real | `target_jobs` | none（读 `analysisStatus`） | P0.098 / P0.099 |
| 3 | `createPracticePlan` | POST `/practice/plans` | `PracticePlans/createPracticePlan.json` | WorkspaceScreen / Report CTA | real | `practice_plans` | none | P0.098 / P0.099（baseline + next_round） |
| 4 | `startPracticeSession` | POST `/practice/sessions` | `PracticeSessions/startPracticeSession.json` | WorkspaceScreen 立即面试 | real + internal runner | `practice_sessions` + `session_events` + outbox | `practice.session.first_question` via stub | P0.098 / P0.099 |
| 5 | `appendSessionEvent` | POST `/practice/sessions/{sessionId}/events` | `PracticeSessions/appendSessionEvent.json` | PracticeScreen 事件循环 | real | `session_events` | session AI via stub | P0.098 / P0.099 |
| 6 | `completePracticeSession` | POST `/practice/sessions/{sessionId}/complete` | `PracticeSessions/completePracticeSession.json` | PracticeScreen 完成 | real + internal runner | `feedback_reports`（reserve）+ `jobs` + outbox | `report_generate` via stub | P0.098 / P0.099 |
| 7 | `getFeedbackReport` | GET `/reports/{reportId}` | `Reports/getFeedbackReport.json` | ReportDashboard / Generating 轮询 | real | `feedback_reports` | none（读 `status`） | P0.098 / P0.099 |
| 8 | `getJob` | GET `/jobs/{jobId}` | `Jobs/getJob.json` | 通用 job 轮询（备选） | real | `jobs` | none | P0.098（备选轮询断言） |

> handoff 链：`registerResume → resumeAssetId`；`importTargetJob → targetJobId`；`getTargetJob.analysisStatus=ready`；`createPracticePlan(targetJobId, resumeAssetId, goal=baseline) → planId`；`startPracticeSession(planId) → sessionId`；`completePracticeSession(sessionId) → reportId`；`getFeedbackReport(reportId).status=ready → nextActions`；`createPracticePlan(goal=next_round, sourceReportId=reportId, targetJobId, resumeAssetId) → 派生 planId`。

## 4 实施步骤

### Phase 0: 真后端环境与前置依赖验证

#### 0.1 dev-stack 与 migration 就绪

确认 `make dev-up` postgres 可达、`make migrate-up` 至最新；记录 `DATABASE_URL` 约定（默认 `postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable`，沿用现有 scenario harness）。

#### 0.2 stub AI 与 secret 边界

确认 `config.LoadCanonical(AppEnv:"test")` 加载成功且漏斗各 AI 步骤（target_import / first_question / session / report_generate）落到 stub provider；确认未读 `AI_PROVIDER_*` 业务 secret。

#### 0.3 operation 真实挂载验证

按 §3.1 operation matrix grep generated server + 真实 handler，确认 8 个 operationId 真实可调用、handler 已挂载（不是 mock-only）。

#### 0.4 journey 前置 seed 设计

设计 journey 内前置：注册 / 受控 seed 一个已认证 user + 一个 ready resume asset（`createPracticePlan` 必需 `resumeAssetId`），并明确 cleanup 边界。

### Phase 1: API-level full-funnel journey（E2E.P0.098）

#### 1.1 journey test harness

在 `backend/cmd/api/full_funnel_journey_scenario_test.go` 复用现有 harness（`httptest.NewServer` + `DATABASE_URL` + `testLoader`/`LoadCanonical(AppEnv:"test")` + 真实 router/handler/store/internal runner/events）；postgres 不可达 `t.Skip`。

#### 1.2 import → parse ready

调用 `importTargetJob`（paste JD）→ 拿 `targetJobId` + `target_import` job；由真实 internal runner 处理后轮询 `getTargetJob` 直到 `analysisStatus=ready`，断言结构化解析结果真实落库。

#### 1.3 createPracticePlan（baseline）

用 `targetJobId` + 前置 `resumeAssetId` + `goal=baseline` 调用 `createPracticePlan` → 拿 `planId`；断言 plan 真实落库并绑定 targetJob / resume。

#### 1.4 session 事件循环

`startPracticeSession(planId)` → `sessionId` + 首题（stub AI）；`appendSessionEvent` 逐题作答推进；断言 session / events 真实落库、`session_started` 等 outbox 仅一次。

#### 1.5 complete → report ready

`completePracticeSession(sessionId)` → `reportId` + `report_generate` job；真实 internal runner 处理后轮询 `getFeedbackReport` 直到 `status=ready`，断言报告与 `nextActions`（含 `next_round` 类型）真实生成。

#### 1.6 next_round 派生（handoff 链闭合）

用 `goal=next_round` + `sourceReportId=reportId` + `targetJobId` + `resumeAssetId` 调用 `createPracticePlan` → 派生 `planId`；断言派生 plan 真实关联 source report、与首个 plan 不同。

#### 1.7 幂等断言

对 `startPracticeSession` / `completePracticeSession` / `createPracticePlan` 用同 Idempotency-Key replay，断言无重复副作用（无第二 session / report / plan、无重复 outbox）。

#### 1.8 隐私红线 + legacy-negative 断言

断言 journey 全程响应 / event / audit / log / DB 可观测面不含 JD 原文 / 答案文本 / 报告 prose；负向断言不出现旧 route / 旧模块 / 旧 `mode=debrief` / 旧 feature_key。

### Phase 2: Playwright full-stack journey（E2E.P0.099）

#### 2.1 全栈环境拉起

脚本拉起真后端进程（连 dev-stack postgres，`APP_ENV=test` stub AI）+ 前端 build/preview 指向真后端 base URL（非 fixture mock transport）；seed 已认证 user + resume asset。

#### 2.2 UI 走完漏斗

`frontend/tests/e2e/full-funnel-journey.spec.ts`：Playwright 从首页导入 JD → ParseScreen 解析 ready → Confirm 进 WorkspaceScreen → 立即面试 → PracticeScreen 完成 session → Generating → ReportDashboard → 点击「进入下一轮」CTA。

#### 2.3 真实轮询 UI

断言解析 loading 与 report generating 的真实轮询 UI 在真后端异步 job 推进下正确过渡到 ready（非 mock 即时返回）。

#### 2.4 CTA handoff

断言 Report「进入下一轮」CTA 触发 `createPracticePlan(next_round, sourceReportId)` + `startPracticeSession`，nav 到新 workspace/practice 且 query 含派生 planId / fresh sessionId。

#### 2.5 隐私 + legacy 红线

断言 URL / localStorage / sessionStorage / console 不泄露 JD 原文 / 答案 / 报告 prose；scenario 树 legacy 负向 grep 0 命中。

### Phase 3: 场景登记与收口

#### 3.1 scenario 资产

创建 `test/scenarios/e2e/p0-098-full-funnel-import-to-next-round-journey/` 与 `p0-099-full-funnel-fullstack-ui-journey/`：`README.md` + `data/` + `scripts/{setup,trigger,verify,cleanup}.sh`；`trigger.sh` 保留 runner exit code，`verify.sh` 检查 runner 日志真实执行证据（命令 marker + 目标 test 路径 + pass marker），拒绝 no-op / skip-as-pass；登记 `test/scenarios/e2e/INDEX.md`。

#### 3.2 scenario 执行

两个场景各按 `setup → trigger → verify → cleanup` 执行通过，证据写入 `.test-output/e2e/<scenario>/trigger.log` 并由 `verify.sh` 消费。

#### 3.3 文档一致性

`validate_context.py` / `sync-doc-index --check` / `make docs-check` / `git diff --check` 通过；operation matrix 终态与实现一致。

## 5 验收标准

- 本 plan §4 列出的实现 / 测试项全部通过（Phase 0-3）。
- spec [C-1~C-8](../../spec.md#6-验收标准) 全部满足。
- BDD-Gate `E2E.P0.098` / `E2E.P0.099` 场景验证通过。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 真后端全链路首次贯通可能暴露各域集成缺口（handoff 字段、异步时序、outbox 时机） | 这正是 journey 的价值；缺口按 `/change-intake` 路由到对应 owner spec/plan 原地修复，不在本 plan 内绕过 |
| postgres / dev-stack 不可达导致 journey 无法运行 | 沿用现有范式 `t.Skip` 并输出 skip 原因；CI 缺依赖不算 pass，verify.sh 拒绝 skip-as-pass |
| 异步 job 轮询超时导致 flaky | 用有界轮询 + 明确超时 + 真实 runner 同步触发；超时视为失败并记录 job 状态，不静默重试 |
| Playwright 全栈环境（真后端进程 + 前端 build）启动成本高、易污染 | setup 显式拉起 / cleanup 显式回收；scenario README 写明 isolation；失败优先检查环境污染（框架 §8） |
| stub AI 输出与真实报告维度结构漂移 | 只断言契约结构 / 状态 / handoff，不断言 AI 文案质量；质量归 F3 eval workstream |
