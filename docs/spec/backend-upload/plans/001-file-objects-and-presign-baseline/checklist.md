# Backend Upload File Objects and Presign Baseline Checklist

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-05-13

**关联计划**: [plan](./plan.md)

## Phase 0: A4 config contract preflight

- [x] 0.1 登记 A4 config-only paths：`objectStorage.provider` / `upload.presignTTLSeconds` / `upload.maxBytes.resume` / `upload.maxBytes.targetJobAttachment` / `upload.maxBytes.privacyExport`，复用现有 `OBJECT_STORAGE_*` env key（验证：A4 spec/config artifacts 同步）
- [x] 0.2 禁止 backend-upload 直接读取未登记 `UPLOAD_*` / `OBJECT_STORE_*` env key（验证：grep negative + `make lint-config` PASS）
- [x] 0.3 config unit test 覆盖默认值、dev override、非法 provider、非正数 TTL / maxBytes fail-fast（验证：config tests PASS）

## Phase 1: handler skeleton + IK + purpose validation

- [x] 1.1 实现 `backend/internal/upload/handler/presign.go`，实现 generated server interface `CreateUploadPresign`（验证：编译 PASS + `go vet` PASS）
- [x] 1.2 实现 IK 校验（缺失 / 24h TTL replay / 24h 外拒绝），返回 `VALIDATION_FAILED`（验证：unit test `TestPresignIdempotency` 3 case PASS）
- [x] 1.3 实现 purpose enum 校验（D-1 锁定集 + 拒绝未知值）（验证：unit test `TestPresignPurposeValidation` PASS）
- [x] 1.4 实现 request `byteSize` 上限校验（A4 `upload.maxBytes.*` per-purpose 配置注入，不向 response 私加 `maxBytes`）（验证：unit test `TestPresignByteSizeLimit` PASS）
- [x] 1.5 handler 返回 201 + `UploadPresign{fileObjectId, uploadUrl, method, headers, expiresAt}` 与 B2 fixture `default` scenario 字节一致（验证：fixture parity test）

## Phase 2: file_objects store + state machine

- [x] 2.1 实现 `backend/internal/upload/store/file_objects.go` Repository，方法签名：`Create / MarkUploaded / MarkScanFailed / MarkDeleted / HardDelete / DeleteFileObjectsForUser / LockForRegister`（或等价 row-lock 查询）（验证：编译 PASS）
- [x] 2.2 在 store 层实现 state transition validation：`pending → uploaded`、`pending|uploaded → scan_failed`、`pending|uploaded|scan_failed → deleted`；非法转换返回 `VALIDATION_FAILED`（验证：unit test `TestStateTransition` 含所有合法 + 非法转换 case PASS）
- [x] 2.3 实现 register row-lock 校验：同 user + purpose 匹配；cross-user / not-found 返回 404；`scan_failed` / `deleted` 返回 `VALIDATION_FAILED`；不写 `registered` 状态（验证：integration test）
- [x] 2.4 integration test：CRUD + state transition + cross-user isolation + FK 约束验证（验证：`go test ./internal/upload/store/... -tags=integration` PASS）

## Phase 3: ObjectStore interface + dev provider

- [x] 3.1 实现 `backend/internal/upload/objectstore/interface.go`，方法签名：`Presign / Delete / Exists`（验证：编译 PASS）
- [x] 3.2 实现 `RegisterFileObject(ctx, fileObjectId, expectedPurpose, ownerUserId)` internal service：`pending` 行先 `ObjectStore.Exists(objectKey)` 再原子 `MarkUploaded`；已 `uploaded` 幂等通过；object missing / 非法状态返回 `VALIDATION_FAILED`（验证：presign → PUT → register integration test PASS）
- [x] 3.3 实现 MinIO provider `objectstore/minio.go`（A2 dev stack）（验证：MinIO smoke `go test ./internal/upload/objectstore/... -tags=integration -run TestMinIO` PASS）
- [x] 3.4 实现 filesystem fallback provider `objectstore/filesystem.go`（unit test fallback，不持久 / 内存映射）（验证：unit test PASS）
- [x] 3.5 A4 config selector `objectStorage.provider=minio|filesystem` 切换（验证：runtime config 测试 + dev stack 启动验证）
- [x] 3.6 真 MinIO smoke：presign URL 在 TTL 内 PUT 接受、超期后 PUT 拒绝（验证：手工 curl smoke 或 integration test）

## Phase 4: privacy delete 链路

