# Backend Review Spec

> **版本**: 1.28
> **状态**: active
> **更新日期**: 2026-07-14

## 1 背景与目标

`backend-review` 将一次已完成的 Practice conversation 转换为会话级证据报告。报告必须直接依据完成时冻结的 JD、绑定简历、轮次、训练目标与完整有序会话，帮助用户判断当前准备度、理解证据、识别缺口并选择复练当前轮或进入下一轮。

当前 P0 不建立跨场次可比的固定评分内核。LLM 是报告用户可见业务语义的唯一生成者；`backend-review` 只消费 `backend-practice/002` 原子冻结的上下文，并负责 report prompt trust boundary、严格校验、单次用户生成动作内 `1 + 3` 次调用的有界恢复、状态机与报告持久化，不得复制 completion 快照逻辑，也不得通过隐藏数值分、算术平均、观察条数或 fallback 文案重新解释报告。

## 2 当前合同

### 2.1 Operation Matrix

| operation / async path | frontend consumer | persistence | AI dependency | verification |
|------------------------|-------------------|-------------|---------------|--------------|
| `getFeedbackReport` | generating/report | `feedback_reports` | none on read | focused handler/store/consumer tests + `E2E.P0.099` real API/UI |
| `listTargetJobReports` | target-scoped ReportsScreen | `feedback_reports` + current TargetJob canonical round summary | none on read | focused handler/store/consumer contract tests |
| `report_generate` | async runner | `feedback_reports`, jobs/outbox/audit/task-runs | `report.generate` | focused service/store/integration tests + independent eval gate + `E2E.P0.099` visible report |

`completePracticeSession` 与 `createPracticePlan` 是跨 owner handoff，不是本 subject 的实现 operation：前者由 `backend-practice/002` 产出 `practice-completion-evidence.v1`，后者由 `backend-practice/004` 产出 server-derived replay markers；本 subject 只消费对应 reference/marker。

### 2.2 冻结输入

`backend-practice/002` 的会话完成事务必须在 `feedback_reports.generation_context` 写入 content-bearing `report-context.v1` 快照；`backend-review` 的 owner gate 先消费 schema-valid `practice-completion-evidence.v1` / `REPORT_CONTEXT_SNAPSHOT_PASS`，再只读取该快照与 terminal ordered messages。同一 B4 migration 增加 nullable-until-ready 的 `feedback_reports.summary`；重试计数属于一次用户生成动作的会话内状态，禁止写入 `feedback_reports`。项目未上线，本次采用开发数据 reset + current-shape migration，不增加 legacy ready-report 兼容分支：

- TargetJob：稳定 ID、title/company/seniority/language、原始 JD、结构化 summary 与 requirements。
- Resume：绑定 ID、display name/language、确定性 source snapshot 与 structured profile。
- Round：canonical `roundId/sequence/type/name/focus/durationMinutes`。
- PracticePlan：stable plan ID、goal、interviewer persona、focus codes。
- Conversation coordinate：session ID、session language、完成时 message count / last sequence。

`backend-practice/002` 保证 completed session 后不可追加 `practice_messages`，并在同一数据库一致性视图校验 TargetJob/Resume/Plan 绑定关系；`backend-review` 生成器只按 `seq_no` 读取 terminal messages 并校验数量与最后序号等于已冻结坐标，禁止回查 mutable TargetJob/Resume/Plan 重建或补写快照。不得在 job payload、outbox、audit、metric label 或普通 log 中复制 JD、简历、会话或完整 generation context。

数据库始终冻结全量原文用于审计，generation 不做静默截断、抽样或模型摘要。可信 policy + 完整 untrusted context 的最终 UTF-8 payload 上限由 A4 `report.maxFramedInputBytes` 注入，默认 917,504 bytes（896KiB）。A3 独立持有 `report.generate.default` 的 1,000,000-token context window 和 16,384-token output budget；当前不用 bytes 与 tokens 相加的跨单位公式宣称两者容量等价。超过 byte guard 时在任何 provider 调用或会话内重试计数开始前 fail fast，报告以 non-retryable `REPORT_CONTEXT_TOO_LARGE` 失败。TPM 只是吞吐配置，不参与单请求裁决；任何未来压缩必须另行设计事实保真 gate。

