# 001 Debrief Screen and Handoff BDD Plan

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-05-17

**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 0 目标

为 frontend-debrief/001-debrief-screen-and-handoff 定义端到端 BDD 场景集。每个场景含 setup / trigger / verify / cleanup 四段，覆盖用户行为流（路由切换、点击、表单提交、跨页 nav、轮询、失败态）；backend-debrief/001 独立占用 scenarios E2E.P0.060-064。

执行入口：每个场景目录必须提供并按顺序运行 `scripts/setup.sh` → `scripts/trigger.sh` → `scripts/verify.sh` → `scripts/cleanup.sh`；`trigger.sh` 必须保留真实 Playwright / Vitest runner exit code，verify.sh 必须断言前端 runner 专属 pass marker（例如 Vitest `Test Files ... passed` 或 Playwright expected pass output）并拒绝 no-op / no tests，同时包含旧口径 grep 反查。

## 1 Scenario Matrix

| 场景 ID | category | 关联 spec AC | 关联 plan phase | 关联 checklist BDD-Gate |
|---------|----------|--------------|-----------------|------------------------|
| E2E.P0.065 | Primary + UI source structure parity | C-1, C-2, C-3, C-11, C-14 | Phase 0-2 | 8.8 |
| E2E.P0.066 | Primary + Failure/recovery | C-4, C-5, C-7 | Phase 3-5 | 8.9 |
| E2E.P0.067 | Primary | C-8, C-12 | Phase 5-6 | 8.10 |
| E2E.P0.068 | Failure/recovery + Cross-layer | C-9, C-10, C-13 | Phase 5-6 | 8.11 |
| E2E.P0.069 | UI visual parity + Privacy + Regression/Legacy-negative | C-15, C-16, C-17, C-18 | Phase 7-8 | 8.12 |

## 2 场景详情

### E2E.P0.065 — Debrief Default Render + 3 Picker Modal

| 字段 | 内容 |
|------|------|
| 目录 | `test/scenarios/e2e/p0-065-debrief-default-render-and-pickers/` |
| Phase | Phase 0-2 |
| 关联 spec AC | C-1, C-2, C-3, C-11, C-14 |
| 执行入口 | `bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh; bash scripts/cleanup.sh`（在该场景目录内执行） |
| Given | 用户已认证（passwordless session cookie）；fixture `listTargetJobs` 返回 3 ready jobs；fixture `listPracticeSessions` 返回 2 completed sessions（Phase 0 addendum）；fixture `listResumes` 返回 2 active assets；fixture `listResumeVersions(resumeAssetId)` 返回对应 ready versions；fixture `getTargetJob/getResumeVersion/getPracticeSession` 返回有效数据；route normalization 已配置 `debrief_full -> debrief` |
| When | (1) nav `/debrief_full` 并确认 URL/route normalize 到 `/debrief`；(2) 用户点击 ContextStrip JD 卡片；(3) 在 JD picker 中选 tj-2 + 确认；(4) 用户点击 Mock Session 卡片 + 选 mock-24 + 确认；(5) 用户点击 Resume 卡片 + 选 resume-v3 + 确认；(6) ContextStrip 三选完成后等 500ms |
| Then | (a) DebriefScreen 渲染（含 Header / ContextStrip / Stepper / Step 0 Record panel）；testid `debrief-screen` / `debrief-header` / `debrief-context-strip` / `debrief-stepper-step-0` 命中；(b) 三个 picker modal 在 in-page 打开（不离开 debrief 页）；testid `debrief-picker-modal-targetJob` / `debrief-picker-modal-mockSession` / `debrief-picker-modal-resume` 各打开一次；(c) ContextStrip 三卡片更新显示 selected title；(d) 500ms 后自动触发 `suggestDebriefQuestions` 调用一次 with {targetJobId:'tj-2', sessionId:'mock-24', resumeVersionId:'resume-v3', language:'zh', count:6}；(e) TopBar 一级导航 `debrief` 高亮；(f) 正式 route catalog / TopBar 不含 `debrief_full` |
| Cleanup | 清空 InterviewContext / sessionStorage / localStorage；登出 |
| Privacy 反查 | verify.sh 含 `! grep "questionText\|notes" localStorage_dump.json` |
| UI source parity 反查 | verify.sh assert DOM 锚点存在 + 控件类型 (button vs link vs menu) 与 prototype 一致 |
| Legacy 反查 | verify.sh 含 `! grep "experience_library\|drill_builder\|mistakes_book" frontend/src/app/screens/debrief/` |

### E2E.P0.066 — Text Mode AI Suggestions + Entries + createDebrief Submit

