# E2E.P0.009 Phone provider failure fallback

> **场景 ID**: E2E.P0.009
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom + Go unit tests)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

用户在电话模式中提交底层 voice turn。STT、chat、TTS 分别通过独立 A3 profiles 解析，realtime voice profile 仍是 unsupported reserved capability。OpenAPI fixtures 提供 `stt-config-missing`、`chat-failed`、`chat-output-invalid`、`tts-failed` scenarios。

## 2 When

测试分别模拟 STT secret missing、chat fallback exhausted、chat 连续两次返回不符合 session language 的问题、TTS provider timeout、TTS disabled profile 和 realtime profile 被误用于 transcription。

## 3 Then

- STT 失败 fail-fast，不调用 chat/TTS
- Chat 失败不调用 TTS
- Chat 输出语言不符合 session language 时只 repair 一次；第二次仍不合法则返回顶层 `AI_OUTPUT_INVALID`，不调用 TTS、不持久化 voice result，前端留在同一 session 并显示本地化错误
- TTS 失败保留 transcript 和 assistant text，返回结构化 `ttsError`，前端仍可文本继续
- Realtime profile 在 STT path fail-closed，不调用 provider
- Voice-turn runtime 不静默走 stub 或 realtime profile，隐私 grep 不包含 raw audio、TTS bytes、provider secret 或 transcript 明文

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```

## 5 关联需求

`practice-voice-mvp` C-6, C-7, C-8, C-9, D-3, D-4, D-5, D-6
