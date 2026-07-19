# Harness v1 渐进迁移 Plan

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-07-19

**关联 Spec**: [Harness 工程框架收敛](../../spec.md) 2.4
**工作模式**: `loop`
**当前授权**: 用户已明确按新 Harness 体系实施代码、Skill 与治理迁移；R3 删除、公开入口切换和其他破坏性动作仍按 Phase 单独确认。

## 1 交付目标与边界

本计划是 Harness Spec 2.4 的唯一当前工作 owner。目标是按 Spec 第 10.1 节的 Arch-first 顺序，把当前 EasyInterview Harness 渐进迁移到 Project Arch v1，并以 A1-A33 的当前可执行证据证明新路径在正确性、恢复能力和风险控制上不退化。

计划只描述当前交付 delta、Phase/Iteration、行内进度、验证、checkpoint、恢复与退出。完成后，有效合同回写 Spec，当前证据留在代码/测试，交付事实进入 Git/work journal，本计划目录从工作树退出。

### 1.1 In Scope

- Project Arch v1 的四层 Docs Arch、bundled template、INDEX/current-truth lookup、docs transaction 和 `arch.*` 接口；
- `/init-arch` 与 `scripts/harness_arch.py` 的 init/check/upgrade/repair、自举发现、幂等、冲突与恢复；
- test/scenario/env 标准接口以及环境 build/operate 分工；
- 20 个现有 Skill 到 14 个目标 Skill 的 rename/merge/keep/new/tool/remove 迁移；
- `AGENTS.md`、`docs/agent-workflow.md`、`docs/development.md`、`docs/README.md`、`test/README.md` 和最近 README 的 owner 重分层；
- 最小 `context.yaml`、独立 checklist/BDD/test plan、`history.md`、checked-in template、`docs/ui-design/` 和已完成 plan 的语义承接；
- finite/loop 单一 `plan.md`、跨 turn/context compaction 的 checkpoint 与恢复；
- fresh/upgrade、R0-R3、单能力/组合链、异构环境和失败恢复回放；
- A1-A33 审计、旧入口负向搜索、当前根级验证与最终 plan 退出。

### 1.2 Out of Scope

- 产品功能、OpenAPI 业务语义、正式前后端用户行为或数据库业务模型变更；
- 为 Harness 迁移虚构产品 BDD 或 `test/scenarios/e2e/` 场景；
- 把 EasyInterview 的 subject、route、operationId、组件、端口、secret、Make target 或环境拓扑固化进通用 Skill/Blueprint；
- 保留无期限旧名 alias、双轨 workflow、通用阻塞状态、扩展 `context.yaml` 语义或等价的第二份 manifest/cache；
- 用文件数量下降、结构 PASS 或历史 checklist 代替 A1-A33 的当前语义与运行证据。

### 1.3 2026-07-18 当前基线

以下数据用于计划启动，不作为永久合同；Phase 0 必须用可重复脚本刷新并记录基线命令与结果。

| 基线项 | 当前事实 |
|--------|----------|
| 顶层 workflow Skill | 20 个，仍使用旧名称和旧调用关系 |
| `docs/agent-workflow.md` | 不存在 |
| `scripts/harness_arch.py` | 不存在 |
| `history.md` | 29 个 |
| `context.yaml` | 50 个（含本 plan 的单文件 target） |
| 独立 `checklist.md` | 49 个 |
| 独立 BDD 文档 | 40 个 |
| 独立 test plan/checklist | 16 个 |
| checked-in `TEMPLATES.md` | 7 个 |
| `docs/ui-design/` 文件 | 13 个 |
| `INDEX.md` | 37 个；机制保留，但尚未由 Project Arch docs transaction 统一重建 |
| Harness 当前执行 plan | 本文件是唯一当前 plan |

### 1.4 Phase 0 当前证据与已知红灯

基线 commit 为 `61e0e8b2419b7c91708a06b704ba7c6145f57f73`。本轮只把 Harness owner gate 作为完成证据；产品 backend/frontend gate 仅在后续 Change 实际触碰产品代码或真实用户链路时条件运行。

| 入口 | 当前结果 | 结论 |
|------|----------|------|
| `make docs-check` | PASS；Header/INDEX/Markdown link 零漂移 | 旧 docs gate 可作为迁移前只读基线 |
| `.agent-skills` focused contract | `152 passed` | 当前 20 个 Skill 的分散合同基线可重复 |
| 旧路由/实施/env/commit 代表 replay | `96 passed` | 可作为 hard cut 前的行为意图基线，但没有统一组合 runner |
| 最小 `context.yaml` validator | 原 `49/49` PASS | 本次需纳入 Harness 单文件 plan，并证明 plan/checklist 同路径与 consumer 去重语义 |
| `scripts/harness_index_test.py` | `8 passed, 1 failed` | 标题断言未从当前 Spec H1 派生；归 Phase 1.2 修复当前 source projection |
| production-script reference gate | `6 passed, 1 failed` | `scripts/harness_index.py` 无当前生产入口且与 Spec 2.2 冲突；归 Phase 2 吸收后删除 |

旧 `make test` 在用户澄清范围前曾被只读执行，并在 Python 聚合段命中上述两个 Harness 红项后停止；它及单独补跑的产品测试不作为本计划完成依据，后续不重复运行无关产品回归。

