# Engineering Roadmap Spec

> **版本**: 1.8
> **状态**: active
> **更新日期**: 2026-04-27

## 1 背景与目标

easyinterview 已沉淀三类输入资产：

- **产品真理源**：`easyinterview-spec-v1-0.md`，定义 5 大模块（M1 Profile / M2 Target Job / M3 Practice / M4 Review / M5 Growth）、P0–P3 阶段范围与伦理红线。
- **技术真理源**：`easyinterview-tech-docs/00`–`06`，覆盖共享约定、技术架构、API（32+ 端点 / 14 tag）、DB（29 表 / 9 域）、指标、日志、事件契约（18 个）。
- **UI 真理源**：`easyinterview-ui/`（CDN React + JSX 原型，28+ 屏，含 `easyinterview-canvas.html` 设计画板与 `src/app.jsx` 路由总览）。

工程体量大，若直接逐域开 spec 会出现：spec 边界不清、依赖不明、前后端缺契约层无法并行、mock-first 自验证缺统一 fixtures 源、横切关注点（observability / prompt registry / privacy）被分散埋没。

本 spec 不实现任何业务功能，只承担**项目分解与排期**这一项工程治理职责，目标是：

1. 把整个工程拆成 ~38 份**边界清晰的 child subspec**，每个 child 都能由独立 owner 串行交付。
2. 给出 child 之间的**依赖 DAG**与 6 个**实施 wave**的硬同步点。
3. 锁定**前后端 mock-first** 集成策略：契约 fixtures 同源、单元测试 stub provider 自验、E2E gate 触发真集成。
4. 把 6 项历史悬而未决的技术方案（认证 / 异步编排 / 分析 / 云部署 / 隐私节奏 / AI 网关与模型路由）固定为 W0 hard gate，避免 W2 业务域反复 rebase。

## 2 范围

### 2.1 In Scope

- 38 份 child subspec 的命名、职责一行描述、上游依赖、阶段（P0/P1/P2）、估算 plan 数。
- 6 层结构（A Foundation / B Contract / C Backend / D Frontend / E Integration / F Quality 横切）的边界与协作约束。
- 6 个 wave 的同步点定义（W0–W5）与每个 wave 的准入 gate。
- mock-first 集成策略：fixtures 来源、msw 注入、unit-test stub provider、3 次集成节点（W2 末 / W4 / W5）。
- 6 项 W0 hard gate 决策清单与各自的默认值（默认值仅作为 ADR 起点，不锁结论）。
- `docs/spec/INDEX.md` 与 `docs/work-journal/` 的同步约定。

### 2.2 Out of Scope

- 任何 child subspec 的 spec.md / plan / checklist 的具体内容（由后续 wave 按本 spec 唯一的 plan 触发 spawn）。
- 6 项 W0 决策的 ADR 正文（仅在本 spec §3 列出待决策项与默认值）。
- 任何代码（含 `repo-scaffold` 的 makefile、`local-dev-stack` 的 docker-compose；这些归对应 child subspec 的 plan）。
- 修改 `easyinterview-spec-v1-0.md` / `easyinterview-tech-docs/` / `easyinterview-ui/` 这三个真理源（本 spec 只读引用）。
- 任何业务域级评估、性能基线、合规审计（归 Layer F 或对应业务 subspec）。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策（由 plan-mode 批准）

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 顶层 subspec 目录名 | `engineering-roadmap` | 文件路径 / INDEX 标识 |
| D-2 | 分层结构 | 6 层（A Foundation / B Contract / C Backend / D Frontend / E Integration / F Quality 横切） | child 总数 38；Layer F 横切自始并行 |
| D-3 | 现有 UI 原型角色 | 作为 mock fixtures 与视觉真理源 | `src/data.jsx` 折成 OpenAPI fixtures；`screens-*.jsx` 为视觉/交互参考；前端 spec 用 React+TS 重做但保留 `easyinterview-ui/` 不动 |
| D-4 | 6 项历史悬而未决方案 | 设为 W0 hard gate（每项产出 1 份 ADR） | W0 不出 ADR 则不进 W1 |

### 3.2 W0 已锁定决策（hard gate · 全部 accepted）

