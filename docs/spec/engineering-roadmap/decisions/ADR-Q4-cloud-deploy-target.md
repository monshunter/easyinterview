# ADR-Q4 · 云部署目标

> **版本**: 1.6
> **状态**: accepted
> **更新日期**: 2026-05-06

## 1 背景

`engineering-roadmap decisions` 历史上把 P0 / P1 部署形态锁定为「3 个运行单元」（`web-app` / `api` / `worker`）+「共享基础设施」。`backend-runtime-topology` v1.0 已将 P0 运行单元收敛为 `frontend` + `backend`，后台任务在 backend internal runner 中执行；`README.md` §「待评审的 5 个决策点」第 4 项只作为历史决策输入。

仓库现状（与决策强相关）：

- `CLAUDE.md` §5 + `test/scenarios/README.md` 已锁定**场景集成测试基于 Kind**（K8s 本地集群）
- `image-cache.sh pull` 与 helm-chart 友好的脚本约定已在 skill `scenario-env` / `scenario-redeploy` 中固化
- A2 `local-dev-stack` 用 docker-compose 起本地最小依赖与项目组件；当前单人阶段 A5 只保留本地质量门禁，不构建远端 CI pipeline
- 没有 vendor lock-in：所有外部依赖通过 SDK 接口接入

业务背景：

- P0 团队 ≤ 数人；运维带宽有限
- 流量预期：MVP 阶段 < 5k MAU，但需支持单日 / 单周高峰（求职窗口期）
- 隐私偏好倾向 EU / 自主可控；多地域延后到 P1+

## 2 选项与取舍

### 选项 A · Kubernetes（managed cluster：EKS / GKE / AKS 之一；staging 用 Kind）

**Pros**：

- 与既有 Kind 场景测试栈天然一致：`test/scenarios/` 的 helm chart / manifest 直接复用到 staging / prod
- 2 个应用 deployment（web / backend）+ 1 ingress + secrets / configmaps / hpa 全标准化；未来拆分后台 runner 需新 ADR
- 可观测性 stack（Prometheus / Loki / OTel Collector / Grafana）以 helm chart 形式分发
- 运维团队 / SRE 文化对齐；社区方案最成熟
- 可平滑迁移到自托管 / 私有云（不锁特定云厂商）
- 成本可控：单个小型 managed cluster 即可承载 P0+P1

**Cons**：

- 需要至少一个 ops 角色（兼职可）熟悉 K8s
- 起步成本高于 Fly.io / Railway 等 PaaS

### 选项 B · AWS ECS（Fargate）

**Pros**：

- AWS 生态内一键化，无需管理节点
- 与 RDS / ElastiCache / S3 集成简单

**Cons**：

- 与现有 Kind 场景测试栈不一致 → 必须维护两套部署描述
- 强 AWS lock-in；切到 GCP / Azure / 自托管成本高
- helm chart / Prometheus stack 需重写为 ECS task definition + CloudWatch；Loki / OTel Collector 不在原生路径上

### 选项 C · Fly.io / Railway / Render（PaaS）

**Pros**：

- 起步速度最快（git push 即部署）
- 内置全球边缘网络 + 自动 TLS
- 无运维负担

**Cons**：

- 无法承载完整 OpenTelemetry / Loki / 自托管 PostHog / Higress 等组件（与 Q-3 / Q-6 / F1 链路冲突）
- 数据库 / Redis / S3 仍需外部 SaaS（成本叠加）
- 不可灵活水平扩展后台 runner（大队列场景受限）
- 与 Kind 测试栈完全脱节

### 选项 D · 自托管裸金属 / 单机 docker-compose

**Pros**：

- 无云厂商成本

**Cons**：

- 无 HA / 无健康检查 / 无自愈 / 无滚动升级 → 与 E4 release-gate 灰度演练冲突
- 安全 / 备份 / 监控全自建
- 与 Q-3 / Q-6 / F1 / E4 部署假设冲突

## 3 决策

**P0 锁定选项 A：Kubernetes 作为唯一部署目标。**

落地约束：