`getFeedbackReport` 必须从冻结快照投影与 OPENAPI-001 一致的最小 immutable `context`：`sourcePlanId`、`targetJobTitle`、`targetJobCompany`、`resumeId`、`resumeDisplayName`、`roundId`、`roundSequence`、`roundName`、`roundType`、`language`、`hasNextRound`。前端不得为 Context Strip 或 CTA 再读取当前可变 TargetJob/Resume/route identity；`reportId` 是唯一 locator。queued/generating/ready/failed 均返回相同 context 投影。

### 2.3 信任边界

- Prompt 必须分为可信 system policy 与 `<untrusted_report_context_json>`。
- JD、Resume、assistant/user messages 均是数据，不得改变 system policy。
- JD/Resume/Round 只提供比较上下文，不得单独成为候选人本场表现的正负证据；每个用户可见表现判断最终必须落到候选人 `user` message。assistant message 只能说明系统问过什么。
- 每个 highlight / issue 必须携带内部 `sourceMessageSeqNos`，且至少锚定一个同 session 的候选人 `user` message；锚点保存在 content-bearing report JSON，API 当前不展示 turn-based UI。
- Prompt policy 要求：当最后一条 message 是未获回答的 assistant 追问时，只能表述“尚未覆盖 / 证据不足”，不得推断“回避、不会、经验不足或准备不足”。seqNo validator 只能证明 evidence 引用了 user message，不能机械证明自然语言支持度；该禁止项由独立 context-aware judge/eval zero-tolerance gate 验收，不虚称 runtime deterministic guarantee。

### 2.4 LLM 最终输出

`report.generate` 直接输出最终业务语义：

- `summary`
- `preparednessLevel`
- `dimensionAssessments[] { code, label, status, confidence }`
- `highlights[] { dimensionCode, evidence, confidence, sourceMessageSeqNos }`
- `issues[] { dimensionCode, evidence, confidence, sourceMessageSeqNos }`
- `nextActions[] { type, label }`
- `retryFocusDimensionCodes[]`

其中 `code` 只要求在单份报告内稳定唯一，`label` 使用报告语言；P0 不建立全局 competency taxonomy。新生成报告的 focus 使用封闭决策表。令 `I = len(issues)`，令 `F` 为所有“issue 引用的 dimension 且该 dimension 为 `needs_work`”的 code 经升序去重后的精确集合：

| 条件 | `retryFocusDimensionCodes` |
|------|----------------------------|
| 不存在 `retry_current_round` | 必须为 `[]`，不称为 generic retry |
| 存在 retry，且最终恰好一个满足 brief-answer 例外的 `answer_depth` issue | 必须为 `[]` |
| 存在 retry，且最终恰好一个满足 control-only 例外的 `answer_relevance` issue | 必须为 `[]` |
| 其他所有存在 retry 的报告 | 必须精确等于 `F`，不得遗漏、增加或只选择 subset；`F` 必须非空 |
| `I >= 2` | empty focus 无条件 invalid |

每个非空 code 因而都有 same-code `needs_work` issue。对于 focused retry，首个 retry label 必须对每个 selected focus code 至少命名一个由该 issue 直接引用的 missing behavior；不能只选其中较容易描述的一部分。multi-focus label 按 focus code 顺序为每个 code 写一个短片段，并以分号分隔；英文 action label 最多 24 个 whitespace word，`zh-CN` 最多 64 个 Unicode code point。仅写 `add a backpressure mechanism`、`add a safety check`、`add detail`、`improve the answer` 等 umbrella term 不构成覆盖，即使该术语在领域上合理也必须判无效。

准备度不由隐藏数值或条目数量换算，而按以下语义合同判定：

