# 001 Debrief Screen and Handoff

> **版本**: 1.7
> **状态**: active
> **更新日期**: 2026-06-13

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 Test Plan**: [test-plan](./test-plan.md)

## 1 目标

落地 frontend-debrief P0 闭环的全部前端实现：正式 `debrief` route 接管 + 历史 `debrief_full` 输入 normalize 到 `debrief` + DebriefScreen 全屏 + 5 个子组件源级复刻 + 3 个 in-page picker modal + 3-step stepper + 文本模式 AI 推荐问题（suggestDebriefQuestions 自动 + 手动触发）+ 非空回答摘要采集 + 跨模式共享 entries + 语音模式 UI shell（D-6 P0 不实现真实 STT）+ createDebrief 提交 + 双轨 polling（getJob + getDebrief）+ 三种失败态（Failure / Missing / Timeout）+ 分析渲染 + 复盘面试 fresh practice plan/session handoff + i18n / 主题 / 响应式 / pixel parity / 隐私红线 / 旧口径负向 / BDD `E2E.P0.065-069`。

落地完成后，DebriefScreen 可作为 P0 用户路径中"刚面完一轮 → 复盘记录 → 复盘分析 → 复盘面试"闭环的前端入口，与 backend-debrief/001 一起完成 P0 整体闭环最后一段域。

2026-05-23 L2 real-backend gate remediation：P0.065-P0.069 trigger 前置 `frontendOwners.realApiMode.test.ts`，verify 检查 `VITE_EI_API_MODE=real`、默认 backend base URL 与测试文件 marker；fixture-backed UI variants 继续覆盖 DebriefScreen 状态分支，真实 debrief / jobs / picker / replay practice generated-client routing 由集中 gate 证明。

## 2 背景

`frontend-debrief` spec v1.0 在 2026-05-16 由本 plan 同时段派生；spec §7 已声明 plan 001 落地全部 D-1~D-18 决策。spec §1 已确认 UI 真理源为 `ui-design/src/screens-p1-depth.jsx::DebriefFullScreen`（lines 38-2180）+ 5 个子组件 + `primitives.jsx` + `app.jsx` + `docs/ui-design/review-module.md`。

**Phase 0 跨域前置依赖**：
- backend-debrief/001 Phase 0 必须先完成 cross-owner addendums（B1 enum / B2 operation 含 suggestDebriefQuestions / B3 events.yaml 修复 / B4 ai_task_runs task_type / F3 baseline），否则 generated client 中没有 `suggestDebriefQuestions` 方法可调用，fixtures 中没有 `Debriefs/*.json` variants 可消费。
- backend-practice 必须同时具备 generated `listPracticeSessions({targetJobId,status?})` operation + fixture + TS client + 真实 `GET /practice/sessions` handler/service/store；缺失真实 backend route 时不得进入 picker 集成。
- backend-resume 现有契约为 `listResumes()` 列资产 + `listResumeVersions(resumeAssetId)` 列某个资产版本；本 plan 不再假设存在全局按 status 过滤的 resume-version 列表入口。
- frontend-workspace-and-practice 已交付 `InterviewContext` reducer + `useRequestAuth` + nav practice 入口；本 plan 在此基础上增量扩展 1 个 reducer action `SET_DEBRIEF_CONTEXT`。
- frontend-shell 已交付 TopBar 一级导航 `debrief` 入口 + route normalization + i18n / theme / pixel parity infrastructure。
- backend-practice 现状已支持 `goal='debrief'` plan 派生 + 合法 `mode IN ('assisted','strict')` session start（由 backend-debrief/001 Phase 0.6 Q-3 与 backend-practice/004 验证保障）。

