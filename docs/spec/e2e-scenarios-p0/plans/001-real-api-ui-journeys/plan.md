# 001 Real API/UI Journeys

> **版本**: 3.11
> **状态**: active
> **更新日期**: 2026-07-15

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

保留三个必须依赖真实运行环境的最小 P0 流程：P0.098 验证 completion 后持久化 round progress，P0.099 验证 report/generating 的真实 UI/API/DB 证据，P0.101 登记由 backend-auth/frontend-shell 持有业务合同的 Mailpit email-code/profile-setup/Settings/logout 流程。代码层测试和 provider 可靠性评估由各自 owner 承接，不在场景 shell 中聚合。

## 2 范围

- P0.098：Mailpit 登录、真实 completion API、刷新 Home/Workspace/TargetJob 后的 round progress 一致性。
- P0.098 不拦截或 fulfill session route，不创建 round-2 plan，也不把预置数据或数据库查询本身作为 E2E PASS。
- P0.099：真实 frontend/backend/provider 下的 en/zh ready report、generating 与 report-owned conversation 页面，authenticated report APIs、只读 PostgreSQL 投影和精确六图 + 有界非图片证据。
- P0.101：真实 frontend/backend/Mailpit 下的 email-code、首次 profile setup、唯一设置齿轮、真实账号字段、Settings-owned logout/relogin；suite 只维护运行资产与 current-run 状态，不复制 auth/settings 业务合同，不执行 deleteMe。
- 删除没有真实应用 API/UI 链路的 provider CLI/eval 场景 plan；相关可靠性验证保留为 code/eval gate。

## 3 质量门禁分类

- **Plan 类型**: real API/UI integration。
- **TDD 策略**: 实现变更由代码 owner 使用 focused tests 快速反馈；阶段完成与 CI 从仓库根执行 `make test`，整体回归 backend/frontend 单测。
- **BDD 策略**: 仅 P0.098、P0.099 与 P0.101；证据来自真实请求、页面、持久化结果和用户可见状态。P0.101 的业务行为仍由 backend-auth/frontend-shell 持有。
- **独立 code/eval gate**: OpenAPI/codegen、migration、prompt/eval、lint、build、fixture parity 与 provider reliability 均独立运行和报告，不进入 E2E 脚本或 PASS marker。

## 4 Coverage Matrix

| 行为 | 类型 | Gate | 负向断言 |
|------|------|------|----------|
| completion → round progress | primary | P0.098 | route interception、mock transport、创建 round-2 plan、只看 DB 不看 UI/API |
| progress after reload | persistence | P0.098 | Home/Workspace/TargetJob 投影不一致或仅内存更新 |
| generating → report UI → conversation → back | primary | P0.099 | fixture-only 页面、sessionId route、公共 session list、伪 ready、跨 run evidence |
| ready report desktop/mobile | responsive/privacy | P0.099 | 少于/多于六图、裁剪、用户数据/secret 泄露、digest 未绑定；普通 PNG 技术元数据不是隐私失败 |
| email-code → profile setup → Settings → logout/relogin | auth/session persistence | P0.101 | fixture transport、mock backend、跳过 Mailpit、完成账号重复补全、旧账号 dropdown/tab、静态账号值、破坏性 deleteMe |

## 5 实施步骤

### Phase 1: P0.098 completion/progress

- 使用 shared host-run frontend/backend/PostgreSQL/Mailpit。
- 为固定用户准备 round-1 plan 与 waiting session；浏览器登录后通过真实 API 完成 session。
- reload Home、Workspace list/detail，并与真实 TargetJob response 核对 `done,current,pending`。
- 明确禁止 application request interception；round-2 plan 不由场景创建。

### Phase 2: P0.099 report/generating

