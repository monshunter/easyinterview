# Product Scope Spec

> **版本**: 2.24
> **状态**: completed
> **更新日期**: 2026-07-15

## 1 背景与目标

### 1.1 当前背景

EasyInterview 当前产品形态由 `docs/spec/product-scope/spec.md`、`docs/ui-design/`、`frontend/` 和各工程 owner spec 共同约束。产品工作台只围绕 `首页 / 面试 / 简历` 三个一级入口组织；`/parse?targetJobId=...` 只承接首页 import 之后的 queued / processing 命令进度，ready 后立即以 history replace 进入 `/workspace?targetJobId=...` 只读规划详情。规划详情内容区可进入 target-scoped ReportsScreen 索引当前规划轮次报告，但它不属于全局导航或第二种报告内容形态。报告内容只有 `Report Dashboard(reportId)` 一种形态，并且必须区分 `复练当前轮` 与 `进入下一轮`。

本 spec 是 EasyInterview 当前产品范围、阶段边界、非目标、质量红线和文档真理源关系的正式入口。所有开发、评审、plan 设计和 owner handoff 都以当前 active truth source 为准。

### 1.2 一句话定义

**EasyInterview 是一款围绕具体目标岗位、JD、简历资产和真实面试流程设计的 AI 面试训练产品。**

它把“看到岗位 -> 导入 JD / 简历 -> 完成模拟面试 -> 获得证据化报告 -> 复练当前轮或进入下一轮”的准备过程标准化、可重复化和可追踪化。

### 1.3 当前版本目标

本 spec 的目标不是重新扩张产品愿景，而是把当前已经收敛的产品事实固定下来：

1. **固定产品真理源**：将当前产品范围固定在 `docs/spec/product-scope/spec.md`，纳入 spec-centric 文档治理。
2. **对齐当前 UI 设计文档**：把 `frontend/` 和 `docs/ui-design/` 已确认的导航、路由和模块边界纳入产品层判断。
3. **保留必要产品水准**：继续覆盖目标用户、JTBD、产品原则、阶段路线、质量评估、隐私伦理和工程治理要求，而不是只写 UI 清单。
4. **约束后续实施**：后续 child spec、plan、OpenAPI、DB、前端实现和场景测试不得绕过当前产品边界创造平行流程。

### 1.4 真理源关系

| 层级 | 当前真理源 | 负责内容 |
|------|------------|----------|
| 产品范围 | 本文档 | 用户、JTBD、P0/P1/P2/P3 边界、非目标、质量和伦理红线 |
| UI / 交互 | `docs/ui-design/` + `frontend/` | 当前静态 UI、一级导航、页面职责、目标路由、视觉和交互目标 |
| 工程拆分 | `docs/spec/engineering-roadmap/spec.md` | child subspec、wave、依赖 DAG、mock-first 集成策略 |
| 技术契约 | Layer A/B/F active spec + 已编码 truth source（`openapi/`、`shared/`、`migrations/`、`config/`） | API、DB、事件、共享枚举、AI provider / model routing、配置、可观测性等工程约束；owner matrix 见 §1.5 |

当文档出现冲突时，产品范围以本文档为准，UI 具体交互以 `docs/ui-design/` 为准，工程落地路径以对应 child spec / plan 为准。若工程 roadmap 与本文档出现阶段漂移，必须先通过 `/plan-review` 或新的设计修订对齐，再进入 `/implement`。

### 1.5 技术契约 owner matrix

本小节是当前项目的工程契约分层索引。后续文档、plan 和实现只能引用下表 owner spec 与编码真理源；如果某个责任没有在对应 owner 中明确字段、模块、schema、事件、指标或 gate，则视为尚未落地，必须先修订 owner spec，再进入 `/implement` 或对应代码计划。

| 契约职责 | 当前 owner spec | 当前可执行真理源 | 统一规则 |
|----------|-----------------|------------------|----------|
| 全局命名、ID、时间、错误码、共享枚举、分页和 API error envelope | [shared-conventions-codified](../shared-conventions-codified/spec.md) | `shared/conventions.yaml`、Go / TS generated shared packages | 新增或修改共享字面量必须先修订 B1 或对应 owner spec，再同步 YAML 和生成物；共享 enum / error code / `PageInfo` / `ApiError` 不得在业务域私造 |
| HTTP API、公共 schema、fixtures、mock contract 和 breaking-change gate | [openapi-v1-contract](../openapi-v1-contract/spec.md) | `openapi/openapi.yaml`、`openapi/fixtures/`、OpenAPI generated packages / baseline | API inventory、tag、auth 形态、header、status code、schema 和 fixture provenance 以 B2 为准；任何破坏性变更必须走 B2 ADR / diff gate |
| DB 表、索引、enum check、迁移、隐私删除 / 导出不可用例外和 audit 表 | [db-migrations-baseline](../db-migrations-baseline/spec.md) | `migrations/`、`migrations/enum-sources.yaml`、migration lint / probe | 表数量、列名、索引、enum source、migration/backfill 策略和 privacy deletion matrix 以 B4 与实际 migration 为准 |
| internal event、jobType、outbox envelope、dispatcher 与 breaking-change baseline | [event-and-outbox-contract](../event-and-outbox-contract/spec.md) | `shared/events.yaml`、`shared/jobs.yaml`、generated schemas / baselines | `eventName`、`jobType`、API-facing subset、outbox 字段、dispatcher retry/dead-letter 语义以 B3 为准 |
| 本地运行、基础架构约束、配置、secret、feature flag、runtime-config | [engineering-roadmap](../engineering-roadmap/spec.md)、[local-dev-stack](../local-dev-stack/spec.md)、[secrets-and-config](../secrets-and-config/spec.md) | `config/`、`config/feature-flags.yaml`、A2/A4 runtime code 与 deploy assets | 运行时、配置字典、secret 归属、feature flag 和公开配置 allowlist 由 A2/A4 决定 |
| AI provider、model profile、fallback、调用观测字段、prompt / rubric 坐标 | [ai-provider-and-model-routing](../ai-provider-and-model-routing/spec.md)、[prompt-rubric-registry](../prompt-rubric-registry/spec.md) | `config/ai-providers.yaml`、`config/ai-profiles.yaml`、F3 plan-defined prompt / rubric runtime assets | 业务代码只依赖 provider capability、model profile、feature key 和 prompt/rubric 坐标；模型、fallback 和 prompt/rubric 版本必须可审计 |
| metrics、logging、trace、dashboard、alerting、敏感字段红线 | [observability-stack](../observability-stack/spec.md) | F1 spec、后续 logger / metric / alert rule 编码 truth source | metric / label / log 字段 / trace attribute / dashboard / alert 和明文红线以 F1 与实现 gate 为准 |
| 产品模块、一级入口、route、UI 画板标签和用户行为流 | 本文档、`docs/ui-design/`、`frontend/` | UI 文档、静态原型、后续 BDD / E2E 场景 | 产品和 UI 范围先于技术契约；未进入当前产品 / UI 范围的模块不得由技术层私自引入 |

