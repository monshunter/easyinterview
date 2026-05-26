# Local Dev Stack

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-05-26

本目录承载 [local-dev-stack/001-bootstrap](../../docs/spec/local-dev-stack/plans/001-bootstrap/plan.md) 的运行时实现。默认 `make dev-up` 只启动 P0 闭环必须的外部依赖；backend / frontend 默认由宿主机 dev command 管理，只有组件 owner 明确接入 optional compose app service 时才进入本栈。**默认本地栈不包含 OTel Collector / Grafana / Loki / Prometheus / AI provider**；本地邮件通过轻量 Mailpit 依赖承接。

## 1 前置条件

| 工具 | 最低版本 | 备注 |
|------|----------|------|
| Docker Engine | 24+ | macOS 用 Docker Desktop ≥ 4.x；Linux 原生 docker 同样 24+ |
| Docker Compose 插件 | v2.20+ | 必须支持 `docker compose up --wait` 以及无 `version:` 字段的 compose v2 schema |
| `jq` | 1.6+ | `make dev-doctor` 强依赖 |
| `lsof` | 任意 | 端口冲突诊断；缺失时 `dev-doctor` 会跳过 host port owner 解析 |
| `curl` | 任意 | 验证已声明 `/metrics` 的项目组件 |
| Bash 4+ / POSIX sh | — | Makefile 与 `dev-doctor.sh` 均按 POSIX sh 编写 |

资源占用：≥ 8GB RAM 推荐；默认依赖镜像（`postgres:18-alpine` + `redis:7-alpine` + `minio/minio` + `minio/mc` + `axllent/mailpit`）首次拉取总体积 < 1.5GB。慢网络下先 `make dev-pull` 预热再 `make dev-up`。

## 2 默认服务

### 2.1 依赖服务

| name | image | host port | 默认凭据 | 命名卷 |
|------|-------|-----------|----------|--------|
| `postgres-dev` | `postgres:18-alpine` | `${POSTGRES_HOST_PORT:-5432}` | `easyinterview` / `dev` (DB `easyinterview`) | `easyinterview-pg-data` mounted at `/var/lib/postgresql`; official image PGDATA stays `/var/lib/postgresql/18/docker` |
| `redis-dev` | `redis:7-alpine` | `${REDIS_HOST_PORT:-6379}` | 无密码（dev only） | `easyinterview-redis-data` |
| `minio-dev` | `minio/minio:RELEASE.2024-12-18T13-15-44Z` | `${MINIO_API_HOST_PORT:-9000}` API + `${MINIO_CONSOLE_HOST_PORT:-9001}` Console | `dev-access-key` / `dev-secret-key` | `easyinterview-minio-data` |
| `minio-init` | `minio/mc:RELEASE.2024-11-21T17-21-54Z` | — | 复用 minio-dev 凭据 | — (一次性 init job) |
| `mailpit-dev` | `axllent/mailpit:v1.30.0` | `${MAILPIT_WEB_HOST_PORT:-8025}` Web UI + `${MAILPIT_SMTP_HOST_PORT:-1025}` SMTP | 无账号（dev only） | — |

所有服务通过 bridge network `easyinterview-dev` 互访，短名解析（`postgres-dev` / `redis-dev` / `minio-dev` / `mailpit-dev`）。MinIO init 幂等创建默认 bucket `easyinterview-dev`；Mailpit 提供本地 magic-link 收信，不需要真实外部邮箱服务或真实邮箱账号。

### 2.2 项目组件

当前没有项目组件接入默认 compose。backend / frontend 本地开发优先使用宿主机 dev command 直接运行，并连接本栈提供的 Postgres / Redis / MinIO。后续只有在组件 owner 明确需要可复现容器化 app service 时，才把 service 接入本 compose 文件并打 label：

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
| `make dev-up` | 启动外部依赖与已显式接入的 optional app service；空仓克隆即跑（D-3 健康超时 60s）；幂等：全部 healthy 时打印 `already healthy` 并退出 0（C-3） |
| `make dev-down` | 停止容器，**保留命名卷**（C-4），下次 `dev-up` 数据完整 |
| `make dev-doctor` | 输出 D-6 锁定的 JSON 健康报告（services + summary）；`summary.down==0 && summary.degraded==0` 才退出 0 |
| `make dev-reset` | 停止容器并 **删除** 三个命名卷（C-5）；交互式 `read` 确认；`DEV_RESET_FORCE=1` 跳过确认 |
| `make dev-logs` | 汇总打印近期容器日志；`SERVICE=<name>` 限定到单个 service |
| `make dev-pull` | 预拉锁定 tag 的依赖镜像（慢网络） |

`dev-up` 启动顺序：先只读检查 Postgres 18 命名卷布局是否仍是旧 `/var/lib/postgresql/data` 或半初始化状态 → 按 `--wait` 等 4 个依赖 healthy → 启动 `minio-init` 并轮询其退出码 (timeout 60s) → 调用 `dev-doctor.sh` 作为最终 gate；任何阶段失败时 `up` 退出非 0 并把每个 dependency / init 服务最近 50 行 `docker logs` 打到 stderr。

## 4 配置

