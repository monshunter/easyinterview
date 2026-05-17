# Frontend Debrief Spec

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-05-17

## 1 背景与目标

`frontend-debrief` 是 [engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) `Debrief` workstream 的前端业务 subspec，承接 [frontend-shell](../frontend-shell/spec.md) 已交付的 App 壳、TopBar 一级导航 `debrief` 入口、route normalization、`requestAuth(pendingAction)`、fixture-backed generated client、UI parity gate；承接 [backend-practice](../backend-practice/spec.md) 的 `createPracticePlan(goal='debrief') + startPracticeSession(mode IN ('assisted','strict'))` 真实 session 创建契约（"复盘面试"启动入口），并把新 session 交给 [frontend-workspace-and-practice](../frontend-workspace-and-practice/spec.md) 的 `PracticeScreen` 消费；同时作为 [backend-debrief](../backend-debrief/spec.md) `createDebrief` / `getDebrief` / `suggestDebriefQuestions` schema 的前端 consumer。

本 subspec 的终稿范围收敛为一条正式 owner 路由 + 一个历史 prototype alias：

- `debrief`：复盘主流程默认入口。源级复刻 `ui-design/src/screens-p1-depth.jsx::DebriefFullScreen`（lines 38-2180）+ 5 个子组件（`DebriefContextStrip` / `DebriefContextPickerModal` / `GuidedDebriefRecord` / `VoiceDebriefRecord` / `DebriefReplayPlan`）。三步骤 stepper（`step=0` 记录 / `step=1` 分析 / `step=2` 复盘面试 launcher）+ 三个 in-page picker modal（JD / Mock Session / Resume）+ 文本模式（功能完整：AI 推荐问题 → occurred/skipped/edit/manual → 写入 entries，且每条 entry 必须采集非空回答摘要）+ 语音模式（UI shell only，无真实 STT）+ 跨模式共享 entries 列表 + createDebrief 提交 + polling getJob + polling getDebrief + 分析渲染 + 创建复盘 practice plan/session 后 nav practice。
- `debrief_full`：历史 prototype alias。UI 真理源 `app.jsx` 仍把 `debrief_full` 映射到 `<DebriefFullScreen>`，但正式前端 route catalog 只保留 `debrief`；本 spec 仅要求 `normalizeRouteName("debrief_full") -> "debrief"`，不得把 `debrief_full` 加回正式 `RouteName`、TopBar entry 或 live navigation。

`workspace` / `practice` / `report` / `generating` / `company_intel` / `home` / `parse` 不在本 subspec 范围。本 subspec 在 step 2 "开始复盘面试" CTA 中先调用 backend `createPracticePlan(goal='debrief', sourceDebriefId)` 与 `startPracticeSession` 创建新的可用复盘练习 session，再把 `planId/sessionId/practiceGoal='debrief'` route payload 交给 frontend-workspace-and-practice owner 的 `PracticeScreen` 处理。

本 subspec 通过 generated client + fixture-backed transport 消费 backend-debrief 已声明的 3 个 OpenAPI operation；任何新增或缺失 operation 先回到 [B2](../openapi-v1-contract/spec.md) / [backend-debrief](../backend-debrief/spec.md) 修订，不能在前端手写 ad hoc fetch 或复制 `ui-design` mock data。

## 2 范围

### 2.1 In Scope

