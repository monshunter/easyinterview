# E2E Scenarios P0 Spec

> **版本**: 2.17
> **状态**: active
> **更新日期**: 2026-07-15

## 1 目标

只维护必须驱动真实运行环境才能证明的 P0 端到端业务流程。E2E 通过真实 HTTP API 或真实 frontend UI 访问当前 backend、PostgreSQL 与 Mailpit；Go test、Vitest、lint、build、fixture parity、provider CLI/eval 均属于代码或评估 gate，不在本 subject 包装为场景。

## 2 当前产品不变量

- Practice 是一个连续聊天窗口，不存在题目侧栏、题号计数、当前题目、逐题分类或题目预算。
- P0.098 只从已存在的 waiting session 执行真实 completion；不启动 session、不发送 chat message、不创建下一轮 plan。
- `practice_session_events` 只保留 session started/completed 生命周期事实；完成后的 canonical round progress 必须从真实 API/UI 读取。
- Report 以完成时 frozen context 和整场对话生成 conversation-level report；`reportId` 是 Report/Generating/ReportConversation 页面唯一 locator。会话记录只作为所属报告的只读附件，不存在公共 session list。
- E2E 浏览器不得 intercept/fulfill 应用请求，不得用 fixture transport、dev mock、jsdom 或进程内 handler 代替 backend。
- Evidence 隐私门禁只保护项目用户数据与认证/运行 secret；PNG 色彩配置、创建工具、文件格式等不含用户内容的开发过程技术元数据不属于隐私数据，不得仅因存在而拒绝当前证据。

## 3 当前场景

| ID | Owner plan | 类型 | 目的 | 环境 |
|----|------------|------|------|------|
| `E2E.P0.098` | `001-real-api-ui-journeys` | automated | 真实登录后调用 completion API，并在刷新后验证 Home、Workspace 与 TargetJob 的持久化 round progress 一致 | shared host-run frontend/backend/PostgreSQL/Mailpit |
| `E2E.P0.099` | `001-real-api-ui-journeys` | hybrid | 验证真实 report/generating/conversation UI、authenticated report API、PostgreSQL 绑定与精确六图 + 有界非图片证据 | shared host-run frontend/backend/provider/PostgreSQL |
| `E2E.P0.101` | `backend-auth/001 + frontend-shell/001` | automated | 验证真实 Mailpit email-code、首次资料补全、设置齿轮/真实账号字段、logout 与完成账号再次登录 | shared host-run frontend/backend/PostgreSQL/Mailpit |

认证首次资料补全仍由独立 auth owner 承接；本 subject 只登记同一真实 E2E 资产，不复制 auth 业务实现或代码测试。

## 4 设计决策

| ID | 决策 | 理由 |
|----|------|------|
| D-1 | P0.098 只验 completion/progress | 场景不拦截 session route，也不创建 round-2 plan；它只证明真实 completion 后 canonical round 投影跨页面一致并在 reload 后保留 |
| D-2 | P0.099 只验 report/generating/report-owned conversation | Practice live UI 与代码级边界由各自 owner 测试；既有截图合同保持精确六图，conversation 只补真实 Report → Conversation → Back 与 API/DB 绑定的有界非图片证据，不增加第七张图 |
| D-3 | provider 内容可靠性不属于应用 E2E | provider CLI、evalkit、context-aware judge 与重复采样没有通过应用 API/UI 驱动用户流程，应作为 code/eval gate 独立报告 |
| D-4 | 代码回归与 E2E 分层 | focused tests 用于开发反馈；阶段完成与 CI 由根 `make test` 统一运行 backend/frontend 全量单测，E2E 脚本不得再次编排这些命令 |
| D-5 | 共享环境由顶层 `test/scenarios/env-*.sh` 管理 | 场景目录不私有化环境 bootstrap，setup/cleanup 只隔离自身数据 |
| D-6 | P0.101 保持独立 auth/settings owner | E2E suite 原地扩展同一真实环境资产与运行状态；backend-auth/frontend-shell 拥有 email-code/profile/Settings 行为，不创建第二个设置场景，也不把破坏性的账号删除加入共享登录生命周期 |
| D-7 | Privacy gate 只识别用户数据与 secret | transcript、JD、简历、Cookie、验证码、密钥和可还原正文必须 fail closed；无用户内容的 PNG `iCCP` 等技术元数据属于开发过程，不作为隐私失败条件 |

