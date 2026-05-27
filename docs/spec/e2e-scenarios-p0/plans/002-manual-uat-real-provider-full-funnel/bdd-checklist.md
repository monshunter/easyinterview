# BDD Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-05-27

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.100 真实 provider hybrid 全漏斗 UAT

- [x] 创建 owner plan 文档并保留场景 ID `E2E.P0.100`
- [x] 准备 hybrid e2e 场景目录 `test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/`
- [x] 准备账号材料：UAT email、Mailpit URL、magic-link 验证步骤、cookie 检查和 cleanup
- [x] 准备输入材料：双语 JD、双语简历、作答样例、期望观察点
- [x] 准备真实环境 runbook：dev-stack、migrate、backend real provider、frontend real mode、secret redline、无 mock/stub 边界
- [x] 准备人工执行 checklist：全漏斗、AI provider evidence、隐私、legacy-negative、证据路径
- [x] 执行一次真实 provider hybrid UAT，记录脱敏证据
  <!-- verified: 2026-05-26 evidence=".test-output/e2e/p0-100-real-provider-full-funnel-hybrid/evidence-20260526.md records ready report 019e6432-e38c-7acd-92d0-be1ca42386df, next-round session 019e6433-da8c-73a1-8ba0-b885f9c2dc94, and post-fix ready report 019e644f-a712-7812-a314-c677b98dac78 with answer_summary/report semantic checks and real provider task runs, without prompt/response bodies or secrets" -->
- [x] 完成 cleanup 或明确保留现场的人工确认
  <!-- verified: 2026-05-26 evidence="DELETE /api/v1/me via the UAT browser session returned privacy request 019e6435-0a9c-7e09-876c-8f3f2109b190 completed and privacy_delete job 019e6435-0a9c-7e0d-b785-bd2e1677126c succeeded; BUG-0106 records that current cleanup semantics leave users.email present" -->
- [x] 迁移为标准 e2e hybrid 场景目录 `test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/`
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="contract test confirms standard scenario directory, data files, and scripts for E2E.P0.100" -->
- [x] 登记 `E2E.P0.100` 到 `test/scenarios/e2e/INDEX.md`，执行方式为 `hybrid`
  <!-- verified: 2026-05-27 command="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q" evidence="contract test confirms INDEX registration as hybrid Ready" -->
- [x] Agent-first 脚本可生成 `MANUAL_REQUIRED` 或 PASS result artifact
  <!-- verified: 2026-05-27 command="bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/trigger.sh && bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/verify.sh && jq . .test-output/e2e/p0-100-real-provider-full-funnel-hybrid/result.json" evidence="AI Agent first-run scripts produced MANUAL_REQUIRED result artifact when local real-provider env file was absent" -->
