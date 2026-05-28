# Engineering Roadmap Spec

> **版本**: 3.22
> **状态**: active
> **更新日期**: 2026-05-28

## 1 背景与目标

`engineering-roadmap` 曾承担一次性拆出 38 个 child subspec、按 W0-W5 wave 逐步 spawn 的顶层规划职责。随着 `docs/spec/product-scope/spec.md`、`docs/ui-design/` 与 `ui-design/` 在 2026-05-03 收敛，旧 roadmap 的几个假设已经不适合继续作为执行入口：

- 当前产品真理源明确：未被当前 UI / UI 文档保留、重定义或列为规划例外的旧能力默认已丢弃。
- 当前 UI 一级入口只保留 `首页 / 岗位推荐 / 模拟面试 / 简历 / 复盘`，报告和语音都是上下文能力，不是一级模块。
- `docs/spec/INDEX.md` 应投影真实存在的 spec，而不是承载大量 `_pending_` backlog 占位。
- 已落地的 A/B/F 工程契约已经有 active spec 与编码 truth source，后续实现应直接引用这些真理源，而不是按旧 wave 重新生成空壳文档。

因此，本 spec 重新定义 engineering roadmap 的职责：

1. 固定当前产品、UI、契约和已编码资产之间的真理源关系。
2. 保留仍有效的基础设施、契约和质量治理 spec，删除未创建 child 的 pending 索引模型。
3. 给出当前 P0 MVP 闭环的实施地图和依赖顺序，只在真正进入设计或实现时创建对应 child spec / plan。
4. 明确哪些旧规划、旧 route、旧模块和旧技术文档不得作为恢复功能的依据。

## 2 范围

### 2.1 In Scope

- 当前 active spec 与已编码 truth source 的工程边界。
- 当前 P0 MVP 闭环的实施 workstream：JD 导入、岗位推荐、模拟面试规划、完整面试 session、报告 Dashboard、简历工坊、真实面试复盘、认证 / 设置 / 用户画像、mock / E2E / release gate。
- 后续 child spec / plan 的创建规则、依赖顺序和质量门禁。
- 已完成 ADR-Q1..Q6 的持续约束：认证、异步编排、分析平台、云部署、隐私节奏、AI 网关与模型路由。
- `docs/spec/INDEX.md` 与 `plans/INDEX.md` 作为 Header 投影视图的治理规则。

### 2.2 Out of Scope

