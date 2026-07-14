# Core Loop Module Pruning BDD Checklist

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-14

> Completed D-22 scenario evidence remains valid. Product-scope/001 Phase 7 的 Parse/Workspace/theme 场景修订与当前源码重跑已完成。

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.001 默认首页只暴露核心三入口

- [x] 更新 scenario README / scripts / expected outcome，声明三入口和用户菜单新目标。
  <!-- verified: 2026-06-29 method=scenario-assets evidence="p0-001 README/expected/verify updated; frontend scenario test now expects three primary nav entries and no debrief/profile entry." -->
- [x] 执行当前 frontend BDD runner 并记录证据。
  <!-- verified: 2026-06-29 method=frontend-bdd-runner evidence="pnpm --filter @easyinterview/frontend test src/app/scenarios/p0-001-default-home-shell.test.tsx PASS (included in 39-test BDD gate run)." -->

## E2E.P0.088 canonical route 与最小 safe params

- [x] 更新 canonical path matrix，删除 `/debrief` 和 `/profile` 正向 deep-link。
  <!-- verified: 2026-06-29 method=scenario-assets evidence="p0-088 README/seed/expected/verify updated; frontend scenario test now treats /debrief and /profile as out-of-scope paths folding to home." -->
- [x] 执行当前 frontend BDD runner 并记录证据。
  <!-- verified: 2026-06-29 method=frontend-bdd-runner evidence="pnpm --filter @easyinterview/frontend test src/app/scenarios/p0-088-url-addressable-routing-canonical.test.tsx PASS (included in 39-test BDD gate run)." -->
- [x] 更新当前 route matrix：Parse/Workspace 均只允许 `targetJobId`；Workspace 允许 direct detail 并丢弃 `planId/resumeId`，bad identity replace `/workspace`。 <!-- verified: 2026-07-14 method=route-matrix+source-gate result=PASS -->
- [x] 执行场景四段脚本，覆盖 direct/reload/back-forward/auth recovery 与 no Back loop。 <!-- verified: 2026-07-14 method=scenario result=PASS evidence="P0.088 fresh; P0.089/P0.102 auth/privacy companion gates PASS" -->

## E2E.P0.090 旧 hash/query 不恢复混合页面或范围外模块

- [x] 增加 `debrief`、`debrief_full`、`profile` out-of-scope-negative 输入。
  <!-- verified: 2026-06-29 method=scenario-assets evidence="p0-090 README/seed/expected/verify updated; frontend scenario test now includes debrief/debrief_full/profile out-of-scope aliases." -->
- [x] 执行当前 frontend BDD runner 并记录证据。
  <!-- verified: 2026-06-29 method=frontend-bdd-runner evidence="pnpm --filter @easyinterview/frontend test src/app/scenarios/p0-090-url-routing-hash-out-of-scope-negative.test.tsx PASS (included in 39-test BDD gate run)." -->
- [x] 更新旧 Parse report params 与 Workspace extra params 的 canonical assertions，阻止 ready detail/report entry materialize 到 Parse。 <!-- verified: 2026-07-14 method=canonical-assertions result=PASS -->
- [x] 执行场景四段脚本并记录 current-negative search 证据。 <!-- verified: 2026-07-14 method=scenario+negative-search result=PASS evidence="P0.090 fresh on current source" -->

## E2E.P0.102 未登录保护路由不把范围外模块当业务目标

- [x] 更新 auth-gated route matrix，删除范围外模块 pendingAction 正向路径。
  <!-- verified: 2026-06-29 method=scenario-assets evidence="p0-102 README/seed/expected/verify updated; AppAuthDispatch/HomeAuthGate tests no longer treat debrief/profile as protected business routes." -->
- [x] 执行当前 frontend BDD runner 并记录证据。
  <!-- verified: 2026-06-29 method=frontend-bdd-runner evidence="pnpm --filter @easyinterview/frontend test src/app/screens/home/HomeAuthGate.test.tsx src/app/AppAuthDispatch.test.tsx PASS (included in 39-test BDD gate run)." -->
  <!-- verified: 2026-07-14 method=scenario result=PASS evidence="E2E.P0.102 fresh on current source" -->

