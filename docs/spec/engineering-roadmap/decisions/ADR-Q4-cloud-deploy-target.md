# ADR-Q4 · 部署与测试目标

> **版本**: 2.0
> **状态**: accepted
> **更新日期**: 2026-07-07

## 1 背景

`backend-runtime-topology` v1.0 已将 P0 运行单元收敛为 `frontend` + `backend`，后台任务在 backend internal runner 中执行；`README.md` §「待评审的 5 个决策点」第 4 项只作为决策输入。

2026-05-22 重新对齐后，仓库当前事实是：

- A2 `local-dev-stack` 只需要 Docker Compose 管理 Postgres / Redis / MinIO / Mailpit 外部依赖；Mailpit 只承接本地 email-code 收信，不表示应用进程进入 compose。
- backend / frontend 当前可通过宿主机 dev command 直接运行并连接这些依赖；把它们强制打包进镜像或 Kind 集群会增加本地反馈成本。
- `test/scenarios/e2e/` 当前只承载真实 HTTP API 或连接真实 backend 的浏览器 UI 流程，不需要部署级 Kind / K8s 环境；代码层回归由根 `make test` 承接。
- 当前单人阶段 A5 只保留本地质量门禁，不构建远端 CI pipeline。
- 没有 vendor lock-in：所有外部依赖通过 SDK 接口、OpenAPI fixture 或 provider registry 接入。

业务背景：

- P0 团队 ≤ 数人；运维带宽有限。
- MVP 阶段更需要快速本地反馈、明确脚本证据和低维护成本，而不是提前引入集群编排。
- staging / production、自动发布、灰度和 SLO gate 尚未进入当前执行计划。

## 2 选项与取舍

### 选项 A · Docker Compose 外部依赖 + 宿主机 app runtime + 真实 API/UI E2E

**Pros**：

- 与当前代码事实一致：`make dev-up` 管理外部依赖，backend/frontend 可在宿主机直接运行。
- E2E 直接操作真实 HTTP API 或连接真实 backend 的浏览器 UI，证据来自请求、页面与持久化结果；Go/Vitest 等代码层回归仍由根 `make test` 统一执行。
- 不要求 Dockerfile、镜像 registry、Helm chart、K8s context 或集群权限作为普通开发/测试前置。
- 保留逐步演进空间：后续组件若确实需要容器化 app service，可由 owner 显式接入 compose。

**Cons**：

- 不提供部署级滚动发布、自动扩缩容或集群级自愈能力。
- 如果进入多人远端环境或公开 release，需要 E4 重新评估部署目标和发布自动化。

### 选项 B · Docker Compose app services / 单机容器部署

**Pros**：

- 比 K8s 简单，仍能复用容器镜像和健康检查。
- 适合作为未来 staging 或小规模 production 的轻量候选。

**Cons**：

- 需要补 Dockerfile、compose app service、镜像构建和部署 runbook；当前尚无必要。
- HA、滚动升级、secret rotation 和备份策略仍需 E4 单独设计。

### 选项 C · Kubernetes / managed cluster

**Pros**：

- 标准化 rollout、HPA、secret/config、observability 和 GitOps 路径。
- 适合多人协作、公开 release、强 SLO 或多环境治理。

**Cons**：

- 对当前阶段过重：需要集群、Helm/Kustomize、镜像 registry、secret 管理和更多运维角色。
- 会把本地 scenario 与部署资产强耦合，拉长普通功能开发和 bugfix 的验证回路。

### 选项 D · PaaS（Fly.io / Railway / Render 等）

**Pros**：

- 起步快，部分平台内置 TLS、日志和部署流水线。

**Cons**：

- 平台差异大，容易形成新的部署描述和 secret 管理分支。
- 与未来自托管观测、隐私和队列能力的兼容性需要重新评估。

## 3 决策

**P0 锁定选项 A：Docker Compose 外部依赖 + 宿主机 app runtime + 真实 API/UI E2E。**

落地约束：

1. **本地依赖**：`make dev-up` 默认只启动 Postgres / Redis / MinIO / Mailpit 以及已显式接入的 optional app service；不得因为组件具备本地运行入口就强制放进 compose。
2. **应用运行**：backend / frontend 默认通过宿主机 dev command 管理，连接 `deploy/dev-stack/.env.example` 暴露的本地依赖端口与连接串。
3. **场景验证**：`test/scenarios/e2e/` 只维护真实 HTTP API 或连接真实 backend 的浏览器 UI 流程；不得把 `go test`、Vitest、pytest、lint、build 或 package smoke 编排进场景脚本或 E2E 证据，也不得默认要求 Kind / K8s / Helm。前后端全量单测统一由根 `make test` 执行，focused test 只作开发反馈。
4. **AI provider（关联 Q-6）**：非测试本地 app run、未来 staging 和 production 必须通过 provider registry / model profile / provider-specific secret env ref 注入真实 provider；`APP_ENV=test` 的 stub/fixture 只能用于单元测试、离线契约测试或显式代码层 mock test，不能作为 E2E backend。
5. **容器化 app service**：后续组件只有在 owner plan 明确需要可复现容器化 app runtime 时，才新增 Dockerfile / compose service / healthcheck / resource budget，并同步 A2 spec。
6. **CI/CD 延后**：当前个人单人开发阶段不构建远端 CI pipeline，也不做 CI deploy；A5 只约束本地手动质量门禁。
7. **Release 目标重评估**：E4 `release-gate-and-rollout` 创建时，必须基于当时真实团队规模、SLO、运维带宽和成本重新评估单机 compose、PaaS、ECS 或 K8s；不得把 Kind / K8s 当作默认答案。

