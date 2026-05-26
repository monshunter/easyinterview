# Manual UAT Companion

> **Status**: active
> **更新日期**: 2026-05-26
> **Owner plan**: [`e2e-scenarios-p0/002-manual-uat-real-provider-full-funnel`](../../docs/spec/e2e-scenarios-p0/plans/002-manual-uat-real-provider-full-funnel/plan.md)

本目录承载人工验收材料包与 runbook。它是 [`test/scenarios/e2e/`](../e2e/) 自动化 BDD 套件的 companion，不是标准 runner 套件：这里不要求每个目录都有 `setup.sh` / `trigger.sh` / `verify.sh` / `cleanup.sh`，也不登记到 `e2e/INDEX.md` 作为自动化场景。

## 1 边界

- 自动化真后端回归：仍由 `E2E.P0.098` / `E2E.P0.099` 承接。
- 人工真实 provider 验收：由 `E2E.P0.100` 和本目录承接，目标是真实前端、真实后端、真实 PostgreSQL、真实 AI LLM provider。
- 本目录中的 JD / 简历 / 作答文本是人工输入材料，不是 mock response fixture。
- `APP_ENV=test`、deterministic stub AI、fixture-backed frontend mock transport、`Prefer: example=<scenario>` 和 P0.099 test server 都不能作为真实 provider UAT 完成证据。

## 2 目录

| 目录 | 内容 | Owner |
|------|------|-------|
| [`full-funnel/`](./full-funnel/) | Home -> Parse -> Workspace -> Practice -> Generating -> Report -> next_round 的真实 provider 人工验收材料包 | `E2E.P0.100` / plan 002 |

## 3 使用顺序

1. 先读 owner plan，确认当前 Phase 是否已经完成 Mailpit 本地邮箱登录边界；`test/scenarios` 不新增 Go / `backend/cmd` helper。
2. 复制并本地填写 `full-funnel/env-template/dev-real.env.example`，真实 key 只放在本地未跟踪文件。
3. 按 `full-funnel/README.md` 启动 dev-stack、migrate、backend、frontend，并通过 Mailpit 完成 synthetic 邮箱登录。
4. 使用 `full-funnel/materials/` 中的 synthetic JD / resume / answer materials 走完整漏斗。
5. 用 `full-funnel/checklist.md` 的副本记录人工结果和 `.test-output/manual-uat/full-funnel/` 证据路径。
