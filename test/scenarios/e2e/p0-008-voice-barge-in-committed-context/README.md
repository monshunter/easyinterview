# E2E.P0.008 Voice barge-in committed context

> **场景 ID**: E2E.P0.008
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom + Go unit tests)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

语音面试中 AI 已返回 TTS chunk metadata，前端正在播放 assistant response。后端已能接收 `tts_chunk_started`、`tts_chunk_played`、`barge_in_detected` 与 `assistant_context_committed` 事件，并可从 playback event stream 构造 committed assistant context。

## 2 When

用户在 AI 播放期间开始下一轮说话。前端停止当前 TTS playback，先上报 `barge_in_detected`，再进入新一轮录音；后端只把已播放或已明确 committed 的 assistant text 片段注入下一轮 prompt。

## 3 Then

- 前端停止 active TTS，并用 `appendSessionEvent` 上报 `barge_in_detected`
- voice playback events 使用 body-level `clientEventId` replay，禁止 `Idempotency-Key`
- complete chunk 可提交为 committed assistant context
- partial barge-in 只提交已播放长度，并生成 interruption note
- no playback 不提交未播放 draft；未播放 assistant draft 不进入下一轮 prompt

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```

## 5 关联需求

`practice-voice-mvp` C-3, C-5, D-8, D-9, D-10, D-11
