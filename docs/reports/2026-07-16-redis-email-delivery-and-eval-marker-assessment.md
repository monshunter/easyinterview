# Redis Email Delivery and Eval Marker 交付复盘报告

> **日期**: 2026-07-16
> **审查人**: Codex

## 1 复盘范围与成功证据

- 后端认证将 6 位邮箱验证码的投递 secret 从单进程内存改为既有 Redis：AES-GCM 加密值、SHA-256 namespaced key、5 分钟 TTL，producer/consumer 跨 backend 实例共享，成功投递后删除，失败保留给异步重试。
- Runtime 启动解析并 ping Redis，producer 与 SMTP writer 注入同一共享 store，shutdown 关闭 client；不新增服务、volume、network 或配置 key。
- 两个独立真实 Redis client 的 integration 覆盖 Put/Get/Delete、密文、TTL 与实际过期；full-container Mailpit Chrome challenge→收码→verify→profile 主流程通过，外部 SMTP job 一次成功，Redis namespace 在成功后为空，doctor 6/6。
- F3 004 owner 当前门禁重新执行：`make eval-offline` 完成 28 cases / 9 resolved prompts single-source drift-check，offline no-network grading，Promptfoo 28/28 PASS；prompt/rubric/profile/hardcode lint、judge/evalkit/registry focused tests全绿。
- 根 `make test` 完整通过：564 个 Python tests、4481 个 subtests、全部 Go packages、frontend 126 files / 1004 tests；`make build`、`make lint-config`、owner contexts、docs/index 和 diff gates 均通过。

## 2 会话中的主要阻点/痛点

- Redis 功能、真实 Mailpit/SMTP 和绝大多数回归已经通过后，根 `make test` 被 `REPORT_RUBRIC_V020_PASS` 缺失阻塞。
  - **证据**：`TestV020ActivationOwnerMarkersReady` 指向 F3 004 checklist；git history 显示 2026-07-15 completed-plan 压缩删除了原 Phase 8 verified markers，但下游 preflight 仍消费它们。
  - **影响**：认证交付无法按根级质量门禁收口，必须切换到相邻 eval owner 做证据恢复。
- 初次记录 RED 时，verified comment 中提到缺失 marker 名称，旧 preflight 立刻把它误判为已验证并继续失败在第二个 marker。
  - **证据**：旧实现只做 `verified` 与 marker 的同一行 substring 匹配；新增四个用例后证明 failure mention、普通 evidence 和 unchecked text 都可能形成假阳性。
  - **影响**：若第二个 marker 恰好仍在，单纯写一条失败说明就可能让 gate false-green。
- local-dev context 验证首次传入不存在的 `infra` target。
  - **证据**：validator 返回 available target 为 `repo`，使用 `--target repo` 后立即 PASS。
  - **影响**：仅一次命令重跑，无代码或文档返工。

## 3 根因归类

- **spec-plan**：completed-plan 压缩把机器消费的跨 owner marker当作历史流水账删除，缺少“仍有消费者的 marker 必须保留”的明确规则。
- **test**：marker verifier 使用宽泛 substring，而非显式结构字段，也没有针对失败说明提及 marker 的负向测试。
- **无需仓库改动**：context target 参数错误属于执行时选择失误，validator 已明确返回合法 target，未暴露工具缺口。

## 4 对流程资产的改进建议

- 已实施：F3 spec/plan/checklist 明确 marker 只能在当前 owner gate 重跑通过后写入，completed-plan 压缩不得删除仍被下游消费的 marker。
  - **落点**：`prompt-rubric-registry` spec/004 plan/checklist
  - **优先级**：high
- 已实施：preflight 只接受 verified comment 的显式 `marker=<name>` 字段；三类非通过文本纳入负向测试，storage owner marker同步规范化。
  - **落点**：registry preflight test + owner checklists
  - **优先级**：high
- 已实施：将该模式补入 Bug 模式库，要求压缩前反查 marker consumer，并拒绝从失败说明或普通 evidence 识别 PASS。
  - **落点**：`docs/bugs/PATTERNS.md` 模式 4、BUG-0179
  - **优先级**：high
- 建议后续：如果 marker 数量继续增长，考虑把跨 owner handoff 从自由文本 comment 提升为统一的结构化 evidence manifest，并由一个 lint 生成/校验 consumer exact set；当前显式属性已足够解决本次问题，不在本提交扩展架构。
  - **落点**：后续 governance/test owner plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一步推荐保持当前 full-container 环境运行，人工抽查 Mailpit 与外部 SMTP 切换时只改 provider 配置，确认没有重新引入 host-run runner 竞争；若无后续调试需求，再执行标准环境清理。
- 中期再评估结构化 evidence manifest；只有出现第二次 marker consumer 漂移时再立项，避免为当前已闭环的显式属性方案过度设计。
