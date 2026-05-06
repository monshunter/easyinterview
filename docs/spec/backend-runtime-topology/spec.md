# Backend Runtime Topology Spec

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-05-07

## 1 背景与目标

`backend-runtime-topology` 固定 P0 阶段的应用运行拓扑：当前项目尚无真实用户，运行单元应收敛为 `frontend` 与 `backend`，不再维护独立 `worker` 进程。后台任务、outbox drain、邮件派发、privacy_delete 等长任务语义仍保留，但作为 backend 进程内的后台 runner / goroutine / DB-backed executor 逐步落地。

本 subject 的目标是减少研发期进程、配置和观测依赖复杂度，同时保留 B3 job/outbox 契约、F1 metrics/logging 红线和未来生产拆分的可选扩展点。

## 2 范围

### 2.1 In Scope

- P0 runtime topology：`frontend` + `backend` 两个应用进程。
- 删除独立 `cmd/worker` 入口、`WORKER_LISTEN_ADDR`、`worker.listenAddr` 及相关 lint / dev-stack / config 文档口径。
- 将 B3 event envelope producer 中独立进程语义的 `worker` 改为 backend 内部执行方语义，例如 `backend_async`。
- 保留 `job_type` / outbox / retry / redaction / idempotency 契约，后续由 backend internal runner 消费。
- 开发期观测边界：应用生产 metrics/logs，单测使用内存 registry，dev 只依赖 `/metrics` 与容器日志；Prometheus / Grafana / OTel Collector / Loki 只作为生产或显式可选 profile。
- 原地修订受影响 owner spec / plan / ADR 口径，避免后续实现继续把 worker 当作默认前置。

### 2.2 Out of Scope

- 不实现完整 backend internal async runner、Asynq server 或 privacy_delete 删除矩阵执行。
- 不删除 B3 `job_type`、`async_jobs`、`outbox_events` 或 API-facing `JobType` 契约。
- 不新增 Prometheus / Grafana / OTel Collector / Loki 本地默认服务。
- 不设计生产期重新拆出独立 worker 的部署方案；若未来真实负载需要，必须新增 superseding ADR。

## 3 用户决策 / 待确认事项

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | P0 进程拓扑 | `frontend` + `backend`；无独立 `worker` 进程 | 删除 `cmd/worker` 与 worker listen addr；backend 内部承接后台任务 |
| D-2 | 后台任务语义 | 保留 job/outbox contract，执行方为 backend internal runner | B3/B4 契约继续可用，但不要求单独部署 worker |
| D-3 | Event producer 命名 | `backend_async` 表示 backend 内部后台执行方 | 避免 envelope producer 暗示独立进程 |
| D-4 | 开发期观测 | 应用只生产 metrics/logs；消费端非 dev gate | Prometheus/Grafana/OTel/Loki 不阻塞 TDD、BDD 或本地研发 |

## 4 设计约束

- `WORKER_LISTEN_ADDR` 与 `worker.listenAddr` 不得出现在 active config truth source、`.env.example`、validator、lint allowlist 或 dev-stack 默认配置中。
- `backend/cmd/worker` 不得作为可构建 binary 保留；`go build ./cmd/...` 只能构建当前真实入口。
- Event producer 只能使用 `api` / `backend_async` / `dispatcher` / `review`；若未来新增 producer，先修订 B3 与本 spec。
- Active owner plan/checklist、runtime code comments、tooling scripts、config/deploy docs、generated contract handoff 与 future subject shorthand 不得继续把 retired standalone worker process 或旧 `backend-async-runtime` 名称当作当前执行口径；`make lint-runtime-topology` 作为负向 gate。
- `job_type`、Asynq dotted task name、outbox retry columns 可以继续存在，因为它们是任务契约，不等于独立进程拓扑。
- F1 metrics helper 和业务埋点不得依赖 Prometheus 实例可用；测试使用内存 registry 或文本 scrape 断言。
- 默认 local dev stack 不启动观测消费端；可选观测 profile 必须显式 opt-in，且不得成为 `make dev-up` 默认依赖。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| Runtime topology | `backend-runtime-topology` | P0 进程拓扑、worker 收敛、开发期观测依赖边界 |
| Config / secrets | A4 `secrets-and-config` | 删除 worker env/config，并保持 env dictionary / validator / lint 同步 |
| Events / jobs | B3 `event-and-outbox-contract` | producer enum、event schemas、baseline、Go/TS generated artifacts |
| Backend auth | C1 `backend-auth` | 继续使用 backend-internal mail dispatcher，不等待独立 worker |
| Observability | F1 `observability-stack` | metric/log/trace 命名与红线；消费端只在生产或可选 profile |
| Local dev stack | A2 `local-dev-stack` | 默认依赖仍为 Postgres / Redis / MinIO，不新增 worker 或观测消费端 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Worker 进程取消 | 当前仓库存在 `cmd/worker` 与 worker config 口径 | 执行本计划 | `cmd/worker`、`WORKER_LISTEN_ADDR`、`worker.listenAddr` 从 active runtime/code/config 中移除 | 001-worker-consolidation |
| C-2 | 后台任务契约保留 | B3 jobs/outbox contract 已存在 | 重命名 producer 并重生成 artifacts | `job_type` / outbox / retry / redaction 契约保持，producer 使用 `backend_async` | 001-worker-consolidation |
| C-3 | 开发期观测不阻塞 | 应用产生 metrics/logs | 运行本地与测试 gate | 无 gate 需要 Prometheus/Grafana/OTel/Loki 实例；内存 registry / `/metrics` 文本断言可通过 | 001-worker-consolidation |
| C-4 | 文档口径一致 | active specs / plans / ADRs 存在旧 worker 进程说法 | 执行负向搜索 | active truth source 不再把独立 worker 作为 P0 默认前置 | 001-worker-consolidation |
| C-5 | 旧 worker 口径回流被拦截 | active code/doc handoff、tooling scripts、generated contract 或 shared event truth source 中误写 retired standalone worker process 术语、`producer: worker` / `"producer": "worker"`、旧 listen addr 口径或 `backend-async-runtime` shorthand | 执行 `make lint-runtime-topology` / `make lint` | lint 失败并定位文件行号；history、tests、本 lint 自身与本 owner 负向断言仍允许保留历史证据 | 001-worker-consolidation |

## 7 关联计划

- [001-worker-consolidation](./plans/001-worker-consolidation/plan.md)
