# Harness 工程框架渐进收敛计划

> **版本**: 3.1
> **状态**: active
> **更新日期**: 2026-07-18

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

依据 Harness Simplification Spec 1.8，以 `Harness 工程框架 = Skill + Docs Arch + Env` 为主轴完成方案 A：把现有 `init-docs` 升级为内置 canonical Blueprint 的 `init-arch`，使它能安装和升级共同的 docs、test/scenario 与 env 骨架；再逐步把当前 20 个 Skill 收敛为 14 个 Arch-aware、虚实结合的独立能力。最后迁移 Project Arch tooling，并用 fresh/upgrade R0-R3 回放和全量 gate 证明质量、恢复与效率不退化。

本目录沿用历史路径 `001-full-migration`，但不再执行一次性方案 B。每个 Phase 只收敛一个清晰重叠面，完成当前证据、回退和退出 gate 后才进入下一 Phase。

## 2 范围

### 2.1 In Scope

- `docs/agent-workflow.md` 唯一编排 owner 与 R0-R3 路由、handoff、失败恢复和退出合同。
- `AGENTS.md` 中稳定公共政策、Git/安全/证据底线和唯一编排入口。
- Project Arch v1 的 Docs Arch、test、scenario、env 四子系统角色、骨架、扩展点与 gate。
- `init-docs` → `init-arch` 的内置 Blueprint、bootstrap/check/upgrade/repair、幂等和回滚实现。
- 14 个目标 `.agent-skills/*/SKILL.md` 的 Arch compatibility、指导思想、可执行 SOP、证据和恢复合同；`/work-journal` 仅按冻结合同 name-only 迁移为 `/delivery-commit`。
- Harness 索引、路由、回放、Skill 合同和 docs/index 确定性工具及测试。
- `/create-doc` 与 `/sync-doc-index` 能力迁移到 Project Arch tooling，并移除顶层 Skill wrapper；删除 `/frontend-design`、`/skill-creator`、`/agent-browser`，不创建同义替代 Skill。
- R0/R1/R2/R3 单 Skill与组合链路回放、成本比较、回退演练和最终全量验证。

### 2.2 Out of Scope

- 产品功能、OpenAPI 业务语义、数据库 schema、正式前后端用户行为和真实 E2E 场景内容。
- 机械改写 Bug、report、work-journal 或 Git 中的历史事实以制造“零残留”。
- 在替代 owner 与回退证据完成前批量删除旧 plan/context/BDD/INDEX。
- 修改 `.agent-skills/work-journal/SKILL.md` 及其 manual/auto、commit、journal、INDEX、ASCII-only 行为合同。

## 3 质量门禁分类

- **Plan 类型**: tooling + docs architecture + governance + migration。
- **TDD 策略**: `scripts/harness_arch_test.py` 先覆盖 fresh bootstrap、same-version no-op、legacy upgrade、conflict/rollback 和项目血肉保留；现有 `scripts/harness_index_test.py`、`scripts/harness_workflow_test.py`、`scripts/harness_skills_test.py`、`scripts/harness_replay_test.py` 与 docs/env gate tests 按新 Arch contract 更新。普通文案迁移复用 contract lint 与回放，不按文件配额新增测试。
- **BDD 策略**: BDD-N/A；本计划不引入用户可感知产品行为。
- **替代验证 gate**: Python contract tests、静态耦合 lint、docs/header/index drift、代表性 R0-R3 replay、rollback replay、负向搜索、`make test`、`make build`、`make lint`、`make docs-check` 和 `git diff --check`。

## 4 声明式边界

1. 当前设计唯一真理源是 `../../spec.md`；本 plan/checklist 只拥有本次有限交付顺序与进度。
2. 测试结果、指标、缓存 commit/worktree fingerprint 和运行日志均为运行时派生证据，不写回 Spec 充当永久事实。
3. Bug、report、work-journal、migration 与 Git 是历史 owner；迁移工具不得把历史 plan/context 路径机械重写成当前 Spec 后改变其原意。
4. 本轮采用渐进迁移：旧入口可在等价影子阶段存在，但每个临时双轨必须有退出条件；最终不得保留无期限兼容层。
5. 用户允许迁移期间忽略旧 Harness 的强制级联与旧 gate；这不豁免新 Spec 的安全、恢复、证据和最终全量验证。

## 5 覆盖矩阵

