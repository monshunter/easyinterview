# Decompose Subspecs Checklist

> **版本**: 1.7
> **状态**: active
> **更新日期**: 2026-04-27

**关联计划**: [plan](./plan.md)

## Phase 1: 决策与冻结（W0 入口）

- [x] 1.1 6 项 W0 hard gate 决策完成（Q-1 认证 / Q-2 异步编排 / Q-3 分析平台 / Q-4 云部署 / Q-5 隐私节奏 / Q-6 AI 网关与模型路由），每项在 `docs/spec/engineering-roadmap/decisions/ADR-Q{n}-*.md` 产出一份 ADR
- [x] 1.2 spec §3.2 表中 6 项决策的最终结论同步更新（ADR 通过后回填）
- [x] 1.3 `docs/spec/INDEX.md` 中 38 行 child subspec 占位行齐全（按 Layer A-F × Phase P0/P1/P2 两轴分组，状态 `pending`，链接占位）
- [x] 1.4 顶层 `engineering-roadmap` spec 通过 `/plan-review`，反馈原地修订完毕
- [x] 1.5 验证 `docs/spec/INDEX.md` 与 `engineering-roadmap/spec.md` Header 一致（运行 `/sync-doc-index --check`）
- [x] 1.6 L2 remediation：修复 async `jobType` 命名漂移、Q-5 P0 导出例外记录、ADR-Q3 自托管 PostHog 决策切换，并同步 spec / plan / history / context / INDEX

## Phase 2: Wave 0（共识与骨架）

- [x] 2.1 spawn `repo-scaffold`：spec.md + history.md + plans 脚手架 + 至少 1 个 plan / checklist / context.yaml
- [x] 2.2 spawn `shared-conventions-codified`：spec.md + history.md + plans 脚手架 + 至少 1 个 plan
- [x] 2.3 W0 收口验证：A1 / B1 docs 结构完整（spec + history + plans/INDEX.md + 001-bootstrap 三件套），plan 规则统一引用 `docs/spec/README.md`，plan / checklist / context 模板统一引用 `docs/spec/TEMPLATES.md`；`validate_context.py --target docs` 对 A1 / B1 / engineering-roadmap 三个 context.yaml 全部通过；A1 / B1 bootstrap 实现（make 占位 / Go module / TS lib / generator）随后由各自 child 的 `/implement` 继续推进，W1 末闭合 spec C-2；`make dev-up` 延后到 A2/W1 gate
- [x] 2.4 `docs/spec/INDEX.md` 中 A1 / B1 两行由占位切换为指向真实 `spec.md` 的链接，状态 `active`、版本 1.0、更新日期 2026-04-26
- [x] 2.5 W0 脚手架结构修复：移除已生成的 `docs/spec/*/plans/README.md` 与 `docs/spec/*/plans/TEMPLATES.md`，并同步 `init-docs` / `create-doc` / `design` skill 与契约测试，防止后续 child subspec spawn 再次复制局部规则或模板

## Phase 3: Wave 1（基础设施 + 契约骨架）

- [x] 3.1 并行 spawn 9 份 spec（A2 / A3 / A4 / A5 / B2 / B3 / B4 / F1 / F3），仅写 spec.md + history，**不写 impl plan**
- [x] 3.2 完成 parent-level W1 cross-spec review：核对 9 份 W1 spec 的 boundary / ownership / ADR-Q1..Q6 继承 / truth-source 引用；本项不声称 9 个 child 已各自拥有独立 plan/context 或已逐个通过 `/plan-review`，child impl plan 必须在逐一核对对应 spec 后再创建
- [x] 3.3 B2 `openapi-v1-contract` 完成 spec-contract lock：v1.0.0 freeze 的 36 endpoint / 14 tag / additive-only 规则 / privacy export 501 例外已写入 spec；`openapi/openapi.yaml`、codegen、fixtures、breaking-change linter 由 B2 后续 `001` plan 验证
- [x] 3.4 F1 `observability-stack` 完成 spec-contract lock：baseline metric 命名、allowed labels、forbidden labels、log 明文红线、dashboard 名称与健康检查契约已写入 spec；helper / lint / dashboard / alerting 实现由 F1 后续 `001` plan 验证
- [x] 3.5 F3 `prompt-rubric-registry` 完成 spec-contract lock：13 个 P0 feature_key、`(feature_key, version, language)` 坐标、Resolve 调用契约、prompt/rubric 文件落点已写入 spec；baseline prompt/rubric 文件与 loader 由 F3 后续 `001` plan 验证
- [x] 3.6 A2 `local-dev-stack` 完成 spec-contract lock：7 个本地服务、`make dev-*` 行为契约、JSON 健康检查口径已写入 spec；`deploy/dev-stack/docker-compose.yaml` 与真实 `make dev-up` 一键健康检查由 A2 后续 `001` plan 验证