本 plan 假设上述依赖在 Phase 0 全部就绪后才进入 Phase 1。

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + code-internal（混合：用户可见 UI + 前端业务状态机 / React Hook / reducer 实现）
- **TDD 策略**: Code plan requires TDD。所有 React 组件 / hook / reducer / fetch logic 必须先写测试（红/绿/重构）；测试文件：`frontend/src/app/screens/debrief/*.test.tsx`、`frontend/src/app/screens/debrief/hooks/*.test.ts`、`frontend/src/app/screens/debrief/reducer.test.ts`；测试命令：`pnpm --filter @easyinterview/frontend test -- src/app/screens/debrief`；Phase 1-7 每个 checklist item 命名其测试断言来源（见 test-plan.md 与 test-checklist.md）。
- **BDD 策略**: Feature plan requires BDD。本 plan 引入用户可见 UI 全屏 + 跨页 nav handoff + 失败态 + AI 推荐 + polling；BDD scenarios `E2E.P0.065-069` 已在 [bdd-plan.md](./bdd-plan.md) 分配，主 [checklist.md](./checklist.md) 在 Phase 8 含 `BDD-Gate:` 项引用每个 scenario ID；执行必须使用当前场景框架的 `scripts/setup.sh` → `scripts/trigger.sh` → `scripts/verify.sh` → `scripts/cleanup.sh` 四段入口（P0.065-P0.068 为 Vitest-backed runner，P0.069 为 Vitest + Playwright debrief parity runner），cleanup 在失败时也必须执行。
- **替代验证 gate**:
  - Phase 0 dep 验证：`grep -rn "suggestDebriefQuestions\|createDebrief\|getDebrief\|listPracticeSessions\|listResumes\|listResumeVersions" frontend/src/api/generated/` 命中；`ls openapi/fixtures/Debriefs/` 含 3 个 fixture file；`ls openapi/fixtures/PracticeSessions/listPracticeSessions.json` 存在；`grep -rn "DebriefRoundType\|DebriefQuestionSource\|DEBRIEF_NOT_FOUND\|IDEMPOTENCY_KEY_MISMATCH" frontend/src/lib/conventions/ frontend/src/api/generated/` 命中
  - UI source parity：Vitest `expect(screen.getByTestId('debrief-*'))` 测试 + jsdom DOM snapshot 匹配 prototype 锚点；Playwright 覆盖 DOM anchors、computed styles 与 viewport geometry
  - Pixel parity：Playwright debrief gate validates desktop (1440×900) + mobile (390×844) DOM anchors, bounding boxes, theme/customAccent computed values, and non-empty screenshot smoke
  - Runtime route / async polling drift gate：`getJob` 或任何 frontend-consumed polling operation 的完成证据必须同时覆盖 generated client + fixture + real `backend/cmd/api` route mount + handler/store owner scope + focused Go tests；BUG-0070 证据为 `go test ./internal/jobs ./internal/api/jobs ./internal/store/jobs ./cmd/api -count=1`
  - 隐私红线：Vitest fixture spy 不接收 raw entries / notes；URL/localStorage/sessionStorage/console.log 扫描
  - Legacy negative：`grep -rn "experience_library\|star_editor\|drill_builder\|mistakes_book\|growth_center\|report_timeline" frontend/src/app/screens/debrief/ frontend/src/app/i18n/locales/ test/scenarios/e2e/p0-06[56789]-*` 不命中

## 2026-05-18 Mock Flow Remediation

用户在默认 Vite dev fixture-backed mock 中无法看到 Step 1 `复盘分析` 与 Step 2 `复盘面试`：`createDebrief` 返回的 `debrief_generate` job id 继续被 `getJob` 的 generic `default` fixture 解析为 `report_generate/running`，导致 Step 1 永久显示 `AI 分析中...`，Step 2 无法访问。本次修复维持现有 UI 与 OpenAPI 语义不变，只补齐 dev mock 的异步 job 场景推进：

- `openapi/fixtures/Jobs/getJob.json` 增加 `debrief-succeeded` 场景，匹配 `createDebrief` default 返回的 debrief job。
- `frontend/src/api/devMockClient.ts` 记录 dev mock 中 `POST /debriefs` 返回的 debrief job id，并在后续无显式 `Prefer` 的 `GET /jobs/{jobId}` 请求上自动选择 `Prefer: example=debrief-succeeded`；同时把 `goal='debrief'` 的 replay plan/session 请求自动映射到 debrief-derived practice fixtures。
- 回归测试覆盖 generated dev mock client 的 `createDebrief -> getJob` 状态推进、debrief-derived practice fixture 选择，以及 `DebriefScreen` 在真实 fixture-backed dev mock client 下从 Step 0 提交进入 Step 1 分析，再进入 Step 2 并触发 fresh practice handoff。

## 4 实施步骤

### Phase 0: 依赖验证 + ui-design source map + 包结构

#### 0.1 backend-debrief Phase 0 完成验证

