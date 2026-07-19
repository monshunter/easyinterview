# Frontend Workspace and Practice Spec

> **版本**: 1.55
> **状态**: completed
> **更新日期**: 2026-07-20

## 1 背景与目标

`frontend-workspace-and-practice` 承接 `/workspace` 面试规划列表、`/workspace?targetJobId` 只读详情与 `practice` 连续文本会话。当前 Practice 不再使用题目/turn 状态机，正式页面只保留 Top Bar、全宽 Transcript 和 Composer；电话入口暂时置灰。

`generating` 与 `report` 由 `frontend-report-dashboard` / `backend-review` 承接；本 owner 只负责 completion 返回 stable `reportId` 后的 route handoff，不拥有生成态文案、轮询或错误动作。

## 2 范围

### 2.1 Workspace

- `/workspace` 无 `targetJobId` 时渲染 ready TargetJob 规划卡片列表；`/workspace?targetJobId=<uuid>` 时通过 `getTargetJob` 渲染统一只读详情母版。
- 卡片主体直接进入 workspace detail，右上角归档，底部只有 `立即面试`；已解析规划不得再经过 Parse 动画。
- 列表视觉按 1916×821 参考稿收敛：TopBar 下方背景层必须覆盖完整 viewport 宽度和剩余高度，不能受居中内容最大宽度或 overflow 裁剪；desktop 内容层约 1508px，header 与 card grid 共享 1456px 右边界，“新建面试规划”按钮右侧必须与第二列卡片右侧对齐；规划卡使用两列等宽宽卡与 28px 间距，单卡保持同一列宽而不横跨整行。卡内依次为公司、岗位、轮次 rail、分隔线与“上次保存 + 开始模拟面试”footer。mobile 保持同一 DOM 顺序并收敛为单列，不横向溢出。
- Workspace 只允许 `targetJobId` 详情 locator；不接受 `planId`、`resumeId`、auto-start 或业务事实，不拥有 Resume Picker、Plan Switcher 或 route-side start side effect。
- 快速启动通过 generated `createPracticePlan` / `startPracticeSession` 创建或复用 plan，然后进入 `practice`。
- Home 最近面试、Workspace 列表、Workspace 详情与 Report 复练/下一轮的启动请求必须共享同一个全屏面试准备过渡态。过渡态从用户触发有效启动后立即出现，覆盖并阻断原页面交互，直到成功导航 `practice` 或失败回到原入口错误；不得只禁用按钮、保留看似静止的页面，也不得伪造百分比、阶段完成或 opening message。
- 卡片 round rail、`立即面试` 和 parse 当前轮只消费 backend `TargetJob.practiceProgress`；TargetJob lifecycle `status` 不参与轮次推断，也不在面试规划卡片展示。卡片只在 `locationText` 非空时展示真实地点，不渲染空地点占位。
- Workspace 详情的轮次假设卡片必须复用列表 rail 的同一严格投影：完成前缀显示 `done / 已进行`，首个未完成轮显示 `current / 即将进行`，其余显示 `pending / 未进行`；三态分别使用 success-soft、accent-soft、neutral-soft 背景与对应边框，并暴露 `data-round-state`。投影缺失或无效时保持中性、隐藏状态文案并禁用启动，不得伪造 pending/current/done。
- Workspace 详情删除独立 Interview Launch / 绑定简历大卡片与页尾启动区。标题 cluster 在“面试规划详情”旁显示“绑定简历”查看链接，点击只使用 `getTargetJob` 返回的 `resumeId` 进入 `resume_versions?resumeId=...`，不调用 `getResume` 预读、不提供 rebind；缺失绑定时显示非链接状态并禁用启动。
- 标题下方首行动作行从左依次展示「立即面试」与「面试报告」；desktop 同排，mobile 保持顺序并在必要时换行。报告入口仍只携带可信 `targetJobId`，启动仍使用后端绑定 resume/round/progress 事实；两者不得回到标题右侧或页尾。

