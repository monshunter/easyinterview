# Cascaded Speech Provider Foundation

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-21

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

为 P0 语音面试 MVP 落地低成本 `stt -> chat -> tts` 级联底座，替代 S2S / realtime voice 作为首发方案。本计划只 owns A3 provider foundation：`tts` capability、`Synthesize` provider-neutral 接口、豆包 / MiniMax provider-specific speech adapter、独立 STT/TTS profile 配置、观测 / 隐私 / 成本 gate。

本计划不实现用户可见 `PracticeScreen` 语音面试流程、不新增 HTTP API handler、不实现媒体留存和删除链路；这些由 `practice-voice-mvp/001-cascaded-stt-llm-tts` 承接。

## 2 背景

用户明确确认 S2S 成本过高，P0 MVP 采用级联方案。关键设计决策是 `stt`、`chat`、`tts` 三个能力独立配置，不绑定同一家 provider；例如 `practice.voice.stt.default` 可指向豆包 STT，`practice.followup.default` 继续走 DeepSeek，`practice.voice.tts.default` 可默认豆包 TTS 并把 MiniMax `speech-02-turbo` 作为高音色质量备选或 fallback。

现有 A3 已具备 `Transcribe` 和 OpenAI-compatible `/v1/audio/transcriptions` 底座，但当前 repo-tracked profile 因 DeepSeek 不提供 STT 仍 fail-closed。语音 provider 的 STT/TTS API 不能默认假设 OpenAI-compatible，因此本计划新增 provider-specific protocol，而不是把豆包 / MiniMax 强行塞进 `openai_compatible` adapter。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + contract + provider-adapter + config + observability/privacy`。
- **TDD 策略**: 通过 `/implement` -> `/tdd` 顺序执行。每个非文档 checklist item 必须先有 Red test 或 drift gate，再实现 B1 capability/codegen、A3接口、provider adapter、profile loader、observability/privacy 与 lint gate。
- **BDD 策略**: BDD 不适用。本计划只打开内部 provider foundation，不直接产生用户行为；用户可见语音面试闭环由 `practice-voice-mvp/001` 的 BDD 覆盖。
- **替代验证 gate**: Go interface / adapter contract tests、stub deterministic tests、profile coverage lint、config lint、conventions/codegen drift gate、observability privacy tests、active-scope negative search。

## 4 Coverage Matrix

| Row | Category | Source | Phase | Verification | Negative scope |
|-----|----------|--------|-------|--------------|----------------|
| A3-004-C1 | Cross-layer contract | A3 spec D-12/D-15 | Phase 1 | B1 Go/TS AI vocabulary parity tests + `make codegen-check` | `tts` 只在 A3 私有常量中出现；绕过 B1/shared conventions |
| A3-004-C2 | Primary path | A3 spec C-16 | Phase 2 | `AIClient.Synthesize` interface tests + stub deterministic test | TTS 混入 `Complete` 或 `Transcribe` |
| A3-004-C3 | Provider contract | A3 spec D-9/D-14 | Phase 3 | doubao/minimax adapter mockserver tests | 把豆包/MiniMax speech 当作 OpenAI-compatible |
| A3-004-C4 | Alternate path | 用户确认 STT/TTS 独立配置 | Phase 4 | profile coverage lint + loader tests | STT/TTS 共享同一 provider 必填 |
| A3-004-C5 | Failure/recovery | A3 spec C-17 | Phase 5 | timeout/secret missing/provider error tests + practice voice handoff reference | A3 provider error 被吞掉；业务 orchestration 在 A3 内实现 |
| A3-004-C6 | Privacy/security/observability | ADR-Q5 + A3 §4.3 | Phase 5 | observability privacy tests + grep gate | audio bytes、transcript、TTS text/audio 明文进 log/DB/metric |
| A3-004-C7 | Regression/legacy-negative | A3 spec D-15 | Phase 6 | active-scope negative search | cascade 被标为 `realtime`；恢复独立 `voice` route；恢复一 profile 一文件 truth source |

### 4.1 Operation Matrix

本计划不新增用户可见 HTTP API，但会修改 AI provider foundation、B1 共享 vocabulary、A3 config catalog 与 `practice-voice-mvp/001` 的 AI dependency handoff。按 `docs/development.md` §2.1，本 matrix 用 `N/A` 明确非 HTTP 边界，并把 downstream API owner 与 scenario gate 分离。

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `N/A - aiCapabilityVocabulary` | `shared/conventions.yaml` + generated Go/TS/OpenAPI artifacts | none | B1 generated vocabulary consumed by `backend/internal/shared/ai` and A3 | none | `capability=tts` vocabulary only; no provider call | B1 parity tests + `make codegen-check` |
| `N/A - aiClientSynthesize` | A3 Go contract fixtures / stub deterministic fixture | none | `backend/internal/ai/aiclient` `AIClient.Synthesize` + provider dispatch | none | `practice.voice.tts.default` capability contract; provider-neutral input/output and `AICallMeta` summary | focused Go tests + observability/privacy tests |
| `N/A - speechProviderProtocols` | adapter mockserver fixtures for `doubao_speech` / `minimax_speech` | none | `backend/internal/ai/aiclient/providers` speech adapters | none | provider-specific STT/TTS or TTS protocol; official docs / captured contract fixture required before adapter Green | adapter contract tests + provider-specific negative tests |
| `N/A - voiceProfileCatalog` | `config/ai-providers.yaml` + `config/ai-profiles.yaml` | none | provider registry / profile loader | none | `practice.voice.stt.default` + `practice.voice.tts.default`; chat remains `practice.followup.default` | `make lint-ai-profile-coverage` + config lint |
| `createPracticeVoiceTurn` downstream handoff | `openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json` owned by `practice-voice-mvp/001` | `PracticeScreen` voice turn controller, owned by `practice-voice-mvp/001` | `backend/internal/practice` voice turn handler/service, owned by `practice-voice-mvp/001` | session events and optional transient audio metadata, owned by `practice-voice-mvp/001` | A3 004 supplies profile names, provider/client error semantics, privacy summary, and cost metadata; full STT -> chat -> TTS orchestration is not implemented in A3 | `E2E.P0.007` / `E2E.P0.009` in `practice-voice-mvp/001` |

## 5 实施步骤

### Phase 1: B1 capability 与 A3 schema 对齐

#### 1.1 扩展 B1 AI capability

在 `shared/conventions.yaml` 与生成物中新增 `tts` capability，保持 Go/TS/OpenAPI 共享枚举一致。不得只在 A3 Go 包内私造 `tts` 字符串。

#### 1.2 扩展 provider registry / profile validator

让 A3 provider registry 接受 `doubao_speech` / `minimax_speech` protocol，并让 profile loader 接受 `capability=tts`。负向 fixture 覆盖未知 protocol、provider capability 不匹配、TTS profile 缺 secret 和 unsupported profile 被调用。

### Phase 2: `Synthesize` provider-neutral 接口

#### 2.1 定义输入输出模型

新增 `SynthesisInput` / `SynthesisResponse`，输入包含文本、voice、format、speaking rate、language 与 `CallMetadata`；输出包含音频 bytes 或 chunk metadata、content type、duration / character count 摘要。原始文本和音频不得写入观测明文字段。

#### 2.2 AIClient / Provider interface 扩展

在 `AIClient` 与 `Provider` 增加 `Synthesize`，client dispatch 要求 `CapabilityTTS`，并复用 canonical meta merge、fallback、timeout 与 fail-closed 语义。

#### 2.3 Stub provider deterministic TTS

stub provider 为 TTS 返回 deterministic audio placeholder 与 meta 摘要，仅用于单元测试 / 离线契约测试，不允许 local deploy / Kind / staging / prod 静默回退到 stub。

### Phase 3: 豆包与 MiniMax speech adapters

#### 3.1 豆包 speech adapter

实现 `doubao_speech` adapter 的 STT / TTS provider-specific wire。STT 优先支持 P0 需要的音频片段转写；TTS 支持文本合成与音频格式配置。实施前必须记录官方 API 文档版本或可审计的 contract fixture 来源；契约测试覆盖成功、provider 4xx/5xx、超时、secret missing、格式不支持。

#### 3.2 MiniMax speech adapter

实现 `minimax_speech` adapter 的 TTS wire，优先支持 `speech-02-turbo`。实施前必须记录官方 API 文档版本或可审计的 contract fixture 来源。本计划不默认打开 MiniMax STT；除非公开文档、权限和契约测试确认，否则 MiniMax provider 只声明 `tts` capability。

#### 3.3 Provider-specific 不兼容 gate

新增测试和文档负向断言：speech adapter 不复用 OpenAI Audio Transcriptions wire 假设；每家 provider 的鉴权、请求、流式/非流式输出和错误码映射必须在 adapter 内部完成。

### Phase 4: Profile catalog 与独立配置

#### 4.1 新增 voice MVP profiles

在 `config/ai-profiles.yaml` 新增 `practice.voice.stt.default`、`practice.voice.tts.default`，保留 `practice.followup.default` / `practice.first_question.default` 作为 chat profile。STT 与 TTS 必须独立 profile，可选择不同 provider。

#### 4.2 默认 provider 策略

建议默认 `practice.voice.stt.default` 指向豆包 STT；`practice.voice.tts.default` 默认豆包 TTS，MiniMax `speech-02-turbo` 作为高质量备选 profile 或受限 fallback。实际 active / disabled 状态由 secret 可用性与 release gate 决定。

#### 4.3 成本 metadata

profile 与 meta 必须能记录 STT 时长、TTS 字符数 / duration、provider/model、cost summary。成本估算只写摘要和数值，不写音频或文本明文。

### Phase 5: Observability / privacy / failure semantics

#### 5.1 观测埋点

复用 A3 7 个 metric family 并确保 `capability=tts` label 有界。若 F1 需要新增 speech-specific dashboard label，先修订 F1 spec，再实现。

#### 5.2 隐私红线

log、DB metadata、audit、metric label 不得出现 raw audio、transcript 明文、TTS 输入文本、TTS 输出音频、provider secret。允许 hash、长度、duration、content type、profile、provider、model、cost micros。

#### 5.3 Failure isolation

A3 owns provider/client failure semantics：provider timeout / secret missing / unsupported capability 必须返回共享 `AI_*` 错误码并记录 meta；`Transcribe` / `Synthesize` 失败不得吞掉 provider error、不得静默回退到 stub 或 `realtime`。完整 `STT -> chat -> TTS` orchestration（STT 失败不调用 chat/TTS、TTS 失败不丢 transcript/chat text）由 `practice-voice-mvp/001` service tests 与 BDD gate 验证，本计划只提供可消费的 error semantics 与 handoff matrix。

### Phase 6: Verification and handoff

#### 6.1 Focused tests

运行 A3 focused Go tests、adapter contract tests、observability privacy tests、profile loader tests、stub deterministic tests。

#### 6.2 Drift gates

运行 `make codegen-check`、`make lint-ai-profile-coverage`、`make lint-config`、`make docs-check`，并记录 negative search：不得把 cascade 标成 `realtime`，不得恢复独立 `voice` route，不得恢复旧 provider key 或一 profile 一目录 truth source。

#### 6.3 Handoff to practice voice MVP

把可消费的 profile name、error semantics、privacy summary 与 cost metadata 写入 `practice-voice-mvp/001` operation matrix。用户可见 API / BDD 不在本计划收口。

## 6 验收标准

- B1 已生成 `tts` capability，Go/TS parity test 与 codegen drift gate 通过。
- `AIClient.Synthesize`、stub provider、豆包 STT/TTS、MiniMax TTS adapter 均有 contract tests 覆盖成功与错误路径。
- `practice.voice.stt.default` 与 `practice.voice.tts.default` 独立存在，profile coverage lint 证明 STT/TTS 不绑定同一 provider。
- 观测与 audit 只写 hash / 长度 / duration / profile / provider / cost 摘要，不写音频、转写或 TTS 文本明文。
- realtime S2S 继续 fail-closed，任何用户可见 voice workflow 由 `practice-voice-mvp` plan 的 BDD gate 覆盖。

## 7 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 把级联语音误实现成 realtime S2S | `tts` 独立 capability；negative search 阻止 cascade profile 标成 `realtime` |
| STT/TTS 配置被强行绑定同一 provider | profile schema 与 lint gate 要求 `practice.voice.stt.default` / `practice.voice.tts.default` 独立存在 |
| 语音 provider API 差异被 OpenAI-compatible 假设吞掉 | 新增 `doubao_speech` / `minimax_speech` provider-specific adapters 与 contract tests |
| TTS 失败导致文本对话丢失 | failure isolation tests 断言 chat text 可返回，TTS error 只影响 audio |
| 语音明文泄漏 | observability privacy tests + grep gate 覆盖 raw audio、transcript、TTS text/audio |
