# UI Design 文档索引

## 1 Active

| 文档 | 版本 | 状态 | 更新日期 | 说明 |
|------|------|------|----------|------|
| [目标总体架构](./ui-architecture.md) | 2.30 | active | 2026-07-14 | 三入口信息架构、规划范围 ReportsScreen、desktop/mobile 响应式 TopBar、连续文本面试与 reportId-only 报告详情 |
| [目标用户流程](./user-flow.md) | 2.26 | active | 2026-07-14 | 首页仅粘贴 JD、规划回访与当前规划报告、连续文本面试、会话级报告、简历与认证设置流程 |
| [目标模块地图](./module-map.md) | 2.17 | active | 2026-07-14 | 三入口模块、target-scoped ReportsScreen、连续文本 Interview Session、reportId-only Report Dashboard 与当前数据依赖 |
| [认证与默认入口](./auth-and-entry.md) | 1.23 | active | 2026-07-14 | 默认进首页后的邮箱验证码登录、首次账号资料补全、opaque pending import 与 reports targetJobId-only 接续、三入口用户菜单和范围外入口负向 |
| [当前面试规划目标模块](./module-job-workspace.md) | 1.36 | active | 2026-07-14 | 面试一级 `workspace` 是纯规划列表页；规划详情右上角进入当前规划报告，Parse 无嵌入列表；Home JD 仅粘贴并绑定 ready Resume |
| [模拟面试与报告目标模块](./module-practice-review.md) | 1.28 | active | 2026-07-13 | 全宽连续文本聊天、即时 user row、持久回复状态、pending 思考/输入锁定、retryable-only 行内重试、零回答完成门禁与 grounded 会话级报告闭环 |
| [报告仪表盘目标结构](./report-dashboard.md) | 1.32 | active | 2026-07-14 | target-scoped current/latest ReportsScreen、reportId-only frozen detail、可信 Back、内部 locator 清理、honest wait 与 desktop+390 parity |
| [简历一级模块](./resume-module.md) | 3.5 | active | 2026-07-10 | 平铺简历列表、上传/粘贴创建后直接打开详情、LLM-derived displayName、fallback 命名、禁止 raw 第一行/文件名命名、上传文件正文提取、只读原始正文、无导出/复制/编辑/改写/原件弹层和范围外入口边界（D-20） |
| [多 JD 与多简历管理](./jd-resume-management.md) | 3.5 | active | 2026-07-13 | 多 JD、平铺简历资产、Home JD 仅粘贴文本并绑定 ready Resume、最近模拟面试与 Interview list 共用固定最大列宽卡片主体和立即面试主按钮；Resume 上传/粘贴能力独立保留 |
| [首次无简历引导](./resume-onboarding.md) | 1.15 | active | 2026-07-10 | 上传/粘贴简历、注册成功后直接打开只读详情、LLM-derived displayName、上传文件正文提取和范围外 onboarding / 轻量问答边界（D-20） |

## 2 参考

暂无。前端视觉实现只以本目录文档和 `ui-design/` 静态原型为真理源；外部品牌设计系统不是项目 UI 参考。
