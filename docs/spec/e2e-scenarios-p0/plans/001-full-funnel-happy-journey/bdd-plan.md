# 001 Full Funnel Happy Journey BDD Plan

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-05-24

**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 0 目标

为 `e2e-scenarios-p0/001-full-funnel-happy-journey` 定义跨模块完整漏斗 happy 主干 journey 的 BDD 场景集。两条 journey 共享真后端全栈环境（dev-stack postgres + scenario stub / fixture AI），区别在 driver 与验证侧重：

- `E2E.P0.098`：API-level，验证后端域间真实贯通（handler/store/internal runner/event/DB + handoff 链 + 幂等 + 隐私）。
- `E2E.P0.099`：Playwright 全栈，验证前端在真后端下走完漏斗（跨屏 nav + 真实轮询 UI + CTA handoff + 隐私）。

执行入口：每个场景目录按顺序运行 `scripts/setup.sh → scripts/trigger.sh → scripts/verify.sh → scripts/cleanup.sh`；wrapper 使用 `rc=0; bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh || rc=$?; cleanup_rc=0; bash scripts/cleanup.sh || cleanup_rc=$?; [ "$rc" -ne 0 ] && exit "$rc"; exit "$cleanup_rc"`，确保 cleanup 总会执行且不会把前置失败改写成通过；`trigger.sh` 保留真实 runner（Go test / Playwright）exit code，`verify.sh` 必须断言 runner 专属 pass marker（Go `ok`/`PASS` + 目标 test 名，或 Playwright `passed` 计数）并拒绝 no-op / skip-as-pass，同时包含隐私与旧口径负向 grep。

## 1 Scenario Matrix

| 场景 ID | category | 关联 spec AC | 关联 plan phase | 关联 checklist BDD-Gate |
|---------|----------|--------------|-----------------|------------------------|
| E2E.P0.098 | Primary + Cross-layer contract + Boundary/idempotency + Privacy + Regression/legacy-negative | C-1, C-2, C-3, C-4, C-5, C-6, C-7 | Phase 1 | 3.2 |
| E2E.P0.099 | Primary + Cross-layer contract + Privacy + Regression/legacy-negative | C-1, C-2, C-4, C-5, C-6 | Phase 2 | 3.3 |

## 2 场景详情

### E2E.P0.098 — API-level Full Funnel: Import to Next Round

| 字段 | 内容 |
|------|------|
| 目录 | `test/scenarios/e2e/p0-098-full-funnel-import-to-next-round-journey/` |
| Phase | Phase 1 |
| 关联 spec AC | C-1, C-2, C-3, C-4, C-5, C-6, C-7 |
| 执行入口 | `rc=0; bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh || rc=$?; cleanup_rc=0; bash scripts/cleanup.sh || cleanup_rc=$?; [ "$rc" -ne 0 ] && exit "$rc"; exit "$cleanup_rc"`（在该场景目录内执行） |
| Given | `make dev-up` postgres 可达且 `make migrate-up` 至最新；`config.LoadCanonical(AppEnv:"test")` 加载且 `resume_parse` / `target_import` / practice / `report_generate` 所需 AI profile/registry 在未配置 `AI_PROVIDER_*` 时可解析，实际 AI 由 scenario harness 注入确定性 stub / fixture client；`backend/cmd/api` httptest server 组装真实 router/handler/store/internal runner/events；journey 内 seed 已认证 user A |
| When | 顺序调用：(0) `registerResume` 创建 resume asset + `resume_parse` job，并用 `resume.parse.default` stub 推进至 ready 得到 `resumeAssetId`；(1) `importTargetJob`（paste JD）→ `targetJobId` + `target_import` job；(2) 真实 runner 处理后轮询 `getTargetJob` 至 `analysisStatus=ready`；(3) `createPracticePlan(targetJobId, resumeAssetId, goal=baseline)` → `planId`；(4) `startPracticeSession(planId)` → `sessionId` + 首题；(5) `appendSessionEvent` 逐题作答；(6) `completePracticeSession(sessionId)` → `reportId` + `report_generate` job；(7) 真实 runner 处理后轮询 `getFeedbackReport` 至 `status=ready`，若返回/记录 jobId 则以 `getJob` 作为备选轮询 / handler gate；(8) `createPracticePlan(goal=next_round, sourceReportId=reportId, targetJobId, resumeAssetId)` → 派生 `planId`；(9) 对 register/start/complete/createPlan 同 Idempotency-Key replay 一次 |
| Then | (a) 每步返回真实响应 envelope，handoff ID 真实传递并被下一步消费：`resumeAssetId → targetJobId → planId → sessionId → reportId → 派生 planId`；(b) `resume_parse` / `target_import` / `report_generate` 经真实 internal runner 完成，resource status 由 queued/processing → ready，对应行真实落库；(c) 派生 plan 真实关联 `sourceReportId` 且 `id ≠` 首个 planId；(d) 同 key replay 无重复副作用（无第二 resume/session/report/plan、无重复 outbox）；(e) journey 全程响应 / event / audit / log / 持久化可观测面不含 JD 原文 / 答案文本 / 报告 prose；(f) `cd backend && go test -v ./cmd/api -run 'TestE2EP0098' -count=1` 输出 `PASS` + 目标 test 名 |
| Cleanup | 删除 journey 创建的 target_jobs / practice_plans / practice_sessions / session_events / feedback_reports / jobs / resume_assets 行；关闭 httptest server；postgres 不可达时记录 skip 原因 |
| Privacy 反查 | `verify.sh` 断言 trigger.log / 响应 dump 不含 JD raw text / answer text / report prose 字段值 |
| Legacy 反查 | `verify.sh` 含 route-aware negative pattern，覆盖 `(^\|[[:space:]'"'/#?&=:-])(welcome\|growth\|mistakes\|drill\|followup\|experiences\|star(_editor)?\|onboarding)([[:space:]'"'/#?&=:-]\|$)\|mode=debrief\|name=['\"](plan\|resume\|voice)['\"]\|route=['\"](plan\|resume\|voice)['\"]\|#route=(plan\|resume\|voice)([[:space:]'"'/#?&=:-]\|$)`，并避免误伤合法 `startPracticeSession` / `createPracticePlan` / `resumeAssetId` |

