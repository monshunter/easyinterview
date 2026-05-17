# Practice Voice Fixture Audio Ref 交付复盘报告

> **日期**: 2026-05-17
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `practice-voice-mvp/001-cascaded-stt-llm-tts` 中 `createPracticeVoiceTurn` fixture 的 `audioRef` 语义漂移，确保 fixture-backed frontend mock 与真实 backend response 一样返回浏览器可播放 media ref。
- 成功证据：
  - `pnpm --dir frontend test src/api/devMockClient.test.ts --run`
  - `make validate-fixtures`
  - `make lint-openapi`
  - `go test ./internal/practice ./internal/api/practice ./internal/store/practice ./cmd/api -count=1`
  - `make docs-check`
  - `make codegen-check`
  - `git diff --check`

## 2 会话中的主要阻点/痛点

- BUG-0070 后真实 backend 已修复 `data:audio` response，但 fixture 仍保留 `fixture-audio://...`。
  - **证据**：新增 Red test 后，`frontend/src/api/devMockClient.test.ts` 收到 `fixture-audio://practice-voice/default/chunk-001` 并失败。
  - **影响**：本地 fixture/mock 路径看起来完成了 voice turn，但浏览器播放器不能消费该 ref，容易再次误判为“接口有返回但语音无反应”。
- 既有 `validate_fixtures.py` 只做 schema/coverage/privacy gate，没有 media semantic gate。
  - **证据**：旧 fixture schema-valid，直到新增 `data:audio` 断言才暴露漂移。
  - **影响**：OpenAPI fixture 可以通过全局校验却违反 feature-specific playback contract。

## 3 根因归类

- `spec-plan`：`practice-voice-mvp` 已写明 response/persistence 分离，但 fixture-level mock 语义没有作为独立 RF gate 固化。
- `README` / tooling：fixture validator 的通用规则无法覆盖 voice/media 特有语义，必须在 lint 中补 feature-specific semantic check。

## 4 对流程资产的改进建议

- 已落地：`scripts/lint/validate_fixtures.py` 增加 `createPracticeVoiceTurn` audioRef gate，拒绝 mock-only scheme。
  - **落点**：tooling
  - **优先级**：high
- 已落地：`practice-voice-mvp` spec/plan/checklist 升到 v1.3，并记录 RF.4 fixture playback ref gate。
  - **落点**：spec-plan
  - **优先级**：high
- 建议后续：如果落地 resolver URL，不要只改 schema；同时新增 resolver route test、fixture validator allowlist 和 frontend playback smoke。
  - **落点**：spec-plan / tooling
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级已完成：fixture 与真实 backend response 的 `audioRef` 语义重新对齐，并加入可执行 gate。
- 后续若继续推进真实语音能力，应优先补 real-provider smoke：验证 dev/Kind 中真实 STT/TTS secret 缺失 fail-fast 与成功配置下的端到端音频播放，而不是继续扩展 fixture-only 断言。