6 项 W0 hard gate 在 2026-04-26 全部签字，ADR 固定落点 `docs/spec/engineering-roadmap/decisions/ADR-Q{n}-*.md`，由本 roadmap subject 直接承接，不另起 sibling spec。本表只承载锁定结论与影响范围；推翻或升级的具体阈值见各 ADR §5「失效与修订条件」。

| ID | 决策项 | 锁定结论 | ADR | 影响 child subspec |
|----|--------|----------|-----|-------------------|
| Q-1 | 认证方案 | 自建 passwordless email magic link + first-party session cookie；OIDC / 托管 Auth 推迟到 P1 / 团队版触发条件出现 | [ADR-Q1](./decisions/ADR-Q1-auth.md) | C1 `backend-auth`、D1 `frontend-shell`、B2、B4、C8、F1、F4 |
| Q-2 | 异步编排 | Asynq + Redis 作为唯一异步 runtime；PG outbox + dispatcher 进程；critical/default/low 三级队列；Temporal 推迟到出现跨日 SLA / 多步补偿 | [ADR-Q2](./decisions/ADR-Q2-async-orchestration.md) | C8 `backend-async-runtime`、A2、B3、B4、C4-C7、F1、A4 |
| Q-3 | 分析平台 | 自托管 PostHog 作为唯一产品分析后端；不依赖 PostHog Cloud / Segment / Warehouse；F2 adapter 抽象、不入侵业务代码；feature flag 使用自托管 PostHog；部署方式由 F2 / E4 验证可运维路径，普通本地 dev 默认 no-op / file-backed mode | [ADR-Q3](./decisions/ADR-Q3-analytics-platform.md) | F2 `analytics-funnel`、D1、C8、A4、F4、E4、B1、F1、A2 |
| Q-4 | 云部署目标 | Kubernetes（managed cluster；staging/prod；本地场景测试 = Kind）；普通本地开发走 A2 docker-compose；3 deployment + helm chart 与 `test/scenarios/` 同源；Kind 直连真实 AI provider endpoint，staging/prod 可指 cluster-internal AI Gateway；当前单人阶段不构建 CI pipeline，部署自动化归 E4 后续；ArgoCD/FluxCD/SOPS 等具体工具不锁定 | [ADR-Q4](./decisions/ADR-Q4-cloud-deploy-target.md) | A2、A5、E4、F1、A3、A4、`test/scenarios/` |
| Q-5 | 隐私节奏 | P0 仅落地删除链路（24h SLA + 跨 17 表硬删 + audit）；导出延后到 P1；`POST /privacy/exports` 在 v1.0.0 freeze 中预留并返回 501；这是对产品 P0「删除与导出路径可用」验收项的 W0 例外，必须在 release gate 中显式记录 | [ADR-Q5](./decisions/ADR-Q5-privacy-cadence.md) | C12 `backend-privacy`、F4 `privacy-and-audit-runtime`、B2、B4、C8、D1、D6、F1、E4 |
| Q-6 | AI 网关与模型路由 | 应用内 `AIClient` + Model Profile + unit-test stub provider；docker compose / Kind / staging / prod 通过 `AI_GATEWAY_BASE_URL` 指向真实 OpenAI-compatible provider 或生产 gateway；fallback / cost cap / rate limit 全部归 provider/gateway 配置；业务零厂商 SDK 入侵 | [ADR-Q6](./decisions/ADR-Q6-ai-gateway-and-model-routing.md) | A3 `ai-gateway-and-model-routing`、A4、F1、F3、C4-C7、C9、C11、C14、E4、B1 |

ADR 推翻或升级时，新 ADR 显式标注 `supersedes: ADR-Q{n}-*.md`，本表对应行同步更新「锁定结论」与 ADR 链接。

### 3.3 已知边界争议（在对应 child spec 中处理，非本 spec 决策）

- C5 `backend-practice` 的 turn-light-review（同步轻量观察）是否独立成 child？默认仍归 C5 内部 plan；如果在 W3 体量超阈值再升格。
- D3 `frontend-workspace-and-practice` 的 4-plan 内部拆分（workspace / practice-core / practice-modes / followup-and-star）由 D3 自己的 plan 文档定义，本 spec 不锁。

## 4 设计约束

### 4.1 文档治理约束

