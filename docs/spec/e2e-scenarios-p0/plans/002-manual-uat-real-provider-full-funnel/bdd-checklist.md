# BDD Checklist

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-05-27

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.100 真实 provider hybrid 全漏斗 UAT

- [x] 创建 owner plan 文档并保留场景 ID `E2E.P0.100`
- [x] 准备 hybrid e2e 场景目录 `test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/`
- [x] 准备账号材料：UAT email、Mailpit URL、email-code 验证步骤、cookie 检查和 cleanup
- [x] 准备输入材料：双语 JD、双语简历、作答样例、期望观察点
- [x] 准备真实环境 runbook：dev-stack、migrate、backend real provider、frontend real mode、secret redline、无 mock/stub 边界
- [x] 准备人工执行 checklist：全漏斗、AI provider evidence、隐私、legacy-negative、证据路径
- [x] 执行一次真实 provider hybrid UAT，记录脱敏证据
  <!-- verified: 2026-05-27 evidence=".test-output/e2e/p0-100-real-provider-full-funnel-hybrid/evidence.md records run_id e2e-p0-100-20260527T081532Z-38774, Mailpit auth, ready resume parse, ready target import, baseline practice, acknowledged ask_follow_up, ready report, next-round practice, 6 ai_task_runs, and screenshots without prompt/response bodies or secrets" -->
- [x] 完成 cleanup 或明确保留现场的人工确认
  <!-- verified: 2026-05-26 evidence="DELETE /api/v1/me via the UAT browser session returned privacy request 019e6435-0a9c-7e09-876c-8f3f2109b190 completed and privacy_delete job 019e6435-0a9c-7e0d-b785-bd2e1677126c succeeded; BUG-0106 records that current cleanup semantics leave users.email present" -->
- [x] 迁移为标准 e2e hybrid 场景目录 `test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/`
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="contract test confirms standard scenario directory, data files, and scripts for E2E.P0.100" -->
- [x] 登记 `E2E.P0.100` 到 `test/scenarios/e2e/INDEX.md`，执行方式为 `hybrid`
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="contract test confirms INDEX registration as hybrid Ready" -->
- [x] Agent-first 脚本可生成 `MANUAL_REQUIRED` 或 PASS result artifact
  <!-- verified: 2026-05-27 command="bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/trigger.sh && bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/verify.sh && cat .test-output/e2e/p0-100-real-provider-full-funnel-hybrid/result.json" evidence="AI Agent first-run scripts produced PASS only after current RUN_ID evidence was present and redline-clean; missing real-provider/browser evidence remains MANUAL_REQUIRED by contract" -->
- [x] 统一使用 `deploy/dev-stack/.env`，不维护场景专属 `dev-real.env`
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="contract test confirms E2E.P0.100 trigger reads deploy/dev-stack/.env and rejects scenario-specific dev-real.env references" -->
- [x] 固化 RUN_ID / evidence redline BDD gate：`setup.env` 的 `RUN_ID` 必须匹配 `evidence.md` 的 `run_id`，`scan_evidence_redline` 必须在 `trigger.sh` 与 `verify.sh` 中拒绝 provider key、auth secret、session cookie、raw email code、prompt body、response body
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q && bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/trigger.sh && bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/verify.sh" evidence="owner docs and BDD docs expose RUN_ID/current-evidence and evidence redline gates; 12 contract tests pass; trigger/verify accepted current run evidence only after scan_evidence_redline passed" -->
- [x] 固化 env consumer gate：`env-setup.sh --with-migrations`、`env-redeploy.sh frontend` 与 `E2E.P0.100` trigger 都必须消费 `deploy/dev-stack/.env`，并拒绝场景专属 env、frontend fixture mock 或 `Prefer: example=<scenario>`
  <!-- verified: 2026-05-27 command="bash test/scenarios/env-setup.sh --with-migrations && cd frontend && set -a && . ../deploy/dev-stack/.env && set +a && pnpm build && python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="owner docs and BDD docs expose env consumer gate for setup migrations, frontend real-mode build, and trigger; 12 contract tests pass" -->
