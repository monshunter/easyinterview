# easyinterview 技术实施文档包（P0 / P1）

> **状态**: historical-input
> **更新日期**: 2026-05-03
> **执行边界**: 本目录是 2026-04-26 基于旧产品 spec 整理的历史技术输入，尚未按 `docs/spec/product-scope/spec.md` v1.4、`docs/ui-design/` 和当前 `ui-design/` 静态原型全量重写。它不得作为当前 API、DB、事件、指标、路由或产品模块的可执行真理源。

本目录保留用于考古和解释早期技术假设。当前可执行契约以以下文件为准：

- 产品范围：`docs/spec/product-scope/spec.md`
- UI / 交互：`docs/ui-design/` 与 `ui-design/`
- OpenAPI：`openapi/openapi.yaml` 与 `docs/spec/openapi-v1-contract/spec.md`
- 共享枚举 / 错误码 / ID：`shared/conventions.yaml` 与 `docs/spec/shared-conventions-codified/spec.md`
- 内部事件 / job：`shared/events.yaml`、`shared/jobs.yaml` 与 `docs/spec/event-and-outbox-contract/spec.md`
- 数据库：`migrations/` 与 `docs/spec/db-migrations-baseline/spec.md`
- 可观测性：`docs/spec/observability-stack/spec.md`

本目录仍包含旧 `mistakes` / `growth` / `warmup` / `single_drill` / `counter_questions`、旧 36 endpoint / 14 tag、旧 27 应用表、旧 18 event 等已经被当前产品 spec 和 Layer B/F contract 修订或删除的内容。若本目录与上述当前真理源冲突，一律以后者为准；不得绕过 Layer B/F 的编码 truth source 直接按本目录实施。

以下原文仅作为历史背景保留。

本目录基于早期产品 spec 的 P0 / P1 主链路整理，面向前后端分离实现落地。文档目标不是复述产品需求，而是把需求收敛成一套可以直接进入开发、联调、评审和上线治理的技术约束。

## 文档清单

1. [`00-shared-conventions.md`](./00-shared-conventions.md)  
   全局统一约定：命名、ID、时间、枚举、错误码、版本化、前后端契约生成方式。

2. [`01-technical-architecture.md`](./01-technical-architecture.md)  
   系统分层、部署形态、模块边界、核心链路、异步任务与 AI 编排方式。

3. [`02-api-definition.md`](./02-api-definition.md)  
   REST API 约定、公共对象模型、主要接口的 request/response、状态码与幂等策略。

4. [`03-db-definition.md`](./03-db-definition.md)  
   PostgreSQL 逻辑模型、核心表结构、索引、向量检索、审计与隐私请求落库方案。

5. [`04-metrics-observability.md`](./04-metrics-observability.md)  
   产品指标、系统指标、AI 质量指标、SLO / SLI、告警与仪表盘。

6. [`05-logging-standard.md`](./05-logging-standard.md)  
   结构化日志规范、字段要求、敏感信息脱敏、审计日志、示例日志。

7. [`06-event-contracts.md`](./06-event-contracts.md)  
   内部事件契约、outbox 模式、异步消费方、幂等与版本化。

## 采用的默认实现决策

这些是为了形成一致文档而做出的默认技术决策，可在评审后替换，但文档内各章节已保持一致：

- 前端：**React + TypeScript** 单页 Web 应用，推荐 `TanStack Query + Zustand + Zod`。
- 后端：**Go 模块化单体**（API 进程 + Worker 进程），推荐 `chi + pgx/sqlc + Redis + Asynq`。
- API：**REST JSON**，`/api/v1` 为基础路径，**OpenAPI 作为契约源头**。
- 数据库：**PostgreSQL 16 + pgvector**，P0 / P1 不引入独立向量库。
- 对象存储：**S3 兼容对象存储**，前端通过预签名 URL 上传文件。
- 可观测性：**OpenTelemetry + Prometheus + Grafana + Loki / ELK + Sentry**。
- 日志：**结构化 JSON 日志**，禁止把原始简历、原始 JD、完整用户回答直接写入应用日志。

## 历史阅读顺序（非当前实施顺序）

下面顺序只用于理解 2026-04-26 旧技术包的内部依赖，不再用于当前实施排期或契约生成。当前实施必须先读取 `docs/spec/product-scope/spec.md`、`docs/ui-design/` 与 Layer B/F active spec，再以 `openapi/`、`shared/`、`migrations/`、`config/` 等已编码 truth source 为准。

1. 可先阅读 `00-shared-conventions.md`，了解历史命名、ID、错误码和枚举背景；当前枚举 / 错误码以 `shared/conventions.yaml` 与 `docs/spec/shared-conventions-codified/spec.md` 为准。
   - JSON 命名、DB 命名、事件命名
   - ID 规范
   - 枚举值
   - 错误码
   - 版本化要求

2. `02-api-definition.md` 只能作为历史 API 输入；当前联调契约以 `docs/spec/openapi-v1-contract/spec.md` 与 `openapi/openapi.yaml` 为准。

3. `03-db-definition.md` 只能作为历史 DB 输入；当前迁移、表数量、字段和索引以 `docs/spec/db-migrations-baseline/spec.md` 与 `migrations/` 为准。

4. `04-metrics-observability.md` 与 `05-logging-standard.md` 只能作为历史观测输入；当前 metric / log / dashboard 契约以 `docs/spec/observability-stack/spec.md` 和后续 F1 编码 truth source 为准。

5. `06-event-contracts.md` 与 `01-technical-architecture.md` 只能作为历史异步和架构输入；当前事件 / job / outbox 契约以 `docs/spec/event-and-outbox-contract/spec.md`、`shared/events.yaml` 与 `shared/jobs.yaml` 为准。

## 本包历史覆盖范围

本包曾以旧版 P0/P1 设想为目标，覆盖过以下主链路；其中已被当前产品 spec 与 UI 删除的模块不得从本包恢复为实现范围：

- JD 导入 / 解析
- 目标岗位工作台
- 模拟面试计划 / 会话 / 逐题事件
- 报告生成
- 错题本与复练（已删除独立错题本模块；当前只保留报告内题目回顾与本轮复练）
- 简历定制
- 真实面试复盘
- 数据导出 / 删除请求
- Prompt / Rubric / 模型调用版本追踪

以下方向只做可扩展预留，不作为 P0 上线阻塞项：

- 语音 / 视频训练
- 复杂多租户 Team 版
- 大规模外部情报抓取
- 多模型动态路由的高级调度策略

## 待评审的 5 个决策点

文档已按默认方案写好，但建议上线前尽快拍板：

1. **认证方案**：自建 passwordless / 第三方 OIDC / Clerk / Auth0。
2. **异步编排**：继续使用 Asynq，还是切到 Temporal。
3. **分析平台**：PostHog 直连，还是 Segment + Warehouse。
4. **云部署目标**：AWS ECS / Kubernetes / Fly.io / 自托管。
5. **隐私与合规节奏**：P0 是否就要做完整导出流程，还是先保证删除与数据可见性。

## 输出物建议

若团队按此文档推进，建议下一步立即补 3 个工程产物：

- `openapi/easyinterview.v1.yaml`
- `migrations/0001_init.sql`
- `docs/adr/` 下的 3~5 个 ADR（认证、队列、对象存储、分析平台）
