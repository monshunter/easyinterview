# Interview Resume Grounding and Round Progress 交付复盘报告

> **日期**: 2026-07-12
> **审查人**: Codex

**关联计划**: [Resume Asset Register, Parse and Listing](../spec/backend-resume/plans/001-asset-register-parse-and-listing/plan.md)、[Practice Plan and Session Orchestration](../spec/backend-practice/plans/001-plan-and-session-orchestration/plan.md)、[Practice Event Loop and Completion](../spec/backend-practice/plans/002-event-loop-and-completion/plan.md)、[TargetJob Import and Parse](../spec/backend-targetjob/plans/001-targetjob-import-and-parse-bootstrap/plan.md)、[Workspace and Interview Context](../spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md)
**关联 Bug**: [BUG-0162](../bugs/BUG-0162.md)、[BUG-0163](../bugs/BUG-0163.md)

## 1 复盘范围与成功证据

- 本次交付修复两个相互独立但在同一用户流程中连续出现的问题：真实简历没有进入 Practice 的有效证据上下文，以及第一轮完成事实没有投影为 TargetJob 当前下一轮。
- Resume parse 不再要求模型回显整份 Markdown；后端从完整原始输入生成持久化快照，结构化输出检查 `finish_reason=length`。Practice start/send 使用快照优先、严格当前简历绑定和 system/untrusted JSON 角色边界，空上下文 fail closed。
- 方案 A 已落地：`practice_plans.round_id/round_sequence` 持久化规范轮次身份，TargetJob Get/List 从当前绑定简历的 completed session/event 台账投影 `completedRounds + currentRound`；前端只消费后端 `practiceProgress`，不维护第二份业务状态。
- migration 000017 在隔离 PostgreSQL 完成 up → down → up，终态 version 17 clean；backfill dry-run/apply/rerun、歧义与非法记录 fail-closed 均有测试。
- P0.098 运行 `e2e-p0-098-20260712111826-75013` 通过：真实 completion 后 Workspace、Home、Parse 跨 reload 均进入 round 2，真实 CreatePlan POST/GET 返回 round 2 / sequence 2，cleanup 后无 seed 残留。
- `make test` 明确 exit 0：UI contract 48/48、Python 473 tests + 3919 subtests、全部 Go packages、frontend 112 files / 764 tests 通过；frontend typecheck/build、Go vet、OpenAPI/codegen/fixture、prompt lint、offline eval 27/27、docs/index/diff gates 通过。
- owner spec/plan/checklist 已原地修订并恢复 completed 生命周期，没有另建同主题 sibling plan；新增治理门禁明确除主题/外观/语言等展示偏好外，业务状态必须由后端持久化。

## 2 会话中的主要阻点/痛点

- “智能客服”问题表面像单纯 prompt hallucination，实际是输入、输出、持久化读取和 prompt trust boundary 四段耦合故障。
  - **证据**：数据库中的真实简历快照完整且无“智能客服”，旧 provider 输出恰好达到 2048 tokens 后解析失败，Practice 又只读取空 `structured_profile`；单独修改面试官文案无法恢复缺失证据。
  - **影响**：如果只修 prompt，会留下同样的无简历上下文；如果只增大 token，又会保留冗余 Markdown 回显和指令注入边界。
- 轮次进度最初看起来是前端刷新问题，但 Home/Workspace 已重新请求后端；真正缺口是 plan 无 round identity、read model 无台账投影。
  - **证据**：session completion/event 已落库，旧 TargetJob 只返回全局 latest ready plan；页面 reload 后仍回到第一轮。新增真实 DB/browser gate 证明方案 A 可跨请求恢复。
  - **影响**：仅在组件状态或 localStorage 推进会形成第二真理源，跨设备、重试与重复完成仍会漂移。
- Prompt 修复首次把 untrusted data 编码为 JSON 后，复核仍发现三条边界：runtime language 可能闭合 system tag，`finish_reason=length` 没有独立失败语义，完整 conversation history 又会把 assistant 自己的上一轮幻觉授权为候选事实。
  - **证据**：closing-tag 注入、连续 length 和“assistant 先误提智能客服”的 prompt contract/eval RED tests 先失败；改为安全 language tag、repair-once/repeat-fail，并把事实来源限制为 persisted resume + candidate-authored user message 后通过。
  - **影响**：仅给 resume/JD 字段加 `_json` 不足以证明整个 system message 的动态输入都安全；若不区分历史角色，一次模型幻觉还会在后续轮次滚雪球。
- 根级 `make test` 首轮暴露若干历史 Python gate 仍断言已删除的 practice mode/report assessment 或旧 helper。
  - **证据**：focused 实现测试已绿，但全量聚合测试在 migrations lint、out-of-scope 和 scenario-env synthetic marker 处失败；把断言迁移到 current contracts 后 473 + 3919 子测试通过。
  - **影响**：跨层合同重构如果只跑 touched package，会让过期治理测试在最终阶段才暴露，增加收口返工。
- 真实浏览器验收没有调用真实 provider 生成首问。
  - **证据**：P0.098 对 completion、TargetJob reload、plan create/read 使用真实 API/数据库，但仅拦截 session start 以避免把无关 AI opening 波动引入进度场景。
  - **影响**：当前证据能证明完整简历传输、角色边界和负向评测，不能保证真实模型在该简历上绝不产生无依据问题。
