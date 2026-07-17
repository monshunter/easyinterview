# Harness 工程框架收敛

> **版本**: 1.8
> **状态**: active
> **更新日期**: 2026-07-18

## 1 目标与范围

本 Spec 是 easyinterview Harness 分层、Skill 分工、业务文档职责和上下文编排的唯一当前设计真理源。目标是在不降低工程正确性、可验证性和风险控制的前提下，以最少独立概念、最少重复 owner 和最短可靠主路径提升可信交付效率。迁移期间暂时保留现有 workflow Skill 与 docs 体系，用可验证、可回退的小步变更消除 Skill 之间的隐式耦合、Skill 与业务知识的耦合、重复规则和无差别预读；目标态不承诺永久保留缺少独立价值的旧资产。

本 Spec 采用以下工程定义作为最高层约束：

```text
Harness 工程框架 = Skill + Docs Arch + Env
```

- **Skill** 持有可执行 SOP，知道如何通过稳定的 Arch 接口完成一类工作；
- **Docs Arch** 持有项目的 owner、设计、交付、知识、导航和历史框架；
- **Env** 持有代码 gate、测试分层、真实 scenario、环境生命周期和当前运行证据的标准接口。

三者是共同运行的工程系统，不是由 caller 临时拼装的三个无关输入。项目之间复用的是骨架、角色、接口、模板参数和 SOP；变化的是 subject、业务合同、模块、数据、依赖、场景内容和环境适配器这些具体“血肉”。

Harness 新体系负责的不只是安装空目录或校验已有文件，而是**构建项目自身的 Harness**：先用内置 Blueprint 安装抽象层、模板参数与可执行接口，再从仓库事实、用户目标和项目 owner Spec 中提取项目参数，通过 design → implement/TDD → scenario/env → check 的同一 Harness 主路径自动生成或修订具体 docs、测试、场景与环境资产。项目血肉由 Harness 体系依据项目 Spec 构建，不要求人手工编写 adapter、场景脚本或目录内容；人只在需求、凭证、不可推导决策或高风险授权上提供输入。

本轮采用**方案 A：保留式渐进收敛**。方案 A 是迁移策略，不是永久目标架构：先证明替代路径不降低质量和恢复能力，再按退出条件删除重复入口。用户已经确认第 7 节的目标 Skill 名称、合并关系和退出项；该确认只固定目标设计，不构成当前会话的实施授权。重命名、合并、删除、plan/checklist 改写或代码变更必须在用户另行明确进入实施阶段后，按替代 owner、调用方迁移、验证和回退条件执行。

本 Spec 的范围包括：

- 根 `AGENTS.md`、独立编排文档、`.agent-skills/`、`docs/` 和各目录 `README.md` 的分层边界；
- `init-arch` 抽象 Blueprint、模板参数、项目事实发现和 Project Arch 自举构建循环；
- Skill 的业务解耦、语义化命名、合并/退出结论和独立调用合同；
- 环境构建与环境运行分离的专责 Skill 合同；
- 请求入口、设计、实施、审查、运行验证和收尾之间的编排 owner；
- `sync-doc-index` 从 Harness Skill 体系移除后的职责承接；
- 不同风险任务应承担的最小充分流程成本；
- 后续简化变更必须满足的兼容、验证和回退条件。

产品功能、OpenAPI 业务语义、正式前后端实现和真实 E2E 场景内容不在本轮设计变更范围内。

仓库根目录或外部传入的同名 `spec.md` 只作为讨论输入，不构成第二份当前设计真理源。讨论中确认的语义必须合并到本文件；未合并内容不自动生效。

### 1.1 目的函数

Harness 的北极星是在质量边界不退化的前提下，提高从用户意图到当前可信证据的端到端交付与恢复能力。不得把文件数量、Skill 调用次数、token 最少、文档完整率或 gate 数量单独当作成功目标。

方案、规则或资产的净价值按以下关系评估：

```text
Harness 净价值
  = 风险避免收益
  + 决策与恢复加速收益
  - 上下文读取成本
  - 文档维护与同步成本
  - 误阻塞成本
  - 漂移与错误路由成本
```

无法说明正向收益、实际成本和退出条件的规则，不得仅因历史存在而永久保留。

## 2 Harness 目标分层

### 2.1 组成模型

<!-- harness-table: harness-components/v3 -->
| 组成 | 回答的问题 | 共同骨架 | Harness 依据项目 Spec 构建的血肉 |
|------|------------|----------|-----------------------------------|
| Skill | 一类任务如何执行 | 输入/输出、SOP、权限、失败恢复、Arch 接口调用 | 当前目标、授权、业务 owner 与任务参数，以及据此生成的项目资产 |
| Docs Arch | 项目事实如何组织和演进 | 文档角色、owner 层次、生命周期、模板参数、导航和确定性投影 | subject、产品合同、模块说明、决策、Bug、报告和历史内容 |
| Env | 如何实施、测试、运行和证明 | gate 角色、测试分层、scenario 生命周期、环境状态、adapter 参数与证据格式 | 语言/框架命令、组件、端口、依赖、场景数据、脚本和部署拓扑 |

仓库编排是 Docs Arch 的标准组成，代码、测试、运行环境与 Git 证据是三者共同运行后的当前结果，不再作为需要 caller 另行拼装的第四套 Harness。

### 2.2 依赖方向

Project Arch 的最小可运行内核必须先于普通能力存在；完整项目实例随后由 Harness 使用自身能力自举构建，而不是由 caller 人工补齐：

```text
/init-arch 内置的 versioned Blueprint
  └─ 安装或升级 Project Arch 内核
       ├─ Docs Arch 角色、模板参数与 owner 生命周期
       ├─ test / scenario / env 标准接口与 adapter schema
       └─ discovery / check / upgrade / repair 工具
            └─ 从仓库事实、用户目标与既有文档形成项目 owner Spec
                 └─ 仓库编排选择 Arch-aware Skill 实施 Spec
                      ├─ delivery 能力生成项目 docs / code / tests / scenarios
                      └─ environment-build 设计并构建项目 env adapter
                           └─ environment-operate / scenario-run 运行真实 gate
                                └─ 将证据和差距反馈给 owner Spec
```

- `init-arch` 不要求 caller 先提供 framework、模板、schema 或 SOP；这些属于其内置 Blueprint；
- `init-arch` 负责发现可推导的项目参数；无法可靠推导的需求必须形成 `needs_decision` 或 owner Spec handoff，不得转化为“请人工手写 adapter”；
- Skill 可以且必须认识兼容版本内的标准 Arch 角色、路径和入口；
- Skill 不反向内置 EasyInterview 的 subject、route、组件、场景 ID 或环境实例；
- Skill 不调用或读取另一个 Skill 来决定全局流程；
- 编排文档不复制 Skill 的具体算法；
- Docs Arch 不复制通用 Skill 算法，但负责把项目事实、要求和决策绑定到标准 owner 角色；
- Env adapter 实现标准生命周期，由 `environment-build` 依据项目环境 owner Spec 设计和构建，再由 `environment-operate` 操作和验证；不要求其他 Skill 猜测项目命令，也不要求人手工落地；
- `AGENTS.md` 不维护第二份 Skill 编排表。

### 2.3 系统属性

Harness 是由用户、Agent、Skill、文档、代码、测试、运行环境和 Git 共同组成的开放系统。设计与验收必须同时覆盖以下属性：

| 系统维度 | Harness 约束 |
|----------|--------------|
| 整体性、目的性 | 优化端到端可信交付与恢复，不以单个 Skill、局部 gate 或 token 指标替代系统结果 |
| 稳定性 | 以当前测试、契约、安全、回滚和失败恢复建立不退化边界 |
| 层次性 | 公共政策、仓库编排、业务 owner 和可执行证据分层，依赖方向保持单向 |
| 相似性 | R0-R3 使用同一判断骨架，但按风险采用不同流程强度，不强制相同步骤 |
| 突发性 | 验证 Skill 组合、规则级联和文档投影产生的误阻塞、上下文放大与虚假高置信路由 |
| 自组织性 | 优先从当前仓库事实生成索引、路由和导航，让局部 owner 提供事实并通过反馈修正最小规则 owner |
| 开放性 | Skill 使用显式输入、输出、权限和失败合同，不绑定特定业务、模型、工具或运行环境 |

### 2.4 Project Arch v1

Project Arch 是由 `init-arch` 内置、安装和升级的版本化工程骨架合同。它定义稳定角色、固定入口、模板参数、生命周期、扩展点和确定性 gate，不硬编码任何产品业务实例。业务实例不是外部手工输入，而是 Harness 读取仓库事实与项目 Spec 后生成、实施并验证的 Project Arch 实例。安装版本由 `docs/README.md` 中唯一的 `<!-- project-arch: v1 -->` 标记持有，不再增加一份需要人工同步的逐任务 manifest。