## 5 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `completePracticeSession` | `PracticeSessions/completePracticeSession.json` | Practice completion action | backend-practice completion handler | session/report job/lifecycle/outbox | indirect report generation；not asserted by P0.098 | P0.098 real API completion |
| `getTargetJob` | `TargetJobs/getTargetJob.json` | Home / Workspace / target detail | backend-targetjob read handler | TargetJob canonical round progress projection | none | P0.098 real reload projection |
| `getFeedbackReport` | `Reports/getFeedbackReport.json` | Generating / Report | backend-review owned-report handler | `feedback_reports` + frozen context | real provider generation precedes ready evidence | P0.099 real API/UI |
| `getReportConversation` | `Reports/getReportConversation.json` | ReportConversation | backend-review conversation handler | owned `feedback_reports.session_id` + ordered `practice_messages` | none | P0.099 real UI/API/DB binding；no list operation |
| `startAuthEmailChallenge` | `Auth/startAuthEmailChallenge.json` | AuthLoginScreen | backend-auth challenge handler | auth challenge + email dispatch job | none | P0.101 real Mailpit delivery |
| `verifyAuthEmailChallenge` | `Auth/verifyAuthEmailChallenge.json` | AuthVerifyScreen | backend-auth verification handler | challenge consumption + first-party session | none | P0.101 real code verification |
| `completeMyProfile` | `Auth/completeMyProfile.json` | AuthProfileSetupScreen | backend-auth profile-completion handler | users display/profile/terms | none | P0.101 first-login setup |
| `getMe` | `Auth/getMe.json` | AppRuntimeProvider + Settings | backend-auth current-user handler | session-owned user identity/profile | none | P0.101 real displayName/full email projection；no duplicate page request；evidence redacts email |
| `logout` | `Auth/logout.json` | AuthLogoutScreen entered from Settings | backend-auth logout handler | session revocation | none | P0.101 real logout/relogin |

Resume/JD/plan/chat/provider 的业务与可靠性由各代码 owner 测试或 eval gate 承接；它们不是本 subject 为扩大覆盖率而拼接的 E2E 步骤。

## 6 验收标准

- P0.098 浏览器登录真实 frontend，经真实 backend 完成已等待的 session；应用请求无 route interception，刷新后 Home、Workspace 与 TargetJob API 一致显示 round 1 `done`、round 2 `current`，且场景不创建 round-2 plan。
- P0.099 对当前 run 的 en/zh ready report 与 generating state 捕获精确六张 `fullPage: true` 图片；每个 row 绑定当前 DB/API 状态、canonical content digest、action/content audit、screenshot digest 与 report/session/context digest。
- P0.099 的两张 390x844 ready report 图片完整覆盖 action 区域，实际 English / zh-CN label 分别满足 `<=24 whitespace words` / `<=64 Unicode code points` 且无 clipping、ellipsis、hidden content 或横向溢出；这不替代代码层 exact boundary test。
- P0.099 还从 ready Report 点击“查看本次面试记录”，验证 URL 只含 `reportId`、真实 API 返回同一 report/context 与 DB 绑定 session 的严格有序消息，并按 Back 回到该 Report；conversation 不新增截图、不抄录正文，只记录 route/status/count/sequence digest、report/session/context digest 与返回目标。公共 `listPracticeSessions` route/API 请求必须为零。
- P0.101 经真实 Mailpit 收取验证码并完成首次 profile setup；点击唯一设置齿轮后显示同一账号的姓名/完整邮箱且不出现旧账号下拉/设置 tab，随后通过 Settings 进入 logout；完整邮箱只在页面断言，不写入 trigger/result 证据；完成账号再次登录不重复补全，且浏览器请求不被 fixture/mock transport 接管。账号删除不在该共享场景执行。
- 三个场景都只接受当前真实环境证据；cookie、邮箱验证码、完整 prompt/response、JD、简历和 transcript 不写入 tracked docs 或验收 evidence。
- 技术元数据可以保留，前提是其中不携带上述用户数据、认证材料或 secret；校验器必须按内容风险判断，不能把普通开发过程信息等同于隐私数据。
- 根 `make test` 与 code/eval gate 单独报告，不得作为 E2E PASS marker。

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-15 | 2.16 | Extend P0.101 in place with the single settings gear, real `/me` account fields and Settings-owned logout；keep delete-account destructive coverage in domain/contract tests. |
| 2026-07-15 | 2.15 | Narrow evidence privacy gates to project user data and secrets; allow benign development metadata such as PNG color profiles. |
| 2026-07-15 | 2.14 | Extend P0.099 with real report-owned Conversation navigation/API/DB binding and deleted-session-list negative evidence; retain the exact-six screenshot contract by using bounded non-image evidence. |
| 2026-07-14 | 2.13 | 将 P0.101 作为独立 auth owner 的真实 E2E 资产纳入 suite 清单，保持业务合同不复制；统一三项场景的 current-run evidence 边界。 |
| 2026-07-14 | 2.12 | 将 E2E 收敛为真实 API/UI：P0.098 只承接 completion/progress，P0.099 只承接 report/generating；删除 provider CLI/eval 场景 owner。 |
