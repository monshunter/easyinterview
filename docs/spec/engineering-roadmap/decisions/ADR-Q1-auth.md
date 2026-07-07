# ADR-Q1 · 认证方案

> **版本**: 1.8
> **状态**: accepted
> **更新日期**: 2026-07-07

## 1 背景

`easyinterview` 是个人向 AI 面试训练产品，目标用户是高意图求职者（C 端单租户），核心场景是 24-72 小时内围绕一份 JD 完成准备闭环。`engineering-roadmap decisions` §5 把 `auth` 列为后端模块之一，但未锁定方案；`engineering-roadmap decision brief` §「待评审的 5 个决策点」第 1 项只作为决策输入。

产品语义上：

- 用户登录不是一次性体验，而是跨设备 / 多次回访（练习→报告→复练）
- 用户携带敏感数据（简历、JD、面试录音）；身份事故 = 数据事故
- 不存在企业 SSO / 组织管理需求（B2B / Team 版明确划入 Out of Scope）
- 邮箱是核心唯一标识（report 通知、email code、隐私请求确认等场景都依赖邮箱）

技术约束：

- 后端是 Go 模块化单体（`chi + pgx + Redis`，后台任务通过 B3 job/outbox contract + backend internal runner 承接），已有 session-cookie 中间件落点
- `B2 openapi-v1-contract` §「Endpoint 总览」无 `/oauth/...` 端点，`POST /auth/...` 与 `GET /me` 已规划
- 03-db `users` / `user_settings` 表已就位，但缺少认证 challenge / token 表

## 2 选项与取舍

### 选项 A · 自建 email-code（first-party session cookie）

**Pros**：

- 零第三方依赖，邮箱即唯一身份；与 `users.email` 表结构天然对齐
- 用户无密码即无密码泄漏风险，免「忘记密码」全链路实现
- 可控 UX / i18n（中英双语登录页可与 D1 `frontend-shell` 共用主题）
- 无 per-MAU 费用；隐私链路（Q-5 删除）只需自家 DB 操作
- 与现有邮件模板和 `email_dispatch` 发件渠道天然复用

**Cons**：

- 需自建邮件服务集成（默认 Resend / Postmark / SES 之一）+ 速率限制 + 防爆破
- email code 链路需要 challenge 表 + 单次失效 + IP/UA 绑定
- 无法天然支持企业 SSO（P0 不需要）

### 选项 B · 第三方 OIDC（Google / GitHub / Apple）

**Pros**：

- 用户无须收件箱往返，登录最快
- 借用 IdP 的反爬 / 风控

**Cons**：

- 至少需要先实现 1 家 OIDC，多家时复杂度叠加；产品定位是个人工具，不强需要 OAuth 圈
- 隐私边界外溢：邮箱 / 头像被 IdP 同步；删除链路需考虑 token 解绑
- 部分目标地区（如国内）无可靠 IdP

### 选项 C · 托管 Auth（Clerk / Auth0 / Supabase Auth）

**Pros**：

- 包含 email-code / SMS / OTP / OIDC 全套
- 内置反爬 / 设备指纹 / 监控

**Cons**：

- 按 MAU 收费；个人向产品高 MAU 成本不可控
- 多一层 vendor lock-in，隐私链路（Q-5）受第三方 SLA 约束
- session 透传与 `chi` 中间件需要适配层
- 与 P0 「保守隐私 / 可解释 / 可关闭 / 可删除」红线冲突更高

## 3 决策

**P0 锁定选项 A：自建 email-code + first-party session cookie。**

落地约束：

1. 唯一登录入口：`POST /api/v1/auth/email/start`（发 6 位 email code）+ `GET /api/v1/auth/email/verify?token=...`（消费 code 并兑换 session）
2. session 存储：cookie 字面量名固定为 `ei_session`；属性为 HttpOnly + SameSite=Lax + Secure，服务端 session 表（带 `last_seen_at` / `revoked_at`）是真理源，默认 30 天滑动续期
3. challenge code：单次失效、短 TTL、绑定请求 IP + UA hash；通过 backend-internal mail dispatcher / `email_dispatch` job contract 异步发邮件
4. 邮件渠道：默认 Resend 抽象在 `notify` 模块，运维可切换 Postmark / SES（不锁厂商）
5. 风控基线：同邮箱 / 同 IP 1 分钟内 ≥ 3 次拒绝；进 audit log
6. SSO（Google / Apple OIDC）保留 P1 扩展点，不在 P0 实现，但 `users` 表预留 `external_identities` 关联表的 migration slot

