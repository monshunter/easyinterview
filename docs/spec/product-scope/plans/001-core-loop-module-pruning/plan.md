# Core Loop Module Pruning Plan

> **版本**: 1.141
> **状态**: active
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

按用户已确认的方案 B，硬清理当前 P0 中的真实面试复盘模块和候选人画像模块，使产品核心回到：

```text
上传 / 粘贴简历 + JD
  -> 模拟面试
  -> 报告
  -> 复练当前轮 / 进入下一轮
```

完成后，`debrief` / `profile` / `CandidateProfile` / `ExperienceCard` 不再作为用户可见入口、OpenAPI tag、后端领域、DB 表、AI feature key、shared event/job 或场景验收对象存在。账号设置、邮箱认证、首次资料补全与隐私删除保留，但不得继续承担“用户画像”产品语义。

## 2 背景

当前产品 scope、engineering roadmap、UI 文档、静态原型、正式前端、OpenAPI、backend、migrations、shared、config 和 E2E 场景都仍包含复盘和候选人画像。用户已明确选择硬删除方案，而不是隐藏或保留兼容层。

由于本项目尚未上线，不要求保留历史 route / API / schema 兼容。删除必须以当前 active spec、`docs/ui-design/`、`ui-design/` 和编码 truth source 为准，避免文档仍把范围外模块作为后续 workstream 自动纳入。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `contract` + `migration` + `code-internal` + `tooling`
- **TDD 策略**: 通过 `/implement product-scope/001-core-loop-module-pruning cross-layer` 进入 `/tdd`。每个 checklist item 在改实现前先补或改对应 red test：route/topbar/i18n/pixel parity、OpenAPI inventory/codegen/fixture validation、Go handler/store/service tests、migration/schema lint、scenario wrapper negative gate。
- **BDD 策略**: 用户可见入口和端到端流程会变化，必须维护 [bdd-plan](./bdd-plan.md) 与 [bdd-checklist](./bdd-checklist.md)。主 checklist 以 `E2E.P0.001`、`E2E.P0.088`、`E2E.P0.090`、`E2E.P0.098`、`E2E.P0.099`、`E2E.P0.102` 作为更新后的行为 gate，删除 `E2E.P0.060`-`E2E.P0.069`、`E2E.P0.071`、`E2E.P0.073`、`E2E.P0.091`-`E2E.P0.093` 的复盘 / 画像正向场景。
- **替代验证 gate**: API、DB、shared、config、prompt/rubric、generated artifacts 通过 `make codegen-check`、`make validate-fixtures`、migration lint / focused Go tests、repo-wide out-of-scope-negative grep 和 `git diff --check` 验证；不以历史 PASS 或 checklist 状态作为当前完成证据。

## 4 实施步骤

### Phase 1: 产品和 UI 真理源改写

#### 1.1 product-scope 与 engineering-roadmap 收敛

把 P0 闭环、一级导航、用户菜单、模块边界、阶段路线、验收标准和范围外能力改为三入口核心链路。`复盘`、`用户画像`、`Progressive Profile` 对用户可见模块和候选人画像数据模型从当前 scope 中删除，并登记为范围外能力。

#### 1.2 UI 文档与静态原型改写

更新 `docs/ui-design/` 与 `ui-design/src/app.jsx`、相关 screen 源码，使一级导航只保留 `首页 / 模拟面试 / 简历`，用户菜单只保留 `设置与隐私 / 退出登录`，`debrief`、`debrief_full`、`profile` 不再是目标 route 或用户入口。

### Phase 2: 前端实现清理

#### 2.1 App shell route 和用户菜单清理

更新 `frontend/src/app/routes.ts`、`normalizeRoute.ts`、`routeUrl.ts`、`App.tsx`、TopBar 和 URL fallback，使 `debrief` / `profile` 不再进入 RouteName、primary nav、user menu、canonical path 或 out-of-scope live route。未知或范围外路径必须归一到当前核心入口，且不暴露范围外模块页面。

#### 2.2 复盘和画像 screen / hook / i18n / tests 删除

删除 `frontend/src/app/screens/debrief/`、`ProfileScreen`、复盘 / 用户画像 i18n key、dev mock debrief special cases、frontend owner tests 和 pixel parity spec 中的正向复盘 / 画像断言，补充负向测试证明范围外 UI contract 不会回流。

### Phase 3: OpenAPI、shared、generated 和 fixture 清理

#### 3.1 API contract 删除

删除 OpenAPI `Profile` / `Debriefs` tags、paths、schemas、fixtures 和 generated Go/TS client/server artifacts。更新 inventory、fixtures、mock transport allowlist 和 frontend/backend API tests。

#### 3.2 shared event/job/enum 清理

删除 `debrief.created`、`debrief.completed`、`debrief.generate`、`PracticeGoal=debrief`、`source_debrief_id` API/shared surface，以及候选人画像 / experience card 相关共享契约。保留账号资料补全和隐私删除必要的账号数据，不把它们命名为用户画像。

### Phase 4: 后端、迁移、AI config 清理

#### 4.1 backend debrief/profile 领域删除

删除 `backend/internal/debrief`、`backend/internal/api/debriefs`、`backend/internal/profile` 的 candidate profile / experience card 领域代码和 cmd/api wiring。更新 practice/report/resume/auth 直接消费者，确保核心链路不依赖 debrief/profile。

#### 4.2 migration 和 seed 清理

在当前未上线前提下修订 baseline/seed migrations，清理 `debriefs`、`candidate_profiles`、`experience_cards`、`practice_plans.source_debrief_id`、`goal='debrief'`、debrief AI prompt/rubric seeds、profile update seed 和相关 enum source。

#### 4.3 AI prompt/rubric/profile 清理

删除 `config/prompts/debrief.*`、`config/rubrics/debrief.*`、`config/evals/debrief.*`、`config/ai-profiles.yaml` 中的 debrief / profile feature key，并更新 prompt/rubric registry gate。

### Phase 5: 场景、文档索引和验收收口

#### 5.1 E2E 场景目录与索引清理

删除复盘 / 画像正向场景目录和索引项，更新核心闭环场景，使它们验证三入口 app shell、范围外路径负向归一、JD 到报告再到复练 / 下一轮的闭环仍通过。

#### 5.2 zero-reference和质量门禁

运行文档、contract、backend、frontend、migration、scenario 和 out-of-scope-negative gate。任何命中 `debrief` / `CandidateProfile` / `ExperienceCard` / `用户画像` / `复盘` 命中必须分类为允许的审计材料、范围外文件路径、报告 / work-journal 记录，或继续清理。

### Phase 6: Active 文档漂移复查

#### 6.1 Event / job active spec 对齐

反查 `shared/events.yaml`、`shared/jobs.yaml` 和 generated events/jobs truth，修订 active B3 spec 中仍保留的 18-event / 11-job / `debrief` / `jd_match` 正向口径，使文档回到当前 14-event / 8-job 合同。

#### 6.2 Workspace / practice active spec 和计划对齐

反查正式前端与 UI 真理源，修订 `frontend-workspace-and-practice` active spec / plan / BDD / test 文档中仍把 `company_intel` 当外部详情 owner 或把 `goal=debrief` 当正向显隐组合的口径；当前只保留 workspace 内嵌公司情报卡片与 `baseline / retry_current_round / next_round` 核心 practice goals。

#### 6.3 Out-of-scope subject 生命周期对齐

修订已 out-of-scope 的 debrief / profile / jobs-recommendations subject 及其 plans index，避免 out-of-scope spec 下面继续投影 active plan 或 future 保留编号建议。

#### 6.4 文档索引与语义 gate

运行 `sync-doc-index --check`、`make docs-check`、out-of-scope-negative active-doc grep 和 `git diff --check`，确认 active 文档不再把范围外模块作为正向 owner / future work。

#### 6.5 B4 migration active spec 与 privacy matrix 收敛

反查 `db-migrations-baseline` active spec、baseline migration truth 与 privacy delete matrix，使 D-22 后 `candidate_profiles` / `experience_cards` / `debriefs` 不再作为当前表项，同时当前仍存在的 `idempotency_records` 等用户关联表必须有明确 privacy disposition 和测试覆盖。

#### 6.6 Context manifest 正向 surface 收敛

反查 `context.yaml` 的 target discovery，确保 `uiRoutes` 与 `apiNames` 只列当前保留 route / operationId；`debrief`、`profile`、`Profile` / `Debriefs` operationId、候选人画像 operationId 只能保留在关键词、operation matrix 或负向断言文本里，不能作为 `/plan-code-review` / `/implement` 的正向 target surface。

#### 6.7 Backend-practice owner 语义收敛

反查 active subject 下的 `backend-practice/003-mode-policies-and-provenance` 与 `backend-practice/004-report-derived-practice-plans`：003 的 mode × goal 矩阵只覆盖 `baseline / retry_current_round / next_round`；004 的正向范围只覆盖 report-derived `retry_current_round` / `next_round` + `sourceReportId`；范围外 source / goal 只能作为负向输入，不得作为当前 owner gate。

#### 6.8 Backend infra / AI contract owner 语义收敛

反查 active subject 下仍被 `context.yaml` 作为执行入口的 owner plans，优先处理 `backend-async-runner/001-internal-job-outbox-runner` 与 `prompt-rubric-registry/002-output-schema-contract`：当前正向 runner 范围必须是 7 个可执行 handler + `privacy_export` contract-only；当前 prompt/rubric truth source 必须是 9 个 chat feature_key。范围外 Debrief / JD Match 正向矩阵、包路径、测试命令和 context discovery 只能作为负向断言，不得作为当前 `/implement` / `/plan-code-review` target surface。

#### 6.9 JD Match migration final-state 审计

反查 `migrations/000009_jd_match_baseline.*`、`migrations/000010_jd_match_seed_registry.*` 与 `migrations/000014_drop_jd_match_module.*` 的当前关系：`000009` / `000010` 只作为 pre-launch DDL / seed 记录保留，`000014` 必须删除 5 张 JD Match 表、清理 `jd_match.*` prompt/rubric registry rows，并在收窄 `async_jobs.job_type` check 前删除范围外 `jd_match_agent_scan` / `jd_match_search` job rows。B4 active spec 不得继续把 B3 canonical job type 误写为范围外 9 项。

#### 6.10 Bootstrap owner 当前口径收敛

反查 active contract subject 下仍作为 owner 入口存在的 bootstrap plans，优先处理 `event-and-outbox-contract/001-bootstrap` 与 `db-migrations-baseline/001-bootstrap`：B3 当前正向口径必须是 14 events / 8 canonical jobs / 6 API-facing job types；B4 当前正向口径必须是 22 应用表 + 3 auth 支撑表、public schema ≥27、B3 8 canonical jobs、B2 6 API-facing jobs。范围外 16-event / 9-job / 30-table 口径不得作为当前 checklist、handoff 或 context discovery 的正向 gate。

#### 6.11 Profile / Jobs Recommendations 范围外 subject 实体删除

反查范围外 subject 下保留的 owner plans，优先处理 `backend-profile/001-candidate-profile-and-experience-cards` 与 `backend-jobs-recommendations/001-jd-match-real-backend-baseline`：若当前 active spec、coded truth source、migration final-state guard 和 scenario negative gate 已能独立承接相关字段、事件、指标、schema 和验证要求，则删除对应 subject 实体目录，不再保留说明文件、standalone note 或 parallel plan。

#### 6.12 Debrief 范围外 subject 实体删除

反查 `backend-debrief/001-debrief-record-and-analysis` 与 `frontend-debrief/001-debrief-screen-and-handoff`：若当前 product-scope、OpenAPI、frontend route normalization、backend route-negative tests、migration/privacy gates 和 E2E negative scenarios 已能独立证明 Debrief 删除，则删除对应 subject 实体目录，不再保留说明文件、standalone note 或 context 收敛历史包。

#### 6.13 Repo-wide context 正向 surface 审计

结构化解析 `docs/spec/**/context.yaml`，只检查 `targets.*.discovery.packages` / `uiRoutes` / `apiNames` 三类会驱动 `/implement`、`/plan-code-review` 或 worker discovery 的正向字段；范围外 Debrief / Profile / JobMatch operationId、route、backend/frontend package、fixtures、prompt/rubric path 不得在这些字段中作为当前 target surface 出现。范围外 subject 不得再以自身文档目录路径或范围外 shorthand 形式保留为执行入口。

#### 6.14 Runtime / generated / config 负向审计