- **Spec owns plan**：每个 child subspec 的 plan 必须挂在 `docs/spec/${child}/plans/` 之下；不得创建 `docs/plan/`。
- **原地修订**：同主题后续修订优先原地更新原 spec 与原 plan，不创建 sibling bugfix/follow-up 目录。
- **设计先行**：任何 child 进入 `/implement` 之前必须先有自己的 spec.md + 至少一个 plan + checklist + context.yaml。
- **顺序执行**：默认串行 phase；并行只发生在不同 child subspec 之间（同一 wave 内）。
- **Header 一致性**：每份 child 的 spec.md 与各 plan 的 plan.md / checklist.md Header 字段顺序固定（版本 / 状态 / 更新日期），状态枚举只取 `draft`/`active`/`completed`/`superseded`/`deprecated`。

### 4.2 真理源约束

- 产品决策与不可逾越红线由 `easyinterview-spec-v1-0.md` 决定，child spec 不得与之冲突。
- 技术契约由 `easyinterview-tech-docs/00`–`06` 决定，child spec 通过 Layer B 的 4 份 contract spec 间接引用，禁止绕过 contract 直接引用未编码约定。
- UI 视觉与交互由 `easyinterview-ui/easyinterview-canvas.html` 决定，前端 child spec 重做 React+TS 时可以重组件库但不得改变页面语义。

### 4.3 mock-first 集成策略

- **前端 mock**：用 `msw`（Mock Service Worker）拦截 fetch；数据来源是 B2 OpenAPI 的 `fixtures/`（按 14 tag 拆 JSON 集合）。**禁止前端 hardcode mock**。`easyinterview-ui/src/data.jsx` 折成 fixtures 的一个命名场景（`scenario: prototype-baseline`）。
- **后端 AI 路由**：A3 `ai-gateway-and-model-routing` 必须交付 provider-neutral `AIClient`、Model Profile 配置、OpenAI-compatible provider/gateway route 与 unit-test `stub`。P0 业务代码只依赖 `AIClient` 与 profile name，不 import Higress SDK 或厂商 SDK；docker compose / Kind / staging / prod 通过 `AI_GATEWAY_BASE_URL` 指向真实 AI provider 或生产 gateway，由配置决定文本、图像、音频、embedding 等场景的 provider / model / fallback / token rate limit。`stub` 输入→输出确定性映射（hash-based），可被 OpenAPI fixtures 反向喂养，但仅用于单元测试、离线契约测试或显式 mock 场景。
- **集成节点**：
  - **W2 末**（软集成）：前端从各自 fixtures 切到 E1 `mock-contract-suite`，前后端 mock 同源。
  - **W4**（硬集成）：每个前端 child 的 `003-integration` plan 切到 W3 跑通的真后端，由 E2 `e2e-scenarios-p0` 担任 BDD-Gate。
  - **W5**（上线集成）：staging 灰度 + 回滚演练，由 E4 `release-gate-and-rollout` 担任 gate。

### 4.4 Layer F 横切约束

- F1 `observability-stack`、F3 `prompt-rubric-registry` 必须从 W1 起就进入；W1 parent phase 先锁 F3 `feature_key + version` 契约，W2 业务域只有在 F3 child `001` 验证 baseline prompt / rubric 文件后才能引用 prompt id，任何阶段都不得在自己 spec 中 hardcode prompt 文本。
- F2 `analytics-funnel` 在 W2 起步埋点 stub，不阻塞业务实现。
- F4 `privacy-and-audit-runtime` 维持 P1 进入，与 C12 `backend-privacy` 协同；[ADR-Q5](./decisions/ADR-Q5-privacy-cadence.md) 锁定 P0 = 删除-only，C12 / F4 不升格 P0；P0 删除链路核心实现下沉到 C8 `backend-async-runtime` 的 `privacy_delete` public `jobType`（内部 Asynq handler 可映射为 `privacy.delete`）；audit_events / privacy_requests schema 在 W1 由 B4 / B1 锁定，W4 release-gate 校验删除 SLA 与 audit 完整性，并记录 P0 导出能力延后这一 W0 例外。

## 5 模块边界

本 spec 把工程拆成 6 层 38 份 child subspec。每个 child 是独立 spec 主题（`docs/spec/${child}/`），自己挂 plan 与 checklist。下表列出全集；child 内部 plan 数为估算上限。