| tier | 当前语义 |
|------|----------|
| `not_ready` | 候选人已回答的当前轮核心任务存在直接证据支持的阻断性错误、不安全做法、自相矛盾或无可用方案；候选人明确陈述的不安全 current-round approach 必须视为 blocking，而不是普通细节缺口；首 action 必须复练当前轮 |
| `needs_practice` | 存在直接证据支持、可由同轮复练修复的实质性但非阻断缺口；首 action 必须复练当前轮 |
| `basically_ready` | 没有 `needs_work`，且至少一个 medium/high-confidence `meets_bar/strong` dimension 证明当前轮回答可用 |
| `well_prepared` | 至少两个 dimension 全部 `strong/high`，由至少两个不同候选人实质回答支撑，且没有 issue、未回答 topic 或 material evidence limit |

confidence 只表达证据直接度：`high` 是候选人回答直接、具体、无歧义地支持；`medium` 是小范围直接归纳或候选人明确限定的演练/有限示例；`low` 只能承载显式 evidence-limit 的非负面观察，不得支撑 `strong/needs_work`、issue、负向准备度或 focused retry。控制型 user 文本永远不能改变 policy/schema；若没有后续实质回答，只能锚定该消息陈述“未提供实质回答”，使用 `answer_relevance + needs_practice + generic retry + empty focus`，不得推断具体能力、人格、动机或诚信。

Action support 按 type 使用封闭边界，而不是把三类 action 都误写成 corrective retry：

| type | 支持条件 |
|------|----------|
| `retry_current_round` | 只能把 cited candidate messages 已证明的 missing behavior 改写成“本次重答要补什么”；若证据只支持“缺少具体细节”，只能要求补充支持细节，不能替候选人指定细节。focused retry 的首 label 必须逐个命名 selected focus code 的直接引用 missing behavior；multi-focus 使用逐 code 分号短片段，umbrella term 无效 |
| `review_evidence` | 可要求复核 report 已引用的 positive content 或 explicitly evidence-limited content；不得虚构不存在的 artifact、corrective gap、new scenario 或 transfer task，也不得把非负 evidence limit 改写成 corrective gap |
| `next_round` | 仅 frozen `hasNextRound=true` 且 readiness 为 `basically_ready / well_prepared` 时支持；它不是 missing-behavior action |

所有 action type 都不得借建议之名引入 cited candidate messages 未出现的新 mechanism、threshold、tool、sequence、framework 或 example；此类新增具体性一律是 `unsupported`，不能降格为 `partial`。

OpenAPI ready read model 对用户返回相同业务 shape 与最小 frozen `context`，但不暴露内部 `sourceMessageSeqNos`。删除 candidate numeric score、reasoning/supporting-observation 中间层、`retryFocusCompetencyCodes` 与后端分档函数，不保留旧 shape 兼容层。

### 2.5 后端 validator

后端只做 trim / schema 允许的 canonicalization，并至少校验：