Project Arch v1 的核心安装面为：

```text
AGENTS.md
docs/
├── README.md
├── agent-workflow.md
├── development.md
├── spec/{README.md,TEMPLATES.md,INDEX.md}
├── apis/{README.md,TEMPLATES.md,INDEX.md}
├── bugs/{README.md,TEMPLATES.md,INDEX.md,PATTERNS.md}
├── discuss/{README.md,TEMPLATES.md,INDEX.md}
├── reports/{README.md,TEMPLATES.md,INDEX.md}
└── work-journal/{README.md,TEMPLATES.md,INDEX.md}
test/
├── README.md
└── scenarios/
    ├── README.md
    ├── _shared/README.md
    ├── e2e/{README.md,INDEX.md}
    ├── env-setup.sh
    ├── env-status.sh
    ├── env-verify.sh
    ├── env-redeploy.sh
    └── env-cleanup.sh
scripts/
└── harness_arch.py
```

<!-- harness-table: project-arch-interfaces/v1 -->
| Interface | Canonical entry | Skill usage |
|-----------|-----------------|-------------|
| `arch.root` | `docs/README.md` | 读取安装版本、角色导航和 extensions |
| `arch.workflow` | `docs/agent-workflow.md` | 编排层选择能力、handoff、确认与退出；Skill 不复制 |
| `arch.development` | `docs/development.md` | 获取项目 code/build/test adapter 与跨层合同 |
| `arch.spec` | `docs/spec/README.md`、`TEMPLATES.md`、`INDEX.md` | spec/delivery 能力获取 owner、生命周期和事务入口；确定性文档写入由 Project Arch tooling 执行 |
| `arch.test` | `test/README.md` | TDD/review 获取测试分层、聚合 gate 和当前证据边界 |
| `arch.scenario` | `test/scenarios/README.md`、`e2e/INDEX.md` | scenario 能力获取真实 E2E 结构、选择和结果合同 |
| `arch.env` | `test/scenarios/env-{setup,status,verify,redeploy,cleanup}.sh` | env/scenario 能力执行统一生命周期；脚本内部绑定项目组件 |
| `arch.check` | `python3 scripts/harness_arch.py check` | 所有能力 preflight 版本/结构兼容；`init-arch` 还拥有 init/upgrade/repair |

项目可以增加 `docs/ui-design/`、OpenAPI、模块 README、更多测试层或部署适配器，但不得改变核心角色含义。共同 SOP 与项目血肉按下表分开：

<!-- harness-table: project-arch-subsystems/v2 -->
| 子系统 | Project Arch 固定合同与参数 | 项目 owner Spec | Harness 实施产物 |
|--------|-----------------------------|-----------------|------------------|
| docs | owner 类型、Spec/plan/current/history 生命周期、模板、导航、文档事务和 docs gate | subject、业务规则、设计内容、决策与历史边界 | 具体 Spec/plan、Bug、报告、README 与确定性导航 |
| test | unit/contract/integration/real-system 证据分层、主证据 owner、聚合回归入口 | 语言框架、风险、测试边界、fixture 与 gate 要求 | 测试实现、fixture、聚合命令和当前结果 |
| scenario | suite/index/case 结构、setup/trigger/verify/cleanup 生命周期、结果状态和脱敏证据 | 用户流程、真实系统边界、数据隔离和可观察结果 | 场景 ID、数据、脚本、浏览器/API 操作和证据产物 |
| env | setup/status/verify/redeploy/cleanup 接口、adapter 参数、readiness、污染控制、日志和恢复语义 | 服务、端口、拓扑、依赖、secret 来源、健康与恢复合同 | 环境配置、启动/检查/重建/清理脚本、Make 入口和运行证据 |

环境脚本在尚未绑定项目实现时必须返回明确的 `NOT_CONFIGURED`/`needs_spec`/`needs_decision` 等价结果，不得用空脚本或 mock PASS 伪装成可运行环境。但这些状态只是 Harness 自举循环的中间结果：编排层必须把缺口交给 design 能力形成或修订环境 owner Spec，再由 implement/TDD 与 scenario/env 能力生成并验证 adapter。只有缺少用户决策、外部凭证、权限或不可获取基础设施时才允许停在 `needs_input`；不得把实现工作转交给人工。

### 2.5 Project Arch 自举构建循环

Project Arch 采用“框架内核 + 项目 Spec + Harness 实施”的自举模型：

1. `init-arch` 安装抽象角色、模板参数、adapter schema 和确定性检查器；
2. discovery 从源码、构建清单、现有 README、测试、部署文件和用户目标提取可证明的项目事实；
3. 缺少普通业务合同时，由编排层调用 `spec-design` 创建或修订唯一 owner Spec；环境合同缺失时，由 `environment-build` 负责发现约束并创建或修订环境 owner Spec，而不是把参数表交给人手工填写；
4. `delivery-execute` 依据 Spec 生成代码、测试、配置和 scenario；`environment-build` 持有环境领域判断并构建具体配置、依赖、脚本和 adapter，通用代码修改可由编排层以显式 handoff 交给 `delivery-execute`；
5. `environment-operate`、`scenario-run` 与 Arch check 在真实项目中执行 setup/status/verify/redeploy/cleanup，失败证据回到对应 owner；
6. Spec、实现和 gate 收敛后，Project Arch 才从 `NOT_CONFIGURED` 进入 `compatible/ready`。

EasyInterview 的 [`local-dev-stack` Spec](../local-dev-stack/spec.md) 是本项目当前环境 owner：它定义 EasyInterview 自己的依赖、端口、拓扑、健康、幂等、污染与恢复合同。它不是 Harness Blueprint、golden fixture、跨项目模板或环境构建方法的 owner；目标 Harness 只把它作为 EasyInterview upgrade/regression 的真实项目输入，由 `environment-build` 读取并维护该项目的环境实现。Harness 框架级验证必须使用专门维护的最小合成项目 fixture，至少覆盖两种异构环境形态，不能从 `local-dev-stack` 反向抽取固定组件清单。

## 3 第一性原理

### 3.1 单一事实 owner

同一语义事实只能有一个当前维护 owner。该 owner 可以由 Harness 创建或修订，但其他位置只能引用、投影或执行，不得复制后形成平行真理源。

- 当前产品或工程合同由 subject Spec 持有；
- 当前 UI 合同由 `docs/ui-design/` 持有；
- wire API 合同由 `openapi/openapi.yaml` 持有；
- 单次有限交付的 delta 与顺序由 plan/checklist 持有；
- 仓库 Skill 编排由 `docs/agent-workflow.md` 持有；
- 目录和模块的具体使用合同由命中路径最近的 `README.md` 持有；
- 可执行结果由代码、测试、真实环境和 Git 持有；
- 历史只由 work journal、Bug、report、migration、Git 等历史 owner 持有。

### 3.2 Arch-aware，而非业务实例内置

Skill 必须直接消费兼容版本的 Project Arch，不要求 caller 在每次调用时重新注入 framework、schema、模板、目录角色、scenario 生命周期或环境 SOP。允许且要求 Skill 固定认识：

- Project Arch 版本和标准角色；
- 核心 Docs Arch 路径、文档事务与 owner 生命周期；
- test/scenario/env 的标准接口、结果状态和恢复语义；
- Project Arch 声明的 deterministic gate 与 extension 解析方式。

Skill 通过当前 owner 文档和相关 README 获得本次任务的业务实例，并可以依据这些合同生成或修改项目资产。它不得硬编码 EasyInterview 的 subject、页面、route、operationId、表、事件、具体 E2E ID、组件、端口、provider、secret、部署拓扑或项目专用命令。标准 Arch 入口不是业务耦合；把项目血肉复制进 Skill 才是业务耦合。

### 3.3 证据靠近执行 owner

测试 PASS、E2E 结果、生成物一致性、数据库状态和运行日志应保留在最接近其生成位置的 owner。文档可以引用当前证据，但不得把历史 PASS 或 checkbox 当作当前实现正确性的替代品。

### 3.4 流程强度与风险相称

流程按变更的可逆性、用户影响、跨层范围和失败代价分级。低风险任务不得无差别承担高风险任务的全部文档和预读成本；高风险任务不得借“简化”绕过设计、确认、验证、安全或回退门禁。

### 3.5 派生信息优先由仓库工具投影

凡能从 Header、链接、源码、OpenAPI、测试资产、Git 或目录结构稳定推导的信息，应由仓库脚本、CI 或文档写入事务投影。此类能力属于仓库 docs/tooling 合同，不需要包装成 Harness Skill。