| 来源 | 类别 | Owner Phase | 验证 | 负向范围 |
|------|------|-------------|------|----------|
| Spec A1-A2 | Project Arch 主路径 | 6-7 | fresh init + same-version no-op + legacy upgrade + four-subsystem gates | caller 外部注入 framework；缺失 env/test/scenario；覆盖项目血肉 |
| Spec A3-A5 | Skill/编排合同 | 2-3/6/8 | Arch compatibility + executable-SOP lint + R0-R3 handoff replay | 只有哲学或只有命令；Skill 调用 Skill；业务实例硬编码 |
| Spec A6-A8 | Skill 退出与冻结迁移 | 8-9 | removed-skill zero-reference + Arch docs projection + delivery-commit normalized parity | 无替代 gate 删除索引能力；修改冻结行为 |
| Spec A9-A10 | 回归/负向 | 7-11 | canonical-interface positive + business-instance negative + full root gates | 伪 E2E、历史 PASS、把标准 Arch 接口误判为业务耦合 |
| Spec A11/A14-A16 | 系统涌现 | 1/2/10 | fresh/upgrade replay metrics + ambiguity/misroute/TDD-closeout/scenario recovery fixtures | 硬编码 replay、永久 alias、无回退删除 |
| Spec A13/A17-A19 | 风险证据经济 | 4 | owner-evidence fixture + duplicate-test lint/review replay | 每项强制新测试、跨层重复同一断言 |
| Spec A20-A21 | 能力寻源 | 5 | build-vs-adopt fixture + dependency boundary review | 通用能力无寻源自研、重复上游测试 |
| Spec A22-A23 | Arch 幂等、恢复与 generated owner | 7/9-10 | same-version zero diff + checkpoint rollback + dirty-aware generated index | 覆盖项目血肉、人工 context 继续充当目标态 owner |
| Spec A24 | 20→14 Skill 收敛 | 8 | target/removed matrix parity + hard-cut zero-reference + permission/risk/evidence replay | 遗漏来源、同义替代、合并后 owner 失真 |
| Spec A25 | 环境构建泛化 | 7/10 | 两个 Harness 自有异构 fixture + EasyInterview upgrade/regression | 把 local-dev-stack 当作 Blueprint 或 golden fixture |
| Spec §10.3 | 不变量 | 全程 | work-journal source hash + delivery-commit normalized contract tests | 修改冻结行为或破坏日志原子性 |

## 6 实施阶段

### Phase 1: 恢复方案 A 基线与真实回放

#### 1.1 纠正 owner Spec、plan/checklist/context/index，使方案 B 不再拥有当前执行语义

#### 1.2 将生成式索引改为 additive 试点：不把现存 plan/context/history/INDEX 判作迁移期 forbidden，并让缓存对 dirty worktree 失效

#### 1.3 建立真实 R0-R3、误路由、误阻塞、失败恢复和回退基线；指标从实际文件/命令/结果采集，不使用硬编码完成数字

### Phase 2: 建立唯一编排 owner 的等价影子

#### 2.1 创建 `docs/agent-workflow.md`，完整承接当前请求分类、R0-R3 顺序、必读 owner、handoff、失败恢复与退出条件

#### 2.2 定义统一 handoff schema，并用代表性 replay 证明影子文档不改变现有触发结果

### Phase 3: 收敛 AGENTS 公共政策

#### 3.1 将不可下沉的安全、Git、高风险确认、证据/TDD 与风险证据经济底线集中到 `AGENTS.md`；能力寻源继续由 Phase 5 的 `docs/development.md` 独占

#### 3.2 删除 AGENTS 中重复 Skill 表与逐步 runbook，只保留 `docs/agent-workflow.md` 唯一入口和稳定 policy IDs

### Phase 4: 试点风险证据经济

#### 4.1 让 design、TDD 与 review 能力消费公共风险政策，以主证据 owner 和故障模式决定最小证据

#### 4.2 用普通/重要/关键 fixture 证明不会按 checklist 项机械新增测试，也不会降低安全、数据、恢复和外部合同证据

### Phase 5: 试点能力寻源

#### 5.1 将寻源顺序与依赖准入放到 `docs/development.md` 唯一 owner

#### 5.2 让入口、design 与 review 能力通过显式输入执行 build-vs-adopt，并验证简单业务逻辑不被强制外部调研

##### Capability fixture decisions

以下仅是 `scripts/harness_sourcing_test.py` 的合成回放 owner，不代表项目实际引入对应依赖，也不得生成独立调研包：

