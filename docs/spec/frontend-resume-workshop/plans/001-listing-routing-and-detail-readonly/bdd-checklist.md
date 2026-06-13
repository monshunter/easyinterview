# 001 BDD Checklist

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-06-13

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.036 resume list tree/flat toggle + retired negative + UI parity

- [x] 创建场景目录 `test/scenarios/e2e/p0-036-resume-list-tree-flat-toggle/`，含 `README.md`（§6 baseline + §7 离线限制）+ `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 fixture variant：`listResumes.json` `default` + `empty` + `paginated`；scenario-scoped `listResumeVersions.json` `default` + `master-only` + `with-targeted-branches`；数量断言从 fixture body 派生，不写死静态原型规模；`listResumes.default` 第二个 asset 无匹配 versions 必须作为 no-versions/partial 状态验证；用户未登录 / 已登录态与 lang fixture
- [x] 实现 `scripts/setup.sh`（A2 dev stack + Vite dev preview fixture-backed + 用户未登录 + lang=zh-CN 注入）/ `scripts/trigger.sh`（未登录路由加载 + 登录恢复 + ViewSwitcher 切换 + lang 切换 + 点击 version row + empty/paginated/partial 状态）/ `scripts/verify.sh`（断言 auth gate no-fetch + ≥ 15 testid + fixture-derived counts + tree/flat 切换 + TARGETED `tab=rewrites` + 当前 mock 不做 path-param-specific fixture selection + lang 切换 + Accept-Language header + UI parity DOM/computed/bounding/screenshot smoke + 隐私 grep + 旧入口 grep + prototype import grep）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录验证证据：`.test-output/e2e/p0-036-resume-list-tree-flat-toggle/trigger.log` + verify 输出 + fixture-derived count summary + no-versions/partial 状态截图或 DOM 证据 + auth no-fetch request spy + UI parity desktop/mobile artifacts + axe a11y report + retired-testid grep 0 + prototype import grep 0
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.036 行（关联需求 `frontend-resume-workshop C-1, C-2, C-3, C-5, C-6, C-7, C-9`，状态 Ready，automated）

## E2E.P0.037 resume detail Preview Tab + 原件弹层 + a11y + 404 fallback

- [x] 创建场景目录 `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 fixture variant：`getResumeVersion.json` `default` / `master-default` / `targeted-with-suggestions` / `not-found-404`；`getResume.json` `default` / `master-default` / `not-found`；`exportResumeVersion.json` `p0-501-not-available`（含 `request.headers.Idempotency-Key` 契约）；user 未登录 / 已登录态 + 已 cached `listResumeVersions`
- [x] 实现 `scripts/setup.sh`（A2 dev stack + Vite dev preview fixture-backed + 用户未登录 + version cache 准备）/ `scripts/trigger.sh`（未登录直达 detail + 登录恢复 + 点击 MASTER version → detail + 点击 TARGETED version → detail + 触发 原件 modal + ESC / 外层 / X 三种关闭路径 + 键盘 Tab focus trap + export PDF 501 + 访问 non-existent versionId）/ `scripts/verify.sh`（断言 detail 全 testid + MASTER/TARGETED 默认 tab 选择 + Preview 内容来源 + `exportResumeVersion` request header `Idempotency-Key` + export toast + modal a11y + 404 fallback + UI parity + 隐私 grep）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录验证证据：`.test-output/e2e/p0-037-resume-detail-preview-readonly/trigger.log` + verify 输出 + auth no-fetch request spy + export `Idempotency-Key` header spy + UI parity artifacts + axe a11y report + focus trap test 录制 + export 501 toast evidence + 404 fallback 截图 + 隐私 grep 0
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.037 行（关联需求 `frontend-resume-workshop C-4, C-5, C-6, C-7, C-8`，状态 Ready，automated）
- [x] 2026-05-23 real-backend gate：listing/detail 关联 scenario trigger scripts now run `frontendOwners.realApiMode.test.ts`, and verify scripts reject missing real-mode marker / default backend base URL / test-file marker; focused real-mode Vitest PASS.
