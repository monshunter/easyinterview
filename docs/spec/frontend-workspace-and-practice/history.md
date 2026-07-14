# Frontend Workspace and Practice History

> **版本**: 1.42
> **状态**: active
> **更新日期**: 2026-07-14

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-14 | 1.42 | 用户确认 T-B/P-A：90 秒 backend lease 对应 95 秒 frontend POST timeout + 同 ID 对账；迟到 response 不覆盖新事实；terminal 展示通用 CTA 并精确返回 `parse(targetJobId)` 当前面试规划。 | [002-practice-text-event-loop](./plans/002-practice-text-event-loop/plan.md) Phase 10 + backend-practice/002 Phase 11 |
| 2026-07-13 | 1.41 | 用户消息提交后立即显示 optimistic row；pending/retry 锁定 composer 并显示面试官思考；失败 retry 仅位于原消息底部且复用同一 ID。 | [002-practice-text-event-loop](./plans/002-practice-text-event-loop/plan.md) Phase 9 |
| 2026-07-12 | 1.40 | 原地重开 Practice plan 002：Finish 在零 committed candidate user message 时原生禁用，显示 zh/en 可访问原因；backend 仍以 `VALIDATION_FAILED` 权威拒绝绕过 UI 的零回答完成。 | [002-practice-text-event-loop](./plans/002-practice-text-event-loop/plan.md) Phase 7 + backend-practice/002 Phase 9 |
| 2026-07-12 | 1.39 | 将 GeneratingScreen 唯一 owner 转交 frontend-report-dashboard；workspace/practice 仅负责 completion 的 stable reportId handoff。 | [002-practice-text-event-loop](./plans/002-practice-text-event-loop/plan.md) + frontend-report-dashboard/001 |
| 2026-07-12 | 1.38 | Frontend round normalizer 接受正 int32 严格递增但不连续的 canonical sequence（如 `1,2,4`），next 取下一条现有 round；P0.098 live browser reload/quick-start 在真实执行前保持未完成。 | [001-workspace-and-interview-context](./plans/001-workspace-and-interview-context/plan.md) Phase 25 |
| 2026-07-12 | 1.37 | Home / Workspace / Parse / Report 与 quick-start 统一消费后端 `practiceProgress` 和精确 round pair；缺失或不一致时 fail closed，不再从 TargetJob lifecycle status、本地缓存或全局 latest plan 猜轮次。 | [001-workspace-and-interview-context](./plans/001-workspace-and-interview-context/plan.md) Phase 25 |
| 2026-07-12 | 1.36 | 结构化面试轮次统一驱动 PracticePlan 时间预算、Top Bar 预算显示和报告下一轮推进；末轮、未知轮次、加载失败与重复点击 fail closed。关联 [BUG-0161](../../bugs/BUG-0161.md) 与 [交付复盘](../../reports/2026-07-12-structured-round-runtime-consistency-assessment.md)。 | [001-workspace-and-interview-context](./plans/001-workspace-and-interview-context/plan.md) |
| 2026-07-12 | 1.35 | 重新打开 002：Error/Retry 按 loader/message/completion 来源绑定正确动作，并在发送/加载/完成期间禁用结束 CTA。 | [002-practice-text-event-loop](./plans/002-practice-text-event-loop/plan.md) |
| 2026-07-12 | 1.34 | Practice 改为全宽连续文本会话，删除题目/hint/mode/phone surface，保留 disabled 电话入口并改用会话级 generating 文案。 | [002-practice-text-event-loop](./plans/002-practice-text-event-loop/plan.md) |
| 2026-07-11 | 1.21 | 重新打开 Practice UI：单一电话图标替代分段/live，挂断回同 session 文本且删除重开/callEnded；Top Bar 消费真实 getTargetJob，公司/岗位和会话内容不得由 mock/内部 questionIntent 填充。 | [002-practice-text-event-loop](./plans/002-practice-text-event-loop/plan.md) |
| 2026-07-10 | 1.20 | Spec v1.32 removes the stale embedded company-insight contract and records the current pure Workspace plan-list boundary. | 001-workspace-and-interview-context Phase 23 |
| 2026-07-10 | 1.19 | Spec v1.31 将 workspace / practice 的负向 UI 边界统一为范围外 / out-of-scope 术语；行为不变。 | tech-debt pruning |
| 2026-07-10 | 1.18 | Spec v1.29 将 workspace route purity 与场景负向锚点口径统一为 out-of-scope/stale wording；行为不变。 | 001-workspace-and-interview-context |
| 2026-07-10 | 1.17 | Spec v1.28 将 `voice` route/query 负向输入统一为 out-of-scope 口径；行为不变。 | tech-debt pruning |
| 2026-07-10 | 1.16 | Spec v1.26 将 records row 口径收敛为 disabled handoff row，并移除 sibling plan 空壳表述。 | tech-debt pruning |
| 2026-07-10 | 1.15 | Spec v1.25 uses out-of-scope wording to exclude the `voice` route/query from the phone-entry scope and C-5 acceptance criteria. | tech-debt pruning |
| 2026-07-10 | 1.14 | Spec v1.24 将电话模式正向入口收敛为 `mode=phone&modality=phone`，out-of-scope `mode/modality=voice` 不再作为电话入口。 | tech-debt pruning |
| 2026-07-10 | 1.13 | Spec v1.22 对齐当前 route fallback 命名，正式前端入口改为 `RouteShellScreen`。 | tech-debt pruning |
| 2026-07-09 | 1.12 | Spec v1.13 and plan 001 v1.17 reopen the workspace owner so context-bearing workspace routes reuse the Parse-derived "面试规划详情 / 面试上下文确认" mother page while workspace keeps list landing and start-practice ownership. | 001-workspace-and-interview-context Phase 12 |
| 2026-07-08 | 1.11 | Spec v1.12 and plan 001 v1.14 reopen the workspace owner to simplify interview plan list cards: remove source/language metadata, use theme accent CTA, and strengthen card/page separation after screenshot review. | 001-workspace-and-interview-context Phase 9 |
| 2026-07-08 | 1.10 | Spec v1.11 and plan 001 v1.13 reopen the workspace owner to harden the no-context interview plan list as visible list cards after screenshot review. | 001-workspace-and-interview-context Phase 8 |
| 2026-07-08 | 1.9 | Spec v1.10 and plan 001 v1.12 reopen the workspace owner to make the first-level `面试` entry land on an interview plan list, while context-bearing `workspace` routes continue to render the current plan detail. | 001-workspace-and-interview-context Phase 7 |
| 2026-07-07 | 1.8 | Spec §7 now lists the current completed 001/002 owner plans and routes voice/report/generating responsibility to the active external owner gates. | product-scope/001-core-loop-module-pruning Phase 6.115 |
| 2026-07-06 | 1.7 | Active spec and context discovery use only current route owner, embedded company insight, flat Resume binding, and the three current practice goals. | product-scope/001-core-loop-module-pruning Phase 6.33 |
| 2026-07-06 | 1.6 | Workspace company insight and practice goal contracts aligned with product-scope current loop. | product-scope/001-core-loop-module-pruning Phase 6 |
| 2026-06-13 | 1.5 | Flat Resume binding synchronized across workspace / practice InterviewContext, ResumePicker, route params, and operation matrix. | D-20 contract impl phase |
| 2026-05-23 | 1.4 | L2 real-backend drift 修订：backend-resume / backend-practice / practice-voice / backend-review operation matrix 改为真实 handler + fixture-backed UI variants 双轨；P0.018-P0.021 与 P0.044-P0.047 trigger 前置 `frontendOwners.realApiMode.test.ts`，verify 检查 `VITE_EI_API_MODE=real`、默认 real base URL 与测试文件 marker，防止 fixture UI PASS 被误判为真实 backend 闭环。 | 001-workspace-and-interview-context / 002-practice-text-event-loop |
| 2026-05-14 | 1.3 | Plan 002-practice-text-event-loop 5 个 Phase 全部完成（76 个 checklist 项）：PracticeScreen 静态壳 + 14 components + 5 hooks（`usePracticeSessionLoader / usePracticeEvents / usePracticeAssistance / usePracticeSession / useCompletePracticeSession`）+ AssistantActionRenderer + handoffParams util；i18n `practice.*` namespace 64 keys；InterviewContext reducer 扩展 `INCREMENT_HINT_COUNT`；fixture variants 扩展 `getPracticeSession.queued / running-with-history / completing` 与 `appendSessionEvent.show-hint / ai-timeout`；`newIdempotencyBatch().complete` Idempotency-Key 链；`test/scenarios/e2e/p0-044~047` 4 scenario 落地（Vitest in-process；trigger 跑 practice 套件、verify 反查 IK 双轨 + voice/out-of-scope 负向 grep）；不动 spec 语义。 | 002-practice-text-event-loop |
| 2026-05-13 | 1.3 | L1 plan-review 修订：按当前 `shared/conventions.yaml` / `openapi/openapi.yaml` 把 PracticeSession 消费状态对齐为 `queued / running / waiting_user_input / completing / completed / failed / cancelled` 七值；将 plan 002 scenario 编号改为 `E2E.P0.044 ~ E2E.P0.047`，避开 backend-practice/002 已占用的 `E2E.P0.038 ~ E2E.P0.043`。 | 002-practice-text-event-loop |
| 2026-05-13 | 1.2 | Plan kickoff only（不更新 spec 版本）：按 spec v1.2 §2.1 / §6 C-4 / §7 拆出 `002-practice-text-event-loop` plan，承接 001 已交付的 InterviewContext + 双步启动契约；fixture 扩展计划由 mock-contract-suite + backend-practice/002 同步落地；不修改 spec 语义。 | 002-practice-text-event-loop |
| 2026-05-08 | 1.2 | 与 backend-practice v1.2 对齐：把 frontend practiceMode 收敛为 `assisted/strict` 二值；同时把 `ReportGenerationParams` 改为 route-only `PracticeDisplayContext`，`completePracticeSession` body 严格遵守 B2 `CompletePracticeSessionRequest{clientCompletedAt}`。 | 暂无（spec-only 修订） |
| 2026-05-08 | 1.1 | 终稿候选修订：按 spec review 结论收敛 owner 范围为 `workspace / practice / generating`；补充 route 最小上下文矩阵、OpenAPI operation matrix、UI 真理源 anchor 修正、current-scope negative gate 分类和 acceptance criteria 映射。 | 无计划目录 |
| 2026-05-08 | 1.0 | 初始创建：从 roadmap、shell、Home/Parse、backend practice、voice controller 和 UI module docs 派生新 subspec；定义 workspace / practice / generating 前端 owner 范围、关键决策、设计约束、模块边界和 acceptance criteria。 | 无计划目录 |
