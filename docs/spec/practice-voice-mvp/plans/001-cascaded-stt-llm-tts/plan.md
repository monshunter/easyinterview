# 001 — Practice Voice Disabled Boundary

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-12

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把既有电话模式正向合同原地收敛为 disabled boundary：前端只显示置灰不可点击图标；phone params 不进入 PhoneSurface；后端 voice endpoint 在任何 provider/store 操作前返回 `AI_UNSUPPORTED_CAPABILITY`。通用 speech adapters/profiles 保留 disabled/unsupported。

## 2 Operation Matrix

| operationId | fixture | frontend | backend | persistence | AI | scenario |
|-------------|---------|----------|---------|-------------|----|----------|
| `createPracticeVoiceTurn` | disabled-only | none | leading fail-closed guard | none | none | E2E.P0.007 |

## 3 质量门禁分类

- **Plan 类型**: feature-disable + contract + frontend/backend cleanup。
- **TDD 策略**: Red tests first require disabled DOM and zero backend downstream calls/writes; then remove positive controllers/events/fixtures.
- **BDD 策略**: P0.007 is the single disabled negative scenario; P0.008/P0.009 positive voice scenarios are deleted.
- **替代验证 gate**: profile lint, OpenAPI fixture, frontend negative search, backend spy/store tests.

## 4 Coverage Matrix

| Behavior | Category | Phase | Verification | Negative |
|----------|----------|-------|--------------|----------|
| disabled icon | UX/a11y | 1 | component/parity + P0.007 | click/route/PhoneSurface |
| backend fail-closed | failure/security | 2 | handler/service spies + P0.007 | audio decode/provider/store |
| profiles disabled | config | 2 | profile lint | active STT/TTS/realtime |
| stale positive assets | regression | 3 | zero-reference | VAD/TTS/barge-in/captions/hangup |

## 5 实施步骤

### Phase 1: Frontend/prototype disabled UI
- Native disabled icon, unavailable copy, no handler, no phone route state.
- Delete PhoneSurface/controllers/hooks/tests and positive prototype behavior.

### Phase 2: Backend guard
- Fail before audio/profile/provider/store with typed unsupported.
- Keep STT/TTS/realtime profiles disabled/unsupported.
- Reduce fixture to disabled error.

### Phase 3: Scenario/docs cleanup
- Rewrite P0.007; delete P0.008/P0.009 and positive voice current docs/index rows.
- Run full negative/profile/privacy/codegen/parity gates.

## 6 验收标准

- User cannot enter phone mode by click, keyboard, URL/query or stale context.
- API cannot reach STT/chat/TTS/store while disabled.
- Generic speech foundation remains internally tested but is not a product release gate.

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 2.0 | Replace cascaded voice happy path with explicit disabled boundary. |
