# Cascaded STT LLM TTS Voice MVP BDD Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-08

## Scenario Matrix

| 场景 ID | 类型 | 阶段 | 场景 | 验证入口 |
|---------|------|------|------|----------|
| E2E.P0.007 | primary | Phase 5 | 完整语音 turn：用户说话、看到 transcript、听到 AI TTS、继续下一题 | `test/scenarios/e2e/p0-007-cascaded-voice-turn/` |
| E2E.P0.008 | failure/recovery | Phase 5 | 用户在 AI 播放中插话，系统只提交已播放 chunk，未播放 draft 不进上下文 | `test/scenarios/e2e/p0-008-voice-barge-in-committed-context/` |
| E2E.P0.009 | alternate/failure | Phase 5 | STT/TTS provider 缺配置或失败时 fail-fast / 文本 fallback，不静默走 stub 或 realtime | `test/scenarios/e2e/p0-009-voice-provider-failure-fallback/` |

## Phase 5: Voice MVP scenarios

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.007 | 完整级联语音面试 turn | 用户已登录并进入 `practice?mode=voice&modality=voice`；fixture 提供 active STT/chat/TTS profiles；session 有当前题目 | 用户录入一段回答并提交 voice turn | 页面展示用户 transcript、assistant 文本与 TTS 播放状态；后端记录 voice turn event；AI meta 摘要含 stt/chat/tts profile；用户可继续下一题 | `test/scenarios/e2e/p0-007-cascaded-voice-turn/` |
| E2E.P0.008 | 插话只提交已播放上下文 | AI 回复被拆成多个 TTS chunks；前端已上报第一个 chunk 完整播放，第二个 chunk 正在播放 | 用户开始说话触发 barge-in | 播放立即停止；后端只提交第一个 chunk 的 assistant 文本；未播放 draft 不进入下一轮 prompt；下一轮 prompt 带 interruption note | `test/scenarios/e2e/p0-008-voice-barge-in-committed-context/` |
| E2E.P0.009 | Provider failure fallback | 用户在语音面试中；fixture 分别模拟 STT secret missing、TTS provider error、unsupported realtime profile | 用户提交 voice turn 或请求播放 | STT 缺配置返回可理解错误且不调用 chat/TTS；TTS 失败保留 assistant 文本并允许继续文本面试；系统不调用 stub 或 realtime；隐私 grep 无明文泄漏 | `test/scenarios/e2e/p0-009-voice-provider-failure-fallback/` |