### 5.1 Layer A · Foundation（5 份，全部 P0）

| ID | Subspec | 一行职责 | 上游依赖 | Plan 数 |
|----|---------|---------|---------|---------|
| A1 | `repo-scaffold` | monorepo 目录骨架（`backend/`、`frontend/`、`openapi/`、`migrations/`、`scripts/`）、根 makefile、git hooks、`.editorconfig`、`.tool-versions` | – | 1 |
| A2 | `local-dev-stack` | docker-compose：Postgres+pgvector / Redis / MinIO + 当前项目可运行组件；`make dev-up` 一键启动本地环境 | A1 | 1 |
| A3 | `ai-gateway-and-model-routing` | provider-neutral `AIClient` + Model Profile + OpenAI-compatible provider/gateway route + unit-test `stub` provider；本地部署直连真实 AI provider，Higress 等 AI Gateway 作为 staging/prod 独立部署组件接入 | A1 | 2 |
| A4 | `secrets-and-config` | 配置分层（`.env.example` / `config.yaml` / env override）、secret manager 抽象、feature flag 文件源 | A1 | 1 |
| A5 | `ci-pipeline-baseline` | 单人阶段的本地质量门禁（lint/test/build/docs/codegen check）；远端 CI pipeline、branch protection、artifact 延后 | A1, A2 | 1 |

### 5.2 Layer B · Contract（4 份，全部 P0）

| ID | Subspec | 一行职责 | 上游依赖 | Plan 数 |
|----|---------|---------|---------|---------|
| B1 | `shared-conventions-codified` | 把 `00-shared-conventions.md` 落到代码：Go types / TS types / ID 工具 / 错误码常量 / 枚举 / `UPPER_SNAKE_CASE` lint | A1 | 2 |
| B2 | `openapi-v1-contract` | 32+ 端点的 OpenAPI 3.1，`/api/v1/...`，14 tags，含 fixtures；前后端 codegen 入口 | B1 | 3 |
| B3 | `event-and-outbox-contract` | 18 事件 envelope + payload schema，outbox 表 schema，dispatcher 协议 | B1 | 1 |
| B4 | `db-migrations-baseline` | 29 表初始迁移 + pgvector 扩展 + 索引；`golang-migrate` / `atlas` 选型 | B1, A2 | 1 |

### 5.3 Layer C · Backend（14 份；P0:8 / P1:4 / P2:2）

业务域纵切（模块化单体）。每个 child 默认 3-5 个 plan：`spec-locked` / `contract-fragment`（贡献到 B2 切片）/ `impl` / `unit-test` / `mock-server`。后两个对纯 CRUD 域可合并。

| ID | Subspec | 一行职责 | 上游依赖 | Plan 数 | 阶段 |
|----|---------|---------|---------|---------|------|
| C1 | `backend-auth` | auth + session + `/me` + 用户上下文中间件 | B1, B2 | 3 | P0 |
| C2 | `backend-upload` | 预签名上传 + `file_objects` + purpose 枚举 + 扫描 hook | B2, B4, A2 | 3 | P0 |
| C3 | `backend-profile` | CandidateProfile + ExperienceCard + 偏好 | B2, B4, C1 | 3 | P0 |
| C4 | `backend-targetjob` | TargetJob + JD 导入 + 解析状态 + 来源记录 + async job 编排 | B2, B3, B4, A3, C2, C3 | 5 | P0 |
| C5 | `backend-practice` | PracticePlan + Session + Event 流 + Turn 物化 + 状态机 + 同步 AI 首题/追问 | B2, B3, B4, A3, C4 | 5 | P0 |
| C6 | `backend-review` | QuestionAssessment + FeedbackReport + MistakeEntry + 异步报告生成 | B3, B4, A3, C5 | 5 | P0 |
| C7 | `backend-resume` | ResumeAsset + 解析 + `/resume/tailor` 异步定制 | B2, B3, B4, A3, C2 | 4 | P0 |
| C8 | `backend-async-runtime` | Asynq + worker 进程骨架 + outbox dispatcher + job 表 + 重试 + 幂等 | B3, A2 | 3 | P0 |
| C9 | `backend-debrief` | 真实面试复盘 + Debrief 异步生成 + 感谢信草稿 | B2, B3, B4, A3, C4, C6 | 4 | P1 |
| C10 | `backend-growth` | 成长看板聚合 + `/growth/overview` + 维度趋势 | B4, C5, C6 | 3 | P1 |
| C11 | `backend-retrieval` | pgvector embedding upsert + 相似题召回 + 跨实体检索 | B4, A3, C3, C4, C6, C9 | 3 | P1 |
| C12 | `backend-privacy` | 导出 / 删除请求 + 隐私 worker + 审计联动 | B4, C8 | 3 | P1 |
| C13 | `backend-source-intel` | 轻量公司情报 + `source_records` + freshness 调度 | B4, C4, C11 | 3 | P2 |
| C14 | `backend-voice-stt` | 语音模式 STT 适配 + 媒体存储 + retention | A3, C2, C5 | 3 | P2 |