### 1.5 代表任务、fixture 与度量合同

| 风险 | 固定代表任务 | 必须观察的结果 |
|------|--------------|----------------|
| R0 | 在合成仓库中分别用精确 subject/identifier 与通用歧义词查询 owner | 精确查询给出唯一证据链；歧义查询低置信并解释候选，不写缓存或文件 |
| R1 | 修复一个合成 subject 的 Markdown 引用并由 docs transaction 原子重建 INDEX | 正文与投影同事务、二次执行 zero-diff、失败可回滚 |
| R2 | 在宿主机进程 fixture 中以 finite plan 增加一个可观察的 lifecycle 状态并完成 TDD/review | 最小 context→Spec/plan→实现→owner test→checkpoint 单链闭环，无独立 checklist |
| R3 | 在 legacy 容器 fixture 上执行带 custom 文件、secret 占位和注入冲突的 upgrade/删除预演 | 未确认时 fail closed；确认后保留项目血肉、脱敏、可 rollback/resume，不产生虚假 ready |

两个 Harness 自有环境 fixture 固定为：

1. **host-process**：使用标准库 HTTP 进程、动态端口和本地状态文件，覆盖 setup/status/verify/redeploy/cleanup 与异常进程恢复，不含 EasyInterview 组件名、端口或依赖；
2. **containerized**：使用独立最小容器服务、不同健康模型和持久卷，覆盖 build、readiness、污染、部分失败与 cleanup，不引用 `local-dev-stack`、产品 secret 或产品镜像。

fresh 输入是只有最小语言/build 事实、没有 Project Arch 文件的临时 Git 仓库；upgrade 输入是包含 legacy docs/env 接口、人类维护扩展、未知文件和脱敏 secret 占位的临时仓库。组合链固定覆盖 `spec-design→delivery-execute→delivery-review`、`environment-build→environment-operate`、`scenario-author→scenario-run→scenario-diagnose`、`delivery-execute→delivery-commit`。

每次 replay 只在 run-local 临时目录记录 monotonic 首次有效证据时间、首次证据前去重读取字节、工具调用数、流程文件触碰数、owner 命中、误路由、误阻塞和恢复结果；结果不得成为跨任务 manifest/cache。相同 commit、同一机器冷热条件各运行 3 次并比较中位数：owner/授权/失败检测与所有注入恢复必须 `100%` 正确，误路由、误阻塞、虚假 PASS、secret 泄露和项目血肉覆盖必须为 `0`；四项效率指标任一不得回退超过 `10%`，且至少两项改善不低于 `20%`，否则不退出旧入口。

### 1.6 Phase 1-7 Change 控制

| Phase | Rollback checkpoint | 止损条件 | 需要单独 R3 确认 | 旧入口退出条件 |
|-------|---------------------|----------|------------------|----------------|
| 1 | Phase 0 commit + 每个 Arch-owned 原子写入前 fixture snapshot | 覆盖 custom/secret/业务内容、同版本有 diff、失败不可恢复 | 对当前仓库执行 upgrade 写入或改变 `/init-docs` 公开入口前 | `init-arch` fresh/check/upgrade/repair owner tests 通过；只保留有期限 alias |
| 2 | 每个资产族迁移前 Git checkpoint + docs transaction rollback | 找不到唯一语义 owner、INDEX 无法重建、查询需要第二份持久化 cache | 批量删除 history、独立 plan 附件或 `docs/ui-design/` 前 | docs transaction/lookup 接管 caller，最小 context 与正文/INDEX 原子一致 |
| 3 | 每个 fixture lifecycle 前后快照与 cleanup 证据 | fixture 反向固化产品栈、污染未清理、虚假 ready | 操作非 fixture 环境、删除真实环境状态或凭证边界变化前 | 两个异构 fixture 闭环通过，build/operate owner 清晰且可恢复 |
| 4 | 目标 Skill/AGENTS/workflow hard cut 前 Git checkpoint | 名称、caller、workflow 或 compat marker 不能同 Change 原子切换 | 切换公开 Skill 名称、AGENTS 入口和删除旧执行入口前 | 14 个 Skill 独立合同与唯一 workflow 通过，旧名除限时 alias 外 zero-reference |
| 5 | delivery-commit 等价 fixture + 删除前 Git checkpoint | manual/auto/commit/journal/INDEX/ASCII 任一可观察行为漂移 | name-only rename 和删除明确退出的 Skill 实体/当前 caller 前 | 冻结合同归一化等价，删除项无同义 wrapper 或活跃引用 |
| 6 | 每组 replay 前 fresh fixture + upgraded fixture snapshot | 质量/恢复回退、连续两轮无 progress、指标改善依赖降低证据 | 仅 fixture 内故障注入已在本计划授权；触碰真实环境或不可逆状态前另行确认 | R0-R3 与四条组合链满足 §1.5 阈值，无新增误路由/误阻塞 |
| 7 | 最终审计前 Git checkpoint 与完整 rollback 清单 | 任一 A1-A33 缺当前 owner 证据、目标路径仍依赖旧合同 | 删除 `/init-docs` alias、旧 contract、迁移 wrapper 和当前 plan 前 | A1-A33、focused Harness gates、replay、docs check、负向搜索全部通过并完成原子收口 |

