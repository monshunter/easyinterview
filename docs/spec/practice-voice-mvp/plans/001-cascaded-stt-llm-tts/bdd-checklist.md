# Cascaded STT LLM TTS Voice MVP BDD Checklist

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-07-09

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.007 完整级联语音面试 turn

- [x] 创建场景目录 `test/scenarios/e2e/p0-007-cascaded-voice-turn/`
- [x] 准备测试数据：登录用户、practice session、active fixture STT/chat/TTS profiles、当前题目、可播放 TTS chunk metadata
- [x] 验证 BUG-0070 playback ref 边界：response `ttsChunks[].audioRef` 可被浏览器播放；persisted voice turn summary 使用 opaque `voice-turn://...` 且不包含 audio bytes
- [x] 实现 setup / trigger / verify / cleanup；trigger 写入真实 runner 日志，verify 断言 runner marker、目标测试路径、pass marker、transcript、assistant text、TTS playback state、voice turn event、profile meta 摘要和下一题入口
- [x] 更新 `test/scenarios/e2e/INDEX.md` 并执行场景，记录验证证据；证据: `test/scenarios/e2e/p0-007-cascaded-voice-turn/scripts/setup.sh && test/scenarios/e2e/p0-007-cascaded-voice-turn/scripts/trigger.sh && test/scenarios/e2e/p0-007-cascaded-voice-turn/scripts/verify.sh && test/scenarios/e2e/p0-007-cascaded-voice-turn/scripts/cleanup.sh`
- [ ] Phase 6 revision: 场景 README / expected outcome 改为用户可见电话模式，不再要求语音分析或手动转写 fallback UI。

## E2E.P0.008 插话只提交已播放上下文

- [x] 创建场景目录 `test/scenarios/e2e/p0-008-voice-barge-in-committed-context/`
- [x] 准备测试数据：multi-chunk TTS response、complete/partial `tts_chunk_played` event、barge-in event、下一轮用户输入
- [x] 实现 setup / trigger / verify / cleanup；trigger 写入真实 runner 日志，verify 断言 runner marker、目标测试路径、pass marker、未播放 draft 不在 committed context / prompt 中，interruption note 存在
- [x] 验证 BUG-0070 committed-context replay 边界：barge-in 前 partial `tts_chunk_played` 含 `playedTextLength`，store replay 生成 committed context 并进入下一轮 prompt
- [x] 更新 `test/scenarios/e2e/INDEX.md` 并执行场景，记录验证证据；证据: `test/scenarios/e2e/p0-008-voice-barge-in-committed-context/scripts/setup.sh && test/scenarios/e2e/p0-008-voice-barge-in-committed-context/scripts/trigger.sh && test/scenarios/e2e/p0-008-voice-barge-in-committed-context/scripts/verify.sh && test/scenarios/e2e/p0-008-voice-barge-in-committed-context/scripts/cleanup.sh`

## E2E.P0.009 Provider failure fallback

- [x] 创建场景目录 `test/scenarios/e2e/p0-009-voice-provider-failure-fallback/`
- [x] 准备测试数据：STT secret missing、TTS provider error、unsupported realtime profile 三类 fixture
- [x] 实现 setup / trigger / verify / cleanup；trigger 写入真实 runner 日志，verify 断言 runner marker、目标测试路径、pass marker、fail-fast / 文本 fallback / 不走 stub 或 realtime / privacy grep 无明文
- [x] 更新 `test/scenarios/e2e/INDEX.md` 并执行场景，记录验证证据；证据: `test/scenarios/e2e/p0-009-voice-provider-failure-fallback/scripts/setup.sh && test/scenarios/e2e/p0-009-voice-provider-failure-fallback/scripts/trigger.sh && test/scenarios/e2e/p0-009-voice-provider-failure-fallback/scripts/verify.sh && test/scenarios/e2e/p0-009-voice-provider-failure-fallback/scripts/cleanup.sh`

## REAL.ENV.PHONE.SCREENSHOT

- [ ] Verify local dev dependencies/backend/frontend are running from the current branch.
- [ ] Open real phone practice flow in browser and capture screenshot evidence under `.test-output/`.
- [ ] Verify screenshot shows phone-mode UI with captions / hang-up / restart and no right panel, voice analysis, manual transcript fallback, start-recording main button or submit-turn main button.
