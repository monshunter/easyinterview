# Frontend Workspace and Practice Spec

> **版本**: 1.42
> **状态**: active
> **更新日期**: 2026-07-14

## 1 背景与目标

`frontend-workspace-and-practice` 承接 `workspace` 面试规划列表与 `practice` 连续文本会话。当前 Practice 不再使用题目/turn 状态机，正式页面只保留 Top Bar、全宽 Transcript 和 Composer；电话入口暂时置灰。

`generating` 与 `report` 由 `frontend-report-dashboard` / `backend-review` 承接；本 owner 只负责 completion 返回 stable `reportId` 后的 route handoff，不拥有生成态文案、轮询或错误动作。

## 2 范围

### 2.1 Workspace

- 一级 `面试 / workspace` 始终渲染 ready TargetJob 规划卡片列表。
- 卡片展示状态、更新时间、岗位、公司和地点；主体进入 `parse` 统一详情，右上角归档，底部只有 `立即面试`。
- `workspace` 不读取或继承详情 route context，不拥有 Resume Picker、Plan Switcher 或 route-side auto start。
- 快速启动通过 generated `createPracticePlan` / `startPracticeSession` 创建或复用 plan，然后进入 `practice`。
- 卡片 round rail、`立即面试` 和 parse 当前轮只消费 backend `TargetJob.practiceProgress`；TargetJob lifecycle `status` 只用于岗位状态展示，不参与轮次推断。

### 2.2 Practice

