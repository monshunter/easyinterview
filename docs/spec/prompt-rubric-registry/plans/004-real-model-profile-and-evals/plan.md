# F3 Real Model Profile and Evals

> **版本**: 1.26
> **状态**: active
> **更新日期**: 2026-07-21

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

维护 F3 当前离线评估闭环：`judge.default` 使用独立 judge capability，`LLMJudge` 按 rubric dimension 返回结构化分数，Promptfoo/evalkit 通过 registry single source 消费 prompt、rubric 和 output schema。

本 plan 不新增用户可见 UI、HTTP API 或端到端业务流程，也不要求真实 provider smoke 作为完成条件。

## 2 当前合同

- `FailClosedJudge` 是未配置或 unsupported 时的安全默认；`judge.default` 仅在 profile/provider 坐标完整时执行。
- report rubric v0.2、context-aware judge 与当前离线 case 集由本 plan 拥有；prompt/schema/parser 由 F3 002 拥有。
- Action `maxLength=200` 只作 malformed fuse；English 24 words、zh-CN 64 code points 是 semantic/UX 约束；targeted repair 使用内部 18/52 余量。
- Generation 与 judge 使用独立 max4 budgets；每轮 generation 完整校验并重新选择 targeted/whole scope；judge 只重试 provider/protocol/schema invalid，valid content rejection 立即终止。
- manifest 只保留脱敏 attempt/retry/reason/scope、usage/latency、结构计数和 digest；raw packet 只允许短期 OS 私有文件并在审计后删除。

## 3 质量门禁

- **Plan 类型**: `code-internal + eval + provider-profile`。
- **TDD**: 开发中运行 evalkit/registry/profile focused tests；阶段收口从仓库根执行 `make test`。
- **BDD**: 不适用。prompt、rubric、schema、judge 与 provider eval 都是内部代码/评估 gate；不得创建 BDD 文件，也不得进入 `test/scenarios/e2e/`。
- **替代 gate**: prompt/rubric lint、registry resolve、offline eval、active-budget floor lint、judge adapter contract、privacy cleanup、context/docs/diff checks 与根 `make test`。

## 4 实施范围

### Phase 1: Judge contract

维持 provider-neutral `Judge` / `[]Score` 合同、fail-closed 默认、judge capability/profile routing 与 non-thinking strict JSON provider wire。

### Phase 2: Offline eval single source

所有离线 case 通过 registry 解析当前 prompt/rubric/schema；fixture/golden 只承接离线确定性输入，任何 drift 必须 fail closed。

### Phase 3: Validation and repair

generation 每轮运行产品同源完整 validator；仅 action-label violation 允许 targeted merge，其余或 mixed invalid 进行 whole-report repair。repair 后再次完整校验，超出预算失败。

### Phase 4: Independent recorded-output audit

使用固定 recorded outputs 计算机械通过率、语义通过率、脱敏 manifest 和独立 reviewer 结果。generation/judge 同模型时 judge 不能单独自证；terminal negative 必须如实保留 FAIL。真实 provider 调用可作为人工诊断，但不属于本 plan gate。

### Phase 5: 收口

运行 offline/lint/profile/privacy gates、根 `make test`、context/docs/index 与 `git diff --check`。下游 UI 的真实浏览器验收由 UI owner 独立承担，不成为 F3 配置或 eval gate。

### Phase 6: 激活证据恢复

原地修复 completed-plan 压缩造成的 marker 漂移。先用 `TestV020ActivationOwnerMarkersReady` 复现 RED，并将 preflight 收紧为只接受 verified comment 中的显式 `marker=<name>` 属性，避免失败说明或历史文字仅提到 marker 就被误判为 PASS。再重新执行当前 28-case offline eval、prompt/rubric/profile/hardcode lint、evalkit/registry/judge focused tests与隐私边界检查。只有这些当前 gate 全部通过后，checklist 才能重新写入 `REPORT_RUBRIC_V020_PASS` 和 `REPORT_CONTEXT_AWARE_EVAL_PASS` verified marker；不得复制历史 marker 充当本次证据。最后运行聚焦 preflight、context/docs/index/diff gate 与根 `make test`，确认下游激活合同恢复。

### Phase 7: Practice interviewer identity rubric and eval

为 `practice.session.chat/v0.3.0` 增加独立 `role_identity` dimension，并重新归一化四个 dimension 权重。`role_identity` 明确评估：面试官招聘方只来自 TargetJob/round；Resume 公司只能作为候选人履历；匿名/模糊目标公司不得猜名；assistant history 中的错误自称不得延续。

在 single-source recorded/offline suite 中加入至少四类 case：目标公司与 Resume 公司不同的强正例、匿名目标公司不报公司名的强正例、把 Resume 公司说成面试官雇主的弱反例、前序 assistant 身份错误后的纠偏正例。所有既有 Practice grounding case 同步补齐新 dimension score，继续零网络可重复执行；配置可用时另做真实 provider 诊断，但不得用其替代 recorded/offline gate。只有当前 rubric/eval/lint 全部通过后，才在 checklist verified comment 中写入 `PRACTICE_INTERVIEWER_IDENTITY_V030_PASS`，供 F3 `002` 激活 preflight 精确消费。

## 5 验收标准

- offline eval 零网络、可重复并从 registry single source 解析。
- offline/recorded eval 对 unsupported、invalid 或不完整 blind review fail closed。
- 不在日志、manifest、DB 或文档证据中泄漏 prompt/transcript/raw model output。
- 不存在本 plan 专属 BDD 文件、E2E ID 或场景完成标记。
- 两个 v0.2 owner marker 与本次实际命令、28-case 结果绑定，并通过下游 marker preflight；后续 completed-plan 压缩不得删除仍有消费者的 verified marker。
- Practice v0.3 rubric/eval 能稳定拒绝 Resume-company interviewer impersonation，并允许匿名 TargetJob 使用不含公司名的自然面试官话术。

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-21 | 1.26 | 原地增加 Phase 7：为 practice v0.3 增加面试官身份来源 rubric dimension 与 named/anonymous/weak/history-correction 评测。 |
| 2026-07-16 | 1.25 | 原地恢复被 completed-plan 压缩删除的 v0.2 rubric/context-aware eval marker，并要求 marker 绑定当前 owner gate 重跑证据。 |
| 2026-07-13 | 1.24 | 压缩为当前 eval owner；删除旧场景编号和逐次 run 流水账，明确内部 eval 与真实 E2E 分层。 |