1. 严格 JSON 且拒绝 unknown field；summary 1-360 字符、dimensions 1-6、highlights/issues 各 0-4 且总 evidence 1-6、actions 1-2、retry focus 0-6。
2. dimension/action label 分别 1-48 / 1-200 code points；200 只属 wire/schema fuse。Backend 执行 English `<=24` whitespace words / zh-CN `<=64` Unicode code points；English delimiter 必须精确镜像 frontend ECMAScript `/\s/u`（含 U+FEFF、不含 U+0085），不能用 `strings.Fields` / `unicode.IsSpace` 近似。超界不得持久化 ready。
3. 每个 dimension 至少被一条 evidence 引用；每个 `strong` 至少有一条 highlight，每个 `needs_work` 至少有一条 issue，且每个 issue 必须引用 `needs_work`。`not_ready / needs_practice` 至少包含一个 `needs_work`，首 action 为 `retry_current_round`，禁止 `next_round`；`basically_ready` 禁止 `needs_work` 并至少有一个 medium/high-confidence 可用 dimension；`well_prepared` 必须满足“两维以上、全部 strong/high、无 issue、至少两个不同 user seqNo”；`low` 禁止用于 strong/needs_work/issue。validator 只镜像这些可机械 cross-field invariant，不用文本启发式猜 blocking/nonblocking 语义。
4. `sourceMessageSeqNos` 为升序、唯一、正整数且不超过 frozen last sequence；每个序号都必须指向同 session 的候选人 `user` message。assistant message（包括 terminal unanswered question）永远不能作为 anchor。
5. `retryFocusDimensionCodes` 升序去重。不存在 retry 时必须为空；存在 retry 时，只有最终恰好一个 `answer_depth` brief issue 或恰好一个 `answer_relevance` control-only issue 可为空，具体文本是否满足对应语义例外由 prompt/judge 审计。除此以外 focus 必须精确等于全部 same-code `needs_work` issue codes 的升序唯一集合，不得是 subset/superset；`I >= 2` 时 empty 无条件 invalid。
6. action type 只能是 `retry_current_round / next_round / review_evidence`，每种类型最多出现一次且顺序表达推荐优先级；所有 ready report 至少有一个 action。lower-tier report 必须恰好有一个 `retry_current_round`，可选再有一个 `review_evidence`，不得有 `next_round`；`next_round` 仅在 frozen canonical round 存在 successor 且准备度为 `basically_ready / well_prepared` 时合法。
7. 禁止录用概率、候选人排名、击败比例，以及基于语速、停顿、情绪、人格等当前不存在输入的判断；禁止复制任一 message 的连续 120 字符以上正文。
8. 输出内容语言必须与 frozen session language 一致。结构、引用、枚举和可机械识别禁区由 runtime validator 判定；“语义是否真正受引用支持、summary 是否引入新事实、focused retry 首 label 是否逐个命名 selected focus code 的直接引用 missing behavior 而非 umbrella term、review 是否只复核 cited positive/evidence-limit 且不虚构 artifact/gap/new scenario/transfer task、next 是否满足 readiness/hasNextRound、所有 action 是否避免未引用的新具体方案”由 context-aware offline judge + 人工事实核对表判定，不伪装成 seqNo 可完全证明的确定性检查。

任何不一致均返回 `AI_OUTPUT_INVALID`，不得静默删除字段、改档位、猜 confidence 或补写用户可见 action。

### 2.6 Repair、失败与 provenance

- 若当前 all-label violation set 只含 schema maxLength200 和/或 24/64，则 `action_labels`；targeted repair prompt 使用 English `<=18` whitespace words / zh-CN `<=52` Unicode code points 的内部生成余量。其它任意 schema、semantic 或 mixed violation 使用 `whole_report`；下一轮根据新输出重新计算 violations 与 scope。
- 每次用户发起一次 report generation 动作时，`GenerateReport` 新建只存在于该次调用会话内的 retry context：1 次 initial + 最多 3 次 retry，总调用数最多 4。动作返回后计数销毁；用户重新发起动作时必须从 0 重新计数。`feedback_reports` 不得保存 `llm_attempt_count`、累计失败次数或“该报告终身不可再试”的额度。
- 同一次动作内，invalid output 与 retryable provider/protocol failure 都消耗一次本地调用并在下一次调用前依次等待 `10s/20s/40s`；attempt4 仍失败即结束本次动作。non-retryable config/secret/unsupported/context-too-large/cancel 立即终止。等待必须响应 context cancellation，测试通过注入无等待 recorder 验证精确序列。
- `async_jobs.attempts` 只表示 runner lease/基础设施执行代次，不是产品 retry count。success/failure 持久化仍必须先验证当前 `async_jobs.status='running' AND attempts=claimed.Attempts`，并保持 report/outbox/audit/job 的既有事务与 stale-worker fencing；不得再用该 lease generation 预占或累计产品 LLM retry budget。
- 每轮输出都复用同一个产品完整 validator，并按该轮当前 violations 重新选择 scope：sole action-label schema200/24-64 使用 `action_labels`，targeted prompt 内部目标18/52且只 merge labels；其它任意 schema、semantic 或 mixed 使用 `whole_report`。Initial output、每次 targeted merge 与每次 whole-report replacement 都执行完整 schema+semantic 复验；attempt4 仍 invalid 才 terminal fail-close，不持久化 partial ready。
- Report 的 `10s/20s/40s` 属于单次用户动作内的 LLM 重试节奏，不复用 runner business job 的 `10s/20s/40s/80s` 调度，也不使用 outbox/infra 的 `30s/2m/10m/1h/6h`。只有尚未调用 provider 的基础设施失败可交给 runner 恢复；不得把 runner replay 当作同一产品重试计数的持久化延续。
- 生成审计以脱敏 bounded coordinate 记录 `attempt_count`、`retry_count`、每轮 `reason` 与 `scope`，并聚合所有调用的 token usage 与 latency；不得保存 raw prompt/output、secret 或候选人正文。label-only 结果只原样 merge 到目标 label，非 label 字段保持不变；服务端不截断、压缩、代写或启发式改写。
- Evalkit 与产品 runtime 都使用动作会话内 generation budget=4 并复用相同 validator/scope 状态机；manifest 输出 attempt/retry/reason/scope 脱敏审计和聚合 usage/latency，但不得把它描述为 report 生命周期持久化额度。
- Judge/evaluator 使用与 generation 相互独立的 budget=4。仅 retryable provider failure 或 judge protocol/schema invalid 可以再次调用；每次调用都消耗 judge attempt 并聚合 usage/latency。一个结构合法的 unsupported、causal mismatch、zero-tolerance 或 critical negative verdict 是有效的 terminal content rejection，必须一次失败并禁止重新抽样到 PASS。实现与 manifest 必须把 retryable `protocol invalid` 和 terminal `content rejected` 表达为不同 typed outcome。
- 完整 payload 超过注入上限时以 `REPORT_CONTEXT_TOO_LARGE` terminal fail；provider 不被调用，前端只提供返回动作。用小型注入上限验证 admitted/overflow 控制流；历史 62,397-byte 症状只作为根因证据，不重建为测试材料，默认上限的 exact/+1 也不作为场景或大材料测试要求。
- ready report 必须持久化实际 `PromptResolution` 与 `AICallMeta` 的 prompt/rubric/model/provider/language/feature/data-source coordinates；禁止使用 store 硬编码占位值。

