# 001 BDD Checklist

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-05-22

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

> 2026-05-22 post-reopen completion：E2E.P0.094-097 的 scenario assets、脚本断言、执行日志和 INDEX Ready/automated 状态已由真实证据闭环；最终 cleanup 未执行是为了保留 `.test-output` trigger logs，cleanup 脚本仍保持可用。

> 2026-05-22 L2 follow-up：E2E.P0.097 的 live cmd/api 断言已追加 agent scan AI payload context 检查；privacy runner domain delete 集成由 focused runner 单测和最终 wrapper gate 覆盖。

## E2E.P0.094 jd-match-profile-and-recommendations-list

- [x] 创建场景目录 `test/scenarios/e2e/p0-094-jd-match-profile-and-recommendations-list/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md` + scripts。
- [x] 准备 B2 fixture scenario + 测试数据：`TestJDMatchFixtureParity` 覆盖 `getJobMatchProfile` D-19 structural parity、`getAgentScanStatus`、`listJobRecommendations`、`getJobRecommendation`、`markJobNotRelevant`；`TestJDMatchHTTPScenario` 覆盖 session / IK / cross-user / privacy delete live path；缺少 `DATABASE_URL` 或出现 skip/no-test marker 时 verify fail。
- [x] 实现 `scripts/setup.sh` / `scripts/trigger.sh` / `scripts/verify.sh` / `scripts/cleanup.sh`；trigger 运行 `TestJDMatchFixtureParity|TestJDMatchHTTPScenario`，verify 断言目标 PASS marker、package-level `ok github.com/monshunter/easyinterview/backend/cmd/api`，并拒绝 skip/no-test/raw email/`--- FAIL`/package `FAIL`。
- [x] 执行 `setup → trigger → verify` 全 PASS（最终 cleanup 未执行以保留 `.test-output` 证据；cleanup 脚本保持可用）。
- [x] 记录验证证据：`.test-output/e2e/p0-094-jd-match-profile-and-recommendations-list/trigger.log` 含 profile structural parity、agent status、recommendations list/detail、dismiss、HTTP scenario PASS。
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表 P0.094 行标记 Ready / automated。

## E2E.P0.095 jd-match-watchlist-and-saved-search-lifecycle

- [x] 创建场景目录 `test/scenarios/e2e/p0-095-jd-match-watchlist-and-saved-search-lifecycle/`，含 README / data / scripts。
- [x] 准备 B2 fixture scenario + 测试数据：`TestJDMatchFixtureParity` 覆盖 `listWatchlist` / `addToWatchlist` / `removeFromWatchlist` / `listSavedSearches` / `createSavedSearch` default fixture；`TestJDMatchHTTPScenario` 覆盖 IK replay、UNIQUE 非重复、tone、cross-user 404。
- [x] 实现 `scripts/setup.sh` / `scripts/trigger.sh` / `scripts/verify.sh` / `scripts/cleanup.sh`；verify 断言目标 PASS marker、package-level `ok github.com/monshunter/easyinterview/backend/cmd/api`，并拒绝 skip/no-test/raw email/`--- FAIL`/package `FAIL`。
- [x] 执行 `setup → trigger → verify` 全 PASS（保留 `.test-output` 证据）。
- [x] 记录验证证据：`.test-output/e2e/p0-095-jd-match-watchlist-and-saved-search-lifecycle/trigger.log`。
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表 P0.095 行标记 Ready / automated。

## E2E.P0.096 jd-match-search-and-market-signals

- [x] 创建场景目录 `test/scenarios/e2e/p0-096-jd-match-search-and-market-signals/`，含 README / data / scripts。
- [x] 准备 B2 fixture scenario + 测试数据：`TestJDMatchFixtureParity` 覆盖 `searchJobs` + `getMarketSignals` default fixture；`TestJDMatchHTTPScenario` 覆盖 search IK replay、`jd_match_search_runs` 非重复、`jd_match.search.completed` outbox privacy、AI_OUTPUT_INVALID 不新增 search_run、market-signals 4 signals。
- [x] 实现 `scripts/setup.sh` / `scripts/trigger.sh` / `scripts/verify.sh` / `scripts/cleanup.sh`；verify 断言目标 PASS marker、package-level `ok github.com/monshunter/easyinterview/backend/cmd/api`，并拒绝 skip/no-test/raw email/`--- FAIL`/package `FAIL`。
- [x] 执行 `setup → trigger → verify` 全 PASS（保留 `.test-output` 证据）。
- [x] 记录验证证据：`.test-output/e2e/p0-096-jd-match-search-and-market-signals/trigger.log`。
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表 P0.096 行标记 Ready / automated。

## E2E.P0.097 jd-match-agent-scan-and-privacy-delete

- [x] 创建场景目录 `test/scenarios/e2e/p0-097-jd-match-agent-scan-and-privacy-delete/`，含 README / data / scripts。
- [x] 准备 fixture + stub provider：`TestJDMatchAgentScanDrainerScenario` 覆盖 `jd_match_agent_scan` drainer claim/finalize、recommendation upsert、completed outbox、shutdown；`TestJDMatchA3F3AdapterUsesRegistryProfilesForSearchAndRecommendation` 覆盖 production adapter 使用 `jd_match.recommendation.default` / `jd_match.search.default`；`TestJDMatchHTTPScenario` 覆盖 privacy delete live path。
- [x] L2 follow-up：`TestJDMatchAgentScanDrainerScenario` 断言 generator payload `candidateProfile` 含 runtime identity/profile 聚合结果，`jobsPool` 含 seeded `jobMatchId`；`TestPrivacyDeleteHandlerDeletesUploadFilesForRequestUser` 断言同一 privacy job 同时调用 upload/profile/JDMatch domain deleter。
- [x] 实现 `scripts/setup.sh` / `scripts/trigger.sh` / `scripts/verify.sh` / `scripts/cleanup.sh`；trigger 运行 `TestJDMatchHTTPScenario|TestJDMatchAgentScanDrainerScenario`，verify 断言目标 PASS marker、package-level `ok github.com/monshunter/easyinterview/backend/cmd/api`，并拒绝 skip/no-test/raw email/`--- FAIL`/package `FAIL`。
- [x] 执行 `setup → trigger → verify` 全 PASS（保留 `.test-output` 证据）。
- [x] 记录验证证据：`.test-output/e2e/p0-097-jd-match-agent-scan-and-privacy-delete/trigger.log`。
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表 P0.097 行标记 Ready / automated。
