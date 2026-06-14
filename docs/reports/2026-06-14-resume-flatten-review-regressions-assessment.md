# Resume Flatten Review Regressions 交付复盘报告

> **日期**: 2026-06-14
> **审查人**: Codex (GPT-5)

## 1 复盘范围与成功证据

本次交付范围是修复 reviewer 对 D-20 resume flatten / D-17 JD Match removal patch 提出的 8 条 runtime、contract、migration、privacy 和 frontend persistence 回归问题，并补齐对应测试和 closeout gate。

已闭环的范围：

- Job polling ownership query 不再引用 dropped resume tables。
- OpenAPI diff baseline、diff config 与 README inventory 同步到 43 operations。
- Resume PATCH display-name-only 不再清空 structured profile。
- Privacy hard delete 会删除 resume parse / tailor async job payload 与 result。
- Resume parse job resource type 回到 `resume_asset` contract。
- Accepted rewrite 保存后会改写 persisted bullets。
- JD Match retired async jobs 与 guided resume rows 在 migration 收窄 constraint 前完成 cleanup。

通过证据：

- `make openapi-diff`
- `make lint-openapi`
- `make validate-fixtures`
- `make codegen-check`
- `go test ./backend/...`
- `make migrate-check`
- temporary database migration verification for retired JD Match jobs and guided resumes
- `pnpm --filter @easyinterview/frontend typecheck`
- `pnpm --filter @easyinterview/frontend test`
- `pnpm --filter @easyinterview/frontend build`
- `git diff --check`
- `make docs-check`
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`

## 2 会话中的主要阻点/痛点

- **OpenAPI baseline drift made the diff gate red after operation deletion**
  - **证据**：当前 `openapi/openapi.yaml` 是 43 operations，但 baseline 仍是旧 60-operation artifact，`openapi/diff-config.yaml` 还写 48。
  - **影响**：breaking-change gate 不能作为 review closeout 的可信信号。

- **Schema deletion fallout escaped runtime ownership and privacy paths**
  - **证据**：job polling 仍引用 `resume_tailor_runs` / `resume_assets`；privacy delete 未清理 `async_jobs.payload/result` 中的 resume tailor data。
  - **影响**：migration 后 job polling 会硬失败，privacy delete 会留下用户材料派生数据。

- **Partial update semantics were under-specified**
  - **证据**：display-name-only PATCH 经 service 填充 `{}` 后被 repository 无条件写入 structured profile。
  - **影响**：允许的轻量编辑会造成 parsed / edited profile 数据丢失。

- **Migration contract tests did not cover valid legacy rows**
  - **证据**：历史合法 `source_type='guided'` 和 JD Match async job rows 会在新 check constraint 前失败。
  - **影响**：真实开发库或 UAT 库可能无法迁移。

## 3 根因归类

- **OpenAPI baseline ownership gap**
  - **类别**：spec-plan / README
  - Baseline freeze、operation inventory 和 deletion patch 没有作为一个原子 contract gate 维护。

- **Runtime SQL negative search gap**
  - **类别**：spec-plan / test
  - 删除表后只覆盖 route / fixture 表层，未覆盖 owner query、privacy cleanup 和 JSON payload references。

- **PATCH field-presence contract gap**
  - **类别**：spec-plan / no repo change needed
  - 本次已用 tests 固化字段存在语义，暂不需要改流程资产。

- **Migration historical-row fixture gap**
  - **类别**：test
  - Migration contract tests 缺少“历史合法行 + constraint narrowing”的 fixture。

## 4 对流程资产的改进建议

- **Deletion checklist 扩展为 runtime SQL / JSON payload sweep**
  - **落点**：owner spec-plan gate 或 `plan-code-review` deep reconcile checklist
  - **优先级**：high

- **OpenAPI operation deletion 必须同 patch freeze baseline**
  - **落点**：`openapi/README.md` 或 OpenAPI owner plan gate
  - **优先级**：high

- **Migration narrowing tests 增加 valid legacy row fixtures**
  - **落点**：backend migration test conventions
  - **优先级**：medium

- **Privacy delete review gate 增加 payload/result retention check**
  - **落点**：privacy runner owner plan 或 backend README
  - **优先级**：medium

## 5 建议优先级与后续动作

下一步最高价值动作是对 D-20 resume flatten owner plan 做一轮 targeted hardening：把 `resume_assets`、`resume_tailor_runs`、`resume_versions`、`resumeVersionId`、`resumeAssetId`、`source_type='guided'`、`jd_match_agent_scan` 和 `jd_match_search` 的 runtime SQL / migration / privacy / OpenAPI baseline negative sweep 固化为 plan gate。

备选动作是先补一个较小的 OpenAPI deletion baseline checklist，确保后续 operation 删除不会再留下 stale baseline。
