# Core Loop Active Doc Pruning 交付复盘报告

> **日期**: 2026-07-06
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付范围是 `product-scope/001-core-loop-module-pruning` Phase 6 active 文档与 runtime 漂移清理：对齐 B3 event/job、frontend workspace/practice、deprecated debrief lifecycle、mock / AI / targetjob / P0 E2E / OpenAPI downstream 文档，以及 B4 migration baseline / privacy matrix / migration README、frontend-shell 三入口与 URL routing plans，并补做 runtime / generated / config 负向审计，使已删除的 debrief、profile、JD Match、独立 company_intel 不再作为当前正向 owner、future work 或正式前端资产出现。

成功证据：

- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS，Header / INDEX / orphan 均零漂移。
- `make docs-check` PASS，`docs/` 与 `docs/spec/` 相对链接检查均通过。
- `git diff --check` PASS。
- `python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml` PASS，当前 10 tags / 35 operations。
- `python3 scripts/lint/events_inventory.py shared/events.yaml shared/jobs.yaml` PASS。
- `go test ./backend/internal/migrations -count=1` PASS，并以红绿测试补齐 `idempotency_records` privacy matrix 覆盖。
- `python3 scripts/lint/migrations_lint.py --repo-root .` PASS。
- `validate_context.py` 对 `frontend-workspace-and-practice` plans 001 / 002 PASS。
- `pnpm --filter @easyinterview/frontend test src/app/screens/workspace/WorkspaceHandoff.test.tsx src/app/screens/practice/__tests__/practiceGoalParity.test.tsx src/app/screens/practice/hooks/usePracticeAssistance.test.ts` PASS，3 files / 16 tests。
- `pnpm --filter @easyinterview/frontend test src/app/topbar/TopBar.test.tsx src/app/normalizeRoute.test.ts src/app/routeUrl.test.ts` PASS，3 files / 52 tests，证明当前 TopBar 三入口、retired route normalization 与 URL codec 行为仍和文档收敛一致。
- `pnpm --filter @easyinterview/frontend test src/app/auth/AuthScreens.test.tsx src/app/auth/AuthVisual.test.tsx src/app/AppAuthDispatch.test.tsx src/app/topbar/TopBar.test.tsx src/app/normalizeRoute.test.ts src/app/routeUrl.test.ts` PASS，6 files / 93 tests；`validate_context.py` 对 `frontend-shell` plans 001 / 002 PASS，证明 `auth_register` / `auth_reset`、`profile`、`debrief`、`jd_match`、`company_intel` 仅保留为 legacy / retired 负向对象，当前 context `uiRoutes` 已收敛到三入口、settings、auth 和保留业务 route。
- `validate_context.py` 对 `product-scope/001-core-loop-module-pruning` 与 `mock-contract-suite/001-fixture-backed-mock-runtime` PASS，且两份 context 的 `apiNames` 与当前 `openapi/openapi.yaml` 35 个 `operationId` 精确一致（`missing=[]`、`extra=[]`），旧 Profile / Debriefs operations 不再作为正向 discovery surface。
- `product-scope/001-core-loop-module-pruning` context 的正向 `uiRoutes` 已从旧 `debrief` / `profile` 收敛到当前保留 routes；旧 key 只保留在 discovery keywords 和 retired-negative 搜索口径中。
- `backend-practice/004-derived-plans-debrief` 已从历史 debrief seeding plan 原地收敛为 report-derived retry / next-round plan：当前 BDD 只保留 P0.070 / P0.072，历史 P0.071 / P0.073 和 `sourceDebriefId` / `PracticeGoalDebrief` 只作为 retired-negative 说明。
- `backend-async-runner/001-internal-job-outbox-runner` 已从旧 9-job runner 口径收敛为当前 7 个可执行 handler + `privacy_export` contract-only，context 不再发现 `backend/internal/debrief`、`backend/internal/jdmatch` 或 deprecated jobs-recommendations owner。
- `prompt-rubric-registry/002-output-schema-contract` 已从旧 13 chat feature_key 口径收敛为当前 9 个 chat feature_key，context 不再发现 retired Debrief / JD Match backend packages，plan/checklist 只把这些 key 作为 retired-negative 或 revision-history 说明。
- `migrations/README.md` 已从旧 “goose / atlas / sqlc 后续定型” 改为当前 `golang-migrate` + `backend/cmd/migrate` + `make migrate-*` 入口，并用负向搜索确认旧待定文案零残留。
- `db-migrations-baseline` 已从旧 B3 canonical 9-job 口径修正为当前 8-job truth source；`TestDropJDMatchMigrationDropsRetiredTablesAndRegistryRows` 与既有 async job check 共同证明 `000014_drop_jd_match_module` 删除 5 张 JD Match 表、清理 `jd_match.*` prompt/rubric registry rows，并在收窄 `async_jobs.job_type` check 前删除退休 job rows。
- `event-and-outbox-contract/001-bootstrap` completed plan 已从旧 16-event / 9-job / 7 API-facing 口径收敛为当前 14-event / 8-job / 6 API-facing contract；`db-migrations-baseline/001-bootstrap` completed plan 已从旧 25-app / 30-public-table / 9-job / 7 API-facing 口径收敛为当前 22-app + 3-auth / 27-public-table / 8-job / 6 API-facing contract。两份 context 均通过 validate_context。
- Deprecated `backend-profile/001` 与 `backend-jobs-recommendations/001` 已补齐退役历史记录 banner；两份 `context.yaml` 不再把旧 operationId、旧 backend package、旧 fixtures、旧 prompt/rubric 或旧场景当作正向 target surface，只保留 retired keyword / negative audit discovery。
- Deprecated `backend-debrief/001` 与 `frontend-debrief/001` 已完成 context 收敛：旧 Debriefs operationId、`backend/internal/debrief`、`frontend/src/app/screens/debrief`、Debriefs fixtures、debrief prompt/rubric 和 `debrief` route 不再作为正向 target surface；backend-debrief 的 plan/checklist/test/BDD 文件补齐退役历史记录 banner。
- Repo-wide context structured audit 覆盖全部 57 个 `docs/spec/**/context.yaml`：旧 Debrief / Profile / JobMatch operationId、route、backend/frontend package、fixtures、prompt/rubric path 均未出现在正向 `packages` / `uiRoutes` / `apiNames` 字段中；剩余命中仅为 deprecated subject 自身文档目录路径或 retired/negative keyword。
- Runtime / generated / config 负向审计确认旧模块命中只剩负向测试、历史迁移、lint guards、legacy alias normalization 或 retired scenario 断言；删除 `frontend/src/app/theme/global.css` 中无人引用的 `jdmatch-*` responsive class，并新增 `frontend/src/app/scope.test.ts` guard。红灯：`npm test -- src/app/scope.test.ts` 失败于旧 CSS；绿灯：同命令 PASS，5 tests。
- `migrations/enum-sources.yaml` 仍保留 `agent_scans.status` / `watchlist_items.tone` 的历史 JD Match enum source，是因为当前 `scripts/lint/migrations_lint.py` 仍扫描全部 historical up migrations，`000009_jd_match_baseline.up.sql` 仍含对应 check；本轮不直接删除，避免 migration gate 从“最终态清理”误退化为“历史链不一致”。后续需要用户确认 pre-launch migration squash，或先把 migration lint 改造为最终态 schema 模型。