目标态不保留 checked-in、人工维护的 `context.yaml` 或等价 discovery manifest。精确 owner 路由、文档关联和候选解释必须由 Project Arch tooling 从路径、Header、Markdown 链接、代码/API/route 标识、Git 和当前任务输入按需生成；生成结果是可删除、可重建、与当前 commit/dirty state 绑定的缓存，不是 Spec、plan 或实现事实源。迁移期现有 `context.yaml` 只作为旧入口兼容与回放输入，替代路由和恢复能力通过验收后必须退出。

### 3.6 当前合同与历史分离

Spec 描述现在应当是什么；plan/checklist 描述一次有限交付正在做什么；work journal、Bug、report 和 Git 描述发生过什么。三类信息不得混写为一个永久增长的工作包。

### 3.7 奥卡姆约束

只有在移除某个资产会丢失独立语义、必要控制或不可重建证据时，才保留它。简化的目标不是最少文件，而是最少独立概念、最少重复 owner 和最短可靠主路径。

每个拟保留的资产、规则或 Skill 必须通过以下判定：

1. 是否承载不可替代的独立语义；
2. 是否拥有与其他资产不同的独立生命周期；
3. 是否保存无法从代码、Git 或其他 owner 重建的证据；
4. 是否以更低总成本防止一个明确、真实且仍适用的失败。

若以上条件均不成立，默认删除、合并、动态生成或按需加载。每条新增或保留的规则必须说明防止的失败、适用风险级别、当前有效证据、执行与维护成本，以及删除或降级条件。

### 3.8 五事态势判断

任务进入流程前应按五个维度建立最小充分态势：

- **道**：用户目的、成功标准和唯一 owner 是否一致；
- **天**：当前时间、分支、工作树、外部依赖和事实时效是否允许执行；
- **地**：代码拓扑、依赖边界、数据/接口路径和影响面是什么；
- **将**：当前 Agent、Skill、工具、权限和协作者是否具备完成能力；
- **法**：适用公共政策、仓库 gate、失败恢复和退出条件是否明确。

五事是运行时判断，不得再落成每个任务强制维护的表单或 manifest。R0/R1 可以隐式完成；R2/R3 或存在多个候选方案时必须显式报告关键差距。需要比较方案时，至少比较目标一致性、owner 置信度、时机与拓扑适配、能力成熟度、gate 可执行性、历史/回放证据、总成本与退出条件；通用词重叠或流程资产数量不得替代比较结论。

### 3.9 度量比较链

重大 Harness 决策必须遵循“地 → 度 → 量 → 数 → 称 → 胜”的证据链：

1. 从真实仓库拓扑、owner 和依赖关系出发；
2. 界定变更边界、影响面和故障半径；
3. 估算上下文、协作、迁移、验证和恢复负担；
4. 采集读取量、首次有效证据时间、工具调用、流程文件触碰、误路由、误阻塞、缺陷逃逸和恢复结果；
5. 在同一风险边界下比较候选路径；
6. 以当前可执行证据证明结论，而不是以结构 PASS、历史完成状态或主观简洁感宣告成功。

不得从“更轻”“更规范”或“历史如此”的偏好直接跳到保留、删除或全量迁移。

### 3.10 谋攻、成本与不败后胜

Harness 优先消除失败产生条件：修正唯一 owner、生成可推导投影、复用靠近执行 owner 的 gate，并在最小责任边界修复根因；增加外围包装、重复文档或事后报告只在它们具有独立价值时采用。

任何规则、预读、校验和交接都会消耗时间、token、维护与协作成本。流程必须设置与风险相称的预算和止损点；在已有证据足以支持决策时停止扩展上下文，在候选路径无法满足质量底线或恢复条件时停止迁移。

迁移必须先建立“不败”条件，再追求效率收益：冻结不可降低的质量不变量，采集旧路径基线，准备回退，使用有限试点和代表性任务回放证明新路径不退化，最后按退出条件删除旧入口。不得用永久双轨换取表面稳定，也不得在替代路径未经证明前先删除当前能力。

### 3.11 风险证据经济

Harness 默认把 AI 编码执行者视为具备高级资深工程师水平。这会降低简单逻辑的出错概率，但不能降低安全、数据损坏、不可逆操作和外部合同失败的影响，因此证据强度必须由失败风险决定，而不是由“AI 很强”或“每个实现项都应有一个新测试”决定。

只有同时满足以下条件的代码路径，才能判定为**普通风险**：

- 逻辑局部、确定、可直接阅读，分支少且不存在组合型边界；
- 不解析不可信外部输入，不实现外部协议、公共格式或跨层合同；
- 不涉及持久化、不可逆副作用、安全、隐私、并发、幂等、重试或故障恢复；
- 失败容易观察、定位、回退，不会阻断关键用户主链；
- 没有对应的历史缺陷模式或已知高逃逸风险。

低频路径不自动等于普通风险。删除、补偿、恢复、权限拒绝、迁移回滚等路径即使调用频率低，也要按失败影响判定为重要或关键风险。

证据购买遵循以下最小充分规则：

| 代码证据风险 | 默认验证 | 何时增加证据 |
|--------------|----------|--------------|
| 普通 | 默认不新增专用单元测试；使用编译、类型检查、lint、现有邻近测试或仓库回归证明未破坏基线 | 只有现有证据不能观察结果时，补一个最小 smoke 或消费者断言 |
| 重要 | 由一个最接近行为 owner 的 focused test 覆盖主要成功或失败分支 | 只有另一层能证明不同故障模式时才增加 contract 或 integration 证据 |
| 关键 | focused owner test 加必要的合同、集成或真实 API/UI 证据 | 安全、隐私、数据一致性、不可逆操作、并发、恢复或外部协议分别按独立风险补证据 |

每个重要行为或重大风险必须有一个**主证据 owner**。同一默认值、边界值或业务结论不得在 loader、composition、domain、frontend 和 scenario 中逐层复制；额外测试必须证明不同故障模式。不得重复测试标准库、框架或已采用依赖的内部行为，也不得为编译器或类型系统已经保证的 getter、setter、简单透传和私有实现细节新增测试。真实缺陷回归、业务状态转换、安全与隐私、持久化原子性、并发/幂等/重试/恢复、外部协议适配、项目错误映射和跨层合同仍是优先证据对象。

TDD 是缩短反馈回路和验证设计假设的方法，不是测试资产配额。Red 可以复用或调整现有测试，也可以从最接近风险 owner 的 contract/integration test 开始；一个测试可以覆盖多个实现项，不要求每个 checklist 项都新增一个测试文件或测试函数。测试代码量、测试/业务代码比和测试文件数只能作为重复与维护成本的趋势信号，不得成为硬性完成门禁。

### 3.12 能力寻源优先

能力寻源只在本 Spec 保留以下 Harness 不变量。分类触发器、寻源顺序、依赖准入、项目自有边界和退出规则的目标唯一执行 owner 是 [docs/development.md](../../development.md)；当前文件尚未包含该执行章节，因此它属于后续迁移产物，不得把计划中的 heading 当成已完成合同。

- 简单、边界清晰或业务内部独有的逻辑允许直接自研，并将寻源明确记为不适用；不得强制外部检索或新建调研文档。
- 通用能力必须先形成与证据风险相称的 build-vs-adopt 结论，再授权实现；采用依赖不等于放弃项目边界，自研也不免除完整所有权与退出条件。
- 判断通过现有 Design Brief、plan 风险/验证策略、Change 或 Decision 传递，不新增强制 manifest、候选配额或平行研究包。
- 标准 handoff 只携带能力分类、寻源深度、结论、摘要、引用以及条件性的边界/退出条件；语义完整性由 design 与 review 能力依据唯一执行 owner 核验。

## 4 当前体系的真实意图与实际效果

### 4.1 应继续保留的真实意图

现有体系的核心意图是正确的：

1. 用 owner discovery 避免脱离当前 Spec 或 plan 工作；
2. 用 design-first 防止把产品或架构决策偷渡为实现细节；
3. 用实施与 TDD 能力维持顺序实施、当前测试证据和 checklist 同步；
4. 用 L1/L2 review、场景验证和契约预读覆盖文档、代码与运行时漂移；
5. 用 Bug、retrospective、work journal 和 Git 保留可复用知识与交付审计；
6. 用 Header、INDEX 和 `context.yaml` 降低跨会话恢复与路由歧义。

### 4.2 需要收敛的实际效果

当前主要成本来自跨层职责混合：

- 同一规则在 `AGENTS.md`、多个 Skill、README、plan 和 checklist 中重复出现；
- 多个 Skill 直接写入 EasyInterview 路径、模块、Make target 和场景约束，难以独立复用；
- Skill 内嵌下一 Skill 的调用关系，形成隐式工作流和连锁预读；
- owner 已明确时仍可能经过重复 locate、预读和上下文验证；
- plan、checklist、BDD、context、INDEX、report 和 journal 可能重复投影同一状态；
- checked-in `context.yaml` 以人工 discovery 重复投影代码与文档事实，结构校验通过仍可能包含失效路径；
- 历史状态容易与当前合同混合，增加“历史 PASS 等于当前完成”的误判风险；
- 低风险局部工作也可能被迫走完整高风险链路。

