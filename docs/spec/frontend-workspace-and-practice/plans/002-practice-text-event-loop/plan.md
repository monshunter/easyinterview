# 002 Practice Text Event Loop

> **版本**: 1.7
> **状态**: active
> **更新日期**: 2026-07-06

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把 [frontend-workspace-and-practice spec](../../spec.md) v1.6 §2.1 / §6 C-4 / §7 锁定的第二个 plan 范围落地，承接 [001-workspace-and-interview-context](../001-workspace-and-interview-context/plan.md) 已交付的 `InterviewContext` + `workspace` 双步启动契约，闭合 P0 用户路径中“文本面试事件循环 + 异步完成 + generating 入口”段：

- `practice` 路由从 `PlaceholderScreen` 切换为正式 `PracticeScreen`，源级复刻 `ui-design/src/screen-practice.jsx::PracticeScreen` 文本面试 surface（顶部工具区 + 题目地图 + 当前问题 + 对话记录 + 输入区 + 右侧 JD 关联/可调用经历/AI 透明度 + 固定底部「结束并生成报告」CTA）。
- 通过 generated client + fixture-backed transport 消费 `getPracticeSession` / `appendSessionEvent` / `completePracticeSession`；严格遵守 spec D-12（appendSessionEvent 用 `clientEventId`、不带 `Idempotency-Key`）与 D-13（completePracticeSession 异步 202 + `ReportWithJob`）。
- 实现 `appendSessionEvent` 5 种 `kind` 路由：`answer_submitted` / `hint_requested` / `turn_skipped` / `session_paused` / `session_resumed`；消费 `SessionEventResult.assistantAction` 5 种 type 渲染下一题 / 追问 / 提示 / 等待 / 完成。
- 消费 `PracticeSession.status` 七值状态机（`queued / running / waiting_user_input / completing / completed / failed / cancelled`），以 `shared/conventions.yaml` / `openapi/openapi.yaml` 当前 `SessionStatus` 为准；前端不重写状态机，只渲染分支。
- 落地 spec D-3 二轴显隐：`practiceMode='strict'` 隐藏提示按钮、左侧实时观察、可调用经历；`practiceGoal∈{baseline,retry_current_round,next_round}` 仅影响题目来源、不影响显隐；模式切换 segmented control 在本 plan 内保留入口。`practice-voice-mvp/001` 后续已接管 voice surface / `createPracticeVoiceTurn`，本 text-loop plan 的当前 gate 只要求 text event loop 不直接轮询 report、不把 voice turn 调用散落到非 voice owner hook。
- 完成动作触发 `completePracticeSession(Idempotency-Key)` → 202 `ReportWithJob{reportId, job}` → nav `generating?...InterviewContext&...PracticeDisplayContext`；`PracticeDisplayContext = {mode, modality, practiceMode, practiceGoal, hintUsed, hintCount}` 仅作为 route handoff，不塞进 backend request body（D-13）。稳定 ID（`planId / targetJobId / jdId / resumeVersionId / roundId / sessionId / reportId`）允许留在 owner route context；隐私红线只禁止 raw answer / question / hint / prompt / provenance 明文泄漏。

完成后用户从 workspace 点击「立即面试」可以进入完整的文本模式模拟面试，按照「答题 → 追问 / 提示 / 跳过 / 暂停 → 完成 → generating」端到端走通。当前仓库中 `backend-practice/002-event-loop-and-completion` 已完成 `getPracticeSession` / `appendSessionEvent` / `completePracticeSession` 真实 handler、service、store 与 `E2E.P0.038-043` Go HTTP scenario；本 plan 仍优先用 fixture-backed transport 做前端 TDD / UI parity，但 Phase 5 必须同时跑真实 backend-practice 002 regression，防止把 mock green 误当真实闭环。`hint_requested → show_hint` 的 assisted 正向路径仍是 backend-practice/003 前置能力，本 plan 只能以 fixture-only UI 合同开发；真实 backend 002 默认返回 `PRACTICE_SESSION_CONFLICT`，前端必须覆盖该防御分支。

## 2 背景

[backend-practice spec](../../../backend-practice/spec.md) v1.7 + plan [002-event-loop-and-completion](../../../backend-practice/plans/002-event-loop-and-completion/plan.md) 已完成 OpenAPI `PracticeSessionEventRequest` / `AssistantAction` / `SessionEventResult` / `ReportWithJob` schema 对齐、真实 handler wiring 与 `E2E.P0.038-043` Go HTTP scenario。当前 `openapi/fixtures/PracticeSessions/` 已有 `getPracticeSession` 的 `default / prototype-baseline / missing-session`，`appendSessionEvent` 的 `default / follow-up / hint-strict-conflict / turn-skipped / pause-resume / replay / mismatch / completed`，以及 `completePracticeSession` 的 `default / replay / mismatch / session-already-completed / cross-user-not-found`；本 plan 只新增仍缺的前端 UI fixture（如 `getPracticeSession.running-with-history / queued / completing`、append `show-hint` fixture-only、必要时 `ai-timeout`），并在 operation matrix 中区分 fixture-only 与真实 backend 已落地能力。`shared/conventions.yaml` 锁定 `PracticeMode∈{assisted,strict}`、`PracticeGoal∈{baseline,retry_current_round,next_round}`、`SessionStatus∈{queued,running,waiting_user_input,completing,completed,failed,cancelled}` 与 `assistantAction.type` 5 值。前端 generated client `frontend/src/api/generated/{client,types}.ts` 已暴露 6 个 practice operation 方法，本 plan 不修改 generated artifacts，仅消费 + 在 owner 边界内创建 hook / component / view-model。

UI 真理源：[`ui-design/src/screen-practice.jsx`](../../../../../ui-design/src/screen-practice.jsx)（`PracticeScreen` 主组件 + `TranscriptMsg` / `RoleDropdown` / `ExpCard` / 输入区分支，含 text input 中的 speech-to-text failure banner lines 583-590）与 [`docs/ui-design/module-practice-review.md`](../../../../ui-design/module-practice-review.md) §3-§6（生命周期、布局、辅助度规则、结束动作显隐）。本 plan 交付时只覆盖 text event loop；`practice-voice-mvp/001` 后续已删除 `VoiceSurfaceComingSoon` 占位并接管 voice surface / STT / TTS / barge-in。当前 P0.044-P0.047 回归 gate 必须允许 voice owner hook 存在，同时继续禁止 text event loop 直接消费 report polling 或绕过 generated client。

001 plan 已经为本 plan 备齐：

- `InterviewContext` reducer 支持 `MERGE_SESSION` action，可承接 `startPracticeSession` 返回 + 本 plan 内 `appendSessionEvent` 响应。
- workspace 启动时已经把 `sessionId / planId / targetJobId / jdId / resumeVersionId / roundId / mode='text' / modality='text' / practiceMode='strict' / practiceGoal='baseline' / hintUsed='false' / hintCount='0'` 通过 route 传到 practice，本 plan 只消费、不重新协商参数。
- `frontend/src/lib/conventions/idempotency.ts::newIdempotencyBatch()` 已提供稳定 `Idempotency-Key` 生成；本 plan 在 `completePracticeSession` 端复用。
- `frontend/src/lib/ids.ts` 已提供 UUIDv7 helper，本 plan 用它派生 `clientEventId`。

### 2.1 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-06 | 1.7 | 对齐 product-scope D-22 与当前 `shared/conventions.yaml`：`practiceGoal` 正向集合收敛为 `baseline / retry_current_round / next_round`，旧 `debrief` 只作为 legacy-negative；P0.045 场景目录同步为 `p0-045-practice-text-loop-mode-policy-display`。 |
| 2026-05-23 | 1.5 | L2 real-backend drift follow-up：`practice-voice-mvp/001` 已落地 `createPracticeVoiceTurn` 与 voice surface，P0.044/P0.047 不再要求 practice 模块 repo-wide 0 命中；scenario verify 改为只禁止 `getFeedbackReport` 进入 practice runtime，并限制 `createPracticeVoiceTurn` 只能出现在 voice owner hook。 |
| 2026-05-23 | 1.4 | L2 real-backend gate remediation：P0.044-P0.047 trigger 前置 `frontendOwners.realApiMode.test.ts`，verify 检查 real-mode marker、默认 backend base URL 与测试文件 marker；front-end UI variants 继续 fixture-backed，真实 practice / report / resume generated-client routing 由集中 gate 证明。 |
| 2026-05-14 | 1.3 | L2 `--fix` 收口：补齐前端错误恢复、privacy / mismatch 专用测试、practice Playwright pixel parity gate、scenario trigger 覆盖，并修正 practice runtime 旧主题 token 漂移。 |
| 2026-05-14 | 1.1 | L1 review fix：按当前 repo truth 修正 backend-practice/002 状态为 completed / real handler landed；operation matrix 标明真实 handler、fixture-only assisted hint 前置到 backend-practice/003，并把 Phase 5 backend regression 扩展为 `E2E.P0.022-026` + `E2E.P0.038-043`。 |
| 2026-05-13 | 1.0 | 初始创建：拆分 spec C-4 文本事件循环 + C-6 generating 入口端 + C-8/9/10/12 横切，落实 §7 plan 序列 002。 |

## 3 质量门禁分类

