# Product Scope Spec

> **版本**: 2.8
> **状态**: active
> **更新日期**: 2026-07-07

## 1 背景与目标

### 1.1 当前背景

EasyInterview 当前产品形态由 `docs/spec/product-scope/spec.md`、`docs/ui-design/`、`ui-design/` 和各工程 owner spec 共同约束。产品工作台只围绕 `首页 / 模拟面试 / 简历` 三个一级入口组织，报告只有 `Report Dashboard(sessionId)` 一种形态，并且必须区分 `复练当前轮` 与 `进入下一轮`。

本 spec 是 EasyInterview 当前产品范围、阶段边界、非目标、质量红线和文档真理源关系的正式入口。所有开发、评审、plan 设计和 owner handoff 都以当前 active truth source 为准。

### 1.2 一句话定义

**EasyInterview 是一款围绕具体目标岗位、JD、简历资产和真实面试流程设计的 AI 面试训练产品。**

它把“看到岗位 -> 导入 JD / 简历 -> 完成模拟面试 -> 获得证据化报告 -> 复练当前轮或进入下一轮”的准备过程标准化、可重复化和可追踪化。

### 1.3 当前版本目标

本 spec 的目标不是重新扩张产品愿景，而是把当前已经收敛的产品事实固定下来：

1. **固定产品真理源**：将当前产品范围固定在 `docs/spec/product-scope/spec.md`，纳入 spec-centric 文档治理。
2. **对齐当前 UI 真理源**：把 `ui-design/` 和 `docs/ui-design/` 已确认的导航、路由和模块边界纳入产品层判断。
3. **保留必要产品水准**：继续覆盖目标用户、JTBD、产品原则、阶段路线、质量评估、隐私伦理和工程治理要求，而不是只写 UI 清单。
4. **约束后续实施**：后续 child spec、plan、OpenAPI、DB、前端实现和场景测试不得绕过当前产品边界创造平行流程。

### 1.4 真理源关系

| 层级 | 当前真理源 | 负责内容 |
|------|------------|----------|
| 产品范围 | 本文档 | 用户、JTBD、P0/P1/P2/P3 边界、非目标、质量和伦理红线 |
| UI / 交互 | `docs/ui-design/` + `ui-design/` | 当前静态 UI、一级导航、页面职责、目标路由、视觉和交互目标 |
| 工程拆分 | `docs/spec/engineering-roadmap/spec.md` | child subspec、wave、依赖 DAG、mock-first 集成策略 |
| 技术契约 | Layer A/B/F active spec + 已编码 truth source（`openapi/`、`shared/`、`migrations/`、`config/`） | API、DB、事件、共享枚举、AI provider / model routing、配置、可观测性等工程约束；owner matrix 见 §1.5 |

当文档出现冲突时，产品范围以本文档为准，UI 具体交互以 `docs/ui-design/` 为准，工程落地路径以对应 child spec / plan 为准。若工程 roadmap 与本文档出现阶段漂移，必须先通过 `/plan-review` 或新的设计修订对齐，再进入 `/implement`。

### 1.5 技术契约 owner matrix

本小节是当前项目的工程契约分层索引。后续文档、plan 和实现只能引用下表 owner spec 与编码真理源；如果某个责任没有在对应 owner 中明确字段、模块、schema、事件、指标或 gate，则视为尚未落地，必须先修订 owner spec，再进入 `/implement` 或对应代码计划。

| 契约职责 | 当前 owner spec | 当前可执行真理源 | 统一规则 |
|----------|-----------------|------------------|----------|
| 全局命名、ID、时间、错误码、共享枚举、分页和 API error envelope | [shared-conventions-codified](../shared-conventions-codified/spec.md) | `shared/conventions.yaml`、Go / TS generated shared packages | 新增或修改共享字面量必须先修订 B1 或对应 owner spec，再同步 YAML 和生成物；共享 enum / error code / `PageInfo` / `ApiError` 不得在业务域私造 |
| HTTP API、公共 schema、fixtures、mock contract 和 breaking-change gate | [openapi-v1-contract](../openapi-v1-contract/spec.md) | `openapi/openapi.yaml`、`openapi/fixtures/`、OpenAPI generated packages / baseline | API inventory、tag、auth 形态、header、status code、schema 和 fixture provenance 以 B2 为准；任何破坏性变更必须走 B2 ADR / diff gate |
| DB 表、索引、enum check、迁移、隐私删除 / 导出占位和 audit 表 | [db-migrations-baseline](../db-migrations-baseline/spec.md) | `migrations/`、`migrations/enum-sources.yaml`、migration lint / probe | 表数量、列名、索引、enum source、migration/backfill 策略和 privacy deletion matrix 以 B4 与实际 migration 为准 |
| internal event、jobType、outbox envelope、dispatcher 与 breaking-change baseline | [event-and-outbox-contract](../event-and-outbox-contract/spec.md) | `shared/events.yaml`、`shared/jobs.yaml`、generated schemas / baselines | `eventName`、`jobType`、API-facing subset、outbox 字段、dispatcher retry/dead-letter 语义以 B3 为准 |
| 本地运行、基础架构约束、配置、secret、feature flag、runtime-config | [engineering-roadmap](../engineering-roadmap/spec.md)、[local-dev-stack](../local-dev-stack/spec.md)、[secrets-and-config](../secrets-and-config/spec.md) | `config/`、`config/feature-flags.yaml`、A2/A4 runtime code 与 deploy assets | 运行时、配置字典、secret 归属、feature flag 和公开配置 allowlist 由 A2/A4 决定 |
| AI provider、model profile、fallback、调用观测字段、prompt / rubric 坐标 | [ai-provider-and-model-routing](../ai-provider-and-model-routing/spec.md)、[prompt-rubric-registry](../prompt-rubric-registry/spec.md) | `config/ai-providers.yaml`、`config/ai-profiles.yaml`、F3 plan-defined prompt / rubric runtime assets | 业务代码只依赖 provider capability、model profile、feature key 和 prompt/rubric 坐标；模型、fallback 和 prompt/rubric 版本必须可审计 |
| metrics、logging、trace、dashboard、alerting、敏感字段红线 | [observability-stack](../observability-stack/spec.md) | F1 spec、后续 logger / metric / alert rule 编码 truth source | metric / label / log 字段 / trace attribute / dashboard / alert 和明文红线以 F1 与实现 gate 为准 |
| 产品模块、一级入口、route、UI 画板标签和用户行为流 | 本文档、`docs/ui-design/`、`ui-design/` | UI 文档、静态原型、后续 BDD / E2E 场景 | 产品和 UI 范围先于技术契约；未进入当前产品 / UI 范围的模块不得由技术层私自引入 |

