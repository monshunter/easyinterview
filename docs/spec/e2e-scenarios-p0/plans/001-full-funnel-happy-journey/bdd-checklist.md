# 001 Full Funnel Happy Journey BDD Checklist

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-05-24

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)

## E2E.P0.098 — API-level Full Funnel: Import to Next Round

- [x] 098.A 创建 scenario 目录 `test/scenarios/e2e/p0-098-full-funnel-import-to-next-round-journey/`
  <!-- verified: 2026-05-24 evidence="scenario directory exists with README, data, and scripts" -->
- [x] 098.B 准备 `data/seed-input.md`（user A + registerResume resume source + paste JD 文本）与 `data/expected-outcome.md`（resume_parse ready + handoff 链 + ready 状态 + 幂等 + 隐私预期）
  <!-- verified: 2026-05-24 evidence="data/seed-input.md and data/expected-outcome.md define registerResume resume seed, paste JD, handoff chain, idempotency, privacy, and legacy-negative expectations" -->
- [x] 098.C 编写 `scripts/setup.sh`：确认 dev-stack postgres + migrate；准备 `.test-output/e2e/p0-098-*/` 与 setup marker
  <!-- verified: 2026-05-24 evidence="setup.sh checks postgres via psql, runs make migrate-status, clears scenario output, and writes setup.env" -->
- [x] 098.D 编写 `scripts/trigger.sh`：`cd backend && go test -v ./cmd/api -run 'TestE2EP0098' -count=1` 并 `tee` 到 trigger.log，保留 exit code
  <!-- verified: 2026-05-24 evidence="trigger.sh runs backend go test with DATABASE_URL and pipes output to .test-output/e2e/p0-098-full-funnel-import-to-next-round-journey/trigger.log without masking failure under set -euo pipefail" -->
- [x] 098.E 编写 `scripts/verify.sh`：断言 trigger.log 含 Go runner marker（`PASS` + `TestE2EP0098`）+ 拒绝 `no tests`/`SKIP`-as-pass + 隐私 grep + route-aware legacy 负向 grep（覆盖旧 route，避免误伤合法 `createPracticePlan` / `resumeAssetId`）
  <!-- verified: 2026-05-24 evidence="verify.sh requires three E2E.P0.098 PASS markers, runner job markers, package ok marker, private-marker absence, and route-aware legacy-negative scan with canonical-token false-positive guard" -->
- [x] 098.F 编写 `scripts/cleanup.sh`：清 journey 创建的 DB 行 + 关闭资源
  <!-- verified: 2026-05-24 evidence="cleanup.sh keeps output evidence and clears scenario scratch markers; journey DB rows are deleted by scenario cleanup helpers in the Go harness" -->
- [x] 098.G 确认四段脚本可独立执行；`trigger.sh` 保留 runner exit code；`verify.sh` 含 pass marker + 负向 grep
  <!-- verified: 2026-05-24 command="cd test/scenarios/e2e/p0-098-full-funnel-import-to-next-round-journey && bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh && bash scripts/cleanup.sh" evidence="setup: ok; trigger: ok; verify: ok; cleanup: ok" -->
- [x] 098.H 编写 scenario README 描述 isolation / setup / cleanup 协议；登记到 `test/scenarios/e2e/INDEX.md`
  <!-- verified: 2026-05-24 evidence="README documents isolation, setup/trigger/verify/cleanup, evidence, and cleanup behavior; INDEX row is Ready" -->
- [x] 098.I 在场景目录内按 `setup → trigger → verify → cleanup` 执行通过；wrapper 必须 cleanup 后按 pre-cleanup 失败码退出；记录证据到 `.test-output/e2e/p0-098-*/trigger.log`
  <!-- verified: 2026-05-24 command="overall=0; for s in p0-098-full-funnel-import-to-next-round-journey p0-099-full-funnel-fullstack-ui-journey; do rc=0; cleanup_rc=0; (cd \"test/scenarios/e2e/$s\" && bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh) || rc=$?; (cd \"test/scenarios/e2e/$s\" && bash scripts/cleanup.sh) || cleanup_rc=$?; if [ \"$rc\" -ne 0 ]; then overall=$rc; break; fi; if [ \"$cleanup_rc\" -ne 0 ]; then overall=$cleanup_rc; break; fi; done; exit \"$overall\"" evidence="P0.098 setup/trigger/verify/cleanup all ok and trigger.log contains TestE2EP0098 pass markers plus resume_parse/target_import/report_generate runner logs" -->
