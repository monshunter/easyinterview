# Manual UAT Real Provider Full Funnel 交付复盘报告

> **日期**: 2026-05-26
> **审查人**: Codex

**关联计划**: [002 Manual UAT Real Provider Full Funnel](../spec/e2e-scenarios-p0/plans/002-manual-uat-real-provider-full-funnel/plan.md)

## 1 复盘范围与成功证据

- 本次交付范围是 `e2e-scenarios-p0/002-manual-uat-real-provider-full-funnel` 的 L2 review/fix：真实 Mailpit 登录、真实 backend/frontend、真实 OpenAI-compatible provider、完整 practice/report/next-round UAT，以及相关 prompt、profile、runtime、文档和 BUG 记录收口。
- 成功证据包括：真实 provider UAT 生成 ready report `019e644f-a712-7812-a314-c677b98dac78`；对应 session `019e644f-78a9-7003-8417-65756ac9320f` 的 `answer_summary` 长度为 197；report/generated text 对 `did not answer` 的语义负向扫描为 0；`question_generate`、`hint_generate`、`followup_generate`、`report_generate`、`report_assessment` 均记录真实 provider/profile/model/latency/task-run evidence。
- 自动门禁包括：`go test` 覆盖 `backend/cmd/api`、AI profile、OpenAI-compatible adapter、AI registry、practice/store、review；frontend Vitest 覆盖 follow-up action renderer；`make lint-prompts`、`make lint-ai-profile-coverage`、`make lint-config`、`validate_context.py --target scenario`、`sync-doc-index --check`、`make docs-check`、`git diff --check`。
- BUG 收口包括：`BUG-0105` 记录并关闭真实 provider full-funnel 语义链路缺陷；`BUG-0106` 记录 privacy delete 已完成作业但仍保留 `users.email` 的剩余清理语义缺口。

## 2 会话中的主要阻点/痛点

- 真实 provider 比 stub path 暴露更多运行时边界。
  - **证据**：UAT 过程中需要补齐 dev CORS、JSON object response format、profile timeout budget 与 prompt/schema drift，stub-AI 自动化路径未覆盖这些行为。
  - **影响**：初始 happy-path 证明不足以代表人工真实联调，必须追加 runtime contract tests 与 prompt/profile gates。
- 报告生成缺少可用的候选人作答摘要，导致真实 provider 合理但错误地推断候选人未回答。
  - **证据**：修复前 report 文本出现未作答语义；修复后同类 UAT 的 `answer_summary` 被持久化，report 文本负向扫描为 0。
  - **影响**：报告质量不能只以 `ready` 状态判断，必须检查报告语义和上游 evidence 是否完整。
- 清理链路的“作业完成”与“PII 字段清除”之间存在语义缺口。
  - **证据**：`DELETE /api/v1/me` 返回完成的 privacy request/job，但 `BUG-0106` 记录当前数据库仍保留 UAT 用户 email。
  - **影响**：manual UAT 可以完成，但隐私 cleanup 不能被简单写成完全通过，需要后续 owner plan 修复。

## 3 根因归类

- Stub / fixture 场景与真实 provider UAT 的门禁层级不同。
  - **类别**：spec-plan
- prompt/schema/profile 的真实 provider 约束缺少跨链路哨兵。
  - **类别**：spec-plan
- report ready gate 只看生成状态，不看报告文本语义和上游 answer evidence。
  - **类别**：spec-plan
- privacy delete cleanup 的完成定义没有覆盖 `users.email` 等账户级字段。
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- 在后续真实 provider UAT plan 中固定两类语义 gate：持久化 `answer_summary` 非空，以及 report 文本不得出现“候选人未回答”等与 evidence 冲突的结论。
  - **落点**：spec-plan
  - **优先级**：high
- 为真实 provider full-funnel 继续保留 provider/profile/model/latency/task-run evidence，但明确禁止记录 prompt/response 明文、cookie、magic-link token 与 API key。
  - **落点**：README
  - **优先级**：high
- 将 `BUG-0106` 提升为下一轮 privacy cleanup owner 工作，补充数据库字段级断言，避免把 async job success 当作数据擦除完成。
  - **落点**：spec-plan
  - **优先级**：high
- 对 report/practice prompt registry 增加更稳定的 schema-drift 哨兵，尤其是真实 provider 会消费但 stub 场景不敏感的可选字段。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：启动 `BUG-0106` 对应的 privacy cleanup 修复计划，目标是让 `DELETE /api/v1/me` 的完成语义覆盖 UAT 用户 email 和相关账户字段的清除，并加入 SQL/handler 级回归测试。
- 次优先级：把 manual UAT runbook 的最终验收项扩展为“状态 + 语义 + 隐私”三层证据，不再只以 ready report 和 task-run 数量作为充分条件。
- 可延后：将真实 provider prompt/schema/profile 的 drift guard 抽成更通用 lint，减少未来新增 AI feature 时的手动补哨兵成本。
