# Cascaded STT LLM TTS Voice MVP Checklist

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 0: Current-state preflight and handoff lock

- [x] 0.1 验证 A3 004 handoff 已在代码中可消费：`tts` capability、`AIClient.Transcribe` / `AIClient.Synthesize`、`practice.voice.stt.default` / `practice.voice.tts.default`、speech adapters、profile coverage 与 privacy tests；验证: focused grep + `cd backend && go test ./internal/ai/aiclient/... -count=1`
- [x] 0.2 固化当前 owner handoff 与待反转负向实现：记录并更新 `backend/internal/api/practice/README.md` 的 voice/audio route owner、`VoiceSurfaceComingSoon` placeholder 测试边界和 `voice` route fallback 负向测试；验证: grep 证明不再存在 non-current voice placeholder owner 口径，且独立 `voice` route 仍 fallback `home`

## 2026-05-17 Review-Fix Evidence

- [x] RF.1 BUG-0070 playback `audioRef` gate：`createPracticeVoiceTurn` HTTP response 的 `ttsChunks[].audioRef` 返回浏览器可播放 `data:audio/...;base64,...`；持久化 voice turn summary 改写为 opaque `voice-turn://...`，不保存 audio bytes；验证: `go test ./internal/practice -run TestCreatePracticeVoiceTurnReturnsPlayableAudioRefWithoutPersistingAudioData -count=1`。
- [x] RF.2 BUG-0070 committed context replay gate：production service 在请求未携带 context 时从 store 加载 latest `follow_up_generated` + 后续 playback events，并注入下一轮 chat payload；验证: `go test ./internal/practice ./internal/store/practice -run 'TestCreatePracticeVoiceTurnLoadsCommittedContext|TestSQLRepositoryLoadCommittedVoiceContext' -count=1`。
- [x] RF.3 BUG-0070 barge-in partial playback gate：frontend `bargeIn()` 先上报 partial `tts_chunk_played`（含 playedTextLength/hash/offset），再上报 `barge_in_detected`；验证: `pnpm --dir frontend test src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx --run`。
- [x] RF.4 BUG-0072 fixture playback ref gate：`createPracticeVoiceTurn` default fixture 的 `ttsChunks[].audioRef` 与真实 HTTP response 语义一致，必须是浏览器可播放 `data:audio/...;base64,...` 或同计划 resolver URL；验证: `pnpm --dir frontend test src/api/devMockClient.test.ts --run` + `python3 scripts/lint/validate_fixtures.py --repo-root .`。
- [x] RF.5 backend-practice non-current lint precision gate：非当前 route 负向 lint 继续拦截独立 `/voice` route / alias，但不得误伤本计划拥有的 `createPracticeVoiceTurn` endpoint、`voice-turn://` persisted summary ref、`practice.voice.*` profile / feature key；验证: `python3 -m pytest scripts/lint/backend_practice_non_current_test.py -q` PASS + `make lint-backend-practice-non-current` PASS + `make lint` PASS。

## Phase 1: Contract and fixture

- [x] 1.1 新增 `createPracticeVoiceTurn` OpenAPI operation 与 schema，锁定 `POST /practice/sessions/{sessionId}/voice-turns`、`Idempotency-Key`、`clientVoiceTurnId`、small audio payload、`userTranscriptFinal`、`assistantTextDraft`、`ttsChunks[]`、`providerMetaSummary` 与 `ttsError`；验证: `make lint-openapi` + operation matrix 字段完整
- [x] 1.2 新增 / 扩展 PracticeSessions fixtures：`createPracticeVoiceTurn` scenarios `default` / `stt-config-missing` / `chat-failed` / `tts-failed`，`appendSessionEvent` scenarios `voice-tts-started` / `voice-tts-played` / `voice-barge-in` / `voice-context-committed`；验证: `make validate-fixtures && make codegen-check`
- [x] 1.3 扩展 generated client allowlist / mock transport 消费路径；验证: frontend contract tests 不允许 ad hoc fetch shape，且未知 `Prefer: example=` voice scenario fail loudly

## Phase 2: Backend orchestration

- [x] 2.1 实现 voice turn service 串联独立 `stt`、`chat`、`tts` profiles；验证: backend service tests 断言三类 profile 可指向不同 provider
- [x] 2.2 实现 STT / chat / TTS 独立失败路径；验证: backend tests 断言 STT 失败不调用 chat/TTS，TTS 失败不丢 transcript/chat text
- [x] 2.3 实现 session event / AI metadata privacy 边界；验证: privacy tests + grep gate 不含 raw audio、TTS audio、provider secret、AI metadata transcript 明文，session event 业务正文与 AI/audit metadata 摘要字段分离

## Phase 3: Playback progress and barge-in context

- [x] 3.1 扩展或复用 `appendSessionEvent` 记录 `tts_chunk_started` / `tts_chunk_played` / `barge_in_detected` / `assistant_context_committed`；验证: API/handler tests 覆盖 event ordering、body-level `clientEventId` replay、禁止 `Idempotency-Key`
- [x] 3.2 实现 committed context builder；验证: unit tests 覆盖完整 chunk、部分 chunk、无播放、重复事件、乱序事件
- [x] 3.3 下一轮 prompt 注入 interruption note；验证: backend tests 断言未播放 draft 不进入 prompt，已播放内容和用户插话进入 prompt

