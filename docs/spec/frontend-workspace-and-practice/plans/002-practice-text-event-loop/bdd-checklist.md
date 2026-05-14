# 002 BDD Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.044 文本面试 assisted happy path

- [ ] 创建场景目录 `test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/`，含 `README.md`（§6 baseline + §7 离线限制）+ `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture variant：`getPracticeSession.json` 至少 1 个 variant（`default`）+ 1 个 `running-with-history`；`appendSessionEvent.json` 至少 3 个 variant（`follow-up` answer→ask_follow_up / `default` answer→ask_question，若当前 default 仍为 ask_follow_up 则 Phase 2 先修订 fixture truth / `turn-skipped`）；按 `mock-contract-suite` 规则配置；通过 fixture/schema 验证（`make validate-fixtures`）
- [ ] 实现 `scripts/setup.sh`（fixture variant + InterviewContext route params + signed-in 状态）/ `scripts/trigger.sh`（运行 PracticeScreen mount + usePracticeSessionLoader + usePracticeEvents + AssistantActionRenderer 覆盖、Vitest 主路径文件 + Playwright pixel parity practice.spec.ts assisted-happy 子用例）/ `scripts/verify.sh`（断言 ≥ 20 个 runtime testid 命中、`appendSessionEvent` 请求 init 反向 grep `Idempotency-Key` 0 命中、`clientEventId` 是 UUIDv7、Transcript ≥ 4 条消息含 Follow-up Tag、负向 grep voice imports / 旧 prototype / `getFeedbackReport` / `createPracticeVoiceTurn` 0 命中）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-044-practice-text-loop-assisted-happy-path/trigger.log` + verify 输出 + pixel parity assisted-happy baseline + retired-testid grep 0 命中
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.044 行（关联需求 `frontend-workspace-and-practice C-4, C-8, C-9`，状态 Ready，automated）

## E2E.P0.045 strict / assisted × baseline / debrief 显隐 + hint / skip / pause-resume + 旧口径负向

- [ ] 创建场景目录 `test/scenarios/e2e/p0-045-practice-text-loop-strict-and-debrief-display/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture variant：`getPracticeSession.json` 至少 1 个 `default` variant；`appendSessionEvent.json` 至少 6 个 variant（`default` answer→ask_question / `show-hint`（fixture-only until backend-practice/003，hint_requested → 200 + assistantAction.type='show_hint'）/ `hint-strict-conflict`（当前真实 backend-practice/002 409 PRACTICE_SESSION_CONFLICT detail.policy='hint_disabled_in_mode'）/ `turn-skipped` / `pause-resume` / `follow-up`）；4 路由组合通过 setup 切换 `practiceMode` × `practiceGoal`
- [ ] 实现 `scripts/setup.sh`（4 路由组合切换 + fixture 切换 + signed-in 状态）/ `scripts/trigger.sh`（运行 usePracticeAssistance + practiceHints / practiceSkip / practicePauseResume / practiceModeSwitch / practiceStrictToggleLocked / practiceGoalParity 覆盖、Vitest 显隐文件、Playwright pixel parity strict / assisted 子用例）/ `scripts/verify.sh`（断言：assisted LIVE NOTES / hint button / experience cards 渲染 + hintCount 自增；strict 三者 DOM 不存在 + strict-mode banner 渲染；assisted + debrief 与 assisted + baseline 显隐快照一致；旧口径负向 grep —— `practiceMode='debrief'` / `切到语音` / voice imports / 旧 testid / `Idempotency-Key.*appendSessionEvent` / 独立 voice route 全部 0 命中；strict toggle 点击触发 toast + 0 backend 调用）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-045-practice-text-loop-strict-and-debrief-display/trigger.log` + verify 输出 + 显隐快照 diff 0 + 旧口径负向 grep 日志
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.045 行（关联需求 `frontend-workspace-and-practice C-4, C-10, C-12`，状态 Ready，automated）

## E2E.P0.046 失败恢复 · AI 502 + session 404 + 409 mismatch + 409 strict conflict + retry 复用