- [x] 098.J BDD-Gate 通过：plan checklist 3.2 勾选
  <!-- verified: 2026-05-24 evidence="main checklist 3.2 is checked with wrapper evidence" -->

## E2E.P0.099 — Full-Stack UI Full Funnel Journey

- [x] 099.A 创建 scenario 目录 `test/scenarios/e2e/p0-099-full-funnel-fullstack-ui-journey/`
  <!-- verified: 2026-05-24 evidence="scenario directory exists with README, data, and scripts" -->
- [x] 099.B 准备 `data/seed-input.md`（user + resume asset + JD 文本）与 `data/expected-outcome.md`（跨屏 nav + 轮询过渡 + CTA handoff + 隐私预期）
  <!-- verified: 2026-05-24 evidence="data files define user/resume/JD seed plus navigation, polling, next_round handoff, privacy, and legacy-negative expectations" -->
- [x] 099.C 编写 `scripts/setup.sh`：确认 dev-stack postgres + migrate；准备 `.test-output/e2e/p0-099-*/` 与 setup marker；真后端测试进程、ready resume seed、前端 build/preview 和 health probe 由 `scripts/trigger.sh` 内的 Playwright `webServer` 托管启动
  <!-- verified: 2026-05-24 evidence="setup.sh checks postgres and migration status; trigger log shows P0.099 backend server listening, resume_parse seed succeeded, Vite preview ready, and Playwright health probe/test execution reached real backend" -->
- [x] 099.D 编写 `frontend/playwright.e2e.config.ts`（`testDir: "./tests/e2e"`；`outputDir` 读取 `EI_PLAYWRIGHT_OUTPUT_DIR`，默认 repo 根 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/playwright`）与 `scripts/trigger.sh`：`EI_PLAYWRIGHT_OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/playwright" pnpm --filter @easyinterview/frontend exec playwright test --config=playwright.e2e.config.ts tests/e2e/full-funnel-journey.spec.ts` 并 `tee` 到 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/trigger.log`，保留 exit code
  <!-- verified: 2026-05-24 evidence="playwright.e2e.config.ts defines testDir ./tests/e2e and scenario outputDir; trigger.sh exports DATABASE_URL and EI_PLAYWRIGHT_OUTPUT_DIR then runs the full-funnel spec through pnpm filter" -->
- [x] 099.E 编写 `scripts/verify.sh`：断言 trigger.log 含 Playwright marker（`passed` 计数 + spec 文件名）+ 拒绝 `no tests`/全 skip + 隐私（URL/storage/console）grep + route-aware legacy 负向 grep + frontend scope gate 或等价 scoped grep + 断言 Playwright trace/screenshot/video/artifacts 位于 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/` 且未写入 `frontend/.playwright-output` / `frontend/test-results`
  <!-- verified: 2026-05-24 evidence="verify.sh requires spec name, backend listening marker, resume_parse/target_import/report_generate runner markers, 1 passed marker, state.json, scenario Playwright output, real API mode config, private-marker absence, artifact-location guard, and scoped legacy-negative scan excluding test-only assertions" -->
- [x] 099.F 编写 `scripts/cleanup.sh`：保留 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/` 下的 trigger.log 与 Playwright 证据产物；前后端进程由 Playwright `webServer` 退出时停止
  <!-- verified: 2026-05-24 evidence="cleanup.sh keeps trigger.log/state.json/playwright evidence; Playwright webServer manages backend/frontend process lifecycle" -->
