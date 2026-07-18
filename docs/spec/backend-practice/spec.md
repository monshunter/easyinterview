# Backend Practice Spec

> **版本**: 1.37
> **状态**: active
> **更新日期**: 2026-07-19

## 1 背景与目标

`backend-practice` 承接基于 JD、简历和面试轮次上下文的连续模拟面试会话。当前 P0 不再维护预设题数、当前题、追问/下一题分类或逐题状态机；AI 与用户只通过按序消息推进对话，用户主动结束后进入会话级报告生成。

当前目标：

- 会话启动时生成一条普通 assistant opening message。
- 用户消息与 assistant 回复写入统一 `practice_messages`，不标记“题目 / 回答 / 追问”。
- 用户消息的回复状态与原 `clientMessageId` 由后端持久化并通过会话读模型返回；请求失败、刷新或重挂载后仍可恢复同一条消息，而不是依赖浏览器存储。
- `pending` 回复 reservation 使用 90 秒服务端租约与单调递增 generation；进程中断、请求丢失或旧 worker 迟到时，GET / 同 ID reserve 可惰性收敛且旧 generation 不得提交。
- `baseline / retry_current_round / next_round` 只决定上下文来源与能力重点，不决定题目集合。
- 暂时禁用电话模式：语音 endpoint 必须 fail-closed，且不得调用 STT / chat / TTS provider 或写入语音事件。
- 会话结束后只生成会话级准备度、能力维度、证据和下一步建议。

## 2 当前合同

### 2.1 Operation Matrix

| operationId | HTTP 行为 | backend handler | persistence | AI / job dependency | scenario coverage |
|-------------|-----------|-----------------|-------------|---------------------|-------------------|
| `createPracticePlan` | `POST /practice/plans`，要求 `Idempotency-Key`；请求不包含 question/mode/hint/focus，可携带 `roundId` 与 report-derived `sourceReportId` | `backend/internal/api/practice.CreatePracticePlan` + `backend/internal/practice.CreatePracticePlan` | `practice_plans.round_id/round_sequence/focus_dimension_codes`, `idempotency_records`, `audit_events` | none；服务端从 TargetJob/source report 推导 sequence/focus | 当前无真实 E2E owner；root `make test` |
| `getPracticePlan` | `GET /practice/plans/{planId}`，用户隔离读取 | `backend/internal/api/practice.GetPracticePlan` | `practice_plans` | none | 当前无真实 E2E owner；root `make test` |
| `startPracticeSession` | `POST /practice/sessions`，要求 `Idempotency-Key`；无活动会话时同步创建 session 与 opening assistant message，同 user/plan 已有 `queued/running` 时恢复该活动会话 | `backend/internal/api/practice.StartPracticeSession` + `backend/internal/practice.StartPracticeSession` | `practice_sessions`, `practice_messages`, lifecycle event, outbox, `idempotency_records`, `ai_task_runs`；恢复路径只完成新幂等记录 | 新建路径使用 `practice.session.chat` + `AIClient.Complete`；恢复路径零 AI | 当前无真实 E2E owner；root `make test` + Chrome 真实 UI 验收 |
| `getPracticeSession` | `GET /practice/sessions/{sessionId}`，返回 session 与有序 messages；user message 含原 `clientMessageId/replyStatus`；读取前惰性收敛已过期 pending lease | `backend/internal/api/practice.GetPracticeSession` + `backend/internal/practice.GetPracticeSession` | `practice_sessions`, `practice_messages.client_message_id/reply_status/reply_generation/reply_lease_expires_at` | none | 当前无真实 E2E owner；root `make test` |
| `sendPracticeMessage` | `POST /practice/sessions/{sessionId}/messages`；body `clientMessageId` 幂等，成功返回唯一 user/assistant pair；失败持久化 user reply status；同 ID reserve 可接管已过期 lease | `backend/internal/api/practice.SendPracticeMessage` + `backend/internal/practice.SendPracticeMessage` + SQL reserve/fail/commit | `practice_messages.client_message_id/reply_status/reply_generation/reply_lease_expires_at`, `ai_task_runs` | `practice.session.chat`; recent ordered messages + plan/session context | 当前无真实 E2E owner；root `make test` |
| `completePracticeSession` | `POST /practice/sessions/{sessionId}/complete`，要求 `Idempotency-Key`；零回答或 pending assistant reply 拒绝，成功返回 `202 ReportWithJob` | `backend-practice/002` 的 practice handler/service/store（唯一 completion owner） | `practice_sessions`, terminal `practice_messages`, `feedback_reports.generation_context`, async job/outbox/idempotency | transaction 内无 AI；随后 `report_generate` | `E2E.P0.098` 仅真实登录、completion API 与进度刷新；其余由 root `make test` |
| `createPracticeVoiceTurn` | 当前禁用；任何请求 fail-closed 为 typed `AI_UNSUPPORTED_CAPABILITY`，不得读取音频后调用 provider | existing voice handler boundary | none | none while disabled | 当前无真实 E2E owner；root `make test` |