1. **集群形态**：staging / prod 各 1 个 managed cluster（云厂商 = ops 选择，初期默认 EKS / GKE / AKS 任一；本 ADR 不锁厂商）；本地场景集成测试继续用 Kind（与 `test/scenarios/` 一致），普通本地开发走 A2 docker-compose，不把 Kind 作为开发前置条件
2. **工作负载拓扑**：P0 为 2 个应用 Deployment（`web-app` 静态资源由 ingress 直 serve 或单独 deploy / `backend` HPA min=2）；outbox / background runner 默认在 backend 内部运行。独立 runner / CronJob 只能作为后续显式扩展，不是 P0 默认拓扑。
3. **共享基础设施**：PostgreSQL + pgvector / Redis 优先用云托管（RDS+pgvector or Neon / ElastiCache）；Object Storage 用云对象存储（S3 / GCS / R2）；OTel Collector / Loki / Prometheus / Grafana 自托管在同一 cluster
4. **AI provider（关联 Q-6）**：业务 deployment 通过 `AI_PROVIDER_REGISTRY_PATH` + `AI_MODEL_PROFILE_PATH` + registry 内 provider-specific secret env ref 注入 AI 连接；`AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 只可作为默认 OpenAI-compatible provider ref 引用的 env 名；Kind 场景测试注入同一 registry/profile/secret 组合，不要求部署 AI provider
5. **Helm chart**：所有组件以 helm chart 形式管理；chart 与 `test/scenarios/` Kind 部署共用同一 values 模板（区别包含 replica / resource / AI registry/profile/secret 注入方式）
6. **CI/CD 延后**：当前个人单人开发阶段不构建远端 CI pipeline，也不做 CI deploy；A5 只约束本地手动质量门禁。自动化 deploy 由 E4 `release-gate-and-rollout` 在公开 release / 多人协作 / 自动发版需求出现后单独管理（GitOps：ArgoCD / FluxCD 任一，本 ADR 不锁工具）
7. **secrets**：Sealed Secrets / SOPS / External Secrets 任一，通过 A4 `secrets-and-config` 抽象注入；不允许明文 ConfigMap

## 4 影响范围

- **A2 `local-dev-stack`** —— docker-compose 只覆盖普通本地开发的最小依赖与项目组件启动；不承接 Kind manifest / Helm chart 同源要求
- **A5 `ci-pipeline-baseline`** —— 当前只保留本地质量门禁；远端 CI pipeline / image push / branch protection 延后
- **E4 `release-gate-and-rollout`** —— 灰度（feature flag + Deployment progressive rollout）+ 回滚 runbook 全部基于 K8s 原语
- **F1 `observability-stack`** —— OTel Collector / Prometheus / Loki / Grafana 以 helm chart 部署在同 cluster
- **A3 `ai-provider-and-model-routing`** —— OpenAI-compatible / stub provider adapter 与 Provider Registry + Capability Model Profile；staging / prod 通过 provider ref 解析连接参数
- **A4 `secrets-and-config`** —— K8s Secret 抽象（含 sealed / external），Kind 与 docker compose 本地部署注入真实 provider registry/profile/secret 组合
- **CLAUDE.md / `test/scenarios/`** —— Kind 场景测试栈与生产栈共用 helm chart 路径，但 Kind 通过真实 provider registry/profile/secret 组合接入外部 AI 能力；与 A2 docker-compose 本地开发栈保持双轨独立

## 5 失效与修订条件

触发推翻或升级本 ADR 的具体阈值：

- MAU < 1k 持续 6 个月 + ops 带宽不足 → 评估降级到 Fly.io / Render（接受重写部署描述）
- 出现强 vendor lock-in 需求（如必须用 AWS Bedrock + IAM 集成）→ 评估迁移到 ECS
- K8s 集群运维成本（含 cluster 自身 + 4 个共享组件）> 业务实际 infra 成本 50% 持续 1 季 → 评估自托管 + lighter orchestrator（Nomad / docker-compose on bare metal + ansible）
- 出现多 region active-active 需求 → 评估 multi-cluster + service mesh（Istio / Linkerd）

修订流程：本 ADR 状态由 `accepted` → `superseded`，新 ADR 显式标注 `supersedes: ADR-Q4-cloud-deploy-target.md`。

## 6 关联

- `engineering-roadmap/spec.md` §3.2 Q-4
- `engineering-roadmap/plans/001-decompose-subspecs/plan.md` checklist 1.1
- 参考背景与当前约束：`engineering-roadmap decisions` §3、`CLAUDE.md` §5 场景测试环境、`test/scenarios/README.md`
- 下游 child：A2 / A5 / E4 / F1 / A3 / A4
- 关联 ADR：ADR-Q6-ai-provider-and-model-routing（Higress as cluster-internal Deployment）

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-05-06 | 1.6 | 对齐 backend-runtime-topology：P0 部署拓扑从 web/api/worker 三应用单元改为 web/backend 两应用单元，后台任务默认由 backend internal runner 承接。 |
| 2026-05-05 | 1.5 | 对齐 A3 003 Provider Registry：部署注入从单一 endpoint/key 口径更新为 registry/profile/provider-specific secret 组合，`AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 仅作为默认 provider ref 可引用 env。 |
| 2026-05-05 | 1.4 | 对齐 ADR-Q6 provider 口径：业务 deployment 只通过 `AI_PROVIDER_BASE_URL` 接入 OpenAI-compatible provider endpoint，不把独立转发层写成应用部署前提。 |
| 2026-04-27 | 1.3 | 对齐个人单人开发阶段决策：P0 当前不构建远端 CI pipeline，不做 CI deploy；A5 只约束本地手动质量门禁，自动化 CI/CD 待多人协作、公开 release 或自动发版需求出现后再建。 |
| 2026-04-27 | 1.2 | 对齐 ADR-Q6 v1.1：Kind 场景测试属于本地部署，默认注入真实 AI provider endpoint / key，不要求部署 AI provider；staging / prod 也复用 `AI_PROVIDER_BASE_URL` 的 endpoint 契约。 |
| 2026-04-27 | 1.1 | 对齐 A2 local-dev-stack v1.2：普通本地开发走 docker-compose 最小依赖 + 项目组件，Kind 仅用于场景集成测试，不再要求 A2 docker-compose 与 Kind manifest 同源。 |