- 验证 `frontend/src/api/generated/` 中存在 `createDebrief` / `getDebrief` / `suggestDebriefQuestions` method types
- 验证 `openapi/fixtures/Debriefs/createDebrief.json` / `getDebrief.json` / `suggestDebriefQuestions.json` 存在并通过 `make validate-fixtures`
- 验证 `frontend/src/lib/conventions/` 与 generated client 含 `DebriefRoundType` / `DebriefQuestionSource` / `DEBRIEF_NOT_FOUND` / `IDEMPOTENCY_KEY_MISMATCH` 字面量
- 验证 `frontend/src/api/generated/` 中存在 `listPracticeSessions`（backend-practice/B2 addendum）；同时验证 `listResumes` / `listResumeVersions(resumeAssetId)` 可用；验证 `openapi/fixtures/PracticeSessions/listPracticeSessions.json` 存在并通过 `make validate-fixtures`
- 验证 backend-practice 现状支持 `goal='debrief'` + 合法 `mode IN ('assisted','strict')`（grep + test names from backend-debrief/001 Phase 0.6 / backend-practice/004）
- 未通过任一验证 → 暂停 plan 001，等 backend-debrief/001 Phase 0 完成

#### 0.2 ui-design source map 记录

在本 plan 的 phase commit message 与 work-journal 记录每个组件的 source anchor：
- `DebriefScreen` → `ui-design/src/screens-p1-depth.jsx:38-368` `DebriefFullScreen`
- `DebriefContextStrip` → `screens-p1-depth.jsx:412-432`
- `DebriefContextPickerModal` → `screens-p1-depth.jsx:434-518`
- `GuidedDebriefRecord` → `screens-p1-depth.jsx:519-619`
- `VoiceDebriefRecord` → `screens-p1-depth.jsx:656-870` (UI shell only)
- `DebriefReplayPlan` → `screens-p1-depth.jsx:1388-1421`

#### 0.3 前端包结构

新建 `frontend/src/app/screens/debrief/`：
- `DebriefScreen.tsx`（主屏组件）
- `components/`（5 个子组件 + 3 个 picker modal）
- `hooks/`（polling + suggestions + reducer integration hooks）
- `reducer.ts`（`SET_DEBRIEF_CONTEXT` reducer action 扩展，与 InterviewContext 整合）
- `types.ts`（local types，import generated types from `frontend/src/api/generated/`）
- `i18n/`（zh.ts / en.ts，新 `debrief.*` namespace；后期与全局 i18n 合并）
- `*.test.tsx` / `*.test.ts`（同位测试文件）

#### 0.4 route 接线（替换 PlaceholderScreen）

修改 `frontend/src/app/App.tsx` / `frontend/src/app/normalizeRoute.ts` 或等价 router config：
- `App.tsx` 中 `case "debrief"` → `<DebriefScreen>`，移除原 PlaceholderScreen 对 debrief 的占位
- `normalizeRoute.ts` 中加入历史 alias：`debrief_full` → `debrief`
- 不在 `frontend/src/app/routes.ts` 的正式 `RouteName` / primary nav / `INTERVIEW_CONTEXT_ROUTES` 中新增 `debrief_full`
- 确认 TopBar 一级导航 `debrief` 入口高亮逻辑保留

### Phase 1: DebriefScreen shell + Header + ContextStrip + Stepper

#### 1.1 DebriefScreen container

实现 `<DebriefScreen>` 组件：
- 接收 `InterviewContext`（targetJobId? + sessionId? + resumeVersionId? + language）
- Internal state：`step` (0/1/2 default 0)、`inputMode` ('text'|'voice' default 'text')、`entries` (DebriefEntry[])、`selectedContext` ({targetJob,mockSession,resume})、`pickerType` (null|'targetJob'|'mockSession'|'resume')、`suggestions` (SuggestedQuestion[] | null)、`activeGuide` (number)、`activeCard` (string)、`pollingState` (idle|running|succeeded|failed|timeout)、`debriefId` / `debriefJobId`
- Render Header + ContextStrip + Stepper + current step panel

#### 1.2 DebriefHeader

- 复刻 prototype lines 122-144：返回按钮 + eyebrow + H1 + 副文案 + 右上 meta
- eyebrow 从 InterviewContext 派生（targetJob.companyName + roundType + day/month）；若缺失显示 default 文案
- 不引入任何 raw text from prototype hardcoded data；所有 dynamic value 来自 generated `getTargetJob` 等

#### 1.3 DebriefContextStrip

