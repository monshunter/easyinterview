# ADR-Q5 · 隐私节奏

> **版本**: 1.9
> **状态**: accepted
> **更新日期**: 2026-07-07

## 1 背景

`B1 shared-conventions-codified` §「隐私请求类型」定义 `export` / `delete` 两个枚举；`B2 openapi-v1-contract` §「privacy tag」预留 `POST /privacy/exports` + `POST /privacy/deletions`；`B4 db-migrations-baseline` §「privacy_requests」表已就位；`F1 observability-stack` §「Privacy Completion」要求「99% 在 24h 内完成」。`README.md` §「待评审的 5 个决策点」第 5 项只作为决策输入。

产品红线（`docs/spec/product-scope/spec.md` §4.4 / §9.3）：

- 默认保守：音视频 / 简历 / 用户数据 / 面试内容必须可解释、可关闭、可删除
- 不做隐形作弊 / 不做企业候选人评估 / 不做实时面试辅助

业务背景：

- P0 用户群体高意图但样本小（早期种子用户），合规风险主要是 GDPR-like 删除请求
- 数据分布在 B4 baseline 的应用 / auth 支撑表中，跨 9 业务域；导出需要稳定的 schema 版本化 + 引用完整性
- 完整 privacy export / advanced audit 仍按 `engineering-roadmap/spec.md` §5.3 的 future candidate 后置；P0 删除链路由 B2 / B4 / C8 / F1 和后续 release gate workstream 承接，不提前创建 C12 / F4 空 spec
- 删除链路风险：物理删除 vs 软删除 vs 冷归档，需要在 P0 锁定一种语义

## 2 选项与取舍

### 选项 A · P0 仅做删除链路；导出延后到 P1

**Pros**：

- 核心法律红线（用户「我要离开」）有完整闭环
- 实现集中：privacy_requests 表 + backend internal delete runner + audit_event；不涉及跨表数据 assembly / 签名 URL / 大文件下载
- 与 P0 团队规模匹配
- 与 roadmap §5.3 的 future candidate 策略一致：P0 只落地可解释的删除链路，不提前空壳化 privacy export / advanced audit
- 用户感知友好：登出 / 删除按钮即时生效，「导出」UI 可显示「即将上线」占位

**Cons**：

- 部分国家 / 地区监管要求同时具备 access + portability（GDPR Art. 15 + Art. 20）；如果首批用户中包含 EU 用户，可能受质询
- 后续补 export 需要重新设计 dump schema，二次成本

### 选项 B · P0 完整导出 + 删除（C12 / F4 升格 P0）

**Pros**：

- 一次到位，完全符合 GDPR 全套要求
- 用户信任度更高（数据可移植 = 数据不被绑架）

**Cons**：

- 会提前创建尚未启动的 privacy export / advanced audit workstream，与 roadmap v3.0 的 on-demand child 创建策略冲突
- 需要冻结 export schema 版本（与 B2 OpenAPI v1.0.0 freeze 同时锁定）
- 跨 B4 baseline 多表 dump 涉及 audio / 简历 / 报告原文等多种格式，复杂度近一个独立子系统
- 会挤压当前 P0 闭环中的 backend-auth、backend async runner、backend-practice / review 与 release gate workstream

### 选项 C · 不做隐私链路（仅靠手动后台处理）

**Pros**：

- 零开发成本

**Cons**：

- 与产品红线直接冲突；不可接受

## 3 决策

**P0 锁定选项 A：P0 仅落地删除链路；导出延后到 P1。**

落地约束：

