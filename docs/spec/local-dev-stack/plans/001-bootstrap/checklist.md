# Local Dev Stack Bootstrap Checklist

> **版本**: 1.27
> **状态**: completed
> **更新日期**: 2026-07-20

**关联计划**: [plan](./plan.md)

## Phase 1: docker-compose 与 init 脚本

- [x] 1.1 落地 `deploy/dev-stack/docker-compose.yaml`：默认最小外部依赖（Postgres / Redis / MinIO）+ 已显式接入的 optional 项目组件；backend/frontend 默认宿主机 dev command 运行；按 D-2 锁定镜像 tag、D-3 端口、D-4 network alias、D-7 命名卷；compose v2 schema；默认不预留也不启动 OTel / Grafana / Loki / Prometheus / AI provider；每个 compose service 配置容器级 healthcheck（≤5s 间隔、≥3 重试）
- [x] 1.2 落地 init 脚本：`init/minio/create-buckets.sh` 创建默认 bucket（幂等）；Postgres 默认不启用未使用扩展；不创建 Grafana / OTel / Loki / Prometheus provisioning
- [x] 1.3 顶层 `volumes:` 声明 3 个命名卷（pg-data / redis-data / minio-data），不使用 bind mount；Postgres 18 命名卷挂到 `/var/lib/postgresql`，保留官方镜像 `PGDATA=/var/lib/postgresql/18/docker`，不挂到 `/var/lib/postgresql/data`；`make dev-up` 启动前只读检测不兼容卷布局并给出显式 reset 指引，不自动删卷
- [x] 1.4 落地 `deploy/dev-stack/.env.example`：连接串 / bucket 名 / 依赖端口 / 项目组件 host port / auth secrets / frontend real mode / `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 示例默认值，字段名与 A4 secrets-and-config spec 对齐；`.env` 由根 `.gitignore` 忽略；`make dev-up` 首次运行时若 `.env` 不存在则从 `.env.example` 复制；`.env.example` 不写真实 AI key 或 auth secret
- [x] 1.5 Phase 1 自检：`docker compose up -d` 后 Postgres / Redis / MinIO 与已接入 optional 项目组件均 healthy；Postgres 内 `select 1` 返回 1 行（关闭 C-6）；`docker compose down`（无 `--volumes`）后命名卷保留

## Phase 2: Make targets 与生命周期语义

- [x] 2.1 接入 repo-scaffold 锁定的 `make dev-up` / `dev-down` 根入口：根 `Makefile` 递归调用 `deploy/dev-stack/Makefile` 真实实现，默认只管理外部依赖与已显式接入 optional app service；新增 `dev-doctor` / `dev-reset` / `dev-logs` / `dev-pull` phony target 并入根 `make help`
- [x] 2.2 实现 dev-up idempotent（C-3）：`docker compose ps` 全 running+healthy 时跳过启动，打印 `already healthy` 并 exit 0；重复执行不重启容器
- [x] 2.3 实现 dev-down 卷保留（C-4）：`docker compose down`（不带 `--volumes` / `--rmi`）；自检写入测试数据 → down → up → 数据仍在
- [x] 2.4 实现 dev-reset 显式清空（C-5）：交互式 `read -p` 确认；`DEV_RESET_FORCE=1` 跳过交互；确认后 `docker compose down --volumes` 删除 3 个命名卷；输入 `no` 时 abort 不删卷
- [x] 2.5 Phase 2 自检：第二次 `make dev-up` 含 `already healthy` 且 <5s；写入-down-up-读取链路数据保留；`DEV_RESET_FORCE=1 make dev-reset` 后 3 卷消失；`make dev-reset` stdin `no` 时不删卷

## Phase 3: dev-doctor 结构化健康检查

- [x] 3.1 落地 `deploy/dev-stack/scripts/dev-doctor.sh`（POSIX sh + jq，≤200 行）：输出 spec D-6 锁定的 JSON 结构（services 含 `type=dependency|app` + summary）；`summary.down==0 && summary.degraded==0` 时 exit 0；不得硬编码固定 7-service 口径
- [x] 3.2 实现 dependency/runtime probe：PG `pg_isready` + `select 1`；Redis set/get/del 一次；MinIO `mc ls` 默认 bucket；optional 项目 HTTP 组件查 `/healthz`，已声明 `/metrics` 的组件查 `/metrics` 非空；宿主机 backend/frontend 由对应 owner 的 dev command / runtime health 验证；启用 AIClient 的组件只校验真实 provider env 已注入，不调用真实 LLM。
- [x] 3.3 dev-up gate 接入（C-1）：`up` target 在 `docker compose up -d --wait` 后调用 dev-doctor；`summary.ok == total` 才 exit 0；否则输出 DOWN/DEGRADED 服务的最近 50 行 `docker logs` 尾段
- [x] 3.4 失败可观察（C-2）：构造 Postgres 5432 或任一已接入 optional 项目组件 host port 冲突复现路径；`make dev-up` 非 0 退出且 stderr 含冲突服务名 + 占用进程；`make dev-doctor` 对冲突服务报 `status=DOWN, reason="port conflict: ..."`，其它服务保持 OK
- [x] 3.5 Phase 3 自检：全员 OK 时 dev-doctor JSON 通过 schema 校验（3 个依赖名固定，项目组件来自 compose）且 exit 0；`docker stop redis-dev` 后报 DOWN/exit 1；缺 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 时启用 AIClient 的组件 fail-fast 且 dev-doctor 报缺真实 provider 配置；端口冲突复现路径日志贴入工作日志

## Phase 4: 指标日志 + 文档 + AC 收口

- [x] 4.1 应用 `/metrics` 与容器日志验证（C-7）：对已接入 compose 且声明 `/metrics` 的 optional 项目组件执行 curl 非空校验；当前没有 optional app service 时验证依赖容器日志；宿主机 backend/frontend 由对应 owner 的 dev command / runtime health 验证；不创建 OTLP smoke，不安装 OTel / Grafana / Loki / Prometheus。
- [x] 4.2 落地 `deploy/dev-stack/README.md`：服务表（name/image/port/credentials/volume）、optional 项目组件表（component/service/host port/health/metrics）与宿主机运行边界、`make dev-*` 命令清单、AI provider 配置说明（非测试本地 app run 使用真实 provider，stub 仅单测）、常见故障排查、与 `test/scenarios/` 真实 API/UI E2E 契约及根 `make test` 单测边界说明、资源占用提示与默认依赖镜像下载体积估算（< 1.5GB），并声明默认本地栈不包含 OTel / Grafana / Loki / Prometheus / AI provider
- [x] 4.3 在 `deploy/dev-stack/README.md` 声明最低 docker engine（24+）与 compose plugin（v2.20+）版本；本 plan 不创建或修改 A5 远端 CI workflow，当前单人开发阶段不在 CI 拉起 dev stack
- [x] 4.4 A2 executable gate handoff（C-8）：依次复跑 C-1 / C-2 / C-3 / C-4 / C-5 / C-6 / C-7 / C-9 八项 AC；执行证据贴入工作日志；不修改 engineering-roadmap parent checklist
- [x] 4.5 文档收口：`deploy/dev-stack/README.md` Header 完整；plans/INDEX.md 把本 plan 切到 completed 段；`/sync-doc-index --check` 通过

## Phase 5: Mailpit 本地邮件 sink revision

- [x] 5.1 在 `deploy/dev-stack/docker-compose.yaml` 默认依赖中新增 `mailpit-dev`，锁定 `axllent/mailpit:v1.30.0`，暴露 8025 Web UI / 1025 SMTP，接入 dependency label 与 `/readyz` healthcheck；`DEPENDENCY_SERVICES`、端口冲突扫描、pull/logs 自动覆盖 Mailpit；验证：`docker compose config --quiet` + `make -C deploy/dev-stack -n up`
  <!-- verified: 2026-05-26 command="docker compose -f deploy/dev-stack/docker-compose.yaml --project-directory deploy/dev-stack config --quiet && make -C deploy/dev-stack -n up" evidence="compose config valid; dry-run up targets postgres-dev redis-dev minio-dev mailpit-dev and scans MAILPIT_WEB_HOST_PORT/MAILPIT_SMTP_HOST_PORT" -->
- [x] 5.2 扩展 `dev-doctor.sh`：Mailpit 容器状态、`GET /readyz`、可用时 SMTP host port probe 均纳入 JSON summary，默认依赖 OK 数从 3 调整为 4；验证：`bash -n deploy/dev-stack/scripts/dev-doctor.sh` + live `make dev-doctor`
  <!-- verified: 2026-05-26 command="bash -n deploy/dev-stack/scripts/dev-doctor.sh && make dev-doctor" evidence="dev-doctor output summary ok=4 degraded=0 down=0 total=4 with mailpit-dev status OK; script remains 192 lines" -->
- [x] 5.3 后端 auth 邮件投递接入 Mailpit：`EMAIL_PROVIDER=mailpit` 时 `cmd/api` 注册 SMTP `DeliveryWriter`，从 `auth_challenges` 查收件人、从 transient delivery secret store 取 token，`email_dispatch` payload 仍保持 redaction；验证：focused auth / cmd/api tests
  <!-- verified: 2026-05-26 command="go test ./backend/internal/auth -run 'TestSMTPDeliveryWriter|TestSQLChallengeEmailLookup|TestEmailDispatchHandler_PayloadRedaction' -count=1 && go test ./backend/cmd/api -run 'TestBuildAuthService(UsesMailpitDeliveryWriterWhenConfigured|RejectsEmptyAuthSecrets)' -count=1 && go build ./backend/cmd/api" evidence="auth SMTP writer, SQL challenge lookup, email_dispatch redaction, cmd/api Mailpit DI, and backend entry build all pass" -->
- [x] 5.4 A4 env/config 字典补齐 `EMAIL_SMTP_HOST` / `EMAIL_SMTP_PORT` / `EMAIL_FROM_ADDRESS` / `EMAIL_VERIFY_BASE_URL`，root `.env.example` 与 dev-stack env 模板同步；验证：focused config test + `make lint-config`
  <!-- verified: 2026-05-26 command="go test ./backend/internal/platform/config -run TestDefaultEmailDictionaryIncludesMailpitSMTPBindings -count=1 && make lint-config" evidence="A4 default env bindings include Mailpit SMTP keys; lint-config reports 32 env keys in .env.example and spec with no leaks" -->
- [x] 5.5 hybrid UAT 账号入口改为 synthetic 邮箱 + Mailpit 6 位 code，删除直接 session bootstrap helper，保留 no-backend-cmd 与 test/scenarios no-Go negative gate；验证：`test ! -d backend/cmd/devsession && test ! -d backend/internal/devsession && test -z "$(find test/scenarios -name '*.go' -type f -print -quit)"`
  <!-- verified: 2026-05-26 evidence="no devsession backend cmd/internal package, no Go files under test/scenarios, and no direct-session bootstrap helper" -->
- [x] 5.6 Mailpit live gate：`make dev-up && make dev-doctor` 输出 Postgres / Redis / MinIO / Mailpit 四个 dependency OK；若无法拉取镜像或本机端口占用，记录 blocker 与复现输出
  <!-- verified: 2026-05-26 command="make dev-up; make dev-doctor" evidence="first docker pull attempt hit registry EOF; retry succeeded. make dev-up completed and standalone make dev-doctor returned postgres-dev/redis-dev/minio-dev/mailpit-dev all OK with summary ok=4 degraded=0 down=0 total=4" -->

## Phase 6: 独立 scenario/local integration environment lifecycle revision

- [x] 6.1 新增 `test/scenarios/env-setup.sh` / `env-status.sh` / `env-verify.sh` / `env-cleanup.sh` / `env-redeploy.sh` 顶层环境入口：setup/status/verify/cleanup/redeploy 独立于具体 `p0-*` 场景目录；支持 `--dry-run`；cleanup 默认保留卷、显式 `--with-volumes` 才 reset；redeploy 支持 `deps|backend|frontend|all`；验证：focused static contract red/green + `bash -n` + dry-run 输出。
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="red failed before env scripts existed; green passed 3 tests covering top-level scripts, bash -n, --dry-run, no p0/manual coupling, and redeploy host-run artifact boundary" -->
- [x] 6.2 根 `Makefile` 集成 `scenario-env-setup` / `scenario-env-status` / `scenario-env-verify` / `scenario-env-cleanup` / `scenario-env-redeploy`，只委派顶层 env scripts 并透传 `ARGS` / `TARGET`；验证：`make scenario-env-* ARGS=--dry-run` 与 `TARGET=backend` dry-run。
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="4 tests passed, including root Makefile scenario-env-* target and dry-run delegation coverage" -->
- [x] 6.3 更新 `.agent-skills/scenario-env/SKILL.md` 与 `.agent-skills/scenario-redeploy/SKILL.md`，让 skill 根据用户意图优先调用顶层 env scripts，覆盖 setup/verify/status/cleanup/rebuild/redeploy，并明确 host-run frontend/backend 边界；验证：focused skill contract pytest。
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="5 tests passed, including scenario-env/scenario-redeploy top-level env script and host-run redeploy contract coverage" -->
- [x] 6.4 更新 `test/scenarios/README.md`、`deploy/dev-stack/README.md`，说明共享环境入口与真实 API/UI 验收、本地联调 runbook 的边界；验证：docs contract + `make docs-check`。
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q && make docs-check" evidence="6 scenario env contract tests passed; docs-check reported zero Header/INDEX drift and link checks OK" -->
- [x] 6.5 Phase 6 live gate：执行 `test/scenarios/env-setup.sh`、`test/scenarios/env-verify.sh`、`test/scenarios/env-cleanup.sh`，证明环境可独立启动/验证/清理；若 Docker/端口/镜像阻塞，记录 blocker 与输出，不用具体场景 runner 代替。
  <!-- verified: 2026-05-27 command="test/scenarios/env-setup.sh && test/scenarios/env-verify.sh && test/scenarios/env-cleanup.sh" evidence="setup reused already healthy dev-stack; verify returned postgres-dev/redis-dev/minio-dev/mailpit-dev OK with summary ok=4 degraded=0 down=0 total=4; cleanup stopped containers and cleaned up easyinterview-dev network while preserving named volumes" -->

## Phase 7: local raw output debug default revision

- [x] 7.1 历史本地 raw debug 默认已由 Phase 15 独立 NDJSON capture/path 合同取代；当前配置与 preflight 不保留 stderr 输出入口，代码层阶段收口仍由根 `make test` 承接。

## Phase 8: developer debug handoff revision

- [x] 8.1 新增 `test/scenarios/_shared/scripts/local-dev-runtime.sh`：从 `deploy/dev-stack/.env` 推导 frontend/backend/Mailpit/MinIO 地址，统一 `.test-output/local-dev/{backend,frontend}.log` 与 `.pid`，输出 `tail -f` 和 `make dev-logs SERVICE=<name>` 调试入口；不得打印 secret、cookie 或 raw email code。
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="contract covers local-dev-runtime helper, start_new_session detached launch, endpoint summary, backend/frontend log paths, and no p0/manual coupling" -->
- [x] 8.2 修订 `test/scenarios/env-redeploy.sh backend|frontend|all`：build artifact gate 通过后重启对应 host-run backend/frontend 进程，等待端口可连接，失败时打印日志尾段；`deps` 仍只管理 Docker Compose 外部依赖。
  <!-- verified: 2026-05-27 command="test/scenarios/env-redeploy.sh all" evidence="dev-stack dependencies OK, backend go build passed and backend listened on 127.0.0.1:8080, frontend build passed and Vite listened on 127.0.0.1:5173" -->
- [x] 8.3 修订 `env-setup.sh` / `env-status.sh` / `env-verify.sh`：创建、巡检或验证共享环境后输出同一调试摘要；status/verify 保持 `dev-doctor` JSON stdout，调试摘要走 stderr。
  <!-- verified: 2026-05-27 command="test/scenarios/env-setup.sh --dry-run; test/scenarios/env-status.sh --dry-run" evidence="setup explains dev-up/dev-doctor and debug summary; status keeps dry-run make dev-doctor on stdout and debug summary notice on stderr" -->
- [x] 8.4 文档与 skill 对齐：更新 `deploy/dev-stack/README.md`、`test/scenarios/README.md`、`scenario-env` / `scenario-redeploy` skill，明确 redeploy 是 build + restart，且输出可接管地址/日志/PID。
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="contract requires dev-stack debug log paths, env-redeploy command, scenario README redeploy restart semantics, and skill host-run restart wording" -->
- [x] 8.5 Phase 8 自检：scenario env contract pytest、dry-run、`env-redeploy.sh all` live restart、端口监听、日志/PID 存在、Mailpit 最新邮件为 code-only 且本地 frontend origin / CORS 来源一致。
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q; test/scenarios/env-redeploy.sh backend --dry-run; test/scenarios/env-setup.sh --dry-run; test/scenarios/env-status.sh --dry-run; test/scenarios/env-redeploy.sh all; lsof -nP -iTCP:8080 -iTCP:5173 -sTCP:LISTEN; curl -X POST http://127.0.0.1:8080/api/v1/auth/email/start ..." evidence="12 contract tests passed; redeploy summary printed endpoints/logs/PIDs; backend/frontend listeners survived after command exit; latest Mailpit email contains frontend /auth/verify callback and not backend verify API" -->

## Phase 9: host-run backend loopback bind revision

- [x] 9.1 Red contract：`scripts/lint/scenario_env_contract_test.py` 覆盖 `local-dev-runtime.sh` 的 `backend_listen_addr` / `APP_LISTEN_ADDR` 导出契约；implementation 前 focused pytest 必须失败。
  <!-- verified: 2026-06-15 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q -k redeploy_script_documents_host_run_artifact_boundary" evidence="red failed before helper implementation: expected backend_listen_addr() in local-dev-runtime.sh" -->
- [x] 9.2 Runtime fix：`restart_backend_runtime` 将通配 `APP_LISTEN_ADDR` 收敛为 `127.0.0.1:${API_HOST_PORT:-8080}` 后再启动 `go run ./backend/cmd/api`，并保留显式具体监听地址。
  <!-- verified: 2026-06-15 evidence="local-dev-runtime.sh now derives backend_listen_addr, exports APP_LISTEN_ADDR before go run, and logs the effective listen address" -->
- [x] 9.3 Docs/runbook：`deploy/dev-stack/README.md` 与 local-dev-stack spec/plan/checklist 说明 loopback host-run backend 契约和 bridge listener regression。
  <!-- verified: 2026-06-15 evidence="local-dev-stack spec 1.20 adds D-15/C-16; plan 1.14 adds Phase 9; dev-stack README 1.7 documents loopback redeploy behavior" -->
- [x] 9.4 Green/static gates：`python3 -m pytest scripts/lint/scenario_env_contract_test.py -q`、`bash -n test/scenarios/_shared/scripts/local-dev-runtime.sh test/scenarios/env-redeploy.sh` 通过。
  <!-- verified: 2026-06-15 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q; bash -n test/scenarios/_shared/scripts/local-dev-runtime.sh; bash -n test/scenarios/env-redeploy.sh" evidence="12 scenario env contract tests passed; both shell scripts parsed successfully; bash emitted only existing locale warning" -->
- [x] 9.5 Live/user regression gate：存在无关 8080 bridge listener 时 `test/scenarios/env-redeploy.sh backend` 成功；`/api/v1/runtime-config` 返回 200；新 synthetic 首次登录用户 `GET /api/v1/resumes` 返回 200 empty list。
  <!-- verified: 2026-06-15 command="test/scenarios/env-redeploy.sh backend; curl http://127.0.0.1:8080/api/v1/runtime-config; node first-login smoke; agent-browser /resume-versions smoke" evidence="netstat still showed unrelated 172.18.0.6:8080 listener; redeploy started backend with APP_LISTEN_ADDR=127.0.0.1:8080; runtime-config 200; new synthetic user resumes before and after profile setup returned 200 empty list; browser page showed resume empty state and in-page fetch /api/v1/resumes returned status 200 itemCount 0 errorCode null" -->

## Phase 10: one-click reset/redeploy Make target revision

- [x] 10.1 Red contract：`scripts/lint/scenario_env_contract_test.py` 覆盖 `scenario-env-reset-redeploy` 的 Makefile target、phony/help、script reuse 和 dry-run 顺序；implementation 前 focused pytest 必须失败。
  <!-- verified: 2026-07-09 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q -k reset_redeploy" evidence="red failed before Makefile target existed; focused gate passed after adding the target and dry-run order assertion" -->
- [x] 10.2 Makefile implementation：根 `Makefile` 新增 `scenario-env-reset-redeploy`，依次调用 `env-cleanup.sh --with-volumes`、`env-setup.sh --with-migrations`、`env-redeploy.sh all`、`env-verify.sh`，并支持 `ARGS=--dry-run` 无副作用预览。
  <!-- verified: 2026-07-09 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q -k reset_redeploy" evidence="Makefile target uses SCENARIO_ENV_* variables and dry-run output shows reset, setup/migrations, redeploy backend/frontend, final verify order" -->
- [x] 10.3 Docs/runbook：`deploy/dev-stack/README.md`、`test/scenarios/README.md` 说明一键清数据重编译重部署入口，并区分普通重启 `scenario-env-redeploy TARGET=all`。
  <!-- verified: 2026-07-09 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q -k \"reset_redeploy or scenario_docs_describe\"" evidence="focused contract verifies Makefile target plus dev-stack/scenario/e2e README reset-redeploy wording" -->
- [x] 10.4 Phase 10 self-check：focused contract pytest、`make scenario-env-reset-redeploy ARGS=--dry-run`、`make docs-check`、`sync-doc-index --check`、`git diff --check` 全部通过。
  <!-- verified: 2026-07-09 command="make scenario-env-reset-redeploy ARGS=--dry-run; python3 -m pytest scripts/lint/scenario_env_contract_test.py -q; python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check; make docs-check; git diff --check" evidence="dry-run previewed reset/setup-migrations/redeploy/verify without changing environment; 13 scenario env contract tests passed; docs/index/link gates and whitespace gate passed" -->

## Phase 11: shared scenario script inventory cleanup

- [x] 11.1 新增 README shared-script inventory contract，先证明 `common.sh` / `image-cache.sh` 引用指向不存在文件，再删除两处路径、无效首次使用命令和 fallback 说明；验证 focused/full scenario contract pytest、shell inventory、owner contexts 与 docs/diff/pruning gates。
  <!-- verified: 2026-07-10 method=scenario-shared-script-inventory-cleanup evidence="Focused RED named only missing common.sh and image-cache.sh. Removed both README entries, the nonexistent pull command and fallback prose; documented the three real shared helpers. Scenario env/script contracts pass 18 tests, eight shell entrypoints parse, reset/redeploy dry-run and owner contexts PASS." -->

## Phase 12: optional full-container local deployment

- [x] 12.1 Red contract：扩展 `scripts/lint/scenario_env_contract_test.py`，覆盖根/子 Makefile `dev-container-*`、同一 Compose `full-container` profile、migrations/backend/frontend、默认 10800/10801、Dockerfile/proxy/docs/skill 和默认 host-run 非回归；先运行 focused pytest 并确认因实现缺失而失败。
  <!-- verified: 2026-07-16 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q -k optional_full_container_profile_contract" evidence="RED failed at missing root dev-container-up target before implementation" -->
- [x] 12.2 Container runtime：新增 backend/frontend 多阶段 Dockerfile 与 frontend SPA/API proxy 配置；Compose 接入一次性 migrations、backend、frontend，使用真实依赖 service name、healthcheck、doctor app labels、loopback host port 和 AI/auth secret fail-fast；验证 focused contract、`docker compose config --quiet` 与两张镜像实际 build 通过。
  <!-- verified: 2026-07-16 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q -k optional_full_container_runtime_contract; docker compose -f deploy/dev-stack/docker-compose.yaml --project-directory deploy/dev-stack --profile full-container config --quiet; docker compose ... build backend-dev frontend-dev" evidence="runtime contract and compose config pass; backend API/migrate and frontend TypeScript/Vite/nginx images built successfully for arm64" -->
- [x] 12.3 Lifecycle/docs/skill：根与 `deploy/dev-stack` Makefile 实现 `dev-container-up/down/doctor/logs`，`.env.example` 增加 `FULL_CONTAINER_FRONTEND_HOST_PORT=10800` / `FULL_CONTAINER_API_HOST_PORT=10801`；更新 dev-stack/scenario README 与 scenario-env/redeploy skill；focused lifecycle contract pytest、根/子 Make help 和 `git diff --check` 已通过。
- [x] 12.4 Static/regression gate：root `make test`（Python 561 passed / 4481 subtests、backend Go、frontend 126 files / 1004 tests）、`make build`、`make docs-check`、context validator、doc index check、full-container compose config 和 `git diff --check` 全部通过；代码 gate 与真实环境验收分层报告。
- [x] 12.5 Live/Chrome gate：`make dev-container-up` 已部署当前工作树，migrations 成功，doctor 6/6 OK，10800 frontend 与 10801 runtime-config 均返回 200；Chrome 在 10800 使用 E2E.P0.101 标准 synthetic 邮箱完成验证码登录、资料补全、简历 AI 解析、JD/面试计划、真实 AI 问答与报告生成，未使用 mock/interception且 console warning/error 为 0；脱敏截图保存于 `.test-output/full-container-chrome/01-auth-profile-home.png`、`02-resume-ready.png`、`03-live-practice.png`、`04-report-complete.png`，报告页与环境保持运行供接管。

## Phase 13: Mailpit / external SMTP provider switch

- [x] 13.1 RED: scenario env contract 要求 root/dev-stack env 模板与 Compose 包含 provider、host/port、username/password、TLS mode/from 透传，且外部 SMTP host 不被 `mailpit-dev` 硬编码覆盖。
  <!-- verified: 2026-07-16 method=focused-red evidence="new scenario env contract failed before x-backend-environment exposed provider-selectable SMTP variables" -->
- [x] 13.2 GREEN: 更新 env 模板、Compose、Make target 与 runbook；Mailpit 仅切 provider 即由 full-container 自动使用 `mailpit-dev:1025`，SMTP 使用用户 `.env` endpoint/secret；切换前停止仓库 PID 文件管理的 host-run app，避免 runner 竞争；focused contract、compose config、`make lint-config` 通过。
  <!-- verified: 2026-07-16 method=focused-green evidence="scenario env contract 10/10, compose config, env dictionary, terminology and secret scan pass" -->
- [x] 13.3 LIVE: full-container 先以 Mailpit 完成真实发码/收码，再以用户 `.env` 标准 SMTP 完成 TLS/auth/实发；检查 backend/doctor/compose evidence 不包含 secret、完整收件人或 raw code。MVP 只运行一个 active backend 实例。
  <!-- verified: 2026-07-16 evidence="dev-container-up stopped repo-managed host-run PID files and reported 6/6 healthy; EMAIL_PROVIDER=mailpit alone resolved mailpit-dev:1025 and delivered in one attempt; fresh external SMTP service accepted authenticated implicit-TLS delivery and application job succeeded once; user confirmed EMAIL_FROM_ADDRESS inbox received EasyInterview sign-in code; redacted artifacts contain no recipient, code, or credential" -->

## Phase 14: reuse existing Redis for delivery secrets

- [x] 14.1 RED/GREEN-CONTRACT: scenario env contract 先拒绝当前单 backend/无 Redis client 口径，再证明 Compose 仅有既有 `redis-dev`，full-container backend 使用 `redis://redis-dev:6379/0`，不新增 service/network/volume/env key。
  <!-- verified: 2026-07-16 method=focused-red-green evidence="focused pytest first failed because runbook lacked shared Redis delivery-secret contract, then passed after README alignment; Compose asserts exactly one redis-dev and canonical backend REDIS_URL" -->
