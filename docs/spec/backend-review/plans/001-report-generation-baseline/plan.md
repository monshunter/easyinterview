# 001 — Grounded Conversation Report Generation

> **版本**: 2.22
> **状态**: active
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

在既有 conversation-level report owner 内消除隐藏数值评分与后端二次语义，消费 `backend-practice/002` 已冻结的完整当轮上下文，让 LLM 直接生成可验证、可解释、可复练的最终报告；completion snapshot 与 report-derived plan 分别由 backend-practice/002、004 唯一负责，本 plan 只消费 named marker，并用真实 provider 内容核对与浏览器截图闭环。

## 2 Operation Matrix

| operationId / job | fixture | consumer | backend | persistence | AI | scenario |
|-------------------|---------|----------|---------|-------------|----|----------|
| `report_generate` job | N/A | generating poll | runner/review service/store | feedback_reports/async_jobs/outbox_events/audit_events/ai_task_runs | report.generate | P0.056/P0.058/P0.099/P0.100 |
| `getFeedbackReport` | `Reports/getFeedbackReport.json`: queued/generating/ready-needs-practice/ready-well-prepared/failed/invalid-contract/long-content | generating + report dashboard | reports handler/review read | feedback_reports + frozen context projection | none | P0.056/P0.058/P0.059/P0.099 |
| `listTargetJobReports` | `Reports/listTargetJobReports.json` | target-scoped ReportsScreen only; no Parse/generating/report consumer | reports handler/review read | feedback_reports + current TargetJob canonical summary | none | focused handler/store gate + P0.059 |

`completePracticeSession` 与 `createPracticePlan` 只作为 references/marker handoff：本 plan 不拥有其 API、fixture、handler/store 或 scenario directory。

## 2.1 Cross-owner Handoff / Execution Order

| Order | Owner | Deliverable / gate | Unblocks |
|-------|-------|--------------------|----------|
| 0 | `openapi-v1-contract/003` | accepted pre-release breaking ADR; snapshot merge-base old baseline | proposed OpenAPI edit |
| 1 | `shared-conventions-codified/001 Phase 9` | canonical non-retryable `REPORT_CONTEXT_TOO_LARGE` + `REPORT_CONTEXT_TOO_LARGE_CONVENTIONS_PASS` | B2 error-enum parity + backend terminal error |
| 2 | `openapi-v1-contract/001/003/002` | edit proposed schema while old baseline stays frozen → exact breaking/additive audit → fixtures + Go/TS codegen | backend API/consumer changes + frontend 7.1 |
| 3 | `db-migrations-baseline/001 Phase 8` + `prompt-rubric-registry/002 Phase 14 prep` | 000018 storage + `REPORT_STORAGE_V18_PASS` + v0.2 prompt/schema/multi-version `REPORT_PROMPT_V020_READY`（尚不激活） | completion snapshot + report runtime prerequisites |
| 4 | `backend-practice/002 Phase 9` | `practice-completion-evidence.v1` + `ZERO_ANSWER_COMPLETION_REJECTED_PASS` / `REPORT_CONTEXT_SNAPSHOT_PASS` / `REPORT_CONTEXT_REPLAY_PASS` | review context consumer |
| 5 | `prompt-rubric-registry/004 Phase 8` | v0.2 rubric/context-aware judge/eval + `REPORT_RUBRIC_V020_PASS` / `REPORT_CONTEXT_AWARE_EVAL_PASS` | F3 final activation + P0.100 reliability |
| 6 | `prompt-rubric-registry/002 Phase 14 activation` | consume both 004 markers, atomically activate prompt/rubric v0.2 via 000019, preserve v0.1 rollback, emit `REPORT_PROMPT_V020_PASS` | backend generation/provenance |
| 7a | `backend-review/001 Phase 6.1-6.4` | consume schema/prompt/context markers, implement frozen request builder, commit deterministic exact 48,000/48,001-byte input plus zh/en worst-case output fixtures and emit `REPORT_BOUNDARY_FIXTURES_READY` | A3 executable capacity/profile gate |
| 7b | `ai-provider-and-model-routing/003 Phase 8` | consume boundary fixtures, prove model context capacity separately from TPM, set 6,144 output budget and emit `REPORT_PROFILE_6144_PASS` | report runtime boundary |
| 7c | `backend-review/001 Phase 6.5-7.5` | consume profile marker; implement runtime byte gate, direct semantics, validator/repair/persistence/read and exact backend evidence | frontend 7.1/7.4 + 056/058 composition |
| 8 | `backend-practice/004 Phase 3` | generic-empty-focus retry, issue-backed non-empty focus, next successor, isolation + IK markers | report replay CTA/scenarios |
| 9 | registry owners (`frontend-report-dashboard` 056-059, `backend-practice` 070/072, `e2e-scenarios-p0` 099/100) | compose named backend/UI/real-provider evidence; only then 003 re-freezes baseline | final closeout |

