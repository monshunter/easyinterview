# Frontend Workspace and Practice History

> **版本**: 1.11
> **状态**: active
> **更新日期**: 2026-07-08

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-08 | 1.11 | Spec v1.12 and plan 001 v1.14 reopen the workspace owner to simplify interview plan list cards: remove source/language metadata, use theme accent CTA, and strengthen card/page separation after screenshot review. | 001-workspace-and-interview-context Phase 9 |
| 2026-07-08 | 1.10 | Spec v1.11 and plan 001 v1.13 reopen the workspace owner to harden the no-context interview plan list as visible list cards after screenshot review. | 001-workspace-and-interview-context Phase 8 |
| 2026-07-08 | 1.9 | Spec v1.10 and plan 001 v1.12 reopen the workspace owner to make the first-level `面试` entry land on an interview plan list, while context-bearing `workspace` routes continue to render the current plan detail. | 001-workspace-and-interview-context Phase 7 |
| 2026-07-07 | 1.8 | Spec §7 now lists the current completed 001/002 owner plans and routes voice/report/generating responsibility to the active external owner gates. | product-scope/001-core-loop-module-pruning Phase 6.115 |
| 2026-07-06 | 1.7 | Active spec and context discovery use only current route owner, embedded company insight, flat Resume binding, and the three current practice goals. | product-scope/001-core-loop-module-pruning Phase 6.33 |
| 2026-07-06 | 1.6 | Workspace company insight and practice goal contracts aligned with product-scope current loop. | product-scope/001-core-loop-module-pruning Phase 6 |
| 2026-06-13 | 1.5 | Flat Resume binding synchronized across workspace / practice InterviewContext, ResumePicker, route params, and operation matrix. | D-20 contract impl phase |
| 2026-05-23 | 1.4 | L2 real-backend drift 修订：backend-resume / backend-practice / practice-voice / backend-review operation matrix 改为真实 handler + fixture-backed UI variants 双轨；P0.018-P0.021 与 P0.044-P0.047 trigger 前置 `frontendOwners.realApiMode.test.ts`，verify 检查 `VITE_EI_API_MODE=real`、默认 real base URL 与测试文件 marker，防止 fixture UI PASS 被误判为真实 backend 闭环。 | 001-workspace-and-interview-context / 002-practice-text-event-loop |
| 2026-05-14 | 1.3 | Plan 002-practice-text-event-loop 5 个 Phase 全部完成（76 个 checklist 项）：PracticeScreen 静态壳 + 14 components + 5 hooks（`usePracticeSessionLoader / usePracticeEvents / usePracticeAssistance / usePracticeSession / useCompletePracticeSession`）+ AssistantActionRenderer + handoffParams util；i18n `practice.*` namespace 64 keys；InterviewContext reducer 扩展 `INCREMENT_HINT_COUNT`；fixture variants 扩展 `getPracticeSession.queued / running-with-history / completing` 与 `appendSessionEvent.show-hint / ai-timeout`；`newIdempotencyBatch().complete` Idempotency-Key 链；`test/scenarios/e2e/p0-044~047` 4 scenario 落地（Vitest in-process；trigger 跑 practice 套件、verify 反查 IK 双轨 + voice/non-current 负向 grep）；不动 spec 语义。 | 002-practice-text-event-loop |
| 2026-05-13 | 1.3 | L1 plan-review 修订：按当前 `shared/conventions.yaml` / `openapi/openapi.yaml` 把 PracticeSession 消费状态对齐为 `queued / running / waiting_user_input / completing / completed / failed / cancelled` 七值；将 plan 002 scenario 编号改为 `E2E.P0.044 ~ E2E.P0.047`，避开 backend-practice/002 已占用的 `E2E.P0.038 ~ E2E.P0.043`。 | 002-practice-text-event-loop |
| 2026-05-13 | 1.2 | Plan kickoff only（不更新 spec 版本）：按 spec v1.2 §2.1 / §6 C-4 / §7 拆出 `002-practice-text-event-loop` plan，承接 001 已交付的 InterviewContext + 双步启动契约；fixture 扩展计划由 mock-contract-suite + backend-practice/002 同步落地；不修改 spec 语义。 | 002-practice-text-event-loop |
| 2026-05-08 | 1.2 | 与 backend-practice v1.2 对齐：把 frontend practiceMode 收敛为 `assisted/strict` 二值；同时把 `ReportGenerationParams` 改为 route-only `PracticeDisplayContext`，`completePracticeSession` body 严格遵守 B2 `CompletePracticeSessionRequest{clientCompletedAt}`。 | 暂无（spec-only 修订） |
| 2026-05-08 | 1.1 | 终稿候选修订：按 spec review 结论收敛 owner 范围为 `workspace / practice / generating`；补充 route 最小上下文矩阵、OpenAPI operation matrix、UI 真理源 anchor 修正、current-scope negative gate 分类和 acceptance criteria 映射。 | 无计划目录 |
| 2026-05-08 | 1.0 | 初始创建：从 roadmap、shell、Home/Parse、backend practice、voice controller 和 UI module docs 派生新 subspec；定义 workspace / practice / generating 前端 owner 范围、关键决策、设计约束、模块边界和 acceptance criteria。 | 无计划目录 |