- [x] 4.1 实现 `DeleteFileObjectsForUser(ctx, userId) → []fileObjectId`（验证：integration test 含 5+ fileObject 删除场景）
- [x] 4.2 对象存储删除失败 retryable + DB 行保留原状态等待重试；不引入 `deleted_pending` 状态（验证：unit test `TestDeleteRetryable` PASS）
- [x] 4.3 audit_events 写入 tombstone（含 fileObjectId / purpose / deletedAt，不含 objectKey）（验证：integration test 验证 audit 行）
- [x] 4.4 在 `backend/internal/privacy/runner/` 中接入 `upload.DeleteFileObjectsForUser`（验证：privacy runner integration test PASS）
- [x] 4.5 privacy_delete 跨域顺序以 B4 §3.1.2 与业务 owner 删除链路为准；backend-upload file_objects 步骤必须先删对象存储、成功后再 hard delete DB 行，且不得单独规定 `resume_assets` / `target_jobs` 全局先后（验证：privacy runner integration test + B4 matrix dry-run 对照）

## Phase 5: 收口与 BDD

- [x] 5.1 跑 `make lint-config` + `make backend-test` + `go test ./internal/upload/...` 全 PASS（验证：exit 0）
- [x] 5.2 mock-first 对齐：handler 真实响应与 B2 fixture `default` scenario 字节比对 PASS（验证：mock-contract-suite 测试集成）
- [x] 5.3 BDD-Gate: 验证 E2E.P0.033 file-presign-register-roundtrip PASS（详见 [bdd-checklist.md](./bdd-checklist.md)）
- [x] 5.4 在 `test/scenarios/e2e/INDEX.md` 追加 E2E.P0.033 行（关联需求 `backend-upload C-1, C-2, C-3, C-4, C-6, C-7, C-8`，状态 Ready，automated）
- [x] 5.5 同步 `docs/spec/engineering-roadmap/spec.md` §5.2 `backend-upload` 状态从 "未创建" 改为 "active"，spec 3.10 → 3.11，history.md 追加 3.11 行（验证：`sync-doc-index --check`）
- [x] 5.6 通知 backend-resume/001 owner：createUploadPresign + Register internal API 已就位（验证：cross-plan 引用 commit）

## Phase 6: L2 remediation hardening

- [x] 6.1 RegisterFileObject 在 row lock 内校验对象存储实际 size 与 `file_objects.byte_size` 精确匹配，object missing / size mismatch 不得标记 `uploaded`（验证：`go test ./backend/internal/upload/store -run 'TestRepositoryRegisterUploaded(ChecksObjectWhileRowLocked|RejectsObjectSizeMismatchWhileRowLocked)' -count=1`；`go test ./backend/internal/upload/service -run 'TestRegisterFileObject(MarksPendingUploadedAfterObjectExists|RejectsMissingObjectAndIllegalStates)' -count=1`；`go test ./backend/internal/upload/objectstore -run 'TestFilesystem' -count=1`）
- [x] 6.2 `createUploadPresign` idempotency TTL 与 `upload.presignTTLSeconds` 对齐，超过 signed URL TTL 不 replay 旧 body（验证：`go test ./backend/cmd/api -run TestBuildUploadRoutesAlignsIdempotencyTTLWithPresignTTL -count=1`；`go test ./backend/internal/upload/handler -run TestCreateUploadPresignIdempotencyReplayAndTTL -count=1`）
- [x] 6.3 `cmd/api` runtime privacy_delete drainer 挂入 upload deleter，`DELETE /api/v1/me` 创建的 job 可调用 `DeleteFileObjectsForUser(userId)`（验证：`go test ./backend/internal/privacy/runner -run TestPrivacyDeleteHandler -count=1`；`go test ./backend/cmd/api -run 'TestBuildTargetJobRuntime(RegistersPrivacyDeleteHandler|WiresDrainerAndAIClient)' -count=1`）
- [x] 6.4 `file_objects` DB hard delete 与 audit tombstone 同事务提交，audit 失败时 row 保留可重试（验证：`go test ./backend/internal/upload/store -run 'TestRepositoryHardDeleteWithAuditTombstone|TestRepositoryInsertAuditTombstoneDoesNotPersistObjectKey' -count=1`；`go test ./backend/internal/upload/service -run 'TestDeleteFileObjectsForUser(DeletesObjectsBeforeDBAndWritesAudit|ObjectDeleteFailureIsRetryableAndKeepsDBRows|UsesAtomicDBDeleteAndAudit)' -count=1`）
- [x] 6.5 BDD-Gate hardening: E2E.P0.033 在缺少 `DATABASE_URL` / `OBJECT_STORAGE_*`、live integration skip 或 focused gate no-op 时 fail；`trigger.sh` 必须执行 `TestUploadPresignRegisterPrivacyDeleteLiveRoundtrip` 覆盖真实 HTTP presign → MinIO PUT → register → `DELETE /api/v1/me` → privacy drainer roundtrip，不能只以离线 focused tests 作为 PASS 证据（验证：`python3 test/scenarios/e2e/p0-033-file-presign-register-roundtrip/scripts/script_contract_test.py`；`go test ./backend/cmd/api -tags=integration -run TestUploadPresignRegisterPrivacyDeleteLiveRoundtrip -count=1 -v`）