- **Plan 类型**: feature-behavior（用户可感知 UI + API 行为 + 业务流程 + 端到端功能）+ cross-layer contract（消费 B2 / B3 / mock-contract-suite 已落地的 wire 契约）。
- **TDD 策略**: Red-Green-Refactor 入口为 `pnpm --filter @easyinterview/frontend test`（Vitest + @testing-library/react + jsdom）；每个 Phase 在新增组件 / hook / utils 前先写失败测试，覆盖 DOM 锚点、控件类型、props/state、generated client 调用断言（method、path、body schema、header 反查）、URL/state 隐私反查与负向旧 testid / 旧 route alias / 旧 enum 值断言；`pnpm --filter @easyinterview/frontend test:pixel-parity` 在 Phase 5 扩展为 `practice.spec.ts`（desktop 1440×900 + mobile 390×844）。新增组件位于 `frontend/src/app/screens/practice/`；测试文件与组件 colocate（`*.test.tsx` / `*.test.ts`）。`test-plan.md` + `test-checklist.md` 拓展测试形态与 phase 映射。
- **BDD 策略**: Feature plan requires BDD；本 plan 在 [bdd-plan.md](./bdd-plan.md) 定义 4 个场景 `E2E.P0.044 / E2E.P0.045 / E2E.P0.046 / E2E.P0.047`，[bdd-checklist.md](./bdd-checklist.md) 跟踪每个场景资产创建与执行；主 [checklist.md](./checklist.md) 在每个 Phase 末尾保留 `BDD-Gate:` 项引用对应场景 ID。
- **替代验证 gate**: 不适用（feature plan，已有完整 BDD + TDD 双层覆盖 + pixel parity + contract drift）。

## 3.5 Coverage Matrix