### 2.2 Practice

- Practice 在全局 76px App TopBar 下使用 `calc(100dvh - 76px)` 可用高度；desktop 外层以浅蓝背景承接约 `1708px` 会话内容面，Session Header 与 Conversation 卡共享左右边界，不得再以自身 `100vh` 造成页面纵向错位。
- Session Header 按参考图分为状态/岗位、轮次与预算、计时、暂停/电话/结束动作五组；主结束动作保持蓝色高强调，其他控件使用白色描边按钮与统一图标，不使用字符 glyph 充当正式 icon。
- Transcript 使用白色大圆角卡、48px 角色标记、独立消息 surface 与 16px 正文字号；Composer 固定在会话卡底部，textarea 与蓝色发送按钮共同位于同一个内层 input surface，按钮位于该表面的底部 action area 并右对齐。action area 与 textarea 不重叠，使正文保持完整可用宽度；按钮不得悬浮在 input surface 与 Composer 外边框之间，也不得移出 Composer 成为脱离输入语境的动作。说明胶囊是 Composer 的固定附属元素，位于输入框正上方且不属于 Transcript 的滚动内容；聊天记录增长、滚动或重渲染不得改变它与输入框的相对间距。加载、pending、retry、terminal 与 disabled 语义不变。

- Route 只需要稳定 `sessionId` 与 target/plan/resume/round IDs；不使用 `mode/modality/practiceMode/hintUsed/hintCount`。
- Top Bar：真实公司/岗位、面试官角色、计时、暂停、disabled phone icon、结束并生成报告。
- `TargetJob.summary.interviewRounds[]` 只定义 canonical 轮次目录、顺序和时长；当前轮由 backend `TargetJob.practiceProgress.currentRound` 选中。启动时把该轮 `durationMinutes` 写入 `PracticePlan.timeBudgetMinutes`，Practice Top Bar 再从当前 plan 读取并显示预算，不使用固定分钟数。
- Conversation：全宽有序 Transcript + Error/Retry + Composer。
- opening message 和后续 assistant reply 统一来自 server messages，不是 QuestionCard。
- 所有持久化 user/assistant `message.text` 使用 `react-markdown` + `remark-gfm` 作为只读 view projection；`skipHtml` 启用且不得接入 `rehypeRaw`。Raw HTML、remote image、`javascript:`/unsafe URI 不得执行或发起请求；link 只允许安全协议并带安全 `rel`。原始 `message.text` 仍是发送、持久化和 same-ID retry 的唯一 payload。
- 用户输入通过 generated `sendPracticeMessage`，不提交 `turnId`，不标记 answer/hint/question。
- 提交后必须立即清空 composer，并先把该条 user message 作为当前页面的瞬时 optimistic row 加入 Transcript；不得等待 assistant response 后才与 reply 一起出现。该 row 只用于请求中的即时反馈，不写浏览器存储、不计入 Finish 资格，也不覆盖 server messages 事实源。
- `getPracticeSession` 刷新恢复完整 ordered messages，并为每条 user message 返回服务端事实 `clientMessageId` 与 `replyStatus=pending|retryable_failed|terminal_failed|complete`；assistant message 不伪造这些 user-only 字段。`pending` 继续显示思考并自动 re-read server truth，`retryable_failed` 恢复 row-local retry，`terminal_failed` 进入无 retry 的事实恢复，`complete` 只展示唯一 user/assistant pair。
- 服务端 pending reservation 的 lease 固定为 90 秒；前端 `sendPracticeMessage` 的单次 POST 等待上限固定为 95 秒。95 秒到达时必须 abort 本次 fetch，并立即用 `getPracticeSession` 对账同一 `clientMessageId`；不得盲目自动重发、生成新 ID 或把客户端时钟当成 reply status 事实源。
- 暂停/恢复只控制当前页面的 composer 与计时显示，不产生 backend 事件；结束通过 `completePracticeSession`。
- Error/Retry 必须按失败来源恢复：session loader 调用 `refresh`，message failure 使用同一 `clientMessageId` 重试 send，completion failure 使用同一 completion idempotency key 重试 finish；不得把完成重试误接到 send。
- message pending/retrying 时 composer disabled，并在 Transcript 中追加 assistant-style、`aria-live` 的面试官思考动画；成功后用 server session 原子替换 optimistic row/思考态，不能重复 user message。失败后隐藏思考态、保留原 user row，只在该 row 底部渲染 retry icon；该 icon 必须调用与 composer submit 相同的 send path，并复用原 `clientMessageId` 与原文本。
- row-local retry 只属于明确可重试的 message failure：无 HTTP response 的网络错误，或 generated `ApiClientError.apiError.retryable=true`（含 AI timeout/5xx）。OpenAPI owner 必须从 error envelope 生成 typed `ApiClientError`，保留 HTTP status、`code/requestId/retryable/details` 与 transport cause；Practice 不得解析普通 `Error.message` 字符串。intentional abort/unmount 不渲染 retry。`VALIDATION_FAILED`、auth/not-found、client-message conflict/mismatch 等终态不得渲染同 ID retry icon；它们必须走 loader/auth/session-lost 等事实恢复，重新读取 server messages 后再决定 composer 是否可用，不能让用户陷入无限重试。
- `terminal_failed` 是服务端权威终态，不显示 row-local retry，也不允许永久空白锁死。页面必须显示一条不泄漏技术细节的通用恢复说明和唯一主动作“返回当前面试规划”，精确执行 `navigate({ name: "workspace", params: { targetJobId: session.targetJobId } })` 并进入 `/workspace?targetJobId=...` 只读详情；不得退回无上下文 workspace 列表、不得再进入 `parse(targetJobId)` 命令进度页、不得携带 `planId`、不得复用 composer submit。
- 当前页面内存中的 `{text, clientMessageId, status}` 只覆盖 submit 到首次 response/read convergence 的即时反馈；一旦刷新或重挂载，恢复必须完全来自 `getPracticeSession` 的 `clientMessageId + replyStatus`，不得用 URL、localStorage、sessionStorage 或 IndexedDB 保存 retry identity。AI 失败后刷新仍必须用服务端返回的原文本与同一 `clientMessageId` 重试，最终收敛为唯一 user/reply pair。
- message failure 后 textarea 可恢复输入以保留下一条草稿，但在失败消息完成同 ID retry 前 submit 仍 disabled，并提供本地化说明；草稿不得改变 retry payload，也不得作为另一条业务消息绕过待回复状态。
- Composer 按 RuntimeConfig `practiceMessageBytes` 默认 32KiB 预检，当前 server-loaded messages 与 draft 按 `practiceSessionTextBytes` 默认 256KiB 预检；统一使用 UTF-8 bytes 而不是字符/rune 计数。limit+1 不调用 send，显示本地化可恢复说明；backend persisted aggregate 仍是最终权威，前端不得把 optimistic/route/storage 作为会话总量事实。
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
| D-7 | 业务状态后端持久化 | 主题/外观偏好由 frontend-shell 账号设置 owner 持久化；本 owner 的轮次进度、当前轮、plan/session/report 和完成事实只来自 backend API。`TargetJob.practiceProgress` 是卡片/详情/quick-start 的 read model；缺失或不一致时 fail closed。 |
| D-16 | Practice 全局 chrome | `practice` 保留全局 App TopBar，并在其下渲染独立 Practice Session Header | 会话页与其他页面拥有一致导航/显示入口；route 切换不得触发 `/me`，会话控制栏不冒充 App chrome |
| D-8 | Finish 最低回答门槛 | 前端只从 server-loaded messages 计算至少一条 committed candidate `user` message；零回答原生 disabled 并显示本地化可访问原因。Backend `completePracticeSession` 独立执行同一事实校验并保持最终权威。 |
| D-9 | 即时消息与失败恢复（方案 A） | user submit 后立即显示瞬时 optimistic row；服务端 `PracticeMessage` 为 user message 投影 `clientMessageId + replyStatus(pending|retryable_failed|terminal_failed|complete)`，`getPracticeSession` 在刷新后恢复 thinking/retry/terminal/complete；OpenAPI owner 生成 typed `ApiClientError.apiError`；retry 复用服务端原文与同 ID。该方案兼顾即时反馈、跨刷新幂等恢复与后端事实源；前端不持久化 retry identity、不解析 error string、不引入第二套消息事实或无限重试。 |
| D-10 | Pending 超时与对账（T-B） | backend lease 为 90 秒；frontend POST timeout 为 95 秒并 abort；随后只用同一 ID `getPracticeSession` 对账；GET/同 ID reserve 负责服务端惰性收敛。这样既覆盖服务端 lease 又不无限挂起；超时不是失败事实，新 ID 与盲目自动重发均禁止。 |
| D-11 | Terminal 恢复入口（P-A） | `terminal_failed` 展示通用安全说明 + 唯一“返回当前面试规划”CTA，精确进入 `/workspace?targetJobId`；无 row retry。 |
| D-12 | Workspace list/detail split | 无 targetJobId 是列表；有 targetJobId 是只读详情，card 直达；详情只执行一次同 key `getTargetJob`，不 import、不 poll、不播放 Parse animation | shell/004 owns ready replace/back；shell/001 owns safe-read single-flight |
| D-13 | Conversation Markdown projection | user/assistant 均由 `react-markdown + remark-gfm` 渲染；`skipHtml`、no `rehypeRaw`、no remote image、safe link；retry 保留原始 text/clientMessageId | Markdown 只是 view，不能改写业务 payload；mobile code block 必须容器内滚动 |
| D-14 | Workspace detail round-state affordance | 详情卡片与列表 mini rail 消费同一 `practiceProgress`：`done/已进行`、`current/即将进行`、`pending/未进行` 使用三种背景、边框、可见标签与 `data-round-state`；无效投影中性 fail closed | 不新增 API/schema/前端状态机，不从 TargetJob lifecycle、URL 或 storage 猜测 |
| D-15 | Practice text limits | `AppRuntimeProvider.contentLimits.practiceMessageBytes/practiceSessionTextBytes` 是唯一前端数据源，缺字段用 A4 同值 code default 32768/262144；`TextEncoder` 计算 bytes；backend error 可覆盖前端估算 | 删除 8,000-rune 本地真理源，保持 composer DOM/视觉不变并防止正常长回答误拒 |
| D-16 | Workspace detail leading controls | 删除独立 Interview Launch/绑定简历 block；标题旁的“绑定简历”只按 `TargetJob.resumeId` 打开对应简历详情，标题下首行动作行左对齐“立即面试 + 面试报告”。缺绑定时 link 不可用且 Start fail closed，Report 仍按可信 target 可用 | 减少重复上下文块，把查看绑定、开始面试和查看报告前置到详情开头，同时保持 backend 事实源与 route 最小化 |
| D-17 | Workspace list reference geometry | desktop 使用两列宽卡、宽松页面标题区和 52px 级删除触控区；卡片 footer 左侧显示 API `updatedAt` 派生的本地化“上次保存”，右侧保留唯一开始 CTA；mobile 单列 | 只改变正式前端信息层级和几何，不修改 route、API、归档、轮次或启动事实源 |