`deploy/dev-stack/.env.example` 列出所有可调字段，字段名与 [secrets-and-config §3.1.1 P0 env 字典](../../docs/spec/secrets-and-config/spec.md#311-p0-必备-env-key-字典) 对齐。`make dev-up` 第一次执行时若 `.env` 不存在会自动从 `.env.example` 复制；`.env` 由根 `.gitignore` 忽略，**真实 `AI_PROVIDER_API_KEY` 不得提交**。

### 4.1 AI provider 配置

非测试本地 app run 与后续部署都连接真实 AI provider / OpenAI-compatible endpoint；当前 repo-tracked 开发主力为 DeepSeek：

- 必须保留 `AI_PROVIDER_REGISTRY_PATH=config/ai-providers.yaml` 与 `AI_MODEL_PROFILE_PATH=config/ai-profiles.yaml`，本地组件从单一 provider registry / profile catalog 加载 AI 路由。
- 必须设置 `AI_PROVIDER_BASE_URL=https://api.deepseek.com`，与 `deepseek` provider ref 对齐。
- 必须设置 `AI_PROVIDER_API_KEY` 为对应 provider 的真实 key。
- 缺 registry/profile 路径或当前 profile 选中的 provider secret 时，启用 AIClient 的组件 fail-fast，`make dev-doctor` 对该组件报 DOWN/DEGRADED 且 reason 指向缺真实 AI provider 配置（C-9）。
- **不允许** 把本地部署降级到单元测试 stub。stub 仅在 `APP_ENV=test`（单元测试 / 离线契约测试）下使用，由 [A3 ai-provider-and-model-routing](../../docs/spec/ai-provider-and-model-routing/spec.md) 承接。

### 4.2 本地邮件配置

本地 passwordless 登录默认走 Mailpit：

- `EMAIL_PROVIDER=mailpit`
- `EMAIL_SMTP_HOST=127.0.0.1`
- `EMAIL_SMTP_PORT=1025`
- `EMAIL_FROM_ADDRESS=noreply@easyinterview.local`
- `EMAIL_VERIFY_BASE_URL=http://127.0.0.1:8080/api/v1/auth/email/verify`

Mailpit Web UI 默认在 `http://127.0.0.1:8025`。backend 以 `APP_ENV=dev` 启动后，`startAuthEmailChallenge` 会通过 `email_dispatch` handler 向 Mailpit SMTP 投递 magic-link 邮件；人工 UAT 使用 synthetic `.example.test` 邮箱即可完成登录，不需要真实邮箱账号。

## 5 与场景测试的关系

本目录是 **应用本地开发依赖** 的 Docker Compose 路径；`test/scenarios/` 是 BDD / E2E 场景契约路径。两条路径互不替代：

- 应用 dev → 用 `make dev-up` 启动 Postgres / Redis / MinIO / Mailpit 依赖；backend/frontend 进程默认在宿主机单独启动并消费这些连接串。
- BDD / E2E 场景 → 以 [test/scenarios/README.md](../../test/scenarios/README.md) 和目标套件 README 为准。当前 P0 场景默认由 shell / Python 脚本编排既有产品 runner（例如已有包测试、Vitest、Playwright、browser smoke）验证同一行为契约；场景专属依赖不得新增正式 `backend/cmd` / Go helper 进程，不要求 Kind / K8s / Helm 环境。
- 需要真实 AI provider 的应用部署不得降级到单元测试 stub；`APP_ENV=test` 以外缺真实 provider config 时必须 fail-fast。

## 6 故障排查

| 现象 | 应对 |
|------|------|
| `make dev-up` 报 `bind: address already in use` | `make dev-doctor` 会指出占用端口的 pid 与 cmd；通过 `.env` 的 `*_HOST_PORT` 字段 override 端口，不要修改容器内端口 |
| Postgres healthcheck 反复失败但容器在跑 | `make dev-logs SERVICE=postgres-dev`；常见原因：旧卷里 `POSTGRES_PASSWORD` / `POSTGRES_USER` 与 `.env` 不一致，或历史 compose 曾把卷挂到 Postgres 18 不兼容的 `/var/lib/postgresql/data`；确认要清空本地开发数据后再执行 `DEV_RESET_FORCE=1 make dev-reset` 并重新 up |
| MinIO 启动报 `volume not writable` | macOS Docker Desktop 偶发权限缓存问题；`docker volume rm easyinterview-minio-data` 后重新 up |
| 镜像首次拉取超过 60s healthy 预算 | 先 `make dev-pull` 预热再 `make dev-up`；预算只针对 image 已在本地的稳态 |
| `make dev-doctor` 对启用 AIClient 的组件报 DOWN | 检查 `.env` 中 `AI_PROVIDER_REGISTRY_PATH` / `AI_MODEL_PROFILE_PATH` 是否指向 repo 内 catalog，并确认 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 填了真实 provider；勿提交真实 key |
| Mailpit 收不到邮件 | 确认 `make dev-doctor` 中 `mailpit-dev` 为 OK；backend env 包含 `EMAIL_PROVIDER=mailpit` / `EMAIL_SMTP_HOST=127.0.0.1` / `EMAIL_SMTP_PORT=1025`；若本机 1025 或 8025 被占用，用 `.env` 调整 `MAILPIT_*_HOST_PORT` 并同步 backend email env |
| macOS Docker Desktop 端口冲突看似没冲突 | docker-desktop 用 IPv6 监听，纯 IPv4 squatter 不冲突；用 `python3 -c "...AF_INET6 + IPV6_V6ONLY=0..."` 或绑双栈监听才会触发真实冲突 |

## 7 升级与扩展

任何对默认依赖镜像的 major 升级、默认端口变更、新增默认依赖、引入新的常驻组件都必须先递增 [local-dev-stack spec](../../docs/spec/local-dev-stack/spec.md) 版本与 history。本 README 与 `.env.example` 的更新随 plan 修订一起走，不单独 bump。

## 8 CI 边界

当前单人开发阶段不在远端 CI 拉起 dev stack；`make lint` / `make test` / `make build` 仅依赖单元测试 stub。本 plan **不创建也不修改** [A5 ci-pipeline-baseline](../../docs/spec/ci-pipeline-baseline/spec.md) 的远端 workflow；A5 的远端验证触发条件成立后，再由 A5 原地评估是否在 CI 集成本地栈。
