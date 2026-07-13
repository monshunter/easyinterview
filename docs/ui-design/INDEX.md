# UI Design 文档索引

## 1 Active

| 文档 | 版本 | 状态 | 更新日期 | 说明 |
|------|------|------|----------|------|
| [目标总体架构](./ui-architecture.md) | 2.27 | active | 2026-07-12 | 三入口信息架构、desktop/mobile 响应式 TopBar、连续文本面试、disabled 电话入口、会话级报告与范围外 route 输入归一边界 |
| [目标用户流程](./user-flow.md) | 2.23 | active | 2026-07-12 | 首页启动、规划回访、连续文本面试、会话级报告、简历与认证设置流程 |
| [目标模块地图](./module-map.md) | 2.15 | active | 2026-07-12 | 三入口模块、连续文本 Interview Session、会话级 Report Dashboard 与当前数据依赖 |
| [认证与默认入口](./auth-and-entry.md) | 1.21 | active | 2026-07-10 | 默认进首页后的邮箱验证码登录、首次账号资料补全、pending action 接续、三入口用户菜单和范围外 `debrief` / `profile` 负向 |
| [当前面试规划目标模块](./module-job-workspace.md) | 1.31 | active | 2026-07-12 | 面试一级 `workspace` 是纯规划列表页；启动进入连续文本 Interview Session，对话只使用已确认的 JD、简历和轮次上下文 |
| [模拟面试与报告目标模块](./module-practice-review.md) | 1.26 | active | 2026-07-12 | 全宽连续文本聊天、零回答完成门禁、disabled 电话入口、grounded 会话级报告和服务端派生 CTA 闭环 |
| [报告仪表盘目标结构](./report-dashboard.md) | 1.29 | active | 2026-07-13 | reportId-only frozen context、action-local10/20/40 + async-attempt separation、maxAttempts49 honest wait、no regenerate claim、200 fuse、24/64 typed-invalid、desktop+390 parity |
| [简历一级模块](./resume-module.md) | 3.5 | active | 2026-07-10 | 平铺简历列表、上传/粘贴创建后直接打开详情、LLM-derived displayName、fallback 命名、禁止 raw 第一行/文件名命名、上传文件正文提取、只读原始正文、无导出/复制/编辑/改写/原件弹层和范围外入口边界（D-20） |
| [多 JD 与多简历管理](./jd-resume-management.md) | 3.3 | active | 2026-07-10 | 多 JD、平铺简历资产、首页新建模拟面试规划在同一 JD 输入卡内整合粘贴 / 上传 / URL source actions、下拉选择已有 ready 简历、创建入口同排、最近模拟面试 3 条 + 更多进入一级面试列表、Home recent 与 Interview list 共用固定最大列宽卡片主体和立即面试主按钮且 Home 不展示删除按钮、`resumeId` 绑定关系、LLM-derived displayName、解析前来源信息、只读原始详情正文和模拟面试规划绑定关系（D-17/D-20） |
| [首次无简历引导](./resume-onboarding.md) | 1.15 | active | 2026-07-10 | 上传/粘贴简历、注册成功后直接打开只读详情、LLM-derived displayName、上传文件正文提取和范围外 onboarding / 轻量问答边界（D-20） |

## 2 参考

暂无。前端视觉实现只以本目录文档和 `ui-design/` 静态原型为真理源；外部品牌设计系统不是项目 UI 参考。
