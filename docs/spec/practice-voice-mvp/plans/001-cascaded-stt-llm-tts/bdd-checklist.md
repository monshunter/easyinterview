# Cascaded STT LLM TTS Voice MVP BDD Checklist

> **版本**: 1.7
> **状态**: active
> **更新日期**: 2026-07-11

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.007 完整级联电话模式 voice turn

- [x] 创建场景目录 `test/scenarios/e2e/p0-007-cascaded-voice-turn/`
- [x] 准备测试数据：登录用户、practice session、active fixture STT/chat/TTS profiles、当前题目、可播放 TTS chunk metadata
- [x] 验证 BUG-0070 playback ref 边界：response `ttsChunks[].audioRef` 可被浏览器播放；persisted voice turn summary 使用 opaque `voice-turn://...` 且不包含 audio bytes
- [x] 实现 setup / trigger / verify / cleanup；trigger 写入真实 runner 日志，verify 断言 runner marker、目标测试路径、pass marker、transcript、assistant text、TTS playback state、voice turn event、profile meta 摘要和下一题入口
- [x] 更新 `test/scenarios/e2e/INDEX.md` 并执行场景，记录验证证据；证据: `test/scenarios/e2e/p0-007-cascaded-voice-turn/scripts/setup.sh && test/scenarios/e2e/p0-007-cascaded-voice-turn/scripts/trigger.sh && test/scenarios/e2e/p0-007-cascaded-voice-turn/scripts/verify.sh && test/scenarios/e2e/p0-007-cascaded-voice-turn/scripts/cleanup.sh`
- [x] Phase 6 revision: 场景 README / expected outcome 改为用户可见电话模式，期望只覆盖当前电话 surface 控件。
  - 2026-07-10 evidence: P0.007-P0.009 scenario README/seed/expected-outcome wording now uses user-visible phone mode while preserving backend `voice turn` engineering names; P0.007-P0.009 setup/trigger/verify/cleanup all passed after the wording and wrapper checks were updated.
- [x] Phase 7 revision: VAD silence auto-submits non-empty capture, chat uses session language/canonical context, and TTS-ended re-arms listening without manual record/submit controls.
  <!-- verified: 2026-07-11 evidence="P0.007 PASS; VAD kernel/monitor, call-scoped recorder, controller and phone integration tests prove silence submit and TTS-ended re-arm, while backend voice tests prove canonical context and persisted session language." -->

## E2E.P0.008 插话只提交已播放上下文

- [x] 创建场景目录 `test/scenarios/e2e/p0-008-voice-barge-in-committed-context/`
- [x] 准备测试数据：multi-chunk TTS response、complete/partial `tts_chunk_played` event、barge-in event、下一轮用户输入
- [x] 实现 setup / trigger / verify / cleanup；trigger 写入真实 runner 日志，verify 断言 runner marker、目标测试路径、pass marker、未播放 draft 不在 committed context / prompt 中，interruption note 存在
- [x] 验证 BUG-0070 committed-context replay 边界：barge-in 前 partial `tts_chunk_played` 含 `playedTextLength`，store replay 生成 committed context 并进入下一轮 prompt
- [x] 更新 `test/scenarios/e2e/INDEX.md` 并执行场景，记录验证证据；证据: `test/scenarios/e2e/p0-008-voice-barge-in-committed-context/scripts/setup.sh && test/scenarios/e2e/p0-008-voice-barge-in-committed-context/scripts/trigger.sh && test/scenarios/e2e/p0-008-voice-barge-in-committed-context/scripts/verify.sh && test/scenarios/e2e/p0-008-voice-barge-in-committed-context/scripts/cleanup.sh`
- [x] Phase 7 variant A: hang-up stops microphone/TTS, may commit heard prefix, emits no barge-in, suppresses later phone TTS and returns to text for the same session.
- [x] Phase 7 variant B: real speech-start during TTS reports partial playback before `barge_in_detected` and keeps unplayed draft out of committed context.
  <!-- verified: 2026-07-11 evidence="P0.008 PASS; playback lifecycle and phone integration tests distinguish hang-up from real speech-start, settle only heard text, suppress late completion and keep unplayed draft out of committed context." -->

