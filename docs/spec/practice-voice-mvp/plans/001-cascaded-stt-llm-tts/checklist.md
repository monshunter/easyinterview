# Cascaded STT LLM TTS Voice MVP Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-08

**关联计划**: [plan](./plan.md)

## Phase 1: Contract and fixture

- [ ] 1.1 新增 `createPracticeVoiceTurn` OpenAPI operation 与 schema；验证: `make lint-openapi` + operation matrix 字段完整
- [ ] 1.2 新增 PracticeSessions fixtures（happy path / barge-in / TTS failure / provider config missing）并 codegen；验证: `make validate-fixtures && make codegen-check`
- [ ] 1.3 扩展 generated client allowlist / mock transport 消费路径；验证: frontend contract tests 不允许 ad hoc fetch shape

## Phase 2: Backend orchestration

- [ ] 2.1 实现 voice turn service 串联独立 `stt`、`chat`、`tts` profiles；验证: backend service tests 断言三类 profile 可指向不同 provider
- [ ] 2.2 实现 STT / chat / TTS 独立失败路径；验证: backend tests 断言 STT 失败不调用 chat/TTS，TTS 失败不丢 transcript/chat text
- [ ] 2.3 实现 session event / AI metadata privacy 边界；验证: privacy tests + grep gate 不含 raw audio、TTS audio、provider secret、AI metadata transcript 明文

## Phase 3: Playback progress and barge-in context

- [ ] 3.1 扩展或复用 `appendSessionEvent` 记录 playback / barge-in / committed context events；验证: API/handler tests 覆盖 event ordering 与 idempotency
- [ ] 3.2 实现 committed context builder；验证: unit tests 覆盖完整 chunk、部分 chunk、无播放、重复事件、乱序事件
- [ ] 3.3 下一轮 prompt 注入 interruption note；验证: backend tests 断言未播放 draft 不进入 prompt，已播放内容和用户插话进入 prompt

## Phase 4: Frontend voice controller

- [ ] 4.1 在 `PracticeScreen` 内复刻 ui-design 语音 Surface；验证: source-structure parity tests + visual geometry / existing pixel parity gate，且不新增 `voice` route
- [ ] 4.2 实现音频采集、voice turn 提交、transcript 展示和文本 fallback；验证: frontend component/controller tests 使用 fixtures/stub
- [ ] 4.3 实现 TTS 播放、播放完成回报、barge-in 停止播放；验证: frontend tests 覆盖 played chunk 上报、用户插话、TTS error fallback

## Phase 5: Verification and negative gates

- [ ] 5.1 BDD-Gate: 创建并执行 `E2E.P0.007` 完整语音 turn 场景
- [ ] 5.2 BDD-Gate: 创建并执行 `E2E.P0.008` 插话 / 打断 committed context 场景
- [ ] 5.3 BDD-Gate: 创建并执行 `E2E.P0.009` provider failure / fallback 场景
- [ ] 5.4 重跑 regression gates：OpenAPI fixture validation、codegen drift、frontend tests、backend tests、A3 profile coverage、privacy grep、旧 route negative search；验证: 所有命令记录实际通过证据