- Route 只需要稳定 `sessionId` 与 target/plan/resume/round IDs；不使用 `mode/modality/practiceMode/hintUsed/hintCount`。
- Top Bar：真实公司/岗位、面试官角色、计时、暂停、disabled phone icon、结束并生成报告。
- `TargetJob.summary.interviewRounds[]` 只定义 canonical 轮次目录、顺序和时长；当前轮由 backend `TargetJob.practiceProgress.currentRound` 选中。启动时把该轮 `durationMinutes` 写入 `PracticePlan.timeBudgetMinutes`，Practice Top Bar 再从当前 plan 读取并显示预算，不使用固定分钟数。
- Conversation：全宽有序 Transcript + Error/Retry + Composer。
- opening message 和后续 assistant reply 统一来自 server messages，不是 QuestionCard。
- 用户输入通过 generated `sendPracticeMessage`，不提交 `turnId`，不标记 answer/hint/question。
- 提交后必须立即清空 composer，并先把该条 user message 作为当前页面的瞬时 optimistic row 加入 Transcript；不得等待 assistant response 后才与 reply 一起出现。该 row 只用于请求中的即时反馈，不写浏览器存储、不计入 Finish 资格，也不覆盖 server messages 事实源。
- `getPracticeSession` 刷新恢复完整 ordered messages，并为每条 user message 返回服务端事实 `clientMessageId` 与 `replyStatus=pending|retryable_failed|terminal_failed|complete`；assistant message 不伪造这些 user-only 字段。`pending` 继续显示思考并自动 re-read server truth，`retryable_failed` 恢复 row-local retry，`terminal_failed` 进入无 retry 的事实恢复，`complete` 只展示唯一 user/assistant pair。
- 服务端 pending reservation 的 lease 固定为 90 秒；前端 `sendPracticeMessage` 的单次 POST 等待上限固定为 95 秒。95 秒到达时必须 abort 本次 fetch，并立即用 `getPracticeSession` 对账同一 `clientMessageId`；不得盲目自动重发、生成新 ID 或把客户端时钟当成 reply status 事实源。
- 暂停/恢复只控制当前页面的 composer 与计时显示，不产生 backend 事件；结束通过 `completePracticeSession`。
- Error/Retry 必须按失败来源恢复：session loader 调用 `refresh`，message failure 使用同一 `clientMessageId` 重试 send，completion failure 使用同一 completion idempotency key 重试 finish；不得把完成重试误接到 send。
- message pending/retrying 时 composer disabled，并在 Transcript 中追加 assistant-style、`aria-live` 的面试官思考动画；成功后用 server session 原子替换 optimistic row/思考态，不能重复 user message。失败后隐藏思考态、保留原 user row，只在该 row 底部渲染 retry icon；该 icon 必须调用与 composer submit 相同的 send path，并复用原 `clientMessageId` 与原文本。
- row-local retry 只属于明确可重试的 message failure：无 HTTP response 的网络错误，或 generated `ApiClientError.apiError.retryable=true`（含 AI timeout/5xx）。OpenAPI owner 必须从 error envelope 生成 typed `ApiClientError`，保留 HTTP status、`code/requestId/retryable/details` 与 transport cause；Practice 不得解析普通 `Error.message` 字符串。intentional abort/unmount 不渲染 retry。`VALIDATION_FAILED`、auth/not-found、client-message conflict/mismatch 等终态不得渲染同 ID retry icon；它们必须走 loader/auth/session-lost 等事实恢复，重新读取 server messages 后再决定 composer 是否可用，不能让用户陷入无限重试。
- `terminal_failed` 是服务端权威终态，不显示 row-local retry，也不允许永久空白锁死。页面必须显示一条不泄漏技术细节的通用恢复说明和唯一主动作“返回当前面试规划”，精确执行 `navigate({ name: "parse", params: { targetJobId: session.targetJobId } })`；不得退回无上下文 `workspace`、不得把 `planId` 伪装为 parse 参数、不得复用 composer submit。
- 当前页面内存中的 `{text, clientMessageId, status}` 只覆盖 submit 到首次 response/read convergence 的即时反馈；一旦刷新或重挂载，恢复必须完全来自 `getPracticeSession` 的 `clientMessageId + replyStatus`，不得用 URL、localStorage、sessionStorage 或 IndexedDB 保存 retry identity。AI 失败后刷新仍必须用服务端返回的原文本与同一 `clientMessageId` 重试，最终收敛为唯一 user/reply pair。
- message failure 后 textarea 可恢复输入以保留下一条草稿，但在失败消息完成同 ID retry 前 submit 仍 disabled，并提供本地化说明；草稿不得改变 retry payload，也不得作为另一条业务消息绕过待回复状态。
- 95 秒 POST timeout 后的对账必须采用有界、可取消的 GET：若读到 `complete/pending/retryable_failed/terminal_failed`，立即采用服务端事实；若尚无该 ID 或 GET 本身失败，保留原 optimistic row 与原 `clientMessageId` 为 unresolved/retryable recovery，继续禁止新 ID submit/Finish，并允许刷新或同 ID 恢复。被 abort 的旧 POST 即使迟到返回也不得覆盖较新的对账或 retry generation。
- message optimistic row 仍处于 pending/retryable-failed/retrying/terminal-recovery、session loading、completion 进行中或 session 已进入 `completing / completed` 时结束 CTA 必须 disabled，避免 UI 主动制造 send/complete 竞态。
- 只有 server-loaded `messages` 中至少存在一条已提交的 candidate `user` message 且不存在 pending assistant reply 时，Finish 才具备前端资格。零回答时使用原生 disabled，并在控件附近展示 zh/en 本地化原因，通过稳定 `aria-describedby` 关联；route params、composer draft 或仅有 opening assistant message 均不能充当回答。
- 前端资格只减少无效操作，不是业务授权：即使绕过 UI 直接调用 `completePracticeSession`，backend 仍必须权威返回 typed `VALIDATION_FAILED`，保持 session mutable，且不创建 completion/report/job/outbox/idempotency success。
- phone icon 使用原生 disabled 控件；phone/voice route params 不得 materialize PhoneSurface。
- 规划时长是预算显示，不是自动结束条件；elapsed 可以超过预算，用户仍通过“结束并生成报告”显式完成会话。

### 2.3 Generating Handoff

- `completePracticeSession` 返回 `ReportWithJob` 后进入 `generating?reportId`。
- handoff 只携带 stable `reportId`；不携带 session/target/resume/round/status/error 业务事实。
- 进入 `generating` 后的轮询、状态、文案和动作由 `frontend-report-dashboard` 唯一承接。

### 2.4 Out of Scope

- SessionMap、“本轮题目”、题号/总题数、QuestionCard、question badge/topic/prompt。
- `PracticeTurn/currentTurn/turnCount/questionIntent` UI 消费。
- 专用 hint button/banner/event/count 与 strict/assisted switch。
- PhoneSurface、麦克风、字幕、VAD、TTS、barge-in、hangup。
- 独立 Voice route、右侧辅助栏、语音分析、跳过、会话内 persona switch。
- Report Dashboard 具体实现。

## 3 用户决策

