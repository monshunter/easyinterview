# Backend Upload Spec

> **版本**: 1.7
> **状态**: active
> **更新日期**: 2026-07-14

## 1 背景与目标

[engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 标记 Resume Workshop 候选 subject 依赖 `C2 backend-upload`（roadmap 3.10 已显式化）。本 subject 是当前仓库的横向 file 上传基础设施 owner：把 `POST /api/v1/uploads/presign` 真实落地为 backend handler，管理 `file_objects` 表 CRUD 与 state machine，并提供 `purpose` 枚举强约束 + 隐私 / 留存策略。

backend-upload 之所以独立于 `backend-resume`：`file_objects` 同时服务简历资产、privacy export 与 DB-local 内部对象。JD 当前只允许粘贴文本，不消费上传基础设施；public upload purpose 只保留 `resume` / `privacy_export`。把文件基础设施收敛到一个 owner 可避免重复实现，同时防止删除 JD 附件能力时误伤简历与隐私链路。

目标：

1. **唯一上传入口**：所有业务文件上传都必须通过 `POST /api/v1/uploads/presign` 获取签名 URL + register 引用，不允许业务 handler 自建 presign 逻辑。
2. **file_objects state machine**：P0 复用当前 B4 baseline `upload_status`：`pending` → `uploaded`，失败为 `scan_failed`，隐私/清理终态为 `deleted`；业务 register 通过 backend-upload internal API 在确认对象已 PUT 后原子完成 `pending → uploaded`，再由业务表 FK（如 `resume_assets.file_object_id`）表达引用，不向 `file_objects.upload_status` 私自加入 `registered` / `deleted_pending`。
3. **purpose 强约束**：public `createUploadPresign` 只接受 `resume` / `privacy_export`；DB-local `source_snapshot` / `audio` / `video` 不构成 public API。新增 purpose 必须先修订 B2 + B4。
4. **隐私 / 留存策略**：对象存储 hard delete 与 DB 行 hard delete 同事务（[B4 §3.1.2 privacy deletion matrix](../db-migrations-baseline/spec.md#312-p0-privacy-deletion-table-matrix)）；presign URL TTL 与 secret 由 [A4 secrets-and-config](../secrets-and-config/spec.md) 锁定。
5. **mock-first 可用**：本 subject 的 `createUploadPresign` fixtures（B2 已存在）可被 `frontend-resume-workshop` / 未来 `frontend-onboarding` / `frontend-resume-workshop/002-create-flow` 直接消费，无需等真实 backend 落地。

本 spec 不实现具体业务文件解析（`resume_parse` 归 [backend-resume](../backend-resume/spec.md)；`target_import` 归 [backend-targetjob](../backend-targetjob/spec.md)）；不实现 backend internal runner（归 backend-runtime-topology 与未来 backend-async-runner）；不实现前端 UI（归各 D 域）。

## 2 范围

### 2.1 In Scope

- **HTTP handler**：`POST /api/v1/uploads/presign` (createUploadPresign) 真实业务逻辑，含 IK 校验、purpose 检查、签名 URL 生成与过期；future `getFileObject(fileObjectId)` 同 owner（[B2 D-18 范围](../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) 落地时引入）。
- **store layer**：`file_objects` 表 repository（Go side `backend/internal/upload/store/`），含创建（pending）、标记上传完成（uploaded）、上传扫描失败（scan_failed）、删除标记（deleted）、row lock / state validation 与 hard delete；业务 register completion 由 `backend/internal/upload/service` 组合 repository + `ObjectStore.Exists` 完成。
- **purpose enum**：本 spec §3.1 锁定值集；handler 拒绝未列举值（返回 `VALIDATION_FAILED`）。
- **state machine 校验**：仅允许 `pending → uploaded`、`pending|uploaded → scan_failed`、`pending|uploaded|scan_failed → deleted`；非法转换统一返回 B1 已登记 `VALIDATION_FAILED`（cross-user / not-found 仍返回 404）。业务 `RegisterFileObject` 对 `pending` 行必须先确认 object storage 中对象存在，再在同一业务临界区标记为 `uploaded`；已是 `uploaded` 时幂等通过。
- **隐私删除链路**：privacy_delete job 调用 backend-upload 提供的 `DeleteFileObjectsForUser(userId)` API；对象存储 hard delete 先执行，DB 行 hard delete 与 audit tombstone 必须在同一 DB 事务内提交（或整体失败保留可重试状态）。
- **mock-first fixtures**：B2 现有 `createUploadPresign.json` fixture 已就位；本 spec 不重复 owner，但通过 P0 happy-path BDD 场景与 fixture 对齐。
- **A4 配置依赖**：对象存储 endpoint / bucket / access key / secret key 使用 A4 现有 `objectStorage.*` / `OBJECT_STORAGE_*` 契约；presign TTL、per-purpose max bytes 与 provider selector 必须在实现前以 A4 additive config path 登记，本 spec 不 hardcode。

### 2.2 Out of Scope

- 业务文件解析（resume 解析归 backend-resume）；JD 文件导入不属于当前产品范围。
- 前端上传组件 UI（归 `frontend-shell` / `frontend-resume-workshop/002`）。
- 对象存储 provider 抉择（S3 / OSS / MinIO / 本地）：归 [A2 local-dev-stack](../local-dev-stack/spec.md) 与 [A4 secrets-and-config](../secrets-and-config/spec.md) 决策。
- 文件病毒扫描 / 内容安全扫描：P0 不实现；P1 评估。
- presign URL 公开访问 CDN / 大文件分片上传：P0 不实现。
- 已上传未 register 的 GC（pending → deleted 自动回收）：P1 引入 backend internal runner 时再设计。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | public purpose enum 锁定集 | `resume`（resume_assets.file_object_id 引用）/ `privacy_export`（privacy export 输出物）；`source_snapshot` / `audio` / `video` 为 DB-local 内部值，不开放 public presign | 业务调用 createUploadPresign 必传 purpose；JD 不使用该 endpoint；扩展需先修订 B2 与 B4 |
| D-2 | state machine 字面量 | `pending` / `uploaded` / `scan_failed` / `deleted`；存于 `file_objects.upload_status` text + check constraint | 与 B4 baseline `migrations/000001_create_baseline.up.sql` 对齐；本 spec 不引入 `registered` / `deleted_pending` |
| D-3 | presign URL TTL | 默认 600s（10 分钟）；通过 A4 additive config path `upload.presignTTLSeconds` 注入；当前不新增 env key，若未来需要 env override 必须先修订 A4 env 字典 | 客户端必须在 TTL 内完成 PUT；超时后重新调用 createUploadPresign |
| D-4 | IK 必带 | `createUploadPresign` 必带 `Idempotency-Key`（与 B2 D-6 / D-18 一致）；upload presign route 的 idempotency record TTL 必须与 `upload.presignTTLSeconds` 对齐，避免在 signed URL 过期后继续 replay stale `uploadUrl` / `expiresAt`；TTL 内重复请求返回同一 fileObjectId + uploadUrl/method/headers/expiresAt | 防止网络抖动产生孤儿 `file_objects` 行，同时避免过期 signed URL 被 response cache 继续 replay |
| D-5 | register 路径 | P0 不引入独立 `POST /api/v1/file-objects/{id}/register` endpoint；register 行为由业务 handler（如 `registerResume`）调用 backend-upload internal `RegisterFileObject(fileObjectId, expectedPurpose, ownerUserId)` 完成。该 API 必须锁定同 user + purpose 行，若状态为 `pending` 则先调用 `ObjectStore.Exists(objectKey)` 确认客户端 PUT 已落对象存储，再原子标记 `uploaded`；若状态已为 `uploaded` 则幂等通过；`scan_failed` / `deleted` 返回 `VALIDATION_FAILED` | 避免独立 HTTP endpoint；business owner 持有 file 引用语义；不新增 `registered` 状态，同时补齐 presign → PUT → business register 的公开上传完成路径 |
| D-6 | 隐私删除 | privacy_delete job 调用 `DeleteFileObjectsForUser(userId)`：先按 owner 反查 file_object 行 → 对象存储 hard delete → DB 行 hard delete；失败 retryable | 与 [B4 §3.1.2](../db-migrations-baseline/spec.md#312-p0-privacy-deletion-table-matrix) `file_objects` / `resume_assets` 行对齐 |
| D-7 | 最大文件大小 | 默认 10MiB（resume）/ 5MiB（privacy_export）；通过 `upload.maxBytes.resume` / `upload.maxBytes.privacyExport` 注入，并有同值 typed code defaults；缺 key 使用缺省，显式非正数启动失败。超限由 presign 拒绝 `VALIDATION_FAILED`，RegisterFileObject 以对象存储实际大小再次裁决 | 默认/override/invalid 只由 A4 typed owner 覆盖；upload owner 以小型注入值验证 handler/service 非平凡裁决，不用真实默认大小文件或场景环境重复证明配置 |

### 3.2 待确认事项

- 是否在 P1 引入 backend-upload 自身的内部 GC runner（`pending` 24h 未 `uploaded` 自动转 `deleted`）：默认 P0 不实现；如 production 观测到孤儿 fileObject 累积再决策。
- `getFileObject(fileObjectId)` HTTP endpoint 是否在 P0 引入：默认随后续 B2 additive plan 决定；本 spec 不预设。

## 4 设计约束

### 4.1 契约约束

- 必须实现 [B2 §3.1.1](../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) op #6 `createUploadPresign` 的 generated server interface；不允许私造 handler 签名。
- 必须复用 [B2 fixtures](../mock-contract-suite/spec.md) `openapi/fixtures/Uploads/createUploadPresign.json` 现有 scenario；如新增 scenario 必须在 B2 plan 修订时同步。
- handler 错误码必须 `$ref` [B1 D-5](../shared-conventions-codified/spec.md#31-已锁定决策) 中已锁错误码（`VALIDATION_FAILED` / `RATE_LIMITED` / `AUTH_*`）；不私造 `UPLOAD_*` 前缀错误码（如需要先修订 B1）。

### 4.2 存储约束

- `file_objects` 表行（[B4 baseline §2.1](../db-migrations-baseline/spec.md#21-in-scope)）必须含 `user_id`（FK users）、`purpose`、`upload_status`、`byte_size`、`content_type`、`object_key`、`original_file_name`、`created_at`、`updated_at`、`deleted_at`。
- `object_key` 命名格式：`{user_id}/{purpose}/{file_object_id}.{ext}`；不允许业务 owner 自定义 path。
- 对象存储 endpoint / bucket / credentials 由 A4 现有 `OBJECT_STORAGE_*` 配置注入；provider selector 使用 additive config path `objectStorage.provider`，实现前必须在 A4/config 层登记；本 spec 不绑定 S3 / OSS / MinIO 具体 SDK，handler 通过 `ObjectStore` interface 抽象。

### 4.3 隐私约束

- 对象存储删除失败时，DB 行保留原状态并由 privacy job retry；不允许 DB 先 hard delete 后留下对象存储孤儿。只有对象删除成功后才能标记 `deleted` 或 hard delete DB 行；DB hard delete 与 audit tombstone 必须同事务提交，防止 tombstone 写入失败后永久丢失删除证据。
- presign URL 必须由对象存储 provider 端签发 server-to-server，不允许把 secret key 透传到客户端。
- 上传内容（raw bytes）不允许在 handler / store / runner / log 中持久或缓存；只允许 `object_key` 与 metadata。
- 隐私删除事件不在 `audit_events` 中持久 `object_key`；只保留 audit tombstone（参考 [B4 §3.1.2](../db-migrations-baseline/spec.md#312-p0-privacy-deletion-table-matrix)）。

### 4.4 BDD / TDD 约束

- 任何 user-visible 修订（如新增 purpose / 改 TTL / 改 state machine）必须先修订本 spec 再落 plan；plan 必须维护 BDD scenario 覆盖 happy path + 至少 1 个 failure / boundary。
- 业务 handler / store 实现必须有 TDD 覆盖：handler unit test（IK replay / purpose validation / TTL 边界）、store integration test（state transition / cross-user isolation / cascade delete）。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `POST /api/v1/uploads/presign` handler | backend-upload | 真实业务逻辑 + IK 校验 + purpose check |
| `file_objects` 表 schema + migration | [B4 db-migrations-baseline](../db-migrations-baseline/spec.md) | 字段 / 索引 / FK / check constraint |
| `purpose` / `upload_status` 枚举字面量 | backend-upload（D-1/D-2）+ [B4 enum-sources.yaml](../db-migrations-baseline/spec.md#21-in-scope) 登记 | DB check constraint 必须与本 spec 对齐 |
| presign URL 签发 | backend-upload + 对象存储 provider | 通过 `ObjectStore` interface 抽象 |
| 对象存储 endpoint / secret / TTL | [A4 secrets-and-config](../secrets-and-config/spec.md) | 不 hardcode |
| 业务 register 引用 | [backend-resume](../backend-resume/spec.md) 等真实文件业务 owner | 通过 backend-upload internal API；TargetJob paste-only 不注册文件 |
| 隐私删除调用 | backend internal privacy runner（[backend-runtime-topology](../backend-runtime-topology/spec.md)） | 调用 backend-upload `DeleteFileObjectsForUser` |
| frontend 上传组件 | [frontend-shell](../frontend-shell/spec.md) / `frontend-resume-workshop/002-create-flow` | 消费 generated client 调 createUploadPresign + 客户端 PUT |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | presign 主路径 | 已登录用户 + 有效 `purpose=resume` + `Idempotency-Key` | 调 `POST /api/v1/uploads/presign`，body 含 `fileName / contentType / byteSize` | 返回 201 + `UploadPresign{fileObjectId, uploadUrl, method, headers, expiresAt}`；DB 创建 `file_objects` 行 `upload_status='pending'` | 001-file-objects-and-presign-baseline |
| C-2 | IK replay | 同 IK 重复调用，且仍在 `upload.presignTTLSeconds` TTL 内 | 同 IK 第二次 | 返回首次 `fileObjectId` + 同一 uploadUrl/method/headers/expiresAt；不创建新 DB 行；超过 presign TTL 后重新执行 presign path，并返回新的 uploadUrl/method/headers/expiresAt | 001 |
| C-3 | purpose 非法 | `purpose='unknown_purpose'` | 调 presign | 422 + `error.code = "VALIDATION_FAILED"` + `details.field = "purpose"` | 001 |
| C-4 | 跨用户隔离 | 用户 A 创建 fileObject；用户 B 调 `getFileObject` 或 register | – | 404（不暴露存在；与 B2 D-15 envelope 对齐） | 001 |
| C-5 | state transition 非法 | DB 行 `upload_status='deleted'` 或 `scan_failed` | business handler 调 internal RegisterFileObject | 422 + `error.code = "VALIDATION_FAILED"`；不使用未登记的状态迁移专用错误码 | 001 |
| C-6 | privacy delete 链路 | 用户 A 有 5 个 fileObject + 1 个 privacy_export pending | `DELETE /api/v1/me` 创建 `privacy_delete` async job，backend runtime runner kernel 执行该 job | 对象存储删除 5 行 → DB 5 行硬删 + audit tombstone 同事务写入；retryable 失败时 DB 行保留原状态等待重试；upload deleter 必须被实际挂入 `cmd/api` runtime privacy_delete path | 001（含隐私章节）+ backend-runtime-topology |
| C-7 | 隐私 / 范围外输入负向 | grep `frontend-` / `backend-` / `docs/spec/` | – | 不出现范围外 `upload-route` / `pre-signed-by-frontend` / hardcode S3 SDK 路径等 out-of-scope 模式 | 001 |
| C-8 | mock-first 对齐 | B2 fixture `createUploadPresign.json` `default` scenario | mock-server 返回该 scenario | 字段集 / status code / IK 行为与真实 handler 字节级一致 | 001 + mock-contract-suite |
| C-9 | JD purpose 收缩 | OpenAPI/DB/config/backend 仍可能存在 JD attachment purpose | 执行 001 Phase 7 | JD attachment purpose 与专属 maxBytes zero-reference；`resume` / `privacy_export` presign、register 与 privacy delete 回归通过 | 001 |
| C-10 | Config owner handoff | A4 typed loader/validator 已覆盖 maxBytes default/override/invalid | presign/register 注入小型 limit | overflow 在 presign/object-size 裁决处零 DB/object side effect；本域不重复默认大小、RuntimeConfig 传播或配置专属场景 | A4 001 Phase 13 + 001 Phase 8 focused tests |

## 7 关联计划

- [001-file-objects-and-presign-baseline](./plans/001-file-objects-and-presign-baseline/plan.md)（active）：保留 resume/privacy presign、register、删除链路与当前无真实 E2E owner 的 BDD 行为合同。