## 2 Loop 工作合同

### 2.1 Progress Predicate

一次 Iteration 只有在至少关闭一个当前最高价值差距，且没有降低质量不变量时才算取得进展。判断依据为：

- A1-A33 中具备当前 owner 证据的验收项增加；
- 目标执行路径对旧 Skill 名称、Skill-to-Skill 调用、`context.yaml` 扩展语义、独立 checklist/BDD/test plan、checked-in templates 和通用阻塞机制的可执行依赖减少；
- fresh/upgrade、单能力/组合链和失败恢复回放的已验证范围增加；
- 在同一风险边界下，首次有效证据时间、预读量、工具调用和流程文件触碰量下降，且误路由、误阻塞、缺陷逃逸和恢复失败不增加；
- 同版本执行保持 zero-diff，升级或失败恢复不覆盖项目血肉、不泄露 secret、不留下环境污染。

单纯删除文件、调整命名、增加 PASS 文案或降低测试数量不构成进展。

### 2.2 不可突破的不变量

1. TDD、真实 E2E、OpenAPI、后端持久化、安全、深度重校对、Git 与当前证据边界不得降低。
2. 同一语义只有一个当前 owner；INDEX 只做可重建导航投影。
3. Skill 必须 Arch-aware、业务解耦、可独立执行，不直接调用另一 Skill。
4. 新路径只读取和维护最小 `context.yaml` 的 target/一等文档链接，不消费扩展语义，也不建立替代 manifest/cache。
5. 单文件 plan 的 `plan`/`checklist` 同指 `plan.md`；consumer 保留角色但按绝对路径去重正文读取。
6. `/work-journal` 到 `/delivery-commit` 只允许 name-only 变化，其 manual/auto、commit、journal、INDEX、ASCII 和原子性合同冻结。
7. `local-dev-stack` 只作为 EasyInterview 环境 owner 与 upgrade/regression 输入，不成为通用 Blueprint 或 golden fixture。
8. loop 不扩大用户授权；R3 删除、重命名、公开入口切换和不可逆操作继续要求人类确认。

### 2.3 Iteration Protocol

每个 Iteration 按以下顺序推进，且一次只收敛一个清晰重叠面：

1. 读取当前 Spec、当前 checkpoint、Git、测试和 Env 状态；
2. 从最早未完成 Phase 选择价值最高、风险边界最清晰的差距；
3. 明确该 Change 的不变量、基线、主证据 owner、成功阈值、止损、rollback 和旧入口退出条件；
4. 对代码/脚本/Skill 逻辑执行 Red-Green-Refactor，普通风险复用已有证据，重要/关键风险只补最接近行为 owner 的不同故障模式；
5. 执行 focused owner gate 和受影响的 Harness 聚合 gate；只有实际触碰产品代码或真实用户链路时才补对应产品 gate，环境/组合行为使用独立 fixture 或真实 lifecycle evidence，不包装为产品 E2E；
6. 将新的当前合同回写 Spec/README，原子刷新 INDEX；
7. 更新本 plan 的行内 checkbox 和第 7 节 checkpoint，只保留恢复所需当前状态；
8. 在可恢复边界使用交付提交能力建立 Git/work-journal checkpoint；
9. 判断继续、完成、暂停或请求决策。

### 2.4 预算、止损与暂停

- 单个 Iteration 不跨越两个独立 owner 切换；若必须跨越，先拆分 Change 或重新确认计划。
- 同一失败条件连续两个 Iteration 没有改善 progress predicate 时暂停，记录证据和 resume condition，不继续堆叠 wrapper、测试或文档。
- 当前 Spec 失效、替代 owner 不完整、命中的 Harness owner gate 出现新增红灯、rollback 不可用、secret/项目血肉可能受损或需要新的架构决策时立即停止。
- 外部凭证、不可获取基础设施或 R3 授权缺失时返回精确 handoff，不以 `NOT_CONFIGURED` 或空脚本伪装 ready。

### 2.5 Bootstrap 执行入口

在 Phase 4 的 `/delivery-execute` 和 `docs/agent-workflow.md` 尚未可用之前，本迁移采用最小 bootstrap 路径：执行 Agent 直接读取 Spec、本 plan、Git/test/Env 当前状态和命中目录最近 README，按第 2.3 节逐 Iteration 实施并执行 TDD/相称验证。该路径不是新 Skill、wrapper 或长期编排 owner，也不修改旧 `/implement`、`/tdd` 去兼容单 plan。

- 使用本 plan 的最小 `context.yaml` 进入 `/implement`/review 主路径，`plan` 与 `checklist` 同指 `plan.md`；不创建临时 manifest、转换器或独立 checklist；
- 每次 Iteration 仍需用户明确实施授权、语义分支、当前测试证据、rollback 和适用 R3 确认；
- `/delivery-commit` 完成前继续使用冻结的现有 `/work-journal` 行为建立 phase checkpoint，不改变其合同；
- Phase 4 完成目标 workflow 与 `/delivery-execute` 切换后，bootstrap 路径立即退出，不保留第二入口。