`listPracticeSessions` 已从当前产品与公共合同删除：没有 Workspace/Report/Practice 用户入口需要按 session 分页，保留公共列表只会制造第二套历史模型。进行中会话恢复仍由 scoped `getPracticeSession(sessionId)` 承担；完成后的复盘读取由 backend-review 的 `getReportConversation(reportId)` 通过 report-owned 唯一关系承担。项目未上线，不保留 route、fixture、generated method、handler 或 compatibility alias。

### 2.2 数据模型

`practice_plans` 只保存会话所需的稳定上下文：

- `target_job_id`
- `resume_id`
- `source_report_id`
- `goal`
- `interviewer_persona`
- `difficulty`
- `language`
- `time_budget_minutes`
- `round_id`
- `round_sequence`
- `focus_dimension_codes`
- `status`

以下字段删除：`question_budget`、`mode`。专用 hint / strict-assistance 语义不再存在。

`practice_sessions` 保存生命周期、语言和关联 ID，不保存 `turn_count`、`current_turn` 或 hint 配置。

`practice_messages` 是对话唯一正文真理源：

| 字段 | 约束 |
|------|------|
| `id` | UUID 主键 |
| `session_id` | session FK，级联删除 |
| `seq_no` | session 内单调递增，唯一 |
| `role` | `user / assistant` |
| `content` | 用户可见消息正文 |
| `client_message_id` | user message 必填；session 内唯一，作为 replay key |
| `reply_to_message_id` | assistant message 指向 user message；唯一，保证每条用户消息最多一个回复 |
| `reply_status` | user message 必填：`pending / retryable_failed / terminal_failed / complete`；assistant message 为 NULL |
| `reply_generation` | user message 必填且从 1 单调递增；每次同 ID 成功 reserve 新 generation；assistant message 为 NULL；内部并发 fence，不进入 OpenAPI |
| `reply_lease_expires_at` | 仅 `pending` user message 必填，值为 reserve 服务端时钟 + 90 秒；其余状态与 assistant message 为 NULL；内部租约，不进入 OpenAPI |
| `created_at` | server timestamp |

不存在 `PracticeTurn`、`question_text`、`question_intent`、`answer_text`、`follow_up_count`、题目状态或题目编号。

### 2.3 连续聊天 AI 边界

