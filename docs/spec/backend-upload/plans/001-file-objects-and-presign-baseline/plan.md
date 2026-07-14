# Backend Upload File Objects and Presign Baseline

> **版本**: 1.6
> **状态**: active
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [backend-upload spec](../../spec.md) §6 验收标准落到 backend Go handler + store + 隐私删除链路；当前 public purpose 只保留 `resume` / `privacy_export`，JD paste-only 不消费上传服务：

- 实现 `POST /api/v1/uploads/presign` (createUploadPresign) handler，含 IK 校验 + public purpose enum 检查 + upload URL 生成（通过 `ObjectStore` interface 抽象）+ TTL 边界（A4 `upload.presignTTLSeconds`）+ byteSize 上限边界（A4 `upload.maxBytes.*` per purpose）；
- 实现 `file_objects` store + service layer：repository 提供 `Create(pending)` / `MarkUploaded` / `MarkScanFailed` / `MarkDeleted` / `HardDelete` / `DeleteFileObjectsForUser` / row-lock state validation；service 提供 `RegisterFileObject` 上传完成确认入口；
- state machine 校验：`pending → uploaded`、`pending|uploaded → scan_failed`、`pending|uploaded|scan_failed → deleted`，非法转换统一返回 B1 已登记 `VALIDATION_FAILED`；业务 register 不写 `registered` 状态，但必须在确认 object exists 后原子完成 `pending → uploaded` 或对已 `uploaded` 幂等通过；
- 隐私删除：privacy_delete job 调用 `DeleteFileObjectsForUser(userId)`，先对象存储 hard delete → DB 行 hard delete → audit tombstone（与 [backend-runtime-topology](../../../backend-runtime-topology/spec.md) 共同维护）；
- mock-first 对齐：handler 实际响应字段集 / status code / IK 行为与 [B2 fixture `createUploadPresign.json` `default` scenario](../../../mock-contract-suite/spec.md) 字节级一致；
- 通过 spec §6 C-1..C-9 验收；E2E.P0.033 继续覆盖 resume presign/register/privacy delete，且增加 JD attachment purpose 不可用断言；
- 不实现 backend internal GC runner（D-3.2 未确认事项 P0 不实现）；不实现独立 `getFileObject` endpoint（D-3.2 由 openapi-v1-contract/004 决定）。

## 2 背景

