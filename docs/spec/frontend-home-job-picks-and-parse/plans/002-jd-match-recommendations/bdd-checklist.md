# 002 BDD Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-10

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.022 jd_match Recommended tab 主路径 + Save/Unsave/Mark not relevant/Open source 闭环 + 4xx revert + auth pending action

- [ ] 创建场景目录 `test/scenarios/e2e/p0-022-jd-match-recommended-and-confirm/`，含 `README.md`（§6 baseline + §7 离线限制）
- [x] 准备 fixture variant：`getJobMatchProfile.json`（`default`）、`getAgentScanStatus.json`（`idle/scanning/error`）、`listJobRecommendations.json`（`empty/one/many/failed`）、`getJobRecommendation.json`（`default/network-intel-empty/failed`）、`addToWatchlist.json`（`default/4xx-validation/5xx`）、`removeFromWatchlist.json`（`default/4xx-not-found`）、`markJobNotRelevant.json`（`default/4xx`）；signed-in / signed-out 两种状态切换入口；`make validate-fixtures` PASS <!-- evidence: 2026-05-10 plan §1.2 fixtures landed, validate-fixtures OK 46 fixtures; signed-in/out via existing Auth/getMe authenticated/unauthenticated fixtures -->
- [ ] 实现 `scripts/setup.sh`（含 fixture variant 切换 + signed-in/out 切换 + 三主题态预置入口）/ `scripts/trigger.sh`（按 A-I 9 个子用例运行）/ `scripts/verify.sh`（断言 9 子用例 Then 子句全部命中、`getJobRecommendation` 详情必调、4 button request body / Idempotency-Key、4xx revert 行为、auth pending action params 隐私字段、Source 不触发 pendingAction 且 `window.open` flags 命中、mockTransport spy 仅 status code、jobMatchId/sourceUrl/freeNote 隐私反查、plan 001 placeholder testid 0 命中）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS（A-I 共 ≥9 子用例）
- [ ] 记录验证证据：mockTransport 调用日志 spy + 4 button request body 截取 + auth pending action 路径流 + 4xx revert 截图 + retired-testid grep 0 命中日志
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.022 行（关联需求 `frontend-home-job-picks-and-parse C-12, C-13, C-15`，状态 Ready，automated）

## E2E.P0.023 jd_match Search tab 自然语言搜索 + Saved searches + 4 chip filter + 5 步 AGENT panel + failure + auth gate + privacy

- [ ] 创建场景目录 `test/scenarios/e2e/p0-023-jd-match-search-and-saved/`，含 `README.md`
- [x] 准备 fixture variant：`searchJobs.json`（`default/empty/failed/slow-response`）、`listSavedSearches.json`（`default/empty/4xx`）、`createSavedSearch.json`（`default/4xx-validation`）；signed-in / signed-out 切换入口；`slow-response` variant 配置可观察延迟（≥1.5s）以便验证 5 步 AGENT panel 持续渲染；`make validate-fixtures` PASS <!-- evidence: 2026-05-10 plan §1.2 fixtures landed; searchJobs.slow-response 携带 X-Mock-Delay-Ms=4500 (≥1.5s); validate-fixtures OK -->
- [ ] 实现 `scripts/setup.sh`（含 fixture variant 切换 + signed-in/out 切换）/ `scripts/trigger.sh`（按 A-I 9 个子用例运行：切 tab listSavedSearches / Run default / Run slow-response / Run failed / Run empty / 4 filter 切换 / Save current + 4xx / 未登录 Run / 切 tab abort in-flight）/ `scripts/verify.sh`（断言 searchJobs body / Idempotency-Key、5 步 AGENT panel `jdmatch-search-searching-panel` + 5 个 `jdmatch-search-searching-step-${i}` testid 持续渲染至 settle、`opacity: 1` 前 3 步与 `opacity: 0.4` 后 2 步 computed style 命中、`● AGENT SCANNING` / `● AGENT 扫描中` accent label + accent 边框命中、动态 JD 数字（`248` / `87` / `unique postings` / `唯一岗位` / `248 → 87`）在 SearchTab DOM 与前端 i18n 0 命中、createSavedSearch body / Idempotency-Key、filter 4 状态、auth pending action params 不含 query / label、隐私反查 query / label / sourceJobUrl 0 命中、in-flight abort）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS（A-I 共 ≥9 子用例）
- [ ] 记录验证证据：mockTransport 调用日志 + 加载文案截图（in-flight + settle 后）+ 4 filter 切换截图 + auth pending action params 截取 + 隐私反查 grep 日志
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.023 行（关联需求 `frontend-home-job-picks-and-parse C-14, C-15`，状态 Ready，automated）