### E2E.P0.099 — Full-Stack UI Full Funnel Journey

| 字段 | 内容 |
|------|------|
| 目录 | `test/scenarios/e2e/p0-099-full-funnel-fullstack-ui-journey/` |
| Phase | Phase 2 |
| 关联 spec AC | C-1, C-2, C-4, C-5, C-6 |
| 执行入口 | `rc=0; bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh || rc=$?; cleanup_rc=0; bash scripts/cleanup.sh || cleanup_rc=$?; [ "$rc" -ne 0 ] && exit "$rc"; exit "$cleanup_rc"`（在该场景目录内执行） |
| Given | `setup.sh` 拉起真后端进程（连 dev-stack postgres，`APP_ENV=test`，场景 AI 使用确定性 stub / fixture client）+ 前端 build/preview 以 `VITE_EI_API_MODE=real` / `VITE_EI_API_BASE_URL=http://127.0.0.1:<backend-port>/api/v1` 指向真后端 base URL（非 fixture mock transport）；通过 `registerResume` + `resume_parse` scenario stub/fixture AI seed 已认证 user 的 ready resume asset；health probe 确认前后端就绪 |
| When | Playwright 驱动真实 UI：(1) 首页导入 JD（paste）；(2) ParseScreen 真实轮询至解析 ready 并 Confirm；(3) 进 WorkspaceScreen 选 resume + 立即面试；(4) PracticeScreen 完成 session；(5) Generating 真实轮询；(6) ReportDashboard 渲染 ready 报告；(7) 点击「进入下一轮」CTA |
| Then | (a) 跨屏 nav 正确：home → parse → workspace → practice → generating → report；(b) 解析 loading 与 report generating 的真实轮询 UI 在异步 job 推进下过渡到 ready（非 mock 即时返回）；(c)「进入下一轮」CTA 触发 `createPracticePlan(next_round, sourceReportId)` + `startPracticeSession`，nav query 含派生 `planId` + fresh `sessionId`；(d) URL / localStorage / sessionStorage / console 不泄露 JD 原文 / 答案 / 报告 prose；(e) scenario 树旧口径 grep 0 命中；(f) `EI_PLAYWRIGHT_OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/playwright" playwright test --config=playwright.e2e.config.ts tests/e2e/full-funnel-journey.spec.ts` 输出 `passed` 计数且无 `failed` / no-tests，trace / screenshot / video / runner artifacts 只写入 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/` |
| Cleanup | 删除 journey 创建的 DB 行；停真后端 / 前端进程；保留 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/` 下的 trigger.log 与 Playwright 证据产物；失败优先检查环境污染（框架 §8） |
| Privacy 反查 | `verify.sh` 断言 URL / storage dump / console dump 不含 raw JD / answer / report prose |
| Legacy 反查 | `verify.sh` 含 route-aware negative pattern，覆盖 `(^\|[[:space:]'"'/#?&=:-])(welcome\|growth\|mistakes\|drill\|followup\|experiences\|star(_editor)?\|onboarding)([[:space:]'"'/#?&=:-]\|$)\|mode=debrief\|name=['\"](plan\|resume\|voice)['\"]\|route=['\"](plan\|resume\|voice)['\"]\|#route=(plan\|resume\|voice)([[:space:]'"'/#?&=:-]\|$)`，并避免误伤合法 `startPracticeSession` / `createPracticePlan` / `resumeAssetId`；同时运行 frontend scope gate 或等价 scoped grep，证明独立 `voice` route 未回流 |

## 3 编号占用

本 plan 占用 `E2E.P0.098` ~ `E2E.P0.099`（2 个）。下一可用编号 `E2E.P0.100`（保留给本 subject 后续 journey：复练 / 真实复盘回流 / 失败恢复，由 `002+` plan 派生时锁定）。

## 4 与现有 slice 场景的关系

- 现有 `E2E.P0.001-097` 是单模块可独立收口切片，验证各 owner spec 的局部 C-* 条件。
- 本 plan 的 `E2E.P0.098-099` 是**跨模块 journey**，在 slice 之上叠加「完整漏斗真实 handoff 端到端贯通」验证，不替代、不重复 slice 的局部断言。
