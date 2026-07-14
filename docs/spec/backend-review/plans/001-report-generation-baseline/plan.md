# 001 — Grounded Conversation Report Generation

> **版本**: 2.25
> **状态**: active
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

在既有 conversation-level report owner 内消除隐藏数值评分与后端二次语义，消费 `backend-practice/002` 已冻结的完整当轮上下文，让 LLM 直接生成可验证、可解释、可复练的最终报告；completion snapshot 与 report-derived plan 分别由 backend-practice/002、004 唯一负责，本 plan 只消费 named marker，并用真实 provider 内容核对与浏览器截图闭环。

## 2 Operation Matrix

| operationId / job | fixture | consumer | backend | persistence | AI | verification |
|-------------------|---------|----------|---------|-------------|----|--------------|
| `report_generate` job | N/A | generating poll | runner/review service/store | feedback_reports/async_jobs/outbox_events/audit_events/ai_task_runs | report.generate | focused service/store/integration + independent eval + P0.099 visible result |
| `getFeedbackReport` | `Reports/getFeedbackReport.json`: queued/generating/ready-needs-practice/ready-well-prepared/failed/invalid-contract/long-content | generating + report dashboard | reports handler/review read | feedback_reports + frozen context projection | none | focused handler/consumer contract + P0.099 real API/UI |
| `listTargetJobReports` | `Reports/listTargetJobReports.json` | target-scoped ReportsScreen only; no Parse/generating/report consumer | reports handler/review read | feedback_reports + current TargetJob canonical summary | none | focused handler/store/consumer contract |

`completePracticeSession` 与 `createPracticePlan` 只作为 references/marker handoff：本 plan 不拥有其 API、fixture、handler/store 或 scenario directory。

## 2.1 Cross-owner Handoff / Execution Order

| Order | Owner | Deliverable / gate | Unblocks |
|-------|-------|--------------------|----------|
| 0 | `openapi-v1-contract/003` | accepted pre-release breaking ADR; snapshot merge-base old baseline | proposed OpenAPI edit |
| 1 | `shared-conventions-codified/001 Phase 9` | canonical non-retryable `REPORT_CONTEXT_TOO_LARGE` + `REPORT_CONTEXT_TOO_LARGE_CONVENTIONS_PASS` | B2 error-enum parity + backend terminal error |
| 2 | `openapi-v1-contract/001/003/002` | edit proposed schema while old baseline stays frozen → exact breaking/additive audit → fixtures + Go/TS codegen | backend API/consumer changes + frontend 7.1 |
| 3 | `db-migrations-baseline/001 Phase 8` + `prompt-rubric-registry/002 Phase 14 prep` | 000018 storage + `REPORT_STORAGE_V18_PASS` + v0.2 prompt/schema/multi-version `REPORT_PROMPT_V020_READY`（尚不激活） | completion snapshot + report runtime prerequisites |
| 4 | `backend-practice/002 Phase 9` | `practice-completion-evidence.v1` + `ZERO_ANSWER_COMPLETION_REJECTED_PASS` / `REPORT_CONTEXT_SNAPSHOT_PASS` / `REPORT_CONTEXT_REPLAY_PASS` | review context consumer |
| 5 | `prompt-rubric-registry/004 Phase 8` | v0.2 rubric/context-aware judge/eval + `REPORT_RUBRIC_V020_PASS` / `REPORT_CONTEXT_AWARE_EVAL_PASS` | F3 final activation + independent reliability eval |
| 6 | `prompt-rubric-registry/002 Phase 14 activation` | consume both 004 markers, atomically activate prompt/rubric v0.2 via 000019, preserve v0.1 rollback, emit `REPORT_PROMPT_V020_PASS` | backend generation/provenance |
| 7a | `backend-review/001 Phase 6.1-6.4` | consume schema/prompt/context contracts and implement the frozen request builder; committed `input-*.json` remain deleted while zh/en worst-case output-schema fixtures remain | report runtime and semantic validation |
| 7b | `ai-provider-and-model-routing/003` | loader contract owns report defaults/override/invalid; canonical coverage lint requires all six active profiles `max_tokens >= 16384` and keeps report context at 1,000,000 | report model-profile configuration |
| 7c | `backend-review/001 Phase 6.5-7.5` | consume the A3 profile configuration contract; implement runtime byte gate, direct semantics, validator/repair/persistence/read and exact backend evidence | frontend report owner + P0.099 visible result |
| 8 | `backend-practice/004 Phase 3` | generic-empty-focus retry, issue-backed non-empty focus, next successor, isolation + IK markers | report replay CTA and code-level consumer gates |
| 9 | code owners + `e2e-scenarios-p0` P0.099 | code owners complete unit/integration/eval gates；P0.099 is reserved for an explicitly run real report/generating API/UI acceptance | final closeout |