1. **删除语义**：用户级 `DELETE /api/v1/me`（同义于 `POST /api/v1/privacy/deletions` body `{type: "delete"}`）→ 同步软删 `users.deleted_at` 同时立即吊销所有 session → 返回 `202 + PrivacyRequestWithJob` → backend internal runner 异步逐域硬删（按 B4 `db-migrations-baseline` §3.1.2 table matrix）
2. **删除范围**（P0 必须覆盖）：B4 baseline 的所有用户关联表、ADR-Q1 auth/session 支撑表、对象存储文件与 AI call/audit/job/outbox 运行痕迹；全局 prompt/rubric 版本与 migration 元数据保留。每表处理策略以 [B4 §3.1.2 P0 privacy deletion table matrix](../../db-migrations-baseline/spec.md#312-p0-privacy-deletion-table-matrix) 为准，`audit_events` 只保留不可反推用户身份的删除完成 tombstone。
3. **保留例外**：billing 类（如未来引入）/ 法律强制留存的合规日志按对应法规另行 ADR；P0 暂无
4. **SLA**：删除请求 99% 在 24h 内完成（与 `F1 observability-stack` §「Privacy Completion」对齐）；超期写 audit + Sentry alert
5. **导出占位 / 产品例外**：`POST /privacy/exports` 在 OpenAPI v1.0.0 freeze 中**预留 endpoint 但返回 `501 Not Implemented`**；Settings / Privacy UI 显示「导出功能即将上线」（i18n 文案锁定）。这显式覆盖产品 spec P0 验收项「删除与导出路径可用」中的导出部分：P0 只保证删除路径可运行，导出路径仅保证契约预留、可观测、可解释地不可用；release gate workstream 必须把该 tradeoff 记录为准入例外
6. **审计**：所有 privacy 状态变迁（request_created / step_started / step_completed / completed / failed）全部进 `audit_events`；advanced audit workstream 后续触发时复用这些事件，不重新定义并行审计协议
7. **自助 vs 工单**：P0 用户自助 UI 直接触发；同时保留 `support@` 邮箱兜底路径
8. **可观测性**：`privacy_request_duration_seconds` + `privacy_request_in_flight` 指标 P0 即上

## 4 影响范围

- **privacy export / advanced audit future candidates** —— 不提前创建空 spec；触发条件成立时先修订 product-scope / roadmap，再创建对应 child spec / plan
- **B2 `openapi-v1-contract`** —— 冻结 `/privacy/deletions` + `/privacy/exports`（后者 stub 501）+ `DELETE /api/v1/me`
- **B4 `db-migrations-baseline`** —— `privacy_requests` / `audit_events` 0001 迁移
- **backend async runner future subject** —— public `privacy_delete` job_type；内部 handler 可为 `privacy.delete`；优先级 critical；P0 不要求独立 worker 进程
- **frontend-shell / Settings & Privacy** —— 「删除我的账号」UI + 「导出即将上线」占位；不得把导出延后误写成独立业务功能
- **F1 `observability-stack`** —— privacy 指标接入
- **`release-gate-and-rollout`** —— 创建时校验 P0 删除链路 SLA、audit 完整性与 privacy export 501 例外说明
- **`engineering-roadmap/spec.md`** —— §3.2 Q-5 记录「P0 删除-only；导出延后并以 501 / UI 占位解释」，§5.3 将完整 privacy export / advanced audit 作为 future candidate

## 5 失效与修订条件

触发推翻或升级本 ADR 的具体阈值：

- 出现 ≥1 例 EU 用户基于 GDPR Art. 20 的正式 portability 请求 → 评估提前升格 export
- 监管环境变更（如所在区域强制双向）→ 升级 export 到 P0
- 用户调研显示 ≥ 30% 受访者把「数据可移植」作为关键决策因素 → 升格 export
- 删除 SLA 24h 在生产无法稳定达到 → 重新评估 backend internal runner 拓扑与跨表 dependency

修订流程：如需推翻本决策，新增修订 ADR；同步 roadmap §3.2 Q-5、§5.3 future candidates、B2/B4/C8/F1 相关 spec 与 Settings / Privacy UI 文档。

## 6 关联

- `engineering-roadmap/spec.md` §3.2 Q-5、§5.3 future candidates、§5.2 release gate workstream
- `engineering-roadmap/plans/001-decompose-subspecs/plan.md` checklist 1.1、checklist 3.3
- 当前约束与参考背景：`docs/spec/product-scope/spec.md` §4.4 / §9.3、`B1 shared-conventions-codified` §「隐私请求」、`B2 openapi-v1-contract` §「privacy」、`B4 db-migrations-baseline` §「privacy_requests」、`F1 observability-stack` §「Privacy Completion」
- 下游 child / future candidate：B2 / B4 / backend async runner / frontend-shell Settings / F1 / release-gate-and-rollout / privacy export / advanced audit

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-07 | 1.8 | 对齐当前产品隐私红线和核心闭环，删除非当前模块叙述。 | product-scope/001-core-loop-module-pruning |
| 2026-05-06 | 1.7 | 对齐 backend-runtime-topology：P0 删除链路执行方从独立 worker 改为 backend internal runner，`privacy_delete` jobType 保留。 | backend-runtime-topology/001-worker-consolidation |
| 2026-05-04 | 1.6 | 对齐 engineering-roadmap v3.0：删除对原 C12/F4/W4 phase 的当前执行口径引用，改为 privacy export / advanced audit future candidate + P0 删除链路由现有契约和后续 release gate 承接。 | engineering-roadmap v3.0 L2 remediation |
| 2026-05-03 | 1.4 | 同步产品真理源迁移：隐私产品红线引用改为 `docs/spec/product-scope/spec.md`，不改变 Q5 的 P0 删除-only 决策。 | docs-only |
| 2026-04-29 | 1.3 | 对齐 B2 / B4 remediation：明确 `DELETE /api/v1/me` 必须进入 OpenAPI freeze 并返回 `202 + PrivacyRequestWithJob`；删除范围改为引用 B4 §3.1.2 per-table matrix，区分 hard delete / cascade / retain / audit tombstone。 | plan-review remediation |
| 2026-04-29 | 1.2 | 对齐 B4 `db-migrations-baseline` v1.4：移除原「29 表」背景口径，改为引用 B4 baseline 多表范围；删除范围中的 `resumes` 改为当前表名 `resume_assets`，并纳入 ADR-Q1 指派给 B4 的 `external_identities` 支撑表。 | db-migrations-baseline plan-review remediation |