| 类别 | 覆盖描述 | UI Source Anchor | Phase | 验证入口 |
|------|----------|------------------|-------|---------|
| Primary path · 文本面试 happy path | 进入 practice 携带完整 InterviewContext + sessionId；渲染 TopBar / SessionMap / QuestionCard / Transcript / InputBar / 右侧 JD 关联 + AI 透明度 + 固定底部 CTA；首题来自 `getPracticeSession.currentTurn` | `screen-practice.jsx::PracticeScreen` lines 74-326（text 分支 196-256） | 1+2+3 | E2E.P0.044 + Vitest `practice/PracticeScreen.test.tsx` |
| Primary path · answer_submitted → ask_follow_up → ask_question | 用户输入回答 → `appendSessionEvent({kind:'answer_submitted', payload:{turnId, answerText}})` → 渲染 `assistantAction.type` 决定下一步 | `screen-practice.jsx::send` lines 59-66 + `TranscriptMsg` lines 592-615 | 2 | E2E.P0.044 + Vitest `hooks/usePracticeEvents.test.ts` + `components/AssistantActionRenderer.test.tsx` |
| Primary path · `appendSessionEvent` 不带 `Idempotency-Key` | hook 内部断言 request init 不含该 header；body 必含 `clientEventId`（UUIDv7 from `lib/ids.ts`）+ `kind` + `occurredAt` + `payload` | spec D-12 + OpenAPI `PracticeSessionEventRequest` + fixture | 2 | Vitest `idempotencyContract.test.ts` 反向断言；scenario verify.sh grep `Idempotency-Key.*appendSessionEvent` 应 0 命中 |
| Primary path · completePracticeSession → generating handoff | 用户点击「结束并生成报告」 → 调 `completePracticeSession({clientCompletedAt})` 带 `Idempotency-Key`；202 返回 `ReportWithJob{reportId, job}` → nav `generating`，携带稳定 InterviewContext ID + `PracticeDisplayContext`，但 body 不带展示字段 | `screen-practice.jsx::finishAndGenerate` lines 38-45 + spec §2.1 D-13 | 4 | E2E.P0.047 + Vitest `hooks/useCompletePracticeSession.test.ts` + `utils/practiceHandoffParams.test.ts` |
| Alternate path · assisted vs strict 显隐 | `practiceMode==='assisted'` 渲染：左侧 LIVE NOTES / hint button / 右侧可调用经历卡片；`practiceMode==='strict'` 隐藏全部上述 + 右侧渲染 strict-mode banner | `screen-practice.jsx` lines 170-179, 219-243, 276-300 | 3 | E2E.P0.045 + Vitest `components/RightPanel.test.tsx` + `hooks/usePracticeAssistance.test.ts` |
| Alternate path · practiceGoal 不改变显隐 | 任意 `practiceMode` × `practiceGoal∈{baseline,retry_current_round,next_round}` 组合，显隐策略一致；practiceGoal 仅影响题目来源（B2 拥有，前端不渲染该差异） | spec D-3 + `module-practice-review.md` §6 | 3 | Vitest `practiceGoalParity.test.tsx` + `usePracticeAssistance.test.ts` 覆盖 6 组合；负向 grep `goal === 'debrief'` 应 0 命中（测试断言除外） |
| Alternate path · 模式切换 segmented control（text/voice） | 顶部 segmented control 渲染 text + voice 两按钮（aria-checked、active 高亮）；text mode 保持本 plan 的 answer/hint/skip/pause/complete event loop；voice mode 由 `practice-voice-mvp/001` 接管 `PracticeVoiceSurface` 与 `usePracticeVoiceTurn` | `screen-practice.jsx` lines 91-114 (segmented) + practice-voice-mvp/001 | 1+3 | Vitest `components/TopBar.test.tsx` + `practiceModeSwitch.test.tsx` + P0.044/P0.047 verify 限定 `createPracticeVoiceTurn` 只在 voice owner hook |
| Alternate path · hint_requested 提示流（assisted） | 用户点击「提示」 → `appendSessionEvent({kind:'hint_requested', payload:{turnId}})`；fixture-only `show-hint` 用于前端 UI / `hintCount++` / `hintUsed='true'` 验证；当前真实 backend-practice/002 对 hint 默认返回 `PRACTICE_SESSION_CONFLICT`（D-34），assisted 200 正向由 backend-practice/003 接手，前端必须同时覆盖 409 防御分支 | `screen-practice.jsx` lines 219-243 + `module-practice-review.md` §6 + backend-practice D-34 | 3 | E2E.P0.045 + Vitest `practiceHints.test.tsx` + `practiceConflict.test.tsx` |
| Alternate path · turn_skipped 跳过 | 用户点击「跳过」 → `appendSessionEvent({kind:'turn_skipped', payload:{turnId}})` → 渲染下一题；当前 turn `status='skipped'` 写入题目地图 | `screen-practice.jsx` lines 249 + `SessionMap` | 2+3 | E2E.P0.045 + Vitest `practiceSkip.test.tsx` |
| Alternate path · session_paused / session_resumed | 用户点击「暂停」 → `appendSessionEvent({kind:'session_paused'})` + 本地 timer 暂停；session.status 切到 `waiting_user_input`；点「继续」 → `session_resumed` + timer 恢复；暂停期间禁用 submit / hint / skip | `screen-practice.jsx` lines 87-89 (pause button) + 17-21 (timer effect) | 3 | E2E.P0.044 / E2E.P0.045 + Vitest `practicePauseResume.test.tsx` |
| Failure / recovery · append AI timeout / complete 5xx | `appendSessionEvent` 返回 502 `AI_PROVIDER_TIMEOUT` 或 `completePracticeSession` 返回网络 / 5xx → 输入区下方 InlineError + retry 按钮；retry 复用同一 `clientEventId`（append）或同一 `Idempotency-Key`（complete） | spec §6 C-12 + backend-practice C-5 | 4 | E2E.P0.046 + Vitest `practiceErrors.test.tsx`（502 / 5xx 子用例） |
| Failure / recovery · session 404 / PRACTICE_SESSION_NOT_FOUND | `getPracticeSession` / `appendSessionEvent` / `completePracticeSession` 返回 404 → 渲染 `PracticeSessionLostState`，CTA「返回 workspace」调 `nav("workspace", {targetJobId, jdId, planId, resumeVersionId})` | `screen-practice.jsx` 兜底 + spec §4 缺 session/plan 时回 workspace | 1+4 | E2E.P0.046 + Vitest `practiceSessionLost.test.tsx` |
| Failure / recovery · appendSessionEvent 409 PRACTICE_SESSION_CONFLICT | 仅在 strict + hint 路径触发（理论上 UI 已隐藏按钮，但 fixture 故意构造）→ 渲染「严格模拟不允许提示」内嵌警告 banner，不重试 | spec D-12 + backend-practice D-34 | 4 | Vitest `practiceConflict.test.tsx` 触发 + fixture variant `hint-strict-conflict` |
| Failure / recovery · `clientEventId` 冲突 (409 mismatch) | 同一 `clientEventId` 改 payload → `appendSessionEvent` 返回 409；UI 显示「同步异常，请刷新」，触发 `getPracticeSession` 重拉，丢弃本地 stale 状态 | OpenAPI fixture `mismatch` scenario + spec D-12 | 4 | E2E.P0.046 + Vitest `practiceClientEventConflict.test.tsx` |
| Failure / recovery · 网络错误 / 5xx | 通用错误占位 + retry；retry 复用同一 batch；3 次失败后展示「返回 workspace」fallback | spec §4 错误态 | 4 | Vitest `practiceErrors.test.tsx`（network 子用例） |
| Boundary · `getPracticeSession` resume / refresh | visibility change / window focus / 网络恢复时调 `getPracticeSession(sessionId)` 重拉；server state 与 local state diff 采用「server wins」（不写本地缓存 transcript 明文） | spec §2.1 + getPracticeSession fixture 3 variant | 1+4 | E2E.P0.046 + Vitest `usePracticeSessionLoader.test.ts` |
| Boundary · 空 transcript 进入 | 用户首次进入 practice 时 transcript 只有首题（来自 `currentTurn`），无历史 user message；空状态文案「— 可以暂停、请求提示，或跳过 —」 | `screen-practice.jsx` lines 212-214 | 1 | Vitest `Transcript.test.tsx`（empty 用例） |
| Boundary · turn budget 用完 | 当 `assistantAction.type='session_completed'`（B2 触发 budget exhausted）→ 自动渲染「面试已完成」提示 + 自动定位到底部 CTA；不再接受用户输入；用户点 CTA 触发 completePracticeSession | spec D-12 + `module-practice-review.md` §3 | 3 | Vitest `practiceCompletion.test.tsx` |
| Boundary · paused 状态 submit / hint / skip 被禁用 | `session.status='waiting_user_input'` 或本地 `paused` → 这三个按钮 disabled 且不发请求；resume 后恢复 | spec D-12 + ui-design pause toggle | 3 | Vitest `practicePauseResume.test.tsx` |
| Cross-layer contract · PracticeSessionEventRequest schema | body `{clientEventId, kind, occurredAt, payload}` 与 OpenAPI schema 一致；kind 限于 5 个 enum；payload 按 kind 类型化（`answer_submitted` → `{turnId, answerText}`；`hint_requested` / `turn_skipped` → `{turnId}`；`session_paused` / `session_resumed` → `{}`）；non-side-effect 调用不带 `Idempotency-Key` | OpenAPI `PracticeSessionEventRequest` + appendSessionEvent fixture | 2 | mock-contract-suite parity + Vitest `appendSessionEventBody.test.ts` |
| Cross-layer contract · CompletePracticeSessionRequest schema | body 仅 `{clientCompletedAt: ISO8601}`；side-effect 调用带 `Idempotency-Key`；展示字段（mode/modality/practiceMode/practiceGoal/hintUsed/hintCount）**不**进入 body | OpenAPI `CompletePracticeSessionRequest` + spec D-13 | 4 | Vitest `completePracticeSessionBody.test.ts` + 负向 grep `hintCount` 在 complete body 应 0 命中 |
| Cross-layer contract · AssistantAction 5 type 消费 | renderer 覆盖 `ask_question` / `ask_follow_up` / `show_hint` / `session_wait` / `session_completed`；`provenance` 字段（promptVersion / rubricVersion / modelId / language / featureFlag / dataSourceVersion）仅渲染到右侧 AI 透明度卡，**不**渲染到主对话流 | OpenAPI `AssistantAction` + `screen-practice.jsx` AI TRANSPARENCY 块 lines 293-298 | 2 | Vitest `AssistantActionRenderer.test.tsx` + `AiTransparency.test.tsx` |
| Cross-layer contract · SessionStatus 七值消费 | `queued / running / waiting_user_input / completing / completed / failed / cancelled` 各自分支渲染：queued=等待首题/会话准备中 occupy；running=主交互；waiting_user_input=禁用输入；completing=按钮置灰 + 「正在生成报告…」；completed=自动 nav generating；failed=显示错误 + retry / 返回 workspace；cancelled=显示「会话已取消」+ 返回 workspace；不引用 `draft / archived` 旧值 | OpenAPI `SessionStatus` + spec §2.1 + getPracticeSession fixture | 2+4 | Vitest `usePracticeSession.test.ts` 七分支断言 + 负向 grep `draft / archived` 旧值 0 命中 |
| Cross-layer contract · `generating` 路由 handoff 参数 | nav `generating` 携带稳定 `InterviewContext` ID（`planId / targetJobId / jdId / resumeVersionId / roundId / sessionId / reportId`）+ `PracticeDisplayContext`（`mode / modality / practiceMode / practiceGoal / hintUsed / hintCount`）；与 spec §2.1 route context + getFeedbackReport(reportId) 入口一致；不写 `report` / 真实 `getFeedbackReport` 调用（归 plan 004）；负向仅禁止 raw answer/question/hint/prompt/provenance 明文进入 URL | spec §2.1 + `screen-practice.jsx::finishAndGenerate` | 4 | Vitest `practiceHandoff.test.ts` 完整字段集断言 + 负向断言 `getFeedbackReport` 调用次数 0 |
| Cross-layer contract · `Prefer: example=<scenario>` fixture variant 切换 | 5 个 hook（loader / events / assistance / complete / pause-resume）通过 fixture-backed transport 切换 variant；scenario verify.sh 通过 `EI_FIXTURE_SCENARIO_*` 环境变量驱动 | `frontend/src/api/mockTransport.ts` createFixtureBackedFetch | 全 phase | Vitest `mockTransport.spy.test.ts`（断言 Prefer header）+ scenario setup.sh |
| Cross-layer contract · `interviewerPersona` UI-only 切换 | 顶部 RoleDropdown 切换 `general_interviewer / hr_screener / hiring_manager` 仅改本地展示文案与右侧 AI 透明度文案，**不**发送任何 backend 请求；plan 创建时 persona 已固定，run-time 切换是观感选项 | `screen-practice.jsx::RoleDropdown` lines 617-641 | 3 | Vitest `RoleDropdown.test.tsx` + 负向断言 generated client 调用次数为 0 |
| Privacy / security · raw answer text | `answerText` 仅出现在 `appendSessionEvent` body（fixture transport 内部）+ Transcript 组件 React state；不出现在 console.log / URL query / localStorage / telemetry / pendingAction.params | spec §4 隐私红线 + CLAUDE.md §2.1.3 | 2+5 | Vitest `practicePrivacy.test.tsx` + scenario verify.sh grep `answerText` 在 URL/localStorage/console 0 命中 |
| Privacy / security · questionText / hint / provenance | `questionText` 渲染到 QuestionCard / Transcript / SessionMap，但不进入 URL / localStorage / telemetry；`hint` 仅渲染到 HintBanner，不缓存；`provenance.modelId` 仅渲染到 AI 透明度卡，不进入 telemetry payload | spec §4 + screen-practice.jsx AI TRANSPARENCY | 2+5 | Vitest + scenario verify.sh grep |
| Privacy / security · `clientEventId` 不复用 sessionId / userId | `clientEventId` 通过 `lib/ids.ts::uuidv7()` 派生，且不嵌入 sessionId / userId 明文；每次 user-side action 生成新 id；重试同一 action 复用同一 id | spec D-12 + B1 idempotency 规则 | 2 | Vitest `clientEventIdContract.test.ts` 反查 |
| Privacy / security · `Idempotency-Key` 双轨边界 | `appendSessionEvent` **必无** `Idempotency-Key` header；`completePracticeSession` **必有**，并通过 `lib/conventions/idempotency.ts::newIdempotencyBatch().complete` 派生 | spec D-12 / D-13 + OpenAPI | 2+4 | Vitest `idempotencyContract.test.ts` 双向断言 |
| Observability | mockTransport spy 仅记录 status / latency / 4xx code / scenario name；不带 body；不带 `answerText` / `hint` / `questionText` / `provenance` | spec §4 + 001 plan 同款 mockTransport spy | 全 phase | Vitest `mockTransport.spy.test.ts` |
| UX · loading state | practice mount 阶段 `getPracticeSession` 拉取占位（≥1 viewport 不闪烁）；TopBar 显示 skeleton 题号；submit / hint / skip 按钮在请求中 disabled + spinner | `screen-practice.jsx` 隐式 skeleton | 1+3 | Vitest fake timer + `practiceLoading.test.tsx` |
| UX · empty state | `currentTurn` 缺失（理论上不应发生，但兜底）→ 渲染「等待首题…」occupy；进入时 `sessionId` 缺失 → 渲染 `PracticeSessionLostState` 并自动返回 workspace | spec §4 缺 session/plan | 1 | Vitest `practiceMissingSession.test.tsx` |
| UX · error state | `getPracticeSession` 5xx / `appendSessionEvent` 5xx / `completePracticeSession` 5xx 各自 inline 错误 + retry | spec §4 | 4 | Vitest `practiceErrors.test.tsx` 三子用例 |
| UX · i18n zh/en | 全文案通过 typed locale helper；新增 `practice.*` namespace ≥ 40 keys（toolbar / question / transcript / input / hint / skip / submit / pause / resume / strict / finish / generatingPending / errorAi / errorSession / errorConflict / sessionLost / voiceComingSoon / role.* / ai.* / 等）；切换立即重绘 | 001 D1 typed locale helper | 1-5 | Vitest `practiceI18n.test.tsx` + namespace 同步断言 |
| UX · dark + customAccent + 主题切换 | practice 三栏 + TopBar + InputBar + RightPanel + 底部 CTA 在 8 主题 × dark 组合下 computed background / color 出现可见变化 | D2 `data-theme / data-mode / data-custom-accent` + `ui-design/src/primitives.jsx` | 5 | Playwright `tests/pixel-parity/practice.spec.ts` 主题循环 |
| UX · responsive layout (mobile 390×844) | TopBar 控件折行（公司/岗位顶部 + 工具次行）；中部 grid 三栏折叠为单列 + 底部 sheet（输入区 + CTA sticky）；左栏题目地图变成折叠 drawer；右栏 JD 关联变成底部 Accordion；CTA 与 BindingPill 不溢出 | `module-practice-review.md` §4 + spec §4 mobile | 1+5 | Playwright mobile project + Vitest jsdom 视口模拟 |
| UI source structure parity · TopBar | 公司/岗位 left block + 右侧控件链（RoleDropdown + Question Tag + Timer Tag + Pause button + segmented mode control + voice live indicator + 严格模拟 toggle role='switch' aria-checked）；testid `practice-topbar-{company,title,role,question,timer,pause,mode-text,mode-voice,live,strict}` | `screen-practice.jsx` lines 76-134 | 1 | Vitest + `practice-topbar-*` testid 命中 + 控件类型断言（不是 select） |
| UI source structure parity · SessionMap | 左栏 SESSION MAP label + 题目 list（圆角圆点 + 索引 + 主题 + duration）+ active / done / pending 三态；strict 模式隐藏 LIVE NOTES；assisted 模式渲染 LIVE NOTES 卡片 | `screen-practice.jsx` lines 138-180 | 1+3 | Vitest + testid `practice-sessionmap-{label,item-${idx},live-notes,live-notes-ok,live-notes-warn}` |
| UI source structure parity · QuestionCard | 顶部 padding + Tag(Q index + topic) + 多 Tag(currentQ.tags) + serif 字体 question prompt | `screen-practice.jsx` lines 196-205 | 1 | Vitest + testid `practice-question-{badge,topic,tag-${idx},prompt}` |
| UI source structure parity · Transcript | 滚动容器 + TranscriptMsg 列表（用户 / AI / followUp Tag / mono 时间戳）+ 底部 helper 文案 | `screen-practice.jsx` lines 208-215 + `TranscriptMsg` lines 592-615 | 2 | Vitest + testid `practice-transcript-{container,message-${idx},follow-up-badge-${idx},helper}` |
| UI source structure parity · InputBar | textarea + dictation-listening banner + transcript-failed banner + hint button (assisted) + dictation toggle + skip + send；textarea placeholder zh/en；button label zh/en | `screen-practice.jsx` lines 218-254 | 2+3 | Vitest + testid `practice-input-{textarea,hint,dictate,skip,send,dictation-banner,transcript-failure-banner}` |
| UI source structure parity · HintBanner | assisted + showHint=true 时 amberSoft 背景 banner + `提示:` / `Hint:` 前缀 + hint 文本 | `screen-practice.jsx` lines 219-223 | 3 | Vitest + testid `practice-hint-banner` |
| UI source structure parity · RightPanel (assisted) | JD LINK 卡片 + RELEVANT EXPERIENCE 卡列表 + AI TRANSPARENCY mono 文本 + 底部 CTA `结束并生成报告` + hint usage 标记 | `screen-practice.jsx` lines 260-322 | 3+4 | Vitest + testid `practice-rightpanel-{jd,exp-${idx},ai-transparency,cta-finish,hint-count}` |
| UI source structure parity · RightPanel (strict) | 用 strict-mode banner 替换 hint / experience；其余结构（JD link + AI transparency + 底部 CTA）保留 | `screen-practice.jsx` lines 276-281 | 3 | Vitest + testid `practice-rightpanel-{strict-banner}` |
| Voice owner co-location boundary | voice surface / audio capture / `createPracticeVoiceTurn` 属于 `practice-voice-mvp/001`，允许在 `usePracticeVoiceTurn.ts` 与 voice tests 出现；text-loop scenario 不把 repo-wide 0 命中作为完成依据 | practice-voice-mvp/001 | 1+5 | P0.044/P0.047 verify 限定 `createPracticeVoiceTurn` 只在 voice owner hook；`frontendOwners.realApiMode.test.ts` 证明 real generated client route |
| UI source structure parity · PracticeSessionLostState | sessionId 缺失 / 404：卡片 + 文案「会话已结束或不存在」+ CTA「返回 workspace」 | spec §4 缺 session 兜底 | 1+4 | Vitest + testid `practice-session-lost-{title,desc,cta}` |
| UI visual geometry parity · desktop | 1440×900 practice 主屏 + voice surface + session-lost + transcript 长滚动场景 bounding box stays in viewport, no overlap；TopBar 高度固定 ≤ 64px；底部 CTA sticky 不被遮挡 | n/a | 5 | Playwright `tests/pixel-parity/practice.spec.ts` desktop project |
| UI visual geometry parity · mobile | 390×844 三栏折叠 + 输入 sheet sticky + RoleDropdown drawer + 顶部工具链不重叠 | n/a | 5 | Playwright mobile project |
| UI visual geometry parity · dark / customAccent / theme | 8 主题 × dark + customAccent oklch 切换可见变化 | n/a | 5 | Playwright |
| UI visual geometry parity · screenshot regression | toHaveScreenshot baseline maxDiffPixels 阈值（与 workspace.spec.ts 同款配置） | n/a | 5 | Playwright + frontend baseline |
| UI stale-contract negative · 旧 route alias | 旧 `voice` route alias、独立 `voice` route entry、`VoicePracticeScreen` testid、`PlanScreen` testid 在 practice 新代码中 0 命中（不计 `normalizeRoute` alias map） | spec §2.2 + frontend-shell D1 alias 表 | 全 phase | Vitest + scenario verify negative grep |
| UI stale-contract negative · 旧 enum / 旧文案 | 旧 `practiceMode='debrief'` value、旧 `practiceGoal='debrief'` value、旧 `切到语音` 文案、旧 `reportLayout` hash、旧 `featureKey` 路由口径、旧 `mistakes` / `growth` / `drill` 入口在 practice 模块 0 命中 | spec §2.2 + history.md | 全 phase | grep negative |
| UI stale-contract negative · 不直接 import prototype | `frontend/src/app/screens/practice/` 不 import `ui-design/src/data.jsx` / `window.EI_DATA` / `getPracticeSampleTranscript` / `getPracticeSampleQuestions` 等 prototype helper | n/a | 全 phase | Vitest + tsc grep |
| UI stale-contract negative · report/voice owner 边界 | practice text event loop 不调 `getFeedbackReport`（归 report/generating owner）；`createPracticeVoiceTurn` 只能由 voice owner hook 消费，不得散落到 text event hooks / completion handoff | spec §5.1 operation matrix + practice-voice-mvp/001 | 全 phase | Vitest spy + P0.044/P0.047 verify |
| Regression / legacy-negative · 工作区 + 后端契约 | `E2E.P0.018-021`（workspace）+ `E2E.P0.022-026`（backend-practice 001 启动 / 首题 / idempotency / 隐私 shell 场景）+ `E2E.P0.038-043`（backend-practice 002 event loop / complete Go HTTP scenario）全部作为真实 regression gate 重跑；fixture-backed PASS 只能证明前端 mock 合同，不替代真实 handler | n/a | 5 | scenario rerun + `cd backend && go test ./cmd/api -run 'TestE2EP0038|TestE2EP0039|TestE2EP0040|TestE2EP0041|TestE2EP0042|TestE2EP0043' -count=1` |
| Regression / legacy-negative · 不直接调用 LLM | practice 模块不出现 AI provider key / provider registry / prompt registry / AIClient / LLM endpoint / bypass generated client 的 ad hoc fetch | n/a | 全 phase | Vitest + grep negative |
| BDD 主路径 + 关键分支 + 失败恢复 + 旧口径负向 | 见 [bdd-plan.md](./bdd-plan.md) 4 场景矩阵 | n/a | 1-5 | E2E.P0.044/045/046/047 + Playwright contract |