- 复刻 prototype lines 412-432：三卡片
- 点击触发 `setPickerType('targetJob'|'mockSession'|'resume')`
- 显示三选当前选择的 title / display name；如选择缺失显示 default "未选择" 提示
- 通过 generated `getTargetJob(targetJobId)` / `getResumeVersion(resumeVersionId)` / `getPracticeSession(sessionId)` 拉取 display name；失败时 fallback 显示 ID

#### 1.4 Stepper

- 复刻 prototype lines 148-156：3 步骤 (`复盘记录` / `复盘分析` / `复盘面试`)
- 当前 step state 显示高亮
- 用户可点击已访问过的 step 返回（step 1/2 已访问后允许回到 step 0，但 entries 状态保留；不允许跳跃前进）

### Phase 2: 3 个 in-page picker modal

#### 2.1 DebriefContextPickerModal 通用骨架

- 复刻 prototype lines 434-518：modal container + title + body description + options 列表 + 确认/取消按钮
- 接收 `kind` ∈ {'targetJob','mockSession','resume'}、`options` 列表、`selectedId`、`onClose`、`onConfirm`
- modal 弹层不离开 debrief 页；点击外部 / Esc 关闭
- 移动端折叠为全屏 sheet（D-spec 设计约束）

#### 2.2 JD picker

- 调用 `listTargetJobs({analysisStatus:'ready'})` 拉已解析完成的用户岗位；不得使用 TargetJob lifecycle `status='ready'`
- 单选；用户选择后 onConfirm 写入 `selectedContext.targetJob`
- 完成后 `SET_DEBRIEF_CONTEXT` reducer action 写入 InterviewContext

#### 2.3 Mock Session picker

- 调用 Phase 0 已生成的 `listPracticeSessions({targetJobId, status:'completed'})` 拉相关 session；若 generated client 缺失该 method，立即 BLOCK 并回 backend-practice/B2 addendum
- 如 operation 已存在但不支持 server-side filter `status`，client-side filter `status==='completed'`；记录 fallback decision
- 单选 + "暂不关联模拟面试" option（Mock 是 optional）
- 用户选择后写入 `selectedContext.mockSession` + InterviewContext

#### 2.4 Resume picker

- 调用 `listResumes()` 拉用户简历资产列表，筛选 active/ready asset；用户选中 asset 后调用 `listResumeVersions(resumeAssetId)` 拉版本列表并筛选 ready version
- 单选 resume version；用户选择后写入 `selectedContext.resumeAsset` + `selectedContext.resumeVersion`（InterviewContext 字段仍使用 `resumeVersionId` 传给 backend-debrief）

#### 2.5 ContextStrip 三选完成后自动触发 suggestions

- 在 ContextStrip useEffect 中检测 `selectedContext.targetJob && selectedContext.resume` 都已选择
- 触发 `setSuggestionsTrigger`（debounce 500ms）→ Phase 4 的 suggestions hook 监听

### Phase 3: Step 0 复盘记录 + 跨模式共享 entries + Voice UI shell

#### 3.1 顶部统一汇总条

- 复刻 prototype lines 162-185：total entries 计数 + source chips（recorded / text / voice / manual）+ 跨模式共享提示

#### 3.2 Mode toggle

- 复刻 prototype lines 187-210：`Text` / `Voice` 双 tab
- `inputMode` state 切换；切换不清空 entries
- toggle 视觉同 prototype（icon + label + highlight ring）

#### 3.3 GuidedDebriefRecord (text mode)

- 复刻 prototype lines 519-619
- 接收 `suggestions[]`（来自 Phase 4 hook）+ `entries[]` + `setEntries` + `activeGuide` + `setActiveGuide`
- 左侧渲染 `currentGuide` (suggestions[activeGuide])：stage / questionText / whyLikelyAsked / source
- 上/下导航按钮切换 `activeGuide`
- 4 个 CTA：
  - 「遇到过，记录」→ 打开 inline editor，用户填写非空回答摘要后 `setEntries(prev => [...prev, makeEntryFromSuggestion(currentGuide, source:'ai_confirmed')])`，然后 `setActiveGuide(activeGuide+1)`
  - 「没问到，跳过」→ `setActiveGuide(activeGuide+1)`，不写 entries
  - 「改成真实问题」→ 打开 inline edit 编辑器；保存真实问题 + 非空回答摘要后 `setEntries(prev => [...prev, makeEntryFromEdit(...)])` source='ai_edited'，setActiveGuide(activeGuide+1)
  - 「手动添加真实问题」→ 打开 inline manual form；保存真实问题 + 非空回答摘要后 source='manual'