- 唯一 feature key 为 `practice.session.chat`。空历史生成 opening message，非空历史生成下一条普通 assistant message。
- Prompt 输入包含 persisted session language、plan goal、targetJob/resume/round context、结构化 semantic focus 和按序消息历史；简历事实源按 `parsed_text_snapshot → original_text → structured_profile` 选择第一份非空完整内容，业务代码不得做字符/token 截断；不得包含题号、题目预算、generation kind、question intent 或 follow-up count。
- backend-practice start/send 只构造 JSON 编码的 `semanticFocus` payload；`practice.session.chat/v0.2.0` prompt/schema+rubric pair、template hash、version parity、8-status dev snapshot 和 DB 000019 激活/回滚全部由 F3/002 拥有。发布前 runtime test 读取 v0.2 exact coordinate；消费 F3 final marker 后 runtime 由 `ResolveActive` 取得 v0.2 pair，v0.1/000002 保持 immutable rollback。
- Prompt 必须把不可被业务数据覆盖的 `system` 策略与不可信业务上下文分层：JD、完整简历、轮次、persona 和历史对话只能作为 JSON 编码后的 user data 传入，任何其中出现的指令式文本都不得进入或改写 system policy。JSON 编码必须覆盖引号、换行、标签和类似 prompt-injection 的内容，避免上下文逃逸。完整历史可用于保持对话连续性，但只有 persisted resume 与 candidate-authored `user` message 能建立候选人事实；`assistant` message 永远不能证明履历事实或把上一轮模型臆造放大为下一轮依据。
- `round context` 必须由 plan 的 exact `roundId/roundSequence` 回查 TargetJob summary 得到 round name/type/focus；不得把 `interviewerPersona` 当作 interview round，也不得只把 persona/generalist 文案写进 `{{interview_round}}`。start/send 必须使用同一 round 投影；新 session 对 legacy null/mismatch identity fail closed。
- `interviewerPersona` 只控制语气、视角和追问风格，不能创造候选人事实、替代真实轮次、改变完成台账或决定当前/下一轮。
- start/send 两条 store reservation 必须使用相同的 resume context 投影，并再次验证 `practice_plans.resume_id = target_jobs.resume_id`；TargetJob 改绑后，旧 plan/session 即使仍属同一用户也必须 fail closed，不能继续使用旧简历。若三种简历内容均为空，必须在 prompt resolve / AI 调用前 fail closed 为 typed `VALIDATION_FAILED`，不得注入 `resume context unavailable` 后继续生成，也不得写 assistant message。
- 面试官只能把 persisted Resume context 或 candidate-authored `user` message 中明确出现的公司、项目、产品和技术栈当作候选人事实；`assistant` history 只用于连续性，不是事实来源。不得声称简历包含实际不存在的项目，也不得继续追问仅由上一条 assistant 引入的项目。用户只说“几个项目”却未给出名称时，应先请用户命名或描述，不能自行补造项目。
- 输出保持最小结构化 envelope `{messageText}`，只用于 schema 与语言校验；它不是题目结构。
- 用户输入大小按 UTF-8 bytes 由 A4 typed config 注入：`practice.maxMessageBytes` 默认 32KiB，`practice.maxSessionTextBytes` 默认 256KiB。会话总量按已持久化 user/assistant 正文加本次 user message 计算；单条或累计 limit+1 在 reservation/AI 调用前返回 typed `VALIDATION_FAILED`，不得静默截断、写入半条 user row 或调用 provider。`maxMessageBytes <= maxSessionTextBytes` 在启动期验证。
- `finish_reason=length` 即使返回内容碰巧是合法 JSON 也视为截断的 `AI_OUTPUT_INVALID`；schema / business parse、语言不一致或 length 截断只 repair 一次，provider/config/secret/timeout/unsupported 不做 business repair。
- 二次 invalid output 返回 typed `AI_OUTPUT_INVALID`，不生成 canned message。

### 2.4 消息幂等与恢复

`startPracticeSession` 的活动会话恢复先于新 session 创建：同一 `user_id + plan_id` 已有 `queued/running` session 时，新 `Idempotency-Key` 必须绑定并返回该 session，不得把唯一索引冲突映射为 `PRACTICE_SESSION_CONFLICT`，也不得再次调用 opening LLM、追加 opening message/lifecycle event/outbox/audit 或取消旧 session。不同 user/plan 不得互相恢复；同一 key 的 fingerprint mismatch、pending/terminal 状态仍保持原 typed conflict。

