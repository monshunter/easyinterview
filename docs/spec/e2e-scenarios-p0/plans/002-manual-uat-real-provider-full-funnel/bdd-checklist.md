# BDD Checklist

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-05-26

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.100 真实 provider 人工全漏斗 UAT

- [x] 创建 owner plan 文档并保留场景 ID `E2E.P0.100`
- [x] 准备 manual UAT 目录 `test/scenarios/manual-uat/full-funnel/`
- [x] 准备账号材料：UAT email、Mailpit URL、magic-link 验证步骤、cookie 检查和 cleanup
- [x] 准备输入材料：双语 JD、双语简历、作答样例、期望观察点
- [x] 准备真实环境 runbook：dev-stack、migrate、backend real provider、frontend real mode、secret redline、无 mock/stub 边界
- [x] 准备人工执行 checklist：全漏斗、AI provider evidence、隐私、legacy-negative、证据路径
- [x] 执行一次真实 provider manual UAT，记录脱敏证据
  <!-- verified: 2026-05-26 evidence=".test-output/manual-uat/full-funnel/evidence-20260526.md records ready report 019e6432-e38c-7acd-92d0-be1ca42386df, next-round session 019e6433-da8c-73a1-8ba0-b885f9c2dc94, and post-fix ready report 019e644f-a712-7812-a314-c677b98dac78 with answer_summary/report semantic checks and real provider task runs, without prompt/response bodies or secrets" -->
- [x] 完成 cleanup 或明确保留现场的人工确认
  <!-- verified: 2026-05-26 evidence="DELETE /api/v1/me via the UAT browser session returned privacy request 019e6435-0a9c-7e09-876c-8f3f2109b190 completed and privacy_delete job 019e6435-0a9c-7e0d-b785-bd2e1677126c succeeded; BUG-0106 records that current cleanup semantics leave users.email present" -->
