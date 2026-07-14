# File Objects and Presign BDD Plan

> **版本**: 1.9
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.UPLOAD.FILE.001` | 用户选择受支持文件；metadata、ownership、size 或 object 状态也可能非法 | 请求 presign、上传、register 或读取 | 只登记已写入且属于当前用户的 object；非法输入 fail closed 且不产生可见脏记录 | `backend/internal/upload/handler/presign_test.go` + `backend/internal/upload/service/register_test.go`，由根 `make test` 承接 |

当前没有覆盖 presign/upload/register roundtrip 的真实 API/UI E2E owner。代码层回归统一由根 `make test` 承接，不能作为 E2E 证据。
