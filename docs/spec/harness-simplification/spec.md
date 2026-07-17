# 轻量级 Harness 上下文与技能体系

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-07-17

## 1 目标与范围

本 Spec 是 easyinterview Harness 文档、上下文加载、owner 路由和 Skill 分工的唯一当前设计真理源。目标是在不降低工程正确性、验证质量和风险控制的前提下，删除人工维护的派生文档、重复规则和无差别预读。

本次范围包括 `docs/spec/` 的信息模型、机器索引、AGENTS 公共政策、`.agent-skills/` 入口合同及旧结构清理。产品代码、OpenAPI 业务语义和真实场景本身不在迁移范围内。

## 2 信息模型

### 2.1 Spec

每个长期 subject 只有一个 `docs/spec/<subject>/spec.md`。Spec 只描述当前目标、范围、不变量、失败语义、owner 边界和可验证验收标准，不记录实施 checkbox、历史 PASS、commit 日志或可从仓库推导的文件清单。

### 2.2 Change

只有跨会话、多层协作、高风险决策、多阶段实施或用户明确要求留档的交付才创建 `changes/<YYYY-MM-DD>-<change>.md`。一个 Change 单文件承接 delta、范围、决策、风险、有序任务、行为和证据摘要；完成后冻结，后续修改创建新 Change，不重开旧 Change。

### 2.3 Decision

只有跨 subject、难以逆转或存在重要长期取舍的决定才创建 `decisions/<decision>.md`。普通实现选择留在 Change，现行结果回写 Spec。

### 2.4 Evidence 与机器索引

可执行证据归代码测试、契约 gate、数据库/API 查询、真实 API/UI E2E 和 Git 自然 owner。机器索引从 Spec/Change 路径、少量 Header、Markdown 链接、Git、源码符号、API、route 和测试资产生成；索引可删除、可重建、绑定当前 commit，禁止 checked-in 人工 discovery manifest。

## 3 目录合同

```text
docs/spec/
└── <subject>/
    ├── spec.md
    ├── changes/      # 仅非平凡交付
    └── decisions/    # 仅长期取舍
```

不再强制或维护 `context.yaml`、独立 `plan.md` / `checklist.md`、普通 `bdd-plan.md` / `bdd-checklist.md`、`history.md`、顶层或 subject 级 INDEX。供人浏览的目录页只能按需生成并标明 generated，不得成为状态 owner。

## 4 风险分级

- **R0**：只读解释，只加载用户指定事实和最小 owner，不创建分支、计划、日志或报告。
- **R1**：局部可逆修改，确认 owner Spec、执行 focused Red/Green 与相称回归，默认不创建 Change、BDD、Bug、retrospective 或工作日志。
- **R2**：行为或跨层交付，先确认/更新 Spec，创建一个 Change，覆盖主路径、失败恢复、跨层契约和持久化；代码逻辑使用 TDD，真实用户流程按风险使用 domain behavior test 或真实 E2E。
- **R3**：公开 API、迁移、安全隐私、删除、生产操作或跨组件架构变更，使用 Spec + Change + 必要 Decision，实施前获得人类确认并覆盖回滚、幂等、安全和可观测性。

不确定时按较高风险处理；范围变化必须显式升级或降级。

## 5 上下文与路由

初始上下文只加载用户请求、适用的 Git 摘要、恢复任务的当前状态、一个候选 owner Spec 和确实继续执行的 Change。只有具体未决问题出现时才读取契约、UI、Decision、历史或失败 gate；每次扩展必须说明它解决的问题。

路由证据优先级为：用户指定文件、精确 owner 声明、精确代码/API/route/表/事件/配置标识符、当前 diff/相关提交、标题与链接图、通用关键词。通用词不得单独产生高置信结果；候选接近或跨 owner 时必须输出 low confidence、命中理由和最多三个候选。

## 6 Skill 与公共政策

Skill 只描述独有操作步骤，能力收敛为 locate、design、execute、review、operate、closeout 六类。分支安全、TDD、E2E、契约、持久化、安全、风险确认和证据优先级由一份公共政策按风险加载，Skill 只引用规则 ID，不复制全文。

只读任务不触发写入 Skill；明确 owner 的 R1 不先 locate；明确 Change 直接 execute；review 不自动 fix；closeout 只在提交、交接或审计需要时触发。

## 7 必须保留的质量边界

1. 代码逻辑必须有可失败断言和当前运行证据。
2. 只有真实 HTTP/API 或连接真实 backend 的浏览器 UI 才是 E2E，代码测试、lint、fixture 和 build wrapper 不是 E2E。
3. OpenAPI、generated artifacts、fixture、handler 和 persistence 必须保持契约一致。
4. 用户可感知的长期业务状态必须由后端持久化。
5. 当前代码、生成物和测试证据优先于历史 PASS 或 completed 状态。
6. 公开接口、数据结构、删除和跨组件架构修改必须先确认。
7. 敏感数据、鉴权、删除、日志和证据产物必须满足安全与隐私边界。
8. 验证必须覆盖与风险直接相关的失败、恢复、幂等和回归负向路径。

## 8 生命周期

新 subject 先创建自洽 Spec；只在 R2/R3 或跨会话需要时创建 Change；只在长期取舍需要时创建 Decision。实现偏离当前 Spec 时修代码和测试；设计改变时先更新 Spec 再创建 Change。Change 完成后冻结，Git 保存详细演进。

Bug 记录只用于可复用根因、严重影响或新失败模式；retrospective 只用于暴露系统性流程问题的交付；工作日志不再作为每次提交的强制投影。

## 9 验收标准

| ID | 标准 |
|----|------|
| A1 | 单个 subject 的 `spec.md` 足以解释当前目标、范围、不变量、失败语义和验收标准 |
| A2 | R1 修改不要求 plan、checklist、BDD、context、INDEX、Bug、retrospective 或工作日志 |
| A3 | R2/R3 一次交付最多使用一个可变 Change，完成后冻结 |
| A4 | 不存在 checked-in `context.yaml` 或等价人工 discovery manifest |
| A5 | 机器索引可从仓库事实重建，绑定当前 commit 并在变化后失效 |
| A6 | 路由依赖精确证据，通用关键词不能产生虚假 high confidence |
| A7 | 引用延迟加载，工具能说明扩展上下文解决的问题 |
| A8 | R2/R3 保留 TDD、契约、持久化、安全、恢复和真实 E2E 边界 |
| A9 | 结构存在性检查和语义正确性检查明确区分 |
| A10 | R1、R2、R2/R3 review 三类任务的首次有效结论前预读量较基线至少下降 50%，缺陷逃逸和误路由不增加 |
| A11 | 代码提交触碰流程文件的中位数由 18.5 至少下降 50% |
| A12 | 公共政策只有一份，Skill 不复制大段政策 |

## 10 失败与恢复

找不到唯一 owner、Spec 与代码冲突、Change 不存在或已冻结、索引 commit 失效、路由只有通用词、风险从 R1 扩张、高风险修改缺确认时必须显式失败。恢复动作只能是刷新索引、加载具体 owner、提升风险、请求判断或创建新 Change，不能复制更多上下文掩盖不确定性。

## 11 用户决策

- 2026-07-17：用户确认采用一次性全量迁移（方案 B），接受同步删除旧文档包装层和旧 Harness 入口的风险；迁移仍按可验证阶段执行，但不保留长期双轨兼容。