当前静态扫描已发现 20 个 Skill 直接包含仓库路径、模块名或仓库命令，Skill 之间还存在显式交叉调用；这证明业务解耦与能力合并必须成为独立验收面，而不能只调整 Skill 名称、移动重复文字或把调用改写成 handoff。

## 5 风险分级与最小流程

| 等级 | 判定 | 最小流程 | 默认不要求 |
|------|------|----------|------------|
| R0 | 只读解释、检索、现状核对 | 读取用户指定事实与最小 owner，给出证据化结论 | 分支、文档、日志、报告、实施 Skill |
| R1 | 局部、可逆、不改变对外合同 | 确认 owner，focused test 或相称校验，必要回归 | 修改 Spec、创建新 plan/BDD/Bug/report |
| R2 | 用户可感知行为或跨层交付 | 更新当前 Spec，使用有限 plan/checklist，代码逻辑 TDD，按行为需要 BDD/真实 E2E | 与行为无关的重复 BDD/E2E |
| R3 | 公开 API、迁移、安全隐私、删除、生产操作或跨组件架构 | R2 全部要求，加人类确认、回滚、幂等、安全与可观测性设计；必要时 Decision | 未确认的一次性破坏性迁移 |

任务范围变化时必须重新分级；不确定时按较高风险处理。具体“风险 → Skill 顺序 → 必读 docs/README → 退出 gate”的映射只在 `docs/agent-workflow.md` 中维护。

风险分级必须基于五事态势，而不是仅按文件数量或用户使用的动词判断。R2/R3 的持久化交付记录应说明关键态势、比较依据和不败条件；R0/R1 不因此新增文档资产。

R0-R3 描述任务需要承担的治理与交付流程；第 3.11 节的普通、重要、关键描述具体代码行为需要购买的证据强度。二者相关但不互相替代：R2 任务可以包含无需新增单测的普通逻辑，R1 局部修复也可能命中安全、删除或恢复等关键证据风险。能力属性构成第三个独立输入；其具体风险组合矩阵只在 `docs/development.md` 的能力寻源合同维护。Harness 不为这些轴新增永久 ID、表单或 manifest。

## 6 Project Arch 与独立编排合同

### 6.1 `init-arch` bootstrap 与 upgrade

`init-docs` 的目标能力名称调整为 `init-arch`。`init-arch` 必须携带 canonical Project Arch Blueprint 和模板资产，不得要求一个尚未初始化的项目先从外部提供 framework manifest、目录 schema、模板、验证规则或 upgrade SOP。

`init-arch` 提供四种可执行模式：

| 模式 | 行为 | 成功结果 |
|------|------|----------|
| `init` | 在 fresh 或 legacy repository 中盘点并安装缺失的 Docs Arch 与 Env 接口，发现项目参数与 owner 缺口 | core skeleton、Project Arch version、parameter evidence、spec/adapter handoff、current gate evidence |
| `check` | 只读验证版本、必需角色、链接、模板、test/scenario/env 接口和 deterministic gates | compatible / drift / not-configured / conflict 清单 |
| `upgrade` | 预览并执行 N→N+1 的 Arch-owned 迁移 | 保留项目血肉的升级结果、迁移证据和 rollback checkpoint |
| `repair` | 只修复已确认的 Arch-owned drift | 最小修复、验证结果和未解决的项目扩展冲突 |

调用输入只有目标 repository、模式、用户目标和授权；现有 build/test 命令、服务启动方式、health endpoint、部署清单和测试资产应由 discovery 从仓库事实读取，不要求 caller 重复传入。无法推导的业务选择、外部凭证或高风险取舍返回结构化 `needs_decision`，项目实现缺口返回 owner Spec handoff；不得把项目 adapter 当成必须由 caller 手工提供的 framework 参数。

执行顺序固定为：识别版本与安装态 → inventory `absent/compatible/custom/conflicting` → 发现项目事实与参数证据 → 生成变更预览 → 检查授权和 checkpoint → 原子写入 Arch-owned 内核 → 运行 docs/test/scenario/env 合同 gate → 输出 `ready/spec_required/decision_required/conflict` handoff → 失败回滚或给出精确 resume condition。同版本第二次执行必须无 diff；不得覆盖项目 subject、业务文档、场景内容、环境 secret、实现代码或已有人类维护的扩展。

迁移期间 `/init-docs` 只可作为指向同一能力的临时 alias，必须声明退出条件；目标态只保留 `/init-arch`。

### 6.2 项目血肉自动构建合同

Project Arch 内核安装完成不等于项目 Harness 建设完成。若 docs、test、scenario 或 env 中存在 `spec_required` / `NOT_CONFIGURED` 缺口，`docs/agent-workflow.md` 必须继续编排以下闭环：

1. 普通项目缺口由编排层把 discovery 证据和未决参数交给 `spec-design`；环境缺口交给 `environment-build`，由它发现环境约束并建立或修订唯一环境 owner Spec；
2. 通过 `spec-review` 确认 Spec 能解释目标、约束、失败、恢复与验收，而不是只填一张参数表；
3. 通过 `delivery-execute` 生成具体代码、文档、测试、配置和 scenario；环境相关资产仍由 `environment-build` 持有领域完成定义和验收责任；
4. 通过 `environment-operate`、`scenario-run` 与 Arch check 在真实项目中验证，失败时回到最小 owner 修订；
5. 只有全部必需接口有真实实现和当前证据时，才报告 Project Arch ready。

Harness 可以自动完成事实提取、文档生成、代码实施、环境构建、环境操作和验证。人工参与只限于明确项目目标、选择不可推导方案、提供无法安全获取的凭证，以及批准 R3/破坏性操作；“请人工创建脚本/补目录/填写完整配置”不是成功 handoff。

### 6.3 环境构建与环境运行分工

环境工程必须拆成两个语义不同、可独立选择的能力：

| Skill | 负责 | 不负责 | 完成结果 |
|-------|------|--------|----------|
| `/environment-build` | 读取仓库事实和项目要求，选择适配拓扑，创建或修订唯一环境 owner Spec，构建/升级环境配置、依赖、脚本、adapter 与验证合同，并持有环境领域验收 | 把 `local-dev-stack` 或任一项目栈当作模板；要求人工补脚本；日常重复启动已有环境 | environment Spec、实施资产、生命周期入口、contract evidence、未决 decision/credential handoff |
| `/environment-operate` | 对已存在且声明兼容的 env adapter 执行 setup/status/verify/redeploy/cleanup，采集日志、readiness、污染和恢复证据 | 决定新环境拓扑、替代缺失 Spec、生成第二套环境合同或修改产品业务 | 当前环境状态、操作证据、失败位置、cleanup/rollback/resume condition |

`environment-build` 是环境建设的领域 owner，不是只写一份 Spec 后把实际搭建留给人。若搭建涉及通用代码修改，`docs/agent-workflow.md` 可以依次调用 `delivery-execute` 执行显式 handoff，但环境约束、资产完整性和最终环境验收仍回到 `environment-build`；两个 Skill 不互相调用。

Project Arch 只提供 lifecycle interface、adapter schema、证据格式和最小测试 fixture。真实项目必须依据自己的环境 owner Spec 形成具体血肉。`local-dev-stack` 只属于 EasyInterview，不参与定义通用环境构建算法；Harness 框架测试应维护独立的合成 fixture repository，并覆盖至少一种宿主机进程型环境和一种容器化环境，证明相同骨架可以承载不同血肉。

### 6.4 唯一编排 owner

仓库应建立 `docs/agent-workflow.md`，作为唯一 Skill 编排真理源，至少定义：

- 请求分类和 owner 已知/未知分支；
- R0-R3 风险到 Skill 序列的映射；
- 每个阶段需要加载哪类 owner doc 和相关 README；
- Skill 之间的显式输入/输出 handoff；
- review 是否修复、场景是否运行、何时收尾等条件分支；
- 失败、暂停、恢复和用户确认点；
- `work-journal` manual/auto 模式在交付边界的调用条件。

`AGENTS.md` 只保留无法下沉的全仓安全、Git、用户确认和证据底线，并以一个明确入口引用 `docs/agent-workflow.md`。不得在 `AGENTS.md` 再维护完整 Skill 表、自动触发矩阵或逐步 runbook。

### 6.5 Skill 的虚实结合合同

一个可用 Skill 必须同时是指导思想和可落地执行框架。每个目标 Skill 至少包含以下五层，不得只保留抽象输入输出表，也不得退化为无判断的命令清单：

