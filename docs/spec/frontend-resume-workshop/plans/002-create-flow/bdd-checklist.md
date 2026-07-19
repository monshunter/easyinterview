# Resume Workshop Create Flow BDD Checklist

> **版本**: 1.16
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.RESUME.CREATE.001` Upload/Paste 创建

- [x] Owner behavior tests 覆盖 upload/paste、direct-detail、输入恢复、类型/大小 guard 与隐私。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。
- [x] 当前无 create-flow 真实 E2E owner；不创建 wrapper 场景。
- [x] Reference revision：owner tests 覆盖 desktop/mobile Header、tab、dropzone、CTA 与 capability labels，同时保留 upload/paste 行为、错误恢复与隐私。
- [x] 根 `make test` 与 Chrome 1916×821 / 390×844 视图验收完成；Chrome 证据不声明 E2E PASS。<!-- verified: 2026-07-19 evidence="root frontend 131 files/1054 tests PASS; formal real-mode frontend create flow has zero document overflow at both viewports" -->