反查 `backend/`、`frontend/`、`openapi/`、`shared/`、`config/`、`scripts/`、`migrations/`、`test/scenarios/` 与 `ui-design/` 中的范围外 Debrief / Profile / JobMatch runtime surface：允许负向测试、历史迁移和 out-of-scope alias normalization 命中，但不得保留无人引用的正式样式、组件、handler、fixture、operationId、prompt/rubric、job/event 或正向场景资产。发现 `frontend/src/app/theme/global.css` 中仅服务范围外 JD Match screen 的 `jdmatch-*` responsive class 时，删除并用 frontend scope gate 固化。`migrations/enum-sources.yaml` 中历史 JD Match enum source 仍受 `000009` historical up migration 与当前 lint 模型约束；除非用户批准 pre-launch migration squash 或先改造 lint 为最终态 schema 模型，否则不得直接删除导致 migration gate 失真。

#### 6.15 无争议范围外文档包实体删除

按用户 2026-07-06 明确要求，已经无争议范围外的模块文档包直接删除。删除 `backend-debrief`、`frontend-debrief`、`backend-profile`、`backend-jobs-recommendations` 等不再承接当前合同的范围外 subject 实体目录，并更新 `docs/spec/INDEX.md`、跨计划引用、context discovery 和验收报告。保留范围仅限 work-journal、bug、report、migration 历史、负向测试和当前 active owner spec 必需的删除证据；若某历史 plan 的内容仍有字段 / schema / gate 价值，必须先迁移到当前 owner spec 或 coded truth source，再删除原目录。

#### 6.16 Runtime / generated allowlist 脚本化

把 6.14 的手工 runtime / generated / config 负向审计沉淀为 `scripts/lint/core_loop_pruning_surface.py`：扫描 `backend/`、`frontend/`、`openapi/`、`shared/`、`config/`、`scripts/`、`migrations/`、`test/scenarios/` 与 `ui-design/` 的范围外 Debrief / Profile / JD Match surface，并分桶输出 `migration_records`、`out_of_scope_normalization`、`negative_tests` 和 `real_residuals`。脚本必须允许 migration 链、out-of-scope route normalization 和负向测试 / lint guard 命中，但 `real_residuals` 非空时失败；若发现真实命中文档或代码，按当前 owner spec 直接删除或修订，不保留 parallel out-of-scope 包。

#### 6.17 范围外输入 / 文档实体删除

按用户 2026-07-06 追加要求继续删除范围外历史入口：删除根目录范围外产品输入、`historical-spec-implementation-review` 执行编排 subject、`docs/ui-design/review-module.md` 和 `docs/ui-design/user-profile-and-settings.md`。同步 `product-scope`、`docs/spec/INDEX.md`、`docs/ui-design/INDEX.md`、context 引用、范围外报告链接和正式前端注释，使当前 active truth source 只指向 `product-scope`、active `docs/ui-design/`、`ui-design/` 与编码 gate，不保留范围外 standalone note、stub 或 active 历史执行入口。

#### 6.18 UI 范围边界文档实体删除

删除对应 UI 范围边界文档包，不新增说明文件。仍需保留的岗位推荐 / 公司情报零入口、范围外 route 归一和当前模块边界约束，由 `docs/ui-design/module-map.md`、`docs/spec/product-scope/spec.md`、`ui-design/src/app.jsx` 和正式前端 `normalizeRoute` 承接；所有 context、INDEX、README、spec 和 checklist 引用不得再指向该文档包。

#### 6.19 根入口产品摘要收敛

反查仓库第一屏和高层索引摘要，删除仍把真实面试复盘、用户画像或范围调整说明写进当前产品主闭环的口径。根 `README.md` 必须只描述当前 `JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮` 链路；`docs/ui-design/INDEX.md` 的 active 行只摘要当前正向 UI 真理源。

#### 6.20 Backend Resume active spec 扁平合同收敛

反查 `backend-resume` active spec 与仍作为执行入口的 context manifest 当前正向合同，删除版本树、主版本、岗位定制版本、服务端逐条采纳 / 拒绝状态机和范围外 operation/table 标识的解释性正文或 discovery 词。`backend-resume/spec.md` 和 `backend-resume` context 必须只以当前 9 个 operation、`resumes` 单表、`ai_task_runs` tailor 输出、`updateResume` 覆盖和 `duplicateResume` 另存作为正向事实；历史计划和 work-journal 可保留事实记录，但不得作为 active spec 当前行为说明。

#### 6.21 Product Scope active spec 当前范围口径收敛

反查 `docs/spec/product-scope/spec.md` 作为当前产品 owner truth source 的正文、决策表、范围章节和验收标准，删除以范围变更过程解释当前行为的口径。当前 spec 必须以正向合同和负向边界描述范围：当前 P0 保留 JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮；复盘、用户画像、岗位推荐、公司情报独立页和简历版本树等只列入范围边界，不写生命周期说明。

#### 6.22 Frontend Resume Workshop active spec 扁平 UI 合同收敛

反查 `frontend-resume-workshop` active spec 与 context manifest 的当前正向 UI / generated-client 合同，删除 D-20 前版本树、范围外 operation、范围外 adapter 和范围外组件名的解释性正文或 discovery 词。`frontend-resume-workshop/spec.md` 与 001/002/003 context 必须只以当前 `Resume` / `resumeId`、upload/paste create、accept-only rewrites、`updateResume` 覆盖和 `duplicateResume` 另存作为正向事实；范围外 UI 入口负向断言由 product-scope pruning gate 承接，不在 active spec 中保留范围外说明。

#### 6.23 UI Design active truth source 范围外口径收敛

反查 `docs/ui-design/` active 文档和 INDEX 的当前目标说明，删除“范围外流程 / 范围外模块 / 范围外组件 / 已随版本删除”这类历史解释口径。UI 真理源必须以当前页面、当前入口、当前 route 归一和范围外边界描述产品形态；`out-of-scope` 仅允许作为 README 中的文档状态枚举，不作为模块说明。

#### 6.24 Engineering Roadmap active spec 当前 workstream 口径收敛

反查 `engineering-roadmap` active spec 与 context 的当前 P0 workstream、mock-first 计数、implementation order 和验收标准，删除以范围变更过程解释当前执行地图的口径和 discovery 词。Roadmap 必须只描述当前 active owner、当前 operation/tag inventory、当前 UI 三入口和范围边界。

#### 6.25 OpenAPI / Shared / Mock contract active spec 当前契约口径收敛

反查 `openapi-v1-contract`、`shared-conventions-codified` 与 `mock-contract-suite` active spec 正文和 context discovery，使 B2/B1/E1 当前 contract 只描述 10 tag / 35 operation、16 个共享枚举、flat Resume vocabulary、current-scope negative search 和 fixture coverage。范围外模块、范围外 operation、范围外 schema 和 out-of-scope tooling 说明不得作为 active scope、决策、schema inventory、模块边界、验收标准或正向 `apiNames` / keyword discovery 口径；必要历史明细只保留在独立 history 文件。

#### 6.26 Frontend Report Dashboard flat Resume contract 收敛

反查 `frontend-report-dashboard` active spec、plan、checklist 与 context discovery，使 ReportContextStrip 当前合同只通过 generated `getResume(resumeId)` 读取 `displayName`，不再把 ResumeVersion / resumeVersionId / resume_versions / ResumeAsset 字段或范围外模块说明作为 active 行为、operation matrix、验收标准、正向 discovery 或文档收口口径。当前边界必须以 `generating / report` 两条 owner route、B2 Reports + TargetJobs + Resumes generated client、stale-contract negative gate 和 product-scope 当前三入口主链路表述。

#### 6.27 Frontend Home / Parse active spec 当前合同收敛

反查 `frontend-home-job-picks-and-parse` active spec 与 context discovery，使当前 owner 只描述 Home + Parse 新建模拟面试入口、当前 UI truth source、TargetJobs / Uploads / Resumes generated-client contract、workspace handoff、privacy gate 与 P0.014-P0.016 BDD。范围外历史模块、历史计划段落和当前实现无关的删除说明不得留在 active spec 正文中。

#### 6.28 OpenAPI README / baseline 当前 inventory 口径收敛

反查 `openapi/README.md`、`openapi/baseline/README.md` 与 `openapi/diff-config.yaml`，使当前 API freeze 只用 10 tag / 35 operation 正向 inventory、flat Resume operations、Practice session / voice、Auth profile-completion 与 privacy export whitelist 表述，不再通过范围外模块或范围外 tooling 说明解释当前 contract。

#### 6.29 Event / outbox active spec 当前 inventory 口径收敛

反查 `event-and-outbox-contract` active spec、bootstrap plan 与 context discovery，使当前 B3 合同只用 14 events / 8 canonical jobs / 6 API-facing job types、6 个 event domain、flat Resume event identity 和当前 `ResumeTailorMode` 字面量表述。范围外 event domain、job、payload 字段或历史迁移说明不得作为 active B3 spec 当前行为、context version 或 completed bootstrap plan 的正向说明。

#### 6.30 Backend Async Runner active spec 当前 runtime 口径收敛

反查 `backend-async-runner` active spec 与 context discovery，使当前 runtime 合同只用单一 kernel、7 个可执行 handler、`privacy_export` contract-only、B3 8 job inventory、current handler owner map 和 generic out-of-scope runner gate 表述。范围外 job、domain runtime、payload/table 字段或历史迁移说明不得作为 active backend runtime 当前行为、context version 或验收标准正向说明。

#### 6.31 Prompt / rubric active spec 当前 feature_key 口径收敛

反查 `prompt-rubric-registry` active spec 与 002 context discovery，使当前 F3 合同只用 9 个 baseline feature_key、provider-neutral output schema、current AI profile coverage 和 current eval case count 表述。范围外 feature_key、prompt/rubric/eval/profile seed、过往 additive 说明或模块边界解释不得作为 active F3 spec 当前行为、context keyword 或 AI profile coverage 说明。

#### 6.32 Frontend Shell active spec 当前 route catalog 口径收敛

反查 `frontend-shell` active spec 与 001/004 context discovery，使当前 D1 合同只用当前 route catalog、TopBar 三入口、用户菜单、auth flow、safe params 与 out-of-scope alias 归一规则表述。范围外 route/module alias、out-of-scope wording、out-of-scope route list 或 pre-D20 resume param 不得作为 active frontend shell 当前行为、context keyword 或验收标准正向说明。

#### 6.33 Frontend Workspace / Practice active spec 当前 owner 合同收敛

反查 `frontend-workspace-and-practice` active spec、history 与 001/002 context discovery，使当前 D2/D3 合同只用三条 owner route、workspace 嵌入式公司轻情报摘要、flat Resume binding 和 `baseline / retry_current_round / next_round` practice goals 表述。独立公司情报页面/API、pre-D20 简历版本字段、范围外 practice goal 或历史模块解释不得作为 active workspace/practice 当前行为、context keyword 或验收标准正向说明。

#### 6.34 Workspace insight route/API alias 实现与文档收敛

反查 `ui-design/`、正式 `frontend/`、`docs/ui-design/`、`frontend-shell` / `frontend-workspace-and-practice` owner plans 和 P0.005 / P0.006 / P0.021 场景资产，使公司信号只保留为 workspace 内嵌 `WorkspaceInsightCard` 摘要。正式前端不保留独立洞察 route/url fallback、SPA fallback、独立 API consumer、范围外组件导出名或范围外场景正向说明；范围外独立详情路径按通用 fallback 处理。

#### 6.35 UI architecture / module-map 当前边界口径收敛

反查 `docs/ui-design/ui-architecture.md`、`docs/ui-design/module-map.md` 与 `docs/ui-design/INDEX.md`，使 active UI truth source 只用当前模块归属、范围外能力、范围外 route 输入归一和当前边界描述目标 UI。不得用范围变更过程或清理动作解释当前 UI 形态；需要保留的 route token 只能作为范围外输入 fallback 合同或负向断言。

#### 6.36 UI design active docs 全目录当前边界口径收敛

反查 `docs/ui-design/` active 文档、README、TEMPLATES 和 INDEX，使报告、用户流程、简历、当前面试规划、面试/报告、无简历引导、认证入口和多 JD / 多简历管理文档统一使用当前唯一形态、范围边界、范围外 route 输入和当前布局描述。全目录不得用范围变更过程、已消失入口/画板或清理动作解释当前 UI 目标；全局 Header 状态枚举元数据除外。

#### 6.37 Frontend route negative test 命名口径收敛

反查正式前端 route codec、route catalog、SPA fallback、scope guard 和 P0.005 visual smoke 测试，使当前执行中的负向 route gate 使用 `out-of-scope` / 范围外输入语言，而不是直接暴露样本来源口径。导出名、文件名、场景 slug 和 wrapper 断言统一使用 out-of-scope 命名；测试行为只验证范围外 route/path 不 materialize。

#### 6.38 Engineering Roadmap 当前执行地图口径二次收敛

