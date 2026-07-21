# UI Design 文档索引

## 1 Active

| 文档 | 版本 | 状态 | 更新日期 | 说明 |
|------|------|------|----------|------|
| [目标总体架构](./ui-architecture.md) | 2.43 | active | 2026-07-21 | JD/简历运行中等待态无内联返回；共享异步场景 action 保持可选，失败恢复不变 |
| [目标用户流程](./user-flow.md) | 2.32 | active | 2026-07-19 | 设置主题本地预览/单次保存/重登恢复；页面切换零重复 `/me`；其余核心流程保持 |
| [目标模块地图](./module-map.md) | 2.24 | active | 2026-07-20 | 账号级 Ocean / Plum / Forest Appearance、圆形 E 设置入口、Practice 双层 header 与 Workspace/Reports 边界 |
| [认证与默认入口](./auth-and-entry.md) | 1.34 | completed | 2026-07-20 | 单一设置入口、Ocean / Plum / Forest 账号主题、bootstrap 单次读取、设置页和 pendingAction 边界 |
| [当前面试规划目标模块](./module-job-workspace.md) | 1.49 | completed | 2026-07-20 | Home 空规划时隐藏整个 recent section；非空记录与 Workspace 共用业务 mapper |
| [模拟面试与报告目标模块](./module-practice-review.md) | 1.35 | active | 2026-07-19 | Practice 保留 global App TopBar，并在其下使用独立 Session Header 与全宽连续文本聊天 |
| [报告仪表盘目标结构](./report-dashboard.md) | 1.41 | active | 2026-07-16 | reportId-only frozen detail、任意报告状态保留面试记录、failed recovery、desktop `4/2/2/2/1` 与 mobile 同序单列 |
| [简历一级模块](./resume-module.md) | 4.8 | active | 2026-07-21 | 简历 queued/processing 等待态移除内联返回；失败态与普通详情返回保持 |
| [多 JD 与多简历管理](./jd-resume-management.md) | 4.3 | active | 2026-07-21 | JD queued/processing 等待态移除内联返回；失败/错误恢复与 ready replace 保持 |
| [首次无简历引导](./resume-onboarding.md) | 1.17 | active | 2026-07-15 | 上传/粘贴创建 selectable 简历；形成可读证据后返回 Home 显式选择，禁止跳过进入训练或报告链路 |

## 2 参考

暂无。本目录定义当前 UI 架构、流程和交互约束；正式实现位于 `frontend/`，项目不维护重复的可运行 UI Demo，也不引入外部品牌设计系统替代当前产品设计。
