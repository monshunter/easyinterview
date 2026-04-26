# easyinterview 技术实施文档包（P0 / P1）

本目录基于产品 spec 的 P0 / P1 主链路整理，面向前后端分离实现落地。文档目标不是复述产品需求，而是把需求收敛成一套可以直接进入开发、联调、评审和上线治理的技术约束。

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

## 使用顺序

建议按下面顺序进入实施：

1. 先确认 `00-shared-conventions.md`，尤其是：
   - JSON 命名、DB 命名、事件命名
   - ID 规范
   - 枚举值
   - 错误码
   - 版本化要求

2. 再以 `02-api-definition.md` 为联调主文档，产出 OpenAPI。

3. 后端根据 `03-db-definition.md` 建模与迁移，前端根据 `02-api-definition.md` 生成 SDK 和类型。

4. 上线前补齐 `04-metrics-observability.md` 与 `05-logging-standard.md` 的埋点、日志、告警。

5. Worker / 异步逻辑按 `06-event-contracts.md` 与 `01-technical-architecture.md` 实现。

## 本包覆盖范围

本包以 P0 必做、P1 可兼容扩展为目标，重点覆盖以下主链路：

- JD 导入 / 解析
- 目标岗位工作台
- 模拟面试计划 / 会话 / 逐题事件
- 报告生成
- 错题本与复练
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