`backend-review/001` 只拥有 `backend/internal/api/reports`、`backend/internal/review`、`backend/internal/store/review`；B2/B4/F3/backend-practice/frontend/scenario 均通过 context references 与 marker handoff。它不重复领取 completion、derived replay、migration、prompt/eval 或 scenario checklist。GeneratingScreen 不在本 plan 写入范围。

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + contract + async backend + migration + AI quality。
- **TDD 策略**: `/implement` → `/tdd`；本 owner 只为 frozen-context consumption、report validator/repair、persistence/provenance 与 reports API 建立 RED/GREEN；OpenAPI、completion、derived replay、prompt/eval、frontend 与 scenarios 只消费对应 owner gate。每个 checklist item 写明聚焦断言入口。
- **BDD 策略**: 复用并原地加强 `E2E.P0.056`、`E2E.P0.058`、`E2E.P0.070`、`E2E.P0.072`、`E2E.P0.099`、`E2E.P0.100`；覆盖 ready、invalid/repair、服务端 replay focus、真实 provider grounding 与截图。
- **替代验证 gate**: prompt/schema drift、migration up/down/up、OpenAPI/codegen/fixture、privacy/negative lint、focused Go/TS tests 与真实内容核对表共同补充 BDD。

## 4 Coverage Matrix

| Source | Category | Phase | Verification | Negative |
|--------|----------|-------|--------------|----------|
| spec C-1 | cross-owner frozen consumer | 6 | consume backend-practice/002 owner artifact + review load tests | review rebuilds mutable JD/Resume context |
| spec C-2 | primary / contract | 7 | prompt/schema/service/persistence/OpenAPI tests | numeric score / backend average / fallback action |
| spec C-3/C-4 | grounding / boundary | 7-8 | deterministic user-anchor tests + context-aware pending-question eval/UAT | assistant-only anchor / ability over-inference |
| spec C-5 | failure/recovery + action-scoped retry | 7-8 | per-user-action `1+3` matrix + exact `10s/20s/40s` wait recorder | persisted retry count / cross-action carry-over / partial ready |
| spec C-6 | replay handoff | 8 | consume backend-practice/004 + P0.070/P0.072 markers | review duplicates create-plan/focus authority |
| spec C-7 | privacy/security | 6-8 | redaction/isolation/PII negative gates | context in job/outbox/audit/log/state artifacts |
| spec C-8 | model quality / UX | 8 | context-aware judge + fact→judgment→action audit + P0.099/P0.100 | reused output anchors / contract-only or screenshot-only claim |
| spec C-9 | deep-link truth | 6-8 | frozen API context + route-tamper frontend scenario | URL/current mutable entity authority |
| spec C-10 | judge retry boundary | 8 | provider/protocol-invalid retry + valid-negative terminal tests + P0.100 | resampling a valid rejection to PASS |
| spec C-12 | canonical-round overview | 10 | focused handler/store/API tests + P0.059 composition | paginated full-report list / mutable round inference / partial response |
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

First commit synthetic, non-sensitive deterministic fixtures under `backend/internal/review/testdata/report-boundary/`: exact final framed inputs at 48,000 and 48,001 UTF-8 bytes, current-schema worst-case zh/en outputs, and a manifest containing byte counts plus SHA-256 values. A focused fixture test must reconstruct/serialize them byte-identically and emit `REPORT_BOUNDARY_FIXTURES_READY`; A3 consumes these files without owning or rewriting report business bounds.

After A3 returns `REPORT_PROFILE_6144_PASS`, freeze full content and measure the exact final UTF-8 policy+context payload immediately before provider call. At >48,000 bytes persist terminal `REPORT_CONTEXT_TOO_LARGE` without provider/repair; never truncate/summarize. Exactly 48,000 bytes reaches the provider unchanged. The context-window capacity proof, 6,144 output budget and actual-token smoke are A3-owned; TPM remains a separate throughput setting.

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

