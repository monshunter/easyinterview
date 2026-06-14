# D-20 Resume Negative Sweep Hardening 交付复盘报告

> **日期**: 2026-06-14
> **审查人**: Codex (GPT-5)

## 1 复盘范围与成功证据

本次交付范围是对 D-20 resume flatten owner plan 做 L2 code-review 修复闭环：把 BUG-0119 后续建议中的 runtime SQL / privacy payload / migration legacy-row / OpenAPI baseline negative sweep 固化为可执行 gate，并完成 `backend-resume/002` Phase 10 当前 D-20 contract 的场景与文档收口。

已完成的交付项：

- `backend-resume/002` Phase 10 增加 L2 negative sweep preflight gate，覆盖 runtime SQL、privacy payload、migration legacy row 与 OpenAPI baseline。
- `db-migrations-baseline/002` Phase 6 增加 legacy-row narrowing cleanup gate。
- `openapi-v1-contract/004` Phase 7 增加 retired fixture key gate。
- `validate_fixtures.py` 递归拒绝 fixture request / response 中的 D-20 退役 key `resumeAssetId` / `resumeVersionId`。
- `openapi/fixtures/Debriefs/suggestDebriefQuestions.json` 修正为 `resumeId`。
- P0.077 / P0.078 / P0.080 场景文档与脚本更新到当前 `async_jobs` + `ai_task_runs` tailor 语义，并修正旧 live-store 测试名。
- `backend-resume/002` plan / checklist / BDD 文档补齐 D-20 current operation matrix、当前 scenario checklist 与 active-vs-historical 执行边界；完成后状态回收为 `completed`。
- runtime 注释中清除 retired table / route literal，确保 D-20 zero-reference gate 不被注释残留误报。
- 新增 [BUG-0120](../bugs/BUG-0120.md)，把 retired fixture key 与 stale scenario script gate 漂移沉淀为 resolved 测试缺陷。

通过证据：

- `go test ./backend/internal/store/jobs ./backend/internal/privacy/runner ./backend/internal/migrations -run 'TestRepositoryGetJobScopesAsyncJobToOwningUser|TestSQLStoreMarkDeleteRequestCompletedDeletesAccountIdentityAndPreservesRequestTombstone|TestResumeFlattenMigrationContract|TestDropJDMatchMigrationDeletesRetiredAsyncJobsBeforeNarrowingCheck' -count=1`
- `python3 -m unittest scripts.lint.validate_fixtures_cli_test`
- `make validate-fixtures`
- `make openapi-diff`
- `make codegen-check`
- `go test ./backend/internal/resume/... ./backend/cmd/api -count=1`
- `go test ./backend/internal/resume/jobs ./backend/internal/resume/store -count=1`
- `test/scenarios/e2e/p0-074-resume-confirm-master-and-version-reads/scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-075-resume-update-version-merge-and-ik/scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-076-resume-branch-version-sync-paths/scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-077-resume-tailor-async-dispatch-and-ready/scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-078-resume-tailor-failure-and-retry/scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-079-resume-suggestion-accept-reject-terminal/scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-080-resume-versions-privacy-legacy/scripts/{setup,trigger,verify,cleanup}.sh`
- scoped runtime negative grep: 0 matches
- scoped OpenAPI fixture/schema/generated negative grep: 0 matches
- scoped scenario retired key / old tailor test negative grep: 0 matches
- `make docs-check`
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
- `git diff --check`

## 2 会话中的主要阻点/痛点

- **OpenAPI fixture validator 没有扫 request / response key**
  - **证据**：`make validate-fixtures` 在修复前通过，但新增 key-level validator 后立即失败于 `suggestDebriefQuestions.scenarios.default.request.body.resumeVersionId`。
  - **影响**：OpenAPI schema / generated 已切到 `resumeId`，fixture 仍可携带旧 key 并作为 mock truth source 继续扩散。

- **owner plan gate 分散在多个 owner 文档**
  - **证据**：runtime SQL、privacy cleanup、migration cleanup 与 OpenAPI baseline 分别落在 backend-resume、B4 migration、B2 OpenAPI 与 BUG-0119 报告中；未被一个 D-20 owner preflight gate 汇总。
  - **影响**：继续 Phase 10 时容易只看局部 gate，漏掉跨层删除副作用；本轮已把汇总 gate 固化到 `backend-resume/002` Phase 10。

- **场景脚本与 BDD 文档仍携带旧 tailor run 口径**
  - **证据**：P0.078 / P0.080 脚本仍引用已退役的 live-store 测试名，P0.077 / P0.078 / P0.080 场景说明仍描述专属 tailor run / suggestion table。
  - **影响**：场景可成为 no-op 或历史语义证明，无法证明当前 D-20 ephemeral tailor contract。

## 3 根因归类

- **Fixture semantic negative gate 缺失**
  - **类别**：spec-plan / test
  - Validator 原先覆盖结构、schema、provenance、privacy 和 operation coverage，但未覆盖退役字段 key 的语义负向检查。

- **Cross-owner deletion gate 没有 owner-level 汇总**
  - **类别**：spec-plan
  - D-20 flatten 是跨 B2/B4/backend-resume 的 schema deletion，单 owner checklist 里缺少一个专门的 L2 sweep anchor。

## 4 对流程资产的改进建议

- **保留 validator 级 retired-key gate**
  - **落点**：`scripts/lint/validate_fixtures.py`
  - **优先级**：high

- **后续 schema deletion 继续使用 owner preflight 汇总**
  - **落点**：对应 owner plan/checklist
  - **优先级**：high

- **下游 debrief D-20 Phase 10 仍需单独执行**
  - **落点**：`frontend-debrief/001` / `backend-debrief/001`
  - **优先级**：medium

## 5 建议优先级与后续动作

下一步建议进入 `/plan-code-review backend-debrief/001-debrief-record-and-analysis backend --fix`，把 `suggestDebriefQuestions` 的 backend/runtime Phase 8 从文档声明推进到真实 handler、fixture parity 和 scenario gate。理由是本轮只修正了 Debriefs fixture 的 retired key，尚未证明 debrief runtime 已完全承接 D-20 `resumeId` contract。

备选路径是 `/plan-code-review frontend-debrief/001-record-mode-ui-parity frontend --fix`，优先确认前端 debrief consumer 是否仍携带旧 resume version key 或 mock-only 语义。