- [x] 14.2 DOC/WIRING: dev-stack README、spec/plan/checklist 与 backend-auth Phase 12 一致；`dev-container-up` 可继续停止 host-run app，但邮件投递正确性不再依赖该动作。
  <!-- verified: 2026-07-16 evidence="dev-stack README v1.9 documents encrypted shared delivery secret, 5m TTL, host/full-container REDIS_URL and runner-stop non-dependency" -->
- [x] 14.3 LIVE/REGRESSION: real Redis cross-client integration、full-container Mailpit/SMTP live、doctor 6/6、Compose config、scenario contract、docs/index/diff gates 全绿。
  <!-- verified: 2026-07-16 evidence="real Redis cross-client integration PASS; Mailpit Chrome login/profile PASS; SMTP job succeeded once; doctor 6/6; Compose/scenario contract and root make test/build/lint-config/docs/context/index/diff gates PASS" -->
- [x] 14.4 PID-OWNERSHIP-REMEDIATION: RED/GREEN pytest 使用真实无关子进程与陈旧 pidfile，证明 `_stop_host_runtimes` 在命令不匹配时不发送 TERM/KILL、只清理 pidfile；匹配逻辑仅接受当前 repo-managed backend/frontend 命令，scenario contract 与 shell/Make gates 通过。
  <!-- verified: 2026-07-16 method=tdd commands="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q; make test" evidence="RED killed unrelated sleep process; GREEN 13 scenario-env contract tests PASS for both stale-unowned preservation and owned-backend termination; root Python 566 tests/4481 subtests PASS" -->

