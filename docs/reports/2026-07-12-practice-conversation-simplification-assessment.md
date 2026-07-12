# Practice Conversation Simplification 交付复盘报告

> **日期**: 2026-07-12
> **审查人**: Codex

**关联计划**: [E2E Current Conversation Funnel Journey](../spec/e2e-scenarios-p0/plans/001-full-funnel-happy-journey/plan.md)
**关联 Bug**: [BUG-0159](../bugs/BUG-0159.md)

## 1 复盘范围与成功证据

- 本次交付把模拟面试从“题目/轮次状态机”收敛为不限定问题数量的连续文字对话：删除题目边栏、题号/题数、逐题 API/表/报告结构；电话模式在前端原生 disabled、后端 fail-closed。
- 报告改为整场 conversation 的 readiness、dimension assessments、evidence、issues/actions，不再保存逐题评分或 turn retry ID。
- 全量后端 `go test ./backend/... -count=1`、`go vet ./backend/...` 通过；前端 111 个 Vitest 文件、708 个测试、typecheck、build 通过；UI contract 45/45、报告 pixel parity 14/14 通过。
- P0.007、P0.022-P0.026、P0.044-P0.047、P0.051、P0.056-P0.059、P0.070、P0.072、P0.098、P0.099 通过；其中 P0.022 直接覆盖真实 Handler/Repository，P0.099 使用共享真实环境和真实 provider 完成浏览器验收。
- 真实链路从 Mailpit 登录、简历解析、JD 导入、plan/session 创建、连续消息、完成会话到 ready report 闭环；桌面/移动端共四张脱敏截图已验收，数据库与 AI observability 证明使用真实持久化和当前 profile。
- migration up/down/up、OpenAPI/codegen、Prompt/Rubric/hardcode lint、offline eval 24/24、active negative-reference 和场景环境 4/4 均通过。

## 2 会话中的主要阻点/痛点

- 主实现通过 focused tests 后，真实 PostgreSQL 仍连续暴露空数组写成 `NULL`、完成事件写旧列、report `text[]` 编码错误与 generating 重试不可重入。
  - **证据**：真实漏斗分别在 plan create、session complete 和 report retry 阶段失败；BUG-0159 的 sqlmock/live DB 复现与新增回归测试确认根因。
  - **影响**：没有真实环境验收时会产生“代码和单测已绿、核心漏斗仍不可用”的假完成。
- Resume parse runtime 没有使用已有 observability decorator，provider 首次失败时缺少 `ai_task_runs` 与 raw debug 证据。
  - **证据**：共享环境首次解析失败后数据库没有对应 task run；补齐 runtime wrapper 后重试产生当前 profile 的可追踪记录。
  - **影响**：AI 输出问题和应用持久化问题无法快速分层，增加诊断往返。
- P0.098/P0.099 的旧专用 harness 与当前 continuous conversation 契约脱节，无法承担真实验收 owner。
  - **证据**：旧 Playwright server/spec 依赖已删除的事件/逐题模型；改为当前跨层组合门禁和共享真实环境浏览器验收后四阶段通过。
  - **影响**：历史 scenario 名称仍在，但运行证据无法证明当前产品路径。
- 旧结构残留不只存在于类型和 route，还藏在 OpenAPI 参数描述、privacy matrix 与 UI 原型文案/网格中。
  - **证据**：最终 active negative search 找到 `/events` 描述、`practice_turns`/`question_assessments` privacy 行及“逐轮读”文案；修订并增加负向断言后归零。
  - **影响**：后续生成物、隐私删除和正式前端 parity 可能重新引入旧模型。

## 3 根因归类

- PostgreSQL array/exact-column 漂移
  - **类别**：spec/plan
  - 当前计划要求 migration 与 focused store tests，但在删除跨层数据模型时没有把 nil/empty array、exact SQL columns 和 live PostgreSQL recovery 固化为同一 gate。
- AI runtime observability 漏接
  - **类别**：无须仓库改动
  - 已用 runtime wiring 测试修复；属于具体 bootstrap 漏接，当前没有证据表明通用 skill 或治理规则缺失。
- 旧 scenario harness 失真
  - **类别**：README / spec-plan
  - 场景说明曾把专用 mock harness 当成 owner，而产品契约已迁移到共享 host-run 环境；本次已原地重写 P0.098/P0.099。
- 描述、隐私矩阵和原型中的残留
  - **类别**：spec/plan
  - 原 negative gate 侧重 schema/path/组件正向符号，没有列出 prose、privacy projection 与 prototype copy 这些二级消费者。
- 两次命令使用错误场景目录/glob，以及第一次迁移命令未注入 `DATABASE_URL`
  - **类别**：无须仓库改动
  - 都被实际目录/标准场景环境快速纠正，未造成实现误判，不足以认定为流程缺陷。

## 4 对流程资产的改进建议

- 在涉及表/字段删除或数组类型变更的 backend plan 中，明确加入“current migration + exact store SQL + nil/empty semantics + live PostgreSQL retry”四联 gate。
  - **落点**：backend-practice / backend-review 的后续 spec-plan 模板或 checklist
  - **优先级**：high
- 场景 owner 选择应优先复用 `test/scenarios` 的共享 host-run 环境；专用 mock server 只能作为 focused gate，不能单独证明真实 full-funnel。
  - **落点**：test/scenarios/README.md 与 e2e-scenarios-p0 spec-plan
  - **优先级**：high
- 数据模型删除的 zero-reference 清单应显式覆盖 OpenAPI prose/baseline、privacy matrix、UI prototype copy/layout、generated artifacts 和 scenario evidence，而不只搜索 API path/schema/type。
  - **落点**：change-intake 或 plan-review 的删除类审查清单
  - **优先级**：medium
- 保持语音模式 disabled/fail-closed，直到 continuous text conversation 的真实用户流程和报告质量稳定；重新启用时使用独立 owner plan，不在当前主链路内提前恢复正向契约。
  - **落点**：practice-voice-mvp spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

1. 下一轮最高价值动作是对 `backend-practice` / `backend-review` 的未来 schema 变更强制执行 live PostgreSQL 四联 gate，优先防止“单测绿、真实漏斗断”的高成本回归。
2. 保留 P0.099 作为共享真实环境的会话级浏览器验收 owner，每次调整 practice/report contract 后至少复跑桌面与 390px 移动端截图。
3. 在连续文字流程经过一轮真实用户验证前，不启动语音接入；届时另开 `practice-voice-mvp` owner revision，先定义 audio provider、设备权限和同会话恢复边界。
4. 删除类 zero-reference 扩展可在下一次 change-intake/plan-review 流程改进中集中处理，本次实现已通过新增负向测试覆盖当前残留点。