## E2E.P0.009 Provider failure fallback

- [x] 创建场景目录 `test/scenarios/e2e/p0-009-voice-provider-failure-fallback/`
- [x] 准备测试数据：STT secret missing、TTS provider error、unsupported realtime profile 三类 fixture
- [x] 实现 setup / trigger / verify / cleanup；trigger 写入真实 runner 日志，verify 断言 runner marker、目标测试路径、pass marker、fail-fast / 文本 fallback / 不走 stub 或 realtime / privacy grep 无明文
- [x] 更新 `test/scenarios/e2e/INDEX.md` 并执行场景，记录验证证据；证据: `test/scenarios/e2e/p0-009-voice-provider-failure-fallback/scripts/setup.sh && test/scenarios/e2e/p0-009-voice-provider-failure-fallback/scripts/trigger.sh && test/scenarios/e2e/p0-009-voice-provider-failure-fallback/scripts/verify.sh && test/scenarios/e2e/p0-009-voice-provider-failure-fallback/scripts/cleanup.sh`
- [x] Phase 7 revision: provider/config/timeout failures do not trigger business repair; malformed or wrong-language follow-up gets exactly one repair; a second invalid result returns the top-level `AI_OUTPUT_INVALID` envelope, leaves the session row unchanged, produces no result/canned question/TTS and permits same-session text-mode exit.
  <!-- verified: 2026-07-11 evidence="P0.009 PASS; focused voice tests prove provider failures receive no business repair, invalid structured/language output receives exactly one repair, and double invalid returns AI_OUTPUT_INVALID before result/TTS/session persistence." -->

## REAL.ENV.PHONE.SCREENSHOT

- [x] Verify local dev dependencies/backend/frontend are running from the current branch.
  - 2026-07-10 evidence: `test/scenarios/env-redeploy.sh all` rebuilt/restarted host-run backend/frontend from the current branch and `test/scenarios/env-verify.sh` reported Postgres, Redis, MinIO, Mailpit, frontend, and backend endpoints healthy.
- [x] Open real phone practice flow in browser and capture screenshot evidence under `.test-output/`.
  - 2026-07-10 evidence: synthetic local account and minimal real practice plan/session/turn rows were created, then the canonical real-mode URL `/practice?...mode=phone&modality=phone` was opened against `http://127.0.0.1:5173` + `http://127.0.0.1:8080/api/v1`. Screenshot evidence: `.test-output/practice-voice-mvp/phone-surface-acceptance/phone-surface-real-env-playwright.png`.
- [x] Historical Phase 6 screenshot captured the then-current captions / hang-up / restart surface; it is superseded and is not current acceptance evidence.
  - 2026-07-10 evidence: DOM evidence `.test-output/practice-voice-mvp/phone-surface-acceptance/phone-surface-real-env-playwright-dom.json` confirms required phone controls are present/in viewport and out-of-scope voice coming-soon, record/submit, manual transcription, voice analysis, start-recording, and submit-turn markers are absent.
- [x] Refresh real-mode evidence for the Phase 7 UI: single Top Bar handset + center red circular hang-up are present; segmented controls, live chip, visible cut-off copy, restart and `callEnded` are absent; hang-up returns to text for the same session.
  <!-- verified: 2026-07-11 evidence="Real API browser screenshots at .test-output/practice-phone-session-flow-0711/practice-text-real.png, practice-phone-real.png and practice-after-hangup-real.png are 1440x900. Before/after URLs retain sessionId=d1202287-c62f-4f85-af7f-059ba848ad1c; DB remains running with turn 1 asked." -->