## 3 质量门禁分类

- **Plan 类型**: `architecture + tooling + migration + internal-contract`。
- **TDD 策略**: 所有 Project Arch tooling、Skill 逻辑、迁移脚本、fixture helper 和检查器先在最接近 owner 的 Python contract/integration test 建立 Red；普通静态文本调整可复用 `make docs-check`、lint、负向搜索与 `git diff --check`，不机械新增专用测试。
- **BDD 策略**: `BDD-N/A`。本 Spec 不新增用户可感知产品 API、UI 或业务流程，不创建 BDD Phase、BDD 文档或产品 E2E。
- **替代验证 gate**: owner contract test、fresh/upgrade fixture integration、同版本幂等/冲突/rollback test、Skill contract、INDEX/current-truth drift check、环境 lifecycle fixture、组合 replay、Harness 聚合 gate、`make docs-check`、`git diff --check` 和目标旧口径负向搜索；产品 gate 只在实际命中产品代码或真实链路时条件运行。
- **真实 E2E 边界**: Harness fixture/replay 属于内部 contract/integration evidence；只有既有产品场景被用于确认 EasyInterview upgrade 未破坏真实用户链路时，才由其现有 `E2E.*` owner 执行，不在本计划中分配新 E2E ID。

### 3.1 主证据 Owner

| 风险面 | 主证据 owner | 补充证据条件 |
|--------|---------------|--------------|
| Project Arch init/check/upgrade/repair | `scripts/harness_arch.py` 的 fixture contract/integration tests | 只有真实仓库组合能证明不同故障模式时补 EasyInterview upgrade smoke |
| Docs transaction、INDEX 和 lookup | Project Arch tooling tests | 原子写入/rollback 需要文件系统 integration evidence |
| Skill 五层合同、名称和独立性 | `.agent-skills/` contract tests | 组合路由在 Phase 6 replay 证明涌现风险 |
| 环境 build/operate | 两个 Harness 自有异构 fixture lifecycle tests | EasyInterview `local-dev-stack` 只补 upgrade/regression evidence |
| `/delivery-commit` 冻结合同 | 现有 work-journal contract 的前后等价测试 | 只归一化允许变化的名称、目录和路径字段 |
| Harness 不退化 | focused owner tests、目标 `make harness-test`、`make docs-check` 与负向搜索 | 命中产品代码或真实链路时才补对应产品 gate/scenario owner，不创建包装场景 |

### 3.2 能力寻源策略

- Project Arch docs transaction、路径解析、Markdown/Header 投影等通用能力在实施前按 `docs/development.md` 的目标寻源合同完成 build-vs-adopt 判断；
- 简单仓库特有 glue、确定性 allowlist 和小型迁移逻辑允许直接自研，并记录不适用理由；
- 采用依赖时只测试项目拥有的 adapter、错误映射、原子性和退出边界；
- 若目标寻源合同尚未落地，相关 Iteration 先建立该唯一 owner，不新增独立调研 manifest 或永久候选清单。

## 4 有序工作队列

### 4.1 Phase 0：基线、不败条件与回放骨架

**目标**：在修改执行入口前，建立可重复的当前基线、代表性任务集和逐 Change 回退合同。

- [x] 0.1 确认 Spec 2.2 是唯一 owner，用户确认采用单一 `plan.md` 的方案 A。
- [x] 0.2 盘点 20 个 Skill 与旧 Docs Arch 资产数量，记录 `docs/agent-workflow.md`、`scripts/harness_arch.py` 当前缺失状态。
- [x] 0.3 运行并记录当前 `make docs-check`、Harness/Skill contract tests、旧路径代表性 replay 与已知 Harness 红项；历史 PASS 和无关产品 gate 只作为线索。
- [x] 0.4 冻结 R0/R1/R2/R3 代表任务、两个异构环境 fixture 边界、fresh/upgrade 输入和组合链路集合。
- [x] 0.5 记录首次有效证据时间、预读量、工具调用、流程文件触碰、误路由、误阻塞、恢复结果的基线采集方式和成功阈值。
- [x] 0.6 为 Phase 1-7 分别声明 rollback checkpoint、止损条件、R3 确认点和旧入口退出条件。
- [x] 0.7 冻结单 plan bootstrap 执行入口：使用最小 `context.yaml` 且不复活独立 checklist；现有 `/implement` 直接消费同文件双角色，不增加旁路兼容入口。

**Phase Exit**：基线可重复、代表任务与 fixture 不绑定 EasyInterview 业务实例、当前红灯被显式列出，且首个 Arch Change 具备 rollback。

### 4.1A Phase 0A：旧 `context.yaml` 合同硬收缩

**目标**：在继续 Phase 1.2 前，执行用户于 2026-07-19 授权的单一 migration Change，把 `context.yaml` 收缩为最小文档链接清单，停止 discovery/references/metadata 扩展继续产生错误 owner 置信度和重复事实。

