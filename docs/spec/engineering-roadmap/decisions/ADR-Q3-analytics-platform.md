# ADR-Q3 · 分析平台

> **版本**: 1.7
> **状态**: accepted
> **更新日期**: 2026-05-04

## 1 背景

`easyinterview-tech-docs/01-technical-architecture.md` §2 把「PostHog / Segment（产品分析）」并列为外部依赖；`04-metrics-observability.md` §1 把「产品分析」分类设为 `PostHog / Segment / Warehouse`，§4 曾定义历史产品分析事件集 + 3 个 P0 漏斗。当前产品漏斗和事件范围以 `docs/spec/product-scope/spec.md` 与 F2 `analytics-funnel` 后续 contract 为准：导入→规划→练习→报告→复练当前轮 / 下一轮→真实复盘；旧「错题」漏斗只保留为报告内题目回顾信号，不形成独立模块。`README.md` §「待评审的 5 个决策点」第 3 项只作为历史决策输入。

业务背景：

- M5 Growth Signals 强依赖事件流（funnel / cohort / retention），但只嵌入报告、画像和面试规划，不生成独立成长中心
- 漏斗对账是 release-gate 的必备项（`release-gate-and-rollout`）
- 隐私红线（Q-5 / 产品 spec §伦理）要求事件不携带原始邮箱 / 简历正文 / 面试录音，只用平台匿名 ID
- 团队规模 P0 阶段 ≤ 数人，没有专职 data engineer

P0 阶段采用与 product-scope / 当前 UI scope 对齐的产品分析事件集 + 3 漏斗，事件量级预期 ≤ 100k/day，无 ETL 重塑需求；具体事件名和 props 由 F2 原地锁定，不继承已移除模块的旧事件。

## 2 选项与取舍

### 选项 A · PostHog Cloud 直连（默认 EU region）

**Pros**：

- SaaS 一键起，无需自建 ingest
- 内置 funnel / cohort / retention / session replay（P0 不开启）/ feature flags
- 前后端双 SDK 成熟（@posthog/posthog-js / posthog-go），双发去重容易
- EU region 满足默认隐私偏好；后续可自托管 PostHog 切换无需改代码（同一 SDK）
- 含 feature flags；灰度字段以 A4/F2 当前 contract 为准，`04-metrics-observability.md` §「灰度发布」只作历史 seed
- 不引入 warehouse 维护成本

**Cons**：

- 数据所有权在 PostHog（自托管可缓解）
- 复杂 cohort / 跨表 join 仍需自托管 ClickHouse 或导出
- 高 MAU 下成本可见

### 选项 B · Segment（CDP）+ 自托管 Warehouse（BigQuery / Snowflake）

**Pros**：

- 一次埋点多平台分发（PostHog / Mixpanel / Amplitude / Warehouse 等）
- Warehouse 中可任意 SQL 分析，data engineer 友好

**Cons**：

- 需立即决定 Warehouse 选型 + dbt 模型 + BI 工具，运维 / 学习成本立刻拉满
- Segment 本身按 MTU 计费，PostHog 又算一份 → 双重成本
- P0 团队没有 data engineer，复杂栈成 ROI 负

### 选项 C · 自托管 PostHog（不依赖 PostHog Cloud）

**Pros**：

- 完全数据自主
- 不依赖第三方产品分析 SaaS，符合「敏感求职数据默认保守」方向
- SDK / 事件模型仍沿用 PostHog，F2 adapter 不需要重写产品分析语义
- feature flags 与 funnel 能留在同一工具内，避免 Segment + Warehouse 的多系统成本

**Cons**：

- 自托管 PostHog 需额外存储 / 队列 / ClickHouse 等组件，备份、升级、容量规划由团队自担
- PostHog Kubernetes / Helm 路径已不再是官方支持主线（见 PostHog `charts-clickhouse` README）；不得把当前 gate 建在已废弃 chart 假设上
- P0 运维复杂度高于 Cloud，F2 / E4 必须先验证可运维 self-host path；A2 默认本地开发栈不承接 PostHog 部署