### 高风险类别 N/A 说明

- **隐私 / 安全 · audio buffer**：本 text-loop plan 不直接处理 audio buffer；voice surface / STT / TTS / barge-in 的 raw audio 红线由 `practice-voice-mvp/001` 承接。P0.044/P0.047 只反查 text-loop 路径不泄漏 raw answer/question/hint/provenance。
- **Privacy · LLM prompt raw text**：B2 在服务端 redact prompt；前端不直接调用 LLM；`provenance` 字段（`promptVersion / modelId`）只是版本/标识，不含 prompt body；前端只渲染版本号到右侧 AI 透明度，因此 prompt-response 明文不在前端泄漏面。N/A 原因记录在此。
- **Voice session failure**：完整 voice session 的 transcription / waveform / TTS 失败恢复由 `practice-voice-mvp/001` owner；本 plan 的 text-mode speech-to-text failure banner 仍属于文本 surface parity，按 `screen-practice.jsx` lines 583-590 派生为正式前端组件，但不得直接 import `ui-design` 源。

## 3.6 Frontend / Backend Operation Matrix

本 plan 走 `docs/development.md` §2.2 Frontend-First Path：正式前端先对齐 `ui-design/` 并通过 generated client + fixture-backed transport 完成 P0 UI/BDD；同时当前 repo 已有 backend-practice/002 真实 handler，Phase 5 必须跑对应真实 handler regression。fixture-backed PASS 不等于真实 backend 闭环；尤其 assisted hint 的 `show_hint` 正向 UI 在本 plan 内仍是 fixture-only，真实 backend 支持由 backend-practice/003 接管。

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getPracticeSession` | 当前已有 `default` / `prototype-baseline` / `missing-session`；本 plan 计划新增 `running-with-history`（含 turn history view-model）/ `queued` / `completing` 用于 UI 状态覆盖 | `usePracticeSessionLoader` mount + visibility/focus refresh；missing-session → `PracticeSessionLostState` | 已落地：`backend/internal/api/practice/handler.go::GetPracticeSession` + `backend/internal/practice.Service.GetPracticeSession` + store read path；真实 path 由 backend-practice 001/002 regression 覆盖 | `practice_sessions` + `practice_turns` | none in frontend | frontend `E2E.P0.044 / E2E.P0.046`；backend `E2E.P0.023` |
| `appendSessionEvent` | 当前已有 `default / follow-up / hint-strict-conflict / turn-skipped / pause-resume / replay / mismatch / completed`；本 plan 只在需要时新增 `show-hint`（fixture-only until backend-practice/003）与 `ai-timeout` UI fixture，并在 mock-contract-suite 记录 | `usePracticeEvents` 5 个 mutation；body 不带 `Idempotency-Key`；`clientEventId` 来自 `lib/ids.ts::uuidv7()` | 已落地：`backend/internal/api/practice/session_event_handlers.go::AppendSessionEvent` + `backend/internal/practice.AppendSessionEvent` + `backend/internal/store/practice.AppendSessionEvent`；真实 backend 002 中 `hint_requested` 默认 409（D-34），assisted 200 归 backend-practice/003 | `practice_session_events` + `practice_turns` + outbox | backend-only F3 `practice.session.follow_up`；`show_hint` / lightweight observe real path 不在 backend 002 | frontend `E2E.P0.044 / E2E.P0.045 / E2E.P0.046`；backend `E2E.P0.038 / E2E.P0.039 / E2E.P0.040 / E2E.P0.043` |
| `completePracticeSession` | 当前已有 `default / replay / mismatch / session-already-completed / cross-user-not-found`；complete path 无 AI，不能要求 `ai-timeout` fixture | `useCompletePracticeSession` 由底部 CTA 触发；带 `Idempotency-Key`（同会话 finish 同一 key）；body 仅 `{clientCompletedAt}` | 已落地：`backend/internal/api/practice/session_event_handlers.go::CompletePracticeSession` + idempotency middleware + `backend/internal/practice.CompletePracticeSession` + `backend/internal/store/practice.CompleteSession` | session status + feedback_reports + async_jobs + outbox + idempotency_records | none in frontend; no AI in completion path | frontend `E2E.P0.047`；backend `E2E.P0.041 / E2E.P0.042 / E2E.P0.043` |
| `getFeedbackReport` | N/A（生成态 + 报告 owner） | 本 plan **不消费**；handoff 到 `generating?reportId` 后由 report/generating owner 轮询 | backend-review real handler | feedback_reports | backend-review only | 负向断言 + E2E.P0.047 + real-mode gate |
| `createPracticeVoiceTurn` | `openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json` | Text event loop **不消费**；voice owner hook `usePracticeVoiceTurn` 消费；P0.044/P0.047 只限制该 operation 不散落到 text event hooks / completion handoff | practice-voice/backend-practice real handler | voice events | STT/LLM/TTS backend-only | practice-voice owner + real-mode gate |

## 3.7 InterviewContext × PracticeDisplayContext View-Model Mapping

正式前端不得从 `ui-design/src/data.jsx` 或未声明 fixture 字段补齐 `InterviewContext` 之外的数据。本 plan 在 001 已落地的 `InterviewContext` reducer 基础上额外消费 `MERGE_SESSION` action 把 `getPracticeSession` / `appendSessionEvent` 返回的 `session` 合并进 store。具体 mapping：

| 字段 | Source | Rule |
|------|--------|------|
| `sessionId` | route param 或 `InterviewContext.sessionId`（由 001 workspace `MERGE_SESSION` 写入） | 必填；缺失 → `PracticeSessionLostState`，CTA 返回 workspace |
| `planId / targetJobId / jdId / resumeVersionId / roundId / roundName` | route param 或 InterviewContext（001 已 hydrate） | 本 plan 只读，不修改；用于 finishAndGenerate 时跳 generating，并允许作为稳定 owner route context 继续传递 |
| `mode / modality` | route param（默认 `text/text`） | 本 plan UI 切换 voice → 仅触发 `nav("practice", {mode:'voice', modality:'voice'})`，不调 backend |
| `practiceMode` | route param（默认 `strict`） | 二值 `assisted / strict`；session mode 由后端/plan 创建时确定，本 plan 渲染层只读；顶部 strict switch 保留视觉与 a11y，但点击只给用户提示，不改 backend mode |
| `practiceGoal` | route param（默认 `baseline`） | 显隐策略与此值**无关**（D-3 verifier） |
| `hintUsed / hintCount` | route param（默认 `'false' / '0'`）+ practice 运行时增量 | 用户每次成功 `hint_requested`（200 OK）→ `hintCount++`，`hintUsed='true'`；写回 InterviewContext（通过新 reducer action `INCREMENT_HINT_COUNT`）；strict 模式不计数（按钮不渲染） |
| `currentTurn` | `getPracticeSession.currentTurn` 或 `appendSessionEvent.session.currentTurn` | 仅在本 plan 内通过 React state 缓存；不持久化到 localStorage |
| `transcript` (本地) | `appendSessionEvent` 成功后 append `{role:'user', text:answerText, t:elapsed}` 与 `{role:'ai', text:assistantAction.questionText \| hint, t:elapsed, followUp:type==='ask_follow_up'}` | 不与 server 同步；refresh 时 server wins 重置 |
| `elapsed` (本地) | React useEffect interval timer；paused 时停止；可被 `getPracticeSession.session.elapsedSeconds`（若 fixture 添加）覆盖 | 不进入 backend body |
| `assistantAction` | `SessionEventResult.assistantAction` | 临时状态；下一个 event 完成后被替换 |

新增 InterviewContext reducer action（在 001 reducer 基础上扩展，不破坏 001 测试）：

- `INCREMENT_HINT_COUNT`：把 `hintCount` 解析为 number 自增并写回 string；`hintUsed='true'`；幂等（依赖 `clientEventId`，由调用方保证）。
- `MERGE_TURN`：把 `currentTurn` 写入 InterviewContext（可选；本 plan 内部 state 也可，但若后续需要把当前题号带到 generating 可能需要）。

> 备注：001 已有 `MERGE_SESSION`，承接 startPracticeSession 返回。本 plan **新增的 reducer action** 必须 reuse 001 的 `interviewContextReducer.test.tsx` 测试结构，并补齐对应单元测试（在 001 测试文件追加，不创建并行 reducer）。

## 4 实施步骤

### Phase 1: PracticeScreen 静态壳 + 路由替换 + i18n + sessionId 守卫

#### 1.1 新增 `frontend/src/app/screens/practice/PracticeScreen.tsx`

按 `ui-design/src/screen-practice.jsx::PracticeScreen` 文本分支（lines 74-326 中 text 部分 184-326）源级复刻渲染：TopBar（公司/岗位 + RoleDropdown + Question Tag + Timer Tag + Pause + segmented mode control + voice live indicator + 严格模拟 toggle）+ 中部 grid（260px / 1fr / 280px）+ 左栏 SessionMap + 中栏 QuestionCard + Transcript + InputBar + 右栏 RightPanel + 底部固定 CTA。本 phase 不接入 API：所有动态字段渲染占位 skeleton；`send` / `requestHint` / `skipTurn` / `pauseSession` / `resumeSession` / `finishAndGenerate` callback 仅记录调用次数；voice mode 当前由 `practice-voice-mvp/001` owner 接管。

#### 1.2 新增 `frontend/src/app/screens/practice/components/{TopBar, SessionMap, QuestionCard, Transcript, InputBar, RightPanel, HintBanner, LiveNotes, FinishCta, PracticeSessionLostState, RoleDropdown, ExpCard, ErrorState}.tsx`

每个组件接受 typed props，对应 §3.5 UI source structure parity 行；从 `ui-design/src/screen-practice.jsx` 同名片段复刻 DOM；不引入 ui-design `VoiceSessionSurface` / `PracticeWaveformBars` / `PracticeAnnotatedWaveform` / `VoiceExpressionPanel` 等 voice session surface import。Text input 的 speech-to-text failure banner 独立实现为正式前端组件（建议 `DictationFailureBanner`），追溯 `screen-practice.jsx` lines 583-590，但不得直接 import `ui-design` 源。

#### 1.3 新增 `frontend/src/app/screens/practice/hooks/usePracticeSessionLoader.ts`

通过 generated client 调 `getPracticeSession(sessionId)`；React state 跟踪 `idle / loading / data / sessionLost / error` 五态；mount 时若 `InterviewContext.sessionId` 缺失 → 立即返回 `sessionLost` 状态（不发请求）；返回数据通过 `MERGE_SESSION` 写入 `InterviewContext`；暴露 `refresh()` 手动重拉；提供 `useEffect` 监听 visibility / focus / online 事件并自动 refresh（在挂载点确认后）。

#### 1.4 路由壳替换

在 `frontend/src/app/App.tsx` `renderRouteScreen` 中绑定 `practice` → `<PracticeScreen route={route} />`（替换 D1 `PlaceholderScreen`）；保持 `NO_CHROME_ROUTES` 隐藏 TopBar；`generating` / `report` 仍由其他 owner 接管，旧 `company_intel` route 必须归一回 `workspace`，不在本 plan materialize。

#### 1.5 i18n locale 文件扩展

在 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 中新增 `practice.*` 命名空间（≥ 40 key 与 `screen-practice.jsx::L` zh/en 字典等价：toolbar.questionTag / toolbar.timer / toolbar.pause / toolbar.resume / toolbar.modeText / toolbar.modeVoice / toolbar.voiceLive / toolbar.voicePaused / toolbar.strict / toolbar.role.general / toolbar.role.hr / toolbar.role.manager / sessionMap.label / sessionMap.liveNotes / sessionMap.liveNotesOk / sessionMap.liveNotesWarn / sessionMap.liveNotesNote / question.tagPrefix / transcript.helper / transcript.followUp / transcript.aiLabel / transcript.userLabel / input.placeholder / input.dictationListening / input.dictationToggleOn / input.dictationToggleOff / input.hint / input.skip / input.send / hint.prefix / rightpanel.jdLink / rightpanel.jdLinkProbes / rightpanel.experienceLabel / rightpanel.strictBanner / rightpanel.aiTransparency / rightpanel.finishCta / rightpanel.hintUsageNote / voiceComingSoon.title / voiceComingSoon.desc / voiceComingSoon.backToText / sessionLost.title / sessionLost.desc / sessionLost.cta / errors.aiTimeout / errors.network / errors.sessionConflict / errors.unknown / errors.retry / errors.backToWorkspace）；`messages.ts` 类型聚合补齐。

#### 1.6 Vitest 红灯 → 绿灯

新增 `practice/__tests__/PracticeScreen.test.tsx`：测 i18n zh/en 切换重绘、≥ 20 个 `practice-*` testid 存在（按 §3.5 UI source structure parity rows）、所有 callback prop 触发记录、控件类型断言（segmented mode control 不是 `<select>`、strict toggle role='switch' aria-checked、RoleDropdown 是 menu hierarchy 而非 select）、负向断言不 import ui-design voice 组件、负向断言不出现旧 prototype testid（`practice-voice-*` 即 voice surface dom、`practice-mode-card-*`、`growth-*`、`mistakes-*`）。

新增 `practice/__tests__/usePracticeSessionLoader.test.ts`：测 idle / loading / data / sessionLost (404 + missing) / error (5xx) 五态、refresh 调用、visibility/focus/online 触发 refresh、`MERGE_SESSION` 调用次数。

#### 1.7 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.044` 中 practice 静态壳 + sessionId 守卫部分资产构建到 ready 态。