## 4 影响范围

- **A2 `local-dev-stack`** —— Docker Compose 默认覆盖外部依赖；backend/frontend 默认宿主机运行；optional app service 必须由 owner 显式接入。
- **A5 `ci-pipeline-baseline`** —— 当前只保留本地质量门禁；远端 CI pipeline / image push / branch protection 延后。
- **`test/scenarios/`** —— 场景环境文档从 Kind 目标改为本地真实 API/UI 契约；场景验证必须读取具体 README / wrapper，并让请求落到真实 backend，而不是假设部署环境或复用代码层测试结果。
- **E4 `release-gate-and-rollout`** —— 未创建；后续 release/staging/prod 目标必须重新设计，不继承 Kind / K8s 默认值。
- **F1 `observability-stack`** —— 当前只消费应用 `/metrics`、日志和真实 API/UI E2E 证据；不要求本地默认启动 Prometheus / Loki / Grafana / OTel Collector。
- **A3 `ai-provider-and-model-routing`** —— 非测试运行环境继续 fail-fast 要求真实 provider registry/profile/secret 注入；不因为去掉 Kind 而放宽 AI provider 边界。
- **A4 `secrets-and-config`** —— 继续 owner env 字典、secret ref 和 fail-fast 规则；不新增 K8s Secret 作为当前 P0 前提。
- **AGENTS.md / CLAUDE.md** —— Agent 执行门禁必须区分 Docker Compose 外部依赖、宿主机 app runtime、代码层 `make test` 与真实 API/UI E2E，不得默认引入 Kind / K8s / Helm。

## 5 失效与修订条件

触发推翻或升级本 ADR 的具体阈值：

- 第二名长期贡献者加入，且本地真实 API/UI 环境无法提供一致验证 → E4 / A5 评估远端 CI 和容器化 app service。
- 创建公开 release branch、付费用户上线或自动发版需求出现 → E4 创建 release/staging/prod plan，并重新比较单机 compose、PaaS、ECS、K8s。
- 单机/宿主机开发环境导致高频环境漂移或故障复现困难 → 评估把具体 app service 接入 compose，而不是直接跳到 K8s。
- 出现 HA、自动扩缩容、多 region active-active、强 SLO 或复杂 secret rotation 需求 → 新 ADR 评估 K8s / managed platform / PaaS。

修订流程：如需推翻本决策，新增修订 ADR 并同步 roadmap Q-4 与相关 owner spec。

## 6 关联

- `engineering-roadmap/spec.md` §3.2 Q-4
- `engineering-roadmap/plans/001-decompose-subspecs/plan.md` checklist 1.1
- 当前运行事实：`local-dev-stack/spec.md`、`deploy/dev-stack/README.md`、`test/scenarios/README.md`
- 下游 child：A2 / A5 / E4 / F1 / A3 / A4
- 关联 ADR：ADR-Q6-ai-provider-and-model-routing

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-07 | 1.9 | 对齐当前 email-code 登录、本地 runner 和 Kind / K8s 非默认部署边界。 |
| 2026-05-26 | 1.8 | 对齐 local-dev-stack Mailpit revision：默认 Docker Compose 外部依赖增加 Mailpit 本地邮箱 sink，仍保持宿主机 backend/frontend app runtime 与本地 scenario runner 口径。 |
| 2026-05-22 | 1.7 | 按方案 A 重定部署与测试目标：P0 默认 Docker Compose 外部依赖 + 宿主机 app runtime + 本地 scenario runner；Kind / K8s / Helm 不再作为默认测试或部署前提。 |
| 2026-05-06 | 1.6 | 对齐 backend-runtime-topology：P0 部署拓扑从 web/api/worker 三应用单元改为 web/backend 两应用单元，后台任务默认由 backend internal runner 承接。 |
| 2026-05-05 | 1.5 | 对齐 A3 003 Provider Registry：部署注入从单一 endpoint/key 口径更新为 registry/profile/provider-specific secret 组合，`AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 仅作为默认 provider ref 可引用 env。 |
| 2026-05-05 | 1.4 | 对齐 ADR-Q6 provider 口径：业务 deployment 只通过 `AI_PROVIDER_BASE_URL` 接入 OpenAI-compatible provider endpoint，不把独立转发层写成应用部署前提。 |
| 2026-04-27 | 1.3 | 对齐个人单人开发阶段决策：P0 当前不构建远端 CI pipeline，不做 CI deploy；A5 只约束本地手动质量门禁，自动化 CI/CD 待多人协作、公开 release 或自动发版需求出现后再建。 |
