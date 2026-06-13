# Backend Jobs Recommendations Spec

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-06-13

> **D-17 退役声明（2026-06-12 product-scope v2.1）**：岗位推荐模块（`jd_match` / JobMatch / Job Picks）已被 [product-scope D-17](../product-scope/spec.md#31-已锁定决策) 整体删除——JD 获取唯一入口是首页导入，一级导航收敛为四项，Q-2 关闭，`全球多平台搜岗` 规划例外随之丢弃。本 spec 自 v2.0 起转为该删除的 backend / 契约侧 owner：§9 定义删除范围与零残留验收 gate；§1-§8 保留为历史 baseline 记录（描述的能力已退役，不得作为新实现依据）。前端删除归 [frontend-home-job-picks-and-parse/002](../frontend-home-job-picks-and-parse/plans/002-jd-match-recommendations/plan.md)。

## 1 背景与目标（历史 baseline 记录，已随 D-17 退役）

[frontend-home-job-picks-and-parse spec §2.2 + §3.1 D-1](../frontend-home-job-picks-and-parse/spec.md#22-out-of-scope) 与 [openapi-v1-contract §3.1.1](../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) 都已显式预占：JobMatch tag 12 个 operationId（`getJobMatchProfile` / `getAgentScanStatus` / `listJobRecommendations` / `getJobRecommendation` / `markJobNotRelevant` / `listWatchlist` / `addToWatchlist` / `removeFromWatchlist` / `searchJobs` / `listSavedSearches` / `createSavedSearch` / `getMarketSignals`）的真实 backend 由独立未来 subspec `backend-jobs-recommendations` 承接；frontend 当前通过 generated client + fixture-backed transport 闭环，不依赖也不实现真实 backend。本 spec 是该承接的正式入口。

`backend-jobs-recommendations` 是 JD-Match 业务域的后端 owner：

1. **承接 12 个 JobMatch endpoint 真实 backend**：消除当前 frontend fixture-backed transport 与未来真实 backend 之间的契约缺口；11 个响应与 [B2 fixtures](../openapi-v1-contract/spec.md) `JobMatch/*.json` 当前 default scenario 字节级对齐，`getJobMatchProfile` 按 D-19 structural parity 验证，frontend 切真时无需修订。
2. **构建 candidate 画像聚合 + JD-Match agent baseline**：从 [backend-profile](../backend-profile/spec.md) candidate_profile 读取 headline / yearsOfExperience，从 experience_cards 读取内部质量信号 / P1 扩展锚点，并从 [backend-resume](../backend-resume/spec.md) resume_assets / resume_versions、[backend-targetjob](../backend-targetjob/spec.md) target_jobs、[backend-practice](../backend-practice/spec.md) practice_sessions、[backend-debrief](../backend-debrief/spec.md) debriefs 通过 internal API 聚合当前 4 项 `sources` 计数。
3. **JD-Match 推荐编排**：基于 candidate 画像 + 内部候选 jobs 池，通过 [A3 AIClient](../ai-provider-and-model-routing/spec.md) + 本 spec 新增 [F3 feature_key](../prompt-rubric-registry/spec.md) `jd_match.recommendation` 生成 `JobMatchRecommendation`（含 score / reasons / risks / highlights / GenerationProvenance）。
4. **Agent scan baseline**：后台周期性 job (`jd_match.agent_scan`) 刷新推荐；前端 polling `getAgentScanStatus` 获取 `idle / scanning / error` 状态；不引入 SSE/WebSocket（与 [frontend D-10](../frontend-home-job-picks-and-parse/spec.md#3-用户决策--待确认事项) 一致）。
5. **Watchlist + Saved Search 服务端持久化**：[frontend D-9](../frontend-home-job-picks-and-parse/spec.md#3-用户决策--待确认事项) 已锁定服务端持久化；本 subject 承接 `watchlist_items` / `saved_searches` 表与对应 endpoint。
6. **自然语言搜索 baseline**：`searchJobs` 通过 AI 在**内部候选 jobs 池**（baseline 阶段由 seed / fixture 提供）中匹配相关 JD；**不直连** LinkedIn / Boss 直聘 / 脉脉 / 拉勾 / 公司官网等外部招聘平台 API（与 [product-scope §3.2 Q-2](../product-scope/spec.md#32-待确认事项) + [§7.3 P2](../product-scope/spec.md#73-p2工程化和数据源扩展) "全球多平台搜岗作为规划例外延后" 一致）。
7. **Market signals + privacy + 删除链路**：4-card 市场信号聚合；search query / watchlist label / sourceJobUrl / freeNote 等敏感字段不进 log / audit / outbox；privacy delete 链路 cascade 清空用户 jd_match 数据。
8. **mock-first 切真**：本 subject 实现的 handler 以 11 strict byte parity + `getJobMatchProfile` structural parity（D-19）对齐 [B2 fixtures](../openapi-v1-contract/spec.md) `JobMatch/*.json`，frontend 通过 generated client 切真不感知 transport 差异。

本 subject **不实现**：前端 UI（归 [frontend-home-job-picks-and-parse/002-jd-match-recommendations](../frontend-home-job-picks-and-parse/plans/002-jd-match-recommendations/plan.md)）；真实外部招聘平台 API 接入（归后续 P2 plan）；岗位推荐"全球多平台搜岗"扩展（[product-scope §7.3 P2](../product-scope/spec.md#73-p2工程化和数据源扩展) 规划例外，独立设计）；隐形实时面试辅助（[product-scope 已丢弃](../product-scope/spec.md#611-已丢弃能力)）。

## 2 范围

### 2.1 In Scope

- **HTTP handler + runtime wiring**：实现 [B2 §3.1.1](../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) JobMatch tag 12 个 operationId，在 `cmd/api` 按当前 session middleware / IK middleware / generated response envelope 口径挂载真实 route。
- **store layer**：新建 5 张表（由本 plan 携带 [B4 cross-owner additive migration](../db-migrations-baseline/spec.md)）：
  - `jd_match_recommendations`：AI 推荐结果存储（score / reasons / risks / highlights / source_url / source_label / interview_hypotheses / network_note / similar_interviewers / seen / dismissed_at / provenance）
  - `watchlist_items`：用户收藏的推荐（linked_job_match_id / label / tone / change）
  - `saved_searches`：用户保存的搜索（label / query / filters jsonb / new_jobs_count / last_run_at）
  - `agent_scans`：agent scan job 状态（status / started_at / finished_at / last_scan_at / next_scan_at / error_message）
  - `jd_match_search_runs`：自然语言搜索运行（search_run_id / query / filters jsonb / result_count / provenance）
- **AI 编排**：
  - `jd_match.recommendation` generator/service：由 `jd_match.agent_scan` job handler 内联调用，不新增独立 canonical job_type；周期性 / on-demand 调 [A3 AIClient](../ai-provider-and-model-routing/spec.md) 生成推荐；通过本 spec 新增 [F3 `jd_match.recommendation` feature_key](../prompt-rubric-registry/spec.md) 注入 prompt / rubric / model profile；成功完成发射 `jd_match.recommendation.completed` 事件
  - `jd_match.search` 同步 / 准异步调用：用户提交 search 时通过 [A3 AIClient](../ai-provider-and-model-routing/spec.md) + 新增 [F3 `jd_match.search` feature_key](../prompt-rubric-registry/spec.md) 在内部 jobs 池中匹配并 rank
  - `jd_match.agent_scan` 后台 job：周期性触发推荐刷新；by polling 而非 SSE / WebSocket
- **画像聚合**：实现 `BuildJobMatchProfile(ctx context.Context, userID string)` internal service，通过 cross-owner internal API 聚合 7 个 owner：
  - `backend-auth.GetUserIdentityForUser(ctx context.Context, userID string) (UserIdentity, error)` → `JobMatchProfile.displayName` / `avatarUrl` （D-17 锁定；read-only；不写 audit；不返回 raw email；与 [B2 `UserContext`](../openapi-v1-contract/spec.md) `displayName` 字段同源）
  - `backend-profile.GetCandidateProfileForUser(ctx context.Context, userID string) (*api.CandidateProfile, error)` → `JobMatchProfile.headline` / `yearsOfExperience`（已锁定 [backend-profile D-13](../backend-profile/spec.md#31-已锁定决策)；read-only 不触发 seed 副作用）
  - `backend-profile.CountExperienceCardsBySource(ctx context.Context, userID string) (profile.SourceCounts, error)` → experience cards 多源计数（暂不入 sources schema，留作 P1 扩展，已锁定 [backend-profile D-11](../backend-profile/spec.md#31-已锁定决策)）
  - `backend-resume.CountResumesForUser(ctx context.Context, userID string) (int, error)` → `sources.resumes`
  - `backend-targetjob.CountTargetJobsForUser(ctx context.Context, userID string) (int, error)` → `sources.jds`
  - `backend-practice.CountPracticeSessionsForUser(ctx context.Context, userID string) (int, error)` → `sources.mocks`
  - `backend-debrief.CountDebriefsForUser(ctx context.Context, userID string) (int, error)` → `sources.debriefs`
  - 上述 cross-owner internal API 中已存在的（backend-profile.GetCandidateProfileForUser/CountExperienceCardsBySource 由 [backend-profile/001](../backend-profile/plans/001-candidate-profile-and-experience-cards/plan.md) 提供）直接调用；缺失的（backend-auth identity / backend-resume / backend-targetjob / backend-practice / backend-debrief 5 个 count/identity API 当前未暴露）由本 subject 携带 cross-owner additive 增补
  - **D-18 baseline 字段映射**：`JobMatchProfile.locationText` / `compensationText` / `avatarUrl` P0 baseline 一律返回 `null`；`JobMatchProfile.skills` P0 baseline 返回 `[]`；这些字段的真实派生（candidate_profile.region 格式化 / target_jobs 聚合 / resume_versions.structured_profile.skills 聚合）归后续 plan，本 subject 不在 P0 引入额外 cross-owner additive 读取 structured_profile
- **隐私链路**：
  - search query / watchlist label / saved-search label / sourceJobUrl / markNotRelevant freeNote 不进 log / audit / outbox payload；只通过 store 行内 jsonb / text 列存储
  - privacy_delete 调用 `DeleteJobMatchDataForUser(userId)` internal API，cascade 删除 5 张表所有 user-owned 行；audit tombstone 仅写 userId / 各表删除计数 / 时间，不含字段值
- **B3 events 发射**：
  - `jd_match.recommendation.completed`（agent_scan 内联 recommendation generator 成功结束时）
  - `jd_match.search.completed`（search run 完成时）
  - envelope 字段集与 [B3 §3.1.4 PII 边界](../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory) 一致；不含 query / reasons / source_url 等敏感字段
- **mock-first 对齐**：本 subject 实现的 12 个 handler 中，11 个 handler 响应字段集 / status code / 字段顺序与 [B2 fixtures](../openapi-v1-contract/spec.md) `JobMatch/*.json` default scenario 字节比对；`getJobMatchProfile` 按 D-19 structural parity 验证；[mock-contract-suite C-9](../mock-contract-suite/spec.md#6-验收标准) 强制 enforce。

### 2.2 Out of Scope

- **前端 UI**：归 [frontend-home-job-picks-and-parse/002-jd-match-recommendations](../frontend-home-job-picks-and-parse/plans/002-jd-match-recommendations/plan.md)；本 subject 不实现 React 组件。
- **真实外部招聘平台 API 接入**：LinkedIn / Boss 直聘 / 脉脉 / 拉勾 / 公司官网 API 直连归后续 `backend-jobs-recommendations/002-external-sources` 或类似 P2 plan；本 subject baseline 只在**内部 jobs 池**（由 seed / fixture 提供，未来由数据源接入扩展）中推荐 / 搜索。
- **岗位推荐"全球多平台搜岗"扩展**：[product-scope §7.3 P2](../product-scope/spec.md#73-p2工程化和数据源扩展) 规划例外；进入实施前必须先解决数据源接入、合规、质量评估、维护成本，由独立后续 plan 设计。
- **company intel 完整聚合 API**：[ui-design module-job-workspace.md](../../ui-design/module-job-workspace.md) `company_intel` page 的数据接口当前由 [backend-targetjob](../backend-targetjob/spec.md) 承接基础字段；JD-Match recommendation `networkNote` / `similarInterviewers` / `interviewHypotheses` 由本 subject 在 recommendation generation 中产出，但深度 company intel（公司新闻 / 招聘历史 / 面试题库）归后续 plan。
- **AI provenance 与 F3 evaluation set**：本 baseline 只确保 recommendation / search 输出含 `GenerationProvenance`；prompt / rubric quality 评估、离线评估集 ≥ 50 题归 [prompt-rubric-registry](../prompt-rubric-registry/spec.md) 后续 plan。
- **多语言 personalized rank**：baseline 推荐按 `candidate_profile.preferred_practice_language` 过滤；多语言混合排序、跨语言权重等归后续 plan。
- **watchlist label / saved-search label / search query 全文检索能力**：本 baseline 只支持精确字符串匹配；ElasticSearch / pgvector 等全文检索能力归后续 plan。
- **server-sent events / WebSocket 实时推送**：[frontend D-10](../frontend-home-job-picks-and-parse/spec.md#3-用户决策--待确认事项) 已锁定 polling 模式，frontend 不消费 SSE/WebSocket；本 subject 不实现实时推送 endpoint。
- **隐形实时面试辅助**：[product-scope §4.4 + §6.12](../product-scope/spec.md#44-伦理与安全约束) 明确丢弃。
- **agent scan 调度配置 UI / 高级管理面板**：用户层面只通过 `getAgentScanStatus` 看到状态；调度策略由 backend 内部决定 + A4 config 提供 tuning，不暴露 admin endpoint。
- **JobMatchProfile 富字段派生 (P0)**：`avatarUrl`（来自 user 上传头像或 Gravatar）/ `locationText`（来自 candidate_profile.region 国际化格式化 + remote 偏好叠加）/ `compensationText`（来自 target_jobs 期望薪资区间聚合）/ `skills`（来自 candidate_profile + resume_versions.structured_profile 聚合去重）P0 baseline 一律返回 null/`[]`，归后续 plan（`backend-profile/002-profile-insights-and-corrections` 或新 `backend-jobs-recommendations/002-profile-enrichment` plan）。本 baseline 不引入 backend-resume `GetResumeStructuredProfilesForUser` 或类似 cross-owner additive；frontend 切真后看到的 JobMatchProfile 字段稀疏度由 D-18 + D-19 锁定。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | baseline 数据源边界 | baseline 推荐 / 搜索只在**内部 jobs 池**（由 seed fixture + 未来 cross-owner data feed 提供）中执行；**不直连**外部招聘平台 API；`searchJobs` 通过 AI 在内部池中 rank，response 与 [B2 fixture](../openapi-v1-contract/spec.md) `searchJobs.json` `default` scenario 字段集一致 | 与 [product-scope §3.2 Q-2 + §7.3 P2](../product-scope/spec.md) 规划例外一致；不在 P0/P1 引入未合规数据源；外部接入归独立 P2 plan |
| D-2 | sources schema 边界 | `JobMatchProfileSourceCounts` 当前 schema 只含 `resumes / jds / mocks / debriefs` 4 项；experience_cards 计数 baseline 不进 sources schema（避免 B2 cross-owner additive 与 frontend fixture drift）；如未来 UI 需要展示 `experienceCards` 计数，由 cross-owner additive 在 B2 + 本 spec 同步修订 | 与 [B2 §3.1 D-X](../openapi-v1-contract/spec.md) JobMatchProfileSourceCounts schema 一致；baseline 不动 frontend fixture |
| D-3 | recommendation generation 触发 | 3 类触发：（1）`jd_match.agent_scan` 周期性后台 job（baseline P0 默认 4 小时一次，配置由 A4 提供）；（2）`addToWatchlist` / `markJobNotRelevant` 等 user feedback 后台触发增量 rescore；（3）frontend 进入 `jd_match` 路由时 polling 触发懒加载（如距上次 scan > N 分钟）；推荐结果存 `jd_match_recommendations` 表，`listJobRecommendations` 直接读 | 推荐生成与读取分离；frontend polling 不引起 AI 调用风暴；agent scan 调度策略由 backend 内部决定 |
| D-4 | search runtime 形态 | `searchJobs` 同步返回（fixture `searchJobs.json` `default` scenario 已固定 sync response，无 `progress` 字段）；backend 通过 AI 在内部 jobs 池中 rank，timeout 上限 backend 内部决定（baseline 30s）；超时返回 `502 + AI_PROVIDER_TIMEOUT`（与 [fixture `failed` scenario](../openapi-v1-contract/spec.md) 一致） | 与 [frontend D-12](../frontend-home-job-picks-and-parse/spec.md#3-用户决策--待确认事项) "frontend 不展示真实步骤进度切换" 一致；后端不暴露 progress 字段 |
| D-5 | side-effect IK 必带 | `addToWatchlist` / `removeFromWatchlist` / `markJobNotRelevant` / `searchJobs` / `createSavedSearch` 5 个 operation 必带 `Idempotency-Key`（[B2 §3.1.1](../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) 已声明）；handler 层 IK middleware 强制 | 防止网络抖动重复创建 / 重复扣除推荐池；与 B2 既有 IK 惯例一致 |
| D-6 | cross-user 隔离 | 所有 endpoint cross-user 访问返回 `404 + RESOURCE_NOT_FOUND`（不暴露存在）；audit_events 不写敏感字段；watchlist / saved_search / jd_match_recommendation 行的 `user_id` 必须强制过滤 | 与其他 backend domain 一致；防止枚举攻击 |
| D-7 | privacy 红线 | search query / watchlist label / saved-search label / sourceJobUrl / markNotRelevant freeNote / linkedJobMatchId 不进 logger 字符串、URL query、telemetry payload；只通过 generated client request body / store 行内 jsonb / text 列传递；fixture redact lint 必须覆盖；privacy delete tombstone 仅含 ID / 计数 / 时间 | 与 [frontend D-11](../frontend-home-job-picks-and-parse/spec.md#3-用户决策--待确认事项) + [product-scope §4.4](../product-scope/spec.md#44-伦理与安全约束) + [§9.3](../product-scope/spec.md#93-数据与隐私) 一致 |
| D-8 | watchlist / saved_searches 持久化 | 服务端持久化在 `watchlist_items` / `saved_searches` 表；不依赖前端 localStorage / sessionStorage / IndexedDB；watchlist 唯一性：`UNIQUE (user_id, linked_job_match_id)` 防止重复加入；saved_searches 无唯一性（用户可创建多个相同 query 的不同 label） | 与 [frontend D-9](../frontend-home-job-picks-and-parse/spec.md#3-用户决策--待确认事项) 一致；服务端兜底所有 jd_match 状态 |
| D-9 | AI provenance 强制 | `JobMatchRecommendation.provenance` 必填（[B2 schema](../openapi-v1-contract/spec.md) `GenerationProvenance` 已约束）；handler 必须从 `ai_task_runs` typed columns 投影 promptVersion / rubricVersion / modelId / language / featureFlag / dataSourceVersion；非 AI 生成的 watchlist / saved_search / market_signals 字段不带 provenance | 与 [B2 §4.6](../openapi-v1-contract/spec.md) + [product-scope §9.2](../product-scope/spec.md#92-prompt--rubric--model-治理) 一致 |
| D-10 | F3 / A3 cross-owner additive | 本 plan 携带 [F3 spec.md §3.1.1](../prompt-rubric-registry/spec.md#311-13-个当前-baseline-feature_key-字典) cross-owner additive：新增 2 个 baseline feature_key `jd_match.recommendation`（model profile `jd_match.recommendation.default`）+ `jd_match.search`（model profile `jd_match.search.default`）；F3 字典从 11 升至 13；新增 prompt / rubric / model profile 文件由本 plan 落地 baseline 文案 | F3 spec 同步修订；[A3 `config/ai-profiles.yaml`](../ai-provider-and-model-routing/spec.md) 必须新增对应 profile entry，并同步 [A3 §4.5 Product/UI AI Capability Catalog](../ai-provider-and-model-routing/spec.md#45-productui-ai-capability-catalog)，把 JD-Match 推荐解释 / 搜索从 `target.import.default` 升级为 `jd_match.recommendation.default` / `jd_match.search.default` |
| D-11 | B4 cross-owner additive | 本 plan 携带 [B4](../db-migrations-baseline/spec.md) cross-owner additive migration 创建 5 张新表（`jd_match_recommendations` / `watchlist_items` / `saved_searches` / `agent_scans` / `jd_match_search_runs`）+ 对应 index / FK / check constraint；B4 表总数从 28 升至 33；migration 必须通过 [B4 enum lint](../db-migrations-baseline/spec.md) / privacy dry-run / drift gate | 与 backend-resume/002 同款 cross-owner B4 additive 模式 |
| D-12 | B3 cross-owner additive | 本 plan 携带 [B3](../event-and-outbox-contract/spec.md) cross-owner additive：新增 2 个 internal event `jd_match.recommendation.completed` / `jd_match.search.completed` + 2 个 canonical job_type `jd_match_agent_scan`（dotted `jd_match.agent_scan`）+ `jd_match_search`（dotted `jd_match.search`）；envelope schema 必须遵守 PII 边界（不含 query / reasons / source_url） | B3 events / jobs total bump；通过 B3 generator + baseline gate |
| D-13 | agent scan 调度策略 | baseline P0 默认 4 小时一次；最近 scan 时间 + 下次 scan 时间存 `agent_scans` 表；frontend `getAgentScanStatus` 直接读最近行；scan 触发条件：（1）周期触发；（2）有新 source data（新 resume / target_job / debrief）通过 backend internal event 触发增量 scan；（3）用户进入 jd_match 路由懒触发（若距上次 scan > 4 小时）；不暴露 admin 调度 endpoint | 平衡推荐时新度与 AI 成本；调度由 backend 内部决定，前端只 polling |
| D-14 | searchJobs 与 jd_match_search_runs 关系 | 每次 `searchJobs` 调用创建 1 行 `jd_match_search_runs`（含 query / filters / result_count / provenance）；search 结果中的每个 JobMatchRecommendation 通过 join 到 `jd_match_recommendations` 表（同一 candidate 池中的行）；不为 search 结果创建 ephemeral recommendation 行；saved_search 的 `last_run_at` / `new_jobs_count` 由 search run 完成后异步更新 | 防止 search 调用污染 recommendation 池；search history 可审计 |
| D-15 | listJobRecommendations 默认排序 | 按 `score DESC, recommended_at DESC, id DESC` 唯一稳定序；cursor pagination；不返回已 dismissed 行（`dismissed_at IS NULL`）；不返回过期行（如有 `expires_at` 字段）；用户 mark not relevant 后该行 dismissed_at 写入，不出现在后续 list 调用 | 与 [B2 fixture](../openapi-v1-contract/spec.md) `listJobRecommendations.json` `default` scenario 一致 |
| D-16 | market_signals baseline 数据来源 | 4 个 market signal 由 backend 内部聚合：basaline P0 由 `watchlist_items` 表 + `jd_match_recommendations` 表 + 内部 jobs 池统计派生；不接入外部市场数据源；signal 字段集 `{k, v, d?, tone}` 与 [B2 fixture](../openapi-v1-contract/spec.md) `getMarketSignals.json` `default` scenario 一致 | 与 baseline 数据源边界 D-1 一致；外部市场数据接入归后续 P2 |
| D-17 | backend-auth cross-owner identity internal API | 新增 `backend-auth.GetUserIdentityForUser(ctx context.Context, userID string) (UserIdentity, error)` cross-owner internal API，`UserIdentity` 含 `{displayName, avatarUrl, emailMasked}`（read-only / 不写 audit / 不返回 raw email，仅返回 `emailMasked`）；签名锁定后由 [backend-auth/001](../backend-auth/plans/001-passwordless-session-bootstrap/plan.md) 在 in-place additive 中实现 `backend/internal/auth/service/identity.go`；本 plan Phase 0 携带 backend-auth spec/history additive 行；模式与 D-11 4 个 counter additive 一致 | 修复 [B2 `JobMatchProfile`](../openapi-v1-contract/spec.md) `required: [displayName, ...]` 与 [B2 `UserContext.displayName`](../openapi-v1-contract/spec.md) 同源约束在 plan §1 cross-owner aggregation 列表中缺位的漂移；P0 `displayName` 必须从 backend-auth 拿，避免 backend-jobs-recommendations 直接跨域 SQL JOIN `users` 表 |
| D-18 | JobMatchProfile P0 字段来源映射 + 稀疏 baseline | `JobMatchProfile` 字段来源在 P0 锁定如下：`displayName` ← `backend-auth.GetUserIdentityForUser(ctx, userID)`（必填，非 null；失败时 fallback 到非 PII anonymous display name）；`avatarUrl` ← P0 baseline `null`（P1 由 user 上传头像 / Gravatar 派生）；`headline` ← `backend-profile.GetCandidateProfileForUser(ctx, userID)`（缺失返回 null per D-13）；`yearsOfExperience` ← `backend-profile.GetCandidateProfileForUser(ctx, userID)`（缺失返回 null）；`locationText` ← P0 baseline `null`（P1 由 candidate_profile.region + remote 偏好格式化）；`compensationText` ← P0 baseline `null`（P1 由 target_jobs 期望薪资区间聚合）；`skills` ← P0 baseline `[]`（P1 由 candidate_profile + resume_versions.structured_profile 聚合去重，需要新 cross-owner additive，本 baseline 不引入）；`sources` ← 4 个 counter internal API 真实计数（必填 object，每字段 ≥ 0） | 防止 P0 baseline implementation 与 [B2 `JobMatchProfile required: [displayName, skills, sources]`](../openapi-v1-contract/spec.md) 漂移；防止 plan §1.1 把 `region / preferredPracticeLanguage` 错误列入响应字段；明确稀疏 baseline 不破坏 schema required 约束（`skills:[]` 满足 array required；optional fields 用 oneOf null 满足）；frontend graceful render 由 [B2 fixture `partial-profile` scenario](../openapi-v1-contract/spec.md) 与 [docs/ui-design/module-job-workspace.md](../../ui-design/module-job-workspace.md) 已覆盖 |
| D-19 | `getJobMatchProfile` fixture parity 例外 | mock-first 字节比对（C-13 / C-9 in [mock-contract-suite](../mock-contract-suite/spec.md#6-验收标准)）对 12 个 endpoint 中 11 个 endpoint 用 strict scenario byte parity with `default`；`getJobMatchProfile` 例外，因为 P0 baseline 稀疏字段（D-18）与 `default` scenario 富字段不一致，且 `partial-profile` scenario sources 全 0 也不匹配真实计数。`getJobMatchProfile` 用 **structural parity**：assertion 验证 (1) JobMatchProfile schema required 字段全部出现（displayName / skills / sources）；(2) optional 字段类型与 schema oneOf null 兼容（avatarUrl / headline / yearsOfExperience / locationText / compensationText 可为 null 或对应类型）；(3) `sources` 子对象 4 字段真实计数从 4 个 counter API 来；(4) headers / status code / X-Request-ID propagation 与 `default` scenario byte 一致 | 避免 B2 fixture `default` 二次编辑或新增 `baseline` scenario 引入 frontend 切真 contract churn；保留 `default` / `partial-profile` 作为 design preview；plan checklist 5.8 列出 11 byte parity + 1 structural parity 区分 |

### 3.2 待确认事项

| ID | 待确认事项 | 影响 | 默认处理 |
|----|------------|------|----------|
| Q-1 | agent scan 频率是否需要在 user setting 中可配置 | 影响 AI 成本与推荐时新度 | 默认 P0 固定 4 小时；A4 提供 `JD_MATCH_AGENT_SCAN_INTERVAL` 全局 config tuning；用户级配置归 P1 plan |
| Q-2 | internal jobs 池如何 seed 与扩展 | 影响 baseline 阶段推荐数据丰富度 | 默认 P0 由 backend-jobs-recommendations baseline plan 提供 seed fixture（mock 50-100 个 job posting）；未来 cross-owner data feed 扩展由独立 plan 设计 |
| Q-3 | searchJobs timeout / retry 策略 | 影响用户体验 | 默认 P0 baseline 30s timeout 同步返回 502；P1 改为 async job (`jd_match.search` job_type 已预占) 由用户主动 poll |
| Q-4 | watchlist `tone` 派生规则 | 影响 UI 视觉提示 | 默认 P0 由 backend 在 watchlist read 时按 recommendation `score` 派生（score ≥ 80 → ok / 50-79 → warn / < 50 → muted）；规则可由 P1 plan 调整 |
| Q-5 | listJobRecommendations pageSize 上限 | 影响一次 list 数据量 | 默认 P0 maximum=100（与 [B2 schema](../openapi-v1-contract/spec.md) 一致）；前端 default=20 |

## 4 设计约束

### 4.1 契约约束

- 实现 [B2 §3.1.1](../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) JobMatch tag 12 个 operationId 的 generated server interface；不允许私造 handler 签名。
- 11 个非 profile 响应的字段集 / status code / 字段顺序与 [B2 fixtures](../openapi-v1-contract/spec.md) `JobMatch/*.json` 字节比对；`getJobMatchProfile` 按 D-19 structural parity 验证；如响应字段在当前 B2 schema 未声明，先修订 B2 fixture，不在 backend 端绕过 fixture 自定义。
- 错误码必须 `$ref` [B1 D-5](../shared-conventions-codified/spec.md#31-已锁定决策) 已锁定的常量集；如需新增 jd_match 专有错误码，先修订 B1。
- side-effect operation（D-5 锁定）必带 IK；handler 层 + middleware 双层校验。
- response header `X-Request-ID` 必须传递；trace 字段通过 `Traceparent` parameter 支持。

### 4.2 AI 约束

- recommendation / search 必须通过 [A3 AIClient](../ai-provider-and-model-routing/spec.md) 调用；不允许业务代码 import 厂商 SDK / 直接 HTTP 调 model endpoint。
- prompt / rubric / 模型版本必须通过 [F3 registered feature_key](../prompt-rubric-registry/spec.md) 引用：`jd_match.recommendation`（model profile `jd_match.recommendation.default`）/ `jd_match.search`（model profile `jd_match.search.default`）；本 subject 不 hardcode prompt 正文。
- AI 输出必须含 `GenerationProvenance`（[B2 §4.6](../openapi-v1-contract/spec.md) + D-9 锁定）；写入 `ai_task_runs.model_profile_*` typed columns + `jd_match_recommendations` 行的 `prompt_version / rubric_version / model_id / data_source_version` 列。
- AI capability：recommendation / search 都使用 `chat` capability（[B1 D-8](../shared-conventions-codified/spec.md#31-已锁定决策)）；不引入 stt / realtime / judge / 向量检索（向量检索归后续 P2 retrieval plan）。
- AI 调用失败路径：timeout / output_invalid / retry exhausted 不写 `*.completed` event；只写 `ai_task_runs` + `async_jobs` retry metadata + audit failure 记录。

### 4.3 存储约束

- 5 张新表通过 [B4 cross-owner additive migration](../db-migrations-baseline/spec.md) 落地（D-11）；不绕过 store 层直接 SQL。
- 跨用户隔离：所有 read endpoint 必须以 `user_id = current_user_id` 过滤；cross-user 访问返回 404（不暴露存在）。
- 隐私删除调用 `DeleteJobMatchDataForUser(userId)`：删除顺序 `watchlist_items → saved_searches → jd_match_search_runs → jd_match_recommendations → agent_scans`（按 FK 依赖反序）；audit tombstone 仅写 userId / 各表删除计数 / 时间。
- query / label / sourceJobUrl / freeNote 等敏感字段不出现在 audit_events / outbox / log 中（D-7 + [B3 §3.1.4 PII 边界](../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory)）。
- jd_match_recommendations 行的 `recommendation_payload` jsonb（reasons / risks / highlights / network_note 等）只在 store 行内存储，不写入 audit / outbox payload。

### 4.4 cross-owner 接口约束

- 本 subject 消费的 cross-owner internal API（7 个）：
  - `backend-auth.GetUserIdentityForUser(ctx context.Context, userID string) (UserIdentity, error)`（本 plan 携带 backend-auth cross-owner additive；D-17 签名锁定；read-only / 不写 audit / 不返回 raw email；P0 baseline `avatarUrl` 来源为 null，仅 `displayName` / `emailMasked` 真实返回）
  - `backend-profile.GetCandidateProfileForUser(ctx context.Context, userID string) (*api.CandidateProfile, error)`（已锁定 [backend-profile D-13](../backend-profile/spec.md#31-已锁定决策)；read-only / 不触发 seed 副作用 / 缺失返回 nil）
  - `backend-profile.CountExperienceCardsBySource(ctx context.Context, userID string) (profile.SourceCounts, error)`（已锁定 [backend-profile D-11](../backend-profile/spec.md#31-已锁定决策)；read-only / 不暴露 card content）
  - `backend-resume.CountResumesForUser(ctx context.Context, userID string) (int, error)`（本 plan 携带 backend-resume cross-owner additive，签名锁定，由 backend-resume owner 同意）
  - `backend-targetjob.CountTargetJobsForUser(ctx context.Context, userID string) (int, error)`（本 plan 携带 backend-targetjob cross-owner additive）
  - `backend-practice.CountPracticeSessionsForUser(ctx context.Context, userID string) (int, error)`（本 plan 携带 backend-practice cross-owner additive）
  - `backend-debrief.CountDebriefsForUser(ctx context.Context, userID string) (int, error)`（本 plan 携带 backend-debrief cross-owner additive）
- 本 subject 暴露的 internal API：
  - `DeleteJobMatchDataForUser(userId)` 仅由 backend internal privacy runner 调用；不暴露 HTTP
- 所有 cross-owner internal API 在本 spec D-11 / D-12 / D-17 锁定；签名锁定后下游 owner 不得绕过 / 私造重复 API。

### 4.5 BDD / TDD 约束

- 每个 endpoint 必须有 handler unit test（参数校验 + IK + cross-user + 错误路径）+ `cmd/api` route wiring test（session middleware / IK middleware / path params）+ store integration test（state transition + cross-user isolation + cursor pagination）+ AI 调用 unit test（stub provider，验证 prompt/profile 路由正确）。
- 用户可见行为（getJobMatchProfile / list recommendations / mark not relevant / add to watchlist / search / save search）必须有 BDD scenario 覆盖；涉及 async job 的场景（agent scan）必须通过 backend-async-runner kernel（`runner.Runtime`）或等价真实 runtime harness 证明可执行，不得只验证包级 handler。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| 12 个 JobMatch HTTP handler | backend-jobs-recommendations | 真实业务逻辑 |
| `jd_match_recommendations` / `watchlist_items` / `saved_searches` / `agent_scans` / `jd_match_search_runs` 5 张表 schema | [B4 db-migrations-baseline](../db-migrations-baseline/spec.md) + 本 plan 携带 cross-owner additive migration | 字段 / 索引 / FK / check constraint |
| candidate 画像聚合 | backend-jobs-recommendations + cross-owner internal API（backend-auth / backend-profile / backend-resume / backend-targetjob / backend-practice / backend-debrief） | `BuildJobMatchProfile` 通过 7 个 internal API 聚合（D-17 / D-18 / D-19 锁定字段来源与稀疏 baseline）|
| `jd_match.agent_scan` job / `jd_match.recommendation` generator / `jd_match.search` sync handler | backend-jobs-recommendations + active [`backend-async-runner`](../backend-async-runner/spec.md) | `jd_match_agent_scan` job handler 注册到 backend-async-runner kernel（单一 `runner.Runtime`）；recommendation 不注册独立 job_type；search P0 仍走同步 HTTP handler 且为 future-async reserved，不注册后台 job |
| `cmd/api` runtime wiring | backend-jobs-recommendations + [backend-runtime-topology](../backend-runtime-topology/spec.md) | 挂载 12 个 JobMatch route 与 idempotency middleware；后台 job 生命周期统一由 backend-async-runner kernel 持有，不引入独立 worker 进程 |
| AI 调用 | [A3 AIClient](../ai-provider-and-model-routing/spec.md) + [F3 feature_key](../prompt-rubric-registry/spec.md) | backend-jobs-recommendations 只引用 profile，不绑定 provider；F3 cross-owner additive 新增 2 个 feature_key |
| 隐私删除调用 | backend internal privacy runner（[backend-runtime-topology](../backend-runtime-topology/spec.md)） | 调用 `DeleteJobMatchDataForUser` |
| frontend JD-Match UI | [frontend-home-job-picks-and-parse/002-jd-match-recommendations](../frontend-home-job-picks-and-parse/plans/002-jd-match-recommendations/plan.md) | 消费 generated TS client；当前通过 fixture-backed transport 闭环，本 subject 落地后切真 |
| mock-first fixtures | [B2 fixtures](../openapi-v1-contract/spec.md) | 11 个非 profile handler 响应字节比对；`getJobMatchProfile` 走 D-19 structural parity |
| company intel 深度数据 | 后续 plan / [backend-targetjob](../backend-targetjob/spec.md) | 本 subject 不实现 deep company intel；recommendation 中的 networkNote / interviewHypotheses 由 AI 生成 |
| 外部招聘平台 API 接入 | 后续 `backend-jobs-recommendations/002-external-sources` 或类似 P2 plan | baseline 不接入 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | getJobMatchProfile aggregation | 已登录用户 A + candidate_profile 行（headline / years 已填，per backend-profile D-1 seed）+ 3 resumes + 5 target_jobs + 8 practice_sessions + 2 debriefs | 调 `GET /api/v1/jd-match/profile` | 返回 200 + JobMatchProfile（D-18 锁定字段来源）：`displayName` 非 null（来自 backend-auth.GetUserIdentityForUser）；`avatarUrl=null`；`headline` 非 null（来自 backend-profile.GetCandidateProfileForUser）；`yearsOfExperience` 非 null；`locationText=null`；`compensationText=null`；`skills=[]`；`sources={resumes:3, jds:5, mocks:8, debriefs:2}` 真实计数；7 个 cross-owner internal API 各调 1 次；不直接 SQL JOIN 其他 owner 表；与 [B2 `JobMatchProfile` schema](../openapi-v1-contract/spec.md) required + oneOf null 字段口径完全一致；fixture parity 走 D-19 structural parity 路径，不走 `default` scenario byte parity | 001-jd-match-real-backend-baseline |
| C-2 | getAgentScanStatus | 用户 A 已登录 + `agent_scans` 表有最近 scan 行（status='idle', last_scan_at=2h ago, next_scan_at=2h later） | 调 `GET /api/v1/jd-match/agent-status` | 返回 200 + `AgentScanStatus{ status:'idle', lastScanAt, nextScanAt, message:null }`；与 [B2 fixture](../openapi-v1-contract/spec.md) `getAgentScanStatus.json` `default` scenario 一致；frontend polling 不触发新 scan | 001-jd-match-real-backend-baseline |
| C-3 | listJobRecommendations cursor pagination | 用户 A 有 25 个 active jd_match_recommendation 行（未 dismissed） | 调 `GET /api/v1/jd-match/recommendations?pageSize=20` 然后 cursor | 第一页返回 20 行 + `pageInfo.nextCursor` 非空 + 按 `score DESC, recommended_at DESC, id DESC` 唯一稳定序；第二页返回 5 行 + `hasMore=false`；已 dismissed 行不出现在 list；每行携带完整 `GenerationProvenance`；cross-user 不可见 | 001-jd-match-real-backend-baseline |
| C-4 | getJobRecommendation detail | 用户 A 有 jobMatchId X | 调 `GET /api/v1/jd-match/recommendations/{X}` | 返回 200 + JobMatchRecommendation 完整对象（包含 list 中可能省略的 networkNote / interviewHypotheses / similarInterviewers 详细字段）；与 [B2 fixture](../openapi-v1-contract/spec.md) `getJobRecommendation.json` `default` scenario 字段集一致；cross-user 返回 404 + RESOURCE_NOT_FOUND | 001-jd-match-real-backend-baseline |
| C-5 | markJobNotRelevant | 用户 A + jobMatchId X (active) + IK | 调 `POST /api/v1/jd-match/recommendations/{X}/dismiss` body `MarkNotRelevantRequest{reason:'wrong_level', freeNote:'...'}` + IK | 返回 200 + `MarkNotRelevantResult{ jobMatchId:X, dismissedAt }`；DB 行 `dismissed_at` 写入；后续 list 不包含 X；IK replay 返回首次结果；freeNote 不进 log / audit / outbox | 001-jd-match-real-backend-baseline |
| C-6 | addToWatchlist + UNIQUE | 用户 A + jobMatchId X + IK | 调 `POST /api/v1/jd-match/watchlist` body `AddToWatchlistRequest{jobMatchId:X}` + IK | 返回 200 + WatchlistItem；DB `watchlist_items` 新增行 `linked_job_match_id=X, user_id=user_A`；同 IK replay 返回首次 item；UNIQUE 约束防止重复加入（不同 IK 重复 add 返回首次 item 而非创建新行） | 001-jd-match-real-backend-baseline |
| C-7 | listWatchlist + tone | 用户 A 已 add 3 watchlist items（关联 score 92 / 78 / 45 三个 recommendation） | 调 `GET /api/v1/jd-match/watchlist` | 返回 200 + 3 items；tone 按 score 派生（92→ok / 78→warn / 45→muted，按 Q-4 默认规则）；与 [B2 fixture](../openapi-v1-contract/spec.md) `listWatchlist.json` `default` scenario 字段集一致 | 001-jd-match-real-backend-baseline |
| C-8 | removeFromWatchlist | 用户 A + jobMatchId X 已在 watchlist + IK | 调 `DELETE /api/v1/jd-match/watchlist/{X}` + IK | 返回 204；DB watchlist 行删除；IK replay 同样返回 204；cross-user 删除返回 404 | 001-jd-match-real-backend-baseline |
| C-9 | searchJobs sync return | 用户 A + IK + 内部 jobs 池有可匹配 JD | 调 `POST /api/v1/jd-match/search` body `SearchJobsRequest{ query:'frontend platform' }` + IK | 返回 200 + `SearchJobsResponse{ searchRunId, items:[JobMatchRecommendation...] }`（同步返回，无 progress 字段）；DB `jd_match_search_runs` 新增 1 行（包含 query / filters / result_count / provenance）；search runtime ≤ 30s；超时返回 502 + AI_PROVIDER_TIMEOUT；query 不进 log / audit / outbox payload；IK replay 返回首次 searchRunId | 001-jd-match-real-backend-baseline |
| C-10 | listSavedSearches + createSavedSearch | 用户 A + IK | 调 `POST /api/v1/jd-match/saved-searches` body `CreateSavedSearchRequest{ label:'frontend remote', query, filters }` + IK；后调 `GET /api/v1/jd-match/saved-searches` | 返回 200 + SavedSearch（首次创建）；DB `saved_searches` 新增；后续 GET 返回包含该行；label / query 不进 log；newJobsCount / lastRunAt 字段在用户尚未 run 时为 null | 001-jd-match-real-backend-baseline |
| C-11 | getMarketSignals 4-card | 用户 A 已登录 + 内部 jobs 池统计已就位 | 调 `GET /api/v1/jd-match/market-signals?window=7d` | 返回 200 + `MarketSignalsResponse{ signals:[{k,v,d?,tone}×4], asOf }`；4 个 signal 由 backend 内部聚合（不接入外部市场数据）；与 [B2 fixture](../openapi-v1-contract/spec.md) `getMarketSignals.json` `default` scenario 字段集一致；window 参数默认 7d，可选 7d / 14d / 30d | 001-jd-match-real-backend-baseline |
| C-12 | agent_scan recommendation generation | `jd_match.agent_scan` job 触发（周期 / on-demand / 用户 lazy trigger） | 通过 [A3 AIClient](../ai-provider-and-model-routing/spec.md) 调 F3 `jd_match.recommendation` feature_key + 内部 jobs 池 | DB `agent_scans` 行 `status='scanning'` → `status='idle' + last_scan_at + next_scan_at`；`jd_match_recommendations` 新增/更新行（含 score / reasons / risks / highlights / provenance）；`ai_task_runs` 行写入 typed columns；outbox `jd_match.recommendation.completed` event 发射；失败路径（AI timeout / output_invalid）不发 completed event；privacy：reasons / risks / source_url 不进 outbox payload；recommendation generator 不注册独立 canonical job_type | 001-jd-match-real-backend-baseline |
| C-13 | mock-first 字节比对（11 + 1 例外）| B2 fixtures `JobMatch/*.json` 12 个 default scenario + D-19 锁定的 `getJobMatchProfile` structural parity 路径 | 通过 `cmd/api` route 调真实 handler | 11 个 endpoint（`getAgentScanStatus` / `listJobRecommendations` / `getJobRecommendation` / `markJobNotRelevant` / `listWatchlist` / `addToWatchlist` / `removeFromWatchlist` / `searchJobs` / `listSavedSearches` / `createSavedSearch` / `getMarketSignals`）响应字段集 / status / header 字节一致；session / IK middleware 不改变 generated response envelope；`getJobMatchProfile` 走 D-19 structural parity（required 字段满足 + optional null 兼容 + sources 真实计数 + status 200 + X-Request-ID propagation）| 001-jd-match-real-backend-baseline |
| C-14 | cross-user 隔离全面 | 用户 A 有 jd_match data；用户 B 调任意 read / write endpoint | – | 全部返回 404 + RESOURCE_NOT_FOUND（不暴露存在）；audit_events 不写敏感字段；cross-user 行不被读 / 写 | 001-jd-match-real-backend-baseline |
| C-15 | privacy 删除链路 | 用户 A 有 10 recommendations + 3 watchlist + 2 saved_searches + 5 search_runs + 4 agent_scans | privacy_delete job 触发，调 `DeleteJobMatchDataForUser(userId)` | 删除顺序：watchlist_items → saved_searches → jd_match_search_runs → jd_match_recommendations → agent_scans；audit tombstone 写 userId / 各表删除计数 / 时间；不含 query / label / reasons / source_url / freeNote；DB 状态 user_A 全部 0 行；其他用户不受影响 | 001-jd-match-real-backend-baseline |
| C-16 | 旧口径负向断言 | grep `mistake|growth|drill|experiences|star|LinkedIn|Boss|脉脉|拉勾` in `backend/internal/jdmatch/` + outbox payload + DB seed data | – | 0 命中（与 [product-scope §3.2 Q-2](../product-scope/spec.md#32-待确认事项) + [§7.3 P2](../product-scope/spec.md#73-p2工程化和数据源扩展) 外部平台不接入一致；与 [engineering-roadmap D-6](../engineering-roadmap/spec.md#31-已锁定决策) 旧模块不恢复一致）；platform 字面量若出现在 schema 文档/注释作为 "未来扩展候选" 描述允许，但不允许出现在 code 实际 HTTP client / SDK / DB seed | 001-jd-match-real-backend-baseline |
| C-17 | AI provenance 完整性 | 任一 recommendation / search result | 返回 response | 每个 JobMatchRecommendation 必须含 `provenance{promptVersion, rubricVersion, modelId, language, featureFlag, dataSourceVersion}`；缺字段返回 500（handler 层断言）；watchlist / saved_search / market_signals 字段不带 provenance | 001-jd-match-real-backend-baseline |
| C-18 | F3 + A3 + B4 + B3 cross-owner additive PASS | 本 plan 携带的 F3 / A3 / B4 / B3 修订 | 运行各自 gate | F3 字典 11 → 13；A3 `config/ai-profiles.yaml` 与 §4.5 Product/UI catalog 均含 `jd_match.recommendation.default` / `jd_match.search.default`，且 provider ref 使用 `deepseek`、模型 ID 使用 `deepseek-v4-flash` / `deepseek-v4-pro`；B4 表 28 → 33 + migration / privacy dry-run / drift PASS；B3 events / jobs 新增 2 event + 2 job_type baseline gate PASS；所有 cross-owner additive 通过对应 owner spec / history / INDEX 修订同步 | 001-jd-match-real-backend-baseline |
| C-19 | cross-owner identity + counter additive PASS | 本 plan 携带的 backend-auth identity + backend-resume/targetjob/practice/debrief counter 共 5 个 cross-owner internal API additive | 各 owner unit test + spec/history 同步 | `backend-auth.GetUserIdentityForUser(ctx context.Context, userID string) (UserIdentity, error)` 返回 `{displayName, avatarUrl, emailMasked}` 字段完整；read-only / 不写 audit_events / 不返回 raw email（emailMasked 形如 `ali***@example.com`）；cross-user 由 caller 提供 userId 保证；4 个 counter API 均使用 `Count*ForUser(ctx context.Context, userID string) (int, error)` 并各自返回 ≥0 整数；所有 5 个 owner spec.md 模块边界表新增对应 internal API 行 + history.md 追加 backend-jobs-recommendations/001 cross-owner additive 行；`sync-doc-index --check` PASS | 001-jd-match-real-backend-baseline |

## 7 关联计划

- [001-jd-match-real-backend-baseline](./plans/001-jd-match-real-backend-baseline/plan.md)：第一批 plan，落地 12 个 JobMatch endpoint real backend + 5 张新表 cross-owner B4 additive + 2 个 feature_key cross-owner F3 additive + 2 个 event / 2 个 job_type cross-owner B3 additive + AI 编排 baseline + agent scan 后台 job + privacy delete + 7 个 cross-owner internal API 整合（backend-auth identity additive + backend-profile 2 个 read API + 4 个 counter additive）；BDD 覆盖 profile aggregation → recommendation generation → search → watchlist → privacy 主路径。
- `002-external-sources-and-scaling`（P2 延后）：落地外部招聘平台 API 接入（评估合规与可维护性后选择 1-2 个数据源）+ vector search / pgvector 检索 + 离线评估集 ≥ 50 题 + 多语言 personalized rank + agent scan user-level config。
- `003-company-intel-and-network`（P1/P2 延后）：落地深度 company intel（招聘历史 / 新闻 / 面试题库）+ network note 真实数据源（mock interview alumni overlap）+ similarInterviewers 精确推断。

## 8 关联文档

- 上游 spec：[`product-scope`](../product-scope/spec.md)、[`engineering-roadmap`](../engineering-roadmap/spec.md)、[`openapi-v1-contract`](../openapi-v1-contract/spec.md)、[`db-migrations-baseline`](../db-migrations-baseline/spec.md)、[`shared-conventions-codified`](../shared-conventions-codified/spec.md)、[`event-and-outbox-contract`](../event-and-outbox-contract/spec.md)、[`ai-provider-and-model-routing`](../ai-provider-and-model-routing/spec.md)、[`prompt-rubric-registry`](../prompt-rubric-registry/spec.md)、[`backend-runtime-topology`](../backend-runtime-topology/spec.md)、[`mock-contract-suite`](../mock-contract-suite/spec.md)
- 上游 owner 内部 API 依赖：[`backend-auth`](../backend-auth/spec.md)（D-17 identity additive）、[`backend-profile`](../backend-profile/spec.md)（D-11 + D-13）、[`backend-resume`](../backend-resume/spec.md)、[`backend-targetjob`](../backend-targetjob/spec.md)、[`backend-practice`](../backend-practice/spec.md)、[`backend-debrief`](../backend-debrief/spec.md)
- 下游 frontend：[`frontend-home-job-picks-and-parse`](../frontend-home-job-picks-and-parse/spec.md)（plan 002 承接前端删除）
- UI 真理源：[`docs/ui-design/removed-modules-and-scope.md`](../../ui-design/removed-modules-and-scope.md) §15（jd_match 删除记录；旧 `ui-design/src/screen-jd-match.jsx` 已随 2026-06-12 第二批裁剪删除）
- 历史：[history.md](./history.md)

## 9 D-17 删除范围与零残留验收（当前 active scope）

### 9.1 删除范围

| 层 | 待删资产 | 处置 |
|----|---------|------|
| OpenAPI | `openapi/openapi.yaml` jobmatch tag 12 个 operationId（`getJobMatchProfile` / `getAgentScanStatus` / `listJobRecommendations` / `getJobRecommendation` / `markJobNotRelevant` / `listWatchlist` / `addToWatchlist` / `removeFromWatchlist` / `searchJobs` / `listSavedSearches` / `createSavedSearch` / `getMarketSignals`）及其专属 schema、`openapi/fixtures/JobMatch/` 全部 fixture | 删除并重新 codegen；[openapi-v1-contract](../openapi-v1-contract/spec.md) freeze 列表同步修订 |
| backend | `backend/internal/jdmatch/`（handler / service / store / jobs / generators）、`backend/cmd/api/jdmatch_runtime.go` 与 3 个 jdmatch 测试文件、`main.go` 挂载点、session policy / generated server 中 jobmatch 入口 | 删除文件 + 原地修改共享文件 |
| cross-owner internal API | `backend-auth.GetUserIdentityForUser`、`backend-{resume,targetjob,practice,debrief}.Count*ForUser` 等为 JobMatchProfile 聚合新增的 additive | 仅当无其他消费方时随删；有消费方（如 privacy 数据概览 / profile 页）则保留并在 plan 中记录留存理由 |
| migrations | `jd_match_recommendations` / `watchlist_items` / `saved_searches` / `agent_scans` / `jd_match_search_runs` 5 张表与 000010 注入的 `jd_match.*` prompt/rubric registry 行 | 新增 drop migration 收口（000009/000010 保留为历史迁移文件）；`migrations/enum-sources.yaml` 同步 |
| shared / B3 | `jd_match.recommendation.completed` / `jd_match.search.completed` 事件、`jd_match.agent_scan` job_type、events/jobs baseline 与 schema 文件、生成常量 | 删除并重新 codegen |
| config / F3 | `jd_match.recommendation` / `jd_match.search` feature_key、`config/prompts|rubrics|evals/jd_match.*`、`config/ai-profiles.yaml` 关联条目、`resolved-prompts.json` 再生成 | 删除 |
| scenarios | `test/scenarios/e2e/p0-094..097-jd-match-*` 4 个真实后端场景目录与 INDEX 行 | 删除（前端 mock 场景 p0-017 / p0-027..031 归 frontend plan 002） |

### 9.2 验收标准（v2.0）

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-R1 | 契约零残留 | D-17 删除完成 | 运行 `make codegen && make codegen-check`、fixtures 校验、`rg -i "jobmatch|jd[-_]match"` 于 `openapi/ backend/ shared/ config/ migrations/`（drop migration、历史 000009/000010 迁移文件与负向断言除外） | jobmatch tag、生成物、fixtures、事件、job_type、feature_key、prompt/rubric/eval 资产零残留；backend 全量 `go test ./...` 通过 | 001 Phase 9 |
| C-R2 | 运行时无 jd_match 表面 | 删除后启动 `cmd/api` | 请求任一旧 `/api/v1/jd-match/*` 路径 | 返回 404（非 500 / 非鉴权拦截后的 panic）；session policy 不再包含 jobmatch operation | 001 Phase 9 |
| C-R3 | 数据删除收口 | 既有 dev DB 含 jd_match 数据 | 应用 drop migration；运行 privacy delete | 5 张表与 registry 行被删除；privacy delete 链路不再引用 jd_match 表且测试通过 | 001 Phase 9 |

## 10 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 2.0 | 2026-06-13 | 对齐 product-scope v2.1 D-17：subject 转为岗位推荐模块删除的 backend / 契约侧 owner；新增 §9 删除范围与 C-R1~C-R3 零残留验收；§1-§8 标注为历史 baseline 记录 |