## 4 UI 设计文档与 parity

- Workspace：`frontend/src`
- Practice：`frontend/src`
- Generating：由 `frontend-report-dashboard` 独占；本 spec 仅引用 completion handoff，不修改其原型或正式屏幕。
- Shared：`frontend/src`
- Docs：`docs/ui-design/module-job-workspace.md`、`module-practice-review.md`、`report-dashboard.md`

用户可见改动必须先更新 `frontend/`，再源级迁移到 frontend。验证必须拆分：

1. DOM/control/a11y formal implementation contract。
2. computed style/bounding box/responsive/screenshot geometry parity。
3. stale question/hint/phone positive-contract negative search。

## 5 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticePlan` | `openapi/fixtures/PracticePlans/createPracticePlan.json`: current variants | `frontend/src/app/interview-context/startPractice.ts` | practice handler/service/store | `practice_plans`, idempotency/audit | none | 当前无真实 E2E owner；root `make test` |
| `getPracticePlan` | `openapi/fixtures/PracticePlans/getPracticePlan.json`: current variants | exact-plan reuse + Practice budget read | practice handler/service/store | `practice_plans` read | none | 当前无真实 E2E owner；root `make test` |
| `listTargetJobs` | current variants | Workspace list rail | targetjob list handler/service/store | TargetJob + completion-ledger projection | none | `E2E.P0.098` 仅 progress refresh |
| `getTargetJob` | current variants | Workspace detail/start/display | targetjob get handler/service/store | TargetJob requirements/progress | none | `E2E.P0.098` 仅 progress/detail read |
| `startPracticeSession` | current fixtures | shared start | practice handler/service/store | sessions/messages/idempotency/AI task | `practice.session.chat` | 当前无真实 E2E owner；root `make test` |
| `getPracticeSession` | current fixtures | loader/rehydration | practice owner | messages/reply facts | none | 当前无真实 E2E owner；root `make test` |
| `sendPracticeMessage` | current fixtures | send/retry | practice owner | messages/idempotency/reply facts | `practice.session.chat` | 当前无真实 E2E owner；root `make test` |
| `completePracticeSession` | current fixtures | Finish CTA | practice completion owner | session/report/job/outbox/idempotency | report job after valid completion | `E2E.P0.098` 仅真实 completion API 与进度刷新 |
| `createPracticeVoiceTurn` | disabled fixture | no frontend consumer | fail-closed handler | none | none | 当前无真实 E2E owner；root `make test` |

