# Local Dev Stack Bootstrap Checklist

> **版本**: 1.19
> **状态**: completed
> **更新日期**: 2026-07-10

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
  <!-- verified: 2026-05-26 command="test ! -d backend/cmd/devsession && test ! -d backend/internal/devsession && test -z \"$(find test/scenarios -name '*.go' -type f -print -quit)\" && test ! -e test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/bootstrap_account.py" evidence="no devsession backend cmd/internal package, no Go files under test/scenarios, and no direct-session bootstrap_account.py helper" -->
- [x] 5.6 Mailpit live gate：`make dev-up && make dev-doctor` 输出 Postgres / Redis / MinIO / Mailpit 四个 dependency OK；若无法拉取镜像或本机端口占用，记录 blocker 与复现输出
  <!-- verified: 2026-05-26 command="make dev-up; make dev-doctor" evidence="first docker pull attempt hit registry EOF; retry succeeded. make dev-up completed and standalone make dev-doctor returned postgres-dev/redis-dev/minio-dev/mailpit-dev all OK with summary ok=4 degraded=0 down=0 total=4" -->

## Phase 6: 独立 scenario/local integration environment lifecycle revision

- [x] 6.1 新增 `test/scenarios/env-setup.sh` / `env-status.sh` / `env-verify.sh` / `env-cleanup.sh` / `env-redeploy.sh` 顶层环境入口：setup/status/verify/cleanup/redeploy 独立于具体 `p0-*` 场景目录；支持 `--dry-run`；cleanup 默认保留卷、显式 `--with-volumes` 才 reset；redeploy 支持 `deps|backend|frontend|all`；验证：focused static contract red/green + `bash -n` + dry-run 输出。
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="red failed before env scripts existed; green passed 3 tests covering top-level scripts, bash -n, --dry-run, no p0/manual coupling, and redeploy host-run artifact boundary" -->
- [x] 6.2 根 `Makefile` 集成 `scenario-env-setup` / `scenario-env-status` / `scenario-env-verify` / `scenario-env-cleanup` / `scenario-env-redeploy`，只委派顶层 env scripts 并透传 `ARGS` / `TARGET`；验证：`make scenario-env-* ARGS=--dry-run` 与 `TARGET=backend` dry-run。
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="4 tests passed, including root Makefile scenario-env-* target and dry-run delegation coverage" -->
- [x] 6.3 更新 `.agent-skills/scenario-env/SKILL.md` 与 `.agent-skills/scenario-redeploy/SKILL.md`，让 skill 根据用户意图优先调用顶层 env scripts，覆盖 setup/verify/status/cleanup/rebuild/redeploy，并明确 host-run frontend/backend 边界；验证：focused skill contract pytest。
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="5 tests passed, including scenario-env/scenario-redeploy top-level env script and host-run redeploy contract coverage" -->
- [x] 6.4 更新 `test/scenarios/README.md`、`test/scenarios/e2e/README.md`、`deploy/dev-stack/README.md`，说明共享环境入口与具体场景 runner、hybrid UAT / 本地联调 runbook 的边界；验证：docs contract pytest + `make docs-check`。
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q && make docs-check" evidence="6 scenario env contract tests passed; docs-check reported zero Header/INDEX drift and link checks OK" -->
- [x] 6.5 Phase 6 live gate：执行 `test/scenarios/env-setup.sh`、`test/scenarios/env-verify.sh`、`test/scenarios/env-cleanup.sh`，证明环境可独立启动/验证/清理；若 Docker/端口/镜像阻塞，记录 blocker 与输出，不用具体场景 runner 代替。
  <!-- verified: 2026-05-27 command="test/scenarios/env-setup.sh && test/scenarios/env-verify.sh && test/scenarios/env-cleanup.sh" evidence="setup reused already healthy dev-stack; verify returned postgres-dev/redis-dev/minio-dev/mailpit-dev OK with summary ok=4 degraded=0 down=0 total=4; cleanup stopped containers and cleaned up easyinterview-dev network while preserving named volumes" -->

## Phase 7: local raw output debug default revision

