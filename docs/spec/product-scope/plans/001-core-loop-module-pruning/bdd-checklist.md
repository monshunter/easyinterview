# Core Loop Module Pruning BDD Checklist

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-19

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## BDD.CORE.001 用户可见范围

- [x] 核心导航与页面代码不 materialize 已删除模块入口。
- [x] frontend behavior tests 与 scope lint 覆盖正向核心页面和旧入口负向断言。

## 分层回归

- [x] 根 `make test` 统一运行前后端全量单测。
- [x] 本计划不创建或复用 E2E ID，不把代码测试包装成场景证据。

## BDD.CORE.SETTINGS.001 账号主题边界

- [x] Product/UI/frontend 当前合同一致声明 TopBar 零主题菜单、Settings Appearance 账号级主题和“设置”页面名称。（2026-07-19 focused current-doc semantic test PASS。）
- [x] Frontend domain tests 与真实 `E2E.P0.101` current run 分别证明失败恢复/请求预算和保存后重登恢复；不把代码测试包装成 E2E。（domain focused 45 PASS；真实 run `e2e-p0-101-20260719082610-75505` PASS + cleanup PASS。）
