# AGENT CODING 指令

## 1 项目概述

easyinterview 是一款围绕真实 JD、目标岗位、简历资产和真实面试复盘设计的 AI 面试训练产品。项目主张是把用户从“拿到一份 JD”到“完成一场有上下文的模拟面试、获得证据化报告、决定复练当前轮或进入下一轮，并在真实面试后复盘”的准备过程做成可执行、可追踪、可反复改进的训练工作台。

当前项目仍处于开发阶段，尚未上线，不需要保留线上兼容性或历史 route / 旧模块兼容层。实现和文档应以当前 active spec、`docs/ui-design/` 与 `ui-design/` 静态原型为准；发现旧 spec、旧 route、旧画板标签或 dead code 与当前设计冲突时，优先删除或原地修订旧内容，而不是为了兼容历史形态保留平行入口。

当前仓库的主要目录结构如下：

- `docs/`：产品范围、工程 spec、plan、UI 设计文档、工作日志、Bug 知识库和报告。
- `ui-design/`：当前静态 UI 原型与 UI 契约测试，是前端交互设计的直接参照。
- `openapi/`、`shared/`、`backend/`、`frontend/`、`migrations/`、`config/`、`scripts/`：API、共享契约、后端、前端、数据库、配置和工程脚本资产。
- `.agent-skills/`：仓库内自定义 Skill、模板和共享校验脚本。
- `test/scenarios/`：场景测试框架、环境说明和 BDD / E2E 场景资产。

### 1.1 前端实施 UI 真理源

前端创建新页面、新组件、新 token 时必须以 `ui-design/` 静态原型源码和 `docs/ui-design/` 文档作为唯一 UI 真理源。正式前端的视觉目标是 100% 源级复刻 `ui-design/`；AI 不得基于自身审美自由生成正式前端视觉，也不得引入外部品牌设计系统作为参考替代。

- **先设计后实现**：新页面或大幅视觉修订必须先在 `ui-design/` 落地静态原型，并同步 `docs/ui-design/` 说明，再进入正式 `frontend/` 实施。
- **源级复刻，不做二次设计**：正式前端必须逐项复刻 `ui-design/` 中的 DOM 构图、布局、间距、字号、字体层级、控件密度、颜色、阴影、边框、圆角、状态、响应式行为和交互节奏；只允许为真实数据、路由、鉴权、可访问性和工程约束做必要适配，不允许重新设计、重新解释、重新组合视觉。
- **样式来源必须可追溯**：每个正式组件的样式、token、className 和布局规则必须能追溯到对应 `ui-design/src/*.jsx`、`ui-design/src/primitives.jsx`、`ui-design/src/app.jsx` 或 `docs/ui-design/`；不得凭 AI 判断补齐未在原型中出现的视觉值。
- **Parity gate 必须可执行**：涉及用户可见 UI 的 plan/checklist 必须包含对 `ui-design/` 的 100% 复刻验证，至少覆盖 DOM 锚点、关键 computed style、bounding box、viewport 布局和必要截图差异；“语义相似”“风格接近”“视觉大致一致”不能作为完成依据。
- **删除外部设计参考口径**：仓库不再保留独立外部品牌设计系统参考文件；后续若需要新的视觉方向，由用户先更新 `ui-design/` 原型，Agent 再做原生迁移。

---

## 2 核心原则

### 2.1 工程三原则（每个任务必须遵守）

**所有产出必须同时满足一致性、完备性和最佳实践。这不是建议，是强制检查项。**

#### 一致性（Consistency）

新代码必须与项目现有约定保持一致。动手前先观察，不要引入新风格。

- **命名**：读取同包已有代码，沿用相同的命名风格（变量、函数、文件）
- **结构**：新增文件时，先查看同目录已有文件的组织方式，保持一致
- **模式**：错误处理、日志格式、依赖注入方式等，必须复用项目已有模式
- **文档**：遵守 `docs/` 各子目录 README.md 与 `TEMPLATES.md` 定义的规则、模板和格式

#### 完备性（Completeness）