`backend-review/001` 只拥有 `backend/internal/api/reports`、`backend/internal/review`、`backend/internal/store/review`；B2/B4/F3/backend-practice/frontend/scenario 均通过 context references 与 marker handoff。它不重复领取 completion、derived replay、migration、prompt/eval 或 scenario checklist。GeneratingScreen 不在本 plan 写入范围。

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + contract + async backend + migration + AI quality。
- **TDD 策略**: `/implement` → `/tdd`；本 owner 只为 frozen-context consumption、report validator/repair、persistence/provenance 与 reports API 建立 RED/GREEN；OpenAPI、completion、derived replay、prompt/eval、frontend 与 scenarios 只消费对应 owner gate。每个 checklist item 写明聚焦断言入口。
- **BDD 策略**: `BDD.REPORT.GENERATE.001` 由代码层 owner tests 验证 frozen-context generation/repair/persistence/replay 行为，并由仓库根 `make test` 统一回归；`E2E.P0.099` 仅作为真实 frontend/backend/provider report/generating API/UI 与六图的独立 handoff，只有显式真实运行后才产生 PASS。provider/judge reliability 仍属于独立 eval gate。
- **替代验证 gate**: prompt/schema drift、migration up/down/up、OpenAPI/codegen/fixture、privacy/negative lint、focused tests 与独立内容核对表；阶段完成由根 `make test` 统一回归 backend/frontend 单测。

## 4 Coverage Matrix

| Source | Category | Phase | Verification | Negative |
|--------|----------|-------|--------------|----------|
| spec C-1 | cross-owner frozen consumer | 6 | consume backend-practice/002 owner artifact + review load tests | review rebuilds mutable JD/Resume context |
| spec C-2 | primary / contract | 7 | prompt/schema/service/persistence/OpenAPI tests | numeric score / backend average / fallback action |
| spec C-3/C-4 | grounding / boundary | 7-8 | deterministic user-anchor tests + context-aware pending-question eval/UAT | assistant-only anchor / ability over-inference |
| spec C-5 | failure/recovery + action-scoped retry | 7-8 | per-user-action `1+3` matrix + exact `10s/20s/40s` wait recorder | persisted retry count / cross-action carry-over / partial ready |
| spec C-6 | replay handoff | 8 | consume backend-practice/004 code-level owner markers | review duplicates create-plan/focus authority |
| spec C-7 | privacy/security | 6-8 | redaction/isolation/PII negative gates | context in job/outbox/audit/log/state artifacts |
| spec C-8 | model quality / UX | 8 | independent context-aware judge/fact audit + P0.099 real UI | reused output anchors / eval or screenshot alone claiming both layers |
| spec C-9 | deep-link truth | 6-8 | frozen API context + route-tamper frontend scenario | URL/current mutable entity authority |
| spec C-10 | judge retry boundary | 8 | provider/protocol-invalid retry + valid-negative terminal code/eval tests | resampling a valid rejection to PASS |
| spec C-12 | canonical-round overview | 10 | focused handler/store/API/consumer tests | paginated full-report list / mutable round inference / partial response |
| current-scope | regression | 8 | repo negative search | `dimension_scores`, `retry_round`, stale question fields |