反查 `engineering-roadmap/spec.md` 的背景、范围、决策、P0 workstream、执行顺序和验收标准，使 roadmap 只以当前 truth source、当前 owner spec、当前 10 tag / 35 operation、当前 UI 三入口、out-of-scope route/module 边界和 no-pending INDEX model 描述执行地图。历史 root spec / route / 技术草稿只能作为范围外输入或边界条件，不得作为当前 owner、future workstream 或纳入依据。

#### 6.39 Engineering Roadmap plan / checklist / context 当前治理口径收敛

反查 `engineering-roadmap/plans/001-decompose-subspecs` 仍作为 completed owner plan 的 plan、checklist 与 context discovery，使其只描述当前 execution-map governance、真实 spec INDEX 投影、按需 child 创建、current owner spec 依赖和 technical-draft zero-reference gate。该 plan 不得用 wave、pending backlog、范围外模块纳入或技术草稿处置过程解释当前治理合同。

#### 6.40 Frontend / UI Resume picker 与 flat Resume locale 当前合同收敛

反查正式前端 workspace 简历选择、Resume Workshop locale、dead component 和静态 UI contract，使当前实现只描述 active flat Resume 列表、`listResumes` 消费、`resumeId` 选择和 overwrite-or-new 改写保存。不得保留 disabled placeholder、`NotImplementedPlaceholder`、未引用 coming-soon tab、`master` / `targeted` / branch 文案 key、version-tree 文案、范围外命名的展示 copy 或把范围外树/分支 namespace 作为可用功能。

#### 6.41 Resume 场景资产与 BDD owner 当前语义收敛

反查 `test/scenarios/e2e` 中承接 backend-resume 与 frontend-resume-workshop flat Resume 合同的场景目录、INDEX、README、data、scripts，以及对应 BDD plan / checklist。保留仍覆盖当前行为的场景编号，但目录 slug、场景描述、输出路径、owner BDD 文档必须用 `flat read`、`updateResume`、`duplicateResume`、`Rewrites accept-only save`、`tailor privacy` 和 `flat UI regression` 描述当前合同；不得用历史目录名、范围外说明、范围外版本树 / 分支流程说明或 accept/reject 范围外状态机作为场景语义。

#### 6.42 Executable out-of-scope gate naming 收敛

反查正式前端 route codec / SPA fallback / scenario tests、后端 route guard tests、lint scripts、P0.059 / P0.090 场景目录和触发脚本，使当前执行中的负向 gate、Make target、文件名和测试名统一使用 `out-of-scope` 命名。仍作为禁止输入样本的字符串只保留在 lint/test fixture 内，不得作为目录名、脚本名、测试名、输出路径或正向场景语义。

#### 6.43 Lint tooling out-of-scope terminology 收敛

反查 `scripts/lint/` 中当前执行 gate 的类名、函数名、bucket、测试名和错误输出，使 core-loop surface、runtime topology、AI provider terminology、mock runtime boundary、fixture validator、prompt/rubric lint 等工具统一用 `out-of-scope` 描述当前负向合同。正则中的历史英文词只作为 forbidden input 匹配样本保留，不得作为工具输出、测试名或 bucket 名。

#### 6.44 Scenario / pixel parity out-of-scope wording 收敛

反查 active P0 scenario README / expected outcome / shell scripts、pixel-parity specs，以及 e2e / workspace / report BDD owner 文档，使当前负向断言、脚本输出、artifact marker 和测试变量统一使用 `out-of-scope` 表述。仍需拦截的历史输入值在 shell 中通过拼接变量保留，不得在文档正文、脚本输出或测试名中直接暴露范围外口径。

#### 6.45 Active code / prototype out-of-scope wording 收敛

反查 active Go code comments、ui-design static prototype data、ui-design contract tests、markdown link tooling docs 和 backend-practice lint samples，使正式代码与原型文案不再用历史英文标签描述当前适配器、drainer、mail dispatch、workspace insight 或 resume negative gate。lint 正则可继续包含范围外词作为 forbidden-context matcher，但不能出现在用户可见 mock data、active comments、test names 或 error output 中。

#### 6.46 Active docs/spec out-of-scope wording 收敛

反查 `docs/spec/` 中除 `history.md` 外的 active spec / plan / checklist / BDD / context 文档，把当前负向边界、验证证据和 owner handoff 中的历史英文标签统一为 `out-of-scope`。该步骤不改文件路径、不改状态枚举、不碰保留为审计材料的 `history.md`。

#### 6.47 Active test out-of-scope wording 收敛

反查后端 Go tests 与前端 generated-contract tests 中的变量名、测试名、错误文案和 fixture sample，使当前负向断言统一使用 `out-of-scope` 表述。需要保留的 forbidden enum / evidence 输入通过字符串拼接维持实际测试值，不在源码中直接暴露范围外英文标签。

#### 6.48 Spec history out-of-scope wording 收敛

反查 `docs/spec/**/history.md` 的审计文字，把英文范围外标签统一为 `out-of-scope`，保留时间线、版本号、owner 和证据来源不变。该步骤不改 work-journal、bug、report 或 migration 审计材料。

#### 6.49 Product-scope 中文范围边界当前合同收敛

反查 `docs/spec/product-scope/spec.md` 与 `history.md` 中的中文范围边界、阶段路线、验收标准和修订记录，使产品范围 owner 只描述当前合同、范围外和当前验证入口。范围变更过程不得作为当前行为依据；隐私删除链路、会话历史和审计类词汇作为当前功能语义保留。

#### 6.50 跨 owner 中文范围外口径精确收敛

反查 active / completed owner docs、frontend README 和当前 E2E scenario 文档中的中文过程性词汇，使范围变更过程词和 earlier route / module / testid / operation 短语不再作为当前合同叙述方式。当前仍需拦截的输入改写为范围外输入、范围外边界或负向 gate；隐私删除、migration drop、cleanup 真实动作和 lint forbidden matcher 不在本批机械改写范围内。

#### 6.51 Product-scope owner plan 自描述收敛

反查本 plan、checklist 与 BDD 文档自身，使 owner 文档使用当前合同、范围外边界、zero-reference gate 和 data-erasure 语义描述执行要求，不再通过范围变更过程词、早期 subject / path / testid 表述或机械替换产物解释当前 owner 事实。校正本轮机械替换带来的中英混合伪文件名和状态枚举误写，确保 owner 文档可继续作为后续扫描入口。

#### 6.52 UI design active docs 当前动作词收敛

反查 `docs/ui-design/` active 文档、README、TEMPLATES 和 INDEX，使登录后 pendingAction 表述统一为“接续”，并把范围外入口、模块边界和模板示例改写为当前 UI 合同语言。模板不再保留范围外模块示例行，UI 文档状态说明只描述当前常用状态集合。

#### 6.53 Frontend README / scenario 入口动作词收敛

反查 `frontend/README.md`、`test/scenarios/README.md`、`test/scenarios/e2e/README.md`、E2E INDEX 和当前 P0 scenario README，使 pendingAction、cleanup marker、失败处理、范围外 auth / route 输入和真实 provider fail-fast 使用当前动作词。保留 privacy data-erasure、retry、history stack 和 cleanup 真实行为语义。

#### 6.54 Scenario data / scripts 当前动作词收敛

反查当前 P0 scenario data 与 shell scripts，使 forbidden input 变量名、cleanup 输出、privacy data-erasure、canonical URL、auth handoff 和真实 provider 样本文案使用当前动作词。保留负向测试值本身、hard delete 语义和场景可执行脚本行为。

#### 6.55 Frontend Home / Parse owner 当前合同措辞收敛

反查 `frontend-home-job-picks-and-parse` active spec、001 plan/checklist/BDD 和 plans INDEX，使 Home + Parse owner 只用当前 auth continuation、route boundary、0-hit gate、data-erasure 与 TargetJobs/Uploads/Resumes generated-client 合同描述当前实施范围。P0.017 / jd_match 不进入当前 BDD 矩阵。

#### 6.56 Prompt registry status / 9-key contract cleanup

反查 `config/prompts/README.md`、`config/ai-profiles.yaml`、`scripts/lint/prompt_lint.py`、`backend/internal/ai/registry`、`backend/internal/ai/aiclient/profile`、`frontend/src/lib/events/events.test.ts`、`frontend/src/app/screens/practice/hooks/usePracticeSession*` 与 `prompt-rubric-registry/001-baseline` owner docs，使 prompt status enum 统一为 `draft | active`，F3 baseline 统一为 9 个 canonical `multi` prompt/rubric 坐标，profile catalog 使用 8 个唯一 profile row 服务 9 个 feature_key，并让当前测试命名、注释和 preflight 断言同步到 spec v2.15。

#### 6.57 Active executable / scenario wording cleanup

反查 dev-stack preflight、lint scripts、runner/auth comments、migration/privacy tests、rubric README、frontend pixel specs、frontend current tests 和 P0 Home/Parse scenario 文案，使当前执行面只用 current / out-of-scope / outside-current-scope 表述。保留实际负向输入样本、placeholder UI 语义、migration SQL 和 route normalization 行为。

#### 6.58 Local-dev / config current wording cleanup

反查 `local-dev-stack` spec/history/001 plan/checklist、`deploy/dev-stack/README.md`、`config/README.md` 与 `config/prompts/README.md`，使 Postgres volume preflight、dev-doctor、AI profile 和 prompt hash 文档使用当前不兼容布局、固定服务口径、out-of-scope alias 与 excluded-field 表述。仅收敛文档措辞，不改变命令、配置或测试逻辑。

#### 6.59 Practice voice/workspace owner current wording cleanup

反查 `practice-voice-mvp` spec/001 plan/checklist/context 与 `frontend-workspace-and-practice` 001/002 plan、checklist、BDD/test 文档，使 voice route、practice enum、workspace active-list、cached context 和 voice owner co-location 负向 gate 使用 out-of-scope / current owner 边界表述。仅收敛 owner 文档措辞，不改变前端、后端、OpenAPI 或场景逻辑。

#### 6.60 Workspace owner process-word cleanup

反查 `frontend-workspace-and-practice/001-workspace-and-interview-context` plan、checklist、BDD/context 与 P0.018/P0.021 场景资产，使当前合同只描述 workspace 当前规划、flat Resume Picker active-list、start-practice、embedded insight、records static affordance 和 privacy gates。保留 records static area 行为，不改变 OpenAPI 或后端契约。

#### 6.61 Report snapshot deletion

删除两份只在 reports INDEX 中引用、且已由后续交付台账承接事实的早期 L2 reconcile snapshot 报告，并同步 `docs/reports/INDEX.md`。不新增说明文件，不改保留的 bug / work-journal / report evidence。

#### 6.62 Report module-positive cleanup

删除 `docs/reports/` 中只被 reports INDEX 引用、且仅承接范围外 Debrief / Profile / JD Match / Jobs Recommendations 正向交付的报告实体，并同步 `docs/reports/INDEX.md`。保留仍被 work-journal 直接链接的报告证据，以及 AI/model profile 和 auth profile 等当前术语报告。

#### 6.63 Active-doc pruning report cleanup

删除只被 reports INDEX 与当天 work-journal 链接、且内容仍承接旧模块生命周期说明的 2026-07-06 active-doc pruning 复盘报告，同步 reports INDEX 和 work-journal 链接。当前证据以本 owner plan/checklist、BUG-0135 和现行 gates 承接，不新增说明文件。

#### 6.64 Product-scope owner note wording cleanup

清理本 owner plan/checklist 自身残留的旧 lifecycle-note 口径与过时 privacy test 名称，统一为当前 `out-of-scope` 术语和 standalone note 删除口径；不改变代码、接口、迁移或场景行为。

#### 6.65 Runtime/generated allowlist report cleanup

删除只被 reports INDEX 引用、且内容已由 `make lint-core-loop-pruning-surface` 与本 owner gate 承接的 2026-07-06 runtime/generated allowlist 复盘报告，并同步 reports INDEX。当前可执行证据保留在 lint gate、6.16/6.64 checklist 与本次收口验证中。

#### 6.66 June core-loop report cleanup

删除 3 份只被 reports INDEX 引用、且其证据已由本 owner plan/checklist、BUG 台账和 work-journal 承接的 2026-06 core-loop 复盘报告，并同步 reports INDEX。当前 gate 以可执行 lint、context validation、docs-check 和本 checklist 证据为准。

#### 6.67 Standalone lifecycle-term report batch cleanup

删除 15 份只被 reports INDEX 精确引用、且正文含旧 lifecycle-term 说明的 standalone 复盘报告，并同步 reports INDEX。保留仍被 work-journal、spec 或其他报告直接引用的报告，避免破坏审计链路。

#### 6.68 Work-journal referenced report cleanup

