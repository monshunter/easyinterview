# Local Dev Stack Spec

> **版本**: 1.16
> **状态**: active
> **更新日期**: 2026-05-27

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将历史 A2 `local-dev-stack` 保留为当前 active Foundation spec（依赖 [A1 `repo-scaffold`](../repo-scaffold/spec.md)）。它承接 A1 已占位的 `make dev-up` / `make dev-down` target；当前执行口径锁定本地开发栈的最小依赖、应用组件启动契约、Make target 行为与健康检查口径，真实「克隆仓库 → 一条命令 → 本项目本地环境启动完成」由 A2 child `001` plan 验证。

本地开发环境不是生产观测环境的缩小版。默认 `make dev-up` 只启动开发 P0 闭环必须的外部依赖；当前仓库内可运行的 backend / frontend 组件默认在宿主机用 dev command 管理，避免把生产才需要的 OTel Collector、Grafana、Loki、Prometheus 等高占用组件变成本地开发前置条件。

目标是：

1. **冻结最小依赖清单与版本**：默认本地依赖只包含 `PostgreSQL`、`Redis`、`MinIO`、`Mailpit`。任何 child 不得把生产观测、分析平台、AI provider 或其它重型组件加入默认 `make dev-up`。
2. **统一本地依赖启动契约**：`make dev-up` 必须通过 docker compose 启动最小外部依赖；backend API、frontend dev server 等项目组件默认通过宿主机 dev command 管理，只有组件 owner 明确接入 optional compose app service 时才进入本栈；P0 不接入独立 worker 进程，重复执行不破坏已有数据卷。
3. **健康检查可机器读**：`make dev-doctor` 对依赖服务与已显式接入 compose 的 optional 项目组件返回 `OK / DEGRADED / DOWN` 的结构化结果，`make dev-up` 的退出码反映整体健康状态。
4. **本地观测轻量化**：本地只要求应用自身暴露 `/metrics`（当组件已具备 HTTP runtime 时）并通过容器日志确认行为；不安装 Grafana / Loki / Prometheus / OTel Collector 作为默认依赖。
5. **本地 AI 调用真实化**：非测试本地 app run 只通过 env 注入真实 AI provider / OpenAI-compatible endpoint，不启动 AI provider 容器，也不把应用切到单元测试 stub。
6. **环境生命周期独立化**：测试环境与本地前后端联调环境必须能通过 scenario environment skill 和 repo-tracked env entrypoints 独立 setup / status / verify / cleanup / redeploy，不依附任一具体 `p0-*` 场景脚本；开发者可只构建环境，再人工或由 Agent 执行目标场景验证。
7. **单一真实 env 来源**：本地真实前后端联调只使用 `deploy/dev-stack/.env` 一个文件承接共享依赖、host-run backend、frontend real mode、auth secrets 与真实 AI provider 配置；具体场景不得复制独立 `.env`。
8. **本地 AI raw output 可调试**：local dev/test 与本地真实联调默认开启 `AI_DEBUG_PRINT_RAW_OUTPUT=true`，让 AI Agent 能确认真实 provider 输出格式；raw output 只进入本机 backend stderr / `.test-output/` 调试日志，不进入持久化审计、runtime-config、staging 或 prod 默认配置。

