# Resume Workshop Create Flow BDD Plan

> **版本**: 1.16
> **状态**: active
> **更新日期**: 2026-07-20

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.RESUME.CREATE.001` | 用户在 desktop/mobile 选择 upload/paste；输入、类型、大小或请求也可能失败 | 创建、修正或重试 resume | 参考图层级的 Header/tab/dropzone/CTA 在两类 viewport 可达；UI 通过 API 进入 readonly detail/waiting；失败时保留可恢复输入，且不提前 presign/register 或泄露原文 | `frontend/src/app/screens/resume-workshop/create/CreateFlowIntegration.test.tsx` + `ResumeCreateFlow.test.tsx` + responsive source gate，由根 `make test` 承接 |
| `BDD.RESUME.CREATE.DROP.002` | 用户在 desktop 将一个文件拖到 Upload 区，或在 mobile/keyboard 使用“选择文件”；也可能拖入多个、非法格式或超限文件 | 进入/离开 drag-active，松开文件，或修正后重试 | 中央大 icon 缺席；drag-active 明确提示“松开以上传”并在离开/drop 后复位；有效单文件完成既有 presign/PUT/register/direct-detail；失败路径保留可恢复页面且零 presign/register；mobile 无横向溢出 | `frontend/src/app/screens/resume-workshop/create/UploadTab.test.tsx` domain behavior tests；current-run Chrome 仅作 UI 证据 |

当前没有覆盖 upload/paste create flow 的真实 API/UI E2E owner。代码层回归统一由根 `make test` 承接，不能作为 E2E 证据。