## E2E.P0.098 API-level 核心闭环不依赖复盘 / 画像

- [x] 更新 full-funnel expected outcome，明确无 debrief/profile API/table/event 参与。
  <!-- verified: 2026-06-29 method=scenario-assets evidence="P0.098 README/expected/verify updated with D-22 negative scan for Debriefs/Profile operations, candidate_profiles/experience_cards tables, source_debrief_id, debrief jobs/events, and profile.update." -->
- [x] 执行 setup / trigger / verify / cleanup 并记录当前 runner 证据。
  <!-- verified: 2026-06-29 method=scenario-runner evidence="test/scenarios/e2e/p0-098-full-funnel-import-to-next-round-journey/scripts/setup.sh && trigger.sh && verify.sh && cleanup.sh PASS." -->

## E2E.P0.099 Fullstack UI 核心闭环不出现复盘 / 画像入口

- [x] 更新 Playwright full-stack assertions，TopBar 与用户菜单无复盘 / 用户画像入口。
  <!-- verified: 2026-06-29 method=scenario-assets evidence="P0.099 README/expected/verify and Playwright route assertions now require three-entry TopBar, settings/logout user menu only, no debrief/profile testids, and D-22 negative UI token scan." -->
- [x] 执行 setup / trigger / verify / cleanup 并记录当前 runner 证据。
  <!-- verified: 2026-06-29 method=scenario-runner evidence="test/scenarios/e2e/p0-099-full-funnel-fullstack-ui-journey/scripts/setup.sh && trigger.sh PASS with target_import/report_generate succeeded; verify.sh PASS; cleanup.sh PASS. Scenario harness now completes first-login profile setup and cleanup targets current resumes table." -->

## Out-of-scope 场景删除矩阵

- [x] 删除 E2E.P0.060-E2E.P0.069 正向场景目录和 `test/scenarios/e2e/INDEX.md` 行。
  <!-- verified: 2026-06-29 method=scenario-assets evidence="Cleaned P0.060-P0.069 directories and INDEX rows." -->
- [x] 删除 E2E.P0.071、E2E.P0.073 正向场景目录和 `test/scenarios/e2e/INDEX.md` 行。
  <!-- verified: 2026-06-29 method=scenario-assets evidence="Cleaned P0.071/P0.073 directories and INDEX rows; P0.070/P0.072 now cover report-derived retry/next-round only." -->
- [x] 删除 E2E.P0.091-E2E.P0.093 正向场景目录和 `test/scenarios/e2e/INDEX.md` 行。
  <!-- verified: 2026-06-29 method=scenario-assets evidence="Cleaned P0.091-P0.093 directories and INDEX rows." -->
- [x] 运行 scenario INDEX negative grep，确认 out-of-scope 场景不再显示 Ready。
  <!-- verified: 2026-06-29 method=scenario-contract-test evidence="python3 -m pytest -q scripts/lint/scenario_script_contract_test.py PASS (2 tests); out-of-scope scenario IDs absent from test/scenarios/e2e/INDEX.md." -->

## E2E.P0.015 新 JD import 进入纯 Parse progress

- [x] 更新 scenario/unit/source assertions：只有 `importTargetJob` 成功后可进入 Parse；queued/processing 显示进度，ready 初读/轮询使用 replace 进入 Workspace。 <!-- verified: 2026-07-14 method=scenario+unit+source result=PASS -->
- [x] 执行 setup / trigger / verify / cleanup，记录真实 transport 次数与 no-ready-Parse-detail 证据。 <!-- verified: 2026-07-14 method=scenario result=PASS run_id=8d1484ca-746b-485a-a004-272970cb8990 evidence="initial GET=1 then one per scheduler tick; ready Parse detail absent" -->