| ID | 决策 | 当前合同 |
|----|------|----------|
| D-1 | Practice 交互模型 | 连续文本 conversation，不区分问题/回答/追问 |
| D-2 | 页面布局 | 删除左栏和 QuestionCard，只保留全宽聊天 |
| D-3 | 专用提示 | 删除；用户需要提示时发送普通消息 |
| D-4 | 电话模式 | 前端入口置灰，phone/voice params 归一为文本 |
| D-5 | 报告 handoff | 只传稳定 IDs；不传 modality/practiceMode/hint fields |
| D-6 | 轮次目录与预算来源 | `TargetJob.summary.interviewRounds[]` 定义 canonical 轮次目录、顺序与时长；sequence 必须正 int32、唯一、严格递增但允许 `1,2,4`，下一轮是数组中下一条已存在 canonical round，不是 `current.sequence + 1`。`TargetJob.practiceProgress` 决定当前/已完成轮次；`PracticePlan.timeBudgetMinutes` 保存所选轮次时长快照；重复派生 ID、未知轮次、空轮次和末轮不得回退到第一轮或固定默认轮次 |
| D-7 | 业务状态后端持久化 | 除主题/外观偏好外，轮次进度、当前轮、plan/session/report 和完成事实只来自 backend API；前端内存、URL、`localStorage`、`sessionStorage`、IndexedDB、自由文本 `nextRound` 或 fixture 不得作为事实源。`TargetJob.practiceProgress` 是卡片/详情/quick-start 的 read model；缺失或不一致时 fail closed。 |
| D-8 | Finish 最低回答门槛 | 前端只从 server-loaded messages 计算至少一条 committed candidate `user` message；零回答原生 disabled 并显示本地化可访问原因。Backend `completePracticeSession` 独立执行同一事实校验并保持最终权威。 |
| D-9 | 即时消息与失败恢复（方案 A） | user submit 后立即显示瞬时 optimistic row；服务端 `PracticeMessage` 为 user message 投影 `clientMessageId + replyStatus(pending|retryable_failed|terminal_failed|complete)`，`getPracticeSession` 在刷新后恢复 thinking/retry/terminal/complete；OpenAPI owner 生成 typed `ApiClientError.apiError`；retry 复用服务端原文与同 ID。该方案兼顾即时反馈、跨刷新幂等恢复与后端事实源；前端不持久化 retry identity、不解析 error string、不引入第二套消息事实或无限重试。 |
| D-10 | Pending 超时与对账（T-B） | backend lease 为 90 秒；frontend POST timeout 为 95 秒并 abort；随后只用同一 ID `getPracticeSession` 对账；GET/同 ID reserve 负责服务端惰性收敛。这样既覆盖服务端 lease 又不无限挂起；超时不是失败事实，新 ID 与盲目自动重发均禁止。 |
| D-11 | Terminal 恢复入口（P-A） | `terminal_failed` 展示通用安全说明 + 唯一“返回当前面试规划”CTA，精确进入 `parse(targetJobId)`；无 row retry，不退回无上下文 workspace。最小动作即可回到当前规划继续决策，避免永久锁死、错误重放和上下文丢失。 |

## 4 UI 真理源与 parity

- Workspace：`ui-design/src/screen-workspace.jsx`
- Practice：`ui-design/src/screen-practice.jsx::PracticeScreen`
- Generating：由 `frontend-report-dashboard` 独占；本 spec 仅引用 completion handoff，不修改其原型或正式屏幕。
- Shared：`ui-design/src/app.jsx`、`ui-design/src/primitives.jsx`
- Docs：`docs/ui-design/module-job-workspace.md`、`module-practice-review.md`、`report-dashboard.md`

用户可见改动必须先更新 `ui-design/`，再源级迁移到 frontend。验证必须拆分：

1. DOM/control/a11y source parity。
2. computed style/bounding box/responsive/screenshot geometry parity。
3. stale question/hint/phone positive-contract negative search。

