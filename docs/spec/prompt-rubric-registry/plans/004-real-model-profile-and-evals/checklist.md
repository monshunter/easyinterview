# F3 Real Model Profile and Evals Checklist

> **版本**: 1.24
> **状态**: completed
> **更新日期**: 2026-07-13

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