## Phase 4: Wave 2（前后端 mock-first 并行）

- [ ] 4.1 spawn 后端 5 份：C1 / C2 / C3 / C8 / E1（每份完整 spec + plan 链）
- [ ] 4.2 spawn 前端 4 份：D1 / D2 / D3 / D4；D1 必须先于 D2-D4 完成基础壳
- [ ] 4.3 spawn 横切 1 份：F2 `analytics-funnel`
- [ ] 4.4 E1 提供 14 tag 全 mock（按 B2 fixtures 自动生成）
- [ ] 4.5 前端 4 域跑通 P0 8 步 happy path（导入→工作台→练习→报告→错题→复练）全部基于 E1 mock
- [ ] 4.6 后端 5 域 mock-server plan 自验证通过
- [ ] 4.7 验证：前后端 mock 同源（fixtures 同一份，禁止前端 hardcode）

## Phase 5: Wave 3（核心业务域后端）

- [ ] 5.1 spawn C4 `backend-targetjob`：完整 spec + plan 链
- [ ] 5.2 spawn C5 `backend-practice`：完整 spec + plan 链；plan 必须显式写出 turn-light-review 边界
- [ ] 5.3 spawn C6 `backend-review`：完整 spec + plan 链
- [ ] 5.4 spawn C7 `backend-resume`：完整 spec + plan 链
- [ ] 5.5 F3 切到真实 Model Profile（由运维配置映射到 AI Gateway route / provider / model），落地 ≥50 题离线评估集
- [ ] 5.6 6 个 P0 后端域（C1 + C4-C7 + C8）通过各自 unit 测试 + mock-server BDD

## Phase 6: Wave 4 + Wave 5（真集成 + 上线 gate）

- [ ] 6.1 spawn E2 `e2e-scenarios-p0`：跨前后端 P0 主漏斗 BDD
- [ ] 6.2 spawn E4 `release-gate-and-rollout`：灰度 / 版本兼容 / 回滚 runbook / SLO 准入
- [ ] 6.3 D2 / D3 / D4 各自的 `003-integration` plan 把 fetch 切到真后端
- [ ] 6.4 F1 指标接齐 5 个 dashboard；F2 漏斗对账完成
- [ ] 6.5 E2 全场景通过
- [ ] 6.6 `04-metrics-observability.md` §15 最低上线门槛全勾
- [ ] 6.7 E4 staging 灰度演练 + 回滚演练通过
- [ ] 6.8 Q-5 ADR 决定的 P0 隐私范围完成验证；若 Q-5 选择 P0 完整导出+删除，C12/F4 已在 W4 前升格并通过各自 gate
- [ ] 6.9 P0 准入

## Phase 7: 收尾

- [ ] 7.1 `engineering-roadmap` spec 状态由 `active` 调整为 `completed`（仅当 P0 全部上线）
- [ ] 7.2 P1 child draft spec 创建：C9 / C10 / C11 / C12 / D5 / D6 / E3 / F4（每份 spec.md 含 §1 §2 §3 §7，状态 `draft`）
- [ ] 7.3 P2 child draft spec 创建：C13 / C14 / D7（每份 spec.md 含 §1 §2 §3 §7，状态 `draft`），INDEX 行从占位切为真实链接
- [ ] 7.4 触发 `/retrospective` 生成 P0 交付复盘报告
- [ ] 7.5 同步 `docs/work-journal/INDEX.md` 与最近一条工作日志，记录 P0 收尾
