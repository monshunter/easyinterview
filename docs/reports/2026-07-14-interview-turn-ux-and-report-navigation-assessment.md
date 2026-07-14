# Interview Turn UX and Report Navigation 交付复盘报告

> **日期**: 2026-07-14
> **审查人**: Codex

**关联计划**: [Backend Practice Event Loop](../spec/backend-practice/plans/002-event-loop-and-completion/plan.md)、[Frontend Practice Text Event Loop](../spec/frontend-workspace-and-practice/plans/002-practice-text-event-loop/plan.md)、[Frontend Home JD Import](../spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/plan.md)、[Frontend Report Screen](../spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/plan.md)、[Frontend URL Routing](../spec/frontend-shell/plans/004-url-addressable-routing/plan.md)、[Backend TargetJob Import](../spec/backend-targetjob/plans/001-targetjob-import-and-parse-bootstrap/plan.md)

**关联 Bug**: [BUG-0167](../bugs/BUG-0167.md)、[BUG-0168](../bugs/BUG-0168.md)、[BUG-0169](../bugs/BUG-0169.md)

## 1 复盘范围与成功证据

- 本次交付完成三个原始目标：面试页面不再显示 model/rubric 等内部调试信息；用户回答提交后立即进入消息流，请求期间锁定输入并显示面试官思考状态，只有 retryable failure 才在原消息下显示同 identity 重试入口；Home/TargetJob 导入收敛为粘贴 JD，不保留 URL、文件或同步表单分支。
- 按用户最终纠正，面试规划详情只在内容区右上角提供“面试报告”页面级入口；独立 `/reports?targetJobId=...` 只展示当前规划各轮 current report 与 latest attempt。Parse 不再嵌入报告列表、请求或 `section=reports` 兼容；Report/Generating 有可信规划上下文时返回列表，缺失时安全 replace 到工作台。
- 6 个既有 owner 原地修订并恢复 `completed`，旧 backend-practice 10.6 与 backend-targetjob 18.6 保持 `HISTORICAL-SUPERSEDED`，本轮完成证据分别由 11.8 与 19.5 承接，没有创建同主题 sibling plan。
- 当前全量验证通过：backend `go test ./...`；frontend 121/121 files、977/977 tests、typecheck 与 production build；UI source contract 62/62；OpenAPI 37 operations、37 fixtures、migration lint、隔离 index codegen drift 与 pruning `real_residuals=0`。
- BDD/路由验证通过：P0.006 160 Playwright tests；P0.016、P0.058、P0.059、P0.088、P0.089、P0.090；practice P0.044/P0.046；paste-only P0.010/P0.012。所有 final wrapper 使用当前代码与当前环境证据，不复用历史 PASS。
- 本地数据按用户授权清空并重建，PostgreSQL/Redis/MinIO/Mailpit 4/4 健康，Mailpit Web 端口为 8026；migration 与真实 PostgreSQL focused gates 均通过。
- 截图验收闭环：规划详情入口与当前规划报告列表各有 1440x900/390x844 正式前端截图和 manifest；报告 Context Strip 有 1440x1200/390x844 full-page 截图、SHA-256、状态、viewport、UUID 不可见审计；practice pending/failed-retry 和 Home paste-only 也保留 desktop/mobile 当前证据。
- `/work-journal` 全量门禁再次通过：`make test` 覆盖 UI contract 62/62、Python 590 tests / 5181 subtests、全部 Go packages、frontend 121 files / 977 tests；修订后的 P0.044/P0.046/P0.059 也重新完成当前源码场景验证与清理。

## 2 会话中的主要阻点/痛点

- 报告列表最初被放入规划详情，用户进一步明确它应是内容区右上角一级入口后的独立当前规划页面。
  - **证据**：最终同时修订 frontend-home Phase 19、frontend-report Phase 10 与 frontend-shell Phase 11；Parse 的 embedded state/effect/section compatibility 全部删除，ReportsScreen 成为唯一 list consumer。
  - **影响**：如果只改入口视觉、不重做 route owner、返回路径与消费边界，会继续把报告列表耦合进 Parse，并留下两个事实源。
- 旧本地数据库首先暴露上一轮 DDL 形态，清库重建后才稳定暴露 P0.058 raw SQL fixture 缺少当前 reply generation 的真实问题。
  - **证据**：干净 baseline 上 `practice_messages_client_id_check` 失败；补齐 `reply_generation=1` 后 focused 四个 subcase 与 P0.058 四阶段 wrapper 通过，见 BUG-0167。
  - **影响**：只在长期复用的本地卷运行测试可能同时产生假失败和假通过，增加 schema/fixture 根因分层成本。
- P0.006 verifier 把报告列表业务 marker `failed=1` 当成 Playwright 失败摘要。
  - **证据**：runner 实际为 160 passed；将失败 pattern 锚定到行首 summary 并加入 source contract 后完整 wrapper 通过，见 BUG-0168。
  - **影响**：混合日志的宽泛 grep 会让新增业务状态覆盖反而破坏交付 gate，掩盖真实 runner 结果。
