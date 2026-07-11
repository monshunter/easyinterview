# Cascaded STT LLM TTS Voice MVP BDD Plan

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-07-11

## Scenario Matrix

| 场景 ID | 类型 | 阶段 | 场景 | 验证入口 |
|---------|------|------|------|----------|
| E2E.P0.007 | primary | Phase 7 | 完整电话 turn：VAD 静音自动提交，用户看到 transcript、听到同语言 AI TTS，TTS 结束自动重新监听 | `test/scenarios/e2e/p0-007-cascaded-voice-turn/` |
| E2E.P0.008 | alternate/recovery | Phase 7 | 挂断与真实插话分离：挂断回文本且不发 barge-in；speech-start 才提交 partial playback + barge-in | `test/scenarios/e2e/p0-008-voice-barge-in-committed-context/` |
| E2E.P0.009 | failure/recovery | Phase 7 | provider failure 不做 business repair；连续 invalid follow-up 返回既有顶层 `AI_OUTPUT_INVALID`，session 行不变且无 result/TTS/canned question | `test/scenarios/e2e/p0-009-voice-provider-failure-fallback/` |
| REAL.ENV.PHONE.SCREENSHOT | final acceptance | Phase 7 | 真实本地环境证明单一电话图标与圆形挂断存在，分段/live/切断文字/重开/callEnded 不存在，挂断回到同 session 文本模式 | `.test-output/` browser screenshot artifacts |

## Phase 7: Phone lifecycle and conversational-integrity scenarios

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.007 | 完整级联电话模式 turn | 用户已登录并进入电话模式；profiles active；session language 与当前问题已确定 | 用户说出非空回答并静音，随后 AI TTS 播放结束 | VAD 自动提交；chat 使用 canonical context + session language；response audioRef 可播放；persisted summary 无 audio bytes；TTS-ended 自动重新监听；用户无需点击录音/提交 | `test/scenarios/e2e/p0-007-cascaded-voice-turn/` |
| E2E.P0.008 | 挂断与真实插话分离 | AI 正播放第二个 TTS chunk，第一 chunk 已完整上报 | A 用户点击挂断；B 用户真实开始说话 | A 立即停麦克风/TTS，按需提交 heard prefix 但不发 barge-in，并回同 session 文本模式；B 先上报 partial `tts_chunk_played` 再上报 `barge_in_detected`；两者都不把未播放 draft 注入下一轮 prompt | `test/scenarios/e2e/p0-008-voice-barge-in-committed-context/` |
| E2E.P0.009 | Provider / follow-up repair failure | 用户在电话模式；fixtures 覆盖 provider failure、结构合法但语言错误，以及连续 malformed structured result | 用户提交 voice turn | STT/config/provider/timeout failure 不做 business repair；TTS 失败保留 assistant text；chat 首次 parser/language invalid 只 repair 一次，第二次返回顶层 `AI_OUTPUT_INVALID`、session 行不变且无 `PracticeVoiceTurnResult`/canned question/TTS；前端可回同一 session 文本模式；系统不调用 stub/realtime | `test/scenarios/e2e/p0-009-voice-provider-failure-fallback/` |

每个场景目录必须遵守 `test/scenarios/README.md` 与 `test/scenarios/e2e/README.md`：`trigger.sh` 写入真实 runner 日志，`verify.sh` 检查 runner marker、目标测试路径和 pass marker；场景创建时同步 `test/scenarios/e2e/INDEX.md`，不得用文件存在检查代替执行证据。
