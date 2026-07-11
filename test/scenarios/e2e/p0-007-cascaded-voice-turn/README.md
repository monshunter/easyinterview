# E2E.P0.007 Cascaded phone turn

> **场景 ID**: E2E.P0.007
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom + Go httptest)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

用户已登录并进入同一 practice session，当前 session 有 active turn。用户通过 TopBar 唯一电话图标进入电话模式；fixture 提供独立 active STT / chat / TTS profiles，浏览器端复用一条 call-scoped microphone stream，并由 VAD 为每次发言切出小音频后提交底层 `createPracticeVoiceTurn`。

## 2 When

用户在电话模式中开始说话并停顿。VAD 自动结束本次 utterance 并提交 voice turn；前端发送 `POST /api/v1/practice/sessions/{sessionId}/voice-turns`，携带 `Idempotency-Key`、`clientVoiceTurnId`、`turnId` 和小型 base64 audio payload；后端使用 server-owned plan / turn status / session language / committed context，按 STT -> chat -> TTS 级联执行并返回 transcript、assistant draft、TTS chunk metadata 和 provider meta summary。

## 3 Then

- 页面展示最终 transcript、assistant 文本和 TTS 播放状态
- 顶部电话图标和中心挂断按钮共用同一退出路径；退出只回到同一 session 的文本模式，不存在重新开始或 call-ended 中间态
- `createPracticeVoiceTurn` 只在 request-level `Idempotency-Key` 下产生一次 side effect，replay 不重复写 session event
- 后端 voice turn event 持久化 provider-neutral metadata，不保存原始音频或 TTS bytes
- response 的 provider meta summary 同时包含 `practice.voice.stt.default`、`practice.followup.default`、`practice.voice.tts.default`
- 用户得到可继续下一题的 running session，当前 turn 进入 `follow_up_requested`

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```

## 5 关联需求

`practice-voice-mvp` C-1, C-2, C-3, C-4, C-5, D-1, D-2, D-3, D-4, D-6, D-8
