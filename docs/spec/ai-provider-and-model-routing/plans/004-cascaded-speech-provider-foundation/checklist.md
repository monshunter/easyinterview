# Cascaded Speech Provider Foundation Checklist

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Phase 1: B1 capability 与 A3 schema 对齐

- [x] 1.1 扩展 B1 AI capability 增加 `tts` 并同步 Go/TS/OpenAPI 生成物；验证: B1 spec/history + `shared/conventions.yaml` 同步、Go/TS AI vocabulary parity tests + `make codegen-check`，且 active-scope negative search 证明 `tts` 未被私造为 A3-only 常量
- [x] 1.2 扩展 registry/profile validator 支持 `doubao_speech`、`minimax_speech` 与 `capability=tts`；验证: profile/registry loader tests 覆盖未知 protocol、capability 不匹配、secret missing、unsupported profile

## Phase 2: `Synthesize` provider-neutral 接口

- [x] 2.1 定义 `SynthesisInput` / `SynthesisResponse`，包含文本、voice、format、rate、language、metadata 与音频摘要；验证: interface contract tests 先 Red 后 Green
- [x] 2.2 扩展 `AIClient` / `Provider` / client dispatch 支持 `Synthesize` + `CapabilityTTS`；验证: client-level tests 覆盖 canonical meta merge、fallback、timeout、fail-closed
- [x] 2.3 stub provider 实现 deterministic TTS fixture；验证: stub deterministic tests 且非测试本地 app run / future deploy anti-stub gate 保持通过

## Phase 3: 豆包与 MiniMax speech adapters

- [x] 3.1 实现 `doubao_speech` STT/TTS adapter；验证: 记录官方 API 文档版本或可审计 contract fixture 来源，mockserver contract tests 覆盖 happy path、4xx/5xx、timeout、secret missing、unsupported audio/text format
- [x] 3.2 实现 `minimax_speech` TTS adapter，优先 `speech-02-turbo`；验证: 记录官方 API 文档版本或可审计 contract fixture 来源，mockserver contract tests 覆盖 happy path、provider error、secret missing；MiniMax 未确认 STT 前不得声明 `stt`
- [x] 3.3 增加 provider-specific 不兼容 gate；验证: negative tests 证明豆包/MiniMax speech adapter 不复用 OpenAI Audio Transcriptions wire 假设

## Phase 4: Profile catalog 与独立配置

- [x] 4.1 新增 `practice.voice.stt.default` 与 `practice.voice.tts.default` profiles；验证: `make lint-ai-profile-coverage` 证明 STT/TTS profile 独立存在
- [x] 4.2 配置默认 provider 策略：STT 默认豆包；TTS 默认豆包、MiniMax `speech-02-turbo` 作为备选 profile 或 fallback；验证: profile loader tests + config lint
- [x] 4.3 成本 metadata 支持 STT 时长、TTS 字符数 / duration、cost micros 摘要；验证: meta tests 不含音频或文本明文

## Phase 5: Observability / privacy / failure semantics

- [x] 5.1 观测埋点支持 `capability=tts` 有界 label；验证: observability focused tests + F1 label gate
- [x] 5.2 隐私红线覆盖 raw audio、transcript、TTS text/audio、provider secret；验证: privacy tests + active-scope grep gate
- [x] 5.3 provider/client failure semantics：provider timeout / secret missing / unsupported capability 返回 shared `AI_*` 错误码并记录 meta，`Transcribe` / `Synthesize` 失败不静默回退 stub 或 `realtime`；验证: client/provider tests + shared `AI_*` error code assertions + `practice-voice-mvp/001` BDD handoff 引用完整

## Phase 6: Verification and handoff

