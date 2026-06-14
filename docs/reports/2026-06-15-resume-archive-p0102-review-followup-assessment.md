# Resume Archive And P0.102 Review Follow-up 交付复盘报告

> **日期**: 2026-06-15
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复 review 指出的两项回归：`archiveResume` 只改响应、不持久化 archive 状态；`E2E.P0.102` trigger/verify 仍引用 retired JD Match middleware 测试名。
- 成功证据：
  - `cd backend && go test ./internal/resume/... -run 'TestArchiveResume|TestRepositoryExposesFlatResumeMethods|TestArchiveResumeSoftHidesAndScopesUser|TestArchiveResumeAlreadyArchivedAndNotFound' -count=1 -v`
  - `cd backend && go test ./cmd/api -run 'TestBuildAPIHandlerMounts(TargetJobRoutes|UploadPresign|ResumeRoutes|PracticeAndProfileRoutes|ReportRoutes|JobRoute)BehindSessionMiddleware|TestJDMatchRoutesAreGonePerD17' -count=1 -v`
  - `bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/setup.sh`
  - `bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/trigger.sh`
  - `bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/verify.sh`
  - `bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/cleanup.sh`
  - `make validate-fixtures`
  - `cd backend && go test ./internal/resume/... ./cmd/api -count=1`
  - `make docs-check`
  - `git diff --check`
- 关联 Bug： [BUG-0125](../bugs/BUG-0125.md)。

## 2 会话中的主要阻点/痛点

- `archiveResume` 原测试只断言 response `status=archived` 和 user scope，未证明 store state changed。
  - **证据**：review 指出 follow-up GET/list 仍读 active row；修复前 `Service.ArchiveResume` 只调用 `ReadStore.Get`。
  - **影响**：endpoint success 与实际持久化状态脱节，二次 archive conflict fixture 无法由 runtime 产生。
- P0.102 场景 README 已写 retired `jd_match` 口径，但 wrapper 仍 grep 被删除的测试名。
  - **证据**：`trigger.sh` / `verify.sh` 修复前引用 `TestJDMatchRoutesRequireSessionOnAllRoutes`；修复后实际 PASS marker 是 `TestJDMatchRoutesAreGonePerD17`。
  - **影响**：backend focused test 可通过，场景 verify 仍失败，BDD gate 变成脚本漂移噪声。

## 3 根因归类

- Side-effect endpoint 缺少 store-level state proof。
  - **类别**：spec-plan
- Scenario wrapper 与 owner checklist 没有随 retired test name 同步。
  - **类别**：spec-plan
- 既有 Bug 模式库已覆盖 wrapper false-green / stale PASS marker 风险。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 在后续 backend side-effect endpoint checklist 中明确要求 store test 覆盖 state mutation、idempotent replay/conflict、follow-up read/list 可见性。
  - **落点**：spec-plan
  - **优先级**：medium
- Retired route/module cleanup 后，除 runtime negative grep 外，增加场景目录 `trigger.sh` / `verify.sh` focused test name 负向搜索。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一步优先在相关 review checklist 中复用 BUG-0125 的检查项：side-effect endpoint 不能只看响应 DTO，scenario wrapper 必须验证当前 PASS marker。
- 可延后处理：历史报告和历史 BUG 中的旧 JD Match 测试名不作为当前执行契约，保留其历史证据，不做批量改写。