## 2 范围

### 2.1 In Scope

- EasyInterview 当前产品定义、目标用户、JTBD 和产品原则。
- 当前 P0 MVP 闭环：`JD / 简历导入 -> 解析进度 -> 只读目标面试规划 -> 模拟面试 -> Report Dashboard -> 复练当前轮 / 进入下一轮`。
- 当前一级 UI 模块：`首页`、`面试`、`简历`。
- 用户菜单能力：`设置与隐私`、`认证流程`。
- 模拟面试会话边界：连续文本聊天、稳定面试官角色、计时/暂停、结束并生成报告；不维护题号、题目总数、当前题或追问分类。
- 报告边界：session-scoped Dashboard、准备度档位、能力维度、会话证据、下一步行动。
- JD 解析与规划读取边界：首页通过 `importTargetJob` 创建解析命令后才进入 `/parse?targetJobId=...`；该 route 只展示 queued / processing 进度，ready 初读或轮询转 ready 都立即 replace 到 `/workspace?targetJobId=...`。既有 ready 卡片不得重放解析动画。
- 报告索引边界：`/workspace?targetJobId=...` 规划详情右上角进入 `/reports?targetJobId=...`，只展示当前规划 canonical rounds 的 current report 与 latest attempt；不做跨规划中心或完整历史版本列表。
- 简历边界：平铺简历列表、上传 / 粘贴创建、Agent/LLM 解析预览确认、LLM 从简历内容生成有意义的 `displayName`、简历详情只读展示简历内容本身。
- P1/P2/P3 的方向性阶段路线：只围绕当前 UI 已保留能力做工程化、质量化和规模化扩展，并单独标注明确规划例外。
- 评分、Prompt/Rubric/Model 版本追踪、隐私、伦理、数据来源和可观测性红线。
- 与 `docs/ui-design/`、`docs/spec/engineering-roadmap/` 和当前工程契约 owner 的职责边界。

### 2.2 Out of Scope