## 5 实施步骤

### Phase 1-5: Conversation-level baseline（已交付）

既有阶段保留为历史：conversation-level contract、generate/store、read/replay、privacy 与 1.0-5.0 score remediation 已完成；Phase 6 起以当前设计替代隐藏 score 合同。

### Phase 6: Frozen context and direct contract

#### 6.1 OpenAPI / shared / fixture direct semantic shape

在 B2 accepted ADR 后，将 FeedbackReport 改为 `summary + frozen context + dimension code/label/status/confidence + dimensionCode evidence + retryFocusDimensionCodes`，删除 request `focusCompetencyCodes` 与无引用 `DimensionResult`，关闭 report schemas 的 additional properties；同步 generated artifacts、fixtures、mapper 与 merge-base breaking audit，不保留旧 shape 兼容层。

#### 6.2 Completion-time report context snapshot

B4 migration 提供 `feedback_reports.generation_context`、ready summary 与 dimension-code naming；它不保存产品重试计数。`backend-practice/002` completion 事务唯一负责冻结 TargetJob/Resume/Round/Plan/session-language/message coordinate。本 owner 先验证/消费 `practice-completion-evidence.v1`，再让 LoadReportContext 只读取快照与 terminal ordered messages并校验坐标；禁止查询 mutable entity 重建或补写快照。开发数据 reset 后只支持 current shape，缺快照统一 fail closed。

#### 6.3 Trust boundary and real provenance

Prompt 使用可信 system policy + untrusted JSON；generation persistence 使用实际 PromptResolution/AICallMeta 与 `report-context.v1` data-source coordinate。

#### 6.4 Executable input/output budget

This historical phase first used committed synthetic input-boundary files plus current-schema worst-case zh/en outputs. Phase 11 supersedes that input-file technique: `backend/internal/review/testdata/report-boundary/` retains only the current-schema output fixtures and their manifest；default-sized framed inputs are not reconstructed, and no boundary-ready marker is emitted.

Freeze full content and measure the final UTF-8 policy+context payload immediately before the provider call. The historical package-local threshold/default-size test is superseded by Phase 11's small injected provider call/no-call guard. A3 independently validates the six active profile output budgets and report context-window configuration through its loader contract and canonical coverage lint; it does not derive token capacity from this byte measurement. TPM remains a separate throughput setting.

### Phase 7: Direct semantics, grounding, repair and persistence

#### 7.1 Prompt/schema direct output

删除 candidate numeric score 与 runtime evaluator rubric 注入；模型直接输出最终业务 shape、内部 user-message anchors、report-local dimension focus 和严格 action enum。

#### 7.2 Business invariant validator

Action `maxLength=200` code points只作fuse。纯label schema200和/或24/64 violations走`action_labels`；targeted repair内部生成目标为18/52，修复后同时满足200+24/64；其它schema或混合semantic才`whole_report`。

覆盖 closed JSON、明确长度/code/anchor 边界、每维 evidence、strong↔highlight、needs_work↔issue、issue→needs_work、focus↔retry action、next-round successor、readiness/language/safety；镜像 lower-tier 首 action retry 且禁止 next-round、每种 action type 最多一次、lower-tier 恰好一个 retry 并可选一个 review、focus exact-set、basically-ready no-needs-work、well-prepared 两维/全 strong-high/两条 user evidence、low-confidence 非负面等机械不变量。不存在 retry 时 focus 为空；存在 retry 时，只有最终恰好一个 `answer_depth` brief issue 或恰好一个 `answer_relevance` control-only issue 可为空，其他 focus 必须精确等于全部 same-code needs-work issue codes 的升序唯一集合，`I >= 2` empty 无条件 invalid。两类 exception 的文本语义和 type-specific action support 属于 prompt/judge gate：focused retry 首 label 必须为每个 selected focus code 命名至少一个直接引用 missing behavior，不能只写 umbrella term；review 只复核 cited positive/explicit evidence-limit 且不虚构 artifact/gap/new scenario/transfer task；next 只在 readiness/hasNextRound 允许；所有类型禁止未引用的新具体方案。validator 不用文本启发式猜 blocking/nonblocking 或 label 因果。assistant/未回答追问不能成为 anchor，其自然语言过度推断由 judge/UAT 验证。

