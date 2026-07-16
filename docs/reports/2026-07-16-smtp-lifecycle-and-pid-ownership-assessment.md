# SMTP Lifecycle and PID Ownership 交付复盘报告

> **日期**: 2026-07-16
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次复盘覆盖 [BUG-0180](../bugs/BUG-0180.md)：`email_dispatch` runner context 贯穿 Redis、SQL 与 SMTP，完整 SMTP 会话有界，DATA 已接受后不因 QUIT 失败重复投递，delivery secret/challenge 跨存储写入具备补偿，以及 `dev-container-up` 停止 host-run runtime 前校验 PID 对应命令。
- 当前工作树根 `make test` 通过：Python 566 tests / 4481 subtests、Go 全包、frontend 126 files / 1004 tests。
- `make build`、`make lint-config`、`make docs-check` 与 `git diff --check` 通过；`python3 -m pytest scripts/lint/scenario_env_contract_test.py -q` 为 13 passed，Header/INDEX 检查为 zero drift。
- `backend-auth/001-email-code-session-bootstrap` 与 `local-dev-stack/001-bootstrap` 已原地修订并恢复 `completed`，Bug 记录、owner history、spec、plan、checklist、context 与 INDEX 已同步。

## 2 会话中的主要阻点/痛点

- 既有 SMTP 测试覆盖配置、TLS 与发送错误，但没有把 runner cancellation 作为所有下游 I/O 的共同生命周期。
  - **证据**：旧 handler 丢弃 context，writer 的 Redis/SQL/SMTP 调用使用后台 context；建连后不发送 greeting 的 RED test 暴露取消无法推进。
  - **影响**：服务端在 greeting、TLS、AUTH 或 DATA 阶段停滞时，job handler 和 graceful shutdown 可能长期阻塞。
- 既有实现没有按远端副作用提交点定义 retry 边界。
  - **证据**：SMTP DATA 获得最终成功响应后，QUIT EOF 仍被映射为 retryable error。
  - **影响**：runner 重试可能重复发送同一验证码。
- Redis secret、challenge 和 pidfile 都存在“存在即归属/成功”的隐含假设。
  - **证据**：Redis Put 发生在 challenge 创建之后；`_stop_host_runtimes` 只用 `kill -0` 判断 PID；专项 RED tests 分别留下限流记录并终止无关 `sleep` 进程。
  - **影响**：失败请求污染限流，PID 复用时可能终止仓库外进程。

## 3 根因归类

- SMTP 与异步 runner 的生命周期合同缺少端到端取消和 accepted-once 不变量。
  - **类别**：spec-plan。当前 backend-auth spec/plan/checklist 已补入 context、deadline、DATA 成功边界与 focused regression。
- 跨存储副作用顺序没有显式区分可过期 secret 与会影响限流的 challenge 事实。
  - **类别**：spec-plan。当前 D-12、C-12 和 Phase 12 remediation 已固化先 Put、后 Create、失败补偿的顺序。
- pidfile contract 过去只验证静态字符串，没有真实进程归属负向测试。
  - **类别**：spec-plan。当前 local-dev owner checklist 与 scenario environment contract 已覆盖复用 PID 和 owner PID 两条路径。
- 本次修复按既有 `/change-intake`、`/tdd`、`/bug-report` 与根级 gate 完成，没有证据表明需要修改通用 Skill、README 或 `AGENTS.md`。
  - **类别**：无需仓库改动。

## 4 对流程资产的改进建议

- 后续修改异步协议 adapter 时，保留 `runner context -> storage lookup -> transport` 的完整传递断言，并以远端协议提交点定义是否可重试。
  - **落点**：backend-auth spec-plan
  - **优先级**：high
- 后续引入跨存储写入时，在 owner plan 中显式列出副作用可见性、补偿 context 与失败后的业务残留，避免仅覆盖 happy path。
  - **落点**：feature owner spec-plan
  - **优先级**：high
- 后续扩展 pidfile 管理角色时，复用“命令身份匹配 + 无关真实进程负向测试”，不要退回 `kill -0` 单条件判断。
  - **落点**：local-dev-stack spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 当前最高优先级是把本次代码、owner 文档、BUG-0180、复盘和工作日志作为一个原子提交，保留完整追溯链。
- 下一轮若继续增强邮件投递，优先围绕 provider retry/idempotency 审查现有 async job 重试预算与可观察性；在出现新需求或新失败证据前，不提前扩张通用治理规则。