每个交付必须是完整的、可用的，不允许留下半成品。

- **功能**：实现必须覆盖正常路径和错误路径
- **测试**：新功能必须有对应测试，测试必须实际运行通过
- **文档**：接口变更必须更新相关文档和 INDEX.md
- **Checklist**：按 Checklist 工作时，每完成一项立即勾选，全部完成才算交付

#### 最佳实践（Best Practice）

遵循 Go 社区和本项目的工程最佳实践。

- **安全**：不引入 OWASP Top 10 漏洞，敏感数据不明文存储或日志输出
- **简洁**：最小化变更，不做未要求的重构或「顺便改进」
- **可测试**：新代码必须可被单元测试覆盖（接口抽象、依赖注入）
- **幂等**：配置变更和部署操作必须可重复执行

### 2.1.1 TDD / BDD 质量门禁（强制）

- **Code plan requires TDD**：凡 plan 涉及前端 / 后端 / 工具脚本 / 迁移 / codegen / 测试辅助等代码逻辑，必须通过 `/implement` 进入 `/tdd` 执行；checklist 中每个实现项必须有对应测试断言和实际运行证据。
- **Feature plan requires BDD**：凡 plan 引入用户可感知 UI、API 行为、业务流程或端到端功能，必须在同一 plan 目录内维护 `bdd-plan.md` 和 `bdd-checklist.md`，主 `checklist.md` 必须包含引用场景编号的 `BDD-Gate:` 项。
- **BDD 不适用时必须说明**：纯内部契约 / 工具 / 迁移 / codegen 若不产生用户行为流，可不创建 BDD 文件，但 plan 必须写明“不适用原因 + 替代验证 gate”（如 contract test、lint、drift check、migration check、smoke），并由 `/plan-review` 审查。

### 2.1.2 深度重校对门禁（强制）

当任务要求 review、reconcile、重新实施、忽略历史状态、校对 spec/plan/checklist 与代码事实，或产品 / UI spec 已大规模重构时，必须执行 deep reconcile，而不是轻量核对。

- **历史状态不算证据**：既有 `completed`、checklist 勾选、历史 PASS、历史测试结果和 diff 大小只能作为线索，不能作为当前完成依据。
- **Artifact-level 反查**：必须直接读取或解析当前真理源、实现代码、生成物、fixtures、baseline、DDL、runtime config、scripts、README、测试断言和 Make target；不得只读 plan/checklist 就进入下一项。
- **新版语义反向审计**：必须从当前 `docs/spec/product-scope/spec.md`、`docs/ui-design/`、`ui-design/` 和 active spec 中提取不变量，反向审查实现是否仍符合当前产品与交互范围。
- **旧口径负向搜索**：必须搜索旧 route、旧 tag/schema/table/event/job/config flag、旧 feature flag、旧 AI model/provider 假设、旧 `feature_key` / `featureKey` 路由口径，以及 Mistakes / Growth / Drill / 独立 Voice 等被当前设计丢弃的模块口径。
- **历史包删除零残留**：当用户要求删除已迁移文档包、旧目录或旧模块时，不得只改链接；必须删除实体目录 / 文件，并在 owner spec / plan gate 中固化目录名、文件名和旧 shorthand 的 zero-reference 搜索，确认当前 owner spec / coded truth source 可独立承接字段、事件、指标、日志、schema 与验证 gate。
- **旧 gate 只是必要条件**：如果现有 gate 只覆盖结构数量或历史断言，必须补充语义 lint、unit test、negative fixture、smoke 或脚本断言后再继续下一个 target。
- **反馈立即固化**：用户指出工作方式、审查深度或 gate 覆盖不足时，必须先把反馈写入当前执行规章、AGENTS.md、对应 skill 或 plan gate，再继续推进。

### 2.1.3 前后端契约执行门禁（强制）

凡任务涉及 `frontend/`、`backend/`、`openapi/`、`migrations/`、`config/ai-*`、`deploy/dev-stack/` 或 `test/scenarios/` 的代码实施、接口变更、mock 数据、场景验证或 L2 code review，Agent 必须在动手前读取并遵守以下当前契约：

