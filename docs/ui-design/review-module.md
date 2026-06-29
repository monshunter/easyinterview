# 复盘目标模块

> **版本**: 2.8
> **状态**: deprecated
> **更新日期**: 2026-06-29

## 1 废弃结论

本文档记录的一级 `复盘` 模块已被 product-scope D-22 删除。`debrief` / `debrief_full` 不再是目标 route、顶部导航、OpenAPI tag、DB 表、AI feature key、shared event/job 或 E2E 正向场景。

## 2 当前替代路径

真实面试复盘没有替代的 P0 产品路径。当前核心闭环只保留：

```text
JD / 简历
  -> 模拟面试
  -> 报告
  -> 复练当前轮 / 进入下一轮
```

## 3 实现约束

- 静态原型不得再加载或渲染 `DebriefFullScreen`。
- 正式前端不得注册 `debrief` RouteName 或 TopBar 入口。
- OpenAPI、backend、migrations、shared、config 和 scenario 的复盘实体由 product-scope/001-core-loop-module-pruning 删除。
- 恢复复盘必须先重新修订 product-scope 和 UI 真理源，不得从本文档恢复。