删除 10 份只被 reports INDEX 与 work-journal 直接引用的 standalone 复盘报告，同步 reports INDEX，并将对应 work-journal 行改为直接记录 BUG、验证或当日工作事实。带有 spec 或其他报告引用的报告继续保留待单独审计。

#### 6.69 Spec/report referenced lifecycle report cleanup

删除剩余 3 份含旧 lifecycle-term 且被 spec/report/work-journal 直接引用的 standalone 复盘报告，改写 `openapi-v1-contract/001`、`frontend-shell/003`、相关 work-journal 与相邻报告中的链接，使当前 owner plan/context 直接承接事实，不再依赖报告互链。

#### 6.70 Work-journal lifecycle-term wording cleanup

清理 `docs/work-journal/` 中仍使用 configured strict lifecycle-term token set 或 lifecycle banner 口径的工作记录与索引行，改为直接事实描述、`out-of-scope` 负向 gate 术语或删除事实；不新增报告、说明文件或生命周期注释。

#### 6.71 Historical-spec report cleanup

删除 `docs/reports/` 中剩余的 `historical-spec` standalone 报告实体，并同步 reports INDEX、BUG-0006 与 work-journal 中指向这些报告/ledger 的表述。当前证据保留为 owner gate、BUG 直接验证行和 work-journal 事实记录，不新增替代报告或说明文件。

#### 6.72 Active spec strict lifecycle-term cleanup

清理 active/completed `docs/spec/**` subject artifacts 中仍使用 configured strict lifecycle-term token set 的说明性残留，改为当前 `out-of-scope`、兼容 alias、warning 或直接删除事实；保留 docs schema/status enum 对该状态值的定义，不新增说明文件。

#### 6.73 Lint rule source strict-token assembly cleanup

调整 `scripts/lint/core_loop_pruning_surface.py` 与 `scripts/lint/runtime_topology.py` 的负向上下文正则，使规则仍能匹配 configured strict lifecycle-term token set，但源码不直接保存这些 token 的完整字面量；用相邻 lint 测试固定匹配能力与源码约束。

#### 6.74 Bug knowledge base strict lifecycle-term wording cleanup

清理 `docs/bugs/` 知识库中 configured strict lifecycle-term token set 的直述残留，改为当前 `out-of-scope`、current contract、removed 或 stale wording；保留 BUG 记录、根因、验证命令和 commit 字段结构，不新增说明文件。

#### 6.75 Governance strict lifecycle-term wording cleanup

清理根级治理指令中 configured strict lifecycle-term token set 的直述残留，改为旧模块生命周期说明、英文旧状态 banner 和归档说明等中性表述；docs lifecycle status enum 与 package lock metadata 作为结构契约/外部元数据单独分类，不在本阶段重命名。

#### 6.76 Broader old-scope standalone report cleanup

删除 `docs/reports/` 中命中 broader old-scope wording、且只被 reports INDEX 引用的 standalone 复盘报告实体，并同步 reports INDEX。仍被 work-journal 或其他 owner 文档直接引用的报告保留待单独审计，不新增说明文件。

#### 6.77 Work-journal-linked broader old-scope report cleanup

删除 `docs/reports/` 中命中 broader old-scope wording、且只被 reports INDEX 与 work-journal 直接引用的 standalone 复盘报告实体；将对应 work-journal 行改为直接工作事实，移除报告链接，并同步 reports INDEX。

#### 6.78 High-confidence obsolete wording cleanup

清理 active docs、Bug 记录与 work-journal 中高置信旧生命周期直述，改为 current `out-of-scope`、直接删除事实或不再支持事实；docs lifecycle status enum 继续作为结构契约分类，不在本阶段迁移。

#### 6.79 Docs lifecycle status enum cleanup

删除 docs Header / INDEX 工具链、模板和 README 中的旧英文 / 中文生命周期状态枚举支持，使文档生命周期只保留当前结构状态集合；同步把内部 context discovery 的旧字段错误文案改为 unsupported，确保工具提示不重新引入旧生命周期口径。

#### 6.80 Remaining old-scope standalone report cleanup

删除 `docs/reports/` 中仍命中 broader old-scope wording、且经引用审计确认只被 reports INDEX 引用的 standalone 复盘报告实体，并同步 reports INDEX。仍承接当前 owner、Bug 或 work-journal 事实的证据不迁入替代报告。

#### 6.81 Active code scope-token naming cleanup

清理 active code、UI truth source 和 lint tooling 中剩余可改名的 scope-token 命名：测试变量改为 out-of-scope 命名，禁用输入字面量通过运行时拼接保留覆盖，lint bucket / 注释改为 current wording，Workspace 文案从 history/historical 改为 records / previous report signals，并同步 ui-design 与正式 frontend i18n。

#### 6.82 Local generated scope-token artifact cleanup

删除未跟踪本地生成物中的 scope-token artifact，包括 `.test-output` 下已不属于当前场景 slug 的旧运行目录，以及 `scripts/lint/__pycache__` 中旧 lint 模块名的 pyc 文件；不触碰依赖目录和仓库外部缓存。

#### 6.83 Active spec current-record wording cleanup

反查 `product-scope`、`frontend-workspace-and-practice`、`frontend-report-dashboard`、`engineering-roadmap`、`openapi-v1-contract` 与 `ci-pipeline-baseline` active spec，使当前用户可见记录能力统一使用会话记录 / 模拟面试记录 / records wording，使基础契约编号统一使用原始 ID / 既有约束表述；不改变运行时代码、OpenAPI、场景或 BDD 范围。

#### 6.137 Context manifest lineage metadata cleanup

清理 plan context 元数据中已无当前用途的空 lineage 字段：生成器不再推断或合并该字段，create/init/spec 模板和共享 context contract 示例不再要求该字段，所有 checked-in `context.yaml` 实例保持当前最小 metadata 形状。测试必须覆盖新生成与已有 manifest reconciliation 两条路径，repo direct grep 不得出现旧字段名。

#### 6.138 Current-owner handoff wording cleanup

清理 active code comments、OpenAPI tooling README、TargetJob owner spec 和 migrations README 中残留的旧计划交接说明：AI registry judge 默认错误只表达当前 fail-closed 状态，runtime config / review handler 注释描述当前 owner 协作，TargetJob spec 使用当前 backend internal runner 事实，migration README 用当前 B4 / lint / privacy matrix 作为 baseline inventory 来源。

#### 6.139 Skill execution terminology cleanup

统一 workflow skills、模板和 contract tests 的 current / out-of-scope / status-alias 术语；保留可执行负向 gate，不保留平行执行标签。

#### 6.140 Product-scope scope-boundary terminology cleanup

统一 product-scope spec、history 与 001 plan/checklist 的中文范围边界和英文 gate 命名；移除机械重复 token 与生命周期式说明，保持删除证据、隐私动作、迁移记录和负向 matcher 的原有语义，并同步 context specVersion 与 INDEX 投影。

## 5 Operation Matrix