## 2 范围

### 2.1 In Scope

- EasyInterview 当前产品定义、目标用户、JTBD 和产品原则。
- 当前 P0 MVP 闭环：`JD / 简历导入 -> 目标面试规划 -> 模拟面试 -> Report Dashboard -> 复练当前轮 / 进入下一轮`。
- 当前一级 UI 模块：`首页`、`模拟面试`、`简历`。
- 用户菜单能力：`设置与隐私`、`认证流程`。
- 模拟面试会话边界：文本面试、语音面试、语音转文字、带提示练习、严格模拟、面试官角色、问题推进、结束并生成报告。
- 报告边界：session-scoped Dashboard、准备度档位、维度状态、题目回顾、证据详情、复练计划。
- 简历边界：平铺简历列表、上传 / 粘贴创建、Agent/LLM 解析预览确认、LLM 从简历内容生成有意义的 `displayName`、简历详情只读展示简历内容本身。
- P1/P2/P3 的方向性阶段路线：只围绕当前 UI 已保留能力做工程化、质量化和规模化扩展，并单独标注明确规划例外。
- 评分、Prompt/Rubric/Model 版本追踪、隐私、伦理、数据来源和可观测性红线。
- 与 `docs/ui-design/`、`docs/spec/engineering-roadmap/` 和当前工程契约 owner 的职责边界。

### 2.2 Out of Scope

- 不在本 spec 中编写具体 API schema、DB schema、事件 payload、prompt 文本或 UI 组件代码。
- 不做岗位推荐、全球搜岗或任何岗位发现 / 聚合模块；JD 获取唯一入口是用户自带 JD（粘贴 / 上传 / URL 导入）。
- 不做简历版本树、主版本 / 岗位定制版本继承或版本管理系统；简历按平铺资产管理。
- 不在本 spec 中承载 engineering-roadmap child subspec 的具体重排；roadmap 拆分与后续 child 创建规则以 `docs/spec/engineering-roadmap/spec.md` 为准。当前 engineering-roadmap active spec 不采用 pending 占位模型，后续 workstream 限定为当前产品 / UI 已保留能力的 on-demand child。
- 不创建实现 plan、TDD checklist 或 BDD scenario；本文档是 docs-only 产品范围结晶。
- 不做独立成长中心、多轮计划、经历库、追问树、单题 Drill、独立错题队列、报告时间线、刊物式报告页、真实面试复盘或用户画像。
- 不定义 Team / EDU、企业端候选人评估、社区或真实面试中的隐形实时辅助。
- 不承诺没有校准证据的通过率、录用概率、击败候选人比例或强结论评分。

### 2.3 当前阶段边界

当前阶段仍以 P0 MVP 闭环为中心。P0 的判断标准不是“功能最多”，而是：

> 用户带着一份具体 JD 和简历来到产品后，能否在最短时间内完成一轮有效练习，并明确知道下一次该复练当前轮还是进入下一轮。

凡不能增强这个闭环的能力，不得进入当前主流程。

### 2.4 当前范围规则

当前 UI 静态设计和 `docs/ui-design/` 是产品范围的正向清单：