## Phase 15: local AI raw I/O capture and single-runner guard

- [x] 15.1 RED-CONTRACT: scenario env contract 要求新 capture/path keys、旧 stderr key 零引用、resolved/realpath P0.099 raw path 不进入 evidence、backend-only full-container host bind，并要求已登记/手工 host 与 container 同 role 互斥。<!-- verified: 2026-07-16 method=scenario-env-focused-red evidence="contract tests fail on absent new keys/backend-only bind/raw hardening and host-container role mutual exclusion while retaining never-kill-unowned rule" -->
- [x] 15.2 GREEN-ENV: 同步 dev-stack `.env.example` / README / Compose / preflight，默认 `AI_DEBUG_CAPTURE_RAW_IO=true`、path 为 `.test-output/local-dev/ai-raw.ndjson`；预建/收紧 host raw directory，container recreate 后文件仍在，raw 不进入 backend log 或场景 evidence。<!-- verified: 2026-07-16 method=raw-env-green evidence="21 scenario contract tests PASS for new env keys, backend-only rw bind, ConfigDir-parent/realpath/path-mode guards and evidence exclusion; Compose config and shell syntax PASS in the implementation run" -->
- [x] 15.3 SINGLE-RUNNER: `env-redeploy.sh backend|frontend|all` 启动 host role 前停止对应 `backend-dev` / `frontend-dev` app service并保留依赖/卷；`env-verify.sh` 结合 repo PID、bounded process/listener inspection 和 Compose state 对同 role host/container 并存 fail closed，且不杀手工/无关进程。<!-- verified: 2026-07-16 method=single-runner-green evidence="scenario contract uses real unrelated processes/stale pidfiles to prove command plus repo-cwd ownership before TERM/KILL, and host/container same-role conflict detection preserves manual/unowned listeners" -->
- [x] 15.4 REGRESSION: focused scenario-env contract、shell syntax、dry-run/isolated process tests、A4 config gates、P0.099 preflight negative cases和旧 key current-scope zero-reference 通过；BDD 不适用。<!-- verified: 2026-07-16 method=raw-env-regression evidence="21 pytest cases, bash -n, Compose config --quiet, backend redeploy dry-run, platform/config and lint-config PASS; no E2E run or PASS is claimed" -->
- [x] 15.5 DOCTOR-ORDER-REGRESSION: RED dry-run contract 复现 `all` 在 app role 切换前先执行 dependency doctor；GREEN 改为在 doctor 前 `docker compose rm -sf backend-dev frontend-dev`，保留依赖/卷，并以真实 `scenario-env-redeploy TARGET=all` + `scenario-env-verify` 证明 backend/frontend build、监听与 4/4 dependency readiness。<!-- verified: 2026-07-16 method=tdd-live commands="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q; make scenario-env-redeploy TARGET=all; make scenario-env-verify" evidence="RED failed because rm-before-doctor order was absent; GREEN 22 tests PASS; stopped app records removed; backend 8080 and frontend 5173 started; dependency doctor 4/4 OK" -->