1. **Purpose and judgment**：解释它解决什么问题、关键判断依据和不负责什么；
2. **Arch contract**：声明兼容的 Project Arch 版本、使用的标准角色、入口和 extension；
3. **Executable SOP**：给出可顺序执行的 preflight、主路径、分支、实际操作和停止条件；
4. **Evidence and result**：说明要运行或采集的 Arch gate、当前证据以及结构化输出；
5. **Failure and recovery**：定义 checkpoint、cleanup、rollback、resume condition 和 fail-closed 边界。

这些层使用统一的 `## Purpose and Judgment`、`## Arch Contract`、`## Executable SOP`、`## Evidence and Result`、`## Failure and Recovery` 标题；`Arch Contract` 中必须包含唯一的 `<!-- project-arch-compat: v1 -->` 标记和本 Skill 实际消费的 `arch.*` interface。结构化 capability table 可以保留为机器合同，但不能替代五层正文。

Skill 的显式任务输入只携带 objective、owner 实例、scope、authorization、当前业务参数和必要证据，不重复传入整个 Arch 定义。若 Project Arch 未安装或版本不兼容，Skill 必须停止并返回 `init-arch init/upgrade` 建议，不得临时发明另一套框架。

### 6.6 Skill 间解耦

1. 每个 Skill 必须能在兼容 Project Arch 上用显式任务输入独立执行，不要求调用者先运行另一个 Skill 才能理解其合同。
2. Skill 不得直接调用另一个 Skill；需要连续能力时，由编排文档规定 Agent 分别调用并传递 handoff。
3. Skill 只返回自身能力范围内的结果、证据、失败原因和建议 next capability，不替编排层决定整个任务生命周期。
4. Skill 可以复用通用共享库，但共享库不得成为隐藏业务规则或隐藏编排层。
5. 兼容入口可以暂时存在，但必须标明唯一能力 owner 和移除条件。

### 6.7 业务上下文加载

编排层应根据任务命中范围选择具体文档：

1. 当前 subject 的 Spec 与适用的有限 plan/checklist；迁移期可读取旧 `context.yaml`，目标态使用按需生成的 owner index；
2. `docs/development.md`、`docs/ui-design/`、`openapi/` 等跨层 owner；
3. 命中代码、配置、迁移、部署或场景目录最近的 `README.md`；
4. 只有出现具体未决问题时才扩展其他引用。

Skill 先按 Arch contract 加载固定角色，再读取编排选定的 owner 和相关 README。Skill 可以枚举 Project Arch 的 canonical 路径和入口，但不得枚举 EasyInterview 的业务目录、业务标识或项目专用命令。

### 6.8 证据与能力决策编排

编排层在进入设计或实现前，只需做两个运行时判断：

1. 依据第 3.11 节判定代码证据风险和主证据 owner，复用已有证据并只补缺口；
2. 依据第 3.12 节判定能力是业务特有还是通用，并选择相称的寻源深度。

这两个判断通过现有 Design Brief、plan 风险/验证策略、Change 或 Decision 传递，不新增强制 manifest、测试配额表或生态调研文档。若实现过程中发现边界、安全、外部兼容或失败影响高于预期，必须重新分类并提高证据或寻源强度；反之，已有证据已覆盖风险时应停止继续堆叠测试和依赖比较。

## 7 Skill 目标能力与命名合同

当前仓库有 20 个顶层 Skill。用户已经确认方案 A 的领域化语义命名、合并关系和明确退出项；目标态固定为 14 个 Project Arch-aware Skill，而不是继续把现有名称当作默认保留候选。具体顺序由 `docs/agent-workflow.md` 持有；后续新增能力仍必须通过以下准入判定：

1. 是否存在无法由确定性工具完成的独立判断；
2. 是否具有与其他能力不同的输入、输出和失败恢复生命周期；
3. 是否需要作为可被用户或编排层独立选择的任务入口；
4. 是否不仅是另一个 Skill 的子步骤、参数模式或收尾动作；
5. 合并后是否会导致权限、风险或证据 owner 失真。

不满足前三项，或属于第四项且不存在第五项风险时，默认并入相邻 Skill、Project Arch tooling 或普通脚本。Skill-to-Skill 调用改成 handoff 只解决耦合，不自动证明两个 Skill 都应保留。

名称统一采用“领域 + 明确动作”，描述用户或编排层可观察的能力结果；不得用 `design`、`implement`、`review` 这类无领域限定的宽泛词，也不得用 `tdd` 这类内部方法充当顶层能力名。除 `/init-docs` 为 legacy upgrade 保留的限时 alias 外，重命名采用 hard cut，不长期保留旧名称兼容入口。

<!-- harness-table: target-skill-matrix/v3 -->
| 目标 Skill | 当前来源 | 唯一职责 | 标准输出 |
|------------|----------|----------|----------|
| `/init-arch` | rename `/init-docs` | 初始化、检查、升级和修复 Project Arch 内核，并发现项目 Spec/adapter 缺口 | Arch 版本、安装/升级结果、事实证据、handoff、四子系统 gate 与恢复状态 |
| `/work-triage` | rename `/change-intake` | 定位唯一 owner、分类问题与风险并选择下一项独立能力 | owner 候选、置信度、命中证据、风险与 capability 建议 |
| `/spec-design` | rename `/design` | 把确认后的需求和取舍结晶为最小充分 Spec/plan/test contract | Design Brief、owner 文档、覆盖矩阵、风险与 decision handoff |
| `/delivery-execute` | merge `/implement` + `/tdd` | 校验交付输入并按 Project Arch test contract 完成实现、测试和进度更新；TDD 是内部 SOP | 实现、最小充分测试证据、checklist/phase 状态与失败恢复 |
| `/spec-review` | rename `/plan-review` | 审查并在获准时修复 Spec、plan 和验证合同的当前自洽性 | L1 findings 或文档修订与 gate 证据 |
| `/delivery-review` | rename `/plan-code-review` | 审查实现、生成物和当前证据是否符合 owner Spec | L2 findings、证据/依赖成本判断与 remediation request |
| `/environment-build` | new | 依据项目要求和仓库事实设计、构建、升级并验收项目专有环境 | 环境 Spec、实现资产、lifecycle adapter、contract evidence 与 handoff |
| `/environment-operate` | merge `/scenario-env` + `/scenario-redeploy` | 操作已建成环境的 setup/status/verify/redeploy/cleanup 生命周期 | 环境状态、日志、readiness、污染、恢复与 cleanup 证据 |
| `/scenario-author` | rename `/scenario-create` | 按真实系统资格和 Project Arch scenario contract 编写场景资产 | 场景资产、导航投影、数据隔离和验证定义 |
| `/scenario-run` | keep `/scenario-run` | 执行项目定义的真实 API/UI 场景并采集当前证据 | 场景结果、失败位置、脱敏证据和 cleanup 状态 |
| `/scenario-diagnose` | rename `/scenario-investigate` | 区分场景设计、环境、资产和实现故障并定位最小 owner | 根因分类、复现证据和 owner 建议 |
| `/bug-record` | rename `/bug-report` | 把可复用故障根因写入 Bug 知识库 | Bug 记录、导航投影与写入证据 |
| `/delivery-retrospect` | rename `/retrospective` | 从当前交付证据提取系统性流程改进 | retrospective、导航投影与后续建议 |
| `/delivery-commit` | name-only rename `/work-journal` | 保持原合同完成 commit、当日日志和 INDEX 原子收口 | journal entry、INDEX、英文 ASCII commit |

数量守恒必须可审计：20 个当前 Skill 中，15 个进入 rename/keep/merge 输入，`/implement`+`/tdd` 与 `/scenario-env`+`/scenario-redeploy` 两次合并后得到 13 个；新增 `/environment-build` 后目标为 14 个。其余 5 个当前 Skill 中，`/create-doc` 与 `/sync-doc-index` 工具化，`/frontend-design`、`/skill-creator`、`/agent-browser` 删除。

以下能力不进入目标 Skill 集合：

<!-- harness-table: removed-skill-matrix/v1 -->
| 当前能力 | 结论 | 承接边界 |
|----------|------|----------|
| `/create-doc` | `tool`：移除顶层 Skill wrapper | 文档写入、Header 和导航投影并入 Project Arch docs transaction |
| `/sync-doc-index` | `tool`：按用户确认删除 Skill | Header/INDEX `check/fix` 并入 Project Arch tooling 与 docs gate |
| `/frontend-design` | `remove`：按用户确认删除且不重命名 | UI 当前合同仍由 `docs/ui-design/` 持有，设计/实施分别由 `/spec-design` 与 `/delivery-execute` 承接 |
| `/skill-creator` | `remove`：按用户确认删除且不重命名 | 通用 Skill 开发不是项目 Harness 主路径；不得在仓库内保留重复平台能力 |
| `/agent-browser` | `remove`：按用户确认删除且不重命名 | 浏览器是 `/scenario-run`、`/delivery-review` 等能力按需使用的执行工具，不是独立项目 workflow Skill |