- `debrief` 屏（正式 route=`debrief`；宽松入口 `debrief_full` 先经 route normalization 归一为 `debrief`；保留默认 App chrome / TopBar；TopBar 一级导航 `debrief` 入口高亮）：
  - 源级复刻 `ui-design/src/screens-p1-depth.jsx::DebriefFullScreen`（lines 38-2180）：
    - **Header**（lines 122-144）：返回按钮 + eyebrow label（"复盘 · 公司 · 轮次 · 4/22"）+ H1 标题 + 副文案 + 右上 meta 区（time / interviewer / modality）；公司、轮次、时间、interviewer 来自 InterviewContext / debrief 上下文派生
    - **DebriefContextStrip**（lines 412-432）：三个上下文卡片（目标岗位 / JD / 关联模拟面试 / 绑定简历）；点击触发对应 picker modal（不离开 debrief 页）
    - **Stepper**（lines 148-156）：3 步骤 `复盘记录 / 复盘分析 / 复盘面试`；当前 step state 控制下方 panel 渲染
    - **Step 0 复盘记录**（lines 158-318）：
      - 顶部统一汇总条（lines 162-185）：total entries 计数 + recorded/text/voice/manual source chips + 跨模式共享提示
      - 添加方式 toggle（lines 187-210）：`Text` / `Voice` 双 tab；切换不丢 entries 状态
      - 文本模式（`<GuidedDebriefRecord>` lines 519-619）：AI 推荐问题面板（左侧 currentGuide 显示 + 上/下导航）+ 「遇到过 / 没问到 / 改成真实问题」3 个 CTA + 右侧问题卡片列表（共享 entries）+ 点击展开详细回答 / 追问 / 面试官反应标签 / 整体感受 / 遗漏点
      - 语音模式（`<VoiceDebriefRecord>` lines 656-870 UI shell only）：录音状态 idle/listening 占位 + 实时观察流（fade-in mock）+ 待确认卡片列表（从 entries 派生）+ 「确认 / 编辑 / 删除」CTA + 「空格暂停/继续」键盘提示（UI only，不绑定真实音频 API）；用户切到 voice 模式后等待功能集成的"功能即将上线"占位提示（**D-6 P0 边界**）
      - 底部 CTA：`<Btn variant="accent" iconRight="arrow_right" onClick={() => setStep(1)}>生成复盘分析</Btn>`（lines 314-316）→ 触发 createDebrief
    - **Step 1 复盘分析**（lines 320-362）：
      - 分析维度卡片：与关联模拟面试对比（题目重合度 + 模拟预测但未出现 + 真实出现但未预测）+ 与目标 JD 对比（考察重点 + 暴露的岗位风险）+ 与绑定简历对比（简历有但回答没讲清 + 简历缺失但被追问）
      - 风险项列表（来自 `risk_items` jsonb）：每项含 label + severity（low/medium/high color tag）
      - "下一步重点准备" suggestions（来自 `risk_items` + 派生）
      - 底部 CTA：`<Btn iconRight="arrow_right" onClick={() => setStep(2)}>生成复盘面试</Btn>`（line 361）
    - **Step 2 复盘面试 launcher**（`<DebriefReplayPlan>` lines 1388-1421）：
      - 复盘面试 plan 预览：复现真实问题 + 薄弱处追问 + 真实顺序 + 简历证据对比
      - CTA：`<Btn variant="accent" icon="play" onClick={...start replay...}>开始复盘面试</Btn>`（line 1414） → 调用 `createPracticePlan(goal='debrief', sourceDebriefId=debriefId)` + `startPracticeSession`，再 nav practice with payload `{practiceGoal:'debrief', mode:'text', modality:'text', planId, sessionId, targetJobId, resumeVersionId, debriefId}`；未登录走 `useRequestAuth({type:'start_debrief_interview', route:'debrief', params:{...}})`
  - 3 个 in-page picker modal（`<DebriefContextPickerModal>` lines 434-518）：
    - JD picker：列出用户已有 `target_jobs`（generated `listTargetJobs`），单选确认；不离开 debrief 页
    - Mock Session picker：列出用户已有 `practice_sessions` filtered by selected targetJobId + status='completed'（Phase 0 必须由 backend-practice/B2 addendum expose generated `listPracticeSessions({targetJobId, status:'completed'})`；未生成前不得进入 Phase 1；若 operation 已存在但不支持 status filter，前端 client-side filter）；单选确认或 "暂不关联"
    - Resume picker：先列出用户简历资产（generated `listResumes`），用户选择 asset 后再调用 generated `listResumeVersions(resumeAssetId)` 展开 ready versions，单选版本确认
  - 文本模式 AI 推荐问题流程：
    - 入口：用户点击 step 0 文本模式中 "生成推荐问题" 按钮（或自动在 ContextStrip 三选完成后触发）
    - 行为：调用 generated `suggestDebriefQuestions({targetJobId, sessionId?, resumeVersionId?, language, count:6})` 同步 API
    - 渲染：拿到 `suggestions[]` → 渲染 `<GuidedDebriefRecord>` 中 `currentGuide` 列表（替代 prototype hardcoded `guideQuestions`）
    - 用户交互：「遇到过」→ 打开 inline editor，保存真实回答摘要后写入 entries 一行 source='ai_confirmed'；「没问到」→ 跳过当前 guide；「改成真实问题」→ 编辑真实问题 + 回答摘要后写入 entries source='ai_edited'；「手动添加」→ 写入 entries source='manual'；保存 entry 时 `questionText` 与 `myAnswerSummary` 均不能为空
    - 失败处理：AI 失败时显示 inline error + "稍后重试" CTA + 用户可降级到完全手工录入（不阻塞 step 0）
    - 推荐次数：用户可重复触发 "重新生成推荐"；UX 流程上不做 rate limit（依赖 backend-debrief Q-5 决策）
  - createDebrief 提交流程：
    - 入口：step 0 底部 "生成复盘分析" CTA
    - 行为：调用 generated `createDebrief({targetJobId, roundType, interviewerRole?, language, questions: entries.map(toDebriefQuestionInput), notes?})` with `Idempotency-Key` （前端生成 UUIDv4）；`questions[*].myAnswerSummary` 必须来自 UI 采集的非空回答摘要；require auth, 未登录走 `useRequestAuth({type:'submit_debrief', route:'debrief', params:{...}})`
    - 响应：202 + `DebriefWithJob{debriefId, job}`；前端 store `debriefId` + `debriefJobId=job.id` in InterviewContext；不得写入现有 `jobId` 字段（正式前端中该字段是 target job alias/fallback）
    - 失败处理：422 VALIDATION_FAILED → inline error 列出失败字段；409 IDEMPOTENCY_KEY_MISMATCH → 自动重生 IK 重试；401 → useRequestAuth；5xx → toast + retry CTA
  - polling 流程：
    - 提交成功后自动 setStep(1) 并启动 polling
    - 双轨 polling：(a) `getJob(debriefJobId)` 指数退避（初始 1.5s × 1.5 上限 8s），max attempts=30；(b) job.status='succeeded' 时停止 getJob 轮询，启动 `getDebrief(debriefId)` 一次性拉取
    - visibility/focus 暂停-恢复 polling
    - job.status='failed' → 显示 DebriefFailureState + errorCode 文案 + retry CTA / 「返回 step 0 编辑」
    - max attempts 达到 → 显示 timeout state + retry CTA
  - 分析渲染（step 1）：
    - 从 `Debrief.questions[*].aiAnalysis` 派生题目深度反馈卡片
    - 从 `Debrief.riskItems` 渲染风险项列表 + severity color tag
    - 不渲染 `nextRoundChecklist`（P0 留空，D-7 backend-debrief 不填充）
    - 不渲染 `thankYouDraft`（P0 留空）
    - `provenance` 6 字段：仅在「关于本次分析」展开区显示（与 frontend-report-dashboard 一致）
  - 复盘面试 handoff（step 2）：
    - CTA 首先调用 `createPracticePlan({goal:'debrief', sourceDebriefId: debriefId, targetJobId, resumeAssetId, mode, language, ...})`，随后调用 `startPracticeSession({planId, hintsEnabled})`
    - nav payload：`{practiceGoal:'debrief', mode:'text', modality:'text', planId, sessionId:newSessionId, targetJobId, resumeVersionId, debriefId, language}`
    - 本 spec **不**实现 practice 任何 UI；PracticeScreen 继续由 frontend-workspace-and-practice owner 消费新 session
    - **D-11 修订决策**：frontend-debrief 主导复盘 CTA 的 plan/session 创建；不得把已完成的 optional mock session id 当作新的 practice `sessionId` 转发，也不得在无 sessionId 时直接 nav practice
- 跨路由共享：
  - `InterviewContext` 在 debrief owner route 内传递 `{targetJobId, jdId, sessionId?, resumeVersionId, roundId?, debriefId?, debriefJobId?, mode:'text', modality:'text', practiceMode?, practiceGoal?, language}`；本 spec 在 step 0 ContextStrip 三选完成后写入 `targetJobId, sessionId?, resumeVersionId, language`；createDebrief 成功后写入 `debriefId` + `debriefJobId`；step 2 nav practice 时 forward stable IDs + `practiceGoal`，但不 forward raw entries / notes
  - 未登录用户点击 `submit_debrief` / `start_debrief_interview` 复盘流程 CTA 通过 `useRequestAuth({type:..., route:'debrief', params:{...InterviewContext}})` 触发鉴权；登录后 `pendingAction` 回到 `debrief`，由 DebriefScreen 自动检测并恢复
  - 隐私：route params 仅传 stable IDs + display knobs；不传 raw `entries[].q` / `entries[].a` / `entries[].follow` / `notes` / risk_items prose；console.log / localStorage / telemetry 同款约束