- 已在 UI / UI 文档中作为目标模块、会话页面、横切控制或重定义能力出现的内容，才进入当前产品范围。
- 未在 UI / UI 文档中出现，也未在本 spec 中明确列为规划例外的内容，默认视为非当前范围。
- 引入任何非当前范围能力，必须先原地修订本 spec 和对应 UI 文档，再进入工程 plan；不得只靠 route、画板标签或组件名称扩展产品范围。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 产品真理源 | `docs/spec/product-scope/spec.md` 是当前产品范围入口 | 后续产品评审、plan 设计、scope 判断只能引用当前 active truth source |
| D-2 | UI 真理源 | `docs/ui-design/` + `ui-design/` 是当前 UI / 交互真理源 | 产品 spec 不复述每个控件细节；交互冲突以 UI 文档为准 |
| D-3 | 默认入口 | App 默认进入首页，用户可直接粘贴 / 上传 / URL 导入 JD | 不设置未登录欢迎页或登录前置页 |
| D-4 | 一级导航 | `首页 / 模拟面试 / 简历` | `复盘`、`岗位推荐`、`当前岗位`、`面试报告`、`成长` 等不作为一级导航 |
| D-5 | 模拟面试语义 | 面试是一场完整 session，不提供热身、反问、单题深钻等入口前模式卡片 | 复练和下一轮都从报告 CTA 直接进入对应 session |
| D-6 | 语音边界 | 语音是面试形式，不是独立业务模块；当前目标路由只允许 `practice` 显式携带 `mode=voice` / `modality=voice` | 文本输入框麦克风只表示语音转文字；不设置独立语音 route |
| D-7 | 报告形态 | 报告只有 session-scoped Dashboard 一种形态 | 报告不作为一级导航，不提供时间线、刊物式报告或 `reportLayout` 多形态 |
| D-8 | 错题价值承载 | 错题本的用户价值放在报告题目回顾、证据缺口和本轮复练中 | 不设置独立错题队列、Drill builder 或单题 retry |
| D-9 | 真实面试复盘边界 | 复盘不属于当前 P0 模块或后续默认能力 | 不设置 `debrief` route、API、DB、AI feature key 或场景 |
| D-10 | 证据和版本化 | 生成结果必须携带 prompt / rubric / model / language / feature flag / data source 等来源信息 | 支撑质量评估、回归检测和问题追踪 |
| D-11 | 默认范围规则 | 当前 UI / UI 文档未定义且本 spec 未标为规划例外的内容，默认不进入当前范围 | P1/P2/P3 不自动扩展非当前能力 |
| D-12 | Roadmap 对齐 | engineering-roadmap active spec 不采用 `docs/spec/INDEX.md` pending child 占位模型，只保留真实 active spec，并要求后续 child 按当前 Home / Mock Interview / Practice / Report / Resume / Settings/Auth 等保留能力 on-demand 创建 | 后续 child spec / plan 不得用 pending 名称或 route token 引入非当前范围模块 |
| D-13 | 技术契约 owner matrix | 本 spec §1.5 定义当前工程契约职责、owner spec、编码真理源和统一规则 | 后续文档必须引用当前 owner spec 和编码 truth source；不得引入未被 owner 承接的实施依据 |
| D-14 | JD 导入单次确认 | `parse` 解析确认页同时承载面试启动决策：核对解析结果、绑定简历、确认 InterviewRound、立即面试；首次导入链路只允许一次全页确认 | `workspace` 不再是首次导入的必经第二确认页，定位为回访枢纽（最近面试、切换/新建规划、公司轻情报、会话记录、再次发起面试）；不得在 parse 与 session 之间设置平行确认页 |
| D-15 | 复盘上下文选一带二 | 复盘上下文选择不属于当前产品能力 | 不保留相关 UI / API / 场景 |
| D-16 | 无密码认证唯一流 | 邮箱验证码是唯一登录方式；不提供独立“重置登录 / 密码重置”页面 | 验证码重发与更换邮箱在 `auth_verify` 内完成；`auth_reset` route 输入归一回 `auth_login`，文档与 UI 不出现“忘记密码 / 密码 / 两步验证”口径 |
| D-17 | 岗位推荐模块边界 | `岗位推荐 / Job Picks / jd_match` 不属于当前一级模块；JD 获取唯一入口是首页导入 | 一级导航为三项；主流程不包含岗位推荐来源；Q-2 关闭；`全球多平台搜岗` 不属于当前范围 |
| D-18 | 公司情报仅保留轻量嵌入卡片 | `company_intel` 独立详情页与 route 不属于当前范围；模拟面试规划页内嵌轻量卡片是公司情报唯一呈现 | 轻量卡片仍展示一句话画像、近期公开信号、反问建议与合规来源说明；P2 生产化仍以嵌入卡片为呈现边界 |
| D-19 | 报告 CTA 单点收敛 | 报告页只保留 Header 一对 CTA：`复练当前轮` 与 `进入下一轮`；复练计划详情只承载路径说明与复练清单，不重复 CTA 按钮 | 题目回顾的`加入本轮复练`是复练计划标记动作，不直接开启 session；报告任何二级详情不得再出现重复的开练按钮 |
| D-20 | 简历资产扁平化与详情只读 | 简历是平铺列表中的独立资产，不区分原始简历 / 结构化主版本 / 岗位定制版本，不做版本树、分叉、继承或版本管理；创建仅上传 / 粘贴；LLM 解析结果必须产出有意义的 `displayName`，不得长期停留在“上传的简历 / 粘贴的简历”等通用标题；详情页只读展示简历内容本身，不提供导出 PDF、复制、编辑、预览 tab、改写建议或查看原始简历预览 | 每份简历仍保留原始来源与解析文本快照作为后台证据；面试绑定对象为简历（`resumeId`）；原始简历预览就是详情中的简历正文 |
| D-21 | 全局呈现收敛 | 设置页只保留个人资料与隐私数据两个 tab；主题保留四个预设 + `自定义 accent`（色相 / 饱和度滑杆）+ 暗色模式 + 语言下拉 + 字体预设，默认主题为`深海` | 通知、订阅等占位能力未来按需重新设计，不以空 tab 形式预留；主题自定义按用户决策保留 |
| D-22 | 核心闭环方案 B | 真实面试复盘和用户画像不属于当前功能范围：`debrief` / `profile` 不再是目标 route，`Debriefs` / `Profile` OpenAPI tags、`debriefs` / `candidate_profiles` / `experience_cards` DB 表、debrief/profile AI feature key、shared event/job 和正向场景不保留 | P0 只保留 JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮；账号资料补全和设置隐私保留，但不承担候选人画像语义 |

### 3.2 待确认事项

| ID | 待确认事项 | 影响 | 默认处理 |
|----|------------|------|----------|
| Q-1 | 生产版语音面试能力的 P0 发布门槛 | UI 已有语音面试目标，但后端 STT、媒体留存和隐私链路可能晚于文本闭环 | 产品层保留语音面试形式；工程 release gate 可用 feature flag 控制上线 |
| Q-3 | 商业包装与价格 | 商业包装尚未定稿 | 当前不把商业包装写成产品功能范围；Team / EDU 不进入产品规划 |

> Q-2（岗位推荐数据来源策略）已关闭，不再是待确认事项；当前 JD 获取唯一入口是首页导入。

## 4 设计约束

### 4.1 产品原则

1. **极速进入价值**：首页必须允许用户直接输入 JD，不以长 onboarding 或登录作为开始前置。
2. **目标岗位优先**：练习、报告和简历建议都必须围绕 `TargetJob / JD / Resume / Round / Session` 建立上下文。
3. **完整面试优先**：产品默认组织一场完整模拟面试，而不是把用户分散到热身、单题、追问树和错题队列中。
4. **证据优先于伪精确**：报告前台展示准备度档位、维度状态、证据片段和下一步动作，不输出未经校准的精确通过率。
5. **复练优先于一次性报告**：报告必须提供 `复练当前轮` 和 `进入下一轮` 两条清晰路径，且二者不能混用。
6. **隐私默认保守**：简历、JD、面试回答、音频和报告都属于敏感数据，采集、留存、导出、删除必须可解释。
8. **工程与评估前置**：模型调用、prompt/rubric、feature flag、数据来源和语言必须可追踪；质量回归必须可复现。

