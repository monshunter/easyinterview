# Backend Upload File Objects and Presign Baseline

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-12

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [backend-upload spec](../../spec.md) §6 验收标准 C-1..C-8 落到 backend Go handler + store + 隐私删除链路：

- 实现 `POST /api/v1/uploads/presign` (createUploadPresign) handler，含 IK 校验 + public purpose enum 检查 + upload URL 生成（通过 `ObjectStore` interface 抽象）+ TTL 边界（A4 `upload.presignTTLSeconds`）+ byteSize 上限边界（A4 `upload.maxBytes.*` per purpose）；
- 实现 `file_objects` store + service layer：repository 提供 `Create(pending)` / `MarkUploaded` / `MarkScanFailed` / `MarkDeleted` / `HardDelete` / `DeleteFileObjectsForUser` / row-lock state validation；service 提供 `RegisterFileObject` 上传完成确认入口；
- state machine 校验：`pending → uploaded`、`pending|uploaded → scan_failed`、`pending|uploaded|scan_failed → deleted`，非法转换统一返回 B1 已登记 `VALIDATION_FAILED`；业务 register 不写 `registered` 状态，但必须在确认 object exists 后原子完成 `pending → uploaded` 或对已 `uploaded` 幂等通过；
- 隐私删除：privacy_delete job 调用 `DeleteFileObjectsForUser(userId)`，先对象存储 hard delete → DB 行 hard delete → audit tombstone（与 [backend-runtime-topology](../../../backend-runtime-topology/spec.md) 共同维护）；
- mock-first 对齐：handler 实际响应字段集 / status code / IK 行为与 [B2 fixture `createUploadPresign.json` `default` scenario](../../../mock-contract-suite/spec.md) 字节级一致；
- 通过 spec §6 C-1..C-8 验收 + 新增 E2E.P0.033 happy path BDD 场景；
- 不实现 backend internal GC runner（D-3.2 未确认事项 P0 不实现）；不实现独立 `getFileObject` endpoint（D-3.2 由 openapi-v1-contract/004 决定）。

## 2 背景