- 右侧渲染 entries 列表：点击选中 `activeCard`；展开详情显示 stage / q / a / follow / reaction / reflection / tag
- 隐私：entries.q / a / follow 仅在用户当前 session 显示，不写入 localStorage / telemetry

#### 3.4 VoiceDebriefRecord (UI shell only, D-6)

- 复刻 prototype lines 656-870 视觉
- 渲染：toggle highlight + idle 状态视觉占位 + 待确认卡片列表（从 entries filter source==='voice_extracted' 派生，初始为空）+ 「空格暂停/继续」键盘提示文案（UI only，不绑定真实 keyboard listener）+ 共享 entries 列表
- 显示固定占位 `<div data-testid="debrief-voice-not-implemented">语音复盘集成中，敬请期待</div>`
- 不绑定真实 Web Audio API / SpeechRecognition；不调用任何 STT endpoint
- 切换回 text 模式 entries 仍保留

#### 3.5 Submit CTA

- 复刻 prototype lines 314-316：「生成复盘分析」CTA
- disabled 条件：`entries.length === 0` 或 `selectedContext.targetJob === null` 或任一 entry 缺少非空 `myAnswerSummary`；voice 模式不阻止 submit，因为用户可能在 text 模式录入后切到 voice 查看，再回 text submit
- 点击触发 Phase 5 createDebrief

### Phase 4: suggestDebriefQuestions 集成（自动 + 手动）

#### 4.1 useSuggestDebriefQuestions hook

实现 React hook：
```ts
function useSuggestDebriefQuestions({
  targetJobId, sessionId, resumeVersionId, language, count = 6, enabled = true
}): { suggestions, loading, error, refetch }
```
- 当 `enabled === true` 且 `targetJobId` 非空时调用 generated `suggestDebriefQuestions`
- 返回 `suggestions: SuggestedQuestion[] | null`、`loading: boolean`、`error: ApiError | null`、`refetch: () => void`
- 自动触发 condition：targetJobId 改变 → debounce 500ms → 调用

#### 4.2 触发整合

- DebriefScreen 在 ContextStrip 三选完成后 enable hook
- GuidedDebriefRecord 接收 `suggestions` 渲染
- 用户可点击 "重新生成推荐" 按钮触发 `refetch`

#### 4.3 失败降级

- `error.code` 为 B1 canonical `AI_PROVIDER_CONFIG_INVALID` / `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_FALLBACK_EXHAUSTED` → 显示 inline error "推荐生成失败，可手动添加问题"
- 不阻塞 step 0；用户可继续手工添加 entries
- 点击 "重新生成推荐" 重试

### Phase 5: createDebrief + 双轨 polling + 失败态

#### 5.1 createDebrief 提交

实现 hook `useSubmitDebrief`:
- 接收 `{targetJobId, roundType, interviewerRole?, language, entries, notes?}`
- 转换 `entries → DebriefQuestionInput[]`（map q→questionText, a→myAnswerSummary, follow+reflection→interviewerReaction）；提交前再次 trim 并拒绝空 `myAnswerSummary`
- 生成 Idempotency-Key (UUIDv4 from `crypto.randomUUID()`)
- 调用 generated `createDebrief(payload, {Idempotency-Key})`
- 处理响应：
  - 202 → 写入 InterviewContext SET_DEBRIEF_CONTEXT (debriefId, debriefJobId=job.id)；不得写现有 `jobId`；调用 setStep(1) + 启动 polling
  - 422 VALIDATION_FAILED → inline error 列出失败字段
  - 409 IDEMPOTENCY_KEY_MISMATCH → 自动重生 IK 重试一次
  - 401 → `useRequestAuth({type:'submit_debrief', route:'debrief', params:{entries,...}})`
  - 5xx → toast + retry CTA

#### 5.2 双轨 polling hook

实现 `useDebriefPolling({debriefJobId, debriefId})`:
- Phase A: 指数退避 polling `getJob(debriefJobId)`（初始 1.5s × 1.5 上限 8s, max attempts=30）
- visibility/focus event listener 暂停-恢复 polling（document.visibilityState）
- job.status='succeeded' → 停止 phase A polling，触发 phase B
- job.status='failed' → 触发 `setPollingState('failed')` + 保存 errorCode
- max attempts → 触发 `setPollingState('timeout')`
- Phase B: 一次性调用 `getDebrief(debriefId)` 拉取 enriched data；返回到组件
- pollingState 状态机：idle → running → succeeded | failed | timeout

