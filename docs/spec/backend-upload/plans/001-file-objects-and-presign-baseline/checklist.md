# Backend Upload File Objects and Presign Baseline Checklist

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

> Phase 0-6 的已勾选项只保留为历史交付证据；Phase 7 是当前 JD attachment purpose 收缩合同。旧 Phase 中的 TargetJob upload/config 正向口径不构成当前实现、验收或兼容要求。

## Phase 0: A4 config contract preflight

- [x] 0.1 登记 A4 config-only paths：`objectStorage.provider` / `upload.presignTTLSeconds` / `upload.maxBytes.resume` / `upload.maxBytes.targetJobAttachment` / `upload.maxBytes.privacyExport`，复用现有 `OBJECT_STORAGE_*` env key（验证：A4 spec/config artifacts 同步）
- [x] 0.2 禁止 backend-upload 直接读取未登记 `UPLOAD_*` / `OBJECT_STORE_*` env key（验证：grep negative + `make lint-config` PASS）
- [x] 0.3 A4 typed owner 单一 contract 覆盖默认值、override、非法 provider、非正数 TTL / maxBytes；backend-upload 不复制 config matrix（验证：A4 contract + lint-config PASS）

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

- [x] 5.1 阶段完成由仓库根 `make test` 承接前后端全量单测；`make lint-config` 作为独立配置 gate。
- [x] 5.2 mock-first 对齐：handler 真实响应与 B2 fixture `default` scenario 字节比对 PASS（验证：mock-contract-suite 测试集成）
- [x] 5.5 同步 `docs/spec/engineering-roadmap/spec.md` §5.2 `backend-upload` 状态从 "未创建" 改为 "active"，spec 3.10 → 3.11，history.md 追加 3.11 行（验证：`sync-doc-index --check`）
- [x] 5.6 通知 backend-resume/001 owner：createUploadPresign + Register internal API 已就位（验证：cross-plan 引用 commit）

## Phase 6: L2 remediation hardening

- [x] 6.1 RegisterFileObject 在 row lock 内校验对象存储实际 size 与 `file_objects.byte_size` 精确匹配，object missing / size mismatch 不得标记 `uploaded`（验证：`go test ./backend/internal/upload/store -run 'TestRepositoryRegisterUploaded(ChecksObjectWhileRowLocked|RejectsObjectSizeMismatchWhileRowLocked)' -count=1`；`go test ./backend/internal/upload/service -run 'TestRegisterFileObject(MarksPendingUploadedAfterObjectExists|RejectsMissingObjectAndIllegalStates)' -count=1`；`go test ./backend/internal/upload/objectstore -run 'TestFilesystem' -count=1`）
- [x] 6.2 `createUploadPresign` idempotency TTL 与 `upload.presignTTLSeconds` 对齐，超过 signed URL TTL 不 replay 旧 body（验证：`go test ./backend/cmd/api -run TestBuildUploadRoutesAlignsIdempotencyTTLWithPresignTTL -count=1`；`go test ./backend/internal/upload/handler -run TestCreateUploadPresignIdempotencyReplayAndTTL -count=1`）
- [x] 6.3 `cmd/api` runtime privacy_delete runner kernel 挂入 upload deleter，`DELETE /api/v1/me` 创建的 job 可调用 `DeleteFileObjectsForUser(userId)`（验证：`go test ./backend/internal/privacy/runner -run TestPrivacyDeleteHandler -count=1`；`go test ./backend/cmd/api -run 'TestBuildTargetJobRuntime(RegistersPrivacyDeleteHandler|WiresRunnerAndAIClient)' -count=1`）
- [x] 6.4 `file_objects` DB hard delete 与 audit tombstone 同事务提交，audit 失败时 row 保留可重试（验证：`go test ./backend/internal/upload/store -run 'TestRepositoryHardDeleteWithAuditTombstone|TestRepositoryInsertAuditTombstoneDoesNotPersistObjectKey' -count=1`；`go test ./backend/internal/upload/service -run 'TestDeleteFileObjectsForUser(DeletesObjectsBeforeDBAndWritesAudit|ObjectDeleteFailureIsRetryableAndKeepsDBRows|UsesAtomicDBDeleteAndAudit)' -count=1`）
- [x] 6.5 上述 focused commands 只作开发反馈；Phase 6 完成证据统一为仓库根 `make test`，integration test 不进入 E2E。

## Phase 7: Remove JD attachment upload purpose

- [x] 7.1 RED: B1/B3/OpenAPI、B4、A4 与 backend-upload 各 owner test 共同证明 JD attachment purpose 与专属 maxBytes 仍可达；本 owner 锁定 purpose validation/handler 分支和 `resume` / `privacy_export` 正向基线。
  <!-- verified: 2026-07-13 method=service-purpose-red evidence="direct upload service test accepted target_job_attachment and returned nil before GREEN; handler/config tests lock rejection plus resume/privacy limits" -->
- [x] 7.2 GREEN: 先消费 OpenAPI/B4 purpose 与 A4 Phase 12 maxBytes handoff，再只删除 backend-upload handler/service 自有分支；不直接修改 A4 config/validator/composition 或 B4 DB constraint，并保留 endpoint、state machine、resume register 与 privacy delete。

## Phase 8: Injected upload size guard

- [x] 8.1 OWNER-GATE: resume/privacy missing/default/override/invalid 只由 A4 typed contract 覆盖；删除 backend-upload 重复 config/wiring tests。
- [x] 8.2 FOCUSED-GATE: presign/register 注入小型 declared/object limits；overflow/size mismatch 零 DB/object state transition，不生成默认大小文件。
- [x] 8.4 仓库根 `make test`、privacy integration、contexts/docs/diff 与 duplicate-limit negative search 通过；不新增配置 E2E。
- [x] 7.4 Zero-ref: OpenAPI/generated/backend/migrations/config/fixtures/scripts 精确搜索旧 purpose 为零；仓库根 `make test` 完成前后端全量单测回归，resume/privacy focused tests 仅作开发反馈。

## BDD Gate

- [x] BDD-Gate: `BDD.UPLOAD.FILE.001` 由 [BDD checklist](./bdd-checklist.md) 关联 presign/register owner behavior tests；不创建或声明真实 E2E PASS。