#### 7.3 Action-scoped generation retries

每次用户发起 report generation 动作时，Service 创建会话内 retry context：initial 后最多重试3次；invalid后每轮动态scope并完整复验，retryable provider/protocol failure重试同一scope；三次重试前精确等待10s/20s/40s。动作返回后计数销毁，用户重新操作从0计数。`feedback_reports` 不保存计数；`async_jobs.attempts/max_attempts`只负责lease与基础设施恢复，不参与产品retry计算。当前 OpenAPI 不提供 failed-report regenerate endpoint，因此本 plan 只证明独立 `GenerateReport` invocation 会重置，不虚称已存在用户可见重试入口。

每次claim的`AsyncJob.Attempts`是lease generation。Success persistence必须在同一事务先校验/锁定`async_jobs.status='running' AND attempts=claimed`，再写ready report、outbox、audit与job succeeded；failure persistence也必须在写report/outbox/audit前校验同一generation，随后kernel finalize再次fence。Reaper/takeover后旧worker的success/retry/failure全部typed stale，不能产生任何report/outbox/audit/job副作用。

#### 7.4 Lossless persistence/read

逐项保存/读取 summary、direct assessments/evidence/actions/focus 与实际 provenance；删除 readiness/status/confidence/default-action 推导函数。

### Phase 8: Server-owned replay focus and reliability closeout

#### 8.1 Source-report focus projection

本 plan 不实现 create-plan。它消费 `backend-practice/004` markers：closed derived request 仅含 `goal + sourceReportId`；retry 空 focus 表示通用同轮复练，非空 focus 必须 issue-backed；next 使用 frozen canonical successor 且 focus 为空；服务端派生其余 settings/identity/duration，客户端无复制字段 authority。本轮 generation 语义收紧不下放为 004 对既有 ready report 的 read-time 重校验：004 仍可从已持久化 ready report 的空 focus 派生合法通用同轮 plan。

#### 8.2 Reliability eval and contract gates

Evalkit复用产品完整 report validator，并维护相互独立的内存 generation budget=4 与 judge budget=4。Generation每轮按当前 violations选择targeted/whole scope并完整复验；judge仅对retryable provider failure或protocol/schema invalid重试。结构合法的unsupported/causal/zero-tolerance/critical negative verdict是typed content rejection并立即终止，不能重采样。两条链分别聚合usage/latency，manifest记录attempt_count/retry_count/reason/scope。

#### 8.3 Scenario and privacy evidence

代码 owner 分别验证 report happy/failure/retry、replay projection 与 evalkit generation/judge；这些 gate 作为 unit/integration/eval 独立执行，不再由 E2E shell 聚合。P0.099 只以真实 frontend/backend/provider、authenticated API、只读 DB 与六图 visual audit 绑定 current-run report/generating 资源，不消费 eval output digest。exact 24/64 属于代码层边界测试；真实截图只证明当前合法内容完整可见。

Reliability eval 曾暴露 generic replay rubric 与完整 runtime validator 未同构、injection summary 缺少直接 evidence mapping、合法 empty-focus exception 被误判等问题。修复后，evalkit 复用完整 validator，judge instruction 与 rubric 使用同一 focus/action-support 决策表，summary clause 必须映射 candidate evidence。历史 live run 结果只保存在 history/report 台账；当前 plan 不把任一 CLI/eval run 编号、部分矩阵或 marker 当作应用 E2E 证据。

### Phase 9: Action-local retry contract refresh