- 创建 current-run en/zh ready reports 与 generating resource。
- 捕获 desktop/mobile 精确六图，并通过 authenticated API + read-only PostgreSQL 绑定当前 run 状态与 digest。
- 人工直接查看六图，验证 action 区域、完整 label、无截断/省略/横溢；不使用 OCR 或抄录正文。
- 从 ready Report 的低强调入口进入只读 Conversation，验证 route 仅含 `reportId`、真实 `getReportConversation` 与只读 PostgreSQL 的 report/session/context/ordered-message binding 一致，再通过 Back 回到原 Report。
- 截图目录、manifest 与人工审计继续精确六张；conversation 只追加 route/status/count/sequence digest、binding digest 与 back target 等有界非图片字段，不保存 transcript 正文。浏览器与网络证据必须证明没有调用公共 `listPracticeSessions`。
- 隐私扫描只拒绝项目用户数据与认证/运行 secret；PNG `iCCP`、编码器信息等不携带用户内容的开发过程技术元数据允许存在，文件完整性仍由 chunk/CRC/digest gate 独立验证。

### Phase 3: 分层回归与收口

- 原地扩展 P0.101 Playwright flow：profile setup 后点击唯一设置齿轮，以当前登录时提交的 displayName 和 `/me` 完整账号邮箱核对真实 Settings；断言旧账号 dropdown/tab 不存在；从 Settings 进入既有 logout 确认，再完成同邮箱 relogin。场景运行时可在页面内断言完整值，但不得把完整邮箱、验证码或 cookie 写入日志或保存到证据；不得调用 deleteMe。
- 开发中运行必要 focused tests；完成时从根执行 `make test`。
- 只在相应 owner 发生变化时独立运行 OpenAPI/codegen、migration、lint、build、prompt/eval gate；不把结果写成 E2E marker。
- 静态校验场景目录、shell 语法、真实 Playwright 请求边界、docs/index/diff 与旧场景引用。真实环境运行由显式 `/scenario-run` 单独触发；未运行时场景保持 `Ready`，不记录 PASS。

## 6 验收标准

- P0.098 脚本通过真实 completion 后读取 Home、Workspace 与 TargetJob API，并要求 reload 后下一轮一致为 current；没有 route interception 或 round-2 plan creation。
- P0.099 runbook 要求 exact-six evidence 绑定当前 run 的 DB/API/report/session/context/screenshot digest，且真实 desktop/390 页面完整显示 report/generating 状态；同一 run 的 bounded evidence 另证明 Report → Conversation → Back、strict message ordering 与零公共 session-list 请求，不增加截图。
- P0.099 的隐私 gate 必须按项目用户数据与 secret 的可泄露性判定，不能因 PNG 色彩配置等开发过程元数据单独失败。
- P0.101 通过真实 Mailpit code 完成首次 profile setup，证明唯一设置齿轮/真实账号字段/Settings logout，并证明完成账号 relogin 后不重复补全；不执行账号删除。
- 场景脚本不调用 `go test`、Vitest/npm test、pytest、lint、build 或 provider CLI/eval。
- 根 `make test` 作为独立全量单测回归通过后才能阶段收口。
- 场景资产的 `Ready` 状态与 current-run 结果分开记录；只有显式 `/scenario-run` 的真实 API/UI 证据可以写入 PASS checklist。

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-15 | 3.11 | Correct P0.101 to assert the complete authenticated account email in Settings while keeping scenario logs/evidence redacted. |
| 2026-07-15 | 3.10 | Extend P0.101 in place with Settings gear, real account fields and Settings-owned logout while excluding destructive account deletion. |
| 2026-07-15 | 3.9 | Scope P0.099 privacy rejection to project user data and secrets while retaining independent PNG integrity and digest checks. |
| 2026-07-15 | 3.8 | Extend P0.099 with report-owned Conversation navigation and API/DB binding while preserving the exact-six screenshot contract through bounded non-image evidence. |
| 2026-07-14 | 3.7 | Register P0.101 as the independent auth-owned real E2E asset and keep all three current-run gates explicitly pending until scenario execution. |
| 2026-07-14 | 3.6 | Rename the owner and scenario slugs to their current real API/UI scope; keep static readiness separate from an explicit scenario run. |
| 2026-07-14 | 3.5 | 按真实 API/UI 边界收敛：P0.098 只验 completion/progress，P0.099 只验 report/generating；代码与 eval gate 分层。 |