### 5.4 Layer D · Frontend（7 份；P0:4 / P1:2 / P2:1）

每个 child 默认 2-3 个 plan：`spec-locked` / `impl-with-mock` / `integration`（拨真 API）。

| ID | Subspec | 一行职责 | 关键路由 | 上游依赖 | Plan 数 | 阶段 |
|----|---------|---------|---------|---------|---------|------|
| D1 | `frontend-shell` | App 壳 + 路由 + TopBar + 主题（warm/dark）+ 双语 i18n + auth gate + 全局 store/query | `welcome` / topbar / 切主题 | A1, B1 | 3 | P0 |
| D2 | `frontend-onboarding-and-target` | M0 欢迎+登录 + Home 收件箱 + Parse 进度 + Onboarding 画像 + JD-Match | `welcome` / `home` / `parse` / `onboarding` / `jd_match` | D1, B2 | 3 | P0 |
| D3 | `frontend-workspace-and-practice` | M2 Workspace + M3 Practice 全套（5 模式 + followup-tree + drill + STAR editor）。**内部 4 plan 拆分**：workspace / practice-core / practice-modes / followup-and-star | `workspace` / `practice` / `followup` / `drill` / `star` | D1, B2 | 4 | P0 |
| D4 | `frontend-review-and-mistakes` | M4 Report（3 layouts）+ 异步 Generating + Mistakes 列表 + Drill builder | `report` / `generating` / `mistakes` | D1, B2 | 3 | P0 |
| D5 | `frontend-resume-and-experiences` | 简历工坊单页 + 多版本对比 + 经历库 | `resume` / `resume_versions` / `experiences` | D1, B2 | 2 | P1 |
| D6 | `frontend-debrief-and-growth` | 真实复盘（简版+完整版+感谢信）+ 成长看板 + Settings/隐私 | `debrief` / `growth` / `settings` | D1, B2 | 3 | P1 |
| D7 | `frontend-voice-and-plan` | 语音练习 + 多轮计划 + Company Intel | `voice` / `plan` / `company_intel` | D1, B2 | 2 | P2 |

### 5.5 Layer E · Integration（4 份）

| ID | Subspec | 一行职责 | 上游依赖 | Plan 数 | 阶段 |
|----|---------|---------|---------|---------|------|
| E1 | `mock-contract-suite` | B2 fixtures 转可运行 mock server（Prism / 自建）+ 后端 mock-server plan 统一壳 | B2 | 1 | P0 |
| E2 | `e2e-scenarios-p0` | 跨前后端 P0 主漏斗 8 步：导入→工作台→练习→报告→错题→复练 | C4, C5, C6, C7, D2, D3, D4, E1 | 1 | P0 |
| E3 | `e2e-scenarios-p1` | 真实复盘漏斗 + 简历定制漏斗 + 多语言场景 | C9, C10, C11, D5, D6, E2 | 1 | P1 |
| E4 | `release-gate-and-rollout` | 灰度开关、版本兼容、回滚 runbook、SLO 准入；按 [ADR-Q5](./decisions/ADR-Q5-privacy-cadence.md)（P0 = 删除-only · 24h SLA）与 [ADR-Q6](./decisions/ADR-Q6-ai-gateway-and-model-routing.md)（AI Gateway 路由可观测 + fallback / cost cap）校验上线门槛 | F1, F2, F3, E2 | 1 | P0+持续 |