#### 5.3 失败态组件

- `<DebriefFailureState errorCode>`：复刻 frontend-report-dashboard 失败卡片 layout；errorCode 文案映射；CTA「返回 step 0 编辑」(setStep(0))、「重试生成」(resubmit createDebrief with new IK)
- `<DebriefMissingContextState>`：缺 targetJobId 时；卡片 + CTA「选择目标岗位」(自动 open JD picker)
- `<DebriefTimeoutState>`：polling 超时；CTA「重试」(重启 polling)、「返回 step 0」

#### 5.4 InterviewContext reducer 扩展

在 InterviewContext reducer 新增：
```ts
case 'SET_DEBRIEF_CONTEXT':
  return { ...state, debriefId: action.payload.debriefId, debriefJobId: action.payload.debriefJobId, ...rest }
```
- 不破坏既有 `SET_PRACTICE_CONTEXT` / `SET_REPORT_CONTEXT` 等 actions
- 不写现有 `jobId` 字段；该字段在当前 frontend-workspace context 中是 target job alias/fallback
- 同步扩展 `PENDING_ACTION_INTERVIEW_KEYS` 覆盖 `practiceGoal` / `debriefId` / `debriefJobId`，并补登录恢复 round-trip 测试
- 增加单元测试覆盖

### Phase 6: Step 1 分析渲染 + Step 2 复盘面试 launcher + handoff

#### 6.1 Step 1 分析渲染

- 当 `pollingState === 'succeeded'` 且 `debrief.status === 'completed'` 时渲染：
- 风险项列表（从 `debrief.riskItems` 派生）：每项 label + severity color tag（low=gray, medium=amber, high=red）
- 维度对比卡片三张（与模拟面试 / JD / 简历）：从 `debrief.questions[*].aiAnalysis` 与 ContextStrip 上下文派生（具体派生逻辑由 backend AI 输出结构决定；plan 内默认按 mock layout 渲染，等 backend AI 输出 schema 稳定后细化）
- 「关于本次分析」展开区：显示 `debrief.provenance` 6 字段（promptVersion / rubricVersion / modelId / language / featureFlag / dataSourceVersion）
- 底部 CTA「生成复盘面试」→ setStep(2)

#### 6.2 Step 2 复盘面试 launcher

- 复刻 `<DebriefReplayPlan>` lines 1388-1421
- 渲染：复盘面试 plan 预览（复现真实问题 + 薄弱处追问 + 真实顺序 + 简历证据对比）
- 内容来自 `debrief.questions` + `debrief.riskItems`（前端组装预览文本）
- CTA「开始复盘面试」→ Phase 6.3

#### 6.3 复盘面试 fresh session handoff

- 已登录时先调用 `createPracticePlan({goal:'debrief', sourceDebriefId: debriefId, targetJobId, resumeAssetId, mode, language, ...})`，再调用 `startPracticeSession({planId, hintsEnabled})`
- 成功后调用 `nav('practice', {practiceGoal:'debrief', mode:'text', modality:'text', planId, sessionId: newSessionId, targetJobId, resumeVersionId, debriefId, language})`
- 未登录走 `useRequestAuth({type:'start_debrief_interview', route:'debrief', params:{debriefId, targetJobId, resumeVersionId}})`，登录恢复后回到 debrief 重新执行 CTA；不得把 optional completed mock session id 作为 replay practice session 转发

### Phase 7: i18n + 主题 + 响应式

#### 7.1 i18n `debrief.*` namespace

- 在 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 新增 `debrief.*` keys：header / contextStrip / stepper / step0Record / step1Analysis / step2Interview / pickers / failureStates / suggestions / voice 等
- 不复用 `workspace.*` / `practice.*` / `report.*`
- 翻译完整 zh + en；不引入其他 locale

#### 7.2 主题适配

- 验证 dark mode / customAccent 在 DebriefScreen 各 step 与 picker modal 中正常工作
- 通过 root `data-theme` / `data-mode` / `data-custom-accent` 已被 frontend-shell 配置

#### 7.3 Mobile 响应式

