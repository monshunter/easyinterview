# E2E Scenarios P0 Spec

> **版本**: 2.0
> **状态**: completed
> **更新日期**: 2026-07-12

## 1 目标

为当前核心漏斗维护两层证据：可重复的跨层契约组合门禁，以及共享真实环境中的浏览器验收。当前核心漏斗是：简历与 JD 就绪 → 创建 practice plan → 连续聊天 → 完成会话 → 生成 conversation-level report。

## 2 当前产品不变量

- Practice 是一个连续聊天窗口，不存在题目侧栏、题号计数、当前题目、逐题分类或题目预算。
- `startPracticeSession` 创建开场 assistant message；`sendPracticeMessage` 追加普通 user/assistant message pair。
- `practice_messages` 是聊天内容真理源；`practice_session_events` 只保留 session started/completed 生命周期事实。
- Report 以整场对话生成 dimensions、evidence、risks、next actions，不生成逐题 assessment。
- Voice 暂时 fail-closed：前端原生 disabled，后端返回 unsupported capability，且不触达 provider 或持久化音频。

## 3 场景分层

| ID | 类型 | 目的 | 环境 |
|----|------|------|------|
| `E2E.P0.098` | automated | 组合当前 handler/service/store/report persistence/retry 契约 | repo tests；不维护场景专属后端 |
| `E2E.P0.099` | hybrid | 验证真实 Mailpit 登录、真实前后端、PostgreSQL、真实 provider 与桌面/移动 UI | shared host-run environment |
| `E2E.P0.100` | hybrid | 长流程真实 provider UAT runbook 与双语材料 | shared host-run environment |

## 4 设计决策记录

| ID | 决策 | 理由 |
|----|------|------|
| D-1 | 删除 P0.099 专属 Playwright test server/config | 对应 Go server 已不存在；继续保留会形成第二套漂移运行时 |
| D-2 | P0.098 使用当前 focused gates 组合 | 复用 owner 测试，不复制业务 orchestration |
| D-3 | P0.099 使用 agent-browser/human hybrid 真实验收 | 能验证真实登录、provider、DB 与可见 UI，并留存截图 |
| D-4 | 共享环境由顶层 `test/scenarios/env-*.sh` 管理 | 场景目录不私有化环境 bootstrap |

## 5 Operation Matrix

| operationId | fixture | consumer | persistence | AI | gate |
|-------------|---------|----------|-------------|----|------|
| `registerResume` | current fixture | Resume Workshop | resumes/jobs | `resume.parse.default` | P0.099/P0.100 |
| `importTargetJob` | current fixture | Home/Parse | target jobs/jobs | `target.import.default` | P0.099/P0.100 |
| `createPracticePlan` | current fixture | Parse/Workspace/Report CTA | `practice_plans` | none | P0.022/P0.098/P0.099 |
| `startPracticeSession` | current fixture | Practice entry | sessions/messages/lifecycle/outbox | `practice.chat.default` | P0.023/P0.098/P0.099 |
| `sendPracticeMessage` | current fixture | Practice chat | `practice_messages`, task-runs | `practice.chat.default` | P0.044/P0.046/P0.098/P0.099 |
| `completePracticeSession` | current fixture | Finish report | session/report/job/lifecycle/outbox | report job | P0.047/P0.098/P0.099 |
| `getFeedbackReport` | current fixture | Generating/Report | `feedback_reports` | none on read | P0.056/P0.058/P0.099 |

## 6 验收标准

- P0.098 在当前代码上运行真实测试项且拒绝 no-test/skip-as-pass。
- P0.099 只在共享环境健康、当前 run evidence 完整、四张截图存在且 focused gates 通过时 PASS。
- P0.099 截图明确显示连续聊天、disabled voice、conversation report 的桌面/移动状态。
- active runtime/scenario 资产不存在 append-event、question counter/sidebar/current-question 合同。
- 真实凭证、邮箱验证码、cookie、完整 prompt/response 不进入 tracked docs 或验收 evidence。

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 2.0 | Rebase E2E owners onto continuous conversation and shared real-browser acceptance. |
