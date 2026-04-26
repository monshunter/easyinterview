# ADR-Q3 · 分析平台

> **版本**: 1.0
> **状态**: accepted
> **更新日期**: 2026-04-26

## 1 背景

`easyinterview-tech-docs/01-technical-architecture.md` §2 把「PostHog / Segment（产品分析）」并列为外部依赖；`04-metrics-observability.md` §1 把「产品分析」分类设为 `PostHog / Segment / Warehouse`，§4 定义了 18 个产品分析事件 + 3 个 P0 漏斗（导入→工作台→报告→错题→复练）。`README.md` §「待评审的 5 个决策点」第 3 项把分析平台选择留作 W0 决策。

业务背景：

- M5 成长系统强依赖事件流（funnel / cohort / retention）
- 漏斗对账是 W4 release-gate 的必备项（E4 release-gate-and-rollout）
- 隐私红线（Q-5 / 产品 spec §伦理）要求事件不携带原始邮箱 / 简历正文 / 面试录音，只用平台匿名 ID
- 团队规模 P0 阶段 ≤ 数人，没有专职 data engineer

P0 阶段 18 事件 / 3 漏斗，事件量级预期 ≤ 100k/day，无 ETL 重塑需求。

## 2 选项与取舍

### 选项 A · PostHog Cloud 直连（默认 EU region）

**Pros**：

- SaaS 一键起，无需自建 ingest
- 内置 funnel / cohort / retention / session replay（P0 不开启）/ feature flags
- 前后端双 SDK 成熟（@posthog/posthog-js / posthog-go），双发去重容易
- EU region 满足默认隐私偏好；后续可自托管 PostHog 切换无需改代码（同一 SDK）
- 含 feature flags，与 `04-metrics-observability.md` §「灰度发布」flag 列表天然联动
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

### 选项 C · 自托管 PostHog（Helm Chart）

**Pros**：

- 完全数据自主
- 与 K8s 部署（Q-4）天然贴合

**Cons**：

- 自托管 PostHog 需 ClickHouse + Kafka + Postgres + Redis + MinIO 一整套，运维负担与产品本身相当
- P0 不应承担

### 选项 D · 不接产品分析（仅靠 Prometheus 业务指标）

**Pros**：

- 零额外工具

**Cons**：

- 漏斗 / cohort / retention 全部要自建 SQL，等于自做一个 BI；与 W4 漏斗对账 gate 直接冲突
- M5 成长系统失去基础数据

## 3 决策

**P0 锁定选项 A：PostHog Cloud（EU region）作为唯一产品分析后端。**

落地约束：

1. **事件契约**：以 `04-metrics-observability.md` §4 的 18 事件 + `06-event-contracts.md` 的 envelope 为唯一来源；前后端双发由 F2 `analytics-funnel` 提供 `idempotency_key` 去重
2. **匿名标识**：使用 PostHog 匿名 `distinct_id`；登录后通过 `identify` 关联 `users.public_id`（不传 email / 真实姓名）；登出后 reset
3. **adapter 抽象**：F2 `analytics-funnel` 提供 `track(event, props)` / `identify` / `featureFlag` 接口，业务代码不直接 import PostHog SDK；切换厂商只改 adapter
4. **opt-out**：`user_settings.analytics_opt_in = false` 时全部 SDK 不初始化（前端）+ 后端事件标 `dnt=true` 进 dead-letter
5. **feature flags**：PostHog feature flags 是唯一灰度源；与 `01-technical-architecture.md` §「ai_fallback_model_enabled」等 flag 落到同一 console
6. **环境隔离**：dev / staging / prod 各自独立 project，避免脏数据污染漏斗
7. **rate limit**：客户端 SDK 默认开启 batch；服务端事件经 C8 outbox 异步发送，避免阻塞主请求

## 4 影响范围

- **F2 `analytics-funnel`** —— 落地 adapter + 18 事件 + 3 漏斗 + feature flag 包装层
- **D1 `frontend-shell`** —— 注入 `<PosthogProvider>`，按 `analytics_opt_in` 条件初始化
- **C8 `backend-async-runtime`** —— `analytics_dispatch` job_type 异步推送 server-side 事件
- **A4 `secrets-and-config`** —— `POSTHOG_PROJECT_API_KEY` / `POSTHOG_HOST` / `POSTHOG_REGION` 配置项
- **F4 `privacy-and-audit-runtime`** —— `analytics_opt_in` 切换写 audit_event；删除链路触发 PostHog `delete person`
- **E4 `release-gate-and-rollout`** —— W4 漏斗对账 gate 引用 PostHog funnel 报表
- **B1 `shared-conventions-codified`** —— 18 事件名 / props schema 作为 TS / Go 共享枚举
- **C9-C12 / D5-D7 P1 child** —— 通过 F2 接口埋点，不引入新 vendor

## 5 失效与修订条件

触发推翻或升级本 ADR 的具体阈值：

- MAU > 50k 且 PostHog Cloud 月成本 > $1500 → 评估自托管 PostHog（K8s Helm）
- 数据合规要求强制本地化（中国 / 俄罗斯）→ 评估区域自托管 / 切换厂商
- 出现 ≥3 个目标平台（如同时需要推送到 Hubspot / Salesforce）→ 引入 Segment / RudderStack 作为 CDP 层
- 漏斗 SQL 复杂度无法在 PostHog 内表达 → 引入 Warehouse + dbt
- PostHog SLA 事故影响 funnel 对账 ≥ 2 次 / 季 → 评估替代

修订流程：本 ADR 状态由 `accepted` → `superseded`，新 ADR 显式标注 `supersedes: ADR-Q3-analytics-platform.md`。

## 6 关联

- `engineering-roadmap/spec.md` §3.2 Q-3
- `engineering-roadmap/plans/001-decompose-subspecs/plan.md` Phase 1.1
- 上游：`easyinterview-tech-docs/01-technical-architecture.md` §2、`04-metrics-observability.md` §1 §4 §15、`06-event-contracts.md`
- 下游 child：F2 / D1 / C8 / A4 / F4 / E4 / B1