### Phase 2: appendSessionEvent 单 endpoint + 5 kind 路由 + AssistantAction 渲染

#### 2.1 新增 `frontend/src/app/screens/practice/hooks/usePracticeEvents.ts`

实现 5 个 mutation：

```ts
function buildEventRequest(kind, payload, occurredAt = isoNow()): PracticeSessionEventRequest {
  return { clientEventId: uuidv7(), kind, occurredAt, payload };
}

async function submitAnswer({ turnId, answerText }) {
  const req = buildEventRequest('answer_submitted', { turnId, answerText });
  return client.appendSessionEvent(sessionId, req); // 不传 Idempotency-Key
}
// requestHint / skipTurn / pauseSession / resumeSession 同款
```

调用必须断言 fetch init 不含 `Idempotency-Key` header；`clientEventId` 由 `lib/ids.ts::uuidv7()` 派生；同一 user action 失败 retry 时复用同一 `clientEventId`（hook 内 ref 缓存当前 batch，retry 调同一 batch；fresh action 生成新 batch）。

#### 2.2 新增 `frontend/src/app/screens/practice/components/AssistantActionRenderer.tsx`

消费 `SessionEventResult.assistantAction`：

- `ask_question` → append AI message + 推进 `qIdx` + 更新 SessionMap turn status
- `ask_follow_up` → append AI message with `followUp=true` badge
- `show_hint` → 设置 `HintBanner.visible=true` + 写入 `hint` 文本；`INCREMENT_HINT_COUNT` 写回 InterviewContext
- `session_wait` → 输入区显示「正在生成下一题…」disabled state
- `session_completed` → 渲染「面试已完成」提示 + auto-scroll 到底部 CTA

