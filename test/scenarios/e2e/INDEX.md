# E2E 场景索引

> 场景按阶段分组，记录编号、关联需求、目录路径和状态。

---

## P0 核心闭环

| 场景 ID | 关联需求 | 目录 | 描述 | 执行方式 | 状态 |
|---------|----------|------|------|----------|------|
| E2E.P0.001 | frontend-shell C-1 | `p0-001-default-home-shell/` | 默认进入首页并呈现三入口 TopBar、单登录入口，复盘 / 用户画像不可达 | automated | Ready |
| E2E.P0.002 | frontend-shell C-2 | `p0-002-auth-pending-action-resume/` | 登录打断后接续原业务动作与上下文 | automated | Ready |
| E2E.P0.003 | backend-auth C-1 | `p0-003-email-code-session-cookie/` | 邮箱挑战验证后签发 first-party session 并支持 /me 与 logout | automated | Ready |
| E2E.P0.004 | frontend-shell C-7 | `p0-004-app-shell-language-switch/` | App Shell 中英语言切换并携带 Accept-Language display hint | automated | Ready |
| E2E.P0.005 | frontend-shell C-8 | `p0-005-app-shell-visual-system-smoke/` | D2 视觉系统 smoke：DOM/className/CSS-variable/customAccent overlay/non-current 负向 + ui-design 源追溯 | automated | Ready |
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
| E2E.P0.016 | frontend-home-job-picks-and-parse C-5, C-7, C-17 | `p0-016-parse-confirm-to-workspace/` | Parse 编辑 + 绑定简历 + Save/Start handoff | automated | Ready |
| E2E.P0.018 | frontend-workspace-and-practice C-2, C-7, C-8, C-9 | `p0-018-workspace-default-render/` | Workspace 默认渲染：plan eyebrow + header + Interview Launcher + Main Left/Right + Modals | automated | Ready |
| E2E.P0.019 | frontend-workspace-and-practice C-2, C-3, C-8, C-9 | `p0-019-workspace-context-loading/` | Workspace context loading：empty/missing-resume 空态 + getPracticePlan refresh | automated | Ready |
| E2E.P0.020 | frontend-workspace-and-practice C-1, C-3, C-12 | `p0-020-workspace-start-practice/` | 立即面试 双步契约 + Idempotency-Key + pendingAction 未登录接续 | automated | Ready |
| E2E.P0.021 | frontend-workspace-and-practice C-7, C-9, C-10, C-12 | `p0-021-workspace-handoff/` | Workspace handoff + 隐私红线 + non-current negative grep | automated | Ready |
| E2E.P0.022 | backend-practice C-1, C-13 | `p0-022-practice-plan-baseline-create-and-read/` | createPracticePlan baseline、idempotency replay、getPracticePlan 与 cross-user 404 隔离 | automated | Ready |
| E2E.P0.023 | backend-practice C-4 | `p0-023-practice-session-start-and-first-question/` | startPracticeSession 同步首题、getPracticeSession、session_started event 与 practice.session.started outbox | automated | Ready |
| E2E.P0.024 | backend-practice C-5, C-21, C-23 | `p0-024-practice-session-ai-failure-retry/` | AI timeout 后 failed_retryable reservation，同 key 重试成功且 outbox 仅一次 | automated | Ready |
| E2E.P0.025 | backend-practice C-10, C-13, C-22, C-23, C-24, C-25 | `p0-025-practice-idempotency-and-isolation-matrix/` | startPracticeSession replay / mismatch / 跨用户隔离 / 同 plan 多 key conflict / cross-user 404 矩阵 | automated | Ready |
| E2E.P0.026 | backend-practice C-16, D-11 | `p0-026-practice-observability-and-privacy-redlines/` | observed AI typed columns、metric allowlist、隐私红线与 backend-practice non-current gate | automated | Ready |
| E2E.P0.032 | frontend-shell C-10 | `p0-032-dev-mock-auth-state-and-user-menu/` | Dev mock 默认非登录、登录后头像 dropdown、settings 分流与 logout 后非登录闭环 | automated | Ready |
| E2E.P0.033 | backend-upload C-1, C-2, C-3, C-4, C-6, C-7, C-8 | `p0-033-file-presign-register-roundtrip/` | file presign、IK replay、register 校验、cross-user 隔离与 privacy delete tombstone | automated | Ready |
| E2E.P0.034 | backend-resume C-1, C-2, C-5, C-6, C-7, C-8 | `p0-034-resume-register-and-list/` | Resume register/get/list：三 sourceType、IK replay、upload handoff、cross-user 404、cursor pagination 与 fixture parity | automated | Ready |
| E2E.P0.035 | backend-resume C-3, C-4, C-13 | `p0-035-resume-parse-async-job-lifecycle/` | resume.parse in-process drainer：AI parse、ready/failed 状态、typed task run、ready-only outbox 与 privacy redlines | automated | Ready |
| E2E.P0.036 | frontend-resume-workshop C-1, C-2, C-10, C-11 | `p0-036-resume-flat-list-auth-boundary/` | Resume Workshop flat list：route shell、auth gate、fixture-derived flat rows、open-to-detail navigation 与 non-current-route negative grep | automated | Ready |
| E2E.P0.037 | frontend-resume-workshop C-3, C-10, C-11 | `p0-037-resume-detail-preview-readonly/` | Resume Workshop detail：Preview Tab 投影、原件 modal a11y、Export PDF Idempotency-Key + P0 501 toast、404 fallback 不回显 fixture error.code | automated | Ready |
| E2E.P0.044 | frontend-workspace-and-practice C-4, C-8, C-9 | `p0-044-practice-text-loop-assisted-happy-path/` | Practice text loop assisted happy path：PracticeScreen + appendSessionEvent default + ask_follow_up renderer + IK 双轨边界 | automated | Ready |
| E2E.P0.045 | frontend-workspace-and-practice C-4, C-10, C-12 | `p0-045-practice-text-loop-mode-policy-display/` | Practice text loop 显隐：strict / assisted × baseline / retry_current_round / next_round + hint / skip / pause-resume + strict toggle 锁定 toast + 非当前口径负向 | automated | Ready |
| E2E.P0.046 | frontend-workspace-and-practice C-4, C-12 | `p0-046-practice-text-loop-failure-and-recovery/` | Practice text loop 失败处理：AI 502 + session 404 + 409 mismatch + 409 strict-hint + retry 复用 IK | automated | Ready |
| E2E.P0.047 | frontend-workspace-and-practice C-4, C-6, C-12 | `p0-047-practice-text-loop-complete-and-generating-handoff/` | Practice text loop 完成 handoff：completePracticeSession 202 + IK + handoff generating + 隐私红线 | automated | Ready |
| E2E.P0.048 | backend-practice C-7, C-8b, C-12 | `p0-048-practice-hint-assisted-across-goals/` | assisted hint 主路径：3 个 goal 下返回 show_hint、写 hint_generate task run 且不推进 turn lifecycle | automated | Ready |
| E2E.P0.049 | backend-practice C-8, C-8b, D-38 | `p0-049-practice-hint-strict-refusal/` | strict hint 拒绝：3 个 goal 下返回 409 hint_disabled_in_mode，same clientEventId replay 不留下 pending reservation | automated | Ready |
| E2E.P0.050 | backend-practice C-12, D-37 | `p0-050-practice-hint-provenance-task-runs/` | AssistantAction wire provenance 仅 6 字段，runtime feature_key 只进入 ai_task_runs typed columns | automated | Ready |
| E2E.P0.051 | backend-practice C-16, C-17, D-36 | `p0-051-practice-hint-degrade-privacy/` | hint AI graceful degrade：200 session_wait、session running、failed task run 与隐私红线/non-current negative | automated | Ready |
| E2E.P0.056 | frontend-report-dashboard C-1, C-2, C-5, C-8, C-11 | `p0-056-generating-to-report-happy-path/` | Generating → Report happy path：5-phase poll → ReportDashboard mount + 5 detail tabs + ContextStrip + testid coverage + read-only contract | automated | Ready |
| E2E.P0.057 | frontend-report-dashboard C-3, C-6 | `p0-057-replay-cta-paths-a-and-b/` | Replay CTA paths A/B：retry_current_round + next_round payload, replay_practice pendingAction round-trip, no raw text | automated | Ready |
| E2E.P0.058 | frontend-report-dashboard C-4, C-12 | `p0-058-report-failure-and-missing-session/` | ReportFailureState AI_* enum + ReportMissingSessionState + cross-user 404 not-found copy + Generating timeout retry | automated | Ready |
| E2E.P0.059 | frontend-report-dashboard C-13, C-14, D-12 | `p0-059-report-pixel-parity-i18n-and-non-current-negative/` | i18n namespace sync + AI_* enum coverage + non-current vocab negative grep + Playwright pixel-parity specs staged | automated | Ready |
| E2E.P0.070 | backend-practice C-2, C-3 | `p0-070-practice-derived-plan-create-read-replay/` | report-derived createPracticePlan、getPracticePlan 与 idempotency replay source 字段 | automated | Ready |
| E2E.P0.072 | backend-practice C-2, C-3, D-11 | `p0-072-practice-derived-source-isolation-privacy/` | report source missing/cross-user/wrong-target 隔离与隐私红线 | automated | Ready |
| E2E.P0.074 | backend-resume C-6, C-14, C-15 | `p0-074-resume-flat-read-api/` | get/list Resume flat reads：fixture parity、pagination、cross-user 404 与非当前 route/catalog gate | automated | Ready |
| E2E.P0.075 | backend-resume C-14 | `p0-075-resume-update-flat-fields-and-ik/` | updateResume flat fields：idempotency replay/mismatch、validation、cross-user/deleted 404 | automated | Ready |
| E2E.P0.076 | backend-resume C-10 | `p0-076-resume-duplicate-save-as-new/` | duplicateResume save-as-new：source snapshot copy、editable overlay、idempotency、cross-user source isolation | automated | Ready |
| E2E.P0.077 | backend-resume C-10, C-16 | `p0-077-resume-tailor-async-dispatch-and-ready/` | resume tailor ai_select dispatch + request/read + drainer ready path with suggestions/task-run/outbox | automated | Ready |
| E2E.P0.078 | backend-resume C-16 | `p0-078-resume-tailor-failure-and-retry/` | resume_tailor timeout retryable, output_invalid terminal, retry-to-ready, and ready-only outbox | automated | Ready |
| E2E.P0.079 | backend-resume C-16 | `p0-079-resume-rewrites-accept-only-save/` | Rewrites accept-only save：flat save fixture parity、frontend overwrite/save-as-new 与非当前 route/catalog gate | automated | Ready |
| E2E.P0.080 | backend-resume C-13 | `p0-080-resume-tailor-privacy-negative/` | resume tailor privacy payload：outbox / ai_task_runs / audit redaction 与非当前 vocabulary negative gate | automated | Ready |
| E2E.P0.081 | frontend-resume-workshop C-10, C-7, C-8, C-9 | `p0-081-resume-create-flow-upload-paste-guided-happy/` | ResumeCreateFlow upload/paste happy + guided-negative + presign + register + parse polling + IK + privacy + non-current negative | automated | Ready |
| E2E.P0.082 | frontend-resume-workshop C-10, C-8 | `p0-082-resume-create-flow-parsing-failure-and-retry/` | Agent Parsing failure / PARSE_TIMEOUT / cancel-and-return + retry-from-input | automated | Ready |
| E2E.P0.083 | frontend-resume-workshop C-10, C-8, C-9 | `p0-083-resume-create-flow-preview-confirm-and-cta-handoff/` | PreviewConfirm save v1 + 409 fallback + 422 inline + Home/Workspace CTA pendingAction handoff | automated | Ready |
| E2E.P0.084 | frontend-resume-workshop C-11, C-8, C-9 | `p0-084-resume-flat-ui-regression/` | Flat Resume UI regression：route/auth/create/detail/Rewrites smoke + non-current operation / prototype import negative | automated | Ready |
| E2E.P0.085 | frontend-resume-workshop C-11, C-8 | `p0-085-resume-rewrites-tab-tailor-run-polling/` | Flat Rewrites Tab tailor polling (queued/generating/ready/failed/timeout) + rerun IK/no-IK read path + unmount cleanup | automated | Ready |
| E2E.P0.086 | frontend-resume-workshop C-11, C-8 | `p0-086-resume-rewrites-edit-save/` | Rewrites accept-only save + flat profile merge + Edit Tab updateResume + targetJobId rerun context + non-current operation gate | automated | Ready |
| E2E.P0.087 | frontend-resume-workshop C-11, C-9, C-8 | `p0-087-resume-detail-export-copy-consistency-and-parity/` | Export PDF 501 stub + copyText fallback + flat Detail/Rewrites/Edit parity + non-current tailor mode / non-current-entry / prototype import negative | automated | Ready |
| E2E.P0.088 | frontend-shell C-11 | `p0-088-url-addressable-routing-canonical/` | Canonical path deep-link / reload / back-forward 在 workspace / practice / generating / report / resume-versions 保留 safe params；非当前 debrief/profile path 折回首页 | automated | Ready |
| E2E.P0.089 | frontend-shell C-12 | `p0-089-url-routing-auth-privacy/` | Auth pendingAction round-trip + URL / history / storage / console 19 类 raw marker 零命中 + safe handoff keys 保留 | automated | Ready |
| E2E.P0.090 | frontend-shell C-13 | `p0-090-url-routing-hash-non-current-negative/` | `#route=...` 兼容 + non-current aliases（含 debrief/debrief_full/profile）不 materialize + SPA fallback 仅服务 canonical path 且不吞后端 / 静态资源 | automated | Ready |
| E2E.P0.098 | e2e-scenarios-p0 C-1, C-2, C-3, C-4, C-5, C-6, C-7 | `p0-098-full-funnel-import-to-next-round-journey/` | API-level full funnel：resume seed + JD import + practice + report + next_round plan + idempotency/privacy/non-current gates | automated | Ready |
| E2E.P0.099 | e2e-scenarios-p0 C-1, C-2, C-4, C-5, C-6 | `p0-099-full-funnel-fullstack-ui-journey/` | Playwright full-stack UI：home → parse → workspace → practice → generating → report → next_round practice with real backend | automated | Ready |
| E2E.P0.100 | e2e-scenarios-p0 C-9, C-10, C-11, C-12, C-13, C-14 | `p0-100-real-provider-full-funnel-hybrid/` | AI Agent first-run preflight + real provider browser UAT handoff：shared env、real backend/frontend、Mailpit login、redacted provider evidence | hybrid | Ready |
| E2E.P0.101 | backend-auth C-9; frontend-shell C-14; local-dev-stack C-10/C-15 | `p0-101-auth-email-code-profile-setup/` | Playwright host-run real-mode auth：单一邮箱验证码入口；首次登录强制资料补全；同一邮箱后续登录不重复补全 | automated | Ready |
| E2E.P0.102 | frontend-shell C-15; backend-auth C-5 | `p0-102-auth-gated-interview-routes/` | 未登录 Home 隐藏 Recent mock interviews 和复盘 CTA；保留业务入口和保护路由先跳登录；非当前 debrief/profile 不作为保护业务目标 | automated | Ready |