### 2.7 复练事实源

- 客户端派生 plan 的 closed request 只提交 `goal + sourceReportId`；baseline request 继续使用独立 non-focus shape。API 删除 `focusCompetencyCodes` 输入，客户端不得在 URL 或请求中重传 persona/difficulty/language/time budget/target/resume/round/focus 事实。
- `backend-practice/004` 唯一负责 server derivation：retry_current_round 复用 source plan settings/current round，并投影 source report 的 focus；空 focus 是合法通用同轮复练，非空必须 issue-backed。next_round 复用 source persona/difficulty/language，使用 frozen canonical successor 的 round/duration且 focus 为空。`backend-review` 只消费 `REPORT_GENERIC_RETRY_PASS` / `REPORT_DERIVED_FOCUS_PASS` / `REPORT_DERIVED_ISOLATION_PASS`，不复制 create-plan 实现或 checklist。
- 本版收紧的是新 report generation 与 quality acceptance，不把 prompt/judge 的“单一宽泛 issue 才可生成 empty focus”规则复制为 `backend-practice/004` 的 read-time validator。对于已经持久化为 ready 的 source report，004 仍按自身 owner 合同将空 focus 视为合法通用 derived-plan 输入；review owner 不在派生时追溯修改或拒绝该 ready report。
- source report 缺失、跨用户、target/resume/round 不匹配、非 ready、缺 current generation context，或非空 focus 含 unsupported/duplicate/non-needs-work code 时 fail closed；空 focus 本身合法。

### 2.8 Report-quality rubric 与评测

`config/rubrics/report.generate` 只用于 offline/live judge，不注入 runtime prompt，也不作为候选人能力 taxonomy。可靠性验证是独立 code/eval gate，不是应用 E2E：generation budget=4 内每轮复用完整 validator，只有完整 schema+semantic validation 通过才进入独立 judge budget=4。Judge transport/protocol/schema retry 与业务 content rejection 必须 typed 分离；结构合法的 content rejection 是终端 FAIL，严禁重采样。逐项 verdict 覆盖 summary、preparedness、每个 dimension/highlight/issue/action 与 retry focus，并按 `supported / partial / unsupported` 判定。失败诊断只保留有界计数、reason/scope 与 digest，不保留 report/context/transcript 正文。