- 契约消费形态：
  - `createDebrief`：step 0 → step 1 transition；按 OpenAPI POST `/debriefs` + Idempotency-Key + 完整 questions[] body；返回 `DebriefWithJob`
  - `getJob`：step 1 polling (job lifecycle)；按 OpenAPI GET `/jobs/{jobId}`，本屏使用本地变量 `debriefJobId`
  - `getDebrief`：step 1 polling (debrief enriched data)；按 OpenAPI GET `/debriefs/{debriefId}`
  - `suggestDebriefQuestions`：step 0 文本模式 "生成推荐问题"；按 OpenAPI POST `/debriefs/question-suggestions`（Phase 0 新增）
  - 3 个 picker modal 列表来源：JD picker 使用既有 `listTargetJobs`；Resume picker 使用既有 `listResumes` + `listResumeVersions(resumeAssetId)`；Mock Session picker 使用 Phase 0 backend-practice/B2 addendum 提供的 `listPracticeSessions`

### 2.2 Out of Scope

- `WorkspaceScreen` / Interview Launcher / Resume Picker（workspace-内 picker）：由 [frontend-workspace-and-practice](../frontend-workspace-and-practice/spec.md) 承接。
- `PracticeScreen` 任何 UI / 状态机消费 / 文本 surface / voice surface / 完成动作：由 [frontend-workspace-and-practice](../frontend-workspace-and-practice/spec.md) 承接；本 spec 只在 step 2 CTA 中创建复盘 practice plan/session 并 nav 回 practice。
- `ReportDashboard` / `GeneratingScreen`：由 [frontend-report-dashboard](../frontend-report-dashboard/spec.md) 承接。
- `CompanyIntelScreen` / `getCompanyIntel`：external company-intel owner 承接。
- `HomeScreen` / `ParseScreen` / `JDMatchScreen` shell：由 [frontend-home-job-picks-and-parse](../frontend-home-job-picks-and-parse/spec.md) 承接。
- `ResumeWorkshopScreen`：由 [frontend-resume-workshop](../frontend-resume-workshop/spec.md) 承接。
- Auth / TopBar / Sidebar / Theme / I18n bootstrap / requestAuth 接线：由 [frontend-shell](../frontend-shell/spec.md) 承接。
- Debriefs 真实 backend handler / service / store / event 发射：由 [backend-debrief](../backend-debrief/spec.md) 承接。
- AI provider / prompt registry / 模型路由：由 [ai-provider-and-model-routing](../ai-provider-and-model-routing/spec.md) / [prompt-rubric-registry](../prompt-rubric-registry/spec.md) / backend-debrief 承接。
- **Voice 真实集成（STT + 音频录制 + 实时转写 + 卡片提取）：D-6 决策固定 P0 仅 UI shell；功能集成由未来 plan 协同 [practice-voice-mvp](../practice-voice-mvp/spec.md) 整体语音上线后接入；本 spec **不**实现 Web Audio API 录制、不接入 SpeechRecognition、不调用 STT endpoint（backend 也没有 debrief STT endpoint）。**
- 复盘历史浏览（list debriefs UI）：当前 OpenAPI 无 `listDebriefs` operation；本 spec P0 不实现；plan 002 与 backend-debrief 002 协同决定是否开启。
- 复盘记录原地修订（update existing debrief）：当前 OpenAPI 无 `updateDebrief` operation；本 spec P0 不实现；用户如需修订，体验上引导重新创建。
- 复盘导出 / 分享：由 future 隐私 spec / 平台 owner 承接。
- 复盘评分质量反馈（用户对复盘分析打分）：由 future [prompt-rubric-registry/003-grayscale-and-quality-feedback](../prompt-rubric-registry/spec.md) 承接。
- 不新增或恢复弃用模块 / 路由 / 术语作为 live UI：独立 `mistakes` route / `drill_builder` / `growth_center` / `experience_library` / `star_editor` / 独立 `voice` route alias / 旧 "把复盘终点定义为下一轮面试" 流程口径 / 旧错题本回收箱口径。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | Route owner 范围 | 本 subspec 只接管正式 `debrief` route；`debrief_full` 仅作为历史 prototype alias 在 route normalization 中归一到 `debrief`，不得加入正式 `RouteName` / TopBar / live nav；`workspace / practice / report / generating / company_intel / home / parse / resume` 是外部 owner | 消除与 frontend-workspace-and-practice / frontend-report-dashboard / external owner 的边界冲突 |
| D-2 | UI 真理源 | `ui-design/src/screens-p1-depth.jsx::DebriefFullScreen`（lines 38-2180）+ 5 个子组件 `DebriefContextStrip` / `DebriefContextPickerModal` / `GuidedDebriefRecord` / `VoiceDebriefRecord` / `DebriefReplayPlan` + `ui-design/src/primitives.jsx` + `ui-design/src/app.jsx`（route mapping / INTERVIEW_CONTEXT_ROUTES / TopBar entry "debrief"）+ `docs/ui-design/review-module.md` v2.5 + `docs/ui-design/module-map.md` 为唯一真理源进行源级复刻；不得二次设计 | 保护 ui-design parity gate；不引入外部审美 |
| D-3 | 三步骤 stepper 状态机 | `step ∈ {0,1,2}`：step=0 复盘记录（默认）/ step=1 复盘分析 / step=2 复盘面试 launcher；step 单向递增（0→1→2 通过 CTA 推进）；step=1 时如 polling 失败可"返回 step 0 编辑"；step=2 后退回 step=1 通过浏览器 history 或显式 back CTA；step 状态不持久化到 URL（保持本地，刷新回 step 0；如未来需要 deep-link，plan 002 再加 URL param） | 与 prototype 状态机 line 39 完全一致；简化 P0 实现 |
| D-4 | 文本模式 AI 推荐问题契约 | 调用 generated `suggestDebriefQuestions({targetJobId, sessionId?, resumeVersionId?, language, count:6})`；不在前端硬编码 prompt；不通过其他 endpoint 派生；AI 失败时 inline error + 降级到手工录入；前端不缓存 suggestions（用户主动重新生成） | 与 backend-debrief D-6 一致；P0 不引入推荐 cache 复杂度 |
| D-5 | suggestDebriefQuestions 触发时机 | 在 ContextStrip 三个上下文（targetJob / mockSession / resume）任意更换或确认后，**自动**触发一次 suggestDebriefQuestions（debounce 500ms）；首次进入 debrief 时如 InterviewContext 已含 targetJobId 也自动触发；用户可手动点击 "重新生成推荐" 按钮再次触发；不限制次数 | 自动触发减少用户操作成本；与 prototype "AI 推荐问题用于降低用户回忆成本" 一致 |
| D-6 | Voice 模式 P0 范围 | step 0 voice mode toggle 渲染完整 UI shell（toggle + idle/listening 视觉占位 + 待确认卡片列表 + 共享 entries + 「空格暂停/继续」键盘提示文本）；**不**绑定真实 Web Audio API / SpeechRecognition / WebRTC；**不**调用任何 STT endpoint（backend-debrief 也未提供）；显示固定占位提示 "语音复盘集成中，敬请期待"；功能集成由未来 plan 与 practice-voice-mvp 协调后接入 | UI source parity gate 仍 100% 复刻 voice 模式视觉；与 practice-voice-mvp §2.2 一致；与 user Q-2 选择一致 |
| D-7 | 跨模式共享 entries 状态 | text 与 voice 模式共享同一个 `entries` 数组（React state）；切换 mode 不清空 entries；文本模式的 occurred/edit/manual 与语音模式的 confirm/edit/delete 都写入同一份 list；每条 entry 含 `source: 'ai_confirmed' | 'ai_edited' | 'manual' | 'voice_extracted'` 来源标记；entries 在 createDebrief 提交时转换为 `DebriefQuestionInput[]` | 与 prototype line 78 + review-module.md §5.2 "shared question cards" 一致 |
| D-8 | Context picker 范围 | 3 个 in-page modal（JD / MockSession / Resume）；所有 modal 在当前 debrief 页打开（不跳转）；JD picker 调 generated `listTargetJobs(user_id)` filtered by `analysisStatus='ready'`；Mock Session picker 依赖 backend-practice 真实 `listPracticeSessions({targetJobId,status:'completed'})` handler；Resume picker 调 generated `listResumes()` 列资产，再对选中 asset 调 `listResumeVersions(resumeAssetId)` 展开 ready versions；用户可"暂不关联模拟面试"（Mock 是 optional） | 与 review-module.md §5.1 "三个上下文动作都不得跳出复盘页" 一致 |
| D-9 | createDebrief 提交 payload | `{targetJobId(必填), roundType(必填，B1 DebriefRoundType enum), interviewerRole?(可选，B1 InterviewerRole enum), language(必填，来自 InterviewContext.language 或 i18n 当前 locale), questions: entries.map(toDebriefQuestionInput)(必填，至少 1 条), notes?(可选)}` + Idempotency-Key（前端 UUIDv4）；toDebriefQuestionInput 映射：`{questionText: entry.q, myAnswerSummary: entry.a, interviewerReaction: entry.follow + entry.reflection (concat)}`；UI 必须在保存 entry 前采集非空 `myAnswerSummary`，Submit CTA 在存在空回答摘要时 disabled；不传 entry.tag / entry.source / entry.id（前端 only） | 与 backend-debrief D-1 + OpenAPI CreateDebriefRequest schema 一致 |
| D-10 | polling 节奏 | 双轨 polling：(a) `getJob(debriefJobId)` 指数退避（初始 1.5s × 1.5 上限 8s，max attempts=30 约 4 分钟）+ visibility/focus 暂停-恢复；(b) job.status='succeeded' 触发 `getDebrief(debriefId)` 一次性拉取（不持续 polling getDebrief）；job.status='failed' 触发 DebriefFailureState；max attempts 达到触发 timeout state；`debriefJobId` 不写入既有 `jobId` 字段 | 与 frontend-report-dashboard D-3 一致原则；适配 debrief 双 endpoint 模型，避免污染 target job context |
| D-11 | 复盘面试 handoff 边界 | frontend-debrief step 2 CTA 在已登录状态下调用 `createPracticePlan(goal='debrief', sourceDebriefId=debriefId)` + `startPracticeSession` 创建 fresh practice session，然后 `nav("practice", {practiceGoal:'debrief', mode:'text', modality:'text', planId, sessionId, targetJobId, resumeVersionId, debriefId, language})`；未登录走 `useRequestAuth({type:'start_debrief_interview', route:'debrief', params:{...}})`，登录恢复后回 debrief 重放 CTA；本 spec 不实现 practice 任何 UI | 当前 `PracticeScreen` 对缺失 `sessionId` 会进入 session-lost，且 optional mock session 是已完成历史会话，不能作为 replay practice session 复用；因此 fresh plan/session 创建必须在 debrief CTA 内完成 |
| D-12 | 失败状态语义 | `DebriefFailureState`（job.status='failed'）：失败卡片 + errorCode 文案映射（按 B1 `AI_*` enum 各自文案）+ CTA「返回 step 0 编辑」（保留 entries）/「重试生成」（resubmit createDebrief）；`DebriefMissingContextState`（缺 targetJobId）：卡片 + CTA「选择目标岗位」（自动 open JD picker）；`DebriefTimeoutState`（polling max attempts）：卡片 + CTA「重试」+「返回 step 0」；不暴露 raw provider error 给用户 | 与 backend-debrief D-9 graceful failed + B1 error_code 一致 |
| D-13 | i18n 命名空间约定 | 新增 `debrief.*` 命名空间；不复用 `workspace.*` / `practice.*` / `report.*`；外部 `home.gotoDebrief` 等已存在的 key 保留不动（由相应 owner 维护） | 命名空间独立避免与其他 owner 冲突 |
| D-14 | InterviewContext reducer 扩展边界 | 在 frontend-workspace-and-practice 已有 `InterviewContext` reducer 基础上**仅 read** + **新增 1 个 reducer action** `SET_DEBRIEF_CONTEXT`（写 `debriefId` + `debriefJobId` + `targetJobId` + `sessionId?` + `resumeVersionId` + `practiceGoal?` + `language`）；该 action 在本 spec 内由 ContextStrip 三选完成 + createDebrief 成功后触发；不得写现有 `jobId` 字段；同步扩展 `PENDING_ACTION_INTERVIEW_KEYS` 覆盖 `practiceGoal` / `debriefId` / `debriefJobId` 并补 round-trip 测试 | 不破坏 frontend-workspace-and-practice reducer 边界；只增量加 1 个 action；避免 `jobId` alias 与 debrief async job id 冲突 |
| D-15 | DOM 锚点 / testid 命名 | `debrief-*` 前缀：`debrief-screen` / `debrief-back` / `debrief-header` / `debrief-context-strip` / `debrief-context-card-{targetJob,mockSession,resume}` / `debrief-stepper-step-{0,1,2}` / `debrief-mode-toggle-{text,voice}` / `debrief-suggested-question-{i}` / `debrief-occur-btn` / `debrief-skip-btn` / `debrief-edit-btn` / `debrief-manual-add-btn` / `debrief-entry-card-{id}` / `debrief-voice-status` / `debrief-voice-pending-card-{id}` / `debrief-submit-btn` / `debrief-loading-state` / `debrief-failure-state` / `debrief-missing-context-state` / `debrief-timeout-state` / `debrief-analysis-risk-item-{i}` / `debrief-analysis-dimension-{mock,jd,resume}` / `debrief-interview-launcher` / `debrief-start-interview-btn` / 3 个 modal `debrief-picker-modal-{type}`；workspace / practice / report 前缀归外部 owner | DOM anchor 锁定让源级 parity test 可执行 |
| D-16 | 隐私红线 | route params / URL search params / InterviewContext / sessionStorage / localStorage / console.log / telemetry payload 不传 raw `entries[].q` / `entries[].a` / `entries[].follow` / `entries[].reflection` / `notes` / 任何 risk_items prose / AI prompt body / AI response body / model_id raw value；只允许 stable owner IDs + display knobs + 数量 / 状态 / error_code；createDebrief request body 直接发送（必要传输）但响应后不持久化到本地；fixture transport spy 不泄漏 body；getDebrief 响应数据存于 React state 不写 localStorage | 与 frontend-workspace-and-practice plan 002 / product-scope §9.3 / backend-debrief D-12 一致 |
| D-17 | backend 契约消费 | 只通过 [B2 generated client](../openapi-v1-contract/spec.md) 消费 OpenAPI operation；字段变化先回 B2 / backend-debrief / backend-practice 修订；不在前端自造 endpoint 或复制 fixture JSON；本 spec 在 plan 001 Phase 0 验证 backend-debrief Phase 0 cross-owner addendum 已落地（B1 enum / B2 operation / fixtures 已 generated），并验证 backend-practice/B2 addendum 已生成 `listPracticeSessions`；对 `getJob` 等 frontend-consumed async polling operation，收口证据必须同时证明 generated client / fixture、真实 `backend/cmd/api` route mount、handler/store auth scope 与 focused route/store tests，不能把 fixture-backed mock 通过等同于真实 backend 闭环 | 与 frontend-workspace-and-practice D-10 / frontend-report-dashboard D-14 一致；BUG-0070 固化 runtime route gate |
| D-18 | retired 术语 negative scope | 旧 `experience_library` / `star_editor` / `drill_builder` / `mistakes_book` / `growth_center` / `report_timeline` / 独立 `voice` route / 旧 "下一轮面试" 作为复盘终点 / 错题本回收箱 / 单题 Drill / 追问树独立流 不得作为 live route / TopBar 项 / 正向 testid / 正向 scenario / 用户可见入口出现；如出现在源码 / 测试中必须有 `// negative grep target` 显式标注 | 与 product-scope D-6/D-11 + ui-design review-module.md §10 一致 |