1. `docs/development.md` §2 Frontend / Backend Contract Workflow
2. 相关模块 README：至少包括命中目录的 `README.md`，例如 `frontend/README.md`、`backend/README.md`、`openapi/README.md`、`deploy/dev-stack/README.md`、`test/scenarios/README.md`
3. 若涉及用户可见 UI：同时读取 `docs/ui-design/` 对应文档与 `ui-design/src/*.jsx` / `ui-design/src/primitives.jsx` / `ui-design/src/app.jsx` 的相关源码
4. 若涉及 API / fixture / generated client / handler：读取 `openapi/openapi.yaml`、相关 `openapi/fixtures/<tag>/<operationId>.json`、generated client/server artifacts，以及计划中的 operation matrix
5. 若涉及本地依赖或场景验证：区分 Docker Compose dev stack 与 Kind scenario target，按 `deploy/dev-stack/README.md` 和 `test/scenarios/README.md` 执行，不得凭历史印象假设环境入口

若计划或 checklist 缺少 operation matrix，或未标明 `operationId`、fixture、frontend consumer、backend handler、persistence、AI dependency、scenario coverage 的当前状态，必须先回到 `/plan-review --fix` 或请求用户批准修订，不得直接实施或宣称验证闭环。

### 2.2 任务开始前必须检查工作日志

**每次开启新任务前，必须：**

1. 读取 `docs/work-journal/INDEX.md` 查看最近进展
2. 读取最新日志文件了解详细内容
3. 确认待处理事项，决定从哪里继续

若仓库尚未初始化 `docs/work-journal/`，先完成文档骨架初始化，再继续执行任务。

### 2.3 故障排查时检查 Bug 知识库

**处理故障排查或修复任务时，还应：**

1. 读取 `docs/bugs/PATTERNS.md` 了解已知 Bug 模式
2. 修复 Bug 后，评估是否需要创建 Bug 记录（使用 `/bug-report`）

### 2.4 只记录真正完成的工作

**严禁虚假记录！**

- 代码必须已写入文件
- 功能必须可运行
- 测试必须已通过
- 未完成的工作记录为「进行中」

---

## 3 可用 Skills

### 3.1 Skills 使用协议与列表（强制）

**开始任何任务前，必须检查是否命中下表中的 Skill。命中则必须调用，不得绕过直接用基础工具完成。**

使用规则：

- 同一请求命中多个 Skill 时，选择能覆盖当前任务的最小集合，并按“问题入口/设计 → 文档 → 实施 → 收尾”的顺序执行
- 继续已有计划或恢复当前 plan 执行 → 必须调用 `/implement`
- 违反本节协议等同于违反 §4.3 禁止事项

使用自定义 Skills（尤其是 `/work-journal`）时，必须严格按步骤逐个 commit 执行工作流。禁止将多个日志条目批量合并到多个 commit 中。每个 commit 必须完成完整周期（分析 → 日志 → 提交）后，再进入下一个。

