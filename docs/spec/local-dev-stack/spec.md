# Local Dev Stack Spec

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-04-27

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-layer-a--foundation5-份全部-p0) 把 A2 `local-dev-stack` 列为 Layer A · Foundation 的第二份 child（依赖 [A1 `repo-scaffold`](../repo-scaffold/spec.md)）。它是 Wave 1 的 9 份契约/基础设施 spec 之一，承接 A1 在 W0 占位的 `make dev-up` / `make dev-down` target，并由 [001-decompose-subspecs Phase 3.6](../engineering-roadmap/plans/001-decompose-subspecs/checklist.md#phase-3-wave-1基础设施--契约骨架) 锁定为 W1 spec-contract lock：parent phase 只固定本地开发栈的服务清单、Make target 行为与健康检查口径；真实「克隆仓库 → 一条命令 → 全部依赖就绪并通过健康检查」由 A2 child `001` plan 验证。

目标是：

1. **冻结依赖清单与版本**：在任何 child 接入数据库 / 缓存 / 对象存储 / 观测组件之前，先把本地依赖的服务名、版本、端口、卷与依赖顺序定下来，避免 W2 多个 child 各自在 `docker-compose.yaml` 里增删服务雪球。
2. **统一 dev-up 契约**：`make dev-up` 必须 idempotent（重复执行不破坏已有数据卷）、可中断重启、退出码反映健康检查结果；任何 child 都不得在自己的 plan 里另起 `dev-up-X` target。
3. **健康检查可机器读**：`make dev-up` 的健康检查不能仅依赖 docker `healthy` 状态，必须额外暴露一个 `make dev-doctor` 或等价 target，对每个服务返回 `OK / DEGRADED / DOWN` 的结构化结果，便于 A5 CI 与 E4 release-gate 后续接入。
4. **观测栈本地化**：Prometheus / Grafana / Loki / OTel Collector 在本地必须能被 backend / worker 进程作为 endpoint 接入，避免 F1 在 W2 才发现本地与生产 endpoint 不一致。

本 spec 不实现业务代码、不接入 K8s、不部署 staging（K8s/Kind 由 [test/scenarios/](../../../test/scenarios/) 与 ADR-Q4 锁定路径承接，[E4 `release-gate-and-rollout`](../engineering-roadmap/spec.md#55-layer-e--integration4-份) 在 W4/W5 验证）。

## 2 范围

### 2.1 In Scope

- 依赖清单：`PostgreSQL 16 + pgvector`、`Redis 7`、`MinIO`、`OTel Collector`、`Grafana`、`Loki`、`Prometheus`（导出 OTel 指标）共 7 个服务的本地版本、端口、卷、network alias。
- 顶层入口：`docker-compose.yaml`（落点 `deploy/dev-stack/docker-compose.yaml`）+ A1 已占位的 `make dev-up` / `make dev-down` 真实实现。
- `make dev-doctor`：结构化健康检查，对每个服务返回 `OK / DEGRADED / DOWN` 与人类可读原因（输出 JSON + 退出码）。
- 初始化脚本：Postgres 启用 `pgvector` 扩展；MinIO 创建默认 bucket；Loki / OTel / Grafana datasource provisioning。
- `.env` 与 `config.yaml` 的最小 dev override（连接串、bucket 名、端口）；具体 secrets 抽象与 feature flag 由 [A4 `secrets-and-config`](../engineering-roadmap/spec.md#51-layer-a--foundation5-份全部-p0) 承接，本 spec 只锁 dev 默认值与字段名。
- 数据卷管理：默认命名 `easyinterview-pg-data` / `easyinterview-minio-data` / `easyinterview-loki-data`；提供 `make dev-reset` 用于显式清空（非默认）。
- 文档：`deploy/dev-stack/README.md` 一屏说明 + 故障排查 + 与 K8s/Kind 路径的 cross-link。

### 2.2 Out of Scope

- 自托管 PostHog：归 [F2 `analytics-funnel`](../engineering-roadmap/spec.md#56-layer-f--quality-横切4-份)；ADR-Q3 锁定不依赖 PostHog Cloud，但本地 dev-up 不强制启动 PostHog（资源占用大）。F2 可在自己的 plan 里另起 `make dev-up-analytics` profile，本 spec 锁定 profile 命名空间但不实现。
- AI Gateway 本地 mock：归 [A3 `ai-gateway-and-model-routing`](../engineering-roadmap/spec.md#51-layer-a--foundation5-份全部-p0)；本 spec 只在 `docker-compose.yaml` 中预留 `ai-gateway-mock` 服务名，不落地。
- K8s / Kind 场景集群：归 `test/scenarios/` 与 [E4](../engineering-roadmap/spec.md#55-layer-e--integration4-份)；ADR-Q4 锁定 staging/dev-cluster 路径走 Kind，与本 spec 的 docker-compose 互不依赖（本地双轨：docker-compose for app dev，Kind for scenario tests）。
- 生产 OTel / Grafana / Loki 部署：归 [F1 `observability-stack`](../engineering-roadmap/spec.md#56-layer-f--quality-横切4-份) 与 deploy chart；本 spec 只交付本地版本。
- 数据库迁移：归 [B4 `db-migrations-baseline`](../engineering-roadmap/spec.md#52-layer-b--contract4-份全部-p0)；A2 仅保证 Postgres 实例可用且 `pgvector` 扩展启用。
- 业务种子数据：归各 C 域 mock-server plan；本 spec 只提供空实例。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | docker-compose 落点 | `deploy/dev-stack/docker-compose.yaml`（A1 已锁定 `deploy/` 根容器） | 任何 child 不得在仓库根另起平行 compose 文件 |
| D-2 | 服务镜像 tag | 锁定 `postgres:16` / `pgvector/pgvector:pg16` / `redis:7-alpine` / `minio/minio:RELEASE.2025-09-01T...` / `otel/opentelemetry-collector-contrib:0.110.0` / `grafana/grafana:11.3.0` / `grafana/loki:3.2.0` / `prom/prometheus:v2.55.0`；具体版本号在 001-bootstrap plan 落地时按当前最新稳定版回填，但 major 不漂移 | 升级须递增 spec 版本 |
| D-3 | 服务端口 | Postgres 5432 / Redis 6379 / MinIO 9000(API) + 9001(Console) / OTel 4317(gRPC) + 4318(HTTP) / Grafana 3000 / Loki 3100 / Prometheus 9090 | 本地 5xx 段保留给 backend/worker 自身；前端 dev server 5173 不冲突 |
| D-4 | network 命名 | `easyinterview-dev`（bridge 模式）；所有服务通过短名互访 | backend/worker 启动时 `host=postgres-dev` 等命名解析 |
| D-5 | Postgres 扩展启用 | `pgvector`（必须）；后续 `pg_trgm` / `pg_stat_statements` 由 B4 决定是否前置；本 spec 只保证 `pgvector` | B4 / C11 不再需要在自己的 plan 里启用扩展 |
| D-6 | dev-up 健康检查口径 | `make dev-doctor` 返回 JSON：`{services:[{name,status:OK\|DEGRADED\|DOWN,reason}], summary:{ok,degraded,down,total}}`；`make dev-up` 在所有服务 OK 后才 exit 0 | A5 CI 与 E4 release-gate 可直接消费此输出 |
| D-7 | 数据持久化默认 | 命名卷（非 bind mount）：`easyinterview-pg-data` / `easyinterview-minio-data` / `easyinterview-loki-data` / `easyinterview-grafana-data`；`make dev-down` 不删卷，`make dev-reset` 才删 | 避免误操作丢失本地开发数据 |

### 3.2 待确认事项

- 是否在本地 dev stack 中加入 Sentry self-hosted：默认不加（占用过大），生产由 F1 负责；如团队需要本地异常聚合，可在后续 spec 修订中加入。
- Prometheus 是否独立保留 vs 仅靠 OTel Collector exporter：默认双轨（OTel Collector → Prometheus + Loki + Tempo（未来））；本 spec 暂不接 Tempo，由 F1 在 W2/W3 决定 trace backend。
- `make dev-up` 是否自动跑 `make migrate`：默认不自动；B4 落地后由开发者显式执行，避免数据库 schema 变更与服务启动耦合。

## 4 设计约束

### 4.1 结构约束

- A1 锁定的 `deploy/` 根容器之内，本 spec 只负责 `deploy/dev-stack/`；K8s manifests / Helm chart 由 E4 与 `test/scenarios/` 各自承接的子目录持有。
- `docker-compose.yaml` 必须使用 compose v2 schema（不显式声明 `version:` 字段）；profiles 用于按需启用（如 `--profile analytics` 触发 PostHog，由 F2 后续 spec 增量贡献）。

### 4.2 健康检查约束

- 每个服务必须配置容器级 `healthcheck`（compose `healthcheck:` 字段），间隔 ≤ 5s，重试 ≥ 3。
- `make dev-doctor` 在容器健康基础上，对 Postgres / Redis / MinIO 必须执行端到端 probe（连接 + 读写最小操作 + 拆解延迟），不能只看容器状态；OTel Collector / Grafana / Loki / Prometheus 至少 HTTP 200。
- 健康检查超时上限 60s（`dev-up` 默认等待时长），超出后输出每个 DOWN 服务的最近日志尾段（≤ 50 行）。

### 4.3 性能与资源约束

- 全部服务并行启动后，2 vCPU + 8GB RAM 的开发机内必须可在 60s 内全部 healthy。
- Postgres / MinIO 默认资源不设硬限制；Grafana / Loki 默认 mem limit 512Mi 各。
- 镜像总下载体积控制在 < 3GB；任何 child 想引入 > 500MB 的新镜像必须在本 spec 修订流程中登记。

### 4.4 文档约束

- `deploy/dev-stack/README.md` 必须包含：服务表（name/port/credentials default）、`make dev-*` 命令清单、常见故障（端口占用 / 卷不可写 / 镜像拉取失败）应对、与 `test/scenarios/` 的 K8s/Kind 路径区别说明。
- 本 spec 修订（新增服务 / 端口变更 / 镜像 major 升级）必须递增版本 + history 记录。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| docker-compose 文件与 Make target | A2 | `deploy/dev-stack/` 全部内容、A1 占位 target 的真实实现 |
| Postgres 扩展启用 | A2 | `pgvector` 由 init script 启用；B4 不再重复 |
| DB schema migration | B4 | A2 提供空实例 + 扩展，schema 由 B4 落地 |
| AI Gateway mock | A3 | A2 仅在 compose 预留服务名占位 |
| 自托管 PostHog | F2 | A2 锁定 profile 命名空间（`--profile analytics`），实现归 F2 |
| 生产观测部署 | F1 | A2 仅交付本地版本 |
| Secrets / config 抽象 | A4 | A2 仅锁 dev 默认值 |
| K8s / Kind 场景 | `test/scenarios/` + E4 | 双轨独立 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 一键拉起 | 干净 worktree（无既存容器与卷），仓库根已有 A1 `Makefile` 占位 `dev-up` | `make dev-up` | 7 个服务（Postgres / Redis / MinIO / OTel / Grafana / Loki / Prometheus）全部启动；`make dev-doctor` 输出 `summary.ok==7`；exit 0 | 001（A2 自身后续 plan） |
| C-2 | 失败可观察 | Postgres 端口 5432 已被占用 | `make dev-up` | 退出码非 0；stderr 输出冲突服务名 + 占用进程提示；其它服务允许已启动；`make dev-doctor` 输出对应服务 `status=DOWN,reason="port conflict"` | 001（A2 自身后续 plan） |
| C-3 | idempotent | 已运行 `make dev-up` 一次 | 再次执行 `make dev-up` | 已 healthy 服务保持运行不重启；输出说明 `already healthy`；exit 0；数据卷不被清空 | 001 |
| C-4 | 安全停止 | 服务正在运行 | `make dev-down` | 容器停止；命名卷保留；下一次 `make dev-up` 数据完整可读 | 001 |
| C-5 | 显式清空 | 服务正在运行 | `make dev-reset` | 容器停止 + 命名卷删除；操作前提示交互确认（CI 模式跳过通过 `DEV_RESET_FORCE=1`） | 001 |
| C-6 | pgvector 可用 | `make dev-up` 完成 | 在 Postgres 中执行 `select extname from pg_extension where extname='vector'` | 返回一行，确认扩展已启用 | 001 |
| C-7 | OTel endpoint 可达 | `make dev-up` 完成 | backend dev 进程指向 OTLP `http://localhost:4318` 发送 trace + metric | Grafana / Prometheus 中能查询到 `service.name=api`；Loki 中能查询到 access log；`make dev-doctor` 报 OK | 001 |
| C-8 | A2 executable gate handoff（来自 [001 Phase 3.6](../engineering-roadmap/plans/001-decompose-subspecs/checklist.md#phase-3-wave-1基础设施--契约骨架)） | 本 spec 的 contract lock 已完成，A2 后续 `001` plan 完成 | C-1 + C-7 都成立 | A2 的 `make dev-up` 可执行 gate 通过；依赖本地栈的 W2 implementation 可启动；parent Phase 3 只记录 spec-contract lock，不单独冒充本项已通过 | A2 后续 `001` |

## 7 关联计划

A2 在本次 W1 spec 阶段不创建 impl plan（参见 [001-decompose-subspecs §3.1](../engineering-roadmap/plans/001-decompose-subspecs/plan.md#3-实施步骤)）。后续由 A2 自己的 `001-bootstrap`（W1 末或 W2 初）承接：

- 落地 `deploy/dev-stack/docker-compose.yaml` 与 init scripts。
- 实现 `make dev-up` / `make dev-down` / `make dev-doctor` / `make dev-reset` 的真实命令体（替换 A1 占位）。
- 提供 `deploy/dev-stack/README.md` 与故障排查。

A2 后续如需扩展（新增观测组件、加入 PostHog profile），在原 spec 上递增版本，原地修订；不创建 sibling spec。