加强 P0.056/P0.058/P0.070/P0.072/P0.099/P0.100，禁止 cookie 写入 state/evidence。P0.058证明产品 generation 的单动作会话内 max4、完整复验、10s/20s/40s与独立调用清零；P0.100继续证明 evalkit generation/judge 独立max4与typed judge outcome。产品内容验收以所有最终输出机械100%和固定五类至少4/5语义PASS收口；严格P0.100继续要求5类/11次与blind review。P0.099 不消费这些 output digests；六图以脱敏 DB/API + hybrid visual audit绑定 current-run 资源，exact 24/64 由 deterministic parity 证明。

真实 run `e2e-p0-100-20260713T011140Z-36625` 的 short-conservative 第1次 judge 曾在 `$.nextActions[0]` 返回 `invalid_partial`。测试、环境与 consumer drift 均已排除；根因是 prompt/case/judge 允许 exact `answer_depth` / `answer_relevance` generic replay，而 `report_action_quality` rubric 把所有 generic action 统一降为 developing。judge instruction 与 rubric 例外已用 TDD 修复，migration 与 active DB 已同步；同 case 复测 weighted=0.82、min=0.70、零违规。该单例和后续run35103都只保留为对应历史prompt/contract证据。

后续调查继续得到三条局部证据，均不得升级为最终 PASS：

- `e2e-p0-100-20260713T014058Z-80338` 第11次把 `needs_practice` 且同时含 `retry_current_round,next_round` 的报告送入 judge。测试/环境/consumer drift 排除后，根因是 evalkit 只执行 schema+label 校验而未复用 runtime 完整 validator。代码GREEN接入full-validator，随后 run `35103` 完成5类/11次验收；该历史失败不计PASS。
- `e2e-p0-100-20260713T012359Z-59906` 的 injection summary 被判 unsupported。修复将每个 summary clause 映射到 candidate evidence，禁止 action 升级对话未声明的质量属性，并要求 `W` exact readiness；direct injection judge 连续3次PASS。
- `e2e-p0-100-20260713T013642Z-75753` short repetition3 的合法 empty focus 被 judge 误判。judge 改为 exact generic empty-focus exception supported；same digest regression 与额外5次均PASS。

`e2e-p0-100-20260713T022140Z-25849` 在11次矩阵的第10次完成后，因用户确认以generation/judge各自max4的新合同替换旧单次预算而主动中止。该run没有完成当前合同矩阵，不能作为PASS或marker证据。

`e2e-p0-100-20260713T030100Z-35622` 在第7/11次后因L2确认report job max_attempts5泄漏、lease generation fencing缺失与frontend in-flight pause恢复重置而主动中止。该run为aborted/not-PASS；三项owner RED/GREEN未通过前不得复用其部分结果或发出marker。

### Phase 9: Action-local retry contract refresh

在不新增OpenAPI regenerate operation的前提下，按TDD删除`feedback_reports.llm_attempt_count`、pre-call durable reservation与`report_generate.max_attempts=4`产品额度耦合；注入context-aware waiter，证明每次`GenerateReport` invocation独立执行initial+最多3次retry及10s/20s/40s，返回时销毁状态，下一次独立invocation从0开始。`AsyncJob.Attempts`继续仅作lease generation，stale worker仍不得提交report/outbox/audit/job副作用。P0.058升级为`report-backend-evidence.v3`，将动作调用/等待事实放入`runtime`，数据库区只保存状态与ready-column fail-closed事实。

### Phase 10: Canonical-round report overview

#### 10.1 Contract owner handoff

消费 OpenAPI / migration owner 对 R-A 的 current-shape handoff：route 与 `listTargetJobReports` operationId 不变，删除 `cursor/pageSize`、完整报告列表与 TargetJob `latestReportId/latest_report_id` pointer；本 owner 不复制 schema/migration checklist。

#### 10.2 Store and service selection

先以 RED 锁定每个 current canonical round 都返回、`currentReport` 与 `latestAttempt` 独立排序、newer failed/generating 不覆盖 prior ready、latest ready 可双占位，再实现 user/target/frozen-context/round-pair 校验与 deterministic tie-break。任一 invalid context/identity/ready-generatedAt 使整个 overview fail closed，不返回 partial。

#### 10.3 Handler projection and privacy