### 5.6 Layer F · Quality 横切（4 份）

| ID | Subspec | 一行职责 | 上游依赖 | Plan 数 | 阶段 |
|----|---------|---------|---------|---------|------|
| F1 | `observability-stack` | 应用 `/metrics` / OTel SDK / Sentry 接线、route/job/AI metrics、access log、5 个 dashboard 与生产观测配置 | A2, B1 | 2 | P0 |
| F2 | `analytics-funnel` | 18 产品事件 + 3 漏斗 + 自托管 PostHog adapter；前后端双发去重；部署路径不得依赖 PostHog Cloud | B1, D1 | 2 | P0 |
| F3 | `prompt-rubric-registry` | Prompt / Rubric / Model Profile 版本表 + 灰度 + 离线评估集（≥50 题）+ LLM Judge | A3, B4 | 3 | P0+持续 |
| F4 | `privacy-and-audit-runtime` | 审计事件 + 留存策略 + 删除/导出可观测性 + 字段红线 lint | B4, C12 | 2 | P1 |

### 5.7 实施 Wave 顺序

| Wave | 周期估算 | 范围 | 包含 child | 准入 gate（W 末） |
|------|---------|------|-----------|------------------|
| **W0** | 1 周 | 共识与骨架 | A1, B1，本 spec freeze | 6 项 ADR 全部签字；A1/B1 scaffold 与 context 校验通过 |
| **W1** | 1.5 周 | 基础设施 + 契约骨架（spec only，不写 impl plan） | A2, A3, A4, A5, B2, B3, B4, F1, F3 | 9 spec 完成 parent-level cross-spec review；A2/B2/F1/F3 的 spec-contract lock 就绪（dev-up 契约 / OpenAPI freeze 范围 / metric 字典 / prompt key 字典）；A5 仅锁本地质量门禁与 deferred CI 边界；可执行 gate 交由各 child `001` plan 逐一闭合，未闭合前不得启动依赖它的 W2 implementation |
| **W2** | 3-4 周 | 前后端 mock-first 并行（最大并行波） | C1, C2, C3, C8, E1（后端）+ D1, D2, D3, D4（前端）+ F2 | E1 提供 14 tag 全 mock；前端跑通 P0 漏斗 happy path（mock）；后端 mock-server 自验证 |
| **W3** | 3 周 | 核心业务域后端 | C4, C5, C6, C7；F3 接入真实 Model Profile + ≥50 题离线评估集 | 6 P0 后端域 unit + mock-server BDD 通过 |
| **W4** | 1.5 周 | 真集成 | D2/D3/D4 的 `003-integration` plan + E2 + F1/F2 接齐 | E2 全场景通过 |
| **W5** | 1 周 | 上线 gate | E4 + 04 文档 §15 最低上线门槛 | P0 准入 |

**P0 总周期估算**：约 10-11 周，与产品 spec 给出的 8-10 周节奏对齐。

P1 / P2 child（C9-C14、D5-D7、E3、F4）按相同结构进入 Wave 6+，由本 spec 唯一的 plan 在 P0 收尾后追加新 phase 调度。

### 5.8 关键依赖（DAG 文字版）

```
F (横切)：F1 / F2 / F3 与 P0 A→B→C/D 同步推进；F4 默认 P1，除非 Q-5 ADR 将完整隐私链路升格为 P0

A1 ─┬─► A2 ─┬─► A4
    ├─► A3 ─┘
    └─► A5 ─────────► (local quality gate; remote CI deferred)

A1 ─► B1 ─┬─► B2 (OpenAPI)   ← 整个 DAG 最关键瓶颈节点
          ├─► B3 (Events)
          └─► B4 (DB)

C1 auth         ◄── B1, B2
C2 upload       ◄── B2, B4, A2
C3 profile      ◄── B2, B4, C1
C4 targetjob    ◄── B2, B3, B4, A3, C2, C3
C5 practice     ◄── B2, B3, B4, A3, C4
C6 review       ◄── B3, B4, A3, C5
C7 resume       ◄── B2, B3, B4, A3, C2
C8 async-runtime ◄── B3, A2  (横切 backend，多个 C 域共用)
C9 debrief      ◄── C4, C6
C10 growth      ◄── C5, C6
C11 retrieval   ◄── C3, C4, C6, C9
C12 privacy     ◄── C8

D1 shell        ◄── A1, B1
D2-D7           ◄── D1, B2  (除 D1 外，前端只跨依赖 B2 fixtures——这是 mock-first 能并行的根本)

E1 mock-suite   ◄── B2
E2 e2e P0       ◄── C4, C5, C6, C7, D2, D3, D4, E1
E3 e2e P1       ◄── C9, C10, D5, D6, E2
E4 release-gate ◄── F1, F2, F3, E2
```