四个用户明确删除项必须在对应阶段完成实体目录、Skill 清单、自动触发、调用方和当前可执行治理入口的 zero-reference；本 Spec 的迁移映射以及历史 journal、Bug、report 和 migration 中的原始名称保留为设计/历史事实，不做机械改写。`/sync-doc-index` 删除前必须先迁移其确定性脚本和 Make/docs gate；其余三个删除项不创建同义替代 Skill。

## 8 Project Arch 与业务扩展职责矩阵

| 文档或目录 | 唯一职责 | 不得承载 |
|------------|----------|----------|
| `init-arch` bundled Blueprint | Project Arch 版本、核心骨架、模板参数、adapter schema、发现/初始化/升级/修复 SOP 和兼容规则 | 固定项目业务实例或运行 secret |
| `AGENTS.md` | Harness bootstrap、全仓安全/Git/证据底线，以及对 `docs/agent-workflow.md` 的唯一入口引用 | 完整 Skill 表、业务目录矩阵、自动触发流程和逐步 runbook |
| `docs/agent-workflow.md` | EasyInterview 唯一 Skill 编排、风险路由、handoff 和退出条件 | 单个 Skill 的内部算法、产品当前合同 |
| `docs/README.md` | Project Arch 安装版本标记、Docs Arch 导航和业务 owner 入口 | 复制各 subject 当前合同或 Skill 流程 |
| `docs/development.md` | 跨层开发合同、OpenAPI operation matrix 工作流和通用工程入口 | 单个业务 subject 的产品语义 |
| `docs/spec/<subject>/spec.md` | subject 唯一当前合同：目标、范围、不变量、失败语义和验收标准 | Skill 流程、实施 checkbox、commit 日志、历史 PASS |
| `docs/spec/<subject>/plans/<plan>/plan.md` | 一次有限交付的 delta、顺序、风险和验证策略 | 通用 Skill 算法、永久当前合同或无限历史 |
| `checklist.md` | 当前 plan 的可执行进度和当前 gate 状态 | Spec 正文、Skill 编排、历史交付叙事 |
| `context.yaml` | 仅作为迁移期旧入口与回放输入；目标态无 checked-in 人工 manifest | 当前设计、人工 discovery、执行命令或永久路由 owner |
| Project Arch owner index/cache | 从路径、Header、链接、代码/API/route 标识和 Git 按需生成精确路由与引用解释 | 人工维护语义、跨 commit 复用陈旧缓存或替代 Spec/plan |
| `bdd-plan.md` / `bdd-checklist.md` | 真实用户行为与可观察结果 | Skill 流程、配置默认值、unit test、lint、build 或伪 E2E |
| `docs/ui-design/` | 当前 UI 信息架构、页面职责、流程、状态和视觉约束 | 第二套运行时或前端 Skill 内置业务知识 |
| `openapi/openapi.yaml` | wire API 唯一合同 | 在 Skill 或 `docs/apis/` 维护平行 schema |
| `docs/apis/` | API 使用说明、示例和补充解释 | 与 OpenAPI 冲突的字段或状态定义 |
| `test/README.md` | Project Arch 测试分层、聚合 gate 和代码证据入口 | 单个业务测试实现或历史 PASS |
| `test/scenarios/README.md` | Project Arch scenario/env 生命周期、真实 E2E、结果与恢复合同 | Skill 内部算法、固定项目组件或具体场景事实 |
| `test/scenarios/e2e/` | EasyInterview 真实 API/UI 场景内容与当前运行证据 | 代码测试包装、mock backend 或通用 Skill SOP |
| `test/scenarios/env-*.sh` | Harness 依据环境 owner Spec 实施的 Project Arch 环境生命周期 adapter | 另一套并行环境合同、人工外置实现责任或虚假 readiness |
| 项目环境 owner Spec | 当前项目的服务、依赖、端口、拓扑、健康、污染、恢复和验收合同 | Harness 通用 Blueprint、其他项目环境模板或平台级 Skill 算法 |
| 命中目录的 `README.md` | 该模块的具体路径、命令、依赖、生成规则、运行方式和局部约束 | 跨仓库通用 Skill 算法 |
| `docs/history/` | 仍有独立价值的重要 subject 版本历史 | 当前有效合同 |
| `docs/bugs/` | 可复用根因、严重故障和新失败模式 | 每次修复流水账或 Bug Skill 算法 |
| `docs/reports/` | 审计、复盘和跨会话 handoff 证据 | 当前 Spec、checklist 或 retrospective Skill 算法的副本 |
| `docs/discuss/` | 尚未收敛的分析和备选方案 | 已确认的当前合同 |
| `docs/work-journal/` | 按现有 Skill 保存提交级工作历史和索引 | 当前产品/工程合同；其现有结构与规则保持不变 |
| 各级 `INDEX.md` | 从 Header 和真实路径投影的导航 | 独立语义、主状态或 Skill 编排 |

## 9 规则唯一 owner

| 规则主题 | 唯一 owner | 其他位置的允许行为 |
|----------|------------|--------------------|
| Project Arch 版本、骨架、模板参数和 upgrade | `init-arch` bundled Blueprint | `docs/README.md` 记录安装版本；仓库 gate 验证，不复制 Blueprint |
| 项目 Harness 构建要求 | 对应 project owner Spec | 编排层驱动 `/spec-design`、`/delivery-execute`、environment/scenario 能力实施；不得要求人工另写 adapter |
| Harness 启动入口 | `AGENTS.md` | 引用唯一编排文档，不复制流程 |
| Skill 选择、排序和 handoff | `docs/agent-workflow.md` | `AGENTS.md` 仅引用；Skill 返回标准结果 |
| 单个 Skill 的指导思想、Arch contract、SOP、证据和恢复 | 对应 `SKILL.md` | 编排文档引用能力，不复制算法 |
| 业务 owner、路径、命令和验收 | 对应 docs 与最近 `README.md` | 编排层选择；Skill 通过标准 Arch 角色消费，不内置实例 |
| 分支、高风险确认、安全与 Git 基线 | `AGENTS.md` | 编排文档引用，Skill 执行适用输入 |
| TDD 原则 | `AGENTS.md` | 编排文档决定何时执行；`/delivery-execute` 将 TDD 作为内部 SOP 而非独立 Skill |
| 风险证据经济与主证据 owner 政策 | `AGENTS.md` | 编排文档选择强度；`/spec-design`、`/delivery-execute` 和 review Skill 执行但不复制政策 |
| 通用能力寻源与依赖准入政策 | `docs/development.md` | 编排文档决定何时寻源；Change/Decision 记录具体结论，模块 README 记录已采用依赖 |
| test 分层与聚合 gate | `test/README.md` | Skills 按 Arch 接口执行；项目测试实现靠近代码 owner |
| 环境设计与项目环境合同 | 项目环境 owner Spec | `/environment-build` 设计和构建；`/environment-operate` 只操作已建成 adapter |
| scenario/env 生命周期与真实 E2E 定义 | `test/scenarios/README.md` | `/environment-operate`、`/scenario-author`、`/scenario-run` 和 `/scenario-diagnose` 执行标准 SOP；项目 adapter 填充组件和命令 |
| 前后端/OpenAPI 契约流程 | `docs/development.md` | 相关 README、plan 和编排文档引用 |
| 文档格式、Header、模板和 INDEX 投影 | 对应目录 README/TEMPLATES、Project Arch docs transaction 与仓库 docs gate | 不设 `/create-doc` 或 `/sync-doc-index` 顶层 Skill |
| owner 路由与引用投影 | Project Arch tooling 的 generated-index contract | Skill 接收候选、命中证据与 owner 路径；不维护 `context.yaml` discovery |
| subject 当前设计 | 对应 `spec.md` | plan、测试和代码引用 |
| 当前 UI 设计 | `docs/ui-design/` | 编排层选择，前端能力与正式代码消费 |
| wire API | `openapi/openapi.yaml` | fixtures、generated artifacts 和 docs 校验一致 |
| commit 语言与格式 | `AGENTS.md` 与 Git gate | `/delivery-commit` 按冻结的既有行为合同执行 |
| delivery commit 能力合同 | 由 `.agent-skills/work-journal/SKILL.md` name-only 迁移为 `.agent-skills/delivery-commit/SKILL.md` | 只改入口名称；manual/auto 与 commit/journal/INDEX/ASCII 行为保持不变；调用时机由 `docs/agent-workflow.md` 编排 |

新增或修改规则前必须先确认 owner。若目标位置不是 owner，只能添加链接、简短执行约束或校验，不得复制完整规则。

## 10 渐进迁移约束

### 10.1 迁移顺序

本 Spec 的后续实施必须先定义所保护的不变量、当前基线、代表性回放、成功阈值、止损条件、回退方式和旧入口退出条件，并按以下 Arch-first 顺序推进：