## Phase 16: unified config-owned local app ports

- [x] 16.1 RED-CONTRACT: scenario env contract 将 host-run/full-container 宿主机默认入口锁为 frontend 10900/backend 10901，覆盖 A4 env/YAML、Compose external mapping、Vite fallback 与 lifecycle fallback；实施前 focused pytest 因旧 5173/8080/10800/10801 默认而失败。<!-- verified: 2026-07-16 method=focused-red evidence="three focused contracts failed on old host-run env, full-container mapping and lifecycle docs before implementation" -->
- [x] 16.2 CONFIG/RUNTIME: root/dev-stack env、canonical YAML、Compose host mapping/label/email origin、Vite、host runtime、P0.099/P0.101 与 Playwright defaults 已统一；容器内部 8080 不变。<!-- verified: 2026-07-16 method=static-green evidence="runtime sources use 10900/10901 external defaults while Compose mapping right side, Dockerfile/nginx and container health retain internal 8080" -->
- [x] 16.3 SKILL-BOUNDARY: `scenario-env` / `scenario-redeploy` Skill 不包含具体 app 端口，只按 README、environment config 与 lifecycle command 输出定位 endpoint；focused contract 拒绝 Skill 重新耦合 10800/10801/10900/10901。<!-- verified: 2026-07-16 method=user-correction evidence="skills contain no app port literals and scenario contract verifies config-owned endpoint consumption" -->
- [x] 16.4 LIVE/CHROME: scenario/config/Compose/shell/build gates全绿；真实 `scenario-env-redeploy TARGET=all` + verify 后 frontend/backend 分别监听 10900/10901，runtime-config 200；Chrome 使用脚本输出的 frontend endpoint 完成中文 auth gate 与失败报告恢复验收。<!-- verified: 2026-07-16 method=live-chrome evidence="scenario env focused gates and build passed; host-run frontend/backend served 10900/10901 with runtime-config 200; Chrome verified localized auth gate, failed-report recovery, generating progress plus conversation, and ready report without mock/interception" -->