- [ ] 创建场景目录 `test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture variant：`getPracticeSession.json` 至少 3 个 variant（`default` / `missing-session`（404）/ 计划新增 `network-fail` 用于 5xx 流）；`appendSessionEvent.json` 至少 4 个 variant（`default` / `ai-timeout` / `mismatch`（client_event_fingerprint_mismatch）/ `hint-strict-conflict`）
- [ ] 实现 `scripts/setup.sh`（5 子场景 fixture 切换）/ `scripts/trigger.sh`（运行 practiceErrors / practiceSessionLost / practiceClientEventConflict / practiceConflict / usePracticeEvents retry 覆盖）/ `scripts/verify.sh`（断言：AI 502 inline error + retry 按钮 + 同 clientEventId 复用；404 → PracticeSessionLostState + nav workspace 携带 targetJobId/jdId/planId/resumeVersionId；409 mismatch → refresh + 锁定 UI + server wins；409 strict → InlineWarning 不重试；3 次失败 → 「返回 workspace」fallback；raw answerText / questionText / hint / provenance modelId 不出现在 console / URL / localStorage / telemetry；AI provider key / prompt registry / AIClient / LLM endpoint 0 命中）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-046-practice-text-loop-failure-and-recovery/trigger.log` + verify 输出 + 错误流截图基线 + clientEventId reuse / refresh race 测试断言日志
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.046 行（关联需求 `frontend-workspace-and-practice C-4, C-12`，状态 Ready，automated）

## E2E.P0.047 completePracticeSession 202 + handoff generating + Idempotency-Key replay + 隐私红线

- [ ] 创建场景目录 `test/scenarios/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture variant：`getPracticeSession.json` 至少 2 个 variant（`default` running / `completing`）；`appendSessionEvent.json` `completed`（assistantAction.type='session_completed'）；`completePracticeSession.json` 至少 4 个 variant（`default` 202 / `replay` 同 key 二次返回首次 response / `mismatch` 同 key 不同 fingerprint / `session-already-completed` 反向查既有 report）
- [ ] 实现 `scripts/setup.sh`（4 子场景 fixture 切换 + InterviewContext hintCount='2' 用于显隐验证）/ `scripts/trigger.sh`（运行 useCompletePracticeSession / practiceHandoff / completePracticeSessionBody / practiceCompletion / practicePrivacy 覆盖、Playwright pixel parity completing / completed 子用例）/ `scripts/verify.sh`（断言：CTA 点击调 `completePracticeSession` body 仅 `{clientCompletedAt}` 且不含展示字段；`Idempotency-Key` header 存在且与 retry 复用；202 后 nav `generating` 携带稳定 InterviewContext ID（planId / targetJobId / jdId / resumeVersionId / roundId / sessionId / reportId）+ PracticeDisplayContext（mode / modality / practiceMode / practiceGoal / hintUsed / hintCount）；URL query string 允许稳定 ID 与显示上下文，不含 raw text / promptHash / modelId；localStorage / console / telemetry 不含 raw text；StrictMode 双触发 generated client 调用次数 = 1 + nav 调用次数 = 1；`getFeedbackReport` / `createPracticeVoiceTurn` 调用 0 命中；旧 testid / 旧 route alias / `Idempotency-Key.*appendSessionEvent` 0 命中）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/trigger.log` + verify 输出 + handoff nav params 断言日志 + 隐私红线 grep 0 命中日志 + Idempotency replay 行为日志
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.047 行（关联需求 `frontend-workspace-and-practice C-4, C-6, C-12`，状态 Ready，automated）

## 整体 Regression（Phase 5 收口）

- [ ] workspace regression：`E2E.P0.018 / 019 / 020 / 021` 全部 `setup → trigger → verify → cleanup` PASS（确认 001 plan 交付的 workspace 行为不被 002 改动破坏）
- [ ] backend-practice 契约 regression：`E2E.P0.022 / 023 / 024 / 025 / 026` 全部 PASS；backend-practice 002 `E2E.P0.038 / 039 / 040 / 041 / 042 / 043` 真实 Go HTTP scenario 全部 PASS（执行入口 `cd backend && go test ./cmd/api -run 'TestE2EP0038|TestE2EP0039|TestE2EP0040|TestE2EP0041|TestE2EP0042|TestE2EP0043' -count=1`）；本 plan 在 fixture-backed 和真实 handler 边界下均满足契约一致
- [ ] `pnpm --filter @easyinterview/frontend test` 全量 Vitest PASS（含本 plan 新增测试文件）
- [ ] `pnpm --filter @easyinterview/frontend test:pixel-parity` 累加 practice spec 全 PASS（在 D2/D3 + home plan + workspace plan 现有基础上）
- [ ] `pnpm --filter @easyinterview/frontend build`（含 `tsc --noEmit` + `vite build`）+ `make build` PASS
- [ ] 文档与索引收口：本 checklist、bdd-checklist、test-checklist 与 plans INDEX 已同步；`make docs-check` / `/sync-doc-index` 作为 post-fix drift gate 执行；`check-md-links` OK；history.md 追加 plan 002 启动条目
