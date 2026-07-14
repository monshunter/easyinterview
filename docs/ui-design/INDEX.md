# UI Design 文档索引

## 1 Active

| 文档 | 版本 | 状态 | 更新日期 | 说明 |
|------|------|------|----------|------|
| [目标总体架构](./ui-architecture.md) | 2.32 | active | 2026-07-14 | 三入口信息架构、Workspace list/detail 与 Parse command split、最小 custom accent、ReportsScreen 和安全 Markdown/GFM 连续文本面试 |
| [目标用户流程](./user-flow.md) | 2.27 | active | 2026-07-14 | 首页仅粘贴 JD、Parse command progress、Workspace targetJobId 详情回访、当前规划报告与连续文本面试流程 |
| [目标模块地图](./module-map.md) | 2.19 | active | 2026-07-14 | 三入口模块、Workspace list/detail、Parse command progress、最小 custom accent、target-scoped ReportsScreen 与安全 Markdown/GFM Interview Session |
| [认证与默认入口](./auth-and-entry.md) | 1.24 | active | 2026-07-14 | 邮箱验证码登录、首次账号资料补全、Workspace detail / Parse command / Reports targetJobId-only pendingAction 与范围外入口负向 |
| [当前面试规划目标模块](./module-job-workspace.md) | 1.37 | active | 2026-07-14 | `/workspace` 无参列表与 targetJobId 只读详情；ready 卡片直达，Parse 仅新导入进度并 ready replace；Reports Back 返回详情 |
| [模拟面试与报告目标模块](./module-practice-review.md) | 1.30 | active | 2026-07-14 | 全宽连续文本聊天、安全 Markdown/GFM、持久回复状态、same-ID retry、Workspace-detail terminal recovery 与 grounded 报告闭环 |
| [报告仪表盘目标结构](./report-dashboard.md) | 1.33 | active | 2026-07-14 | target-scoped current/latest ReportsScreen、Workspace-detail Back、reportId-only frozen detail、honest wait 与 desktop+390 parity |
| [简历一级模块](./resume-module.md) | 3.5 | active | 2026-07-10 | 平铺简历列表、上传/粘贴创建后直接打开详情、LLM-derived displayName、fallback 命名、禁止 raw 第一行/文件名命名、上传文件正文提取、只读原始正文、无导出/复制/编辑/改写/原件弹层和范围外入口边界（D-20） |
| [多 JD 与多简历管理](./jd-resume-management.md) | 3.6 | active | 2026-07-14 | Home POST 绑定 ready Resume，Parse 仅进度、Workspace targetJobId 详情只读绑定；最近卡片与 Interview list 共用动作主体，Resume 上传/粘贴独立保留 |
| [首次无简历引导](./resume-onboarding.md) | 1.15 | active | 2026-07-10 | 上传/粘贴简历、注册成功后直接打开只读详情、LLM-derived displayName、上传文件正文提取和范围外 onboarding / 轻量问答边界（D-20） |

## 2 参考

暂无。前端视觉实现只以本目录文档和 `ui-design/` 静态原型为真理源；外部品牌设计系统不是项目 UI 参考。
