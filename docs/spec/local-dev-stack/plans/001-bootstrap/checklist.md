# Local Dev Stack Bootstrap Checklist

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-05-22

**关联计划**: [plan](./plan.md)

## Phase 1: docker-compose 与 init 脚本

- [x] 1.1 落地 `deploy/dev-stack/docker-compose.yaml`：默认最小外部依赖（Postgres / Redis / MinIO）+ 已显式接入的 optional 项目组件；backend/frontend 默认宿主机 dev command 运行；按 D-2 锁定镜像 tag、D-3 端口、D-4 network alias、D-7 命名卷；compose v2 schema；默认不预留也不启动 OTel / Grafana / Loki / Prometheus / AI provider；每个 compose service 配置容器级 healthcheck（≤5s 间隔、≥3 重试）
- [x] 1.2 落地 init 脚本：`init/minio/create-buckets.sh` 创建默认 bucket（幂等）；Postgres 默认不启用未使用扩展；不创建 Grafana / OTel / Loki / Prometheus provisioning
- [x] 1.3 顶层 `volumes:` 声明 3 个命名卷（pg-data / redis-data / minio-data），不使用 bind mount；Postgres 18 命名卷挂到 `/var/lib/postgresql`，保留官方镜像 `PGDATA=/var/lib/postgresql/18/docker`，不挂到 `/var/lib/postgresql/data`；`make dev-up` 启动前只读检测旧卷布局并给出显式 reset 指引，不自动删卷
- [x] 1.4 落地 `deploy/dev-stack/.env.example`：连接串 / bucket 名 / 依赖端口 / 项目组件 host port / `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 占位默认值，字段名与 A4 secrets-and-config spec 对齐；`.env` 由根 `.gitignore` 忽略；`make dev-up` 首次运行时若 `.env` 不存在则从 `.env.example` 复制；`.env.example` 不写真实 AI key
- [x] 1.5 Phase 1 自检：`docker compose up -d` 后 Postgres / Redis / MinIO 与已接入 optional 项目组件均 healthy；Postgres 内 `select 1` 返回 1 行（关闭 C-6）；`docker compose down`（无 `--volumes`）后命名卷保留

## Phase 2: Make targets 与生命周期语义

- [x] 2.1 替换 repo-scaffold 占位 `make dev-up` / `dev-down`：根 `Makefile` 递归调用 `deploy/dev-stack/Makefile` 真实实现，默认只管理外部依赖与已显式接入 optional app service；新增 `dev-doctor` / `dev-reset` / `dev-logs` / `dev-pull` phony target 并入根 `make help`
- [x] 2.2 实现 dev-up idempotent（C-3）：`docker compose ps` 全 running+healthy 时跳过启动，打印 `already healthy` 并 exit 0；重复执行不重启容器
- [x] 2.3 实现 dev-down 卷保留（C-4）：`docker compose down`（不带 `--volumes` / `--rmi`）；自检写入测试数据 → down → up → 数据仍在
- [x] 2.4 实现 dev-reset 显式清空（C-5）：交互式 `read -p` 确认；`DEV_RESET_FORCE=1` 跳过交互；确认后 `docker compose down --volumes` 删除 3 个命名卷；输入 `no` 时 abort 不删卷
- [x] 2.5 Phase 2 自检：第二次 `make dev-up` 含 `already healthy` 且 <5s；写入-down-up-读取链路数据保留；`DEV_RESET_FORCE=1 make dev-reset` 后 3 卷消失；`make dev-reset` stdin `no` 时不删卷

## Phase 3: dev-doctor 结构化健康检查

- [x] 3.1 落地 `deploy/dev-stack/scripts/dev-doctor.sh`（POSIX sh + jq，≤200 行）：输出 spec D-6 锁定的 JSON 结构（services 含 `type=dependency|app` + summary）；`summary.down==0 && summary.degraded==0` 时 exit 0；不得硬编码旧 7-service 口径
- [x] 3.2 实现 e2e probe：PG `pg_isready` + `select 1`；Redis set/get/del 一次；MinIO `mc ls` 默认 bucket；optional 项目 HTTP 组件查 `/healthz`，已声明 `/metrics` 的组件查 `/metrics` 非空；宿主机 backend/frontend 由对应 owner 的 dev command / scenario runner 验证；启用 AIClient 的组件只校验真实 provider env 已注入，不调用真实 LLM
- [x] 3.3 dev-up gate 接入（C-1）：`up` target 在 `docker compose up -d --wait` 后调用 dev-doctor；`summary.ok == total` 才 exit 0；否则输出 DOWN/DEGRADED 服务的最近 50 行 `docker logs` 尾段
- [x] 3.4 失败可观察（C-2）：构造 Postgres 5432 或任一已接入 optional 项目组件 host port 冲突复现路径；`make dev-up` 非 0 退出且 stderr 含冲突服务名 + 占用进程；`make dev-doctor` 对冲突服务报 `status=DOWN, reason="port conflict: ..."`，其它服务保持 OK
- [x] 3.5 Phase 3 自检：全员 OK 时 dev-doctor JSON 通过 schema 校验（3 个依赖名固定，项目组件来自 compose）且 exit 0；`docker stop redis-dev` 后报 DOWN/exit 1；缺 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 时启用 AIClient 的组件 fail-fast 且 dev-doctor 报缺真实 provider 配置；端口冲突复现路径日志贴入工作日志

## Phase 4: 指标日志 + 文档 + AC 收口

- [x] 4.1 应用 `/metrics` 与容器日志验证（C-7）：对已接入 compose 且声明 `/metrics` 的 optional 项目组件执行 curl 非空校验；当前没有 optional app service 时验证依赖容器日志；宿主机 backend/frontend 由对应 owner 的 dev command / scenario runner 验证；不创建 OTLP smoke，不安装 OTel / Grafana / Loki / Prometheus
- [x] 4.2 落地 `deploy/dev-stack/README.md`：服务表（name/image/port/credentials/volume）、optional 项目组件表（component/service/host port/health/metrics）与宿主机运行边界、`make dev-*` 命令清单、AI provider 配置说明（非测试本地 app run 使用真实 provider，stub 仅单测）、常见故障排查、与 `test/scenarios/` 本地 runner 契约说明、资源占用提示与默认依赖镜像下载体积估算（< 1.5GB），并声明默认本地栈不包含 OTel / Grafana / Loki / Prometheus / AI provider
- [x] 4.3 在 `deploy/dev-stack/README.md` 声明最低 docker engine（24+）与 compose plugin（v2.20+）版本；本 plan 不创建或修改 A5 远端 CI workflow，当前单人开发阶段不在 CI 拉起 dev stack
- [x] 4.4 A2 executable gate handoff（C-8）：依次复跑 C-1 / C-2 / C-3 / C-4 / C-5 / C-6 / C-7 / C-9 八项 AC；执行证据贴入工作日志；不修改 engineering-roadmap parent checklist
- [x] 4.5 文档收口：`deploy/dev-stack/README.md` Header 完整；plans/INDEX.md 把本 plan 切到 completed 段；`/sync-doc-index --check` 通过
