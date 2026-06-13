# JD-Match Real Backend Baseline

> **版本**: 2.0
> **状态**: completed
> **更新日期**: 2026-06-13

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 0.0 D-17 Module Removal Reopen (v2.0)

2026-06-12 [product-scope v2.1 D-17](../../../product-scope/spec.md#31-已锁定决策) 整体删除岗位推荐模块。本 plan 原地重开（completed -> active），新增 Phase 9 承接 [spec v2.0 §9](../../spec.md#9-d-17-删除范围与零残留验收当前-active-scope) 的 backend / 契约侧删除：OpenAPI jobmatch tag 12 个 operation 与 fixtures、`backend/internal/jdmatch/` 全包、cmd/api wiring、5 张表 drop migration、B3 事件 / job_type、F3 feature_key 与 prompt/rubric/eval 资产、A3 profile 条目、P0.094-097 场景目录。Phase 0-8 为历史完成记录，其断言的能力随 D-17 退役，不得作为当前验收依据。

**Phase 9 质量门禁**：删除型 phase 不新增用户行为流，BDD 不适用；替代验证 gate 为 spec §9.2 C-R1~C-R3——`make codegen && make codegen-check`、fixtures / mock-contract lint、`cd backend && go test ./...`、`/api/v1/jd-match/*` 404 运行时断言、drop migration up/down 测试、跨仓零残留负向搜索（drop migration、历史迁移文件 000009/000010 与负向断言除外）。删除属 TDD 范畴：先以失败断言（如 route 404 测试、负向 grep gate）表达目标态，再执行删除使其转绿。

### Phase 9 operation matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| jobmatch tag 12 个 operation（全部删除） | `openapi/fixtures/JobMatch/`（全部删除） | frontend jd_match 模块（由 frontend-home-job-picks-and-parse/002 同步删除） | `backend/internal/jdmatch/` + `cmd/api` jdmatch wiring（全部删除） | 5 张 jd_match 表（drop migration 收口） | `jd_match.recommendation` / `jd_match.search` feature_key（删除） | C-R1~C-R3 替代 gate；P0.094-097 场景目录删除 |

### Phase 9: D-17 module removal

#### 9.1 契约删除与再生成

删除 `openapi/openapi.yaml` jobmatch tag、12 个 operation、专属 schema 与 `openapi/fixtures/JobMatch/`；运行 `make codegen && make codegen-check` 再生成 frontend/backend client、server、types；同步修订 `openapi-v1-contract` spec freeze 列表与 `mock-contract-suite` 中 JobMatch 口径；fixtures / mock-contract lint 通过。

#### 9.2 backend 包与运行时删除

删除 `backend/internal/jdmatch/` 全包、`cmd/api/jdmatch_runtime.go`、`jdmatch_live_scenario_test.go` / `jdmatch_fixture_parity_test.go` / `jdmatch_http_scenario_test.go`、`main.go` 挂载点与 session policy jobmatch 条目；原地修改 privacy runner（去除 `DeleteJobMatchDataForUser` 链路）与 `backend/internal/auth/identity.go` 等共享触点；保留仍被其他消费方使用的 cross-owner counter / identity API 并记录留存理由（如 privacy 数据概览）。先写 route 404 / 负向断言（Red），删除后转绿。

#### 9.3 数据与注册表收口

新增 drop migration：删除 `jd_match_recommendations` / `watchlist_items` / `saved_searches` / `agent_scans` / `jd_match_search_runs` 5 张表与 000010 注入的 `jd_match.*` prompt/rubric registry 行；`migrations/enum-sources.yaml` 同步；migration up/down 测试与迁移 gate 通过。

#### 9.4 shared / config 资产删除

删除 `shared/events.yaml` / `shared/jobs.yaml` 及 baseline / schema 中 `jd_match.*` 事件与 job_type，重新生成共享常量；删除 `config/prompts|rubrics|evals/jd_match.*`、`config/ai-profiles.yaml` 对应 entry，并再生成 `config/evals/resolved-prompts.json`；相关 lint（events / jobs / rubric / config）通过。

#### 9.5 场景与文档收口

删除 `test/scenarios/e2e/p0-094..097-jd-match-*` 4 个目录并更新 `test/scenarios/e2e/INDEX.md`；同步 `engineering-roadmap` §5.2 对本 subject 的描述为 D-17 删除完成；`sync-doc-index --check` 零漂移。

#### 9.6 零残留与全量回归 gate

`rg -i "jobmatch|jd[-_]match"` 于 `openapi/ backend/ shared/ config/ frontend/src/api/generated/`（drop migration、历史迁移文件、负向断言与本 plan 文档除外）零残留；`cd backend && go test ./...`、`make codegen-check`、`make lint-mock-contract`（如适用）、`make docs-check` 通过；`/api/v1/jd-match/*` 返回 404 的运行时断言通过。

## 0 Post-Reopen Completion Note

2026-05-22 `/plan-code-review --fix` 曾将 Phase 5.5-5.8 与 Phase 6.1-6.8 / 6.12-6.13 退回 active：当时 `TestJDMatchHTTPScenario` 仅是 live smoke，`bdd-checklist.md` 与 `test/scenarios/e2e/INDEX.md` 仍无完成证据。后续补齐 `buildJDMatchRuntime` lifecycle gate、12-route session/IK/cross-user/live scenario、`TestJDMatchAgentScanDrainerScenario`、`TestJDMatchFixtureParity` 与 E2E.P0.094-097 wrapper `setup -> trigger -> verify` PASS 证据；本 plan lifecycle 恢复 completed。

## 0.1 L2 Follow-up Completion Note

2026-05-22 第二轮 `/plan-code-review backend-jobs-recommendations/001-jd-match-real-backend-baseline --fix` 修复 review 遗留问题：`privacy_delete` runner 现在会调用 backend-profile `DeleteCandidateProfileForUser` 与 JD-Match `DeleteJobMatchDataForUser`；`jd_match_agent_scan` 在调用 `jd_match.recommendation` generator 前组装 runtime `JobMatchProfile` 与内部 jobs pool JSON；JDMatch handler 错误响应统一为 `ApiErrorResponse` envelope 并使用 shared error registry 的 retryable 值；本地 `.claude/scheduled_tasks.lock` 从版本控制中移除并加入 ignore。验证覆盖 focused 单测、`cmd/api` live JDMatch matrix、`cd backend && go test ./...` 与 E2E.P0.097 wrapper。

## 0.2 Repo Lint Follow-up Evidence Correction

2026-05-23 review correction：对照 merge base 后确认 JD-Match 六个 rubric dimension 已在 `config/rubrics/README.md` 与 `scripts/lint/rubric_lint.py` 中存在；上一轮 repo lint follow-up 对该项的记录属于证据漂移，并引入了重复 allowlist stanza。Phase 8 现在只负责移除重复项、保留 unknown dimension negative fixture 与 `rubric_lint` 通过证据；practice voice / mock runtime / runtime topology / Go revive 仍是 BUG-0092 的实际 repo lint 修复范围。

## 1 目标

把 [backend-jobs-recommendations spec](../../spec.md) §6 C-1..C-19 全部落到 backend Go handler + store + async jobs + AI 编排 + cmd/api wiring + 多个 cross-owner additive：

- 实现 12 个 JobMatch endpoint handler：`getJobMatchProfile` / `getAgentScanStatus` / `listJobRecommendations` / `getJobRecommendation` / `markJobNotRelevant` / `listWatchlist` / `addToWatchlist` / `removeFromWatchlist` / `searchJobs` / `listSavedSearches` / `createSavedSearch` / `getMarketSignals`；
- 携带 [B4 cross-owner additive migration](../../../db-migrations-baseline/spec.md) 新建 5 张表（`jd_match_recommendations` / `watchlist_items` / `saved_searches` / `agent_scans` / `jd_match_search_runs`）+ 对应 index / FK / check constraint + enum source；表总数 28 → 33；
- 携带 [F3 cross-owner additive](../../../prompt-rubric-registry/spec.md) 新增 2 个 baseline feature_key `jd_match.recommendation`（model profile `jd_match.recommendation.default`）+ `jd_match.search`（model profile `jd_match.search.default`）+ baseline prompt / rubric YAML / Markdown 文件；F3 字典 11 → 13；
- 携带 [A3 cross-owner additive](../../../ai-provider-and-model-routing/spec.md) `config/ai-profiles.yaml` 新增 2 个 model profile entry（capability=chat，`provider_ref=deepseek`，model 使用 `deepseek-v4-flash` / `deepseek-v4-pro`）并同步 A3 §4.5 Product/UI AI Capability Catalog；
- 携带 [B3 cross-owner additive](../../../event-and-outbox-contract/spec.md) 新增 2 个 internal event `jd_match.recommendation.completed` / `jd_match.search.completed` + 2 个 canonical job_type `jd_match_agent_scan`（dotted `jd_match.agent_scan`）+ `jd_match_search`（dotted `jd_match.search`）；envelope 遵守 PII 边界；
- 携带 5 个 cross-owner internal API additive：`backend-auth.GetUserIdentityForUser`（spec D-17 锁定，本 plan 携带 backend-auth in-place additive）+ `backend-resume.CountResumesForUser` / `backend-targetjob.CountTargetJobsForUser` / `backend-practice.CountPracticeSessionsForUser` / `backend-debrief.CountDebriefsForUser`（这 4 个由本 plan 携带各 owner spec / handler / unit test additive）；`backend-profile.GetCandidateProfileForUser`（D-13 / read-only / 不触发 seed 副作用）与 `CountExperienceCardsBySource`（D-11）已由 [backend-profile/001](../../../backend-profile/plans/001-candidate-profile-and-experience-cards/plan.md) 提供；总计 7 个 cross-owner API 供 `BuildJobMatchProfile` 聚合；
- 实现 store layer：5 张表 Repository + cursor pagination + cross-user 隔离 + privacy delete；
- 实现 AI 编排：`jd_match.recommendation` generator/service（由 `jd_match_agent_scan` 内联调用，不注册独立 job_type）+ `jd_match.search` 同步调用 + `jd_match.agent_scan` 后台周期 job；通过 [A3 AIClient](../../../ai-provider-and-model-routing/spec.md) + [F3 registered feature_key](../../../prompt-rubric-registry/spec.md)；
- 在 `cmd/api` 挂载 12 个 route + session middleware + IK middleware（5 个 side-effect op）+ in-process drainer（P0 只注册 `jd_match_agent_scan` 1 个后台 handler；`jd_match_search` 仅作为 future async 预占，recommendation generator 不注册独立 job_type）；
- 实现 `BuildJobMatchProfile(ctx context.Context, userID string)` internal service：通过 7 个 cross-owner internal API 聚合画像数据；P0 baseline 按 spec D-18 字段映射返回稀疏 JobMatchProfile（`displayName` 必填非 null；`avatarUrl/locationText/compensationText=null`；`headline/yearsOfExperience` 来自 backend-profile；`skills=[]`；`sources` 真实计数）；
- 实现 `DeleteJobMatchDataForUser(userId)` internal API + audit tombstone；
- 通过 spec §6 C-1..C-19 验收 + 新增 E2E.P0.094 + E2E.P0.095 + E2E.P0.096 + E2E.P0.097 4 个 BDD 场景；
- 不接入外部招聘平台（LinkedIn / Boss / 脉脉 / 拉勾 / 公司官网 等）；不实现 vector search / pgvector；不实现深度 company intel（归 plan 002 / 003 / future P2）。

## 2 背景

[engineering-roadmap §5.2](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) "Home / Job Picks / Parse" workstream 在 backend-jobs-recommendations spec 创建期（roadmap 3.17）已把旧备注 "JobMatch real backend subject not yet created" 替换为 `backend-jobs-recommendations active（001 real backend baseline ...）`；[frontend-home-job-picks-and-parse §2.2 + §3.1 D-1](../../../frontend-home-job-picks-and-parse/spec.md) 与 [openapi-v1-contract §3.1.1](../../../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) 都已显式预占 `backend-jobs-recommendations` 承接 JobMatch 12 个 endpoint 的真实 backend；frontend `plan 002-jd-match-recommendations` 当前通过 fixture-backed transport 闭环，等待真实 backend 落地后切真。

本 plan 是 backend-jobs-recommendations 第一批 plan，承担 P0 "用户进入 jd_match 路由 → 看到画像快照 + agent scan 状态 → 浏览推荐 / 详情 / 加入收藏 / 标不相关 → 自然语言搜索 + 保存搜索 → 看 watchlist + market signals" 完整 backend 端到端。

JD-Match 业务域当前 backend zero state：
- DB：0 张表（B4 baseline 不含 jd_match_* 表）
- F3：0 个 feature_key（11 个 baseline 不含 `jd_match.*`）
- A3：0 个 model profile entry（`config/ai-profiles.yaml` 不含 `jd_match.*`）
- B3：0 个 event / job_type（events.yaml / jobs.yaml 不含 `jd_match.*`）
- Code：0 个 handler / store / service（`backend/internal/` 不含 `jdmatch/` 目录）
- generated server interface：12 个 method 已存在但全部 unimplemented

每个 phase 是可独立验证的纵向切片：
- Phase 0：cross-owner additive 准备（B4 / F3 / A3 / B3 / 5 个 counter API spec lock）
- Phase 1：getJobMatchProfile + getAgentScanStatus + 5 个 cross-owner counter integration
- Phase 2：listJobRecommendations + getJobRecommendation + markJobNotRelevant + recommendation generator + agent_scan AI integration
- Phase 3：listWatchlist + addToWatchlist + removeFromWatchlist + watchlist UNIQUE + tone 派生
- Phase 4：searchJobs + listSavedSearches + createSavedSearch + jd_match_search_runs 写入
- Phase 5：getMarketSignals + agent_scan 后台 job + privacy delete + cmd/api runtime wiring
- Phase 6：收口 + BDD + cross-owner handoff（frontend 切真 + B4/F3/A3/B3 owner sign-off）

执行本 plan 前必须确认：

- [B2 §3.1.1](../../../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) JobMatch tag 12 个 operationId 已 freeze；fixture `openapi/fixtures/JobMatch/*.json` 12 个文件已存在并通过 `make validate-fixtures`
- [B4 baseline migration](../../../db-migrations-baseline/spec.md) 已就位（baseline 000001-000008）；本 plan 携带 000009 cross-owner additive migration
- [F3 baseline](../../../prompt-rubric-registry/spec.md) 11 个 feature_key 已 ready；本 plan 携带 cross-owner additive
- [A3 003](../../../ai-provider-and-model-routing/plans/003-provider-registry-and-capability-profiles/plan.md) 已 ready（AIClient + provider registry + Capability Model Profile）；本 plan 携带 `config/ai-profiles.yaml` cross-owner additive
- [B3 baseline](../../../event-and-outbox-contract/spec.md) 已 ready；本 plan 携带 events.yaml / jobs.yaml cross-owner additive
- [backend-profile/001](../../../backend-profile/plans/001-candidate-profile-and-experience-cards/plan.md) completed：`GetCandidateProfileForUser`（D-13 read-only 不触发 seed 副作用） / `CountExperienceCardsBySource`（D-11）internal API 已可用
- [backend-resume/001](../../../backend-resume/plans/001-asset-register-parse-and-listing/plan.md) completed：resume_assets store 可用；本 plan 携带 `CountResumesForUser` cross-owner additive
- [backend-targetjob](../../../backend-targetjob/spec.md) 001 completed：target_jobs store 可用；本 plan 携带 `CountTargetJobsForUser` cross-owner additive
- [backend-practice](../../../backend-practice/spec.md) 各 plan completed：practice_sessions store 可用；本 plan 携带 `CountPracticeSessionsForUser` cross-owner additive
- [backend-debrief](../../../backend-debrief/spec.md) 001 completed：debriefs store 可用；本 plan 携带 `CountDebriefsForUser` cross-owner additive
- [backend-auth/001](../../../backend-auth/plans/001-passwordless-session-bootstrap/plan.md) completed：session middleware 可用
- [backend-runtime-topology](../../../backend-runtime-topology/spec.md) `cmd/api` in-process composition 模式已就绪

## 3 质量门禁分类

- **Plan 类型**: `code-internal + feature-behavior + contract + migration + tooling`。本 plan 实现 backend handler / store / async job / AI 调用；用户可见 HTTP API 行为；含多个 cross-owner additive（B4 migration / F3 prompt / A3 profile / B3 event-job / 5 counter API）。
- **TDD 策略**: 适用。Red-Green-Refactor 入口：
  1. handler unit test：12 个 endpoint 各自参数校验 + IK（5 个 side-effect op）+ cross-user + 错误路径；
  2. store integration test：5 张新表 CRUD + cursor pagination + cross-user 隔离 + UNIQUE 约束 + privacy delete cascade；
  3. cross-owner additive test：B4 migration / F3 prompt loader / A3 profile registry / B3 events generator / 5 counter API 各自 owner gate PASS；
  4. AI 编排 unit test（stub AIClient）：`jd_match.recommendation` generator 成功 / output_invalid / timeout retryable；`jd_match.search` 同步成功 / 超时 502；`jd_match.agent_scan` 周期触发 / 增量触发；
  5. outbox event unit test：2 个 event envelope 字段集 + PII 边界（不含 query / reasons / source_url）；
  6. privacy delete internal API test：5 表删除顺序 + audit tombstone 完整 + 无敏感字段泄漏；
  7. BuildJobMatchProfile aggregation test：7 个 cross-owner internal API 调用（D-17 / D-18 字段映射 + cross-owner failure fallback）+ sources 计数正确 + skills baseline=[] + locationText/compensationText/avatarUrl baseline=null + displayName 必填非 null；
  8. `cmd/api` route/runtime test：session middleware、IK middleware（5 个 op）、12 个 route 真实可达、`jd_match_agent_scan` 1 个后台 handler 注册 / shutdown。
  执行入口：`/implement backend-jobs-recommendations/001-jd-match-real-backend-baseline` → `/tdd`。
- **BDD 策略**: 适用（Feature plan requires BDD）。E2E.P0.094 jd-match-profile-and-recommendations-list + E2E.P0.095 jd-match-watchlist-and-saved-search-lifecycle + E2E.P0.096 jd-match-search-and-market-signals + E2E.P0.097 jd-match-agent-scan-and-privacy-delete。详见 [bdd-plan.md](./bdd-plan.md) / [bdd-checklist.md](./bdd-checklist.md)。
- **替代验证 gate**:
  - `cd backend && go test ./...`
  - `cd backend && go test ./internal/jdmatch/handler/... -count=1`
  - `cd backend && go test ./internal/jdmatch/store/... -tags=integration -count=1`
  - `cd backend && go test ./internal/jdmatch/jobs/... -count=1`（stub AIClient）
  - `cd backend && go test ./internal/jdmatch/service/... -count=1`（BuildJobMatchProfile + privacy delete）
  - `cd backend && DATABASE_URL='postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' go test ./cmd/api -run '^(TestBuildJDMatchRuntimeWiresRoutesDrainerAndLifecycle|TestJDMatchRoutesRequireSessionOnAllRoutes|TestJDMatchHTTPScenario|TestJDMatchAgentScanDrainerScenario|TestJDMatchFixtureParity)$' -count=1 -v`
  - migration gate：`migrations/lint.sh` + `make migrate-check` + `make privacy-delete-dry-run` PASS（本地若缺 `DATABASE_URL`，必须至少记录 `python3 scripts/lint/migrations_lint.py --repo-root .` PASS 与 live DB 子步骤 blocker，不得冒充 migrate-check 全绿）
  - F3 gate：`make lint-ai-profile-coverage` + F3 prompt loader test PASS
  - B3 gate：events / jobs generator PASS + baseline manifest update
  - smoke：`curl -X GET /api/v1/jd-match/profile` 通过 D-19 structural parity，另选 1 个非 profile endpoint 与 mock-server fixture 字节比对
  - grep `LinkedIn|Boss|脉脉|拉勾|mistake|growth|drill|experiences|star` in `backend/internal/jdmatch/` + outbox payload + DB seed：0 命中（C-16 negative；外部平台名作为 future 描述允许在 spec 文档注释中，禁止在实际 code / DB seed 中）
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

### 3.1 Frontend / Backend Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getJobMatchProfile` | `openapi/fixtures/JobMatch/getJobMatchProfile.json` `default` (design preview) / `partial-profile` (零计数 reference) — **byte parity 例外**: P0 baseline 走 spec D-19 structural parity (required 字段 + optional null oneOf + sources 真实计数 + status + X-Request-ID) | `frontend-home-job-picks-and-parse/002-jd-match-recommendations` (currently fixture-backed; switches real after this plan) | `backend/internal/jdmatch/handler/get_profile.go` + `cmd/api` GET `/api/v1/jd-match/profile` route with session middleware | 7 cross-owner internal API aggregation (`backend-auth.GetUserIdentityForUser` + `backend-profile.GetCandidateProfileForUser` + `backend-profile.CountExperienceCardsBySource` + 4 个 counter — D-17 / D-18 / D-19 锁定); no own table | none | E2E.P0.094 |
| `getAgentScanStatus` | `openapi/fixtures/JobMatch/getAgentScanStatus.json` `default` / `scanning` / `error` | plan 002 (real switch) | `backend/internal/jdmatch/handler/get_agent_status.go` + `cmd/api` GET `/api/v1/jd-match/agent-status` | `agent_scans` latest row read | none | E2E.P0.094 + E2E.P0.097 |
| `listJobRecommendations` | `openapi/fixtures/JobMatch/listJobRecommendations.json` `default` / `empty` / `one` / `many` / `failed` | plan 002 (real switch) | `backend/internal/jdmatch/handler/list_recommendations.go` + `cmd/api` GET `/api/v1/jd-match/recommendations` | `jd_match_recommendations` cursor pagination (dismissed filtered) | none (read-only; AI generation in agent_scan job) | E2E.P0.094 |
| `getJobRecommendation` | `openapi/fixtures/JobMatch/getJobRecommendation.json` `default` | plan 002 (real switch) | `backend/internal/jdmatch/handler/get_recommendation.go` + `cmd/api` GET `/api/v1/jd-match/recommendations/{jobMatchId}` | `jd_match_recommendations` row read with cross-user | none | E2E.P0.094 |
| `markJobNotRelevant` | `openapi/fixtures/JobMatch/markJobNotRelevant.json` `default` | plan 002 (real switch) | `backend/internal/jdmatch/handler/mark_not_relevant.go` + `cmd/api` POST `/api/v1/jd-match/recommendations/{jobMatchId}/dismiss` with session + IK | `jd_match_recommendations.dismissed_at` UPDATE | none | E2E.P0.094 |
| `listWatchlist` | `openapi/fixtures/JobMatch/listWatchlist.json` `default` | plan 002 (real switch) | `backend/internal/jdmatch/handler/list_watchlist.go` + `cmd/api` GET `/api/v1/jd-match/watchlist` | `watchlist_items` cursor + JOIN to `jd_match_recommendations` for tone derivation | none | E2E.P0.095 |
| `addToWatchlist` | `openapi/fixtures/JobMatch/addToWatchlist.json` `default` | plan 002 (real switch) | `backend/internal/jdmatch/handler/add_watchlist.go` + `cmd/api` POST `/api/v1/jd-match/watchlist` with session + IK | `watchlist_items` INSERT + UNIQUE constraint | none | E2E.P0.095 |
| `removeFromWatchlist` | `openapi/fixtures/JobMatch/removeFromWatchlist.json` `default` | plan 002 (real switch) | `backend/internal/jdmatch/handler/remove_watchlist.go` + `cmd/api` DELETE `/api/v1/jd-match/watchlist/{jobMatchId}` with session + IK | `watchlist_items` DELETE | none | E2E.P0.095 |
| `searchJobs` | `openapi/fixtures/JobMatch/searchJobs.json` `default` / `failed` | plan 002 (real switch) | `backend/internal/jdmatch/handler/search.go` + `cmd/api` POST `/api/v1/jd-match/search` with session + IK | `jd_match_search_runs` INSERT + `jd_match_recommendations` JOIN (no ephemeral insert) | `jd_match.search` via A3/F3 (sync 30s timeout) | E2E.P0.096 |
| `listSavedSearches` | `openapi/fixtures/JobMatch/listSavedSearches.json` `default` | plan 002 (real switch) | `backend/internal/jdmatch/handler/list_saved.go` + `cmd/api` GET `/api/v1/jd-match/saved-searches` | `saved_searches` cursor | none | E2E.P0.095 |
| `createSavedSearch` | `openapi/fixtures/JobMatch/createSavedSearch.json` `default` | plan 002 (real switch) | `backend/internal/jdmatch/handler/create_saved.go` + `cmd/api` POST `/api/v1/jd-match/saved-searches` with session + IK | `saved_searches` INSERT | none | E2E.P0.095 |
| `getMarketSignals` | `openapi/fixtures/JobMatch/getMarketSignals.json` `default` | plan 002 (real switch) | `backend/internal/jdmatch/handler/get_market_signals.go` + `cmd/api` GET `/api/v1/jd-match/market-signals` | aggregation from `watchlist_items` + `jd_match_recommendations` + internal pool stats | none | E2E.P0.096 |

### 3.2 Cross-owner Additive Summary

Phase 0 必须同时携带多个 cross-owner spec / artifact additive 修订：

| Owner | Additive 内容 | Gate |
|-------|--------------|------|
| [db-migrations-baseline (B4)](../../../db-migrations-baseline/spec.md) | `migrations/000009_jd_match_baseline.{up,down}.sql` 创建 5 张表 + index + FK + CHECK constraints；`migrations/enum-sources.yaml` 追加新 enum 来源；B4 spec.md §X 表数 28 → 33 + D-X 决策行；B4 history.md 追加 backend-jobs-recommendations/001 cross-owner additive 行 | `migrations/lint.sh` PASS + `make migrate-check` PASS（无 `DATABASE_URL` 时记录 live DB blocker）+ `make privacy-delete-dry-run` PASS + B4 sync-doc-index PASS |
| [prompt-rubric-registry (F3)](../../../prompt-rubric-registry/spec.md) | `config/prompts/jd_match.recommendation/v0.1.0.{yaml,md}` + `config/prompts/jd_match.search/v0.1.0.{yaml,md}`；`config/rubrics/jd_match.recommendation/v0.1.0.yaml` + `config/rubrics/jd_match.search/v0.1.0.yaml`；F3 spec.md §3.1.1 字典 11 → 13 + D-X 决策行；F3 history.md 追加 cross-owner additive 行 | F3 prompt loader test PASS + `make lint-ai-profile-coverage` PASS |
| [ai-provider-and-model-routing (A3)](../../../ai-provider-and-model-routing/spec.md) | `config/ai-profiles.yaml` 新增 `jd_match.recommendation.default` + `jd_match.search.default` 两个 capability=chat profile entry（`provider_ref=deepseek`，model 只能使用 `deepseek-v4-flash` / `deepseek-v4-pro`）；同步 A3 §4.5 Product/UI AI Capability Catalog，把 JD-Match 推荐解释 / 搜索指向新 profile，不再复用 `target.import.default` | `make lint-ai-profile-coverage` PASS + provider registry runtime bootstrap test PASS |
| [event-and-outbox-contract (B3)](../../../event-and-outbox-contract/spec.md) | `shared/events.yaml` 新增 `jd_match.recommendation.completed` + `jd_match.search.completed` 2 个 internal event；`shared/jobs.yaml` 新增 `jd_match_agent_scan` (dotted `jd_match.agent_scan`) + `jd_match_search` (dotted `jd_match.search`) 2 个 canonical job_type；B3 spec.md §3.1 events / jobs total bump + D-X 决策行；B3 history.md 追加 cross-owner additive 行 | B3 generator PASS + baseline manifest update PASS |
| [backend-resume](../../../backend-resume/spec.md) | `backend/internal/resume/service/count.go` 新增 `CountResumesForUser(ctx context.Context, userID string) (int, error)` internal service；backend-resume spec.md cross-owner exposed internal API 行追加；backend-resume history.md 追加 backend-jobs-recommendations/001 cross-owner additive 行 | backend-resume unit test PASS |
| [backend-targetjob](../../../backend-targetjob/spec.md) | `backend/internal/targetjob/service/count.go` 新增 `CountTargetJobsForUser(ctx context.Context, userID string) (int, error)`；同上 spec / history 修订 | backend-targetjob unit test PASS |
| [backend-practice](../../../backend-practice/spec.md) | `backend/internal/practice/service/count.go` 新增 `CountPracticeSessionsForUser(ctx context.Context, userID string) (int, error)`；同上 spec / history 修订 | backend-practice unit test PASS |
| [backend-debrief](../../../backend-debrief/spec.md) | `backend/internal/debrief/service/count.go` 新增 `CountDebriefsForUser(ctx context.Context, userID string) (int, error)`；同上 spec / history 修订 | backend-debrief unit test PASS |
| [backend-auth](../../../backend-auth/spec.md) | `backend/internal/auth/service/identity.go` 新增 `GetUserIdentityForUser(ctx context.Context, userID string) (UserIdentity, error)` internal service，`UserIdentity` 含 `{displayName, avatarUrl, emailMasked}` (D-17 read-only / 不写 audit / 不返回 raw email；emailMasked 形如 `ali***@example.com`)；backend-auth spec.md 模块边界表追加 internal API 行；history.md 追加 backend-jobs-recommendations/001 cross-owner additive 行 | backend-auth unit test PASS (cross-user 由 caller userId 保证；emailMasked 字段 redact; 不写 audit_events 断言) |

## 4 实施步骤

### Phase 0: cross-owner additive 准备

#### 0.1 B4 cross-owner additive migration
- 撰写 `migrations/000009_jd_match_baseline.up.sql` + `down.sql`：5 张表 schema + index + FK + CHECK constraints
- 表设计：
  - `jd_match_recommendations`: `id uuid PK / user_id uuid FK / job_match_id uuid UNIQUE per user / title / company / company_tag / level / location / comp / posted_label / score smallint 0-100 / fit jsonb / reasons text[] / risks text[] / highlights text[] / seen bool / dismissed_at / source_url / source_label / network_note / similar_interviewers int / interview_hypotheses text[] / prompt_version / rubric_version / model_id / language / feature_flag / data_source_version / recommended_at / updated_at / deleted_at`
  - `watchlist_items`: `id uuid PK / user_id uuid FK / linked_job_match_id uuid FK → jd_match_recommendations / label / tone text CHECK IN ('ok','warn','muted') / added_at / change text / UNIQUE (user_id, linked_job_match_id)`
  - `saved_searches`: `id uuid PK / user_id uuid FK / label / query / filters jsonb / new_jobs_count / last_run_at / created_at / updated_at`
  - `agent_scans`: `id uuid PK / user_id uuid FK / status text CHECK IN ('idle','scanning','error') / started_at / finished_at / last_scan_at / next_scan_at / error_message`
  - `jd_match_search_runs`: `id uuid PK / user_id uuid FK / search_run_id uuid UNIQUE / query / filters jsonb / result_count / prompt_version / rubric_version / model_id / data_source_version / created_at`
- 更新 `migrations/enum-sources.yaml`：追加 `agent_scans.status` + `watchlist_items.tone` enum source 来源
- 修订 [B4 spec.md](../../../db-migrations-baseline/spec.md) §3.X 表数 28 → 33 + 新 D-X 决策行 "JD-Match baseline tables cross-owner additive"
- 修订 [B4 history.md](../../../db-migrations-baseline/history.md) 追加 backend-jobs-recommendations/001 cross-owner additive 行
- 运行 `migrations/lint.sh` PASS + `make migrate-check` PASS（或记录 live DB blocker）+ `make privacy-delete-dry-run` PASS + sync-doc-index --check PASS

#### 0.2 F3 cross-owner additive
- 撰写 `config/prompts/jd_match.recommendation/v0.1.0.yaml` + `v0.1.0.md`（baseline prompt 文本）
- 撰写 `config/prompts/jd_match.search/v0.1.0.yaml` + `v0.1.0.md`
- 撰写 `config/rubrics/jd_match.recommendation/v0.1.0.yaml`（dimensions: relevance_to_jd / risk_clarity / actionability）
- 撰写 `config/rubrics/jd_match.search/v0.1.0.yaml`（dimensions: query_alignment / diversity / privacy_compliance）
- 计算 + 写入 `template_hash`（sha256 canonical）
- 修订 [F3 spec.md](../../../prompt-rubric-registry/spec.md) §3.1.1 字典 11 → 13 + 新 D-X 决策行 "Add jd_match.recommendation + jd_match.search feature_keys"
- 修订 [F3 history.md](../../../prompt-rubric-registry/history.md) 追加 cross-owner additive 行
- 运行 F3 prompt loader test PASS + `make lint-ai-profile-coverage` PASS

#### 0.3 A3 cross-owner additive
- 修订 `config/ai-profiles.yaml` 新增 2 个 profile entry：
  ```yaml
  - name: jd_match.recommendation.default
    capability: chat
    status: active
    default:
      provider_ref: deepseek
      model: deepseek-v4-flash
  - name: jd_match.search.default
    capability: chat
    status: active
    default:
      provider_ref: deepseek
      model: deepseek-v4-flash
  ```
- 修订 [A3 spec.md](../../../ai-provider-and-model-routing/spec.md) §4.5 Product/UI AI Capability Catalog：新增或改写 JD-Match 推荐解释 / 搜索行，默认 profile 分别为 `jd_match.recommendation.default` / `jd_match.search.default`，不再复用 JD 导入解析的 `target.import.default`
- 运行 `make lint-ai-profile-coverage` PASS + A3 provider registry runtime bootstrap test PASS

#### 0.4 B3 cross-owner additive
- 修订 `shared/events.yaml` 新增 2 个 internal event：
  - `jd_match.recommendation.completed`: payload `{userId, agentScanId, recommendationCount, completedAt}`
  - `jd_match.search.completed`: payload `{userId, searchRunId, resultCount, completedAt}`
  - PII 边界断言：不含 query / reasons / source_url
- 修订 `shared/jobs.yaml` 新增 2 个 canonical job_type：
  - `jd_match_agent_scan` (dotted `jd_match.agent_scan`)
  - `jd_match_search` (dotted `jd_match.search`)
- 修订 [B3 spec.md](../../../event-and-outbox-contract/spec.md) §3.1 events 总数 + jobs 总数 bump + 新 D-X 决策行
- 修订 [B3 history.md](../../../event-and-outbox-contract/history.md) 追加 cross-owner additive 行
- 运行 B3 generator PASS + baseline manifest update PASS

#### 0.5 4 个 cross-owner counter internal API additive
- 在 [backend-resume](../../../backend-resume/spec.md) / [backend-targetjob](../../../backend-targetjob/spec.md) / [backend-practice](../../../backend-practice/spec.md) / [backend-debrief](../../../backend-debrief/spec.md) 各自 `internal/<domain>/service/count.go` 新增 `Count*ForUser(ctx context.Context, userID string) (int, error)` internal service
- 每个 counter 必须 cross-user 隔离（仅返回 `user_id = current_user_id` 行计数）
- 修订各 owner spec.md 在模块边界表追加新 internal API 行；history.md 追加 cross-owner additive 行
- 各 owner 单元测试 PASS

#### 0.6 backend-auth identity cross-owner internal API additive（spec D-17）
- 在 [backend-auth](../../../backend-auth/spec.md) `internal/auth/service/identity.go` 新增 `GetUserIdentityForUser(ctx context.Context, userID string) (UserIdentity, error)` internal service
- `UserIdentity` 字段：`{displayName string, avatarUrl *string, emailMasked string}`
- read-only / 不写 `audit_events` / 不更新 `users` 表 / 不触发 session 副作用
- 返回 `emailMasked`（形如 `ali***@example.com`）；**禁止返回 raw email**；不返回 `password_hash` / `email_verified_at` / `created_at` 等非 identity 字段
- cross-user 隔离由 caller 提供 `userId` 保证；若 userId 不存在返回 `(UserIdentity{}, ErrUserNotFound)`，caller fallback 到 anonymous display name (在 backend-jobs-recommendations BuildJobMatchProfile 中作为 cross-owner failure fallback 处理；不让单一 cross-owner 缺失阻塞整个 endpoint)
- 修订 [backend-auth spec.md](../../../backend-auth/spec.md) 在模块边界表追加 internal API 行 + §3 新增 D-X 决策行 "cross-owner identity internal API for backend-jobs-recommendations aggregation"
- 修订 [backend-auth history.md](../../../backend-auth/history.md) 追加 1.X 行记录本 cross-owner additive
- backend-auth unit test PASS：(1) seeded user 返回完整 identity 字段；(2) emailMasked 格式断言（contains `***`，不含 `@<original>` 中的 local-part 字符）；(3) 不存在 userId 返回 ErrUserNotFound；(4) 调用不写 audit_events / 不 bump 任何 user 字段

### Phase 1: getJobMatchProfile + getAgentScanStatus + 5 个 cross-owner counter integration

#### 1.1 实现 `backend/internal/jdmatch/handler/get_profile.go`
- 实现 generated server interface `GetJobMatchProfile`
- 调 `service.BuildJobMatchProfile(ctx, userID)` 聚合并返回 `JobMatchProfile`（spec D-18 锁定字段来源映射）
- 字段映射 (P0 baseline)：
  - `displayName` ← `backend-auth.GetUserIdentityForUser(ctx, userID).displayName`（必填非 null；失败时 fallback 到非 PII anonymous display name）
  - `avatarUrl` ← P0 baseline `null`（D-18；P1 由 user 头像 / Gravatar 派生）
  - `headline` ← `backend-profile.GetCandidateProfileForUser(ctx, userID).headline`（缺失 null）
  - `yearsOfExperience` ← `backend-profile.GetCandidateProfileForUser(ctx, userID).yearsOfExperience`（缺失 null）
  - `locationText` ← P0 baseline `null`（D-18；P1 由 candidate_profile.region + remote 偏好格式化）
  - `compensationText` ← P0 baseline `null`（D-18；P1 由 target_jobs 期望薪资聚合）
  - `skills` ← P0 baseline `[]`（D-18；P1 由 candidate_profile + resume_versions.structured_profile 聚合去重，需要后续 plan 的新 cross-owner additive）
  - `sources` ← 4 个 counter API 真实计数（resumes / jds / mocks / debriefs）

#### 1.2 实现 `backend/internal/jdmatch/service/build_profile.go`
- `BuildJobMatchProfile(ctx context.Context, userID string, deps)` orchestrator：调用 7 个 cross-owner internal API（backend-auth identity + 2 个 backend-profile read + 4 个 counter），聚合并返回 `JobMatchProfile` value object
- 字段映射严格按 spec D-18；不引入 `region` / `preferredPracticeLanguage` 等 JobMatchProfile schema 不存在的字段
- 错误处理：任一 cross-owner API 失败回退到 default（如 `backend-auth.GetUserIdentityForUser(ctx, userID)` 失败返回非 PII anonymous display name，例如 `Candidate`，不得返回空字符串或 raw email；sources.resumes=0 而非整个 endpoint 失败）；记录 audit warn（不含 PII）

#### 1.3 实现 `backend/internal/jdmatch/handler/get_agent_status.go`
- 实现 generated server interface `GetAgentScanStatus`
- 调 store `agent_scans.GetLatestForUser(userId)` 读最近行
- 若用户从未触发 scan，返回 `{status:'idle', lastScanAt:null, nextScanAt:null, message:null}`（懒触发 scan 在 list recommendations 时按 D-3 进行）
- 返回 `AgentScanStatus`

#### 1.4 实现 `backend/internal/jdmatch/store/agent_scans.go`
- Repository: `GetLatestForUser(userId)` / `Create(userId, status)` / `UpdateStatus(id, status, ...)` / `DeleteForUser(userId)`

#### 1.5 unit test + integration test
- `get_profile_test.go`: 聚合 happy path + cross-owner failure fallback + skills 去重
- `get_agent_status_test.go`: 首次 idle / 已有 scan 行
- `agent_scans_integration_test.go`: CRUD + cross-user isolation

### Phase 2: listJobRecommendations + getJobRecommendation + markJobNotRelevant + recommendation generation

#### 2.1 实现 `backend/internal/jdmatch/store/recommendations.go`
- Repository: `ListByUser(userId, cursor, pageSize)` / `GetByIdForUser(jobMatchId, userId)` / `Upsert(userId, recommendation)` / `MarkDismissed(jobMatchId, userId, reason, freeNote)` / `DeleteForUser(userId)` / `CountActiveByUser(userId)`
- ListByUser 排序：`score DESC, recommended_at DESC, id DESC`；过滤 `dismissed_at IS NULL AND deleted_at IS NULL`
- cursor pagination

#### 2.2 实现 `backend/internal/jdmatch/handler/list_recommendations.go`
- 实现 generated server interface `ListJobRecommendations`
- cursor pagination + cross-user
- 投影 `jd_match_recommendations` row → `JobMatchRecommendation` schema（含 provenance）

#### 2.3 实现 `backend/internal/jdmatch/handler/get_recommendation.go`
- 实现 generated server interface `GetJobRecommendation`
- cross-user 404 + RESOURCE_NOT_FOUND
- 返回完整 recommendation（包含 list 可能省略的详细字段）

#### 2.4 实现 `backend/internal/jdmatch/handler/mark_not_relevant.go`
- 实现 generated server interface `MarkJobNotRelevant`
- IK 校验
- 调 store `MarkDismissed(jobMatchId, userId, reason, freeNote)`
- 返回 `MarkNotRelevantResult{ jobMatchId, dismissedAt }`
- freeNote 不进 log / audit / outbox

#### 2.5 实现 `backend/internal/jdmatch/generators/recommendation.go`
- `RunRecommendationGenerator(ctx, userId)` **service** 函数（**不是 canonical job_type**；spec D-12 锁定的 2 个 canonical job_type 是 `jd_match_agent_scan` + `jd_match_search`，没有 `jd_match_recommendation`）；本 generator 由 `jd_match_agent_scan` job handler 内联调用（per user batch），不暴露独立 job_type / drainer 注册
- 通过 [A3 AIClient](../../../ai-provider-and-model-routing/spec.md) 调 F3 `jd_match.recommendation` feature_key（feature_key 是 F3 prompt routing key，**不要求对应 canonical job_type**）
- 解析 LLM JSON 输出 → upsert `jd_match_recommendations` 行
- 写入 `ai_task_runs` typed columns（`task_type='jd_match_recommendation_generation'` 或类似命名，由 owner 决定）
- 成功完成时通过 outbox 写入 `jd_match.recommendation.completed` event（由 agent_scan job 在 generator 返回后发射）；失败路径不发 completed event
- agent_scans 行的 status / last_scan_at / next_scan_at 更新由调用方 `jd_match_agent_scan` job handler 负责（见 §5.3）

#### 2.6 unit test
- `list_recommendations_test.go`: 空 / 25 行 + cursor / dismissed 过滤 / cross-user / 排序
- `get_recommendation_test.go`: 200 / 404 cross-user / 404 not exist
- `mark_not_relevant_test.go`: 成功 / IK replay / IK conflict / freeNote 不泄漏
- `generators/recommendation_test.go`（stub AIClient）: 成功 / output_invalid / timeout retryable / 输出投影到 `jd_match_recommendations` upsert / ai_task_runs typed columns 写入

### Phase 3: listWatchlist + addToWatchlist + removeFromWatchlist + tone derivation

#### 3.1 实现 `backend/internal/jdmatch/store/watchlist.go`
- Repository: `ListByUser(userId)` / `Add(userId, linkedJobMatchId)` / `Remove(userId, jobMatchId)` / `DeleteForUser(userId)`
- ListByUser JOIN `jd_match_recommendations` 取 title / company / score（用于 tone 派生）
- UNIQUE (user_id, linked_job_match_id) 约束

#### 3.2 实现 `backend/internal/jdmatch/handler/list_watchlist.go`
- 实现 generated server interface `ListWatchlist`
- tone 派生（Q-4 默认规则）：score ≥ 80 → ok / 50-79 → warn / < 50 → muted
- 返回 `WatchlistResponse{items}`

#### 3.3 实现 `backend/internal/jdmatch/handler/add_watchlist.go`
- 实现 generated server interface `AddToWatchlist`
- IK 校验
- UNIQUE 处理：重复 add 返回首次 item（不创建新行）
- 验证 linkedJobMatchId 属于当前用户（cross-user 404）

#### 3.4 实现 `backend/internal/jdmatch/handler/remove_watchlist.go`
- 实现 generated server interface `RemoveFromWatchlist`
- IK 校验
- DELETE WHERE user_id = current AND linked_job_match_id = path param
- 返回 204
- cross-user 删除返回 404

#### 3.5 unit test + integration test
- `watchlist_integration_test.go`: add / list / remove / UNIQUE / cross-user
- `add_watchlist_test.go`: 成功 / IK replay / UNIQUE 重复加入 / cross-user 404
- `remove_watchlist_test.go`: 成功 / IK replay / 不存在 / cross-user
- `list_watchlist_test.go`: tone 派生 / cross-user / empty

### Phase 4: searchJobs + listSavedSearches + createSavedSearch

#### 4.1 实现 `backend/internal/jdmatch/store/saved_searches.go`
- Repository: `ListByUser(userId, cursor, pageSize)` / `Create(userId, label, query, filters)` / `UpdateRunInfo(id, lastRunAt, newJobsCount)` / `DeleteForUser(userId)`

#### 4.2 实现 `backend/internal/jdmatch/store/search_runs.go`
- Repository: `Create(userId, query, filters, provenance, resultCount)` / `DeleteForUser(userId)`

#### 4.3 实现 `backend/internal/jdmatch/handler/search.go`
- 实现 generated server interface `SearchJobs`
- IK 校验
- 同步调 [A3 AIClient](../../../ai-provider-and-model-routing/spec.md) F3 `jd_match.search` feature_key + 内部 jobs 池 rank
- timeout 30s → 502 + AI_PROVIDER_TIMEOUT
- 写入 `jd_match_search_runs` 行（provenance / result_count / query / filters）
- 返回 `SearchJobsResponse{searchRunId, items}`；items 通过 JOIN 已有 `jd_match_recommendations`（不创建 ephemeral 行）
- query 不进 log / audit / outbox

#### 4.4 实现 `backend/internal/jdmatch/handler/list_saved.go`
- 实现 generated server interface `ListSavedSearches`
- 返回 `SavedSearchesResponse{items}`

#### 4.5 实现 `backend/internal/jdmatch/handler/create_saved.go`
- 实现 generated server interface `CreateSavedSearch`
- IK 校验
- 必填字段校验（label / query）
- 返回 `SavedSearch`

#### 4.6 unit test
- `search_test.go`（stub AIClient）: 成功 / timeout 502 / IK replay
- `list_saved_test.go`: 空 / 多行 / cross-user
- `create_saved_test.go`: 成功 / IK replay / 必填校验

### Phase 5: getMarketSignals + agent_scan 后台 job + privacy delete + cmd/api wiring

#### 5.1 实现 `backend/internal/jdmatch/service/market_signals.go`
- `BuildMarketSignals(userId, window)`：从 `watchlist_items` + `jd_match_recommendations` + 内部 jobs 池统计派生 4 个 signal
- baseline 4 个 signal 示例（待 fixture 锁定）：`new_jobs_this_week` / `avg_score_movement` / `top_company_intent` / `compensation_band`
- tone 派生：每个 signal 按业务规则

#### 5.2 实现 `backend/internal/jdmatch/handler/get_market_signals.go`
- 实现 generated server interface `GetMarketSignals`
- window 参数（7d / 14d / 30d）；default 7d
- 返回 `MarketSignalsResponse{signals, asOf}`

#### 5.3 实现 `backend/internal/jdmatch/jobs/agent_scan.go`
- `jd_match.agent_scan` 后台周期 job（canonical job_type `jd_match_agent_scan` per spec D-12 + Phase 0.4 jobs.yaml additive）
- 触发条件（D-3）：周期（4h）/ 增量（new resume / target_job / debrief 触发）/ 用户 lazy（list recommendations 时距上次 > 4h）
- 内部逻辑：per user batch 调用 `generators.RunRecommendationGenerator(ctx, userId)`（§2.5）；generator 返回成功后 agent_scan job 发射 `jd_match.recommendation.completed` outbox event
- 更新 `agent_scans` 行：开始时 status='scanning' + started_at；成功结束时 status='idle' + last_scan_at + next_scan_at；失败时 status='error' + error_message
- 错误路径：status='error' + error_message；不发 `jd_match.recommendation.completed` event；写 audit failure（不含 prompt 内容 / PII）

#### 5.4 实现 `backend/internal/jdmatch/service/privacy.go`
- `DeleteJobMatchDataForUser(userId)`:
  1. 调 store `watchlist_items.DeleteForUser(userId)` (count)
  2. 调 store `saved_searches.DeleteForUser(userId)` (count)
  3. 调 store `jd_match_search_runs.DeleteForUser(userId)` (count)
  4. 调 store `jd_match_recommendations.DeleteForUser(userId)` (count)
  5. 调 store `agent_scans.DeleteForUser(userId)` (count)
  6. 写入 audit_events tombstone（userId / 5 个 count / 删除时间 / job_id）；不含 query / label / reasons / source_url / freeNote
  7. 事务失败回滚 + 写 audit failure

#### 5.5 `cmd/api` runtime wiring
- 新增 `buildJDMatchRuntime`（或等价 composition helper），组合 jdmatch store / cross-owner counter services / AIClient / F3 registry / idempotency middleware / 1 个 `jd_match_agent_scan` 后台 drainer handler
- 挂载：
  - `GET /api/v1/jd-match/profile` → session middleware + `GetJobMatchProfile`
  - `GET /api/v1/jd-match/agent-status` → session + `GetAgentScanStatus`
  - `GET /api/v1/jd-match/recommendations` → session + `ListJobRecommendations`
  - `GET /api/v1/jd-match/recommendations/{jobMatchId}` → session + path param + `GetJobRecommendation`
  - `POST /api/v1/jd-match/recommendations/{jobMatchId}/dismiss` → session + IK middleware + `MarkJobNotRelevant`
  - `GET /api/v1/jd-match/watchlist` → session + `ListWatchlist`
  - `POST /api/v1/jd-match/watchlist` → session + IK + `AddToWatchlist`
  - `DELETE /api/v1/jd-match/watchlist/{jobMatchId}` → session + IK + path param + `RemoveFromWatchlist`
  - `POST /api/v1/jd-match/search` → session + IK + `SearchJobs`
  - `GET /api/v1/jd-match/saved-searches` → session + `ListSavedSearches`
  - `POST /api/v1/jd-match/saved-searches` → session + IK + `CreateSavedSearch`
  - `GET /api/v1/jd-match/market-signals` → session + `GetMarketSignals`
- in-process drainer：spec D-12 锁定 2 个 canonical job_type — drainer 实际**只注册 1 个后台 job handler `jd_match_agent_scan`**（per spec D-12 + Phase 0.4 jobs.yaml additive）；`jd_match_search` 虽然是 canonical job_type 但 P0 走 sync HTTP handler（Phase 4.3），不在 drainer 注册（保留 job_type 是为 Q-3 / future P1 改异步留预占）；`recommendation` 不是 canonical job_type，由 agent_scan job 内联调用 `generators.RunRecommendationGenerator`（§2.5）
- `Start(ctx)` / `Shutdown(ctx)` 必须随 `cmd/api` lifecycle 管理；shutdown 不泄漏 goroutine（test 验证）

#### 5.6 unit test + integration test
- `market_signals_test.go`: 4 个 signal 派生 + tone + asOf
- `agent_scan_test.go`: 周期 / 增量 / lazy 触发
- `privacy_test.go`: 5 表删除顺序 + audit tombstone 完整 + 无敏感字段泄漏 + 失败回滚
- `cmd/api` test：12 route 真实可达 + auth 401 + IK middleware + drainer 启停

### Phase 6: 收口 + BDD + cross-owner handoff

#### 6.1 跨 gate 收口

按 §3 替代验证 gate 依序运行：
- `cd backend && go test ./...` PASS
- `cd backend && go test ./internal/jdmatch/...` PASS
- `cd backend && go test ./cmd/api -run 'TestBuildJDMatchRuntime|TestJDMatchHTTPScenario|TestJDMatchAgentScanDrainerScenario' -count=1` PASS
- mock-first 对齐：11 个非 profile endpoint 通过 `cmd/api` 真实 route 响应与对应 fixture default scenario 字节比对 PASS，`getJobMatchProfile` 通过 D-19 structural parity PASS
- migration gate：`migrations/lint.sh` + `make migrate-check` + `make privacy-delete-dry-run` PASS（无 `DATABASE_URL` 时记录 live DB blocker，不得把 `make migrate-check` 记为全绿）
- F3 gate：prompt loader test + `make lint-ai-profile-coverage` PASS
- A3 gate：`make lint-ai-profile-coverage` + provider registry runtime bootstrap PASS
- B3 gate：events / jobs generator + baseline manifest PASS
- 4 个 cross-owner counter unit test PASS
- grep `LinkedIn|Boss|脉脉|拉勾|mistake|growth|drill|experiences|star` in `backend/internal/jdmatch/` + outbox payload + DB seed：0 命中（C-16 negative）
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS
- `make docs-check` PASS

#### 6.2 BDD 场景验证

- 执行 `test/scenarios/e2e/p0-094-jd-match-profile-and-recommendations-list/` 全 PASS
- 执行 `test/scenarios/e2e/p0-095-jd-match-watchlist-and-saved-search-lifecycle/` 全 PASS
- 执行 `test/scenarios/e2e/p0-096-jd-match-search-and-market-signals/` 全 PASS
- 执行 `test/scenarios/e2e/p0-097-jd-match-agent-scan-and-privacy-delete/` 全 PASS
- 在 `test/scenarios/e2e/INDEX.md` 追加 P0.094 / P0.095 / P0.096 / P0.097 行

#### 6.3 cross-owner handoff 信号

通知下游 owner：
- [frontend-home-job-picks-and-parse/002-jd-match-recommendations](../../../frontend-home-job-picks-and-parse/plans/002-jd-match-recommendations/plan.md) owner：12 个 JobMatch endpoint real backend 已可用，可启动 frontend fixture-backed transport → real transport 切换原地修订（plan 002 spec 1.X → 1.Y / plan checklist 切真改造）
- backend internal privacy runner owner：`DeleteJobMatchDataForUser` internal API 已可用，可纳入 privacy_delete job dispatcher 链路
- B4 / F3 / A3 / B3 / backend-resume / backend-targetjob / backend-practice / backend-debrief owners：cross-owner additive 已落地，本 plan 在各自 spec / history 中追加完成行

本 plan 不直接修订下游 owner code，只在 6.3 完成 "可消费" 信号传递；frontend owner 与各 cross-owner owner 独立完成各自后续修订。

#### 6.4 spec / history / INDEX 同步

- backend-jobs-recommendations spec.md 维持 1.2 active；本 completion 不新增 D-20，只同步更新日期与 INDEX 投影。
- backend-jobs-recommendations history.md 追加 v1.5 "post-reopen completion" 行：记录 12 个 endpoint real backend、drainer、fixture parity、E2E.P0.094-097 Ready/automated 与 cross-owner handoff。
- `docs/spec/INDEX.md` §5 P0 Implementation 表中 `backend-jobs-recommendations` 行同步到 spec Header：1.2 active 2026-05-22。
- [engineering-roadmap §5.2](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) "Home / Job Picks / Parse" workstream 行从 deferred implementation wording 调整为 `backend-jobs-recommendations active（001 real backend baseline completed）`；roadmap 从 3.18 bump 到 3.19 并在 engineering-roadmap history.md 追加对应 completion 行。

## 5 验收标准

- 本计划列出的 §4 所有 Phase task 全部完成（包含 Phase 0 cross-owner additive 5 个 owner = backend-auth identity + backend-resume / backend-targetjob / backend-practice / backend-debrief counter）
- §3 替代验证 gate 全部通过
- spec §6 C-1..C-19 全部 PASS（含 C-19 cross-owner identity + counter additive 验证）
- `cmd/api` route/runtime gate PASS：session middleware、IK middleware（5 op）、12 个 route 真实可达、in-process drainer 注册 1 个后台 job handler `jd_match_agent_scan` + 启停 / shutdown 不泄漏 goroutine、均有测试证据
- 4 个 cross-owner additive 全部 PASS：B4 migration + F3 prompt + A3 profile + B3 events
- 5 个 cross-owner internal API additive（backend-auth identity + backend-resume / backend-targetjob / backend-practice / backend-debrief counter）unit test PASS
- BDD E2E.P0.094 + E2E.P0.095 + E2E.P0.096 + E2E.P0.097 全 PASS（含 E2E.P0.094 7 个 cross-owner API trace + D-19 structural parity 断言）
- 下游 frontend owner / backend internal privacy runner owner / 5 个 cross-owner additive owner 已收到 real backend / privacy API / internal API 可用信号
- `docs/spec/INDEX.md` + engineering-roadmap §5.2 + engineering-roadmap history.md 已同步本 subject `plan 001 完成` 状态描述 + roadmap 版本 bump（与 backend-profile/001 完成顺序协同，见 §6.4）

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: cross-owner additive 范围大（B4 + F3 + A3 + B3 + 4 个 counter API）导致 review 链长 | Phase 0 优先完成所有 cross-owner additive；与各 owner 提前对齐签名；本 plan checklist 把每个 owner 修订作为独立 item，可独立 review |
| R2: 内部 jobs 池 seed 缺失导致 baseline 推荐为空 | Phase 0 配套提供 seed fixture（50-100 个 mock job posting）；fixture 通过 `make validate-fixtures`；seed 数据通过 store layer 写入而非直接 SQL |
| R3: AI 调用 timeout / output_invalid 比例高 | F3 baseline prompt 包含 output schema example + retry 策略 + fallback 到 cache（最近一次成功 recommendation）；ai_task_runs 跟踪 retryability |
| R4: agent scan 频率过高导致 AI cost burst | A4 config 提供 `JD_MATCH_AGENT_SCAN_INTERVAL` tunable；baseline 4 小时；增量触发去重（同用户 5 分钟内重复触发合并） |
| R5: searchJobs 30s timeout 影响用户体验 | baseline accept；Q-3 已记录 P1 改 async；frontend 在 inflight 期间渲染 5 步 AGENT panel（[frontend D-12](../../../frontend-home-job-picks-and-parse/spec.md#3-用户决策--待确认事项)） |
| R6: privacy delete 5 表顺序错误导致 FK 违反 | Phase 5.4 严格按 watchlist → saved_searches → search_runs → recommendations → agent_scans 顺序；FK CASCADE 由 migration 兜底；integration test 覆盖 |
| R7: B4 cross-owner additive 触发 breaking-change gate | additive only（5 张新表 + 新 enum 来源）；不修改既有表；migration lint / drift gate 应允许新表 additive；若误判，先修订 B4 diff-config |
| R8: F3 prompt 文本 baseline 质量影响推荐 | F3 baseline prompt 包含可用文案（不写 TBD 占位）；future plan 通过离线评估集 ≥ 50 题逐步优化 |
| R9: 多 worker 并发触发 agent_scan 重复 | agent_scans 表使用 SELECT FOR UPDATE 锁 + status='scanning' 唯一性；同用户同时只有 1 个 scanning 行 |
| R10: 外部 watcher 误以为 baseline 接入外部平台 | spec §2.2 + plan §1 明确 baseline 不接 LinkedIn/Boss/脉脉/拉勾 等；C-16 negative grep 强制；外部平台扩展归独立 P2 plan |
| R11: frontend 切真时发现字段缺失 / 字节漂移 | 11 个非 profile endpoint 严格按 fixture default scenario 语义等价验证（字段集、显式 null、status、`X-Request-ID`），`getJobMatchProfile` 按 D-19 structural parity 验证；frontend 002 plan 通过 generated client + same fixture 自动验证；本 plan 6.3 主动信号 |
| R12: counter cross-owner API 形态不一致 | spec §4.4 锁定签名 `Count*ForUser(ctx context.Context, userID string) (int, error)`；各 owner 实现必须匹配；Phase 0.5 检查所有 owner 一致 |
| R13: backend-auth identity additive 泄漏 raw email 或越权 | spec D-17 锁定 `UserIdentity` 字段集（仅 displayName / avatarUrl / emailMasked）；Phase 0.6 backend-auth unit test 强制 `emailMasked` 格式断言（含 `***` / 不含 raw local-part 字符）；调用不写 audit_events（防止跨域读取被审计放大）；cross-user 隔离由 caller userId 保证，调用方（jdmatch）必须先通过 session middleware 解析得到 current_user_id 再传入 |
| R14: P0 baseline JobMatchProfile 稀疏字段引起 frontend 切真"看起来缺字段"误判 | spec D-18 / D-19 锁定稀疏 baseline；[B2 fixture `partial-profile` scenario](../../../openapi-v1-contract/spec.md) 已展示稀疏形态作为 design preview；frontend `frontend-home-job-picks-and-parse/002-jd-match-recommendations` 切真后应该按 partial-profile shape 处理（[docs/ui-design/module-job-workspace.md](../../../../ui-design/module-job-workspace.md) graceful render null）；handoff 信号 §6.3 主动告知 frontend owner P0 baseline 字段稀疏度 |
| R15: drainer 注册 job_type 数量与 spec D-12 不一致 | spec D-12 锁定 2 个 canonical job_type（agent_scan + search）；§5.5 显式说明 drainer P0 只注册 1 个后台 job handler（agent_scan），search 走 sync HTTP handler，recommendation 不是独立 job_type；route wiring test 验证 drainer registry 实际只含 1 个后台 entry，避免 §5.5 / Phase 2.5 描述漂移 |
