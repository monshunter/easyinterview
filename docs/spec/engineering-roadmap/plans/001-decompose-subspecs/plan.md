# Decompose Subspecs

> **版本**: 2.4
> **状态**: active
> **更新日期**: 2026-05-03

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [engineering-roadmap spec](../../spec.md) §5 列出的 6 层 38 份 child subspec 按 6 个 wave 逐步落地：P0 child 必须完整 spawn 为 spec.md / plan / checklist / context.yaml 并通过 `/plan-review`；P1/P2 child 在 P0 收尾先创建最小 draft spec，占位进入 Wave 6+ 前再补齐 plan 链。本 plan 是 `engineering-roadmap` 的唯一 plan；其他治理类 plan（灰度发布、release gate、隐私链路）归入对应的 child（E4 `release-gate-and-rollout` / F4 `privacy-and-audit-runtime`），不堆在顶层。

## 2 背景

工程拆分一旦在顶层 spec 定下来，就有两种落地方式：

1. **一次性 spawn 全部 38 份 child 空壳**：`docs/spec/` 立刻铺出 38 个目录与 ~95 份 plan。优点是结构清晰，缺点是大量空 spec 缺乏 owner 与决策上下文，会沦为僵尸文档；同时 W1 之前的 6 项 hard gate 还未签字，spec 内容会反复推翻。
2. **按 wave 分批 spawn**：每个 wave 只 spawn 当前 wave 需要的 child，每份 spec 创建时其上游决策已经签字、依赖 child 已经实现，spec 内容可以直接落地不返工。

本 plan 选择第 2 条路线。Phase 1 处理 6 项 W0 hard gate 与 INDEX 占位；Phase 2-6 与 wave W0-W5 一一对应；Phase 7 收尾。

## 3 实施步骤

### Phase 1: 决策与冻结（W0 入口）

#### 1.1 6 项 W0 hard gate 决策