- 独立复核还指出 `resume.parse` 的解析指令与 raw resume 仍处在同一个 user role。
  - **证据**：`buildPromptMessages` 当前把模板渲染后的指令和 `{{resume_text}}` 一起提交；本次真实简历不含注入指令，Practice 又优先使用后端源快照，因此未构成本次两个故障的阻断项。
  - **影响**：恶意或意外包含指令式文本的简历可能影响结构化摘要；这是独立的防御纵深事项，不应与已关闭的上下文截断混为一谈。

## 3 根因归类

- 长文本输出合同与 downstream context precedence 不完整
  - **类别**：spec/plan
  - 原计划没有区分 1M 输入窗口、provider 输出预算、确定性源快照和 Practice 实际读取字段，也没有要求 `finish_reason=length` gate。
- 业务进度缺少后端事实模型
  - **类别**：AGENTS.md / spec-plan
  - 旧合同允许前端从 lifecycle status 推断轮次，没有声明 plan identity、completion ledger projection、Get/List parity 和浏览器 storage 禁区；本次已在治理规则和 owner plans 中原地修复。
- Prompt 动态值信任边界不完整
  - **类别**：spec/plan
  - 原 gate 关注模板文本与 schema，却没有逐一审计 system message 中所有 runtime substitution，也没有把 prompt data 当作不可信内容进行角色隔离，更没有区分 candidate-authored 与 assistant-authored history 的事实权限。
- 已删除合同的 Python consumer 漂移
  - **类别**：spec/plan
  - current code/场景已迁移，但聚合 lint 的 synthetic fixture 和 public helper 断言没有进入同一 consumer matrix；完整 `make test` 才发现。
- 本次未运行真实 provider 的简历针对性验收
  - **类别**：无需仓库改动（当前交付边界）
  - 进度持久化场景刻意隔离 AI opening；真实 provider 质量验证应作为独立、可重复且有成本/隐私控制的 UAT，而不是伪装在 deterministic BDD 中。
- Resume parse 原始文本仍与解析指令同 role
  - **类别**：spec/plan
  - 当前 Practice grounding 已不依赖 provider 回显 Markdown，但 parser 自身仍可进一步把可信解析策略与不可信简历正文分成 system/user 两个 role。

## 4 对流程资产的改进建议

- 对所有长文本 structured-output feature，owner checklist 必须同时列出 input byte boundary、禁止静默截断、output token budget、禁止冗余源文本回显、`finish_reason=length`、完整源快照和 downstream precedence。
  - **落点**：prompt-rubric / backend-resume 后续 spec-plan
  - **优先级**：high
- 保持本次新增的业务状态持久化门禁：除 display preferences 外，任何进度、选择、审批、草稿与流程状态都必须有 backend persistence、OpenAPI read model、幂等/重试和跨 reload negative gate。
  - **落点**：AGENTS.md 与各业务 owner spec-plan
  - **优先级**：high
- 对“台账投影进度”保留 round identity、当前绑定简历、completion provenance、out-of-order gap、非连续 sequence、Get/List parity 和 exact ready-plan reuse 的测试矩阵。
  - **落点**：backend-practice / backend-targetjob / frontend-workspace-and-practice spec-plan
  - **优先级**：high
- Prompt contract review 应枚举所有进入 system role 的动态值，并要求安全枚举/tag 或 JSON 编码；history 必须声明各 role 的事实权限，assistant-only claim 只能维持语境、不得建立履历事实；structured message 还需测试首次 length repair 与连续 length fail-closed。
  - **落点**：prompt-rubric registry 与相关 feature spec-plan
  - **优先级**：medium
- 合同删除或大规模重构时，在 focused RED/GREEN 之前和收口时各跑一次根级聚合测试，确保 Python lint、synthetic fixtures、生成物和场景 verifier 都在 consumer matrix 内。
  - **落点**：implement/tdd 执行约定或对应 plan checklist
  - **优先级**：medium
- 新增一个受控的真实 provider UAT，使用本次提供的简历，正向断言首问引用 Ferry / GitOps / Kubernetes 等真实事实，负向断言不出现“智能客服”或其他不存在项目，并保存脱敏 prompt/task-run 证据。
  - **落点**：e2e-scenarios-p0 / prompt-rubric 后续 spec-plan
  - **优先级**：high
- 在独立的 Resume parse 防御纵深修订中，把解析策略放入 system role，把 raw resume 作为 JSON 编码的不可信 user data，并增加 closing-tag / instruction-like resume 负例；不要与当前已验证的 Practice 修复混成无边界扩展。
  - **落点**：backend-resume / prompt-rubric 后续 spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

1. 下一步最高优先级是运行真实 provider 的简历针对性 UAT：使用同一份简历创建全新 session，检查首问与至少一轮追问的证据来源，并保留“智能客服”负向断言和脱敏 task-run 记录。
2. 持续保留 P0.098 的真实 completion → Home/Workspace/Parse reload → next-round plan create/read 链路；任何 plan、session、TargetJob 或轮次 UI 变更都必须重跑。
3. 后续长文本 AI feature 先用统一 checklist 审计输入窗口与输出预算，避免再次把“模型支持长上下文”误解为整个应用链路不会截断。
4. Resume parse 的 system/user role 分层可作为下一轮 medium-priority 安全加固；它不阻塞当前真实简历 grounding 与轮次持久化交付。
5. 根级聚合测试的前置运行可以在下一次跨层合同重构中验证收益；本次 stale consumers 已修复，不需要为当前交付再修改治理资产。
