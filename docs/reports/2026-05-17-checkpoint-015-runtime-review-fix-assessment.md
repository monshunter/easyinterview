# Checkpoint 015 Runtime Review Fix 交付复盘报告

> **日期**: 2026-05-17
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖 `main` 相较 `checkpoint/015` review 后确认的 runtime drift：Debrief `getJob` 真实 route 缺失、Practice voice TTS `audioRef` 不可播放、voice committed context 未进入下一轮 prompt。
- 已建档 [BUG-0070](../bugs/BUG-0070.md)，并将修复记录到 [2026-05-17 工作日志](../work-journal/2026-05-17.md)。
- 通过验证：
  - `go test ./internal/jobs ./internal/api/jobs ./internal/store/jobs ./internal/practice ./internal/store/practice ./cmd/api -count=1`
  - `pnpm --dir frontend typecheck`
  - `pnpm --dir frontend test src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx --run`
  - `git diff --check`

## 2 会话中的主要阻点/痛点

- Generated/client/fixture evidence 没有证明真实 `cmd/api` runtime route 已挂载。
  - **证据**：`getJob` contract 与 generated client 存在，但真实后端缺少 `GET /api/v1/jobs/{jobId}` mux registration。
  - **影响**：Debrief Step 1 真实 backend polling 会 404，mock / fixture 通过不能代表真实路径可用。
- Opaque media ref 穿过了 backend/frontend 边界。
  - **证据**：voice service 返回 `voice-turn://...`，frontend 直接交给 `new Audio()`。
  - **影响**：TTS chunk 元数据存在，但浏览器主路径不可播放。
- Committed context builder 与生产 replay 链路脱节。
  - **证据**：`BuildCommittedVoiceContext` 有局部测试，但 `CreatePracticeVoiceTurn` 没有从已存 events 加载 context。
  - **影响**：barge-in 后下一轮 AI payload 可能不知道用户已经听到哪些 assistant 内容。
- Barge-in event evidence 不足。
  - **证据**：frontend 原先只发送 `barge_in_detected`，没有 partial `tts_chunk_played` 的 `playedTextLength`。
  - **影响**：backend 即使回放事件，也无法精确构造已提交上下文。

## 3 根因归类

- `spec-plan`：相关 owner plan 的完成 gate 没有强制证明 frontend-consumed async operation 从 OpenAPI 到 `cmd/api` route 的端到端 runtime wiring。
- `spec-plan`：voice plan gate 没有把 response playback ref 的浏览器可消费性与持久化 privacy 边界分开验证。
- `spec-plan`：barge-in committed context gate 偏重 builder 行为，没有要求 frontend event evidence、store replay 和下一轮 chat payload 三段联测。
- `no repo change needed`：本次实现代码修复本身已补齐 runtime/test 证据；流程资产改进应作为后续 owner plan gate hardening 单独处理。

## 4 对流程资产的改进建议

- 在 `frontend-debrief/001-debrief-screen-and-handoff` 或共用 L2 review gate 中加入 async polling operation matrix：OpenAPI、fixture、generated client、backend handler/store、`cmd/api` route、focused route test 必须逐项有证据。
  - **落点**：spec-plan
  - **优先级**：high
- 在 `practice-voice-mvp/001-cascaded-stt-llm-tts` 的 voice playback gate 中加入 media-ref rule：response ref 必须浏览器可播放，或 resolver route 必须同 plan 落地；store event 必须验证不持久化 audio bytes。
  - **落点**：spec-plan
  - **优先级**：high
- 在 voice barge-in gate 中加入 committed-context tripwire：barge-in 前必须上报 partial played evidence，store replay 必须生成 context，下一轮 AI payload 必须包含 interruption / committed assistant context。
  - **落点**：spec-plan
  - **优先级**：high

## 5 建议优先级与后续动作

- 下一轮最值得做的是一个小型 docs hardening pass：原地修订 `frontend-debrief/001-debrief-screen-and-handoff` 与 `practice-voice-mvp/001-cascaded-stt-llm-tts` 的 review / BDD gates，把本次 BUG-0070 的三类 runtime tripwire 写回 owner plan。
- 可以延后处理的是通用 skill 层规则抽象；先让两个直接 owner plan 具备可执行 gate，再评估是否沉淀到 `/plan-code-review` 共享规则。