不同 key 的 start 请求通过 user/plan-scoped transaction lock 串行决定“恢复或创建”，数据库 active-session 唯一索引继续作为最终防线。若恢复目标仍为 `queued`，service 以 100ms 起步、1 秒封顶的有界退避读取状态，并以该 session 的持久化更新时间为起点、最多等待 35 秒让原启动请求把它推进为 `running`；只有读到最终可进入的会话后，才把新 key 的精确响应标记为 succeeded。超过边界仍为 `queued` 时，必须以 `AI_PROVIDER_TIMEOUT` 原子收敛该 session 与当前恢复 key 为 retryable failure，使下一次 start 可以重新创建；原启动提交必须以 `status='queued'` fencing，迟到 worker 不得复活已失败会话或留下 opening message/lifecycle event/outbox/audit 事实。恢复成功最终化必须锁定 session row，再校验 `running` 并写入精确响应，因此与 completion 线性化：锁前已终态则冲突，锁后才完成则返回在提交点真实的 running 快照。等待期间原请求失败、目标进入非活动状态或 caller context 取消时不得伪造成功，也不得产生第二个 session。浏览器关闭不改变服务端生命周期；之后从任一正式入口再次开始同一 plan 时，后端恢复该活动会话。

`sendPracticeMessage` 使用三段式边界：

1. 短事务按 `(session_id, client_message_id)` reserve 或 replay user message，并拒绝前一条 user message 尚无 assistant reply 时的并发新消息。
2. 事务外执行 `practice.session.chat`。
3. 短事务写 assistant reply；`reply_to_message_id` 唯一约束防止重复生成结果落库。提交时必须再次锁定并校验 session 仍处于 `running / waiting_user_input`；若完成请求已把 session 推进到 `completing / completed`，则回滚迟到的 assistant reply 并返回 typed conflict，不得把 session 改回 `running`。

reserve 时 user message 写 `reply_status='pending'`、`reply_generation=1` 与 `reply_lease_expires_at=serverNow+90s`；同一 `clientMessageId` 从 `retryable_failed` 或已过期 `pending` 重新 reserve 时只更新原 user row，generation 原子加一并刷新 90 秒 lease，不得重复写 user message。未过期 `pending`、`terminal_failed`、`complete` 或 payload mismatch 都不得被接管；不同 `clientMessageId` 在任何 unresolved user row 尚无 assistant reply 时继续冲突。

`getPracticeSession` 必须在同一授权事务内、返回 read model 前把 `reply_status='pending' AND reply_lease_expires_at <= serverNow` 惰性收敛为 `retryable_failed`，保留原 generation 并清空 lease。同 ID reserve 也必须在 session lock 下原子完成“识别过期 → generation+1 → 新 lease”，因此恢复不依赖后台 job、浏览器轮询是否持续或进程内计时器。服务端 `Service.now` / request-scoped clock 是 lease 判断的唯一时钟；数据库 `NOW()` 与客户端时钟不得形成第二套判定。

reserve 成功必须把本次 `reply_generation` 返回给 service 内部；`CommitPracticeMessage` 与 `FailPracticeMessage` 都必须携带 expected generation，并且只允许 `pending + generation match` 的 worker 写状态或 assistant reply。旧 generation、旧 lease owner 或迟到的 Commit/Fail 返回 typed conflict 且零写入；唯一 assistant reply 与 user `complete` 仍在同一事务提交。`reply_generation/reply_lease_expires_at` 只属于 persistence/domain reservation，不进入 `PracticeMessage`、OpenAPI、URL、日志或前端状态。

成功提交唯一 assistant reply 时同事务把 user 改为 `complete` 并清空 lease。AI/provider/contract 失败必须在返回 error envelope 前按 B1 `retryable` 元数据原子写为 `retryable_failed` 或 `terminal_failed` 并清空 lease，不得留下只能靠前端猜测的无状态 user row。`getPracticeSession` 对 user message 仍只公开原 `clientMessageId/replyStatus`，刷新后前端据此重建 thinking、失败 row 与同 ID retry；URL、local/session storage、IndexedDB 与 fixture 都不能成为恢复事实源。