## 2 会话中的主要阻点/痛点

- Completed checklist 不能代表 active 文档仍然一致。
  - **证据**：`event-and-outbox-contract` active spec 曾保留旧 18-event / 11-job 口径；`frontend-workspace-and-practice` active docs 曾有 `goal='debrief'` 语义；`mock-contract-suite` completed plan/context 曾保留 12 tag、34 operation 与旧 Profile / Debriefs 正向 operations。
  - **影响**：如果只相信历史 PASS，会让后续 owner 从 stale spec 派生已删除模块。

- Deprecated subject 的 plan lifecycle 表达不够统一。
  - **证据**：`frontend-debrief` subject 已 deprecated，但把历史 plan 文件 Header 改为 deprecated 后，`sync-doc-index` 将它们判为 orphan；`backend-profile` / `backend-jobs-recommendations` 采用 subject deprecated + plan completed 的既有模式。
  - **影响**：容易在“历史审计记录”和“当前可执行计划”之间产生索引漂移。

- 文档标题语义变化会破坏历史 plan anchor。
  - **证据**：B3 spec 标题从旧事件全集更新为 14-event 后，`make docs-check` 发现 `event-and-outbox-contract/plans/001-bootstrap` 两处旧 `#313-18-个事件全集v1` 断链。
  - **影响**：即使历史 plan 内容保留，链接仍需要随 current spec 标题维护。

