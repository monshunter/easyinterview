# Cascaded Speech Provider Foundation Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-08

**关联计划**: [plan](./plan.md)

## Phase 1: B1 capability 与 A3 schema 对齐

- [x] 1.1 扩展 B1 AI capability 增加 `tts` 并同步 Go/TS/OpenAPI 生成物；验证: B1 spec/history + `shared/conventions.yaml` 同步、Go/TS AI vocabulary parity tests + `make codegen-check`，且 active-scope negative search 证明 `tts` 未被私造为 A3-only 常量
- [x] 1.2 扩展 registry/profile validator 支持 `doubao_speech`、`minimax_speech` 与 `capability=tts`；验证: profile/registry loader tests 覆盖未知 protocol、capability 不匹配、secret missing、unsupported profile

## Phase 2: `Synthesize` provider-neutral 接口

- [x] 2.1 定义 `SynthesisInput` / `SynthesisResponse`，包含文本、voice、format、rate、language、metadata 与音频摘要；验证: interface contract tests 先 Red 后 Green
- [x] 2.2 扩展 `AIClient` / `Provider` / client dispatch 支持 `Synthesize` + `CapabilityTTS`；验证: client-level tests 覆盖 canonical meta merge、fallback、timeout、fail-closed
- [x] 2.3 stub provider 实现 deterministic TTS placeholder；验证: stub deterministic tests 且 local deploy / Kind / staging / prod anti-stub gate 保持通过

## Phase 3: 豆包与 MiniMax speech adapters

- [ ] 3.1 实现 `doubao_speech` STT/TTS adapter；验证: 记录官方 API 文档版本或可审计 contract fixture 来源，mockserver contract tests 覆盖 happy path、4xx/5xx、timeout、secret missing、unsupported audio/text format
- [ ] 3.2 实现 `minimax_speech` TTS adapter，优先 `speech-02-turbo`；验证: 记录官方 API 文档版本或可审计 contract fixture 来源，mockserver contract tests 覆盖 happy path、provider error、secret missing；MiniMax 未确认 STT 前不得声明 `stt`
- [ ] 3.3 增加 provider-specific 不兼容 gate；验证: negative tests 证明豆包/MiniMax speech adapter 不复用 OpenAI Audio Transcriptions wire 假设

## Phase 4: Profile catalog 与独立配置

- [ ] 4.1 新增 `practice.voice.stt.default` 与 `practice.voice.tts.default` profiles；验证: `make lint-ai-profile-coverage` 证明 STT/TTS profile 独立存在
- [ ] 4.2 配置默认 provider 策略：STT 默认豆包；TTS 默认豆包、MiniMax `speech-02-turbo` 作为备选 profile 或 fallback；验证: profile loader tests + config lint
- [ ] 4.3 成本 metadata 支持 STT 时长、TTS 字符数 / duration、cost micros 摘要；验证: meta tests 不含音频或文本明文

## Phase 5: Observability / privacy / failure semantics

- [ ] 5.1 观测埋点支持 `capability=tts` 有界 label；验证: observability focused tests + F1 label gate
- [ ] 5.2 隐私红线覆盖 raw audio、transcript、TTS text/audio、provider secret；验证: privacy tests + active-scope grep gate
- [ ] 5.3 provider/client failure semantics：provider timeout / secret missing / unsupported capability 返回 shared `AI_*` 错误码并记录 meta，`Transcribe` / `Synthesize` 失败不静默回退 stub 或 `realtime`；验证: client/provider tests + shared `AI_*` error code assertions + `practice-voice-mvp/001` BDD handoff 引用完整

## Phase 6: Verification and handoff

- [ ] 6.1 focused Go tests、adapter contract tests、stub deterministic tests、observability privacy tests 全部通过；验证: `cd backend && go test ./internal/ai/aiclient/... -count=1`
- [ ] 6.2 drift gates 与 negative search 通过；验证: `make codegen-check`、`make lint-ai-profile-coverage`、`make lint-config`、`make docs-check`，且 cascade 不标为 `realtime`、不恢复独立 `voice` route、一 profile 一目录 truth source
- [ ] 6.3 向 `practice-voice-mvp/001` handoff profile names、provider/client error semantics、privacy summary 与 cost metadata；验证: 本 plan operation matrix 与 practice plan operation matrix / BDD 引用 A3 004 输出，且 A3 不实现 `PracticeScreen` / `createPracticeVoiceTurn` 业务 orchestration
