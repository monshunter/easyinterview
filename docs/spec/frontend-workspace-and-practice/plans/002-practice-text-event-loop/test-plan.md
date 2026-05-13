# 002 — Practice Text Event Loop Test Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-13

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)

## 1 测试策略概览

本 plan 是 feature-behavior + cross-layer contract 组合。TDD（Vitest + @testing-library/react + jsdom）覆盖 hook / 组件 / utils；contract drift 覆盖 fixture parity + OpenAPI schema 同步；Pixel parity（Playwright）覆盖 UI visual geometry parity；BDD scenario 覆盖用户可见行为切片（answer_submitted 主路径 / hint / skip / pause-resume / strict×debrief 显隐 / 完成 + 生成态导航 / 失败恢复 / 旧口径负向）。本测试计划只覆盖 **单元 / 集成 / contract / drift / pixel-parity / negative-grep** 层；BDD 场景见 [bdd-plan](./bdd-plan.md) 与 [bdd-checklist](./bdd-checklist.md)。

测试目标按 phase 与 [plan §3.5 Coverage Matrix](./plan.md#35-coverage-matrix) 行映射；不引入硬编码代码覆盖率百分比作为 gate；如团队需要观测覆盖率，仅作背景指标。

## 2 Coverage Matrix（测试视角投射）

| 测试源 | Coverage Matrix 行 | Phase | 测试形态 | 文件 / 命令 |
|--------|-------------------|-------|----------|------------|
| PracticeScreen 静态壳源级复刻 | UI source structure parity (TopBar / SessionMap / QuestionCard / Transcript / InputBar / RightPanel / HintBanner / FinishCta) | Phase 1 | 单元 + DOM 锚点 | `pnpm --filter @easyinterview/frontend test --run practice/__tests__/PracticeScreen.test.tsx` |
| 路由壳替换 | Primary path · 文本面试 happy path（路由分支） | Phase 1 | 单元 | `pnpm --filter @easyinterview/frontend test --run app/__tests__/App.test.tsx`（既有，断言新增 practice case） |
| usePracticeSessionLoader 5 态 + auto refresh | Cross-layer contract · getPracticeSession refresh / resume；Failure · 404 sessionLost | Phase 1 + Phase 4 | 单元 + fake timer + jsdom event | `practice/__tests__/usePracticeSessionLoader.test.ts` |
| i18n zh/en namespace parity | UX · i18n zh/en | Phase 1 | 单元 | `pnpm --filter @easyinterview/frontend test --run i18n` |
| VoiceSurfaceComingSoon scoped 占位 + 不 import voice session surface | Scoped temporary placeholder · VoiceSurfaceComingSoon；UI stale-contract negative · voice imports | Phase 1 + Phase 3 | 单元 + 静态 grep | `practice/__tests__/PracticeScreen.test.tsx` + `pnpm --filter @easyinterview/frontend test --run practiceModeSwitch.test.tsx` + `grep -R "VoiceSessionSurface\\|PracticeWaveformBars\\|PracticeAnnotatedWaveform\\|VoiceExpressionPanel" frontend/src/app/screens/practice` 0 命中；text dictation failure banner 使用本地组件 |
| usePracticeEvents 5 kind body & header | Primary · answer_submitted；Cross-layer contract · PracticeSessionEventRequest schema | Phase 2 | 单元 | `practice/__tests__/usePracticeEvents.test.ts` |
| Idempotency 双轨边界 | Privacy · Idempotency-Key 双轨；Primary · appendSessionEvent 不带 Idempotency-Key；Primary · completePracticeSession 带 Idempotency-Key | Phase 2 + Phase 4 | 单元 + grep | `practice/__tests__/idempotencyContract.test.ts` + `grep -R "Idempotency-Key" frontend/src/app/screens/practice` 中只允许 useCompletePracticeSession 命中 |
| AssistantActionRenderer 5 type 渲染 + provenance 隔离 | Cross-layer contract · AssistantAction 5 type | Phase 2 | 单元 | `practice/__tests__/AssistantActionRenderer.test.tsx` |
| usePracticeSession 七个 status 分支 | Cross-layer contract · SessionStatus 七值消费 | Phase 2 + Phase 4 | 单元 | `practice/__tests__/usePracticeSession.test.ts` |
| appendSessionEvent body schema parity | Cross-layer contract · PracticeSessionEventRequest schema | Phase 2 | 单元 + contract | `practice/__tests__/appendSessionEventBody.test.ts` + `make validate-fixtures` |
| usePracticeAssistance strict / assisted × baseline / debrief | Alternate · assisted/strict 显隐；Alternate · goal=debrief 不改变显隐 | Phase 3 | 单元 + 显隐快照 | `practice/__tests__/usePracticeAssistance.test.ts` + `practiceGoalParity.test.tsx` |
| hint / skip / pause-resume 流 | Alternate · hint_requested；Alternate · turn_skipped；Alternate · session_paused/resumed；Boundary · paused 状态禁用 | Phase 3 | 单元 + integration | `practice/__tests__/practiceHints.test.tsx` + `practiceSkip.test.tsx` + `practicePauseResume.test.tsx` |
| RoleDropdown UI-only | Cross-layer contract · interviewerPersona UI-only | Phase 3 | 单元 + 调用次数反查 | `practice/__tests__/RoleDropdown.test.tsx` |
| SessionMap turn 历史 | UI source structure parity · SessionMap；Boundary · 空 transcript | Phase 3 | 单元 | `practice/__tests__/SessionMap.test.tsx` + `Transcript.test.tsx`（empty 用例） |
| 模式切换 segmented control + strict toggle 锁定 | Alternate · 模式切换；Risk · strict toggle 锁定 | Phase 3 | 单元 + a11y | `practice/__tests__/practiceModeSwitch.test.tsx` + `practiceStrictToggleLocked.test.tsx` |
| useCompletePracticeSession happy / replay / mismatch / network-5xx / StrictMode | Primary · completePracticeSession；Cross-layer · CompletePracticeSessionRequest schema；Failure · complete 5xx / 409 mismatch | Phase 4 | 单元 + fake timer + StrictMode | `practice/__tests__/useCompletePracticeSession.test.ts` |
| practiceHandoffParams 完整字段 + 不含展示字段 | Cross-layer · generating handoff 参数；Privacy · body 不含展示字段 | Phase 4 | 单元 | `practice/__tests__/practiceHandoff.test.ts` |
| completePracticeSession body parity | Cross-layer · CompletePracticeSessionRequest schema | Phase 4 | 单元 + contract | `practice/__tests__/completePracticeSessionBody.test.ts` |
| 错误映射（append 502 / complete 5xx / 404 / 409 strict / 409 mismatch / 网络） | Failure · AI 502；Failure · complete 5xx；Failure · 404；Failure · 409 conflict；Failure · 409 mismatch；Failure · network | Phase 4 | 单元 + fixture variant | `practice/__tests__/practiceErrors.test.tsx` + `practiceSessionLost.test.tsx` + `practiceConflict.test.tsx` + `practiceClientEventConflict.test.tsx` |
| INCREMENT_HINT_COUNT reducer action | InterviewContext × PracticeDisplayContext mapping | Phase 4 | 单元 | `interview-context/InterviewContext.test.tsx`（在 001 测试文件追加） |
| Privacy redaction（answerText / questionText / hint / provenance） | Privacy · raw text；Privacy · provenance | Phase 4 + Phase 5 | 单元 + scenario verify | `practice/__tests__/practicePrivacy.test.tsx` + scenario verify.sh grep |
| Pixel parity practice.spec.ts | UI visual geometry parity · desktop / mobile / dark / customAccent | Phase 5 | Playwright | `pnpm --filter @easyinterview/frontend test:pixel-parity --grep practice` |
| Negative grep · prototype imports / 旧 testid / 旧 route / 旧 enum / voice imports / getFeedbackReport / createPracticeVoiceTurn / Idempotency-Key appendSessionEvent | UI stale-contract negative；Regression · 不直接调用 LLM | Phase 5 | grep gate (Vitest + scenario verify.sh) | `practice/__tests__/legacyNegative.test.ts` + `test/scenarios/e2e/p0-043.../scripts/verify.sh` + CI grep |
| Fixture parity / drift | Cross-layer contract · openapi fixture variants | Phase 2 + Phase 4 + Phase 5 | drift + contract | `make validate-fixtures` + `make codegen-check` + `python3 scripts/lint/conventions_drift.py --repo-root .` |
| Regression rerun（workspace + backend-practice contract gate） | Regression · workspace + 后端契约 | Phase 5 | scenario + Vitest | `test/scenarios/e2e/p0-018-021` + `p0-022-026` + 全量 Vitest |

## 3 Phase 1: PracticeScreen 静态壳 + 路由替换 + i18n + sessionId 守卫

| 任务 | 测试文件 / 命令 | 预期 Red/Green 证据 |
|------|----------------|---------------------|
| PracticeScreen 静态壳源级复刻 + ≥ 20 个 testid + 控件类型断言 | `practice/__tests__/PracticeScreen.test.tsx` | Red: testid 缺失 / `<select>` 用作 segmented mode / RoleDropdown 是 `<select>`；Green: ≥ 20 个 `practice-*` testid 命中 + segmented mode 是 `<button>` 列表 + RoleDropdown 展开后是 menu hierarchy（`role="menu"` / `role="menuitem"`） + strict toggle role='switch' aria-checked 切换 |
| usePracticeSessionLoader 五态 | `practice/__tests__/usePracticeSessionLoader.test.ts` | Red: idle / loading / data / sessionLost / error 任一态错；Green: 五态分支 + sessionId 缺失立即 sessionLost（不发请求） + MERGE_SESSION 调用次数 1 |
| usePracticeSessionLoader auto refresh | 同上（`TestVisibilityRefresh` / `TestFocusRefresh` / `TestOnlineRefresh`） | Red: 切换 visibility 不触发 refresh；Green: 三个事件各自触发一次 `getPracticeSession` 调用 |
| 路由壳替换 | `app/__tests__/App.test.tsx`（在 001 文件追加 case） | Red: `practice` 命中 PlaceholderScreen；Green: `practice` 命中 PracticeScreen，`generating` / `report` / `company_intel` 仍命中 PlaceholderScreen |
| i18n practice.* namespace parity | `pnpm --filter @easyinterview/frontend test --run i18n` | Red: zh / en namespace 有缺漏；Green: ≥ 40 key 双语对齐 |
| VoiceSurfaceComingSoon 占位 + 不 import voice 组件 | `practice/__tests__/PracticeScreen.test.tsx` + `practiceModeSwitch.test.tsx` + grep | Red: voice surface DOM 出现 / voice 组件 import 存在；Green: 占位卡渲染 + 点击「返回 text」回到 text + grep 0 命中 |

## 4 Phase 2: appendSessionEvent + AssistantAction + SessionStatus 消费

| 任务 | 测试文件 / 命令 | 预期 Red/Green 证据 |
|------|----------------|---------------------|
| usePracticeEvents 5 kind body & header | `practice/__tests__/usePracticeEvents.test.ts` | Red: kind 缺失 / body 不一致 / header 含 Idempotency-Key；Green: 5 kind body 与 OpenAPI schema 一致 + fetch init 反向断言 Idempotency-Key 0 命中 + clientEventId 来自 uuidv7 |
| usePracticeEvents retry 复用 + fresh action | 同上（`TestRetryReusesClientEventId` / `TestFreshActionAssignsNewClientEventId`） | Red: retry 生成新 id / fresh action 复用旧 id；Green: retry 复用 + fresh 切换 |
| Idempotency 双轨 | `practice/__tests__/idempotencyContract.test.ts` | Red: appendSessionEvent 含 Idempotency-Key / completePracticeSession 缺；Green: 双向断言通过 |
| AssistantActionRenderer 5 type | `practice/__tests__/AssistantActionRenderer.test.tsx` | Red: ask_question / ask_follow_up / show_hint / session_wait / session_completed 任一分支错；Green: 5 分支渲染正确 + provenance 仅出现在 AI transparency 卡 |
| usePracticeSession 七个 status 分支 | `practice/__tests__/usePracticeSession.test.ts` | Red: queued/completing 不禁用 / completed 不防抖 / failed/cancelled 不返回 workspace；Green: 七分支 + completed 防抖 nav generating + `draft/archived` 旧值 0 命中 |
| appendSessionEvent body schema parity | `practice/__tests__/appendSessionEventBody.test.ts` + `make validate-fixtures` | Red: payload 按 kind 类型化错 / fixture variant 缺；Green: canonical 命名 scenarios + 类型化 body |
| fixture variants（appendSessionEvent + getPracticeSession） | `make validate-fixtures` + `make codegen-check` + `python3 scripts/lint/conventions_drift.py --repo-root .` | Red: 命名场景缺失 / drift；Green: appendSessionEvent canonical variants（`default/follow-up/hint-strict-conflict/turn-skipped/pause-resume/replay/mismatch/completed`）+ getPracticeSession `default/prototype-baseline/missing-session/running-with-history/queued/completing` 齐全 + 0 drift |

## 5 Phase 3: assisted / strict 显隐 + RoleDropdown + 提示 / 跳过 / 暂停-恢复

| 任务 | 测试文件 / 命令 | 预期 Red/Green 证据 |
|------|----------------|---------------------|
| usePracticeAssistance 显隐 | `practice/__tests__/usePracticeAssistance.test.ts` | Red: practiceGoal 参与计算；Green: 仅依赖 practiceMode；4 组合（strict / assisted × baseline / debrief）显隐一致 |
| practiceGoalParity 4 组合显隐 | `practice/__tests__/practiceGoalParity.test.tsx` | Red: debrief 改变显隐；Green: assisted+baseline / assisted+debrief / strict+baseline / strict+debrief 显隐快照一致 |
| Hint flow（assisted + 200） | `practice/__tests__/practiceHints.test.tsx` | Red: hintCount 未自增 / strict 模式 hint button DOM 渲染；Green: hint 200 后 hintCount++ + hintUsed='true' + HintBanner 显示；strict 模式 hint button DOM 不存在 |
| Skip flow | `practice/__tests__/practiceSkip.test.tsx` | Red: SessionMap turn 状态未标记 skipped；Green: skip → API → SessionMap 标记 skipped + renderer 推进 ask_question |
| Pause / Resume flow | `practice/__tests__/practicePauseResume.test.tsx` | Red: 暂停期间按钮可用；Green: pause → API + timer 停止 + 三按钮 disabled；resume → API + timer 恢复 + 解禁 |
| RoleDropdown UI-only | `practice/__tests__/RoleDropdown.test.tsx` | Red: dropdown 切换发请求；Green: generated client 调用次数 = 0 + AI transparency role label 切换 |
| SessionMap turn 历史 | `practice/__tests__/SessionMap.test.tsx` | Red: turn 状态未渲染；Green: done/active/pending/skipped/follow_up_requested 五态渲染 |
| Mode 切换 segmented control | `practice/__tests__/practiceModeSwitch.test.tsx` | Red: voice 渲染真实 surface；Green: voice 渲染 ComingSoon 占位 + 「返回 text」回到 text |
| Strict toggle 运行时锁定 | `practice/__tests__/practiceStrictToggleLocked.test.tsx` | Red: click 改 backend；Green: click 触发 toast + 不调 backend |

## 6 Phase 4: completePracticeSession + handoff + 错误恢复 + sessionLost / conflict

| 任务 | 测试文件 / 命令 | 预期 Red/Green 证据 |
|------|----------------|---------------------|
| useCompletePracticeSession happy + replay + StrictMode | `practice/__tests__/useCompletePracticeSession.test.ts` | Red: replay 重复 nav / StrictMode 多次调用；Green: 202 后 nav 一次 + replay 同 key 二次返回原值 + StrictMode 调用 1 次 |
| useCompletePracticeSession mismatch / retry | 同上 | Red: 409 mismatch 不显示 / retry 不复用 key；Green: 错误码映射正确 + retry 复用 Idempotency-Key |
| practiceHandoffParams 字段完整性 | `practice/__tests__/practiceHandoff.test.ts` | Red: 缺 reportId / 缺稳定 InterviewContext ID / raw text 进 URL；Green: 字段集完整正确 + nav 路径 generating + body 不含展示字段 |
| completePracticeSession body parity | `practice/__tests__/completePracticeSessionBody.test.ts` + `make validate-fixtures` | Red: body 含展示字段；Green: body 仅 `{clientCompletedAt}` + canonical 命名 scenarios (`default`/`replay`/`mismatch`/`session-already-completed`/`cross-user-not-found`) |
| sessionLost 兜底 | `practice/__tests__/practiceSessionLost.test.tsx` | Red: 404 不渲染 PracticeSessionLostState；Green: 404 → 渲染 + CTA 返回 workspace + InterviewContext workspace 上下文保留 |
| client_event_fingerprint_mismatch | `practice/__tests__/practiceClientEventConflict.test.tsx` | Red: race 未锁定；Green: 409 mismatch → refresh + UI 锁定 → server wins 重置 |
| INCREMENT_HINT_COUNT reducer | `interview-context/InterviewContext.test.tsx`（在 001 文件追加） | Red: action 未实现 / hintUsed 未更新；Green: hintCount 字符串 → 数字自增 + hintUsed='true' + 001 已有 action 不受影响 |
| Privacy redaction | `practice/__tests__/practicePrivacy.test.tsx` + scenario verify.sh grep | Red: answerText / questionText / hint / provenance modelId 出现在 URL / localStorage / console / telemetry；Green: 全部 0 命中 |
| Practice completion flow | `practice/__tests__/practiceCompletion.test.tsx` | Red: session_completed 不触发 CTA 高亮；Green: 高亮 + auto-scroll + 点击 CTA 完成 |

## 7 Phase 5: Pixel parity + Scenario + Regression + Negative grep

| 任务 | 测试文件 / 命令 | 预期 Red/Green 证据 |
|------|----------------|---------------------|
| Pixel parity practice.spec.ts | `pnpm --filter @easyinterview/frontend test:pixel-parity --grep practice` | Red: bounding box 溢出 / 主题切换不可见；Green: desktop + mobile + 8 主题 × dark + customAccent + 5 状态截图基线全 PASS |
| Scenario 资产 + INDEX 更新 | `test/scenarios/e2e/p0-042-045/` 目录 + `INDEX.md` | Red: 目录或脚本缺失；Green: 4 目录齐全 + INDEX 4 行追加 |
| Regression rerun（workspace + backend-practice contract） | scenario rerun + 全量 Vitest + build | Red: workspace P0.018-021 任一 FAIL / backend-practice P0.022-026 任一 FAIL；Green: 9 个 regression scenario 全 PASS + Vitest 全 PASS + build PASS |
| Negative grep | `practice/__tests__/legacyNegative.test.ts` + CI grep | Red: 任何旧口径 / voice imports / getFeedbackReport / createPracticeVoiceTurn / `Idempotency-Key.*appendSessionEvent` / raw text 泄漏命中；Green: 全部 0 命中 |
| 文档与索引同步 | `make docs-check` + `/sync-doc-index --fix-index` + `check-md-links` | Red: drift / 链接断；Green: 0 drift + 0 broken |