| 字段 | 内容 |
|------|------|
| 目录 | `test/scenarios/e2e/p0-066-debrief-text-suggestions-and-submit/` |
| Phase | Phase 3-5 |
| 关联 spec AC | C-4, C-5, C-7 |
| 执行入口 | `bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh; bash scripts/cleanup.sh`（在该场景目录内执行） |
| Given | 用户已认证；fixture `suggestDebriefQuestions=default` 返回 6 suggestions；fixture `createDebrief=default` 返回 202 + DebriefWithJob{debriefId:'D', job:{id:'J'}}；用户已通过 E2E.P0.065 完成三选 |
| When | (1) 等待 suggestions 自动加载；(2) 用户对 suggestions[0] 点击 "遇到过，记录"；(3) 用户对 suggestions[1] 点击 "没问到，跳过"；(4) 用户对 suggestions[2] 点击 "改成真实问题" + inline edit + save；(5) 用户点击 "手动添加真实问题" + 表单 + save；(6) 用户点击 "重新生成推荐"（mock 返回 502 AI_PROVIDER_TIMEOUT）；(7) 用户切到 voice 模式查看 UI shell；(8) 切回 text 模式；(9) 用户点击 "生成复盘分析" CTA |
| Then | (a) suggestions 渲染 6 项；testid `debrief-suggested-question-{0..5}` 命中；(b) entries 写入 3 行（source: ai_confirmed / ai_edited / manual）且每行 `myAnswerSummary` 非空；testid `debrief-entry-card-{id}` 各显示；(c) 跳过的 suggestion 不入 entries；(d) 重新生成推荐失败时显示 inline error，manual CTA 仍可用；不阻塞 step 0；(e) Voice 模式 testid `debrief-voice-not-implemented` 占位提示出现；entries 列表保留；(f) 切回 text 模式 entries 仍为 3 行；(g) Submit CTA 点击后触发 `createDebrief` 调用 with Idempotency-Key UUIDv4 + 完整 questions[3] body（每项 `myAnswerSummary` 非空）；返回 202 + DebriefWithJob；(h) InterviewContext 写入 debriefId='D' + debriefJobId='J'，且不覆盖既有 jobId；(i) 自动 setStep(1) + 启动 polling |
| Cleanup | 同 P0.065 + 清空 entries |
| Privacy 反查 | verify.sh assert (a) URL 不含 raw text; (b) localStorage 不含 entries body; (c) console.log 不含 raw questionText |

### E2E.P0.067 — Polling Happy + Analysis Render

| 字段 | 内容 |
|------|------|
| 目录 | `test/scenarios/e2e/p0-067-debrief-polling-happy-and-analysis/` |
| Phase | Phase 5-6 |
| 关联 spec AC | C-8, C-12 |
| 执行入口 | `bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh; bash scripts/cleanup.sh`（在该场景目录内执行） |
| Given | 用户已通过 P0.066 完成 createDebrief submit + setStep(1)；fixture `getJob` 配置为前 3 次返回 status='running'，第 4 次返回 status='succeeded'；fixture `getDebrief=default` 返回 completed Debrief with riskItems=[3 items] + provenance 6 字段；runtime gate 已证明真实 `GET /api/v1/jobs/{jobId}` route 挂载并按 owner scope 查询 |
| When | (1) Polling 自动启动；(2) 用户等待 polling 完成；(3) Step 1 渲染；(4) 用户点击 "关于本次分析" 展开 provenance |
| Then | (a) `getJob('J')` 调用 4 次（按指数退避节奏）；(b) status='succeeded' 后 `getDebrief('D')` 调用 1 次；(c) Step 1 panel 渲染：风险项列表 3 项 + 维度卡 3 张 + provenance 展开区；testid `debrief-analysis-risk-item-{0,1,2}` / `debrief-analysis-dimension-{mock,jd,resume}` 命中；(d) 不渲染 nextRoundChecklist / thankYouDraft；grep `data-testid="debrief-next-round-checklist"` 0 命中；(e) Provenance 展开显示 6 字段（promptVersion / rubricVersion / modelId / language / featureFlag / dataSourceVersion）；不显示 feature_key / cost 等运行时字段；(f) 后端 route/store gate 覆盖 `getJob` owner scope，避免 mock-only polling false-green |
| Cleanup | 清空 InterviewContext + DB |

### E2E.P0.068 — Failure States + Cross-Owner Handoff