[engineering-roadmap §5.2](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 标记 `backend-upload` (C2) 为 Resume Workshop 阶段 1 第 1 个 subspec：必须在 `backend-resume`（C7）落地 `registerResume` handler 之前先把 `createUploadPresign` 真实可用，否则 frontend `ResumeCreateFlow` (upload tab) mock 验证后无法切真。

本 plan 已把 createUploadPresign 从 fixture 切到真实业务逻辑。Phase 0-6 保留为既有 file infrastructure 交付证据；当前待执行 Phase 7 只收缩 JD 专属 purpose/config，不改变简历与隐私链路。

每个 phase 是可独立验证的纵向切片：Phase 0 先补齐 A4/config 契约门禁；Phase 1 起来就有 handler skeleton + IK + purpose validation；Phase 2 起来就有 store + state machine；Phase 3 起来就有 ObjectStore interface + dev MinIO；Phase 4 起来就有 privacy delete 链路；Phase 5 收口 + BDD。

执行本 plan 前必须确认：

- [B2 createUploadPresign fixture](../../../mock-contract-suite/spec.md) 已就位（C-8 mock-first 对齐源）。
- [B4 baseline `file_objects` 表](../../../db-migrations-baseline/spec.md#21-in-scope) 保持 state machine，public purpose 收敛为 `resume` / `privacy_export`；DB-local purpose 由各 owner 管理，不扩大 OpenAPI。
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
  7. Phase 7 先以 OpenAPI/DB/config/handler tests 证明 JD attachment purpose 与专属 maxBytes 仍可达（RED），再删除最小分支并保持 resume/privacy tests 通过（GREEN）。
  执行入口：`/implement backend-upload/001-file-objects-and-presign-baseline` → `/tdd`。
- **BDD 策略**: 适用（Feature plan requires BDD）。E2E.P0.033 继续覆盖 `purpose=resume` 的 presign → PUT → register → privacy delete，并在 Phase 7 增加 JD attachment purpose 被拒绝、`privacy_export` 仍被合同接受的边界断言。
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
| `createUploadPresign` | `openapi/fixtures/Uploads/createUploadPresign.json` resume default；purpose/size/auth/IK failure 由 handler tests 与 E2E.P0.033 断言 | `frontend-resume-workshop/002-create-flow` 与 privacy export consumer；TargetJob 不消费 | `backend/internal/upload/handler/presign.go` | `file_objects` + object storage object；public purpose=`resume|privacy_export` | none | E2E.P0.033 + handler/store unit/integration tests |

Config dependency: A4 `objectStorage.*` / `OBJECT_STORAGE_*`、`upload.presignTTLSeconds` 与 `upload.maxBytes.{resume,privacyExport}` 保留；JD attachment 专属 maxBytes 必须删除。不得引入 `UPLOAD_*` 或 `OBJECT_STORE_*` 旁路。

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
- 通知 `frontend-resume-workshop/002-create-flow`：upload tab 可消费真实 backend presign；
- 通知 `backend-targetjob` owner：现有 fixture-backed file 流可保持不变，未来切真时调用相同 handler。

### Phase 6: L2 remediation hardening

#### 6.1 Register-time actual size enforcement

- `RegisterFileObject` 必须在 row lock 临界区内读取对象存储实际 size，并与 `file_objects.byte_size` 精确匹配；object missing 或 size mismatch 均返回 `VALIDATION_FAILED`，不得标记 `uploaded`。
- MinIO provider 可继续通过 signed PUT URL 接收对象，但 register path 必须以 `StatObject` 结果作为最终大小裁决，防止客户端低报 `byteSize` 后上传超限对象。

#### 6.2 Upload presign idempotency TTL alignment

- `createUploadPresign` route 的 idempotency cache TTL 必须与 `upload.presignTTLSeconds` 对齐；超过 signed URL TTL 的 retry 不得 replay 旧 `uploadUrl` / `expiresAt`。

#### 6.3 Runtime privacy delete wiring

- `cmd/api` 必须把 upload deleter 接入 backend runtime `privacy_delete` job runner kernel；`DELETE /api/v1/me` 创建的 privacy handoff job 在运行时必须调用 `DeleteFileObjectsForUser(userId)`。

#### 6.4 Atomic DB delete and audit tombstone

- 对象存储删除成功后，`file_objects` DB hard delete 与 audit tombstone 必须在同一 DB transaction 内提交；audit 写入失败不得让 row 先消失。

#### 6.5 Live scenario gate hardening

- E2E.P0.033 必须要求 `DATABASE_URL` 与 `OBJECT_STORAGE_*` live env；integration-tag tests 出现 skip 或 focused gate no-op 时 scenario 必须 fail，不得作为 Ready/PASS BDD 证据。
- `trigger.sh` 必须执行 `go test ./cmd/api -tags=integration -run TestUploadPresignRegisterPrivacyDeleteLiveRoundtrip -count=1 -v`，该测试必须覆盖真实 `POST /api/v1/uploads/presign` → MinIO signed `PUT` → internal `RegisterFileObject` → `DELETE /api/v1/me` → runtime runner kernel 处理 `privacy_delete` → DB hard delete + audit tombstone 的 live roundtrip。

### Phase 7: Remove JD attachment upload purpose

本批次依赖顺序固定为：统一 RED → B1/B3/OpenAPI 真理源与生成物 → A4/B4/F3/backend-upload/backend-async-runner 各自 owner surface → backend-targetjob Phase 18 集成 → BDD/全局 zero-reference。backend-upload 只在 OpenAPI/B4 purpose 收缩与 A4 Phase 12 maxBytes handoff 可消费后进入本 owner GREEN；任一上游 handoff 未完成时不得宣称本 Phase 完成。

#### 7.1 RED contract

统一 RED 分别由 B1/B3/OpenAPI、B4、A4 与 backend-upload owner tests 证明 JD attachment purpose 和专属 maxBytes 当前仍可达；backend-upload 自身只新增 purpose validation 与 handler/service 分支断言，同时固定 `resume` / `privacy_export` 正向基线。

#### 7.2 GREEN scope reduction

消费 OpenAPI/B4 的 purpose 收缩与 A4 Phase 12 maxBytes 删除 handoff；backend-upload 只删除自己拥有的 public purpose validation、handler/service 分支，不直接修改 A4 config/validator/composition 或 B4 DB constraint。`createUploadPresign` endpoint、file_objects state machine、resume register 与 privacy delete 保持不变。

#### 7.3 BDD and zero-reference

E2E.P0.033 继续验证 resume presign→PUT→register→privacy delete，并增加 JD attachment purpose 被拒绝与 privacy export purpose 仍合法的边界断言。zero-reference gate 覆盖 OpenAPI/generated/backend/migrations/config/fixtures/scripts，同时以正向断言防止误删 resume/privacy 能力。

### Phase 8: Typed upload defaults and exact size boundaries

Add missing/default/override/invalid tests for resume 10MiB and privacy_export 5MiB. Route A4 typed values into presign and register; no package-local fallback may differ from the shared code default. E2E.P0.033 must exercise exact limit/limit+1 for both purposes with actual object size verification and zero DB/object side effects on overflow. RuntimeConfig exposes only resume upload bytes for frontend P0.081; privacy export remains internal.

## 5 验收标准

- 本计划列出的 §4 所有 Phase task 全部完成
- §3 替代验证 gate 全部通过
- spec §6 C-1..C-9 全部 PASS
- BDD E2E.P0.033 PASS（含 setup → trigger → verify → cleanup 全脚本；trigger 必须包含 `TestUploadPresignRegisterPrivacyDeleteLiveRoundtrip` live roundtrip evidence）
- backend-resume owner 已收到 createUploadPresign 落地信号；backend-targetjob 当前 paste-only 合同不消费 upload handoff
- Phase 6 L2 remediation hardening 全部通过；若本机 live DB / MinIO env 未就绪，只能记录 scenario gate hardening 验证，不能把 E2E.P0.033 记为 live PASS。
- Phase 7 消费 OpenAPI/B4 purpose 与 A4 maxBytes 删除 handoff、完成 backend-upload 自有分支删除后，E2E.P0.033 与 resume/privacy focused tests 通过，旧 purpose 精确 zero-reference。
- Phase 8 missing/default/override/invalid 与 exact limit/+1 通过；resume 10MiB public projection 与 privacy 5MiB internal boundary 不漂移。

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: 对象存储 provider 选型未定 | ObjectStore interface 抽象 + 双实现（MinIO / filesystem）；A4 `objectStorage.provider` 配置可切换；P0 dev 用 MinIO，filesystem 仅供 unit test fallback |
| R2: presign URL 在 TTL 内泄漏导致越权 | TTL ≤ 600s + objectKey 含 userId 前缀；server-side 校验 register 时 fileObject.userId 匹配 |
| R3: state machine 复杂度与现有 file_objects.upload_status 字段对齐 | B4 baseline `file_objects.upload_status` 已存在；本 plan Phase 2.1 不改 schema，只补 store-layer validation |
| R4: privacy delete 对象存储失败 + DB 一致性 | retryable job + 保留 DB 原状态 + 重试上限；超过上限触发 P2 告警进入人工排查；不得私加 `deleted_pending` 状态 |
| R5: 历史 fixture/mock 让 TargetJob upload consumer 口径回流 | Phase 7 current operation matrix、E2E.P0.033 负向断言与全局 zero-reference 共同证明 TargetJob 不调用 upload endpoint；只保留 resume/privacy consumer |