## 5 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticePlan` | `openapi/fixtures/PracticePlans/createPracticePlan.json`: current `default`, `next-derived`, `retry-derived`, `round-mismatch` | `frontend/src/app/interview-context/startPractice.ts` | `backend/internal/api/practice.Handler.CreatePracticePlan` → `backend/internal/practice.Service.CreatePracticePlan` → `backend/internal/store/practice.SQLRepository.CreatePlan` | `practice_plans`, `idempotency_records`, `audit_events` | none | `E2E.P0.021`, `E2E.P0.057`, `E2E.P0.098` |
| `getPracticePlan` | `openapi/fixtures/PracticePlans/getPracticePlan.json`: current `default`, `legacy-null-round-identity`, `not-found` | `startPractice.ts` exact-plan reuse + `PracticeScreen` budget read | `Handler.GetPracticePlan` → `Service.GetPracticePlan` → `SQLRepository.GetPlan` | `practice_plans` read | none | `E2E.P0.021`, `E2E.P0.045`, `E2E.P0.098` |
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json`: current `default`, `empty`, `one-job`, `twelve-plus`, `not-started-progress`, `all-completed-progress`, `prototype-baseline` | Home/Workspace plan rails | `backend/internal/targetjob.Handler.ListTargetJobs` → `Service.ListTargetJobs` → `SQLStore.ListTargetJobsForUser` | `target_jobs` + completion-ledger read projection | none | `E2E.P0.018`, `E2E.P0.098` |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json`: current `default`, `not-started-progress`, `all-completed-progress`, `prototype-baseline` | `startPractice.ts` + `usePracticeTargetDisplay` | `backend/internal/targetjob.Handler.GetTargetJob` → `Service.GetTargetJob` → `SQLStore.GetTargetJobByUser` | `target_jobs` + requirements/progress projection | none | `E2E.P0.045`, `E2E.P0.098` |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json`: current `default`, `ai-timeout-502` | `frontend/src/app/interview-context/startPractice.ts` | `Handler.StartPracticeSession` → `Service.StartPracticeSession` → `SQLRepository.ReserveSessionStart/CommitSessionStart` | `practice_sessions`, opening `practice_messages`, lifecycle/outbox/idempotency/AI task records | `practice.session.chat` | `E2E.P0.023`, `E2E.P0.057` |
| `getPracticeSession` | current `openapi/fixtures/PracticeSessions/getPracticeSession.json`: `default`, `missing-session`, `prototype-baseline`, `reply-pending`, `reply-retryable-failed`, `reply-terminal-failed`, `reply-complete` | `usePracticeSessionLoader.ts` + `PracticeScreen` rehydration/95-second reconciliation | `Handler.GetPracticeSession` → `Service.GetPracticeSession` → `SQLRepository.GetSession`；backend-practice/002 Phase 11 adds expired-lease lazy convergence | `practice_sessions`, public `client_message_id/reply_status`；internal generation/lease remain backend-only | none | `E2E.P0.044`, `E2E.P0.046` |
| `sendPracticeMessage` | current `openapi/fixtures/PracticeSessions/sendPracticeMessage.json`: `default`, `ai-timeout-retryable`, `validation-empty-text`, `auth-unauthorized`, `session-not-found`, `reply-pending-conflict`, `client-message-mismatch`, `retry-success-same-client-message`；transport timeout stays a fetch/Abort test | `usePracticeMessages.ts` + `PracticeScreen` row-local retry/95-second abort | `Handler.SendPracticeMessage` → `Service.SendPracticeMessage` → SQL reserve/fail/commit with backend-only generation fence | public `client_message_id/reply_status`；internal 90-second lease/generation；task-runs | `practice.session.chat` | `E2E.P0.044`, `E2E.P0.046` |
| `completePracticeSession` | `openapi/fixtures/PracticeSessions/completePracticeSession.json`: current `default`, `replay`, `mismatch`, `cross-user-not-found`, `session-already-completed`; **planned** `zero-answer-rejected`, `one-answer-ready` | `useCompletePracticeSession` + Finish CTA | `Handler.CompletePracticeSession` → `Service.CompletePracticeSession` → backend-practice completion store transaction | zero-answer none；success writes session/report/job/outbox/idempotency | report job only after valid completion | `E2E.P0.047` |
| `createPracticeVoiceTurn` | `openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json`: current `default` disabled response | no frontend consumer | `backend/internal/api/practice.Handler.CreatePracticeVoiceTurn` fail-closed | none | none | `E2E.P0.007`, `E2E.P0.045` |

## 6 Conversation 状态

- Loading：conversation skeleton，不展示假 opening message。
- Running：ordered messages + enabled composer。
- Running / zero-answer：composer enabled，Finish native disabled；可见 zh/en reason 与按钮通过 `aria-describedby` 关联。
- Sending：提交后立即清空 composer 并显示 optimistic user row；composer disabled，Transcript 显示面试官思考动画，retry icon 不渲染；POST 最多等待 95 秒，timeout abort 后进入同 ID reconciliation。
- AI failure：思考动画消失，原 optimistic user row 保留；retry icon 只出现在该 row 底部。同一 `clientMessageId` 与原文本重试，不重复 user message；textarea 可保存下一条草稿但 submit disabled，直至失败消息恢复成功。
- Retry pending：复用 Sending 的 composer lock 与思考动画；retry icon 暂时隐藏，成功后 server session 替换 optimistic row，失败后同一 icon 恢复。
- Terminal message failure：不显示 row-local retry；保持 composer/Finish disabled，显示通用恢复说明与唯一“返回当前面试规划”CTA，进入 `parse(targetJobId)`；auth/session-lost 仍走各自全局边界。
- Reloaded pending：`getPracticeSession` 返回原 user message 的 `clientMessageId + replyStatus=pending`；不追加第二条 optimistic row、不再次 send，保持 composer/Finish disabled 与思考动画，并单飞 re-read。服务端 90 秒 lease 到期后，下一次 GET 必须惰性返回 `retryable_failed`；前端不得无限等待永不变化的 pending。
- Reloaded retryable failure：服务端返回 `retryable_failed` 时在原持久 user row 下恢复唯一 retry icon；点击后使用该 row 的原文本与 `clientMessageId`，成功只产生一个 assistant reply。
- Reloaded terminal / complete：`terminal_failed` 无 retry，显示通用安全说明并提供唯一 `parse(targetJobId)` 当前规划 CTA；`complete` 直接显示服务端唯一 pair。以上状态均不从浏览器 storage 恢复。
- Paused：仅当前页面 composer disabled、计时显示暂停，可恢复；刷新后以 server session 状态重新进入 Running。
- Completing/completed：composer disabled，finish CTA guarded。
- Missing/cross-user：session-lost state 返回 workspace。

## 7 Layout

- Desktop：Top Bar 下只有一个 conversation column；内容 max-width 居中，不留 260px sidebar 空白。
- Mobile：单列，Top Bar controls wrap；Transcript 和 Composer 不横向溢出。
- Transcript 独立滚动，Composer 保持在会话区底部。
- disabled phone icon 不得在 narrow layout 变成可点击入口。

## 8 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Workspace | ready plans | 进入 workspace | 当前卡片列表/启动/归档行为保持 | 001 |
| C-2 | Practice 首屏 | session 有 opening message | 进入 practice | 只见 Top Bar + 全宽 Conversation | 002 |
| C-3 | 连续聊天 | session running，或刷新后服务端 user message 为 `replyStatus=pending` | 提交一条消息并等待 AI / 刷新页面 | user message 立即进入 Transcript、composer 立即清空并禁用、面试官思考动画可访问；刷新从 `getPracticeSession` 重建同一 row + thinking，不重复 send；成功后 server messages 按序收敛且无重复/题目分类 | 002 |
| C-4 | 消息失败恢复 | AI/网络首次可重试失败，或 validation/auth/not-found/conflict 终态失败 | 查看失败 row、刷新、编辑下一条草稿并恢复 | generated `ApiClientError.apiError` 或 transport failure 决定当前请求分类；刷新后以 server `clientMessageId + replyStatus` 重建。可重试失败只在原 user row 底部显示 retry，复用原文本/同 ID 且保留草稿；终态错误无 retry icon并转入事实恢复；两类状态均保持 Finish disabled，AI failure → reload → retry 成功后 user message 与 reply 各唯一一条 | 002 |
| C-5 | 暂停/完成 | session running，可能存在加载/发送/完成失败 | pause/resume/finish/retry | 暂停为页面本地状态；retry 调用原失败操作；完成期间 CTA guarded 并进入 generating | 002 |
| C-6 | phone disabled | 任意 route params | 查看/操作 phone icon | disabled，仍为文本 conversation | 002 + voice/001 |
| C-7 | DOM parity | prototype 已更新 | Vitest | 结构/控件/a11y 与 source 一致 | 002 |
| C-8 | Visual parity | desktop/mobile | Playwright | geometry/screenshot 与 source 一致 | 002 |
| C-9 | Stale negative | current tree | lint/search | 无 SessionMap/QuestionCard/hint/PhoneSurface 正向残留 | 002 |
| C-10 | Privacy | conversation 完成 | 检查 URL/storage/log | raw messages 不泄漏 | 002 |
| C-11 | 轮次预算与推进 | TargetJob 有严格递增但可能非连续的结构化轮次，如 `1,2,4` | 启动当前轮或在报告点击进入下一轮 | plan/计时预算与所选轮次时长一致；从 sequence 2 推进到 canonical 列表中的 4，不构造不存在的 3；重复派生 ID、末轮、单轮、空轮次、未知轮次、加载失败和重复点击不创建错误 plan/session | 001 + 002 + frontend-report-dashboard/001 |
| C-12 | 持久化进度与卡片刷新 | 完成一轮后重新进入/刷新 Home、Workspace 或 Parse，可能有非连续/相邻等时长轮次、legacy plan、全部完成或旧报告 | API 重新加载 TargetJob 并点击 `立即面试` / `进入下一轮` | rail 显示 backend 已完成前缀与当前 canonical successor；只复用 exact current round plan；legacy null/错轮不复用；全部完成禁用启动；生命周期 status 变化不改变轮次；业务进度未写入任何前端持久化介质；真实浏览器刷新与 quick-start 只有在 live frontend/backend 实际执行后才可作为完成证据 | 001 + frontend-report-dashboard/001 |
| C-13 | 零回答完成门禁 | session 只有 opening assistant message，或已有一条 committed user message | 查看 Finish / 绕过 UI 调 completion | 零回答 Finish disabled 且有本地化可访问原因；直接 API 仍由 backend `VALIDATION_FAILED` 拒绝且零副作用；一条回答满足资格后可正常完成 | 002 + backend-practice/002 Phase 9 |
| C-14 | 95 秒 timeout 对账 | POST 在服务端已 reserve 后无响应，或 abort 后旧 response 迟到 | 等待 95 秒并执行同 ID `getPracticeSession` | fetch 被 abort；服务端 pending/failed/complete 被采用；读失败/未找到时原 row 与 ID 保留且新 ID/Finish 仍锁定；迟到旧 response 不覆盖较新事实 | 002 + backend-practice/002 Phase 11 |
| C-15 | terminal 当前规划恢复 | server row 为 `terminal_failed` 且 session 有 authoritative `targetJobId` | 查看终态并点击恢复 CTA | 无 retry icon；显示通用安全说明；唯一 CTA 精确进入 `parse(targetJobId)` 当前面试规划，不进入无上下文 workspace | 002 |

## 9 关联计划

- [001-workspace-and-interview-context](./plans/001-workspace-and-interview-context/plan.md)
- [002-practice-text-event-loop](./plans/002-practice-text-event-loop/plan.md)

## 10 关联文档

- [product-scope](../product-scope/spec.md)
- [backend-practice](../backend-practice/spec.md)
- [practice-voice-mvp](../practice-voice-mvp/spec.md)
- [frontend-report-dashboard](../frontend-report-dashboard/spec.md)
- [openapi-v1-contract](../openapi-v1-contract/spec.md)
- [module-practice-review](../../ui-design/module-practice-review.md)

## 11 修订记录

| 版本 | 日期 | 变更 |
|------|------|------|
| 1.42 | 2026-07-14 | Confirm T-B/P-A: pair the 90-second backend lease with a 95-second abort-and-reconcile client timeout, ignore stale responses, and give terminal failures one generic CTA to the exact current `parse(targetJobId)` plan. |
| 1.41 | 2026-07-13 | 用户消息改为即时 optimistic row；pending/retry 显示面试官思考并锁定 composer；失败仅在原消息底部显示同 ID retry icon，成功回归 server messages。 |
| 1.40 | 2026-07-12 | 原地重开 002：零回答 Finish 原生禁用并提供本地化可访问原因；backend completion 保持权威拒绝与零副作用。 |
| 1.39 | 2026-07-12 | 将 GeneratingScreen 唯一 owner 转交 frontend-report-dashboard；本 owner 仅保留 completion 的 stable reportId handoff，避免两个计划并行修改同一屏幕。 |
| 1.38 | 2026-07-12 | 明确 sequence 可严格递增但不连续，下一轮取现有 canonical successor；区分真实 PostgreSQL/单测组合证据与尚需实际执行的 live browser 刷新门禁。 |
| 1.37 | 2026-07-12 | 采用方案 A：卡片/详情/quick-start 消费 backend-persisted `practiceProgress`，plan 以 exact round pair 复用，移除 status/时长/前端存储轮次推断。 |
| 1.36 | 2026-07-12 | 重新打开轮次 handoff owner：结构化轮次成为时间预算与下一轮推进的单一来源，禁止固定 25 分钟、固定轮次表和末轮/未知轮次 fallback。 |
| 1.35 | 2026-07-12 | 重新打开 Practice owner：按 loader/message/completion 错误来源路由 retry，并在发送/加载/完成边界禁用结束 CTA。 |
| 1.34 | 2026-07-12 | Practice 改为全宽连续文本会话；删除题目/hint/mode UI，电话入口置灰，generating 改用会话级文案。 |