### 4.2 UI 约束

- 顶部导航只能出现当前 UI 真理源确认的三个一级入口。
- `workspace` 的产品语义是 `模拟面试 / 当前面试规划`，不是一级岗位资产管理模块。
- `report` 必须带 `sessionId` 或等价会话上下文；无上下文时不得展示假报告。
- `debrief` / `debrief_full` / `profile` 不是目标 route；对应路径或 hash 输入必须归一到当前核心入口。
- 语音面试只能通过 `practice?mode=voice&modality=voice` 或等价显式参数进入；不得保留或新增 `voice` 目标 route / route alias。
- route、画板标签和组件名称不得单独作为新增产品能力的依据。

### 4.3 评分与反馈约束

报告可以展示：

- 准备度档位：例如 `未就绪 / 建议再练 / 基本可面 / 较为充分`。
- 维度状态：例如 `强项 / 达标 / 待加强`。
- 置信度或证据充分性说明。
- 具体题目、回答片段、证据缺口、推荐框架和下一步复练建议。

报告不得展示：

- 未经校准的精确录用概率。
- “击败多少候选人”式营销指标。
- 没有证据片段支撑的强烈通过 / 淘汰判断。
- 仅由语速、停顿、口头禅等表达信号推出的能力结论。

### 4.4 伦理与安全约束

- 不做真实面试中的隐形实时辅助或作弊能力。
- 不把音视频或情绪识别结果用于企业端候选人淘汰决策。
- 不把用户面试准备材料默认公开或用于他人训练样本。
- 不在日志中输出原始简历、原始 JD、完整回答、音频转写全文或敏感个人信息。
- 任何 AI 生成的岗位分析、报告和简历建议都必须保留来源、版本、语言和模型调用元数据。

### 4.5 文档治理约束

- 本 spec 后续修订必须原地更新，不创建同主题 sibling spec。
- 任何改变当前 P0 用户行为流的修订，必须同步检查 `docs/ui-design/`、`docs/spec/engineering-roadmap/` 和相关 child spec。
- 涉及代码实现的计划必须通过 `/implement` -> `/tdd`，涉及用户行为的计划必须维护 BDD gate。
- 非当前范围能力必须先在本 spec 中重新设计和批准。
- route、画板标签或组件名称不能单独作为新增产品能力的依据。

## 5 模块边界

### 5.1 产品能力层

EasyInterview 的完整能力仍围绕五个产品层组织，但 UI 一级入口和底层能力层不是一一对应关系。

| 能力层 | 当前职责 | 当前 UI 承载 |
|--------|----------|--------------|
| M1 · Context Inputs | 通过简历、JD 和模拟面试上下文支撑面试定制与报告解释 | 首页、简历、模拟面试和报告中的上下文，不提供用户画像页 |
| M2 · Target Job Workspace | 围绕 JD 管理岗位要点、简历绑定、轮次假设、公司轻情报和会话记录 | 首页 JD 导入、模拟面试规划 |
| M3 · Mock Interview Orchestrator | 组织完整模拟面试、问题推进、追问、提示和 session 状态 | Interview Session |
| M4 · Evidence-based Review | 产出证据化报告、题目回顾和复练计划 | Report Dashboard |
| M5 · Growth Signals | 准备度变化、报告后动作和练习趋势等横切信号 | 不作为独立成长中心；仅能嵌入报告、模拟面试规划等现有模块 |

### 5.2 当前一级 UI 模块

| 一级模块 | 用户任务 | P0 职责 | 不承担的职责 |
|----------|----------|---------|--------------|
| 首页 | 快速开始一次岗位准备 | JD 粘贴 / 上传 / URL 导入，最近模拟面试，创建简历入口 | 不做登录前营销页；不做复盘辅助入口 |
| 模拟面试 | 回访并管理既有面试规划，再次发起 session | 当前面试规划（回访枢纽）、切换/新建规划、公司轻情报嵌入卡片、会话记录、立即面试 | 不作为泛岗位资产管理中心；不作为首次 JD 导入的必经第二确认页 |
| 简历 | 管理可被岗位和面试消费的简历资产 | 平铺简历列表、上传 / 粘贴创建、解析预览确认、LLM 生成可识别简历名称、只读简历详情 | 不做版本树 / 主版本 / 岗位定制继承；不做复杂排版设计器；详情页不做导出、复制、编辑、改写建议或原件弹层 |

### 5.3 会话级和上下文页面

| 页面 | 归属 | 进入方式 | 关键上下文 |
|------|------|----------|------------|
| `parse` | JD 解析确认 + 面试启动 | 首页 JD 导入 | `jdId / targetJobId / resumeId / roundId` |
| `practice` | Interview Session | 模拟面试规划、报告复练、进入下一轮 | `sessionId / targetJobId / resumeId / roundId / mode / modality / practiceMode` |
| `generating` | 报告生成过渡态 | 面试结束 | `sessionId / practiceMode / hintUsed / hintCount` |
| `report` | Report Dashboard | 面试结束、会话记录、相关入口 | `sessionId` |
| `settings` | 设置与隐私 | 用户菜单 | `userId` |

### 5.4 当前范围外能力