- [x] 099.G 确认四段脚本可独立执行；`trigger.sh` 保留 runner exit code；`verify.sh` 含 pass marker + 负向 grep
  <!-- verified: 2026-05-24 command="cd test/scenarios/e2e/p0-099-full-funnel-fullstack-ui-journey && bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh && bash scripts/cleanup.sh" evidence="setup: ok; Playwright 1 passed; trigger: ok; verify: ok; cleanup: ok" -->
- [x] 099.H 编写 scenario README 描述 isolation / setup / cleanup 协议；登记到 `test/scenarios/e2e/INDEX.md`
  <!-- verified: 2026-05-24 evidence="README documents isolation, real backend/frontend execution, evidence, and cleanup behavior; INDEX row is Ready" -->
- [x] 099.I 在场景目录内按 `setup → trigger → verify → cleanup` 执行通过；wrapper 必须 cleanup 后按 pre-cleanup 失败码退出；记录证据到 `.test-output/e2e/p0-099-*/trigger.log`
  <!-- verified: 2026-05-24 command="overall=0; for s in p0-098-full-funnel-import-to-next-round-journey p0-099-full-funnel-fullstack-ui-journey; do rc=0; cleanup_rc=0; (cd \"test/scenarios/e2e/$s\" && bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh) || rc=$?; (cd \"test/scenarios/e2e/$s\" && bash scripts/cleanup.sh) || cleanup_rc=$?; if [ \"$rc\" -ne 0 ]; then overall=$rc; break; fi; if [ \"$cleanup_rc\" -ne 0 ]; then overall=$cleanup_rc; break; fi; done; exit \"$overall\"" evidence="P0.099 setup/trigger/verify/cleanup all ok; trigger.log contains backend listening, Vite preview, Playwright 1 passed, and runner logs for resume_parse/target_import/report_generate" -->
- [x] 099.J BDD-Gate 通过：plan checklist 3.3 勾选
  <!-- verified: 2026-05-24 evidence="main checklist 3.3 is checked with wrapper evidence" -->

## 收口

- [x] 9.A 两个 scenario `Ready` 状态登记到 `test/scenarios/e2e/INDEX.md`
  <!-- verified: 2026-05-24 evidence="test/scenarios/e2e/INDEX.md contains Ready rows for E2E.P0.098 and E2E.P0.099" -->
- [x] 9.B 两个 scenario 一次性顺序执行通过：`overall=0; for s in p0-098-full-funnel-import-to-next-round-journey p0-099-full-funnel-fullstack-ui-journey; do rc=0; cleanup_rc=0; (cd test/scenarios/e2e/$s && bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh) || rc=$?; (cd test/scenarios/e2e/$s && bash scripts/cleanup.sh) || cleanup_rc=$?; if [ "$rc" -ne 0 ]; then overall=$rc; break; fi; if [ "$cleanup_rc" -ne 0 ]; then overall=$cleanup_rc; break; fi; done; exit "$overall"`
  <!-- verified: 2026-05-24 command="overall=0; for s in p0-098-full-funnel-import-to-next-round-journey p0-099-full-funnel-fullstack-ui-journey; do rc=0; cleanup_rc=0; (cd \"test/scenarios/e2e/$s\" && bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh) || rc=$?; (cd \"test/scenarios/e2e/$s\" && bash scripts/cleanup.sh) || cleanup_rc=$?; if [ \"$rc\" -ne 0 ]; then overall=$rc; break; fi; if [ \"$cleanup_rc\" -ne 0 ]; then overall=$cleanup_rc; break; fi; done; exit \"$overall\"" evidence="exit code 0; P0.098 and P0.099 both reported setup: ok, trigger: ok, verify: ok, cleanup: ok" -->
- [x] 9.C 全部 scenario 证据 `.test-output/e2e/<scenario>/trigger.log` 已记录并由对应 `verify.sh` 消费；P0.099 Playwright 产物仅存在于 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/`
  <!-- verified: 2026-05-24 evidence="both verify.sh scripts consume scenario trigger.log; P0.099 verify.sh asserts state.json/playwright output and rejects newer files under frontend/.playwright-output or frontend/test-results" -->