为 [spec §3.2](../../spec.md#32-w0-已锁定决策hard-gate--全部-accepted) 中 Q-1（认证）/ Q-2（异步编排）/ Q-3（分析平台）/ Q-4（云部署）/ Q-5（隐私节奏）/ Q-6（AI 网关与模型路由）每项产出一份 ADR。ADR 文件固定放在 `docs/spec/engineering-roadmap/decisions/ADR-Q{n}-*.md`，通过即视为决策锁定，本 plan 在 spec §3.2 表中同步更新最终结论。**已于 2026-04-26 全部 accepted**：ADR-Q1（自建 passwordless）/ ADR-Q2（Asynq+Redis）/ ADR-Q3（自托管 PostHog，不依赖第三方 Cloud）/ ADR-Q4（Kubernetes）/ ADR-Q5（P0 仅删除，导出延后作为 W0 例外）/ ADR-Q6（AIClient + Model Profile + OpenAI-compatible provider/gateway route + unit-test stub）。

#### 1.2 docs/spec/INDEX.md 占位 38 行

在 `docs/spec/INDEX.md` 中按 Layer A-F × Phase（P0/P1/P2）两轴分组，为 38 份 child subspec 各占一行，状态填 `pending`，链接为占位（不指向真实文件）。本 wave 完成时 INDEX 应有 38 行 + `engineering-roadmap` 主行。

#### 1.3 顶层 spec 自身审查

对 `engineering-roadmap` spec 跑 `/plan-review`；如有反馈在原文件原地修订，不创建 sibling。

L2 remediation 约束：公共 API / DB / event / metrics 中的 `jobType` 沿用技术真理源既有 snake_case 值，内部 Asynq handler 可用 dotted task name，但必须由 C8 / B3 / B4 显式维护映射；Q-5 导出延后必须作为产品验收项例外写入 ADR 与 release gate；Q-3 分析平台按用户决策切换为自托管 PostHog 优先，部署路径由 F2 / E4 在后续 child spec 中验证，不依赖 PostHog Cloud；A2 普通本地 `make dev-up` 默认不启动 PostHog。

### Phase 2: Wave 0（共识与骨架）

#### 2.1 spawn A1 + B1

按 `engineering-roadmap` spec §5.1 / §5.2 描述，spawn `repo-scaffold` 与 `shared-conventions-codified` 两份 child subspec：每份生成 `spec.md` + `history.md` + `plans/` 脚手架 + 至少一个 plan 与 checklist 与 `context.yaml`。`plans/` 脚手架只包含 `INDEX.md`；plan 规则统一引用 `docs/spec/README.md`，plan / checklist / context 模板统一引用 `docs/spec/TEMPLATES.md`，不得在每个 subject 下复制 README 或模板文件。

#### 2.2 W0 收口验证

执行 A1 与 B1 各自 plan 的 checklist；验证两份 child 的 `context.yaml` 可被共享 validator 解析；通过 `docs/spec/INDEX.md` 中 A1 / B1 两行状态由 `pending` 调整为 `active`。`make dev-up` 由 A2 `local-dev-stack` 在 W1 收口验证。

#### 2.3 W0 脚手架结构修复

移除已生成的 `docs/spec/*/plans/README.md` 与 `docs/spec/*/plans/TEMPLATES.md`，并同步 `init-docs` / `create-doc` / `design` skill 与契约测试，防止后续 child subspec spawn 再次复制局部规则或模板。

### Phase 3: Wave 1（基础设施 + 契约骨架）

#### 3.1 并行 spawn 9 份 spec

为 A2 `local-dev-stack` / A3 `ai-gateway-and-model-routing` / A4 `secrets-and-config` / A5 `ci-pipeline-baseline` / B2 `openapi-v1-contract` / B3 `event-and-outbox-contract` / B4 `db-migrations-baseline` / F1 `observability-stack` / F3 `prompt-rubric-registry` 各创建 `spec.md` + `history.md` + plans 脚手架。本 wave **只写 spec，不写 impl plan**——目的是让 9 份契约/基础设施 spec 互相 review 时尽早发现冲突。A5 在当前单人阶段只锁本地质量门禁与远端 CI deferred 边界，不创建 CI pipeline plan。

#### 3.2 W1 collective gate

本阶段的 collective gate 是 parent-level cross-spec review：集中核对 9 份 W1 spec 之间的职责边界、真理源引用、ADR-Q1..Q6 继承关系和后续 handoff，不声称 9 个 child 已各自拥有可被 `/plan-review` 解析的 plan/context。9 个 child 在本阶段只创建 `spec.md` + `history.md` + `plans/INDEX.md`，独立 impl plan 必须在逐一核对对应 spec 后再创建；不得批量预生成还未审清的 child plan。

#### 3.3 B2 OpenAPI spec-contract lock

B2 `openapi-v1-contract` 在本阶段只锁定当时的 spec 合同：v1.0.0 freeze 清单、字段命名、additive-only 规则、privacy export P0 返回 501 例外与 fixtures 同源边界。该历史 lock 后续已经由 B2 `001` / `002` / `003` 原地 remediation 按 product-scope v1.2 收敛；当前可执行 OpenAPI 真理源是 12 tag / 34 operation，含 `DELETE /api/v1/me` 账号删除别名，不再包含独立 `Mistakes` / `Growth`。

#### 3.4 F1 observability spec-contract lock

F1 `observability-stack` 在本阶段只锁定 baseline metric 名称、allowed labels、forbidden labels、log 明文红线、dashboard 名称与健康检查契约。OTel/logx helper、lint-metrics、dashboard JSON 与 alerting rules 的真实落地由 F1 自身后续 `001` plan 承接。

#### 3.5 F3 prompt/rubric spec-contract lock

F3 `prompt-rubric-registry` 在本阶段只锁定 13 个 P0 feature_key 字典、`(feature_key, version, language)` 坐标、Resolve 调用契约、prompt/rubric 文件落点与 lint 边界。`config/prompts/`、`config/rubrics/`、Resolve 实现与 baseline prompt 文件由 F3 自身后续 `001` plan 承接；W2 业务域在 F3 `001` 未通过前不得 hardcode prompt 文本。

#### 3.6 A2 local dev stack spec-contract lock

A2 `local-dev-stack` 在本阶段只锁定最小本地依赖（Postgres+pgvector / Redis / MinIO）、项目组件启动语义、端口、卷、`make dev-up` / `make dev-doctor` / `make dev-down` / `make dev-reset` / `make dev-logs` 的行为契约与 JSON 健康检查口径。`deploy/dev-stack/docker-compose.yaml` 与 Make target 真实实现由 A2 自身后续 `001` plan 承接；依赖本地栈的 W2 implementation 在 A2 `001` 未通过前不得启动。

#### 3.7 A5 local quality gate spec-contract lock

A5 `ci-pipeline-baseline` 在本阶段只锁定本地手动质量门禁（`make lint` / `make test` / `make build` / `make docs-check` / `make codegen-check`）与远端 CI 延后条件。当前个人单人开发阶段不得把 GitHub Actions、branch protection、required checks、artifact、nightly 或 CI secret 写成 P0 前置；如未来触发多人协作 / 公开 release / 自动发版，再由 A5 原地新增远端 CI plan。

### Phase 4: Wave 2（前后端 mock-first 并行）

#### 4.1 spawn W2 child

后端 5 份：C1 `backend-auth` / C2 `backend-upload` / C3 `backend-profile` / C8 `backend-async-runtime` / E1 `mock-contract-suite` 各自创建 spec + 完整 plan 链；E1 是把 B2 fixtures 转成可运行 mock server 的统一壳。
前端 6 份：D1 `frontend-shell` / D2 `frontend-home-job-picks-and-parse` / D3 `frontend-workspace-and-practice` / D4 `frontend-report-dashboard` / D5 `frontend-resume-workshop` / D6 `frontend-debrief`；D1 必须先于 D2-D6 完成基础壳。
横切 1 份：F2 `analytics-funnel`。

#### 4.2 W2 collective gate

E1 提供当前 B2 12 tag / 34 operation 全 mock（按 B2 fixtures 自动生成）；前端 6 域跑通 P0 happy path（导入→规划→练习→报告→复练当前轮 / 下一轮→真实复盘，且简历绑定可用）全部基于 E1 mock；后端 5 域 mock-server plan 自验证；前后端 mock 同源（同一份 fixtures，禁止前端 hardcode）。

### Phase 5: Wave 3（核心业务域后端）

#### 5.1 spawn C4-C7

为 C4 `backend-targetjob` / C5 `backend-practice` / C6 `backend-review` / C7 `backend-resume` / C9 `backend-debrief` 各创建 spec + 完整 plan 链。`backend-practice` 内部 plan 必须显式写出 turn-light-review 边界（同步轻量观察 vs 异步完整报告 vs 跨 C5/C6 边界）；`backend-debrief` 在 P0 只承接真实面试复现 / 复盘文本流，感谢信草稿与完整跟进建议延后到 C9 P1 plan。

#### 5.2 F3 接入真实 Model Profile

`prompt-rubric-registry` 此时切到真实 Model Profile（由配置映射到真实 AI provider / gateway endpoint、provider、model），落地至少 50 题的离线评估集（覆盖行为题、动机题、会话内追问、候选人反问建议、不同语言）；F3 所有 child（C4/C5/C6/C7）通过 `prompt_version + rubric_version + model_profile` 引用。

#### 5.3 W3 collective gate

7 个 P0 后端域（C1+C4-C7+C8+C9 共 7 域；C2/C3 已在 W2 完成）通过各自的 unit 测试与 mock-server BDD（每域内部 plan 的 BDD-Gate）。

### Phase 6: Wave 4 + Wave 5（真集成 + 上线 gate）

#### 6.1 spawn E2 + E4

E2 `e2e-scenarios-p0` 创建跨前后端的 P0 主漏斗 BDD；E4 `release-gate-and-rollout` 创建灰度开关 / 版本兼容 / 回滚 runbook / SLO 准入。

#### 6.2 D2-D6 切真后端

每份 P0 前端 child 的 `003-integration` plan 把 fetch 从 E1 mock 切到真后端（W3 跑通的服务）。F1 `observability-stack` 此时把指标接齐 5 个 dashboard；F2 `analytics-funnel` 完成漏斗对账。

#### 6.3 W4 + W5 collective gate

E2 全场景通过；`04-metrics-observability.md` §15 最低上线门槛全勾；E4 staging 灰度演练 + 回滚演练通过；Q-5 ADR 决定的 P0 隐私范围完成验证（若 Q-5 选择 P0 完整导出+删除，则 C12/F4 必须在 W4 前升格并通过各自 gate）；P0 准入。

### Phase 7: 收尾

#### 7.1 状态切换

`engineering-roadmap` spec 状态由 `active` 调整为 `completed`（仅当 P0 全部上线）。本 plan 状态保持 `active` 直到 P1 child 全部 spawn。

#### 7.2 P1 / P2 child draft

为 C10 `backend-readiness-signals`、C11 `backend-retrieval`、C12 `backend-privacy`、E3 `e2e-scenarios-p1`、F4 `privacy-and-audit-runtime` 创建 draft spec.md（只含 §1 §2 §3 §7，标 `状态: draft`）；P1 前端增强不另建恢复旧模块的 child，后续只允许挂到 D5/D6 等已保留 P0 前端 child 的新 plan。为 C13/C14（P2）、D7 `frontend-voice-production`（P2）也创建同等最小 draft spec.md，避免 P2 只停留在 INDEX 占位。

#### 7.3 交付复盘

触发 `/retrospective` 生成 P0 交付复盘报告，固化本次 wave 编排的经验教训。

## 4 验收标准

- 本 plan 7 个 phase 的所有 checklist 项全部勾选。
- 本 plan 关联的 6 个 wave 同步点（W0-W5 collective gate）全部通过。
- Phase 3 的完成口径是 W1 spec-contract lock 与 cross-spec review 通过；A2/B2/F1/F3 的可执行 gate 由各自 child `001` plan 验证，不在 parent checklist 中冒充完成。
- `engineering-roadmap` spec §6 表中 C-1 至 C-7 的 7 条验收场景全部成立。
- `docs/spec/INDEX.md` 中 38 行 child 全部不为 `pending`（P0 row 至少 `active`、P1/P2 row 至少 `draft` 占位）。

## 5 风险与应对

| 风险 | 应对措施 |
|------|----------|
| **B2 OpenAPI v1.0.0 freeze 后被频繁破坏性变更** | B2 plan 必须自带 breaking change linter；任何破坏性变更必须开 ADR 走 B2 spec 修订流程；P0 中后期默认只允许 additive |
| **C5 `backend-practice` 与 C6 `backend-review` 边界模糊**（turn-light-review） | C5 plan 必须开一个 `turn-light-review` plan 专门讲清楚同步轻量观察 vs 异步完整报告的边界；如果 W3 末发现体量超阈值，把 lightweight-observer 拆为 `backend-review/006-lightweight-observer`，**禁止塞进 C5** |
| **D3 体量过大**（Mock Interview Plan + 完整 Interview Session + 文本 / 语音形式 + 辅助程度 + company intel 集中在一份 spec） | D3 内部 plan 拆分为 workspace / session-core / assistance-and-strict / company-intel，每个 plan 单独 review、单独 PR；W2 优先完成 workspace + session-core，语音生产化只保留折返点与 feature flag，真实 STT / 媒体留存归 D7 / C14 P2 |
| **F3 `prompt-rubric-registry` 没有 baseline 时 W2 业务域偷偷 hardcode prompt** | W1 parent phase 必须先锁 `feature_key + version` 契约；W2 业务域 implementation 只能在 F3 child `001` 验证 baseline prompt / rubric 文件后引用 F3 prompt id，任何业务 spec/plan 不得 hardcode prompt 文本 |
| **6 项 W0 待决策方案未签字就进入 W1** | Phase 1.1 设为 hard gate；任一 ADR 未签字时 W1 不得开始；具体由 Phase 2 收口的 `/plan-review` 强制执行 |
| **自托管 PostHog 运维复杂度超过 P0 带宽** | ADR-Q3 锁定“不依赖第三方 Cloud”，但不锁已废弃的 K8s Helm chart；F2 / E4 必须先验证可运维 self-host path、备份、升级与漏斗对账，再允许 W4 release gate 通过；A2 普通本地栈只提供 no-op / file-backed dev mode |