`provenance` 字段不出现在主对话流；只在 RightPanel.AI TRANSPARENCY 卡渲染 `promptVersion / rubricVersion / modelId / language / featureFlag`（不渲染 `dataSourceVersion`，保留兜底）；切换 fixture variant 时 transparency 卡更新。

#### 2.3 新增 `frontend/src/app/screens/practice/hooks/usePracticeSession.ts`

`SessionStatus` 七值消费器（严格以 generated `SessionStatus` 为准，不引入 `draft / archived` 旧值）：

- `queued` → 「正在准备面试…」occupy；输入 / hint / skip / finish 全部 disabled；允许 refresh
- `running` → 主交互（输入 / 按钮可用）
- `waiting_user_input` → 输入禁用 + 暂停态显示
- `completing` → 按钮置灰 + 「正在生成报告…」occupy
- `completed` → 自动 nav generating（仅一次；防抖）
- `failed` → 渲染 ErrorState + retry / 返回 workspace
- `cancelled` → 渲染「会话已取消」 + 返回 workspace

#### 2.4 fixture variant 扩展

与 `mock-contract-suite` owner / backend-practice/002 owner 同步，在 `openapi/fixtures/PracticeSessions/appendSessionEvent.json` 使用 backend-practice/002 已声明的 canonical scenario 名称；当前 `default` 文件可能仍是 `ask_follow_up`，Phase 2 必须先确认/修订 fixture truth 后再让前端测试依赖具体分支：

- `default`：answer_submitted → `ask_question`
- `follow-up`：answer_submitted → `ask_follow_up`
- `hint-strict-conflict`：hint_requested → 409 PRACTICE_SESSION_CONFLICT + `detail.policy='hint_disabled_in_mode'`
- `turn-skipped`：turn_skipped → 200 + `ask_question`
- `pause-resume`：session_paused / session_resumed → 200 + `session_wait` / resume current question
- `replay`：同 clientEventId 同 payload → 首次结果
- `mismatch`：同 clientEventId 改 payload → 409 `PRACTICE_SESSION_CONFLICT` + `detail.reason='client_event_fingerprint_mismatch'`
- `completed`：answer_submitted → `assistantAction.type='session_completed'` + `session.status='completing'`
- `ai-timeout`（frontend failure UI supplement）：502 `AI_PROVIDER_TIMEOUT` + retryable=true；若 mock-contract-suite 不接受该命名，BDD 必须改用 contract owner 确认的等价名称

在 `openapi/fixtures/PracticeSessions/getPracticeSession.json` 新增：

- `running-with-history`：current turn=3、turn history view-model（前端不需要 turn history fixture row 字段，仅用于截图基线）
- `queued`：status='queued'、currentTurn=null，用于准备态 UI
- `completing`：status='completing'、currentTurn.status='answered'

`make validate-fixtures`（或 `python3 scripts/lint/validate_fixtures.py --repo-root .`）通过；`make codegen-check` 零 drift。

#### 2.5 Vitest

新增 `practice/__tests__/usePracticeEvents.test.ts`：测 5 个 kind 的 request body 与 fetch init（method/path/header 反查 `Idempotency-Key` 不存在）；`clientEventId` 重试复用 + 新动作分配新 id；reset on success。

新增 `practice/__tests__/AssistantActionRenderer.test.tsx`：5 type 分支渲染 + provenance 只渲染到 AI transparency 卡的负向断言。

新增 `practice/__tests__/usePracticeSession.test.ts`：七个 generated `SessionStatus` 分支渲染 + 防抖 nav + 负向断言 `draft / archived` 旧值在实现中 0 命中。

新增 `practice/__tests__/idempotencyContract.test.ts`：appendSessionEvent 反向断言 `Idempotency-Key` 0 命中；completePracticeSession 正向断言（在 Phase 4 完善）。

新增 `practice/__tests__/appendSessionEventBody.test.ts`：body 与 OpenAPI `PracticeSessionEventRequest` schema 一致；payload 按 kind 类型化。

#### 2.6 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.044` 中 answer_submitted 主路径 + AssistantAction `ask_follow_up` / `ask_question` 渲染部分通过。

### Phase 3: assisted / strict 显隐 + RoleDropdown + 提示 / 跳过 / 暂停-恢复 + transcript

#### 3.1 新增 `frontend/src/app/screens/practice/hooks/usePracticeAssistance.ts`

派生显隐策略：

```ts
function deriveAssistance(practiceMode: string) {
  const isStrict = practiceMode === 'strict';
  return {
    showLiveNotes: !isStrict,
    showHintButton: !isStrict,
    showExperienceCards: !isStrict,
    showStrictBanner: isStrict,
  };
}
```

`practiceGoal` 完全不参与显隐计算（D-3 verifier）。Vitest 负向断言旧 `goal === 'debrief'` 0 命中（测试断言除外）。

#### 3.2 InputBar 提示流接线

assisted + 用户点击「提示」按钮 → `usePracticeEvents.requestHint({ turnId })`；成功 200 后 `INCREMENT_HINT_COUNT` 写回 InterviewContext（`hintCount++`、`hintUsed='true'`）；HintBanner 渲染 `assistantAction.hint`；再点击「提示」隐藏 banner（不重复发请求，除非用户重新打开输入或 banner 消失）。strict 模式下按钮 DOM 不渲染（不存在 disabled 形态）。

#### 3.3 InputBar 跳过 / 暂停-恢复接线

跳过按钮 → `skipTurn({ turnId })`；renderer 收到 `ask_question` 推进；当前 turn 在 SessionMap 标记 `status='skipped'`。

暂停按钮 → `pauseSession({})` + 本地 timer 暂停 + 设置 `session.status` 显示「已暂停」；恢复按钮 → `resumeSession({})` + 本地 timer 恢复。暂停期间 submit / hint / skip 三按钮 disabled 且不发请求。

#### 3.4 RoleDropdown 接线（UI-only）

通过 InterviewContext 派生当前 `interviewerPersona`（默认从 plan 创建时携带，001 已传到 route）；切换 dropdown 选项仅改本地 React state + RightPanel.AI TRANSPARENCY 卡的 role label，不发任何 backend 请求；Vitest 负向断言 generated client 调用次数 = 0。

#### 3.5 SessionMap turn 历史

题目地图通过 `assistantAction` 推进 `qIdx` + 维护本地 turn 状态数组（client-side cache，不持久化）；done / active / pending 三态；`turn.status` 渲染为 hashable 标记（`answered` ✓ / `skipped` ↷ / `follow_up_requested` 圆点 + 颜色）。