| Skill | 用途 | 调用策略 | 自动匹配条件 |
|-------|------|----------|-------------|
| `/change-intake` | 问题入口与 plan 自动发现 | **必须自动** | 用户报 bug / 回归 / 现象 / 特性修订且未明确给出 plan 时 |
| `/create-doc` | 创建或维护项目文档 | **必须自动** | 在 `docs/` 目录创建或修改文档时 |
| `/tdd` | TDD 开发流程 | **必须自动** | 按 checklist 实现代码时 |
| `/bug-report` | 创建 Bug 知识库记录 | **必须自动** | 修复 Bug 后建档时 |
| `/scenario-env` | 管理本地场景测试环境 | **必须自动** | 创建/验证/清理测试环境时 |
| `/sync-doc-index` | 检查/修复文档 Header 与 INDEX 一致性 | **必须自动** | 检查或修复文档 Header/INDEX 漂移时 |
| `/skill-creator` | 创建或更新 Skill | **必须自动** | 创建或更新 Skill 时 |
| `/implement` | 薄入口计划实施（解析上下文 + 编排执行 + 交接 `/tdd`） | **必须自动** | 用户要求实施、继续已有计划或恢复当前 plan 执行时 |
| `/plan-review` | L1 文档审查与文档修复 | **必须自动** | 用户要求审查或修复 spec/plan/checklist 一致性时 |
| `/plan-code-review` | L2 代码审查与代码修复 | **必须自动** | 用户要求审查或修复代码与 spec/plan/checklist 一致性时 |
| `/retrospective` | 成功交付后的会话复盘与改进建议沉淀 | **必须自动** | 功能或 bugfix 完成并通过验证后 |
| `/design` | 设计结晶：讨论 → 与需求匹配的 spec/plan/test/BDD 文档集 | **必须自动** | 设计讨论收敛，需落地为 spec/plan 文档时 |
| `/work-journal` | 记录工作日志并提交 | 显式调用 + `/tdd` phase-commit | 用户决定提交时机 / phase 边界自动提交时 |
| `/scenario-create` | 创建新的场景用例目录与索引项 | 显式调用 | 用户指定新增场景用例时 |
| `/scenario-run` | 执行场景集成测试 | 显式调用 | 用户指定运行测试时 |
| `/scenario-investigate` | 调查场景测试失败原因 | 显式调用 | 测试失败需诊断时 |
| `/scenario-redeploy` | 重建场景环境所需组件 | 显式调用 | 用户指定重新部署时 |
| `/init-docs` | 初始化 docs 目录结构 | 显式调用 | 新项目或首次添加文档时 |
| `/agent-browser` | 浏览器自动化交互 | 显式调用 | 网页测试、截图、表单填充时 |

---

## 4 AI Agent 行为规范

### 4.1 必须咨询用户的情况

1. **架构决策**：多组件交互方式变更
2. **接口变更**：影响外部调用方的 API 修改
3. **风险操作**：删除代码、重命名公开类型、修改数据结构

咨询时必须采用结构化方案格式：
- 列出至少两个可行方案（方案 A / B / ...），每个附简要理由与取舍
- 明确标注推荐方案及推荐原因
- 如无法推荐，说明需要用户补充的判断依据

### 4.2 可自行决策的情况

1. **实现细节**：不影响接口的内部重构
2. **测试补充**：增加测试用例
3. **文档完善**：补充注释或 README
4. **明显 bug**：行为明显不符合设计
5. **交付复盘**：功能或 bugfix 已完成且验证通过后的复盘建议

### 4.3 禁止事项

1. **禁止假设文件内容** — 必须实际读取
2. **禁止跳过测试运行** — 声称「测试通过」前必须实际执行
3. **禁止脱离计划开发** — TDD 必须按计划执行
4. **禁止延迟更新 Checklist** — 每完成一个小节立即更新
5. **禁止擅自修改计划** — 需修改时先与用户对齐
6. **禁止拆分同主题 sibling plan** — 命中 `completed` plan 时，不得新建同主题 sibling follow-up / bugfix plan；应在原 spec/plan/checklist 上原地修订
7. **禁止悬空原地修订** — 若当前只处于分析/建议阶段，不得先改写完成态 plan/checklist；若已开始修订，则不得在 owner handoff 前结束会话
8. **禁止浅层收口** — 不得把“小 diff”“历史 gate 通过”“历史 checklist 已完成”当成 spec/plan 与当前代码事实已经闭环的证明

### 4.4 必须事项

