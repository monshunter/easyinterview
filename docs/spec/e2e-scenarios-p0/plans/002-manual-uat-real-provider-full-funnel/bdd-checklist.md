# BDD Checklist

> **版本**: 2.10
> **状态**: active
> **更新日期**: 2026-07-13

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.100 真实 provider hybrid 全漏斗 UAT

- [x] 创建 owner plan 文档并保留场景 ID `E2E.P0.100`
- [x] 准备 hybrid e2e 场景目录 `test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/`
- [x] 准备账号材料：UAT email、Mailpit URL、email-code 验证步骤、cookie 检查和 cleanup
- [x] 准备输入材料：双语 JD、双语简历、作答样例、期望观察点
- [x] 准备真实环境 runbook：dev-stack、migrate、backend real provider、frontend real mode、secret redline、无 mock/stub 边界
- [x] 准备人工执行 checklist：全漏斗、AI provider evidence、隐私、out-of-scope-negative、证据路径
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

## E2E.P0.100 Phase 8 grounded report content reliability

- [x] 消费 F3/004 在 corrected evalkit completion gate 后重发的 `REPORT_RUBRIC_V020_PASS` / `REPORT_CONTEXT_AWARE_EVAL_PASS` 与 F3/002 `REPORT_PROMPT_V020_PASS`；场景只调用 registered completion/grade path，不复制 output schema、repair prompt、rubric 或 evaluator。
  <!-- verified: 2026-07-13 evidence="prompt/rubric/resolved single-source gates and exact offline eval 28/28 PASS" -->
- [x] HISTORICAL-SUPERSEDED Product generation durable max4/pre-call reserve/runner requeue/crash-cap evidence is retained only as former-contract audit evidence.
- [x] Evalkit generation/judge independent max4：judge provider/protocol invalid retry；valid negative typed terminal no retry；usage/latency aggregate。
- [x] 准备完整作答、部分作答、短答、待追问、提示注入五类互相隔离的当前轮输入与输出 manifest。
- [x] 让 context-aware judge 与人工抽查逐项记录 `supported` / `partial` / `unsupported`，并核对“事实 -> 判断 -> 行动”；partial 仅在显式证据受限且不驱动负面 readiness/action 时允许。
- [x] 验证 exact focus decision table；focused multi-code 首 action 每 code 一个 directly cited 分号短片段，English 按空白分词 `<=24 words`、zh-CN 按 Unicode code point `<=64 code points`，umbrella-only 无效；`ReportNextAction.label.maxLength=200` code points仅是 wire/schema malformed-output fuse，不能替代 UX gate；18/52只作targeted-repair内部生成余量。
  <!-- verified: 2026-07-13 evidence="current P0.100 action audits pass 24/64 semantic limits; ui-design contract 54/54 and focused Playwright 34/34 separately prove exact en24/zh64 desktop/mobile wrapping" -->
- [x] 验证每轮full-validator scope、labels-only不改其它字段、whole replacement、attempt2/3/4成功与attempt4 fail-close；manifest attempt_count/retry_count/reason/scope连续且无raw数据。
- [x] 验证product每次invocation独立initial+3、10s/20s/40s、返回销毁/新动作清零且async job attempts不影响product attempt；lease takeover仍证明stale worker零report/outbox/audit/job副作用。
  <!-- verified: 2026-07-13 evidence="P0.058 v3 owns product action/reset/separation proof; current PostgreSQL/race fencing gates PASS" -->
- [x] 验证frontend timer/in-flight hidden/blur均恢复n+1，单run<=49且无重复/并发请求。
  <!-- verified: 2026-07-13 evidence="post-L2 focused Playwright suite 34/34 PASS before the fresh complete scenario rerun" -->
- [x] 产品验收的五类固定代表样本均独立运行且失败后不替换；所有已生成最终输出先通过机械合同，每类仍按既有threshold/zero-tolerance规则判定，至少4/5通过。关键3x、11/11与blind review只属于更严格P0.100诊断。
  <!-- verified: 2026-07-13 run="e2e-p0-100-20260713T101214Z-59381" evidence="mechanical9/9; semantic8/9; fixed categories4/5; strict P0.100 FAIL on injection summary and blind audit not run" -->
  - [x] Preserve run36625 generic-replay focused evidence and run35103 only as the historical strict PASS for its exact prompt；final evidence uses run59381。
  - [x] Preserve run80338 as historical full-validator escape；run59381 proves all nine emitted finals pass the corrected validator。
  - [x] Preserve run59906/run75753 focused regressions without using either as current matrix evidence。
  - [x] Preserve run25849 as aborted/not-PASS；do not promote it or historical run35103 over final run59381。
  - [x] Preserve `e2e-p0-100-20260713T030100Z-35622` as aborted at7/11 due L2 findings and not PASS；fresh post-L2 run 35103 supersedes its pending restart。
- [x] 证明 `contract-safe`、schema valid 或页面可渲染不能单独形成内容可靠性 PASS。
- [x] 证明 evidence 只含当前 run/resource IDs、样本类别、逐项 verdict、阈值和 provider 摘要，不含原始上下文、完整 prompt/response、cookie、email code 或 secret。
- [x] 最终run以固定五类4/5、completed attempts 8/9语义与9/9机械摘要闭环；严格runner的FAIL与未执行blind audit保持可见，不向P0.099提供output-digest前置条件。
  <!-- verified: 2026-07-13 evidence="run59381 redacted failure evidence and privacy cleanup PASS" -->
- [x] BDD-Gate: product acceptance consumes action-local/reset/separation、机械100%、fixed-five至少4/5与脱敏证据；strict P0.100 remains FAIL unless its own 11/11+blind-review contract passes。P0.099独立验收UX。
  <!-- verified: 2026-07-13 run="e2e-p0-100-20260713T101214Z-59381" evidence="product acceptance met; strict diagnostic not promoted" -->