### 3.2 待确认事项

- D-3 step state 是否持久化到 URL（深链能力）：plan 001 默认不持久化（刷新回 step 0 + 保留 entries 在 React state）；plan 002 若需深链 `?step=1&debriefId=D` 再考虑（前提是 polling state 也要可恢复）。
- Mock Session picker 所需 `listPracticeSessions({targetJobId,status:'completed'})` 已要求 backend-practice 提供真实 handler/service/store 与 generated client；后续变更必须保持 OpenAPI / fixture / generated / backend route / frontend consumer 同步。
- entries 本地持久化（草稿保存）：默认 P0 不持久化（页面离开 entries 丢失）；用户体验上需要在 step 0 离开 / 刷新时显示 "草稿将丢失" 确认对话框；plan 002 评估是否引入 localStorage 草稿。

## 4 设计约束

- 视觉与交互必须以 `ui-design/src/screens-p1-depth.jsx::DebriefFullScreen`（含 5 个子组件）、`ui-design/src/primitives.jsx`、`ui-design/src/app.jsx`（route mapping / INTERVIEW_CONTEXT_ROUTES / TopBar）为唯一真理源进行源级复刻；不得二次设计；不引入外部审美。
- DebriefFullScreen 的 Header / ContextStrip / Stepper / Step 0 (record) / Step 1 (analysis) / Step 2 (interview launcher) 主结构必须与 `screens-p1-depth.jsx` 当前 DOM 一致；primitives（`<Btn>` / `<Icon>` / `<ChipBtn>` / `<DimRow>` / `<StatCard>` 等）直接复用 `primitives.jsx`。
- Voice 模式必须渲染完整 UI shell（per D-6），不允许直接隐藏或缺失 voice toggle / 录音状态占位 / 待确认卡片视觉。
- route context 最小键必须按下表执行：

