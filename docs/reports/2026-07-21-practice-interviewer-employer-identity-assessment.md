# Practice Interviewer Employer Identity 交付复盘报告

> **日期**: 2026-07-21
> **审查人**: Codex

**关联计划**: [backend-practice/001-plan-and-session-orchestration](../spec/backend-practice/plans/001-plan-and-session-orchestration/plan.md)、[prompt-rubric-registry/002-output-schema-contract](../spec/prompt-rubric-registry/plans/002-output-schema-contract/plan.md)、[prompt-rubric-registry/004-real-model-profile-and-evals](../spec/prompt-rubric-registry/plans/004-real-model-profile-and-evals/plan.md)
**关联 Bug**: [BUG-0197](../bugs/BUG-0197.md)

## 1 复盘范围与成功证据

- 交付范围：原地重开三个 owner plan，明确 Practice 面试官雇主身份来源，发布 immutable v0.3 prompt/rubric pair，补充身份评测、激活 preflight、registry/runtime tests 与可回滚 PostgreSQL migration。
- 离线证据：32 个 registry/offline cases 与 32/32 Promptfoo cases 通过；11 个 Practice cases 全部锁定 v0.3 exact coordinate 和包含 `role_identity` 的四维评分。
- 数据库证据：在一次性 PostgreSQL 数据库完成 `22 -> 23 -> 22 -> 23`，验证 v0.2/v0.3 精确切换、report v0.2 不变和 rollback 可用；临时数据库已删除。
- 真实模型证据：DeepSeek 对 5 个身份边界输入完成 5/5 有效 JSON 响应，未发生 Resume-company interviewer impersonation；匿名目标不自称具体公司，实名目标使用正确公司，历史错误自称能够纠偏。
- 仓库与运行时证据：根 `make test`、`make build`、`make lint`、`make docs-check`、context/index/diff gates 通过；backend 重新部署后端口监听和真实 API 行为符合预期。

## 2 会话中的主要阻点/痛点

- 原 prompt 只拥有“候选人事实 grounding”，没有面试官身份 owner。
  - **证据**：v0.2 同时接收 TargetJob、Resume 与 assistant history，却没有声明招聘方身份只能来自 TargetJob/round，也没有匿名目标公司的中立降级规则。
  - **影响**：模型可从更具体的简历公司补全自身身份，且错误输出进入 history 后可能自我延续。
- rubric 演进暴露了 fixture 对旧维度集合与旧 active version 的硬编码。
  - **证据**：RED 阶段既有 Practice cases 全部缺少 `role_identity`，evalkit 仍要求 28 cases，registry/cache/Practice tests 仍锁定 active v0.2。
  - **影响**：行为修复必须同步更新跨 owner 的 exact coordinates；遗漏任何一处都会 fail closed，但发现路径分散。
- migration parity 曾把文件命名约定误当成内容 owner。
  - **证据**：Go/Python lint 只扫描文件名包含 `prompt_rubric` 的 up migration；本次语义化文件名 `activate_practice_interviewer_identity_v030` 最初不会进入 parity 检查。
  - **影响**：合法 migration 可以绕过 prompt/rubric seed parity，降低激活 drift 的可见性。
- 真实 provider 的 completion 与 live judge 可用性不是同一件事。
  - **证据**：5 次 completion 均成功并满足人工可判定身份边界，但现有 live judge 在解析返回 JSON 时失败。
  - **影响**：可完成关键行为验收，却不能诚实声明 live judge PASS；需要保留 unavailable 边界。

## 3 根因归类

- 身份事实 owner、冲突顺序与匿名降级未进入原始行为合同。
  - **类别**：spec-plan
- exact prompt/rubric version 与 dimension set 分布在 registry、evalkit、runtime tests，版本演进时依赖当前 fail-closed tests 逐一暴露。
  - **类别**：spec-plan
- migration parity 按历史文件名 glob 选 owner，而不是检查所有 migration 内容。
  - **类别**：README
- live judge JSON 解析失败不属于本次身份 prompt 修复；当前 completion 证据已足够验收本缺陷，但 judge 诊断能力仍不可用。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 后续新增多来源 prompt 时，在 plan 的 context matrix 中显式列出“事实类型 -> 唯一 owner -> 冲突顺序 -> 缺失降级 -> 历史消息可信度”，覆盖身份、候选人事实和招聘方事实。
  - **落点**：spec-plan
  - **优先级**：high
- 将 prompt/rubric version bump 的受影响面固定为一个复核清单：immutable assets、active resolver/cache、caller fixtures、dimension exactness、eval count、migration activation/rollback 和 non-target feature isolation。
  - **落点**：spec-plan
  - **优先级**：medium
- migration lint 继续扫描全部 up migrations，再按内容识别 prompt/rubric seed；不要把语义 owner 绑定到文件名包含某个历史 shorthand。
  - **落点**：README
  - **优先级**：medium
- 单独诊断 live judge 的 JSON extraction/normalization；在修复前继续把 completion acceptance 与 judge verdict 分开报告。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- high：本次三个 plan 完成 lifecycle closeout 后，下一项最值得处理的是 live judge JSON extraction 诊断，由 prompt-rubric-registry 的 real-model eval owner 承接，先稳定复现再决定是否修代码。
- medium：下一次 prompt version bump 时执行统一的 exact-coordinate 影响面复核；如果再次出现手工漏项，再把复核清单固化为自动 drift gate。
- low：无需为本缺陷新增 HTTP API、业务表或浏览器 E2E；现有 owner contract、真实 provider completion、migration integration 与 runtime redeploy 已覆盖实际风险边界。
