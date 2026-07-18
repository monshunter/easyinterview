# Backend Practice History

> **版本**: 1.36
> **状态**: active
> **更新日期**: 2026-07-18

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-18 | 1.36 | 用户批准方案 A：关闭未对话的面试窗口后，再次开始同一 plan 应恢复服务端既有 queued/running session；不清理、不取消、不重复生成 opening，并用现有活动会话做 Chrome 真实验收。 | [001 Phase 9](./plans/001-plan-and-session-orchestration/plan.md) |
| 2026-07-15 | 1.35 | 用户确认删除无产品入口的 `listPracticeSessions`；保留 live `getPracticeSession`，完成会话复盘改由 report-owned `getReportConversation` 承接，无兼容 route/handler/fixture。 | [001](./plans/001-plan-and-session-orchestration/plan.md) + backend-review/001 + openapi/001/002/003 |
| 2026-07-14 | 1.33 | 用户确认 T-B/P-A：pending 使用 90 秒服务端 lease 与内部 generation fence；GET / 同 ID reserve 惰性收敛；客户端 95 秒超时后按同 ID 对账；terminal 恢复返回当前面试规划。 | [002](./plans/002-event-loop-and-completion/plan.md) + frontend-workspace-and-practice/002 |
| 2026-07-13 | 1.32 | 用户确认方案 A：Practice user message 持久化 reply status，并在会话读模型返回原 `clientMessageId/replyStatus`，支持刷新后同 ID 恢复且不以浏览器存储为事实源。 | [002](./plans/002-event-loop-and-completion/plan.md) + openapi-v1-contract/001 + frontend-workspace-and-practice/002 |
| 2026-07-12 | 1.31 | 完成 004：active v0.2 semantic focus 限定为 code+label+issues，空 focus 不伪造，raw/anchor/code-only fail closed；P0.070/P0.072 PostgreSQL v19、IK/isolation/privacy 与 legacy-negative markers 闭环。 | [004](./plans/004-report-derived-practice-plans/plan.md) + F3/002 |
| 2026-07-12 | 1.30 | 方案 A 将 backend-practice 的结构化 `semanticFocus` runtime payload 与 F3/002-owned immutable practice v0.2 pair、8-status/000019 激活边界对齐。 | [004](./plans/004-report-derived-practice-plans/plan.md) + F3/002 |
| 2026-07-12 | 1.29 | 将零回答/pending-reply completion 拒绝与 `report-context.v1` 原子快照收口到 002 唯一 owner；004 允许空 focus 作为通用同轮复练，非空 focus 保持 issue-backed。 | [002](./plans/002-event-loop-and-completion/plan.md) + [004](./plans/004-report-derived-practice-plans/plan.md) |
| 2026-07-12 | 1.28 | Report-derived retry focus 改为服务端投影 report-local dimension codes；客户端 focus 输入删除，completion 冻结 report context。 | [004](./plans/004-report-derived-practice-plans/plan.md) + [002](./plans/002-event-loop-and-completion/plan.md) |
| 2026-07-12 | 1.27 | Practice 事实来源收紧为 persisted resume 与 candidate-authored user message；assistant history 仅保持连续性，不能把上一轮模型臆造转化为后续履历事实。 | [001](./plans/001-plan-and-session-orchestration/plan.md) |
| 2026-07-12 | 1.26 | CreatePlan、source report 与 completion 台账统一约束为 TargetJob 绑定 resume；canonical round 增加非空 provenance、小写 type allowlist、正 int32 严格递增但可不连续的约束；Practice prompt 分离 system policy 与 JSON 编码的不可信 JD/简历/历史，persona 只控制风格。 | [001](./plans/001-plan-and-session-orchestration/plan.md) / [002](./plans/002-event-loop-and-completion/plan.md) |
| 2026-07-12 | 1.25 | PracticePlan 持久化规范化 `roundId + roundSequence`；baseline / retry / next 由完成 session 台账和 TargetJob canonical rounds 校验，真实 round name/type/focus 注入 AI 上下文。 | [001](./plans/001-plan-and-session-orchestration/plan.md) / [002](./plans/002-event-loop-and-completion/plan.md) |
| 2026-07-12 | 1.24 | 会话 AI grounding 改为完整简历正文优先，正文为空时 fail closed；提示词要求问题必须能引用简历或 JD 证据，禁止臆造项目。 | [001-plan-and-session-orchestration](./plans/001-plan-and-session-orchestration/plan.md) |
| 2026-07-12 | 1.23 | 重新打开 002：assistant reply commit 仅允许 mutable session；completion 赢得竞态后迟到 reply 必须回滚，P0.046/P0.047 必须执行对应失败恢复断言。 | [002-event-loop-and-completion](./plans/002-event-loop-and-completion/plan.md) |
| 2026-07-12 | 1.22 | Practice 收敛为连续 message conversation：删除题目/turn/hint/mode 合同，暂停改为前端本地状态，电话模式 fail-closed。 | [001](./plans/001-plan-and-session-orchestration/plan.md) / [002](./plans/002-event-loop-and-completion/plan.md) / [003](./plans/003-mode-policies-and-provenance/plan.md) |
| 2026-07-11 | 1.19 | 重新打开 003：用户可见 hint cue 必须匹配 persisted session language，错误语言输出按既有 invalid-output graceful degrade 返回 `session_wait`，禁止混合语言提示落库或回显。 | [003-mode-policies-and-provenance](./plans/003-mode-policies-and-provenance/plan.md) |
| 2026-07-11 | 1.18 | 重新打开 002：文本/电话追问统一使用 server-owned canonical context 与 session language，structured output 恰好 repair 一次；二次失败分别返回 retryable session_wait 或既有 voice 错误，禁止 canned English question。 | [002-event-loop-and-completion](./plans/002-event-loop-and-completion/plan.md) |
| 2026-07-07 | 1.17 | 压缩 active spec 为当前 PracticePlans / PracticeSessions / VoiceTurn 合同；001 计划完成 flat Resume `resumeId` 绑定与首题 `resumes.structured_profile` prompt context。 | [001-plan-and-session-orchestration](./plans/001-plan-and-session-orchestration/plan.md) |
| 2026-07-07 | 1.16 | 002 文档收敛为当前 text event loop、completion handoff、source-event-only report job 与双轨幂等合同。 | [002-event-loop-and-completion](./plans/002-event-loop-and-completion/plan.md) |
| 2026-07-06 | 1.15 | 004 owner 路径与正文收敛为当前 `sourceReportId` 派生计划合同。 | [004-report-derived-practice-plans](./plans/004-report-derived-practice-plans/plan.md) |
| 2026-06-29 | 1.13 | PracticeGoal 收敛为 `baseline` / `retry_current_round` / `next_round`。 | product-scope/001-core-loop-module-pruning |
| 2026-05-15 | 1.9 | 003 增补 `show_hint` replay 不变量与 hint lifecycle 边界。 | [003-mode-policies-and-provenance](./plans/003-mode-policies-and-provenance/plan.md) |
| 2026-05-13 | 1.7 | 002 落地 append event loop、complete queued report/job handoff、practice turn/session outbox 与双轨幂等。 | [002-event-loop-and-completion](./plans/002-event-loop-and-completion/plan.md) |
| 2026-05-09 | 1.4 | 001 派生 baseline plan/session foundation、shared idempotency、PracticeMode 二值和 first-question AI flow。 | [001-plan-and-session-orchestration](./plans/001-plan-and-session-orchestration/plan.md) |