调用 judge 前，机械 gate 必须拒绝非例外 empty focus、focus 与全部 needs-work same-code issue 集合不精确相等、`I >= 2` empty focus 和重复 action type，避免让 judge 为结构性 invalid 输出背书。固定五类、重复采样与 blind review 可作为独立可靠性诊断，但不得被称作 E2E 或成为 `E2E.P0.099` 的前置条件。

### 2.9 TargetJob 轮次报告概览

`GET /targets/{targetJobId}/reports` 与 operationId `listTargetJobReports` 保持不变，但返回合同从“分页的完整报告记录列表”收敛为当前 TargetJob 的 canonical round 概览。请求不再接受 `cursor/pageSize`；响应是 closed object `TargetJobReportsOverview`，只包含 `targetJobId` 与按当前 `TargetJob.summary.interviewRounds[]` canonical 顺序排列的 `rounds[]`：

```text
TargetJobReportsOverview
├─ targetJobId: uuid
└─ rounds: TargetJobReportRoundOverview[]
   ├─ round: PracticeRoundRef
   ├─ currentReport: { id: uuid, generatedAt: date-time } | null
   └─ latestAttempt: { id: uuid, status: ReportStatus, errorCode: ApiErrorCode | null, createdAt: date-time } | null
```

所有 object 均为 closed，列出的 properties 均 required；可空性只通过 `currentReport | null`、`latestAttempt | null` 与 `errorCode | null` 显式表达。

- `rounds[]` 必须逐项覆盖当前 canonical round，即使该轮 `currentReport` 与 `latestAttempt` 都为 `null`；显示名称、type、duration、focus 等信息由独立 ReportsScreen 将 `PracticeRoundRef` 与当前 TargetJob summary join，overview 不复制可变展示字段。
- `currentReport` 只从 owned、同 TargetJob、同 canonical round pair、`status=ready` 且 `generated_at IS NOT NULL` 的报告中选择，稳定顺序为 `generated_at DESC, created_at DESC, id DESC`。
- `latestAttempt` 从同一 ownership / TargetJob / canonical round 边界内的全部状态选择，稳定顺序为 `created_at DESC, id DESC`；`errorCode` 必须显式 nullable。
- 两个指针彼此独立：较新的 queued/generating/failed attempt 不得覆盖较早但仍有效的 ready `currentReport`；最新 attempt 自身 ready 时允许同一 report 同时出现在两个位置。
- TargetJob 不存在、已删除或非当前用户所有时沿用 hidden 404。当前 TargetJob canonical summary 无效、report frozen context 缺失/无效、row user/target/session identity 不一致、冻结 round pair 不属于当前 canonical rounds，或 ready row 缺 `generated_at` 时，整个 overview fail closed；不得返回 partial rounds、从 mutable row/URL 推断 round、或使用 fallback identity。
- 本 operation 只服务 target-scoped ReportsScreen，不承载完整报告内容、summary/readiness/provenance/model/rubric，不建立全局/跨规划 Report Center、timeline、完整历史或第二套报告详情。Parse、Report 与 Generating 不消费该 operation；后两者继续只通过 `reportId` 调用 `getFeedbackReport`。

## 3 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| Practice transcript / completion snapshot | `backend-practice/002` | terminal messages、零回答/pending-reply gate、`generation_context` 与 completion owner artifact；review 只消费 |
| API schema | `openapi-v1-contract` | direct semantic FeedbackReport shape |
| DB | `db-migrations-baseline` | report context snapshot 与 direct semantic JSON |
| Prompt / eval | `prompt-rubric-registry` | runtime generation policy、output schema、judge-only rubric |
| UI | `frontend-report-dashboard` | honest generating、summary/dimensions/evidence/actions |

本次 001 只拥有 `backend/internal/api/reports`、`backend/internal/review`、`backend/internal/store/review` 的 report-specific 实现。B2/B4/F3、`backend-practice/002` completion、`backend-practice/004` replay、frontend 与 scenario owner 通过 context references + named marker handoff；本 subject 不把跨 owner package 纳入自身 discovery 或复制其 checklist。GeneratingScreen 已正式转交 `frontend-report-dashboard/001`。

