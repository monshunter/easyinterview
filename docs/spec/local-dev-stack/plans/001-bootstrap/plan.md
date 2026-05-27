# Local Dev Stack Bootstrap

> **版本**: 1.13
> **状态**: completed
> **更新日期**: 2026-05-27

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [local-dev-stack spec](../../spec.md) §3.1 已锁定的 D-1..D-10 决策落到仓库：在 `deploy/dev-stack/` 下创建默认最小 compose、init 脚本与 optional 项目组件接入约定，把 [repo-scaffold §2.1](../../../repo-scaffold/plans/001-bootstrap/plan.md#21-根-makefile) 占位的 `make dev-up` / `make dev-down` 替换为真实实现并新增 `make dev-doctor` / `make dev-reset` / `make dev-logs`，使「克隆仓库 → `make dev-up` → Postgres / Redis / MinIO / Mailpit healthy；backend / frontend 通过宿主机 dev command 连接这些依赖」可由开发者本机重复跑通；其中启用 AIClient 的非测试组件必须连接真实 AI provider / OpenAI-compatible endpoint，不默认走单元测试 stub。

本 plan 是 `local-dev-stack` 唯一的 plan；后续如需扩展默认依赖或新增项目组件接入，递增 spec 与本 plan 版本，原地修订，不再开 sibling plan。本次 1.13 revision 将本地 redeploy 收口为 build + 重启 host-run backend/frontend，并把服务地址、日志路径、PID 文件与容器日志入口作为 env 脚本固定输出，避免开发者在 Agent 启动环境后无法接管调试。

## 2 背景

[engineering-roadmap §5.1](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 A2 保留为当前 active Foundation spec；后续 workstream 依赖本地数据库 / 缓存 / 对象存储以及统一项目启动入口。本 plan 通过 §4 的 4 个 phase 验收 spec §6 C-1..C-9，关闭 roadmap 历史 rebaseline 中保留的 A2 executable gate 承诺。

每个 phase 是可独立部署 / 验证的纵向行为切片：Phase 1 起来就能用 `docker compose` 直连最小外部依赖；Phase 2 起来就能用 `make` 管理生命周期；Phase 3 起来就能机器消费 `make dev-doctor` JSON；Phase 4 收口 optional 应用 `/metrics`、依赖容器日志与文档。本 plan 不引入 BDD 资产；当前 `test/scenarios/` 场景覆盖由具体 feature plan 通过 repo-tracked 本地 runner 维护，AC 验证完全由 `make dev-*` 命令驱动。

## 3 质量门禁分类

- **Plan 类型**: `tooling + dev-infra + code-internal`。本 plan 修改本地 docker-compose dev stack、Make targets、doctor 脚本、README 与健康检查约定；不产生用户可见 UI、HTTP API 行为或业务 workflow。
- **TDD 策略**: 历史实现以 checklist 中每个 phase 的 `自检` 命令作为 Red-Green-Refactor 断言来源；重进本 plan 时必须通过 `/implement` -> `/tdd` 顺序执行，优先以 `make dev-*`、`dev-doctor` JSON schema/probe、端口冲突复现、volume idempotency 和 README smoke 作为 focused assertions。
- **BDD 策略**: BDD 不适用。本 plan 只交付开发环境基础设施；后续 P0 用户行为场景由 `e2e-scenarios-p0` 或具体 feature plan 维护 BDD。
- **替代验证 gate**: `make dev-up`、`make dev-doctor`、`make dev-down`、`make dev-reset`、`make scenario-env-setup` / `status` / `verify` / `cleanup` / `redeploy` dry-run 与 focused live gate、端口冲突复现、Postgres connectivity probe、AI provider fail-fast smoke、local raw debug config tests、P0.100 preflight contract、`sync-doc-index --check`、Markdown link check、`git diff --check`。

## 4 实施步骤

### Phase 1: docker-compose 与 init 脚本

#### 1.1 `deploy/dev-stack/docker-compose.yaml`

按 spec D-1..D-4 落地默认最小依赖服务（Postgres / Redis / MinIO）以及已显式接入的 optional 项目组件。backend / frontend 默认通过宿主机 dev command 运行，不因具备本地入口就强制进入 compose。固定 D-2 镜像 tag、D-3 端口、D-7 命名卷与 D-4 `easyinterview-dev` bridge network。compose v2 schema（不写 `version:` 字段）；默认 compose 不预留也不启动 OTel Collector / Grafana / Loki / Prometheus / AI provider。

每个服务必须配置容器级 `healthcheck`（间隔 ≤5s，重试 ≥3）：

- Postgres：`pg_isready -U $POSTGRES_USER`
- Redis：`redis-cli ping`
- MinIO：`mc ready local` 或 HTTP `/minio/health/live`
- Optional 项目 HTTP 组件：`GET /healthz`；backend internal runner 随 backend 组件观测，不单独声明进程 probe

#### 1.2 init 脚本与 provisioning

- Postgres 默认不挂载 extension init 脚本；新增 DB extension 必须先由 B4 owner spec 决策并同步本 plan。
- `deploy/dev-stack/init/minio/create-buckets.sh`：通过 `mc` 创建默认 bucket（bucket 名按 spec §2.1 与 [A4 secrets-and-config](../../../secrets-and-config/spec.md) dev defaults 对齐为 `easyinterview-dev`），bucket 已存在不报错。
- 不落地 Grafana / OTel / Loki / Prometheus provisioning；本地观测通过应用 `/metrics` 与容器日志完成。

#### 1.3 数据卷命名（D-7）

`docker-compose.yaml` 顶层 `volumes:` 节声明 `easyinterview-pg-data` / `easyinterview-redis-data` / `easyinterview-minio-data` 三个命名卷；不使用 bind mount。Postgres 18 的 `easyinterview-pg-data` 必须挂到 `/var/lib/postgresql`，不挂到 `/var/lib/postgresql/data`，让官方镜像保持 `PGDATA=/var/lib/postgresql/18/docker` 并把真实数据库目录放进命名卷。`make dev-up` 在启动 Postgres 前必须只读检测旧根目录 `PG_VERSION`、旧 `/var/lib/postgresql/data/PG_VERSION` 或半初始化 `/var/lib/postgresql/18` 布局；命中时明确提示用户确认数据后用 `make dev-reset` 重建，不自动删除卷。

#### 1.4 dev `.env` 与 config 默认值

`deploy/dev-stack/.env.example` 落地连接串、bucket 名、端口、项目组件 host/port、auth secrets、frontend real-mode env、AI provider / OpenAI-compatible endpoint 的本地默认占位；字段名与 [A4 secrets-and-config spec](../../../secrets-and-config/spec.md) 对齐（如 `DATABASE_URL` / `REDIS_URL` / `OBJECT_STORAGE_ENDPOINT` / `OBJECT_STORAGE_BUCKET` / `API_HOST_PORT` / `FRONTEND_HOST_PORT` / `SESSION_COOKIE_SECRET` / `AUTH_CHALLENGE_TOKEN_PEPPER` / `VITE_EI_API_MODE` / `VITE_EI_API_BASE_URL` / `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY`）。`.env`（无 `.example` 后缀）由根 `.gitignore` 忽略；`make dev-up` 第一次运行时若 `.env` 不存在则从 `.env.example` 复制。`AI_PROVIDER_API_KEY`、auth secrets 与其它本地 secret 在 `.env.example` 中只能为空占位，不能写真实值；真实本地联调和 hybrid 场景都必须读取这一个 `.env` 文件。

#### 1.5 Phase 1 自检

- `cd deploy/dev-stack && docker compose up -d`：Postgres / Redis / MinIO 与已接入的 optional 项目组件全部进入 `healthy`（用 `docker compose ps --format json | jq` 校验）。
- `docker exec` 进入 Postgres 容器执行 `psql -U $POSTGRES_USER -d $POSTGRES_DB -c "select 1"` 必须返回 1 行（关闭 spec C-6）。
- 拆下后 `docker compose down`（不带 `--volumes`），命名卷在 `docker volume ls` 中保留。

### Phase 2: Make targets 与生命周期语义

#### 2.1 替换 A1 占位 `make dev-up` / `make dev-down`

根 `Makefile` 中 [repo-scaffold §2.1](../../../repo-scaffold/plans/001-bootstrap/plan.md#21-根-makefile) 的两个占位 target 改为递归调用 `$(MAKE) -C deploy/dev-stack up` / `down`；`deploy/dev-stack/Makefile` 承载真实实现（`docker compose -f docker-compose.yaml --project-name easyinterview-dev up -d --wait` 等），只默认管理外部依赖和已显式接入的 optional app service。

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

- Postgres：`pg_isready` + 一次 `select 1`。
- Redis：`redis-cli set __doctor__ ok EX 5` + `get` + `del`。
- MinIO：`mc ls` 默认 bucket（不存在则报 DEGRADED + reason）。

Optional 项目 HTTP 组件：`GET /healthz` 返回 2xx；若该组件声明 `/metrics`，`GET /metrics` 必须返回非空文本。宿主机运行的 backend / frontend 由对应 owner 的 dev command、unit/integration test 或 scenario runner 验证；backend internal runner 通过 backend 组件状态、最近日志与 `/metrics` 出口观测，不单独注册本地进程 probe。对启用 AIClient 的组件，doctor 只校验 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 是否注入并报告缺失；不得在 doctor 中调用真实 LLM。

#### 3.3 dev-up gate 接入（C-1）

`deploy/dev-stack/Makefile` 的 `up` target 在 `docker compose up -d --wait` 完成后调用 `dev-doctor.sh`；只在 `summary.ok == total` 时 exit 0。否则输出 DOWN / DEGRADED 服务的最近 50 行 `docker logs` 尾段并 exit 1。

#### 3.4 失败可观察（C-2）

构造 Postgres 端口 5432 或任一已接入 optional 项目组件 host port 被本机进程占用的复现路径（例如 `nc -l 5432 &`）：

- `make dev-up` 退出码非 0；
- stderr 输出冲突服务名 + 占用进程提示（用 `lsof -nP -iTCP:5432` 或等价命令）；
- `make dev-doctor` 对冲突服务报 `status=DOWN, reason="port conflict: ..."`，其它已起来的服务保持 OK。

#### 3.5 Phase 3 自检

- 全员 OK 时 `make dev-doctor` JSON 通过 `jq` schema 校验（自动化验证脚本固定 3 个依赖名，并校验项目组件来自 compose），exit 0。
- 故意 `docker stop redis-dev` 后 `make dev-doctor`：redis 报 DOWN，summary.down=1，exit 1。
- 删除 `.env` 中 `AI_PROVIDER_BASE_URL` 或 `AI_PROVIDER_API_KEY` 后，启用 AIClient 的组件必须 fail-fast，`make dev-doctor` 对该组件报 DOWN/DEGRADED 且 reason 指向缺真实 AI provider 配置；补齐真实 provider endpoint / key 后恢复。
- §3.4 的端口冲突复现路径手动跑通并贴日志。

### Phase 4: 指标日志 + 文档 + AC 收口

#### 4.1 应用 `/metrics` 与容器日志验证（C-7）

对每个已接入 compose 的 optional HTTP 项目组件执行轻量验证；当前没有 optional app service 时，本项以依赖容器日志和 README 边界说明收口：

- `curl http://localhost:${PORT}/metrics` 返回非空文本指标（仅对已声明 `/metrics` 的组件强制）。
- `make dev-logs SERVICE=<name>` 能输出对应容器最近日志。
- `make dev-doctor` 对该组件维持 OK。
- 宿主机运行 backend / frontend 的验证不由 A2 伪造，归对应 feature / runtime owner 的本地 dev command、单测或 scenario runner。

本 plan 不创建 OTLP smoke，不安装 OTel Collector / Grafana / Loki / Prometheus。F1 后续可基于 `/metrics` 与 stdout/stderr 日志接入生产或可选观测链路。

#### 4.2 `deploy/dev-stack/README.md`

按 spec §4.4 必须包含：

- 服务表（name / image / port / 默认 credentials / 命名卷）。
- Optional 项目组件表（component / compose service / host port / health endpoint / metrics endpoint）与宿主机运行边界。
- `make dev-*` 命令清单与典型用例。
- 常见故障排查（端口占用 / 卷不可写 / 镜像拉取失败 / Postgres 连接失败）。
- AI provider 配置说明：非测试本地 app run 使用真实 provider / OpenAI-compatible endpoint；`.env.example` 只提供 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 占位，真实 key 写入被 `.gitignore` 忽略的 `.env`；单元测试 stub 不适用于本地部署。
- 与 `test/scenarios/` 的边界说明：本地 docker-compose 走外部依赖，场景验证走 repo-tracked Go / Vitest / Playwright / browser runner，默认不需要 Kind / K8s / Helm。
- 资源占用提示（≥ 8GB RAM 推荐）+ 默认依赖镜像总下载体积估算（< 1.5GB，对 spec §4.3 兑现）。
- 明确默认本地栈不包含 OTel Collector / Grafana / Loki / Prometheus / AI provider。

#### 4.3 工具版本与 CI 兼容性核对

- `.tool-versions` 不需要新增字段（docker / docker compose 不通过 asdf 管理）；但 `deploy/dev-stack/README.md` 必须声明最低 docker engine（24+）与 compose plugin（v2.20+）版本。
- 本 plan 不创建或修改 [A5 ci-pipeline-baseline](../../../ci-pipeline-baseline/spec.md) 的远端 CI workflow；当前单人开发阶段不在 CI 拉起 dev stack。若未来满足 A5 触发条件，再由 A5 原地评估是否增加远端验证。

#### 4.4 A2 executable gate handoff（C-8）

收口验证依次跑：

- `make dev-up`（C-1）→ exit 0，`make dev-doctor` summary.ok==summary.total，且 3 个默认依赖均 OK；backend/frontend 可通过宿主机 dev command 连接这些依赖。
- AI provider 配置校验（C-9）→ 缺 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` fail-fast，补齐真实 provider endpoint / key 后恢复；全程不启动 AI provider 容器，不切 stub。
- Postgres SQL probe（C-6）→ `select 1` 返回 1 行。
- `/metrics` + `make dev-logs` 验证（C-7）→ 已接入 optional app service 且声明 metrics 的项目组件返回非空指标；当前没有 optional app service 时验证依赖容器日志可按服务名查看。
- `make dev-down` → 卷保留；`make dev-up` 数据完整（C-4 复跑一次）。
- `DEV_RESET_FORCE=1 make dev-reset` → 卷清空（C-5 复跑一次）。
- 端口冲突路径（C-2 复跑一次）。
- 重复 `make dev-up`（C-3 复跑一次）。

完成后在工作日志贴出 8 条 AC 的执行证据；spec §6 表格中 C-1..C-9 全部成立。roadmap 历史 rebaseline 中保留的 A2 executable gate 承诺由本 phase 关闭；不再修改 parent checklist。

#### 4.5 文档收口

- `deploy/dev-stack/README.md` Header / 内容完整。
- `docs/spec/local-dev-stack/plans/INDEX.md` 把本 plan 从 active 切到 completed。
- 调用 `/sync-doc-index --check` 确认 `docs/spec/INDEX.md` 与 plans/INDEX 对 Header 无 drift。

### Phase 5: Mailpit 本地邮件 sink revision

#### 5.1 默认依赖接入

在 `deploy/dev-stack/docker-compose.yaml` 默认依赖中新增 `mailpit-dev`，镜像锁定 `axllent/mailpit:v1.30.0`，暴露 host Web UI `${MAILPIT_WEB_HOST_PORT:-8025}` 与 SMTP `${MAILPIT_SMTP_HOST_PORT:-1025}`，加入 `easyinterview-dev` network 与 dependency label，并配置 `/readyz` healthcheck。`DEPENDENCY_SERVICES`、端口冲突扫描、`make dev-pull`、`make dev-logs` 必须自动覆盖 Mailpit。

#### 5.2 dev-doctor probe

`dev-doctor.sh` 增加 `probe_mailpit`：容器状态必须 `running/healthy`，HTTP `GET /readyz` 必须 2xx；若本机有 `nc`，额外检查 SMTP host port 可连接。JSON summary 中 Mailpit 与 Postgres / Redis / MinIO 同为 dependency，C-1 依赖 OK 数从 3 调整为 4。

#### 5.3 backend-auth 与 A4 配置边界

后端 `cmd/api` 在 `EMAIL_PROVIDER=mailpit` 时必须注册 SMTP `DeliveryWriter`，通过 `auth_challenges` 查询收件邮箱、从 C1 transient delivery secret store 取 magic token，并将 magic-link 邮件投递到 Mailpit SMTP。`email_dispatch` payload 仍不得包含 raw email、magic token、完整 URL、邮件正文或标题；默认 `DevMailSink` 保留给单元测试和非 Mailpit 配置。A4 env/config 字典新增 `EMAIL_SMTP_HOST` / `EMAIL_SMTP_PORT` / `EMAIL_FROM_ADDRESS` / `EMAIL_VERIFY_BASE_URL`。

#### 5.4 hybrid UAT 账号入口回收

`E2E.P0.100` 不再通过直接 DB session bootstrap 取得账号；runbook 改为 synthetic `.example.test` 邮箱 + Mailpit magic link。场景工具仍只允许 shell / Python，且不得新增 `backend/cmd` / Go helper 作为验收依赖。

#### 5.5 Phase 5 自检

- Focused backend auth tests：SMTP writer 通过 fake SendMail 捕获 magic link，缺 delivery secret 不发送，lookup error 不泄露 raw email/token。
- Focused `cmd/api` test：`EMAIL_PROVIDER=mailpit` 时 `buildAuthService` 注册 `*auth.SMTPDeliveryWriter`。
- Focused config test：canonical env bindings 覆盖 Mailpit SMTP keys。
- Compose/static gates：`docker compose config --quiet`、`bash -n dev-doctor.sh`、`make -C deploy/dev-stack -n up`。
- Dev-stack live gate：`make dev-up && make dev-doctor` 输出 Postgres / Redis / MinIO / Mailpit 四个依赖 OK。
- Negative gate：`find test/scenarios -name '*.go' -type f` 无命中，`test ! -d backend/cmd/devsession && test ! -d backend/internal/devsession`。

### Phase 6: 独立 scenario/local integration environment lifecycle revision

#### 6.1 framework-owned env scripts

在 `test/scenarios/` 新增顶层环境入口，作为 `scenario-env` / `scenario-redeploy` skill 的唯一 repo-tracked 执行层：

- `env-setup.sh`：默认执行 `make dev-up` + `make dev-doctor`；`--with-migrations` 时在默认 `DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable` 下执行 `make migrate-up`。该脚本只能准备共享本地环境，不得引用具体 `test/scenarios/e2e/p0-*` 场景目录。
- `env-status.sh`：只读执行 `make dev-doctor`，用于 skill status 与人工巡检。
- `env-verify.sh`：执行 `make dev-doctor` 并检查 JSON summary 全 OK；可作为 `/scenario-run` 前置环境 gate。
- `env-cleanup.sh`：默认执行 `make dev-down` 并保留命名卷；只有显式 `--with-volumes` / `--reset` 时才执行 `DEV_RESET_FORCE=1 make dev-reset`。
- `env-redeploy.sh [deps|backend|frontend|all]`：`deps` 重跑 `make dev-up` + `make dev-doctor`；`backend` 执行 `cd backend && go build ./cmd/...` 后从 `deploy/dev-stack/.env` 重启 `go run ./backend/cmd/api`；`frontend` 执行 `pnpm --filter @easyinterview/frontend build` 后从同一 `.env` 重启 Vite dev server；`all` 顺序覆盖 deps/backend/frontend。当前 host-run 口径下 redeploy 不等于 Kind / Helm rollout，但必须让浏览器访问到当前代码和当前 env。

所有 env 脚本必须支持 `--dry-run`，输出将执行的 repo command，便于 skill 在不改变环境时解释下一步，也便于 unit/static gate 验证 Makefile 集成。脚本必须用 shell/Python 实现，不新增 `backend/cmd` / Go helper。

#### 6.2 根 Makefile 集成

根 `Makefile` 新增：

- `scenario-env-setup`
- `scenario-env-status`
- `scenario-env-verify`
- `scenario-env-cleanup`
- `scenario-env-redeploy`

这些 target 只委派 `test/scenarios/env-*.sh`，并透传 `ARGS` / `TARGET` 等显式参数；不得把业务场景 ID 或 `e2e/p0-100-real-provider-full-funnel-hybrid` 路径写死进 Makefile。

#### 6.3 skill 集成

更新 `.agent-skills/scenario-env/SKILL.md`：

- description 与 Usage 覆盖 setup / verify / cleanup / status / redeploy / rebuild。
- setup / verify / cleanup / status 优先调用 `test/scenarios/env-*.sh`，再退到 README 手动引导；不得从具体场景脚本提取环境 bootstrap。
- local integration 解释必须区分 Docker Compose 外部依赖、host-run backend/frontend command、repo-tracked scenario runner；redeploy backend/frontend/all 必须重启对应 host-run 进程，并在缺真实 secret 时报告具体 blocker。

更新 `.agent-skills/scenario-redeploy/SKILL.md`：把 `test/scenarios/env-redeploy.sh [component]` 作为当前 repo 首选入口，并声明当前 host-run redeploy 是 build + 重启 host-run 进程，不是 Kind / Helm / cluster rollout。

#### 6.4 场景 README 与 dev-stack README 对齐

更新 `test/scenarios/README.md`、`test/scenarios/e2e/README.md`、`deploy/dev-stack/README.md`：

- 明确 `env-*.sh` 是共享环境入口，独立于具体场景用例。
- 明确具体场景 `setup.sh` 只准备场景数据/输出目录，不能私有化共享环境 bootstrap。
- 明确最新 hybrid UAT / 前后端联调可先通过 env setup + redeploy 准备基础环境，再按 runbook 启动真实 backend/frontend 进程并由 Agent/人工验证目标场景。

#### 6.5 Phase 6 自检

- Red/green static contract：新增 `scripts/lint/scenario_env_contract_test.py`，断言 `env-*.sh` 存在、可执行、`bash -n` 通过、支持 `--dry-run`、不引用 `p0-*` 场景目录或 `backend/cmd` 场景 helper。
- Makefile dry-run：`make scenario-env-setup ARGS=--dry-run`、`make scenario-env-status ARGS=--dry-run`、`make scenario-env-verify ARGS=--dry-run`、`make scenario-env-cleanup ARGS="--dry-run --with-volumes"`、`make scenario-env-redeploy TARGET=backend ARGS=--dry-run` 均输出对应脚本命令。
- Skill contract：focused pytest 断言 `scenario-env` / `scenario-redeploy` skill 使用顶层 env scripts，并支持 redeploy/rebuild 意图。
- Live gate：执行 `test/scenarios/env-setup.sh`、`test/scenarios/env-verify.sh`、`test/scenarios/env-cleanup.sh`，证明环境可独立启动/验证/清理；若 Docker/端口/镜像阻塞，记录具体 blocker，不用具体场景 runner 代替。

### Phase 7: local raw output debug default revision

#### 7.1 dev/test config defaults

`config/dev.yaml` 与 `config/test.yaml` 必须默认 `ai.debugPrintRawOutput=true`；`config/config.yaml` 与 staging/prod overrides 不得把该默认扩大到非本地环境。根 `.env.example` 与 `deploy/dev-stack/.env.example` 必须将 `AI_DEBUG_PRINT_RAW_OUTPUT=true` 作为 local test/integration 默认值。

#### 7.2 P0.100 hybrid preflight

`E2E.P0.100` 的 `scripts/trigger.sh` 必须从 `deploy/dev-stack/.env` 校验 `AI_DEBUG_PRINT_RAW_OUTPUT=true`；缺失或非 true 时输出 `MANUAL_REQUIRED`，防止真实 provider hybrid 场景在无法调试 raw output 的状态下给出 false-green。

#### 7.3 Phase 7 自检

- Red/green config test：`go test ./backend/internal/platform/config -run TestRepoLocalConfigEnablesRawOutputDebugOnlyForLocalEnvironments -count=1`。
- Red/green scenario contract：`python3 -m pytest scripts/lint/scenario_env_contract_test.py -q -k real_provider_hybrid_uat_uses_dev_stack_env_as_single_source`。
- Live hybrid gate：`scenario-run -i E2E.P0.100` 在当前本地 `.env` 下产出 PASS，并在 redacted evidence 中保留 provider/profile/model/task-run 摘要，不复制 raw response。

### Phase 8: developer debug handoff revision

#### 8.1 调试摘要 helper

新增共享 helper `test/scenarios/_shared/scripts/local-dev-runtime.sh`，负责：

- 从 `deploy/dev-stack/.env` 读取非 secret 端口配置，推导 frontend dev URL、backend API base、Mailpit URL 与 MinIO Console URL。
- 将 backend/frontend host-run 日志固定写入 `.test-output/local-dev/backend.log` 与 `.test-output/local-dev/frontend.log`。
- 将启动进程组 PID 写入 `.test-output/local-dev/backend.pid` 与 `.test-output/local-dev/frontend.pid`。
- 输出 `tail -f` 与 `make dev-logs SERVICE=<name>` 调试命令；不得打印 `AI_PROVIDER_API_KEY`、auth secret、session cookie 或 magic-link token。

#### 8.2 redeploy 闭环语义

`test/scenarios/env-redeploy.sh backend|frontend|all` 必须在 build artifact gate 通过后：

- 停止旧 PID 文件指向的进程组，并 fallback 停止对应端口 listener。
- 使用 detached host-run process 重新启动 backend/frontend，让 Agent 启动的服务在命令结束后仍可由开发者访问。
- 等待对应端口可连接；失败时打印对应日志尾段并非零退出。
- 输出统一调试摘要。

`deps` target 仍只负责 Docker Compose 外部依赖和 `make dev-doctor`。

#### 8.3 setup/status/verify 可接管输出

`env-setup.sh` / `env-status.sh` / `env-verify.sh` 必须输出同一调试摘要；其中 status/verify 的 `make dev-doctor` JSON 保持 stdout，调试摘要走 stderr，避免破坏机器消费。

#### 8.4 Phase 8 自检

- Static contract：`python3 -m pytest scripts/lint/scenario_env_contract_test.py -q` 覆盖 helper、redeploy restart semantics、summary 输出和 README。
- Dry-run：`test/scenarios/env-redeploy.sh backend --dry-run`、`test/scenarios/env-setup.sh --dry-run`、`test/scenarios/env-status.sh --dry-run` 均解释即将输出调试信息。
- Live redeploy：`test/scenarios/env-redeploy.sh all` 后 `lsof` 证明 backend/frontend 监听当前端口；`.test-output/local-dev/*.log` 与 `.pid` 存在；重新触发 Mailpit 登录邮件，最新 magic-link 指向当前 `EMAIL_VERIFY_BASE_URL` 的 frontend `/auth/verify`。

## 5 验收标准

- spec [§6 验收标准](../../spec.md#6-验收标准) C-1 到 C-15 全部成立，证据贴入工作日志或当前 `.test-output/`。
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
| 未来组件没有 Dockerfile 或稳定 dev command，导致无法纳入 `make dev-up` | 默认不纳入 compose：对应组件先提供宿主机 dev command；只有确实需要 optional app service 时，组件 plan 才补齐 Dockerfile、健康检查与资源预算后声明受 `make dev-up` 覆盖 |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-05-27 | 1.13 | Developer debug handoff revision：`env-redeploy.sh backend|frontend|all` 从 build-only 修订为 build + 重启 host-run 进程，并输出 endpoint/log/PID/container log 调试入口。 | user feedback |
| 2026-05-27 | 1.12 | Raw debug local default revision：local dev/test 与本地真实联调默认开启 `AI_DEBUG_PRINT_RAW_OUTPUT=true`，P0.100 preflight 校验该开关，staging/prod 默认关闭。 | user feedback |
| 2026-05-27 | 1.11 | Single env source revision：`deploy/dev-stack/.env` 成为本地真实前后端联调唯一 env 来源，`.env.example` 补齐 auth secrets、frontend real mode、AI provider 与共享依赖 keys，禁止场景复制独立 `.env`。 | user feedback / BUG-0110 follow-up |
| 2026-05-27 | 1.10 | Environment lifecycle revision：把共享测试环境与本地前后端联调环境 lifecycle 抽到 `test/scenarios/env-*.sh`、根 `scenario-env-*` Make target 与 scenario skill 入口，支持独立 setup/status/verify/cleanup/redeploy。 | user objective / scenario-env independent lifecycle |
| 2026-05-26 | 1.9 | Mailpit revision：默认 dev-stack 新增 Mailpit，backend `EMAIL_PROVIDER=mailpit` 走 SMTP writer，manual UAT 账号入口回到真实 magic-link flow；`test/scenarios` 继续禁止新增 Go / `backend/cmd` 场景 helper。 | user feedback / manual UAT boundary fix |
| 2026-05-22 | 1.8 | L2 runtime remediation：修复 Postgres 18 命名卷挂载路径，`easyinterview-pg-data` 改挂 `/var/lib/postgresql` 以兼容官方镜像 PGDATA 布局，并增加旧卷布局 preflight。 | local-dev-stack/001 L2 code review |
| 2026-05-04 | 1.4 | L1 plan-review remediation：补齐当前强制的质量门禁分类，不改变已完成 dev stack 范围。 | historical-spec-implementation-review/001 |
| 2026-05-08 | 1.5 | 对齐 A3/B4 当前决策：默认 dev stack 删除未使用扩展依赖与 probe，仅保留普通 Postgres；未来需要时重新设计。 | ai-provider-and-model-routing/003 Phase 6 |
| 2026-05-08 | 1.6 | 按用户决策将默认 dev stack Postgres 镜像升级到 18，并同步迁移基线前提。 | local-dev-stack/001 post-pass revision |
| 2026-05-22 | 1.7 | 按方案 A 对齐本地部署与测试环境：compose 默认只管理外部依赖，backend/frontend 默认宿主机运行；`test/scenarios/` 默认本地 runner 验证，不再把 Kind / K8s / Helm 当作 P0 前提。 | local-dev-stack/001 post-pass revision |
