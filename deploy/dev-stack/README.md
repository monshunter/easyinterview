# Local Dev Stack

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-04-28

本目录承载 [local-dev-stack/001-bootstrap](../../docs/spec/local-dev-stack/plans/001-bootstrap/plan.md) 的运行时实现。默认 `make dev-up` 只启动 P0 闭环必须的最小依赖与当前仓库已具备本地运行入口的项目组件，**默认本地栈不包含 OTel Collector / Grafana / Loki / Prometheus / AI provider**。

## 1 前置条件

| 工具 | 最低版本 | 备注 |
|------|----------|------|
| Docker Engine | 24+ | macOS 用 Docker Desktop ≥ 4.x；Linux 原生 docker 同样 24+ |
| Docker Compose 插件 | v2.20+ | 必须支持 `docker compose up --wait` 以及无 `version:` 字段的 compose v2 schema |
| `jq` | 1.6+ | `make dev-doctor` 强依赖 |
| `lsof` | 任意 | 端口冲突诊断；缺失时 `dev-doctor` 会跳过 host port owner 解析 |
| `curl` | 任意 | 验证已声明 `/metrics` 的项目组件 |
| Bash 4+ / POSIX sh | — | Makefile 与 `dev-doctor.sh` 均按 POSIX sh 编写 |

资源占用：≥ 8GB RAM 推荐；默认依赖镜像（`pgvector/pgvector:pg16` + `redis:7-alpine` + `minio/minio` + `minio/mc`）首次拉取总体积 < 1.5GB。慢网络下先 `make dev-pull` 预热再 `make dev-up`。

## 2 默认服务

### 2.1 依赖服务

| name | image | host port | 默认凭据 | 命名卷 |
|------|-------|-----------|----------|--------|
| `postgres-dev` | `pgvector/pgvector:pg16` | `${POSTGRES_HOST_PORT:-5432}` | `easyinterview` / `dev` (DB `easyinterview`) | `easyinterview-pg-data` |
| `redis-dev` | `redis:7-alpine` | `${REDIS_HOST_PORT:-6379}` | 无密码（dev only） | `easyinterview-redis-data` |
| `minio-dev` | `minio/minio:RELEASE.2024-12-18T13-15-44Z` | `${MINIO_API_HOST_PORT:-9000}` API + `${MINIO_CONSOLE_HOST_PORT:-9001}` Console | `dev-access-key` / `dev-secret-key` | `easyinterview-minio-data` |
| `minio-init` | `minio/mc:RELEASE.2024-11-21T17-21-54Z` | — | 复用 minio-dev 凭据 | — (一次性 init job) |

所有服务通过 bridge network `easyinterview-dev` 互访，短名解析（`postgres-dev` / `redis-dev` / `minio-dev`）。Postgres init 自动启用 `pgvector` 扩展（D-5）。MinIO init 幂等创建默认 bucket `easyinterview-dev`（D-1..D-9）。

### 2.2 项目组件

当前没有具备本地运行入口的项目组件接入 compose。后续 backend / frontend / worker 等 child 落地 Dockerfile 或 dev command 时，必须把 service 接入本 compose 文件并打 label：

| label | 含义 |
|-------|------|
| `easyinterview.dev-stack.role=app` | `dev-doctor` 自动发现并 probe |
| `easyinterview.dev-stack.host-port=<port>` | host 上暴露的端口（`/healthz` / `/metrics` 拉取入口） |
| `easyinterview.dev-stack.healthz=/healthz` | HTTP health endpoint 路径 |
| `easyinterview.dev-stack.metrics=/metrics` | metrics endpoint 路径（声明后强制 curl 非空校验） |
| `easyinterview.dev-stack.aiclient=true` | 启用 AIClient 的组件，`dev-doctor` 校验 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 已注入；缺失即报 DOWN |

## 3 命令清单

| 命令 | 用途 |
|------|------|
| `make dev-up` | 启动依赖 + 项目组件；空仓克隆即跑（D-3 健康超时 60s）；幂等：全部 healthy 时打印 `already healthy` 并退出 0（C-3） |
| `make dev-down` | 停止容器，**保留命名卷**（C-4），下次 `dev-up` 数据完整 |
| `make dev-doctor` | 输出 D-6 锁定的 JSON 健康报告（services + summary）；`summary.down==0 && summary.degraded==0` 才退出 0 |
| `make dev-reset` | 停止容器并 **删除** 三个命名卷（C-5）；交互式 `read` 确认；`DEV_RESET_FORCE=1` 跳过确认 |
| `make dev-logs` | 汇总打印近期容器日志；`SERVICE=<name>` 限定到单个 service |
| `make dev-pull` | 预拉锁定 tag 的依赖镜像（慢网络） |

