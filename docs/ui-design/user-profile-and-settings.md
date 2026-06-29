# 用户画像与设置目标结构

> **版本**: 1.8
> **状态**: deprecated
> **更新日期**: 2026-06-29

## 1 废弃结论

本文档历史上定义 `用户画像` 和 `设置与隐私` 的边界。当前 product-scope D-22 已删除用户画像：`profile` 不再是目标 route、用户菜单入口、OpenAPI tag、DB 表或前后端模块。

## 2 当前保留范围

`设置与隐私` 保留为用户菜单入口，只承载：

- 账号基础信息。
- 登录方式。
- 界面偏好和字体预设。
- 隐私数据控制。

`auth_profile_setup` 保留为首次账号资料补全页，但不沉淀 CandidateProfile 或 ExperienceCard。

## 3 实现约束

- 静态原型不得加载或渲染 `UserProfileScreen`。
- 正式前端用户菜单不得出现 `用户画像`。
- OpenAPI、backend、migrations、shared、config 和 scenario 的 CandidateProfile / ExperienceCard 实体由 product-scope/001-core-loop-module-pruning 删除。
- 恢复用户画像必须先重新修订 product-scope 和 UI 真理源，不得从本文档恢复。
