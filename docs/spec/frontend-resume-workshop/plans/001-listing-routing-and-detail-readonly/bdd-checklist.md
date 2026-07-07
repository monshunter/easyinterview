# 001 BDD Checklist

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.036 flat list + auth boundary

- [x] 场景目录为 `test/scenarios/e2e/p0-036-resume-flat-list-auth-boundary/`，含 `README.md`、`data/seed-input.md`、`data/expected-outcome.md` 与四段脚本。
- [x] `trigger.sh` 执行 `src/app/scenarios/p0-036-resume-flat-list-auth-boundary.test.tsx`，覆盖未登录 no-fetch、已登录 flat table、row open navigation 和 non-current route testid negative。
- [x] `verify.sh` 检查 trigger log 的 4 tests passed、测试文件 marker、non-current testid negative 和 fallback-text negative。
- [x] 场景在 `test/scenarios/e2e/INDEX.md` 登记为 Ready / automated，描述当前 flat list contract。

## E2E.P0.037 detail preview + original modal + 404 fallback

- [x] 场景目录为 `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/`，含 `README.md`、`data/seed-input.md`、`data/expected-outcome.md` 与四段脚本。
- [x] `trigger.sh` 执行 `src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx`，覆盖 preview default、explicit rewrites tab、original modal、export 501 和 not-found fallback。
- [x] `verify.sh` 检查 trigger log 的 5 tests passed、测试文件 marker、non-current testid negative、fallback-text negative 和 error-code no-echo。
- [x] 场景在 `test/scenarios/e2e/INDEX.md` 登记为 Ready / automated，描述当前 detail preview contract。
