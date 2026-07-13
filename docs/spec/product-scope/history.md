# Product Scope History

> **版本**: 2.20
> **状态**: active
> **更新日期**: 2026-07-13

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-13 | 2.20 | 用户确认方案 A：Practice 发送采用即时 user row、pending thinking/输入锁、failed-row retry，并由服务端持久化原 message identity/reply state 保障刷新恢复。 | backend-practice/002 + frontend-workspace-and-practice/002 + openapi-v1-contract/001 |
| 2026-07-13 | 2.19 | Home JD intake 只保留粘贴文本框和 `{ rawText, targetLanguage, resumeId }` 请求合同；删除其他 JD 导入形态，保留 Resume 上传能力。 | frontend-home-job-picks-and-parse/001 + contract/backend owners |
| 2026-07-12 | 2.18 | 删除首次 Parse 页已不存在的“仅保存规划”分叉；确认一次提交同时持久化规划并直接进入 Practice，回访仍从既有入口承接。 | frontend-workspace-and-practice/001 + UI module-job-workspace |
| 2026-07-12 | 2.17 | 明确零回答不可完成、空 focus 通用同轮复练和 reportId-only 缺失态，修正报告 locator 漂移。 | backend-practice/002 + backend-practice/004 + frontend-workspace-and-practice/002 + frontend-report-dashboard/001 |
| 2026-07-12 | 2.16 | 报告统一为 reportId-only 深链，状态/上下文/CTA identity 来自冻结后端投影；复练 focus 明确为单份报告内 dimension code。 | backend-review/001 + frontend-report-dashboard/001 |
| 2026-07-12 | 2.15 | 报告采用完成时冻结上下文与 LLM direct semantic output；UI 锁定三指标四常驻区块，grounding evidence、服务端 replay focus 与诚实生成态成为 P0 质量红线。 | backend-review 001 / frontend-report-dashboard 001 |
| 2026-07-12 | 2.14 | 模拟面试收敛为不限制问题范畴的连续文本会话；删除题目结构与专用 hint/mode，电话模式暂时禁用，报告改为会话级。 | backend-practice 001-003 / frontend-workspace 002 / voice 001 / review 001 |
| 2026-07-10 | 2.13 | 统一 product-scope 中文范围边界与执行 plan 的 scope-boundary gate 口径。 | tech-debt pruning |
| 2026-07-10 | 2.12 | 将 privacy export、pending child 与设置通知/订阅边界收敛为不可用例外 / 索引模型 / 范围外能力口径。 | tech-debt pruning |
| 2026-07-10 | 2.11 | 用户可见面试形式统一为文本面试 / 电话模式；`voice` 只保留为底层工程能力名或负向 route 输入。 | tech-debt pruning |
| 2026-07-10 | 2.10 | 电话模式正向入口收敛为 `practice?mode=phone&modality=phone`；旧 `voice` query 不再作为电话模式入口。 | tech-debt pruning |
| 2026-07-07 | 2.8 | 简历资产范围收敛：详情页只读展示简历正文，移除详情二次编辑/导出/复制/原件弹层；新增 LLM-derived 简历名称要求。 | resume detail readonly / LLM display name follow-up |
| 2026-07-07 | 2.6 | 将 product-scope 正文收敛为当前合同表达；中文范围边界只描述当前行为和范围外。 | product-scope/001-core-loop-module-pruning |
| 2026-07-06 | 2.5 | 将 active product-scope 中的范围变更过程说明收敛为当前范围合同与负向边界表述。 | product-scope/001-core-loop-module-pruning |
| 2026-07-06 | 2.4 | 将 UI 范围边界证据集中到 `module-map` 与当前 product-scope。 | product-scope/001-core-loop-module-pruning |
| 2026-07-06 | 2.3 | 将产品真理源验收改为 product-scope active；同步 UI 文档和执行 subject 入口。 | product-scope/001-core-loop-module-pruning |
| 2026-06-29 | 2.2 | 用户确认方案 B：P0 收敛为 JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮；新增 D-22、C-19 和 `001-core-loop-module-pruning` owner plan。 | product-scope/001-core-loop-module-pruning |
| 2026-05-05 | 1.8 | owner matrix 的 AI provider / model profile 可执行真理源改为 `config/ai-providers.yaml` 与 `config/ai-profiles.yaml`。 | ai-provider-and-model-routing/003 L2 remediation |
| 2026-05-05 | 1.7 | 明确技术契约 owner matrix 是当前唯一分层入口：API / DB / event / metrics / logging / config / AI 等责任由当前 owner spec 与编码 truth source 独立承接。 | engineering-roadmap/001-decompose-subspecs |
| 2026-05-05 | 1.6 | 增加当前技术契约统一 owner matrix，把 API / DB / event / metrics / logging 等职责映射到当前 A/B/F owner spec 与编码 truth source。 | docs-only |
| 2026-05-03 | 1.5 | 同步 engineering-roadmap v3.0：`docs/spec/INDEX.md` 只投影真实 active spec，后续 child 按当前产品 / UI 能力 on-demand 创建。 | engineering-roadmap/001-decompose-subspecs |
| 2026-05-03 | 1.4 | 明确当前 API / DB / event / metrics 以 Layer B/F active spec 与已编码 truth source 为准。 | docs-only |
| 2026-05-03 | 1.3 | 明确语音面试只通过 `practice` 显式携带 `mode=voice` / `modality=voice` 进入。 | docs-only |
| 2026-05-03 | 1.2 | 记录 engineering-roadmap v2.2 产品范围对齐：P0 前端 child 改为当前 UI 五入口相关拆分。 | engineering-roadmap/001-decompose-subspecs |
| 2026-05-03 | 1.1 | 修正范围语义：当前 UI / UI 文档和本 spec 是正向范围清单；仅保留 `全球多平台搜岗` 作为规划例外。 | docs-only |
| 2026-05-03 | 1.0 | 初始创建：将产品范围真理源迁入 spec-centric 文档体系，吸收当前 `ui-design/` 与 `docs/ui-design/` 已确认的导航、模块边界、报告、简历和当前范围，并保留目标用户、阶段路线、隐私伦理与质量评估水准。 | docs-only |