1. 定义 Project Arch v1 的最小内核、模板参数、adapter schema、自举状态、`/init-arch` 合同、环境 build/operate 分工和 20→14 目标映射；
2. 实现 `/init-arch` 与 `scripts/harness_arch.py`，先在 fresh fixture 和当前 EasyInterview 验证安装、同版本幂等、legacy upgrade、冲突恢复与业务内容保留；
3. 建立 `/environment-build` 与 `/environment-operate`，使用至少两个 Harness 自有异构最小 fixture repository 回放“项目要求 → 环境 owner Spec → 环境资产 → lifecycle/gates”；`local-dev-stack` 只作为 EasyInterview upgrade/regression 输入；
4. 按已确认映射 rename/merge 14 个目标 Skill，使其 Arch-aware，并将 workflow/AGENTS 的当前执行入口一次性切换为新名称；除 `/init-docs` 限时 upgrade alias 外不保留旧名 alias；
5. name-only 迁移 `/work-journal` → `/delivery-commit`，在归一化允许变化字段后证明 manual/auto、commit/journal/INDEX/ASCII 合同等价；
6. 无替代删除 `/frontend-design`、`/skill-creator`、`/agent-browser`，并证明 UI、浏览器工具和平台 Skill 开发边界没有形成同义 wrapper；
7. 用 generated owner index/cache 替代 `context.yaml` 路由和引用职责，将文档事务和 Header/INDEX 能力并入 Project Arch tooling；完成调用方迁移后删除 `/create-doc`、`/sync-doc-index`；
8. 执行 fresh/upgrade 两种安装态的 R0-R3 单能力和组合回放，覆盖项目血肉生成、规则级联、上下文放大、错误路由、delivery-execute→delivery-commit owner 和环境恢复；
9. 完成全量负向搜索与 root gates 后，按退出条件删除 checked-in `context.yaml`、`/init-docs` alias、旧 contract tests 和无独立价值资产；
10. 完成 A1-A25 审计、retrospective 与计划生命周期收口。

### 10.2 迁移期保留

1. 当前 workflow Skill 与 docs 目录在对应阶段完成前继续可用。
2. 不得以本 Spec 为依据批量删除旧 plan、checklist、BDD、`context.yaml`、INDEX 或历史日志。第 7 节明确退出的 Skill 只能在各自替代 owner、调用方迁移、zero-reference、验证和回退条件满足后按 Phase 8/9 删除，不得提前跳序。
3. 已存在的 Header、INDEX、context 和 plan lifecycle 继续作为恢复证据与迁移输入，但 `context.yaml` 不得成为目标态 owner；旧 Harness 的自动级联、强制预读和 gate 在新体系实施期间只作基线观测，不得阻塞新 owner 的顺序实施。最终收口前必须由新体系的等价 gate 承接必要质量边界，并清除所有未解释的失败。
4. 一次 Change 只收敛一个清晰重叠面；Project Arch 合同与 `init-arch` 必须先于依赖它的 Skill API 调整。
5. 任何删除必须先证明唯一语义已被 docs/README 或可执行 gate 完整承接，并有负向搜索和回退证据。
6. 临时兼容入口必须记录退出条件和最晚复核点；退出条件已满足后继续保留重复入口，视为目标态漂移。
7. “保留”只保护迁移期间的稳定性，不得推翻第 3.7 节对最终目标态的奥卡姆判定。

### 10.3 `delivery-commit` 行为冻结例外

用户确认把 `/work-journal` 入口 name-only 迁移为 `/delivery-commit`，同时冻结其全部可观察行为：

- 只允许修改目录名、frontmatter `name`、自身路径引用和编排入口；
- 不修改其 manual/auto 参数和 commit、日志、INDEX、ASCII-only 合同；
- 不修改其逐 commit 原子收口和现有验证步骤；
- 独立编排文档只定义何时调用该 Skill，不复制或改写其内部步骤；
- 其他 Skill 不得以这个冻结例外为理由继续内置 EasyInterview 业务内容。

当前 `/tdd` 到 `/work-journal` 的调用关系在 `/delivery-execute` 合并阶段迁到独立编排层，并把入口切换为 `/delivery-commit`；迁移前后必须保持输入和可观察结果等价。这属于名称、调用方与编排 owner 的调整，不属于修改 commit/journal 行为。

### 10.4 BDD 适用性

本 Spec 描述内部 Harness 治理与文档职责，不新增用户可感知产品行为，因此本轮不创建 BDD 文档。后续若某个收敛 Change 改变真实 API/UI 行为，仍由对应业务 docs/README 按现有 BDD/E2E 门禁定义资产。

## 11 验收标准

| ID | 标准 |
|----|------|
| A1 | `/init-arch` 内置 canonical Blueprint、模板参数和 adapter schema，无需 caller 外部提供 framework/schema/templates，即可安装 Project Arch v1 的 Docs Arch 与 Env/Test/Scenario 最小内核 |
| A2 | fresh bootstrap 能从仓库事实、用户目标和既有 owner 自动形成 Spec/decision handoff，并由 Harness 主路径生成或修订具体 docs、tests、scenarios 与 env adapter；除决策、凭证和高风险授权外，不要求人工实现项目血肉 |
| A3 | 14 个目标 Skill 全部使用第 7 节确认的语义化名称，声明兼容 Arch 版本，并同时具备指导思想、Arch contract、可执行 SOP、证据结果和失败恢复，不直接调用另一个 Skill |
| A4 | Skill 可以固定认识 canonical Arch 路径、角色和入口，并依据项目 owner 生成业务实例，但不得硬编码 EasyInterview subject、route、operationId、表、事件、具体 E2E ID、组件、端口或项目专用命令 |
| A5 | `docs/agent-workflow.md` 成为唯一 Skill 编排 owner，`AGENTS.md` 只保留稳定底线和一个入口；R0-R3 顺序、必读 owner、handoff、失败恢复和退出 gate 只维护一次 |
| A6 | `/sync-doc-index`、`/frontend-design`、`/skill-creator`、`/agent-browser` 的实体 Skill、Skill 清单、自动触发、当前调用方和当前可执行治理入口均退出；本 Spec 的迁移映射与历史 owner 中的原始事实不机械改写 |
| A7 | Header/INDEX 必要校验由 docs 目录合同和确定性仓库 gate 承接，不因移除 Skill 而丢失 |
| A8 | `/work-journal` name-only 迁移为 `/delivery-commit`；除目录/frontmatter name/自身路径和编排入口外，manual/auto 输入及 commit/journal/INDEX/ASCII 输出合同保持等价 |
| A9 | TDD、真实 E2E、OpenAPI、后端持久化、安全、深度重校对和当前证据边界不得因解耦而降低 |
| A10 | 每个 Arch/Skill Change 都有 canonical-interface 正向测试、业务实例负向搜索、fresh/upgrade 等价验证、失败恢复和回退方式；项目血肉生成还必须有 Spec→资产→真实 gate 的纵向证据 |
| A11 | 代表性 R0/R1/R2/R3 任务的首次有效证据时间、预读量、工具调用和流程文件触碰量下降，且误路由、误阻塞、缺陷逃逸和恢复失败不增加 |
| A12 | 迁移可按当前 owner plan 顺序修改 Skill、`AGENTS.md`、编排文档和确定性 gate；旧 Harness gate 只作中间基线，最终必须由新合同下的全量验证证明无质量退化 |
| A13 | 每条新增或保留规则都能定位其防止的失败、适用风险、当前证据、总成本和退出条件；不能回答时不得成为永久全局规则 |
| A14 | fresh 与 upgraded 安装态的代表性回放同时覆盖 Project Arch 自举、单 Skill 和组合链路，包括规则级联、上下文放大、虚假高置信路由、delivery-execute→delivery-commit owner、scenario/env 污染与恢复失败 |
| A15 | 每个迁移阶段在执行前声明不变量、基线、成功阈值、止损、回退和旧入口退出条件，并实际验证回退可用 |
| A16 | 目标架构与迁移策略明确分离；`/init-docs` 等临时 alias 和双轨只用于 upgrade，满足退出条件后删除，不形成永久兼容层 |
| A17 | 普通风险逻辑在编译、类型、lint、现有邻近测试或仓库回归已提供充分证据时，可以以零新增专用单元测试收口 |
| A18 | 每个重要或关键风险有且只有一个主证据 owner；新增层级测试必须证明与既有测试不同的故障模式 |
| A19 | TDD 和 checklist 不要求每个实现项新增测试文件或测试函数；测试/业务代码比、测试文件数只作诊断信号，不作硬门禁 |
| A20 | 通用能力实施前能按寻源顺序给出 build-vs-adopt 结论；简单或业务特有逻辑不被强制创建外部调研和独立文档 |
| A21 | 采用成熟依赖后只测试项目拥有的 adapter 与风险边界；通用能力选择自研时明确完整所有权、范围上限和退出条件 |
| A22 | 同版本 `init-arch` 二次执行无 diff；legacy N→N+1 upgrade 保留所有非 Arch-owned 项目血肉，冲突和部分失败可从 checkpoint 恢复 |
| A23 | 目标仓库不存在 checked-in、人工维护的 `context.yaml` 或等价 discovery manifest；owner index/cache 可从当前仓库事实重建并对 commit/dirty state 失效 |
| A24 | 当前 20 个 Skill 全部映射到 keep/rename/merge/tool/remove 结论；目标态为 14 个 Skill，旧名称 hard-cut zero-reference，且合并不破坏权限、风险、证据或恢复 owner |
| A25 | `environment-build` 能依据至少两个 Harness 自有异构 fixture 的独立环境 Spec 构建不同环境血肉；`local-dev-stack` 只作为 EasyInterview 环境 owner 和 upgrade/regression 输入，不被引用为 golden fixture、Blueprint 或跨项目模板 |

