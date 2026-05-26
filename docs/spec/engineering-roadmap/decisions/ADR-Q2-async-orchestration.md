# ADR-Q2 · 异步编排

> **版本**: 1.6
> **状态**: accepted
> **更新日期**: 2026-05-26

## 1 背景

`engineering-roadmap` 历史上曾把后端拆成 `api` + `worker` 两个进程；`backend-runtime-topology` v1.0 已将 P0 拓扑收敛为 `frontend` + `backend`，后台任务由 backend internal runner 承接。`B1 shared-conventions-codified` §4 把所有长耗时操作统一为「异步 Job 模式」；`B4 db-migrations-baseline` 已定义 `async_jobs` 与 `outbox_events` 两张表。`README.md` §「待评审的 5 个决策点」第 2 项只作为历史决策输入。

P0 已识别的异步链路：

- JD 解析（public `jobType=target_import`，internal Asynq handler 可映射为 `target.import`）
- 模拟面试报告生成（public `jobType=report_generate`，internal Asynq handler 可映射为 `report.generate`）
- 简历定制（`resume.tailor`）
- 报告题目回顾与本轮复练上下文物化（`review.materialize_report_items`）；不生成独立错题本或 Drill 队列
- Email magic link / 通知派发（internal-only canonical `jobType=email_dispatch`，backend internal handler 可映射为 `email.dispatch`）
- Outbox event publish（backend internal runner drain / `outbox.dispatch` handler）

特征：

- 单 job 耗时秒级到分钟级，无跨日 SLA、无人工签字 / 多步补偿
- 幂等键由调用方提供（`clientEventId` / `dedupe_key`）；不需要 saga
- 队列容量预期 P0 ≤ 数千并发任务量级
- 现有技术栈已默认 Redis（session / cache / rate limit），引入额外队列 broker 成本高

## 2 选项与取舍

### 选项 A · Asynq（Go-native，Redis-backed）

**Pros**：

- 与已选 Redis 栈天然复用，不引入额外组件
- Go 一等公民，与 `chi + pgx + sqlc` 心智无缝
- 内置：retry + backoff + 优先级队列 + delayed/scheduled / unique queue + cron / web UI
- runner 可先内嵌在 backend 进程内，真实负载出现后再评估是否拆出部署单元
- 与 outbox dispatcher 同库不同表，事务边界清晰

**Cons**：

- 不天然支持 long-running workflow / signal / query / human-in-loop
- 跨日持久化依赖 Redis 持久化策略（AOF + 磁盘），需运维约束
- 复杂 saga 需自行编排（业务层组合多个 task）

### 选项 B · Temporal（独立编排服务）

**Pros**：

- workflow as code、versioning、replay、长运行历史
- 内置 signal / query / 人工干预；适合多步补偿
- 多语言 SDK，跨服务编排自然

**Cons**：

- 需独立部署 Temporal cluster + Cassandra/PostgreSQL backend，运维复杂度 +1 个数量级
- 对 P0「秒级 - 分钟级 + 幂等」场景明显过设
- 学习曲线陡峭（workflow / activity 心智模型）；与现有 Asynq DB 表设计需重新对齐
- 主仓代码无 Temporal 引入痕迹，全部从零接入

### 选项 C · 数据库轮询（DB-only）

**Pros**：

- 无新依赖

**Cons**：

- 锁竞争、长尾延迟、调度不公平；与 Asynq 相比无任何架构优势
- 失败重试 / 优先级 / 调度 / 监控全部需自建

## 3 决策

**P0 锁定选项 A 的任务契约：B3 canonical `job_type` + PG outbox / `async_jobs` + Redis/Asynq-compatible handler naming；运行形态由 backend internal runner 承接，不再要求独立 worker 进程。**

落地约束：

1. **进程拓扑**：P0 只有 `frontend` + `backend` 应用进程；backend 负责 enqueue、outbox 写入和 backend internal runner drain。不得把独立 worker 进程作为本地开发、BDD 或 P0 auth/业务闭环前置。
2. **任务命名边界**：API / DB / event / metrics 暴露或持久化的 canonical `jobType` 必须沿用 snake_case 值（例如 `target_import`、`report_generate`、`privacy_delete`、`email_dispatch`）；内部 handler 可使用 `<domain>.<action>` dotted name（例如 `target.import`、`report.generate`、`privacy.delete`、`email.dispatch`），但 backend async runner 必须维护双向映射，B3 / B4 新增 job 时必须先 additive 更新 canonical `jobType`
3. **幂等**：所有 task payload 必须含 `dedupe_key`；进 `async_jobs.dedupe_key` unique index（per `job_type`）
4. **重试策略**：默认指数退避（30s/2m/10m/1h/6h），最多 5 次；超限后落 `dead_letter` 并写 audit_event
5. **优先级队列**：`critical`（user-facing：`report_generate` / `privacy_delete`）/ `default`（`target_import` / `resume_tailor`）/ `low`（`email_dispatch` / `analytics_dispatch` / batch）；`email_dispatch` 由 B3 / B4 作为 internal-only canonical jobType 纳入契约，`analytics_dispatch` 新增前必须由 B3 / B4 additive 更新
6. **可观测性**：每个 task 必须落 `async_job_duration_seconds` + `async_jobs_processed_total{result=succ|fail|retry}` + Sentry breadcrumb
7. **outbox**：`outbox_events` 表由业务事务写入；backend internal runner 按 B3 协议 drain 未发布行并投递对应 handler。未来如拆出独立部署单元，必须由新 ADR 明确 supersede 本拓扑。
8. **Asynq Web UI**：不作为 P0 默认 dev/staging 前置；如未来采用 Asynq server，可由可选运维 profile 暴露给 ops。

