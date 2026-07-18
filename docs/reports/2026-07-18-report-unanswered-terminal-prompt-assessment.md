# Report Unanswered Terminal Prompt 交付复盘报告

> **日期**: 2026-07-18
> **审查人**: Codex

**关联计划**: [Report Generation Baseline](../spec/backend-review/plans/001-report-generation-baseline/plan.md)
**关联 Bug**: [BUG-0186](../bugs/BUG-0186.md)

## 1 复盘范围与成功证据

- 本次修复把报告使用的 provider assessment transcript 与完整审计会话分开：只排除一条末尾未回答的 assistant 消息，完整数据库 transcript 与 report conversation read 保持不变。
- TDD RED 先证明旧 `reportCompletePayload` 仍发送末尾 assistant；GREEN 后 focused test、完整 `backend/internal/review` 包和根 `make test` 均通过。根回归结果为 Python 584 tests / 4583 subtests、Go 全包、frontend 126 files / 1027 tests。
- backend 重新部署到本地真实场景环境后，新建真实 practice session 和 report；真实 provider 生成 ready 报告。Chrome desktop 与 390×844 均验证报告只评价实际 user answer，会话页仍显示末尾追问，且无横向溢出、无 console error/warn。
- PostgreSQL 侧以同一 report/session 复核：冻结坐标和 `practice_messages` 都是 3 条、末条 assistant；report summary/highlights/issues/actions/dimensions 不含末尾追问的规模、分片或锁竞争主题。
- Chrome 自动 `setFiles` 权限不足后，使用同一 Chrome 的 macOS 原生 chooser 完成真实 Markdown、TXT、PDF 上传；三份 resume 均从 processing 到 ready，数据库 MIME、byte size 与 `upload_status=uploaded` 一致，PDF 详情实际渲染页面图像。
- spec、plan、test/BDD checklist 与 context discovery 已在原 owner `backend-review/001-report-generation-baseline` 原地修订；没有创建 sibling plan，也没有改变 API、OpenAPI、migration 或前端事实源。

## 2 会话中的主要阻点/痛点

- 初始 owner 定位曾偏向 prompt/schema 的 `F3/002`。
  - **证据**：反查 production `reportCompletePayload`、当前 spec 和 plan operation boundary 后，确认运行时 transcript builder 的 owner 是 `backend-review/001-report-generation-baseline`。
  - **影响**：增加了一轮 owner 重校对；若未反查代码，可能把修复写进错误的 plan。
- 原浏览器回归把 Workspace 点击删除后立即移出列表误判为缺少确认。
  - **证据**：`docs/ui-design/module-job-workspace.md` 明确要求删除图标直接调用 generated `archiveTargetJob` 持久软归档；只有账号删除要求二次确认。
  - **影响**：产生一个假阳性 finding，后续必须用 owner UI 文档纠正证据矩阵。
- Chrome 自动文件注入受扩展本地文件权限阻塞。
  - **证据**：file chooser 成功打开，但 `setFiles` 返回 `Not allowed`；改用同一 Chrome 的 macOS 原生 chooser 后，真实 Markdown/TXT/PDF 均上传、解析和展示成功。
  - **影响**：增加了一段原生 UI 操作与证据采集成本，但最终没有留下应用功能缺口。
- ready 状态本身无法揭示语义污染。
  - **证据**：缺陷报告通过 provider/schema/validator 并渲染为 ready；只有把报告自然语言、完整 conversation 和数据库消息角色并排比较，才发现未回答主题被当作能力缺口。
  - **影响**：单靠状态码、schema 和 evidence anchor 的历史 gate 会得到假绿。

## 3 根因归类

- 报告把完整可审计 transcript 同时当作可评价 transcript，且 gate 只约束 evidence role，没有约束 provider 可见的未回答主题。
  - **类别**：spec-plan。已在本次原 owner spec/plan/checklist 中补充 assessment projection 与真实 Chrome/DB 语义 gate。
- Owner 发现依赖名称近似，初始没有先从 production consumer 反向确认职责。
  - **类别**：spec-plan。已在 `context.yaml` 增加 trailing unanswered / assessment transcript discovery keywords；现有 change-intake 与 deep reconcile 规则足够，不需要修改 skill。
- Workspace 删除假阳性来自执行时没有在写 finding 前完成 current UI owner 对照。
  - **类别**：无需仓库改动。`AGENTS.md` 已明确要求用户可见 UI 读取对应 `docs/ui-design/`，本次应作为执行纪律而非新增规则处理。
- Chrome upload 的拒绝来自用户浏览器扩展权限，不是仓库 runtime、前端或后端合同。
  - **类别**：无需仓库改动。权限只能由用户在 Chrome 扩展详情显式开启。

## 4 对流程资产的改进建议

- 后续报告语义变更继续在 `backend-review/001-report-generation-baseline` 保留“三面对照”gate：provider assessment input、持久化完整 conversation、最终用户可见自然语言。
  - **落点**：spec-plan
  - **优先级**：high
- 报告类浏览器验收不能以 `ready`、schema valid 或合法 user anchor 单独收口；至少加入一个未回答 terminal assistant 的负向主题断言，并绑定 desktop/mobile conversation 截图。
  - **落点**：spec-plan
  - **优先级**：high
- 全页面回归在形成 finding 前先逐项引用当前 UI owner 的预期行为；账号删除与业务资产软归档不得共用同一确认假设。
  - **落点**：无需仓库改动
  - **优先级**：medium
- Chrome 文件上传应在长链路开始前做权限 preflight；未授权时立即报告精确设置路径，并在用户已授权上传测试时使用原生 chooser 作为真实交互 fallback，而不是把应用功能永久标为 blocked。
  - **落点**：skill（上游 Chrome file-upload troubleshooting / preflight）
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：保留 Phase 15 的 focused transcript projection test 与真实 Chrome/DB 语义对照，防止未来 prompt、repair 或 report validator 调整重新把未回答主题带回评分。
- 下一步推荐：后续 report/provider 修改继续复跑 terminal unanswered 负向主题和三面对照；三格式上传已完成，无需再等待扩展权限。
- 可以延后：把 Chrome upload permission preflight 反馈给上游插件 skill；它不阻塞本次代码修复，也不应诱发仓库内兼容层或测试绕过。