## Phase 17: host Mailpit SMTP route fail-fast

- [x] 17.1 RED-CONTRACT: provider 为 Mailpit、SMTP host 为 loopback 且 `EMAIL_SMTP_PORT` 与 `MAILPIT_SMTP_HOST_PORT` 不一致时，focused contract 要求 backend 进程启动前失败。<!-- verified: 2026-07-16 method=focused-red evidence="runtime contract reproduced the host mapping mismatch that left Mailpit empty and required a pre-start rejection" -->
- [x] 17.2 GREEN-RUNTIME: `local-dev-runtime.sh` 的 `assert_host_mailpit_smtp_route` 校验动态有效 host mapping；只报告字段名/端口，不输出 secret；full-container internal Mailpit 与 external SMTP 不受影响。<!-- verified: 2026-07-16 method=focused-green evidence="25 scenario_env_contract pytest cases passed; guard is called after env load and before host backend start; no port or email business configuration was added to Skills" -->
- [x] 17.3 LIVE-MAILPIT: 同步 host-run endpoint 后 redeploy backend，发起 fresh email-code challenge，Mailpit 收到一封新邮件；证据不记录完整邮箱或验证码。<!-- verified: 2026-07-16 method=chrome-live evidence="Mailpit mailbox changed from empty to one fresh code-only message and the matching browser login completed against the real backend" -->
- [x] 17.4 REGRESSION/DOCS: focused pytest、shell syntax、root gates、docs/context/INDEX/diff 全绿；BDD 不适用。<!-- verified: 2026-07-16 method=full-regression evidence="25 scenario environment contract tests and shell syntax PASS; live Mailpit delivery + Chrome login PASS; make test/build, context/docs/index and diff gates PASS" -->