## 4 影响范围

- **`backend-runtime-topology`** —— 持有 P0 无独立 worker 进程和 backend internal runner 运行形态
- **backend async runner future subject** —— 落地 backend 内部 task registry、outbox drain、retry、metrics adapter；不默认创建独立进程
- **A2 `local-dev-stack`** —— 默认启动 Postgres / Redis / MinIO / Mailpit；不包含 Asynq Web UI 或独立 worker health gate
- **B3 `event-and-outbox-contract`** —— 当前 16 个内部事件 envelope + outbox publish_status 状态机；新增事件必须走 B3 additive 更新
- **B4 `db-migrations-baseline`** —— `async_jobs` / `outbox_events` 0001 迁移；`async_jobs.job_type` check 必须包含 internal-only `email_dispatch`
- **C1 `backend-auth`** —— magic link 邮件派发通过 C1 backend-internal dispatcher / future `email_dispatch` job contract，handler 不同步等待 provider
- **C4 / C5 / C6 / C7** —— 所有异步链路统一通过 backend async runner / B3 job contract enqueue，禁止业务代码直接 import concrete queue SDK
- **F1 `observability-stack`** —— `async_jobs_*` + `outbox_publish_lag_seconds` 指标接入；queue depth alert 阈值锁定
- **A4 `secrets-and-config`** —— Redis 连接串 / Asynq 队列权重作为 config 节点

## 5 失效与修订条件

触发推翻或升级本 ADR 的具体阈值：

- 出现单工作流 ≥ 24h SLA / 跨日定时器需求（如「面试后 7 天提醒」演化成 saga）→ 评估 Temporal 或 river queue
- 出现多步补偿 / 强一致 saga 需求（如跨 4+ 域的金融级回滚）→ Temporal
- Redis AOF + 磁盘持久化在生产无法满足 RPO < 5min → 评估迁移到持久化更强的 broker（NATS JetStream / Postgres-only river）
- backend internal runner 吞吐影响 API SLO，或 task throughput ≥ 10k/s 持续 → 评估拆出独立运行单元和更强调度器
- 出现需要 workflow versioning / replay 调试的复杂业务 → Temporal

修订流程：本 ADR 状态由 `accepted` → `superseded`，新 ADR 显式标注 `supersedes: ADR-Q2-async-orchestration.md`。

## 6 关联

- `engineering-roadmap/spec.md` §3.2 Q-2
- `engineering-roadmap/plans/001-decompose-subspecs/plan.md` checklist 1.1
- 参考背景：`B1 shared-conventions-codified` §4 异步任务约定、`engineering-roadmap decisions` §3、`B4 db-migrations-baseline` §5.9、`F1 observability-stack` §「async_jobs」、`B3 event-and-outbox-contract`
- 下游 child：backend-runtime-topology / backend async runner future subject / B3 / B4 / A2 / A4 / F1；间接：C4-C7 全部

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-05-26 | 1.6 | 对齐 local-dev-stack Mailpit revision：默认依赖增加 Mailpit 本地邮箱 sink，但仍不新增独立 worker 或 Asynq Web UI 前置。 | local-dev-stack/001 Mailpit revision |
| 2026-05-06 | 1.5 | 按 backend-runtime-topology v1.0 取消 P0 独立 worker 进程前置：保留 B3 job/outbox/handler 命名契约，将运行形态改为 backend internal runner。 | backend-runtime-topology/001-worker-consolidation |
| 2026-05-03 | 1.4 | 对齐 B3 当前可执行事件契约：event/outbox 当前为 16 个内部事件，不再沿用旧 18 事件口径。 | event-and-outbox-contract / product-scope v1.2 |
| 2026-05-03 | 1.3 | 对齐 product-scope v1.1：异步 review 物化只服务报告内题目回顾与本轮复练上下文，不再承接独立 Mistake / Drill 队列。 | product-scope / engineering-roadmap v2.2 |
| 2026-04-29 | 1.2 | 将 magic link / 通知派发所需的 `email_dispatch` 明确纳入 internal-only canonical jobType，锁定 `email.dispatch` Asynq dotted name、low priority 队列与 B3/B4 契约同步要求；同时把早期示例 dotted name 对齐 B3 规范。 | plan-review remediation |
