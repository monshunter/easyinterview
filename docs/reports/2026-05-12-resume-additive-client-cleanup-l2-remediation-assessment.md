# Resume Additive Client Cleanup L2 Remediation 交付复盘报告

> **日期**: 2026-05-12
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 Resume Workshop additive L2 review 暴露的 generated TS client response typing、P0 export 501 typed fallback、dev fixture consumer 漏同步，以及 `resume_versions` live migration test cleanup 泄漏。
- 关联 owner docs 已原地修订并完成：`openapi-v1-contract/004-resume-additive-coverage` v1.1、`db-migrations-baseline/002-resume-versions-additive` v1.1、B2 spec v1.18、B4 spec v1.16、[BUG-0044](../bugs/BUG-0044.md)。
- 成功证据：`go test ./backend/cmd/codegen/openapi -count=1`、`pnpm --filter @easyinterview/frontend test src/api/devMockClient.test.ts`、`pnpm --filter @easyinterview/frontend typecheck`、`make lint-openapi`、`make validate-fixtures`、`make openapi-diff`、`make docs-check` 均通过。
- 迁移 live test 证据：`cd backend && go test ./internal/migrations/... -run 'TestResumeVersions|TestResumeAssetDeleteRequiresVersionCleanup' -count=2 -v` 通过；当前 `DATABASE_URL` 未设置，live cases 明确 skip，contract-only migration test PASS。

## 2 会话中的主要阻点/痛点

- Generated client gate 没覆盖 response matrix。
  - **证据**：`branchResumeVersion` 同时声明 `201` 与 `202`，generator 只取第一个 preferred status；`exportResumeVersion` / `requestPrivacyExport` 声明 `501`，runtime 却先 throw。
  - **影响**：fixture 与 generated client 同时存在但类型语义不一致，调用方会在合法 async path 上读错对象。
- Fixture consumer 同步不是 OpenAPI additive 的默认闭环。
  - **证据**：`frontend/src/api/devMockClient.test.ts` 初始失败，9 个 Resumes operation fixture 未导入 dev mock registry。
  - **影响**：dev fixture-backed client 无法启动，说明 contract-ready 和 frontend mock-ready 被误判为同一状态。
- Live DB cleanup 风险依赖环境才暴露。
  - **证据**：当前本机缺 `DATABASE_URL`，只能确认 skip 与 helper 代码路径；review 指出有 live DB 时固定 UUID 会残留。
  - **影响**：本地无 live DB 时容易把 contract test PASS 误当 cleanup PASS。

## 3 根因归类

- Response matrix 漏测。
  - **类别**：spec-plan
  - **说明**：OpenAPI additive plan 的 codegen gate 更关注 operation/schema 数量，没有要求 generator 单测覆盖多 explicit response status 与 declared non-OK typed response。
- Consumer registry 漏同步。
  - **类别**：README / spec-plan
  - **说明**：`frontend/README.md` 要求 fixtures 是 mock 唯一来源，但 B2 plan checklist 没把 dev mock registry coverage 作为新增 operation 的完成项。
- Live cleanup gate 表达不够硬。
  - **类别**：spec-plan
  - **说明**：B4 plan 有 FK/privacy 删除顺序，但测试 helper cleanup 没被列为可重复运行 gate；本次已在 B4 spec C-14 固化。

## 4 对流程资产的改进建议

- 在 B2 OpenAPI plan 模板或 checklist 约定中增加 response matrix gate。
  - **落点**：spec-plan
  - **优先级**：high
  - **内容**：新增 operation 若有多个 explicit 2xx 或 explicit P0 non-OK response，必须有 generator unit test 和 frontend typed runtime test。
- 把 `frontend/src/api/devMockClient.test.ts` 纳入新增 operation 的固定替代验证 gate。
  - **落点**：README / spec-plan
  - **优先级**：medium
  - **内容**：B2 contract-ready 后必须证明 fixture-backed frontend registry 覆盖新增 operationId，避免未启动 UI 就破坏 dev preview。
- 对 live migration tests 增加环境分层记录。
  - **落点**：spec-plan
  - **优先级**：medium
  - **内容**：每次 live DB gate 都要区分 `DATABASE_URL` absent skip、contract-only PASS、live DB PASS，不允许合并成单一 PASS 口径。

## 5 建议优先级与后续动作

- 最高优先级：后续 B2 contract/codegen 变更先补 response matrix gate，尤其是多状态码与 P0 exception。
- 次优先级：将 dev mock operationId coverage 写进 OpenAPI additive checklist 模板或 B2 README 的新增 operation checklist。
- 可延后：在有 live Postgres 的本地环境重跑 B4 C-14，取得真正的 fixed UUID rerun-safe cleanup PASS 证据。
