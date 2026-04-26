# ADR-Q1 · 认证方案

> **版本**: 1.0
> **状态**: accepted
> **更新日期**: 2026-04-26

## 1 背景

`easyinterview` 是个人向 AI 面试训练产品，目标用户是高意图求职者（C 端单租户），核心场景是 24-72 小时内围绕一份 JD 完成准备闭环。`easyinterview-tech-docs/01-technical-architecture.md` §5 把 `auth` 列为后端模块之一，但未锁定方案；`easyinterview-tech-docs/README.md` §「待评审的 5 个决策点」第 1 项明确把认证方案留作 W0 决策。

产品语义上：

- 用户登录不是一次性体验，而是跨设备 / 多次回访（练习→报告→复练）
- 用户携带敏感数据（简历、JD、面试录音）；身份事故 = 数据事故
- 不存在企业 SSO / 组织管理需求（B2B / Team 版明确划入 Out of Scope）
- 邮箱是核心唯一标识（report 发送、隐私请求确认、感谢信草稿场景都依赖邮箱）

技术约束：

- 后端是 Go 模块化单体（`chi + pgx + Redis + Asynq`），已有 session-cookie 中间件落点
- `02-api-definition.md` §「Endpoint 总览」无 `/oauth/...` 端点，`POST /auth/...` 与 `GET /me` 已规划
- 03-db `users` / `user_settings` 表已就位，但缺少认证 challenge / token 表

## 2 选项与取舍

### 选项 A · 自建 passwordless（email magic link + first-party session cookie）

**Pros**：

- 零第三方依赖，邮箱即唯一身份；与 `users.email` 表结构天然对齐
- 用户无密码即无密码泄漏风险，免「忘记密码」全链路实现
- 可控 UX / i18n（中英双语登录页可与 D1 `frontend-shell` 共用主题）
- 无 per-MAU 费用；隐私链路（Q-5 删除）只需自家 DB 操作
- 与现有 Asynq + 邮件模板（debrief 感谢信复用同一发件渠道）天然复用

**Cons**：

- 需自建邮件服务集成（默认 Resend / Postmark / SES 之一）+ 速率限制 + 防爆破
- magic link 链路需要 challenge 表 + 单次失效 + IP/UA 绑定
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

- 包含 magic link / SMS / OTP / OIDC 全套
- 内置反爬 / 设备指纹 / 监控

**Cons**：

- 按 MAU 收费；个人向产品高 MAU 成本不可控
- 多一层 vendor lock-in，隐私链路（Q-5）受第三方 SLA 约束
- session 透传与 `chi` 中间件需要适配层
- 与 P0 「保守隐私 / 可解释 / 可关闭 / 可删除」红线冲突更高

## 3 决策

**P0 锁定选项 A：自建 passwordless email magic link + first-party session cookie。**

落地约束：

1. 唯一登录入口：`POST /api/v1/auth/email/start`（发 magic link）+ `GET /api/v1/auth/email/verify?token=...`（兑换 session）
2. session 存储：HttpOnly + SameSite=Lax + Secure 的 cookie，服务端 session 表（带 `last_seen_at` / `revoked_at`），默认 30 天滑动续期
3. challenge token：单次失效、15 分钟 TTL、绑定请求 IP + UA hash；进 outbox 异步发邮件
4. 邮件渠道：默认 Resend 抽象在 `notify` 模块，运维可切换 Postmark / SES（不锁厂商）
5. 风控基线：同邮箱 / 同 IP 1 分钟内 ≥ 3 次拒绝；进 audit log
6. SSO（Google / Apple OIDC）保留 P1 扩展点，不在 P0 实现，但 `users` 表预留 `external_identities` 关联表的 migration slot

## 4 影响范围

- **C1 `backend-auth`** —— 落地 magic link 全链路 + session 中间件 + challenge 表迁移
- **D1 `frontend-shell`** —— 实现 `/welcome` 登录页（`screens-welcome.jsx` 重构）+ auth gate + session 自动续期
- **B2 `openapi-v1-contract`** —— 在 `auth` tag 下冻结 `/auth/email/{start,verify}` + `/auth/logout` + `/me`
- **B4 `db-migrations-baseline`** —— `auth_challenges` / `sessions` / `external_identities`（空表）三张表的 0001 迁移
- **C8 `backend-async-runtime`** —— 邮件发送作为 `email_dispatch` job_type，复用 outbox dispatcher
- **F1 `observability-stack`** —— `auth_challenge_started_total` / `auth_session_minted_total` / `auth_failure_total` 指标接入
- **F4 `privacy-and-audit-runtime`** —— magic link / session 创建 / 撤销均进 audit_events

## 5 失效与修订条件

触发推翻或升级本 ADR 的具体阈值：

- 出现企业 / 团队版需求（>1 用户共享同一 workspace）→ 升级到 OIDC + 组织模型
- magic link 发送失败率 ≥ 5% 持续 2 周（邮件可达性问题）→ 评估增加 SMS OTP 或托管 Auth
- 出现合规要求（如 SOC2 / ISO27001 强制）需要现成审计闭环 → 评估 Clerk / Auth0
- 跨地域用户对邮箱依赖产生明显流失（漏斗对账显示 verify 步 < 60%）→ 评估补充 OIDC

修订流程：本 ADR 状态由 `accepted` → `superseded`，新 ADR 显式标注 `supersedes: ADR-Q1-auth.md`。

## 6 关联

- `engineering-roadmap/spec.md` §3.2 Q-1
- `engineering-roadmap/plans/001-decompose-subspecs/plan.md` Phase 1.1
- 上游：`easyinterview-tech-docs/01-technical-architecture.md` §「auth」、`02-api-definition.md` §「auth tag」、`03-db-definition.md` §「users」
- 下游 child：C1 / D1 / B2 / B4 / C8 / F1 / F4