- 新 Reports parity spec 让当前 inventory 从 12 增至 13，同时 paste-only 合同删除导致 generated event/job artifacts 与 source truth 暂时漂移。
  - **证据**：scaffold 先因固定数量失败；常规 codegen-check 在大型 dirty worktree 中比较旧 index，改用临时 index 纳入当前 intended changes 后证明生成结果 byte-clean。
  - **影响**：新增真理源消费者时，若只更新实现而未反查 inventory/generated baseline，最终全量 gate 才会暴露机械漂移。
- 提交前根级聚合又暴露 4 个 contract failures：旧事件/表总数、P0.044/P0.046 重复 Vitest parser、P0.059 旧 marker 与 parity inventory。
  - **证据**：改为 exact-name sets、共享 verifier 单一所有权和当前 P0.059 marker 后，focused Python、根级 `make test`、P0.044/P0.046/P0.059 全部通过，见 BUG-0169。
  - **影响**：只依赖 focused PASS 会遗漏跨 owner 的防漂移测试；复制共享 parser 或维护裸数量会让正确删改被误判为回归。

## 3 根因归类

- 报告入口与列表 owner 边界
  - **类别**：spec/plan
  - “一级入口”必须同时约束页面归属、route authority、data consumer、返回路径和旧兼容删除；单纯 UI 定位不足以防止耦合。
- P0.058 raw SQL fixture 漂移
  - **类别**：spec/plan
  - durable reply schema 的 final gate 已覆盖 production store，但跨 owner report scenario 的直接 SQL fixture 未被 schema change matrix 反向列出。
- P0.006 混合日志误判
  - **类别**：README
  - 场景 verifier 已要求精确 runner marker，但缺少“失败检测也只能匹配 runner-owned summary/status”的明确口径。
- parity inventory 与 generated artifact 同步
  - **类别**：无需仓库改动
  - 当前已用 exact spec-set contract、codegen 与隔离 index 验证闭环；本次没有证据证明需要改变全局 codegen 的 index-based 设计。
- closeout contract 与当前 source ownership 漂移
  - **类别**：spec/plan
  - 三个原 owner checklist 已原地重新打开最终 gate，并用 exact set、共享 verifier 委托和当前场景证据重新闭环；无需建立 sibling fix plan。
- validator 入口从旧路径迁移到 implement-owned shared script
  - **类别**：无需仓库改动
  - 当前 skill 与仓库入口可发现且 6 个 context 全部通过；旧路径只存在于历史报告，不是 active owner 文档。

## 4 对流程资产的改进建议

- 对“入口从嵌入区迁移为独立页面”的修订，在 owner checklist 中固定五项成套 gate：entry location、route identity、single data consumer、trusted return/fallback、legacy compatibility zero-reference。
  - **落点**：change-intake / plan-review 的 UI route revision 检查清单
  - **优先级**：high
- 对新增必填列、联动 CHECK 或状态机字段的 migration，增加 active raw SQL fixture inventory，至少覆盖相邻 owner 的 scenario/evidence tests，并要求一次 clean baseline rebuild。
  - **落点**：backend plan checklist 与 test/scenarios README
  - **优先级**：high
- 场景 verifier 的失败检测应只接受 runner-owned summary/status 行，并用与业务日志冲突的 marker（如 `failed=1`）作为 source contract 负例。
  - **落点**：test/scenarios README；后续经用户确认可补充 PATTERNS 模式 4
  - **优先级**：high
- 新增 parity page/spec 时继续使用“目录 exact set = verifier marker exact set”，不要维护独立的手工数量真理源。
  - **落点**：frontend pixel parity scaffold
  - **优先级**：medium
- 对 events/schema/scenario inventory 的 contract tests 统一使用 exact-name sets；共享 verifier 的 consumer 只验证委托和禁止重复 parser。
  - **落点**：owner plan final gate 与 `scripts/lint/*contract_test.py`
  - **优先级**：high

## 5 建议优先级与后续动作

1. 下一轮最高价值动作是把 raw SQL fixture inventory + clean baseline rebuild 固化到涉及 migration/state-machine 的 owner checklist，避免长期本地卷掩盖当前约束漂移。
2. 在用户确认后，将 BUG-0168 的 runner-owned summary 规则补入 PATTERNS 模式 4；当前 source contract 已先形成可执行防线。
3. 当前功能范围不再扩展全局报告中心、完整历史版本或新的 OpenAPI/数据库结构；本轮以 `/work-journal` 冻结 `fix(interview): close turn UX and report navigation (BUG-0167, BUG-0168, BUG-0169)` 并提交，随后针对该提交执行一次相较 `main` 的分支 L2 review。
