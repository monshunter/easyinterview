# Backend Upload 001 L2 Remediation 交付复盘报告

> **日期**: 2026-05-12
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-upload/001-file-objects-and-presign-baseline` L2 code review remediation，覆盖 `createUploadPresign` runtime route、real service implementation、`file_objects` register 事务边界、MinIO smoke 和 E2E.P0.033 scenario gate hardening。
- Bug 记录：新增 [BUG-0045](../bugs/BUG-0045.md)，记录 upload presign runtime path 只被 fake/wrapper gate 覆盖的问题。
- 红测证据：`TestCreateUploadPresignCreatesPendingFileObjectAndPresignsObject`、`TestRepositoryRegisterUploadedChecksObjectWhileRowLocked`、`TestBuildAPIHandlerMountsUploadPresignBehindSessionMiddleware` 在修复前分别暴露 service missing、store missing 与 route 404。
- 通过证据：上述 focused tests 均 PASS；`go test ./backend/internal/upload/... ./backend/internal/privacy/runner ./backend/cmd/api -count=1`、`cd backend && go test ./...`、`make validate-fixtures`、`make docs-check`、`git diff --check`、`make test`、`make lint`、`make build` 均 PASS。
- BDD 证据：E2E.P0.033 `setup -> trigger -> verify -> cleanup` PASS，trigger/verify 已纳入 real service、transactional register 和 API route focused tests。
- 环境说明：当前未设置 `DATABASE_URL` / `OBJECT_STORAGE_*`，DB integration 与 MinIO live smoke 按 env-gated contract skip；MinIO smoke 已不再无条件 skip。

## 2 会话中的主要阻点/痛点

- 原实现用 handler fake 和 scenario wrapper 形成假绿灯。
  - **证据**：`createUploadPresign` handler tests 只注入 fake service；真实 `cmd/api` route 红测返回 `404`。
  - **影响**：实现看似完成，但真实 HTTP client 无法调用 upload presign API。
- Presign service 没有 real persistence/provider contract。
  - **证据**：service 红测编译失败，`CreateUploadPresign` 与 input type 不存在。
  - **影响**：即使 route 挂载，真实 runtime 也无法创建 pending `file_objects` 或返回 ObjectStore presign。
- Register 状态机缺少事务边界 gate。
  - **证据**：store 红测编译失败；旧 service 在事务外 lock，再用另一条 store 方法 update。
  - **影响**：对象存在性检查与状态更新不能证明受同一个 row lock 保护。
- Live provider smoke 被历史无条件 skip 屏蔽。
  - **证据**：MinIO smoke 在 env 检查后仍 `t.Skip("live MinIO SDK smoke is deferred until provider SDK is introduced")`。
  - **影响**：配置齐全时也无法验证真实 signed header presign 和 PUT 上传。
- 当前本机缺少 live DB / MinIO env。
  - **证据**：scenario 与 integration smoke 均因 `DATABASE_URL` / `OBJECT_STORAGE_*` unset 跳过 live 分支。
  - **影响**：本轮已证明 unit、route、scenario script 与 offline harness；真实 dev stack 端到端仍需在环境就绪后补跑。

## 3 根因归类

- Handler envelope、route mounting、real service 和 persistence/provider side effects 没有同测。
  - **类别**：spec-plan
- Scenario scripts 没有反查 real-path focused tests，导致 fake-only gate 可通过。
  - **类别**：spec-plan
- Row lock 语义停留在 service 组合层，没有被 store 层事务方法固化。
  - **类别**：spec-plan
- Provider SDK 引入后，历史 defer skip 没有被 integration smoke gate 捕获。
  - **类别**：spec-plan
- Live env 缺失是环境前置，不代表本轮修复失败。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 后续 backend API plan 的 L2 gate 应默认包含 `cmd/api` route mounting test，并证明 session/idempotency middleware 能包住真实 handler。
  - **落点**：spec-plan
  - **优先级**：high
- 后续 handler 使用 fake service 时，必须另有 real service focused test 覆盖 persistence side effect、provider call 与 response contract。
  - **落点**：spec-plan
  - **优先级**：high
- 涉及 row lock 或状态机的 store 变更，应在 repository 层提供单事务方法，并用 focused test 断言 `FOR UPDATE`、callback 和 update 顺序。
  - **落点**：spec-plan
  - **优先级**：high
- Scenario verify 脚本应检查关键 focused test 名称或输出，避免场景只运行 wrapper tests 却没有覆盖真实 runtime path。
  - **落点**：spec-plan
  - **优先级**：medium
- Provider integration smoke 允许 env-gated skip，但不允许 SDK 已存在后继续保留无条件 skip。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：在 `backend-resume/001-asset-register-parse-and-listing` 启动前，把 upload dependency 明确为已通过的 offline/runtime contract，并把 live DB + MinIO dev-stack smoke 单独列为环境就绪后的 BDD gate。
- 下一步修复：执行 `/work-journal` 时使用 commit title `fix(backend-upload): wire presign runtime contracts`，并在日志中引用 [BUG-0045](../bugs/BUG-0045.md)。
- 可延后：将 “route mounting + real service side effect + transaction boundary” 抽成 backend API plan 的通用 L2 checklist 模板，降低后续 plan 重复漏项风险。