- 不在本 spec 中编写具体 API schema、DB schema、事件 payload、prompt 文本或 UI 组件代码。
- 不做岗位推荐、全球搜岗或任何岗位发现 / 聚合模块；JD 获取唯一入口是用户在首页文本框粘贴自带 JD。
- 不做简历版本树、主版本 / 岗位定制版本继承或版本管理系统；简历按平铺资产管理。
- 不在本 spec 中承载 engineering-roadmap child subspec 的具体重排；roadmap 拆分与后续 child 创建规则以 `docs/spec/engineering-roadmap/spec.md` 为准。当前 engineering-roadmap active spec 不采用 pending child 索引模型，后续 workstream 限定为当前产品 / UI 已保留能力的 on-demand child。
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
- 未在 UI / UI 文档中出现，也未在本 spec 中明确列为规划例外的内容，默认视为范围外。
- 引入任何范围外能力，必须先原地修订本 spec 和对应 UI 文档，再进入工程 plan；不得只靠 route、画板标签或组件名称扩展产品范围。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 产品真理源 | `docs/spec/product-scope/spec.md` 是当前产品范围入口 | 后续产品评审、plan 设计、scope 判断只能引用当前 active truth source |
| D-2 | UI 设计文档 | `docs/ui-design/` + `frontend/` 是当前 UI / 交互真理源 | 产品 spec 不复述每个控件细节；交互冲突以 UI 文档为准 |
| D-3 | 默认入口 | App 默认进入首页，用户通过唯一文本框粘贴 JD | 不设置未登录欢迎页或登录前置页，不保留 JD 文件、岗位链接或结构化表单等平行导入入口 |
| D-4 | 一级导航 | `首页 / 面试 / 简历` | `复盘`、`岗位推荐`、`当前岗位`、`面试报告`、`成长` 等不作为一级导航；`面试` 是面试规划列表和统一面试规划详情入口，不改变完整模拟面试 session 语义 |
| D-5 | 模拟面试语义 | 面试是一场连续 conversation session；opening、追问、话题转换都是普通 assistant message | 不提供题号、题目预算、当前题、问题/回答/追问分类、热身或单题深钻 |
| D-6 | 语音边界 | 当前 P0 暂不开放电话模式；前端电话图标置灰，后端 voice endpoint fail-closed；通用 speech foundation 可保留 | phone/voice 参数不得 materialize PhoneSurface；重新启用必须重新设计和验收 |
| D-7 | 报告形态 | 报告只有 session-scoped conversation Dashboard 一种形态；desktop 自上而下采用 `3/2/2/2/1`：三项冻结上下文、两个数量指标、两行各两个能力/证据/行动区块、底部全宽面试总评；准备度与 LLM direct semantic summary 只在面试总评展示，mobile 同序单列 | 不按题目/turn，不展示隐藏数值分，前端不二次推导或重复展示报告语义 |
| D-8 | 复练价值承载 | 复练当前轮由后端从 source report 投影；报告存在 issue-backed needs-work dimension 时可以携带 report-local focus，否则创建空 focus 的通用同轮复练；code 仅在单份报告内稳定 | 不设置全局 competency taxonomy、题目回顾、逐题评分、独立错题队列、Drill builder、turn-based retry 或客户端 focus 事实源 |
| D-9 | 真实面试复盘边界 | 复盘不属于当前 P0 模块或后续默认能力 | 不设置 `debrief` route、API、DB、AI feature key 或场景 |
| D-10 | 证据和版本化 | 生成结果必须携带 prompt / rubric / model / language / feature flag / data source 等来源信息 | 支撑质量评估、回归检测和问题追踪 |
| D-11 | 默认范围规则 | 当前 UI / UI 文档未定义且本 spec 未标为规划例外的内容，默认不进入当前范围 | P1/P2/P3 不自动扩展范围外能力 |
| D-12 | Roadmap 对齐 | engineering-roadmap active spec 不采用 `docs/spec/INDEX.md` pending child 索引模型，只保留真实 active spec，并要求后续 child 按当前 Home / Interview / Practice / Report / Resume / Settings/Auth 等保留能力 on-demand 创建 | 后续 child spec / plan 不得用 pending 名称或 route token 引入范围外模块 |
| D-13 | 技术契约 owner matrix | 本 spec §1.5 定义当前工程契约职责、owner spec、编码真理源和统一规则 | 后续文档必须引用当前 owner spec 和编码 truth source；不得引入未被 owner 承接的实施依据 |
| D-14 | JD 导入命令与规划读取分离 | 首页先调用 `importTargetJob` 创建解析命令，再进入 `/parse?targetJobId=...` 查看 queued / processing 进度；ready 初读或轮询转 ready 后立即 history replace 到 `/workspace?targetJobId=...` 只读规划详情 | `/parse` 不展示 ready 规划详情、不承接报告入口，也不得为既有 ready 规划重放解析动画；`/workspace` 详情不重新触发 import |
| D-15 | 复盘上下文选一带二 | 复盘上下文选择不属于当前产品能力 | 不保留相关 UI / API / 场景 |
| D-16 | 无密码认证唯一流 | 邮箱验证码是唯一登录方式；不提供独立“重置登录 / 密码重置”页面 | 验证码重发与更换邮箱在 `auth_verify` 内完成；`auth_reset` route 输入归一回 `auth_login`，文档与 UI 不出现“忘记密码 / 密码 / 两步验证”口径 |
| D-17 | 岗位推荐模块边界 | `岗位推荐 / Job Picks / jd_match` 不属于当前一级模块；JD 获取唯一入口是首页粘贴 | 一级导航为三项；主流程不包含岗位推荐来源；Q-2 关闭；`全球多平台搜岗` 不属于当前范围 |
| D-18 | 公司情报仅保留轻量嵌入卡片 | `company_intel` 独立详情页与 route 不属于当前范围；模拟面试规划页内嵌轻量卡片是公司情报唯一呈现 | 轻量卡片仍展示一句话画像、近期公开信号、反问建议与合规来源说明；P2 生产化仍以嵌入卡片为呈现边界 |
| D-19 | 报告 CTA 单点收敛 | 报告页只保留 Header 一对 CTA：`复练当前轮` 与 `进入下一轮`；Next 详情只承载路径说明与能力重点 | 报告任何二级详情不得再出现重复开练按钮或 per-question replay toggle |
| D-20 | 简历资产扁平化与详情只读 | 简历是平铺列表中的独立资产，不区分原始简历 / 结构化主版本 / 岗位定制版本，不做版本树、分叉、继承或版本管理；创建仅上传 / 粘贴；LLM 解析结果必须产出有意义的 `displayName`，不得长期停留在“上传的简历 / 粘贴的简历”等通用标题；详情页只读展示简历内容本身，不提供导出 PDF、复制、编辑、预览 tab、改写建议或查看原始简历预览 | 每份简历仍保留原始来源与解析文本快照作为后台证据；面试绑定对象为简历（`resumeId`）；原始简历预览就是详情中的简历正文 |
| D-21 | 全局呈现收敛 | 设置页只保留个人资料与隐私数据两个 tab；主题只保留 `深海（Ocean）` / `梅紫（Plum）` 两个预设 + `自定义 accent` 的色相 / 饱和度滑杆 + 暗色模式 + 语言下拉 + 字体预设，默认主题为 `深海（Ocean）` | 自定义 accent 不展示 preview、数值文本或“恢复主题默认色”按钮；选择 Ocean / Plum 即退出 custom。通知、订阅等范围外能力未来按需重新设计，不以空 tab 形式预留 |
| D-22 | 核心闭环方案 B | 真实面试复盘和用户画像不属于当前功能范围：`debrief` / `profile` 不再是目标 route，`Debriefs` / `Profile` OpenAPI tags、`debriefs` / `candidate_profiles` / `experience_cards` DB 表、debrief/profile AI feature key、shared event/job 和正向场景不保留 | P0 只保留 JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮；账号资料补全和设置隐私保留，但不承担候选人画像语义 |
| D-23 | 面试入口 landing | `/workspace` 是面试规划列表；`/workspace?targetJobId=...` 是当前面试规划只读详情，Workspace query 只允许 `targetJobId`，丢弃 `planId` / `resumeId` 等冗余身份 | ready 规划卡片直接进入 Workspace 详情；列表使用当前 `listTargetJobs` 规划候选，不新增独立多轮计划或泛岗位资产中心 |
| D-24 | Practice conversation 简化 | 删除前后端全部题目/turn 结构、专用 hint/mode 和逐题报告；Practice 只保留 ordered user/assistant messages | `questionBudget`、`PracticeTurn`、QuestionCard/SessionMap、question assessments、hint event/count 均不属于当前范围 |
| D-25 | Home JD intake 唯一入口 | Home 只保留 JD textarea、ready Resume 下拉框和「立即面试」CTA；`importTargetJob` 的当前请求体固定为 `{ rawText, targetLanguage, resumeId }` | JD 文件、岗位链接和结构化表单导入不属于当前产品范围；Resume 模块自己的上传 / 粘贴能力不受影响 |
| D-26 | Practice 即时反馈与可恢复发送 | 提交后 user message 立即出现；等待期间锁定输入并显示面试官思考；只有服务端标记可重试的失败消息在原 row 下显示 retry。原 `clientMessageId/replyStatus` 由后端持久化并经会话读模型恢复 | 刷新/重挂载不丢失待回复身份，不用 local/session storage 充当业务事实源；同 ID 重试最终只产生一个 assistant reply |
| D-27 | 当前规划报告入口 | Workspace 只读规划详情内容区右上角进入 `/reports?targetJobId=...`；ReportsScreen 只索引当前 TargetJob canonical rounds 的 current report 与 latest attempt，Reports Back 返回 `/workspace?targetJobId=...`；Report/Generating trusted Back 返回该列表 | 不加入 TopBar，不嵌入 Parse，不显示其他规划或完整历史，不改数据库/OpenAPI Schema；无可信 target 时安全回 `/workspace` |
| D-28 | ready 规划统一回访入口 | Home / Workspace ready 卡片、Reports Back 和 Practice terminal recovery 都进入 `/workspace?targetJobId=...` | 这些只读回访动作不得导航到 `/parse`、重放解析动画或再次调用 `importTargetJob` |
| D-29 | 报告归属的只读会话记录 | 一份报告附属一份完成会话记录；用户只以 `reportId` 进入 `/report-conversation?reportId=...` | 不提供会话历史列表或 `sessionId` 用户路由；复用现有 `feedback_reports.session_id` 唯一关系，不新增关系表；queued/generating/ready/failed 报告都可查看只读 Markdown transcript |

