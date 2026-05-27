# Local Dev Stack History

> **版本**: 1.19
> **状态**: active
> **更新日期**: 2026-05-27

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-27 | 1.19 | 对齐 backend-auth / frontend-shell 的 email-code 修订：Mailpit 本地邮件改为 6 位验证码、5 分钟有效，`EMAIL_VERIFY_BASE_URL` 仅保留为本地 frontend origin / CORS 推导来源，不再拼入邮件链接。 | backend-auth/001 Phase 7 + frontend-shell/001 Phase 8 |
| 2026-05-27 | 1.18 | 将 scenario redeploy 闭环修订为 rebuild + 重启 host-run backend/frontend，并要求 env setup/status/verify/redeploy 输出服务地址、PID、日志路径和容器日志命令，便于开发者接管调试。 | local-dev-stack/001 developer debug handoff |
| 2026-05-27 | 1.17 | 修订本地 Mailpit 登录闭环：默认邮件链接进入 frontend `/auth/verify` callback，由前端调用 backend verify API、刷新 session 并清理 URL token；手动 token 仅保留为 fallback。 | frontend-shell/001 Phase 7 |
| 2026-05-27 | 1.16 | 本地测试与本地真实联调默认开启 `AI_DEBUG_PRINT_RAW_OUTPUT=true`，并要求 P0.100 hybrid preflight 校验该开关；staging/prod 默认仍关闭，raw output 不进入持久化审计。 | local-dev-stack/001 raw debug local default |
| 2026-05-27 | 1.15 | 明确 `deploy/dev-stack/.env` 是本地真实前后端联调唯一 env 来源，`.env.example` 覆盖 auth secrets、frontend real mode、AI provider 与共享依赖配置；场景不得维护独立 `.env`。 | local-dev-stack/001 + e2e-scenarios-p0/002 |
| 2026-05-27 | 1.14 | 将共享测试环境与本地前后端联调环境生命周期从具体场景脚本中抽离：新增 `test/scenarios/env-*.sh` 与根 `scenario-env-*` Make target 作为 setup / status / verify / cleanup / redeploy 真理源，供 `/scenario-env` 与 `/scenario-redeploy` skill 调用。 | local-dev-stack/001 environment lifecycle revision |
| 2026-05-26 | 1.13 | 将 Mailpit 纳入默认本地依赖：新增 `mailpit-dev`（Web 8025 / SMTP 1025）、dev-doctor `/readyz` 与 SMTP 端口探测、`.env.example` 邮件配置；本地 magic-link 登录走真实 backend auth flow，不再依赖真实外部邮箱服务、真实邮箱账号或场景专属 backend cmd。 | local-dev-stack/001 Mailpit revision |
| 2026-05-22 | 1.12 | 修复 Postgres 18 官方镜像 PGDATA / volume 挂载契约：`easyinterview-pg-data` 挂到 `/var/lib/postgresql`，由镜像管理 `/18/docker` 子目录；`make dev-up` 增加只读旧卷布局 preflight，避免旧 `/var/lib/postgresql/data` 或半初始化卷表现为不明原因 unhealthy。 | local-dev-stack/001 L2 runtime remediation |
| 2026-05-22 | 1.11 | 按用户确认的方案 A 对齐部署与测试环境：默认 `make dev-up` 只管理 Docker Compose 外部依赖，backend/frontend 由宿主机 dev command 管理；`test/scenarios/` 默认使用 repo-tracked 本地 runner，不再把 Kind / K8s / Helm 作为 P0 本地测试、smoke 或部署前提。 | local-dev-stack/001 post-pass revision |
| 2026-05-08 | 1.10 | 按用户决策将默认本地 Postgres 镜像从 16 升级到 18，并同步 B4 迁移基线的本地 DB 前提。 | local-dev-stack/001 post-pass revision |
| 2026-05-08 | 1.9 | 对齐 A3/B4 当前决策：默认本地依赖收敛为普通 Postgres / Redis / MinIO；删除未使用扩展 init/probe 口径，未来需要时重新设计。 | ai-provider-and-model-routing/003 Phase 6 |
| 2026-05-06 | 1.8 | 对齐 backend-runtime-topology：默认本地栈不接入独立 worker 进程或 worker host port，backend background runner 随 backend 应用组件观测。 | backend-runtime-topology/001-worker-consolidation |
| 2026-05-05 | 1.7 | 本地 dev-stack AI provider 配置纳入 A3/A4 单文件 registry/profile catalog path，`.env.example` 必须暴露 `AI_PROVIDER_REGISTRY_PATH` 与 `AI_MODEL_PROFILE_PATH` canonical 值。 | ai-provider-and-model-routing/003 L2 remediation |
| 2026-05-05 | 1.6 | 收口 AI provider 口径：A2 本地栈只传递真实 provider endpoint 配置，不启动 provider mock / proxy 容器，也不把部署切到单元测试 stub。 | ai-provider-and-model-routing/001 remediation |
| 2026-04-29 | 1.5 | 文档侧 reconcile：把已落地 compose 的 MinIO / mc 不可变 tag 写回 D-2；A2 executable gate 明确同时覆盖 AI provider fail-fast C-9；§7 从未来计划改为 `001-bootstrap` 已完成事实，不新开 plan。 | plan-review remediation |
| 2026-04-27 | 1.4 | 对齐 A5 单人开发阶段决策：`make dev-doctor` JSON 仍保持可被未来 CI 消费，但当前不把 A5 CI 作为本地开发栈前置。 | [001-bootstrap](./plans/001-bootstrap/plan.md) |
| 2026-04-27 | 1.3 | 对齐 A3 / A4 AI provider 规则：docker compose 本地部署不启动 AI provider 容器，也不使用单元测试 stub；A2 只传递 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 占位，启用 AIClient 的组件缺真实 provider 配置时 fail-fast。 | [001-bootstrap](./plans/001-bootstrap/plan.md) |
| 2026-04-27 | 1.2 | 按 L1 plan-review 与用户确认修订本地开发栈边界：默认依赖收敛为当时最小 Postgres / Redis / MinIO；`make dev-up` 改为启动最小依赖 + 当前项目可运行组件；本地观测改为应用 `/metrics` + 容器日志；默认排除 OTel Collector / Grafana / Loki / Prometheus 与 AI provider。 | [001-bootstrap](./plans/001-bootstrap/plan.md) |
| 2026-04-27 | 1.1 | spawn `001-bootstrap` impl plan：把 spec §3.1 D-1..D-7 与 §6 C-1..C-8 落到 4 个 phase（compose+init / make 生命周期 / dev-doctor / OTel 通路+收口），关闭 [engineering-roadmap/001 Phase 3.6](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 的「executable gate by A2 child」承诺。spec 内容未变，版本不动 | [001-bootstrap](./plans/001-bootstrap/plan.md) |
| 2026-04-27 | 1.1 | 修正 W1 gate 口径：parent Phase 3 只锁定 A2 spec-contract；真实 `make dev-up` / `make dev-doctor` 可执行 gate 由 A2 child `001` plan 验证后再放行依赖本地栈的 W2 implementation | engineering-roadmap/001 Phase 3 remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定 `deploy/dev-stack/docker-compose.yaml` 落点、当时的本地依赖服务清单、`make dev-up` / `dev-down` / `dev-doctor` / `dev-reset` 契约、健康检查口径与命名卷策略；承接 [engineering-roadmap §5.7](../engineering-roadmap/spec.md#6-实施顺序) 的 W1 dev-up spec-contract lock；引用 `engineering-roadmap decisions §2.3` 的本地依赖列表。 | engineering-roadmap/001 Phase 3 |