在不新增OpenAPI regenerate operation的前提下，按TDD删除`feedback_reports.llm_attempt_count`、pre-call durable reservation与`report_generate.max_attempts=4`产品额度耦合；注入context-aware waiter，证明每次`GenerateReport` invocation独立执行initial+最多3次retry及10s/20s/40s，返回时销毁状态，下一次独立invocation从0开始。`AsyncJob.Attempts`继续仅作lease generation，stale worker仍不得提交report/outbox/audit/job副作用。证据保留在 review/store/runner 的 focused unit/integration tests，不生成伪 E2E artifact。

### Phase 10: Canonical-round report overview

#### 10.1 Contract owner handoff

消费 OpenAPI / migration owner 对 R-A 的 current-shape handoff：route 与 `listTargetJobReports` operationId 不变，删除 `cursor/pageSize`、完整报告列表与 TargetJob `latestReportId/latest_report_id` pointer；本 owner 不复制 schema/migration checklist。

#### 10.2 Store and service selection

先以 RED 锁定每个 current canonical round 都返回、`currentReport` 与 `latestAttempt` 独立排序、newer failed/generating 不覆盖 prior ready、latest ready 可双占位，再实现 user/target/frozen-context/round-pair 校验与 deterministic tie-break。任一 invalid context/identity/ready-generatedAt 使整个 overview fail closed，不返回 partial。

#### 10.3 Handler projection and privacy

Handler 只投影 `targetJobId`、`PracticeRoundRef`、`currentReport{id,generatedAt}` 与 `latestAttempt{id,status,errorCode,createdAt}`；无 pagination/pageInfo、完整 report semantic、provenance/model/rubric。TargetJob hidden 404 与跨用户零泄漏保持一致。

#### 10.4 ReportsScreen BDD handoff and closeout

用 focused backend/frontend contract tests 证明独立 ReportsScreen 可按 TargetJob summary join 并隔离当前规划，而 Parse/Report/Generating 不消费 list operation。运行 focused review/API/store、OpenAPI/fixture/generated handoff、根 `make test`、context/docs/diff 与旧分页/`latest_report_id` scoped negative gate 后再完成本 Phase。

### Phase 11: Injected report input guard

#### 11.1 RED: current regression and configured boundary

Historical diagnosis established that the old 48,000-byte package constant rejected an observed 62,397-byte request before provider admission. Both numbers remain root-cause evidence only，not current test inputs or boundaries. Do not commit `input-*.json` boundary files、reconstruct the historical payload or construct default-sized limit/+1 materials.

#### 11.2 GREEN: injected small-limit behavior

Inject A4 `report.maxFramedInputBytes` into the report service/context builder. A focused test uses a small injected limit to prove an admitted canonical frame reaches the provider unchanged and overflow persists terminal `REPORT_CONTEXT_TOO_LARGE` before provider/repair without consuming action retry。No truncation, sampling or summary fallback is allowed；default/override/invalid remain A4-owned。

#### 11.3 Capacity and substitute gate

不保留 bytes + tokens 的跨单位容量公式。BDD 不适用：配置 guard 不新增用户流程。Report happy path、fail-closed、重试恢复与返回动作由代码 owner tests 承接，不承担默认数值证明。替代 gate 为 A4 单一 typed default/override/invalid contract、A3 六个 active profile `max_tokens >= 16384` coverage lint 与 report 1M context loader contract、小值 provider call/no-call、历史/default-sized payload zero-reconstruction、`input-*.json` zero-reference 与根 `make test`。

## 6 验收标准