| operationId / contract | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|------------------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getMyProfile` | delete `openapi/fixtures/Profile/getMyProfile.json` | remove `ProfileScreen` and generated client calls | delete profile handler | delete `candidate_profiles` | none after cleanup | delete `E2E.P0.091` |
| `updateMyProfile` | delete `openapi/fixtures/Profile/updateMyProfile.json` | remove profile correction UI | delete profile handler | delete `candidate_profiles` | remove `profile.update` if present | delete `E2E.P0.091` |
| `listExperienceCards` | delete `openapi/fixtures/Profile/listExperienceCards.json` | no consumer after cleanup | delete profile handler | delete `experience_cards` | none | delete `E2E.P0.092` |
| `createExperienceCard` | delete `openapi/fixtures/Profile/createExperienceCard.json` | no consumer after cleanup | delete profile handler | delete `experience_cards` + idempotency consumer | none | delete `E2E.P0.092` |
| `updateExperienceCard` | delete `openapi/fixtures/Profile/updateExperienceCard.json` | no consumer after cleanup | delete profile handler | delete `experience_cards` + idempotency consumer | none | delete `E2E.P0.092` |
| `createDebrief` | delete `openapi/fixtures/Debriefs/createDebrief.json` | delete DebriefScreen submit hook | delete debrief handler/job | delete `debriefs`, `task_runs` debrief usage, `source_debrief_id` | delete `debrief.generate` | delete `E2E.P0.060`, `E2E.P0.066` |
| `suggestDebriefQuestions` | delete `openapi/fixtures/Debriefs/suggestDebriefQuestions.json` | delete DebriefScreen suggestions hook | delete debrief handler | none after cleanup | delete `debrief.suggest_questions` | delete `E2E.P0.063`, `E2E.P0.066` |
| `getDebrief` | delete `openapi/fixtures/Debriefs/getDebrief.json` | delete DebriefScreen polling hook | delete debrief handler | delete `debriefs` | none after cleanup | delete `E2E.P0.061`, `E2E.P0.067` |
| `createPracticePlan goal=debrief` | update existing fixture scenarios only if present | remove debrief replay launcher | update practice handler/store to reject/omit debrief source goal | delete `practice_plans.source_debrief_id`, `goal='debrief'` | none | update `E2E.P0.070`, delete `E2E.P0.071`, `E2E.P0.073` |

## 6 Coverage Matrix

| Source | Category | Plan phase | Verification | Negative scope |
|--------|----------|------------|--------------|----------------|
| 用户确认方案 B | Primary path | Phase 1-5 | `E2E.P0.098`, `E2E.P0.099` | 核心链路不得依赖复盘或画像 |
| product-scope P0 改写 | Cross-layer contract | Phase 1 | `make docs-check`, sync-doc-index | `复盘` / `用户画像` 不得作为 active 产品能力出现 |
| UI 真理源 | UI source structure parity | Phase 1-2 | frontend route/topbar tests, pixel parity smoke, `ui-design/canvas.html` out-of-scope-artboard negative gate | `debrief` primary nav, `profile` user menu, positive out-of-scope design-canvas artboards |
| Out-of-scope route behavior | Regression / out-of-scope-negative | Phase 2 | `E2E.P0.088`, `E2E.P0.090`, routeUrl tests | `/debrief`, `#route=debrief_full`, `/profile` live page |
| API contract boundary | Cross-layer contract | Phase 3 | `make codegen-check`, `make validate-fixtures` | `Profile` / `Debriefs` tags, generated methods |
| DB contract boundary | Migration | Phase 4 | migration lint / focused DB tests | `debriefs`, `candidate_profiles`, `experience_cards`, `source_debrief_id` |
| Async/event contract boundary | Cross-layer contract | Phase 3-4 | shared jobs/events codegen/lint | `debrief.created`, `debrief.completed`, `debrief.generate` |
| AI config contract boundary | Cross-layer contract | Phase 4 | prompt/rubric/profile lint/eval inventory | `debrief.*`, `profile.update` feature keys |
| Privacy boundary | Privacy / security | Phase 4-5 | backend privacy delete tests, out-of-scope grep for out-of-scope profile cleanup hooks | account delete must still clean retained core data without candidate-profile runtime hooks |
| Scenario scope | Regression / out-of-scope-negative | Phase 5 | scenario INDEX and script verification | out-of-scope P0 debrief/profile scenarios must not remain Ready |
| Out-of-scope input package cleanup | Docs / governance | Phase 6.17 | `make docs-check`, `sync-doc-index --check`, target zero-reference grep | root out-of-scope product input, out-of-scope UI docs, out-of-scope executable subject |
| UI boundary doc cleanup | Docs / UI truth source | Phase 6.18 | `make docs-check`, `sync-doc-index --check`, target zero-reference grep, affected context validation | UI boundary document package references must not remain |
| Root product summary cleanup | Docs / governance | Phase 6.19 | `make docs-check`, `git diff --check`, targeted grep | root README and UI index summaries must not describe cleaned modules as current loop |
| Backend Resume active spec cleanup | Docs / backend contract | Phase 6.20 | backend-resume context validation, `make docs-check`, `sync-doc-index --check`, targeted grep | active backend-resume spec/context must not explain or discover pre-D20 version-tree operation/table surface as current behavior |
| Product Scope current boundary wording cleanup | Docs / product owner | Phase 6.21 | product-scope spec targeted grep, `make docs-check`, `sync-doc-index --check`, `git diff --check` | product owner spec current sections must describe current behavior through product contract and negative boundaries |
| Frontend Resume Workshop active spec cleanup | Docs / frontend contract | Phase 6.22 | frontend-resume-workshop context validation, targeted grep, `make docs-check`, `sync-doc-index --check` | active frontend-resume-workshop spec/context must not discover pre-D20 version-tree operation/schema surface as current behavior |
| UI Design current boundary wording cleanup | Docs / UI truth source | Phase 6.23 | UI docs targeted grep, `make docs-check`, `git diff --check` | active UI truth source docs must describe current targets through current pages and route-input boundaries |
| Engineering Roadmap current-workstream cleanup | Docs / roadmap owner | Phase 6.24 | engineering-roadmap targeted grep, context validation, `make docs-check`, `sync-doc-index --check` | roadmap current workstream map/context must describe current owners through the active execution map |
| OpenAPI / Shared / Mock contract current-contract cleanup | Docs / contract owner | Phase 6.25 | targeted grep, context validation, `openapi_inventory.py`, `sync-doc-index --check`, `make docs-check`, `git diff --check` | B2/B1/E1 active specs and context discovery must not explain or discover current contract through out-of-scope module / out-of-scope operation / out-of-scope-tooling history |
| Frontend Report Dashboard flat Resume contract cleanup | Docs / frontend contract | Phase 6.26 | frontend-report-dashboard context validation, targeted grep, `openapi_inventory.py`, `sync-doc-index --check`, `make docs-check`, `git diff --check` | active report dashboard spec/context must not explain or discover ResumeVersion / resumeVersionId / ResumeAsset / out-of-scope module surface as current behavior |
| Frontend Home / Parse current-contract cleanup | Docs / frontend contract | Phase 6.27 | frontend-home-job-picks-and-parse context validation, targeted grep, `sync-doc-index --check`, `make docs-check`, `git diff --check` | active Home / Parse spec must explain current behavior through Home, Parse, workspace handoff, and privacy gates |
| OpenAPI README / baseline inventory cleanup | Docs / contract owner | Phase 6.28 | targeted grep, `openapi_inventory.py`, `make docs-check`, `git diff --check` | OpenAPI README, baseline README, and diff config must describe current freeze by current inventory rather than out-of-scope modules or out-of-scope tooling |
| Event / outbox active inventory cleanup | Docs / contract owner | Phase 6.29 | event-and-outbox context validation, `events_inventory.py`, targeted grep, `make docs-check`, `git diff --check` | B3 active spec, bootstrap plan, and context must describe current event/job contract by current inventory rather than out-of-scope domain, job, field, or table history |
| Backend Async Runner active runtime cleanup | Docs / backend contract | Phase 6.30 | backend-async-runner context validation, targeted grep, `make docs-check`, `git diff --check` | active backend runtime spec/context must describe current runner scope by current handler inventory and generic out-of-scope-runner gate rather than out-of-scope job/domain history |
| Prompt / rubric active feature-key cleanup | Docs / AI contract | Phase 6.31 | prompt-rubric context validation, targeted grep, `prompt_lint.py`, `ai_profile_coverage.py`, `make docs-check`, `git diff --check` | active F3 spec/context must describe current 9-key prompt/rubric contract through current feature_key inventory |
| Frontend Shell current route catalog cleanup | Docs / frontend contract | Phase 6.32 | frontend-shell 001/004 context validation, targeted route-token grep, `make docs-check`, `git diff --check` | active D1 spec/context must describe current route catalog and out-of-scope alias normalization rather than out-of-scope route lists or out-of-scope wording |
| Frontend Workspace / Practice current owner cleanup | Docs / frontend contract | Phase 6.33 | frontend-workspace-and-practice 001/002 context validation, targeted route/API/token grep, `make docs-check`, `git diff --check` | active workspace/practice spec/history/context must describe current owner routes, embedded company insight, flat Resume binding, and three practice goals rather than independent insight API, versioned Resume, or out-of-scope goal explanations |
| Workspace insight implementation alias cleanup | Frontend / UI truth source / scenarios | Phase 6.34 | UI contract test, focused frontend route/workspace tests, frontend-workspace 001/002 context validation, targeted route/API/token grep | ui-design, formal frontend, owner plans, and P0 scenarios must keep company signal only as embedded workspace insight and must not materialize an independent route, API consumer, component export, or positive scenario path |
| UI architecture/module-map current boundary cleanup | Docs / UI truth source | Phase 6.35 | UI docs targeted wording grep, `make docs-check`, `git diff --check` | active UI truth source must describe current modules and out-of-scope route input fallback through current UI boundaries |
| UI design active docs directory wording cleanup | Docs / UI truth source | Phase 6.36 | `docs/ui-design` targeted wording grep, `make docs-check`, `git diff --check` | all active UI design docs must describe current shape and boundaries through current layout and route-input contracts, except global status enum values |
| Frontend route negative naming cleanup | Frontend tests / route comments | Phase 6.37 | focused frontend route/scope/visual-smoke tests, targeted route wording grep | current executable negative route gates must use out-of-scope route/path wording and continue proving out-of-scope inputs do not materialize |
| Engineering Roadmap current-map wording cleanup | Docs / roadmap owner | Phase 6.38 | engineering-roadmap spec targeted wording grep, context validation, `make docs-check`, `git diff --check`, pruning-surface lint | roadmap active spec must describe current execution map through current truth sources, owner specs, operation inventory, UI entry set, out-of-scope boundaries, and no-pending INDEX model |
| Engineering Roadmap plan governance cleanup | Docs / roadmap owner | Phase 6.39 | engineering-roadmap 001 plan/checklist/context targeted wording grep, context validation, `sync-doc-index --check`, `make docs-check`, `git diff --check` | roadmap completed owner plan and discovery must describe current governance contract without wave/backlog/module-recovery/technical-draft disposition narrative |
| Frontend Resume picker / locale cleanup | Frontend / UI truth source | Phase 6.40 | focused Resume picker / workspace / locale / Resume Workshop tests, UI contract test, targeted frontend/ui grep, `make docs-check`, `git diff --check`, pruning-surface lint | workspace resume picker must use active `listResumes` choices and current locale/UI files must not expose disabled placeholders, coming-soon dead components, or master / targeted / version-tree / branch display copy |
| Resume scenario asset cleanup | Scenarios / BDD docs | Phase 6.41 | scenario script contract test, shell syntax check, scenario INDEX directory check, targeted current-slug grep, `make docs-check`, `git diff --check`, pruning-surface lint | Resume scenario directories and owner BDD docs must describe current flat Resume behavior through current scenario slugs and flat Resume semantics |
| Executable out-of-scope gate naming cleanup | Frontend / backend / scripts / scenarios | Phase 6.42 | focused pytest / Go / Vitest gates, shell syntax check, exact out-of-scope-name grep, `make docs-check`, `git diff --check`, pruning-surface lint | executable negative gates must use current `out-of-scope` names and paths while preserving forbidden-input fixture coverage |
| Lint tooling terminology cleanup | Tooling / tests | Phase 6.43 | focused pytest / unittest lint suites, `make lint-core-loop-pruning-surface`, `git diff --check` | executable lint tools must report current out-of-scope terminology and keep forbidden-input matching coverage |
| Scenario and pixel wording cleanup | Scenarios / frontend visual tests / BDD docs | Phase 6.44 | targeted wording grep, scenario shell syntax, scenario contract test, `make docs-check`, `git diff --check` | active scenario and visual-test prose must describe negative gates as out-of-scope without direct forbidden-label exposition |
| Active code and prototype wording cleanup | Backend / ui-design / tooling | Phase 6.45 | focused Go tests, ui-design contract test, lint script tests, active non-test grep | active comments, mock data, and test labels must use current wording except regex forbidden-context matchers |
| Active docs/spec wording cleanup | Docs / governance | Phase 6.46 | active docs/spec wording grep, `sync-doc-index --check`, `make docs-check`, `git diff --check` | active specs and plans must use out-of-scope wording without changing history files or path contracts |
| Active test wording cleanup | Backend / frontend tests | Phase 6.47 | focused Go package tests, focused frontend Vitest, active residual grep | test names, variables, and failure messages must use out-of-scope wording while preserving forbidden values as constructed inputs |
| Spec history wording cleanup | Docs / history | Phase 6.48 | docs/spec full wording grep, `sync-doc-index --check`, `make docs-check`, `git diff --check` | spec-local history files must use out-of-scope wording while preserving chronology and evidence |
| Product-scope Chinese boundary cleanup | Docs / product owner | Phase 6.49 | product-scope wording grep, `sync-doc-index --check`, `make docs-check`, `git diff --check` | product owner spec/history must describe current contract and out-of-scope boundaries through current verification entries |
| Cross-owner Chinese out-of-scope precision cleanup | Docs / frontend README / scenarios | Phase 6.50 | targeted Chinese wording grep, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | active/completed owner docs and scenario README prose must describe current out-of-scope boundaries with precise input-boundary wording; lint forbidden matcher remains code-owned |
| Product-scope owner self-description cleanup | Docs / product owner | Phase 6.51 | product-scope plan/checklist/BDD wording grep, `validate_context.py`, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | owner plan artifacts must describe current cleanup through current boundaries and zero-reference gates without self-triggering process wording |
| UI design current action wording cleanup | Docs / UI truth source | Phase 6.52 | `docs/ui-design` targeted wording grep, `sync-doc-index --check`, `make docs-check`, `git diff --check` | active UI design docs must describe current auth continuation, route boundaries, and templates without self-triggering process wording |
| Frontend README and scenario current action wording cleanup | Docs / scenarios | Phase 6.53 | targeted frontend/scenario wording grep, `make docs-check`, `git diff --check` | current README and P0 scenario prose must describe auth continuation, marker cleanup, failure handling, and out-of-scope route inputs with current action wording |
| Scenario data and script current action wording cleanup | Scenarios | Phase 6.54 | targeted scenario data/script wording grep, shell syntax check, `make docs-check`, `git diff --check` | current scenario data and scripts must preserve negative inputs while using current variable names, cleanup output, and data-erasure wording |
| Frontend Home / Parse owner wording cleanup | Docs / frontend owner | Phase 6.55 | frontend-home owner wording grep, `validate_context.py`, `sync-doc-index --check`, `make docs-check`, `git diff --check` | Home / Parse owner docs must describe current auth continuation, route boundaries, generated-client contracts, and 0-hit gates |
| Prompt registry status / 9-key contract cleanup | AI contract / docs / tests | Phase 6.56 | F3 target wording grep, `make lint-prompts`, focused Go and Vitest tests, `validate_context.py`, `sync-doc-index --check`, `make docs-check`, `git diff --check` | Prompt registry owner docs and executable gates must describe current status enum, 9 canonical `multi` baselines, and current v2.15 spec assertion |
| Active executable / scenario wording cleanup | Backend / frontend / tooling / scenarios | Phase 6.57 | targeted executable/scenario wording grep, focused pytest/unittest/Go/Vitest gates, `make lint-rubrics`, shell syntax, `make docs-check`, `git diff --check` | Active executable comments, tests, lint scripts, and P0 Home/Parse scenario prose must express current scope without process-word explanations |
| Local-dev / config current wording cleanup | Docs / config / dev-stack | Phase 6.58 | targeted local-dev/config wording grep, `make lint-prompts`, shell syntax, context validation, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | Local-dev and config docs must express Postgres volume preflight, dev-doctor, AI profile aliases, and prompt hash rules through current terminology without process wording |
| Practice voice/workspace owner wording cleanup | Docs / frontend owner / voice owner | Phase 6.59 | targeted practice voice/workspace wording grep, context validation, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | Practice voice/workspace owner docs must express out-of-scope route/enum/input gates and owner co-location through current owner boundary wording |
| Workspace owner process-word cleanup | Docs / frontend owner | Phase 6.60 | targeted workspace process-word grep, context validation, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | Workspace owner docs must use current owner-boundary wording while preserving current records static area semantics |
| Report snapshot deletion | Reports / docs | Phase 6.61 | targeted report snapshot zero-reference grep, `make docs-check`, `git diff --check`, pruning-surface lint | redundant report snapshot files must not remain as standalone report docs or INDEX rows |
| Report module-positive cleanup | Reports / docs | Phase 6.62 | targeted report basename zero-reference grep, `make docs-check`, `git diff --check`, pruning-surface lint | unreferenced module-positive report files must not remain as standalone report docs or INDEX rows; work-journal-linked evidence remains intact |
| Active-doc pruning report cleanup | Reports / work-journal / docs | Phase 6.63 | targeted report basename zero-reference grep, `make docs-check`, `git diff --check`, pruning-surface lint | standalone report text must not preserve old module lifecycle narrative once owner plan evidence and BUG link carry the current gate |
| Product-scope owner note wording cleanup | Docs / product owner | Phase 6.64 | product-scope owner wording grep, context validation, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | owner plan/checklist must not preserve outdated lifecycle-note wording or stale privacy test names |
| Runtime/generated allowlist report cleanup | Reports / docs | Phase 6.65 | targeted report basename/title zero-reference grep, `make docs-check`, `git diff --check`, pruning-surface lint | standalone report text must not duplicate executable lint gate evidence already carried by the owner plan |
| June core-loop report cleanup | Reports / docs | Phase 6.66 | targeted report basename/title zero-reference grep, `make docs-check`, `git diff --check`, pruning-surface lint | standalone report text must not duplicate owner plan, BUG, and work-journal evidence for the same module cleanup |
| Standalone lifecycle-term report batch cleanup | Reports / docs | Phase 6.67 | generated candidate reference audit, targeted report basename zero-reference grep, `make docs-check`, `git diff --check`, pruning-surface lint | reports that only self-index old lifecycle-term discussion must not remain as standalone docs; externally referenced evidence stays intact |
| Work-journal referenced report cleanup | Reports / work-journal / docs | Phase 6.68 | targeted report basename zero-reference grep, `make docs-check`, `git diff --check`, pruning-surface lint | work-journal must not keep direct links to deleted standalone reports; historical fact lines remain without report indirection |
| Spec/report referenced lifecycle report cleanup | Reports / spec / work-journal / docs | Phase 6.69 | targeted report basename zero-reference grep, affected context validation, `make docs-check`, `git diff --check`, pruning-surface lint | spec plans and reports must not preserve links to deleted standalone lifecycle reports; facts remain directly in owner docs |
| Work-journal lifecycle-term wording cleanup | Work-journal / docs | Phase 6.70 | targeted work-journal lifecycle-term grep, broader scope-token/historical residual count, context validation, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | work-journal records must not preserve configured strict token or banner wording; broader scope-token/historical terminology remains a separately counted follow-up surface |
| Historical-spec report cleanup | Reports / bugs / work-journal / docs | Phase 6.71 | historical-spec report basename zero-reference grep, reports filename grep, `make docs-check`, `git diff --check`, pruning-surface lint | historical-spec reports must not remain as standalone report files, INDEX rows, BUG links, or work-journal report links |
| Active spec strict lifecycle-term cleanup | Docs / specs | Phase 6.72 | docs/spec strict lifecycle-term grep, context validation, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | active/completed spec subject artifacts must not preserve configured strict lifecycle-token wording outside schema/status enum docs |
| Lint rule source strict-token assembly cleanup | Tooling / lint tests | Phase 6.73 | focused lint tests, repo strict lifecycle-term grep, `make lint-core-loop-pruning-surface`, `git diff --check` | active lint scripts must keep detection behavior without direct configured strict token literals in source |
| Bug knowledge base strict lifecycle-term wording cleanup | Bugs / docs | Phase 6.74 | docs/bugs strict lifecycle-term grep, `make docs-check`, `git diff --check` | bug records may preserve evidence, but must not preserve direct configured strict lifecycle-token wording |
| Governance strict lifecycle-term wording cleanup | Governance / docs | Phase 6.75 | root strict lifecycle-term grep classification, `make docs-check`, `git diff --check` | governance instructions must not preserve direct configured strict lifecycle-token wording; structural enum and lockfile metadata stay classified |
| Broader old-scope standalone report cleanup | Reports / docs | Phase 6.76 | report basename zero-reference grep, reports INDEX sync, `make docs-check`, `git diff --check` | broader old-scope standalone reports that are only self-indexed must not remain as report entities |
| Work-journal-linked broader old-scope report cleanup | Reports / work-journal / docs | Phase 6.77 | report basename zero-reference grep, reports INDEX sync, `make docs-check`, `git diff --check` | work-journal-linked broader old-scope standalone reports must not remain once work-journal carries direct facts |
| High-confidence obsolete wording cleanup | Docs / bugs / work-journal | Phase 6.78 | obsolete wording grep with status-enum exception, `make docs-check`, `git diff --check` | high-confidence old lifecycle wording must use current out-of-scope or direct removal facts |
| Docs lifecycle status enum cleanup | Docs / skills / tooling | Phase 6.79 | focused sync-doc-index / implement / change-intake tests, status enum grep, `make docs-check`, `git diff --check` | docs Header / INDEX tooling and templates must not retain old lifecycle status enum or scope-token field wording |
| Remaining old-scope standalone report cleanup | Reports / docs | Phase 6.80 | report basename reference audit, reports INDEX sync, `make docs-check`, `git diff --check` | standalone report files with old-scope wording and no non-INDEX references must not remain as report entities |
| Active code scope-token naming cleanup | Frontend / ui-design / backend tests / lint tooling | Phase 6.81 | focused Go/Vitest/node/pytest lint tests, active code grep, `make docs-check`, `git diff --check` | active code and UI copy must not retain scope-token names where runtime-assembled negative inputs or current wording can carry the same gate |
| Local generated scope-token artifact cleanup | Local generated files | Phase 6.82 | generated artifact filename grep, `git status` scope check, `git diff --check` | untracked local test output and pycache files must not preserve scope-token directory or module names |
| Active spec current-record wording cleanup | Docs / specs | Phase 6.83 | targeted active-spec grep, context validation, `sync-doc-index --check`, `make docs-check`, `git diff --check` | active specs must use records / original-id / existing-constraint wording where those are current product concepts |
| Active spec broad stale-wording cleanup | Docs / specs | Phase 6.84 | targeted active-spec grep, spec INDEX sync, context validation, `sync-doc-index --check`, `make docs-check`, `git diff --check` | active spec.md files must not preserve direct stale lifecycle labels when current wording or out-of-scope boundary wording carries the same contract |
| Engineering roadmap ADR current-decision cleanup | Docs / ADR | Phase 6.85 | targeted ADR grep, `sync-doc-index --check`, `make docs-check`, `git diff --check` | accepted ADRs must express current email-code auth, local runner, current product loop, privacy, and backend AI owner boundaries without standalone obsolete-module narration |
| Backend runtime topology owner wording cleanup | Docs / backend runtime owner | Phase 6.86 | targeted owner grep, context validation, runtime-topology lint, `sync-doc-index --check`, `make docs-check`, `git diff --check` | completed worker-consolidation owner docs must describe current backend internal runner boundary without preserving standalone-process examples in current handoff prose |
| Local dev stack owner wording cleanup | Docs / local dev owner | Phase 6.87 | targeted owner grep, context validation, plans INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check` | completed dev-stack owner docs must express current rebaseline, existing gate, and email-code terminology without old lifecycle or obsolete auth-link wording |
| Backend resume 001 owner wording cleanup | Docs / backend resume owner | Phase 6.88 | targeted owner grep, context validation, plans INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check` | active backend-resume asset-register owner docs must describe existing chronology and stored rows without old lifecycle wording |
| Secrets and config owner wording cleanup | Docs / config owner | Phase 6.89 | targeted owner grep, context validation, config lint, plans INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check` | completed config owner docs must express current out-of-scope feature flag and existing-gate terminology without old lifecycle wording |
| Backend auth owner history-block cleanup | Docs / backend auth owner | Phase 6.90 | targeted owner grep, context validation, config lint, plans INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check` | completed auth owner docs must retain current email-code single-entry contract without registration-split phase records |
| Frontend shell owner history-block cleanup | Docs / frontend shell owner | Phase 6.91 | targeted owner grep, context validation, focused frontend tests, plans INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check` | completed frontend-shell owner docs must retain current topbar/auth/settings contract without mail-link or registration-split phase records |
| Prompt rubric schema owner wording cleanup | Docs / prompt-rubric owner | Phase 6.92 | targeted owner grep, context validation, prompt/rubric lint, plans INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check` | completed output-schema owner docs must express migration audit rows, seed net-state, and parser aliases without old lifecycle wording |
| Event outbox mode owner wording cleanup | Docs / event contract owner | Phase 6.93 | targeted owner grep, context validation, codegen/lint events, plans INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check` | completed ResumeTailorMode drift owner docs must express out-of-scope literal scanning and diff scope without old lifecycle wording |
| Backend async runner owner wording cleanup | Docs / backend runner owner | Phase 6.94 | targeted owner grep, context validation, runner lint, plans INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check` | completed runner owner docs must express out-of-scope runner/drainer and email-code smoke wording without old lifecycle wording |
| AI provider registry owner wording cleanup | Docs / AI provider owner | Phase 6.95 | targeted owner grep, context validation, AI profile coverage lint, config lint, plans INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check` | completed A3 provider-registry owner docs must express current 9-key profile coverage, DeepSeek chat baseline, and speech / judge profile boundaries without old lifecycle wording |
| Frontend resume workshop 003 owner compression | Docs / frontend resume owner | Phase 6.96 | owner package grep, context validation, focused Resume Workshop Vitest, plans INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check` | resume-workshop 003 owner docs must describe only current flat Resume/Rewrites/Edit contract and completed BDD gates, without keeping out-of-scope phase prose |
| Backend review spec and 001 owner compression | Docs / backend review owner | Phase 6.97 | active spec grep, owner package grep, context validation, plans/spec INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | backend-review active spec and 001 owner docs must describe current async-runner-backed report generation/read contract without stale baseline or staged implementation prose |
| Prompt rubric language coordinate owner compression | Docs / prompt-rubric owner | Phase 6.98 | owner package grep, context validation, prompt/rubric/profile lint, plans INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | prompt-rubric 003 owner docs must describe current 9-key canonical `multi` contract and executable evidence without stale coordinate inventory |
| OpenAPI bootstrap owner compression | Docs / OpenAPI contract owner | Phase 6.99 | owner package grep, context validation, OpenAPI lint/codegen gates, plans INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | openapi-v1-contract 001 owner docs must describe current 35-operation / 10-tag contract, codegen, docs renderer, and child handoff without stale endpoint/tag inventories |
| Frontend Shell 001 owner compression | Docs / frontend shell owner | Phase 6.123 | owner package grep, context validation, focused frontend shell Vitest, plans INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | frontend-shell 001 owner docs must describe only current App shell, email-code auth, settings, display, route guard and BDD gates, without keeping stale execution prose |
| Frontend Shell active spec/history compression | Docs / frontend shell owner | Phase 6.124 | active spec/history grep, frontend-shell context validation, spec INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | frontend-shell active spec and history must describe only the current App shell, email-code auth, settings, display, canonical URL and protected route guard contract |
| Frontend Shell 004 URL routing owner compression | Docs / frontend shell owner | Phase 6.125 | owner package grep, context validation, focused URL/auth privacy Vitest, plans INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | frontend-shell 004 owner docs must describe only current canonical URL, safe-param, hash adapter, privacy and host fallback contracts |
| Frontend Resume Workshop 001 owner compression | Docs / frontend resume owner / scenarios | Phase 6.126 | owner package grep, context validation, focused Resume Workshop Vitest, P0.036/P0.037 scenario scripts, scenario INDEX sync, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | frontend-resume-workshop 001 owner docs and P0.036 assets must describe only current flat list, `resumeId` detail preview, `listResumes` / `getResume` / `exportResume`, auth/privacy and scenario gates |
| Frontend fallback shell phase-label cleanup | Frontend / docs / scenarios | Phase 6.127 | red-green ScreensVisual test, focused App/P0.005/P0.036/P0.037 tests, targeted marker grep, `make docs-check`, `git diff --check`, pruning-surface lint | frontend fallback shell runtime, README, comments and scenario evidence must use current fallback-shell wording rather than old phase labels |
| Backend auth email-code owner and redaction contract cleanup | Backend / frontend / OpenAPI / shared / docs / scenarios | Phase 6.128 | old auth/path/redaction grep, backend-auth context validation, focused backend Go tests, focused frontend Vitest, events/openapi lint, P0.003 scenario scripts, `make docs-check`, `git diff --check`, pruning-surface lint | auth owner package, Go service naming, scenario slugs, OpenAPI summary and `email_dispatch` payload fields must use current email-code/code-only semantics |
| Engineering roadmap history net-state compression | Docs / roadmap owner | Phase 6.129 | targeted roadmap removed-workstream grep, product context validation, spec INDEX sync, `make docs-check`, `git diff --check`, pruning-surface lint | engineering-roadmap history must retain current net-state and counts without positive delivery prose for removed workstreams |
| Active plan/history process wording cleanup | Docs / specs | Phase 6.130 | high-confidence process/lifecycle wording grep, product context validation, sync-doc-index, `make docs-check`, `git diff --check`, pruning-surface lint | active plan/history/INDEX docs must use current owner, current-scope and follow-up wording without stale process labels or removed-workstream narration |
| Product/UI current-status wording cleanup | Docs / UI truth source | Phase 6.131 | targeted current-status grep, product context validation, sync-doc-index, `make docs-check`, `git diff --check`, pruning-surface lint | product owner records and current UI design docs must describe current status without stale status-transition wording or empty out-of-scope sections |
| Active spec/history current-status wording cleanup | Docs / specs | Phase 6.132 | targeted docs/spec current-status grep, product context validation, sync-doc-index, `make docs-check`, `git diff --check`, pruning-surface lint | active spec/history/owner docs must describe net-state without status-transition narration |
| Docs lifecycle tooling status cleanup | Skills / docs / tooling | Phase 6.133 | focused sync-doc-index/change-intake/implement tests, status residual grep, product context validation, sync-doc-index, `make docs-check`, `git diff --check`, pruning-surface lint | docs tooling, templates and README must support only `draft` / `active` / `completed` document status values |
| ADR decision lifecycle wording cleanup | Docs / roadmap ADRs | Phase 6.134 | targeted ADR status-transition grep, product context validation, spec INDEX sync, `make docs-check`, `git diff --check`, pruning-surface lint | accepted ADRs and current active specs must describe future decision changes as revision ADRs without stale status-transition wording |
| Prompt-rubric report/work-journal cleanup | Reports / work-journal / docs | Phase 6.135 | report basename zero-reference grep, targeted report/work-journal old-wording grep, sync-doc-index, `make docs-check`, `git diff --check`, pruning-surface lint | standalone report and historical journal prose must not preserve stale static-bridge lifecycle wording once owner history carries current facts |
| Residual skill/work-journal scope-token cleanup | Skills / work-journal / backend tests | Phase 6.136 | focused skill contract tests, broad scope-token grep, sync-doc-index, `make docs-check`, `git diff --check`, pruning-surface lint | skills, work-journal, and active test comments must use out-of-scope/removal wording instead of stale scope-token labels |
| Context manifest lineage metadata cleanup | Skills / docs / context manifests | Phase 6.137 | focused implement/change-intake context tests, batch context validation, direct old-field grep, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | plan context generator, templates, contract docs, and checked-in manifests must use current minimal metadata without empty lineage fields |
| Current-owner handoff wording cleanup | Backend comments / OpenAPI docs / TargetJob spec / migrations docs | Phase 6.138 | focused Go tests, targeted handoff-wording grep, spec INDEX sync, `make docs-check`, `git diff --check`, pruning-surface lint | active comments and docs must describe current owners and final inventory truth without old plan handoff prose |
| Skill execution terminology cleanup | `.agent-skills` workflow docs, templates and contract tests | Phase 6.139 | focused skill contract tests, `.agent-skills` scope-token grep, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | active skills and templates must use current/out-of-scope/status-alias wording instead of old execution labels |
| Product-scope scope-boundary terminology cleanup | Docs / product owner | Phase 6.140 | targeted owner token grep, product context validation, `sync-doc-index --check`, `make docs-check`, `git diff --check`, pruning-surface lint | product owner docs use one scope-boundary vocabulary without duplicate tokens or lifecycle-style module notes |

