# ADR-Q4 · 云部署目标

> **版本**: 1.0
> **状态**: accepted
> **更新日期**: 2026-04-26

## 1 背景

`easyinterview-tech-docs/01-technical-architecture.md` §3 把 P0 / P1 部署形态锁定为「3 个运行单元」（`web-app` / `api` / `worker`）+「共享基础设施」（PostgreSQL / Redis / Object Storage / 监控 / 日志 / 追踪 / 外部 AI Provider）；`README.md` §「待评审的 5 个决策点」第 4 项把云部署目标留作 W0 决策。

仓库现状（与决策强相关）：

- `CLAUDE.md` §5 + `test/scenarios/README.md` 已锁定**场景集成测试基于 Kind**（K8s 本地集群）
- `image-cache.sh pull` 与 helm-chart 友好的脚本约定已在 skill `scenario-env` / `scenario-redeploy` 中固化
- A2 `local-dev-stack` 用 docker-compose 起本地依赖，A5 `ci-pipeline-baseline` 不做 deploy
- 没有 vendor lock-in：所有外部依赖通过 SDK 接口接入

业务背景：

- P0 团队 ≤ 数人；运维带宽有限
- 流量预期：MVP 阶段 < 5k MAU，但需支持单日 / 单周高峰（求职窗口期）
- 隐私偏好倾向 EU / 自主可控；多地域延后到 P1+

## 2 选项与取舍

### 选项 A · Kubernetes（managed cluster：EKS / GKE / AKS 之一；staging 用 Kind）

**Pros**：

- 与既有 Kind 场景测试栈天然一致：`test/scenarios/` 的 helm chart / manifest 直接复用到 staging / prod
- 3 deployment（web / api / worker）+ 1 ingress + secrets / configmaps / hpa 全标准化
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
- 不可灵活水平扩展长任务 worker（Asynq 大队列场景受限）
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

1. **集群形态**：staging / prod 各 1 个 managed cluster（云厂商 = ops 选择，初期默认 EKS / GKE / AKS 任一；本 ADR 不锁厂商）；本地 dev 继续用 Kind（与 `test/scenarios/` 一致）
2. **工作负载拓扑**：3 个 Deployment（`web-app` 静态资源由 ingress 直 serve 或单独 deploy / `api` HPA min=2 / `worker` HPA min=1），1 个 CronJob（outbox dispatcher 兜底重试）
3. **共享基础设施**：PostgreSQL + pgvector / Redis 优先用云托管（RDS+pgvector or Neon / ElastiCache）；Object Storage 用云对象存储（S3 / GCS / R2）；OTel Collector / Loki / Prometheus / Grafana 自托管在同一 cluster
4. **AI Gateway（关联 Q-6）**：作为同 cluster 内独立 Deployment（Higress 默认候选），通过内部 Service 对外暴露 OpenAI-compatible route；业务 deployment 通过 `AI_GATEWAY_BASE_URL` 引用
5. **Helm chart**：所有组件以 helm chart 形式管理；chart 与 `test/scenarios/` Kind 部署共用同一 values 模板（区别只在 replica / resource）
6. **CI 不直接 deploy**：A5 `ci-pipeline-baseline` 只跑 lint / test / build / image push；deploy 由 E4 `release-gate-and-rollout` 单独管理（GitOps：ArgoCD / FluxCD 任一，本 ADR 不锁工具）
7. **secrets**：Sealed Secrets / SOPS / External Secrets 任一，通过 A4 `secrets-and-config` 抽象注入；不允许明文 ConfigMap

## 4 影响范围

- **A2 `local-dev-stack`** —— docker-compose 与 Kind manifest 同源（同一 image / 同一健康检查）
- **A5 `ci-pipeline-baseline`** —— 仅 build + push image；不做 deploy
- **E4 `release-gate-and-rollout`** —— 灰度（feature flag + Deployment progressive rollout）+ 回滚 runbook 全部基于 K8s 原语
- **F1 `observability-stack`** —— OTel Collector / Prometheus / Loki / Grafana 以 helm chart 部署在同 cluster
- **A3 `ai-gateway-and-model-routing`** —— Higress / 替代 AI Gateway 作为 cluster-internal Deployment
- **A4 `secrets-and-config`** —— K8s Secret 抽象（含 sealed / external）
- **CLAUDE.md / `test/scenarios/`** —— Kind 场景测试栈与生产栈共用 helm chart 路径，避免双轨

## 5 失效与修订条件

触发推翻或升级本 ADR 的具体阈值：

- MAU < 1k 持续 6 个月 + ops 带宽不足 → 评估降级到 Fly.io / Render（接受重写部署描述）
- 出现强 vendor lock-in 需求（如必须用 AWS Bedrock + IAM 集成）→ 评估迁移到 ECS
- K8s 集群运维成本（含 cluster 自身 + 4 个共享组件）> 业务实际 infra 成本 50% 持续 1 季 → 评估自托管 + lighter orchestrator（Nomad / docker-compose on bare metal + ansible）
- 出现多 region active-active 需求 → 评估 multi-cluster + service mesh（Istio / Linkerd）

修订流程：本 ADR 状态由 `accepted` → `superseded`，新 ADR 显式标注 `supersedes: ADR-Q4-cloud-deploy-target.md`。

## 6 关联

- `engineering-roadmap/spec.md` §3.2 Q-4
- `engineering-roadmap/plans/001-decompose-subspecs/plan.md` Phase 1.1
- 上游：`easyinterview-tech-docs/01-technical-architecture.md` §3、`CLAUDE.md` §5 场景测试环境、`test/scenarios/README.md`
- 下游 child：A2 / A5 / E4 / F1 / A3 / A4
- 关联 ADR：ADR-Q6-ai-gateway-and-model-routing（Higress as cluster-internal Deployment）