### 选项 D · 不接产品分析（仅靠 Prometheus 业务指标）

**Pros**：

- 零额外工具

**Cons**：

- 漏斗 / cohort / retention 全部要自建 SQL，等于自做一个 BI；与 release-gate 漏斗对账直接冲突
- M5 Growth Signals 失去基础数据

## 3 决策

**P0 锁定选项 C：自托管 PostHog 作为唯一产品分析后端；不依赖 PostHog Cloud / Segment / Warehouse。**

落地约束：

1. **事件契约**：以 product-scope / 当前 UI scope 对齐后的 F2 产品分析事件集为唯一产品分析来源；若需要复用内部 envelope / trace 语义，必须引用 B3 `event-and-outbox-contract` 当前 contract，不再直接以 `06-event-contracts.md` 为准。前后端双发由 F2 `analytics-funnel` 提供 `idempotency_key` 去重
2. **匿名标识**：使用 PostHog 匿名 `distinct_id`；登录后通过 `identify` 关联 `users.public_id`（不传 email / 真实姓名）；登出后 reset
3. **adapter 抽象**：F2 `analytics-funnel` 提供 `track(event, props)` / `identify` / `featureFlag` 接口，业务代码不直接 import PostHog SDK；如果 self-host path 被后续 ADR 推翻，切换厂商只改 adapter
4. **opt-out**：`user_settings.analytics_opt_in = false` 时全部 SDK 不初始化（前端）+ 后端事件标 `dnt=true` 进 dead-letter
5. **feature flags**：自托管 PostHog feature flags 是唯一灰度源；具体 flag 字典由 A4/F2 当前 contract 决定，`01-technical-architecture.md` §「ai_fallback_model_enabled」只作历史 seed
6. **环境隔离**：dev / staging / prod 各自独立 project；普通本地 dev 默认使用 no-op / file-backed mode，不启动 PostHog；staging 必须跑与 prod 同等的 self-host 验证路径
7. **部署边界**：本 ADR 只锁定 self-host，不锁死 Helm / K8s chart；F2 / E4 必须在对应 child spec 中验证当前官方支持的部署路径、备份、升级、恢复、容量与告警。若唯一可行路径无法满足 Q4/E4 gate，必须新 ADR 推翻本决策
8. **rate limit**：客户端 SDK 默认开启 batch；服务端事件经 C8 outbox 异步发送，避免阻塞主请求；`analytics_dispatch` 若落入 `async_jobs`，必须先由 B3 / B4 additive 加入 public `jobType` 契约

## 4 影响范围

- **F2 `analytics-funnel`** —— 落地 adapter + 与 product-scope / 当前 UI scope 对齐的产品分析事件集 + 3 漏斗 + feature flag 包装层，并验证 self-host PostHog 的 funnel / flag 能力
- **D1 `frontend-shell`** —— 注入 `<PosthogProvider>`，按 `analytics_opt_in` 条件初始化
- **C8 `backend-async-runtime`** —— `analytics_dispatch` job_type 异步推送 server-side 事件；新增前必须由 B3 / B4 更新 public job 契约
- **A2 `local-dev-stack`** —— 默认 `make dev-up` 不启动 PostHog；只需确保 no-op / file-backed dev 配置不阻塞本地启动
- **A4 `secrets-and-config`** —— `POSTHOG_PROJECT_API_KEY` / `POSTHOG_HOST` / `POSTHOG_SELF_HOSTED` / backup credentials 配置项
- **F1 `observability-stack`** —— 自托管 PostHog 可用性、ingestion lag、存储水位、备份成功率进入 dashboard
- **privacy export / advanced audit future candidate** —— `analytics_opt_in` 切换写 audit_event；删除链路触发 PostHog `delete person`
- **`release-gate-and-rollout`** —— 创建时将漏斗对账 gate 引用自托管 PostHog funnel 报表，并验证备份 / 恢复 / 升级 runbook
- **B1 `shared-conventions-codified`** —— F2 锁定后的事件名 / props schema 作为 TS / Go 共享枚举
- **当前 P0 backend / frontend workstream 与 future candidates** —— 通过 F2 接口埋点，不引入新 vendor；readiness / trends 只能作为嵌入式信号进入报告、画像或面试规划，不恢复独立 growth overview

