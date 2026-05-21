# 001 BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-21

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.094 jd-match-profile-and-recommendations-list

- [ ] 创建场景目录 `test/scenarios/e2e/p0-094-jd-match-profile-and-recommendations-list/`，含 `README.md`（背景 + baseline + 离线限制 + spec D-18 / D-19 稀疏 baseline 与 structural parity 例外说明）+ `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 B2 fixture scenario + 测试数据：`getJobMatchProfile.json` `default`（用于 X-Request-ID / status / headers 取样）+ `partial-profile`（用于 reference 稀疏字段 shape）、`getAgentScanStatus.json` `default`、`listJobRecommendations.json` `default`、`getJobRecommendation.json` `default`、`markJobNotRelevant.json` `default`；2 个测试用户（A / B）；用户 A users 表已 seed 含 displayName + email；用户 A 已通过 store 直接写入 candidate_profile (headline / yearsOfExperience 真实) + 3 resumes + 5 target_jobs + 8 practice_sessions + 2 debriefs + 25 个 active jd_match_recommendation（混合 score 分布以验证排序）+ agent_scans 行；缺少 live env、integration-tag 测试 skip 或 focused gate no-op 时本场景必须 fail
- [ ] 实现 `scripts/setup.sh`（A2 dev stack + `cmd/api` jdmatch route + session middleware + IK middleware + 5 个 cross-owner additive owner runtime 注入 + 用户登录 + 批量 store-write 测试数据）/ `scripts/trigger.sh`（依序触发 A1-A7 + A8 7 个 cross-owner API trace + B1-B3，并运行 `cd backend && go test ./cmd/api -run TestJDMatchHTTPScenario -count=1 -v` + `cd backend && go test ./internal/jdmatch/service -run TestBuildJobMatchProfileAggregation -count=1 -v`）/ `scripts/verify.sh`（断言 `method=cmd-api-http+internal-api` 或等价 live runtime evidence + no-op / skip 不可 PASS + 7 个 cross-owner API 调用证据 + identity emailMasked 格式 + identity 不写 audit_events + GetCandidateProfileForUser 不 seed 副作用 + JobMatchProfile structural parity (required 字段 / optional null oneOf 兼容 / sources 真实计数) + headers byte 一致 + sources 计数 + recommendation 排序 + provenance 完整 + dismiss 后列表更新 + cross-user 404 + 4 个 fixture 字节比对 + validation 直接断言 + 隐私 grep（含 raw email 负向）+ 旧口径 grep）/ `scripts/cleanup.sh`（清理 5 张 jd_match 表 + 跨域数据 + 用户）
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-094-jd-match-profile-and-recommendations-list/trigger.log` + `cmd/api` HTTP scenario log + BuildJobMatchProfile service log + verify 输出 + 7 个 cross-owner API 调用 trace（含 identity API + 2 个 backend-profile read + 4 个 counter）+ identity API emailMasked 格式与 audit 0 写入证据 + getJobMatchProfile structural parity 断言输出（required / optional null / sources 真实计数）+ recommendation list 4 个 fixture byte diff 0 + provenance 字段完整性 + dismissed 行不在 list 验证 + cross-user 404 + raw email 负向 grep 0 命中 + `method=cmd-api-http+internal-api` 或等价 live runtime evidence + no no-op / no skip 证据
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.094 行（关联需求 `backend-jobs-recommendations C-1, C-2, C-3, C-4, C-5, C-13, C-14, C-17, C-19`，状态 Ready，automated）

## E2E.P0.095 jd-match-watchlist-and-saved-search-lifecycle

- [ ] 创建场景目录 `test/scenarios/e2e/p0-095-jd-match-watchlist-and-saved-search-lifecycle/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 B2 fixture scenario + 测试数据：`listWatchlist.json` / `addToWatchlist.json` / `removeFromWatchlist.json` / `listSavedSearches.json` / `createSavedSearch.json` `default`；2 个测试用户（A / B）；用户 A 已有 3 个 jd_match_recommendation (score 92 / 78 / 45)；用户 B 无 jd_match data
- [ ] 实现 `scripts/setup.sh`（A2 dev stack + `cmd/api` jdmatch route + 用户登录 + store-write recommendation）/ `scripts/trigger.sh`（依序触发 A1-A11 + B1，并运行 `cd backend && go test ./cmd/api -run TestJDMatchHTTPScenario -count=1 -v`）/ `scripts/verify.sh`（断言 `method=cmd-api-http` + no-op / skip 不可 PASS + DB watchlist UNIQUE 约束 + IK replay vs UNIQUE 兜底 + tone 派生（score→tone）+ saved_searches CRUD + cross-user 404 + label / query 不进 log + 5 个 fixture 字节比对 + 隐私 grep + 旧口径 grep）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-095-jd-match-watchlist-and-saved-search-lifecycle/trigger.log` + `cmd/api` HTTP scenario log + verify 输出 + DB watchlist UNIQUE 验证 + IK replay 证据 + tone 派生 trace + saved_searches 5 fixture byte diff 0 + cross-user 404 验证 + label/query 隐私 grep 0 命中 + `method=cmd-api-http` + no no-op / no skip 证据
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.095 行（关联需求 `backend-jobs-recommendations C-6, C-7, C-8, C-10, C-13, C-14, C-17`，状态 Ready，automated）