- 不在本 spec 中编写具体 API schema、DB schema、事件 payload、prompt 文本、UI 组件代码或 BDD 场景脚本。
- 不为尚未启动的 P1/P2 方向创建 draft spec、pending INDEX 行或空 plan。
- 不恢复旧根 spec、旧 route、旧画板标签或历史组件中出现但当前产品 / UI 已丢弃的能力。
- 不保留已被当前 owner spec / coded truth source 承接的技术草稿目录或文件。
- 不新增 sibling roadmap plan；本 subject 继续原地修订 `001-decompose-subspecs`。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 产品真理源 | `docs/spec/product-scope/spec.md` | child spec / plan 不能绕过产品范围恢复旧能力 |
| D-2 | UI 真理源 | `docs/ui-design/` + `ui-design/` | 前端与端到端流程以当前静态 UI 和 UI 文档为准 |
| D-3 | 技术契约真理源 | [product-scope §1.5](../product-scope/spec.md#15-技术契约-owner-matrix) 定义的 Layer A/B/F owner spec + 已编码 truth source（`openapi/`、`shared/`、`migrations/`、`config/`） | 后续实现必须复用现有契约；技术契约 owner matrix 由 product-scope 持有 |
| D-4 | INDEX 语义 | `docs/spec/INDEX.md` 只记录真实存在的 spec | 删除 pending 占位；未 spawn 的 subject 只能在 roadmap 正文中作为候选 workstream 描述 |
| D-5 | child 创建策略 | 只在进入设计或实现时创建 child spec / plan / checklist / context | 避免空 spec、僵尸 plan 和未审清的 P1/P2 占位 |
| D-6 | 当前 P0 前端边界 | Home / Job Picks / Mock Interview / Practice Session / Report Dashboard / Resume / Debrief / User Menu | 不恢复 Welcome、Growth、Plan、Mistakes、Drill、Followup、Experiences、STAR 或独立 Voice 页面 |
| D-7 | ADR-Q1..Q6 | 保留为 engineering-roadmap 的历史架构约束 | 后续推翻必须新增 superseding ADR，并同步本 spec |

### 3.2 ADR-Q1..Q6 当前约束

| ID | 主题 | 当前结论 | 当前落地边界 |
|----|------|----------|--------------|
| Q-1 | 认证方案 | 自建 passwordless email-code + first-party session cookie；邮箱是唯一账号标识，注册/登录合并为单一登录入口 | 认证是操作级拦截；默认入口仍为 Home；首次邮箱验证后如资料未完成先进入资料补全，再恢复 `pendingAction` |
| Q-2 | 异步编排 | B3 job/outbox contract + backend internal runner；Redis/Asynq 仅作为后续 runner implementation option | P0 异步任务服务 JD 解析、报告生成、简历处理、复盘分析和删除链路；不生成独立错题 / Drill 队列，也不要求独立 worker 进程 |
| Q-3 | 分析平台 | 自托管 PostHog，普通本地 dev 可 no-op / file-backed | 分析漏斗围绕导入 -> 规划 -> 练习 -> 报告 -> 复练 / 下一轮 -> 真实复盘 |
| Q-4 | 部署与测试目标 | 当前 P0 锁定 Docker Compose 外部依赖 + 宿主机 app runtime + repo-tracked 本地 scenario runner；Kubernetes / Kind / Helm 不再是默认测试或部署目标 | 部署自动化和 rollout gate 进入后续 release workstream，届时按实际规模重新评估容器编排 |
| Q-5 | 隐私节奏 | P0 删除-only；导出延后并以 501 / UI 占位解释 | 删除链路、audit 和 redaction 是 P0 gate；完整导出归后续隐私增强 |
| Q-6 | AI provider 与模型路由 | 应用内 `AIClient` + Provider Registry + Capability Model Profile + OpenAI-compatible / stub provider | 业务代码只依赖 profile / feature_key，不 import 厂商 SDK；provider/profile/capability drift 由 A3/B1/A4/F3 gate 拦截 |

### 3.3 待确认事项

| ID | 待确认事项 | 默认处理 |
|----|------------|----------|
| Q-R1 | P0 实现是否按前端先行还是后端先行切片 | 默认先做 mock-first：B2 fixtures -> E1 mock -> D1-D6 UI 集成，再切真后端 |
| Q-R2 | 生产级语音是否提前进入 P0 | 默认不提前；P0 保留 UI 语音形式和显式入口，真实 STT、媒体留存、隐私开关归后续语音生产化 workstream |
| Q-R3 | 全球多平台搜岗是否进入近期 roadmap | 默认不进入当前 MVP；只作为 product-scope 明确规划例外保留，另行设计数据源与合规边界 |

## 4 设计约束

### 4.1 产品与 UI 约束

- P0 闭环必须围绕 `JD / 推荐岗位 -> 当前面试规划 -> 完整模拟面试 -> Report Dashboard -> 复练当前轮 / 进入下一轮 -> 真实面试复盘`。
- 顶部导航只能出现当前 UI 真理源确认的五个一级入口。
- `workspace` 的产品语义是当前模拟面试规划，不是旧 `当前岗位` 一级模块。
- `practice` 是文本和语音面试共享的会话页面；语音面试只能通过 `practice?mode=voice&modality=voice` 或等价显式参数进入。
- `report` 必须带 `sessionId` 或等价会话上下文；报告不作为一级导航或无上下文历史中心。
- `debrief` 处理真实面试复盘，文本添加和语音添加共享同一份复盘记录。
- 旧 route key / 旧画板标签 `welcome`、`growth`、`plan`、`mistakes`、`drill`、`followup`、`experiences`、`star`、`resume`、`onboarding`、`voice` 不得作为独立目标模块恢复；当前简历一级模块入口以 UI 文档定义的 `resume_versions` / Resume Workshop 为准。

### 4.2 文档治理约束

- `docs/spec/INDEX.md` 只投影真实 `docs/spec/*/spec.md` Header，不保留 pending 行。
- 任何同主题修订必须优先原地修改既有 spec / plan / checklist / context，不创建 sibling plan。
- 尚未进入设计或实现的 P1/P2 能力不得提前创建空 spec、空 plans/INDEX 或 draft 占位。
- 新增或修订代码逻辑 plan 必须写明 TDD 策略，并通过 `/implement` -> `/tdd` 执行。
- 新增或修订用户可见 UI、API 行为、业务流程或端到端功能 plan 必须维护 BDD gate。
- 删除旧规划时优先删除索引占位和死文档；仍作为历史证据的 completed plan 可保留，但正文必须明确它不是当前执行入口。
- 删除已迁移技术草稿前，必须确认当前文档与代码注释不再引用其目录名或文件名；技术责任必须由当前 owner spec / coded truth source 独立承接。

### 4.3 契约与 mock-first 约束

- 前端 mock 数据来源必须是 B2 OpenAPI fixtures（当前 13 tag / 60 operation；JobMatch tag 12 operation 由 B2 spec D-17 additive 升级纳入，Resumes tag 10 additive operation 由 B2 spec D-18 与 D-23 纳入，Debrief suggestions / PracticeSessions listing / Practice voice turn 分别由 D-20 / D-21 / D-22 纳入，Auth `completeMyProfile` 由 D-25 纳入），禁止前端重新 hardcode product data truth source。
- `ui-design/src/data.jsx` 只能作为 prototype-baseline 场景输入，不能越过 OpenAPI fixtures 直接驱动实现。
- 后端 AI 调用必须通过 A3 `AIClient` 和 F3 prompt/rubric/model profile 契约。
- 业务 spec 不得 hardcode prompt 正文、rubric 文本、模型名、厂商 SDK 或 feature flag 绕过 A3/A4/F3。
- OpenAPI、events、migrations、feature flags 和 runtime config 的破坏性变更必须先修订对应 Layer B/A/F spec 与 drift gate。
- 技术契约分层以 [product-scope §1.5](../product-scope/spec.md#15-技术契约-owner-matrix) 为唯一 owner matrix；本 roadmap 只消费该表，不在子章节复制第二套映射。

## 5 模块边界

### 5.1 当前已存在的 active spec

| 层级 | 历史 ID | Subject | 当前职责 | 是否保留 |
|------|---------|---------|----------|----------|
| 顶层 | - | `product-scope` | 产品范围、阶段边界、丢弃规则和质量红线 | 保留 |
| 顶层 | - | `engineering-roadmap` | 当前实施地图、依赖顺序和文档治理 | 保留 |
| Foundation | A1 | `repo-scaffold` | 仓库骨架、根 Makefile、hook 基础 | 保留 |
| Foundation | A2 | `local-dev-stack` | 本地 dev stack、dev doctor、端口与健康检查契约 | 保留 |
| Foundation | A3 | `ai-provider-and-model-routing` | AIClient、provider registry、capability model profile、DeepSeek V4 Flash/Pro 开发主力、OpenAI-compatible / stub provider、profile coverage lint | 保留 |
| Foundation | A4 | `secrets-and-config` | 配置、secret、feature flag、runtime config 边界 | 保留 |
| Foundation | A5 | `ci-pipeline-baseline` | 当前本地质量门禁，远端 CI deferred | 保留 |
| Foundation | - | `backend-runtime-topology` | P0 frontend/backend 进程拓扑、worker 收敛与开发期观测依赖边界 | 保留 |
| Contract | B1 | `shared-conventions-codified` | Go/TS 共享枚举、错误码、ID、codegen / drift gate | 保留 |
| Contract | B2 | `openapi-v1-contract` | 当前 60 endpoint / 13 tag OpenAPI + fixtures | 保留 |
| Contract | B3 | `event-and-outbox-contract` | 当前 16 internal event、jobType、outbox 契约 | 保留 |
| Contract | B4 | `db-migrations-baseline` | 当前 28 应用表 + auth / migration 支撑表 | 保留 |
| Quality | F1 | `observability-stack` | metrics/log/trace/dashboard/alerting 命名和红线 | 保留 |
| Quality | F3 | `prompt-rubric-registry` | 10 个当前 baseline feature_key、prompt/rubric/model profile 治理 | 保留 |

这些 spec 是当前 engineering roadmap 的基础层。若其中某个计划已完成，后续改动应在该 subject 原地修订，而不是从 roadmap 再 spawn 同主题 plan。P0 implementation subject 在进入设计或实现时创建，并在 §5.2 的当前状态列跟踪。

### 5.2 当前 P0 实施 workstream 候选

以下 subject 在进入设计或实现时创建；未创建的 subject 不进入 `docs/spec/INDEX.md` pending 占位，已创建的 subject 必须作为当前 owner 原地修订。

| Workstream | 建议 subject | 当前状态 | 当前产品 / UI 范围 | 主要依赖 |
|------------|--------------|----------|-------------------|----------|
| App shell + auth + settings | `frontend-shell`、`backend-auth`、`backend-profile` | `frontend-shell` active；`backend-auth` active（001 auth bootstrap completed）；`backend-profile` active（001 candidate profile + experience cards baseline completed） | TopBar、用户菜单、单入口邮箱登录、验证、首次资料补全、退出、pendingAction、用户画像入口、设置与隐私 | A4、B1、B2、B4、ADR-Q1 |
| Home / Job Picks / Parse | `frontend-home-job-picks-and-parse`、`backend-targetjob`、`backend-jobs-recommendations` | `frontend-home-job-picks-and-parse` completed；`backend-targetjob` active（001 import / parse completed）；`backend-jobs-recommendations` active（001 real backend baseline completed: 12 JobMatch endpoints, drainer, fixture parity, E2E.P0.094-097 Ready/automated） | 首页 JD 导入、岗位推荐、解析确认、目标岗位 / JD / 轮次假设 | B2、B3、B4、A3、F3、D1 |
| Mock Interview + Practice | `frontend-workspace-and-practice`、`backend-practice`、`practice-voice-mvp` | `frontend-workspace-and-practice` active；`backend-practice` active；`practice-voice-mvp` active（001 voice MVP completed） | 当前面试规划、简历绑定、公司轻情报卡片、完整文本 / 语音 session、带提示 / 严格模拟 | B2、B3、B4、A3、C4 |
| Report Dashboard | `frontend-report-dashboard`、`backend-review` | `frontend-report-dashboard` active；`backend-review` active | 报告生成、上下文条、准备度、维度、题目回顾、复练当前轮 / 进入下一轮 | B2、B3、B4、A3、C5、F3 |
| Resume Workshop | `frontend-resume-workshop`、`backend-resume`、`backend-upload` | `frontend-resume-workshop` active（001 / 002 / 003 completed）；`backend-resume` active；`backend-upload` active（001 file_objects + presign baseline completed） | 原始简历树、结构化主版本、岗位定制版本、创建/解析/确认、版本详情 | B2、B3、B4、A3、C2 `backend-upload` |
| Debrief | `frontend-debrief`、`backend-debrief` | `frontend-debrief` active；`backend-debrief` active | 真实面试上下文选择、文本 / 语音共享记录、复盘分析、复盘面试 | B2、B3、B4、A3、C4、C6 |
| Backend async runner | `backend-async-runner` | active（001 internal job + outbox runner baseline，承接单一 backend in-process runtime kernel、outbox dispatcher、retry/reaper/shutdown、`email_dispatch` 收口） | backend 内部 job/outbox runner、retry、删除链路执行；不创建独立 worker 进程 | B3、B4、A2、ADR-Q2/Q5、backend-runtime-topology |
| Mock + E2E + release | `mock-contract-suite`、`e2e-scenarios-p0`、`analytics-funnel`、`release-gate-and-rollout` | `mock-contract-suite` active；`test/scenarios/e2e` framework 已创建，默认本地 runner；其余未创建 | fixture-backed mock、P0 主漏斗 BDD、产品漏斗、后续 release / rollback / SLO gate | B2、D1-D6、C4-C9、F1-F3 |

### 5.3 Future candidates（不自动 spawn）

| 能力方向 | 当前处理 | 创建条件 |
|----------|----------|----------|
| 嵌入式 readiness / trends | 可作为报告、画像、面试规划的增强，不是独立成长中心 | P1 设计明确数据来源、展示位置和验收指标 |
| personal knowledge retrieval | 当前不保留实现或基础设施；业务接入后置 | 有明确大量资料场景、隐私边界和质量评估后重新设计 |
| privacy export / advanced audit | P0 删除-only，导出延后 | product-scope 或隐私合规要求升格 |
| company/source intel | 当前 UI 只有轻量公司情报详情 | 数据源、freshness、合规和维护成本已设计 |
| production voice | 当前 UI 保留语音形式；真实 STT / 媒体留存 / 隐私开关后置 | 延迟、成本、retention 和删除链路可验证 |
| multi-platform job search | product-scope 明确规划例外，不属于当前 MVP | 单独设计数据接入、合规、质量和运维边界 |

## 6 实施顺序

### 6.1 S0 · 已完成基础层

当前已完成或已创建 active spec 的基础层包括 A1-A5、B1-B4、F1、F3，以及 ADR-Q1..Q6。后续实现不得绕过这些文档和编码 truth source。

### 6.2 S1 · Contract-backed mock runway

目标是让当前 UI 五入口和会话级页面能基于 B2 fixtures 跑通 P0 happy path：

1. 创建或修订 `mock-contract-suite`，把 60 operation fixtures 提供给前端和后端 mock。
2. 创建或修订 `frontend-shell`，锁 TopBar、用户菜单、display controls、auth pendingAction、settings/profile 入口。
3. 创建或修订 D2-D6 前端 workstream，严格按 `docs/ui-design/` 和 `ui-design/src/app.jsx` 目标路由实现。
4. 在每个用户可见 workstream 的 plan 中维护 BDD gate。

### 6.3 S2 · Backend domain implementation

目标是把 P0 主流程所需后端域按契约落地：

1. `backend-auth`、`backend-upload`、`backend-profile` 提供身份、文件与画像基础。
2. `backend-async-runner` 提供 backend 内部 job、outbox、retry 和删除链路执行；P0 不拆独立 worker 进程。
3. `backend-targetjob`、`backend-practice`、`backend-review`、`backend-resume`、`backend-debrief` 分别承接当前产品闭环。
4. 所有 AI 输出必须引用 A3 / F3，所有长任务必须走 B3 job/outbox contract 并由 backend internal runner 承接，所有敏感数据必须符合 product-scope 隐私红线。

### 6.4 S3 · True integration and release gate

目标是从 fixtures / mock 切到真实后端，并证明 P0 闭环可上线：

1. D2-D6 各自 integration plan 切真 API。
2. `e2e-scenarios-p0` 覆盖导入 -> 规划 -> 练习 -> 报告 -> 复练当前轮 / 下一轮 -> 真实复盘。
3. `analytics-funnel` 对齐当前产品漏斗，不恢复错题或成长独立漏斗。
4. `release-gate-and-rollout` 验证 SLO、rollback、AI provider 可观测、删除 SLA、audit 和导出延后例外。

### 6.5 S4 · Future staged capabilities

P1/P2 能力只在产品 / UI / 合规设计重新确认后创建 subject。不得提前把 future candidates 写入 `docs/spec/INDEX.md` 或创建空 plan。

## 7 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|----------|
| C-1 | Roadmap 真理源一致 | product-scope、docs/ui-design、ui-design 已收敛 | 读取本 spec | 本 spec 不再规划当前 UI 已丢弃的独立模块 | 001 checklist 2.1 / 2.2 |
| C-2 | INDEX 无僵尸占位 | 某 subject 尚未创建真实 `spec.md` | 查看 `docs/spec/INDEX.md` | 不出现 `_pending_` 行或待 spawn 占位 | 001 checklist 2.4 |
| C-3 | Active spec 保留 | A/B/F 基础契约仍有编码 truth source | 查看 `docs/spec/*/spec.md` | 这些 spec 保持 active，并在后续实现中作为依赖 | 001 checklist 1.2 |
| C-4 | Future 不自动恢复旧能力 | 旧 root spec、旧 route 或旧画板标签提到 Growth / Mistakes / Drill / Plan / Voice | 评审后续需求或 plan | 默认视为已丢弃，必须先修订 product-scope 和 UI 文档才能恢复 | 001 checklist 3 |
| C-5 | P0 workstream 可进入实现 | 某当前 UI 模块准备进入实现 | 创建 child spec / plan | plan 有 context、TDD 策略；用户行为有 BDD gate | 后续 child plan |
| C-6 | 文档一致性 | roadmap 修订完成 | 运行校验 | `validate_context`、`sync-doc-index --check`、Markdown link check、`git diff --check` 通过 | 001 checklist 2.6 |
| C-7 | 技术草稿可删除性 | product-scope §1.5 已持有 owner matrix | 删除已迁移技术草稿前运行 zero-reference gate | 当前 owner spec / coded truth source 可独立说明 API / DB / event / metrics / logging / config 契约；仓库中不再出现已迁移草稿的目录名或文件名 | 001 checklist 4 |

## 8 关联计划

- [001-decompose-subspecs](./plans/001-decompose-subspecs/plan.md)：历史上承接 38 child / wave 分解；当前已原地修订为 roadmap rebaseline 与后续 child 创建治理计划，不再维护 pending 占位模型。