## 5 失效与修订条件

触发推翻或升级本 ADR 的具体阈值：

- 自托管 PostHog 升级 / 备份 / 恢复演练连续 2 次失败，或 ingestion lag p95 > 5min 持续 1 周 → 优先评估更轻量自托管分析栈；如需切回 Cloud 必须新 ADR + 用户批准
- 自托管基础设施成本 + 运维人力超过等价 Cloud 方案预估成本 2 倍且没有合规收益 → 重新评估自托管方案边界；如需切 Cloud 必须新 ADR + 用户批准
- 出现 ≥3 个目标平台（如同时需要推送到 Hubspot / Salesforce）→ 引入 Segment / RudderStack 作为 CDP 层
- 漏斗 SQL 复杂度无法在 PostHog 内表达 → 引入 Warehouse + dbt
- 自托管 PostHog feature flag 延迟或故障影响灰度发布 ≥ 2 次 / 季 → 评估独立 feature flag 服务

修订流程：本 ADR 状态由 `accepted` → `superseded`，新 ADR 显式标注 `supersedes: ADR-Q3-analytics-platform.md`。如果 F2 / E4 发现 self-host path 不可运维，不得静默降级到 Cloud，必须走新 ADR。

## 6 关联

- `engineering-roadmap/spec.md` §3.2 Q-3
- `engineering-roadmap/plans/001-decompose-subspecs/plan.md` checklist 1.1
- 历史输入：`easyinterview-tech-docs/01-technical-architecture.md` §2、`04-metrics-observability.md` §1 §4 §15、`06-event-contracts.md`
- 当前契约：`docs/spec/product-scope/spec.md`、F2 `analytics-funnel` 后续 spec、F1 `observability-stack`、B3 `event-and-outbox-contract`
- 参考：PostHog `charts-clickhouse` README（Kubernetes support sunset）
- 下游 child / future candidate：F2 / frontend-shell / backend-async-runtime / A4 / release-gate-and-rollout / B1 / F1 / A2 / advanced audit

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-05-04 | 1.7 | 对齐 engineering-roadmap v3.0：删除当前影响范围中的旧 C10-C12 / D7 / F4 / E4 编号口径，改为当前 P0 workstream 与 future candidate 表述。 |
| 2026-05-03 | 1.6 | 将 `06-event-contracts.md` / `04-metrics-observability.md` / `01-technical-architecture.md` 在分析 ADR 中降级为历史输入；当前产品分析事件、灰度和 envelope 语义分别以 F2、A4/F2 与 B3 active contract 为准。 |
| 2026-05-03 | 1.5 | 移除 ADR 当前段落中的旧产品分析事件数量硬编码，明确 F2 必须按 product-scope / 当前 UI scope 重新锁定事件集，不继承已移除模块事件。 |
| 2026-05-03 | 1.4 | 对齐 product-scope v1.1 与 engineering-roadmap v2.2：P0 漏斗改为导入→规划→练习→报告→复练 / 下一轮→真实复盘；M5 改为嵌入式 Growth Signals，不恢复独立成长中心或错题漏斗。 |
| 2026-04-29 | 1.3 | 对齐 engineering-roadmap v2.0：C9 `backend-debrief` 已升格为 P0 真实面试复现 / 复盘文本流，P1 child 范围收窄为 C10-C12 / D5-D7。 |
| 2026-04-27 | 1.2 | 对齐 A2 local-dev-stack v1.2：普通本地 `make dev-up` 不启动 PostHog；自托管 PostHog 部署验证归 F2/E4，A2 只保证 no-op / file-backed dev 模式不阻塞本地启动。 |