## E2E.P0.024 jd_match Watchlist tab + Market signals + chevron handoff + boundary + privacy

- [ ] 创建场景目录 `test/scenarios/e2e/p0-024-jd-match-watchlist-and-signals/`，含 `README.md`
- [x] 准备 fixture variant：`listWatchlist.json`（`default/empty/few/4xx`）、`getMarketSignals.json`（`default/partial-data/failed`）；预置两套 `listJobRecommendations` 状态（含已加载 linkedJobMatchId / 缺失 linkedJobMatchId）；`make validate-fixtures` PASS <!-- evidence: 2026-05-10 plan §1.2 fixtures landed; listWatchlist.default linkedJobMatchId 指向 jm-1/jm-2/jm-4 与 listJobRecommendations.many 的 jm-1..jm-4 同集；listWatchlist.few 仅 jm-1，与 listJobRecommendations.one/many 共存可演示已加载 / 缺失两路径；validate-fixtures OK -->
- [ ] 实现 `scripts/setup.sh`（含 fixture variant 切换 + Recommended 列表预置入口）/ `scripts/trigger.sh`（按 A-F 6 个子用例运行：default / empty / partial-data / chevron 已加载 linkedJobMatchId / chevron 未加载 linkedJobMatchId / listWatchlist 4xx）/ `scripts/verify.sh`（断言 listWatchlist + getMarketSignals 各调 1 次不重复、3 tone 渲染 + 3px 左边框、refresh footer i18n、empty CTA、partial-data fallback 文案、chevron 两路径行为、隐私反查 linkedJobMatchId/label/sourceJobUrl 0 命中）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS（A-F 共 ≥6 子用例）
- [ ] 记录验证证据：mockTransport 调用日志 + tone 渲染截图 + chevron handoff 流截图 + partial-data fallback 截图 + 隐私反查 grep 日志
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.024 行（关联需求 `frontend-home-job-picks-and-parse C-16`，状态 Ready，automated）

## E2E.P0.025 jd_match Profile chip + AGENT scan status + Auth pending action 跨 tab 综合

- [ ] 创建场景目录 `test/scenarios/e2e/p0-025-jd-match-profile-and-agent-status/`，含 `README.md`
- [x] 准备 fixture variant：`getJobMatchProfile.json`（`default/unauthenticated/partial-profile`）、`getAgentScanStatus.json`（`idle/scanning/error/next-scan-soon`）；signed-in / signed-out 切换入口；i18n zh/en + warm/light + dark + customAccent 全状态切换入口 <!-- evidence: 2026-05-10 plan §1.2 fixtures landed (getJobMatchProfile 3 variants + getAgentScanStatus 5 variants 含 next-scan-soon)；signed-in/out via Auth/getMe；i18n + theme 切换入口由 D1/D2 frontend-shell 提供 -->
- [ ] 实现 `scripts/setup.sh`（含 fixture variant + 主题切换 + 登录态切换入口）/ `scripts/trigger.sh`（按 A-J 10 个子用例运行：default profile + idle、切回 Recommended 验证 agent 再调 1 次、scanning agent、error agent、partial-profile、unauthenticated profile、未登录 Recommended/Search API side-effect 触发 auth_login 自动恢复、Source 与 Watchlist chevron 不触发 pendingAction、i18n 切换、三主题态切换）/ `scripts/verify.sh`（断言 profile 调 1 次 + agent 在切回 Recommended 时再调 1 次、不引入 setInterval/SSE/WebSocket spy 0 命中、AGENT badge tone 三态可见变化、profile chip 缺失字段 fallback、pendingAction params 不含隐私字段、selectedJobMatchId 自动恢复 + action 自动重新触发、Source/chevron 无 pendingAction、i18n 立即重绘、computed 颜色三态可见变化、plan 001 placeholder testid 0 命中）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS（A-J 共 ≥10 子用例）
- [ ] 记录验证证据：mockTransport 调用日志 + pendingAction 路径流 + Source/chevron 非 pendingAction 截取 + AGENT badge 三态截图 + i18n 切换截图 + 三主题态截图
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.025 行（关联需求 `frontend-home-job-picks-and-parse C-12, C-15`，状态 Ready，automated）