- `FIXTURE-QUICK-REUSE`：普通通用能力复用 `stdlib:pathlib`；项目边界是路径归一化 wrapper，平台合同变化时替换 wrapper。
- `FIXTURE-FORMAL-ADOPT`：重要通用能力在合成官方/成熟候选引用上选择 adopt；摘要覆盖适配、维护、许可证、安全/供应链、稳定性与替换回滚，项目只拥有薄 adapter，任一结论未知时保持 pending。
- `FIXTURE-FORMAL-BOUNDED-BUILD`：重要通用能力在候选均不适用时选择 bounded build；边界限制为 owner parser，退出条件是平台方案满足当前合同或维护成本超过该边界。

### Phase 6: 建立 Project Arch 与 `init-arch` 设计合同

#### Phase 6 execution contract

- **不变量**：`Harness = Skill + Docs Arch + Env`；当前仓库的 docs 与 test/scenario 来源于原 `init-docs`，其有效骨架和 SOP 不得因“去耦合”丢失；标准 handoff、POL-TEST、能力寻源与 R0-R3 确认/恢复语义保持不变；`work-journal` Skill 行为与内容保持不变。
- **当前基线**：原 `init-docs` bundle 仍保存 docs 与 scenario 模板，EasyInterview 已有完整 Docs Arch、`test/README.md`、scenario/env scripts 和 root gate；当前错误草案把 `init-docs` 改成要求 caller 提供 framework/templates/validation 的薄壳，尚未删除原模板。
- **成功阈值**：Spec 明确 Project Arch v1 文件树、四子系统、共同 SOP 与项目血肉边界；`init-arch` 拥有 `init/check/upgrade/repair` 和内置 Blueprint；每个目标 Skill 的质量合同要求指导思想、Arch contract、SOP、证据和恢复；plan/checklist 先实施 Arch 再迁移 Skill。
- **止损条件**：需要新增逐任务 architecture manifest、caller 仍需重传 framework、共同骨架无法覆盖当前 docs/test/scenario、env adapter 只能假 PASS，或 Project Arch 会接管业务事实时停止。
- **回退方式**：只回退本阶段 Spec/plan/checklist 修订；保留原 `init-docs` Skill、templates、当前 docs/test/scenario 与 env 脚本，不执行删除或 rename。
- **旧入口退出条件**：本阶段不退出 `/init-docs`、`/sync-doc-index` 或任何模板；只有 `init-arch` 实现、fresh/upgrade replay 和 caller 迁移通过后才允许删除兼容入口。
- **冻结证据**：`.agent-skills/work-journal/SKILL.md` SHA256 为 `4ceeb4567a655ec265728b40cbd44241ce8c6abc9033df83e6c59406beebc275`；任何变化均阻断 Phase 6。

#### 6.1 定义 Project Arch v1 的 Docs Arch、test、scenario、env 固定角色和扩展点

#### 6.2 定义 `init-docs` → `init-arch` 的 bootstrap/check/upgrade/repair、版本、幂等、冲突和回滚合同

#### 6.3 定义 Skill 的“指导思想 + Arch contract + 可执行 SOP + 证据 + 恢复”结构门禁

### Phase 7: 实现 `init-arch` 与四子系统 Blueprint

#### 7.1 先建立 `scripts/harness_arch_test.py`，覆盖 fresh init、同版本 no-op、legacy upgrade、custom/conflict、部分失败回滚和项目血肉保留

#### 7.2 将当前通用 docs/test/scenario 模板整理到 `.agent-skills/init-arch/blueprint/`，以 `.agent-skills/init-arch/scripts/init_arch.py` 实现 `init/check/upgrade/repair`；`/init-docs` 暂时保留为 alias

#### 7.3 让 `docs/README.md` 持有 `<!-- project-arch: v1 -->` 唯一安装标记，由目标 `scripts/harness_arch.py` 提供仓库内 check 入口，并证明 fresh fixture 与当前仓库四子系统合同完整

#### 7.4 建立 `/environment-build` 与 `/environment-operate`，使用至少两个 Harness 自有异构环境 fixture 证明环境 Spec、资产和 lifecycle adapter 可按项目血肉变化；`local-dev-stack` 只作 EasyInterview upgrade/regression 输入

### Phase 8: 收敛 14 个独立 Skill 与退出项

#### 8.1 按 Spec 第 7 节把当前 20 个 Skill 全量映射为 keep/rename/merge/tool/remove，并为 14 个目标 Skill 声明 Arch 兼容版本、输入/输出和依赖的 canonical interfaces

#### 8.2 保留或精炼每个 Skill 的指导思想、可执行 SOP、证据与恢复，移除 Skill-to-Skill 调用及 EasyInterview 业务实例