| 能力 | 当前边界 | 说明 |
|------|----------|------|
| 未登录欢迎页 | 范围外 | App 默认进入首页，降低首次价值路径阻力 |
| 当前岗位一级导航 | 范围外 | 岗位信息归属于模拟面试规划 |
| 面试报告一级导航 | 范围外 | 报告隶属于 session |
| 练习模式卡片 | 范围外 | 面试是一场完整过程，辅助程度在 session 内切换 |
| 热身 / 单题深钻 / 反问专练 | 范围外 | 反问内容可作为完整面试中的问题或公司情报建议出现 |
| 追问树 | 范围外 | 追问发生在面试会话内 |
| 独立错题队列 | 范围外 | 错题价值放入报告题目回顾和本轮复练 |
| 成长中心 | 范围外 | 准备度和趋势信号只嵌入报告或面试规划 |
| 多轮计划 | 范围外 | 轮次节点在模拟面试规划和报告 CTA 中表达 |
| 经历库 / STAR 编辑器 | 范围外 | 经历证据由简历和面试上下文承载 |
| 报告时间线 / 刊物式报告 | 范围外 | 报告统一为 Dashboard |
| 岗位推荐一级模块 | 范围外 | JD 获取唯一入口是首页导入 |
| 公司情报独立详情页 | 范围外 | 轻量情报由模拟面试规划页嵌入卡片承载 |
| 简历版本树 / 主版本 / 岗位定制版本 | 范围外 | 简历按平铺资产管理，不做版本继承 |
| 轻量问答建档 | 范围外 | 创建简历只保留上传 / 粘贴 |
| 设置页通知 / 订阅占位 tab | 范围外 | 设置页只保留个人资料与隐私数据 |
| 真实面试复盘 / Debrief | 范围外 | 当前闭环聚焦模拟面试报告后的复练 / 下一轮 |
| 用户画像 / CandidateProfile / ExperienceCard | 范围外 | 账号资料和设置隐私保留，但不沉淀独立候选人画像产品或数据模型 |

## 6 P0 MVP 详细规格

### 6.1 MVP 目标

P0 只回答一个问题：

> 当用户带着一份具体 JD 和简历来到 EasyInterview，产品能否在最短路径内帮他完成一次针对性练习，拿到可追溯证据的报告，并知道下一步应该复练当前轮还是进入下一轮。

P0 成功标准：

- 用户不登录也能看到 JD 导入和准备路径。
- 用户可以围绕一个目标岗位、绑定简历和面试轮次进入完整模拟面试。
- 面试结束后生成一份有会话上下文、有证据、有下一步动作的报告。
- 报告可以直接发起当前轮复练或下一轮 session。

### 6.2 主流程 A：带着 JD 来的用户

```text
Home
  -> 粘贴 / 上传 / URL 导入 JD
  -> Parse & Confirm（解析确认 + 面试启动）
     ├─ 核对 JD 基础信息 / 必需项 / 加分项 / 隐性关注点
     ├─ 绑定简历
     ├─ 确认 InterviewRound
     ├─ 立即面试
     └─ 仅保存规划 -> Mock Interview Plan（回访枢纽）
  -> Interview Session
  -> 结束并生成报告
  -> Report Dashboard(sessionId)
     ├─ 复练当前轮
     └─ 进入下一轮
```

首次导入链路只有一次全页确认：解析确认页同时完成核对与启动决策。`Mock Interview Plan` 服务回访场景：从首页最近模拟面试、报告或会话记录回到既有规划，切换/新建规划，并再次发起面试。

### 6.3 主流程 B：先补简历的用户

```text
Home 或 Resume
  -> 1 分钟创建简历 / 新建简历
  -> 上传 / 粘贴
  -> Agent 解析
  -> 预览确认保存
  -> 在解析确认页或模拟面试规划中绑定这份简历
```

简历模块的价值不是排版，而是形成可被 JD 解析、模拟面试和报告调用的证据资产。

### 6.4 当前流程边界

用户完成模拟面试后的下一步只保留报告内的 `复练当前轮` 与 `进入下一轮`。真实面试复盘不属于当前 P0 主流程、一级入口、后续默认 workstream 或场景验收对象。

### 6.5 M1：上下文输入

#### 目标

上下文来自简历、JD、面试 session 和报告，只作为当前任务输入和证据来源。候选人画像页、画像纠偏、Experience Card 和独立画像数据模型不属于当前功能范围。

#### 输入来源

- 简历资产（原始来源、解析文本快照和结构化内容）。
- 用户导入的 JD、岗位偏好、地区、语言和目标职级。
- 模拟面试中的回答、提示使用、追问风险和报告结果。

#### 输出

- 可用于面试和简历建议的经历证据。
- 岗位偏好、语言偏好和界面偏好边界。
- 待补充信息提示。

#### 当前不进入本模块

- 冗长职业咨询 onboarding。
- 性格测试式大段画像。
- 用户画像页、画像纠偏和 Experience Card。
- 独立经历库或 STAR 编辑器。
- 无来源的能力结论。

#### 验收标准

- 没有完整简历也能开始 JD 导入和模拟面试规划。
- 面试和报告使用的上下文必须来自当前 JD、简历或 session 证据。
- 不存在 `profile` 目标 route、`CandidateProfile` / `ExperienceCard` API 或画像页入口。

### 6.6 M2：目标岗位工作台

#### 目标

把一份 JD 变成可练、可关联简历和会话记录的面试规划。

#### 用户入口

- 首页粘贴 JD。
- 首页上传 JD 文件。
- 首页粘贴岗位 URL。
- 最近模拟面试回到当前规划。

#### 系统处理

- 解析岗位标题、公司、地区、语言、职级和来源。
- 提炼必需项、加分项、隐性关注点和风险提示。
- 关联简历和当前面试轮次。
- 展示公司轻情报嵌入卡片和面试轮次节点。
- 展示当前规划下的模拟面试记录。

#### 当前不进入本模块

- 岗位推荐 / 全球多平台搜岗：不属于当前功能范围，不再是规划例外。
- 复杂职位收藏生态。
- 大规模外部公司情报拼图与独立情报详情页。
- 把 `当前岗位` 做成独立一级模块。

#### 验收标准

- 任意一份可读 JD 能进入解析确认流程，并在同一页完成简历绑定、轮次确认和立即面试。
- 首次导入链路在 parse 与 session 之间不得出现第二个全页确认。
- 模拟面试规划（回访枢纽）必须明确 `TargetJob / JD / 简历 / InterviewRound`。
- 当前规划记录不得混入其他公司、岗位或 JD 的会话。

### 6.7 M3：模拟面试编排器

#### 目标

