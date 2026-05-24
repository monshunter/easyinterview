# 001 Full Funnel Happy Journey BDD Checklist

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-05-24

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)

## E2E.P0.098 — API-level Full Funnel: Import to Next Round

- [ ] 098.A 创建 scenario 目录 `test/scenarios/e2e/p0-098-full-funnel-import-to-next-round-journey/`
- [ ] 098.B 准备 `data/seed-input.md`（user A + registerResume resume source + paste JD 文本）与 `data/expected-outcome.md`（resume_parse ready + handoff 链 + ready 状态 + 幂等 + 隐私预期）
- [ ] 098.C 编写 `scripts/setup.sh`：确认 dev-stack postgres + migrate；准备 `.test-output/e2e/p0-098-*/` 与 setup marker
- [ ] 098.D 编写 `scripts/trigger.sh`：`cd backend && go test -v ./cmd/api -run 'TestE2EP0098' -count=1` 并 `tee` 到 trigger.log，保留 exit code
- [ ] 098.E 编写 `scripts/verify.sh`：断言 trigger.log 含 Go runner marker（`PASS` + `TestE2EP0098`）+ 拒绝 `no tests`/`SKIP`-as-pass + 隐私 grep + route-aware legacy 负向 grep（覆盖旧 route，避免误伤合法 `createPracticePlan` / `resumeAssetId`）
- [ ] 098.F 编写 `scripts/cleanup.sh`：清 journey 创建的 DB 行 + 关闭资源
- [ ] 098.G 确认四段脚本可独立执行；`trigger.sh` 保留 runner exit code；`verify.sh` 含 pass marker + 负向 grep
- [ ] 098.H 编写 scenario README 描述 isolation / setup / cleanup 协议；登记到 `test/scenarios/e2e/INDEX.md`
- [ ] 098.I 在场景目录内按 `setup → trigger → verify → cleanup` 执行通过；wrapper 必须 cleanup 后按 pre-cleanup 失败码退出；记录证据到 `.test-output/e2e/p0-098-*/trigger.log`
- [ ] 098.J BDD-Gate 通过：plan checklist 3.2 勾选

## E2E.P0.099 — Full-Stack UI Full Funnel Journey

- [ ] 099.A 创建 scenario 目录 `test/scenarios/e2e/p0-099-full-funnel-fullstack-ui-journey/`
- [ ] 099.B 准备 `data/seed-input.md`（user + resume asset + JD 文本）与 `data/expected-outcome.md`（跨屏 nav + 轮询过渡 + CTA handoff + 隐私预期）
- [ ] 099.C 编写 `scripts/setup.sh`：拉起真后端进程（dev-stack postgres + stub AI）+ 前端 build/preview 通过 `VITE_EI_API_MODE=real` / `VITE_EI_API_BASE_URL=http://127.0.0.1:<backend-port>/api/v1` 指向真后端；health probe；通过 `registerResume` + `resume_parse` stub seed user + ready resume
- [ ] 099.D 编写 `frontend/playwright.e2e.config.ts`（`testDir: "./tests/e2e"`；`outputDir` 读取 `EI_PLAYWRIGHT_OUTPUT_DIR`，默认 repo 根 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/playwright`）与 `scripts/trigger.sh`：`EI_PLAYWRIGHT_OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/playwright" pnpm --filter @easyinterview/frontend exec playwright test --config=playwright.e2e.config.ts tests/e2e/full-funnel-journey.spec.ts` 并 `tee` 到 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/trigger.log`，保留 exit code
- [ ] 099.E 编写 `scripts/verify.sh`：断言 trigger.log 含 Playwright marker（`passed` 计数 + spec 文件名）+ 拒绝 `no tests`/全 skip + 隐私（URL/storage/console）grep + route-aware legacy 负向 grep + frontend scope gate 或等价 scoped grep + 断言 Playwright trace/screenshot/video/artifacts 位于 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/` 且未写入 `frontend/.playwright-output` / `frontend/test-results`
- [ ] 099.F 编写 `scripts/cleanup.sh`：清 DB 行 + 停前后端进程；保留 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/` 下的 trigger.log 与 Playwright 证据产物
- [ ] 099.G 确认四段脚本可独立执行；`trigger.sh` 保留 runner exit code；`verify.sh` 含 pass marker + 负向 grep
- [ ] 099.H 编写 scenario README 描述 isolation / setup / cleanup 协议；登记到 `test/scenarios/e2e/INDEX.md`
- [ ] 099.I 在场景目录内按 `setup → trigger → verify → cleanup` 执行通过；wrapper 必须 cleanup 后按 pre-cleanup 失败码退出；记录证据到 `.test-output/e2e/p0-099-*/trigger.log`
- [ ] 099.J BDD-Gate 通过：plan checklist 3.3 勾选

## 收口

- [ ] 9.A 两个 scenario `Ready` 状态登记到 `test/scenarios/e2e/INDEX.md`
- [ ] 9.B 两个 scenario 一次性顺序执行通过：`status=0; for s in p0-098-full-funnel-import-to-next-round-journey p0-099-full-funnel-fullstack-ui-journey; do rc=0; cleanup_rc=0; (cd test/scenarios/e2e/$s && bash scripts/setup.sh && bash scripts/trigger.sh && bash scripts/verify.sh) || rc=$?; (cd test/scenarios/e2e/$s && bash scripts/cleanup.sh) || cleanup_rc=$?; if [ "$rc" -ne 0 ]; then status=$rc; break; fi; if [ "$cleanup_rc" -ne 0 ]; then status=$cleanup_rc; break; fi; done; exit "$status"`
- [ ] 9.C 全部 scenario 证据 `.test-output/e2e/<scenario>/trigger.log` 已记录并由对应 `verify.sh` 消费；P0.099 Playwright 产物仅存在于 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/`