前端当前请求期间禁用输入和结束 CTA；`pending` 显示面试官思考，`retryable_failed` 只在原 user row 下显示 retry，`terminal_failed` 不显示同 ID retry并提供返回当前面试规划的安全恢复动作，`complete` 与唯一 assistant reply 一起收敛。前端 POST 最多等待 95 秒（覆盖 90 秒服务端 lease 与少量网络/调度余量）；超时、abort 或 transport outcome 不确定时，必须先调用 `getPracticeSession` 对账同一消息的服务端状态，再按权威 `replyStatus` 恢复 thinking、retry、terminal 或 complete；不得盲目二次提交或改用新 ID。

### 2.5 生命周期事件

`practice_session_events` 只保存生命周期事实：

- `session_started`
- `session_completed`

`session_completed` 在完成事务内与 `practice_sessions.completed_at/status='completing'`、report/job/outbox 一起提交，是轮次已完成的唯一台账事实。报告生成的 queued/generating/ready/failed 不得反向改变这项完成事实；同一 session 的完成重放不得重复贡献轮次进度。

消息正文只存在于 `practice_messages`，不得复制进 event、outbox、audit、log、metric 或 task-run payload。

#### 2.5.1 Reportable completion 唯一 owner

`backend-practice/002-event-loop-and-completion` 是 `completePracticeSession` 零回答拒绝、pending-reply 拒绝、完成事务与 `report-context.v1` 冻结的唯一 owner：

- completion 至少要求一条已提交的 candidate `user` message，且不存在等待提交的 assistant reply；零回答返回 typed `VALIDATION_FAILED`，session 保持可继续对话，不写 completion fact、report、job、outbox 或 idempotency success。
- 成功事务从同一数据库一致性视图冻结 TargetJob raw/structured data、绑定 Resume source/profile、canonical round ladder/current round、source Plan settings、session language、terminal message count/last sequence，并写入 `feedback_reports.generation_context`；事务内不调用 AI。
- completion owner 的代码层测试由根 `make test` 统一回归；`backend-review/001` 只消费持久化快照，不得重新查询 mutable TargetJob/Resume/Plan 或复制 completion 测试所有权。

### 2.6 电话模式禁用

- `practice.voice.stt.default`、`practice.voice.tts.default` 和 realtime profile 保持 disabled / unsupported。
- `createPracticeVoiceTurn` 在读取/解析业务音频前 fail-closed；禁止调用 STT、chat、TTS，禁止写 voice event / committed context / audio metadata。
- 正式前端只展示不可点击的置灰电话图标，不产生 `phone` route state。
- 通用 speech adapter 可保留；重新启用必须重新修订 Product/UI/OpenAPI owner 与真实 provider gate。

### 2.7 隐私与隔离

- 所有读取和写入按 `user_id` 隔离；跨用户访问返回 404。
- message content 只进入授权的 session/read/report 输入，不进入 URL、localStorage、outbox、audit、log、metric label、task-run payload 或 provider metadata。
- `DELETE /me` 通过 FK cascade 删除 plans、sessions、messages、events 和 reports。

### 2.8 轮次身份、创建校验与进度事实