## 6 Conversation 状态

- Loading：conversation skeleton，不展示假 opening message。
- Pre-session launch：`createPracticePlan/getPracticePlan/startPracticeSession` 尚未完成时显示本地化 indeterminate 面试准备过渡态；使用 `role=status`、`aria-live=polite`、`aria-busy=true`，支持 `prefers-reduced-motion`，不写 URL/storage，不改变 API/idempotency。未登录认证跳转不提前展示；失败关闭覆盖层并保留原入口可恢复错误。
- Running：ordered messages + enabled composer。
- Running / zero-answer：composer enabled，Finish native disabled；可见 zh/en reason 与按钮通过 `aria-describedby` 关联。
- Sending：提交后立即清空 composer 并显示 optimistic user row；composer disabled，Transcript 显示面试官思考动画，retry icon 不渲染；POST 最多等待 95 秒，timeout abort 后进入同 ID reconciliation。
- AI failure：思考动画消失，原 optimistic user row 保留；retry icon 只出现在该 row 底部。同一 `clientMessageId` 与原文本重试，不重复 user message；textarea 可保存下一条草稿但 submit disabled，直至失败消息恢复成功。
- Retry pending：复用 Sending 的 composer lock 与思考动画；retry icon 暂时隐藏，成功后 server session 替换 optimistic row，失败后同一 icon 恢复。
- Terminal message failure：不显示 row-local retry；保持 composer/Finish disabled，显示通用恢复说明与唯一“返回当前面试规划”CTA，进入 `/workspace?targetJobId` 只读详情；auth/session-lost 仍走各自全局边界。
- Reloaded pending：`getPracticeSession` 返回原 user message 的 `clientMessageId + replyStatus=pending`；不追加第二条 optimistic row、不再次 send，保持 composer/Finish disabled 与思考动画，并单飞 re-read。服务端 90 秒 lease 到期后，下一次 GET 必须惰性返回 `retryable_failed`；前端不得无限等待永不变化的 pending。
- Reloaded retryable failure：服务端返回 `retryable_failed` 时在原持久 user row 下恢复唯一 retry icon；点击后使用该 row 的原文本与 `clientMessageId`，成功只产生一个 assistant reply。
- Reloaded terminal / complete：`terminal_failed` 无 retry，显示通用安全说明并提供唯一 `/workspace?targetJobId` 当前规划详情 CTA；`complete` 直接显示服务端唯一 pair。以上状态均不从浏览器 storage 恢复，当前范围不得保留 `parse(targetJobId)` 正向恢复路径。
- Paused：仅当前页面 composer disabled、计时显示暂停，可恢复；刷新后以 server session 状态重新进入 Running。
- Completing/completed：composer disabled，finish CTA guarded。
- Missing/cross-user：session-lost state 返回 workspace。