- Ready report 来自冻结完整上下文，模型最终语义经严格 validator 后无损持久化/API 返回。
- 任一 evidence 有同 session candidate-user anchor；未回答追问不产生能力负面推断。
- Product generation 每次用户动作最多4次调用：计数只在该次 `GenerateReport` invocation 内存在，三次重试前依次等待10s/20s/40s；每轮完整验证与动态scope；attempt4 invalid或retryable failure结束本次动作，nonretryable零重试；新的独立 invocation 从0计数。
- Judge/evaluator独立最多4次调用：仅retryable provider或protocol/schema invalid重试；结构合法negative verdict typed terminal FAIL且不重采样。
- Focused code tests覆盖product provider transient→session retry success、invalid→targeted/whole多轮成功、attempt4仍错终止、nonretryable零重试与独立 invocation 清零；evalkit tests覆盖generation/judge max4、judge invalid retry和valid negative不retry。
- 新生成报告的 generic empty focus 只用于 exact single `answer_depth` brief 或 exact single `answer_relevance` control-only issue；其他 retry focus 精确等于全部 same-code needs-work issue codes 的升序唯一集合，禁止 subset/superset。已有 ready report 的 derived-plan 空 focus 合法性仍由 backend-practice/004 owner 合同承接。
- 200仅fuse；P0.099 desktop+390证明合法24/64完整换行，超限typed invalid/no raw；18/52不作为UI门禁。
- OpenAPI、generated、fixture、DB、prompt/schema/eval、frontend consumer 与 BDD 无旧 numeric-score/unknown-action 漂移。
- Provider/eval 可靠性结果作为独立诊断记录，不转成应用 E2E。P0.099 的 desktop/mobile zh/en full-page 截图由其自身 current-run DB/API/content/action/screenshot/report/session/context 摘要闭合，并完整覆盖满足 `<=24 whitespace words` / `<=64 Unicode code points` 的实际 action；确定性 boundary fixture 由代码层测试证明恰好24/64。
- `listTargetJobReports` 对每个当前 canonical round 返回最小 `currentReport/latestAttempt` 概览；两个指针独立、排序稳定、invalid identity/context 整体 fail closed，且不形成顶层报告中心或完整报告列表。
- Report framed-input default/override/invalid 归 A4；小型 injected guard 证明 admitted frame 进入 provider、overflow 在 provider 前 terminal fail，历史 62,397-byte input 不重建；A3 独立锁定六个 active profile 至少 16K 与 report 1M context，不宣称 byte/token 换算，TPM 不参与裁决。

## 7 风险与应对

