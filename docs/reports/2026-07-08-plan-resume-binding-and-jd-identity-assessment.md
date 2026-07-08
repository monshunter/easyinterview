# Plan Resume Binding and JD Identity 交付复盘报告

> **日期**: 2026-07-08
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `进入规划` 丢失已绑定简历上下文，以及 `target.import.parse` 有时丢失 JD 公司 / 岗位身份字段。
- Owner 文档：`frontend-workspace-and-practice/001`、`backend-targetjob/001`、`backend-practice/001` 已原地重开并恢复 completed。
- Bug 记录：[BUG-0144](../bugs/BUG-0144.md)、[BUG-0145](../bugs/BUG-0145.md) 已建档。
- 成功证据：
  - `pnpm --filter @easyinterview/frontend typecheck`
  - `pnpm --filter @easyinterview/frontend test src/api/frontendOwners.realApiMode.test.ts src/app/screens/workspace/WorkspaceEmptyState.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceHandoff.test.tsx src/app/screens/workspace/WorkspaceStartPractice.test.tsx`
  - `cd backend && go test ./internal/targetjob ./internal/api/practice ./internal/practice ./internal/store/practice -count=1`
  - `cd backend && go test ./cmd/api -run 'TestBuildTargetJobRuntimeWiresDrainerAndAIClient|TestE2EP0010HTTPTextImportParseReady|TestE2EP0022PracticePlanBaselineCreateAndRead' -count=1`
  - `make lint-prompts`, `make validate-fixtures`, `make lint-openapi`, `make docs-check`

## 2 会话中的主要阻点/痛点

- Plan list UI 的数据源是 `TargetJob`，真实绑定源却在 `practice_plans`。
  - **证据**：前端旧实现合成 `plan-${targetJobId}` 且不传 `resumeId`；后端 `practice_plans.resume_id` 已存在但 response contract 未暴露。
  - **影响**：需要同时改 OpenAPI、generated types、backend projection、PracticePlan response、frontend route 和 fixtures，不能只修一个组件。

- Prompt 输出 schema 与持久化 identity 字段脱节。
  - **证据**：`target.import.parse` 旧 schema 只要求 6 个分析字段；`parseAIResponse` 和 `CompleteParseSuccess` 没有 title/company 字段。
  - **影响**：单改提示词无法保证落库，必须同步 parser、store、seed migration、resolved eval prompts 和 prompt linter required paths。

## 3 根因归类

- TargetJob plan-list consumer 缺少 current practice binding 投影。
  - **类别**：spec-plan
  - **说明**：`frontend-workspace-and-practice/001` 的 card navigation gate 没有要求真实 `currentPracticePlanId` / `resumeId`。

- PracticePlan response contract 漏返回 `resumeId`。
  - **类别**：spec-plan
  - **说明**：`backend-practice/001` 已写入 `practice_plans.resume_id`，但 OpenAPI / generated / handler response 没闭合。

- Prompt schema gate 没覆盖 TargetJob identity。
  - **类别**：spec-plan
  - **说明**：`prompt-rubric-registry/002` 和 prompt linter required paths 只覆盖分析字段，未把 `title` / `companyName` 纳入 parse success contract。

## 4 对流程资产的改进建议

- 在 workspace plan-list owner gate 固化“列表入口不得合成 plan/resume id，必须使用 API 声明字段”。
  - **落点**：spec-plan
  - **优先级**：high

- 在 prompt output schema owner gate 中增加“持久化 identity 字段必须同时出现在 prompt schema、parser struct、store success path 和 seed/eval resolved prompt”检查项。
  - **落点**：spec-plan
  - **优先级**：medium

- 在 OpenAPI contract review 中增加“request required field 若是 persisted binding，response 是否也需要返回”的检查口径。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：把 `WorkspacePlanList` 的 current binding 投影要求保留在 `frontend-workspace-and-practice/001` 的 BDD / scenario gate 中，后续若新增 plan list API 时沿用同一断言。
- 次优先级：在下一轮 prompt-rubric owner review 中补一个跨层 negative gate，搜索 required prompt keys 与 consumer struct / persistence path 是否一致。
- 可延后：把 `PracticePlan.resumeId` response 规则补进 openapi-v1-contract 的 schema review checklist，避免同类绑定字段只进 request 不进 response。
