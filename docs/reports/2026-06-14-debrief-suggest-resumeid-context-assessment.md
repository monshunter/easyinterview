# Debrief Suggest ResumeId Context 交付复盘报告

> **日期**: 2026-06-14
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-debrief/001-debrief-record-and-analysis` Phase 8 D-20 修复，闭环 `suggestDebriefQuestions` 从 fixture `resumeId` contract 到真实 backend handler/service/store、fixture parity 和 E2E.P0.063 scenario gate。
- 代码证据：`Repository.GetSuggestionContext` 在 request 携带 `ResumeID` 时按 `(user_id, resume_id)` 查询 `resumes.structured_profile`；service 将 resume summary 注入 AI prompt；API handler 将 generated `ResumeId` 映射到 domain request；cmd/api route gate 证明真实 route 被 session middleware 挂载。
- 测试证据：store Red tests 先失败，Green 后 focused tests 覆盖 store/service/API/cmd-api；P0.063 scenario wrapper 运行 store/service/API/cmd-api focused tests、`make validate-fixtures`、fixture `resumeId` marker 与 `resumeVersionId` 负向 gate。
- 文档证据：backend-debrief/001 plan/checklist/test/BDD/history 原地更新，plans/INDEX 恢复 completed；`make docs-check` 与 sync-doc-index check 零漂移。
- Bug 记录：新增 [BUG-0121](../bugs/BUG-0121.md)，记录 `resumeId` contract 只停在 fixture / generated 层、store 未读取 `resumes.structured_profile` 的根因与修复证据。

## 2 会话中的主要阻点/痛点

- Fixture 与 generated request 已是 `resumeId`，但真实 store 仍只查 target job。
  - **证据**：Red tests `TestStoreGetSuggestionContext_LoadsResumeStructuredProfile` / `CrossUserResumeNotFound` 初次失败，显示 ResumeSummary 为空且 missing resume 没有 fail-closed。
  - **影响**：Debriefs owner 可被旧 fixture parity 误判为完成，真实 AI prompt 没有 resume 上下文。
- P0.063 scenario wrapper 证据过窄。
  - **证据**：修复前 trigger 只运行 service/API focused tests；修复后补入 store、cmd/api、fixture validator 与 stale-field negative gate。
  - **影响**：场景 PASS 无法证明真实 persistence path 或 route 装载，只能证明 mock-like service/API 层。
- Owner plan 文档保留旧路径和旧字段口径。
  - **证据**：`bdd-plan.md` 仍引用 `resume_versions` 与 `resumeVersionId`，plan 仍引用旧场景目录 `p0-063-suggest-debrief-questions`。
  - **影响**：后续 L2 review 容易继续沿用历史语义，而不是当前 D-20 contract。

## 3 根因归类

- Contract rename 的实现审查没有沿 operation matrix 逐层反查。
  - **类别**：spec-plan
  - **说明**：Phase 8 缺少明确 operation matrix，把 fixture、generated request、handler/service/store、AI prompt 和 scenario gate 串成一个验收对象。
- Scenario verify 对 wrapper 本身的证据要求不足。
  - **类别**：spec-plan
  - **说明**：场景脚本存在，但只证明 focused test 执行，没有证明关键 contract 字段从 fixture 到真实 persistence path。
- 无需新增 AGENTS.md 或 skill 规则。
  - **类别**：no repo change needed
  - **说明**：现有 `/plan-code-review` 已要求 operation matrix、scenario wrapper 和 production caller 审查；本次修复是 owner plan/gate 落地不足，不是通用流程规则缺失。

## 4 对流程资产的改进建议

- 在 backend-debrief 后续计划中，把 D-20 类字段重命名的完成条件写成 operation matrix 行，而不是单独写 fixture 或 handler gate。
  - **落点**：spec-plan
  - **优先级**：high
- 对已有 Debrief scenarios 做一次轻量 sweep，确认 `verify.sh` 都至少证明目标测试名、package `ok`、PASS marker、no-op negative gate 和关键 contract 字段负向搜索。
  - **落点**：spec-plan
  - **优先级**：medium
- 保持 BUG-0121 与 PATTERNS.md 现有模式 4 / 5 的关联即可，暂不新增模式。
  - **落点**：no repo change needed
  - **优先级**：low

## 5 建议优先级与后续动作

- 最高优先级：继续用 `/plan-code-review backend-debrief/001-debrief-record-and-analysis backend --fix` 对剩余 Debrief operation 做同样的 operation matrix 反查，尤其是 session summary 可选上下文是否也仍停在声明层。
- 次优先级：在后续 debrief plan 002 设计时，把 `suggestDebriefQuestions` rate limit / quota 与真实 backend evidence gate 一起写入，而不是仅写产品风险。
