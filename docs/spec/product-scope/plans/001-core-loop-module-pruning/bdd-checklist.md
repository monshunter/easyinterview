# Core Loop Module Pruning BDD Checklist

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-07-07

> Product-scope/001 Phase 6 tracks owner-document current boundary verification; completed D-22 scenario updates remain evidence, and active verification is tracked in the main checklist Phase 6.

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.001 默认首页只暴露核心三入口

- [x] 更新 scenario README / scripts / expected outcome，声明三入口和用户菜单新目标。
  <!-- verified: 2026-06-29 method=scenario-assets evidence="p0-001 README/expected/verify updated; frontend scenario test now expects three primary nav entries and no debrief/profile entry." -->
- [x] 执行当前 frontend BDD runner 并记录证据。
  <!-- verified: 2026-06-29 method=frontend-bdd-runner evidence="pnpm --filter @easyinterview/frontend test src/app/scenarios/p0-001-default-home-shell.test.tsx PASS (included in 39-test BDD gate run)." -->

## E2E.P0.088 canonical route 不再包含 debrief/profile

- [x] 更新 canonical path matrix，删除 `/debrief` 和 `/profile` 正向 deep-link。
  <!-- verified: 2026-06-29 method=scenario-assets evidence="p0-088 README/seed/expected/verify updated; frontend scenario test now treats /debrief and /profile as out-of-scope paths folding to home." -->
- [x] 执行当前 frontend BDD runner 并记录证据。
  <!-- verified: 2026-06-29 method=frontend-bdd-runner evidence="pnpm --filter @easyinterview/frontend test src/app/scenarios/p0-088-url-addressable-routing-canonical.test.tsx PASS (included in 39-test BDD gate run)." -->

## E2E.P0.090 hash out-of-scope route 不 materialize out-of-scope modules

- [x] 增加 `debrief`、`debrief_full`、`profile` out-of-scope-negative 输入。
  <!-- verified: 2026-06-29 method=scenario-assets evidence="p0-090 README/seed/expected/verify updated; frontend scenario test now includes debrief/debrief_full/profile out-of-scope aliases." -->
- [x] 执行当前 frontend BDD runner 并记录证据。
  <!-- verified: 2026-06-29 method=frontend-bdd-runner evidence="pnpm --filter @easyinterview/frontend test src/app/scenarios/p0-090-url-routing-hash-out-of-scope-negative.test.tsx PASS (included in 39-test BDD gate run)." -->

## E2E.P0.102 未登录保护路由不把范围外模块当业务目标

- [x] 更新 auth-gated route matrix，删除范围外模块 pendingAction 正向路径。
  <!-- verified: 2026-06-29 method=scenario-assets evidence="p0-102 README/seed/expected/verify updated; AppAuthDispatch/HomeAuthGate tests no longer treat debrief/profile as protected business routes." -->
- [x] 执行当前 frontend BDD runner 并记录证据。
  <!-- verified: 2026-06-29 method=frontend-bdd-runner evidence="pnpm --filter @easyinterview/frontend test src/app/screens/home/HomeAuthGate.test.tsx src/app/AppAuthDispatch.test.tsx PASS (included in 39-test BDD gate run)." -->

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