## 12 失败与恢复

出现以下任一情况时，后续解耦 Change 必须停止：

- `/init-arch` 仍要求 caller 提供本应内置的 framework、schema、模板或 upgrade SOP；
- `/init-arch` 只安装空目录或参数表，随后要求人工编写项目 Spec、测试、场景或 env adapter；
- fresh repository 初始化后缺少 Docs Arch、test、scenario 或 env 任一核心接口，或 gate 只能返回虚假 PASS；
- `NOT_CONFIGURED` / `spec_required` 未被编排为 spec/environment design→delivery/environment build→operate/scenario verify 闭环，却被当作 Project Arch 建设完成；
- `environment-build` 只生成环境建议或参数表，没有负责形成 owner Spec、环境资产和真实 lifecycle evidence；
- `environment-operate` 反向决定环境拓扑，或 `local-dev-stack` 被当作通用 fixture/模板复制到其他项目；
- 同版本重复执行产生 drift，upgrade 覆盖项目业务内容，或部分失败无法恢复到 checkpoint；
- Skill 只剩抽象原则/合同表而没有可执行 SOP，或只剩命令而没有判断、证据和恢复框架；
- Skill 无法从 Project Arch 与项目 owner 恢复必要上下文；
- 同一 Skill 仍隐式读取、调用或依赖另一个 Skill 的内部实现；
- 现有 Skill 只做格式改写或 handoff 外置，没有形成逐项 keep/merge/tool/remove 结论，目标数量和依赖图未收敛；
- 新路径仍使用 `/design`、`/implement`、`/tdd`、`/plan-review`、`/scenario-env`、`/work-journal` 等旧入口，或为 hard-cut 重命名保留无期限 alias；
- `/frontend-design`、`/skill-creator`、`/agent-browser` 被同义改名后继续留在项目 Harness，或 `/sync-doc-index` 仍以 Skill wrapper 存在；
- 业务规则从 Skill 移出后没有唯一文档 owner，或在多个 README 重复；
- `AGENTS.md` 与 `docs/agent-workflow.md` 同时维护完整编排；
- 移除 `/sync-doc-index` 后 Header/INDEX 漂移失去确定性 gate；
- owner 路由准确性下降，或恢复已有 plan 所需信息丢失；
- checked-in `context.yaml` 或等价人工 discovery manifest 仍是新路径必需入口；
- 候选方案没有经过同一风险边界下的度量比较，只凭文件数量、通用词或主观简洁感决策；
- 新增规则无法说明防止的失败、实际成本、有效证据或退出条件；
- 普通风险逻辑被按 checklist 项强制新增专用单元测试，或同一合同在多个层级重复断言；
- 低频但高影响的删除、补偿、恢复、安全或数据一致性路径被误判为普通风险；
- 通用能力的分类、寻源状态、准入结论或边界不满足 `docs/development.md` 的唯一能力寻源合同，却继续进入实现或宣告完成；
- 采用或自研后的实际依赖、项目自有边界、验证对象或退出条件偏离已确认的 sourcing handoff；
- 单项 gate 均通过，但 Skill 组合回放出现误阻塞、上下文放大、错误路由或不可恢复状态；
- 临时兼容入口缺少退出条件，或退出条件满足后仍形成无期限双轨；
- TDD、E2E、安全、Git 或高风险确认门禁发生降级；
- `/delivery-commit` 除 name-only 迁移外改变输入、提交、日志/索引原子性或 ASCII commit 合同；
- 当前工作树包含来源不明的同范围修改，无法安全区分。

恢复动作应优先回退单个 Project Arch/Skill Change、恢复升级 checkpoint、原编排入口或原仓库 gate。不得通过要求 caller 每次重传 Arch、复制业务规则回 Skill 或建立长期双轨掩盖职责不清。

## 13 用户决策

- 2026-07-17：用户否决一次性全量迁移方案，确认采用方案 A——暂时保留当前 workflow Skill 与 docs 体系，先完成职责归类和编排边界，再按独立 Change 渐进优化。
- 2026-07-17：用户明确要求 `work-journal` Skill 行为保持不变；2026-07-18 进一步批准 name-only 迁移为 `/delivery-commit`，内部事务与可观察合同继续冻结。
- 2026-07-17：用户确认 `/sync-doc-index` 将从 Harness Skill 体系移除；必要索引能力转由 docs 目录合同和仓库确定性 gate 承接。
- 2026-07-17：用户确认 Skill 必须与业务解耦，只负责通用流程和框架；具体业务内容由 `docs/` 和相关 `README.md` 定义。
- 2026-07-17：用户确认跨 Skill 编排由独立仓库文档持有，并在 `AGENTS.md` 中引用；Skill 不再内置相互调用关系。
- 2026-07-18：用户明确当前会话只对齐 Harness Spec，不实施 Spec；“忽略旧版工作流”表示本 Spec 讨论不受旧 Harness 级联约束，不代表授权修改 plan/checklist、代码、Skill 实体或执行迁移。
- 2026-07-17：用户确认以“轻量目标态 + 不败后胜的渐进迁移”为统一总纲；方案 A 只描述迁移策略，不构成永久保留旧资产的理由。
- 2026-07-17：用户确认将五事七计、度量比较链、战争成本、谋攻、不败后胜、系统论与奥卡姆原则转化为判断算法、迁移门禁和验收指标，不新增强制哲学表单或第二套文档流程。
- 2026-07-17：根目录或外部传入的 `spec.md` 只作为讨论输入；本文件继续作为唯一 active Harness owner。
- 2026-07-17：用户确认普通风险逻辑无需追求圆满测试；已有编译、类型、lint 和回归证据充分时，允许零新增专用单元测试，但重要与关键风险仍按失败影响补足唯一 owner 证据。
- 2026-07-17：用户确认简单或业务内部独有的能力允许直接自研；明显通用的能力必须先完成相称的生态寻源与 build-vs-adopt 比较。
- 2026-07-17：AI 高级资深工程师水平只作为错误概率判断输入，不替代安全、数据、外部合同、恢复和不可逆操作的风险门禁。
- 2026-07-17：用户将最高层工程哲学明确为 `Harness 工程框架 = Skill + Docs Arch + Env`；三者必须在落地项目中共同运行。
- 2026-07-17：用户确认当前项目的 docs 与 test/scenario 骨架来自 `init-docs`，目标是将其升级并重命名为内置共同 Blueprint 的 `init-arch`，而不是改成要求人工外部注入 framework 的薄壳。
- 2026-07-17：用户确认骨架与 SOP 跨项目相通，只有业务血肉不同；好的 Skill 必须“指导思想 + 可落地执行框架”虚实结合，并有实际操作、证据和恢复措施。
- 2026-07-18：用户确认 Harness 新体系负责构建项目自身的 Harness。Blueprint 是抽象层和模板参数，项目血肉由 Harness 根据项目要求与 owner Spec 自动设计、实施和验证，不由人工手工补齐；EasyInterview 的本地/场景环境建设 Spec 是该自举模型的实例。
- 2026-07-18：用户确认 `local-dev-stack` 不是 golden fixture 或通用模板；环境建设必须由专责 `/environment-build` 引导并完成，已建成环境的生命周期操作由 `/environment-operate` 承接。
- 2026-07-18：用户确认采用“领域 + 明确动作”的方案 A 命名：当前 20 个 Skill 收敛为第 7 节的 14 个目标 Skill；TDD 等方法下沉为内部 SOP，旧名称除 `/init-docs` 限时 upgrade alias 外直接 hard cut。
- 2026-07-18：用户明确删除 `/sync-doc-index`、`/frontend-design`、`/skill-creator`、`/agent-browser`。其中索引能力进入 Project Arch tooling，后三项不创建同义替代 Skill；同时确认 `/work-journal` 只改名为 `/delivery-commit`，行为合同继续冻结。