围绕目标岗位组织一场完整模拟面试，而不是随机抛题或让用户先选择复杂练习模式。

#### 当前面试形式

- 文本面试：用户主要通过文本回答。
- 语音面试：用户与 AI 进行实时语音对话。
- 语音转文字：文本输入框中的麦克风只负责插入转写，不改变面试形式。

#### 当前辅助程度

- `带提示练习`：展示提示、实时观察、可调用经历和现场提示。
- `严格模拟`：隐藏提示、实时观察、可调用经历和语音现场提示。

#### 关键逻辑

- 面试会话必须有 `sessionId` 和 InterviewContext。
- 面试官角色可以在同一 session 内表达 HR、综合面试官、用人经理等语气差异。
- 追问必须和用户刚才回答、目标 JD 或简历证据直接相关。
- `结束并生成报告` 固定在右侧底部，并传递 `practiceMode / hintUsed / hintCount`。

#### 当前不进入本模块

- 入口前热身、反问专练、单题深钻、追问树或 Drill builder。
- 把严格模拟做成新的面试类型。
- 把语音做成独立页面或独立产品模块。
- 真实面试中的隐形实时辅助。

#### 验收标准

- 用户从模拟面试规划点击 `立即面试` 后直接进入完整 session。
- 用户能明确当前是文本面试还是语音面试。
- 严格模拟时提示和辅助信息必须隐藏。
- 报告生成必须带上练习方式和提示使用记录。

### 6.8 M4：证据化报告

#### 目标

给出足够具体、可操作、可复练的反馈，让用户知道下一次该怎么练。

#### 报告结构

- Header：目标岗位、轮次、会话、绑定简历和报告归属说明。
- Context Strip：`sessionId`、目标岗位、轮次、简历、沟通形式、练习方式、提示记录。
- Summary Cards：准备度、维度详情、题目回顾、下一动作。
- Detail Surface：准备度详情、维度详情、题目回顾页、证据详情、复练计划。
- Next Actions：Header 的 `复练当前轮` 和 `进入下一轮`，是报告唯一一对开练 CTA（D-19）；复练计划详情只承载路径说明与复练清单。

#### 题目回顾

题目回顾承载题目级反馈、证据缺口和本轮复练标记，不形成独立模块。每题至少需要说明：

- 原始问题。
- 用户回答摘要或证据片段。
- 有效点。
- 缺口。
- 推荐框架。
- 可能追问。
- 是否加入本轮复练。

#### 当前不进入本模块

- 报告一级导航。
- 无 session 上下文的报告页。
- 时间线报告。
- 刊物式报告页。
- 独立错题队列或单题 retry。
- 精确通过率或录用概率。

#### 验收标准

- 报告的每个主要判断能回溯到题目、回答、JD、简历或 session 来源。
- 用户看完报告后能直接选择复练当前轮或进入下一轮。
- 无 `sessionId` 时必须显示缺失状态并回到当前面试规划或记录列表。

### 6.9 M5：简历资产

#### 目标

让简历成为可被岗位、面试和报告复用的证据资产，而不是孤立的文档上传入口。

#### 当前范围

- 平铺简历列表：每份简历是一份独立资产，不区分原始简历 / 主版本 / 岗位定制版本。
- 上传 / 粘贴创建，Agent/LLM 解析，预览确认后保存。
- LLM 根据简历内容生成有意义的简历名称，列表和详情不以“上传的简历 / 粘贴的简历”等通用来源标题作为最终名称。
- 每份简历只读保留原始来源（文件或粘贴文本）和解析文本快照。
- 简历详情只读展示简历内容本身；不提供预览 tab、改写建议、手动编辑、原件预览、导出或复制。

#### 当前不进入本模块

- 原始简历树、结构化主版本、岗位定制版本与分叉流程。
- 简历版本管理、版本继承、复制为新版本等版本系统。
- 轻量问答建档。
- 复杂排版编辑器。
- 简历详情二次编辑工作台、改写建议工作台、导出/复制工具栏或原件弹层。
- 多模板视觉设计系统。
- 深度求职信工作台。
- 独立经历库。

#### 验收标准

- 用户可以从无简历状态上传或粘贴创建一份经预览确认的简历。
- 简历列表是单层平铺，不出现树、主版本或定制版本概念。
- 新简历保存后必须有 LLM-derived 可识别名称。
- 用户打开简历详情时看到的就是只读简历正文，旧导出 / 复制 / 编辑 / 改写 / 原件预览入口不存在。
- 模拟面试规划可以绑定列表中的任意一份简历。

### 6.10 报告后行动边界

报告后的行动只保留 `复练当前轮` 与 `进入下一轮`。真实面试复盘、复盘分析、复盘面试和 debrief-derived practice plan 不属于当前功能范围；`debrief` 不作为 route、OpenAPI tag、DB 表、AI feature key、shared event/job 或 E2E 场景。

### 6.11 非当前能力

本节是当前产品范围的负向边界。

