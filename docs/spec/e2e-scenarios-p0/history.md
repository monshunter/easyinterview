# E2E Scenarios P0 History

> **版本**: 2.12
> **状态**: active
> **更新日期**: 2026-07-13

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-13 | 2.12 | 区分报告产品验收与严格稳定性诊断：机械合同必须100%；固定五类代表场景至少4/5表达约80%语义置信度；P0.100的11/11、关键3/3与blind review继续严格fail-closed。最终run59381为机械9/9、语义8/9、场景4/5，产品验收满足但严格P0.100保持FAIL。 | [002](./plans/002-manual-uat-real-provider-full-funnel/plan.md) + F3/004 |
| 2026-07-13 | 2.11 | Supersede 2.9/2.10中的product durable/report-job max4：每次`GenerateReport` invocation独立initial+3、等待10s/20s/40s、返回销毁且新动作清零；async attempts只作基础设施，lease side-effect fencing继续有效。 | [002](./plans/002-manual-uat-real-provider-full-funnel/plan.md) + backend-review/001 |
| 2026-07-13 | 2.10 | 补report job max4、lease fencing、frontend in-flight resume gate；run35622于7/11中止非PASS。 | [002](./plans/002-manual-uat-real-provider-full-funnel/plan.md) |
| 2026-07-13 | 2.9 | 用户确认generation/judge各自max4 calls；P0.100新增transient recovery、multi-round invalid、attempt4 terminal、nonretryable、judge invalid/valid-negative与crash/replay cap矩阵。 | [002](./plans/002-manual-uat-real-provider-full-funnel/plan.md) + backend-review/001 + F3/004 |
| 2026-07-13 | 2.9 | `e2e-p0-100-20260713T022140Z-25849` 因合同替换在10/11主动中止，不计PASS或marker。 | [002](./plans/002-manual-uat-real-provider-full-funnel/plan.md) |
| 2026-07-13 | 2.8 | 80338 attempt11暴露evalkit未复用产品完整validator，needs_practice retry+next进入judge；合同改为sole-label targeted/其它或mixed whole-report repair，全阶段复验。Code GREEN待主agent验收。 | [002](./plans/002-manual-uat-real-provider-full-funnel/plan.md) + backend-review/001 + F3/004 |
| 2026-07-13 | 2.8 | 59906 injection修复summary clause evidence mapping、action未声明质量属性升级与W exact readiness，judge3x PASS；75753 short empty-focus修复exact generic exception，同digest+5次PASS；最终矩阵/markers仍pending。 | [002](./plans/002-manual-uat-real-provider-full-funnel/plan.md) + F3/004 |
| 2026-07-13 | 2.7 | P0.100 run `e2e-p0-100-20260713T011140Z-36625` short-conservative attempt1因exact generic replay与report_action_quality rubric冲突在`$.nextActions[0]`得到invalid_partial；TDD修复并同步migration/active DB，同case复测0.82/0.70零违规；完整矩阵仍pending。 | [002](./plans/002-manual-uat-real-provider-full-funnel/plan.md) + F3/004 |
| 2026-07-13 | 2.7 | 方案 A 最终边界：wire/schema fuse200 code points；semantic/UX English24 whitespace words / zh-CN64 Unicode code points；targeted repair内部目标18/52并按200+24/64复验；P0.099/P0.100均等待当前合同新证据。 | [001](./plans/001-full-funnel-happy-journey/plan.md) + [002](./plans/002-manual-uat-real-provider-full-funnel/plan.md) |
| 2026-07-13 | 2.6 | label>120 仍属纯 action-label violation，必须 action_labels 而非 whole_report；本轮 P0.100 live 因误分类 FAIL，5类/11次与 markers 均保持 pending。 | [002](./plans/002-manual-uat-real-provider-full-funnel/plan.md) |
| 2026-07-13 | 2.5 | P0.100 建立 evalkit scoped repair、runner no-repair/zero-judge 与 one-shot judge 分界；具体 label>120 scope由2.6校正。 | [002](./plans/002-manual-uat-real-provider-full-funnel/plan.md) |
| 2026-07-13 | 2.4 | P0.100 独立证明 5 类/11 次内容可靠性；P0.099 以 current-run canonical audit 闭环，exact 14/40 由 deterministic parity 证明。 | [001](./plans/001-full-funnel-happy-journey/plan.md) + [002](./plans/002-manual-uat-real-provider-full-funnel/plan.md) |
| 2026-07-13 | 2.3 | P0.100 generation completion 新增同源 output-schema validation、一次 `$ / output_schema_invalid` repair、second-invalid fail、usage/latency 聚合与 repair_used；judge 零 repair/retry；action 固化 en<=14 words、zh-CN<=40 chars、multi-focus 逐 code 分号短片段。 | [002](./plans/002-manual-uat-real-provider-full-funnel/plan.md) Phase 8 |
| 2026-07-12 | 2.2 | 将新增 P0.100 Phase/BDD/内容可靠性工作完整迁回真实 owner plan 002；plan 001 只保留 P0.099 精确六图与 manifest。 | [001](./plans/001-full-funnel-happy-journey/plan.md) Phase 5 + [002](./plans/002-manual-uat-real-provider-full-funnel/plan.md) Phase 8 |
| 2026-07-12 | 2.1 | P0.099/100 补 frozen direct report、精确六图矩阵与 fact→judgment→action 内容可靠性验收。 | [001-full-funnel-happy-journey](./plans/001-full-funnel-happy-journey/plan.md) Phase 5 |
| 2026-07-12 | 2.0 | Current funnel rebased to continuous conversation and shared real-browser acceptance. | [001-full-funnel-happy-journey](./plans/001-full-funnel-happy-journey/plan.md) |
| 2026-07-10 | 1.9 | 统一 P0 journey 的范围外负向 gate 口径，清理混合范围术语，并同步 001 plan、BDD 文档与 context 版本。 | tech-debt pruning |
| 2026-07-06 | 1.8 | 对齐 product-scope D-17/D-22 后的 P0 journey 范围：full-funnel 正向目标删除真实复盘 / debrief 回流、jobs-recommendations 与 profile 业务域；复练仍以 `retry_current_round` / `next_round` 报告派生计划表达，`debrief` 仅保留在 out-of-scope-negative 反向 gate 中。 | product-scope/001-core-loop-module-pruning Phase 6 |
| 2026-05-27 | 1.7 | 对齐 C1/D1 email-code 修订：真实 provider hybrid UAT 账号入口改为 Mailpit 6 位 code，证据红线禁止 raw email code。 | backend-auth/001 Phase 7 + frontend-shell/001 Phase 8 |
| 2026-05-27 | 1.6 | 修正真实 provider hybrid 场景 env 边界：`deploy/dev-stack/.env` 成为本地真实联调唯一 env 来源，删除 `E2E.P0.100` 场景专属 `dev-real.env` 模板口径。 | 002-manual-uat-real-provider-full-funnel |
| 2026-05-27 | 1.5 | 将 `E2E.P0.100` 从独立 `manual-uat` companion 迁回标准 `e2e` 场景框架：AI Agent 先运行环境 preflight 与四段脚本，缺真实凭证/浏览器证据时输出 `MANUAL_REQUIRED`，人工或浏览器 Agent 在同一场景输出目录补证。 | 002-manual-uat-real-provider-full-funnel |
| 2026-05-26 | 1.4 | 对齐 local-dev-stack Mailpit revision：manual UAT 账号入口改为 synthetic 邮箱 + Mailpit email-code，删除直接 session bootstrap 口径；继续保留 `test/scenarios` 只允许 shell/Python、不得新增 `backend/cmd` / Go helper 的边界。 | 002-manual-uat-real-provider-full-funnel + local-dev-stack/001 |
| 2026-05-26 | 1.3 | 修正 002 manual UAT 账号边界：本地栈账号入口由 Mailpit email-code 承接，账号/session 辅助不得进入正式 `backend/cmd` / Go helper。 | 002-manual-uat-real-provider-full-funnel |
| 2026-05-26 | 1.2 | 新增真实 provider manual UAT 验收层：锁定 D-8~D-10，明确 002 不改写 001 的 stub-AI 自动化边界；新增 C-9~C-13，要求账号/session bootstrap、真实前后端、真实 AI provider、无 mock/stub 冒充、脱敏证据与完整材料包。 | 002-manual-uat-real-provider-full-funnel |
| 2026-05-24 | 1.1 | L1 plan-review 修订：校正 P0 场景实施前基线为 87 条切片场景（最高编号 `E2E.P0.097`），将 operation matrix 口径统一为 9 行（8 个主链必经 operation + `getJob` 备选轮询 / handler gate），明确 Playwright 全栈必须用 `VITE_EI_API_MODE=real` / `VITE_EI_API_BASE_URL` 指向真后端，并把 out-of-scope-negative 加固为 route-aware 范围外 route / 独立 voice / `mode=debrief` 反查且避免误伤合法 `createPracticePlan` / `resumeAssetId`。 | 001-full-funnel-happy-journey |
| 2026-05-24 | 1.0 | 初始创建：定义 P0 完整漏斗端到端 journey owner subject；锁定 D-1~D-7（真后端全栈 + stub AI + happy 主干 + 两种 driver + 接续编号）；派生 `001-full-funnel-happy-journey`（`E2E.P0.098` API-level + `E2E.P0.099` Playwright 全栈） | 001-full-funnel-happy-journey |