## Phase 18: scenario cleanup host runtime remediation

- [x] 18.1 RED-CONTRACT: `scripts/lint/scenario_env_contract_test.py` 复现标准 cleanup 只调用 `make dev-down`、遗漏 repo-managed backend/frontend，并要求默认与 `--with-volumes` 两条路径都先调用共享 host runtime stop helper。<!-- verified: 2026-07-20 method=focused-red command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q -k scenario_cleanup_stops_repo_owned_host_runtimes_before_dependencies" evidence="1 failed, 25 deselected because env-cleanup.sh did not source the shared runtime helper" -->
- [x] 18.2 GREEN-RUNTIME: `local-dev-runtime.sh` 暴露组合停止 backend/frontend 的 helper；`env-cleanup.sh` 复用该 helper 后再执行 `make dev-down` 或显式 reset，沿用 PID/命令/repo cwd ownership 检查，不终止手工/无关进程。<!-- verified: 2026-07-20 method=focused-green command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q -k 'scenario_cleanup_stops_repo_owned_host_runtimes_before_dependencies or stop_host_runtimes'" evidence="3 passed, including cleanup ordering plus owned-process termination and unowned-process preservation" -->
- [x] 18.3 SAFETY/LIVE: focused contract、shell syntax、dry-run 顺序与隔离进程测试通过；真实标准 cleanup 后 repo-managed backend/frontend、pidfile、Compose 依赖均消失，默认命名卷保留。<!-- verified: 2026-07-20 method=tdd-live commands="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q; bash -n test/scenarios/env-cleanup.sh test/scenarios/_shared/scripts/local-dev-runtime.sh; test/scenarios/env-setup.sh; test/scenarios/env-redeploy.sh all; test/scenarios/env-cleanup.sh" evidence="27 contract tests passed; cleanup closed repo-owned 10900/10901 listeners, removed both pidfiles and all easyinterview containers, retained three named volumes; unrelated Ferry kubectl listener on 8025 remained untouched" -->
- [x] 18.4 DOC/REGRESSION: dev-stack/scenario runbook、Spec/plan/checklist/History 与 Skill cleanup 语义一致；context、Header/INDEX、docs links、根 `make test` 与 `git diff --check` 通过；BDD 不适用。<!-- verified: 2026-07-20 method=full-regression evidence="28 scenario environment contracts, context validator, Header/INDEX, docs links, Spec ID and diff gates passed; root make test rerun passed Python 624 tests/4628 subtests, Go all packages and frontend 137 files/1126 tests after one unrelated HomeRecentMocks timing retry passed focused 14/14" -->
