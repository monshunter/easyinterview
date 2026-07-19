# Resume Listing and Readonly Detail BDD Checklist

> **版本**: 2.10
> **状态**: completed
> **更新日期**: 2026-07-19

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.RESUME.READ.001` 列表与只读详情

- [x] Owner behavior tests 覆盖 list/detail、waiting/failure、删除、route 与隐私。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。
- [x] 当前无 list/detail/delete 真实 E2E owner；不创建 wrapper 场景。

## `BDD.RESUME.LIST.002` 语义卡片列表

- [x] Owner behavior tests 覆盖语义 card/list、closed 摘要字段与 table/header/row 负向语义；参考稿几何由 `BDD.RESUME.LIST.VISUAL.003` owner。（验证：list/fixture/StrictMode tests PASS）
- [x] 卡片“打开”只导航 `resume_versions?resumeId=...`；删除成功隐藏、失败保留并显示可恢复错误；loading/empty/error/pagination 不回退。（验证：Resume Workshop 20 files / 118 tests PASS）
- [x] 根 `make test` 与独立 responsive/a11y gates 通过；该结果是代码层行为证据，不声明真实 E2E PASS。（验证：Chrome 1440/390 geometry+screenshot；根回归 PASS）

## `BDD.RESUME.LIST.VISUAL.003` 简历列表参考稿层级

- [x] 确认验证入口为 Resume list/CSS domain behavior tests 与 current-run Chrome UI acceptance，不创建 E2E wrapper。
- [x] 执行 owner tests，验证 desktop 双列等宽卡、mobile 满宽单列、与 Workspace 一致的 22px circled-plus create icon、文件 icon、名称/摘要、meta、删除与 footer 打开层级。（验证：create icon RED 后 focused 8 tests PASS；其余 Resume list/CSS owner tests PASS）
- [x] 执行 create/open/archive success/failure 与 loading/empty/error/pagination 回归，确认 route/generated client/closed DTO 不变。（验证：owner scope 24 files / 150 tests PASS）
- [x] 记录 1916×821 / 390×844 bbox、circled-plus 一致性、截图、keyboard、theme、console 与 no-overflow 证据。（验证：Resume/Workspace 图标均为 22×22、同一 viewBox/圆/十字路径/1.8 线宽；desktop 双列 690px，mobile 358px 单列，overflow 0；截图位于 `.test-output/list-ui-acceptance/`）