### 3.2 待确认事项

| ID | 待确认事项 | 影响 | 默认处理 |
|----|------------|------|----------|
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

- 顶部导航只能出现当前 UI 设计文档确认的三个一级入口。
- “面试报告”只允许作为 Workspace 规划详情内容区的页面级入口和 target-scoped 上下文页面；不得进入 Parse、TopBar 或成为无上下文全局中心。
- `/parse` 的产品语义是 import 后的命令进度页，只允许 `targetJobId`；queued / processing 可轮询，ready 必须 replace 到 Workspace 详情，failed 显示受控失败与恢复路径。
- `workspace` 的产品语义是 `面试 / 面试规划列表 / 当前面试规划`，不是一级岗位资产管理模块；`/workspace` 展示列表，`/workspace?targetJobId=...` 只读展示详情，query 只保留 `targetJobId`，缺数据时给出导入 JD 的友好空态。
- `generating` / `report` 只以 `reportId` 定位；status/error、session/岗位/简历/轮次与 CTA identity 必须来自报告完成时冻结的后端 context 投影，route 不得成为业务事实源。
- `report-conversation` 是 Report 的只读附属页，只以 `reportId` 定位；不得暴露 `sessionId`、创建会话历史一级入口或让 Workspace 消费会话列表。主入口位于 Report Dashboard Context Strip 下方，ReportsScreen 当轮 current report 行可提供快捷入口。
- `debrief` / `debrief_full` / `profile` 不是目标 route；对应路径或 hash 输入必须归一到当前核心入口。
- 当前 phone/voice route/query 输入统一归一为文本 `practice`；不得 materialize PhoneSurface 或独立 voice route。
- route、画板标签和组件名称不得单独作为新增产品能力的依据。

### 4.3 评分与反馈约束

报告可以展示：

- 准备度档位：例如 `未就绪 / 建议再练 / 基本可面 / 较为充分`。
- 维度状态：例如 `强项 / 达标 / 待加强`。
- 置信度或证据充分性说明。
- grounded 会话证据摘要、证据缺口、推荐框架和下一步复练建议；用户可见报告不恢复题目/turn UI。

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
- 范围外能力必须先在本 spec 中重新设计和批准。
- route、画板标签或组件名称不能单独作为新增产品能力的依据。

## 5 模块边界

### 5.1 产品能力层

EasyInterview 的完整能力仍围绕五个产品层组织，但 UI 一级入口和底层能力层不是一一对应关系。

| 能力层 | 当前职责 | 当前 UI 承载 |
|--------|----------|--------------|
| M1 · Context Inputs | 通过简历、JD 和模拟面试上下文支撑面试定制与报告解释 | 首页、简历、面试和报告中的上下文，不提供用户画像页 |
| M2 · Target Job Workspace | 围绕 JD 管理岗位要点、简历绑定、轮次假设、公司轻情报和会话记录 | 首页 JD 导入、面试规划列表与统一面试规划详情 |
| M3 · Mock Interview Orchestrator | 组织连续文本 conversation、稳定上下文和 session 生命周期 | Interview Session |
| M4 · Evidence-based Review | 产出准备度、能力维度、会话证据和下一步行动 | Report Dashboard |
| M5 · Growth Signals | 准备度变化、报告后动作和练习趋势等横切信号 | 不作为独立成长中心；仅能嵌入报告、模拟面试规划等现有模块 |

### 5.2 当前一级 UI 模块

| 一级模块 | 用户任务 | P0 职责 | 不承担的职责 |
|----------|----------|---------|--------------|
| 首页 | 快速开始一次岗位准备 | 粘贴 JD、选择 ready 简历、最近模拟面试、创建简历入口 | 不做登录前营销页；不做复盘辅助入口；不提供其他 JD 导入形态 |
| 面试 | 浏览并回访既有面试规划，再次发起 session | `/workspace` 面试规划列表、`/workspace?targetJobId=...` 当前面试规划只读详情（回访枢纽）、公司轻情报嵌入卡片、会话记录、报告入口、立即面试 | 不作为泛岗位资产管理中心；不在 `/parse` 复刻 ready 详情或解析动画 |
| 简历 | 管理可被岗位和面试消费的简历资产 | 平铺简历列表、上传 / 粘贴创建、解析预览确认、LLM 生成可识别简历名称、只读简历详情 | 不做版本树 / 主版本 / 岗位定制继承；不做复杂排版设计器；详情页不做导出、复制、编辑、改写建议或原件弹层 |

### 5.3 会话级和上下文页面

| 页面 | 归属 | 进入方式 | 关键上下文 |
|------|------|----------|------------|
| `parse` | JD 解析命令进度 | 首页 `importTargetJob` 成功后 | `targetJobId` |
| `workspace` | 面试规划列表 / 只读详情 | 一级导航、ready 规划卡片、Reports Back、Practice terminal recovery | 可选 `targetJobId`；不得保留 `planId / resumeId` |
| `practice` | Interview Session | 模拟面试规划、报告复练、进入下一轮 | `sessionId / targetJobId / resumeId / roundId` |
| `reports` | 当前规划报告索引 | Workspace 规划详情内容区右上入口、Report/Generating trusted Back | `targetJobId` |
| `generating` | 报告生成过渡态 | 面试结束 | `reportId` |
| `report` | Report Dashboard | 面试结束、会话记录、相关入口 | `reportId` |
| `report-conversation` | 报告附属只读会话记录 | Report Dashboard 主入口、ReportsScreen 当轮 current report 快捷入口 | `reportId`；不得使用 `sessionId` |
| `settings` | 设置与隐私 | 用户菜单 | `userId` |

### 5.4 当前范围外能力