## 4 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 冻结生成消费 | 002 已产出 completion owner artifact，且 JD/Resume 可随后修改 | runner 校验 marker/坐标并生成报告 | payload 只使用完成时快照与 terminal messages，不漂移、不重建 completion context | 001（消费）/ backend-practice 002（产出） |
| C-2 | 最终语义 | 模型返回 direct report shape | validate/persist/read | API 逐项无损返回，无隐藏数值重算 | 001 |
| C-3 | Grounding | candidate user message 支持判断 | 生成报告 | 每个 evidence 有合法内部 anchor | 001 |
| C-4 | 未回答追问 | 最后一条为 assistant 新追问 | prompt + context-aware quality gate | 只标记未覆盖/证据不足；任一能力负面推断使 eval/UAT 失败 | 001 |
| C-5 | Generation retry / repair / fail closed | 单次用户动作的会话内 retry context | runtime/evalkit执行 | 每轮按当前violations选择action_labels/whole_report并完整复验；最多4调用，重试前依次等待10s/20s/40s；attempt4 invalid/provider failure结束本次动作；用户重新操作从0计数；无report级持久化额度 | 001 |
| C-6 | 复练当前轮 | ready report focus 为空或含 issue-backed needs-work codes | 创建 retry plan | 004 服务端派生通用同轮或定向 focus，客户端不能覆盖；review 只消费 marker | backend-practice 004 |
| C-7 | 隐私与隔离 | 跨用户或非内容存储面 | 读取/生成/审计 | 404/fail closed 且无 raw context 泄漏 | 001 |
| C-8 | 可靠性与真实 UI | 独立 code/eval inputs + P0.099 current run | eval + BDD | code/eval gate 验机械与语义可靠性；P0.099 只验真实 report/generating API/UI、desktop+390 完整显示与 current-run digest，不互相冒充 | 001 |
| C-9 | 深链事实源 | 仅有 reportId 或 URL 带冲突 identity/status | 读取报告/点击 CTA | 状态、Context Strip 与动作 identity 均来自 frozen report context，route 不能覆盖 | 001 |
| C-10 | Judge bounded retry | generation output完整合法 | evaluator调用 | 独立budget=4；仅provider retryable或protocol/schema invalid重试；结构合法negative verdict typed content-rejected并终端FAIL，不重采样 | 001 + F3 004 code/eval gate |
| C-11 | Report job/fencing | report job创建、reap/takeover、迟到worker | generation persistence/finalize | async `max_attempts`只约束基础设施执行恢复；仅running+claimed attempt可写report/outbox/audit/job，stale worker零领域副作用；不得充当产品retry counter | 001 + backend-async-runner/001 |
| C-12 | TargetJob 轮次报告概览 | owned ready TargetJob 含 canonical rounds，且各轮可有 ready / queued / generating / failed 历史 | 调用 `listTargetJobReports` | 无分页地返回每个 canonical round 的 `PracticeRoundRef + currentReport + latestAttempt`；两个指针按各自稳定顺序独立选择，非法 ownership/context/round identity 整体 fail closed，不返回完整报告内容 | 001 |

## 5 关联计划