- [x] 0A.1 Red：更新 shared validator/generator 与 change-intake matcher 的 owner tests，要求 `metadata` 仅含 `name`、拒绝顶层/target `discovery` 和 target `references`，并证明 owner 匹配不再消费这些字段。
- [x] 0A.2 Green：收缩 validator、generator、candidate/matcher 和分支/版本推导；移除 preserve/normalize/score 兼容逻辑，所有额外 metadata/spec/target 字段 fail closed。
- [x] 0A.3 迁移 `.agent-skills/`、`AGENTS.md`、`docs/spec/TEMPLATES.md` 与全部现有 `context.yaml`，保证 generator 重跑确定性、清单链接仍可解析，且不保留旧字段说明或调用方。
- [x] 0A.4 验证 focused contract、49 份清单 batch validation、受影响 Skill contract、Harness/docs gate、字段 zero-reference 与 `git diff --check`；完成 post-pass reconcile、retrospective、work journal 和原子提交。

**BDD**：不适用。本 Phase 只收缩内部文档/tooling 契约，不新增或改变产品 UI、API、业务流程；替代 gate 为 owner contract tests、batch validation、Skill contract、docs check 与负向搜索。

**回退与止损**：以本 Phase 开始前 Git HEAD 为 rollback checkpoint。若任一旧 caller 无法从路径、Spec Header、`AGENTS.md` 或当前仓库事实推导原执行必需信息，停止 hard cut 并回到 owner 设计，不恢复 discovery 置信度或静默保留未知字段。

**Phase Exit**：全仓当前清单满足唯一最小 schema，所有生产 caller 不读取退出字段；旧字段只允许保留在本 Spec/plan 的迁移/负向 gate、测试 fixture 以及已完成 plan/checklist/work journal/Bug/report 的冻结历史证据中；随后恢复 Phase 1.2。

### 4.2 Phase 1：Project Arch v1 与 `init-arch`

**目标**：先建立其他目标能力能够消费的 Project Arch 最小内核。

- [x] 1.1 Red：为 fresh `init`、只读 `check`、legacy `upgrade`、最小 `repair`、同版本 zero-diff、custom/conflicting 分类、部分失败 resume 和项目血肉保留建立 fixture assertions。
- [x] 1.2 Green：按本轮 code review 修复四项合同漂移：为本 plan 与 fresh Project Arch 安装最小 `context.yaml`；让 matcher 只接受 manifest owner、generator 在保留 target identity 时补齐新增一等链接；实现 versioned Blueprint、bundled templates、adapter schema、`<!-- project-arch: v1 -->` 安装标记、`scripts/harness_arch.py` 的 init/check/upgrade/repair 与根 `make harness-test`；让 Harness INDEX 标题断言跟随当前 Spec H1。Red/Green owner 分别为 shared/change-intake Skill contract tests、`scripts/harness_arch_test.py`、`scripts/harness_index_test.py` 和 Make 聚合 gate。
- [ ] 1.3 实现仓库事实 discovery 和 `ready/spec_required/decision_required/conflict` handoff；不得要求 caller 提供 framework/schema/templates 或手工编写项目 adapter。
- [ ] 1.4 安装四层 Docs Arch 与 test/scenario/env 最小接口；按需扩展不得预建空目录。
- [ ] 1.5 在 fresh fixture 与当前 EasyInterview checkout 验证幂等、冲突、rollback、secret/业务内容保留和精确 resume condition。
- [ ] 1.6 仅在同一能力 owner 已证明后保留 `/init-docs` 限时 upgrade alias，并记录最晚退出点。

**Phase Exit**：A1、A2、A10、A15、A22、A27 的 Project Arch 部分具备当前证据；`arch.*` 接口可被后续 Phase 消费。

### 4.3 Phase 2：Docs transaction、owner lookup 与旧文档迁移

**目标**：让文档写入、导航和 owner 解析由 Project Arch tooling 承接，并移除平行当前状态资产。

- [ ] 2.1 Red：以当前两个已知 Harness 红项为迁移输入，覆盖 Header/路径/Markdown 链接生成 INDEX、唯一 active plan 导航、原子写入/rollback、current-truth lookup、多候选解释和无持久化查询缓存；不通过改旧标题制造假绿。
- [ ] 2.2 Green：实现 docs transaction、INDEX check/repair/rebuild 和 current-truth lookup；`INDEX.md` 不拥有正文语义或独立状态。
- [ ] 2.3 按“当前事实 / 当前工作 / 独立知识 / 可删除投影”逐 subject 重分类，单个 Iteration 只迁移一个资产族或 owner 边界。
- [ ] 2.4 把最小 `context.yaml` 纳入 docs transaction：正文附件新增/删除时原子更新链接，validator 对未知字段 fail closed，consumer 只用它定位正文且不建立替代 manifest/cache。
- [ ] 2.5 把仍属当前工作的 checklist/BDD/test plan 内容折叠进唯一 plan，把有效 history/UI 合同回写 subject Spec，把独立知识迁入 Bug/report/Decision 后删除旧文件与空目录。
- [ ] 2.6 将 templates 迁入 Blueprint bundled assets，更新 `docs/README.md` 四层合同和按需扩展，原子重建受管 INDEX。
- [ ] 2.7 调用方全部切换后工具化确定性脚本，并删除 `/create-doc` 与 `/sync-doc-index` Skill wrapper；work-journal INDEX 仍由冻结交付提交合同维护。
- [ ] 2.8 用全量负向搜索证明目标执行路径不存在 `context.yaml` 扩展字段或第二份 manifest/cache、独立 checklist/BDD/test plan reader/writer、`history.md`、checked-in templates 或 `docs/ui-design/` owner。