| 能力 | 当前边界 | 说明 |
|------|----------|------|
| 未登录欢迎页 | 范围外 | App 默认进入首页，降低首次价值路径阻力 |
| 当前岗位一级导航 | 范围外 | 岗位信息归属于模拟面试规划 |
| 面试报告全局一级导航 | 范围外 | 仅保留规划详情内容区入口和 target-scoped ReportsScreen；报告详情仍隶属于 session |
| 练习模式卡片 | 范围外 | 面试是一场连续 conversation，不提供 strict/assisted 选择 |
| 热身 / 单题深钻 / 反问专练 | 范围外 | 对话由 AI 结合当前上下文自然推进，不建立子模式 |
| 追问树 | 范围外 | 不维护问题/追问分类或树结构 |
| 独立错题队列 | 范围外 | 复练使用能力缺口和会话证据，不使用题目集合 |
| 成长中心 | 范围外 | 准备度和趋势信号只嵌入报告或面试规划 |
| 多轮计划 | 范围外 | 轮次节点在模拟面试规划和报告 CTA 中表达 |
| 经历库 / STAR 编辑器 | 范围外 | 经历证据由简历和面试上下文承载 |
| 报告时间线 / 刊物式报告 | 范围外 | 报告统一为 Dashboard |
| 岗位推荐一级模块 | 范围外 | JD 获取唯一入口是首页导入 |
| 公司情报独立详情页 | 范围外 | 轻量情报由模拟面试规划页嵌入卡片承载 |
| 简历版本树 / 主版本 / 岗位定制版本 | 范围外 | 简历按平铺资产管理，不做版本继承 |
| 轻量问答建档 | 范围外 | 创建简历只保留上传 / 粘贴 |
| 设置页通知 / 订阅 tab | 范围外 | 设置页只保留个人资料与隐私数据 |
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
  -> 在唯一文本框粘贴 JD
  -> importTargetJob（创建解析命令）
  -> Parse Progress（仅 queued / processing）
  -> ready 后 replace Workspace Detail（只读规划详情）
     ├─ 查看 JD 基础信息 / 必需项 / 加分项 / 隐性关注点
     ├─ 查看已绑定简历
     ├─ 查看 canonical InterviewRound
     ├─ 查看公司轻情报 / 会话记录 / 报告入口
     └─ 立即面试
  -> Interview Session
  -> 结束并生成报告
  -> Report Dashboard(reportId)
     ├─ 查看本次面试记录
     ├─ 复练当前轮
     └─ 进入下一轮
```

首次导入链路把命令进度与只读查询明确分开：`importTargetJob` 是唯一创建解析工作的命令，`/parse?targetJobId=...` 只展示 queued / processing；ready 初读或轮询转 ready 后立即 replace `/workspace?targetJobId=...`。已创建的 `Mock Interview Plan` 统一从 Home / Workspace ready 卡片、Reports Back、Practice terminal recovery 回到 Workspace 详情，不重新进入 Parse 或重放解析动画。

### 6.3 主流程 B：先补简历的用户

```text
Home 或 Resume
  -> 1 分钟创建简历 / 新建简历
  -> 上传 / 粘贴
  -> Agent 解析
  -> 预览确认保存
  -> 回到 Home，在提交 JD 前从 ready 简历下拉框选择这份简历
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
- Home / Workspace ready 卡片直接回到 `/workspace?targetJobId=...` 当前规划。

#### 系统处理

- 解析岗位标题、公司、地区、语言、职级和来源。
- 提炼必需项、加分项、隐性关注点和风险提示。
- 使用 import 时已选择的 ready 简历，并展示解析得到的 canonical 面试轮次。
- 展示公司轻情报嵌入卡片和面试轮次节点。
- 展示当前规划下的模拟面试记录。

#### 当前不进入本模块

- 岗位推荐 / 全球多平台搜岗：不属于当前功能范围，不再是规划例外。
- 复杂职位收藏生态。
- 大规模外部公司情报拼图与独立情报详情页。
- 把 `当前岗位` 做成独立一级模块。

#### 验收标准

- 任意一份可读 JD 都先经 `importTargetJob` 创建解析工作；只有 queued / processing 可留在 Parse，ready 必须进入 Workspace 只读详情。
- 直接打开 ready 卡片只执行只读查询，不进入 Parse、不重放动画、不再次 import。
- 模拟面试规划（回访枢纽）必须明确 `TargetJob / JD / 简历 / InterviewRound`，且 Workspace route 只以 `targetJobId` 定位。
- 当前规划记录不得混入其他公司、岗位或 JD 的会话。

### 6.7 M3：连续模拟面试

#### 目标

围绕目标岗位组织一场连续文本模拟面试，由 AI 根据上下文和聊天历史自然推进，而不是维护预设题目集合或让用户选择复杂练习模式。

#### 当前面试形式

- 只开放文本 conversation。
- 电话图标置灰且不可点击；phone/voice 参数不产生电话页面。
- 用户需要提示时直接发送普通聊天消息，不存在专用提示模式。

#### 关键逻辑

- 面试会话必须有 `sessionId` 和稳定 InterviewContext。
- 面试官角色由当前 round/plan 决定，同一 session 内不切换。
- opening 与后续回复都是普通 assistant message；AI 必须结合最近消息、目标 JD、简历和能力重点自然推进。
- 用户提交后先立即看到自己的消息；服务端未返回时输入框不可提交并显示面试官思考。失败后思考消失，仅可重试失败在原消息下提供 retry；刷新后由会话 API 恢复 pending/failed/complete 状态和原消息 ID。
- 页面只保留 Top Bar、全宽 Transcript、Composer 和全局 `结束并生成报告`。
- opening assistant message 不算候选人作答；至少提交一条 user message 后才允许结束。前端在此之前禁用结束 CTA 并给出本地化、可访问原因，后端仍以 typed `VALIDATION_FAILED` 作权威校验，且不得创建 report/job/outbox 或完成 session。

#### 当前不进入本模块

- 入口前热身、反问专练、单题深钻、追问树或 Drill builder。
- 题号、题目总数、当前题、题目地图、QuestionCard、追问/下一题分类。
- 专用 hint button/event/count 或 strict/assisted mode。
- 可用 PhoneSurface、麦克风、字幕、VAD、TTS 或底层 voice 产品入口。
- 真实面试中的隐形实时辅助。

#### 验收标准