Handler 只投影 `targetJobId`、`PracticeRoundRef`、`currentReport{id,generatedAt}` 与 `latestAttempt{id,status,errorCode,createdAt}`；无 pagination/pageInfo、完整 report semantic、provenance/model/rubric。TargetJob hidden 404 与跨用户零泄漏保持一致。

#### 10.4 ReportsScreen BDD handoff and closeout

向 `E2E.P0.059` 提供 focused backend contract evidence；证明独立 ReportsScreen 可按 TargetJob summary join 并隔离当前规划，而 Parse/Report/Generating 不消费 list operation。运行 focused/full review/API/store、OpenAPI/fixture/generated handoff、context/docs/diff 与旧分页/`latest_report_id` scoped negative gate 后再完成本 Phase。

## 6 验收标准

- Ready report 来自冻结完整上下文，模型最终语义经严格 validator 后无损持久化/API 返回。
- 任一 evidence 有同 session candidate-user anchor；未回答追问不产生能力负面推断。
- Product generation 每次用户动作最多4次调用：计数只在该次 `GenerateReport` invocation 内存在，三次重试前依次等待10s/20s/40s；每轮完整验证与动态scope；attempt4 invalid或retryable failure结束本次动作，nonretryable零重试；新的独立 invocation 从0计数。
- Judge/evaluator独立最多4次调用：仅retryable provider或protocol/schema invalid重试；结构合法negative verdict typed terminal FAIL且不重采样。
- P0.058覆盖product provider transient→session retry success、invalid→targeted/whole多轮成功、attempt4仍错终止、nonretryable零重试与独立 invocation 清零；evalkit focused tests覆盖generation/judge max4、judge invalid retry和valid negative不retry，最终run59381也证明第9次valid negative立即终止且不重采样。
- 新生成报告的 generic empty focus 只用于 exact single `answer_depth` brief 或 exact single `answer_relevance` control-only issue；其他 retry focus 精确等于全部 same-code needs-work issue codes 的升序唯一集合，禁止 subset/superset。已有 ready report 的 derived-plan 空 focus 合法性仍由 backend-practice/004 owner 合同承接。
- 200仅fuse；P0.099 desktop+390证明合法24/64完整换行，超限typed invalid/no raw；18/52不作为UI门禁。
- OpenAPI、generated、fixture、DB、prompt/schema/eval、frontend consumer 与 BDD 无旧 numeric-score/unknown-action 漂移。
- 最终prompt真实provider run59381的机械输出9/9、语义judge8/9、固定五类4/5满足当前产品验收；strict P0.100因unsupported summary保持FAIL，未运行blind audit。P0.099 的 desktop/mobile zh/en full-page 截图由其自身 current-run DB/API/content/action/screenshot/report/session/context 摘要闭合，并完整覆盖满足 `<=24 whitespace words` / `<=64 Unicode code points` 的实际 action。确定性 boundary fixture 另行证明恰好24/64的pixel parity。
- `listTargetJobReports` 对每个当前 canonical round 返回最小 `currentReport/latestAttempt` 概览；两个指针独立、排序稳定、invalid identity/context 整体 fail closed，且不形成顶层报告中心或完整报告列表。

## 7 风险与应对

