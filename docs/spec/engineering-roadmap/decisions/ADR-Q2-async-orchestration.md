# ADR-Q2 · 异步编排

> **版本**: 1.0
> **状态**: accepted
> **更新日期**: 2026-04-26

## 1 背景

`easyinterview-tech-docs/01-technical-architecture.md` §3 把后端拆成 `api` + `worker` 两个进程；`00-shared-conventions.md` §4 把所有长耗时操作统一为「异步 Job 模式」；`03-db-definition.md` §5.9 已定义 `async_jobs` 与 `outbox_events` 两张表。`README.md` §「待评审的 5 个决策点」第 2 项把异步编排选择留作 W0 决策。

P0 已识别的异步链路：

- JD 解析（`target_job.parse`）
- 模拟面试报告生成（`practice.report_generate`）
- 简历定制（`resume.tailor`）
- Mistake / Drill 物化（`review.materialize_mistakes`）
- Email magic link / 通知派发（`email_dispatch`）
- Outbox event publish（`outbox.dispatch`）

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
- worker 拓扑简单，K8s deployment 即起即停
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

**P0 锁定选项 A：Asynq + Redis 作为唯一异步 runtime；outbox 事件由 PG outbox 表 + 独立 dispatcher 进程消费推 Asynq enqueue。**

落地约束：

1. **进程拓扑**：`api` 进程负责 enqueue + outbox 写入；`worker` 进程独占 Asynq consumer + outbox dispatcher
2. **任务命名**：`<domain>.<action>`，例如 `target_job.parse`、`practice.report_generate`，与 `04-metrics-observability.md` §「`async_jobs_*`」label 对齐
3. **幂等**：所有 task payload 必须含 `dedupe_key`；进 `async_jobs.dedupe_key` unique index（per `job_type`）
4. **重试策略**：默认指数退避（30s/2m/10m/1h/6h），最多 5 次；超限后落 `dead_letter` 并写 audit_event
5. **优先级队列**：`critical`（user-facing：report_generate）/ `default`（target_job.parse / resume.tailor）/ `low`（email_dispatch / batch）
6. **可观测性**：每个 task 必须落 `async_job_duration_seconds` + `async_jobs_processed_total{result=succ|fail|retry}` + Sentry breadcrumb
7. **outbox**：`outbox_events` 表由业务事务写入；dispatcher 独立 cron job（每 5s 扫一次未发布行）→ Asynq enqueue → publish 到事件 bus（P0 内网直发）
8. **Asynq Web UI**：本地 dev / staging 默认开启；prod 通过 ingress + auth 暴露给 ops

## 4 影响范围

- **C8 `backend-async-runtime`** —— 落地 Asynq client / server 抽象、outbox dispatcher、task registry、metrics adapter
- **A2 `local-dev-stack`** —— `docker-compose.yml` 含 Redis 7 + AOF；`make dev-up` 健康检查包含 Asynq ping
- **B3 `event-and-outbox-contract`** —— 18 事件 envelope + outbox publish_status 状态机
- **B4 `db-migrations-baseline`** —— `async_jobs` / `outbox_events` 0001 迁移
- **C4 / C5 / C6 / C7** —— 所有异步链路统一通过 C8 SDK enqueue，禁止业务代码直接 import `asynq`
- **F1 `observability-stack`** —— `async_jobs_*` + `outbox_publish_lag_seconds` 指标接入；queue depth alert 阈值锁定
- **A4 `secrets-and-config`** —— Redis 连接串 / Asynq 队列权重作为 config 节点

## 5 失效与修订条件

触发推翻或升级本 ADR 的具体阈值：

- 出现单工作流 ≥ 24h SLA / 跨日定时器需求（如「面试后 7 天提醒」演化成 saga）→ 评估 Temporal 或 river queue
- 出现多步补偿 / 强一致 saga 需求（如跨 4+ 域的金融级回滚）→ Temporal
- Redis AOF + 磁盘持久化在生产无法满足 RPO < 5min → 评估迁移到持久化更强的 broker（NATS JetStream / Postgres-only river）
- worker 节点数 ≥ 50 或 task throughput ≥ 10k/s 持续 → 评估更强调度器
- 出现需要 workflow versioning / replay 调试的复杂业务 → Temporal

修订流程：本 ADR 状态由 `accepted` → `superseded`，新 ADR 显式标注 `supersedes: ADR-Q2-async-orchestration.md`。

## 6 关联

- `engineering-roadmap/spec.md` §3.2 Q-2
- `engineering-roadmap/plans/001-decompose-subspecs/plan.md` Phase 1.1
- 上游：`easyinterview-tech-docs/00-shared-conventions.md` §4 异步任务约定、`01-technical-architecture.md` §3、`03-db-definition.md` §5.9、`04-metrics-observability.md` §「async_jobs」、`06-event-contracts.md`
- 下游 child：C8 / B3 / B4 / A2 / A4 / F1；间接：C4-C7 全部
