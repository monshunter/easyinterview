# 001 BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-11

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.036 resume list tree/flat toggle + retired negative + UI parity

- [ ] 创建场景目录 `test/scenarios/e2e/p0-036-resume-list-tree-flat-toggle/`，含 `README.md`（§6 baseline + §7 离线限制）+ `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture variant：`listResumes.json` `default` (5 resume_asset) + `empty` + `paginated`；`listResumeVersions.json` `default` (3 master + 9 targeted 共 12 version)；用户登录态 / lang fixture
- [ ] 实现 `scripts/setup.sh`（A2 dev stack + Vite dev preview fixture-backed + 用户登录 + lang=zh-CN 注入）/ `scripts/trigger.sh`（路由加载 + ViewSwitcher 切换 + lang 切换 + 点击 version row）/ `scripts/verify.sh`（断言 ≥ 15 testid + tree/flat 切换 + lang 切换 + Accept-Language header + UI parity baseline + 隐私 grep + 旧入口 grep + data.jsx import grep）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-036-resume-list-tree-flat-toggle/trigger.log` + verify 输出 + UI parity baseline screenshot desktop + mobile + axe a11y report + retired-testid grep 0 + data.jsx import grep 0
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.036 行（关联需求 `frontend-resume-workshop C-1, C-2, C-3, C-5, C-6, C-7, C-9`，状态 Ready，automated）

## E2E.P0.037 resume detail Preview Tab + 原件弹层 + a11y + 404 fallback

- [ ] 创建场景目录 `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture variant：`getResumeVersion.json` `master-default` / `targeted-with-suggestions` / `not-found-404`；user 登录态 + 已 cached `listResumeVersions`
- [ ] 实现 `scripts/setup.sh`（A2 dev stack + Vite dev preview fixture-backed + 用户登录 + version cache 准备）/ `scripts/trigger.sh`（点击 MASTER version → detail + 点击 TARGETED version → detail + 触发 原件 modal + ESC / 外层 / X 三种关闭路径 + 键盘 Tab focus trap + 访问 non-existent versionId）/ `scripts/verify.sh`（断言 detail 全 testid + 默认 tab 选择 + Preview 内容来源 + modal a11y + 404 fallback + UI parity + 隐私 grep）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-037-resume-detail-preview-readonly/trigger.log` + verify 输出 + UI parity baseline screenshot + axe a11y report + focus trap test 录制 + 404 fallback 截图 + 隐私 grep 0
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.037 行（关联需求 `frontend-resume-workshop C-4, C-5, C-6, C-7, C-8`，状态 Ready，automated）