- Header 紧凑布局（meta 右上隐藏，移到 secondary line）
- ContextStrip 三卡片单列折叠
- Stepper 横向滑动或缩短文案
- Step 0 双栏（guide + entries）折叠为单列 + Tab 切换；Voice UI shell 单列
- Step 1 风险列表 + 维度卡片单列；展开区可折叠
- Step 2 launcher CTA sticky bottom
- 3 个 Picker modal 转为全屏 sheet
- Vitest + jsdom 媒体查询测试 + Playwright mobile viewport 测试

### Phase 8: Playwright pixel parity + 隐私 + legacy negative + BDD

#### 8.1 Playwright pixel parity

- 新增 `frontend/tests/pixel-parity/debrief.spec.ts`
- 对 `frontend/src/app/screens/debrief/DebriefScreen.tsx` 渲染结果执行 source-level parity smoke；source anchor 仍以 `ui-design/src/screens-p1-depth.jsx` / `docs/ui-design/` 为 truth source
- 断言：DOM 锚点（testid）、computed style key 值（color/spacing/font-size/border-radius）、bounding box 区域比例、非空 screenshot smoke；checked-in screenshot baseline diff 留给后续专门基线 plan
- viewport: desktop 1440×900 + mobile 390×844
- 主题: light / dark / customAccent 各一次

#### 8.2 隐私 + telemetry 验证

- Vitest fixture spy 注入 marker `__SECRET_RAW_TEXT__` 在 entries 中；submit createDebrief 后 spy 接收 raw body 但不应写入 console.log / localStorage / sessionStorage / telemetry payload
- grep gate: `grep -rn "questionText\|myAnswerSummary\|interviewerReaction\|notes" frontend/src/app/screens/debrief/ frontend/src/app/i18n/locales/ | grep -v "_test\|generated\|.types" | grep -v "// privacy reviewed"` 应只命中合理位置（如 type definitions / mapped value）
- URL/search params 检查：nav 时 params 仅含 stable IDs + display knobs；不传 entries body

#### 8.3 Legacy negative grep

- `grep -rn "experience_library\|star_editor\|drill_builder\|mistakes_book\|growth_center\|report_timeline" frontend/src/app/screens/debrief/ frontend/src/app/i18n/locales/ test/scenarios/e2e/p0-06[56789]-*` 0 命中
- 在 `scripts/lint/` 新增 `frontend_debrief_legacy.py`，作为 pytest lint script
- `python3 -m pytest scripts/lint -q` 通过

#### 8.4 BDD scenarios P0.065-069

- 在 `test/scenarios/e2e/` 新建：`p0-065-debrief-default-render-and-pickers/` / `p0-066-debrief-text-suggestions-and-submit/` / `p0-067-debrief-polling-happy-and-analysis/` / `p0-068-debrief-failure-and-handoff/` / `p0-069-debrief-pixel-parity-and-legacy-negative/`
- 每个 scenario 含 `scripts/setup.sh` / `scripts/trigger.sh` / `scripts/verify.sh` / `scripts/cleanup.sh` 四段脚本；不创建或引用额外 wrapper
- 登记到 `test/scenarios/e2e/INDEX.md`
- 执行：在每个 scenario 目录内按 `scripts/setup.sh -> scripts/trigger.sh -> scripts/verify.sh -> scripts/cleanup.sh` 顺序运行

#### 8.5 Plan 收口

- `pnpm --filter @easyinterview/frontend test -- src/app/screens/debrief` 通过
- `pnpm --filter @easyinterview/frontend test -- --run` 全量通过
- `pnpm --filter @easyinterview/frontend lint` 通过
- `pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/debrief.spec.ts` (debrief Playwright gate)
- `python3 -m pytest scripts/lint -q` 通过
- `make docs-check` + `git diff --check` 通过
- 更新 plans/INDEX.md 把 001 移到 completed
- 更新 frontend-debrief/history.md 增加最新 completion / review-fix 行

### Phase 10: D-20 简历扁平化 picker + resumeId

> product-scope D-20 / spec D-19。依赖 B2 004 Phase 7（contract collapse）+ generated client 重生。Resume picker 改扁平——直接列 `listResumes` 单选简历，删除 asset→version 二级展开 / `listResumeVersions`；`InterviewContext` / nav payload / `suggestDebriefQuestions` / `createPracticePlan` 的 `resumeVersionId` / `resumeAssetId`→`resumeId`；`getResumeVersion`→`getResume`；`SET_DEBRIEF_CONTEXT` 写 `resumeId`。详见 spec D-19。（Phase 9 为既有 Plan 收口，D-20 续编为 Phase 10。）

