# UI Design 文档索引

## 1 Active

| 文档 | 版本 | 状态 | 更新日期 | 说明 |
|------|------|------|----------|------|
| [目标总体架构](./ui-architecture.md) | 2.33 | active | 2026-07-15 | 三入口信息架构、单一设置齿轮与真实账号/隐私单页、固定字体栈、Workspace/Reports/Practice 边界 |
| [目标用户流程](./user-flow.md) | 2.29 | active | 2026-07-15 | 首页仅粘贴 JD、Workspace/Reports/Practice 主线，以及设置齿轮到真实账号/隐私单页流程 |
| [目标模块地图](./module-map.md) | 2.20 | active | 2026-07-15 | 三入口模块、单一设置齿轮、真实账号/隐私单页、固定字体、Workspace/Reports 与安全 Markdown/GFM Interview Session |
| [认证与默认入口](./auth-and-entry.md) | 1.25 | active | 2026-07-15 | 邮箱验证码登录、单一设置齿轮、真实账号/隐私单页、账号删除状态与 pendingAction 边界 |
| [当前面试规划目标模块](./module-job-workspace.md) | 1.39 | active | 2026-07-15 | `/workspace` 无参列表与 targetJobId 只读详情；标题旁绑定简历链接，立即面试/面试报告位于左对齐首行动作行；无独立绑定 block |
| [模拟面试与报告目标模块](./module-practice-review.md) | 1.31 | active | 2026-07-15 | 全宽连续文本聊天、持久回复状态、Workspace-detail recovery，以及准备度/summary 下沉到面试总评的 grounded 报告闭环 |
| [报告仪表盘目标结构](./report-dashboard.md) | 1.37 | active | 2026-07-15 | reportId-only frozen detail、honest wait、desktop `3/2/2/2/1`、mobile 同序单列及设置按钮 TopBar 响应式边界 |
| [简历一级模块](./resume-module.md) | 3.7 | active | 2026-07-15 | 响应式简历卡片网格、desktop 固定最大列宽/mobile 单列、上传/粘贴创建、closed 摘要与只读来源格式详情边界（D-20） |
| [多 JD 与多简历管理](./jd-resume-management.md) | 3.7 | active | 2026-07-15 | Home POST 绑定 ready Resume；Resume 响应式卡片；Workspace 标题旁绑定简历详情链接与立即面试/面试报告首行动作行 |
| [首次无简历引导](./resume-onboarding.md) | 1.15 | active | 2026-07-10 | 上传/粘贴简历、注册成功后直接打开只读详情、LLM-derived displayName、上传文件正文提取和范围外 onboarding / 轻量问答边界（D-20） |

## 2 参考

暂无。本目录定义当前 UI 架构、流程和交互约束；正式实现位于 `frontend/`，项目不维护重复的可运行 UI Demo，也不引入外部品牌设计系统替代当前产品设计。
