# Backend Upload History

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-05-12

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-12 | 1.2 | L2 review remediation 修订：upload presign idempotency TTL 与 signed URL TTL 对齐；RegisterFileObject 以对象存储实际 size 校验 declared byteSize；privacy_delete runtime 必须挂入 upload deleter；DB hard delete 与 audit tombstone 必须同事务提交；E2E.P0.033 不再允许 live DB / MinIO gate skip 后作为 PASS 证据。 | 001-file-objects-and-presign-baseline |
| 2026-05-12 | 1.1 | L1 plan-review 修订：对齐 `createUploadPresign` OpenAPI / fixture 成功状态码为 201，明确 backend-upload 001 只消费 B2 `default` fixture，failure/boundary 由 handler 与 BDD 断言覆盖；收敛 privacy delete 跨域顺序说明，避免 backend-upload 单独规定 `resume_assets` / `target_jobs` 全局删除顺序。 | 001-file-objects-and-presign-baseline |
| 2026-05-11 | 1.0 | 初始创建：从 engineering-roadmap 3.10 §5.2 派生 `backend-upload` (C2) subject，作为横向 file 上传基础设施 owner；锁定 D-1..D-7 决策（purpose enum / state machine / TTL / IK / register 路径 / 隐私删除 / 最大文件大小）；首批 plan `001-file-objects-and-presign-baseline` 覆盖 presign handler + file_objects store + state machine + privacy delete 链路。 | 001-file-objects-and-presign-baseline |