**Phase Exit**：A7、A19、A23、A28-A31 具备当前证据，`make docs-check` 使用新 tooling 且所有 INDEX 可从当前仓库重建。

### 4.4 Phase 3：环境建设与运行分工

**目标**：证明 Project Arch 骨架能依据项目 Spec 构建不同环境血肉，而不是复制 EasyInterview 栈。

- [ ] 3.1 Red：为 `/environment-build` 的事实发现、owner Spec handoff、资产构建和验收，以及 `/environment-operate` 的 setup/status/verify/redeploy/cleanup、污染和恢复建立 contract assertions。
- [ ] 3.2 建立至少两个 Harness 自有最小 fixture repository：一种宿主机进程型，一种容器化；两者使用不同组件、端口和依赖形态。
- [ ] 3.3 Green：实现 `/environment-build`，使其形成/修订环境 owner Spec、构建配置/依赖/脚本/adapter 并完成环境领域验收。
- [ ] 3.4 Green：由 `/environment-operate` 只操作兼容 adapter，不反向决定拓扑或生成第二套合同。
- [ ] 3.5 回放 `NOT_CONFIGURED/spec_required` 到 design→build→operate→verify 的闭环，证明非决策缺口不会被转交人工。
- [ ] 3.6 以 EasyInterview `local-dev-stack` 做 upgrade/regression，验证项目 secret、拓扑和现有真实场景 owner 不被 Blueprint 污染。

**Phase Exit**：A2、A10、A14、A25 的环境部分通过；两种 fixture 都有真实 lifecycle、cleanup 与失败恢复证据。

### 4.5 Phase 4：14 个 Arch-aware Skill 与唯一编排切换

**目标**：完成 Skill 语义化收敛、业务解耦和 `AGENTS.md`/workflow owner 重分层。

- [ ] 4.1 Red：为目标 14 个名称、五层标题、唯一 Arch compat 标记、实际消费的 `arch.*` interface、业务实例负向搜索和 Skill-to-Skill 调用零残留建立 contract tests。
- [ ] 4.2 建立 `docs/agent-workflow.md`，唯一维护 owner discovery、R0-R3 路由、finite/loop、handoff、review/scenario/commit 条件、失败与恢复。
- [ ] 4.3 将 `AGENTS.md` 收敛为七项工程原则、全仓安全/Git/证据底线和一个 workflow 入口；移除重复 Skill 表、业务目录矩阵和 runbook。
- [ ] 4.4 更新 `docs/development.md` 的能力寻源/依赖准入和跨层合同，更新 `test/README.md` 与 scenario/env 标准接口；业务路径与命令继续由最近 README 持有。
- [ ] 4.5 按 Spec 第 7 节映射 rename/merge/keep/new 目标 Skill；`/delivery-execute` 合并 implement/TDD，TDD 作为内部 SOP，不保留顶层 `/tdd`。
- [ ] 4.6 每个 Skill 验证独立执行、显式输入、结构化结果、fail-closed、rollback/resume 和不兼容 Arch handoff。
- [ ] 4.7 一次性切换 workflow、AGENTS、Skill 清单、触发器和当前调用方；除 `/init-docs` 限时 upgrade alias 外，旧名称 hard cut。

**Phase Exit**：A3-A5、A9、A12-A13、A17-A18、A20-A21、A24、A26 通过；目标执行路径只包含 14 个 Project Arch-aware Skill。

### 4.6 Phase 5：冻结交付提交合同与明确退出项

**目标**：完成必须独立验证的 name-only 迁移和无替代删除。

- [ ] 5.1 Red：为 `/work-journal` manual/auto 输入及 commit/journal/INDEX/ASCII/逐 commit 原子结果建立前后等价 fixture，只归一化允许变化的名称、目录与路径字段。
- [ ] 5.2 Green：name-only 迁移 `/work-journal` → `/delivery-commit`，把调用时机移入 workflow，不修改其内部事务与验证步骤。
- [ ] 5.3 删除 `/frontend-design`、`/skill-creator`、`/agent-browser` 的实体、清单、触发器和当前调用方，不建立同义 Skill wrapper。
- [ ] 5.4 证明 UI 合同、浏览器执行工具和平台 Skill 开发边界分别由 subject Spec/frontend、scenario/review tool usage 和平台能力承接。
- [ ] 5.5 对 `/sync-doc-index`、`/frontend-design`、`/skill-creator`、`/agent-browser` 及全部旧 Skill 名称执行目标路径 zero-reference；Spec 映射与历史 journal/Bug/report 原始事实允许保留。

**Phase Exit**：A6、A8、A16、A24 通过，冻结合同等价且明确删除项没有同义替代入口。

### 4.7 Phase 6：fresh/upgrade、R0-R3 与组合回放

**目标**：验证单项正确之外的路由、级联、上下文、恢复和环境涌现风险。

