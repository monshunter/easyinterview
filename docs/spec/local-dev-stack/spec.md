# Local Dev Stack Spec

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-04-29

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-layer-a--foundation5-份全部-p0) 把 A2 `local-dev-stack` 列为 Layer A · Foundation 的第二份 child（依赖 [A1 `repo-scaffold`](../repo-scaffold/spec.md)）。它承接 A1 在 W0 占位的 `make dev-up` / `make dev-down` target，并由 [001-decompose-subspecs Phase 3.6](../engineering-roadmap/plans/001-decompose-subspecs/checklist.md#phase-3-wave-1基础设施--契约骨架) 锁定为 W1 spec-contract lock：parent phase 只固定本地开发栈的最小依赖、应用组件启动契约、Make target 行为与健康检查口径；真实「克隆仓库 → 一条命令 → 本项目本地环境启动完成」由 A2 child `001` plan 验证。

本地开发环境不是生产观测环境的缩小版。默认 `make dev-up` 只启动开发 P0 闭环必须的本地依赖与当前仓库内可运行的项目组件，避免把生产才需要的 OTel Collector、Grafana、Loki、Prometheus 等高占用组件变成本地开发前置条件。

目标是：

1. **冻结最小依赖清单与版本**：默认本地依赖只包含 `PostgreSQL + pgvector`、`Redis`、`MinIO`。任何 child 不得把生产观测、分析平台、AI gateway 或其它重型组件加入默认 `make dev-up`。
2. **统一本地环境启动契约**：`make dev-up` 必须通过 docker compose 启动最小依赖以及当前项目所有已具备本地运行入口的组件（例如 backend API、worker、frontend dev server）；重复执行不破坏已有数据卷。
3. **健康检查可机器读**：`make dev-doctor` 对依赖服务与项目组件返回 `OK / DEGRADED / DOWN` 的结构化结果，`make dev-up` 的退出码反映整体健康状态。
4. **本地观测轻量化**：本地只要求应用自身暴露 `/metrics`（当组件已具备 HTTP runtime 时）并通过容器日志确认行为；不安装 Grafana / Loki / Prometheus / OTel Collector 作为默认依赖。
5. **本地 AI 调用真实化**：docker compose 本地部署只通过 env 注入真实 AI provider / OpenAI-compatible endpoint，不启动 AI gateway 容器，也不把应用切到单元测试 stub。

本 spec 不实现业务代码、不接入 K8s、不部署 staging（K8s/Kind 由 `test/scenarios/`（W4 spawn，E2 owner）与 ADR-Q4 锁定路径承接，[E4 `release-gate-and-rollout`](../engineering-roadmap/spec.md#55-layer-e--integration4-份) 在 W4/W5 验证）。

## 2 范围

### 2.1 In Scope

- 最小本地依赖清单：`PostgreSQL 16 + pgvector`、`Redis 7`、`MinIO` 的本地版本、端口、卷、network alias。
- 项目组件启动：`docker-compose.yaml` 必须包含当前仓库内所有已具备本地运行入口的项目组件；后续 backend / frontend / worker 等 child 落地运行入口时，必须接入同一个 compose，而不是另起平行本地环境。
- 顶层入口：`docker-compose.yaml`（落点 `deploy/dev-stack/docker-compose.yaml`）+ A1 已占位的 `make dev-up` / `make dev-down` 真实实现。
- `make dev-doctor`：结构化健康检查，对每个依赖服务与项目组件返回 `OK / DEGRADED / DOWN` 与人类可读原因（输出 JSON + 退出码）。
- 初始化脚本：Postgres 启用 `pgvector` 扩展；MinIO 创建默认 bucket。
- `.env` 与 `config.yaml` 的最小 dev override（连接串、bucket 名、端口、应用组件默认 host/port、AI provider / gateway endpoint 与 key 占位）；具体 secrets 抽象与 feature flag 由 [A4 `secrets-and-config`](../engineering-roadmap/spec.md#51-layer-a--foundation5-份全部-p0) 承接，本 spec 只锁 dev 默认值与字段名。
- 数据卷管理：默认命名 `easyinterview-pg-data` / `easyinterview-redis-data` / `easyinterview-minio-data`；提供 `make dev-reset` 用于显式清空（非默认）。
- 文档：`deploy/dev-stack/README.md` 一屏说明 + 故障排查 + 与 K8s/Kind 路径的 cross-link。

### 2.2 Out of Scope

- OTel Collector / Grafana / Loki / Prometheus 默认本地部署：归 [F1 `observability-stack`](../engineering-roadmap/spec.md#56-layer-f--quality-横切4-份) 或生产部署路径；A2 默认 `make dev-up` 不安装这些组件，也不以它们作为健康检查前提。
- AI Gateway 本地 mock 或基础设施服务：AI gateway 不是 A2 本地依赖；A2 不启动 `ai-gateway` / `ai-gateway-mock` 容器。单元测试 stub / 离线 provider mock 归 A3；docker compose 本地部署必须把真实 AI provider / OpenAI-compatible endpoint 配置传给应用组件。
- 自托管 PostHog：归 [F2 `analytics-funnel`](../engineering-roadmap/spec.md#56-layer-f--quality-横切4-份)；ADR-Q3 锁定不依赖 PostHog Cloud，但本地 dev-up 不强制启动 PostHog（资源占用大）。
- K8s / Kind 场景集群：归 `test/scenarios/` 与 [E4](../engineering-roadmap/spec.md#55-layer-e--integration4-份)；ADR-Q4 锁定场景集成测试路径走 Kind，与本 spec 的 docker-compose 互不依赖（本地双轨：docker-compose for app dev，Kind for scenario tests）。
- 数据库迁移：归 [B4 `db-migrations-baseline`](../engineering-roadmap/spec.md#52-layer-b--contract4-份全部-p0)；A2 仅保证 Postgres 实例可用且 `pgvector` 扩展启用。
- 业务种子数据：归各 C 域 mock-server plan；本 spec 只提供空实例。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | docker-compose 落点 | `deploy/dev-stack/docker-compose.yaml`（A1 已锁定 `deploy/` 根容器） | 任何 child 不得在仓库根另起平行 compose 文件 |
| D-2 | 服务镜像 tag | 默认依赖锁定 `pgvector/pgvector:pg16` / `redis:7-alpine` / `minio/minio:RELEASE.2024-12-18T13-15-44Z`；MinIO bucket init 工具锁 `minio/mc:RELEASE.2024-11-21T17-21-54Z`。项目组件优先从仓库内 Dockerfile 或 dev command 构建，不从外部拉取业务镜像 | 升级须递增 spec 版本；默认 compose 不含生产观测镜像 |
| D-3 | 服务端口 | Postgres 5432 / Redis 6379 / MinIO 9000(API) + 9001(Console)；项目组件端口由各组件 dev defaults 声明（frontend 默认 5173，backend API 默认 8080，worker 默认不暴露 host port） | 不预留 Grafana 3000 / Prometheus 9090 / Loki 3100 / OTLP 4317/4318 给默认本地栈 |
| D-4 | network 命名 | `easyinterview-dev`（bridge 模式）；依赖服务与项目组件通过短名互访 | backend/worker 启动时 `host=postgres-dev` / `redis-dev` / `minio-dev` 等命名解析 |
| D-5 | Postgres 扩展启用 | `pgvector`（必须）；后续 `pg_trgm` / `pg_stat_statements` 由 B4 决定是否前置；本 spec 只保证 `pgvector` | B4 / C11 不再需要在自己的 plan 里启用扩展 |
| D-6 | dev-up 健康检查口径 | `make dev-doctor` 返回 JSON：`{services:[{name,type:dependency\|app,status:OK\|DEGRADED\|DOWN,reason}], summary:{ok,degraded,down,total}}`；`make dev-up` 在所有启用服务 OK 后才 exit 0 | E4 release-gate 与未来 A5 远端 CI（仅触发条件成立后）可消费此输出；不得硬编码旧的 7-service 口径 |
| D-7 | 数据持久化默认 | 命名卷（非 bind mount）：`easyinterview-pg-data` / `easyinterview-redis-data` / `easyinterview-minio-data`；`make dev-down` 不删卷，`make dev-reset` 才删 | 避免误操作丢失本地开发数据 |
| D-8 | 本地观测口径 | 默认依赖容器日志与应用 `/metrics`；`make dev-logs` 汇总容器日志，`make dev-doctor` 可检查已启用 HTTP 组件的 `/metrics` | F1 可以消费这些出口，但不能要求 A2 默认安装观测栈 |
| D-9 | 本地 AI provider 配置 | `deploy/dev-stack/.env.example` 必须列出 `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY` 占位；启用 AIClient 的项目组件启动时缺少真实 provider endpoint / key 必须 fail-fast；A2 不启动 AI gateway 容器 | 本地部署验证真实 LLM 服务，同时保持 A2 依赖最小化 |

### 3.2 待确认事项

- `make dev-up` 是否自动跑 `make migrate`：默认不自动；B4 落地后由开发者显式执行，避免数据库 schema 变更与服务启动耦合。
- 若某个未来组件暂不具备 Dockerfile 或稳定 dev command，该组件 plan 必须先补齐本地运行入口，再声明自己受 `make dev-up` 覆盖。

## 4 设计约束

### 4.1 结构约束

- A1 锁定的 `deploy/` 根容器之内，本 spec 只负责 `deploy/dev-stack/`；K8s manifests / Helm chart 由 E4 与 `test/scenarios/` 各自承接的子目录持有。
- `docker-compose.yaml` 必须使用 compose v2 schema（不显式声明 `version:` 字段）；默认 profile 只包含最小依赖与项目组件。任何可选 profile 都不得成为 `make dev-up` 的默认前提。
- 项目组件的 Dockerfile / dev command 由对应 child owner 提供，A2 负责把已具备运行入口的组件接入统一 compose 与 `make dev-doctor`。

### 4.2 健康检查约束

- 每个 compose service 必须配置容器级 `healthcheck`（compose `healthcheck:` 字段），间隔 ≤ 5s，重试 ≥ 3；纯一次性 init job 可通过退出码进入 `dev-doctor` 汇总。
- `make dev-doctor` 在容器健康基础上，对 Postgres / Redis / MinIO 必须执行端到端 probe（连接 + 读写最小操作 + 拆解延迟），不能只看容器状态。
- 对已具备 HTTP runtime 的项目组件，`make dev-doctor` 至少检查 `/healthz`；若组件已声明 `/metrics`，还必须检查 `/metrics` 可访问。
- 对已启用 AIClient 的项目组件，A2 只检查必要 env 是否存在并把结果纳入 `dev-doctor`；不得在 `dev-doctor` 中发起真实 LLM 付费调用。真实 provider smoke 由 A3 的本地部署验证负责。
- 健康检查超时上限 60s（`dev-up` 默认等待时长），超出后输出每个 DOWN 服务的最近日志尾段（≤ 50 行）。

### 4.3 性能与资源约束

- 默认本地环境并行启动后，2 vCPU + 8GB RAM 的开发机内必须可在 60s 内全部 healthy。
- Postgres / Redis / MinIO 默认资源不设硬限制；不得因为默认本地栈引入 Grafana / Loki / Prometheus 等常驻内存组件。
- 默认依赖镜像总下载体积控制在 < 1.5GB；任何 child 想引入 > 500MB 的新镜像或常驻服务必须在本 spec 修订流程中登记，并证明不阻塞默认 `make dev-up`。

### 4.4 文档约束

- `deploy/dev-stack/README.md` 必须包含：服务表（name/port/credentials default）、项目组件表、`make dev-*` 命令清单、AI provider 配置说明（本地部署走真实 provider，stub 仅用于测试）、常见故障（端口占用 / 卷不可写 / 镜像拉取失败）应对、与 `test/scenarios/` 的 K8s/Kind 路径区别说明。
- 本 spec 修订（新增默认依赖 / 端口变更 / 镜像 major 升级 / 默认项目组件启动语义变更）必须递增版本 + history 记录。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| docker-compose 文件与 Make target | A2 | `deploy/dev-stack/` 全部内容、A1 占位 target 的真实实现 |
| 本地依赖服务 | A2 | Postgres+pgvector / Redis / MinIO 的版本、端口、卷、健康检查 |
| 项目组件运行入口 | 对应 child owner | backend / frontend / worker 等组件提供 Dockerfile 或 dev command；A2 负责统一 compose 接入 |
| Postgres 扩展启用 | A2 | `pgvector` 由 init script 启用；B4 不再重复 |
| DB schema migration | B4 | A2 提供空实例 + 扩展，schema 由 B4 落地 |
| AI provider 运行时配置 | A3 + A4 + A2 | A3 决定 AIClient / provider 行为；A4 决定 env 字典与 fail-fast；A2 只在 compose 中传递 `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY` 占位，不启动 AI gateway，不切 stub |
| 产品分析 / 自托管 PostHog | F2 | 不阻塞普通 `make dev-up` |
| 观测 SDK / 指标命名 / dashboard | F1 | F1 消费应用 `/metrics` 与日志出口；生产或可选观测栈不归 A2 默认依赖 |
| Secrets / config 抽象 | A4 | A2 仅锁 dev 默认值 |
| K8s / Kind 场景 | `test/scenarios/` + E4 | 双轨独立 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 一键拉起 | 干净 worktree（无既存容器与卷），仓库根已有 A1 `Makefile` 占位 `dev-up` | `make dev-up` | Postgres / Redis / MinIO 与当前所有已具备本地运行入口的项目组件全部启动；`make dev-doctor` 输出 `summary.ok==summary.total` 且依赖服务 OK 数为 3；exit 0 | 001（A2 自身后续 plan） |
| C-2 | 失败可观察 | Postgres 5432 或任一已启用项目组件 host port 已被占用 | `make dev-up` | 退出码非 0；stderr 输出冲突服务名 + 占用进程提示；其它服务允许已启动；`make dev-doctor` 输出对应服务 `status=DOWN,reason="port conflict"` | 001（A2 自身后续 plan） |
| C-3 | idempotent | 已运行 `make dev-up` 一次 | 再次执行 `make dev-up` | 已 healthy 服务保持运行不重启；输出说明 `already healthy`；exit 0；数据卷不被清空 | 001 |
| C-4 | 安全停止 | 服务正在运行 | `make dev-down` | 容器停止；命名卷保留；下一次 `make dev-up` 数据完整可读 | 001 |
| C-5 | 显式清空 | 服务正在运行 | `make dev-reset` | 容器停止 + 命名卷删除；操作前提示交互确认（CI 模式跳过通过 `DEV_RESET_FORCE=1`） | 001 |
| C-6 | pgvector 可用 | `make dev-up` 完成 | 在 Postgres 中执行 `select extname from pg_extension where extname='vector'` | 返回一行，确认扩展已启用 | 001 |
| C-7 | 本地指标与日志可查 | `make dev-up` 完成，至少一个已启用 HTTP 项目组件声明 `/metrics` | 访问该组件 `/metrics` 并执行 `make dev-logs` | `/metrics` 返回文本指标；容器日志可按服务名查看；不依赖 Grafana / Loki / Prometheus / OTel Collector | 001 |
| C-8 | A2 executable gate handoff（来自 [001 Phase 3.6](../engineering-roadmap/plans/001-decompose-subspecs/checklist.md#phase-3-wave-1基础设施--契约骨架)） | 本 spec 的 contract lock 已完成，A2 `001-bootstrap` plan 完成 | C-1 + C-7 + C-9 都成立 | A2 的 `make dev-up` 可执行 gate 通过；依赖本地栈的 W2 implementation 可启动；parent Phase 3 只记录 spec-contract lock，不单独冒充本项已通过 | 001-bootstrap |
| C-9 | 本地 AI provider 配置不走 stub | 启用了需要 AIClient 的 API / worker；`.env` 缺 `AI_GATEWAY_BASE_URL` 或 `AI_GATEWAY_API_KEY` | `make dev-up` / `make dev-doctor` | 组件启动失败或 dev-doctor 报 DOWN/DEGRADED 并说明缺真实 AI provider 配置；补齐真实 provider endpoint / key 后组件健康；不启动 AI gateway 容器，也不把部署切到 stub | 001 |

## 7 关联计划

A2 `001-bootstrap` 已完成并落地本地依赖 compose、dev lifecycle Make targets、dev-doctor 与文档入口；本 spec 现在以 `001-bootstrap` 作为唯一已完成实现计划：

- 落地 `deploy/dev-stack/docker-compose.yaml` 与 init scripts。
- 实现 `make dev-up` / `make dev-down` / `make dev-doctor` / `make dev-reset` / `make dev-logs` 的真实命令体（替换 A1 占位）。
- 确保默认 compose 启动最小依赖与当前项目所有可运行组件，不包含 OTel Collector / Grafana / Loki / Prometheus / AI gateway。
- 确保默认 compose 将真实 AI provider / OpenAI-compatible endpoint 配置传给需要 AIClient 的项目组件；缺失时 fail-fast，不默认降级到 stub。
- 提供 `deploy/dev-stack/README.md` 与故障排查。

A2 后续如需扩展（新增默认依赖、接入新项目组件），在原 spec 上递增版本，原地修订；不创建 sibling spec。