| 风险 | 应对措施 |
|------|----------|
| LLM retry 被用成抽样到PASS | generation/judge各自硬cap4；judge valid negative终端；manifest记录reason/scope，禁止第五次调用 |
| 把动作重试误作报告终身额度 | retry context仅存在于单次GenerateReport；DB无计数列；同一动作max4，新的独立invocation从0开始 |
| context snapshot 扩大敏感面 | 仅 content-bearing report 字段保存；job/outbox/audit/log/state 全部负向扫描 |
| 完整上下文超出运行时输入 guard | 全量冻结供审计；A4 typed preflight fail，不静默截断；runtime test 使用小型注入值；A3 独立校验 16,384-token output 与 1M context 配置，不使用跨单位算式 |
| 维度 code 过早全局固化 | code 仅在单份报告内唯一，P0 不建立全局 taxonomy |
| API 改动扩散遗漏 consumer | B2 accepted ADR、merge-base diff、OpenAPI-first、closed schemas、codegen、fixtures、mapper、frontend、scenarios exact-set gate |
| 历史 PASS 继续掩盖语义缺口 | 所有 completed marker 只作线索，重新运行 current-state gates 与真实 provider UAT |
| latest attempt 覆盖最后可用报告 | current/latest 使用独立查询与排序断言；failed/generating 只能更新 latestAttempt，不能清空 currentReport |
| overview 复制可变轮次展示事实 | wire 只携带 PracticeRoundRef；ReportsScreen 从当前 TargetJob canonical summary join display，pair 异常整体 fail closed |

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 2.25 | Separate code-owned report-generation BDD from the pending explicitly run P0.099 real acceptance gate. |
| 2026-07-14 | 2.24 | Remove the deleted boundary-ready marker and byte/token capacity formula; use typed owner contracts, active-profile coverage lint, small focused guards and root `make test`. |
| 2026-07-14 | 2.23 | Revise Phase 11 to a 62,397 regression, small injected provider call/no-call and A3 16K profile setting; remove default-size BDD gates. The former capacity-formula interpretation is superseded by 2.24. |
| 2026-07-14 | 2.22 | Move the unchanged overview handoff to target-scoped ReportsScreen and focused consumer contract tests. |
| 2026-07-14 | 2.21 | Reopen Phase 10 for the R-A canonical-round report overview, minimal wire, independent ready/latest selection and fail-closed identity checks. |
| 2026-07-13 | 2.20 | Separate mechanical/semantic provider eval from application E2E and retain it as an independent reliability diagnostic. |
| 2026-07-13 | 2.19 | Add Phase 9 owner handoff for action-local retry implementation, no retry-column/job-max4 coupling and preserved lease fencing. |
| 2026-07-13 | 2.18 | Replace report-lifetime durable attempt ledger with per-user-action in-memory `1+3`; wait `10s/20s/40s`, destroy on return, and reset on a new independent generation action. |
| 2026-07-13 | 2.17 | L2：report job explicit max_attempts4 and running+claimed-attempt lease fencing across report success/failure/outbox/audit/job transactions；record run35622 aborted7/11 not PASS. |
| 2026-07-13 | 2.16 | Replace single repair/judge call with independent max4 generation/judge budgets；durable pre-call reservation、typed retry boundary、audit manifest and complete TDD/BDD matrix；run25849 aborted10/11，not PASS. |
| 2026-07-13 | 2.15 | Reuse product full semantic validator in evalkit；sole-label targeted repair，all other/mixed violations whole-report repair，full revalidation and second-invalid zero-judge；record three focused live regressions while full matrix stays pending. |
| 2026-07-13 | 2.14 | Record and repair short-conservative generic-replay rubric contradiction in independent eval. |
| 2026-07-13 | 2.14 | Finalize A：wire fuse200；semantic/UX 24 whitespace words / 64 Unicode code points；targeted action-label repair margin18/52；separate eval and P0.099 evidence. |
| 2026-07-13 | 2.13 | A-200 fuse200；keep14/40 UX and refresh independent eval plus P0.099 evidence. |
| 2026-07-13 | 2.12 | Record live FAIL and require label>120 plus language-bound violations to remain action_labels，not mechanical whole_report. |
| 2026-07-13 | 2.11 | Split runtime full semantic scoped repair from evalkit schema+14/40 repair；runner other semantic invalid is no-repair/zero-judge. |
| 2026-07-13 | 2.10 | Separate the wire fuse, provider reliability eval, P0.099 current-run screenshot audit and deterministic boundary tests. |
| 2026-07-13 | 2.9 | Require evalkit same-source output-schema validation, bounded repair, aggregate usage/latency and typed judge outcome. |
| 2026-07-13 | 2.7 | Tighten generic empty focus, multi-issue focus, lower-tier action and unsafe blocking semantics in pre-judge mechanical/redacted gates. |
| 2026-07-12 | 2.6 | Add reproducible readiness/confidence semantics, runtime cross-field rejection, full judge score bands and control-only/injection reliability cases before live acceptance. |
| 2026-07-12 | 2.5 | Linearize storage/boundary/profile/evidence markers, split legacy negative owners and make scenario verifier the sole evidence artifact writer. |
| 2026-07-12 | 2.4 | Narrow review ownership, consume 002/004 markers, allow generic empty-focus retry and lock exact backend evidence. |
| 2026-07-12 | 2.3 | Close owner DAG, immutable API context, report-local dimension focus, durable repair, closed contract and context-aware reliability acceptance. |
| 2026-07-12 | 2.2 | Reopen for frozen context, LLM direct semantics, grounding anchors, one repair, real provenance, server-owned replay focus and content-level acceptance. |
| 2026-07-12 | 2.1 | Reopen to make the report score scale explicit and fail closed before readiness calculation. |
| 2026-07-12 | 2.0 | Reopen for conversation-level report and competency-focused retry. |
