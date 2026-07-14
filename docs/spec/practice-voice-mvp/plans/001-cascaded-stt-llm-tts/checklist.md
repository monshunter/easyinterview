# 001 — Practice Voice Disabled Boundary Checklist

> **版本**: 2.0
> **状态**: completed
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md)

## Phase 1: Frontend/prototype
- [x] 1.1 RED-GREEN: phone icon is native disabled with unavailable a11y/copy and no handler/route change.
- [x] 1.2 RED-GREEN: remove PhoneSurface/controllers/hooks/prototype and positive UI tests.

## Phase 2: Backend guard
- [x] 2.1 RED-GREEN: voice endpoint returns AI_UNSUPPORTED_CAPABILITY before audio/provider/store.
- [x] 2.2 RED-GREEN: disabled profiles and zero side effects are proven; fixture is disabled-only.

## Phase 3: Test/docs closeout
- [x] 3.1 删除 voice 场景 wrapper 与 BDD 文档；保留 frontend component 和 backend handler 的代码层测试。
- [x] 3.2 BDD-N/A: 当前电话模式不可进入，不存在真实端到端业务流程。
- [x] 3.3 Run profile/codegen/privacy/negative/docs gates, then run root `make test` for full frontend/backend regression.