| 能力 | 当前判断 | 原因 |
|------|----------|------|
| 岗位推荐一级模块 | 非当前范围 | 超出 MVP 闭环；JD 获取唯一入口是首页导入 |
| 全球多平台搜岗 | 非当前范围 | JD 获取唯一入口是首页导入；如需引入新来源，先修订 §2.4 |
| 公司情报独立详情页 | 非当前范围 | 轻量情报由模拟面试规划页嵌入卡片承载 |
| 简历版本树 / 主版本 / 岗位定制版本 / 轻量问答 | 非当前范围 | 简历按平铺资产管理；创建只保留上传 / 粘贴 |
| 视频情绪识别 | 非当前范围 | 解释风险高，训练价值不稳定，且容易引入不当评估 |
| 社区 | 非当前范围 | 稀释单人岗位准备闭环，不在当前产品定位内 |
| Team / EDU | 非当前范围 | 当前产品只面向个人训练闭环，不规划团队版或组织评估产品 |
| 企业端候选人评估 | 非当前范围 | 与训练产品定位冲突，伦理负担高 |
| 独立成长中心 | 非当前范围 | 当前只保留嵌入报告和面试规划的准备度 / 趋势信号 |
| 独立多轮计划 | 非当前范围 | 面试轮次只在模拟面试规划和报告 CTA 中表达 |
| 独立经历库 / STAR 编辑器 | 非当前范围 | 经历证据由简历和面试上下文承载 |
| 独立错题本 / 单题 Drill / 追问树 | 非当前范围 | 题目问题留在报告题目回顾和本轮复练，不形成单题流程 |
| 报告时间线 / 刊物式报告 | 非当前范围 | 报告统一为 session-scoped Dashboard |
| 真实面试复盘 / Debrief | 非当前范围 | 当前核心闭环收敛到模拟面试报告后的复练 / 下一轮，不维护平行复盘系统 |
| 用户画像 / CandidateProfile / ExperienceCard | 非当前范围 | 账号资料和设置隐私保留，但不再沉淀独立候选人画像产品或数据模型 |
| 隐形实时面试辅助 | 非当前范围 / 长期不做 | 明确违反产品伦理边界 |

## 7 阶段路线

### 7.1 P0：当前 MVP

P0 聚焦一条完整闭环：

```text
JD 导入
  -> 当前面试规划
  -> 完整模拟面试
  -> Dashboard 报告
  -> 复练当前轮 / 进入下一轮
```

P0 退出标准：

- 用户能在无登录前置下进入 JD 导入。
- 至少一条从 JD 到报告再到复练的路径可闭合。
- 报告判断具备证据来源。
- 简历资产能被绑定到模拟面试规划。

### 7.2 P1：加深单人价值

P1 在 P0 闭环内加深当前能力：

- 简历改写建议增强：更稳定的针对岗位建议、与报告缺口联动、操作反馈更完整。
- 多语言能力增强：目标岗位语言和界面语言可以分离。
- 准备度和趋势信号只嵌入报告和岗位规划。

P1 退出标准：

- 用户愿意围绕同一岗位进行 2 次以上复练。
- 简历改写建议能反向补齐报告中暴露的证据缺口。
- 多语言输出通过人工抽检和用户反馈。

### 7.3 P2：工程化和数据源扩展

P2 只扩展当前 UI 已保留能力：

- 生产级语音能力：STT、媒体留存、表达层反馈、隐私开关和删除链路。
- 轻量岗位情报：官方信息、公开新闻、职位上下文、来源和抓取时间；呈现边界仍是模拟面试规划页嵌入卡片（D-18）。

P2 退出标准：

- 语音练习成本、延迟、隐私留存和删除链路可控。
- 公开情报能力不会成为维护负担或合规风险。

### 7.4 P3：可规模化的训练产品

P3 把训练闭环做得更稳定、更可评估、更可扩展：

- 更成熟的 JD 解析和数据来源治理。
- 更稳定的 AI 质量评估、离线集、prompt/rubric 回归和灰度策略。
- 更完整的隐私、导出、删除、审计和用户数据控制。
- 更可靠的跨语言训练、语音训练和公司轻情报。

P3 仍然不做：

- 社区、Team / EDU 或企业端训练营产品。
- 企业端候选人淘汰评估。
- 视频情绪识别。
- 岗位推荐、全球搜岗或岗位聚合平台。
- 独立成长中心、独立多轮计划、独立经历库、独立错题本或单题 Drill。
- 真实面试作弊辅助。
- 无来源公司情报。
- 没有校准的录用概率承诺。

## 8 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 产品真理源明确 | 当前 UI 与工程 owner 需要统一产品入口 | 读本文档和 `docs/spec/INDEX.md` | 本文档明确 active，并作为产品范围入口被索引 | docs-only |
| C-2 | 首页直达价值 | 用户未登录且有 JD | 打开 App 首页 | 可以直接粘贴 / 上传 / URL 导入 JD | 后续 frontend child |
| C-3 | 一级导航一致 | 当前 UI 文档已锁定三个一级入口 | 查看 TopBar | 只出现首页、模拟面试、简历 | `docs/ui-design` |
| C-4 | 模拟面试上下文 | 用户在 JD 解析确认页完成核对并绑定简历和轮次 | 点击立即面试 | session 带入目标岗位、JD、简历版本和轮次，中途不出现第二个全页确认 | 后续 frontend/backend child |
| C-5 | 辅助程度清晰 | 用户在 Interview Session 中切换严格模拟 | 开启严格模拟 | 提示、实时观察、可调用经历和语音现场提示隐藏 | `docs/ui-design/module-practice-review.md` |
| C-6 | 报告归属清晰 | 一场面试完成 | 进入报告页 | 报告展示 session、岗位、轮次、简历、形式和提示记录 | `docs/ui-design/report-dashboard.md` |
| C-7 | 下一步动作清晰 | 用户在报告页查看复练计划 | 点击 CTA | `复练当前轮` 和 `进入下一轮` 分别直接进入对应 session | `docs/ui-design/report-dashboard.md` |
| C-8 | 错题价值不独立成模块 | 用户查看某题表现 | 打开题目回顾 | 看到证据、缺口、建议和加入本轮复练，不进入独立错题队列 | `docs/ui-design/module-practice-review.md` |
| C-9 | 简历资产可绑定 | 用户已有简历 | 打开解析确认页或模拟面试规划 | 可以选择绑定列表中的任意一份简历 | `docs/ui-design/resume-module.md` |
| C-12 | 证据化和版本化 | 任一 AI 生成结果产生 | 保存或展示结果 | 结果可追踪 prompt / rubric / model / language / feature flag / data source | 后续 backend / quality child |
| C-13 | 默认范围规则 | 某能力没有进入当前 product-scope 和 UI truth source | 评审后续需求或 plan | 该能力默认不进入当前范围，除非先修订本 spec 和对应 UI 文档 | docs-only |
| C-15 | 无密码认证唯一流 | 产品只有邮箱验证码登录 | 查看认证页面流 | 不存在独立重置登录页；验证码重发与更换邮箱在 `auth_verify` 内完成 | `docs/ui-design/auth-and-entry.md` |
| C-16 | 岗位推荐与情报独立页零入口 | 岗位推荐模块与公司情报独立页不属于当前范围 | 走查导航、首页与静态原型路由 | 不存在 `jd_match` / `company_intel` 目标 route、岗位推荐入口或独立情报详情页；公司情报只出现在模拟面试规划页嵌入卡片 | `docs/ui-design/module-map.md` |
| C-17 | 简历平铺与采纳收口 | 用户在简历模块管理资产并接受改写 | 查看简历列表并采纳改写建议 | 列表是单层平铺、无树 / 主版本 / 定制版本概念；改写建议仅有`采纳`，采纳后确认前预览可选覆盖原简历或保存为新简历 | `docs/ui-design/resume-module.md` |
| C-18 | 报告 CTA 单点 | 报告已生成 | 走查报告页全部区块 | 只有 Header 一对 `复练当前轮 / 进入下一轮` CTA；复练计划详情与题目回顾不再出现重复开练按钮 | `docs/ui-design/report-dashboard.md` |
| C-19 | 复盘和用户画像零入口 | 当前 P0 只保留 JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮 | 走查 TopBar、用户菜单、URL/hash route、OpenAPI、DB、shared、config 和场景索引 | 不存在 `debrief` / `profile` 目标 route、`Debriefs` / `Profile` OpenAPI tag、`debriefs` / `candidate_profiles` / `experience_cards` 表或正向场景 | [001-core-loop-module-pruning](./plans/001-core-loop-module-pruning/plan.md) |

