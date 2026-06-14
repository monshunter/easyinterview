# UX Funnel Refactor Drift 交付复盘报告

> **日期**: 2026-06-14
> **审查人**: Codex (GPT-5)

## 1 复盘范围与成功证据

本次会话从 `7e6ec756d91280b9046d34f362a3cde58bbf2d29` 之后的变更出发，对 UX funnel simplification 下游实现做 deep reconcile，并修复 D-17 模块删除、D-18 workspace 嵌入、D-20 简历扁平化后遗留在场景、contract lint、frontend tests、backend e2e helper 和 fixture README 中的不一致口径。

已闭环的主要范围：

- OpenAPI / fixture / mock contract inventory 对齐到当前 12 tags、43 operations、43 fixtures，删除旧 `resume_versions` operation 口径。
- Backend route catalog 和 full-funnel helper 对齐到 flat `resumeId`、`resume_id` schema、6 位 auth code 和语义 JSONB 比较。
- Frontend route、README、i18n、pixel/e2e tests 对齐四入口 TopBar、workspace 嵌入、无 live `company_intel` route、无 guided create。
- Scenario assets 和 verify gates 对齐当前字段名、当前导航结构和 retired route 负向断言。
- Engineering roadmap spec/history 和 docs index 对齐本次 D-17/D-18/D-20 refactor 后的当前事实。

关键通过证据：

- `bash test/scenarios/env-setup.sh --with-migrations`
- `E2E.P0.035`、`E2E.P0.004`、`E2E.P0.016`、`E2E.P0.098` setup / trigger / verify / cleanup
- `E2E.P0.079`、`E2E.P0.083`、`E2E.P0.089`、`E2E.P0.006` setup / trigger / verify / cleanup
- `cd backend && go test ./...`
- `pnpm --filter @easyinterview/frontend typecheck`
- `pnpm --filter @easyinterview/frontend test -- --run`
- `python3 -m pytest scripts/lint/events_inventory_test.py scripts/lint/openapi_inventory_test.py scripts/lint/validate_fixtures_test.py scripts/mock_contract/fixture_registry_test.py -q`
- `make validate-fixtures`
- `make codegen-check`
- `make docs-check`
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
- `pnpm --filter @easyinterview/frontend build`
- `pnpm --filter @easyinterview/frontend test:pixel-parity`
- `git diff --check`

## 2 会话中的主要阻点/痛点

1. **历史 PASS 掩盖了 scenario wrapper 漂移**
   - **证据**：P0.016 live Playwright evidence 输出 `resumeId`，verify 仍查旧 `resumeVersionId`；P0.098 backend trigger 使用已删除 `resume_asset_id` 字段。
   - **影响**：runtime 代码已经迁移，但 BDD gate 仍可能 false-fail 或 false-green，导致交付状态无法信任。

2. **README / inventory 计数属于人工维护残留面**
   - **证据**：`openapi/fixtures/README.md` 仍写 60 operations 与 `/resume-versions/{resumeVersionId}/exports`，而当前 codegen 和 fixture gate 已是 43 operations / 43 fixtures。
   - **影响**：新成员或后续 agent 会从 README 获得错误 contract 事实，review 时也容易把旧路径当成可用 API。

3. **backend helper 同时踩中 auth contract 与 schema drift**
   - **证据**：P0.098 先因 auth challenge token 不是当前 6 位格式失败，修复后又暴露 `practice_plans.resume_asset_id` 已删除。
   - **影响**：单个 full-funnel gate 隐藏多个独立漂移点，需要逐层跑到真实 failure 才能定位。

4. **结构化数据断言用了字符串格式比较**
   - **证据**：P0.035 在 migration 恢复后失败于 JSONB 字符串精确比较；Postgres JSONB 合法地重排空格和序列化格式。
   - **影响**：测试失败不是业务语义错误，但会阻断场景闭环并误导排查方向。

## 3 根因归类

- **Scenario wrapper drift**：归类 **spec-plan / test**。多层 refactor 后没有把 `data/`、`expected-outcome.md`、`scripts/verify.sh` 和 trigger helper 纳入同一轮当前态审计。
- **Contract documentation drift**：归类 **README / docs**。fixture README 的 endpoint 计数和示例路径缺少由 codegen 或 fixture registry 派生的检查。
- **Backend e2e helper drift**：归类 **test**。full-funnel helper SQL 和 browser state 没有跟 schema rename 一起被负向搜索覆盖。
- **JSONB assertion drift**：归类 **test**。对结构化存储结果使用字符串比较，不符合数据库 JSONB 语义。

## 4 对流程资产的改进建议

- **删除 / 重命名 refactor checklist**：落点为 `plan-code-review` 或 `implement` 的 deep reconcile 入口，优先级 **high**。每次删除 route、operation、schema field 或 feature module 时，强制同时搜索 runtime、generated artifacts、fixtures、scenario data、verify scripts、README 计数、pixel/e2e specs 和 backend helper SQL。
- **Scenario verify 语义 gate**：落点为 `test/scenarios/README.md` 或 scenario skill，优先级 **high**。verify 脚本应优先断言当前字段语义和本轮 evidence；历史 marker 只能作为辅助。
- **Fixture README drift check**：落点为 `scripts/lint`，优先级 **medium**。将 README 中的 operation / fixture 计数和示例路径纳入现有 OpenAPI inventory 或 fixture validation。
- **JSONB / map 断言规范**：落点为 backend README 或测试 helper，优先级 **medium**。结构化 JSON 统一 parse 后比较，避免字符串序列化格式成为测试结果。

## 5 建议优先级与后续动作

下一步建议先用 `/plan-code-review frontend-resume-workshop/001-listing-routing-and-detail-readonly --fix` 或当前 D-20 owner plan 的等价入口，专门做一轮 **D-20 flat resume negative sweep hardening**：把本次修复过的 `resumeVersionId` / `resumeAssetId` / `getResumeVersion` / `resume_versions` 搜索范围固化到 owner checklist 和 lint gate，确保后续 D-14 parse 单次确认漏斗与 D-15 debrief 选一带二不会再次继承旧版本树口径。

备选路径是先落地 `scripts/lint` 的 fixture README drift check。它范围更小，但只能覆盖 OpenAPI/fixture 文档，不能覆盖 scenario wrapper 和 backend helper drift。