| Route | 本 spec owner | 最小上下文 | 缺失处理 |
|-------|---------------|------------|----------|
| `debrief` | 是 | 无强制必填（首次进入可触发 ContextStrip 三选）；推荐携带 `targetJobId` 跳过 JD picker；宽松 `debrief_full` 输入必须先 normalize 为 `debrief` | 缺 targetJobId 显示 ContextStrip 三选默认态 + 自动打开 JD picker；entries 始终从空开始 |
| `workspace` | 否 | `targetJobId` | 由 frontend-workspace-and-practice 处理 |
| `practice` | 否 | `practiceGoal='debrief'` + `sessionId` + `planId` + `targetJobId` + `resumeVersionId` + `debriefId` | 由 frontend-workspace-and-practice `PracticeScreen` 处理已创建的新复盘 session；frontend-debrief CTA 负责创建 plan/session |
| `report` | 否 | `sessionId + reportId` | 由 frontend-report-dashboard 处理 |
| `company_intel` | 否 | `targetJobId + jdId` | 由 company-intel owner 处理 |

- 隐私红线（D-16）：raw entries / notes / risk_items prose 不得进入 console.log / URL query / localStorage / sessionStorage / telemetry payload / fixture transport spy；createDebrief request body 直接发送但响应后不持久化；getDebrief 响应仅存 React state；i18n 翻译串可包含 placeholder（如 `{{questionText}}`）但 fixture 测试必须验证 placeholder 在产线不解析为 raw text 暴露。
- 暗色 / customAccent / 主题切换必须在 owner 屏（正式 `debrief`，含 `debrief_full` normalize 后入口）通过 root `data-theme` / `data-mode` / `data-custom-accent` 生效，与 frontend-shell parity gate 一致。
- I18n 必须支持 zh / en；新增 `debrief.*` 命名空间；workspace / practice / report 文案归外部 owner。
- Pixel parity gate 必须在 desktop (1440×900) + mobile (390×844) 两个 viewport 下断言 owner 屏的 DOM 锚点 / computed style / bounding box / 截图差异。
- Mobile 响应式：Header 紧凑布局；ContextStrip 三卡片折叠为单列；Stepper 横向滑动或缩短；Step 0 双栏（guide + entries）折叠为单列 + Tab 切换；Step 1 风险列表 + 维度卡片单列；Step 2 launcher CTA sticky bottom；Picker modal 全屏 sheet。
- `data-testid` 遵循 D-15 命名，使用 `debrief-*` 前缀；workspace / practice / report 前缀归外部 owner。
- stale-contract negative gate（D-18）必须区分"禁止作为 live UI/runtime 正向入口"和"允许出现在负向断言/禁止清单/历史说明"。旧 route / module 名称不得作为 active route / TopBar 项 / 正向 testid / 正向 scenario / 用户可见入口重新出现。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| debrief UI | `frontend-debrief`（本 spec） | DebriefFullScreen React 组件、5 个子组件、3 picker modal、Stepper 状态机、polling hook、状态分支（dashboard/failure/missing/timeout）、复盘面试 fresh practice plan/session 创建 + nav CTA、source parity、visual parity、i18n、a11y、responsive；`debrief_full` 仅作为 normalize alias，不是独立 UI owner |
| Workspace / Practice UI | [`frontend-workspace-and-practice`](../frontend-workspace-and-practice/spec.md) | workspace 屏、practice 屏、PracticeScreen 对 `sessionId/planId/practiceGoal='debrief'` 的消费；本 spec 在复盘面试 CTA 创建 session 后 nav 回该 owner |
| Report / Generating UI | [`frontend-report-dashboard`](../frontend-report-dashboard/spec.md) | report dashboard 与 generating handoff；本 spec 不交互 |
| App shell / routes / auth / runtime / theme | [`frontend-shell`](../frontend-shell/spec.md) | TopBar 一级导航 debrief 入口、route normalization、requestAuth、generated client bootstrap、mock transport、display preferences |
| Home / Parse / JD Match | [`frontend-home-job-picks-and-parse`](../frontend-home-job-picks-and-parse/spec.md) | home 屏；本 spec 不交互 |
| Resume Workshop UI | [`frontend-resume-workshop`](../frontend-resume-workshop/spec.md) | resume 屏；本 spec 不交互 |
| Debriefs backend | [`backend-debrief`](../backend-debrief/spec.md) | `createDebrief` / `getDebrief` / `suggestDebriefQuestions` handler / service / store / drainer-registered worker handler |
| Practice backend | [`backend-practice`](../backend-practice/spec.md) | `listPracticeSessions` picker 列表；`createPracticePlan(goal='debrief')` + `startPracticeSession(mode IN ('assisted','strict'))` handler；本 spec 通过 generated client 在 Step 2 CTA 调用 |
| OpenAPI / fixtures / codegen | [`openapi-v1-contract`](../openapi-v1-contract/spec.md) + [`mock-contract-suite`](../mock-contract-suite/spec.md) | `openapi/openapi.yaml`、fixtures `Debriefs/*.json`（Phase 0 由 backend-debrief/001 扩展既有 create/get fixture，并新增 suggest fixture）、generated Go/TS artifacts、fixture-backed mock transport |
| TargetJob data | [`backend-targetjob`](../backend-targetjob/spec.md) | `target_jobs` 行；本 spec 通过 generated `listTargetJobs` 显示 JD picker |
| Resume data | [`backend-resume`](../backend-resume/spec.md) | 简历 binding 字段；本 spec 通过 generated `listResumes` + `listResumeVersions(resumeAssetId)` 显示 Resume picker |
| Practice session data | [`backend-practice`](../backend-practice/spec.md) | practice_sessions 行；本 spec 通过 Phase 0 backend-practice/B2 addendum 生成的 `listPracticeSessions({targetJobId,status:'completed'})` 显示 Mock picker |
| Voice 真实集成 | future plan + [`practice-voice-mvp`](../practice-voice-mvp/spec.md) | STT / WebRTC / 音频留存 / 隐私链路；本 spec P0 不实现 |