- 可执行 projection 和 active spec 可以同时漂移，但互相掩盖。
  - **证据**：B4 spec 仍把 `candidate_profiles` / `experience_cards` / `debriefs` 写成当前表项；同时 `backend/internal/migrations/privacy.go` 和 `TestPrivacyMatrixCoversEveryBaselineTableExactly` 都漏掉当前仍存在的 `idempotency_records`。
  - **影响**：只看 spec 会保留已删除表，只看测试又会漏掉当前用户关联表，privacy matrix 形成双向 false-green。

- Completed frontend-shell plans 仍会误导当前 UI / route contract。
  - **证据**：`frontend-shell/spec.md` 和 completed plans 001/002/003/004 仍把五入口、`jd_match` / `debrief` / `profile`、`auth_register` / `auth_reset` 或 `selectedJobMatchId` / `debriefId` 写成正向导航、screen、protected route、parity 或 URL survival 对象；当前 `frontend/README.md`、`ui-design/src/app.jsx`、`frontend/src/app/routes.ts`、`routeUrl.ts` 均已收敛为 `home` / `workspace` / `resume_versions` 三入口与 retired alias 归一。
  - **影响**：后续 `/plan-code-review` 或 `/implement` 若只读取 completed plan，会重新引入旧 route / TopBar / URL allowlist。

- Active subject 下的 completed historical plan 比 deprecated subject 更容易误导后续实现。
  - **证据**：`backend-practice` spec v1.13 已声明 `debrief/sourceDebriefId/source_debrief_id` 退役，但同 subject 的 completed `004-derived-plans-debrief` 仍把 debrief source、P0.071/P0.073 和 first-turn seeding 写成正向交付，且 context discovery 仍携带旧 `goal=debrief` / `PracticeGoalDebrief` 关键词。
  - **影响**：后续按 active subject 继续 `/implement backend-practice/004` 或做 L2 review 时，会从历史 plan 反向恢复已删除路径。

- Backend infra 与 AI contract completed plans 也会携带已删除模块的正向矩阵和包路径。
  - **证据**：`backend-async-runner/001` 仍把 `debrief_generate` / `jd_match_agent_scan` 写入当前 runner matrix、BDD rerun 和 context discovery；`prompt-rubric-registry/002` 仍把 Debrief / JD Match 写成 13-key schema/seed/caller 正向范围。
  - **影响**：后续按 completed plan context 做 L2 review 或恢复实施时，会误读已删除 backend package、prompt key 和 BDD 场景为当前 target。

- 历史 migration 可以合法保留已删除模块的建表记录，但必须有最终删除迁移和 focused guard 兜底。
  - **证据**：`000009_jd_match_baseline` / `000010_jd_match_seed_registry` 是历史建表 / seed 记录，当前 truth 由 `000014_drop_jd_match_module` 删除 5 张 JD Match 表和 `jd_match.*` registry rows；B4 active spec 曾仍把当前 B3 job type 误写为 9 项。
  - **影响**：如果只做全文负向搜索，会误删迁移历史；如果只允许历史命中而不加最终删除断言，又会让 retired module 重新出现在当前 schema / seed / check 口径中。