本 spec 不实现业务代码、不接入 K8s / Kind / Helm，也不部署 staging。当前 P0 本地测试与 smoke 以 Docker Compose 外部依赖 + 宿主机前后端运行 + `test/scenarios/` 本地 runner 脚本为准；[E4 `release-gate-and-rollout`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 后续如需部署级环境，必须重新评估并显式修订。

## 2 范围

### 2.1 In Scope

- 最小本地依赖清单：`PostgreSQL 18`、`Redis 7`、`MinIO`、`Mailpit` 的本地版本、端口、卷、network alias。
- 项目组件运行边界：backend / frontend 等项目组件默认通过宿主机 dev command 运行并连接本地依赖；只有对应 child owner 明确需要可复现容器化 app service 时，才接入同一个 compose，而不是另起平行本地环境。后台任务随 backend internal runner 观测，不单独接入 worker service。
- 顶层入口：`docker-compose.yaml`（落点 `deploy/dev-stack/docker-compose.yaml`）+ A1 已占位的 `make dev-up` / `make dev-down` 真实实现。
- `make dev-doctor`：结构化健康检查，对每个依赖服务与项目组件返回 `OK / DEGRADED / DOWN` 与人类可读原因（输出 JSON + 退出码）。
- 初始化脚本：MinIO 创建默认 bucket；Postgres 不启用未使用扩展。
- `.env` 与 `config.yaml` / `config/dev.yaml` / `config/test.yaml` 的最小 dev/test override（连接串、bucket 名、端口、应用组件默认 host/port、auth secrets、frontend real-mode env、AI provider endpoint 与 key 占位、local raw output debug 默认值）；具体 secrets 抽象与 feature flag 由 [A4 `secrets-and-config`](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 承接，本 spec 只锁 dev 默认值与字段名。
- 数据卷管理：默认命名 `easyinterview-pg-data` / `easyinterview-redis-data` / `easyinterview-minio-data`；Postgres 18 卷挂载到 `/var/lib/postgresql`，由官方镜像管理 `PGDATA=/var/lib/postgresql/18/docker`；`make dev-up` 必须只读检测旧布局或半初始化卷并给出显式 reset 指引；提供 `make dev-reset` 用于显式清空（非默认）。
- 文档：`deploy/dev-stack/README.md` 一屏说明 + 故障排查 + 本地 Mailpit 登录边界 + 与 `test/scenarios/` 本地 runner 场景契约的 cross-link。
- 场景环境入口：`test/scenarios/env-setup.sh` / `env-status.sh` / `env-verify.sh` / `env-cleanup.sh` / `env-redeploy.sh` 作为 framework-owned 环境生命周期入口；根 `Makefile` 提供 `scenario-env-*` target 委派这些脚本，供 `/scenario-env` 和 `/scenario-redeploy` skill 调用。

### 2.2 Out of Scope

- OTel Collector / Grafana / Loki / Prometheus 默认本地部署：归 [F1 `observability-stack`](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 或生产部署路径；A2 默认 `make dev-up` 不安装这些组件，也不以它们作为健康检查前提。
- AI provider 本地 mock 或基础设施服务：AI provider 不是 A2 本地依赖；A2 不启动 provider mock 容器。单元测试 stub / 离线 provider mock 归 A3；非测试本地 app run 或 optional compose app service 必须把真实 AI provider / OpenAI-compatible endpoint 配置传给应用组件。
- 自托管 PostHog：归 [F2 `analytics-funnel`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)；ADR-Q3 锁定不依赖 PostHog Cloud，但本地 dev-up 不强制启动 PostHog（资源占用大）。
- K8s / Kind / Helm 场景集群：当前不属于 P0 本地测试、smoke 或默认部署路径；`test/scenarios/` 只维护 repo-tracked 本地 runner 场景契约。若后续 release owner 需要部署级环境，必须先修订 ADR-Q4 / E4 与本 spec 的边界。
- 数据库迁移：归 [B4 `db-migrations-baseline`](../engineering-roadmap/spec.md#51-当前已存在的-active-spec)；A2 仅保证 Postgres 实例可用。
- 业务种子数据：归各 C 域 mock-server plan；本 spec 只提供空实例。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | docker-compose 落点 | `deploy/dev-stack/docker-compose.yaml`（A1 已锁定 `deploy/` 根容器） | 任何 child 不得在仓库根另起平行 compose 文件 |
| D-2 | 服务镜像 tag | 默认依赖锁定 `postgres:18-alpine` / `redis:7-alpine` / `minio/minio:RELEASE.2024-12-18T13-15-44Z` / `axllent/mailpit:v1.30.0`；MinIO bucket init 工具锁 `minio/mc:RELEASE.2024-11-21T17-21-54Z`。项目组件优先使用仓库内宿主机 dev command；只有 owner 显式接入 optional app service 时才使用仓库内 Dockerfile 构建，不从外部拉取业务镜像 | 升级须递增 spec 版本；默认 compose 不含生产观测镜像 |
| D-3 | 服务端口 | Postgres 5432 / Redis 6379 / MinIO 9000(API) + 9001(Console) / Mailpit 8025(Web UI) + 1025(SMTP)；项目组件端口由各组件 dev defaults 声明（frontend 默认 5173，backend 默认 8080） | 不预留 worker host port、Grafana 3000 / Prometheus 9090 / Loki 3100 / OTLP 4317/4318 给默认本地栈 |
| D-4 | network 命名 | `easyinterview-dev`（bridge 模式）；依赖服务与 optional compose app service 通过短名互访 | 宿主机 backend 默认通过 localhost 连接依赖；optional app service 启动时使用 `postgres-dev` / `redis-dev` / `minio-dev` 等命名解析 |
| D-5 | Postgres 扩展启用 | 当前不启用未使用扩展；后续 `pg_trgm` / `pg_stat_statements` 或向量扩展由 B4 决定是否前置 | A2 默认栈保持最小依赖 |
| D-6 | dev-up 健康检查口径 | `make dev-doctor` 返回 JSON：`{services:[{name,type:dependency\|app,status:OK\|DEGRADED\|DOWN,reason}], summary:{ok,degraded,down,total}}`；`make dev-up` 在所有启用服务 OK 后才 exit 0 | E4 release-gate 与未来 A5 远端 CI（仅触发条件成立后）可消费此输出；不得硬编码旧的 7-service 口径 |
| D-7 | 数据持久化默认 | 命名卷（非 bind mount）：`easyinterview-pg-data` / `easyinterview-redis-data` / `easyinterview-minio-data`；Postgres 18 命名卷必须挂到 `/var/lib/postgresql`，保持官方镜像 `PGDATA=/var/lib/postgresql/18/docker` 位于卷内，不挂到 `/var/lib/postgresql/data`；`make dev-down` 不删卷，`make dev-reset` 才删 | 避免误操作丢失本地开发数据，并兼容 Postgres 18 官方镜像目录布局 |
| D-8 | 本地观测口径 | 默认依赖容器日志与应用 `/metrics`；`make dev-logs` 汇总容器日志，`make dev-doctor` 可检查已启用 HTTP 组件的 `/metrics` | F1 可以消费这些出口，但不能要求 A2 默认安装观测栈 |
| D-9 | 本地 AI provider 配置 | `deploy/dev-stack/.env.example` 必须列出 `AI_PROVIDER_REGISTRY_PATH=config/ai-providers.yaml`、`AI_MODEL_PROFILE_PATH=config/ai-profiles.yaml` 与 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 占位；启用 AIClient 的非测试项目组件启动时缺少 catalog path 或当前 provider endpoint / key 必须 fail-fast；A2 不启动 AI provider 容器 | 本地 app run 验证真实 LLM 服务，同时保持 A2 依赖最小化 |
| D-10 | 本地邮件 sink | 默认依赖包含 Mailpit；`deploy/dev-stack/.env.example` 必须列出 `MAILPIT_WEB_HOST_PORT` / `MAILPIT_SMTP_HOST_PORT` 与 C1/A4 邮件 env（`EMAIL_PROVIDER=mailpit`、SMTP host/port、from、verify base URL）。host-run backend 默认通过 `127.0.0.1:1025` 投递 magic-link 到 Mailpit，人工通过 `http://127.0.0.1:8025` 收信 | 本地测试不依赖真实外部邮箱服务或真实邮箱账号；账号验收走真实 passwordless flow，不需要场景专属 backend cmd |
| D-11 | 独立场景环境生命周期 | `test/scenarios/env-setup.sh` 调 `make dev-up` 并可选执行 migrations；`env-status.sh` / `env-verify.sh` 消费 `make dev-doctor`；`env-cleanup.sh` 默认 `make dev-down`，显式 `--with-volumes` 才走 `DEV_RESET_FORCE=1 make dev-reset`；`env-redeploy.sh` 支持 `deps` / `backend` / `frontend` / `all`，在当前 host-run 口径下等价于依赖重启与 repo-tracked build artifact 刷新，不创建场景专属 backend/cmd 或长驻 helper | 环境管理从具体 `test/scenarios/e2e/p0-*` 场景脚本中抽离；skill 可按用户意图只管理环境、只重建组件，或再交给场景 runner/人工 UAT |
| D-12 | 单一真实 env 来源 | `deploy/dev-stack/.env.example` 必须列出真实本地联调所需的 backend auth secrets、AI provider、邮件、共享依赖、frontend real-mode env；`deploy/dev-stack/.env` 是唯一被 host-run backend/frontend 与 hybrid 场景读取的本地真实 env 文件 | 防止每个场景复制独立 `.env`，保证本地测试环境和真实联调环境通过同一配置构建 |
| D-13 | local raw output debug 默认开启 | `config/dev.yaml`、`config/test.yaml` 与 `deploy/dev-stack/.env.example` 必须默认 `AI_DEBUG_PRINT_RAW_OUTPUT=true`；`config/config.yaml`、staging、prod 仍默认 false；本地 hybrid 场景 preflight 必须拒绝未开启 raw debug 的真实 provider run | 支持 AI Agent 以本机 raw log 调试 schema/格式问题，同时不扩大生产或共享持久化泄露面 |

### 3.2 待确认事项

- `make dev-up` 是否自动跑 `make migrate`：默认不自动；B4 落地后由开发者显式执行，避免数据库 schema 变更与服务启动耦合。
- 若某个未来组件需要容器化 app service，该组件 plan 必须先补齐稳定运行入口、健康检查与资源预算，再声明自己受 `make dev-up` 覆盖；普通本地开发默认仍使用宿主机 dev command。
- `env-redeploy.sh` 当前不启动长期运行的 backend/frontend 进程；真实 provider hybrid UAT 的进程启动命令仍由目标场景 README 与用户本地 secret 文件控制，避免 skill 在无凭证上下文中生成悬挂进程或泄露 secrets。

## 4 设计约束

### 4.1 结构约束

- A1 锁定的 `deploy/` 根容器之内，本 spec 只负责 `deploy/dev-stack/`；K8s manifests / Helm chart 不属于当前 P0 本地验证路径，后续 release workstream 如需引入必须原地修订 owner spec / ADR。
- `docker-compose.yaml` 必须使用 compose v2 schema（不显式声明 `version:` 字段）；默认 profile 只包含最小依赖与已显式接入的 optional 项目组件。任何可选 profile 都不得成为 `make dev-up` 的默认前提。
- 项目组件的宿主机 dev command / optional Dockerfile 由对应 child owner 提供；A2 只负责统一外部依赖与已显式接入 compose 的 optional app service。

### 4.2 健康检查约束

- 每个 compose service 必须配置容器级 `healthcheck`（compose `healthcheck:` 字段），间隔 ≤ 5s，重试 ≥ 3；纯一次性 init job 可通过退出码进入 `dev-doctor` 汇总。Mailpit healthcheck 使用 HTTP `/readyz`。
- `make dev-up` 在启动 Postgres 前必须只读检测 `easyinterview-pg-data` 是否包含旧根目录 `PG_VERSION`、旧 `/var/lib/postgresql/data/PG_VERSION` 或半初始化 `/var/lib/postgresql/18` 布局；命中时退出非 0，提示用户确认本地数据后用 `make dev-reset` 重建，不得自动删除卷。
- `make dev-doctor` 在容器健康基础上，对 Postgres / Redis / MinIO 必须执行端到端 probe（连接 + 读写最小操作 + 拆解延迟），不能只看容器状态。
- 对已接入 compose 的 optional HTTP 项目组件，`make dev-doctor` 至少检查 `/healthz`；若组件已声明 `/metrics`，还必须检查 `/metrics` 可访问。P0 不接入独立 worker 进程；backend background runner 随 backend 组件观测。
- 对已接入 compose 且启用 AIClient 的 optional 项目组件，A2 只检查必要 env 是否存在并把结果纳入 `dev-doctor`；不得在 `dev-doctor` 中发起真实 LLM 付费调用。宿主机运行组件的真实 provider smoke 由 A3 / feature owner 的本地验证负责。
- 健康检查超时上限 60s（`dev-up` 默认等待时长），超出后输出每个 DOWN 服务的最近日志尾段（≤ 50 行）。

### 4.3 性能与资源约束

- 默认本地环境并行启动后，2 vCPU + 8GB RAM 的开发机内必须可在 60s 内全部 healthy。
- Postgres / Redis / MinIO 默认资源不设硬限制；不得因为默认本地栈引入 Grafana / Loki / Prometheus 等常驻内存组件。
- 默认依赖镜像总下载体积控制在 < 1.5GB；任何 child 想引入 > 500MB 的新镜像或常驻服务必须在本 spec 修订流程中登记，并证明不阻塞默认 `make dev-up`。

### 4.4 文档约束

- `deploy/dev-stack/README.md` 必须包含：服务表（name/port/credentials default）、optional 项目组件表、`make dev-*` 命令清单、AI provider 配置说明（非测试本地 app run 走真实 provider，stub 仅用于测试）、常见故障（端口占用 / 卷不可写 / 镜像拉取失败）应对、与 `test/scenarios/` 本地 runner 场景契约的区别说明。
- AI provider 配置说明必须明确 local dev/test 与本地真实联调默认开启 `AI_DEBUG_PRINT_RAW_OUTPUT=true`，raw output 仅用于本机调试，不得进入持久化审计、runtime-config 或 staging/prod 默认配置。
- `test/scenarios/README.md` 与 `test/scenarios/e2e/README.md` 必须说明独立环境入口与具体场景 `setup/trigger/verify/cleanup` 的边界：环境入口只能构建共享本地环境，不得引用固定 `p0-*` 场景目录；场景脚本只能消费环境，不得把共享环境 bootstrap 私有化。
- 本 spec 修订（新增默认依赖 / 端口变更 / 镜像 major 升级 / 默认项目组件启动语义变更）必须递增版本 + history 记录。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| docker-compose 文件与 Make target | A2 | `deploy/dev-stack/` 全部内容、A1 占位 target 的真实实现 |
| 本地依赖服务 | A2 | Postgres / Redis / MinIO / Mailpit 的版本、端口、卷、健康检查 |
| 项目组件运行入口 | 对应 child owner | backend / frontend 等组件提供宿主机 dev command；可选 Dockerfile / compose app service 需由 owner 显式设计 |
| Postgres 扩展启用 | B4 | A2 默认不启用未使用扩展；新增 DB extension 必须由 B4 owner spec 决策并同步 A2 |
| DB schema migration | B4 | A2 提供空实例 + 扩展，schema 由 B4 落地 |
| AI provider 运行时配置 | A3 + A4 + A2 | A3 决定 AIClient / provider 行为；A4 决定 env 字典与 fail-fast；A2 只在 compose 中传递 `AI_PROVIDER_REGISTRY_PATH` / `AI_MODEL_PROFILE_PATH` catalog 路径和 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 占位，不启动 AI provider，不切 stub |
| 产品分析 / 自托管 PostHog | F2 | 不阻塞普通 `make dev-up` |
| 观测 SDK / 指标命名 / dashboard | F1 | F1 消费应用 `/metrics` 与日志出口；生产或可选观测栈不归 A2 默认依赖 |
| Secrets / config 抽象 | A4 | A2 仅锁 dev 默认值；Mailpit email env 字典由 A4/C1 承接 |
| 场景 runner 契约 | `test/scenarios/` + feature owner | repo-tracked Go / Vitest / Playwright / browser runner；不默认要求 Kind / K8s |
| 场景环境生命周期 | A2 + `test/scenarios/` framework | framework-owned `env-*.sh` 与根 `scenario-env-*` Make target；只管理共享本地环境、build artifact 与 cleanup，不承接具体场景断言 |
| Release 部署环境 | E4 / release owner | 未创建；如需 staging / prod / K8s / Helm，必须单独原地设计并修订 ADR-Q4 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 一键拉起 | 干净 worktree（无既存容器与卷），仓库根已有 A1 `Makefile` 占位 `dev-up` | `make dev-up` | Postgres / Redis / MinIO / Mailpit 与已显式接入的 optional app service 全部启动；`make dev-doctor` 输出 `summary.ok==summary.total` 且依赖服务 OK 数为 4；exit 0；backend/frontend 可在宿主机 dev command 中连接这些依赖 | 001（A2 自身后续 plan） |
| C-2 | 失败可观察 | Postgres 5432 / Redis 6379 / MinIO 9000/9001 / Mailpit 8025/1025 或任一已接入 optional app service host port 已被占用 | `make dev-up` | 退出码非 0；stderr 输出冲突服务名 + 占用进程提示；其它服务允许已启动；`make dev-doctor` 输出对应服务 `status=DOWN,reason="port conflict"` | 001（A2 自身后续 plan） |
| C-3 | idempotent | 已运行 `make dev-up` 一次 | 再次执行 `make dev-up` | 已 healthy 服务保持运行不重启；输出说明 `already healthy`；exit 0；数据卷不被清空 | 001 |
| C-4 | 安全停止 | 服务正在运行 | `make dev-down` | 容器停止；命名卷保留；下一次 `make dev-up` 数据完整可读 | 001 |
| C-5 | 显式清空 | 服务正在运行 | `make dev-reset` | 容器停止 + 命名卷删除；操作前提示交互确认（CI 模式跳过通过 `DEV_RESET_FORCE=1`） | 001 |
| C-6 | Postgres 可用 | `make dev-up` 完成 | 在 Postgres 中执行 `select 1` | 返回一行，确认基础数据库连接可用 | 001 |
| C-7 | 本地指标与日志可查 | `make dev-up` 完成；若存在已接入 compose 的 HTTP optional app service，则该组件声明 `/metrics` | 访问该组件 `/metrics` 并执行 `make dev-logs`；宿主机运行组件由对应 owner 的 dev command / test gate 验证 | `/metrics` 返回文本指标；容器日志可按服务名查看；不依赖 Grafana / Loki / Prometheus / OTel Collector；未接入 app service 时本项以依赖日志 gate 收口 | 001 |
| C-8 | A2 executable gate handoff | 本 spec 的 contract lock 已完成，A2 `001-bootstrap` plan 完成 | C-1 + C-7 + C-9 都成立 | A2 的 `make dev-up` 可执行 gate 通过；依赖本地栈的后续 implementation 可启动；roadmap 只保留 active spec 关系，不单独冒充本项已通过 | 001-bootstrap |
| C-9 | 本地 AI provider 配置不走 stub | 启用了需要 AIClient 的 backend 运行路径；`.env` 缺 `AI_PROVIDER_REGISTRY_PATH` / `AI_MODEL_PROFILE_PATH` 或当前 profile 选中的 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` | 启动非测试 backend runtime；若该 runtime 已接入 compose，则同时通过 `make dev-up` / `make dev-doctor` 检查 | 组件启动失败或 dev-doctor 报 DOWN/DEGRADED 并说明缺真实 AI provider 配置；补齐 catalog path 与真实 provider endpoint / key 后组件健康；不启动 AI provider 容器，也不把部署切到 stub | 001 |
| C-10 | 本地邮箱登录 | `make dev-up` 完成，backend 以 `EMAIL_PROVIDER=mailpit` 和真实 auth secrets 在宿主机运行 | 用户调用 `POST /api/v1/auth/email/start` 或前端登录页提交 synthetic `.example.test` 邮箱 | Mailpit Web UI 出现 magic-link 邮件；点击或复制 token 后 `GET /api/v1/auth/email/verify` 签发 first-party `ei_session` cookie；不依赖真实外部邮箱服务、真实邮箱账号或场景专属 backend cmd | 001 Mailpit revision |
| C-11 | 独立环境 setup / verify / cleanup | 开发者或 Agent 只想准备共享本地环境，不准备立即运行具体场景 | 通过 `/scenario-env setup` / `verify` / `status` / `cleanup`，或等价根 `make scenario-env-*` target | 调用 `test/scenarios/env-*.sh` 顶层脚本；setup 启动 dev-stack，verify/status 返回 dev-doctor JSON，cleanup 默认保留命名卷；全流程不引用任何 `test/scenarios/e2e/p0-*` 目录，也不新增 `backend/cmd` / Go helper | 001 environment lifecycle revision |
| C-12 | 独立 rebuild / redeploy | 开发者或 Agent 已有共享本地环境，希望刷新当前 backend/frontend consumer artifacts 后再人工或自动验证 | `/scenario-redeploy backend|frontend|all` 或 `make scenario-env-redeploy TARGET=backend|frontend|all` | backend target 执行 `go build ./cmd/...`，frontend target 执行 `pnpm --filter @easyinterview/frontend build`，all 同时覆盖依赖环境 verify；不会运行具体 BDD 场景，也不会启动真实 provider 付费调用 | 001 environment lifecycle revision |
| C-13 | 单一真实 env 文件 | 开发者或 Agent 准备真实前后端联调或 `E2E.P0.100` hybrid 场景 | 编辑 `deploy/dev-stack/.env` 并运行目标 backend/frontend/scenario preflight | `deploy/dev-stack/.env.example` 覆盖 auth、AI、邮件、共享依赖、frontend real-mode keys；场景脚本不得要求场景专属 `.env`；缺真实 key 时输出可解释的 `MANUAL_REQUIRED` 或 fail-fast，不降级到 stub | 001 + e2e-scenarios-p0/002 |
| C-14 | 本地 raw output debug 默认开启 | 开发者或 Agent 准备本地测试、host-run backend/frontend 联调或 `E2E.P0.100` hybrid 场景 | 加载 `config/dev.yaml` / `config/test.yaml` 或 `deploy/dev-stack/.env`，再启动 backend / 场景 preflight | local dev/test effective config 中 `ai.debugPrintRawOutput=true`；`deploy/dev-stack/.env.example` 含 `AI_DEBUG_PRINT_RAW_OUTPUT=true`；`E2E.P0.100` preflight 对缺失或非 true 输出 `MANUAL_REQUIRED`；staging/prod 默认仍关闭 | 001 raw debug local default revision |

## 7 关联计划

A2 `001-bootstrap` 已完成并落地本地依赖 compose、dev lifecycle Make targets、dev-doctor 与文档入口；本 spec 现在以 `001-bootstrap` 作为唯一已完成实现计划：

- 落地 `deploy/dev-stack/docker-compose.yaml` 与 MinIO init scripts。
- 实现 `make dev-up` / `make dev-down` / `make dev-doctor` / `make dev-reset` / `make dev-logs` 的真实命令体（替换 A1 占位）。
- 确保默认 compose 启动最小外部依赖与已显式接入的 optional app service，不包含 OTel Collector / Grafana / Loki / Prometheus / AI provider。
- 确保非测试 app runtime 将真实 AI provider / OpenAI-compatible endpoint 配置传给需要 AIClient 的项目组件；缺失时 fail-fast，不默认降级到 stub。
- 确保 local dev/test 与真实本地联调默认开启 `AI_DEBUG_PRINT_RAW_OUTPUT=true`，并把 raw output 限定在本机 stderr / `.test-output/` 调试日志。
- 提供 `deploy/dev-stack/README.md` 与故障排查。

A2 后续如需扩展（新增默认依赖、接入 optional 项目组件、改变本地场景 runner 边界），在原 spec 上递增版本，原地修订；不创建 sibling spec。