- [001-report-generation-baseline](./plans/001-report-generation-baseline/plan.md)

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 1.28 | Remove the cross-unit byte/token capacity formula; keep the 917,504-byte input guard and the 1M/16K profile contract as independently owned configuration facts. |
| 2026-07-14 | 1.27 | Raise the A3 report output setting to 16,384 and replace default-sized report/scenario boundaries with one 62,397 regression plus small injected provider call/no-call tests; the former capacity-formula interpretation is superseded by 1.28. |
| 2026-07-14 | 1.26 | 方案 A：report framed input 默认提升至 896KiB，由 A4 注入并以 1M context window 公式验证；62,397-byte 回归样本不得再被本地拒绝。 |
| 2026-07-14 | 1.25 | Move the unchanged canonical-round overview consumer from embedded Parse to the independent target-scoped ReportsScreen and focused consumer contract gate; no wire, schema, persistence, or selection change. |
| 2026-07-14 | 1.24 | 将 `listTargetJobReports` 收敛为 Parse 消费的 canonical-round overview，锁定最小 wire、独立 current/latest 排序、完整 round coverage 与 fail-closed 边界。 |
| 2026-07-13 | 1.23 | 区分机械100%、固定五类4/5语义产品验收与严格可靠性诊断；评估结果不作为应用 E2E。 |
| 2026-07-13 | 1.22 | Replace report-lifetime durable retry budget with per-user-action in-memory `1+3` and `10s/20s/40s`; new action resets, async job attempts remain infrastructure-only. |
| 2026-07-13 | 1.21 | L2：report job explicit max_attempts4；running+claimed-attempt lease fencing covers success/failure/report/outbox/audit/job；run35622 aborted7/11 not PASS. |
| 2026-07-13 | 1.20 | 用户确认generation/judge各自最多4次调用；generation durable pre-call reservation + crash-safe cap，judge仅重试provider/protocol invalid，valid negative终端FAIL；旧单次预算废止。 |
| 2026-07-13 | 1.19 | Evalkit复用产品完整semantic validator；sole-label→action_labels，其它schema/semantic/mixed→one whole_report repair；所有阶段完整复验，second invalid zero judge。 |
| 2026-07-13 | 1.18 | 方案 A 最终边界：wire fuse200 code points；semantic/UX 24 whitespace words / 64 Unicode code points；targeted action-label repair internal margin18/52。 |
| 2026-07-13 | 1.17 | A-200：wire fuse改200；sole label200/14-40仍action_labels；14/40 UX与desktop+390 gate不变。 |
| 2026-07-13 | 1.16 | Normalize all action-label schema120/14-40 violations into action_labels scope，including label>120 schema-invalid；record the live eval failure until corrected. |
| 2026-07-13 | 1.15 | Runtime 锁定整报告 / 唯一 action-length label-only repair 决策表与全量复验；evalkit 仅共享 schema+14/40 repair，runner 其它 semantic gate 不 repair且零 judge。 |
| 2026-07-13 | 1.14 | 将 action 120-char 定位为 wire/schema fuse；可靠性 eval 与 P0.099 current-run UX audit 保持独立证据链。 |
| 2026-07-13 | 1.13 | 锁定 evalkit 同源 output-schema validation、一次 `$ / output_schema_invalid` repair、second-invalid fail、generation aggregate + repair_used、judge one-shot，以及 action 长度/逐 code 分号片段。 |
| 2026-07-13 | 1.12 | 真实 judge eval 暴露 action 合同矛盾后，按 retry/review/next 分类型锁定 support 边界，并将 focus 封闭为两个 exact empty 例外或完整升序 needs-work issue-code set。 |
| 2026-07-13 | 1.11 | 初步收紧 generic focus、multi-issue 与 lower-tier action 因果；最终 focus 集合合同由 1.12 exact decision table 取代，并保留 004 对既有 ready report 空 focus 派生的 owner 边界。 |
| 2026-07-12 | 1.10 | 锁定四档准备度、三档证据 confidence、control-only 非回答与可立即执行 action 语义；runtime validator 镜像可机械 cross-field invariant，judge 接收完整分档并逐项审计 preparedness/focus。 |
| 2026-07-12 | 1.9 | 收窄 review 为 frozen-context consumer；completion/replay 改为 002/004 marker handoff，并允许空 focus 表达通用同轮复练。 |
| 2026-07-12 | 1.8 | 补齐 immutable report context 投影、跨字段 grounding、durable one-repair、report-local dimension focus、深链事实源和 context-aware 可靠性判据。 |
| 2026-07-12 | 1.7 | 重新打开 001：采用 LLM direct semantic output，冻结完整报告上下文，增加 grounding anchor、严格 validator、一次 repair、真实 provenance 与服务端 replay focus。 |
| 2026-07-12 | 1.6 | 重新打开 001：统一 candidate score 为 1.0-5.0，并在 prompt/schema/runtime 三层校验后计算 readiness。 |
| 2026-07-12 | 1.5 | 删除逐题评估与 turn focus，报告改为 conversation-level dimensions/evidence/actions。 |