- 已删除 screen 的 global CSS 可以脱离组件继续存活。
  - **证据**：JD Match screen 已删除且无 class consumer，但 `frontend/src/app/theme/global.css` 仍保留 `jdmatch-recommended-grid` / `jdmatch-detail-panel` / `jdmatch-search-*` / `jdmatch-market-signals-inner` responsive class。
  - **影响**：只查组件、route、operationId 会漏掉全局样式死资产，后续重构可能误以为这些 class 仍属当前视觉系统。

- Migration enum source 的当前 lint 模型仍按历史 up migration 而非最终态 schema 判断。
  - **证据**：`migrations/enum-sources.yaml` 中 `agent_scans.status` / `watchlist_items.tone` 对应的表已由 `000014` 删除，但 `000009` historical up 仍含 check，当前 `migrations_lint.py` 会要求它们继续登记。
  - **影响**：直接删除 enum source 会让 lint 失败；继续保留则会在全文负向搜索里表现为 JD Match 残留。这个边界需要 migration squash 或 lint final-state modeling 单独处理。

## 3 根因归类

- Active truth source 缺少集中二次复查 gate。
  - **类别**：spec-plan

- Deprecated plan 投影约定没有在执行规则里明确为 “subject deprecated, historical plans completed”。
  - **类别**：README / tooling

- Anchor drift 只能由 docs-check 捕获，单纯 sync-doc-index 不覆盖。
  - **类别**：spec-plan

- Privacy matrix 测试复制了实现中的不完整 allowlist，没有从当前 baseline 表集合反推。
  - **类别**：spec-plan / test

- Completed plan 仍被后续工具当作 discovery/source context 使用，但缺少 current-scope reconciliation note。
  - **类别**：spec-plan / tooling

- Context manifest 是执行入口，不是普通历史附件；completed plan 的 `context.yaml` 即使 plan 正文已修，也可能继续把退休 route / operation 当成正向 discovery surface。
  - **类别**：spec-plan / tooling

- Active subject 的历史目录名可以保留，但当前正向 contract 必须在 plan/test/BDD/context 同步改写。
  - **类别**：spec-plan

- Backend infra / AI contract plans 也必须从当前 truth source 反推正向 package、job_type 和 feature_key 集合。
  - **类别**：spec-plan / tooling

- Historical migration cleanup needs two-sided guards: allow immutable historical create/seed files, but require final deletion migrations and tests that prove retired tables, seed rows and enum/check values are removed from the current end state.
  - **类别**：spec-plan / test

- Global CSS is a runtime asset and must be included in deleted-module negative sweeps, not only component and route directories.
  - **类别**：frontend / test

- Migration lint currently validates historical up-file check registration, not only end-state schema; removed-module cleanup must either keep historical enum sources or explicitly choose migration squash/final-state lint semantics.
  - **类别**：migration / tooling

- Bootstrap plans stay dangerous after completion because they often become copied setup templates; current positive counts in completed bootstrap plans must be reconciled whenever product-scope removes modules, even if later active specs are already correct.
  - **类别**：spec-plan / tooling

- Deprecated subject specs can be correct while their historical plan/context files still look executable.
  - **类别**：spec-plan / tooling

- Deprecated frontend/backend paired subjects need both sides checked; one side may already have banner text while its context still exposes old route/API surfaces.
  - **类别**：spec-plan / tooling

## 4 对流程资产的改进建议

- 在 `product-scope/001-core-loop-module-pruning` 后续清理 gate 中保留 Phase 6 这种 active-doc semantic sweep，而不是只跑已完成 checklist。
  - **落点**：spec-plan
  - **优先级**：high

- 明确 deprecated subject 的文档生命周期约定：subject spec/history 可为 `deprecated`，历史 plan/checklist/BDD/test 文档保留 `completed`，并用顶部 Deprecated note 防回流。
  - **落点**：docs README / sync-doc-index 说明
  - **优先级**：medium

- 对会改动 spec 标题的文档清理，固定把 `make docs-check` 放在收口 gate 中，避免历史 plan anchor 断裂。
  - **落点**：spec-plan
  - **优先级**：medium