## 7 Layout

- Workspace list desktop：页面背景层全宽且至少覆盖 TopBar 下方剩余 viewport；居中内容层最大宽度 1508px，header 和卡片网格最大宽度 1456px，标题组与新建 CTA 分列且 CTA 右侧与第二列卡片右侧对齐；规划卡两列等宽，卡片最小高度约 384px，round rail 在正文中部，footer 贴近卡片底部。单卡不得扩展为两列宽度。
- Workspace list mobile：标题组、创建 CTA、规划卡和 footer actions 顺序堆叠；按钮与长公司/岗位/轮次名保持可见且 document 无横向溢出。
- Desktop：Top Bar 下只有一个 conversation column；内容 max-width 居中，不留 260px sidebar 空白。
- Mobile：单列，Top Bar controls wrap；Transcript 和 Composer 不横向溢出。
- Transcript 独立滚动，Composer 保持在会话区底部；说明胶囊与输入框共同位于 Composer 固定区，短/长聊天记录下都保持相同垂直间距。
- Desktop/mobile 均由一个内层 input surface 同时包裹 textarea 与 send；send 在不覆盖 textarea 的底部 action area 右对齐，任意长度文本、placeholder 与光标保持完整可用宽度。
- disabled phone icon 不得在 narrow layout 变成可点击入口。