## 9 质量、安全与评估

### 9.1 质量评估

P0 质量评估必须覆盖三类指标：

| 类型 | 指标 |
|------|------|
| 体验类 | JD 导入完成率、进入模拟面试率、报告查看率、报告后复练率、进入下一轮率 |
| 内容类 | JD 解析准确性、问题与岗位相关性、追问相关性、报告证据可追溯率、简历建议可用性 |
| 工程类 | 报告生成延迟、AI 调用失败率、异步任务失败率、OpenAPI/fixtures drift、prompt/rubric 回归 |

### 9.2 Prompt / Rubric / Model 治理

所有 AI 输出必须至少能追踪：

- `feature_key`
- `prompt_id`
- `prompt_version`
- `rubric_id`
- `rubric_version`
- `model_profile`
- `model_name`
- `language`
- `feature_flag`
- `data_source`
- `input_snapshot_ref`
- `generated_at`

业务域不得 hardcode prompt 文本。Prompt、Rubric 和模型 profile 的管理归对应 quality / AI child spec。

### 9.3 数据与隐私

敏感对象包括：

- 原始简历和结构化简历。
- 原始 JD、岗位链接和来源记录。
- 面试问题、回答、追问、报告和题目回顾。
- 音频、媒体对象和 STT 中间结果。

隐私要求：

- 用户必须能理解哪些数据被用于生成建议。
- 数据删除链路必须可审计。
- 音频和媒体留存必须可关闭、可解释、可删除。
- 日志和指标不得包含原始敏感内容。

### 9.4 安全边界

- API 和前端必须按用户身份隔离 TargetJob、Resume、InterviewSession 和 Report。
- 任何导出、删除、上传、解析和报告生成都必须记录审计或任务来源。
- 任何外部数据源必须记录 source、retrieved_at、freshness 和使用范围。
- 任何 feature flag 改变用户可见 AI 能力时，必须能在报告或审计信息中追踪。

## 10 关联文档

- UI 真理源：[docs/ui-design/README.md](../../ui-design/README.md)、[docs/ui-design/ui-architecture.md](../../ui-design/ui-architecture.md)、[docs/ui-design/module-map.md](../../ui-design/module-map.md)。
- 工程 roadmap：[docs/spec/engineering-roadmap/spec.md](../engineering-roadmap/spec.md)。
- 当前技术契约 owner matrix：本文档 §1.5。
- OpenAPI 契约：[docs/spec/openapi-v1-contract/spec.md](../openapi-v1-contract/spec.md)。
- Prompt / Rubric：[docs/spec/prompt-rubric-registry/spec.md](../prompt-rubric-registry/spec.md)。

本 spec 当前挂载 [001-core-loop-module-pruning](./plans/001-core-loop-module-pruning/plan.md) 作为核心闭环范围收敛 owner。后续如果要调整 P0/P1 阶段或新增用户可见行为，必须先在本 subject 下创建或原地修订 plan / checklist / context.yaml，再进入 `/implement`。

## 11 修订记录

| 版本 | 日期 | 修订内容 |
|------|------|----------|
| 2.7 | 2026-07-07 | 将会话记录、模拟面试记录和报告返回入口统一为当前产品术语，避免 active spec 把记录能力写成过期口径。 |
| 2.6 | 2026-07-07 | 将 product-scope 正文收敛为当前合同表达；中文范围边界只描述当前行为和非当前范围。 |
| 2.5 | 2026-07-06 | 将 active product-scope 中的范围变更过程说明改为当前范围合同与负向边界表述。 |
| 2.2 | 2026-06-29 | 锁定当前核心闭环：JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮；新增 D-22、C-19 和 001-core-loop-module-pruning owner plan。 |
| 2.1 | 2026-06-12 | 确认主题 `自定义 accent` 模式保留，默认主题改为 `深海`；设置页范围收敛为当前 tab 结构。 |
| 2.0 | 2026-06-12 | 收敛一级导航、公司情报嵌入卡片、报告 CTA、flat Resume 和设置页范围；更新 §5.1-§5.4、§7、C-3/C-9 并新增 C-16/C-17/C-18。 |
| 1.9 | 2026-06-12 | 锁定 JD 导入单次确认、workspace 回访枢纽、复盘上下文范围边界和无密码认证唯一流；同步主流程、模块页面表与验收场景。 |
| 1.8 | 2026-05-05 | 既有基线（见 history.md） |