- 用户从 Workspace 只读规划详情点击 `立即面试` 后直接进入完整 session。
- 用户只看到连续文本聊天，不看到任何题目结构。
- 刷新后完整 ordered messages 可恢复。
- 电话图标置灰，后端 voice 调用 fail-closed。
- 零回答会话不能结束或生成报告；提交首条 user message 后结束动作才可用。
- report job / event 只携带稳定 session/context IDs；服务端在会话完成时冻结 JD、绑定简历、轮次、训练目标和消息坐标，并在生成时按 ID 加载该快照与 terminal ordered messages。

### 6.8 M4：证据化报告

#### 目标

给出足够具体、可操作、可复练的反馈，让用户知道下一次该怎么练。

#### 报告结构

- ReportsScreen：从 Workspace 只读规划详情进入，仅按当前 TargetJob canonical rounds 显示 `currentReport/latestAttempt`，覆盖 loading/empty/error/identity mismatch；它是规划范围索引，不是报告内容页，不展示完整历史，Back 返回同一 Workspace 详情。
- Header：目标岗位、轮次、会话、绑定简历和报告归属说明。
- Context Strip：`sessionId`、目标岗位、轮次、简历。
- Summary Metrics：准备度 + summary、能力维度数量、会话证据数量。
- Detail Grid：能力维度、优势证据、风险 / 待加强证据、下一步行动四个常驻区块，不设置 tab。
- Next Actions：Header 的复练当前轮和进入下一轮是报告唯一一对开练 CTA（D-19）；第一条服务端 action 只决定两枚现有 CTA 的推荐主次。

#### 会话证据

Evidence Surface 展示会话级 highlights / issues 的 dimension label、证据摘要和本地化置信度，不复制完整 transcript，也不按题号或 turn 分组。每个主要判断在后端保留候选人消息 grounding anchor；未回答的 assistant 追问只能表达为“未覆盖 / 证据不足”，不能据此形成负面能力结论。复练当前轮由后端从 source report 投影可选的 report-local dimension focus；没有可支持 focus 时创建空 focus 的通用同轮复练，不使用题目 ID 或客户端 focus。

#### 当前不进入本模块

- 报告全局一级导航或跨规划中心。
- 无 session 上下文的报告页。
- 时间线报告。
- 刊物式报告页。
- 题目回顾、逐题评分、独立错题队列或单题 retry。
- 精确通过率或录用概率。
- 与后端状态无关的生成百分比、固定“实时观察”或未实现的通知承诺。

#### 验收标准

- 报告直接使用完成时冻结的 JD、简历、轮次、目标与 terminal conversation；每个 evidence 能回溯到候选人 user message，且模型最终语义不会被后端隐藏分数二次改写。
- 用户看完报告后能直接选择复练当前轮或进入下一轮。
- 无 `reportId` 或 API 返回 missing/invalid current contract 时必须显示专用缺失/失败状态，并回到当前面试规划；route 中的 session/岗位/简历/轮次参数不得补造报告事实。
- 报告资源存在时，用户可从报告详情或 ReportsScreen 当轮 current report 行打开只读会话记录；queued/generating/ready/failed 共享同一 report-owned 访问边界，跨用户或缺失 report 统一 hidden 404。

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
- Home import 可以从 ready 简历列表选择任意一份简历作为新 TargetJob 的绑定对象；Workspace 只读详情展示该绑定事实。

### 6.10 报告后行动边界

报告后的行动只保留 `复练当前轮` 与 `进入下一轮`。真实面试复盘、复盘分析、复盘面试和 debrief-derived practice plan 不属于当前功能范围；`debrief` 不作为 route、OpenAPI tag、DB 表、AI feature key、shared event/job 或 E2E 场景。

### 6.11 范围外能力

本节是当前产品范围的负向边界。

| 能力 | 当前判断 | 原因 |
|------|----------|------|
| 岗位推荐一级模块 | 范围外 | 超出 MVP 闭环；JD 获取唯一入口是首页导入 |
| 全球多平台搜岗 | 范围外 | JD 获取唯一入口是首页导入；如需引入新来源，先修订 §2.4 |
| 公司情报独立详情页 | 范围外 | 轻量情报由模拟面试规划页嵌入卡片承载 |
| 简历版本树 / 主版本 / 岗位定制版本 / 轻量问答 | 范围外 | 简历按平铺资产管理；创建只保留上传 / 粘贴 |
| 视频情绪识别 | 范围外 | 解释风险高，训练价值不稳定，且容易引入不当评估 |
| 社区 | 范围外 | 稀释单人岗位准备闭环，不在当前产品定位内 |
| Team / EDU | 范围外 | 当前产品只面向个人训练闭环，不规划团队版或组织评估产品 |
| 企业端候选人评估 | 范围外 | 与训练产品定位冲突，伦理负担高 |
| 独立成长中心 | 范围外 | 当前只保留嵌入报告和面试规划的准备度 / 趋势信号 |
| 独立多轮计划 | 范围外 | 面试轮次只在模拟面试规划和报告 CTA 中表达 |
| 独立经历库 / STAR 编辑器 | 范围外 | 经历证据由简历和面试上下文承载 |
| 独立错题本 / 单题 Drill / 追问树 | 范围外 | 当前没有题目/turn 模型；复练基于能力缺口和会话证据 |
| 报告时间线 / 刊物式报告 | 范围外 | 报告统一为 session-scoped Dashboard |
| 真实面试复盘 / Debrief | 范围外 | 当前核心闭环收敛到模拟面试报告后的复练 / 下一轮，不维护平行复盘系统 |
| 用户画像 / CandidateProfile / ExperienceCard | 范围外 | 账号资料和设置隐私保留，但不再沉淀独立候选人画像产品或数据模型 |
| 隐形实时面试辅助 | 范围外 / 长期不做 | 明确违反产品伦理边界 |

## 7 阶段路线

### 7.1 P0：当前 MVP

P0 聚焦一条完整闭环：

