# Report Retry and Local Diagnostics 交付复盘报告

> **日期**: 2026-07-16
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖失败面试报告的同 ID 人工恢复、明确语义修复提示、状态无关的面试记录入口、中文 auth gate、dev/test 原始 Complete 请求/响应捕获、Mailpit host-run 路由和本地 10900/10901 运行环境。
- [BUG-0182](../bugs/BUG-0182.md) 的真实失败报告从 failed 经同 ID regenerate 到 generating/ready；raw evidence 证明最终只引用可信 user anchor，没有接口不兼容、输入截断或上一轮输出污染。
- [BUG-0183](../bugs/BUG-0183.md) 的 host SMTP mapping 修复后，fresh challenge 在 Mailpit 产生新邮件并通过真实浏览器登录；证据未保留完整邮箱或验证码。
- Chrome 插件在真实 frontend/backend 上验证 failed 行的“重新生成报告 + 查看面试记录”、generating 行的“查看生成进度 + 查看面试记录”、ready 报告详情和中文文案，无请求拦截或 mock backend。
- 真实 PostgreSQL integration tests 证明不同幂等键并发时最多一个 active report job，并验证冻结 transcript、跨用户、超限、非 failed 和 active old job 的原子边界。
- 最终根 `make test` 通过：Python 584 tests / 4583 subtests、Go 全包、frontend 126 files / 1026 tests；`make build`、OpenAPI lint/inventory、38 fixtures、16 项 live Prism parity、openapi diff 0 和生成前后 hash 幂等检查通过。

## 2 会话中的主要阻点/痛点

- 连续多次报告失败起初看起来像 AI client 接口或 prompt 污染，但 provider/schema 都成功，真正错误发生在业务语义 validator。
  - **证据**：连续调用请求结构稳定且不附带上一轮模型输出；每次都选中末尾 assistant 序号并返回 `not_user_message`。
  - **影响**：如果只增加自动/人工重试，会稳定浪费调用并重复失败，无法恢复核心报告功能。
- Mailpit Web UI 健康掩盖了 host-run backend 使用错误 SMTP 地址空间的问题。
  - **证据**：Compose internal route 正确，动态 host mapping 与 backend SMTP port 不一致；任务存在但连接失败。
  - **影响**：doctor 看似通过、Inbox 却为空，容易把配置漂移误判为异步任务或 Mailpit 故障。
- OpenAPI 历史 decision tests 曾读取当前 worktree source，而不是不可变历史 snapshot。
  - **证据**：新增第 38 个 operation 后，旧 37-operation 审计错误吸收当前 additive error code；改为 Git SHA 链重放后 61 个 wrapper tests 通过。
  - **影响**：合法的新契约会使历史 gate 伪失败，甚至诱导修改旧审计结论。
- `make codegen-check` 的最后一步直接比较 HEAD，无法区分合法未提交生成物和生成器二次漂移。
  - **证据**：OpenAPI 生成/lint/inventory 均通过，但 dirty worktree 的预期新增代码触发 `git diff --exit-code`；独立生成前后 diff hash 完全一致。
  - **影响**：实施阶段不能直接把该 target 作为幂等 PASS，需要额外证据解释，增加收口成本。
- 终审补抓到两个非阻断竞态：raw capture 失败事件落入无人消费的内存 logger；failed transcript owner 未解析前 Back 会短暂默认 workspace。
  - **证据**：两个 RED test 分别证明进程日志为零和 owner promise pending 时 Back 已可点击。
  - **影响**：诊断承诺不可观测，且快速点击可能进入错误页面。

## 3 根因归类

- 报告重复失败：repair contract 只描述内部错误坐标，没有模型可执行的意图、可信 anchor 值域和多问题组合规则。
  - **类别**：spec-plan。
- Mailpit 空邮箱：host/container 地址空间没有在 host backend 启动前闭合；环境健康检查只覆盖服务健康，不覆盖消费者实际 route。
  - **类别**：README / spec-plan。
- 历史 OpenAPI gate 漂移：审计测试缺少 immutable source ownership，把当前 source 当成历史事实。
  - **类别**：spec-plan。
- dirty-worktree codegen gate：一个 target 同时承担生成器执行、lint 和 HEAD 清洁度，未提供实施期幂等模式。
  - **类别**：README / tooling plan。
- 两个终审 P2：合同写了“安全事件”和“可信 Back”，但没有锁定 process-visible sink 与 resolving 首帧。
  - **类别**：spec-plan。

## 4 对流程资产的改进建议

- 为 `make codegen-check` 增加 worktree-friendly 幂等模式：执行前记录 owner 生成物内容，二次生成后比较前后，而不是要求相对 HEAD 零差异；现有 CI clean-tree 模式继续保留。
  - **落点**：OpenAPI/current codegen owner plan + 根 Makefile。
  - **优先级**：high。
- 在 backend-review owner 中保留“transcript 以未回答 assistant 结束”的 regression corpus，并强制每个可达 validator issue family 具有明确意图、可信值域与 provider-before-call fail-closed 测试。
  - **落点**：backend-review spec-plan。
  - **优先级**：high。
- local environment doctor 输出经过脱敏的已解析 Mailpit host route，并把 loopback Mailpit mapping equality 保持为 runtime pre-start invariant；不要把端口写入 Skill。
  - **落点**：local-dev-stack README/spec-plan。
  - **优先级**：medium。
- OpenAPI 历史 decision/audit 测试继续从固定 Git SHA 或 preserved artifact 加载，不允许通过当前 source 推导旧 inventory；新增 contract 时必须有一个 unrelated-additive negative test。
  - **落点**：openapi-v1-contract spec-plan。
  - **优先级**：medium。
- 对异步 owner 解析的导航控件统一要求 `resolving` 可观察状态；在可信 destination 未确定前隐藏或禁用动作。
  - **落点**：frontend-report-dashboard spec-plan。
  - **优先级**：low。

## 5 建议优先级与后续动作

- 下一步最高价值动作：使用 `/change-intake` 原地重开当前 OpenAPI/codegen owner，把 worktree-friendly idempotence 模式加入现有 plan/checklist，避免下一次合法 API 变更再次遇到假 drift。
- 次优先：在 local-dev-stack 原 owner 中补一条脱敏 doctor 输出断言，让“依赖健康”和“host consumer route 可达”同时可见。
- 已在本轮闭合、无需再开新计划：报告 repair intent/user-anchor 合同、raw capture process WARN、failed conversation Back resolving fence、Mailpit mapping guard 和历史 OpenAPI snapshot 重放。