## E2E.P0.026 Confirm interview from jd_match → parse 出口 params 完整性 + parse 屏不破坏

- [ ] 创建场景目录 `test/scenarios/e2e/p0-026-jd-match-confirm-interview-handoff/`，含 `README.md`
- [x] 准备 fixture variant：复用 P0.022 `listJobRecommendations.many` + plan 001 已有的 `importTargetJob/getTargetJob/updateTargetJob` fixture；signed-in 入口；prefer 选中卡 ID = "jm-2" <!-- evidence: 2026-05-10 listJobRecommendations.many 含 jm-2 (01918fa0-0000-7000-8000-00000000a002) 与 plan 001 TargetJobs/* fixtures 可联用；validate-fixtures OK -->
- [ ] 实现 `scripts/setup.sh`（fixture variant 切换 + 选中 jm-2）/ `scripts/trigger.sh`（在 jd_match Recommended tab 选中 jm-2 → 点 Confirm interview → 进入 parse 屏 → 完成 plan 001 主路径）/ `scripts/verify.sh`（断言 nav 调 1 次 + params 仅含 `source: "jd_match"` 与 `sourceJobMatchId: "jm-2"` 两字段、source/sourceJobMatchId 不写入 URL/localStorage/sessionStorage/telemetry、parse 屏 D1 已有 testid 全部命中、plan 002 注入字段不修改 parse 屏 DOM 与现有 testid、E2E.P0.015 + E2E.P0.016 子用例 setup→trigger→verify→cleanup 仍 PASS）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：nav params 截取 + parse 屏 testid 命中日志 + URL/storage 隐私反查 grep + E2E.P0.015/016 regression PASS 日志
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.026 行（关联需求 `frontend-home-job-picks-and-parse C-13`，状态 Ready，automated）

## E2E.P0.017 升级（不再 grep placeholder）

- [ ] 升级 `test/scenarios/e2e/p0-017-jd-match-placeholder/scripts/verify.sh`：删除 `jdmatch-placeholder` / `jdmatch-placeholder-cta` / "Coming in P1" / "Coming Soon" 等文案 grep 断言；改为断言三 tab 数据驱动渲染（Profile chip + AGENT badge + Recommended 列表 + 三 tab 切换）；保留旧业务 testid 负向断言（plan 001 之前的 prototype 业务 testid 仍 0 命中）
- [ ] 同步更新 `test/scenarios/e2e/p0-017-jd-match-placeholder/README.md` 描述：从「P1 placeholder shell smoke」改为「三 tab 数据驱动 smoke + 旧 prototype 负向 grep」
- [ ] 同步 `test/scenarios/e2e/INDEX.md` P0.017 行描述
- [ ] 重跑 P0.017 setup→trigger→verify→cleanup PASS

## 整体 Regression（Phase 6 收口）

- [ ] D1+D2+D3+plan001 Regression 重跑：`E2E.P0.001 / 002 / 004 / 005 / 006 / 014 / 015 / 016` setup→trigger→verify→cleanup 全部 PASS（D2 视觉系统、D1 auth pending action、plan001 home/parse 流不被 plan 002 改动破坏）
- [ ] `pnpm --filter @easyinterview/frontend test` 全量 Vitest PASS（含本 plan 新增 ≥25 spec）
- [ ] `pnpm --filter @easyinterview/frontend test:pixel-parity` 在 plan 001 已升级 baseline 上累加新增三 tab × 三主题 spec，总数全 PASS，并确认升级后 jd_match SearchTab 的单一加载文案与 ui-design 真理源结构一致、5 步 panel DOM 0 命中
- [ ] `pnpm --filter @easyinterview/frontend typecheck` + `pnpm --filter @easyinterview/frontend build` + `make build` + `make validate-fixtures` 全 PASS
- [ ] `make docs-check` zero drift；`/sync-doc-index --fix-index` post-fix zero drift；`check_md_links` 双 OK