- [x] 6.1 focused Go tests、adapter contract tests、stub deterministic tests、observability privacy tests 全部通过；验证: `cd backend && go test ./internal/ai/aiclient/... -count=1`
- [x] 6.2 drift gates 与 negative search 通过；验证: `make codegen-check`、`make lint-ai-profile-coverage`、`make lint-config`、`make docs-check`，且 cascade 仅使用 `stt` / `chat` / `tts` capability、顶层 `voice` route 缺席、profile truth source 保持单一 catalog
- [x] 6.3 向 `practice-voice-mvp/001` handoff profile names、provider/client error semantics、privacy summary 与 cost metadata；验证: 本 plan operation matrix 与 practice plan operation matrix / BDD 引用 A3 004 输出，且 A3 不实现 `PracticeScreen` / `createPracticeVoiceTurn` 业务 orchestration

## Phase 7: Doubao speech trivial wrapper removal

- [x] 7.1 Structural red: `util.go` 的三个 helper 都是单调用标准库转发且各只有一个生产调用点；验证: `rg` 命中 `encodeBase64Audio` / `decodeBase64Audio` / `readAll` 定义与各一处调用，contract tests 已覆盖 wire 行为
- [x] 7.2 Green: adapter 直接调用 `encoding/base64` 与 `io`，删除 `util.go`，不创建替代 helper
  <!-- verified: 2026-07-10 command="cd backend && go test ./internal/ai/aiclient/providers/doubao_speech -count=1" result="pass; wrapper search zero; util.go absent" -->
- [x] 7.3 Verify/closure: 运行 Doubao focused/package、A3 aiclient、lint/context/docs/diff/pruning gates，wrapper/file 负向搜索为零，并确认状态为 `completed`
  <!-- verified: 2026-07-10 commands="go test ./internal/ai/aiclient/... -count=1; make lint; validate A3 004 context; docs/index/diff/pruning gates" result="pass; wrapper search zero; util.go absent; real_residuals=0" -->

## Phase 8: MiniMax speech dead wrapper and return removal

- [x] 8.1 Structural red: `encodeBase64Audio` 零调用，`decodeBase64Audio` 单次标准库转发，`postJSON` headers 只被 `_ = headers` 丢弃；现有 MiniMax contract tests 覆盖 TTS decode/error/timeout
- [x] 8.2 Green: 直接调用 base64 decoder，删除两个 wrapper，并把 `postJSON` 收窄为 body/status/error 三返回值
  <!-- verified: 2026-07-10 command="cd backend && go test ./internal/ai/aiclient/providers/minimax_speech -count=1" result="pass; wrapper and unused-header searches zero" -->
- [x] 8.3 Verify/closure: 运行 MiniMax/A3/lint/context/docs/diff/pruning gates，wrapper/unused-header 负向搜索为零，并确认状态为 `completed`
  <!-- verified: 2026-07-10 commands="go test ./internal/ai/aiclient/... -count=1; make lint; validate A3 004 context; docs/index/diff/pruning gates" result="pass; wrapper/unused-header searches zero; real_residuals=0" -->

## Phase 9: Speech adapter dead parameter removal

- [x] 9.1 Structural red: 两家 `errMeta.msg`、Doubao `buildMeta.headers/contentType` 均未使用，Doubao `postJSON` headers 无真实 consumer；现有 provider contract tests 覆盖响应和 meta 行为
- [x] 9.2 Green: 删除上述死参数/实参，把 Doubao `postJSON` 收窄为 body/status/error，保持业务响应与 error/meta 行为不变
  <!-- verified: 2026-07-10 command="cd backend && go test ./internal/ai/aiclient/providers/doubao_speech ./internal/ai/aiclient/providers/minimax_speech -count=1" result="pass; old signature search zero" -->
- [x] 9.3 Verify/closure: 运行 speech/A3/lint/context/docs/diff/pruning gates，死参数/旧签名负向搜索为零，并确认状态为 `completed`
  <!-- verified: 2026-07-10 commands="go test ./internal/ai/aiclient/... -count=1; make lint; validate A3 004 context; docs/index/diff/pruning gates" result="pass; dead-parameter/old-signature searches zero; real_residuals=0" -->
