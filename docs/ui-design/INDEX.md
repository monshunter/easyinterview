# UI Design 文档索引

## 1 Active

| 文档 | 版本 | 状态 | 更新日期 | 说明 |
|------|------|------|----------|------|
| [目标总体架构](./ui-architecture.md) | 2.32 | active | 2026-07-14 | 三入口信息架构、Workspace list/detail 与 Parse command split、最小 custom accent、ReportsScreen 和安全 Markdown/GFM 连续文本面试 |
| [目标用户流程](./user-flow.md) | 2.28 | active | 2026-07-15 | 首页仅粘贴 JD、Parse command progress、Workspace 标题旁绑定简历与首行动作行、当前规划报告及连续文本面试流程 |
| [目标模块地图](./module-map.md) | 2.19 | active | 2026-07-14 | 三入口模块、Workspace list/detail、Parse command progress、最小 custom accent、target-scoped ReportsScreen 与安全 Markdown/GFM Interview Session |
| [认证与默认入口](./auth-and-entry.md) | 1.24 | active | 2026-07-14 | 邮箱验证码登录、首次账号资料补全、Workspace detail / Parse command / Reports targetJobId-only pendingAction 与范围外入口负向 |
| [当前面试规划目标模块](./module-job-workspace.md) | 1.39 | active | 2026-07-15 | `/workspace` 无参列表与 targetJobId 只读详情；标题旁绑定简历链接，立即面试/面试报告位于左对齐首行动作行；无独立绑定 block |
| [模拟面试与报告目标模块](./module-practice-review.md) | 1.31 | active | 2026-07-15 | 全宽连续文本聊天、持久回复状态、Workspace-detail recovery，以及准备度/summary 下沉到面试总评的 grounded 报告闭环 |
| [报告仪表盘目标结构](./report-dashboard.md) | 1.35 | active | 2026-07-15 | reportId-only frozen detail、honest wait、desktop `3/2/2/2/1`、mobile 同序单列及 Workspace 首行动作行报告入口 |
| [简历一级模块](./resume-module.md) | 3.7 | active | 2026-07-15 | 响应式简历卡片网格、desktop 固定最大列宽/mobile 单列、上传/粘贴创建、closed 摘要与只读来源格式详情边界（D-20） |
| [多 JD 与多简历管理](./jd-resume-management.md) | 3.7 | active | 2026-07-15 | Home POST 绑定 ready Resume；Resume 响应式卡片；Workspace 标题旁绑定简历详情链接与立即面试/面试报告首行动作行 |
| [首次无简历引导](./resume-onboarding.md) | 1.15 | active | 2026-07-10 | 上传/粘贴简历、注册成功后直接打开只读详情、LLM-derived displayName、上传文件正文提取和范围外 onboarding / 轻量问答边界（D-20） |

## 2 参考

暂无。本目录定义当前 UI 架构、流程和交互约束；正式实现位于 `frontend/`，项目不维护重复的可运行 UI Demo，也不引入外部品牌设计系统替代当前产品设计。