- [ ] 6.1 在 fresh 与 upgraded 安装态分别执行代表性 R0/R1/R2/R3 单能力任务，比较当前基线指标。
- [ ] 6.2 回放 spec-design→delivery-execute→delivery-review、environment-build→environment-operate、scenario author/run/diagnose 和 delivery-execute→delivery-commit 组合链。
- [ ] 6.3 覆盖错误 owner 候选、多候选低置信、规则级联、上下文放大、错误路由、虚假 PASS、部分失败和 rollback/resume。
- [ ] 6.4 回放 finite plan 和 loop Iteration；跨 Agent turn、context compaction 和进程重启后只使用 Spec、plan checkpoint、INDEX、Git、test 与 Env 恢复。
- [ ] 6.5 验证仍有安全有效下一步时不会提前结束，命中决策/预算/无进展/风险条件时留下精确 resume condition。
- [ ] 6.6 执行两个环境 fixture 的污染/恢复组合回放和 EasyInterview upgrade regression；未知文件、secret 和部分状态不得被误报 ready。

**Phase Exit**：A10-A16、A22、A25、A32-A33 的组合与恢复证据全部通过，效率改善不以质量或恢复退化换取。

### 4.8 Phase 7：A1-A33 审计、旧路径退出与 plan 收口

**目标**：只在目标合同、实现、当前证据和恢复边界完整后删除临时兼容和当前 plan。

- [ ] 7.1 逐项审计 A1-A33，每项链接到唯一 owner 的当前证据；历史 PASS 和本 plan checkbox 不单独构成验收。
- [ ] 7.2 执行目标 Skill/旧名称、`context.yaml` 非法扩展与第二份 manifest/cache、独立 checklist/BDD/test plan、history、templates、ui-design、通用阻塞机制、Skill-to-Skill 调用和 `/init-docs` alias 的全量负向搜索。
- [ ] 7.3 删除已满足退出条件的 `/init-docs` alias、旧 contract tests、迁移 wrapper、空扩展和无独立价值资产；验证 rollback checkpoint 在删除前可用。
- [ ] 7.4 运行全部 focused Harness owner gates、`make harness-test`、fixture/replay、`make docs-check`、`git diff --check` 和目标旧口径负向搜索；只有实际触碰产品代码或真实链路时才条件运行对应产品 gate/现有真实场景。
- [ ] 7.5 将仍有效合同回写 Spec/README，把验证留在代码/测试，把交付事实交给 Git/work journal；只在具有独立知识价值时创建 Bug/report/Decision/retrospective。
- [ ] 7.6 原子刷新 INDEX，删除本 plan 目录；不得把 Header 改成 `completed` 后永久保留交付包。

**Phase Exit**：A1-A33 全部有当前证据，目标路径无旧运行合同，工作树和环境干净，Git/work journal 可追溯，当前 plan 已退出。

## 5 覆盖矩阵

| Source | 类别 | Plan Phase | 验证 | Negative Scope / 失败面 |
|--------|------|------------|------|-------------------------|
| A1、A2、A27 | Primary path | 1 | fresh fixture init→spec/adapter handoff→current checks | 外部 framework/schema/templates、空目录即 ready、要求人工补血肉 |
| A10、A15、A22 | Alternate / recovery / boundary | 0-1、6 | same-version zero-diff、legacy upgrade、custom/conflict、partial-failure resume、rollback | 覆盖项目血肉、无 checkpoint、失败后长期双轨 |
| A7、A19、A23、A28、A29、A30、A31 | Cross-layer contract / regression | 1-2 | 最小 context、docs transaction、INDEX rebuild、current-truth lookup、资产族迁移审计 | context 扩展语义/链接漂移/第二份 cache、独立 checklist/BDD/test/history/template/ui-design owner |
| A25 | Alternate / environment | 3、6 | host-process fixture、container fixture、EasyInterview upgrade regression | 把 local-dev-stack 当 Blueprint/golden fixture、虚假 readiness、污染未清理 |
| A3、A4、A5、A6、A24、A26 | Contract / non-current-negative | 4-5 | 14-Skill contract、compat marker、依赖图、trigger/caller zero-reference | 旧 Skill 名称、Skill-to-Skill 调用、业务路径/命令硬编码、同义 wrapper |
| A8 | Regression / frozen contract | 5 | manual/auto normalized equivalence、ASCII commit、journal/INDEX atomicity | 借改名改变参数、提交边界、日志结构或验证步骤 |
| A9、A13、A17-A18、A20-A21 | Risk / evidence / sourcing | 0、4、6 | ordinary/important/critical replay、主证据 owner、build-vs-adopt handoff | 普通逻辑测试配额、多层重复断言、高风险欠验证、重复调研 manifest |
| A11-A12、A14、A16 | System / emergence | 0、6-7 | R0-R3 指标对比、组合 replay、错误路由/误阻塞/恢复结果 | 通用 PASS 状态、注册表、规则级联、永久 alias/双轨 |
| A32-A33 | Loop / recovery | 6 | turn/context/process resume，仅使用 Spec/plan/INDEX/Git/test/Env | session manifest、第二状态机、无 progress predicate 无限循环、提前结束 |
| §2.4、§6.1-6.3 | Privacy / security / observability | 1、3、6 | secret redaction、project-owned preservation、日志/冲突/rollback evidence | secret 持久化、未知文件保留、冲突静默覆盖、环境拓扑越权 |
| §1 Out of Scope、§10.4 | BDD/UX N/A | 全阶段 | internal contract/integration/replay；必要时复用既有产品 E2E owner | 为 Harness 内部迁移虚构产品行为、把代码测试包装为 E2E |

