# ADR-Q5 · 隐私节奏

> **版本**: 1.2
> **状态**: accepted
> **更新日期**: 2026-04-29

## 1 背景

`easyinterview-tech-docs/00-shared-conventions.md` §「隐私请求类型」定义 `export` / `delete` 两个枚举；`02-api-definition.md` §「privacy tag」预留 `POST /privacy/exports` + `POST /privacy/deletions`；`03-db-definition.md` §「privacy_requests」表已就位；`04-metrics-observability.md` §「Privacy Completion」要求「99% 在 24h 内完成」。`README.md` §「待评审的 5 个决策点」第 5 项把隐私节奏选择留作 W0 决策。

产品红线（`easyinterview-spec-v1-0.md` §伦理）：

- 默认保守：音视频 / 简历 / 画像 / 面试内容必须可解释、可关闭、可删除
- 不做隐形作弊 / 不做企业候选人评估 / 不做实时面试辅助

业务背景：

- P0 用户群体高意图但样本小（早期种子用户），合规风险主要是 GDPR-like 删除请求
- 数据分布在 B4 baseline 的应用 / auth 支撑表中，跨 9 业务域；导出需要稳定的 schema 版本化 + 引用完整性
- C12 `backend-privacy` / F4 `privacy-and-audit-runtime` 默认归 P1（`engineering-roadmap/spec.md` §5.3 §5.6）
- 删除链路风险：物理删除 vs 软删除 vs 冷归档，需要在 P0 锁定一种语义

## 2 选项与取舍

### 选项 A · P0 仅做删除链路；导出延后到 P1

**Pros**：

- 核心法律红线（用户「我要离开」）有完整闭环
- 实现集中：privacy_requests 表 + delete worker + audit_event；不涉及跨表数据 assembly / 签名 URL / 大文件下载
- 与 P0 团队规模匹配
- 与 spec §5.6 默认（C12 / F4 = P1）天然一致
- 用户感知友好：登出 / 删除按钮即时生效，「导出」UI 可显示「即将上线」占位

**Cons**：

- 部分国家 / 地区监管要求同时具备 access + portability（GDPR Art. 15 + Art. 20）；如果首批用户中包含 EU 用户，可能受质询
- 后续补 export 需要重新设计 dump schema，二次成本

### 选项 B · P0 完整导出 + 删除（C12 / F4 升格 P0）

**Pros**：

- 一次到位，完全符合 GDPR 全套要求
- 用户信任度更高（数据可移植 = 数据不被绑架）

**Cons**：

- C12 / F4 进入 W4 关键路径（`engineering-roadmap/spec.md` §6 C-6 验收：「Q-5 ADR 决定的 P0 隐私范围已验证」）
- 需要冻结 export schema 版本（与 B2 OpenAPI v1.0.0 freeze 同时锁定）
- 跨 B4 baseline 多表 dump 涉及 audio / 简历 / 报告原文等多种格式，复杂度近一个独立子系统
- W2-W4 团队带宽吃紧，可能挤压 C4-C7 业务域

### 选项 C · 不做隐私链路（仅靠手动后台处理）

**Pros**：

- 零开发成本

**Cons**：

- 与产品红线直接冲突；不可接受

## 3 决策

**P0 锁定选项 A：P0 仅落地删除链路；导出延后到 P1。**

落地约束：

1. **删除语义**：用户级 `DELETE /api/v1/me`（同义于 `POST /privacy/deletions` body `{type: "delete"}`）→ 同步软删 `users.deleted_at` 同时立即吊销所有 session → 异步 worker 逐域硬删（按 `03-db-definition.md` §「删除策略」表）
2. **删除范围**（P0 必须覆盖）：users / user_settings / candidate_profiles / experience_cards / target_jobs / resume_assets / file_objects / practice_sessions / practice_session_events / question_assessments / feedback_reports / mistake_entries / async_jobs / outbox_events / sessions / auth_challenges / external_identities / audit_events（最后才删；先写 `delete_completed` audit）
3. **保留例外**：billing 类（如未来引入）/ 法律强制留存的合规日志按对应法规另行 ADR；P0 暂无
4. **SLA**：删除请求 99% 在 24h 内完成（与 `04-metrics-observability.md` §「Privacy Completion」对齐）；超期写 audit + Sentry alert
5. **导出占位 / 产品例外**：`POST /privacy/exports` 在 OpenAPI v1.0.0 freeze 中**预留 endpoint 但返回 `501 Not Implemented`**；前端 D6 settings 显示「导出功能即将上线」（i18n 文案锁定）。这显式覆盖产品 spec P0 验收项「删除与导出路径可用」中的导出部分：P0 只保证删除路径可运行，导出路径仅保证契约预留、可观测、可解释地不可用；E4 release gate 必须把该 W0 tradeoff 记录为准入例外
6. **审计**：所有 privacy 状态变迁（request_created / step_started / step_completed / completed / failed）全部进 `audit_events`；F4 P1 升格时这些事件已就位
7. **自助 vs 工单**：P0 用户自助 UI 直接触发；同时保留 `support@` 邮箱兜底路径
8. **可观测性**：`privacy_request_duration_seconds` + `privacy_request_in_flight` 指标 P0 即上

