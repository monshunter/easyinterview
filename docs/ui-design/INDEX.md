# UI Design 文档索引

## 1 Active

| 文档 | 版本 | 状态 | 更新日期 | 说明 |
|------|------|------|----------|------|
| [目标总体架构](./ui-architecture.md) | 2.25 | active | 2026-07-10 | 三入口信息架构、TopBar、用户菜单、核心 JD / 简历 -> 面试 -> 报告 -> 复练 / 下一轮闭环、面试入口默认展示面试规划列表，面试规划详情 / 面试上下文确认以 Parse 母版统一承接首次核对与回访，首页新建规划把上传 JD / URL source actions 整合进 JD 输入卡底部，简历下拉框与创建入口同排，简历详情只读原始正文，文本面试 / 电话模式会话和范围外 route 输入归一边界 |
| [目标用户流程](./user-flow.md) | 2.22 | active | 2026-07-10 | 首页启动、解析确认、面试规划列表、当前面试规划回访、文本面试 / 电话模式与报告、简历创建后直接打开详情、LLM-derived displayName、fallback 命名、禁止 raw 第一行/文件名命名、只读原始简历详情、认证设置和范围外流程边界 |
| [目标模块地图](./module-map.md) | 2.14 | active | 2026-07-10 | 三入口模块、面试规划列表 landing、文本面试与电话模式共享的会话级页面、简历 LLM-derived displayName、禁止 raw 第一行/文件名命名、设置认证、通知/订阅不作为当前设置页 tab、范围外 route 输入归一和当前目标数据依赖 |
| [认证与默认入口](./auth-and-entry.md) | 1.21 | active | 2026-07-10 | 默认进首页后的邮箱验证码登录、首次账号资料补全、pending action 接续、三入口用户菜单和范围外 `debrief` / `profile` 负向 |
| [当前面试规划目标模块](./module-job-workspace.md) | 1.30 | active | 2026-07-10 | 面试一级 `workspace` 是纯规划列表页：只展示 ready 且标题非空的 TargetJob 卡片，卡片主体与 Home 最近模拟面试同源，包含公司/状态 eyebrow、岗位、地点、mini round rail、稳定固定最大列宽、卡片点击进入规划详情、footer `立即面试` 和删除图标；启动进入文本面试 / 电话模式共享的 Interview Session；解析失败、非 ready 或空标题 JD 不得进入列表 |
| [模拟面试与报告目标模块](./module-practice-review.md) | 1.21 | active | 2026-07-11 | 完整模拟面试、统一文本/电话模式外层骨架、单一电话图标进入/退出、红色圆形挂断回到同一 session 文本模式、无重新开始、VAD 静音提交与真实 speech-start 打断、真实岗位上下文、out-of-scope 复盘边界和报告 CTA 直接进入面试 session 的仪表盘闭环 |
| [报告仪表盘目标结构](./report-dashboard.md) | 1.14 | active | 2026-07-10 | session-scoped 报告、文本 / 电话模式沟通形式展示、Summary Cards、二级详情、失败/缺 session 状态不展示估算评分、Header 唯一一对复练当前轮与进入下一轮 CTA（D-19）、题目回顾加入本轮复练为标记动作 |
| [简历一级模块](./resume-module.md) | 3.5 | active | 2026-07-10 | 平铺简历列表、上传/粘贴创建后直接打开详情、LLM-derived displayName、fallback 命名、禁止 raw 第一行/文件名命名、上传文件正文提取、只读原始正文、无导出/复制/编辑/改写/原件弹层和范围外入口边界（D-20） |
| [多 JD 与多简历管理](./jd-resume-management.md) | 3.3 | active | 2026-07-10 | 多 JD、平铺简历资产、首页新建模拟面试规划在同一 JD 输入卡内整合粘贴 / 上传 / URL source actions、下拉选择已有 ready 简历、创建入口同排、最近模拟面试 3 条 + 更多进入一级面试列表、Home recent 与 Interview list 共用固定最大列宽卡片主体和立即面试主按钮且 Home 不展示删除按钮、`resumeId` 绑定关系、LLM-derived displayName、解析前来源信息、只读原始详情正文和模拟面试规划绑定关系（D-17/D-20） |
| [首次无简历引导](./resume-onboarding.md) | 1.15 | active | 2026-07-10 | 上传/粘贴简历、注册成功后直接打开只读详情、LLM-derived displayName、上传文件正文提取和范围外 onboarding / 轻量问答边界（D-20） |

## 2 参考

暂无。前端视觉实现只以本目录文档和 `ui-design/` 静态原型为真理源；外部品牌设计系统不是项目 UI 参考。