- canonical round 只来自带完整非空 provenance 的 `TargetJob.summary.interviewRounds[]`：`promptVersion`、`rubricVersion`、`modelId`、`language`、`dataSourceVersion` 均不得为空。每个 sequence 必须是正 `int32`，按数组顺序严格递增且唯一，但允许 `1,2,4` 这类非连续目录；type 必须保持小写并属于 `hr / technical / manager / cross_functional / culture / final / other`。服务端按 sequence 派生 `round-{sequence}-{type}`，不得用 TargetJob lifecycle `status`、数组第一项、固定轮次表或时长碰撞猜测当前轮。
- baseline plan 必须选择完成台账投影出的第一个未完成轮次；`retry_current_round` 必须复用 source report 所属 plan 的 exact round pair；`next_round` 必须选择 source round 之后 canonical 列表中的紧邻项，并且该项仍是当前第一个未完成轮次。sequence 不连续时选择下一条已存在 canonical round，不硬编码 `sourceSequence + 1`。
- 请求 `roundId` 仅用于校验客户端意图；`roundSequence` 由服务端推导。请求 `resumeId` 必须等于当前用户 TargetJob 持久化绑定的 `resume_id`；新 plan 必须成对持久化 round identity，time budget 必须等于该 round 的 `durationMinutes`。即使另一份简历也属于同一用户，也不得用于该 TargetJob 创建 plan。
- baseline 已完成台账、retry/next 的 source report/session/plan 事实都必须来自同一 TargetJob 绑定 resume；wrong-resume completion、source plan 或 ready plan 必须被忽略或 fail closed，不得贡献完成前缀、授权派生计划或成为当前 plan。
- report-derived 请求只接收 `goal + sourceReportId`。`retry_current_round` 从 owned ready source report 原子投影 `retryFocusDimensionCodes`：空数组是合法的通用同轮复练；非空时每个 code 必须唯一且由同一报告的 `needs_work` dimension + issue 支持。`next_round` focus 固定为空。其余 plan settings/target/resume/round/duration 由 source plan + frozen canonical successor 派生，不能由客户端覆盖。
- retry session start/send 不把 report-local code（如 `d1`）直接当自然语言提示；focus 非空时必须按 `source_report_id` 从 immutable source report 解析 `code → label + issues[]`，只把选中维度的 code/label/issue 摘要作为 untrusted focus context 注入 `practice.session.chat`；focus 为空时执行通用同轮复练且不伪造维度提示。不复制 transcript/anchors；非空 focus 缺 label/issue/cross-ref 时 fail closed。next/baseline 不加载 report focus。
- session start 与后续 message prompt 必须从 plan pair 加载同一真实 round 的 name/type/focus，并在每次 reservation 时重新校验 plan resume 仍等于 TargetJob 当前绑定 resume；persona 是独立的面试官风格字段，不是轮次身份。
- legacy null round identity 不参与当前 plan 复用或完成投影；只有 migration/backfill owner 证明唯一匹配后才能补齐。
- 除主题/外观偏好外，轮次、计划、完成事实和当前进度均为后端业务数据。前端内存、URL、local/session storage 或 fixture 不得充当事实源。

## 3 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| API contract | `openapi-v1-contract` | message schemas、fixtures、generated client/server |
| Persistence | `db-migrations-baseline` | `practice_messages` 与 session/report schema |
| Prompt | `prompt-rubric-registry` | `practice.session.chat` schema/hash/eval |
| Frontend | `frontend-workspace-and-practice` | 单聊天窗口、消息发送、本地暂停、结束、语音 disabled UI |
| Report | `backend-review` / `frontend-report-dashboard` | 会话级维度、证据、建议和复练 |
| Voice | `practice-voice-mvp` | disabled fail-closed 边界与 generic provider 保留范围 |

