# F3 Real Model Profile and Evals Checklist

> **版本**: 1.26
> **状态**: active
> **更新日期**: 2026-07-21

**关联计划**: [plan](./plan.md)

## Phase 1: Judge contract

- [x] 1.1 `Judge` 返回逐维度结构化 score；unsupported/missing profile 使用 `FailClosedJudge`。
- [x] 1.2 `judge.default` 通过独立 capability/provider routing，non-thinking strict JSON wire 与 reasoning-only fail-close 受 contract test 保护。

## Phase 2: Offline eval

- [x] 2.1 Promptfoo/evalkit 从 registry single source 解析 prompt、rubric 和 output schema。
- [x] 2.2 当前离线 case 集零网络执行并覆盖多语言、readiness、focus/action、safety 与 invalid outputs。
- [x] 2.3 prompt/rubric/profile coverage 与 drift lint 通过。

## Phase 3: Validation and repair

- [x] 3.1 Generation/judge 使用独立 max4 budgets，且每轮重新完整校验。
- [x] 3.2 仅 action-label violation 走 targeted repair；其它或 mixed invalid 走 whole-report repair。
- [x] 3.3 judge 只重试 provider/protocol/schema invalid；valid negative 立即终止。

## Phase 4: Recorded-output audit and privacy

- [x] 4.1 固定 recorded-output eval 分别报告机械与语义结果，不把不完整矩阵或 skipped blind review 宣称为 PASS。
- [x] 4.2 同模型 generation/judge 结果需要独立 reviewer；raw packet 使用 OS 私有临时文件并清理。
- [x] 4.3 manifest 只保留脱敏 usage/latency、attempt/retry/reason/scope、结构计数与 digest。
- [x] 4.4 真实 provider smoke 不属于完成条件，不写入 BDD/E2E 或配置 gate。

## Phase 5: 分层收口

- [x] 5.1 BDD-N/A: 本 plan 是内部 prompt/rubric/judge/eval owner，不创建 BDD 文件，也不进入 `test/scenarios/e2e/`。
- [x] 5.2 开发中 focused tests 与 offline gates 通过；阶段收口从仓库根执行 `make test`。
- [x] 5.3 context/docs/index/privacy/diff gates 通过。

## Phase 6: 激活证据恢复

- [x] 6.1 RED: 运行 `TestV020ActivationOwnerMarkersReady`，确认 completed-plan 压缩删除了 004 verified marker，而下游激活 preflight 仍消费该合同。
  <!-- verified: 2026-07-16 command="focused owner-marker preflight" result="FAIL at the first missing 004 owner marker" root_cause="2026-07-15 owner compression removed the verified marker comments but retained the downstream preflight consumer" -->
- [x] 6.1a RED-GREEN: preflight 只接受 verified comment 的显式 `marker=<name>` 属性；失败说明或普通 evidence 文本提到 marker 不得误判为 PASS。
  <!-- verified: 2026-07-16 red="helper missing caused focused build failure" green="four cases pass: explicit attribute accepted; failure mention, evidence mention and unchecked text rejected; activation preflight still fails at the first absent 004 marker" -->
- [x] 6.2 CURRENT-GATE: 实际重跑 28-case `make eval-offline`、prompt/rubric/profile/hardcode lint、evalkit/registry/judge focused tests与隐私边界检查；记录本次结果，不复用历史 PASS。
  <!-- verified: 2026-07-16 commands="make eval-offline; make lint-prompts lint-rubrics lint-ai-profile-coverage lint-prompts-hardcode; focused eval/evalkit/registry/profile/provider/bootstrap/judge/observability Go tests; config/evals generated-output and dirty-truth-source redlines" result="drift-check 28 cases and 9 resolved prompts; offline no-network grading 28; Promptfoo 28 passed, 0 failed, 0 errors; lints and all focused Go packages pass; no config/evals truth-source drift or .generated output" -->
- [x] 6.3 MARKER-GATE: 仅在 6.2 全绿后重新写入 `REPORT_RUBRIC_V020_PASS` 与 `REPORT_CONTEXT_AWARE_EVAL_PASS` verified marker，并使 `TestV020ActivationOwnerMarkersReady` 转绿。
  <!-- verified: 2026-07-16 marker=REPORT_RUBRIC_V020_PASS basis="current 28-case single-source offline eval, rubric/profile/hardcode lint and focused judge contract gates all pass" -->
  <!-- verified: 2026-07-16 marker=REPORT_CONTEXT_AWARE_EVAL_PASS basis="five grounded report cases, full validator, typed judge retry/content rejection and redacted audit tests pass in the current owner run" -->
- [x] 6.4 CLOSEOUT: 运行 context/docs/index/diff 与根 `make test`；恢复 plan/checklist `completed` 生命周期并保留 marker。
  <!-- verified: 2026-07-16 evidence="activation preflight and explicit-marker parser PASS; root make test PASS with 564 Python tests, 4481 subtests, all Go packages, and frontend 126 files/1004 tests; build, owner contexts, lint-config, docs/index and git diff checks PASS" -->

## Phase 7: Practice interviewer identity rubric and eval

- [x] 7.1 RED: rubric/eval lint rejects the missing v0.3 `role_identity` dimension and missing TargetJob-versus-Resume identity cases.
  <!-- verified: 2026-07-21 method=tdd-red evidence="focused evalkit count test expected 32 but implementation required 28; real-suite tests found 28 total, seven unpinned Practice cases, seven missing role_identity scores and all four required identity cases absent" -->
- [x] 7.2 GREEN: add the immutable practice v0.3 rubric, normalize weights, and update every Practice case with an exact dimension score set.
  <!-- verified: 2026-07-21 method=tdd-green evidence="practice v0.3 rubric weights sum to 1.0 with role_identity=0.4; all eleven Practice cases pin v0.3 prompt/rubric and contain the exact four-dimension score set; focused evalkit/eval tests and prompt/rubric lint pass" -->
- [x] 7.3 EVAL-GATE: named target, anonymous target, Resume-company impersonation weak negative and assistant-history correction cases pass the single-source offline suite together with all existing cases.
  <!-- verified: 2026-07-21 evidence="make eval-offline: single-source drift-check 32 cases/9 resolved prompts; deterministic offline grading 32; Promptfoo 32 passed, 0 failed, 0 errors; existing resume-grounding cases retained" -->
- [x] 7.4 OWNER-MARKER: only after current rubric/eval/lint gates pass, emit verified `PRACTICE_INTERVIEWER_IDENTITY_V030_PASS`; F3 `002` activation preflight must reject failure prose or ordinary text that merely mentions the marker.
  <!-- verified: 2026-07-21 marker=PRACTICE_INTERVIEWER_IDENTITY_V030_PASS basis="current prompt/rubric lint, exact four-dimension Practice suite, evalkit registry resolution and 32-case offline/Promptfoo gates pass; explicit-marker preflight first failed while this marker was absent" -->
- [x] 7.5 CLOSEOUT: focused judge/registry/eval tests, root `make test`, context/docs/index/privacy/diff gates pass before restoring `completed`.
  <!-- verified: 2026-07-21 method=full-closeout evidence="focused eval/evalkit/registry and privacy/lint gates PASS; make eval-offline drift/offline/Promptfoo 32/32; make test PASS Python 626/4628, Go all, frontend 137/1126; make build/lint/docs-check/context/index/diff PASS; live completion 5/5 and live judge parser limitation recorded separately" -->