## 8 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Workspace list/detail | ready plans | 进入 `/workspace` 或 `/workspace?targetJobId` | 无参列表；有 target 只读详情；card 直达；详情 same-key `getTargetJob` 底层 count=1，零 import/poll/Parse animation；标题旁绑定简历只按 saved `resumeId` 打开详情，无独立 binding/launch block；标题下首行动作行左对齐“立即面试 + 面试报告” | 001 |
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
| C-12 | 持久化进度与卡片刷新 | 完成一轮后重新进入/刷新 Home、Workspace 或 TargetJob detail，可能有非连续/相邻等时长轮次、legacy plan、全部完成或旧报告 | API 重新加载 TargetJob | rail 显示 backend 已完成前缀与当前 canonical successor；只复用 exact current round plan；legacy null/错轮不复用；全部完成禁用启动；生命周期 status 变化不改变轮次；业务进度未写入任何前端持久化介质。`E2E.P0.098` 只证明真实登录、completion API、Home/Workspace/TargetJob 进度刷新与 TargetJob detail read；`createPracticePlan`、quick-start、`startPracticeSession`、进入下一轮和会话启动当前无真实 E2E owner，由代码层测试随根 `make test` 回归 | 001 + frontend-report-dashboard/001 |
| C-13 | 零回答完成门禁 | session 只有 opening assistant message，或已有一条 committed user message | 查看 Finish / 绕过 UI 调 completion | 零回答 Finish disabled 且有本地化可访问原因；直接 API 仍由 backend `VALIDATION_FAILED` 拒绝且零副作用；一条回答满足资格后可正常完成 | 002 + backend-practice/002 Phase 9 |
| C-14 | 95 秒 timeout 对账 | POST 在服务端已 reserve 后无响应，或 abort 后旧 response 迟到 | 等待 95 秒并执行同 ID `getPracticeSession` | fetch 被 abort；服务端 pending/failed/complete 被采用；读失败/未找到时原 row 与 ID 保留且新 ID/Finish 仍锁定；迟到旧 response 不覆盖较新事实 | 002 + backend-practice/002 Phase 11 |
| C-15 | terminal 当前规划恢复 | server row 为 `terminal_failed` 且 session 有 authoritative `targetJobId` | 查看终态并点击恢复 CTA | 无 retry icon；唯一 CTA 精确进入 `/workspace?targetJobId` 当前规划详情 | 002 |
| C-16 | Safe Markdown/GFM | persisted user/assistant text 含 GFM 与恶意 HTML/image/link/code | 渲染、retry、mobile 查看 | 两类角色都渲染 GFM；HTML/remote image/unsafe URI 不执行；safe link hardened；retry exact raw text/ID；code 不撑破 viewport | 002 |
| C-17 | Composer 说明定位 | session 中存在短或长聊天记录 | 聊天增长并滚动 Transcript | 说明胶囊始终作为 Composer 子元素贴在输入框上方；不随 Transcript 内容移动，二者间距在 desktop/mobile 保持稳定 | 002 |
| C-17 | Workspace 详情轮次三态 | ready TargetJob 有 2~5 条 canonical rounds 与合法/完成/无效 `practiceProgress` | 打开或刷新 `/workspace?targetJobId` | 合法进行中显示完成前缀 `done/已进行`、唯一 `current/即将进行`、其余 `pending/未进行`，三态背景/边框不同且与列表 rail 一致；全完成全部 done；无效投影中性且启动 disabled | 001 |
| C-18 | Practice byte boundaries | owner config provides message/session UTF-8 limits | submit / reload | 注入小型 boundary 验证 overflow zero send and draft recovery；backend remains authoritative；默认/override/invalid 由 typed config owner 覆盖，不构造默认大小文本或配置 E2E | 002 Phase 12 |
| C-19 | 面试规划卡片元信息 | ready TargetJob 的 lifecycle status 为任意值，地点可能有值或缺失 | 查看 Home 最近面试或 Workspace 规划卡片 | 卡片不展示 lifecycle status 文案/徽标；有地点时展示真实值，缺失或空白时不渲染地点占位行；轮次 rail 仍表达真实训练进度 | 001 |
| C-20 | 会话启动等待反馈 | 用户从 Home、Workspace 列表/详情或 Report 复练/下一轮发起有效面试，session opening LLM 请求持续未返回或失败 | 点击启动并等待 | 立即展示统一全屏、可访问、阻断交互且 reduced-motion 兼容的诚实过渡态；不伪造进度/opening；成功进入 `practice`，失败关闭过渡态并在原入口显示错误；API、route、idempotency 与持久化合同不变 | 001 |
| C-21 | 面试列表参考稿还原 | desktop/mobile 打开 query-free Workspace，存在 1~N 个 ready TargetJob | 查看标题区、规划卡、轮次 rail 与 footer actions | 背景层覆盖 TopBar 下方完整 viewport，不在内容区右侧形成空白带；desktop 以参考稿的 1508px 内容区和双列宽卡呈现；mobile 单列；公司/岗位/进度/上次保存/删除/启动层级一致，卡片打开、归档和启动行为不回退，控制台无错误且无横向溢出 | 001 Phase 32 |
| C-22 | Composer 发送动作归属 | desktop/mobile 的 Practice Composer 可输入短文本或长文本 | 聚焦 textarea、输入并查看/触发发送按钮 | textarea 与 send 同属一个内层 input surface；send 在表面内部的底部 action area 右对齐，不位于内外边框之间且不覆盖 textarea；文本保持完整可用宽度，发送、快捷键、disabled 语义不变 | 002 Phase 15 |