所有 A1-A33 至少映射到一个 Phase 和一个当前证据 owner；Phase 7 必须复核映射未因实施 delta 漂移。

## 6 风险、回退与恢复

| 风险 | 早期信号 | 止损与回退 | Resume Condition |
|------|----------|------------|------------------|
| bootstrap 期间新旧文档合同冲突 | 新 tooling 未 ready，consumer 拒绝单 plan 双角色或新 INDEX | 保留当前可执行 gate并修复统一最小 manifest 合同；不创建旁路入口 | Project Arch docs transaction 可验证并接管全仓调用方 |
| 批量删除丢失唯一语义 | 删除前找不到当前 Spec/README/代码 owner | 停止该资产族，恢复最近 Git checkpoint，先完成语义分类和 owner 承接 | owner 路径、负向搜索、focused gate 和 rollback 均明确 |
| hard cut 导致路由不可用 | trigger/caller/AGENTS/workflow 不一致 | 回退单个 Skill Change，不保留半切换名称 | 同一 Change 内名称、实体、调用方和 contract tests 可原子切换 |
| `init-arch` 覆盖项目血肉 | fixture custom 文件或 EasyInterview 业务内容变化 | fail closed 并回滚 Arch-owned 写入 | inventory ownership 与冲突策略可区分 Arch-owned/custom |
| 环境 fixture 反向固化 EasyInterview | fixture 出现本项目组件、端口或 local-dev-stack 引用 | 删除污染 fixture，回到抽象 adapter schema | 两个异构 fixture 独立表达各自 owner Spec |
| `/delivery-commit` 可观察行为漂移 | normalized equivalence 出现非允许字段差异 | 回退 name-only Change，保留旧入口 | 所有 manual/auto/commit/journal/INDEX/ASCII 断言等价 |
| 指标优化诱导降级 | 文件/调用量下降但缺陷、误路由或恢复失败增加 | 拒绝退出旧入口，恢复到上一 checkpoint | 同一风险边界下质量与恢复不退化 |
| loop 无进展或无限扩张 | 连续两个 Iteration 无法改善 predicate | 暂停并压缩 checkpoint，请求决策或重校 Spec | 存在新的安全可执行差距、授权与可观察证据 |

## 7 当前 Checkpoint

- **最近完成**: 完成 Phase 1.2：本 plan 与 fresh Project Arch 均使用 single-plan 最小 context；matcher、generator、candidate、Skill/模板适配完成；Project Arch 四模式 CLI、回滚与 `make harness-test` 落地；INDEX 标题按当前 Spec H1 投影；聚合 Harness gate `181 passed`。
- **当前 Phase**: Phase 1。
- **当前未完成**: Phase 1.3：扩展仓库事实 discovery 与 `ready/spec_required/decision_required/conflict` handoff，不要求 caller 提供项目 adapter。
- **当前阻塞**: 无；本轮不执行 Phase 2 的 context 删除，后续 Phase 也不得以删除 manifest 为退出条件。
- **验证摘要**: 本轮 Red 明确复现 matcher 无 manifest 旁路、generator 后增链接丢失、single-plan candidate 无进度、Project Arch/Make 缺失；Green/Refactor 后 `make harness-test` 最终为 `181 passed`，Project Arch owner tests `13 passed`，50 份 manifest batch validation 全通过且 generator drift 为 0。
- **下一动作**: 完成本次全量 manifest/docs/static reconcile；后续按用户另行继续指令进入 Phase 1.3，不自动扩展当前四项 review remediation。
- **恢复入口**: 读取 Spec 2.4、本 plan 的最小 `context.yaml`、本节 checkpoint、`git status/diff/log`、focused Harness/Skill tests、`make docs-check` 和 fixture Env status；manifest 只定位正文，不作为 owner 事实或 session state。

## 8 完成与退出条件

本计划只有同时满足以下条件才可收口：

1. A1-A33 每项都有当前、唯一 owner 的可执行或可审计证据；
2. fresh 与 upgraded 安装态、R0-R3、单能力/组合链、两类环境和 loop resume 全部通过；
3. 当前 Harness owner/聚合 gates 与 docs gate 通过，目标旧口径负向搜索只剩 Spec 映射与允许的历史事实；产品 gate 只在实际命中产品代码或真实链路时适用；
4. rollback、cleanup、secret redaction、项目血肉保留和恢复条件均已实际验证；
5. 仍有效合同已进入 Spec/README，测试和检查进入代码 owner，提交事实可由 Git/work journal 追溯；
6. `/init-docs` alias、旧 Skill 名称、旧 contract/wrapper 和无独立价值资产按退出条件删除；
7. 本 plan 与其目录被删除，plans INDEX 和顶层 Spec INDEX 原子刷新。

若目标已满足、需要新的架构决策/凭证/权限、达到止损点、连续 Iteration 无进展、当前 Spec 失效或继续会扩大风险，则暂停而不是伪造完成；暂停必须更新第 7 节精确 resume condition。