#### 3.6 fixture variant

`appendSessionEvent.json` 已在 Phase 2.4 扩展；本 phase 仅复用。`getPracticeSession.json` `running-with-history` 用于截图基线。

#### 3.7 Vitest

新增 `practice/__tests__/usePracticeAssistance.test.ts`：strict / assisted × baseline / retry_current_round / next_round 6 组合策略；负向断言 goal 不影响显隐。

新增 `practice/__tests__/practiceHints.test.tsx`：assisted hint 流（点 hint → API → renderer → InterviewContext hintCount 自增 → 报告字段 hintUsed='true'）；strict 模式下 hint button DOM 不存在。

新增 `practice/__tests__/practiceSkip.test.tsx`：跳过流（点 skip → API → SessionMap 标记）。

新增 `practice/__tests__/practicePauseResume.test.tsx`：暂停 → timer 停止、按钮 disabled、API 调用；恢复 → timer 启动、按钮可用、API 调用。

新增 `practice/__tests__/RoleDropdown.test.tsx`：切换选项仅改本地 state + AI 透明度卡 role label；generated client 调用次数 = 0。

新增 `practice/__tests__/practiceModeSwitch.test.tsx`：覆盖 text/voice segmented control 的路由与 owner 边界；`practice-voice-mvp/001` 后续接管 voice surface 渲染与 voice turn controller。

新增 `practice/__tests__/practiceGoalParity.test.tsx`：当前 core-loop practiceGoal 组合（assisted+baseline / assisted+retry_current_round / strict+baseline / strict+next_round）的显隐快照一致性。

#### 3.8 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.045` 中 strict / assisted × current practiceGoal 显隐主路径 + hint / skip / pause-resume 副路径通过。

### Phase 4: completePracticeSession + handoff + 错误恢复 + sessionLost / conflict 兜底

#### 4.1 新增 `frontend/src/app/screens/practice/hooks/useCompletePracticeSession.ts`

实现：

```ts
async function finish() {
  const idempotencyBatch = newIdempotencyBatch();
  const req: CompletePracticeSessionRequest = { clientCompletedAt: isoNow() };
  // body 不含 mode/modality/practiceMode/practiceGoal/hintUsed/hintCount
  const result = await client.completePracticeSession(sessionId, req, {
    idempotencyKey: idempotencyBatch.complete,
  });
  // result: ReportWithJob{reportId, job}
  navigate({
    name: 'generating',
    params: buildPracticeHandoffParams(ctx, result.reportId),
  });
}
```

`buildPracticeHandoffParams` 在 `frontend/src/app/screens/practice/utils/practiceHandoffParams.ts` 中实现，输出稳定 InterviewContext ID + 展示上下文：`{planId, targetJobId, jdId, resumeVersionId, roundId, sessionId, reportId, mode, modality, practiceMode, practiceGoal, hintUsed, hintCount}`；body 仍只包含 `{clientCompletedAt}`。URL / route 负向只禁止 `answerText / questionText / hint / promptHash / modelId / provenance` 等明文或 AI 细节。

`inFlightRef` + `Promise` 缓存防 StrictMode 双触发；retry 复用同一 `Idempotency-Key`；3 次失败后展示 `回到 workspace` fallback CTA。

#### 4.2 错误映射

| 错误码 / 类型 | 显示 | 动作 |
|--------------|------|------|
| 502 `AI_PROVIDER_TIMEOUT` (append) | InlineError「AI 暂时不可达，请重试」+ retry | 复用 `clientEventId` 重发 appendSessionEvent |
| 网络 / 5xx (complete) | InlineError「报告生成失败，请重试」+ retry | 复用 `Idempotency-Key` 重发 completePracticeSession |
| 404 `PRACTICE_SESSION_NOT_FOUND` | `PracticeSessionLostState` | 返回 workspace |
| 409 `PRACTICE_SESSION_CONFLICT` (strict + hint) | InlineWarning「严格模拟不允许提示」 | 不重试；按钮被禁用（理论上 UI 已 隐藏，但作为防御层） |
| 409 client_event_fingerprint_mismatch | InlineError「同步异常，请刷新」 | 自动 refresh `getPracticeSession`；server wins 重置本地 transcript |
| 网络 / 5xx 通用 | InlineError「网络错误」 + retry | 复用同 batch |

#### 4.3 sessionLost / completing / completed / cancelled 流转

- `getPracticeSession` 404 → `PracticeSessionLostState`
- `session.status='queued'` → 输入禁用 + 「正在准备面试…」
- `session.status='completing'` → 输入禁用 + CTA 置灰 + 「正在生成报告…」
- `session.status='completed'` → 自动 nav generating（仅一次）
- `session.status='cancelled'` → `PracticeSessionLostState`（带 cancelled 文案）+ 返回 workspace

#### 4.4 InterviewContext 扩展

在 001 reducer 基础上追加：

- `INCREMENT_HINT_COUNT`：把 `hintCount` 数字自增 + `hintUsed='true'`

在 001 `interview-context/InterviewContext.test.tsx` 文件追加该 action 测试（不创建并行 reducer）。

#### 4.5 Vitest

新增 `practice/__tests__/useCompletePracticeSession.test.ts`：happy path（点 CTA → 202 → nav generating）；replay（同 key 再发只触发一次 nav）；mismatch 409；网络 / 5xx retry 复用 key；StrictMode 双触发去重。

新增 `practice/__tests__/practiceHandoff.test.ts`：完整字段集断言（稳定 InterviewContext ID + `PracticeDisplayContext`）；body 不含展示字段；nav `generating` 路径正确；负向断言 raw answer / question / hint / prompt / provenance 不进 URL。

新增 `practice/__tests__/practiceErrors.test.tsx`：6 错误码各自渲染 + retry 行为。

新增 `practice/__tests__/practiceSessionLost.test.tsx`：404 → PracticeSessionLostState；返回 workspace 时 InterviewContext 不被清空（roundId / planId / targetJobId 仍在）。

新增 `practice/__tests__/practiceClientEventConflict.test.tsx`：`mismatch` fixture 触发 refresh + 重置本地 state。

新增 `practice/__tests__/completePracticeSessionBody.test.ts`：body 仅 `{clientCompletedAt}`；负向断言 mode/modality/practiceMode/practiceGoal/hintUsed/hintCount 不出现在 body JSON。

新增 `practice/__tests__/practiceCompletion.test.tsx`：session_completed assistant action 触发底部 CTA 高亮 + auto-scroll；点击 CTA → completePracticeSession。

#### 4.6 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.046` 失败恢复 + `E2E.P0.047` 完成 handoff 主路径 + 隐私红线通过。

### Phase 5: Pixel parity + 4 个 scenario + regression 重跑 + 文档与索引同步

#### 5.1 新增 `frontend/tests/pixel-parity/practice.spec.ts`

覆盖 desktop (1440×900) + mobile (390×844) 两 chromium project：

- DOM 锚点（TopBar + SessionMap + QuestionCard + Transcript + InputBar + RightPanel + 底部 CTA + HintBanner + PracticeSessionLostState + voice owner surface）
- 关键元素 bounding box stays in viewport, no overlap；TopBar 高度 ≤ 64px；底部 CTA sticky 不被遮挡；mobile 输入 sheet sticky bottom
- warm/light → dark → customAccent 三态切换 computed background / color 可见变化
- toHaveScreenshot baseline 区域：practice 主屏（assisted + strict + voice surface + session-lost + completing + completed）+ HintBanner

`pnpm --filter @easyinterview/frontend test:pixel-parity` 全 PASS（在 D2/D3 + home plan + workspace plan 现有基础上累加）。

#### 5.2 Scenario 资产

派生 4 个新 scenario 目录：

- `test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/`
- `test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/`
- `test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/`
- `test/scenarios/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/`

每个目录含 `README.md`（§6 baseline 维护、§7 离线运行限制）+ `scripts/{setup,trigger,verify,cleanup}.sh`（按 `test/scenarios/README.md` + `test/scenarios/e2e/README.md` 规范）+ `data/seed-input.md` + `data/expected-outcome.md`。

#### 5.3 Scenario INDEX 更新

`test/scenarios/e2e/INDEX.md` P0 表追加 4 行（`E2E.P0.044` / `E2E.P0.045` / `E2E.P0.046` / `E2E.P0.047`），关联需求列指向 `frontend-workspace-and-practice C-4 / C-8 / C-9 / C-10 / C-12`（按 §3.5 矩阵），状态 Ready；执行方式 automated。

#### 5.4 Regression 重跑

- `E2E.P0.018 / 019 / 020 / 021`（workspace fixture-backed contract gate）全部 PASS
- `E2E.P0.022 / 023 / 024 / 025 / 026`（backend-practice 001 plan/session/start/idempotency/privacy shell scenario）全部 PASS
- `E2E.P0.038 / 039 / 040 / 041 / 042 / 043`（backend-practice 002 event loop / complete / privacy Go HTTP scenario）全部 PASS，执行入口：`cd backend && go test ./cmd/api -run 'TestE2EP0038|TestE2EP0039|TestE2EP0040|TestE2EP0041|TestE2EP0042|TestE2EP0043' -count=1`
- `pnpm --filter @easyinterview/frontend test`（全量 Vitest）+ `pnpm --filter @easyinterview/frontend typecheck` + `pnpm --filter @easyinterview/frontend build` + `make build` 全 PASS

#### 5.5 文档与索引同步