| 风险 | 应对措施 |
|------|----------|
| LLM retry 被用成抽样到PASS | generation/judge各自硬cap4；judge valid negative终端；manifest记录reason/scope，禁止第五次调用 |
| 把动作重试误作报告终身额度 | retry context仅存在于单次GenerateReport；DB无计数列；同一动作max4，新的独立invocation从0开始 |
| context snapshot 扩大敏感面 | 仅 content-bearing report 字段保存；job/outbox/audit/log/state 全部负向扫描 |
| 完整上下文超出模型预算 | 全量冻结供审计；48,000-byte preflight typed fail，不静默截断；输出合同与 6,144-token profile 做 worst-case gate |
| 维度 code 过早全局固化 | code 仅在单份报告内唯一，P0 不建立全局 taxonomy |
| API 改动扩散遗漏 consumer | B2 accepted ADR、merge-base diff、OpenAPI-first、closed schemas、codegen、fixtures、mapper、frontend、scenarios exact-set gate |
| 历史 PASS 继续掩盖语义缺口 | 所有 completed marker 只作线索，重新运行 current-state gates 与真实 provider UAT |
| latest attempt 覆盖最后可用报告 | current/latest 使用独立查询与排序断言；failed/generating 只能更新 latestAttempt，不能清空 currentReport |
| overview 复制可变轮次展示事实 | wire 只携带 PracticeRoundRef；ReportsScreen 从当前 TargetJob canonical summary join display，pair 异常整体 fail closed |

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 2.22 | Move the unchanged overview handoff to target-scoped ReportsScreen/P0.059 and require Parse/Report/Generating zero-consumer evidence. |
| 2026-07-14 | 2.21 | Reopen Phase 10 for the R-A canonical-round report overview, minimal wire, independent ready/latest selection, fail-closed identity checks and P0.016 handoff. |
| 2026-07-13 | 2.20 | Separate mechanical 100%, fixed-five 4/5 semantic product acceptance and strict P0.100 11/11 diagnostics; record final-prompt run59381 without promoting its strict FAIL. |
| 2026-07-13 | 2.19 | Add unchecked Phase 9 owner handoff for action-local retry implementation, no retry-column/job-max4 coupling, preserved lease fencing, and P0.058 evidence v3. |
| 2026-07-13 | 2.18 | Replace report-lifetime durable attempt ledger with per-user-action in-memory `1+3`; wait `10s/20s/40s`, destroy on return, and reset on a new independent generation action. |
| 2026-07-13 | 2.17 | L2：report job explicit max_attempts4 and running+claimed-attempt lease fencing across report success/failure/outbox/audit/job transactions；record run35622 aborted7/11 not PASS. |
| 2026-07-13 | 2.16 | Replace single repair/judge call with independent max4 generation/judge budgets；durable pre-call reservation、typed retry boundary、audit manifest and complete TDD/BDD matrix；run25849 aborted10/11，not PASS. |
| 2026-07-13 | 2.15 | Reuse product full semantic validator in evalkit；sole-label targeted repair，all other/mixed violations whole-report repair，full revalidation and second-invalid zero-judge；record three focused live regressions while full matrix stays pending. |
| 2026-07-13 | 2.14 | Record and repair short-conservative generic-replay rubric contradiction；single-case retest0.82/0.70 with zero violations，full P0.100 matrix remains pending. |
| 2026-07-13 | 2.14 | Finalize A：wire fuse200；semantic/UX 24 whitespace words / 64 Unicode code points；targeted action-label repair margin18/52；reopen P0.100/P0.099 evidence. |
| 2026-07-13 | 2.13 | A-200 fuse200；keep14/40 UX and reopen P0.100/P0.099 evidence for new contract. |
| 2026-07-13 | 2.12 | Record live FAIL and require label>120 plus language-bound violations to remain action_labels，not mechanical whole_report. |
| 2026-07-13 | 2.11 | Split runtime full semantic scoped repair from evalkit schema+14/40 repair；runner other semantic invalid is no-repair/zero-judge. |
| 2026-07-13 | 2.10 | Separate the 120-char fuse, P0.100 5-case/11-attempt reliability, P0.099 current-run canonical screenshot audit, and deterministic exact-14/40 parity. |
| 2026-07-13 | 2.9 | Require P0.100/evalkit same-source output-schema validation, one `$ / output_schema_invalid` product repair, aggregate generation usage/latency + repair_used, and zero judge repair/retry；bound actions to en 14 words / zh-CN 40 chars with per-code semicolon fragments. |
| 2026-07-13 | 2.7 | Tighten generic empty focus, multi-issue focus, lower-tier action and unsafe blocking semantics; add P0.100 pre-judge mechanical/redacted gates without changing 004 ready-report derivation ownership. |
| 2026-07-12 | 2.6 | Add reproducible readiness/confidence semantics, runtime cross-field rejection, full judge score bands and control-only/injection reliability cases before live acceptance. |
| 2026-07-12 | 2.5 | Linearize storage/boundary/profile/evidence markers, split legacy negative owners and make scenario verifier the sole evidence artifact writer. |
| 2026-07-12 | 2.4 | Narrow review ownership, consume 002/004 markers, allow generic empty-focus retry and lock exact 056/058 backend evidence. |
| 2026-07-12 | 2.3 | Close owner DAG, immutable API context, report-local dimension focus, durable repair, closed contract and context-aware reliability acceptance. |
| 2026-07-12 | 2.2 | Reopen for frozen context, LLM direct semantics, grounding anchors, one repair, real provenance, server-owned replay focus and content-level acceptance. |
| 2026-07-12 | 2.1 | Reopen to make the report score scale explicit and fail closed before readiness calculation. |
| 2026-07-12 | 2.0 | Reopen for conversation-level report and competency-focused retry. |