## 4 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 创建 baseline plan | 有 JD 和 resume | 创建 plan | 不读写 question/mode/hint 字段 | 001 |
| C-2 | 启动会话 | plan ready | start session | 返回 opening message，无 currentTurn | 001 |
| C-3 | 连续聊天 | session running | 用户发送多条消息 | 每次只追加 user/assistant message pair，无题目分类 | 002 |
| C-4 | 幂等恢复 | assistant generation 首次失败 | 同 clientMessageId 重试 | user message 不重复，最终只有一个 assistant reply | 002 |
| C-5 | 生命周期 | session running，可能存在进行中的 AI reply | 本地暂停、恢复、完成 | 暂停只影响前端输入；完成写生命周期事实并进入 report job；迟到 reply 不得重新打开 session | 002 |
| C-6 | 无专用提示 | session running | 用户需要提示 | 通过普通聊天表达，不存在 hint event/action/count | 003 |
| C-7 | 语音禁用 | 任意 session | 点击电话按钮或调用 voice endpoint | 前端不可点击；后端 fail-closed 且零 provider/持久化副作用 | practice-voice-mvp/001 |
| C-8 | 隐私 | 完成聊天与报告 | 检查非正文存储面 | raw messages 不泄漏 | 001/002/003 |
| C-9 | 真实简历 grounding | resume 的 `structured_profile` 为空或 parse failed，但存在完整 `parsed_text_snapshot` / `original_text` | 启动会话并继续发送消息 | start/send prompt 均包含完整 source snapshot 尾部 marker；面试官不得发明简历未出现的项目；所有简历内容都为空时返回 `VALIDATION_FAILED` 且零 AI / assistant message | 001/002 |
| C-10 | 轮次持久化与完成事实 | TargetJob 有结构化轮次，可能相邻轮次时长相同、sequence 不连续、完成请求重放或存在 legacy plan | 创建 baseline/retry/next plan 并完成 session | 新 plan 持久化 exact `roundId/roundSequence`；请求轮次/预算不匹配、全轮完成和非法 source fail closed；`session_completed` 每个 round pair 去重贡献进度，report 状态不影响完成事实 | 001/002 |
| C-11 | server-owned report focus | ready source report 的 retry focus 为空、为 issue-backed dimension codes，或 source/context/request 不匹配 | 创建 retry/next plan | retry 空 focus 执行通用同轮复练；非空 focus 原子投影且全部 issue-backed，并通过 F3-owned active v0.2 `semanticFocus/{{semantic_focus_json}}` pair 传入 prompt；next 为空；客户端无 focus 输入，非法 source fail closed 且 IK replay 精确 | 004 + F3/002 |
| C-12 | 绑定、目录与 prompt 信任边界 | TargetJob 绑定 resume A；同用户另有 resume B；summary 可能缺 provenance、含大于 int32 / 非连续 sequence 或大小写错误 type；简历/JD/历史可能含指令式文本，assistant history 可能已臆造项目 | 创建 plan 并启动/继续会话 | 只有绑定 resume A 的 source/completion/ready-plan 事实有效；`1,2,4` 目录按 canonical successor 推进，溢出/非法 type/缺 provenance fail closed；system policy 不被 JSON 编码的不可信上下文覆盖，persona 只影响风格，assistant-only claim 不成为候选人事实 | 001/002 |
| C-13 | reportable completion | running session 为零回答、存在 pending reply 或已有至少一条 committed user message | 调用 complete 并重放 | 002 唯一负责拒绝不可报告 completion 或原子冻结 `report-context.v1`；review 只消费持久化快照 | 002 |
| C-14 | 刷新可恢复消息 | user message 已持久化为 pending/retryable/terminal 状态，页面刷新或重挂载 | `getPracticeSession` 后按需用同一 ID 重试 | 读模型返回原 `clientMessageId/replyStatus`；pending 继续 thinking，retryable failure 原 row 可重试，terminal failure 无 retry；成功后仅一个 assistant reply，浏览器存储不参与 | 002 |
| C-16 | 消息与会话文本边界 | owner config 提供单条与累计 UTF-8 byte limit | `sendPracticeMessage` | 注入小型边界验证 overflow 返回 `VALIDATION_FAILED`、零持久化与零 AI；默认/override/invalid 只由 typed config owner 覆盖，不分配默认大小字符串或建立配置 E2E | 002 Phase 12 |
| C-15 | pending lease 与 generation fence | G1 worker 在写 reply 前失联或迟到，90 秒 lease 已过期，随后发生 GET 或两个同 ID 并发 retry | 读取会话并 reserve G2，再释放 G1 Commit/Fail | GET 或同 ID reserve 惰性收敛过期 pending；只一个调用取得 G2；G1 Commit/Fail 均 typed conflict 且零写入；G2 最终只写一个 assistant reply | 002 |
| C-17 | 无会话列表公共入口 | 当前 UI 只需要 live session 恢复与 report-owned 复盘 | 检查 OpenAPI/generated/router/handler/fixture/mock/frontend | `listPracticeSessions` current positive surface 为零；`getPracticeSession` 保留 live recovery，完成记录只由 `getReportConversation(reportId)` 读取；无 compatibility route 或新关系表 | 001 + openapi 001/002/003 + backend-review 001 |
| C-18 | 重入开始恢复活动会话 | 同一用户同一 plan 已有 `queued/running` session，可能由关闭浏览器前的启动请求留下 | 从任一正式入口再次调用 `startPracticeSession` | running 返回既有 session 且新 key 精确幂等；queued 最多等待 35 秒，超时原子失败并允许后续重试；恢复最终化锁 session row，原启动迟到提交受 queued fence 拒绝；零重复 opening/AI/lifecycle/outbox/audit；不同 user/plan 与 fingerprint mismatch 仍隔离或冲突 | 001 Phase 9 |

