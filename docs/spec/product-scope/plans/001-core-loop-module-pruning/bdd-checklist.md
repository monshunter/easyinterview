# Core Loop Module Pruning BDD Checklist

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## BDD.CORE.001 用户可见范围

- [x] 核心导航与页面代码不 materialize 已删除模块入口。
- [x] frontend behavior tests 与 scope lint 覆盖正向核心页面和旧入口负向断言。

## 分层回归

- [x] 根 `make test` 统一运行前后端全量单测。
- [x] 本计划不创建或复用 E2E ID，不把代码测试包装成场景证据。