`/sync-doc-index --fix-index` 把 `docs/spec/INDEX.md` 与 `docs/spec/frontend-workspace-and-practice/plans/INDEX.md` 同步到当前 Header；`make docs-check` zero drift；`check-md-links` OK；history.md 追加 plan 002 启动条目。

#### 5.6 负向搜索

- `frontend/src/app/screens/practice/` 不 import `ui-design/src/data.jsx` / `window.EI_DATA` / `getPracticeSampleQuestions` / `getPracticeSampleTranscript` / `getPracticeWaveformSamples` 等 prototype helper（0 命中）
- `frontend/src/app/screens/practice/` 不 import 已废弃 voice placeholder / 旧 voice route owner；voice surface 允许由 `practice-voice-mvp/001` 的 `PracticeVoiceSurface` / `usePracticeVoiceTurn` 承接；text-mode speech-to-text failure banner 必须是正式前端本地组件，禁止直接 import `ui-design` 源
- 旧 prototype practice 业务 testid（`practice-mode-card-*` / `growth-*` / `drill-builder-*` / `mistake-queue-*` / `practice-voice-*` voice surface dom）grep 0 命中（除负向断言文件）
- 旧 route alias（独立 `voice` / `voice_practice` / `welcome` / `growth` / `mistakes` / `drill` / `followup` / `experiences` / `star`）在 practice 模块中 grep 0 命中（除 `app/normalizeRoute.ts` alias map）
- 旧 enum 值 `practiceMode='debrief'` / `practiceGoal='debrief'`、旧文案 `切到语音` / `Switch to voice` 在 practice 模块 grep 0 命中
- raw answer text / questionText / hint text / AI provenance 详情 grep — 仅出现在 React state / generated client request body / fixture，不出现在 `console.log` / URL / `localStorage` / telemetry 调用
- LLM/provider grep — practice 模块不出现 provider key、provider registry、prompt registry、AIClient、LLM endpoint 或 ad hoc 绕过 generated client 的 fetch；`provenance.modelId` 仅渲染到 AI 透明度卡
- generated client `getFeedbackReport` 调用次数为 0；`createPracticeVoiceTurn` 只允许在 voice owner hook 中出现，不能进入 text event hooks / complete handoff
- `appendSessionEvent` 请求 init 含 `Idempotency-Key` header → 0 命中（仅在 completePracticeSession 命中）

#### 5.7 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.044 / E2E.P0.045 / E2E.P0.046 / E2E.P0.047` 全部 `setup → trigger → verify → cleanup` PASS；workspace `E2E.P0.018-021` regression 全部 PASS；backend-practice `E2E.P0.022-026` 与 `E2E.P0.038-043` 真实 regression 全部 PASS。

### Phase 6: D-20 简历扁平化 resumeId 透传

> product-scope D-20 / spec D-15。依赖 B2 004 Phase 7（contract collapse）+ generated client 重生。practice session context / handoff payload 的 `resumeVersionId`→`resumeId`（透传 D-Z InterviewContext 的扁平 resume 绑定）。详见 spec D-15。

#### 6.1 实施

practice session context / handoff payload 的 `resumeVersionId`→`resumeId`（透传 D-Z InterviewContext 的扁平 resume 绑定）。详见 spec D-15。

（验证：vitest 组件/adapter/route + pixel parity + typecheck + build PASS）

#### 6.2 收口

零版本树残留 grep（`resumeVersionId` / `resumeAssetId` / `listResumeVersions` / 版本树组件，generated adapter 除外）+ `sync-doc-index --check`。

（验证：全 gate PASS + 负向 grep 0 命中）

## 5 验收标准

- 本计划列出的 Phase 1-5 全部 checklist 项通过
- spec C-4（practice 文本 happy path 主要部分）覆盖完整且通过对应测试；C-8 / C-9 / C-10 / C-12 practice 子集（含 BDD 主流程 + 关键分支）通过；C-6 仅覆盖 generating 入口跳转 + handoff 参数完整性，generating 屏渲染 + getFeedbackReport 轮询由 report/generating owner 承接；C-5（voice surface）由 `practice-voice-mvp/001` 承接，本 plan 只验证 text-loop 与 voice owner co-location 不互相污染
- 关联 BDD-Gate（`E2E.P0.044 / 045 / 046 / 047`）全部通过；workspace regression（`P0.018-021`）全部 PASS；backend-practice regression（`P0.022-026` + `P0.038-043`）全部 PASS，其中 `P0.038-043` 必须跑当前真实 Go HTTP scenario
- pixel parity 在 desktop + mobile 两 viewport 下 practice 主屏 + voice surface + session-lost + completing + completed + HintBanner 新增 spec 全 PASS
- `make docs-check` zero drift；`check-md-links` OK；`pnpm typecheck` 0 错；`pnpm build` + `make build` PASS
- 负向搜索（ui-design prototype 直接 import / 旧 testid / 旧 route alias / `practiceMode='debrief'` / `practiceGoal='debrief'` / `切到语音` / `getFeedbackReport` 调用 / `createPracticeVoiceTurn` 非 voice-owner-hook 调用 / `Idempotency-Key.*appendSessionEvent` / raw answer/question/hint/provenance 泄漏到 URL/localStorage/console）全部 0 命中

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 当前真实 backend-practice/002 已落地，但前端 plan 仍有 fixture-only UI 分支（尤其 assisted `show_hint`）可能被误读为真实 backend 闭环 | Operation matrix 必须把真实 handler 与 fixture-only 分支分开；Phase 5 同时跑 mock fixture、workspace regression、backend-practice `E2E.P0.022-026` 与 `E2E.P0.038-043`；assisted hint 200 正向在本 plan 标记为 backend-practice/003 前置，当前真实 backend 002 只要求 409 防御分支 |
| `clientEventId` 在 retry 路径下复用策略复杂（同一 user action 失败 retry 复用，新 action 必须 fresh） | hook 内部 `inFlightRef`（per user-action token）缓存当前 batch；retry 调同一 batch；用户主动重新输入 / 切换 turn / 点新 action 时清 ref 并生成新 batch；Vitest 锁定行为 |
| `completePracticeSession` 在 StrictMode 双触发或用户多次点击导致重复 nav | hook 内 `inFlightRef` + `Promise` 缓存当前 in-flight；StrictMode 重复调用复用 promise；nav 仅在 first response 成功后触发一次；Vitest 在 StrictMode 下断言 generated client `completePracticeSession` 调用次数 = 1，nav 调用次数 = 1 |
| voice owner 与 text-loop owner 共用 `PracticeScreen`，旧 text-loop negative grep 误判 `createPracticeVoiceTurn` 为漂移 | P0.044/P0.047 verify 只禁止 `getFeedbackReport` 进入 practice runtime，并限制 `createPracticeVoiceTurn` 只能出现在 `hooks/usePracticeVoiceTurn.ts`；real-mode gate 负责证明 voice operation 已走 generated client |
| `practiceMode` 切换运行时不可改导致用户误操作（顶部 toggle 看起来可改但只是 UI-only） | 顶部 strict switch 在本 plan 保持 ui-design 视觉和 a11y，但点击只显示 toast「严格模拟需在新建规划时设定，本场已锁定」；不调 backend、不写 console；Vitest 锁定 toast 文案和 generated client 调用次数 0；后续如需运行时切换必须先回 B2/backend-practice 修订 contract |
| `getPracticeSession` 在 fixture-backed transport 下不返回真实 turn history → SessionMap 只能展示当前 turn + 推进过的本地 history | SessionMap 用客户端 cache + `assistantAction.type` 推进；Vitest 锁定 cache 行为；如果 fixture 后续添加 turn history view-model，hook 可选择 server-side hydrate（保留 forward-compat 注释） |
| `appendSessionEvent` 409 mismatch 触发 `getPracticeSession` refresh 期间用户继续输入 → race condition | 收到 409 后立即设置 `session.status='refreshing'` UI 锁定（输入禁用 + spinner）；refresh 完成后才解锁；Vitest 锁定 race |
| Pixel parity 跨字体子像素差异（D3 retrospective 经验） | 沿用 D3：`practice.spec.ts` toHaveScreenshot 仅作 frontend 内部 regression（含 maxDiffPixels 阈值），不与 ui-design golden 跨字体源做硬 diff |
| 旧 prototype data 渗透（开发者从 `screen-practice.jsx` 复制粘贴时把 `D.questions` / `D.sessionTranscript` / `D.targetJobs` 一并带过来） | Vitest negative grep + `eslint-rules` 反查（`no-restricted-imports` 限制 `ui-design/`）；scenario verify 阶段 grep `window.EI_DATA` / `D.questions` / `D.sessionTranscript` literal |
| backend-practice/002 落地 D-32 / D-33 后 OpenAPI fixtures parity 漂移 | Phase 5.4 在 `make validate-fixtures` 后再跑 `python3 scripts/lint/conventions_drift.py` + `make codegen-check`；漂移由 PR 描述显式列出 |
| InterviewContext 新增 `INCREMENT_HINT_COUNT` reducer action 可能破坏 001 测试 | 在 001 `interview-context/InterviewContext.test.tsx` 文件追加测试；reducer exhaustive switch 编译保证；Vitest 锁定 001 已有 actions 行为不变 |
| `Prefer: example=<scenario>` fixture variant 切换在 jsdom 与 Playwright 之间行为不一致 | Vitest（jsdom）通过 `EI_FIXTURE_SCENARIO_DEFAULT_OVERRIDE` 环境变量；Playwright 通过 init script 注入 same env；scenario setup.sh 同源生成两边的 env 文件；保证 single source |