[engineering-roadmap §5.2](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 标记 `backend-upload` (C2) 为 Resume Workshop 阶段 1 第 1 个 subspec：必须在 `backend-resume`（C7）落地 `registerResume` handler 之前先把 `createUploadPresign` 真实可用，否则 frontend `ResumeCreateFlow` (upload tab) mock 验证后无法切真。

本 subject 之前的所有 P0 业务（`backend-targetjob/001` 等）已在自己的 plan 中消费 `createUploadPresign` fixture，但没有真实 backend handler；本 plan 是第一份把 createUploadPresign 从 fixture 切到真实业务逻辑的实施 plan。

每个 phase 是可独立验证的纵向切片：Phase 0 先补齐 A4/config 契约门禁；Phase 1 起来就有 handler skeleton + IK + purpose validation；Phase 2 起来就有 store + state machine；Phase 3 起来就有 ObjectStore interface + dev MinIO；Phase 4 起来就有 privacy delete 链路；Phase 5 收口 + BDD。

执行本 plan 前必须确认：

- [B2 createUploadPresign fixture](../../../mock-contract-suite/spec.md) 已就位（C-8 mock-first 对齐源）。
- [B4 baseline `file_objects` 表](../../../db-migrations-baseline/spec.md#21-in-scope) 字段就位（D-2 state machine + purpose check constraint：`purpose=[resume,target_job_attachment,privacy_export,source_snapshot,audio,video]`，`upload_status=[pending,uploaded,scan_failed,deleted]`）。
- [A4 secrets-and-config](../../../secrets-and-config/spec.md) 当前已暴露对象存储 endpoint / bucket / credentials 的 `objectStorage.*` / `OBJECT_STORAGE_*` 契约；本 plan Phase 0 必须先以 additive config path 登记 `objectStorage.provider` / `upload.presignTTLSeconds` / `upload.maxBytes.*`，不得在 backend-upload 代码中直接 `os.Getenv` 或私造未登记 env key。
- [A2 local-dev-stack](../../../local-dev-stack/spec.md) 已就位（dev 环境可拉起对象存储 mock，如 MinIO 或本地 filesystem fallback）。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + feature-behavior + contract`。本 plan 实现 backend handler / store / state machine / privacy delete 链路；涉及用户可见 HTTP API 行为。
- **TDD 策略**: 适用。Red-Green-Refactor 入口：
  1. config contract test：A4 typed config / validator 能读取 `objectStorage.provider`、`upload.presignTTLSeconds`、`upload.maxBytes.*`，且 `.env.example` 不新增未登记 key；
  2. handler unit test（`backend/internal/upload/handler/*_test.go`）：IK replay / purpose validation / TTL / byteSize limit / auth check；
  3. store integration test（`backend/internal/upload/store/*_integration_test.go`）：state transition / FK / cross-user isolation；
  4. ObjectStore interface mock + dev MinIO smoke：presign 签发可消费、URL 在 TTL 内可 PUT、超期拒绝；
  5. privacy delete unit test：对象存储删除失败 retryable / DB 行硬删幂等；
  6. 跑 `make backend-test` + `cd backend && go test ./internal/upload/...` 全 PASS。
  执行入口：`/implement backend-upload/001-file-objects-and-presign-baseline` → `/tdd`。
- **BDD 策略**: 适用（Feature plan requires BDD）。E2E.P0.033 file-presign-register-roundtrip 覆盖 happy path：presign → 客户端 PUT → register → delete。详见 [bdd-plan.md](./bdd-plan.md) / [bdd-checklist.md](./bdd-checklist.md)。
- **替代验证 gate**:
  - `make backend-test`（含 `internal/upload/...`）
  - `make lint-config`
  - `go test ./internal/upload/handler/... -run TestPresignIdempotency`
  - `go test ./internal/upload/store/... -run TestStateTransition`
  - smoke：`curl -X POST /api/v1/uploads/presign -H 'Idempotency-Key: ...' -d '{"purpose":"resume","fileName":"resume.pdf","contentType":"application/pdf","byteSize":1048576}'`
  - `sync-doc-index --check`

### 3.1 Frontend / Backend Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createUploadPresign` | `openapi/fixtures/Uploads/createUploadPresign.json` `default`（201 success）；purpose / size / auth / IK failure 由 handler tests 与 E2E.P0.033 直接断言，除非 B2 后续修订 fixture，否则本 plan 不私自声明 error fixture variant | `frontend-resume-workshop/002-create-flow-and-onboarding` upload tab (future), `backend-targetjob` file-import consumers via generated client | `backend/internal/upload/handler/presign.go` real handler | `file_objects` (`purpose`, `upload_status`, `byte_size`, `object_key`) + object storage object | none | E2E.P0.033 + handler/store unit/integration tests |

Config dependency: existing A4 `objectStorage.endpoint` / `bucket` / `accessKey` / `secretKey` map to `OBJECT_STORAGE_*`; this plan must add config-only paths `objectStorage.provider`, `upload.presignTTLSeconds`, and `upload.maxBytes.{resume,targetJobAttachment,privacyExport}` before handler code. If an env override is required, first revise A4 env dictionary and `.env.example`; do not introduce `UPLOAD_*` or `OBJECT_STORE_*` reads locally.

## 4 实施步骤

### Phase 0: A4 config contract preflight

#### 0.1 修订 / 验证 A4 config schema

- 在 [A4 secrets-and-config](../../../secrets-and-config/spec.md) 与 config artifacts 中登记 config-only paths：
  - `objectStorage.provider` (`minio | filesystem`)
  - `upload.presignTTLSeconds`（默认 600）
  - `upload.maxBytes.resume`（默认 10485760）
  - `upload.maxBytes.targetJobAttachment`（默认 10485760）
  - `upload.maxBytes.privacyExport`（默认 5242880）
- 复用现有 `OBJECT_STORAGE_ENDPOINT` / `OBJECT_STORAGE_BUCKET` / `OBJECT_STORAGE_ACCESS_KEY` / `OBJECT_STORAGE_SECRET_KEY` env key；不得使用未登记的 `OBJECT_STORE_*` 或 `UPLOAD_*` env key。

#### 0.2 config tests

- `make lint-config` 必须 PASS。
- config unit test 覆盖默认值、dev override、非法 provider、非正数 TTL / maxBytes fail-fast。

### Phase 1: handler skeleton + IK + purpose validation

#### 1.1 实现 `internal/upload/handler/presign.go`
- 实现 generated server interface `CreateUploadPresign` 方法
- 校验 `Idempotency-Key` header 存在 + 24h TTL（B1 idempotency 工具）
- 校验 purpose enum（D-1 锁定集）
- 校验 request `byteSize` 不超过 A4 `upload.maxBytes.*` per-purpose limit（配置注入；不把 limit 私加到 response）
- 返回 201 + `UploadPresign{fileObjectId, uploadUrl, method, headers, expiresAt}`，字段与当前 B2 schema / fixture 一致

#### 1.2 unit test
- `presign_test.go`：IK 缺失 / IK 24h 内 replay / IK >24h 拒绝 / purpose 非法 / size 超限

### Phase 2: file_objects store + state machine

#### 2.1 实现 `internal/upload/store/file_objects.go`
- Repository 方法：`Create(ctx, userId, purpose, fileName, contentType, byteSize, objectKey) → fileObjectId`
- `MarkUploaded(ctx, fileObjectId)` / `MarkScanFailed(ctx, fileObjectId, reason)` / `MarkDeleted(ctx, fileObjectId)` / `HardDelete`
- `LockForRegister(ctx, fileObjectId, ownerUserId)` 或等价 row-lock 查询：只返回同 user 行；cross-user / not-found 返回 404，不暴露存在
- state transition validation：在 store 层 enforce `pending → uploaded`、`pending|uploaded → scan_failed`、`pending|uploaded|scan_failed → deleted`；非法转换返回 `VALIDATION_FAILED`

#### 2.2 integration test
- `file_objects_integration_test.go`：CRUD + state transition + cross-user isolation + pending / uploaded / scan_failed / deleted register row-lock 校验 + privacy hard delete；不依赖 `resume_assets` 级联表达 register 状态

### Phase 3: ObjectStore interface + dev provider

#### 3.1 实现 `internal/upload/objectstore/interface.go`
- `Presign(ctx, objectKey, contentType, byteSize, ttl) → (url, method, headers, expiresAt)`
- `Delete(ctx, objectKey)` retryable
- `Exists(ctx, objectKey)` (用于 register 验证已上传)

#### 3.2 实现 internal `RegisterFileObject` completion service
- 在 `backend/internal/upload/service/register.go`（或同等 service 层）组合 repository row lock + `ObjectStore.Exists(objectKey)`：
  - 同 user + purpose 匹配 + `upload_status='pending'` + object exists：同一事务 / 临界区内 `MarkUploaded`，返回 fileObject metadata 给业务 handler 写 FK
  - 同 user + purpose 匹配 + `upload_status='uploaded'`：幂等通过，不重复写状态
  - `scan_failed` / `deleted`：返回 `VALIDATION_FAILED`
  - object missing：返回 `VALIDATION_FAILED`，提示客户端重新 PUT 或重新 presign；不得让业务 register 接受未完成上传
- unit / integration test 覆盖 presign → PUT → RegisterFileObject 主路径，防止只能通过测试直接调用 `MarkUploaded` 才能完成上传。

#### 3.3 实现 MinIO / filesystem 双 provider
- `objectstore/minio.go`：dev/staging（A2 dev stack）
- `objectstore/filesystem.go`：unit test fallback
- A4 config selector：`objectStorage.provider=minio|filesystem`

#### 3.4 smoke test
- 真 MinIO（A2 dev stack）：presign signs → PUT 接受 → URL 过期后 PUT 拒绝
- `internal/upload/objectstore/minio_smoke_test.go`（默认 skip，需要 `--integration` flag）

### Phase 4: privacy delete 链路

#### 4.1 实现 `DeleteFileObjectsForUser(ctx, userId) → []fileObjectId`
- 反查所有 user 拥有的 file_object 行
- 按 batch 调用 ObjectStore.Delete + DB hard delete（同事务 或 retryable job）
- 对象存储删除失败：DB 行保留原状态 + retryable job 重试；只有对象删除成功后才标记 `deleted` 或 hard delete DB 行
- audit tombstone：写入 audit_events（仅 fileObjectId / purpose / deletedAt，不含 objectKey）

#### 4.2 与 backend-runtime-topology privacy runner 集成
- 在 `backend/internal/privacy/runner/` 中调用 `upload.DeleteFileObjectsForUser`
- privacy_delete 跨域顺序以 [B4 §3.1.2](../../../db-migrations-baseline/spec.md#312-p0-privacy-deletion-table-matrix) 与各业务 owner 删除链路为准；backend-upload 只负责 file_objects 步骤中先删对象存储、成功后再 hard delete DB 行，不单独规定 `resume_assets` / `target_jobs` 的全局先后顺序

#### 4.3 unit + integration test
- `delete_for_user_test.go`：成功路径 / 对象存储失败重试 / 部分成功幂等

### Phase 5: 收口与 BDD

#### 5.1 跨 gate 收口

按 §3 替代验证 gate 依序运行：
- `make backend-test` + `go test ./internal/upload/...` PASS
- `make lint-config` PASS
- handler smoke：curl 真实端口验证 201 + IK replay 一致
- mock-first 对齐验证：本地 mock-server 与 真实 handler 同 endpoint 字节比对
- `sync-doc-index --check` PASS

#### 5.2 BDD 场景验证

- 执行 `test/scenarios/e2e/p0-033-file-presign-register-roundtrip/` 全套 setup → trigger → verify → cleanup PASS（详见 [bdd-checklist.md](./bdd-checklist.md)）
- 在 `test/scenarios/e2e/INDEX.md` 追加 E2E.P0.033 行

#### 5.3 spec / history / INDEX 同步

- backend-upload spec.md 本次 L1 修订后保持 1.1 active；实施完成时再追加完成行
- backend-upload history.md 已记录本次 L1 修订；plan 001 落地阶段维持 active
- 同步 `docs/spec/engineering-roadmap/spec.md` §5.2 `backend-upload` 状态从 "未创建" 改为 "active"（roadmap spec 3.10 → 3.11）

#### 5.4 通知下游 owner

- 通知 `backend-resume/001-asset-register-parse-and-listing` owner：createUploadPresign + file_objects + Register internal API 已就位，可启动 registerResume 真实落地；
- 通知 `frontend-resume-workshop/002-create-flow-and-onboarding`（未来 plan）：upload tab 可消费真实 backend presign；
- 通知 `backend-targetjob` owner：现有 fixture-backed file 流可保持不变，未来切真时调用相同 handler。

## 5 验收标准

- 本计划列出的 §4 所有 Phase task 全部完成
- §3 替代验证 gate 全部通过
- spec §6 C-1..C-8 全部 PASS
- BDD E2E.P0.033 PASS（含 setup → trigger → verify → cleanup 全脚本）
- backend-resume / backend-targetjob owner 已收到 createUploadPresign 落地信号

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: 对象存储 provider 选型未定 | ObjectStore interface 抽象 + 双实现（MinIO / filesystem）；A4 `objectStorage.provider` 配置可切换；P0 dev 用 MinIO，filesystem 仅供 unit test fallback |
| R2: presign URL 在 TTL 内泄漏导致越权 | TTL ≤ 600s + objectKey 含 userId 前缀；server-side 校验 register 时 fileObject.userId 匹配 |
| R3: state machine 复杂度与现有 file_objects.upload_status 字段对齐 | B4 baseline `file_objects.upload_status` 已存在；本 plan Phase 2.1 不改 schema，只补 store-layer validation |
| R4: privacy delete 对象存储失败 + DB 一致性 | retryable job + 保留 DB 原状态 + 重试上限；超过上限触发 P2 告警进入人工排查；不得私加 `deleted_pending` 状态 |
| R5: backend-targetjob 已 mock 消费 createUploadPresign，切真时可能出现 fixture / handler 字段不一致 | Phase 5.1 mock-first 对齐验证：handler 响应字段集 / status / header 与 fixture 字节比对；新增 scenario 必须 B2 plan 修订同步 |