1. **必须复述理解** — 执行复杂任务前先确认
2. **必须报告异常** — 遇到预期外情况立即告知
3. **必须按序执行** — 按 Checklist 顺序，不得跳跃
4. **必须咨询用户修改计划** — 发现需修改时暂停报告
5. **必须设计先行** — 若入口判断为设计或特性变更，必须先修订 spec/plan，再编码
6. **必须收尾复查** — bugfix 或特性修订完成后，必须执行一次 post-pass doc reconcile（plan/spec/index/bug/retrospective）
7. **必须让原计划成为当前 owner，先更新 spec/plan/checklist，再继续 `/implement` 或其他明确 owner skill** — 命中已完成主题时，必要时先将 Header `状态` 调整回 `active`，完成验证后再恢复 `completed`
8. **必须使用中文与用户沟通** — 所有面向用户的聊天回复（含状态更新、问询、复述、总结）以及思考过程一律使用简体中文，确保交互语言一致；代码、标识符、命令、文件路径、技术术语、引用的英文原文以及代码内注释/文档遵循各自既有约定，不强制翻译
9. **必须主动给出下一步** — 每次完成一个复杂任务、阶段收口、提交或验证后，最终回复必须主动给出一个明确、可执行的下一步建议，包括建议 owner skill / plan、目标范围、为什么它是下一步；不要让用户反复追问“下一步是什么”。若存在多个合理路径，给出推荐路径和备选路径；若不能推进，明确 blocker 与解除条件。
10. **必须主动维护执行规章** — 当一次任务暴露出流程缺陷、审查盲点、误判模式或用户反复纠正的协作成本时，必须把规则沉淀到合适位置（AGENTS.md、skill、spec/plan gate、README 或报告台账），而不是只在当前对话中口头记住。
11. **必须执行前后端契约预读** — 命中 §2.1.3 范围的实施、验证或 review，必须在编码或下结论前读取 `docs/development.md` §2 与相关模块 README，并在结果中说明已遵守的关键契约或发现的缺口。

---

## 5 场景测试环境

**创建或操作测试环境前，必须先阅读：** `test/scenarios/README.md`

关键要点：
- 首次使用若存在镜像缓存脚本，应先运行 `image-cache.sh pull` 预热外部依赖镜像
- Kind 场景环境的部署、重建、验证入口以 `test/scenarios/README.md` 与对应层级 README 为准
- 禁止预设 Helm Chart、组件名、命名空间或外部依赖平台，必须以仓库当前测试框架文档为真理源

---

## 6 代码域所有权与协作边界

### 6.1 当前仓库目录职责映射

本节描述的是**当前 harness 仓库实际存在的协作边界**，不是历史产品代码域，也不是 plan 中的并行执行协议。

- easyinterview 当前仓库以 governance docs、skills、docs 流程和场景测试框架为主
- 当需要 delegate / teammate / 并行拆分时，以下 role ID 仅作为当前仓库的写入边界 shorthand
- 当前仓库尚不存在的目录，不得提前写入 ownership 映射

| 角色 ID | 角色 | 负责文件范围 | 说明 |
|---------|------|-------------|------|
| `governance` | 治理文档 | `AGENTS.md`、`CLAUDE.md`、`GEMINI.md` | 根级 agent 指令与全局协作约束 |
| `docs` | 项目文档 | `docs/` | spec、plan、report、work-journal、INDEX 等文档资产 |
| `skills` | Skills 与共享脚本 | `.agent-skills/` | skill 指令、模板、共享脚本与契约测试 |
| `scenarios` | 场景测试框架 | `test/scenarios/` | 场景 README、INDEX、脚本约定与环境说明 |

### 6.2 使用方式

- 顺序计划仍是默认执行模型；本节不改变 `/implement` → `/tdd` 的串行主路径
- 需要委派任务时，一个 worker 应尽量只拥有一个当前仓库边界的写入范围
- 若同一 checklist 小节横跨多个边界，应优先拆分任务或先更新计划，再进入实现
- role ID 可用于 review、讨论、任务拆分和脚本兼容，但不是新 plan 的强制元数据

### 6.3 协作约束

- 委派任务时，必须同时给出明确目标、交付结果与允许修改的路径范围
- 多 agent 协作完成后，必须由当前 owner 负责集成检查，并继续按 checklist 顺序推进

---

## 7 Git 分支策略

- 默认父分支: dev（优先自动探测；若未配置则使用当前主开发分支）
- `/implement` 自动从父分支创建 feature branch
- phase commit 后自动 merge 回父分支
