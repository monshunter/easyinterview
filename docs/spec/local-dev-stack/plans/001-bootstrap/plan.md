# Local Dev Stack Bootstrap

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-05-04

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [local-dev-stack spec](../../spec.md) §3.1 已锁定的 D-1..D-9 决策落到仓库：在 `deploy/dev-stack/` 下创建默认最小 compose、init 脚本与项目组件启动约定，把 [repo-scaffold §2.1](../../../repo-scaffold/plans/001-bootstrap/plan.md#21-根-makefile) 占位的 `make dev-up` / `make dev-down` 替换为真实实现并新增 `make dev-doctor` / `make dev-reset` / `make dev-logs`，使「克隆仓库 → `make dev-up` → Postgres / Redis / MinIO + 当前项目可运行组件全部 healthy」可由开发者本机重复跑通；其中启用 AIClient 的组件必须连接真实 AI provider / OpenAI-compatible endpoint，不默认走单元测试 stub。

本 plan 是 `local-dev-stack` 唯一的 plan；后续如需扩展默认依赖或新增项目组件接入，递增 spec 与本 plan 版本，原地修订，不再开 sibling plan。

## 2 背景

[engineering-roadmap §5.1](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 A2 保留为当前 active Foundation spec；后续 workstream 依赖本地数据库 / 缓存 / 对象存储以及统一项目启动入口。本 plan 通过 §4 的 4 个 phase 验收 spec §6 C-1..C-9，关闭 roadmap 历史 rebaseline 中保留的 A2 executable gate 承诺。

每个 phase 是可独立部署 / 验证的纵向行为切片：Phase 1 起来就能用 `docker compose` 直连最小依赖与项目组件；Phase 2 起来就能用 `make` 管理生命周期；Phase 3 起来就能机器消费 `make dev-doctor` JSON；Phase 4 收口应用 `/metrics`、容器日志与文档。本 plan 不引入 BDD 资产（场景覆盖由后续 [e2e-scenarios-p0](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) workstream 承接），AC 验证完全由 `make dev-*` 命令驱动。

## 3 质量门禁分类

- **Plan 类型**: `tooling + dev-infra + code-internal`。本 plan 修改本地 docker-compose dev stack、Make targets、doctor 脚本、README 与健康检查约定；不产生用户可见 UI、HTTP API 行为或业务 workflow。
- **TDD 策略**: 历史实现以 checklist 中每个 phase 的 `自检` 命令作为 Red-Green-Refactor 断言来源；重进本 plan 时必须通过 `/implement` -> `/tdd` 顺序执行，优先以 `make dev-*`、`dev-doctor` JSON schema/probe、端口冲突复现、volume idempotency 和 README smoke 作为 focused assertions。
- **BDD 策略**: BDD 不适用。本 plan 只交付开发环境基础设施；后续 P0 用户行为场景由 `e2e-scenarios-p0` 或具体 feature plan 维护 BDD。
- **替代验证 gate**: `make dev-up`、`make dev-doctor`、`make dev-down`、`make dev-reset`、端口冲突复现、pgvector probe、AI provider fail-fast smoke、`sync-doc-index --check`、Markdown link check、`git diff --check`。

## 4 实施步骤

### Phase 1: docker-compose 与 init 脚本

#### 1.1 `deploy/dev-stack/docker-compose.yaml`

按 spec D-1..D-4 落地默认最小依赖服务（Postgres+pgvector / Redis / MinIO）以及当前仓库内所有已具备本地运行入口的项目组件。固定 D-2 镜像 tag、D-3 端口、D-7 命名卷与 D-4 `easyinterview-dev` bridge network。compose v2 schema（不写 `version:` 字段）；默认 compose 不预留也不启动 OTel Collector / Grafana / Loki / Prometheus / AI gateway。

每个服务必须配置容器级 `healthcheck`（间隔 ≤5s，重试 ≥3）：

- Postgres：`pg_isready -U $POSTGRES_USER`
- Redis：`redis-cli ping`
- MinIO：`mc ready local` 或 HTTP `/minio/health/live`
- 项目 HTTP 组件：`GET /healthz`；无 HTTP 端口的 worker 类组件使用进程存活检查 + 最近日志探测

#### 1.2 init 脚本与 provisioning

- `deploy/dev-stack/init/postgres/01-pgvector.sql`：`CREATE EXTENSION IF NOT EXISTS vector;`，由 Postgres image `/docker-entrypoint-initdb.d/` 自动执行（D-5）。
- `deploy/dev-stack/init/minio/create-buckets.sh`：通过 `mc` 创建默认 bucket（bucket 名按 spec §2.1 与 [A4 secrets-and-config](../../../secrets-and-config/spec.md) dev defaults 对齐为 `easyinterview-dev`），bucket 已存在不报错。
- 不落地 Grafana / OTel / Loki / Prometheus provisioning；本地观测通过应用 `/metrics` 与容器日志完成。

#### 1.3 数据卷命名（D-7）

`docker-compose.yaml` 顶层 `volumes:` 节声明 `easyinterview-pg-data` / `easyinterview-redis-data` / `easyinterview-minio-data` 三个命名卷；不使用 bind mount。

#### 1.4 dev `.env` 与 config 默认值

`deploy/dev-stack/.env.example` 落地连接串、bucket 名、端口、项目组件 host/port、AI provider / OpenAI-compatible endpoint 的本地默认占位；字段名与 [A4 secrets-and-config spec](../../../secrets-and-config/spec.md) 对齐（如 `DATABASE_URL` / `REDIS_URL` / `OBJECT_STORAGE_ENDPOINT` / `OBJECT_STORAGE_BUCKET` / `API_HOST_PORT` / `FRONTEND_HOST_PORT` / `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY`）。`.env`（无 `.example` 后缀）由根 `.gitignore` 忽略；`make dev-up` 第一次运行时若 `.env` 不存在则从 `.env.example` 复制。`AI_GATEWAY_API_KEY` 在 `.env.example` 中只能写占位说明，不能写真实 key。

#### 1.5 Phase 1 自检

- `cd deploy/dev-stack && docker compose up -d`：Postgres / Redis / MinIO 与当前已接入的项目组件全部进入 `healthy`（用 `docker compose ps --format json | jq` 校验）。
- `docker exec` 进入 Postgres 容器执行 `psql -U $POSTGRES_USER -d $POSTGRES_DB -c "select extname from pg_extension where extname='vector'"` 必须返回 1 行（关闭 spec C-6）。
- 拆下后 `docker compose down`（不带 `--volumes`），命名卷在 `docker volume ls` 中保留。

### Phase 2: Make targets 与生命周期语义

#### 2.1 替换 A1 占位 `make dev-up` / `make dev-down`

根 `Makefile` 中 [repo-scaffold §2.1](../../../repo-scaffold/plans/001-bootstrap/plan.md#21-根-makefile) 的两个占位 target 改为递归调用 `$(MAKE) -C deploy/dev-stack up` / `down`；`deploy/dev-stack/Makefile` 承载真实实现（`docker compose -f docker-compose.yaml --project-name easyinterview-dev up -d --wait` 等）。

新增 phony target：`dev-doctor`、`dev-reset`、`dev-logs`、`dev-pull`，并在根 `make help` 输出。

#### 2.2 idempotent 与 already-healthy 处理（C-3）

`deploy/dev-stack/Makefile` 的 `up` target 在调用 `docker compose up -d --wait` 前先 `docker compose ps --filter "status=running" --format json` 探测；若全部服务都已 running 且 healthy，跳过启动并打印 `already healthy` 后 exit 0。重复执行不重启容器，命名卷不被清空。

#### 2.3 dev-down 卷保留（C-4）

`down` target 调用 `docker compose down`（不带 `--volumes`、不带 `--rmi`）。下一次 `up` 必须读取到 §2.2 留下的命名卷里的数据（用「写入 → down → up → 读取」的自检覆盖）。

#### 2.4 dev-reset 显式清空（C-5）

`reset` target 必须先打印一段 stderr 警告（列出将删除的命名卷名 + 当前占用 size）并要求 `read -p` 输入 `yes` 确认；环境变量 `DEV_RESET_FORCE=1` 时跳过交互（CI 用）。确认后调用 `docker compose down --volumes` 删除命名卷。

#### 2.5 Phase 2 自检

- 干净仓库 `make dev-up` 一次 → 再次 `make dev-up`：第二次 < 5s 完成且日志含 `already healthy`。
- 写入 Postgres 一行测试数据 → `make dev-down` → `make dev-up` → 再次查询：数据仍在。
- `DEV_RESET_FORCE=1 make dev-reset` → `docker volume ls`：3 个命名卷已删除；下一次 `make dev-up` 从空卷重建。
- 不带 `DEV_RESET_FORCE` 时 stdin 输入 `no`：reset 必须 abort 且不删卷。

### Phase 3: dev-doctor 结构化健康检查

#### 3.1 dev-doctor 输出契约（D-6）

`deploy/dev-stack/scripts/dev-doctor.sh`（POSIX sh + jq；不引入 Go binary 以控镜像体积与依赖）输出 JSON：

```json
{
  "services": [
    {"name":"postgres","type":"dependency|app","status":"OK|DEGRADED|DOWN","reason":"..."}
  ],
  "summary": {"ok": 0, "degraded": 0, "down": 0, "total": 0}
}
```

退出码：`summary.down == 0 && summary.degraded == 0` 时 exit 0；否则 exit 1。`make dev-doctor` 把脚本输出原样打到 stdout，stderr 留给执行过程日志。脚本可以固定 3 个默认依赖名，但项目组件列表必须来自 compose service labels 或统一配置，不能硬编码旧的 7-service 口径。

#### 3.2 端到端 probe 实现（spec §4.2）

容器健康基础上对 PG / Redis / MinIO 必须执行 e2e probe：

- Postgres：`pg_isready` + 一次 `select 1` + 一次 `vector` 扩展检查。
- Redis：`redis-cli set __doctor__ ok EX 5` + `get` + `del`。
- MinIO：`mc ls` 默认 bucket（不存在则报 DEGRADED + reason）。

项目 HTTP 组件：`GET /healthz` 返回 2xx；若该组件声明 `/metrics`，`GET /metrics` 必须返回非空文本。worker 类组件使用容器运行状态 + 最近 50 行日志无启动失败特征作为最小 probe。对启用 AIClient 的组件，doctor 只校验 `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY` 是否注入并报告缺失；不得在 doctor 中调用真实 LLM。

#### 3.3 dev-up gate 接入（C-1）

`deploy/dev-stack/Makefile` 的 `up` target 在 `docker compose up -d --wait` 完成后调用 `dev-doctor.sh`；只在 `summary.ok == total` 时 exit 0。否则输出 DOWN / DEGRADED 服务的最近 50 行 `docker logs` 尾段并 exit 1。

#### 3.4 失败可观察（C-2）

构造 Postgres 端口 5432 或任一已启用项目组件 host port 被本机进程占用的复现路径（例如 `nc -l 5432 &`）：

- `make dev-up` 退出码非 0；
- stderr 输出冲突服务名 + 占用进程提示（用 `lsof -nP -iTCP:5432` 或等价命令）；
- `make dev-doctor` 对冲突服务报 `status=DOWN, reason="port conflict: ..."`，其它已起来的服务保持 OK。

#### 3.5 Phase 3 自检

- 全员 OK 时 `make dev-doctor` JSON 通过 `jq` schema 校验（自动化验证脚本固定 3 个依赖名，并校验项目组件来自 compose），exit 0。
- 故意 `docker stop redis-dev` 后 `make dev-doctor`：redis 报 DOWN，summary.down=1，exit 1。
- 删除 `.env` 中 `AI_GATEWAY_BASE_URL` 或 `AI_GATEWAY_API_KEY` 后，启用 AIClient 的组件必须 fail-fast，`make dev-doctor` 对该组件报 DOWN/DEGRADED 且 reason 指向缺真实 AI provider 配置；补齐真实 provider endpoint / key 后恢复。
- §3.4 的端口冲突复现路径手动跑通并贴日志。

### Phase 4: 指标日志 + 文档 + AC 收口

#### 4.1 应用 `/metrics` 与容器日志验证（C-7）

对每个已启用 HTTP 项目组件执行轻量验证：

- `curl http://localhost:${PORT}/metrics` 返回非空文本指标（仅对已声明 `/metrics` 的组件强制）。
- `make dev-logs SERVICE=<name>` 能输出对应容器最近日志。
- `make dev-doctor` 对该组件维持 OK。

本 plan 不创建 OTLP smoke，不安装 OTel Collector / Grafana / Loki / Prometheus。F1 后续可基于 `/metrics` 与 stdout/stderr 日志接入生产或可选观测链路。

#### 4.2 `deploy/dev-stack/README.md`

按 spec §4.4 必须包含：

- 服务表（name / image / port / 默认 credentials / 命名卷）。
- 项目组件表（component / compose service / host port / health endpoint / metrics endpoint）。
- `make dev-*` 命令清单与典型用例。
- 常见故障排查（端口占用 / 卷不可写 / 镜像拉取失败 / pgvector 扩展未启用）。
- AI provider 配置说明：docker compose 本地部署使用真实 provider / OpenAI-compatible endpoint；`.env.example` 只提供 `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY` 占位，真实 key 写入被 `.gitignore` 忽略的 `.env`；单元测试 stub 不适用于本地部署。
- 与 `test/scenarios/`（Kind 路径，归 E2/E4）的双轨说明：本地 docker-compose 走应用 dev，Kind 走场景测试，互不依赖。
- 资源占用提示（≥ 8GB RAM 推荐）+ 默认依赖镜像总下载体积估算（< 1.5GB，对 spec §4.3 兑现）。
- 明确默认本地栈不包含 OTel Collector / Grafana / Loki / Prometheus / AI gateway。

#### 4.3 工具版本与 CI 兼容性核对

- `.tool-versions` 不需要新增字段（docker / docker compose 不通过 asdf 管理）；但 `deploy/dev-stack/README.md` 必须声明最低 docker engine（24+）与 compose plugin（v2.20+）版本。
- 本 plan 不创建或修改 [A5 ci-pipeline-baseline](../../../ci-pipeline-baseline/spec.md) 的远端 CI workflow；当前单人开发阶段不在 CI 拉起 dev stack。若未来满足 A5 触发条件，再由 A5 原地评估是否增加远端验证。

#### 4.4 A2 executable gate handoff（C-8）

收口验证依次跑：

- `make dev-up`（C-1）→ exit 0，`make dev-doctor` summary.ok==summary.total，且 3 个默认依赖均 OK。
- AI provider 配置校验（C-9）→ 缺 `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY` fail-fast，补齐真实 provider endpoint / key 后恢复；全程不启动 AI gateway 容器，不切 stub。
- pgvector SQL probe（C-6）→ 返回 1 行。
- `/metrics` + `make dev-logs` 验证（C-7）→ 已声明 metrics 的项目组件返回非空指标，容器日志可按服务名查看。
- `make dev-down` → 卷保留；`make dev-up` 数据完整（C-4 复跑一次）。
- `DEV_RESET_FORCE=1 make dev-reset` → 卷清空（C-5 复跑一次）。
- 端口冲突路径（C-2 复跑一次）。
- 重复 `make dev-up`（C-3 复跑一次）。

完成后在工作日志贴出 8 条 AC 的执行证据；spec §6 表格中 C-1..C-9 全部成立。roadmap 历史 rebaseline 中保留的 A2 executable gate 承诺由本 phase 关闭；不再修改 parent checklist。

#### 4.5 文档收口

- `deploy/dev-stack/README.md` Header / 内容完整。
- `docs/spec/local-dev-stack/plans/INDEX.md` 把本 plan 从 active 切到 completed。
- 调用 `/sync-doc-index --check` 确认 `docs/spec/INDEX.md` 与 plans/INDEX 对 Header 无 drift。

## 5 验收标准

- spec [§6 验收标准](../../spec.md#6-验收标准) C-1 到 C-9 全部成立，证据贴入工作日志。
- 本 plan checklist 全部勾选；Phase 3 / Phase 4 的 `make dev-*` 自检命令日志贴入工作日志。
- engineering-roadmap 历史 rebaseline 中保留的 A2 executable gate 承诺由 Phase 4.4 关闭；不重复修改父 roadmap checklist。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 依赖镜像首次拉取在慢网络下超过 spec §4.3 60s healthy 预算 | `make dev-pull` 单独提供 image 预拉；README 提示首次执行 `make dev-pull` 后再 `make dev-up`；预算只针对 image 已在本地的稳态 |
| macOS Docker Desktop 与 Linux 原生 docker daemon 在 healthcheck / volume 性能上行为差异 | Phase 1.5 / 2.5 / 3.5 自检在 macOS 至少手跑一轮；任何平台特定 hack 必须落到 README 故障排查 |
| `dev-doctor.sh` 在 POSIX sh + jq 下复杂度上升后难维护 | 单脚本 ≤ 200 行；超出阈值时升级到 Go binary（落 `scripts/dev-doctor/`），但本 plan 不预先升级 |
| 默认端口（5432 / 6379 / 9000 / 9001 / 项目组件端口）与开发者本机已运行的服务冲突 | C-2 已覆盖端口冲突报错路径；README 提示用 `.env` override `*_HOST_PORT` 字段，不修改容器内端口；本 plan 不实现 host port 自动避让 |
| init 脚本中 MinIO bucket 创建在 image 升级后字段格式漂移 | 镜像 tag 锁定在 spec D-2；任何 major 升级走 spec 修订流程而非本 plan 静默 bump |
| 未来组件没有 Dockerfile 或稳定 dev command，导致无法纳入 `make dev-up` | 对应组件 plan 必须先补齐本地运行入口，再声明自己受 `make dev-up` 覆盖；A2 只接入已具备运行入口的组件 |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-05-04 | 1.4 | L1 plan-review remediation：补齐当前强制的质量门禁分类，不改变已完成 dev stack 范围。 | historical-spec-implementation-review/001 |
