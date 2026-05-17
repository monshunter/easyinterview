# E2E 场景索引

> 场景按阶段分组，记录编号、关联需求、目录路径和状态。

---

## P0 核心闭环

| 场景 ID | 关联需求 | 目录 | 描述 | 执行方式 | 状态 |
|---------|----------|------|------|----------|------|
| E2E.P0.001 | frontend-shell C-1 | `p0-001-default-home-shell/` | 默认进入首页并呈现五入口 TopBar 与用户菜单 | automated | Ready |
| E2E.P0.002 | frontend-shell C-2 | `p0-002-auth-pending-action-resume/` | 登录打断后恢复原业务动作与上下文 | automated | Ready |
| E2E.P0.003 | backend-auth C-1 | `p0-003-passwordless-session-cookie/` | 邮箱挑战验证后签发 first-party session 并支持 /me 与 logout | automated | Ready |
| E2E.P0.004 | frontend-shell C-7 | `p0-004-app-shell-language-switch/` | App Shell 中英语言切换并携带 Accept-Language display hint | automated | Ready |
| E2E.P0.005 | frontend-shell C-8 | `p0-005-app-shell-visual-system-smoke/` | D2 视觉系统 smoke：DOM/className/CSS-variable/customAccent overlay/legacy 负向 + ui-design 源追溯 | automated | Ready |
| E2E.P0.006 | frontend-shell C-9 | `p0-006-ui-design-pixel-parity-gate/` | D2 follow-up Playwright + chromium pixel parity：desktop+mobile viewport DOM/computed style/bounding box/screenshot regression + dark/customAccent 状态 | automated | Ready |
| E2E.P0.007 | practice-voice-mvp C-1/C-2/C-3/C-4/C-5 | `p0-007-cascaded-voice-turn/` | 完整语音 turn：前端提交 voice audio，后端级联 STT/chat/TTS，返回 transcript、TTS chunk、provider meta 与幂等 replay | automated | Ready |
| E2E.P0.008 | practice-voice-mvp C-3/C-5 | `p0-008-voice-barge-in-committed-context/` | 语音插话：前端停止 TTS 并上报 barge-in，后端只提交已播放上下文并阻止未播放 draft 进入 prompt | automated | Ready |
| E2E.P0.009 | practice-voice-mvp C-6/C-7/C-8/C-9 | `p0-009-voice-provider-failure-fallback/` | Provider failure fallback：STT fail-fast、chat/TTS 隔离、TTS 文本 fallback、realtime/stub 负向与 privacy gate | automated | Ready |
| E2E.P0.010 | backend-targetjob C-1/C-3/C-6/C-7/C-12 | `p0-010-targetjob-text-import-parse-ready/` | manual_text TargetJob 导入、异步解析、列表、详情、更新与 idempotency | automated | Ready |
| E2E.P0.011 | backend-targetjob C-2/C-3/C-9 | `p0-011-targetjob-url-import-fetch-and-parse/` | URL TargetJob 导入、SSRF 守护、抓取 snapshot、解析与 source_refresh 占位 | automated | Ready |
| E2E.P0.012 | backend-targetjob C-4/C-5/C-10 | `p0-012-targetjob-parse-failure-retryable/` | TargetJob parse 失败 retryable / non-retryable 语义与隐私红线 | automated | Ready |
| E2E.P0.013 | backend-targetjob C-3/C-6/C-9/C-11/C-13 | `p0-013-targetjob-manual-form-ready/` | manual_form TargetJob 同步 ready、terminal job、列表详情与 no-runner 断言 | automated | Ready |
| E2E.P0.014 | frontend-home-job-picks-and-parse C-1, C-4 | `p0-014-home-default-render/` | Home 默认渲染：empty/non-empty/12+ 三态，DOM 锚点，排序，TopBar 高亮 | automated | Ready |
| E2E.P0.015 | frontend-home-job-picks-and-parse C-2, C-3, C-6 | `p0-015-jd-import-and-parse/` | Paste/Upload/URL → import → parse loading → preview 主路径 + failed | automated | Ready |
| E2E.P0.016 | frontend-home-job-picks-and-parse C-5, C-7 | `p0-016-parse-confirm-to-workspace/` | Parse 编辑 + Confirm → workspace + auth pending action | automated | Ready |
| E2E.P0.017 | frontend-home-job-picks-and-parse C-8 | `p0-017-jd-match-placeholder/` | jd_match P1 placeholder shell smoke：hero/tabs/placeholder + legacy negative | automated | Ready |
| E2E.P0.018 | frontend-workspace-and-practice C-2, C-7, C-8, C-9 | `p0-018-workspace-default-render/` | Workspace 默认渲染：plan eyebrow + header + Interview Launcher + Main Left/Right + Modals | automated | Ready |
| E2E.P0.019 | frontend-workspace-and-practice C-2, C-3, C-8, C-9 | `p0-019-workspace-context-loading/` | Workspace context loading：empty/missing-resume 空态 + getPracticePlan refresh | automated | Ready |
| E2E.P0.020 | frontend-workspace-and-practice C-1, C-3, C-12 | `p0-020-workspace-start-practice/` | 立即面试 双步契约 + Idempotency-Key + pendingAction 未登录恢复 | automated | Ready |
| E2E.P0.021 | frontend-workspace-and-practice C-7, C-9, C-10, C-12 | `p0-021-workspace-handoff/` | Workspace handoff + 隐私红线 + legacy negative grep | automated | Ready |
| E2E.P0.022 | backend-practice C-1, C-13 | `p0-022-practice-plan-baseline-create-and-read/` | createPracticePlan baseline、idempotency replay、getPracticePlan 与 cross-user 404 隔离 | automated | Ready |
| E2E.P0.023 | backend-practice C-4 | `p0-023-practice-session-start-and-first-question/` | startPracticeSession 同步首题、getPracticeSession、session_started event 与 practice.session.started outbox | automated | Ready |
| E2E.P0.024 | backend-practice C-5, C-21, C-23 | `p0-024-practice-session-ai-failure-retry/` | AI timeout 后 failed_retryable reservation，同 key 重试成功且 outbox 仅一次 | automated | Ready |
| E2E.P0.025 | backend-practice C-10, C-13, C-22, C-23, C-24, C-25 | `p0-025-practice-idempotency-and-isolation-matrix/` | startPracticeSession replay / mismatch / 跨用户隔离 / 同 plan 多 key conflict / cross-user 404 矩阵 | automated | Ready |
| E2E.P0.026 | backend-practice C-16, D-11 | `p0-026-practice-observability-and-privacy-redlines/` | observed AI typed columns、metric allowlist、隐私红线与 backend-practice legacy gate | automated | Ready |
| E2E.P0.027 | frontend-home-job-picks-and-parse C-12, C-13, C-15 | `p0-027-jd-match-recommended-and-confirm/` | jd_match Recommended tab 主路径 + 4 button 闭环 + auth pending action + 隐私反查 | automated | Ready |
| E2E.P0.028 | frontend-home-job-picks-and-parse C-14, C-15 | `p0-028-jd-match-search-and-saved/` | jd_match Search tab + Saved searches + 4 chip filter + 5 步 AGENT panel + failure + auth gate + privacy | automated | Ready |
| E2E.P0.029 | frontend-home-job-picks-and-parse C-16 | `p0-029-jd-match-watchlist-and-signals/` | jd_match Watchlist tab + Market signals + chevron handoff + boundary + privacy | automated | Ready |
| E2E.P0.030 | frontend-home-job-picks-and-parse C-12, C-15 | `p0-030-jd-match-profile-and-agent-status/` | jd_match Profile chip + AGENT scan status + Auth pending action 跨 tab 综合 | automated | Ready |
| E2E.P0.031 | frontend-home-job-picks-and-parse C-13 | `p0-031-jd-match-confirm-interview-handoff/` | Confirm interview from jd_match → parse 出口 params 完整性 + parse 屏不破坏 | automated | Ready |
| E2E.P0.032 | frontend-shell C-10 | `p0-032-dev-mock-auth-state-and-user-menu/` | Dev mock 默认非登录、登录后头像 dropdown、profile/settings 分流与 logout 后非登录闭环 | automated | Ready |
| E2E.P0.033 | backend-upload C-1, C-2, C-3, C-4, C-6, C-7, C-8 | `p0-033-file-presign-register-roundtrip/` | file presign、IK replay、register 校验、cross-user 隔离与 privacy delete tombstone | automated | Ready |
| E2E.P0.034 | backend-resume C-1, C-2, C-5, C-6, C-7, C-8 | `p0-034-resume-register-and-list/` | Resume register/get/list：三 sourceType、IK replay、upload handoff、cross-user 404、cursor pagination 与 fixture parity | automated | Ready |
| E2E.P0.035 | backend-resume C-3, C-4, C-13 | `p0-035-resume-parse-async-job-lifecycle/` | resume.parse in-process drainer：AI parse、ready/failed 状态、typed task run、ready-only outbox 与 privacy redlines | automated | Ready |
| E2E.P0.036 | frontend-resume-workshop C-1, C-2, C-3, C-5, C-6, C-7, C-8, C-9 | `p0-036-resume-list-tree-flat-toggle/` | Resume Workshop list：路由替换、auth gate、StatsStrip 从 fixture 派生计数、ViewSwitcher tree/flat、第二个 asset 的 no-versions 占位与 retired-route negative grep | automated | Ready |
| E2E.P0.037 | frontend-resume-workshop C-4, C-5, C-6, C-7, C-8 | `p0-037-resume-detail-preview-readonly/` | Resume Workshop detail：Preview Tab 投影、原件 modal a11y、Export PDF Idempotency-Key + P0 501 toast、404 fallback 不回显 fixture error.code | automated | Ready |
| E2E.P0.044 | frontend-workspace-and-practice C-4, C-8, C-9 | `p0-044-practice-text-loop-assisted-happy-path/` | Practice text loop assisted happy path：PracticeScreen + appendSessionEvent default + ask_follow_up renderer + IK 双轨边界 | automated | Ready |
| E2E.P0.045 | frontend-workspace-and-practice C-4, C-10, C-12 | `p0-045-practice-text-loop-strict-and-debrief-display/` | Practice text loop 显隐：strict / assisted × baseline / debrief 4 组合 + hint / skip / pause-resume + strict toggle 锁定 toast + 旧口径负向 | automated | Ready |
| E2E.P0.046 | frontend-workspace-and-practice C-4, C-12 | `p0-046-practice-text-loop-failure-and-recovery/` | Practice text loop 失败恢复：AI 502 + session 404 + 409 mismatch + 409 strict-hint + retry 复用 IK | automated | Ready |
| E2E.P0.047 | frontend-workspace-and-practice C-4, C-6, C-12 | `p0-047-practice-text-loop-complete-and-generating-handoff/` | Practice text loop 完成 handoff：completePracticeSession 202 + IK + handoff generating + 隐私红线 | automated | Ready |
| E2E.P0.048 | backend-practice C-7, C-8b, C-12 | `p0-048-practice-hint-assisted-across-goals/` | assisted hint 主路径：4 个 goal 下返回 show_hint、写 hint_generate task run 且不推进 turn lifecycle | automated | Ready |
| E2E.P0.049 | backend-practice C-8, C-8b, D-38 | `p0-049-practice-hint-strict-refusal/` | strict hint 拒绝：4 个 goal 下返回 409 hint_disabled_in_mode，same clientEventId replay 不留下 pending reservation | automated | Ready |
| E2E.P0.050 | backend-practice C-12, D-37 | `p0-050-practice-hint-provenance-task-runs/` | AssistantAction wire provenance 仅 6 字段，runtime feature_key 只进入 ai_task_runs typed columns | automated | Ready |
| E2E.P0.051 | backend-practice C-16, C-17, D-36 | `p0-051-practice-hint-degrade-privacy/` | hint AI graceful degrade：200 session_wait、session running、failed task run 与隐私红线/legacy negative | automated | Ready |
| E2E.P0.056 | frontend-report-dashboard C-1, C-2, C-5, C-8, C-11 | `p0-056-generating-to-report-happy-path/` | Generating → Report happy path：5-phase poll → ReportDashboard mount + 5 detail tabs + ContextStrip + testid coverage + read-only contract | automated | Ready |
| E2E.P0.057 | frontend-report-dashboard C-3, C-6 | `p0-057-replay-cta-paths-a-and-b/` | Replay CTA paths A/B：retry_current_round + next_round payload, replay_practice pendingAction round-trip, no raw text | automated | Ready |
| E2E.P0.058 | frontend-report-dashboard C-4, C-12 | `p0-058-report-failure-and-missing-session/` | ReportFailureState AI_* enum + ReportMissingSessionState + cross-user 404 not-found copy + Generating timeout retry | automated | Ready |
| E2E.P0.059 | frontend-report-dashboard C-13, C-14, D-12 | `p0-059-report-pixel-parity-i18n-and-legacy-negative/` | i18n namespace sync + AI_* enum coverage + legacy vocab negative grep + Playwright pixel-parity specs staged | automated | Ready |
| E2E.P0.060 | backend-debrief C-1, C-2, C-3, C-5 | `p0-060-debrief-create-worker-happy/` | createDebrief draft + queued worker + worker completed happy path + outbox/task-run privacy | automated | Ready |
| E2E.P0.061 | backend-debrief C-6, C-7, C-8 | `p0-061-debrief-get-isolation/` | getDebrief draft/completed 双态与 cross-user/not-found 404 隔离 | automated | Ready |
| E2E.P0.062 | backend-debrief C-11, C-12 | `p0-062-debrief-worker-retry-failure/` | debrief_generate F3/A3/parse failure、retry backoff 与 max-attempt permanent failure | automated | Ready |
| E2E.P0.063 | backend-debrief C-9, C-10 | `p0-063-debrief-suggest-questions/` | suggestDebriefQuestions 成功、count boundary、AI failure 与 task/audit 写入 | automated | Ready |
| E2E.P0.064 | backend-debrief C-14, C-15 | `p0-064-debrief-privacy-legacy/` | debrief outbox/audit/task privacy marker 与 retired vocabulary negative lint | automated | Ready |
| E2E.P0.065 | frontend-debrief C-1, C-2, C-3, C-11, C-14 | `p0-065-debrief-default-render-and-pickers/` | DebriefScreen 默认渲染 + 3 picker modal + route normalize debrief_full | automated | Ready |
| E2E.P0.066 | frontend-debrief C-4, C-5, C-7 | `p0-066-debrief-text-suggestions-and-submit/` | text-mode AI suggestions + 4 CTA entries + createDebrief 提交 + IK | automated | Ready |
| E2E.P0.067 | frontend-debrief C-8, C-12 | `p0-067-debrief-polling-happy-and-analysis/` | dual-track polling getJob + getDebrief + Step 1 analysis 渲染 | automated | Ready |
| E2E.P0.068 | frontend-debrief C-9, C-10, C-13 | `p0-068-debrief-failure-and-handoff/` | failure / missing / timeout 状态 + Step 2 launcher nav 复盘面试 handoff | automated | Ready |
| E2E.P0.069 | frontend-debrief C-15, C-16, C-17, C-18 | `p0-069-debrief-pixel-parity-and-legacy-negative/` | i18n 覆盖 + 隐私边界 + debrief Playwright pixel parity + legacy-negative gate | automated | Ready |
| E2E.P0.070 | backend-practice C-2, C-3 | `p0-070-practice-derived-plan-create-read-replay/` | report/debrief derived createPracticePlan、getPracticePlan 与 idempotency replay source 字段 | automated | Ready |
| E2E.P0.071 | backend-practice C-3, C-4 | `p0-071-practice-debrief-start-source-question/` | debrief startPracticeSession 使用 source raw_questions 首题且 first_question AI 零调用 | automated | Ready |
| E2E.P0.072 | backend-practice C-2, C-3, D-11 | `p0-072-practice-derived-source-isolation-privacy/` | derived source missing/cross-user/wrong-target/draft/empty 隔离与隐私红线 | automated | Ready |
| E2E.P0.073 | backend-practice D-5, D-14 | `p0-073-practice-debrief-mode-regression/` | goal=debrief × assisted/strict 可启动，legacy mode=debrief 负向 | automated | Ready |
| E2E.P0.074 | backend-resume C-6, C-14, C-15 | `p0-074-resume-confirm-master-and-version-reads/` | confirm structured master + get/list resume versions + unique index + pagination + cross-user 404 | automated | Ready |
| E2E.P0.075 | backend-resume C-14 | `p0-075-resume-update-version-merge-and-ik/` | update resume version partial merge + idempotency replay/mismatch + cross-user/deleted 404 | automated | Ready |
| E2E.P0.076 | backend-resume C-10 | `p0-076-resume-branch-version-sync-paths/` | branch resume version copy_master/blank sync paths + idempotency + cross-user target isolation | automated | Ready |
| E2E.P0.077 | backend-resume C-10, C-16 | `p0-077-resume-tailor-async-dispatch-and-ready/` | resume tailor ai_select dispatch + request/read + drainer ready path with suggestions/task-run/outbox | automated | Ready |
| E2E.P0.078 | backend-resume C-16 | `p0-078-resume-tailor-failure-and-retry/` | resume_tailor timeout retryable, output_invalid terminal, retry-to-ready, and ready-only outbox | automated | Ready |
| E2E.P0.079 | backend-resume C-16 | `p0-079-resume-suggestion-accept-reject-terminal/` | resume suggestion accept/reject terminal decision + idempotency + cross-user isolation + profile stability | automated | Ready |
