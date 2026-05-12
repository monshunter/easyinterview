# 001 BDD Plan

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-12

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 类别 | 关联 Phase | 关联 Spec C-* | 关联 BDD-Gate（主 checklist） |
|---------|------|-----------|--------------|----------------------------|
| E2E.P0.033 | primary + boundary + failure · file presign → 客户端 PUT → register → privacy delete roundtrip | Phase 1 + 2 + 3 + 4 + 5 | C-1, C-2, C-3, C-4, C-6, C-7, C-8 | Phase 5.3 |

---

## Phase 1 + 2 + 3 + 4 + 5: file presign / register / delete roundtrip

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.033 | file presign → 客户端 PUT → register 校验 → privacy delete 完整 roundtrip + IK replay + purpose 校验 + cross-user 隔离 | A2 dev stack 已拉起（含 MinIO 或等价对象存储）；用户 A 已登录；A4 config paths 已就位：`objectStorage.provider=minio` + `upload.presignTTLSeconds=600` + `upload.maxBytes.resume=10485760`，并复用现有 `OBJECT_STORAGE_*` endpoint/bucket/credential；分多子场景：（A）happy path：`purpose=resume` + 有效 IK；（B）IK replay：同 IK 重复调用；（C）purpose 非法：`purpose=unknown`；（D）跨用户：用户 B 尝试 register 用户 A 的 fileObjectId；（E）privacy delete：用户 A 有 3 个 fileObject + 调 `DELETE /api/v1/me` 触发 privacy_delete job | （A1）`POST /api/v1/uploads/presign` body `{purpose: resume, fileName: resume.pdf, contentType: application/pdf, byteSize: 1048576}` + `Idempotency-Key: <uuid>`；（A2）客户端拿到 `uploadUrl` 后按 response `method/headers` `PUT <uploadUrl>` with 1MB binary；（A3）业务 handler（test-only harness 或 mocked registerResume）调用 backend-upload internal `RegisterFileObject(fileObjectId, expectedPurpose=resume, ownerUserId=A)`，再由业务表自行写入 FK；（B1）相同 IK 重复 POST；（C1）`purpose=unknown` POST；（D1）用户 B 登录后调 internal `RegisterFileObject(用户 A 的 fileObjectId, resume, B)`；（E1）用户 A 调 `DELETE /api/v1/me`，触发 privacy_delete job | （A1）返回 201 + `UploadPresign{fileObjectId, uploadUrl, method, headers, expiresAt=now+600s}`，且与 B2 fixture `default` scenario 字节一致；DB `file_objects` 行 `upload_status='pending'` / `user_id=A` / `purpose='resume'` / `object_key='{A.id}/resume/{fileObjectId}.pdf'`；audit_events 不写入（presign 不算 audit 事件）；（A2）对象存储 PUT 接受，DB 可保持 `pending` 直到业务 register 完成确认；（A3）`RegisterFileObject` 调 `ObjectStore.Exists(objectKey)` 确认对象存在后原子标记 `upload_status='uploaded'` 并校验通过，但不写 `registered` 状态；业务 FK 写入由 backend-resume / backend-targetjob owner 负责；（B1）返回与 A1 相同的 fileObjectId + uploadUrl/method/headers/expiresAt（或新签名但 fileObjectId 不变）；不创建新 DB 行（count(*) before/after 一致）；（C1）返回 422 + `error.code = "VALIDATION_FAILED"` + `details.field = "purpose"`；不创建 DB 行；（D1）返回 404 或 422；用户 B 不可见用户 A 的 fileObject；audit_events 不暴露 fileObjectId；（E1）backend-upload file_objects 步骤先对象存储 hard delete 3 file，再 DB hard delete 3 行，再写 audit tombstone（含 fileObjectId / purpose / deletedAt，不含 objectKey）；对象删除失败时 DB 行保留原状态等待 retry；`resume_assets` / `target_jobs` 等跨域删除顺序由 B4 matrix 和对应 owner 链路协调；（F 隐私）file binary content 不出现在 console / URL（除 uploadUrl 已签发）/ localStorage / log；fixture transport 日志不泄漏 secret key；（G 旧口径）grep `frontend-` / `backend-` 不出现旧 `upload-route-frontend-signed` / hardcode S3 SDK 直接 import 等 retired 模式（除 `internal/upload/objectstore/minio.go` 中合法 import） | `test/scenarios/e2e/p0-033-file-presign-register-roundtrip/` |