## E2E.P0.096 jd-match-search-and-market-signals

- [ ] 创建场景目录 `test/scenarios/e2e/p0-096-jd-match-search-and-market-signals/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 B2 fixture scenario + 测试数据：`searchJobs.json` `default` + `failed`、`getMarketSignals.json` `default`；A3 AIClient stub provider 3 variant（success / timeout / output_invalid）；F3 `jd_match.search` feature_key prompt 已就位；内部 jobs 池 seed 50 个 mock JD；用户 A 已有 5 个 jd_match_recommendation 用于 market_signals 派生
- [ ] 实现 `scripts/setup.sh`（A2 dev stack + `cmd/api` runtime + stub provider variant 注入 + 用户登录 + 内部 jobs 池 seed + store-write recommendations）/ `scripts/trigger.sh`（依序触发 A1-A7，并运行 `cd backend && go test ./cmd/api -run TestJDMatchHTTPScenario -count=1 -v`）/ `scripts/verify.sh`（断言 `method=cmd-api-http` + no-op / skip 不可 PASS + searchJobs 同步返回 + searchRunId + 30s timeout 502 + output_invalid 502 + IK replay + DB search_runs 行 + market_signals 4 个 signal 派生 + window enum + fixture 字节比对 + query 不进 log/outbox + 隐私 grep + 外部平台 grep 0 命中）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-096-jd-match-search-and-market-signals/trigger.log` + `cmd/api` HTTP scenario log + verify 输出 + searchJobs success / timeout / output_invalid 三 variant 行为证据 + search_runs DB dump + market_signals 4 signal 派生 trace + 2 fixture byte diff 0 + query / filters 隐私 grep 0 命中 + LinkedIn/Boss/脉脉/拉勾 grep 0 命中 + `method=cmd-api-http` + no no-op / no skip 证据
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.096 行（关联需求 `backend-jobs-recommendations C-9, C-11, C-13, C-14, C-17`，状态 Ready，automated）

## E2E.P0.097 jd-match-agent-scan-and-privacy-delete

- [ ] 创建场景目录 `test/scenarios/e2e/p0-097-jd-match-agent-scan-and-privacy-delete/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture + stub provider：A3 AIClient stub provider 2 variant（success / output_invalid）；`cmd/api` in-process drainer 启动并注册 `jd_match_agent_scan` job handler；F3 `jd_match.recommendation` feature_key prompt / rubric / profile 就位；用户 A 已登录 + 有 candidate_profile + 3 resumes + 10 jd_match_recommendation + 3 watchlist + 2 saved_searches + 5 search_runs + 4 agent_scans；用户 B / C 不受影响；privacy_delete job 模拟入口已就位；缺少 live env、drainer no-op 或 focused gate skip 时本场景必须 fail
- [ ] 实现 `scripts/setup.sh`（`cmd/api` runtime + in-process drainer + stub provider variant 注入 + 用户登录 + 批量 store-write user_A 真实内容）/ `scripts/trigger.sh`（分子场景 A drainer RunOnce success + B drainer output_invalid + C 调 internal DeleteJobMatchDataForUser + D 模拟事务回滚 + E shutdown；运行 `cd backend && go test ./cmd/api -run TestJDMatchAgentScanDrainerScenario -count=1 -v` + `cd backend && go test ./internal/jdmatch/service -run TestPrivacyDeleteOrderAndAudit -count=1 -v`）/ `scripts/verify.sh`（断言 `method=cmd-api+internal-api` + no-op / skip 不可 PASS + agent_scans 状态机轨迹 + ai_task_runs 多行 + ready-only outbox event 写入 + failure no completed event + 5 表删除顺序 + audit tombstone 完整无敏感字段 + DB 状态 user_A 全 0 / 其他用户不受影响 + 失败回滚 + shutdown 无 goroutine leak + 隐私 grep + 外部平台 grep + 旧口径 grep）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-097-jd-match-agent-scan-and-privacy-delete/trigger.log` + `cmd/api` drainer scenario log + privacy service log + verify 输出 + DB agent_scans 状态转换轨迹 + ai_task_runs 行 dump + outbox_events 行 dump（ready-only completed event，failure no completed event，PII grep 0 命中）+ stub provider call log + 5 表删除顺序证据 + audit tombstone payload dump（验证无敏感字段）+ 失败回滚证据 + cross-user 不影响验证 + shutdown / no goroutine leak 证据 + `method=cmd-api+internal-api` + no no-op / no skip 证据
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.097 行（关联需求 `backend-jobs-recommendations C-12, C-15, C-16, C-18`，状态 Ready，automated）
