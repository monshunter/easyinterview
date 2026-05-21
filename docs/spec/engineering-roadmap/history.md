# Engineering Roadmap History

> **版本**: 3.16
> **状态**: active
> **更新日期**: 2026-05-21

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-21 | 3.16 | 对齐当前 `docs/spec/INDEX.md` 与 plan checklist 事实：刷新 §5.2 P0 workstream 状态，移除已创建 subject 的“未创建”旧口径，并保留 backend-async-runner / e2e-scenarios-p0 / analytics-funnel / release-gate-and-rollout 为未创建候选。 | roadmap/spec lifecycle rebaseline |
| 2026-05-17 | 3.15 | 对齐 B2 D-23 与 recent additive contract 落地：roadmap B2 摘要与 mock-first runway 从 55 operation 升到 59 operation，承接 Debrief suggestions、PracticeSessions listing、Practice voice turn 与 backend-resume `confirmResumeStructuredMaster` fixtures。 | backend-resume/002-versions-tailor-runs-and-save-v1 Phase 1 |
| 2026-05-12 | 3.14 | 对齐 backend-upload/001 落地：Resume Workshop §5.2 中 `backend-upload` 状态从首批 plan 骨架待实施更新为 file_objects + presign baseline 已落地，backend-resume 可接续。 | backend-upload/001-file-objects-and-presign-baseline |
| 2026-05-12 | 3.13 | 对齐 B2 D-18 落地：roadmap B2 摘要与 mock-first runway 从 46 operation 升到 55 operation，承接 Resume Workshop `Resumes` tag 新增 9 operation 与 fixtures。 | openapi-v1-contract/004-resume-additive-coverage |
| 2026-05-12 | 3.12 | 对齐 B4 D-17 落地：roadmap B4 摘要从旧应用表数量升级为 28 应用表 + auth / migration 支撑表，承接 Resume Workshop `resume_versions` / `resume_version_suggestions` migration。 | db-migrations-baseline/002-resume-versions-additive |
| 2026-05-11 | 3.11 | Resume Workshop 阶段 1 三个新 subspec 骨架就位：`backend-upload` (C2，001-file-objects-and-presign-baseline) / `backend-resume` (C7，001-asset-register-parse-and-listing) / `frontend-resume-workshop` (001-listing-routing-and-detail-readonly) 三个 subspec 的 spec.md / history.md / plans/INDEX.md / 第一批 plan（含 BDD）全套骨架就位；§5.2 Resume Workshop 行当前状态从 "未创建" 改为各 subspec "active（首批 plan 骨架就位，待 `/implement` 执行）"。 | docs-only（Resume Workshop 阶段 1 设计结晶） |
| 2026-05-11 | 3.10 | §5.2 Resume Workshop 行依赖 C2 显式化为 `C2 backend-upload`：基于 `shared/jobs.yaml` C-domain 占用映射与历史 work-journal 记录，消除 C2 占位歧义，为 Resume Workshop 阶段 0 contract additive 与阶段 1 `backend-upload` subspec 创建提供明确 ownerDomain。 | docs-only（Resume Workshop 阶段 0 设计结晶） |
| 2026-05-08 | 3.8 | 对齐 A3/B4/A2 当时决策：DeepSeek V4 Flash/Pro 成为开发期 chat provider 主力；默认本地和云基础设施不再保留向量扩展或资料检索预埋；B4 baseline 收敛为应用表 + 3 auth 支撑表 + 2 迁移元数据表。 | ai-provider-and-model-routing/003 Phase 6 |
| 2026-05-06 | 3.7 | 新增 `backend-runtime-topology` active owner，Q-2/S2 从独立 worker runtime 调整为 B3 job/outbox contract + backend internal runner，P0 不再要求 `backend-async-runtime` 独立进程。 | backend-runtime-topology/001-worker-consolidation |
| 2026-05-05 | 3.6 | 按 S1/S2 默认路径派生首批 P0 implementation owner：`mock-contract-suite`、`frontend-shell`、`backend-auth`，并创建 `test/scenarios/e2e` 场景框架；§5.2 增加当前状态列，避免已创建 subject 仍被误读为 pending 候选。 | mock-contract-suite/001 + frontend-shell/001 + backend-auth/001 |
| 2026-05-05 | 3.5 | 同步 A3 003 Phase 4：roadmap 的 Q-6 / A3 职责改为 Provider Registry + Capability Model Profile + profile coverage lint，不再沿用单一 endpoint 作为当前目标架构。 | ai-provider-and-model-routing/003 Phase 4 |
| 2026-05-05 | 3.4 | 将已迁移技术草稿从当前项目实体与命名空间中移除：删除旧目录，要求当前文档、代码注释、生成源和生成物只引用 product-scope §1.5、owner spec 与编码 truth source。 | 001-decompose-subspecs |
| 2026-05-05 | 3.3 | 增加已迁移技术草稿删除前 gate：旧名称不得再作为 Markdown 链接或外部真理源，当前替代必须由 product-scope §1.5、owner spec 与编码 truth source 独立承载。 | 001-decompose-subspecs |
| 2026-05-05 | 3.2 | 将技术契约替代表述指向 product-scope §1.5 owner matrix，roadmap 只消费统一 truth-source 映射，不复制第二套口径。 | docs-only |
| 2026-05-05 | 3.1 | 同步 A3 subject/ADR 重命名与 AI provider 口径：roadmap 只保留 provider endpoint 契约，不再把独立转发层作为项目关心的应用层事实。 | ai-provider-and-model-routing/001 remediation |
| 2026-05-03 | 3.0 | 根据 product-scope v1.5、当前 `docs/ui-design/` 与 `ui-design/` 重新规划 roadmap：删除 `docs/spec/INDEX.md` 的 pending child 占位模型，保留已存在 active spec 与编码 truth source，后续 child spec / plan 仅在进入设计或实现时按当前 P0 workstream 原地创建。 | 001-decompose-subspecs |
| 2026-05-03 | 2.4 | 当前技术契约改由 B1/B2/B3/B4/F1/F2/F3 active spec 与已编码 truth source 决定，禁止后续 child 绕过 Layer B/F owner。 | docs-only |
| 2026-05-03 | 2.3 | 对齐 product-scope v1.2 后的可执行契约口径：B2 改为 34 endpoint / 12 tag，B3 internal events 改为 16 个，B4 baseline 改为当时的应用表 + auth 支撑表 + 迁移元数据表，E1 mock gate 改为当前 B2 全量 operation。 | product-contract alignment review |
| 2026-05-03 | 2.2 | 对齐 product-scope v1.1 与当前 UI 真理源：P0 前端 child 改为 Home / Job Picks / Practice / Report / Resume / Debrief 六域；移除 onboarding、独立 mistakes、独立 growth、Drill、followup-tree、STAR、多轮计划和独立 voice page 的 roadmap 口径；C10 改为嵌入式 readiness signals。 | 001-decompose-subspecs |
| 2026-05-03 | 2.1 | 同步产品真理源迁移：将 parent roadmap 的产品输入从根目录 `easyinterview-spec-v1-0.md` 改为 `docs/spec/product-scope/spec.md`，并明确旧根 spec 仅作历史参考。 | docs-only |
| 2026-04-29 | 2.0 | 收口 A/B spec 全面审查 remediation：A1 根目录契约纳入 `shared/` / `config/`；Q2/B3/B4 接入 internal-only `email_dispatch`；Q6 以 ADR-Q6 为准重新锁定 AIClient 连接参数与 gateway fallback 边界；C9 真实面试复现升格为 P0 后端范围；B4 增补 AI call meta 字段与隐私删除表矩阵。 | plan-review remediation |
| 2026-04-29 | 1.9 | 同步 B4 `db-migrations-baseline` v1.4：parent roadmap 摘要从旧「29 表初始迁移」口径更新为 30 张应用 / auth 支撑表 + 迁移元数据表 + backfill ledger，并吸收 B3 outbox retry 字段承载，与 ADR-Q1 auth/session 表归属和 B4 migration 真理源一致。 | db-migrations-baseline plan-review remediation |
| 2026-04-27 | 1.8 | 对齐个人单人开发阶段决策：A5 不再作为 P0 远端 CI pipeline，当前只保留本地质量门禁；GitHub Actions、branch protection、artifact、nightly 与 CI secret 延后到多人协作、公开 release 或自动发版触发条件出现后再建。 | 001-decompose-subspecs |
| 2026-04-27 | 1.7 | 同步修订 ADR-Q4/Q6 与 A2/A3/A4 AI provider 边界：unit test / 离线契约测试才走 stub；docker compose 与 Kind 本地部署必须注入真实 AI provider endpoint / key，staging/prod 可指 cluster-internal gateway。 | 001-decompose-subspecs |
| 2026-04-27 | 1.6 | 同步修订 ADR-Q3/Q4 与 A2 本地开发栈边界：PostHog 部署验证改归 F2/E4，普通本地 dev 默认 no-op / file-backed；Kind 仅用于场景集成测试，A2 docker-compose 不再与 Kind manifest 同源。 | 001-decompose-subspecs |
| 2026-04-27 | 1.5 | 对齐 A2 local-dev-stack v1.2：parent roadmap 中 A2 口径改为当时的最小依赖 + 项目组件一键启动；F1 口径改为消费应用 `/metrics` / 日志与生产观测配置，不再要求 A2 默认提供 OTel/Grafana/Loki/Prometheus。 | 001-decompose-subspecs |
| 2026-04-27 | 1.4 | 修正 W1 Phase 3 gate 口径：parent plan 只完成 9 份 child spec 的 cross-spec review 与 spec-contract lock；A2/B2/F1/F3 的可执行 gate 交由各 child `001` plan 逐一验证，未通过前不得启动依赖它的 W2 implementation | 001-decompose-subspecs |
| 2026-04-26 | 1.3 | L2 code review remediation：ADR-Q3 从 PostHog Cloud 切换为自托管 PostHog 优先；补齐 async public `jobType` 与内部 Asynq handler 的命名边界；明确 Q-5 P0 导出延后是产品验收项的 W0 例外 | 001-decompose-subspecs |
| 2026-04-26 | 1.2 | W0 hard gate 6 项 ADR 全部 accepted（Q-1 自建 passwordless / Q-2 Asynq+Redis / Q-3 PostHog Cloud EU / Q-4 Kubernetes / Q-5 P0 仅删除 / Q-6 AIClient+Model Profile+外部 AI Gateway）；§3.2 表替换为锁定结论 + ADR 链接；§4.4 / §5.5 E4 / §6 C-1 / C-6 同步引用 ADR | 001-decompose-subspecs |
| 2026-04-26 | 1.1 | 补充 Q-6 AI 网关与模型路由 W0 决策项；A3 改为 provider-neutral `ai-provider-and-model-routing`；明确 Higress 等 AI Gateway 作为独立部署组件而非业务 SDK | 001-decompose-subspecs |
| 2026-04-26 | 1.0 | 初始创建：定义 6 层 38 child subspec、6 wave 实施顺序、mock-first 集成策略、5 项 W0 hard gate 决策项 | 001-decompose-subspecs |