## E2E.P0.016 ready handoff 与 Workspace 报告入口

- [x] 把只读详情与报告入口断言迁到 Workspace，删除 Parse ready detail/report entry/animation 的正向断言并增加负向 gate。 <!-- verified: 2026-07-14 method=source+unit result=PASS -->
- [x] 执行 unit/source/pixel runner 和场景四段脚本，记录 1440/390 证据。 <!-- verified: 2026-07-14 method=scenario+browser result=PASS evidence="P0.016 fresh; 8/8 desktop/mobile browser checks" -->

## E2E.P0.018 ready 规划卡片直达详情

- [x] 更新 Home/Workspace card、direct/reload 与 list/detail assertions：ready card 直达 targetJobId-only Workspace detail，`/workspace` 无 target 时仍为列表。 <!-- verified: 2026-07-14 method=scenario+route-tests result=PASS -->
- [x] 执行场景四段脚本，证明 direct detail 只读 `getTargetJob` 一次、不调用 import 或 Parse animation。 <!-- verified: 2026-07-14 method=scenario+transport-count result=PASS evidence="P0.018/P0.021 fresh; direct detail has no import, mutation, or Parse animation" -->

## E2E.P0.046 Practice terminal recovery 回当前规划

- [x] 更新 terminal failure CTA 断言为 `/workspace?targetJobId=...`，保留 no-retry/no-duplicate-banner 与可信 identity gate。 <!-- verified: 2026-07-14 method=scenario-contract result=PASS -->
- [x] 执行场景四段脚本并记录恢复路径不进入 Parse 的证据。 <!-- verified: 2026-07-14 method=scenario result=PASS run_id=04fd68b6-1ed7-49d7-b271-a37c236ea541 -->

## E2E.P0.058 Report/Generating 可信与不可信 Back 分层

- [x] 保持可信 Report/Generating Back 到 ReportsScreen、无可信 identity 回 `/workspace`，并增加 no-Parse fallback 断言。 <!-- verified: 2026-07-14 method=route+source-gates result=PASS -->
- [x] 执行场景四段脚本并记录 trusted/untrusted 两类证据。 <!-- verified: 2026-07-14 method=scenario result=PASS evidence="P0.058 fresh on current source" -->

## E2E.P0.059 Reports Back 回 Workspace detail

- [x] 更新 ReportsScreen Back 与 pixel/i18n assertions：可信 target 返回 `/workspace?targetJobId=...`，报告详情仍返回 ReportsScreen。 <!-- verified: 2026-07-14 method=source+unit+pixel result=PASS -->
- [x] 执行场景四段脚本和 browser acceptance，记录 desktop/mobile Back 与 screenshot 证据。 <!-- verified: 2026-07-14 method=scenario+browser result=PASS evidence="P0.059 fresh on current source" -->

## E2E.P0.005 Custom Accent 最小 DOM

- [x] 更新 source/jsdom smoke：只保留 Ocean/Plum 与 hue/saturation；preview/value/reset testid、copy、prop 与 DOM 为零，preset 选择退出 custom。 <!-- verified: 2026-07-14 method=source+jsdom result=PASS evidence="UI contract 65/65" -->
- [x] 执行场景四段脚本并记录 custom/preset/light/dark 证据。 <!-- verified: 2026-07-14 method=scenario result=PASS evidence="P0.005 fresh on current source" -->

## E2E.P0.006 Ocean/Plum 与 custom desktop/mobile parity

- [x] 更新原型与正式 TopBar parity anchors，删除 preview/value/reset baseline，保留 hue/saturation 与 preset exit-custom 状态。 <!-- verified: 2026-07-14 method=prototype-formal-contract result=PASS -->
- [x] 执行 Playwright 1440/390 DOM/computed-style/bounding-box/pixel gate，保存非空截图证据。 <!-- verified: 2026-07-14 method=scenario+browser result=PASS evidence="P0.006 fresh; parity 170/170 on desktop/mobile" -->