**3 个关键观察**：

1. **B2（OpenAPI）是 DAG 瓶颈节点**：C 全域和 D 全域都直接依赖。一旦 codegen 投产，破坏性变更会触发跨 spec 雪球。W1 parent phase 先锁 v1.0.0 freeze 范围与 additive-only 规则；`openapi/openapi.yaml` 与 breaking change linter 由 B2 child `001` 验证后，才能放行依赖 B2 的 W2 implementation。
2. **C8（async-runtime）+ F3（prompt-rubric-registry）是横切型基础**：必须先于 C4/C5/C6 起来，否则业务域会偷偷 hardcode prompt 与 outbox 行为。
3. **D1 之外的前端 child 几乎只依赖 B2 fixtures**：是 mock-first 能在 W2 大并行的结构性原因。

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|----------|
| C-1 | 顶层 spec freeze | 本 spec §3.2 全部 6 项 ADR 签字（已于 2026-04-26 完成） | W0 收尾 | 本 spec 状态可保持 `active`，进入 W1 | 001 Phase 1-2 |
| C-2 | W1 spec 契约锁定 | A1 + B1 完成，A2-A5 / B2-B4 / F1 / F3 共 9 份 spec 已 spawn | W1 末 / 001 Phase 3 | 9 份 spec 的 parent-level cross-spec review 证据已记录；A2/B2/F1/F3 的 spec-contract lock 已在各自 spec 中固定；A5 明确当前只做本地质量门禁且远端 CI 延后；`make dev-up`、`openapi/openapi.yaml` v1.0.0、baseline prompt 文件等可执行 gate 不在 parent Phase 3 中冒充完成，必须由对应 child `001` plan 验证通过后才能放行依赖它的 W2 implementation | 001 Phase 3 + 后续 child `001` plans |
| C-3 | mock-first 软集成 | E1 提供 14 tag 全 mock | W2 末 | 前端 4 域跑通 P0 happy path（mock）；前后端 mock 同源（fixtures 同一份） | 001 Phase 4 |
| C-4 | 业务域 ready | C4–C7 实现完毕 | W3 末 | 6 个 P0 后端域 unit + mock-server BDD 通过；F3 接入真实 Model Profile + ≥50 题离线评估集 | 001 Phase 5 |
| C-5 | 真集成贯通 | D2/D3/D4 切真 API | W4 末 | E2 `e2e-scenarios-p0` 全场景通过 | 001 Phase 6 |
| C-6 | 上线 gate | E4 release-gate 跑完 staging 灰度 + 回滚 | W5 末 | `04-metrics-observability.md` §15 最低上线门槛全勾；[ADR-Q5](./decisions/ADR-Q5-privacy-cadence.md) 决定的 P0 删除链路 24h SLA 与 audit 已验证，且导出延后例外已作为 W0 tradeoff 记录；[ADR-Q6](./decisions/ADR-Q6-ai-gateway-and-model-routing.md) 的 AI provider/gateway 路由 / fallback 可观测；P0 准入 | 001 Phase 6 |
| C-7 | 收尾归档 | P0 全部上线 | W5 后 | 本 spec 状态由 `active` 调整为 `completed`；P1/P2 child draft spec 创建；触发 `/retrospective` | 001 Phase 7 |

## 7 关联计划

- [001-decompose-subspecs](./plans/001-decompose-subspecs/plan.md)：按 6 wave spawn P0 child 并通过各自 review；P1/P2 child 在 P0 收尾创建 draft spec，进入 Wave 6+ 前再补齐 plan / checklist / context。这是本 spec 唯一挂的 plan；其他治理类 plan（灰度、release-gate、隐私）归入对应的 child（E4 / F4）。
