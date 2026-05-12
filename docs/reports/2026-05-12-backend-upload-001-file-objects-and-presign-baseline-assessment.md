# Backend Upload 001 File Objects And Presign Baseline 交付复盘报告

> **日期**: 2026-05-12
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-upload/001-file-objects-and-presign-baseline`，覆盖 A4 upload config paths、`createUploadPresign` handler、`file_objects` store/state machine、ObjectStore MinIO/filesystem provider、`RegisterFileObject` internal service、privacy delete object-first 链路与 E2E.P0.033 BDD 资产。
- 成功证据：已提交 phase commits `feat(backend-upload): complete config contract preflight`、`feat(backend-upload): add presign handler validation`、`feat(backend-upload): add file object store`、`feat(backend-upload): add object store register service`、`feat(backend-upload): add privacy file deletion`、`test(backend-upload): add presign register bdd scenario`。
- 验证证据：`make lint-config`、`cd backend && go test ./...`、`go test ./backend/internal/upload/...`、`go test ./internal/upload/handler -run TestCreateUploadPresignFixtureParity -count=1`、E2E.P0.033 `setup → trigger → verify → cleanup`、`make docs-check`、`git diff --check` 均通过。
- 环境说明：当前未配置 live `DATABASE_URL` / `OBJECT_STORAGE_*`，integration-tag DB / MinIO smoke 按测试契约 skip；unit、sqlmock、fixture parity、negative grep 与 BDD harness gate 已通过。

## 2 会话中的主要阻点/痛点

- 全量 config test 暴露 Phase 0 focused gate 覆盖不足。
  - **证据**：`cd backend && go test ./...` 初次失败于 `backend/internal/platform/config`，原因是 prod test fixture 未补 `objectStorage.provider` 与 `upload.*` 默认值。
  - **影响**：Phase 5 收口前需要返修 `validator_test.go`，说明 focused config tests 不足以替代全量后端 gate。
- BDD live 环境未就绪，E2E.P0.033 只能以 offline harness + integration-tag skip 方式闭合。
  - **证据**：scenario setup log 中 `DATABASE_URL` / `OBJECT_STORAGE_*` 均为空；DB / MinIO integration tests 输出 skip。
  - **影响**：当前可证明 backend-owned 行为与隐私红线，但不能声称真实 dev stack HTTP + MinIO + `DELETE /api/v1/me` 全链路已跑通。
- Checklist 写了不存在的 `make backend-test` 目标。
  - **证据**：Makefile 当前有 `test` 与 `lint-config`，未定义 `backend-test`；本次以 `cd backend && go test ./...` 作为等价后端 gate。
  - **影响**：执行时需要人工判断等价命令，增加收口歧义。

## 3 根因归类

- Phase gate 粒度与全量 gate 的关系没有在 backend-upload checklist 中强制表达。
  - **类别**：spec-plan
- E2E.P0.033 的 live stack 前置与 offline fallback 边界在 BDD checklist 中不够明确。
  - **类别**：spec-plan
- Make target 名称漂移未被 plan-review 提前发现。
  - **类别**：spec-plan / README

## 4 对流程资产的改进建议

- 在后续 backend plan 的 Phase 0/Phase 5 checklist 中同时保留 focused tests 与 `cd backend && go test ./...`，避免只跑 focused config tests 后遗漏共享 test fixture。
  - **落点**：spec-plan
  - **优先级**：high
- 对需要 live DB / MinIO 的 BDD checklist 增加两层 gate：`offline backend-owned harness` 与 `live dev-stack smoke`，分别记录 skip 条件，避免把 skip 误写成完整 live PASS。
  - **落点**：spec-plan
  - **优先级**：high
- 将 `make backend-test` 统一替换为当前仓库真实存在的 Make target，或在 Makefile 中补一个别名，减少后续执行歧义。
  - **落点**：README / spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮最值得优先处理：启动 `backend-resume/001-asset-register-parse-and-listing` 前，把它的 checklist 中 backend-upload 前置从 blocker 改为 passed dependency，并明确 live BDD 是否需要等待 A2 dev stack。
- 可以延后处理：为 Makefile 增加 `backend-test` alias 或批量修订历史 plan 中的 stale target 名称。