## Phase 4: Frontend voice controller

- [x] 4.1 在 `PracticeScreen` 内复刻 ui-design 语音 Surface，并删除 / 反转 `VoiceSurfaceComingSoon` placeholder 语义；验证: source-structure parity tests + visual geometry / existing pixel parity gate 覆盖 `practice-voice-waveform`、`practice-voice-annotated-waveform`、`practice-voice-expression-panel`，且不新增 `voice` route；证据: `pnpm --dir frontend test src/app/screens/practice/PracticeScreen.test.tsx src/app/screens/practice/__tests__/practiceModeSwitch.test.tsx src/app/App.test.tsx src/app/screens/practice/__tests__/nonCurrentNegative.test.ts`、`pnpm --dir frontend typecheck`、`pnpm --dir frontend build`、`pnpm --dir frontend test:pixel-parity tests/pixel-parity/practice.spec.ts --grep "voice mode"`
- [x] 4.2 实现音频采集、voice turn 提交、transcript 展示和文本 fallback；验证: frontend component/controller tests 使用 fixtures/stub；证据: `pnpm --dir frontend test src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx src/app/screens/practice/__tests__/nonCurrentNegative.test.ts src/app/screens/practice/PracticeScreen.test.tsx src/app/screens/practice/__tests__/practiceModeSwitch.test.tsx src/app/App.test.tsx`、`pnpm --dir frontend typecheck`、`pnpm --dir frontend build`、`pnpm --dir frontend test:pixel-parity tests/pixel-parity/practice.spec.ts --grep "voice mode"`
- [x] 4.3 实现 TTS 播放、播放完成回报、barge-in 停止播放；验证: frontend tests 覆盖 played chunk 上报、用户插话、TTS error fallback；证据: `pnpm --dir frontend test src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx src/app/screens/practice/__tests__/nonCurrentNegative.test.ts src/app/screens/practice/PracticeScreen.test.tsx src/app/screens/practice/__tests__/practiceModeSwitch.test.tsx src/app/App.test.tsx`、`pnpm --dir frontend typecheck`、`pnpm --dir frontend build`、`pnpm --dir frontend test:pixel-parity tests/pixel-parity/practice.spec.ts --grep "voice mode"`

## Phase 5: Verification and negative gates

- [x] 5.1 BDD-Gate: 创建并执行 `E2E.P0.007` 完整语音 turn 场景；证据: `test/scenarios/e2e/p0-007-cascaded-voice-turn/scripts/setup.sh && test/scenarios/e2e/p0-007-cascaded-voice-turn/scripts/trigger.sh && test/scenarios/e2e/p0-007-cascaded-voice-turn/scripts/verify.sh && test/scenarios/e2e/p0-007-cascaded-voice-turn/scripts/cleanup.sh`
- [x] 5.2 BDD-Gate: 创建并执行 `E2E.P0.008` 插话 / 打断 committed context 场景；证据: `test/scenarios/e2e/p0-008-voice-barge-in-committed-context/scripts/setup.sh && test/scenarios/e2e/p0-008-voice-barge-in-committed-context/scripts/trigger.sh && test/scenarios/e2e/p0-008-voice-barge-in-committed-context/scripts/verify.sh && test/scenarios/e2e/p0-008-voice-barge-in-committed-context/scripts/cleanup.sh`
- [x] 5.3 BDD-Gate: 创建并执行 `E2E.P0.009` provider failure / fallback 场景；证据: `test/scenarios/e2e/p0-009-voice-provider-failure-fallback/scripts/setup.sh && test/scenarios/e2e/p0-009-voice-provider-failure-fallback/scripts/trigger.sh && test/scenarios/e2e/p0-009-voice-provider-failure-fallback/scripts/verify.sh && test/scenarios/e2e/p0-009-voice-provider-failure-fallback/scripts/cleanup.sh`
- [x] 5.4 重跑 regression gates：OpenAPI fixture validation、codegen drift、frontend tests、backend tests、A3 profile coverage、privacy grep、非当前 route negative search、scenario wrapper contract；证据: `make codegen-check`、`make validate-fixtures`、`pnpm --filter @easyinterview/frontend test src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx src/api/devMockClient.test.ts src/app/screens/practice/PracticeScreen.test.tsx src/app/screens/practice/__tests__/practiceModeSwitch.test.tsx src/app/App.test.tsx`、`cd backend && go test ./internal/practice ./internal/api/practice ./internal/store/practice ./cmd/api ./internal/ai/aiclient ./internal/ai/aiclient/profile -count=1`、`python3 scripts/lint/ai_profile_coverage.py --repo-root .`、runtime privacy grep 零命中、non-current voice route / coming-soon 负向搜索零命中、`test/scenarios/e2e/p0-007-cascaded-voice-turn/scripts/verify.sh`、`test/scenarios/e2e/p0-008-voice-barge-in-committed-context/scripts/verify.sh`、`test/scenarios/e2e/p0-009-voice-provider-failure-fallback/scripts/verify.sh`、`make docs-check`、`git diff --check`