## 4 影响范围

- **C12 `backend-privacy`** —— 维持 P1 占位（draft spec），但删除链路核心实现下沉到 P0：在 C8 `backend-async-runtime` 中以 public `jobType=privacy_delete` 落地，内部 Asynq handler 可映射为 `privacy.delete`；C12 升格时只接管 export + 完整 worker 管理面
- **F4 `privacy-and-audit-runtime`** —— 维持 P1，但 audit_events schema 与 retention 在 W1 锁定（B4 / B1）
- **B2 `openapi-v1-contract`** —— 冻结 `/privacy/deletions` + `/privacy/exports`（后者 stub 501）+ `DELETE /me`
- **B4 `db-migrations-baseline`** —— `privacy_requests` / `audit_events` 0001 迁移
- **C8 `backend-async-runtime`** —— public `privacy_delete` job_type；内部 handler 可为 `privacy.delete`；优先级 critical
- **D6 `frontend-debrief-and-growth`**（P1） + **D1 `frontend-shell` Settings 面板**（P0 子功能） —— 「删除我的账号」UI + 「导出即将上线」占位
- **F1 `observability-stack`** —— privacy 指标接入
- **E4 `release-gate-and-rollout`** —— W4 gate 校验 P0 删除链路 SLA 与 audit 完整性
- **`engineering-roadmap/spec.md`** —— §6 C-6 验收行的「Q-5 ADR 决定的 P0 隐私范围」具体化为「P0 仅删除」，并显式记录导出延后是 W0 产品验收例外

## 5 失效与修订条件

触发推翻或升级本 ADR 的具体阈值：

- 出现 ≥1 例 EU 用户基于 GDPR Art. 20 的正式 portability 请求 → 评估提前升格 export
- 监管环境变更（如所在区域强制双向）→ 升级 export 到 P0
- 用户调研显示 ≥ 30% 受访者把「数据可移植」作为关键决策因素 → 升格 export
- 删除 SLA 24h 在生产无法稳定达到 → 重新评估 worker 拓扑与跨表 dependency

修订流程：本 ADR 状态由 `accepted` → `superseded`，新 ADR 显式标注 `supersedes: ADR-Q5-privacy-cadence.md`；同步 §3.2 表与 §5.6 中 C12 / F4 阶段标记。

## 6 关联

- `engineering-roadmap/spec.md` §3.2 Q-5、§5.3 C12、§5.6 F4、§6 C-6
- `engineering-roadmap/plans/001-decompose-subspecs/plan.md` Phase 1.1、Phase 6.8
- 上游：`easyinterview-spec-v1-0.md` §伦理、`easyinterview-tech-docs/00-shared-conventions.md` §「隐私请求」、`02-api-definition.md` §「privacy」、`03-db-definition.md` §「privacy_requests」、`04-metrics-observability.md` §「Privacy Completion」
- 下游 child：C12 / F4 / B2 / B4 / C8 / D1 / D6 / F1 / E4

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-04-29 | 1.2 | 对齐 B4 `db-migrations-baseline` v1.4：移除旧「29 表」背景口径，改为引用 B4 baseline 多表范围；删除范围中的 `resumes` 改为当前表名 `resume_assets`，并纳入 ADR-Q1 指派给 B4 的 `external_identities` 支撑表。 | db-migrations-baseline plan-review remediation |
