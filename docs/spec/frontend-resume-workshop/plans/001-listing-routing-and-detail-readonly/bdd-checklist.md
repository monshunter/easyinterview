# Resume Listing and Readonly Detail BDD Checklist

> **版本**: 2.8
> **状态**: active
> **更新日期**: 2026-07-15

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.RESUME.READ.001` 列表与只读详情

- [x] Owner behavior tests 覆盖 list/detail、waiting/failure、删除、route 与隐私。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。
- [x] 当前无 list/detail/delete 真实 E2E owner；不创建 wrapper 场景。

## `BDD.RESUME.LIST.002` 响应式卡片列表

- [ ] Owner behavior tests 覆盖 desktop 固定最大列宽/左对齐、mobile 单列/no-overflow、closed 摘要字段与 table/header/row 负向语义。
- [ ] 卡片“打开”只导航 `resume_versions?resumeId=...`；删除成功隐藏、失败保留并显示可恢复错误；loading/empty/error/pagination 不回退。
- [ ] 根 `make test` 与独立 responsive/a11y gates 通过；该结果是代码层行为证据，不声明真实 E2E PASS。