#### 10.1 实施

Resume picker 改扁平——直接列 `listResumes` 单选简历，删除 asset→version 二级展开 / `listResumeVersions`；`InterviewContext` / nav payload / `suggestDebriefQuestions` / `createPracticePlan` 的 `resumeVersionId` / `resumeAssetId`→`resumeId`；`getResumeVersion`→`getResume`；`SET_DEBRIEF_CONTEXT` 写 `resumeId`。详见 spec D-19。（Phase 9 为既有 Plan 收口，D-20 续编为 Phase 10。）

（验证：vitest 组件/adapter/route + pixel parity + typecheck + build PASS）

#### 10.2 收口

零版本树残留 grep（`resumeVersionId` / `resumeAssetId` / `listResumeVersions` / 版本树组件，generated adapter 除外）+ `sync-doc-index --check`。

（验证：全 gate PASS + 负向 grep 0 命中）

## 5 验收标准

- C-1 ~ C-18（[spec §6](../../spec.md#6-验收标准)）全部通过
- 本 plan 列出的 Phase 0-8 实现项全部按 checklist 勾选
- [BDD-Gate](./bdd-checklist.md) `E2E.P0.065-069` 全部通过
- [test-checklist](./test-checklist.md) 单元测试与 Playwright 测试项全部通过
- backend-debrief/001 Phase 0 cross-owner addendum 已完成
- legacy negative gate 通过

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| backend-debrief/001 Phase 0 cross-owner addendum 未及时落地，frontend-debrief Phase 0 验证失败 | Phase 0.1 验证失败时立即暂停 plan 001，与 backend-debrief team 协调 PR 优先级；不在 frontend 端 mock 任何 generated client 缺失的字面量 |
| Mock Session picker 所需 `listPracticeSessions` operation 未生成 | Phase 0.1 直接 BLOCK，回 backend-practice/B2 addendum 新增 operation + fixture + generated client；不得在 frontend 手写 ad hoc fetch 或复制 mock data |
| `listPracticeSessions` 已生成但不支持 `status='completed'` server-side filter | Phase 2.3 client-side filter 并记录 fallback；plan 002 可考虑回 B2 加 filter |
| AI 推荐 suggestions 的具体内容格式与 GuidedDebriefRecord prototype mock 数据结构不一致 | Phase 4.1 hook 内部 schema 标准化：将 `SuggestedQuestion{stage?, questionText, whyLikelyAsked, source}` 映射到 GuidedDebriefRecord 期待的字段；如缺 stage 默认 "通用"；如缺 source 默认 'jd' |
| Step 1 分析渲染时 backend AI 输出的 risk_items 与 questions[*].aiAnalysis 结构不稳定（前端期待 prototype mock data 形态） | Phase 6.1 内做 defensive parsing：缺少字段时显示默认占位；不抛出错误；通过 Vitest 断言至少正常渲染 schema-minimum 输出 |
| Voice UI shell 在 mobile 端 sticky CTA 布局冲突 | Phase 7.3 单独测试 mobile viewport；如冲突，调整 sticky 优先级（提交 CTA > 占位提示）|
| Playwright pixel parity 因字体加载 / 动画 / blur 等差异多次失败 | Phase 8.1 使用 deterministic settings：禁用 transition / animation；font 用本地 self-hosted；screenshot 用 `disableAnimations` flag；本 plan 的 close-out gate 先使用 DOM / computed style / bounding box / screenshot smoke，checked-in screenshot baseline diff 留给后续专门基线 plan |
| InterviewContext reducer 新增 `SET_DEBRIEF_CONTEXT` action 与 frontend-workspace-and-practice 既有 action 冲突 | Phase 5.4 在新增 action 前 grep 既有 reducer；与 frontend-workspace-and-practice owner co-review；如必要新建独立 sub-state slice 而不是直接 merge |
| createDebrief 提交后 page navigation 或刷新丢失 entries 草稿 | spec §3.2 已声明默认 P0 不持久化；如用户反馈强烈，plan 002 评估 localStorage 草稿（注意隐私边界）；本 plan 显示 "草稿将丢失" 离开提示对话框 |
