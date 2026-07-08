# Target Job Resume Binding 交付复盘报告

> **日期**: 2026-07-09
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `进入规划` 仍提示绑定简历的问题，采用方案 A 将 Home 导入 JD 时选择的简历持久化为 `target_jobs.resume_id`，并保持 `TargetJob.resumeId` 在 list/detail 响应中可恢复。
- Owner 文档：`backend-targetjob/001`、`frontend-home-job-picks-and-parse/001`、`frontend-workspace-and-practice/001` 已原地修订并恢复 `completed`。
- Bug 记录：[BUG-0144](../bugs/BUG-0144.md) 已更新为 target job-level binding 根因；[BUG-0145](../bugs/BUG-0145.md) 覆盖 JD identity prompt/schema 修复。
- 成功证据：
  - `make codegen-openapi`
  - `cd backend && go test ./internal/targetjob -count=1`
  - `cd backend && go test ./cmd/api -run 'TestE2EP0010HTTPTextImportParseReady|TestE2EP0011HTTPURLImportFetchAndParse|TestE2EP0012HTTPParseFailureRetryableAndNonRetryable|TestE2EP0013HTTPManualFormReady|TestE2EP0098FullFunnelImportToNextRound|TestE2EP0098CreatePracticePlanAcceptsEmptyFocusCodes' -count=1`
  - `pnpm --filter @easyinterview/frontend test src/api/targetJob.realApiMode.test.ts src/app/screens/home/HomeImport.test.tsx src/app/screens/home/HomeResumeSelection.test.tsx src/app/screens/home/HomeAuthGate.test.tsx src/app/screens/workspace/WorkspaceEmptyState.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceHandoff.test.tsx src/app/interview-context/InterviewContext.test.tsx`
  - `make lint-prompts`, `make validate-fixtures`, `make lint-openapi`
  - `test/scenarios/env-setup.sh --with-migrations`, `test/scenarios/env-redeploy.sh backend && test/scenarios/env-redeploy.sh frontend`, `test/scenarios/env-verify.sh`
  - Local HTTP smoke: `POST /targets/import` with `resumeId` returned 202, and `target_jobs.resume_id`, `GET /targets/{id}.resumeId`, `GET /targets.items[].resumeId` all matched before any `practice_plans` row existed.
  - `E2E.P0.018` scenario wrapper PASS; `sync-doc-index --check`, `make docs-check`, `git diff --check` PASS.

## 2 会话中的主要阻点/痛点

- 第一轮修复把绑定恢复点放在 latest ready `practice_plans` 投影上，但用户截图里的问题行没有 practice plan。
  - **证据**：本地 DB 里截图同款 TargetJob `analysis_status=ready`，`practice_plans` 表无对应行；修复前 `target_jobs` 也无 `resume_id`。
  - **影响**：Workspace 从列表重进时只能拿到 JD 身份，拿不到导入时选择的简历，继续落入 missing-resume 空态。

- Home 导入、Parse 路由、Workspace 列表三个 owner 对“绑定简历”的生命周期定义不一致。
  - **证据**：Home 已要求用户选择 ready 简历，但旧合同只把 `resumeId` 作为前端 route param；Workspace list/detail 需要后端返回该字段。
  - **影响**：只改前端导航或只改 practice plan response 都不足以修复刷新、重进、无 plan 行的主路径。

## 3 根因归类

- TargetJob 级业务事实没有持久化字段。
  - **类别**：spec-plan
  - **说明**：`backend-targetjob/001` 原合同没有声明 `ImportTargetJobRequest.resumeId` 必填与 `target_jobs.resume_id` 存储。

- Workspace plan-list gate 没覆盖“无 practice plan 行但已有 target job resume binding”的 re-entry。
  - **类别**：spec-plan
  - **说明**：`frontend-workspace-and-practice/001` 旧 gate 只覆盖有 current plan 投影的卡片导航。

## 4 对流程资产的改进建议

- 保留 TargetJob import/list/get 的 target job-level resume binding gate。
  - **落点**：spec-plan
  - **优先级**：high

- 后续审查类似“创建时已选择 X，列表重进丢 X”的问题时，先区分 route state、target entity persistence、derived child entity projection。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：对已存在的修复前 TargetJob 行，决定是否需要产品层面的重新绑定/重新导入入口；代码无法可靠推断当时用户选择的简历。
- 次优先级：在下一轮 workspace owner review 中增加一个负向 fixture：`TargetJob.resumeId` 存在但 `currentPracticePlanId=null` 时，`进入规划` 仍进入详情而不是 missing-resume。