## 4 影响范围

- **C1 `backend-auth`** —— 落地 email code 全链路 + session 中间件 + challenge 表迁移
- **D1 `frontend-shell`** —— 实现操作级 AuthGate / 单入口登录 + 首次资料补全 + session 自动续期；默认入口仍为 `home`
- **B2 `openapi-v1-contract`** —— 在 `auth` tag 下冻结 `/auth/email/{start,verify}` + `/auth/logout` + `/me`
- **B4 `db-migrations-baseline`** —— `auth_challenges` / `sessions` / `external_identities`（空表）三张表的 0001 迁移
- **backend-runtime-topology / backend async runner** —— 邮件发送作为 internal-only canonical `email_dispatch` job_type（由 B3 / B4 加入内部契约，不进入 B2 API-facing `JobType`）；P0 backend-auth 可用 backend-internal dispatcher 完成本地闭环
- **F1 `observability-stack`** —— `auth_challenge_started_total` / `auth_session_minted_total` / `auth_failure_total` 指标接入
- **advanced audit future candidate** —— email challenge / session 创建 / 撤销均进 `audit_events`；后续 advanced audit workstream 只消费既有事件，不重新定义认证审计协议

## 5 失效与修订条件

触发推翻或升级本 ADR 的具体阈值：

- 出现企业 / 团队版需求（>1 用户共享同一 workspace）→ 升级到 OIDC + 组织模型
- email code 发送失败率 ≥ 5% 持续 2 周（邮件可达性问题）→ 评估增加 SMS OTP 或托管 Auth
- 出现合规要求（如 SOC2 / ISO27001 强制）需要现成审计闭环 → 评估 Clerk / Auth0
- 跨地域用户对邮箱依赖产生明显流失（漏斗对账显示 verify 步 < 60%）→ 评估补充 OIDC

修订流程：如需推翻本决策，新增修订 ADR 并同步 roadmap Q-1 与相关 owner spec。

## 6 关联

- `engineering-roadmap/spec.md` §3.2 Q-1
- `engineering-roadmap/plans/001-decompose-subspecs/plan.md` checklist 1.1
- 参考背景：`engineering-roadmap decisions` §「auth」、`B2 openapi-v1-contract` §「auth tag」、`B4 db-migrations-baseline` §「users」
- 下游 child / future candidate：backend-auth / frontend-shell / B2 / B4 / backend-runtime-topology / F1 / advanced audit

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-07 | 1.7 | 对齐当前 backend-auth / frontend-shell：认证采用 email-code challenge、单入口登录与首次资料补全。 | product-scope/001-core-loop-module-pruning |
| 2026-05-06 | 1.6 | 对齐 backend-runtime-topology：email code 邮件派发不等待独立 worker 进程，P0 由 backend-internal dispatcher 承接，同时保留 `email_dispatch` jobType 契约。 | backend-runtime-topology/001-worker-consolidation |
| 2026-05-04 | 1.5 | 对齐 engineering-roadmap v3.0：将原 F4 child 编号改为 advanced audit future candidate，不提前创建空 spec。 | engineering-roadmap v3.0 L2 remediation |
| 2026-05-03 | 1.4 | 对齐 product-scope v1.1：认证是操作级拦截，默认入口由 Home 承接；邮件身份场景移除感谢信草稿范围。 | product-scope / engineering-roadmap v2.2 |
| 2026-04-29 | 1.3 | 将 `email_dispatch` 从 public jobType 修正为 internal-only canonical jobType：仅 DB/C8 内部使用，不进入 B2 API-facing `JobType`；继续复用 outbox dispatcher。 | event-and-outbox-contract/001-bootstrap plan-review remediation |
| 2026-04-28 | 1.2 | 锁定 first-party session cookie 字面量名为 `ei_session`，供 B2 OpenAPI security scheme、A4 config 文档与后续 C1 backend-auth 实现复用。 | openapi-v1-contract/001-bootstrap assessment remediation |