## 5 关联计划

- [001-plan-and-session-orchestration](./plans/001-plan-and-session-orchestration/plan.md)
- [002-event-loop-and-completion](./plans/002-event-loop-and-completion/plan.md)
- [003-mode-policies-and-provenance](./plans/003-mode-policies-and-provenance/plan.md)
- [004-report-derived-practice-plans](./plans/004-report-derived-practice-plans/plan.md)

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-19 | 1.37 | 修复 Phase 9 并发缺口：恢复最终化锁定 session row；queued 恢复采用 35 秒边界、retryable timeout 收敛与原启动 queued fencing，防止无限轮询、迟到复活和终态竞态快照。 |
| 2026-07-18 | 1.36 | 用户批准方案 A：同 user/plan 重入 start 时恢复既有 queued/running session，不取消旧会话、不重复 opening LLM，并以 plan-scoped lock 与新 key 最终快照保持并发和幂等精确。 |
| 2026-07-15 | 1.35 | 删除无产品入口的 listPracticeSessions 公共列表；保留 getPracticeSession 进行中恢复，并把完成 transcript 读取交给 report-owned getReportConversation，不保留兼容层。 |
| 2026-07-14 | 1.34 | 新增单条与会话累计文本 typed config，并将默认值测试收敛到 typed config owner。 |
| 2026-07-14 | 1.33 | Confirm T-B/P-A recovery contract: 90-second server lease, internal reply-generation fence, GET/same-ID-reserve lazy convergence, 95-second client timeout reconciliation and terminal return-to-current-plan handoff. |
| 2026-07-13 | 1.32 | Reopen 002 for server-persisted reply status and same-client-message recovery across refresh, plus typed frontend error consumption handoff. |
| 2026-07-12 | 1.31 | 完成 server-owned report focus 的 projection、idempotency、isolation 与 privacy 合同。 |
| 2026-07-12 | 1.30 | 将 report-derived prompt handoff 固化为 backend 构造 `semanticFocus`、F3/002 提供 immutable practice v0.2 pair 并独占激活/回滚。 |
| 2026-07-12 | 1.29 | 将零回答拒绝与 `report-context.v1` completion 快照唯一归属 002；允许空 report focus 表示通用同轮复练，并修正验收编号。 |
| 2026-07-12 | 1.28 | Report-derived retry focus 改为服务端投影 report-local dimension codes；客户端 focus 输入删除，completion 冻结 report context。 |
| 2026-07-12 | 1.27 | Restrict candidate facts to the persisted resume and candidate-authored user messages; assistant history remains continuity-only and cannot amplify model-invented projects. |
| 2026-07-12 | 1.26 | Tighten TargetJob-resume binding, canonical round provenance/type/int32 constraints, non-contiguous successor semantics, and the system-policy versus JSON-encoded untrusted-context prompt boundary. |
| 2026-07-12 | 1.25 | Add normalized practice-plan round identity, server-side current/next validation, and session-completion ledger semantics for backend progress projection. |
| 2026-07-12 | 1.24 | Reopen 001/002 so start and send use the complete resume source snapshot, fail closed without evidence, and forbid invented resume projects. |
| 2026-07-12 | 1.23 | 重新打开 002：关闭 message commit 与 completion 的竞态，并强化失败恢复与生命周期合同。 |
| 2026-07-12 | 1.22 | 按奥卡姆剃刀将 Practice 从题目/turn 状态机重构为连续 message conversation；删除专用 hint/mode/question 合同并暂时禁用语音。 |