`dev-up` 启动顺序：先按 `--wait` 等 3 个依赖 healthy → 启动 `minio-init` 并轮询其退出码 (timeout 60s) → 调用 `dev-doctor.sh` 作为最终 gate；任何阶段失败时 `up` 退出非 0 并把每个 dependency / init 服务最近 50 行 `docker logs` 打到 stderr。

## 4 配置

`deploy/dev-stack/.env.example` 列出所有可调字段，字段名与 [secrets-and-config §3.1.1 P0 env 字典](../../docs/spec/secrets-and-config/spec.md#311-p0-必备-env-key-字典) 对齐。`make dev-up` 第一次执行时若 `.env` 不存在会自动从 `.env.example` 复制；`.env` 由根 `.gitignore` 忽略，**真实 `AI_PROVIDER_API_KEY` 不得提交**。

### 4.1 AI provider 配置

docker compose 与 Kind 本地部署都连接真实 AI provider / OpenAI-compatible endpoint：

- 必须设置 `AI_PROVIDER_BASE_URL`（任何 OpenAI-compatible URL，例如 `https://api.openai.com/v1`、自托管 vLLM endpoint 等）。
- 必须设置 `AI_PROVIDER_API_KEY` 为对应 provider 的真实 key。
- 缺任一字段时启用 AIClient 的组件 fail-fast，`make dev-doctor` 对该组件报 DOWN/DEGRADED 且 reason 指向缺真实 AI provider 配置（C-9）。
- **不允许** 把本地部署降级到单元测试 stub。stub 仅在 `APP_ENV=test`（单元测试 / 离线契约测试）下使用，由 [A3 ai-provider-and-model-routing](../../docs/spec/ai-provider-and-model-routing/spec.md) 承接。

## 5 与场景测试的关系

本目录是 **应用本地开发** 的 docker compose 路径；场景集成测试的 Kind / K8s 路径尚未在当前仓库落地，后续由 [engineering-roadmap S3](../../docs/spec/engineering-roadmap/spec.md#64-s3--true-integration-and-release-gate) 的 E2E / release workstream 按 on-demand 规则创建。两条路径互不依赖：

- 应用 dev → 用 `make dev-up`
- BDD / E2E 场景 → 待对应 workstream 创建 `test/scenarios/` 后，以该目录 README 锁定的入口为准

## 6 故障排查

| 现象 | 应对 |
|------|------|
| `make dev-up` 报 `bind: address already in use` | `make dev-doctor` 会指出占用端口的 pid 与 cmd；通过 `.env` 的 `*_HOST_PORT` 字段 override 端口，不要修改容器内端口 |
| Postgres healthcheck 反复失败但容器在跑 | `make dev-logs SERVICE=postgres-dev`；常见原因：旧卷里 `POSTGRES_PASSWORD` / `POSTGRES_USER` 与 `.env` 不一致 → `DEV_RESET_FORCE=1 make dev-reset` 后重新 up |
| MinIO 启动报 `volume not writable` | macOS Docker Desktop 偶发权限缓存问题；`docker volume rm easyinterview-minio-data` 后重新 up |
| `pgvector` 扩展未启用 | 数据卷是历史遗留时 init 脚本不会重跑：`docker exec easyinterview-postgres-dev psql -U easyinterview -d easyinterview -c "CREATE EXTENSION IF NOT EXISTS vector;"` 或重置卷 |
| 镜像首次拉取超过 60s healthy 预算 | 先 `make dev-pull` 预热再 `make dev-up`；预算只针对 image 已在本地的稳态 |
| `make dev-doctor` 对启用 AIClient 的组件报 DOWN | 检查 `.env` 中 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 是否填了真实 provider；勿提交真实 key |
| macOS Docker Desktop 端口冲突看似没冲突 | docker-desktop 用 IPv6 监听，纯 IPv4 squatter 不冲突；用 `python3 -c "...AF_INET6 + IPV6_V6ONLY=0..."` 或绑双栈监听才会触发真实冲突 |

## 7 升级与扩展

任何对默认依赖镜像的 major 升级、默认端口变更、新增默认依赖、引入新的常驻组件都必须先递增 [local-dev-stack spec](../../docs/spec/local-dev-stack/spec.md) 版本与 history。本 README 与 `.env.example` 的更新随 plan 修订一起走，不单独 bump。

## 8 CI 边界

当前单人开发阶段不在远端 CI 拉起 dev stack；`make lint` / `make test` / `make build` 仅依赖单元测试 stub。本 plan **不创建也不修改** [A5 ci-pipeline-baseline](../../docs/spec/ci-pipeline-baseline/spec.md) 的远端 workflow；A5 的远端验证触发条件成立后，再由 A5 原地评估是否在 CI 集成本地栈。