### 8.1 Practice 启动过渡构图

- 从任一合法入口创建/恢复会话后，在 opening request pending 期间使用共享 `AsyncTransitionScene` 的 `brand` variant：全局 TopBar 仍可见，抽象蓝白画布、同心轨道与中心 E 标识对应参考稿。
- transition 继续承担既有 portal、背景 inert、focus/scroll lock 与成功/失败恢复，不展示无 producer 的百分比或 opening 内容；reduced-motion 下停止非必要轨道动画。

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
| 1.55 | 2026-07-20 | Reopen the Practice owner so send belongs to a non-overlapping bottom action area inside the input surface, preserving full-width text on narrow screens. |
| 1.54 | 2026-07-19 | Reopen the Practice owner for the screenshot-aligned brand transition while preserving blocking, focus and honest opening-request semantics. |
| 1.51 | 2026-07-19 | Require a full-viewport Workspace canvas and align the header CTA right edge with the two-column card grid. |
| 1.50 | 2026-07-19 | 按提供的面试列表参考稿重开 Workspace list 视觉 owner：桌面双列宽卡、参考级标题与动作层级、上次保存 footer，并保留现有 route/API/启动/归档合同。 |
| 1.49 | 2026-07-19 | Practice 恢复全局 App TopBar，并把会话控制栏明确为独立 Practice Session Header；route 切换零账号重复读取。 |
| 1.48 | 2026-07-18 | Add a shared accessible pre-session launch transition across Home, Workspace detail/list, and Report replay/next-round entry points while preserving the existing start-session contract. |
| 1.47 | 2026-07-17 | Remove lifecycle status and empty-location placeholders from shared Home/Workspace interview-plan cards while retaining real location values and persisted round progress. |
| 1.45 | 2026-07-14 | Add RuntimeConfig-backed 32KiB message and 256KiB persisted-session UTF-8 byte limits, replacing the 8,000-rune frontend/backend drift. |
| 1.44 | 2026-07-14 | Add persisted done/current/pending round-state cards to Workspace detail with rail-consistent visuals and invalid-projection fail-closed behavior. |
| 1.43 | 2026-07-14 | Add Workspace query-addressed detail, supersede terminal recovery to that detail, and add safe Markdown/GFM conversation projection with exact raw retry payloads. |
| 1.42 | 2026-07-14 | Confirm T-B/P-A: pair the 90-second backend lease with a 95-second abort-and-reconcile client timeout, ignore stale responses, and give terminal failures one generic CTA to the exact current `parse(targetJobId)` plan. |
| 1.41 | 2026-07-13 | 用户消息改为即时 optimistic row；pending/retry 显示面试官思考并锁定 composer；失败仅在原消息底部显示同 ID retry icon，成功回归 server messages。 |
| 1.40 | 2026-07-12 | 原地重开 002：零回答 Finish 原生禁用并提供本地化可访问原因；backend completion 保持权威拒绝与零副作用。 |
| 1.39 | 2026-07-12 | 将 GeneratingScreen 唯一 owner 转交 frontend-report-dashboard；本 owner 仅保留 completion 的 stable reportId handoff，避免两个计划并行修改同一屏幕。 |
| 1.38 | 2026-07-12 | 明确 sequence 可严格递增但不连续，下一轮取现有 canonical successor；区分真实 PostgreSQL/单测组合证据与尚需实际执行的 live browser 刷新门禁。 |
| 1.37 | 2026-07-12 | 采用方案 A：卡片/详情/quick-start 消费 backend-persisted `practiceProgress`，plan 以 exact round pair 复用，移除 status/时长/前端存储轮次推断。 |
| 1.36 | 2026-07-12 | 重新打开轮次 handoff owner：结构化轮次成为时间预算与下一轮推进的单一来源，禁止固定 25 分钟、固定轮次表和末轮/未知轮次 fallback。 |
| 1.35 | 2026-07-12 | 重新打开 Practice owner：按 loader/message/completion 错误来源路由 retry，并在发送/加载/完成边界禁用结束 CTA。 |
| 1.34 | 2026-07-12 | Practice 改为全宽连续文本会话；删除题目/hint/mode UI，电话入口置灰，generating 改用会话级文案。 |