- 对 migration privacy matrix，要求测试先枚举当前 public baseline 表集合，再断言 disposition；不要只维护历史固定 count。
  - **落点**：spec-plan / test
  - **优先级**：high

- 对 completed frontend-shell plans 增加 current-scope reconciliation gate：当 product-scope 删除 route/module 后，同步修 plan/checklist/BDD/context 中的 positive route list、URL allowlist 与 discovery uiRoutes，只把旧 key 留在 retired-negative 断言里。
  - **落点**：spec-plan / tooling
  - **优先级**：high

- 对 completed plan 的 `context.yaml` 增加 OpenAPI / route 正向集合反查：`apiNames` 必须与当前 operation matrix 精确一致，退休 operation 只能出现在 negative search 或 discovery keywords 中。
  - **落点**：spec-plan / tooling
  - **优先级**：high

- 对 active subject 下含退休模块名的 completed plan 增加 current-scope rewrite gate：目录名可保留作历史锚点，但 plan/test/BDD/context 的当前场景、命令和 operation matrix 必须只列保留功能。
  - **落点**：spec-plan
  - **优先级**：high

- 对 backend infra / AI contract completed plans 增加“current positive package/key set” gate：context discovery packages、feature_key count、job_type matrix、BDD rerun 命令必须从当前 truth source 反推，历史迁移或 revision 只能留在 retired-negative 语境。
  - **落点**：spec-plan / tooling
  - **优先级**：high

- 对历史 migration / seed 残留增加 final-state deletion guard：允许 pre-launch historical migration 保留旧建表记录，但必须有后一条 migration 删除表、seed rows、retired job/check 值，并用 focused Go/Python test 固化。
  - **落点**：spec-plan / test
  - **优先级**：high

- 对正式前端全局 CSS 增加 deleted-module scope guard：删除 screen 后必须同步搜索 `theme/*.css`、shared CSS 和 token 文件，旧模块 class name 不得继续作为正式资产保留。
  - **落点**：frontend test / spec-plan
  - **优先级**：medium

- 对 JD Match historical migration enum source 做单独决策：若用户批准 pre-launch migration squash，可删除 `000009` / `000010` 与对应 enum source；若不 squash，应先让 migration lint 能证明最终态 schema，再减少历史全文命中。
  - **落点**：migration / tooling
  - **优先级**：high

- 对 completed bootstrap plans 增加 current-count reconcile gate：事件数、job type 数、API-facing subset、应用表数、public schema gate 必须从当前 truth source 反推，旧计数只能保留在明确历史 revision row 中。
  - **落点**：spec-plan / tooling
  - **优先级**：high

- 对 deprecated subject 的 historical plans 增加退役 banner + context surface gate：plan/checklist/BDD 顶部必须说明不可实施，`context.yaml` 不得继续暴露旧 operationId、旧 package、旧 fixture 或旧 scenario 作为正向 discovery。
  - **落点**：spec-plan / tooling
  - **优先级**：high

- 对成对退役模块执行 backend/frontend 双向 context gate：后端 operation/package/prompt 与前端 route/screen/fixture 都必须同时收敛，否则后续 owner 仍可能从另一侧恢复旧模块。
  - **落点**：spec-plan / tooling
  - **优先级**：high

## 5 建议优先级与后续动作

最高价值后续动作：继续以 `product-scope/001-core-loop-module-pruning` 为 owner，先做一次更机械的 runtime/generated allowlist 脚本化（把负向测试、历史迁移、legacy normalization 与真实残留分桶输出），然后再请用户决定是否批准 pre-launch migration squash。当前不建议在未确认前删除 `000009` / `000010`，因为它们仍是历史迁移链的一部分，已由 `000014` 和 focused test 证明最终状态清理；但如果目标是“历史文件也零残留”，下一步必须进入 squash 决策。

可延后处理：扩展 `sync-doc-index` 对 deprecated plan section 的原生支持；当前已有的 subject deprecated + plan completed 模式已经能通过门禁。