### 5.1 Operation Matrix

| operationId | Fixture | Frontend consumer | Backend handler | Persistence | AI dependency | Scenario / status |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createDebrief` | `openapi/fixtures/Debriefs/createDebrief.json`（既有文件，backend-debrief/001 Phase 0 扩展：`default` = 202 + DebriefWithJob） | DebriefScreen step 0 "生成复盘分析" CTA | backend-debrief 001 Phase 2 真实 handler | `debriefs` write + `async_jobs` write + outbox `debrief.created` | none in frontend；worker AI 调用由 backend-debrief 完成 | `E2E.P0.066` |
| `getDebrief` | `openapi/fixtures/Debriefs/getDebrief.json`（既有文件，backend-debrief/001 Phase 0 扩展：`default` = completed 完整字段 / `debrief-draft` = draft + 空字段 / `prototype-baseline` = 中文示例） | DebriefScreen step 1 polling 拉取数据 | backend-debrief 001 Phase 5 真实 handler | `debriefs` read | none in frontend | `E2E.P0.067` |
| `suggestDebriefQuestions` | `openapi/fixtures/Debriefs/suggestDebriefQuestions.json`（backend-debrief/001 Phase 0 新增：`default` = 6 suggestions / `empty` = 0 / `prototype-baseline`） | DebriefScreen step 0 文本模式自动 / 手动触发 | backend-debrief 001 Phase 3 真实 handler | `ai_task_runs` write + `audit_events` write | F3 `debrief.suggest_questions`（backend） | `E2E.P0.066` happy + AI failure |
| `getJob` | `openapi/fixtures/Jobs/getJob.json`（既有） | DebriefScreen step 1 polling job lifecycle | `backend/internal/api/jobs` handler + `backend/internal/store/jobs` owner-scoped read；`backend/cmd/api/main.go` 必须挂载 `GET /api/v1/jobs/{jobId}` | `async_jobs` read，按 `resource_type` 回查 owner resource（含 `debriefs.user_id`） | none | `E2E.P0.067` + BUG-0070 focused route/store gate |
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json`（既有） | DebriefContextPickerModal JD picker | backend-targetjob 既有 handler | `target_jobs` read | none | `E2E.P0.065` |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json`（既有） | DebriefContextPickerModal Resume picker asset list | backend-resume 既有 handler | `resume_assets` read | none | `E2E.P0.065` |
| `listResumeVersions` | `openapi/fixtures/Resumes/listResumeVersions.json`（既有；调用参数必须是 `resumeAssetId`） | DebriefContextPickerModal Resume picker version list after asset selection | backend-resume 既有 handler | `resume_versions` read | none | `E2E.P0.065` |
| `listPracticeSessions` | `openapi/fixtures/PracticeSessions/listPracticeSessions.json`（支持 targetJobId + status filter） | DebriefContextPickerModal Mock Session picker | backend-practice real handler | `practice_sessions` read | none | `E2E.P0.065` |
| `getTargetJob` / `getResumeVersion` / `getPracticeSession` | 既有 fixtures | DebriefContextStrip 展示 title / display name；失败时 fallback 显示 ID | backend 既有 handler | read | none | `E2E.P0.065` ContextStrip 子断言 |
| `createPracticePlan` / `startPracticeSession` | `openapi/fixtures/PracticePlans/createPracticePlan.json` / `openapi/fixtures/PracticeSessions/startPracticeSession.json` | DebriefScreen Step 2 "开始复盘面试" CTA | backend-practice real handlers | `practice_plans` write + `practice_sessions` write | first-question AI dependency in backend-practice | `E2E.P0.068` fresh session handoff |
| `getFeedbackReport` / `getCompanyIntel` | — | 本 spec **不**调用 | — | — | — | 负向断言 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Owner route 专属 Screen 接管 | frontend-shell D1 已交付；backend-debrief Phase 0 cross-owner addendum 已落地；debrief route 当前由 PlaceholderScreen 占位 | 进入 `debrief` 或宽松初始 route `debrief_full` | `debrief` 渲染正式 DebriefScreen；`debrief_full` 先 normalize 为 `debrief` 后渲染同屏；保留默认 App chrome / TopBar；TopBar 一级导航 `debrief` 入口高亮；不展示 PlaceholderScreen；不新增正式 `debrief_full` RouteName | 001 |
| C-2 | Default render + ContextStrip + Stepper | 用户已认证；InterviewContext 含 / 不含 targetJobId | 进入 `debrief` | 渲染 Header + ContextStrip (3 cards) + Stepper (3 steps, current=0) + Step 0 Record panel（mode toggle + entries 空态 + Submit CTA disabled）；如 InterviewContext 含 targetJobId 自动填充 JD 卡片，否则显示 default 提示 | 001 |
| C-3 | 3 个 in-page picker modal | C-2 已渲染；Phase 0 已生成 `listPracticeSessions` | 点击 ContextStrip 的 targetJob / mockSession / resume 卡片 | 在当前页打开对应 modal；调 generated `listTargetJobs` / `listPracticeSessions` / `listResumes` + `listResumeVersions(resumeAssetId)`；用户选择后关闭 modal；ContextStrip 卡片更新；不离开 debrief 页 | 001 |
| C-4 | 文本模式 AI 推荐问题 | C-3 已选完 3 个上下文；fixture `suggestDebriefQuestions=default` 返回 6 suggestions | ContextStrip 三选完成后 debounce 500ms 触发 | 自动调用 `suggestDebriefQuestions({targetJobId, sessionId?, resumeVersionId?, language, count:6})`；GuidedDebriefRecord 左侧渲染 6 条 suggestions；用户可点击 occurred/skipped/edit/manual；保存 entry 前必须填写非空回答摘要；每次操作写入 entries 一行 with source 标记 | 001 |
| C-5 | AI 推荐失败降级 | fixture `suggestDebriefQuestions=fail` 返回 canonical B1 `AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_PROVIDER_CONFIG_INVALID` 之一 | suggestDebriefQuestions 调用 | 显示 inline error "推荐生成失败，可手动添加问题"；不阻塞 step 0；用户可继续手工添加 entries；可点击 "重新生成推荐" 重试 | 001 |
| C-6 | Voice 模式 UI shell | C-2 已渲染 step 0；text 模式当前激活 | 用户点击 mode toggle 切换到 voice | 渲染 VoiceDebriefRecord UI shell（toggle highlighted + idle 状态视觉 + 待确认卡片列表（空）+ 「空格暂停/继续」键盘提示 + entries 列表保留）；显示 "语音复盘集成中，敬请期待" 占位提示；切换回 text 模式 entries 仍保留；UI 视觉 100% 源级复刻 prototype VoiceDebriefRecord | 001 |
| C-7 | createDebrief submit 主路径 | C-4 已写入 entries (≥1 条) 且每条 entry 有非空回答摘要；fixture `createDebrief=default` 返回 202 | 用户点击 "生成复盘分析" CTA | 调用 `createDebrief({targetJobId, roundType, interviewerRole, language, questions, notes}) + Idempotency-Key=UUIDv4`，其中 `questions[*].myAnswerSummary` 非空；响应 202 + DebriefWithJob{debriefId, job}；InterviewContext 写入 debriefId + debriefJobId；不写现有 jobId；自动 setStep(1) 并启动 polling | 001 |
| C-8 | polling getJob + getDebrief happy | C-7 已成功；fixture `getJob` 配置为 queued→running→succeeded 三次；fixture `getDebrief=default` 完整字段 | step 1 启动 polling | 调用 `getJob(debriefJobId)` 多次（按指数退避节奏）；status='succeeded' 时停止 getJob，调 `getDebrief(debriefId)` 一次；渲染 step 1 分析面板（risk_items + dimensions + provenance 展开区） | 001 |
| C-9 | DebriefFailureState | fixture `getJob=failed` 返回 status='failed' + errorCode='AI_PROVIDER_TIMEOUT' | step 1 polling 命中 failed | 渲染 DebriefFailureState 卡片 + errorCode 文案映射 + CTA「返回 step 0 编辑」（保留 entries）+「重试生成」（resubmit createDebrief with new IK） | 001 |
| C-10 | DebriefTimeoutState | fixture `getJob` 永久返回 status='queued'（模拟 backend 卡住） | step 1 polling 达到 max attempts=30 | 渲染 DebriefTimeoutState 卡片 + CTA「重试」（重启 polling）/「返回 step 0」 | 001 |
| C-11 | DebriefMissingContextState | 用户直接进入 `debrief` 无任何 InterviewContext | 进入 debrief | 渲染 ContextStrip 三卡片 default 态 + 自动打开 JD picker modal；entries 空；step 0 默认；Submit CTA disabled | 001 |
| C-12 | Step 1 分析渲染（completed） | C-8 已完成；getDebrief 返回 completed Debrief with risk_items 非空 | step=1 渲染 | 显示 风险项列表（每项 label + severity color tag）+ 与模拟面试对比维度卡 + 与目标 JD 对比维度卡 + 与绑定简历对比维度卡 + provenance 展开区 6 字段；不渲染 nextRoundChecklist / thankYouDraft（P0 留空） | 001 |
| C-13 | Step 2 复盘面试 launcher + handoff | C-12 已渲染；debriefId + targetJobId + resumeAssetId 可用 | 用户点击 "生成复盘面试" CTA（先 setStep=2），再点 "开始复盘面试" CTA | step 2 渲染 DebriefReplayPlan 内容预览；点击 "开始" 调用 `createPracticePlan(goal='debrief', sourceDebriefId=debriefId)` + `startPracticeSession` 创建 fresh session；随后 `nav("practice", {practiceGoal:'debrief', mode:'text', modality:'text', planId, sessionId:newSessionId, targetJobId, resumeVersionId, debriefId, language})`；未登录走 useRequestAuth 并回 debrief 恢复；不得复用已完成 mock session id | 001 |
| C-14 | UI source structure parity | C-1~C-13 通过 | Vitest+jsdom 加载 DebriefScreen 各 step | DOM 锚点、控件类型、icon、aria、keyboard、menu/modal 层级、5 个子组件嵌套关系可追溯到 `screens-p1-depth.jsx::DebriefFullScreen` 与 `primitives.jsx`；testid 命名按 D-15 一致 | 001 |
| C-15 | UI visual geometry parity | C-14 通过 | Playwright desktop + mobile 加载 owner 屏 | 关键区块不重叠且 stays in viewport；theme/dark/customAccent 可见；Mobile 折叠 + sticky CTA；3 个 picker modal 在 mobile 转为全屏 sheet | 001 |
| C-16 | UI stale-contract negative search | C-14 + C-15 通过 | lint/grep gate 扫描 active runtime、positive tests、README、scenario | 旧 `experience_library` / `star_editor` / `drill_builder` / `mistakes_book` / `growth_center` / `report_timeline` / 独立 `voice` route / 旧 "下一轮面试作为终点" 流程 不作为 live route / TopBar / 正向 testid / 正向 scenario / 用户入口出现；负向断言 / 禁止清单命中被分类允许 | 001 |
| C-17 | Privacy 红线 | 用户完成 step 0 → step 1 → step 2 完整流程 | 检查 URL/localStorage/sessionStorage/console.log/telemetry/fixture transport | raw entries.q/a/follow/reflection/notes / risk_items prose / AI prompt body / model_id raw value 不泄漏；只允许 stable owner IDs + display knobs + 数量 / 状态 / error_code | 001 |
| C-18 | BDD 主流程 + 关键分支 | DebriefScreen + parity gate 已就绪 | 创建并执行 E2E 场景 P0.065-069 | 覆盖 default render + 3 picker、文本模式 AI suggestions + createDebrief happy、polling getJob + getDebrief happy、DebriefFailureState、复盘面试 nav handoff、pixel parity + i18n + privacy + legacy negative | 001 |

## 7 关联计划

本 spec v1.0 已创建首个 active plan 目录 `001-debrief-screen-and-handoff`；其余计划编号仍为预留，后续通过 `/design` 创建对应 plan/context 后再进入 `/implement`：

- `001-debrief-screen-and-handoff` — DebriefScreen 全屏 + 3 picker modal + 3-step stepper + 文本模式 AI suggestions + createDebrief 提交 + 双轨 polling + 分析渲染 + 复盘面试 nav + 失败 / 缺 context / timeout 兜底 + i18n + Playwright pixel parity + 旧口径负向 + BDD `E2E.P0.065-069`。
- `002-debrief-voice-integration-and-history` — Voice 模式真实 STT 集成（依赖 practice-voice-mvp 整体语音上线）+ 复盘历史浏览 UI（如 backend-debrief plan 002 开放 listDebriefs）+ 本地草稿持久化（localStorage entries draft）。
- `003-debrief-export-and-share` — 复盘导出 / 分享（依赖未来隐私 spec）。

## 8 关联文档

- 上游 spec：[`engineering-roadmap`](../engineering-roadmap/spec.md) §5.2、[`product-scope`](../product-scope/spec.md) §6.5（主流程 D）+ §6.11（M4 复盘）+ §4.1（产品原则）、[`frontend-shell`](../frontend-shell/spec.md)、[`frontend-workspace-and-practice`](../frontend-workspace-and-practice/spec.md)、[`backend-debrief`](../backend-debrief/spec.md)、[`backend-practice`](../backend-practice/spec.md)、[`openapi-v1-contract`](../openapi-v1-contract/spec.md)、[`mock-contract-suite`](../mock-contract-suite/spec.md)、[`shared-conventions-codified`](../shared-conventions-codified/spec.md)、[`prompt-rubric-registry`](../prompt-rubric-registry/spec.md)、[`practice-voice-mvp`](../practice-voice-mvp/spec.md)
- UI 真理源：`ui-design/src/screens-p1-depth.jsx::DebriefFullScreen` + 5 个子组件、`ui-design/src/primitives.jsx`、`ui-design/src/app.jsx`（route mapping / INTERVIEW_CONTEXT_ROUTES / TopBar entry "debrief"）、`ui-design/src/data.jsx`、[`docs/ui-design/review-module.md`](../../ui-design/review-module.md)、[`docs/ui-design/module-map.md`](../../ui-design/module-map.md)、[`docs/ui-design/ui-architecture.md`](../../ui-design/ui-architecture.md)、[`docs/ui-design/user-flow.md`](../../ui-design/user-flow.md)
- 当前正式前端入口：`frontend/src/app/{routes.ts,App.tsx,screens/PlaceholderScreen.tsx}`、`frontend/src/api/{generated/client.ts,mockTransport.ts}`、`frontend/src/app/runtime/AppRuntimeProvider.tsx`、`frontend/src/app/auth/pendingAction.ts`、`frontend/src/app/i18n/locales/{zh,en}.ts`、`frontend/src/app/theme/`、`frontend/src/app/interview-context/`、`frontend/tests/pixel-parity/`
- Fixture：`openapi/fixtures/Debriefs/createDebrief.json`、`openapi/fixtures/Debriefs/getDebrief.json`、`openapi/fixtures/Debriefs/suggestDebriefQuestions.json`（均由 backend-debrief/001 Phase 0 新增）
- 治理 / 流程：[`AGENTS.md`](../../../AGENTS.md)、[`docs/development.md`](../../development.md) §2、[`docs/spec/README.md`](../README.md)、[`docs/spec/TEMPLATES.md`](../TEMPLATES.md)、[`test/scenarios/README.md`](../../../test/scenarios/README.md)
- 历史：[history.md](./history.md)