#### 8.3 统一语义化入口和唯一编排接口；除 `/init-docs` 限时 upgrade alias 外旧名称 hard cut，并把 `/work-journal` name-only 迁移为 `/delivery-commit`，以归一化合同回放证明行为不变

#### 8.4 删除 `/frontend-design`、`/skill-creator`、`/agent-browser` 的实体、清单、自动触发、调用方和当前治理入口，不创建同义替代 Skill

### Phase 9: 文档工具化与索引归位

#### 9.1 将 Header/INDEX 检查、修复和投影能力并入 Project Arch tooling

#### 9.2 将 `/create-doc` 与 `/sync-doc-index` 的文档事务、Header/INDEX check/fix/projection 能力并入 Project Arch tooling，迁移调用方后删除两个顶层 Skill wrapper

### Phase 10: fresh/upgrade 系统回放与旧入口退出

#### 10.1 在 fresh 与 upgraded EasyInterview 上执行 R0/R1/R2/R3 单 Skill 和组合链路，比较首次有效证据、预读、调用、流程文件、误路由/误阻塞、缺陷逃逸和恢复

#### 10.2 演练 ambiguity、规则级联、上下文放大、`delivery-execute`→`delivery-commit`、scenario/env 污染、失败恢复和 rollback

#### 10.3 满足退出条件后删除 aliases、重复编排、旧 contract tests 和无独立价值资产；保留真实历史 owner

### Phase 11: 全量验证与生命周期收口

#### 11.1 审计 A1-A25、Arch 正向接口、业务实例负向搜索、owner 唯一性和工作树范围

#### 11.2 执行新合同下 focused tests、`make test`、`make build`、`make lint`、`make docs-check`、链接/索引/fresh-upgrade replay 和 `git diff --check`

#### 11.3 完成 retrospective、plan/checklist/index 生命周期与交付收口

## 7 风险与恢复

| 风险 | 止损与恢复 |
|------|------------|
| 当前脏变更再次混入方案 B | 每阶段检查 name-status 与 owner 范围；方案 B 快照保存在仓库外，可审计但不自动重放 |
| 编排迁移改变触发语义 | 影子 workflow 与组合 replay 未等价前不移除旧入口；失败时回退当前阶段文件 |
| `init-arch` 退化为 caller 注入薄壳 | fresh init 必须只依赖 bundled Blueprint 与仓库事实；发现 framework/schema/template 外部前置时停止并恢复原 `init-docs` |
| Arch upgrade 覆盖项目血肉 | 先 inventory 与 preview，checkpoint 所有 Arch-owned target；custom/conflict 默认不覆盖，失败恢复精确 pre-state |
| Skill 去业务化后只剩哲学或空合同 | 每个 Skill 同时通过 judgment、Arch contract、SOP、evidence、recovery 语义 gate；缺一即恢复原有效算法再精简 |
| Skill 业务解耦丢上下文 | Skill 必须能从 Project Arch 标准角色和项目 owner 恢复；不得用 caller 重传整个框架补洞 |
| docs gate 迁移造成 Header/INDEX 漂移 | 新 deterministic check/fix 在删除 Skill 前完成等价、幂等和失败恢复验证 |
| work-journal 被误改 | 每阶段验证冻结 SHA256 和行为合同；出现差异立即恢复该文件并停止 |
| 只测单项却出现组合涌现失败 | Phase 10 同时在 fresh/upgrade 安装态执行单能力与组合链路回放，任何误阻塞/错误路由/不可恢复状态阻断退出 |

## 8 完成标准

- Spec A1-A25 均有当前文件或可执行证据，历史 completed/PASS 不作替代。
- 每个迁移阶段都记录不变量、基线、成功阈值、止损、回退和旧入口退出条件。
- `init-arch` 可以从空仓库安装 Project Arch，并对当前仓库执行无损 upgrade；同版本重复执行无 diff。
- 每个目标 Skill 都能基于 Project Arch 执行真实 SOP，且无需 caller 重传整个 Project Arch。
- 新 Harness 成为唯一可执行主路径，`/init-docs` alias、旧 Harness 约束和 gate 已被删除或明确降级为非阻塞历史资产。
- `/work-journal` 已 name-only 迁移为 `/delivery-commit`，归一化后的 manual/auto、commit、journal、INDEX 与 ASCII-only 可观察行为保持不变。
- 新合同下全量测试、构建、lint、docs、链接、索引、回放与 diff gate 全部通过。