```text
JD 导入
  -> Parse 命令进度
  -> Workspace 当前面试规划只读详情
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
| C-2 | 首页直达价值 | 用户未登录且有 JD | 打开 App 首页 | 可以在唯一文本框粘贴 JD；页面不存在其他 JD 导入入口 | `frontend-home-job-picks-and-parse/001` |
| C-3 | 一级导航一致 | 当前 UI 文档已锁定三个一级入口 | 查看 TopBar | 只出现首页、面试、简历；点击面试默认进入面试规划列表 | `docs/ui-design` |
| C-4 | 模拟面试上下文 | TargetJob 已 ready，用户位于 `/workspace?targetJobId=...` 只读规划详情 | 点击立即面试 | session 带入目标岗位、JD、已绑定简历和轮次；不返回 Parse 或出现第二个全页确认 | frontend-workspace-and-practice/001 |
| C-5 | 连续聊天清晰 | 用户进入 Interview Session | 连续发送消息 | 只看到有序 user/assistant messages，无题号、题目地图、专用提示或 mode switch | `docs/ui-design/module-practice-review.md` |
| C-6 | 报告归属清晰 | 一场面试完成 | 进入报告页 | 报告展示 session、岗位、轮次、简历、能力维度和会话证据 | `docs/ui-design/report-dashboard.md` |
| C-7 | 下一步动作清晰 | 用户在报告页查看复练计划 | 点击 CTA | `复练当前轮` 和 `进入下一轮` 分别直接进入对应 session | `docs/ui-design/report-dashboard.md` |
| C-8 | 复练不依赖题目 | 用户查看报告 | 点击复练当前轮 | 能力缺口进入新 plan，不存在题目选择或 turn ID | `docs/ui-design/module-practice-review.md` |
| C-9 | 简历资产可绑定 | 用户已有 ready 简历 | 在 Home 提交 JD 前打开简历下拉框 | 可以选择列表中的任意一份 ready 简历，并由 `importTargetJob` 持久化到新 TargetJob；Workspace 详情只读展示绑定结果 | `docs/ui-design/resume-module.md` |
| C-12 | 证据化和版本化 | 任一 AI 生成结果产生 | 保存或展示结果 | 结果可追踪 prompt / rubric / model / language / feature flag / data source | 后续 backend / quality child |
| C-13 | 默认范围规则 | 某能力没有进入当前 product-scope 和 UI design document | 评审后续需求或 plan | 该能力默认不进入当前范围，除非先修订本 spec 和对应 UI 文档 | docs-only |
| C-15 | 无密码认证唯一流 | 产品只有邮箱验证码登录 | 查看认证页面流 | 不存在独立重置登录页；验证码重发与更换邮箱在 `auth_verify` 内完成 | `docs/ui-design/auth-and-entry.md` |
| C-16 | 岗位推荐与情报独立页零入口 | 岗位推荐模块与公司情报独立页不属于当前范围 | 走查导航、首页与静态原型路由 | 不存在 `jd_match` / `company_intel` 目标 route、岗位推荐入口或独立情报详情页；公司情报只出现在模拟面试规划页嵌入卡片 | `docs/ui-design/module-map.md` |
| C-17 | 简历平铺与采纳收口 | 用户在简历模块管理资产并接受改写 | 查看简历列表并采纳改写建议 | 列表是单层平铺、无树 / 主版本 / 定制版本概念；改写建议仅有`采纳`，采纳后确认前预览可选覆盖原简历或保存为新简历 | `docs/ui-design/resume-module.md` |
| C-18 | 报告 CTA 单点 | 报告已生成 | 走查报告页全部区块 | 只有 Header 一对 `复练当前轮 / 进入下一轮` CTA；详情不出现重复开练按钮或 per-question toggle | `docs/ui-design/report-dashboard.md` |
| C-20 | 题目模型零残留 | D-24 已生效 | 检查 UI/API/DB/Prompt/report/scenarios | 无 questionBudget/PracticeTurn/QuestionCard/question assessment/hint positive contract；voice 当前 fail-closed | backend/frontend/report owners |
| C-21 | JD intake 单一合同 | D-25 已生效 | 检查 UI 设计文档、Home、OpenAPI、generated artifacts、backend 与 active scenarios | 唯一正向请求是 `{ rawText, targetLanguage, resumeId }`；JD 文件、岗位链接、结构化表单及其专属 handler、fixture、场景和 UI 锚点为零；Resume 上传仍可用 | `frontend-home-job-picks-and-parse/001` + contract/backend owners |
| C-22 | Practice 发送与刷新恢复 | D-26 已生效，AI 首次成功、可重试失败或终态失败 | 提交、等待、失败、刷新并按需重试 | user row 立即出现；pending 锁输入并显示 thinking；retry 只在可重试失败 row 下；刷新恢复原 `clientMessageId/replyStatus`；同 ID 成功后 user/reply 各唯一一条 | backend-practice/002 + frontend-workspace-and-practice/002 + openapi-v1-contract/001 |
| C-23 | 当前规划报告隔离 | D-27 已生效，用户从 Workspace 详情进入报告列表或从 Report/Generating 返回 | 打开 `/reports?targetJobId=...` | 只显示当前规划 current/latest；Reports Back 返回同一 Workspace 详情；TopBar 与 Parse 无报告入口/嵌入列表；跨 target/mismatch/stale fail closed，无完整历史 | frontend-workspace 001 + frontend-report 001 + frontend-shell 004 |
| C-24 | JD 命令与只读详情分路 | 用户刚完成 import，或点击既有 ready 规划卡片 | 进入相应 route | 仅刚 import 的 queued/processing 工作进入 `/parse?targetJobId=...`；ready 初读/轮询用 replace 进入 `/workspace?targetJobId=...`；ready 卡片、Reports Back、Practice terminal recovery 直达 Workspace，且 Workspace query 只有 `targetJobId` | frontend-home 001 + frontend-workspace 001/002 + frontend-report 001 + frontend-shell 004 |
| C-25 | Custom Accent 最小控制 | 用户打开主题菜单 | 选择自定义 accent 或 Ocean/Plum | custom 只显示色相/饱和度滑杆，无 preview/value/reset；选择 Ocean/Plum 退出 custom，主题仍可正常切换 | frontend-shell/002 component/parity gates + root `make test` |
| C-26 | 报告附属会话记录 | owned report 已创建，状态可能为 queued / generating / ready / failed | 用户从报告详情或 ReportsScreen 当轮 current report 快捷入口打开记录 | 以 `reportId` 返回该报告唯一的 ordered user/assistant Markdown transcript；页面只读并返回同一报告，不暴露 `sessionId`，不存在会话列表或额外关系表 | frontend-report 001 + backend-review 001 + openapi-v1-contract 001/002/003 + E2E.P0.099 |
| C-19 | 复盘和用户画像零入口 | 当前 P0 只保留 JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮 | 走查 TopBar、用户菜单、URL/hash route、OpenAPI、DB、shared、config 和场景索引 | 不存在 `debrief` / `profile` 目标 route、`Debriefs` / `Profile` OpenAPI tag、`debriefs` / `candidate_profiles` / `experience_cards` 表或正向场景 | [001-core-loop-module-pruning](./plans/001-core-loop-module-pruning/plan.md) |

## 9 质量、安全与评估

### 9.1 质量评估

P0 质量评估必须覆盖三类指标：

| 类型 | 指标 |
|------|------|
| 体验类 | JD 导入完成率、进入模拟面试率、报告查看率、报告后复练率、进入下一轮率 |
| 内容类 | JD 解析准确性、对话与岗位相关性、上下文连续性、报告证据可追溯率、简历建议可用性 |
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
- 原始 JD 与导入记录。
- 面试会话消息与报告证据。
- 若未来重新开放语音，音频、媒体对象和 STT 中间结果必须重新进入隐私设计；当前路径不得产生这些数据。

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

- UI 设计文档：[docs/ui-design/README.md](../../ui-design/README.md)、[docs/ui-design/ui-architecture.md](../../ui-design/ui-architecture.md)、[docs/ui-design/module-map.md](../../ui-design/module-map.md)。
- 工程 roadmap：[docs/spec/engineering-roadmap/spec.md](../engineering-roadmap/spec.md)。
- 当前技术契约 owner matrix：本文档 §1.5。
- OpenAPI 契约：[docs/spec/openapi-v1-contract/spec.md](../openapi-v1-contract/spec.md)。
- Prompt / Rubric：[docs/spec/prompt-rubric-registry/spec.md](../prompt-rubric-registry/spec.md)。

本 spec 当前挂载 [001-core-loop-module-pruning](./plans/001-core-loop-module-pruning/plan.md) 作为核心闭环范围收敛 owner。后续如果要调整整体 P0/P1 阶段，必须原地修订该 owner；命中既有 child subject 的用户可见行为修订，必须先原地重开对应 child plan / checklist / context.yaml 并同步本文，再进入 `/implement`，不得创建同主题 sibling plan。

## 11 修订记录

| 版本 | 日期 | 修订内容 |
|------|------|----------|
| 2.24 | 2026-07-15 | D-29：合并 report-owned 只读会话记录，以 reportId-only Markdown 页呈现；报告资源存在即允许查看，不新增 session 历史列表、sessionId 用户路由或关系表。 |
| 2.23 | 2026-07-15 | D-7：用户确认报告 ready 页采用 `3/2/2/2/1`；准备度与服务端 summary 从顶部指标下移为底部全宽面试总评，mobile 保持同序单列。 |
| 2.22 | 2026-07-14 | D-14/D-23/D-27/D-28：Parse 仅承接 import 后 queued/processing 命令进度，ready replace 到 targetJobId-only Workspace 只读详情；ready 卡片、Reports Back、Practice terminal recovery 统一回 Workspace；报告入口迁至 Workspace。D-21 同步为 Ocean/Plum + hue/saturation-only custom accent，无 preview/value/reset。 |
| 2.21 | 2026-07-14 | D-27：规划详情右上角进入 target-scoped ReportsScreen，当前规划 current/latest-only；Parse 解耦、TopBar 无入口、trusted Back 返回 Reports、无可信上下文回 workspace。 |
| 2.20 | 2026-07-13 | D-26：Practice 用户消息即时显示、等待态 thinking/输入锁、失败 row-local retry，并由后端持久化 reply state 支持刷新后同 ID 恢复。 |
| 2.19 | 2026-07-13 | D-25：Home JD intake 收敛为唯一粘贴文本框与 `{ rawText, targetLanguage, resumeId }` 请求合同，删除其他 JD 导入形态，同时明确 Resume 上传不受影响。 |
| 2.17 | 2026-07-12 | 明确零回答不可完成、空 focus 的通用同轮复练，以及 reportId-only 缺失态；移除 sessionId 报告 locator 漂移。 |
| 2.16 | 2026-07-12 | 报告统一为 reportId-only 深链，状态/上下文/CTA identity 来自冻结后端投影；复练 focus 明确为单份报告内 dimension code。 |
| 2.15 | 2026-07-12 | D-7/D-8：报告采用冻结完整上下文与 LLM direct semantic output；三指标四常驻区块展示 grounded evidence，复练 focus 由后端 source report 投影，生成页不得伪造进度/观察/通知。 |
| 2.14 | 2026-07-12 | D-24：Practice 收敛为连续文本 conversation，删除题目/hint/逐题报告，电话模式置灰且后端 fail-closed。 |
| 2.13 | 2026-07-10 | 统一 owner spec 的中文范围边界为“范围外”，并同步执行 plan 的 scope-boundary gate 口径。 |
| 2.9 | 2026-07-08 | 将一级导航 `模拟面试` 收敛为更简洁的 `面试`，并规定 `workspace` 无上下文 landing 为面试规划列表，带上下文时进入面试规划详情。 |
| 2.7 | 2026-07-07 | 将会话记录、模拟面试记录和报告返回入口统一为当前产品术语，避免 active spec 把记录能力写成过期口径。 |
| 2.6 | 2026-07-07 | 将 product-scope 正文收敛为当前合同表达；中文范围边界只描述当前行为和范围外。 |
| 2.5 | 2026-07-06 | 将 active product-scope 中的范围变更过程说明改为当前范围合同与负向边界表述。 |
| 2.2 | 2026-06-29 | 锁定当前核心闭环：JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮；新增 D-22、C-19 和 001-core-loop-module-pruning owner plan。 |
| 2.1 | 2026-06-12 | 确认主题 `自定义 accent` 模式保留，默认主题改为 `深海`；设置页范围收敛为当前 tab 结构。 |
| 2.0 | 2026-06-12 | 收敛一级导航、公司情报嵌入卡片、报告 CTA、flat Resume 和设置页范围；更新 §5.1-§5.4、§7、C-3/C-9 并新增 C-16/C-17/C-18。 |
| 1.9 | 2026-06-12 | 锁定 JD 导入单次确认、workspace 回访枢纽、复盘上下文范围边界和无密码认证唯一流；同步主流程、模块页面表与验收场景。 |
| 1.8 | 2026-05-05 | 既有基线（见 history.md） |
