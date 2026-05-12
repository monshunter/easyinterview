# 001 BDD Checklist

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-12

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.033 file presign → register → delete roundtrip

- [x] 创建场景目录 `test/scenarios/e2e/p0-033-file-presign-register-roundtrip/`，含 `README.md`（§6 baseline + §7 离线限制）+ `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 B2 `createUploadPresign.default` fixture + scenario 数据：A2 dev stack MinIO 拉起；A4 config 注入；2 个测试用户（A / B）；3 个测试 file binary（小 PDF / 5MB binary / 11MB 超限）；failure / boundary 走直接断言，不声明不存在的 B2 error fixture
- [x] 实现 `scripts/setup.sh`（A2 dev stack 拉起 + MinIO 健康检查 + 测试用户登录 + 测试文件准备）/ `scripts/trigger.sh`（依序触发 A/B/C/D/E 子场景）/ `scripts/verify.sh`（断言 DB state machine + 对象存储 key / IK replay 不变 / cross-user 隔离 / privacy delete tombstone / 隐私负向 grep）/ `scripts/cleanup.sh`（清理 file_objects / 用户 / MinIO bucket / audit_events）
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录验证证据：`.test-output/e2e/p0-033-file-presign-register-roundtrip/trigger.log` + verify 输出 + `createUploadPresign.default` 201 fixture byte diff 0 + DB state machine 轨迹 + 对象存储 key list before/after + privacy delete audit tombstone 内容 + 隐私反查日志
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.033 行（关联需求 `backend-upload C-1, C-2, C-3, C-4, C-6, C-7, C-8`，状态 Ready，automated）
