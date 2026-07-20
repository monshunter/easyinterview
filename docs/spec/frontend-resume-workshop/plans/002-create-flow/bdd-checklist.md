# Resume Workshop Create Flow BDD Checklist

> **版本**: 1.17
> **状态**: active
> **更新日期**: 2026-07-20

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.RESUME.CREATE.001` Upload/Paste 创建

- [x] Owner behavior tests 覆盖 upload/paste、direct-detail、输入恢复、类型/大小 guard 与隐私。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。
- [x] 当前无 create-flow 真实 E2E owner；不创建 wrapper 场景。
- [x] Reference revision：owner tests 覆盖 desktop/mobile Header、tab、dropzone、CTA 与 capability labels，同时保留 upload/paste 行为、错误恢复与隐私。
- [x] 根 `make test` 与 Chrome 1916×821 / 390×844 视图验收完成；Chrome 证据不声明 E2E PASS。<!-- verified: 2026-07-19 evidence="root frontend 131 files/1054 tests PASS; formal real-mode frontend create flow has zero document overflow at both viewports" -->

## `BDD.RESUME.CREATE.DROP.002` 单文件拖放上传

- [x] 确认验证入口为 `UploadTab.test.tsx` domain behavior tests 与 current-run Chrome scoped UI acceptance，不创建 E2E wrapper。<!-- verified: 2026-07-20 method=owner-routing evidence="BDD.RESUME.CREATE.DROP.002 remains code-level plus scoped Chrome evidence only" -->
- [x] 执行 dragenter/dragover/dragleave/drop 行为断言，验证 active copy/style 复位与中央 icon 零残留。<!-- verified: 2026-07-20 method=focused-vitest evidence="UploadTab + ResumeCreateVisual 2 files / 13 tests PASS" -->
- [x] 执行有效单文件 drop 的 presign/PUT/register/direct-detail 断言，以及多文件、非法格式、超限、disabled/submitting 零请求断言。<!-- verified: 2026-07-20 method=focused-vitest evidence="valid dropped.md handoff and local failure guards PASS" -->
- [ ] 使用 Chrome skill 在真实 frontend/backend 记录 desktop default/drag-active、有效 drop handoff、mobile no-overflow、截图与 console；该证据不声明 E2E PASS。<!-- partial: 2026-07-20 evidence="Desktop/mobile default UI, zero legacy icon, containment and zero warning/error pass on the combined branch. Local file injection is blocked by Chrome extension permission; no live drop handoff PASS is claimed." -->