| 字段 | 内容 |
|------|------|
| 目录 | `test/scenarios/e2e/p0-068-debrief-failure-and-handoff/` |
| Phase | Phase 5-6 |
| 关联 spec AC | C-9, C-10, C-13 |
| 执行入口 | `bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh; bash scripts/cleanup.sh`（在该场景目录内执行） |
| Given | 用户已认证；scenarios 模拟 4 类失败 + 1 类成功 handoff |
| When | (1) 用户进入 `/debrief` 无 InterviewContext → DebriefMissingContextState；(2) 用户重新进入完整流程 → submit createDebrief → fixture `getJob=failed` 返回 status='failed' + errorCode='AI_PROVIDER_TIMEOUT' → DebriefFailureState；(3) 用户点击 "重试生成"（new IK） → 这次 fixture `getJob` 永久 queued → DebriefTimeoutState；(4) 用户点击 "返回 step 0 编辑"；(5) 重新 submit → fixture 成功 polling → Step 1 → Step 2 → 用户点击 "开始复盘面试" CTA |
| Then | (a) DebriefMissingContextState 渲染；JD picker 自动打开；testid `debrief-missing-context-state` 命中；(b) DebriefFailureState 渲染 errorCode 文案 + CTA「返回 step 0 编辑」+「重试生成」；testid `debrief-failure-state` 命中；errorCode 显示按 B1 AI_PROVIDER_TIMEOUT 文案映射，不暴露 raw provider error；(c) DebriefTimeoutState 渲染 timeout 卡片 + CTA「重试」+「返回 step 0」；testid `debrief-timeout-state` 命中；(d) "返回 step 0 编辑" 后 entries 保留；(e) Step 2 "开始复盘面试" CTA 触发 `createPracticePlan(goal='debrief', sourceDebriefId)` + `startPracticeSession`，然后 `nav("practice", {practiceGoal:'debrief', mode:'text', modality:'text', planId, sessionId:newSessionId, targetJobId, resumeVersionId, debriefId, language})`；(f) scenario 关键断言：fixture transport spy 确认新 session id 来自 `startPracticeSession` 响应，不复用 optional completed mock session id |
| Cleanup | 清空 InterviewContext + DB |
| Cross-owner 反查 | verify.sh assert Step 2 先创建 fresh debrief practice plan/session，nav 触发后 URL 切到 `/practice?...` 且包含 practiceGoal=debrief + fresh sessionId |

### E2E.P0.069 — Pixel Parity + i18n + Privacy + Legacy Negative

| 字段 | 内容 |
|------|------|
| 目录 | `test/scenarios/e2e/p0-069-debrief-pixel-parity-and-legacy-negative/` |
| Phase | Phase 7-8 |
| 关联 spec AC | C-15, C-16, C-17, C-18 |
| 执行入口 | `bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh; bash scripts/cleanup.sh`（在该场景目录内执行） |
| Given | 完整 DebriefScreen + Playwright debrief parity spec 已就绪 |
| When | (1) Vitest 跑 i18n / privacy / devMock fixture gates；(2) build frontend dist；(3) Playwright 加载 frontend `/debrief` desktop 1440×900 与 mobile 390×844；(4) 断言 `debrief_full` alias normalize、Step 0 source anchors、viewport bounding boxes、mobile overflow negative；(5) 切到 dark mode 与 customAccent；(6) 截图 smoke；(7) grep legacy terms in active runtime 与 P0.065-P0.069 scenario tree |
| Then | (a) Vitest runner 通过；(b) Playwright desktop + mobile debrief parity gate 通过（DOM anchors / computed style / bounding box / non-empty screenshot smoke）；(c) dark / customAccent 主题应用正确（root data-theme / data-mode / data-custom-accent）；(d) privacy boundary tests 通过；(e) retired terms `experience_library` / `star_editor` / `drill_builder` / `mistakes_book` / `growth_center` / `report_timeline` 在 `frontend/src/app/screens/debrief/` / `frontend/src/app/i18n/locales/` / `test/scenarios/e2e/p0-06[56789]-*` 全部 0 命中；(f) `getFeedbackReport` / `getCompanyIntel` 在 debrief 模块内 0 调用；Step 2 practice 创建调用仅限 handoff handler |
| Cleanup | clean |

## 3 编号占用

本 plan 占用 E2E.P0.065 ~ E2E.P0.069（5 个）。下一可用编号 E2E.P0.070（保留给未来 P0 plan）。

## 4 编号策略与与 backend-debrief 的对齐

- backend-debrief/001 占用 P0.060-064
- frontend-debrief/001（本 plan）占用 P0.065-069

完整 P0 闭环 P0.001-069 含全部 backend + frontend 域。Debrief 是 P0 闭环最后一个域，跨 backend (P0.060-064) + frontend (P0.065-069) 共 10 个 scenarios。
