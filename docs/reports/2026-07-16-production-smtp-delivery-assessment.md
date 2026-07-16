# Production SMTP Delivery 交付复盘报告

> **日期**: 2026-07-16
> **审查人**: Codex

**关联计划**: [backend-auth/001-email-code-session-bootstrap](../spec/backend-auth/plans/001-email-code-session-bootstrap/plan.md)

## 1 复盘范围与成功证据

- 本次交付为既有 email-code 主流程增加 `mailpit|smtp` 环境选择、标准 SMTP STARTTLS/隐式 TLS 与用户名密码认证，并保持邮件仅包含 6 位验证码和有效期。
- A4 typed config owner 覆盖 provider 条件字段、TLS mode、staging/prod fail-fast 与 secret redaction；backend transport/runtime wiring 覆盖 Mailpit、STARTTLS、implicit TLS、AUTH、TLS 1.2 下限和脱敏失败路径。
- 根 `make test`、`make build`、`make lint-config`、docs/context/index/diff、Compose 与 scenario env contract 均通过。
- live 证据覆盖 Mailpit challenge->收码->verify->session->me，以及用户 `.env` 外部 SMTP 的 TLS/auth、授权 sender、provider DATA 接收和应用 job 一次成功；用户确认真实收件箱收到 “EasyInterview sign-in code”。
- owner spec/plan/checklist 已恢复 `completed`，MVP 单 active backend 限制已写入 backend-auth 与 local-dev-stack 当前合同。

## 2 会话中的主要阻点/痛点

### 2.1 SMTP TLS mode 需要以真实服务能力纠正

- **证据**：最初按 STARTTLS 探测超时，implicit TLS 可完成握手与认证，因此本地 `.env` 的 `EMAIL_SMTP_TLS_MODE` 改为 `tls`。
- **影响**：仅靠常见端口/协议印象无法证明 provider 模式，必须执行真实握手；首次 live 尝试产生一次配置返工。

### 2.2 provider 拒绝未授权 sender

- **证据**：连接和认证成功后，初始 `EMAIL_FROM_ADDRESS` 在 `MAIL FROM` 阶段收到 provider 553；改为 provider 已授权 identity 后 DATA 被接受。
- **影响**：SMTP 凭据有效不代表任意 From 地址可用；若不拆分阶段错误，容易误判为网络或密码问题。

### 2.3 两个 backend runner 竞争进程内 delivery secret

- **证据**：full-container 创建的 job 被 host-run backend lease；停止仓库 PID 文件管理的 host-run app 后，Mailpit 与外部 SMTP 均一次成功。详见 [BUG-0178](../bugs/BUG-0178.md)。
- **影响**：Compose/HTTP healthy 仍会出现邮件不投递；诊断跨越环境生命周期、队列 lease 与 transient secret 所有权。

## 3 根因归类

- TLS mode 与授权 sender 属于外部 provider 配置事实，现有分阶段 transport 错误已能定位，归类为 `no repo change needed`。
- 双 runner 竞争属于 `spec-plan` 和本地环境生命周期合同缺口；本轮已将单实例边界写回 owner，并由 `dev-container-up` 与 scenario contract 固化。
- 多副本共享验证码是明确延后的生产扩容架构，不是当前 MVP 缺陷；在没有扩容需求前不新增依赖。

## 4 对流程资产的改进建议

- 生产部署 runbook 应明确 `EMAIL_SMTP_TLS_MODE` 必须按 provider 文档/握手结果选择，并要求 `EMAIL_FROM_ADDRESS` 是 provider 已授权 sender。**落点**：未来 release/deployment README；**优先级**：medium。
- 启用第二个 backend 副本前，原地重开 backend-auth/001，设计共享一次性 delivery secret、TTL、消费后销毁、重放与跨实例失败语义。**落点**：backend-auth spec/plan；**优先级**：扩容前 high，MVP 当前不实施。
- 保留 Mailpit 与外部 SMTP 两条 live smoke：前者验证完整登录闭环，后者验证真实 provider 连接与收件，不把 TCP/auth 成功冒充邮件闭环。**落点**：backend-auth/local-dev-stack checklist；**优先级**：已完成。

## 5 建议优先级与后续动作

1. 当前 MVP 直接维持单个 active backend 实例，使用已验证的 `.env` SMTP 配置上线试用，不增加 Redis 等新依赖。
2. 上线前由部署 owner 把同一组 SMTP env 以 secret manager 或平台 secret 注入，并做一次目标环境冷启动发码验收。
3. 只有确定要横向扩容 backend 时，才执行 backend-auth 原 plan 的多副本 secret-store 修订；不得直接复制当前进程内实现。