- [x] 7.1 默认开启本地 raw output debug：`config/dev.yaml` / `config/test.yaml` / 根 `.env.example` / `deploy/dev-stack/.env.example` 使用 `AI_DEBUG_PRINT_RAW_OUTPUT=true`，`config/config.yaml` 与 staging/prod 默认保持关闭；`E2E.P0.100` trigger 从 `deploy/dev-stack/.env` 校验 true；验证：focused config test、scenario env contract test、真实 `scenario-run -i E2E.P0.100` PASS。
  <!-- verified: 2026-05-27 command="go test ./backend/internal/platform/config -run TestRepoLocalConfigEnablesRawOutputDebugOnlyForLocalEnvironments -count=1; python3 -m pytest scripts/lint/scenario_env_contract_test.py -q -k real_provider_hybrid_uat_uses_dev_stack_env_as_single_source; scenario-run -i E2E.P0.100" evidence="dev/test config raw debug true and staging/prod false; P0.100 env contract requires AI_DEBUG_PRINT_RAW_OUTPUT=true; real-provider full funnel run p0-100-debug-1779866312146 passed with redacted provider/profile/model/task-run evidence" -->

## Phase 8: developer debug handoff revision

- [x] 8.1 新增 `test/scenarios/_shared/scripts/local-dev-runtime.sh`：从 `deploy/dev-stack/.env` 推导 frontend/backend/Mailpit/MinIO 地址，统一 `.test-output/local-dev/{backend,frontend}.log` 与 `.pid`，输出 `tail -f` 和 `make dev-logs SERVICE=<name>` 调试入口；不得打印 secret、cookie 或 raw email code。
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="contract covers local-dev-runtime helper, start_new_session detached launch, endpoint summary, backend/frontend log paths, and no p0/manual coupling" -->
- [x] 8.2 修订 `test/scenarios/env-redeploy.sh backend|frontend|all`：build artifact gate 通过后重启对应 host-run backend/frontend 进程，等待端口可连接，失败时打印日志尾段；`deps` 仍只管理 Docker Compose 外部依赖。
  <!-- verified: 2026-05-27 command="test/scenarios/env-redeploy.sh all" evidence="dev-stack dependencies OK, backend go build passed and backend listened on 127.0.0.1:8080, frontend build passed and Vite listened on 127.0.0.1:5173" -->
- [x] 8.3 修订 `env-setup.sh` / `env-status.sh` / `env-verify.sh`：创建、巡检或验证共享环境后输出同一调试摘要；status/verify 保持 `dev-doctor` JSON stdout，调试摘要走 stderr。
  <!-- verified: 2026-05-27 command="test/scenarios/env-setup.sh --dry-run; test/scenarios/env-status.sh --dry-run" evidence="setup explains dev-up/dev-doctor and debug summary; status keeps dry-run make dev-doctor on stdout and debug summary notice on stderr" -->
- [x] 8.4 文档与 skill 对齐：更新 `deploy/dev-stack/README.md`、`test/scenarios/README.md`、`test/scenarios/e2e/README.md`、`scenario-env` / `scenario-redeploy` skill，明确 redeploy 是 build + restart，且输出可接管地址/日志/PID。
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
- [x] 10.3 Docs/runbook：`deploy/dev-stack/README.md`、`test/scenarios/README.md`、`test/scenarios/e2e/README.md` 说明一键清数据重编译重部署入口，并区分普通重启 `scenario-env-redeploy TARGET=all`。
  <!-- verified: 2026-07-09 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q -k \"reset_redeploy or scenario_docs_describe\"" evidence="focused contract verifies Makefile target plus dev-stack/scenario/e2e README reset-redeploy wording" -->
- [x] 10.4 Phase 10 self-check：focused contract pytest、`make scenario-env-reset-redeploy ARGS=--dry-run`、`make docs-check`、`sync-doc-index --check`、`git diff --check` 全部通过。
  <!-- verified: 2026-07-09 command="make scenario-env-reset-redeploy ARGS=--dry-run; python3 -m pytest scripts/lint/scenario_env_contract_test.py -q; python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check; make docs-check; git diff --check" evidence="dry-run previewed reset/setup-migrations/redeploy/verify without changing environment; 13 scenario env contract tests passed; docs/index/link gates and whitespace gate passed" -->
