# Structured Interview Rounds Data Binding 交付复盘报告

> **日期**: 2026-07-09
> **审查人**: Codex

## 1 复盘范围与成功证据

- 交付范围：修复 `frontend-home-job-picks-and-parse/001-home-jd-import-and-parse` 中 TargetJob 轮次规划的数据绑定与契约漂移。`target.import.parse`、OpenAPI/backend persistence、Parse 详情、Home 最近模拟面试卡片、Workspace 回访 handoff 和 shared navigation context 必须使用保存的 2~5 条 `TargetJob.summary.interviewRounds[]`，由 LLM 推断轮数、类型、标题、时长和 focus。
- 根因判断证据：本地 runtime 使用 real backend；数据库中最近 TargetJob 的历史 `summary.interviewHypotheses` 已存在差异化内容；`ai_task_runs` provenance 显示 DeepSeek `jd_parse` 成功运行。因此最初问题不是后端 mock，而是前端相关 consumer 使用静态轮次文案；进一步修订后，string-only hypotheses 合同也被结构化 `interviewRounds[]` 取代。
- 修复证据：
  - `make lint-prompts`
  - `make codegen-openapi && make lint-openapi && make validate-fixtures`
  - `cd backend && go test ./internal/targetjob -run 'TestParseExecutor|TestDecodeParseResponse|TestTargetImportParse' -count=1`
  - `pnpm --dir frontend test --run src/app/navigation/interviewContext.test.ts src/app/screens/home/MockInterviewCard.test.tsx src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/parse/ParseScreen.test.tsx src/app/screens/parse/ParseEdit.test.tsx src/api/targetJob.realApiMode.test.ts`
  - `node --test ui-design/ui-design-contract.test.mjs`
  - `pnpm --dir frontend typecheck`
  - `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/trigger.sh && .../verify.sh`，其中 Playwright readonly detail 验收产出 screenshot attachment 与 `screenshotBytes=` marker。
  - context validation, `sync-doc-index --check`, `make docs-check`, `git diff --check`, and `make lint-core-loop-pruning-surface` all passed.

## 2 会话中的主要阻点/痛点

- 最初症状出现在 Parse 详情截图，但同一数据对象还被 Home recent cards 和 shared navigation context 消费。
  - **证据**：用户补充要求“关联的页面和对象，要一并适配”；随后定位到 `MockInterviewCard` 的 local rounds 和 `interviewContextFromTargetJob` 的 static round name。
  - **影响**：若只修 Parse 详情，首页卡片和启动面试上下文仍会展示或传递另一套静态轮次，用户会继续看到不一致结果。
- 旧测试主要保护布局和只读行为，缺少差异化 LLM structured round fixture。
  - **证据**：本次新增 focused tests、UI contract test 和 P0.016 scenario checks，才覆盖 `parse-round-*`、`home-recent-mock-rail-*`、route round context 和截图验收的同源数据绑定。
  - **影响**：静态 fallback 能长期作为“正常布局”通过测试，直到真实用户比较多个规划时才暴露。
- UI 真理源、正式前端和场景资产都含有 round assumptions 口径。
  - **证据**：本次同时更新 `ui-design/`、`docs/ui-design/module-job-workspace.md`、owner plan/BDD、formal frontend 和 P0.016 scenario。
  - **影响**：单点修复容易与 UI truth source 或 scenario gate 重新漂移，必须在同一 owner 计划内收口。

## 3 根因归类

- Round assumptions 没有被定义成 TargetJob 的结构化 LLM display object，而是分散在 prompt string-only output、页面静态文案、卡片默认值和 route context 里。
  - **类别**：spec-plan
- 前后端数据绑定测试没有要求“有 backend structured rounds 时不得回退静态文案/固定 4 轮模板”，也没有截图验收。
  - **类别**：spec-plan
- 现有 `AGENTS.md` 和 skill 已要求关联页面、真理源和 BDD 同步；本次主要是 owner plan 旧 gate 覆盖不足，不需要直接修改全局规章。
  - **类别**：no repo change needed

## 4 对流程资产的改进建议

- 后续涉及 AI/后端生成字段的展示变更，应在 owner checklist 中显式列出 output schema、persistence、shared mapper、所有 consumer、fallback 边界、静态字符串负向搜索和 screenshot gate。
  - **落点**：spec-plan
  - **优先级**：high
- P0.016 后续继续作为 Parse readonly detail 和 Home recent card 共同 gate，保留对 2~5 条 `summary.interviewRounds[]`、old static round strings 和 screenshot marker 的断言。
  - **落点**：spec-plan
  - **优先级**：high
- 若未来抽象更多 TargetJob display object，可考虑把 title/company/resume/round assumptions 的显示派生统一收敛到同一 frontend domain helper，避免页面重复解释 API shape。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- high：执行 `/work-journal`，用 `fix(frontend-home): derive structured interview rounds from JD (BUG-0148)` 提交本次代码、owner docs、scenario、Bug 和复盘资产。
- high：下一轮若继续改 TargetJob 规划详情，先从 `frontend/src/app/interview-context/roundAssumptions.ts` 反查所有 consumer，不要在页面内新建独立 rounds/defaults。
- medium：在后续 plan review 中推广“AI output display object consumer matrix”，优先覆盖 Home、Parse、Workspace handoff 和 Practice start params 这类跨页面对象。