## 7 验收标准

- Product scope、engineering roadmap、UI 文档、静态原型和正式前端都只把 `首页 / 模拟面试 / 简历` 作为一级业务入口。
- 用户菜单不再出现 `用户画像`，设置与隐私仍可进入。
- OpenAPI inventory、fixtures、generated Go/TS artifacts 不再包含 `Profile` / `Debriefs` tags 或对应 operationId。
- Backend、migrations、shared、config 不再包含运行时复盘或候选人画像领域。
- 核心 BDD 场景仍能证明 JD / 简历 -> 模拟面试 -> 报告 -> 复练 / 下一轮闭环。
- 复盘 / 画像范围外 route、testid、table、event、job、feature key、prompt/rubric、scenario 通过负向搜索归零；历史 work-journal、bug、report 记录可保留为历史上下文，但不得作为 active truth source。

## 8 风险与应对

| 风险 | 应对措施 |
|------|----------|
| `profile` 与认证资料补全命名混用 | 只删除候选人画像 / experience card 领域；账号资料补全保留，并在文档中改称账号资料或资料补全 |
| 复盘 goal 已进入 practice 派生链路 | Phase 3/4 同步删除 `goal=debrief`、`source_debrief_id` 和 debrief first-question reservation，并补充 practice focused tests |
| OpenAPI 删除导致 generated consumer 大面积编译失败 | 先写 contract red tests / inventory gate，再代码删除并运行 codegen，最后修 frontend/backend consumers |
| 历史场景索引保留 Ready 状态 | Phase 5 删除正向场景目录和 INDEX 行，保留核心闭环场景的替代覆盖 |
| 文档把范围外能力写为 P1/P2 自动纳入对象 | Phase 1 和 zero-reference gate 覆盖 product-scope、engineering-roadmap、docs/ui-design、docs/spec/INDEX |

## 9 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-10 | 1.141 | Normalize product owner scope-boundary terminology, remove duplicate tokens, and sync context/index metadata. |
| 2026-07-10 | 1.140 | Rename workspace records-area wording to records static affordance across owner and pruning gates. |
| 2026-07-07 | 1.139 | Clean current skill execution terminology across workflow docs, templates, sync tooling labels and contract tests. |
| 2026-07-07 | 1.138 | Reconcile active code comments, OpenAPI tooling docs, TargetJob owner boundaries and migration README to current-owner wording. |
| 2026-07-07 | 1.137 | Remove unused context manifest lineage metadata from generators, templates, shared contract docs and checked-in plan contexts. |
| 2026-07-07 | 1.136 | Normalize residual skill, work-journal and active-test scope-token wording to out-of-scope/removal language. |
| 2026-07-07 | 1.135 | Remove prompt-rubric standalone report and old static-bridge lifecycle wording from history records. |
| 2026-07-07 | 1.134 | Remove stale status-transition wording from accepted ADRs and related active specs. |
| 2026-07-07 | 1.133 | Remove old document status support from docs tooling, templates and README files. |
| 2026-07-07 | 1.132 | Remove high-confidence status-transition wording from active spec/history owner docs. |
| 2026-07-07 | 1.131 | Remove stale status wording from product owner records and current UI design docs. |
| 2026-07-07 | 1.130 | Remove high-confidence stale process labels and removed-workstream narration from active plan/history/INDEX docs. |
| 2026-07-07 | 1.129 | Compress engineering-roadmap history to current net-state and remove positive delivery prose for removed workstreams. |
| 2026-07-07 | 1.128 | Rename backend-auth active owner, service, scenarios, OpenAPI summary, and email_dispatch redaction contract to current email-code semantics. |
| 2026-07-07 | 1.127 | Remove old frontend fallback-shell phase labels from runtime text, comments, README, pixel parity prose and scenario evidence. |
| 2026-07-07 | 1.126 | Compress frontend-resume-workshop 001 owner docs and P0.036 scenario slug to the current flat Resume list/detail preview contract. |
| 2026-07-07 | 1.125 | Compress frontend-shell 004 URL routing owner docs to the current canonical URL, safe-param, hash adapter, privacy and host fallback contract. |
| 2026-07-07 | 1.124 | Compress frontend-shell active spec and history to the current App shell, email-code auth, settings, display, canonical URL and route-guard contract. |
| 2026-07-07 | 1.123 | Compress frontend-shell 001 owner docs to the current App shell, email-code auth, settings, display and route-guard contract. |
| 2026-07-07 | 1.122 | Compress frontend-resume-workshop active spec and change log to the current flat Resume Workshop contract and completed 002/003 owner state. |
| 2026-07-07 | 1.121 | Rename and compress frontend-resume-workshop 002 into the completed current CreateFlow owner, update references, and remove stale create-flow contract prose. |
| 2026-07-07 | 1.120 | Rename and compress db-migrations-baseline 002 into the completed flat Resume migration owner, update references, and remove stale owner package naming. |
| 2026-07-07 | 1.119 | Compress db-migrations-baseline active spec and change log to the current net-state schema, flat Resume binding, privacy matrix, backfill ledger, and executable migration gates. |
| 2026-07-07 | 1.118 | Compress backend-practice active spec and 001 owner docs to the current PracticePlans / PracticeSessions contract, complete flat Resume first-question context, and close 001 lifecycle. |
| 2026-07-07 | 1.117 | Compress frontend-home-job-picks-and-parse active spec and 001 owner docs to the current Home / Parse generated-client contract and P0.014-P0.016 gates. |
| 2026-07-07 | 1.116 | Compress openapi-v1-contract 004 owner docs to the current flat Resume / ResumeTailor 35-operation contract and generated drift gates. |
| 2026-07-07 | 1.115 | Reconcile frontend-workspace-and-practice spec/history to current completed 001/002 owner plans and external voice/report owner gates. |
| 2026-07-07 | 1.114 | Compress frontend-workspace-and-practice 002 owner docs to the current text event loop, resumeId handoff, generated-client boundary and P0.044-P0.047 gates. |
| 2026-07-07 | 1.113 | Compress prompt-rubric-registry 001 owner docs to the current 9-key registry, lint, TargetJob adapter and provenance contract. |
| 2026-07-07 | 1.112 | Compress frontend-shell 002 owner docs to the current ui-design-native visual system contract and visual-smoke gates. |
| 2026-07-07 | 1.111 | Compress shared-conventions 001 owner docs to the current conventions truth source, generator, Go/TS helper and local lint contract. |
| 2026-07-07 | 1.110 | Compress AI provider 001 owner docs to the current AIClient, provider registry, model profile and observability foundation contract. |
| 2026-07-07 | 1.109 | Compress backend-practice 002 owner docs to the current append event, completion, idempotency, event/job and privacy contract. |
| 2026-07-07 | 1.108 | Compress event-and-outbox 001 owner docs and active spec to the current 14-event / 8-job event-job contract. |
| 2026-07-07 | 1.107 | Compress prompt-rubric-registry 004 owner docs to the current judge/default and 36-case eval-offline contract, and sync prompt-rubric spec-version preflight. |
| 2026-07-07 | 1.106 | Compress mock-contract-suite 001 owner docs and active spec to the current 35-operation fixture-backed mock runtime contract. |
| 2026-07-07 | 1.105 | Compress openapi-v1-contract 002 owner docs to the current 35-operation fixture truth source and sync generated fixture baselines. |
| 2026-07-07 | 1.104 | Compress frontend-shell 003 owner docs to the current 13-spec Playwright pixel parity contract. |
| 2026-07-07 | 1.103 | Compress frontend-workspace-and-practice 001 owner docs and scenario gates to the current workspace, flat Resume Picker, start-practice and records static area contract. |
| 2026-07-07 | 1.102 | Compress openapi-v1-contract 003 owner docs to the current 35-operation breaking-change gate. |
| 2026-07-07 | 1.101 | Compress backend-practice 003 owner docs to the current hint mode policy and provenance contract. |
| 2026-07-07 | 1.100 | Rename and compress backend-resume 002 owner docs to the current flat Resume tailor/save contract. |
| 2026-07-07 | 1.99 | Compress openapi-v1-contract 001 owner docs to the current 35-operation / 10-tag OpenAPI contract. |
| 2026-07-07 | 1.98 | Compress prompt-rubric-registry 003 owner docs to the current 9-key canonical multi contract. |
| 2026-07-07 | 1.97 | Compress backend-review active spec and 001 owner docs to the current async-runner-backed report generation/read contract. |
| 2026-07-07 | 1.96 | Compress frontend-resume-workshop 003 owner docs to the current flat Resume Workshop Rewrites/Edit contract. |
| 2026-07-07 | 1.95 | Reconcile ai-provider-and-model-routing 003 owner docs to current 9-key profile coverage and DeepSeek chat baseline wording. |
| 2026-07-07 | 1.94 | Reconcile backend-async-runner owner docs to current out-of-scope runner/drainer and email-code smoke wording. |
| 2026-07-07 | 1.93 | Reconcile event-and-outbox ResumeTailorMode owner docs to current out-of-scope literal scanning and diff-scope wording. |
| 2026-07-07 | 1.92 | Reconcile prompt-rubric output-schema owner docs to current migration audit, seed net-state, and parser alias wording. |
| 2026-07-07 | 1.91 | Delete frontend-shell auth history blocks and reconcile owner docs to current topbar/auth/settings contract. |
| 2026-07-07 | 1.90 | Delete backend-auth registration-split phase records and reconcile auth owner wording to current email-code contract. |
| 2026-07-07 | 1.89 | Reconcile secrets-and-config bootstrap owner docs to current out-of-scope feature flag and existing-gate wording. |
| 2026-07-07 | 1.88 | Reconcile backend-resume asset-register owner docs away from old lifecycle wording. |
| 2026-07-07 | 1.87 | Reconcile local-dev-stack bootstrap owner docs to current rebaseline, existing gate, and email-code wording. |
| 2026-07-07 | 1.86 | Reconcile backend-runtime-topology worker-consolidation owner docs and context to current backend internal runner wording. |
| 2026-07-07 | 1.85 | Reconcile engineering-roadmap ADR decisions to current email-code auth, local-runner deployment, product-loop analytics, privacy, and backend AI owner boundaries. |
| 2026-07-07 | 1.84 | Reconcile active spec.md broad stale wording across foundation, frontend, backend, AI, contract, and migration owner specs. |
| 2026-07-07 | 1.83 | Reconcile active specs to current records wording, original-id labels, and existing-constraint phrasing. |
| 2026-07-07 | 1.82 | Delete untracked local generated scope-token artifacts from test output and pycache. |
| 2026-07-07 | 1.81 | Reconcile active code, lint tooling, and Workspace UI copy away from scope-token naming. |
| 2026-07-07 | 1.80 | Delete remaining old-scope standalone reports referenced only by reports INDEX. |
| 2026-07-07 | 1.79 | Remove obsolete docs lifecycle status enum and rename context unsupported-field diagnostics. |
| 2026-07-07 | 1.78 | Reconcile high-confidence obsolete wording across active docs, Bug records, and work-journal entries. |
| 2026-07-07 | 1.77 | Delete work-journal-linked broader old-scope standalone reports and replace links with direct facts. |
| 2026-07-07 | 1.76 | Delete broader old-scope standalone reports referenced only by reports INDEX. |
| 2026-07-07 | 1.75 | Reconcile root governance wording and classify remaining structural enum and lockfile metadata hits. |
| 2026-07-07 | 1.74 | Reconcile Bug knowledge base wording away from configured strict lifecycle-term tokens while preserving evidence records. |
| 2026-07-07 | 1.73 | Reconcile lint rule source literals while preserving strict lifecycle-token detection through tests. |
| 2026-07-07 | 1.72 | Reconcile active spec subject artifacts away from configured strict lifecycle-term wording while preserving schema enum docs. |
| 2026-07-07 | 1.71 | Delete remaining historical-spec standalone reports and replace report links with direct evidence facts. |
| 2026-07-07 | 1.70 | Replace work-journal strict lifecycle-token wording with direct facts and current negative-gate terminology. |
| 2026-07-07 | 1.69 | Delete remaining spec/report referenced lifecycle reports and replace links with direct owner facts. |
| 2026-07-07 | 1.68 | Delete work-journal referenced standalone reports and replace report links with direct work facts. |
| 2026-07-07 | 1.67 | Delete standalone lifecycle-term report batch and prune reports INDEX references. |
| 2026-07-07 | 1.66 | Delete redundant June core-loop reports and prune reports INDEX references. |
| 2026-07-07 | 1.65 | Delete redundant runtime/generated allowlist report and prune reports INDEX reference. |
| 2026-07-07 | 1.64 | Reconcile product-scope owner lifecycle-note wording and stale privacy test evidence to current terms. |
| 2026-07-07 | 1.63 | Delete active-doc pruning report narrative and prune reports INDEX plus work-journal link. |
| 2026-07-07 | 1.62 | Delete unreferenced module-positive report records and prune reports INDEX references. |
| 2026-07-07 | 1.61 | Delete redundant report snapshots and prune reports INDEX references. |
| 2026-07-07 | 1.60 | Reconcile workspace owner process wording to current owner-boundary and active-list boundary terms. |
| 2026-07-07 | 1.59 | Reconcile practice-voice-mvp and frontend-workspace-and-practice owner docs to current out-of-scope route/enum/input and owner-boundary wording. |
| 2026-07-07 | 1.58 | Reconcile local-dev-stack, dev-stack README, config README, and prompt hash documentation to current wording without behavior changes. |
| 2026-07-07 | 1.57 | Reconcile active executable comments, lint docs, tests, pixel specs, and Home/Parse scenario prose to current-scope wording without behavior changes. |
| 2026-07-07 | 1.56 | Reconcile prompt registry status enum, 9-key canonical baseline contract, profile catalog count, and executable preflight assertions to current F3 v2.15 truth. |
| 2026-07-07 | 1.55 | Reconcile frontend-home-job-picks-and-parse owner docs to current Home / Parse BDD matrix, generated-client contracts, route boundaries, and 0-hit gates. |
| 2026-07-07 | 1.54 | Reconcile current scenario data and shell scripts to current action wording while preserving negative input values and shell behavior. |
| 2026-07-07 | 1.53 | Reconcile frontend README and current scenario prose to current action wording for auth continuation, cleanup markers, failure handling, and out-of-scope route inputs. |
| 2026-07-07 | 1.52 | Reconcile active UI design docs, README, templates, and INDEX to current action and boundary wording. |
| 2026-07-07 | 1.51 | Reconcile product-scope owner plan, checklist, and BDD self-description wording to current boundary and zero-reference language after the cross-owner cleanup pass. |
| 2026-07-07 | 1.50 | Reconcile cross-owner Chinese boundary and out-of-scope phrasing wording to precise out-of-scope boundary language while leaving privacy data-erasure, migration drop, cleanup actions, and lint forbidden matchers intact. |
| 2026-07-07 | 1.49 | Reconcile product-scope Chinese boundary wording to current contract language while preserving privacy data-erasure and records semantics. |
| 2026-07-06 | 1.48 | Reconcile docs/spec history wording to out-of-scope terminology while preserving chronology and evidence references. |
| 2026-07-06 | 1.47 | Reconcile backend and frontend active test names, variables, messages, and fixture samples to out-of-scope wording while preserving forbidden-value coverage. |
| 2026-07-06 | 1.46 | Reconcile active docs/spec wording outside history files from forbidden English labels to `out-of-scope` without changing paths or status enums. |
| 2026-07-06 | 1.45 | Reconcile active Go comments, ui-design mock data/tests, markdown link tooling docs, and backend-practice lint samples to current out-of-scope wording. |
| 2026-07-06 | 1.44 | Reconcile active scenario, pixel-parity, and related BDD owner wording to `out-of-scope`, with shell forbidden values kept as constructed inputs only. |
| 2026-07-06 | 1.43 | Reconcile active lint/tooling class names, test names, bucket output, and error messages to `out-of-scope` terminology while preserving forbidden-input matching. |
| 2026-07-06 | 1.42 | Rename executable negative gate files, Make targets, scenario directories, Go/Vitest test names, and lint scripts to current `out-of-scope` wording with focused verification. |
| 2026-07-06 | 1.41 | Rename and reconcile Resume BDD scenario directories, INDEX rows, scripts, and owner BDD docs to current flat Resume semantics without historical labels. |
| 2026-07-06 | 1.40 | Reconcile frontend workspace Resume picker, Resume Workshop locale keys, dead coming-soon/placeholder components, and UI contract wording to current flat Resume behavior. |
| 2026-07-06 | 1.39 | Reconcile engineering-roadmap 001 plan, checklist, and context discovery to current execution-map governance and technical-draft zero-reference wording. |
| 2026-07-06 | 1.38 | Reconcile engineering-roadmap active spec wording to current execution-map language, current owner boundaries, and no-pending INDEX model. |
| 2026-07-06 | 1.37 | Reword frontend route negative tests and route comments to out-of-scope route/path terminology while preserving behavior and focused route test coverage. |
| 2026-07-06 | 1.36 | Reconcile all active docs/ui-design wording to current shape and boundary language; leave only the global Header status metadata. |
| 2026-07-06 | 1.35 | Reconcile UI architecture and module-map wording to current module ownership and out-of-scope route-input fallback through current UI boundaries. |
| 2026-07-06 | 1.34 | Clean standalone company insight alias/API/component wording from ui-design, frontend routing, workspace components, owner plans, and P0 scenario assets; keep company signal as embedded workspace insight only. |
| 2026-07-06 | 1.33 | Reconcile frontend-workspace-and-practice active spec, history, and 001/002 contexts to current owner route, embedded insight, flat Resume, and current practice-goal wording. |
| 2026-07-06 | 1.32 | Reconcile frontend-shell active spec and 001/004 contexts to current route catalog, auth flow, safe params, and out-of-scope alias normalization wording. |
| 2026-07-06 | 1.31 | Reconcile prompt-rubric-registry active spec and 002 context to current 9-key prompt/rubric contract wording through current feature_key inventory. |
| 2026-07-06 | 1.30 | Reconcile backend-async-runner active spec and context to current single-kernel runtime wording with 7 executable handlers and `privacy_export` contract-only. |
| 2026-07-06 | 1.29 | Reconcile event-and-outbox active spec, bootstrap plan, and context to current 14-event / 8-job inventory wording without out-of-scope domain, job, field, or table explanations. |
| 2026-07-06 | 1.28 | Reconcile OpenAPI README, baseline README, and diff config to current 10 tag / 35 operation inventory wording without current boundary or out-of-scope-tooling explanations. |
| 2026-07-06 | 1.27 | Reconcile frontend-home-job-picks-and-parse active spec to the current Home + Parse contract and data-erasure truth source. |
| 2026-07-06 | 1.26 | Reconcile frontend-report-dashboard active spec, plan, checklist, and context to current flat Resume `getResume(resumeId)` contract and current boundary wording. |
| 2026-07-06 | 1.25 | Reconcile OpenAPI, Shared Conventions, and Mock Contract active specs and B2 context discovery to current executable contract wording without out-of-scope module or out-of-scope-tooling explanations. |
| 2026-07-06 | 1.24 | Reconcile engineering-roadmap active spec to current workstream and out-of-scope boundary wording through the active execution map. |
| 2026-07-06 | 1.23 | Reconcile active UI design truth-source wording to current targets and out-of-scope boundaries. |
| 2026-07-06 | 1.22 | Reconcile frontend-resume-workshop active spec and contexts to the current flat Resume UI contract and remove explanatory version-tree operation/schema leftovers. |
| 2026-07-06 | 1.21 | Reconcile product-scope active spec current boundary wording so current behavior is expressed as product contract and negative boundaries. |
| 2026-07-06 | 1.20 | Reconcile backend-resume active spec to the current flat Resume contract and remove explanatory version-tree leftovers from current behavior sections. |
| 2026-07-06 | 1.19 | Align root README and UI index summaries with the current core loop; remove stale debrief/profile product-loop wording from first-screen docs. |
| 2026-07-06 | 1.18 | Move remaining route alias and module boundary evidence to current UI/product truth sources after UI boundary document package cleanup. |
| 2026-07-06 | 1.17 | Rename backend-practice/004 to report-derived owner and keep prohibited source/goal terms only as negative inputs. |
| 2026-07-06 | 1.16 | Retarget root out-of-scope product input, out-of-scope UI docs, and out-of-scope orchestration references to current product/UI truth sources. |
| 2026-07-06 | 1.15 | Add executable runtime/generated allowlist audit to bucket migrations, out-of-scope normalization, negative tests, and real residuals before further pruning. |
| 2026-07-06 | 1.14 | Delete obsolete Debrief / Profile / Jobs Recommendations subject directories and the out-of-scope frontend JD Match plan package; update indexes, links, context audit, and current owner evidence without preserving out-of-scope standalone notes. |
| 2026-07-06 | 1.13 | Reopen to delete unambiguous obsolete module document packages instead of preserving out-of-scope standalone notes. |
| 2026-07-06 | 1.12 | Reconcile runtime/generated/config negative sweep: remove stale JD Match CSS assets, add frontend scope guard, and record the migration enum-source squash/lint-model decision boundary. |
| 2026-07-06 | 1.6 | Reconcile completed backend infra / AI contract plans after D-22 pruning: `backend-async-runner/001` current runner scope is 7 executable handlers + `privacy_export` contract-only; `prompt-rubric-registry/002` current prompt/rubric scope is 9 chat feature_keys. |
| 2026-07-06 | 1.5 | Reconcile active-subject completed `backend-practice/004-report-derived-practice-plans`: current positive scope is report-derived retry / next-round only; out-of-scope source fields and P0.071/P0.073 are negative-only. |
| 2026-07-06 | 1.4 | Reconcile completed owner plan context manifests after D-22 pruning: current `uiRoutes` / `apiNames` now list only retained routes and 35 live OpenAPI operationIds; out-of-scope debrief/profile operations remain negative/search terms only. |
| 2026-07-06 | 1.3 | Reopen for active document-drift cleanup after D-22 pruning: reconcile B3 event/job spec, frontend workspace/practice active docs, out-of-scope debrief lifecycle projection, and active-doc out-of-scope-negative gates. |
| 2026-06-30 | 1.2 | Review remediation: reconcile stale AI profile specs, OpenAPI operation-count tests, event schema-count tests, and frontend envelope fixture expectations after D-22 pruning. |
| 2026-06-29 | 1.1 | L2 review remediation: reopen cross-layer cleanup to remove stale design-canvas Profile/Debrief artboards and privacy runner profile cleanup hook drift. |
